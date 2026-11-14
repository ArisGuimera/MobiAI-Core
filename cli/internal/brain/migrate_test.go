package brain

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// oldBrain builds a brain as 0.2.3 left it: config plus the five
// single-file memories with their templates, and the given notes.
func oldBrain(t *testing.T, root string, files map[string]string) BrainPaths {
	t.Helper()
	p := NewBrainPaths(root)
	if err := os.MkdirAll(p.MemoriesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig(root)
	if err := cfg.Save(p); err != nil {
		t.Fatal(err)
	}
	for _, mf := range MemoryFiles {
		content := mf.Template
		if extra, ok := files[mf.Name]; ok {
			content += extra
		}
		mustWrite(t, filepath.Join(p.MemoriesDir, mf.Name), content)
	}
	return p
}

const oldDecision = `
## Use Koin as DI

- id: use-koin-as-di-20260510-123045
- type: architecture_decision
- status: active
- platform: shared
- area: dependency_injection
- date: 2026-05-10

### Decision
Use Koin.

### Files
- composeApp/src/commonMain/Module.kt
`

const oldBugfixWithC5 = `
## Login crash on screen rotation

- id: login-crash-on-screen-rotation-20260928-185828
- type: bug_fix
- status: active
- area: auth
- date: 2026-09-28

## Symptom
The app crashes when the device rotates on the login screen.

## Root cause
The form state is kept in the Activity instead of the ViewModel.

## Fix
Move the form state to the ViewModel.
`

// snapshot returns every file under dir keyed by slash path relative to it.
func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestMigrate_CreatesOneCardPerLegacyNote(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{
		"decisions.md": oldDecision,
		"bugfixes.md":  oldBugfixWithC5,
	})
	rep, err := Migrate(p)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Created != 2 {
		t.Fatalf("Created = %d, want 2", rep.Created)
	}
	card, err := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions", "use-koin-as-di-20260510-123045.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := `## Use Koin as DI

- id: use-koin-as-di-20260510-123045
- type: architecture_decision
- status: active
- platform: shared
- area: dependency_injection
- date: 2026-05-10
- verified: 2026-05-10

### Decision
Use Koin.

### Files
- composeApp/src/commonMain/Module.kt
`
	if string(card) != want {
		t.Errorf("card mismatch.\n--- got ---\n%s\n--- want ---\n%s", card, want)
	}
}

func TestMigrate_KeepsLegacyNotes(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	if !strings.Contains(string(got), strings.TrimSpace(oldDecision)) {
		t.Errorf("the old note must stay in decisions.md:\n%s", got)
	}
}

func TestMigrate_IsDeterministic(t *testing.T) {
	files := map[string]string{
		"decisions.md": oldDecision + "\n## Hand written\n\n- status: active\n- date: 2026-03-01\n\nNo id here.\n",
		"bugfixes.md":  oldBugfixWithC5,
	}
	a := oldBrain(t, filepath.Join(t.TempDir(), "a"), files)
	b := oldBrain(t, filepath.Join(t.TempDir(), "b"), files)

	orig := timeNow
	defer func() { timeNow = orig }()
	timeNow = func() time.Time { return time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC) }
	if _, err := Migrate(a); err != nil {
		t.Fatal(err)
	}
	timeNow = func() time.Time { return time.Date(2027, 2, 3, 23, 59, 59, 0, time.UTC) }
	if _, err := Migrate(b); err != nil {
		t.Fatal(err)
	}

	sa, sb := snapshot(t, a.MemoriesDir), snapshot(t, b.MemoriesDir)
	if len(sa) != len(sb) {
		t.Fatalf("different file sets: %d vs %d", len(sa), len(sb))
	}
	for name, ca := range sa {
		if sb[name] != ca {
			t.Errorf("%s differs between the two migrations", name)
		}
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision, "bugfixes.md": oldBugfixWithC5})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, p.MemoriesDir)
	rep, err := Migrate(p)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Changed() {
		t.Errorf("second run reported changes: %+v", rep)
	}
	after := snapshot(t, p.MemoriesDir)
	for name, c := range before {
		if after[name] != c {
			t.Errorf("%s changed on the second run", name)
		}
	}
}

