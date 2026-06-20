# 📂 Complete T.A.M.K Project Structure

> Full mapping of all files and directories of T.A.M.K v2026.3.0-HMR, now built with Go and Clean Architecture v4.

---

## Complete Tree

```
tamk/                                    # Project root
│
├── AGENTS.md                           # Agent instructions (this file)
├── CLAUDE.md                           # Claude AI instructions
├── GEMINI.md                           # Gemini AI instructions
├── README.md                           # Main documentation
├── LICENSE                             # MIT License
├── go.mod                              # Go module definition
├── go.sum                              # Go module checksum
├── Makefile                            # Build, test, lint targets
├── Dockerfile                          # Containerized build environment
├── .golangci.yml                       # Linter configuration
├── .githooks/                          # Git hooks (pre-commit via core.hooksPath)
└── .agents/
    └── hooks/                          # Agent lifecycle hooks (work-finished)
│
├── cmd/                                # Entry points
│   └── tamk/
│       └── main.go                     # Go entry point -> cobra root command
│
├── internal/                           # Clean Architecture layers
│   ├── config/                         # Config (config.go, env detection)
│   ├── delivery/
│   │   ├── cli/
│   │   │   ├── root.go                # Cobra root + subcommand wiring
│   │   │   ├── wizard.go              # Interactive project creation wizard
│   │   │   └── shell.go               # Interactive development REPL
│   ├── domain/
│   │   ├── entity/                     # Core domain types
│   │   │   ├── project.go             # Project, ProjectType, WebContentMode
│   │   │   ├── build.go               # BuildResult, build state
│   │   │   ├── template.go            # Template definition
│   │   │   ├── keystore.go            # Keystore metadata
│   │   │   └── update.go              # VersionInfo, UpdateLevel
│   │   ├── valueobject/               # Immutable value objects
│   │   │   ├── project_name.go        # Sanitized project name
│   │   │   ├── version.go             # SEMVER validation
│   │   │   ├── package_name.go        # Android package format
│   │   │   └── errors.go              # Value object validation errors
│   │   └── repository/                # Interface contracts (ports)
│   │       ├── project_repository.go
│   │       ├── build_repository.go
│   │       ├── template_repository.go
│   │       └── update_repository.go
│   ├── repository/
│   │   ├── filesystem/
│   │   │   ├── project_repository.go  # Project CRUD on disk
│   │   │   ├── build_repository.go    # Build pipeline (aapt2, kotlinc, d8, ...)
│   │   │   └── template_repository.go # Template loading + rendering
│   │   └── update_repository.go       # GitHub API client
│   └── usecase/                       # Business logic
│       ├── project_create.go          # Project creation orchestration
│       ├── project_build.go           # Full + incremental APK build
│       ├── development.go             # HMR dev mode orchestrator
│       ├── install.go                 # HTTP server + QR code for APK install
│       ├── environment_setup.go       # SDK download + debug keystore
│       ├── update.go                  # Update check + apply
│       └── zip.go                     # Archive helper (zipDir)
│
├── pkg/                               # Shared utilities
│   ├── logger/
│   │   └── logger.go                  # Structured slog logger with ANSI colors
│   ├── errors/
│   │   └── errors.go                  # Domain error types
│   ├── watcher/
│   │   └── watcher.go                 # fsnotify-based file watcher
│   └── qrcode/
│       └── qrcode.go                  # QR code terminal renderer
│
├── configs/
│   └── config.yaml                    # User-facing config sample
│
├── assets/                            # Static resources
│   ├── images/
│   │   └── logo.png                   # Project logo
│   └── templates/                     # Template system (.tmpl)
│       ├── webapp/                    # WebApp (12 files)
│       │   ├── AndroidManifest.xml.tmpl
│       │   ├── MainActivity.kt.tmpl
│       │   ├── dev_bridge.js.tmpl     # HMR client bridge
│       │   ├── index.html.tmpl
│       │   ├── strings.xml.tmpl
│       │   ├── styles.xml.tmpl
│       │   ├── icon.xml.tmpl
│       │   ├── network_security_config.xml.tmpl
│       │   ├── gitignore_root.tmpl
│       │   ├── gitignore_assets.tmpl
│       │   ├── css/
│       │   │   └── styles.css.tmpl
│       │   └── js/
│       │       └── app.js.tmpl
│       ├── console/                   # Console (1 file)
│       │   └── Main.kt.tmpl
│       └── ui_apk/                    # UI APK (6 files)
│           ├── AndroidManifest.xml.tmpl
│           ├── MainActivity.kt.tmpl
│           ├── activity_main.xml.tmpl
│           ├── strings.xml.tmpl
│           ├── styles.xml.tmpl
│           └── icon.xml.tmpl
│
├── bin/                               # Compiled binary output
│   └── tamk                           # Go binary
│
├── PRD/                               # Product requirements (workflow.md)
├── documentation/                     # Documentation (21 .md + 1 .txt)
│   ├── ARCHITECTURE.md                # Architecture and data flow
│   ├── API_COMPONENTS.md              # Package/module reference
│   ├── DEV_GUIDE.md                   # Development guide
│   ├── QUICKSTART.md                  # Quick start guide (10 min)
│   ├── STRUCTURE.md                   # This file
│   ├── HMR_SYSTEM.md                  # Complete HMR documentation
│   ├── HMR_GUIDE.md                   # HMR quick guide
│   ├── HMR_EXAMPLES.md                # HMR practical examples
│   ├── UPDATE_SYSTEM.md               # Update system
│   ├── BANNER_UTILS.md                # Banner API
│   ├── WEBAPP_TEMPLATES.md            # WebApp templates
│   ├── CONSOLE_TEMPLATES.md           # Console templates
│   ├── INSTALLATION.md                # Multi-platform installation
│   ├── INSTALLATION_TERMUX.md         # Termux installation
│   ├── INSTALLATION_LINUX.md          # Linux installation
│   ├── INSTALLATION_WINDOWS.md        # Windows installation
│   ├── INSTALLATION_MACOS.md          # macOS installation
│   ├── ANDROID_APK_INSTALL.md         # APK installation via ADB
│   ├── FAQ.md                         # Frequently asked questions
│   ├── CHANGELOG.md                   # Version history
│   ├── CONTRIBUTING.md                # Contribution guide
│   └── VERSIONING.txt                 # Versioning scheme
│
├── tests/                             # Test resources
├── development/                       # Development environment (SDK cache)
├── setup-install.sh                   # Termux installation script
└── tamk                               # Shell launcher wrapper
```

