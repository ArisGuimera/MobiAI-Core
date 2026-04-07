---
name: android-architecture
description: Use when creating features, refactoring, or navigating an Android codebase — understand project structure, find where to make changes, follow existing patterns
version: 0.1.0
license: MIT
author: Matias Rosenstein
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android]
---

# Android Architecture

Common architecture patterns and project structures in Android apps.

## When to Use

- Navigating an unfamiliar Android codebase
- Understanding where to make changes for a bug fix
- Writing code that follows the project's existing patterns

## Detecting the Architecture

### Check project structure

```bash
# Find the main source directory
find app/src/main -type d | head -30

# Look for common architecture indicators
grep -r "ViewModel\|UseCase\|Repository\|Interactor\|Presenter" app/src/main --include="*.kt" -l | head -20
```

### Common Patterns

| Pattern | Indicator Files | Layer Structure |
|---------|----------------|-----------------|
| **MVVM + Clean** | `*ViewModel.kt`, `*UseCase.kt`, `*Repository.kt` | UI → ViewModel → UseCase → Repository → DataSource |
| **MVI** | `*Intent.kt`, `*State.kt`, `*Reducer.kt` | UI → ViewModel (reducer) → Repository |
| **MVP** | `*Presenter.kt`, `*View.kt` (interface) | View → Presenter → Model |
| **Simple MVVM** | `*ViewModel.kt`, no UseCases | UI → ViewModel → Repository |

## Clean Architecture (Most Common)

### Layer Structure

```
feature/
  presentation/        # UI Layer
    MyFragment.kt         # Fragment (XML) or Screen (Compose)
    MyViewModel.kt        # ViewModel — UI state + user actions
    MyUiState.kt          # State model for the UI
  domain/              # Business Logic Layer
    usecase/
      GetItemsUseCase.kt  # Single-purpose business operation
    model/
      Item.kt              # Domain model (plain Kotlin)
    repository/
      ItemRepository.kt   # Interface — defined here, implemented in data layer
  data/                # Data Layer
    repository/
      ItemRepositoryImpl.kt  # Implementation of domain interface
    local/
      ItemDao.kt           # Room DAO
      ItemEntity.kt        # Database entity
    remote/
      ItemApi.kt           # Retrofit API interface
      ItemDto.kt           # Network response model
    mapper/
      ItemMapper.kt        # DTO ↔ Entity ↔ Domain model mapping
```

### Tracing a Bug Through Layers

1. **Start from the UI**: Find the Fragment/Activity/Composable for the screen
2. **Follow to ViewModel**: The UI observes ViewModel state. Find what triggers the bug.
3. **Follow to UseCase**: The ViewModel calls UseCases. Which one is involved?
4. **Follow to Repository**: The UseCase calls Repository methods. Where does data come from?
5. **Check Data Sources**: Repository uses DAOs (local) and APIs (remote). Is the data correct?

### Quick Navigation

```bash
# Find ViewModel for a feature
grep -r "class.*ViewModel" app/src/main --include="*.kt" | grep -i "<feature>"

# Find UseCase
grep -r "class.*UseCase\|class.*Interactor" app/src/main --include="*.kt" | grep -i "<feature>"

# Find Repository interface
grep -r "interface.*Repository" app/src/main --include="*.kt" | grep -i "<feature>"

# Find Repository implementation
grep -r "class.*RepositoryImpl\|class.*Repository(" app/src/main --include="*.kt" | grep -i "<feature>"
```

## Jetpack Compose

### Project Structure (Compose)
```
feature/
  presentation/
    MyScreen.kt           # @Composable function
    MyViewModel.kt        # ViewModel with StateFlow
    components/           # Reusable composables for this feature
      ItemCard.kt
      LoadingIndicator.kt
```

### Common Compose Patterns

```kotlin
// ViewModel exposes state
class MyViewModel : ViewModel() {
    private val _state = MutableStateFlow(MyState())
    val state: StateFlow<MyState> = _state.asStateFlow()

    fun onAction(action: MyAction) { ... }
}

// Screen collects state
@Composable
fun MyScreen(viewModel: MyViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    MyScreenContent(state = state, onAction = viewModel::onAction)
}
```

### Finding Compose Screens

```bash
# Find all @Composable screens
grep -r "@Composable" app/src/main --include="*.kt" -l | head -20

# Find navigation graph
grep -r "NavHost\|composable(" app/src/main --include="*.kt" -l
```

## XML Views (Legacy/Hybrid)

### Project Structure (XML)
```
feature/
  presentation/
    MyFragment.kt        # Fragment with ViewBinding
    MyViewModel.kt       # ViewModel
    adapter/             # RecyclerView adapters
      MyAdapter.kt
res/
  layout/
    fragment_my.xml      # Layout file
    item_my.xml          # RecyclerView item layout
```

### Finding XML Layouts

```bash
# Find layout for a fragment
grep -r "R.layout\.\|inflate(" app/src/main --include="*.kt" | grep -i "<feature>"

# Find string resources
grep -r "<feature-keyword>" app/src/main/res/values/strings.xml
```

## Dependency Injection

### Hilt (Most Common)
```bash
# Find modules
grep -r "@Module\|@Provides\|@Binds" app/src/main --include="*.kt" -l

# Find injected classes
grep -r "@Inject\|@HiltViewModel" app/src/main --include="*.kt" | grep -i "<feature>"
```

### Koin
```bash
grep -r "module\s*{" app/src/main --include="*.kt" -l
```

## Build Flavors and Resources

### Country/Environment-specific code
```bash
# Check for flavor-specific source sets
ls app/src/*/java/ 2>/dev/null

# Check for resource overrides
ls app/src/main/res/values*/strings.xml 2>/dev/null
```

### String Resources
```bash
# Find a string by its visible text
grep -r "visible text here" app/src/main/res/values/strings.xml

# Find a string by its ID
grep -r "@string/my_string_id" app/src/main --include="*.xml" --include="*.kt"
```

## Navigation

### Jetpack Navigation
```bash
# Find nav graphs
find app/src/main/res -name "*nav*" -o -name "*navigation*"

# Find navigation actions
grep -r "navigate\|findNavController\|NavController" app/src/main --include="*.kt" | grep -i "<feature>"
```

### Activity-based Navigation
```bash
grep -r "startActivity\|Intent(" app/src/main --include="*.kt" | grep -i "<feature>"
```
