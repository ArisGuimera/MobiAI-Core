package brain

import (
	"fmt"
	"os"
	"path/filepath"
)

// MemoryFile describes one Markdown memory file under <brain>/memories/.
// Order in MemoryFiles is the order used in the generated context.
type MemoryFile struct {
	Name        string // filename, e.g. "decisions.md"
	Title       string // human-readable section title
	ContextHead string // section heading printed by `brain context`
	Template    string // initial content written by `brain init`
	// Dir is the card directory under memories/ (one file per note).
	// Empty for sections that only exist in the old single-file format.
	Dir string
	// Section is the canonical section name when it differs from the old
	// filename (integrations.md feeds the knowledge section).
	Section string
}

// MemoryFiles is the canonical list of memory categories. Phase 1 ships
// 5 categories; future `brain save` subcommands will append to them.
var MemoryFiles = []MemoryFile{
	{
		Name:        "decisions.md",
		Title:       "Architecture Decisions",
		ContextHead: "## Architecture Decisions",
		Template:    decisionsTemplate,
		Dir:         "decisions",
	},
	{
		Name:        "bugfixes.md",
		Title:       "Known Bugfixes",
		ContextHead: "## Known Bugfixes",
		Template:    bugfixesTemplate,
		Dir:         "bugfixes",
	},
	{
		Name:        "testing.md",
		Title:       "Testing Patterns",
		ContextHead: "## Testing Patterns",
		Template:    testingTemplate,
		Dir:         "testing",
	},
	{
		// How the app and the services it talks to behave. Replaces the
		// reserved integrations section: notes written by hand in
		// integrations.md migrate to knowledge/.
		Name:        "integrations.md",
		Section:     "knowledge",
		Title:       "Knowledge",
		ContextHead: "## Knowledge",
		Template:    integrationsTemplate,
		Dir:         "knowledge",
	},
	{
		Name:        "releases.md",
		Title:       "Release Notes",
		ContextHead: "## Release Notes",
		Template:    releasesTemplate,
	},
}

// EnsureMemoryFiles creates the card directories and any missing old
// single-file memory that carries the format notice (decisions.md,
// bugfixes.md, testing.md), holding only its heading and the notice. Those
// files exist so a teammate on 0.2.x sees the notice instead of an empty
// brain. Existing files are left untouched (idempotent — `brain init` may
// run many times). Returns the list of files actually created.
func EnsureMemoryFiles(p BrainPaths) ([]string, error) {
	for _, mf := range MemoryFiles {
		if mf.Dir == "" {
			continue
		}
		dir := filepath.Join(p.MemoriesDir, mf.Dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", dir, err)
		}
	}
	var created []string
	for _, name := range noticeFiles {
		path := filepath.Join(p.MemoriesDir, name)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return created, fmt.Errorf("stat %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(noticeFileContent(memoryFile(name))), 0o644); err != nil {
			return created, fmt.Errorf("write %s: %w", path, err)
		}
		created = append(created, name)
	}
	return created, nil
}

// readMemory returns the file's content, or an empty string if missing.
func readMemory(p BrainPaths, name string) string {
	data, err := os.ReadFile(filepath.Join(p.MemoriesDir, name))
	if err != nil {
		return ""
	}
	return string(data)
}

const decisionsTemplate = `# Decisions

<!--
Architecture decisions specific to this project.
Append entries with: mobiai brain save decision (coming in Phase 2).
Each entry should record: title, status (active|deprecated), platform,
area, date, decision, reason, files.
-->
`

const bugfixesTemplate = `# Bugfixes

<!--
Bugfixes and workarounds worth remembering for this project.
Append entries with: mobiai brain save bugfix (coming in Phase 2).
Mark temporary workarounds as status: temporary so the agent does not
treat them as permanent decisions.
-->
`

const testingTemplate = `# Testing Patterns

<!--
Reusable testing patterns discovered for this project.
Append entries with: mobiai brain save testing (coming in Phase 2).
Include the problem, the pattern that solved it and a minimal example.
-->
`

const integrationsTemplate = `# Integrations

<!--
Notes on third-party integrations (Firebase, analytics, push, payments,
etc.) and their project-specific configuration quirks.
-->
`

const releasesTemplate = `# Releases

<!--
Release notes, checklists and lessons learned during shipping.
-->
`
