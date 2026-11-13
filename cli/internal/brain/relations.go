package brain

import (
	"fmt"
	"strings"
	"unicode"
)

// SimilarNoteError is returned by AppendEntry when the section already has
// an active note with a similar title. `save` stops instead of warning:
// with a warning, the duplicate would already be saved.
type SimilarNoteError struct {
	ID    string
	Title string
}

func (e *SimilarNoteError) Error() string {
	return fmt.Sprintf("A similar note already exists: %s %q. Edit that note file, pass --supersedes %s to replace it, or pass --new to save a separate note.", e.ID, e.Title, e.ID)
}

// similarThreshold is the share of title words two notes must have in
// common (over the longer title) to count as similar.
const similarThreshold = 0.6

// stopWords are left out when comparing titles (English and Spanish).
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "of": true, "on": true, "in": true, "at": true, "to": true,
	"for": true, "and": true, "or": true, "with": true, "by": true, "from": true, "as": true,
	"is": true, "are": true, "be": true, "when": true, "into": true, "via": true, "vs": true,
	"de": true, "del": true, "la": true, "el": true, "los": true, "las": true, "un": true,
	"una": true, "en": true, "y": true, "o": true, "con": true, "por": true, "para": true,
	"al": true, "que": true, "se": true, "su": true, "sus": true, "es": true, "cuando": true,
}

// titleWords returns the distinct meaningful words of a title, lowercased.
func titleWords(title string) []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len([]rune(w)) < 2 || stopWords[w] || seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
	}
	return out
}

// similarTitles reports whether two titles share at least 60% of their
// meaningful words (counted over the longer one).
func similarTitles(a, b string) bool {
	wa, wb := titleWords(a), titleWords(b)
	if len(wa) == 0 || len(wb) == 0 {
		return false
	}
	inA := map[string]bool{}
	for _, w := range wa {
		inA[w] = true
	}
	common := 0
	for _, w := range wb {
		if inA[w] {
			common++
		}
	}
	longest := len(wa)
	if len(wb) > longest {
		longest = len(wb)
	}
	return float64(common) >= similarThreshold*float64(longest)
}

// findSimilar returns the first active (or temporary) note in section whose
// title is similar to title and that nothing supersedes, or nil.
func findSimilar(notes []MemoryEntry, section, title string) *MemoryEntry {
	superseded := map[string]bool{}
	for _, n := range notes {
		for _, id := range splitList(n.Get("supersedes")) {
			superseded[id] = true
		}
	}
	for i, n := range notes {
		if n.Section != section || superseded[n.Get("id")] ||
			strings.EqualFold(strings.TrimSpace(n.Get("status")), StatusDeprecated) {
			continue
		}
		if similarTitles(n.Title, title) {
			return &notes[i]
		}
	}
	return nil
}

// applySupersedes marks every note that a newer note replaces. The
// relation lives only in the new note (`supersedes:`), so a replacement
// writes one file and can't conflict. A replacement that is itself
// unconfirmed doesn't retire the old note yet: an unreviewed note must not
// switch off a project rule; the old note just points at the proposal.
// Flags.Unconfirmed must already be set.
func applySupersedes(notes []MemoryEntry) {
	index := map[string]int{}
	for i, n := range notes {
		index[n.Get("id")] = i
	}
	for _, n := range notes {
		for _, old := range splitList(n.Get("supersedes")) {
			i, ok := index[old]
			if !ok || old == n.Get("id") {
				continue
			}
			if n.Flags.Unconfirmed {
				if notes[i].Flags.ReplacementProposed == "" {
					notes[i].Flags.ReplacementProposed = n.Get("id")
				}
				continue
			}
			if notes[i].Flags.SupersededBy == "" {
				notes[i].Flags.SupersededBy = n.Get("id")
			}
		}
	}
}

// splitList splits a comma-separated metadata value, trimming blanks.
func splitList(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
