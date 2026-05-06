package host

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir_NestedFiles(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, body := range map[string]string{
		"root.txt":     "root",
		"a/alpha.txt":  "alpha",
		"a/b/beta.txt": "beta",
	} {
		full := filepath.Join(src, filepath.FromSlash(rel))
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	dst := filepath.Join(t.TempDir(), "out")
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	for rel, want := range map[string]string{
		"root.txt":     "root",
		"a/alpha.txt":  "alpha",
		"a/b/beta.txt": "beta",
	} {
		got, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("read %s: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("%s: got %q, want %q", rel, string(got), want)
		}
	}
}

func TestCopyDir_Overwrites(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(dst, "f.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "f.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("got %q, want %q", string(got), "new")
	}
}

func TestCopyDir_SrcMissing(t *testing.T) {
	if err := copyDir(filepath.Join(t.TempDir(), "nope"), t.TempDir()); err == nil {
		t.Error("expected error for missing src, got nil")
	}
}

func TestCopyDir_SrcIsFileNotDir(t *testing.T) {
	src := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(src, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyDir(src, t.TempDir()); err == nil {
		t.Error("expected error when src is a file, got nil")
	}
}
