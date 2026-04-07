# Como crear un nuevo skill

Queres contribuir un skill a MobiAI? Genial! Aca te explicamos como.

## Que es un skill?

Un skill es un conjunto de instrucciones que le ensenan a la IA como hacer una tarea especifica de desarrollo mobile. Por ejemplo, el skill `android-device` le ensena a usar ADB, y `fix-issue` le ensena el flujo completo para corregir un bug.

## La forma mas facil: usa el skill `writing-skills`

No necesitas escribir el SKILL.md a mano. Podes pedirle a tu asistente de IA que lo haga por vos:

```
Quiero crear un nuevo skill para MobiAI que [describe lo que queres que haga]
```

El agente va a usar el skill `writing-skills` automaticamente para guiarte en el proceso: te pregunta que tiene que hacer, genera el archivo con el formato correcto, y lo agrega al catalogo.

## Si preferis hacerlo manual

### 1. Crea el directorio

```
skills/mi-skill/
  SKILL.md              # Requerido
  references/           # Opcional: docs detalladas
  scripts/              # Opcional: scripts de ayuda
```

### 2. Escribe el SKILL.md

Los skills se escriben en **ingles** (son instrucciones tecnicas para la IA, no documentacion para el usuario).

El archivo necesita:
- **Frontmatter YAML** con nombre, descripcion, version, licencia, compatibilidad
- **When to Use** — cuando debe activarse
- **Steps** — instrucciones paso a paso
- **Secciones por plataforma** (si aplica)

Mira cualquier skill existente como ejemplo, por ejemplo [android-device](../android-device/SKILL.md).

### 3. Agregalo al catalogo

Anade tu skill a la tabla en `skills/using-mobiai/SKILL.md`.

### 4. Probalo

Usa el skill en un proyecto real antes de enviarlo. Verifica que el agente puede seguir las instrucciones sin ambiguedad.

### 5. Abre un PR

Abre un pull request con:
- Que hace tu skill
- Como lo probaste
- Que plataformas soporta
