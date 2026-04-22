---
name: mobile-debugging
description: "You MUST use this before proposing any fix for a mobile bug, test failure, crash, or unexpected behavior. Walks root-cause analysis as phased evidence gathering. For small fixes with clear evidence, runs autonomously and returns the root cause to the caller. For complex or uncertain cases, gates at known-vs-assumed and root-cause steps."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Systematic Mobile Debugging

Random fixes waste time and create new bugs. Quick patches mask underlying issues.

**Core principle:** ALWAYS find root cause before attempting fixes. Symptom fixes are failure.

<HARD-GATE>
You MUST NOT propose any fix, write any code, edit any production file, refactor any surrounding code, or invoke `mobile-tdd` until you have (a) gathered evidence, (b) separated what is known from what is assumed, (c) formed and verified a hypothesis, and (d) stated the root cause.

Whether you also need **explicit user confirmation** at steps (b) and (d) depends on the scope classification inherited from the caller (or declared at Pre-flight below). Fast path: no confirmation gate, proceed autonomously. Gated path: explicit confirmation required at both steps.
</HARD-GATE>

## Anti-Pattern: "It's Obviously X"

The most common failure mode is pattern-matching the symptom to a familiar-looking cause and proposing a fix before the evidence supports it — especially when a sub-agent or exploration returns a confident diagnosis that was never cross-checked against the code. Threading, null, lifecycle, cache, "the backend changed" — all tempting guesses. Every bug goes through the phased flow.

The scope classification only changes **whether you wait for user confirmation**, not **whether you do the analysis**. Fast path still gathers evidence, splits known vs. assumed, forms and verifies a hypothesis, and states the root cause — it just returns that result to the caller instead of pausing for approval.

## The Iron Law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

## Pre-flight: Scope Classification

This skill inherits the scope from its caller (typically `fix-issue`). If invoked standalone, classify at the start:

**Fast path** — work autonomously through all phases; return the verified root cause to the caller at Phase 6:
- Clear symptom with obvious evidence (stack trace points at a specific line, log message is explicit)
- Single root cause category (not contradictory evidence)
- No architectural question, no cross-module mystery

**Gated path** — stop and wait for user confirmation at Phase 2 and Phase 5:
- Contradictory evidence or multiple plausible causes
- Cross-module / cross-platform investigation required
- Root cause involves architectural choices (e.g., "should this be on the main thread?")
- Previous fix attempts failed

In doubt → gated path.

## Checklist

You MUST create a task for each of these items and complete them in order. Do not skip phases. Do not merge phases.

1. **Gather evidence** — logs, stack traces, reproduction steps, affected platforms, recent changes
2. **State known vs. assumed** — label each fact as observed or inferred (gated path: present to the user and wait for confirmation; fast path: record the split internally and proceed)
3. **Form a hypothesis** — a single, specific, testable theory about the root cause
4. **Verify the hypothesis** — against code and evidence, not against intuition
5. **Present the root cause** (gated path: wait for user confirmation before Phase 6; fast path: state the root cause and proceed to Phase 6)
6. **Hand off to implementation** — return control to `fix-issue` if the broader pipeline is active, or invoke `mobile-tdd`

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
- A previous fix didn't work
- A sub-agent or exploration returned a confident diagnosis

**Don't skip when:**
- Issue seems simple (simple bugs have root causes too)
- You're in a hurry (rushing guarantees rework)

## Phase 1: Gather Evidence

Before forming any opinion, collect the raw material.

1. **Read error messages carefully**
   - Don't skip past errors or warnings
   - Read stack traces completely
   - Note line numbers, file paths, error codes

2. **Confirm reproducibility**
   - Can you trigger it reliably?
   - What are the exact steps?
   - If not reproducible → gather more data, don't guess

   | Platform | How to reproduce |
   |----------|-----------------|
   | **Android** | `adb logcat` for logs, reproduce on emulator, check specific API level |
   | **iOS** | Xcode console or `simctl` logs, reproduce on simulator, check iOS version |
   | **Flutter** | `flutter run` with `--verbose`, check both platforms |
   | **React Native** | Metro console, `adb logcat` / Xcode console, check native bridge errors |
   | **KMP** | Reproduce on each target platform; check whether the issue is in shared code or platform-specific `actual` implementations |

