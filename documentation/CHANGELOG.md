# 📋 T.A.M.K Changelog

All notable changes. Format based on [Keep a Changelog](https://keepachangelog.com/), adhering to [Semantic Versioning](https://semver.org/).

---

## [2026.3.0-HMR] - 2026-06-18

### ✨ Added

- **Development Mode (HMR)**: `tamk dev` with complete Hot Module Replacement
  - CSS hot reload without state loss
  - JavaScript HMR with module injection and state preservation
  - JSON real-time data update
  - WebSocket server (port 8765) + HTTP fallback (port 8080)
  - File watcher (fsnotify) with 500ms debounce
  - Incremental asset build (`AssetsOnlyBuild()`)
  - ADB auto-install + asset push + broadcast refresh
  - HMR server with module registry and client state tracking
  - Bridge JS `tamk-dev-bridge.js` automatically injected
  - Backup and restore of `index.html`
- **Automatic Update System**:
  - `Checker` with 6h cache and GitHub API
  - 5 priority levels: CRITICAL, MAJOR, MINOR, PATCH, OPTIONAL
  - Auto-updater with backup, 3 methods (git/go install/script)
- **Remote URL WebApp**: Support for loading external URLs in WebView
- **Java Support**: Run controller now executes `.java` besides `.kt`

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
- **Dev mode**: `onFileChanged()`, `AssetsOnlyBuild()`, status command, 500ms debounce
- **HMR server**: Client-side WebSocket bridge (server-side pending)
- **Build pipeline**: `FullBuild()`, `AssetsOnlyBuild()` incremental, `PushAssetToDevice()` via ADB
- **Template system**: `TemplateRepository` loading + placeholder rendering
- **Config**: Environment detection in `internal/config/config.go`, `GetTemplateDir()` multi-path fallback
- **Dependency injection**: All use cases wired in `root.go` with constructor injection

### 📚 Documentation

- AGENTS.md completely rewritten (Go + Clean Architecture, 16 sections, Mermaid diagrams)
- ARCHITECTURE.md updated with Clean Architecture layer diagrams
- API_COMPONENTS.md updated with Go package reference
- QUICKSTART.md updated for Go binary usage
- STRUCTURE.md updated with Go project tree
- README.md restructured with technology table and architecture sections
- HMR_SYSTEM.md, HMR_GUIDE.md, HMR_EXAMPLES.md (new)
- UPDATE_SYSTEM.md (new)
- BANNER_UTILS.md (new)
- CLAUDE.md, GEMINI.md updated
- MIGRATION_PLAN.md (internal migration tracking)

---

## [2.3.2] - 2026-03-29

### 🔧 Fixed

- Installation path: path resolution corrected, environment variables updated

---

## [2.3.1] - 2026-03-29

### 🎨 Added

- Visual identity: project logo integration

### 🔧 Fixed

- Banner system: rendering and update, toilet adjusted

---

## [2.3.0] - 2026-03-29

### ✨ Added

- HMR - Hot Module Replacement (initial version)
  - CSS Hot Reload, JavaScript HMR, JSON HMR
  - State Preservation
  - DevServer with module registry

### 🔧 Changed

- DevController: `_on_file_changed()` with per-type HMR
- `_quick_assets_build()` with reload notification
- `dev_bridge.js.tmpl` updated for 2026.3.0-HMR

---

## [2.2.0] - 2026-02-27

### ✨ Added

- Hosted WebApps: remote URL support in WebView

---

## [2.1.2] - 2026-01-21

### 📌 Updated

- Contributor credits in README

---

## [2.1.1] - 2026-01-21

### 🔧 Fixed

- Console projects: structure and `Main.kt` template

---

## [2.1.0] - 2026-01-20

### ✨ Added

- Modernized CLI with improved UX
- Input system with `tput`, layout with `toilet`

---

## [2.0.1] - 2026-01-20

### 📝 Fixed

- README: title formatting

---

## [2.0.0] - 2026-01-20

### ✨ Added

- **WebApp Support**: WebView + HTML/CSS/JS in APK
- WebAppStructure, webapp templates
- Complete documentation (ARCHITECTURE, API_COMPONENTS, DEV_GUIDE, etc.)

### 🔧 Changed

- BuildController: `-A` flag for assets
- ProjectFactory: webapp mapping
- README with WebApp flow

---

## [1.0.0] - 2026-01-20

### ✨ Added

- Initial release
- UI APK (native XML) and Console (Kotlin CLI)
- Build pipeline (aapt2, kotlinc, d8, apksigner, zipalign)
- `.tmpl` template system
- `setup-install.sh` installer
- Per-project private keystore

---

## Versioning Format

`MAJOR.MINOR.PATCH` (with optional suffix):
- **MAJOR**: Incompatible changes
- **MINOR**: New compatible features
- **PATCH**: Bug fixes
- **-HMR**: Release with Hot Module Replacement support

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Changelog</sub>
</div>
