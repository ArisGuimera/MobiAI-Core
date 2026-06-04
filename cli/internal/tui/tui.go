// Package tui implements the MobiAI picker TUI (Bubble Tea).
package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/resolver"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// Mode is the TUI screen the user is currently looking at.
type Mode int

const (
	ModePicker Mode = iota
	ModeCommunityPicker
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

	// Community sub-picker: the community pack is the one pack installable
	// skill-by-skill, so selecting it drills into ModeCommunityPicker instead
	// of toggling the whole pack. subCursor indexes communitySkills there.
	subCursor        int
	communitySkills  []catalog.Skill   // community pack's skills, sorted by ID (cached)
	communityActions map[string]Action // pending action per community skill ID

	userActions   map[string]Action // pending action per pack (Install/Uninstall/None)
	requiredByDep map[string]bool   // packs marked required because of dep expansion (Install only)

	installPlan    []installStep
	installDone    int
	installErrors  []installStepDoneMsg
	installSaveErr error // error saving installed.json after the install plan finishes

	width, height int
}

// installStep is one (pack, host, kind) tuple to apply. skillIDs, when
// non-empty, restricts the step to a subset of the pack's skills — used for
// per-skill community install/uninstall. Empty means "the whole pack" (every
// platform pack, and bare-community installs, take this path unchanged).
type installStep struct {
	pack     string
	host     string
	kind     stepKind
	skillIDs []string
}

// label is the human-facing name of the step shown in progress/result views.
// A community step carrying a skill subset reads as "community/foo" (single)
// or "community (N skills)" (batched), instead of the bare pack name.
func (s installStep) label() string {
	if s.pack == catalog.CommunityPack && len(s.skillIDs) > 0 {
		if len(s.skillIDs) == 1 {
			return state.CommunitySkillKey(s.skillIDs[0])
		}
		return fmt.Sprintf(i18n.T("community (%d skills)"), len(s.skillIDs))
	}
	return s.pack
}

// installStepDoneMsg is emitted by the install runner once a step finishes.
type installStepDoneMsg struct {
	pack     string
	host     string
	kind     stepKind
	skillIDs []string
	err      error
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
		catalog:         c,
		state:           s,
		paths:           paths,
		hosts:           hosts,
		mode:            mode,
		communitySkills: loadCommunitySkills(c),
	}
}

