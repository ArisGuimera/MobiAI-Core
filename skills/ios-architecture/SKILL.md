---
name: ios-architecture
description: iOS architecture patterns — SwiftUI, UIKit, TCA, MVVM+Combine, Clean Architecture
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
platforms: [ios]
---

# iOS Architecture

> **Community contribution welcome!** Help improve this skill with your architecture insights, patterns, and real-world examples.

Common architecture patterns in iOS apps.

## When to Use

- Navigating an unfamiliar iOS codebase
- Understanding where to make changes
- Writing code that follows existing patterns

## Detecting the Architecture

```bash
# Check for SwiftUI vs UIKit
grep -r "import SwiftUI" --include="*.swift" -l | wc -l
grep -r "import UIKit" --include="*.swift" -l | wc -l

# Check for architecture indicators
grep -r "ViewModel\|UseCase\|Repository\|Interactor\|Presenter\|Store\|Reducer" --include="*.swift" -l | head -20

# Check for Combine usage
grep -r "import Combine\|@Published\|AnyPublisher" --include="*.swift" -l | wc -l

# Check for TCA (The Composable Architecture)
grep -r "import ComposableArchitecture\|ReducerOf\|Store<" --include="*.swift" -l | head -10
```

## MVVM + SwiftUI (Most Common Modern Pattern)

### Structure
```
Feature/
  Views/
    MyView.swift              # SwiftUI View
    MyDetailView.swift
  ViewModels/
    MyViewModel.swift         # ObservableObject with @Published
  Models/
    MyModel.swift             # Data models
  Services/
    MyService.swift           # API/data access
```

### Patterns

```swift
// ViewModel
@MainActor
class MyViewModel: ObservableObject {
    @Published var items: [Item] = []
    @Published var isLoading = false
    @Published var error: Error?

    private let repository: ItemRepository

    init(repository: ItemRepository) {
        self.repository = repository
    }

    func loadItems() async {
        isLoading = true
        defer { isLoading = false }
        do {
            items = try await repository.fetchItems()
        } catch {
            self.error = error
        }
    }
}

// View
struct MyView: View {
    @StateObject private var viewModel = MyViewModel(repository: .live)

    var body: some View {
        List(viewModel.items) { item in
            Text(item.name)
        }
        .task { await viewModel.loadItems() }
    }
}
```

## MVVM + UIKit + Combine

### Structure
```
Feature/
  MyViewController.swift      # UIViewController
  MyViewModel.swift           # ViewModel with Combine publishers
  MyCoordinator.swift         # Navigation coordinator (optional)
```

### Patterns

```swift
class MyViewModel {
    @Published var items: [Item] = []

    private let repository: ItemRepository
    private var cancellables = Set<AnyCancellable>()

    func loadItems() {
        repository.fetchItems()
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { _ in },
                receiveValue: { [weak self] items in
                    self?.items = items
                }
            )
            .store(in: &cancellables)
    }
}
```

## The Composable Architecture (TCA)

### Structure
```
Feature/
  MyFeature.swift             # Reducer + State + Action
  MyView.swift                # View with Store
```

### Patterns

```swift
@Reducer
struct MyFeature {
    @ObservableState
    struct State: Equatable {
        var items: [Item] = []
        var isLoading = false
    }

    enum Action {
        case loadItems
        case itemsLoaded([Item])
    }

    @Dependency(\.itemClient) var itemClient

    var body: some ReducerOf<Self> {
        Reduce { state, action in
            switch action {
            case .loadItems:
                state.isLoading = true
                return .run { send in
                    let items = try await itemClient.fetch()
                    await send(.itemsLoaded(items))
                }
            case .itemsLoaded(let items):
                state.isLoading = false
                state.items = items
                return .none
            }
        }
    }
}
```

## Clean Architecture (iOS)

### Layer Structure
```
Domain/
  Entities/Item.swift          # Business models
  UseCases/GetItemsUseCase.swift
  Repositories/ItemRepository.swift  # Protocol
Data/
  Repositories/ItemRepositoryImpl.swift
  Network/ItemAPI.swift
  Persistence/ItemStore.swift
Presentation/
  Scenes/ItemList/
    ItemListView.swift
    ItemListViewModel.swift
```

## Navigation Patterns

### NavigationStack (SwiftUI)
```bash
grep -r "NavigationStack\|NavigationLink\|navigationDestination" --include="*.swift" -l | head -10
```

### Coordinator Pattern (UIKit)
```bash
grep -r "Coordinator\|protocol.*Coordinator" --include="*.swift" -l | head -10
```

### Router
```bash
grep -r "Router\|protocol.*Router" --include="*.swift" -l | head -10
```

## Quick Code Navigation

```bash
# Find all ViewModels
grep -r "class.*ViewModel\|struct.*ViewModel" --include="*.swift" -l

# Find all Views/ViewControllers
grep -r "struct.*View.*:.*View\|class.*ViewController" --include="*.swift" -l

# Find dependency injection setup
grep -r "@Environment\|@EnvironmentObject\|container.register" --include="*.swift" -l
```

---

**Want to improve this skill?** Share your architecture patterns, navigation strategies, and real-world iOS expertise via a PR.
