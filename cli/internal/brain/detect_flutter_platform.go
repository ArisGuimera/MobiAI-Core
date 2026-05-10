package brain

import "strings"

// detectFlutter registers Flutter signals. Triggered by pubspec.yaml
// containing a `flutter:` key (Dart packages without Flutter use the same
// file but lack that section).
func detectFlutter(idx *scanIndex, s *Scan, types, platforms map[string]struct{}) {
	pubspec, ok := idx.files["pubspec.yaml"]
	if !ok {
		return
	}
	content := string(pubspec)
	if !strings.Contains(content, "\nflutter:") && !strings.HasPrefix(content, "flutter:") {
		// pure Dart package — not Flutter.
		return
	}
	types[ProjectTypeFlutter] = struct{}{}
	s.DetectedFiles["pubspec_yaml"] = "pubspec.yaml"

	if idx.hasDir("android") {
		platforms[PlatformAndroid] = struct{}{}
	}
	if idx.hasDir("ios") {
		platforms[PlatformIOS] = struct{}{}
	}
	platforms[PlatformShared] = struct{}{}
}
