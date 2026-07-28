# MobiAI Graph

> **Read in another language:** **English** · [Español](README.es.md)

> Semantic exploration of mobile code, local and regenerable. MobiAI Graph understands Android, iOS, KMP, Flutter, and React Native apps.

## What MobiAI Graph is

MobiAI Graph is an agent **capability**, not a separate product: semantic exploration of mobile code over a local index that's regenerated on demand. It's not a service, it's not a remote database, it requires no account or cloud.

It lives in `<repo>/.mobiai/graph/index.json` — local-first, regenerable, no cloud. In V1 it covers **Kotlin** and **Swift**, the two native languages of the mobile stack MobiAI prioritizes.

## Skills vs Brain vs Graph

MobiAI has three complementary pieces. Each solves a different problem:

| Piece | What it stores | Who maintains it | Key property |
|---|---|---|---|
| Skills | How to work (process, flows) | The MobiAI team (in plugins) | Global, shared across projects |
| Brain | What was decided, which workarounds are active | The project's user | Human historical memory, diff-friendly |
| Graph | Current structure of the code | Auto-generated | Computed, regenerable, don't edit by hand |

The three are complementary. A well-made decision combines all three: the skill sets the flow, Brain adds historical context, Graph points to where the affected code lives today.

## How V1 works

Full pipeline in a single pass (`mobiai graph init`):

```
filepath.WalkDir(root)
   │
   ├─ skip dirs: .git .gradle .idea .vscode .mobiai
   │             .claude .claude-plugin .worktrees
   │             build dist out target Pods DerivedData node_modules
   │
   ├─ for each .kt / .swift:
   │     1. read file
   │     2. strip comments + strings (preserving line numbers)
   │     3. regex line-by-line → symbols (class/fun/struct/...)
   │     4. brace-depth stack → assign container to the symbol
   │
   └─ sort by path → atomic write JSON (.tmp + fsync + rename)
```

**Firm technical decisions**:

