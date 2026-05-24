---
name: mobiai-review-code
description: Use when reviewing mobile code changes — check for lifecycle issues, memory leaks, thread safety, and platform-specific pitfalls.
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Review Code

Perform a mobile-specific code review, checking for common issues that static analysis tools miss.

## When to Use

- User asks for a code review of mobile code changes
- After applying a fix, to self-review before creating a PR
- User shares a diff or PR and asks for feedback

## Pre-flight: check the Graph for code structure

The Graph indexes Kotlin/Swift symbols across the repo, so callers of any symbol modified in the diff can be located instantly. Used during Step 3 (Review Checklist) and Step 4 (Verify Test Coverage) to detect call sites and tests that the diff may have missed.

**If `<repo>/.mobiai/graph/index.json` exists**:

```bash
# Freshness — if "hace Xd" with X ≥ 1, suggest refresh (don't auto-run):
mobiai graph status

# For each symbol modified/renamed in the diff, find callers outside the diff:
mobiai graph callers <SymbolFromDiff>

# If a symbol was renamed or removed, search for stale references:
mobiai graph search <OldSymbolName>
```

The callers list surfaces call sites that the diff may not have updated — a common source of silent breakage when the PR is partial.

If the index is ≥1 day old:

> "El índice del Graph es de hace Xd. Si la rama tiene cambios recientes, considerá `mobiai graph init` antes de revisar."

Don't run `init` autonomously.

**If `.mobiai/graph/index.json` does not exist** and the project has `.kt` or `.swift` files:

> "No veo el índice del Graph. Generalo con `mobiai graph init` para que pueda detectar callers fuera del diff."

Don't run `init` autonomously.

**Skip silently when**:

- The project has no Kotlin/Swift files (Flutter or RN-only).
- The diff is purely a config / resource / build-file change.

## Workflow

### Step 1: Identify the Correct Base Branch

<STOP>
You MUST complete this step before reading ANY code or generating ANY diff.
Do NOT skip this. Do NOT assume `main` or `master`. Get the real target branch first.
</STOP>

Determine the target branch this change will be merged into:

1. **If reviewing a PR**: `gh pr view --json baseRefName --jq '.baseRefName'` — this is the authoritative answer
2. **If the user mentioned a branch**: use that branch
3. **If no PR and no branch mentioned**: ask the user "What branch is this being merged into?"
4. **Only as absolute last resort**: default to the repo's default branch via `gh repo view --json defaultBranchRef --jq '.defaultBranchRef.name'`

Then generate the diff against that branch:
```bash
BASE_BRANCH="<the branch from above>"
git diff "$BASE_BRANCH"...HEAD
```

**Why this matters**: If you diff against the wrong branch, you'll report hundreds of false changes that belong to the base branch, not to this PR. The user will see a noisy, useless review.

### Step 2: Understand the Change

1. Read the diff against the correct base branch (from Step 1)
2. Understand the intent — what problem does this solve?
3. Check the related issue/ticket for context

### Step 3: Review Checklist

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

### Step 4: Verify Test Coverage

Check if the changed code has tests. This is a separate step — do not skip it.

1. Look at the diff: identify new logic, bug fixes, or behavior changes
2. Search for existing tests related to the modified files:
   - Android: look for test classes in `src/test/` or `src/androidTest/` that test the same classes
   - iOS: look for test files in the test target that cover the same types
   - Flutter: look for `_test.dart` files matching the changed files
   - React Native: look for `.test.js` / `.test.tsx` files
3. If there are NO tests for the changed code, report it as a **blocking issue** in your review:
   - "No unit tests found for [changed class/function]. This [bug fix / new logic / behavior change] should have test coverage."
4. If tests exist but don't cover the new changes, flag it as a suggestion

Always include a "Test coverage" section in your review output, even if tests are present. State what you found.

### Step 5: Provide Feedback

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
