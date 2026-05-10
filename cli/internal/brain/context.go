package brain

import (
	"fmt"
	"strings"
)

// projectTypeLabel returns a human-friendly name for a project_type
// value. Falls back to the raw value when unknown.
func projectTypeLabel(t string) string {
	switch t {
	case ProjectTypeAndroid:
		return "Android"
	case ProjectTypeIOS:
		return "iOS"
	case ProjectTypeKMP:
		return "Kotlin Multiplatform"
	case ProjectTypeFlutter:
		return "Flutter"
	case ProjectTypeReactNative:
		return "React Native"
	case ProjectTypeUnknown, "":
		return "Unknown"
	default:
		return t
	}
}

// BuildContext assembles the Markdown context document consumed by AI
// agents. It is intentionally close to the spec layout — section names
// and order matter because skills/agents grep for them.
func BuildContext(cfg Config, scan *Scan, p BrainPaths) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# MobiAI Brain Context\n\n")
	fmt.Fprintf(&b, "Project: %s\n", cfg.ProjectName)
	fmt.Fprintf(&b, "Type: %s\n", projectTypeLabel(cfg.ProjectType))
	if scan != nil && cfg.ProjectType == ProjectTypeUnknown && scan.ProjectType != ProjectTypeUnknown {
		// scan has a fresher answer than config — show both so the user
		// notices and re-runs `brain init` if they want config refreshed.
		fmt.Fprintf(&b, "Type (scan): %s\n", projectTypeLabel(scan.ProjectType))
	}
	platforms := cfg.Platforms
	if len(platforms) == 0 && scan != nil {
		platforms = scan.Platforms
	}
	if len(platforms) > 0 {
		fmt.Fprintf(&b, "Platforms: %s\n", strings.Join(platforms, ", "))
	}
	b.WriteString("\n")

	writeStackSection(&b, scan)
	writeRulesSection(&b, cfg)

	for _, mf := range MemoryFiles {
		writeMemorySection(&b, p, mf)
	}

	writeWarningsSection(&b, scan)
	return b.String()
}

func writeStackSection(b *strings.Builder, scan *Scan) {
	b.WriteString("## Detected Stack\n\n")
	if scan == nil {
		b.WriteString("_No scan yet. Run `mobiai brain scan`._\n\n")
		return
	}
	any := false
	for _, row := range []struct {
		label string
		items []string
	}{
		{"UI", scan.UI},
		{"DI", scan.DI},
		{"Network", scan.Network},
		{"Persistence", scan.Persistence},
		{"Serialization", scan.Serialization},
		{"Testing", scan.Testing},
		{"Integrations", scan.Integrations},
		{"CI/CD", scan.CICD},
		{"Build systems", scan.BuildSystems},
	} {
		if len(row.items) == 0 {
			continue
		}
		any = true
		fmt.Fprintf(b, "- %s: %s\n", row.label, strings.Join(row.items, ", "))
	}
	if !any {
		b.WriteString("_No technologies detected yet._\n")
	}
	b.WriteString("\n")
}

func writeRulesSection(b *strings.Builder, cfg Config) {
	b.WriteString("## Project Rules\n\n")
	if len(cfg.Rules) == 0 {
		b.WriteString("_No rules defined._\n\n")
		return
	}
	for _, r := range cfg.Rules {
		fmt.Fprintf(b, "- %s\n", r)
	}
	b.WriteString("\n")
}

func writeMemorySection(b *strings.Builder, p BrainPaths, mf MemoryFile) {
	body := readMemory(p, mf.Name)
	body = stripLeadingTitle(body)
	body = strings.TrimSpace(body)

	fmt.Fprintf(b, "%s\n\n", mf.ContextHead)
	if body == "" || onlyHTMLComments(body) {
		b.WriteString("_No entries yet._\n\n")
		return
	}
	b.WriteString(body)
	b.WriteString("\n\n")
}

// stripLeadingTitle drops the first `# ...` line if the file starts with
// one — the section heading printed by BuildContext already covers it.
func stripLeadingTitle(content string) string {
	trimmed := strings.TrimLeft(content, "\n")
	if !strings.HasPrefix(trimmed, "# ") {
		return content
	}
	if idx := strings.Index(trimmed, "\n"); idx >= 0 {
		return trimmed[idx+1:]
	}
	return ""
}

// onlyHTMLComments returns true if the (trimmed) content has no real
// prose — only the `<!-- ... -->` placeholder we wrote at init time.
func onlyHTMLComments(content string) bool {
	stripped := content
	for {
		start := strings.Index(stripped, "<!--")
		if start < 0 {
			break
		}
		end := strings.Index(stripped[start:], "-->")
		if end < 0 {
			break
		}
		stripped = stripped[:start] + stripped[start+end+3:]
	}
	return strings.TrimSpace(stripped) == ""
}

func writeWarningsSection(b *strings.Builder, scan *Scan) {
	b.WriteString("## Warnings\n\n")
	if scan == nil || len(scan.Warnings) == 0 {
		b.WriteString("_None._\n")
		return
	}
	for _, w := range scan.Warnings {
		fmt.Fprintf(b, "- %s\n", w)
	}
}
