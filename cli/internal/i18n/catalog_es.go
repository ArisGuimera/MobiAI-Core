package i18n

// catalogES maps English source strings (the T() keys, taken verbatim from the
// call sites) to their Spanish translations. Keys MUST be byte-identical to the
// i18n.T("...") arguments in the code — TestCatalogCompleteness enforces that
// every wrapped string has an entry here (or sits on intentionallyEnglish), and
// TestCatalogVerbParity enforces that the %-verbs match between en and es.
//
// English strings absent from this map fall back to English at runtime.
var catalogES = map[string]string{
	// ── internal/cmd/update.go ───────────────────────────────────────────
	"Refresh the catalog from the remote (or a local path with --catalog-root)":               "Refrescar el catálogo desde el remoto (o un local con --catalog-root)",
	"Skills catalog updated to v%s.\n":                                                        "Catálogo de skills actualizado a v%s.\n",
	"%d packs available at %s.\n":                                                             "%d packs disponibles en %s.\n",
	"mobiai binary: %s (--skip-binary, did not check for an update).\n":                       "binario mobiai: %s (--skip-binary, no se verificó si hay update).\n",
	"Warning: the catalog was updated, but the binary could not be updated: %v\n":             "Aviso: el catálogo se actualizó, pero no pude actualizar el binario: %v\n",
	"Retry `mobiai update` later, or reinstall with the install script: https://mobiai.dev\n": "Reintentá `mobiai update` más tarde, o reinstalá con el install script: https://mobiai.dev\n",
	"path to a local catalog (override)":                                                      "ruta a un catálogo local (override)",
	"if the cache dir exists but is not a git repo, delete it and re-clone":                   "si el directorio cache existe sin ser un repo git, borralo y cloná de nuevo",
	"only query GitHub releases and cache the result (does not touch the catalog)":            "solo consultar GitHub releases y cachear el resultado (no toca el catálogo)",
	"print nothing on exit (intended for running from a background hook)":                     "no imprime nada al salir (pensado para correr desde un hook en background)",
	"do not update the mobiai binary, only refresh the catalog":                               "no actualizar el binario mobiai, solo refrescar el catálogo",

	// ── internal/cmd/updatecheck.go ──────────────────────────────────────
	"MobiAI %s → %s available. Run: mobiai update\n":  "MobiAI %s → %s disponible. Actualizá con: mobiai update\n",
	"MobiAI %s is up to date (latest release: %s).\n": "MobiAI %s está al día (último release: %s).\n",

	// ── internal/cmd/selfupdate.go ───────────────────────────────────────
	"mobiai binary updated to %s. Restart your terminal (or re-run mobiai) to use the new version.\n": "binario mobiai actualizado a %s. Reiniciá la terminal (o volvé a correr mobiai) para usar la nueva versión.\n",
	"mobiai binary %s: up to date.\n":            "binario mobiai %s: al día.\n",
	"mobiai binary %s → %s: downloading %s...\n": "binario mobiai %s → %s: descargando %s...\n",

	// ── internal/cmd/lang.go ─────────────────────────────────────────────
	"Show or set the CLI language (en/es)": "Mostrar o cambiar el idioma del CLI (en/es)",
	"Current language: %s\n":               "Idioma actual: %s\n",
	"Language set to %s.\n":                "Idioma cambiado a %s.\n",
}

// intentionallyEnglish lists wrapped strings that have no Spanish translation on
// purpose (e.g. brand names, identifiers). Entries here satisfy
// TestCatalogCompleteness without a catalogES mapping. Empty for now.
var intentionallyEnglish = map[string]bool{}

// registerES merges additional en→es pairs into catalogES. It lets each command
// keep its translations in its own catalog_<cmd>_es.go file (registered from an
// init func) instead of one giant map. Package-var initialization runs before
// any init func, so the base catalogES literal above is always populated first;
// last writer wins on duplicate keys.
func registerES(m map[string]string) {
	for k, v := range m {
		catalogES[k] = v
	}
}
