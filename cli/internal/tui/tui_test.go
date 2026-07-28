package tui

import (
	"path/filepath"
	"strings"
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
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

// communityModel builds a picker model from the on-disk sample catalog, which
// includes a community pack with three skills (alpha/bravo/charlie-skill). The
// community sub-picker reads real skills off disk, so these tests can't use the
// hand-built inline catalogs the other tests use.
func communityModel(t *testing.T) Model {
	t.Helper()
	root := filepath.Join("..", "catalog", "testdata", "sample")
	c, err := catalog.Load(root)
	if err != nil {
		t.Fatalf("load sample catalog: %v", err)
	}
	return NewModel(c, &state.Installed{Packs: map[string][]string{}}, state.Paths{}, []host.HostAdapter{stubHost{}})
}

// communityRowIndex returns the cursor index of the community row in the main
// picker, or -1 if absent.
func communityRowIndex(m Model) int {
	for i, r := range m.PackRows() {
		if r.IsCommunity {
			return i
		}
	}
	return -1
}

func TestCommunityModel_CachesSkillsSorted(t *testing.T) {
	m := communityModel(t)
	if len(m.communitySkills) != 3 {
		t.Fatalf("communitySkills: got %d, want 3", len(m.communitySkills))
	}
	if m.communitySkills[0].ID != "alpha-skill" || m.communitySkills[2].ID != "charlie-skill" {
		t.Errorf("communitySkills not sorted: got %s..%s", m.communitySkills[0].ID, m.communitySkills[2].ID)
	}
	if m.communitySkills[0].Description != "first community fixture skill" {
		t.Errorf("description not loaded: got %q", m.communitySkills[0].Description)
	}
}

func TestUpdate_EnterOnCommunity_OpensSubPicker(t *testing.T) {
	m := communityModel(t)
	m.cursor = communityRowIndex(m)
	if m.cursor < 0 {
		t.Fatal("no community row found")
	}
	updated, _ := m.Update(keyMsg("enter"))
	if updated.(Model).mode != ModeCommunityPicker {
		t.Errorf("mode after enter on community: got %v, want ModeCommunityPicker", updated.(Model).mode)
	}
}

func TestUpdate_SpaceOnCommunity_OpensSubPicker(t *testing.T) {
	m := communityModel(t)
	m.cursor = communityRowIndex(m)
	updated, _ := m.Update(keyMsg("space"))
	final := updated.(Model)
	if final.mode != ModeCommunityPicker {
		t.Errorf("mode after space on community: got %v, want ModeCommunityPicker", final.mode)
	}
	// Space on community must NOT toggle the whole pack as a user action.
	if final.userActions[catalog.CommunityPack] != ActionNone {
		t.Errorf("community pack should not be toggled at pack level; got %v", final.userActions)
	}
}

func TestCommunitySubPicker_ToggleSelectsSkill(t *testing.T) {
	m := communityModel(t)
	m.mode = ModeCommunityPicker
	m.subCursor = 0 // alpha-skill
	updated, _ := m.Update(keyMsg("space"))
	final := updated.(Model)
	if final.communityActions["alpha-skill"] != ActionInstall {
		t.Errorf("alpha-skill action after space: got %v, want ActionInstall", final.communityActions["alpha-skill"])
	}
	// Second space cycles back to None.
	updated, _ = final.Update(keyMsg("space"))
	if updated.(Model).communityActions["alpha-skill"] != ActionNone {
		t.Errorf("second space should reset to None; got %v", updated.(Model).communityActions["alpha-skill"])
	}
}

func TestCommunitySubPicker_EnterWithPendingGoesToConfirm(t *testing.T) {
	m := communityModel(t)
	m.mode = ModeCommunityPicker
	m.communityActions = map[string]Action{"alpha-skill": ActionInstall}
	updated, _ := m.Update(keyMsg("enter"))
	if updated.(Model).mode != ModeConfirm {
		t.Errorf("mode after enter with pending: got %v, want ModeConfirm", updated.(Model).mode)
	}
}

func TestCommunitySubPicker_EscReturnsToPickerKeepingSelections(t *testing.T) {
	m := communityModel(t)
	m.mode = ModeCommunityPicker
	m.communityActions = map[string]Action{"alpha-skill": ActionInstall}
	updated, _ := m.Update(keyMsg("esc"))
	final := updated.(Model)
	if final.mode != ModePicker {
		t.Errorf("mode after esc: got %v, want ModePicker", final.mode)
	}
	if final.communityActions["alpha-skill"] != ActionInstall {
		t.Errorf("esc should keep pending selections; got %v", final.communityActions)
	}
}

func TestCommunitySubPicker_ToggleInstalledSkill_MarksUninstall(t *testing.T) {
	m := communityModel(t)
	// alpha-skill is already installed in the only detected host.
	m.state.Packs[state.CommunitySkillKey("alpha-skill")] = []string{"stub"}
	m.mode = ModeCommunityPicker
	m.subCursor = 0 // alpha-skill (sorted first)

	updated, _ := m.Update(keyMsg("space"))
	final := updated.(Model)
	if final.communityActions["alpha-skill"] != ActionUninstall {
		t.Fatalf("space on an installed community skill should mark Uninstall; got %v", final.communityActions["alpha-skill"])
	}

	plan := final.buildInstallPlan()
	var sawUninstall bool
	for _, s := range plan {
		if s.pack == catalog.CommunityPack && s.kind == stepUninstall {
			sawUninstall = true
			if len(s.skillIDs) != 1 || s.skillIDs[0] != "alpha-skill" {
				t.Errorf("uninstall step skillIDs: got %v, want [alpha-skill]", s.skillIDs)
			}
		}
	}
	if !sawUninstall {
		t.Error("expected a community stepUninstall in the plan")
	}
}

func TestBuildInstallPlan_CommunitySkill_PullsCoreAndFiltersStep(t *testing.T) {
	m := communityModel(t)
	m.communityActions = map[string]Action{"alpha-skill": ActionInstall}

	plan := m.buildInstallPlan()
	if len(plan) == 0 {
		t.Fatal("expected non-empty plan")
	}

	var coreIdx, commIdx = -1, -1
	for i, s := range plan {
		switch s.pack {
		case "core":
			if s.kind == stepInstall {
				coreIdx = i
			}
		case catalog.CommunityPack:
			if s.kind == stepInstall {
				commIdx = i
				if len(s.skillIDs) != 1 || s.skillIDs[0] != "alpha-skill" {
					t.Errorf("community step skillIDs: got %v, want [alpha-skill]", s.skillIDs)
				}
			}
		}
	}
	if coreIdx < 0 {
		t.Error("expected a core install step (community depends on core)")
	}
	if commIdx < 0 {
		t.Fatal("expected a community install step")
	}
	if coreIdx > commIdx {
		t.Errorf("core (%d) must install before community (%d)", coreIdx, commIdx)
	}
}

func TestHandleInstallStep_CommunitySkill_RecordsCompositeKey(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	paths, _ := state.NewPaths()
	if err := paths.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	c := &catalog.Catalog{Marketplace: &catalog.Marketplace{}}
	c.Reindex()
	installed := &state.Installed{Packs: map[string][]string{}}
	m := NewModel(c, installed, paths, []host.HostAdapter{stubHost{}})
	m.mode = ModeProgress
	m.installPlan = []installStep{{pack: "community", host: "stub", kind: stepInstall, skillIDs: []string{"alpha-skill"}}}

	updated, _ := m.Update(installStepDoneMsg{pack: "community", host: "stub", kind: stepInstall, skillIDs: []string{"alpha-skill"}})
	final := updated.(Model)

	if hosts := final.state.HostsFor(state.CommunitySkillKey("alpha-skill")); len(hosts) != 1 || hosts[0] != "stub" {
		t.Errorf("community/alpha-skill state: got %v, want [stub]", hosts)
	}
	if _, ok := final.state.Packs["community"]; ok {
		t.Errorf("must not write a bare community key; got %v", final.state.Packs)
	}
}

func TestCommunityMarker_ReflectsInstallState(t *testing.T) {
	m := communityModel(t)

	// None installed → empty marker.
	if got := m.communityMarker(); strings.ContainsAny(got, "✓~+-") {
		t.Errorf("no-install marker should be empty-ish; got %q", got)
	}

	// One of three installed → partial "[~]".
	m.state.Packs[state.CommunitySkillKey("alpha-skill")] = []string{"stub"}
	if got := m.communityMarker(); !strings.Contains(got, "~") {
		t.Errorf("partial marker: got %q, want a ~", got)
	}

	// All three installed in the only host → "[✓]".
	m.state.Packs[state.CommunitySkillKey("bravo-skill")] = []string{"stub"}
	m.state.Packs[state.CommunitySkillKey("charlie-skill")] = []string{"stub"}
	if got := m.communityMarker(); !strings.Contains(got, "✓") {
		t.Errorf("all-installed marker: got %q, want a ✓", got)
	}

	// A pending install takes priority → "[+]".
	m.communityActions = map[string]Action{"alpha-skill": ActionInstall}
	if got := m.communityMarker(); !strings.Contains(got, "+") {
		t.Errorf("pending-install marker: got %q, want a +", got)
	}
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
	// cursor=0 → "android" (core is filtered out, not installed)
	updated, _ := m.Update(keyMsg("space"))
	rows := updated.(Model).PackRows()
	if rows[0].Action != ActionInstall {
		t.Errorf("expected android Action=ActionInstall after space at cursor 0; got %v", rows[0].Action)
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
	if byName["kmp"].Action != ActionInstall {
		t.Errorf("kmp Action: got %v, want ActionInstall", byName["kmp"].Action)
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
		if r.Action != ActionNone || r.Required {
			t.Errorf("after toggle off, %s still has Action=%v Required=%v", r.Name, r.Action, r.Required)
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

// TestUpdate_EnterAutoSelectsCursorWhenNothingSelected verifies that pressing
// Enter from the picker with no pending actions auto-selects the pack under
// the cursor before transitioning to ModeConfirm. Matches user intuition
// ("estoy parado en kmp, le doy enter, instalo kmp").
func TestUpdate_EnterAutoSelectsCursorWhenNothingSelected(t *testing.T) {
	m := sampleModel(t)
	// Precondition: nothing pending, cursor at 0 (android in sampleModel).
	if hasPendingActions(m) {
		t.Fatalf("precondition: userActions should be empty, got %v", m.userActions)
	}

	updated, _ := m.Update(keyMsg("enter"))
	final := updated.(Model)

	if final.mode != ModeConfirm {
		t.Errorf("mode: got %v, want ModeConfirm", final.mode)
	}
	if final.userActions["android"] != ActionInstall {
		t.Errorf("Enter with no pending actions should auto-mark cursor row 'android' for Install; got userActions=%v", final.userActions)
	}
}

// hasPendingActions reports whether the model has any non-None action set.
func hasPendingActions(m Model) bool {
	for _, a := range m.userActions {
		if a != ActionNone {
			return true
		}
	}
	return false
}

// TestUpdate_EnterPreservesPreviousSelections verifies that if user already
// toggled some packs with space, Enter does NOT auto-add the cursor row —
// it just transitions to confirm with the explicit selection intact.
func TestUpdate_EnterPreservesPreviousSelections(t *testing.T) {
	m := sampleModel(t)
	// User explicitly selected ios. Cursor still at 0 (android).
	m.userActions = map[string]Action{"ios": ActionInstall}

	updated, _ := m.Update(keyMsg("enter"))
	final := updated.(Model)

	if final.mode != ModeConfirm {
		t.Errorf("mode: got %v, want ModeConfirm", final.mode)
	}
	if final.userActions["ios"] != ActionInstall {
		t.Errorf("ios selection should persist; got userActions=%v", final.userActions)
	}
	if final.userActions["android"] != ActionNone {
		t.Errorf("android should NOT be auto-added when user already selected something else; got userActions=%v", final.userActions)
	}
}

// TestToggle_OnInstalledPack_MarksForUninstall verifies that pressing space
// on a pack already installed in all detected hosts cycles the Action to
// Uninstall (visual `[-]`), so the user can remove it via the picker.
func TestToggle_OnInstalledPack_MarksForUninstall(t *testing.T) {
	m := sampleModel(t)
	// android is currently installed in the only host (stub). Cursor at 0.
	m.state.Packs["android"] = []string{"stub"}

	updated, _ := m.Update(keyMsg("space"))
	final := updated.(Model)

	rows := final.PackRows()
	if rows[0].Name != "android" {
		t.Fatalf("precondition: expected cursor on android, got %s", rows[0].Name)
	}
	if rows[0].Action != ActionUninstall {
		t.Errorf("space on installed pack should mark for Uninstall; got Action=%v", rows[0].Action)
	}
}

// TestToggle_OnUninstallAction_ResetsToNone verifies that pressing space
// twice on an installed pack (Install→Uninstall→None) cycles back without
// committing accidentally.
func TestToggle_OnUninstallAction_ResetsToNone(t *testing.T) {
	m := sampleModel(t)
	m.state.Packs["android"] = []string{"stub"}

	// First space: ActionUninstall
	updated, _ := m.Update(keyMsg("space"))
	m = updated.(Model)
	// Second space: ActionNone
	updated, _ = m.Update(keyMsg("space"))
	m = updated.(Model)

	if m.userActions["android"] != ActionNone {
		t.Errorf("two spaces on installed pack should cycle back to None; got %v", m.userActions["android"])
	}
}

// TestBuildInstallPlan_MixedInstallAndUninstall verifies that when the user
// has both Install and Uninstall actions pending, the plan contains the
// right step kinds for each (pack, host) pair.
func TestBuildInstallPlan_MixedInstallAndUninstall(t *testing.T) {
	m := sampleModel(t)
	m.state.Packs["android"] = []string{"stub"} // currently installed
	m.userActions = map[string]Action{
		"ios":     ActionInstall,
		"android": ActionUninstall,
	}

	plan := m.buildInstallPlan()
	if len(plan) == 0 {
		t.Fatal("expected non-empty plan")
	}
	byPackKind := map[string]stepKind{}
	for _, s := range plan {
		byPackKind[s.pack] = s.kind
	}
	if byPackKind["android"] != stepUninstall {
		t.Errorf("android step kind: got %v, want stepUninstall", byPackKind["android"])
	}
	// ios install also pulls core (transitive dep).
	if byPackKind["ios"] != stepInstall {
		t.Errorf("ios step kind: got %v, want stepInstall", byPackKind["ios"])
	}
}

// TestHandleInstallStep_UninstallSuccess_RemovesFromState verifies that a
// successful uninstall step calls state.Remove (so installed.json reflects
// the removal).
func TestHandleInstallStep_UninstallSuccess_RemovesFromState(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", tmp)
	paths, _ := state.NewPaths()
	if err := paths.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	c := &catalog.Catalog{Marketplace: &catalog.Marketplace{}}
	c.Reindex()
	installed := &state.Installed{Packs: map[string][]string{"android": {"stub"}}}
	m := NewModel(c, installed, paths, []host.HostAdapter{stubHost{}})
	m.mode = ModeProgress
	m.installPlan = []installStep{{pack: "android", host: "stub", kind: stepUninstall}}

	updated, _ := m.Update(installStepDoneMsg{pack: "android", host: "stub", kind: stepUninstall})
	final := updated.(Model)

	if final.mode != ModeResult {
		t.Errorf("mode: got %v, want ModeResult", final.mode)
	}
	hosts := final.state.HostsFor("android")
	if len(hosts) != 0 {
		t.Errorf("after Uninstall step, android should be gone from state; got hosts=%v", hosts)
	}
}

func TestUpdate_ConfirmYGoesToProgress(t *testing.T) {
	m := sampleModel(t)
	m.mode = ModeConfirm
	m.userActions = map[string]Action{"android": ActionInstall}
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
