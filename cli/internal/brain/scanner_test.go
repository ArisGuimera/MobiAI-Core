package brain

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScan_AndroidProject(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), `rootProject.name = "app"`)
	mustWrite(t, filepath.Join(root, "app", "build.gradle.kts"),
		"plugins { id(\"com.android.application\") }\n"+
			"dependencies { implementation(\"androidx.compose.ui:ui:1.6.0\") }\n"+
			"dependencies { implementation(\"com.google.dagger.hilt.android:hilt-android:2.50\") }\n")
	mustWrite(t, filepath.Join(root, "app", "src", "main", "AndroidManifest.xml"),
		`<manifest package="com.example"/>`)

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProjectType != ProjectTypeAndroid {
		t.Errorf("ProjectType = %q, want %q", s.ProjectType, ProjectTypeAndroid)
	}
	if !contains(s.Platforms, PlatformAndroid) {
		t.Errorf("Platforms = %v, want android", s.Platforms)
	}
	if !contains(s.UI, "compose") {
		t.Errorf("UI = %v, want compose", s.UI)
	}
	if !contains(s.DI, "hilt") {
		t.Errorf("DI = %v, want hilt", s.DI)
	}
	if !contains(s.BuildSystems, "gradle") {
		t.Errorf("BuildSystems = %v, want gradle", s.BuildSystems)
	}
}

func TestScan_KMPProject_WithComposeAndKoin(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), `rootProject.name = "kmp"`)
	mustWrite(t, filepath.Join(root, "composeApp", "build.gradle.kts"),
		"plugins { kotlin(\"multiplatform\") }\n"+
			"dependencies { implementation(\"io.insert-koin:koin-core:3.5.0\") }\n"+
			"dependencies { implementation(\"org.jetbrains.compose.runtime:runtime:1.6.0\") }\n"+
			"dependencies { implementation(\"io.ktor:ktor-client-core:2.3.0\") }\n")
	mustMkdir(t, filepath.Join(root, "composeApp", "src", "commonMain"))
	mustMkdir(t, filepath.Join(root, "composeApp", "src", "androidMain"))
	mustMkdir(t, filepath.Join(root, "composeApp", "src", "iosMain"))
	mustWrite(t, filepath.Join(root, "composeApp", "src", "androidMain", "AndroidManifest.xml"),
		`<manifest/>`)

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProjectType != ProjectTypeKMP {
		t.Errorf("ProjectType = %q, want %q", s.ProjectType, ProjectTypeKMP)
	}
	for _, want := range []string{PlatformAndroid, PlatformIOS, PlatformShared} {
		if !contains(s.Platforms, want) {
			t.Errorf("Platforms = %v, missing %q", s.Platforms, want)
		}
	}
	if !contains(s.DI, "koin") {
		t.Errorf("DI = %v, want koin", s.DI)
	}
	if !contains(s.UI, "compose_multiplatform") {
		t.Errorf("UI = %v, want compose_multiplatform", s.UI)
	}
	if !contains(s.Network, "ktor") {
		t.Errorf("Network = %v, want ktor", s.Network)
	}
}

func TestScan_FlutterProject(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "pubspec.yaml"),
		"name: app\nflutter:\n  uses-material-design: true\n")
	mustMkdir(t, filepath.Join(root, "android"))
	mustMkdir(t, filepath.Join(root, "ios"))
	mustMkdir(t, filepath.Join(root, "lib"))

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProjectType != ProjectTypeFlutter {
		t.Errorf("ProjectType = %q, want %q", s.ProjectType, ProjectTypeFlutter)
	}
	for _, want := range []string{PlatformAndroid, PlatformIOS} {
		if !contains(s.Platforms, want) {
			t.Errorf("Platforms = %v, missing %q", s.Platforms, want)
		}
	}
	if s.DetectedFiles["pubspec_yaml"] != "pubspec.yaml" {
		t.Errorf("DetectedFiles[pubspec_yaml] = %q", s.DetectedFiles["pubspec_yaml"])
	}
}

func TestScan_IOSWithPodfile(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Podfile"), "platform :ios, '15.0'\n")
	mustMkdir(t, filepath.Join(root, "App.xcodeproj"))
	mustWrite(t, filepath.Join(root, "App", "View.swift"), "import SwiftUI\n")

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProjectType != ProjectTypeIOS {
		t.Errorf("ProjectType = %q, want %q", s.ProjectType, ProjectTypeIOS)
	}
	if !contains(s.Platforms, PlatformIOS) {
		t.Errorf("Platforms = %v, want ios", s.Platforms)
	}
	if !contains(s.BuildSystems, "cocoapods") {
		t.Errorf("BuildSystems = %v, want cocoapods", s.BuildSystems)
	}
}

