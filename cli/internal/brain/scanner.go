package brain

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// maxScanDepth is the deepest directory level (relative to the project
// root) the scanner descends into. 6 is plenty for typical mobile
// monorepos and keeps the walk fast on big trees.
const maxScanDepth = 6

// maxFileSize is the largest file the scanner is willing to read into
// memory for content matching. Anything bigger is skipped (binary blobs,
// generated artefacts).
const maxFileSize = 1 << 20 // 1 MiB

// skipDirs are directory names ignored during the walk regardless of
// depth. Mostly noise: VCS, build outputs, IDE state, vendor caches,
// and AI-agent workspaces (which often contain full repo copies and
// would duplicate every signal).
var skipDirs = map[string]struct{}{
	".git":         {},
	".mobiai":      {},
	".gradle":      {},
	".idea":        {},
	".kotlin":      {},
	".dart_tool":   {},
	".swiftpm":     {},
	"build":        {},
	"out":          {},
	"target":       {},
	"node_modules": {},
	"Pods":         {},
	"DerivedData":  {},
	"vendor":       {},
	".claude":      {},
	".cursor":      {},
	".copilot":     {},
	".gemini":      {},
	".codex":       {},
	".junie":       {},
}

// interestingFiles lists the basenames whose full content we slurp into
// the index for downstream regex/string matching. The rest of the tree
// is recorded by path only (presence checks, suffix checks).
//
// Sensitive files (Firebase configs, .env, signing material) are NOT in
// this list — they're flagged by `detectSensitiveFiles` via path-only
// checks so we never load their bytes. See `sensitiveBasenames` below
// for the explicit denylist enforced in buildIndex.
var interestingFiles = map[string]struct{}{
	"build.gradle":        {},
	"build.gradle.kts":    {},
	"settings.gradle":     {},
	"settings.gradle.kts": {},
	"pubspec.yaml":        {},
	"pubspec.lock":        {},
	"package.json":        {},
	"metro.config.js":     {},
	"Podfile":             {},
	"Package.swift":       {},
	"AndroidManifest.xml": {},
	"libs.versions.toml":  {},
	"gradle.properties":   {},
}

// sensitiveBasenames are file basenames the scanner refuses to read
// regardless of context. Defense-in-depth: even if a future change adds
// one of these to interestingFiles by mistake, buildIndex still rejects
// it. Their *presence* is still reported via detectSensitiveFiles, just
// without ever loading bytes into memory.
var sensitiveBasenames = map[string]struct{}{
	".env":                     {},
	"google-services.json":     {},
	"GoogleService-Info.plist": {},
	"local.properties":         {},
	"keystore.properties":      {},
}

// scanIndex captures everything the detectors need: opened files
// (full content) and a flat list of all paths walked.
type scanIndex struct {
	root     string
	files    map[string][]byte // relpath → content (only for interestingFiles)
	allPaths map[string]bool   // relpath → isDir (every entry visited)
	warnings []string          // unexpected read errors during the walk
}

func newScanIndex(root string) *scanIndex {
	return &scanIndex{
		root:     root,
		files:    map[string][]byte{},
		allPaths: map[string]bool{},
	}
}

// hasFile reports whether a file with relpath was seen (relative to root).
func (i *scanIndex) hasFile(relpath string) bool {
	isDir, ok := i.allPaths[relpath]
	return ok && !isDir
}

// hasDir reports whether a directory with relpath was seen.
func (i *scanIndex) hasDir(relpath string) bool {
	isDir, ok := i.allPaths[relpath]
	return ok && isDir
}

// hasAnyFileSuffix reports whether any indexed path ends with suffix.
func (i *scanIndex) hasAnyFileSuffix(suffix string) bool {
	for p, isDir := range i.allPaths {
		if !isDir && strings.HasSuffix(p, suffix) {
			return true
		}
	}
	return false
}

// hasAnyDirSuffix reports whether any indexed directory path ends with
// suffix (e.g. "/commonMain", "/composeApp").
func (i *scanIndex) hasAnyDirSuffix(suffix string) bool {
	for p, isDir := range i.allPaths {
		if !isDir {
			continue
		}
		if p == strings.TrimPrefix(suffix, "/") || strings.HasSuffix(p, suffix) {
			return true
		}
	}
	return false
}

