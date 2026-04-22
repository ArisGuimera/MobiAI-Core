---
name: mobiai-mobile-executing-plans
description: "ALWAYS invoke this when a written mobile implementation plan exists and you are about to carry it out in a fresh session. Load and critique the plan before executing any task from it. Activate via the `Skill` tool; do not paraphrase this skill's workflow from memory."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Executing Mobile Plans

## Activation

You are reading this because `Skill(mobiai-mobile-executing-plans)` was invoked — correct. Every HARD-GATE, phase checkpoint, and approval gate in this document is binding **because** the `Skill` tool was called. None of them bind if a future step, subagent, or different session reproduces this workflow from memory without another `Skill(mobiai-mobile-executing-plans)` call.

If you need any of this skill's steps in another context, invoke `Skill(mobiai-mobile-executing-plans)` again. Paraphrasing from memory is not activation.

## Overview

Load plan, review critically, execute all tasks, report when complete.

**Announce at start:** "I'm using the mobiai-mobile-executing-plans skill to implement this plan."

**Note:** If subagents are available, use `mobiai-mobile-subagent-development` instead of this skill — quality is significantly higher with subagent support.

## The Process

### Step 1: Load and Review Plan
1. Read plan file
2. Review critically - identify any questions or concerns about the plan
3. If concerns: Raise them with your human partner before starting
4. If no concerns: Create TodoWrite and proceed

### Step 2: Execute Tasks

For each task:
1. Mark as in_progress
2. Follow each step exactly (plan has bite-sized steps)
3. Run verifications as specified
4. Mark as completed

### Step 3: Complete Development

After all tasks complete and verified:
- Announce: "I'm using the mobiai-mobile-finishing-branch skill to complete this work."
- **REQUIRED:** Use `mobiai-mobile-finishing-branch`
- Follow that skill to verify tests, present options, execute choice

## When to Stop and Ask for Help

**STOP executing immediately when:**
- Hit a blocker (missing dependency, test fails, instruction unclear)
- Plan has critical gaps
- You don't understand an instruction
- Verification fails repeatedly

**Ask for clarification rather than guessing.**

## Remember
- Review plan critically first
- Follow plan steps exactly
- Don't skip verifications
- Stop when blocked, don't guess
- Never start implementation on main/master branch without explicit user consent

## Integration

**Required workflow skills:**
- **`mobiai-mobile-worktrees`** - Set up isolated workspace before starting
- **`mobiai-mobile-planning`** - Creates the plan this skill executes
- **`mobiai-mobile-finishing-branch`** - Complete development after all tasks