func TestScan_ReactNativeProject(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "package.json"),
		`{"name":"app","dependencies":{"react-native":"0.74.0"}}`)
	mustWrite(t, filepath.Join(root, "metro.config.js"), "module.exports = {};\n")
	mustMkdir(t, filepath.Join(root, "android"))
	mustMkdir(t, filepath.Join(root, "ios"))

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProjectType != ProjectTypeReactNative {
		t.Errorf("ProjectType = %q, want %q", s.ProjectType, ProjectTypeReactNative)
	}
	for _, want := range []string{PlatformAndroid, PlatformIOS} {
		if !contains(s.Platforms, want) {
			t.Errorf("Platforms = %v, missing %q", s.Platforms, want)
		}
	}
	if !contains(s.BuildSystems, "npm") {
		t.Errorf("BuildSystems = %v, want npm", s.BuildSystems)
	}
}

func TestScan_DetectsGitHubActions(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), "")
	mustMkdir(t, filepath.Join(root, ".github", "workflows"))
	mustWrite(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "name: CI\n")

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(s.CICD, "github_actions") {
		t.Errorf("CICD = %v, want github_actions", s.CICD)
	}
}

func TestScan_FlagsSensitiveFilesWithoutReading(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), "")
	mustWrite(t, filepath.Join(root, ".env"), "API_KEY=should_not_be_read")

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	hit := false
	for _, w := range s.Warnings {
		if strings.Contains(w, ".env") {
			hit = true
		}
		if strings.Contains(w, "should_not_be_read") {
			t.Error("scanner leaked .env content into warnings")
		}
	}
	if !hit {
		t.Errorf("Warnings = %v, want one mentioning .env", s.Warnings)
	}
}

func TestScan_SkipsNoiseDirs(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), "")
	// Place a fake settings.gradle.kts inside a build/ dir — must be ignored.
	mustWrite(t, filepath.Join(root, "build", "settings.gradle.kts"), "")
	mustMkdir(t, filepath.Join(root, "node_modules", "react-native"))

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	for path := range s.DetectedFiles {
		if strings.Contains(s.DetectedFiles[path], "build/") ||
			strings.Contains(s.DetectedFiles[path], "node_modules/") {
			t.Errorf("DetectedFiles leaked noise dir: %q", s.DetectedFiles[path])
		}
	}
}

func TestScan_SkipsAgentWorkspaceDirs(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), "")
	// Real .env at root must be flagged.
	mustWrite(t, filepath.Join(root, ".env"), "API_KEY=x")
	// Copies inside agent workspaces (Claude Code worktrees, Cursor, etc.)
	// must NOT appear in warnings — those are full repo clones and would
	// produce duplicate signals on every run.
	mustWrite(t, filepath.Join(root, ".claude", "worktrees", "abc", ".env"), "API_KEY=x")
	mustWrite(t, filepath.Join(root, ".claude", "worktrees", "abc", "GoogleService-Info.plist"), "")
	mustWrite(t, filepath.Join(root, ".cursor", "session", "google-services.json"), "")

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	envWarnings := 0
	for _, w := range s.Warnings {
		if strings.Contains(w, ".claude/") || strings.Contains(w, ".cursor/") {
			t.Errorf("agent workspace leaked into warnings: %q", w)
		}
		if strings.Contains(w, ".env") {
			envWarnings++
		}
	}
	if envWarnings != 1 {
		t.Errorf(".env should be flagged exactly once, got %d (warnings=%v)", envWarnings, s.Warnings)
	}
}

