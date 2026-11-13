package brain

import (
	"strconv"
	"strings"
)

// memoriesPathspec is the brain's memories folder relative to the project
// root, as git pathspecs and tree paths use it.
const memoriesPathspec = ".mobiai/brain/memories/"

// confirmation answers "has this note been reviewed?" for one brain. A note
// is confirmed when its card is on the main branch, or when it is the
// reader's own: not committed yet, or added by a commit whose author email
// is the reader's git user.email. A migrated note is confirmed when its id
// was already in the main branch's old files. Everything else (notes by
// other people that haven't gone through a PR into the main branch) is
// unconfirmed.
type confirmation struct {
	branch       string            // "" = no git: everything is confirmed
	me           string            // git config user.email
	onBranch     map[string]bool   // card paths on the main branch
	inHead       map[string]bool   // card paths committed in HEAD
	addedBy      map[string]string // card path → author email of the commit that added it (branch..HEAD)
	legacyOnMain map[string]bool   // ids in the main branch's old single-file memories
}

// loadConfirmation gathers everything with a handful of git calls,
// whatever the number of notes.
func loadConfirmation(p BrainPaths, cfg Config) confirmation {
	c := confirmation{branch: detectConfirmedBranch(p.Root, cfg)}
	if c.branch == "" {
		return c
	}
	c.onBranch = lsTree(p.Root, c.branch)
	c.inHead = lsTree(p.Root, "HEAD")
	c.addedBy = addedBy(p.Root, c.branch)
	if out, err := runGit(p.Root, "", "config", "user.email"); err == nil {
		c.me = strings.TrimSpace(out)
	}
	c.legacyOnMain = legacyIDsAt(p.Root, c.branch)
	return c
}

// confirmed reports whether n counts as reviewed for this reader.
func (c confirmation) confirmed(n MemoryEntry) bool {
	if c.branch == "" {
		return true
	}
	if c.legacyOnMain[n.Get("id")] {
		return true
	}
	if n.Legacy {
		return false
	}
	if c.onBranch[n.Path] || !c.inHead[n.Path] {
		return true
	}
	author := c.addedBy[n.Path]
	return author != "" && c.me != "" && strings.EqualFold(author, c.me)
}

// lsTree returns the memories paths present in rev.
func lsTree(root, rev string) map[string]bool {
	out, err := runGit(root, "", "ls-tree", "-r", "--name-only", rev, "--", memoriesPathspec)
	set := map[string]bool{}
	if err != nil {
		return set
	}
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			set[line] = true
		}
	}
	return set
}

// addedBy maps each card added between branch and HEAD to the author
// email of the most recent commit that added it.
func addedBy(root, branch string) map[string]string {
	out, err := runGit(root, "", "log", "--diff-filter=A", "--relative", "--format=%x00%ae", "--name-only",
		branch+"..HEAD", "--", memoriesPathspec)
	m := map[string]string{}
	if err != nil {
		return m
	}
	email := ""
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "\x00") {
			email = strings.TrimPrefix(line, "\x00")
			continue
		}
		if line = strings.TrimSpace(line); line != "" {
			if _, seen := m[line]; !seen {
				m[line] = email
			}
		}
	}
	return m
}

// legacyIDsAt returns the ids of the notes in rev's old single-file
// memories, read in one `git cat-file --batch` call.
func legacyIDsAt(root, rev string) map[string]bool {
	var req strings.Builder
	for _, mf := range MemoryFiles {
		req.WriteString(rev + ":./" + memoriesPathspec + mf.Name + "\n")
	}
	out, err := runGit(root, req.String(), "cat-file", "--batch")
	ids := map[string]bool{}
	if err != nil {
		return ids
	}
	for out != "" {
		header, rest, ok := strings.Cut(out, "\n")
		if !ok {
			break
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[1] != "blob" {
			out = rest
			continue
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil || size > len(rest) {
			break
		}
		for _, e := range parseLegacyContent(rest[:size]) {
			id := e.Get("id")
			if id == "" {
				id = contentID(e)
			}
			ids[id] = true
		}
		out = strings.TrimPrefix(rest[size:], "\n")
	}
	return ids
}
