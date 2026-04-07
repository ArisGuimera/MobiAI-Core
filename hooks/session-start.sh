#!/usr/bin/env bash
# MobiAI-Core — SessionStart hook
# Loads the bootstrap skill into the agent's context on every session start.

set -euo pipefail

PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
SKILL_FILE="$PLUGIN_ROOT/skills/using-mobiai/SKILL.md"

if [ ! -f "$SKILL_FILE" ]; then
  echo '{"message": "MobiAI-Core: bootstrap skill not found at '"$SKILL_FILE"'"}'
  exit 0
fi

# Read the skill content
SKILL_CONTENT=$(cat "$SKILL_FILE")

# Escape for JSON
SKILL_JSON=$(printf '%s' "$SKILL_CONTENT" | python3 -c 'import sys,json; print(json.dumps(sys.stdin.read()))' 2>/dev/null || printf '%s' "$SKILL_CONTENT" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\t/\\t/g' | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

cat <<EOF
{"message": "<MOBIAI_CONTEXT>\n${SKILL_JSON}\n</MOBIAI_CONTEXT>"}
EOF
