#!/usr/bin/env bash
# setup-emulator.sh — Create and start an Android emulator for testing
#
# Usage: ./setup-emulator.sh [avd-name]
# Default AVD name: mobiai_avd

set -euo pipefail

AVD_NAME="${1:-mobiai_avd}"
SDK_ROOT="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"

if [ -z "$SDK_ROOT" ]; then
  echo "Error: ANDROID_HOME or ANDROID_SDK_ROOT must be set"
  exit 1
fi

EMULATOR="$SDK_ROOT/emulator/emulator"
AVDMANAGER="$SDK_ROOT/cmdline-tools/latest/bin/avdmanager"

# Check if AVD already exists
if "$EMULATOR" -list-avds 2>/dev/null | grep -q "^${AVD_NAME}$"; then
  echo "AVD '$AVD_NAME' already exists"
else
  echo "Creating AVD '$AVD_NAME'..."

  # Find an installed system image (prefer google_apis x86_64)
  SYSTEM_IMAGE=""
  for api_dir in "$SDK_ROOT"/system-images/android-*; do
    [ -d "$api_dir" ] || continue
    for variant in google_apis google_apis_playstore default; do
      if [ -d "$api_dir/$variant/x86_64" ]; then
        api_name=$(basename "$api_dir")
        SYSTEM_IMAGE="system-images;${api_name};${variant};x86_64"
        break 2
      fi
    done
  done

  if [ -z "$SYSTEM_IMAGE" ]; then
    echo "No x86_64 system image found. Install one via Android Studio SDK Manager."
    exit 1
  fi

  echo "Using system image: $SYSTEM_IMAGE"
  "$AVDMANAGER" create avd -n "$AVD_NAME" -k "$SYSTEM_IMAGE" -d "pixel_6" --force
fi

# Start emulator
echo "Starting emulator..."
"$EMULATOR" -avd "$AVD_NAME" \
  -no-window \
  -no-audio \
  -no-boot-anim \
  -gpu swiftshader_indirect \
  -no-snapshot-load &

# Wait for boot
echo "Waiting for boot..."
adb wait-for-device
while [ "$(adb -e shell getprop sys.boot_completed 2>/dev/null)" != "1" ]; do
  sleep 3
  echo "  Still booting..."
done

echo "Emulator '$AVD_NAME' is ready!"
adb -e shell wm size
