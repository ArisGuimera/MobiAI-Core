package host

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestAdapter() *genericAdapter {
	return &genericAdapter{
		id:         "test",
		name:       "Test Host",
		homepage:   "https://example.com",
		homeSubdir: ".test-host",
		caps:       Caps{Skills: true},
	}
}

func TestGenericAdapter_Identity(t *testing.T) {
	a := newTestAdapter()
	if a.ID() != "test" {
		t.Errorf("ID: got %q", a.ID())
	}
	if a.Name() != "Test Host" {
		t.Errorf("Name: got %q", a.Name())
	}
	if a.Homepage() != "https://example.com" {
		t.Errorf("Homepage: got %q", a.Homepage())
	}
}

func TestGenericAdapter_Capabilities(t *testing.T) {
	caps := newTestAdapter().Capabilities()
	if !caps.Skills {
		t.Error("Capabilities.Skills should be true")
	}
}

func TestGenericAdapter_SkillsDir(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	want := filepath.Join(tmp, ".test-host", "skills")
	if got := newTestAdapter().SkillsDir(); got != want {
		t.Errorf("SkillsDir: got %q, want %q", got, want)
	}
}

func TestGenericAdapter_Detect_Found(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	if err := os.MkdirAll(filepath.Join(tmp, ".test-host"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := newTestAdapter().Detect()
	if !r.Found {
		t.Errorf("Detect.Found: got false (Searched=%v)", r.Searched)
	}
}

func TestGenericAdapter_Detect_NotFound(t *testing.T) {
	tmp := t.TempDir()
	setFakeHome(t, tmp)
	r := newTestAdapter().Detect()
	if r.Found {
		t.Error("Detect.Found: got true, want false")
	}
	if len(r.Searched) == 0 {
		t.Error("Detect.Searched: got empty")
	}
}
