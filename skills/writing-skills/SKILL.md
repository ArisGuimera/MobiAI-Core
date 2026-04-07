---
name: writing-skills
description: How to create new MobiAI skills — the meta skill for community contributions
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
---

# Crear Skills para MobiAI

Este skill te ensena como crear nuevos skills para el ecosistema MobiAI.

## Formato de SKILL.md

Cada skill es un directorio dentro de `skills/` que contiene como minimo un archivo `SKILL.md`:

```
skills/
  mi-nuevo-skill/
    SKILL.md              # Requerido: definicion del skill
    references/           # Opcional: documentacion detallada cargada bajo demanda
      detalle.md
    scripts/              # Opcional: scripts de ayuda que el agente puede ejecutar
      helper.sh
    assets/               # Opcional: templates, ejemplos
      template.kt
```

## Estructura del SKILL.md

```yaml
---
name: mi-nuevo-skill
description: Descripcion de una linea usada para matching de relevancia
version: 1.0.0
license: MIT
author: Tu Nombre
compatibility: [claude-code, cursor, copilot]
platforms: [android, ios, kmp, flutter, react-native]  # Opcional: a que plataformas aplica
---

# Titulo del Skill

## When to Use
Describir las condiciones de activacion — cuando debe invocarse este skill?

## Steps
Instrucciones numeradas para que el agente siga.

## Platform-Specific Sections (si aplica)
### Android
Instrucciones especificas de Android.

### iOS
Instrucciones especificas de iOS.

## References
Apuntar a archivos en references/ para profundizar.

## Common Pitfalls
A que tener cuidado.
```

## Guia para escribir skills

1. **Se especifico y accionable.** Los agentes siguen instrucciones de forma literal. "Busca el crash signal en el codebase" es mejor que "investiga el problema."

2. **Incluir comandos exactos.** Cuando el skill involucra comandos de terminal, poner el comando completo con placeholders:
   ```bash
   adb -s <serial> shell uiautomator dump /sdcard/ui.xml
   ```

3. **Proveer arboles de decision.** Los agentes necesitan reglas claras para logica de ramificacion:
   - Si X → hacer A
   - Si Y → hacer B
   - Si ninguno → hacer C

4. **Mantener el SKILL.md principal conciso.** Mover material de referencia detallado a archivos en `references/` que se cargan bajo demanda.

5. **Probar con escenarios reales.** Antes de enviar, verificar que el skill funciona de punta a punta en un proyecto real.

6. **Agnositco de plataforma cuando sea posible.** Si el skill aplica a multiples plataformas, incluir secciones especificas por plataforma en vez de crear skills separados.

## Campos del frontmatter

| Campo | Requerido | Descripcion |
|-------|-----------|-------------|
| `name` | Si | Identificador unico (kebab-case) |
| `description` | Si | Descripcion de una linea para matching de relevancia |
| `version` | Si | Version semver |
| `license` | Si | Licencia (MIT recomendado) |
| `author` | Si | Nombre del autor o "MobiAI Community" |
| `compatibility` | Si | Que herramientas de IA soportan este skill |
| `platforms` | No | A que plataformas mobile aplica |

## Enviar un nuevo Skill

1. Haz fork del [repo de MobiAI-Core](https://github.com/ArisGuimera/MobiAI-Core)
2. Crea tu directorio de skill dentro de `skills/`
3. Escribe y proba tu `SKILL.md`
4. Anade tu skill a la tabla en `skills/using-mobiai/SKILL.md`
5. Abre un pull request con una descripcion de que hace tu skill y como lo probaste
