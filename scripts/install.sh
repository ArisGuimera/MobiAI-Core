#!/usr/bin/env sh
# MobiAI CLI installer (Mac/Linux)
# Usage: curl -fsSL https://mobiai.dev/install.sh | sh
# Override: MOBIAI_INSTALL_BASE=<url> MOBIAI_INSTALL_DIR=<path> MOBIAI_VERSION=<version>
set -e

INSTALL_BASE="${MOBIAI_INSTALL_BASE:-https://github.com/ArisGuimera/MobiAI-Core/releases}"
INSTALL_DIR="${MOBIAI_INSTALL_DIR:-$HOME/.mobiai/bin}"

# Detect OS
case "$(uname -s)" in
    Linux*)  OS="linux";;
    Darwin*) OS="darwin";;
    *)
        echo "Error: unsupported OS: $(uname -s)" >&2
        echo "Supported: Linux, Darwin (macOS)" >&2
        exit 1
        ;;
esac

# Detect arch
case "$(uname -m)" in
    x86_64|amd64)  ARCH="amd64";;
    arm64|aarch64) ARCH="arm64";;
    *)
        echo "Error: unsupported arch: $(uname -m)" >&2
        echo "Supported: amd64 (x86_64), arm64 (aarch64)" >&2
        exit 1
        ;;
esac

# Resolve version
if [ -z "${MOBIAI_VERSION}" ]; then
    API="https://api.github.com/repos/ArisGuimera/MobiAI-Core/releases?per_page=20"
    if command -v jq >/dev/null 2>&1; then
        LATEST_TAG="$(curl -fsSL "${API}" | jq -r '.[] | select(.tag_name | startswith("cli-v")) | .tag_name' | head -n1)"
    else
        LATEST_TAG="$(curl -fsSL "${API}" | grep -oE '"tag_name":[[:space:]]*"cli-v[^"]+"' | head -n1 | sed 's/.*"\(cli-v[^"]*\)".*/\1/')"
    fi
    if [ -z "${LATEST_TAG}" ]; then
        echo "Error: could not detect latest mobiai CLI version from GitHub releases." >&2
        echo "Set MOBIAI_VERSION explicitly (e.g., MOBIAI_VERSION=0.1.0) or check the repo." >&2
        exit 1
    fi
    MOBIAI_VERSION="${LATEST_TAG#cli-v}"
fi

TAG="${LATEST_TAG:-cli-v${MOBIAI_VERSION}}"
echo "Resolved version: ${MOBIAI_VERSION}"

# Resolve URL — note: archive uses "tar.gz" for Mac/Linux
ARCHIVE="mobiai-${MOBIAI_VERSION}-${OS}-${ARCH}.tar.gz"
URL="${INSTALL_BASE}/download/${TAG}/${ARCHIVE}"

# Download to temp dir
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "MobiAI CLI installer"
echo "Detected: ${OS} ${ARCH}"
echo "Downloading from: ${URL}"

if ! curl -fsSL "${URL}" -o "${TMP}/${ARCHIVE}"; then
    echo "Error: download failed" >&2
    echo "Check your network or set MOBIAI_INSTALL_BASE to a working mirror." >&2
    exit 1
fi

# Extract
tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"

# Install
mkdir -p "${INSTALL_DIR}"
mv "${TMP}/mobiai" "${INSTALL_DIR}/mobiai"
chmod +x "${INSTALL_DIR}/mobiai"

echo ""
echo "Installed to: ${INSTALL_DIR}/mobiai"

# PATH hint
case "${SHELL}" in
    */zsh)  RC="${HOME}/.zshrc";;
    */bash) RC="${HOME}/.bashrc";;
    */fish) RC="${HOME}/.config/fish/config.fish";;
    *)      RC="";;
esac

if [ -n "${RC}" ] && ! echo "${PATH}" | grep -q "${INSTALL_DIR}"; then
    echo ""
    echo "Add to PATH (run this manually):"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ${RC}"
    echo "  source ${RC}"
fi

echo ""
echo "Next step: mobiai --version"
