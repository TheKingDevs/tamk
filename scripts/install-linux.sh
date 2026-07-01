#!/bin/bash
# =============================================================================
# T.A.M.K - Termux APK Manager Kit • Linux Installer
# Version: 1.0.0
# Supports: Debian, Ubuntu, Arch, Fedora, Alpine, and derivatives
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
# DETECTION
# =============================================================================

detect_distro() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        DISTRO="$ID"
        DISTRO_LIKE="${ID_LIKE:-}"
    else
        err "Cannot detect Linux distribution"
        exit 1
    fi
    
    if [[ "$DISTRO" == "debian" ]] || [[ "$DISTRO" == "ubuntu" ]] || [[ "$DISTRO_LIKE" == *"debian"* ]]; then
        PKG_MGR="apt"
        PKG_INSTALL="sudo apt-get install -y"
    elif [[ "$DISTRO" == "arch" ]] || [[ "$DISTRO" == "manjaro" ]] || [[ "$DISTRO_LIKE" == *"arch"* ]]; then
        PKG_MGR="pacman"
        PKG_INSTALL="sudo pacman -S --noconfirm"
    elif [[ "$DISTRO" == "fedora" ]] || [[ "$DISTRO" == "rhel" ]] || [[ "$DISTRO_LIKE" == *"fedora"* ]]; then
        PKG_MGR="dnf"
        PKG_INSTALL="sudo dnf install -y"
    elif [[ "$DISTRO" == "alpine" ]]; then
        PKG_MGR="apk"
        PKG_INSTALL="sudo apk add"
    else
        warn "Unknown distro ($DISTRO), trying apt..."
        PKG_MGR="apt"
        PKG_INSTALL="sudo apt-get install -y"
    fi
    
    ok "Detected: $DISTRO ($PKG_MGR)"
}

# =============================================================================
# CHECKS
# =============================================================================

check_go() {
    if ! command -v go &>/dev/null; then
        warn "Go not found. Installing..."
        $PKG_INSTALL golang
    fi
    local ver
    ver=$(go version | sed -n 's/.*go\([0-9]*\.[0-9]*\).*/\1/p')
    ok "Go $ver found"
}

# =============================================================================
# INSTALLATION
# =============================================================================

deploy_assets() {
    log "Deploying templates, libs, and configs..."
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local repo_root
    repo_root="$(dirname "$script_dir")"

    mkdir -p "$INSTALL_DIR/templates"
    cp -r "$repo_root/templates/"* "$INSTALL_DIR/templates/"
    ok "Templates deployed"

    mkdir -p "$INSTALL_DIR/libs"
    cp "$repo_root/libs/libraries.json" "$INSTALL_DIR/libs/"
    ok "Libraries deployed"

    mkdir -p "$INSTALL_DIR/configs"
    cp "$repo_root/configs/"* "$INSTALL_DIR/configs/"
    ok "Configs deployed"
}

install_binary() {
    log "Installing T.A.M.K v${VERSION}..."
    
    mkdir -p "$BIN_DIR"
    
    if [[ -f "$BIN_DIR/tamk" ]]; then
        local current_version
        current_version=$("$BIN_DIR/tamk" version 2>/dev/null | sed -n 's/.*v\([0-9.]*\).*/\1/p' || echo "")
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
    if [[ -f "$HOME/.bashrc" ]]; then
        shell_rc="$HOME/.bashrc"
    elif [[ -f "$HOME/.zshrc" ]]; then
        shell_rc="$HOME/.zshrc"
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
    echo -e "${BLUE}║${NC}   Linux (Multi-distro)              ${BLUE}║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
    echo ""
    
    detect_distro
    check_go
    echo ""
    
    install_binary
    deploy_assets
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
