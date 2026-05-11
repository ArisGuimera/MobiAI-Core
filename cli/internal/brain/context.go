package brain

import (
	"fmt"
	"sort"
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

// stackSection is the canonical name for the auto-detected stack block.
// Not a memory section, but addressable via --section like the rest.
const (
	stackSection    = "stack"
	rulesSection    = "rules"
	warningsSection = "warnings"
)

// ContextOptions tweaks the output of BuildContext. Zero value renders
// everything (current behavior — back-compat with the original signature
// via BuildContext).
type ContextOptions struct {
	// Sections is the canonical list of sections to render, in order.
	// Empty = render all (stack, rules, memories…, warnings).
	Sections []string
	// Filter is applied to every memory entry. Stack/rules/warnings
	// ignore the filter (those aren't entry-shaped).
	Filter EntryFilter
}

// BuildContext assembles the Markdown context document consumed by AI
// agents with no filters or section restriction — the original entry
// point, preserved for callers that want everything.
func BuildContext(cfg Config, scan *Scan, p BrainPaths) string {
	return BuildContextWith(cfg, scan, p, ContextOptions{})
}

// BuildContextWith renders the Markdown context using opts. Section
// names and order match the spec because skills/agents grep for them.
func BuildContextWith(cfg Config, scan *Scan, p BrainPaths, opts ContextOptions) string {
	groups, err := ParseAllMemories(p)
	if err != nil {
		// Parser errors are unexpected (files are normal Markdown);
		// fall back to an empty groups map so the document still renders.
		groups = map[string][]MemoryEntry{}
	}

	include := includeMap(opts.Sections)
	var b strings.Builder

	writeHeader(&b, cfg, scan)

	for _, section := range renderOrder() {
		if include != nil {
			if _, ok := include[section]; !ok {
				continue
			}
		}
		switch section {
		case stackSection:
			writeStackSection(&b, scan)
		case rulesSection:
			writeRulesSection(&b, cfg)
		case warningsSection:
			writeWarningsSection(&b, scan)
		default: // memory section
			entries := groups[section]
			if !opts.Filter.Empty() {
				entries = FilterEntries(entries, opts.Filter)
			}
			writeMemorySection(&b, section, entries, !opts.Filter.Empty())
		}
	}
	return b.String()
}

// renderOrder returns the canonical section sequence for the context
// document: stack, rules, then each memory section in MemoryFiles
// order, then warnings.
func renderOrder() []string {
	out := []string{stackSection, rulesSection}
	for _, mf := range MemoryFiles {
		out = append(out, canonicalSection(mf.Name))
	}
	out = append(out, warningsSection)
	return out
}

// includeMap returns a lookup set of requested sections, normalized to
// lowercase trimmed. Returns nil when the request is empty (= render
// all) so callers can short-circuit the `_, ok :=` check.
func includeMap(sections []string) map[string]struct{} {
	if len(sections) == 0 {
		return nil
	}
	m := map[string]struct{}{}
	for _, s := range sections {
		key := strings.ToLower(strings.TrimSpace(s))
		if key == "" {
			continue
		}
		m[key] = struct{}{}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func writeHeader(b *strings.Builder, cfg Config, scan *Scan) {
	fmt.Fprintf(b, "# MobiAI Brain Context\n\n")
	fmt.Fprintf(b, "Project: %s\n", cfg.ProjectName)
	fmt.Fprintf(b, "Type: %s\n", projectTypeLabel(cfg.ProjectType))
	if scan != nil && cfg.ProjectType == ProjectTypeUnknown && scan.ProjectType != ProjectTypeUnknown {
		fmt.Fprintf(b, "Type (scan): %s\n", projectTypeLabel(scan.ProjectType))
	}
	platforms := cfg.Platforms
	if len(platforms) == 0 && scan != nil {
		platforms = scan.Platforms
	}
	if len(platforms) > 0 {
		fmt.Fprintf(b, "Platforms: %s\n", strings.Join(platforms, ", "))
	}
	b.WriteString("\n")
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

// writeMemorySection renders a memory section from parsed entries.
// `filtered` distinguishes "no entries exist" from "filter dropped them
// all" so the empty-state message is accurate.
func writeMemorySection(b *strings.Builder, section string, entries []MemoryEntry, filtered bool) {
	heading := memoryHeading(section)
	fmt.Fprintf(b, "%s\n\n", heading)
	if len(entries) == 0 {
		if filtered {
			b.WriteString("_No entries match the current filter._\n\n")
		} else {
			b.WriteString("_No entries yet._\n\n")
		}
		return
	}
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n")
		}
		writeEntry(b, e)
	}
	b.WriteString("\n")
}

// writeEntry serializes one MemoryEntry back to its canonical Markdown
// shape — same layout that `save` writes, so context output is
// round-tripable through the parser.
func writeEntry(b *strings.Builder, e MemoryEntry) {
	fmt.Fprintf(b, "## %s\n\n", e.Title)
	for _, k := range metaOrder(e.Meta) {
		fmt.Fprintf(b, "- %s: %s\n", k, e.Meta[k])
	}
	body := strings.TrimSpace(e.Body)
	if body != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}
}

// metaOrder returns the metadata keys in the canonical order used by
// `save`. Unknown keys are appended sorted at the end so user-added
// fields render deterministically.
func metaOrder(m map[string]string) []string {
	canonical := []string{"id", "type", "status", "platform", "area", "date", "review_after"}
	seen := map[string]bool{}
	out := make([]string, 0, len(m))
	for _, k := range canonical {
		if _, ok := m[k]; ok {
			out = append(out, k)
			seen[k] = true
		}
	}
	extras := make([]string, 0)
	for k := range m {
		if !seen[k] {
			extras = append(extras, k)
		}
	}
	sort.Strings(extras)
	out = append(out, extras...)
	return out
}

// memoryHeading maps a canonical section name to its display heading.
// Falls back to a Title-Case version for unknown sections.
func memoryHeading(section string) string {
	for _, mf := range MemoryFiles {
		if canonicalSection(mf.Name) == section {
			return mf.ContextHead
		}
	}
	return "## " + strings.Title(section)
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
