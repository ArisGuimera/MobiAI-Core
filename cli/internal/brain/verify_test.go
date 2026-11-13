package brain

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// commitAt commits everything with author and committer date set to day.
func commitAt(t *testing.T, root, day, msg string) {
	t.Helper()
	gitRun(t, root, "add", "-A")
	cmd := exec.Command("git", "commit", "-qm", msg)
	cmd.Dir = root
	stamp := day + "T12:00:00+00:00"
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_DATE="+stamp, "GIT_COMMITTER_DATE="+stamp)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}
}

// cardWithFiles renders a bugfix card verified on `verified` that lists files.
func cardWithFiles(id, verified string, files ...string) string {
	var b strings.Builder
	b.WriteString("## Login crash\n\n- id: " + id + "\n- type: bug_fix\n- status: active\n- date: 2026-09-01\n- verified: " + verified + "\n\nFixed.\n\n### Files\n")
	for _, f := range files {
		b.WriteString("- " + f + "\n")
	}
	return b.String()
}

func warningsOf(t *testing.T, p BrainPaths, now time.Time) map[string][]string {
	t.Helper()
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	annotateNotes(p, DefaultConfig(p.Root), notes, now)
	out := map[string][]string{}
	for _, n := range notes {
		out[n.Get("id")] = n.Flags.Warnings
	}
	return out
}

func TestStale_FileChangedAfterVerified(t *testing.T) {
	root, p := gitBrain(t)
	mustWrite(t, filepath.Join(root, "app", "LoginViewModel.kt"), "class A\n")
	commitAt(t, root, "2026-09-10", "add vm")
	writeCard(t, p, "bugfixes", "login-crash-20260901-000000", cardWithFiles("login-crash-20260901-000000", "2026-09-14", "app/LoginViewModel.kt"))
	commitAt(t, root, "2026-09-14", "note")

	if w := warningsOf(t, p, indexNow)["login-crash-20260901-000000"]; len(w) != 0 {
		t.Fatalf("nothing changed after verification yet, got %v", w)
	}
	mustWrite(t, filepath.Join(root, "app", "LoginViewModel.kt"), "class B\n")
	commitAt(t, root, "2026-10-01", "change vm")
	w := warningsOf(t, p, indexNow)["login-crash-20260901-000000"]
	if len(w) != 1 || w[0] != "⚠ LoginViewModel.kt changed since verification" {
		t.Errorf("warnings = %v", w)
	}
}

func TestStale_SameDayCommitNotFlagged(t *testing.T) {
	root, p := gitBrain(t)
	mustWrite(t, filepath.Join(root, "app", "LoginViewModel.kt"), "class A\n")
	commitAt(t, root, "2026-09-14", "fix and note on the same day")
	writeCard(t, p, "bugfixes", "login-crash-20260901-000000", cardWithFiles("login-crash-20260901-000000", "2026-09-14", "app/LoginViewModel.kt"))
	if w := warningsOf(t, p, indexNow)["login-crash-20260901-000000"]; len(w) != 0 {
		t.Errorf("a commit on the verification day is not a change after it, got %v", w)
	}
}

func TestStale_FileDeleted(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "login-crash-20260901-000000", cardWithFiles("login-crash-20260901-000000", "2026-12-01", "app/LoginViewModel.kt"))
	w := warningsOf(t, p, indexNow)["login-crash-20260901-000000"]
	if len(w) != 1 || w[0] != "⚠ LoginViewModel.kt no longer exists" {
		t.Errorf("warnings = %v", w)
	}
}

func TestStale_NoFilesThreeMonthRule(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "old-20260801-000000", decisionCard("Old", "old-20260801-000000", "2026-08-01", "active", 10))
	writeCard(t, p, "decisions", "recent-20261001-000000", decisionCard("Recent", "recent-20261001-000000", "2026-10-01", "active", 10))
	writeCard(t, p, "decisions", "retired-20260101-000000", decisionCard("Retired", "retired-20260101-000000", "2026-01-01", "deprecated", 10))
	w := warningsOf(t, p, indexNow)
	if got := w["old-20260801-000000"]; len(got) != 1 || got[0] != "⚠ not verified in 3+ months" {
		t.Errorf("old note warnings = %v", got)
	}
	if got := w["recent-20261001-000000"]; len(got) != 0 {
		t.Errorf("a note verified 2.5 months ago is fine, got %v", got)
	}
	if got := w["retired-20260101-000000"]; len(got) != 0 {
		t.Errorf("deprecated notes are not checked, got %v", got)
	}
}

