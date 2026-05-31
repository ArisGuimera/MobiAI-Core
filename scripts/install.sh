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
        echo "Error: unsupported architecture: $(uname -m)" >&2
        echo "Supported: amd64 (x86_64), arm64 (aarch64)" >&2
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
        echo "Error: could not detect the latest MobiAI CLI version from GitHub releases." >&2
        echo "Set MOBIAI_VERSION manually (e.g. MOBIAI_VERSION=0.1.0) or check the repo." >&2
        exit 1
    fi
    MOBIAI_VERSION="${LATEST_TAG#cli-v}"
fi

TAG="${LATEST_TAG:-cli-v${MOBIAI_VERSION}}"
echo "Version: ${MOBIAI_VERSION}"

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
    echo "Check your connection or set MOBIAI_INSTALL_BASE to a mirror." >&2
    exit 1
fi

# Extract
tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"

# Install
mkdir -p "${INSTALL_DIR}"
mv "${TMP}/mobiai" "${INSTALL_DIR}/mobiai"
chmod +x "${INSTALL_DIR}/mobiai"

echo ""
echo "Installed at: ${INSTALL_DIR}/mobiai"

# PATH hint
case "${SHELL:-}" in
    */zsh)  RC="${HOME}/.zshrc";;
    */bash) RC="${HOME}/.bashrc";;
    */fish) RC="${HOME}/.config/fish/config.fish";;
    *)      RC="";;
esac

if [ -n "${RC}" ] && ! echo "${PATH}" | grep -q "${INSTALL_DIR}"; then
    echo ""
    echo "Add it to your PATH (run this manually):"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ${RC}"
    echo "  source ${RC}"
fi

# Welcome banner — printed last so the user's eye lands on the
# wordmark + next step together. Respect NO_COLOR per the spec
# (https://no-color.org/) and skip ANSI when stdout isn't a tty.
echo ""
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
    # POSIX sh no interpreta \033 dentro de comillas dobles; hay que materializar
    # el byte ESC con printf para que las secuencias ANSI viajen como tales y no
    # como cuatro caracteres literales.
    BANNER_COLOR="$(printf '\033[1;36m')"   # bold cyan
    TAGLINE_COLOR="$(printf '\033[2;37m')"  # dim white
    RESET="$(printf '\033[0m')"
else
    BANNER_COLOR=""
    TAGLINE_COLOR=""
    RESET=""
fi

printf '%s' "${BANNER_COLOR}"
cat <<'BANNER'
███╗   ███╗  ██████╗  ██████╗  ██╗       ██████╗  ██╗
████╗ ████║ ██╔═══██╗ ██╔══██╗ ██║      ██╔═══██╗ ██║
██╔████╔██║ ██║   ██║ ██████╔╝ ██║      ███████║ ██║
██║╚██╔╝██║ ██║   ██║ ██╔══██╗ ██║      ██╔══██║ ██║
██║ ╚═╝ ██║ ╚██████╔╝ ██████╔╝ ██║      ██║  ██║ ██║
╚═╝     ╚═╝  ╚═════╝  ╚═════╝  ╚═╝      ╚═╝  ╚═╝ ╚═╝
BANNER
printf '%s\n\n' "${RESET}"
printf '%s            AI FOR MOBILE DEVELOPERS%s\n\n' "${TAGLINE_COLOR}" "${RESET}"

echo "Next step: mobiai --version"
