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

// Action is the pending action a user has marked for a pack in the picker.
// Tri-state: nada, instalar, desinstalar. El cycle por espacio depende del
// estado actual del pack:
//   - pack no instalado: None ↔ Install
//   - pack instalado en todos los hosts: None ↔ Uninstall
//   - pack parcialmente instalado: None ↔ Install (relleno)
type Action int

const (
	ActionNone Action = iota
	ActionInstall
	ActionUninstall
)

// stepKind discriminates between Install and Uninstall steps in the
// install plan executed by the picker after confirmation.
type stepKind int

const (
	stepInstall stepKind = iota
	stepUninstall
)

// Model is the Bubble Tea model.
type Model struct {
	catalog *catalog.Catalog
	state   *state.Installed
	paths   state.Paths
	hosts   []host.HostAdapter

	mode   Mode
	cursor int

	userActions   map[string]Action // pending action per pack (Install/Uninstall/None)
	requiredByDep map[string]bool   // packs marked required because of dep expansion (Install only)

	installPlan    []installStep
	installDone    int
	installErrors  []installStepDoneMsg
	installSaveErr error // error saving installed.json after the install plan finishes

	width, height int
}

// installStep is one (pack, host, kind) tuple to apply.
type installStep struct {
	pack string
	host string
	kind stepKind
}

// installStepDoneMsg is emitted by the install runner once a step finishes.
type installStepDoneMsg struct {
	pack string
	host string
	kind stepKind
	err  error
}

// NewModel builds the picker model from already-loaded catalog/state/hosts.
// paths is used to persist installed.json after a successful install. If
// paths.Home is empty (e.g., in unit tests), persistence is skipped.
func NewModel(c *catalog.Catalog, s *state.Installed, paths state.Paths, hosts []host.HostAdapter) Model {
	mode := ModePicker
	if len(hosts) == 0 {
		mode = ModeNoHosts
	}
	return Model{
		catalog: c,
		state:   s,
		paths:   paths,
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

	Action      Action   // pending action (None/Install/Uninstall)
	Required    bool     // required because another Install-marked pack depends on it
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
			Action:      m.userActions[p.Ref.Name],
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
	} else if m.state != nil {
		// Mirror the persistence behavior of `mobiai skills add/remove`: every
		// successful (pack, host) is recorded in installed.json so that
		// `mobiai status`/`skills list` reflect TUI changes too.
		switch msg.kind {
		case stepUninstall:
			m.state.Remove(msg.pack, msg.host)
		default:
			m.state.Add(msg.pack, msg.host)
		}
	}
	m.installDone++
	if m.installDone >= len(m.installPlan) {
		m.mode = ModeResult
		if m.paths.Home != "" && m.state != nil {
			if err := m.state.Save(m.paths); err != nil {
				m.installSaveErr = err
			}
		}
		return m, nil
	}
	return m, m.runInstallStep(m.installDone)
}

// hasPendingActions reports whether the user has any non-None action queued.
func (m Model) hasPendingActions() bool {
	for _, a := range m.userActions {
		if a != ActionNone {
			return true
		}
	}
	return false
}

// buildInstallPlan produces the full ordered list of (pack, host, kind)
// steps to apply. Install steps come first (in resolved dep order so deps
// install before dependents), uninstall steps come second (no transitive
// removal — user-listed packs only).
func (m Model) buildInstallPlan() []installStep {
	if !m.hasPendingActions() || len(m.hosts) == 0 {
		return nil
	}

	var installs, uninstalls []string
	for name, act := range m.userActions {
		switch act {
		case ActionInstall:
			installs = append(installs, name)
		case ActionUninstall:
			uninstalls = append(uninstalls, name)
		}
	}

	var plan []installStep
	if len(installs) > 0 {
		order, err := resolver.Resolve(m.catalog, installs)
		if err != nil {
			return nil
		}
		for _, packName := range order {
			for _, h := range m.hosts {
				plan = append(plan, installStep{pack: packName, host: h.ID(), kind: stepInstall})
			}
		}
	}
	for _, packName := range uninstalls {
		for _, h := range m.hosts {
			plan = append(plan, installStep{pack: packName, host: h.ID(), kind: stepUninstall})
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
			return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind, err: fmt.Errorf("host %q not found", step.host)}
		}
		pack, err := c.Get(step.pack)
		if err != nil {
			return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind, err: err}
		}
		skills, err := c.Skills(pack)
		if err != nil {
			return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind, err: err}
		}
		switch step.kind {
		case stepUninstall:
			ids := make([]string, len(skills))
			for i, s := range skills {
				ids[i] = s.ID
			}
			if err := hostAdapter.Uninstall(ids); err != nil {
				return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind, err: err}
			}
		default:
			if err := hostAdapter.Install(skills); err != nil {
				return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind, err: err}
			}
		}
		return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind}
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
		// Si no hay acciones pendientes, auto-cyclear el pack bajo el
		// cursor antes de transicionar. Matchea expectativa "estoy
		// parado en kmp, le doy enter, instalo (o desinstalo) kmp"
		// sin requerir que conozca el toggle por espacio.
		if !m.hasPendingActions() && m.cursor < len(rows) {
			m = m.toggleAt(m.cursor)
		}
		m.mode = ModeConfirm
	}
	return m, nil
}

// toggleAt cycles the action on the pack at index i (in PackRows order)
// and recomputes the required-by-dep set.
//
// Cycle:
//   - pack fully installed (in all detected hosts) → None ↔ Uninstall
//   - pack not installed (or partial) → None ↔ Install
//
// Required-by-dep is computed only for Install actions (uninstall doesn't
// auto-pull deps; user removes packs explicitly to avoid surprise removals).
func (m Model) toggleAt(i int) Model {
	rows := m.PackRows()
	row := rows[i]
	target := row.Name

	if m.userActions == nil {
		m.userActions = map[string]Action{}
	}

	current := m.userActions[target]
	fullyInstalled := len(row.InstalledIn) > 0 && len(row.InstalledIn) == len(m.hosts)

	switch current {
	case ActionNone:
		if fullyInstalled {
			m.userActions[target] = ActionUninstall
		} else {
			m.userActions[target] = ActionInstall
		}
	default:
		// Either Install or Uninstall → reset to None.
		delete(m.userActions, target)
	}

	// Recompute required-by-dep: only Install actions pull transitive deps.
	required := map[string]bool{}
	var installReq []string
	for name, act := range m.userActions {
		if act == ActionInstall {
			installReq = append(installReq, name)
		}
	}
	if len(installReq) > 0 {
		order, err := resolver.Resolve(m.catalog, installReq)
		if err == nil {
			for _, name := range order {
				if m.userActions[name] != ActionInstall && name != "core" {
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
		return "No AI client detected.\nInstall Claude Code, Cursor, Gemini CLI or Codex and run `mobiai` again.\n"
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
