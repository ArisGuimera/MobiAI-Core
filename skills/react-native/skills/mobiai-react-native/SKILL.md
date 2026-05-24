---
name: mobiai-react-native
description: Use when working on a React Native project — building, testing, debugging, understanding navigation and state management.
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [react-native]
---

# React Native

> **Community contribution welcome!** Help flesh out this skill with React Native patterns, debugging tips, and real-world workflows.

Guide for working with React Native projects.

## When to Use

- Working on a React Native project
- Debugging JavaScript/TypeScript issues, native module problems, or navigation bugs
- Writing tests for React Native code

## Detecting a React Native Project

```bash
# Check for React Native indicators
ls metro.config.js app.json package.json android/ ios/ 2>/dev/null
grep "react-native" package.json | head -5
cat app.json 2>/dev/null | head -10
```

## Project Structure

```
project/
  src/
    screens/
      HomeScreen.tsx
      DetailScreen.tsx
    components/
      Button.tsx
      Card.tsx
    navigation/
      AppNavigator.tsx
    hooks/
      useItems.ts
    services/
      api.ts
    store/                        # State management
      slices/
        itemSlice.ts
  __tests__/
    HomeScreen.test.tsx
  android/                        # Native Android code
  ios/                            # Native iOS code
  package.json
  metro.config.js
  app.json
  tsconfig.json
```

## Build & Run

```bash
# Start Metro bundler
npx react-native start

# Run on Android
npx react-native run-android

# Run on iOS
npx react-native run-ios

# Run on specific simulator
npx react-native run-ios --simulator="iPhone 15 Pro"

# Clean build
cd android && ./gradlew clean && cd ..
cd ios && xcodebuild clean && cd ..

# Install dependencies
npm install  # or yarn install
cd ios && pod install && cd ..
```

## State Management Detection

```bash
# Redux / Redux Toolkit
grep -r "redux\|@reduxjs/toolkit\|createSlice\|configureStore" src/ --include="*.ts" --include="*.tsx" | head -5

# Zustand
grep -r "zustand\|create(" src/ --include="*.ts" --include="*.tsx" | head -5

# MobX
grep -r "mobx\|observable\|@observer" src/ --include="*.ts" --include="*.tsx" | head -5

# React Query / TanStack Query
grep -r "react-query\|@tanstack/react-query\|useQuery" src/ --include="*.ts" --include="*.tsx" | head -5

# Context API
grep -r "createContext\|useContext" src/ --include="*.ts" --include="*.tsx" | head -5
```

## Testing

### Unit Tests (Jest)

```typescript
// __tests__/useItems.test.ts
import { renderHook, act } from '@testing-library/react-hooks';
import { useItems } from '../src/hooks/useItems';

describe('useItems', () => {
  it('returns empty array initially', () => {
    const { result } = renderHook(() => useItems());
    expect(result.current.items).toEqual([]);
  });

  it('loads items on fetch', async () => {
    const { result, waitForNextUpdate } = renderHook(() => useItems());
    act(() => { result.current.fetchItems(); });
    await waitForNextUpdate();
    expect(result.current.items).toHaveLength(3);
  });
});
```

### Component Tests

```typescript
import { render, fireEvent } from '@testing-library/react-native';
import { HomeScreen } from '../src/screens/HomeScreen';

describe('HomeScreen', () => {
  it('renders title', () => {
    const { getByText } = render(<HomeScreen />);
    expect(getByText('Home')).toBeTruthy();
  });

  it('navigates on button press', () => {
    const navigate = jest.fn();
    const { getByText } = render(
      <HomeScreen navigation={{ navigate }} />
    );
    fireEvent.press(getByText('Go to Details'));
    expect(navigate).toHaveBeenCalledWith('Details');
  });
});
```

### Run Tests

```bash
# All tests
npx jest

# Specific file
npx jest __tests__/HomeScreen.test.tsx

# Watch mode
npx jest --watch

# With coverage
npx jest --coverage
```

## Common Issues

- **Metro bundler cache**: `npx react-native start --reset-cache`
- **Native module linking**: `cd ios && pod install && cd ..`
- **Android build failure**: `cd android && ./gradlew clean`
- **TypeScript errors**: `npx tsc --noEmit`
- **Flipper issues**: Check if Flipper is properly configured in native code

## Debugging

```bash
# Check for JavaScript errors in Android logcat
adb logcat | grep ReactNativeJS

# Check for native errors
adb logcat | grep -E "FATAL|crash|Exception"

# iOS logs
xcrun simctl spawn booted log stream --predicate 'process == "MyApp"'
```

## Navigation

```bash
# Detect navigation library
grep -r "react-navigation\|@react-navigation" package.json
grep -r "react-native-navigation" package.json  # Wix navigation
```

---

**Want to improve this skill?** Share your React Native expertise, architecture patterns, and native module tips via a PR.
