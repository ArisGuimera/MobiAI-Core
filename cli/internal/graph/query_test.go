package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// -- Search ---------------------------------------------------------------

func TestSearch_RanksExactBeforeSubstring(t *testing.T) {
	idx := &Index{
		Files: []FileIndex{
			{Path: "a.kt", Language: "kotlin", Symbols: []Symbol{{Name: "Login", Kind: "class", Line: 1}}},
			{Path: "b.kt", Language: "kotlin", Symbols: []Symbol{{Name: "LoginViewModel", Kind: "class", Line: 1}}},
			{Path: "c.kt", Language: "kotlin", Symbols: []Symbol{{Name: "helloLogin", Kind: "fun", Line: 1}}},
		},
	}
	hits := Search(idx, "Login")
	if len(hits) != 3 {
		t.Fatalf("want 3 hits, got %d", len(hits))
	}
	if hits[0].Symbol.Name != "Login" || hits[0].Score != 3 {
		t.Errorf("hit[0]: got name=%q score=%d, want Login/3", hits[0].Symbol.Name, hits[0].Score)
	}
	if hits[1].Symbol.Name != "LoginViewModel" || hits[1].Score != 2 {
		t.Errorf("hit[1]: got name=%q score=%d, want LoginViewModel/2", hits[1].Symbol.Name, hits[1].Score)
	}
	if hits[2].Symbol.Name != "helloLogin" || hits[2].Score != 1 {
		t.Errorf("hit[2]: got name=%q score=%d, want helloLogin/1", hits[2].Symbol.Name, hits[2].Score)
	}
	if hits[0].File != "a.kt" || hits[0].Language != "kotlin" {
		t.Errorf("hit[0]: file=%q lang=%q", hits[0].File, hits[0].Language)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	idx := &Index{
		Files: []FileIndex{
			{Path: "b.kt", Language: "kotlin", Symbols: []Symbol{{Name: "LoginViewModel", Kind: "class", Line: 1}}},
		},
	}
	hits := Search(idx, "loginVIEWMODEL")
	if len(hits) != 1 {
		t.Fatalf("want 1 hit, got %d", len(hits))
	}
	if hits[0].Score != 3 {
		t.Errorf("want score 3 (exact case-insensitive), got %d", hits[0].Score)
	}
}

func TestSearch_EmptyTerm(t *testing.T) {
	idx := &Index{
		Files: []FileIndex{
			{Path: "a.kt", Symbols: []Symbol{{Name: "Foo", Line: 1}}},
		},
	}
	if got := Search(idx, ""); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestSearch_NilIndex(t *testing.T) {
	if got := Search(nil, "x"); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestSearch_TieBreakByFileThenLine(t *testing.T) {
	idx := &Index{
		Files: []FileIndex{
			{Path: "a.kt", Language: "kotlin", Symbols: []Symbol{
				{Name: "Foo", Kind: "class", Line: 10},
				{Name: "Foo", Kind: "class", Line: 5},
			}},
			{Path: "b.kt", Language: "kotlin", Symbols: []Symbol{
				{Name: "Foo", Kind: "class", Line: 1},
			}},
		},
	}
	hits := Search(idx, "Foo")
	if len(hits) != 3 {
		t.Fatalf("want 3 hits, got %d", len(hits))
	}
	if hits[0].File != "a.kt" || hits[0].Symbol.Line != 5 {
		t.Errorf("hit[0]: %s:%d, want a.kt:5", hits[0].File, hits[0].Symbol.Line)
	}
	if hits[1].File != "a.kt" || hits[1].Symbol.Line != 10 {
		t.Errorf("hit[1]: %s:%d, want a.kt:10", hits[1].File, hits[1].Symbol.Line)
	}
	if hits[2].File != "b.kt" || hits[2].Symbol.Line != 1 {
		t.Errorf("hit[2]: %s:%d, want b.kt:1", hits[2].File, hits[2].Symbol.Line)
	}
}

// -- Callers --------------------------------------------------------------

func TestCallers_FindsReferences(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.kt"), "class Foo {}\n")
	writeFile(t, filepath.Join(root, "b.kt"), "fun bar() {\n    val x = Foo()\n    val y = Foo\n}\n")

	idx, err := Build(root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	hits := Callers(idx, root, "Foo")
	if len(hits) != 2 {
		t.Fatalf("want 2 hits, got %d: %+v", len(hits), hits)
	}
	for _, h := range hits {
		if h.File != "b.kt" {
			t.Errorf("unexpected file in hit: %q", h.File)
		}
	}
	if hits[0].Line != 2 || hits[1].Line != 3 {
		t.Errorf("lines: got %d,%d want 2,3", hits[0].Line, hits[1].Line)
	}
	if hits[0].Snippet != "    val x = Foo()" {
		t.Errorf("snippet[0]: %q", hits[0].Snippet)
	}
	if hits[1].Snippet != "    val y = Foo" {
		t.Errorf("snippet[1]: %q", hits[1].Snippet)
	}
}

func TestCallers_ExcludesDefinitionLine(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.kt"), "class Foo {}\n// usage of Foo\n")

	idx, err := Build(root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	hits := Callers(idx, root, "Foo")
	if len(hits) != 1 {
		t.Fatalf("want 1 hit (comment line only), got %d: %+v", len(hits), hits)
	}
	if hits[0].Line != 2 {
		t.Errorf("want line 2 (comment), got %d", hits[0].Line)
	}
}

func TestCallers_WholeWordOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "b.kt"), "val a = Foobar\nval b = Foo()\n")

	idx, err := Build(root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	hits := Callers(idx, root, "Foo")
	if len(hits) != 1 {
		t.Fatalf("want 1 hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].Line != 2 {
		t.Errorf("want line 2, got %d", hits[0].Line)
	}
}

func TestCallers_NilIndex(t *testing.T) {
	if got := Callers(nil, "/tmp", "x"); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestCallers_MissingFileSkipped(t *testing.T) {
	root := t.TempDir()
	idx := &Index{
		Root: root,
		Files: []FileIndex{
			{Path: "does/not/exist.kt", Language: "kotlin"},
		},
	}
	hits := Callers(idx, root, "Foo")
	if hits != nil && len(hits) != 0 {
		t.Errorf("want no hits, got %+v", hits)
	}
}

// -- Context --------------------------------------------------------------

func TestContext_RanksByTokenOverlap(t *testing.T) {
	idx := &Index{
		Files: []FileIndex{
			{Path: "login.kt", Language: "kotlin", Symbols: []Symbol{
				{Name: "LoginViewModel"}, {Name: "LoginUseCase"},
			}},
			{Path: "logout.kt", Language: "kotlin", Symbols: []Symbol{
				{Name: "LogoutViewModel"},
			}},
			{Path: "unrelated.kt", Language: "kotlin", Symbols: []Symbol{
				{Name: "WeatherWidget"},
			}},
		},
	}
	hits := Context(idx, "fix login bug", 0)
	if len(hits) != 1 {
		t.Fatalf("want 1 hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].File != "login.kt" || hits[0].Score != 2 {
		t.Errorf("got %s score=%d, want login.kt score=2", hits[0].File, hits[0].Score)
	}
}

func TestContext_TokenizesProperly(t *testing.T) {
	// Tokens after tokenization of "Fix the login-flow & navigation":
	//   {"fix", "login", "flow", "navigation"} ("the" dropped as stopword).
	idx := &Index{
		Files: []FileIndex{
			{Path: "fix.kt", Symbols: []Symbol{{Name: "Fixer"}}},
			{Path: "login.kt", Symbols: []Symbol{{Name: "LoginPage"}}},
			{Path: "flow.kt", Symbols: []Symbol{{Name: "FlowController"}}},
			{Path: "nav.kt", Symbols: []Symbol{{Name: "NavigationStack"}}},
			{Path: "the.kt", Symbols: []Symbol{{Name: "TheThing"}}}, // "the" is stopword
			{Path: "noise.kt", Symbols: []Symbol{{Name: "Unrelated"}}},
		},
	}
	hits := Context(idx, "Fix the login-flow & navigation", 0)
	// Expect 4 hits: fix.kt, flow.kt, login.kt, nav.kt (sorted by Score desc then File asc).
	files := map[string]bool{}
	for _, h := range hits {
		files[h.File] = true
	}
	want := []string{"fix.kt", "login.kt", "flow.kt", "nav.kt"}
	for _, w := range want {
		if !files[w] {
			t.Errorf("missing expected file %q in hits %+v", w, hits)
		}
	}
	if files["the.kt"] {
		t.Errorf("the.kt should not match — 'the' is a stopword")
	}
	if files["noise.kt"] {
		t.Errorf("noise.kt should not match")
	}
}

func TestContext_DropsStopwordsAndShortTokens(t *testing.T) {
	// Tokens after tokenization of "fix the los con bug":
	//   {"fix", "bug"} ("the", "los", "con" are stopwords).
	idx := &Index{
		Files: []FileIndex{
			{Path: "fix.kt", Symbols: []Symbol{{Name: "Fixer"}}},
			{Path: "bug.kt", Symbols: []Symbol{{Name: "Bugfix"}}},
			{Path: "the.kt", Symbols: []Symbol{{Name: "TheThing"}}},
			{Path: "los.kt", Symbols: []Symbol{{Name: "LosFoo"}}},
			{Path: "con.kt", Symbols: []Symbol{{Name: "ConBar"}}},
		},
	}
	hits := Context(idx, "fix the los con bug", 0)
	files := map[string]bool{}
	for _, h := range hits {
		files[h.File] = true
	}
	if !files["fix.kt"] || !files["bug.kt"] {
		t.Errorf("expected fix.kt and bug.kt in hits, got %+v", hits)
	}
	if files["the.kt"] || files["los.kt"] || files["con.kt"] {
		t.Errorf("stopword files should not match: %+v", hits)
	}
}

func TestContext_MaxResults(t *testing.T) {
	// Each file has a symbol containing "widget"; with task "widget",
	// each file scores 1.
	idx := &Index{
		Files: []FileIndex{
			{Path: "a.kt", Symbols: []Symbol{{Name: "WidgetA"}}},
			{Path: "b.kt", Symbols: []Symbol{{Name: "WidgetB"}}},
			{Path: "c.kt", Symbols: []Symbol{{Name: "WidgetC"}}},
			{Path: "d.kt", Symbols: []Symbol{{Name: "WidgetD"}}},
			{Path: "e.kt", Symbols: []Symbol{{Name: "WidgetE"}}},
		},
	}
	if got := len(Context(idx, "widget", 2)); got != 2 {
		t.Errorf("max=2: got %d", got)
	}
	if got := len(Context(idx, "widget", 0)); got != 5 {
		t.Errorf("max=0 (all): got %d", got)
	}
}

func TestContext_NilIndex(t *testing.T) {
	if got := Context(nil, "anything", 0); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

// Sanity: ensure writeFile from indexer_test.go is reachable (compile-time).
var _ = os.WriteFile
