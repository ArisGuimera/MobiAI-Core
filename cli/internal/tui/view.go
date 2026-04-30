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
