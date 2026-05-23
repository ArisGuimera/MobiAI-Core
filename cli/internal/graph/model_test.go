package graph

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestIndexVersion_IsOne(t *testing.T) {
	if IndexVersion != 1 {
		t.Fatalf("IndexVersion = %d, want 1", IndexVersion)
	}
}

func TestIndex_RoundTripJSON(t *testing.T) {
	generated := time.Date(2026, time.May, 23, 12, 30, 45, 0, time.UTC)
	original := Index{
		Version:     IndexVersion,
		GeneratedAt: generated,
		Root:        "/abs/path/to/project",
		Files: []FileIndex{
			{
				Path:     "app/src/main/kotlin/Main.kt",
				Language: "kotlin",
				Lines:    42,
				Imports:  []string{"kotlinx.coroutines.flow.Flow"},
				Symbols: []Symbol{
					{
						Name:      "MainViewModel",
						Kind:      "class",
						Line:      10,
						Container: "Main",
						Modifiers: []string{"public", "final"},
					},
				},
			},
		},
	}

	data, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Index
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Version != original.Version {
		t.Errorf("Version = %d, want %d", decoded.Version, original.Version)
	}
	if !decoded.GeneratedAt.Equal(original.GeneratedAt) {
		t.Errorf("GeneratedAt = %v, want %v", decoded.GeneratedAt, original.GeneratedAt)
	}
	if decoded.Root != original.Root {
		t.Errorf("Root = %q, want %q", decoded.Root, original.Root)
	}
	if len(decoded.Files) != 1 {
		t.Fatalf("Files length = %d, want 1", len(decoded.Files))
	}

	gotFile := decoded.Files[0]
	wantFile := original.Files[0]
	if gotFile.Path != wantFile.Path {
		t.Errorf("Path = %q, want %q", gotFile.Path, wantFile.Path)
	}
	if gotFile.Language != wantFile.Language {
		t.Errorf("Language = %q, want %q", gotFile.Language, wantFile.Language)
	}
	if gotFile.Lines != wantFile.Lines {
		t.Errorf("Lines = %d, want %d", gotFile.Lines, wantFile.Lines)
	}
	if len(gotFile.Imports) != 1 || gotFile.Imports[0] != "kotlinx.coroutines.flow.Flow" {
		t.Errorf("Imports = %v, want [kotlinx.coroutines.flow.Flow]", gotFile.Imports)
	}
	if len(gotFile.Symbols) != 1 {
		t.Fatalf("Symbols length = %d, want 1", len(gotFile.Symbols))
	}

	gotSym := gotFile.Symbols[0]
	wantSym := wantFile.Symbols[0]
	if gotSym.Name != wantSym.Name {
		t.Errorf("Symbol.Name = %q, want %q", gotSym.Name, wantSym.Name)
	}
	if gotSym.Kind != wantSym.Kind {
		t.Errorf("Symbol.Kind = %q, want %q", gotSym.Kind, wantSym.Kind)
	}
	if gotSym.Line != wantSym.Line {
		t.Errorf("Symbol.Line = %d, want %d", gotSym.Line, wantSym.Line)
	}
	if gotSym.Container != wantSym.Container {
		t.Errorf("Symbol.Container = %q, want %q", gotSym.Container, wantSym.Container)
	}
	if len(gotSym.Modifiers) != 2 || gotSym.Modifiers[0] != "public" || gotSym.Modifiers[1] != "final" {
		t.Errorf("Symbol.Modifiers = %v, want [public final]", gotSym.Modifiers)
	}
}

func TestSymbol_OmitEmpty(t *testing.T) {
	sym := Symbol{
		Name: "doThing",
		Kind: "fun",
		Line: 5,
	}

	data, err := json.Marshal(&sym)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(data)

	if strings.Contains(out, "container") {
		t.Errorf("expected no \"container\" key in JSON, got: %s", out)
	}
	if strings.Contains(out, "modifiers") {
		t.Errorf("expected no \"modifiers\" key in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"name":"doThing"`) {
		t.Errorf("expected name field in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"kind":"fun"`) {
		t.Errorf("expected kind field in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"line":5`) {
		t.Errorf("expected line field in JSON, got: %s", out)
	}
}
