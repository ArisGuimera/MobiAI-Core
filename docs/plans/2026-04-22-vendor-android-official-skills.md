# Vendor `android/skills` into `plugins/android/` Plan

> **For agentic workers:** Use `mobile-subagent-development` (recommended) or `mobile-executing-plans` to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Eliminar el plugin separado `android-official-skills` del marketplace. Vendorear los 6 SKILL.md de `github.com/android/skills` (Apache 2.0) dentro de `plugins/android/skills/google/`, manteniendo auto-sync vía GitHub Action semanal. Single plugin `android` contiene todo.

**Architecture:** Cambio de modelo "externa como dependencia" a "externa vendoreada". El workflow `sync-android-skills.yml` pasa de editar un array JSON a copiar archivos al repo. Apache 2.0 requiere preservar `LICENSE.txt` y atribuir; lo cumplimos con `plugins/android/skills/google/LICENSE` + `NOTICE.md` en raíz.

**Tech Stack:** GitHub Actions (workflow refactoreado), jq para JSON edits, git para commits automáticos, bash para copia de archivos, Apache 2.0 compliance.

**Platform:** Infrastructure (cambio de empaquetado del marketplace, afecta todas las plataformas indirectamente pero el contenido vive bajo `android`).

---

## Motivación

Usuario reporta problema de UX: `/plugin list` muestra `android` y `android-official-skills` como dos entradas, y la gente que no lee la description no sabe cuál instalar. Además `android-official-skills` aparece en el autocomplete de `/plugin install` y es técnicamente válido instalarlo directo (aunque no debería — sólo se usa como dep automática de `android`).

Vendoreando el contenido de Google dentro de nuestro plugin `android`, el marketplace queda con **un único plugin por plataforma** (`android`, `ios`, `kmp`, `flutter`, `react-native` + `mobile` meta + `core` interno). El mismo patrón escala a futuras fuentes externas (performance samples, compose samples, lo que venga).

Verificación legal: Apache 2.0 §4 permite redistribución incluyendo modificada, con condiciones:
1. Incluir copia de la License → preservamos `LICENSE.txt` upstream en `plugins/android/skills/google/LICENSE`
2. Atribuir cambios → no modificamos SKILL.md, sólo copiamos; documentamos en NOTICE
3. Preservar copyright headers → el workflow copia verbatim
4. Incluir `NOTICE` si upstream tiene → verificamos al import que no hay (ya lo verificamos antes, puede cambiar)

§6 (marcas) nos impide dar a entender endorsement. Lo evitamos con lenguaje factual en README: "MobiAI includes official Android skills maintained by Google (Apache 2.0)".

---

## Estado actual vs objetivo

**Actual:**
- `.claude-plugin/marketplace.json` lista 8 plugins incluyendo `android-official-skills` con `source: url https://github.com/android/skills.git` y `strict: false`
- `plugins/android/.claude-plugin/plugin.json` tiene `"dependencies": ["core", "android-official-skills"]`
- `.github/workflows/sync-android-skills.yml` detecta nuevas carpetas raíz en Google y abre PR editando el array `skills` del plugin entry
- El cache local en `~/.claude/plugins/cache/mobiai/android-official-skills/2.0.0/` tiene el contenido clonado

**Objetivo:**
- Marketplace lista 7 plugins (removida `android-official-skills`)
- `android/plugin.json` tiene sólo `"dependencies": ["core"]`
- `plugins/android/skills/google/` contiene las 6 skills vendoreadas + LICENSE + README
- `NOTICE.md` nuevo en raíz listando software de terceros
- `README.md` menciona la integración
- Workflow refactoreado: clona android/skills → copia SKILL.md → actualiza LICENSE si cambió → abre PR con diff real de archivos (no de array JSON)
- `/plugin install android@mobiai` trae nuestros 4 skills + los 6 de Google en un solo plugin

---

## Decisiones ya tomadas

1. **Subdir `google/` dentro de `plugins/android/skills/`** — más limpio que flatten (aísla origen, evita colisiones futuras). Los paths finales son `plugins/android/skills/google/<skill-name>/SKILL.md`.
2. **No renombrar nada** — nombres de carpeta quedan igual a upstream (`agp-9-upgrade`, `edge-to-edge`, etc.). Flatten de la jerarquía upstream: `system/edge-to-edge` → `google/edge-to-edge`, `build/agp/agp-9-upgrade` → `google/agp-9-upgrade`. Perdemos la categoría intermedia de Google pero simplifica y no afecta funcionalidad (el `name` del frontmatter es lo que importa).
3. **No modificar el contenido** de los SKILL.md de Google. Copia verbatim.
4. **Cron semanal se mantiene** (lunes 06:00 UTC) + `workflow_dispatch` manual.
5. **Fallo explícito** si el workflow detecta cambio en LICENSE (requiere intervención humana).
6. **Primer import manual** en este mismo PR (no esperar al workflow) — ejecutamos la lógica localmente una vez para traer el estado actual.
7. **Branch**: `feat/vendor-android-official-skills` apilada sobre `feat/autonomy-fast-path`. PR #3 stack.

---

## Riesgos identificados

| Riesgo | Mitigación |
|---|---|
| Google cambia la licencia upstream | Workflow verifica hash/contenido de LICENSE.txt antes de commit; falla con error explícito si cambió. Humano revisa. |
| Colisión de nombres entre nuestros skills y los de Google | Inventario: nuestros 4 son `android-device`, `android-build`, `android-testing`, `android-architecture`. Los 6 de Google son `agp-9-upgrade`, `migrate-xml-views-to-jetpack-compose`, `navigation-3`, `r8-analyzer`, `play-billing-library-version-upgrade`, `edge-to-edge`. Sin colisión. Regla para el futuro: CI check que bloquee PR si dos `<name>/SKILL.md` tienen frontmatter `name:` igual. |
| Google elimina una skill | Workflow detecta skills que estaban y ya no: las borra del vendoreado + commit. |
| Google renombra una carpeta | Workflow lo ve como "una borrada + una agregada". El PR lo hace evidente al reviewer. |
| Skills tienen copyright headers al inicio del markdown | Preservados verbatim por ser copia byte-exacta. |
| Alguien hace `/plugin install android-official-skills@mobiai` antes de que el marketplace update llegue | Claude Code tira "not found in marketplace". Sin efectos colaterales. |
| Trademark de Google/Android en nuestro README | Ya cuidamos lenguaje factual: "mantenidos por Google (Apache 2.0)". Sin "endorsed by", sin logos. |

---

## Estructura de archivos target

