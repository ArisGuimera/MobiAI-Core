package embedded

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPlugins_HasMarketplaceJSON is a smoke test confirming `go generate`
// was run before the build and the embedded FS contains the marketplace.
// It is the only test that exercises the embed; the rest of the package is
// verified by the integration into skills.go's fallback chain.
func TestPlugins_HasMarketplaceJSON(t *testing.T) {
	if IsEmpty() {
		t.Fatal("embedded Plugins is empty — run `go generate ./...` before testing")
	}
	// The repo's marketplace.json lives at .claude-plugin/marketplace.json,
	// outside the skills/ tree, so the embedded FS will NOT contain it.
	// Instead, verify that at least one expected pack manifest is present.
	if _, err := Plugins.ReadFile("skills/core/.claude-plugin/plugin.json"); err != nil {
		t.Errorf("expected core/.claude-plugin/plugin.json in embed: %v", err)
	}
}

func TestExtract_WritesFilesToDestination(t *testing.T) {
	if IsEmpty() {
		t.Skip("skipping; run `go generate ./...` first")
	}

	dst := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(dst); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "core", ".claude-plugin", "plugin.json")); err != nil {
		t.Errorf("expected core/.claude-plugin/plugin.json after Extract: %v", err)
	}
}
