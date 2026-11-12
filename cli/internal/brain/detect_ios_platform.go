package brain

import (
	"sort"
	"strings"
)

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
	// allPaths is a map, so its iteration order is random. Collect every
	// candidate and pick the first in sorted order so repeated scans
	// record the same file.
	var podfiles, xcodeprojs, xcworkspaces []string
	for path, isDir := range idx.allPaths {
		if strings.HasSuffix(path, "/Podfile") {
			podfiles = append(podfiles, path)
		}
		if isDir && strings.HasSuffix(path, ".xcodeproj") {
			xcodeprojs = append(xcodeprojs, path)
		}
		if isDir && strings.HasSuffix(path, ".xcworkspace") {
			xcworkspaces = append(xcworkspaces, path)
		}
	}
	if len(podfiles) > 0 {
		hit = true
		if _, set := s.DetectedFiles["podfile"]; !set {
			s.DetectedFiles["podfile"] = firstSorted(podfiles)
		}
		s.BuildSystems = append(s.BuildSystems, "cocoapods")
	}
	if _, ok := idx.files["Package.swift"]; ok {
		hit = true
		s.DetectedFiles["package_swift"] = "Package.swift"
		s.BuildSystems = append(s.BuildSystems, "spm")
	}
	if len(xcodeprojs) > 0 {
		hit = true
		s.DetectedFiles["xcodeproj"] = firstSorted(xcodeprojs)
	}
	if len(xcworkspaces) > 0 {
		hit = true
		// Every .xcodeproj carries an internal project.xcworkspace. Prefer
		// a standalone workspace (the one CocoaPods generates and users
		// open), otherwise the first one in sorted order.
		sort.Strings(xcworkspaces)
		chosen := xcworkspaces[0]
		for _, path := range xcworkspaces {
			if !strings.Contains(path, ".xcodeproj/") {
				chosen = path
				break
			}
		}
		s.DetectedFiles["xcworkspace"] = chosen
	}
	if !hit {
		return
	}
	types[ProjectTypeIOS] = struct{}{}
	platforms[PlatformIOS] = struct{}{}
}

// firstSorted returns the lexicographically smallest path in paths.
func firstSorted(paths []string) string {
	sort.Strings(paths)
	return paths[0]
}
