package branding

import (
	"os"
	"strings"
)

// ANSI escape sequences for the cobra help/usage templates. Bold cyan reads
// well on both light and dark terminals; dim works as a secondary tier; the
// reset is always applied via wrap() so we never leak escapes downstream.
const (
	ansiReset    = "\033[0m"
	ansiBold     = "\033[1m"
	ansiDim      = "\033[2m"
	ansiCyan     = "\033[36m"
	ansiBoldCyan = "\033[1;36m"
)

// noColorOverride lets the root command force-disable color (used by tests
// that want deterministic output). When false (the default), shouldStyle
// scans os.Args for --no-color itself — cobra's PersistentPreRun doesn't
// run for non-runnable command groups like `mobiai skills`, so we can't
// rely on a hook to wire the flag in time for the help template.
var noColorOverride bool

// SetNoColor lets callers force-disable color regardless of flag parsing.
// Mainly for tests; runtime users should pass --no-color or set NO_COLOR.
func SetNoColor(noColor bool) { noColorOverride = noColor }

// shouldStyle decides whether to emit ANSI for cobra help. Disabling
// signals (any one wins):
//   - SetNoColor(true) was called
//   - --no-color present anywhere in os.Args
//   - NO_COLOR env is non-empty (https://no-color.org/)
//   - os.Stdout is not a character device (pipe/file redirect)
//
// Override: FORCE_COLOR=1 bypasses the TTY check so you can pipe the
// styled output into a pager that handles ANSI (less -R, etc).
func shouldStyle() bool {
	if noColorOverride {
		return false
	}
	for _, a := range os.Args {
		if a == "--no-color" {
			return false
		}
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	if strings.TrimSpace(os.Getenv("FORCE_COLOR")) != "" {
		return true
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// StyleHeading wraps section titles ("Uso:", "Comandos disponibles:") in
// bold cyan. Exposed to cobra templates as the `head` func.
func StyleHeading(s string) string { return wrap(ansiBoldCyan, s) }

// StyleCmd wraps command names and other primary tokens in plain cyan.
// Exposed to cobra templates as the `styleCmd` func.
func StyleCmd(s string) string { return wrap(ansiCyan, s) }

// StyleDim is for secondary, lower-emphasis tokens (e.g., hints).
// Exposed to cobra templates as the `dim` func.
func StyleDim(s string) string { return wrap(ansiDim, s) }

// StyleBold is for occasional emphasis without a color shift.
func StyleBold(s string) string { return wrap(ansiBold, s) }

func wrap(prefix, s string) string {
	if !shouldStyle() {
		return s
	}
	return prefix + s + ansiReset
}
