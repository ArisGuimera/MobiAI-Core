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
		Short: "Summary of detected hosts, installed packs and catalog state",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "MobiAI %s\n\n", version)

			g := flagsFromAnyCmd(cmd)
			r := newRegistry(g)
			present := r.Detect()
			fmt.Fprintln(out, "Hosts:")
			if len(present) == 0 {
				fmt.Fprintln(out, "  (none detected — install Claude Code, Cursor, Gemini CLI, Codex, or another agentskills.io-compatible client)")
			} else {
				for _, a := range present {
					det := a.Detect()
					note := ""
					if a.Experimental() {
						note = " (experimental)"
					}
					fmt.Fprintf(out, "  ✓ %-14s %s%s\n", a.Name(), det.Path, note)
				}
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
			fmt.Fprintln(out, "Installed packs:")
			if len(installed.Packs) == 0 {
				fmt.Fprintln(out, "  (none)")
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
