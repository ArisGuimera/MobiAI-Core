package tui

import (
	"testing"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func TestNewModel_Smoke(t *testing.T) {
	c := &catalog.Catalog{Marketplace: &catalog.Marketplace{}}
	c.Reindex()
	m := NewModel(c, &state.Installed{Packs: map[string][]string{}}, []host.HostAdapter{stubHost{}})
	if m.Mode() != ModePicker {
		t.Errorf("default mode: got %v, want ModePicker", m.Mode())
	}
}

func TestNewModel_NoHostsTriggersModeNoHosts(t *testing.T) {
	c := &catalog.Catalog{Marketplace: &catalog.Marketplace{}}
	c.Reindex()
	m := NewModel(c, &state.Installed{Packs: map[string][]string{}}, nil)
	if m.Mode() != ModeNoHosts {
		t.Errorf("got %v, want ModeNoHosts", m.Mode())
	}
}

// stubHost is a minimal HostAdapter for tests that don't care about install behavior.
type stubHost struct{}

func (s stubHost) ID() string                                          { return "stub" }
func (s stubHost) Name() string                                        { return "stub" }
func (s stubHost) Homepage() string                                    { return "" }
func (s stubHost) Detect() host.DetectResult                           { return host.DetectResult{Found: true} }
func (s stubHost) SkillsDir() string                                   { return "" }
func (s stubHost) Capabilities() host.Caps                             { return host.Caps{Skills: true} }
func (s stubHost) Install(skills []catalog.Skill) error                { return nil }
func (s stubHost) Uninstall(skillIDs []string) error                   { return nil }
func (s stubHost) List() ([]host.InstalledSkill, error)                { return nil, nil }
func (s stubHost) Verify() host.DriftReport                            { return host.DriftReport{} }

func TestNewModel_BuildsPackRowsFromCatalog(t *testing.T) {
	c := &catalog.Catalog{
		Marketplace: &catalog.Marketplace{},
		Packs: []catalog.Pack{
			{Ref: catalog.PluginRef{Name: "core", Description: "Core"}, Manifest: catalog.PluginManifest{Version: "2.3.0"}},
			{Ref: catalog.PluginRef{Name: "android", Description: "Android"}, Manifest: catalog.PluginManifest{Version: "2.3.0", Dependencies: []string{"core"}}},
			{Ref: catalog.PluginRef{Name: "kmp", Description: "KMP"}, Manifest: catalog.PluginManifest{Version: "2.2.0", Dependencies: []string{"core", "android"}}},
		},
	}
	c.Reindex()
	m := NewModel(c, &state.Installed{Packs: map[string][]string{}}, []host.HostAdapter{stubHost{}})

	// `core` should be filtered out (universal dep, not user-pickable).
	rows := m.PackRows()
	if len(rows) != 2 {
		t.Fatalf("PackRows: got %d, want 2 (excluding core)", len(rows))
	}
	if rows[0].Name != "android" || rows[1].Name != "kmp" {
		t.Errorf("PackRows order: got [%s %s], want [android kmp]", rows[0].Name, rows[1].Name)
	}
}
