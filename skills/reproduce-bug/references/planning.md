# Reproduction Planning

How to plan reproduction steps before interacting with the device.

## Approach

1. **Parse the bug report** for explicit steps. If the description says "tap X → go to Y → crash", follow those exactly.

2. **Explore source code** to understand real UI flows:
   - Read layout XMLs, Compose screens, or SwiftUI views for the relevant feature
   - Check string resources for button labels and screen titles
   - Read ViewModels/Controllers to understand navigation flow
   - Use Grep to find relevant files by feature name or class name

3. **Generate element-based steps** with this targeting priority:
   - `elementText` — visible text on the button/element (most reliable)
   - `resourceId` / `accessibilityIdentifier` — programmatic ID
   - `contentDesc` / `accessibilityLabel` — accessibility description
   - Coordinates — only for swipes or when no other targeting is available

## Step Types

| Action | Description | Key Parameters |
|--------|-------------|----------------|
| `tap` | Tap an element | elementText, resourceId, contentDesc, x, y |
| `long_press` | Long press an element | Same as tap + durationMs |
| `swipe` | Swipe gesture | x, y, direction (up/down/left/right), durationMs |
| `type_text` | Enter text | text |
| `press_key` | Press a hardware/soft key | keycode |
| `wait` | Wait for loading/animation | durationMs |

## Non-Reproducible Bugs

Set `canReproduce=false` ONLY if:
- The bug requires physical hardware (payment terminal, printer, Bluetooth device)
- The bug is purely server-side with no triggerable UI path
- The bug depends on specific account data that can't be replicated
- The description is completely vague with no actionable information

## Wait Times

- After tap: 1-2 seconds
- After screen transition: 2-3 seconds
- After network call: 3-5 seconds
- After app launch: 5-10 seconds
