package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	removalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func (m Model) viewPicker() string {
	var b strings.Builder
	hostNames := make([]string, 0, len(m.hosts))
	for _, h := range m.hosts {
		hostNames = append(hostNames, h.Name())
	}
	b.WriteString(titleStyle.Render("MobiAI") + "    " + dimStyle.Render(fmt.Sprintf(i18n.T("Detected hosts (%d): %s"), len(m.hosts), strings.Join(hostNames, " · "))) + "\n\n")

	rows := m.PackRows()
	for i, r := range rows {
		marker := pickerMarker(r, len(m.hosts))
		desc := r.Description
		if r.IsCommunity {
			// Community is the one pack you drill into to pick skills one by one;
			// its marker reflects the per-skill selection/install state.
			marker = m.communityMarker()
			desc = strings.TrimSpace(desc + "  " + i18n.T("→ pick skills"))
		}

		prefix := "  "
		if i == m.cursor {
			prefix = cursorStyle.Render("▶ ")
		}
		line := fmt.Sprintf("%s%s %-14s %s", prefix, marker, r.Name, dimStyle.Render(desc))
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + dimStyle.Render(i18n.T("↑↓ navigate  [space] change action  [enter] apply  [q] quit")) + "\n")
	return b.String()
}

// viewCommunityPicker renders the per-skill community sub-picker: one toggleable
// row per community skill, sorted alphabetically.
func (m Model) viewCommunityPicker() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(i18n.T("Community skills")) + "\n\n")

	if len(m.communitySkills) == 0 {
		b.WriteString(dimStyle.Render(i18n.T("No community skills available yet.")) + "\n")
		b.WriteString("\n" + dimStyle.Render(i18n.T("[esc] back")) + "\n")
		return b.String()
	}

	for i, s := range m.communitySkills {
		marker := m.communitySkillMarker(s.ID)
		prefix := "  "
		if i == m.subCursor {
			prefix = cursorStyle.Render("▶ ")
		}
		line := fmt.Sprintf("%s%s %-24s %s", prefix, marker, s.ID, dimStyle.Render(s.Description))
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + dimStyle.Render(i18n.T("↑↓ navigate  [space] change action  [enter] apply  [esc] back")) + "\n")
	return b.String()
}

// communityMarker is the marker for the community row in the main picker. It
// aggregates the per-skill state: pending selections take priority, otherwise
// it shows how many community skills are already installed.
func (m Model) communityMarker() string {
	var pi, pu int
	for _, a := range m.communityActions {
		switch a {
		case ActionInstall:
			pi++
		case ActionUninstall:
			pu++
		}
	}
	switch {
	case pi > 0 && pu == 0:
		return checkboxStyle.Render("[+]")
	case pu > 0 && pi == 0:
		return removalStyle.Render("[-]")
	case pi > 0 && pu > 0:
		return checkboxStyle.Render("[~]")
	}

	total := len(m.communitySkills)
	if total == 0 {
		return "[ ]"
	}
	full, some := 0, 0
	for _, s := range m.communitySkills {
		hosts := m.state.HostsFor(state.CommunitySkillKey(s.ID))
		if len(hosts) == 0 {
			continue
		}
		some++
		if len(hosts) == len(m.hosts) {
			full++
		}
	}
	switch {
	case some == 0:
		return "[ ]"
	case full == total:
		return checkboxStyle.Render("[✓]")
	default:
		return checkboxStyle.Render("[~]")
	}
}

// communitySkillMarker is the per-skill marker inside the community sub-picker.
func (m Model) communitySkillMarker(id string) string {
	switch m.communityActions[id] {
	case ActionInstall:
		return checkboxStyle.Render("[+]")
	case ActionUninstall:
		return removalStyle.Render("[-]")
	}
	hosts := m.state.HostsFor(state.CommunitySkillKey(id))
	switch {
	case len(hosts) == 0:
		return "[ ]"
	case len(hosts) == len(m.hosts):
		return checkboxStyle.Render("[✓]")
	default:
		return checkboxStyle.Render("[~]")
	}
}

