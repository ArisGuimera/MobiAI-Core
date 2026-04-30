package tui

import (
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// ErrNoTTY signals that the picker cannot run because stdin or stdout is
// not a terminal. Callers should print a friendly message and exit, or
// fall back to non-interactive flags.
var ErrNoTTY = errors.New("no hay terminal interactiva")

// Run launches the picker and blocks until the user quits. Returns ErrNoTTY
// if stdin or stdout is not a terminal.
func Run(c *catalog.Catalog, s *state.Installed, hosts []host.HostAdapter) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return ErrNoTTY
	}
	m := NewModel(c, s, hosts)
	prog := tea.NewProgram(m)
	_, err := prog.Run()
	return err
}