```
MobiAI-Core/
├── .claude-plugin/
│   └── marketplace.json                   ← entrada android-official-skills REMOVIDA
├── .github/
│   └── workflows/
│       └── sync-android-skills.yml        ← REFACTOREADO a vendoring
├── plugins/
│   └── android/
│       ├── .claude-plugin/
│       │   └── plugin.json                ← dependency android-official-skills REMOVIDA, skills array expandida
│       └── skills/
│           ├── android-device/            ← existentes
│           ├── android-build/
│           ├── android-testing/
│           ├── android-architecture/
│           └── google/                    ← NUEVO directorio completo
│               ├── LICENSE                ← copia de upstream
│               ├── README.md              ← copia de upstream (atribución)
│               ├── agp-9-upgrade/
│               │   └── SKILL.md           ← verbatim de Google
│               ├── edge-to-edge/
│               │   └── SKILL.md
│               ├── migrate-xml-views-to-jetpack-compose/
│               │   └── SKILL.md
│               ├── navigation-3/
│               │   └── SKILL.md
│               ├── play-billing-library-version-upgrade/
│               │   └── SKILL.md
│               └── r8-analyzer/
│                   └── SKILL.md
├── NOTICE.md                              ← NUEVO, atribución de software de terceros
├── README.md                              ← actualizado
└── ... (resto igual)
```

---

## Pre-flight

- [ ] **Verificar rama y estado limpio**

```bash
cd "C:\dev\repos\MobiAI-Core\.claude\worktrees\multi-plugin-restructure"
git branch --show-current
```

Expected: `feat/vendor-android-official-skills`

```bash
git status
```

Expected: working tree clean (el plan file queda tracked tras el primer commit).

---

## Phase 1 — Primer import manual del contenido upstream

Antes de refactorear el workflow, traemos el contenido actual de Google al repo manualmente. Esto sirve también de test case para el workflow nuevo.

### Task 1.1: Clonar android/skills en carpeta temporal

- [ ] **Step 1.1.1: Clonar upstream a `_upstream/` (temporal, ignorado por git)**

```bash
cd "C:\dev\repos\MobiAI-Core\.claude\worktrees\multi-plugin-restructure"
git clone --depth 1 https://github.com/android/skills.git _upstream
```

Verify:
```bash
ls _upstream/
```
Expected: `build/  jetpack-compose/  LICENSE.txt  navigation/  performance/  play/  README.md  system/  .github/`

### Task 1.2: Preparar directorio `plugins/android/skills/google/`

- [ ] **Step 1.2.1: Crear el directorio**

```bash
mkdir -p plugins/android/skills/google
```

- [ ] **Step 1.2.2: Copiar LICENSE upstream**

```bash
cp _upstream/LICENSE.txt plugins/android/skills/google/LICENSE
```

Verify it's Apache 2.0:
```bash
head -5 plugins/android/skills/google/LICENSE
```
Expected: starts with `Apache License` and `Version 2.0, January 2004`.

- [ ] **Step 1.2.3: Copiar README upstream como atribución**

```bash
cp _upstream/README.md plugins/android/skills/google/README.md
```

Verify:
```bash
head -3 plugins/android/skills/google/README.md
```
Expected: título del README de android/skills.

- [ ] **Step 1.2.4: Verificar ausencia de NOTICE upstream**

```bash
test -f _upstream/NOTICE && echo "EXISTS — debe ser vendoreado" || echo "not present — ok"
```
Expected: `not present — ok`

Si existe, copiarlo también:
```bash
test -f _upstream/NOTICE && cp _upstream/NOTICE plugins/android/skills/google/NOTICE
```

### Task 1.3: Copiar los 6 SKILL.md preservando nombre de carpeta terminal

Cada SKILL.md upstream vive en `<category>/.../<skill-name>/SKILL.md`. Flattenamos a `<skill-name>/SKILL.md` bajo `google/`.

- [ ] **Step 1.3.1: Detectar todos los SKILL.md upstream y sus nombres de carpeta terminal**

```bash
find _upstream -name SKILL.md -type f | sort
```
Expected output:
```
_upstream/build/agp/agp-9-upgrade/SKILL.md
_upstream/jetpack-compose/migration/migrate-xml-views-to-jetpack-compose/SKILL.md
_upstream/navigation/navigation-3/SKILL.md
_upstream/performance/r8-analyzer/SKILL.md
_upstream/play/play-billing-library-version-upgrade/SKILL.md
_upstream/system/edge-to-edge/SKILL.md
```

- [ ] **Step 1.3.2: Copiar cada uno preservando su carpeta contenedora (y sólo esa)**

Loop manual por ahora (el workflow lo automatizará):

```bash
mkdir -p plugins/android/skills/google/agp-9-upgrade
cp -r _upstream/build/agp/agp-9-upgrade/* plugins/android/skills/google/agp-9-upgrade/

mkdir -p plugins/android/skills/google/migrate-xml-views-to-jetpack-compose
cp -r _upstream/jetpack-compose/migration/migrate-xml-views-to-jetpack-compose/* plugins/android/skills/google/migrate-xml-views-to-jetpack-compose/

mkdir -p plugins/android/skills/google/navigation-3
cp -r _upstream/navigation/navigation-3/* plugins/android/skills/google/navigation-3/

mkdir -p plugins/android/skills/google/r8-analyzer
cp -r _upstream/performance/r8-analyzer/* plugins/android/skills/google/r8-analyzer/

mkdir -p plugins/android/skills/google/play-billing-library-version-upgrade
cp -r _upstream/play/play-billing-library-version-upgrade/* plugins/android/skills/google/play-billing-library-version-upgrade/

mkdir -p plugins/android/skills/google/edge-to-edge
cp -r _upstream/system/edge-to-edge/* plugins/android/skills/google/edge-to-edge/
```

Verify:
```bash
find plugins/android/skills/google -name SKILL.md -type f | sort
```
Expected: 6 paths matching the 6 names above.

- [ ] **Step 1.3.3: Verificar que el contenido es byte-idéntico al upstream**

```bash
for name in agp-9-upgrade migrate-xml-views-to-jetpack-compose navigation-3 r8-analyzer play-billing-library-version-upgrade edge-to-edge; do
  # Find the upstream path (it varies per skill)
  upstream=$(find _upstream -name SKILL.md -path "*$name*" -type f | head -1)
  vendored="plugins/android/skills/google/$name/SKILL.md"
  if diff -q "$upstream" "$vendored" > /dev/null; then
    echo "OK: $name"
  else
    echo "FAIL: $name — diff detected"
    diff "$upstream" "$vendored" | head -20
  fi
done
```
Expected: 6x `OK:` lines.

### Task 1.4: Limpiar el clon temporal

- [ ] **Step 1.4.1: Eliminar `_upstream/`**

```bash
rm -rf _upstream
```

- [ ] **Step 1.4.2: Agregar `_upstream/` a `.gitignore`**

Edit `.gitignore`. Current content:
```
# OS
.DS_Store
Thumbs.db

# IDE
.idea/
.vscode/
*.swp
*.swo

# Node (if any scripts need it)
node_modules/

# Claude Code local config
.claude/

# Output
output/
*.log
```

Append:
```

# Temporary upstream clones (used by sync-android-skills workflow)
_upstream/
```

