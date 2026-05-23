// Package graph implements the MobiAI semantic code index: a lightweight
// representation of mobile codebases (Kotlin, Swift) used by the CLI to power
// fast symbol/import queries without a full language server.
package graph

import "time"

// IndexVersion is the current schema version of the persisted Graph index.
// Bumped when the on-disk JSON shape changes in a non-backwards-compatible way.
const IndexVersion = 1

// Index is the root document of the Graph: a snapshot of every scanned file
// under Root at GeneratedAt.
type Index struct {
	Version     int         `json:"version"`
	GeneratedAt time.Time   `json:"generatedAt"`
	Root        string      `json:"root"`
	Files       []FileIndex `json:"files"`
}

// FileIndex captures the symbols and imports discovered in a single source
// file. Path is repo-relative (relative to Index.Root).
type FileIndex struct {
	Path     string   `json:"path"`
	Language string   `json:"language"`
	Lines    int      `json:"lines"`
	Imports  []string `json:"imports"`
	Symbols  []Symbol `json:"symbols"`
}

// Symbol is a top-level or nested declaration extracted from a source file.
// Kind values are platform-specific tags (e.g. "class", "fun", "struct",
// "protocol"). Container is the enclosing class/struct name when the symbol
// is nested; empty for top-level declarations.
type Symbol struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Line      int      `json:"line"`
	Container string   `json:"container,omitempty"`
	Modifiers []string `json:"modifiers,omitempty"`
}
