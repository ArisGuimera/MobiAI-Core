package brain

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// projectVersions maps a dependency name (see depKey) to every version
// string the project declares for it, as written.
type projectVersions map[string][]string

// depKey normalizes a dependency name so the same library matches however
// it is written: case-insensitive, and `.`/`_`/`-` are equivalent (a Gradle
// catalog alias firebase-bom is libs.firebase.bom in build files).
func depKey(name string) string {
	return strings.NewReplacer(".", "-", "_", "-").Replace(strings.ToLower(strings.TrimSpace(name)))
}

func (v projectVersions) add(name, version string) {
	name, version = strings.TrimSpace(name), strings.TrimSpace(version)
	if name == "" || version == "" {
		return
	}
	k := depKey(name)
	v[k] = append(v[k], version)
}

// highest returns the highest version the project declares for name (when
// none can be compared, the first one found) and whether the dependency
// was found at all.
func (v projectVersions) highest(name string) (string, bool) {
	all, ok := v[depKey(name)]
	if !ok || len(all) == 0 {
		return "", false
	}
	best := ""
	for _, ver := range all {
		if _, _, ok := parseVersion(ver); !ok {
			continue
		}
		if best == "" {
			best = ver
			continue
		}
		if c, ok := compareVersions(ver, best); ok && c > 0 {
			best = ver
		}
	}
	if best == "" {
		best = all[0]
	}
	return best, true
}

// versionSkipDirs are never walked: build output, caches, vendored code.
var versionSkipDirs = map[string]bool{
	".git": true, ".gradle": true, ".idea": true, ".mobiai": true, ".dart_tool": true,
	".build": true, "build": true, "node_modules": true, "Pods": true, "DerivedData": true,
}

// versionMaxDepth bounds the walk so a huge repo stays quick.
const versionMaxDepth = 8

// readProjectVersions reads the dependency versions the project declares:
// Gradle version catalogs and build files, Podfile.lock, Package.resolved,
// pubspec.lock, package-lock.json (package.json when there is no lock).
// Plain line readers and encoding/json: no new dependency in the binary.
func readProjectVersions(root string) projectVersions {
	v := projectVersions{}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root {
				rel, _ := filepath.Rel(root, path)
				if versionSkipDirs[d.Name()] || strings.Count(filepath.ToSlash(rel), "/") >= versionMaxDepth {
					return filepath.SkipDir
				}
			}
			return nil
		}
		switch name := d.Name(); {
		case strings.HasSuffix(name, ".versions.toml"):
			readGradleCatalog(path, v)
		case name == "build.gradle" || name == "build.gradle.kts":
			readGradleBuild(path, v)
		case name == "Podfile.lock":
			readPodfileLock(path, v)
		case name == "Package.resolved":
			readPackageResolved(path, v)
		case name == "pubspec.lock":
			readPubspecLock(path, v)
		case name == "package-lock.json":
			readPackageLock(path, v)
		case name == "package.json":
			if !isFile(filepath.Join(filepath.Dir(path), "package-lock.json")) {
				readPackageJSON(path, v)
			}
		}
		return nil
	})
	return v
}

// eachLine calls fn for every line of the file (CR stripped).
func eachLine(path string, fn func(string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for s.Scan() {
		fn(strings.TrimRight(s.Text(), "\r"))
	}
}

var (
	tomlTableRe   = regexp.MustCompile(`^\[([A-Za-z0-9_.-]+)\]\s*(#.*)?$`)
	tomlEntryRe   = regexp.MustCompile(`^([A-Za-z0-9_.-]+)\s*=\s*(.+)$`)
	tomlFieldRe   = regexp.MustCompile(`([A-Za-z0-9_.]+)\s*=\s*"([^"]*)"`)
	tomlStringRe  = regexp.MustCompile(`^"([^"]*)"`)
	gradleCoordRe = regexp.MustCompile(`["']([A-Za-z0-9_.\-]+):([A-Za-z0-9_.\-]+):([^"'@:\s]+)(@[A-Za-z]+)?["']`)
	podLineRe     = regexp.MustCompile(`^  - ("?)([^\s"]+)("?) \(([^)]+)\)`)
)

// readGradleCatalog reads a Gradle version catalog (libs.versions.toml):
// [versions] values, and [libraries]/[plugins] entries with an inline
// version, version.ref or "group:artifact:version" string. Each library is
// recorded under its coordinates and its alias.
func readGradleCatalog(path string, v projectVersions) {
	versions := map[string]string{}
	type entry struct{ alias, fields, inline string }
	var libs []entry
	table := ""
	eachLine(path, func(line string) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			return
		}
		if m := tomlTableRe.FindStringSubmatch(line); m != nil {
			table = m[1]
			return
		}
		m := tomlEntryRe.FindStringSubmatch(line)
		if m == nil {
			return
		}
		key, value := m[1], strings.TrimSpace(m[2])
		switch table {
		case "versions":
			if s := tomlStringRe.FindStringSubmatch(value); s != nil {
				versions[key] = s[1]
				return
			}
			fields := tomlFields(value)
			for _, k := range []string{"strictly", "require", "prefer"} {
				if fields[k] != "" {
					versions[key] = fields[k]
					return
				}
			}
		case "libraries", "plugins":
			if s := tomlStringRe.FindStringSubmatch(value); s != nil {
				libs = append(libs, entry{alias: key, inline: s[1]})
			} else {
				libs = append(libs, entry{alias: key, fields: value})
			}
		}
	})
	for _, l := range libs {
		var name, version string
		if l.inline != "" {
			parts := strings.Split(l.inline, ":")
			if len(parts) < 3 {
				continue
			}
			name, version = parts[0]+":"+parts[1], parts[2]
		} else {
			f := tomlFields(l.fields)
			switch {
			case f["module"] != "":
				name = f["module"]
			case f["group"] != "" && f["name"] != "":
				name = f["group"] + ":" + f["name"]
			default:
				name = f["id"]
			}
			version = f["version"]
			if version == "" {
				version = versions[f["version.ref"]]
			}
		}
		if version == "" {
			continue
		}
		v.add(name, version)
		v.add(l.alias, version)
	}
}

