package catalog

import (
	"path/filepath"
	"testing"
)

func TestLoadPlugin_Core(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	pm, err := LoadPlugin(root, "./skills/core")
	if err != nil {
		t.Fatalf("LoadPlugin: %v", err)
	}
	if pm.Name != "core" {
		t.Errorf("Name: got %q, want %q", pm.Name, "core")
	}
	if pm.Version != "0.1.0" {
		t.Errorf("Version: got %q, want %q", pm.Version, "0.1.0")
	}
	if len(pm.Dependencies) != 0 {
		t.Errorf("Dependencies: got %v, want empty", pm.Dependencies)
	}
}

func TestLoadPlugin_AndroidWithDeps(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	pm, err := LoadPlugin(root, "./skills/android")
	if err != nil {
		t.Fatalf("LoadPlugin: %v", err)
	}
	if got, want := pm.Dependencies, []string{"core"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("Dependencies: got %v, want %v", got, want)
	}
	if got := pm.Skills; len(got) != 2 || got[0] != "./skills/" || got[1] != "./skills/google/" {
		t.Errorf("Skills: got %v, want [./skills/ ./skills/google/]", got)
	}
}

func TestLoadPlugin_KMPMultipleDeps(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	pm, err := LoadPlugin(root, "./skills/kmp")
	if err != nil {
		t.Fatalf("LoadPlugin: %v", err)
	}
	want := []string{"core", "android", "ios"}
	if len(pm.Dependencies) != len(want) {
		t.Fatalf("Dependencies length: got %d, want %d", len(pm.Dependencies), len(want))
	}
	for i, d := range want {
		if pm.Dependencies[i] != d {
			t.Errorf("Dependencies[%d]: got %q, want %q", i, pm.Dependencies[i], d)
		}
	}
}

func TestLoadPlugin_MissingPluginJSON(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	_, err := LoadPlugin(root, "./skills/does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing plugin.json, got nil")
	}
}
