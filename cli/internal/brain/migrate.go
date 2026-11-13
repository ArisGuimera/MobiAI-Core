package brain

import (
	"fmt"
	"os"
	"path/filepath"
)

// MigratedMessage is printed (CLI) or prepended to tool output (MCP) after
// a migration changed files, for the user and for the agent reading it.
const MigratedMessage = "Brain migrated to the new format: %d notes. Include .mobiai/brain/ in your next commit."

// MigrationReport describes what Migrate did.
type MigrationReport struct {
	Created  int      // cards created from old entries
	Notices  []string // old memories that got the notice (or were recreated with it)
	ReadOnly bool     // the brain could not be written; notes are read from both formats
	Err      error    // the write error behind ReadOnly
}

// Changed reports whether the migration touched any file.
func (r MigrationReport) Changed() bool {
	return r.Created > 0 || len(r.Notices) > 0
}

// migrationWrite writes one migrated file. Tests replace it to simulate a
// read-only brain.
var migrationWrite = func(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return atomicWrite(path, data)
}

// Migrate brings an old (0.2) brain to the card format. It creates a card
// for every old entry that has none yet and makes sure decisions.md,
// bugfixes.md and testing.md carry the notice for 0.2.x users. The old
// entries stay where they are.
//
// The result depends only on the files' text, so it is the same on every
// machine and branch, and running it again changes nothing. It never
// commits. When the brain can't be written it reports ReadOnly instead of
// failing: LoadNotes reads both formats, so commands keep working.
func Migrate(p BrainPaths) (MigrationReport, error) {
	var rep MigrationReport
	if !p.Exists() {
		return rep, nil
	}

	existing := map[string]bool{}
	for _, mf := range MemoryFiles {
		if mf.Dir == "" {
			continue
		}
		cards, err := loadCardDir(p, mf)
		if err != nil {
			return rep, err
		}
		for _, c := range cards {
			existing[c.Get("id")] = true
		}
	}

	legacy, err := loadLegacyNotes(p)
	if err != nil {
		return rep, err
	}
	for _, n := range legacy {
		dir := sectionDir(n.Section)
		id := n.Get("id")
		if dir == "" || existing[id] {
			continue
		}
		path := filepath.Join(p.MemoriesDir, dir, id+".md")
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := migrationWrite(path, []byte(renderMigratedCard(n))); err != nil {
			rep.ReadOnly, rep.Err = true, err
			return rep, nil
		}
		existing[id] = true
		rep.Created++
	}

	for _, name := range noticeFiles {
		path := filepath.Join(p.MemoriesDir, name)
		var updated string
		data, err := os.ReadFile(path)
		switch {
		case os.IsNotExist(err):
			updated = noticeFileContent(memoryFile(name))
		case err != nil:
			return rep, fmt.Errorf("read %s: %w", name, err)
		default:
			var changed bool
			if updated, changed = insertNotice(string(data)); !changed {
				continue
			}
		}
		if err := migrationWrite(path, []byte(updated)); err != nil {
			rep.ReadOnly, rep.Err = true, err
			return rep, nil
		}
		rep.Notices = append(rep.Notices, name)
	}
	return rep, nil
}

// sectionDir returns the card directory of a canonical section, or "" when
// the section has no cards.
func sectionDir(section string) string {
	for _, mf := range MemoryFiles {
		if canonicalSection(mf.Name) == section {
			return mf.Dir
		}
	}
	return ""
}

// memoryFile returns the MemoryFile with the given basename.
func memoryFile(name string) MemoryFile {
	for _, mf := range MemoryFiles {
		if mf.Name == name {
			return mf
		}
	}
	return MemoryFile{Name: name}
}
