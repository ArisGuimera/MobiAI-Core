// Package tui implements the MobiAI picker TUI (Bubble Tea).
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// Mode is the TUI screen the user is currently looking at.
type Mode int

const (
	ModePicker Mode = iota
	ModeConfirm
	ModeProgress
	ModeResult
	ModeNoHosts
)

// Model is the Bubble Tea model.
type Model struct {
	catalog *catalog.Catalog
	state   *state.Installed
	hosts   []host.HostAdapter

	mode   Mode
	cursor int

	width, height int
}

// NewModel builds the picker model from already-loaded catalog/state/hosts.
func NewModel(c *catalog.Catalog, s *state.Installed, hosts []host.HostAdapter) Model {
	mode := ModePicker
	if len(hosts) == 0 {
		mode = ModeNoHosts
	}
	return Model{
		catalog: c,
		state:   s,
		hosts:   hosts,
		mode:    mode,
	}
}

// Mode returns the current screen.
func (m Model) Mode() Mode { return m.mode }

// PackRow is a single row in the picker.
type PackRow struct {
	Name        string
	Description string
	Version     string
	Deps        []string

	Selected    bool     // user toggled on
	Required    bool     // required because another selected pack depends on it
	InstalledIn []string // host IDs where it's already installed
}

// PackRows builds the list of rows shown in the picker, filtering out the
// universal `core` dep (which is never user-pickable per design Sec 5.2).
func (m Model) PackRows() []PackRow {
	if m.catalog == nil {
		return nil
	}
	var rows []PackRow
	for _, p := range m.catalog.Packs {
		if p.Ref.Name == "core" {
			continue
		}
		row := PackRow{
			Name:        p.Ref.Name,
			Description: p.Ref.Description,
			Version:     p.Manifest.Version,
			Deps:        append([]string(nil), p.Manifest.Dependencies...),
		}
		if hosts, ok := m.state.Packs[p.Ref.Name]; ok {
			row.InstalledIn = append([]string(nil), hosts...)
		}
		rows = append(rows, row)
	}
	return rows
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.mode {
	case ModeNoHosts:
		return "No detecté ningún cliente de IA.\nInstalá Claude Code, Cursor, Gemini CLI o Codex y volvé a correr `mobiai`.\n"
	case ModePicker:
		return m.viewPicker()
	default:
		return "(placeholder)\n"
	}
}
