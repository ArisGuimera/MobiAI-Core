package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateEntry_ChangesStatus(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`# Bugfixes

## Old workaround

- id: old-wa-20260101-000000
- type: platform_workaround
- status: temporary
- platform: ios
- date: 2026-01-01
- review_after: 2026-06-01

### Problem
The thing breaks.

### Solution
Workaround it.
`)

	res, err := UpdateEntry(p, "old-wa-20260101-000000", UpdateOptions{
		Status: "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrevStatus != "temporary" || res.NewStatus != "active" {
		t.Errorf("status transition: prev=%q new=%q", res.PrevStatus, res.NewStatus)
	}
	if res.Section != "bugfixes" {
		t.Errorf("section = %q, want bugfixes", res.Section)
	}

	// Verify on disk: status updated AND type re-derived (bugfix +
	// active → bug_fix).
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	got := string(data)
	for _, want := range []string{
		"- status: active",
		"- type: bug_fix",
		"### Problem",
		"### Solution",
		"The thing breaks.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q after update:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"- status: temporary",
		"- type: platform_workaround",
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("stale %q still present:\n%s", unwanted, got)
		}
	}
}

func TestUpdateEntry_BumpReviewAfter(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## WA

- id: wa-1
- type: platform_workaround
- status: temporary
- date: 2026-01-01
- review_after: 2026-03-01

body
`)

	res, err := UpdateEntry(p, "wa-1", UpdateOptions{
		ReviewAfter: "2027-01-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrevReviewAfter != "2026-03-01" || res.NewReviewAfter != "2027-01-01" {
		t.Errorf("review_after transition: prev=%q new=%q", res.PrevReviewAfter, res.NewReviewAfter)
	}
	// Status should NOT have changed.
	if res.PrevStatus != res.NewStatus {
		t.Errorf("bump should not touch status; prev=%q new=%q", res.PrevStatus, res.NewStatus)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	if !strings.Contains(string(data), "- review_after: 2027-01-01") {
		t.Errorf("new review_after missing:\n%s", data)
	}
	if strings.Contains(string(data), "- review_after: 2026-03-01") {
		t.Errorf("old review_after still present:\n%s", data)
	}
}

func TestUpdateEntry_AddReviewAfter(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## WA without date

- id: nodate-1
- type: platform_workaround
- status: temporary
- date: 2026-01-01

body
`)
	res, err := UpdateEntry(p, "nodate-1", UpdateOptions{
		ReviewAfter: "2026-09-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrevReviewAfter != "" || res.NewReviewAfter != "2026-09-01" {
		t.Errorf("review_after transition: prev=%q new=%q", res.PrevReviewAfter, res.NewReviewAfter)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	if !strings.Contains(string(data), "- review_after: 2026-09-01") {
		t.Errorf("review_after not added:\n%s", data)
	}
}

func TestUpdateEntry_ClearReviewAfter(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## WA

- id: wa-clr
- type: platform_workaround
- status: temporary
- date: 2026-01-01
- review_after: 2026-06-01

body
`)
	res, err := UpdateEntry(p, "wa-clr", UpdateOptions{
		ClearReviewAfter: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrevReviewAfter != "2026-06-01" || res.NewReviewAfter != "" {
		t.Errorf("review_after transition: prev=%q new=%q", res.PrevReviewAfter, res.NewReviewAfter)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	if strings.Contains(string(data), "review_after") {
		t.Errorf("review_after should be gone:\n%s", data)
	}
}

func TestUpdateEntry_PromoteAndClearAtOnce(t *testing.T) {
	// Promoting a temporary bugfix to active and dropping its
	// review_after in a single call — common workflow when closing
	// out a workaround.
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## WA combo

- id: combo-1
- type: platform_workaround
- status: temporary
- date: 2026-01-01
- review_after: 2026-06-01

body
`)
	res, err := UpdateEntry(p, "combo-1", UpdateOptions{
		Status:           "active",
		ClearReviewAfter: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.NewStatus != "active" || res.NewReviewAfter != "" {
		t.Errorf("combo update failed: status=%q review_after=%q",
			res.NewStatus, res.NewReviewAfter)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	got := string(data)
	if !strings.Contains(got, "- status: active") || !strings.Contains(got, "- type: bug_fix") {
		t.Errorf("status/type not updated:\n%s", got)
	}
	if strings.Contains(got, "review_after") {
		t.Errorf("review_after still present:\n%s", got)
	}
}

func TestUpdateEntry_DecisionDoesNotFlipType(t *testing.T) {
	// architecture_decision should stay architecture_decision regardless
	// of status (the type-flip behavior is bugfix-only).
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		`## Use Koin

- id: koin-1
- type: architecture_decision
- status: active
- date: 2026-01-01

body
`)
	_, err := UpdateEntry(p, "koin-1", UpdateOptions{Status: "deprecated"})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	got := string(data)
	if !strings.Contains(got, "- type: architecture_decision") {
		t.Errorf("decision type should not flip:\n%s", got)
	}
	if !strings.Contains(got, "- status: deprecated") {
		t.Errorf("status not updated:\n%s", got)
	}
}

func TestUpdateEntry_PreservesBodyAndFiles(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		`## Use Koin

- id: k-1
- type: architecture_decision
- status: active
- date: 2026-01-01

### Decision
Use Koin everywhere.

Multiple paragraphs.
With blank lines.

### Files
- composeApp/src/commonMain/di/Module.kt
- iosApp/iosApp/AppDelegate.swift
`)

	_, err := UpdateEntry(p, "k-1", UpdateOptions{Status: "deprecated"})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	got := string(data)
	for _, want := range []string{
		"### Decision",
		"Use Koin everywhere.",
		"Multiple paragraphs.",
		"With blank lines.",
		"### Files",
		"- composeApp/src/commonMain/di/Module.kt",
		"- iosApp/iosApp/AppDelegate.swift",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("body/files corrupted: missing %q:\n%s", want, got)
		}
	}
}

