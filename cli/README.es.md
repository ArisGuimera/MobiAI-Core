# MobiAI CLI

> **Leer en otro idioma:** [English](README.md) · **Español**

CLI standalone para instalar las skills de MobiAI en clientes IA que cumplen el estándar [agentskills.io](https://agentskills.io).

> Arquitectura interna y razonamiento de diseño: ver el design doc interno (gitignored).

## Instalación pública

```bash
# Mac/Linux
curl -fsSL https://mobiai.dev/install.sh | sh

# Windows (cmd)
curl.exe -fsSL https://mobiai.dev/install.cmd -o "%TEMP%\i.cmd" && "%TEMP%\i.cmd"

# Windows (PowerShell)
iwr -useb https://mobiai.dev/install.ps1 | iex
```

## Desarrollo

### Requisitos

- Go 1.22+
- (Para releases) GoReleaser v2

### Build local

```bash
cd cli
go build -o /tmp/mobiai ./cmd/mobiai
/tmp/mobiai --version
```

### Tests

```bash
cd cli
go test ./...
```

### Snapshot release local

```bash
cd cli
goreleaser build --snapshot --clean
ls dist/
```

### Disparar una release pública

Los tags con formato `cli-v*` lanzan el workflow `release-cli`:

```bash
git tag cli-v0.1.1
git push origin cli-v0.1.1
```

Esto compila los binarios para los 6 targets OS+arch y crea una GitHub Release con checksums.

### Override de URLs para testing

Los scripts de instalación respetan variables de entorno:

```bash
MOBIAI_INSTALL_BASE=http://localhost:8080 sh scripts/install.sh
MOBIAI_INSTALL_DIR=/tmp/mobiai-test sh scripts/install.sh
```

## Estructura del proyecto

```
cli/
├── cmd/mobiai/          # entry point
│   ├── main.go
│   └── main_test.go
├── go.mod  go.sum
└── .goreleaser.yml      # config de release
```

## Licencia

MIT — ver el LICENSE de la raíz del repositorio.
