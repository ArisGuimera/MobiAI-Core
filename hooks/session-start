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

# Escape for JSON (cross-platform: try python3, python, node, then sed/awk fallback)
escape_json() {
  printf '%s' "$1" | python3 -c 'import sys,json; print(json.dumps(sys.stdin.read()))' 2>/dev/null && return
  printf '%s' "$1" | python  -c 'import sys,json; print(json.dumps(sys.stdin.read()))' 2>/dev/null && return
  printf '%s' "$1" | node -e "process.stdin.resume();let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>console.log(JSON.stringify(d)))" 2>/dev/null && return
  # Last resort: sed/awk (limited support for special characters)
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\t/\\t/g' | awk '{printf "%s\\n", $0}' | sed 's/\\n$//'
}
SKILL_JSON=$(escape_json "$SKILL_CONTENT")

cat <<EOF
{"message": "<MOBIAI_CONTEXT>\n${SKILL_JSON}\n</MOBIAI_CONTEXT>"}
EOF
