package host

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

// claudeCodeMinSkillBudget is the minimum value we ensure for
// skillListingBudgetFraction in ~/.claude/settings.json. Claude Code's default
// (1%) truncates skill descriptions once the user has roughly 20+ skills, and
// MobiAI's full catalog crosses that threshold. 3% (~6k tokens on a 200k
// context) fits every current pack with headroom.
const claudeCodeMinSkillBudget = 0.03

// newClaudeCode returns a HostAdapter for Anthropic Claude Code.
// Claude Code stores skills under ~/.claude/skills/ and additionally
// supports hooks and slash commands; those bonus capabilities are flagged
// here but actual hook installation is deferred to a later phase.
func newClaudeCode() HostAdapter {
	return &claudeCodeAdapter{
		genericAdapter: &genericAdapter{
			id:         "claude-code",
			name:       "Claude Code",
			homepage:   "https://claude.com/code",
			homeSubdir: ".claude",
			caps:       Caps{Skills: true, Hooks: true, Commands: true, MCPs: true},
		},
	}
}

// claudeCodeAdapter extends genericAdapter with a post-install step that
// raises Claude Code's skillListingBudgetFraction so the full MobiAI skill
// catalog lists without truncation.
type claudeCodeAdapter struct {
	*genericAdapter
}

func (a *claudeCodeAdapter) Install(skills []catalog.Skill) error {
	if err := a.genericAdapter.Install(skills); err != nil {
		return err
	}
	dir, err := a.hostDir()
	if err != nil {
		return nil
	}
	settingsPath := filepath.Join(dir, "settings.json")
	bumped, berr := ensureSkillListingBudget(settingsPath, claudeCodeMinSkillBudget)
	if berr != nil {
		fmt.Fprintf(os.Stderr, "  aviso: no pude ajustar skillListingBudgetFraction en %s: %v\n", settingsPath, berr)
		return nil
	}
	if bumped {
		fmt.Fprintf(os.Stderr, "  aviso: subí skillListingBudgetFraction a %.0f%% en %s para listar todas las skills sin truncar\n", claudeCodeMinSkillBudget*100, settingsPath)
	}
	return nil
}

// ensureSkillListingBudget raises skillListingBudgetFraction in settingsPath
// to min when the current value is missing or lower. Returns (bumped, error):
// bumped=true means the file was rewritten, false means no change was needed.
// Safe to call repeatedly; the value comparison short-circuits second calls.
func ensureSkillListingBudget(settingsPath string, min float64) (bool, error) {
	settings := map[string]any{}
	raw, err := os.ReadFile(settingsPath)
	switch {
	case err == nil:
		if uerr := json.Unmarshal(raw, &settings); uerr != nil {
			return false, fmt.Errorf("parse settings: %w", uerr)
		}
	case os.IsNotExist(err):
		// First-time install — create a fresh settings.json with just the field.
	default:
		return false, fmt.Errorf("read settings: %w", err)
	}

	if current, ok := settings["skillListingBudgetFraction"].(float64); ok && current >= min {
		return false, nil
	}
	settings["skillListingBudgetFraction"] = min

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return false, fmt.Errorf("mkdir: %w", err)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(settings); err != nil {
		return false, fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(settingsPath, buf.Bytes(), 0o644); err != nil {
		return false, fmt.Errorf("write settings: %w", err)
	}
	return true, nil
}
