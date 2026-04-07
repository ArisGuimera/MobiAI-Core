---
name: android-device
description: ADB device interaction — emulator management, UI automation, screenshots, logcat, screen recording
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android]
---

# Android Device Interaction

Expert-level knowledge for interacting with Android devices and emulators via ADB.

## When to Use

- Reproducing a bug on an Android emulator or device
- Taking screenshots or capturing logs
- Automating UI interactions for testing
- Managing emulators

## Device Setup

### Detect Connected Devices

```bash
adb devices -l
```

This shows all connected devices and emulators. Use `-s <serial>` to target a specific device, or `-e` for the only running emulator.

### Start an Emulator

Use the helper script or run manually:

```bash
# List available AVDs
emulator -list-avds

# Start with recommended flags for automation
emulator -avd <avd-name> \
  -no-window \
  -no-audio \
  -no-boot-anim \
  -gpu swiftshader_indirect \
  -no-snapshot-load
```

Wait for boot to complete:
```bash
adb -e wait-for-device
adb -e shell getprop sys.boot_completed  # Returns "1" when ready
```

### Screen Size

```bash
adb shell wm size
# Output: Physical size: 1080x2340
```

## UI Inspection — CRITICAL

**Always dump the UI before every interaction.** This is the most important rule.

### Dump + Read UI Hierarchy (One Command)

```bash
adb shell uiautomator dump /sdcard/ui.xml && adb pull /sdcard/ui.xml ./ui-dump.xml && cat ./ui-dump.xml
```

The XML contains every visible element with:
- `text` — visible text on the element
- `resource-id` — programmatic ID (e.g., `com.example.app:id/btn_submit`)
- `content-desc` — accessibility label
- `bounds="[x1,y1][x2,y2]"` — pixel coordinates

**To find tap coordinates**: center = ((x1+x2)/2, (y1+y2)/2)

### Element Targeting Priority

1. **Text** — most reliable, search for `text="Submit"`
2. **Resource ID** — search for `resource-id=".*btn_submit"`
3. **Content description** — search for `content-desc="Submit button"`
4. **Coordinates** — last resort, computed from bounds

## UI Interaction

### Tap + Observe Pattern (USE THIS FOR EVERY TAP)

**Critical**: Every tap MUST be followed by a UI dump AND logcat check in a single command. Dialogs auto-dismiss in 2-3 seconds — if you dump UI separately, the dialog is gone.

```bash
adb shell input tap <x> <y> && sleep 1 && adb shell uiautomator dump /sdcard/ui.xml && adb pull /sdcard/ui.xml ./ui-dump.xml && cat ./ui-dump.xml && echo "=== LOGCAT ===" && adb logcat -d -t 30
```

### Other Interactions

```bash
# Back button
adb shell input keyevent 4

# Home button
adb shell input keyevent 3

# Enter/OK
adb shell input keyevent 66

# Delete/Backspace
adb shell input keyevent 67

# Swipe (scroll down)
adb shell input swipe 540 1500 540 500 300

# Long press (swipe with same start/end)
adb shell input swipe <x> <y> <x> <y> 1000
```

### Typing Text

**Never use `adb shell input text`** for text that goes into app fields — it triggers an Android "pasted from clipboard" notification that can break the app.

**Type digits using keyevents:**
| Digit | Keycode |
|-------|---------|
| 0 | 7 |
| 1 | 8 |
| 2 | 9 |
| 3 | 10 |
| 4 | 11 |
| 5 | 12 |
| 6 | 13 |
| 7 | 14 |
| 8 | 15 |
| 9 | 16 |

Example — type "1234":
```bash
adb shell input keyevent 8 9 10 11
```

For text fields where keyevents won't work (search, filters), use `input text` as a fallback.

## Logcat — Why It Matters More Than UI Dumps

Dialogs, toasts, and error popups auto-dismiss in 2-3 seconds. The UI dump may miss them entirely. But **logcat always contains the error message** that triggered the popup.

### Check for Crashes

```bash
# Check for fatal exceptions
adb logcat -d | grep -A 10 "FATAL EXCEPTION"

# Check for ANR (Application Not Responding)
adb logcat -d | grep "ANR in"

# Dump last 50 lines with timestamps
adb logcat -d -t 50
```

### Filter Logcat by App

```bash
# Get the app's PID
adb shell pidof com.example.myapp

# Filter by PID
adb logcat -d | grep <pid>
```

### Clear Logcat

```bash
adb logcat -c
```

Always clear logcat before a test sequence to get clean logs.

## Screenshots

```bash
adb shell screencap -p /sdcard/screenshot.png
adb pull /sdcard/screenshot.png ./screenshot.png
adb shell rm -f /sdcard/screenshot.png
```

## Screen Recording

```bash
# Start recording (max 180 seconds per Android limit)
adb shell screenrecord /sdcard/recording.mp4 &

# Stop recording
adb shell pkill -INT screenrecord

# Wait for finalization, then pull
sleep 3
adb pull /sdcard/recording.mp4 ./recording.mp4
```

For sessions longer than 180s, use segmented recording (start a new segment before the previous one ends).

## App Management

```bash
# Launch an app
adb shell monkey -p com.example.myapp -c android.intent.category.LAUNCHER 1

# Force stop an app
adb shell am force-stop com.example.myapp

# Clear app data
adb shell pm clear com.example.myapp

# Install APK
adb install -r app-debug.apk

# Install split APKs
adb install-multiple base.apk config.apk

# Grant permissions
adb shell pm grant com.example.myapp android.permission.ACCESS_FINE_LOCATION
```

## Handling Common Issues

### "Display over other apps" dialog
```bash
adb shell appops set <package> SYSTEM_ALERT_WINDOW allow
adb shell am force-stop com.android.settings
```

### Emulator not responding
```bash
adb emu kill  # Graceful kill
# If that fails:
taskkill /F /IM qemu-system-x86_64.exe  # Windows
pkill qemu-system-x86_64                # macOS/Linux
```

### adb server issues
```bash
adb kill-server && adb start-server
```

## Wait Times

- After tap: 1-2 seconds (included in the standard tap command)
- After screen transition: 2-3 seconds
- After network call: 3-5 seconds
- After app launch: 5-10 seconds
- After emulator boot: check `sys.boot_completed` property
