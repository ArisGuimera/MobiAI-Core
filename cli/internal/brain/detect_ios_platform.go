package brain

import "strings"

// detectIOS registers iOS signals. Triggered by any of:
//   - Podfile
//   - Package.swift
//   - *.xcodeproj or *.xcworkspace directory
//
// Info.plist on its own is intentionally NOT a trigger: it appears in
// many non-iOS repos (vendored frameworks, code-signing artefacts) and
// would cause false positives.
func detectIOS(idx *scanIndex, s *Scan, types, platforms map[string]struct{}) {
	hit := false

	if _, ok := idx.files["Podfile"]; ok {
		hit = true
		s.DetectedFiles["podfile"] = "Podfile"
		s.BuildSystems = append(s.BuildSystems, "cocoapods")
	}
	for path := range idx.allPaths {
		if strings.HasSuffix(path, "/Podfile") {
			hit = true
			if _, set := s.DetectedFiles["podfile"]; !set {
				s.DetectedFiles["podfile"] = path
			}
			s.BuildSystems = append(s.BuildSystems, "cocoapods")
			break
		}
	}
	if _, ok := idx.files["Package.swift"]; ok {
		hit = true
		s.DetectedFiles["package_swift"] = "Package.swift"
		s.BuildSystems = append(s.BuildSystems, "spm")
	}
	for path, isDir := range idx.allPaths {
		if !isDir {
			continue
		}
		if strings.HasSuffix(path, ".xcodeproj") {
			hit = true
			s.DetectedFiles["xcodeproj"] = path
		}
		if strings.HasSuffix(path, ".xcworkspace") {
			hit = true
			s.DetectedFiles["xcworkspace"] = path
		}
	}
	if !hit {
		return
	}
	types[ProjectTypeIOS] = struct{}{}
	platforms[PlatformIOS] = struct{}{}
}
