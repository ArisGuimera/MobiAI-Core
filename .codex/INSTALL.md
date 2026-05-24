# Installing MobiAI-Core for Codex

Enable MobiAI skills in Codex via native skill discovery. Clone once, create one symlink per plugin.

## Prerequisites

- Git

## Installation

1. **Clone the MobiAI-Core repository:**

   ```bash
   git clone https://github.com/ArisGuimera/MobiAI-Core.git ~/.codex/mobiai-core
   ```

2. **Create symlinks for each plugin's skills:**

   **Linux / macOS:**
   ```bash
   mkdir -p ~/.agents/skills
   for plugin in core android ios kmp flutter react-native; do
     ln -sf ~/.codex/mobiai-core/skills/$plugin/skills ~/.agents/skills/mobiai-$plugin
   done
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.agents\skills"
   foreach ($plugin in @("core", "android", "ios", "kmp", "flutter", "react-native")) {
     cmd /c mklink /J "$env:USERPROFILE\.agents\skills\mobiai-$plugin" "$env:USERPROFILE\.codex\mobiai-core\skills\$plugin\skills"
   }
   ```

3. **Restart Codex** (quit and relaunch the CLI) to discover the skills.

## Verify

```bash
ls -la ~/.agents/skills/ | grep mobiai
```

You should see 6 symlinks (or junctions on Windows): `mobiai-core`, `mobiai-android`, `mobiai-ios`, `mobiai-kmp`, `mobiai-flutter`, `mobiai-react-native`.

## Google Android Skills (opcional)

Para recibir también los skills oficiales de Google (Apache 2.0):

```bash
git clone https://github.com/android/skills.git ~/.codex/android-skills
ln -sf ~/.codex/android-skills ~/.agents/skills/android-official
```

## Updating

```bash
cd ~/.codex/mobiai-core && git pull
# Si instalaste los de Google:
cd ~/.codex/android-skills && git pull
```

## Uninstalling

```bash
rm ~/.agents/skills/mobiai-*
rm ~/.agents/skills/android-official  # si lo instalaste
```

Optionally delete the clones: `rm -rf ~/.codex/mobiai-core ~/.codex/android-skills`.
