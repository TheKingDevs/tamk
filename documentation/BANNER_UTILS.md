# BANNER UTILITIES

Nao existe um pacote separado de Banner. Banners sao construidos diretamente na camada CLI (`internal/delivery/cli/wizard.go`) com codigo inline usando constantes ANSI, `tput cols` para deteccao de largura do terminal, e caracteres Unicode de borda.

## Logger (`pkg/logger/logger.go`)

Logger baseado em `log/slog` com niveis customizados e cores ANSI.

### Niveis

- `LevelDebug(-4)` — cinza
- `LevelInfo(0)` — azul
- `LevelStep(2)` — ciano
- `LevelSuccess(4)` — verde
- `LevelWarn(4)` — amarelo (reusa nivel slog)
- `LevelError(8)` — vermelho

### Inicializacao

```go
logger.Init(verbose bool) // verbose = true ativa LevelDebug
```

### Funcoes

```go
logger.Debug(msg string, args ...any)
logger.Info(msg string, args ...any)
logger.Step(msg string, args ...any)
logger.Success(msg string, args ...any)
logger.Warn(msg string, args ...any)
logger.Error(msg string, args ...any)
```

### Constantes ANSI

```go
ansiReset, ansiCyan, ansiGreen, ansiYellow, ansiRed, ansiGray, ansiBlue
```

## Banner no CLI (`internal/delivery/cli/wizard.go`)

Funcoes auxiliares para o assistente interativo:

- `termWidth()` — executa `tput cols` para largura do terminal
- `center(text, width)` — centraliza texto (ignora codigos ANSI)
- `separator(char, color, width)` — linha repetindo caractere
- `showBanner()` — limpa tela, exibe arte ASCII (via `toilet`) ou fallback com bordas

```go
showSuccessBanner(name, projType, projPath, pkg string)
```

Nao existem classes `Banner`, `DynamicBanner` ou enum `BannerStyle`. A arte do logo e carregada de `assets/images/logo.png` via `toilet`.

---
