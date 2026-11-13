package brain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// claudeInput builds the PostToolUse stdin Claude Code sends for a file tool.
func claudeInput(t *testing.T, session, file string) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"session_id":      session,
		"cwd":             filepath.Dir(file),
		"hook_event_name": "PostToolUse",
		"tool_name":       "Read",
		"tool_input":      map[string]any{"file_path": file},
		"tool_response":   map[string]any{},
	})
	return b
}

func claudeContext(t *testing.T, out []byte) string {
	t.Helper()
	if len(out) == 0 {
		return ""
	}
	var got struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("hook output is not the Claude Code JSON: %v\n%s", err, out)
	}
	if got.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("hookEventName = %q", got.HookSpecificOutput.HookEventName)
	}
	return got.HookSpecificOutput.AdditionalContext
}

func hookOpts(t *testing.T) HookOptions {
	return HookOptions{StateDir: t.TempDir()}
}

func authCard(id, title string) string {
	return "## " + title + "\n\n- id: " + id + "\n- type: bug_fix\n- status: active\n- date: 2026-09-14\n- verified: 2026-09-14\n\n" +
		title + " body.\n\n### Files\n- app/src/auth/LoginViewModel.kt\n"
}

func TestHook_NoBrainNoOutput(t *testing.T) {
	dir := t.TempDir()
	out := FileContextHook(claudeInput(t, "s1", filepath.Join(dir, "app", "A.kt")), hookOpts(t))
	if len(out) != 0 {
		t.Errorf("no brain → no output, got %s", out)
	}
}

func TestHook_DeliversNotesForFolder_ClaudeFormat(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "login-crash-20260914-163044", authCard("login-crash-20260914-163044", "Login crash on rotation"))
	writeCard(t, p, "bugfixes", "net-20260914-163044", strings.Replace(authCard("net-20260914-163044", "Network retry"), "app/src/auth/", "app/src/net/", 1))

	out := FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "app", "src", "auth", "LoginScreen.kt")), hookOpts(t))
	ctx := claudeContext(t, out)
	for _, want := range []string{"app/src/auth/LoginScreen.kt", "## Login crash on rotation", "Login crash on rotation body."} {
		if !strings.Contains(ctx, want) {
			t.Errorf("missing %q in:\n%s", want, ctx)
		}
	}
	if strings.Contains(ctx, "Network retry") {
		t.Errorf("a note for another folder must not be delivered:\n%s", ctx)
	}
}

func TestHook_CursorFormat(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "login-crash-20260914-163044", authCard("login-crash-20260914-163044", "Login crash"))
	in, _ := json.Marshal(map[string]any{
		"conversation_id": "c1",
		"hook_event_name": "postToolUse",
		"tool_name":       "Write",
		"tool_input":      map[string]any{"path": filepath.Join(p.Root, "app", "src", "auth", "X.kt")},
	})
	out := FileContextHook(in, hookOpts(t))
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	ctx, _ := got["additional_context"].(string)
	if !strings.Contains(ctx, "## Login crash") || len(got) != 1 {
		t.Errorf("Cursor output = %s", out)
	}
}

func TestHook_OncePerSession(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "login-crash-20260914-163044", authCard("login-crash-20260914-163044", "Login crash"))
	opts := hookOpts(t)
	file := filepath.Join(p.Root, "app", "src", "auth", "X.kt")
	if ctx := claudeContext(t, FileContextHook(claudeInput(t, "s1", file), opts)); ctx == "" {
		t.Fatal("first call must deliver the note")
	}
	if out := FileContextHook(claudeInput(t, "s1", file), opts); len(out) != 0 {
		t.Errorf("second call in the same session must say nothing, got %s", out)
	}
	if ctx := claudeContext(t, FileContextHook(claudeInput(t, "s2", file), opts)); ctx == "" {
		t.Errorf("another session gets the note again")
	}
}

func TestHook_CapsSize(t *testing.T) {
	p := initBrainForTest(t)
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		id := "note-" + n + "-20260914-163044"
		card := strings.Replace(authCard(id, "Note "+n), "body.", strings.Repeat("y", 1500), 1)
		writeCard(t, p, "bugfixes", id, card)
	}
	opts := hookOpts(t)
	opts.MaxBytes = 4000
	ctx := claudeContext(t, FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "app", "src", "auth", "X.kt")), opts))
	if len(ctx) > 4300 {
		t.Errorf("context is %d bytes, over the cap", len(ctx))
	}
	if !strings.Contains(ctx, "More notes apply to this folder") || !strings.Contains(ctx, "mobiai brain show") {
		t.Errorf("the notes that don't fit must be listed by id:\n%s", ctx)
	}
}

func TestHook_DoesNotMigrate(t *testing.T) {
	root := t.TempDir()
	p := oldBrain(t, root, map[string]string{"decisions.md": oldDecision})
	before := snapshot(t, p.MemoriesDir)
	ctx := claudeContext(t, FileContextHook(claudeInput(t, "s1", filepath.Join(root, "composeApp", "src", "commonMain", "App.kt")), hookOpts(t)))
	if !strings.Contains(ctx, "## Use Koin as DI") {
		t.Errorf("an old-format note must still be delivered:\n%s", ctx)
	}
	after := snapshot(t, p.MemoriesDir)
	if len(before) != len(after) {
		t.Errorf("the hook must not write: %d files before, %d after", len(before), len(after))
	}
	for name, c := range before {
		if after[name] != c {
			t.Errorf("the hook changed %s", name)
		}
	}
}

