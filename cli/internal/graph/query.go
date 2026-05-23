package graph

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SearchHit is a single hit returned by Search.
type SearchHit struct {
	Symbol   Symbol
	File     string // FileIndex.Path of the symbol's home file
	Language string // FileIndex.Language
	Score    int    // higher = better; ranking: exact match > prefix > substring
}

// Search looks up symbols by name in the index. Matching is case-insensitive
// substring; ranking is: exact case-insensitive match (score 3) > prefix
// match (score 2) > substring match (score 1). Ties broken by File then Line.
// Empty term returns nil.
func Search(idx *Index, term string) []SearchHit {
	if idx == nil || term == "" {
		return nil
	}
	needle := strings.ToLower(term)

	var hits []SearchHit
	for _, file := range idx.Files {
		for _, sym := range file.Symbols {
			name := strings.ToLower(sym.Name)
			score := 0
			switch {
			case name == needle:
				score = 3
			case strings.HasPrefix(name, needle):
				score = 2
			case strings.Contains(name, needle):
				score = 1
			}
			if score == 0 {
				continue
			}
			hits = append(hits, SearchHit{
				Symbol:   sym,
				File:     file.Path,
				Language: file.Language,
				Score:    score,
			})
		}
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if hits[i].File != hits[j].File {
			return hits[i].File < hits[j].File
		}
		return hits[i].Symbol.Line < hits[j].Symbol.Line
	})

	return hits
}

// CallerHit is a textual occurrence of a symbol name in a file other than
// (or other than the line of) its definition.
type CallerHit struct {
	File    string // FileIndex.Path
	Line    int    // 1-indexed
	Snippet string // the raw line, trimmed of trailing whitespace
}

// Callers returns textual references to a symbol name across the index's
// files. It reads each file under root (using FileIndex.Path joined to root)
// and looks for the symbol name as a whole word (\b<name>\b). Definition
// lines of the SAME name are excluded (a line whose definition Symbol.Line
// matches). Returns hits sorted by File then Line.
//
// If a file fails to read, that file is skipped silently — Callers is
// best-effort and never returns a partial-failure error.
func Callers(idx *Index, root, symbol string) []CallerHit {
	if idx == nil || symbol == "" {
		return nil
	}

	re, err := regexp.Compile(`\b` + regexp.QuoteMeta(symbol) + `\b`)
	if err != nil {
		return nil
	}

	// Build set of definition (file, line) pairs to exclude.
	type fileLine struct {
		file string
		line int
	}
	defLines := make(map[fileLine]struct{})
	for _, file := range idx.Files {
		for _, sym := range file.Symbols {
			if sym.Name == symbol {
				defLines[fileLine{file.Path, sym.Line}] = struct{}{}
			}
		}
	}

	var hits []CallerHit
	for _, file := range idx.Files {
		abs := filepath.Join(root, filepath.FromSlash(file.Path))
		f, err := os.Open(abs)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if !re.MatchString(line) {
				continue
			}
			if _, isDef := defLines[fileLine{file.Path, lineNo}]; isDef {
				continue
			}
			hits = append(hits, CallerHit{
				File:    file.Path,
				Line:    lineNo,
				Snippet: strings.TrimRight(line, " \t\r\n"),
			})
		}
		f.Close()
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].File != hits[j].File {
			return hits[i].File < hits[j].File
		}
		return hits[i].Line < hits[j].Line
	})

	return hits
}

// FileHit is a file scored by relevance to a free-form task description.
type FileHit struct {
	File     string
	Language string
	Score    int // sum of token-symbol-name matches; ties broken by File
}

// contextStopwords is the small stopword list dropped during tokenization.
var contextStopwords = map[string]struct{}{
	"the":  {},
	"and":  {},
	"for":  {},
	"with": {},
	"from": {},
	"into": {},
	"que":  {},
	"con":  {},
	"para": {},
	"por":  {},
	"los":  {},
	"las":  {},
}

// reContextSplit splits a task string on any non-alphanumeric (Unicode-aware)
// run. Compiled once at package load.
var reContextSplit = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// tokenizeContext lowercases task, splits on non-alphanumerics, then drops
// tokens shorter than 3 chars and stopwords. Duplicates are removed.
func tokenizeContext(task string) []string {
	lower := strings.ToLower(task)
	raw := reContextSplit.Split(lower, -1)
	seen := make(map[string]struct{})
	var tokens []string
	for _, t := range raw {
		if len(t) < 3 {
			continue
		}
		if _, stop := contextStopwords[t]; stop {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		tokens = append(tokens, t)
	}
	return tokens
}

// Context tokenizes the task string (split on whitespace and punctuation,
// lowercase, drop tokens shorter than 3 chars and common stopwords), then
// scores each file by counting how many of its Symbols' names contain any
// token (case-insensitive substring). Files with score 0 are excluded.
// Returns up to maxResults hits; if maxResults <= 0, returns all positive
// hits. Sort by Score desc, then File asc.
func Context(idx *Index, task string, maxResults int) []FileHit {
	if idx == nil {
		return nil
	}
	tokens := tokenizeContext(task)
	if len(tokens) == 0 {
		return nil
	}

	var hits []FileHit
	for _, file := range idx.Files {
		score := 0
		for _, sym := range file.Symbols {
			lname := strings.ToLower(sym.Name)
			for _, tok := range tokens {
				if strings.Contains(lname, tok) {
					score++
				}
			}
		}
		if score == 0 {
			continue
		}
		hits = append(hits, FileHit{
			File:     file.Path,
			Language: file.Language,
			Score:    score,
		})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].File < hits[j].File
	})

	if maxResults > 0 && len(hits) > maxResults {
		hits = hits[:maxResults]
	}
	return hits
}
