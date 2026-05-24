---
name: mobiai-mobile-brainstorming
description: "You MUST use this before any creative mobile work - creating features, building components, adding functionality, or modifying behavior. Explores user intent, requirements and design before implementation."
license: MIT
compatibility: [claude-code, cursor, copilot, codex, gemini]
platforms: [android, ios, kmp, flutter, react-native]
---

# Brainstorming Mobile Ideas Into Designs

Help turn ideas into fully formed mobile designs and specs through natural collaborative dialogue.

Start by understanding the current project context, then ask questions one at a time to refine the idea. Once you understand what you're building, present the design and get user approval.

<HARD-GATE>
Do NOT invoke any implementation skill, write any code, scaffold any project, or take any implementation action until you have presented a design and the user has approved it. This applies to EVERY project regardless of perceived simplicity.
</HARD-GATE>

## Anti-Pattern: "This Is Too Simple To Need A Design"

Every project goes through this process. A new screen, a single ViewModel, a config change — all of them. "Simple" projects are where unexamined assumptions cause the most wasted work. The design can be short (a few sentences for truly simple projects), but you MUST present it and get approval.

## Checklist

You MUST create a task for each of these items and complete them in order:

1. **Explore project context** — check files, docs, recent commits, detect platform
2. **Ask clarifying questions** — one at a time, understand purpose/constraints/success criteria
3. **Propose 2-3 approaches** — with trade-offs and your recommendation
4. **Present design** — in sections scaled to their complexity, get user approval after each section
5. **Write design doc** — save to `docs/designs/YYYY-MM-DD-<topic>-design.md` and commit
6. **Spec self-review** — quick inline check for placeholders, contradictions, ambiguity, scope (see below)
7. **User reviews written spec** — ask user to review the spec file before proceeding
8. **Transition to implementation** — invoke `mobiai-mobile-planning` skill to create implementation plan

## Pre-flight: check the Brain for prior context

Before exploring the idea, if `<repo>/.mobiai/brain/config.json` exists, do a quick review pass — you don't want to re-litigate decisions the team already made, or design around a workaround that's already been recognised as temporary debt.

**Preferred — MCP tool** (when `mobiai-brain` MCP server is registered):

Invoke `mobile_review` with default args. From the returned `overdue` list, focus on entries whose `platform` matches the project's platform and whose `area` overlaps with the feature space (DI, navigation, networking, persistence, whichever applies). Run `mobile_scan` first if you don't already know the project's platform.

**Fallback — CLI**:

```bash
mobiai brain review --no-fail
```

If 1+ overdue entries match, **mention them to the user** before The Process:

> "Heads up — there are N overdue workaround(s) in <platform>/<area> that touch this space: <title>. I'll continue with the brainstorm; let me know if you want to factor those in first."

Then proceed to The Process. Skip silently when:

- The brain isn't initialized.
- No overdue entries match the feature's area / platform.

## The Process

**Understanding the idea:**

- Check out the current project state first (files, docs, recent commits)
- Detect platform: `build.gradle` → Android, `*.xcodeproj` → iOS, `shared/` + `composeApp/` → KMP, `pubspec.yaml` → Flutter, `package.json` + `metro.config.js` → React Native
- Before asking detailed questions, assess scope: if the request describes multiple independent features, flag this immediately. Don't spend questions refining details of a project that needs to be decomposed first.
- If the project is too large for a single spec, help the user decompose into sub-projects. Each sub-project gets its own spec → plan → implementation cycle.
- For appropriately-scoped projects, ask questions one at a time to refine the idea
- Prefer multiple choice questions when possible, but open-ended is fine too
- Only one question per message
- Focus on understanding: purpose, constraints, success criteria

**Mobile-specific questions to explore (ask only the relevant ones):**

| Category | Questions |
|----------|-----------|
| **Platforms** | Which platforms? Android only? iOS only? Both? KMP shared code? |
| **UI** | New screen or modification? What's the layout? List, form, detail, dashboard? |
| **Navigation** | How does the user get here? Modal, push, tab, deep link? |
| **Data** | Where does the data come from? API, local DB, cache, real-time? |
| **Offline** | Does it need to work offline? What happens without connectivity? |
| **State** | What state needs to be managed? Local, shared, persistent? |
| **Permissions** | Does this need new permissions? Camera, location, notifications, storage? |
| **Lifecycle** | What happens on background/foreground? Screen rotation? Process death? |
| **Performance** | Large lists? Heavy images? Animation? How much data? |
| **Accessibility** | Screen reader support? Dynamic text sizes? Color contrast? |

**Exploring approaches:**

- Propose 2-3 different approaches with trade-offs
- Present options conversationally with your recommendation and reasoning
- Lead with your recommended option and explain why
- Consider platform-native patterns (MVVM, MVI, TCA, BLoC, etc.)

