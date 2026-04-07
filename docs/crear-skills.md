# Cómo crear un nuevo skill

¿Querés contribuir un skill a MobiAI? ¡Genial! Acá te explicamos cómo.

## ¿Qué es un skill?

Un skill es contexto experto que adapta lo que la IA ya sabe al escenario adecuado. No le enseña a programar — le da las herramientas, los comandos y los flujos específicos de cada plataforma para que aplique su conocimiento de forma precisa. Por ejemplo, el skill `android-device` le da el contexto para usar ADB correctamente, y `fix-issue` le define el flujo completo para corregir un bug.

## La forma más fácil: usá el skill `writing-skills`

No necesitás escribir el SKILL.md a mano. Podés pedirle a tu asistente de IA que lo haga por vos:

```
Quiero crear un nuevo skill para MobiAI que [describí lo que querés que haga]
```

El agente va a usar el skill `writing-skills` automáticamente para guiarte en el proceso: te pregunta qué tiene que hacer, genera el archivo con el formato correcto, y lo agrega al catálogo.

## Si preferís hacerlo manual

### 1. Creá el directorio

```
skills/mi-skill/
  SKILL.md              # Requerido
  references/           # Opcional: docs detalladas
  scripts/              # Opcional: scripts de ayuda
```

### 2. Escribí el SKILL.md

Los skills se escriben en **inglés** (son instrucciones técnicas que le dan contexto a la IA, no documentación para el usuario).

El archivo tiene dos partes:

**Frontmatter YAML** (obligatorio) — metadata del skill:
- `name` — identificador único (kebab-case)
- `description` — **lo más importante**: una línea en inglés que le dice a la IA *cuándo* debe usar este skill. No es una lista de keywords, es una instrucción de cuándo activarlo. Ejemplo: `"Use when creating features, refactoring, or navigating an Android codebase"`.
- `version`, `license`, `author`, `compatibility`, `platforms`

**Cuerpo en Markdown** (flexible) — las instrucciones para la IA. No hay una estructura rígida obligatoria. Organizalo como tenga sentido para tu skill: podés usar secciones como "Workflow", "Steps", secciones por plataforma, tablas de referencia, comandos exactos, árboles de decisión, etc. Lo importante es que sea claro y accionable.

Mirá los skills existentes como ejemplo — cada uno tiene la estructura que mejor se adapta a lo que hace:
- [analyze-crash](../skills/analyze-crash/SKILL.md) — flujo paso a paso con decisiones
- [android-device](../skills/android-device/SKILL.md) — referencia de comandos y patrones
- [fix-issue](../skills/fix-issue/SKILL.md) — pipeline completo con reglas de decisión

### 3. Agregalo al catálogo

Añadí tu skill a la tabla en `skills/using-mobiai/SKILL.md`.

### 4. Probalo

Usá el skill en un proyecto real antes de enviarlo. Verificá que el agente puede seguir las instrucciones sin ambigüedad.

### 5. Abrí un PR

Abrí un pull request con:
- Qué hace tu skill
- Cómo lo probaste
- Qué plataformas soporta
