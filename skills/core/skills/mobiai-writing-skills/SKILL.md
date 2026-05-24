---
name: mobiai-writing-skills
description: Guide the user through creating a new MobiAI skill — generate SKILL.md with proper structure, frontmatter, and actionable instructions.
license: MIT
compatibility: [claude-code, cursor, copilot, codex]
---

# Writing Skills

You are helping the user create a new skill for the MobiAI ecosystem. Guide them through the process step by step.

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

A SKILL.md has two parts: **frontmatter** (metadata) and **body** (instructions).

### Frontmatter (required)

```yaml
---
name: my-new-skill
description: Use when [trigger condition] — [what the skill helps with]
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]  # Optional
---
```

The `description` field is critical — it tells the AI tool WHEN to activate this skill. Write it as a trigger condition, not a list of keywords. Example: `"Use when creating features, refactoring, or navigating an Android codebase"`.

### Body (flexible)

The markdown body contains the instructions for the agent. There is no rigid required structure — organize it in whatever way makes sense for the skill. You can use:

- Step-by-step workflows
- Reference tables with commands
- Decision trees
- Platform-specific sections
- Any combination of the above

Look at existing skills for examples — each has the structure that best fits what it does.

## Language

All skills MUST be written in **English**. Skills are technical instructions consumed by AI agents, not user-facing documentation. English ensures compatibility across all models and tools.

## How Skills Are Loaded (Important)

Skills are NOT all loaded at once. The loading model is:

1. On session start, only `using-mobiai/SKILL.md` is loaded — a lightweight table with skill names and short descriptions.
2. The agent reads the `description` field to decide if a skill is relevant to the current task.
3. Only if it matches, the agent reads the full SKILL.md.

This means:
- The `description` in frontmatter is the most important field — it's the ONLY thing the agent sees when deciding whether to load the skill. If it's bad, the skill will never activate.
- The body can be as long and detailed as needed — it's only consumed when relevant.
- We can have many skills without token cost concerns.

## Writing Guidelines

1. **Be specific and actionable.** Agents follow instructions literally. "Search for the crash signal in the codebase" is better than "investigate the issue."

2. **Include exact commands.** When the skill involves terminal commands, provide the full command with placeholders:
   ```bash
   adb -s <serial> shell uiautomator dump /sdcard/ui.xml
   ```

3. **Provide decision trees.** Agents need clear rules for branching logic:
   - If X -> do A
   - If Y -> do B
   - If neither -> do C

4. **Keep the main SKILL.md concise.** Move detailed reference material to `references/` files that are loaded on-demand.

5. **Test with real scenarios.** Before submitting, verify the skill works end-to-end on a real project.

6. **Platform-agnostic where possible.** If the skill applies to multiple platforms, include platform-specific sections rather than creating separate skills.

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier (kebab-case) |
| `description` | Yes | **Critical**: one-line trigger condition that tells the AI WHEN to load this skill. Write "Use when..." not a list of keywords. |
| `license` | Yes | License (MIT recommended) |
| `compatibility` | Yes | Which AI tools support this skill |
| `platforms` | No | Which mobile platforms this applies to |

## Workflow

When the user wants to create a skill:

1. **Ask what the skill should do** — understand the use case, platform, and trigger conditions.
2. **Generate the SKILL.md** — create the file with proper frontmatter and structure, following all guidelines above.
3. **Create the directory** — `skills/<skill-name>/SKILL.md`
4. **Add to the catalog** — update the table in `skills/using-mobiai/SKILL.md`
   - Also verify that the skill's category is represented in the summary list in `README.md`
5. **Suggest testing** — remind the user to test the skill on a real project before submitting.
