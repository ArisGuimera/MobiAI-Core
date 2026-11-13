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
	// candidate and pick with shallowest so repeated scans record the
	// same file.
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
			s.DetectedFiles["podfile"] = shallowest(podfiles, nil)
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
		s.DetectedFiles["xcodeproj"] = shallowest(xcodeprojs, nil)
	}
	if len(xcworkspaces) > 0 {
		hit = true
		// Every .xcodeproj carries an internal project.xcworkspace, one
		// level below the standalone workspace CocoaPods generates next to
		// it, so depth already prefers the standalone one. On equal depth,
		// prefer a standalone workspace too.
		s.DetectedFiles["xcworkspace"] = shallowest(xcworkspaces, func(path string) bool {
			return !strings.Contains(path, ".xcodeproj/")
		})
	}
	if !hit {
		return
	}
	types[ProjectTypeIOS] = struct{}{}
	platforms[PlatformIOS] = struct{}{}
}

// shallowest returns the path closest to the repo root. Dependency
// checkouts (Carthage/Checkouts/...) and sample apps (Example/...) sit
// deeper than the app's own project but often sort first. Ties go to a
// path for which preferred reports true (nil means no preference), then
// to sorted order, so repeated scans pick the same path.
func shallowest(paths []string, preferred func(string) bool) string {
	sort.Slice(paths, func(i, j int) bool {
		a, b := paths[i], paths[j]
		if da, db := strings.Count(a, "/"), strings.Count(b, "/"); da != db {
			return da < db
		}
		if preferred != nil {
			if pa, pb := preferred(a), preferred(b); pa != pb {
				return pa
			}
		}
		return a < b
	})
	return paths[0]
}
