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

MobiAI es un ecosistema Open Source diseñado para llevar la Inteligencia Artificial al corazón de tu flujo de trabajo mobile. Cuatro piezas que encajan entre sí — pulsa en cualquiera para ver el detalle:

| | | |
|---|---|---|
| 🛠️ **[CLI `mobiai`](cli/README.md)** | Una sola herramienta para gestionar todo el stack: instalar skills, indexar código, gestionar la memoria del proyecto y mantenerlo todo actualizado. | [Ver detalle →](cli/README.md) |
| 🧩 **[Skills](skills/README.md)** | Contexto experto por plataforma: corregir bugs, reproducir incidencias, escribir tests, analizar crashes, crear PRs… siguiendo las prácticas de cada stack. | [Ver detalle →](skills/README.md) |
| 🔍 **[Graph](docs/graph/README.md)** | Exploración semántica del código mobile. Resuelve "qué archivos tocar para esta tarea" o "quién llama a este símbolo" sin recorrer el repo a ciegas. | [Ver detalle →](docs/graph/README.md) |
| 🧠 **[Brain](brain/README.md)** | Memoria viva por proyecto: decisiones, bugfixes, workarounds e integraciones específicas, separadas del estado global de MobiAI. | [Ver detalle →](brain/README.md) |

Y dos piezas más en el horizonte:

- **Agentes** — Agentes especializados que ejecutan tareas complejas de forma autónoma (en desarrollo).
- **Pipeline automatizado** — Bot que recibe tickets desde tu plataforma de mensajería, reproduce el bug, aplica el fix y abre el PR (planificado).

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

Verifica que ha quedado instalada:

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

Mantener al día (`mobiai update` refresca el catálogo) y diagnóstico (`mobiai doctor`, `mobiai status`) viven dentro de la CLI — ver [cli/README.md](cli/README.md).

## ¿Cómo funciona?

1. **Instala la CLI `mobiai`** — un solo binario que detecta tu asistente (Claude Code, Cursor, Copilot CLI, Codex o Gemini CLI) y registra los skills en el lugar correcto
2. **Elige tus packs** — `mobiai skills init` te abre el selector, o instala lo que necesites con `mobiai skills add <pack>`
3. **Empieza a programar** — Los skills se activan automáticamente según el contexto; opcionalmente indexa el repo con `mobiai graph init` y arranca la memoria con `mobiai brain init`
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
