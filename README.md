<div align="center">
  <img src="docs/assets/mobiai-hero.png" alt="MobiAI — Brain · Skills · Graph para Android, iOS, React Native, Flutter y KMP" width="100%">
</div>

# MobiAI-Core

<div align="center">

<img src="https://img.shields.io/badge/Android-3DDC84?style=for-the-badge&logo=android&logoColor=white" alt="Android">
<img src="https://img.shields.io/badge/iOS-000000?style=for-the-badge&logo=apple&logoColor=white" alt="iOS">
<img src="https://img.shields.io/badge/Kotlin_Multiplatform-7F52FF?style=for-the-badge&logo=kotlin&logoColor=white" alt="Kotlin Multiplatform">
<img src="https://img.shields.io/badge/Flutter-02569B?style=for-the-badge&logo=flutter&logoColor=white" alt="Flutter">
<img src="https://img.shields.io/badge/React_Native-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React Native">

<img src="https://img.shields.io/badge/Kotlin-7F52FF?style=for-the-badge&logo=kotlin&logoColor=white" alt="Kotlin">
<img src="https://img.shields.io/badge/Swift-F05138?style=for-the-badge&logo=swift&logoColor=white" alt="Swift">
<img src="https://img.shields.io/badge/Dart-0175C2?style=for-the-badge&logo=dart&logoColor=white" alt="Dart">
<img src="https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript">
<img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">

<img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="License: MIT">
<img src="https://img.shields.io/badge/Open_Source-❤-red?style=for-the-badge" alt="Open Source">

</div>

Evoluciona tu stack móvil con Inteligencia Artificial.

## ¿Qué es MobiAI?

MobiAI es un ecosistema Open Source diseñado para llevar la Inteligencia Artificial al corazón de tu flujo de trabajo. Más que una librería de funciones, es una plataforma evolutiva que automatiza lo complejo para que tú te centres en crear.

- **CLI `mobiai`** — Una sola herramienta para gestionar todo el stack: instalar skills, indexar tu código (Graph), mantener la memoria por proyecto (Brain) y mantener todo actualizado en cualquier asistente compatible.
- **Skills** — Contexto experto que adapta lo que la IA ya sabe al escenario adecuado. Los skills no le enseñan a programar — le dan el contexto, las herramientas y los flujos específicos de cada plataforma para que aplique su conocimiento de forma precisa. Corregir bugs, reproducir incidencias, escribir tests, analizar crashes, crear PRs... todo siguiendo las mejores prácticas de cada plataforma.
- **Graph** — Exploración semántica del código mobile. Indexa tu proyecto y resuelve preguntas tipo "qué archivos tocar para esta tarea" o "quién llama a este símbolo" sin recorrer el repo a ciegas.
- **Brain** — El cerebro contextual de cada proyecto: entiende tu stack mobile, recuerda decisiones, bugfixes y patrones clave, y entrega a la IA contexto preciso y filtrado para trabajar alineada con tu arquitectura, evitando ruido y ahorrando tokens.
- **Agentes** — Agentes especializados que ejecutan tareas complejas de forma autónoma: un agente que analiza código, otro que interactúa con el dispositivo, otro que escribe tests.
- **Pipeline automatizado** — Un flujo completo de corrección de bugs: recibe un error o ticket desde cualquier plataforma de mensajería, reproduce el bug en un emulador, encuentra la causa raíz, aplica el fix, ejecuta tests y crea el PR. Todo sin intervención humana.

Compatible con **Claude Code**, **Cursor**, **Copilot CLI**, **Codex** y **Gemini CLI**.

## Estado actual

| Componente | Estado |
|------------|--------|
| CLI `mobiai` | Estable |
| Skills (multi-plataforma) | Estable |
| Graph (indexador semántico) | Alpha |
| Brain (memoria por proyecto) | Alpha |
| Agentes autónomos | En desarrollo |
| Pipeline automatizado (bot de mensajería) | Planificado |
| Marketplace de skills comunitarios | Planificado |

