package brain

import "strings"

// detectKMP registers Kotlin Multiplatform signals. Triggered when any
// Gradle file declares the multiplatform plugin OR a commonMain source
// set is present. KMP usually implies both Android and iOS targets, but
// we only add platforms we can directly observe (so single-target KMP
// libs scan cleanly too).
func detectKMP(idx *scanIndex, s *Scan, types, platforms map[string]struct{}) {
	pluginNeedles := []string{
		"kotlin(\"multiplatform\")",
		"kotlin('multiplatform')",
		"org.jetbrains.kotlin.multiplatform",
	}
	hasPlugin := false
	for _, n := range pluginNeedles {
		if idx.containsAny(n) {
			hasPlugin = true
			break
		}
	}
	hasCommonMain := false
	for path, isDir := range idx.allPaths {
		if !isDir {
			continue
		}
		if strings.HasSuffix(path, "/commonMain") || path == "commonMain" {
			hasCommonMain = true
			break
		}
	}
	if !hasPlugin && !hasCommonMain {
		return
	}

	types[ProjectTypeKMP] = struct{}{}

	// Walk the dir set once to attribute Android/iOS targets.
	for path, isDir := range idx.allPaths {
		if !isDir {
			continue
		}
		if strings.HasSuffix(path, "/androidMain") || path == "androidMain" {
			platforms[PlatformAndroid] = struct{}{}
		}
		if strings.HasSuffix(path, "/iosMain") || path == "iosMain" ||
			strings.HasSuffix(path, "/iosArm64Main") || strings.HasSuffix(path, "/iosX64Main") ||
			strings.HasSuffix(path, "/iosSimulatorArm64Main") {
			platforms[PlatformIOS] = struct{}{}
		}
	}
	platforms[PlatformShared] = struct{}{}
	s.BuildSystems = append(s.BuildSystems, "gradle")
}
