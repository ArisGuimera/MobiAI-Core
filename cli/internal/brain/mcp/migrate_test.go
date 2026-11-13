package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCP_ContextTriggersMigration(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	old := "# Decisions\n\n## Use Koin\n\n- id: use-koin-20260510-123045\n- type: architecture_decision\n- status: active\n- date: 2026-05-10\n\nUse Koin.\n"
	if err := os.WriteFile(filepath.Join(paths.MemoriesDir, "decisions.md"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_context",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := firstTextContent(res)
	if !strings.HasPrefix(text, "Brain migrated to the new format: 1 notes. Include .mobiai/brain/ in your next commit.") {
		t.Errorf("context should start with the migration message:\n%s", text)
	}
	if _, err := os.Stat(filepath.Join(paths.MemoriesDir, "decisions", "use-koin-20260510-123045.md")); err != nil {
		t.Errorf("card not created: %v", err)
	}
}
