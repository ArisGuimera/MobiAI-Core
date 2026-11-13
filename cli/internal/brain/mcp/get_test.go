package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
)

func TestMCP_GetReturnsCard(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	id, err := brain.AppendEntry(paths, &brain.SaveEntry{Type: brain.SaveTypeDecision, Title: "Use Koin", Body: "Use Koin."})
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_get",
		Arguments: map[string]any{"id": id},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("mobile_get failed: %s", firstTextContent(res))
	}
	text := firstTextContent(res)
	if !strings.Contains(text, "## Use Koin") || !strings.Contains(text, "- id: "+id) {
		t.Errorf("mobile_get text:\n%s", text)
	}
	raw, _ := json.Marshal(res.StructuredContent)
	var view brain.NoteView
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if view.ID != id || view.Section != "decisions" || !strings.HasSuffix(view.Path, "/decisions/"+id+".md") {
		t.Errorf("structured result = %+v", view)
	}
}

func TestMCP_GetUnknownIDFails(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()
	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_get",
		Arguments: map[string]any{"id": "nope"},
	})
	if err == nil && (res == nil || !res.IsError) {
		t.Errorf("unknown id should fail")
	}
}
