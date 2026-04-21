# Fix Issue — Workflow Rules

Detailed decision logic for the fix-issue pipeline.

## Multiple Issues in a Single Request

When the user sends multiple issues (e.g., "fix PROJ-123 and PROJ-456"):

1. **Determine versions/branches** for each issue first
2. **If they share the same version**, build once and reuse
3. **Work each issue one at a time, in order**:
   - Fetch issue
   - Analyze/reproduce
   - Fix
   - Verify
   - Commit and PR
   - Move to next issue
4. **Never skip an issue.** Even if one fails, move on to the next.
5. **Report all results at the end** with links to all PRs created.

## Version/Branch Handling

Before building, determine the correct branch:

1. **User's message** — e.g., "fix PROJ-123 on v3.2" → checkout v3.2 branch
2. **Issue metadata** — check affected versions field
3. **Ask the user** — if neither source has version info

## Code-Only Mode (No Device Reproduction)

Skip build and device interaction when:
- User explicitly says "no repro", "just look at the code", "skip reproduction"
- No device/emulator/simulator is available
- The bug has a clear stack trace that points to the exact code

Flow becomes: fetch issue → search codebase → fix → test → PR

## Handling Build Failures

If the build fails:
- Report the error to the user
- **Skip directly to code analysis** — you can still read and fix code without a running app
- Do NOT stop the entire workflow

## Handling User Feedback After PR

When the user gives feedback on the PR:
- **Changelog/README fix** → edit the file directly, commit, push
- **Tests missing or failing** → write/fix tests, commit, push
- **Code change requested** → re-analyze with the feedback as context, apply fix
- **Approval** → done

## Budget Recovery

If the analysis runs out of context or turns but code changes were already applied:
- Check `git diff --name-only` to identify changed files
- Recover the partial work rather than losing it
- Report what was done and what remains
