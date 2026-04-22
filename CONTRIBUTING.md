# Contribuir a MobiAI-Core

¡Gracias por tu interés en contribuir! MobiAI es un proyecto comunitario — cada skill, agente, corrección y mejora hace que el desarrollo mobile sea mejor para todos.

## Formas de contribuir

### 1. Añadir un nuevo Skill
Consulta la [guía para crear skills](docs/crear-skills.md).

**Pasos rápidos:**
1. Haz fork del repo
2. Decidí a qué plugin va tu skill:
   - Workflow cross-platform → `plugins/core/skills/`
   - Android-specific → `plugins/android/skills/`
   - iOS-specific → `plugins/ios/skills/`
   - KMP/Flutter/RN → `plugins/kmp/skills/` | `plugins/flutter/skills/` | `plugins/react-native/skills/`
3. Crea `plugins/<plugin>/skills/nombre-de-tu-skill/SKILL.md`
4. Añade tu skill a la tabla en `plugins/core/skills/using-mobiai/SKILL.md`
5. Pruébalo en un proyecto real
6. Abre un PR

**Regla crítica:** los nombres de skill deben ser únicos en TODO el catálogo. Si tu skill es platform-specific, prefijá con la plataforma (`mobiai-android-build`, `mobiai-ios-build`).

### 2. Mejorar Skills o Agentes existentes
¿Encontraste algo que podría ser mejor? Mejoras comunes:
- Mejores comandos o instrucciones
- Casos edge que faltan
- Secciones específicas de plataforma
- Árboles de decisión más claros

### 3. Reportar problemas
Si un skill o agente da consejo incorrecto o le falta algo, abre un issue con:
- Qué skill o agente
- Qué salió mal
- Cuál debería ser el comportamiento correcto
- (Bonus) El contexto del proyecto donde ocurrió

## Estándares de calidad

Antes de enviar una contribución:

- [ ] **Probado en un proyecto real** — No solo teoría, realmente lo usaste
- [ ] **Instrucciones accionables** — El agente puede seguirlas sin adivinar
- [ ] **Comandos exactos** — Los comandos shell incluyen todos los flags y placeholders
- [ ] **Árboles de decisión** — Lógica de ramificación clara para diferentes escenarios
- [ ] **Secciones por plataforma** — Si es multi-plataforma, cada plataforma tiene su sección
- [ ] **Conciso** — Mueve explicaciones detalladas a archivos en `references/`
- [ ] **Sin secretos** — Sin API keys, tokens ni credenciales

## Proceso de Pull Request

1. Haz fork y crea una rama: `feat/mi-nuevo-skill` o `fix/android-device-typo`
2. Haz tus cambios
3. Prueba de principio a fin
4. Abre un PR con:
   - Qué hace tu contribución
   - Cómo lo probaste
   - Qué plataformas soporta

## Código de conducta

Sé respetuoso, sé útil, sé constructivo. Todos estamos aquí para aprender y construir mejores herramientas juntos.
