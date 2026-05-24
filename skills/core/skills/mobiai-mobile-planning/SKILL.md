---
name: mobiai-mobile-planning
description: "You MUST use this whenever a mobile task spans multiple steps, files, or subsystems — before touching code. Produce a written plan the user approves; do not improvise multi-step work from memory."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Writing Mobile Plans

## Overview

Write comprehensive implementation plans assuming the engineer has zero context for our codebase. Document everything they need: which files to touch, code, testing, how to verify. Give them the whole plan as bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

**Announce at start:** "I'm using the mobile-planning skill to create the implementation plan."

**Save plans to:** `docs/plans/YYYY-MM-DD-<feature-name>.md`
- (User preferences for plan location override this default)

## Numbered Task Lists Are INPUT, Not Substitute

If the user hands you a numbered list of 3+ tasks, that list is raw input — not a finished plan. You MUST still invoke this skill to validate:

- **Dependencies**: which task must finish before another can start
- **File-level conflicts**: can tasks be parallelized without agent collisions (e.g. two agents editing the same DI module or build file)
- **Scope boundaries**: what each task pulls in (a "fix bug X" task may require related tests and touch adjacent code)
- **Approval checkpoint**: present the dependency graph and wait for explicit user approval before dispatching any agent or writing any code

A user-supplied list describes *what*. Planning determines *how, in what order, and by whom*.

## Scope Check

If the spec covers multiple independent subsystems, suggest breaking into separate plans — one per subsystem. Each plan should produce working, testable software on its own.

## File Structure

Before defining tasks, map out which files will be created or modified. Design units with clear boundaries. Prefer smaller, focused files. In existing codebases, follow established patterns.

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Plan Document Header

**Every plan MUST start with this header:**

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** Use `mobiai-mobile-executing-plans-with-subagents` (recommended) or `mobiai-mobile-executing-plans` to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

**Platform:** [Android / iOS / KMP / Flutter / React Native]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.kt`
- Modify: `exact/path/to/existing.kt:123-145`
- Test: `exact/path/to/test.kt`

- [ ] **Step 1: Write the failing test**

```kotlin
@Test
fun `specific behavior description`() {
    val result = function(input)
    assertEquals(expected, result)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `./gradlew testDebugUnitTest --tests "com.example.MyTest"`
Expected: FAIL with "function not defined"

- [ ] **Step 3: Write minimal implementation**

```kotlin
fun function(input: Type): ReturnType {
    return expected
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `./gradlew testDebugUnitTest --tests "com.example.MyTest"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add tests/path/test.kt src/path/file.kt
git commit -m "feat: add specific feature"
```
````

## Mobile-Specific Considerations

Include these in the plan when relevant:

| Consideration | Include when |
|---------------|-------------|
| **Build verification** | After every 2-3 tasks: `./gradlew assembleDebug` or `xcodebuild build` |
| **DB migrations** | Any Room/CoreData schema change — include migration code |
| **Permissions** | New permission needed — include manifest/plist entry + runtime request |
| **ProGuard/R8 rules** | New serialization or reflection — include keep rules |
| **Multi-platform** | KMP/Flutter/RN — separate tasks for shared code vs platform-specific |
| **Navigation** | New screen — include route registration and navigation call |
| **DI registration** | New injectable class — include module/component registration |

## No Placeholders

Every step must contain the actual content. These are **plan failures** — never write them:
- "TBD", "TODO", "implement later", "fill in details"
- "Add appropriate error handling" / "add validation" / "handle edge cases"
- "Write tests for the above" (without actual test code)
- "Similar to Task N" (repeat the code)
- Steps that describe what to do without showing how

## Remember
- Exact file paths always
- Complete code in every step
- Exact commands with expected output
- DRY, YAGNI, TDD, frequent commits

## Self-Review

After writing the complete plan, check:

**1. Spec coverage:** Skim each requirement. Can you point to a task that implements it? List any gaps.

**2. Placeholder scan:** Search for red flags from the "No Placeholders" section. Fix them.

**3. Type consistency:** Do types, method signatures, and property names match across tasks?

If you find issues, fix them inline.

## Execution Handoff

After saving the plan, offer execution choice:

**"Plan complete and saved to `<path>`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks

**2. Inline Execution** - Execute tasks in this session with checkpoints

**Which approach?"**

**If Subagent-Driven chosen:** Use `mobiai-mobile-executing-plans-with-subagents`

**If Inline Execution chosen:** Use `mobiai-mobile-executing-plans`
