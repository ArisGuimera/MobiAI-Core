// gen copies the repo-level skills/ directory into cli/internal/embedded/skills/
// so that go:embed can pick it up. Runs via `go generate ./...`.
//
// This file is excluded from normal builds with the `ignore` build tag.
//go:build ignore

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Working directory when run via `go generate` is the directory of the
	// file containing the //go:generate directive, which is
	// cli/internal/embedded/. Repo root is 3 levels up.
	src, err := filepath.Abs(filepath.Join("..", "..", "..", "skills"))
	if err != nil {
		log.Fatal(err)
	}
	dst, err := filepath.Abs(filepath.Join("skills"))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(src); err != nil {
		log.Fatalf("skills/ source not found at %s: %v", src, err)
	}

	if err := os.RemoveAll(dst); err != nil {
		log.Fatalf("clean dst: %v", err)
	}
	if err := copyTree(src, dst); err != nil {
		log.Fatalf("copy: %v", err)
	}
	fmt.Printf("embedded: %s -> %s\n", src, dst)
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyOne(path, target, info.Mode())
	})
}

func copyOne(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
