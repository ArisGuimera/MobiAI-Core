package catalog

import (
	"path/filepath"
	"sort"
	"testing"
)

func TestLoad_Sample(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Marketplace.Name != "mobiai" {
		t.Errorf("Marketplace.Name: got %q, want %q", c.Marketplace.Name, "mobiai")
	}
	if len(c.Packs) != 5 {
		t.Fatalf("Packs length: got %d, want 5", len(c.Packs))
	}
}

func TestCatalog_Skills_ReadsDescription(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	comm, err := c.Get(CommunityPack)
	if err != nil {
		t.Fatalf("Get(community): %v", err)
	}
	skills, err := c.Skills(comm)
	if err != nil {
		t.Fatalf("Skills(community): %v", err)
	}
	if len(skills) != 3 {
		t.Fatalf("community skills: got %d, want 3", len(skills))
	}
	// os.ReadDir yields entries alphabetically, so the fixture comes back sorted.
	if skills[0].ID != "alpha-skill" {
		t.Errorf("first skill: got %q, want alpha-skill", skills[0].ID)
	}
	if skills[0].Description != "first community fixture skill" {
		t.Errorf("description: got %q, want %q", skills[0].Description, "first community fixture skill")
	}
}

func TestCatalog_Get(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	kmp, err := c.Get("kmp")
	if err != nil {
		t.Fatalf("Get(kmp): %v", err)
	}
	if kmp.Manifest.Name != "kmp" {
		t.Errorf("Manifest.Name: got %q, want %q", kmp.Manifest.Name, "kmp")
	}
	if kmp.Ref.Source != "./skills/kmp" {
		t.Errorf("Ref.Source: got %q, want %q", kmp.Ref.Source, "./skills/kmp")
	}
	if _, err := c.Get("does-not-exist"); err == nil {
		t.Error("Get(does-not-exist): expected error, got nil")
	}
}

func TestCatalog_Skills_DefaultDir(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	core, _ := c.Get("core")
	skills, err := c.Skills(core)
	if err != nil {
		t.Fatalf("Skills(core): %v", err)
	}
	if len(skills) != 1 || skills[0].ID != "using-mobiai" {
		t.Errorf("core skills: got %+v, want [{ID:using-mobiai ...}]", skills)
	}
}

func TestCatalog_Skills_MultipleDirs(t *testing.T) {
	root := filepath.Join("testdata", "sample")
	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	android, _ := c.Get("android")
	skills, err := c.Skills(android)
	if err != nil {
		t.Fatalf("Skills(android): %v", err)
	}
	ids := make([]string, 0, len(skills))
	for _, s := range skills {
		ids = append(ids, s.ID)
	}
	sort.Strings(ids)
	want := []string{"android-build", "edge-to-edge"}
	if len(ids) != len(want) || ids[0] != want[0] || ids[1] != want[1] {
		t.Errorf("android skill IDs: got %v, want %v", ids, want)
	}
}
