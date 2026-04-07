---
name: writing-skills
description: How to create new MobiAI skills — the meta skill for community contributions
version: 0.1.0
license: MIT
author: MobiAI Community
compatibility: [claude-code, cursor, copilot]
---

# Writing MobiAI Skills

This skill teaches you how to create new skills for the MobiAI ecosystem.

## SKILL.md Format

Every skill is a directory under `skills/` containing at minimum a `SKILL.md` file:

```
skills/
  my-new-skill/
    SKILL.md              # Required: skill definition
    references/           # Optional: detailed docs loaded on-demand
      deep-dive.md
    scripts/              # Optional: helper scripts the agent can execute
      helper.sh
    assets/               # Optional: templates, examples
      template.kt
```

## SKILL.md Structure

```yaml
---
name: my-new-skill
description: One-line description used for relevance matching
version: 1.0.0
license: MIT
author: Your Name
compatibility: [claude-code, cursor, copilot]
platforms: [android, ios, kmp, flutter, react-native]  # Optional: which platforms this applies to
---

# Skill Title

## When to Use
Describe the trigger conditions — when should this skill be invoked?

## Steps
Numbered instructions for the agent to follow.

## Platform-Specific Sections (if applicable)
### Android
Android-specific instructions.

### iOS
iOS-specific instructions.

## References
Point to files in references/ for deep dives.

## Common Pitfalls
What to watch out for.
```

## Writing Guidelines

1. **Be specific and actionable.** Agents follow instructions literally. "Search for the crash signal in the codebase" is better than "investigate the issue."

2. **Include exact commands.** When the skill involves terminal commands, provide the full command with placeholders:
   ```bash
   adb -s <serial> shell uiautomator dump /sdcard/ui.xml
   ```

3. **Provide decision trees.** Agents need clear rules for branching logic:
   - If X → do A
   - If Y → do B
   - If neither → do C

4. **Keep the main SKILL.md concise.** Move detailed reference material to `references/` files that are loaded on-demand.

5. **Test with real scenarios.** Before submitting, verify the skill works end-to-end on a real project.

6. **Platform-agnostic where possible.** If the skill applies to multiple platforms, include platform-specific sections rather than creating separate skills.

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier (kebab-case) |
| `description` | Yes | One-line description for relevance matching |
| `version` | Yes | Semver version |
| `license` | Yes | License (MIT recommended) |
| `author` | Yes | Author name or "MobiAI Community" |
| `compatibility` | Yes | Which AI tools support this skill |
| `platforms` | No | Which mobile platforms this applies to |

## Submitting a New Skill

1. Fork the [MobiAI-Core repo](https://github.com/ArisGuimera/MobiAI-Core)
2. Create your skill directory under `skills/`
3. Write and test your `SKILL.md`
4. Add your skill to the table in `skills/using-mobiai/SKILL.md`
5. Open a pull request with a description of what your skill does and how you tested it
