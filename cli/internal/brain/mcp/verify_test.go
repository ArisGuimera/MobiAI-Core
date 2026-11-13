package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const staleCard = "## Use Koin\n\n- id: use-koin-20260101-000000\n- type: architecture_decision\n- status: active\n- date: 2026-01-01\n- verified: 2026-01-01\n\nUse Koin.\n"

func writeStaleCard(t *testing.T, memoriesDir string) {
	t.Helper()
	path := filepath.Join(memoriesDir, "decisions", "use-koin-20260101-000000.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(staleCard), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMCP_Verify(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	writeStaleCard(t, paths.MemoriesDir)

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_verify",
		Arguments: map[string]any{"id": "use-koin-20260101-000000", "files": []string{"app/Module.kt"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("mobile_verify failed: %s", firstTextContent(res))
	}
	got, _ := readSectionCards(paths.MemoriesDir, "decisions")
	if strings.Contains(string(got), "- verified: 2026-01-01") || !strings.Contains(string(got), "- app/Module.kt") {
		t.Errorf("card after mobile_verify:\n%s", got)
	}
}

func TestMCP_ReviewListsStale(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	writeStaleCard(t, paths.MemoriesDir)

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_review",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := decodeReviewResult(t, res)
	if len(out.Stale) != 1 || out.Stale[0].ID != "use-koin-20260101-000000" ||
		len(out.Stale[0].Warnings) != 1 || out.Stale[0].Warnings[0] != "⚠ not verified in 3+ months" {
		t.Errorf("stale = %+v", out.Stale)
	}
	if !strings.Contains(firstTextContent(res), "may be out of date") {
		t.Errorf("text summary:\n%s", firstTextContent(res))
	}
}