func TestScan_NeverReadsSensitiveFileContent(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), "")
	// google-services.json (Firebase config) frequently contains an
	// `api_key` value. The scanner must NEVER load this content into
	// the index — neither in idx.files nor reachable via containsAny.
	const secret = "AIzaSy-FAKE-FIREBASE-KEY-DO-NOT-LEAK"
	mustWrite(t, filepath.Join(root, "composeApp", "src", "google-services.json"),
		`{"client":[{"api_key":[{"current_key":"`+secret+`"}]}]}`)
	mustWrite(t, filepath.Join(root, "iosApp", "GoogleService-Info.plist"),
		`<?xml version="1.0"?>
<plist><dict><key>API_KEY</key><string>`+secret+`</string></dict></plist>`)
	mustWrite(t, filepath.Join(root, "local.properties"),
		"sdk.dir=/Users/me/Android/sdk\nMAPS_API_KEY="+secret)
	mustWrite(t, filepath.Join(root, ".env"), "API_KEY="+secret)

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range s.Warnings {
		if strings.Contains(w, secret) {
			t.Errorf("warnings leaked secret: %q", w)
		}
	}
	// Round-trip the scan to JSON and verify the secret never makes it
	// into anything we'd persist on disk.
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}
	persisted, err := os.ReadFile(p.ScanFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(persisted), secret) {
		t.Errorf("scan.json persisted the secret")
	}

	// All four sensitive files should still be flagged via path-only
	// detection (no content read involved).
	wantWarnings := []string{
		".env",
		"google-services.json",
		"GoogleService-Info.plist",
	}
	for _, want := range wantWarnings {
		hit := false
		for _, w := range s.Warnings {
			if strings.Contains(w, want) {
				hit = true
				break
			}
		}
		if !hit {
			t.Errorf("expected warning mentioning %q; got: %v", want, s.Warnings)
		}
	}
}

func TestScan_ReportsUnreadableInterestingFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows ignores Unix permission bits on regular files, so
		// chmod 0o000 doesn't actually make the file unreadable.
		// The behavior under test (warning on os.ReadFile error) is
		// cross-platform; only this way of triggering the error isn't.
		t.Skip("chmod-based unreadable file does not work on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root — chmod-based unreadable test is meaningless")
	}
	root := t.TempDir()
	gradleFile := filepath.Join(root, "build.gradle.kts")
	mustWrite(t, gradleFile, `plugins { id("com.android.application") }`)
	if err := os.Chmod(gradleFile, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(gradleFile, 0o644) })

	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	hit := false
	for _, w := range s.Warnings {
		if strings.Contains(w, "could not read") && strings.Contains(w, "build.gradle.kts") {
			hit = true
			break
		}
	}
	if !hit {
		t.Errorf("expected read-error warning for build.gradle.kts; got: %v", s.Warnings)
	}
}

func TestScan_EmptyDirectoryIsUnknown(t *testing.T) {
	root := t.TempDir()
	s, err := ScanProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.ProjectType != ProjectTypeUnknown {
		t.Errorf("ProjectType = %q, want unknown", s.ProjectType)
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func TestScan_PrefersMainAndroidManifestOverOverlays(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "settings.gradle.kts"), `include(":app")`)
	mustWrite(t, filepath.Join(root, "app", "build.gradle.kts"), `plugins { id("com.android.application") }`)
	for _, v := range []string{"debug", "flavora", "flavorb", "main", "zeta"} {
		mustWrite(t, filepath.Join(root, "app", "src", v, "AndroidManifest.xml"), `<manifest/>`)
	}

	// Map iteration order is randomized per run; repeat to catch it.
	for i := 0; i < 30; i++ {
		s, err := ScanProject(root)
		if err != nil {
			t.Fatal(err)
		}
		if got := s.DetectedFiles["android_manifest"]; got != "app/src/main/AndroidManifest.xml" {
			t.Fatalf("run %d: android_manifest = %q, want app/src/main/AndroidManifest.xml", i, got)
		}
	}
}

func TestScan_AndroidManifestWithoutMainIsStable(t *testing.T) {
	root := t.TempDir()
	for _, v := range []string{"zeta", "debug", "flavora"} {
		mustWrite(t, filepath.Join(root, "app", "src", v, "AndroidManifest.xml"), `<manifest/>`)
	}

	for i := 0; i < 30; i++ {
		s, err := ScanProject(root)
		if err != nil {
			t.Fatal(err)
		}
		if got := s.DetectedFiles["android_manifest"]; got != "app/src/debug/AndroidManifest.xml" {
			t.Fatalf("run %d: android_manifest = %q, want app/src/debug/AndroidManifest.xml", i, got)
		}
	}
}

