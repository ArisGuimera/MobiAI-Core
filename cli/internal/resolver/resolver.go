// Package resolver expands a list of requested packs into the full
// install order, including all transitive dependencies, with cycle
// detection and missing-dep errors.
package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/catalog"
)

// NotFoundError signals that a pack the user explicitly requested is
// not present in the catalog (typo, stale catalog, etc.).
type NotFoundError struct {
	Pack string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("requested pack %q not found in catalog", e.Pack)
}

// MissingDepError signals that pack From declares a dependency on
// Missing, which is not present in the catalog.
type MissingDepError struct {
	From    string
	Missing string
}

func (e *MissingDepError) Error() string {
	return fmt.Sprintf("pack %q depends on %q, which is not in the catalog", e.From, e.Missing)
}

// CycleError signals a dependency cycle. Cycle is the list of packs
// involved, in order, with the first repeating at the end.
type CycleError struct {
	Cycle []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("dependency cycle detected: %s", strings.Join(e.Cycle, " → "))
}

// Resolve returns the install order for the requested packs:
// every transitive dep appears before its dependents, and the result
// is deduplicated. Order among independent siblings is alphabetical
// for determinism.
func Resolve(c *catalog.Catalog, requested []string) ([]string, error) {
	// 1. Validate every requested name exists in the catalog.
	for _, name := range requested {
		if !c.Has(name) {
			return nil, &NotFoundError{Pack: name}
		}
	}

	// 2. DFS-based topological sort.
	const (
		white = 0 // unvisited
		gray  = 1 // on stack (cycle detection)
		black = 2 // done
	)
	color := make(map[string]int)
	var order []string
	var stack []string

	var visit func(name string) error
	visit = func(name string) error {
		switch color[name] {
		case black:
			return nil
		case gray:
			// Build the cycle slice from the stack starting at name.
			start := -1
			for i, s := range stack {
				if s == name {
					start = i
					break
				}
			}
			cycle := append([]string{}, stack[start:]...)
			cycle = append(cycle, name)
			return &CycleError{Cycle: cycle}
		}
		color[name] = gray
		stack = append(stack, name)

		pack, err := c.Get(name)
		if err != nil {
			// Should not happen — validated above for requested; for transitive,
			// missing dep triggers MissingDepError below.
			return err
		}
		deps := append([]string(nil), pack.Manifest.Dependencies...)
		sort.Strings(deps)
		for _, d := range deps {
			if !c.Has(d) {
				return &MissingDepError{From: name, Missing: d}
			}
			if err := visit(d); err != nil {
				return err
			}
		}

		stack = stack[:len(stack)-1]
		color[name] = black
		order = append(order, name)
		return nil
	}

	roots := append([]string(nil), requested...)
	sort.Strings(roots)
	seen := make(map[string]bool)
	uniqueRoots := roots[:0]
	for _, r := range roots {
		if seen[r] {
			continue
		}
		seen[r] = true
		uniqueRoots = append(uniqueRoots, r)
	}
	for _, r := range uniqueRoots {
		if err := visit(r); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// rebuildIndex is a test helper used by resolver_test.go to seed
// catalog.Catalog's internal index when constructed by hand.
// In production, catalog.Load builds it.
func rebuildIndex(c *catalog.Catalog) {
	// We can't access unexported fields from another package, so the
	// helper exists only as a no-op; resolver tests rely on Has/Get,
	// which scan Packs. We instead use the public Reindex on catalog.
	c.Reindex()
}
