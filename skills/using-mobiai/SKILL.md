---
name: using-mobiai
description: Bootstrap skill loaded on every session — teaches the agent about MobiAI skills, agents, and capabilities
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
---

# MobiAI — Mobile Development Ecosystem

You have access to MobiAI, an ecosystem of skills, agents, and automation tools for mobile development. MobiAI gives you expert-level knowledge for Android, iOS, KMP, Flutter, and React Native projects.

## Available Skills

Invoke these skills when the context matches. Use `/skill <name>` or load them automatically when the situation calls for it.

### Core Workflow Skills
| Skill | When to Use |
|-------|-------------|
| `fix-issue` | User asks to fix a bug from an issue tracker (Jira, GitHub Issues, Linear) |
| `reproduce-bug` | User asks to reproduce a bug on a device, emulator, or simulator |
| `analyze-crash` | User shares a crash log, stack trace, or Crashlytics report |
| `write-tests` | User asks to write tests for mobile code |
| `review-code` | User asks for a code review of mobile code changes |
| `create-pr` | User asks to create a pull request for mobile changes |

### Platform Skills
| Skill | When to Use |
|-------|-------------|
| `android-device` | Need to interact with an Android device/emulator (adb, UI automation, screenshots) |
| `android-build` | Need to build an Android project (Gradle, flavors, signing, APK/AAB) |
| `android-testing` | Writing or running Android tests (JUnit, MockK, Espresso, Compose Testing) |
| `android-architecture` | Working with Android architecture patterns (MVVM, MVI, Clean Architecture, Compose) |
| `ios-device` | Need to interact with an iOS Simulator (simctl, screenshots, logs) |
| `ios-build` | Need to build an iOS project (xcodebuild, schemes, CocoaPods, SPM) |
| `ios-testing` | Writing or running iOS tests (XCTest, Quick/Nimble, snapshot tests) |
| `ios-architecture` | Working with iOS architecture patterns (SwiftUI, UIKit, TCA, MVVM+Combine) |

### Cross-Platform Skills
| Skill | When to Use |
|-------|-------------|
| `kmp` | Working with Kotlin Multiplatform projects |
| `flutter` | Working with Flutter/Dart projects |
| `react-native` | Working with React Native projects |

### Meta Skills
| Skill | When to Use |
|-------|-------------|
| `writing-skills` | User wants to create a new MobiAI skill |

## How to Use Skills

1. **Auto-detection**: When you recognize a mobile development task, load the relevant skill(s) before responding.
2. **Platform detection**: Check the project structure to determine the platform:
   - `build.gradle` / `build.gradle.kts` → Android
   - `*.xcodeproj` / `*.xcworkspace` → iOS
   - `shared/` + `composeApp/` / `iosApp/` → KMP
   - `pubspec.yaml` → Flutter
   - `package.json` + `metro.config.js` / `app.json` → React Native
3. **Combine skills**: Complex tasks often need multiple skills. For example, `fix-issue` may invoke `reproduce-bug`, `analyze-crash`, `write-tests`, and `create-pr` as sub-steps.
4. **Project context**: Always read the project's `CLAUDE.md` or similar config files first — they contain project-specific conventions that override general skill guidance.

## Principles

- **Minimal fixes**: Change only what's necessary. Don't refactor, don't add unrelated improvements.
- **Evidence-based**: Always verify your work compiles and tests pass before declaring success.
- **Platform-native**: Follow each platform's conventions and idioms. Don't apply Android patterns to iOS or vice versa.
- **Community-driven**: These skills and agents are maintained by mobile developers worldwide. If you find something wrong or missing, check the `writing-skills` skill to learn how to contribute.
