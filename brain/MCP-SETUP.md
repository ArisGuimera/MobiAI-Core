# MobiAI Brain — MCP server

`mobiai brain mcp` arranca un servidor [MCP](https://modelcontextprotocol.io) que expone el Brain como **tools** para que tu cliente de IA (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI, ...) las invoque directamente.

Sin MCP, el agente solo conoce el Brain si la skill `mobiai-brain` se lo dice — y aun así tiene que ejecutar comandos CLI. Con MCP el Brain pasa a estar siempre en su toolbox y se invoca como cualquier otra tool.

## Tools expuestas

| Tool | Equivalente CLI | Salida |
|---|---|---|
| `mobile_context` | `mobiai brain context` (con filtros) | Markdown del contexto: completo hasta 25 KB, con índice por encima |
| `mobile_get` | `mobiai brain show` | Markdown de una nota + `{id, section, path, markdown}` |
| `mobile_search` | `mobiai brain search` | JSON estructurado `{hits: [{section, title, id, status, platform, area, snippet, path}]}`; acepta `crash` para buscar por firma de crash |
| `mobile_scan` | `mobiai brain scan` | JSON del stack detectado (project_type, platforms, libs, integraciones, ...) |
| `mobile_review` | `mobiai brain review` | JSON `{overdue: [...], no_date: [...], stale: [...]}`: temporales vencidas y notas que pueden estar desactualizadas (⚠) |
| `mobile_verify` | `mobiai brain verify` | `{section, file, title, prev_verified, new_verified, files}` |
| `mobile_promote` | `mobiai brain promote` | `{section, file, title, prev_status, new_status, prev_review_after, new_review_after}` |
| `mobile_bump` | `mobiai brain bump` | Mismo schema que `mobile_promote`; bumpea solo `review_after` |
| `mobile_save_decision` | `mobiai brain save decision` | `{id, section, path}` de la nota creada |
| `mobile_save_bugfix` | `mobiai brain save bugfix` | `{id, section, path}` |
| `mobile_save_testing` | `mobiai brain save testing` | `{id, section, path}` |
| `mobile_save_knowledge` | `mobiai brain save knowledge` | `{id, section, path}` |

Las tools `mobile_save_*` aceptan, además de los campos de siempre, `supersedes` (ids que esta nota reemplaza), `until` (`<dependencia>@<versión>`), `new` (guardar aunque haya un título parecido) y, solo en bugfix, `crash` (firmas de crash).

Todas requieren un brain inicializado (`mobiai brain init`). Si no existe, la tool falla con un error claro indicando cómo arreglar. Todas, menos la del hook, migran un Brain del formato 0.2 la primera vez que lo tocan y lo avisan al principio de su salida.

## Registro automático (recomendado)

```bash
mobiai brain install-mcp                  # detecta clientes presentes y registra
mobiai brain install-mcp --client claude  # registro específico
mobiai brain install-mcp --dry-run        # previsualiza sin tocar archivos
mobiai brain install-mcp --uninstall      # quita el registro
```

El comando:

- **Detecta** clientes presentes (ahora mismo: Claude Code y Cursor) buscando `~/.claude/` y `~/.cursor/`. Si un cliente no está instalado, lo salta sin error.
- **Preserva** el resto del archivo de config — solo añade/quita el bloque `mcpServers.mobiai-brain`.
- **Idempotente**: re-correrlo no duplica nada.
- **Atomic** (temp+rename): un crash a media escritura nunca deja el config corrupto.
- **Usa el path absoluto** del binario en uso (vía `os.Executable()`) para que el cliente AI lo encuentre incluso si su `$PATH` es distinto al del shell. Override con `--binary <ruta>`.

Tras correrlo, reiniciá el cliente para que cargue el server MCP. **Validá** con `mobiai brain mcp` (debería arrancar sin errores; salí con Ctrl+C — el cliente lo lanzará como subproceso después).

Si tu cliente no es Claude Code ni Cursor todavía, seguí con el registro manual abajo. Vamos a ir sumando soporte (Copilot CLI, Codex, Gemini CLI) en próximas versiones.

## Registro manual (fallback)

Si preferís editar a mano, o tu cliente todavía no está soportado por `install-mcp`, los pasos son los mismos para todos: añadir un bloque `mcpServers` que apunta a `mobiai brain mcp`. Solo cambia dónde se declara el server. El binario `mobiai` tiene que estar en `$PATH` (o pasalo con ruta absoluta).

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