func TestMigrate_RepairsC5(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"bugfixes.md": oldBugfixWithC5})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	cards, _ := filepath.Glob(filepath.Join(p.MemoriesDir, "bugfixes", "*.md"))
	if len(cards) != 1 {
		t.Fatalf("expected the three ## fragments to stay in one card, got %d cards", len(cards))
	}
	got, _ := os.ReadFile(cards[0])
	for _, want := range []string{
		"### Symptom\nThe app crashes when the device rotates on the login screen.",
		"### Root cause\nThe form state is kept in the Activity",
		"### Fix\nMove the form state to the ViewModel.",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("card lost %q:\n%s", want, got)
		}
	}
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 {
		t.Errorf("LoadNotes = %v, want the single repaired note", titles(notes))
	}
}

func TestLoadNotes_RepairsC5InMemory(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"bugfixes.md": oldBugfixWithC5})
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || !strings.Contains(notes[0].Body, "### Fix") {
		t.Errorf("legacy read must repair C5 too: %v", titles(notes))
	}
}

var contentIDRe = regexp.MustCompile(`^hand-written-20260301-[0-9a-f]{8}$`)

func TestMigrate_NoteWithoutIDGetsContentID(t *testing.T) {
	note := "\n## Hand written\n\n- status: active\n- date: 2026-03-01\n\nNo id here.\n"
	idFor := func(body string) string {
		p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": body})
		if _, err := Migrate(p); err != nil {
			t.Fatal(err)
		}
		cards, _ := filepath.Glob(filepath.Join(p.MemoriesDir, "decisions", "*.md"))
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		return strings.TrimSuffix(filepath.Base(cards[0]), ".md")
	}
	id1, id2 := idFor(note), idFor(note)
	if !contentIDRe.MatchString(id1) {
		t.Errorf("id %q does not look like slug-date-hash", id1)
	}
	if id1 != id2 {
		t.Errorf("same content gave different ids: %q vs %q", id1, id2)
	}
	if id3 := idFor(strings.Replace(note, "No id here.", "Different text.", 1)); id3 == id1 {
		t.Errorf("different content must give a different id, both %q", id1)
	}
}

func TestMigrate_NoteWithoutIDKeepsAccentedWordsInID(t *testing.T) {
	note := "\n## Reintentar si se corta la conexión\n\n- status: active\n- date: 2026-03-01\n\nNo id here.\n"
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": note})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	cards, _ := filepath.Glob(filepath.Join(p.MemoriesDir, "decisions", "*.md"))
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	id := strings.TrimSuffix(filepath.Base(cards[0]), ".md")
	re := regexp.MustCompile(`^reintentar-si-se-corta-la-conexion-20260301-[0-9a-f]{8}$`)
	if !re.MatchString(id) {
		t.Errorf("id %q: accented letters must become plain letters, not hyphens", id)
	}
}

func TestMigrate_DuplicateIDsKeepBoth(t *testing.T) {
	dup := strings.Replace(oldDecision, "Use Koin.", "Use Koin, ported version.", 1)
	p := oldBrain(t, t.TempDir(), map[string]string{
		"decisions.md": oldDecision + oldDecision + dup,
	})
	rep, err := Migrate(p)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Created != 2 {
		t.Errorf("Created = %d, want 2 (identical copies collapse, the edited one is kept)", rep.Created)
	}
	notes, _ := LoadNotes(p)
	if len(notes) != 2 {
		t.Errorf("LoadNotes = %d notes, want 2", len(notes))
	}
}

func TestNotice_InsertedAfterH1_LF(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	want := "# Decisions\n\n" + formatNotice + "\n<!--\n"
	if !strings.HasPrefix(string(got), want) {
		t.Errorf("notice must sit right after the H1.\n--- got ---\n%s", got)
	}
}

