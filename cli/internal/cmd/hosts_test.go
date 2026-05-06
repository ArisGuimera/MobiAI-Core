package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostsList_PrintsAdapters(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	cmd := NewHostsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, n := range []string{"Claude Code", "Cursor", "Gemini CLI", "Codex"} {
		if !strings.Contains(got, n) {
			t.Errorf("output should contain %q; got: %q", n, got)
		}
	}
	if !strings.Contains(got, "no detectado") {
		t.Errorf("expected 'no detectado'; got: %q", got)
	}
}

func TestHostsList_MarksDetected(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := NewHostsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "✓ Claude Code") {
		t.Errorf("expected ✓ Claude Code; got: %q", out.String())
	}
}
