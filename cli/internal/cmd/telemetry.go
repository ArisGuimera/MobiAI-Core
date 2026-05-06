package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

type telemetryConfig struct {
	Enabled bool `json:"enabled"`
}

func telemetryConfigPath(p state.Paths) string {
	return filepath.Join(p.Home, "telemetry.json")
}

func loadTelemetry(p state.Paths) (telemetryConfig, error) {
	data, err := os.ReadFile(telemetryConfigPath(p))
	if err != nil {
		if os.IsNotExist(err) {
			return telemetryConfig{Enabled: false}, nil
		}
		return telemetryConfig{}, err
	}
	var c telemetryConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return telemetryConfig{}, err
	}
	return c, nil
}

func saveTelemetry(p state.Paths, c telemetryConfig) error {
	if err := os.MkdirAll(p.Home, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(telemetryConfigPath(p), data, 0o644)
}

func NewTelemetryCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "telemetry",
		Short: "Activar, desactivar o consultar el estado de la telemetría",
	}

	on := &cobra.Command{
		Use:   "on",
		Short: "Activar la telemetría opt-in",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := state.NewPaths()
			if err != nil {
				return err
			}
			if err := saveTelemetry(p, telemetryConfig{Enabled: true}); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Telemetría: on (gracias por ayudar a mejorar MobiAI).")
			return nil
		},
	}
	off := &cobra.Command{
		Use:   "off",
		Short: "Desactivar la telemetría",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := state.NewPaths()
			if err != nil {
				return err
			}
			if err := saveTelemetry(p, telemetryConfig{Enabled: false}); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Telemetría: off.")
			return nil
		},
	}
	status := &cobra.Command{
		Use:   "status",
		Short: "Mostrar el estado actual de la telemetría",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := state.NewPaths()
			if err != nil {
				return err
			}
			c, err := loadTelemetry(p)
			if err != nil {
				return err
			}
			label := "off"
			if c.Enabled {
				label = "on"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Telemetría: %s\n", label)
			return nil
		},
	}

	root.AddCommand(on, off, status)
	return root
}