Verify:
```bash
grep -c "_upstream" .gitignore
```
Expected: `1`

### Task 1.5: Commit del import manual

- [ ] **Step 1.5.1: Stage y commit**

```bash
git add plugins/android/skills/google/ .gitignore
git commit -m "feat(android): vendorear skills oficiales de Google (Apache 2.0)

Trae las 6 SKILL.md de github.com/android/skills a plugins/android/skills/google/:

- agp-9-upgrade — upgrade a Android Gradle Plugin 9
- migrate-xml-views-to-jetpack-compose — migrar XML Views a Compose
- navigation-3 — Jetpack Navigation 3
- r8-analyzer — analisis de keep rules R8/ProGuard
- play-billing-library-version-upgrade — upgrade de Play Billing Library
- edge-to-edge — adaptive edge-to-edge en Compose

Incluye LICENSE.txt y README.md de upstream (preservados verbatim).
Apache 2.0 §4 requiere preservar LICENSE y attribution; cumplido.
Apache 2.0 §6 prohibe implicar endorsement; lenguaje factual en docs.

Import manual en este commit; el workflow sync-android-skills.yml se
refactorea en commits siguientes para automatizar el mantenimiento."
```

---

## Phase 2 — Actualizar configs del plugin y marketplace

### Task 2.1: Eliminar entrada `android-official-skills` de marketplace.json

**Files:**
- Modify: `.claude-plugin/marketplace.json`

- [ ] **Step 2.1.1: Usar jq para remover la entrada**

```bash
jq '.plugins = (.plugins | map(select(.name != "android-official-skills")))' \
  .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
```

- [ ] **Step 2.1.2: Verificar**

```bash
jq '.plugins | length' .claude-plugin/marketplace.json
```
Expected: `7` (antes eran 8)

```bash
jq -r '.plugins[].name' .claude-plugin/marketplace.json | sort
```
Expected:
```
android
core
flutter
ios
kmp
mobile
react-native
```

Sin `android-official-skills`.

### Task 2.2: Actualizar `plugins/android/.claude-plugin/plugin.json`

**Files:**
- Modify: `plugins/android/.claude-plugin/plugin.json`

Eliminar `android-official-skills` de `dependencies`. Agregar `skills` array para que Claude descubra tanto `./skills/` como `./skills/google/`.

- [ ] **Step 2.2.1: Leer el archivo actual**

Expected content:
```json
{
  "name": "android",
  "description": "MobiAI para Android — skills de device, build, testing, architecture + skills oficiales de Google vía android-official-skills",
  "version": "2.0.0",
  "author": { ... },
  "homepage": "...",
  "repository": "...",
  "license": "MIT",
  "keywords": ["android", "kotlin", "adb", "gradle", "mobile"],
  "dependencies": [
    "core",
    "android-official-skills"
  ]
}
```

- [ ] **Step 2.2.2: Reemplazar con Write**

Nuevo contenido:
```json
{
  "name": "android",
  "description": "MobiAI para Android — skills de device, build, testing, architecture, e integración con skills oficiales de Google (Apache 2.0)",
  "version": "2.1.0",
  "author": {
    "name": "Matias Rosenstein",
    "url": "https://github.com/ArisGuimera/MobiAI-Core"
  },
  "homepage": "https://github.com/ArisGuimera/MobiAI-Core",
  "repository": "https://github.com/ArisGuimera/MobiAI-Core",
  "license": "MIT",
  "keywords": ["android", "kotlin", "adb", "gradle", "mobile"],
  "dependencies": ["core"],
  "skills": [
    "./skills/",
    "./skills/google/"
  ]
}
```

Cambios:
- `description` actualizada (sin mención del plugin separado)
- `version` 2.0.0 → 2.1.0 (minor bump: content nuevo vendoreado)
- `dependencies` sin `android-official-skills`
- `skills` array nuevo con los dos directorios

- [ ] **Step 2.2.3: Validar JSON**

```bash
python -c "import json; json.load(open('plugins/android/.claude-plugin/plugin.json'))" && echo "OK"
```

### Task 2.3: Validar marketplace

- [ ] **Step 2.3.1: Correr el validator**

```bash
claude plugin validate .
```
Expected: `✔ Validation passed`

Si falla con "skill path not found" verificar:
- `./skills/` existe (sí, con las 4 skills nuestras)
- `./skills/google/` existe (sí, con las 6 vendoreadas)

### Task 2.4: Commit config

- [ ] **Step 2.4.1: Commit**

```bash
git add .claude-plugin/marketplace.json plugins/android/.claude-plugin/plugin.json
git commit -m "feat(android): eliminar plugin separado android-official-skills

El contenido de Google ahora vive vendoreado en plugins/android/skills/google/
(commit anterior). Con eso:

- marketplace.json queda en 7 plugins (antes 8), sin entrada confusa para
  el usuario final
- android/plugin.json saca la dependency android-official-skills (ya no
  es un plugin separado) y declara skills array con ambos paths:
  ./skills/ (nuestros) + ./skills/google/ (vendoreados)
- version bump de android a 2.1.0 (minor: content nuevo integrado, no
  breaking para quien tenga 2.0.0)

UX result: /plugin list muestra solo android, no hay riesgo de que el
usuario instale el plugin interno por equivocacion."
```

---

## Phase 3 — NOTICE.md y README

### Task 3.1: Crear `NOTICE.md` en raíz

**Files:**
- Create: `NOTICE.md`

- [ ] **Step 3.1.1: Escribir el archivo**

```markdown
# NOTICE

MobiAI-Core includes software developed by third parties, redistributed under their respective open-source licenses.

## Third-party software vendored in this repository

### Android Skills (maintained by Google)

- **Upstream:** https://github.com/android/skills
- **License:** Apache License 2.0
- **Vendored to:** `plugins/android/skills/google/`
- **License file:** `plugins/android/skills/google/LICENSE`

The SKILL.md files under `plugins/android/skills/google/` are copied verbatim from the upstream repository. MobiAI-Core does not modify their content. Redistribution complies with Apache 2.0 §4: the original `LICENSE.txt` is preserved, and attribution is documented both in this file and in `README.md`.

MobiAI-Core is not affiliated with, endorsed by, or sponsored by Google or the Android Open Source Project. References to "Android" and links to upstream resources are factual attribution, not trademark usage implying endorsement.

Automated synchronization of this content is performed by `.github/workflows/sync-android-skills.yml` on a weekly cadence.
```

### Task 3.2: Actualizar `README.md` con la atribución

**Files:**
- Modify: `README.md`

Agregar una subsección en "Skills disponibles" que mencione la integración.

- [ ] **Step 3.2.1: Leer el bloque actual de "Skills disponibles"**

Ubicar (alrededor de línea 90):
```markdown
- **Android** (en `android`) — android-device, android-build, android-testing, android-architecture + skills oficiales de Google vía `android-official-skills`
```

