package brain

import "strings"

// EntryFilter narrows the set of MemoryEntry values that will be
// rendered or searched. Semantics are AND across non-empty fields:
// every condition must pass for the entry to match. Empty fields are
// no-ops (don't restrict).
//
// String comparisons are case-insensitive and trimmed.
type EntryFilter struct {
	Platform string // matches the `platform:` meta field
	Status   string // matches the `status:` meta field
	Area     string // matches the `area:` meta field (substring)
	Query    string // free-text, matches Title or Body (substring)
}

// Empty reports whether the filter would let everything through.
func (f EntryFilter) Empty() bool {
	return f.Platform == "" && f.Status == "" && f.Area == "" && f.Query == ""
}

// Matches reports whether e satisfies every set condition.
func (f EntryFilter) Matches(e MemoryEntry) bool {
	if f.Platform != "" && !eqFold(e.Get("platform"), f.Platform) {
		return false
	}
	if f.Status != "" && !eqFold(e.Get("status"), f.Status) {
		return false
	}
	if f.Area != "" && !containsFold(e.Get("area"), f.Area) {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(strings.TrimSpace(f.Query))
		if !strings.Contains(strings.ToLower(e.Title), q) &&
			!strings.Contains(strings.ToLower(e.Body), q) {
			return false
		}
	}
	return true
}

// FilterEntries returns the subset of in for which f.Matches is true.
// Returns a non-nil empty slice when nothing matches so callers can
// distinguish "no entries to begin with" from "nothing matched".
func FilterEntries(in []MemoryEntry, f EntryFilter) []MemoryEntry {
	out := make([]MemoryEntry, 0, len(in))
	for _, e := range in {
		if f.Matches(e) {
			out = append(out, e)
		}
	}
	return out
}

func eqFold(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(
		strings.ToLower(strings.TrimSpace(haystack)),
		strings.ToLower(strings.TrimSpace(needle)),
	)
}
