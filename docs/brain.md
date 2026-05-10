# MobiAI Brain (preview)

MobiAI Brain es la **memoria viva por proyecto** del ecosistema MobiAI. Mientras las skills enseñan a la IA *cómo* trabajar, el Brain guarda el *conocimiento específico* de cada proyecto mobile: decisiones, bugfixes, workarounds, patrones de testing e integraciones.

> Esta es la Fase 1 (`init` + `scan` + `context`). Los subcomandos `save` y `search` llegan en la Fase 2.

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
mobiai brain init      # Crea .mobiai/brain/ en el proyecto actual (idempotente)
mobiai brain scan      # Detecta stack: Android, iOS, KMP, Flutter, RN, librerías, CI/CD
mobiai brain context   # Imprime Markdown con config + scan + memorias
```

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

## Diferencia frente a otros conceptos

| | Skills | CLAUDE.md / AGENTS.md / GEMINI.md | MobiAI Brain |
|---|---|---|---|
| **Qué guarda** | Metodología, buenas prácticas | Reglas estables del proyecto | Conocimiento histórico cambiante |
| **Alcance** | Global (compartido entre proyectos) | Por proyecto | Por proyecto |
| **Estructura** | SKILL.md | Markdown libre | Markdown estructurado + JSON |
| **Caduca?** | No (se actualiza vía PR) | Pocas veces | A menudo (workarounds temporales, decisiones revisables) |

## Roadmap

**Fase 1 (este PR)** — `init`, `scan`, `context`.

**Fase 2** — `save decision|bugfix|testing|integration|release`, `search`, flags `--section` / `--platform` en `context`.

**Fase 3** — MCP tools (`mobile_context`, `mobile_save_decision`, `mobile_search`, ...) para que el agente llame al Brain directamente.

**Fase futura** — migración a SQLite + FTS5 si Markdown se queda corto.
