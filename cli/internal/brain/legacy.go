package brain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// legacyBlock is one `## ` section of an old single-file memory, before
// C5 repair.
type legacyBlock struct {
	title   string
	meta    map[string]string
	hasMeta bool
	lines   []string // every line after the heading
	body    []string // lines after the metadata block
}

// parseLegacyContent parses an old single-file memory the way 0.3 reads
// it. Unlike the 0.2 parser (ParseMemoryFile), a `## ` heading without a
// metadata block that follows a real note is a heading of that note's body
// that 0.2 split off by mistake (C5): it is glued back as `### `.
func parseLegacyContent(content string) []MemoryEntry {
	var blocks []*legacyBlock
	var cur *legacyBlock
	for _, line := range strings.Split(normalizeNewlines(content), "\n") {
		if strings.HasPrefix(line, "## ") {
			cur = &legacyBlock{title: strings.TrimSpace(strings.TrimPrefix(line, "## "))}
			blocks = append(blocks, cur)
			continue
		}
		if cur != nil {
			cur.lines = append(cur.lines, line)
		}
	}

	var notes []*legacyBlock
	for _, b := range blocks {
		b.meta = map[string]string{}
		i := 0
		for ; i < len(b.lines); i++ {
			if k, v, ok := parseMetaLine(b.lines[i]); ok {
				b.meta[k] = v
				b.hasMeta = true
				continue
			}
			if strings.TrimSpace(b.lines[i]) == "" {
				continue
			}
			break
		}
		b.body = b.lines[i:]

		if !b.hasMeta && len(notes) > 0 && notes[len(notes)-1].hasMeta {
			prev := notes[len(notes)-1]
			prev.body = append(prev.body, "### "+b.title)
			prev.body = append(prev.body, b.lines...)
			continue
		}
		notes = append(notes, b)
	}

	out := make([]MemoryEntry, 0, len(notes))
	for _, b := range notes {
		out = append(out, MemoryEntry{
			Title: b.title,
			Meta:  b.meta,
			Body:  strings.TrimSpace(strings.Join(b.body, "\n")),
		})
	}
	return out
}

// loadLegacyNotes reads every old single-file memory with the 0.3 rules:
// C5 repair, a content-derived id for notes written by hand without one,
// and no lost note when two entries share an id (identical copies collapse,
// a different one gets "-<hash>" appended). The format notice is skipped.
func loadLegacyNotes(p BrainPaths) ([]MemoryEntry, error) {
	var notes []MemoryEntry
	seen := map[string]string{} // id → rendered card
	for _, mf := range MemoryFiles {
		data, err := os.ReadFile(filepath.Join(p.MemoriesDir, mf.Name))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", mf.Name, err)
		}
		for _, e := range parseLegacyContent(string(data)) {
			if e.Get("id") == formatNoticeID {
				continue
			}
			e.File = mf.Name
			e.Section = canonicalSection(mf.Name)
			e.Path = ".mobiai/brain/memories/" + mf.Name
			e.Legacy = true
			e.Files = parseFilesSection(e.Body)
			if e.Get("id") == "" {
				e.Meta["id"] = contentID(e)
			}

			card := renderMigratedCard(e)
			id := e.Get("id")
			if prev, ok := seen[id]; ok {
				if prev == card {
					continue
				}
				e.Meta["id"] = id + "-" + shortHash(card)
				card = renderMigratedCard(e)
				id = e.Get("id")
				if _, dup := seen[id]; dup {
					continue
				}
			}
			seen[id] = card
			notes = append(notes, e)
		}
	}
	return notes, nil
}

// contentID builds an id from the note's own text, so every machine that
// migrates the same note picks the same id: slug(title)-YYYYMMDD-<hash>,
// without the date part when the note has no valid date.
func contentID(e MemoryEntry) string {
	slug := slugify(e.Title)
	if slug == "" {
		slug = "entry"
	}
	keys := make([]string, 0, len(e.Meta))
	for k := range e.Meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(e.Title + "\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %s\n", k, e.Meta[k])
	}
	b.WriteString("\n" + e.Body)
	hash := shortHash(b.String())
	if d, err := time.Parse("2006-01-02", e.Get("date")); err == nil {
		return fmt.Sprintf("%s-%s-%s", slug, d.Format("20060102"), hash)
	}
	return slug + "-" + hash
}

// shortHash returns the first 8 hex characters of the SHA-256 of s.
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

// renderMigratedCard renders an old entry as a card. It depends only on
// the entry's text (never on the clock), so two people migrating on two
// branches write byte-identical files. `verified` starts at `date`.
func renderMigratedCard(e MemoryEntry) string {
	meta := make(map[string]string, len(e.Meta)+1)
	for k, v := range e.Meta {
		meta[k] = v
	}
	if meta["verified"] == "" && meta["date"] != "" {
		meta["verified"] = meta["date"]
	}
	var b strings.Builder
	b.WriteString("## " + e.Title + "\n\n")
	for _, k := range metaOrder(meta) {
		fmt.Fprintf(&b, "- %s: %s\n", k, meta[k])
	}
	if body := strings.TrimSpace(e.Body); body != "" {
		b.WriteString("\n" + body + "\n")
	}
	return b.String()
}
