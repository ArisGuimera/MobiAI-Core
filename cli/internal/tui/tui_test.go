package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/host"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func TestNewModel_Smoke(t *testing.T) {
	c := &catalog.Catalog{Marketplace: &catalog.Marketplace{}}
	c.Reindex()
	m := NewModel(c, &state.Installed{Packs: map[string][]string{}}, state.Paths{}, []host.HostAdapter{stubHost{}})
	if m.Mode() != ModePicker {
		t.Errorf("default mode: got %v, want ModePicker", m.Mode())
	}
}

func TestNewModel_NoHostsTriggersModeNoHosts(t *testing.T) {
	c := &catalog.Catalog{Marketplace: &catalog.Marketplace{}}
	c.Reindex()
	m := NewModel(c, &state.Installed{Packs: map[string][]string{}}, state.Paths{}, nil)
	if m.Mode() != ModeNoHosts {
		t.Errorf("got %v, want ModeNoHosts", m.Mode())
	}
}

func TestUpdate_ProgressTransition_PersistsInstalledJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	paths, _ := state.NewPaths()
	if err := paths.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	c := &catalog.Catalog{
		Marketplace: &catalog.Marketplace{},
		Packs: []catalog.Pack{
			{Ref: catalog.PluginRef{Name: "core"}, Manifest: catalog.PluginManifest{}},
			{Ref: catalog.PluginRef{Name: "android"}, Manifest: catalog.PluginManifest{Dependencies: []string{"core"}}},
		},
	}
	c.Reindex()
	installed := &state.Installed{Packs: map[string][]string{}}
	m := NewModel(c, installed, paths, []host.HostAdapter{stubHost{}})
	m.mode = ModeProgress
	m.installPlan = []installStep{{pack: "android", host: "stub"}}

	updated, _ := m.Update(installStepDoneMsg{pack: "android", host: "stub"})
	final := updated.(Model)
	if final.mode != ModeResult {
		t.Fatalf("mode: got %v, want ModeResult", final.mode)
	}
	if final.installSaveErr != nil {
		t.Errorf("installSaveErr: %v", final.installSaveErr)
	}

	reloaded, err := state.LoadInstalled(paths)
	if err != nil {
		t.Fatal(err)
	}
	hosts := reloaded.HostsFor("android")
	if len(hosts) != 1 || hosts[0] != "stub" {
		t.Errorf("HostsFor(android): got %v, want [stub]", hosts)
	}
}

// stubHost is a minimal HostAdapter for tests that don't care about install behavior.
type stubHost struct{}

func (s stubHost) ID() string                           { return "stub" }
func (s stubHost) Name() string                         { return "stub" }
func (s stubHost) Homepage() string                     { return "" }
func (s stubHost) Detect() host.DetectResult            { return host.DetectResult{Found: true} }
func (s stubHost) SkillsDir() string                    { return "" }
func (s stubHost) Capabilities() host.Caps              { return host.Caps{Skills: true} }
func (s stubHost) Install(skills []catalog.Skill) error { return nil }
func (s stubHost) Uninstall(skillIDs []string) error    { return nil }
func (s stubHost) List() ([]host.InstalledSkill, error) { return nil, nil }
func (s stubHost) Verify() host.DriftReport             { return host.DriftReport{} }
func (s stubHost) Experimental() bool                   { return false }

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
	m := NewModel(c, &state.Installed{Packs: map[string][]string{}}, state.Paths{}, []host.HostAdapter{stubHost{}})

	// `core` should be filtered out (universal dep, not user-pickable).
	rows := m.PackRows()
	if len(rows) != 2 {
		t.Fatalf("PackRows: got %d, want 2 (excluding core)", len(rows))
	}
	if rows[0].Name != "android" || rows[1].Name != "kmp" {
		t.Errorf("PackRows order: got [%s %s], want [android kmp]", rows[0].Name, rows[1].Name)
	}
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func sampleModel(t *testing.T) Model {
	t.Helper()
	c := &catalog.Catalog{
		Marketplace: &catalog.Marketplace{},
		Packs: []catalog.Pack{
			{Ref: catalog.PluginRef{Name: "core"}, Manifest: catalog.PluginManifest{}},
			{Ref: catalog.PluginRef{Name: "android"}, Manifest: catalog.PluginManifest{Dependencies: []string{"core"}}},
			{Ref: catalog.PluginRef{Name: "ios"}, Manifest: catalog.PluginManifest{Dependencies: []string{"core"}}},
			{Ref: catalog.PluginRef{Name: "kmp"}, Manifest: catalog.PluginManifest{Dependencies: []string{"core", "android", "ios"}}},
		},
	}
	c.Reindex()
	return NewModel(c, &state.Installed{Packs: map[string][]string{}}, state.Paths{}, []host.HostAdapter{stubHost{}})
}

