package brain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// A note can say it holds only up to a dependency version:
// `until: com.google.firebase:firebase-bom@34.0.0` means "valid while the
// project uses a version lower than 34.0.0". Several conditions are
// comma-separated.

// untilCond is one `<dependency>@<version>` condition.
type untilCond struct {
	raw, dep, version string
}

// parseUntil splits an `until:` value into conditions and the parts that
// are not `<dependency>@<version>`. The last @ separates the version, so
// npm scopes (@scope/pkg@1.0) work.
func parseUntil(value string) (conds []untilCond, bad []string) {
	for _, raw := range splitList(value) {
		i := strings.LastIndex(raw, "@")
		if i <= 0 || i == len(raw)-1 {
			bad = append(bad, raw)
			continue
		}
		conds = append(conds, untilCond{raw: raw, dep: strings.TrimSpace(raw[:i]), version: strings.TrimSpace(raw[i+1:])})
	}
	return conds, bad
}

var versionRe = regexp.MustCompile(`^[vV]?(\d+(?:\.\d+)*)(.*)$`)

// parseVersion reads the numeric segments of a version and any pre-release
// or build suffix (-beta01, +meta). Anything else (1.+, ranges, variables,
// latest.release) can't be interpreted.
func parseVersion(s string) (nums []int, suffix string, ok bool) {
	m := versionRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return nil, "", false
	}
	if m[2] != "" && m[2][0] != '-' && m[2][0] != '+' {
		return nil, "", false
	}
	for _, part := range strings.Split(m[1], ".") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, "", false
		}
		nums = append(nums, n)
	}
	return nums, m[2], true
}

// compareVersions compares two versions by numeric segments (34.0.0 >
// 33.9.1, 1.0 == 1.0.0). ok is false when either can't be interpreted, or
// when they only differ by a pre-release suffix: then nothing is flagged.
func compareVersions(a, b string) (int, bool) {
	na, sa, okA := parseVersion(a)
	nb, sb, okB := parseVersion(b)
	if !okA || !okB {
		return 0, false
	}
	for i := 0; i < len(na) || i < len(nb); i++ {
		x, y := 0, 0
		if i < len(na) {
			x = na[i]
		}
		if i < len(nb) {
			y = nb[i]
		}
		if x != y {
			if x > y {
				return 1, true
			}
			return -1, true
		}
	}
	if sa != sb {
		return 0, false
	}
	return 0, true
}

// shortDep is how a dependency is named in warnings: the artifact of
// Gradle coordinates (firebase-bom), the name as written otherwise.
func shortDep(dep string) string {
	if i := strings.LastIndex(dep, ":"); i >= 0 {
		return dep[i+1:]
	}
	return dep
}

// checkUntil adds a warning to every note whose `until` condition is met
// (the project now uses that version or a higher one) or whose dependency
// the project no longer declares. A version that can't be interpreted
// never flags anything; `doctor` reports those (UntilProblems). The
// project's dependency files are read only when some note uses `until`.
func checkUntil(root string, notes []MemoryEntry) {
	var versions projectVersions
	for i := range notes {
		n := &notes[i]
		if strings.EqualFold(strings.TrimSpace(n.Get("status")), StatusDeprecated) {
			continue
		}
		conds, _ := parseUntil(n.Get("until"))
		if len(conds) == 0 {
			continue
		}
		if versions == nil {
			versions = readProjectVersions(root)
		}
		for _, c := range conds {
			current, found := versions.highest(c.dep)
			if !found {
				n.Flags.Warnings = append(n.Flags.Warnings,
					fmt.Sprintf("⚠ %s: dependency no longer found (note valid until %s)", shortDep(c.dep), c.version))
				continue
			}
			if cmp, ok := compareVersions(current, c.version); ok && cmp >= 0 {
				n.Flags.Warnings = append(n.Flags.Warnings,
					fmt.Sprintf("⚠ %s is now %s (note valid until %s): check whether this workaround is still needed", shortDep(c.dep), current, c.version))
			}
		}
	}
}

// validateUntil checks the `until` conditions given to `save`.
func validateUntil(values []string) error {
	for _, raw := range values {
		conds, bad := parseUntil(raw)
		if len(bad) > 0 {
			return fmt.Errorf("--until %q must be <dependency>@<version>", bad[0])
		}
		for _, c := range conds {
			if _, _, ok := parseVersion(c.version); !ok {
				return fmt.Errorf("--until %q: %q is not a version that can be compared (use numbers like 34.0.0)", c.raw, c.version)
			}
		}
	}
	return nil
}

// UntilProblems lists the `until` conditions that can never flag anything
// because they can't be interpreted — for `mobiai doctor`.
func UntilProblems(p BrainPaths) ([]string, error) {
	notes, err := LoadNotes(p)
	if err != nil {
		return nil, err
	}
	var versions projectVersions
	var out []string
	for _, n := range notes {
		conds, bad := parseUntil(n.Get("until"))
		if len(conds) == 0 && len(bad) == 0 {
			continue
		}
		id := n.Get("id")
		for _, b := range bad {
			out = append(out, fmt.Sprintf("%s: %q is not <dependency>@<version>", id, b))
		}
		if versions == nil {
			versions = readProjectVersions(p.Root)
		}
		for _, c := range conds {
			if _, _, ok := parseVersion(c.version); !ok {
				out = append(out, fmt.Sprintf("%s: %q in %s is not a version", id, c.version, c.raw))
				continue
			}
			current, found := versions.highest(c.dep)
			if !found {
				continue
			}
			if _, ok := compareVersions(current, c.version); !ok {
				out = append(out, fmt.Sprintf("%s: %s is at %q in the project, which can't be compared", id, c.dep, current))
			}
		}
	}
	return out, nil
}