- [ ] **Step 3.2.2: Reemplazar con texto actualizado**

Nuevo texto:
```markdown
- **Android** (en `android`) — android-device, android-build, android-testing, android-architecture, más los [skills oficiales de Android mantenidos por Google](https://github.com/android/skills) (Apache 2.0, vendoreados en `plugins/android/skills/google/` con auto-sync semanal). Ver [NOTICE.md](NOTICE.md).
```

- [ ] **Step 3.2.3: Actualizar la tabla de plugins disponibles**

Ubicar la tabla:
```markdown
| `android` | Skills Android + skills oficiales de Google | `/plugin install android@mobiai` |
```

Dejar igual — la descripción ya es precisa y no expone el detalle interno de implementación.

### Task 3.3: Commit docs

- [ ] **Step 3.3.1: Commit**

```bash
git add NOTICE.md README.md
git commit -m "docs: agregar NOTICE.md y actualizar README con atribucion a android/skills

Apache 2.0 §4 exige preservar atribucion de forma visible en el producto
derivado. NOTICE.md en raiz lista third-party software vendoreado, linkea
al LICENSE preservado, y aclara explicitamente que MobiAI no implica
endorsement de Google ni AOSP (Apache 2.0 §6).

README linkea a NOTICE.md y documenta que los skills de Google vienen
con auto-sync semanal, sin exponer detalles de implementacion innecesarios
al usuario final."
```

---

## Phase 4 — Refactorear el workflow de sync

El workflow actual edita el array `skills` del plugin entry de `android-official-skills` cuando detecta nuevas carpetas raíz en upstream. Ya no existe ese plugin entry. El workflow nuevo:

1. Clona upstream a `_upstream/`
2. Para cada SKILL.md upstream, determina la carpeta terminal y copia a `plugins/android/skills/google/<name>/`
3. Detecta skills vendoreadas que ya no existen upstream (eliminadas) y las borra
4. Compara `LICENSE` — si cambió, FALLA con error explícito (requiere intervención humana)
5. Copia `README.md` upstream (con cambio o no — siempre se refresca)
6. Si hay cambios reales en el repo, abre PR

### Task 4.1: Reescribir `.github/workflows/sync-android-skills.yml`

**Files:**
- Modify: `.github/workflows/sync-android-skills.yml`

- [ ] **Step 4.1.1: Escribir el workflow nuevo**

Contenido completo:

```yaml
name: Sync android/skills (vendoring)

on:
  schedule:
    - cron: '0 6 * * 1'  # cada lunes a las 06:00 UTC
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write

concurrency:
  group: sync-android-skills
  cancel-in-progress: false

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout MobiAI-Core
        uses: actions/checkout@v4

      - name: Checkout android/skills (shallow)
        uses: actions/checkout@v4
        with:
          repository: android/skills
          path: _upstream

      - name: Verify upstream LICENSE is unchanged Apache 2.0
        run: |
          set -euo pipefail
          if ! grep -q "Apache License" _upstream/LICENSE.txt; then
            echo "::error::Upstream LICENSE.txt no longer contains 'Apache License'. Manual review required."
            exit 1
          fi
          if ! grep -q "Version 2.0" _upstream/LICENSE.txt; then
            echo "::error::Upstream LICENSE.txt no longer references Version 2.0. Manual review required."
            exit 1
          fi
          # Compare to our vendored LICENSE byte-for-byte
          if ! diff -q _upstream/LICENSE.txt plugins/android/skills/google/LICENSE > /dev/null; then
            echo "::notice::Upstream LICENSE differs from vendored copy. Will update in the sync PR."
          fi

      - name: Refresh vendored LICENSE and README
        run: |
          set -euo pipefail
          cp _upstream/LICENSE.txt plugins/android/skills/google/LICENSE
          cp _upstream/README.md plugins/android/skills/google/README.md
          # NOTICE file (present upstream → vendor; absent → remove our copy if any)
          if [ -f _upstream/NOTICE ]; then
            cp _upstream/NOTICE plugins/android/skills/google/NOTICE
          else
            rm -f plugins/android/skills/google/NOTICE
          fi

      - name: Sync SKILL.md files
        env:
          VENDOR_ROOT: plugins/android/skills/google
        run: |
          set -euo pipefail

          # Build list of upstream skill dirs (terminal names)
          mapfile -t upstream_paths < <(find _upstream -name SKILL.md -type f | sort)
          declare -A upstream_names

          for path in "${upstream_paths[@]}"; do
            # path is _upstream/<category>/.../<name>/SKILL.md
            skill_dir=$(dirname "$path")
            name=$(basename "$skill_dir")
            upstream_names[$name]=1

            target_dir="$VENDOR_ROOT/$name"
            mkdir -p "$target_dir"
            # Copy entire skill dir contents (SKILL.md + any references/ or scripts/)
            cp -r "$skill_dir"/* "$target_dir/"
            echo "Synced: $name"
          done

          # Remove vendored skills that no longer exist upstream
          for vendored_dir in "$VENDOR_ROOT"/*/; do
            [ -d "$vendored_dir" ] || continue
            name=$(basename "$vendored_dir")
            if [ -z "${upstream_names[$name]:-}" ]; then
              echo "Removing (no longer upstream): $name"
              rm -rf "$vendored_dir"
            fi
          done

      - name: Clean up _upstream clone
        run: rm -rf _upstream

      - name: Detect changes
        id: changes
        run: |
          set -euo pipefail
          if git diff --quiet --exit-code; then
            echo "has_changes=false" >> "$GITHUB_OUTPUT"
            echo "No changes detected."
          else
            echo "has_changes=true" >> "$GITHUB_OUTPUT"
            echo "Changes detected:"
            git diff --stat
          fi

      - name: Open PR
        if: steps.changes.outputs.has_changes == 'true'
        uses: peter-evans/create-pull-request@v6
        with:
          branch: chore/sync-android-skills
          delete-branch: true
          commit-message: "chore: sincronizar android/skills vendoreados"
          title: "chore: sincronizar android/skills vendoreados"
          body: |
            Sincronizacion automatica del contenido vendoreado desde
            [android/skills](https://github.com/android/skills) upstream.

            El diff de este PR muestra los archivos que cambiaron: SKILL.md
            actualizados, skills agregadas, skills removidas, o cambios en
            LICENSE/README.

            Auto-generado por `.github/workflows/sync-android-skills.yml`.
            Apache 2.0 attribution preserved (ver `NOTICE.md`).
```

- [ ] **Step 4.1.2: Validar sintaxis YAML**

```bash
python -c "import yaml; yaml.safe_load(open('.github/workflows/sync-android-skills.yml'))" && echo "YAML_OK"
```

Si Python no tiene PyYAML instalado, skip (el workflow se valida al corro en GitHub).

### Task 4.2: Commit workflow refactoreado

- [ ] **Step 4.2.1: Commit**

