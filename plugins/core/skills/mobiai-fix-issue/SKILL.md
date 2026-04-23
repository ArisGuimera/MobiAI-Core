---
name: mobiai-fix-issue
description: "You MUST use this before starting any bug fix from a ticket or issue, OR when re-opening a ticket whose previous fix attempt didn't work (user shares a commit SHA and says 'no anda' / 'didn't fix it' / 'still broken'). Fetch, understand, investigate root cause, apply the fix with tests, and verify. For small line-level fixes, runs autonomously end-to-end and gates only before push. For complex changes, gates at each phase boundary."
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Fix Issue

End-to-end workflow for fixing a bug reported in an issue tracker (Jira, GitHub Issues, Linear, etc.). The flow runs all phases in order — understand, reproduce (only if needed), investigate root cause, apply fix, verify, open PR — but the level of user gating adapts to the scope of the change.

<HARD-GATE>
You MUST NOT push to a remote branch or open a PR without the user approving the final diff and test evidence. This final gate is non-negotiable for every ticket.

For SMALL line-level fixes (clear root cause, ≤5 lines of production change in ≤2 files, no schema / permissions / build-config / API impact): work autonomously through phases 1–6, then gate at Phase 7 (via `mobiai-create-pr`).

For LARGE changes (multi-file refactors, unclear root cause, new behavior, breaking API, architectural choice): enforce user-approval gates at each phase boundary as detailed below.
</HARD-GATE>

## Anti-Pattern: "I Already Know What's Wrong"

The most common failure mode is jumping from the ticket description straight to a fix, or spawning an Explore/sub-agent that returns a confidently wrong diagnosis which is then implemented without verification. Every ticket, including the "obvious" ones, goes through the phased flow — Phases 1 through 6 run end-to-end regardless of scope. The scope assessment only changes whether you **wait for user confirmation** between phases, not whether you **do the work** in each phase.

If you think "I already know what this is", you may still be right — but you must still run investigation, reproduce if needed, and verify. What you skip in fast path is the ceremony of asking permission, not the analysis itself.

## Pre-flight: Scope Assessment

Before starting Phase 1, classify the task. State your classification to the user in one sentence at the start so they know what to expect:

**Fast path** — go end-to-end autonomously, gate only before push:
- Clear root cause visible from the error message / stack trace / description
- Fix is line-level: ≤5 lines of production change in ≤2 files
- No DB migration, no new permission, no ProGuard/R8 rule change, no build-config change
- No new API, no architectural decision, no breaking change
- No new feature behavior

**Gated path** — stop and wait for user confirmation at each phase boundary:
- Multi-file refactor or architectural change
- Unclear root cause, contradictory evidence, or multiple plausible causes
- DB schema change, permission added, ProGuard rule, build config change
- New feature or behavior addition
- Breaking change to public API, SDK, or cross-module contract

In doubt → gated path. The cost of an unnecessary gate is small (an extra user confirmation); the cost of applying a risky change without review is large.

Declare the decision upfront:

> "Fast path — clear root cause, small scope. Arranco la investigación y te muestro el diff antes de pushear."

or

> "Gated path — <reason: unclear root cause / multi-file refactor / new feature / schema change>. Voy a confirmar contigo en cada fase."

## Checklist

You MUST create a task for each of these items and complete them in order. Do not skip phases. Do not merge phases.

Pre-flight: **Classify scope** — fast path or gated path (Pre-flight above)

1. **Understand the bug** — fetch the ticket, read it fully, detect platform, gather context
2. **Reproduce the bug** — only if the user explicitly asked for device reproduction or the bug is non-deterministic and code alone can't explain it; otherwise document existing evidence (stack trace, logs, description)
3. **Investigate root cause** — invoke skill `mobiai-mobile-debugging` and follow its phased flow to completion
4. **Propose the fix** — gated path: require explicit user approval before Phase 5; fast path: note the proposal for the PR description and proceed
5. **Implement the fix** — invoke skill `mobiai-mobile-tdd` (test first, then minimal implementation)
6. **Verify the fix** — invoke skill `mobiai-mobile-verification` (compile, tests, platform checks)
7. **Open the PR** — invoke skill `mobiai-create-pr` (final user-approval gate before push — always)

