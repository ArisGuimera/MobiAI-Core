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

// section returns the canonical section a SaveType writes to.
func (t SaveType) section() string {
	switch t {
	case SaveTypeDecision:
		return "decisions"
	case SaveTypeBugfix:
		return "bugfixes"
	case SaveTypeTesting:
		return "testing"
	}
	return ""
}

// CardDir returns the directory under memories/ where cards of this type
// live (e.g. "decisions").
func (t SaveType) CardDir() string {
	for _, mf := range MemoryFiles {
		if canonicalSection(mf.Name) == t.section() {
			return mf.Dir
		}
	}
	return ""
}

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
// writes the note as its own card at memories/<section>/<id>.md. Returns
// the generated id (slug + timestamp) so callers can echo it. Two saves
// with the same title in the same second get "-2", "-3"... appended.
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
	dir := filepath.Join(p.MemoriesDir, e.Type.CardDir())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	base := makeID(e.Title, e.now())
	id := base
	for n := 2; ; n++ {
		if _, err := os.Stat(filepath.Join(dir, id+".md")); os.IsNotExist(err) {
			break
		}
		id = fmt.Sprintf("%s-%d", base, n)
	}
	if err := atomicWrite(filepath.Join(dir, id+".md"), []byte(renderCard(id, e))); err != nil {
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

// renderCard produces the content of a card file. Layout matches the
// 0.2 entries (H2 title, metadata bullets, body, optional Files list) plus
// `verified`, which starts equal to `date`. Empty optional fields are
// omitted entirely (no `platform: ` lines with trailing whitespace).
func renderCard(id string, e *SaveEntry) string {
	var b strings.Builder
	b.WriteString("## ")
	b.WriteString(strings.TrimSpace(e.Title))
	b.WriteString("\n\n")

	date := e.now().UTC().Format("2006-01-02")
	fmt.Fprintf(&b, "- id: %s\n", id)
	fmt.Fprintf(&b, "- type: %s\n", e.Type.fullTypeLabel(e.Status))
	fmt.Fprintf(&b, "- status: %s\n", e.Status)
	if e.Platform != "" {
		fmt.Fprintf(&b, "- platform: %s\n", e.Platform)
	}
	if e.Area != "" {
		fmt.Fprintf(&b, "- area: %s\n", e.Area)
	}
	fmt.Fprintf(&b, "- date: %s\n", date)
	fmt.Fprintf(&b, "- verified: %s\n", date)
	if e.ReviewAfter != "" {
		fmt.Fprintf(&b, "- review_after: %s\n", e.ReviewAfter)
	}

	body := strings.TrimSpace(demoteBodyH2(normalizeNewlines(e.Body)))
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

// demoteBodyH2 turns `## ` headings in a note body into `### ` so the card
// also reads correctly with tools that split on H2. Lines inside fenced
// code blocks (``` or ~~~) are left alone: there `##` is code, not Markdown.
func demoteBodyH2(body string) string {
	lines := strings.Split(body, "\n")
	fence := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			fence = "```"
			continue
		}
		if strings.HasPrefix(trimmed, "~~~") {
			fence = "~~~"
			continue
		}
		if strings.HasPrefix(line, "## ") {
			lines[i] = "#" + line
		}
	}
	return strings.Join(lines, "\n")
}
