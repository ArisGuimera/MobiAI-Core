---
name: mobiai-brain
description: "Use when the user asks about past decisions, workarounds, bugfixes, testing patterns or integrations specific to *this* project, OR before proposing non-trivial architecture/integration changes. MobiAI Brain is per-project living memory at <repo>/.mobiai/brain/. Loads context the user already captured so suggestions respect real project conventions instead of generic best practices."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# MobiAI Brain

Per-project living memory for mobile projects. Lives at `<repo>/.mobiai/brain/` — separate from MobiAI's global state in `~/.mobiai/`. Each project has its own brain; nothing is shared between projects.

## When to invoke this skill

- User asks "what did we decide about X?" / "why is Y like that in this project?" / "is this workaround still needed?"
- You are about to propose changes to architecture, dependency injection, networking, persistence, testing strategy, or third-party integrations.
- User mentions an integration that often has project-specific quirks (Firebase, push, payments, auth).
- User starts a fresh session in a mobile project and you need quick orientation.

If the project has no `.mobiai/brain/` yet and the user wants project memory, suggest running `mobiai brain init`. Don't run it yourself unless asked.

## Phase 1 commands available today

```
mobiai brain init      # Create .mobiai/brain/ in the current project (idempotent)
mobiai brain scan      # Detect stack: Android / iOS / KMP / Flutter / RN + libraries
mobiai brain context   # Print Markdown context: config + scan + memories
```

`save` and `search` subcommands are coming in Phase 2. Until then, the user (or you, with their consent) appends entries to `.mobiai/brain/memories/*.md` by hand.

## Recommended flow

1. **Read context first.** Before proposing arch/integration/testing changes, run `mobiai brain context` (or read `.mobiai/brain/memories/*.md` directly) and respect what's there.
2. **Honor `status: active` decisions.** Treat them as project rules.
3. **Treat `status: temporary` workarounds as provisional.** They are not permanent decisions; do not promote them to `CLAUDE.md`.
4. **Don't conflate Brain with CLAUDE.md.**
   - `CLAUDE.md` holds *stable* rules: architecture, commands, conventions.
   - Brain holds *historical and changing knowledge*: decisions with dates, bugfixes, workarounds with review dates, testing patterns discovered along the way.
5. **Don't conflate Brain with skills.**
   - Skills teach the agent *how* to work (methodology).
   - Brain stores *what* the project actually did and decided.
6. **Never store secrets.** The scanner already flags sensitive files (`.env`, `GoogleService-Info.plist`, `google-services.json`) and refuses to read their contents. If you're about to write a memory, do the same: paths and behavior, not credentials.

## What the Brain looks like on disk

```
<repo>/.mobiai/brain/
├── config.json           # version, project_name, project_type, platforms, rules
├── scan.json             # last scan: stack, integrations, CI/CD, warnings
└── memories/
    ├── decisions.md
    ├── bugfixes.md
    ├── testing.md
    ├── integrations.md
    └── releases.md
```

Memory entries are Markdown sections. Recommended shape (also what Phase 2 `save` will produce):

```
## <Title>

- id: <slug>
- type: architecture_decision | platform_workaround | testing_pattern | integration_note
- status: active | temporary | deprecated
- platform: android | ios | shared | flutter | react-native
- area: <free-form>
- date: <ISO date>
- review_after: <ISO date>   (optional, for temporary entries)

### Decision / Problem / Pattern
...

### Reason / Root Cause / Solution
...

### Files
- path/to/file
```

## Behavior contract

- Brain is **per-project**. Don't mix decisions from different repos.
- Brain is **complementary** to `CLAUDE.md`/`AGENTS.md`/`GEMINI.md`, not a replacement.
- The agent **must not** silently overwrite a memory entry. Append; if changing status, leave a note explaining why.
- If the scan in `scan.json` is older than the user's last large refactor, suggest re-running `mobiai brain scan` before relying on it.
