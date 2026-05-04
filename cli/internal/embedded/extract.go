package embedded

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Extract walks the embedded Plugins FS and writes every file to dstRoot,
// preserving the directory structure. Existing files are overwritten.
// dstRoot is created if missing.
//
// After Extract, dstRoot contains a layout identical to the source plugins/
// directory: <dstRoot>/<pack>/.claude-plugin/plugin.json,
// <dstRoot>/<pack>/skills/..., etc.
func Extract(dstRoot string) error {
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dstRoot, err)
	}
	return fs.WalkDir(Plugins, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		// Strip the leading "plugins/" segment so dstRoot acts as the
		// equivalent of the repo's plugins/ dir.
		rel := path
		if rel == "plugins" {
			return os.MkdirAll(dstRoot, 0o755)
		}
		if len(rel) > len("plugins/") && rel[:len("plugins/")] == "plugins/" {
			rel = rel[len("plugins/"):]
		}
		target := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := Plugins.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
		return nil
	})
}

// ErrEmpty is returned when the embedded FS has no plugins (e.g., go generate
// was not run before the build, or the plugins directory was empty).
var ErrEmpty = errors.New("embedded plugins are empty (was `go generate` run?)")

// IsEmpty reports whether the embedded FS contains any plugins.
func IsEmpty() bool {
	entries, err := Plugins.ReadDir("plugins")
	if err != nil {
		return true
	}
	return len(entries) == 0
}
