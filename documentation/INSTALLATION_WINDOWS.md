# 🪟 Instalação do TAMK no Windows

Guia completo para instalar o T.A.M.K (Termux APK Manager Kit) no Windows 10/11.

## 📋 Pré-requisitos

- Windows 10 ou Windows 11
- Permissões de Administrador
- ~250MB de espaço em disco livre
- 7-Zip, WinRAR ou ferramenta similar (ou use Windows Subsystem for Linux)

## 🚀 Instalação Rápida (GUI)

### 1. Download do Pacote

Baixe o arquivo ZIP mais recente:

```
https://seu-servidor.com/tamk-proprietary-2026.3.0-HMR-windows.zip
```

### 2. Extrair o Arquivo

Clique com botão direito no arquivo ZIP → **Extrair Tudo...**

Ou use 7-Zip:
- Clique direito → 7-Zip → Extrair para "tamk"

### 3. Instalar Globalmente

Dentro da pasta extraída, crie um arquivo `install.bat`:

```batch
@echo off
REM TAMK Installer para Windows
setlocal enabledelayedexpansion

REM Verificar permissões de admin
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Erro: Execute como Administrador
    timeout /t 5
    exit /b 1
)

echo [INFO] Iniciando instalacao do TAMK...

REM Criar diretorio de instalacao
if not exist "C:\Program Files\TAMK" mkdir "C:\Program Files\TAMK"

REM Copiar arquivos
echo [INFO] Copiando arquivos...
xcopy /E /I /Y tamk "C:\Program Files\TAMK\tamk"

REM Adicionar ao PATH (criar comando .bat em System32)
echo [INFO] Criando comando global...
(
    echo @echo off
    echo "C:\Program Files\TAMK\tamk\tamk.exe" %%*
) > "C:\Windows\System32\tamk.bat"

echo.
echo ========================================
echo   TAMK instalado com sucesso!
echo ========================================
echo.
echo Abra um novo Command Prompt e execute:
echo   tamk --help
echo.
pause
```

Duplo clique em `install.bat` → **Executar como Administrador**

### 4. Verificar Instalação

Abra um novo **Command Prompt** ou **PowerShell**:

```cmd
tamk --version
tamk --help
```

✅ Se os comandos funcionarem, está pronto!

## 📍 Locais de Instalação

Após a instalação:

- **Executável**: `C:\Program Files\TAMK\tamk\tamk.exe`
- **Comando**: `tamk.bat` em `C:\Windows\System32\`
- **Configuração**: `%USERPROFILE%\.tamk\`

## 🔧 Uso Global

```cmd
REM Criar novo projeto
tamk --create

REM Compilar projeto
tamk --build -p sua-senha

REM Modo desenvolvimento
tamk --dev

REM Ver opções
tamk --help
```

## 🗑️ Desinstalação

### Via GUI

1. Abra **Painel de Controle** → **Programas e Recursos**
2. Procure por "TAMK" (se instalado via MSI)
3. Clique em **Desinstalar**

### Via Command Prompt (Como Administrador)

```cmd
REM Remover comando global
del C:\Windows\System32\tamk.bat

REM Remover arquivos de instalação
rmdir /s /q "C:\Program Files\TAMK"

REM Remover configurações (opcional)
rmdir /s /q "%USERPROFILE%\.tamk"
```

## 🔄 Atualização

```cmd
REM Verificar atualizações
tamk --update

REM Ou reinstalar manualmente
REM 1. Desinstale a versão atual
REM 2. Extraia o novo pacote ZIP
REM 3. Execute install.bat novamente
```

## 🆘 Solução de Problemas

### Problema: "tamk não é reconhecido como comando interno"

**Solução 1**: Feche e reabra o Command Prompt/PowerShell

**Solução 2**: Adicione ao PATH manualmente:
- Clique direito em **Este PC** → **Propriedades**
- → **Configurações avançadas do sistema**
- → **Variáveis de Ambiente...**
- Edite `PATH` e adicione: `C:\Program Files\TAMK\tamk`
- Clique OK e reinicie o terminal

**Solução 3**: Use o caminho completo:
```cmd
"C:\Program Files\TAMK\tamk\tamk.exe" --version
```

### Problema: "Acesso Negado" ao instalar

```cmd
REM Execute o install.bat como Administrador:
REM Clique direito em install.bat → Executar como Administrador
```

### Problema: Antivírus bloqueia a execução

- Adicione `C:\Program Files\TAMK\` à lista de exclusão do antivírus
- Ou desabilite temporariamente o antivírus durante a instalação

### Problema: Script de instalação não funciona

Use o **PowerShell** em vez de Command Prompt:

```powershell
# Como Administrador
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\install.ps1
```

## 💾 Instalação Portável (Sem instalar globalmente)

Se preferir usar sem instalar:

1. Extraia o ZIP para uma pasta (ex: `C:\TAMK`)
2. Use o caminho completo: `C:\TAMK\tamk\tamk.exe --version`
3. Ou crie um atalho no desktop

## 🔌 WSL (Windows Subsystem for Linux)

Se usará TAMK no WSL, siga o guia de **Instalação Linux**:

```bash
# Dentro do WSL
wsl
wget https://seu-servidor.com/tamk-proprietary-2026.3.0-HMR-linux.zip
unzip tamk-proprietary-2026.3.0-HMR-linux.zip
sudo ./install.sh
```

## 📝 Versões Testadas

- ✅ Windows 10 (build 19041+)
- ✅ Windows 11 (build 22000+)
- ✅ Windows Server 2019+

## 📞 Suporte

Se encontrar problemas:

1. Verifique a versão: `tamk --version`
2. Use PowerShell em vez de Command Prompt
3. Execute como Administrador
4. Consulte: `C:\Program Files\TAMK\README.md`
5. Reporte issues em: https://github.com/Shadw-Developer/tamk/issues

---

<div align="center">
  <sub>TAMK v2026.3.0-HMR | Instalação Windows</sub>
</div>