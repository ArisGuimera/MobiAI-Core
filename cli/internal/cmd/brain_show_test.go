package cmd

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

var savedIDRe = regexp.MustCompile(`id: (\S+)`)

func TestBrainShow_PrintsCard(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	out := runBrain(t, []string{"save", "decision", "--root", root, "--title", "Use Koin", "--body", "Use Koin."})
	m := savedIDRe.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("no id in save output:\n%s", out)
	}

	shown := runBrain(t, []string{"show", m[1], "--root", root})
	for _, want := range []string{"## Use Koin", "- id: " + m[1], "Use Koin."} {
		if !strings.Contains(shown, want) {
			t.Errorf("missing %q in show output:\n%s", want, shown)
		}
	}
}

func TestBrainShow_UnknownID(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	cmd := NewBrainCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"show", "nope", "--root", root})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "no entry with id") {
		t.Errorf("show of an unknown id should fail, err = %v", err)
	}
}
