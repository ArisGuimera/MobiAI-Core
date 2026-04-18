---
name: mobile-executing-plans
description: Use when you have a written mobile implementation plan to execute in a separate session with review checkpoints
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Executing Mobile Plans

## Overview

Load plan, review critically, execute all tasks, report when complete.

**Announce at start:** "I'm using the mobile-executing-plans skill to implement this plan."

**Note:** If subagents are available, use `mobile-subagent-development` instead of this skill — quality is significantly higher with subagent support.

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
- Announce: "I'm using the mobile-finishing-branch skill to complete this work."
- **REQUIRED:** Use `mobile-finishing-branch`
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
- **`mobile-worktrees`** - Set up isolated workspace before starting
- **`mobile-planning`** - Creates the plan this skill executes
- **`mobile-finishing-branch`** - Complete development after all tasks
