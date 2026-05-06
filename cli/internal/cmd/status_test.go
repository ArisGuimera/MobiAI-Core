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
	for _, expected := range []string{"Hosts:", "✓ Claude Code", "✓ Cursor", "MobiAI"} {
		if !strings.Contains(got, expected) {
			t.Errorf("status should contain %q; got: %q", expected, got)
		}
	}
	// Status muestra solo los detectados — nada de "X de Y soportados"
	// ni listado completo con "-" markers.
	for _, unwanted := range []string{"de 36 soportados", "de 9 soportados", "  - "} {
		if strings.Contains(got, unwanted) {
			t.Errorf("status should NOT contain %q; got: %q", unwanted, got)
		}
	}
}

func TestStatus_NoHostsDetected_ShowsHint(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))

	cmd := NewStatusCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "ninguno detectado") {
		t.Errorf("status should hint when no hosts detected; got: %q", got)
	}
}
