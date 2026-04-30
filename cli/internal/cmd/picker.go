package cmd

import (
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/tui"
)

// RunPicker loads catalog/state/hosts and launches the TUI picker.
// Used by `mobiai` (no args) when there is a TTY.
func RunPicker(g GlobalFlags) error {
	// Check TTY before loading anything — avoids spurious errors in CI/tests.
	if !tui.IsTTY() {
		return tui.ErrNoTTY
	}
	c, err := loadCatalog(g)
	if err != nil {
		return err
	}
	paths, err := state.NewPaths()
	if err != nil {
		return err
	}
	installed, err := state.LoadInstalled(paths)
	if err != nil {
		return err
	}
	hosts := host.NewDefaultRegistry().Detect()
	return tui.Run(c, installed, hosts)
}
