# Guia de Contribuicao

Agradecemos seu interesse em contribuir com o **T.A.M.K**!

---

## Arquitetura

O projeto segue **Clean Architecture v4** com dependencia rigorosa: camadas internas nunca importam camadas externas.

```
cmd/tamk/main.go (entrada)
    |
delivery/cli/ (comandos Cobra)
    |
usecase/ (orquestracao de negocios)
    |
domain/entity/ + domain/valueobject/ + domain/repository/ (contratos)
    |
repository/filesystem/ (implementacao concreta)
    |
Sistema de Arquivos
```

### Componentes Chave

| Componente | Funcao |
| :--- | :--- |
| **CLI** (`internal/delivery/cli/root.go`) | Comandos Cobra, parsing de flags, injecao de dependencias |
| **Use Cases** (`internal/usecase/`) | Logica de negocio: `CreateProjectUseCase`, `BuildProjectUseCase`, `DevModeUseCase`, etc. |
| **Template Repository** (`internal/repository/filesystem/template_repository.go`) | Carregamento de `.tmpl` e substituicao de placeholders com `strings.Replacer` (`{{PLACEHOLDER}}`) |

| **Build Repository** (`internal/repository/filesystem/build_repository.go`) | Pipeline de build: `aapt2`, `kotlinc`, `d8`, `zipalign`, `apksigner` via `exec.Command()` |
---

## Areas de Contribuicao

### 1. Templates (Prioridade Alta)

Novos layouts em `assets/templates/`:
- `login_activity.xml.tmpl` + `LoginActivity.kt.tmpl`
- `settings_activity.xml.tmpl` com PreferenceFragment
- `compose_activity.tmpl` (Jetpack Compose)

### 2. Use Cases

- `BuildProjectUseCase`: Compilacao incremental, cache SHA-256, build paralelo
- `SetupEnvironmentUseCase`: Download paralelo, checksum SHA-256, mirrors

### 3. Utilitarios

- `pkg/logger/`: Niveis de log customizados (Debug, Info, Step, Success, Warn, Error)
- `pkg/watcher/`: File watcher com fsnotify e debounce

### 4. Testes

```bash
go test ./...
```

- Unitarios para Template Repository (`ApplyPlaceholders`)
- Integracao para Build Pipeline
- Benchmarks de performance

---

## Regras de Estilo

### Go

```go
func (uc *BuildProjectUseCase) FullBuild(ctx context.Context, input BuildInput) (*entity.BuildResult, error) {
    // Compila o projeto APK
    return &entity.BuildResult{Success: true, APKPath: finalPath}, nil
}
```

### Kotlin e XML

- Kotlin: 4 espacos indentacao, nomes descritivos
- XML: IDs primeiro, layout, depois estilo — 4 espacos
- Placeholders: **MAIUSCULAS** com chaves duplas (`{{NAME}}`, `{{PACKAGE}}`)

---

## Pull Request

1. Fork e clone
2. Branch descritiva (`feature/`, `fix/`, `docs/`)
3. Commits atomicos com mensagens claras
4. Rebase com `upstream/main`
5. Push e PR detalhando mudancas

---

## Reportando Bugs

- **Ambiente**: Dispositivo, Android, Termux, T.A.M.K
- **Comando**: Exato com `-V` (verbose)
- **Comportamento**: Esperado vs real
- **Logs**: Saida completa

---