---

## Generated Project Structure

### WebApp (internal mode)

```
MeuWebApp/
├── AndroidManifest.xml                 # Android manifest
├── tamk.config                         # Metadata
├── .gitignore                          # Ignores APK, cache, keystore
├── secret/
│   └── project.keystore               # RSA 2048 signing key
├── res/
│   ├── values/strings.xml             # Name, version, author
│   ├── values/styles.xml              # Material theme
│   ├── drawable/ic_launcher.xml       # Icon
│   ├── drawable/ic_launcher_round.xml # Rounded icon
│   └── xml/network_security_config.xml
├── src/main/
│   ├── assets/                        # ⭐ WEB CONTENT
│   │   ├── index.html                 # Entry point
│   │   ├── css/styles.css             # Styles
│   │   ├── js/app.js                  # JavaScript
│   │   ├── js/tamk-dev-bridge.js      # (dev mode only)
│   │   └── .gitignore                 # Assets
│   └── kotlin/com/author/appname/
│       └── MainActivity.kt            # WebView + HMR bridge
├── development/
│   ├── sdk/android.jar                # Isolated SDK
│   └── secret/debug.keystore          # Debug fallback
├── assets/cache/                      # Build temp files
├── .build_cache                       # Hash cache (incremental build)
├── {name}-{version}-release.apk       # Signed APK
└── {name}-{version}-dev.apk           # Development APK
```

### WebApp (external/URL mode)

```
MeuWebApp/
├── AndroidManifest.xml                 # With INTERNET permission
├── tamk.config                         # web_url=https://example.com
├── secret/project.keystore            # Keystore
├── res/                               # Android resources (no web assets)
├── src/main/kotlin/.../MainActivity.kt # WebView loading URL
└── development/                       # Local SDK
```

### UI APK

```
MeuApp/
├── AndroidManifest.xml
├── tamk.config
├── secret/project.keystore
├── res/
│   ├── layout/activity_main.xml
│   ├── values/strings.xml
│   ├── values/styles.xml
│   └── drawable/ic_launcher*.xml
├── src/main/kotlin/.../MainActivity.kt  # Native Activity
└── development/
```

### Console

```
MeuConsole/
├── tamk.config                          # type=console
├── src/Main.kt                          # Kotlin entry point
└── libs/                                # Libraries (optional)
```

---

## Important Notes

- **`src/main/assets/`**: Most important folder for WebApps. All content is packaged in the APK and accessible via `file:///android_asset/`.
- **`.gitignore`**: Root ignores APKs, cache, keystore. `assets/.gitignore` ignores temp files.
- **Keystore**: `secret/` contains sensitive data. **Never version.**
- **Dev Mode**: `tamk-dev-bridge.js` is created/removed automatically. `index.html.tamk_backup` preserves edits.
- **Build Cache**: `assets/cache/` and `.build_cache` are safe to delete.
- **Clean Architecture**: The `internal/` package enforces the Dependency Rule — inner layers never import outer layers.

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Project Structure</sub>
</div>
