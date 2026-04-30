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
