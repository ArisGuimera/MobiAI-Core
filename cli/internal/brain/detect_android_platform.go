package brain

import (
	"sort"
	"strings"
)

// detectAndroid registers Android signals in s. Triggered by:
//   - any AndroidManifest.xml in the tree
//   - build.gradle{,.kts} containing the Android plugin id
func detectAndroid(idx *scanIndex, s *Scan, types, platforms map[string]struct{}) {
	// allPaths is a map, so its iteration order is random. Collect every
	// manifest, sort them and pick deterministically.
	var manifests []string
	for path, isDir := range idx.allPaths {
		if !isDir && (strings.HasSuffix(path, "/AndroidManifest.xml") || path == "AndroidManifest.xml") {
			manifests = append(manifests, path)
		}
	}
	hasManifest := len(manifests) > 0
	if hasManifest {
		sort.Strings(manifests)
		s.DetectedFiles["android_manifest"] = pickAndroidManifest(idx, manifests)
	}

	hasAndroidPlugin := false
	for _, needle := range []string{
		"com.android.application",
		"com.android.library",
		"org.jetbrains.kotlin.android",
	} {
		if idx.containsAny(needle) {
			hasAndroidPlugin = true
			break
		}
	}

	if hasManifest || hasAndroidPlugin {
		types[ProjectTypeAndroid] = struct{}{}
		platforms[PlatformAndroid] = struct{}{}
		s.BuildSystems = append(s.BuildSystems, "gradle")

		// Record the most authoritative gradle file we have.
		for _, candidate := range []string{
			"settings.gradle.kts",
			"settings.gradle",
			"build.gradle.kts",
			"build.gradle",
		} {
			if _, ok := idx.files[candidate]; ok {
				s.DetectedFiles["settings_gradle"] = candidate
				break
			}
		}
	}
}

// mainManifest is the base manifest Gradle merges flavor and build-type
// overlays into.
const mainManifest = "src/main/AndroidManifest.xml"

// pickAndroidManifest chooses from sorted manifest paths, in order of
// preference: the src/main manifest of a module applying the Android
// application plugin, then any src/main manifest, then the first one.
// In multi-module repos library modules (analytics/, api/) often sort
// before the app module.
func pickAndroidManifest(idx *scanIndex, sorted []string) string {
	firstMain := ""
	for _, path := range sorted {
		if path != mainManifest && !strings.HasSuffix(path, "/"+mainManifest) {
			continue
		}
		if isAndroidApplicationModule(idx, strings.TrimSuffix(path, mainManifest)) {
			return path
		}
		if firstMain == "" {
			firstMain = path
		}
	}
	if firstMain != "" {
		return firstMain
	}
	return sorted[0]
}

// isAndroidApplicationModule reports whether the build file of the module
// at dir ("" for the root, otherwise with a trailing slash) applies the
// Android application plugin. Matching "android.application" covers the
// plugin id (com.android.application), the version catalog alias
// (libs.plugins.android.application) and convention plugins named
// <prefix>.android.application; "androidApplication" covers camelCase
// catalog aliases. Convention plugins named otherwise go undetected.
func isAndroidApplicationModule(idx *scanIndex, dir string) bool {
	for _, name := range []string{"build.gradle.kts", "build.gradle"} {
		content := string(idx.content(dir + name))
		if strings.Contains(content, "android.application") || strings.Contains(content, "androidApplication") {
			return true
		}
	}
	return false
}
