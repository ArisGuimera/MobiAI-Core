package brain

import "strings"

// detectAndroid registers Android signals in s. Triggered by:
//   - any AndroidManifest.xml in the tree
//   - build.gradle{,.kts} containing the Android plugin id
func detectAndroid(idx *scanIndex, s *Scan, types, platforms map[string]struct{}) {
	hasManifest := false
	for path, isDir := range idx.allPaths {
		if isDir {
			continue
		}
		if strings.HasSuffix(path, "/AndroidManifest.xml") || path == "AndroidManifest.xml" {
			hasManifest = true
			s.DetectedFiles["android_manifest"] = path
			break
		}
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
