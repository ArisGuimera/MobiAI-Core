package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// NewLangCmd builds `mobiai lang`: with no args it prints the active language;
// with `en`/`es` it persists the choice to <home>/config.json. The preference
// is applied at the next startup via i18n.Init (and immediately for the
// confirmation message below).
func NewLangCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lang [en|es]",
		Short: i18n.T("Show or set the CLI language (en/es)"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if len(args) == 0 {
				fmt.Fprintf(out, i18n.T("Current language: %s\n"), i18n.Current())
				return nil
			}

			lang, ok := i18n.Parse(args[0])
			if !ok {
				return fmt.Errorf("unsupported language %q (supported: en, es)", args[0])
			}

			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			cfg, err := state.LoadConfig(paths)
			if err != nil {
				return err
			}
			cfg.Lang = string(lang)
			if err := cfg.Save(paths); err != nil {
				return err
			}

			i18n.SetLang(lang) // so the confirmation prints in the new language
			fmt.Fprintf(out, i18n.T("Language set to %s.\n"), lang)
			return nil
		},
	}
}
