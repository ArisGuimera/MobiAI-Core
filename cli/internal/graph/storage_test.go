package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sampleIndex() *Index {
	return &Index{
		Version:     IndexVersion,
		GeneratedAt: time.Date(2026, time.May, 23, 12, 30, 45, 0, time.UTC),
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
}

func TestStorage_WriteThenRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	original := sampleIndex()
	if err := Write(path, original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	decoded, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
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
		t.Errorf("File.Path = %q, want %q", gotFile.Path, wantFile.Path)
	}
	if gotFile.Language != wantFile.Language {
		t.Errorf("File.Language = %q, want %q", gotFile.Language, wantFile.Language)
	}
	if gotFile.Lines != wantFile.Lines {
		t.Errorf("File.Lines = %d, want %d", gotFile.Lines, wantFile.Lines)
	}
	if len(gotFile.Imports) != 1 || gotFile.Imports[0] != "kotlinx.coroutines.flow.Flow" {
		t.Errorf("File.Imports = %v", gotFile.Imports)
	}
	if len(gotFile.Symbols) != 1 {
		t.Fatalf("File.Symbols length = %d, want 1", len(gotFile.Symbols))
	}

	gotSym := gotFile.Symbols[0]
	wantSym := wantFile.Symbols[0]
	if gotSym.Name != wantSym.Name || gotSym.Kind != wantSym.Kind || gotSym.Line != wantSym.Line {
		t.Errorf("Symbol = %+v, want %+v", gotSym, wantSym)
	}
	if gotSym.Container != wantSym.Container {
		t.Errorf("Symbol.Container = %q, want %q", gotSym.Container, wantSym.Container)
	}
	if len(gotSym.Modifiers) != 2 || gotSym.Modifiers[0] != "public" || gotSym.Modifiers[1] != "final" {
		t.Errorf("Symbol.Modifiers = %v", gotSym.Modifiers)
	}
}

func TestStorage_WriteCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "index.json")

	if err := Write(path, sampleIndex()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s, got: %v", path, err)
	}
}

func TestStorage_WriteIsAtomic_NoLeftoverTmp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	if err := Write(path, sampleIndex()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("expected no leftover tmp file at %s, stat err = %v", tmpPath, err)
	}

	// Also verify no stray tmp-* files (in case implementation uses CreateTemp).
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if name == "index.json" {
			continue
		}
		if strings.Contains(name, ".tmp") {
			t.Errorf("leftover tmp artifact found: %s", name)
		}
	}
}

func TestStorage_ReadVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	bogus := `{
  "version": 999,
  "generatedAt": "2026-05-23T12:30:45Z",
  "root": "/x",
  "files": []
}`
	if err := os.WriteFile(path, []byte(bogus), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	_, err := Read(path)
	if err == nil {
		t.Fatal("expected error for version mismatch, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "999") {
		t.Errorf("error message %q should mention 999", msg)
	}
	if !strings.Contains(msg, fmt.Sprintf("%d", IndexVersion)) {
		t.Errorf("error message %q should mention IndexVersion (%d)", msg, IndexVersion)
	}
}

func TestStorage_ReadMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	_, err := Read(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
