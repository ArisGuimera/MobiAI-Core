package brain

import (
	"path/filepath"
	"testing"
)

func TestSearch_MatchesTitleAndBody(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		"## Use Koin as DI\n\n- id: d1\n- status: active\n\nKoin is KMP-friendly.\n")
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		"## FirebaseAuth iOS module renaming\n\n- id: b1\n- status: temporary\n- platform: ios\n\n### Solution\nKeep composeApp lowercase.\n")

	hits, err := Search(p, "koin", EntryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Entry.Title != "Use Koin as DI" {
		t.Fatalf("expected 1 hit for koin, got %d: %v", len(hits), hitTitles(hits))
	}

	hits, _ = Search(p, "firebaseauth", EntryFilter{})
	if len(hits) != 1 || hits[0].Section != "bugfixes" {
		t.Errorf("firebaseauth should hit bugfixes, got: %v", hits)
	}
}

func TestSearch_FilterANDsWithQuery(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Firebase crash on iOS

- id: a1
- status: temporary
- platform: ios

solution body

## Firebase crash on Android

- id: a2
- status: active
- platform: android

solution body
`)
	// Query matches both, but filter narrows to iOS only.
	hits, err := Search(p, "firebase", EntryFilter{Platform: "ios"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Entry.Get("platform") != "ios" {
		t.Errorf("expected 1 ios hit, got %d: %v", len(hits), hitTitles(hits))
	}
}

func TestSearch_NoMatchReturnsEmpty(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		"## X\n\n- id: x\n\nbody\n")
	hits, err := Search(p, "nonexistent-query", EntryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(hits))
	}
}

func TestSearch_SnippetPrefersBodyLineWithQuery(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		"## Generic title\n\n- id: x\n\nFirst line of body.\nLine mentioning koin specifically.\nThird line.\n")
	hits, err := Search(p, "koin", EntryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("got %d hits", len(hits))
	}
	if hits[0].Snippet != "Line mentioning koin specifically." {
		t.Errorf("snippet = %q", hits[0].Snippet)
	}
}

func TestSearch_PreservesSectionOrder(t *testing.T) {
	p := initBrainForTest(t)
	// Plant a "foo" match in both bugfixes and decisions. Search should
	// return decisions first (canonical order: decisions → bugfixes …).
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		"## foo bug\n\n- id: b\n\nfoo\n")
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		"## foo decision\n\n- id: d\n\nfoo\n")
	hits, err := Search(p, "foo", EntryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 || hits[0].Section != "decisions" || hits[1].Section != "bugfixes" {
		t.Errorf("section order broken: got %v", hitTitles(hits))
	}
}

func hitTitles(hits []SearchHit) []string {
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.Section + "/" + h.Entry.Title
	}
	return out
}