## When to Use

- User provides an issue key/URL and asks to fix it (e.g., "fix PROJ-123", "fix #456")
- User shares a bug report and wants it resolved with a PR
- User pastes a crash, log, or description and asks for an end-to-end fix

## Phase 1: Understand the Bug

Get the full issue details before anything else — summary, description, reproduction steps, stack traces, affected versions, attachments.

**Jira**: Use the Atlassian MCP tools (`getJiraIssue`, `searchJiraIssuesUsingJql`) if available, or ask the user to paste the details.

**GitHub Issues**: Use `gh issue view <number>` via Bash.

**Manual**: If no integration, ask the user to paste the issue content.

Then, before touching code:

1. **Detect the platform** by checking project structure:
   - `build.gradle(.kts)` → Android → load `mobiai-android-build` and `mobiai-android-architecture` skills
   - `*.xcodeproj` / `*.xcworkspace` → iOS → load `mobiai-ios-build` and `mobiai-ios-architecture` skills
   - `shared/` + `composeApp/` → KMP → load `mobiai-kmp` skill
   - `pubspec.yaml` → Flutter → load `mobiai-flutter` skill
   - `metro.config.js` / `app.json` + `package.json` → React Native → load `mobiai-react-native` skill

2. **Read the project's CLAUDE.md** or similar configuration for project-specific conventions.

3. **Search for similar past fixes** in git history:
   ```bash
   git log --oneline --all --grep="<keyword from bug>" | head -20
   ```

Present a short summary: what the bug is, which platform(s) it affects, what you understood from the ticket.

- **Fast path**: proceed directly to Phase 2.
- **Gated path**: wait for the user to confirm understanding before Phase 2.

## Phase 2: Reproduce the Bug

**Default: skip device/emulator/simulator reproduction.** Reproduction is slow (boot, install, navigation) and most bugs are diagnosable from stack traces, logs, or clear descriptions.

**Only invoke `mobiai-reproduce-bug` when**:
- The user's original request explicitly asked for device reproduction ("reproducí esto en el simulador", "probalo en el emulador", "try to reproduce")
- The bug is non-deterministic or purely UI-driven and code analysis genuinely cannot explain it

Otherwise: document the existing evidence (stack trace, logs, description) as the reproduction evidence and move on.

- **Fast path**: proceed directly to Phase 3 after documenting evidence.
- **Gated path**: state the evidence collected and wait for user confirmation before Phase 3.

## Phase 3: Investigate Root Cause

**Invoke skill `mobiai-mobile-debugging`** and follow its phased flow. The child skill inherits the scope classification from this task — fast path runs `mobiai-mobile-debugging` autonomously and returns with the root cause identified; gated path triggers `mobiai-mobile-debugging`'s own confirmation gate at the root-cause step.

Do NOT shortcut investigation regardless of path. Do NOT propose a fix in this phase. Do NOT edit code in this phase. The output is a stated, evidence-backed root cause.

When `mobiai-mobile-debugging` returns with a root cause, proceed to Phase 4.

## Phase 4: Propose the Fix

Prepare the fix proposal:

1. **The root cause** (one or two sentences)
2. **The proposed fix** — which file(s), which function(s), the shape of the change
3. **The scope** — exactly what will change and, explicitly, what will NOT change
4. **Risks and side effects** — other flows that touch the same code, platform-specific concerns (Android lifecycle, iOS main thread, KMP expect/actual, Flutter widget rebuilds, RN bridge)
5. **Test strategy** — what tests will be added or updated

Rules for the proposed fix (apply in both paths):

- Change **only** what is necessary to fix the bug
- Do NOT refactor surrounding code, rename variables, add comments, or change formatting
- Do NOT add features or improve error handling beyond what's needed
- Do NOT modify changelog files, README, or CI configuration unless the bug is in them

