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
}

// MemoryFiles is the canonical list of memory categories. Phase 1 ships
// 5 categories; future `brain save` subcommands will append to them.
var MemoryFiles = []MemoryFile{
	{
		Name:        "decisions.md",
		Title:       "Architecture Decisions",
		ContextHead: "## Architecture Decisions",
		Template:    decisionsTemplate,
	},
	{
		Name:        "bugfixes.md",
		Title:       "Known Bugfixes",
		ContextHead: "## Known Bugfixes",
		Template:    bugfixesTemplate,
	},
	{
		Name:        "testing.md",
		Title:       "Testing Patterns",
		ContextHead: "## Testing Patterns",
		Template:    testingTemplate,
	},
	{
		Name:        "integrations.md",
		Title:       "Integrations",
		ContextHead: "## Integrations",
		Template:    integrationsTemplate,
	},
	{
		Name:        "releases.md",
		Title:       "Release Notes",
		ContextHead: "## Release Notes",
		Template:    releasesTemplate,
	},
}

// EnsureMemoryFiles writes any missing memory file with its template.
// Existing files are left untouched (idempotent — `brain init` may run
// many times). Returns the list of files actually created.
func EnsureMemoryFiles(p BrainPaths) ([]string, error) {
	if err := os.MkdirAll(p.MemoriesDir, 0o755); err != nil {
		return nil, fmt.Errorf("crear %s: %w", p.MemoriesDir, err)
	}
	var created []string
	for _, mf := range MemoryFiles {
		path := filepath.Join(p.MemoriesDir, mf.Name)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return created, fmt.Errorf("stat %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(mf.Template), 0o644); err != nil {
			return created, fmt.Errorf("escribir %s: %w", path, err)
		}
		created = append(created, mf.Name)
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