## Instalación

MobiAI se instala como una CLI standalone (`mobiai`) que detecta tu asistente (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI) y empuja los skills al lugar correcto.

```bash
# Mac / Linux
curl -fsSL https://mobiai.dev/install.sh | sh

# Windows (cmd)
curl.exe -fsSL https://mobiai.dev/install.cmd -o "%TEMP%\i.cmd" && "%TEMP%\i.cmd"

# Windows (PowerShell)
iwr -useb https://mobiai.dev/install.ps1 | iex
```

Verificá que quedó instalada:

```bash
mobiai --version
mobiai doctor      # diagnostica hosts detectados, catálogo y permisos
```

Detalle completo de instalación, build local y releases en [cli/README.md](cli/README.md).

### Primer uso

```bash
mobiai             # banner + ayuda
mobiai skills init # selector interactivo para elegir packs (mobile, android, ios, kmp, flutter, react-native)
```

## Actualizar

`mobiai update` refresca el catálogo de skills desde el remoto. El binario `mobiai` se actualiza re-ejecutando el script de instalación.

```bash
mobiai update          # refresca el catálogo
mobiai status          # ver qué hosts están conectados y qué packs están instalados
```

Dentro de Claude Code también podés disparar `/mobiai-update` desde cualquier sesión — un notificador en `SessionStart` te avisa si hay catálogo nuevo.

## La CLI en tres bloques

### `mobiai skills` — gestionar skills

Instala, lista y desinstala packs de skills en los hosts detectados. Hay un meta-pack `mobile` que trae todo, o packs por plataforma.

```bash
mobiai skills init                 # selector interactivo
mobiai skills add mobile           # instala todo el stack
mobiai skills add android ios      # solo las plataformas que usás
mobiai skills list                 # qué tenés instalado
mobiai skills remove flutter       # desinstalar un pack
```

Packs disponibles: `mobile` (meta), `android`, `ios`, `kmp`, `flutter`, `react-native`. Detalle del flujo, hosts soportados y troubleshooting en [cli/README.md](cli/README.md). Catálogo completo de skills en [skills/core/skills/using-mobiai/SKILL.md](skills/core/skills/using-mobiai/SKILL.md).

### `mobiai graph` — exploración semántica del código

Indexa tu proyecto mobile y responde preguntas tipo "qué archivos tocar para esta tarea" o "quién llama a este símbolo" sin que el agente tenga que recorrer el repo a ciegas. Los skills downstream (debugging, fix-issue, planning) hacen pre-flight de Graph automáticamente cuando está disponible.

```bash
mobiai graph init                  # indexa el proyecto en .mobiai/graph/
mobiai graph status                # info del índice actual
mobiai graph search <símbolo>      # buscar símbolos por nombre
mobiai graph callers <símbolo>     # referencias textuales
mobiai graph context <tarea...>    # archivos relevantes para una tarea
```

Visión completa, formato del índice y roadmap en [docs/graph.md](docs/graph.md).

### `mobiai brain` — memoria por proyecto

Memoria viva por repo: decisiones, bugfixes, workarounds e integraciones específicas, separadas del estado global de MobiAI. Se guarda en `<repo>/.mobiai/brain/`.

```bash
mobiai brain init                  # crea .mobiai/brain/ (idempotente)
mobiai brain scan                  # detecta stack: Android/iOS/KMP/Flutter/RN + librerías
mobiai brain context               # imprime Markdown listo para agentes
mobiai brain search <query>        # busca texto libre en memorias
mobiai brain review                # entradas temporales que toca revisar
```

Detalle completo, formato de las memorias y workflow de save en [brain/README.md](brain/README.md).

## Skills disponibles

MobiAI incluye skills para todo el ciclo de desarrollo mobile:

