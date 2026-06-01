package i18n

// catalog_misc_es.go holds the Spanish translations for the misc command set:
// internal/cmd/doctor.go, catalog.go, hosts.go and status.go. Keys are
// byte-identical to the i18n.T("...") arguments at those call sites.
func init() {
	registerES(map[string]string{
		// ── internal/cmd/doctor.go ───────────────────────────────────────────
		"Thorough diagnostics for CLI, hosts and catalog": "Diagnóstico exhaustivo del CLI, hosts y catálogo",
		"MobiAI %s — diagnostics\n\n":                     "MobiAI %s — diagnóstico\n\n",
		"Supported hosts:":                                "Hosts soportados:",
		"Catalog:":                                        "Catálogo:",
		"  (never synced)":                                "  (nunca se sincronizó)",
		"  Last sync: %s\n":                               "  Última sincronización: %s\n",
		"  Version:   %s\n":                               "  Versión: %s\n",
		"Drift check:":                                    "Chequeo de drift:",
		"  ✓ no drift detected (Verify is a stub in this version)": "  ✓ sin drift detectado (Verify es stub en esta versión)",

		// ── internal/cmd/catalog.go ──────────────────────────────────────────
		"Inspect the catalog (debug)": "Inspeccionar el catálogo (debug)",
		"Hidden subcommand for catalog/resolver smoke tests. Not intended for end users.": "Subcomando oculto para smoke tests del catálogo/resolver. No es para usuarios finales.",
		"Path to the catalog root (directory containing .claude-plugin/marketplace.json)": "Ruta a la raíz del catálogo (el directorio que contiene .claude-plugin/marketplace.json)",
		"List the packs in the catalog": "Lista los packs del catálogo",
		"Catalog: %s (version %s)\n":    "Catálogo: %s (versión %s)\n",
		"none":                          "ninguna",
		"Resolve the install order for the given packs": "Resuelve el orden de instalación para los packs dados",
		"Install order:\n": "Orden de instalación:\n",

		// ── internal/cmd/hosts.go ────────────────────────────────────────────
		"Inspect supported hosts (debug)":                                               "Inspeccionar los hosts soportados (debug)",
		"Hidden subcommand for hosts-registry smoke tests. Not intended for end users.": "Subcomando oculto para smoke tests del registry de hosts. No es para usuarios finales.",
		"List supported hosts and which are detected":                                   "Lista los hosts soportados y cuáles están detectados",
		"Supported hosts (%d adapters loaded):\n":                                       "Hosts soportados (%d adapters cargados):\n",

		// ── shared: doctor.go + hosts.go ─────────────────────────────────────
		"not detected": "no detectado",

		// ── internal/cmd/status.go ───────────────────────────────────────────
		"Summary of detected hosts, installed packs and catalog state":                                                    "Resumen de hosts detectados, packs instalados y estado del catálogo",
		"  (none detected — install Claude Code, Cursor, Gemini CLI, Codex, or another agentskills.io-compatible client)": "  (ninguno detectado — instalá Claude Code, Cursor, Gemini CLI, Codex u otro cliente compatible con agentskills.io)",
		"Installed packs:": "Packs instalados:",
		"  (none)":         "  (ninguno)",
	})
}
