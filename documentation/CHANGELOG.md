# 📋 T.A.M.K Changelog

All notable changes. Format based on [Keep a Changelog](https://keepachangelog.com/), adhering to [Semantic Versioning](https://semver.org/).

---

## [1.0.0] - 2026-06-23

### ✨ Added

- **Clean Architecture v4**: Full migration from Python to Go
  - Domain entities + value objects + repository interfaces
  - Use cases: CreateProject, BuildProject, DevMode, Setup, Install, Update, Run
  - CLI with Cobra + interactive wizard + REPL shell
  - Dependency injection via constructor wiring in `root.go`
- **Project Creation**: `tamk create` wizard
  - WebApp (WebView + HTML/CSS/JS) — internal (assets) or external (URL)
  - UI APK Nativo (XML + Kotlin)
  - Console Kotlin CLI — boilerplate with Calculator and StringUtils
  - Validation: name, author, SEMVER version, Android package name
  - Per-project keystore RSA 2048 (`secret/project.keystore`)
  - 18 `.tmpl` templates organized in `xml/`, `kotlin/`, `css/`, `js/`
- **APK Build Pipeline**: `tamk build -p SENHA`
  - Full pipeline via ToolManager: AAPT2 → kotlinc → D8 → zipalign → apksigner
  - Incremental assets build (`AssetsOnlyBuild()`) for dev mode
  - SHA-256 build cache (skip if nothing changed)
  - Keystore password validation before build (`keytool -list`)
  - Timeout: 10min full build, 5min incremental
- **Development Mode (HMR)**: `tamk dev`
  - File watcher (fsnotify) with 500ms debounce
  - Bridge JS inline injected into `index.html`
  - **WebSocket server on port 8765** for real-time hot reload
  - Messages: `reload`, `css-update`, `js-update`
  - Fallback to full assets rebuild when no HMR clients connected
  - ADB push for HTML/CSS/JS + broadcast `ACTION_REFRESH`
  - Backup and restore of `index.html`
- **Code Execution**: `tamk run [file]`
  - Compiles and executes Kotlin/Java files natively
  - Auto-detection of `Main.kt` in `src/`
  - Automatic main class resolution
  - Uses ToolManager for tool paths
- **Environment Setup**: `tamk setup`
  - Tool extraction to `~/.tamk/tools/{os}/` ("Setup Once, Run Forever")
  - SDK download (android.jar platform-30)
  - Debug keystore generation
  - Version marker (`.version`) for re-extraction only on TAMK update
- **APK Installation**: `tamk install [port]`
  - Built-in HTTP server + QR code in terminal
  - ADB side-load support
- **Automatic Update System**: `tamk update`
  - GitHub API release checker (6h cache)
  - 5 priority levels: CRITICAL, MAJOR, MINOR, PATCH, OPTIONAL
  - **Auto-install via `git pull --rebase --autostash`** (if `.git` in TAMK_HOME)
  - **Auto-install via `go install @latest`** (fallback)
  - CRITICAL/PATCH: auto-install silently
  - MAJOR/MINOR: prompt user before install
- **Interactive Shell**: `tamk shell`
  - Commands: `create`, `build`, `dev`, `setup`, `install`, `update`, `run`, `version`
  - `cd`, `pwd`, `history`, `help`
- **Guardian Security System**: `tamk build --guardian`
  - Anti-Debug: 5 detection methods (API, TracerPid, Timing, Props, JDWP)
  - Anti-Root: 4 layers (Files, /data/data, Apps, Magisk)
  - Anti-Frida: 5 methods (Process, Maps, Ports, TCP, Files)
  - Anti-Emulator: Build props, files, QEMU, CPU info
  - Asset Encryption: AES-256-GCM with magic header `TAMK_ENC_1`
  - RASP: Periodic runtime checks with threat callbacks
  - Security levels: none, basic, standard, maximum
- **ToolManager**: Multi-platform tool management
  - Linux/macOS/Android: system tools (lightweight)
  - Windows: embedded tools via `go:embed` (zero config)
- **8 GitHub Actions workflows**: CI, build, release, tags, security, maintenance, Windows test, T2T audit
- **Platform-specific installers**: Termux, Linux (5 distros), macOS, Windows
- **13 test files**: usecase, entity, valueobject, errors packages

### 🔧 Changed

- **Complete migration from Python to Go with Clean Architecture v4**
  - Language: Python 3.x → Go 1.26
  - CLI framework: argparse → Cobra
  - Architecture: Controllers + Factory → Clean Architecture (domain → usecase → repository → delivery)
  - Logger: Custom Python logger → structured slog-based logger with ANSI colors
  - File watcher: watchdog (Python) → fsnotify (Go)
  - Build: `python src/main.py` → `go build ./cmd/tamk`
  - Binary: PyInstaller bundle → native Go binary at `bin/tamk`
  - Template engine: string replacements → `internal/repository/filesystem/template_repository.go`
  - Update checker: standalone Python → `internal/usecase/update.go` + `internal/repository/update_repository.go`
- Shell `run` command: Now calls `RunUseCase.Execute()` (was stub)
- Stack: Added `gorilla/websocket` dependency

### 📚 Documentation

- AGENTS.md: Go + Clean Architecture, 16 sections, Mermaid diagrams
- ARCHITECTURE.md: Clean Architecture layer diagrams
- API_COMPONENTS.md: Go package reference
- QUICKSTART.md, STRUCTURE.md, README.md updated
- HMR_SYSTEM.md, HMR_GUIDE.md, HMR_EXAMPLES.md
- UPDATE_SYSTEM.md, BANNER_UTILS.md
- GUARDIAN_SECURITY.md: Complete security documentation
- CONTRIBUTING.md, FAQ.md, CHANGELOG.md
- 21 docs + VERSIONING.txt total

---

## Versioning Format

`MAJOR.MINOR.PATCH` (with optional suffix):
- **MAJOR**: Incompatible changes
- **MINOR**: New compatible features
- **PATCH**: Bug fixes
- **-HMR**: Release with Hot Module Replacement support

---

<div align="center">
  <sub>T.A.M.K v1.0.0 — Changelog</sub>
</div>
