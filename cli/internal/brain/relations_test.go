package brain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func save(t *testing.T, p BrainPaths, e *SaveEntry) string {
	t.Helper()
	id, err := AppendEntry(p, e)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestSupersedes_OldShownAsSupersededAndNotActive(t *testing.T) {
	p := initBrainForTest(t)
	oldID := save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Use Hilt for DI", Body: "Old reasoning body."})
	newID := save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Move DI to Koin", Body: "Koin works in KMP.", Supersedes: []string{oldID}})

	card, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions", newID+".md"))
	if !strings.Contains(string(card), "- supersedes: "+oldID+"\n") {
		t.Errorf("the new card must declare supersedes:\n%s", card)
	}

	full := BuildContextWith(DefaultConfig(p.Root), nil, p, ContextOptions{Now: indexNow})
	if !strings.Contains(full, "- state: superseded by "+newID) {
		t.Errorf("full context must mark the old note:\n%s", full)
	}
	if strings.Contains(full, "Old reasoning body.") {
		t.Errorf("a superseded note's text is not delivered:\n%s", full)
	}

	writeCard(t, p, "bugfixes", "big-20260101-000000", decisionCard("Big", "big-20260101-000000", "2026-01-01", "active", 3000))
	cfg := DefaultConfig(p.Root)
	cfg.ContextMaxBytes = 2500
	idx := BuildContextWith(cfg, nil, p, ContextOptions{Now: indexNow})
	active := idx[strings.Index(idx, "## Active Decisions"):strings.Index(idx, "## Index")]
	if strings.Contains(active, "Use Hilt") || !strings.Contains(active, "Move DI to Koin") {
		t.Errorf("only the replacement is an active decision:\n%s", active)
	}
	if !strings.Contains(idx, "- [architecture_decision] Use Hilt for DI · superseded by "+newID+" ·") {
		t.Errorf("index line must say superseded by:\n%s", idx)
	}
}

func TestSupersedes_DoesNotTouchOldCard(t *testing.T) {
	p := initBrainForTest(t)
	oldID := save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Use Hilt for DI", Body: "b"})
	path := filepath.Join(p.MemoriesDir, "decisions", oldID+".md")
	before, _ := os.ReadFile(path)
	save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Move DI to Koin", Body: "b", Supersedes: []string{oldID}})
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("superseding must only write the new card")
	}
}

func TestSupersedes_UnknownIDRejected(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{Type: SaveTypeDecision, Title: "X", Supersedes: []string{"nope"}})
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Errorf("err = %v", err)
	}
}

func TestSupersedes_UnconfirmedReplacementDoesNotRetireOld(t *testing.T) {
	root, p := gitBrain(t)
	oldID := save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Use Hilt for DI", Body: "Hilt body."})
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "old decision on main")
	gitRun(t, root, "checkout", "-qb", "feature")
	newID := save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Move DI to Koin", Body: "b", Supersedes: []string{oldID}})
	commitAs(t, root, "other@example.test", "someone proposes a replacement")

	full := BuildContextWith(DefaultConfig(root), nil, p, ContextOptions{Now: indexNow})
	if strings.Contains(full, "superseded by") || !strings.Contains(full, "Hilt body.") {
		t.Errorf("an unconfirmed replacement must not retire the old note:\n%s", full)
	}
	if !strings.Contains(full, "- state: replacement proposed by "+newID+" (unconfirmed)") {
		t.Errorf("the old note should point at the proposal:\n%s", full)
	}
}

func TestSimilarTitle_BlocksSave(t *testing.T) {
	p := initBrainForTest(t)
	oldID := save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Login crash on screen rotation", Body: "b"})
	_, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Login crash on rotation of the screen", Body: "b"})
	var similar *SimilarNoteError
	if !errors.As(err, &similar) {
		t.Fatalf("err = %v, want SimilarNoteError", err)
	}
	want := "A similar note already exists: " + oldID + " \"Login crash on screen rotation\". Edit that note file, pass --supersedes " + oldID + " to replace it, or pass --new to save a separate note."
	if err.Error() != want {
		t.Errorf("message:\n got: %s\nwant: %s", err.Error(), want)
	}
	cards, _ := filepath.Glob(filepath.Join(p.MemoriesDir, "bugfixes", "*.md"))
	if len(cards) != 1 {
		t.Errorf("the brake must not save anything, %d cards", len(cards))
	}
}

