#!/bin/bash

# =============================================================================
# T.A.M.K - Termux APK Manager Kit • Installer (2026)
# Script de instalação automática
# =============================================================================

# Define a variável PREFIX se ela não estiver configurada
: "${PREFIX:=/data/data/com.termux/files/usr}"

# Cores e Estilos
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'
BOLD='\033[1m'

# =============================================================================
# FUNÇÕES AUXILIARES
# =============================================================================

center_block() {
    local input="$1"
    local cols=$(tput cols 2>/dev/null || echo 80)
    while IFS= read -r line; do
        local clean_line=$(echo -e "$line" | sed 's/\x1b\[[0-9;]*m//g')
        local length=${#clean_line}
        local padding=$(( (cols - length) / 2 ))
        [ $padding -lt 0 ] && padding=0
        printf "%${padding}s" ""
        echo -e "$line"
    done <<< "$input"
}

show_banner() {
    clear
    local art=$(toilet -f standard -F metal "T.A.M.K" 2>/dev/null || echo "T.A.M.K")
    center_block "$art"
    center_block "${BLUE}===========================================${NC}"
    center_block "${BLUE}Termux Apk Manager Kit • Installer (2026)${NC}"
    center_block "${BLUE}===========================================${NC}"
    echo ""
}

loading_animation() {
    local pid=$1
    local spin='⣷⣯⣟⡿⢿⣻⣽⣾'
    while kill -0 "$pid" 2>/dev/null; do
        for i in {0..7}; do
            printf "\r${YELLOW}[${spin:$i:1}]${NC} Processando... "
            sleep 0.1
        done
    done
    wait "$pid"
}

log_info() {
    echo -e "${BLUE}==>${NC} $1"
}

log_success() {
    echo -e "${GREEN} ✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW} ⚠${NC} $1"
}

log_error() {
    echo -e "${RED} ✗${NC} $1"
}

# =============================================================================
# VALIDAÇÕES PRÉVIAS
# =============================================================================

check_environment() {
    log_info "Verificando ambiente..."
    
    # Verifica se está no Termux/SmartIDE
    if [[ "$PREFIX" == *"/org.smartide.code"* ]]; then
        ENV_NAME="SmartIDE"
    elif command -v pkg &> /dev/null; then
        ENV_NAME="Termux"
    else
        log_error "Ambiente não suportado. Este script funciona apenas no Termux ou SmartIDE."
        exit 1
    fi
    
    log_success "Ambiente detectado: ${GREEN}$ENV_NAME${NC}"
}

check_dependencies() {
    log_info "Verificando dependências essenciais..."
    
    local missing=()
    
    # Verifica Python
    if ! command -v python3 &> /dev/null && ! command -v python &> /dev/null; then
        missing+=("python")
    fi
    
    # Verifica Java
    if ! command -v java &> /dev/null; then
        missing+=("openjdk-21")
    fi
    
    # Verifica ferramentas Android
    for cmd in aapt2 apksigner zip; do
        if ! command -v $cmd &> /dev/null; then
            missing+=("$cmd")
        fi
    done
    
    # Verifica toilet (para banners)
    if ! command -v toilet &> /dev/null; then
        missing+=("toilet")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_warning "Dependências faltando: ${missing[*]}"
        return 1
    fi
    
    log_success "Todas as dependências estão instaladas"
    return 0
}

install_dependencies() {
    log_info "Instalando/atualizando dependências..."
    
    if ! command -v pkg &> /dev/null; then
        log_error "pkg não encontrado. Impossível instalar dependências."
        return 1
    fi
    
    # Lista completa de pacotes
    local packages=(
        "python"
        "termux-tools"
        "aapt2"
        "apksigner"
        "openjdk-21"
        "kotlin"
        "zip"
        "unzip"
        "wget"
        "git"
        "ncurses-utils"
        "toilet"
    )
    
    # Instala em background com animação
    (pkg install -y "${packages[@]}" > /dev/null 2>&1) &
    loading_animation $!
    
    # Verifica se instalação foi bem sucedida
    if check_dependencies; then
        log_success "Dependências instaladas com sucesso"
        return 0
    else
        log_error "Falha ao instalar algumas dependências"
        return 1
    fi
}

# =============================================================================
# INSTALAÇÃO
# =============================================================================

setup_directories() {
    local install_path="$1"
    
    log_info "Instalando arquivos em $install_path..."
    
    # Cria diretório de instalação
    mkdir -p "$install_path"
    
    if [ $? -ne 0 ]; then
        log_error "Falha ao criar diretório de instalação"
        return 1
    fi
    
    # Copia todos os arquivos (preservando estrutura)
    cp -rf . "$install_path/"
    
    if [ $? -ne 0 ]; then
        log_error "Falha ao copiar arquivos"
        return 1
    fi
    
    log_success "Arquivos copiados com sucesso"
    return 0
}

create_executable() {
    local install_path="$1"
    
    log_info "Criando executável global 'tamk'..."
    
    # Cria o script wrapper
    cat << EOF > "$PREFIX/bin/tamk"
#!/bin/bash
# T.A.M.K Global Executor (2026)
# Termux APK Manager Kit

export TAMK_HOME="$install_path"
export PATH="\$TAMK_HOME:\$PATH"

# Executa o main.py com todos os argumentos
python3 "\$TAMK_HOME/src/main.py" "\$@"
EOF
    
    chmod +x "$PREFIX/bin/tamk"
    
    if [ $? -ne 0 ]; then
        log_error "Falha ao criar executável"
        return 1
    fi
    
    log_success "Executável criado em $PREFIX/bin/tamk"
    return 0
}

setup_environment_config() {
    local install_path="$1"
    
    log_info "Configurando ambiente..."
    
    # Cria arquivo de configuração local
    local config_file="$install_path/settings.local.json"
    
    if [ ! -f "$config_file" ]; then
        cat << EOF > "$config_file"
{
    "sdk_path": "$install_path/sdk",
    "build_tools_version": "34.0.0",
    "platform_version": "34",
    "auto_update_check": true,
    "verbose_mode": false
}
EOF
        log_success "Configuração criada"
    else
        log_info "Configuração já existe, mantendo configurações atuais"
    fi
    
    return 0
}

# =============================================================================
# PÓS-INSTALAÇÃO
# =============================================================================

post_install_setup() {
    log_info "Executando configuração pós-instalação..."
    
    # Cria diretórios necessários
    local tamk_home="$PREFIX/opt/tamk"
    mkdir -p "$tamk_home/sdk"
    mkdir -p "$tamk_home/.cache"
    mkdir -p "$tamk_home/projects"
    
    log_success "Diretórios criados"
}

show_completion_message() {
    clear
    echo ""
    
    local art_sucesso=$(toilet -f standard -F metal "SUCESSO" 2>/dev/null || echo "SUCESSO")
    center_block "$art_sucesso"
    
    center_block "${GREEN}===========================================${NC}"
    center_block "${BOLD}   INSTALAÇÃO NO $ENV_NAME CONCLUÍDA!     ${NC}"
    center_block "${GREEN}===========================================${NC}"
    
    echo ""
    log_info "O comando ${YELLOW}tamk${NC} agora está disponível globalmente."
    log_info "Local da instalação: ${CYAN}$INSTALL_PATH${NC}"
    echo ""
    
    # Dicas de uso
    echo -e "${BLUE}───────────────────────────────────────────${NC}"
    echo -e "${BOLD}📚 PRÓXIMOS PASSOS:${NC}"
    echo -e "${BLUE}───────────────────────────────────────────${NC}"
    echo ""
    echo -e "  1. Teste a instalação:"
    echo -e "     ${GREEN}tamk --version${NC}"
    echo ""
    echo -e "  2. Configure o ambiente (primeiro uso):"
    echo -e "     ${GREEN}tamk --setup${NC}"
    echo ""
    echo -e "  3. Crie seu primeiro projeto:"
    echo -e "     ${GREEN}tamk --create${NC}"
    echo ""
    echo -e "  4. Consulte a documentação:"
    echo -e "     ${CYAN}https://github.com/Shadw-Developer/tamk/tree/main/documentation${NC}"
    echo ""
    echo -e "${BLUE}───────────────────────────────────────────${NC}"
    echo -e "${YELLOW}💡 Dica: ${NC}Execute ${GREEN}tamk --help${NC} para ver todos os comandos"
    echo -e "${BLUE}───────────────────────────────────────────${NC}"
    echo ""
}

show_quick_reference() {
    echo -e "${MAGENTA}═══════════════════════════════════════════════${NC}"
    echo -e "${BOLD}📖 COMANDOS PRINCIPAIS:${NC}"
    echo -e "${MAGENTA}═══════════════════════════════════════════════${NC}"
    echo ""
    printf "  %-30s %s\n" "${GREEN}tamk --create${NC}" "Cria novo projeto"
    printf "  %-30s %s\n" "${GREEN}tamk --build -p <senha>${NC}" "Compila APK"
    printf "  %-30s %s\n" "${GREEN}tamk --install${NC}" "Instala APK no dispositivo"
    printf "  %-30s %s\n" "${GREEN}tamk --dev${NC}" "Modo desenvolvimento (live reload)"
    printf "  %-30s %s\n" "${GREEN}tamk --setup${NC}" "Configura SDK e ambiente"
    printf "  %-30s %s\n" "${GREEN}tamk --update${NC}" "Atualiza o T.A.M.K"
    printf "  %-30s %s\n" "${GREEN}tamk --help${NC}" "Mostra ajuda completa"
    echo ""
    echo -e "${MAGENTA}═══════════════════════════════════════════════${NC}"
}

# =============================================================================
# FLUXO PRINCIPAL
# =============================================================================

main() {
    # Mostra banner inicial
    show_banner
    
    # Detecta ambiente
    check_environment
    
    echo ""
    
    # Verifica e instala dependências
    if ! check_dependencies; then
        if ! install_dependencies; then
            log_error "Não foi possível instalar as dependências necessárias."
            echo ""
            echo -e "Instale manualmente: ${GREEN}pkg install python openjdk-21 aapt2 apksigner zip toilet${NC}"
            exit 1
        fi
    fi
    
    echo ""
    
    # Define caminho de instalação
    INSTALL_PATH="$PREFIX/opt/tamk"
    
    # Instala arquivos
    if ! setup_directories "$INSTALL_PATH"; then
        exit 1
    fi
    
    echo ""
    
    # Cria executável
    if ! create_executable "$INSTALL_PATH"; then
        exit 1
    fi
    
    echo ""
    
    # Configura ambiente
    setup_environment_config "$INSTALL_PATH"
    
    echo ""
    
    # Pós-instalação
    post_install_setup
    
    # Aguarda breve para usuário ler
    sleep 2
    
    # Mostra mensagem de conclusão
    show_completion_message
    show_quick_reference
    
    echo ""
    echo -e "${GREEN}Bom desenvolvimento! 🚀${NC}"
    echo ""
}

# Executa instalação
main "$@"
