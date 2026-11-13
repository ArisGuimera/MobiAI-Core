package brain

import (
	"strings"
	"testing"
)

// gitBrain returns an initialized brain at the root of a fresh git repo
// whose first commit (by "me") is on main.
func gitBrain(t *testing.T) (string, BrainPaths) {
	t.Helper()
	root := gitRepo(t)
	p := NewBrainPaths(root)
	cfg := DefaultConfig(root)
	if err := cfg.Save(p); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureMemoryFiles(p); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "init brain")
	return root, p
}

// commitAs commits everything with the given author email.
func commitAs(t *testing.T, root, email, msg string) {
	t.Helper()
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.email="+email, "-c", "user.name=someone", "commit", "-qm", msg)
}

func saveDecision(t *testing.T, p BrainPaths, title string) string {
	t.Helper()
	id, err := AppendEntry(p, &SaveEntry{Type: SaveTypeDecision, Title: title, Body: title + " body."})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func unconfirmedIDs(t *testing.T, p BrainPaths) map[string]bool {
	t.Helper()
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	annotateNotes(p, DefaultConfig(p.Root), notes, indexNow)
	out := map[string]bool{}
	for _, n := range notes {
		if n.Flags.Unconfirmed {
			out[n.Get("id")] = true
		}
	}
	return out
}

func TestConfirm_NoGitEverythingConfirmed(t *testing.T) {
	p := initBrainForTest(t)
	saveDecision(t, p, "Use Koin")
	if got := unconfirmedIDs(t, p); len(got) != 0 {
		t.Errorf("without git everything is confirmed, got %v", got)
	}
}

func TestConfirm_CardOnMainIsConfirmed(t *testing.T) {
	root, p := gitBrain(t)
	id := saveDecision(t, p, "Use Koin")
	commitAs(t, root, "other@example.test", "other's note, reviewed and merged")
	gitRun(t, root, "checkout", "-qb", "feature")
	if got := unconfirmedIDs(t, p); got[id] {
		t.Errorf("a card on main is confirmed")
	}
}

func TestConfirm_UncommittedCardIsOwn(t *testing.T) {
	root, p := gitBrain(t)
	gitRun(t, root, "checkout", "-qb", "feature")
	id := saveDecision(t, p, "Use Koin")
	if got := unconfirmedIDs(t, p); got[id] {
		t.Errorf("a card just saved here counts as confirmed for its author")
	}
}

func TestConfirm_CommittedByMeOnBranchIsConfirmed(t *testing.T) {
	root, p := gitBrain(t)
	gitRun(t, root, "checkout", "-qb", "feature")
	id := saveDecision(t, p, "Use Koin")
	commitAs(t, root, "Me@Example.test", "mine")
	if got := unconfirmedIDs(t, p); got[id] {
		t.Errorf("my own committed card is confirmed for me (email compared case-insensitively)")
	}
}

func TestConfirm_CommittedByOtherOnBranchIsUnconfirmed(t *testing.T) {
	root, p := gitBrain(t)
	gitRun(t, root, "checkout", "-qb", "feature")
	id := saveDecision(t, p, "Use Koin")
	commitAs(t, root, "other@example.test", "someone else's note")
	if got := unconfirmedIDs(t, p); !got[id] {
		t.Errorf("another person's card that is not on main must be unconfirmed")
	}
}

func TestConfirm_MigratedNoteOnMainLegacyIsConfirmed(t *testing.T) {
	root := gitRepo(t)
	p := oldBrain(t, root, map[string]string{"decisions.md": oldDecision})
	commitAs(t, root, "other@example.test", "old brain on main")
	gitRun(t, root, "checkout", "-qb", "feature")
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	commitAs(t, root, "other@example.test", "other migrated")
	if got := unconfirmedIDs(t, p); got["use-koin-as-di-20260510-123045"] {
		t.Errorf("a migrated note that was on main's old files is confirmed")
	}
}

func TestDetectBranch_OriginHEAD(t *testing.T) {
	root, _ := gitBrain(t)
	gitRun(t, root, "update-ref", "refs/remotes/origin/develop", "HEAD")
	gitRun(t, root, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/develop")
	if got := detectConfirmedBranch(root, DefaultConfig(root)); got != "origin/develop" {
		t.Errorf("branch = %q, want origin/develop (the remote's default)", got)
	}
}

func TestDetectBranch_FallbackOrder(t *testing.T) {
	root, _ := gitBrain(t)
	if got := detectConfirmedBranch(root, DefaultConfig(root)); got != "main" {
		t.Errorf("only local main exists: got %q", got)
	}
	gitRun(t, root, "update-ref", "refs/remotes/origin/master", "HEAD")
	if got := detectConfirmedBranch(root, DefaultConfig(root)); got != "origin/master" {
		t.Errorf("origin/master beats local main: got %q", got)
	}
	gitRun(t, root, "update-ref", "refs/remotes/origin/main", "HEAD")
	if got := detectConfirmedBranch(root, DefaultConfig(root)); got != "origin/main" {
		t.Errorf("origin/main beats origin/master: got %q", got)
	}
}

func TestDetectBranch_ConfigOverride(t *testing.T) {
	root, _ := gitBrain(t)
	gitRun(t, root, "branch", "release")
	cfg := DefaultConfig(root)
	cfg.ConfirmedBranch = "release"
	if got := detectConfirmedBranch(root, cfg); got != "release" {
		t.Errorf("confirmed_branch override ignored: got %q", got)
	}
	cfg.ConfirmedBranch = "does-not-exist"
	if got := detectConfirmedBranch(root, cfg); got != "main" {
		t.Errorf("a missing override falls back to detection: got %q", got)
	}
}

func TestDetectBranch_NoGit(t *testing.T) {
	root := t.TempDir()
	if got := detectConfirmedBranch(root, DefaultConfig(root)); got != "" {
		t.Errorf("no git → no branch, got %q", got)
	}
}

func TestContext_UnconfirmedDecision(t *testing.T) {
	root, p := gitBrain(t)
	gitRun(t, root, "checkout", "-qb", "feature")
	saveDecision(t, p, "Use Hilt")
	commitAs(t, root, "other@example.test", "someone else's decision")

	cfg := DefaultConfig(root)
	full := BuildContextWith(cfg, nil, p, ContextOptions{Now: indexNow})
	for _, want := range []string{"- state: unconfirmed", "_Notes marked `unconfirmed`"} {
		if !strings.Contains(full, want) {
			t.Errorf("full mode missing %q:\n%s", want, full)
		}
	}

	writeCard(t, p, "bugfixes", "big-20260101-000000", decisionCard("Big", "big-20260101-000000", "2026-01-01", "active", 3000))
	cfg.ContextMaxBytes = 2500
	idx := BuildContextWith(cfg, nil, p, ContextOptions{Now: indexNow})
	active := idx[strings.Index(idx, "## Active Decisions"):strings.Index(idx, "## Index")]
	if strings.Contains(active, "Use Hilt") {
		t.Errorf("an unconfirmed decision is not a project rule:\n%s", active)
	}
	if !strings.Contains(idx, "- [architecture_decision] Use Hilt · unconfirmed ·") {
		t.Errorf("index line must say unconfirmed:\n%s", idx)
	}
}
