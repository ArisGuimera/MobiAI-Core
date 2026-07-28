package brain

import (
	"fmt"
	"strings"
)

// SearchHit is a single match emitted by Search.
type SearchHit struct {
	Section string      // canonical section name (decisions/bugfixes/...)
	Entry   MemoryEntry // the matched entry
	Snippet string      // first body line containing the query (or title if matched there)
}

// Search returns every entry that matches both query and filter. The
// query is a case-insensitive substring matched against title and body
// — filter constraints (platform/status/area) apply on top with AND
// semantics.
//
// Returns hits in canonical section order (decisions → bugfixes →
// testing → integrations → releases), then in the order entries appear
// in their source file.
func Search(p BrainPaths, query string, filter EntryFilter) ([]SearchHit, error) {
	groups, err := ParseAllMemories(p)
	if err != nil {
		return nil, fmt.Errorf("parse memories: %w", err)
	}
	full := filter
	if q := strings.TrimSpace(query); q != "" {
		full.Query = q
	}

	var hits []SearchHit
	for _, section := range AllSectionNames() {
		for _, e := range groups[section] {
			if !full.Matches(e) {
				continue
			}
			hits = append(hits, SearchHit{
				Section: section,
				Entry:   e,
				Snippet: snippetFor(e, full.Query),
			})
		}
	}
	return hits, nil
}

// snippetFor returns the most relevant single line of e for query.
// Preference: the first body line containing query (case-insensitive);
// falls back to the title.
func snippetFor(e MemoryEntry, query string) string {
	if query == "" {
		return firstNonEmptyLine(e.Body)
	}
	q := strings.ToLower(query)
	for _, line := range strings.Split(e.Body, "\n") {
		if strings.Contains(strings.ToLower(line), q) {
			return strings.TrimSpace(line)
		}
	}
	return e.Title
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}
