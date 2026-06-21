# T.A.M.K — Termux APK Manager Kit

T.A.M.K (Termux APK Manager Kit) v1.0.0 — framework de automacao para desenvolvimento Android nativo diretamente no Termux. Escrito em Go com Clean Architecture v4. Permite criar, compilar, assinar e instalar APKs Android sem um PC.

<p align="center">
  <img src="assets/images/logo.png" alt="T.A.M.K Logo" width="300">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Version-1.0.0-blueviolet?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/Go-1.26-blue?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/TheKingDevs/tamk/ci.yml?style=for-the-badge&label=CI" alt="CI">
  <br>
  <img src="https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black" alt="Linux">
  <img src="https://img.shields.io/badge/Termux-000?style=for-the-badge&logo=terminal&logoColor=white" alt="Termux">
  <img src="https://img.shields.io/badge/macOS-000?style=for-the-badge&logo=apple" alt="macOS">
  <img src="https://img.shields.io/badge/Windows-0078D6?style=for-the-badge&logo=windows" alt="Windows">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
</p>

---

## Recursos

| Funcionalidade | Descricao |
| :--- | :--- |
| **3 Tipos de Projeto** | UI APK nativo, Console Kotlin, WebApp hibrido (WebView) |
| **WebApp com URL Remota** | Suporte a carregamento de URLs externas no WebView |
| **Hot Module Replacement** | Bridge JS inline + file watcher + assets rebuild |
| **Build Incremental** | Atualizacao rapida de assets sem rebuild completo |
| **Assinatura Segura** | Keystore privada por projeto + validacao de senha inline |
| **Cache Inteligente** | Build pulado se nada mudou (hash SHA-256) |
| **ADB Integration** | Push de assets e broadcast via ADB |
| **Template Engine** | Arquivos `.tmpl` com placeholders — customizacao sem alterar o nucleo |
| **Sistema de Updates** | Verificacao via GitHub API com niveis de prioridade |
| **Instalacao via QR Code** | Servidor HTTP + QR code no terminal (`tamk install`) |
| **Shell Interativo** | REPL com comandos create, build, dev, setup, install, update |

---

## Instalacao

### Pre-requisitos

```bash
pkg update && pkg upgrade
pkg install -y golang openjdk-21 kotlin wget zip apksigner aapt2 git
```

### Instalacao Rapida

```bash
git clone https://github.com/TheKingDevs/tamk.git
cd tamk
make build
cp bin/tamk $PREFIX/bin/
```

### Via Script

```bash
bash setup-install.sh
```

### Verificacao

```bash
tamk version
```

---

## Comandos

| Comando | Descricao |
| :--- | :--- |
| `tamk create` | Assistente interativo de criacao de projeto |
| `tamk build -p senha` | Compila e assina APK |
| `tamk dev` | Modo desenvolvimento com HMR |
| `tamk setup` | Download SDK + geracao debug keystore |
| `tamk install [porta]` | Servidor HTTP + QR code para instalar APK |
| `tamk update` | Verifica atualizacoes via GitHub API |
| `tamk version` | Exibe versao e ambiente |
| `tamk shell` | REPL interativo |

### Flags Globais

| Flag | Descricao |
| :--- | :--- |
| `-p, --password SENHA` | Senha da keystore |
| `-V, --verbose` | Logs detalhados (DEBUG) |

---

## Tipos de Projeto

### Fluxo de Criacao

```
tamk create
  1. Nome do projeto
  2. Author
  3. Versao (semver)
  4. Engine:
      [1] Console          → Kotlin CLI
      [2] UI/APK           → Android nativo XML/Kotlin
      [3] WebApp           → WebView hibrido
          - Interno (assets/)
          - Externo (URL remota)
  5. Senha da Keystore (se aplicavel)
```

---

## Pipeline de Build

### Completo (`tamk build`)

1. **Validacao** — SDK existe, senha valida (min 6 chars)
2. **Hash Check** — SHA-256 dos fontes vs `.build_cache`
3. **AAPT2 Compile** — `aapt2 compile --dir res -o res.zip`
4. **AAPT2 Link** — `aapt2 link -I android.jar --manifest AndroidManifest.xml -o app.apk`
5. **Kotlin Compile** — `kotlinc src gen/ -cp android.jar -d obj/`
6. **D8 DEX** — `d8 --lib android.jar --release --output . *.class`
7. **Zipalign** — `zipalign -f 4 app.apk app-unsigned.apk`
8. **Assinatura** — `apksigner sign --ks keystore --out app-final.apk`
9. **Cleanup** — Remove temporarios, salva `.build_cache`

