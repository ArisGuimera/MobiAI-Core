package host

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestNewDefaultRegistry_Has36Adapters(t *testing.T) {
	r := NewDefaultRegistry()
	if got := len(r.Adapters()); got != 36 {
		t.Fatalf("Adapters: got %d, want 36 (4 tier-1 + 5 tier-2 + 27 tier-3)", got)
	}
}

func TestRegistry_Get_ByID(t *testing.T) {
	r := NewDefaultRegistry()
	a, err := r.Get("cursor")
	if err != nil || a.ID() != "cursor" {
		t.Errorf("Get(cursor): got %v / %v", a, err)
	}
	if _, err := r.Get("nope"); err == nil {
		t.Error("Get(nope): expected error")
	}
}

func TestRegistry_Detect_OnlyPresent(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	for _, sub := range []string{".claude", ".gemini"} {
		if err := os.MkdirAll(filepath.Join(tmp, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	present := NewDefaultRegistry().Detect()
	ids := make([]string, 0, len(present))
	for _, a := range present {
		ids = append(ids, a.ID())
	}
	sort.Strings(ids)
	want := []string{"claude-code", "gemini"}
	if len(ids) != 2 || ids[0] != want[0] || ids[1] != want[1] {
		t.Errorf("Detect: got %v, want %v", ids, want)
	}
}

func TestRegistry_Detect_ExcludesTier3ByDefault(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	// Create dirs for tier-1 (.claude) AND a couple of tier-3 (.mux, .factory).
	// .mux/.factory are precisely the kind of directories users may have for
	// completely unrelated reasons — auto-detection must not hijack them.
	for _, sub := range []string{".claude", ".mux", ".factory"} {
		if err := os.MkdirAll(filepath.Join(tmp, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	got := NewDefaultRegistry().Detect()
	ids := make([]string, 0, len(got))
	for _, a := range got {
		ids = append(ids, a.ID())
	}
	sort.Strings(ids)
	if len(ids) != 1 || ids[0] != "claude-code" {
		t.Errorf("Detect (default): got %v, want [claude-code] only", ids)
	}
}

func TestRegistry_Detect_IncludesTier3WhenOptedIn(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	for _, sub := range []string{".claude", ".mux"} {
		if err := os.MkdirAll(filepath.Join(tmp, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	r := NewDefaultRegistry()
	r.SetIncludeExperimental(true)
	got := r.Detect()
	ids := make([]string, 0, len(got))
	for _, a := range got {
		ids = append(ids, a.ID())
	}
	sort.Strings(ids)
	want := []string{"claude-code", "mux"}
	if len(ids) != 2 || ids[0] != want[0] || ids[1] != want[1] {
		t.Errorf("Detect (experimental): got %v, want %v", ids, want)
	}
}

func TestRegistry_Get_ResolvesTier3Regardless(t *testing.T) {
	r := NewDefaultRegistry()
	a, err := r.Get("mux")
	if err != nil {
		t.Fatalf("Get(mux): %v — tier-3 must be resolvable for explicit --host", err)
	}
	if !a.Experimental() {
		t.Error("mux should report Experimental() == true")
	}
}