func TestUpdateEntry_FindsAcrossSections(t *testing.T) {
	// Plant the same-id-less but distinct entries in multiple files.
	// Update targets the bugfixes one — decisions must remain untouched.
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		`## Some decision

- id: dec-1
- type: architecture_decision
- status: active

body
`)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Some bug

- id: bug-1
- type: bug_fix
- status: active

body
`)
	res, err := UpdateEntry(p, "bug-1", UpdateOptions{Status: "deprecated"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Section != "bugfixes" {
		t.Errorf("section = %q, want bugfixes", res.Section)
	}
	// decisions.md should be byte-identical.
	dec, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	if !strings.Contains(string(dec), "- status: active") {
		t.Errorf("decisions.md was modified incorrectly:\n%s", dec)
	}
}

func TestUpdateEntry_IDNotFound(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## A

- id: real-id
- status: temporary

body
`)
	_, err := UpdateEntry(p, "ghost-id", UpdateOptions{Status: "active"})
	if err == nil {
		t.Fatal("expected error for unknown id")
	}
	if !strings.Contains(err.Error(), "ghost-id") {
		t.Errorf("error should mention the id; got: %v", err)
	}
}

func TestUpdateEntry_InvalidStatus(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## A

- id: x

body
`)
	_, err := UpdateEntry(p, "x", UpdateOptions{Status: "weird"})
	if err == nil {
		t.Fatal("expected error for bad status")
	}
}

func TestUpdateEntry_InvalidReviewAfter(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## A

- id: x
- status: temporary

body
`)
	_, err := UpdateEntry(p, "x", UpdateOptions{ReviewAfter: "tomorrow"})
	if err == nil {
		t.Fatal("expected error for bad date format")
	}
}

func TestUpdateEntry_ContradictoryFlags(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## A

- id: x
- status: temporary
- review_after: 2026-06-01

body
`)
	_, err := UpdateEntry(p, "x", UpdateOptions{
		ReviewAfter:      "2027-01-01",
		ClearReviewAfter: true,
	})
	if err == nil {
		t.Fatal("expected error: review-after + clear-review-after is contradictory")
	}
}

func TestUpdateEntry_NoChangesRequested(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## A

- id: x
- status: temporary

body
`)
	_, err := UpdateEntry(p, "x", UpdateOptions{})
	if err == nil {
		t.Fatal("expected error when nothing requested")
	}
}

func TestUpdateEntry_BrainNotInitialized(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	_, err := UpdateEntry(p, "x", UpdateOptions{Status: "active"})
	if err == nil {
		t.Fatal("expected error when brain missing")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("error should mention not initialized: %v", err)
	}
}

func TestUpdateEntry_PreservesWrappingBlankLines(t *testing.T) {
	// save.renderEntry produces a blank line both between `## Title`
	// and the first meta, and between the last meta and the body.
	// Updating an entry must keep that exact layout so the file stays
	// visually consistent with freshly saved entries.
	p := initBrainForTest(t)
	original := `## Workaround

- id: wa-fmt
- type: platform_workaround
- status: temporary
- review_after: 2026-06-01

body line 1
body line 2
`
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"), original)
	_, err := UpdateEntry(p, "wa-fmt", UpdateOptions{Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	got := string(data)
	// Title → blank line → first meta.
	if !strings.Contains(got, "## Workaround\n\n- id: wa-fmt") {
		t.Errorf("leading blank line between title and meta lost:\n%s", got)
	}
	// Last meta → blank line → body. The canonical order puts
	// review_after as the last meta line.
	if !strings.Contains(got, "- review_after: 2026-06-01\n\nbody line 1") {
		t.Errorf("trailing blank line between meta and body lost:\n%s", got)
	}
}

func TestUpdateEntry_ReparseAfterUpdate(t *testing.T) {
	// Round-trip: write entry, update, re-parse, verify the parser
	// sees the new values. Catches subtle whitespace/format bugs that
	// silent string asserts miss.
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Round trip

- id: rt-1
- type: platform_workaround
- status: temporary
- platform: ios
- date: 2026-01-01
- review_after: 2026-06-01

body
`)
	_, err := UpdateEntry(p, "rt-1", UpdateOptions{
		Status:      "active",
		ReviewAfter: "2027-01-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := ParseAllMemories(p)
	if err != nil {
		t.Fatal(err)
	}
	var got MemoryEntry
	for _, e := range groups["bugfixes"] {
		if e.Get("id") == "rt-1" {
			got = e
			break
		}
	}
	if got.Get("status") != "active" {
		t.Errorf("re-parsed status = %q, want active", got.Get("status"))
	}
	if got.Get("type") != "bug_fix" {
		t.Errorf("re-parsed type = %q, want bug_fix", got.Get("type"))
	}
	if got.Get("review_after") != "2027-01-01" {
		t.Errorf("re-parsed review_after = %q, want 2027-01-01", got.Get("review_after"))
	}
	if got.Get("platform") != "ios" {
		t.Errorf("platform should be preserved untouched; got %q", got.Get("platform"))
	}
}
