---
name: mobiai-ios-device
description: Use when interacting with an iOS Simulator — run simctl commands, capture screenshots, read device logs, manage simulators. Activate via the `Skill` tool; do not paraphrase this skill's workflow from memory.
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [ios]
---

# iOS Device Interaction

## Activation

You are reading this because `Skill(mobiai-ios-device)` was invoked — correct. Every HARD-GATE, phase checkpoint, and approval gate in this document is binding **because** the `Skill` tool was called. None of them bind if a future step, subagent, or different session reproduces this workflow from memory without another `Skill(mobiai-ios-device)` call.

If you need any of this skill's steps in another context, invoke `Skill(mobiai-ios-device)` again. Paraphrasing from memory is not activation.

> **Community contribution welcome!** This skill is a skeleton. If you have iOS development experience, please help flesh it out with battle-tested commands and workflows.

Interact with iOS Simulators using `xcrun simctl` and related tools.

## When to Use

- Reproducing a bug on an iOS Simulator
- Taking screenshots or capturing logs
- Managing Simulators

## Simulator Management

### List Available Simulators

```bash
xcrun simctl list devices available
```

### Boot a Simulator

```bash
# Boot by name
xcrun simctl boot "iPhone 15 Pro"

# Boot by UDID
xcrun simctl boot <UDID>

# Open Simulator.app (if you need the GUI)
open -a Simulator
```

### Shut Down

```bash
xcrun simctl shutdown <UDID>
xcrun simctl shutdown all  # Shut down all simulators
```

### Erase (reset to clean state)

```bash
xcrun simctl erase <UDID>
```

## App Management

### Install

```bash
xcrun simctl install booted /path/to/MyApp.app
```

### Launch

```bash
xcrun simctl launch booted com.example.myapp
```

### Terminate

```bash
xcrun simctl terminate booted com.example.myapp
```

### Uninstall

```bash
xcrun simctl uninstall booted com.example.myapp
```

## UI Inspection

### Accessibility Snapshot

```bash
# Get accessibility hierarchy (equivalent to Android's UIAutomator dump)
xcrun simctl spawn booted accessibility_inspector  # Limited CLI support

# Alternative: Use Xcode's Accessibility Inspector GUI
# Or use XCUITest for programmatic UI inspection
```

> **Note**: iOS doesn't have a direct CLI equivalent to Android's `uiautomator dump`. For automated UI interaction, consider using:
> - **XCUITest** for reliable UI automation
> - **Accessibility Inspector** (GUI) for manual inspection
> - **`xcrun simctl io booted screenshot`** combined with vision analysis

### Screenshots

```bash
xcrun simctl io booted screenshot screenshot.png
```

### Screen Recording

```bash
# Start recording
xcrun simctl io booted recordVideo recording.mp4

# Stop with Ctrl+C
```

## Logs

### View App Logs

```bash
# Stream logs
xcrun simctl spawn booted log stream --predicate 'process == "MyApp"' --level debug

# Show recent logs
xcrun simctl spawn booted log show --predicate 'process == "MyApp"' --last 5m

# Check for crashes
xcrun simctl spawn booted log show --predicate 'eventMessage contains "crash"' --last 10m
```

### Crash Reports

```bash
# Crash reports are saved to:
ls ~/Library/Logs/DiagnosticReports/ | grep MyApp
```

## UI Interaction

### Simulated Input

```bash
# Open a URL (deep link)
xcrun simctl openurl booted "myapp://screen/settings"

# Send a push notification
xcrun simctl push booted com.example.myapp notification.json

# Set device location
xcrun simctl location booted set 40.7128,-74.0060

# Add photos to simulator
xcrun simctl addmedia booted photo.jpg
```

### Keyboard Input

```bash
# Type text via pasteboard
xcrun simctl pbcopy booted <<< "Hello World"
# Then paste in the simulator
```

## Privacy & Permissions

```bash
# Grant permissions
xcrun simctl privacy booted grant photos com.example.myapp
xcrun simctl privacy booted grant camera com.example.myapp
xcrun simctl privacy booted grant location-always com.example.myapp

# Revoke permissions
xcrun simctl privacy booted revoke all com.example.myapp

# Reset all privacy settings
xcrun simctl privacy booted reset all com.example.myapp
```

## Status Bar Override

```bash
# Set clean status bar for screenshots
xcrun simctl status_bar booted override \
  --time "9:41" \
  --batteryState charged \
  --batteryLevel 100 \
  --wifiBars 3 \
  --cellularBars 4
```

## Common Issues

### Simulator won't boot
```bash
xcrun simctl shutdown all
xcrun simctl erase all
# Or delete derived data: rm -rf ~/Library/Developer/Xcode/DerivedData
```

### App crashes on launch
Check crash logs and Console.app for details.

---

**Want to improve this skill?** This is a community skeleton. Add your battle-tested workflows, common pitfalls, and advanced simctl techniques via a PR to [MobiAI-Core](https://github.com/ArisGuimera/MobiAI-Core).
