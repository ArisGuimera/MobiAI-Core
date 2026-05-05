package cmd

import (
	"github.com/spf13/cobra"
)

// GlobalFlags hold the persistent flags shared across all subcommands.
// Subcommands read them via FlagsFromCmd(cmd).
type GlobalFlags struct {
	Hosts                []string // --host=<id,id> (default empty = all detected)
	NoTUI                bool     // --no-tui
	Yes                  bool     // --yes / -y
	Verbose              bool     // --verbose / -V
	Quiet                bool     // --quiet / -q
	NoColor              bool     // --no-color
	Offline              bool     // --offline
	CacheDir             string   // --cache-dir
	Root                 string   // --catalog-root: path to a local catalog (overrides ~/.mobiai/cache/catalog/)
	IncludeExperimental  bool     // --include-experimental: enable tier-3 adapter auto-detection
}

// AddPersistentFlags attaches the global flags to the root command.
func AddPersistentFlags(root *cobra.Command) {
	root.PersistentFlags().StringSliceP("host", "", nil, "fuerza adapters específicos (default: todos los detectados)")
	root.PersistentFlags().Bool("no-tui", false, "desactiva el picker interactivo")
	root.PersistentFlags().BoolP("yes", "y", false, "asume sí en confirmaciones (CI-friendly)")
	root.PersistentFlags().BoolP("verbose", "V", false, "más output")
	root.PersistentFlags().BoolP("quiet", "q", false, "menos output")
	root.PersistentFlags().Bool("no-color", false, "deshabilita colores ANSI")
	root.PersistentFlags().Bool("offline", false, "no refresca el catálogo, usa cache")
	root.PersistentFlags().String("cache-dir", "", "override del cache (~/.mobiai/cache)")
	root.PersistentFlags().String("catalog-root", "", "ruta a un catálogo local (override de ~/.mobiai/cache/catalog)")
	root.PersistentFlags().Bool("include-experimental", false, "incluir adapters tier-3 (paths especulativos) en la auto-detección")
}

// FlagsFromCmd extracts the global flags from a cobra command's flag set.
func FlagsFromCmd(cmd *cobra.Command) GlobalFlags {
	g := GlobalFlags{}
	g.Hosts, _ = cmd.Flags().GetStringSlice("host")
	g.NoTUI, _ = cmd.Flags().GetBool("no-tui")
	g.Yes, _ = cmd.Flags().GetBool("yes")
	g.Verbose, _ = cmd.Flags().GetBool("verbose")
	g.Quiet, _ = cmd.Flags().GetBool("quiet")
	g.NoColor, _ = cmd.Flags().GetBool("no-color")
	g.Offline, _ = cmd.Flags().GetBool("offline")
	g.CacheDir, _ = cmd.Flags().GetString("cache-dir")
	g.Root, _ = cmd.Flags().GetString("catalog-root")
	g.IncludeExperimental, _ = cmd.Flags().GetBool("include-experimental")
	return g
}
