package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
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
		marker := "[ ]"
		if r.Required {
			marker = checkboxStyle.Render("[●]")
		} else if r.Selected {
			marker = checkboxStyle.Render("[✓]")
		} else if len(r.InstalledIn) > 0 && len(r.InstalledIn) == len(m.hosts) {
			marker = checkboxStyle.Render("[✓]")
		} else if len(r.InstalledIn) > 0 {
			marker = checkboxStyle.Render("[~]")
		}

		prefix := "  "
		if i == m.cursor {
			prefix = cursorStyle.Render("▶ ")
		}
		line := fmt.Sprintf("%s%s %-14s %s", prefix, marker, r.Name, dimStyle.Render(r.Description))
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + dimStyle.Render("↑↓ navegar  [espacio] toggle  [enter] aplicar  [q] salir") + "\n")
	return b.String()
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Confirmar cambios") + "\n\n")
	if len(m.userSelected) == 0 {
		b.WriteString("No hay packs seleccionados.\n")
	} else {
		b.WriteString("A instalar:\n")
		for name := range m.userSelected {
			b.WriteString("  + " + name + "\n")
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
	if len(m.installErrors) == 0 {
		b.WriteString(checkboxStyle.Render("✓ Listo") + "\n\n")
	} else {
		b.WriteString(fmt.Sprintf("⚠ Terminado con %d errores\n\n", len(m.installErrors)))
		for _, e := range m.installErrors {
			b.WriteString(fmt.Sprintf("  ✗ %s → %s: %v\n", e.pack, e.host, e.err))
		}
	}
	b.WriteString(dimStyle.Render("\n[enter] cerrar"))
	return b.String()
}
