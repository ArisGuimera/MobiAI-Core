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
# Mediante el sistema de plugins de Claude Code
claude plugin install ArisGuimera/MobiAI-Core
```

## Skills disponibles

### Flujo de trabajo
| Skill | Descripción |
|-------|-------------|
| [fix-issue](skills/fix-issue/SKILL.md) | Pipeline completo: obtener issue, analizar, corregir, testear, PR |
| [reproduce-bug](skills/reproduce-bug/SKILL.md) | Reproducir bugs en dispositivo/emulador/simulador |
| [analyze-crash](skills/analyze-crash/SKILL.md) | Analizar un crash a partir de cualquier fuente y encontrar la causa raíz |
| [crashlytics](skills/crashlytics/SKILL.md) | Investigar un crash de Firebase Crashlytics en profundidad |
| [write-tests](skills/write-tests/SKILL.md) | Escribir tests con las convenciones de cada plataforma |
| [review-code](skills/review-code/SKILL.md) | Code review mobile: lifecycle, memoria, thread safety |
| [create-pr](skills/create-pr/SKILL.md) | Crear PRs con contexto mobile y plan de test |

### Android
| Skill | Descripción |
|-------|-------------|
| [android-device](skills/android-device/SKILL.md) | ADB, emulador, automatización de UI, screenshots, logcat |
| [android-build](skills/android-build/SKILL.md) | Gradle, flavors, firmado, ProGuard/R8 |
| [android-testing](skills/android-testing/SKILL.md) | JUnit, MockK, Espresso, Compose Testing |
| [android-architecture](skills/android-architecture/SKILL.md) | Clean Architecture, MVVM, MVI, Compose |

### iOS
| Skill | Descripción |
|-------|-------------|
| [ios-device](skills/ios-device/SKILL.md) | Simulador vía simctl, screenshots, logs |
| [ios-build](skills/ios-build/SKILL.md) | xcodebuild, schemes, CocoaPods, SPM |
| [ios-testing](skills/ios-testing/SKILL.md) | XCTest, Quick/Nimble, snapshot testing |
| [ios-architecture](skills/ios-architecture/SKILL.md) | SwiftUI, UIKit, TCA, MVVM+Combine |

### Multiplataforma
| Skill | Descripción |
|-------|-------------|
| [kmp](skills/kmp/SKILL.md) | Kotlin Multiplatform |
| [flutter](skills/flutter/SKILL.md) | Flutter/Dart |
| [react-native](skills/react-native/SKILL.md) | React Native |

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