func TestNotice_InsertedAfterH1_CRLF(t *testing.T) {
	root := t.TempDir()
	p := oldBrain(t, root, nil)
	crlf := strings.ReplaceAll(decisionsTemplate+oldDecision, "\n", "\r\n")
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"), crlf)
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "decisions.md"))
	if strings.Count(string(got), "\n") != strings.Count(string(got), "\r\n") {
		t.Errorf("a CRLF file must stay CRLF on every line:\n%q", got)
	}
	if !strings.Contains(string(got), strings.ReplaceAll(formatNotice, "\n", "\r\n")) {
		t.Errorf("notice missing:\n%q", got)
	}
}

func TestNotice_AtTopWithoutH1(t *testing.T) {
	got, changed := insertNotice("## Note\n\n- id: x\n")
	if !changed || !strings.HasPrefix(got, formatNotice+"\n## Note\n") {
		t.Errorf("without an H1 the notice goes first:\n%s", got)
	}
}

func TestNotice_NotDuplicated(t *testing.T) {
	once, _ := insertNotice(decisionsTemplate)
	twice, changed := insertNotice(once)
	if changed || twice != once {
		t.Errorf("notice inserted twice:\n%s", twice)
	}
}

func TestNotice_CreatesMissingLegacyFile(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision})
	if err := os.Remove(filepath.Join(p.MemoriesDir, "testing.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(p.MemoriesDir, "testing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# Testing Patterns\n\n"+formatNotice {
		t.Errorf("recreated testing.md:\n%s", got)
	}
}

func TestNotice_NotAddedToIntegrationsOrReleases(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"integrations.md", "releases.md"} {
		got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, name))
		if strings.Contains(string(got), formatNoticeID) {
			t.Errorf("%s must not get the notice", name)
		}
	}
}

