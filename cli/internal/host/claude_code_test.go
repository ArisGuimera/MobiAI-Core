package host

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

func readSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return got
}

func TestEnsureSkillListingBudget_FileMissing_CreatesWithField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	bumped, err := ensureSkillListingBudget(path, 0.03)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bumped {
		t.Fatal("expected bumped=true on missing file")
	}
	got := readSettings(t, path)
	if v := got["skillListingBudgetFraction"]; v != 0.03 {
		t.Errorf("field: got %v, want 0.03", v)
	}
}

func TestEnsureSkillListingBudget_FieldAbsent_AddsField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	original := map[string]any{
		"permissions": map[string]any{"defaultMode": "default"},
		"effortLevel": "xhigh",
	}
	raw, _ := json.Marshal(original)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	bumped, err := ensureSkillListingBudget(path, 0.03)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bumped {
		t.Fatal("expected bumped=true when field absent")
	}
	got := readSettings(t, path)
	if v := got["skillListingBudgetFraction"]; v != 0.03 {
		t.Errorf("field: got %v, want 0.03", v)
	}
	if v := got["effortLevel"]; v != "xhigh" {
		t.Errorf("effortLevel preserved: got %v", v)
	}
	if _, ok := got["permissions"]; !ok {
		t.Error("permissions key lost")
	}
}

func TestEnsureSkillListingBudget_FieldBelowMin_Raises(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	original := map[string]any{"skillListingBudgetFraction": 0.01}
	raw, _ := json.Marshal(original)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	bumped, err := ensureSkillListingBudget(path, 0.03)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bumped {
		t.Fatal("expected bumped=true when below min")
	}
	got := readSettings(t, path)
	if v := got["skillListingBudgetFraction"]; v != 0.03 {
		t.Errorf("field: got %v, want 0.03", v)
	}
}

func TestEnsureSkillListingBudget_FieldEqualToMin_NoOp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	original := map[string]any{"skillListingBudgetFraction": 0.03}
	raw, _ := json.Marshal(original)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	beforeStat, _ := os.Stat(path)

	bumped, err := ensureSkillListingBudget(path, 0.03)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if bumped {
		t.Fatal("expected bumped=false when field already at min")
	}
	afterStat, _ := os.Stat(path)
	if !afterStat.ModTime().Equal(beforeStat.ModTime()) {
		t.Error("file was rewritten despite no-op")
	}
}

func TestEnsureSkillListingBudget_FieldAboveMin_NoOp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	original := map[string]any{"skillListingBudgetFraction": 0.05}
	raw, _ := json.Marshal(original)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	bumped, err := ensureSkillListingBudget(path, 0.03)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if bumped {
		t.Fatal("expected bumped=false when field already above min")
	}
	got := readSettings(t, path)
	if v := got["skillListingBudgetFraction"]; v != 0.05 {
		t.Errorf("user value clobbered: got %v, want 0.05", v)
	}
}

func TestEnsureSkillListingBudget_InvalidJSON_ReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ensureSkillListingBudget(path, 0.03); err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
}

func TestClaudeCodeAdapter_Install_BumpsBudget(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)

	a := newClaudeCode()
	if err := a.Install([]catalog.Skill{makeFakeSkill(t, "demo")}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	got := readSettings(t, settingsPath)
	if v := got["skillListingBudgetFraction"]; v != claudeCodeMinSkillBudget {
		t.Errorf("budget after install: got %v, want %v", v, claudeCodeMinSkillBudget)
	}
}

func TestClaudeCodeAdapter_Install_PreservesUserBudgetIfHigher(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	claudeDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	raw, _ := json.Marshal(map[string]any{"skillListingBudgetFraction": 0.10})
	if err := os.WriteFile(settingsPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	a := newClaudeCode()
	if err := a.Install([]catalog.Skill{makeFakeSkill(t, "demo")}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	got := readSettings(t, settingsPath)
	if v := got["skillListingBudgetFraction"]; v != 0.10 {
		t.Errorf("user budget clobbered: got %v, want 0.10", v)
	}
}
