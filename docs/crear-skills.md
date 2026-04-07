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

El archivo necesita:
- **Frontmatter YAML** con nombre, descripción, versión, licencia, compatibilidad
- **When to Use** — cuándo debe activarse
- **Steps** — instrucciones paso a paso
- **Secciones por plataforma** (si aplica)

Mirá cualquier skill existente como ejemplo, por ejemplo [android-device](../skills/android-device/SKILL.md).

### 3. Agregalo al catálogo

Añadí tu skill a la tabla en `skills/using-mobiai/SKILL.md`.

### 4. Probalo

Usá el skill en un proyecto real antes de enviarlo. Verificá que el agente puede seguir las instrucciones sin ambigüedad.

### 5. Abrí un PR

Abrí un pull request con:
- Qué hace tu skill
- Cómo lo probaste
- Qué plataformas soporta
