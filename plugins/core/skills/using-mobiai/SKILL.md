---
name: using-mobiai
description: Bootstrap skill loaded on every session — teaches the agent about MobiAI skills, agents, and capabilities
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
---

# MobiAI — Mobile Development Ecosystem

You have access to MobiAI, an ecosystem of skills, agents, and automation tools for mobile development. MobiAI gives you expert-level knowledge for Android, iOS, KMP, Flutter, and React Native projects.

## Quick Decision Guide

Start here to find the right skill:

- **User reports a bug / shares a ticket** → `mobiai-fix-issue` (full pipeline)
- **User shares a crash, stack trace, or error** → `mobiai-analyze-crash`
- **User mentions Firebase Crashlytics specifically** → `mobiai-crashlytics`
- **User wants to reproduce a bug on device** → `mobiai-reproduce-bug`
- **User asks to write or add tests** → `mobiai-write-tests`
- **User asks for code review** → `mobiai-review-code`
- **User wants to create a PR** → `mobiai-create-pr`
- **Need to interact with Android emulator/device** → `mobiai-android-device`
- **Need to build an Android project** → `mobiai-android-build`
- **Need to interact with iOS Simulator** → `mobiai-ios-device`
- **Need to build an iOS project** → `mobiai-ios-build`
- **Working on a KMP project** → `mobiai-kmp`
- **Working on a Flutter project** → `mobiai-flutter`
- **Working on a React Native project** → `mobiai-react-native`
- **Need to understand Android project structure** → `mobiai-android-architecture`
- **Need to understand iOS project structure** → `mobiai-ios-architecture`
- **Writing Android tests specifically** → `mobiai-android-testing`
- **Writing iOS tests specifically** → `mobiai-ios-testing`
- **User wants to design a new feature before coding** → `mobiai-mobile-brainstorming`
- **Bug or unexpected behavior, need to find root cause** → `mobiai-mobile-debugging`
- **Implementing a feature or fix with tests first** → `mobiai-mobile-tdd`
- **Planning a multi-step feature** → `mobiai-mobile-planning`
- **About to claim work is done** → `mobiai-mobile-verification`
- **Have a plan to execute step by step** → `mobiai-mobile-executing-plans`
- **Multiple independent tasks to parallelize** → `mobiai-mobile-parallel-agents`
- **Starting feature work in isolation** → `mobiai-mobile-worktrees`
- **Implementation done, need to integrate** → `mobiai-mobile-finishing-branch`
- **Creating a new MobiAI skill** → `mobiai-writing-skills`

## MANDATORY: Invoke, Don't Paraphrase

