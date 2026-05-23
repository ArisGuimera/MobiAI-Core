---
name: mobiai-graph
description: "Use when the user asks about code impact, call graph, callers/callees, semantic exploration of the mobile codebase, or 'where is X used'. Routes the agent to query MobiAI Graph (mobiai graph search/callers/context) instead of grep/read. Pre-flight checks that .mobiai/graph/index.json exists; if not, suggests `mobiai graph init` without running it."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# MobiAI Graph

CodeGraph entiende código. MobiAI Graph entiende apps mobile. Índice semántico per-project del codebase mobile, vive en `<repo>/.mobiai/graph/index.json` — regenerable, complementario a Brain (Brain = memoria humana del proyecto, Graph = estructura computada del código).

## When to invoke this skill

- User asks "where is X used", "impacto de cambiar X", "dónde se usa", "qué llama a Y", "callers/callees".
- User pide el call graph o quiere explorar la arquitectura del proyecto.
- User pregunta qué tests están afectados por un cambio.
- Antes de cambios cross-file no triviales: para acotar el blast radius antes de tocar código.
- Cuando un fix exige saber qué pantallas / ViewModels / repositorios dependen de un símbolo concreto.

**Skip** cuando el archivo / símbolo ya está identificado, cuando el fix es one-liner, o cuando el output esperado es más grande que lo que devuelve una sola llamada al CLI — en ese caso delegar a un sub-agente Explore.

## Pre-flight

**Verificá una única vez** al inicio de la conversación (o al invocar esta skill por primera vez), no antes de cada subcomando:

1. Verificá que existe `<repo>/.mobiai/graph/index.json`.
2. Si **no existe**: avisá una sola vez al usuario:
   > "No encuentro `<repo>/.mobiai/graph/index.json`. Podés crearlo con: `mobiai graph init`"
   No corras `init` por tu cuenta. Esperá confirmación explícita.
3. Si **existe pero tiene más de 24h**: mencionalo casualmente y ofrecé re-correr `mobiai graph init` antes de confiar en los resultados. Sólo re-indexá si el usuario lo pide.

`mobiai graph status` muestra la antigüedad del índice y el conteo de símbolos — sirve como check rápido.

## Commands available today

```
mobiai graph init                  # Indexa el proyecto (Kotlin + Swift en V1)
mobiai graph status                # Stats del índice y antigüedad
mobiai graph search <term>         # Lookup de símbolos
mobiai graph callers <symbol>      # Referencias textuales
mobiai graph context <task...>     # Archivos relevantes para una tarea
```

```bash
# Ambas formas son válidas (Cobra acepta 1+ argumentos posicionales):
mobiai graph context fix login bug
mobiai graph context "fix login bug"
```

Cada uno devuelve output autocontenido. Preferí una llamada por pregunta concreta; no encadenes 5 búsquedas si una sola `context` resuelve el caso.

### Recipes rápidas

```bash
# Antes de tocar un símbolo cross-file:
mobiai graph callers AuthRepository

# Para arrancar una feature nueva y entender qué archivos están en juego:
mobiai graph context "agregar refresh token al flujo de login"

# Para chequear si existe un nombre antes de inventarlo:
mobiai graph search SessionManager

# Para confirmar que el índice no está stale antes de confiar en callers:
mobiai graph status
```

Si vas a hacer varias preguntas seguidas sobre el mismo área, **una sola `context` suele ser más barata** que tres `search` + dos `callers`.

## When to prefer Graph over grep/read

| Pregunta | Herramienta preferida |
|---|---|
| "¿Dónde se usa X?" | `mobiai graph callers X` |
| "¿Qué pantallas tocan este Repository?" | `mobiai graph context "<descripción>"` |
| "¿Qué clases tienen 'Login' en el nombre?" | `mobiai graph search Login` |
| "Leer el contenido exacto de Foo.kt" | `Read` (no necesitás el grafo) |
| "Buscar la string literal 'TODO: fix'" | `grep` / `rg` (no es semántico) |

Regla mental: si la pregunta es **sobre un símbolo** (clase, función, propiedad), Graph. Si es **sobre texto exacto**, grep. Si es **leer un archivo conocido**, Read.

### Anatomía del output

- `search` devuelve nombres de símbolos con el archivo donde están declarados. Útil para confirmar que un nombre existe antes de citarlo.
- `callers` devuelve la lista de archivos y líneas que **mencionan** el símbolo. Es match textual: filtrá mentalmente los falsos positivos obvios (comentarios, strings).
- `context` devuelve un set de archivos rankeados por overlap con la descripción de la tarea. Es el punto de entrada típico cuando arrancás un ticket sin saber dónde mirar.
- `status` devuelve métricas: cantidad de símbolos, archivos indexados, timestamp del último `init`.

