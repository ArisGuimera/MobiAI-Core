package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/embedded"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/resolver"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/tui"
)

// NewSkillsCmd builds `mobiai skills <add|remove|list>`.
// list is added in Task 16; this task ships add and remove.
func NewSkillsCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "skills",
		Short: "Manage MobiAI skills",
	}
	// When invoked standalone (in tests), register the persistent flags
	// locally so callers can pass --catalog-root and --yes without going
	// through the root command's persistent flag set.
	root.PersistentFlags().StringSlice("host", nil, "force specific adapters (default: all detected)")
	root.PersistentFlags().Bool("yes", false, "assume yes on confirmations")
	root.PersistentFlags().String("catalog-root", "", "path to a local catalog")

	addCmd := &cobra.Command{
		Use:   "add <pack>...",
		Short: "Install packs into detected hosts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkillsAdd(cmd, args)
		},
	}
	removeCmd := &cobra.Command{
		Use:   "remove <pack>...",
		Short: "Uninstall packs from detected hosts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkillsRemove(cmd, args)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			installed, err := state.LoadInstalled(paths)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(installed.Packs) == 0 {
				fmt.Fprintln(out, "No packs installed.")
				return nil
			}
			fmt.Fprintln(out, "Pack          | Hosts")
			fmt.Fprintln(out, "──────────────┼─────────────────────────────")
			names := make([]string, 0, len(installed.Packs))
			for name := range installed.Packs {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, n := range names {
				fmt.Fprintf(out, "%-13s │ %s\n", n, strings.Join(installed.Packs[n], ", "))
			}
			return nil
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive picker to install/uninstall skills",
		Long:  "Launches the TUI picker to choose skill packs and apply changes to detected clients (Claude Code, Cursor, Codex, etc).",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := flagsFromAnyCmd(cmd)
			if err := RunPicker(g); err != nil {
				if errors.Is(err, tui.ErrNoTTY) {
					return fmt.Errorf("this command requires an interactive terminal — run `mobiai skills add <pack>` to install without TUI")
				}
				return err
			}
			return nil
		},
	}

	root.AddCommand(addCmd, removeCmd, listCmd, initCmd)
	return root
}

func runSkillsAdd(cmd *cobra.Command, packs []string) error {
	g := flagsFromAnyCmd(cmd)

	c, err := loadCatalog(g)
	if err != nil {
		return err
	}
	hosts, err := selectHosts(g)
	if err != nil {
		return err
	}
	order, err := resolver.Resolve(c, packs)
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
	installed, err := state.LoadInstalled(paths)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()

	if !g.Yes {
		ok, err := confirmInstall(out, cmd.InOrStdin(), order, hosts)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "Cancelled.")
			return nil
		}
	}

	for _, packName := range order {
		pack, _ := c.Get(packName)
		skills, err := c.Skills(pack)
		if err != nil {
			return fmt.Errorf("enumerate skills for %s: %w", packName, err)
		}
		for _, h := range hosts {
			if err := h.Install(skills); err != nil {
				return fmt.Errorf("install %s into %s: %w", packName, h.ID(), err)
			}
			installed.Add(packName, h.ID())
			fmt.Fprintf(out, "✓ %s → %s\n", packName, h.Name())
		}
	}
	if err := installed.Save(paths); err != nil {
		return err
	}
	fmt.Fprintln(out, "Done.")
	return nil
}

// confirmInstall prints the install plan and reads y/N from in. Returns
// (true, nil) if the user accepts. If in is not a terminal and the caller
// did not pass --yes, it returns an error so CI scripts fail loudly instead
// of hanging on a closed pipe.
func confirmInstall(out io.Writer, in io.Reader, packs []string, hosts []host.HostAdapter) (bool, error) {
	hostNames := make([]string, 0, len(hosts))
	for _, h := range hosts {
		hostNames = append(hostNames, h.Name())
	}
	fmt.Fprintln(out, "Install plan:")
	fmt.Fprintf(out, "  Packs (%d): %s\n", len(packs), strings.Join(packs, ", "))
	fmt.Fprintf(out, "  Hosts (%d): %s\n", len(hosts), strings.Join(hostNames, ", "))

	if f, ok := in.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
		return false, fmt.Errorf("stdin is not interactive: pass --yes to confirm without a prompt")
	}

	fmt.Fprint(out, "Continue? [y/N]: ")
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes" || answer == "s" || answer == "si" || answer == "sí", nil
}

func runSkillsRemove(cmd *cobra.Command, packs []string) error {
	g := flagsFromAnyCmd(cmd)

	hosts, err := selectHosts(g)
	if err != nil {
		return err
	}
	paths, err := state.NewPaths()
	if err != nil {
		return err
	}
	installed, err := state.LoadInstalled(paths)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	c, err := loadCatalog(g)
	if err != nil {
		return err
	}
	for _, packName := range packs {
		pack, err := c.Get(packName)
		if err != nil {
			return err
		}
		skills, err := c.Skills(pack)
		if err != nil {
			return err
		}
		ids := make([]string, len(skills))
		for i, s := range skills {
			ids[i] = s.ID
		}
		for _, h := range hosts {
			if err := h.Uninstall(ids); err != nil {
				return fmt.Errorf("uninstall %s from %s: %w", packName, h.ID(), err)
			}
			installed.Remove(packName, h.ID())
			fmt.Fprintf(out, "✓ %s ← %s\n", packName, h.Name())
		}
	}
	if err := installed.Save(paths); err != nil {
		return err
	}
	fmt.Fprintln(out, "Done.")
	return nil
}

