package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refrescar el catálogo y reinstalar packs cambiados",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlag, _ := cmd.Flags().GetString("catalog-root")
			if rootFlag == "" {
				return fmt.Errorf("falta --catalog-root (clone remoto via git automático queda para polish posterior)")
			}
			c, err := catalog.Load(rootFlag)
			if err != nil {
				return err
			}
			paths, err := state.NewPaths()
			if err != nil {
				return err
			}
			if err := paths.EnsureDirs(); err != nil {
				return err
			}
			meta := &state.CatalogMeta{
				LastSync: time.Now().UTC(),
				Version:  c.Marketplace.Metadata.Version,
			}
			if err := meta.Save(paths); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Catálogo actualizado a versión %s.\n", meta.Version)
			fmt.Fprintf(out, "%d packs disponibles.\n", len(c.Packs))
			return nil
		},
	}
	// Register the flag locally so the subcommand can be tested standalone
	// without going through the root command's persistent flag set.
	cmd.Flags().String("catalog-root", "", "ruta a un catálogo local")
	return cmd
}
