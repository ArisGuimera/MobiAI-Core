# Workflow Patterns

This reference captures reusable patterns extracted from Android CI workflows and a `gradle-ci.properties` baseline.

## Common patterns

### Concurrency

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

### Permissions for test-report publishing

```yaml
permissions:
  contents: read
  checks: write
  pull-requests: write
```

### Gradle CI setup

```yaml
- name: Configure Gradle for CI
  run: |
    chmod +x ./gradlew
    cp gradle-ci.properties gradle.properties || true
    echo "GRADLE_OPTS=-Dorg.gradle.daemon=false -Dorg.gradle.parallel=true -Dorg.gradle.caching=true" >> $GITHUB_ENV
```

### Report publication

```yaml
- name: Publish test results
  uses: EnricoMi/publish-unit-test-result-action@v2
  if: always()
  with:
    files: |
      <xml-glob>
    check_name: <check-name>
    comment_title: <comment-title>
    comment_mode: always
```

### Upload HTML report on failure

```yaml
- name: Upload HTML report
  uses: actions/upload-artifact@v4
  if: failure()
  with:
    name: <artifact-name>-${{ github.run_number }}
    path: <report-path>
    retention-days: <days>
    compression-level: 6
```

## `unittest` patterns

### Suggested job shape

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 15
```

### Cache pattern

```yaml
- name: Restore Gradle cache
  id: gradle-cache
  uses: actions/cache/restore@v4
  with:
    path: |
      ~/.gradle/caches
      ~/.gradle/wrapper
    key: gradle-${{ runner.os }}-${{ hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') }}
    restore-keys: |
      gradle-${{ runner.os }}-
```

```yaml
- name: Save Gradle cache
  if: steps.gradle-cache.outputs.cache-hit != 'true'
  uses: actions/cache/save@v4
  with:
    path: |
      ~/.gradle/caches
      ~/.gradle/wrapper
    key: gradle-${{ runner.os }}-${{ hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') }}
```

### Unit-test command pattern

```bash
./gradlew <unit-test-task> \
  --parallel \
  --build-cache \
  --configuration-cache \
  --no-daemon \
  --warning-mode none \
  --console=plain \
  --stacktrace
```

### Unit-test report paths

```yaml
files: |
  <module>/build/test-results/**/*.xml
path: <module>/build/reports/tests/<report-dir>/
```

Do not hard-code module names or report directories unless they match the repository being edited.

## `androidtest` patterns

### KVM setup

```yaml
- name: Enable KVM permissions
  run: |
    echo 'KERNEL=="kvm", GROUP="kvm", MODE="0666", OPTIONS+="static_node=kvm"' | sudo tee /etc/udev/rules.d/99-kvm4all.rules
    sudo udevadm control --reload-rules
    sudo udevadm trigger --name-match=kvm
```

### AVD cache pattern

```yaml
- name: AVD cache
  uses: actions/cache@v4
  id: avd-cache
  with:
    path: |
      ~/.android/avd/*
      ~/.android/adb*
    key: avd-<key-suffix>
```

### Snapshot creation

```yaml
- name: Create AVD snapshot
  if: steps.avd-cache.outputs.cache-hit != 'true'
  uses: reactivecircus/android-emulator-runner@v2
  with:
    api-level: <api-level>
    target: <target>
    arch: <arch>
    force-avd-creation: false
    emulator-options: <options>
    disable-animations: false
    emulator-boot-timeout: <timeout-seconds>
    script: echo "AVD snapshot created"
```

### Instrumented test execution

```yaml
- name: Run instrumented tests
  uses: reactivecircus/android-emulator-runner@v2
  with:
    api-level: <api-level>
    target: <target>
    arch: <arch>
    force-avd-creation: false
    emulator-options: -no-snapshot-save <options>
    disable-animations: true
    emulator-boot-timeout: <timeout-seconds>
    script: ./gradlew <android-test-task> --no-daemon --warning-mode none --console=plain --stacktrace
```

### Instrumented-test report paths

```yaml
files: |
  <module>/build/outputs/androidTest-results/**/*.xml
path: <module>/build/reports/androidTests/connected/
```

Do not hard-code module names unless they match the target repository.

## `gradle-ci.properties` baseline

When a repository provides `gradle-ci.properties`, treat it as the source of truth for CI Gradle properties.

Typical contents include:

- larger JVM memory for CI
- configuration cache
- build cache
- parallel workers
- daemon disabled
- Kotlin incremental/caching options
- AndroidX / non-transitive R settings
- reduced warning noise

## Dependency-check reminder

Before finalizing a workflow, check whether these action references are still appropriate:

- `actions/checkout`
- `actions/setup-java`
- `actions/cache`
- `actions/upload-artifact`
- `gradle/actions/setup-gradle`
- `reactivecircus/android-emulator-runner`
- `EnricoMi/publish-unit-test-result-action`

Suggested commands:

```bash
gh api repos/actions/checkout/tags --jq '.[0].name'
gh api repos/actions/setup-java/tags --jq '.[0].name'
gh api repos/actions/cache/tags --jq '.[0].name'
gh api repos/actions/upload-artifact/tags --jq '.[0].name'
gh api repos/gradle/actions/tags --jq '.[0].name'
gh api repos/reactivecircus/android-emulator-runner/tags --jq '.[0].name'
gh api repos/EnricoMi/publish-unit-test-result-action/tags --jq '.[0].name'
```
