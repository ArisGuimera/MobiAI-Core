package graph

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// defaultExcludedDirs is the set of directory base names skipped during Build.
// Walks do not descend into these directories. Defined as a var (not const)
// so a future task can override or extend it.
var defaultExcludedDirs = map[string]struct{}{
	".git":           {},
	".gradle":        {},
	".idea":          {},
	".vscode":        {},
	".mobiai":        {},
	".claude":        {}, // Claude Code state: settings, agents, worktrees
	".claude-plugin": {}, // Claude plugin metadata (this very repo uses it)
	".worktrees":     {}, // Project-local git worktrees
	"build":          {},
	"dist":           {},
	"out":            {},
	"target":         {},
	"Pods":           {},
	"DerivedData":    {},
	"node_modules":   {},
}

// Build walks root recursively, scans every .kt and .swift file with the
// appropriate scanner, and returns a populated Index. The Root field is set
// to the absolute path of root. Paths in FileIndex.Path are repo-relative
// (relative to root, using forward slashes).
//
// Directories whose base name is in the default exclusion list are skipped
// entirely (their subtrees are not walked):
//
//	.git .gradle .idea .vscode .mobiai
//	.claude .claude-plugin .worktrees
//	build dist out target
//	Pods DerivedData
//	node_modules
//
// Symlinks to directories are not followed.
func Build(root string) (*Index, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory %s does not exist", root)
		}
		return nil, fmt.Errorf("stat %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path of %s: %w", root, err)
	}

	idx := &Index{
		Version:     IndexVersion,
		GeneratedAt: time.Now().UTC(),
		Root:        absRoot,
		Files:       []FileIndex{},
	}

	walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Never skip the root itself, even if its base name happens to
			// match an excluded entry.
			if path == absRoot {
				return nil
			}
			if _, skip := defaultExcludedDirs[d.Name()]; skip {
				return fs.SkipDir
			}
			return nil
		}

		// Only regular files are considered. filepath.WalkDir does not follow
		// directory symlinks by default (DirEntry.Type reports the link type).
		if !d.Type().IsRegular() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".kt" && ext != ".swift" {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return fmt.Errorf("relative path of %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", relPath, err)
		}

		var fi FileIndex
		switch ext {
		case ".kt":
			fi = ScanKotlin(relPath, content)
		case ".swift":
			fi = ScanSwift(relPath, content)
		}
		fi.Path = relPath
		idx.Files = append(idx.Files, fi)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(idx.Files, func(i, j int) bool {
		return idx.Files[i].Path < idx.Files[j].Path
	})

	return idx, nil
}
