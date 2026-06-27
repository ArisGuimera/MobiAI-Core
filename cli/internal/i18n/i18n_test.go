package i18n

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestT(t *testing.T) {
	t.Cleanup(func() { SetLang(EN) })

	SetLang(EN)
	const key = "Skills catalog updated to v%s.\n"
	if got := T(key); got != key {
		t.Errorf("EN should return the source string; got %q", got)
	}

	SetLang(ES)
	if got := T(key); got != "Catálogo de skills actualizado a v%s.\n" {
		t.Errorf("ES translation wrong; got %q", got)
	}
	if got := T("a string with no translation"); got != "a string with no translation" {
		t.Errorf("ES must fall back to the English source for unknown keys; got %q", got)
	}
}

func TestInit_Precedence(t *testing.T) {
	t.Cleanup(func() { SetLang(EN) })

	t.Setenv("MOBIAI_LANG", "") // empty == unset for Init's purposes
	Init("")
	if Current() != EN {
		t.Errorf("default should be EN; got %s", Current())
	}
	Init("es")
	if Current() != ES {
		t.Errorf("persisted preference should win; got %s", Current())
	}
	Init("garbage")
	if Current() != EN {
		t.Errorf("unrecognized persisted value should fall back to EN; got %s", Current())
	}

	t.Setenv("MOBIAI_LANG", "en")
	Init("es")
	if Current() != EN {
		t.Errorf("MOBIAI_LANG should override the persisted preference; got %s", Current())
	}
}

// verbRe matches Go fmt verbs (with optional flags/width/precision/arg-index).
var verbRe = regexp.MustCompile(`%[#0\-+ '\[\]0-9.*]*[a-zA-Z%]`)

// formatVerbs returns the ordered sequence of verb letters in s, skipping the
// literal "%%". e.g. "got %d of %s" -> ["d","s"].
func formatVerbs(s string) []string {
	var out []string
	for _, m := range verbRe.FindAllString(s, -1) {
		if m == "%%" {
			continue
		}
		out = append(out, m[len(m)-1:])
	}
	return out
}

// TestCatalogVerbParity guards against %!s(MISSING) / wrong-order bugs: every
// (en, es) pair must have the identical ordered sequence of format verbs.
func TestCatalogVerbParity(t *testing.T) {
	for en, es := range catalogES {
		ev, sv := formatVerbs(en), formatVerbs(es)
		if !slices.Equal(ev, sv) {
			t.Errorf("format-verb mismatch:\n  en %q -> %v\n  es %q -> %v", en, ev, es, sv)
		}
	}
}

// TestCatalogCompleteness is the safety net that makes English-as-key safe: it
// statically scans every i18n.T("literal") call in the module and asserts each
// has an es translation (or sits on intentionallyEnglish). Without it, a single
// byte difference between a call site and its catalog key would silently leak
// English in es mode with the build still green.
func TestCatalogCompleteness(t *testing.T) {
	moduleRoot := filepath.Join("..", "..") // internal/i18n -> cli/
	fset := token.NewFileSet()
	var missing []string

	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			switch info.Name() {
			case "embedded", ".git", "testdata", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return fmt.Errorf("parse %s: %w", path, perr)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := ce.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "i18n" || sel.Sel.Name != "T" || len(ce.Args) == 0 {
				return true
			}
			lit, ok := ce.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				t.Errorf("%s: i18n.T() must be called with a string literal (catalog keys are extracted statically)", fset.Position(ce.Pos()))
				return true
			}
			key, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				t.Errorf("%s: unquote %s: %v", fset.Position(ce.Pos()), lit.Value, uerr)
				return true
			}
			if _, ok := catalogES[key]; ok {
				return true
			}
			if intentionallyEnglish[key] {
				return true
			}
			missing = append(missing, fmt.Sprintf("%s: %q", fset.Position(ce.Pos()), key))
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) > 0 {
		t.Errorf("%d wrapped string(s) lack an es translation — add to catalogES or intentionallyEnglish:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}
