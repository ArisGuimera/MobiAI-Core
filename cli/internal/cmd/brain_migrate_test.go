package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeOldBrainNote turns an initialized brain back into the 0.2.3 shape
// for decisions.md (no notice) holding one note.
func writeOldBrainNote(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, ".mobiai", "brain", "memories", "decisions.md")
	content := "# Decisions\n\n## Use Koin\n\n- id: use-koin-20260510-123045\n- type: architecture_decision\n- status: active\n- date: 2026-05-10\n\nUse Koin.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBrainContext_PrintsMigrationMessage(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	writeOldBrainNote(t, root)

	out := runBrain(t, []string{"context", "--root", root})
	first := strings.SplitN(out, "\n", 2)[0]
	if first != "Brain migrated to the new format: 1 notes. Include .mobiai/brain/ in your next commit." {
		t.Errorf("first line = %q\nfull output:\n%s", first, out)
	}
	if _, err := os.Stat(filepath.Join(root, ".mobiai", "brain", "memories", "decisions", "use-koin-20260510-123045.md")); err != nil {
		t.Errorf("card not created: %v", err)
	}
	if !strings.Contains(out, "## Use Koin") {
		t.Errorf("context should still include the note:\n%s", out)
	}

	again := runBrain(t, []string{"context", "--root", root})
	if strings.Contains(again, "Brain migrated") {
		t.Errorf("the message must appear only once:\n%s", again)
	}
}

func TestBrainSearch_TriggersMigration(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	writeOldBrainNote(t, root)

	out := runBrain(t, []string{"search", "Koin", "--root", root})
	if !strings.HasPrefix(out, "Brain migrated to the new format: 1 notes.") {
		t.Errorf("search should migrate first:\n%s", out)
	}
}

func TestBrainInit_WritesNoticeFiles(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	got, err := os.ReadFile(filepath.Join(root, ".mobiai", "brain", "memories", "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- id: mobiai-brain-format-notice") {
		t.Errorf("decisions.md should hold only the notice:\n%s", got)
	}
	for _, name := range []string{"integrations.md", "releases.md"} {
		if _, err := os.Stat(filepath.Join(root, ".mobiai", "brain", "memories", name)); err == nil {
			t.Errorf("%s should no longer be created by init", name)
		}
	}
}
