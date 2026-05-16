package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrainSave_RequiresInit(t *testing.T) {
	root := setupKMPProject(t)
	cmd := NewBrainCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"save", "decision",
		"--root", root,
		"--title", "x",
		"--body", "y",
	})
	if err := cmd.Execute(); err == nil {
		t.Errorf("save without init should fail; output:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "no inicializado") {
		t.Errorf("error message should mention 'no inicializado'; got:\n%s", buf.String())
	}
}

func TestBrainSave_DecisionAppendsToFile(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})

	out := runBrain(t, []string{
		"save", "decision",
		"--root", root,
		"--title", "Use Koin as DI",
		"--platform", "shared",
		"--area", "dependency_injection",
		"--body", "### Decision\nUse Koin.\n\n### Reason\nKMP-friendly.",
		"--files", "composeApp/src/commonMain/Module.kt,composeApp/build.gradle.kts",
	})
	if !strings.Contains(out, "✓ guardado") {
		t.Errorf("expected success message, got:\n%s", out)
	}

	got, err := os.ReadFile(filepath.Join(root, ".mobiai", "brain", "memories", "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Use Koin as DI",
		"- type: architecture_decision",
		"- platform: shared",
		"- area: dependency_injection",
		"### Decision",
		"### Files",
		"- composeApp/src/commonMain/Module.kt",
		"- composeApp/build.gradle.kts",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestBrainSave_BugfixTemporaryWritesReviewAfter(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})

	_ = runBrain(t, []string{
		"save", "bugfix",
		"--root", root,
		"--title", "FirebaseAuth iOS module name",
		"--platform", "ios",
		"--status", "temporary",
		"--review-after", "2026-09-01",
		"--body", "### Problem\n...\n\n### Solution\n...",
	})

	got, err := os.ReadFile(filepath.Join(root, ".mobiai", "brain", "memories", "bugfixes.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## FirebaseAuth iOS module name",
		"- type: platform_workaround",
		"- status: temporary",
		"- review_after: 2026-09-01",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestBrainSave_MultipleSavesDoNotOverwrite(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})

	for _, title := range []string{"First", "Second", "Third"} {
		_ = runBrain(t, []string{
			"save", "testing",
			"--root", root,
			"--title", title,
			"--body", "body for " + title,
		})
	}
	got, err := os.ReadFile(filepath.Join(root, ".mobiai", "brain", "memories", "testing.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## First", "## Second", "## Third"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestBrainSave_RejectsBadStatus(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	cmd := NewBrainCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"save", "decision",
		"--root", root,
		"--title", "x",
		"--status", "wat",
		"--body", "y",
	})
	if err := cmd.Execute(); err == nil {
		t.Error("save with invalid status should fail")
	}
}

func TestBrainSave_TitleRequired(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	cmd := NewBrainCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"save", "decision",
		"--root", root,
		"--body", "y",
	})
	if err := cmd.Execute(); err == nil {
		t.Error("save without --title should fail")
	}
}

func TestBrainSave_ContextShowsSavedEntries(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	_ = runBrain(t, []string{
		"save", "decision",
		"--root", root,
		"--title", "Use Koin",
		"--body", "### Decision\nUse Koin.",
	})
	out := runBrain(t, []string{"context", "--root", root})
	if !strings.Contains(out, "## Use Koin") {
		t.Errorf("context should include the saved decision; got:\n%s", out)
	}
	// The "_No entries yet._" placeholder should disappear from
	// Architecture Decisions now that one is saved.
	idx := strings.Index(out, "## Architecture Decisions")
	rest := out[idx:]
	nextSection := strings.Index(rest[len("## Architecture Decisions"):], "##")
	section := rest
	if nextSection > 0 {
		section = rest[:len("## Architecture Decisions")+nextSection]
	}
	if strings.Contains(section, "_No entries yet._") {
		t.Errorf("Architecture Decisions section should not show 'No entries yet' after save:\n%s", section)
	}
}
