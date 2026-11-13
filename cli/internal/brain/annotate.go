package brain

import "time"

// NoteFlags are computed every time the brain is read. They are never
// written to the cards, so they can't go stale or cause merge conflicts.
type NoteFlags struct {
	// Unconfirmed: written by someone else and not on the main branch yet
	// (see confirmation).
	Unconfirmed bool
	// Warnings: why the note may be out of date, each starting with "⚠"
	// (see checkVerification).
	Warnings []string
	// SupersededBy: id of the confirmed note that replaces this one; its
	// text is no longer delivered (see applySupersedes).
	SupersededBy string
	// ReplacementProposed: id of an unconfirmed note that says it replaces
	// this one; this note stays in force until that one is confirmed.
	ReplacementProposed string
}

// annotation is what annotateNotes learned about the brain as a whole.
type annotation struct {
	Branch string // main branch notes are confirmed against ("" = no git)
}

// annotateNotes fills the computed flags of every note in place.
func annotateNotes(p BrainPaths, cfg Config, notes []MemoryEntry, now time.Time) annotation {
	conf := loadConfirmation(p, cfg)
	for i := range notes {
		notes[i].Flags = NoteFlags{Unconfirmed: !conf.confirmed(notes[i])}
	}
	applySupersedes(notes)
	checkVerification(p.Root, notes, now)
	checkUntil(p.Root, notes)
	return annotation{Branch: conf.branch}
}

// groupBySection buckets notes by canonical section, with an entry (maybe
// empty) for every known section.
func groupBySection(notes []MemoryEntry) map[string][]MemoryEntry {
	out := make(map[string][]MemoryEntry, len(MemoryFiles))
	for _, mf := range MemoryFiles {
		out[canonicalSection(mf.Name)] = nil
	}
	for _, n := range notes {
		out[n.Section] = append(out[n.Section], n)
	}
	return out
}
