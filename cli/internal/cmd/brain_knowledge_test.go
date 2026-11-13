package cmd

import (
	"strings"
	"testing"
)

func TestBrainSaveKnowledge_WritesCard(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	out := runBrain(t, []string{"save", "knowledge", "--root", root,
		"--title", "Server returns 409 when the order is closed",
		"--body", "POST /orders/{id}/items answers 409 once the order is closed."})
	if !strings.Contains(out, "memories") || !strings.Contains(out, "knowledge") {
		t.Errorf("save output should point at the knowledge card:\n%s", out)
	}
	got, err := readSectionCards(root, "knowledge")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- type: knowledge") {
		t.Errorf("card:\n%s", got)
	}
}

func TestBrainSaveKnowledge_HelpStatesRule(t *testing.T) {
	out := runBrain(t, []string{"save", "knowledge", "--help"})
	if !strings.Contains(out, "only what you can't learn by reading the code") {
		t.Errorf("help must state the knowledge rule:\n%s", out)
	}
}
