package brain

import (
	"strings"
)

// formatNotice is the entry 0.3 adds to the old single-file memories.
// It is a regular old-format entry on purpose: a 0.2.x binary shows it in
// `context` and, because it is `temporary` with a past `review_after`,
// also in `review`, so an agent running an old mobiai tells the user to
// update. 0.3 never treats it as a note (see formatNoticeID).
const formatNotice = `## MobiAI Brain moved to a new format: update mobiai

- id: mobiai-brain-format-notice
- type: notice
- status: temporary
- review_after: 2000-01-01

This project's Brain now stores one note per file under ` + "`.mobiai/brain/memories/<section>/`" + `.
The mobiai version on this machine is older than 0.3.0 and only reads this file, so it
cannot see notes saved after the migration. Tell the user to update mobiai with
` + "`mobiai update`" + ` (versions older than 0.2.2 must reinstall with the install script).
Notes saved with the old version are not lost: the new version picks them up automatically.
`

// noticeFiles are the old memories that carry the notice. integrations.md
// and releases.md don't: 0.2.x never saved into them.
var noticeFiles = []string{"decisions.md", "bugfixes.md", "testing.md"}

// insertNotice returns content with the notice inserted right after the
// file's `# ` heading (or at the very top when there is none), and whether
// it changed anything. The position is fixed, so two branches that insert
// it produce the same change and git merges them cleanly; being near the
// top also keeps it away from the end, where 0.2.x appends new notes.
// A CRLF file stays CRLF.
func insertNotice(content string) (string, bool) {
	normalized := normalizeNewlines(content)
	if strings.Contains(normalized, "- id: "+formatNoticeID) {
		return content, false
	}
	crlf := strings.Contains(content, "\r\n")

	lines := strings.Split(normalized, "\n")
	at := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			break
		}
		if strings.HasPrefix(line, "# ") {
			at = i + 1
			break
		}
	}

	notice := strings.Split(strings.TrimSuffix(formatNotice, "\n"), "\n")
	out := make([]string, 0, len(lines)+len(notice)+2)
	out = append(out, lines[:at]...)
	if at > 0 {
		out = append(out, "")
	}
	out = append(out, notice...)
	rest := lines[at:]
	if len(rest) > 0 && rest[0] != "" {
		out = append(out, "")
	}
	out = append(out, rest...)

	result := strings.Join(out, "\n")
	if crlf {
		result = strings.ReplaceAll(result, "\n", "\r\n")
	}
	return result, true
}

// noticeFileContent is what a missing old memory is (re)created with:
// its heading and the notice, nothing else.
func noticeFileContent(mf MemoryFile) string {
	heading := strings.SplitN(mf.Template, "\n", 2)[0]
	return heading + "\n\n" + formatNotice
}
