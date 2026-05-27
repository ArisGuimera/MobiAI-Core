package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Thorough diagnostics for CLI, hosts and catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "MobiAI %s — diagnostics\n\n", version)

			fmt.Fprintln(out, "Supported hosts:")
			g := flagsFromAnyCmd(cmd)
			r := newRegistry(g)
			for _, a := range r.Adapters() {
				det := a.Detect()
				marker := "-"
				note := "not detected"
				if det.Found {
					marker = "✓"
					note = "OK"
				}
				if a.Experimental() {
					note += " (experimental)"
				}
				fmt.Fprintf(out, "  %s %-14s %-30s %s\n", marker, a.Name(), det.Path, note)
			}
			fmt.Fprintln(out)

			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			meta, _ := state.LoadMeta(paths)
			fmt.Fprintln(out, "Catalog:")
			if meta.LastSync.IsZero() {
				fmt.Fprintln(out, "  (never synced)")
			} else {
				fmt.Fprintf(out, "  Last sync: %s\n", meta.LastSync.Format("2006-01-02 15:04:05"))
				fmt.Fprintf(out, "  Version:   %s\n", meta.Version)
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
				fmt.Fprintln(out, "  ✓ no drift detected (Verify is a stub in this version)")
			}
			return nil
		},
	}
}
