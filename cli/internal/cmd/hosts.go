package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
)

// NewHostsCmd builds the hidden `mobiai hosts` debug command tree.
// User-visible strings are in Spanish per the project's i18n convention,
// even though the command itself is hidden from `--help` (a dev who
// invokes it explicitly still reads the output).
func NewHostsCmd() *cobra.Command {
	root := &cobra.Command{
		Use:    "hosts",
		Short:  "Inspeccionar los hosts soportados (debug)",
		Long:   "Subcomando oculto para smoke tests del registry de hosts. No es para usuarios finales.",
		Hidden: true,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista los hosts soportados y cuáles están detectados",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := host.NewDefaultRegistry()
			out := cmd.OutOrStdout()
			adapters := r.Adapters()
			fmt.Fprintf(out, "Hosts soportados (%d adapters cargados):\n", len(adapters))
			for _, a := range adapters {
				det := a.Detect()
				marker, status, path := "-", "no detectado", ""
				if det.Found {
					marker, status, path = "✓", "OK", det.Path
				} else if len(det.Searched) > 0 {
					path = det.Searched[0]
				}
				fmt.Fprintf(out, "  %s %-14s %-30s %s\n", marker, a.Name(), path, status)
			}
			return nil
		},
	}
	root.AddCommand(listCmd)
	return root
}
