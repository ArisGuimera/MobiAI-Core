package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/cmd"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

// usageTemplate mirrors cobra's default but with Spanish labels.
const usageTemplate = `Uso:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [comando]{{end}}{{if gt (len .Aliases) 0}}

Alias:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Ejemplos:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Comandos disponibles:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Flags globales:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Temas adicionales:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Usá "{{.CommandPath}} [comando] --help" para más información sobre un comando.{{end}}
`

func newRootCmd(v string) *cobra.Command {
	root := &cobra.Command{
		Use:     "mobiai",
		Short:   "MobiAI CLI — gestiona el ecosistema MobiAI para desarrollo móvil con IA",
		Long:    "MobiAI CLI es la herramienta unificada del ecosistema MobiAI: skills, agentes, MCPs y orquestación entre clientes para desarrollo móvil asistido por IA. Hoy gestiona skills compatibles con el standard agentskills.io en cualquier cliente compatible; próximamente sumará agentes y servidores MCP.",
		Version: v,
		// When called with no args, show help instead of erroring.
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
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
	return root
}

func main() {
	if err := newRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
