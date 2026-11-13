package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrainHook_FileContext(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	writeCardFile(t, root, "bugfixes", "login-crash-20260914-163044",
		"## Login crash\n\n- id: login-crash-20260914-163044\n- type: bug_fix\n- status: active\n- date: 2026-09-14\n\nKeep state in the ViewModel.\n\n### Files\n- app/src/auth/LoginViewModel.kt\n")

	payload, _ := json.Marshal(map[string]any{
		"session_id":      "cmd-test-" + filepath.Base(root),
		"hook_event_name": "PostToolUse",
		"tool_name":       "Edit",
		"tool_input":      map[string]any{"file_path": filepath.Join(root, "app", "src", "auth", "LoginScreen.kt")},
	})
	cmd := NewBrainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(bytes.NewReader(payload))
	cmd.SetArgs([]string{"hook", "file-context"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not hook JSON: %v\n%s", err, out.String())
	}
	if !strings.Contains(got.HookSpecificOutput.AdditionalContext, "Keep state in the ViewModel.") {
		t.Errorf("context = %q", got.HookSpecificOutput.AdditionalContext)
	}
}

func TestBrainHook_SilentOnGarbage(t *testing.T) {
	cmd := NewBrainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader("garbage"))
	cmd.SetArgs([]string{"hook", "file-context"})
	if err := cmd.Execute(); err != nil || out.Len() != 0 {
		t.Errorf("a hook must never fail or print on bad input: err=%v out=%q", err, out.String())
	}
}
