package brain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// reviewAfterLayout is the on-disk format of the review_after metadata
// field. Plain calendar date in UTC — no time component, no timezone.
const reviewAfterLayout = "2006-01-02"

// ReviewItem is a single entry surfaced by Review. It bundles the parsed
// MemoryEntry with the derived "how overdue is this" information so
// callers don't recompute it.
type ReviewItem struct {
	Section     string      // canonical section name (decisions/bugfixes/...)
	Entry       MemoryEntry // the overdue entry
	ReviewAfter string      // raw review_after meta value (YYYY-MM-DD); "" if HasDate is false
	DaysOverdue int         // (today - review_after) in days; 0 = due today; only meaningful when HasDate
	HasDate     bool        // true when entry had a valid review_after; false for "no-date" entries
}

// ReviewOptions configures Review. The zero value is valid and matches
// the CLI defaults (only overdue entries, "today" = real now).
type ReviewOptions struct {
	// Now is the reference "today" used to decide if an entry is
	// overdue. When zero, time.Now().UTC() is used. Tests inject a
	// fixed value to keep date assertions deterministic.
	Now time.Time

	// IncludeNoDate, when true, also includes temporary entries that
	// lack a review_after meta field. They land with HasDate=false on
	// the result so the caller can render them separately.
	IncludeNoDate bool
}

// Review returns every memory entry with `status: temporary` whose
// `review_after` has passed (review_after <= today). With
// IncludeNoDate, also returns temporary entries without a review_after.
//
// Entries with malformed review_after dates are skipped silently — the
// rest of the brain is permissive about user edits, and we don't want a
// typo in one entry to break the whole audit. Active and deprecated
// entries are always excluded, regardless of review_after.
//
// Results are sorted most-overdue first; no-date entries come after all
// dated ones, in canonical section order.
func Review(p BrainPaths, opts ReviewOptions) ([]ReviewItem, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	// Truncate to UTC date so comparisons are calendar-based: a
	// review_after of 2026-05-13 is "due" the entire day in UTC, not
	// from a specific second onward.
	today := truncateToDay(now)

	groups, err := ParseAllMemories(p)
	if err != nil {
		return nil, fmt.Errorf("parsear memorias: %w", err)
	}

	var items []ReviewItem
	for _, section := range AllSectionNames() {
		for _, e := range groups[section] {
			if !strings.EqualFold(strings.TrimSpace(e.Get("status")), "temporary") {
				continue
			}
			raw := strings.TrimSpace(e.Get("review_after"))
			if raw == "" {
				if !opts.IncludeNoDate {
					continue
				}
				items = append(items, ReviewItem{
					Section: section,
					Entry:   e,
					HasDate: false,
				})
				continue
			}
			t, err := time.Parse(reviewAfterLayout, raw)
			if err != nil {
				// Malformed date — skip silently. The entry is still in
				// the brain; user can fix the date and re-run review.
				continue
			}
			t = truncateToDay(t)
			if t.After(today) {
				// Not overdue yet.
				continue
			}
			days := int(today.Sub(t).Hours() / 24)
			items = append(items, ReviewItem{
				Section:     section,
				Entry:       e,
				ReviewAfter: raw,
				DaysOverdue: days,
				HasDate:     true,
			})
		}
	}

	// Sort: dated entries first (most-overdue → least-overdue), then
	// no-date entries last in canonical section order.
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].HasDate != items[j].HasDate {
			return items[i].HasDate
		}
		if !items[i].HasDate {
			return false // preserve insertion order for no-date
		}
		return items[i].DaysOverdue > items[j].DaysOverdue
	})

	return items, nil
}

// truncateToDay returns t in UTC with hour/minute/second/nanosecond zero.
func truncateToDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