// flagsFromAnyCmd reads global flags from cmd, preferring the local lookup
// but falling back to the root command's persistent flag set when invoked
// from a child subcommand.
func flagsFromAnyCmd(cmd *cobra.Command) GlobalFlags {
	g := GlobalFlags{}
	if f, err := cmd.Flags().GetStringSlice("host"); err == nil {
		g.Hosts = f
	}
	if f, err := cmd.Flags().GetBool("yes"); err == nil {
		g.Yes = f
	}
	if f, err := cmd.Flags().GetString("catalog-root"); err == nil {
		g.Root = f
	}
	if f, err := cmd.Flags().GetBool("include-experimental"); err == nil {
		g.IncludeExperimental = f
	}
	// Honor MOBIAI_INCLUDE_EXPERIMENTAL env var as a fallback so users can
	// opt in once across a shell session without typing the flag every time.
	if !g.IncludeExperimental && os.Getenv("MOBIAI_INCLUDE_EXPERIMENTAL") != "" {
		g.IncludeExperimental = true
	}
	return g
}

func loadCatalog(g GlobalFlags) (*catalog.Catalog, error) {
	root := g.Root
	if root == "" {
		root = defaultCatalogRoot()
	}
	if root == "" {
		return nil, fmt.Errorf("could not find the catalog. Options:\n" +
			"  - pass --catalog-root <path>\n" +
			"  - set MOBIAI_CATALOG_ROOT=<path>\n" +
			"  - run 'mobiai update --catalog-root <path>' to populate ~/.mobiai/cache/catalog/")
	}
	return catalog.Load(root)
}

// defaultCatalogRoot returns the resolved catalog root using a fallback chain.
// Returns empty string if nothing is found.
func defaultCatalogRoot() string {
	// 1. Env var
	if env := os.Getenv("MOBIAI_CATALOG_ROOT"); env != "" {
		return env
	}
	paths, err := state.NewPaths()
	if err == nil {
		// 2. ~/.mobiai/cache/catalog/ (populated by 'mobiai update')
		cacheCatalog := paths.CatalogDir()
		if _, err := os.Stat(filepath.Join(cacheCatalog, ".claude-plugin", "marketplace.json")); err == nil {
			return cacheCatalog
		}
	}
	// 3. Walk up from cwd looking for .claude-plugin/marketplace.json
	cwd, err := os.Getwd()
	if err == nil {
		dir := cwd
		for {
			if _, err := os.Stat(filepath.Join(dir, ".claude-plugin", "marketplace.json")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	// 4. Embedded catalog: extract to ~/.mobiai/cache/embedded/ on first use,
	//    then return its path. The embed has no top-level marketplace.json
	//    (it lives at .claude-plugin/marketplace.json which is outside the
	//    skills/ tree the embed captures). We synthesize a minimal
	//    marketplace.json from plugin.json files at extract time.
	if !embedded.IsEmpty() && err == nil {
		paths, perr := state.NewPaths()
		if perr == nil {
			embedRoot := filepath.Join(paths.Cache(), "embedded")
			if eerr := ensureEmbeddedExtracted(embedRoot); eerr == nil {
				return embedRoot
			}
		}
	}
	return ""
}

// ensureEmbeddedExtracted writes the bundled catalog to embedRoot (idempotent).
// The embedded FS only ships skills/ — to make catalog.Load happy, we also
// synthesize a .claude-plugin/marketplace.json that lists every embedded pack.
func ensureEmbeddedExtracted(embedRoot string) error {
	marketplaceFile := filepath.Join(embedRoot, ".claude-plugin", "marketplace.json")
	if _, err := os.Stat(marketplaceFile); err == nil {
		return nil // already extracted
	}

	skillsDst := filepath.Join(embedRoot, "skills")
	if err := embedded.Extract(skillsDst); err != nil {
		return err
	}

	// Synthesize marketplace.json by listing every plugin dir.
	entries, err := os.ReadDir(skillsDst)
	if err != nil {
		return err
	}
	var plugins []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(skillsDst, e.Name(), ".claude-plugin", "plugin.json")); err == nil {
			plugins = append(plugins, e.Name())
		}
	}
	sort.Strings(plugins)

	if err := os.MkdirAll(filepath.Dir(marketplaceFile), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString(`{"name":"mobiai","owner":{"name":"MobiAI","email":""},"metadata":{"description":"Embedded","version":"embed"},"plugins":[`)
	for i, name := range plugins {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"name":%q,"source":"./skills/%s","description":"","category":""}`, name, name)
	}
	b.WriteString("]}")
	return os.WriteFile(marketplaceFile, []byte(b.String()), 0o644)
}

// newRegistry returns a fresh registry with the experimental flag
// applied per the global config. Centralized so every callsite agrees
// on how --include-experimental affects auto-detection.
func newRegistry(g GlobalFlags) *host.Registry {
	r := host.NewDefaultRegistry()
	r.SetIncludeExperimental(g.IncludeExperimental)
	return r
}

func selectHosts(g GlobalFlags) ([]host.HostAdapter, error) {
	r := newRegistry(g)
	if len(g.Hosts) > 0 {
		var out []host.HostAdapter
		for _, id := range g.Hosts {
			a, err := r.Get(id)
			if err != nil {
				return nil, err
			}
			out = append(out, a)
		}
		return out, nil
	}
	detected := r.Detect()
	if len(detected) == 0 {
		return nil, fmt.Errorf("no AI client detected — install Claude Code, Cursor, Gemini CLI, or Codex")
	}
	return detected, nil
}

// detectHosts is selectHosts minus the "no hosts found" error: returns the
// auto-detected adapters (possibly empty). Used by the picker, which has a
// dedicated UI state for the empty case.
func detectHosts(g GlobalFlags) []host.HostAdapter {
	return newRegistry(g).Detect()
}
