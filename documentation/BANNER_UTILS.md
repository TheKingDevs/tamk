# BANNER UTILITIES

Banners são construídos diretamente na camada CLI (`internal/delivery/cli/wizard.go`) com código inline usando constantes ANSI, `golang.org/x/term.GetSize()` para detecção de largura do terminal, e caracteres Unicode de borda.

## Logger (`pkg/logger/logger.go`)

Logger baseado em `log/slog` com níveis customizados e cores ANSI.

### Níveis

- `LevelDebug(-4)` — cinza
- `LevelInfo(0)` — azul
- `LevelStep(2)` — ciano
- `LevelSuccess(4)` — verde
- `LevelWarn(-4)` — amarelo (reusa nível slog)
- `LevelError(8)` — vermelho

### Inicialização

```go
logger.Init(verbose bool) // verbose = true ativa LevelDebug
```

### Funções

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

Funções auxiliares para o assistente interativo:

- `termWidth()` — usa `golang.org/x/term.GetSize()` para largura do terminal (fallback para 80)
- `center(text, width)` — centraliza texto (ignora códigos ANSI)
- `separator(char, color, width)` — linha repetindo caractere
- `showBanner()` — limpa tela, exibe arte ASCII (via `toilet`) ou fallback com bordas
- `renderArt(text, width)` — tenta `toilet` com cadeia de fallback de fontes, retorna string vazia se todas falharem
- `tryToilet(text, font, width)` — executa `toilet` com uma fonte específica, retorna string vazia em erro

### Cadeia de Fontes `toilet`

```go
fonts := []string{"", "mono12", "bigmono12", "ascii12", "future", "smblock"}
```

A ordem tenta: fonte padrão (sem `-f`, mais portável) → `mono12` (melhor saída Unicode) → `bigmono12` → `ascii12` → `future` → `smblock`. A primeira que produz saída é usada. Se todas falham (`toilet` não instalado ou sem fontes), `showBanner()` exibe o fallback Go puro com caixa de bordas.

```go
showSuccessBanner(name, projType, projPath, pkg string)
```

Não existem classes `Banner`, `DynamicBanner` ou enum `BannerStyle`. A arte do logo é gerada via `toilet` com fallback para caracteres Unicode de borda.

## Mudanças da v1.0.0

| Antes | Depois |
| :--- | :--- |
| `tput cols` para largura do terminal | `golang.org/x/term.GetSize()` |
| `toilet -f standard -F metal` | `toilet` (fonte padrão) ou cadeia de fallback |
| Falha se `standard.tlf` não existia | Fallback para 5 fontes conhecidas + Go puro |

---

