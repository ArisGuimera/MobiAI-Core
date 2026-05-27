package host

import (
	"os"
	"path/filepath"
	"testing"
)

type adapterCase struct {
	factory      func() HostAdapter
	wantID       string
	wantName     string
	wantHomeSub  string
	wantHooks    bool
	wantCommands bool
}

func adapterCases() []adapterCase {
	return []adapterCase{
		// Tier 1
		{newClaudeCode, "claude-code", "Claude Code", ".claude", true, true},
		{newCursor, "cursor", "Cursor", ".cursor", true, true},
		{newGemini, "gemini", "Gemini CLI", ".gemini", false, false},
		{newCodex, "codex", "Codex", ".codex", false, false},
		// Tier 2 (best-effort)
		{newGoose, "goose", "Goose", ".config/goose", false, false},
		{newGitHubCopilot, "github-copilot", "GitHub Copilot", ".copilot", false, false},
		{newOpenCode, "opencode", "OpenCode", ".opencode", false, false},
		{newJunie, "junie", "Junie", ".junie", false, false},
		{newRooCode, "roo-code", "Roo Code", ".roo-code", false, false},
		{newFirebender, "firebender", "Firebender", ".firebender", false, false},
		// Tier 3 (speculative, paths a confirmar)
		{newAutohand, "autohand", "Autohand", ".autohand", false, false},
		{newOpenHands, "openhands", "OpenHands", ".openhands", false, false},
		{newMux, "mux", "Mux", ".mux", false, false},
		{newAmp, "amp", "Amp", ".amp", false, false},
		{newLetta, "letta", "Letta", ".letta", false, false},
		{newClaudeDesktop, "claude-desktop", "Claude Desktop", ".claude-desktop", false, false},
		{newPiebald, "piebald", "Piebald", ".piebald", false, false},
		{newFactory, "factory", "Factory", ".factory", false, false},
		{newPi, "pi", "pi", ".pi", false, false},
		{newDatabricksGenie, "databricks-genie", "Databricks Genie Code", ".databricks/genie-code", false, false},
		{newAgentman, "agentman", "Agentman", ".agentman", false, false},
		{newTRAE, "trae", "TRAE", ".trae", false, false},
		{newSpringAI, "spring-ai", "Spring AI", ".spring-ai", false, false},
		{newMistralVibe, "mistral-vibe", "Mistral AI Vibe", ".mistral-vibe", false, false},
		{newCommandCode, "command-code", "Command Code", ".command-code", false, false},
		{newOna, "ona", "Ona", ".ona", false, false},
		{newVTCode, "vt-code", "VT Code", ".vt-code", false, false},
		{newQodo, "qodo", "Qodo", ".qodo", false, false},
		{newLaravelBoost, "laravel-boost", "Laravel Boost", ".config/laravel-boost", false, false},
		{newEmdash, "emdash", "Emdash", ".emdash", false, false},
		{newSnowflakeCortex, "snowflake-cortex", "Snowflake Cortex Code", ".snowflake/cortex-code", false, false},
		{newKiro, "kiro", "Kiro", ".kiro", false, false},
		{newWorkshop, "workshop", "Workshop", ".workshop", false, false},
		{newAIEdgeGallery, "ai-edge-gallery", "Google AI Edge Gallery", ".google/ai-edge-gallery", false, false},
		{newNanobot, "nanobot", "nanobot", ".nanobot", false, false},
		{newFastAgent, "fast-agent", "fast-agent", ".fast-agent", false, false},
	}
}

func TestAdapters_Identity(t *testing.T) {
	for _, c := range adapterCases() {
		t.Run(c.wantID, func(t *testing.T) {
			a := c.factory()
			if a.ID() != c.wantID {
				t.Errorf("ID: got %q, want %q", a.ID(), c.wantID)
			}
			if a.Name() != c.wantName {
				t.Errorf("Name: got %q, want %q", a.Name(), c.wantName)
			}
			if a.Homepage() == "" {
				t.Error("Homepage empty")
			}
		})
	}
}

func TestAdapters_SkillsDir(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	for _, c := range adapterCases() {
		t.Run(c.wantID, func(t *testing.T) {
			if c.wantID == "firebender" {
				t.Skip("firebender uses project dir, tested separately")
			}
			want := filepath.Join(tmp, c.wantHomeSub, "skills")
			if got := c.factory().SkillsDir(); got != want {
				t.Errorf("SkillsDir: got %q, want %q", got, want)
			}
		})
	}
}

func TestFirebenderAdapter_SkillsDir_UsesProjectDir(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	// macOS: os.Getwd() resolves /var to /private/var; want must match.
	resolvedTmp, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatal(err)
	}

	a := newFirebender()
	want := filepath.Join(resolvedTmp, ".firebender", "skills")
	if got := a.SkillsDir(); got != want {
		t.Errorf("SkillsDir: got %q, want %q", got, want)
	}
}

func TestAdapters_Capabilities(t *testing.T) {
	for _, c := range adapterCases() {
		t.Run(c.wantID, func(t *testing.T) {
			caps := c.factory().Capabilities()
			if !caps.Skills {
				t.Error("Skills should be true")
			}
			if caps.Hooks != c.wantHooks {
				t.Errorf("Hooks: got %v, want %v", caps.Hooks, c.wantHooks)
			}
			if caps.Commands != c.wantCommands {
				t.Errorf("Commands: got %v, want %v", caps.Commands, c.wantCommands)
			}
		})
	}
}

func TestAdapters_DetectFlow(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	for _, c := range adapterCases() {
		t.Run(c.wantID+"_not_found", func(t *testing.T) {
			if c.factory().Detect().Found {
				t.Error("Detect should be false on empty home")
			}
		})
		t.Run(c.wantID+"_found", func(t *testing.T) {
			if err := os.MkdirAll(filepath.Join(tmp, c.wantHomeSub), 0o755); err != nil {
				t.Fatal(err)
			}
			if !c.factory().Detect().Found {
				t.Errorf("Detect should be true after creating %s", c.wantHomeSub)
			}
		})
	}
}