// In a multi-module repo every module has its own src/main manifest; the
// one describing the app belongs to the module applying the Android
// application plugin, which rarely sorts first (analytics/, api/ < app/).
func TestScan_PrefersApplicationModuleManifest(t *testing.T) {
	libraries := map[string]string{
		"settings.gradle.kts":                    `include(":analytics", ":api", ":app")`,
		"analytics/build.gradle.kts":             `plugins { alias(libs.plugins.android.library) }`,
		"analytics/src/main/AndroidManifest.xml": `<manifest/>`,
		"api/build.gradle":                       `apply plugin: 'com.android.library'`,
		"api/src/main/AndroidManifest.xml":       `<manifest/>`,
		"app/src/main/AndroidManifest.xml":       `<manifest/>`,
		"app/src/debug/AndroidManifest.xml":      `<manifest/>`,
	}
	cases := []struct {
		name      string
		buildFile string
		plugins   string
	}{
		{"plugin id", "build.gradle.kts", `plugins { id("com.android.application") }`},
		{"groovy apply plugin", "build.gradle", `apply plugin: 'com.android.application'`},
		{"version catalog alias", "build.gradle.kts", `plugins { alias(libs.plugins.android.application) }`},
		{"camelCase catalog alias", "build.gradle.kts", `plugins { alias(libs.plugins.androidApplication) }`},
		{"convention plugin", "build.gradle.kts", `plugins { id("miapp.android.application") }`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeTree(t, root, libraries)
			mustWrite(t, filepath.Join(root, "app", tc.buildFile), tc.plugins)
			assertStableManifest(t, root, "app/src/main/AndroidManifest.xml")
		})
	}
}

func TestScan_PrefersFirstOfSeveralApplicationModules(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"analytics/build.gradle.kts":             `plugins { id("com.android.library") }`,
		"analytics/src/main/AndroidManifest.xml": `<manifest/>`,
		"wear/build.gradle.kts":                  `plugins { id("com.android.application") }`,
		"wear/src/main/AndroidManifest.xml":      `<manifest/>`,
		"mobile/build.gradle.kts":                `plugins { id("com.android.application") }`,
		"mobile/src/main/AndroidManifest.xml":    `<manifest/>`,
	})
	assertStableManifest(t, root, "mobile/src/main/AndroidManifest.xml")
}

func TestScan_PrefersRootApplicationModuleManifest(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"build.gradle":                           `apply plugin: 'com.android.application'`,
		"src/main/AndroidManifest.xml":           `<manifest/>`,
		"analytics/build.gradle":                 `apply plugin: 'com.android.library'`,
		"analytics/src/main/AndroidManifest.xml": `<manifest/>`,
	})
	assertStableManifest(t, root, "src/main/AndroidManifest.xml")
}

func TestScan_WithoutApplicationModuleKeepsFirstMainManifest(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"analytics/build.gradle.kts":             `plugins { id("com.android.library") }`,
		"analytics/src/main/AndroidManifest.xml": `<manifest/>`,
		"app/build.gradle.kts":                   `plugins { id("com.android.library") }`,
		"app/src/main/AndroidManifest.xml":       `<manifest/>`,
		"app/src/debug/AndroidManifest.xml":      `<manifest/>`,
	})
	assertStableManifest(t, root, "analytics/src/main/AndroidManifest.xml")
}

// assertStableManifest scans root repeatedly: map iteration order is
// randomized per run, so a single scan could pick the right file by luck.
func assertStableManifest(t *testing.T, root, want string) {
	t.Helper()
	for i := 0; i < 30; i++ {
		s, err := ScanProject(root)
		if err != nil {
			t.Fatal(err)
		}
		if got := s.DetectedFiles["android_manifest"]; got != want {
			t.Fatalf("run %d: android_manifest = %q, want %q", i, got, want)
		}
	}
}

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		mustWrite(t, filepath.Join(root, filepath.FromSlash(rel)), content)
	}
}

func TestScanSave_PreservesCreatedAtOnRescan(t *testing.T) {
	root := t.TempDir()
	paths := NewBrainPaths(root)
	first := newScan("app")
	first.CreatedAt = "2026-08-07T10:00:00Z"
	if err := first.Save(paths); err != nil {
		t.Fatal(err)
	}

	// A re-scan builds a fresh Scan with empty CreatedAt.
	if err := newScan("app").Save(paths); err != nil {
		t.Fatal(err)
	}
	got, err := LoadScan(paths)
	if err != nil {
		t.Fatal(err)
	}
	if got.CreatedAt != "2026-08-07T10:00:00Z" {
		t.Errorf("CreatedAt = %q, want preserved 2026-08-07T10:00:00Z", got.CreatedAt)
	}
	if got.UpdatedAt == "" || got.UpdatedAt == got.CreatedAt {
		t.Errorf("UpdatedAt = %q, want refreshed timestamp", got.UpdatedAt)
	}
}