// pickerMarker returns the visual checkbox-like marker for a pack row,
// based on its pending action, required-by-dep flag, and how many of the
// detected hosts already have it installed.
//
//   - Required (dep of an Install): [●]
//   - ActionInstall:                [+] (cyan) — will install
//   - ActionUninstall:              [-] (red)  — will uninstall
//   - ActionNone, fully installed:  [✓]        — already there, no change
//   - ActionNone, partial install:  [~]        — some hosts only
//   - ActionNone, not installed:    [ ]
func pickerMarker(r PackRow, totalHosts int) string {
	if r.Required {
		return checkboxStyle.Render("[●]")
	}
	switch r.Action {
	case ActionInstall:
		return checkboxStyle.Render("[+]")
	case ActionUninstall:
		return removalStyle.Render("[-]")
	}
	switch {
	case len(r.InstalledIn) == 0:
		return "[ ]"
	case len(r.InstalledIn) == totalHosts:
		return checkboxStyle.Render("[✓]")
	default:
		return checkboxStyle.Render("[~]")
	}
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(i18n.T("Confirm changes")) + "\n\n")

	var installs, uninstalls []string
	for name, act := range m.userActions {
		switch act {
		case ActionInstall:
			installs = append(installs, name)
		case ActionUninstall:
			uninstalls = append(uninstalls, name)
		}
	}
	// Per-skill community changes render as "community/<id>" alongside packs.
	for id, act := range m.communityActions {
		switch act {
		case ActionInstall:
			installs = append(installs, state.CommunitySkillKey(id))
		case ActionUninstall:
			uninstalls = append(uninstalls, state.CommunitySkillKey(id))
		}
	}
	sort.Strings(installs)
	sort.Strings(uninstalls)

	if len(installs) == 0 && len(uninstalls) == 0 {
		b.WriteString(i18n.T("No pending changes.\n"))
	}
	if len(installs) > 0 {
		b.WriteString(i18n.T("To install:\n"))
		for _, name := range installs {
			b.WriteString(checkboxStyle.Render("  + ") + name + "\n")
		}
	}
	if len(uninstalls) > 0 {
		if len(installs) > 0 {
			b.WriteString("\n")
		}
		b.WriteString(i18n.T("To uninstall:\n"))
		for _, name := range uninstalls {
			b.WriteString(removalStyle.Render("  - ") + name + "\n")
		}
	}
	b.WriteString(fmt.Sprintf(i18n.T("\nOn %d detected hosts.\n\n"), len(m.hosts)))
	b.WriteString(dimStyle.Render(i18n.T("[y] confirm  [n] back to picker  [esc] cancel")))
	return b.String()
}

func (m Model) viewProgress() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf(i18n.T("Applying changes... %d/%d"), m.installDone, len(m.installPlan))) + "\n\n")
	for i, s := range m.installPlan {
		marker := "  "
		if i < m.installDone {
			marker = checkboxStyle.Render("✓ ")
		} else if i == m.installDone {
			marker = "⠋ "
		}
		b.WriteString(fmt.Sprintf("%s%s → %s\n", marker, s.label(), s.host))
	}
	return b.String()
}

func (m Model) viewResult() string {
	var b strings.Builder

	// Bubble Tea reemplaza la View entera en cada render: si solo
	// dibujáramos "✓ Listo" la lista de viewProgress desaparece y el
	// usuario pierde la confirmación de qué se instaló. Re-renderizamos
	// el plan completo con su marker final.
	if len(m.installErrors) == 0 {
		b.WriteString(checkboxStyle.Render(i18n.T("✓ Done")) + "\n\n")
	} else {
		b.WriteString(fmt.Sprintf(i18n.T("⚠ Finished with %d errors\n\n"), len(m.installErrors)))
	}

	// Key by (pack, host, kind): a community install and a community uninstall
	// can target the same host, so pack+host alone would collide.
	type errKey struct {
		pack, host string
		kind       stepKind
	}
	errBy := make(map[errKey]error, len(m.installErrors))
	for _, e := range m.installErrors {
		errBy[errKey{e.pack, e.host, e.kind}] = e.err
	}
	for _, s := range m.installPlan {
		if err, bad := errBy[errKey{s.pack, s.host, s.kind}]; bad {
			b.WriteString(fmt.Sprintf("%s%s → %s: %v\n", removalStyle.Render("✗ "), s.label(), s.host, err))
		} else {
			b.WriteString(fmt.Sprintf("%s%s → %s\n", checkboxStyle.Render("✓ "), s.label(), s.host))
		}
	}

	if m.installSaveErr != nil {
		b.WriteString(fmt.Sprintf(i18n.T("\n⚠ Could not save installed.json: %v\n"), m.installSaveErr))
	}
	b.WriteString(dimStyle.Render(i18n.T("\n[enter] close")))
	return b.String()
}
