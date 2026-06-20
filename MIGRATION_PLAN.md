# Migration Plan: Python → Go (Clean Architecture)

## Overview
Migrate T.A.M.K (Termux APK Manager Kit) v2026.3.0-HMR from Python to Go using Clean Architecture v4.

## Current Architecture (Python)
```
src/
├── main.py                     # CLI entry point (argparse)
├── config/tamk_config.py       # Central config (VERSION, paths, env)
├── controllers/                # Business logic (6 controllers)
├── organization/               # Project factory + structures (4 files)
└── utils/                      # Helpers (6 modules)
```

## Target Architecture (Go)
```
tamk/
├── cmd/tamk/main.go            # Entry point
├── internal/                   # NOT importable by external modules
│   ├── domain/                 # Entities, value objects, repository interfaces
│   ├── usecase/                # Application business rules
│   ├── repository/             # Concrete implementations
│   └── delivery/               # CLI + HTTP handlers
├── pkg/                        # Shared utilities
└── config/                     # Configuration
```

## Migration Order (Layer by Layer)

### Phase 1: Domain (Innermost)
- Entities: Project, Template, BuildResult, UpdateInfo, Keystore
- Value Objects: Version (SEMVER), ProjectName, PackageName
- Repository Interfaces: ProjectRepository, TemplateRepository, BuildRepository, UpdateRepository

### Phase 2: Usecase
- CreateProjectUseCase: Interactive project creation wizard
- BuildProjectUseCase: Full APK build pipeline + incremental assets build
- SetupEnvironmentUseCase: SDK download + debug keystore generation
- DevModeUseCase: HMR with WebSocket, file watcher, ADB bridge
- UpdateUseCase: GitHub release checker + auto-updater

### Phase 3: Repository (Infrastructure)
- FilesystemProjectRepo: Read/write tamk.config, project structure
- TemplateRepo: Load .tmpl files, apply replacements
- BuildRepo: Build cache (SHA-256 hash comparison)
- UpdateRepo: GitHub API client with caching

### Phase 4: Delivery
- CLI (cobra): All tamk commands (--create, --build, --dev, --setup, --run, --install, --update, --version)
- HTTP: HMR dev server (WebSocket + HTTP fallback)

### Phase 5: Tests + Infrastructure
- Unit tests for domain + usecase (testify + gomock)
- Integration tests for repository
- Makefile, Dockerfile, golangci-lint config

## Key Risk Areas
1. **HMR System** (790 lines Python): Complex asyncio WebSocket + file watcher + ADB bridge
2. **Build Pipeline** (571 lines Python): 11-step APK build with AAPT2, kotlinc, D8, zipalign, apksigner
3. **Template Engine**: 19 .tmpl files with 10 placeholders
4. **External Commands**: All Android SDK tools must be shell-executed
5. **Termux-Specific Features**: termux-share, Termux environment detection

## Architecture Decisions (ADRs)
- **ADR-001**: Clean Architecture v4 with `internal` package enforcement
- **ADR-002**: Cobra CLI framework for command routing
- **ADR-003**: slog for structured logging
- **ADR-004**: fsnotify for file watching (instead of watchdog)
- **ADR-005**: gorilla/websocket for HMR WebSocket server
- **ADR-006**: Viper for configuration management
