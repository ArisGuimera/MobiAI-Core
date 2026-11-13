package brain

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// verifyLogFormat marks each commit in the verification `git log` with its
// committer date (YYYY-MM-DD), the day the change landed.
const verifyLogFormat = "--format=%x00%cs"

// maxPathspecBytes keeps one `git log` command line well under the
// Windows limit (32 KiB); a brain with more files uses more calls.
const maxPathspecBytes = 20000

// checkVerification adds the "possibly out of date" warnings: a note with
// `### Files` is flagged when one of them changed (a commit on a later day
// than `verified`) or no longer exists; a note without files when it was
// last verified more than 3 months ago. Deprecated notes are not checked.
// Nothing is ever deleted or expired: the agent re-checks a flagged note
// when it is about to rely on it.
func checkVerification(root string, notes []MemoryEntry, now time.Time) {
	threeMonthsAgo := truncateToDay(now).AddDate(0, -3, 0).Format("2006-01-02")
	var withFiles []int
	files := map[string]bool{}
	oldest := ""
	for i := range notes {
		n := &notes[i]
		if strings.EqualFold(strings.TrimSpace(n.Get("status")), StatusDeprecated) {
			continue
		}
		ver := verifiedDay(*n)
		switch {
		case ver == "":
			n.Flags.Warnings = append(n.Flags.Warnings, "⚠ never verified")
		case len(n.Files) == 0:
			if ver < threeMonthsAgo {
				n.Flags.Warnings = append(n.Flags.Warnings, "⚠ not verified in 3+ months")
			}
		default:
			withFiles = append(withFiles, i)
			for _, f := range n.Files {
				files[cleanRepoPath(f)] = true
			}
			if oldest == "" || ver < oldest {
				oldest = ver
			}
		}
	}
	if len(withFiles) == 0 {
		return
	}

	latest := lastChangeDays(root, files, oldest)
	for _, i := range withFiles {
		n := &notes[i]
		ver := verifiedDay(*n)
		seen := map[string]bool{}
		for _, f := range n.Files {
			f = cleanRepoPath(f)
			var w string
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(f))); err != nil {
				w = "⚠ " + path.Base(f) + " no longer exists"
			} else if changedAfter(latest, f, ver) {
				w = "⚠ " + path.Base(f) + " changed since verification"
			}
			if w != "" && !seen[w] {
				seen[w] = true
				n.Flags.Warnings = append(n.Flags.Warnings, w)
			}
		}
	}
}

// verifiedDay is the note's `verified` date, or its `date` for notes that
// predate the field, as YYYY-MM-DD ("" when neither parses).
func verifiedDay(n MemoryEntry) string {
	for _, k := range []string{"verified", "date"} {
		if d, err := time.Parse("2006-01-02", strings.TrimSpace(n.Get(k))); err == nil {
			return d.Format("2006-01-02")
		}
	}
	return ""
}

// cleanRepoPath normalizes a `### Files` entry to a slash path relative
// to the project root.
func cleanRepoPath(f string) string {
	return path.Clean(strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(f)), "./"))
}

// changedAfter reports whether f (a file, or a folder: any file under it)
// has a commit on a day after ver.
func changedAfter(latest map[string]string, f, ver string) bool {
	if day, ok := latest[f]; ok {
		return day > ver
	}
	prefix := strings.TrimSuffix(f, "/") + "/"
	for p, day := range latest {
		if strings.HasPrefix(p, prefix) && day > ver {
			return true
		}
	}
	return false
}

