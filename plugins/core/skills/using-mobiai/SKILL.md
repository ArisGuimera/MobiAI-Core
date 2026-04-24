---
name: using-mobiai
description: Use when starting any conversation — establishes how to find and invoke MobiAI skills, requiring `Skill` tool invocation before ANY response including clarifying questions, git/file reads, or code exploration
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific, scoped task, skip this skill. Your parent agent is responsible for skill routing.
</SUBAGENT-STOP>

<EXTREMELY-IMPORTANT>
If you think there is even a 1% chance a MobiAI skill might apply to what you are doing, you ABSOLUTELY MUST invoke the skill via the `Skill` tool.

IF A SKILL APPLIES TO YOUR TASK, YOU DO NOT HAVE A CHOICE. YOU MUST USE IT.

This is not negotiable. This is not optional. You cannot rationalize your way out of this.
</EXTREMELY-IMPORTANT>

## Instruction Priority

MobiAI skills override default system-prompt behavior, but **user instructions always take precedence**:

1. **User's explicit instructions** (CLAUDE.md, GEMINI.md, AGENTS.md, direct requests) — highest priority
2. **MobiAI skills** — override default behavior where they conflict
3. **Default system prompt** — lowest priority

If a CLAUDE.md or user message conflicts with a skill's workflow, follow the user. The user is in control.

## How to Access Skills

**In Claude Code:** Call the `Skill` tool with the skill name (e.g. `Skill(mobiai-fix-issue)`). The SKILL.md body is returned as the tool result — follow it directly. **Never use the `Read` tool on `SKILL.md` files** — that bypasses invocation and is not activation.

**In Copilot CLI:** Call the `skill` tool. Skills are auto-discovered from installed plugins.

**In Gemini CLI:** Call `activate_skill` with the skill name.

**In Codex:** Skills are discovered natively from the symlinked skills directory. See `.codex/INSTALL.md`.

**In other environments:** Check your platform's documentation for how skills are invoked.

## Quick Decision Guide

Match the user's intent against the rows below. If any row matches even at 1% confidence — invoke that skill FIRST, before any other tool call.

| User intent | Skill to invoke |
|-------------|-----------------|
| User reports a bug / shares a ticket (Jira, Linear, GitHub issues) | `mobiai-fix-issue` |
| User shares a commit SHA + says the fix didn't work, or asks whether a commit resolves a bug | `mobiai-mobile-debugging` (or `mobiai-fix-issue` if the ticket is still open) |
| User shares a crash, stack trace, or error | `mobiai-analyze-crash` |
| User mentions Firebase Crashlytics specifically | `mobiai-crashlytics` |
| User wants to reproduce a bug on device/emulator/simulator | `mobiai-reproduce-bug` |
| User asks to write or add tests | `mobiai-write-tests` |
| User asks for a code review | `mobiai-review-code` |
| User wants to commit, push, or open a PR (says "mergear", "push", "subir", "dale PR", "shipeá", "commit", "merge") | `mobiai-create-pr` |
| Interact with Android device/emulator (adb, screenshots, logcat) | `mobiai-android-device` |
| Build an Android project (Gradle, flavors, variants) | `mobiai-android-build` |
| Interact with iOS Simulator (simctl) | `mobiai-ios-device` |
| Build an iOS project (xcodebuild, schemes) | `mobiai-ios-build` |
| Working on a Kotlin Multiplatform project | `mobiai-kmp` |
| Working on a Flutter project | `mobiai-flutter` |
| Working on a React Native project | `mobiai-react-native` |
| Need to understand Android project structure | `mobiai-android-architecture` |
| Need to understand iOS project structure | `mobiai-ios-architecture` |
| Writing Android tests specifically | `mobiai-android-testing` |
| Writing iOS tests specifically | `mobiai-ios-testing` |
| Design a new feature with **unclear intent or requirements** (user still exploring what to build) | `mobiai-mobile-brainstorming` |
| Bug or unexpected behavior, need root cause | `mobiai-mobile-debugging` |
| Implementing a feature/fix with tests first | `mobiai-mobile-tdd` |
| User gives a **defined task list (3+ items)** or a multi-step spec — intent clear, "how" missing | `mobiai-mobile-planning` |
| About to claim work is done | `mobiai-mobile-verification` |
| Have a written plan to execute step by step | `mobiai-mobile-executing-plans` |
| 2+ independent tasks to parallelize (says "en paralelo", "orquestador", "agentes al mismo tiempo", "waves", "al mismo tiempo") | `mobiai-mobile-parallel-agents` |
| Executing a plan via subagents (says "ejecutá el plan", "implementalo", "arrancá con esto", "hacelo con agentes") | `mobiai-mobile-executing-plans-with-subagents` |
| Starting feature work in isolation | `mobiai-mobile-worktrees` |
| Implementation done, need to integrate | `mobiai-mobile-finishing-branch` |
| Creating a new MobiAI skill | `mobiai-writing-skills` |

