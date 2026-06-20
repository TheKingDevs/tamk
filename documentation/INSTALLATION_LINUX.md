# 🐧 Instalação do TAMK no Linux

Guia completo para instalar o T.A.M.K (Termux APK Manager Kit) no Linux.

## 📋 Pré-requisitos

- Linux com suporte a bash (Ubuntu, Debian, Fedora, Arch, etc.)
- Permissões de sudo para instalação global
- Espaço em disco: ~250MB

## 🚀 Instalação Rápida

### 1. Download do Pacote

Baixe o arquivo ZIP mais recente:

```bash
# Opção 1: Usando wget
wget https://seu-servidor.com/tamk-proprietary-2026.3.0-HMR-linux.zip

# Opção 2: Usando curl
curl -O https://seu-servidor.com/tamk-proprietary-2026.3.0-HMR-linux.zip
```

### 2. Extrair e Instalar

```bash
# Extrair o arquivo
unzip tamk-proprietary-2026.3.0-HMR-linux.zip

# Oferecer permissão ao script
chmod +x install.sh

# Executar instalação
sudo ./install.sh
```

### 3. Verificar Instalação

```bash
# Testar versão
tamk version

# Ver ajuda
tamk --help
```

✅ Se os comandos acima funcionarem, a instalação foi bem-sucedida!

## 📍 Locais de Instalação

Após a instalação, os seguintes locais são usados:

- **Binário**: `/opt/tamk/tamk` (diretório completo com dependências)
- **Link**: `/usr/local/bin/tamk` (comando global)
- **Configuração**: `~/.tamk/` (arquivos do usuário)

## 🔧 Uso Global

Após a instalação, você pode usar TAMK em qualquer terminal:

```bash
# Criar novo projeto
tamk create

# Compilar projeto atual
tamk build -p sua-senha

# Instalar APK no dispositivo
tamk install

# Modo desenvolvimento com live reload
tamk dev

# Ver opções
tamk --help
```

## 🗑️ Desinstalação

Para remover completamente o TAMK:

```bash
# Remover instalação global
sudo rm -rf /opt/tamk

# Remover link simbólico
sudo rm -f /usr/local/bin/tamk

# Remover configurações do usuário (opcional)
rm -rf ~/.tamk/
```

## 🔄 Atualização

Para atualizar para uma versão mais recente:

```bash
# Verificar atualizações disponíveis
tamk update

# Ou reinstalar manualmente
sudo rm -rf /opt/tamk /usr/local/bin/tamk
sudo ./install.sh [novo-arquivo.zip]
```

## 🆘 Solução de Problemas

### Problema: Permissão Negada ao Instalar

```bash
# Solução: Use sudo
sudo ./install.sh
```

### Problema: zip: command not found

```bash
# Instale unzip
sudo apt install unzip          # Debian/Ubuntu
sudo dnf install unzip          # Fedora/RHEL
sudo pacman -S unzip            # Arch
```

### Problema: tamk command not found após instalação

```bash
# Verifique se o link existe
ls -la /usr/local/bin/tamk

# Se não existir, reinstale
sudo ./install.sh
```

### Problema: /usr/local/bin não está no PATH

```bash
# Tente com caminho completo
/usr/local/bin/tamk version

# Se funcionar, adicione ao PATH em ~/.bashrc ou ~/.zshrc
export PATH="/usr/local/bin:$PATH"
```

## 📝 Distribuições Testadas

- ✅ Ubuntu 20.04+
- ✅ Debian 11+
- ✅ Fedora 35+
- ✅ Arch Linux
- ✅ Linux Mint 20+

## 📞 Suporte

Se encontrar problemas:

1. Verifique a versão: `tamk version`
2. Veja os logs: `tamk --verbose --help`
3. Consulte a documentação: `/opt/tamk/README.md`
4. Reporte issues em: https://github.com/Shadw-Developer/tamk/issues

---

<div align="center">
  <sub>TAMK v2026.3.0-HMR | Instalação Linux</sub>
</div>