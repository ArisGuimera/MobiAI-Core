package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// updateCheckCache is the on-disk record consumed by the SessionStart hooks
// in each client. The schema is intentionally narrow — clients should treat
// missing/unparseable fields as "no update info available".
type updateCheckCache struct {
	CheckedAt       int64  `json:"checked_at"`
	Installed       string `json:"installed"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url,omitempty"`
}

// defaultUpdateCheckURL is the GitHub Releases API endpoint we consult. The
// tag filter (`cli-v*`) is applied client-side after fetching the list.
// Overridable for tests via MOBIAI_UPDATE_CHECK_URL.
const defaultUpdateCheckURL = "https://api.github.com/repos/ArisGuimera/MobiAI-Core/releases?per_page=20"

// updateCheckCachePath returns the absolute path to update-check.json under
// the user cache dir (XDG_CACHE_HOME on Linux, ~/Library/Caches on macOS,
// %LocalAppData% on Windows). Mirrors the convention gsd uses with ~/.cache/gsd.
func updateCheckCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mobiai", "update-check.json"), nil
}

// runUpdateCheck fetches the latest cli-v* release from GitHub, compares it
// against the installed version, and writes the result to update-check.json.
// Designed to be cheap and side-effect-free aside from the cache write —
// safe to call from a SessionStart hook or a periodic worker.
//
// installedVersion is the build-time version string (e.g., "0.1.1"). When
// it equals "dev" (local builds), the check writes update_available=false
// because we can't meaningfully compare against semver tags.
func runUpdateCheck(out io.Writer, installedVersion string) error {
	url := os.Getenv("MOBIAI_UPDATE_CHECK_URL")
	if url == "" {
		url = defaultUpdateCheckURL
	}

	latest, latestURL, err := latestRelease(url)
	if err != nil {
		return err
	}

	available := false
	if installedVersion != "" && installedVersion != "dev" {
		available = semverLess(installedVersion, latest)
	}

	cache := updateCheckCache{
		CheckedAt:       time.Now().Unix(),
		Installed:       installedVersion,
		Latest:          latest,
		UpdateAvailable: available,
		ReleaseURL:      latestURL,
	}
	path, err := updateCheckCachePath()
	if err != nil {
		return fmt.Errorf("cache path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}

	if available {
		fmt.Fprintf(out, "MobiAI %s → %s disponible. Actualizá con: mobiai update\n", installedVersion, latest)
	} else if out != io.Discard {
		fmt.Fprintf(out, "MobiAI %s está al día (último release: %s).\n", installedVersion, latest)
	}
	return nil
}

// latestRelease fetches the GitHub releases list at url and returns the
// version and HTML URL of the newest non-draft, non-prerelease `cli-v*`
// release. Releases come newest-first per GitHub docs, so the first match
// wins. Shared by the SessionStart update-check and the binary self-update.
func latestRelease(url string) (version, htmlURL string, err error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "mobiai-cli")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch GitHub releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub releases API returned %d", resp.StatusCode)
	}

	var releases []struct {
		TagName    string `json:"tag_name"`
		HTMLURL    string `json:"html_url"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", "", fmt.Errorf("parse releases JSON: %w", err)
	}

	for _, r := range releases {
		if r.Draft || r.Prerelease {
			continue
		}
		if !strings.HasPrefix(r.TagName, "cli-v") {
			continue
		}
		return strings.TrimPrefix(r.TagName, "cli-v"), r.HTMLURL, nil
	}
	return "", "", fmt.Errorf("no cli-v* release found in the last 20 entries")
}

// semverLess reports whether a < b using a tolerant dotted-number compare.
// Non-numeric suffixes (e.g., "1.2.3-rc1") fall back to byte order on the
// trailing piece, which is good enough for the CLI's flat MAJOR.MINOR.PATCH
// release scheme. Empty strings sort lowest so "" < "0.1.0".
func semverLess(a, b string) bool {
	aa := strings.Split(a, ".")
	bb := strings.Split(b, ".")
	n := len(aa)
	if len(bb) > n {
		n = len(bb)
	}
	for i := 0; i < n; i++ {
		var as, bs string
		if i < len(aa) {
			as = aa[i]
		}
		if i < len(bb) {
			bs = bb[i]
		}
		ai, aErr := strconv.Atoi(as)
		bi, bErr := strconv.Atoi(bs)
		if aErr == nil && bErr == nil {
			if ai != bi {
				return ai < bi
			}
			continue
		}
		// Component contains non-numeric (e.g. "3-rc1"): byte-order compare.
		if as != bs {
			return as < bs
		}
	}
	return false
}
