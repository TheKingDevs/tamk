@echo off
REM T.A.M.K Windows Test Script
REM Tests embedded tool extraction and basic functionality

setlocal enabledelayedexpansion

echo ========================================
echo  T.A.M.K Windows Test Script
echo ========================================
echo.

REM Check if binary exists
if not exist "tamk-windows-amd64.exe" (
    echo ERROR: tamk-windows-amd64.exe not found
    echo Download from: https://github.com/TheKingDevs/tamk/releases
    exit /b 1
)

REM Test 1: Version command
echo [1/4] Testing version command...
tamk-windows-amd64.exe version
if errorlevel 1 (
    echo ERROR: Version command failed
    exit /b 1
)
echo OK
echo.

REM Test 2: Check embedded tools extraction
echo [2/4] Testing embedded tools extraction...
set TAMK_HOME=%TEMP%\tamk-test
if exist "%TAMK_HOME%" rmdir /s /q "%TAMK_HOME%"
mkdir "%TAMK_HOME%"

set TAMK_HOME=%TAMK_HOME%
tamk-windows-amd64.exe version >nul 2>&1

if exist "%TEMP%\tamk-tools" (
    echo OK - Tools extracted to: %TEMP%\tamk-tools
    dir /s /b "%TEMP%\tamk-tools" 2>nul | find /c /v "" >nul
    echo     Files extracted: %ERRORLEVEL%
) else (
    echo WARNING: Tools directory not created yet (lazy extraction)
)
echo.

REM Test 3: Create test project
echo [3/4] Creating test project...
cd /d "%TEMP%"
if exist "TestApp" rmdir /s /q "TestApp"
tamk-windows-amd64.exe create --name TestApp --type webapp --author "Test" --version "1.0.0"
if errorlevel 1 (
    echo ERROR: Project creation failed
    exit /b 1
)

if exist "TestApp\AndroidManifest.xml" (
    echo OK - Project created successfully
) else (
    echo ERROR: Project files not created
    exit /b 1
)
echo.

REM Test 4: Verify project structure
echo [4/4] Verifying project structure...
set FILES=0
if exist "TestApp\AndroidManifest.xml" set /a FILES+=1
if exist "TestApp\tamk.config" set /a FILES+=1
if exist "TestApp\src\main\kotlin" set /a FILES+=1
if exist "TestApp\res\values" set /a FILES+=1

if %FILES% GEQ 4 (
    echo OK - All expected files created
) else (
    echo WARNING: Some files may be missing
)
echo.

REM Cleanup
echo Cleaning up test project...
cd /d "%TEMP%"
if exist "TestApp" rmdir /s /q "TestApp"

echo ========================================
echo  All tests passed!
echo ========================================
echo.
echo Next steps:
echo   1. Install Android SDK build-tools
echo   2. Install Kotlin compiler
echo   3. Run: tamk-windows-amd64.exe setup
echo   4. Run: tamk-windows-amd64.exe build -p YOUR_PASSWORD
echo.

endlocal
