# MobiAI Brain

> **Leer en otro idioma:** [English](README.md) · **Español**

> La **memoria viva por proyecto** del ecosistema MobiAI. Mientras las skills enseñan a la IA *cómo* trabajar, el Brain guarda el *conocimiento específico* de cada proyecto mobile: decisiones, bugfixes, workarounds, patrones de testing y conocimiento sobre cómo se comportan la app y sus servicios.

**Tabla de contenidos**

- [Quickstart (60 segundos)](#quickstart-60-segundos)
- [¿Qué resuelve?](#qué-resuelve)
- [Cómo encaja con Skills y CLAUDE.md](#cómo-encaja-con-skills-y-claudemd)
- [Filosofía](#filosofía)
- [Comandos](#comandos)
- [Ejemplos end-to-end](#ejemplos-end-to-end)
- [Arquitectura interna](#arquitectura-interna)
- [Integración con Skills](#integración-con-skills)
- [FAQ y troubleshooting](#faq-y-troubleshooting)
- [Roadmap](#roadmap)

---

## Quickstart (60 segundos)

Asumiendo que ya tienes el binario `mobiai` instalado (ver el [README principal](../README.md) si no):

```bash
# 1. En la raíz de tu proyecto mobile
cd ~/code/mi-app

# 2. Inicializa el Brain (idempotente — no rompe nada si ya existe)
mobiai brain init

# 3. Deja que detecte tu stack
mobiai brain scan

# 4. Mira qué encontró
mobiai brain context
```

A partir de aquí, las skills de MobiAI te van a **proponer** guardar decisiones, bugfixes y patrones de testing al final de sus flujos. También puedes guardar manualmente:

```bash
mobiai brain save decision --title "Use Koin como DI" --platform shared \
  --body "KMP-friendly, sin code generation, integrado en composeApp/iosApp."
```

Y consultar cuando lo necesites:

```bash
mobiai brain search firebase
mobiai brain context --section bugfixes --platform ios --status temporary
```

**Recomendado:** registra el Brain como servidor MCP en tu cliente. Para Claude Code y Cursor es un comando:

```bash
mobiai brain install-mcp
```

Detecta los clientes presentes y registra `mobiai-brain` preservando el resto del config. Para otros clientes (Copilot CLI, Codex, Gemini CLI) seguí el registro manual en [`MCP-SETUP.md`](MCP-SETUP.md).

---

## ¿Qué resuelve?

Todos los proyectos mobile acumulan *conocimiento histórico cambiante* que no entra cómodo en ningún lado:

- **"Usamos Koin, no Hilt"** — decisión de marzo 2026, sigue activa. No es metodología (no va en una skill), pero es demasiado específico para CLAUDE.md.
- **"FirebaseAuth iOS necesita `composeApp` en minúsculas hasta el próximo major de CocoaPods"** — workaround temporal. Si no lo trackeas, en 6 meses nadie recuerda por qué está así.
- **"El test de DataStore tiene que esperar `dataStore.data.first { it.asMap().isEmpty() }` después de `clear()`"** — patrón no obvio que descubriste a la mala. Lo vas a necesitar otra vez.

Sin Brain, ese conocimiento vive en commits enterrados, en el Slack del equipo, o en la cabeza de quien lo descubrió. Con Brain vive **dentro del repo**, estructurado, buscable, y accesible para cualquier agente IA conectado al proyecto.

---

## Cómo encaja con Skills y CLAUDE.md

| | Skills | CLAUDE.md / AGENTS.md / GEMINI.md | MobiAI Brain |
|---|---|---|---|
| **Qué guarda** | Metodología, buenas prácticas | Reglas estables del proyecto | Conocimiento histórico cambiante |
| **Alcance** | Global (compartido entre proyectos) | Por proyecto | Por proyecto |
| **Estructura** | SKILL.md | Markdown libre | Markdown estructurado + JSON |
| **¿Caduca?** | No (se actualiza vía PR) | Pocas veces | A menudo (workarounds temporales, decisiones revisables) |
| **¿Cómo se consulta?** | Cargada al iniciar el agente | Cargada al iniciar el agente | On-demand: `context`, `search`, MCP tools |

Los tres coexisten. CLAUDE.md te dice **cómo** trabajar; Brain te dice **qué se decidió antes**.

---

## Filosofía

- **Por proyecto**, no global. Cada repo mobile tiene su propio cerebro local en `<repo>/.mobiai/brain/`.
- **Local primero**. Markdown plano, JSON simple. Sin SQLite, embeddings ni nube por ahora.
- **Mobile-first**. El valor diferencial es entender Android / iOS / KMP / Flutter / React Native, no guardar texto genérico.
- **Sin secretos**. El scanner detecta archivos sensibles (`.env`, `GoogleService-Info.plist`, `google-services.json`) y solo registra su existencia, nunca su contenido.

---

## Comandos

```bash
mobiai brain init                  # Crea .mobiai/brain/ en el proyecto actual (idempotente)
mobiai brain scan                  # Detecta stack: Android, iOS, KMP, Flutter, RN, librerías, CI/CD
mobiai brain context               # Imprime Markdown con config + scan + memorias (acepta filtros)

mobiai brain save decision ...     # Guarda una decisión de arquitectura
mobiai brain save bugfix ...       # Guarda un bugfix o workaround
mobiai brain save testing ...      # Guarda un patrón de testing reusable
mobiai brain save knowledge ...    # Guarda cómo se comporta la app o un servicio externo

mobiai brain search <query>        # Busca texto libre en las memorias (o --crash <firma>)
mobiai brain show <id>             # Imprime una nota completa
mobiai brain review                # Temporales vencidas + notas que pueden estar desactualizadas
mobiai brain promote <id> ...      # Cambia el status de una entrada existente
mobiai brain bump <id> ...         # Extiende review_after de una entrada existente
mobiai brain verify <id>           # Marca una nota como revisada contra el código actual

mobiai brain mcp                   # Arranca un server MCP que expone el Brain como tools
mobiai brain install-mcp           # Registra el server MCP en clientes IA (Claude Code, Cursor)
```

Todos aceptan `--root <ruta>` para apuntar a un proyecto distinto del directorio actual.

### `mobiai brain init`

Crea la estructura local:

```
<repo>/.mobiai/brain/
├── config.json
└── memories/
    ├── decisions/       ← un archivo por nota
    ├── bugfixes/
    ├── testing/
    ├── knowledge/
    ├── decisions.md     ← formato viejo: solo el aviso para mobiai < 0.3
    ├── bugfixes.md
    └── testing.md
```

Es idempotente: si `config.json` o cualquier archivo de memoria ya existe, no se sobrescribe. Los tres `.md` existen solo para que un compañero con un mobiai anterior vea un aviso que le pide actualizar (ver [Actualizar desde 0.2](#actualizar-desde-02)) en vez de un Brain vacío.

Para localizar la raíz del proyecto, sube por los ancestros del cwd buscando, en este orden de prioridad:

1. `.mobiai/brain/config.json` (un brain ya inicializado).
2. `.git/`.
3. `settings.gradle.kts` / `settings.gradle` (Android / KMP).
4. `pubspec.yaml` (Flutter).
5. `Package.swift`.
6. `*.xcworkspace` / `*.xcodeproj`.
7. `Podfile`.
8. `package.json` con dependencia `react-native`.

Si nada coincide, usa el directorio actual y avisa con un warning.

### Actualizar desde 0.2

Hasta 0.2.x, cada nota de un tipo se agregaba al final de un único archivo (`decisions.md`, `bugfixes.md`, …). Dos ramas que guardaban notas chocaban siempre al final de ese archivo. Desde 0.3 cada nota es su propio archivo en `memories/<sección>/<id>.md`, así que las notas nuevas nunca chocan.

**No hay que correr nada.** La primera vez que un comando (o una tool MCP) de 0.3 toca un Brain viejo:

1. crea un archivo por cada entrada vieja, conservando su id (las entradas escritas a mano sin id reciben uno que sale de su texto);
2. agrega un aviso justo debajo del título de `decisions.md`, `bugfixes.md` y `testing.md`;
3. lo avisa: `Brain migrated to the new format: N notes. Include .mobiai/brain/ in your next commit.`

Las entradas viejas quedan donde estaban. El resultado depende solo del texto de los archivos, así que dos personas que migran en dos ramas generan archivos idénticos y git los junta sin conflicto. Nunca hace commit. También repara las entradas cuyo cuerpo usaba títulos `##`, que 0.2 partía en entradas vacías.

**Los compañeros que siguen en 0.2.x** solo leen los archivos viejos. Su `brain context` y su `brain review` muestran el aviso, y su agente les pide actualizar. Una nota que guardan se agrega al archivo viejo, y 0.3 la convierte en archivo la próxima vez que corre. Un `promote` o `bump` hecho con 0.2.x sobre una nota migrada cambia solo la copia vieja y se ignora: primero hay que actualizar. Ojo: en 0.2.x, `brain review` sale con exit 1 por el aviso (tiene forma de entrada temporal vencida), salvo que pases `--no-fail`.

### `mobiai brain scan`

Recorre el árbol del proyecto (profundidad máx. 6, ignorando `.git`, `node_modules`, `build`, `Pods`, `DerivedData`, `.dart_tool`, etc.) y produce `.mobiai/brain/scan.json` con:

- `project_type`: `android` | `ios` | `kmp` | `flutter` | `react_native` | `unknown`.
- `platforms`: `android`, `ios`, `shared` (lo que se haya detectado).
- `build_systems`: `gradle`, `cocoapods`, `spm`, `npm`.
- `ui` / `di` / `network` / `persistence` / `serialization`: librerías reconocidas (Compose, SwiftUI, Koin, Hilt, Ktor, Retrofit, Room, DataStore, kotlinx.serialization, ...).
- `testing`: JUnit, MockK, Mockito, Turbine, Espresso, ...
- `integrations`: Firebase y otras integraciones detectables por strings en `build.gradle.kts` / `Podfile` / `package.json`.
- `ci_cd`: GitHub Actions, Bitrise, Codemagic, Fastlane.
- `warnings`: archivos sensibles detectados (sin leer su contenido).

`scan` requiere haber corrido `init` antes.

### `mobiai brain save <type>`

Cuatro subcomandos: `decision`, `bugfix`, `testing`, `knowledge`. Cada uno escribe la nota como su propio archivo, `memories/<sección>/<id>.md`. Todos comparten los mismos flags:

| Flag | ¿Requerido? | Notas |
|---|---|---|
| `--title <str>` | sí | Título corto, se usa como heading H2 |
| `--platform <plat>` | no | `android` \| `ios` \| `shared` \| `kmp` \| `flutter` \| `react-native` |
| `--area <str>` | no | Libre (`firebase_auth`, `dependency_injection`, `datastore`, ...) |
| `--status <s>` | no | `active` (default) \| `temporary` \| `deprecated` |
| `--review-after YYYY-MM-DD` | no | Sentido sobre todo en `temporary` |
| `--files a,b,c` | no | Lista de paths relevantes. Dicen a qué carpetas aplica la nota, y la nota se marca cuando cambian (ver `verify`) |
| `--body <md>` | no | Cuerpo Markdown. Si se omite, lee de stdin (útil para pipear contenido multilínea). Los títulos `##` pasan a `###` |
| `--supersedes <id>,…` | no | Esta nota reemplaza a esas: dejan de contar y su texto ya no se entrega. Las notas viejas no se modifican |
| `--until <dep>@<versión>,…` | no | La nota vale mientras el proyecto use una versión menor (por ejemplo `com.google.firebase:firebase-bom@34.0.0`). Cuando el proyecto llega a esa versión, la nota se marca para revisar |
| `--crash <firma>,…` | no | Solo `bugfix`: issue ID de Crashlytics, fingerprint de Sentry o firma de la excepción que resuelve este fix (se encuentra después con `search --crash`) |
| `--new` | no | Guarda aunque exista una nota con un título parecido (ver abajo) |

**`knowledge`** guarda cómo se comportan la app y los servicios con los que habla: happy paths, respuestas y códigos de error del servidor, comportamiento de SDKs y de terceros. Guarda solo lo que no se ve leyendo el código: "el servidor devuelve 409 si el pedido está cerrado" va; "la función X hace Y" no va.

**Un título parecido frena el guardado.** Si en la sección ya hay una nota activa cuyo título comparte el 60% o más de sus palabras, `save` no escribe nada y responde `A similar note already exists: <id> "<title>". Edit that note file, pass --supersedes <id> to replace it, or pass --new to save a separate note.`

Ejemplo:

```bash
mobiai brain save decision \
  --title "Use Koin como DI" \
  --platform shared \
  --area dependency_injection \
  --files "composeApp/src/commonMain/di/Module.kt" \
  --body "### Decision
Usar Koin como framework de DI para todo el código compartido.

### Reason
KMP-friendly, sin code generation, ya integrado con composeApp/iosApp."
```

**Guard**: si no existe `.mobiai/brain/config.json` el comando falla con un error claro y exit code 1. No crea el brain solo — sugiere correr `mobiai brain init` primero.

El tipo interno (`type:` en la entrada renderizada) se deriva del subcomando + status:

- `decision` → `architecture_decision`
- `bugfix` + `temporary` → `platform_workaround`
- `bugfix` + `active`/`deprecated` → `bug_fix`
- `testing` → `testing_pattern`
- `knowledge` → `knowledge`

Cada entrada recibe un `id` estable basado en el slug del título + timestamp UTC, así dos saves con el mismo título no colisionan.

### `mobiai brain search <query>`

Búsqueda case-insensitive sobre el título y el cuerpo de cada entrada de las memorias. Devuelve por sección, con título, status/platform, un snippet de la primera línea que contiene la query y el archivo de la nota.

```bash
mobiai brain search koin
mobiai brain search --platform ios firebase
mobiai brain search --status temporary
mobiai brain search --area firebase_auth keystore
mobiai brain search --crash 6a1b2c3d4e5f        # bugfixes para una firma de crash
```

Acepta los mismos filtros que `context` (`--platform`, `--status`, `--area`) con semántica **AND**. La query se combina con los filtros: `--platform ios firebase` devuelve solo entradas iOS que mencionen "firebase". Con `--crash` la query es opcional: busca coincidencias exactas en el campo `crash` de los bugfixes y el texto en el cuerpo de las notas.

### `mobiai brain show <id>`

Imprime una nota completa. Cuando `context` lista una nota en su índice, el agente abre el archivo de la nota; esto es para clientes que no pueden leer archivos (MCP: `mobile_get`).

### `mobiai brain review`

Audita la **deuda temporal** del Brain: recorre las memorias y muestra todas las entradas con `status: temporary` cuyo `review_after` ya pasó (`review_after <= hoy`). Es el comando que cierra el ciclo de "memoria con caducidad" — sin él, los workarounds temporales se vuelven permanentes por inercia.

```bash
mobiai brain review                        # solo vencidas, exit 1 si hay
mobiai brain review --include-no-date      # + temporales sin review_after asignado
mobiai brain review --no-fail              # siempre exit 0 (modo informativo)
```

**Output ejemplo:**

```
⚠ 2 entrada(s) temporal(es) vencida(s):

bugfixes.md
  ⚠ Podfile composeApp en minúsculas hasta CocoaPods 1.16
    review_after: 2026-08-15 (vencido hace 89 días)
    platform: ios
    id: podfile-composeapp-minusculas-20260512-094523

  ⚠ Suppress Deprecation en Compose 1.7
    review_after: 2026-10-01 (vencido hace 42 días)
    platform: android
    id: suppress-deprecation-compose-20260801-101200
```

**Exit codes:**

| Caso | Default | Con `--no-fail` |
|---|---|---|
| Hay vencidas | exit 1 | exit 0 |
| No hay vencidas | exit 0 | exit 0 |
| Brain no inicializado / error de IO | exit 1 (mensaje en stderr) | exit 1 |

El exit 1 por defecto está pensado para usarse como **gate en CI o pre-commit**: el build falla si alguien acumula workarounds temporales sin revisar. Si solo quieres informar sin bloquear, usa `--no-fail`.

**Qué entra y qué no:**

| `status` | `review_after` | ¿Aparece? |
|---|---|---|
| `temporary` | fecha pasada o hoy | **Sí** (vencida) |
| `temporary` | fecha futura | No |
| `temporary` | sin fecha | Solo con `--include-no-date` (sección separada) |
| `temporary` | malformada | Se ignora silenciosamente (corregí el `.md`) |
| `active` | cualquiera | No (no es deuda) |
| `deprecated` | cualquiera | No (ya está cerrada) |

`review` **no edita ni borra** nada — solo informa. Para bajar una entrada del listado, edita el `.md` y cambia `review_after`, el `status` a `active`/`deprecated`, o elimina el bloque.

**Notas que pueden estar desactualizadas.** `review` también lista, en una sección aparte, las notas marcadas con ⚠ (ver `verify`): sus archivos cambiaron o desaparecieron, llevan más de 3 meses sin verificar o se cumplió una condición `until`. Nunca cambian el código de salida, así que los gates de CI siguen fallando solo por temporales vencidas.

### `mobiai brain verify <id>`

Una nota que se apoya en el código puede quedar vieja sin que nadie lo note. El Brain la revisa contra git cada vez que lee las notas, con un solo `git log` para todo el Brain, y la marca con ⚠ cuando:

- alguno de sus `### Files` tiene commits de un día posterior a su `verified` (`⚠ LoginViewModel.kt changed since verification`), o ya no existe;
- no tiene archivos y no se verificó en los últimos 3 meses (`⚠ not verified in 3+ months`);
- se cumplió una condición `until` (`⚠ firebase-bom is now 34.1.0 (note valid until 34.0.0): check whether this workaround is still needed`) o esa dependencia ya no está. Las versiones se leen de los catálogos y build files de Gradle, `Podfile.lock`, `Package.resolved`, `pubspec.lock` y `package-lock.json`/`package.json`.

Nada vence ni se borra. Cuando tú (o el agente) revisas una nota marcada y sigue siendo cierta:

```bash
mobiai brain verify <id>                          # verified = hoy
mobiai brain verify <id> --files a/New.kt,b/X.kt  # además reemplaza sus archivos
```

Si ya no es cierta, corrige la nota o pásala a `deprecated` con `promote`. Una condición `until` que no se puede interpretar nunca marca nada; `mobiai doctor` las lista.

### `mobiai brain promote <id>`

Cambia el `status` de una entrada existente. Pensado como flujo de cierre tras `brain review`: cuando una entrada `temporary` ya no debería seguir siéndolo, la promovés a `active` (se volvió definitiva) o `deprecated` (ya no aplica).

```bash
mobiai brain promote <id> --status active
mobiai brain promote <id> --status deprecated
mobiai brain promote <id> --status temporary --review-after 2027-01-01

# Promover y eliminar review_after en una sola llamada
mobiai brain promote <id> --status active --clear-review-after
```

| Flag | ¿Requerido? | Notas |
|---|---|---|
| `--status <s>` | sí | `active` \| `temporary` \| `deprecated` |
| `--review-after YYYY-MM-DD` | no | También actualiza `review_after` |
| `--clear-review-after` | no | Elimina `review_after` (incompatible con `--review-after`) |

Para entradas tipo `bugfix`, el campo `type:` se recalcula automáticamente desde el nuevo `status` (`temporary` → `platform_workaround`, `active`/`deprecated` → `bug_fix`), manteniendo la invariante que `save` establece.

El body, los archivos en `### Files` y cualquier metadata custom se preservan byte-perfect — `promote` solo toca los campos que le pediste.

### `mobiai brain bump <id>`

Extiende solo el `review_after` de una entrada, sin tocar el `status`. Útil cuando un workaround temporal sigue siendo válido y querés correr el plazo de revisión:

```bash
mobiai brain bump <id> --review-after 2027-01-01
```

| Flag | ¿Requerido? | Notas |
|---|---|---|
| `--review-after YYYY-MM-DD` | sí | Nueva fecha |

Para cambios de `status`, usá `promote`. Para eliminar la fecha completamente, usá `promote --clear-review-after`.

### `mobiai brain context`

Lee `config.json`, `scan.json` (si existe) y todas las memorias. Imprime un Markdown listo para agentes:

```
# MobiAI Brain Context

Project: taykus-kmp
Type: Kotlin Multiplatform
Platforms: android, ios, shared

## Detected Stack
- UI: compose_multiplatform
- DI: koin
- Network: ktor
- Build systems: gradle

## Project Rules
- Prioritize clean architecture
- ...

## Architecture Decisions
...

## Known Bugfixes
...

## Testing Patterns
...

## Knowledge
...

## Release Notes
...

## Warnings
- archivo sensible detectado: .env (no leído)
```

**Hasta 25 KB, todo; por encima, un índice.** Se arma el documento completo y se mide. Si supera `context_max_bytes` (25600 por defecto, configurable en `config.json`), el agente recibe el stack, las reglas y los warnings, el texto completo de las decisiones activas (de la más nueva a la más vieja, mientras entren) y una línea por cada otra nota:

```
- [bug_fix] Login crash on screen rotation · android · auth · created 2026-09-14 (3 months ago) · verified 2026-09-14 · ⚠ LoginViewModel.kt changed since verification · applies to app/src/main/java/com/example/auth/ → memories/bugfixes/login-crash-on-screen-rotation-20260914-163044.md
```

El agente abre las notas que necesita (o usa `show` / `mobile_get`). Los filtros se aplican antes de medir, así que una consulta acotada sigue recibiendo los textos completos.

**Marcas que vas a ver** (en modo completo como líneas `- state:` / `- check:`, en el índice dentro de la línea):

- `unconfirmed`: la nota la escribió otra persona y su archivo todavía no llegó a la rama principal, así que nadie la revisó en un PR. Tus propias notas (sin commit todavía, o commiteadas con tu `git config user.email`) cuentan como confirmadas para ti. Una decisión sin confirmar no se toma como regla del proyecto. La rama principal es `origin/HEAD` (por ejemplo `develop` en equipos con git-flow), después `origin/main`, `origin/master`, `main`, `master`; se puede forzar con `"confirmed_branch"` en `config.json`. Sin git, todo cuenta como confirmado.
- `⚠ …`: la nota puede estar desactualizada (ver `verify`).
- `superseded by <id>`: una nota más nueva la reemplaza; su texto ya no se entrega. Si el reemplazo todavía está sin confirmar, la nota vieja sigue vigente y muestra `replacement proposed by <id> (unconfirmed)`.

**Filtros disponibles**:

| Flag | Notas |
|---|---|
| `--section` | Coma-separado o repetible. Nombres canónicos: `stack`, `rules`, `decisions`, `bugfixes`, `testing`, `knowledge`, `releases`, `warnings` (`integrations` es un alias de `knowledge`). |
| `--platform` | Filtra entradas por `platform:` (exacto, case-insensitive). |
| `--status` | Filtra por `status:` (exacto). |
| `--area` | Filtra por `area:` (substring). |

Los filtros aplican solo a entradas de memorias — `stack`, `rules` y `warnings` se incluyen/excluyen únicamente vía `--section`. Si tras filtrar una sección no quedan entradas, aparece `_No entries match the current filter._` (distinto del `_No entries yet._` para secciones genuinamente vacías).

```bash
# Solo bugfixes temporales para iOS
mobiai brain context --section bugfixes --platform ios --status temporary

# Solo el stack detectado, sin memorias
mobiai brain context --section stack

# Decisiones + bugfixes de plataforma shared
mobiai brain context --section decisions,bugfixes --platform shared
```

---

## Ejemplos end-to-end

Tres historias completas para hacer tangible el valor del Brain.

### 1. Workaround temporal de Firebase en iOS

**Lunes 9:00** — Te explota el build de iOS tras subir Firebase a 11.x. Después de 40 minutos descubres que `composeApp` tiene que estar en minúsculas en el `Podfile` hasta que CocoaPods 1.16 salga con el fix.

**Lunes 9:45** — Arreglas el bug con `mobiai-fix-issue`. Al final del flujo, la skill detecta que tienes `.mobiai/brain/config.json` y propone:

> *Detecté que esto fue un workaround más que un fix de raíz. ¿Lo guardo en el Brain como bugfix temporal para revisar en 3 meses?*

Confirmas. La skill ejecuta:

```bash
mobiai brain save bugfix \
  --title "Podfile composeApp en minúsculas hasta CocoaPods 1.16" \
  --platform ios \
  --area cocoapods \
  --status temporary \
  --review-after 2026-08-15 \
  --files "iosApp/Podfile" \
  --body "..."
```

**Tres meses después** — Otro miembro del equipo (o tu yo del futuro) ve un crash relacionado. Pregunta al agente, que invoca `mobile_search "cocoapods composeApp"` y encuentra la entrada con todo el contexto: cuándo se decidió, qué archivo, cuándo revisar. Cinco minutos de investigación en lugar de cuarenta.

### 2. Decisión de DI en un proyecto KMP

**Brainstorm de arquitectura** — Necesitas elegir framework de DI para tu app KMP nueva. Lanzas `mobiai-mobile-brainstorming`, comparas Koin vs Hilt+Dagger vs manual, llegas a una conclusión.

Tras el User Review Gate, la skill propone guardar la decisión. Confirmas y se guarda como `architecture_decision` en `memories/decisions/` con el razonamiento, los trade-offs descartados y los archivos donde vive.

**Dos meses después** — Llega un dev nuevo, abre el repo en Cursor. Pregunta "¿por qué Koin y no Hilt?". El agente invoca `mobile_search "DI"` y le responde con la decisión completa, fecha y motivos. No tiene que reabrir el debate.

### 3. Patrón de testing no obvio con DataStore

**Tests intermitentes** — Tu test de DataStore pasa local pero falla en CI 1 de cada 4 veces. Investigas con `mobiai-mobile-debugging`: resulta que `dataStore.edit { it.clear() }` no flushea sincrónicamente, hay una ventana donde `.data.first()` devuelve el valor previo.

Aprendes que el patrón correcto es esperar `dataStore.data.first { it.asMap().isEmpty() }` después del clear. Aplicas el fix con `mobiai-write-tests`.

Al final, la skill propone:

> *Este patrón parece reutilizable para otros tests de DataStore. ¿Lo guardo como testing pattern?*

Confirmas. Se guarda en `memories/testing/`. La próxima vez que alguien (o tú) escriba un test contra DataStore, `mobile_context --section testing` se lo va a mostrar antes de que tropiece con el mismo flake.

---

## Arquitectura interna

### Layout en disco

Todo lo que sabe el Brain vive en `<repo>/.mobiai/brain/`:

```
.mobiai/brain/
├── config.json          ← metadatos del proyecto + reglas custom
├── scan.json            ← stack detectado (regenerable con `scan`)
└── memories/
    ├── decisions/<id>.md   ← architecture_decision
    ├── bugfixes/<id>.md    ← bug_fix + platform_workaround
    ├── testing/<id>.md     ← testing_pattern
    ├── knowledge/<id>.md   ← knowledge
    ├── decisions.md        ← formato viejo (0.2): entradas viejas + aviso
    ├── bugfixes.md         ← ídem
    ├── testing.md          ← ídem
    ├── integrations.md     ← formato viejo: las notas escritas a mano migran a knowledge/
    └── releases.md         ← reservado
```

Solo `config.json` es obligatorio. El resto se crea durante `init` y se va llenando con el uso. No hay un archivo de índice en el repo: el CLI lo arma al leer las notas, así que no puede generar conflictos de merge. Las notas se ordenan por `date` y después por `id`.

### Formato de las entradas

Las memorias son Markdown plano legible a ojo. Cada archivo de nota tiene un título H2 (`##`) seguido de metadata YAML-ish y un cuerpo libre:

```markdown
## Podfile composeApp en minúsculas hasta CocoaPods 1.16

- id: podfile-composeapp-minusculas-20260512-094523
- type: platform_workaround
- status: temporary
- platform: ios
- area: cocoapods
- date: 2026-05-12
- verified: 2026-05-12
- review_after: 2026-08-15

### Reason
CocoaPods 1.15.x se rompe con módulos en CamelCase cuando ...

### Files
- iosApp/Podfile
```

Solo la primera línea `##` es el título: un `##` dentro del cuerpo queda en el cuerpo. Campos opcionales: `verified` (arranca igual a `date`; lo actualiza `verify`), `supersedes`, `until`, `crash` y `paths` (globs separados por coma, como `feature/**/checkout/*.kt`, que reemplazan a las carpetas que salen de `### Files` para el hook por archivo). Las marcas `unconfirmed`, `⚠` o `superseded by` no se escriben en los archivos: se calculan al leer las notas.

### IDs estables

Cada entrada recibe un ID generado al guardarla: `<slug-del-título>-<YYYYMMDD-HHMMSS>` en UTC. Por ejemplo:

```
use-koin-como-di-20260301-143025
```

Esto permite dos cosas:

- **Dedup**: dos saves con el mismo título no colisionan (timestamps distintos).
- **Referencia**: si necesitas linkar a una entrada desde otra, el ID es estable aunque edites el título.

### Atomicidad de writes

Las escrituras a `config.json`, `scan.json` y a los archivos de nota usan el patrón **temp + rename**: escribimos primero a un archivo temporal en el mismo directorio y luego hacemos `rename` atómico al destino. Si el proceso se mata a media escritura, no quedan archivos corruptos.

### Scanner: alcance y límites

El `scan` recorre el árbol con estos parámetros:

- **Profundidad máxima**: 6 niveles desde la raíz. Cubre monorepos típicos sin perderse en `node_modules` o caches.
- **Tamaño máximo por archivo**: 1 MiB. Evita cargar binarios o artefactos generados.
- **Dirs ignoradas**: `.git`, `.mobiai`, `.gradle`, `.idea`, `.kotlin`, `.dart_tool`, `.swiftpm`, `build`, `out`, `target`, `node_modules`, `Pods`, `DerivedData`, `vendor`, `.claude`, `.cursor`, `.copilot`, `.gemini`, `.codex`, `.junie`.
- **Archivos sensibles**: `.env`, `google-services.json`, `GoogleService-Info.plist`, `local.properties`, `keystore.properties` se **detectan** (warning) pero **nunca se leen**.

La detección de librerías funciona por substring search (case-insensitive) sobre archivos de configuración (`build.gradle.kts`, `Podfile`, `package.json`, `pubspec.yaml`...). No hay parsing AST — es deliberadamente simple, mantenible y fast.

### ¿Por qué Markdown y no SQLite?

A escala de 1-100 entradas por proyecto (el rango realista hoy), Markdown gana en todas las dimensiones que importan:

- **Legible a ojo**: puedes abrir `decisions.md` en cualquier editor y entender el contenido.
- **Diff-friendly**: cada save es un archivo nuevo, fácil de revisar en un PR, y dos ramas que guardan notas nunca chocan.
- **Sin runtime**: no requiere base de datos, no hay schema migrations.
- **Editable manualmente**: si necesitas corregir algo, edita el `.md` directamente.

Si algún proyecto supera ~500 entradas y la búsqueda se vuelve lenta, migraremos a SQLite + FTS5 manteniendo el mismo CLI y formato de export. Por ahora, no hay razón.

---

## Integración con Skills

Las mismas 4 skills tienen **dos puntos de contacto** con el Brain — uno al entrar, otro al salir.

### Hook de salida: proponer `save`

Las skills proponen guardar al final de su flujo, todas con el mismo patrón: **solo si** existe `.mobiai/brain/config.json` en el proyecto, y **siempre** pidiendo aprobación al usuario antes de invocar el save.

| Skill | Cuándo propone | Tipo guardado |
|---|---|---|
| `mobiai-fix-issue` | Tras Phase 6 (verification), antes del PR gate | `bugfix` (status según haya fix real o workaround) |
| `mobiai-write-tests` | Tras Step 4 (tests pasan), si el test captura un patrón no obvio | `testing` |
| `mobiai-mobile-debugging` | Tras Phase 6 (root cause confirmada), **solo en invocación standalone** — si se invocó desde fix-issue, se salta (fix-issue ya tiene su propio hook) | `bugfix` |
| `mobiai-mobile-brainstorming` | Tras User Review Gate (spec aprobada), una entrada por cada decisión arquitectónica no trivial del spec | `decision` |

Si el proyecto **no** tiene brain, el hook se salta sin ruido. El agente nunca invoca `save` solo: siempre pide una línea de confirmación al usuario antes de ejecutarlo.

### Hook de entrada: pre-flight de `review`

Las mismas 4 skills, al **arrancar**, invocan `mobile_review` (CLI fallback: `mobiai brain review --no-fail`), filtran las vencidas por `platform` y `area` del trabajo en curso, y mencionan al usuario las que matchean antes de empezar:

> "Heads up — there are N overdue workaround(s) in `<platform>/<area>` that may be relevant: `<title>`. I'll proceed; let me know if you want to look at those first."

No pausa el flujo. El usuario decide si quiere ir a `brain promote` / `brain bump` antes o continuar y revisarlo después. `mobile-debugging` y `write-tests` skipean el pre-flight cuando se invocan anidadas desde `fix-issue` (que ya hizo el suyo) para no doble-promptear.

**Doble path MCP / CLI**: cada uno de los hooks documenta dos rutas equivalentes. Si el cliente tiene `mobiai-brain` registrado como server MCP, el agente invoca la tool `mobile_save_*` directamente — más rápido, sin spawn de proceso, output estructurado. Si no, cae al `mobiai brain save ...` de siempre. Ambas terminan escribiendo exactamente el mismo archivo de nota.

### Entrega por archivo: el hook `PostToolUse`

El plugin `core` también registra un hook que corre después de que el agente lee o edita un archivo (Claude Code: `PostToolUse` sobre `Read|Edit|Write`; Cursor: `postToolUse` sobre `Read|Write`). Llama a `mobiai brain hook file-context`, que le entrega al agente, junto al resultado de la tool, las notas cuyas carpetas contienen ese archivo (las de `### Files`, o las de `paths:`). Cada nota se entrega como mucho una vez por sesión, hasta unos 4 KB por llamada; las que no entran se listan por id. Nunca escribe el Brain, y no dice nada si no hay Brain o si mobiai es anterior a 0.3. El resto de las herramientas se apoya en las carpetas `applies to` del índice de `context`.

Otras skills que podrían integrarse en el futuro: `mobiai-crashlytics` (crashes recurrentes → `save bugfix`), `mobiai-mobile-planning` (planes con decisiones embebidas → `save decision`).

---

## FAQ y troubleshooting

### ¿Debo commitear `.mobiai/brain/` en git?

**Sí, es lo recomendado.** El Brain es conocimiento del equipo: decisiones, workarounds y patrones que quieres compartir con los demás miembros (humanos y agentes IA). Versionar `.mobiai/brain/` en git hace que ese conocimiento viaje con el repo.

La única excepción razonable sería un proyecto experimental donde solo tú usas el Brain como notebook personal — ahí podrías añadirlo a `.gitignore`.

### ¿Se sincroniza el Brain entre máquinas?

No tiene sincronización nativa. La sincronización pasa por git: si lo commiteas, tus compañeros lo tienen al hacer `git pull`. No hay servidor central, no hay cuenta, no hay nube. Simple por diseño.

### ¿Qué pasa si edito un archivo de nota a mano?

Está permitido y soportado. El parser re-lee los archivos en cada `context` / `search`, así que cualquier edición manual se refleja inmediatamente. Las únicas restricciones son:

- Mantén la estructura `## Título` + lista de metadata + cuerpo.
- Mantén el nombre del archivo igual al `id:` (una nota por archivo).
- Si añades metadata custom, no la garantizamos para filtros futuros.

### ¿Cómo borro o modifico una entrada?

**Para cambiar status** (cerrar un workaround, deprecar una decisión): usa `mobiai brain promote <id> --status deprecated` (o `active`).

**Para extender el plazo de revisión** de un workaround vigente: usa `mobiai brain bump <id> --review-after YYYY-MM-DD`.

**Para reemplazar una nota** por otra más nueva: guarda la nueva con `--supersedes <id>`.

**Para borrado real**: borra el archivo de la nota. No hay `mobiai brain delete` por diseño — preferimos marcar como `deprecated` para preservar la historia. Cuidado con las notas que vinieron del formato 0.2: su copia vieja sigue en `decisions.md`/`bugfixes.md`/`testing.md`, así que si borras el archivo de la nota vuelve a aparecer en la próxima corrida. Pásalas a `deprecated`, o borra también su bloque del archivo viejo (por ejemplo, datos sensibles guardados por error).

### ¿Qué significa `unconfirmed`?

La nota la escribió otra persona y todavía no llegó a la rama principal, así que nadie la revisó en un PR. El agente la usa con cautela en vez de tomarla como regla del proyecto. Se confirma sola cuando se mergea el PR que la agrega (después de un `git fetch`). Si una nota tuya aparece sin confirmar, probablemente tu `git config user.email` es distinto del email del commit que la agregó.

### ¿Por qué el scanner no detecta mi librería X?

La detección funciona por lista hardcoded de strings conocidos (Koin, Ktor, Compose, etc.). Si tu librería no aparece, abre un issue en [MobiAI-Core](https://github.com/ArisGuimera/MobiAI-Core/issues) con el nombre del paquete y un snippet del archivo de configuración donde aparece. Añadirla es trivial — ver `cli/internal/brain/detect_deps.go`.

### ¿Funciona offline?

Completamente. Todo es local: Markdown + JSON en disco, ninguna llamada de red. El servidor MCP también es local (stdio). Esto es por diseño: tus decisiones de arquitectura son tuyas.

### ¿Mis secretos pueden acabar en el Brain?

No. El scanner detecta archivos sensibles (`.env`, `google-services.json`, `GoogleService-Info.plist`, `local.properties`, `keystore.properties`) pero **nunca lee su contenido**. Solo los registra como `warnings` con su nombre, para que sepas que existen.

Las entradas `save` solo guardan lo que tú (o el agente con tu aprobación) le pasas explícitamente vía `--body` y `--files`. Si nunca le pasas un secreto, nunca se guarda uno.

### ¿Cuál es el cliente MCP recomendado?

Cualquiera de los cinco soportados (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI) funciona igual de bien — todos hablan el mismo protocolo MCP. Usa el que ya usas para programar. Setup paso a paso en [`MCP-SETUP.md`](MCP-SETUP.md).

### ¿Puedo tener más de un Brain por repo?

No. La detección de raíz busca el primer `.mobiai/brain/config.json` subiendo desde el cwd, así que solo el más cercano gana. Si necesitas separar conocimiento de subproyectos, usa la flag `--area` o `--platform` en los saves para segmentar.

### ¿Qué pasa con los monorepos?

Cada monorepo tiene **un solo Brain** en la raíz. La forma de segmentar conocimiento entre subproyectos es vía `--area` (por ejemplo `--area app-android` vs `--area app-ios` vs `--area shared-kmp`). Los filtros de `context` y `search` te dejan recuperar solo lo relevante para una zona del monorepo.

### Me salta "brain no inicializado" pero corrí `init`

Probablemente estás en un subdirectorio y la detección de raíz no encuentra el `config.json` esperado. Soluciones:

- Asegúrate de haber corrido `mobiai brain init` desde la raíz del proyecto (donde está `.git/` o `settings.gradle.kts` o `Package.swift`).
- O usa `--root <ruta-absoluta>` para apuntar explícitamente al proyecto.

### `scan` detecta cosas raras / no detecta nada

`scan` tiene profundidad máxima 6 desde la raíz. Si tu monorepo tiene módulos a profundidad 7+, no los va a ver. Para casos extremos, considera ejecutar `scan` desde el subdirectorio del módulo con `--root` apuntando ahí (aunque solo se reflejará en `scan.json` de ese brain).

---

## Roadmap

**Fase 1** — `init`, `scan`, `context`. ✓
**Fase 2** — `save decision|bugfix|testing` + hook en `mobiai-fix-issue`. ✓
**Fase 3** — `search` + filtros (`--section`, `--platform`, `--status`, `--area`) en `context`. ✓
**Fase 4** — hooks de save en `mobiai-write-tests` (testing pattern), `mobiai-mobile-debugging` (bugfix sin pasar por fix-issue) y `mobiai-mobile-brainstorming` (decision tras spec aprobada). ✓
**Fase 5** — servidor MCP (`mobiai brain mcp`) que expone las 6 operaciones como tools nativas para Claude Code, Cursor, Copilot CLI, Codex y Gemini CLI. El brain pasa a estar siempre en el toolbox del agente. Setup en [`MCP-SETUP.md`](MCP-SETUP.md). ✓
**Fase 6** — `mobiai brain review` (CLI) + `mobile_review` (MCP tool) para auditar entradas `status: temporary` cuyo `review_after` ya pasó. Cierra el ciclo de "memoria con caducidad" — los workarounds temporales ya no se vuelven permanentes por inercia. ✓
**Fase 7** — `mobiai brain promote` y `mobiai brain bump` (CLI) + `mobile_promote` y `mobile_bump` (MCP tools) para cerrar el ciclo de vida de una entrada desde CLI/agente: cambiar `status` o extender `review_after` sin editar el `.md` a mano. Atomic rewrite; preserva body, files y metadata custom byte-perfect. ✓
**Fase 8** — los 4 hooks de skills (`mobiai-fix-issue`, `mobiai-write-tests`, `mobiai-mobile-debugging`, `mobiai-mobile-brainstorming`) ahora documentan la tool MCP equivalente como ruta preferida, manteniendo el CLI como fallback explícito. Cuando el cliente tiene `mobiai-brain` registrado como server MCP, el agente invoca `mobile_save_*` directamente; si no, cae al `mobiai brain save ...` de siempre. ✓
**Fase 9** — pre-flight automático en las 4 skills: antes de empezar el trabajo, el agente invoca `mobile_review` (o `mobiai brain review --no-fail`), filtra por platform/area del trabajo en curso, y menciona al usuario cualquier workaround vencido relevante. No pausa: el usuario decide si quiere revisarlo primero. Cierra el bucle "memoria con caducidad" al hacer la deuda **proactivamente visible** en el momento exacto en que importa. ✓
**Fase 10** — `mobiai brain install-mcp` registra el server MCP en clientes soportados (Claude Code y Cursor en v1) con un solo comando. Detecta automáticamente qué clientes tiene instalados el usuario, preserva el resto del config, es idempotente y reversible (`--uninstall`). Path absoluto al binario vía `os.Executable()` para que el cliente AI lo encuentre incluso con `$PATH` distinto al del shell. Eliminamos la fricción de tener que editar JSON a mano. ✓
**Brain 0.3** — una nota por archivo (se terminan los conflictos de merge) con migración automática y un aviso para las versiones anteriores; `context` pasa a índice por encima de 25 KB; tipo `knowledge`; notas `unconfirmed` hasta que llegan a la rama principal; verificación contra git (`verify`, marcas ⚠); hook de entrega por archivo; `supersedes` y freno por títulos parecidos; condiciones `until` atadas a versiones de dependencias; firmas de `crash` y `search --crash`. ✓

**Próximos pasos**:

- **`install-mcp` para Copilot CLI, Codex y Gemini CLI** — el v1 cubre Claude Code y Cursor (mismo shape de JSON). Los otros tres tienen formatos distintos (Copilot/Gemini JSON con shape propia, Codex TOML) y se añaden cuando haya demanda.

- **`brain review --upcoming N`** — además de las vencidas, listar entradas temporales que vencen en los próximos N días (default propuesto: 30). Da margen de planning para sprints/retros sin esperar a que algo esté ya vencido. Pendiente.
- **`save release`** (categoría de menor volumen, baja prioridad).
- **`mobiai brain lint`**: validar los archivos de nota y buscar instrucciones sospechosas, para CI.

**Fase futura** — migración a SQLite + FTS5 si Markdown se queda corto (>~500 entradas por proyecto). Para el rango actual (1-100 entradas) los filtros sobre Markdown bastan.
