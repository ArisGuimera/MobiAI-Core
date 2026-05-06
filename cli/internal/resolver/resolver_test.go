package resolver

import (
	"errors"
	"strings"
	"testing"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

// fakeCatalog builds a minimal Catalog-shaped value for resolver tests.
// We don't need the real Load() — just the byName/Packs structure.
func fakeCatalog(packs map[string][]string) *catalog.Catalog {
	c := &catalog.Catalog{
		Marketplace: &catalog.Marketplace{},
	}
	names := make([]string, 0, len(packs))
	for name := range packs {
		names = append(names, name)
	}
	for _, name := range names {
		c.Packs = append(c.Packs, catalog.Pack{
			Ref:      catalog.PluginRef{Name: name},
			Manifest: catalog.PluginManifest{Name: name, Dependencies: packs[name]},
		})
	}
	// Index — Catalog doesn't expose byName; rebuild via re-Load semantics.
	// We instead use the public Has/Get which scan Packs in our exported helper.
	// To keep tests pure, we call newCatalogIndex from the resolver pkg.
	rebuildIndex(c)
	return c
}

func TestResolve_SinglePackNoDeps(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"core": nil,
	})
	got, err := Resolve(c, []string{"core"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 1 || got[0] != "core" {
		t.Errorf("got %v, want [core]", got)
	}
}

func TestResolve_LinearDeps(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"core":    nil,
		"android": {"core"},
	})
	got, err := Resolve(c, []string{"android"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"core", "android"}
	if !equalStringSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResolve_DiamondDeps(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"core":    nil,
		"android": {"core"},
		"ios":     {"core"},
		"kmp":     {"core", "android", "ios"},
	})
	got, err := Resolve(c, []string{"kmp"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// Expectation: deps come before dependents; core first; kmp last.
	if len(got) != 4 {
		t.Fatalf("got %v, want 4 entries", got)
	}
	if got[0] != "core" {
		t.Errorf("first = %q, want core", got[0])
	}
	if got[len(got)-1] != "kmp" {
		t.Errorf("last = %q, want kmp", got[len(got)-1])
	}
	if !contains(got, "android") || !contains(got, "ios") {
		t.Errorf("missing android or ios in %v", got)
	}
}

func TestResolve_DuplicateRequest(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"core":    nil,
		"android": {"core"},
	})
	got, err := Resolve(c, []string{"android", "android", "core"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %v, want 2 unique entries", got)
	}
}

func TestResolve_MissingDep(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"android": {"core"}, // core not in catalog
	})
	_, err := Resolve(c, []string{"android"})
	if err == nil {
		t.Fatal("expected error for missing dep, got nil")
	}
	var me *MissingDepError
	if !errors.As(err, &me) {
		t.Fatalf("expected *MissingDepError, got %T", err)
	}
	if me.Missing != "core" {
		t.Errorf("Missing: got %q, want %q", me.Missing, "core")
	}
}

func TestResolve_MissingRequested(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"core": nil,
	})
	_, err := Resolve(c, []string{"androyd"})
	if err == nil {
		t.Fatal("expected error for missing requested pack, got nil")
	}
	if !strings.Contains(err.Error(), "androyd") {
		t.Errorf("error should mention pack name; got: %v", err)
	}
}

func TestResolve_Cycle(t *testing.T) {
	c := fakeCatalog(map[string][]string{
		"a": {"b"},
		"b": {"a"},
	})
	_, err := Resolve(c, []string{"a"})
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *CycleError, got %T", err)
	}
	if len(ce.Cycle) < 2 {
		t.Errorf("Cycle should have at least 2 entries, got %v", ce.Cycle)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
