---
name: mobile-verification
description: Use when about to claim mobile work is complete, fixed, or passing, before committing or creating PRs - requires running verification commands and confirming output before making any success claims; evidence before assertions always
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Mobile Verification Before Completion

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
