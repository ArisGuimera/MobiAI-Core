package i18n

func init() {
	registerES(map[string]string{
		// ── internal/cmd/brain.go ────────────────────────────────────────────
		// root + brainCommonFlags
		"Per-project memory for mobile agents": "Memoria por proyecto para agentes mobile",
		"MobiAI Brain stores living per-project context: detected stack, decisions, bugfixes, testing patterns and integrations. Lives at <repo>/.mobiai/brain/ — separate from the CLI's global state.": "MobiAI Brain guarda contexto vivo por proyecto: stack detectado, decisiones, bugfixes, patrones de testing e integraciones. Vive en <repo>/.mobiai/brain/ — separado del estado global de la CLI.",
		"project path (default: detected from cwd)": "ruta del proyecto (default: detectado desde el cwd)",

		// init
		"Initialize .mobiai/brain in the current project":     "Inicializa .mobiai/brain en el proyecto actual",
		"⚠ %s (using %s)\n":                                   "⚠ %s (usando %s)\n",
		"Detected project: %s\n":                              "Proyecto detectado: %s\n",
		"✓ created %s\n":                                      "✓ creado %s\n",
		"= %s already exists (not overwritten)\n":             "= %s ya existe (no se sobrescribe)\n",
		"✓ memories initialized: %s\n":                        "✓ memorias inicializadas: %s\n",
		"= memories already exist (not overwritten)":          "= memorias ya existentes (no se sobrescriben)",
		"Next step: `mobiai brain scan` to detect the stack.": "Siguiente paso: `mobiai brain scan` para detectar el stack.",

		// scan
		"Scan the project and detect the mobile stack": "Escanea el proyecto y detecta el stack mobile",
		"Scanning %s ...\n":                            "Escaneando %s ...\n",
		"✓ scan saved to %s\n":                         "✓ scan guardado en %s\n",

		// context
		"Print the Brain's Markdown context (config + scan + memories)": "Imprime el contexto Markdown del Brain (config + scan + memorias)",
		"Prints the Brain context as Markdown. Without flags, dumps everything. With --section, --platform, --status or --area, filters memory entries so only what's relevant appears (useful as the brain grows).": "Imprime el contexto del Brain como Markdown. Sin flags, vuelca todo. Con --section, --platform, --status o --area filtra entradas de las memorias para que solo aparezca lo relevante (útil cuando el brain crece).",
		"limit sections to render (stack,rules,decisions,bugfixes,testing,integrations,releases,warnings). Repeatable or comma-separated.":                                                                           "limita las secciones a renderizar (stack,rules,decisions,bugfixes,testing,integrations,releases,warnings). Repetible o coma-separado.",
		"filter entries by platform (android|ios|shared|kmp|flutter|react-native)":                                                                                                                                   "filtra entradas por platform (android|ios|shared|kmp|flutter|react-native)",
		"filter entries by status (active|temporary|deprecated)":                                                                                                                                                     "filtra entradas por status (active|temporary|deprecated)",
		"filter entries whose area contains this string (substring)":                                                                                                                                                 "filtra entradas cuyo area contenga este string (substring)",

		// search
		"Free-text search across Brain memories": "Busca texto libre en las memorias del Brain",
		"Performs a case-insensitive match against the title and body of each memory entry. The --platform/--status/--area flags filter results with AND semantics.": "Hace match case-insensitive sobre el título y el cuerpo de cada entrada de las memorias. Los flags --platform/--status/--area filtran los resultados con semántica AND.",
		"limit to entries with this platform":              "limita a entradas con este platform",
		"limit to entries with this status":                "limita a entradas con este status",
		"limit to entries whose area contains this string": "limita a entradas cuya area contenga este string",
		"No results for %q.\n":                             "Sin resultados para %q.\n",
		"%d result(s) for %q:\n\n":                         "%d resultado(s) para %q:\n\n",

		// review
		"List temporary entries whose review_after has passed": "Lista entradas temporales cuyo review_after ya pasó",
		"Walks the memories and shows entries with status: temporary whose review_after <= today. Designed to keep temporary workarounds from drifting into permanence by inertia. By default exits with status 1 if there are overdue entries (useful as a CI / pre-commit gate); pass --no-fail to only report. With --include-no-date it also lists temporary entries without a review_after assigned.": "Recorre las memorias y muestra las entradas con status: temporary cuyo review_after <= hoy. Pensado para evitar que workarounds temporales se vuelvan permanentes por inercia. Por defecto sale con exit 1 si hay vencidas (útil como gate en CI / pre-commit); usá --no-fail para que solo informe. Con --include-no-date también lista las temporary sin review_after asignado.",
		"also list temporary entries without review_after (printed under a separate heading)": "también lista entradas temporary sin review_after (sección aparte en la salida)",
		"always exit 0, even if there are overdue entries (informational mode)":               "siempre exit 0, incluso si hay entradas vencidas (modo solo-informativo)",
		"✓ No overdue temporary entries.":                                                     "✓ No hay entradas temporales vencidas.",
		"⚠ %d overdue temporary entry(ies):\n\n":                                              "⚠ %d entrada(s) temporal(es) vencida(s):\n\n",
		"✓ No overdue entries.":                                                               "✓ No hay entradas vencidas.",
		"\n%d temporary entry(ies) without review_after:\n\n":                                 "\n%d entrada(s) temporal(es) sin review_after:\n\n",
		"    review_after: %s (overdue by %s)\n":                                              "    review_after: %s (vencido hace %s)\n",

		// promote
		"Change the status of an existing entry": "Cambia el status de una entrada existente",
		"Updates the status: of the entry with the given id (active|temporary|deprecated). Designed as the wrap-up flow after `brain review`: once you no longer need a temporary entry, promote it to active (it became permanent) or deprecated (no longer applies). Optionally also updates review_after in the same call, or removes it with --clear-review-after.": "Actualiza el status: de la entrada con el id dado (active|temporary|deprecated). Pensado como flujo de cierre tras `brain review`: cuando ya no necesitás una entrada temporary, la promovés a active (se volvió definitiva) o deprecated (ya no aplica). Opcionalmente también actualiza review_after en la misma llamada, o lo elimina con --clear-review-after.",
		"new status: active|temporary|deprecated (required)":          "nuevo status: active|temporary|deprecated (requerido)",
		"also update review_after to this YYYY-MM-DD date (optional)": "actualizar también review_after a esta fecha YYYY-MM-DD (opcional)",
		"remove review_after (incompatible with --review-after)":      "eliminar review_after (incompatible con --review-after)",

		// bump
		"Extend the review_after of an existing entry": "Extiende review_after de una entrada existente",
		"Updates review_after of the entry with the given id, without touching status. Useful after `brain review` when a temporary entry is still valid and you want to extend its review deadline. For status changes use `promote`.": "Actualiza review_after de la entrada con el id dado, sin tocar status. Útil tras `brain review` cuando una entrada temporary sigue siendo válida y querés extender su plazo de revisión. Para cambios de status usá `promote`.",
		"new YYYY-MM-DD date for review_after (required)": "nueva fecha YYYY-MM-DD para review_after (requerido)",

		// doBrainUpdate
		"✓ %s updated (%s)\n": "✓ %s actualizada (%s)\n",

		// save
		"Save an entry into the Brain's memories": "Guardar una entrada en las memorias del Brain",
		"Appends a structured entry to the matching memory file. Requires the Brain to be initialized in the project (run `mobiai brain init` first).": "Añade una entrada estructurada al archivo de memoria correspondiente. Requiere que el Brain esté inicializado en el proyecto (corré `mobiai brain init` primero).",
		"Save an architecture decision":                          "Guardar una decisión de arquitectura",
		"Save a bugfix or workaround":                            "Guardar un bugfix o workaround",
		"Save a reusable testing pattern":                        "Guardar un patrón de testing reusable",
		"short entry title (required)":                           "título corto de la entrada (requerido)",
		"android|ios|shared|kmp|flutter|react-native (optional)": "android|ios|shared|kmp|flutter|react-native (opcional)",
		"project area (free-form, optional)":                     "área del proyecto (libre, opcional)",
		"active|temporary|deprecated":                            "active|temporary|deprecated",
		"YYYY-MM-DD date to review (optional)":                   "fecha YYYY-MM-DD para revisar (opcional)",
		"Markdown body (if omitted, read from stdin)":            "cuerpo Markdown (si se omite, se lee de stdin)",
		"related files, comma-separated (optional)":              "archivos relacionados, separados por coma (opcional)",
		"✓ saved to %s\n  id: %s\n":                              "✓ guardado en %s\n  id: %s\n",

		// mcp
		"Start an MCP server that exposes the Brain as tools": "Arranca un servidor MCP que expone el Brain como tools",
		"Starts an MCP (Model Context Protocol) server that exposes the Brain's operations (context, search, scan, save) as tools the agent can invoke directly. Communicates over stdio — designed for an MCP client (Claude Code, Cursor, Copilot CLI, etc.) to launch it as a subprocess. See brain/MCP-SETUP.md.": "Arranca un servidor MCP (Model Context Protocol) que expone las operaciones del Brain (context, search, scan, save) como tools que el agente puede invocar directamente. Comunica por stdio — pensado para que un cliente MCP (Claude Code, Cursor, Copilot CLI, etc.) lo lance como subproceso. Ver brain/MCP-SETUP.md.",

		// install-mcp
		"Register the Brain MCP server in AI clients (Claude Code, Cursor)": "Registra el server MCP de Brain en clientes IA (Claude Code, Cursor)",
		"Adds the `mobiai-brain` server to each supported AI client's config, preserving the rest of the file. By default detects which clients are present (presence of ~/.claude or ~/.cursor); use --client to force a single one. Idempotent: re-running it with the same config is a no-op. Use --dry-run to preview without touching files, or --uninstall to remove the registration.": "Añade el server `mobiai-brain` al config de cada cliente IA soportado, preservando el resto del archivo. Por defecto detecta los clientes presentes (presencia de ~/.claude o ~/.cursor); con --client podés forzar uno solo. Idempotente: re-correrlo con la misma config no hace nada. Usá --dry-run para previsualizar sin tocar archivos, o --uninstall para quitar el registro.",
		"clients to register: claude|cursor (repeatable or comma-separated; default: all detected)": "clientes a registrar: claude|cursor (repetible o coma-separado; default: todos los detectados)",
		"show which files would be touched without writing anything":                                "muestra qué archivos se tocarían sin escribir nada",
		"remove the registration instead of creating it":                                            "elimina el registro en lugar de crearlo",
		"path to the mobiai binary to register (default: the binary in use)":                        "ruta al binario mobiai a registrar (default: el binario en uso)",

		// printInstallResults
		"Done. Restart the client for changes to take effect.": "Listo. Reiniciá el cliente para que tome efecto.",
		"Done. Restart the client so it loads the MCP server.": "Listo. Reiniciá el cliente para que cargue el server MCP.",
	})
}
