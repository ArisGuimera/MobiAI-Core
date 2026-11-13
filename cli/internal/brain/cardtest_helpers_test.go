package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// readSectionCards concatenates every card in <memoriesDir>/<dir>/ sorted
// by filename. Tests use it where they used to read the old single file.
func readSectionCards(memoriesDir, dir string) ([]byte, error) {
	matches, err := filepath.Glob(filepath.Join(memoriesDir, dir, "*.md"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no cards in %s", filepath.Join(memoriesDir, dir))
	}
	sort.Strings(matches)
	var out []byte
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}
