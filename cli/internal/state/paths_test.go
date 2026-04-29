package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPaths_DefaultUnderHome(t *testing.T) {
	// Force MOBIAI_HOME to a temp dir to make the test deterministic.
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)

	p, err := NewPaths()
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}
	if p.Home != tmp {
		t.Errorf("Home: got %q, want %q", p.Home, tmp)
	}
	if p.Bin() != filepath.Join(tmp, "bin") {
		t.Errorf("Bin: got %q, want %q", p.Bin(), filepath.Join(tmp, "bin"))
	}
	if p.Cache() != filepath.Join(tmp, "cache") {
		t.Errorf("Cache: got %q, want %q", p.Cache(), filepath.Join(tmp, "cache"))
	}
	if p.State() != filepath.Join(tmp, "state") {
		t.Errorf("State: got %q, want %q", p.State(), filepath.Join(tmp, "state"))
	}
	if p.InstalledFile() != filepath.Join(tmp, "state", "installed.json") {
		t.Errorf("InstalledFile: got %q, want %q", p.InstalledFile(), filepath.Join(tmp, "state", "installed.json"))
	}
	if p.MetaFile() != filepath.Join(tmp, "cache", "meta.json") {
		t.Errorf("MetaFile: got %q, want %q", p.MetaFile(), filepath.Join(tmp, "cache", "meta.json"))
	}
}

func TestPaths_EnsureDirs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)

	p, err := NewPaths()
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}
	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}
	for _, dir := range []string{p.Bin(), p.Cache(), p.State(), p.Logs()} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Stat(%q): %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q exists but is not a directory", dir)
		}
	}
}

func TestPaths_FallbackToUserHomeDir(t *testing.T) {
	// Without MOBIAI_HOME, NewPaths should land somewhere under os.UserHomeDir.
	t.Setenv("MOBIAI_HOME", "")

	p, err := NewPaths()
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no user home dir available on this runner")
	}
	want := filepath.Join(home, ".mobiai")
	if p.Home != want {
		t.Errorf("Home: got %q, want %q", p.Home, want)
	}
}
