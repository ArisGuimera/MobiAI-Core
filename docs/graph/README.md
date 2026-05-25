# MobiAI Graph

> Exploración semántica del código mobile, local y regenerable. MobiAI Graph entiende apps Android, iOS, KMP, Flutter y React Native.

## Qué es MobiAI Graph

MobiAI Graph es una **capacidad** del agente, no un producto separado: exploración semántica del código mobile sobre un índice local que se regenera bajo demanda. No es un servicio, no es una base de datos remota, no requiere cuenta ni cloud.

Vive en `<repo>/.mobiai/graph/index.json` — local-first, regenerable, sin nube. En V1 cubre **Kotlin** y **Swift**, los dos lenguajes nativos del stack mobile que MobiAI prioriza.

## Skills vs Brain vs Graph

MobiAI tiene tres piezas complementarias. Cada una resuelve un problema distinto:

| Pieza | Qué guarda | Quién la mantiene | Propiedad clave |
|---|---|---|---|
| Skills | Cómo trabajar (proceso, flujos) | El equipo MobiAI (en plugins) | Globales, compartidas entre proyectos |
| Brain | Qué se decidió, qué workarounds hay activos | El usuario del proyecto | Memoria histórica humana, diff-friendly |
| Graph | Estructura actual del código | Generado automáticamente | Computado, regenerable, no editar a mano |

Las tres son complementarias. Una decisión bien tomada combina las tres: la skill marca el flujo, Brain aporta el contexto histórico, Graph señala dónde vive el código afectado hoy.

## Cómo funciona V1

Pipeline completo en una pasada (`mobiai graph init`):

```
filepath.WalkDir(root)
   │
   ├─ skip dirs: .git .gradle .idea .vscode .mobiai
   │             .claude .claude-plugin .worktrees
   │             build dist out target Pods DerivedData node_modules
   │
   ├─ por cada .kt / .swift:
   │     1. read file
   │     2. strip comments + strings (preservando números de línea)
   │     3. regex line-by-line → símbolos (class/fun/struct/...)
   │     4. brace-depth stack → asigna container al símbolo
   │
   └─ sort por path → atomic write JSON (.tmp + fsync + rename)
```

**Decisiones técnicas firmes**:

- **Regex line-by-line, no AST**. Para 2 lenguajes y este nivel de detalle (símbolos top-level y anidados, no resolución de tipos), regex es suficiente y elimina dependencias pesadas como tree-sitter. Revisaremos cuando entre el 3er lenguaje.
- **JSON, no SQLite**. El índice cabe holgadamente en JSON (TaykusKMP: 647 archivos, 2464 símbolos). JSON es diff-friendly, inspeccionable, y mantiene cero dependencias externas en la CLI Go.
- **Sin watcher**. `mobiai graph init` regenera el índice entero en milisegundos. Un watcher con `fsnotify` queda para V2 cuando se justifique.
- **Strip comments+strings preservando líneas**. Los regex parserían falsos positivos en docstrings y literales. El strip reemplaza con espacios (no borra) — así los números de línea reportados por el indexer apuntan al lugar real en el archivo.
- **Brace-depth stack para containers**. El símbolo más anidado gana: una `fun login()` dentro de `class LoginViewModel` registra `container = "LoginViewModel"`.
- **Atomic write**. Escribir `.tmp` + `fsync` + `rename` garantiza que un Ctrl-C a mitad de `init` nunca corrompe `index.json`.
- **Cero dependencias externas**. Stdlib Go + `cobra` (ya presente en la CLI). No CGo, no tree-sitter, no SQLite.

Ejemplo de uso real:

```
$ mobiai graph init
✓ Índice generado: /Users/.../proyecto/.mobiai/graph/index.json
  Archivos: 1
  Símbolos: 1
  Kotlin: 1
  Swift: 0

$ mobiai graph search Login
app/LoginViewModel.kt:5 class LoginViewModel
```

El índice es **regenerable**: si los regex evolucionan o el código cambia, basta con volver a correr `mobiai graph init`.