func TestUpdate_NavigationDown(t *testing.T) {
	m := sampleModel(t)
	if m.cursor != 0 {
		t.Fatalf("initial cursor: got %d", m.cursor)
	}
	updated, _ := m.Update(keyMsg("down"))
	if updated.(Model).cursor != 1 {
		t.Errorf("cursor after down: got %d, want 1", updated.(Model).cursor)
	}
}

func TestUpdate_NavigationUp_DoesNotGoNegative(t *testing.T) {
	m := sampleModel(t)
	updated, _ := m.Update(keyMsg("up"))
	if updated.(Model).cursor != 0 {
		t.Errorf("cursor after up at 0: got %d, want 0", updated.(Model).cursor)
	}
}

func TestUpdate_ToggleSelectsPack(t *testing.T) {
	m := sampleModel(t)
	// cursor=0 → "android" (core is filtered out)
	updated, _ := m.Update(keyMsg("space"))
	rows := updated.(Model).PackRows()
	if !rows[0].Selected {
		t.Error("expected android to be selected after space at cursor 0")
	}
}

func TestUpdate_ToggleKMP_MarksAndroidAndIOSAsRequired(t *testing.T) {
	m := sampleModel(t)
	// PackRows order = [android, ios, kmp]; need cursor=2 for kmp
	var um tea.Model
	um, _ = m.Update(keyMsg("down"))
	m = um.(Model)
	um, _ = m.Update(keyMsg("down"))
	m = um.(Model)
	um, _ = m.Update(keyMsg("space"))
	m = um.(Model)

	rows := m.PackRows()
	byName := map[string]PackRow{}
	for _, r := range rows {
		byName[r.Name] = r
	}
	if !byName["kmp"].Selected {
		t.Error("kmp should be Selected")
	}
	if !byName["android"].Required {
		t.Error("android should be Required (dep of kmp)")
	}
	if !byName["ios"].Required {
		t.Error("ios should be Required (dep of kmp)")
	}
}

func TestUpdate_UntoggleKMP_ReleasesRequiredFlags(t *testing.T) {
	m := sampleModel(t)
	for _, k := range []string{"down", "down", "space", "space"} {
		var um tea.Model
		um, _ = m.Update(keyMsg(k))
		m = um.(Model)
	}
	rows := m.PackRows()
	for _, r := range rows {
		if r.Selected || r.Required {
			t.Errorf("after toggle off, %s still has Selected=%v Required=%v", r.Name, r.Selected, r.Required)
		}
	}
}

func TestUpdate_EnterFromPickerGoesToConfirm(t *testing.T) {
	m := sampleModel(t)
	updated, _ := m.Update(keyMsg("enter"))
	if updated.(Model).mode != ModeConfirm {
		t.Errorf("mode after enter: got %v, want ModeConfirm", updated.(Model).mode)
	}
}

func TestUpdate_ConfirmYGoesToProgress(t *testing.T) {
	m := sampleModel(t)
	m.mode = ModeConfirm
	m.userSelected = map[string]bool{"android": true}
	updated, cmd := m.Update(keyMsg("y"))
	if updated.(Model).mode != ModeProgress {
		t.Errorf("mode after Y: got %v, want ModeProgress", updated.(Model).mode)
	}
	if cmd == nil {
		t.Error("expected install Cmd to start, got nil")
	}
}

func TestUpdate_ConfirmNGoesBackToPicker(t *testing.T) {
	m := sampleModel(t)
	m.mode = ModeConfirm
	updated, _ := m.Update(keyMsg("n"))
	if updated.(Model).mode != ModePicker {
		t.Errorf("mode after N: got %v, want ModePicker", updated.(Model).mode)
	}
}

func TestUpdate_ProgressMsgsTransitionToResult(t *testing.T) {
	m := sampleModel(t)
	m.mode = ModeProgress
	m.installPlan = []installStep{
		{pack: "android", host: "stub"},
	}
	updated, _ := m.Update(installStepDoneMsg{pack: "android", host: "stub"})
	if updated.(Model).mode != ModeResult {
		t.Errorf("mode after last step: got %v, want ModeResult", updated.(Model).mode)
	}
}
