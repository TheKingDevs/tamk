# 📱 Instalação do TAMK no Termux (Android)

Guia para usar T.A.M.K (Termux APK Manager Kit) diretamente no Termux sem precisar de compilação.

## 📋 Por que não usar o executável compilado?

O executável `tamk` foi compilado para **Linux x86_64** em um PC/servidor. O Termux roda em **Android ARM/ARM64**, que é uma arquitetura diferente.

**Solução**: Use Python nativamente no Termux! ✅

## 🚀 Instalação Rápida no Termux

### 1. Instalar Dependências

```bash
# Atualizar Termux
pkg update
pkg upgrade

# Instalar Python 3
pkg install python3

# Instalar pip
pkg install python3-pip

# Instalar dependências do TAMK
pip install watchdog websockets

# Opcional: para melhor experiência
pkg install git nano vim
```

### 2. Clonar ou Descarregar o TAMK

**Opção A: Clonar do GitHub**
```bash
cd ~
git clone https://github.com/Shadw-Developer/tamk.git
cd tamk
```

**Opção B: Descarregar do ZIP**
```bash
cd ~
wget https://github.com/Shadw-Developer/tamk/archive/refs/heads/main.zip
unzip main.zip
cd tamk-main
```

### 3. Usar TAMK

```bash
# Método 1: Executar direto
python3 src/main.py --version

# Método 2: Criar alias para facilitar
alias tamk='python3 ~/tamk/src/main.py'

# Agora usar normalmente
tamk --help
tamk --create
tamk --build -p sua-senha
```

## 🔧 Configurar Alias Permanente

Para não digitar `python3 ~/tamk/src/main.py` sempre:

### 1. Editar ~/.bashrc (ou ~/.profile no Termux)

```bash
# Abrir editor
nano ~/.bashrc
```

### 2. Adicionar no final do arquivo

```bash
# TAMK alias
alias tamk='python3 ~/tamk/src/main.py'
```

### 3. Aplicar mudanças

```bash
# Sair do Termux e reabrir, ou:
source ~/.bashrc
```

### 4. Agora use normalmente

```bash
tamk --version
tamk --help
tamk --create
```

## 📁 Estrutura de Uso

Após instalar:

```
~/tamk/                    # Código fonte do TAMK
  ├── src/main.py         # Entrada principal
  ├── src/controllers/
  ├── src/organization/
  ├── src/utils/
  ├── assets/              # Templates e recursos
  └── documentation/

~/.tamk/                   # Configuração do usuário
  ├── projects/            # Seus projetos Android
  ├── config.json          # Configurações
  └── keystore/            # Keystores assinadas
```

## 🎯 Primeiros Passos

```bash
# 1. Ver versão
tamk --version

# 2. Criar novo projeto
tamk --create
# Escolher tipo: UI, Console ou WebApp

# 3. Entrar no diretório do projeto
cd ~/.tamk/projects/meu-projeto

# 4. Compilar
tamk --build -p minha-senha

# 5. Instalar no dispositivo
tamk --install
```

## 🔐 Permissões no Termux

**Importante**: Termux não requer `sudo` (já roda como usuário sem root necessário em muitos casos).

Se precisar de permissões especiais:

```bash
# Alguns comandos podem pedir:
su

# Ou usar:
sudo -u termux ...
```

## 🆘 Solução de Problemas

### Problema: "ModuleNotFoundError: No module named 'watchdog'"

```bash
# Solução: Instalar novamente
pip install watchdog websockets
```

### Problema: "Permission denied" ao executar

```bash
# Solução: Tornar executable
chmod +x ~/tamk/src/main.py

# Ou adicionar shebang e executar direto:
~/tamk/src/main.py --version
```

### Problema: "java: command not found"

```bash
# Solução: Instalar OpenJDK no Termux
pkg install openjdk-17

# Verificar
java -version
```

### Problema: "aapt2: command not found"

```bash
# Aapt2 precisa ser baixado ou compilado
# Para Termux, use alternativa:
pkg install android-tools

# Ou configure SDK manualmente:
tamk --setup
```

### Problema: Arquivo muito grande ao clonar

```bash
# Use shallow clone
git clone --depth 1 https://github.com/Shadw-Developer/tamk.git
```

## 💡 Dicas para Termux

### 1. Usar com Screen

Para manter sessões ativas:

```bash
# Instalar screen
pkg install screen

# Iniciar sessão nomeada
screen -S tamk

# Rodar TAMK
tamk --dev

# Desconectar: Ctrl+A, depois D
# Reconectar: screen -r tamk
```

### 2. Função Shell Personalizada

```bash
# Adicionar à ~/.bashrc
tamk_quick() {
    echo "🚀 TAMK Quick Menu"
    echo "1) tamk --version"
    echo "2) tamk --create"
    echo "3) tamk --build"
    echo "4) tamk --dev"
    echo "5) tamk --help"
    read -p "Escolha: " opt
    
    case $opt in
        1) tamk --version ;;
        2) tamk --create ;;
        3) read -p "Senha: " pass && tamk --build -p "$pass" ;;
        4) tamk --dev ;;
        5) tamk --help ;;
    esac
}

# Usar: tamk_quick
```

### 3. Executar em Background

```bash
# Rodar em background
nohup tamk --dev &

# Ver processos
ps aux | grep tamk

# Matar processo
pkill -f "tamk --dev"
```

## 📦 Alternativa: Compilar Executável para ARM

Se realmente quiser um executável:

```bash
# 1. Instalar cross-compilation tools
pkg install clang build-essential

# 2. Usar PyInstaller com target ARM
# (Avançado - requer configuração complexa)

# 3. Ou usar Buildozer (recomendado para Android)
pip install buildozer
buildozer android debug
```

**Nota**: Este método é mais complexo, não recomendado para iniciantes.

## 📚 Próximas Ações

1. ✅ Instale Python e dependências
2. ✅ Clone ou baixe TAMK
3. ✅ Configure o alias
4. ✅ Execute `tamk --create`
5. ✅ Desenvolva seu projeto Android!

## 🔗 Recursos Úteis

- [Termux Wiki](https://wiki.termux.com/)
- [Termux GitHub](https://github.com/termux/termux-app)
- [TAMK Documentation](../documentation/)
- [TAMK Quick Start](../documentation/QUICKSTART.md)

## 📞 Suporte

- Termux issues: https://github.com/termux/termux-app/issues
- TAMK issues: https://github.com/Shadw-Developer/tamk/issues

---

<div align="center">
  <sub>TAMK v2026.3.0-HMR | Termux Edition</sub>
</div>