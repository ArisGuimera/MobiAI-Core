package brain

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MemoryEntry is a parsed section of a memories/<file>.md file. The
// parser reads the `## Title` heading, the `- key: value` metadata
// bullets immediately after, and everything else up to the next `## `
// (or EOF) as Body.
//
// This is the structured view of the Markdown that `save` writes — same
// shape, parsed back. Used by `context` filtering and `search` so we
// don't have to grep raw markdown.
type MemoryEntry struct {
	File  string            // basename, e.g. "decisions.md"
	Title string            // H2 text (without the leading "## ")
	Meta  map[string]string // id/type/status/platform/area/date/review_after/...
	Body  string            // everything after metadata, before next entry
}

// Get returns the metadata value for key, or "" if missing.
func (e MemoryEntry) Get(key string) string {
	if e.Meta == nil {
		return ""
	}
	return e.Meta[key]
}

// ParseMemoryFile reads path and returns each `## Title` block as a
// MemoryEntry. Content before the first `## ` (file title, HTML
// comments, intro prose) is ignored — those aren't entries.
//
// Returns ([], nil) if the file does not exist; the parser is permissive
// about user edits, so an entry without metadata bullets is still
// emitted with an empty Meta map.
func ParseMemoryFile(path string) ([]MemoryEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer f.Close()

	base := filepath.Base(path)
	var entries []MemoryEntry
	var current *MemoryEntry
	// inMeta is true while we're still consuming the contiguous block
	// of `- key: value` lines right after the title. The first non-meta
	// line flips it to false and everything thereafter goes into Body.
	inMeta := false
	var body strings.Builder

	flush := func() {
		if current == nil {
			return
		}
		current.Body = strings.TrimSpace(body.String())
		entries = append(entries, *current)
		current = nil
		body.Reset()
		inMeta = false
	}

	scanner := bufio.NewScanner(f)
	// Allow generously large lines for big code blocks in entry bodies.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			flush()
			current = &MemoryEntry{
				File:  base,
				Title: strings.TrimSpace(strings.TrimPrefix(line, "## ")),
				Meta:  map[string]string{},
			}
			inMeta = true
			continue
		}
		if current == nil {
			// Pre-first-entry content (file title, HTML comment header,
			// intro prose). Discarded — only entries are interesting.
			continue
		}
		if inMeta {
			if k, v, ok := parseMetaLine(line); ok {
				current.Meta[k] = v
				continue
			}
			// Skip blank lines between title and the meta block — they
			// don't end the meta region. The first non-blank, non-meta
			// line does.
			if strings.TrimSpace(line) == "" {
				continue
			}
			inMeta = false
		}
		body.WriteString(line)
		body.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("leer %s: %w", path, err)
	}
	flush()
	return entries, nil
}

// parseMetaLine extracts (key, value, true) from a `- key: value` line.
// Returns (_, _, false) for anything else (including non-bullet lines
// or bullets without a colon).
func parseMetaLine(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- ") {
		return "", "", false
	}
	content := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
	colon := strings.Index(content, ":")
	if colon <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(content[:colon])
	val := strings.TrimSpace(content[colon+1:])
	// Reject obvious non-meta bullets — anything with whitespace in the
	// key isn't a metadata pair. Keeps us from treating a `### Files`
	// list (e.g. `- composeApp/src/...`) as metadata.
	if strings.ContainsAny(key, " \t") {
		return "", "", false
	}
	return key, val, true
}

// ParseAllMemories returns the parsed entries for every memory file
// known to the brain, keyed by canonical section name (decisions,
// bugfixes, testing, integrations, releases). Files that don't exist
// resolve to empty slices.
func ParseAllMemories(p BrainPaths) (map[string][]MemoryEntry, error) {
	out := make(map[string][]MemoryEntry, len(MemoryFiles))
	for _, mf := range MemoryFiles {
		section := canonicalSection(mf.Name)
		entries, err := ParseMemoryFile(filepath.Join(p.MemoriesDir, mf.Name))
		if err != nil {
			return nil, err
		}
		out[section] = entries
	}
	return out, nil
}

// canonicalSection maps a memory filename to the lowercase section name
// used by filters and CLI flags (e.g. "decisions.md" → "decisions").
func canonicalSection(filename string) string {
	return strings.TrimSuffix(filename, ".md")
}

// sectionForFile is the reverse of canonicalSection. Returns "" if the
// section name doesn't correspond to any known memory file.
func fileForSection(section string) string {
	want := strings.ToLower(strings.TrimSpace(section)) + ".md"
	for _, mf := range MemoryFiles {
		if mf.Name == want {
			return mf.Name
		}
	}
	return ""
}

// AllSectionNames returns the canonical section names in display order.
func AllSectionNames() []string {
	out := make([]string, 0, len(MemoryFiles))
	for _, mf := range MemoryFiles {
		out = append(out, canonicalSection(mf.Name))
	}
	return out
}