- **Flujo de trabajo** (en `core`, auto-instalado con cualquier plataforma) — mobiai-fix-issue, mobiai-reproduce-bug, mobiai-analyze-crash, mobiai-crashlytics, mobiai-write-tests, mobiai-review-code, mobiai-create-pr
- **Proceso** (en `core`) — mobiai-mobile-brainstorming, mobiai-mobile-debugging, mobiai-mobile-tdd, mobiai-mobile-planning, mobiai-mobile-verification, mobiai-mobile-executing-plans, mobiai-mobile-parallel-agents, mobiai-mobile-executing-plans-with-subagents, mobiai-mobile-worktrees, mobiai-mobile-finishing-branch, mobiai-writing-skills
- **CLI / ecosistema** (en `core`) — mobiai-graph, mobiai-brain, mobiai-update, using-mobiai
- **Android** (en `android`) — `mobiai-android-device`, `mobiai-android-build`, `mobiai-android-testing`, `mobiai-android-architecture`, más los [skills oficiales de Android mantenidos por Google](https://github.com/android/skills) (Apache 2.0, vendoreados en `skills/android/skills/google/` con auto-sync semanal). Ver [NOTICE.md](NOTICE.md).
- **iOS** (en `ios`) — mobiai-ios-device, mobiai-ios-build, mobiai-ios-testing, mobiai-ios-architecture
- **Multiplataforma** — mobiai-kmp, mobiai-flutter, mobiai-react-native (cada uno en su pack)

Consulta el [catálogo completo de skills](skills/core/skills/using-mobiai/SKILL.md) para ver cuándo usar cada uno.

## ¿Cómo funciona?

1. **Instala la CLI `mobiai`** — un solo binario que detecta tu asistente (Claude Code, Cursor, Copilot CLI, Codex o Gemini CLI) y registra los skills en el lugar correcto
2. **Elegí tus packs** — `mobiai skills init` te abre el selector, o instalá lo que necesites con `mobiai skills add <pack>`
3. **Empezá a programar** — Los skills se activan automáticamente según el contexto; opcionalmente indexá el repo con `mobiai graph init` y arrancá memoria con `mobiai brain init`
4. **Detección de plataforma** — MobiAI detecta tu proyecto (Android, iOS, Flutter...) y aplica el contexto adecuado
5. **Tu proyecto, tus reglas** — Los skills respetan las convenciones de tu proyecto (`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`)

## Roadmap

### Fase 1 — Skills (actual)
Plugin multi-plataforma con skills para todas las plataformas mobile. Compatible con Claude Code, Cursor, Copilot CLI, Codex y Gemini CLI. La comunidad puede contribuir nuevos skills.

### Fase 2 — Agentes
Agentes especializados que van más allá de las instrucciones: analizan código de forma autónoma, interactúan con dispositivos, y toman decisiones inteligentes basándose en el contexto.

### Fase 3 — Pipeline automatizado
Un bot que se conecta a tu plataforma de mensajería favorita, recibe errores o tickets, orquesta los agentes, y produce PRs listos para review — el ciclo completo de corrección de bugs sin intervención humana.

### Fase 4 — Marketplace
Un marketplace donde la comunidad publica y descarga skills, agentes y configuraciones para diferentes tipos de proyectos mobile.

## Contribuir

¡Nos encantaría tu ayuda! MobiAI está creado por desarrolladores mobile, para desarrolladores mobile.

- **Añadir un nuevo skill**: Consulta la [guía para crear skills](docs/crear-skills.md)
- **Mejorar skills existentes**: Abre un PR con mejores instrucciones o manejo de casos edge
- **Reportar problemas**: Abre un issue si un skill da mal consejo o le falta algo

Consulta [CONTRIBUTING.md](CONTRIBUTING.md) para más detalles.

## 👀 ¿Quieres estar actualizado/a?

Únete al [discord de la comunidad](https://bit.ly/3bmeQvm) donde tenemos un canal para hablar del proyecto (`🤖-mobiai`). 

## 👨‍💻 Colaboradores

<a href="https://github.com/ArisGuimera/JetpackComposePro/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=ArisGuimera/mobiai-core" />
</a>


---

<div align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
</div>