func TestStale_NoDateNeverVerified(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "hand-written-abcd1234", "## Hand written\n\n- status: active\n\nNo dates.\n")
	if got := warningsOf(t, p, indexNow)["hand-written-abcd1234"]; len(got) != 1 || got[0] != "⚠ never verified" {
		t.Errorf("warnings = %v", got)
	}
}

func TestStale_SingleGitLogCall(t *testing.T) {
	root, p := gitBrain(t)
	for _, f := range []string{"a/A.kt", "b/B.kt", "c/C.kt"} {
		mustWrite(t, filepath.Join(root, filepath.FromSlash(f)), "x\n")
	}
	commitAt(t, root, "2026-09-10", "files")
	writeCard(t, p, "bugfixes", "a-20260901-000000", cardWithFiles("a-20260901-000000", "2026-09-14", "a/A.kt"))
	writeCard(t, p, "bugfixes", "b-20260901-000000", cardWithFiles("b-20260901-000000", "2026-09-14", "b/B.kt"))
	writeCard(t, p, "bugfixes", "c-20260901-000000", cardWithFiles("c-20260901-000000", "2026-09-14", "c/C.kt", "a/A.kt"))

	orig := runGit
	defer func() { runGit = orig }()
	calls := 0
	runGit = func(dir, stdin string, args ...string) (string, error) {
		if strings.Contains(strings.Join(args, " "), verifyLogFormat) {
			calls++
		}
		return orig(dir, stdin, args...)
	}
	warningsOf(t, p, indexNow)
	if calls != 1 {
		t.Errorf("verification must use one git log for the whole brain, used %d", calls)
	}
}

func TestIndexLine_ShowsWarnings(t *testing.T) {
	e := ParseCardContent(sampleCard)
	e.Section = "bugfixes"
	e.Path = ".mobiai/brain/memories/bugfixes/login-crash-on-screen-rotation-20260914-163044.md"
	e.Flags.Warnings = []string{"⚠ LoginViewModel.kt changed since verification"}
	got := indexLine(e, indexNow)
	if !strings.Contains(got, "· verified 2026-09-14 · ⚠ LoginViewModel.kt changed since verification · applies to app/") {
		t.Errorf("index line: %s", got)
	}
}

func TestContext_FullModeShowsWarnings(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "login-crash-20260901-000000", cardWithFiles("login-crash-20260901-000000", "2026-12-01", "app/LoginViewModel.kt"))
	md := BuildContextWith(DefaultConfig(p.Root), nil, p, ContextOptions{Now: indexNow})
	if !strings.Contains(md, "- check: ⚠ LoginViewModel.kt no longer exists\n") {
		t.Errorf("full context must show the warning:\n%s", md)
	}
}

func TestVerify_SetsTodayAndOptionallyFiles(t *testing.T) {
	p := initBrainForTest(t)
	id := "login-crash-20260901-000000"
	writeCard(t, p, "bugfixes", id, cardWithFiles(id, "2026-09-14", "app/Old.kt"))

	res, err := VerifyEntry(p, id, nil, indexNow)
	if err != nil {
		t.Fatal(err)
	}
	if res.PrevVerified != "2026-09-14" || res.NewVerified != "2026-12-14" {
		t.Errorf("verified %q → %q", res.PrevVerified, res.NewVerified)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes", id+".md"))
	if !strings.Contains(string(got), "- verified: 2026-12-14\n") || !strings.Contains(string(got), "- app/Old.kt\n") {
		t.Errorf("after verify:\n%s", got)
	}

	if _, err := VerifyEntry(p, id, []string{"app/New.kt", "app/Other.kt"}, indexNow); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes", id+".md"))
	s := string(got)
	if strings.Contains(s, "Old.kt") || !strings.Contains(s, "### Files\n- app/New.kt\n- app/Other.kt\n") {
		t.Errorf("--files must replace the Files section:\n%s", s)
	}
	if !strings.Contains(s, "Fixed.") {
		t.Errorf("the body must stay:\n%s", s)
	}
}

func TestVerify_UnknownID(t *testing.T) {
	p := initBrainForTest(t)
	if _, err := VerifyEntry(p, "nope", nil, indexNow); err == nil || !strings.Contains(err.Error(), "no entry with id") {
		t.Errorf("err = %v", err)
	}
}

func TestStaleNotes_ForReview(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "decisions", "old-20260801-000000", decisionCard("Old", "old-20260801-000000", "2026-08-01", "active", 10))
	writeCard(t, p, "decisions", "recent-20261201-000000", decisionCard("Recent", "recent-20261201-000000", "2026-12-01", "active", 10))
	stale, err := StaleNotes(p, DefaultConfig(p.Root), indexNow)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 1 || stale[0].Get("id") != "old-20260801-000000" {
		t.Errorf("stale = %v", titles(stale))
	}
}
