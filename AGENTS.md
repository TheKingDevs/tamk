# T.A.M.K — Agent Instructions

**T.A.M.K (Termux APK Manager Kit)** v2026.3.0-HMR — Professional automation framework for native Android app development directly in Termux. Built with Go and Kotlin.

## DYNAMIC INDEX

1. [Overview & Philosophy](#overview--philosophy)
2. [Architecture & Data Flow](#architecture--data-flow)
3. [Project Structure](#project-structure)
4. [CLI Reference](#cli-reference)
5. [Project Types & Creation](#project-types--creation)
6. [Template System](#template-system)
7. [Build Pipeline](#build-pipeline)
8. [Dev Mode & HMR](#dev-mode--hmr)
9. [Clean Architecture Layers](#clean-architecture-layers)
10. [Update System](#update-system)
11. [Banner & UI Utilities](#banner--ui-utilities)
12. [Utility Modules](#utility-modules)
13. [Documentation Reference](#documentation-reference)
14. [Security & Stability](#security--stability)
15. [Agent Best Practices](#agent-best-practices)
16. [Critical Notes](#critical-notes)

---

## OVERVIEW & PHILOSOPHY

The project abandons rigid code generation in favor of a **template-oriented architecture** with **factory pattern** for project types, implemented in **Go with Clean Architecture v4** enforcing `internal/` package boundaries.

- **Clean Architecture:** Dependency Rule enforced — domain → usecase → repository → delivery. Inner layers never import outer layers.
- **Modularity:** Each project type (UI, Console, WebApp) is isolated in its own domain entity with use case orchestration.
- **Template Engine:** Code generation via `.tmpl` files with placeholder injection (`{{NAME}}`, `{{PACKAGE}}`, etc.) in `internal/repository/filesystem/template_repository.go`.
- **Tech Stack:** Go 1.26+ (core automation), Kotlin (Android templates), OpenJDK 21, Android SDK tools.
- **HMR System:** Hot Module Replacement for real-time WebApp development with inline JS bridge + file watcher.
- **Update System:** Automatic update checker via GitHub API with priority levels.
- **Build Pipeline:** `go build ./cmd/tamk` → `bin/tamk` binary.
- **Philosophy:** "Code for humans: simple, readable, and decoupled."
- **Target:** Developers seeking hardware independence for Android development.

---

## ARCHITECTURE & DATA FLOW

### Dependency Rule (Clean Architecture v4)

```
+--------------------------------------------------+
|                   delivery/cli/                   |  Cobra commands, CLI handlers
+--------------------------------------------------+
|                   usecase/                        |  Business logic, orchestration
+--------------------------------------------------+
|              repository/filesystem/               |  Data access, file I/O
|                  repository/                      |  Remote (GitHub API)
+--------------------------------------------------+
|     domain/entity/    domain/valueobject/         |  Core types, validation
|     domain/repository/                           |  Interface contracts
+--------------------------------------------------+
|              pkg/ (shared utilities)              |  Logger, errors, watcher, qrcode
+--------------------------------------------------+
|          internal/config/                         |  Central configuration
+--------------------------------------------------+
```

### Entry Point Sequence

```
User Input → cmd/tamk/main.go → cobra.Command → delivery/cli/ → usecase → repository → File System
```

### Component Map

```
cmd/tamk/main.go
└── internal/delivery/cli/root.go       # Cobra root command, dependency wiring
    ├── create                          # CreateProjectUseCase
    ├── build                           # BuildProjectUseCase
    ├── dev                             # DevModeUseCase
    ├── setup                           # SetupEnvironmentUseCase
    ├── run                             # Run (stub)
    ├── install                         # InstallUseCase
    ├── update                          # UpdateUseCase
    ├── version                         # Config display
    └── shell                           # Interactive REPL
├── internal/usecase/
│   ├── project_create.go           # Project creation wizard logic
│   ├── project_build.go            # Full APK build pipeline + incremental
│   ├── development.go              # HMR dev mode orchestrator
│   ├── environment_setup.go        # SDK download + keystore generation
│   ├── install.go                  # APK install with HTTP server + QR code
│   ├── update.go                   # Update check & apply
│   └── zip.go                      # Archive helper (zipDir)
├── internal/domain/
│   ├── entity/                        # Project, Build, Template, Keystore, Update
│   ├── valueobject/                   # ProjectName, Version, PackageName
│   └── repository/                    # Interface contracts (ports)
├── internal/config/
│   └── config.go                     # Version, SDK, env, paths
├── internal/repository/
│   ├── filesystem/                    # ProjectRepo, TemplateRepo, BuildRepo
│   └── update_repository.go           # Remote GitHub API
├── configs/
│   └── config.yaml                   # User-facing config sample
└── pkg/
    ├── logger/logger.go               # Structured slog-based logger with ANSI colors
    ├── errors/errors.go               # Domain error types
    ├── watcher/watcher.go             # fsnotify-based file watcher
    └── qrcode/qrcode.go               # QR code terminal renderer
```

### Data Flow Diagram

```mermaid
graph TD
    CLI[cmd/tamk/main.go cobra] -->|create| CUC[CreateProjectUseCase]
    CLI -->|build| BUC[BuildProjectUseCase]
    CLI -->|dev| DUC[DevModeUseCase]
    CLI -->|setup| SUC[SetupEnvironmentUseCase]
    CLI -->|run| RC[Run (stub)]
    CLI -->|install| IC[InstallUseCase]
    CLI -->|update| UUC[UpdateUseCase]
    CLI -->|version| CFG[Config]
    CLI -->|shell| SH[Shell REPL]

    CUC --> PR[ProjectRepository]
    CUC --> TR[TemplateRepository]
    BUC --> BR[BuildRepository]
    BUC --> PR
    DUC --> BUC
    DUC --> PR
    UUC --> UR[UpdateRepository]

    PR & TR & BR --> FS[(File System)]
    UR -->|GitHub API| REL([Releases])

    DUC --> FW[File Watcher]
    DUC --> ADB[ADB Bridge]
```

---

## PROJECT STRUCTURE

```
tamk/
├── cmd/tamk/main.go                 # Go entry point
├── internal/                        # Clean Architecture layers
│   ├── delivery/
│   │   ├── cli/root.go              # Cobra root + subcommands
│   │   ├── cli/wizard.go            # Interactive wizard
│   │   ├── cli/shell.go             # Interactive REPL
│   ├── usecase/                     # Business logic
│   │   ├── project_create.go        # Project creation orchestration
│   │   ├── project_build.go         # Full + incremental APK build
│   │   ├── development.go           # HMR dev mode
│   │   ├── environment_setup.go     # SDK setup
│   │   ├── install.go               # HTTP server + QR install
│   │   ├── update.go                # Update logic
│   │   └── zip.go                   # Archive helper (zipDir)
│   ├── domain/
│   │   ├── entity/                  # Core domain types
│   │   │   ├── project.go           # Project, ProjectType, WebContentMode
│   │   │   ├── build.go             # BuildResult, build state
│   │   │   ├── template.go          # Template definition
│   │   │   ├── keystore.go          # Keystore metadata
│   │   │   └── update.go            # VersionInfo, UpdateLevel
│   │   ├── valueobject/             # Value objects with validation
│   │   │   ├── project_name.go      # Sanitized project name
│   │   │   ├── version.go           # SEMVER validation
│   │   │   ├── package_name.go      # Android package format
│   │   │   └── errors.go            # Value object validation errors
│   │   └── repository/              # Interface contracts
│   │       ├── project_repository.go
│   │       ├── build_repository.go
│   │       ├── template_repository.go
│   │       └── update_repository.go
│   ├── repository/
│   │   ├── filesystem/
│   │   │   ├── project_repository.go # Project CRUD on disk
│   │   │   ├── build_repository.go   # Build pipeline (aapt2, kotlinc, d8, etc.)
│   │   │   └── template_repository.go # Template loading + rendering
│   │   └── update_repository.go      # GitHub API client
│   └── config/                       # Internal config (config.go, env detection)
├── configs/
│   └── config.yaml                  # User-facing config sample
├── pkg/                             # Shared utilities
│   ├── logger/logger.go             # Structured slog logger with ANSI colors
│   ├── errors/errors.go             # Domain error types
│   ├── watcher/watcher.go           # fsnotify-based FileWatcher
│   └── qrcode/qrcode.go             # QR code terminal renderer
├── assets/
│   ├── templates/                   # .tmpl files per project type
│   │   ├── webapp/                  # 11 templates + css/ + js/
│   │   ├── console/                 # 1 template
│   │   └── ui_apk/                  # 6 templates
│   └── images/logo.png
├── documentation/                   # 23 markdown docs
├── bin/                             # Compiled binary output
│   └── tamk                         # Go binary
├── tests/                           # Test fixture data (DupTest, TestConsole, TestWebApp)
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksum
├── .githooks/                       # Git hooks (pre-commit: fmt, vet, tests, bench)
├── .agents/
│   └── hooks/                       # Agent lifecycle hooks (work-finished)
├── Makefile                         # Build, test, lint targets
├── Dockerfile                       # Containerized build environment
└── .golangci.yml                    # Linter configuration
```

---

## CLI REFERENCE

### All Commands

| Command | Handler Layer | Description |
| :--- | :--- | :--- |
| `tamk version` | `newVersionCmd` | Show version + environment info |
| `tamk create` | `newCreateCmd` | Interactive project creation wizard |
| `tamk build -p SENHA` | `newBuildCmd` | Full APK build (compile + sign + align) |
| `tamk dev` | `newDevCmd` | HMR development mode with live reload |
| `tamk run [ARQUIVO]` | `newRunCmd` | Execute Kotlin/Java snippet (stub) |
| `tamk install [port]` | `newInstallCmd` | Serve APK download via HTTP + QR code |
| `tamk setup` | `newSetupCmd` | Download SDK + generate debug keystore |
| `tamk update` | `newUpdateCmd` | Check and install updates |
| `tamk shell` | `newShellCmd` | Interactive development REPL |

### Global Flags

| Flag | Description |
| :--- | :--- |
| `-p, --password SENHA` | Keystore password (prompted if missing) |
| `-V, --verbose` | Debug-level logging |

---

## PROJECT TYPES & CREATION

### Supported Types (from `entity.ProjectType`)

| Key | Entity Constant | Description |
| :--- | :--- | :--- |
| `webapp` | `ProjectTypeWebApp` | WebView wrapping HTML/CSS/JS + remote URL support |
| `ui_apk` | `ProjectTypeUIAPK` | Native Android UI with XML layouts + Kotlin |
| `console` | `ProjectTypeConsole` | CLI Kotlin application |

### WebApp Content Modes

During `tamk create`, when selecting **WebApp**:

| Mode | Description | URL |
| :--- | :--- | :--- |
| **Internal** (1) | Static files in `src/main/assets/` | `file:///android_asset/index.html` |
| **External** (2) | Remote URL loaded in WebView | Any `http://` or `https://` URL |

### Generated Project Structure (WebApp example)

```
MyWebApp/
├── AndroidManifest.xml
├── tamk.config
├── .gitignore
├── secret/
│   └── project.keystore       # RSA 2048 signing key
├── res/
│   ├── values/strings.xml
│   ├── values/styles.xml
│   ├── drawable/ic_launcher.xml
│   └── xml/network_security_config.xml
├── src/main/
│   ├── assets/                 # Web content goes here
│   │   ├── index.html
│   │   ├── css/styles.css
│   │   ├── js/app.js
│   │   └── .gitignore
│   └── kotlin/com/author/name/
│       └── MainActivity.kt
├── development/
│   ├── sdk/android.jar
│   └── secret/debug.keystore
└── assets/cache/               # Build temp files
```

---

## TEMPLATE SYSTEM

### Location: `assets/templates/{type}/`

### Template Files

| Type | Templates | Placeholders |
| :--- | :--- | :--- |
| **webapp** | `AndroidManifest.xml`, `MainActivity.kt`, `index.html`, `strings.xml`, `styles.xml`, `icon.xml`, `network_security_config.xml`, `dev_bridge.js`, `css/styles.css`, `js/app.js`, `gitignore_root`, `gitignore_assets` | `{{NAME}}`, `{{PACKAGE}}`, `{{VERSION}}`, `{{AUTHOR}}`, `{{WEB_URL}}`, `{{DEV_MODE}}`, `{{MIN_SDK}}`, `{{TARGET_SDK}}`, `{{TAMK_VERSION}}`, `{{DEV_PORT}}` |
| **ui_apk** | `AndroidManifest.xml`, `MainActivity.kt`, `activity_main.xml`, `strings.xml`, `styles.xml`, `icon.xml` | `{{NAME}}`, `{{PACKAGE}}`, `{{VERSION}}`, `{{AUTHOR}}` |
| **console** | `Main.kt` | `{{NAME}}`, `{{VERSION}}`, `{{AUTHOR}}` |

### Template Processing Flow (in `internal/repository/filesystem/template_repository.go`)

```
1. Read .tmpl file from tmpl_dir via TemplateRepository
2. Apply replacements via strings.Replacer ({{PLACEHOLDER}})
3. Write to final_path in project directory
4. Skip missing templates with WARNING (non-fatal)
```

### MainActivity.kt.tmpl Key Features (WebApp)

- `BroadcastReceiver` for `tamk.ACTION_REFRESH_ASSET` and `tamk.ACTION_REFRESH_ALL`
- JavaScript bridge injection via `injectHMRBridge()` after page load
- `WebViewClient` + `WebChromeClient` with logging
- `onBackPressed()` for WebView history navigation
- Error handling with `Toast` notifications
- `onResume()` / `onPause()` lifecycle for broadcast receiver

### WebView Configuration

```kotlin
webView.settings.apply {
    javaScriptEnabled = true
    domStorageEnabled = true
    allowFileAccess = true
    allowContentAccess = false
    allowFileAccessFromFileURLs = true
    allowUniversalAccessFromFileURLs = true
    cacheMode = WebSettings.LOAD_DEFAULT
    setSupportZoom(true)
    builtInZoomControls = true
    displayZoomControls = false
}
```

---

## BUILD PIPELINE

### Full Build (`BuildProjectUseCase.FullBuild()`)

1. **Validation**: Check SDK exists, validate keystore password (min 6 chars)
2. **Hash Check**: SHA-256 of all source files via `BuildRepository.CalculateHash()`
3. **Cache Hit**: If `.build_cache` matches current hash AND APK exists on disk → skip build ("Nada mudou")
4. **AAPT2 Compile**: `aapt2 compile --dir res -o res.zip`
5. **AAPT2 Link**: `aapt2 link -I sdk/android.jar --manifest AndroidManifest.xml -o app.apk [-A src/main/assets] res.zip --auto-add-overlay`
6. **Kotlin Compile**: `kotlinc src/main/kotlin gen/ -cp sdk/android.jar -d obj/`
7. **D8 DEX**: `d8 --lib sdk/android.jar --release --output . *.class`
8. **Package DEX**: `zip -j app.apk classes.dex`
9. **Zipalign**: `zipalign -f 4 app.apk app-unsigned.apk`
10. **Sign**: `apksigner sign --ks keystore --ks-pass pass:xxx --out app-final.apk app-unsigned.apk`
11. **Cleanup**: Remove temp files, save `.build_cache`

### Incremental Assets Build (`BuildProjectUseCase.AssetsOnlyBuild()`)

Used by HMR dev mode for fast iteration:

1. Extract existing APK (app-final.apk or app-dev.apk)
2. Remove old assets/ folder
3. Copy current `src/main/assets/` into extracted APK
4. Repackage and sign → `app-dev.apk`
5. Auto-install via ADB if device connected

### ADB Push (`DevModeUseCase.PushAssetToDevice()` - stub)

For HTML hot-reload without full rebuild:

1. Verifies file is inside `src/main/assets/`
2. Creates directory on device via `adb shell mkdir -p`
3. Pushes file via `adb push`
4. Broadcasts refresh intent via `adb shell am broadcast`

### Security Measures

- `SanitizePath()` blocks path traversal (`[;&|`$]`, `..`)
- Commands use `exec.Command()` (list-based, no shell=True equivalent)
- Passwords redacted in logs (`--ks-pass [REDACTED]`)
- `ValidateKeystorePassword()` pre-validates before build
- Password cleared from memory after use (`clear` slice)
- 10-minute timeout on full build, 5-minute on assets build

---

## DEV MODE & HMR

### Architecture (`DevModeUseCase` in `internal/usecase/development.go`)

```
DevModeUseCase
├── FileWatcher (fsnotify)        — File system monitoring
│   ├── onFileChanged             — Debounced callback (500ms)
│   └── Watch extensions: .html, .css, .js, .json, .png, .jpg, .svg, .webp, .xml, .kt
├── ADB Bridge                    — Auto-install + push + broadcast
└── HMR Bridge (inline JS)       — Injected into index.html as WebSocket client
    ├── Connects to ws://localhost:8765
    ├── Handles reload, css-update messages
    └── Sends hello on connect
```

### Client-Side Bridge (injected inline by `injectDevBridge()`)

A WebSocket client is injected into `index.html` before `</body>`:

```javascript
(function() {
    var ws = new WebSocket('ws://localhost:8765');
    ws.onmessage = function(e) {
        var msg = JSON.parse(e.data);
        if (msg.type === 'reload') { location.reload(); }
        else if (msg.type === 'css-update') {
            document.querySelectorAll('link[rel="stylesheet"]').forEach(function(link) {
                link.href = link.href.split('?')[0] + '?t=' + Date.now();
            });
        }
    };
    ws.onopen = function() {
        ws.send(JSON.stringify({type: 'hello', url: window.location.href}));
    };
})();
```

### Dev Mode Lifecycle

```
Start (tamk dev):
1. Validate project is WebApp (check tamk.config via ProjectRepository)
2. Check prerequisites (assets dir exists, APK base exists)
3. Inject dev bridge (inline JS into index.html, backup saved)
4. Start file watcher on src/main/assets/
5. Show instructions banner

Change detected:
1. .css/.js files → log (HMR-ready, server-side WS pending)
2. Other files → quick assets build via AssetsOnlyBuild()

Stop (Ctrl+C):
1. Stop file watcher
2. Restore index.html from .tamk_backup
```

---

## CLEAN ARCHITECTURE LAYERS

### Dependency Rule

```
domain/entity → domain/valueobject → domain/repository (interfaces)
    ↑
usecase/  (imports domain, repository interfaces)
    ↑
repository/filesystem/  (implements domain repository interfaces)
    ↑
delivery/cli/  (wires dependencies, handles CLI input)
```

### Domain Layer (`internal/domain/`)

| Package | Types | Purpose |
| :--- | :--- | :--- |
| `entity/` | `Project`, `ProjectType`, `BuildResult`, `Template`, `Keystore`, `VersionInfo`, `UpdateLevel` | Core domain types with no external dependencies |
| `valueobject/` | `ProjectName`, `Version`, `PackageName`, validation errors | Immutable value objects with built-in validation |
| `repository/` | `ProjectRepository`, `BuildRepository`, `TemplateRepository`, `UpdateRepository` | Interface contracts (Go interfaces) |

### Use Case Layer (`internal/usecase/`)

| Use Case | Key Methods | Description |
| :--- | :--- | :--- |
| `CreateProjectUseCase` | `Execute(input)` | Project creation wizard, template processing |
| `BuildProjectUseCase` | `FullBuild(input)`, `AssetsOnlyBuild(input)` | APK build pipeline + incremental |
| `DevModeUseCase` | `Start(ctx, projectPath, password)`, `Stop(ctx)`, `GetStatus()` | HMR dev mode orchestrator |
| `SetupEnvironmentUseCase` | `Execute()` | SDK download + keystore generation |
| `InstallUseCase` | `Serve(ctx, projectPath, port)`, `findAPK()` | HTTP server + QR code for APK download |
| `UpdateUseCase` | `Check(ctx)`, `ShouldAutoInstall(info)`, `PrintUpdateInfo(info)` | GitHub release check |

### Repository Layer (`internal/repository/`)

| Repository | File | Implements |
| :--- | :--- | :--- |
| `filesystem.ProjectRepository` | `project_repository.go` | `domain/repository.ProjectRepository` |
| `filesystem.BuildRepository` | `build_repository.go` | `domain/repository.BuildRepository` |
| `filesystem.TemplateRepository` | `template_repository.go` | `domain/repository.TemplateRepository` |
| `update_repository.go` | `update_repository.go` | `domain/repository.UpdateRepository` |

### Delivery Layer (`internal/delivery/`)

| Package | File | Purpose |
| :--- | :--- | :--- |
| `cli/` | `root.go` | Cobra root command + subcommand wiring |

---

## UPDATE SYSTEM

### Components

| Package | Type | Responsibility |
| :--- | :--- | :--- |
| `internal/repository/update_repository.go` | `UpdateRepository` | HTTP client for GitHub releases API + 6h cache |
| `internal/usecase/update.go` | `UpdateUseCase` | Orchestrates check, print, auto-install logic |

### Update Priority Levels

| Level | Detection | Behavior |
| :--- | :--- | :--- |
| **CRITICAL** | Keywords: `critical`, `security`, `urgent` | Auto-install immediately |
| **PATCH** | Keywords: `patch`, `bugfix`, `fix` | Auto-install silently |
| **MAJOR** | major version bump | Prompt user |
| **MINOR** | minor version bump | Prompt user |
| **OPTIONAL** | Cosmetic/docs only | Notify only |

### Update Methods (auto-detected)

1. **git**: `git pull --rebase --autostash` in TAMK_HOME
2. **go install**: `go install github.com/Shadw-Developer/tamk/cmd/tamk@latest`

---

## BANNER & UI UTILITIES

### Banner (`internal/delivery/cli/`)

Banners are built directly in the CLI layer using ANSI terminal utilities. Shared helpers live in the CLI package:

- Dynamic width detection (`tput cols` equivalent via `golang.org/x/term`)
- Text centering, box drawing with Unicode chars
- Color styles via ANSI constants
- SIGWINCH handling for terminal resize

### Color Utilities (`pkg/logger/logger.go`)

| Function/Type | Description |
| :--- | :--- |
| `slog.Level` extensions | Custom levels (Debug, Info, Step, Success, Warning, Error) |
| `logger.Init(verbose bool)` | Initialize structured logger |
| `logger.Debug(msg, args...)` | Debug-level logging |
| `logger.Info(msg, args...)` | Info-level logging |
| `logger.Success(msg, args...)` | Success-level logging |
| `logger.Warn(msg, args...)` | Warning-level logging |
| `logger.Error(msg, args...)` | Error-level logging |
| ANSI constant helpers | Terminal color codes in logger output |

### Logo Output

ASCII art logo loaded and rendered during `tamk create` and `tamk version`.

---

## UTILITY MODULES

### File Watcher (`pkg/watcher/watcher.go`)

| Component | Function |
| :--- | :--- |
| `Watcher` | fsnotify-based file watcher with single-timer debounced callback |
| `New(onChange func(string), debounce time.Duration)` | Create new watcher with debounce |
| `Start(path string)` | Begin watching directory recursively |
| `Stop()` | Stop watcher gracefully |
| `IsRunning()` | Check if watcher is active |
| Watch extensions | `.html`, `.css`, `.js`, `.json`, `.png`, `.jpg`, `.jpeg`, `.svg`, `.webp` |
| Ignore patterns | `.git`, `node_modules`, `assets/cache`, `secret` |

### Config (`internal/config/config.go`)

| Feature | Details |
| :--- | :--- |
| `Version` | `"2026.3.0-HMR"` |
| Environment detection | Termux (`/com.termux`), SmartIDE (`/org.smartide.code`), unknown |
| Secure paths | `SecurePath()` with `filepath.Abs()` + `filepath.Clean()` |
| Path validation | Ensures all paths are under home or cwd |
| Template resolution | `GetTemplateDir(type)` — searches TAMK_HOME, source dir, cwd, system paths |

### QR Code (`pkg/qrcode/qrcode.go`)

| Component | Description |
| :--- | :--- |
| `Print(url string)` | Render QR code in terminal using Unicode block chars |
| `PrintWithFrame(url string)` | Render QR code with ASCII frame border |

### Errors (`pkg/errors/errors.go`)

| Sentinel | Description |
| :--- | :--- |
| `ErrProjectNotFound` | No project in current directory |
| `ErrBuildFailed` | Generic build failure |
| `ErrKeystoreInvalidPass` | Wrong keystore password |
| `ErrNotAWebAppProject` | Dev mode only supports WebApp |
| `ErrWebSocketNotAvailable` | WebSocket server unavailable |
| `ErrWatchdogNotAvailable` | fsnotify unavailable |
| `BuildError` | Structured build error with Phase field |

---

## DOCUMENTATION REFERENCE

| File | Purpose |
| :--- | :--- |
| `ARCHITECTURE.md` | System architecture and data flow diagrams |
| `API_COMPONENTS.md` | Class and module reference |
| `DEV_GUIDE.md` | Development workflow and debugging |
| `QUICKSTART.md` | Step-by-step getting started guide |
| `STRUCTURE.md` | Complete directory tree reference |
| `WEBAPP_TEMPLATES.md` | WebApp-specific template docs |
| `CONSOLE_TEMPLATES.md` | Console template documentation |
| `HMR_SYSTEM.md` | Hot Module Replacement system (detailed) |
| `HMR_GUIDE.md` | HMR quick reference and API |
| `HMR_EXAMPLES.md` | HMR practical patterns and recipes |
| `UPDATE_SYSTEM.md` | Auto-update system with priority levels |
| `BANNER_UTILS.md` | Banner API reference |
| `INSTALLATION.md` | Multi-platform installation guide |
| `INSTALLATION_TERMUX.md` | Termux-specific installation |
| `INSTALLATION_LINUX.md` | Linux installation |
| `INSTALLATION_WINDOWS.md` | Windows installation |
| `INSTALLATION_MACOS.md` | macOS installation |
| `ANDROID_APK_INSTALL.md` | ADB/root APK installation guide |
| `FAQ.md` | Frequently asked questions |
| `CHANGELOG.md` | Version history and changes |
| `CONTRIBUTING.md` | Contribution guidelines |
| `VERSIONING.txt` | Version scheme reference |

---

## HOOKS SYSTEM

Two independent hooks systems coexist in the project:

| Directory | Trigger | Purpose | Standard |
| :--- | :--- | :--- | :--- |
| `.githooks/` | `git commit` | Code quality gates (fmt, vet, tests, bench) | Git convention via `core.hooksPath` |
| `.agents/hooks/` | Agent task completion | User notification (vibrate/notify) | opencode agent lifecycle |

**Do not conflate them.** They serve different layers:
- `.githooks/pre-commit` runs on every `git commit` — enforces Go formatting, vet, tests, and benchmarks.
- `.agents/hooks/work-finished` runs when an AI agent finishes a task — notifies the user.

### Setup

```bash
make setup  # Includes: git config core.hooksPath .githooks
```

---

## SECURITY & STABILITY

### Security Measures

- **Path sanitization**: All user-supplied paths are sanitized (path traversal blocked, special chars filtered)
- **Password handling**: Zeroed after use (`clear` byte slice), never logged, redacted in error output
- **Subprocess safety**: `exec.Command()` list-based (no shell interpolation), shell=True equivalent never used
- **File permissions**: Keystore files set to `0o600` (owner read/write only)
- **SDK validation**: SHA-256 hash verification on downloaded SDK
- **Keystore isolation**: Per-project keystore (`secret/project.keystore`) with debug fallback
- **Input validation**: Project names, versions, packages validated via value objects before creation

### Error Handling

- All use cases return `error` with user-friendly messages
- Verbose mode (`-V`) enables debug-level logging
- 10-minute timeout on full build, 5-minute on assets build (`context.WithTimeout`)
- Graceful cleanup of cache on build failures
- `os.RemoveAll` cleanup on project creation failures

### Common Errors

| Error | Cause | Solution |
| :--- | :--- | :--- |
| `Keystore password incorrect` | Wrong password in `--build` | Re-enter correct password |
| `aapt2 not found` | Missing dependency | `pkg install aapt2` |
| `SDK not configured` | SDK path not set | `tamk setup` |
| `Build failed: D8 error` | Kotlin syntax error | Check generated Kotlin files |
| `no project found in current directory` | No tamk.config | Run from project directory |
| `APK base not found` | No prior build for dev mode | `tamk build` first |

---

## AGENT BEST PRACTICES

1. **Always read** `./AGENTS.md` first for context.
2. **Check existing docs** in `documentation/` before creating new files.
3. **Preserve template structure** when modifying `.tmpl` files.
4. **Test changes** with `go test ./...` and `make build`.
5. **Update CHANGELOG.md** for new features or fixes.
6. **Respect user settings** in `settings.local.json`.
7. **Never expose credentials** (Keystore passwords, paths).
8. **Understand HMR system** — Review `HMR_SYSTEM.md` for WebApp dev workflows.
9. **Incremental builds** — `AssetsOnlyBuild()` is key for HMR; do not break it.
10. **File Watcher** depends on `fsnotify` library — ensure import fallback.
11. **Async cleanup** — Dev server has complex goroutine lifecycle; preserve proper shutdown.
12. **Version updates** — Update `config.Version` constant in `internal/config/config.go` for new releases.
13. **Build cache** — APK existence is verified (`os.Stat`) alongside hash comparison; do not remove this check.

---

## CRITICAL NOTES

- **Templates**: All templates in `assets/templates/` — never modify without testing both creation and build.
- **Keystore**: Sensitive files in `secret/` — never log or expose passwords in error output.
- **Environment**: Assumes Termux with Go 1.26+, OpenJDK 21, Kotlin, aapt2, apksigner, zipalign.
- **SDK Path**: Configured via `internal/config/config.go` from `TAMK_HOME` or auto-detected; falls back to `tamk setup`.
- **Build Cache**: `.build_cache` file for hash comparison + APK existence check; `assets/cache/` for temp files — safe to delete.
- **Dev Bridge**: Inline JS injected/removed by dev mode; `index.html.tamk_backup` preserved.
- **Config Singleton**: `config.New()` instantiates fresh config per call — consistent paths rely on `TAMK_HOME` env var.
- **Async Lifecycle**: Dev server runs goroutines; proper shutdown via context cancellation + WaitGroup.
- **fsnotify**: Optional dependency; dev mode fails gracefully if not available.
- **Clean Architecture**: Never break the Dependency Rule — `internal/` packages must only import inward.

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Made with ❤️ by @mrx_dev</sub>
</div>
