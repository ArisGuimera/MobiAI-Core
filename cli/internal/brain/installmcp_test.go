package brain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeBinary is the path we pretend mobiai lives at across these
// tests. Using a fixed string keeps the JSON assertions readable.
const fakeBinary = "/usr/local/bin/mobiai"

// makeFakeHome builds a TempDir that looks like $HOME and returns the
// path. By default both ~/.claude and ~/.cursor exist (so install
// doesn't skip them) — individual tests can wipe one to exercise the
// skipped-client path.
func makeFakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	for _, sub := range []string{".claude", ".cursor"} {
		if err := os.MkdirAll(filepath.Join(home, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

func writeJSON(t *testing.T, path string, v map[string]interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("parse %s: %v\n%s", path, err, data)
	}
	return out
}

func TestInstallMCP_CreatesEntryWhenSettingsExisted(t *testing.T) {
	home := makeFakeHome(t)
	// User already has a settings.json with unrelated fields.
	settings := filepath.Join(home, ".claude", "settings.json")
	writeJSON(t, settings, map[string]interface{}{
		"alwaysThinkingEnabled": true,
	})

	results, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Action != ActionInstalled {
		t.Fatalf("unexpected results: %+v", results)
	}

	got := readJSON(t, settings)
	// Other fields preserved.
	if got["alwaysThinkingEnabled"] != true {
		t.Errorf("alwaysThinkingEnabled lost: %+v", got)
	}
	// MCP entry present and shaped correctly.
	servers, _ := got["mcpServers"].(map[string]interface{})
	entry, _ := servers["mobiai-brain"].(map[string]interface{})
	if entry["command"] != fakeBinary {
		t.Errorf("command = %v, want %s", entry["command"], fakeBinary)
	}
	args, _ := entry["args"].([]interface{})
	if !reflect.DeepEqual(args, []interface{}{"brain", "mcp"}) {
		t.Errorf("args = %v, want [brain mcp]", args)
	}
}

func TestInstallMCP_CreatesFileWhenMissingButDirExists(t *testing.T) {
	// .claude/ exists but settings.json doesn't — common state for a
	// fresh Claude Code install.
	home := makeFakeHome(t)
	settings := filepath.Join(home, ".claude", "settings.json")

	results, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionInstalled {
		t.Fatalf("want ActionInstalled, got %s", results[0].Action)
	}
	got := readJSON(t, settings)
	servers, _ := got["mcpServers"].(map[string]interface{})
	if _, ok := servers["mobiai-brain"]; !ok {
		t.Errorf("mobiai-brain missing from created file: %+v", got)
	}
}

func TestInstallMCP_SkipsWhenClientDirMissing(t *testing.T) {
	// Build a home with only ~/.claude/ — Cursor isn't installed.
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	results, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientCursor},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionSkipped {
		t.Errorf("want ActionSkipped for missing client dir, got %s", results[0].Action)
	}
	// And no file was created sneakily.
	if _, err := os.Stat(filepath.Join(home, ".cursor", "mcp.json")); !os.IsNotExist(err) {
		t.Errorf("install should not create files for non-existent clients (got err=%v)", err)
	}
}

func TestInstallMCP_IdempotentOnSecondRun(t *testing.T) {
	home := makeFakeHome(t)
	opts := InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	}
	// First run installs.
	first, err := InstallMCP(opts)
	if err != nil {
		t.Fatal(err)
	}
	if first[0].Action != ActionInstalled {
		t.Fatalf("first run: want Installed, got %s", first[0].Action)
	}
	// Second run with same inputs should report Unchanged.
	second, err := InstallMCP(opts)
	if err != nil {
		t.Fatal(err)
	}
	if second[0].Action != ActionUnchanged {
		t.Errorf("second run: want Unchanged, got %s", second[0].Action)
	}
}

func TestInstallMCP_UpdatesWhenBinaryPathChanges(t *testing.T) {
	home := makeFakeHome(t)
	// Pretend we previously installed pointing at /old/path/mobiai.
	settings := filepath.Join(home, ".claude", "settings.json")
	writeJSON(t, settings, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"mobiai-brain": map[string]interface{}{
				"command": "/old/path/mobiai",
				"args":    []interface{}{"brain", "mcp"},
			},
		},
	})

	results, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionUpdated {
		t.Errorf("want Updated when binary path differs, got %s", results[0].Action)
	}
	got := readJSON(t, settings)
	servers := got["mcpServers"].(map[string]interface{})
	entry := servers["mobiai-brain"].(map[string]interface{})
	if entry["command"] != fakeBinary {
		t.Errorf("command not updated: got %v", entry["command"])
	}
}

func TestInstallMCP_PreservesOtherMCPServers(t *testing.T) {
	home := makeFakeHome(t)
	settings := filepath.Join(home, ".claude", "settings.json")
	writeJSON(t, settings, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"some-other-server": map[string]interface{}{
				"command": "/usr/bin/some-other-thing",
				"args":    []interface{}{"--serve"},
			},
		},
	})

	_, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, settings)
	servers := got["mcpServers"].(map[string]interface{})
	if _, ok := servers["mobiai-brain"]; !ok {
		t.Errorf("mobiai-brain not added: %+v", servers)
	}
	other, ok := servers["some-other-server"].(map[string]interface{})
	if !ok {
		t.Fatalf("other server lost: %+v", servers)
	}
	if other["command"] != "/usr/bin/some-other-thing" {
		t.Errorf("other server clobbered: %+v", other)
	}
}

