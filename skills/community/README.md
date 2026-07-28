# Community skills

> Skills contributed by the MobiAI community. This is the **one pack anyone can add to** — the platform packs (`core`, `android`, `ios`, `kmp`, `flutter`, `react-native`) are curated by the maintainers and a CI guard blocks PRs that change them unless they come from a maintainer.

Install it like any other pack:

```bash
mobiai skills add community
```

It pulls in `core` automatically, so a community skill is discoverable as soon as it's installed.

## Catalog

<!-- Add one row per skill you contribute. Keep it sorted by name. -->

| Skill | What it does | Author |
|---|---|---|
| _none yet_ | Be the first — see below. | — |

## Contribute a community skill

1. Fork the repo and create a branch (`feat/community-<your-skill>`).
2. Create `skills/community/skills/<your-skill-name>/SKILL.md`.
   - The name must be **unique across the whole catalog**. Prefix it so it doesn't collide — e.g. `mobiai-community-<topic>`.
   - Follow the [skill authoring guide](../../docs/crear-skills.md) and the quality checklist in [CONTRIBUTING.md](../../CONTRIBUTING.md).
3. **Register it** by adding a row to the *Catalog* table above. Updating this file is what satisfies the `check-skills-docs` CI gate for community PRs — you don't need to touch the central catalog.
4. Test it on a real project.
5. Open a PR. A maintainer reviews and merges.

You can only add or change files **under `skills/community/`**. If your idea belongs in a platform pack (Android, iOS, …) open an issue or PR proposing it — a maintainer will land it.
