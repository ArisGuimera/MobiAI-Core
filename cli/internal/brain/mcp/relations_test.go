package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func callTool(t *testing.T, client *sdkmcp.ClientSession, name string, args map[string]any) (*sdkmcp.CallToolResult, error) {
	t.Helper()
	return client.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: name, Arguments: args})
}

func decodeSaveID(t *testing.T, res *sdkmcp.CallToolResult) string {
	t.Helper()
	raw, _ := json.Marshal(res.StructuredContent)
	var out SaveResult
	if err := json.Unmarshal(raw, &out); err != nil || out.ID == "" {
		t.Fatalf("no id in save result: %s", raw)
	}
	return out.ID
}

func TestMCP_SaveWithSupersedesAndNew(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	res, err := callTool(t, client, "mobile_save_bugfix", map[string]any{"title": "Login crash on screen rotation", "body": "b"})
	if err != nil || res.IsError {
		t.Fatalf("first save: %v %s", err, firstTextContent(res))
	}
	oldID := decodeSaveID(t, res)

	res, err = callTool(t, client, "mobile_save_bugfix", map[string]any{"title": "Login crash on rotation of the screen", "body": "b"})
	text := ""
	if err != nil {
		text = err.Error()
	} else if res.IsError {
		text = firstTextContent(res)
	} else {
		t.Fatal("a similar title must be refused")
	}
	if !strings.Contains(text, "A similar note already exists: "+oldID) || !strings.Contains(text, "new") || !strings.Contains(text, "supersedes") {
		t.Errorf("MCP brake message: %s", text)
	}

	if res, err := callTool(t, client, "mobile_save_bugfix", map[string]any{"title": "Login crash on rotation of the screen", "body": "b", "new": true}); err != nil || res.IsError {
		t.Errorf("new: true must save: %v", err)
	}
	if res, err := callTool(t, client, "mobile_save_bugfix", map[string]any{"title": "Login crash fixed for good", "body": "b",
		"supersedes": []string{oldID}, "crash": []string{"6a1b2c3d4e5f"}}); err != nil || res.IsError {
		t.Fatalf("save with supersedes/crash: %v", err)
	}
	got, _ := readSectionCards(paths.MemoriesDir, "bugfixes")
	if !strings.Contains(string(got), "- supersedes: "+oldID) || !strings.Contains(string(got), "- crash: 6a1b2c3d4e5f") {
		t.Errorf("cards:\n%s", got)
	}
}

func TestMCP_SearchByCrash(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()
	if res, err := callTool(t, client, "mobile_save_bugfix", map[string]any{"title": "NPE on login", "body": "b", "crash": []string{"6a1b2c3d4e5f"}}); err != nil || res.IsError {
		t.Fatalf("save: %v", err)
	}
	res, err := callTool(t, client, "mobile_search", map[string]any{"crash": "6A1B2C3D4E5F"})
	if err != nil || res.IsError {
		t.Fatalf("search by crash: %v %s", err, firstTextContent(res))
	}
	hits := decodeSearchHits(t, res)
	if len(hits) != 1 || hits[0].Title != "NPE on login" || !strings.Contains(hits[0].Path, "/bugfixes/") {
		t.Errorf("hits = %+v", hits)
	}
}