- **Regex line-by-line, not AST**. For two languages and this level of detail (top-level and nested symbols, no type resolution), regex is enough and eliminates heavy dependencies like tree-sitter. We'll reconsider when a third language enters.
- **JSON, not SQLite**. The index fits comfortably in JSON (real KMP project measured: 647 files, 2464 symbols). JSON is diff-friendly, inspectable, and keeps zero external dependencies in the Go CLI.
- **No watcher**. `mobiai graph init` regenerates the whole index in milliseconds. A watcher with `fsnotify` is parked for V2 once it's justified.
- **Strip comments+strings preserving lines**. The regexes would hit false positives in docstrings and literals. The strip replaces with spaces (it doesn't delete) — that way the indexer's reported line numbers point to the real spot in the file.
- **Brace-depth stack for containers**. The most-nested symbol wins: a `fun login()` inside `class LoginViewModel` registers `container = "LoginViewModel"`.
- **Atomic write**. Writing `.tmp` + `fsync` + `rename` guarantees that a Ctrl-C mid-`init` never corrupts `index.json`.
- **Zero external dependencies**. Go stdlib + `cobra` (already in the CLI). No CGo, no tree-sitter, no SQLite.

Real usage example:

```
$ mobiai graph init
✓ Index generated: /Users/.../project/.mobiai/graph/index.json
  Files: 1
  Symbols: 1
  Kotlin: 1
  Swift: 0

$ mobiai graph search Login
app/LoginViewModel.kt:5 class LoginViewModel
```

The index is **regenerable**: if the regexes evolve or the code changes, just re-run `mobiai graph init`.

## Real benchmark: Graph vs grep

Measured on a **real production KMP project** (Kotlin Multiplatform, Android + iOS).

**Index context**:

- Files indexed: **647**
- Symbols: **2464** (635 Kotlin + 12 Swift)
- Time for `mobiai graph init`: **0.47 s**
- Base commit: `dcbe040` · Date: 2026-05-25

Tokens estimated as `bytes / 4` (standard GPT/Claude proxy).

| Question | Tool | Time | Bytes | Tokens ≈ | Winner |
|---|---|---|---|---|---|
| **Q1** — List every ViewModel in the project | `mobiai graph search ViewModel` | 0.01 s | 9446 | 2361 | — |
|  | `rg "class \w+ViewModel"` | 0.15 s | 7426 | 1856 | **grep** (more compact output) |
| **Q2** — Where is `AppScaffold` used? | `mobiai graph callers AppScaffold` | 0.07 s | 6766 | 1691 | — |
|  | `rg "AppScaffold"` | 0.01 s | 6888 | 1722 | **technical tie** |
| **Q3** — Context for login + refresh token fix | `mobiai graph context "fix login bug refresh token"` | 0.01 s | 1021 | 255 | **graph** (6.3× fewer tokens) |
|  | `rg -l -i "login\|refresh.?token"` | 0.03 s | 6403 | 1600 | — |
| **Q4** — Does `AuthRepository` exist and where? | `mobiai graph search AuthRepository` | 0.01 s | 237 | 59 | **graph** (7.5× fewer tokens) |
|  | `rg "AuthRepository"` | 0.03 s | 1762 | 440 | — |

### Aggregated totals

```
graph │██████████████████████░░░░░░  4366 tokens · 0.10 s
grep  │████████████████████████████  5618 tokens · 0.22 s
                                     ─22 % tokens
```

- graph tokens: **4366** · grep: **5618** → **overall savings ≈ 22%**
- Total graph time: **0.10 s** · grep: **0.22 s**

## How an AI agent uses it

Graph exists so AI agents start a task **knowing where to look**, instead of guessing with grep.

### Brain → Graph → action pattern

The canonical flow when an agent gets a task on a mobile project:

1. **Brain** brings the historical context: "this crash already happened, was fixed with workaround X, no rollback".
2. **Graph** brings the current structure: "today's relevant files are these 10, ranked by score".
3. **Action**: the agent reads only the top files, not the 74 grep would dump.

This reduces tokens read and eliminates speculative reads of irrelevant files.

### Pre-flight pattern

Downstream skills (`mobiai-mobile-debugging`, `mobiai-fix-issue`, `mobiai-write-tests`, `mobiai-review-code`) run at startup:

```bash
# 1. Does the index exist?
test -f .mobiai/graph/index.json || mobiai graph init

# 2. Is it fresh? (status shows "Xm/Xh/Xd ago")
mobiai graph status

# 3. Query relevant to the issue:
mobiai graph context "<bug/feature description>"
mobiai graph callers <AffectedSymbol>
```

Pre-flight runs **once at the start of the conversation**, not per subcommand. The results stay in the agent's context for reuse.

### Typical case: "fix the login + refresh token bug"

Without Graph: the agent runs `rg -l -i "login|refresh.?token"` and gets **6403 bytes** (~1600 tokens) of flat paths with no clue which one is central — tests, fakes, repositories, viewmodels, and screens all mixed together.

With Graph:

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

**1021 bytes (~255 tokens)** ranked by score: the agent reads the screen, the ViewModel, and the repository that touch the token, and proposes the fix. **6.3× fewer tokens, all relevant.**

## Supported languages

| Language | V1 | Roadmap V2/V3 |
|---|---|---|
| Kotlin | ✅ | improve precision with AST |
| Swift | ✅ | improve precision with AST |
| Dart (Flutter) | ❌ | V2 |
| JavaScript / TypeScript (RN) | ❌ | V2 |
| Java | ❌ | V2 |
| KMP `expect/actual` semantics | partial (each side indexed separately) | V3 |

## Inspiration

The idea of a local semantic graph to accelerate AI agents isn't new. [CodeGraph](https://github.com/colbymchenry/codegraph) (TypeScript + tree-sitter + SQLite/FTS5) popularized it in web projects. MobiAI Graph takes that idea and brings it to mobile ground, with a native Go stack and a focus on languages and patterns that matter for Android, iOS, KMP, Flutter, and React Native.

## Roadmap (future, not yet implemented)

- **V1.2** (next concrete step):
  - Graph pre-flight integrated into downstream skills (`mobiai-mobile-debugging`, `mobiai-fix-issue`, `mobiai-write-tests`, `mobiai-review-code`) — without this, AI agents won't invoke Graph autonomously.
- **V2** (future):
  - Dart, JavaScript/TypeScript, and Java support.
  - Incremental watcher — the index refreshes when files change.
  - Exact search (`--exact`) and regex search (`--regex`).
  - Possible migration to tree-sitter if regexes fall short (decision pending).
- **V3** (future):
  - Mobile-specific semantics: Compose `@Composable` → ViewModel → UseCase → Repository.
  - KMP `expect/actual` mapped to platform implementations.
  - Navigation Compose / SwiftUI routes → screens.
  - Hilt / Koin module bindings.
- **V4** (speculative, no date):
  - A dedicated MCP server so AI clients query Graph as a native tool.

None of these pieces are available today. The documentation will be updated when they are.

## Policy

- MobiAI never installs external backends automatically. Graph is native Go and lives inside the `mobiai` CLI.
- The index (`.mobiai/graph/index.json`) is **regenerable**: committing it is not required. If your team prefers to keep it out of the repo, gitignore it without worry — it's not critical information.
- Graph **does not replace Brain**: if a decision, workaround, or convention needs to persist across code changes, it goes to Brain. Graph only reflects the current structure of the code.
- Zero user data leaves the machine. Graph is 100% local.
