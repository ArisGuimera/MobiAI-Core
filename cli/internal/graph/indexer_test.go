package graph

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// writeFile creates parent directories and writes content to path.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestBuild_FindsKotlinAndSwift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app", "src", "main", "kotlin", "Foo.kt"), "class Foo {}\n")
	writeFile(t, filepath.Join(dir, "ios", "Sources", "Bar.swift"), "struct Bar {}\n")

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := len(idx.Files); got != 2 {
		t.Fatalf("want 2 files, got %d", got)
	}

	// Sorted by Path.
	paths := []string{idx.Files[0].Path, idx.Files[1].Path}
	if !sort.StringsAreSorted(paths) {
		t.Errorf("files not sorted: %v", paths)
	}

	// Each FileIndex has the correct Language.
	byPath := map[string]string{}
	for _, f := range idx.Files {
		byPath[f.Path] = f.Language
	}
	if byPath["app/src/main/kotlin/Foo.kt"] != "kotlin" {
		t.Errorf("kotlin file language: %q", byPath["app/src/main/kotlin/Foo.kt"])
	}
	if byPath["ios/Sources/Bar.swift"] != "swift" {
		t.Errorf("swift file language: %q", byPath["ios/Sources/Bar.swift"])
	}
}

func TestBuild_SkipsExcludedDirs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app", "src", "main", "kotlin", "Real.kt"), "class Real {}\n")
	writeFile(t, filepath.Join(dir, "build", "generated", "Fake.kt"), "class Fake {}\n")
	writeFile(t, filepath.Join(dir, "node_modules", "lib", "Ignore.kt"), "class Ignore {}\n")
	writeFile(t, filepath.Join(dir, "Pods", "SomePod", "Ignore.swift"), "struct Ignore {}\n")
	writeFile(t, filepath.Join(dir, ".git", "hooks", "Ignore.kt"), "class Ignore {}\n")
	writeFile(t, filepath.Join(dir, ".mobiai", "cached", "Ignore.kt"), "class Ignore {}\n")

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.Files) != 1 {
		t.Fatalf("want 1 file, got %d: %+v", len(idx.Files), idx.Files)
	}
	if idx.Files[0].Path != "app/src/main/kotlin/Real.kt" {
		t.Errorf("unexpected file: %q", idx.Files[0].Path)
	}
}

func TestBuild_SkipsClaudeAndWorktreeDirs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app", "Real.kt"), "class Real {}\n")
	// Real Claude Code clutter that showed up in user projects: agent
	// worktrees nested under .claude. The whole .claude tree should be
	// invisible to the indexer, otherwise we double-index everything.
	writeFile(t, filepath.Join(dir, ".claude", "worktrees", "ios-v5", "src", "FakeViewModel.kt"), "class FakeViewModel {}\n")
	writeFile(t, filepath.Join(dir, ".claude", "skills", "custom", "Notes.kt"), "class Notes {}\n")
	writeFile(t, filepath.Join(dir, ".claude-plugin", "marketplace", "Bundle.kt"), "class Bundle {}\n")
	writeFile(t, filepath.Join(dir, ".worktrees", "feature-x", "App.kt"), "class App {}\n")

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.Files) != 1 {
		t.Fatalf("want 1 file (only app/Real.kt), got %d: %+v", len(idx.Files), idx.Files)
	}
	if idx.Files[0].Path != "app/Real.kt" {
		t.Errorf("unexpected file: %q", idx.Files[0].Path)
	}
}

func TestBuild_PathsAreRepoRelativeForwardSlashes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a", "b", "c", "Foo.kt"), "class Foo {}\n")

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.Files) != 1 {
		t.Fatalf("want 1 file, got %d", len(idx.Files))
	}
	if idx.Files[0].Path != "a/b/c/Foo.kt" {
		t.Errorf("path = %q, want %q", idx.Files[0].Path, "a/b/c/Foo.kt")
	}
}

func TestBuild_RootIsAbsolute(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Foo.kt"), "class Foo {}\n")

	// Use a relative path by chdir'ing into the parent.
	parent := filepath.Dir(dir)
	base := filepath.Base(dir)
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(parent); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	idx, err := Build(base)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !filepath.IsAbs(idx.Root) {
		t.Errorf("Root is not absolute: %q", idx.Root)
	}
}

func TestBuild_VersionAndTimestamp(t *testing.T) {
	dir := t.TempDir()

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if idx.Version != IndexVersion {
		t.Errorf("version = %d, want %d", idx.Version, IndexVersion)
	}
	if idx.GeneratedAt.IsZero() {
		t.Errorf("GeneratedAt is zero")
	}
	if idx.GeneratedAt.Location() != time.UTC {
		t.Errorf("GeneratedAt location = %v, want UTC", idx.GeneratedAt.Location())
	}
}

func TestBuild_NonExistentDir(t *testing.T) {
	idx, err := Build("/this/path/does/not/exist")
	if err == nil {
		t.Fatalf("want error, got nil")
	}
	if idx != nil {
		t.Errorf("want nil index, got %+v", idx)
	}
	if !strings.Contains(err.Error(), "no existe") && !strings.Contains(err.Error(), "/this/path/does/not/exist") {
		t.Errorf("error message %q lacks expected hints", err.Error())
	}
}

func TestBuild_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "regular.txt")
	if err := os.WriteFile(filePath, []byte("hi"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	idx, err := Build(filePath)
	if err == nil {
		t.Fatalf("want error, got nil")
	}
	if idx != nil {
		t.Errorf("want nil index, got %+v", idx)
	}
	if !strings.Contains(err.Error(), "no es un directorio") {
		t.Errorf("error message %q does not mention 'no es un directorio'", err.Error())
	}
}

func TestBuild_IgnoresOtherFileExtensions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app", "Foo.java"), "class Foo {}\n")
	writeFile(t, filepath.Join(dir, "app", "Bar.cpp"), "int main() {}\n")
	writeFile(t, filepath.Join(dir, "app", "Doc.md"), "# Doc\n")
	writeFile(t, filepath.Join(dir, "app", "Foo.kt"), "class Foo {}\n")

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.Files) != 1 {
		t.Fatalf("want 1 file, got %d: %+v", len(idx.Files), idx.Files)
	}
	if idx.Files[0].Path != "app/Foo.kt" {
		t.Errorf("unexpected file: %q", idx.Files[0].Path)
	}
}

func TestBuild_EmptyProject(t *testing.T) {
	dir := t.TempDir()

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.Files) != 0 {
		t.Errorf("want 0 files, got %d", len(idx.Files))
	}
}

func TestBuild_SortedByPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "z", "Z.kt"), "class Z {}\n")
	writeFile(t, filepath.Join(dir, "a", "A.kt"), "class A {}\n")
	writeFile(t, filepath.Join(dir, "m", "M.swift"), "struct M {}\n")

	idx, err := Build(dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.Files) != 3 {
		t.Fatalf("want 3 files, got %d", len(idx.Files))
	}
	want := []string{"a/A.kt", "m/M.swift", "z/Z.kt"}
	for i, w := range want {
		if idx.Files[i].Path != w {
			t.Errorf("Files[%d].Path = %q, want %q", i, idx.Files[i].Path, w)
		}
	}
}
