# MobiAI-Core

El plugin de IA para desarrolladores mobile. Skills, agentes y pipelines automatizados para Android, iOS, KMP, Flutter y React Native.

## ¿Qué es MobiAI?

MobiAI es un ecosistema open source que potencia tu desarrollo mobile con IA. No es solo un conjunto de skills — es una plataforma completa que crece con la comunidad:

- **Skills** — Contexto experto que adapta lo que la IA ya sabe al escenario adecuado. Los skills no le enseñan a programar — le dan el contexto, las herramientas y los flujos específicos de cada plataforma para que aplique su conocimiento de forma precisa. Corregir bugs, reproducir incidencias, escribir tests, analizar crashes, crear PRs... todo siguiendo las mejores prácticas de cada plataforma.
- **Agentes** — Agentes especializados que ejecutan tareas complejas de forma autónoma: un agente que analiza código, otro que interactúa con el dispositivo, otro que escribe tests.
- **Pipeline automatizado** — Un flujo completo de corrección de bugs: recibe un error o ticket desde cualquier plataforma de mensajería, reproduce el bug en un emulador, encuentra la causa raíz, aplica el fix, ejecuta tests y crea el PR. Todo sin intervención humana.

Empezamos con [Claude Code](https://docs.anthropic.com/en/docs/claude-code) como plataforma principal, pero la visión es que funcione con cualquier herramienta de IA.

## Estado actual

| Componente | Estado |
|------------|--------|
| Skills (Claude Code plugin) | Disponible |
| Agentes autónomos | En desarrollo |
| Pipeline automatizado (bot de mensajería) | Planificado |
| Marketplace de skills comunitarios | Planificado |

## Instalación

```bash
# 1. Agregá el marketplace de MobiAI
/plugin marketplace add ArisGuimera/MobiAI-Core

# 2. Instalá el plugin
/plugin install mobiai-core@mobiai
```

También podés instalarlo directamente desde la CLI:
```bash
claude plugin install mobiai-core@mobiai
```

## Skills disponibles

MobiAI incluye skills para todo el ciclo de desarrollo mobile:

- **Flujo de trabajo** — fix-issue, reproduce-bug, analyze-crash, crashlytics, write-tests, review-code, create-pr
- **Android** — android-device, android-build, android-testing, android-architecture
- **iOS** — ios-device, ios-build, ios-testing, ios-architecture
- **Multiplataforma** — kmp, flutter, react-native

Consultá el [catálogo completo de skills](skills/using-mobiai/SKILL.md) para ver cuándo usar cada uno.

## ¿Cómo funciona?

1. **Instalá el plugin** — MobiAI registra sus skills en tu asistente de IA
2. **Empezá a programar** — Los skills se activan automáticamente según el contexto
3. **Detección de plataforma** — MobiAI detecta tu proyecto (Android, iOS, Flutter...) y aplica el contexto adecuado
4. **Tu proyecto, tus reglas** — Los skills respetan las convenciones de tu `CLAUDE.md`

## Roadmap

### Fase 1 — Skills (actual)
Plugin de Claude Code con skills para todas las plataformas mobile. La comunidad puede contribuir nuevos skills.

### Fase 2 — Agentes
Agentes especializados que van más allá de las instrucciones: analizan código de forma autónoma, interactúan con dispositivos, y toman decisiones inteligentes basándose en el contexto.

### Fase 3 — Pipeline automatizado
Un bot que se conecta a tu plataforma de mensajería favorita, recibe errores o tickets, orquesta los agentes, y produce PRs listos para review — el ciclo completo de corrección de bugs sin intervención humana.

### Fase 4 — Marketplace
Un marketplace donde la comunidad publica y descarga skills, agentes y configuraciones para diferentes tipos de proyectos mobile.

## Contribuir

¡Nos encantaría tu ayuda! MobiAI está creado por desarrolladores mobile, para desarrolladores mobile.

- **Añadir un nuevo skill**: Consultá la [guía para crear skills](docs/crear-skills.md)
- **Mejorar skills existentes**: Abrí un PR con mejores instrucciones o manejo de casos edge
- **Reportar problemas**: Abrí un issue si un skill da mal consejo o le falta algo

Consultá [CONTRIBUTING.md](CONTRIBUTING.md) para más detalles.

## Comunidad

Unite a la comunidad de desarrolladores mobile que comparten conocimiento y herramientas. Próximamente más información.

## Licencia

MIT — ver [LICENSE](LICENSE)
