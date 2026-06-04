package i18n

func init() {
	registerES(map[string]string{
		// ── internal/tui/tui.go ──────────────────────────────────────────────
		"No AI client detected.\nInstall Claude Code, Cursor, Gemini CLI or Codex and run `mobiai` again.\n": "No detecté ningún cliente de IA.\nInstalá Claude Code, Cursor, Gemini CLI o Codex y volvé a correr `mobiai`.\n",

		// ── internal/tui/view.go ─────────────────────────────────────────────
		"Detected hosts (%d): %s": "Hosts detectados (%d): %s",
		"↑↓ navigate  [space] change action  [enter] apply  [q] quit":   "↑↓ navegar  [espacio] cambiar acción  [enter] aplicar  [q] salir",
		"↑↓ navigate  [space] change action  [enter] apply  [esc] back": "↑↓ navegar  [espacio] cambiar acción  [enter] aplicar  [esc] volver",
		"→ pick skills":                                 "→ elegí skills",
		"Community skills":                              "Skills de la comunidad",
		"No community skills available yet.":            "Todavía no hay skills de la comunidad.",
		"[esc] back":                                    "[esc] volver",
		"community (%d skills)":                         "community (%d skills)",
		"Confirm changes":                               "Confirmar cambios",
		"No pending changes.\n":                         "No hay cambios pendientes.\n",
		"To install:\n":                                 "A instalar:\n",
		"To uninstall:\n":                               "A desinstalar:\n",
		"\nOn %d detected hosts.\n\n":                   "\nEn %d hosts detectados.\n\n",
		"[y] confirm  [n] back to picker  [esc] cancel": "[y] confirmar  [n] volver al picker  [esc] cancelar",
		"Applying changes... %d/%d":                     "Aplicando cambios... %d/%d",
		"✓ Done":                                        "✓ Listo",
		"⚠ Finished with %d errors\n\n":                 "⚠ Terminado con %d errores\n\n",
		"\n⚠ Could not save installed.json: %v\n":       "\n⚠ No pude guardar installed.json: %v\n",
		"\n[enter] close":                               "\n[enter] cerrar",
	})
}
