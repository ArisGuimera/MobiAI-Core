---
name: mobiai-android-testing
description: Use when writing or running tests in an Android project — unit tests, UI tests, choosing the right framework and patterns. Activate via the `Skill` tool; do not paraphrase this skill's workflow from memory.
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
platforms: [android]
---

# Android Testing

## Activation

You are reading this because `Skill(mobiai-android-testing)` was invoked — correct. Every HARD-GATE, phase checkpoint, and approval gate in this document is binding **because** the `Skill` tool was called. None of them bind if a future step, subagent, or different session reproduces this workflow from memory without another `Skill(mobiai-android-testing)` call.

If you need any of this skill's steps in another context, invoke `Skill(mobiai-android-testing)` again. Paraphrasing from memory is not activation.

Comprehensive guide to writing and running tests in Android projects.

## When to Use

- Writing unit tests for Android/Kotlin code
- Setting up test infrastructure
- Running and debugging test suites
- Choosing the right testing approach

## Test Types

| Type | Location | Framework | Speed | Uses Device |
|------|----------|-----------|-------|-------------|
| Unit | `src/test/` | JUnit + MockK | Fast | No |
| Integration | `src/test/` | JUnit + Robolectric | Medium | No (simulated) |
| UI / Instrumented | `src/androidTest/` | Espresso / Compose Testing | Slow | Yes |

## Unit Testing (JUnit + MockK)

### Setup

```kotlin
// build.gradle.kts
dependencies {
    testImplementation("junit:junit:4.13.2")
    testImplementation("io.mockk:mockk:1.13.12")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")
    testImplementation("app.cash.turbine:turbine:1.2.0") // For Flow testing
}
```

### File Location

Mirror the source path:
- Source: `app/src/main/java/com/example/feature/MyViewModel.kt`
- Test: `app/src/test/java/com/example/feature/MyViewModelTest.kt`

### Test Structure

```kotlin
class MyViewModelTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val repository: MyRepository = mockk()
    private lateinit var viewModel: MyViewModel

    @Before
    fun setUp() {
        viewModel = MyViewModel(repository)
    }

    @Test
    fun `loadData sets state to Success when repository returns data`() = runTest {
        // Given
        val expected = listOf(Item("1", "Test"))
        coEvery { repository.getData() } returns expected

        // When
        viewModel.loadData()

        // Then
        assertEquals(UiState.Success(expected), viewModel.state.value)
        coVerify(exactly = 1) { repository.getData() }
    }

    @Test
    fun `loadData sets state to Error when repository throws`() = runTest {
        // Given
        coEvery { repository.getData() } throws IOException("Network error")

        // When
        viewModel.loadData()

        // Then
        assertTrue(viewModel.state.value is UiState.Error)
    }
}
```

### MainDispatcherRule (for ViewModel tests)

```kotlin
class MainDispatcherRule(
    private val dispatcher: TestDispatcher = UnconfinedTestDispatcher()
) : TestWatcher() {
    override fun starting(description: Description) {
        Dispatchers.setMain(dispatcher)
    }
    override fun finished(description: Description) {
        Dispatchers.resetMain()
    }
}
```

### MockK Cheat Sheet

```kotlin
// Create mock
val repo: Repository = mockk()
val repo: Repository = mockk(relaxed = true) // Returns defaults for unconfigured calls

// Stub behavior
every { repo.getData() } returns listOf(item)
coEvery { repo.fetchData() } returns result       // For suspend functions
every { repo.getData() } throws IOException()
every { repo.save(any()) } just runs               // Void functions

// Verify
verify { repo.getData() }
verify(exactly = 1) { repo.save(any()) }
coVerify { repo.fetchData() }
verify(exactly = 0) { repo.delete(any()) }         // Never called

// Capture arguments
val slot = slot<String>()
every { repo.search(capture(slot)) } returns emptyList()
// After call: slot.captured == "search term"

// Mock objects/companions
mockkObject(MySingleton)
every { MySingleton.instance } returns mockk()
```

### Flow Testing with Turbine

```kotlin
@Test
fun `events flow emits navigation event`() = runTest {
    viewModel.events.test {
        viewModel.onSubmitClicked()
        assertEquals(Event.NavigateToSuccess, awaitItem())
        cancelAndConsumeRemainingEvents()
    }
}
```

## Compose UI Testing

### Setup

```kotlin
dependencies {
    androidTestImplementation("androidx.compose.ui:ui-test-junit4")
    debugImplementation("androidx.compose.ui:ui-test-manifest")
}
```

### Test Structure

```kotlin
class MyScreenTest {
    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun submitButton_isDisabled_whenFieldIsEmpty() {
        composeRule.setContent {
            MyScreen(state = MyState(text = ""))
        }

        composeRule.onNodeWithText("Submit").assertIsNotEnabled()
    }

    @Test
    fun errorMessage_isDisplayed_whenStateIsError() {
        composeRule.setContent {
            MyScreen(state = MyState(error = "Something went wrong"))
        }

        composeRule.onNodeWithText("Something went wrong").assertIsDisplayed()
    }
}
```

## Espresso (XML Views)

```kotlin
@Test
fun clickSubmit_showsConfirmation() {
    onView(withId(R.id.btn_submit)).perform(click())
    onView(withText("Confirmed")).check(matches(isDisplayed()))
}
```

## Running Tests

```bash
# All unit tests
./gradlew testDebugUnitTest

# Specific flavor
./gradlew test<Flavor>DebugUnitTest

# Single test class
./gradlew testDebugUnitTest --tests "com.example.MyViewModelTest"

# Single test method
./gradlew testDebugUnitTest --tests "com.example.MyViewModelTest.loadData sets state to Success*"

# Instrumented tests (requires device/emulator)
./gradlew connectedDebugAndroidTest
```

## Common Patterns

### Testing Use Cases
```kotlin
@Test
fun `invoke returns transformed data`() = runTest {
    coEvery { repository.getItems() } returns rawItems
    val result = useCase()
    assertEquals(expectedTransformed, result)
}
```

### Testing Repository with Room
```kotlin
// Use in-memory database for tests
val db = Room.inMemoryDatabaseBuilder(context, AppDatabase::class.java).build()
val dao = db.myDao()
```

### Testing Mapper/Converter Functions
```kotlin
@Test
fun `toEntity maps all fields correctly`() {
    val dto = MyDto(id = "1", name = "Test", value = 42.0)
    val entity = dto.toEntity()
    assertEquals("1", entity.id)
    assertEquals("Test", entity.name)
    assertEquals(42.0, entity.value, 0.001)
}
```

## Troubleshooting

### Tests pass locally but fail in CI
- Check timezone-dependent tests
- Check tests that depend on system locale
- Check flaky tests with `@FlakyTest` or increase timeouts

### MockK issues with final classes
Add to `src/test/resources/mockk-extensions/io.mockk.ifc`:
```
io.mockk.classmocking.enabled=true
```

### Coroutine test timing issues
Use `advanceUntilIdle()` or `runCurrent()` in `runTest` blocks.
