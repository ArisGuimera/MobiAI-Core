package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// A card is the 0.3 storage format: one note per file at
// memories/<section>/<id>.md. The layout is the same Markdown that `save`
// wrote in 0.2 (H2 title, `- key: value` metadata, body), but because the
// file holds a single note, only the FIRST `## ` line is the title — any
// later `## ` line belongs to the body (fixes C5 for cards).

// canonicalMetaKeys is the order in which metadata is written to a card
// (and rendered by `context`). Keys outside this list follow, in their
// original order.
var canonicalMetaKeys = []string{
	"id", "type", "status", "platform", "area", "date", "verified", "review_after",
	"supersedes", "until", "crash", "paths",
}

// ParseCard reads a card file. A card without an `id` metadata line takes
// its id from the filename, which is the id by construction.
func ParseCard(path string) (MemoryEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MemoryEntry{}, fmt.Errorf("read %s: %w", path, err)
	}
	e := ParseCardContent(string(data))
	base := filepath.Base(path)
	e.File = base
	if e.Get("id") == "" {
		e.Meta["id"] = strings.TrimSuffix(base, ".md")
	}
	return e, nil
}

// ParseCardContent parses the text of a card. Line endings are normalized,
// so a CRLF checkout parses exactly like an LF one.
func ParseCardContent(content string) MemoryEntry {
	lines := strings.Split(normalizeNewlines(content), "\n")
	e := MemoryEntry{Meta: map[string]string{}}

	i := 0
	for ; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			e.Title = strings.TrimSpace(strings.TrimPrefix(lines[i], "## "))
			i++
			break
		}
	}
	// Metadata: contiguous `- key: value` bullets, blank lines allowed
	// before and between them (same tolerance as the old parser).
	for ; i < len(lines); i++ {
		if k, v, ok := parseMetaLine(lines[i]); ok {
			e.Meta[k] = v
			continue
		}
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		break
	}
	if i < len(lines) {
		e.Body = strings.TrimSpace(strings.Join(lines[i:], "\n"))
	}
	e.Files = parseFilesSection(e.Body)
	return e
}

// parseFilesSection returns the bullet paths listed under a `### Files`
// heading, stopping at the next heading or the first non-bullet line.
// Surrounding backticks are stripped (`path` and path are the same file).
func parseFilesSection(body string) []string {
	var files []string
	in := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			in = strings.EqualFold(strings.TrimSpace(strings.TrimLeft(trimmed, "#")), "files")
			continue
		}
		if !in {
			continue
		}
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			in = false
			continue
		}
		f := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), "`")
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

// normalizeNewlines converts CRLF (and stray CR) line endings to LF.
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}
