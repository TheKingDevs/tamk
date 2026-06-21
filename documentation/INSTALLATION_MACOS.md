# 🍎 Instalação do TAMK no macOS

Guia completo para instalar o T.A.M.K (Termux APK Manager Kit) no macOS 10.15+.

## 📋 Pré-requisitos

- macOS 10.15 (Catalina) ou posterior
- M1/M2 (Apple Silicon) ou Intel processor
- ~250MB de espaço em disco livre
- Acesso a terminal (application > utilities > terminal)

## 🚀 Instalação Rápida

### 1. Download do Pacote

Baixe o arquivo ZIP mais recente:

```bash
# Usando curl
curl -O https://seu-servidor.com/tamk-proprietary-1.0.0-macos.zip

# Ou usando wget
wget https://seu-servidor.com/tamk-proprietary-1.0.0-macos.zip
```

Ou clique direto do navegador para baixar.

### 2. Extrair o Arquivo

```bash
# Extrair (automático no Finder, ou via terminal)
unzip tamk-proprietary-1.0.0-macos.zip

# Entrar no diretório
cd tamk
```

### 3. Executar Script de Instalação

```bash
# Oferecer permissão de execução
chmod +x install.sh

# Executar instalação
sudo ./install.sh
```

### 4. Verificar Instalação

```bash
# Feche o terminal e abra um novo, depois:
tamk version
tamk --help
```

✅ Se os comandos funcionarem, está pronto!

## 📍 Locais de Instalação

Após a instalação:

- **Binário**: `/opt/tamk/tamk` (diretório completo com dependências)
- **Link**: `/usr/local/bin/tamk` (comando global)
- **Configuração**: `~/.tamk/` (arquivos do usuário)

## ⚠️ Aviso de Segurança macOS

Na primeira execução, o macOS pode mostrar um alerta:

> "Não é possível abrir tamk porque não foi verificado por Apple"

**Solução:**

1. Abra **System Preferences** → **Security & Privacy**
2. Na aba **General**, clique em **Allow Anyway** para "tamk"
3. Ou use o terminal:

```bash
# Remover quarentena (se necessário)
xattr -d com.apple.quarantine /opt/tamk/tamk
```

## 🔧 Uso Global

```bash
# Criar novo projeto
tamk create

# Compilar projeto
tamk build -p sua-senha

# Instalar APK no dispositivo
tamk install

# Modo desenvolvimento
tamk dev

# Ver opções
tamk --help
```

## 🗑️ Desinstalação

```bash
# Remover instalação global
sudo rm -rf /opt/tamk

# Remover link simbólico
sudo rm -f /usr/local/bin/tamk

# Remover configurações (opcional)
rm -rf ~/.tamk/
```

## 🔄 Atualização

```bash
# Verificar atualizações
tamk update

# Ou reinstalar manualmente
sudo rm -rf /opt/tamk /usr/local/bin/tamk
sudo ./install.sh [novo-arquivo.zip]
```

## 🆘 Solução de Problemas

### Problema: "tamk: command not found"

```bash
# Opção 1: Feche e reabra o terminal

# Opção 2: Verifique se o link existe
ls -la /usr/local/bin/tamk

# Opção 3: Reinstale
sudo ./install.sh
```

### Problema: "Operation not permitted" ao desinstalar

```bash
# macOS Monterey+ requer SIP desabilitado para /usr/local
# Tente colocar em /usr/local/bin com caminho diferente

# Verificar permissão
ls -la /usr/local/bin/

# Se precisar de acesso de admin
sudo chmod +x /usr/local/bin/tamk
```

### Problema: Antivírus bloqueia execução

Se estiver usando Malwarebytes, Avast, etc:

1. Adicione `/opt/tamk/` à lista de exclusão
2. Ou desabilite temporariamente durante uso

### Problema: "unzip: command not found"

"Architecture or byte order )", espaço é automaticamente criado:

```bash
# Use o Finder
# Duplo clique no arquivo ZIP
```

### Problema: Shell script não funciona (.zsh vs bash)

```bash
# Se usar zsh (padrão em Big Sur+), adicione ao ~/.zshrc:
export PATH="/usr/local/bin:$PATH"

# Depois reinicie:
source ~/.zshrc
```

## 💡 Dicas para macOS

### Usar com Homebrew (opcional)

Se preferir instalar via Homebrew no futuro:

```bash
# Adicionar tab (quando disponível)
brew tap TheKingDevs/tamk
brew install tamk
```

### M1/M2 Apple Silicon

O executável é universal (suporta Intel e Apple Silicon). Se encontrar problemas:

```bash
# Verifique arquitetura
file /opt/tamk/tamk

# Deve mostrar: "Mach-O 64-bit dynamically linked shared library universal"
```

### Variáveis de Ambiente Personalizadas

Edite `~/.zshrc` ou `~/.bash_profile`:

```bash
# Adicionar ao final do arquivo
export TAMK_HOME="/opt/tamk"
export PATH="$TAMK_HOME:$PATH"

# Aplicar mudanças
source ~/.zshrc  # ou ~/.bash_profile
```

## 📦 Instalação em Múltiplos Macs

Para distribuir para vários Macs:

```bash
# 1. Em um Mac, crie instalação personalizada
./install.sh

# 2. Compacte tudo
cd /opt
tar -czf tamk-backup.tar.gz tamk/

# 3. Distribua tamk-backup.tar.gz via USB ou rede
# 4. Em outro Mac:
tar -xzf tamk-backup.tar.gz -C /opt
sudo chown -R root:wheel /opt/tamk
sudo ln -s /opt/tamk/tamk /usr/local/bin/tamk
```

## 📝 Versões Testadas

- ✅ macOS 10.15 (Catalina)
- ✅ macOS 11 (Big Sur)
- ✅ macOS 12 (Monterey)
- ✅ macOS 13 (Ventura)
- ✅ macOS 14 (Sonoma)
- ✅ Intel e Apple Silicon (M1/M2/M3)

## 📞 Suporte

Se encontrar problemas:

1. Verifique a versão: `tamk version`
2. Confira a arquitetura: `uname -m` (deve ser `arm64` ou `x86_64`)
3. Veja permissions: `ls -la /opt/tamk/`
4. Consulte: `/opt/tamk/README.md`
5. Reporte issues em: https://github.com/TheKingDevs/tamk/issues

---

<div align="center">
  <sub>TAMK v1.0.0 | Instalação macOS</sub>
</div>