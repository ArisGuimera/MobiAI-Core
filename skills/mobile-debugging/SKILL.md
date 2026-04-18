---
name: mobile-debugging
description: Use when encountering any mobile bug, test failure, or unexpected behavior, before proposing fixes
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Systematic Mobile Debugging

## Overview

Random fixes waste time and create new bugs. Quick patches mask underlying issues.

**Core principle:** ALWAYS find root cause before attempting fixes. Symptom fixes are failure.

**Violating the letter of this process is violating the spirit of debugging.**

## The Iron Law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

If you haven't completed Phase 1, you cannot propose fixes.

## When to Use

Use for ANY mobile technical issue:
- Test failures
- Bugs in production
- Unexpected behavior
- Performance problems
- Build failures
- Crashes

**Use this ESPECIALLY when:**
- Under time pressure (emergencies make guessing tempting)
- "Just one quick fix" seems obvious
- You've already tried multiple fixes
- Previous fix didn't work

**Don't skip when:**
- Issue seems simple (simple bugs have root causes too)
- You're in a hurry (rushing guarantees rework)

## The Four Phases

You MUST complete each phase before proceeding to the next.

### Phase 1: Root Cause Investigation

**BEFORE attempting ANY fix:**

1. **Read Error Messages Carefully**
   - Don't skip past errors or warnings
   - Read stack traces completely
   - Note line numbers, file paths, error codes

2. **Reproduce Consistently**
   - Can you trigger it reliably?
   - What are the exact steps?
   - If not reproducible → gather more data, don't guess

   | Platform | How to reproduce |
   |----------|-----------------|
   | **Android** | `adb logcat` for logs, reproduce on emulator, check specific API level |
   | **iOS** | Xcode console or `simctl` logs, reproduce on simulator, check iOS version |
   | **Flutter** | `flutter run` with `--verbose`, check both platforms |
   | **React Native** | Metro console, `adb logcat` / Xcode console, check native bridge errors |

3. **Check Recent Changes**
   - What changed that could cause this?
   - `git diff`, `git log --oneline -10`, `git blame <file>`
   - New dependencies, config changes

4. **Gather Evidence**

   Add diagnostic instrumentation at component boundaries:

   | Platform | Instrumentation |
   |----------|----------------|
   | **Android** | `Log.d(TAG, "value=$value")` at ViewModel, Repository, DAO boundaries |
   | **iOS** | `print("value=\(value)")` or `os_log` at ViewModel, Service, Store boundaries |
   | **Flutter** | `debugPrint("value=$value")` at BLoC/Provider, Repository, API boundaries |
   | **React Native** | `console.log("value=", value)` at component, hook, API boundaries |

5. **Trace Data Flow**
   - Where does bad value originate?
   - What called this with bad value?
   - Keep tracing up until you find the source
   - Fix at source, not at symptom

### Phase 2: Mobile-Specific Pattern Analysis

**Find the pattern before fixing:**

1. **Check common mobile bug categories:**

   **Lifecycle Bugs:**
   - Android: Activity/Fragment recreated after rotation? ViewModel cleared after process death? `onSaveInstanceState` missing?
   - iOS: View controller deallocated? `viewWillAppear` vs `viewDidLoad` timing? SceneDelegate lifecycle?
   - Flutter: Widget rebuilt unexpectedly? State lost on navigation? `dispose()` not called?
   - React Native: Component unmounted during async operation? Navigation state stale?

   **Threading Bugs:**
   - Android: Accessing UI from background thread? Coroutine on wrong dispatcher?
   - iOS: UI update not on main thread? Data race in concurrent queue? `@MainActor` missing?
   - Flutter: Isolate communication issue? `setState` after dispose?
   - React Native: Native module callback on wrong thread?

   **Memory Bugs:**
   - Android: Context leak in singleton? Inner class holding Activity reference?
   - iOS: Retain cycle in closure? Missing `[weak self]`?
   - Flutter: Stream subscription not cancelled? Controller not disposed?
   - React Native: Event listener not removed in cleanup?

   **Data Bugs:**
   - SQL/Room/CoreData: Wrong query? Migration missing? Schema mismatch?
   - SharedPreferences/UserDefaults: Wrong key? Type mismatch?
   - API: Response format changed? Null field not handled?

2. **Find Working Examples**
   - Locate similar working code in same codebase
   - What works that's similar to what's broken?

3. **Compare Against References**
   - What's different between working and broken?
   - List every difference, however small

### Phase 3: Hypothesis and Testing

1. **Form Single Hypothesis**
   - State clearly: "I think X is the root cause because Y"
   - Be specific, not vague

2. **Test Minimally**
   - Make the SMALLEST possible change to test hypothesis
   - One variable at a time

3. **Verify Before Continuing**
   - Did it work? Yes → Phase 4
   - Didn't work? Form NEW hypothesis
   - DON'T add more fixes on top

### Phase 4: Implementation

1. **Create Failing Test Case**
   - Use `mobile-tdd` skill for writing proper failing tests
   - Platform-specific test commands:

   | Platform | Test command |
   |----------|-------------|
   | **Android** | `./gradlew testDebugUnitTest --tests "com.example.MyTest"` |
   | **iOS** | `xcodebuild test -scheme <Scheme> -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyTests/MyTest` |
   | **Flutter** | `flutter test test/my_test.dart` |
   | **React Native** | `npx jest __tests__/my.test.ts` |

2. **Implement Single Fix**
   - Address the root cause identified
   - ONE change at a time
   - No "while I'm here" improvements

3. **Verify Fix**
   - Test passes now?
   - No other tests broken?
   - Issue actually resolved?

4. **If Fix Doesn't Work**
   - STOP
   - Count: How many fixes have you tried?
   - If < 3: Return to Phase 1, re-analyze with new information
   - **If ≥ 3: STOP and question the architecture**
   - Discuss with your human partner before attempting more fixes

## Red Flags - STOP and Follow Process

If you catch yourself thinking:
- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "Add multiple changes, run tests"
- "It's probably a threading issue" (without evidence)
- "I'll add a null check here" (without knowing why it's null)
- "Works on my emulator" (without checking other devices)
- **"One more fix attempt" (when already tried 2+)**

**ALL of these mean: STOP. Return to Phase 1.**

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "Issue is simple, don't need process" | Simple issues have root causes too. Process is fast for simple bugs. |
| "Emergency, no time for process" | Systematic debugging is FASTER than guess-and-check thrashing. |
| "Just try this first, then investigate" | First fix sets the pattern. Do it right from the start. |
| "I'll write test after confirming fix works" | Untested fixes don't stick. Test first proves it. |
| "This null check should handle it" | Why is it null? That's the actual bug. |
| "Works on my device" | There are thousands of device/OS combinations. |
| "One more fix attempt" (after 2+ failures) | 3+ failures = architectural problem. Question pattern, don't fix again. |

## Quick Reference

| Phase | Key Activities | Success Criteria |
|-------|---------------|------------------|
| **1. Root Cause** | Read errors, reproduce, check changes, gather evidence | Understand WHAT and WHY |
| **2. Pattern** | Check mobile categories, find working examples, compare | Identify differences |
| **3. Hypothesis** | Form theory, test minimally | Confirmed or new hypothesis |
| **4. Implementation** | Create test, fix, verify | Bug resolved, tests pass |

## Related Skills

- **`mobile-tdd`** - For creating failing test case (Phase 4, Step 1)
- **`mobile-verification`** - Verify fix worked before claiming success
