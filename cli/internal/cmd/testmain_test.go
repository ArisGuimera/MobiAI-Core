package cmd

import (
	"os"
	"testing"
)

// TestMain clears MobiAI env vars that leak from the developer's shell
// and would otherwise short-circuit catalog/registry lookups under test.
// Tests still call t.Setenv to set their own values per-case.
func TestMain(m *testing.M) {
	for _, key := range []string{
		"MOBIAI_CATALOG_ROOT",
		"MOBIAI_CATALOG_GIT_URL",
		"MOBIAI_HOME",
		"MOBIAI_INCLUDE_EXPERIMENTAL",
	} {
		_ = os.Unsetenv(key)
	}
	os.Exit(m.Run())
}
