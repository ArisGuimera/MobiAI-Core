// Package cmd holds CLI subcommand wiring. The catalog subcommand is
// hidden — it exists for end-to-end smoke testing of the catalog +
// resolver layers, not for end users.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/resolver"
)

// NewCatalogCmd builds the hidden `mobiai catalog` command tree.
func NewCatalogCmd() *cobra.Command {
	root := &cobra.Command{
		Use:    "catalog",
		Short:  "Inspeccionar el catálogo (debug)",
		Long:   "Subcomando oculto para smoke tests del catálogo/resolver. No es para usuarios finales.",
		Hidden: true,
	}

	var rootFlag string
	addRootFlag := func(c *cobra.Command) {
		c.Flags().StringVar(&rootFlag, "root", ".", "Ruta a la raíz del catálogo (el directorio que contiene .claude-plugin/marketplace.json)")
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista los packs del catálogo",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := catalog.Load(rootFlag)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Catálogo: %s (versión %s)\n", c.Marketplace.Name, c.Marketplace.Metadata.Version)
			fmt.Fprintf(out, "Packs (%d):\n", len(c.Packs))
			for _, p := range c.Packs {
				deps := "ninguna"
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
		Short: "Resuelve el orden de instalación para los packs dados",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := catalog.Load(rootFlag)
			if err != nil {
				return err
			}
			order, err := resolver.Resolve(c, args)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Orden de instalación:\n")
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
