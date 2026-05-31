package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// defaultInstallBase is the GitHub Releases base URL. Release assets live at
// <base>/download/cli-v<ver>/<archive>. Overridable via MOBIAI_INSTALL_BASE
// (same env var the install scripts honor) for tests and self-hosted mirrors.
const defaultInstallBase = "https://github.com/ArisGuimera/MobiAI-Core/releases"

// selfUpdateBinary updates the running mobiai binary in place when a newer
// release exists. It is the in-process equivalent of re-running the install
// script, but works while mobiai itself is the running process (the install
// script can't overwrite a locked .exe on Windows).
//
// Flow: resolve latest release → download the platform asset → verify its
// SHA256 against checksums.txt → extract the binary → replace the running
// executable via rename-then-write.
//
// Callers treat a returned error as non-fatal: `mobiai update` has already
// refreshed the catalog by the time this runs, so a binary-update failure is
// surfaced as a warning, not a hard exit.
func selfUpdateBinary(out io.Writer, currentVersion string) error {
	bin, latest, err := fetchUpdateBinary(out, currentVersion)
	if err != nil {
		return err
	}
	if bin == nil {
		// Up to date, or a "dev" build we can't compare. fetchUpdateBinary
		// already printed the relevant message (or stayed silent for dev).
		return nil
	}
	// selfReplace is the one line we can't exercise in tests (it resolves the
	// real running executable). Everything above is covered via httptest.
	if err := selfReplace(bin); err != nil {
		return fmt.Errorf("instalar binario nuevo: %w", err)
	}
	fmt.Fprintf(out, "Binario mobiai actualizado a %s. Reiniciá la terminal (o volvé a correr mobiai) para usar la nueva versión.\n", latest)
	return nil
}

// fetchUpdateBinary resolves the latest release, and if it is newer than
// currentVersion, downloads + verifies + extracts the platform binary,
// returning its bytes and the resolved version. It returns (nil, "", nil)
// when there is nothing to do: a "dev"/empty build (can't compare) or the
// binary is already current. All network/IO lives here so the orchestrator's
// only untestable step is the actual on-disk replace.
func fetchUpdateBinary(out io.Writer, currentVersion string) (bin []byte, latest string, err error) {
	if currentVersion == "" || currentVersion == "dev" {
		return nil, "", nil // local/dev build — nothing to compare against
	}

	checkURL := os.Getenv("MOBIAI_UPDATE_CHECK_URL")
	if checkURL == "" {
		checkURL = defaultUpdateCheckURL
	}
	latest, _, err = latestRelease(checkURL)
	if err != nil {
		return nil, "", fmt.Errorf("consultar releases: %w", err)
	}
	if !semverLess(currentVersion, latest) {
		fmt.Fprintf(out, "Binario mobiai %s: al día.\n", currentVersion)
		return nil, "", nil
	}

	base := os.Getenv("MOBIAI_INSTALL_BASE")
	if base == "" {
		base = defaultInstallBase
	}
	archive := assetName(latest, runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(out, "Binario mobiai %s → %s: descargando %s...\n", currentVersion, latest, archive)

	archiveURL := fmt.Sprintf("%s/download/cli-v%s/%s", base, latest, archive)
	data, err := downloadBytes(archiveURL)
	if err != nil {
		return nil, "", fmt.Errorf("descargar %s: %w", archive, err)
	}

	sumsURL := fmt.Sprintf("%s/download/cli-v%s/checksums.txt", base, latest)
	if err := verifyChecksum(data, archive, sumsURL); err != nil {
		return nil, "", err
	}

	binary, err := extractBinary(data, archive)
	if err != nil {
		return nil, "", err
	}
	return binary, latest, nil
}

// assetName builds the release asset filename for the given platform, matching
// the naming the install scripts and goreleaser use: Windows ships a .zip,
// every other OS a .tar.gz. goos/goarch are runtime.GOOS/GOARCH values, which
// map directly onto the published asset names (windows/darwin/linux,
// amd64/arm64).
func assetName(version, goos, goarch string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("mobiai-%s-%s-%s.%s", version, goos, goarch, ext)
}

// downloadBytes fetches url and returns the full body. Uses a generous timeout
// because release archives are a few MB. Redirects (GitHub → object storage)
// are followed by the default client.
func downloadBytes(url string) ([]byte, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "mobiai-cli")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d en %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// verifyChecksum downloads the goreleaser-style checksums.txt (lines of
// "<sha256hex>  <filename>") from sumsURL and confirms data's SHA256 matches
// the entry for filename. A mismatch — or a missing entry — is a hard error:
// we never replace the binary with unverified bytes.
func verifyChecksum(data []byte, filename, sumsURL string) error {
	sums, err := downloadBytes(sumsURL)
	if err != nil {
		return fmt.Errorf("descargar checksums.txt: %w", err)
	}
	var want string
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("no encontré checksum para %s en checksums.txt", filename)
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum no coincide para %s (esperado %s, obtenido %s) — abortando para no instalar un binario corrupto", filename, want, got)
	}
	return nil
}

