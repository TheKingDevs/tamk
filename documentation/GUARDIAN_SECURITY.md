# T.A.M.K Guardian Security System

> **T.A.M.K (Termux APK Manager Kit)** — Security documentation for APK protection

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Security Layers](#security-layers)
4. [Usage](#usage)
5. [Configuration](#configuration)
6. [Asset Encryption](#asset-encryption)
7. [Security Classes](#security-classes)
8. [Threat Detection](#threat-detection)
9. [Best Practices](#best-practices)
10. [Limitations](#limitations)

---

## Overview

The T.A.M.K Guardian Security System provides **multi-layered APK protection** without requiring NDK/native code compilation. It uses pure Kotlin security checks combined with Go-side asset encryption to create a comprehensive defense-in-depth strategy.

### Key Features

- **Anti-Debug**: Detects debugger attachment via multiple methods
- **Anti-Root**: Identifies rooted devices and root management apps
- **Anti-Frida**: Detects Frida hooking framework and similar tools
- **Anti-Emulator**: Identifies virtual/emulated environments
- **Asset Encryption**: AES-256-GCM encryption for sensitive files
- **RASP**: Runtime Application Self-Protection with periodic checks

### Architecture Decision

T.A.M.K uses **direct kotlinc compilation** (no Gradle), which means:
- ❌ R8/ProGuard obfuscation not available
- ❌ NDK/native code compilation not supported
- ✅ Pure Kotlin security checks (reversible but layered)
- ✅ Go-side asset encryption (strong protection)
- ✅ Template-based injection (automatic at build time)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  GUARDIAN SECURITY ARCHITECTURE                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │  tamk build     │───▶│  Go Build       │                │
│  │  --guardian     │    │  Pipeline       │                │
│  └─────────────────┘    └────────┬────────┘                │
│                                  │                          │
│                    ┌─────────────┴─────────────┐            │
│                    │                           │            │
│                    ▼                           ▼            │
│  ┌─────────────────────────┐  ┌─────────────────────────┐  │
│  │  Asset Encryption       │  │  Template Injection     │  │
│  │  (AES-256-GCM)          │  │  (Kotlin classes)       │  │
│  │                         │  │                         │  │
│  │  - Encrypts assets/     │  │  - GuardianBridge.kt    │  │
│  │  - Magic header         │  │  - RASPSecurityModule.kt│  │
│  │  - Key from password    │  │  - Package injection    │  │
│  └─────────────────────────┘  └─────────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Runtime Protection (Kotlin)                        │   │
│  │                                                     │   │
│  │  GuardianBridge:                                    │   │
│  │    - isDebuggerAttached()                          │   │
│  │    - isDeviceRooted()                              │   │
│  │    - detectFrida()                                 │   │
│  │    - isEmulator()                                  │   │
│  │    - verifyAppSignature()                          │   │
│  │    - verifyDexIntegrity()                          │   │
│  │                                                     │   │
│  │  RASPSecurityModule:                                │   │
│  │    - Periodic security checks                      │   │
│  │    - Threat callback system                        │   │
│  │    - Lifecycle management                          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Security Layers

### Layer 1: Anti-Debug Detection

**Purpose**: Detect if the app is being debugged or analyzed.

| Method | Description | Difficulty to Bypass |
|--------|-------------|---------------------|
| `Debug.isDebuggerConnected()` | Android API check | Easy (hookable) |
| TracerPid check | Reads `/proc/self/status` | Medium |
| Timing anomaly detection | Multiple CPU-bound checks with std dev analysis | Hard |
| System property check | `ro.debuggable` and `ro.secure` props | Medium |
| JDWP thread detection | Checks for Java Debug Wire Protocol threads | Hard |

**Timing Anomaly Detection**:
```kotlin
// Performs 5 rounds of CPU-intensive calculations
// Calculates standard deviation of timings
// High std dev (>5ms) indicates timing manipulation
// Any single timing >50ms is suspicious
```

### Layer 2: Anti-Root Detection

**Purpose**: Identify rooted devices and root management apps.

| Layer | Check | Description |
|-------|-------|-------------|
| 1 | File paths | Checks 18+ common root file locations |
| 2 | /data/data access | Tests if app can read other apps' data |
| 3 | Root apps | Scans for Magisk, SuperSU, etc. via `pm list packages` |
| 4 | Magisk Hide | Checks Magisk module directories |

**Root Indicators Checked**:
- `/system/bin/su`, `/system/xbin/su`, `/sbin/su`
- `/data/adb/magisk`, `/sbin/.magisk`
- `/system/app/Superuser.apk`, `/system/app/Kinguser.apk`
- `/system/bin/busybox`, `/system/etc/init.d`
- And more...

### Layer 3: Anti-Frida Detection

**Purpose**: Detect Frida hooking framework and similar tools.

| Method | Description |
|--------|-------------|
| Process scan | `ps | grep frida` |
| Maps analysis | Checks `/proc/self/maps` for frida/gadget/xposed |
| Port scanning | Scans ports 27042-27050 (Frida defaults) |
| TCP table | Checks `/proc/net/tcp` for hex port patterns |
| File detection | Checks `/data/local/tmp/frida-server` |

**Frida Indicators**:
- Library names: `frida`, `gadget`, `linjector`, `xposed`, `substrate`
- Ports: 27042, 27043, 27046-27050
- Files: `frida-server`, `re.frida.server`

### Layer 4: Anti-Emulator Detection

**Purpose**: Identify virtual/emulated environments.

| Check | Description |
|-------|-------------|
| Build properties | MODEL, PRODUCT, HARDWARE, BRAND, MANUFACTURER |
| Emulator files | `/dev/socket/qemud`, `/dev/qemu_pipe`, etc. |
| QEMU property | `Build.QEMU_EMU` field |
| CPU info | `/proc/cpuinfo` for "qemu" or "virtual" |

**Emulator Indicators**:
- Properties: `goldfish`, `ranchu`, `generic`, `sdk_gphone`, `emulator`
- Files: `qemud`, `qemu_pipe`, `qemu-props`
- CPU: QEMU or virtual CPU features

---

## Usage

### Build with Guardian Protection

```bash
# Basic build with security
tamk build -p <password> --guardian

# Verbose output to see security steps
tamk build -p <password> --guardian -V
```

### Build Output

```
[STEP] Applying Guardian security protection...
[INFO] Encrypted assets count=3
[STEP] Compiling resources (AAPT2)...
[STEP] Linking resources (AAPT2)...
[DEBUG] Security templates injected dir=...
[STEP] Compiling Kotlin sources...
[STEP] Converting to DEX (D8)...
[STEP] Packaging DEX into APK...
[STEP] Aligning (zipalign)...
[STEP] Signing APK...
[ OK ] Build completed
```

### Verify Security Features

```bash
# Decompile APK
jadx -d decompiled app.apk

# Check for security classes
find decompiled -name "*.java" | grep -E "(Guardian|RASP)"

# Check for encrypted assets
ls decompiled/resources/assets/*.enc

# Check for API key exposure
grep -r "api_key\|secret\|password" decompiled/
```

---

## Configuration

### Security Levels

| Level | Features Enabled | Use Case |
|-------|------------------|----------|
| `none` | No security | Development |
| `basic` | Anti-debug only | Testing |
| `standard` | + Anti-root, Anti-Frida, Integrity | Internal apps |
| `maximum` | + Anti-emulator, RASP | Production |

### tamk.config Security Section

```json
{
  "type": "webapp",
  "name": "MyApp",
  "version": "1.0.0",
  "security": {
    "enabled": true,
    "level": "maximum",
    "guardian": {
      "anti_debug": true,
      "anti_root": true,
      "anti_frida": true,
      "anti_emulator": true,
      "integrity": true,
      "rasp": true
    }
  }
}
```

---

## Asset Encryption

### Encryption Details

| Property | Value |
|----------|-------|
| Algorithm | AES-256-GCM |
| Key Derivation | SHA-256 (password + salt) |
| Salt Size | 16 bytes |
| IV Size | 12 bytes |
| Magic Header | `TAMK_ENC_1` |

### Encrypted File Format

```
┌────────────────┬────────┬────────┬──────────────────┐
│ Magic Header   │ Salt   │ IV     │ Ciphertext       │
│ (10 bytes)     │ (16 B) │ (12 B) │ (variable)       │
└────────────────┴────────┴────────┴──────────────────┘
```

### Files Encrypted

- `index.html` → `index.html.enc`
- `css/styles.css` → `css/styles.css.enc`
- `js/app.js` → `js/app.js.enc`

### Decryption at Runtime

The Kotlin code decrypts assets using the same password:

```kotlin
fun decryptAsset(encryptedData: ByteArray, password: String): ByteArray {
    // Extract salt, IV, ciphertext from TAMK_ENC_1 format
    // Derive key from password + salt
    // Decrypt with AES-256-GCM
    // Return plaintext
}
```

---

## Security Classes

### GuardianBridge.kt

**Location**: `templates/security/kotlin/GuardianBridge.kt.tmpl`

**Package**: `{{PACKAGE}}.security`

**Methods**:

| Method | Return | Description |
|--------|--------|-------------|
| `initialize(context)` | `Unit` | Initializes security module |
| `isDebuggerAttached()` | `Boolean` | Detects debugger |
| `isDeviceRooted()` | `Boolean` | Detects root |
| `detectFrida()` | `Boolean` | Detects Frida |
| `isEmulator()` | `Boolean` | Detects emulator |
| `verifyAppSignature(context, hash)` | `Boolean` | Verifies APK signature |
| `verifyDexIntegrity(context, hash)` | `Boolean` | Verifies DEX integrity |
| `performSecurityCheck(context)` | `SecurityThreat?` | Runs all checks |
| `getObfuscatedString(parts, xorKey)` | `String` | Deobfuscates strings |

**SecurityThreat Enum**:

```kotlin
enum class SecurityThreat {
    DEBUGGER,   // Critical - immediate response
    FRIDA,      // Critical - immediate response
    ROOT,       // High - degraded functionality
    EMULATOR,   // Medium - limited features
    TAMPERED    // Critical - integrity violation
}
```

### RASPSecurityModule.kt

**Location**: `templates/security/kotlin/RASPSecurityModule.kt.tmpl`

**Package**: `{{PACKAGE}}.security`

**Methods**:

| Method | Description |
|--------|-------------|
| `start(context, intervalMs, onThreat)` | Starts RASP protection |
| `stop()` | Stops RASP protection |
| `onResume()` | Resumes checks (call in Activity) |
| `onPause()` | Pauses checks (call in Activity) |

**Usage in Activity**:

```kotlin
class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        RASPSecurityModule.start(this, 5000L) { threat ->
            when (threat) {
                GuardianBridge.SecurityThreat.DEBUGGER -> finish()
                GuardianBridge.SecurityThreat.FRIDA -> finish()
                GuardianBridge.SecurityThreat.ROOT -> {
                    // Disable sensitive features
                }
                GuardianBridge.SecurityThreat.EMULATOR -> {
                    // Show warning
                }
                GuardianBridge.SecurityThreat.TAMPERED -> finish()
            }
        }
    }
    
    override fun onResume() {
        super.onResume()
        RASPSecurityModule.onResume()
    }
    
    override fun onPause() {
        super.onPause()
        RASPSecurityModule.onPause()
    }
}
```

---

## Threat Detection

### Detection Matrix

| Threat | Detection Methods | Response |
|--------|-------------------|----------|
| Debugger | API, TracerPid, Timing, Props, JDWP | Critical: Exit |
| Frida | Process, Maps, Ports, TCP, Files | Critical: Exit |
| Root | Files, /data/data, Apps, Magisk | High: Degrade |
| Emulator | Props, Files, QEMU, CPU | Medium: Warn |
| Tampered | Signature, DEX hash | Critical: Exit |

### Response Strategies

```kotlin
// Option 1: Immediate exit
android.os.Process.killProcess(android.os.Process.myPid())

// Option 2: Delayed exit (harder to debug)
Thread {
    Thread.sleep((1000..5000).random().toLong())
    android.os.Process.killProcess(android.os.Process.myPid())
}.start()

// Option 3: Degraded functionality
fun onSecurityThreat(threat: SecurityThreat) {
    when (threat) {
        SecurityThreat.ROOT -> {
            // Disable payment features
            // Disable sensitive data access
        }
        SecurityThreat.EMULATOR -> {
            // Show warning message
            // Limit functionality
        }
        else -> finish()
    }
}
```

---

## Best Practices

### 1. Always Use --guardian for Production

```bash
# Development (fast, no security)
tamk build -p senha

# Production (secure)
tamk build -p senha --guardian
```

### 2. Protect Sensitive Data

- Never hardcode API keys in source code
- Store secrets in encrypted assets
- Use the encryption system for sensitive configs

### 3. Implement Proper Response

```kotlin
// Don't show clear error messages
// ❌ "Security violation detected: debugger"
// ✅ Silent exit or degraded functionality

// Don't reveal security implementation
// ❌ Log: "GuardianBridge.isDebuggerAttached() returned true"
// ✅ Silent response
```

### 4. Test Security Features

```bash
# Test with Frida attached
frida-server &
tamk build -p senha --guardian
# Should detect Frida

# Test on rooted device
tamk build -p senha --guardian
# Should detect root

# Test with debugger
tamk build -p senha --guardian
# Should detect debugger
```

### 5. Layer Your Security

- Use `--guardian` flag for build-time protection
- Implement RASP for runtime protection
- Add certificate pinning for network security
- Use server-side validation for critical operations

---

## Limitations

### What Guardian CAN Do

- ✅ Detect common debugging tools
- ✅ Detect rooted devices
- ✅ Detect Frida and similar frameworks
- ✅ Detect emulators
- ✅ Encrypt sensitive assets
- ✅ Verify app integrity
- ✅ Provide runtime protection

### What Guardian CANNOT Do

- ❌ Prevent all reverse engineering (determined attackers)
- ❌ Protect against custom ROMs
- ❌ Prevent memory dumps on rooted devices
- ❌ Replace server-side security validation
- ❌ Guarantee 100% tamper resistance

### Security Philosophy

> **"Security is about making attacks expensive, not impossible."**

The Guardian system adds multiple layers of protection that:
1. Increase the time and skill required for attacks
2. Deter casual tampering attempts
3. Provide detection and response capabilities
4. Protect sensitive data at rest and in transit

For maximum security, always combine:
- Client-side protection (Guardian)
- Server-side validation
- Secure communication (TLS)
- Regular security audits

---

## File Structure

```
templates/security/
└── kotlin/
    ├── GuardianBridge.kt.tmpl      # Main security class
    └── RASPSecurityModule.kt.tmpl  # Runtime protection

internal/
├── domain/entity/
│   └── security.go                 # SecurityConfig entity
├── usecase/
│   ├── project_build.go           # Guardian build integration
│   └── asset_encryptor.go         # AES-256-GCM encryption
└── repository/filesystem/
    └── project_repository.go       # Security config persistence
```

---

## Changelog

### v1.0.0 (2026-06-21)

- Initial Guardian security system implementation
- Kotlin-based security checks (no NDK required)
- AES-256-GCM asset encryption
- `--guardian` CLI flag
- Security levels: none, basic, standard, maximum
- RASP module with periodic checks
- Anti-debug, anti-root, anti-frida, anti-emulator detection
- Timing anomaly detection
- Signature and DEX integrity verification

---

<div align="center">
  <sub><strong>T.A.M.K Guardian Security v1.0</strong></sub><br>
  <sub>Termux APK Manager Kit — Security Documentation</sub>
</div>
