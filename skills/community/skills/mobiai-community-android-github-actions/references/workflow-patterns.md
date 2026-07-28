# Workflow Patterns

This reference captures reusable patterns extracted from Android CI workflows.

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

AVD cache is always tied to the API level (and architecture) — never to the branch — so the same snapshot is safely reused across branches.

```yaml
- name: AVD cache
  uses: actions/cache@v4
  id: avd-cache
  with:
    path: |
      ~/.android/avd/*
      ~/.android/adb*
    key: avd-${{ matrix.api-level }}-${{ matrix.target }}-${{ matrix.arch }}
```

### Gradle cache pattern (per-branch key)

Default behavior of `gradle/actions/setup-gradle` is a shared cache key across branches. If the user wants per-branch reuse (each branch keeps its own warm cache), include `github.ref_name` in the key:

```yaml
- name: 🐘 Gradle cache
  uses: gradle/actions/setup-gradle@v6
  with:
    cache-read-only: ${{ github.ref_name != 'develop' }}
    cache-cleanup: always
    gradle-home-cache-cleanup: true
```

For Java (Maven/Gradle dependency caches), mirror the same branching decision:

```yaml
- name: 🐘 Java cache
  uses: actions/setup-java@v5
  with:
    distribution: temurin
    java-version: <version>
    cache: gradle
    cache-key: ${{ runner.os }}-java-${{ github.ref_name }}-${{ hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') }}
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
gh api repos/actions/checkout/releases/latest --jq '.tag_name'
gh api repos/actions/setup-java/releases/latest --jq '.tag_name'
gh api repos/actions/cache/releases/latest --jq '.tag_name'
gh api repos/actions/upload-artifact/releases/latest --jq '.tag_name'
gh api repos/gradle/actions/releases/latest --jq '.tag_name'
gh api repos/reactivecircus/android-emulator-runner/releases/latest --jq '.tag_name'
gh api repos/EnricoMi/publish-unit-test-result-action/releases/latest --jq '.tag_name'
```
