package host

import "testing"

// setFakeHome sets HOME and USERPROFILE so os.UserHomeDir() returns dir
// on every supported platform.
func setFakeHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}