func TestNotice_OldParserSeesItAsOverdueTemporary(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	// ParseMemoryFile is the 0.2.3 parser, kept unchanged.
	for _, name := range []string{"decisions.md", "bugfixes.md", "testing.md"} {
		entries, err := ParseMemoryFile(filepath.Join(p.MemoriesDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 0 {
			t.Fatalf("%s: old parser sees no entries", name)
		}
		n := entries[0]
		if n.Title != "MobiAI Brain moved to a new format: update mobiai" ||
			n.Get("status") != "temporary" || n.Get("review_after") != "2000-01-01" {
			t.Errorf("%s: first old entry = %q %v", name, n.Title, n.Meta)
		}
		if !strings.Contains(n.Body, "mobiai update") {
			t.Errorf("%s: notice body lost the update instruction:\n%s", name, n.Body)
		}
	}
}

func TestMigrate_OldSaveAfterMigrationIsPickedUp(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"bugfixes.md": oldBugfixWithC5})
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	// What 0.2.3's save does: append the rendered entry at the end.
	f, err := os.OpenFile(filepath.Join(p.MemoriesDir, "bugfixes.md"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("\n## Saved by an old mobiai\n\n- id: saved-by-an-old-mobiai-20261101-080000\n- type: bug_fix\n- status: active\n- date: 2026-11-01\n\nBody.\n")
	f.Close()

	rep, err := Migrate(p)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Created != 1 {
		t.Errorf("Created = %d, want 1", rep.Created)
	}
	if _, err := os.Stat(filepath.Join(p.MemoriesDir, "bugfixes", "saved-by-an-old-mobiai-20261101-080000.md")); err != nil {
		t.Errorf("card for the old save missing: %v", err)
	}
}

func TestMigrate_ReadOnlyKeepsWorking(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{"decisions.md": oldDecision})
	orig := migrationWrite
	defer func() { migrationWrite = orig }()
	migrationWrite = func(string, []byte) error { return &fs.PathError{Op: "open", Path: "x", Err: fs.ErrPermission} }

	rep, err := Migrate(p)
	if err != nil {
		t.Fatalf("a read-only brain must not fail the command: %v", err)
	}
	if !rep.ReadOnly || !errors.Is(rep.Err, fs.ErrPermission) {
		t.Errorf("report = %+v, want ReadOnly with the write error", rep)
	}
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Get("id") != "use-koin-as-di-20260510-123045" {
		t.Errorf("notes still readable from the old file: %v", titles(notes))
	}
}

func TestMigrate_NotABrainIsNoop(t *testing.T) {
	rep, err := Migrate(NewBrainPaths(t.TempDir()))
	if err != nil || rep.Changed() {
		t.Errorf("rep=%+v err=%v", rep, err)
	}
}

func TestInit_CreatesDirsAndNoticeFiles(t *testing.T) {
	p := NewBrainPaths(t.TempDir())
	created, err := EnsureMemoryFiles(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(created, ",") != "decisions.md,bugfixes.md,testing.md" {
		t.Errorf("created = %v", created)
	}
	for _, dir := range []string{"decisions", "bugfixes", "testing"} {
		if info, err := os.Stat(filepath.Join(p.MemoriesDir, dir)); err != nil || !info.IsDir() {
			t.Errorf("missing card dir %s: %v", dir, err)
		}
	}
	got, _ := os.ReadFile(filepath.Join(p.MemoriesDir, "bugfixes.md"))
	if string(got) != "# Bugfixes\n\n"+formatNotice {
		t.Errorf("bugfixes.md:\n%s", got)
	}
	for _, name := range []string{"integrations.md", "releases.md"} {
		if _, err := os.Stat(filepath.Join(p.MemoriesDir, name)); err == nil {
			t.Errorf("%s should no longer be created", name)
		}
	}
}

// git helpers for the merge test.

func gitAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func gitRepo(t *testing.T) string {
	t.Helper()
	gitAvailable(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	gitRun(t, dir, "config", "user.email", "me@example.test")
	gitRun(t, dir, "config", "user.name", "me")
	gitRun(t, dir, "config", "core.autocrlf", "false")
	gitRun(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func TestMigrate_TwoBranchesMergeClean(t *testing.T) {
	root := gitRepo(t)
	p := oldBrain(t, root, map[string]string{"decisions.md": oldDecision, "bugfixes.md": oldBugfixWithC5})
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "old brain")

	gitRun(t, root, "checkout", "-qb", "a")
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Note from A", Body: "a",
		now: clockAt(2026, 11, 1, 10, 0, 0)}); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "a")

	gitRun(t, root, "checkout", "-q", "main")
	gitRun(t, root, "checkout", "-qb", "b")
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Note from B", Body: "b",
		now: clockAt(2026, 11, 2, 10, 0, 0)}); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "b")

	gitRun(t, root, "merge", "-q", "--no-edit", "a")
	notes, err := LoadNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 4 {
		t.Errorf("after the merge: %v", titles(notes))
	}
}

func TestMigrate_OpenOldBranchMergesClean(t *testing.T) {
	root := gitRepo(t)
	p := oldBrain(t, root, nil)
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "empty old brain")

	// An old mobiai on a branch appends a note at the end of bugfixes.md...
	gitRun(t, root, "checkout", "-qb", "old")
	f, _ := os.OpenFile(filepath.Join(p.MemoriesDir, "bugfixes.md"), os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.WriteString(oldBugfixWithC5)
	f.Close()
	gitRun(t, root, "commit", "-qam", "old save")

	// ...while main migrates (notice only: there are no notes yet).
	gitRun(t, root, "checkout", "-q", "main")
	if _, err := Migrate(p); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "migrate")

	gitRun(t, root, "merge", "-q", "--no-edit", "old")
	rep, err := Migrate(p)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Created != 1 {
		t.Errorf("the old branch's note should become a card, Created = %d", rep.Created)
	}
}
