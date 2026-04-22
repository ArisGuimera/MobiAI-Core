# MobiAI-Core

El plugin de IA para desarrolladores mobile. Skills, agentes y pipelines automatizados para Android, iOS, KMP, Flutter y React Native.

## ¿Qué es MobiAI?

MobiAI es un ecosistema open source que potencia tu desarrollo mobile con IA. No es solo un conjunto de skills — es una plataforma completa que crece con la comunidad:

- **Skills** — Contexto experto que adapta lo que la IA ya sabe al escenario adecuado. Los skills no le enseñan a programar — le dan el contexto, las herramientas y los flujos específicos de cada plataforma para que aplique su conocimiento de forma precisa. Corregir bugs, reproducir incidencias, escribir tests, analizar crashes, crear PRs... todo siguiendo las mejores prácticas de cada plataforma.
- **Agentes** — Agentes especializados que ejecutan tareas complejas de forma autónoma: un agente que analiza código, otro que interactúa con el dispositivo, otro que escribe tests.
- **Pipeline automatizado** — Un flujo completo de corrección de bugs: recibe un error o ticket desde cualquier plataforma de mensajería, reproduce el bug en un emulador, encuentra la causa raíz, aplica el fix, ejecuta tests y crea el PR. Todo sin intervención humana.

Compatible con **Claude Code**, **Cursor**, **Copilot CLI**, **Codex** y **Gemini CLI**.

## Estado actual

| Componente | Estado |
|------------|--------|
| Skills (multi-plataforma) | Disponible |
| Agentes autónomos | En desarrollo |
| Pipeline automatizado (bot de mensajería) | Planificado |
| Marketplace de skills comunitarios | Planificado |

## Instalación

MobiAI se organiza en plugins granulares. Instalá sólo las plataformas que uses, o instalá el meta-plugin `mobile` para todo el stack.

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
- **Proceso** (en `core`) — mobiai-mobile-brainstorming, mobiai-mobile-debugging, mobiai-mobile-tdd, mobiai-mobile-planning, mobiai-mobile-verification, mobiai-mobile-executing-plans, mobiai-mobile-parallel-agents, mobiai-mobile-subagent-development, mobiai-mobile-worktrees, mobiai-mobile-finishing-branch, mobiai-writing-skills
- **Android** (en `android`) — mobiai-android-device, mobiai-android-build, mobiai-android-testing, mobiai-android-architecture + skills oficiales de Google vía `android-official-skills`
- **iOS** (en `ios`) — mobiai-ios-device, mobiai-ios-build, mobiai-ios-testing, mobiai-ios-architecture
- **Multiplataforma** — mobiai-kmp, mobiai-flutter, mobiai-react-native (cada uno en su plugin)

Consulta el [catálogo completo de skills](plugins/core/skills/using-mobiai/SKILL.md) para ver cuándo usar cada uno.

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

## Comunidad

Únete a la comunidad de desarrolladores mobile que comparten conocimiento y herramientas. Próximamente más información.

## Licencia

MIT — ver [LICENSE](LICENSE)
