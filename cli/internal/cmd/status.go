package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// version is injected by main via SetVersion. Keeps cmd package decoupled
// from the build-time -ldflags machinery.
var version = "dev"

// SetVersion is called from main to inject the build-time version string.
func SetVersion(v string) { version = v }

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Resumen de hosts detectados, packs instalados y estado del catálogo",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "MobiAI %s\n\n", version)

			g := flagsFromAnyCmd(cmd)
			r := newRegistry(g)
			adapters := r.Adapters()
			present := r.Detect()
			fmt.Fprintf(out, "Hosts presentes: %d de %d soportados\n", len(present), len(adapters))
			for _, a := range adapters {
				det := a.Detect()
				marker := "-"
				if det.Found {
					marker = "✓"
				}
				note := ""
				if a.Experimental() {
					note = " (experimental)"
				}
				fmt.Fprintf(out, "  %s %-14s %s%s\n", marker, a.Name(), det.Path, note)
			}
			fmt.Fprintln(out)

			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			installed, err := state.LoadInstalled(paths)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, "Packs instalados:")
			if len(installed.Packs) == 0 {
				fmt.Fprintln(out, "  (ninguno)")
			} else {
				names := make([]string, 0, len(installed.Packs))
				for n := range installed.Packs {
					names = append(names, n)
				}
				sort.Strings(names)
				for _, n := range names {
					fmt.Fprintf(out, "  %-13s → %s\n", n, strings.Join(installed.Packs[n], ", "))
				}
			}
			return nil
		},
	}
}
