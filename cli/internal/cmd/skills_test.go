package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRootForFixture returns the path to the actual MobiAI-Core repo so we
// test against the real catalog. From cli/internal/cmd/, we go up 3 dirs.
func repoRootForFixture(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestSkillsAdd_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))

	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	root := repoRootForFixture(t)
	cmd := NewSkillsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", "android", "--catalog-root", root, "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "mobiai-android-build", "SKILL.md")); err != nil {
		t.Errorf("expected skill installed: %v", err)
	}
	if !strings.Contains(out.String(), "android") {
		t.Errorf("output should mention android; got: %q", out.String())
	}
}

func TestSkillsAdd_UnknownPack(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := NewSkillsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", "androyd", "--catalog-root", repoRootForFixture(t), "--yes"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for unknown pack")
	}
}

func TestSkillsRemove_RemovesInstalledPack(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	root := repoRootForFixture(t)

	addCmd := NewSkillsCmd()
	addCmd.SetArgs([]string{"add", "android", "--catalog-root", root, "--yes"})
	if err := addCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	removeCmd := NewSkillsCmd()
	var out bytes.Buffer
	removeCmd.SetOut(&out)
	removeCmd.SetErr(&out)
	removeCmd.SetArgs([]string{"remove", "android", "--catalog-root", root, "--yes"})
	if err := removeCmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "mobiai-android-build")); !os.IsNotExist(err) {
		t.Errorf("expected android skills removed; stat err: %v", err)
	}
}

func TestWithFirebenderHost_AppendsProjectFirebender(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	hosts, err := withFirebenderHost(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0].ID() != "firebender" {
		t.Fatalf("hosts: got %#v, want only firebender", hosts)
	}
	// Resolve symlinks on the base tmp dir (macOS /var -> /private/var).
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(resolvedRoot, ".firebender", "skills")
	if got := hosts[0].SkillsDir(); got != want {
		t.Fatalf("SkillsDir: got %q, want %q", got, want)
	}
}

func TestSkillsList_PrintsTable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	root := repoRootForFixture(t)

	addCmd := NewSkillsCmd()
	addCmd.SetArgs([]string{"add", "android", "--catalog-root", root, "--yes"})
	if err := addCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	listCmd := NewSkillsCmd()
	var out bytes.Buffer
	listCmd.SetOut(&out)
	listCmd.SetErr(&out)
	listCmd.SetArgs([]string{"list"})
	if err := listCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "android") {
		t.Errorf("list should mention android; got: %q", got)
	}
	if !strings.Contains(got, "claude-code") {
		t.Errorf("list should mention claude-code; got: %q", got)
	}
}
