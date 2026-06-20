# 📋 API Components — Complete Reference

> **Version:** 2026.3.0-HMR — Technical reference for packages, templates, and utilities of T.A.M.K.

---

## Index

- [1. Go Packages](#1-go-packages)
- [2. Templates](#2-templates)
- [3. Placeholders](#3-placeholders)
- [4. Utilities](#4-utilities)
- [5. Types and Constants](#5-types-and-constants)

---

## 1. Go Packages

### `cmd/tamk/main.go` — Entry Point

| Function | Cobra Command | Description |
| :--- | :--- | :--- |
| `main()` | — | Initializes config, repositories, use cases; executes root cobra command |

### `internal/delivery/cli/root.go` — Cobra Root + Wiring

| Function | Command | Description |
| :--- | :--- | :--- |
| `NewRootCmd()` | — | Creates root cobra command with all subcommands wired |
| `newCreateCmd(uc)` | `create` | Instantiates `CreateProjectUseCase` interactive |
| `newBuildCmd(uc, repo)` | `build` | Instantiates `BuildProjectUseCase` with password prompt |
| `newDevCmd(uc, repo, cfg)` | `dev` | Starts `DevModeUseCase` for HMR |
| `newSetupCmd(uc)` | `setup` | Executes `SetupEnvironmentUseCase` |
| `newRunCmd()` | `run` | Kotlin/Java runner (stub) |
| `newInstallCmd(projRepo)` | `install` | APK installer via HTTP server + QR code |
| `newUpdateCmd(uc)` | `update` | Executes `UpdateUseCase.Check()` |
| `newShellCmd(cfg, createUC, buildUC, devUC, setupUC, updateUC, projRepo)` | `shell` | Interactive development REPL |
| `newVersionCmd()` | `version` | Displays version + environment info |

### `internal/usecase/project_create.go`

```go
type CreateProjectUseCase struct { /* ... */ }
func NewCreateProjectUseCase(cfg, projRepo, tmplRepo) *CreateProjectUseCase
func (uc *CreateProjectUseCase) Execute(input CreateProjectInput) (*CreateProjectOutput, error)
```

| Field/Method | Type | Description |
| :--- | :--- | :--- |
| `CreateProjectInput.Name` | `string` | Sanitized project name |
| `CreateProjectInput.Version` | `string` | SEMVER version |
| `CreateProjectInput.Author` | `string` | Author identifier |
| `CreateProjectInput.Password` | `string` | Keystore password |
| `CreateProjectInput.WebURL` | `string` | URL for external WebApp mode |
| `CreateProjectInput.WebMode` | `WebContentMode` | Internal vs External |

### `internal/usecase/project_build.go`

```go
type BuildProjectUseCase struct { /* ... */ }
func NewBuildProjectUseCase(cfg, buildRepo, projRepo) *BuildProjectUseCase
func (uc *BuildProjectUseCase) FullBuild(ctx context.Context, input BuildInput) (*entity.BuildResult, error)
func (uc *BuildProjectUseCase) AssetsOnlyBuild(ctx context.Context, input BuildInput) (*entity.BuildResult, error)
```

### `internal/usecase/development.go`

```go
type DevModeUseCase struct { /* ... */ }
func NewDevModeUseCase(cfg, buildUC, projRepo) *DevModeUseCase
func (uc *DevModeUseCase) Start(ctx context.Context, projectPath, password string) error
func (uc *DevModeUseCase) Stop(ctx context.Context)
```

*(Watch extensions and ignore patterns are defined in `pkg/watcher/watcher.go`)*

### `internal/usecase/environment_setup.go`

```go
type SetupEnvironmentUseCase struct { /* ... */ }
func NewSetupEnvironmentUseCase(cfg) *SetupEnvironmentUseCase
func (uc *SetupEnvironmentUseCase) Execute(ctx context.Context) error
```

| Constant | Value |
| :--- | :--- |
| `SDK_URL` | `https://dl.google.com/android/repository/platform-30_r03.zip` |

### `internal/usecase/update.go`

```go
type UpdateUseCase struct { /* ... */ }
func NewUpdateUseCase(cfg, updateRepo) *UpdateUseCase
func (uc *UpdateUseCase) Check(ctx context.Context) (*entity.UpdateInfo, error)
func (uc *UpdateUseCase) ShouldAutoInstall(info *entity.UpdateInfo) bool
func (uc *UpdateUseCase) ShouldPrompt(info *entity.UpdateInfo) bool
func (uc *UpdateUseCase) PrintUpdateInfo(info *entity.UpdateInfo)
```

### `internal/usecase/install.go`

```go
type InstallUseCase struct { /* ... */ }
func NewInstallUseCase() *InstallUseCase
func (uc *InstallUseCase) Serve(ctx context.Context, projectPath string, port int) (*InstallOutput, error)
func (uc *InstallUseCase) FindAPK(projectPath string) (string, error)
func (uc *InstallUseCase) FindAPKs(projectPath string) []string
```

### `internal/usecase/zip.go`

```go
func zipDir(source, target string) error  // Archive helper (zipDir)
```

### `internal/repository/filesystem/project_repository.go`

| Method | Description |
| :--- | :--- |
| `Exists(path) bool` | Check if tamk.config exists |
| `Create(project) error` | Create project directories and files |
| `LoadConfig(path) (*entity.Project, error)` | Read and parse tamk.config |
| `SaveConfig(project) error` | Write tamk.config |

### `internal/repository/filesystem/build_repository.go`

| Method | Description |
| :--- | :--- |
| `CalculateHash(projectPath) (string, error)` | SHA-256 recursive hash |
| `MustRecompile(cachedHash string) bool` | Compare hash vs `.build_cache` |
| `CompileResources(projectPath) error` | aapt2 compile + link |
| `CompileKotlin(projectPath) error` | kotlinc -d obj/ |
| `RunD8(projectPath) error` | d8 -> classes.dex |
| `PackageDex(projectPath) error` | zip -j app.apk classes.dex |
| `Zipalign(projectPath) error` | zipalign -f 4 |
| `Sign(projectPath, password) error` | apksigner sign |

### `internal/repository/filesystem/template_repository.go`

| Method | Description |
| :--- | :--- |
| `LoadTemplate(tmplDir, name) (string, error)` | Read .tmpl file |
| `Render(tmpl, dest, replacements) error` | Apply replacements, write file |
| `RenderAll(type, destDir, replacements) error` | Render all templates for type |

### `internal/config/config.go`

```go
const Version = "2026.3.0-HMR"

type Config struct {
    Version    string       // "2026.3.0-HMR"
    Env        Environment  // "termux" | "debian" | "ubuntu" | "arch" | "fedora" | "unknown"
    EnvType    string       // TAMK_ENV env var
    TAMKHome   string       // Installation root
    DevDir     string       // development/
    SDKPath    string       // development/sdk/android.jar
    Keystore   string       // development/secret/debug.keystore
    PkgMgr     string       // Detected package manager
    ProjectDir string       // Project base directory
}

func New() *Config
func ConfigFromEnv() *Config
func (c *Config) GetTemplateDir(templateType string) string
func (c *Config) GetProjectDir(name string) string
func (c *Config) IsTermux() bool
func (c *Config) Validate() error
func (c *Config) DetectPackageName(author, name string) string
```

---

## 2. Templates

### WebApp (`templates/webapp/`)

| Template | Destination | Purpose |
| :--- | :--- | :--- |
| `AndroidManifest.xml.tmpl` | `AndroidManifest.xml` | Permissions, activities, theme |
| `MainActivity.kt.tmpl` | `src/.../MainActivity.kt` | WebView + BroadcastReceiver |
| (injected inline) | injected in dev mode | HMR WebSocket client bridge |
| `index.html.tmpl` | `src/main/assets/index.html` | Landing page (internal) |
| `css/styles.css.tmpl` | `src/main/assets/css/styles.css` | Global styles (internal) |
| `js/app.js.tmpl` | `src/main/assets/js/app.js` | Main JS (internal) |
| `strings.xml.tmpl` | `res/values/strings.xml` | Resource strings |
| `styles.xml.tmpl` | `res/values/styles.xml` | Material theme |
| `icon.xml.tmpl` | `res/drawable/ic_launcher*.xml` | Vector icon |
| `network_security_config.xml.tmpl` | `res/xml/network_security_config.xml` | Network traffic config |
| `gitignore_root.tmpl` | `.gitignore` | Project root |
| `gitignore_assets.tmpl` | `src/main/assets/.gitignore` | Assets (internal) |

### UI APK (`templates/ui_apk/`)

| Template | Destination | Purpose |
| :--- | :--- | :--- |
| `AndroidManifest.xml.tmpl` | `AndroidManifest.xml` | Standard manifest |
| `MainActivity.kt.tmpl` | `src/.../MainActivity.kt` | Button Activity |
| `activity_main.xml.tmpl` | `res/layout/activity_main.xml` | XML layout |
| `strings.xml.tmpl` | `res/values/strings.xml` | Strings |
| `styles.xml.tmpl` | `res/values/styles.xml` | Theme |
| `icon.xml.tmpl` | `res/drawable/ic_launcher*.xml` | Icon |

### Console (`templates/console/`)

| Template | Destination | Purpose |
| :--- | :--- | :--- |
| `Main.kt.tmpl` | `src/Main.kt` | Console entry point |

---

## 3. Placeholders

| Placeholder | Usage | Example |
| :--- | :--- | :--- |
| `{{NAME}}` | All | `MeuApp` |
| `{{PACKAGE}}` | WebApp, UI APK | `com.author.meuapp` |
| `{{VERSION}}` | All | `1.0.0` |
| `{{AUTHOR}}` | All | `JoaoSilva` |
| `{{WEB_URL}}` | WebApp | `file:///android_asset/index.html` |
| `{{DEV_MODE}}` | WebApp | `false` |
| `{{MIN_SDK}}` | WebApp | `21` |
| `{{TARGET_SDK}}` | WebApp | `30` |
| `{{TAMK_VERSION}}` | WebApp | `2026.3.0-HMR` |
| `{{DEV_PORT}}` | dev_bridge.js | `8765` |

---

## 4. Utilities

### Logger (`pkg/logger/logger.go`)

```go
// Custom slog levels
const (
    LevelDebug   = slog.Level(-4)
    LevelInfo    = slog.Level(0)
    LevelStep    = slog.Level(1)
    LevelSuccess = slog.Level(2)
    LevelWarning = slog.Level(4)
    LevelError   = slog.Level(8)
)

func Init(verbose bool)
func Debug(msg string, args ...any)
func Info(msg string, args ...any)
func Step(msg string, args ...any)
func Success(msg string, args ...any)
func Warn(msg string, args ...any)
func Error(msg string, args ...any)
```

### File Watcher (`pkg/watcher/watcher.go`)

```go
type FileChangeHandler func(path string)
type ExtSet map[string]bool

type Watcher struct { /* ... */ }

var WatchExtensions = ExtSet{".html": true, ".css": true, ".js": true, ".json": true, ".png": true, ".jpg": true, ".jpeg": true, ".svg": true, ".webp": true, ".xml": true, ".kt": true}
var IgnoreDirs = map[string]bool{".git": true, "node_modules": true, "secret": true, ".idea": true}

func New(handler FileChangeHandler, debounce time.Duration) *Watcher
func (w *Watcher) Start(path string) error
func (w *Watcher) Stop() error
func (w *Watcher) IsRunning() bool
```

*(Template engine is in `internal/repository/filesystem/template_repository.go`)*

*(Update logic is in `internal/usecase/update.go` and `internal/repository/update_repository.go`)*

---

## 5. Types and Constants

### Config (`tamk.config` — project file)

```
type=webapp|ui_apk|console
name=ProjectName
version=1.0.0
author=AuthorName
package=com.author.name
web_url=file:///android_asset/index.html
created_with=2026.3.0-HMR
```

### CLI Exit Codes

| Code | Meaning |
| :--- | :--- |
| 0 | Success |
| 1 | Generic error |
| 130 | Cancelled (Ctrl+C) |

### Supported File Extensions

| Context | Extensions |
| :--- | :--- |
| File Watcher | `.html`, `.css`, `.js`, `.json`, `.png`, `.jpg`, `.jpeg`, `.svg`, `.webp`, `.xml`, `.kt` |
| Build hash | Everything in `src/`, `res/`, `AndroidManifest.xml` |

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Reference Documentation</sub>
</div>
