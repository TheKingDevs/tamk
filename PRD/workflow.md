# Product Requirements Document — T.A.M.K Go

## Visão Geral
T.A.M.K (Termux APK Manager Kit) v1.0.0 — framework para desenvolvimento Android nativo diretamente no Termux. Escrito em Go com Clean Architecture v4. Permite criar, compilar, assinar e instalar APKs Android sem um PC.

## Funcionalidades (MVP)

### 1. Criação de Projetos
- **WebApp** (WebView + HTML/CSS/JS) — modo interno (assets) ou externo (URL remota)
- **UI APK Nativo** (XML + Kotlin)
- **Console Kotlin CLI**
- Validação: nome, autor, versão (SEMVER), package name Android
- Geração de keystore RSA 2048 por projeto (`secret/project.keystore`)
- 19 templates `.tmpl` com placeholders (`{{NAME}}`, `{{PACKAGE}}`, etc.)

### 2. Build de APKs
- Pipeline completo: AAPT2 → kotlinc → D8 → zipalign → apksigner
- Build incremental (assets only) para modo dev
- Cache SHA-256 (pular se nada mudou)
- Validação de keystore (mín 6 caracteres)
- Timeout: 10min full build, 5min incremental

### 3. Modo Dev (HMR)
- File watcher (fsnotify) em `src/main/assets/`
- Bridge JS inline injetada no index.html (WebSocket client-side)
- Assets build rápido via `AssetsOnlyBuild()` — extrai APK, troca assets, re-assina
- ADB push para HTML/CSS/JS + broadcast `ACTION_REFRESH`
- Extensões monitoradas: `.html`, `.css`, `.js`, `.json`, `.png`, `.jpg`, `.svg`, `.webp`, `.xml`, `.kt`

### 4. Setup de Ambiente
- Download SDK android.jar (platform-30)
- Geração de debug keystore
- Verificação de dependências (aapt2, kotlinc, d8, apksigner, zipalign)

### 5. Instalação de APKs
- Servidor HTTP embutido + QR code terminal via `tamk install`
- side-load por ADB

### 6. Sistema de Atualização
- GitHub API release checker (cache 6h)
- Prioridades: CRITICAL/PATCH (auto-install), MAJOR/MINOR (prompt), OPTIONAL (notify)
- Métodos: `git pull --rebase --autostash` ou `go install`

### 7. Shell Interativo (REPL)
- Comandos: `create`, `build`, `dev`, `setup`, `install`, `update`, `version`
- Auto-completion de comandos

## Requisitos Não-Funcionais
- **Performance**: Build completo < 2min, incremental < 10s
- **Segurança**: Senhas zeradas após uso, caminhos sanitizados, `exec.Command` (sem shell=True)
- **Portabilidade**: Termux, SmartIDE, Linux desktop
- **Observabilidade**: Logs estruturados (slog) com níveis DEBUG → ERROR + ANSI colors
- **Testabilidade**: ≥ 70% cobertura domain + usecase (testify)

## Roadmap

### ✅ Fase 1 — Core (Concluído)
- Domain entities + value objects + repository interfaces
- Use cases: CreateProject, BuildProject, DevMode, Setup, Install, Update
- CLI com Cobra + wizard interativo + REPL
- 19 templates `.tmpl`
- Build pipeline completo + cache

### ✅ Fase 2 — Dev Mode (Concluído)
- File watcher (fsnotify)
- Assets incremental build
- HMR bridge inline JS
- ADB push (stub funcional)

### 🔄 Fase 3 — Qualidade (Em andamento)
- Testes unitários domain + usecase
- Benchmarks + regression guard (pre-commit)
- Docker + CI/CD (GitHub Actions)
- Documentação completa (23 docs + AGENTS.md)

## Stack
- **Linguagem**: Go 1.26+
- **Android**: Kotlin, OpenJDK 21, Android SDK (platform-30)
- **Ferramentas**: AAPT2, D8, zipalign, apksigner
- **CLI**: Cobra + golang.org/x/term
- **Logger**: slog + ANSI colors
- **File Watcher**: fsnotify
- **Testes**: testing padrão + testify
