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

- **User reports a bug / shares a ticket** → `fix-issue` (full pipeline)
- **User shares a crash, stack trace, or error** → `analyze-crash`
- **User mentions Firebase Crashlytics specifically** → `crashlytics`
- **User wants to reproduce a bug on device** → `reproduce-bug`
- **User asks to write or add tests** → `write-tests`
- **User asks for code review** → `review-code`
- **User wants to create a PR** → `create-pr`
- **Need to interact with Android emulator/device** → `android-device`
- **Need to build an Android project** → `android-build`
- **Need to interact with iOS Simulator** → `ios-device`
- **Need to build an iOS project** → `ios-build`
- **Working on a KMP project** → `kmp`
- **Working on a Flutter project** → `flutter`
- **Working on a React Native project** → `react-native`
- **Need to understand Android project structure** → `android-architecture`
- **Need to understand iOS project structure** → `ios-architecture`
- **Writing Android tests specifically** → `android-testing`
- **Writing iOS tests specifically** → `ios-testing`
- **User wants to design a new feature before coding** → `mobile-brainstorming`
- **Bug or unexpected behavior, need to find root cause** → `mobile-debugging`
- **Implementing a feature or fix with tests first** → `mobile-tdd`
- **Planning a multi-step feature** → `mobile-planning`
- **About to claim work is done** → `mobile-verification`
- **Have a plan to execute step by step** → `mobile-executing-plans`
- **Multiple independent tasks to parallelize** → `mobile-parallel-agents`
- **Starting feature work in isolation** → `mobile-worktrees`
- **Implementation done, need to integrate** → `mobile-finishing-branch`
- **Creating a new MobiAI skill** → `writing-skills`

## Pre-flight: Before You Act

These rules are NOT optional. They apply to every task, every time.

- **Before spawning any subagent** (Explore, general-purpose, Plan), **before reading code for investigation**, and **before writing Grep/Glob queries** related to a new task: STOP. Consult the Quick Decision Guide above and load the matching skill FIRST. Exploration, mapping, and investigation are part of the skill's workflow — they are NOT pre-skill activities.
- **On any task context switch within a session** — new bug, new feature, new ticket, merged branch moving to the next branch, or any pivot to unrelated work — re-consult the Quick Decision Guide. The previous task's skill does NOT carry over. Each task gets its own gate check.
- If the current request doesn't clearly map to any row in the guide, ask the user to clarify before touching code. Do not improvise a workflow.

### Anti-Pattern: "Momentum"

The single most common failure mode for this ecosystem is **operating from momentum** after the first task of a session:

> "I already fixed one bug this session, so when the user reported the next one I jumped straight to spawning an Explore agent and reading code — I didn't re-check the Decision Guide or load `mobile-debugging`."

This is wrong. Skills are NOT just for the "fix" step — they govern investigation, reproduction, exploration, and verification too. Reading code without a skill loaded means you are improvising a process that the skill already defines rigorously.

Signs you are in Momentum mode (stop and re-consult the guide):

- You're about to spawn a subagent to "map the code" or "find where X lives" for a new bug/feature, without naming which skill's workflow demands that exploration.
- You just finished a task and are reacting to the next message without a gate check.
- You're writing a Grep/Glob query for investigation and haven't loaded `mobile-debugging` (for bugs) or the relevant architecture skill (for features).
- You feel you "already know" the codebase from the previous task, so the skill would be redundant.

If any of these apply: stop, open the Decision Guide, load the matching skill, and let the skill drive the investigation.

## Available Skills

These skills activate automatically based on context. When a task matches a skill's description, load and follow it.

### Core Workflow Skills
| Skill | When to Use |
|-------|-------------|
| `fix-issue` | User asks to fix a bug from an issue tracker (Jira, GitHub Issues, Linear) |
| `reproduce-bug` | User asks to reproduce a bug on a device, emulator, or simulator |
| `analyze-crash` | User shares a crash from any source (stack trace, log, screenshot, description) |
| `crashlytics` | User shares a Firebase Crashlytics link, crash ID, or mentions Crashlytics specifically |
| `write-tests` | User asks to write tests for mobile code |
| `review-code` | User asks for a code review of mobile code changes |
| `create-pr` | User asks to create a pull request for mobile changes |