Skills must be **invoked via the `Skill` tool** (or your platform's equivalent). It is NOT enough to read the description in this bootstrap, remember the skill's steps, and reproduce them from memory — that is the single most common failure of this ecosystem.

Concretely, this means:

- When a task matches an entry in the Quick Decision Guide above, your **first tool call** must be `Skill(<skill-name>)`. It must happen **before** any `Bash`, `Read`, `Grep`, `Glob`, `Edit`, `Write`, `Agent`, or MCP call related to that task.
- "Consulting" the guide or "mentally loading" a skill is not activation. If the `Skill` tool was not called, the skill is not active and its hard-gates, approval gates, and phase workflows are NOT in effect — even if you copy its steps.
- The phrasing "load the matching skill", "follow the skill", "let the skill drive X" always means: **call the `Skill` tool with that skill's name**. There is no other form of activation.

If you catch yourself about to run a tool for a task without having first invoked the matching skill: stop, call `Skill(<name>)`, and restart from inside the skill.

## Pre-flight: Before You Act

These rules are NOT optional. They apply to every task, every time.

- **Before spawning any subagent** (Explore, general-purpose, Plan), **before reading code for investigation**, and **before writing Grep/Glob queries** related to a new task: STOP. Consult the Quick Decision Guide above and invoke the matching skill via the `Skill` tool FIRST. Exploration, mapping, and investigation are part of the skill's workflow — they are NOT pre-skill activities.
- **On any task context switch within a session** — new bug, new feature, new ticket, merged branch moving to the next branch, or any pivot to unrelated work — re-consult the Quick Decision Guide and invoke the new task's skill. The previous task's skill does NOT carry over. Each task gets its own `Skill` tool invocation.
- If the current request doesn't clearly map to any row in the guide, ask the user to clarify before touching code. Do not improvise a workflow.

### Anti-Pattern: "Momentum"

The single most common failure mode for this ecosystem is **operating from momentum** after the first task of a session:

> "I already fixed one bug this session, so when the user reported the next one I jumped straight to spawning an Explore agent and reading code — I didn't re-check the Decision Guide or invoke `mobiai-mobile-debugging`."

This is wrong. Skills are NOT just for the "fix" step — they govern investigation, reproduction, exploration, and verification too. Reading code without having invoked the `Skill` tool means you are improvising a process that the skill already defines rigorously.

Signs you are in Momentum mode (stop and re-consult the guide):

- You're about to spawn a subagent to "map the code" or "find where X lives" for a new bug/feature, without naming which skill's workflow demands that exploration.
- You just finished a task and are reacting to the next message without a gate check.
- You're writing a Grep/Glob query for investigation and the `Skill` tool has not yet been invoked for `mobiai-mobile-debugging` (for bugs) or the relevant architecture skill (for features).
- You feel you "already know" the codebase from the previous task, so the skill would be redundant.

If any of these apply: stop, open the Decision Guide, invoke the matching skill via the `Skill` tool, and let the skill's loaded content drive the investigation.

### Anti-Pattern: "From Memory"

A second, subtler failure: you read the skill's row in this bootstrap, you remember what the skill does from a previous session, and you reproduce the workflow in-session without ever calling the `Skill` tool.

This is also wrong. The description you see here is only a pointer. The skill's HARD-GATEs, phase boundaries, approval checkpoints, and platform-specific branching only come into force once the `Skill` tool has loaded the full SKILL.md body. Reproducing the steps from memory means none of those gates apply.

If you find yourself thinking "I don't need to invoke it, I remember what it does": that is the exact moment to invoke it.

## Available Skills

Each entry below names a skill you activate by calling `Skill(<name>)` when the task matches its "When to Use" column. Do not paraphrase these skills from this table — invoke them.

### Core Workflow Skills
| Skill | When to Use |
|-------|-------------|
| `mobiai-fix-issue` | User asks to fix a bug from an issue tracker (Jira, GitHub Issues, Linear) |
| `mobiai-reproduce-bug` | User asks to reproduce a bug on a device, emulator, or simulator |
| `mobiai-analyze-crash` | User shares a crash from any source (stack trace, log, screenshot, description) |
| `mobiai-crashlytics` | User shares a Firebase Crashlytics link, crash ID, or mentions Crashlytics specifically |
| `mobiai-write-tests` | User asks to write tests for mobile code |
| `mobiai-review-code` | User asks for a code review of mobile code changes |
| `mobiai-create-pr` | User asks to create a pull request for mobile changes |

### Platform Skills
| Skill | When to Use |
|-------|-------------|
| `mobiai-android-device` | Need to interact with an Android device/emulator (adb, UI automation, screenshots) |
| `mobiai-android-build` | Need to build an Android project (Gradle, flavors, signing, APK/AAB) |
| `mobiai-android-testing` | Writing or running Android tests (JUnit, MockK, Espresso, Compose Testing) |
| `mobiai-android-architecture` | Working with Android architecture patterns (MVVM, MVI, Clean Architecture, Compose) |
| `mobiai-ios-device` | Need to interact with an iOS Simulator (simctl, screenshots, logs) |
| `mobiai-ios-build` | Need to build an iOS project (xcodebuild, schemes, CocoaPods, SPM) |
| `mobiai-ios-testing` | Writing or running iOS tests (XCTest, Quick/Nimble, snapshot tests) |
| `mobiai-ios-architecture` | Working with iOS architecture patterns (SwiftUI, UIKit, TCA, MVVM+Combine) |

### Cross-Platform Skills
| Skill | When to Use |
|-------|-------------|
| `mobiai-kmp` | Working with Kotlin Multiplatform projects |
| `mobiai-flutter` | Working with Flutter/Dart projects |
| `mobiai-react-native` | Working with React Native projects |

### Process Skills
| Skill | When to Use |
|-------|-------------|
| `mobiai-mobile-brainstorming` | Before creating any feature — explore design, platform constraints, tradeoffs |
| `mobiai-mobile-debugging` | Any bug, crash, or unexpected behavior — systematic root cause investigation |
| `mobiai-mobile-tdd` | Implementing any feature or fix — write tests first, then code |
| `mobiai-mobile-planning` | Planning a multi-step feature — detailed plans with exact paths and commands |
| `mobiai-mobile-verification` | Before claiming work is done — run builds and tests, verify output |
| `mobiai-mobile-executing-plans` | Executing an approved plan task by task with build checkpoints |
| `mobiai-mobile-parallel-agents` | 2+ independent tasks that can run simultaneously |
| `mobiai-mobile-subagent-development` | Executing plans with subagents — one per task, with spec and quality review |
| `mobiai-mobile-worktrees` | Starting feature work in isolation — git worktrees with platform setup |
| `mobiai-mobile-finishing-branch` | Implementation complete — decide how to integrate (merge, PR, keep, discard) |

### Meta Skills
| Skill | When to Use |
|-------|-------------|
| `mobiai-writing-skills` | User wants to create a new MobiAI skill |

## How to Access Skills

A skill is only active once you have invoked it via your platform's skill tool and received the loaded SKILL.md body back. Matching a description is NOT activation — invocation is.

**In Claude Code:** Call the `Skill` tool with the skill name (e.g. `Skill(mobiai-fix-issue)`). The SKILL.md body is returned as the tool result — follow it directly. If the `Skill` tool has not been called, the skill is not active.

**In Copilot CLI:** Call the `skill` tool. Skills are auto-discovered from installed plugins.

**In Gemini CLI:** Call `activate_skill` with the skill name. Gemini loads skill metadata at session start and activates the full content on demand.

**In Codex:** Skills are discovered natively from the symlinked skills directory. See `.codex/INSTALL.md` for setup.

**In other environments:** Check your platform's documentation for how skills are invoked.

## Platform Adaptation

Skills use Claude Code tool names (Read, Edit, Write, Bash, Grep, Glob, etc.) as the canonical reference. If you are on a different platform, consult the tool mapping references:

- **Copilot CLI**: `references/copilot-tools.md`
- **Codex**: `references/codex-tools.md`
- **Gemini CLI**: `references/gemini-tools.md` (loaded automatically via GEMINI.md)

## How to Use Skills

1. **Auto-detection (first gate, per task)**: This is the mandatory first step for every task. Before any other action — before spawning subagents, before reading code, before writing Grep/Glob queries — match the task against the Quick Decision Guide and **invoke the relevant skill via the `Skill` tool**. "Loading from memory" or "consulting" does not count; only a successful `Skill` tool call activates the skill. This gate fires **per-task, not per-session**: a new bug, new feature, new ticket, or context switch re-triggers it. The previous task's skill does not carry over. See "Pre-flight: Before You Act" above.
2. **Platform detection**: Check the project structure to determine the platform:
   - `build.gradle` / `build.gradle.kts` → Android
   - `*.xcodeproj` / `*.xcworkspace` → iOS
   - `shared/` + `composeApp/` / `iosApp/` → KMP
   - `pubspec.yaml` → Flutter
   - `package.json` + `metro.config.js` / `app.json` → React Native
3. **Combine skills**: Complex tasks often need multiple skills. For example, `mobiai-fix-issue` may invoke `mobiai-reproduce-bug`, `mobiai-analyze-crash`, `mobiai-write-tests`, and `mobiai-create-pr` as sub-steps.
4. **Project context**: Always read the project's `CLAUDE.md` or similar config files first — they contain project-specific conventions that override general skill guidance.

## Principles

- **Minimal fixes**: Change only what's necessary. Don't refactor, don't add unrelated improvements.
- **Evidence-based**: Always verify your work compiles and tests pass before declaring success.
- **Platform-native**: Follow each platform's conventions and idioms. Don't apply Android patterns to iOS or vice versa.
- **Community-driven**: These skills and agents are maintained by mobile developers worldwide. If you find something wrong or missing, check the `mobiai-writing-skills` skill to learn how to contribute.
