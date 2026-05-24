package embedded

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Extract walks the embedded Plugins FS and writes every file to dstRoot,
// preserving the directory structure. Existing files are overwritten.
// dstRoot is created if missing.
//
// After Extract, dstRoot contains a layout identical to the source skills/
// directory: <dstRoot>/<pack>/.claude-plugin/plugin.json,
// <dstRoot>/<pack>/skills/..., etc.
//
// Files that look executable (shebang or .sh/.bash/.zsh/.fish extension)
// are written with 0o755; everything else with 0o644. embed.FS does not
// preserve unix modes, so we synthesize them from content cues.
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
		// Strip the leading "skills/" segment so dstRoot acts as the
		// equivalent of the repo's skills/ dir.
		rel := path
		if rel == "skills" {
			return os.MkdirAll(dstRoot, 0o755)
		}
		if len(rel) > len("skills/") && rel[:len("skills/")] == "skills/" {
			rel = rel[len("skills/"):]
		}
		target := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := Plugins.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}
		mode := os.FileMode(0o644)
		if isExecutable(rel, data) {
			mode = 0o755
		}
		if err := os.WriteFile(target, data, mode); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
		return nil
	})
}

// isExecutable returns true when the file looks like a script we should
// preserve the +x bit for. Detects by shebang or known extensions.
func isExecutable(name string, data []byte) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".sh", ".bash", ".zsh", ".fish":
		return true
	}
	return bytes.HasPrefix(data, []byte("#!"))
}

// ErrEmpty is returned when the embedded FS has no plugins (e.g., go generate
// was not run before the build, or the plugins directory was empty).
var ErrEmpty = errors.New("embedded plugins are empty (was `go generate` run?)")

// IsEmpty reports whether the embedded FS contains any plugins. Returns
// true on read errors (e.g., the skills/ directory does not exist in the
// embed) so callers fall back to other catalog sources rather than panicking.
func IsEmpty() bool {
	entries, err := Plugins.ReadDir("skills")
	if err != nil {
		return true
	}
	return len(entries) == 0
}
