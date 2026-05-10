package brain

import (
	"path/filepath"
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
