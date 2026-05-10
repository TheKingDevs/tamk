# 📦 Guia de Instalação do TAMK

Bem-vindo ao T.A.M.K (Termux APK Manager Kit)! Este guia cobre a instalação em todos os sistemas operacionais suportados.

## 🎯 Selecione seu Sistema Operacional

### 🐧 Linux
**Para Ubuntu, Debian, Fedora, Arch e outras distribuições Linux.**

- **Download recomendado**: `tamk-proprietary-2026.3.0-HMR-linux.zip`
- **Método**: Script bash automático
- **Tempo de instalação**: ~2 minutos

👉 [Instruções Completas para Linux →](INSTALLATION_LINUX.md)

**Quick start:**
```bash
unzip tamk-proprietary-2026.3.0-HMR-linux.zip
cd tamk
sudo ./install.sh
tamk --version
```

---

### 🪟 Windows
**Para Windows 10 e Windows 11.**

- **Download recomendado**: `tamk-proprietary-2026.3.0-HMR-windows.zip`
- **Método**: Script batch ou manual
- **Tempo de instalação**: ~3 minutos
- **Requer**: Permissões de Administrador

👉 [Instruções Completas para Windows →](INSTALLATION_WINDOWS.md)

**Quick start:**
```cmd
REM Extrair arquivo
REM Clicar direito em install.bat → Executar como Administrador
REM Abrir novo Command Prompt
tamk --version
```

---

### 🍎 macOS
**Para macOS 10.15+ (Catalina, Big Sur, Monterey, Ventura, Sonoma).**

- **Download recomendado**: `tamk-proprietary-2026.3.0-HMR-macos.zip`
- **Métodos**: Script bash (como Linux)
- **Tempo de instalação**: ~2 minutos
- **Suporta**: Intel e Apple Silicon (M1/M2/M3)

👉 [Instruções Completas para macOS →](INSTALLATION_MACOS.md)

**Quick start:**
```bash
unzip tamk-proprietary-2026.3.0-HMR-macos.zip
cd tamk
chmod +x install.sh
sudo ./install.sh
tamk --version
```

---

## ✅ Verificar Instalação

Após instalar, confirme que TAMK está funcionando:

```bash
# Ver versão (funciona em todos os SOs)
tamk --version

# Ver ajuda
tamk --help

# Criar novo projeto
tamk --create
```

## 🗑️ Desinstalar TAMK

**Linux:**
```bash
sudo rm -rf /opt/tamk /usr/local/bin/tamk
```

**Windows:**
```cmd
del C:\Windows\System32\tamk.bat
rmdir /s /q "C:\Program Files\TAMK"
```

**macOS:**
```bash
sudo rm -rf /opt/tamk /usr/local/bin/tamk
```

## 🔄 Atualizar TAMK

```bash
# Verificar atualizações (todos os SOs)
tamk --update

# Ou instalar manualmente
# 1. Desinstale a versão atual
# 2. Extraia o novo pacote
# 3. Execute o script de instalação novamente
```

## 🆘 Solução de Problemas Rápida

| Problema | Solução |
|----------|--------|
| "comando não encontrado" | Feche e reabra o terminal |
| "permissão negada" | Use `sudo` no Linux/macOS, execute como Admin no Windows |
| "antivírus bloqueia" | Adicione TAMK à lista de exclusão |
| Zip não extrai | Use 7-Zip, WinRAR, ou ferramenta nativa |
| Script não funciona | Verifique permissões: `chmod +x install.sh` |

💡 **Para soluções detalhadas**, consulte o guia específico do seu SO.

## 📋 Requisitos de Sistema

### Mínimos
- **RAM**: 2GB
- **Disco**: 250MB livres
- **Processador**: Qualquer moderno

### Recomendados
- **RAM**: 8GB ou mais
- **Disco**: 1GB livres (para projetos Android)
- **Conexão**: Internet alta velocidade (pra download do SDK)

## 🌐 Downloads

Procure os arquivos nos locais:

- **Servidor oficial**: [https://seu-servidor.com/tamk/](https://seu-servidor.com/tamk/)
- **GitHub Releases**: [https://github.com/Shadw-Developer/tamk/releases](https://github.com/Shadw-Developer/tamk/releases)
- **Package Managers**: Disponível em Homebrew, apt, etc.

## 📝 Arquivos de Distribuição

```
tamk-proprietary-2026.3.0-HMR-linux.zip      # Linux
tamk-proprietary-2026.3.0-HMR-windows.zip    # Windows
tamk-proprietary-2026.3.0-HMR-macos.zip      # macOS
tamk-proprietary-2026.3.0-HMR-linux.tar.gz   # Linux alternativo (tar)
```

## 🔐 Segurança

TAMK sempre:
- ✅ Fornece o código fonte verificável
- ✅ Usa certificados SSL para downloads
- ✅ Valida integridade via SHA256
- ✅ Respeita privacidade do usuário
- ✅ É compatível com antivírus conhecidos

Para verificar autenticidade:

```bash
# Comparar SHA256
sha256sum tamk-proprietary-2026.3.0-HMR-linux.zip
# Comparar com hash oficial em SHA256SUMS
```

## 📞 Suporte e Contribuição

- **Issues**: [GitHub Issues](https://github.com/Shadw-Developer/tamk/issues)
- **Discussões**: [GitHub Discussions](https://github.com/Shadw-Developer/tamk/discussions)
- **Pull Requests**: [GitHub PRs](https://github.com/Shadw-Developer/tamk/pulls)

## 📚 Documentação Relacionada

- [Quick Start Guide](QUICKSTART.md) - Começar com TAMK em 5 minutos
- [User Guide](DEV_GUIDE.md) - Guia completo de desenvolvedor
- [Architecture](ARCHITECTURE.md) - Arquitetura interna
- [FAQ](FAQ.md) - Perguntas Frequentes

## 🎉 Próximos Passos

Após instalar, você pode:

1. **Criar seu primeiro projeto**
   ```bash
   tamk --create
   ```

2. **Explorar templates**
   - UI/APK (Android nativo)
   - Console (aplicação CLI)
   - WebApp (híbrido com WebView)

3. **Compilar e testar**
   ```bash
   tamk --build -p sua-senha
   tamk --install
   ```

4. **Modo desenvolvimento**
   ```bash
   tamk --dev
   ```

---

<div align="center">
  <p><strong>TAMK v2026.3.0-HMR</strong></p>
  <p>Termux APK Manager Kit - Desenvolvendo Android em qualquer lugar</p>
  <p><a href="https://github.com/Shadw-Developer/tamk">GitHub</a> • <a href="https://seu-site.com">Website</a> • <a href="https://discord.gg/seu-servidor">Discord</a></p>
</div>