If the request doesn't match any row, ask the user to clarify before touching code. Do not improvise a workflow.

## The Rule

**Invoke relevant skills BEFORE any response or action** — including clarifying questions, reading files, running `git show` or `git log`, spawning subagents, or exploring the filesystem. Even a 1% chance a skill might apply means invoke it to check. If the invoked skill turns out not to fit, you can move on — but the check must happen first.

## Announce What You're Using

Immediately after invoking a skill, announce one sentence to the user:

> "Using `mobiai-fix-issue` to investigate the QA-548 ticket you shared."

This makes the invocation visible and commits you to following the skill's workflow instead of drifting back to improvisation.

## Red Flags

These thoughts mean STOP — you're rationalizing:

| Thought | Reality |
|---------|---------|
| "This is just a diagnostic question, not a fix request" | Diagnostic questions about broken code are tasks. `mobiai-mobile-debugging` exists for exactly this. Invoke it. |
| "Let me `git show` this commit quickly first" | `git show` output lacks ticket context. Invoke the skill first — it tells you HOW to read the commit. |
| "Let me grep the codebase to understand first" | Skills tell you HOW to explore. Check the guide first. |
| "The user is asking a question, not a task" | Questions that need code investigation are tasks. Check the guide. |
| "I already fixed a bug this session, I know the pattern" | Each task gets its own skill check. Momentum ≠ authorization. |
| "I remember what this skill does from last session" | Skills evolve. Invoke it to read the current version. |
| "This looks like a simple one-liner, the skill is overkill" | HARD-GATEs apply regardless of size. Simple fixes still need invocation. |
| "The user shared a commit SHA, so it's a research request" | Commit SHA + "doesn't work" = debugging task. Invoke `mobiai-mobile-debugging`. |
| "Let me check current branch/state quickly first" | `git status`/`git log` without a skill loaded is improvisation. Invoke first. |
| "The user said 'mirá este error' — that's passive, not a task" | Requests to look at broken code are tasks. Invoke. |
| "I need more context before I can pick a skill" | Pick the closest match at 1% confidence. Wrong-skill is recoverable; no-skill is not. |
| "I'll just do this one thing first" | Check the guide BEFORE doing anything. |

## Anti-Pattern: "From Memory"

A subtle variant: you remember what a skill does from a previous session and you reproduce its steps in-session without ever calling `Skill(...)`.

That is not activation. The skill's HARD-GATEs, phase boundaries, approval checkpoints, and platform-specific branching only come into force once the `Skill` tool has loaded the current SKILL.md body. If you catch yourself thinking "I don't need to invoke it, I remember what it does" — that is the exact moment to invoke it.

## Skill Priority

When multiple skills could apply:

1. **Process skills first** (`mobiai-mobile-brainstorming`, `mobiai-mobile-debugging`) — determine HOW to approach the task.
2. **Workflow skills second** (`mobiai-fix-issue`, `mobiai-create-pr`) — orchestrate multi-phase work and delegate to process/implementation skills as sub-steps.
3. **Implementation skills third** (`mobiai-android-build`, `mobiai-ios-device`, platform-specifics) — guide execution.

Examples:
- "Fix QA-548" → `mobiai-fix-issue` (workflow). It will internally invoke debugging, TDD, verification, and create-pr.
- "Design feature X" → `mobiai-mobile-brainstorming` first, then planning/implementation.

## Platform Detection

Read the project structure to pick platform-specific skills:

- `build.gradle(.kts)` → Android → `mobiai-android-*`
- `*.xcodeproj` / `*.xcworkspace` → iOS → `mobiai-ios-*`
- `shared/` + `composeApp/` → KMP → `mobiai-kmp`
- `pubspec.yaml` → Flutter → `mobiai-flutter`
- `package.json` + `metro.config.js` / `app.json` → React Native → `mobiai-react-native`

Always read the project's `CLAUDE.md` (or equivalent) for project-specific conventions.

## Principles

- **Minimal fixes**: Change only what's necessary. No drive-by refactors.
- **Evidence-based**: Verify compile + tests pass before declaring done.
- **Platform-native**: Follow each platform's conventions and idioms.
- **Community-driven**: Found a gap? Use `mobiai-writing-skills` to contribute.

---

*The structure and tone of this bootstrap are inspired by Jesse Vincent's [Superpowers](https://github.com/obra/superpowers) (MIT). See `NOTICE.md` for attribution.*