func TestHook_PathsGlobOverride(t *testing.T) {
	p := initBrainForTest(t)
	card := "## Checkout rounding\n\n- id: checkout-20260914-163044\n- type: knowledge\n- status: active\n- date: 2026-09-14\n- paths: feature/**/checkout/*.kt\n\nRound half-even.\n"
	writeCard(t, p, "knowledge", "checkout-20260914-163044", card)
	opts := hookOpts(t)
	if ctx := claudeContext(t, FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "feature", "pay", "ui", "checkout", "Cart.kt")), opts)); !strings.Contains(ctx, "Checkout rounding") {
		t.Errorf("** must match nested folders:\n%s", ctx)
	}
	if out := FileContextHook(claudeInput(t, "s2", filepath.Join(p.Root, "feature", "pay", "Cart.kt")), opts); len(out) != 0 {
		t.Errorf("a file outside the glob must not match: %s", out)
	}
}

func TestHook_RootFileDoesNotMatchEverything(t *testing.T) {
	p := initBrainForTest(t)
	card := "## Gradle cache\n\n- id: gradle-20260914-163044\n- type: bug_fix\n- status: active\n- date: 2026-09-14\n\nBody.\n\n### Files\n- build.gradle.kts\n"
	writeCard(t, p, "bugfixes", "gradle-20260914-163044", card)
	opts := hookOpts(t)
	if out := FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "app", "A.kt")), opts); len(out) != 0 {
		t.Errorf("a root file must not apply to every file: %s", out)
	}
	if ctx := claudeContext(t, FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "build.gradle.kts")), opts)); !strings.Contains(ctx, "Gradle cache") {
		t.Errorf("the root file itself matches:\n%s", ctx)
	}
}

func TestHook_SkipsDeprecated(t *testing.T) {
	p := initBrainForTest(t)
	writeCard(t, p, "bugfixes", "old-20260914-163044", strings.Replace(authCard("old-20260914-163044", "Old"), "status: active", "status: deprecated", 1))
	if out := FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "app", "src", "auth", "X.kt")), hookOpts(t)); len(out) != 0 {
		t.Errorf("deprecated notes are not delivered: %s", out)
	}
}

func TestHook_MarksUnconfirmedOncePerSession(t *testing.T) {
	root, p := gitBrain(t)
	gitRun(t, root, "checkout", "-qb", "feature")
	writeCard(t, p, "bugfixes", "login-crash-20260914-163044", authCard("login-crash-20260914-163044", "Login crash"))
	writeCard(t, p, "bugfixes", "net-20260914-163044", strings.Replace(authCard("net-20260914-163044", "Network retry"), "app/src/auth/", "app/src/net/", 1))
	commitAs(t, root, "other@example.test", "someone else's notes")

	orig := runGit
	defer func() { runGit = orig }()
	calls := 0
	runGit = func(dir, stdin string, args ...string) (string, error) {
		calls++
		return orig(dir, stdin, args...)
	}
	opts := hookOpts(t)
	ctx := claudeContext(t, FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "app", "src", "auth", "X.kt")), opts))
	if !strings.Contains(ctx, "- state: unconfirmed") {
		t.Errorf("an unconfirmed note must be marked in the hook too:\n%s", ctx)
	}
	first := calls
	if first == 0 {
		t.Fatal("the first call of a session must check confirmation with git")
	}
	ctx = claudeContext(t, FileContextHook(claudeInput(t, "s1", filepath.Join(p.Root, "app", "src", "net", "Api.kt")), opts))
	if !strings.Contains(ctx, "- state: unconfirmed") {
		t.Errorf("second call lost the mark:\n%s", ctx)
	}
	if calls != first {
		t.Errorf("later calls in the session must not run git again (%d extra calls)", calls-first)
	}
}

func TestHook_BadInputIsSilent(t *testing.T) {
	for _, in := range []string{"", "not json", `{"tool_input":{}}`, `{"tool_input":{"file_path":"relative/x.kt"}}`} {
		if out := FileContextHook([]byte(in), hookOpts(t)); len(out) != 0 {
			t.Errorf("input %q → %s, want silence", in, out)
		}
	}
}

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern, name string
		want          bool
	}{
		{"feature/**/checkout/*.kt", "feature/a/b/checkout/X.kt", true},
		{"feature/**/checkout/*.kt", "feature/checkout/X.kt", true},
		{"feature/**/checkout/*.kt", "feature/a/X.kt", false},
		{"core/net/", "core/net/Api.kt", true},
		{"core/net", "core/net/sub/Api.kt", true},
		{"core/net", "core/network/Api.kt", false},
		{"**/*.gradle.kts", "app/build.gradle.kts", true},
		{"build.gradle.kts", "build.gradle.kts", true},
	}
	for _, c := range cases {
		if got := matchAppliesTo(c.pattern, c.name); got != c.want {
			t.Errorf("matchAppliesTo(%q, %q) = %v, want %v", c.pattern, c.name, got, c.want)
		}
	}
}

func TestHooksJSON_Valid(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "skills", "core", "hooks")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("skills/core/hooks not found from the test directory")
	}
	for _, name := range []string{"hooks.json", "hooks-cursor.json"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		var v map[string]any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Errorf("%s is not valid JSON: %v", name, err)
		}
		if !strings.Contains(string(data), "brain-file-context") {
			t.Errorf("%s does not register the file-context hook", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "brain-file-context")); err != nil {
		t.Errorf("hook script missing: %v", err)
	}
}
