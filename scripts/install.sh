#!/usr/bin/env sh
# MobiAI CLI installer (Mac/Linux)
# Usage: curl -fsSL https://mobiai.dev/install.sh | sh
# Override: MOBIAI_INSTALL_BASE=<url> MOBIAI_INSTALL_DIR=<path> MOBIAI_VERSION=<version>
#
# set -e: fail on first error.
# set -u: fail on undefined variable — protects against typos like $INSTAL_DIR.
set -eu

INSTALL_BASE="${MOBIAI_INSTALL_BASE:-https://github.com/ArisGuimera/MobiAI-Core/releases}"
INSTALL_DIR="${MOBIAI_INSTALL_DIR:-$HOME/.mobiai/bin}"

# Detect OS
case "$(uname -s)" in
    Linux*)  OS="linux";;
    Darwin*) OS="darwin";;
    *)
        echo "Error: SO no soportado: $(uname -s)" >&2
        echo "Soportados: Linux, Darwin (macOS)" >&2
        exit 1
        ;;
esac

# Detect arch
case "$(uname -m)" in
    x86_64|amd64)  ARCH="amd64";;
    arm64|aarch64) ARCH="arm64";;
    *)
        echo "Error: arquitectura no soportada: $(uname -m)" >&2
        echo "Soportadas: amd64 (x86_64), arm64 (aarch64)" >&2
        exit 1
        ;;
esac

# Resolve version
if [ -z "${MOBIAI_VERSION:-}" ]; then
    API="https://api.github.com/repos/ArisGuimera/MobiAI-Core/releases?per_page=20"
    if command -v jq >/dev/null 2>&1; then
        LATEST_TAG="$(curl -fsSL "${API}" | jq -r '.[] | select(.tag_name | startswith("cli-v")) | .tag_name' | head -n1)"
    else
        LATEST_TAG="$(curl -fsSL "${API}" | grep -oE '"tag_name":[[:space:]]*"cli-v[^"]+"' | head -n1 | sed 's/.*"\(cli-v[^"]*\)".*/\1/')"
    fi
    if [ -z "${LATEST_TAG}" ]; then
        echo "Error: no pude detectar la última versión de MobiAI CLI en GitHub releases." >&2
        echo "Configurá MOBIAI_VERSION manualmente (ej: MOBIAI_VERSION=0.1.0) o revisá el repo." >&2
        exit 1
    fi
    MOBIAI_VERSION="${LATEST_TAG#cli-v}"
fi

TAG="${LATEST_TAG:-cli-v${MOBIAI_VERSION}}"
echo "Versión: ${MOBIAI_VERSION}"

# Resolve URL — note: archive uses "tar.gz" for Mac/Linux
ARCHIVE="mobiai-${MOBIAI_VERSION}-${OS}-${ARCH}.tar.gz"
URL="${INSTALL_BASE}/download/${TAG}/${ARCHIVE}"

# Download to temp dir
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Instalador de MobiAI CLI"
echo "Detectado: ${OS} ${ARCH}"
echo "Descargando desde: ${URL}"

if ! curl -fsSL "${URL}" -o "${TMP}/${ARCHIVE}"; then
    echo "Error: falló la descarga" >&2
    echo "Revisá tu conexión o configurá MOBIAI_INSTALL_BASE apuntando a un mirror." >&2
    exit 1
fi

# Extract
tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"

# Install
mkdir -p "${INSTALL_DIR}"
mv "${TMP}/mobiai" "${INSTALL_DIR}/mobiai"
chmod +x "${INSTALL_DIR}/mobiai"

echo ""
echo "Instalado en: ${INSTALL_DIR}/mobiai"

# PATH hint
case "${SHELL:-}" in
    */zsh)  RC="${HOME}/.zshrc";;
    */bash) RC="${HOME}/.bashrc";;
    */fish) RC="${HOME}/.config/fish/config.fish";;
    *)      RC="";;
esac

if [ -n "${RC}" ] && ! echo "${PATH}" | grep -q "${INSTALL_DIR}"; then
    echo ""
    echo "Agregalo al PATH (ejecutá esto a mano):"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ${RC}"
    echo "  source ${RC}"
fi

echo ""
echo "Próximo paso: mobiai --version"
