package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSemverLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"0.1.0", "0.1.1", true},
		{"0.1.1", "0.1.1", false},
		{"0.1.2", "0.1.1", false},
		{"0.2.0", "0.10.0", true}, // numeric compare, not lexicographic
		{"1.0.0", "0.9.9", false},
		{"0.1", "0.1.1", true},     // missing patch = 0 < 1
		{"", "0.0.1", true},        // empty < anything
		// Pre-release suffixes are NOT semver-aware; we fall back to byte
		// compare of the trailing component. "0-rc1" > "0" lexicographically,
		// so semverLess returns false. That's acceptable for MobiAI's flat
		// MAJOR.MINOR.PATCH scheme — we don't publish pre-releases.
		{"0.1.0-rc1", "0.1.0", false},
	}
	for _, c := range cases {
		got := semverLess(c.a, c.b)
		if got != c.want {
			t.Errorf("semverLess(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestRunUpdateCheck_WritesCacheWithUpdateAvailable(t *testing.T) {
	// Stub GitHub releases endpoint returning a single cli-v* release.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tag_name": "cli-v0.2.0", "html_url": "https://example.test/r", "draft": false, "prerelease": false}
		]`))
	}))
	defer srv.Close()

	t.Setenv("MOBIAI_UPDATE_CHECK_URL", srv.URL)
	// Redirect UserCacheDir to a tempdir so the test doesn't touch the
	// user's real cache. Go's os.UserCacheDir uses LOCALAPPDATA on Windows
	// and XDG_CACHE_HOME on Linux; we override both for portability.
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("HOME", tmp) // macOS fallback

	if err := runUpdateCheck(io.Discard, "0.1.0"); err != nil {
		t.Fatalf("runUpdateCheck: %v", err)
	}

	cachePath, err := updateCheckCachePath()
	if err != nil {
		t.Fatalf("cache path: %v", err)
	}
	if !filepath.IsAbs(cachePath) {
		t.Errorf("expected absolute cache path, got %q", cachePath)
	}
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}
	var cache updateCheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		t.Fatalf("unmarshal cache: %v", err)
	}
	if cache.Installed != "0.1.0" {
		t.Errorf("Installed: got %q, want %q", cache.Installed, "0.1.0")
	}
	if cache.Latest != "0.2.0" {
		t.Errorf("Latest: got %q, want %q", cache.Latest, "0.2.0")
	}
	if !cache.UpdateAvailable {
		t.Error("UpdateAvailable: got false, want true")
	}
	if cache.CheckedAt == 0 {
		t.Error("CheckedAt: should be non-zero")
	}
}

func TestRunUpdateCheck_DevVersionMarksNoUpdate(t *testing.T) {
	// When installedVersion == "dev" we can't meaningfully compare, so
	// the cache records update_available=false regardless of latest.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"tag_name": "cli-v9.9.9"}]`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	t.Setenv("MOBIAI_UPDATE_CHECK_URL", srv.URL)
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("HOME", tmp)

	if err := runUpdateCheck(io.Discard, "dev"); err != nil {
		t.Fatal(err)
	}
	cachePath, _ := updateCheckCachePath()
	data, _ := os.ReadFile(cachePath)
	var cache updateCheckCache
	_ = json.Unmarshal(data, &cache)
	if cache.UpdateAvailable {
		t.Errorf("UpdateAvailable should be false for 'dev' build; got true")
	}
}

func TestRunUpdateCheck_SkipsDraftsAndPrereleases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"tag_name": "cli-v2.0.0-rc1", "prerelease": true},
			{"tag_name": "cli-v1.5.0-draft", "draft": true},
			{"tag_name": "cli-v1.0.0", "html_url": "https://example.test/stable"}
		]`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	t.Setenv("MOBIAI_UPDATE_CHECK_URL", srv.URL)
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("HOME", tmp)

	if err := runUpdateCheck(io.Discard, "0.9.0"); err != nil {
		t.Fatal(err)
	}
	cachePath, _ := updateCheckCachePath()
	data, _ := os.ReadFile(cachePath)
	var cache updateCheckCache
	_ = json.Unmarshal(data, &cache)
	if cache.Latest != "1.0.0" {
		t.Errorf("Latest: got %q, want %q (should skip rc/draft)", cache.Latest, "1.0.0")
	}
}