3. **Check recent changes & find the breaking commit**
   - What changed that could cause this?
   - `git log --oneline -20 -- <affected-file>` — commits that touched the affected file
   - `git blame -L <line>,<line> <file>` — which commit last changed the specific failing line
   - `git log --all --grep="<keyword>"` — past related fixes that might show the introduction of a conflicting pattern
   - **If this is a regression**: identify the commit that introduced the bug. Record its SHA and subject. Use `git bisect` when the history is dense or the search space is large. The breaking commit goes into the PR description (it helps the reviewer assess scope and decide backport priority).
   - **If the bug has always existed** (not a regression): state that explicitly. "Always-existed, not a regression" is valid and important information for the reviewer — don't leave it ambiguous.
   - New dependencies, config changes, Gradle/SPM/pubspec bumps can themselves be the breaking commit.

4. **Collect concrete evidence**

   Add diagnostic instrumentation at component boundaries:

   | Platform | Instrumentation |
   |----------|----------------|
   | **Android** | `Log.d(TAG, "value=$value")` at ViewModel, Repository, DAO boundaries |
   | **iOS** | `print("value=\(value)")` or `os_log` at ViewModel, Service, Store boundaries |
   | **Flutter** | `debugPrint("value=$value")` at BLoC/Provider, Repository, API boundaries |
   | **React Native** | `console.log("value=", value)` at component, hook, API boundaries |
   | **KMP** | Log at the shared/expect boundary AND at each `actual` implementation |

5. **Identify affected platforms precisely**
   - Does it happen on Android only? iOS only? Both? Specific OS versions? Specific devices?
   - Precise platform scope often points directly at the root cause category.

## Phase 2: State Known vs. Assumed

Separate what you actually observed from what you are inferring. List in two columns:

- **Known** (direct evidence): log lines seen, stack traces captured, reproduction confirmed, git diff contents
- **Assumed** (not yet verified): "probably caused by X", "likely a threading issue", "the backend probably returned null"

Every item in "Assumed" is a candidate to either verify (moving it to Known) or to drop.

- **Fast path**: record the split internally, proceed directly to Phase 3.
- **Gated path**: present the split to the user and wait for confirmation before Phase 3. If the user corrects an assumption, update the list.

## Phase 3: Form a Hypothesis

One hypothesis at a time, stated specifically. Not "it's a race condition" — rather "the ViewModel's `observeUser()` coroutine is cancelled by the Activity being recreated on rotation, so the subsequent emission is dropped".

Check common mobile bug categories to frame the hypothesis:

**Lifecycle bugs:**
- Android: Activity/Fragment recreated after rotation? ViewModel cleared after process death? `onSaveInstanceState` missing?
- iOS: View controller deallocated? `viewWillAppear` vs `viewDidLoad` timing? SceneDelegate lifecycle?
- Flutter: Widget rebuilt unexpectedly? State lost on navigation? `dispose()` not called?
- React Native: Component unmounted during async operation? Navigation state stale?

**Threading bugs:**
- Android: Accessing UI from background thread? Coroutine on wrong dispatcher?
- iOS: UI update not on main thread? Data race in concurrent queue? `@MainActor` missing?
- Flutter: Isolate communication issue? `setState` after dispose?
- React Native: Native module callback on wrong thread?

**Memory bugs:**
- Android: Context leak in singleton? Inner class holding Activity reference?
- iOS: Retain cycle in closure? Missing `[weak self]`?
- Flutter: Stream subscription not cancelled? Controller not disposed?
- React Native: Event listener not removed in cleanup?

**Data bugs:**
- SQL/Room/CoreData: Wrong query? Migration missing? Schema mismatch?
- SharedPreferences/UserDefaults: Wrong key? Type mismatch?
- API: Response format changed? Null field not handled?

**KMP-specific:**
- Divergence between `actual` implementations across platforms
- Freezing/threading models in legacy Kotlin/Native
- Serialization differences between JVM and Native

State the single hypothesis in the form: "I think X is the root cause because Y". Be specific.

## Phase 4: Verify the Hypothesis

Verify against code and evidence, not against intuition.

1. **Find working examples**
   - Locate similar code in the same codebase that works
   - What is different between the working and broken paths?
   - List every difference, however small.

