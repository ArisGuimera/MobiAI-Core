---
name: mobiai-analyze-crash
description: Use when the user shares a crash from any source (stack trace, log, screenshot, error description) and wants to find the root cause and fix it
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Analyze Crash

Analyze a crash that the user provides, find the root cause in the codebase, and fix it.

## When to Use

- User shares a stack trace, crash log, error screenshot, or any crash information
- User pastes a crash from any source (Crashlytics, Sentry, Bugsnag, logcat, Xcode console, a Slack message, an email — anything)
- User asks to investigate why something is crashing

## What You Receive

The user may give you the crash in any format:
- A raw stack trace pasted into the chat
- A screenshot of a crash
- A log file
- A description like "the app crashes when I tap X"
- A URL to a crash reporting tool
- Anything else — work with what you have

## Workflow

### Step 1: Extract the Signal

From whatever the user gave you, identify the key information:

1. **Exception / error type** — e.g., `NullPointerException`, `EXC_BAD_ACCESS`, `RangeError`, `SIGSEGV`
2. **Crash location** — file, line number, method name (if available in the stack trace)
3. **Call stack** — the sequence of calls that led to the crash
4. **Context** — what the user was doing when it happened, which screen, which action

If the information is incomplete, ask the user for more context. But don't block on it — work with what you have.

### Step 2: Classify the Crash Type

Identify the category to guide your investigation:

| Type | Common Signals | Where to Look |
|------|---------------|---------------|
| **Null reference** | NullPointerException, EXC_BAD_ACCESS, forced unwrap nil | Trace where the null value originates — usually an upstream initialization issue |
| **Concurrency** | ConcurrentModificationException, data race, random EXC_BAD_ACCESS | Shared mutable state, missing synchronization, wrong thread |
| **Index out of bounds** | IndexOutOfBoundsException, ArrayIndexOutOfBoundsException | Array/list access without bounds checking, empty collections |
| **Lifecycle** | IllegalStateException, "view not attached", "no activity" | Fragment/Activity lifecycle, accessing UI after destroy |
| **Memory** | OutOfMemoryError, jetsam kill | Leaks, large bitmap allocations, unbounded caches |
| **Network** | SocketTimeoutException, SSLException | Missing error handling in network layer |
| **Database** | SQLiteException, Core Data error | Schema migrations, thread safety, constraint violations |
| **Type casting** | ClassCastException, Swift `as!` failure | Wrong type assumptions, generics, serialization |

### Step 3: Find the Root Cause in the Codebase

For root cause investigation, follow the `mobiai-mobile-debugging` skill methodology: no guessing, trace data backward, form hypotheses based on evidence.

Starting from the crash location:

1. **Find the crashing code** — use the file and line from the stack trace:
   ```
   Grep for the class name or method name from the stack trace
   Read the file at the crash location
   ```

2. **Trace backwards** — the crash site is usually not the root cause. Follow the data flow upstream:
   - Where does the null value come from?
   - Who populates this list that's empty?
   - What thread is calling this method?

3. **Check recent changes** — the crash may have been introduced recently:
   ```bash
   git log --oneline -10 -- <crashing-file>
   git blame <crashing-file>
   ```

4. **Understand the full context** — read surrounding code, the class structure, how the component is used.

### Step 4: Decide — Fix or Propose

Based on confidence and complexity:

- **High confidence, simple fix** — apply the fix directly. Use the `mobiai-fix-issue` skill workflow: edit the code, verify it compiles, run tests.
- **Complex fix or uncertain root cause** — explain what you found, what the root cause is, and propose how to fix it. Let the user decide.
- **Can't find the root cause** — explain what you investigated, what you ruled out, and ask the user for more context.

### Step 5: If Fixing

Apply a minimal fix:
- Change **only** what is necessary
- Do NOT refactor surrounding code
- Verify it compiles
- Run existing tests to check for regressions
- If applicable, load the `mobiai-write-tests` skill to add a regression test

## Platform-Specific Patterns

### Android
- `FATAL EXCEPTION` in logcat — app process crashed
- ANR (Application Not Responding) — main thread blocked >5 seconds
- `Process: <package>` lines identify which app crashed
- StrictMode violations in debug builds hint at production issues

### iOS
- `EXC_BAD_ACCESS` — memory access violation (nil dereference, dangling pointer)
- `EXC_BREAKPOINT` — `fatalError()`, `preconditionFailure()`, Swift runtime checks
- Release builds need symbolication — `atos` or Xcode Organizer
- Crash reports in `~/Library/Logs/DiagnosticReports/`

### Flutter
- `FlutterError` — widget build errors, assertion failures
- Platform channel errors — bridge between Dart and native
- Check BOTH the Dart stack trace AND the native (Android/iOS) logs

### React Native
- Red screen errors in development
- JavaScript errors in `adb logcat | grep ReactNativeJS`
- Native crashes appear in normal Android/iOS logs — not in JS

### KMP
- Shared code crashes may appear differently per platform
- Check if the crash is in `commonMain` (shared) or platform-specific (`androidMain`/`iosMain`)
- Kotlin/Native memory model issues on iOS side
