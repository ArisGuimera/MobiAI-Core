package i18n

func init() {
	registerES(map[string]string{
		// ── internal/cmd/graph.go ────────────────────────────────────────────
		"Semantic exploration of mobile code": "Exploración semántica del código mobile",
		"MobiAI Graph indexes the project and lets you search symbols, find references, and get relevant context for a task. Lives at <repo>/.mobiai/graph/.": "MobiAI Graph indexa el proyecto y permite buscar símbolos, encontrar referencias y obtener contexto relevante para una tarea. Vive en <repo>/.mobiai/graph/.",
		"project path (default: cwd)":                         "ruta del proyecto (default: cwd)",
		"Index the project and save .mobiai/graph/index.json": "Indexa el proyecto y guarda .mobiai/graph/index.json",
		"✓ Index generated: %s\n":                             "✓ Índice generado: %s\n",
		"  Files:   %d\n":                                     "  Archivos: %d\n",
		"  Symbols: %d\n":                                     "  Símbolos: %d\n",
		"  Kotlin:  %d\n":                                     "  Kotlin: %d\n",
		"  Swift:   %d\n":                                     "  Swift: %d\n",
		"Show info about the current index":                   "Muestra info del índice actual",
		"Index:     %s\n":                                     "Índice: %s\n",
		"Generated: %s (%s)\n":                                "Generado: %s (%s)\n",
		"Files:     %d\n":                                     "Archivos: %d\n",
		"Symbols:   %d\n":                                     "Símbolos: %d\n",
		"Kotlin:    %d\n":                                     "Kotlin: %d\n",
		"Swift:     %d\n":                                     "Swift: %d\n",
		"Search symbols by name":                              "Busca símbolos por nombre",
		"max results to show (0 = no limit)":                  "máximo de resultados a mostrar (0 = sin límite)",
		"filter by symbol kind (class, fun, object, interface, struct, enum, protocol, actor, func, extension)": "filtrar por tipo de símbolo (class, fun, object, interface, struct, enum, protocol, actor, func, extension)",
		"No matches.": "Sin coincidencias.",
		"… (showing %d, use --limit 0 to see all)\n": "… (mostrando %d, usá --limit 0 para ver todos)\n",
		"List textual references to the symbol":      "Lista referencias textuales del símbolo",
		"No references.":                             "Sin referencias.",
		"Relevant files for a task (heuristic)":      "Archivos relevantes para una tarea (heurística)",
		"No relevant files found.":                   "Sin archivos relevantes encontrados.",
	})
}