**Presenting the design:**

- Once you believe you understand what you're building, present the design
- Scale each section to its complexity: a few sentences if straightforward, up to 200-300 words if nuanced
- Ask after each section whether it looks right so far
- Cover: architecture, components, data flow, error handling, testing, platform differences
- Be ready to go back and clarify if something doesn't make sense

**Working in existing codebases:**

- Explore the current structure before proposing changes. Follow existing patterns.
- Where existing code has problems that affect the work, include targeted improvements as part of the design.
- Don't propose unrelated refactoring. Stay focused on what serves the current goal.

## After the Design

**Documentation:**

- Write the validated design (spec) to `docs/designs/YYYY-MM-DD-<topic>-design.md`
  - (User preferences for spec location override this default)
- Commit the design document to git

**Spec Self-Review:**
After writing the spec document, look at it with fresh eyes:

1. **Placeholder scan:** Any "TBD", "TODO", incomplete sections, or vague requirements? Fix them.
2. **Internal consistency:** Do any sections contradict each other? Does the architecture match the feature descriptions?
3. **Scope check:** Is this focused enough for a single implementation plan, or does it need decomposition?
4. **Ambiguity check:** Could any requirement be interpreted two different ways? If so, pick one and make it explicit.

Fix any issues inline. No need to re-review — just fix and move on.

**User Review Gate:**
After the spec review passes, ask the user to review the written spec before proceeding:

> "Spec written and committed to `<path>`. Please review it and let me know if you want to make any changes before we start writing out the implementation plan."

Wait for the user's response. Only proceed once the user approves.

**Optional: capture decisions in MobiAI Brain**

After the user approves the spec, check whether `<repo>/.mobiai/brain/config.json` exists. If it does NOT, skip this step silently.

If it does, scan the spec for **architectural decisions worth remembering**. A decision belongs in the brain when it:

- Picks a specific library, pattern or approach over alternatives (e.g. "Koin over Hilt", "MVI over MVVM", "Ktor over Retrofit").
- Sets a project-wide convention that future code will follow (e.g. "all features expose a `Repository` interface", "errors flow as `Result<T, E>`").
- Resolves a tradeoff that took real thought (e.g. "we accept duplicate state across screens to avoid a global store").

Skip routine implementation details that are obvious from the spec itself ("the button is blue", "the screen has a back arrow"). Those don't need to live in the brain — the spec already captures them.

If you find 1+ decisions worth saving, **propose** each save to the user (one-line confirmation per entry). Never invoke silently. A single brainstorming session may produce several distinct decisions — save them as separate entries so future `brain search` and filtering work cleanly.

**Preferred — MCP tool** (if your client has the `mobiai-brain` MCP server registered, you'll see this tool in your toolbox):

Invoke `mobile_save_decision` with:

- `title`: short decision name, e.g. `"Use Koin for DI"`
- `platform`: `android | ios | shared | kmp | flutter | react-native`
- `area`: free-form, e.g. `"dependency_injection" | "navigation" | "state_management"`
- `status`: `active` (decisions are active unless deprecated later via `mobile_promote`)
- `files`: array of repo-relative paths (typically `[<spec_path>, <other_relevant_file>]`)
- `body`: Markdown with `### Decision` / `### Reason` / `### Alternatives considered`

**Fallback — CLI** (if MCP isn't configured):

```bash
mobiai brain save decision \
  --title "<short decision name, e.g. 'Use Koin for DI'>" \
  --platform <android|ios|shared|kmp|flutter|react-native> \
  --area <free-form, e.g. dependency_injection | navigation | state_management> \
  --status active \
  --files "<spec_path>,<any_other_relevant_file>" \
  --body "### Decision
What we decided, in one sentence.

### Reason
Why this over the alternatives. What constraints made the alternatives worse.

### Alternatives considered
- <option A> — why rejected
- <option B> — why rejected"
```

The body's `### Alternatives considered` block is the high-leverage part — it's why future agents won't re-litigate the same decision. Include it when the decision had real tradeoffs.

**Implementation:**

- Invoke the `mobiai-mobile-planning` skill to create a detailed implementation plan
- Do NOT invoke any other skill. `mobiai-mobile-planning` is the next step.

## Key Principles

- **One question at a time** - Don't overwhelm with multiple questions
- **Multiple choice preferred** - Easier to answer than open-ended when possible
- **YAGNI ruthlessly** - Remove unnecessary features from all designs
- **Explore alternatives** - Always propose 2-3 approaches before settling
- **Incremental validation** - Present design, get approval before moving on
- **Platform-native** - Propose patterns the platform community uses, not cross-platform abstractions forced onto native code
- **Be flexible** - Go back and clarify when something doesn't make sense
