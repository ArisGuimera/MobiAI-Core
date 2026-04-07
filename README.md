# MobiAI-Core

El plugin de IA para desarrolladores mobile. Skills, agentes y pipelines automatizados para Android, iOS, KMP, Flutter y React Native.

## Que es MobiAI?

MobiAI es un ecosistema open source que potencia tu desarrollo mobile con IA. No es solo un conjunto de skills — es una plataforma completa que crece con la comunidad:

- **Skills** — Conocimiento experto empaquetado en instrucciones reutilizables. Corregir bugs, reproducir incidencias, escribir tests, analizar crashes, crear PRs... todo siguiendo las mejores practicas de cada plataforma.
- **Agentes** — Agentes especializados que ejecutan tareas complejas de forma autonoma: un agente que analiza codigo, otro que interactua con el dispositivo, otro que escribe tests.
- **Pipeline automatizado** — Un flujo completo de correccion de bugs: recibe un error o ticket desde cualquier plataforma de mensajeria, reproduce el bug en un emulador, encuentra la causa raiz, aplica el fix, ejecuta tests y crea el PR. Todo sin intervencion humana.

Empezamos con [Claude Code](https://docs.anthropic.com/en/docs/claude-code) como plataforma principal, pero la vision es que funcione con cualquier herramienta de IA.

## Estado actual

| Componente | Estado |
|------------|--------|
| Skills (Claude Code plugin) | Disponible |
| Agentes autonomos | En desarrollo |
| Pipeline automatizado (bot de mensajeria) | Planificado |
| Marketplace de skills comunitarios | Planificado |

## Instalacion

```bash
# Mediante el sistema de plugins de Claude Code
claude plugin install ArisGuimera/MobiAI-Core
```

## Skills disponibles

### Flujo de trabajo
| Skill | Descripcion |
|-------|-------------|
| [fix-issue](skills/fix-issue/SKILL.md) | Pipeline completo: obtener issue, analizar, corregir, testear, PR |
| [reproduce-bug](skills/reproduce-bug/SKILL.md) | Reproducir bugs en dispositivo/emulador/simulador |
| [analyze-crash](skills/analyze-crash/SKILL.md) | Analizar crash logs, stack traces y Crashlytics |
| [write-tests](skills/write-tests/SKILL.md) | Escribir tests con las convenciones de cada plataforma |
| [review-code](skills/review-code/SKILL.md) | Code review mobile: lifecycle, memoria, thread safety |
| [create-pr](skills/create-pr/SKILL.md) | Crear PRs con contexto mobile y plan de test |

### Android
| Skill | Descripcion |
|-------|-------------|
| [android-device](skills/android-device/SKILL.md) | ADB, emulador, automatizacion de UI, screenshots, logcat |
| [android-build](skills/android-build/SKILL.md) | Gradle, flavors, firmado, ProGuard/R8 |
| [android-testing](skills/android-testing/SKILL.md) | JUnit, MockK, Espresso, Compose Testing |
| [android-architecture](skills/android-architecture/SKILL.md) | Clean Architecture, MVVM, MVI, Compose |

### iOS
| Skill | Descripcion |
|-------|-------------|
| [ios-device](skills/ios-device/SKILL.md) | Simulador via simctl, screenshots, logs |
| [ios-build](skills/ios-build/SKILL.md) | xcodebuild, schemes, CocoaPods, SPM |
| [ios-testing](skills/ios-testing/SKILL.md) | XCTest, Quick/Nimble, snapshot testing |
| [ios-architecture](skills/ios-architecture/SKILL.md) | SwiftUI, UIKit, TCA, MVVM+Combine |

### Multiplataforma
| Skill | Descripcion |
|-------|-------------|
| [kmp](skills/kmp/SKILL.md) | Kotlin Multiplatform |
| [flutter](skills/flutter/SKILL.md) | Flutter/Dart |
| [react-native](skills/react-native/SKILL.md) | React Native |

## Como funciona?

1. **Instala el plugin** — MobiAI registra sus skills en tu asistente de IA
2. **Empieza a programar** — Los skills se activan automaticamente segun el contexto
3. **Deteccion de plataforma** — MobiAI detecta tu proyecto (Android, iOS, Flutter...) y aplica el conocimiento adecuado
4. **Tu proyecto, tus reglas** — Los skills respetan las convenciones de tu `CLAUDE.md`

## Roadmap

### Fase 1 — Skills (actual)
Plugin de Claude Code con skills para todas las plataformas mobile. La comunidad puede contribuir nuevos skills.

### Fase 2 — Agentes
Agentes especializados que van mas alla de las instrucciones: analizan codigo de forma autonoma, interactuan con dispositivos, y toman decisiones inteligentes basandose en el contexto.

### Fase 3 — Pipeline automatizado
Un bot que se conecta a tu plataforma de mensajeria favorita, recibe errores o tickets, orquesta los agentes, y produce PRs listos para review — el ciclo completo de correccion de bugs sin intervencion humana.

### Fase 4 — Marketplace
Un marketplace donde la comunidad publica y descarga skills, agentes y configuraciones para diferentes tipos de proyectos mobile.

## Contribuir

Nos encantaria tu ayuda! MobiAI esta creado por desarrolladores mobile, para desarrolladores mobile.

- **Anadir un nuevo skill**: Consulta [writing-skills](skills/writing-skills/SKILL.md)
- **Mejorar skills existentes**: Abre un PR con mejores instrucciones o manejo de casos edge
- **Reportar problemas**: Abre un issue si un skill da mal consejo o le falta algo

Consulta [CONTRIBUTING.md](CONTRIBUTING.md) para mas detalles.

## Comunidad

Unite a la comunidad de desarrolladores mobile que comparten conocimiento y herramientas. Proximamente mas informacion.

## Licencia

MIT — ver [LICENSE](LICENSE)