2. **Test minimally**
   - Make the SMALLEST possible change that would confirm or refute the hypothesis (a log line, a breakpoint, a one-line guard used only as a probe — not as a fix)
   - One variable at a time

3. **Read the code path carefully**
   - Trace from the entry point down to where the bad value appears
   - Fix at the source, not at the symptom
   - If the hypothesis contradicts the code, the hypothesis is wrong — do not bend the evidence to match

4. **If verification fails**
   - Discard the hypothesis
   - Return to Phase 3 and form a new one
   - Do NOT stack fixes on top of an unverified theory

## Phase 5: Present Root Cause

Once the hypothesis is verified, state:

1. **The root cause** — one or two sentences, stated plainly
2. **The evidence that proves it** — specific log lines, specific code locations, specific reproduction results
3. **Why the symptom follows from the cause** — the causal chain from root cause to observed behavior
4. **What is still uncertain**, if anything

- **Fast path**: state the root cause concisely and proceed to Phase 6 without pausing. If you discover during Phase 4 that the evidence is contradictory or the hypothesis doesn't hold cleanly, upgrade to gated path before presenting.
- **Gated path**: ask explicitly:

  > "This is the root cause I've identified and verified. Do you confirm this is correct before I propose a fix?"

  Wait for the user's explicit confirmation. If the user pushes back or points to evidence you missed, return to Phase 3 or Phase 4 as appropriate.

## Phase 6: Hand Off to Implementation

The fix itself is not proposed inside this skill. `mobile-debugging` ends when the root cause is stated (fast path) or confirmed (gated path).

- If invoked from `fix-issue`, return control — `fix-issue`'s Phase 4 (Propose the Fix) takes over with its own gating rules.
- Otherwise, **invoke skill `mobile-tdd`** to write the failing test first, then the implementation.

## Red Flags — STOP and Follow Process

If you catch yourself thinking:
- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "Add multiple changes, run tests"
- "It's probably a threading issue" (without evidence)
- "I'll add a null check here" (without knowing why it's null)
- "Works on my emulator" (without checking other devices)
- "The sub-agent said it's X, I'll go with that" (without verifying against code)
- **"One more fix attempt" (when already tried 2+)**

**ALL of these mean: STOP. Return to Phase 1.**

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "Issue is simple, don't need process" | Simple issues have root causes too. The phased flow is fast for simple bugs. |
| "Emergency, no time for process" | Systematic debugging is FASTER than guess-and-check thrashing. |
| "Just try this first, then investigate" | First fix sets the pattern. Do it right from the start. |
| "I'll write tests after confirming the fix works" | Untested fixes don't stick. Test first proves it. |
| "This null check should handle it" | Why is it null? That's the actual bug. |
| "Works on my device" | There are thousands of device/OS combinations. |
| "The Explore agent already diagnosed it" | Sub-agents can be confidently wrong. Verify against the code. |
| "One more fix attempt" (after 2+ failures) | 3+ failures = architectural problem. Question the pattern, don't fix again. |

## Quick Reference

| Phase | Key Activity | Fast path | Gated path |
|-------|--------------|-----------|------------|
| **1. Gather evidence** | Read errors, reproduce, check changes, instrument | Proceed autonomously | Proceed autonomously |
| **2. Known vs. assumed** | Split facts from inferences | Record internally, proceed | Present to user, wait for confirmation |
| **3. Hypothesis** | Single specific testable theory | Proceed autonomously | Proceed autonomously |
| **4. Verify** | Compare to working code, test minimally | Proceed autonomously | Proceed autonomously |
| **5. Root cause** | Present cause + evidence + causal chain | State and proceed to Phase 6 | Ask user to confirm before Phase 6 |
| **6. Hand off** | Return to `fix-issue` or invoke `mobile-tdd` | Implementation begins elsewhere | Implementation begins elsewhere |

## Related Skills

- **`fix-issue`** — the broader pipeline; calls this skill for its Phase 3. The scope classification (fast/gated) declared in `fix-issue` is inherited by this skill.
- **`mobile-tdd`** — for creating the failing test once the root cause is confirmed
- **`mobile-verification`** — to verify the fix worked before claiming success
- **`reproduce-bug`** — for driving a device/emulator/simulator to reproduce the bug in Phase 1 (use only when `fix-issue` Phase 2 authorized reproduction; do not invoke autonomously)
