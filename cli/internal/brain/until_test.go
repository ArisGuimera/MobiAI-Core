package brain

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
		ok   bool
	}{
		{"34.0.0", "33.9.1", 1, true},
		{"1.10", "1.9", 1, true},
		{"1.0", "1.0.0", 0, true},
		{"v2.1.0", "2.1", 0, true},
		{"2.0.0-beta01", "1.9.0", 1, true},
		{"2.0.0-beta01", "2.0.0", 0, false},
		{"1.+", "1.0", 0, false},
		{"latest.release", "1.0", 0, false},
		{"[1.0,2.0)", "1.0", 0, false},
	}
	for _, c := range cases {
		got, ok := compareVersions(c.a, c.b)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("compareVersions(%q, %q) = %d, %v; want %d, %v", c.a, c.b, got, ok, c.want, c.ok)
		}
	}
}

func TestReadVersions_GradleCatalog(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "gradle", "libs.versions.toml"), `[versions]
firebaseBom = "34.1.0"
koin = { strictly = "3.5.6" }

[libraries]
firebase-bom = { group = "com.google.firebase", name = "firebase-bom", version.ref = "firebaseBom" }
koin-core = { module = "io.insert-koin:koin-core", version.ref = "koin" }
okhttp = "com.squareup.okhttp3:okhttp:4.12.0"
retrofit = { module = "com.squareup.retrofit2:retrofit", version = "2.11.0" }

[plugins]
android-application = { id = "com.android.application", version = "8.7.2" }
`)
	v := readProjectVersions(root)
	for name, want := range map[string]string{
		"com.google.firebase:firebase-bom": "34.1.0",
		"firebase-bom":                     "34.1.0",
		"firebase.bom":                     "34.1.0",
		"io.insert-koin:koin-core":         "3.5.6",
		"com.squareup.okhttp3:okhttp":      "4.12.0",
		"com.squareup.retrofit2:retrofit":  "2.11.0",
		"com.android.application":          "8.7.2",
	} {
		if got, ok := v.highest(name); !ok || got != want {
			t.Errorf("%s = %q (found %v), want %q", name, got, ok, want)
		}
	}
}

func TestReadVersions_GradleBuildFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "app", "build.gradle.kts"), `dependencies {
    implementation(platform("com.google.firebase:firebase-bom:33.9.0"))
    implementation("com.squareup.okhttp3:okhttp:4.11.0")
}
`)
	mustWrite(t, filepath.Join(root, "lib", "build.gradle"), `dependencies {
    implementation 'com.squareup.okhttp3:okhttp:4.12.0'
}
`)
	mustWrite(t, filepath.Join(root, "build", "ignored", "build.gradle"), `implementation 'com.squareup.okhttp3:okhttp:9.9.9'`)
	v := readProjectVersions(root)
	if got, _ := v.highest("com.google.firebase:firebase-bom"); got != "33.9.0" {
		t.Errorf("firebase-bom = %q", got)
	}
	if got, _ := v.highest("com.squareup.okhttp3:okhttp"); got != "4.12.0" {
		t.Errorf("okhttp = %q (the highest of the modules, build/ ignored)", got)
	}
}

