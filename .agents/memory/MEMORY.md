# TAMK Project Memory
_Maintained by MiMoCode agent. Updated after significant changes._

## Project Identity
- **Name**: T.A.M.K (Termux APK Manager Kit)
- **Version**: 2026.3.0-HMR
- **Language**: Go 1.26+ (core), Kotlin (Android templates)
- **Architecture**: Clean Architecture v4 with Dependency Rule

## Session Log
- **2026-06-20**: T2T analysis completed. Fixed 25 discrepancies in AGENTS.md, 8 in documentation/ files, 1 Makefile bug. Created `doc-audit` skill.
- **2026-06-20**: Applied clean code fixes. Created 7 GitHub Actions workflows.
- **2026-06-20**: Implemented ToolManager Lazy Extraction ecosystem. Updated project_build.go to use ToolManager.

## Applied Clean Code Fixes (2026-06-20)
- `development.go`: Removed misleading unimplemented command log, extracted `findAPK()` helper
- `project_build.go`: Fixed redundant error logging in `failedResult()`
- `install.go`: Removed redundant parentheses in `FindAPKs()`

## GitHub Actions Workflows (2026-06-20)
| Workflow | File | Purpose |
|:--|:--|:--|
| CI | `ci.yml` | Lint, test, build on push/PR |
| Release | `release.yml` | Multi-platform build + GitHub release |
| Tests T2T | `tests-tt.yml` | Doc-code cross-verification |
| Tags | `tags.yml` | Tag validation + creation |
| Build | `build.yml` | Multi-platform matrix build |
| Security | `security.yml` | govulncheck, staticcheck, golangci-lint |
| Maintenance | `maintenance.yml` | Weekly dependency updates |

## Platform Support
| Platform | GOOS/GOARCH | Binary Name | Tool Strategy |
|:--|:--|:--|:--|
| Linux x64 | linux/amd64 | `tamk-linux-amd64` | System tools |
| Linux ARM64 | linux/arm64 | `tamk-linux-arm64` | System tools |
| **Android ARM64** | **android/arm64** | **`tamk-android-arm64`** | **System tools (pkg)** |
| macOS ARM64 | darwin/arm64 | `tamk-darwin-arm64` | System tools |
| Windows x64 | windows/amd64 | `tamk-windows-amd64.exe` | **Embedded tools** |

## ToolManager Architecture (2026-06-20)
- **Interface**: `internal/tools/manager.go`
- **System tools**: `internal/tools/system.go` (Linux/macOS/Android)
- **Embedded tools**: `internal/tools/embedded.go` (Windows, go:embed)
- **Stub**: `internal/tools/embedded_stub.go` (non-Windows fallback)
- **Binaries**: `internal/tools/binaries/windows/` (aapt2, apksigner, zipalign, d8, kotlinc)
- **Strategy**: Lazy extraction on first use, cached in temp dir

### Windows Embedded Tools (baixados)
| Arquivo | Tamanho | Fonte |
|:--|:--|:--|
| aapt2.exe | 4.4MB | Android SDK Build-Tools r34 |
| apksigner.jar | 1MB | Android SDK Build-Tools r34 |
| zipalign.exe | 799KB | Android SDK Build-Tools r34 |
| d8.jar | 15MB | Android SDK Build-Tools r34 |
| kotlinc.zip | 82MB | Kotlin Compiler 2.0.21 |
| **Total** | **~103MB** | |

### Windows Testing (2026-06-20)
- **Cross-compile**: `GOOS=windows GOARCH=amd64 go build` → 118MB binary ✓
- **GitHub Actions**: `test-windows.yml` com runners windows-2022
- **Scripts locais**: `scripts/test-windows.bat` e `scripts/test-windows.ps1`
- **Limitação Termux**: Sem Wine, sem teste real Windows local
- **Solução**: Usar GitHub Actions para teste em runner Windows real

## Termux Support
- Install script: `scripts/install-termux.sh`
- Binary: `tamk-android-arm64`
- Install path: `$PREFIX/bin/tamk`

## Key Metrics
- Entity types: 11 (Project, ProjectType, WebContentMode, BuildResult, BuildPhase, BuildCache, Template, TemplateMapping, Keystore, UpdateInfo, UpdateLevel)
- Error sentinels: 15 + BuildError struct
- Watch extensions: .html .css .js .json .png .jpg .jpeg .svg .webp .xml .kt
- Ignore dirs: .git node_modules secret .idea
- Template count: 11 webapp + 1 console + 6 ui_apk = 18 total
- Documentation: 21 .md files + VERSIONING.txt

## Architecture Decisions
- Clean Architecture v4 enforced via `internal/` package boundaries
- Template-oriented code generation (not rigid code gen)
- Factory pattern for project types (WebApp, UIAPK, Console)
- HMR bridge injected inline into index.html (not separate file)
- Build cache uses SHA-256 hash + APK existence check

## Rules
- Execute `.agents/hooks/work-finished` after completing agent tasks
- User communicates in Portuguese (PT-BR)
- Code and comments in English (market standard)
