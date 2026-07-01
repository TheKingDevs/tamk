@echo off
REM =============================================================================
REM T.A.M.K - Termux APK Manager Kit • Windows Installer
REM Version: 1.0.0
REM Supports: Windows 10/11 (x64)
REM =============================================================================

setlocal enabledelayedexpansion

set VERSION=1.0.0
set INSTALL_DIR=%USERPROFILE%\.tamk
set BIN_DIR=%INSTALL_DIR%\bin

echo.
echo ========================================
echo  T.A.M.K Installer v%VERSION%
echo  Windows (x64)
echo ========================================
echo.

REM Check Go
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go not found
    echo   Download from: https://go.dev/dl/
    echo   Then run this installer again
    exit /b 1
)
echo [OK] Go found

REM Create directories
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

REM Check if already installed
if exist "%BIN_DIR%\tamk.exe" (
    for /f "tokens=*" %%i in ('"%BIN_DIR%\tamk.exe" version 2^>nul ^| findstr /r "v[0-9]"') do set CURRENT_VER=%%i
    echo Current version: !CURRENT_VER!
)

echo.
echo Building T.A.M.K v%VERSION%...

REM Build
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w -X 'github.com/TheKingDevs/tamk/internal/config.Version=%VERSION%'" -o "%BIN_DIR%\tamk.exe" .\cmd\tamk

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Build failed
    exit /b 1
)

echo [OK] Binary installed at %BIN_DIR%\tamk.exe

REM Deploy assets
echo.
echo [INFO] Deploying assets...
if not exist "%INSTALL_DIR%\templates" mkdir "%INSTALL_DIR%\templates"
xcopy /E /I /Y "%~dp0..\templates" "%INSTALL_DIR%\templates" >nul 2>&1
echo [OK] Templates deployed
if not exist "%INSTALL_DIR%\libs" mkdir "%INSTALL_DIR%\libs"
copy /Y "%~dp0..\libs\libraries.json" "%INSTALL_DIR%\libs\" >nul 2>&1
echo [OK] Libraries deployed
if not exist "%INSTALL_DIR%\configs" mkdir "%INSTALL_DIR%\configs"
copy /Y "%~dp0..\configs\*" "%INSTALL_DIR%\configs\" >nul 2>&1
echo [OK] Configs deployed

REM Add to PATH (current session)
set "PATH=%BIN_DIR%;%PATH%"

REM Add to user PATH (persistent)
setx PATH "%BIN_DIR%;%PATH%" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] PATH configured
)

REM Create TAMK_HOME environment variable
setx TAMK_HOME "%INSTALL_DIR%" >nul 2>&1

echo.
echo ========================================
echo  INSTALLATION COMPLETE
echo ========================================
echo.
echo   Next steps:
echo     tamk setup     Download SDK + keystore
echo     tamk create    Create a project
echo     tamk build     Build APK
echo.
echo   https://github.com/TheKingDevs/tamk
echo.

endlocal
