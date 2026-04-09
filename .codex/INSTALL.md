# Installing MobiAI-Core for Codex

Enable MobiAI skills in Codex via native skill discovery. Just clone and symlink.

## Prerequisites

- Git

## Installation

1. **Clone the MobiAI-Core repository:**
   ```bash
   git clone https://github.com/ArisGuimera/MobiAI-Core.git ~/.codex/mobiai-core
   ```

2. **Create the skills symlink:**
   ```bash
   mkdir -p ~/.agents/skills
   ln -s ~/.codex/mobiai-core/skills ~/.agents/skills/mobiai-core
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.agents\skills"
   cmd /c mklink /J "$env:USERPROFILE\.agents\skills\mobiai-core" "$env:USERPROFILE\.codex\mobiai-core\skills"
   ```

3. **Restart Codex** (quit and relaunch the CLI) to discover the skills.

## Verify

```bash
ls -la ~/.agents/skills/mobiai-core
```

You should see a symlink (or junction on Windows) pointing to your MobiAI-Core skills directory.

## Updating

```bash
cd ~/.codex/mobiai-core && git pull
```

Skills update instantly through the symlink.

## Uninstalling

```bash
rm ~/.agents/skills/mobiai-core
```

Optionally delete the clone: `rm -rf ~/.codex/mobiai-core`.
