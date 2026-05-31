package brain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// MCPServerName is the key under `mcpServers` that we add/remove. Kept
// in sync with the name advertised by the brain MCP server itself (see
// mcp.ServerName).
const MCPServerName = "mobiai-brain"

// MCPClient identifies a supported AI client. Phase 1 ships two
// (Claude Code, Cursor) — both happen to share the same JSON shape so
// the merge logic is identical. Future clients (Copilot, Codex,
// Gemini) can plug in by extending this enum and clientConfigPath.
type MCPClient string

const (
	MCPClientClaudeCode MCPClient = "claude"
	MCPClientCursor     MCPClient = "cursor"
)

// AllSupportedClients returns the clients we know how to register. The
// CLI uses this when --client wasn't given (detect everything we can).
func AllSupportedClients() []MCPClient {
	return []MCPClient{MCPClientClaudeCode, MCPClientCursor}
}

// InstallAction describes what would happen / did happen for one
// client during an install or uninstall operation. Used both by the
// CLI for human-readable output and by tests to verify behavior.
type InstallAction string

const (
	// ActionInstalled: entry was added to the client's config.
	ActionInstalled InstallAction = "installed"
	// ActionUpdated: an existing entry was rewritten because it pointed
	// to a different binary path or args.
	ActionUpdated InstallAction = "updated"
	// ActionUnchanged: the entry already matched the desired config.
	ActionUnchanged InstallAction = "unchanged"
	// ActionUninstalled: entry was removed from the client's config.
	ActionUninstalled InstallAction = "uninstalled"
	// ActionNotPresent: nothing to uninstall — the entry wasn't there.
	ActionNotPresent InstallAction = "not_present"
	// ActionSkipped: the client's config dir doesn't exist (probably
	// the client isn't installed). We don't create config files for
	// clients that aren't there — that would be invasive.
	ActionSkipped InstallAction = "skipped"
)

// InstallResult bundles per-client outcomes for one call. Multiple
// clients can be processed in one InstallMCP/UninstallMCP invocation.
type InstallResult struct {
	Client     MCPClient
	ConfigPath string        // absolute path to the file we touched (or would have)
	Action     InstallAction // what happened
	DryRun     bool          // true when --dry-run is set; no file was written
}

// InstallOptions configures InstallMCP and UninstallMCP. The CLI fills
// this from flags; tests pass it directly so HomeDir can be redirected
// to a TempDir.
type InstallOptions struct {
	// Clients to operate on. Empty means "all supported clients".
	Clients []MCPClient

	// BinaryPath is what we write into the `command` field of the
	// mcpServers entry. Empty means "use os.Executable() — the path of
	// the currently-running mobiai binary". Tests pass a fake path.
	BinaryPath string

	// HomeDir overrides $HOME for config lookups. Empty means
	// os.UserHomeDir(). Tests pass a TempDir.
	HomeDir string

	// DryRun: don't write anything to disk. Results still report what
	// would have happened so the CLI can preview accurately.
	DryRun bool
}

// InstallMCP registers `mobiai-brain` as an MCP server in each
// requested client's config file. Preserves any other fields in the
// JSON byte-as-best-as-possible (JSON re-marshaling drops original
// key ordering and whitespace, but no fields are lost).
//
// Idempotent: re-running with identical inputs yields ActionUnchanged.
// Returns one InstallResult per client requested.
func InstallMCP(opts InstallOptions) ([]InstallResult, error) {
	binary, err := resolveBinaryPath(opts.BinaryPath)
	if err != nil {
		return nil, err
	}
	home, err := resolveHomeDir(opts.HomeDir)
	if err != nil {
		return nil, err
	}
	clients := opts.Clients
	if len(clients) == 0 {
		clients = AllSupportedClients()
	}

	desired := mcpEntry(binary)
	results := make([]InstallResult, 0, len(clients))
	for _, c := range clients {
		path, err := clientConfigPath(c, home)
		if err != nil {
			return nil, err
		}
		action, err := applyInstall(path, desired, opts.DryRun)
		if err != nil {
			return nil, fmt.Errorf("[%s] %w", c, err)
		}
		results = append(results, InstallResult{
			Client:     c,
			ConfigPath: path,
			Action:     action,
			DryRun:     opts.DryRun,
		})
	}
	return results, nil
}

// UninstallMCP removes the `mobiai-brain` entry from each requested
// client's config. Files that didn't exist or didn't contain the entry
// yield ActionNotPresent (no error). Other fields of the file are
// preserved.
func UninstallMCP(opts InstallOptions) ([]InstallResult, error) {
	home, err := resolveHomeDir(opts.HomeDir)
	if err != nil {
		return nil, err
	}
	clients := opts.Clients
	if len(clients) == 0 {
		clients = AllSupportedClients()
	}

	results := make([]InstallResult, 0, len(clients))
	for _, c := range clients {
		path, err := clientConfigPath(c, home)
		if err != nil {
			return nil, err
		}
		action, err := applyUninstall(path, opts.DryRun)
		if err != nil {
			return nil, fmt.Errorf("[%s] %w", c, err)
		}
		results = append(results, InstallResult{
			Client:     c,
			ConfigPath: path,
			Action:     action,
			DryRun:     opts.DryRun,
		})
	}
	return results, nil
}

