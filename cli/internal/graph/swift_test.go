package graph

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// loadSwiftFixture reads a Swift fixture file from testdata/swift/.
func loadSwiftFixture(t *testing.T, name string) (string, []byte) {
	t.Helper()
	path := filepath.Join("testdata", "swift", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return path, content
}

func TestScanSwift_Basic(t *testing.T) {
	path, content := loadSwiftFixture(t, "basic.swift")
	idx := ScanSwift(path, content)

	if idx.Language != "swift" {
		t.Errorf("expected language=swift, got %q", idx.Language)
	}

	wantImports := []string{"Foundation", "Combine", "SwiftUI"}
	if len(idx.Imports) != len(wantImports) {
		t.Fatalf("expected %d imports, got %d: %v", len(wantImports), len(idx.Imports), idx.Imports)
	}
	for i, want := range wantImports {
		if idx.Imports[i] != want {
			t.Errorf("imports[%d]: want %q, got %q", i, want, idx.Imports[i])
		}
	}

	mustHave := []string{
		"LoginViewModel", "User", "AuthRepository", "AuthError",
		"SessionStore", "login", "logout", "authenticate", "clear",
		"set", "displayName",
	}
	for _, name := range mustHave {
		if findSymbol(idx.Symbols, name) == nil {
			t.Errorf("expected symbol %q in symbols; got: %v", name, symbolNames(idx.Symbols))
		}
	}

	// Extension surfaces with the extended-type name: there must be a
	// symbol "User" with kind "extension".
	foundExtUser := false
	for _, s := range idx.Symbols {
		if s.Name == "User" && s.Kind == "extension" {
			foundExtUser = true
			break
		}
	}
	if !foundExtUser {
		t.Errorf("expected an extension symbol named \"User\"; got: %v", idx.Symbols)
	}

	if login := findSymbol(idx.Symbols, "login"); login != nil {
		if login.Container != "LoginViewModel" {
			t.Errorf("login.Container: want LoginViewModel, got %q", login.Container)
		}
	}

	// SessionStore must be present with Kind == "actor".
	foundActor := false
	for _, s := range idx.Symbols {
		if s.Name == "SessionStore" && s.Kind == "actor" {
			foundActor = true
			break
		}
	}
	if !foundActor {
		t.Errorf("expected SessionStore with Kind=actor; got: %v", idx.Symbols)
	}

	// The User struct (Kind=="struct") must have empty Container.
	foundUserStruct := false
	for _, s := range idx.Symbols {
		if s.Name == "User" && s.Kind == "struct" {
			foundUserStruct = true
			if s.Container != "" {
				t.Errorf("User struct Container should be empty, got %q", s.Container)
			}
		}
	}
	if !foundUserStruct {
		t.Errorf("expected a struct symbol named \"User\"; got: %v", idx.Symbols)
	}
}

func TestScanSwift_IgnoresComments(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

	forbidden := []string{"NotARealClass", "AlsoNotReal", "alsoIgnoredFn"}
	for _, name := range forbidden {
		if findSymbol(idx.Symbols, name) != nil {
			t.Errorf("symbol %q should be ignored (inside comment), but was extracted", name)
		}
	}
}

func TestScanSwift_IgnoresStringLiterals(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

	forbidden := []string{"FakeClass", "fakeFn", "TripleQuotedFake", "anotherFake"}
	for _, name := range forbidden {
		if findSymbol(idx.Symbols, name) != nil {
			t.Errorf("symbol %q should be ignored (inside string literal), but was extracted", name)
		}
	}
}

func TestScanSwift_HandlesGenerics(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

	if findSymbol(idx.Symbols, "transform") == nil {
		t.Errorf("expected generic function 'transform' to be extracted; got: %v", symbolNames(idx.Symbols))
	}
}

func TestScanSwift_HandlesNestedTypes(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

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

func TestScanSwift_HandlesModifiers(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

	finalClass := findSymbol(idx.Symbols, "FinalClass")
	if finalClass == nil {
		t.Fatalf("expected 'FinalClass' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if !contains(finalClass.Modifiers, "public") || !contains(finalClass.Modifiers, "final") {
		t.Errorf("FinalClass.Modifiers should contain public and final, got %v", finalClass.Modifiers)
	}

	staticFn := findSymbol(idx.Symbols, "staticFn")
	if staticFn == nil {
		t.Fatalf("expected 'staticFn' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if !contains(staticFn.Modifiers, "public") || !contains(staticFn.Modifiers, "static") {
		t.Errorf("staticFn.Modifiers should contain public and static, got %v", staticFn.Modifiers)
	}

	isolatedFn := findSymbol(idx.Symbols, "isolatedFn")
	if isolatedFn == nil {
		t.Fatalf("expected 'isolatedFn' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if !contains(isolatedFn.Modifiers, "nonisolated") {
		t.Errorf("isolatedFn.Modifiers should contain nonisolated, got %v", isolatedFn.Modifiers)
	}
}

func TestScanSwift_HandlesExtensions(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

	foundExt := false
	for _, s := range idx.Symbols {
		if s.Name == "String" && s.Kind == "extension" {
			foundExt = true
			break
		}
	}
	if !foundExt {
		t.Errorf("expected an extension symbol named \"String\"; got: %v", idx.Symbols)
	}

	extensionFn := findSymbol(idx.Symbols, "extensionFn")
	if extensionFn == nil {
		t.Fatalf("expected 'extensionFn' in symbols; got: %v", symbolNames(idx.Symbols))
	}
	if extensionFn.Container != "String" {
		t.Errorf("extensionFn.Container: want String, got %q", extensionFn.Container)
	}
}

func TestScanSwift_HandlesAnnotations(t *testing.T) {
	path, content := loadSwiftFixture(t, "edge_cases.swift")
	idx := ScanSwift(path, content)

	if findSymbol(idx.Symbols, "mainActorFn") == nil {
		t.Errorf("expected 'mainActorFn' (annotated function) to be extracted; got: %v", symbolNames(idx.Symbols))
	}
}

func TestScanSwift_UnclosedBlockCommentAtEOF(t *testing.T) {
	src := []byte(`import Foundation

class RealClass {}

/* unclosed comment from here to EOF
class FakeClass {}
func fakeFn() {}`)

	idx := ScanSwift("frag.swift", src)

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

func TestScanSwift_LineCount(t *testing.T) {
	data, err := os.ReadFile("testdata/swift/basic.swift")
	if err != nil {
		t.Fatal(err)
	}
	expected := bytes.Count(data, []byte("\n"))
	if len(data) > 0 && data[len(data)-1] != '\n' {
		expected++
	}
	idx := ScanSwift("basic.swift", data)
	if idx.Lines != expected {
		t.Errorf("Lines = %d, want %d", idx.Lines, expected)
	}
}

func TestScanSwift_PathSetVerbatim(t *testing.T) {
	idx := ScanSwift("custom/path.swift", []byte("struct X {}"))
	if idx.Path != "custom/path.swift" {
		t.Errorf("Path: want %q, got %q", "custom/path.swift", idx.Path)
	}
}
