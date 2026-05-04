package cmd

import (
	"bytes"
	"os"
	"os/exec"
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

// makeBareCatalogRepo creates a bare git repo seeded from the project's
// catalog testdata fixture. Returns the file:// URL the bare repo can be
// cloned from. Skips the test if `git` is not installed.
func makeBareCatalogRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH; skipping remote-catalog test")
	}

	src, err := filepath.Abs(filepath.Join("..", "catalog", "testdata", "sample"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a working repo from the fixture.
	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyDirContents(src, work); err != nil {
		t.Fatalf("seed work: %v", err)
	}
	mustRun(t, work, "git", "init", "--initial-branch=main")
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "add", ".")
	mustRun(t, work, "git", "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "seed")

	// Create a bare repo and push to it.
	// Use --initial-branch=main so HEAD points to main (not master), matching
	// the working repo's branch name.
	bare := filepath.Join(t.TempDir(), "bare.git")
	mustRun(t, "", "git", "init", "--bare", "--initial-branch=main", bare)
	mustRun(t, work, "git", "remote", "add", "origin", bare)
	mustRun(t, work, "git", "push", "origin", "main")

	// Convert filesystem path to a file:// URL portable across OSes.
	abs, err := filepath.Abs(bare)
	if err != nil {
		t.Fatal(err)
	}
	return "file:///" + filepath.ToSlash(abs)
}

func copyDirContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	c := exec.Command(name, args...)
	if dir != "" {
		c.Dir = dir
	}
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestUpdate_GitClone_FreshCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	t.Setenv("MOBIAI_CATALOG_GIT_URL", makeBareCatalogRepo(t))

	cmd := NewUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	// Verify the repo was cloned to ~/.mobiai/cache/catalog/.
	cacheRoot := filepath.Join(tmp, ".mobiai", "cache", "catalog")
	if _, err := os.Stat(filepath.Join(cacheRoot, ".claude-plugin", "marketplace.json")); err != nil {
		t.Errorf("expected marketplace.json in cloned cache: %v", err)
	}
	if !strings.Contains(out.String(), "Clonando catálogo") {
		t.Errorf("expected clone message; got: %q", out.String())
	}
	if !strings.Contains(out.String(), "Catálogo actualizado") {
		t.Errorf("expected success message; got: %q", out.String())
	}

	// CatalogMeta should be fresh.
	paths, _ := state.NewPaths()
	meta, err := state.LoadMeta(paths)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(meta.LastSync) > time.Minute {
		t.Errorf("LastSync should be recent: %v", meta.LastSync)
	}
}

func TestUpdate_GitPull_ExistingCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	url := makeBareCatalogRepo(t)
	t.Setenv("MOBIAI_CATALOG_GIT_URL", url)

	// First update: clone.
	first := NewUpdateCmd()
	var out1 bytes.Buffer
	first.SetOut(&out1)
	first.SetErr(&out1)
	if err := first.Execute(); err != nil {
		t.Fatalf("first execute: %v", err)
	}

	// Second update: pull.
	second := NewUpdateCmd()
	var out2 bytes.Buffer
	second.SetOut(&out2)
	second.SetErr(&out2)
	if err := second.Execute(); err != nil {
		t.Fatalf("second execute: %v", err)
	}

	if !strings.Contains(out2.String(), "Sincronizando catálogo (git pull)") {
		t.Errorf("expected pull message on 2nd run; got: %q", out2.String())
	}
}
