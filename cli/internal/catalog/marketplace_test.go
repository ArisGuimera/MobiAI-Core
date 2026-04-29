package catalog

import (
	"path/filepath"
	"testing"
)

func TestLoadMarketplace_Sample(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	mp, err := LoadMarketplace(root)
	if err != nil {
		t.Fatalf("LoadMarketplace: %v", err)
	}
	if mp.Name != "mobiai" {
		t.Errorf("Name: got %q, want %q", mp.Name, "mobiai")
	}
	if mp.Metadata.Version != "0.1.0" {
		t.Errorf("Metadata.Version: got %q, want %q", mp.Metadata.Version, "0.1.0")
	}
	if len(mp.Plugins) != 4 {
		t.Fatalf("Plugins length: got %d, want 4", len(mp.Plugins))
	}
	if mp.Plugins[0].Name != "core" {
		t.Errorf("Plugins[0].Name: got %q, want %q", mp.Plugins[0].Name, "core")
	}
	if mp.Plugins[1].Source != "./plugins/android" {
		t.Errorf("Plugins[1].Source: got %q, want %q", mp.Plugins[1].Source, "./plugins/android")
	}
	if got := mp.Plugins[1].Tags; len(got) != 2 || got[0] != "android" {
		t.Errorf("Plugins[1].Tags: got %v, want [android kotlin]", got)
	}
}

func TestLoadMarketplace_MissingFile(t *testing.T) {
	_, err := LoadMarketplace(filepath.Join("testdata", "does-not-exist"))
	if err == nil {
		t.Fatal("expected error for missing marketplace.json, got nil")
	}
}
