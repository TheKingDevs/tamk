# TAMK Project Memory
_Maintained by MiMoCode agent. Updated after significant changes._

## Project Identity
- **Name**: T.A.M.K (Termux APK Manager Kit)
- **Version**: 1.0.0
- **Language**: Go 1.26+ (core), Kotlin (Android templates)
- **Architecture**: Clean Architecture v4 with Dependency Rule

## Session Log
- **2026-06-20**: T2T analysis completed. Fixed 25 discrepancies in AGENTS.md, 8 in documentation/ files, 1 Makefile bug. Created `doc-audit` skill.
- **2026-06-20**: Applied clean code fixes. Created 7 GitHub Actions workflows.
- **2026-06-20**: Implemented ToolManager Lazy Extraction ecosystem. Updated project_build.go to use ToolManager.
- **2026-06-21**: Implemented "Setup Once, Run Forever" architecture. Added `tamk run` for native Kotlin execution. Reorganized templates into xml/kotlin/css/js subdirectories. Created platform-specific installers. Reset version to 1.0.0. Updated PRD/workflow.md.
- **2026-06-21**: Implemented "Setup Once, Run Forever" architecture. Reset version to 1.0.0. Created platform-specific installers. Implemented `tamk run` command. Fixed console project config.
- **2026-06-23**: Implemented HMR WebSocket server (gorilla/websocket on :8765). Fixed shell doRun() stub. Added tests for zip.go, errors.go, apk_finder.go. Wired update auto-install (git/go).
- **2026-06-23**: Performance optimizations: hash skip ignored dirs, template content cache (sync.Map), config singleton (sync.Once), SDKJar path caching. Updated PRD/performance.md.

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
| Test Windows | `test-windows.yml` | Windows CI on runner |

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
- **SDK**: `internal/tools/binaries/sdk/android.jar` (21MB)
- **Strategy**: "Setup Once, Run Forever" - extract to `~/.tamk/tools/{os}/` on first run

### Setup Once, Run Forever (2026-06-21)
- **Tools dir**: `~/.tamk/tools/{linux|darwin|windows}/`
- **Version marker**: `.version` file with TAMK version
- **Setup command**: `tamk setup` extracts tools on first run
- **Idempotent**: Re-running setup is fast (version check)
- **Binary size**: ~35MB (includes embedded android.jar)

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

## Installation Scripts (2026-06-21)
| Script | Platform | Features |
|:--|:--|:--|
| `scripts/install-termux.sh` | Termux (Android) | pkg install, PATH setup |
| `scripts/install-linux.sh` | Debian/Ubuntu/Arch/Fedora/Alpine | Auto-detect distro, package manager |
| `scripts/install-macos.sh` | macOS (Intel + Apple Silicon) | Homebrew integration |
| `scripts/install-windows.bat` | Windows 10/11 x64 | PATH, TAMK_HOME env vars |
| `setup-install.sh` | All platforms | Auto-detect and redirect |

## Run Command (2026-06-21)
- **UseCase**: `internal/usecase/run.go` — `RunUseCase`
- **Flow**: Find .kt file → compile with `kotlinc` to `.tamk-run/` → execute with `kotlin -cp`
- **Supports**: Main.kt auto-detection or specific file argument
- **CLI**: `newRunCmd(toolMgr)` wired in `internal/delivery/cli/root.go:57`

## Console Project Fix (2026-06-21)
- **Issue**: Console projects included `web_url`, `web_mode`, `min_sdk`, `target_sdk` in tamk.config
- **Fix**: Clear these fields on project entity before `Save()` in `project_create.go:81-86`
- **Console boilerplate**: `Main.kt.tmpl` now includes Calculator class and StringUtils object

## Termux Support
- Install script: `scripts/install-termux.sh`
- Binary: `tamk-android-arm64`
- Install path: `$PREFIX/bin/tamk`

## Key Metrics
- Entity types: 14 (Project, ProjectType, WebContentMode, BuildResult, BuildPhase, BuildCache, Template, TemplateMapping, Keystore, UpdateInfo, UpdateLevel, SecurityConfig, SecurityLevel, GuardianConfig)
- Error sentinels: 14 + BuildError struct
- Watch extensions: .html .css .js .json .png .jpg .jpeg .svg .webp .xml .kt
- Ignore dirs: .git node_modules secret .idea
- Template count: 11 webapp + 1 console + 6 ui_apk + 5 security = 23 total
- Documentation: 22 .md files + VERSIONING.txt

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

## Anti-Regression Rules (MANDATORY)
Before marking ANY task as done, agent MUST:
1. **Rebuild tamk**: `go build -o /root/.tamk/bin/tamk ./cmd/tamk`
2. **Create + Build ALL 3 project types** in `/root/testes/` using flags:
   - UIAPK: `tamk create -n RegTestUIAPK -t ui_apk -v 1.0.0 -a Agent`
   - WebAppLocal: `tamk create -n RegTestWebLocal -t webapp -v 1.0.0 -a Agent --web-mode internal`
   - WebAppExtern: `tamk create -n RegTestWebExtern -t webapp -v 1.0.0 -a Agent --web-mode external --url https://example.com`
3. **Build all 3**: `cd /root/testes/RegTest* && tamk build -p test123` — ALL must succeed
4. **Run Go tests**: `go test ./... -count=1` — ALL must pass
5. **Run go vet**: `go vet ./...` — 0 warnings
6. **Clean up**: `rm -rf /root/testes/RegTest*`
7. **NEVER commit if any check fails**

## Security Fixes (2026-06-20)
- **Keystore password validation**: Password now validated via `keytool -list` BEFORE build starts
- Wrong password → immediate error: `keystore password incorrect`
- Correct password → build proceeds normally
- Validation in both `FullBuild()` and `AssetsOnlyBuild()`

## Test Structure (Clean Architecture)
- **Location**: Tests live next to code (`internal/usecase/*_test.go`)
- **Removed**: Empty `tests/` directory (was leftover from old fixtures)
- **Added**: `project_build_test.go` with tests for `apkFilename()`, `validateKeystorePassword()`, `failedResult()`
- **Added (2026-06-23)**: `apk_finder_test.go` (7 tests), `zip_test.go` (4 tests), `errors_test.go` (4 test groups), `hmr_server_test.go` (8 tests), `update_install_test.go` (3 tests)
- **Coverage**: 11.2% → 19.0% (usecase layer), `pkg/errors`: 0% → 100%
- **Pattern**: Table-driven tests following Go conventions

## HMR WebSocket Server (2026-06-23)
- **File**: `internal/usecase/hmr_server.go`
- **Library**: `gorilla/websocket v1.5.3`
- **Port**: `:8765`
- **Integration**: `DevModeUseCase` creates and starts HMR server on `Start()`
- **Message types**: `reload`, `css-update`, `js-update` (HTML changes send `reload`)
- **Flow**: File change → `onFileChanged()` → `hmr.OnFileChanged()` → broadcast to connected clients
- **Fallback**: When no HMR clients connected, falls back to full assets rebuild

## Update Auto-Install (2026-06-23)
- **Methods**: `git pull --rebase --autostash` (if `.git` in TAMK_HOME) or `go install @latest`
- **Detection**: `detectUpdateMethod()` checks for `.git` dir or `go` binary
- **CLI behavior**: CRITICAL/PATCH → auto-install, MAJOR/MINOR → prompt user
- **Shell fix**: `doRun()` now calls `RunUseCase.Execute()` instead of printing "(stub)"
