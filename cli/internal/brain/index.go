package brain

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// indexFooter tells the agent how to read what the index only lists.
const indexFooter = "Open the note files you need; use `mobiai brain show <id>` or `mobile_get` if you can't read files."

// renderIndexed renders the context in index mode, used when the full
// document is over limit: header, stack, rules and warnings as usual; the
// full text of active decisions, newest first, while they fit; and one
// index line for every other note. The index itself is never cut — hiding
// notes would be worse than going over the limit.
func renderIndexed(cfg Config, scan *Scan, groups map[string][]MemoryEntry, include map[string]struct{}, limit int, now time.Time) string {
	var candidates []MemoryEntry
	if included(include, "decisions") {
		for _, e := range groups["decisions"] {
			if isActiveDecision(e) {
				candidates = append(candidates, e)
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Get("date") != b.Get("date") {
			return a.Get("date") > b.Get("date")
		}
		return a.Get("id") > b.Get("id")
	})

	inFull := map[string]bool{}
	budget := limit - len(composeIndexed(cfg, scan, groups, include, nil, inFull, limit, now))
	var full []MemoryEntry
	for _, e := range candidates {
		var entry strings.Builder
		writeEntry(&entry, e)
		cost := len(entry.String()) + 1 - len(indexLine(e, now)) - 1
		if cost > budget {
			break
		}
		budget -= cost
		full = append(full, e)
		inFull[e.Get("id")] = true
	}
	return composeIndexed(cfg, scan, groups, include, full, inFull, limit, now)
}

// composeIndexed lays out an index-mode document with the given notes in
// full and every other included note as an index line.
func composeIndexed(cfg Config, scan *Scan, groups map[string][]MemoryEntry, include map[string]struct{}, full []MemoryEntry, inFull map[string]bool, limit int, now time.Time) string {
	var b strings.Builder
	writeHeader(&b, cfg, scan)
	if included(include, stackSection) {
		writeStackSection(&b, scan)
	}
	if included(include, rulesSection) {
		writeRulesSection(&b, cfg)
	}

	if included(include, "decisions") {
		b.WriteString("## Active Decisions\n\n")
		fmt.Fprintf(&b, "_The Brain context is over %d KB, so only active decisions are shown in full; every other note is listed in the index below._\n\n", limit/1024)
		for i, e := range full {
			if i > 0 {
				b.WriteString("\n")
			}
			writeEntry(&b, e)
		}
		if len(full) > 0 {
			b.WriteString("\n")
		}
	}

	b.WriteString("## Index\n\n")
	b.WriteString("_One line per note. Paths are relative to `.mobiai/brain/`._\n\n")
	for _, section := range renderOrder() {
		if !included(include, section) {
			continue
		}
		for _, e := range groups[section] {
			if inFull[e.Get("id")] {
				continue
			}
			b.WriteString(indexLine(e, now))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n" + indexFooter + "\n\n")

	if included(include, warningsSection) {
		writeWarningsSection(&b, scan)
	}
	return b.String()
}

// isActiveDecision reports whether a decision counts as a project rule and
// gets its full text in index mode.
func isActiveDecision(e MemoryEntry) bool {
	return strings.EqualFold(strings.TrimSpace(e.Get("status")), StatusActive)
}

// indexLine renders one note as a single index line, e.g. "- [bug_fix]
// Login crash · android · auth · created 2026-09-14 (3 months ago) ·
// verified 2026-09-14 · applies to app/auth/ → memories/bugfixes/<id>.md".
func indexLine(e MemoryEntry, now time.Time) string {
	label := e.Get("type")
	if label == "" {
		label = e.Section
	}
	parts := []string{fmt.Sprintf("[%s] %s", label, e.Title)}
	if st := strings.TrimSpace(e.Get("status")); st != "" && !strings.EqualFold(st, StatusActive) {
		parts = append(parts, st)
	}
	for _, k := range []string{"platform", "area"} {
		if v := e.Get(k); v != "" {
			parts = append(parts, v)
		}
	}
	if d, err := time.Parse("2006-01-02", e.Get("date")); err == nil {
		parts = append(parts, fmt.Sprintf("created %s (%s)", e.Get("date"), relativeAge(d, now)))
	}
	if v := e.Get("verified"); v != "" {
		parts = append(parts, "verified "+v)
	}
	if at := appliesTo(e); len(at) > 0 {
		parts = append(parts, "applies to "+strings.Join(at, ", "))
	}
	return "- " + strings.Join(parts, " · ") + " → " + strings.TrimPrefix(e.Path, ".mobiai/brain/")
}

// appliesTo returns where a note applies: the `paths:` globs when set,
// otherwise the folders of its `### Files`, without repeats. A file at the
// repo root stands for itself, not for the whole repository.
func appliesTo(e MemoryEntry) []string {
	if raw := strings.TrimSpace(e.Get("paths")); raw != "" {
		var out []string
		for _, g := range strings.Split(raw, ",") {
			if g = strings.TrimSpace(g); g != "" {
				out = append(out, g)
			}
		}
		return out
	}
	seen := map[string]bool{}
	var out []string
	for _, f := range e.Files {
		f = path.Clean(strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(f)), "./"))
		item := f
		if dir := path.Dir(f); dir != "." {
			item = dir + "/"
		}
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

// relativeAge renders how long ago d was, in whole days, months (30 days)
// or years (365 days).
func relativeAge(d, now time.Time) string {
	days := int(truncateToDay(now).Sub(truncateToDay(d)).Hours() / 24)
	plural := func(n int, unit string) string {
		if n == 1 {
			return "1 " + unit + " ago"
		}
		return fmt.Sprintf("%d %ss ago", n, unit)
	}
	switch {
	case days <= 0:
		return "today"
	case days < 30:
		return plural(days, "day")
	case days < 365:
		return plural(days/30, "month")
	default:
		return plural(days/365, "year")
	}
}

// NoteView is one note as `show` / `mobile_get` return it.
type NoteView struct {
	ID       string `json:"id"`
	Section  string `json:"section"`
	Path     string `json:"path"` // relative to the project root
	Markdown string `json:"markdown"`
}

// ShowNote returns the full text of the note with the given id: the card
// verbatim, or, for a note still only in an old single-file memory, the
// card it would become.
func ShowNote(p BrainPaths, id string) (NoteView, error) {
	id = strings.TrimSpace(id)
	notes, err := LoadNotes(p)
	if err != nil {
		return NoteView{}, err
	}
	for _, n := range notes {
		if n.Get("id") != id {
			continue
		}
		v := NoteView{ID: id, Section: n.Section, Path: n.Path, Markdown: renderMigratedCard(n)}
		if !n.Legacy {
			data, err := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(n.Path)))
			if err != nil {
				return NoteView{}, fmt.Errorf("read %s: %w", n.Path, err)
			}
			v.Markdown = normalizeNewlines(string(data))
		}
		return v, nil
	}
	return NoteView{}, fmt.Errorf("no entry with id %q found in the brain", id)
}
