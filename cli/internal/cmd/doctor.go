package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnóstico exhaustivo del CLI, hosts y catálogo",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "MobiAI %s — diagnóstico\n\n", version)

			fmt.Fprintln(out, "Hosts soportados:")
			r := host.NewDefaultRegistry()
			for _, a := range r.Adapters() {
				det := a.Detect()
				marker := "-"
				note := "no detectado"
				if det.Found {
					marker = "✓"
					note = "OK"
				}
				fmt.Fprintf(out, "  %s %-14s %-30s %s\n", marker, a.Name(), det.Path, note)
			}
			fmt.Fprintln(out)

			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			meta, _ := state.LoadMeta(paths)
			fmt.Fprintln(out, "Catálogo:")
			if meta.LastSync.IsZero() {
				fmt.Fprintln(out, "  (nunca se sincronizó)")
			} else {
				fmt.Fprintf(out, "  Última sincronización: %s\n", meta.LastSync.Format("2006-01-02 15:04:05"))
				fmt.Fprintf(out, "  Versión: %s\n", meta.Version)
			}
			fmt.Fprintln(out)

			fmt.Fprintln(out, "Drift check:")
			anyDrift := false
			for _, a := range r.Detect() {
				rep := a.Verify()
				if len(rep.Issues) > 0 {
					anyDrift = true
					for _, issue := range rep.Issues {
						fmt.Fprintf(out, "  ⚠ %s/%s: %s\n", a.Name(), issue.SkillID, issue.Detail)
					}
				}
			}
			if !anyDrift {
				fmt.Fprintln(out, "  ✓ sin drift detectado (Verify es stub en esta versión)")
			}
			return nil
		},
	}
}
