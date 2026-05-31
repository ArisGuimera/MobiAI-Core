package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// defaultCatalogGitURL is the upstream MobiAI-Core repo. Overridable via
// MOBIAI_CATALOG_GIT_URL for testing or self-hosted forks.
const defaultCatalogGitURL = "https://github.com/ArisGuimera/MobiAI-Core.git"

func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refrescar el catálogo desde el remoto (o un local con --catalog-root)",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			g := flagsFromAnyCmd(cmd)

			// --check: solo consulta GitHub releases, escribe el cache
			// y termina. No toca el catálogo. Pensado para correr en
			// background desde los hooks SessionStart.
			if check, _ := cmd.Flags().GetBool("check"); check {
				silent, _ := cmd.Flags().GetBool("silent")
				w := out
				if silent {
					w = io.Discard
				}
				return runUpdateCheck(w, version)
			}

			rootFlag, _ := cmd.Flags().GetString("catalog-root")
			if rootFlag == "" {
				rootFlag = os.Getenv("MOBIAI_CATALOG_ROOT")
			}

			if rootFlag == "" {
				// Clone or pull the remote into ~/.mobiai/cache/catalog/.
				paths, err := state.NewPaths()
				if err != nil {
					return err
				}
				if err := paths.EnsureDirs(); err != nil {
					return err
				}
				cacheRoot := paths.CatalogDir()
				url := os.Getenv("MOBIAI_CATALOG_GIT_URL")
				if url == "" {
					url = defaultCatalogGitURL
				}
				force, _ := cmd.Flags().GetBool("force")
				if err := syncRemoteCatalog(out, url, cacheRoot, g.Verbose, force); err != nil {
					return err
				}
				rootFlag = cacheRoot
			}

			c, err := catalog.Load(rootFlag)
			if err != nil {
				return err
			}
			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			if err := paths.EnsureDirs(); err != nil {
				return err
			}
			meta := &state.CatalogMeta{
				LastSync: time.Now().UTC(),
				Version:  c.Marketplace.Metadata.Version,
			}
			if err := meta.Save(paths); err != nil {
				return err
			}
			fmt.Fprintf(out, "Catálogo de skills actualizado a v%s.\n", meta.Version)
			fmt.Fprintf(out, "%d packs disponibles en %s.\n", len(c.Packs), rootFlag)

			// Catalog is refreshed; now update the binary itself if a newer
			// release exists. A binary-update failure is non-fatal: the
			// catalog sync above already succeeded, so we warn and exit 0.
			if skip, _ := cmd.Flags().GetBool("skip-binary"); skip {
				fmt.Fprintf(out, "Binario mobiai: %s (--skip-binary, no se verificó si hay update).\n", version)
			} else if err := selfUpdateBinary(out, version); err != nil {
				fmt.Fprintf(out, "Aviso: el catálogo se actualizó, pero no pude actualizar el binario: %v\n", err)
				fmt.Fprintf(out, "Reintentá `mobiai update` más tarde, o reinstalá con el install script: https://mobiai.dev\n")
			}
			return nil
		},
	}
	cmd.Flags().String("catalog-root", "", "ruta a un catálogo local (override)")
	cmd.Flags().Bool("force", false, "si el directorio cache existe sin ser un repo git, borralo y cloná de nuevo")
	cmd.Flags().Bool("check", false, "solo consultar GitHub releases y cachear el resultado (no toca el catálogo)")
	cmd.Flags().Bool("silent", false, "no imprime nada al salir (pensado para correr desde un hook en background)")
	cmd.Flags().Bool("skip-binary", false, "no actualizar el binario mobiai, solo refrescar el catálogo")
	return cmd
}

// syncRemoteCatalog clones or pulls the remote catalog repo into cacheRoot.
// Requires `git` on PATH. Returns a clear error if git is missing or auth fails.
//
// If verbose is true, git's stdout/stderr stream to the user (out).
// Otherwise we capture them into a buffer and surface only on failure, so
// real errors (auth, DNS, proxy) don't get swallowed.
//
// If cacheRoot exists but isn't a git repo, syncRemoteCatalog refuses to
// clobber it unless force is true. This protects content the user may have
// placed there manually (e.g., a hand-curated catalog or extracted embed).
func syncRemoteCatalog(out io.Writer, url, cacheRoot string, verbose, force bool) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("'git' no está en PATH — instalalo o pasá --catalog-root <ruta-local>: %w", err)
	}

	configureCmdIO := func(c *exec.Cmd) *bytes.Buffer {
		if verbose {
			c.Stdout = out
			c.Stderr = out
			return nil
		}
		buf := &bytes.Buffer{}
		c.Stdout = buf
		c.Stderr = buf
		return buf
	}

	dotGit := filepath.Join(cacheRoot, ".git")
	if _, err := os.Stat(dotGit); err == nil {
		// Already cloned — pull.
		fmt.Fprintf(out, "Sincronizando catálogo (git pull)...\n")
		c := exec.Command("git", "-C", cacheRoot, "pull", "--ff-only")
		buf := configureCmdIO(c)
		if err := c.Run(); err != nil {
			detail := ""
			if buf != nil {
				detail = strings.TrimSpace(buf.String())
			}
			if detail != "" {
				return fmt.Errorf("git pull en %s falló: %w\n%s", cacheRoot, err, detail)
			}
			return fmt.Errorf("git pull en %s falló: %w (probá -V/--verbose para ver el output de git)", cacheRoot, err)
		}
		return nil
	}

	// Fresh clone.
	if err := os.MkdirAll(filepath.Dir(cacheRoot), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(cacheRoot), err)
	}
	if _, err := os.Stat(cacheRoot); err == nil {
		// Existing dir, not a git repo. Refuse to clobber unless --force,
		// so we don't silently destroy hand-curated content.
		if !force {
			return fmt.Errorf("%s existe pero no es un repo git — pasá --force para borrarlo y clonar, o usá --catalog-root para apuntar a otra ruta", cacheRoot)
		}
		if err := os.RemoveAll(cacheRoot); err != nil {
			return fmt.Errorf("clean %s: %w", cacheRoot, err)
		}
	}
	fmt.Fprintf(out, "Clonando catálogo desde %s...\n", url)
	c := exec.Command("git", "clone", "--depth=1", url, cacheRoot)
	buf := configureCmdIO(c)
	if err := c.Run(); err != nil {
		detail := ""
		if buf != nil {
			detail = strings.TrimSpace(buf.String())
		}
		if detail != "" {
			return fmt.Errorf("git clone %s falló: %w\n%s\n(si el repo es privado, configurá MOBIAI_CATALOG_GIT_URL o pasá --catalog-root)", url, err, detail)
		}
		return fmt.Errorf("git clone %s falló: %w (probá -V/--verbose para ver el output de git; si el repo es privado, configurá MOBIAI_CATALOG_GIT_URL o pasá --catalog-root)", url, err)
	}
	return nil
}
