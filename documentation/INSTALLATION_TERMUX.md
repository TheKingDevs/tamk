# 📱 Instalação do TAMK no Termux (Android)

Guia para usar T.A.M.K (Termux APK Manager Kit) diretamente no Termux sem precisar de compilação.

## 📋 TAMK como Binário Go

O T.A.M.K agora é compilado como um binário Go, suportando arquitetura ARM64. Basta compilar diretamente no Termux ou usar o script de instalação.

## 🚀 Instalação Rápida no Termux

### 1. Instalar Dependências

```bash
# Atualizar Termux
pkg update
pkg upgrade

# Instalar Go e ferramentas essenciais
pkg install golang git openjdk-21 kotlin wget zip apksigner aapt2

# Opcional: para melhor experiência
pkg install nano vim
```

### 2. Clonar ou Descarregar o TAMK

**Opção A: Clonar do GitHub**
```bash
cd ~
git clone https://github.com/TheKingDevs/tamk.git
cd tamk
```

**Opção B: Descarregar do ZIP**
```bash
cd ~
wget https://github.com/TheKingDevs/tamk/archive/refs/heads/main.zip
unzip main.zip
cd tamk-main
```

### 3. Compilar o Binário

```bash
# Compilar o binário Go
cd ~/tamk
go build -o bin/tamk ./cmd/tamk

# Verificar
bin/tamk version
```

### 4. Adicionar ao PATH

```bash
# Adicionar ao PATH da sessão atual
export PATH="$HOME/tamk/bin:$PATH"

# Verificar
tamk version
tamk --help
```

## 🔧 Configurar PATH Permanente

Para não precisar exportar o PATH manualmente sempre:

### 1. Editar ~/.bashrc (ou ~/.profile no Termux)

```bash
# Abrir editor
nano ~/.bashrc
```

### 2. Adicionar no final do arquivo

```bash
# Adicionar TAMK ao PATH
export PATH="$HOME/tamk/bin:$PATH"
```

### 3. Aplicar mudanças

```bash
# Sair do Termux e reabrir, ou:
source ~/.bashrc
```

### 4. Agora use normalmente

```bash
tamk version
tamk --help
tamk create
```

## 📁 Estrutura de Uso

Após instalar:

```
~/tamk/                    # Código fonte do TAMK
  ├── cmd/tamk/main.go    # Entrada principal (Go)
  ├── internal/           # Clean Architecture layers
  ├── pkg/                # Pacotes compartilhados
  ├── templates/          # Templates de projeto
  └── documentation/

~/.tamk/                   # Configuração do usuário
  ├── projects/            # Seus projetos Android
  ├── config.json          # Configurações
  └── keystore/            # Keystores assinadas
```

## 🎯 Primeiros Passos

```bash
# 1. Ver versão
tamk version

# 2. Criar novo projeto
tamk create
# Escolher tipo: UI, Console ou WebApp

# 3. Entrar no diretório do projeto
cd ~/.tamk/projects/meu-projeto

# 4. Compilar
tamk build -p minha-senha

# 5. Instalar no dispositivo
tamk install
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

### Problema: "tamk: command not found"

```bash
# Solução: Verificar PATH
export PATH="$HOME/tamk/bin:$PATH"

# Ou recompilar o binário
cd ~/tamk && go build -o bin/tamk ./cmd/tamk
```

### Problema: "Permission denied" ao executar

```bash
# Solução: Tornar executable
chmod +x ~/tamk/bin/tamk

# Executar diretamente:
~/tamk/bin/tamk version
```

### Problema: "java: command not found"

```bash
# Solução: Instalar OpenJDK no Termux
pkg install openjdk-21

# Verificar
java -version
```

### Problema: "aapt2: command not found"

```bash
# Aapt2 precisa ser baixado ou compilado
# Para Termux, use alternativa:
pkg install android-tools

# Ou configure SDK manualmente:
tamk setup
```

### Problema: Arquivo muito grande ao clonar

```bash
# Use shallow clone
git clone --depth 1 https://github.com/TheKingDevs/tamk.git
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
tamk dev

# Desconectar: Ctrl+A, depois D
# Reconectar: screen -r tamk
```

### 2. Função Shell Personalizada

```bash
# Adicionar à ~/.bashrc
tamk_quick() {
    echo "🚀 TAMK Quick Menu"
    echo "1) tamk version"
    echo "2) tamk create"
    echo "3) tamk build"
    echo "4) tamk dev"
    echo "5) tamk --help"
    read -p "Escolha: " opt
    
    case $opt in
        1) tamk version ;;
        2) tamk create ;;
        3) read -p "Senha: " pass && tamk build -p "$pass" ;;
        4) tamk dev ;;
        5) tamk --help ;;
    esac
}

# Usar: tamk_quick
```

### 3. Executar em Background

```bash
# Rodar em background
nohup tamk dev &

# Ver processos
ps aux | grep tamk

# Matar processo
pkill -f "tamk dev"
```

## 📦 Compilar o Binário Go para ARM

Para compilar o binário Go diretamente no Termux:

```bash
# 1. Instalar Go
pkg install golang

# 2. Compilar o binário
cd ~/tamk
go build -o bin/tamk ./cmd/tamk

# 3. Adicionar ao PATH
export PATH="$HOME/tamk/bin:$PATH"
```

**Nota**: O binário compilado será para ARM64 (arquitetura do Termux).

## 📚 Próximas Ações

1. ✅ Instale Go e dependências
2. ✅ Clone ou baixe TAMK
3. ✅ Compile o binário com `go build`
4. ✅ Execute `tamk create`
5. ✅ Desenvolva seu projeto Android!

## 🔗 Recursos Úteis

- [Termux Wiki](https://wiki.termux.com/)
- [Termux GitHub](https://github.com/termux/termux-app)
- [TAMK Documentation](../documentation/)
- [TAMK Quick Start](../documentation/QUICKSTART.md)

## 📞 Suporte

- Termux issues: https://github.com/termux/termux-app/issues
- TAMK issues: https://github.com/TheKingDevs/tamk/issues

---

<div align="center">
  <sub>TAMK v1.0.0 | Termux Edition</sub>
</div>