func TestReadVersions_PodfileLock(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "ios", "Podfile.lock"), `PODS:
  - Alamofire (5.8.1)
  - Firebase/Auth (10.18.0):
    - Firebase/CoreOnly
  - Firebase/CoreOnly (10.18.0):
    - FirebaseCore (= 10.18.0)

DEPENDENCIES:
  - Alamofire (~> 5.8)
`)
	v := readProjectVersions(root)
	for name, want := range map[string]string{"Alamofire": "5.8.1", "Firebase/Auth": "10.18.0", "Firebase": "10.18.0"} {
		if got, _ := v.highest(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestReadVersions_PackageResolved(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "App.xcworkspace", "xcshareddata", "swiftpm", "Package.resolved"), `{
  "pins" : [
    { "identity" : "firebase-ios-sdk", "kind" : "remoteSourceControl", "location" : "https://github.com/firebase/firebase-ios-sdk", "state" : { "revision" : "abc", "version" : "11.4.0" } }
  ],
  "version" : 2
}`)
	mustWrite(t, filepath.Join(root, "Old", "Package.resolved"), `{
  "object": { "pins": [ { "package": "Alamofire", "repositoryURL": "https://github.com/Alamofire/Alamofire.git", "state": { "branch": null, "revision": "x", "version": "5.9.0" } } ] },
  "version": 1
}`)
	v := readProjectVersions(root)
	if got, _ := v.highest("firebase-ios-sdk"); got != "11.4.0" {
		t.Errorf("SPM v2 = %q", got)
	}
	if got, _ := v.highest("alamofire"); got != "5.9.0" {
		t.Errorf("SPM v1 (case-insensitive) = %q", got)
	}
}

func TestReadVersions_PubspecLock(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "pubspec.lock"), `packages:
  http:
    dependency: "direct main"
    description:
      name: http
      url: "https://pub.dev"
    source: hosted
    version: "1.2.2"
  firebase_core:
    dependency: "direct main"
    source: hosted
    version: "3.6.0"
sdks:
  dart: ">=3.4.0 <4.0.0"
`)
	v := readProjectVersions(root)
	for name, want := range map[string]string{"http": "1.2.2", "firebase_core": "3.6.0"} {
		if got, _ := v.highest(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if _, ok := v.highest("dart"); ok {
		t.Errorf("sdks are not packages")
	}
}

func TestReadVersions_Npm(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "package-lock.json"), `{
  "lockfileVersion": 3,
  "packages": {
    "": { "dependencies": { "react-native": "0.76.1" } },
    "node_modules/react-native": { "version": "0.76.1" },
    "node_modules/@react-native-firebase/app": { "version": "21.2.0" }
  }
}`)
	mustWrite(t, filepath.Join(root, "legacy", "package-lock.json"), `{
  "lockfileVersion": 1,
  "dependencies": { "lodash": { "version": "4.17.21" } }
}`)
	mustWrite(t, filepath.Join(root, "nolock", "package.json"), `{
  "dependencies": { "axios": "^1.7.7" },
  "devDependencies": { "jest": "~29.7.0" }
}`)
	mustWrite(t, filepath.Join(root, "node_modules", "x", "package.json"), `{"dependencies": {"axios": "9.9.9"}}`)
	v := readProjectVersions(root)
	for name, want := range map[string]string{
		"react-native": "0.76.1", "@react-native-firebase/app": "21.2.0", "lodash": "4.17.21", "axios": "1.7.7", "jest": "29.7.0",
	} {
		if got, _ := v.highest(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestParseUntil(t *testing.T) {
	conds, bad := parseUntil("com.google.firebase:firebase-bom@34.0.0, Alamofire@6.0, broken, x@")
	if len(conds) != 2 || conds[0].dep != "com.google.firebase:firebase-bom" || conds[0].version != "34.0.0" || conds[1].dep != "Alamofire" {
		t.Errorf("conds = %+v", conds)
	}
	if strings.Join(bad, ",") != "broken,x@" {
		t.Errorf("bad = %v", bad)
	}
}

func untilCard(id, until string) string {
	return "## Firebase workaround\n\n- id: " + id + "\n- type: platform_workaround\n- status: temporary\n- date: 2026-11-01\n- verified: 2026-11-01\n- until: " + until + "\n\nWorkaround.\n"
}

func TestUntil_TriggeredNotYetMissingUnparseable(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.Root, "gradle", "libs.versions.toml"), "[versions]\nbom = \"34.1.0\"\nok = \"1.+\"\n\n[libraries]\nfirebase-bom = { module = \"com.google.firebase:firebase-bom\", version.ref = \"bom\" }\nweird = { module = \"com.example:weird\", version.ref = \"ok\" }\n")
	writeCard(t, p, "bugfixes", "trig-20261101-000000", untilCard("trig-20261101-000000", "com.google.firebase:firebase-bom@34.0.0"))
	writeCard(t, p, "bugfixes", "notyet-20261101-000000", untilCard("notyet-20261101-000000", "com.google.firebase:firebase-bom@35.0.0"))
	writeCard(t, p, "bugfixes", "gone-20261101-000000", untilCard("gone-20261101-000000", "com.example:removed@2.0.0"))
	writeCard(t, p, "bugfixes", "unparse-20261101-000000", untilCard("unparse-20261101-000000", "com.example:weird@2.0.0"))

	w := warningsOf(t, p, indexNow)
	if got := w["trig-20261101-000000"]; len(got) != 1 || got[0] != "⚠ firebase-bom is now 34.1.0 (note valid until 34.0.0): check whether this workaround is still needed" {
		t.Errorf("triggered = %v", got)
	}
	if got := w["notyet-20261101-000000"]; len(got) != 0 {
		t.Errorf("not yet = %v", got)
	}
	if got := w["gone-20261101-000000"]; len(got) != 1 || got[0] != "⚠ removed: dependency no longer found (note valid until 2.0.0)" {
		t.Errorf("missing = %v", got)
	}
	if got := w["unparse-20261101-000000"]; len(got) != 0 {
		t.Errorf("an unparseable project version never flags, got %v", got)
	}
}

