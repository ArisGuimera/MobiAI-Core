---
name: fix-issue
description: Full pipeline to fix a bug from an issue tracker ticket — fetch, analyze, fix, test, and create a PR
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
platforms: [android, ios, kmp, flutter, react-native]
---

# Fix Issue

End-to-end workflow for fixing a bug reported in an issue tracker (Jira, GitHub Issues, Linear, etc.).

## When to Use

- User provides an issue key/URL and asks to fix it (e.g., "fix PROJ-123", "fix #456")
- User shares a bug report and wants it resolved with a PR

## Workflow

### Step 1: Fetch the Issue

Get the full issue details — summary, description, reproduction steps, stack traces, affected versions, attachments.

**Jira**: Use the Atlassian MCP tools (`getJiraIssue`, `searchJiraIssuesUsingJql`) if available, or ask the user to paste the details.

**GitHub Issues**: Use `gh issue view <number>` via Bash.

**Manual**: If no integration, ask the user to paste the issue content.

### Step 2: Understand the Context

Before touching code:

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

### Step 3: Decide the Approach

Choose based on what information is available:

- **Has stack trace / crash signal** → go straight to code analysis (Step 4)
- **Has reproduction steps + device available** → load `reproduce-bug` skill first, then fix
- **Vague description, no stack trace** → try to reproduce if possible, else ask user for more context
- **User says "skip repro" / "just fix the code"** → go straight to code analysis

### Step 4: Find the Root Cause

1. **Start from the signal**: stack trace class names, error messages, crash signals, UI element names
2. **Search the codebase systematically**:
   ```
   Grep for: class names from stack trace, error strings, resource IDs, method names
   ```
3. **Trace the code path**: UI layer → ViewModel/Controller → Business logic → Data layer
4. **Read surrounding context**: Understand how the component works before changing it
5. **Check git history** for recent changes to the affected files:
   ```bash
   git log --oneline -10 -- <file>
   git blame <file> | head -50
   ```

### Step 5: Apply a Minimal Fix

Rules for the fix:
- Change **only** what is necessary to fix the bug
- Do NOT refactor surrounding code, rename variables, add comments, or change formatting
- Do NOT add features or improve error handling beyond what's needed
- Keep the fix as small and safe as possible
- Do NOT modify changelog files, README, or CI configuration

### Step 6: Verify the Fix Compiles

Run the appropriate compile check for the platform:

- **Android**: `./gradlew compile<Flavor>DebugKotlin --no-daemon`
- **iOS**: `xcodebuild build -scheme <scheme> -destination 'platform=iOS Simulator,name=<sim>' -quiet`
- **Flutter**: `flutter analyze && flutter build apk --debug`
- **React Native**: `npx tsc --noEmit`
- **KMP**: `./gradlew compileKotlin`

If compilation fails, fix the errors before proceeding.

### Step 7: Run Existing Tests

Run the project's unit tests to ensure no regressions:

- **Android**: `./gradlew test<Flavor>DebugUnitTest --no-daemon`
- **iOS**: `xcodebuild test -scheme <scheme> -destination 'platform=iOS Simulator,name=<sim>' -quiet`
- **Flutter**: `flutter test`
- **React Native**: `npx jest`

If tests fail, investigate whether your fix caused the failure. If so, fix the regression.

### Step 8: Write Tests

Load the `write-tests` skill and write tests that cover:
- The specific behavior that was broken and is now fixed
- Edge cases related to the fix
- Regression prevention

### Step 9: Create a PR

Load the `create-pr` skill to create a pull request with proper context.

## Decision Rules

- **NEVER stop after a single failure.** If one step fails, try the next approach.
- **Max 2 fix attempts.** If the fix fails verification twice, report to the user and ask for guidance.
- **If you can't find the cause**, explain what you investigated and ask the user for hints.
- **If the fix is too risky** (database migration, multi-module architecture changes), explain what's needed and let the user decide.

## Communication

Keep the user informed at each major step:
1. What you found in the issue
2. Your approach (repro vs code-only)
3. The root cause
4. What you changed and why
5. Test results
6. PR link
