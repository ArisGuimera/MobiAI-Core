package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	b.WriteString(titleStyle.Render("MobiAI") + "    " + dimStyle.Render(fmt.Sprintf("Hosts detectados (%d): %s", len(m.hosts), strings.Join(hostNames, " · "))) + "\n\n")

	rows := m.PackRows()
	for i, r := range rows {
		marker := pickerMarker(r, len(m.hosts))

		prefix := "  "
		if i == m.cursor {
			prefix = cursorStyle.Render("▶ ")
		}
		line := fmt.Sprintf("%s%s %-14s %s", prefix, marker, r.Name, dimStyle.Render(r.Description))
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + dimStyle.Render("↑↓ navegar  [espacio] cambiar acción  [enter] aplicar  [q] salir") + "\n")
	return b.String()
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
	b.WriteString(titleStyle.Render("Confirmar cambios") + "\n\n")

	var installs, uninstalls []string
	for name, act := range m.userActions {
		switch act {
		case ActionInstall:
			installs = append(installs, name)
		case ActionUninstall:
			uninstalls = append(uninstalls, name)
		}
	}
	sort.Strings(installs)
	sort.Strings(uninstalls)

	if len(installs) == 0 && len(uninstalls) == 0 {
		b.WriteString("No hay cambios pendientes.\n")
	}
	if len(installs) > 0 {
		b.WriteString("A instalar:\n")
		for _, name := range installs {
			b.WriteString(checkboxStyle.Render("  + ") + name + "\n")
		}
	}
	if len(uninstalls) > 0 {
		if len(installs) > 0 {
			b.WriteString("\n")
		}
		b.WriteString("A desinstalar:\n")
		for _, name := range uninstalls {
			b.WriteString(removalStyle.Render("  - ") + name + "\n")
		}
	}
	b.WriteString(fmt.Sprintf("\nEn %d hosts detectados.\n\n", len(m.hosts)))
	b.WriteString(dimStyle.Render("[y] confirmar  [n] volver al picker  [esc] cancelar"))
	return b.String()
}

func (m Model) viewProgress() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Aplicando cambios... %d/%d", m.installDone, len(m.installPlan))) + "\n\n")
	for i, s := range m.installPlan {
		marker := "  "
		if i < m.installDone {
			marker = checkboxStyle.Render("✓ ")
		} else if i == m.installDone {
			marker = "⠋ "
		}
		b.WriteString(fmt.Sprintf("%s%s → %s\n", marker, s.pack, s.host))
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
		b.WriteString(checkboxStyle.Render("✓ Listo") + "\n\n")
	} else {
		b.WriteString(fmt.Sprintf("⚠ Terminado con %d errores\n\n", len(m.installErrors)))
	}

	errBy := make(map[[2]string]error, len(m.installErrors))
	for _, e := range m.installErrors {
		errBy[[2]string{e.pack, e.host}] = e.err
	}
	for _, s := range m.installPlan {
		if err, bad := errBy[[2]string{s.pack, s.host}]; bad {
			b.WriteString(fmt.Sprintf("%s%s → %s: %v\n", removalStyle.Render("✗ "), s.pack, s.host, err))
		} else {
			b.WriteString(fmt.Sprintf("%s%s → %s\n", checkboxStyle.Render("✓ "), s.pack, s.host))
		}
	}

	if m.installSaveErr != nil {
		b.WriteString(fmt.Sprintf("\n⚠ No pude guardar installed.json: %v\n", m.installSaveErr))
	}
	b.WriteString(dimStyle.Render("\n[enter] cerrar"))
	return b.String()
}
