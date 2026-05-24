// Package branding centralizes MobiAI's visual identity in the CLI —
// banners, taglines, color choices. Kept in a separate package so any
// command can pull it in without cyclic imports.
package branding

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// banner is the MOBIAI wordmark in block-style ASCII (figlet "ANSI
// Shadow"). All 6 rows are exactly 49 columns; letters M/O/B/I/A/I are
// composed from the canonical figlet glyphs joined by a single space,
// with a 3-space gap separating "MOBI" from "AI".
const banner = `███╗   ███╗  ██████╗  ██████╗  ██╗    █████╗  ██╗
████╗ ████║ ██╔═══██╗ ██╔══██╗ ██║   ██╔══██╗ ██║
██╔████╔██║ ██║   ██║ ██████╔╝ ██║   ███████║ ██║
██║╚██╔╝██║ ██║   ██║ ██╔══██╗ ██║   ██╔══██║ ██║
██║ ╚═╝ ██║ ╚██████╔╝ ██████╔╝ ██║   ██║  ██║ ██║
╚═╝     ╚═╝  ╚═════╝  ╚═════╝  ╚═╝   ╚═╝  ╚═╝ ╚═╝`

// tagline appears centered-ish under the wordmark. Short and stable —
// don't seasonally rewrite, it's part of the brand identity.
const tagline = "            AI FOR MOBILE DEVELOPERS"

// ANSI color codes used for the colorized variant. Bright cyan for
// the wordmark (reads well on both light and dark terminals), dim
// white for the tagline so it doesn't shout louder than the logo.
const (
	colorBannerStart  = "\033[1;36m" // bold cyan
	colorTaglineStart = "\033[2;37m" // dim white
	colorReset        = "\033[0m"
)

// Render returns the full banner string (wordmark + blank line +
// tagline + trailing newline). When useColor is true, wraps the parts
// in ANSI escape codes; otherwise returns plain UTF-8.
//
// Callers should pass false when:
//   - The --no-color flag is set.
//   - The NO_COLOR env var is non-empty (https://no-color.org/).
//   - Output is being piped/redirected (not a TTY).
//
// shouldUseColor() below covers the env / TTY checks; the --no-color
// flag check is the caller's job since this package has no view of
// cobra flags.
func Render(useColor bool) string {
	if useColor {
		return colorBannerStart + banner + colorReset + "\n\n" +
			colorTaglineStart + tagline + colorReset + "\n"
	}
	return banner + "\n\n" + tagline + "\n"
}

// Print writes the rendered banner to w, deciding color automatically
// based on noColorFlag (typically GlobalFlags.NoColor) combined with
// env-var and TTY checks. Convenience wrapper for the common case.
func Print(w io.Writer, noColorFlag bool) {
	useColor := !noColorFlag && shouldUseColor(w)
	fmt.Fprint(w, Render(useColor))
}

// shouldUseColor reports whether ANSI escapes are appropriate for w.
// Returns false when:
//   - NO_COLOR env var is set (any non-empty value, per the spec).
//   - w is a file but not a terminal (piped/redirected output).
//
// Conservative on the side of "no color" so CI logs and `cmd > file`
// stay clean.
func shouldUseColor(w io.Writer) bool {
	if v := strings.TrimSpace(os.Getenv("NO_COLOR")); v != "" {
		return false
	}
	// If the writer is *os.File, check whether it's a terminal. Other
	// writer types (bytes.Buffer in tests, log writers, etc.) we leave
	// to the caller's noColorFlag decision.
	if f, ok := w.(*os.File); ok {
		fi, err := f.Stat()
		if err != nil {
			return false
		}
		// Character device = terminal. Pipes and files don't have this
		// mode bit set.
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return true
}
