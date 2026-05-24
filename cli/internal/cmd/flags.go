package cmd

import (
	"github.com/spf13/cobra"
)

// GlobalFlags hold the persistent flags shared across all subcommands.
// Subcommands read them via FlagsFromCmd(cmd).
type GlobalFlags struct {
	Hosts               []string // --host=<id,id> (default empty = all detected)
	Yes                 bool     // --yes / -y
	Verbose             bool     // --verbose / -V
	NoColor             bool     // --no-color
	Root                string   // --catalog-root: path to a local catalog (overrides ~/.mobiai/cache/catalog/)
	IncludeExperimental bool     // --include-experimental: enable tier-3 adapter auto-detection
}

// AddPersistentFlags attaches the global flags to the root command.
func AddPersistentFlags(root *cobra.Command) {
	root.PersistentFlags().StringSliceP("host", "", nil, "fuerza adapters específicos (default: todos los detectados)")
	root.PersistentFlags().BoolP("yes", "y", false, "asume sí en confirmaciones (CI-friendly)")
	root.PersistentFlags().BoolP("verbose", "V", false, "más output (hoy sólo afecta a 'mobiai update')")
	root.PersistentFlags().Bool("no-color", false, "deshabilita colores ANSI en el help")
	root.PersistentFlags().String("catalog-root", "", "ruta a un catálogo local (override de ~/.mobiai/cache/catalog)")
	root.PersistentFlags().Bool("include-experimental", false, "incluir adapters tier-3 (paths especulativos) en la auto-detección")
}

// FlagsFromCmd extracts the global flags from a cobra command's flag set.
func FlagsFromCmd(cmd *cobra.Command) GlobalFlags {
	g := GlobalFlags{}
	g.Hosts, _ = cmd.Flags().GetStringSlice("host")
	g.Yes, _ = cmd.Flags().GetBool("yes")
	g.Verbose, _ = cmd.Flags().GetBool("verbose")
	g.NoColor, _ = cmd.Flags().GetBool("no-color")
	g.Root, _ = cmd.Flags().GetString("catalog-root")
	g.IncludeExperimental, _ = cmd.Flags().GetBool("include-experimental")
	return g
}
