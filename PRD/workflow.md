# Product Requirements Document — T.A.M.K Go

## Visão Geral
T.A.M.K (Termux APK Manager Kit) v1.0.0 — framework para desenvolvimento Android nativo diretamente no Termux. Escrito em Go com Clean Architecture v4. Permite criar, compilar, assinar e instalar APKs Android sem um PC.

---

## Funcionalidades

### 1. Criação de Projetos
- **WebApp** (WebView + HTML/CSS/JS) — modo interno (assets) ou externo (URL remota)
- **UI APK Nativo** (XML + Kotlin)
- **Console Kotlin CLI** — boilerplate com Calculator e StringUtils
- Validação: nome, autor, versão (SEMVER), package name Android
- Geração de keystore RSA 2048 por projeto (`secret/project.keystore`)
- 18 templates `.tmpl` organizados em `xml/`, `kotlin/`, `css/`, `js/`
- Projetos console não incluem campos web (web_url, web_mode)

### 2. Build de APKs
- Pipeline completo via ToolManager: AAPT2 → kotlinc → D8 → zipalign → apksigner
- Build incremental (assets only) para modo dev
- Cache SHA-256 (pular se nada mudou)
- Validação de keystore **antes** do build (keytool -list)
- Timeout: 10min full build, 5min incremental

### 3. Execução de Código (`tamk run`)
- Compila e executa arquivos Kotlin/Java nativamente
- Auto-detecção de `Main.kt` em `src/`
- Resolução automática de classe principal
- Uso do ToolManager para caminhos de ferramentas

### 4. Modo Dev (HMR)
- File watcher (fsnotify) em `src/main/assets/`
- Bridge JS inline injetada no index.html
- **WebSocket server em `:8765`** para push de atualizações em tempo real
- **Mensagens HMR**: `reload`, `css-update`, `js-update` via WebSocket
- Assets build rápido via `AssetsOnlyBuild()`
- ADB push para HTML/CSS/JS + broadcast `ACTION_REFRESH`
- Extensões: `.html`, `.css`, `.js`, `.json`, `.png`, `.jpg`, `.jpeg`, `.svg`, `.webp`, `.xml`, `.kt`
- **Fallback**: Sem clientes HMR conectados → rebuild completo dos assets

### 5. Setup de Ambiente (`tamk setup`)
- Extração de ferramentas embutidas para `~/.tamk/tools/{os}/`
- Download SDK android.jar (platform-30)
- Geração de debug keystore
- Version marker (`.version`) para re-extração apenas na atualização do TAMK

### 6. Instalação de APKs (`tamk install`)
- Servidor HTTP embutido + QR code terminal
- side-load por ADB

### 7. Sistema de Atualização (`tamk update`)
- GitHub API release checker (cache 6h)
- Prioridades: CRITICAL/PATCH (auto-install), MAJOR/MINOR (prompt), OPTIONAL (notify)
- **Auto-install via `git pull --rebase --autostash`** (se `.git` em TAMK_HOME)
- **Auto-install via `go install @latest`** (fallback)
- Detecção automática do método de atualização

### 8. Shell Interativo (`tamk shell`)
- Comandos: `create`, `build`, `dev`, `setup`, `install`, `update`, `run`, `version`
- **`run <arquivo>` agora executa Kotlin/Java via RunUseCase** (não mais stub)

---

## Arquitetura ToolManager

### Conceito
Sistema de gerenciamento de ferramentas com suporte multiplataforma:
- **Linux/macOS/Android**: Usa ferramentas do sistema (leve)
- **Windows**: Extrai ferramentas embutidas via `go:embed` (zero config)

### "Setup Once, Run Forever"
1. Primeira execução: extrai ferramentas para `~/.tamk/tools/{os}/`
2. Marca versão em `.version`
3. Execuções futuras: usa ferramentas extraídas (instantâneo)
4. Re-extração: apenas quando versão do TAMK muda

### Ferramentas Embutidas (Windows)
| Arquivo | Tamanho | Fonte |
|:--|:--|:--|
| aapt2.exe | 4.4MB | Android SDK Build-Tools r34 |
| apksigner.jar | 1MB | Android SDK Build-Tools r34 |
| zipalign.exe | 799KB | Android SDK Build-Tools r34 |
| d8.jar | 15MB | Android SDK Build-Tools r34 |
| kotlinc.zip | 82MB | Kotlin Compiler 2.0.21 |
| android.jar | 21MB | Android SDK platform-30 |

---

## Plataformas Suportadas

