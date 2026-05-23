package graph

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// loadFixture reads a Kotlin fixture file from testdata/kotlin/.
func loadFixture(t *testing.T, name string) (string, []byte) {
	t.Helper()
	path := filepath.Join("testdata", "kotlin", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return path, content
}

// findSymbol returns the first symbol matching the given name, or nil if not found.
func findSymbol(symbols []Symbol, name string) *Symbol {
	for i := range symbols {
		if symbols[i].Name == name {
			return &symbols[i]
		}
	}
	return nil
}

// symbolNames returns just the symbol names for assertions.
func symbolNames(symbols []Symbol) []string {
	out := make([]string, len(symbols))
	for i, s := range symbols {
		out[i] = s.Name
	}
	return out
}

// contains reports whether s is present in xs.
func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func TestScanKotlin_Basic(t *testing.T) {
	path, content := loadFixture(t, "basic.kt")
	idx := ScanKotlin(path, content)

	if idx.Language != "kotlin" {
		t.Errorf("expected language=kotlin, got %q", idx.Language)
	}

	wantImports := []string{"com.x.y.Foo", "kotlinx.coroutines.flow.Flow"}
	if len(idx.Imports) != len(wantImports) {
		t.Fatalf("expected %d imports, got %d: %v", len(wantImports), len(idx.Imports), idx.Imports)
	}
	for i, want := range wantImports {
		if idx.Imports[i] != want {
			t.Errorf("imports[%d]: want %q, got %q", i, want, idx.Imports[i])
		}
	}

	mustHave := []string{
		"LoginViewModel", "Config", "AuthRepository", "User", "Result",
		"Success", "Error", "login", "logout", "authenticate", "clear",
	}
	for _, name := range mustHave {
		if findSymbol(idx.Symbols, name) == nil {
			t.Errorf("expected symbol %q in symbols; got: %v", name, symbolNames(idx.Symbols))
		}
	}

	if findSymbol(idx.Symbols, "baseUrl") != nil {
		t.Errorf("did not expect val/var %q to be extracted in V1", "baseUrl")
	}

	if login := findSymbol(idx.Symbols, "login"); login != nil {
		if login.Container != "LoginViewModel" {
			t.Errorf("login.Container: want LoginViewModel, got %q", login.Container)
		}
	}

	if success := findSymbol(idx.Symbols, "Success"); success != nil {
		if success.Container != "Result" {
			t.Errorf("Success.Container: want Result, got %q", success.Container)
		}
	}

	if user := findSymbol(idx.Symbols, "User"); user != nil {
		found := false
		for _, m := range user.Modifiers {
			if m == "data" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("User.Modifiers should contain \"data\", got %v", user.Modifiers)
		}
	}

	if vm := findSymbol(idx.Symbols, "LoginViewModel"); vm != nil {
		if vm.Container != "" {
			t.Errorf("LoginViewModel.Container should be empty (top-level), got %q", vm.Container)
		}
	}
}

func TestScanKotlin_IgnoresComments(t *testing.T) {
	path, content := loadFixture(t, "edge_cases.kt")
	idx := ScanKotlin(path, content)

	forbidden := []string{"NotARealClass", "AlsoNotReal", "alsoIgnoredFn"}
	for _, name := range forbidden {
		if findSymbol(idx.Symbols, name) != nil {
			t.Errorf("symbol %q should be ignored (inside comment), but was extracted", name)
		}
	}
}

func TestScanKotlin_IgnoresStringLiterals(t *testing.T) {
	path, content := loadFixture(t, "edge_cases.kt")
	idx := ScanKotlin(path, content)

	forbidden := []string{"FakeClass", "fakeFn", "TripleQuotedFake", "anotherFake"}
	for _, name := range forbidden {
		if findSymbol(idx.Symbols, name) != nil {
			t.Errorf("symbol %q should be ignored (inside string literal), but was extracted", name)
		}
	}
}

func TestScanKotlin_UnclosedBlockCommentAtEOF(t *testing.T) {
	// Input ends mid block-comment. Scanner must:
	//  - Not panic / not loop infinitely.
	//  - Treat everything inside the unclosed comment as stripped (no symbols extracted).
	//  - Return the FileIndex without error (no error channel; just verify it returns).
	src := []byte(`package x

class RealClass {}

/* unclosed comment from here to EOF
class FakeClass {}
fun fakeFn() {}`)

	idx := ScanKotlin("frag.kt", src)

	names := symbolNames(idx.Symbols)
	if !contains(names, "RealClass") {
		t.Errorf("expected RealClass to be extracted, got %v", names)
	}
	for _, forbidden := range []string{"FakeClass", "fakeFn"} {
		if contains(names, forbidden) {
			t.Errorf("symbol %q must NOT appear (inside unclosed block comment), got %v", forbidden, names)
		}
	}
}

func TestScanKotlin_HandlesGenerics(t *testing.T) {
	path, content := loadFixture(t, "edge_cases.kt")
	idx := ScanKotlin(path, content)

	if findSymbol(idx.Symbols, "transform") == nil {
		t.Errorf("expected generic function 'transform' to be extracted; got: %v", symbolNames(idx.Symbols))
	}
}

func TestScanKotlin_HandlesExtensionFunctions(t *testing.T) {
	path, content := loadFixture(t, "edge_cases.kt")
	idx := ScanKotlin(path, content)

	if findSymbol(idx.Symbols, "extensionFn") == nil {
		t.Errorf("expected extension function 'extensionFn' to be extracted; got: %v", symbolNames(idx.Symbols))
	}
}

func TestScanKotlin_HandlesNestedTypes(t *testing.T) {
	path, content := loadFixture(t, "edge_cases.kt")
	idx := ScanKotlin(path, content)

	outer := findSymbol(idx.Symbols, "Outer")
	if outer == nil {
		t.Fatalf("expected 'Outer' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if outer.Container != "" {
		t.Errorf("Outer.Container should be empty, got %q", outer.Container)
	}

	inner := findSymbol(idx.Symbols, "Inner")
	if inner == nil {
		t.Fatalf("expected 'Inner' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if inner.Container != "Outer" {
		t.Errorf("Inner.Container: want Outer, got %q", inner.Container)
	}

	innerMethod := findSymbol(idx.Symbols, "innerMethod")
	if innerMethod == nil {
		t.Fatalf("expected 'innerMethod' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if innerMethod.Container != "Inner" {
		t.Errorf("innerMethod.Container: want Inner, got %q", innerMethod.Container)
	}
}

func TestScanKotlin_HandlesAnnotations(t *testing.T) {
	path, content := loadFixture(t, "edge_cases.kt")
	idx := ScanKotlin(path, content)

	if findSymbol(idx.Symbols, "ComposableScreen") == nil {
		t.Errorf("expected 'ComposableScreen' (annotated function) to be extracted; got: %v", symbolNames(idx.Symbols))
	}
}

func TestScanKotlin_LineCount(t *testing.T) {
	data, err := os.ReadFile("testdata/kotlin/basic.kt")
	if err != nil {
		t.Fatal(err)
	}
	expected := bytes.Count(data, []byte("\n"))
	if len(data) > 0 && data[len(data)-1] != '\n' {
		expected++
	}
	idx := ScanKotlin("basic.kt", data)
	if idx.Lines != expected {
		t.Errorf("Lines = %d, want %d", idx.Lines, expected)
	}
}

func TestScanKotlin_PathSetVerbatim(t *testing.T) {
	idx := ScanKotlin("custom/path.kt", []byte("class X {}"))
	if idx.Path != "custom/path.kt" {
		t.Errorf("Path: want %q, got %q", "custom/path.kt", idx.Path)
	}
}
