package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildContext_FullDocument(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureMemoryFiles(p); err != nil {
		t.Fatal(err)
	}
	// Write a real entry into decisions.md to verify the empty-template
	// detector doesn't swallow real content.
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		"# Decisions\n\n## Use Koin as DI\n\n- status: active\n")

	cfg := Config{
		Version:     ConfigVersion,
		ProjectName: "demo-kmp",
		ProjectType: ProjectTypeKMP,
		Platforms:   []string{"android", "ios", "shared"},
		Rules:       []string{"Prioritize clean architecture"},
	}
	scan := &Scan{
		Version:      ScanVersion,
		ProjectType:  ProjectTypeKMP,
		Platforms:    []string{"android", "ios"},
		BuildSystems: []string{"gradle"},
		UI:           []string{"compose_multiplatform"},
		DI:           []string{"koin"},
		Integrations: []string{"firebase"},
		Warnings:     []string{"archivo sensible detectado: .env (no leído)"},
	}

	md := BuildContext(cfg, scan, p)

	for _, want := range []string{
		"# MobiAI Brain Context",
		"Project: demo-kmp",
		"Type: Kotlin Multiplatform",
		"Platforms: android, ios, shared",
		"## Detected Stack",
		"- UI: compose_multiplatform",
		"- DI: koin",
		"- Integrations: firebase",
		"## Project Rules",
		"- Prioritize clean architecture",
		"## Architecture Decisions",
		"## Use Koin as DI",
		"## Known Bugfixes",
		"_No entries yet._",
		"## Testing Patterns",
		"## Integrations",
		"## Release Notes",
		"## Warnings",
		"archivo sensible detectado: .env",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("BuildContext output missing %q\n--- output ---\n%s", want, md)
		}
	}
}

func TestBuildContext_NoScanYet(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureMemoryFiles(p); err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		Version:     ConfigVersion,
		ProjectName: "fresh",
		ProjectType: ProjectTypeUnknown,
	}
	md := BuildContext(cfg, nil, p)

	for _, want := range []string{
		"Project: fresh",
		"Type: Unknown",
		"## Detected Stack",
		"_No scan yet. Run `mobiai brain scan`._",
		"## Warnings",
		"_None._",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q in output:\n%s", want, md)
		}
	}
}

func TestEnsureMemoryFiles_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	created, err := EnsureMemoryFiles(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != len(MemoryFiles) {
		t.Errorf("first run created %d files, want %d", len(created), len(MemoryFiles))
	}

	// Mutate one file: re-running must not overwrite it.
	target := filepath.Join(p.MemoriesDir, "decisions.md")
	if err := os.WriteFile(target, []byte("# user content"), 0o644); err != nil {
		t.Fatal(err)
	}
	created2, err := EnsureMemoryFiles(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(created2) != 0 {
		t.Errorf("second run created %d files, want 0", len(created2))
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# user content" {
		t.Errorf("decisions.md was overwritten: %q", string(got))
	}
}
