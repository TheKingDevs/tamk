# Performance Requirements Document — T.A.M.K

## Visão Geral
Define requisitos de performance, métricas alvo, e otimizações para T.A.M.K v1.0.0.

---

## Métricas Atuais (Baseline)

| Métrica | Valor Atual | Notas |
|:--|:--|:--|
| Build do binário TAMK | ~9s | `go build` em ARM64 |
| Tamanho do binário | 35MB | Inclui android.jar (21MB) via go:embed |
| CalculateHash | 1.1ms/op | SHA-256 de arquivos |
| StreamingHash | 3.4ms/op | SHA-256 streaming |
| Memory (StreamingHash) | 99KB/op | Buffer de leitura |
| Testes unitários | ~20s | Suite completa |

---

## Requisitos de Performance

### 1. Build Pipeline (Full Build)
| Métrica | Alvo | Atual | Status |
|:--|:--|:--|:--|
| Full build (clean) | < 120s | ~90s | ✅ |
| Incremental (assets) | < 10s | ~5s | ✅ |
| Cache hit (nada mudou) | < 1s | ~0.5s | ✅ |
| Keystore validation | < 2s | ~1s | ✅ |

### 2. Startup & CLI
| Métrica | Alvo | Notas |
|:--|:--|:--|
| Tempo de startup | < 100ms | Go binário nativo |
| `tamk version` | < 50ms | Sem I/O |
| `tamk create` (wizard) | < 5s | Interativo |
| `tamk setup` (first run) | < 30s | Extração de ferramentas |
| `tamk setup` (subsequent) | < 1s | Version check only |

### 3. ToolManager
| Métrica | Alvo | Notas |
|:--|:--|:--|
| Tool resolution | < 10ms | Cached after first call |
| Setup (first time) | < 30s | Extrai ~103MB (Windows) |
| Setup (idempotent) | < 100ms | Version check |
| Memory footprint | < 50MB | Ferramentas em disco |

### 4. Build Cache
| Métrica | Alvo | Notas |
|:--|:--|:--|
| Hash calculation | < 5ms | SHA-256 de fontes |
| Cache lookup | < 1ms | Leitura de .build_cache |
| Cache validation | < 10ms | Compara hash |

### 5. File Watcher (Dev Mode)
| Métrica | Alvo | Notas |
|:--|:--|:--|
| Debounce interval | 500ms | Configurável |
| File change detection | < 100ms | fsnotify |
| Assets rebuild | < 10s | Incremental |
| Memory per watcher | < 10MB | fsnotify overhead |

---

## Oportunidades de Otimização

### Prioridade Alta

#### 1. ~~Binary Size Reduction (35MB → ~15MB)~~ (Adiado)
**Problema:** Binário inclui android.jar (21MB) embutido.
**Status:** Adiado — prioridade menor para lancamento.

#### 2. ~~Build Pipeline Paralelismo~~ (Nao aplicavel)
**Problema:** Steps do build são sequenciais.
**Status:** Análise mostrou dependências de dados reais entre AAPT2→link→Kotlin→D8. Paralelismo não seguro.

#### 3. Template Loading Cache ✅ IMPLEMENTADO
**Problema:** Templates são lidos do disco a cada criação.
**Solução:** Cache `sync.Map` em `TemplateRepository` — cada template lido uma vez, servido da memória.
**Arquivo:** `internal/repository/filesystem/template_repository.go`

### Prioridade Média

#### 4. Hash Calculation — Skip Ignored Dirs ✅ IMPLEMENTADO
**Problema:** `collectSourceFiles()` não ignorava `.git`, `node_modules`, `secret`, etc.
**Solução:** Mapa `ignoredDirs` com `filepath.SkipDir` — previne hash de milhares de arquivos irrelevantes.
**Arquivo:** `internal/repository/filesystem/build_repository.go`
**Impacto:** 10-100x em projetos com `node_modules` em assets.

#### 5. ~~Tool Path Caching~~ (Já implementado)
**Status:** `cachedTool` com `sync.RWMutex` já funciona. SDKJar agora também cacheado.

#### 6. ~~Keystore Validation Async~~ (Adiado)
**Status:** Validação síncrona é aceitável (~1s). Async adiciona complexidade sem ganho significativo.

### Prioridade Baixa

#### 7. Config Singleton ✅ IMPLEMENTADO
**Problema:** `config.New()` criava nova instância a cada chamada.
**Solução:** `config.Get()` com `sync.Once` — lê `/etc/os-release` apenas uma vez.
**Arquivo:** `internal/config/config.go`

---

## Benchmarking Strategy

### Ferramentas
- `go test -bench=.` — Benchmarks Go
- `go tool pprof` — CPU profiling
- `go tool trace` — Execution tracing
- `benchstat` — Comparação de benchmarks

### Comandos Úteis
```bash
# Run benchmarks
make bench

# CPU profile
go test -cpuprofile=cpu.out -bench=. ./internal/...
go tool pprof cpu.out

# Memory profile
go test -memprofile=mem.out -bench=. ./internal/...
go tool pprof mem.out

# Compare benchmarks
benchstat old.txt new.txt
```

### CI Integration
- `tests-tt.yml` já executa testes
- Adicionar `benchstat` comparison no CI
- Flag performance regression no pre-commit

---

## Performance Budget

| Componente | Budget | Atual | Status |
|:--|:--|:--|:--|
| Binary size | < 50MB | 35MB | ✅ |
| Build time (full) | < 120s | ~90s | ✅ |
| Build time (incremental) | < 10s | ~5s | ✅ |
| CLI startup | < 100ms | ~50ms | ✅ |
| Test suite | < 60s | ~20s | ✅ |
| Memory (idle) | < 50MB | ~20MB | ✅ |
| Memory (build) | < 200MB | ~100MB | ✅ |

---

## Roadmap de Performance

### Fase 1 (Completa)
- ✅ Build cache com SHA-256
- ✅ Incremental assets build
- ✅ ToolManager com cache
- ✅ Keystore validation antes do build
- ✅ Template content cache (sync.Map)
- ✅ Hash: skip ignored directories
- ✅ Config singleton (sync.Once)
- ✅ SDKJar path caching

### Fase 2 (Próxima)
- [ ] Benchmark CI integration (benchstat)
- [ ] Binary compression (upx) — opcional

### Fase 3 (Futuro)
- [ ] Download sob demanda de ferramentas
- [ ] Build distribuído

---

## Monitoring

### Métricas a Monitorar
1. **Build duration** — Tempo total do build
2. **Cache hit rate** — % de builds que pulam compilação
3. **Tool resolution time** — Tempo para encontrar ferramentas
4. **Memory usage** — Pico de memória durante build
5. **Binary size** — Tamanho do binário por release

### Logging
```go
logger.Step("Building...", "phase", "aapt2_compile")
// Output: [STEP] Building... phase=aapt2_compile
```

### Profiling Points
- `FullBuild()` — CPU + memory profile
- `AssetsOnlyBuild()` — I/O profile
- `ToolManager.Setup()` — Disk I/O profile