func TestSimilarTitle_NewFlagBypasses(t *testing.T) {
	p := initBrainForTest(t)
	save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Login crash on screen rotation", Body: "b"})
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Login crash on screen rotation", Body: "b", New: true}); err != nil {
		t.Errorf("--new must skip the brake: %v", err)
	}
}

func TestSimilarTitle_SupersedesBypasses(t *testing.T) {
	p := initBrainForTest(t)
	oldID := save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Login crash on screen rotation", Body: "b"})
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Login crash on screen rotation", Body: "b", Supersedes: []string{oldID}}); err != nil {
		t.Errorf("--supersedes must skip the brake: %v", err)
	}
}

func TestSimilarTitle_OtherSectionDeprecatedOrBelowThresholdPass(t *testing.T) {
	p := initBrainForTest(t)
	save(t, p, &SaveEntry{Type: SaveTypeDecision, Title: "Login crash on screen rotation", Body: "b"})
	save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Checkout total ignores discount", Status: StatusDeprecated, Body: "b"})
	for _, e := range []*SaveEntry{
		{Type: SaveTypeBugfix, Title: "Login crash on screen rotation", Body: "other section"},
		{Type: SaveTypeBugfix, Title: "Checkout total ignores the discount", Body: "the similar one is deprecated"},
		{Type: SaveTypeBugfix, Title: "Login button color", Body: "1 of 3 words in common"},
	} {
		if _, err := AppendEntry(p, e); err != nil {
			t.Errorf("%q: %v", e.Title, err)
		}
	}
}

func TestSimilarTitles(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Login crash on screen rotation", "Login crash on rotation of the screen", true},
		{"Use Koin as DI", "Use Hilt as DI", true},
		{"Login crash on rotation", "Login button color", false},
		{"Firebase", "Firebase", true},
		{"Usar Koin para la inyección", "Usar Koin en la inyección de dependencias", true},
		{"Checkout", "Login", false},
	}
	for _, c := range cases {
		if got := similarTitles(c.a, c.b); got != c.want {
			t.Errorf("similarTitles(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestCrash_OnlyForBugfix(t *testing.T) {
	p := initBrainForTest(t)
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeDecision, Title: "X", Crash: []string{"abc"}}); err == nil {
		t.Error("crash on a decision must be rejected")
	}
	id := save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "NPE on login", Body: "b",
		Crash: []string{"6a1b2c3d4e5f", "NullPointerException at LoginViewModel.kt:42"}})
	card, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes", id+".md"))
	if !strings.Contains(string(card), "- crash: 6a1b2c3d4e5f, NullPointerException at LoginViewModel.kt:42\n") {
		t.Errorf("crash field:\n%s", card)
	}
}

func TestSearchCrash_MatchesFieldExactlyAndBodyText(t *testing.T) {
	p := initBrainForTest(t)
	save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "NPE on login", Body: "b", Crash: []string{"6a1b2c3d4e5f"}})
	save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Cart crash", Body: "Sentry fingerprint 9f8e7d was this one."})
	save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Prefix only", Body: "b", Crash: []string{"6a1b2c3d4e5f00"}})

	for crash, want := range map[string]string{
		"6A1B2C3D4E5F": "bugfixes/NPE on login",
		"9f8e7d":       "bugfixes/Cart crash",
	} {
		hits, err := Search(p, "", EntryFilter{Crash: crash})
		if err != nil {
			t.Fatal(err)
		}
		if got := hitTitles(hits); len(got) != 1 || got[0] != want {
			t.Errorf("crash %q → %v, want [%s]", crash, got, want)
		}
	}
}