## Benchmark real: Graph vs grep

Medido sobre **TaykusKMP** (Kotlin Multiplatform, Android + iOS).

**Contexto del índice**:

- Archivos indexados: **647**
- Símbolos: **2464** (635 Kotlin + 12 Swift)
- Tiempo de `mobiai graph init`: **0.47 s**
- Commit base: `dcbe040` · Fecha: 2026-05-25

Tokens estimados como `bytes / 4` (proxy estándar GPT/Claude).

| Pregunta | Herramienta | Tiempo | Bytes | Tokens ≈ | Ganador |
|---|---|---|---|---|---|
| **Q1** — Lista todos los ViewModel del proyecto | `mobiai graph search ViewModel` | 0.01 s | 9446 | 2361 | — |
|  | `rg "class \w+ViewModel"` | 0.15 s | 7426 | 1856 | **grep** (output más compacto) |
| **Q2** — ¿Dónde se usa `TaykusScaffold`? | `mobiai graph callers TaykusScaffold` | 0.07 s | 6766 | 1691 | — |
|  | `rg "TaykusScaffold"` | 0.01 s | 6888 | 1722 | **empate técnico** |
| **Q3** — Contexto para fix de login + refresh token | `mobiai graph context "fix login bug refresh token"` | 0.01 s | 1021 | 255 | **graph** (6.3× menos tokens) |
|  | `rg -l -i "login\|refresh.?token"` | 0.03 s | 6403 | 1600 | — |
| **Q4** — ¿Existe `AuthRepository` y dónde se declara? | `mobiai graph search AuthRepository` | 0.01 s | 237 | 59 | **graph** (7.5× menos tokens) |
|  | `rg "AuthRepository"` | 0.03 s | 1762 | 440 | — |

### Totales agregados

```
graph │██████████████████████░░░░░░  4366 tokens · 0.10 s
grep  │████████████████████████████  5618 tokens · 0.22 s
                                     ─22 % de tokens
```

- Tokens graph: **4366** · grep: **5618** → **ahorro global ≈ 22 %**
- Tiempo total graph: **0.10 s** · grep: **0.22 s**

### Lectura honesta

- Graph **no** sustituye a grep en consultas de texto libre o cuando ya conoces el nombre exacto (Q1, Q2): grep es igual o mejor.
- Graph **brilla** cuando la consulta es semántica o la herramienta tradicional escupiría mucho ruido (Q3, Q4): el ranking y el filtrado estructural recortan el output a una fracción.
- El ahorro de tokens es **acumulativo**: cuanto más se apoye el agente en `graph context`/`graph search` para arrancar la tarea, más tokens libera para razonamiento real.

### Regla mental

| Caso | Herramienta |
|---|---|
| Buscas un **símbolo** (clase, función, repo, VM…) | `mobiai graph` |
| Buscas **texto literal** (strings, comentarios, valores) | `rg` / `grep` |
| Ya **conoces el archivo** | `Read` directo |

Graph complementa a grep, no lo reemplaza.

## Cómo lo usa un agente IA

Graph existe para que los agentes IA arranquen una tarea **sabiendo dónde mirar**, en lugar de adivinar con grep.

### Patrón Brain → Graph → action

El flujo canónico cuando un agente recibe una tarea sobre un proyecto mobile:

1. **Brain** aporta el contexto histórico: "este crash ya pasó, se fijó con un workaround X, no rollback".
2. **Graph** aporta la estructura actual: "los archivos relevantes hoy son estos 10 rankeados por score".
3. **Acción**: el agente lee solo los archivos top, no los 74 que grep arrojaría.

Esto reduce tokens leídos y elimina lectura especulativa de archivos irrelevantes.

### Pre-flight pattern

Las skills downstream (`mobiai-mobile-debugging`, `mobiai-fix-issue`, `mobiai-write-tests`, `mobiai-review-code`) ejecutan al inicio:

