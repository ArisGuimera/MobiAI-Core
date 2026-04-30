package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRoot points at the catalog package's testdata fixture.
// Path is relative to this test file: ../catalog/testdata/sample.
func fixtureRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "catalog", "testdata", "sample"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	return abs
}

func TestCatalogList_PrintsPacks(t *testing.T) {
	cmd := NewCatalogCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--root", fixtureRoot(t)})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, pack := range []string{"core", "android", "ios", "kmp"} {
		if !strings.Contains(got, pack) {
			t.Errorf("output should contain pack %q; got: %q", pack, got)
		}
	}
}

func TestCatalogResolve_KMP(t *testing.T) {
	cmd := NewCatalogCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"resolve", "kmp", "--root", fixtureRoot(t)})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, pack := range []string{"core", "android", "ios", "kmp"} {
		if !strings.Contains(got, pack) {
			t.Errorf("resolve output should contain %q; got: %q", pack, got)
		}
	}
	// core must appear before kmp in the output.
	corePos := strings.Index(got, "core")
	kmpPos := strings.Index(got, "kmp")
	if corePos < 0 || kmpPos < 0 || corePos > kmpPos {
		t.Errorf("expected core to appear before kmp; got: %q", got)
	}
}

func TestCatalogResolve_UnknownPack(t *testing.T) {
	cmd := NewCatalogCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"resolve", "androyd", "--root", fixtureRoot(t)})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unknown pack, got nil")
	}
}
