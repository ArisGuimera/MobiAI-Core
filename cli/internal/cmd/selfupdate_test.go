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
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"windows", "amd64", "mobiai-1.2.3-windows-amd64.zip"},
		{"windows", "arm64", "mobiai-1.2.3-windows-arm64.zip"},
		{"darwin", "amd64", "mobiai-1.2.3-darwin-amd64.tar.gz"},
		{"darwin", "arm64", "mobiai-1.2.3-darwin-arm64.tar.gz"},
		{"linux", "amd64", "mobiai-1.2.3-linux-amd64.tar.gz"},
		{"linux", "arm64", "mobiai-1.2.3-linux-arm64.tar.gz"},
	}
	for _, c := range cases {
		if got := assetName("1.2.3", c.goos, c.goarch); got != c.want {
			t.Errorf("assetName(1.2.3, %s, %s) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}

// makeTarGz returns a .tar.gz containing a single entry named entryName.
func makeTarGz(t *testing.T, entryName string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	if err := tw.WriteHeader(&tar.Header{Name: entryName, Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// makeZip returns a .zip containing a single entry named entryName.
func makeZip(t *testing.T, entryName string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(entryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractBinary_TarGz(t *testing.T) {
	want := []byte("fake-mobiai-binary")
	// Entry under a leading dir to confirm base-name matching.
	archive := makeTarGz(t, "dist/mobiai", want)
	got, err := extractBinary(archive, "mobiai-1.0.0-linux-amd64.tar.gz")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractBinary_Zip(t *testing.T) {
	want := []byte("fake-mobiai-exe")
	archive := makeZip(t, "mobiai.exe", want)
	got, err := extractBinary(archive, "mobiai-1.0.0-windows-amd64.zip")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractBinary_MissingEntry(t *testing.T) {
	archive := makeTarGz(t, "README.md", []byte("not the binary"))
	if _, err := extractBinary(archive, "mobiai-1.0.0-linux-amd64.tar.gz"); err == nil {
		t.Error("expected error when binary entry is absent")
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("the-archive-bytes")
	sum := sha256.Sum256(data)
	hexsum := hex.EncodeToString(sum[:])
	filename := "mobiai-1.0.0-linux-amd64.tar.gz"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// goreleaser format: "<hash>  <filename>", plus a decoy line.
		fmt.Fprintf(w, "deadbeef  some-other-file.zip\n%s  %s\n", hexsum, filename)
	}))
	defer srv.Close()

	if err := verifyChecksum(data, filename, srv.URL); err != nil {
		t.Errorf("verifyChecksum (matching): %v", err)
	}

	if err := verifyChecksum([]byte("tampered"), filename, srv.URL); err == nil {
		t.Error("expected error on checksum mismatch")
	}

	if err := verifyChecksum(data, "not-listed.tar.gz", srv.URL); err == nil {
		t.Error("expected error when filename is absent from checksums.txt")
	}
}

func TestApplyTo(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "mobiai-bin")
	if err := os.WriteFile(target, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := applyTo(target, []byte("NEW")); err != nil {
		t.Fatalf("applyTo: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "NEW" {
		t.Errorf("target content = %q, want NEW", got)
	}
	// The temp ".new" must not linger.
	if _, err := os.Stat(target + ".new"); !os.IsNotExist(err) {
		t.Errorf("%s.new should have been renamed away", target)
	}
}

func TestApplyTo_RemovesStaleOld(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "mobiai-bin")
	if err := os.WriteFile(target, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Simulate a leftover ".old" from a previous update (the Windows case
	// where the running process couldn't be removed last time).
	if err := os.WriteFile(target+".old", []byte("STALE"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := applyTo(target, []byte("NEW")); err != nil {
		t.Fatalf("applyTo with stale .old: %v", err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "NEW" {
		t.Errorf("target content = %q, want NEW", got)
	}
}

func TestFetchUpdateBinary_DevSkips(t *testing.T) {
	var out bytes.Buffer
	bin, latest, err := fetchUpdateBinary(&out, "dev")
	if err != nil {
		t.Fatalf("fetchUpdateBinary(dev): %v", err)
	}
	if bin != nil || latest != "" {
		t.Errorf("dev build should be a no-op; got bin=%v latest=%q", bin != nil, latest)
	}
	if out.Len() != 0 {
		t.Errorf("dev build should print nothing; got %q", out.String())
	}
}

func TestFetchUpdateBinary_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"tag_name":"cli-v0.2.0","html_url":"https://example.test/r"}]`))
	}))
	defer srv.Close()
	t.Setenv("MOBIAI_UPDATE_CHECK_URL", srv.URL)

	var out bytes.Buffer
	bin, _, err := fetchUpdateBinary(&out, "0.2.0")
	if err != nil {
		t.Fatalf("fetchUpdateBinary: %v", err)
	}
	if bin != nil {
		t.Error("expected no binary when already up to date")
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Errorf("expected 'up to date' message; got %q", out.String())
	}
}

func TestFetchUpdateBinary_DownloadsAndExtracts(t *testing.T) {
	const latestVer = "0.3.0"
	binContent := []byte("the-new-mobiai-binary")
	archiveName := assetName(latestVer, runtime.GOOS, runtime.GOARCH)

	var archive []byte
	if runtime.GOOS == "windows" {
		archive = makeZip(t, "mobiai.exe", binContent)
	} else {
		archive = makeTarGz(t, "mobiai", binContent)
	}
	sum := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/releases":
			fmt.Fprintf(w, `[{"tag_name":"cli-v%s","html_url":"https://example.test/r"}]`, latestVer)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			_, _ = io.WriteString(w, checksums)
		case strings.HasSuffix(r.URL.Path, archiveName):
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("MOBIAI_UPDATE_CHECK_URL", srv.URL+"/releases")
	t.Setenv("MOBIAI_INSTALL_BASE", srv.URL)

	var out bytes.Buffer
	bin, latest, err := fetchUpdateBinary(&out, "0.2.0")
	if err != nil {
		t.Fatalf("fetchUpdateBinary: %v", err)
	}
	if latest != latestVer {
		t.Errorf("latest = %q, want %q", latest, latestVer)
	}
	if !bytes.Equal(bin, binContent) {
		t.Errorf("extracted binary = %q, want %q", bin, binContent)
	}
}

func TestFetchUpdateBinary_ChecksumMismatchAborts(t *testing.T) {
	const latestVer = "0.3.0"
	archiveName := assetName(latestVer, runtime.GOOS, runtime.GOARCH)
	var archive []byte
	if runtime.GOOS == "windows" {
		archive = makeZip(t, "mobiai.exe", []byte("real"))
	} else {
		archive = makeTarGz(t, "mobiai", []byte("real"))
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/releases":
			fmt.Fprintf(w, `[{"tag_name":"cli-v%s"}]`, latestVer)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			// Wrong hash on purpose.
			fmt.Fprintf(w, "%s  %s\n", strings.Repeat("0", 64), archiveName)
		case strings.HasSuffix(r.URL.Path, archiveName):
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("MOBIAI_UPDATE_CHECK_URL", srv.URL+"/releases")
	t.Setenv("MOBIAI_INSTALL_BASE", srv.URL)

	var out bytes.Buffer
	if _, _, err := fetchUpdateBinary(&out, "0.2.0"); err == nil {
		t.Error("expected checksum mismatch to abort the update")
	}
}
