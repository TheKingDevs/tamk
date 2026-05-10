# 🤖 T.A.M.K — Agent Instructions

**T.A.M.K (Termux APK Manager Kit)** is a professional automation framework for native Android app development directly in Termux. Built with Python and Kotlin, it enables developers to create, build, sign, and install APK applications using only a mobile device.

## 📑 DYNAMIC INDEX

1. [🌍 Overview & Philosophy](#-overview--philosophy)
2. [🏗️ Architecture & Data Flow](#-architecture--data-flow)
3. [⌨️ Project Development Guide](#-project-development-guide)
4. [🛠️ Template System & Customization](#-template-system--customization)
5. [🗄️ Build System & Configuration](#-build-system--configuration)
6. [🔌 Services & Integrations](#-services--integrations)
7. [🛡️ Stability, Debugging & Errors](#-stability-debugging--errors)
8. [📋 Command Catalog](#-command-catalog)

---

## 🌍 OVERVIEW & PHILOSOPHY

The project abandons rigid code generation in favor of a **template-oriented architecture**.

*   **Modularity:** Each project type (UI, Console, WebApp) is isolated in its own structure class.
*   **Template Engine:** Code generation via `.tmpl` files with placeholder injection.
*   **Tech Stack:** Python 3+ (core), Kotlin (Android templates), OpenJDK 21, Termux tools.
*   **Philosophy:** "Code for humans: simple, readable, and decoupled."
*   **Target:** Developers seeking hardware independence for Android development.

---

## 🏗️ ARCHITECTURE & DATA FLOW

### 1. CLI Entry Point (`src/main.py`)
*   Interprets command-line arguments (`--create`, `--build`, `--install`, `--setup`).
*   Routes to appropriate controllers based on user input.

### 2. Controllers (`src/controllers/`)
*   **`ProjectManager`**: Orchestrates project creation, runs interactive wizard.
*   **`BuildController`**: Manages compilation, signing, and APK alignment.
*   **`SetupController`**: Configures environment and SDK paths.

### 3. Factory Pattern (`src/organization/factory.py`)
*   **`ProjectFactory`**: Central distribution point for project types.
*   Instantiates the correct structure class based on user selection.

```python
class ProjectFactory:
    @staticmethod
    def create(p_type, name, version, author, password):
        structures = {
            "ui_apk": UIAppStructure(),
            "console": ConsoleStructure(),
            "webapp": WebAppStructure()
        }
        builder = structures.get(p_type)
        if builder:
            return builder.setup(name, version, author, password)
        return False
```

### 4. Structure Classes (`src/organization/structures/`)
*   Define the skeleton for each project type.
*   Create folder trees and process templates.
*   Generate private Keystore for each project.

### 5. Template Engine (Embedded in structures)
*   Reads `.tmpl` files from `assets/templates/`.
*   Replaces placeholders (`{{NAME}}`, `{{PACKAGE}}`, `{{VERSION}}`).
*   Writes final files to project directory.

---

## ⌨️ PROJECT DEVELOPMENT GUIDE

### Project Types

| Type | Description | Use Case |
| :--- | :--- | :--- |
| **UI/APK** | Full Android app with native UI (XML + Kotlin) | Traditional Android apps |
| **Console** | Command-line Kotlin application | Learning, scripts, tools |
| **WebApp** | WebView wrapper for HTML/CSS/JS | Hybrid apps, web portfolios |

### Creating a New Project

```bash
tamk --create
```

1.  Select project type (UI, Console, WebApp).
2.  Provide metadata: Name, Version, Author, Keystore Password.
3.  T.A.M.K generates the complete project structure.

### Project Structure (WebApp Example)

```
MyWebApp/
├── src/main/
│   ├── assets/          # Web content (HTML, CSS, JS)
│   ├── kotlin/          # Generated Kotlin code
│   └── res/             # Android resources
├── AndroidManifest.xml
├── tamk.config
└── secret/
    └── project.keystore # Private signing key
```

### Build & Install

```bash
# Build APK
tamk --build -p YOUR_PASSWORD

# Install on device
tamk --install
```

---

## 🛠️ TEMPLATE SYSTEM & CUSTOMIZATION

### Template Files (`assets/templates/`)

Templates use `.tmpl` extension with placeholder syntax.

| Template | Purpose |
| :--- | :--- |
| `webapp/MainActivity.kt.tmpl` | WebView activity logic |
| `webapp/index.html.tmpl` | Default landing page |
| `ui_apk/MainActivity.kt.tmpl` | Native UI activity |
| `console/Main.kt.tmpl` | Console entry point |

### Placeholders Reference

| Placeholder | Description | Example |
| :--- | :--- | :--- |
| `{{NAME}}` | App name defined by user | `Meu WebApp Incrível` |
| `{{PACKAGE}}` | Java/Kotlin package | `com.exemplo.meuwebapp` |
| `{{VERSION}}` | App version string | `1.0.0` |
| `{{AUTHOR}}` | Project author name | `João da Silva` |

### Customizing Templates

1.  Modify `.tmpl` files in `assets/templates/`.
2.  Test changes with `tamk --create`.
3.  **Never** modify generated project files directly in `assets/`.

### WebView Configuration (WebApp)

The generated `MainActivity.kt` configures WebView for optimal performance:

```kotlin
webView.settings.apply {
    javaScriptEnabled = true      // Enable JS execution
    domStorageEnabled = true      // Enable localStorage
    allowFileAccess = true        // Allow local file access
}
webView.loadUrl("file:///android_asset/index.html")
```

---

## 🗄️ BUILD SYSTEM & CONFIGURATION

### Build Pipeline (`src/controllers/build_controller.py`)

1.  **Validate Keystore**: Check password and credentials.
2.  **Compile Resources**: `aapt2` processes XML, images, etc.
3.  **Package Assets**: Include `assets/` folder (WebApp specific).
4.  **Compile Kotlin**: Generate DEX from Kotlin source.
5.  **Link & Sign**: Combine resources, sign with `apksigner`.
6.  **Align APK**: Optimize with `zipalign`.

### Key Build Flags

```python
# Assets flag for WebApps
assets_flag = "-A src/main/assets" if os.path.exists(assets_path) else ""

# aapt2 link command
link_cmd = (f"aapt2 link -I {sdk_path} --manifest AndroidManifest.xml "
            f"--java {cache_dir}/gen -o {cache_dir}/app.apk "
            f"{assets_flag} {cache_dir}/res.zip --auto-add-overlay")
```

### Configuration Files

| File | Purpose |
| :--- | :--- |
| `settings.local.json` | Local environment config (SDK paths, etc.) |
| `tamk.config` | Project-specific metadata (version, name) |
| `secret/project.keystore` | Private signing key (per project) |

---

## 🔌 SERVICES & INTEGRATIONS

### Termux Tools Integration

| Tool | Purpose |
| :--- | :--- |
| `aapt2` | Android Asset Packaging Tool |
| `apksigner` | APK signing and verification |
| `zipalign` | APK optimization for performance |
| `d8` | DEX compiler for Kotlin/Java |
| `termux-tools` | Package installation |

### SDK Isolation

Each project contains its own `android.jar` copy:
*   Prevents version conflicts between projects.
*   Ensures portability across environments.
*   Managed automatically by `SetupController`.

### Keystore Management

*   **Per-project**: Each project gets a unique `project.keystore`.
*   **Debug fallback**: Global debug keystore used if project keystore is missing.
*   **Security**: Never share or log keystore passwords.

---

## 🛡️ STABILITY, DEBUGGING & ERRORS

### Common Errors

| Error | Cause | Solution |
| :--- | :--- | :--- |
| `Keystore password incorrect` | Wrong password in `--build` | Re-enter correct password |
| `aapt2 not found` | Missing dependency | Run `pkg install aapt2` |
| `SDK not configured` | SDK path not set | Run `tamk --setup` |
| `Build failed: D8 error` | Kotlin syntax error | Check generated Kotlin files |

### Debugging WebApps

1.  **JavaScript Alerts**: Use `alert()` for quick variable inspection.
2.  **Remote Debugging**: Enable in `MainActivity.kt`:
    ```kotlin
    WebView.setWebContentsDebuggingEnabled(true)
    ```
3.  **Chrome Inspect**: Connect device via USB, visit `chrome://inspect`.

### Logging

*   Build logs are output to console in real-time.
*   Errors include stack traces for debugging.
*   Enable verbose mode for detailed output.

### Best Practices for Stability

1.  **Validate before build**: Check Keystore credentials early.
2.  **Cache management**: Use `cache_dir` for temporary build files.
3.  **Clean builds**: Remove `cache/` folder if encountering strange errors.
4.  **Template testing**: Always test template changes with a fresh project.

---

## 📋 COMMAND CATALOG

### 🔧 CORE COMMANDS

| Command | Description | Args |
| :--- | :--- | :--- |
| `tamk --version` | Check installed version | None |
| `tamk --create` | Interactive project wizard | Interactive |
| `tamk --build` | Compile and sign APK | `-p <password>` |
| `tamk --install` | Install APK on device | None |
| `tamk --setup` | Configure environment | Interactive |

### 📁 PROJECT TYPES (via `--create`)

| Type | Folder | Templates |
| :--- | :--- | :--- |
| **UI/APK** | `ui_apk/` | XML layouts, Kotlin activities |
| **Console** | `console/` | Kotlin main function |
| **WebApp** | `webapp/` | WebView, HTML/CSS/JS |

---

## 📚 DOCUMENTATION REFERENCE

| File | Purpose |
| :--- | :--- |
| `ARCHITECTURE.md` | System architecture and data flow diagrams |
| `API_COMPONENTS.md` | Class and module reference |
| `DEV_GUIDE.md` | Development workflow and debugging |
| `QUICKSTART.md` | Step-by-step getting started guide |
| `WEBAPP_TEMPLATES.md` | WebApp-specific template docs |
| `HMR_SYSTEM.md` | **Hot Module Replacement** - Desenvolvimento em tempo real |
| `CONSOLE_TEMPLATES.md` | Console template documentation |
| `FAQ.md` | Frequently asked questions |
| `CHANGELOG.md` | Version history and changes |
| `CONTRIBUTING.md` | Contribution guidelines |

---

## 🎯 AGENT BEST PRACTICES

1.  **Always read** `./AGENTS.md` first for context.
2.  **Check existing docs** in `documentation/` before creating new files.
3.  **Preserve template structure** when modifying `.tmpl` files.
4.  **Test changes** with `tamk --create` and `tamk --build`.
5.  **Update CHANGELOG.md** for new features or fixes.
6.  **Respect user settings** in `settings.local.json`.
7.  **Never expose credentials** (Keystore passwords, paths).
8.  **Understand HMR system** - Review `HMR_SYSTEM.md` for WebApp dev workflows.
9.  **Incremental builds** - `build_assets_only()` is key for HMR; do not break it.

---

## ⚠️ CRITICAL NOTES

*   **Templates**: All templates are in `assets/templates/` — do not modify without testing.
*   **Keystore**: Sensitive files in `secret/` — never log or expose passwords.
*   **Environment**: Assumes Termux with Python 3, OpenJDK 21, Kotlin, aapt2, apksigner.
*   **SDK Path**: Configured via `settings.local.json` or `tamk --setup`.
*   **Build Cache**: Temporary files in `cache/` — safe to delete for clean builds.

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Made with ❤️ by @mrx_dev</sub>
</div>
