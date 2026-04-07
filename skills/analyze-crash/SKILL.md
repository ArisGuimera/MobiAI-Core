---
name: analyze-crash
description: Analyze crash logs, stack traces, and Crashlytics reports to identify root causes in mobile apps
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
platforms: [android, ios, kmp, flutter, react-native]
---

# Analyze Crash

Systematic approach to analyzing mobile app crashes from logs, stack traces, or crash reporting tools.

## When to Use

- User shares a stack trace, crash log, or crash reporting URL
- A reproduction attempt captured a crash signal
- User asks to investigate a production crash

## Workflow

### Step 1: Gather Crash Data

Collect all available information:

1. **Stack trace** — the most valuable signal. Identify:
   - Exception class (e.g., `NullPointerException`, `EXC_BAD_ACCESS`, `RangeError`)
   - Crash location (file, line, method)
   - Full call stack (trace the path from trigger to crash)

2. **Crash reporting tools** — if available:
   - **Firebase Crashlytics**: Use Firebase MCP tools to fetch crash details, event counts, affected users
   - **Sentry**: Check the Sentry dashboard for breadcrumbs and device context
   - **Bugsnag/AppCenter**: Similar — look for breadcrumbs, device info, app version

3. **Device logs**:
   - **Android**: `adb logcat -d | grep -A 20 "FATAL EXCEPTION"`
   - **iOS**: Check Console.app or `xcrun simctl spawn booted log show --predicate 'process == "<AppName>"' --last 5m`

4. **Context from the issue** — reproduction steps, affected versions, user actions before crash

### Step 2: Classify the Crash

| Type | Signal | Approach |
|------|--------|----------|
| **Null reference** | NullPointerException, EXC_BAD_ACCESS, nil unwrap | Trace where the null value originates |
| **Concurrency** | ConcurrentModificationException, data race, EXC_BAD_ACCESS on collection | Look for shared mutable state, missing synchronization |
| **Out of bounds** | IndexOutOfBoundsException, ArrayIndexOutOfBoundsException | Check array/list access patterns, empty collections |
| **Lifecycle** | IllegalStateException after onSaveInstanceState, view not attached | Check Fragment/Activity lifecycle handling |
| **Memory** | OutOfMemoryError, jetsam | Look for leaks, large allocations, bitmap handling |
| **Network** | SocketTimeoutException, SSLException | Check error handling in network layer |
| **Database** | SQLiteException, Core Data error | Check migrations, schema mismatches, thread safety |
| **Type casting** | ClassCastException, as! failure | Check type assumptions, generics |

### Step 3: Trace the Code Path

Starting from the crash location in the stack trace:

1. **Read the crashing method** — understand what it's doing at the crash point
2. **Trace backwards** through the call stack — how did we get here?
3. **Identify the root cause** — usually it's not at the crash site but upstream:
   - A variable that should have been initialized but wasn't
   - A state that should have been checked but wasn't
   - A threading issue where state was mutated from the wrong thread
4. **Check recent changes** to the affected files:
   ```bash
   git log --oneline -10 -- <crashing-file>
   git blame <crashing-file> | grep -A 2 -B 2 "<line-number>"
   ```

### Step 4: Assess Severity

| Severity | Criteria | Action |
|----------|----------|--------|
| **Critical** | >50 affected users, data loss, security issue | Fix immediately |
| **High** | >10 affected users, core feature broken | Fix in current sprint |
| **Medium** | 5-10 affected users, workaround exists | Schedule for next sprint |
| **Low** | <5 affected users, edge case | Monitor and fix when possible |

### Step 5: Report Findings

Provide:
- **Root cause**: What code is wrong and why
- **Impact**: Who is affected and how severely
- **Recommended fix**: What should change (high-level)
- **Risk assessment**: Is the fix safe? Could it introduce regressions?

## Platform-Specific Crash Patterns

### Android
- `FATAL EXCEPTION` in logcat — the app process crashed
- ANR (Application Not Responding) — main thread blocked for >5 seconds
- `Process: <package>` lines identify which app crashed
- StrictMode violations in debug builds indicate potential production issues

### iOS
- `EXC_BAD_ACCESS` — memory access violation (often nil dereference or dangling pointer)
- `EXC_BREAKPOINT` — triggered by `fatalError()`, `preconditionFailure()`, or Swift runtime checks
- Crash reports in `~/Library/Logs/DiagnosticReports/`
- Symbolication needed for release builds — use `atos` or Xcode Organizer

### Flutter
- `FlutterError` — widget build errors, assertion failures
- Platform channel errors — bridge between Dart and native code
- Check both Dart stack trace AND native (Android/iOS) logs

### React Native
- Red screen errors in development
- JavaScript errors in `adb logcat | grep ReactNativeJS`
- Native crashes may appear in normal Android/iOS logs
