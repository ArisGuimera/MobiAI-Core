package state

import (
	"sort"
	"testing"
)

func TestLoadInstalled_MissingFileReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	p, _ := NewPaths()
	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	s, err := LoadInstalled(p)
	if err != nil {
		t.Fatalf("LoadInstalled: %v", err)
	}
	if len(s.Packs) != 0 {
		t.Errorf("Packs: got %v, want empty", s.Packs)
	}
}

func TestInstalled_AddSaveLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	p, _ := NewPaths()
	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	s, _ := LoadInstalled(p)
	s.Add("core", "claude-code")
	s.Add("core", "cursor")
	s.Add("android", "claude-code")
	s.Add("core", "claude-code") // duplicate; should be a no-op

	if err := s.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	s2, err := LoadInstalled(p)
	if err != nil {
		t.Fatalf("LoadInstalled (round-trip): %v", err)
	}
	hosts := s2.HostsFor("core")
	sort.Strings(hosts)
	want := []string{"claude-code", "cursor"}
	if len(hosts) != 2 || hosts[0] != want[0] || hosts[1] != want[1] {
		t.Errorf("HostsFor(core): got %v, want %v", hosts, want)
	}
	if got := s2.HostsFor("android"); len(got) != 1 || got[0] != "claude-code" {
		t.Errorf("HostsFor(android): got %v, want [claude-code]", got)
	}
	if got := s2.HostsFor("ios"); got != nil {
		t.Errorf("HostsFor(ios): got %v, want nil", got)
	}
}

func TestInstalled_Remove(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	p, _ := NewPaths()
	_ = p.EnsureDirs()

	s, _ := LoadInstalled(p)
	s.Add("core", "claude-code")
	s.Add("core", "cursor")

	s.Remove("core", "claude-code")
	hosts := s.HostsFor("core")
	if len(hosts) != 1 || hosts[0] != "cursor" {
		t.Errorf("after Remove(core, claude-code): got %v, want [cursor]", hosts)
	}

	s.Remove("core", "cursor")
	if got := s.HostsFor("core"); got != nil {
		t.Errorf("after removing all hosts of core: got %v, want nil", got)
	}
}
