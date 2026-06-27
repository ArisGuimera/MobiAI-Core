package state

import (
	"path/filepath"
	"testing"
)

func TestConfig_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	p, err := NewPaths()
	if err != nil {
		t.Fatal(err)
	}

	// Missing file -> zero-valued config, no error.
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("load missing config: %v", err)
	}
	if cfg.Lang != "" {
		t.Errorf("fresh config should have empty Lang; got %q", cfg.Lang)
	}

	cfg.Lang = "es"
	if err := cfg.Save(p); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Lang != "es" {
		t.Errorf("Lang after round-trip: got %q, want es", got.Lang)
	}
}
