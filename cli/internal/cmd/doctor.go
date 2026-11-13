package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
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
			printBrainDoctor(out)
			return nil
		},
	}
}

// printBrainDoctor adds a Brain section when the current directory is
// inside a project with a Brain: the `until` conditions that can't be
// interpreted, which would otherwise silently never flag their note.
// Read-only: doctor never migrates the Brain.
func printBrainDoctor(out io.Writer) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	info, err := brain.FindProjectRoot(cwd)
	if err != nil {
		return
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, i18n.T("Brain:"))
	problems, err := brain.UntilProblems(paths)
	switch {
	case err != nil:
		fmt.Fprintf(out, i18n.T("  ⚠ could not read the Brain: %v\n"), err)
	case len(problems) == 0:
		fmt.Fprintln(out, i18n.T("  ✓ every until condition can be checked"))
	default:
		fmt.Fprintf(out, i18n.T("  ⚠ %d until condition(s) can't be checked and will never flag their note:\n"), len(problems))
		for _, p := range problems {
			fmt.Fprintf(out, "    %s\n", p)
		}
	}
}
