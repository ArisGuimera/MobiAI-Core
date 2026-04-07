---
name: android-build
description: Gradle build system — flavors, variants, signing, APK/AAB, ProGuard/R8, dependency management
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android]
---

# Android Build System

Expert knowledge for building Android projects with Gradle.

## When to Use

- Building the app for testing or deployment
- Configuring build variants and flavors
- Troubleshooting build failures
- Managing dependencies

## Build Commands

### Assemble (build APK)

```bash
# Debug APK (default flavor)
./gradlew assembleDebug

# Specific flavor
./gradlew assemble<Flavor>Debug
# Example: ./gradlew assembleArgGastronomicDebug

# Release APK
./gradlew assemble<Flavor>Release

# App Bundle (for Play Store)
./gradlew bundle<Flavor>Release
```

### Compile Check (fast, no APK)

Use this to quickly verify code compiles without building a full APK:
```bash
./gradlew compile<Flavor>DebugKotlin --no-daemon
```

### Run Unit Tests

```bash
./gradlew test<Flavor>DebugUnitTest --no-daemon
```

### Run Instrumented Tests

```bash
./gradlew connected<Flavor>DebugAndroidTest
```

### Clean Build

```bash
./gradlew clean assembleDebug
```

## Build Variants

Android builds are a combination of **Build Type** × **Product Flavor**.

### Build Types
- `debug` — debuggable, no minification, debug signing
- `release` — minified (R8/ProGuard), release signing

### Product Flavors

Flavors are defined in `app/build.gradle(.kts)` under `productFlavors`. Common patterns:

```kotlin
// Example: country-based flavors
flavorDimensions += "country"
productFlavors {
    create("arg") { dimension = "country"; applicationIdSuffix = ".arg" }
    create("mx")  { dimension = "country"; applicationIdSuffix = ".mx" }
}
```

To find available flavors:
```bash
grep -r "productFlavors" app/build.gradle* --include="*.gradle*" -A 20
```

Or list all build variants:
```bash
./gradlew tasks --all | grep assemble
```

## APK Installation

```bash
# Single APK
adb install -r app/build/outputs/apk/<flavor>/debug/app-<flavor>-debug.apk

# Split APKs (when using multiple APK output)
adb install-multiple app-base.apk app-config.apk

# Replace existing installation
adb install -r <path-to-apk>
```

## Version & Branch Management

Projects often use branch naming conventions tied to versions:

```bash
# Check available release branches
git branch -r | grep release

# Common patterns:
# release/v3.2.x
# release/VERSION_V2.121.x
# release/3.2.0
```

To build a specific version:
1. Determine the branch name from version info
2. `git checkout <branch>`
3. Build as normal

## Dependency Management

### Check for dependency issues
```bash
./gradlew dependencies --configuration <flavor>DebugRuntimeClasspath
```

### Force a dependency version
```kotlin
configurations.all {
    resolutionStrategy.force("com.example:library:1.2.3")
}
```

## Troubleshooting Common Build Failures

### Out of memory
```bash
# In gradle.properties
org.gradle.jvmargs=-Xmx4g -XX:MaxMetaspaceSize=1g
```

### Cache corruption
```bash
./gradlew clean
rm -rf .gradle/
rm -rf ~/.gradle/caches/
```

### Kotlin version mismatch
```bash
# Check Kotlin version
grep -r "kotlin" gradle/libs.versions.toml build.gradle* --include="*.toml" --include="*.gradle*" | grep version
```

### JDK version issues
```bash
# Check which JDK Gradle uses
./gradlew --version

# Set JDK explicitly
export JAVA_HOME=/path/to/jdk
```

## Signing

### Debug signing
Automatic — uses `~/.android/debug.keystore`

### Release signing
Usually configured in `app/build.gradle(.kts)`:
```bash
grep -r "signingConfigs" app/build.gradle* -A 10
```

**Never commit signing credentials to git.**

## ProGuard / R8

### Check if minification is enabled
```bash
grep -r "isMinifyEnabled\|minifyEnabled" app/build.gradle* --include="*.gradle*"
```

### Keep rules
Located in `proguard-rules.pro` or `consumer-rules.pro`. Common issues:
- Serialization classes stripped → add `@Keep` or keep rules
- Reflection targets removed → add keep rules
- Enum values stripped → add keep rules for enum classes
