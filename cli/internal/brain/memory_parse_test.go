package brain

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMemoryFile_BasicEntry(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "decisions.md")
	mustWrite(t, path, `# Decisions

<!-- header comment -->

## Use Koin as DI

- id: use-koin-as-di-20260510-123045
- type: architecture_decision
- status: active
- platform: shared
- area: dependency_injection
- date: 2026-05-10

### Decision
Use Koin.

### Reason
KMP-friendly.

### Files
- composeApp/src/commonMain/Module.kt
`)

	entries, err := ParseMemoryFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.File != "decisions.md" {
		t.Errorf("File = %q", e.File)
	}
	if e.Title != "Use Koin as DI" {
		t.Errorf("Title = %q", e.Title)
	}
	wantMeta := map[string]string{
		"id":       "use-koin-as-di-20260510-123045",
		"type":     "architecture_decision",
		"status":   "active",
		"platform": "shared",
		"area":     "dependency_injection",
		"date":     "2026-05-10",
	}
	for k, v := range wantMeta {
		if got := e.Meta[k]; got != v {
			t.Errorf("Meta[%q] = %q, want %q", k, got, v)
		}
	}
	if !strings.Contains(e.Body, "### Decision") || !strings.Contains(e.Body, "Use Koin.") {
		t.Errorf("Body missing expected content:\n%s", e.Body)
	}
	// `### Files` list items must NOT be parsed as metadata pairs — the
	// keys would contain `/` and produce garbage. Verify they end up
	// in the Body instead.
	if !strings.Contains(e.Body, "- composeApp/src/commonMain/Module.kt") {
		t.Errorf("Files list should stay in body:\n%s", e.Body)
	}
	if _, leaked := e.Meta["composeApp/src/commonMain/Module.kt"]; leaked {
		t.Error("file path leaked into Meta as a key")
	}
}

func TestParseMemoryFile_MultipleEntries(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bugfixes.md")
	mustWrite(t, path, `# Bugfixes

## First

- id: first-1
- status: active

body of first

## Second

- id: second-2
- status: temporary
- review_after: 2026-09-01

body of second
`)
	entries, err := ParseMemoryFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Title != "First" || entries[1].Title != "Second" {
		t.Errorf("titles = %q, %q", entries[0].Title, entries[1].Title)
	}
	if entries[1].Get("review_after") != "2026-09-01" {
		t.Errorf("review_after for second = %q", entries[1].Get("review_after"))
	}
	if !strings.Contains(entries[0].Body, "body of first") {
		t.Errorf("first.Body = %q", entries[0].Body)
	}
}

func TestParseMemoryFile_MissingFileIsEmpty(t *testing.T) {
	entries, err := ParseMemoryFile(filepath.Join(t.TempDir(), "nope.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseMemoryFile_EntryWithoutMetadata(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "decisions.md")
	mustWrite(t, path, `## Just a title with no meta

Some body content.
`)
	entries, err := ParseMemoryFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Meta) != 0 {
		t.Errorf("Meta should be empty for entry without bullets, got %v", entries[0].Meta)
	}
	if !strings.Contains(entries[0].Body, "Some body content") {
		t.Errorf("Body missing: %q", entries[0].Body)
	}
}

func TestParseAllMemories_GroupsBySection(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		"## Dec1\n\n- id: d1\n- status: active\n\nbody d1\n")
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		"## Bug1\n\n- id: b1\n- status: temporary\n\nbody b1\n")

	groups, err := ParseAllMemories(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups["decisions"]) != 1 || groups["decisions"][0].Title != "Dec1" {
		t.Errorf("decisions = %+v", groups["decisions"])
	}
	if len(groups["bugfixes"]) != 1 || groups["bugfixes"][0].Title != "Bug1" {
		t.Errorf("bugfixes = %+v", groups["bugfixes"])
	}
	// Sections with no entries are still present (empty slice) so the
	// renderer can show "No entries yet" without a map-lookup miss.
	if _, ok := groups["testing"]; !ok {
		t.Error("testing section should exist as empty slice")
	}
}

func TestFileForSection(t *testing.T) {
	cases := map[string]string{
		"decisions":    "decisions.md",
		"  Bugfixes ":  "bugfixes.md",
		"testing":      "testing.md",
		"integrations": "integrations.md",
		"releases":     "releases.md",
		"nope":         "",
		"":             "",
	}
	for in, want := range cases {
		if got := fileForSection(in); got != want {
			t.Errorf("fileForSection(%q) = %q, want %q", in, got, want)
		}
	}
}
