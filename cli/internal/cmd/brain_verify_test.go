package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeCardFile(t *testing.T, root, dir, id, content string) {
	t.Helper()
	path := filepath.Join(root, ".mobiai", "brain", "memories", dir, id+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const oldVerifiedCard = "## Use Koin\n\n- id: use-koin-20260101-000000\n- type: architecture_decision\n- status: active\n- date: 2026-01-01\n- verified: 2026-01-01\n\nUse Koin.\n"

func TestBrainVerify_UpdatesCard(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	writeCardFile(t, root, "decisions", "use-koin-20260101-000000", oldVerifiedCard)

	out := runBrain(t, []string{"verify", "use-koin-20260101-000000", "--root", root, "--files", "composeApp/src/Module.kt"})
	today := time.Now().UTC().Format("2006-01-02")
	if !strings.Contains(out, "verified: 2026-01-01 → "+today) {
		t.Errorf("verify output:\n%s", out)
	}
	got, err := readSectionCards(root, "decisions")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- verified: "+today) || !strings.Contains(string(got), "### Files\n- composeApp/src/Module.kt") {
		t.Errorf("card after verify:\n%s", got)
	}
}

func TestBrainReview_ListsStaleNotesWithoutFailing(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	writeCardFile(t, root, "decisions", "use-koin-20260101-000000", oldVerifiedCard)

	// No --no-fail: stale notes alone must not make review exit 1 (that
	// would kill this test process).
	out := runBrain(t, []string{"review", "--root", root})
	for _, want := range []string{"may be out of date", "Use Koin", "⚠ not verified in 3+ months", "use-koin-20260101-000000"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in review output:\n%s", want, out)
		}
	}
}
