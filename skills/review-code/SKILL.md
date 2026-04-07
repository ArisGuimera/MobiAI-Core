---
name: review-code
description: Use when reviewing mobile code changes — check for lifecycle issues, memory leaks, thread safety, and platform-specific pitfalls
version: 0.1.0
license: MIT
author: Matias Rosenstein
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Review Code

Perform a mobile-specific code review, checking for common issues that static analysis tools miss.

## When to Use

- User asks for a code review of mobile code changes
- After applying a fix, to self-review before creating a PR
- User shares a diff or PR and asks for feedback

## Workflow

### Step 1: Understand the Change

1. Read the diff or changed files
2. Understand the intent — what problem does this solve?
3. Check the related issue/ticket for context

### Step 2: Review Checklist

#### Universal Checks
- [ ] **Minimal change**: Does the diff only contain what's necessary?
- [ ] **No unrelated changes**: No formatting, renaming, or refactoring mixed in?
- [ ] **Error handling**: Are errors handled appropriately (not swallowed silently)?
- [ ] **Thread safety**: Is shared mutable state properly synchronized?
- [ ] **Null safety**: Are nullable values handled correctly?
- [ ] **Edge cases**: Empty collections, null inputs, boundary values?

#### Android-Specific
- [ ] **Lifecycle awareness**: Are observers/listeners registered and unregistered properly?
- [ ] **Memory leaks**: No Activity/Fragment/Context references held in long-lived objects?
- [ ] **Main thread**: No blocking operations (network, disk I/O) on the main thread?
- [ ] **Configuration changes**: Does the code survive rotation/dark mode changes?
- [ ] **ProGuard/R8**: Are keep rules needed for reflection or serialization?
- [ ] **Backward compatibility**: Uses `@RequiresApi` or checks `Build.VERSION.SDK_INT`?

#### iOS-Specific
- [ ] **Retain cycles**: No strong reference cycles in closures? Using `[weak self]` where needed?
- [ ] **Main thread UI**: All UI updates on the main thread?
- [ ] **Optional handling**: No force unwraps (`!`) on values that could be nil?
- [ ] **Memory**: No large allocations in tight loops?
- [ ] **App lifecycle**: Handles `sceneDidEnterBackground` / `applicationWillTerminate`?

#### Flutter-Specific
- [ ] **Widget rebuilds**: No expensive operations in `build()` methods?
- [ ] **State management**: Proper use of the chosen state management solution?
- [ ] **Platform channels**: Error handling for platform-specific calls?
- [ ] **Dispose**: Controllers, streams, subscriptions disposed properly?

#### React Native-Specific
- [ ] **Re-renders**: Proper use of `useMemo`, `useCallback`, `React.memo`?
- [ ] **Native bridges**: Error handling for native module calls?
- [ ] **Navigation**: Proper cleanup in `useEffect` return?
- [ ] **Platform-specific code**: `Platform.OS` checks where needed?

### Step 3: Provide Feedback

Structure your review as:

1. **Summary**: One-line assessment (looks good / needs changes / has issues)
2. **Issues found**: Specific problems with file paths and line numbers
3. **Suggestions**: Optional improvements (clearly marked as non-blocking)

## Handling User Feedback on Your Review

When the user responds to your review:

- **Simple edits** (README, changelog, docs, test adjustments) → make the change directly
- **Code changes in the app** → re-analyze with their feedback as context, apply the fix
- **Questions** → answer with specific code references
- **Approval** → done
