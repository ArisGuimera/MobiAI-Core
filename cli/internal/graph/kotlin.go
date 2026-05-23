package graph

import (
	"bytes"
	"regexp"
	"strings"
)

// Compiled regexes used by ScanKotlin. Defined at package level so they are
// compiled once.
var (
	// reImport matches `import com.foo.Bar` or `import com.foo.*`.
	reImport = regexp.MustCompile(`^\s*import\s+([\w.]+(?:\.\*)?)\s*$`)

	// reType matches class/object/interface declarations with optional
	// modifiers (group 1 = modifiers, group 2 = keyword, group 3 = name).
	reType = regexp.MustCompile(`^\s*((?:(?:public|private|internal|protected|open|abstract|sealed|data|inline|inner|enum)\s+)*)(class|object|interface)\s+(\w+)`)

	// reFun matches `fun foo(...)`. Skips generic params and extension
	// receiver. Group 1 = modifiers, group 2 = function name.
	reFun = regexp.MustCompile(`^\s*((?:(?:public|private|internal|protected|open|abstract|inline|suspend|override|operator|tailrec)\s+)*)fun\s+(?:<[^>]+>\s+)?(?:[\w.]+\.)?(\w+)\s*\(`)
)

// containerEntry tracks an open class/object/interface in the brace stack.
type containerEntry struct {
	Name  string
	Depth int // brace depth at which the container body opened
}

// ScanKotlin parses a Kotlin source file's bytes and returns a FileIndex
// populated with imports, symbols, language="kotlin", and line count.
// The path field is set to the path argument verbatim (caller is responsible
// for passing a repo-relative path).
func ScanKotlin(path string, content []byte) FileIndex {
	idx := FileIndex{
		Path:     path,
		Language: "kotlin",
		Lines:    countLines(content),
		Imports:  []string{},
		Symbols:  []Symbol{},
	}

	stripped := stripCommentsAndStrings(content)
	lines := bytes.Split(stripped, []byte("\n"))

	var stack []containerEntry
	depth := 0

	for i, rawLine := range lines {
		lineNum := i + 1
		line := string(rawLine)

		// Imports: matched against the stripped line (still preserves
		// identifiers and dots since strings/comments don't contain
		// well-formed import statements after stripping).
		if m := reImport.FindStringSubmatch(line); m != nil {
			idx.Imports = append(idx.Imports, m[1])
		}

		// Container at the START of this line is the innermost open
		// type, if any.
		container := ""
		if len(stack) > 0 {
			container = stack[len(stack)-1].Name
		}

		// Type declarations (class/object/interface).
		if m := reType.FindStringSubmatch(line); m != nil {
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

			// If this line opens a body (`{`), push onto the stack.
			// Otherwise (single-line declaration without body, or
			// `class Foo` with body on next line), still push: when
			// the `{` arrives on a later line we'll already be at
			// the correct depth. We use the depth AFTER processing
			// this line's braces below.
			//
			// Simplest approach: push with depth = current depth.
			// When `}` closes, we pop when we drop back BELOW that
			// depth. We update braces below, then check for pops.
			//
			// We push BEFORE processing this line's braces so that
			// "depth at which body opens" matches the line's `{`.
			stack = append(stack, containerEntry{Name: name, Depth: depth})
		} else {
			// Functions (only checked if line is not a type decl).
			if m := reFun.FindStringSubmatch(line); m != nil {
				modifiers := splitModifiers(m[1])
				name := m[2]
				sym := Symbol{
					Name:      name,
					Kind:      "fun",
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

		// Pop containers whose `body-open depth` is now >= current
		// depth. A container pushed at depth D opens its body on the
		// `{` that follows: that brace takes us to D+1. When we
		// return to depth <= D, the body is closed.
		for len(stack) > 0 && depth <= stack[len(stack)-1].Depth {
			stack = stack[:len(stack)-1]
		}
	}

	return idx
}

// splitModifiers splits the modifier prefix captured by the type/fun regexes
// into a slice of individual modifiers. Returns nil if there are none, so the
// resulting FileIndex omits the field in JSON.
func splitModifiers(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}
	return parts
}

// countLines returns the line count of the original content. Trailing newline
// or not, an empty file is 0 lines and a non-empty file is at least 1.
func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	n := bytes.Count(content, []byte("\n"))
	if !bytes.HasSuffix(content, []byte("\n")) {
		n++
	}
	return n
}

// stripCommentsAndStrings walks the bytes once and replaces the contents of
// line comments, block comments, and string literals with spaces. Newlines
// inside the stripped regions are preserved so that line numbers and column
// positions in the output match the original. This makes downstream regex
// matching immune to "class X" appearing inside a comment or string.
//
// State machine handles:
//   - line comments: `// ... \n`
//   - block comments: `/* ... */` (NOT nested — Kotlin's spec says block
//     comments do nest, but in practice almost no real code relies on it,
//     and the V1 tests use commented-out inner `/* ... */` text on the same
//     line which is fine).
//   - double-quoted strings: `"..."` with `\"` escape.
//   - triple-quoted raw strings: `"""..."""` (no escapes inside).
func stripCommentsAndStrings(content []byte) []byte {
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
				// Leave the opening `"""` in place — they aren't
				// identifiers and won't match our regexes. Just
				// transition state and advance.
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
				// Replace `//` and the rest with spaces until newline.
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
				// Escape: blank out the escape and the next byte
				// (but preserve newline if it's `\` + literal newline).
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
				// Closing quote — leave it.
				i++
				state = stateCode
				continue
			}
			if c == '\n' {
				// Unterminated string — bail back to code mode
				// to avoid eating the whole file on a stray quote.
				state = stateCode
				i++
				continue
			}
			out[i] = ' '
			i++

		case stateTripleString:
			if c == '"' && i+2 < n && content[i+1] == '"' && content[i+2] == '"' {
				// Closing triple quote — leave it in place.
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
