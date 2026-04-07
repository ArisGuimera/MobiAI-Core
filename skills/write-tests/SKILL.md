---
name: write-tests
description: Write unit and integration tests for mobile code following platform-specific conventions
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
platforms: [android, ios, kmp, flutter, react-native]
---

# Write Tests

Write tests that cover a bug fix or new feature, following the project's existing patterns and platform conventions.

## When to Use

- After applying a fix, to prevent regression
- User asks to add tests for specific code
- As part of the `fix-issue` pipeline

## Workflow

### Step 1: Understand What to Test

1. **Read the changed files** to understand what was fixed or added
2. **Identify the testable unit** — the method, class, or behavior that changed
3. **Define test cases**:
   - The specific behavior that was broken and is now fixed
   - Edge cases related to the change
   - Boundary conditions (null inputs, empty collections, max values)

### Step 2: Find Existing Test Patterns

Before writing tests, find and read existing tests in the project:

```bash
# Find test files near the changed code
find . -path "*/test*" -name "*Test*" | head -20

# Or use Grep to find test files for a specific class
```

Read existing tests to understand:
- Which test framework is used
- Naming conventions
- Mocking approach
- How dependencies are set up
- Where test files are located relative to source files

### Step 3: Write Tests

Follow the project's existing patterns. If no patterns exist, follow platform defaults below.

### Step 4: Run Tests

Run the test suite to verify:
- Your new tests pass
- Existing tests still pass (no regressions from the fix)

If tests fail, fix them. If existing tests fail and your fix caused the regression, fix that too.

## Platform Conventions

### Android (Kotlin)

**Default framework**: JUnit 4 + MockK

**Test location**: Mirror source path under test directory
- Source: `app/src/main/java/com/example/feature/MyClass.kt`
- Test: `app/src/test/java/com/example/feature/MyClassTest.kt`

**Naming**: Backtick-enclosed descriptive names
```kotlin
@Test
fun `calculateTotal returns zero when cart is empty`() { ... }

@Test
fun `formatPrice throws when amount is negative`() { ... }
```

**Mocking**: Use `io.mockk` (`mockk`, `every`, `verify`, `coEvery` for coroutines)

**Coroutines**: Use `kotlinx-coroutines-test` with `runTest`

**Run tests**:
```bash
./gradlew test<Flavor>DebugUnitTest --no-daemon
```

**Rules**:
- No Android framework in unit tests — use Robolectric only if absolutely needed
- Mock external dependencies (repositories, APIs, DAOs)
- Test ViewModels with `MainDispatcherRule` or `Dispatchers.setMain()`

### iOS (Swift)

**Default framework**: XCTest

**Test location**: Inside the test target
- Source: `MyApp/Feature/MyClass.swift`
- Test: `MyAppTests/Feature/MyClassTests.swift`

**Naming**: `test` prefix + descriptive name
```swift
func testCalculateTotal_returnsZero_whenCartIsEmpty() { ... }
func testFormatPrice_throws_whenAmountIsNegative() { ... }
```

**Run tests**:
```bash
xcodebuild test -scheme <scheme> -destination 'platform=iOS Simulator,name=iPhone 15' -quiet
```

### Flutter (Dart)

**Default framework**: `flutter_test`

**Test location**: `test/` directory mirroring `lib/`
- Source: `lib/feature/my_class.dart`
- Test: `test/feature/my_class_test.dart`

**Naming**: `test()` with descriptive strings
```dart
test('calculateTotal returns zero when cart is empty', () { ... });
group('formatPrice', () {
  test('throws when amount is negative', () { ... });
});
```

**Run tests**:
```bash
flutter test
```

### React Native (TypeScript/JavaScript)

**Default framework**: Jest

**Test location**: `__tests__/` directory or co-located `.test.ts` files

**Naming**: `describe` + `it` blocks
```typescript
describe('calculateTotal', () => {
  it('returns zero when cart is empty', () => { ... });
});
```

**Run tests**:
```bash
npx jest
```

### KMP (Kotlin Multiplatform)

**Test location**: `shared/src/commonTest/kotlin/` for shared code

**Run tests**:
```bash
./gradlew :shared:testDebugUnitTest  # Android target
./gradlew :shared:iosSimulatorArm64Test  # iOS target
```

## When Tests Are Not Possible

Set `testsWritten=false` if:
- The fix is in tightly coupled UI code with no way to unit test without massive refactoring
- The fix is a configuration change (manifest, plist, build file)
- The fix is purely a resource change (strings, layouts, assets)

Always explain why tests couldn't be written.
