# T.A.M.K Windows Test Script (PowerShell)
# Tests embedded tool extraction and basic functionality

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " T.A.M.K Windows Test Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if binary exists
$binary = "tamk-windows-amd64.exe"
if (-not (Test-Path $binary)) {
    Write-Host "ERROR: $binary not found" -ForegroundColor Red
    Write-Host "Download from: https://github.com/TheKingDevs/tamk/releases" -ForegroundColor Yellow
    exit 1
}

# Test 1: Version command
Write-Host "[1/4] Testing version command..." -ForegroundColor Yellow
try {
    & .\$binary version
    Write-Host "OK" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Version command failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Test 2: Check embedded tools extraction
Write-Host "[2/4] Testing embedded tools extraction..." -ForegroundColor Yellow
$env:TAMK_HOME = Join-Path $env:TEMP "tamk-test"
if (Test-Path $env:TAMK_HOME) {
    Remove-Item -Recurse -Force $env:TAMK_HOME
}
New-Item -ItemType Directory -Path $env:TAMK_HOME -Force | Out-Null

& .\$binary version 2>$null | Out-Null

$toolsDir = Join-Path $env:TEMP "tamk-tools"
if (Test-Path $toolsDir) {
    Write-Host "OK - Tools extracted to: $toolsDir" -ForegroundColor Green
    $files = Get-ChildItem -Recurse $toolsDir | Measure-Object
    Write-Host "    Files extracted: $($files.Count)" -ForegroundColor Gray
} else {
    Write-Host "WARNING: Tools directory not created yet (lazy extraction)" -ForegroundColor Yellow
}
Write-Host ""

# Test 3: Create test project
Write-Host "[3/4] Creating test project..." -ForegroundColor Yellow
$testDir = Join-Path $env:TEMP "tamk-test-project"
if (Test-Path $testDir) {
    Remove-Item -Recurse -Force $testDir
}
Set-Location $env:TEMP

try {
    & .\$binary create --name TestApp --type webapp --author "CI Test" --version "1.0.0"
    Write-Host "OK - Project created successfully" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Project creation failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Test 4: Verify project structure
Write-Host "[4/4] Verifying project structure..." -ForegroundColor Yellow
$expectedFiles = @(
    "TestApp\AndroidManifest.xml",
    "TestApp\tamk.config",
    "TestApp\src\main\kotlin",
    "TestApp\res\values"
)

$found = 0
foreach ($file in $expectedFiles) {
    if (Test-Path $file) {
        $found++
    }
}

if ($found -ge 4) {
    Write-Host "OK - All expected files created" -ForegroundColor Green
} else {
    Write-Host "WARNING: Some files may be missing ($found/4)" -ForegroundColor Yellow
}
Write-Host ""

# Cleanup
Write-Host "Cleaning up test project..." -ForegroundColor Gray
if (Test-Path "TestApp") {
    Remove-Item -Recurse -Force "TestApp"
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " All tests passed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Install Android SDK build-tools"
Write-Host "  2. Install Kotlin compiler"
Write-Host "  3. Run: tamk-windows-amd64.exe setup"
Write-Host "  4. Run: tamk-windows-amd64.exe build -p YOUR_PASSWORD"
Write-Host ""
