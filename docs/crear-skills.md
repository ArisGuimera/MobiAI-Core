# Cómo crear un nuevo skill

¿Quieres contribuir un skill a MobiAI? ¡Genial! Aquí te explicamos cómo.

## ¿Qué es un skill?

Un skill es contexto experto que adapta lo que la IA ya sabe al escenario adecuado. No le enseña a programar — le da las herramientas, los comandos y los flujos específicos de cada plataforma para que aplique su conocimiento de forma precisa. Por ejemplo, el skill `android-device` le da el contexto para usar ADB correctamente, y `fix-issue` le define el flujo completo para corregir un bug.

## La forma más fácil: usa el skill `writing-skills`

No necesitas escribir el SKILL.md a mano. Puedes pedirle a tu asistente de IA que lo haga por ti:

```
Quiero crear un nuevo skill para MobiAI que [describe lo que quieres que haga]
```

El agente va a usar el skill `writing-skills` automáticamente para guiarte en el proceso: te pregunta qué tiene que hacer, genera el archivo con el formato correcto, y lo agrega al catálogo.

## ¿Cómo se cargan los skills?

Es importante entender esto para escribir buenos skills:

1. **Al iniciar la sesión**, la IA solo carga `using-mobiai/SKILL.md` — una tabla liviana con el nombre y la descripción corta de cada skill. No lee ningún SKILL.md completo.
2. **Cuando la IA necesita un skill**, lee la descripción corta de la tabla y decide si es relevante para lo que está haciendo.
3. **Solo si coincide**, la IA lee el SKILL.md completo de ese skill.

Esto significa que:
- **Podemos tener muchos skills** sin problema de tokens — la cantidad no importa porque solo se cargan bajo demanda.
- **La `description` del frontmatter es clave** — es lo único que la IA lee para decidir si necesita ese skill. Si la description es mala, el skill nunca se va a activar.
- **El cuerpo del SKILL.md puede ser largo y detallado** — solo se consume cuando es relevante.

## Si prefieres hacerlo manual

### 1. Crea el directorio

```
skills/mi-skill/
  SKILL.md              # Requerido
  references/           # Opcional: docs detalladas
  scripts/              # Opcional: scripts de ayuda
```

### 2. Escribe el SKILL.md

Los skills se escriben en **inglés** (son instrucciones técnicas que le dan contexto a la IA, no documentación para el usuario).

El archivo tiene dos partes:

**Frontmatter YAML** (obligatorio) — metadata del skill:
- `name` — identificador único (kebab-case)
- `description` — **lo más importante**: una línea en inglés que le dice a la IA *cuándo* debe usar este skill. No es una lista de keywords, es una instrucción de cuándo activarlo. Ejemplo: `"Use when creating features, refactoring, or navigating an Android codebase"`.
- `license`, `compatibility`, `platforms`

**Cuerpo en Markdown** (flexible) — las instrucciones para la IA. No hay una estructura rígida obligatoria. Organízalo como tenga sentido para tu skill: puedes usar secciones como "Workflow", "Steps", secciones por plataforma, tablas de referencia, comandos exactos, árboles de decisión, etc. Lo importante es que sea claro y accionable.

Mira los skills existentes como ejemplo — cada uno tiene la estructura que mejor se adapta a lo que hace:
- [analyze-crash](../skills/analyze-crash/SKILL.md) — flujo paso a paso con decisiones
- [android-device](../skills/android-device/SKILL.md) — referencia de comandos y patrones
- [fix-issue](../skills/fix-issue/SKILL.md) — pipeline completo con reglas de decisión

### 3. Agrégalo al catálogo

Añade tu skill a la tabla en `skills/using-mobiai/SKILL.md`.

### 4. Pruébalo

Usa el skill en un proyecto real antes de enviarlo. Verifica que el agente puede seguir las instrucciones sin ambigüedad.

### 5. Abre un PR

Abre un pull request con:
- Qué hace tu skill
- Cómo lo probaste
- Qué plataformas soporta
