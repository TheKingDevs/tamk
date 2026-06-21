#!/bin/bash
# =============================================================================
# T.A.M.K - Termux APK Manager Kit • Legacy Installer
# This script redirects to the platform-specific installer.
# =============================================================================

VERSION="1.0.0"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

echo ""
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}      ${BOLD}T.A.M.K Installer v${VERSION}${NC}       ${BLUE}║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

# Detect platform
detect_platform() {
    case "$(uname -s)" in
        Linux*)
            if [[ -d "/data/data/com.termux" ]] || [[ "${PREFIX:-}" == *"/com.termux"* ]]; then
                echo "termux"
            elif [[ -f /etc/os-release ]]; then
                . /etc/os-release
                case "$ID" in
                    debian|ubuntu|linuxmint|pop) echo "debian" ;;
                    arch|manjaro|endeavouros) echo "arch" ;;
                    fedora|rhel|centos) echo "fedora" ;;
                    alpine) echo "alpine" ;;
                    *) echo "linux" ;;
                esac
            else
                echo "linux"
            fi
            ;;
        Darwin*)
            echo "macos"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

PLATFORM=$(detect_platform)

case "$PLATFORM" in
    termux)
        echo -e "${GREEN}Detected: Termux (Android)${NC}"
        echo "Running Termux installer..."
        echo ""
        exec bash "$(dirname "$0")/scripts/install-termux.sh" "$@"
        ;;
    debian|arch|fedora|alpine|linux)
        echo -e "${GREEN}Detected: Linux ($PLATFORM)${NC}"
        echo "Running Linux installer..."
        echo ""
        exec bash "$(dirname "$0")/scripts/install-linux.sh" "$@"
        ;;
    macos)
        echo -e "${GREEN}Detected: macOS${NC}"
        echo "Running macOS installer..."
        echo ""
        exec bash "$(dirname "$0")/scripts/install-macos.sh" "$@"
        ;;
    windows)
        echo -e "${GREEN}Detected: Windows${NC}"
        echo "Please run install-windows.bat instead"
        echo ""
        echo -e "  ${BOLD}cmd.exe /c scripts\\install-windows.bat${NC}"
        exit 0
        ;;
    *)
        echo -e "${RED}Unknown platform: $(uname -s)${NC}"
        echo "Please use the appropriate installer from scripts/"
        exit 1
        ;;
esac
