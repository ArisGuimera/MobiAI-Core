// Package host defines the HostAdapter interface and per-host implementations.
package host

import "github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"

// HostAdapter abstracts every operation the CLI performs against one AI client.
type HostAdapter interface {
	ID() string
	Name() string
	Homepage() string

	Detect() DetectResult
	SkillsDir() string
	Capabilities() Caps

	Install(skills []catalog.Skill) error
	Uninstall(skillIDs []string) error
	List() ([]InstalledSkill, error)
	Verify() DriftReport

	// Experimental returns true for adapters whose home subdir is
	// speculative (tier-3). The default Registry excludes them from
	// auto-detection so we don't hijack unrelated directories like
	// ~/.mux or ~/.factory; users opt in via --include-experimental
	// or by passing --host=<id> explicitly.
	Experimental() bool
}

type DetectResult struct {
	Found    bool
	Path     string
	Searched []string
}

type Caps struct {
	Skills   bool
	Hooks    bool
	Commands bool
	MCPs     bool
}

type InstalledSkill struct {
	ID   string
	Path string
}

// DriftReport is a stub in Phase 3; full drift detection ships with
// `mobiai doctor` polish in a later phase.
type DriftReport struct {
	Issues []DriftIssue
}

type DriftIssue struct {
	SkillID string
	Kind    string // "missing", "modified", "extra"
	Detail  string
}
