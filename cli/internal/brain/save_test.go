package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendEntry_FailsIfBrainNotInitialized(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	_, err := AppendEntry(p, &SaveEntry{
		Type:  SaveTypeDecision,
		Title: "Use Koin",
	})
	if err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected 'not initialized' error, got: %v", err)
	}
}

func TestAppendEntry_DecisionWritesAllFields(t *testing.T) {
	p := initBrainForTest(t)
	frozenNow := func() time.Time {
		return time.Date(2026, 5, 10, 12, 30, 45, 0, time.UTC)
	}
	id, err := AppendEntry(p, &SaveEntry{
		Type:     SaveTypeDecision,
		Title:    "Use Koin as DI",
		Platform: "shared",
		Area:     "dependency_injection",
		Status:   StatusActive,
		Body:     "### Decision\nUse Koin.\n\n### Reason\nKMP-friendly.",
		Files:    []string{"composeApp/src/commonMain/Module.kt"},
		now:      frozenNow,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantID := "use-koin-as-di-20260510-123045"
	if id != wantID {
		t.Errorf("id = %q, want %q", id, wantID)
	}

	got, err := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Use Koin as DI",
		"- id: " + wantID,
		"- type: architecture_decision",
		"- status: active",
		"- platform: shared",
		"- area: dependency_injection",
		"- date: 2026-05-10",
		"### Decision",
		"### Files",
		"- composeApp/src/commonMain/Module.kt",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestAppendEntry_TemporaryBugfixUsesPlatformWorkaround(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{
		Type:        SaveTypeBugfix,
		Title:       "FirebaseAuth iOS module name",
		Platform:    "ios",
		Status:      StatusTemporary,
		ReviewAfter: "2026-09-01",
		Body:        "### Problem\n...\n\n### Solution\n...",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- type: platform_workaround") {
		t.Errorf("temporary bugfix should map to platform_workaround:\n%s", got)
	}
	if !strings.Contains(string(got), "- review_after: 2026-09-01") {
		t.Errorf("review_after missing:\n%s", got)
	}
}

func TestAppendEntry_PermanentBugfixUsesBugFix(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{
		Type:   SaveTypeBugfix,
		Title:  "Crash on empty cart",
		Status: StatusActive,
		Body:   "fixed null deref",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	if !strings.Contains(string(got), "- type: bug_fix") {
		t.Errorf("active bugfix should map to bug_fix:\n%s", got)
	}
}

func TestAppendEntry_TestingDefaultsStatusToActive(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{
		Type:  SaveTypeTesting,
		Title: "DataStore clear waits for empty emission",
		Body:  "use first { ... }",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "testing.md"))
	if !strings.Contains(string(got), "- status: active") {
		t.Errorf("default status should be active:\n%s", got)
	}
	if !strings.Contains(string(got), "- type: testing_pattern") {
		t.Errorf("testing type missing:\n%s", got)
	}
}

func TestAppendEntry_AppendsDoNotOverwrite(t *testing.T) {
	p := initBrainForTest(t)
	for _, title := range []string{"First decision", "Second decision"} {
		if _, err := AppendEntry(p, &SaveEntry{
			Type:  SaveTypeDecision,
			Title: title,
			Body:  "body for " + title,
			now:   func() time.Time { return time.Now() },
		}); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	for _, want := range []string{"## First decision", "## Second decision"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q after two saves; file:\n%s", want, got)
		}
	}
}

func TestAppendEntry_RejectsInvalidStatus(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{
		Type:   SaveTypeDecision,
		Title:  "x",
		Status: "wat",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Errorf("expected status error, got: %v", err)
	}
}

func TestAppendEntry_RejectsBadReviewAfter(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{
		Type:        SaveTypeBugfix,
		Title:       "x",
		Status:      StatusTemporary,
		ReviewAfter: "tomorrow",
	})
	if err == nil || !strings.Contains(err.Error(), "review-after") {
		t.Errorf("expected review-after error, got: %v", err)
	}
}

func TestAppendEntry_RequiresTitle(t *testing.T) {
	p := initBrainForTest(t)
	_, err := AppendEntry(p, &SaveEntry{
		Type:  SaveTypeDecision,
		Title: "   ",
	})
	if err == nil || !strings.Contains(err.Error(), "title") {
		t.Errorf("expected title error, got: %v", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Use Koin as DI":                     "use-koin-as-di",
		"  Trim Spaces  ":                    "trim-spaces",
		"FirebaseAuth iOS — module renaming": "firebaseauth-ios-module-renaming",
		"!!!":                                "",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

// initBrainForTest creates a fully initialized brain in a TempDir and
// returns its paths. Used to avoid the brain-not-initialized guard in
// tests that exercise append behavior.
func initBrainForTest(t *testing.T) BrainPaths {
	t.Helper()
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig(tmp)
	if err := cfg.Save(p); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureMemoryFiles(p); err != nil {
		t.Fatal(err)
	}
	return p
}
