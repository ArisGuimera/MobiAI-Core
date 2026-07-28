package i18n

// Translations for helper functions that return user-facing fragments embedded
// in larger lines (relative-time, overdue labels, install-action labels, scan
// summary). These are kept here because they were left unwrapped by the
// per-command localization pass (they are return values / closures, not direct
// print sites) and would otherwise leak English into Spanish output.
func init() {
	registerES(map[string]string{
		// ── internal/cmd/graph.go: humanizeAge ───────────────────────────
		"%dm ago": "hace %dm",
		"%dh ago": "hace %dh",
		"%dd ago": "hace %dd",

		// ── internal/cmd/brain.go: daysOverdueLabel ──────────────────────
		"today":   "hoy",
		"1 day":   "1 día",
		"%d days": "%d días",

		// ── internal/cmd/brain.go: installIconAndLabel ───────────────────
		"registered":                     "registrado",
		"updated":                        "actualizado",
		"unchanged (already registered)": "sin cambios (ya estaba registrado)",
		"removed":                        "eliminado",
		"was not registered":             "no estaba registrado",
		"skipped (client not installed)": "saltado (cliente no instalado)",

		// ── internal/cmd/brain.go: printScanSummary ──────────────────────
		"\nSummary:\n":        "\nResumen:\n",
		"  Type:        %s\n": "  Tipo:        %s\n",
		"  Platforms:   %s\n": "  Plataformas: %s\n",
		"  Warnings:    %d\n": "  Advertencias: %d\n",
	})
}
