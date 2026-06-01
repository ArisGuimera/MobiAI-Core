package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
)

// NewHostsCmd builds the hidden `mobiai hosts` debug command tree.
// User-visible strings are in English per the project's i18n convention,
// even though the command itself is hidden from `--help` (a dev who
// invokes it explicitly still reads the output).
func NewHostsCmd() *cobra.Command {
	root := &cobra.Command{
		Use:    "hosts",
		Short:  i18n.T("Inspect supported hosts (debug)"),
		Long:   i18n.T("Hidden subcommand for hosts-registry smoke tests. Not intended for end users."),
		Hidden: true,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("List supported hosts and which are detected"),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := host.NewDefaultRegistry()
			out := cmd.OutOrStdout()
			adapters := r.Adapters()
			fmt.Fprintf(out, i18n.T("Supported hosts (%d adapters loaded):\n"), len(adapters))
			for _, a := range adapters {
				det := a.Detect()
				marker, status, path := "-", i18n.T("not detected"), ""
				if det.Found {
					marker, status, path = "✓", "OK", det.Path
				} else if len(det.Searched) > 0 {
					path = det.Searched[0]
				}
				if a.Experimental() {
					status += " (experimental)"
				}
				fmt.Fprintf(out, "  %s %-14s %-30s %s\n", marker, a.Name(), path, status)
			}
			return nil
		},
	}
	root.AddCommand(listCmd)
	return root
}
