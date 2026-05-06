package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_RunsAndPrintsSections(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))

	cmd := NewDoctorCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, section := range []string{"Hosts soportados", "Catálogo", "Drift"} {
		if !strings.Contains(got, section) {
			t.Errorf("doctor should mention %q; got: %q", section, got)
		}
	}
}
