package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveKnowledge_WritesCardInKnowledgeDir(t *testing.T) {
	p := initBrainForTest(t)
	id, err := AppendEntry(p, &SaveEntry{
		Type:  SaveTypeKnowledge,
		Title: "Server returns 409 when the order is closed",
		Body:  "POST /orders/{id}/items answers 409 once the order is closed.",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(p.MemoriesDir, "knowledge", id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- type: knowledge\n") {
		t.Errorf("knowledge card:\n%s", got)
	}
	groups, _ := ParseAllMemories(p)
	if len(groups["knowledge"]) != 1 {
		t.Errorf("knowledge section = %v", titles(groups["knowledge"]))
	}
}

func TestMigrate_IntegrationsNotesGoToKnowledge(t *testing.T) {
	p := oldBrain(t, t.TempDir(), map[string]string{
		"integrations.md": "\n## Firebase push quirk\n\n- id: firebase-push-quirk-20260401-000000\n- type: integration_note\n- status: active\n- date: 2026-04-01\n\nAPNs token arrives late.\n",
	})
	rep, err := Migrate(p)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Created != 1 {
		t.Errorf("Created = %d", rep.Created)
	}
	if _, err := os.Stat(filepath.Join(p.MemoriesDir, "knowledge", "firebase-push-quirk-20260401-000000.md")); err != nil {
		t.Errorf("integrations note not migrated to knowledge/: %v", err)
	}
	notes, _ := LoadNotes(p)
	if len(notes) != 1 || notes[0].Section != "knowledge" {
		t.Errorf("notes = %+v", notes)
	}
}

func TestContext_SectionIntegrationsAliasesKnowledge(t *testing.T) {
	p := initBrainForTest(t)
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeKnowledge, Title: "Checkout happy path", Body: "Cart → pay → receipt."}); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig(p.Root)
	for _, s := range []string{"knowledge", "integrations"} {
		md := BuildContextWith(cfg, nil, p, ContextOptions{Sections: []string{s}})
		if !strings.Contains(md, "## Knowledge") || !strings.Contains(md, "## Checkout happy path") {
			t.Errorf("--section %s:\n%s", s, md)
		}
	}
}
