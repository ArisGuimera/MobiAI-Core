# Changelog Management

How to update changelog files in different formats.

## Detecting the Format

1. Check if `CHANGELOG.md` exists → likely [Keep a Changelog](https://keepachangelog.com/)
2. Check if `README.md` has a changelog section → read the format from existing entries
3. Check for project-specific formats in contributing guides

## Keep a Changelog Format

```markdown
## [Unreleased]

### Fixed
- **PROJ-123**: Resolve crash when opening checkout with zero amount

### Added
- **PROJ-456**: Add dark mode support to settings screen
```

Categories: Added, Changed, Deprecated, Removed, Fixed, Security

## Custom Formats

Read the existing entries and match:
- Same indentation
- Same bullet style (* vs - vs +)
- Same issue key format (bold, brackets, links)
- Same language (English, Spanish, etc.)
- Same grouping (by version, by epic/category, by type)

## Rules

1. **Never create duplicate entries** — check if the issue key already exists in the file
2. **Add to the correct version section** — usually the first/most recent one
3. **Match existing style exactly** — indentation, formatting, language
4. **Don't touch anything else** in the file
