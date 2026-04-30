package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatus_PrintsHostsAndInstalled(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	for _, h := range []string{".claude", ".cursor"} {
		if err := os.MkdirAll(filepath.Join(tmp, h), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	cmd := NewStatusCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, expected := range []string{"Hosts presentes", "Claude Code", "Cursor", "MobiAI"} {
		if !strings.Contains(got, expected) {
			t.Errorf("status should contain %q; got: %q", expected, got)
		}
	}
}
