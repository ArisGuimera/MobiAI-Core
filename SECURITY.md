# Security Policy

## Reporting vulnerabilities

If you discover a security issue, please email <security@mobiai.dev> (or open a private security advisory on GitHub) instead of filing a public issue.

We aim to acknowledge reports within 72 hours.

## For contributors

### Never commit credentials

- Do not push API keys, tokens, GPG keys, or any other credential to this repo. Even if the repo were private, the credential is compromised the moment it lands in git history.
- All CI credentials live in **GitHub Actions repo secrets** (`Settings → Secrets and variables → Actions`).
- Local `.env` files are git-ignored. Each developer maintains their own.

### If you accidentally pushed a credential

1. **Notify the maintainers immediately**.
2. **Rotate the credential.** Assume it is compromised, regardless of how briefly it was visible.
3. Do **not** rely on `git push --force` to "remove" it from history — once pushed, it is public and may already be cached.

### Automated scanning

Every pull request runs:
- **gitleaks** — scans the diff and full history for secret patterns.
- **GitHub native secret scanning** — backstop on the server side.

A red `secret-scan` check blocks merging.

### What is NOT in this repo

- Production credentials of any kind
- Telemetry backend secrets
- Code-signing keys
- Customer data, internal commercial documents

Internal commercial documents (strategy, partnerships, designs) are gitignored under `docs/designs/`, `docs/plans/`, `docs/strategy/`, `docs/roadmap-next-steps.md`. Do not lift the gitignore on those paths without first confirming the contents are safe to publish.
