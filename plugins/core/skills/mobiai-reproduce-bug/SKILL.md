---
name: mobiai-reproduce-bug
description: Use when the user wants to reproduce a bug on a mobile device, emulator, or simulator using UI automation
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android, ios, flutter, react-native]
---

# Reproduce Bug

Interact with a running mobile app on a device/emulator/simulator to reproduce a reported bug.

## When to Use

- Before fixing a bug, to confirm it exists and capture evidence
- To verify a fix by re-running the reproduction steps
- When a user says "try to reproduce this on the emulator"

## Two Modes

### Reproduce Mode
Explore the app to find and trigger the bug. You have freedom to navigate, try different paths, and investigate.

### Verify Mode
After a fix is applied, re-run the **exact same steps** from the original reproduction. Only check if the original bug still occurs — don't explore other features or treat unrelated behavior as bugs.

## Workflow

### Step 1: Detect Platform

Check which platform the project targets and load the appropriate device skill:
- Android → load `mobiai-android-device` skill
- iOS → load `mobiai-ios-device` skill
- Flutter → load `mobiai-android-device` or `mobiai-ios-device` depending on the target platform
- React Native → load `mobiai-android-device` or `mobiai-ios-device` depending on the target platform

### Step 2: Plan Reproduction Steps

Before interacting with the device:

1. **Read the bug report** — extract explicit reproduction steps if provided
2. **Explore source code** to understand the app's navigation:
   - Find the relevant screen/feature in the source code
   - Read layout files, string resources, or Compose/SwiftUI views
   - Understand what setup or data the feature needs
3. **Generate a step-by-step plan** with element-based targeting:
   - Prefer targeting by text label, resource ID, or accessibility label
   - Use coordinates only as a fallback for swipe gestures

### Step 3: Execute on Device

Follow these mandatory rules:

1. **Dump UI before every action.** Always inspect the screen state before interacting.
2. **One action at a time.** Every tap/swipe must be followed by a UI dump to see the result.
3. **Never guess coordinates.** Always read them from the UI hierarchy dump.
4. **Check logs after every action.** Transient dialogs and errors may not appear in UI dumps but will be in the device logs.
5. **Handle unexpected screens.** Dismiss dialogs, handle permission requests, deal with loading states.

### Step 4: Capture Evidence

After each significant action, capture:
- **UI state** — dump the UI hierarchy
- **Device logs** — check for crashes, exceptions, error messages
- **Screenshots** — if the bug is visual

### Step 5: Report Results

Return a clear report:
- **Reproduced / Not reproduced**
- **Step-by-step narrative** of what you did and what you saw
- **Crash signal** (if a crash was observed) — exception class, error message

## When You're Blocked

- **Never retry the same action more than twice.** If something fails, try an alternative path.
- **If still blocked after 2 attempts**, stop and report what happened.
- **Focus on reproducing, not fixing.** If you spot an obvious root cause while reproducing, note it — but don't get sidetracked into a code investigation. Finish the reproduction first.

## App Crash on Startup

If the app crashes before any screen loads:
1. Capture the crash from device logs
2. Report the crash signal and stack trace
3. Then proceed to investigate the root cause in the code using the `mobiai-analyze-crash` skill

## Platform-Specific Details

### Android
Load the `mobiai-android-device` skill for:
- ADB commands (tap, swipe, type, keyevent)
- UIAutomator UI dumps
- Logcat crash detection
- Screenshot capture

### iOS
Load the `mobiai-ios-device` skill for:
- simctl commands
- Accessibility snapshot for UI hierarchy
- Console log monitoring
- Screenshot capture

### Flutter
Flutter apps run on Android emulators and iOS simulators. Load the `mobiai-android-device` or `mobiai-ios-device` skill depending on your target platform. Additionally, check Flutter-specific logs with `flutter logs` to capture framework-level errors and widget rebuild information that may not appear in platform-native logs.

### React Native
React Native apps run on Android emulators and iOS simulators. Load the `mobiai-android-device` or `mobiai-ios-device` skill depending on your target platform. On Android, check JS-layer errors with `adb logcat | grep ReactNativeJS` to capture JavaScript exceptions and bridge errors that may not surface in the UI hierarchy.