## Combinación con Brain

Orden recomendado cuando ambas skills aplican:

1. **Brain primero** — para entender decisiones, convenciones y workarounds activos del proyecto (`mobiai brain context`).
2. **Graph después** — para descubrir la estructura actual del código que esas decisiones produjeron.
3. **Acción al final**, respetando ambos.

Ejemplo concreto:

> El usuario pide migrar `LoginViewModel` a otra librería de DI.
> Brain dice: "decisión activa — usamos Koin, no Hilt, por compatibilidad con KMP".
> Graph dice: `mobiai graph callers LoginViewModel` → 3 archivos (LoginScreen, LoginNavGraph, LoginViewModelTest); `mobiai graph context "login flow"` → suma LoginUseCase y AuthRepository.
> Acción: respetar Koin (Brain), tocar exactamente esos archivos (Graph), no inventar refactor más amplio.

## When NOT to use Graph

- Archivo ya identificado y cambio de una sola línea.
- Búsqueda de string literal, no de símbolo (usá `grep` / `rg`).
- El índice nunca se generó y el usuario no quiere correr `init`. V1 cubre Kotlin + Swift; en proyectos Flutter/Dart, React Native o web puro el grafo está vacío y no aporta nada.
- El output esperado es enorme (miles de hits) y va a inflar el contexto — delegar a un subagente Explore con instrucción de devolver sólo el resumen.
- Tarea puramente de UI / assets / configuración donde el grafo de símbolos no aplica.

## Cómo se enlaza con otras skills MobiAI

Graph es un proveedor de contexto estructural; varias skills de metodología pueden consultarlo sin reemplazarse:

- `mobiai-mobile-debugging` puede usar `callers` / `context` para entender qué se ve afectado por una hipótesis de root cause antes de proponer fix.
- `mobiai-fix-issue` puede usar `context` para acotar el blast radius del ticket antes de tocar archivos.
- `mobiai-write-tests` puede usar `callers` para encontrar tests existentes relacionados al símbolo bajo cambio y evitar duplicar cobertura.
- `mobiai-review-code` puede usar `callers` para evaluar el impacto real de un diff cross-file.

Ninguna de esas skills se modifica desde acá: Graph es opt-in, cada skill decide cuándo llamarlo.

## Limitaciones de V1

Sé honesto sobre lo que Graph **no** hace todavía:

- **Sólo Kotlin y Swift.** Dart, JS/TS, Java puro no se indexan en V1.
- **Scanners regex, no AST** → posibles falsos positivos en símbolos dentro de strings raros, comentarios multilínea exóticos o macros.
- **No hay resolución de tipos** → `callers` es match textual de nombre, no análisis de overload ni de shadowing.
- **`context` es heurística por overlap de tokens**, no búsqueda semántica con embeddings. Funciona bien para tareas con vocabulario que ya aparece en el código; degrada si la descripción es muy abstracta.
- **El índice no es incremental**: `init` lo regenera entero cada vez. En repos grandes puede tardar.

Cuando una de estas limitaciones aplica al caso, decilo al usuario y caé al método más adecuado (Read directo, grep, o sub-agente Explore).

## Roadmap

Lo siguiente **no está implementado**, es sólo dirección futura — no prometas estas features como disponibles:

- **V2 (planeado, no implementado):** AST real (posiblemente tree-sitter), soporte para más lenguajes (Dart, JS/TS, Java) y watcher incremental para no regenerar el índice entero.
- **V3 (planeado, no implementado):** semántica mobile-específica — cadenas `Composable → ViewModel → UseCase → Repository`, KMP `expect/actual`, rutas de Compose Navigation, etc.
- **V4 (planeado, no implementado):** servidor MCP propio para que clientes IA consulten Graph como herramienta nativa (paralelo a `mobiai brain mcp`).

Si el usuario pregunta por algo de este roadmap, dejá claro que hoy no está disponible y, si aplica, sugerí abrir un issue.

## Behavior contract

- Graph es **per-project**. No mezcles resultados entre repos ni asumas que un símbolo del proyecto A existe en el proyecto B.
- Graph es **complementario** a Brain, a `CLAUDE.md` y a las skills de metodología — nunca las reemplaza.
- El índice es **regenerable y desechable**. No lo edites a mano: corré `mobiai graph init`.
- Si un resultado de `callers` o `context` se ve sospechoso (cero hits para algo que claramente existe, o un hit obviamente erróneo), chequeá `mobiai graph status` y considerá re-indexar antes de actuar.
