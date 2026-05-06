package state

import (
	"testing"
	"time"
)

func TestLoadMeta_MissingReturnsZero(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	p, _ := NewPaths()
	_ = p.EnsureDirs()

	m, err := LoadMeta(p)
	if err != nil {
		t.Fatalf("LoadMeta: %v", err)
	}
	if !m.LastSync.IsZero() {
		t.Errorf("LastSync: got %v, want zero", m.LastSync)
	}
	if m.IsStale(time.Hour) != true {
		t.Errorf("IsStale on zero meta should be true")
	}
}

func TestMeta_SaveLoadFresh(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	p, _ := NewPaths()
	_ = p.EnsureDirs()

	now := time.Now().UTC().Truncate(time.Second)
	m := &CatalogMeta{
		LastSync: now,
		Version:  "2.3.0",
	}
	if err := m.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m2, err := LoadMeta(p)
	if err != nil {
		t.Fatalf("LoadMeta: %v", err)
	}
	if !m2.LastSync.Equal(now) {
		t.Errorf("LastSync: got %v, want %v", m2.LastSync, now)
	}
	if m2.Version != "2.3.0" {
		t.Errorf("Version: got %q, want %q", m2.Version, "2.3.0")
	}
	if m2.IsStale(time.Hour) {
		t.Errorf("IsStale(1h) on just-saved meta should be false")
	}
	if !m2.IsStale(time.Nanosecond) {
		t.Errorf("IsStale(1ns) should be true (any age > 1ns)")
	}
}
