---
name: ios-testing
description: iOS testing — XCTest, Quick/Nimble, snapshot testing, UI testing with XCUITest
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot, codex]
platforms: [ios]
---

# iOS Testing

> **Community contribution welcome!** Help improve this skill with your XCTest patterns, mocking strategies, and testing tips.

Guide to writing and running tests in iOS projects.

## When to Use

- Writing unit tests for Swift code
- Setting up test infrastructure
- Running and debugging test suites

## Test Types

| Type | Target | Framework | Speed |
|------|--------|-----------|-------|
| Unit | `MyAppTests` | XCTest | Fast |
| UI | `MyAppUITests` | XCUITest | Slow |
| Snapshot | `MyAppTests` | SnapshotTesting (pointfree) | Medium |

## Unit Testing (XCTest)

### File Location

- Source: `MyApp/Feature/MyViewModel.swift`
- Test: `MyAppTests/Feature/MyViewModelTests.swift`

### Test Structure

```swift
import XCTest
@testable import MyApp

final class MyViewModelTests: XCTestCase {

    private var sut: MyViewModel!
    private var mockRepository: MockRepository!

    override func setUp() {
        super.setUp()
        mockRepository = MockRepository()
        sut = MyViewModel(repository: mockRepository)
    }

    override func tearDown() {
        sut = nil
        mockRepository = nil
        super.tearDown()
    }

    func testLoadData_setsStateToSuccess_whenRepositoryReturnsData() async {
        // Given
        mockRepository.stubbedResult = [Item(id: "1", name: "Test")]

        // When
        await sut.loadData()

        // Then
        XCTAssertEqual(sut.state, .success([Item(id: "1", name: "Test")]))
    }

    func testLoadData_setsStateToError_whenRepositoryThrows() async {
        // Given
        mockRepository.stubbedError = NSError(domain: "", code: -1)

        // When
        await sut.loadData()

        // Then
        if case .error = sut.state {
            // Expected
        } else {
            XCTFail("Expected error state")
        }
    }
}
```

### Naming Conventions

```swift
// Pattern: testMethodName_expectedBehavior_whenCondition
func testCalculateTotal_returnsZero_whenCartIsEmpty() { ... }
func testFormatPrice_throws_whenAmountIsNegative() { ... }
```

### Async Testing

```swift
// async/await
func testFetchData_returnsItems() async throws {
    let items = try await sut.fetchData()
    XCTAssertFalse(items.isEmpty)
}

// Combine
func testPublisher_emitsValue() {
    let expectation = expectation(description: "Value emitted")
    var received: String?

    sut.publisher
        .sink { value in
            received = value
            expectation.fulfill()
        }
        .store(in: &cancellables)

    sut.trigger()
    wait(for: [expectation], timeout: 1.0)
    XCTAssertEqual(received, "expected")
}
```

### Manual Mocking (No Framework)

```swift
class MockRepository: RepositoryProtocol {
    var stubbedResult: [Item] = []
    var stubbedError: Error?
    var fetchDataCallCount = 0

    func fetchData() async throws -> [Item] {
        fetchDataCallCount += 1
        if let error = stubbedError { throw error }
        return stubbedResult
    }
}
```

## XCUITest (UI Testing)

```swift
final class MyAppUITests: XCTestCase {
    let app = XCUIApplication()

    override func setUp() {
        super.setUp()
        continueAfterFailure = false
        app.launch()
    }

    func testLogin_navigatesToHome() {
        app.textFields["Email"].tap()
        app.textFields["Email"].typeText("test@test.com")
        app.secureTextFields["Password"].tap()
        app.secureTextFields["Password"].typeText("password")
        app.buttons["Login"].tap()

        XCTAssertTrue(app.staticTexts["Welcome"].waitForExistence(timeout: 5))
    }
}
```

## Running Tests

```bash
# All unit tests
xcodebuild test -scheme MyApp -destination 'platform=iOS Simulator,name=iPhone 15 Pro' -quiet

# Specific test class
xcodebuild test -scheme MyApp -destination '...' -only-testing:MyAppTests/MyViewModelTests

# Specific test method
xcodebuild test -scheme MyApp -destination '...' -only-testing:MyAppTests/MyViewModelTests/testLoadData_setsStateToSuccess

# UI tests
xcodebuild test -scheme MyApp -destination '...' -only-testing:MyAppUITests
```

## Quick/Nimble (BDD Style)

```swift
import Quick
import Nimble
@testable import MyApp

class MyViewModelSpec: QuickSpec {
    override class func spec() {
        describe("MyViewModel") {
            var sut: MyViewModel!

            beforeEach { sut = MyViewModel() }

            context("when data loads successfully") {
                it("sets state to success") {
                    expect(sut.state).toEventually(equal(.success))
                }
            }
        }
    }
}
```

---

**Want to improve this skill?** Add your testing patterns, mocking strategies, and CI tips via a PR.