// extractBinary pulls the mobiai executable out of a release archive in
// memory. Windows assets are .zip (entry "mobiai.exe"); everything else is
// .tar.gz (entry "mobiai"). Matching is by base name so a leading directory
// in the archive doesn't trip us up.
func extractBinary(archive []byte, archiveName string) ([]byte, error) {
	if strings.HasSuffix(archiveName, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, fmt.Errorf("abrir zip: %w", err)
		}
		for _, f := range zr.File {
			if isBinaryEntry(f.Name) {
				rc, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("leer %s del zip: %w", f.Name, err)
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("no encontré mobiai.exe dentro de %s", archiveName)
	}

	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("abrir gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("leer tar: %w", err)
		}
		if isBinaryEntry(hdr.Name) {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("no encontré mobiai dentro de %s", archiveName)
}

// isBinaryEntry reports whether an archive entry is the mobiai binary.
// Archives use forward slashes regardless of OS, so path.Base (not
// filepath.Base) is correct here.
func isBinaryEntry(name string) bool {
	base := path.Base(name)
	return base == "mobiai" || base == "mobiai.exe"
}

// selfReplace overwrites the currently running executable with newBin. It
// resolves the real path (following any symlink) and delegates the swap to
// applyTo. This is the only step that touches the live binary, so it stays a
// thin wrapper: applyTo holds the testable logic.
func selfReplace(newBin []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolver ejecutable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return applyTo(exe, newBin)
}

// applyTo replaces the file at targetPath with newBin using the
// rename-then-write dance that works even when targetPath is a running
// executable on Windows:
//
//  1. write newBin to "<target>.new" in the SAME directory (same volume, so
//     the rename is atomic / can't fail with a cross-device error);
//  2. remove any stale "<target>.old" (idempotent — this is what cleans up a
//     previous update, not a next-boot sweep);
//  3. rename the current target to "<target>.old" — Windows allows renaming a
//     locked .exe even though it forbids overwriting it;
//  4. rename "<target>.new" into place.
//
// On Windows the final os.Remove(old) fails while the old process is still
// running; that's expected and harmless — the next update's step 2 removes it.
func applyTo(targetPath string, newBin []byte) error {
	tmp := targetPath + ".new"
	if err := os.WriteFile(tmp, newBin, 0o755); err != nil {
		return fmt.Errorf("escribir %s: %w", tmp, err)
	}

	old := targetPath + ".old"
	_ = os.Remove(old) // clean a leftover from a prior update; ignore if absent

	if err := os.Rename(targetPath, old); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("apartar binario actual: %w", err)
	}
	if err := os.Rename(tmp, targetPath); err != nil {
		_ = os.Rename(old, targetPath) // roll back
		_ = os.Remove(tmp)
		return fmt.Errorf("mover binario nuevo a su lugar: %w", err)
	}

	_ = os.Remove(old) // best-effort; on Windows the running .old stays locked
	return nil
}