// tomlFields reads the `key = "value"` pairs of a TOML inline table.
func tomlFields(s string) map[string]string {
	out := map[string]string{}
	for _, m := range tomlFieldRe.FindAllStringSubmatch(s, -1) {
		out[m[1]] = m[2]
	}
	return out
}

// readGradleBuild reads literal "group:artifact:version" coordinates in a
// build.gradle(.kts), including platform("...") BOMs.
func readGradleBuild(path string, v projectVersions) {
	eachLine(path, func(line string) {
		for _, m := range gradleCoordRe.FindAllStringSubmatch(line, -1) {
			v.add(m[1]+":"+m[2], m[3])
		}
	})
}

// readPodfileLock reads the PODS: section of a Podfile.lock. A subspec
// (Firebase/Auth) also counts for its pod (Firebase).
func readPodfileLock(path string, v projectVersions) {
	inPods := false
	eachLine(path, func(line string) {
		if !strings.HasPrefix(line, " ") {
			inPods = strings.TrimSpace(line) == "PODS:"
			return
		}
		if !inPods {
			return
		}
		if m := podLineRe.FindStringSubmatch(line); m != nil {
			name, version := m[2], m[4]
			v.add(name, version)
			if base, _, ok := strings.Cut(name, "/"); ok {
				v.add(base, version)
			}
		}
	})
}

// readPackageResolved reads a SwiftPM Package.resolved (format 1, 2 or 3).
func readPackageResolved(path string, v projectVersions) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	type pin struct {
		Identity string `json:"identity"`
		Package  string `json:"package"`
		State    struct {
			Version string `json:"version"`
		} `json:"state"`
	}
	var doc struct {
		Pins   []pin `json:"pins"`
		Object struct {
			Pins []pin `json:"pins"`
		} `json:"object"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return
	}
	for _, p := range append(doc.Pins, doc.Object.Pins...) {
		v.add(p.Identity, p.State.Version)
		v.add(p.Package, p.State.Version)
	}
}

// readPubspecLock reads the packages: section of a pubspec.lock.
func readPubspecLock(path string, v projectVersions) {
	inPackages := false
	current := ""
	eachLine(path, func(line string) {
		if !strings.HasPrefix(line, " ") {
			inPackages = strings.TrimSpace(line) == "packages:"
			return
		}
		if !inPackages {
			return
		}
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "   ") && strings.HasSuffix(line, ":") {
			current = strings.TrimSuffix(strings.TrimSpace(line), ":")
			return
		}
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(line, "    ") && strings.HasPrefix(trimmed, "version:") {
			v.add(current, strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "version:")), `"'`))
		}
	})
}

// readPackageLock reads an npm package-lock.json (lockfile 1, 2 or 3).
func readPackageLock(path string, v projectVersions) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	type pkg struct {
		Version string `json:"version"`
	}
	var doc struct {
		Packages     map[string]pkg `json:"packages"`
		Dependencies map[string]pkg `json:"dependencies"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return
	}
	for key, p := range doc.Packages {
		if i := strings.LastIndex(key, "node_modules/"); i >= 0 {
			v.add(key[i+len("node_modules/"):], p.Version)
		}
	}
	for name, p := range doc.Dependencies {
		v.add(name, p.Version)
	}
}

// readPackageJSON reads the declared ranges of a package.json, keeping the
// version they start from (^1.7.7 → 1.7.7).
func readPackageJSON(path string, v projectVersions) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var doc map[string]json.RawMessage
	if json.Unmarshal(data, &doc) != nil {
		return
	}
	for _, field := range []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies"} {
		var deps map[string]string
		if json.Unmarshal(doc[field], &deps) != nil {
			continue
		}
		for name, rng := range deps {
			v.add(name, strings.TrimLeft(strings.TrimSpace(rng), "^~>=<v "))
		}
	}
}
