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
		Short: "Refresh the catalog from the remote (or a local path with --catalog-root)",
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
			fmt.Fprintf(out, "Skills catalog updated to v%s.\n", meta.Version)
			fmt.Fprintf(out, "mobiai binary: %s.\n", version)
			fmt.Fprintf(out, "%d packs available at %s.\n", len(c.Packs), rootFlag)
			return nil
		},
	}
	cmd.Flags().String("catalog-root", "", "path to a local catalog (override)")
	cmd.Flags().Bool("force", false, "if the cache dir exists but is not a git repo, delete it and re-clone")
	cmd.Flags().Bool("check", false, "only query GitHub releases and cache the result (does not touch the catalog)")
	cmd.Flags().Bool("silent", false, "print nothing on exit (intended for running from a background hook)")
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
		return fmt.Errorf("'git' is not in PATH — install it or pass --catalog-root <local-path>: %w", err)
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
		fmt.Fprintf(out, "Syncing catalog (git pull)...\n")
		c := exec.Command("git", "-C", cacheRoot, "pull", "--ff-only")
		buf := configureCmdIO(c)
		if err := c.Run(); err != nil {
			detail := ""
			if buf != nil {
				detail = strings.TrimSpace(buf.String())
			}
			if detail != "" {
				return fmt.Errorf("git pull in %s failed: %w\n%s", cacheRoot, err, detail)
			}
			return fmt.Errorf("git pull in %s failed: %w (try -V/--verbose to see git's output)", cacheRoot, err)
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
			return fmt.Errorf("%s exists but is not a git repo — pass --force to delete it and clone, or use --catalog-root to point at another path", cacheRoot)
		}
		if err := os.RemoveAll(cacheRoot); err != nil {
			return fmt.Errorf("clean %s: %w", cacheRoot, err)
		}
	}
	fmt.Fprintf(out, "Cloning catalog from %s...\n", url)
	c := exec.Command("git", "clone", "--depth=1", url, cacheRoot)
	buf := configureCmdIO(c)
	if err := c.Run(); err != nil {
		detail := ""
		if buf != nil {
			detail = strings.TrimSpace(buf.String())
		}
		if detail != "" {
			return fmt.Errorf("git clone %s failed: %w\n%s\n(if the repo is private, set MOBIAI_CATALOG_GIT_URL or pass --catalog-root)", url, err, detail)
		}
		return fmt.Errorf("git clone %s failed: %w (try -V/--verbose to see git's output; if the repo is private, set MOBIAI_CATALOG_GIT_URL or pass --catalog-root)", url, err)
	}
	return nil
}