```bash
git add .github/workflows/sync-android-skills.yml
git commit -m "ci: refactorear sync-android-skills.yml a modelo de vendoring

Antes: el workflow detectaba carpetas raiz nuevas en upstream y editaba
el array skills del plugin entry android-official-skills en marketplace.json.

Ahora que el plugin separado no existe, el workflow:

1. Clona android/skills upstream
2. Verifica que la LICENSE sigue siendo Apache 2.0 (falla explicito si cambio)
3. Copia cada SKILL.md (+ references/ + scripts/) a plugins/android/skills/google/<name>/
4. Borra skills vendoreadas que ya no existen upstream
5. Refresca LICENSE y README (siempre)
6. Si hay cambios reales en el repo, abre PR

Incluye concurrency group para prevenir races de cron + manual dispatch,
y mantiene la semantica de 'falla si Apache 2.0 cambia' para obligar review
humano ante cambios de licencia upstream."
```

---

## Phase 5 — Validación local end-to-end

### Task 5.1: Correr validador

- [ ] **Step 5.1.1: Validar marketplace**

```bash
claude plugin validate .
```
Expected: `✔ Validation passed`

### Task 5.2: Re-instalar el plugin android desde el worktree

- [ ] **Step 5.2.1: En Claude Code (nueva sesión en otro directorio)**

```
/plugin marketplace add C:/dev/repos/MobiAI-Core/.claude/worktrees/multi-plugin-restructure
/plugin install android@mobiai
```

Expected: instala `android` + `core` (2 plugins, NO android-official-skills).

### Task 5.3: Verificar que los 10 skills Android están disponibles

- [ ] **Step 5.3.1: Probar un skill nuestro**

En el agente: "usá el skill `android-device`". Debería cargar desde `plugins/android/skills/android-device/SKILL.md`.

- [ ] **Step 5.3.2: Probar un skill de Google vendoreado**

En el agente: "usá el skill `edge-to-edge`". Debería cargar desde `plugins/android/skills/google/edge-to-edge/SKILL.md` (path contiene `/google/`).

- [ ] **Step 5.3.3: Verificar que android-official-skills no existe**

```
/plugin list
```
Expected: sin entrada `android-official-skills`.

```
/plugin install android-official-skills@mobiai
```
Expected: `Plugin "android-official-skills" not found in any marketplace`.

---

## Phase 6 — Dry-run del workflow nuevo (opcional pero recomendado)

### Task 6.1: Simular el workflow localmente

El objetivo es correr la lógica del workflow en local para confirmar que NO genera cambios cuando el contenido ya está sincronizado.

- [ ] **Step 6.1.1: Correr la lógica de sync a mano**

```bash
# Simular el clone
git clone --depth 1 https://github.com/android/skills.git _upstream

# Refresh LICENSE/README
cp _upstream/LICENSE.txt plugins/android/skills/google/LICENSE
cp _upstream/README.md plugins/android/skills/google/README.md
[ -f _upstream/NOTICE ] && cp _upstream/NOTICE plugins/android/skills/google/NOTICE || rm -f plugins/android/skills/google/NOTICE

# Sync SKILL.md files (loop igual al workflow)
find _upstream -name SKILL.md -type f | while read path; do
  name=$(basename $(dirname "$path"))
  mkdir -p "plugins/android/skills/google/$name"
  cp -r "$(dirname "$path")"/* "plugins/android/skills/google/$name/"
done

# Detect changes
git diff --stat
```

Expected: `0 files changed` (el estado post-Phase 1 ya refleja upstream, nada nuevo).

- [ ] **Step 6.1.2: Limpiar el clon**

```bash
rm -rf _upstream
```

---

## Phase 7 — Push y PR

### Task 7.1: Revisar log de commits antes de push

- [ ] **Step 7.1.1: Ver commits de este branch**

```bash
git log --oneline feat/autonomy-fast-path..HEAD
```

Expected: 4 commits:
- `feat(android): vendorear skills oficiales de Google (Apache 2.0)`
- `feat(android): eliminar plugin separado android-official-skills`
- `docs: agregar NOTICE.md y actualizar README con atribucion a android/skills`
- `ci: refactorear sync-android-skills.yml a modelo de vendoring`

### Task 7.2: Pedir aprobación al usuario para push

⚠️ Per CLAUDE.md rules: NO push sin aprobación explícita.

- [ ] **Step 7.2.1: Mostrar resumen al usuario**

```
=== Branch: feat/vendor-android-official-skills ===
=== Commits: 4 ===
1. feat(android): vendorear skills oficiales de Google (Apache 2.0)
2. feat(android): eliminar plugin separado android-official-skills
3. docs: agregar NOTICE.md y actualizar README con atribucion a android/skills
4. ci: refactorear sync-android-skills.yml a modelo de vendoring

=== File changes summary ===
(git diff --stat feat/autonomy-fast-path..HEAD)

¿Apruebo push y PR contra feat/autonomy-fast-path?
```

### Task 7.3: Push si aprobado

- [ ] **Step 7.3.1: Push**

```bash
git push -u origin feat/vendor-android-official-skills
```

### Task 7.4: Crear PR

- [ ] **Step 7.4.1: Con gh CLI**

```bash
gh pr create \
  --base feat/autonomy-fast-path \
  --head feat/vendor-android-official-skills \
  --title "feat: vendorear skills oficiales de Google, eliminar plugin separado" \
  --body-file - << 'PR_BODY'
## Motivacion

UX issue reportado por usuario: el marketplace mostraba dos entradas separadas — `android` y `android-official-skills` — y gente que no lee descriptions no sabia cual instalar. Este PR consolida todo bajo un solo plugin `android`.

## Cambios

1. Vendoring de las 6 SKILL.md de github.com/android/skills a `plugins/android/skills/google/` (verbatim, Apache 2.0 preservado)
2. Eliminacion del plugin `android-official-skills` del marketplace.json
3. `plugins/android/plugin.json`: sin dependency `android-official-skills`, skills array expandida a `["./skills/", "./skills/google/"]`, version bump 2.0.0 → 2.1.0
4. `NOTICE.md` nuevo en raiz del repo con atribucion de third-party vendored software (cumple Apache 2.0 §4)
5. README.md actualizado con link al NOTICE y aclaracion de auto-sync
6. `.github/workflows/sync-android-skills.yml` refactoreado: ahora copia archivos en vez de editar array JSON. Mantiene cron semanal, agrega concurrency group, agrega falla explicita ante cambio de LICENSE upstream.

## Compliance Apache 2.0

- §4.a: LICENSE preservada en `plugins/android/skills/google/LICENSE` (copia byte-exacta)
- §4.b: Modificaciones documentadas — no modificamos contenido, solo reubicamos
- §4.c: Copyright headers preservados verbatim (copia byte-exacta)
- §4.d: NOTICE upstream inspeccionada — no existe (si aparece en el futuro, el workflow la vendorea)
- §6: Cero lenguaje de endorsement en README / NOTICE / descripciones. "Maintained by Google" es atribucion factual.

## Verificacion

- `claude plugin validate .` pasa
- Install local de `android@mobiai` trae `android` + `core` (sin el plugin externo)
- Los 10 skills Android (4 nuestros + 6 vendoreados) cargan correctamente

## UX resultante

/plugin list post-install:
- `mobile` (si instalaste el meta)
- `core`
- `android`
- `ios`, `kmp`, `flutter`, `react-native` (si instalaste otros)

Sin `android-official-skills`. Un plugin por plataforma. Escalable a futuras fuentes.

Base: `feat/autonomy-fast-path` (stack PR sobre PR #2).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
PR_BODY
```

Si gh CLI no responde o no está configurada, fallback: dar al usuario el URL `https://github.com/ArisGuimera/MobiAI-Core/compare/feat/autonomy-fast-path...feat/vendor-android-official-skills` y el body para paste manual.

---

---

## Phase 8 — Prefijo `mobiai-` en skills propios (UX: distinguir nuestros de los de Google)

### Motivación

Cuando Claude dice "Using skill `android-device`" o "Using skill `edge-to-edge`", el usuario no puede distinguir cuáles son de MobiAI y cuáles de Google. Prefijar nuestros skills con `mobiai-` deja explícito el origen cada vez que un skill se carga.

### Alcance del rename

**Skills que SE renombran con prefijo `mobiai-`** (29 skills):

En `plugins/core/skills/`:
- `fix-issue` → `mobiai-fix-issue`
- `create-pr` → `mobiai-create-pr`
- `review-code` → `mobiai-review-code`
- `write-tests` → `mobiai-write-tests`
- `analyze-crash` → `mobiai-analyze-crash`
- `crashlytics` → `mobiai-crashlytics`
- `reproduce-bug` → `mobiai-reproduce-bug`
- `writing-skills` → `mobiai-writing-skills`
- `mobile-brainstorming` → `mobiai-mobile-brainstorming`
- `mobile-debugging` → `mobiai-mobile-debugging`
- `mobile-tdd` → `mobiai-mobile-tdd`
- `mobile-planning` → `mobiai-mobile-planning`
- `mobile-verification` → `mobiai-mobile-verification`
- `mobile-executing-plans` → `mobiai-mobile-executing-plans`
- `mobile-parallel-agents` → `mobiai-mobile-parallel-agents`
- `mobile-subagent-development` → `mobiai-mobile-subagent-development`
- `mobile-worktrees` → `mobiai-mobile-worktrees`
- `mobile-finishing-branch` → `mobiai-mobile-finishing-branch`

En `plugins/android/skills/`:
- `android-device` → `mobiai-android-device`
- `android-build` → `mobiai-android-build`
- `android-testing` → `mobiai-android-testing`
- `android-architecture` → `mobiai-android-architecture`

En `plugins/ios/skills/`:
- `ios-device` → `mobiai-ios-device`
- `ios-build` → `mobiai-ios-build`
- `ios-testing` → `mobiai-ios-testing`
- `ios-architecture` → `mobiai-ios-architecture`

En `plugins/kmp/skills/`: `kmp` → `mobiai-kmp`
En `plugins/flutter/skills/`: `flutter` → `mobiai-flutter`
En `plugins/react-native/skills/`: `react-native` → `mobiai-react-native`

**Skills que NO se renombran:**
- `using-mobiai` — ya tiene la marca en el nombre, renombrarlo a `mobiai-using-mobiai` sería redundante y tocaría el bootstrap hook que referencia el path.
- Los 6 skills vendoreados de Google (`plugins/android/skills/google/*`) — son de Google, mantienen sus nombres upstream para que el auto-sync no genere diffs espurios.

### Task 8.1: Rename de carpetas con `git mv`

**Files:** 29 directorios bajo `plugins/*/skills/<skill>/` renombrados.

- [ ] **Step 8.1.1: Rename `core` skills (18 skills, excluyendo using-mobiai)**

```bash
cd "C:\dev\repos\MobiAI-Core\.claude\worktrees\multi-plugin-restructure"

git mv plugins/core/skills/fix-issue plugins/core/skills/mobiai-fix-issue
git mv plugins/core/skills/create-pr plugins/core/skills/mobiai-create-pr
git mv plugins/core/skills/review-code plugins/core/skills/mobiai-review-code
git mv plugins/core/skills/write-tests plugins/core/skills/mobiai-write-tests
git mv plugins/core/skills/analyze-crash plugins/core/skills/mobiai-analyze-crash
git mv plugins/core/skills/crashlytics plugins/core/skills/mobiai-crashlytics
git mv plugins/core/skills/reproduce-bug plugins/core/skills/mobiai-reproduce-bug
git mv plugins/core/skills/writing-skills plugins/core/skills/mobiai-writing-skills
git mv plugins/core/skills/mobile-brainstorming plugins/core/skills/mobiai-mobile-brainstorming
git mv plugins/core/skills/mobile-debugging plugins/core/skills/mobiai-mobile-debugging
git mv plugins/core/skills/mobile-tdd plugins/core/skills/mobiai-mobile-tdd
git mv plugins/core/skills/mobile-planning plugins/core/skills/mobiai-mobile-planning
git mv plugins/core/skills/mobile-verification plugins/core/skills/mobiai-mobile-verification
git mv plugins/core/skills/mobile-executing-plans plugins/core/skills/mobiai-mobile-executing-plans
git mv plugins/core/skills/mobile-parallel-agents plugins/core/skills/mobiai-mobile-parallel-agents
git mv plugins/core/skills/mobile-subagent-development plugins/core/skills/mobiai-mobile-subagent-development
git mv plugins/core/skills/mobile-worktrees plugins/core/skills/mobiai-mobile-worktrees
git mv plugins/core/skills/mobile-finishing-branch plugins/core/skills/mobiai-mobile-finishing-branch
```

- [ ] **Step 8.1.2: Rename `android` skills (4 skills)**

```bash
git mv plugins/android/skills/android-device plugins/android/skills/mobiai-android-device
git mv plugins/android/skills/android-build plugins/android/skills/mobiai-android-build
git mv plugins/android/skills/android-testing plugins/android/skills/mobiai-android-testing
git mv plugins/android/skills/android-architecture plugins/android/skills/mobiai-android-architecture
```

- [ ] **Step 8.1.3: Rename `ios` skills (4 skills)**

```bash
git mv plugins/ios/skills/ios-device plugins/ios/skills/mobiai-ios-device
git mv plugins/ios/skills/ios-build plugins/ios/skills/mobiai-ios-build
git mv plugins/ios/skills/ios-testing plugins/ios/skills/mobiai-ios-testing
git mv plugins/ios/skills/ios-architecture plugins/ios/skills/mobiai-ios-architecture
```

- [ ] **Step 8.1.4: Rename multi-platform skills (3 skills)**

```bash
git mv plugins/kmp/skills/kmp plugins/kmp/skills/mobiai-kmp
git mv plugins/flutter/skills/flutter plugins/flutter/skills/mobiai-flutter
git mv plugins/react-native/skills/react-native plugins/react-native/skills/mobiai-react-native
```

- [ ] **Step 8.1.5: Verificar que using-mobiai NO se renombró**

```bash
ls plugins/core/skills/ | grep using-mobiai
```
Expected: `using-mobiai` (sin prefijo).

- [ ] **Step 8.1.6: Verificar count total de skills renombradas**

```bash
find plugins/core/skills plugins/android/skills plugins/ios/skills plugins/kmp/skills plugins/flutter/skills plugins/react-native/skills -maxdepth 1 -type d -name 'mobiai-*' | wc -l
```
Expected: `29`

### Task 8.2: Actualizar frontmatter `name:` en cada SKILL.md renombrado

Cada `SKILL.md` tiene un frontmatter con `name: <nombre>`. Ese nombre debe coincidir con el nombre de la carpeta terminal (Claude Code lo usa para skill invocation).

- [ ] **Step 8.2.1: Script para actualizar los 29 nombres**

```bash
for dir in $(find plugins/core/skills plugins/android/skills plugins/ios/skills plugins/kmp/skills plugins/flutter/skills plugins/react-native/skills -maxdepth 1 -type d -name 'mobiai-*'); do
  new_name=$(basename "$dir")
  skill_file="$dir/SKILL.md"
  if [ ! -f "$skill_file" ]; then
    echo "SKIP: no SKILL.md in $dir"
    continue
  fi
  # Extract current name from frontmatter
  current_name=$(sed -n 's/^name:[[:space:]]*\(.*\)$/\1/p' "$skill_file" | head -1 | tr -d '\r')
  if [ "$current_name" = "$new_name" ]; then
    echo "OK (already): $new_name"
    continue
  fi
  # Replace the first occurrence of `name: <anything>` with `name: $new_name`
  sed -i "0,/^name:[[:space:]]*.*/s//name: $new_name/" "$skill_file"
  echo "Updated: $current_name → $new_name"
done
```

- [ ] **Step 8.2.2: Verificar que todos los frontmatter matchean nombre de carpeta**

```bash
mismatch=0
for dir in $(find plugins/core/skills plugins/android/skills plugins/ios/skills plugins/kmp/skills plugins/flutter/skills plugins/react-native/skills -maxdepth 1 -type d); do
  name=$(basename "$dir")
  [ "$name" = "skills" ] && continue
  # Skip Google vendored
  [[ "$dir" == *"/google/"* ]] && continue
  skill_file="$dir/SKILL.md"
  [ ! -f "$skill_file" ] && continue
  frontmatter_name=$(sed -n 's/^name:[[:space:]]*\(.*\)$/\1/p' "$skill_file" | head -1 | tr -d '\r')
  if [ "$name" != "$frontmatter_name" ]; then
    echo "MISMATCH: folder=$name frontmatter=$frontmatter_name"
    mismatch=1
  fi
done
[ $mismatch -eq 0 ] && echo "All frontmatter names match folder names"
```

Expected: `All frontmatter names match folder names`.

### Task 8.3: Actualizar cross-references en todos los SKILL.md

Muchos skills referencian otros por nombre (ej. "invoke skill `mobile-debugging`"). Con el rename, esas referencias rompen.

- [ ] **Step 8.3.1: Detectar referencias obsoletas**

```bash
# Buscar todas las menciones de los nombres viejos
old_names="fix-issue create-pr review-code write-tests analyze-crash crashlytics reproduce-bug writing-skills mobile-brainstorming mobile-debugging mobile-tdd mobile-planning mobile-verification mobile-executing-plans mobile-parallel-agents mobile-subagent-development mobile-worktrees mobile-finishing-branch android-device android-build android-testing android-architecture ios-device ios-build ios-testing ios-architecture kmp flutter react-native"

echo "=== Mentions of old skill names in SKILL.md files ==="
for name in $old_names; do
  count=$(grep -rl "\`$name\`" plugins/*/skills/*/SKILL.md 2>/dev/null | wc -l)
  [ "$count" -gt 0 ] && echo "$name: $count files"
done
```

- [ ] **Step 8.3.2: Reemplazar referencias con backticks**

El patrón más común es `` `skill-name` `` (con backticks). Reemplazar:

```bash
old_names=(fix-issue create-pr review-code write-tests analyze-crash crashlytics reproduce-bug writing-skills mobile-brainstorming mobile-debugging mobile-tdd mobile-planning mobile-verification mobile-executing-plans mobile-parallel-agents mobile-subagent-development mobile-worktrees mobile-finishing-branch android-device android-build android-testing android-architecture ios-device ios-build ios-testing ios-architecture kmp flutter react-native)

for name in "${old_names[@]}"; do
  # Use word boundaries via sed: replace `name` only when it's in backticks
  find plugins -name 'SKILL.md' -type f ! -path '*/google/*' -exec sed -i "s/\`$name\`/\`mobiai-$name\`/g" {} +
done

# Verify (should show 0 un-replaced backticked old names)
remaining=0
for name in "${old_names[@]}"; do
  count=$(grep -r "\`$name\`" plugins --include='SKILL.md' 2>/dev/null | grep -v '/google/' | wc -l)
  if [ "$count" -gt 0 ]; then
    echo "REMAINING: \`$name\` in $count places"
    remaining=$((remaining + count))
  fi
done
[ $remaining -eq 0 ] && echo "All backticked old names replaced"
```

- [ ] **Step 8.3.3: Revisión manual de referencias sin backticks**

Algunos skills pueden mencionar otros sin backticks. Hacer una pasada visual:

```bash
for name in "${old_names[@]}"; do
  echo "=== $name ==="
  grep -rn "$name" plugins/*/skills/*/SKILL.md 2>/dev/null | grep -v '/google/' | grep -v "mobiai-$name" | head -10
done
```

Cualquier match visible (que NO sea ya el reemplazo `mobiai-<name>`) requiere edit manual con Edit tool.

### Task 8.4: Actualizar `using-mobiai/SKILL.md` (el bootstrap)

El bootstrap tiene una **Quick Decision Guide** con tabla de todos los skills por nombre. Todas esas referencias rompen.

**Files:**
- Modify: `plugins/core/skills/using-mobiai/SKILL.md`

- [ ] **Step 8.4.1: Aplicar reemplazos en el bootstrap**

```bash
file="plugins/core/skills/using-mobiai/SKILL.md"
old_names=(fix-issue create-pr review-code write-tests analyze-crash crashlytics reproduce-bug writing-skills mobile-brainstorming mobile-debugging mobile-tdd mobile-planning mobile-verification mobile-executing-plans mobile-parallel-agents mobile-subagent-development mobile-worktrees mobile-finishing-branch android-device android-build android-testing android-architecture ios-device ios-build ios-testing ios-architecture kmp flutter react-native)

for name in "${old_names[@]}"; do
  sed -i "s/\`$name\`/\`mobiai-$name\`/g" "$file"
done
```

- [ ] **Step 8.4.2: Verificación visual**

Leer el bootstrap completo y confirmar que la Quick Decision Guide y cualquier listado mencionan `mobiai-*` correctamente (salvo las 6 skills de Google + `using-mobiai` mismo).

### Task 8.5: Actualizar CLAUDE.md, README.md, CONTRIBUTING.md, docs/crear-skills.md

Cualquier doc que referencie skills por nombre necesita update.

- [ ] **Step 8.5.1: Detectar menciones**

```bash
docs="CLAUDE.md README.md CONTRIBUTING.md docs/crear-skills.md .claude/CLAUDE.md"
old_names=(fix-issue create-pr review-code write-tests analyze-crash crashlytics reproduce-bug writing-skills mobile-brainstorming mobile-debugging mobile-tdd mobile-planning mobile-verification mobile-executing-plans mobile-parallel-agents mobile-subagent-development mobile-worktrees mobile-finishing-branch android-device android-build android-testing android-architecture ios-device ios-build ios-testing ios-architecture kmp flutter react-native)

for file in $docs; do
  [ ! -f "$file" ] && continue
  for name in "${old_names[@]}"; do
    count=$(grep -c "\`$name\`" "$file" 2>/dev/null || echo 0)
    [ "$count" -gt 0 ] && echo "$file : \`$name\` x$count"
  done
done
```

- [ ] **Step 8.5.2: Reemplazar backticked references**

```bash
for file in $docs; do
  [ ! -f "$file" ] && continue
  for name in "${old_names[@]}"; do
    sed -i "s/\`$name\`/\`mobiai-$name\`/g" "$file"
  done
done
```

- [ ] **Step 8.5.3: Revisión manual del README**

El README tiene listas tipo "Flujo de trabajo — fix-issue, reproduce-bug, ..." sin backticks. Editar manualmente con Edit tool si se detectan casos.

### Task 8.6: Validar marketplace

- [ ] **Step 8.6.1: Correr validator**

```bash
claude plugin validate .
```
Expected: `✔ Validation passed`

### Task 8.7: Commit del rename

- [ ] **Step 8.7.1: Ver magnitud del cambio**

```bash
git diff --stat | tail -5
```

Expected: ~60+ archivos tocados (29 renames + 29 frontmatter edits + cross-refs + docs).

- [ ] **Step 8.7.2: Commit**

```bash
git add -A
git commit -m "feat(skills): prefijar skills propios con mobiai- para distinguirlos

Todos los skills mantenidos por MobiAI reciben prefijo mobiai-. Cuando Claude
dice 'Using skill X', el usuario sabe al toque si es nuestro o de Google.

Renames: 29 skills (18 en core, 4 en android, 4 en ios, 1 en kmp/flutter/rn).
Excepciones:
- using-mobiai: ya contiene la marca, renombrarlo es redundante y el
  hook session-start referencia su path.
- plugins/android/skills/google/*: son de Google, mantienen nombres upstream
  para que el auto-sync no genere diffs espurios.

Cambios colaterales:
- Frontmatter name: de cada SKILL.md actualizado para matchear folder name
- Cross-references entre skills actualizadas (ej. fix-issue → mobiai-fix-issue)
- Quick Decision Guide en using-mobiai/SKILL.md re-escrita
- CLAUDE.md, README.md, CONTRIBUTING.md, docs/crear-skills.md actualizados

UX: /plugin list sigue mostrando plugins; cuando Claude invoca skills el
output es 'Using skill mobiai-fix-issue' vs 'Using skill edge-to-edge' —
origen visible sin leer descriptions."
```

### Task 8.8: Verificación manual post-rename

- [ ] **Step 8.8.1: El usuario recarga plugins y prueba un skill nuestro y uno de Google**

```
/plugin marketplace update mobiai
/reload-plugins
```

En nueva sesión:
- Pedile al agente algo que dispare un skill nuestro — debería verse "Using skill mobiai-<name>"
- Pedile algo que dispare un skill de Google (ej. "Ayúdame con edge-to-edge en Compose") — debería verse "Using skill edge-to-edge" (sin prefijo)

Esperado: contraste visual inmediato entre skills de ambos orígenes.

---

## Self-review

**1. Spec coverage:**

| Requisito del user prompt | Cubierto en |
|---|---|
| Eliminar entrada `android-official-skills` del marketplace | Task 2.1 |
| Eliminar dep en `android/plugin.json` | Task 2.2 |
| 6 skills de Google vivir bajo `plugins/android/skills/google/` | Task 1.3 |
| LICENSE.txt copiado verbatim | Task 1.2.2 |
| README.md de Google copiado | Task 1.2.3 |
| NOTICE.md nuevo en raíz | Task 3.1 |
| README actualizado con mención a Google | Task 3.2 |
| Workflow refactoreado a vendoring | Task 4.1 |
| Skills array en plugin.json `["./skills/", "./skills/google/"]` | Task 2.2.2 |
| `/plugin list` post-install sin `android-official-skills` | Task 5.3.3 |
| Prefijo `mobiai-` en skills propios | Phase 8 (todas las tasks) |
| Cross-references y docs actualizadas post-rename | Tasks 8.3, 8.4, 8.5 |
| Excepciones del rename documentadas | Phase 8 §"Alcance del rename" |

Todo cubierto. ✓

**2. Placeholder scan:**
- Sin "TBD", "implement later", "handle edge cases" genéricos.
- Código exacto, paths exactos, commits exactos.
- Excepción: Task 5.2 depende de que el usuario abra Claude Code — eso es inevitable (validación manual).

**3. Type consistency:**
- Nombres de skill consistentes (agp-9-upgrade, edge-to-edge, etc.) a lo largo del plan.
- Paths consistentes (`plugins/android/skills/google/<name>/SKILL.md`).
- Commit messages siguen formato Conventional Commits del proyecto.

**4. Riesgos reconocidos:**
- Cambio de LICENSE upstream — workflow falla explícito (Task 4.1)
- Skills removidas upstream — workflow detecta y borra (Task 4.1)
- Colisión de nombres — inventario muestra sin colisión hoy; regla documentada en CONTRIBUTING sería un follow-up
- `gh` CLI no disponible — fallback a paste manual documentado

---

## Handoff

Plan guardado en `docs/plans/2026-04-22-vendor-android-official-skills.md`.

**Dos opciones de ejecución:**

**1. Subagent-Driven (recomendado)** — dispatch subagent fresco por task, review entre tareas. Alcance moderado (8 fases, ~28 tareas), varias con componente de validación. Minimiza riesgo de momentum drift.

**2. Inline Execution** — ejecución directa en esta sesión con checkpoints. Más rápido para un scope contenido como este.

**¿Cuál preferís?**
