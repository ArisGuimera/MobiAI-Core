---
name: create-pr
description: Create a pull request with proper mobile context, test plans, and changelog updates
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
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

### Step 3: Push and Create PR

```bash
git push -u origin fix/<issue-key>
```

Then create the PR with proper structure:

```bash
gh pr create --title "fix: <short description>" --body "$(cat <<'EOF'
## Summary
- **Issue**: <issue-key>
- **Root cause**: <what was wrong>
- **Fix**: <what was changed>

## Changes
- `path/to/file.kt` — <what changed in this file>

## Test Plan
- [ ] Unit tests added for <what>
- [ ] Existing tests pass
- [ ] Compile check passes
- [ ] <Manual test steps if applicable>

## Screenshots / Evidence
<If visual change, add before/after screenshots>
EOF
)"
```

### Step 4: Verify

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

- **Never commit secrets** (.env, credentials, API keys). Warn the user if they ask to.
- **Prefer specific files** over `git add -A` or `git add .`
- **One logical change per commit** — don't mix unrelated changes
- **Never force push** unless the user explicitly asks
