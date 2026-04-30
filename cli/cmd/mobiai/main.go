package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/cmd"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func newRootCmd(v string) *cobra.Command {
	root := &cobra.Command{
		Use:     "mobiai",
		Short:   "MobiAI CLI — manage skills across AI clients",
		Long:    "MobiAI CLI installs and manages mobile-development skills across any AI client that supports the agentskills.io standard.",
		Version: v,
		// When called with no args, show help instead of erroring.
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	root.SetVersionTemplate("mobiai {{.Version}}\n")
	root.AddCommand(cmd.NewCatalogCmd())
	return root
}

func main() {
	if err := newRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
