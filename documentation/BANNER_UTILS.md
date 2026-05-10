# BANNER UTILITIES

Utilitário para exibição dinâmica de banners no terminal, adaptando-se automaticamente ao tamanho da tela.

## Localização

```
src/utils/banner.py
```

## Importação

```python
from utils.banner import (
    Banner,           # Classe completa
    DynamicBanner,    # Banner com redraw automático no resize
    banner,           # Função simplificada
    banner_art,       # Banner com arte ASCII
    banner_box,       # Caixa com bordas
    success_banner,   # Banner de sucesso
    error_banner,     # Banner de erro
    warning_banner,   # Banner de aviso
    info_banner,      # Banner informativo
    header_art,       # Header com arte
    success_art,      # Arte de sucesso
)
```

## Estilos Disponíveis

- `default` - Padrão (ciano)
- `success` - Sucesso (verde)
- `error` - Erro (vermelho)
- `warning` - Aviso (amarelo)
- `info` - Informativo (azul)
- `header` - Cabeçalho (magenta)

## Exemplos de Uso

### Banner Simples

```python
from utils.banner import banner

# Banner padrão
banner("TÍTULO", "Subtítulo opcional")

# Banner com estilo
banner("SUCESSO", "Operação concluída", style="success")
banner("ERRO", "Falha na operação", style="error")
banner("AVISO", "Atenção necessária", style="warning")
```

### Banner Dinâmico (Redimensionamento Automático)

```python
from utils.banner import DynamicBanner, BannerStyle

# Banner que se redesenha ao redimensionar o terminal
dyn_banner = DynamicBanner(
    title="T.A.M.K",
    subtitle="Termux Apk Manager Kit",
    style=BannerStyle.HEADER,
    separator_top='═',      # Separador superior
    separator_bottom='═'    # Separador inferior
)

# Exibe com redraw automático (Ctrl+C para sair)
dyn_banner.show_continuous(
    top_sep='═',
    bottom_sep='═'
)

# Ou exibe apenas uma vez
dyn_banner.show_custom(top_sep='─', bottom_sep='─')
```

### Banner com Conteúdo Dinâmico e Redraw Automático

```python
from utils.banner import DynamicBanner

dyn_banner = DynamicBanner("T.A.M.K", "Servidor Rodando")

def render_status():
    '''Renderiza conteúdo adicional.'''
    return [
        f"Clientes: {client_count}",
        f"Uptime: {uptime}s",
        f"Largura: {terminal_width}cols"
    ]

# Banner + conteúdo que se redesenha no resize
dyn_banner.show_continuous(
    render_callback=render_status,
    top_sep='═',
    bottom_sep='═'
)
```

### Separadores Customizados

```python
from utils.banner import DynamicBanner

dyn = DynamicBanner(
    title="T.A.M.K",
    separator_top='─',      # Linha simples
    separator_bottom='─'
)

# Diferentes estilos de separadores
dyn.show_custom(top_sep='═', bottom_sep='═')  # Linha dupla
dyn.show_custom(top_sep='─', bottom_sep='─')  # Linha simples
dyn.show_custom(top_sep='=', bottom_sep='=')  # Igual
dyn.show_custom(top_sep='*', bottom_sep='*')  # Asteriscos
```

### Banner com Arte ASCII

```python
from utils.banner import banner_art, success_art, header_art

# Arte com toilet
banner_art("T.A.M.K", subtitle="Termux Apk Manager Kit")

# Arte de sucesso
success_art("PRONTO", subtitle="Projeto criado!")

# Header com arte
header_art("DEV MODE", subtitle="Servidor iniciado")
```

### Caixa com Bordas

```python
from utils.banner import banner_box

banner_box(
    lines=[
        "Linha 1 de conteúdo",
        "Linha 2 de conteúdo",
        "Linha 3 de conteúdo",
    ],
    style="info",
    title="Título da Caixa"
)
```

### Classe Banner (Controle Total)

```python
from utils.banner import Banner, BannerStyle

b = Banner()

# Banner completo
b.show(
    title="TÍTULO",
    subtitle="Subtítulo",
    lines=["Linha 1", "Linha 2"],
    style=BannerStyle.DEFAULT,
    separator='=',
    show_top=True,
    show_bottom=True
)

# Arte ASCII
b.show_art("TEXTO", font="standard", effects="metal")

# Banner com arte
b.show_art_banner(
    art_text="T.A.M.K",
    subtitle="Kit de Desenvolvimento",
    style=BannerStyle.HEADER
)

# Caixa com bordas
b.show_box(
    lines=["Conteúdo 1", "Conteúdo 2"],
    style=BannerStyle.SUCCESS,
    title="Minha Caixa"
)
```

## Casos de Uso no Projeto

### Project Manager (Criação de Projetos)

```python
from utils.banner import banner_art, success_art

# Banner inicial
banner_art("T.A.M.K", subtitle="Termux Apk Manager Kit • Factory (2026)")

# Banner de sucesso
success_art("PRONTO", subtitle=f"PROJETO '{name}' CRIADO!")
```

### Dev Controller (Instruções)

```python
from utils.banner import banner_box

banner_box(
    lines=[
        "✨ RECURSOS HMR:",
        "  CSS → Atualiza sem reload",
        "  JS → Injeta módulos",
        "",
        "⌨️ COMANDOS:",
        "  Ctrl+C → Encerrar | b → Rebuild",
    ],
    style="info",
    title="🚀 MODO DESENVOLVIMENTO"
)
```

### Run Controller (Header de Saída)

```python
from utils.banner import Banner

banner = Banner()
print(banner._create_separator('=', TColor.CYAN))
print(banner._center_text(f"{TColor.BOLD}SAÍDA DO PROGRAMA{TColor.RESET}"))
print(banner._create_separator('=', TColor.CYAN))
```

## Vantagens

1. **Adaptável**: Sempre se ajusta à largura do terminal
2. **Limpo**: Código mais legível e manutenível
3. **Consistente**: Padronização visual em todo o projeto
4. **Flexível**: Múltiplos estilos e formatos
5. **Fallback**: Funciona mesmo sem `toilet` instalado

## Dependências

- `toilet` (opcional): Para arte ASCII estilizada
- `tput`: Para detectar largura do terminal (já incluso na maioria dos terminais)

## Fallback

Se o `toilet` não estiver disponível, o utilitário usa um fallback simples:
```
=== TEXTO ===
```
