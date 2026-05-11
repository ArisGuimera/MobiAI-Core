# MobiAI-Core

Evoluciona tu stack móvil con Inteligencia Artificial.

## ¿Qué es MobiAI?

MobiAI es un ecosistema Open Source diseñado para llevar la Inteligencia Artificial al corazón de tu flujo de trabajo. Más que una librería de funciones, es una plataforma evolutiva que automatiza lo complejo para que tú te centres en crear.

- **Skills** — Contexto experto que adapta lo que la IA ya sabe al escenario adecuado. Los skills no le enseñan a programar — le dan el contexto, las herramientas y los flujos específicos de cada plataforma para que aplique su conocimiento de forma precisa. Corregir bugs, reproducir incidencias, escribir tests, analizar crashes, crear PRs... todo siguiendo las mejores prácticas de cada plataforma.
- **Brain** - El cerebro contextual de cada proyecto: entiende tu stack mobile, recuerda decisiones, bugfixes y patrones clave, y entrega a la IA contexto preciso y filtrado para trabajar alineada con tu arquitectura, evitando ruido y ahorrando tokens.
- **Agentes** — Agentes especializados que ejecutan tareas complejas de forma autónoma: un agente que analiza código, otro que interactúa con el dispositivo, otro que escribe tests.
- **Pipeline automatizado** — Un flujo completo de corrección de bugs: recibe un error o ticket desde cualquier plataforma de mensajería, reproduce el bug en un emulador, encuentra la causa raíz, aplica el fix, ejecuta tests y crea el PR. Todo sin intervención humana.

Compatible con **Claude Code**, **Cursor**, **Copilot CLI**, **Codex** y **Gemini CLI**.

## Estado actual

| Componente | Estado |
|------------|--------|
| Skills (multi-plataforma) | Estable |
| Brain | Alpha |
| Agentes autónomos | En desarrollo |
| Pipeline automatizado (bot de mensajería) | Planificado |
| Marketplace de skills comunitarios | Planificado |

## Instalación

MobiAI se organiza en plugins granulares. Instala sólo las plataformas que uses, o instala el meta-plugin `mobile` para todo el stack.

### Claude Code

```bash
# 1. Agregá el marketplace
/plugin marketplace add ArisGuimera/MobiAI-Core

# 2a. Instalá todo (recomendado)
/plugin install mobile@mobiai

# 2b. O instalá sólo la plataforma que necesites
/plugin install android@mobiai      # Android + deps (core + skills oficiales de Google)
/plugin install ios@mobiai           # iOS + deps
/plugin install kmp@mobiai           # KMP (trae Android e iOS)
/plugin install flutter@mobiai       # Flutter
/plugin install react-native@mobiai  # React Native
```

### Cursor / Copilot CLI

Misma secuencia de comandos (`/plugin marketplace add` + `/plugin install`).

### Gemini CLI

```bash
gemini extension install ArisGuimera/MobiAI-Core
```

Gemini instala la extension completa (todos los plugins vienen juntos).

### Codex

Consultá [`.codex/INSTALL.md`](.codex/INSTALL.md) para instrucciones con symlinks.

## Actualizar

### Claude Code / Cursor / Copilot CLI

```bash
/plugin update mobile@mobiai   # o el plugin específico que instalaste
```

Los skills oficiales de Google (`android-official-skills`) se actualizan automáticamente vía el marketplace.

### Gemini CLI

```bash
gemini extension update mobiai-core
```

### Codex

```bash
cd ~/.codex/mobiai-core && git pull
```

## Plugins disponibles

| Plugin | Incluye | Comando |
|---|---|---|
| `mobile` | Todo (meta) | `/plugin install mobile@mobiai` |
| `android` | Skills Android + skills oficiales de Google | `/plugin install android@mobiai` |
| `ios` | Skills iOS | `/plugin install ios@mobiai` |
| `mobiai-kmp` | KMP (incluye Android + iOS como deps) | `/plugin install kmp@mobiai` |
| `mobiai-flutter` | Flutter / Dart | `/plugin install flutter@mobiai` |
| `mobiai-react-native` | React Native | `/plugin install react-native@mobiai` |

## Skills disponibles

MobiAI incluye skills para todo el ciclo de desarrollo mobile:

- **Flujo de trabajo** (en `core`, auto-instalado con cualquier plataforma) — mobiai-fix-issue, mobiai-reproduce-bug, mobiai-analyze-crash, mobiai-crashlytics, mobiai-write-tests, mobiai-review-code, mobiai-create-pr
- **Proceso** (en `core`) — mobiai-mobile-brainstorming, mobiai-mobile-debugging, mobiai-mobile-tdd, mobiai-mobile-planning, mobiai-mobile-verification, mobiai-mobile-executing-plans, mobiai-mobile-parallel-agents, mobiai-mobile-executing-plans-with-subagents, mobiai-mobile-worktrees, mobiai-mobile-finishing-branch, mobiai-writing-skills
- **Android** (en `android`) — `mobiai-android-device`, `mobiai-android-build`, `mobiai-android-testing`, `mobiai-android-architecture`, más los [skills oficiales de Android mantenidos por Google](https://github.com/android/skills) (Apache 2.0, vendoreados en `plugins/android/skills/google/` con auto-sync semanal). Ver [NOTICE.md](NOTICE.md).
- **iOS** (en `ios`) — mobiai-ios-device, mobiai-ios-build, mobiai-ios-testing, mobiai-ios-architecture
- **Multiplataforma** — mobiai-kmp, mobiai-flutter, mobiai-react-native (cada uno en su plugin)

Consulta el [catálogo completo de skills](plugins/core/skills/using-mobiai/SKILL.md) para ver cuándo usar cada uno.

## MobiAI Brain (preview)

`mobiai brain` añade **memoria viva por proyecto** a la CLI: decisiones, bugfixes, workarounds e integraciones específicas de cada repo mobile, separadas del estado global de MobiAI.

```bash
mobiai brain init      # crea <repo>/.mobiai/brain/ (idempotente)
mobiai brain scan      # detecta stack: Android/iOS/KMP/Flutter/RN + librerías
mobiai brain context   # imprime Markdown listo para agentes
```

Detalle completo en [docs/brain.md](docs/brain.md). Los subcomandos `save` y `search` llegan en una próxima fase.

## ¿Cómo funciona?

1. **Instala el plugin** — MobiAI registra sus skills en tu asistente de IA (Claude Code, Cursor, Copilot CLI, Codex o Gemini CLI)
2. **Empieza a programar** — Los skills se activan automáticamente según el contexto
3. **Detección de plataforma** — MobiAI detecta tu proyecto (Android, iOS, Flutter...) y aplica el contexto adecuado
4. **Tu proyecto, tus reglas** — Los skills respetan las convenciones de tu proyecto (`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`)

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
