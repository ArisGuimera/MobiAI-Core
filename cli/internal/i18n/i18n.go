// Package i18n provides minimal gettext-style localization for the CLI.
//
// The English source string IS the catalog key. T("...") returns the Spanish
// translation when the active language is ES and a translation exists;
// otherwise it returns the English source (the fallback). For Printf-style
// calls, wrap the FORMAT string — the format verbs are preserved by the
// translation (enforced by TestCatalogVerbParity).
//
// Localization boundary (what gets wrapped in T): command output, cobra
// Short/Long/Use, flag usages, and usage-template headings. Go error strings
// (fmt.Errorf, %w-wrapped) are intentionally left in English and are NOT
// wrapped in T — translating partial error chains is out of scope.
package i18n

import (
	"os"
	"sync"
)

// Lang is a supported CLI language code.
type Lang string

const (
	EN Lang = "en"
	ES Lang = "es"
)

var (
	mu      sync.RWMutex
	current = EN
)

// Supported returns the languages the CLI can switch between, in display order.
func Supported() []Lang { return []Lang{EN, ES} }

// Parse maps a string to a supported Lang, reporting whether it was valid.
func Parse(s string) (Lang, bool) {
	switch Lang(s) {
	case EN:
		return EN, true
	case ES:
		return ES, true
	default:
		return EN, false
	}
}

// Current returns the active language.
func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// SetLang sets the active language for this process. It does NOT persist the
// choice — use internal/state.Config for that.
func SetLang(l Lang) {
	mu.Lock()
	current = l
	mu.Unlock()
}

// Init resolves the active language and sets it for this process. Precedence:
//
//	MOBIAI_LANG env (one-off override) > persisted preference > EN default.
//
// persisted is the saved preference (e.g. state.Config.Lang); pass "" if none.
// Unrecognized values are ignored (fall through to the next source). English
// is the default — Spanish is an explicit opt-in, never auto-detected from the
// OS locale.
func Init(persisted string) {
	lang := EN
	if l, ok := Parse(persisted); ok {
		lang = l
	}
	if env := os.Getenv("MOBIAI_LANG"); env != "" {
		if l, ok := Parse(env); ok {
			lang = l
		}
	}
	SetLang(lang)
}

// T translates an English source string to the active language (gettext model).
// Returns the English source unchanged when the active language is EN or when
// no translation exists for the current language.
func T(en string) string {
	mu.RLock()
	cur := current
	mu.RUnlock()
	if cur == EN {
		return en
	}
	// catalogES is built once at package init and never mutated, so it is
	// safe to read without holding the lock.
	if s, ok := catalogES[en]; ok {
		return s
	}
	return en
}
