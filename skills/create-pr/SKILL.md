---
name: create-pr
description: Use when the user wants to commit changes and create a pull request — proper branch, commit message, PR description with mobile context
version: 0.1.0
license: MIT
author: Matias Rosenstein
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Create PR

Create a well-structured pull request for mobile code changes.

## When to Use

- After fixing a bug and writing tests
- User asks to create a PR for their changes
- As the final step in the `fix-issue` pipeline

## Workflow

### Step 1: Prepare the Branch

1. **Create a feature branch** if not already on one:
   ```bash
   git checkout -b fix/<issue-key>
   ```

2. **Stage the changes**:
   ```bash
   git add <specific-files>  # Prefer specific files over git add -A
   ```

3. **Commit with a descriptive message**:
   ```bash
   git commit -m "fix: <short description>

   <issue-key>: <what was wrong and why>

   - Root cause: <explanation>
   - Fix: <what was changed>
   - Tests: <what tests were added>"
   ```

### Step 2: Update Changelog (if applicable)

If the project has a changelog file (CHANGELOG.md, HISTORY.md, or a section in README.md):

1. Read the changelog to understand the format
2. Add an entry under the current/unreleased version
3. Follow the project's existing style (Keep a Changelog, custom format, etc.)
4. Don't duplicate entries — check if the issue key already exists

### Step 3: Determine the Base Branch

**CRITICAL**: Do NOT assume the PR targets `main` or `master`.

1. **Ask the user** if they haven't specified: "What branch should this be merged into?"
2. If the current branch was created from another branch, use that as the base:
   ```bash
   # Check what branch this was branched from
   git log --oneline --first-parent main..HEAD | tail -1  # if few commits → likely from main
   git log --oneline --first-parent develop..HEAD | tail -1  # check develop too
   ```
3. Check the repo's default branch: `gh repo view --json defaultBranchRef`
4. If the project uses gitflow or similar (develop, release/*), respect that workflow

Use `--base <branch>` when creating the PR to target the correct branch.

### Step 4: Push and Create PR

```bash
git push -u origin fix/<issue-key>
```

Then create the PR with proper structure:

```bash
gh pr create --base <target-branch> --title "fix: <short description>" --body "$(cat <<'EOF'
## Summary
- **Issue**: <issue-key>
- **Root cause**: <what was wrong>
- **Fix**: <what was changed>

## Platform
- **Affected platform**: Android / iOS / both
- **Min SDK / deployment target affected**: <e.g. API 24, iOS 15>
- **New permissions or entitlements**: <list any, or "none">

## Changes
- `path/to/file.kt` — <what changed in this file>

## Test Plan
- [ ] Unit tests added for <what>
- [ ] Existing tests pass
- [ ] Compile check passes
- [ ] Tested on device / emulator: <model, OS version>
- [ ] Tested both orientations (portrait & landscape)
- [ ] Checked dark mode appearance
- [ ] Accessibility reviewed (TalkBack / VoiceOver, content descriptions)
- [ ] <Manual test steps if applicable>

## Regression Risk
- [ ] Touches Activity/Fragment lifecycle or ViewController lifecycle
- [ ] Touches threading / coroutines / Dispatchers / GCD
- [ ] Touches navigation graph or deep links
- [ ] Modifies ProGuard / R8 rules or iOS build settings
- [ ] Changes dependency versions
- **Risk notes**: <brief explanation of what could break>

## Screenshots / Evidence
<If visual change, add before/after screenshots or screen recordings>
EOF
)"
```

### Step 5: Verify

After creating the PR:
1. Check that CI passes (if configured)
2. Verify the diff looks correct
3. Report the PR URL to the user

## PR Title Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/) if the project uses them:
- `fix: resolve crash when opening checkout with zero amount`
- `feat: add dark mode support to settings screen`
- `refactor: extract payment logic into separate use case`

If the project doesn't use conventional commits, follow whatever pattern exists in recent PRs:
```bash
git log --oneline -20  # Check recent commit message style
```

## Commit Rules

- **Never commit secrets** (.env, `google-services.json`, `local.properties`, signing keystores, API keys). Warn the user if they ask to.
- **Prefer specific files** over `git add -A` or `git add .`
- **One logical change per commit** — don't mix unrelated changes
- **Never force push** unless the user explicitly asks
