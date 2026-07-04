#!/bin/bash
# =============================================================================
# T.A.M.K - Update Script
# Version: 1.0.0
# Runs tests, builds binary, and deploys assets to TAMK_HOME (~/.tamk)
# =============================================================================

set -euo pipefail

VERSION="1.0.0"
INSTALL_DIR="${TAMK_HOME:-$HOME/.tamk}"
BIN_DIR="$INSTALL_DIR/bin"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

log()      { echo -e "${BLUE}==>${NC} $1"; }
ok()       { echo -e "${GREEN}  \xE2\x9C\x93${NC} $1"; }
warn()     { echo -e "${YELLOW}  \xE2\x9A\xA0${NC} $1"; }
err()      { echo -e "${RED}  \xE2\x9C\x97${NC} $1"; }
section()  { echo ""; echo -e "${BOLD}${CYAN}$1${NC}"; echo ""; }

# =============================================================================
# VALIDATION
# =============================================================================

validate_go() {
    if ! command -v go &>/dev/null; then
        err "Go is not installed"
        exit 1
    fi
    local ver
    ver=$(go version | sed -n 's/.*go\([0-9]*\.[0-9]*\).*/\1/p')
    ok "Go $ver found"
}

# =============================================================================
# TESTS
# =============================================================================

run_vet() {
    log "Running go vet..."
    if go vet ./... 2>/tmp/tamk_vet.log; then
        ok "go vet passed"
    else
        err "go vet failed:"
        cat /tmp/tamk_vet.log
        exit 1
    fi
}

run_tests() {
    log "Running go tests..."
    if go test ./... -count=1 2>/tmp/tamk_test.log; then
        ok "All tests passed"
    else
        err "Tests failed:"
        cat /tmp/tamk_test.log
        exit 1
    fi
}

# =============================================================================
# BUILD
# =============================================================================

build_binary() {
    log "Building T.A.M.K v${VERSION}..."
    
    export CGO_ENABLED=0
    if ! go build -trimpath \
        -ldflags="-s -w -X 'github.com/TheKingDevs/tamk/internal/config.Version=${VERSION}' -X 'github.com/TheKingDevs/tamk/internal/config.Commit=$(git log --format=%h -1 2>/dev/null || echo unknown)' -X 'github.com/TheKingDevs/tamk/internal/config.Date=$(date +%Y-%m-%d)'" \
        -o "$REPO_ROOT/bin/tamk" \
        ./cmd/tamk 2>/tmp/tamk_build.log; then
        err "Build failed:"
        cat /tmp/tamk_build.log
        exit 1
    fi
    
    chmod +x "$REPO_ROOT/bin/tamk"
    ok "Binary built at $REPO_ROOT/bin/tamk ($(du -h "$REPO_ROOT/bin/tamk" | cut -f1))"
}

# =============================================================================
# DEPLOY
# =============================================================================

deploy_assets() {
    log "Deploying assets to $INSTALL_DIR..."
    
    mkdir -p "$INSTALL_DIR/bin"
    mkdir -p "$INSTALL_DIR/templates"
    mkdir -p "$INSTALL_DIR/libs"
    mkdir -p "$INSTALL_DIR/configs"
    
    # Binary
    cp "$REPO_ROOT/bin/tamk" "$BIN_DIR/tamk"
    chmod +x "$BIN_DIR/tamk"
    ok "Binary deployed to $BIN_DIR/tamk"
    
    # Templates
    rm -rf "$INSTALL_DIR/templates"/*
    cp -r "$REPO_ROOT/templates/"* "$INSTALL_DIR/templates/"
    ok "Templates deployed ($(find "$INSTALL_DIR/templates" -type f | wc -l) files)"
    
    # Libraries
    cp "$REPO_ROOT/libs/libraries.json" "$INSTALL_DIR/libs/"
    ok "Libraries deployed"
    
    # Configs
    rm -rf "$INSTALL_DIR/configs"/*
    cp -r "$REPO_ROOT/configs/"* "$INSTALL_DIR/configs/"
    ok "Configs deployed"
}

# =============================================================================
# VERIFICATION
# =============================================================================

verify_install() {
    log "Verifying installation..."
    
    if [[ ! -f "$BIN_DIR/tamk" ]]; then
        err "Binary not found at $BIN_DIR/tamk"
        exit 1
    fi
    
    local ver
    ver=$("$BIN_DIR/tamk" version 2>/dev/null | head -3)
    ok "Binary version check passed"
    
    if [[ -d "$INSTALL_DIR/templates/ui_apk" ]]; then
        ok "Templates directory OK"
    else
        err "Templates directory missing"
        exit 1
    fi
    
    if [[ -f "$INSTALL_DIR/libs/libraries.json" ]]; then
        ok "Libraries registry OK"
    else
        err "Libraries registry missing"
        exit 1
    fi
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}      ${BOLD}T.A.M.K Update v${VERSION}${NC}         ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   Build \xE2\x86\x92 Test \xE2\x86\x92 Deploy              ${BLUE}║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
    
    section "[1/4] Validating environment"
    validate_go
    
    section "[2/4] Running quality checks"
    run_vet
    run_tests
    
    section "[3/4] Building binary"
    build_binary
    
    section "[4/4] Deploying to TAMK_HOME"
    deploy_assets
    verify_install
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}      ${BOLD}UPDATE COMPLETE${NC}                  ${GREEN}║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BOLD}Summary:${NC}"
    echo -e "    Binary:  ${CYAN}$BIN_DIR/tamk${NC}"
    echo -e "    Version: ${CYAN}$VERSION${NC}"
    echo -e "    TAMK_HOME: ${CYAN}$INSTALL_DIR${NC}"
    echo ""
}

main "$@"
