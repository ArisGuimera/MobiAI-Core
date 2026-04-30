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
		{newClaudeCode, "claude-code", "Claude Code", ".claude", true, true},
		{newCursor, "cursor", "Cursor", ".cursor", true, true},
		{newGemini, "gemini", "Gemini CLI", ".gemini", false, false},
		{newCodex, "codex", "Codex", ".codex", false, false},
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
			want := filepath.Join(tmp, c.wantHomeSub, "skills")
			if got := c.factory().SkillsDir(); got != want {
				t.Errorf("SkillsDir: got %q, want %q", got, want)
			}
		})
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
