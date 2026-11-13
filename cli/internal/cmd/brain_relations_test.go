package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// runBrainErr runs a brain command that is expected to fail and returns
// its error text plus everything it printed.
func runBrainErr(t *testing.T, args []string) string {
	t.Helper()
	cmd := NewBrainCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("brain %v should fail; output:\n%s", args, buf.String())
	}
	return err.Error() + "\n" + buf.String()
}

func TestBrainSave_SimilarTitleBrakeAndFlags(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	out := runBrain(t, []string{"save", "bugfix", "--root", root, "--title", "Login crash on screen rotation", "--body", "b"})
	oldID := savedIDRe.FindStringSubmatch(out)[1]

	msg := runBrainErr(t, []string{"save", "bugfix", "--root", root, "--title", "Login crash on rotation of the screen", "--body", "b"})
	if !strings.Contains(msg, "A similar note already exists: "+oldID) || !strings.Contains(msg, "--new") {
		t.Errorf("brake message:\n%s", msg)
	}

	_ = runBrain(t, []string{"save", "bugfix", "--root", root, "--title", "Login crash on rotation of the screen", "--body", "b", "--new"})
	out = runBrain(t, []string{"save", "bugfix", "--root", root, "--title", "Login crash on screen rotation, final fix", "--body", "b", "--supersedes", oldID})
	newID := savedIDRe.FindStringSubmatch(out)[1]
	got, _ := readSectionCards(root, "bugfixes")
	if !strings.Contains(string(got), "- supersedes: "+oldID) {
		t.Errorf("supersedes not written:\n%s", got)
	}
	ctx := runBrain(t, []string{"context", "--root", root})
	if !strings.Contains(ctx, "- state: superseded by "+newID) {
		t.Errorf("context should mark the replaced note:\n%s", ctx)
	}
}

func TestBrainSave_CrashOnlyOnBugfix(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	_ = runBrain(t, []string{"save", "bugfix", "--root", root, "--title", "NPE on login", "--body", "b",
		"--crash", "6a1b2c3d4e5f,NullPointerException at LoginViewModel.kt:42"})
	got, _ := readSectionCards(root, "bugfixes")
	if !strings.Contains(string(got), "- crash: 6a1b2c3d4e5f, NullPointerException at LoginViewModel.kt:42") {
		t.Errorf("crash field:\n%s", got)
	}
	msg := runBrainErr(t, []string{"save", "decision", "--root", root, "--title", "X", "--body", "b", "--crash", "abc"})
	if !strings.Contains(msg, "unknown flag: --crash") {
		t.Errorf("--crash must only exist on bugfix:\n%s", msg)
	}

	out := runBrain(t, []string{"search", "--root", root, "--crash", "6A1B2C3D4E5F"})
	if !strings.Contains(out, "NPE on login") || !strings.Contains(out, "1 result") {
		t.Errorf("search --crash:\n%s", out)
	}
	if !strings.Contains(out, ".mobiai/brain/memories/bugfixes/") {
		t.Errorf("search must show each note's file:\n%s", out)
	}
}

func TestBrainSearch_NeedsQueryOrCrash(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	msg := runBrainErr(t, []string{"search", "--root", root})
	if !strings.Contains(msg, "query") {
		t.Errorf("search without query nor --crash:\n%s", msg)
	}
}
