# Contribuir a MobiAI-Core

Gracias por tu interes en contribuir! MobiAI es un proyecto comunitario — cada skill, agente, correccion y mejora hace que el desarrollo mobile sea mejor para todos.

## Formas de contribuir

### 1. Anadir un nuevo Skill
Consulta la [guia para crear skills](docs/crear-skills.md).

**Pasos rapidos:**
1. Haz fork del repo
2. Crea `skills/nombre-de-tu-skill/SKILL.md`
3. Anade tu skill a la tabla en `skills/using-mobiai/SKILL.md`
4. Pruebalo en un proyecto real
5. Abre un PR

### 2. Mejorar Skills o Agentes existentes
Encontraste algo que podria ser mejor? Mejoras comunes:
- Mejores comandos o instrucciones
- Casos edge que faltan
- Secciones especificas de plataforma
- Arboles de decision mas claros

### 3. Reportar problemas
Si un skill o agente da consejo incorrecto o le falta algo, abre un issue con:
- Que skill o agente
- Que salio mal
- Cual deberia ser el comportamiento correcto
- (Bonus) El contexto del proyecto donde ocurrio

## Estandares de calidad

Antes de enviar una contribucion:

- [ ] **Probado en un proyecto real** — No solo teoria, realmente lo usaste
- [ ] **Instrucciones accionables** — El agente puede seguirlas sin adivinar
- [ ] **Comandos exactos** — Los comandos shell incluyen todos los flags y placeholders
- [ ] **Arboles de decision** — Logica de ramificacion clara para diferentes escenarios
- [ ] **Secciones por plataforma** — Si es multi-plataforma, cada plataforma tiene su seccion
- [ ] **Conciso** — Mueve explicaciones detalladas a archivos en `references/`
- [ ] **Sin secretos** — Sin API keys, tokens ni credenciales

## Proceso de Pull Request

1. Haz fork y crea una rama: `feat/mi-nuevo-skill` o `fix/android-device-typo`
2. Haz tus cambios
3. Prueba de principio a fin
4. Abre un PR con:
   - Que hace tu contribucion
   - Como lo probaste
   - Que plataformas soporta

## Codigo de conducta

Se respetuoso, se util, se constructivo. Todos estamos aqui para aprender y construir mejores herramientas juntos.
