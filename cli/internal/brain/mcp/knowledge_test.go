package mcp

import (
	"context"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCP_SaveKnowledge(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_save_knowledge",
		Arguments: map[string]any{
			"title": "Server returns 409 when the order is closed",
			"body":  "POST /orders/{id}/items answers 409 once the order is closed.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("save failed: %s", firstTextContent(res))
	}
	got, err := readSectionCards(paths.MemoriesDir, "knowledge")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- type: knowledge") {
		t.Errorf("card:\n%s", got)
	}

	tools, err := client.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name == "mobile_save_knowledge" && !strings.Contains(tool.Description, "can't learn by reading the code") {
			t.Errorf("tool description must state the rule: %s", tool.Description)
		}
	}
}
