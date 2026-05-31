package i18n

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestNoUntranslatedUserStrings is the reverse of TestCatalogCompleteness: that
// test proves every wrapped string HAS a translation; this one proves every
// user-facing string IS wrapped. Together they make "mobiai lang es" trustworthy
// and fail CI when someone adds a new CLI string without localizing it.
//
// It statically scans the user-facing packages (cmd, tui, main) for:
//   - cobra Short/Long values that are raw string literals (must be i18n.T(...))
//   - flag usage descriptions that are raw string literals (must be i18n.T(...))
//   - prose strings printed to the user (fmt.Fprint*/Print*, cmd.Print*) that
//     are not wrapped
//
// It deliberately does NOT scan internal/brain, internal/graph, internal/host,
// or skills: error messages, MCP tool descriptions and skill content stay in
// English by design.
//
// An intentional exception (a symbol, layout string, or identifier that is the
// same in every language) goes in guardAllow below — a visible, reviewable
// opt-out rather than a silent gap.
func TestNoUntranslatedUserStrings(t *testing.T) {
	// User-facing packages, relative to internal/i18n.
	dirs := []string{
		filepath.Join("..", "cmd"),
		filepath.Join("..", "tui"),
		filepath.Join("..", "..", "cmd", "mobiai"),
	}

	var violations []string
	fset := token.NewFileSet()

	for _, dir := range dirs {
		pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", dir, err)
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				ast.Inspect(file, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.KeyValueExpr:
						checkShortLong(t, fset, node, &violations)
					case *ast.CallExpr:
						checkFlagUsage(fset, node, &violations)
						checkPrint(fset, node, &violations)
					}
					return true
				})
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("%d user-facing string(s) are not localized.\n"+
			"Wrap each with i18n.T(\"...\") so `mobiai lang es` can translate it (then add the\n"+
			"Spanish to a catalog_*_es.go). If a string is intentionally English (a symbol,\n"+
			"layout, or identifier — NOT skill/MCP content, which is English by design), add it\n"+
			"to guardAllow in i18n_guard_test.go.\n\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// checkShortLong flags cobra Short:/Long: values that aren't i18n.T(...) calls.
func checkShortLong(t *testing.T, fset *token.FileSet, kv *ast.KeyValueExpr, out *[]string) {
	key, ok := kv.Key.(*ast.Ident)
	if !ok || (key.Name != "Short" && key.Name != "Long") {
		return
	}
	if litStr, ok := stringLiteralOf(kv.Value); ok {
		if !guardAllow[litStr] {
			*out = append(*out, fmt.Sprintf("%s: cobra %s: %q", fset.Position(kv.Pos()), key.Name, litStr))
		}
	}
}

// flagSetters are the pflag methods whose LAST argument is a usage string.
var flagSetters = map[string]bool{
	"String": true, "StringP": true, "StringArray": true, "StringArrayP": true,
	"StringSlice": true, "StringSliceP": true, "Bool": true, "BoolP": true,
	"Int": true, "IntP": true, "Int64": true, "Duration": true, "Float64": true,
}

func checkFlagUsage(fset *token.FileSet, ce *ast.CallExpr, out *[]string) {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok || !flagSetters[sel.Sel.Name] || len(ce.Args) == 0 {
		return
	}
	last := ce.Args[len(ce.Args)-1]
	if litStr, ok := stringLiteralOf(last); ok {
		if !guardAllow[litStr] {
			*out = append(*out, fmt.Sprintf("%s: flag usage: %q", fset.Position(last.Pos()), litStr))
		}
	}
}

// printFuncs are output funcs whose format/string arg should be localized.
var printFuncs = map[string]bool{
	"Fprintf": true, "Fprintln": true, "Fprint": true,
	"Printf": true, "Println": true, "Print": true,
}

func checkPrint(fset *token.FileSet, ce *ast.CallExpr, out *[]string) {
	sel, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok || !printFuncs[sel.Sel.Name] {
		return
	}
	// For fmt.Fprint*, the writer is arg 0 and the format/string is arg 1.
	// For Print*/cmd.Print*, the format/string is arg 0.
	idx := 0
	if strings.HasPrefix(sel.Sel.Name, "Fprint") {
		idx = 1
	}
	if len(ce.Args) <= idx {
		return
	}
	litStr, ok := stringLiteralOf(ce.Args[idx])
	if !ok {
		return // already wrapped (i18n.T) or a variable — fine
	}
	if guardAllow[litStr] || !isProse(litStr) {
		return // symbol/layout/format-only, or an explicit exception
	}
	*out = append(*out, fmt.Sprintf("%s: printed string: %q", fset.Position(ce.Args[idx].Pos()), litStr))
}

// stringLiteralOf returns the unquoted value when expr is a bare string literal
// or a concatenation of string literals (`"a" + "b"`). It returns ok=false for
// anything else (e.g. an i18n.T(...) call or a variable), which is treated as
// "not a raw user-facing literal".
func stringLiteralOf(expr ast.Expr) (string, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(e.Value)
		if err != nil {
			return "", false
		}
		return s, true
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return "", false
		}
		l, lok := stringLiteralOf(e.X)
		r, rok := stringLiteralOf(e.Y)
		if lok && rok {
			return l + r, true
		}
		return "", false
	}
	return "", false
}

// verbStripper removes Go format verbs so isProse only sees real words.
var verbStripper = regexp.MustCompile(`%[#+\- 0-9.*]*[a-zA-Z]`)

// proseWord matches a run of 4+ letters — enough to call a string "prose"
// rather than a symbol/layout/identifier fragment.
var proseWord = regexp.MustCompile(`\p{L}{4,}`)

// isProse reports whether s reads as natural language (and so should be
// localized) rather than pure layout/symbols (e.g. "%-13s │ %s\n", "✓ %s\n").
func isProse(s string) bool {
	return proseWord.MatchString(verbStripper.ReplaceAllString(s, ""))
}

// guardAllow lists user-facing-position string literals that are intentionally
// left English: layout/table glue, tech-category labels, and identifiers that
// are identical in every language. Each entry is a deliberate, reviewable
// opt-out from localization. Skill/MCP content is NOT here — those packages are
// not scanned at all.
var guardAllow = map[string]bool{
	// Memory-schema field labels — these mirror the YAML frontmatter keys of a
	// memory entry (identifiers, not prose). brain.go review/overdue/promote.
	"    platform: %s\n":        true,
	"    area: %s\n":            true,
	"  status: %s → %s\n":       true,
	"  review_after: %s → %s\n": true,

	// Domain terms kept English across the whole product (pack/host are used
	// untranslated everywhere, like "skills"). Headers in catalog/skills/status.
	"Packs (%d):\n":         true,
	"Pack          | Hosts": true,
	"  Packs (%d): %s\n":    true,
	"  Hosts (%d): %s\n":    true,
	"Hosts:":                true,

	// Pure layout / technical annotation (no translatable prose).
	"  - %-14s v%-8s deps: %s\n": true,
	"%s  (score %d)\n":           true,

	// Brand wordmark line (mobiai status header).
	"MobiAI %s\n\n": true,

	// Machine-readable JSON output — not UI prose. (mobiai skills, --json path)
	"{\"name\":%q,\"source\":\"./skills/%s\",\"description\":\"\",\"category\":\"\"}": true,
}
