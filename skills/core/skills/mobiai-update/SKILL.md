---
name: mobiai-update
description: Actualiza el binario `mobiai` a la última versión publicada en GitHub Releases. Usá esta skill cuando el banner de SessionStart muestre "MobiAI update available" o cuando el usuario pida explícitamente actualizar MobiAI.
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
---

## Objetivo

Actualizar el binario `mobiai` instalado en el sistema del usuario. La vía
canónica es `mobiai update`: desde la v0.2.1 refresca el catálogo de skills **y**
se auto-actualiza el binario (descarga el asset de la plataforma desde GitHub
Releases, verifica su checksum SHA256, y reemplaza `~/.mobiai/bin/mobiai` con el
truco rename-then-write — funciona incluso en Windows con el `.exe` en uso). El
install script queda como fallback para cuando el auto-update falla.

## Workflow

### 1. Confirmá la versión actual

```bash
mobiai --version
```

Anotá la versión (ej: `mobiai 0.2.0`).

### 2. Mostrá al usuario qué vas a hacer

Antes de tocar nada, decile al usuario:

> "Voy a actualizar MobiAI con `mobiai update`. Esto refresca el catálogo de
> skills y, si hay binario nuevo, descarga el release y reemplaza
> `~/.mobiai/bin/mobiai`. ¿Procedo?"

Esperá confirmación. **No actualices sin que el usuario diga sí.**

### 3. Corré `mobiai update`

```bash
mobiai update
```

Refresca el catálogo y, si hay una versión más nueva, se auto-actualiza el
binario. Vas a ver algo como:

```
Catálogo de skills actualizado a vX.Y.Z.
N packs disponibles en ...
Binario mobiai 0.2.0 → 0.2.1: descargando mobiai-0.2.1-<os>-<arch>...
Binario mobiai actualizado a 0.2.1. Reiniciá la terminal (o volvé a correr mobiai) para usar la nueva versión.
```

Si solo querés refrescar el catálogo sin tocar el binario: `mobiai update --skip-binary`.

### 4. Verificá

```bash
mobiai --version
```

Confirmá que la versión cambió. En Windows el `.exe` viejo queda como
`mobiai.exe.old` hasta el próximo update (que lo limpia en su primer paso); es
esperado, no hace falta borrarlo a mano.

### 5. Fallback: install script

Solo si `mobiai update` mostró un **aviso de que no pudo actualizar el binario**
(red, permisos, o instalado en una ruta de solo lectura), corré el install
script oficial.

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

## Troubleshooting

**El binario no se reemplaza en Windows:**
`mobiai update` usa rename-then-write, así que el `.exe` en uso no debería
bloquear el update. Si aun así falla, puede haber otro proceso `mobiai`
corriendo o falta permiso de escritura en `~/.mobiai/bin`. Cerrá otras
terminales y reintentá. Si persiste:

```powershell
Get-Process mobiai -ErrorAction SilentlyContinue | Stop-Process -Force
```

**`mobiai: command not found` después del install:**
El binario vive en `$HOME/.mobiai/bin/`. Confirmá que ese dir esté en `PATH`:

```bash
echo "$PATH" | tr ':' '\n' | grep mobiai
```

Si no aparece, agregalo a tu `~/.bashrc` / `~/.zshrc` / perfil de PowerShell.

**El update aborta con "no encontré checksum" o "checksum no coincide":**
Es una protección a propósito: `mobiai update` nunca instala un binario que no
matchea el SHA256 publicado en `checksums.txt`. Suele ser una descarga corrupta
o parcial — reintentá más tarde, o usá el install script como fallback.

**El install script falla con "no se detectó última versión":**
Pasá la versión manualmente:

```bash
MOBIAI_VERSION=0.2.1 curl -fsSL https://raw.githubusercontent.com/ArisGuimera/MobiAI-Core/main/scripts/install.sh | sh
```

## Notas

- `mobiai update` no requiere permisos de admin (escribe en el home del usuario).
- Reinstalar la misma versión es seguro: sobrescribe el binario existente.
- El cache de skills (`~/.mobiai/cache/`) y el catálogo persisten entre updates.
- Tras un update exitoso, el banner de "update available" se apaga solo en la
  próxima sesión: el hook corre el binario nuevo y reescribe `update-check.json`.
