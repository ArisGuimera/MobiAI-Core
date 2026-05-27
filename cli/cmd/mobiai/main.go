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
const usageTemplate = `{{head "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{head "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{head "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{head "Available commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding | styleCmd}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{head "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{head "Global flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{head "Additional topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding | styleCmd}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{styleCmd (printf "%s [command] --help" .CommandPath)}}" for more information about a command.{{end}}
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
		Short:   "MobiAI CLI — manage the MobiAI ecosystem for AI-assisted mobile development",
		Long:    "MobiAI CLI is the unified tool of the MobiAI ecosystem: skills, agents, MCPs, and orchestration across clients for AI-assisted mobile development. Today it manages skills compatible with the agentskills.io standard on any supported client; agents and MCP servers are coming soon.",
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

	root.SetHelpCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long:  "Help about any mobiai command.",
		Run: func(c *cobra.Command, args []string) {
			target, _, err := c.Root().Find(args)
			if target == nil || err != nil {
				c.Printf("Unknown command %q\n", args)
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
		f.Usage = "help for mobiai"
	}
	root.InitDefaultVersionFlag()
	if f := root.Flags().Lookup("version"); f != nil {
		f.Usage = "mobiai version"
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
