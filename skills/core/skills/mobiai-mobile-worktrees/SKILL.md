---
name: mobiai-mobile-worktrees
description: "You MUST use this before starting mobile work that needs isolation from the current workspace — new feature branches, parallel efforts, or any plan execution that should not contaminate the active tree. Do not create worktrees ad hoc; this skill enforces safe directory selection and verification."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Using Git Worktrees for Mobile

## Overview

Git worktrees create isolated workspaces sharing the same repository, allowing work on multiple branches simultaneously without switching.

**Core principle:** Systematic directory selection + safety verification = reliable isolation.

**Announce at start:** "I'm using the mobile-worktrees skill to set up an isolated workspace."

## Directory Selection Process

Follow this priority order:

### 1. Check Existing Directories

```bash
ls -d .worktrees 2>/dev/null     # Preferred (hidden)
ls -d worktrees 2>/dev/null      # Alternative
```

**If found:** Use that directory. If both exist, `.worktrees` wins.

### 2. Check CLAUDE.md

```bash
grep -i "worktree.*director" CLAUDE.md 2>/dev/null
```

**If preference specified:** Use it without asking.

### 3. Ask User

If no directory exists and no CLAUDE.md preference:

```
No worktree directory found. Where should I create worktrees?

1. .worktrees/ (project-local, hidden)
2. ~/worktrees/<project-name>/ (global location)

Which would you prefer?
```

## Safety Verification

### For Project-Local Directories

**MUST verify directory is ignored before creating worktree:**

```bash
git check-ignore -q .worktrees 2>/dev/null
```

**If NOT ignored:** Add to .gitignore, commit the change, then proceed.

### For Global Directory

No .gitignore verification needed — outside project entirely.

## Creation Steps

### 1. Detect Project Name

```bash
project=$(basename "$(git rev-parse --show-toplevel)")
```

### 2. Create Worktree

```bash
git worktree add "$path" -b "$BRANCH_NAME"
cd "$path"
```

### 3. Run Mobile Project Setup

Auto-detect and run appropriate setup:

| Platform | Setup commands |
|----------|---------------|
| **Android** | `./gradlew assembleDebug` |
| **iOS** | `pod install` (if CocoaPods) then `xcodebuild build -scheme <Scheme> -destination '...'` |
| **Flutter** | `flutter pub get` then `flutter analyze` |
| **React Native** | `npm install` then `npx pod-install` (if iOS) |
| **KMP** | `./gradlew assemble` |

**Note:** First build in a fresh worktree takes longer — Gradle cache and Xcode derived data are per-directory.

### 4. Verify Clean Baseline

Run tests to ensure worktree starts clean:

| Platform | Test command |
|----------|-------------|
| **Android** | `./gradlew testDebugUnitTest` |
| **iOS** | `xcodebuild test -scheme <Scheme> -destination '...'` |
| **Flutter** | `flutter test` |
| **React Native** | `npx jest` |

**If tests fail:** Report failures, ask whether to proceed or investigate.

**If tests pass:** Report ready.

### 5. Report Location

```
Worktree ready at <full-path>
Tests passing (<N> tests, 0 failures)
Ready to implement <feature-name>
```

## Quick Reference

| Situation | Action |
|-----------|--------|
| `.worktrees/` exists | Use it (verify ignored) |
| `worktrees/` exists | Use it (verify ignored) |
| Both exist | Use `.worktrees/` |
| Neither exists | Check CLAUDE.md → Ask user |
| Directory not ignored | Add to .gitignore + commit |
| Tests fail during baseline | Report failures + ask |

## Common Mistakes

- **Skipping ignore verification** — worktree contents get tracked
- **Assuming directory location** — follow priority: existing > CLAUDE.md > ask
- **Proceeding with failing tests** — can't distinguish new bugs from pre-existing
- **Skipping platform setup** — bare checkout won't build until dependencies installed

## Red Flags

**Never:**
- Create worktree without verifying it's ignored (project-local)
- Skip baseline test verification
- Proceed with failing tests without asking
- Assume directory location when ambiguous

## Integration

**Called by:**
- **`mobiai-mobile-brainstorming`** - When design approved and implementation follows
- **`mobiai-mobile-executing-plans-with-subagents`** - Before executing any tasks
- **`mobiai-mobile-executing-plans`** - Before executing any tasks

**Pairs with:**
- **`mobiai-mobile-finishing-branch`** - Cleans up worktree after work complete
