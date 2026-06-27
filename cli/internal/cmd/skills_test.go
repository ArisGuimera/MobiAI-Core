package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

// sampleCatalogRoot points at the on-disk fixture catalog, which (unlike the
// real repo) has a community pack with skills (alpha/bravo/charlie-skill) to
// exercise per-skill install/remove.
func sampleCatalogRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "catalog", "testdata", "sample"))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

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

// communityTestHome wires up a temp HOME with a detectable Claude Code dir and
// returns the temp path so callers can assert on installed skills/state.
func communityTestHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	return tmp
}

func TestSkillsAdd_CommunitySingleSkill(t *testing.T) {
	tmp := communityTestHome(t)
	root := sampleCatalogRoot(t)

	cmd := NewSkillsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", "community/alpha-skill", "--catalog-root", root, "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	// Only the requested skill is installed.
	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "alpha-skill", "SKILL.md")); err != nil {
		t.Errorf("alpha-skill should be installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "bravo-skill")); !os.IsNotExist(err) {
		t.Errorf("bravo-skill should NOT be installed; stat err: %v", err)
	}
	// The community pack's dep (core) is installed too.
	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "using-mobiai")); err != nil {
		t.Errorf("core dep using-mobiai should be installed: %v", err)
	}

	// State records the composite key, not a bare "community".
	paths, _ := state.NewPaths()
	inst, _ := state.LoadInstalled(paths)
	if hosts := inst.HostsFor(state.CommunitySkillKey("alpha-skill")); len(hosts) == 0 {
		t.Errorf("installed.json should record community/alpha-skill; got %v", inst.Packs)
	}
	if _, ok := inst.Packs["community"]; ok {
		t.Errorf("installed.json should not have a bare community key; got %v", inst.Packs)
	}
}

func TestSkillsAdd_CommunityUnknownSkill(t *testing.T) {
	communityTestHome(t)
	root := sampleCatalogRoot(t)

	cmd := NewSkillsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", "community/does-not-exist", "--catalog-root", root, "--yes"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected an error for an unknown community skill")
	}
}

func TestSkillsAdd_BareCommunityInstallsAll(t *testing.T) {
	tmp := communityTestHome(t)
	root := sampleCatalogRoot(t)

	cmd := NewSkillsCmd()
	cmd.SetArgs([]string{"add", "community", "--catalog-root", root, "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, id := range []string{"alpha-skill", "bravo-skill", "charlie-skill"} {
		if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", id)); err != nil {
			t.Errorf("bare community should install %s: %v", id, err)
		}
	}
	// Bare community keeps the legacy whole-pack state key.
	paths, _ := state.NewPaths()
	inst, _ := state.LoadInstalled(paths)
	if hosts := inst.HostsFor("community"); len(hosts) == 0 {
		t.Errorf("bare community should record a community key; got %v", inst.Packs)
	}
}

func TestSkillsRemove_CommunitySingleSkill(t *testing.T) {
	tmp := communityTestHome(t)
	root := sampleCatalogRoot(t)

	addCmd := NewSkillsCmd()
	addCmd.SetArgs([]string{"add", "community/alpha-skill", "community/bravo-skill", "--catalog-root", root, "--yes"})
	if err := addCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	rmCmd := NewSkillsCmd()
	var out bytes.Buffer
	rmCmd.SetOut(&out)
	rmCmd.SetErr(&out)
	rmCmd.SetArgs([]string{"remove", "community/alpha-skill", "--catalog-root", root, "--yes"})
	if err := rmCmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "alpha-skill")); !os.IsNotExist(err) {
		t.Errorf("alpha-skill should be removed; stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude", "skills", "bravo-skill", "SKILL.md")); err != nil {
		t.Errorf("bravo-skill should remain installed: %v", err)
	}

	paths, _ := state.NewPaths()
	inst, _ := state.LoadInstalled(paths)
	if hosts := inst.HostsFor(state.CommunitySkillKey("alpha-skill")); len(hosts) != 0 {
		t.Errorf("community/alpha-skill should be gone from state; got %v", hosts)
	}
	if hosts := inst.HostsFor(state.CommunitySkillKey("bravo-skill")); len(hosts) == 0 {
		t.Errorf("community/bravo-skill should remain in state; got %v", inst.Packs)
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
