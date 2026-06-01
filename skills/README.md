# MobiAI Skills

> **Read in another language:** **English** · [Español](README.es.md)

> The catalog of **skills** that the `mobiai` CLI distributes to your assistant (Claude Code, Cursor, Copilot CLI, Codex, Gemini CLI). Each skill is expert context the AI loads on demand to fix bugs, write tests, analyze crashes, create PRs, and much more — following each platform's best practices.

**Table of contents**

- [How they're organized](#how-theyre-organized)
- [Available packs](#available-packs)
- [Installation](#installation)
- [Skill catalog](#skill-catalog)
- [How the assistant picks which skill to use](#how-the-assistant-picks-which-skill-to-use)
- [Contribute a new skill](#contribute-a-new-skill)
- [Third-party skills (Google)](#third-party-skills-google)

---

## How they're organized

Skills are grouped into **packs**. A pack is an installable unit and, except for `core` and `mobile`, maps to a platform.

```
skills/
├── core/              # cross-platform: flow, process, ecosystem, bootstrap
│   ├── hooks/         # SessionStart hook (update notifier)
│   └── skills/        # 22 cross-platform skills (using-mobiai, fix-issue, …)
├── android/
│   └── skills/
│       ├── google/    # official Google skills (Apache 2.0, weekly sync)
│       └── mobiai-android-*
├── ios/
├── kmp/               # depends on android + ios
├── flutter/           # depends on android + ios
├── react-native/      # depends on android + ios
├── community/         # community-contributed skills — open to everyone
└── mobile/            # meta-pack: brings everything above
```

Each pack contains a `.claude-plugin/plugin.json` with metadata and dependencies. Dependencies are resolved automatically on install (for example, installing `kmp` pulls in `core`, `android`, and `ios`).

## Available packs

| Pack | Includes | Depends on | Command |
|---|---|---|---|
| `mobile` | The whole stack (meta-pack) | core, android, ios, kmp, flutter, react-native | `mobiai skills add mobile` |
| `android` | Android skills + official Google skills | core | `mobiai skills add android` |
| `ios` | iOS skills | core | `mobiai skills add ios` |
| `kmp` | Kotlin Multiplatform skills | core, android, ios | `mobiai skills add kmp` |
| `flutter` | Flutter / Dart skills | core, android, ios | `mobiai skills add flutter` |
| `react-native` | React Native skills | core, android, ios | `mobiai skills add react-native` |
| `community` | Skills contributed by the community — the one pack open to everyone | core | `mobiai skills add community` |
| `core` | ⚠️ Internal — cross-platform skills, `using-mobiai` bootstrap, and `SessionStart` hook. Not installed standalone | — | — |

> `core` is installed automatically as a dependency of any other pack. Don't pick it manually.

## Installation

Packs are managed with `mobiai skills`. Full flow details (supported hosts, troubleshooting, local builds) in [../cli/README.md](../cli/README.md).

```bash
# Interactive selector (recommended for first-time use)
mobiai skills init

# Install specific packs
mobiai skills add mobile
mobiai skills add android ios

# List what's installed
mobiai skills list

# Uninstall
mobiai skills remove flutter
```

## Skill catalog

Skills are grouped by their role in the workflow. The live catalog, with the guide for when to invoke each one, lives in [core/skills/using-mobiai/SKILL.md](core/skills/using-mobiai/SKILL.md) (it's loaded automatically at the start of every session).

### Workflow (in `core`)

Orchestrator skills covering end-to-end cycles:

- `mobiai-fix-issue` — from ticket to PR (Jira/Linear/GitHub issues)
- `mobiai-reproduce-bug` — reproduce bugs on device/emulator/simulator
- `mobiai-analyze-crash` — analyze crashes from a stack trace, log, or screenshot
- `mobiai-crashlytics` — investigate Firebase Crashlytics issues
- `mobiai-write-tests` — add tests (unit, UI, regression)
- `mobiai-review-code` — code review for mobile changes
- `mobiai-create-pr` — package work into a well-structured PR

### Process (in `core`)

Skills that define *how* to approach the work:

- `mobiai-mobile-brainstorming` — explore intent and requirements before writing code
- `mobiai-mobile-debugging` — root-cause analysis as evidence gathering
- `mobiai-mobile-tdd` — tests before implementation
- `mobiai-mobile-planning` — turn a spec into a step-by-step plan
- `mobiai-mobile-verification` — mandatory gate before declaring "done"
- `mobiai-mobile-executing-plans` — execute a written plan
- `mobiai-mobile-executing-plans-with-subagents` — execution via subagents with two-stage review
- `mobiai-mobile-parallel-agents` — orchestrate independent agents
- `mobiai-mobile-worktrees` — isolate work in worktrees
- `mobiai-mobile-finishing-branch` — close/merge/clean up the branch
- `mobiai-writing-skills` — create new skills following the MobiAI format

### CLI and ecosystem (in `core`)

- `using-mobiai` — catalog bootstrap (loaded at session start)
- `mobiai-graph` — AI ↔ `mobiai graph` interface (semantic code search)
- `mobiai-brain` — AI ↔ `mobiai brain` interface (per-project memory)
- `mobiai-update` — `/mobiai-update` command to refresh the catalog

### Platform — Android (in `android`)

- `mobiai-android-device` — `adb`, screenshots, logcat, UI automation
- `mobiai-android-build` — Gradle, flavors, variants, troubleshooting
- `mobiai-android-testing` — Android testing frameworks
- `mobiai-android-architecture` — Android project structure
- Plus the [official skills maintained by Google](https://github.com/android/skills) (Apache 2.0, vendored with weekly sync). See [../NOTICE.md](../NOTICE.md).

### Platform — iOS (in `ios`)

- `mobiai-ios-device` — `simctl`, simulator, logs
- `mobiai-ios-build` — `xcodebuild`, schemes, SPM/CocoaPods
- `mobiai-ios-testing` — XCTest, snapshot tests, UI tests
- `mobiai-ios-architecture` — iOS project structure

### Multiplatform

Each in its own pack:

- `mobiai-kmp` — Kotlin Multiplatform (in `kmp`)
- `mobiai-flutter` — Flutter / Dart (in `flutter`)
- `mobiai-react-native` — React Native (in `react-native`)

## How the assistant picks which skill to use

The `using-mobiai` bootstrap loads at the start of every session and contains the decision guide: given a user intent ("reproduce this bug", "open a PR", "review this code"), the assistant identifies the right skill and invokes it before touching code.

The idea is that the assistant doesn't improvise: if there's a skill covering the task, it uses it. If there's no match, it asks you before acting.

## Contribute a new skill

1. Read the [skill-authoring guide](../docs/crear-skills.md).
2. Invoke the `mobiai-writing-skills` skill from your assistant — it walks you through the structure, frontmatter, and actionable instructions.
3. Place the skill under `community/skills/` — that's the pack open to everyone. Register it in [`community/README.md`](community/README.md).
   - The platform packs (`core`, `android`, `ios`, …) are maintainer-curated; a CI guard (`guard-community-skills`) blocks non-maintainer PRs that touch them. If your skill belongs in one, propose it via an issue.
4. Open a PR. CI blocks PRs that add or modify skills without updating the relevant documentation.

General contribution guidance in [../CONTRIBUTING.md](../CONTRIBUTING.md).

## Third-party skills (Google)

`skills/android/skills/google/` contains the [official Android skills maintained by Google](https://github.com/android/skills) (Apache 2.0). They're auto-synced weekly via CI. Don't edit them by hand: your changes would be overwritten in the next sync. If you find a problem, open the upstream issue in Google's repository.

Full attribution and terms in [../NOTICE.md](../NOTICE.md).
