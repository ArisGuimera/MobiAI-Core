package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/branding"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/cmd"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

// usageTemplate mirrors cobra's default but with Spanish labels and
// ANSI styling. The `head`, `styleCmd` and `dim` funcs come from the
// `branding` package and short-circuit to plain text when --no-color,
// NO_COLOR, or a non-TTY destination is detected.
//
// We apply the style funcs AFTER rpad so cobra's column width calc
// (which counts bytes) doesn't see the escape sequences. Wrapping the
// padded string is fine visually — the escape brackets the whole token
// including its trailing spaces.
const usageTemplate = `{{head "Uso:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [comando]{{end}}{{if gt (len .Aliases) 0}}

{{head "Alias:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{head "Ejemplos:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{head "Comandos disponibles:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding | styleCmd}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{head "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{head "Flags globales:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{head "Temas adicionales:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding | styleCmd}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Usá "{{styleCmd (printf "%s [comando] --help" .CommandPath)}}" para más información sobre un comando.{{end}}
`

func newRootCmd(v string) *cobra.Command {
	cmd.SetVersion(v)

	// Register cobra template funcs that style help output. They check
	// shouldStyle() at call time (which reads the --no-color override set
	// by PersistentPreRunE below), so help for subcommands is styled too.
	cobra.AddTemplateFunc("head", branding.StyleHeading)
	cobra.AddTemplateFunc("styleCmd", branding.StyleCmd)
	cobra.AddTemplateFunc("dim", branding.StyleDim)

	root := &cobra.Command{
		Use:     "mobiai",
		Short:   "MobiAI CLI — gestiona el ecosistema MobiAI para desarrollo móvil con IA",
		Long:    "MobiAI CLI es la herramienta unificada del ecosistema MobiAI: skills, agentes, MCPs y orquestación entre clientes para desarrollo móvil asistido por IA. Hoy gestiona skills compatibles con el standard agentskills.io en cualquier cliente compatible; próximamente sumará agentes y servidores MCP.",
		Version: v,
		// Sin subcomando: banner + help. El selector interactivo vive en
		// `mobiai skills init` (consistente con `mobiai brain init`).
		Run: func(c *cobra.Command, args []string) {
			g := cmd.FlagsFromCmd(c)
			out := c.OutOrStdout()
			branding.Print(out, g.NoColor)
			fmt.Fprintln(out)
			_ = c.Help()
		},
	}
	root.SetVersionTemplate("mobiai {{.Version}}\n")
	root.SetUsageTemplate(usageTemplate)
	cmd.AddPersistentFlags(root)

	// Spanish help subcommand (replaces cobra's auto-added English one).
	root.SetHelpCommand(&cobra.Command{
		Use:   "help [comando]",
		Short: "Ayuda sobre cualquier comando",
		Long:  "Ayuda sobre cualquier comando de mobiai.",
		Run: func(c *cobra.Command, args []string) {
			target, _, err := c.Root().Find(args)
			if target == nil || err != nil {
				c.Printf("Comando desconocido %q\n", args)
				_ = c.Root().Usage()
				return
			}
			_ = target.Help()
		},
	})

	// Disable cobra's auto-added `completion` subcommand for now;
	// re-enabled with Spanish copy in Phase 5 when shell completions
	// become useful for real users.
	root.CompletionOptions.DisableDefaultCmd = true

	// Translate the auto-added help and version flag descriptions.
	root.InitDefaultHelpFlag()
	if f := root.Flags().Lookup("help"); f != nil {
		f.Usage = "ayuda de mobiai"
	}
	root.InitDefaultVersionFlag()
	if f := root.Flags().Lookup("version"); f != nil {
		f.Usage = "versión de mobiai"
	}

	root.AddCommand(cmd.NewCatalogCmd())
	root.AddCommand(cmd.NewHostsCmd())
	root.AddCommand(cmd.NewSkillsCmd())
	root.AddCommand(cmd.NewStatusCmd())
	root.AddCommand(cmd.NewDoctorCmd())
	root.AddCommand(cmd.NewUpdateCmd())
	root.AddCommand(cmd.NewBrainCmd())
	root.AddCommand(cmd.NewGraphCmd())
	return root
}

func main() {
	if err := newRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
