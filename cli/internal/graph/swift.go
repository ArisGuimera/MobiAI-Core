package graph

import (
	"bytes"
	"regexp"
	"strings"
)

// Compiled regexes used by ScanSwift. Defined at package level so they are
// compiled once.
var (
	// reSwiftImport matches `import Foundation` or `import Combine.Publishers`.
	// Single capture group = module path.
	reSwiftImport = regexp.MustCompile(`^\s*import\s+([\w.]+)\s*$`)

	// reSwiftType matches class/struct/enum/protocol/actor/extension
	// declarations with optional modifiers (group 1 = modifiers,
	// group 2 = keyword, group 3 = name).
	reSwiftType = regexp.MustCompile(`^\s*((?:(?:public|private|fileprivate|internal|open|final)\s+)*)(class|struct|enum|protocol|actor|extension)\s+(\w+)`)

	// reSwiftFunc matches `func foo(...)` or `func foo<T>(...)`.
	// Group 1 = modifiers, group 2 = function name.
	// The trailing `[<(]` matches either generic params or the opening paren.
	reSwiftFunc = regexp.MustCompile(`^\s*((?:(?:public|private|fileprivate|internal|open|static|class|mutating|override|final|nonisolated)\s+)*)func\s+(\w+)\s*[<(]`)
)

// ScanSwift parses a Swift source file's bytes and returns a FileIndex
// populated with imports, symbols, language="swift", and line count.
// The path field is set to the path argument verbatim (caller is responsible
// for passing a repo-relative path).
func ScanSwift(path string, content []byte) FileIndex {
	idx := FileIndex{
		Path:     path,
		Language: "swift",
		Lines:    countLines(content),
		Imports:  []string{},
		Symbols:  []Symbol{},
	}

	stripped := stripSwiftCommentsAndStrings(content)
	lines := bytes.Split(stripped, []byte("\n"))

	var stack []containerEntry
	depth := 0

	for i, rawLine := range lines {
		lineNum := i + 1
		line := string(rawLine)

		// Imports.
		if m := reSwiftImport.FindStringSubmatch(line); m != nil {
			idx.Imports = append(idx.Imports, m[1])
		}

		// Container at the START of this line is the innermost open
		// type, if any.
		container := ""
		if len(stack) > 0 {
			container = stack[len(stack)-1].Name
		}

		// Type declarations (class/struct/enum/protocol/actor/extension).
		if m := reSwiftType.FindStringSubmatch(line); m != nil {
			modifiers := splitModifiers(m[1])
			kind := m[2]
			name := m[3]

			sym := Symbol{
				Name:      name,
				Kind:      kind,
				Line:      lineNum,
				Container: container,
				Modifiers: modifiers,
			}
			idx.Symbols = append(idx.Symbols, sym)

			// Push onto the stack so subsequent symbols know their
			// container. Pop happens when brace depth drops back.
			stack = append(stack, containerEntry{Name: name, Depth: depth})
		} else {
			// Functions (only if line is not a type decl).
			if m := reSwiftFunc.FindStringSubmatch(line); m != nil {
				modifiers := splitModifiers(m[1])
				name := m[2]
				sym := Symbol{
					Name:      name,
					Kind:      "func",
					Line:      lineNum,
					Container: container,
					Modifiers: modifiers,
				}
				idx.Symbols = append(idx.Symbols, sym)
			}
		}

		// Update brace depth based on this line's braces, then pop
		// any containers whose body has closed.
		opens := strings.Count(line, "{")
		closes := strings.Count(line, "}")
		depth += opens
		depth -= closes

		for len(stack) > 0 && depth <= stack[len(stack)-1].Depth {
			stack = stack[:len(stack)-1]
		}
	}

	return idx
}

// stripSwiftCommentsAndStrings walks the bytes once and replaces the contents
// of line comments, block comments, and string literals with spaces. Newlines
// inside the stripped regions are preserved so that line numbers match the
// original. This makes downstream regex matching immune to "class X" appearing
// inside a comment or string.
//
// State machine handles:
//   - line comments: `// ... \n`
//   - block comments: `/* ... */`
//   - double-quoted strings: `"..."` with `\"` escape.
//   - triple-quoted multiline strings: `"""..."""`
//
// This is intentionally a near-copy of stripCommentsAndStrings in kotlin.go:
// Swift and Kotlin share the same comment and string-literal forms. We keep
// them separate per task constraints (no shared extraction yet).
func stripSwiftCommentsAndStrings(content []byte) []byte {
	out := make([]byte, len(content))
	copy(out, content)

	const (
		stateCode = iota
		stateLineComment
		stateBlockComment
		stateString       // "..."
		stateTripleString // """..."""
	)

	state := stateCode
	i := 0
	n := len(content)

	for i < n {
		c := content[i]

		switch state {
		case stateCode:
			// Triple-quoted string opener.
			if c == '"' && i+2 < n && content[i+1] == '"' && content[i+2] == '"' {
				i += 3
				state = stateTripleString
				continue
			}
			if c == '"' {
				i++
				state = stateString
				continue
			}
			if c == '/' && i+1 < n && content[i+1] == '/' {
				out[i] = ' '
				out[i+1] = ' '
				i += 2
				state = stateLineComment
				continue
			}
			if c == '/' && i+1 < n && content[i+1] == '*' {
				out[i] = ' '
				out[i+1] = ' '
				i += 2
				state = stateBlockComment
				continue
			}
			i++

		case stateLineComment:
			if c == '\n' {
				state = stateCode
				i++
				continue
			}
			out[i] = ' '
			i++

		case stateBlockComment:
			if c == '*' && i+1 < n && content[i+1] == '/' {
				out[i] = ' '
				out[i+1] = ' '
				i += 2
				state = stateCode
				continue
			}
			if c != '\n' {
				out[i] = ' '
			}
			i++

		case stateString:
			if c == '\\' && i+1 < n {
				if content[i+1] != '\n' {
					out[i] = ' '
					out[i+1] = ' '
					i += 2
					continue
				}
				out[i] = ' '
				i++
				continue
			}
			if c == '"' {
				i++
				state = stateCode
				continue
			}
			if c == '\n' {
				// Unterminated string — bail back to code mode.
				state = stateCode
				i++
				continue
			}
			out[i] = ' '
			i++

		case stateTripleString:
			if c == '"' && i+2 < n && content[i+1] == '"' && content[i+2] == '"' {
				i += 3
				state = stateCode
				continue
			}
			if c != '\n' {
				out[i] = ' '
			}
			i++
		}
	}

	return out
}