// clientConfigPath returns the absolute path to a client's MCP-bearing
// config file under the given home directory.
func clientConfigPath(c MCPClient, home string) (string, error) {
	switch c {
	case MCPClientClaudeCode:
		return filepath.Join(home, ".claude", "settings.json"), nil
	case MCPClientCursor:
		return filepath.Join(home, ".cursor", "mcp.json"), nil
	default:
		return "", fmt.Errorf("unknown client: %q (supported: claude, cursor)", c)
	}
}

// mcpEntry is the JSON value we install under mcpServers["mobiai-brain"].
// Returning a fresh map per call keeps callers from accidentally
// mutating shared state.
func mcpEntry(binary string) map[string]interface{} {
	return map[string]interface{}{
		"command": binary,
		"args":    []interface{}{"brain", "mcp"},
	}
}

// applyInstall reads the config at path (or starts from empty when
// missing), merges in the mobiai-brain entry, and writes back atomic
// when not in dry-run. Returns the action that classifies the outcome.
func applyInstall(path string, desired map[string]interface{}, dryRun bool) (InstallAction, error) {
	cfg, fileExisted, err := readJSONFile(path)
	if err != nil {
		return "", err
	}

	// If the client's config dir doesn't exist at all, skip — we don't
	// want to create a fake `.cursor/` for someone who doesn't use
	// Cursor. The user can always install Cursor and re-run.
	if !fileExisted {
		if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
			return ActionSkipped, nil
		}
	}

	servers, _ := cfg["mcpServers"].(map[string]interface{})
	if servers == nil {
		servers = map[string]interface{}{}
	}

	existing, ok := servers[MCPServerName]
	var action InstallAction
	switch {
	case !ok:
		action = ActionInstalled
	case reflect.DeepEqual(existing, desired):
		action = ActionUnchanged
	default:
		action = ActionUpdated
	}

	if action == ActionUnchanged {
		return action, nil
	}

	servers[MCPServerName] = desired
	cfg["mcpServers"] = servers

	if dryRun {
		return action, nil
	}
	if err := writeJSONFile(path, cfg); err != nil {
		return "", err
	}
	return action, nil
}

// applyUninstall reads the config, removes the mobiai-brain entry,
// and writes back atomic when not in dry-run.
func applyUninstall(path string, dryRun bool) (InstallAction, error) {
	cfg, fileExisted, err := readJSONFile(path)
	if err != nil {
		return "", err
	}
	if !fileExisted {
		return ActionNotPresent, nil
	}
	servers, _ := cfg["mcpServers"].(map[string]interface{})
	if servers == nil {
		return ActionNotPresent, nil
	}
	if _, ok := servers[MCPServerName]; !ok {
		return ActionNotPresent, nil
	}
	delete(servers, MCPServerName)
	// If mcpServers became empty after the removal, keep it as an
	// empty object rather than dropping the key. Removing it would
	// surprise users who later expect to see the same shape.
	cfg["mcpServers"] = servers

	if dryRun {
		return ActionUninstalled, nil
	}
	if err := writeJSONFile(path, cfg); err != nil {
		return "", err
	}
	return ActionUninstalled, nil
}

// readJSONFile loads path as a JSON object. Returns:
//   - ({}, false, nil) when the file doesn't exist
//   - ({}, true, nil) when the file exists but is empty
//   - the parsed object, true, nil on success
//   - ({}, true, err) on parse errors (with a friendly message)
//
// Callers can distinguish "no file" from "empty file" via the bool.
func readJSONFile(path string) (map[string]interface{}, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, false, nil
		}
		return nil, true, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return map[string]interface{}{}, true, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, true, fmt.Errorf("parse %s as JSON: %w (check it manually — it may have an extra comma or be truncated)", path, err)
	}
	if out == nil {
		out = map[string]interface{}{}
	}
	return out, true, nil
}

// writeJSONFile writes cfg to path with indented JSON, creating parent
// directories if needed. Uses the atomic temp+rename pattern shared
// with the rest of the brain package — a crash mid-write leaves the
// original file untouched.
func writeJSONFile(path string, cfg map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	// Indent with 2 spaces — matches what Claude Code itself writes
	// and reads, so a roundtrip through our tooling doesn't look
	// gratuitously different from what the user had.
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	// Trailing newline so the file is POSIX-friendly and diffs cleanly.
	data = append(data, '\n')
	return atomicWrite(path, data)
}

// resolveBinaryPath returns the absolute path mobiai-brain should
// invoke. Prefers --binary when set, otherwise falls back to the
// currently-running binary via os.Executable. We use the absolute
// path (not "mobiai") so MCP clients launching the server don't have
// to share the user's shell PATH.
func resolveBinaryPath(override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", fmt.Errorf("absolutize --binary: %w", err)
		}
		return abs, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("detect running binary: %w (pass --binary <path> to override)", err)
	}
	// os.Executable() can return a symlink on some platforms;
	// EvalSymlinks resolves it. This is what users expect: if you
	// `brew install mobiai`, the binary ends up at the resolved path
	// rather than the symlink in ~/.brew/bin.
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		// Symlink resolution failure is non-fatal — fall back to the
		// path as reported. Better to register something that works
		// most of the time than to refuse to install.
		return exe, nil
	}
	return resolved, nil
}

// resolveHomeDir returns the home directory to use for config
// lookups, honoring the test-only override.
func resolveHomeDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("detect HOME: %w", err)
	}
	return home, nil
}