```bash
# 1. ¿Existe el índice?
test -f .mobiai/graph/index.json || mobiai graph init

# 2. ¿Está fresco? (status muestra "hace Xm/Xh/Xd")
mobiai graph status

# 3. Query relevante al issue:
mobiai graph context "<descripción del bug/feature>"
mobiai graph callers <SímboloAfectado>
```

El pre-flight se hace **una vez al inicio de la conversación**, no por cada subcomando. Los resultados quedan en el contexto del agente para reusar.

### Caso típico: "arregla el bug de login + refresh token"

Sin Graph: el agente hace `rg -l -i "login|refresh.?token"` y recibe **6403 bytes** (~1600 tokens) de paths planos sin pista de cuál es central — incluye tests, fakes, repositorios, viewmodels y pantallas mezclados.

Con Graph:

```
$ mobiai graph context "fix login bug refresh token"
LoginScreenTest.kt          (score 10)
UserPreferencesManager.kt   (score 4)
UserRepositoryImpl.kt       (score 4)
UserApiClient.kt            (score 4)
UserRepository.kt           (score 4)
LoginScreen.kt              (score 4)
LoginViewModel.kt           (score 4)
…
```

**1021 bytes (~255 tokens)** rankeados por score: el agente lee la pantalla, el ViewModel y el repositorio que tocan el token y propone el fix. **6.3× menos tokens, todos relevantes.**

## Lenguajes soportados

| Lenguaje | V1 | Roadmap V2/V3 |
|---|---|---|
| Kotlin | ✅ | mejorar precisión con AST |
| Swift | ✅ | mejorar precisión con AST |
| Dart (Flutter) | ❌ | V2 |
| JavaScript / TypeScript (RN) | ❌ | V2 |
| Java | ❌ | V2 |
| KMP `expect/actual` semantics | parcial (cada lado se indexa por separado) | V3 |

## Inspiración

La idea de un grafo semántico local para acelerar a los agentes IA no es nueva. [CodeGraph](https://github.com/colbymchenry/codegraph) (TypeScript + tree-sitter + SQLite/FTS5) la popularizó en proyectos web. MobiAI Graph toma esa idea y la lleva al terreno mobile, con stack Go nativo y prioridad en lenguajes y patrones que importan en Android, iOS, KMP, Flutter y React Native.

## Roadmap (futuro, no implementado)

- **V1.2** (próximo paso real):
  - Pre-flight de Graph integrado en skills downstream (`mobiai-mobile-debugging`, `mobiai-fix-issue`, `mobiai-write-tests`, `mobiai-review-code`) — sin esto, los agentes IA no invocan Graph autónomamente.
- **V2** (futuro):
  - Soporte Dart, JavaScript/TypeScript y Java.
  - Watcher incremental — el índice se refresca cuando cambian archivos.
  - Búsqueda exacta (`--exact`) y por regex (`--regex`).
  - Posible migración a tree-sitter si los regex se quedan cortos (decisión pendiente).
- **V3** (futuro):
  - Semántica mobile específica: Compose `@Composable` → ViewModel → UseCase → Repository.
  - KMP `expect/actual` mapeado a sus implementaciones por plataforma.
  - Navigation Compose / SwiftUI routes → screens.
  - Hilt / Koin module bindings.
- **V4** (especulativo, sin fecha):
  - MCP server propio para que clientes IA consulten Graph como tool nativa.

Ninguna de estas piezas está disponible hoy. La documentación se actualizará cuando lo estén.

## Política

- MobiAI nunca instala backends externos automáticamente. Graph es Go nativo y vive dentro de la CLI `mobiai`.
- El índice (`.mobiai/graph/index.json`) es **regenerable**: no es obligatorio commitearlo. Si tu equipo prefiere mantenerlo fuera del repo, gitignorealo sin miedo — no es información crítica.
- Graph **no reemplaza a Brain**: si una decisión, workaround o convención debe persistir entre cambios de código, va a Brain. Graph solo refleja la estructura actual del código.
- Cero datos del usuario salen de su máquina. Graph es 100% local.
