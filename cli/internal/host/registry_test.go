package host

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestNewDefaultRegistry_Has4Adapters(t *testing.T) {
	r := NewDefaultRegistry()
	if got := len(r.Adapters()); got != 4 {
		t.Fatalf("Adapters: got %d, want 4", got)
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
