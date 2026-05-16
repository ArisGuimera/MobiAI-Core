package brain

import "testing"

func entry(meta map[string]string, title, body string) MemoryEntry {
	return MemoryEntry{Title: title, Body: body, Meta: meta}
}

func TestEntryFilter_EmptyMatchesEverything(t *testing.T) {
	f := EntryFilter{}
	if !f.Empty() {
		t.Error("zero filter should be empty")
	}
	if !f.Matches(entry(nil, "x", "y")) {
		t.Error("empty filter should match any entry")
	}
}

func TestEntryFilter_Platform(t *testing.T) {
	e1 := entry(map[string]string{"platform": "ios"}, "a", "")
	e2 := entry(map[string]string{"platform": "android"}, "b", "")
	e3 := entry(map[string]string{"platform": "IOS"}, "c", "") // case-insensitive

	f := EntryFilter{Platform: "ios"}
	if !f.Matches(e1) || f.Matches(e2) || !f.Matches(e3) {
		t.Errorf("platform filter wrong: e1=%v e2=%v e3=%v",
			f.Matches(e1), f.Matches(e2), f.Matches(e3))
	}
}

func TestEntryFilter_StatusExactMatch(t *testing.T) {
	f := EntryFilter{Status: "temporary"}
	// Status is an exact (fold) match — `active` must NOT pass when
	// `temporary` is asked, even though they're both valid statuses.
	if f.Matches(entry(map[string]string{"status": "active"}, "x", "")) {
		t.Error("active should not match status=temporary")
	}
	if !f.Matches(entry(map[string]string{"status": "Temporary"}, "x", "")) {
		t.Error("Temporary should match status=temporary (case-insensitive)")
	}
}

func TestEntryFilter_AreaSubstring(t *testing.T) {
	// Area uses substring match (folded) so `firebase` matches both
	// `firebase_auth` and `Firebase Core`.
	f := EntryFilter{Area: "firebase"}
	for _, area := range []string{"firebase_auth", "Firebase Core", "firebase"} {
		if !f.Matches(entry(map[string]string{"area": area}, "x", "")) {
			t.Errorf("area %q should match `firebase`", area)
		}
	}
	if f.Matches(entry(map[string]string{"area": "dependency_injection"}, "x", "")) {
		t.Error("dependency_injection should not match firebase")
	}
}

func TestEntryFilter_QueryMatchesTitleOrBody(t *testing.T) {
	f := EntryFilter{Query: "koin"}
	if !f.Matches(entry(nil, "Use Koin as DI", "")) {
		t.Error("query should match title")
	}
	if !f.Matches(entry(nil, "DI choice", "we use Koin for KMP")) {
		t.Error("query should match body")
	}
	if f.Matches(entry(nil, "Hilt setup", "Dagger Hilt config")) {
		t.Error("entry without koin should not match")
	}
}

func TestEntryFilter_AndSemantics(t *testing.T) {
	f := EntryFilter{Platform: "ios", Status: "temporary"}
	cases := []struct {
		meta map[string]string
		want bool
	}{
		{map[string]string{"platform": "ios", "status": "temporary"}, true},
		{map[string]string{"platform": "ios", "status": "active"}, false},
		{map[string]string{"platform": "android", "status": "temporary"}, false},
	}
	for _, c := range cases {
		if got := f.Matches(entry(c.meta, "x", "")); got != c.want {
			t.Errorf("meta=%v: got %v, want %v", c.meta, got, c.want)
		}
	}
}

func TestFilterEntries_DropsNonMatches(t *testing.T) {
	in := []MemoryEntry{
		entry(map[string]string{"platform": "ios"}, "a", ""),
		entry(map[string]string{"platform": "android"}, "b", ""),
		entry(map[string]string{"platform": "ios"}, "c", ""),
	}
	out := FilterEntries(in, EntryFilter{Platform: "ios"})
	if len(out) != 2 || out[0].Title != "a" || out[1].Title != "c" {
		t.Errorf("got %v", titles(out))
	}
}

func titles(es []MemoryEntry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Title
	}
	return out
}
