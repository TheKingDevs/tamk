#!/usr/bin/env bash
# T.A.M.K Termux Install Script
# Downloads and installs the latest tamk binary for Android ARM64

set -euo pipefail

REPO="TheKingDevs/tamk"
BINARY="tamk-android-arm64"
INSTALL_DIR="${PREFIX:-/data/data/com.termux/files/usr}/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    warn "This script is designed for Termux. Continuing anyway..."
fi

# Check dependencies
for cmd in wget chmod; do
    if ! command -v "$cmd" &>/dev/null; then
        error "Required command '$cmd' not found"
        error "Run: pkg install wget"
        exit 1
    fi
done

# Get latest release tag
info "Fetching latest release info..."
LATEST=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    error "Failed to fetch latest release"
    exit 1
fi

info "Latest version: ${LATEST}"

# Download binary
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}"
TEMP_FILE=$(mktemp)

info "Downloading ${BINARY}..."
if ! wget -qO "$TEMP_FILE" "$DOWNLOAD_URL"; then
    error "Failed to download binary"
    rm -f "$TEMP_FILE"
    exit 1
fi

# Verify it's a valid binary
if ! file "$TEMP_FILE" | grep -q "ELF"; then
    error "Downloaded file is not a valid ELF binary"
    rm -f "$TEMP_FILE"
    exit 1
fi

# Install
info "Installing to ${INSTALL_DIR}/tamk..."
chmod +x "$TEMP_FILE"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_FILE" "${INSTALL_DIR}/tamk"
else
    warn "Need root access to install to ${INSTALL_DIR}"
    sudo mv "$TEMP_FILE" "${INSTALL_DIR}/tamk"
fi

# Verify installation
if command -v tamk &>/dev/null; then
    info "Installation successful!"
    info "Run 'tamk version' to verify"
    tamk version
else
    warn "tamk installed but not in PATH"
    info "Add ${INSTALL_DIR} to your PATH:"
    info "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
