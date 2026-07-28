package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SaveType identifies the memory category an entry targets. The set is
// closed: each value maps to a fixed memories/<file>.md and a default
// fully-qualified type label.
type SaveType string

const (
	SaveTypeDecision SaveType = "decision"
	SaveTypeBugfix   SaveType = "bugfix"
	SaveTypeTesting  SaveType = "testing"
)

// memoryFileFor returns the basename inside memories/ for a SaveType.
func (t SaveType) memoryFileFor() (string, error) {
	switch t {
	case SaveTypeDecision:
		return "decisions.md", nil
	case SaveTypeBugfix:
		return "bugfixes.md", nil
	case SaveTypeTesting:
		return "testing.md", nil
	default:
		return "", fmt.Errorf("unknown save type: %q (expected: decision|bugfix|testing)", t)
	}
}

// fullTypeLabel returns the type: <label> field for the rendered entry.
// `bugfix` flips between `platform_workaround` and `bug_fix` based on
// status — that's how the spec disambiguates a real fix from a temporary
// hack while sharing the same .md file.
func (t SaveType) fullTypeLabel(status string) string {
	switch t {
	case SaveTypeDecision:
		return "architecture_decision"
	case SaveTypeBugfix:
		if strings.EqualFold(status, "temporary") {
			return "platform_workaround"
		}
		return "bug_fix"
	case SaveTypeTesting:
		return "testing_pattern"
	}
	return string(t)
}

// Status values accepted for a SaveEntry.
const (
	StatusActive     = "active"
	StatusTemporary  = "temporary"
	StatusDeprecated = "deprecated"
)

// validStatuses lists every allowed status value.
var validStatuses = map[string]struct{}{
	StatusActive:     {},
	StatusTemporary:  {},
	StatusDeprecated: {},
}

// SaveEntry is a structured memory append. The CLI builds it from flags
// (and stdin for Body); save_test.go drives it directly.
type SaveEntry struct {
	Type        SaveType
	Title       string
	Platform    string   // optional
	Area        string   // optional
	Status      string   // active|temporary|deprecated, defaults to active
	ReviewAfter string   // optional, ISO date — meaningful for temporary entries
	Body        string   // free-form Markdown (may include H3 subheadings)
	Files       []string // optional list of repo-relative file paths
	now         func() time.Time
}

// Validate enforces required fields and value domains. Returns a
// human-readable error in Spanish (matching CLI tone) on first failure.
func (e *SaveEntry) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("type is required")
	}
	if _, err := e.Type.memoryFileFor(); err != nil {
		return err
	}
	if strings.TrimSpace(e.Title) == "" {
		return fmt.Errorf("--title is required")
	}
	if e.Status == "" {
		e.Status = StatusActive
	}
	if _, ok := validStatuses[e.Status]; !ok {
		return fmt.Errorf("invalid --status %q (expected: active|temporary|deprecated)", e.Status)
	}
	if e.ReviewAfter != "" {
		if _, err := time.Parse("2006-01-02", e.ReviewAfter); err != nil {
			return fmt.Errorf("--review-after must be YYYY-MM-DD: %w", err)
		}
	}
	return nil
}

// AppendEntry validates the entry, ensures the brain exists at p, and
// appends a Markdown section to the relevant memories/<file>.md. Returns
// the generated id (slug + timestamp) so callers can echo it.
func AppendEntry(p BrainPaths, e *SaveEntry) (string, error) {
	if !p.Exists() {
		return "", fmt.Errorf("brain not initialized at %s — run `mobiai brain init` first", p.Root)
	}
	if err := e.Validate(); err != nil {
		return "", err
	}
	if e.now == nil {
		e.now = timeNow
	}
	id := makeID(e.Title, e.now())
	rendered := renderEntry(id, e)

	fileBase, _ := e.Type.memoryFileFor()
	target := filepath.Join(p.MemoriesDir, fileBase)
	if err := appendToFile(target, rendered); err != nil {
		return "", err
	}
	return id, nil
}

// makeID builds a stable identifier: <slug-of-title>-<yyyymmdd-hhmmss>.
// The timestamp keeps two saves with identical titles distinguishable.
func makeID(title string, now time.Time) string {
	slug := slugify(title)
	if slug == "" {
		slug = "entry"
	}
	return fmt.Sprintf("%s-%s", slug, now.UTC().Format("20060102-150405"))
}

var nonSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// slugify lowercases, replaces non-alphanumerics with `-`, trims dashes,
// and caps length so an oversized title doesn't bloat the id.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonSlugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	const maxLen = 60
	if len(s) > maxLen {
		s = strings.TrimRight(s[:maxLen], "-")
	}
	return s
}

// renderEntry produces the Markdown section appended to memories/<file>.md.
// Layout matches the spec: H2 title, metadata bullet list, body as-is,
// optional Files list. Empty optional fields are omitted entirely (no
// `platform: ` lines with trailing whitespace).
func renderEntry(id string, e *SaveEntry) string {
	var b strings.Builder
	// Always start with a blank line so we don't fuse with whatever was
	// previously the last line of the file.
	b.WriteString("\n## ")
	b.WriteString(strings.TrimSpace(e.Title))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "- id: %s\n", id)
	fmt.Fprintf(&b, "- type: %s\n", e.Type.fullTypeLabel(e.Status))
	fmt.Fprintf(&b, "- status: %s\n", e.Status)
	if e.Platform != "" {
		fmt.Fprintf(&b, "- platform: %s\n", e.Platform)
	}
	if e.Area != "" {
		fmt.Fprintf(&b, "- area: %s\n", e.Area)
	}
	fmt.Fprintf(&b, "- date: %s\n", e.now().UTC().Format("2006-01-02"))
	if e.ReviewAfter != "" {
		fmt.Fprintf(&b, "- review_after: %s\n", e.ReviewAfter)
	}

	body := strings.TrimSpace(e.Body)
	if body != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}

	if len(e.Files) > 0 {
		b.WriteString("\n### Files\n")
		for _, f := range e.Files {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			fmt.Fprintf(&b, "- %s\n", f)
		}
	}
	return b.String()
}

// appendToFile opens target in append mode (creating it if needed) and
// writes content. Used instead of write-temp+rename because Markdown
// memories are append-only — we never need atomic full-file replacement.
func appendToFile(target, content string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(target), err)
	}
	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", target, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	return nil
}
