# Product Requirements Document — T.A.M.K Go

## Visão Geral
T.A.M.K (Termux APK Manager Kit) é um framework para desenvolvimento Android nativo diretamente no Termux. Permite criar, compilar, assinar e instalar APKs Android sem um PC.

## Funcionalidades (MVP)

### 1. Criação de Projetos
- WebApp (WebView + HTML/CSS/JS) — interno (assets) ou externo (URL)
- UI APK Nativo (XML + Kotlin)
- Console Kotlin CLI
- Validação: nome, autor, versão (SEMVER), package name
- Geração de keystore RSA 2048 por projeto
- 19 templates .tmpl com placeholders ({{NAME}}, {{PACKAGE}}, etc.)

### 2. Build de APKs
- Pipeline completo: AAPT2 → kotlinc → D8 → zipalign → apksigner
- Build incremental (assets only) para HMR
- Cache baseado em SHA-256 (pular se nada mudou)
- Keystore validation (mín 6 chars)

### 3. Modo Dev (HMR)
- WebSocket server para hot reload
- HTTP server fallback
- File watcher (fsnotify) para HTML/CSS/JS/JSON
- ADB bridge: push + broadcast refresh intent
- Client-side JS bridge (TAMK_HMR API)

### 4. Setup de Ambiente
- Download SDK android.jar (platform-30)
- Geração de debug keystore (password: android)
- Verificação de dependências (aapt2, kotlinc, d8, apksigner)

### 5. Execução de Snippets
- Compilar/executar .kt (kotlinc -include-runtime) e .java (javac)
- Auto-detecção de src/Main.kt em projetos Console

### 6. Instalação de APKs
- termux-share (nativo Termux)
- Fallback via am start + Intent.VIEW

### 7. Sistema de Atualização
- GitHub API release checker (6h cache)
- Prioridades: CRITICAL (auto), PATCH (auto), MAJOR/MINOR (prompt)
- Métodos: git pull, pip, script

## Requisitos Não-Funcionais
- **Performance**: Build completo < 2min, incremental < 10s
- **Segurança**: Senhas zeradas após uso, caminhos sanitizados, sem shell=True
- **Portabilidade**: Linux (Termux, SmartIDE, desktop)
- **Observabilidade**: Logs estruturados com níveis (DEBUG → ERROR)
- **Testabilidade**: ≥ 70% cobertura domain + usecase

## Roadmap
- Fase 1: Core domain + usecases + CLI
- Fase 2: Build pipeline + HMR dev mode
- Fase 3: Testes + Docker + CI/CD
