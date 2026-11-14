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
	// SaveTypeKnowledge is how the app and the services it talks to behave
	// (flows, server error answers, SDK behavior): only what reading the
	// code doesn't tell.
	SaveTypeKnowledge SaveType = "knowledge"
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
	case SaveTypeKnowledge:
		return "knowledge"
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
	case SaveTypeKnowledge:
		return "integrations.md", nil
	default:
		return "", fmt.Errorf("unknown save type: %q (expected: decision|bugfix|testing|knowledge)", t)
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
	case SaveTypeKnowledge:
		return "knowledge"
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
	Supersedes  []string // ids of the notes this one replaces (optional)
	Until       []string // <dependency>@<version> conditions (optional)
	Crash       []string // crash signatures, bugfix only (optional)
	New         bool     // save even if a note with a similar title exists
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
	e.Supersedes = cleanList(e.Supersedes)
	e.Until = cleanList(e.Until)
	if err := validateUntil(e.Until); err != nil {
		return err
	}
	e.Crash = cleanList(e.Crash)
	if len(e.Crash) > 0 && e.Type != SaveTypeBugfix {
		return fmt.Errorf("--crash is only for bugfix notes")
	}
	return nil
}

// cleanList trims the items of a list flag and drops empty ones.
func cleanList(in []string) []string {
	var out []string
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// AppendEntry validates the entry, ensures the brain exists at p, and
// writes the note as its own card at memories/<section>/<id>.md. Returns
// the generated id (slug + timestamp) so callers can echo it. Two saves
// with the same title in the same second get "-2", "-3"... appended.
//
// It refuses to save (SimilarNoteError) when the section already has an
// active note with a similar title, unless e.New is set or the entry
// supersedes another note; the ids in e.Supersedes must exist.
func AppendEntry(p BrainPaths, e *SaveEntry) (string, error) {
	if !p.Exists() {
		return "", fmt.Errorf("brain not initialized at %s — run `mobiai brain init` first", p.Root)
	}
	if err := e.Validate(); err != nil {
		return "", err
	}
	if len(e.Supersedes) > 0 || !e.New {
		notes, err := LoadNotes(p)
		if err != nil {
			return "", err
		}
		known := map[string]bool{}
		for _, n := range notes {
			known[n.Get("id")] = true
		}
		for _, id := range e.Supersedes {
			if !known[id] {
				return "", fmt.Errorf("--supersedes: no note with id %q in the brain", id)
			}
		}
		if len(e.Supersedes) == 0 && !e.New {
			if s := findSimilar(notes, e.Type.section(), e.Title); s != nil {
				return "", &SimilarNoteError{ID: s.Get("id"), Title: s.Title}
			}
		}
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

// combiningMarkRe matches the accents of decomposed text (NFD), where "ó"
// is written as "o" followed by U+0301, as some macOS tools produce it.
var combiningMarkRe = regexp.MustCompile(`[\x{0300}-\x{036f}]`)

// latinFold turns the accented Latin letters people type in titles into
// plain letters, so "conexión" gives "conexion" instead of "conexi-n". It
// runs after ToLower, so only lowercase forms are listed. Letters it doesn't
// know still become a hyphen, as before.
//
// Migration derives the ids of hand-written notes from this function, so
// changing it changes those ids: two mobiai versions that disagree here
// would migrate the same note into two cards.
var latinFold = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ä", "a", "ã", "a", "å", "a", "ā", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e", "ē", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i", "ī", "i",
	"ó", "o", "ò", "o", "ô", "o", "ö", "o", "õ", "o", "ø", "o", "ō", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u", "ū", "u",
	"ñ", "n", "ç", "c", "ý", "y", "ÿ", "y",
	"ß", "ss", "æ", "ae", "œ", "oe",
)

// slugify lowercases, turns accented letters into plain ones, replaces the
// remaining non-alphanumerics with `-`, trims dashes, and caps length so an
// oversized title doesn't bloat the id.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = combiningMarkRe.ReplaceAllString(s, "")
	s = latinFold.Replace(s)
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
	if len(e.Supersedes) > 0 {
		fmt.Fprintf(&b, "- supersedes: %s\n", strings.Join(e.Supersedes, ", "))
	}
	if len(e.Until) > 0 {
		fmt.Fprintf(&b, "- until: %s\n", strings.Join(e.Until, ", "))
	}
	if len(e.Crash) > 0 {
		fmt.Fprintf(&b, "- crash: %s\n", strings.Join(e.Crash, ", "))
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
