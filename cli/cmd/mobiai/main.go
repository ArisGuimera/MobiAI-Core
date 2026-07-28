package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/branding"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/cmd"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

// usageTemplate mirrors cobra's default but with localized labels and ANSI
// styling. The `head`, `styleCmd` and `dim` funcs come from the `branding`
// package and short-circuit to plain text when --no-color, NO_COLOR, or a
// non-TTY destination is detected. It is a function (not a const) so the
// headings resolve through i18n.T at build time, after the language is set.
//
// We apply the style funcs AFTER rpad so cobra's column width calc (which
// counts bytes) doesn't see the escape sequences. "Flags:" is left untranslated
// (identical in both languages).
func usageTemplate() string {
	return `{{head "` + i18n.T("Usage:") + `"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} ` + i18n.T("[command]") + `{{end}}{{if gt (len .Aliases) 0}}

{{head "` + i18n.T("Aliases:") + `"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{head "` + i18n.T("Examples:") + `"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{head "` + i18n.T("Available commands:") + `"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding | styleCmd}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{head "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{head "` + i18n.T("Global flags:") + `"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{head "` + i18n.T("Additional topics:") + `"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding | styleCmd}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

` + i18n.T("Use ") + `"{{styleCmd (printf "%s ` + i18n.T("[command]") + ` --help" .CommandPath)}}" ` + i18n.T("for more information about a command.") + `{{end}}
`
}

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
		Short:   i18n.T("MobiAI CLI — manage the MobiAI ecosystem for AI-assisted mobile development"),
		Long:    i18n.T("MobiAI CLI is the unified tool of the MobiAI ecosystem: skills, agents, MCPs, and orchestration across clients for AI-assisted mobile development. Today it manages skills compatible with the agentskills.io standard on any supported client; agents and MCP servers are coming soon."),
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
	root.SetUsageTemplate(usageTemplate())
	cmd.AddPersistentFlags(root)

	root.SetHelpCommand(&cobra.Command{
		Use:   "help [command]",
		Short: i18n.T("Help about any command"),
		Long:  i18n.T("Help about any mobiai command."),
		Run: func(c *cobra.Command, args []string) {
			target, _, err := c.Root().Find(args)
			if target == nil || err != nil {
				c.Printf(i18n.T("Unknown command %q\n"), args)
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

	// Translate the auto-added version flag description (root-only flag).
	root.InitDefaultVersionFlag()
	if f := root.Flags().Lookup("version"); f != nil {
		f.Usage = i18n.T("mobiai version")
	}

	root.AddCommand(cmd.NewCatalogCmd())
	root.AddCommand(cmd.NewHostsCmd())
	root.AddCommand(cmd.NewSkillsCmd())
	root.AddCommand(cmd.NewStatusCmd())
	root.AddCommand(cmd.NewDoctorCmd())
	root.AddCommand(cmd.NewUpdateCmd())
	root.AddCommand(cmd.NewBrainCmd())
	root.AddCommand(cmd.NewGraphCmd())
	root.AddCommand(cmd.NewLangCmd())

	// Localize the auto-generated `--help` flag usage for every command in the
	// tree. Cobra defaults it to "help for <cmd>", which would otherwise leak
	// English into the Flags section of each subcommand's help in es mode.
	localizeHelpFlags(root)
	return root
}

// localizeHelpFlags walks the command tree and rewrites each command's
// auto-generated `--help` flag usage to the active language.
func localizeHelpFlags(c *cobra.Command) {
	c.InitDefaultHelpFlag()
	if f := c.Flags().Lookup("help"); f != nil {
		f.Usage = fmt.Sprintf(i18n.T("help for %s"), c.Name())
	}
	for _, sub := range c.Commands() {
		localizeHelpFlags(sub)
	}
}

func main() {
	// Resolve the CLI language BEFORE building the command tree, so cobra
	// Short/Long/flag-usage strings render in the active language. Precedence
	// (MOBIAI_LANG > persisted preference > EN) lives in i18n.Init; here we
	// just feed it the persisted value. Config errors are non-fatal — we fall
	// back to English rather than block startup.
	var persisted string
	if paths, err := state.NewPaths(); err == nil {
		if cfg, err := state.LoadConfig(paths); err == nil {
			persisted = cfg.Lang
		}
	}
	i18n.Init(persisted)

	if err := newRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
