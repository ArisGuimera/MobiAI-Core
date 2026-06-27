package i18n

// catalog_skills_es registers the Spanish translations for internal/cmd/skills.go.
// Keys MUST be byte-identical to the i18n.T("...") arguments at the call sites.
func init() {
	registerES(map[string]string{
		"Manage MobiAI skills":                            "Gestionar skills de MobiAI",
		"force specific adapters (default: all detected)": "fuerza adapters específicos (default: todos los detectados)",
		"assume yes on confirmations":                     "asume sí en confirmaciones",
		"path to a local catalog":                         "ruta a un catálogo local",
		"Install packs into detected hosts":               "Instalar packs en los hosts detectados",
		"Uninstall packs from detected hosts":             "Desinstalar packs de los hosts detectados",
		"List installed packs":                            "Listar packs instalados",
		"No packs installed.":                             "No hay packs instalados.",
		"Interactive picker to install/uninstall skills":  "Selector interactivo para instalar/desinstalar skills",
		"Launches the TUI picker to choose skill packs and apply changes to detected clients (Claude Code, Cursor, Codex, etc).": "Lanza el picker TUI para elegir packs de skills y aplicar los cambios sobre los clientes detectados (Claude Code, Cursor, Codex, etc).",
		"Cancelled.":                    "Cancelado.",
		"Done.":                         "Listo.",
		"Install plan:":                 "Plan de instalación:",
		"Continue? [y/N]: ":             "¿Continuar? [y/N]: ",
		"  Community skills (%d): %s\n": "  Skills de la comunidad (%d): %s\n",
	})
}
