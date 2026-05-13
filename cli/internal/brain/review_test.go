package brain

import (
	"path/filepath"
	"testing"
	"time"
)

// fixedNow is the reference "today" used across review tests. Anchored
// to a real date so calendar arithmetic is easy to eyeball.
var fixedNow = time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)

func TestReview_OverdueTemporaryAppears(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Podfile composeApp en minúsculas

- id: podfile-x
- status: temporary
- platform: ios
- review_after: 2026-03-01

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 overdue item, got %d", len(items))
	}
	it := items[0]
	if it.Entry.Title != "Podfile composeApp en minúsculas" {
		t.Errorf("wrong entry: %q", it.Entry.Title)
	}
	if !it.HasDate {
		t.Errorf("HasDate should be true")
	}
	// 2026-05-13 minus 2026-03-01 = 31 (March) + 30 (April) + 12 (May 1..12 inclusive minus 1) = 73 days
	if it.DaysOverdue != 73 {
		t.Errorf("DaysOverdue = %d, want 73", it.DaysOverdue)
	}
	if it.Section != "bugfixes" {
		t.Errorf("Section = %q, want bugfixes", it.Section)
	}
}

func TestReview_DueTodayCountsAsOverdue(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Due today

- id: today
- status: temporary
- review_after: 2026-05-13

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("review_after = today should be overdue; got %d", len(items))
	}
	if items[0].DaysOverdue != 0 {
		t.Errorf("DaysOverdue = %d, want 0", items[0].DaysOverdue)
	}
}

func TestReview_FutureNotOverdue(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Future review

- id: future
- status: temporary
- review_after: 2026-12-01

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("future review_after should not be overdue; got %d items", len(items))
	}
}

func TestReview_ActiveAndDeprecatedExcluded(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "decisions.md"),
		`## Active with past review

- id: a1
- status: active
- review_after: 2026-01-01

body

## Deprecated with past review

- id: a2
- status: deprecated
- review_after: 2026-01-01

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("review only surfaces temporary entries; got %d items", len(items))
	}
}

func TestReview_NoDateExcludedByDefault(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Temporary without review_after

- id: nodate
- status: temporary

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("temporary without date should be skipped by default; got %d", len(items))
	}
}

func TestReview_NoDateIncludedWithFlag(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Temporary without review_after

- id: nodate
- status: temporary

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow, IncludeNoDate: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("IncludeNoDate should surface dateless temporaries; got %d", len(items))
	}
	if items[0].HasDate {
		t.Errorf("HasDate should be false for no-date entry")
	}
	if items[0].ReviewAfter != "" {
		t.Errorf("ReviewAfter should be empty; got %q", items[0].ReviewAfter)
	}
}

func TestReview_MalformedDateSkippedSilently(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Malformed date

- id: bad
- status: temporary
- review_after: not-a-real-date

body

## Valid overdue

- id: valid
- status: temporary
- review_after: 2026-01-01

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatalf("malformed date in one entry should not error the whole review: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 (only valid one); got %d", len(items))
	}
	if items[0].Entry.Title != "Valid overdue" {
		t.Errorf("wrong entry survived: %q", items[0].Entry.Title)
	}
}

func TestReview_SortedMostOverdueFirst(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Older

- id: older
- status: temporary
- review_after: 2026-01-01

body

## Newer

- id: newer
- status: temporary
- review_after: 2026-04-01

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Entry.Title != "Older" {
		t.Errorf("most-overdue should come first; got %q first", items[0].Entry.Title)
	}
	if items[0].DaysOverdue <= items[1].DaysOverdue {
		t.Errorf("DaysOverdue should be decreasing; got %d then %d",
			items[0].DaysOverdue, items[1].DaysOverdue)
	}
}

func TestReview_DatedBeforeNoDate(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Without date

- id: nodate
- status: temporary

body

## With overdue date

- id: dated
- status: temporary
- review_after: 2026-01-01

body
`)
	items, err := Review(p, ReviewOptions{Now: fixedNow, IncludeNoDate: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if !items[0].HasDate {
		t.Errorf("dated entries should come first; got no-date first")
	}
	if items[1].HasDate {
		t.Errorf("no-date entries should come last; got dated last")
	}
}

func TestReview_DefaultsToRealNowWhenZero(t *testing.T) {
	// Plant a review_after far in the past so it's overdue regardless
	// of what "now" the implementation picks up.
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.MemoriesDir, "bugfixes.md"),
		`## Ancient temp

- id: ancient
- status: temporary
- review_after: 2000-01-01

body
`)
	items, err := Review(p, ReviewOptions{}) // Now: zero → uses time.Now
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("a 2000-01-01 review_after should always be overdue; got %d items", len(items))
	}
}

func TestReview_EmptyBrainNoItems(t *testing.T) {
	p := initBrainForTest(t)
	items, err := Review(p, ReviewOptions{Now: fixedNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("empty brain should produce no items; got %d", len(items))
	}
}
