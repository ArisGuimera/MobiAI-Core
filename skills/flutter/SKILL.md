---
name: flutter
description: Flutter/Dart development — widgets, state management, platform channels, testing
version: 0.1.0
license: MIT
author: Matias Rosenstein
compatibility: [claude-code, cursor, copilot, codex]
platforms: [flutter]
---

# Flutter / Dart

> **Community contribution welcome!** Help flesh out this skill with Flutter patterns, state management approaches, and real-world workflows.

Guide for working with Flutter projects.

## When to Use

- Working on a Flutter/Dart project
- Debugging widget issues, state management problems, or platform channel errors
- Writing tests for Flutter code

## Detecting a Flutter Project

```bash
# Check for Flutter indicators
ls pubspec.yaml lib/ test/ 2>/dev/null
cat pubspec.yaml | head -20
flutter --version
```

## Project Structure

```
project/
  lib/
    main.dart                     # Entry point
    app.dart                      # App widget
    features/
      home/
        home_screen.dart          # Screen widget
        home_bloc.dart            # State management (BLoC example)
        home_state.dart
        home_event.dart
        widgets/                  # Feature-specific widgets
    core/
      models/                     # Data models
      services/                   # API, database
      utils/                      # Helpers
  test/
    features/
      home/
        home_bloc_test.dart
    widget_test.dart
  pubspec.yaml                    # Dependencies
  analysis_options.yaml           # Lint rules
```

## Build Commands

```bash
# Run in debug mode
flutter run

# Build debug APK
flutter build apk --debug

# Build release APK
flutter build apk --release

# Build iOS
flutter build ios --release

# Run analysis (lint)
flutter analyze

# Format code
dart format lib/
```

## State Management Detection

```bash
# BLoC
grep -r "flutter_bloc\|BlocProvider\|BlocBuilder" lib/ --include="*.dart" | head -5

# Riverpod
grep -r "flutter_riverpod\|ProviderScope\|ConsumerWidget" lib/ --include="*.dart" | head -5

# Provider
grep -r "^import.*provider" lib/ --include="*.dart" | head -5

# GetX
grep -r "get:\|GetMaterialApp\|GetxController" lib/ --include="*.dart" | head -5

# MobX
grep -r "flutter_mobx\|Observable\|@action" lib/ --include="*.dart" | head -5
```

## Testing

### Unit Tests

```dart
// test/features/home/home_bloc_test.dart
import 'package:test/test.dart';

void main() {
  group('HomeBloc', () {
    late HomeBloc bloc;

    setUp(() {
      bloc = HomeBloc(repository: MockRepository());
    });

    test('initial state is HomeInitial', () {
      expect(bloc.state, isA<HomeInitial>());
    });

    test('emits [Loading, Loaded] when data fetch succeeds', () {
      expectLater(
        bloc.stream,
        emitsInOrder([isA<HomeLoading>(), isA<HomeLoaded>()]),
      );
      bloc.add(LoadHomeData());
    });
  });
}
```

### Widget Tests

```dart
testWidgets('displays loading indicator', (tester) async {
  await tester.pumpWidget(
    MaterialApp(home: HomeScreen()),
  );
  expect(find.byType(CircularProgressIndicator), findsOneWidget);
});
```

### Run Tests

```bash
# All tests
flutter test

# Specific file
flutter test test/features/home/home_bloc_test.dart

# With coverage
flutter test --coverage
```

## Common Issues

- **Widget rebuild performance**: Check for unnecessary rebuilds with `const` constructors
- **State management leaks**: Ensure BLoC/controllers are properly disposed
- **Platform channel errors**: Check both Dart and native (Android/iOS) code
- **Hot reload doesn't work**: Try hot restart (`R` in terminal) or full restart

## Tracing Bugs

1. **Check the error in the console** — Flutter errors are usually descriptive
2. **Widget tree issues**: Use Flutter Inspector or `debugDumpApp()`
3. **State management**: Add logging to state transitions
4. **Platform channel**: Check both sides (Dart + native)

---

**Want to improve this skill?** Share your Flutter expertise, architecture patterns, and debugging techniques via a PR.
