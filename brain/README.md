# MobiAI Brain

> **Read in another language:** **English** · [Español](README.es.md)

> The **per-project living memory** of the MobiAI ecosystem. While skills teach the AI *how* to work, Brain stores the *project-specific knowledge* of each mobile codebase: decisions, bugfixes, workarounds, testing patterns, and knowledge about how the app and its services behave.

**Table of contents**

- [Quickstart (60 seconds)](#quickstart-60-seconds)
- [What does it solve?](#what-does-it-solve)
- [How it fits with Skills and CLAUDE.md](#how-it-fits-with-skills-and-claudemd)
- [Philosophy](#philosophy)
- [Commands](#commands)
- [End-to-end examples](#end-to-end-examples)
- [Internal architecture](#internal-architecture)
- [Integration with Skills](#integration-with-skills)
- [FAQ and troubleshooting](#faq-and-troubleshooting)
- [Roadmap](#roadmap)

---

## Quickstart (60 seconds)

Assuming you already have the `mobiai` binary installed (see the [main README](../README.md) if not):

```bash
# 1. At the root of your mobile project
cd ~/code/my-app

# 2. Initialize Brain (idempotent — won't break anything if it already exists)
mobiai brain init

# 3. Let it detect your stack
mobiai brain scan

# 4. See what it found
mobiai brain context
```

From here on, MobiAI skills will **propose** saving decisions, bugfixes, and testing patterns at the end of their flows. You can also save manually:

```bash
mobiai brain save decision --title "Use Koin for DI" --platform shared \
  --body "KMP-friendly, no code generation, integrated into composeApp/iosApp."
```

And query whenever you need to:

```bash
mobiai brain search firebase
mobiai brain context --section bugfixes --platform ios --status temporary
```

**Recommended:** register Brain as an MCP server in your client. For Claude Code and Cursor it's one command:

```bash
mobiai brain install-mcp
```

It detects which clients are present and registers `mobiai-brain` while preserving the rest of the config. For other clients (Copilot CLI, Codex, Gemini CLI) follow the manual registration in [`MCP-SETUP.md`](MCP-SETUP.md).

---

## What does it solve?

Every mobile project accumulates *changing historical knowledge* that doesn't fit anywhere comfortably:

- **"We use Koin, not Hilt"** — decision from March 2026, still active. It's not methodology (doesn't belong in a skill), but too project-specific for CLAUDE.md.
- **"FirebaseAuth on iOS needs `composeApp` lowercase until the next CocoaPods major"** — temporary workaround. If you don't track it, in six months nobody remembers why it's like that.
- **"The DataStore test has to wait on `dataStore.data.first { it.asMap().isEmpty() }` after `clear()`"** — non-obvious pattern you discovered the hard way. You'll need it again.

Without Brain, that knowledge lives in buried commits, the team Slack, or the head of whoever discovered it. With Brain it lives **inside the repo**, structured, searchable, and accessible to any AI agent connected to the project.

---

## How it fits with Skills and CLAUDE.md

| | Skills | CLAUDE.md / AGENTS.md / GEMINI.md | MobiAI Brain |
|---|---|---|---|
| **What it stores** | Methodology, best practices | Stable project rules | Changing historical knowledge |
| **Scope** | Global (shared across projects) | Per project | Per project |
| **Structure** | SKILL.md | Free Markdown | Structured Markdown + JSON |
| **Does it expire?** | No (updated via PR) | Rarely | Often (temporary workarounds, revisitable decisions) |
| **How is it queried?** | Loaded at agent start | Loaded at agent start | On-demand: `context`, `search`, MCP tools |

The three coexist. CLAUDE.md tells you **how** to work; Brain tells you **what was decided before**.

---

## Philosophy

- **Per project**, not global. Every mobile repo has its own local brain in `<repo>/.mobiai/brain/`.
- **Local-first**. Plain Markdown, simple JSON. No SQLite, embeddings, or cloud for now.
- **Mobile-first**. The differentiator is understanding Android / iOS / KMP / Flutter / React Native, not storing generic text.
- **No secrets**. The scanner detects sensitive files (`.env`, `GoogleService-Info.plist`, `google-services.json`) and only records their existence, never their content.

---

## Commands

```bash
mobiai brain init                  # Creates .mobiai/brain/ in the current project (idempotent)
mobiai brain scan                  # Detects stack: Android, iOS, KMP, Flutter, RN, libraries, CI/CD
mobiai brain context               # Prints Markdown with config + scan + memories (accepts filters)

mobiai brain save decision ...     # Save an architecture decision
mobiai brain save bugfix ...       # Save a bugfix or workaround
mobiai brain save testing ...      # Save a reusable testing pattern
mobiai brain save knowledge ...    # Save how the app or an external service behaves

mobiai brain search <query>        # Free-text search across memories (or --crash <signature>)
mobiai brain show <id>             # Print one note in full
mobiai brain review                # Overdue temporary entries + notes that may be out of date
mobiai brain promote <id> ...      # Change the status of an existing entry
mobiai brain bump <id> ...         # Extend review_after on an existing entry
mobiai brain verify <id>           # Mark a note as checked against the current code

mobiai brain mcp                   # Starts an MCP server that exposes Brain as tools
mobiai brain install-mcp           # Registers the MCP server in AI clients (Claude Code, Cursor)
```

All accept `--root <path>` to point at a project other than the current directory.

### `mobiai brain init`

Creates the local structure:

```
<repo>/.mobiai/brain/
├── config.json
└── memories/
    ├── decisions/       ← one file per note
    ├── bugfixes/
    ├── testing/
    ├── knowledge/
    ├── decisions.md     ← old format: only the notice for mobiai < 0.3
    ├── bugfixes.md
    └── testing.md
```

It's idempotent: if `config.json` or any memory file already exists, it is not overwritten. The three `.md` files exist only so a teammate on an older mobiai sees a notice asking them to update (see [Upgrading from 0.2](#upgrading-from-02)) instead of an empty Brain.

To locate the project root, it walks the cwd's ancestors looking, in this order of priority, for:

1. `.mobiai/brain/config.json` (a brain already initialized).
2. `.git/`.
3. `settings.gradle.kts` / `settings.gradle` (Android / KMP).
4. `pubspec.yaml` (Flutter).
5. `Package.swift`.
6. `*.xcworkspace` / `*.xcodeproj`.
7. `Podfile`.
8. `package.json` with a `react-native` dependency.

If none match, it uses the current directory and warns.

### Upgrading from 0.2

Up to 0.2.x, every note of a type was appended to one file (`decisions.md`, `bugfixes.md`, …). Two branches that saved notes always conflicted at the end of that file. Since 0.3 each note is its own file under `memories/<section>/<id>.md`, so new notes never collide.

**There's nothing to run.** The first time any 0.3 command (or MCP tool) touches an old Brain, it:

1. creates one note file per old entry, keeping its id (entries written by hand without an id get one derived from their text);
2. adds a notice right below the heading of `decisions.md`, `bugfixes.md` and `testing.md`;
3. says so: `Brain migrated to the new format: N notes. Include .mobiai/brain/ in your next commit.`

The old entries stay where they were. The result depends only on the files' text, so two people migrating on two branches produce identical files and git merges them cleanly. It never commits. It also repairs entries whose body used `##` headings, which 0.2 split into empty entries.

**Teammates still on 0.2.x** only read the old files. Their `brain context` and `brain review` show the notice, so their agent tells them to update. A note they save is appended to the old file, and 0.3 turns it into a note file the next time it runs. A `promote` or `bump` done with 0.2.x on a migrated note changes only the old copy and is ignored: update first. Note that `brain review` on 0.2.x exits 1 because of the notice (it's shaped as an overdue temporary entry) unless you pass `--no-fail`.

### `mobiai brain scan`

Walks the project tree (max depth 6, ignoring `.git`, `node_modules`, `build`, `Pods`, `DerivedData`, `.dart_tool`, etc.) and produces `.mobiai/brain/scan.json` with:

- `project_type`: `android` | `ios` | `kmp` | `flutter` | `react_native` | `unknown`.
- `platforms`: `android`, `ios`, `shared` (whatever was detected).
- `build_systems`: `gradle`, `cocoapods`, `spm`, `npm`.
- `ui` / `di` / `network` / `persistence` / `serialization`: recognized libraries (Compose, SwiftUI, Koin, Hilt, Ktor, Retrofit, Room, DataStore, kotlinx.serialization, …).
- `testing`: JUnit, MockK, Mockito, Turbine, Espresso, …
- `integrations`: Firebase and other integrations detectable by strings in `build.gradle.kts` / `Podfile` / `package.json`.
- `ci_cd`: GitHub Actions, Bitrise, Codemagic, Fastlane.
- `warnings`: sensitive files detected (without reading their contents).

`scan` requires `init` to have been run first.

### `mobiai brain save <type>`

Four subcommands: `decision`, `bugfix`, `testing`, `knowledge`. Each writes the note as its own file, `memories/<section>/<id>.md`. All share the same flags:

| Flag | Required? | Notes |
|---|---|---|
| `--title <str>` | yes | Short title, used as H2 heading |
| `--platform <plat>` | no | `android` \| `ios` \| `shared` \| `kmp` \| `flutter` \| `react-native` |
| `--area <str>` | no | Free-form (`firebase_auth`, `dependency_injection`, `datastore`, …) |
| `--status <s>` | no | `active` (default) \| `temporary` \| `deprecated` |
| `--review-after YYYY-MM-DD` | no | Mostly meaningful with `temporary` |
| `--files a,b,c` | no | List of relevant paths. They tell which folders the note applies to, and the note is flagged when they change (see `verify`) |
| `--body <md>` | no | Markdown body. If omitted, reads from stdin (useful for piping multi-line content). `##` headings become `###` |
| `--supersedes <id>,…` | no | This note replaces those: they stop counting and their text is no longer delivered. The old notes are not modified |
| `--until <dep>@<version>,…` | no | The note holds while the project uses a lower version (e.g. `com.google.firebase:firebase-bom@34.0.0`). Once the project reaches it, the note is flagged for review |
| `--crash <signature>,…` | no | `bugfix` only: Crashlytics issue id, Sentry fingerprint or exception signature this fix solves (found later with `search --crash`) |
| `--new` | no | Save even if a note with a similar title exists (see below) |

**`knowledge`** stores how the app and the services it talks to behave: happy paths, server answers and error codes, SDK and third-party behavior. Save only what you can't learn by reading the code: "the server returns 409 when the order is closed" belongs here; "function X does Y" does not.

**Similar titles stop the save.** If the section already has an active note whose title shares 60% or more of its words, `save` writes nothing and answers `A similar note already exists: <id> "<title>". Edit that note file, pass --supersedes <id> to replace it, or pass --new to save a separate note.`

Example:

```bash
mobiai brain save decision \
  --title "Use Koin for DI" \
  --platform shared \
  --area dependency_injection \
  --files "composeApp/src/commonMain/di/Module.kt" \
  --body "### Decision
Use Koin as the DI framework for all shared code.

### Reason
KMP-friendly, no code generation, already integrated with composeApp/iosApp."
```

**Guard**: if `.mobiai/brain/config.json` doesn't exist, the command fails with a clear error and exit code 1. It will not create the brain by itself — it suggests running `mobiai brain init` first.

The internal type (`type:` in the rendered entry) is derived from the subcommand + status:

- `decision` → `architecture_decision`
- `bugfix` + `temporary` → `platform_workaround`
- `bugfix` + `active`/`deprecated` → `bug_fix`
- `testing` → `testing_pattern`
- `knowledge` → `knowledge`

Every entry gets a stable `id` based on the title slug + UTC timestamp, so two saves with the same title don't collide.

### `mobiai brain search <query>`

Case-insensitive search across the title and body of every memory entry. Returns results grouped by section, with title, status/platform, a snippet from the first line that contains the query, and the note's file.

```bash
mobiai brain search koin
mobiai brain search --platform ios firebase
mobiai brain search --status temporary
mobiai brain search --area firebase_auth keystore
mobiai brain search --crash 6a1b2c3d4e5f        # bugfixes for a crash signature
```

Accepts the same filters as `context` (`--platform`, `--status`, `--area`) with **AND** semantics. The query combines with the filters: `--platform ios firebase` returns only iOS entries mentioning "firebase". With `--crash`, the query is optional: it matches a bugfix's `crash` value exactly, or the text in a note's body.

### `mobiai brain show <id>`

Prints one note in full. When `context` lists a note in its index, agents read the note file directly; this is for clients that can't read files (MCP: `mobile_get`).

### `mobiai brain review`

Audits the Brain's **temporary debt**: walks the memories and shows every entry with `status: temporary` whose `review_after` has passed (`review_after <= today`). It's the command that closes the loop on "memory with an expiration date" — without it, temporary workarounds become permanent by inertia.

```bash
mobiai brain review                        # only expired, exit 1 if any
mobiai brain review --include-no-date      # + temporary entries without review_after
mobiai brain review --no-fail              # always exit 0 (informational mode)
```

**Example output:**

```
⚠ 2 expired temporary entry(ies):

bugfixes.md
  ⚠ Podfile composeApp lowercase until CocoaPods 1.16
    review_after: 2026-08-15 (expired 89 days ago)
    platform: ios
    id: podfile-composeapp-lowercase-20260512-094523

  ⚠ Suppress Deprecation in Compose 1.7
    review_after: 2026-10-01 (expired 42 days ago)
    platform: android
    id: suppress-deprecation-compose-20260801-101200
```

**Exit codes:**

| Case | Default | With `--no-fail` |
|---|---|---|
| Expired entries present | exit 1 | exit 0 |
| No expired entries | exit 0 | exit 0 |
| Brain not initialized / I/O error | exit 1 (stderr) | exit 1 |

The default exit 1 is meant to be used as a **CI or pre-commit gate**: the build fails if someone is piling up temporary workarounds without review. If you only want to inform without blocking, use `--no-fail`.

**What's in and what's out:**

| `status` | `review_after` | Listed? |
|---|---|---|
| `temporary` | past date or today | **Yes** (expired) |
| `temporary` | future date | No |
| `temporary` | no date | Only with `--include-no-date` (separate section) |
| `temporary` | malformed | Silently ignored (fix the `.md`) |
| `active` | any | No (not debt) |
| `deprecated` | any | No (already closed) |

`review` **does not edit or delete** anything — it only reports. To drop an entry from the list, edit the `.md` and change `review_after`, set `status` to `active`/`deprecated`, or remove the block.

**Notes that may be out of date.** `review` also lists, in their own section, the notes flagged with ⚠ (see `verify`): their files changed or disappeared, they haven't been verified in 3+ months, or an `until` condition was met. They never change the exit code, so CI gates keep failing only on overdue temporary entries.

### `mobiai brain verify <id>`

A note that relies on code can go stale without anyone noticing. Brain checks it against git every time it reads the notes, with a single `git log` for the whole Brain, and flags a note with ⚠ when:

- one of its `### Files` has commits on a later day than its `verified` date (`⚠ LoginViewModel.kt changed since verification`), or no longer exists;
- it has no files and wasn't verified in the last 3 months (`⚠ not verified in 3+ months`);
- an `until` condition is met (`⚠ firebase-bom is now 34.1.0 (note valid until 34.0.0): check whether this workaround is still needed`) or that dependency is gone. Versions are read from Gradle catalogs and build files, `Podfile.lock`, `Package.resolved`, `pubspec.lock` and `package-lock.json`/`package.json`.

Nothing expires or gets deleted. When you (or the agent) check a flagged note and it still holds:

```bash
mobiai brain verify <id>                          # verified = today
mobiai brain verify <id> --files a/New.kt,b/X.kt  # also replace its files
```

If it no longer holds, fix the note or `promote --status deprecated` it. An `until` condition that can't be interpreted never flags anything; `mobiai doctor` lists those.

### `mobiai brain promote <id>`

Changes the `status` of an existing entry. Designed as the close-out flow after `brain review`: when a `temporary` entry shouldn't be temporary any more, you promote it to `active` (it became permanent) or `deprecated` (no longer applies).

```bash
mobiai brain promote <id> --status active
mobiai brain promote <id> --status deprecated
mobiai brain promote <id> --status temporary --review-after 2027-01-01

# Promote and clear review_after in a single call
mobiai brain promote <id> --status active --clear-review-after
```

| Flag | Required? | Notes |
|---|---|---|
| `--status <s>` | yes | `active` \| `temporary` \| `deprecated` |
| `--review-after YYYY-MM-DD` | no | Also updates `review_after` |
| `--clear-review-after` | no | Removes `review_after` (mutually exclusive with `--review-after`) |

For `bugfix` entries, the `type:` field is recomputed automatically from the new `status` (`temporary` → `platform_workaround`, `active`/`deprecated` → `bug_fix`), preserving the invariant `save` establishes.

The body, files under `### Files`, and any custom metadata are preserved byte-perfect — `promote` only touches the fields you asked for.

### `mobiai brain bump <id>`

Extends only the `review_after` of an entry, without touching `status`. Useful when a temporary workaround is still valid and you want to push the review deadline:

```bash
mobiai brain bump <id> --review-after 2027-01-01
```

| Flag | Required? | Notes |
|---|---|---|
| `--review-after YYYY-MM-DD` | yes | New date |

For status changes, use `promote`. To clear the date completely, use `promote --clear-review-after`.

### `mobiai brain context`

Reads `config.json`, `scan.json` (if it exists), and all the memories. Prints agent-ready Markdown:

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
- sensitive file detected: .env (not read)
```

**Up to 25 KB, everything; above that, an index.** The full document is built and measured. If it's over `context_max_bytes` (25600 by default, set in `config.json`), the agent gets the stack, rules and warnings, the full text of the active decisions (newest first, while they fit), and one line per every other note:

```
- [bug_fix] Login crash on screen rotation · android · auth · created 2026-09-14 (3 months ago) · verified 2026-09-14 · ⚠ LoginViewModel.kt changed since verification · applies to app/src/main/java/com/example/auth/ → memories/bugfixes/login-crash-on-screen-rotation-20260914-163044.md
```

The agent opens the note files it needs (or uses `show` / `mobile_get`). Filters apply before measuring, so a narrow query still gets full texts.

**Marks you'll see** (in full mode as `- state:` / `- check:` lines, in the index inline):

- `unconfirmed`: the note was written by someone else and its file hasn't reached the main branch yet, so nobody reviewed it in a PR. Your own notes (not committed yet, or committed with your `git config user.email`) count as confirmed for you. An unconfirmed decision is not treated as a project rule. The main branch is `origin/HEAD` (e.g. `develop` in git-flow teams), then `origin/main`, `origin/master`, `main`, `master`; force one with `"confirmed_branch"` in `config.json`. Without git, everything counts as confirmed.
- `⚠ …`: the note may be out of date (see `verify`).
- `superseded by <id>`: a newer note replaces it; its text is no longer delivered. If the replacement is itself unconfirmed, the old note stays in force and shows `replacement proposed by <id> (unconfirmed)`.

**Available filters**:

| Flag | Notes |
|---|---|
| `--section` | Comma-separated or repeatable. Canonical names: `stack`, `rules`, `decisions`, `bugfixes`, `testing`, `knowledge`, `releases`, `warnings` (`integrations` is an alias of `knowledge`). |
| `--platform` | Filter entries by `platform:` (exact, case-insensitive). |
| `--status` | Filter by `status:` (exact). |
| `--area` | Filter by `area:` (substring). |

Filters apply only to memory entries — `stack`, `rules`, and `warnings` are included/excluded solely via `--section`. If a section has no entries after filtering, it shows `_No entries match the current filter._` (different from `_No entries yet._` for genuinely empty sections).

```bash
# Only temporary iOS bugfixes
mobiai brain context --section bugfixes --platform ios --status temporary

# Only the detected stack, no memories
mobiai brain context --section stack

# Decisions + bugfixes on the shared platform
mobiai brain context --section decisions,bugfixes --platform shared
```

---

## End-to-end examples

Three full stories that make the value of Brain tangible.

### 1. Temporary Firebase workaround on iOS

**Monday 9:00** — Your iOS build blows up after upgrading Firebase to 11.x. Forty minutes later you discover that `composeApp` has to be lowercase in the `Podfile` until CocoaPods 1.16 ships with the fix.

**Monday 9:45** — You fix the bug with `mobiai-fix-issue`. At the end of the flow, the skill detects you have `.mobiai/brain/config.json` and proposes:

> *I detected this was more of a workaround than a root fix. Should I save it in Brain as a temporary bugfix, to revisit in three months?*

You confirm. The skill runs:

```bash
mobiai brain save bugfix \
  --title "Podfile composeApp lowercase until CocoaPods 1.16" \
  --platform ios \
  --area cocoapods \
  --status temporary \
  --review-after 2026-08-15 \
  --files "iosApp/Podfile" \
  --body "..."
```

**Three months later** — Another teammate (or future-you) sees a related crash. They ask the agent, which invokes `mobile_search "cocoapods composeApp"` and finds the entry with full context: when it was decided, which file, when to review. Five minutes of investigation instead of forty.

### 2. DI decision in a KMP project

**Architecture brainstorm** — You need to pick a DI framework for your new KMP app. You launch `mobiai-mobile-brainstorming`, compare Koin vs Hilt+Dagger vs manual, reach a conclusion.

After the User Review Gate, the skill proposes saving the decision. You confirm and it's stored as an `architecture_decision` in `memories/decisions/` with the reasoning, rejected trade-offs, and files where it lives.

**Two months later** — A new dev joins, opens the repo in Cursor. They ask "why Koin and not Hilt?". The agent invokes `mobile_search "DI"` and answers with the full decision, date, and reasons. No need to reopen the debate.

### 3. Non-obvious DataStore testing pattern

**Flaky tests** — Your DataStore test passes locally but fails in CI 1 in 4 runs. You investigate with `mobiai-mobile-debugging`: turns out `dataStore.edit { it.clear() }` doesn't flush synchronously, leaving a window where `.data.first()` returns the previous value.

You learn the correct pattern is to wait on `dataStore.data.first { it.asMap().isEmpty() }` after the clear. You apply the fix with `mobiai-write-tests`.

At the end, the skill proposes:

> *This pattern looks reusable for other DataStore tests. Should I save it as a testing pattern?*

You confirm. It's stored in `memories/testing/`. The next time someone (or you) writes a test against DataStore, `mobile_context --section testing` will surface it before they hit the same flake.

---

## Internal architecture

### On-disk layout

Everything Brain knows lives under `<repo>/.mobiai/brain/`:

```
.mobiai/brain/
├── config.json          ← project metadata + custom rules
├── scan.json            ← detected stack (regenerable with `scan`)
└── memories/
    ├── decisions/<id>.md   ← architecture_decision
    ├── bugfixes/<id>.md    ← bug_fix + platform_workaround
    ├── testing/<id>.md     ← testing_pattern
    ├── knowledge/<id>.md   ← knowledge
    ├── decisions.md        ← old format (0.2): old entries + notice
    ├── bugfixes.md         ← idem
    ├── testing.md          ← idem
    ├── integrations.md     ← old format: hand-written notes migrate to knowledge/
    └── releases.md         ← reserved
```

Only `config.json` is mandatory. The rest is created during `init` and fills in with use. There's no index file in the repo: the CLI builds it when it reads the notes, so it can't cause merge conflicts. Notes are ordered by `date`, then `id`.

### Entry format

Memories are plain Markdown, readable by eye. Each note file holds one H2 (`##`) title followed by YAML-ish metadata and a free-form body:

```markdown
## Podfile composeApp lowercase until CocoaPods 1.16

- id: podfile-composeapp-lowercase-20260512-094523
- type: platform_workaround
- status: temporary
- platform: ios
- area: cocoapods
- date: 2026-05-12
- verified: 2026-05-12
- review_after: 2026-08-15

### Reason
CocoaPods 1.15.x breaks with CamelCase modules when ...

### Files
- iosApp/Podfile
```

Only the first `##` line is the title: a `##` inside the body stays in the body. Optional fields: `verified` (starts equal to `date`; set by `verify`), `supersedes`, `until`, `crash`, and `paths` (comma-separated globs such as `feature/**/checkout/*.kt` that replace the folders taken from `### Files`, for the per-file hook). Notes with `unconfirmed`, `⚠` or `superseded by` marks keep their files untouched: those marks are computed when the notes are read.

### Stable IDs

Every entry gets an ID generated at save time: `<title-slug>-<YYYYMMDD-HHMMSS>` in UTC. For example:

```
use-koin-for-di-20260301-143025
```

This gives us two things:

- **Dedup**: two saves with the same title don't collide (different timestamps).
- **Reference**: if you need to link to an entry from another, the ID is stable even if you edit the title.

### Write atomicity

Writes to `config.json`, `scan.json`, and the note files use the **temp + rename** pattern: we write first to a temp file in the same directory, then atomically `rename` to the destination. If the process dies mid-write, you never end up with a corrupted file.

### Scanner: scope and limits

`scan` walks the tree with these parameters:

- **Max depth**: 6 levels from the root. Covers typical monorepos without getting lost in `node_modules` or caches.
- **Max file size**: 1 MiB. Skips binaries and generated artifacts.
- **Ignored dirs**: `.git`, `.mobiai`, `.gradle`, `.idea`, `.kotlin`, `.dart_tool`, `.swiftpm`, `build`, `out`, `target`, `node_modules`, `Pods`, `DerivedData`, `vendor`, `.claude`, `.cursor`, `.copilot`, `.gemini`, `.codex`, `.junie`.
- **Sensitive files**: `.env`, `google-services.json`, `GoogleService-Info.plist`, `local.properties`, `keystore.properties` are **detected** (warning) but **never read**.

Library detection runs by case-insensitive substring search over config files (`build.gradle.kts`, `Podfile`, `package.json`, `pubspec.yaml`…). No AST parsing — deliberately simple, maintainable, fast.

### Why Markdown and not SQLite?

At the realistic scale of 1–100 entries per project (today's range), Markdown wins on every dimension that matters:

- **Eye-readable**: you can open `decisions.md` in any editor and understand the contents.
- **Diff-friendly**: every save is a new file, easy to review in a PR, and two branches that save notes never conflict.
- **Zero runtime**: no database required, no schema migrations.
- **Hand-editable**: if you need to correct something, edit the `.md` directly.

If a project ever crosses ~500 entries and search slows down, we'll migrate to SQLite + FTS5 while keeping the same CLI and export format. For now, there's no reason.

---

## Integration with Skills

The same four skills have **two contact points** with Brain — one on entry, one on exit.

### Exit hook: propose `save`

The skills propose saving at the end of their flow, all with the same pattern: **only if** `.mobiai/brain/config.json` exists in the project, and **always** asking the user for approval before invoking `save`.

| Skill | When it proposes | Type saved |
|---|---|---|
| `mobiai-fix-issue` | After Phase 6 (verification), before the PR gate | `bugfix` (status depending on real fix vs workaround) |
| `mobiai-write-tests` | After Step 4 (tests pass), if the test captures a non-obvious pattern | `testing` |
| `mobiai-mobile-debugging` | After Phase 6 (root cause confirmed), **only on standalone invocation** — if invoked from fix-issue, it's skipped (fix-issue already has its own hook) | `bugfix` |
| `mobiai-mobile-brainstorming` | After the User Review Gate (spec approved), one entry per non-trivial architectural decision in the spec | `decision` |

If the project has no brain, the hook silently skips. The agent never invokes `save` on its own: it always asks for a one-line confirmation from the user before running it.

### Entry hook: pre-flight `review`

The same four skills, on **startup**, invoke `mobile_review` (CLI fallback: `mobiai brain review --no-fail`), filter the expired entries by `platform` and `area` of the current work, and surface to the user any that match before starting:

> "Heads up — there are N overdue workaround(s) in `<platform>/<area>` that may be relevant: `<title>`. I'll proceed; let me know if you want to look at those first."

It doesn't pause the flow. The user decides whether to detour into `brain promote` / `brain bump` first or continue and revisit later. `mobile-debugging` and `write-tests` skip the pre-flight when invoked nested from `fix-issue` (which already did its own) to avoid double-prompting.

**Dual MCP / CLI path**: every hook documents two equivalent routes. If the client has `mobiai-brain` registered as an MCP server, the agent invokes the `mobile_save_*` tool directly — faster, no process spawn, structured output. Otherwise it falls back to the same `mobiai brain save ...`. Both end up writing the exact same note file.

### Per-file delivery: the `PostToolUse` hook

The `core` plugin also registers a hook that runs after the agent reads or edits a file (Claude Code: `PostToolUse` on `Read|Edit|Write`; Cursor: `postToolUse` on `Read|Write`). It calls `mobiai brain hook file-context`, which hands the agent, next to the tool result, the notes whose folders contain that file (taken from `### Files`, or from `paths:`). Each note is delivered at most once per session, up to about 4 KB per call; notes that don't fit are listed by id. It never writes the Brain, and it stays silent when there's no Brain or mobiai is older than 0.3. Other tools rely on the `applies to` folders in `context`'s index.

Other skills that could integrate in the future: `mobiai-crashlytics` (recurrent crashes → `save bugfix`), `mobiai-mobile-planning` (plans with embedded decisions → `save decision`).

---

## FAQ and troubleshooting

### Should I commit `.mobiai/brain/` to git?

**Yes, that's recommended.** Brain is team knowledge: decisions, workarounds, and patterns you want to share with the rest of the team (humans and AI agents alike). Versioning `.mobiai/brain/` in git makes that knowledge travel with the repo.

The only reasonable exception is an experimental project where only you use Brain as a personal notebook — there you could add it to `.gitignore`.

### Is Brain synced across machines?

There's no native sync. Sync happens through git: if you commit it, your teammates have it after `git pull`. No central server, no account, no cloud. Simple by design.

### What happens if I edit a note file by hand?

It's allowed and supported. The parser re-reads the files on every `context` / `search`, so any manual edit is reflected immediately. The only restrictions are:

- Keep the `## Title` + metadata list + body structure.
- Keep the file name equal to the `id:` (one note per file).
- If you add custom metadata, we don't guarantee it for future filters.

### How do I delete or modify an entry?

**To change status** (close a workaround, deprecate a decision): use `mobiai brain promote <id> --status deprecated` (or `active`).

**To extend the review deadline** of a still-valid workaround: use `mobiai brain bump <id> --review-after YYYY-MM-DD`.

**To replace a note** with a newer one: save the new one with `--supersedes <id>`.

**For actual deletion**: delete the note file. There's no `mobiai brain delete` by design — we prefer marking entries as `deprecated` to preserve history. Careful with notes that came from the 0.2 format: their old copy still lives in `decisions.md`/`bugfixes.md`/`testing.md`, so deleting the note file brings it back on the next run. Deprecate those, or also remove their block from the old file (for example, sensitive data saved by accident).

### What does `unconfirmed` mean?

The note was written by someone else and hasn't reached the main branch yet, so nobody reviewed it in a PR. The agent uses it with caution instead of as a project rule. It becomes confirmed by itself once the PR that adds it is merged (after a `git fetch`). If a note of yours shows as unconfirmed, your `git config user.email` probably differs from the email of the commit that added it.

### Why doesn't the scanner detect my library X?

Detection works via a hardcoded list of known strings (Koin, Ktor, Compose, etc.). If your library doesn't show up, open an issue at [MobiAI-Core](https://github.com/ArisGuimera/MobiAI-Core/issues) with the package name and a snippet of the config file where it appears. Adding it is trivial — see `cli/internal/brain/detect_deps.go`.

### Does it work offline?

Completely. Everything is local: Markdown + JSON on disk, no network calls. The MCP server is also local (stdio). This is by design: your architecture decisions are yours.

### Can my secrets end up in Brain?

No. The scanner detects sensitive files (`.env`, `google-services.json`, `GoogleService-Info.plist`, `local.properties`, `keystore.properties`) but **never reads their content**. It only records them as `warnings` with the name, so you know they exist.

`save` entries only store what you (or the agent with your approval) pass in explicitly via `--body` and `--files`. If you never pass it a secret, none get saved.

### Which MCP client is recommended?

Any of the five supported (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI) works equally well — they all speak the same MCP protocol. Use whichever you already use to code. Step-by-step setup in [`MCP-SETUP.md`](MCP-SETUP.md).

### Can I have more than one Brain per repo?

No. Root detection finds the first `.mobiai/brain/config.json` walking up from the cwd, so only the closest one wins. If you need to separate subproject knowledge, use the `--area` or `--platform` flags on saves to segment.

### What about monorepos?

Every monorepo has **a single Brain** at the root. The way to segment knowledge across subprojects is via `--area` (for example `--area app-android` vs `--area app-ios` vs `--area shared-kmp`). The `context` and `search` filters let you retrieve only what's relevant to a zone of the monorepo.

### I'm getting "brain not initialized" but I ran `init`

You're probably in a subdirectory and root detection isn't finding the expected `config.json`. Fixes:

- Make sure you ran `mobiai brain init` from the project root (where `.git/` or `settings.gradle.kts` or `Package.swift` lives).
- Or use `--root <absolute-path>` to point at the project explicitly.

### `scan` detects weird things / nothing at all

`scan` has a max depth of 6 from the root. If your monorepo has modules at depth 7+, it won't see them. For extreme cases, consider running `scan` from the module subdirectory with `--root` pointing there (though it will only reflect in that brain's `scan.json`).

---

## Roadmap

**Phase 1** — `init`, `scan`, `context`. ✓
**Phase 2** — `save decision|bugfix|testing` + hook in `mobiai-fix-issue`. ✓
**Phase 3** — `search` + filters (`--section`, `--platform`, `--status`, `--area`) in `context`. ✓
**Phase 4** — save hooks in `mobiai-write-tests` (testing pattern), `mobiai-mobile-debugging` (bugfix without going through fix-issue), and `mobiai-mobile-brainstorming` (decision after spec approved). ✓
**Phase 5** — MCP server (`mobiai brain mcp`) that exposes the 6 operations as native tools for Claude Code, Cursor, Copilot CLI, Codex, and Gemini CLI. Brain is now always in the agent's toolbox. Setup in [`MCP-SETUP.md`](MCP-SETUP.md). ✓
**Phase 6** — `mobiai brain review` (CLI) + `mobile_review` (MCP tool) to audit `status: temporary` entries whose `review_after` has passed. Closes the loop on "memory with an expiration date" — temporary workarounds no longer become permanent by inertia. ✓
**Phase 7** — `mobiai brain promote` and `mobiai brain bump` (CLI) + `mobile_promote` and `mobile_bump` (MCP tools) to close the entry lifecycle from CLI/agent: change `status` or extend `review_after` without editing the `.md` by hand. Atomic rewrite; preserves body, files, and custom metadata byte-perfect. ✓
**Phase 8** — the 4 skill hooks (`mobiai-fix-issue`, `mobiai-write-tests`, `mobiai-mobile-debugging`, `mobiai-mobile-brainstorming`) now document the equivalent MCP tool as the preferred route, keeping the CLI as the explicit fallback. When the client has `mobiai-brain` registered as an MCP server, the agent invokes `mobile_save_*` directly; otherwise it falls back to the usual `mobiai brain save ...`. ✓
**Phase 9** — automatic pre-flight in the 4 skills: before starting work, the agent invokes `mobile_review` (or `mobiai brain review --no-fail`), filters by platform/area of the current work, and surfaces any relevant expired workarounds. It doesn't pause: the user decides whether to look at them first. Closes the "memory with an expiration date" loop by making debt **proactively visible** at the exact moment it matters. ✓
**Phase 10** — `mobiai brain install-mcp` registers the MCP server in supported clients (Claude Code and Cursor in v1) with a single command. It auto-detects which clients the user has installed, preserves the rest of the config, is idempotent and reversible (`--uninstall`). Absolute path to the binary via `os.Executable()` so the AI client finds it even with a different `$PATH` than the shell. We eliminate the friction of editing JSON by hand. ✓
**Brain 0.3** — one file per note (no more merge conflicts) with automatic migration and a notice for older versions; `context` switches to an index above 25 KB; `knowledge` type; `unconfirmed` notes until they reach the main branch; verification against git (`verify`, ⚠ marks); per-file delivery hook; `supersedes` and a brake for similar titles; `until` conditions tied to dependency versions; `crash` signatures and `search --crash`. ✓

**Next steps**:

- **`install-mcp` for Copilot CLI, Codex, and Gemini CLI** — v1 covers Claude Code and Cursor (same JSON shape). The other three have different formats (Copilot/Gemini JSON with their own shape, Codex TOML) and will be added when there's demand.

- **`brain review --upcoming N`** — in addition to the expired ones, list temporary entries that expire in the next N days (proposed default: 30). Gives planning margin for sprints/retros without waiting until something is already overdue. Pending.
- **`save release`** (lower-volume category, low priority).
- **`mobiai brain lint`**: validate note files and look for suspicious instructions, for CI.

**Future phase** — migration to SQLite + FTS5 if Markdown falls short (>~500 entries per project). For the current range (1–100 entries), filters over Markdown are enough.
