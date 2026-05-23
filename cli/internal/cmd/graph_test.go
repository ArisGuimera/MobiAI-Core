package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runGraph executes a graph subcommand and returns combined stdout/stderr.
// Fails the test on any execution error.
func runGraph(t *testing.T, args []string) string {
	t.Helper()
	cmd := NewGraphCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("graph %v: %v\noutput:\n%s", args, err, buf.String())
	}
	return buf.String()
}

// writeFile writes content to <root>/<rel>, creating intermediate dirs.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGraphInit_CreatesIndex(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("app", "Foo.kt"), "class Foo {}\n")

	cmd := NewGraphCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"init", "--root", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Índice generado") {
		t.Errorf("expected 'Índice generado' in output, got: %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".mobiai", "graph", "index.json")); err != nil {
		t.Errorf("index.json not created: %v", err)
	}
}

func TestGraphStatus_NoIndex(t *testing.T) {
	dir := t.TempDir()

	cmd := NewGraphCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"status", "--root", dir})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when index missing, got nil. output:\n%s", out.String())
	}
	combined := out.String() + err.Error()
	if !strings.Contains(combined, "mobiai graph init") {
		t.Errorf("expected suggestion to run 'mobiai graph init', got: %s", combined)
	}
}

func TestGraphStatus_WithIndex(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("app", "Foo.kt"), "class Foo {}\n")

	_ = runGraph(t, []string{"init", "--root", dir})
	out := runGraph(t, []string{"status", "--root", dir})

	if !strings.Contains(out, "Archivos:") {
		t.Errorf("expected 'Archivos:' in status output, got: %s", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("expected file count '1' in status output, got: %s", out)
	}
}

func TestGraphSearch_FindsSymbol(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("app", "Foo.kt"), "class Foo {}\n")

	_ = runGraph(t, []string{"init", "--root", dir})
	out := runGraph(t, []string{"search", "Foo", "--root", dir})

	if !strings.Contains(out, "Foo.kt") {
		t.Errorf("expected 'Foo.kt' in search output, got: %s", out)
	}
	if !strings.Contains(out, "Foo") {
		t.Errorf("expected symbol 'Foo' in search output, got: %s", out)
	}
}

func TestGraphSearch_NoHits(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("app", "Foo.kt"), "class Foo {}\n")

	_ = runGraph(t, []string{"init", "--root", dir})
	out := runGraph(t, []string{"search", "Nonexistent", "--root", dir})

	if !strings.Contains(out, "Sin coincidencias.") {
		t.Errorf("expected 'Sin coincidencias.' message, got: %s", out)
	}
}

func TestGraphCallers_FindsRefs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("app", "a.kt"), "class Foo {}\n")
	writeFile(t, dir, filepath.Join("app", "b.kt"), "fun bar() { Foo() }\n")

	_ = runGraph(t, []string{"init", "--root", dir})
	out := runGraph(t, []string{"callers", "Foo", "--root", dir})

	if !strings.Contains(out, "b.kt") {
		t.Errorf("expected reference in 'b.kt', got: %s", out)
	}
}

func TestGraphContext_FindsFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("app", "login.kt"), "class LoginViewModel {}\n")
	writeFile(t, dir, filepath.Join("app", "unrelated.kt"), "class WeatherWidget {}\n")

	_ = runGraph(t, []string{"init", "--root", dir})
	out := runGraph(t, []string{"context", "fix", "login", "bug", "--root", dir})

	if !strings.Contains(out, "login.kt") {
		t.Errorf("expected 'login.kt' in context output, got: %s", out)
	}
	if strings.Contains(out, "unrelated.kt") {
		t.Errorf("did not expect 'unrelated.kt' in context output, got: %s", out)
	}
}
