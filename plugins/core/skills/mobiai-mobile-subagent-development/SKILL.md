---
name: mobiai-mobile-subagent-development
description: "ALWAYS invoke this when executing a mobile implementation plan in the current session with subagent support available. Drives fresh-subagent-per-task execution with mandatory two-stage review; do not hand-execute plan tasks in-session when this workflow applies."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Mobile Subagent-Driven Development

Execute plan by dispatching fresh subagent per task, with two-stage review after each: spec compliance review first, then code quality review.

**Why subagents:** Fresh context per task prevents pollution. Precise instructions ensure focus. This preserves your own context for coordination.

**Core principle:** Fresh subagent per task + two-stage review (spec then quality) = high quality, fast iteration

## When to Use

- Have implementation plan from `mobiai-mobile-planning`
- Tasks mostly independent
- Want to stay in this session

**vs. mobiai-mobile-executing-plans:** Same session, fresh subagent per task, two-stage review, faster iteration.

## The Process

```
[Read plan, extract all tasks with full text, create TodoWrite]

For each task:
  → Dispatch implementer subagent
  → Answer questions if any
  → Implementer implements, tests, commits, self-reviews
  → Dispatch spec reviewer subagent
  → If gaps: implementer fixes, spec reviewer re-reviews
  → Dispatch code quality reviewer subagent
  → If issues: implementer fixes, quality reviewer re-reviews
  → Mark task complete in TodoWrite

After all tasks:
  → Dispatch final code reviewer for entire implementation
  → Use mobiai-mobile-finishing-branch
```

## Model Selection

**Mechanical tasks** (isolated functions, clear specs, 1-2 files): fast, cheap model.

**Integration tasks** (multi-file coordination, debugging): standard model.

**Architecture and review tasks**: most capable model.

## Handling Implementer Status

**DONE:** Proceed to spec compliance review.

**DONE_WITH_CONCERNS:** Read concerns. If about correctness, address before review. If observations, note and proceed.

**NEEDS_CONTEXT:** Provide missing context and re-dispatch.

**BLOCKED:** Assess: context problem → provide more. Task too large → break into pieces. Plan wrong → escalate to human.

**Never** ignore an escalation or force retry without changes.

## Prompt Templates

**Implementer prompt:**
```
Your task: [exact task text from plan]

Context:
- Platform: [Android/iOS/Flutter/etc.]
- Architecture: [patterns used in project]
- Relevant files: [list]

Instructions:
1. Implement exactly what the task describes
2. Write tests first (TDD: test → fail → implement → pass)
3. Run tests and verify they pass
4. Commit your changes
5. Self-review: completeness, quality, testing
6. Report: Status | What implemented | Test results | Files changed | Issues

Build command: [platform-specific]
Test command: [platform-specific]
```

**Spec reviewer prompt:**
```
Review this implementation for spec compliance.

Task spec: [original task text]
Files changed: [list from implementer]

Check:
- Does implementation match spec exactly?
- Anything missing from requirements?
- Any extra work not in spec? (scope creep)
- Tests covering specified behavior?

Report: Spec compliant or Issues found (with specifics)
```

**Code quality reviewer prompt:**
```
Review this implementation for code quality.

Files changed: [list]
Platform: [Android/iOS/Flutter/etc.]

Check:
- Mobile-specific: lifecycle handling, thread safety, memory management
- Architecture: follows existing patterns
- Tests: meaningful assertions, not just coverage
- Naming: clear, consistent with project conventions

Report: Strengths, Issues (Critical/Important/Minor), Assessment
```

## Red Flags

**Never:**
- Start implementation on main/master branch without explicit user consent
- Skip reviews (spec compliance OR code quality)
- Proceed with unfixed issues
- Dispatch multiple implementation subagents in parallel (conflicts)
- Make subagent read plan file (provide full text instead)
- Accept "close enough" on spec compliance
- **Start code quality review before spec compliance passes**

**If reviewer finds issues:**
- Implementer fixes them
- Reviewer reviews again
- Repeat until approved

## Integration

**Required workflow skills:**
- **`mobiai-mobile-worktrees`** - Set up isolated workspace before starting
- **`mobiai-mobile-planning`** - Creates the plan this skill executes
- **`mobiai-mobile-finishing-branch`** - Complete development after all tasks

**Subagents should use:**
- **`mobiai-mobile-tdd`** - Subagents follow TDD for each task
