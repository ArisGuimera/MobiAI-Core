---
name: mobiai-update
description: Actualiza el binario `mobiai` a la última versión publicada en GitHub Releases. Usá esta skill cuando el banner de SessionStart muestre "MobiAI update available" o cuando el usuario pida explícitamente actualizar MobiAI.
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
---

## Objetivo

Actualizar el binario `mobiai` instalado en el sistema del usuario corriendo el install script oficial. El install script detecta la plataforma, descarga el binario correcto desde GitHub Releases, y reemplaza el binario existente en `~/.mobiai/bin/`.

## Workflow

### 1. Confirmá la versión actual

```bash
mobiai --version
```

Anotá la versión que sale (ej: `mobiai 0.1.1`).

### 2. Mostrá al usuario qué vas a hacer

Antes de tocar nada, decile al usuario:

> "Voy a actualizar MobiAI corriendo el install script oficial. Esto descarga el último release desde GitHub y reemplaza `~/.mobiai/bin/mobiai`. ¿Procedo?"

Esperá confirmación. **No actualices sin que el usuario diga sí.**

### 3. Corré el install script según la plataforma

**Windows (PowerShell):**

```powershell
iwr -useb https://mobiai.dev/install.ps1 | iex
```

**macOS / Linux:**

```bash
curl -fsSL https://mobiai.dev/install.sh | sh
```

Si el dominio `mobiai.dev` no resuelve (todavía no está configurado), usá el endpoint directo:

```bash
curl -fsSL https://raw.githubusercontent.com/ArisGuimera/MobiAI-Core/main/scripts/install.sh | sh
```

### 4. Verificá

```bash
mobiai --version
```

Confirmá que la versión cambió. Si quedó igual, algo falló y hay que diagnosticar (ver "Troubleshooting" abajo).

### 5. Refrescá el cache de update-check

Para que el banner deje de aparecer en futuras sesiones:

```bash
mobiai update --check --silent
```

Esto regenera `update-check.json` con la nueva versión instalada.

## Troubleshooting

**El binario no se reemplaza en Windows:**
Posible que `mobiai.exe` esté en uso (otra terminal lo tiene abierto). Cerrá todas las terminales y reintentá. Si persiste:

```powershell
Get-Process mobiai -ErrorAction SilentlyContinue | Stop-Process -Force
```

**`mobiai: command not found` después del install:**
El install script lo deja en `$HOME/.mobiai/bin/`. Confirmá que ese dir esté en `PATH`:

```bash
echo "$PATH" | tr ':' '\n' | grep mobiai
```

Si no aparece, agregalo a tu `~/.bashrc` / `~/.zshrc` / perfil de PowerShell.

**El install script falla con "no se detectó última versión":**
Pasá la versión manualmente:

```bash
MOBIAI_VERSION=0.1.2 curl -fsSL https://raw.githubusercontent.com/ArisGuimera/MobiAI-Core/main/scripts/install.sh | sh
```

## Notas

- El install script no requiere permisos de admin (instala en el home del usuario).
- Reinstalar la misma versión es seguro: sobrescribe el binario existente.
- El cache de skills (`~/.mobiai/cache/`) y el catálogo persisten entre updates.
