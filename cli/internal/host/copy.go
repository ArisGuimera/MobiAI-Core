package host

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// copyDir recursively copies the contents of src into dst.
// dst is created if missing. Existing files inside dst are overwritten.
// File permissions are preserved.
func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("src is not a directory: %s", src)
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return fmt.Errorf("mkdir dst: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", src, err)
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

// replaceDir makes dst an exact copy of src: files dst has but src no longer
// does are removed, so a file dropped from a new version of a skill does not
// linger from the old copy. src is checked first so a missing source never
// deletes what is already installed.
func replaceDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("src is not a directory: %s", src)
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("remove old %s: %w", dst, err)
	}
	return copyDir(src, dst)
}

func copyFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src %s: %w", src, err)
	}
	defer sf.Close()

	info, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("stat src %s: %w", src, err)
	}

	df, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("create dst %s: %w", dst, err)
	}
	defer df.Close()

	if _, err := io.Copy(df, sf); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	return nil
}
