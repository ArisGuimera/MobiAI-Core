package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrainSave_UntilFlag(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	_ = runBrain(t, []string{"save", "bugfix", "--root", root, "--title", "Firebase workaround", "--body", "b",
		"--status", "temporary", "--until", "com.google.firebase:firebase-bom@34.0.0"})
	got, _ := readSectionCards(root, "bugfixes")
	if !strings.Contains(string(got), "- until: com.google.firebase:firebase-bom@34.0.0") {
		t.Errorf("until not written:\n%s", got)
	}
	msg := runBrainErr(t, []string{"save", "bugfix", "--root", root, "--title", "Other", "--body", "b", "--until", "nonsense"})
	if !strings.Contains(msg, "<dependency>@<version>") {
		t.Errorf("bad --until message:\n%s", msg)
	}
}

func TestDoctor_ReportsUnparseableUntil(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))

	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	writeCardFile(t, root, "bugfixes", "weird-20261101-000000",
		"## Weird\n\n- id: weird-20261101-000000\n- type: bug_fix\n- status: active\n- date: 2026-11-01\n- until: com.example:lib@soon\n\nb\n")
	if err := os.MkdirAll(filepath.Join(root, "app", "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "app", "src"))

	cmd := NewDoctorCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"Brain:", "can't be checked", `weird-20261101-000000: "soon" in com.example:lib@soon is not a version`} {
		if !strings.Contains(got, want) {
			t.Errorf("doctor output missing %q:\n%s", want, got)
		}
	}
}
