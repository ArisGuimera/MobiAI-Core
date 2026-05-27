package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrainInit_CreatesStructure(t *testing.T) {
	root := setupKMPProject(t)

	out := runBrain(t, []string{"init", "--root", root})
	if !strings.Contains(out, "Detected project") {
		t.Errorf("expected detection message, got: %s", out)
	}
	if !strings.Contains(out, "config.json") {
		t.Errorf("expected config.json mention, got: %s", out)
	}

	for _, rel := range []string{
		filepath.Join(".mobiai", "brain", "config.json"),
		filepath.Join(".mobiai", "brain", "memories", "decisions.md"),
		filepath.Join(".mobiai", "brain", "memories", "bugfixes.md"),
		filepath.Join(".mobiai", "brain", "memories", "testing.md"),
		filepath.Join(".mobiai", "brain", "memories", "integrations.md"),
		filepath.Join(".mobiai", "brain", "memories", "releases.md"),
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
	}

	// Re-running init must be idempotent: don't overwrite, and don't error.
	out2 := runBrain(t, []string{"init", "--root", root})
	if !strings.Contains(out2, "already exists") {
		t.Errorf("second init should mention existing files, got: %s", out2)
	}
}

func TestBrainScan_WritesScanJSON(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})

	out := runBrain(t, []string{"scan", "--root", root})
	if !strings.Contains(out, "scan saved") {
		t.Errorf("expected scan-saved message, got: %s", out)
	}

	scanPath := filepath.Join(root, ".mobiai", "brain", "scan.json")
	data, err := os.ReadFile(scanPath)
	if err != nil {
		t.Fatalf("read scan.json: %v", err)
	}
	var parsed struct {
		ProjectType string   `json:"project_type"`
		Platforms   []string `json:"platforms"`
		DI          []string `json:"di"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse scan.json: %v", err)
	}
	if parsed.ProjectType != "kmp" {
		t.Errorf("project_type = %q, want kmp", parsed.ProjectType)
	}
	if !sliceContains(parsed.Platforms, "android") || !sliceContains(parsed.Platforms, "ios") {
		t.Errorf("platforms = %v, want android+ios", parsed.Platforms)
	}
	if !sliceContains(parsed.DI, "koin") {
		t.Errorf("di = %v, want koin", parsed.DI)
	}
}

func TestBrainScan_RequiresInit(t *testing.T) {
	root := setupKMPProject(t)
	cmd := NewBrainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"scan", "--root", root})
	if err := cmd.Execute(); err == nil {
		t.Error("scan without init should fail")
	}
}

func TestBrainContext_OutputsMarkdown(t *testing.T) {
	root := setupKMPProject(t)
	_ = runBrain(t, []string{"init", "--root", root})
	_ = runBrain(t, []string{"scan", "--root", root})

	out := runBrain(t, []string{"context", "--root", root})
	for _, want := range []string{
		"# MobiAI Brain Context",
		"Type: Kotlin Multiplatform",
		"## Detected Stack",
		"- DI: koin",
		"## Project Rules",
		"## Architecture Decisions",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("context output missing %q. full output:\n%s", want, out)
		}
	}
}

// runBrain executes a brain subcommand and returns combined stdout/stderr.
// Fails the test on any execution error.
func runBrain(t *testing.T, args []string) string {
	t.Helper()
	cmd := NewBrainCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("brain %v: %v\noutput:\n%s", args, err, buf.String())
	}
	return buf.String()
}

// setupKMPProject creates a minimal Kotlin Multiplatform tree (Compose
// Multiplatform + Koin + Android + iOS) and returns its root.
func setupKMPProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkdir := func(rel string) {
		if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	write("settings.gradle.kts", `rootProject.name = "demo"`)
	write("composeApp/build.gradle.kts", `
plugins { kotlin("multiplatform") }
dependencies {
  implementation("io.insert-koin:koin-core:3.5.0")
  implementation("org.jetbrains.compose.runtime:runtime:1.6.0")
}
`)
	mkdir("composeApp/src/commonMain")
	mkdir("composeApp/src/androidMain")
	mkdir("composeApp/src/iosMain")
	write("composeApp/src/androidMain/AndroidManifest.xml", `<manifest/>`)
	return root
}

func sliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
