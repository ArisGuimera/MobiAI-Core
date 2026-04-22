---
name: mobiai-ios-build
description: Use when building an iOS project, configuring schemes, troubleshooting build failures, or managing dependencies with CocoaPods or SPM
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [ios]
---

# iOS Build System

> **Community contribution welcome!** This skill is a skeleton. Help flesh it out with real-world xcodebuild workflows and troubleshooting tips.

Build iOS projects using xcodebuild and manage dependencies with CocoaPods/SPM.

## When to Use

- Building the app for testing or deployment
- Troubleshooting build failures
- Managing dependencies (CocoaPods, SPM)

## Build Commands

### Discover Project Structure

```bash
# List schemes
xcodebuild -list

# List schemes (workspace)
xcodebuild -workspace MyApp.xcworkspace -list

# Check if it uses workspace or project
ls *.xcworkspace *.xcodeproj 2>/dev/null
```

### Build for Simulator

```bash
# With workspace (CocoaPods projects)
xcodebuild build \
  -workspace MyApp.xcworkspace \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 15 Pro' \
  -quiet

# With project (no CocoaPods)
xcodebuild build \
  -project MyApp.xcodeproj \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 15 Pro' \
  -quiet
```

### Build for Device

```bash
xcodebuild build \
  -workspace MyApp.xcworkspace \
  -scheme MyApp \
  -destination 'generic/platform=iOS' \
  CODE_SIGNING_REQUIRED=NO
```

### Run Tests

```bash
xcodebuild test \
  -workspace MyApp.xcworkspace \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 15 Pro' \
  -quiet
```

### Clean Build

```bash
xcodebuild clean build \
  -workspace MyApp.xcworkspace \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 15 Pro'
```

## Dependency Management

### CocoaPods

```bash
# Install dependencies
pod install

# Update a specific pod
pod update <PodName>

# Check for issues
pod outdated
```

**Always open `.xcworkspace`** (not `.xcodeproj`) after running `pod install`.

### Swift Package Manager

SPM dependencies are defined in the Xcode project or `Package.swift`.

```bash
# Resolve packages
xcodebuild -resolvePackageDependencies \
  -workspace MyApp.xcworkspace \
  -scheme MyApp

# Check Package.swift
cat Package.swift
```

## Code Signing

### Development (Automatic)

Xcode handles signing automatically in most cases. For CLI builds:
```bash
xcodebuild build \
  -allowProvisioningUpdates \
  -scheme MyApp \
  -destination 'generic/platform=iOS'
```

### Skip Signing (for CI or testing)

```bash
xcodebuild build \
  CODE_SIGNING_REQUIRED=NO \
  CODE_SIGNING_ALLOWED=NO \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 15 Pro'
```

## Troubleshooting

### DerivedData corruption
```bash
rm -rf ~/Library/Developer/Xcode/DerivedData
```

### Module not found
```bash
# Regenerate module maps
pod deintegrate && pod install
# Or clean SPM cache
rm -rf ~/Library/Caches/org.swift.swiftpm
```

### Swift version mismatch
```bash
# Check Swift version
swift --version
xcrun swift --version

# Check project's Swift version
grep SWIFT_VERSION *.xcodeproj/project.pbxproj
```

### Provisioning profile issues
```bash
# List installed profiles
ls ~/Library/MobileDevice/Provisioning\ Profiles/
```

---

**Want to improve this skill?** Add your xcodebuild expertise, CI/CD patterns, and signing workflows via a PR.