func TestInstallMCP_DryRunDoesNotWrite(t *testing.T) {
	home := makeFakeHome(t)
	settings := filepath.Join(home, ".claude", "settings.json")

	results, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
		DryRun:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionInstalled || !results[0].DryRun {
		t.Errorf("dry-run result wrong: %+v", results[0])
	}
	if _, err := os.Stat(settings); !os.IsNotExist(err) {
		t.Errorf("dry-run should not write file (got err=%v)", err)
	}
}

func TestInstallMCP_AllClientsByDefault(t *testing.T) {
	home := makeFakeHome(t)
	results, err := InstallMCP(InstallOptions{
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should have processed both Claude Code and Cursor.
	seen := map[MCPClient]bool{}
	for _, r := range results {
		seen[r.Client] = true
	}
	if !seen[MCPClientClaudeCode] || !seen[MCPClientCursor] {
		t.Errorf("expected both clients to be processed; got %+v", results)
	}
}

func TestUninstallMCP_RemovesEntry(t *testing.T) {
	home := makeFakeHome(t)
	settings := filepath.Join(home, ".claude", "settings.json")
	writeJSON(t, settings, map[string]interface{}{
		"alwaysThinkingEnabled": true,
		"mcpServers": map[string]interface{}{
			"mobiai-brain": map[string]interface{}{"command": fakeBinary, "args": []interface{}{"brain", "mcp"}},
			"keep-this":    map[string]interface{}{"command": "/x"},
		},
	})

	results, err := UninstallMCP(InstallOptions{
		Clients: []MCPClient{MCPClientClaudeCode},
		HomeDir: home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionUninstalled {
		t.Fatalf("want ActionUninstalled, got %s", results[0].Action)
	}
	got := readJSON(t, settings)
	// Our entry is gone.
	servers := got["mcpServers"].(map[string]interface{})
	if _, ok := servers["mobiai-brain"]; ok {
		t.Errorf("mobiai-brain not removed: %+v", servers)
	}
	// Other entries preserved.
	if _, ok := servers["keep-this"]; !ok {
		t.Errorf("unrelated mcpServer was clobbered: %+v", servers)
	}
	// Unrelated top-level fields preserved.
	if got["alwaysThinkingEnabled"] != true {
		t.Errorf("alwaysThinkingEnabled lost: %+v", got)
	}
}

func TestUninstallMCP_NotPresentWhenFileMissing(t *testing.T) {
	home := makeFakeHome(t)
	results, err := UninstallMCP(InstallOptions{
		Clients: []MCPClient{MCPClientClaudeCode},
		HomeDir: home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionNotPresent {
		t.Errorf("want ActionNotPresent for missing file, got %s", results[0].Action)
	}
}

func TestUninstallMCP_NotPresentWhenEntryAbsent(t *testing.T) {
	home := makeFakeHome(t)
	settings := filepath.Join(home, ".claude", "settings.json")
	writeJSON(t, settings, map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"only-other-server": map[string]interface{}{"command": "/x"},
		},
	})
	results, err := UninstallMCP(InstallOptions{
		Clients: []MCPClient{MCPClientClaudeCode},
		HomeDir: home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionNotPresent {
		t.Errorf("want ActionNotPresent, got %s", results[0].Action)
	}
	// And we didn't accidentally delete the other server.
	got := readJSON(t, settings)
	servers := got["mcpServers"].(map[string]interface{})
	if _, ok := servers["only-other-server"]; !ok {
		t.Errorf("untouched server got removed")
	}
}

func TestInstallMCP_ErrorOnMalformedJSON(t *testing.T) {
	home := makeFakeHome(t)
	settings := filepath.Join(home, ".claude", "settings.json")
	if err := os.WriteFile(settings, []byte("{this isn't valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{MCPClientClaudeCode},
		BinaryPath: fakeBinary,
		HomeDir:    home,
	})
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
	if !strings.Contains(err.Error(), "JSON") {
		t.Errorf("error should mention JSON parsing: %v", err)
	}
}

func TestInstallMCP_UnknownClient(t *testing.T) {
	_, err := InstallMCP(InstallOptions{
		Clients:    []MCPClient{"copilot"}, // not yet supported
		BinaryPath: fakeBinary,
		HomeDir:    t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for unknown client")
	}
}

func TestInstallMCP_DefaultsBinaryToCurrentExecutable(t *testing.T) {
	// Pass no BinaryPath; resolveBinaryPath should fill it from
	// os.Executable. We can't assert the exact value (depends on
	// test runner), but it should be non-empty and absolute.
	home := makeFakeHome(t)
	results, err := InstallMCP(InstallOptions{
		Clients: []MCPClient{MCPClientClaudeCode},
		HomeDir: home,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Action != ActionInstalled {
		t.Fatalf("install failed: %+v", results[0])
	}
	got := readJSON(t, filepath.Join(home, ".claude", "settings.json"))
	servers := got["mcpServers"].(map[string]interface{})
	entry := servers["mobiai-brain"].(map[string]interface{})
	cmd, _ := entry["command"].(string)
	if cmd == "" || !filepath.IsAbs(cmd) {
		t.Errorf("command should be an absolute path, got %q", cmd)
	}
}
