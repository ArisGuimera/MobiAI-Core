package host

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

func newTestAdapter() *genericAdapter {
	return &genericAdapter{
		id:         "test",
		name:       "Test Host",
		homepage:   "https://example.com",
		homeSubdir: ".test-host",
		caps:       Caps{Skills: true},
	}
}

func TestGenericAdapter_Identity(t *testing.T) {
	a := newTestAdapter()
	if a.ID() != "test" {
		t.Errorf("ID: got %q", a.ID())
	}
	if a.Name() != "Test Host" {
		t.Errorf("Name: got %q", a.Name())
	}
	if a.Homepage() != "https://example.com" {
		t.Errorf("Homepage: got %q", a.Homepage())
	}
}

func TestGenericAdapter_Capabilities(t *testing.T) {
	caps := newTestAdapter().Capabilities()
	if !caps.Skills {
		t.Error("Capabilities.Skills should be true")
	}
}

func TestGenericAdapter_SkillsDir(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	want := filepath.Join(tmp, ".test-host", "skills")
	if got := newTestAdapter().SkillsDir(); got != want {
		t.Errorf("SkillsDir: got %q, want %q", got, want)
	}
}

func TestGenericAdapter_Detect_Found(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	if err := os.MkdirAll(filepath.Join(tmp, ".test-host"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := newTestAdapter().Detect()
	if !r.Found {
		t.Errorf("Detect.Found: got false (Searched=%v)", r.Searched)
	}
}

func TestGenericAdapter_Detect_NotFound(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	r := newTestAdapter().Detect()
	if r.Found {
		t.Error("Detect.Found: got true, want false")
	}
	if len(r.Searched) == 0 {
		t.Error("Detect.Searched: got empty")
	}
}

func makeFakeSkill(t *testing.T, id string) catalog.Skill {
	t.Helper()
	dir := filepath.Join(t.TempDir(), id)
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+id), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "x.sh"), []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	return catalog.Skill{ID: id, AbsPath: dir}
}

func TestGenericAdapter_Install_CopiesSkills(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)

	a := newTestAdapter()
	if err := a.Install([]catalog.Skill{makeFakeSkill(t, "alpha")}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	for _, rel := range []string{"SKILL.md", "scripts/x.sh"} {
		full := filepath.Join(tmp, ".test-host", "skills", "alpha", filepath.FromSlash(rel))
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected %s to exist: %v", full, err)
		}
	}
}

func TestGenericAdapter_Install_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)

	a := newTestAdapter()
	skill := makeFakeSkill(t, "beta")
	if err := a.Install([]catalog.Skill{skill}); err != nil {
		t.Fatalf("first Install: %v", err)
	}
	if err := a.Install([]catalog.Skill{skill}); err != nil {
		t.Fatalf("second Install: %v", err)
	}
}

func TestGenericAdapter_Uninstall_RemovesSkill(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)

	a := newTestAdapter()
	if err := a.Install([]catalog.Skill{makeFakeSkill(t, "gamma")}); err != nil {
		t.Fatal(err)
	}
	if err := a.Uninstall([]string{"gamma"}); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(tmp, ".test-host", "skills", "gamma")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected %s gone, err=%v", dir, err)
	}
}

func TestGenericAdapter_Uninstall_NoOpForMissing(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	if err := newTestAdapter().Uninstall([]string{"nada"}); err != nil {
		t.Errorf("should be no-op, got: %v", err)
	}
}

func TestGenericAdapter_List_Empty(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	got, err := newTestAdapter().List()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("List: got %v, want empty", got)
	}
}

func TestGenericAdapter_List_AfterInstall(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)

	a := newTestAdapter()
	if err := a.Install([]catalog.Skill{
		makeFakeSkill(t, "one"),
		makeFakeSkill(t, "two"),
	}); err != nil {
		t.Fatal(err)
	}

	got, err := a.List()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, s := range got {
		ids[s.ID] = true
	}
	if !ids["one"] || !ids["two"] {
		t.Errorf("List IDs: got %v", ids)
	}
}

func TestGenericAdapter_List_IgnoresFiles(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)

	a := newTestAdapter()
	if err := os.MkdirAll(a.SkillsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a.SkillsDir(), "stray.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := a.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("List: got %v, want empty (loose files ignored)", got)
	}
}

func TestGenericAdapter_Verify_StubEmpty(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	rep := newTestAdapter().Verify()
	if len(rep.Issues) != 0 {
		t.Errorf("Verify (stub): got %d issues", len(rep.Issues))
	}
}
