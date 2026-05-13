# MobiAI Brain — MCP server

`mobiai brain mcp` arranca un servidor [MCP](https://modelcontextprotocol.io) que expone el Brain como **tools** para que tu cliente de IA (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI, ...) las invoque directamente.

Sin MCP, el agente solo conoce el Brain si la skill `mobiai-brain` se lo dice — y aun así tiene que ejecutar comandos CLI. Con MCP el Brain pasa a estar siempre en su toolbox y se invoca como cualquier otra tool.

## Tools expuestas

| Tool | Equivalente CLI | Salida |
|---|---|---|
| `mobile_context` | `mobiai brain context` (con filtros) | Markdown completo del contexto |
| `mobile_search` | `mobiai brain search` | JSON estructurado `{hits: [{section, title, id, status, platform, area, snippet}]}` |
| `mobile_scan` | `mobiai brain scan` | JSON del stack detectado (project_type, platforms, libs, integraciones, ...) |
| `mobile_review` | `mobiai brain review` | JSON `{overdue: [...], no_date: [...]}` con entradas `temporary` cuyo `review_after` ya pasó |
| `mobile_save_decision` | `mobiai brain save decision` | `{id, section}` de la entrada creada |
| `mobile_save_bugfix` | `mobiai brain save bugfix` | `{id, section}` |
| `mobile_save_testing` | `mobiai brain save testing` | `{id, section}` |

Todas requieren un brain inicializado (`mobiai brain init`). Si no existe, la tool falla con un error claro indicando cómo arreglar.

## Registro en cada cliente

La invocación es siempre la misma — solo cambia dónde se declara el server. El binario `mobiai` tiene que estar en `$PATH` (o pasalo con ruta absoluta).

### Claude Code

Edita `~/.claude/settings.json` (o el archivo de settings del proyecto en `.claude/settings.json`):

```json
{
  "mcpServers": {
    "mobiai-brain": {
      "command": "mobiai",
      "args": ["brain", "mcp"]
    }
  }
}
```

Si querés que el server se ata a un proyecto específico independiente del cwd:

```json
{
  "mcpServers": {
    "mobiai-brain": {
      "command": "mobiai",
      "args": ["brain", "mcp", "--root", "/ruta/absoluta/al/proyecto"]
    }
  }
}
```

Reiniciá Claude Code para que recargue la config. En la sesión, las tools `mobile_*` aparecen disponibles.

### Cursor

`~/.cursor/mcp.json` (o `.cursor/mcp.json` en el proyecto):

```json
{
  "mcpServers": {
    "mobiai-brain": {
      "command": "mobiai",
      "args": ["brain", "mcp"]
    }
  }
}
```

### Copilot CLI

`~/.copilot/mcp-config.json`:

```json
{
  "mcpServers": {
    "mobiai-brain": {
      "type": "local",
      "command": "mobiai",
      "args": ["brain", "mcp"],
      "tools": ["*"]
    }
  }
}
```

### Codex

Codex usa TOML. En `~/.codex/config.toml`:

```toml
[mcp_servers.mobiai-brain]
command = "mobiai"
args = ["brain", "mcp"]
```

### Gemini CLI

`~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "mobiai-brain": {
      "command": "mobiai",
      "args": ["brain", "mcp"]
    }
  }
}
```

## Detección del proyecto

Cuando el cliente lanza `mobiai brain mcp` como subproceso, el cwd suele ser el del proyecto abierto. El server usa `brain.FindProjectRoot(cwd)` (la misma lógica que el resto de la CLI) y captura el root **una vez al arrancar**. No cambia entre tool calls.

Override explícito con `--root <ruta>` si necesitás que el server siempre apunte a un proyecto fijo.

## Comportamiento

- **Mismas garantías que la CLI**: archivos sensibles nunca se leen, escrituras atómicas, idempotencia de `init`.
- **Errores claros**: si el brain no está inicializado, todas las tools fallan con `brain no inicializado en <root> — corré 'mobiai brain init' primero`.
- **Versión**: el server reporta el `version` del binario (`mobiai --version`).
- **Transporte**: solo stdio. HTTP queda fuera de scope por ahora — no hay razón para que el Brain sea remoto.

## Diferencia con las skills

Las skills (`mobiai-fix-issue`, `mobiai-write-tests`, etc.) **siguen funcionando** con la CLI shell-based aunque haya MCP configurado. Las tools MCP son una alternativa de invocación, no un reemplazo.

A medida que iteremos podemos migrar los hooks de las skills a invocar las tools MCP directamente (más limpio que shellear out al binario). Por ahora la coexistencia es deliberada — la CLI sigue siendo el fallback universal.
