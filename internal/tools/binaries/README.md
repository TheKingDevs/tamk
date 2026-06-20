# Windows Embedded Binaries

This directory contains the Windows binaries embedded into the TAMK binary via `go:embed`.

## Required Files

| File | Source | Size |
|------|--------|------|
| `aapt2.exe` | Android SDK Build-Tools | ~1.5MB |
| `apksigner.jar` | Android SDK Build-Tools | ~2MB |
| `zipalign.exe` | Android SDK Build-Tools | ~1MB |
| `d8.jar` | Android SDK Build-Tools | ~1.5MB |
| `kotlinc.zip` | Kotlin Compiler | ~80MB |

## How to Obtain

### Option 1: Download from Android SDK

```bash
# Install Android SDK command-line tools
# Then download build-tools:
sdkmanager "build-tools;34.0.0"

# Copy files from:
# $ANDROID_HOME/build-tools/34.0.0/aapt2.exe
# $ANDROID_HOME/build-tools/34.0.0/lib/apksigner.jar
# $ANDROID_HOME/build-tools/34.0.0/zipalign.exe
# $ANDROID_HOME/build-tools/34.0.0/lib/d8.jar
```

### Option 2: Download Kotlin Compiler

```bash
# Download from https://kotlinlang.org/docs/command-line.html
# Extract kotlinc.zip from the distribution
```

## Build Impact

- **Linux/macOS/Termux builds**: These files are NOT embedded (build tag `!windows`)
- **Windows builds**: These files ARE embedded, increasing binary size to ~280MB

## Security Note

These binaries are third-party tools. Verify checksums before embedding:

```bash
sha256sum binaries/windows/*
```