// loadCommunitySkills returns the community pack's skills sorted by ID, or nil
// if the catalog has no community pack. Cached at construction so the picker
// doesn't hit disk on every render.
func loadCommunitySkills(c *catalog.Catalog) []catalog.Skill {
	if c == nil || !c.Has(catalog.CommunityPack) {
		return nil
	}
	pack, err := c.Get(catalog.CommunityPack)
	if err != nil {
		return nil
	}
	skills, err := c.Skills(pack)
	if err != nil {
		return nil
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].ID < skills[j].ID })
	return skills
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
	IsCommunity bool     // community pack: drills into the per-skill sub-picker
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
			IsCommunity: p.Ref.Name == catalog.CommunityPack,
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
		case ModeCommunityPicker:
			return m.updateCommunityPicker(msg)
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
		// `mobiai status`/`skills list` reflect TUI changes too. A community
		// per-skill step records one "community/<id>" key per skill instead of
		// a bare "community" key.
		keys := stateKeysFor(msg.pack, msg.skillIDs)
		switch msg.kind {
		case stepUninstall:
			for _, k := range keys {
				m.state.Remove(k, msg.host)
			}
		default:
			for _, k := range keys {
				m.state.Add(k, msg.host)
			}
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

// hasPendingActions reports whether the user has any non-None action queued,
// across both whole-pack actions and per-skill community actions.
func (m Model) hasPendingActions() bool {
	for _, a := range m.userActions {
		if a != ActionNone {
			return true
		}
	}
	for _, a := range m.communityActions {
		if a != ActionNone {
			return true
		}
	}
	return false
}

// stateKeysFor returns the installed.json keys a step writes. A community step
// with a skill subset maps to one "community/<id>" key per skill; everything
// else maps to the bare pack name.
func stateKeysFor(pack string, skillIDs []string) []string {
	if pack == catalog.CommunityPack && len(skillIDs) > 0 {
		keys := make([]string, len(skillIDs))
		for i, id := range skillIDs {
			keys[i] = state.CommunitySkillKey(id)
		}
		return keys
	}
	return []string{pack}
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

	// Community per-skill actions (kept separate from whole-pack actions).
	var commInstall, commUninstall []string
	for id, act := range m.communityActions {
		switch act {
		case ActionInstall:
			commInstall = append(commInstall, id)
		case ActionUninstall:
			commUninstall = append(commUninstall, id)
		}
	}
	sort.Strings(commInstall)
	sort.Strings(commUninstall)

	// Resolve install order at the PACK level. When any community skill is
	// being installed we feed the resolver the "community" pack (not a synthetic
	// per-skill name it doesn't know) so its dep `core` installs first; the
	// selected skill IDs ride along as a filter on the community step.
	resolveReq := append([]string(nil), installs...)
	if len(commInstall) > 0 {
		resolveReq = append(resolveReq, catalog.CommunityPack)
	}

	var plan []installStep
	if len(resolveReq) > 0 {
		order, err := resolver.Resolve(m.catalog, resolveReq)
		if err != nil {
			return nil
		}
		for _, packName := range order {
			for _, h := range m.hosts {
				step := installStep{pack: packName, host: h.ID(), kind: stepInstall}
				if packName == catalog.CommunityPack {
					step.skillIDs = commInstall
				}
				plan = append(plan, step)
			}
		}
	}
	for _, packName := range uninstalls {
		for _, h := range m.hosts {
			plan = append(plan, installStep{pack: packName, host: h.ID(), kind: stepUninstall})
		}
	}
	// Community per-skill uninstalls (no transitive removal — listed skills only).
	if len(commUninstall) > 0 {
		for _, h := range m.hosts {
			plan = append(plan, installStep{pack: catalog.CommunityPack, host: h.ID(), kind: stepUninstall, skillIDs: commUninstall})
		}
	}
	return plan
}

func (m Model) runInstallStep(idx int) tea.Cmd {
	step := m.installPlan[idx]
	hosts := m.hosts
	c := m.catalog
	return func() tea.Msg {
		done := func(err error) tea.Msg {
			return installStepDoneMsg{pack: step.pack, host: step.host, kind: step.kind, skillIDs: step.skillIDs, err: err}
		}

		var hostAdapter host.HostAdapter
		for _, h := range hosts {
			if h.ID() == step.host {
				hostAdapter = h
				break
			}
		}
		if hostAdapter == nil {
			return done(fmt.Errorf("host %q not found", step.host))
		}
		pack, err := c.Get(step.pack)
		if err != nil {
			return done(err)
		}
		skills, err := c.Skills(pack)
		if err != nil {
			return done(err)
		}
		// Per-skill steps (community) restrict the operation to a subset.
		if len(step.skillIDs) > 0 {
			skills = filterSkillsByID(skills, step.skillIDs)
		}
		switch step.kind {
		case stepUninstall:
			ids := make([]string, len(skills))
			for i, s := range skills {
				ids[i] = s.ID
			}
			if err := hostAdapter.Uninstall(ids); err != nil {
				return done(err)
			}
		default:
			if err := hostAdapter.Install(skills); err != nil {
				return done(err)
			}
		}
		return done(nil)
	}
}

// filterSkillsByID keeps only the skills whose ID is in ids, preserving order.
func filterSkillsByID(skills []catalog.Skill, ids []string) []catalog.Skill {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	out := make([]catalog.Skill, 0, len(ids))
	for _, s := range skills {
		if want[s.ID] {
			out = append(out, s)
		}
	}
	return out
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
			if rows[m.cursor].IsCommunity {
				return m.enterCommunityPicker(), nil
			}
			m = m.toggleAt(m.cursor)
		}
	case "enter":
		// Parado en el row community, enter abre la sub-pantalla de skills
		// (es el punto de entrada al selector por-skill), no aplica.
		if m.cursor < len(rows) && rows[m.cursor].IsCommunity {
			return m.enterCommunityPicker(), nil
		}
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

// enterCommunityPicker switches to the per-skill community sub-picker with the
// sub-cursor reset to the top.
func (m Model) enterCommunityPicker() Model {
	m.mode = ModeCommunityPicker
	m.subCursor = 0
	return m
}

// updateCommunityPicker drives the per-skill community sub-picker.
//
// Keys:
//   - ↑↓/kj   move the sub-cursor
//   - space/x toggle the skill under the cursor (None ↔ Install, or
//     None ↔ Uninstall when already installed in every host)
//   - enter   apply: go to confirm if anything is pending, else back to picker
//   - esc/←/h back to the main picker, keeping the pending selections
//   - q/ctrl+c quit
func (m Model) updateCommunityPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	skills := m.communitySkills
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "left", "h":
		m.mode = ModePicker
	case "up", "k":
		if m.subCursor > 0 {
			m.subCursor--
		}
	case "down", "j":
		if m.subCursor < len(skills)-1 {
			m.subCursor++
		}
	case " ", "x":
		if m.subCursor < len(skills) {
			m = m.toggleCommunityAt(m.subCursor)
		}
	case "enter":
		if m.hasPendingActions() {
			m.mode = ModeConfirm
		} else {
			m.mode = ModePicker
		}
	}
	return m, nil
}

// toggleCommunityAt cycles the action on the community skill at index i (in
// communitySkills order). Mirrors toggleAt's installed-aware cycle:
//   - skill installed in every host → None ↔ Uninstall
//   - otherwise                     → None ↔ Install
func (m Model) toggleCommunityAt(i int) Model {
	if m.communityActions == nil {
		m.communityActions = map[string]Action{}
	}
	id := m.communitySkills[i].ID
	current := m.communityActions[id]
	hosts := m.state.HostsFor(state.CommunitySkillKey(id))
	fullyInstalled := len(hosts) > 0 && len(hosts) == len(m.hosts)

	switch current {
	case ActionNone:
		if fullyInstalled {
			m.communityActions[id] = ActionUninstall
		} else {
			m.communityActions[id] = ActionInstall
		}
	default:
		delete(m.communityActions, id)
	}
	return m
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
		return i18n.T("No AI client detected.\nInstall Claude Code, Cursor, Gemini CLI or Codex and run `mobiai` again.\n")
	case ModePicker:
		return m.viewPicker()
	case ModeCommunityPicker:
		return m.viewCommunityPicker()
	case ModeConfirm:
		return m.viewConfirm()
	case ModeProgress:
		return m.viewProgress()
	case ModeResult:
		return m.viewResult()
	}
	return ""
}