Gate behavior:

- **Fast path**: save the proposal as a short note for the PR description (Phase 7 will use it) and proceed directly to Phase 5. The user will review the full diff + test evidence at Phase 7 before push.
- **Gated path**: present the proposal to the user and ask:

  > "This is my proposed fix. Do you approve it, or do you want to adjust the scope or approach before I write any code?"

  Wait for explicit approval. If the user asks for changes, revise and re-present.

## Phase 5: Implement the Fix

**Invoke skill `mobiai-mobile-tdd`** (test first, then minimal implementation).

- Write the failing test first, covering the specific behavior that was broken
- Add edge-case tests related to the fix
- Add regression tests so the bug cannot silently return
- Implement the minimum production change that makes the tests pass
- One change at a time, no "while I'm here" improvements

`mobiai-mobile-tdd` also inherits the scope classification — fast path runs red-green-refactor autonomously; gated path gates at the appropriate checkpoint.

## Phase 6: Verify the Fix

**Invoke skill `mobiai-mobile-verification`**. Do not claim the fix works until this skill has run its checks and you have the evidence in hand.

Platform-specific compile and test commands (handled by `mobiai-mobile-verification`, listed here for reference):

- **Android**: `./gradlew compile<Flavor>DebugKotlin --no-daemon` then `./gradlew test<Flavor>DebugUnitTest --no-daemon`
- **iOS**: `xcodebuild build` and `xcodebuild test` with the appropriate scheme and simulator
- **Flutter**: `flutter analyze && flutter build apk --debug` then `flutter test`
- **React Native**: `npx tsc --noEmit` then `npx jest`
- **KMP**: `./gradlew compileKotlin` then `./gradlew test`

If verification fails, do not rationalize. Return to Phase 3 with the new evidence.

## Phase 7: Open the PR — FINAL GATE (always)

**Invoke skill `mobiai-create-pr`**. This is the non-negotiable gate for every ticket regardless of path — before pushing, the user will see the full diff plus the context below and must explicitly approve.

What to present to the user at this gate (and include in the PR description):

1. **Ticket reference** — link/key of the issue
2. **Root cause** — the confirmed root cause from Phase 3, in plain language (not the symptom)
3. **Breaking commit** — identified during Phase 3 investigation. The SHA + subject of the commit that introduced the bug; or `"always-existed, not a regression"` if applicable. Do not leave this ambiguous.
4. **What changed** — file-by-file factual summary (what code was touched)
5. **Why each change** — the reasoning that connects the root cause to each specific modification. NOT what the diff already shows — why it had to change. One "why" per logical change.
6. **Evidence it works** — test results, compile output, platform-specific checks from Phase 6
7. **Platform notes** — risks, side effects, anything platform-specific the reviewer should know

Do not push until the user has seen all seven and approved.

## Decision Rules

- **Never skip a phase.** Every ticket runs through the full checklist in order. Scope classification only changes gate behavior, not which execution phases run.
- **Upgrade from fast to gated mid-flow if needed.** If Phase 3 reveals that the root cause is more complex than expected, or Phase 5 shows the fix needs more files than estimated, explicitly switch to gated path and notify the user.
- **Never stop after a single failure.** If a phase surfaces a problem, loop back and re-run from the right point.
- **Max 2 fix attempts after proposal.** If the fix fails verification twice, return to Phase 3, do not keep patching.
- **If you can't find the cause**, stay in Phase 3 and ask the user for hints — do not jump to Phase 4 with a guess.
- **If the fix turns out to be risky** (schema migrations, multi-module architectural changes), escalate to gated path before proposing the fix.

## Communication

Keep the user informed at phase boundaries with a brief status line — what you found, what you're doing next.

- **Fast path**: one-line updates per phase, then full diff + test output at Phase 7.
- **Gated path**: explicit pause at each phase boundary for user confirmation.

No silent progression — the user must be able to follow along even in fast path.
