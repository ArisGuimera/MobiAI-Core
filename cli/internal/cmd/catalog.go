// Package cmd holds CLI subcommand wiring. The catalog subcommand is
// hidden — it exists for end-to-end smoke testing of the catalog +
// resolver layers, not for end users.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/resolver"
)

// NewCatalogCmd builds the hidden `mobiai catalog` command tree.
func NewCatalogCmd() *cobra.Command {
	root := &cobra.Command{
		Use:    "catalog",
		Short:  i18n.T("Inspect the catalog (debug)"),
		Long:   i18n.T("Hidden subcommand for catalog/resolver smoke tests. Not intended for end users."),
		Hidden: true,
	}

	var rootFlag string
	addRootFlag := func(c *cobra.Command) {
		c.Flags().StringVar(&rootFlag, "root", ".", i18n.T("Path to the catalog root (directory containing .claude-plugin/marketplace.json)"))
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("List the packs in the catalog"),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootFlag
			if root == "" || root == "." {
				if r := defaultCatalogRoot(); r != "" {
					root = r
				}
			}
			c, err := catalog.Load(root)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, i18n.T("Catalog: %s (version %s)\n"), c.Marketplace.Name, c.Marketplace.Metadata.Version)
			fmt.Fprintf(out, "Packs (%d):\n", len(c.Packs))
			for _, p := range c.Packs {
				deps := i18n.T("none")
				if len(p.Manifest.Dependencies) > 0 {
					deps = strings.Join(p.Manifest.Dependencies, ", ")
				}
				fmt.Fprintf(out, "  - %-14s v%-8s deps: %s\n", p.Ref.Name, p.Manifest.Version, deps)
			}
			return nil
		},
	}
	addRootFlag(listCmd)

	resolveCmd := &cobra.Command{
		Use:   "resolve <pack>...",
		Short: i18n.T("Resolve the install order for the given packs"),
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootFlag
			if root == "" || root == "." {
				if r := defaultCatalogRoot(); r != "" {
					root = r
				}
			}
			c, err := catalog.Load(root)
			if err != nil {
				return err
			}
			order, err := resolver.Resolve(c, args)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprint(out, i18n.T("Install order:\n"))
			for i, name := range order {
				fmt.Fprintf(out, "  %d. %s\n", i+1, name)
			}
			return nil
		},
	}
	addRootFlag(resolveCmd)

	root.AddCommand(listCmd, resolveCmd)
	return root
}
