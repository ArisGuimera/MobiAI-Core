package cmd

import (
	"bufio"
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
)

// NewSkillsCmd builds `mobiai skills <add|remove|list>`.
// list is added in Task 16; this task ships add and remove.
func NewSkillsCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "skills",
		Short: "Gestionar skills de MobiAI",
	}
	// When invoked standalone (in tests), register the persistent flags
	// locally so callers can pass --catalog-root and --yes without going
	// through the root command's persistent flag set.
	root.PersistentFlags().StringSlice("host", nil, "fuerza adapters específicos (default: todos los detectados)")
	root.PersistentFlags().Bool("yes", false, "asume sí en confirmaciones")
	root.PersistentFlags().String("catalog-root", "", "ruta a un catálogo local")

	addCmd := &cobra.Command{
		Use:   "add <pack>...",
		Short: "Instalar packs en los hosts detectados",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkillsAdd(cmd, args)
		},
	}
	removeCmd := &cobra.Command{
		Use:   "remove <pack>...",
		Short: "Desinstalar packs de los hosts detectados",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkillsRemove(cmd, args)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Listar packs instalados",
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
				fmt.Fprintln(out, "No hay packs instalados.")
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

	root.AddCommand(addCmd, removeCmd, listCmd)
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
			fmt.Fprintln(out, "Cancelado.")
			return nil
		}
	}

	for _, packName := range order {
		pack, _ := c.Get(packName)
		skills, err := c.Skills(pack)
		if err != nil {
			return fmt.Errorf("enumerar skills de %s: %w", packName, err)
		}
		for _, h := range hosts {
			if err := h.Install(skills); err != nil {
				return fmt.Errorf("instalar %s en %s: %w", packName, h.ID(), err)
			}
			installed.Add(packName, h.ID())
			fmt.Fprintf(out, "✓ %s → %s\n", packName, h.Name())
		}
	}
	if err := installed.Save(paths); err != nil {
		return err
	}
	fmt.Fprintln(out, "Listo.")
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
	fmt.Fprintln(out, "Plan de instalación:")
	fmt.Fprintf(out, "  Packs (%d): %s\n", len(packs), strings.Join(packs, ", "))
	fmt.Fprintf(out, "  Hosts (%d): %s\n", len(hosts), strings.Join(hostNames, ", "))

	if f, ok := in.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
		return false, fmt.Errorf("stdin no es interactivo: pasá --yes para confirmar sin prompt")
	}

	fmt.Fprint(out, "¿Continuar? [y/N]: ")
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("leer confirmación: %w", err)
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
				return fmt.Errorf("desinstalar %s de %s: %w", packName, h.ID(), err)
			}
			installed.Remove(packName, h.ID())
			fmt.Fprintf(out, "✓ %s ← %s\n", packName, h.Name())
		}
	}
	if err := installed.Save(paths); err != nil {
		return err
	}
	fmt.Fprintln(out, "Listo.")
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
		return nil, fmt.Errorf("no encontré el catálogo. Opciones:\n" +
			"  - pasá --catalog-root <ruta>\n" +
			"  - configurá MOBIAI_CATALOG_ROOT=<ruta>\n" +
			"  - corré 'mobiai update --catalog-root <ruta>' para popular ~/.mobiai/cache/catalog/")
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
	//    plugins/ tree the embed captures). We synthesize a minimal
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
// The embedded FS only ships plugins/ — to make catalog.Load happy, we also
// synthesize a .claude-plugin/marketplace.json that lists every embedded pack.
func ensureEmbeddedExtracted(embedRoot string) error {
	marketplaceFile := filepath.Join(embedRoot, ".claude-plugin", "marketplace.json")
	if _, err := os.Stat(marketplaceFile); err == nil {
		return nil // already extracted
	}

	pluginsDst := filepath.Join(embedRoot, "plugins")
	if err := embedded.Extract(pluginsDst); err != nil {
		return err
	}

	// Synthesize marketplace.json by listing every plugin dir.
	entries, err := os.ReadDir(pluginsDst)
	if err != nil {
		return err
	}
	var plugins []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(pluginsDst, e.Name(), ".claude-plugin", "plugin.json")); err == nil {
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
		fmt.Fprintf(&b, `{"name":%q,"source":"./plugins/%s","description":"","category":""}`, name, name)
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
		return nil, fmt.Errorf("no detecté ningún cliente de IA — instalá Claude Code, Cursor, Gemini CLI o Codex")
	}
	return detected, nil
}

// detectHosts is selectHosts minus the "no hosts found" error: returns the
// auto-detected adapters (possibly empty). Used by the picker, which has a
// dedicated UI state for the empty case.
func detectHosts(g GlobalFlags) []host.HostAdapter {
	return newRegistry(g).Detect()
}
