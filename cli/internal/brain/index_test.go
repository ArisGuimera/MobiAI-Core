package brain

import (
	"strings"
	"testing"
	"time"
)

var indexNow = time.Date(2026, 12, 14, 12, 0, 0, 0, time.UTC)

// decisionCard renders a decision card with a body of roughly size bytes.
func decisionCard(title, id, date, status string, size int) string {
	body := strings.Repeat("x", size)
	return "## " + title + "\n\n- id: " + id + "\n- type: architecture_decision\n- status: " + status +
		"\n- date: " + date + "\n- verified: " + date + "\n\n" + body + "\n"
}

func contextFor(t *testing.T, p BrainPaths, maxBytes int, opts ContextOptions) string {
	t.Helper()
	cfg := DefaultConfig(p.Root)
	cfg.ContextMaxBytes = maxBytes
	opts.Now = indexNow
	return BuildContextWith(cfg, nil, p, opts)
}

func TestContext_SmallBrainIsFull(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "use-koin-20260510-123045", decisionCard("Use Koin", "use-koin-20260510-123045", "2026-05-10", "active", 50))
	md := contextFor(t, p, 0, ContextOptions{})
	if strings.Contains(md, "## Index") {
		t.Errorf("a small brain must be delivered in full:\n%s", md)
	}
	if !strings.Contains(md, "## Use Koin") || !strings.Contains(md, strings.Repeat("x", 50)) {
		t.Errorf("full text missing:\n%s", md)
	}
}

func TestContext_DefaultLimitIs25KB(t *testing.T) {
	p := initBrainForTest(t)
	for i, d := range []string{"2026-01-01", "2026-02-01"} {
		id := "big-" + string(rune('a'+i)) + "-20260101-000000"
		writeCard(t, p, "bugfixes", id, decisionCard("Big "+id, id, d, "active", 13000))
	}
	md := contextFor(t, p, 0, ContextOptions{})
	if !strings.Contains(md, "## Index") {
		t.Errorf("26 KB of notes must switch to index mode with the default limit")
	}
	if len(md) > DefaultContextMaxBytes {
		t.Errorf("index-mode output is %d bytes, over the default limit", len(md))
	}
}

func TestContext_LargeBrainSwitchesToIndex(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "crash-a-20260101-000000", decisionCard("Crash A", "crash-a-20260101-000000", "2026-01-01", "active", 3000))
	md := contextFor(t, p, 2000, ContextOptions{})
	for _, want := range []string{
		"## Index",
		"- [architecture_decision] Crash A",
		"→ memories/bugfixes/crash-a-20260101-000000.md",
		"Open the note files you need; use `mobiai brain show <id>` or `mobile_get` if you can't read files.",
		"## Project Rules",
		"## Warnings",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q in:\n%s", want, md)
		}
	}
	if strings.Contains(md, strings.Repeat("x", 3000)) {
		t.Errorf("the bugfix body must not be delivered in index mode")
	}
}

func TestContext_ActiveDecisionsNewestFirstUntilLimit(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "d1-20260101-000000", decisionCard("D1", "d1-20260101-000000", "2026-01-01", "active", 1300))
	writeCard(t, p, "decisions", "d2-20260201-000000", decisionCard("D2", "d2-20260201-000000", "2026-02-01", "active", 1300))
	writeCard(t, p, "decisions", "d3-20260301-000000", decisionCard("D3", "d3-20260301-000000", "2026-03-01", "active", 1300))
	writeCard(t, p, "bugfixes", "big-20260101-000000", decisionCard("Big", "big-20260101-000000", "2026-01-01", "active", 5000))

	md := contextFor(t, p, 4000, ContextOptions{})
	active := md[strings.Index(md, "## Active Decisions"):strings.Index(md, "## Index")]
	i3, i2 := strings.Index(active, "## D3"), strings.Index(active, "## D2")
	if i3 < 0 || i2 < 0 || i3 > i2 {
		t.Errorf("D3 then D2 must be in full, newest first:\n%s", active)
	}
	if strings.Contains(active, "## D1") {
		t.Errorf("D1 does not fit and must go to the index:\n%s", active)
	}
	if !strings.Contains(md, "- [architecture_decision] D1 ·") {
		t.Errorf("D1 missing from the index:\n%s", md)
	}
	if strings.Contains(md, "- [architecture_decision] D3 ·") {
		t.Errorf("D3 is in full and must not be repeated in the index")
	}
	if len(md) > 4000 {
		t.Errorf("output is %d bytes, over the 4000 limit", len(md))
	}
}

