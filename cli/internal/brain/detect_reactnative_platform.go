package brain

import "strings"

// detectReactNative registers RN signals. Triggered by package.json with
// a "react-native" dependency (devDependencies counts too) AND either a
// metro.config.js or an android/ + ios/ pair.
func detectReactNative(idx *scanIndex, s *Scan, types, platforms map[string]struct{}) {
	pkg, ok := idx.files["package.json"]
	if !ok {
		return
	}
	if !strings.Contains(string(pkg), "\"react-native\"") {
		return
	}
	if _, ok := idx.files["metro.config.js"]; !ok && !(idx.hasDir("android") && idx.hasDir("ios")) {
		return
	}
	types[ProjectTypeReactNative] = struct{}{}
	platforms[PlatformShared] = struct{}{}
	if idx.hasDir("android") {
		platforms[PlatformAndroid] = struct{}{}
	}
	if idx.hasDir("ios") {
		platforms[PlatformIOS] = struct{}{}
	}
	s.DetectedFiles["package_json"] = "package.json"
	if _, ok := idx.files["metro.config.js"]; ok {
		s.DetectedFiles["metro_config"] = "metro.config.js"
	}
	s.BuildSystems = append(s.BuildSystems, "npm")
}
