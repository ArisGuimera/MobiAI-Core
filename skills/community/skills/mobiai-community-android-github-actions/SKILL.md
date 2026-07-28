---
name: mobiai-community-android-github-actions
description: Use when creating, updating, or reviewing reusable GitHub Actions workflows for Android unit tests or instrumented tests, especially when the workflow should ask about triggers, caching, emulator setup, and action-version checks.
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android]
---

# Android GitHub Actions Workflows

Create or adapt reusable GitHub Actions workflows for Android test automation.

This skill supports two modes:

- `unittest`
- `androidtest`

It is optimized for Gradle-based Android CI, test result publishing, artifacts, and emulator-backed instrumented tests.

## When to Use

- The user wants a new GitHub Actions workflow for Android tests
- The user wants to refactor an existing Android test workflow into a reusable pattern
- The user wants to revise Gradle cache, artifact retention, triggers, or test-report publication for Android CI
- The user wants to review whether workflow actions such as report publishers or emulator runners should be updated

## Non-Negotiable Rules

1. Ask decisions **one at a time** with a recommendation.
2. Do not assume hard-coded module names, branches, paths, Gradle tasks, or emulator settings.
3. Reuse repository conventions when they exist.
4. Separate **common workflow blocks** from mode-specific blocks.
5. Before finalizing the workflow, verify whether third-party and GitHub action dependencies should stay pinned as-is or be updated.

## Workflow

### Step 1: Detect the job type

Determine whether the user needs:

- `unittest` — Gradle unit tests running on the host
- `androidtest` — instrumented tests running on one or more emulators

If unclear, ask this first.

### Step 2: Inspect the repository baseline

Search for these files before drafting the workflow:

```bash
find .github/workflows -maxdepth 1 \( -name '*.yml' -o -name '*.yaml' \) 2>/dev/null
```

### Step 3: Ask the common decisions

Ask these one at a time:

1. Trigger strategy
   - `push`
   - `pull_request`
   - both
   - `workflow_dispatch`
   - branch filters
   - path filters
2. Test scope
   - exact Gradle task
   - test class
   - package/group of tests
3. Gradle cache strategy
   - enable/disable cache
   - restore/save behavior
   - artifact retention policy
   - whether to use a **branch-specific key** so the same cache can be reused across runs on the same branch (for both Gradle and Java); ask which strategy to apply:
     - **shared key** across branches (default `gradle/actions/setup-gradle` behavior)
     - **per-branch key** (e.g. key includes `github.ref_name`) so each branch keeps its own warm cache
4. Reporting
   - publish XML results
   - upload HTML reports on failure or always
   - artifact retention days

### Step 4: Ask the `androidtest`-only decisions

If the mode is `androidtest`, also ask:

1. API level or API matrix
2. Single emulator or multiple emulators
3. Emulator target and architecture
4. AVD cache / snapshot strategy
   - AVD cache key is **always tied to the API level** (and architecture), never to the branch — emulator images and system snapshots depend on the system image, so reusing them across branches is safe and desirable
5. Emulator boot timeout
6. Emulator options
7. Whether to enable KVM permissions setup

### Step 5: Verify workflow dependencies

Before generating or updating the workflow, verify the current tags for the actions you plan to use.

Use `gh api` and check the exact repositories involved:

```bash
gh api repos/actions/checkout/releases/latest --jq '.tag_name'
gh api repos/actions/setup-java/releases/latest --jq '.tag_name'
gh api repos/actions/cache/releases/latest --jq '.tag_name'
gh api repos/actions/upload-artifact/releases/latest --jq '.tag_name'
gh api repos/gradle/actions/releases/latest --jq '.tag_name'
gh api repos/reactivecircus/android-emulator-runner/releases/latest --jq '.tag_name'
gh api repos/EnricoMi/publish-unit-test-result-action/releases/latest --jq '.tag_name'
```

Rules:

- Keep stable major tags unless the user asks for exact SHAs.
- Do not bump versions blindly; compare the workflow's current `uses:` entries against the latest stable tags.
- Prioritize dependency checks for report-publishing actions and AVD/emulator actions.
- If an update looks risky or breaking, surface it instead of silently changing it.

### Step 6: Build the workflow from reusable blocks

Always structure the generated workflow in layers:

#### Common blocks

- `name`
- `on`
- `concurrency`
- `permissions`
- checkout
- JDK setup
- Gradle setup/cache
- CI Gradle properties setup
- test result publication
- artifact upload
- cleanup / summary steps when they add value

#### `unittest`-specific blocks

- host-only Gradle execution
- unit-test XML paths
- unit-test HTML report paths
- optional manual Gradle cache restore/save strategy

#### `androidtest`-specific blocks

- KVM permissions
- AVD cache
- AVD snapshot creation
- emulator-runner execution
- instrumented-test XML paths
- instrumented-test HTML report paths

### Step 7: Output contract

When you finish, provide:

1. The workflow file content
2. A short note listing the decisions made
3. A short note listing which action dependencies were checked and whether any were updated

## Decision Tree

### If the user wants `unittest`

- Prefer a fast Ubuntu job
- Prefer host Gradle execution
- Prefer XML publication and HTML upload on failure
- Ask whether cache should use `actions/cache` explicitly or `setup-java cache: gradle` only

### If the user wants `androidtest`

- Prefer one emulator first unless the user explicitly wants a matrix
- Ask whether AVD snapshots should be cached
- Ask whether multiple API levels should run in parallel
- Verify `reactivecircus/android-emulator-runner` before reusing its version

## Output Quality Bar

- The workflow must be reusable outside the current repository
- Repo-specific defaults should stay parameterized or clearly labeled
- The final YAML must be internally consistent: trigger filters, cache keys, report paths, and Gradle commands should match each other
- Do not mix unit-test and instrumented-test report paths

## Reference Files

If you need more detail, load:

- `references/workflow-patterns.md`
