package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestTelemetry_OnOffStatus(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))

	exec := func(args ...string) string {
		cmd := NewTelemetryCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute %v: %v", args, err)
		}
		return out.String()
	}

	if got := exec("status"); !strings.Contains(got, "off") {
		t.Errorf("default status should say off; got: %q", got)
	}
	exec("on")
	if got := exec("status"); !strings.Contains(got, "on") {
		t.Errorf("status after on should say on; got: %q", got)
	}
	exec("off")
	if got := exec("status"); !strings.Contains(got, "off") {
		t.Errorf("status after off should say off; got: %q", got)
	}
}
