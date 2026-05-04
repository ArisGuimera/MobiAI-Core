package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
				if err := syncRemoteCatalog(out, url, cacheRoot); err != nil {
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
			fmt.Fprintf(out, "Catálogo actualizado a versión %s.\n", meta.Version)
			fmt.Fprintf(out, "%d packs disponibles en %s.\n", len(c.Packs), rootFlag)
			return nil
		},
	}
	cmd.Flags().String("catalog-root", "", "ruta a un catálogo local (override)")
	return cmd
}

// syncRemoteCatalog clones or pulls the remote catalog repo into cacheRoot.
// Requires `git` on PATH. Returns a clear error if git is missing or auth fails.
func syncRemoteCatalog(out interface{ Write([]byte) (int, error) }, url, cacheRoot string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("'git' no está en PATH — instalalo o pasá --catalog-root <ruta-local>: %w", err)
	}

	dotGit := filepath.Join(cacheRoot, ".git")
	if _, err := os.Stat(dotGit); err == nil {
		// Already cloned — pull.
		fmt.Fprintf(out, "Sincronizando catálogo (git pull)...\n")
		c := exec.Command("git", "-C", cacheRoot, "pull", "--ff-only")
		c.Stdout = noopWriter{}
		c.Stderr = noopWriter{}
		if err := c.Run(); err != nil {
			return fmt.Errorf("git pull en %s: %w", cacheRoot, err)
		}
		return nil
	}

	// Fresh clone.
	if err := os.MkdirAll(filepath.Dir(cacheRoot), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(cacheRoot), err)
	}
	// Remove cacheRoot if it exists but isn't a git repo (e.g., from embedded
	// extract). Safer to preserve and abort, but for update semantics we expect
	// to control this dir.
	if _, err := os.Stat(cacheRoot); err == nil {
		if err := os.RemoveAll(cacheRoot); err != nil {
			return fmt.Errorf("clean %s: %w", cacheRoot, err)
		}
	}
	fmt.Fprintf(out, "Clonando catálogo desde %s...\n", url)
	c := exec.Command("git", "clone", "--depth=1", url, cacheRoot)
	c.Stdout = noopWriter{}
	c.Stderr = noopWriter{}
	if err := c.Run(); err != nil {
		return fmt.Errorf("git clone %s: %w (si el repo es privado, configurá MOBIAI_CATALOG_GIT_URL o pasá --catalog-root)", url, err)
	}
	return nil
}

// noopWriter swallows git's stdout/stderr to keep the user-facing output clean.
// On verbose mode (Phase 5+) we'd swap this for the cmd's actual writer.
type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }
