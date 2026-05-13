# MobiAI Brain (preview)

MobiAI Brain es la **memoria viva por proyecto** del ecosistema MobiAI. Mientras las skills enseñan a la IA *cómo* trabajar, el Brain guarda el *conocimiento específico* de cada proyecto mobile: decisiones, bugfixes, workarounds, patrones de testing e integraciones.

> Comandos disponibles hoy: `init`, `scan`, `context` (con filtros), `save decision|bugfix|testing`, `search`. Próximos pasos en el [Roadmap](#roadmap).

## ¿Por qué Brain?

Las skills son metodología. CLAUDE.md / AGENTS.md / GEMINI.md son reglas estables. Pero todos los proyectos tienen *conocimiento histórico cambiante*:

- "Usamos Koin, no Hilt — decisión de marzo 2026, sigue activa."
- "FirebaseAuth iOS necesita mantener `composeApp` en minúsculas hasta el próximo major de CocoaPods. Es temporal."
- "El test de DataStore tiene que esperar `dataStore.data.first { it.asMap().isEmpty() }` después de `clear`."

Eso no entra cómodo en CLAUDE.md (no es regla estable) ni en una skill (no es metodología). Entra en el Brain.

## Filosofía

- **Por proyecto**, no global. Cada repo mobile tiene su propio cerebro local en `<repo>/.mobiai/brain/`.
- **Local primero**. Markdown plano, JSON simple. Sin SQLite, embeddings ni nube por ahora.
- **Mobile-first**. El valor diferencial es entender Android / iOS / KMP / Flutter / React Native, no guardar texto genérico.
- **Sin secretos**. El scanner detecta archivos sensibles (`.env`, `GoogleService-Info.plist`, `google-services.json`) y solo registra su existencia, nunca su contenido.

## Comandos disponibles

```bash
mobiai brain init                  # Crea .mobiai/brain/ en el proyecto actual (idempotente)
mobiai brain scan                  # Detecta stack: Android, iOS, KMP, Flutter, RN, librerías, CI/CD
mobiai brain context               # Imprime Markdown con config + scan + memorias (acepta filtros)

mobiai brain save decision ...     # Guarda una decisión de arquitectura
mobiai brain save bugfix ...       # Guarda un bugfix o workaround
mobiai brain save testing ...      # Guarda un patrón de testing reusable

mobiai brain search <query>        # Busca texto libre en las memorias

mobiai brain mcp                   # Arranca un server MCP que expone el Brain como tools
```

Si tu cliente (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI) tiene MCP soportado, registralo desde [`MCP-SETUP.md`](MCP-SETUP.md) y el agente invoca las operaciones del Brain directamente como tools (`mobile_context`, `mobile_search`, `mobile_save_*`, `mobile_scan`) en vez de shellear la CLI.

Todos aceptan `--root <ruta>` para apuntar a un proyecto distinto del directorio actual.

### `mobiai brain init`

Crea la estructura local:

```
<repo>/.mobiai/brain/
├── config.json
└── memories/
    ├── decisions.md
    ├── bugfixes.md
    ├── testing.md
    ├── integrations.md
    └── releases.md
```

Es idempotente: si `config.json` o cualquier `memories/*.md` ya existe, no se sobrescribe.

Para localizar la raíz del proyecto, sube por los ancestros del cwd buscando, en este orden de prioridad:

1. `.mobiai/brain/config.json` (un brain ya inicializado).
2. `.git/`.
3. `settings.gradle.kts` / `settings.gradle` (Android / KMP).
4. `pubspec.yaml` (Flutter).
5. `Package.swift`.
6. `*.xcworkspace` / `*.xcodeproj`.
7. `Podfile`.
8. `package.json` con dependencia `react-native`.

Si nada coincide, usa el directorio actual y avisa con un warning.

### `mobiai brain scan`

Recorre el árbol del proyecto (profundidad máx. 6, ignorando `.git`, `node_modules`, `build`, `Pods`, `DerivedData`, `.dart_tool`, etc.) y produce `.mobiai/brain/scan.json` con:

- `project_type`: `android` | `ios` | `kmp` | `flutter` | `react_native` | `unknown`.
- `platforms`: `android`, `ios`, `shared` (lo que se haya detectado).
- `build_systems`: `gradle`, `cocoapods`, `spm`, `npm`.
- `ui` / `di` / `network` / `persistence` / `serialization`: librerías reconocidas (Compose, SwiftUI, Koin, Hilt, Ktor, Retrofit, Room, DataStore, kotlinx.serialization, ...).
- `testing`: JUnit, MockK, Mockito, Turbine, Espresso, ...
- `integrations`: Firebase y otras integraciones detectables por strings en `build.gradle.kts` / `Podfile` / `package.json`.
- `ci_cd`: GitHub Actions, Bitrise, Codemagic, Fastlane.
- `warnings`: archivos sensibles detectados (sin leer su contenido).

`scan` requiere haber corrido `init` antes.

### `mobiai brain save <type>`

Tres subcomandos: `decision`, `bugfix`, `testing`. Todos comparten los mismos flags:

| Flag | ¿Requerido? | Notas |
|---|---|---|
| `--title <str>` | sí | Título corto, se usa como heading H2 |
| `--platform <plat>` | no | `android` \| `ios` \| `shared` \| `kmp` \| `flutter` \| `react-native` |
| `--area <str>` | no | Libre (`firebase_auth`, `dependency_injection`, `datastore`, ...) |
| `--status <s>` | no | `active` (default) \| `temporary` \| `deprecated` |
| `--review-after YYYY-MM-DD` | no | Sentido sobre todo en `temporary` |
| `--files a,b,c` | no | Lista de paths relevantes |
| `--body <md>` | no | Cuerpo Markdown. Si se omite, lee de stdin (útil para pipear contenido multilínea). |

Ejemplo:

```bash
mobiai brain save decision \
  --title "Use Koin como DI" \
  --platform shared \
  --area dependency_injection \
  --files "composeApp/src/commonMain/di/Module.kt" \
  --body "### Decision
Usar Koin como framework de DI para todo el código compartido.

### Reason
KMP-friendly, sin code generation, ya integrado con composeApp/iosApp."
```

**Guard**: si no existe `.mobiai/brain/config.json` el comando falla con un error claro y exit code 1. No crea el brain solo — sugiere correr `mobiai brain init` primero.

El tipo interno (`type:` en la entrada renderizada) se deriva del subcomando + status:
- `decision` → `architecture_decision`
- `bugfix` + `temporary` → `platform_workaround`
- `bugfix` + `active`/`deprecated` → `bug_fix`
- `testing` → `testing_pattern`

Cada entrada recibe un `id` estable basado en el slug del título + timestamp UTC, así dos saves con el mismo título no colisionan.

### `mobiai brain search <query>`

Búsqueda case-insensitive sobre el título y el cuerpo de cada entrada de las memorias. Devuelve por sección, con título, status/platform y un snippet de la primera línea que contiene la query.

```bash
mobiai brain search koin
mobiai brain search --platform ios firebase
mobiai brain search --status temporary
mobiai brain search --area firebase_auth keystore
```

Acepta los mismos filtros que `context` (`--platform`, `--status`, `--area`) con semántica **AND**. La query se combina con los filtros: `--platform ios firebase` devuelve solo entradas iOS que mencionen "firebase".

### `mobiai brain context`

Lee `config.json`, `scan.json` (si existe) y todas las memorias. Imprime un Markdown listo para agentes:

```
# MobiAI Brain Context

Project: taykus-kmp
Type: Kotlin Multiplatform
Platforms: android, ios, shared

## Detected Stack
- UI: compose_multiplatform
- DI: koin
- Network: ktor
- Build systems: gradle

## Project Rules
- Prioritize clean architecture
- ...

## Architecture Decisions
...

## Known Bugfixes
...

## Testing Patterns
...

## Integrations
...

## Release Notes
...

## Warnings
- archivo sensible detectado: .env (no leído)
```

**Filtros disponibles**:

| Flag | Notas |
|---|---|
| `--section` | Coma-separado o repetible. Nombres canónicos: `stack`, `rules`, `decisions`, `bugfixes`, `testing`, `integrations`, `releases`, `warnings`. |
| `--platform` | Filtra entradas por `platform:` (exacto, case-insensitive). |
| `--status` | Filtra por `status:` (exacto). |
| `--area` | Filtra por `area:` (substring). |

Los filtros aplican solo a entradas de memorias — `stack`, `rules` y `warnings` se incluyen/excluyen únicamente vía `--section`. Si tras filtrar una sección no quedan entradas, aparece `_No entries match the current filter._` (distinto del `_No entries yet._` para secciones genuinamente vacías).

```bash
# Solo bugfixes temporales para iOS
mobiai brain context --section bugfixes --platform ios --status temporary

# Solo el stack detectado, sin memorias
mobiai brain context --section stack

# Decisiones + bugfixes de plataforma shared
mobiai brain context --section decisions,bugfixes --platform shared
```

## Diferencia frente a otros conceptos

| | Skills | CLAUDE.md / AGENTS.md / GEMINI.md | MobiAI Brain |
|---|---|---|---|
| **Qué guarda** | Metodología, buenas prácticas | Reglas estables del proyecto | Conocimiento histórico cambiante |
| **Alcance** | Global (compartido entre proyectos) | Por proyecto | Por proyecto |
| **Estructura** | SKILL.md | Markdown libre | Markdown estructurado + JSON |
| **Caduca?** | No (se actualiza vía PR) | Pocas veces | A menudo (workarounds temporales, decisiones revisables) |

## Integración con skills

Tres skills proponen guardar al final de su flujo, todas con el mismo patrón: **solo si** existe `.mobiai/brain/config.json` en el proyecto, y **siempre** pidiendo aprobación al usuario antes de invocar el save.

| Skill | Cuándo propone | Tipo guardado |
|---|---|---|
| `mobiai-fix-issue` | Tras Phase 6 (verification), antes del PR gate | `bugfix` (status según haya fix real o workaround) |
| `mobiai-write-tests` | Tras Step 4 (tests pasan), si el test captura un patrón no obvio | `testing` |
| `mobiai-mobile-debugging` | Tras Phase 6 (root cause confirmada), **solo en invocación standalone** — si se invocó desde fix-issue, se salta (fix-issue ya tiene su propio hook) | `bugfix` |
| `mobiai-mobile-brainstorming` | Tras User Review Gate (spec aprobada), una entrada por cada decisión arquitectónica no trivial del spec | `decision` |

Si el proyecto **no** tiene brain, el hook se salta sin ruido. El agente nunca invoca `save` solo: siempre pide una línea de confirmación al usuario antes de ejecutarlo.

Otras skills que podrían integrarse en el futuro: `mobiai-crashlytics` (crashes recurrentes → `save bugfix`), `mobiai-mobile-planning` (planes con decisiones embebidas → `save decision`).

## Roadmap

**Fase 1** — `init`, `scan`, `context`. ✓
**Fase 2** — `save decision|bugfix|testing` + hook en `mobiai-fix-issue`. ✓
**Fase 3** — `search` + filtros (`--section`, `--platform`, `--status`, `--area`) en `context`. ✓
**Fase 4** — hooks de save en `mobiai-write-tests` (testing pattern), `mobiai-mobile-debugging` (bugfix sin pasar por fix-issue) y `mobiai-mobile-brainstorming` (decision tras spec aprobada). ✓
**Fase 5** — servidor MCP (`mobiai brain mcp`) que expone las 6 operaciones como tools nativas para Claude Code, Cursor, Copilot CLI, Codex y Gemini CLI. El brain pasa a estar siempre en el toolbox del agente. Setup en [`MCP-SETUP.md`](MCP-SETUP.md). ✓

**Próximos pasos**:
- `mobiai brain review` — listar entradas `status: temporary` cuyo `review_after` ya pasó (para que workarounds temporales no se vuelvan permanentes por inercia).
- `save integration` y `save release` (categorías de menor volumen, baja prioridad).
- Migrar los hooks de las skills (fix-issue, write-tests, mobile-debugging, brainstorming) a invocar las tools MCP directamente cuando estén disponibles, en lugar de shellear out al binario. Hoy coexisten — la CLI es el fallback universal.

**Fase futura** — migración a SQLite + FTS5 si Markdown se queda corto (>~500 entradas por proyecto). Para el rango actual (1-100 entradas) los filtros sobre Markdown bastan.
