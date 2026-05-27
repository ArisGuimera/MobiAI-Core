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
	if err := paths.EnsureDirs(); err != nil {
		return err
	}
	installed, err := state.LoadInstalled(paths)
	if err != nil {
		return err
	}
	// For the picker we want explicit --host filtering to apply, but
	// "no hosts detected" is a UI state (ModeNoHosts), not a fatal error.
	var hosts []host.HostAdapter
	if len(g.Hosts) > 0 {
		hosts, err = selectHosts(g)
		if err != nil {
			return err
		}
	} else {
		hosts = detectHosts(g)
		hosts, err = withFirebenderHost(hosts)
		if err != nil {
			return err
		}
	}
	return tui.Run(c, installed, paths, hosts)
}
