---
name: mobiai-mobile-verification
description: "You MUST use this before declaring any mobile work done, fixed, green, or ready — and before committing, pushing, or opening a PR. No completion claim may be made without fresh verification evidence from this session. Activate via the `Skill` tool; do not paraphrase this skill's workflow from memory."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Mobile Verification Before Completion

## Activation

You are reading this because `Skill(mobiai-mobile-verification)` was invoked — correct. Every HARD-GATE, phase checkpoint, and approval gate in this document is binding **because** the `Skill` tool was called. None of them bind if a future step, subagent, or different session reproduces this workflow from memory without another `Skill(mobiai-mobile-verification)` call.

If you need any of this skill's steps in another context, invoke `Skill(mobiai-mobile-verification)` again. Paraphrasing from memory is not activation.

<HARD-GATE>
Do NOT state, imply, or signal that work is complete, fixed, passing, ready, or merged-ready until you have just run the verification commands in this message and observed their output. Stale memory of earlier runs does not count.
</HARD-GATE>

## Overview

Claiming work is complete without verification is dishonesty, not efficiency.

**Core principle:** Evidence before claims, always.

**Violating the letter of this rule is violating the spirit of this rule.**

## The Iron Law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

If you haven't run the verification command in this message, you cannot claim it passes.

## The Gate Function

```
BEFORE claiming any status or expressing satisfaction:

1. IDENTIFY: What command proves this claim?
2. RUN: Execute the FULL command (fresh, complete)
3. READ: Full output, check exit code, count failures
4. VERIFY: Does output confirm the claim?
   - If NO: State actual status with evidence
   - If YES: State claim WITH evidence
5. ONLY THEN: Make the claim

Skip any step = lying, not verifying
```

## Mobile Verification Commands

| Claim | Android | iOS | Flutter | React Native |
|-------|---------|-----|---------|--------------|
| **Compiles** | `./gradlew assembleDebug` | `xcodebuild build -scheme <S> -destination '...'` | `flutter build apk --debug` | `npx react-native build-android` |
| **Unit tests** | `./gradlew testDebugUnitTest` | `xcodebuild test -scheme <S> -destination '...'` | `flutter test` | `npx jest` |
| **All tests** | `./gradlew connectedDebugAndroidTest` | same as above | `flutter test` | `npx jest` |
| **KMP shared** | `./gradlew allTests` | — | — | — |

## Common Failures

| Claim | Requires | Not Sufficient |
|-------|----------|----------------|
| Tests pass | Test command output: 0 failures | Previous run, "should pass" |
| Build succeeds | Build command: exit 0 | "It compiled before my changes" |
| Bug fixed | Test original symptom: passes | "I changed the line that was wrong" |
| No regressions | Run full test suite | "I only changed one file" |
| Works on both platforms | Verify on EACH platform | "If it works on Android it should work on iOS" |

## Red Flags - STOP

- Using "should", "probably", "seems to"
- Expressing satisfaction before verification ("Great!", "Done!")
- About to commit/push/PR without verification
- Relying on partial verification
- **ANY wording implying success without having run verification**

## Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "Should work now" | RUN the verification |
| "I'm confident" | Confidence ≠ evidence |
| "Just this once" | No exceptions |
| "Linter passed" | Linter ≠ compiler |
| "I'm tired" | Exhaustion ≠ excuse |
| "Partial check is enough" | Partial proves nothing |

## Multi-Platform Verification

For KMP, Flutter, and React Native projects, verify EACH platform separately:

1. Build for platform A → confirm success
2. Run tests for platform A → confirm pass
3. Build for platform B → confirm success
4. Run tests for platform B → confirm pass
5. Only then claim "works on both platforms"

## When To Apply

**ALWAYS before:**
- ANY variation of success/completion claims
- ANY expression of satisfaction
- Committing, PR creation, task completion
- Moving to next task

## The Bottom Line

**No shortcuts for verification.**

Run the command. Read the output. THEN claim the result.

This is non-negotiable.
