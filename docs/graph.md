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

- Scanners por **regex línea-a-línea** para Kotlin y Swift.
- Persistencia **JSON plano** en `.mobiai/graph/index.json`.
- Sin dependencias externas (stdlib Go + `cobra` ya presente en la CLI).
- Sin watcher: `mobiai graph init` regenera el índice entero.
- Sin AST, sin resolución de tipos, sin análisis cross-file.

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

- **V2** (futuro):
  - Soporte Dart, JavaScript/TypeScript y Java (no implementado).
  - Watcher incremental — el índice se refresca cuando cambian archivos (no implementado).
  - Posible migración a tree-sitter como parser si los regex se quedan cortos (decisión pendiente).
- **V3** (futuro):
  - Semántica mobile específica: Compose `@Composable` → ViewModel → UseCase → Repository (no implementado).
  - KMP `expect/actual` mapeado a sus implementaciones por plataforma (no implementado).
  - Navigation Compose / SwiftUI routes → screens (no implementado).
  - Hilt / Koin module bindings (no implementado).
- **V4** (especulativo, sin fecha):
  - MCP server propio para que clientes IA consulten Graph como tool nativa (no implementado).

Ninguna de estas piezas está disponible hoy. La documentación se actualizará cuando lo estén.

## Política

- MobiAI nunca instala backends externos automáticamente. Graph es Go nativo y vive dentro de la CLI `mobiai`.
- El índice (`.mobiai/graph/index.json`) es **regenerable**: no es obligatorio commitearlo. Si tu equipo prefiere mantenerlo fuera del repo, gitignorealo sin miedo — no es información crítica.
- Graph **no reemplaza a Brain**: si una decisión, workaround o convención debe persistir entre cambios de código, va a Brain. Graph solo refleja la estructura actual del código.
- Cero datos del usuario salen de su máquina. Graph es 100% local.