// content returns the bytes of an interesting file, or nil if not loaded.
func (i *scanIndex) content(relpath string) []byte {
	return i.files[relpath]
}

// containsAny scans all loaded file contents for needle (case-insensitive)
// and returns true at the first hit.
func (i *scanIndex) containsAny(needle string) bool {
	low := strings.ToLower(needle)
	for _, content := range i.files {
		if strings.Contains(strings.ToLower(string(content)), low) {
			return true
		}
	}
	return false
}

// ScanProject walks the project at root, builds a scan index, runs the
// platform detectors and returns a populated Scan. The walk never aborts
// on a per-entry error: if reading an interesting file fails, the path
// is recorded in scan.Warnings so the user sees what was missed. Depth
// and size limits are enforced silently (they're a scanner choice, not
// a problem with the file itself).
func ScanProject(root string) (*Scan, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("absolutize root: %w", err)
	}
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("root is not a directory: %s", abs)
	}

	idx := newScanIndex(abs)
	if err := buildIndex(idx); err != nil {
		return nil, err
	}

	s := newScan(filepath.Base(abs))
	s.Warnings = append(s.Warnings, idx.warnings...)
	platforms := map[string]struct{}{}
	types := map[string]struct{}{}

	detectAndroid(idx, s, types, platforms)
	detectKMP(idx, s, types, platforms)
	detectFlutter(idx, s, types, platforms)
	detectIOS(idx, s, types, platforms)
	detectReactNative(idx, s, types, platforms)
	detectDeps(idx, s)

	s.ProjectType = pickProjectType(types)
	s.Platforms = sortedKeys(platforms)
	dedupAllBuckets(s)
	return s, nil
}

// buildIndex walks the tree, applies skip rules and depth cap, and reads
// the content of interesting files into the index.
func buildIndex(idx *scanIndex) error {
	return filepath.WalkDir(idx.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Don't abort the entire scan because one entry is unreadable.
			return nil
		}
		if path == idx.root {
			return nil
		}
		rel, rerr := filepath.Rel(idx.root, path)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		depth := strings.Count(rel, "/") + 1
		if depth > maxScanDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if _, skip := skipDirs[d.Name()]; skip {
				return fs.SkipDir
			}
			idx.allPaths[rel] = true
			return nil
		}
		idx.allPaths[rel] = false
		base := filepath.Base(path)
		if _, ok := sensitiveBasenames[base]; ok {
			// Hard guard: never load the bytes of files known to carry
			// secrets, even if some future change adds them to
			// interestingFiles. Their presence is still reported
			// downstream by detectSensitiveFiles via the path index.
			return nil
		}
		if _, ok := interestingFiles[base]; !ok {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > maxFileSize {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			idx.warnings = append(idx.warnings,
				fmt.Sprintf("could not read %s: %v", rel, err))
			return nil
		}
		idx.files[rel] = data
		return nil
	})
}

// pickProjectType applies precedence: KMP > Flutter > RN > Android > iOS.
// KMP wins over Android because a KMP project usually contains an Android
// module. Flutter and RN win over their host platforms for the same reason.
func pickProjectType(types map[string]struct{}) string {
	for _, t := range []string{
		ProjectTypeKMP,
		ProjectTypeFlutter,
		ProjectTypeReactNative,
		ProjectTypeAndroid,
		ProjectTypeIOS,
	} {
		if _, ok := types[t]; ok {
			return t
		}
	}
	return ProjectTypeUnknown
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// dedupAllBuckets normalizes every []string field on Scan: sorts and
// deduplicates. Keeps scan.json diffs minimal between runs.
func dedupAllBuckets(s *Scan) {
	s.Platforms = dedupSorted(s.Platforms)
	s.BuildSystems = dedupSorted(s.BuildSystems)
	s.UI = dedupSorted(s.UI)
	s.DI = dedupSorted(s.DI)
	s.Network = dedupSorted(s.Network)
	s.Persistence = dedupSorted(s.Persistence)
	s.Serialization = dedupSorted(s.Serialization)
	s.Testing = dedupSorted(s.Testing)
	s.Integrations = dedupSorted(s.Integrations)
	s.CICD = dedupSorted(s.CICD)
	s.Warnings = dedupSorted(s.Warnings)
}

func dedupSorted(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