// lastChangeDays returns, for every file touched since `since`, the day of
// its most recent commit — with one `git log` for the whole brain (more
// only when the file list would not fit in one command line). Without
// git, it returns nothing and only the existence checks apply.
func lastChangeDays(root string, files map[string]bool, since string) map[string]string {
	latest := map[string]string{}
	base := []string{"log", "--relative", verifyLogFormat, "--name-only", "--since=" + since, "--"}
	var batch []string
	size := 0
	flush := func() {
		if len(batch) == 0 {
			return
		}
		out, err := runGit(root, "", append(base, batch...)...)
		batch, size = nil, 0
		if err != nil {
			return
		}
		day := ""
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimRight(line, "\r")
			if strings.HasPrefix(line, "\x00") {
				day = strings.TrimPrefix(line, "\x00")
				continue
			}
			if line = strings.TrimSpace(line); line != "" && day > latest[line] {
				latest[line] = day
			}
		}
	}
	for f := range files {
		if size+len(f)+1 > maxPathspecBytes {
			flush()
		}
		batch = append(batch, f)
		size += len(f) + 1
	}
	flush()
	return latest
}

// StaleNotes returns the notes flagged as possibly out of date (for
// `review`).
func StaleNotes(p BrainPaths, cfg Config, now time.Time) ([]MemoryEntry, error) {
	notes, err := LoadNotes(p)
	if err != nil {
		return nil, err
	}
	annotateNotes(p, cfg, notes, now)
	var out []MemoryEntry
	for _, n := range notes {
		if len(n.Flags.Warnings) > 0 {
			out = append(out, n)
		}
	}
	return out, nil
}

// VerifyResult describes a `verify` call.
type VerifyResult struct {
	Section      string
	File         string // under memories/
	Title        string
	PrevVerified string
	NewVerified  string
	Files        []string // the new Files list, when replaced
}

// VerifyEntry marks a note as checked against the code today: `verified`
// becomes now's date and, when files is not nil, the `### Files` section is
// replaced with it. Only cards can be verified.
func VerifyEntry(p BrainPaths, id string, files []string, now time.Time) (*VerifyResult, error) {
	id = strings.TrimSpace(id)
	notes, err := LoadNotes(p)
	if err != nil {
		return nil, err
	}
	for _, n := range notes {
		if n.Get("id") != id {
			continue
		}
		if n.Legacy {
			return nil, fmt.Errorf("note %q is still only in an old memory file (the brain could not be migrated), so it can't be verified", id)
		}
		fpath := filepath.Join(p.Root, filepath.FromSlash(n.Path))
		data, err := os.ReadFile(fpath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", fpath, err)
		}
		today := now.UTC().Format("2006-01-02")
		updated, res, found, err := rewriteEntryByID(normalizeNewlines(string(data)), id, n.Section, UpdateOptions{Verified: today})
		if err != nil {
			return nil, err
		}
		if !found {
			break
		}
		if files != nil {
			updated = replaceFilesSection(updated, files)
		}
		if err := atomicWrite(fpath, []byte(updated)); err != nil {
			return nil, err
		}
		return &VerifyResult{
			Section:      n.Section,
			File:         strings.TrimPrefix(n.Path, ".mobiai/brain/memories/"),
			Title:        res.Title,
			PrevVerified: res.PrevVerified,
			NewVerified:  res.NewVerified,
			Files:        files,
		}, nil
	}
	return nil, fmt.Errorf("no entry with id %q found in the brain", id)
}

// replaceFilesSection drops the card's `### Files` list and appends one
// with files at the end.
func replaceFilesSection(content string, files []string) string {
	var kept []string
	in := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			in = strings.HasPrefix(trimmed, "###") && strings.EqualFold(strings.TrimSpace(strings.TrimLeft(trimmed, "#")), "files")
			if in {
				continue
			}
		}
		if in && (trimmed == "" || strings.HasPrefix(trimmed, "- ")) {
			continue
		}
		in = false
		kept = append(kept, line)
	}
	var b strings.Builder
	b.WriteString(strings.TrimRight(strings.Join(kept, "\n"), "\n \t"))
	b.WriteString("\n\n### Files\n")
	for _, f := range files {
		if f = strings.TrimSpace(f); f != "" {
			b.WriteString("- " + f + "\n")
		}
	}
	return b.String()
}