### Incremental (Dev Mode)

```
AssetsOnlyBuild():
1. Extrai APK existente
2. Remove pasta assets/ antiga
3. Copia src/main/assets/ atual
4. Reempacota + assina → app-dev.apk
```

---

## Modo Dev & HMR

```bash
cd MeuWebApp
tamk dev
```

### Fluxo

```
File Watcher (fsnotify) → Assets Build → ADB Install → WebView
  1. Edita arquivo em assets/
  2. fsnotify detecta mudanca (500ms debounce)
  3. AssetsOnlyBuild() reconstroi APK
  4. ADB instala APK atualizado
  5. Broadcast ACTION_REFRESH para WebView recarregar
```

### Extensoes Monitoradas

`.html`, `.css`, `.js`, `.json`, `.png`, `.jpg`, `.svg`, `.webp`, `.xml`, `.kt`

### Bridge JS Inline

Um WebSocket client e injetado no `index.html` para conectar em `ws://localhost:8765`:

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
})();
```

---

## Sistema de Atualizacoes

| Nivel | Exemplos | Comportamento |
| :--- | :--- | :--- |
| **CRITICAL** | Seguranca, bugs graves | Instalacao automatica |
| **PATCH** | Correcoes menores | Instalacao automatica |
| **MAJOR/MINOR** | Novas funcionalidades | Pergunta ao usuario |
| **OPTIONAL** | Docs, cosmeticos | Notifica apenas |

### Metodos

1. **git** — `git pull --rebase --autostash` em TAMK_HOME
2. **go install** — `go install github.com/TheKingDevs/tamk/cmd/tamk@latest`

---

## Estrutura do Projeto

### WebApp (Interno)

```
MeuWebApp/
├── AndroidManifest.xml
├── tamk.config
├── .gitignore
├── secret/
│   └── project.keystore
├── res/
│   ├── values/strings.xml
│   ├── values/styles.xml
│   ├── drawable/ic_launcher.xml
│   └── xml/network_security_config.xml
├── src/main/
│   ├── assets/
│   │   ├── index.html
│   │   ├── css/styles.css
│   │   ├── js/app.js
│   │   └── .gitignore
│   └── kotlin/com/author/name/
│       └── MainActivity.kt
├── development/
│   ├── sdk/android.jar
│   └── secret/debug.keystore
└── app-final.apk
```

---

## Arquitetura

```
CLI (Cobra) → Use Cases → Domain Entities → Templates → Projeto Gerado
                         ↓
                    Build Pipeline → APK Assinado
```

### Camadas (Clean Architecture v4)

| Camada | Diretorio | Responsabilidade |
| :--- | :--- | :--- |
| **Delivery** | `internal/delivery/cli/` | Cobra commands, wizard, REPL |
| **Use Case** | `internal/usecase/` | Logica de negocio (create, build, dev, etc.) |
| **Domain** | `internal/domain/` | Entities, value objects, repository interfaces |
| **Repository** | `internal/repository/` | Filesystem I/O, GitHub API |
| **Config** | `internal/config/` | Configuracao central |
| **Shared** | `pkg/` | Logger, errors, watcher, qrcode |

---

## Documentacao

| Documento | Descricao |
| :--- | :--- |
| [AGENTS.md](AGENTS.md) | Instrucoes completas para agentes de IA |
| [documentation/ARCHITECTURE.md](documentation/ARCHITECTURE.md) | Arquitetura do sistema |
| [documentation/QUICKSTART.md](documentation/QUICKSTART.md) | Guia rapido |
| [documentation/DEV_GUIDE.md](documentation/DEV_GUIDE.md) | Guia de desenvolvimento |
| [documentation/HMR_SYSTEM.md](documentation/HMR_SYSTEM.md) | Documentacao HMR |
| [documentation/UPDATE_SYSTEM.md](documentation/UPDATE_SYSTEM.md) | Sistema de atualizacoes |
| [PRD/workflow.md](PRD/workflow.md) | Product Requirements Document |

---

<div align="center">
  <sub>T.A.M.K v1.0.0 — Feito com ❤️ por @mrx_dev</sub>
</div>
