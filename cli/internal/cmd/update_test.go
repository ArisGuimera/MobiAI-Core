package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func TestUpdate_FromLocalRoot_UpdatesMeta(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	root := filepath.Join("..", "catalog", "testdata", "sample")

	cmd := NewUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--catalog-root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !strings.Contains(out.String(), "Catálogo actualizado") {
		t.Errorf("output: %q", out.String())
	}

	paths, _ := state.NewPaths()
	meta, err := state.LoadMeta(paths)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(meta.LastSync) > time.Minute {
		t.Errorf("LastSync should be recent: %v", meta.LastSync)
	}
	if meta.Version == "" {
		t.Error("Version should be set after update")
	}
}
