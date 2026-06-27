# MobiAI CLI

> **Read in another language:** **English** · [Español](README.es.md)

Standalone CLI for installing MobiAI skills across AI clients that conform to the [agentskills.io](https://agentskills.io) standard.

> Internal architecture and design rationale: see internal design doc (gitignored).

## Public install

```bash
# Mac/Linux
curl -fsSL https://mobiai.dev/install.sh | sh

# Windows (cmd)
curl.exe -fsSL https://mobiai.dev/install.cmd -o "%TEMP%\i.cmd" && "%TEMP%\i.cmd"

# Windows (PowerShell)
iwr -useb https://mobiai.dev/install.ps1 | iex
```

## Development

### Requirements

- Go 1.22+
- (For releases) GoReleaser v2

### Build locally

```bash
cd cli
go build -o /tmp/mobiai ./cmd/mobiai
/tmp/mobiai --version
```

### Run tests

```bash
cd cli
go test ./...
```

### Snapshot release locally

```bash
cd cli
goreleaser build --snapshot --clean
ls dist/
```

### Triggering a public release

Tags matching `cli-v*` trigger the `release-cli` workflow:

```bash
git tag cli-v0.1.1
git push origin cli-v0.1.1
```

This builds binaries for all 6 OS+arch targets and creates a GitHub Release with checksums.

### Override URLs for testing

The install scripts honor environment variables:

```bash
MOBIAI_INSTALL_BASE=http://localhost:8080 sh scripts/install.sh
MOBIAI_INSTALL_DIR=/tmp/mobiai-test sh scripts/install.sh
```

## Project structure

```
cli/
├── cmd/mobiai/          # entry point
│   ├── main.go
│   └── main_test.go
├── go.mod  go.sum
└── .goreleaser.yml      # release config
```

## License

MIT — see repository root LICENSE.
