package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: i18n.T("Thorough diagnostics for CLI, hosts and catalog"),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, i18n.T("MobiAI %s — diagnostics\n\n"), version)

			fmt.Fprintln(out, i18n.T("Supported hosts:"))
			g := flagsFromAnyCmd(cmd)
			r := newRegistry(g)
			for _, a := range r.Adapters() {
				det := a.Detect()
				marker := "-"
				note := i18n.T("not detected")
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
			fmt.Fprintln(out, i18n.T("Catalog:"))
			if meta.LastSync.IsZero() {
				fmt.Fprintln(out, i18n.T("  (never synced)"))
			} else {
				fmt.Fprintf(out, i18n.T("  Last sync: %s\n"), meta.LastSync.Format("2006-01-02 15:04:05"))
				fmt.Fprintf(out, i18n.T("  Version:   %s\n"), meta.Version)
			}
			fmt.Fprintln(out)

			fmt.Fprintln(out, i18n.T("Drift check:"))
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
				fmt.Fprintln(out, i18n.T("  ✓ no drift detected (Verify is a stub in this version)"))
			}
			return nil
		},
	}
}
