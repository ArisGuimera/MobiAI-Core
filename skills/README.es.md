# MobiAI Skills

> **Leer en otro idioma:** [English](README.md) · **Español**

> El catálogo de **skills** que la CLI `mobiai` distribuye a tu asistente (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI). Cada skill es contexto experto que la IA carga bajo demanda para corregir bugs, escribir tests, analizar crashes, crear PRs y mucho más — siguiendo las prácticas de cada plataforma.

**Tabla de contenidos**

- [Cómo se organizan](#cómo-se-organizan)
- [Packs disponibles](#packs-disponibles)
- [Instalación](#instalación)
- [Catálogo de skills](#catálogo-de-skills)
- [Cómo elige el asistente qué skill usar](#cómo-elige-el-asistente-qué-skill-usar)
- [Contribuir un skill nuevo](#contribuir-un-skill-nuevo)
- [Skills de terceros (Google)](#skills-de-terceros-google)

---

## Cómo se organizan

Los skills están agrupados en **packs**. Un pack es una unidad instalable y, salvo `core` y `mobile`, se corresponde con una plataforma.

```
skills/
├── core/              # cross-platform: flujo, proceso, ecosistema, bootstrap
│   ├── hooks/         # SessionStart hook (notificador de updates)
│   └── skills/        # 22 skills cross-platform (using-mobiai, fix-issue, …)
├── android/
│   └── skills/
│       ├── google/    # skills oficiales de Google (Apache 2.0, sync semanal)
│       └── mobiai-android-*
├── ios/
├── kmp/               # depende de android + ios
├── flutter/           # depende de android + ios
├── react-native/      # depende de android + ios
└── mobile/            # meta-pack: trae todo lo de arriba
```

Cada pack contiene un `.claude-plugin/plugin.json` con metadatos y dependencias. Las dependencias se resuelven automáticamente al instalar (por ejemplo, instalar `kmp` arrastra `core`, `android` e `ios`).

## Packs disponibles

| Pack | Incluye | Depende de | Comando |
|---|---|---|---|
| `mobile` | Todo el stack (meta-pack) | core, android, ios, kmp, flutter, react-native | `mobiai skills add mobile` |
| `android` | Skills de Android + skills oficiales de Google | core | `mobiai skills add android` |
| `ios` | Skills de iOS | core | `mobiai skills add ios` |
| `kmp` | Skills de Kotlin Multiplatform | core, android, ios | `mobiai skills add kmp` |
| `flutter` | Skills de Flutter / Dart | core, android, ios | `mobiai skills add flutter` |
| `react-native` | Skills de React Native | core, android, ios | `mobiai skills add react-native` |
| `core` | ⚠️ Interno — skills cross-platform, bootstrap `using-mobiai` y `SessionStart` hook. No se instala suelto | — | — |

> `core` se instala automáticamente como dependencia de cualquier otro pack. No lo elijas a mano.

## Instalación

Los packs se gestionan con `mobiai skills`. Detalle del flujo completo (hosts soportados, troubleshooting, builds locales) en [../cli/README.md](../cli/README.md).

```bash
# Selector interactivo (recomendado la primera vez)
mobiai skills init

# Instalar paquetes concretos
mobiai skills add mobile
mobiai skills add android ios

# Listar lo instalado
mobiai skills list

# Desinstalar
mobiai skills remove flutter
```

## Catálogo de skills

Los skills se agrupan por su rol dentro del flujo de trabajo. El catálogo en tiempo real, con la guía de cuándo invocar cada uno, vive en [core/skills/using-mobiai/SKILL.md](core/skills/using-mobiai/SKILL.md) (se carga automáticamente al inicio de cada sesión).

### Flujo de trabajo (en `core`)

Skills orquestadores que cubren un ciclo end-to-end:

- `mobiai-fix-issue` — del ticket al PR (Jira/Linear/GitHub issues)
- `mobiai-reproduce-bug` — reproducir bugs en device/emulator/simulator
- `mobiai-analyze-crash` — analizar crashes desde stack trace, log o captura
- `mobiai-crashlytics` — investigar issues de Firebase Crashlytics
- `mobiai-write-tests` — añadir tests (unit, UI, regresión)
- `mobiai-review-code` — code review para cambios mobile
- `mobiai-create-pr` — empaquetar el trabajo en un PR bien estructurado

### Proceso (en `core`)

Skills que definen *cómo* abordar el trabajo:

- `mobiai-mobile-brainstorming` — explorar intención y requisitos antes de escribir código
- `mobiai-mobile-debugging` — root-cause analysis como recolección de evidencia
- `mobiai-mobile-tdd` — tests antes de la implementación
- `mobiai-mobile-planning` — convertir una spec en plan paso a paso
- `mobiai-mobile-verification` — gating obligatorio antes de declarar "hecho"
- `mobiai-mobile-executing-plans` — ejecutar un plan escrito
- `mobiai-mobile-executing-plans-with-subagents` — ejecución vía subagentes con review en dos etapas
- `mobiai-mobile-parallel-agents` — orquestar agentes independientes
- `mobiai-mobile-worktrees` — aislar trabajo en worktrees
- `mobiai-mobile-finishing-branch` — cierre/merge/limpieza de la rama
- `mobiai-writing-skills` — crear skills nuevos siguiendo el formato MobiAI

### CLI y ecosistema (en `core`)

- `using-mobiai` — bootstrap del catálogo (se carga al inicio de sesión)
- `mobiai-graph` — interfaz IA ↔ `mobiai graph` (búsqueda semántica de código)
- `mobiai-brain` — interfaz IA ↔ `mobiai brain` (memoria por proyecto)
- `mobiai-update` — comando `/mobiai-update` para refrescar el catálogo

### Plataforma — Android (en `android`)

- `mobiai-android-device` — `adb`, screenshots, logcat, UI automation
- `mobiai-android-build` — Gradle, flavors, variants, troubleshooting
- `mobiai-android-testing` — frameworks de testing en Android
- `mobiai-android-architecture` — estructura de proyectos Android
- Más los [skills oficiales mantenidos por Google](https://github.com/android/skills) (Apache 2.0, vendorizados con sync semanal). Ver [../NOTICE.md](../NOTICE.md).

### Plataforma — iOS (en `ios`)

- `mobiai-ios-device` — `simctl`, simulador, logs
- `mobiai-ios-build` — `xcodebuild`, schemes, SPM/CocoaPods
- `mobiai-ios-testing` — XCTest, snapshot tests, UI tests
- `mobiai-ios-architecture` — estructura de proyectos iOS

### Multiplataforma

Cada uno en su propio pack:

- `mobiai-kmp` — Kotlin Multiplatform (en `kmp`)
- `mobiai-flutter` — Flutter / Dart (en `flutter`)
- `mobiai-react-native` — React Native (en `react-native`)

## Cómo elige el asistente qué skill usar

El bootstrap `using-mobiai` se carga al inicio de cada sesión y contiene la guía de decisión: ante una intención del usuario ("reproduce este bug", "abre un PR", "revisa este código"), el asistente identifica el skill adecuado y lo invoca antes de tocar código.

La idea es que el asistente no improvise: si hay un skill que cubre la tarea, lo usa. Si no hay match, te pregunta antes de actuar.

## Contribuir un skill nuevo

1. Lee la [guía para crear skills](../docs/crear-skills.md).
2. Invoca el skill `mobiai-writing-skills` desde tu asistente — te lleva por la estructura, frontmatter e instrucciones accionables.
3. Coloca el skill en el pack adecuado:
   - Cross-platform o de proceso → `core/skills/`
   - Específico de plataforma → `android/skills/`, `ios/skills/`, etc.
4. Abre un PR. El CI bloquea PRs que añaden o modifican skills sin actualizar la documentación correspondiente.

Detalle de la guía de contribución general en [../CONTRIBUTING.md](../CONTRIBUTING.md).

## Skills de terceros (Google)

`skills/android/skills/google/` contiene los [skills oficiales de Android mantenidos por Google](https://github.com/android/skills) (licencia Apache 2.0). Se sincronizan automáticamente cada semana vía CI. No los edites a mano: tus cambios serían sobrescritos en el próximo sync. Si encuentras un problema, abre el issue upstream en el repositorio de Google.

Atribución completa y términos en [../NOTICE.md](../NOTICE.md).
