---
name: fix-issue
description: "You MUST use this before starting any bug fix from a ticket or issue — fetch, understand, reproduce, diagnose, and get explicit user approval before writing any fix. Orchestrates the full bug-fix pipeline with mandatory approval gates."
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Fix Issue

End-to-end workflow for fixing a bug reported in an issue tracker (Jira, GitHub Issues, Linear, etc.), with explicit user-approval gates between phases.

<HARD-GATE>
Do NOT write any fix code, edit any production file for the fix, invoke `mobile-tdd`, or open a PR until you have (a) completed root-cause investigation, (b) presented the diagnosis and proposed fix to the user, and (c) received explicit user approval. This applies to EVERY ticket regardless of perceived simplicity, including "obvious" one-line fixes and emergency hotfixes.
</HARD-GATE>

## Anti-Pattern: "I Already Know What's Wrong"

The most common failure mode is jumping from the ticket description straight to a fix, or spawning an Explore/sub-agent that returns a confidently wrong diagnosis which is then implemented without verification. Every ticket, including the "obvious" ones, goes through the phased flow. A misread symptom costs more than a few minutes of structured investigation. If you find yourself thinking "I already know what this is", STOP and run the phases anyway.

## Checklist

You MUST create a task for each of these items and complete them in order. Do not skip phases. Do not merge phases. Do not proceed without the user's explicit approval where required.

1. **Understand the bug** — fetch the ticket, read it fully, detect platform, gather context
2. **Reproduce the bug** — invoke skill `reproduce-bug` when applicable, or document why repro is skipped
3. **Investigate root cause** — invoke skill `mobile-debugging` and follow its phased flow to completion
4. **Propose the fix — USER APPROVAL REQUIRED before writing any code**
5. **Implement the fix** — invoke skill `mobile-tdd` (test first, then implementation)
6. **Verify the fix** — invoke skill `mobile-verification` (compile, tests, platform checks)
7. **Open the PR** — invoke skill `create-pr`

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
   - `build.gradle(.kts)` → Android → load `android-build` and `android-architecture` skills
   - `*.xcodeproj` / `*.xcworkspace` → iOS → load `ios-build` and `ios-architecture` skills
   - `shared/` + `composeApp/` → KMP → load `kmp` skill
   - `pubspec.yaml` → Flutter → load `flutter` skill
   - `metro.config.js` / `app.json` + `package.json` → React Native → load `react-native` skill

2. **Read the project's CLAUDE.md** or similar configuration for project-specific conventions.

3. **Search for similar past fixes** in git history:
   ```bash
   git log --oneline --all --grep="<keyword from bug>" | head -20
   ```

Present a short summary to the user: what the bug is, which platform(s) it affects, what you understood from the ticket. **Wait for the user to confirm the understanding is correct before proceeding to Phase 2.**

## Phase 2: Reproduce the Bug

Decide the reproduction strategy based on what's available:

- **Has reproduction steps + device available** → invoke skill `reproduce-bug` to reproduce on device/emulator/simulator
- **Has stack trace / crash signal only** → document the crash signal as the reproduction evidence
- **Vague description, no repro steps** → ask the user for more detail before continuing
- **User explicitly says "skip repro" / "just fix the code"** → document the skip reason and move on

State to the user either (a) the bug was reproduced and here is the evidence, or (b) reproduction was skipped for this specific reason. **Wait for the user to confirm before proceeding to Phase 3.**

## Phase 3: Investigate Root Cause

**Invoke skill `mobile-debugging`** and follow its phased flow end-to-end. That skill has its own user-confirmation gate at the root-cause step — do not bypass it.

Do NOT shortcut this phase. Do NOT propose a fix here. Do NOT edit code here. The output of this phase is a clearly stated root cause that has been confirmed by the user inside `mobile-debugging`.

When `mobile-debugging` returns with a user-confirmed root cause, proceed to Phase 4.

## Phase 4: Propose the Fix — USER APPROVAL REQUIRED

Present to the user, in plain language:

1. **The root cause** (one or two sentences, already confirmed in Phase 3)
2. **The proposed fix** — which file(s), which function(s), the shape of the change
3. **The scope** — exactly what will change and, explicitly, what will NOT change
4. **Risks and side effects** — other flows that touch the same code, platform-specific concerns (Android lifecycle, iOS main thread, KMP expect/actual, Flutter widget rebuilds, RN bridge)
5. **Test strategy** — what tests will be added or updated

Rules for the proposed fix:
- Change **only** what is necessary to fix the bug
- Do NOT refactor surrounding code, rename variables, add comments, or change formatting
- Do NOT add features or improve error handling beyond what's needed
- Do NOT modify changelog files, README, or CI configuration unless the bug is in them

Then ask explicitly:

> "This is my proposed fix. Do you approve it, or do you want to adjust the scope or approach before I write any code?"

**Wait for the user's explicit approval. Do not proceed, do not write tests, do not edit files, until the user has said yes. If the user asks for changes, revise and re-present.**

## Phase 5: Implement the Fix

Only after Phase 4 approval: **invoke skill `mobile-tdd`**.

- Write the failing test first, covering the specific behavior that was broken
- Add edge-case tests related to the fix
- Add regression tests so the bug cannot silently return
- Implement the minimum production change that makes the tests pass
- One change at a time, no "while I'm here" improvements

## Phase 6: Verify the Fix

**Invoke skill `mobile-verification`**. Do not claim the fix works until this skill has run its checks and you have the evidence in hand.

Platform-specific compile and test commands (handled by `mobile-verification`, listed here for reference):

- **Android**: `./gradlew compile<Flavor>DebugKotlin --no-daemon` then `./gradlew test<Flavor>DebugUnitTest --no-daemon`
- **iOS**: `xcodebuild build` and `xcodebuild test` with the appropriate scheme and simulator
- **Flutter**: `flutter analyze && flutter build apk --debug` then `flutter test`
- **React Native**: `npx tsc --noEmit` then `npx jest`
- **KMP**: `./gradlew compileKotlin` then `./gradlew test`

If verification fails, do not rationalize. Return to Phase 3 with the new evidence.

## Phase 7: Open the PR

**Invoke skill `create-pr`**. The PR description must include:

- The ticket reference
- The root cause (from Phase 3)
- The fix (from Phase 4, as approved)
- The evidence that it works (from Phase 6)
- Platform-specific notes where relevant

## Decision Rules

- **Never skip a phase.** Every ticket runs through 1 → 7 in order.
- **Never stop after a single failure.** If one phase surfaces a problem, loop back and re-run from the right point, do not give up.
- **Max 2 fix attempts after approval.** If the approved fix fails verification twice, return to Phase 3, do not keep patching.
- **If you can't find the cause**, stay in Phase 3 and ask the user for hints, do not jump to Phase 4 with a guess.
- **If the fix is too risky** (schema migrations, multi-module architectural changes), surface this in Phase 4 and let the user decide the approach before approval.

## Communication

Keep the user informed at each phase boundary: what you found, what you're asking them to confirm, and what the next phase will do. The user must see each step and approve where required — no silent progression.
