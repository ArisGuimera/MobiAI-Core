// Package tui implements the MobiAI picker TUI (Bubble Tea).
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/resolver"
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

	userSelected  map[string]bool // packs the user toggled on
	requiredByDep map[string]bool // packs marked required because of dep expansion

	installPlan   []installStep
	installDone   int
	installErrors []installStepDoneMsg

	width, height int
}

// installStep is one (pack, host) pair to install.
type installStep struct {
	pack string
	host string
}

// installStepDoneMsg is emitted by the install runner once a step finishes.
type installStepDoneMsg struct {
	pack string
	host string
	err  error
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
			Selected:    m.userSelected[p.Ref.Name],
			Required:    m.requiredByDep[p.Ref.Name],
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
		switch m.mode {
		case ModePicker:
			return m.updatePicker(msg)
		case ModeConfirm:
			return m.updateConfirm(msg)
		case ModeResult:
			if msg.String() == "enter" || msg.String() == "q" {
				return m, tea.Quit
			}
		}
	case installStepDoneMsg:
		return m.handleInstallStep(msg)
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "y":
		m.mode = ModeProgress
		m.installPlan = m.buildInstallPlan()
		m.installDone = 0
		m.installErrors = nil
		if len(m.installPlan) == 0 {
			m.mode = ModeResult
			return m, nil
		}
		return m, m.runInstallStep(0)
	case "n":
		m.mode = ModePicker
	case "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleInstallStep(msg installStepDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.installErrors = append(m.installErrors, msg)
	}
	m.installDone++
	if m.installDone >= len(m.installPlan) {
		m.mode = ModeResult
		return m, nil
	}
	return m, m.runInstallStep(m.installDone)
}

func (m Model) buildInstallPlan() []installStep {
	if len(m.userSelected) == 0 || len(m.hosts) == 0 {
		return nil
	}
	req := make([]string, 0, len(m.userSelected))
	for name := range m.userSelected {
		req = append(req, name)
	}
	order, err := resolver.Resolve(m.catalog, req)
	if err != nil {
		return nil
	}
	var plan []installStep
	for _, packName := range order {
		for _, h := range m.hosts {
			plan = append(plan, installStep{pack: packName, host: h.ID()})
		}
	}
	return plan
}

func (m Model) runInstallStep(idx int) tea.Cmd {
	step := m.installPlan[idx]
	hosts := m.hosts
	c := m.catalog
	return func() tea.Msg {
		var hostAdapter host.HostAdapter
		for _, h := range hosts {
			if h.ID() == step.host {
				hostAdapter = h
				break
			}
		}
		if hostAdapter == nil {
			return installStepDoneMsg{pack: step.pack, host: step.host, err: fmt.Errorf("host %q no encontrado", step.host)}
		}
		pack, err := c.Get(step.pack)
		if err != nil {
			return installStepDoneMsg{pack: step.pack, host: step.host, err: err}
		}
		skills, err := c.Skills(pack)
		if err != nil {
			return installStepDoneMsg{pack: step.pack, host: step.host, err: err}
		}
		if err := hostAdapter.Install(skills); err != nil {
			return installStepDoneMsg{pack: step.pack, host: step.host, err: err}
		}
		return installStepDoneMsg{pack: step.pack, host: step.host}
	}
}

func (m Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.PackRows()
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(rows)-1 {
			m.cursor++
		}
	case " ", "x":
		if m.cursor < len(rows) {
			m = m.toggleAt(m.cursor)
		}
	case "enter":
		m.mode = ModeConfirm
	}
	return m, nil
}

// toggleAt flips the user-selected flag on the pack at index i (in PackRows
// order) and recomputes the required-by-dep set via the resolver.
func (m Model) toggleAt(i int) Model {
	rows := m.PackRows()
	target := rows[i].Name

	if m.userSelected == nil {
		m.userSelected = map[string]bool{}
	}
	if m.userSelected[target] {
		delete(m.userSelected, target)
	} else {
		m.userSelected[target] = true
	}

	required := map[string]bool{}
	if len(m.userSelected) > 0 {
		req := make([]string, 0, len(m.userSelected))
		for name := range m.userSelected {
			req = append(req, name)
		}
		order, err := resolver.Resolve(m.catalog, req)
		if err == nil {
			for _, name := range order {
				if !m.userSelected[name] && name != "core" {
					required[name] = true
				}
			}
		}
	}
	m.requiredByDep = required
	return m
}

func (m Model) View() string {
	switch m.mode {
	case ModeNoHosts:
		return "No detecté ningún cliente de IA.\nInstalá Claude Code, Cursor, Gemini CLI o Codex y volvé a correr `mobiai`.\n"
	case ModePicker:
		return m.viewPicker()
	case ModeConfirm:
		return m.viewConfirm()
	case ModeProgress:
		return m.viewProgress()
	case ModeResult:
		return m.viewResult()
	}
	return ""
}
