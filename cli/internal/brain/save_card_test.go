package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func clockAt(y int, mo time.Month, d, h, mi, s int) func() time.Time {
	return func() time.Time { return time.Date(y, mo, d, h, mi, s, 0, time.UTC) }
}

func TestAppendEntry_WritesCardFile(t *testing.T) {
	p := initBrainForTest(t)
	id, err := AppendEntry(p, &SaveEntry{
		Type:     SaveTypeBugfix,
		Title:    "Login crash on screen rotation",
		Platform: "android",
		Area:     "auth",
		Body:     "The form state lived in the Activity; moved it to the ViewModel.",
		Files:    []string{"app/src/main/java/com/example/auth/LoginViewModel.kt"},
		now:      clockAt(2026, 9, 14, 16, 30, 44),
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != "login-crash-on-screen-rotation-20260914-163044" {
		t.Fatalf("id = %q", id)
	}
	got, err := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes", id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sampleCard {
		t.Errorf("card content mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, sampleCard)
	}
}

func TestAppendEntry_DoesNotTouchLegacyFiles(t *testing.T) {
	p := initBrainForTest(t)
	legacy := filepath.Join(p.MemoriesDir, "decisions.md")
	before, err := os.ReadFile(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeDecision, Title: "Use Koin", Body: "b"}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(legacy)
	if string(before) != string(after) {
		t.Errorf("decisions.md changed after save:\n%s", after)
	}
}

func TestAppendEntry_ConvertsH2InBodyToH3(t *testing.T) {
	p := initBrainForTest(t)
	id, err := AppendEntry(p, &SaveEntry{
		Type:  SaveTypeBugfix,
		Title: "Login crash",
		Body:  "## Symptom\nCrash.\n\n## Root cause\nState in the Activity.\n##NotAHeading stays",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes", id+".md"))
	s := string(got)
	for _, want := range []string{"\n### Symptom\n", "\n### Root cause\n", "\n##NotAHeading stays"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
	if strings.Count(s, "\n## ") != 0 || !strings.HasPrefix(s, "## Login crash\n") {
		t.Errorf("only the title may be an H2:\n%s", s)
	}
}

func TestAppendEntry_KeepsH2InsideCodeFence(t *testing.T) {
	p := initBrainForTest(t)
	body := "### Fix\n```bash\n## not a heading, a shell comment\necho ok\n```\n~~~\n## also code\n~~~\n## Real heading"
	id, err := AppendEntry(p, &SaveEntry{Type: SaveTypeTesting, Title: "Fence", Body: body})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "testing", id+".md"))
	s := string(got)
	for _, want := range []string{"\n## not a heading, a shell comment\n", "\n## also code\n", "\n### Real heading"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
}

func TestAppendEntry_IDCollisionGetsSuffix(t *testing.T) {
	p := initBrainForTest(t)
	now := clockAt(2026, 6, 1, 10, 0, 0)
	id1, err := AppendEntry(p, &SaveEntry{Type: SaveTypeDecision, Title: "Same title", Body: "one", now: now})
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AppendEntry(p, &SaveEntry{Type: SaveTypeDecision, Title: "Same title", Body: "two", New: true, now: now})
	if err != nil {
		t.Fatal(err)
	}
	if id1 != "same-title-20260601-100000" || id2 != "same-title-20260601-100000-2" {
		t.Errorf("ids = %q, %q", id1, id2)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions", id2+".md"))
	if !strings.Contains(string(got), "- id: "+id2+"\n") || !strings.Contains(string(got), "two") {
		t.Errorf("second card wrong:\n%s", got)
	}
}

func TestUpdateEntry_EditsCard(t *testing.T) {
	p := initBrainForTest(t)
	id, err := AppendEntry(p, &SaveEntry{
		Type:   SaveTypeDecision,
		Title:  "Use Koin",
		Status: StatusTemporary,
		Body:   "### Decision\nUse Koin.\n\n## Moved to H3",
		now:    clockAt(2026, 5, 10, 12, 30, 45),
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(p.MemoriesDir, "decisions", id+".md")
	before, _ := os.ReadFile(path)

	res, err := UpdateEntry(p, id, UpdateOptions{Status: StatusDeprecated, ReviewAfter: "2026-12-01"})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrevStatus != StatusTemporary || res.NewStatus != StatusDeprecated {
		t.Errorf("status %q → %q", res.PrevStatus, res.NewStatus)
	}
	if res.Section != "decisions" || res.File != "decisions/"+id+".md" {
		t.Errorf("Section=%q File=%q", res.Section, res.File)
	}
	after, _ := os.ReadFile(path)
	if !strings.Contains(string(after), "- status: deprecated\n") ||
		!strings.Contains(string(after), "- review_after: 2026-12-01\n") {
		t.Errorf("metadata not updated:\n%s", after)
	}
	bodyOf := func(b []byte) string { return ParseCardContent(string(b)).Body }
	if bodyOf(before) != bodyOf(after) {
		t.Errorf("body changed:\n--- before ---\n%s\n--- after ---\n%s", bodyOf(before), bodyOf(after))
	}
	// review_after goes right after verified, matching the canonical order.
	if !strings.Contains(string(after), "- verified: 2026-05-10\n- review_after: 2026-12-01\n") {
		t.Errorf("canonical order broken:\n%s", after)
	}
}

func TestUpdateEntry_BugfixTypeRederivedOnCard(t *testing.T) {
	p := initBrainForTest(t)
	id, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Crash", Body: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateEntry(p, id, UpdateOptions{Status: StatusTemporary}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes", id+".md"))
	if !strings.Contains(string(got), "- type: platform_workaround\n") {
		t.Errorf("type should be re-derived from the section:\n%s", got)
	}
}

func TestUpdateEntry_CardWithCRLFStaysConsistent(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "use-koin-20260510-123045",
		strings.ReplaceAll(cardText("Use Koin", "use-koin-20260510-123045", "2026-05-10", "Body line."), "\n", "\r\n"))
	if _, err := UpdateEntry(p, "use-koin-20260510-123045", UpdateOptions{Status: StatusDeprecated}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions", "use-koin-20260510-123045.md"))
	if strings.Contains(string(got), "\r") {
		t.Errorf("card should be rewritten with LF only, got %q", got)
	}
	if !strings.Contains(string(got), "- status: deprecated\n") || !strings.Contains(string(got), "Body line.") {
		t.Errorf("unexpected card:\n%s", got)
	}
}
