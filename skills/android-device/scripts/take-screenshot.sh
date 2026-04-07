#!/usr/bin/env bash
# take-screenshot.sh — Capture a screenshot from an Android device
#
# Usage: ./take-screenshot.sh [name] [output-dir]
# Default: ./screenshot.png

set -euo pipefail

NAME="${1:-screenshot}"
OUTPUT_DIR="${2:-.}"
OUTPUT_FILE="$OUTPUT_DIR/${NAME}.png"

adb shell screencap -p "/sdcard/${NAME}.png"
adb pull "/sdcard/${NAME}.png" "$OUTPUT_FILE"
adb shell rm -f "/sdcard/${NAME}.png"

echo "Screenshot saved to: $OUTPUT_FILE"
