#!/bin/bash
# =============================================================================
# T.A.M.K - Termux APK Manager Kit • macOS Installer
# Version: 1.0.0
# Supports: macOS (Intel & Apple Silicon)
# =============================================================================

set -euo pipefail

VERSION="1.0.0"
INSTALL_DIR="${TAMK_HOME:-$HOME/.tamk}"
BIN_DIR="$INSTALL_DIR/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

log()  { echo -e "${BLUE}==>${NC} $1"; }
ok()   { echo -e "${GREEN}  ✓${NC} $1"; }
warn() { echo -e "${YELLOW}  ⚠${NC} $1"; }
err()  { echo -e "${RED}  ✗${NC} $1"; }

# =============================================================================
# CHECKS
# =============================================================================

check_macos() {
    if [[ "$(uname)" != "Darwin" ]]; then
        err "This script is for macOS only"
        exit 1
    fi
    ok "macOS detected ($(uname -m))"
}

check_homebrew() {
    if ! command -v brew &>/dev/null; then
        warn "Homebrew not found. Installing..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    ok "Homebrew found"
}

check_go() {
    if ! command -v go &>/dev/null; then
        warn "Go not found. Installing via Homebrew..."
        brew install go
    fi
    local ver
    ver=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
    ok "Go $ver found"
}

check_java() {
    if ! command -v java &>/dev/null; then
        warn "Java not found. Installing OpenJDK 21 via Homebrew..."
        brew install openjdk@21
        # Symlink for macOS
        if [[ -d "/opt/homebrew/opt/openjdk@21" ]]; then
            sudo ln -sfn /opt/homebrew/opt/openjdk@21/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-21.jdk
        fi
    fi
    ok "Java found"
}

# =============================================================================
# INSTALLATION
# =============================================================================

install_binary() {
    log "Installing T.A.M.K v${VERSION}..."
    
    mkdir -p "$BIN_DIR"
    
    if [[ -f "$BIN_DIR/tamk" ]]; then
        local current_version
        current_version=$("$BIN_DIR/tamk" version 2>/dev/null | grep -oP 'v\K[0-9.]+' || echo "")
        if [[ "$current_version" == "$VERSION" ]]; then
            ok "T.A.M.K v${VERSION} already installed"
            return 0
        fi
    fi
    
    local tmp_dir
    tmp_dir=$(mktemp -d)
    
    export CGO_ENABLED=0
    if ! go build -trimpath \
        -ldflags="-s -w -X 'github.com/TheKingDevs/tamk/internal/config.Version=${VERSION}'" \
        -o "$tmp_dir/tamk" \
        ./cmd/tamk 2>/tmp/tamk_build.log; then
        err "Build failed:"
        tail -5 /tmp/tamk_build.log
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    mv "$tmp_dir/tamk" "$BIN_DIR/tamk"
    chmod +x "$BIN_DIR/tamk"
    rm -rf "$tmp_dir"
    
    ok "Binary installed at $BIN_DIR/tamk"
}

setup_path() {
    local shell_rc=""
    if [[ -f "$HOME/.zshrc" ]]; then
        shell_rc="$HOME/.zshrc"
    elif [[ -f "$HOME/.bashrc" ]]; then
        shell_rc="$HOME/.bashrc"
    fi
    
    if [[ -n "$shell_rc" ]]; then
        if ! grep -q "TAMK_HOME" "$shell_rc" 2>/dev/null; then
            echo "" >> "$shell_rc"
            echo "# T.A.M.K" >> "$shell_rc"
            echo "export TAMK_HOME=\"$INSTALL_DIR\"" >> "$shell_rc"
            echo "export PATH=\"\$TAMK_HOME/bin:\$PATH\"" >> "$shell_rc"
            ok "PATH configured in $shell_rc"
        else
            ok "PATH already configured"
        fi
    fi
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}      ${BOLD}T.A.M.K Installer v${VERSION}${NC}       ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   macOS (Intel & Apple Silicon)      ${BLUE}║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
    echo ""
    
    check_macos
    check_homebrew
    check_go
    check_java
    echo ""
    
    install_binary
    setup_path
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}      ${BOLD}INSTALLATION COMPLETE${NC}          ${GREEN}║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BOLD}Next steps:${NC}"
    echo -e "    ${GREEN}tamk setup${NC}     Download SDK + keystore"
    echo -e "    ${GREEN}tamk create${NC}    Create a project"
    echo -e "    ${GREEN}tamk build${NC}     Build APK"
    echo ""
    echo -e "  ${CYAN}https://github.com/TheKingDevs/tamk${NC}"
    echo ""
}

main "$@"
