---
name: kmp
description: Kotlin Multiplatform — shared code, expect/actual, platform-specific implementations
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot, codex]
platforms: [kmp]
---

# Kotlin Multiplatform (KMP)

> **Community contribution welcome!** Help flesh out this skill with KMP patterns, common pitfalls, and real-world workflows.

Guide for working with Kotlin Multiplatform projects.

## When to Use

- Working on shared Kotlin code that targets multiple platforms
- Debugging platform-specific issues in KMP projects
- Writing tests for shared modules

## Detecting a KMP Project

```bash
# Check for KMP indicators
ls shared/ composeApp/ iosApp/ 2>/dev/null
grep -r "kotlin(\"multiplatform\")\|KotlinMultiplatform" build.gradle* --include="*.gradle*" | head -5
cat shared/build.gradle.kts | head -30
```

## Project Structure

```
project/
  shared/                          # Shared KMP module
    src/
      commonMain/kotlin/           # Shared code (all platforms)
      commonTest/kotlin/           # Shared tests
      androidMain/kotlin/          # Android-specific implementations
      iosMain/kotlin/              # iOS-specific implementations
    build.gradle.kts
  composeApp/                      # Android app (Compose)
    src/main/
  iosApp/                          # iOS app (SwiftUI/UIKit)
    iosApp.xcodeproj
  build.gradle.kts                 # Root build file
```

## Key Concepts

### expect/actual

Shared interface with platform-specific implementations:

```kotlin
// commonMain
expect class PlatformLogger() {
    fun log(message: String)
}

// androidMain
actual class PlatformLogger {
    actual fun log(message: String) = Log.d("KMP", message)
}

// iosMain
actual class PlatformLogger {
    actual fun log(message: String) = NSLog(message)
}
```

### Finding expect/actual declarations

```bash
# Find expect declarations
grep -r "expect " shared/src/commonMain --include="*.kt" -l

# Find actual implementations
grep -r "actual " shared/src/androidMain --include="*.kt" -l
grep -r "actual " shared/src/iosMain --include="*.kt" -l
```

## Build Commands

```bash
# Build shared module
./gradlew :shared:build

# Run common tests
./gradlew :shared:testDebugUnitTest

# Run iOS tests (on macOS)
./gradlew :shared:iosSimulatorArm64Test

# Build Android app
./gradlew :composeApp:assembleDebug

# Build iOS framework
./gradlew :shared:linkDebugFrameworkIosSimulatorArm64
```

## Testing

### Shared Tests (commonTest)
```kotlin
// shared/src/commonTest/kotlin/MyTest.kt
class MyTest {
    @Test
    fun testSharedLogic() {
        val result = calculate(2, 3)
        assertEquals(5, result)
    }
}
```

### Platform-Specific Tests
```kotlin
// shared/src/androidTest/kotlin/AndroidSpecificTest.kt
class AndroidSpecificTest {
    @Test
    fun testAndroidImplementation() { ... }
}
```

## Common Issues

- **iOS build fails**: Check that the shared framework is linked correctly in Xcode
- **expect/actual mismatch**: Ensure all `expect` declarations have `actual` implementations for all targets
- **Dependency conflicts**: Use `api()` vs `implementation()` carefully in shared module
- **Kotlin/Native memory model**: Be aware of the new memory model (default since Kotlin 1.7.20)

## Tracing Bugs

1. **Determine if the bug is in shared or platform code** — check the stack trace
2. **If shared**: fix in `commonMain`, test in `commonTest`
3. **If platform-specific**: fix in `androidMain`/`iosMain`, test in platform-specific test source set
4. **If the bug manifests differently per platform**: check `expect`/`actual` implementations

---

**Want to improve this skill?** Share your KMP expertise, Compose Multiplatform patterns, and platform bridging tips via a PR.