### Platform Skills
| Skill | When to Use |
|-------|-------------|
| `android-device` | Need to interact with an Android device/emulator (adb, UI automation, screenshots) |
| `android-build` | Need to build an Android project (Gradle, flavors, signing, APK/AAB) |
| `android-testing` | Writing or running Android tests (JUnit, MockK, Espresso, Compose Testing) |
| `android-architecture` | Working with Android architecture patterns (MVVM, MVI, Clean Architecture, Compose) |
| `ios-device` | Need to interact with an iOS Simulator (simctl, screenshots, logs) |
| `ios-build` | Need to build an iOS project (xcodebuild, schemes, CocoaPods, SPM) |
| `ios-testing` | Writing or running iOS tests (XCTest, Quick/Nimble, snapshot tests) |
| `ios-architecture` | Working with iOS architecture patterns (SwiftUI, UIKit, TCA, MVVM+Combine) |

### Cross-Platform Skills
| Skill | When to Use |
|-------|-------------|
| `kmp` | Working with Kotlin Multiplatform projects |
| `flutter` | Working with Flutter/Dart projects |
| `react-native` | Working with React Native projects |

### Process Skills
| Skill | When to Use |
|-------|-------------|
| `mobile-brainstorming` | Before creating any feature — explore design, platform constraints, tradeoffs |
| `mobile-debugging` | Any bug, crash, or unexpected behavior — systematic root cause investigation |
| `mobile-tdd` | Implementing any feature or fix — write tests first, then code |
| `mobile-planning` | Planning a multi-step feature — detailed plans with exact paths and commands |
| `mobile-verification` | Before claiming work is done — run builds and tests, verify output |
| `mobile-executing-plans` | Executing an approved plan task by task with build checkpoints |
| `mobile-parallel-agents` | 2+ independent tasks that can run simultaneously |
| `mobile-subagent-development` | Executing plans with subagents — one per task, with spec and quality review |
| `mobile-worktrees` | Starting feature work in isolation — git worktrees with platform setup |
| `mobile-finishing-branch` | Implementation complete — decide how to integrate (merge, PR, keep, discard) |

### Meta Skills
| Skill | When to Use |
|-------|-------------|
| `writing-skills` | User wants to create a new MobiAI skill |

## How to Access Skills

**In Claude Code:** Use the `Skill` tool. When you invoke a skill, its content is loaded and presented to you — follow it directly.

**In Copilot CLI:** Use the `skill` tool. Skills are auto-discovered from installed plugins.

**In Gemini CLI:** Skills activate via the `activate_skill` tool. Gemini loads skill metadata at session start and activates the full content on demand.

**In Codex:** Skills are discovered natively from the symlinked skills directory. See `.codex/INSTALL.md` for setup.

**In other environments:** Check your platform's documentation for how skills are loaded.

## Platform Adaptation

Skills use Claude Code tool names (Read, Edit, Write, Bash, Grep, Glob, etc.) as the canonical reference. If you are on a different platform, consult the tool mapping references:

- **Copilot CLI**: `references/copilot-tools.md`
- **Codex**: `references/codex-tools.md`
- **Gemini CLI**: `references/gemini-tools.md` (loaded automatically via GEMINI.md)

## How to Use Skills

1. **Auto-detection (first gate, per task)**: This is the mandatory first step for every task. Before any other action — before spawning subagents, before reading code, before writing Grep/Glob queries — match the task against the Quick Decision Guide and load the relevant skill. This gate fires **per-task, not per-session**: a new bug, new feature, new ticket, or context switch re-triggers it. The previous task's skill does not carry over. See "Pre-flight: Before You Act" above.
2. **Platform detection**: Check the project structure to determine the platform:
   - `build.gradle` / `build.gradle.kts` → Android
   - `*.xcodeproj` / `*.xcworkspace` → iOS
   - `shared/` + `composeApp/` / `iosApp/` → KMP
   - `pubspec.yaml` → Flutter
   - `package.json` + `metro.config.js` / `app.json` → React Native
3. **Combine skills**: Complex tasks often need multiple skills. For example, `fix-issue` may invoke `reproduce-bug`, `analyze-crash`, `write-tests`, and `create-pr` as sub-steps.
4. **Project context**: Always read the project's `CLAUDE.md` or similar config files first — they contain project-specific conventions that override general skill guidance.

## Principles

- **Minimal fixes**: Change only what's necessary. Don't refactor, don't add unrelated improvements.
- **Evidence-based**: Always verify your work compiles and tests pass before declaring success.
- **Platform-native**: Follow each platform's conventions and idioms. Don't apply Android patterns to iOS or vice versa.
- **Community-driven**: These skills and agents are maintained by mobile developers worldwide. If you find something wrong or missing, check the `writing-skills` skill to learn how to contribute.
