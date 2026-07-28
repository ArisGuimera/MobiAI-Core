package i18n

func init() {
	registerES(map[string]string{
		// ── cmd/mobiai/main.go: usage template headings ──────────────────
		"Usage:":                                "Uso:",
		"[command]":                             "[comando]",
		"Aliases:":                              "Alias:",
		"Examples:":                             "Ejemplos:",
		"Available commands:":                   "Comandos disponibles:",
		"Global flags:":                         "Flags globales:",
		"Additional topics:":                    "Temas adicionales:",
		"Use ":                                  "Usá ",
		"for more information about a command.": "para más información sobre un comando.",

		// ── cmd/mobiai/main.go: root + help command ──────────────────────
		"MobiAI CLI — manage the MobiAI ecosystem for AI-assisted mobile development": "MobiAI CLI — gestiona el ecosistema MobiAI para desarrollo móvil con IA",
		"MobiAI CLI is the unified tool of the MobiAI ecosystem: skills, agents, MCPs, and orchestration across clients for AI-assisted mobile development. Today it manages skills compatible with the agentskills.io standard on any supported client; agents and MCP servers are coming soon.": "MobiAI CLI es la herramienta unificada del ecosistema MobiAI: skills, agentes, MCPs y orquestación entre clientes para desarrollo móvil asistido por IA. Hoy gestiona skills compatibles con el standard agentskills.io en cualquier cliente compatible; próximamente sumará agentes y servidores MCP.",
		"Help about any command":         "Ayuda sobre cualquier comando",
		"Help about any mobiai command.": "Ayuda sobre cualquier comando de mobiai.",
		"Unknown command %q\n":           "Comando desconocido %q\n",
		"help for %s":                    "ayuda de %s",
		"mobiai version":                 "versión de mobiai",

		// ── internal/cmd/flags.go: global persistent flags ───────────────
		"force specific adapters (default: all detected)":               "fuerza adapters específicos (default: todos los detectados)",
		"assume yes on confirmations (CI-friendly)":                     "asume sí en confirmaciones (CI-friendly)",
		"more output (currently only affects 'mobiai update')":          "más output (hoy sólo afecta a 'mobiai update')",
		"disable ANSI colors in help output":                            "deshabilita colores ANSI en el help",
		"path to a local catalog (overrides ~/.mobiai/cache/catalog)":   "ruta a un catálogo local (override de ~/.mobiai/cache/catalog)",
		"include tier-3 adapters (speculative paths) in auto-detection": "incluir adapters tier-3 (paths especulativos) en la auto-detección",

		// ── internal/cmd/update.go: syncRemoteCatalog prints ─────────────
		"Syncing catalog (git pull)...\n": "Sincronizando catálogo (git pull)...\n",
		"Cloning catalog from %s...\n":    "Clonando catálogo desde %s...\n",
	})
}
