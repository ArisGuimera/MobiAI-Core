#!/usr/bin/env bash
# ui-dump.sh — Dump the current UI hierarchy from an Android device
#
# Usage: ./ui-dump.sh [output-dir]
# Default output: ./ui-dump.xml

set -euo pipefail

OUTPUT_DIR="${1:-.}"
OUTPUT_FILE="$OUTPUT_DIR/ui-dump.xml"

adb shell uiautomator dump /sdcard/ui.xml
adb pull /sdcard/ui.xml "$OUTPUT_FILE"
adb shell rm -f /sdcard/ui.xml

echo "UI dump saved to: $OUTPUT_FILE"
echo ""
echo "Elements found:"
grep -oP 'text="[^"]*"' "$OUTPUT_FILE" | sort -u | head -30
