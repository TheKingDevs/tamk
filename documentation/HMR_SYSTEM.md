# Hot Module Replacement (HMR) — Sistema de Desenvolvimento em Tempo Real

**Versao:** 2026.3.0-HMR
**Status:** Em desenvolvimento
**Aplica-se a:** WebApps (HTML/CSS/JS)

---

## Visao Geral

O HMR do T.A.M.K permite iteracao rapida durante o desenvolvimento de WebApps, detectando mudancas em arquivos e reconstruindo assets sem rebuild completo do APK.

### Arquitetura

```
Terminal (Termux)
    | tamk dev
    v
DevModeUseCase (internal/usecase/development.go)
├── FileWatcher (fsnotify) — Monitora src/main/assets/
├── Injector — Injeta bridge JS no index.html
├── Incremental Builder (AssetsOnlyBuild) — Reempacota APK
└── ADB Bridge — Instalacao automatica
    |
    v
WebView (App)
└── Bridge JS (WebSocket client) — Conecta em ws://localhost:8765
```

### Fluxo

1. `tamk dev` valida que o projeto e WebApp
2. Faz backup de `index.html` para `index.html.tamk_backup`
3. Injeta bridge JS (WebSocket client) antes de `</body>`
4. Inicia FileWatcher em `src/main/assets/`
5. Em mudanca:
   - `.css` / `.js` — apenas log (sem acao automatica)
   - Outros arquivos — `AssetsOnlyBuild()` (reempacota + assina APK)

### Bridge JS (inline, sem servidor WebSocket)

```javascript
(function() {
    var ws = new WebSocket('ws://localhost:8765');
    ws.onmessage = function(e) {
        var msg = JSON.parse(e.data);
        if (msg.type === 'reload') { location.reload(); }
        else if (msg.type === 'css-update') {
            document.querySelectorAll('link[rel="stylesheet"]').forEach(function(link) {
                link.href = link.href.split('?')[0] + '?t=' + Date.now();
            });
        }
    };
    ws.onopen = function() {
        ws.send(JSON.stringify({type: 'hello', url: window.location.href}));
    };
})();
```

**Nota:** O bridge e apenas cliente-side. Nao ha servidor WebSocket implementado — `ws://localhost:8765` e预备 para implementacao futura.

### Como Usar

```bash
cd MeuWebApp
tamk dev
```

Comandos planejados no modo dev (ainda NAO implementados no codigo Go):
- `b` — Rebuild forçado
- `i` — Instalar APK via ADB
- `s` — Status
- `h` — Ajuda
- `q` — Encerrar

> **Nota:** Estes comandos sao apenas planejados. O modo dev atual inicia o FileWatcher e executa `AssetsOnlyBuild()` automaticamente; nao ha REPL interativo.

### Diferencas do Planejado

| Funcionalidade | Status |
| :--- | :--- |
| Servidor WebSocket (:8765) | Nao implementado (cliente-side apenas) |
| Hot Module Swap (JS/CSS sem reload) | Nao implementado |
| Module Registry | Nao implementado |
| `--ws-port` / `--no-ws` | Flags definidas mas nao implementadas |
| `TAMK_HMR.*` API | Nao implementada |

### AssetsOnlyBuild

Reconstrucao incremental de assets:
1. Extrai APK existente (`unzip`)
2. Remove pasta `assets/` antiga
3. Copia `src/main/assets/` atual
4. Reempacota (`zip`)
5. Assina com `apksigner`
6. Gera `{name}-{version}-dev.apk`

### Parada (Ctrl+C)

1. FileWatcher parado
2. `index.html` restaurado do backup (`.tamk_backup` removido)

---
