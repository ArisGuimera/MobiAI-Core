package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// formatNoticeID is the id of the entry 0.3 writes into the old
// single-file memories so 0.2.x users are told to update. It is not a note.
const formatNoticeID = "mobiai-brain-format-notice"

// LoadNotes returns every note in the brain: the cards under
// memories/<section>/ plus the entries still in the old single-file
// memories. When the same id exists in both places the card wins. The
// format notice is never returned.
//
// Notes are grouped by section (MemoryFiles order) and sorted by date and
// then id inside each section, so the order never depends on how a merge
// laid out the files.
func LoadNotes(p BrainPaths) ([]MemoryEntry, error) {
	var notes []MemoryEntry
	cardIDs := map[string]bool{}

	for _, mf := range MemoryFiles {
		if mf.Dir == "" {
			continue
		}
		cards, err := loadCardDir(p, mf)
		if err != nil {
			return nil, err
		}
		for _, c := range cards {
			cardIDs[c.Get("id")] = true
		}
		notes = append(notes, cards...)
	}

	legacy, err := loadLegacyNotes(p)
	if err != nil {
		return nil, err
	}
	for _, e := range legacy {
		if !cardIDs[e.Get("id")] {
			notes = append(notes, e)
		}
	}

	order := map[string]int{}
	for i, mf := range MemoryFiles {
		order[canonicalSection(mf.Name)] = i
	}
	sort.SliceStable(notes, func(i, j int) bool {
		a, b := notes[i], notes[j]
		if order[a.Section] != order[b.Section] {
			return order[a.Section] < order[b.Section]
		}
		if a.Get("date") != b.Get("date") {
			return a.Get("date") < b.Get("date")
		}
		return a.Get("id") < b.Get("id")
	})
	return notes, nil
}

// loadCardDir parses every card in memories/<mf.Dir>/. Hidden files
// (temporary files from an interrupted atomic write) and anything that is
// not .md are skipped. A missing directory is an empty section.
func loadCardDir(p BrainPaths, mf MemoryFile) ([]MemoryEntry, error) {
	dir := filepath.Join(p.MemoriesDir, mf.Dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	var cards []MemoryEntry
	for _, de := range entries {
		name := de.Name()
		if de.IsDir() || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			continue
		}
		c, err := ParseCard(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		c.Section = canonicalSection(mf.Name)
		c.Path = ".mobiai/brain/memories/" + mf.Dir + "/" + name
		cards = append(cards, c)
	}
	return cards, nil
}
