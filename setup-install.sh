#!/bin/bash

# =============================================================================
# T.A.M.K - Termux APK Manager Kit • Installer (2026)
# Script de instalação automática — Go version
# =============================================================================

: "${PREFIX:=/data/data/com.termux/files/usr}"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

log_info()    { echo -e "${BLUE}==>${NC} $1"; }
log_success() { echo -e "${GREEN} ✓${NC} $1"; }
log_warning() { echo -e "${YELLOW} ⚠${NC} $1"; }
log_error()   { echo -e "${RED} ✗${NC} $1"; }

loading_animation() {
	local pid=$1
	local spin='⣷⣯⣟⡿⢿⣻⣽⣾'
	while kill -0 "$pid" 2>/dev/null; do
		for i in {0..7}; do
			printf "\r${YELLOW}[${spin:$i:1}]${NC} Instalando... "
			sleep 0.1
		done
	done
	wait "$pid"
}

show_banner() {
	clear
	echo ""
	echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
	echo -e "${BLUE}║${NC}        ${BOLD}T.A.M.K Installer (Go)${NC}         ${BLUE}║${NC}"
	echo -e "${BLUE}║${NC}   Termux APK Manager Kit v2026.3   ${BLUE}║${NC}"
	echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
	echo ""
}

# =============================================================================
# CHECKS
# =============================================================================

check_environment() {
	log_info "Checking environment..."
	if command -v pkg &>/dev/null; then
		ENV_NAME="Termux"
	elif [[ "$PREFIX" == *"/org.smartide.code"* ]]; then
		ENV_NAME="SmartIDE"
	else
		log_error "Unsupported environment. This script requires Termux."
		exit 1
	fi
	log_success "Environment: ${GREEN}$ENV_NAME${NC}"
}

check_go() {
	if command -v go &>/dev/null; then
		local ver
		ver=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
		log_info "Go ${ver} found"
		if awk "BEGIN {exit !($ver < 1.26)}"; then
			log_warning "Go ${ver} detected, but 1.26+ is recommended"
		fi
		return 0
	fi
	return 1
}

install_go() {
	log_info "Installing Go..."
	(
		pkg install -y golang 2>/dev/null
	) &
	loading_animation $!
	if check_go; then
		log_success "Go installed"
		return 0
	fi
	log_error "Go installation failed"
	return 1
}

check_deps() {
	local missing=()
	for cmd in go java aapt2 apksigner zipalign kotlin; do
		if ! command -v "$cmd" &>/dev/null; then
			missing+=("$cmd")
		fi
	done
	if [ ${#missing[@]} -gt 0 ]; then
		log_warning "Missing: ${missing[*]}"
		return 1
	fi
	log_success "All dependencies installed"
	return 0
}

install_deps() {
	log_info "Installing dependencies..."
	local pkgs=(
		golang openjdk-21 kotlin aapt2 apksigner zipalign
		zip unzip wget git
	)
	(
		pkg install -y "${pkgs[@]}" 2>/dev/null
	) &
	loading_animation $!
	check_deps
}

# =============================================================================
# INSTALLATION
# =============================================================================

install_binary() {
	local install_path="$1"
	log_info "Building T.A.M.K binary..."

	mkdir -p "$install_path"
	cd "$install_path" || return 1

	export GOPATH="$HOME/go"
	export PATH="$PATH:$GOPATH/bin"
	export CGO_ENABLED=0

	go build -o "$install_path/bin/tamk" ./cmd/tamk 2>/tmp/tamk_build.log
	if [ $? -ne 0 ]; then
		log_error "Build failed: $(tail -3 /tmp/tamk_build.log)"
		return 1
	fi

	log_success "Binary built at $install_path/bin/tamk"
	return 0
}

create_wrapper() {
	local install_path="$1"
	log_info "Creating global wrapper..."

	cat << EOF > "$PREFIX/bin/tamk"
#!/bin/bash
export TAMK_HOME="$install_path"
exec "$install_path/bin/tamk" "\$@"
EOF
	chmod +x "$PREFIX/bin/tamk"
	log_success "Wrapper created at $PREFIX/bin/tamk"
}

# =============================================================================
# POST-INSTALL
# =============================================================================

post_install() {
	local install_path="$1"
	mkdir -p "$install_path/development/sdk"
	mkdir -p "$install_path/development/secret"
	mkdir -p "$install_path/assets/templates"
	log_info "Run 'tamk setup' to download SDK and generate keystore"
}

show_done() {
	echo ""
	echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
	echo -e "${GREEN}║${NC}       ${BOLD}INSTALLATION COMPLETE${NC}       ${GREEN}║${NC}"
	echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
	echo ""
	echo -e "  ${BOLD}Usage:${NC}"
	echo -e "    ${GREEN}tamk setup${NC}     — Download SDK + keystore"
	echo -e "    ${GREEN}tamk create${NC}    — Create a project"
	echo -e "    ${GREEN}tamk build${NC}     — Build APK"
	echo -e "    ${GREEN}tamk install${NC}   — Serve APK via QR code"
	echo -e "    ${GREEN}tamk dev${NC}       — HMR dev mode"
	echo -e "    ${GREEN}tamk help${NC}      — All commands"
	echo ""
	echo -e "  ${CYAN}https://github.com/TheKingDevs/tamk${NC}"
	echo ""
}

# =============================================================================
# MAIN
# =============================================================================

main() {
	show_banner
	check_environment
	echo ""

	if ! check_go; then
		install_go
	fi
	echo ""

	if ! check_deps; then
		install_deps
	fi
	echo ""

	INSTALL_PATH="${TAMK_HOME:-$PREFIX/opt/tamk}"

	install_binary "$INSTALL_PATH" || exit 1
	echo ""

	create_wrapper "$INSTALL_PATH"
	echo ""

	post_install "$INSTALL_PATH"
	sleep 2
	show_done
}

main "$@"
