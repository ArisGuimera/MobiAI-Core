package cmd

import (
	"strings"
	"testing"
)

// brainWithEntries spins up a brain in a KMP project tree, runs scan,
// and seeds memories with one decision (active/shared) and two bugfixes
// (one ios/temporary, one android/active). Returns the project root.
func brainWithEntries(t *testing.T) string {
	t.Helper()
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	_ = runBrain(t, []string{"scan", "--root", root})
	_ = runBrain(t, []string{
		"save", "decision",
		"--root", root,
		"--title", "Use Koin as DI",
		"--platform", "shared",
		"--area", "dependency_injection",
		"--body", "Koin keeps KMP simple.",
	})
	_ = runBrain(t, []string{
		"save", "bugfix",
		"--root", root,
		"--title", "FirebaseAuth iOS module renaming",
		"--platform", "ios",
		"--status", "temporary",
		"--area", "firebase_auth",
		"--body", "Keep composeApp lowercase.",
	})
	_ = runBrain(t, []string{
		"save", "bugfix",
		"--root", root,
		"--title", "Crash on empty cart Android",
		"--platform", "android",
		"--area", "cart",
		"--body", "Null-check items before [0].",
	})
	return root
}

func TestBrainContext_FilterByPlatform(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{"context", "--root", root, "--platform", "ios"})

	// iOS bugfix passes the filter.
	if !strings.Contains(out, "FirebaseAuth iOS") {
		t.Errorf("expected iOS bugfix in output:\n%s", out)
	}
	// android bugfix drops out.
	if strings.Contains(out, "Crash on empty cart Android") {
		t.Errorf("android bugfix should be filtered out:\n%s", out)
	}
	// shared decision drops out (platform=shared, filter=ios).
	if strings.Contains(out, "Use Koin as DI") {
		t.Errorf("shared decision should be filtered out:\n%s", out)
	}
	// Memory sections without matches show the filter empty-state, not
	// the "No entries yet" string used for genuinely empty sections.
	if !strings.Contains(out, "_No entries match the current filter._") {
		t.Errorf("expected filter empty-state marker:\n%s", out)
	}
}

func TestBrainContext_FilterByStatus(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{"context", "--root", root, "--status", "temporary"})

	if !strings.Contains(out, "FirebaseAuth iOS") {
		t.Errorf("temporary bugfix should pass:\n%s", out)
	}
	if strings.Contains(out, "Crash on empty cart Android") {
		t.Errorf("active bugfix should be filtered out:\n%s", out)
	}
	if strings.Contains(out, "Use Koin as DI") {
		t.Errorf("active decision should be filtered out:\n%s", out)
	}
}

func TestBrainContext_SectionRestriction(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{"context", "--root", root, "--section", "decisions"})

	if !strings.Contains(out, "## Architecture Decisions") {
		t.Errorf("Architecture Decisions section missing:\n%s", out)
	}
	if !strings.Contains(out, "Use Koin as DI") {
		t.Errorf("decision should be rendered:\n%s", out)
	}
	// Bugfixes / Testing / Stack / Rules / Warnings all dropped.
	for _, banned := range []string{
		"## Known Bugfixes",
		"## Testing Patterns",
		"## Detected Stack",
		"## Project Rules",
		"## Warnings",
	} {
		if strings.Contains(out, banned) {
			t.Errorf("section %q should NOT appear with --section decisions:\n%s", banned, out)
		}
	}
}

func TestBrainContext_MultipleSections(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{
		"context", "--root", root,
		"--section", "stack,decisions",
	})
	if !strings.Contains(out, "## Detected Stack") {
		t.Errorf("stack should appear:\n%s", out)
	}
	if !strings.Contains(out, "## Architecture Decisions") {
		t.Errorf("decisions should appear:\n%s", out)
	}
	if strings.Contains(out, "## Known Bugfixes") {
		t.Errorf("bugfixes should NOT appear:\n%s", out)
	}
}

func TestBrainSearch_FindsByQuery(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{"search", "--root", root, "koin"})
	if !strings.Contains(out, "Use Koin as DI") {
		t.Errorf("koin search should find the decision:\n%s", out)
	}
	if !strings.Contains(out, "[decisions]") {
		t.Errorf("output should include section tag:\n%s", out)
	}
	if !strings.Contains(out, "id: use-koin") {
		t.Errorf("output should include id:\n%s", out)
	}
}

func TestBrainSearch_FilterANDsWithQuery(t *testing.T) {
	root := brainWithEntries(t)
	// Both bugfixes match "ios" via... actually only one does. Use
	// firebase + platform filter.
	out := runBrain(t, []string{
		"search", "--root", root,
		"--platform", "ios",
		"firebase",
	})
	if !strings.Contains(out, "FirebaseAuth iOS") {
		t.Errorf("expected firebase iOS hit:\n%s", out)
	}
	if strings.Contains(out, "Crash on empty cart Android") {
		t.Errorf("android bugfix should not appear:\n%s", out)
	}
}

func TestBrainSearch_NoResultsMessage(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{"search", "--root", root, "kubernetes"})
	if !strings.Contains(out, "No results") {
		t.Errorf("expected no-results message:\n%s", out)
	}
}

func TestBrainSearch_RequiresInit(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewBrainCmd()
	cmd.SetArgs([]string{"search", "--root", tmp, "anything"})
	if err := cmd.Execute(); err == nil {
		t.Error("search without init should fail")
	}
}

func TestBrainContext_NoFilterPreservesOldOutput(t *testing.T) {
	root := brainWithEntries(t)
	out := runBrain(t, []string{"context", "--root", root})

	// Sanity: full output includes all 3 entries + all the standard
	// headers. This guards against the filter refactor breaking the
	// default no-flag behavior.
	for _, want := range []string{
		"# MobiAI Brain Context",
		"## Detected Stack",
		"## Project Rules",
		"## Architecture Decisions",
		"Use Koin as DI",
		"## Known Bugfixes",
		"FirebaseAuth iOS",
		"Crash on empty cart Android",
		"## Warnings",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in default output:\n%s", want, out)
		}
	}
}