func TestUntilProblems_ForDoctor(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.Root, "gradle", "libs.versions.toml"), "[libraries]\nweird = \"com.example:weird:1.+\"\n")
	writeCard(t, p, "bugfixes", "a-20261101-000000", untilCard("a-20261101-000000", "com.example:weird@2.0.0, nonsense"))
	writeCard(t, p, "bugfixes", "b-20261101-000000", untilCard("b-20261101-000000", "com.example:other@not-a-version"))
	problems, err := UntilProblems(p)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(problems, "\n")
	for _, want := range []string{
		`a-20261101-000000: "nonsense" is not <dependency>@<version>`,
		`a-20261101-000000: com.example:weird is at "1.+" in the project, which can't be compared`,
		`b-20261101-000000: "not-a-version" in com.example:other@not-a-version is not a version`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
}

func TestSaveUntil_ValidatesSyntax(t *testing.T) {
	p := initBrainForTest(t)
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "X", Until: []string{"no-at-sign"}}); err == nil {
		t.Error("an until without @ must be rejected")
	}
	if _, err := AppendEntry(p, &SaveEntry{Type: SaveTypeBugfix, Title: "Y", Until: []string{"lib@latest"}}); err == nil {
		t.Error("an until whose version can't be compared must be rejected")
	}
	id := save(t, p, &SaveEntry{Type: SaveTypeBugfix, Title: "Z", Until: []string{"com.google.firebase:firebase-bom@34.0.0", "Alamofire@6.0"}})
	got, _ := readSectionCards(p.MemoriesDir, "bugfixes")
	if !strings.Contains(string(got), "- until: com.google.firebase:firebase-bom@34.0.0, Alamofire@6.0\n") {
		t.Errorf("card %s:\n%s", id, got)
	}
}

func TestReview_ListsTriggeredUntil(t *testing.T) {
	p := initBrainForTest(t)
	mustWrite(t, filepath.Join(p.Root, "pubspec.lock"), "packages:\n  http:\n    version: \"1.2.2\"\n")
	writeCard(t, p, "bugfixes", "http-20261101-000000", untilCard("http-20261101-000000", "http@1.0.0"))
	stale, err := StaleNotes(p, DefaultConfig(p.Root), indexNow)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 1 || !strings.Contains(stale[0].Flags.Warnings[0], "http is now 1.2.2") {
		t.Errorf("review must list the triggered until: %+v", stale)
	}
}