| Plataforma | GOOS/GOARCH | Status | Ferramentas |
|:--|:--|:--|:--|
| Linux x64 | linux/amd64 | ✅ | Sistema |
| Linux ARM64 | linux/arm64 | ✅ | Sistema |
| Android ARM64 | android/arm64 | ✅ | Sistema (pkg) |
| macOS ARM64 | darwin/arm64 | ✅ | Sistema |
| Windows x64 | windows/amd64 | ✅ | Embutidas |

---

## Scripts de Instalação

| Script | Plataforma | Destaques |
|:--|:--|:--|
| `scripts/install-termux.sh` | Termux | Auto-detect deps, build from source |
| `scripts/install-linux.sh` | Debian/Ubuntu/Arch/Fedora/Alpine | Multi-distro, package manager detection |
| `scripts/install-macos.sh` | macOS | Homebrew integration, Intel + Apple Silicon |
| `scripts/install-windows.bat` | Windows | PATH setup, environment variables |
| `setup-install.sh` | Todos | Dispatcher automático por plataforma |

---

## CI/CD (GitHub Actions)

| Workflow | Trigger | Propósito |
|:--|:--|:--|
| `ci.yml` | push/PR main | Lint + Test + Build |
| `build.yml` | push/PR + manual | Matrix build multi-plataforma |
| `release.yml` | tag `v*.*.*` | Cross-platform build + GitHub Release |
| `tags.yml` | tag + manual | Validação semântica + release |
| `security.yml` | push/PR + diário | govulncheck + staticcheck + golangci-lint |
| `maintenance.yml` | semanal | Auto dependency updates |
| `test-windows.yml` | push/PR (tools) | Cross-compile + teste Windows runner |
| `tests-tt.yml` | push/PR + semanal | T2T doc-code audit |

---

## Requisitos Não-Funcionais

- **Performance**: Build completo < 2min, incremental < 10s
- **Segurança**: Validação keystore antes do build, caminhos sanitizados, `exec.Command` (sem shell=True)
- **Portabilidade**: Termux, Linux (5 distros), macOS, Windows
- **Observabilidade**: Logs estruturados (slog) com níveis DEBUG → ERROR + ANSI colors
- **Testabilidade**: Testes em `internal/usecase/*_test.go` (Clean Architecture)

---

## Stack

- **Linguagem**: Go 1.26+
- **Android**: Kotlin, OpenJDK 21, Android SDK (platform-30)
- **Ferramentas**: AAPT2, D8, zipalign, apksigner (via ToolManager)
- **CLI**: Cobra
- **Logger**: slog + ANSI colors
- **File Watcher**: fsnotify
- **WebSocket**: gorilla/websocket (HMR server)
- **Embed**: go:embed (ferramentas Windows + android.jar)
- **Testes**: testing padrão

---

## Roadmap

### ✅ Fase 1 — Core
- Domain entities + value objects + repository interfaces
- Use cases: CreateProject, BuildProject, DevMode, Setup, Install, Update, Run
- CLI com Cobra + wizard interativo + REPL
- 18 templates `.tmpl` organizados por tipo
- Build pipeline completo via ToolManager

### ✅ Fase 2 — Dev Mode
- File watcher (fsnotify)
- Assets incremental build
- HMR bridge inline JS
- **WebSocket server (:8765) para push de atualizações**
- ADB push + broadcast

### ✅ Fase 3 — Plataforma & CI/CD
- ToolManager "Setup Once, Run Forever"
- Windows embedded tools (go:embed)
- 8 GitHub Actions workflows
- Platform-specific installers
- T2T doc-code audit
- Security scanning (govulncheck, staticcheck, golangci-lint)

### 🔄 Fase 4 — Qualidade (Em andamento)
- Cobertura de testes ≥ 70% (atual: 13.8% usecase, 100% errors, 85.7% entity, 97.1% valueobject)
- Benchmarks + regression guard
- Docker multi-stage build
- Documentação completa (23 docs + AGENTS.md)
- **HMR WebSocket server com gorilla/websocket**
- **Update auto-install (git/go)**
- **Shell run command funcional**

---

## Comandos CLI

| Comando | Descrição | Status |
|:--|:--|:--|
| `tamk version` | Mostra versão + ambiente | ✅ |
| `tamk create` | Wizard interativo de criação | ✅ |
| `tamk build -p SENHA` | Build completo de APK | ✅ |
| `tamk run [arquivo]` | Executa Kotlin/Java nativamente | ✅ |
| `tamk dev` | Modo dev com HMR | ✅ |
| `tamk setup` | Download SDK + keystore + tools | ✅ |
| `tamk install [porta]` | Serve APK via HTTP + QR | ✅ |
| `tamk update` | Verifica atualizações | ✅ |
| `tamk shell` | REPL interativo | ✅ |
