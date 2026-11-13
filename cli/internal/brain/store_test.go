package brain

import (
	"path/filepath"
	"strings"
	"testing"
)

// writeCard drops a card file under memories/<dir>/<id>.md.
func writeCard(t *testing.T, p BrainPaths, dir, id, content string) {
	t.Helper()
	mustWrite(t, filepath.Join(p.MemoriesDir, dir, id+".md"), content)
}

// cardText renders a minimal card for tests.
func cardText(title, id, date, body string) string {
	return "## " + title + "\n\n- id: " + id + "\n- type: architecture_decision\n- status: active\n- date: " + date + "\n- verified: " + date + "\n\n" + body + "\n"
}

func TestLoadNotes_ReadsCardsAndLegacy(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "use-koin-20260510-123045",
		cardText("Use Koin", "use-koin-20260510-123045", "2026-05-10", "Koin body."))
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"), `# Bugfixes

## Old crash

- id: old-crash-20260101-101010
- type: bug_fix
- status: active
- date: 2026-01-01

Old body.
`)

	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d: %v", len(notes), titles(notes))
	}
	byID := map[string]MemoryEntry{}
	for _, n := range notes {
		byID[n.Get("id")] = n
	}
	card := byID["use-koin-20260510-123045"]
	if card.Section != "decisions" || card.Legacy {
		t.Errorf("card: Section=%q Legacy=%v", card.Section, card.Legacy)
	}
	if card.Path != ".mobiai/brain/memories/decisions/use-koin-20260510-123045.md" {
		t.Errorf("card Path = %q", card.Path)
	}
	old := byID["old-crash-20260101-101010"]
	if old.Section != "bugfixes" || !old.Legacy {
		t.Errorf("legacy: Section=%q Legacy=%v", old.Section, old.Legacy)
	}
	if old.Path != ".mobiai/brain/memories/bugfixes.md" {
		t.Errorf("legacy Path = %q", old.Path)
	}
}

func TestLoadNotes_CardWinsOverLegacySameID(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"), `# Decisions

## Use Koin

- id: use-koin-20260510-123045
- type: architecture_decision
- status: active
- date: 2026-05-10

Legacy text.
`)
	writeCard(t, p, "decisions", "use-koin-20260510-123045",
		cardText("Use Koin", "use-koin-20260510-123045", "2026-05-10", "Card text."))

	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected the note once, got %d", len(notes))
	}
	if !strings.Contains(notes[0].Body, "Card text.") || notes[0].Legacy {
		t.Errorf("the card must win; got Legacy=%v body=%q", notes[0].Legacy, notes[0].Body)
	}
}

func TestLoadNotes_SkipsFormatNotice(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"), `# Decisions

## MobiAI Brain moved to a new format: update mobiai

- id: mobiai-brain-format-notice
- type: notice
- status: temporary
- review_after: 2000-01-01

Notice text.

## Real decision

- id: real-decision-20260510-123045
- status: active
- date: 2026-05-10

Body.
`)
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := titles(notes); len(got) != 1 || got[0] != "Real decision" {
		t.Errorf("notes = %v, want only the real decision", got)
	}
}

func TestLoadNotes_SortedByDateThenID(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "b-note-20260601-000000", cardText("B", "b-note-20260601-000000", "2026-06-01", "b"))
	writeCard(t, p, "decisions", "a-note-20260601-000000", cardText("A", "a-note-20260601-000000", "2026-06-01", "a"))
	writeCard(t, p, "decisions", "z-note-20260101-000000", cardText("Z", "z-note-20260101-000000", "2026-01-01", "z"))

	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(titles(notes), ",")
	if got != "Z,A,B" {
		t.Errorf("order = %s, want Z,A,B (date, then id)", got)
	}
}

func TestLoadNotes_IgnoresTempAndNonMarkdown(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "use-koin-20260510-123045",
		cardText("Use Koin", "use-koin-20260510-123045", "2026-05-10", "Koin."))
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions", ".update-123.tmp"), "## Half written\n")
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions", "notes.txt"), "## Not a card\n")
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions", ".hidden.md"), "## Hidden\n")

	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := titles(notes); len(got) != 1 || got[0] != "Use Koin" {
		t.Errorf("notes = %v, want only Use Koin", got)
	}
}

func TestParseAllMemories_IncludesCards(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "crash-20260601-000000",
		cardText("Crash", "crash-20260601-000000", "2026-06-01", "Fixed."))
	groups, err := ParseAllMemories(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := titles(groups["bugfixes"]); len(got) != 1 || got[0] != "Crash" {
		t.Errorf("bugfixes = %v, want [Crash]", got)
	}
}
