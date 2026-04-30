package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
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

	root.AddCommand(addCmd, removeCmd)
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
	return g
}

func loadCatalog(g GlobalFlags) (*catalog.Catalog, error) {
	if g.Root == "" {
		return nil, fmt.Errorf("falta --catalog-root y el cache local todavía no está implementado (Phase 5 polish)")
	}
	return catalog.Load(g.Root)
}

func selectHosts(g GlobalFlags) ([]host.HostAdapter, error) {
	r := host.NewDefaultRegistry()
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
