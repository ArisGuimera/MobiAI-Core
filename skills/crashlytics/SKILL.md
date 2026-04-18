---
name: crashlytics
description: Use when the user shares a Firebase Crashlytics crash link, crash ID, or asks to investigate a Crashlytics issue in depth
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, flutter, react-native]
---

# Crashlytics Investigation

Deep investigation of a crash reported through Firebase Crashlytics. Goes beyond a raw stack trace — uses Crashlytics data to understand the full context: user journey, breadcrumbs, affected versions, device info, and event frequency.

## When to Use

- User shares a Crashlytics link, crash ID, or issue title
- User asks to investigate a specific Crashlytics issue
- User mentions Firebase crashes or production crash reports from Crashlytics

## Prerequisites

This skill requires access to Firebase Crashlytics data. Check what tools are available:

1. **Firebase MCP tools** — if the Firebase MCP server is configured, use these tools:
   - `firebase_update_environment` — set the active Firebase project before querying
   - `firebase_list_apps` — get the app IDs for the project
   - `crashlytics_get_issue` — fetch issue metadata: event count, user count, error type. Parameters: `appId`, `issueId`
   - `crashlytics_list_events` — fetch crash events with full stack traces. Parameters: `appId`, `issueId`, `pageSize`
2. **Firebase CLI** — `firebase crashlytics:list` and related commands
3. **User provides the data** — if no tools are available, ask the user to paste or export the crash details from the Crashlytics console

If none of these are available, fall back to the `analyze-crash` skill with whatever information the user can provide manually.

## Workflow

### Step 1: Fetch Crash Details

Get the full crash information from Crashlytics:

- **Stack trace** — the complete symbolicated stack trace, not just the top frame
- **Event count** — how many times this crash has occurred
- **Affected users** — how many unique users have been affected
- **Affected versions** — which app versions are seeing this crash
- **First seen / last seen** — when did it start and is it still happening?
- **Device breakdown** — which devices and OS versions are most affected

### Step 2: Analyze the User Journey

Crashlytics provides breadcrumbs — the sequence of events before the crash:

- **Screen flow** — which screens did the user visit before crashing?
- **User actions** — taps, navigation events, custom log events
- **Network calls** — did a network request fail before the crash?
- **App state transitions** — did the app go to background and come back?
- **Custom keys/logs** — any custom data the app logs to Crashlytics

This context is what makes Crashlytics investigation more powerful than a raw stack trace. Use it to understand the conditions that trigger the crash.

### Step 3: Identify Patterns

Look across multiple crash events for patterns:

- Does it only happen on specific devices or OS versions?
- Does it only happen after a specific user action?
- Did it start with a specific app version? (regression)
- Is it correlated with a specific time of day or network condition?
- Check `git log` for changes introduced in the first affected version

### Step 4: Find the Root Cause in Code

With the full context from Crashlytics, trace the crash in the codebase:

1. **Start from the stack trace** — find the crashing file and method
2. **Use breadcrumbs to understand the state** — what screen was the user on? What data was loaded?
3. **Check the conditions** — what combination of state + action triggers this?
4. **Read the code path** from the triggering action through to the crash point
5. **Check recent changes** in the affected files:
   ```bash
   git log --oneline -10 -- <crashing-file>
   ```

### Step 5: Decide — Fix or Report

Based on what you found:

- **Clear root cause, simple fix** — apply the fix directly. Verify it compiles, run tests.
- **Clear root cause, complex fix** — explain the root cause, the conditions that trigger it, and propose a fix approach. Include which files need to change and why.
- **Unclear root cause** — report what you found, what patterns you identified, and what additional information would help narrow it down.

In all cases, include:
- **Root cause** — what code is wrong and why
- **Trigger conditions** — what combination of state/actions causes the crash
- **Affected scope** — which versions, devices, user flows
- **Suggested fix** — what to change, with risk assessment
