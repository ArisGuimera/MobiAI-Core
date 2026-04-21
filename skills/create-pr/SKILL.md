---
name: create-pr
description: Use whenever the user expresses intent to save, share, or submit mobile changes — regardless of wording or language. Covers committing, pushing, opening a PR, or approving finished work as ready to integrate. Also triggers proactively after a fix when the user signals completion or approval (always ask before pushing). Works with any git hosting provider.
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, kmp, flutter, react-native]
---

# Create PR

Create a well-structured pull request for mobile code changes.

## When to Use

Match on **intent**, not specific words or languages. The user's phrasing is a signal, not a keyword match — read meaning in context.

**Explicit intent** (user directly expresses they want to save, share, or submit changes):
- Intent to commit the current work
- Intent to push the current branch
- Intent to open a pull request
- Instruction to "ship", "deliver", "submit", or otherwise mark work as ready
- Final step of the `fix-issue` pipeline

When intent is explicit, proceed through the workflow — but still confirm before pushing (see "Never" below).

**Implicit intent after completed work** (user approves something you just finished):
After you report a fix, review, or implementation as done, any user signal that the work is acceptable — approval, agreement, confirmation, green light, a simple "ok" in any language — is an implicit cue that the next action is probably to ship it. Do not act on implicit intent. Instead, ASK: propose creating the PR and wait for explicit confirmation. The language of your question should match the user's.

Distinguish implicit approval from continued discussion: if the user keeps asking questions, requests more changes, or is still evaluating, do not offer a PR yet.

**Never**:
- Act on implicit intent without asking — always confirm before pushing or opening a PR
- Commit directly to protected or long-lived branches (`main`, `master`, `develop`, `release/*`, `staging`) — always move to a feature branch first
- Push without explicit user confirmation
- Skip the changelog/README update step when the project has one

## Workflow

### Step 1: Prepare the Branch

1. **Check the current branch first.** If you're on a protected or long-lived branch (`main`, `master`, `develop`, `release/*`, `staging`), NEVER commit there — create a feature branch. Name it after the issue key: `fix/<issue-key>`, `feat/<issue-key>`, or `chore/<description>`:
   ```bash
   git checkout -b fix/<issue-key>
   ```

   If you've already committed to the wrong branch (e.g. a release branch), move the commit to a new feature branch before pushing:
   ```bash
   git branch fix/<issue-key>                 # save the commit under a new branch name
   git reset --hard origin/<current-branch>   # rewind the protected branch to the remote
   git checkout fix/<issue-key>               # switch to the feature branch
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

### Step 4: Push and Open the PR

Confirm with the user before pushing (project rule: never push without explicit approval).

#### Step 4.1: Identify the hosting provider

Inspect the remote to figure out which provider this repo uses:

```bash
git remote get-url origin
```

Adapt the rest of the workflow to that provider. Prefer a native CLI when one is installed and authenticated (it lets you set title, body, and base branch in one shot). Fall back to constructing the provider's web URL for creating a PR/MR when no CLI is available. You are expected to know the conventions for the major providers — if unsure, check the provider's docs rather than guessing URL formats.

#### Step 4.2: Push the branch

```bash
git push -u origin <feature-branch>
```

#### Step 4.3: Create the PR / MR

Produce the PR with this exact title format and body template, regardless of provider:

**Title**: `fix: <short description>` (or `feat:`, `refactor:`, etc. per Conventional Commits, unless the project uses a different style — check `git log` for the pattern)

**Body**:

```markdown
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
```

**Submit via CLI** if one is installed and authenticated for the provider — pass the title, body, and base branch as arguments so everything is set in a single command.

**Submit via web URL** if no CLI is available. Push the branch first, then build the provider's "new PR/MR" URL with source branch and target branch as query parameters. Give the user:
1. The clickable URL (prefills source and target)
2. The title to paste
3. The full body to paste

Do not omit the title or body in the web-URL case — most providers do not auto-populate them from commit messages, and the template is the main value this skill delivers.

### Step 5: Verify

After creating the PR:
1. Verify the diff on the PR looks correct (no unintended files, no secrets)
2. Report the PR URL to the user
3. Check CI status if the provider's CLI supports it; otherwise ask the user to confirm the PR was created and CI passed

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
