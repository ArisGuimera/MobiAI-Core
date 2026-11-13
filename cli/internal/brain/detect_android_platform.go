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
	// manifest and pick deterministically: prefer the base src/main
	// manifest Gradle merges flavor/build-type overlays into, otherwise
	// the first one in sorted order.
	var manifests []string
	for path, isDir := range idx.allPaths {
		if !isDir && (strings.HasSuffix(path, "/AndroidManifest.xml") || path == "AndroidManifest.xml") {
			manifests = append(manifests, path)
		}
	}
	hasManifest := len(manifests) > 0
	if hasManifest {
		sort.Strings(manifests)
		chosen := manifests[0]
		for _, path := range manifests {
			if strings.HasSuffix(path, "src/main/AndroidManifest.xml") {
				chosen = path
				break
			}
		}
		s.DetectedFiles["android_manifest"] = chosen
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
