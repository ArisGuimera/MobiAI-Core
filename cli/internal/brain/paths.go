// Package brain implements MobiAI Brain — per-project memory for mobile
// agents. Each project gets its own .mobiai/brain/ directory with config,
// stack scan and Markdown memories. This package is pure (no cobra deps)
// so it can be reused from MCP tools or other commands later.
package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BrainPaths describes the on-disk layout of a project-local brain.
// The global tool state lives in "~/.mobiai/" (managed by the state
// package); a brain lives inside the *project*, never in $HOME.
type BrainPaths struct {
	Root        string // absolute project root
	Dir         string // <root>/.mobiai/brain
	ConfigFile  string // <root>/.mobiai/brain/config.json
	ScanFile    string // <root>/.mobiai/brain/scan.json
	MemoriesDir string // <root>/.mobiai/brain/memories
}

// NewBrainPaths returns the canonical brain layout for projectRoot.
// It does not touch the filesystem.
func NewBrainPaths(projectRoot string) BrainPaths {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}
	dir := filepath.Join(abs, ".mobiai", "brain")
	return BrainPaths{
		Root:        abs,
		Dir:         dir,
		ConfigFile:  filepath.Join(dir, "config.json"),
		ScanFile:    filepath.Join(dir, "scan.json"),
		MemoriesDir: filepath.Join(dir, "memories"),
	}
}

// EnsureDirs creates Dir and MemoriesDir if they do not exist.
func (p BrainPaths) EnsureDirs() error {
	if err := os.MkdirAll(p.MemoriesDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", p.MemoriesDir, err)
	}
	return nil
}

// Exists reports whether a brain has already been initialized at this root
// (i.e. config.json exists).
func (p BrainPaths) Exists() bool {
	_, err := os.Stat(p.ConfigFile)
	return err == nil
}

// RootSource describes how the project root was located. Useful for
// debugging surprising cwd → root resolutions.
type RootSource string

const (
	RootSourceBrain   RootSource = "brain"   // existing .mobiai/brain/config.json found
	RootSourceGit     RootSource = "git"     // .git directory
	RootSourceGradle  RootSource = "gradle"  // settings.gradle(.kts)
	RootSourcePubspec RootSource = "pubspec" // pubspec.yaml (Flutter)
	RootSourceSwift   RootSource = "swift"   // Package.swift
	RootSourceXcode   RootSource = "xcode"   // *.xcworkspace or *.xcodeproj
	RootSourcePodfile RootSource = "podfile" // Podfile
	RootSourceRN      RootSource = "rn"      // package.json with react-native dep
	RootSourceCwd     RootSource = "cwd"     // fallback: nothing matched
)

// RootInfo is the result of FindProjectRoot.
type RootInfo struct {
	Path    string
	Source  RootSource
	Warning string // populated when falling back to cwd
}

// FindProjectRoot walks up from start looking for a project marker.
// Order of priority (most specific first):
//
//  1. .mobiai/brain/config.json — an already-initialized brain wins.
//  2. .git directory.
//  3. settings.gradle(.kts) — Android / KMP.
//  4. pubspec.yaml — Flutter.
//  5. Package.swift — iOS SPM.
//  6. *.xcworkspace or *.xcodeproj — iOS Xcode.
//  7. Podfile — iOS CocoaPods.
//  8. package.json with "react-native" dep — RN.
//
// If nothing matches, returns start (absolutized) with a warning.
func FindProjectRoot(start string) (RootInfo, error) {
	if start == "" {
		return RootInfo{}, fmt.Errorf("empty start path")
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return RootInfo{}, fmt.Errorf("absolutize %s: %w", start, err)
	}
	dir := abs
	for {
		if src, ok := detectMarker(dir); ok {
			return RootInfo{Path: dir, Source: src}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return RootInfo{
		Path:    abs,
		Source:  RootSourceCwd,
		Warning: "no project marker found, using the current directory",
	}, nil
}

// detectMarker checks dir for any known marker and returns the matching
// source. Order matters — most specific first.
func detectMarker(dir string) (RootSource, bool) {
	if isFile(filepath.Join(dir, ".mobiai", "brain", "config.json")) {
		return RootSourceBrain, true
	}
	if isDir(filepath.Join(dir, ".git")) || isFile(filepath.Join(dir, ".git")) {
		// .git can be a file inside worktrees; either is fine.
		return RootSourceGit, true
	}
	if isFile(filepath.Join(dir, "settings.gradle.kts")) || isFile(filepath.Join(dir, "settings.gradle")) {
		return RootSourceGradle, true
	}
	if isFile(filepath.Join(dir, "pubspec.yaml")) {
		return RootSourcePubspec, true
	}
	if isFile(filepath.Join(dir, "Package.swift")) {
		return RootSourceSwift, true
	}
	if hasGlob(dir, ".xcworkspace") || hasGlob(dir, ".xcodeproj") {
		return RootSourceXcode, true
	}
	if isFile(filepath.Join(dir, "Podfile")) {
		return RootSourcePodfile, true
	}
	if isFile(filepath.Join(dir, "package.json")) {
		if pkg, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
			if strings.Contains(string(pkg), "\"react-native\"") {
				return RootSourceRN, true
			}
		}
	}
	return "", false
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// hasGlob returns true if dir contains an entry whose name ends with suffix.
func hasGlob(dir, suffix string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			return true
		}
	}
	return false
}