func TestContext_DeprecatedOnlyInIndex(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "old-20260101-000000", decisionCard("Old", "old-20260101-000000", "2026-01-01", "deprecated", 100))
	writeCard(t, p, "bugfixes", "big-20260101-000000", decisionCard("Big", "big-20260101-000000", "2026-01-01", "active", 3000))
	md := contextFor(t, p, 2000, ContextOptions{})
	if strings.Contains(md, "## Old") {
		t.Errorf("a deprecated decision must not be in full in index mode:\n%s", md)
	}
	if !strings.Contains(md, "- [architecture_decision] Old · deprecated ·") {
		t.Errorf("deprecated decision missing from the index:\n%s", md)
	}
}

func TestContext_FilteredUnderLimitIsFull(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "big-20260101-000000", decisionCard("Big", "big-20260101-000000", "2026-01-01", "active", 3000))
	ios := "## iOS fix\n\n- id: ios-fix-20260101-000000\n- type: bug_fix\n- status: active\n- platform: ios\n- date: 2026-01-01\n\nSmall body.\n"
	writeCard(t, p, "bugfixes", "ios-fix-20260101-000000", ios)
	md := contextFor(t, p, 2000, ContextOptions{Filter: EntryFilter{Platform: "ios"}})
	if strings.Contains(md, "## Index") || !strings.Contains(md, "Small body.") {
		t.Errorf("a filtered result under the limit must be full:\n%s", md)
	}
}

func TestIndexLine_Format(t *testing.T) {
	e := ParseCardContent(sampleCard)
	e.Section = "bugfixes"
	e.Path = ".mobiai/brain/memories/bugfixes/login-crash-on-screen-rotation-20260914-163044.md"
	got := indexLine(e, indexNow)
	want := "- [bug_fix] Login crash on screen rotation · android · auth · created 2026-09-14 (3 months ago) · verified 2026-09-14 · applies to app/src/main/java/com/example/auth/ → memories/bugfixes/login-crash-on-screen-rotation-20260914-163044.md"
	if got != want {
		t.Errorf("index line:\n got: %s\nwant: %s", got, want)
	}
}

func TestAppliesTo_DerivedFromFilesAndPaths(t *testing.T) {
	e := MemoryEntry{Meta: map[string]string{}, Files: []string{
		"app/src/auth/Login.kt", "app/src/auth/LoginViewModel.kt", "build.gradle.kts", "./app/src/net/Api.kt",
	}}
	if got := strings.Join(appliesTo(e), ", "); got != "app/src/auth/, build.gradle.kts, app/src/net/" {
		t.Errorf("appliesTo = %s", got)
	}
	e.Meta["paths"] = "feature/**/checkout/*.kt, core/net/"
	if got := strings.Join(appliesTo(e), ", "); got != "feature/**/checkout/*.kt, core/net/" {
		t.Errorf("paths: must replace the derived folders, got %s", got)
	}
}

func TestRelativeAge(t *testing.T) {
	cases := map[string]string{
		"2026-12-14": "today",
		"2026-12-13": "1 day ago",
		"2026-12-01": "13 days ago",
		"2026-11-10": "1 month ago",
		"2026-09-14": "3 months ago",
		"2025-12-14": "1 year ago",
		"2023-01-01": "3 years ago",
	}
	for date, want := range cases {
		d, _ := time.Parse("2006-01-02", date)
		if got := relativeAge(d, indexNow); got != want {
			t.Errorf("relativeAge(%s) = %q, want %q", date, got, want)
		}
	}
}

func TestFindNote(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "use-koin-20260510-123045", decisionCard("Use Koin", "use-koin-20260510-123045", "2026-05-10", "active", 10))
	got, err := ShowNote(p, "use-koin-20260510-123045")
	if err != nil {
		t.Fatal(err)
	}
	if got.Markdown != decisionCard("Use Koin", "use-koin-20260510-123045", "2026-05-10", "active", 10) {
		t.Errorf("ShowNote must return the card verbatim, got:\n%s", got.Markdown)
	}
	if got.Path != ".mobiai/brain/memories/decisions/use-koin-20260510-123045.md" || got.Section != "decisions" {
		t.Errorf("Path=%q Section=%q", got.Path, got.Section)
	}
	if _, err := ShowNote(p, "nope"); err == nil || !strings.Contains(err.Error(), "no entry with id") {
		t.Errorf("unknown id error = %v", err)
	}
}
