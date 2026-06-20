# T.A.M.K ToolManager Architecture

## Overview

The ToolManager implements a "Lazy Extraction" pattern for embedded tools:
- **Linux/macOS/Termux**: Use system tools (lightweight, updatable)
- **Windows**: Auto-extract embedded tools from binary (zero config)

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    BuildProjectUseCase                       │
│  (executes build pipeline: aapt2 → kotlinc → d8 → sign)    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      ToolManager                            │
│  - Resolves tool paths based on platform                    │
│  - Lazy extracts embedded tools on first use                │
│  - Caches extracted paths                                   │
└─────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  SystemTools  │   │ EmbeddedTools │   │  CacheManager │
│  (Unix-like)  │   │  (Windows)    │   │  (temp dir)   │
└───────────────┘   └───────────────┘   └───────────────┘
```

## Directory Structure

```
internal/
├── tools/
│   ├── manager.go          # ToolManager interface + implementation
│   ├── system.go           # SystemTools (Unix-like platforms)
│   ├── embedded.go         # EmbeddedTools (Windows extraction)
│   ├── cache.go            # CacheManager for extracted tools
│   └── embed.go            # go:embed directives (Windows only)
├── binaries/
│   └── windows/
│       ├── aapt2.exe
│       ├── apksigner
│       ├── zipalign
│       ├── d8.jar
│       └── kotlinc/
```

## ToolManager Interface

```go
type ToolManager interface {
    // AAPT2 returns path to aapt2 binary
    AAPT2(ctx context.Context) (string, error)
    
    // ApkSigner returns path to apksigner script/binary
    ApkSigner(ctx context.Context) (string, error)
    
    // Zipalign returns path to zipalign binary
    Zipalign(ctx context.Context) (string, error)
    
    // D8 returns path to d8.jar or d8 binary
    D8(ctx context.Context) (string, error)
    
    // KotlinCompiler returns path to kotlinc script/binary
    KotlinCompiler(ctx context.Context) (string, error)
    
    // SDKJar returns path to android.jar
    SDKJar() (string, error)
    
    // Cleanup removes extracted temporary files
    Cleanup() error
}
```

## Lazy Extraction Flow

1. **First Call**: `tm.AAPT2(ctx)`
   - Check if system tool exists → return system path
   - Check cache → return cached path
   - Extract from embedded → cache → return path

2. **Subsequent Calls**: Return cached path immediately

3. **Cleanup**: On `tamk clean` or exit

## Platform Behavior

| Platform | Strategy | Binary Size |
|----------|----------|-------------|
| Linux amd64 | System tools | ~25MB |
| Linux arm64 | System tools | ~25MB |
| macOS arm64 | System tools | ~25MB |
| Android arm64 | System tools (pkg) | ~25MB |
| Windows amd64 | Embedded tools | ~280MB |

## Build Tags

```go
//go:build !windows
// system.go - uses exec.LookPath

//go:build windows
// embedded.go - uses go:embed extraction
```
