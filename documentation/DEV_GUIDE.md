# ⚙️ Guia de Desenvolvimento

> Guia prático de fluxo de trabalho para desenvolver WebApps com T.A.M.K — da configuração ao deployment.

---

## 📋 Índice

- [1. Ambiente de Desenvolvimento](#1-ambiente-de-desenvolvimento)
- [2. Ciclo de Vida de um WebApp](#2-ciclo-de-vida-de-um-webapp)
- [3. Modo Desenvolvimento com HMR](#3-modo-desenvolvimento-com-hmr)
- [4. Debugging](#4-debugging)
- [5. Customização Nativa](#5-customização-nativa)
- [6. Build e Deploy](#6-build-e-deploy)
- [7. CI/CD e Automação](#7-cicd-e-automação)

---

## 1. Ambiente de Desenvolvimento

### Instalação

```bash
bash setup-install.sh
tamk version
```

### Verificação do Ambiente

```bash
tamk setup  # Configura SDK e keystore de debug
```

### Estrutura de um Projeto WebApp

```
MeuWebApp/
├── src/main/assets/     ← Seu conteúdo web
│   ├── index.html
│   ├── css/
│   ├── js/
│   └── images/
├── AndroidManifest.xml  ← Permissões e configurações
├── tamk.config          ← Metadados do projeto
└── secret/              ← Keystore (NÃO versionar)
```

---

## 2. Ciclo de Vida de um WebApp

### Fase 1: Criação

```bash
tamk create
# → Escolha WebApp
# → Nome, versão, autor, keystore
```

### Fase 2: Desenvolvimento

Edite arquivos em `src/main/assets/`:

```bash
cd MeuWebApp
nano src/main/assets/index.html
nano src/main/assets/css/styles.css
nano src/main/assets/js/app.js
```

**Dica:** Frameworks modernos (React, Vue, Angular) podem ser compilados para estático:

```bash
npm run build
cp -r dist/* MeuWebApp/src/main/assets/
```

### Fase 3: Modo Dev (HMR)

```bash
tamk dev
```

### Fase 4: Build de Produção

```bash
tamk build -p SUA_SENHA
```

### Fase 5: Instalação

```bash
tamk install
```

---

## 3. Modo Desenvolvimento com HMR

### Iniciar

```bash
cd MeuWebApp
tamk dev
```

### Opções

```bash
tamk dev --verbose       # Logs detalhados
```

### Comportamento por Tipo de Arquivo

| Arquivo | Ação | Tempo | Estado |
| :--- | :--- | :--- | :--- |
| `styles.css` | Logs only (HMR-ready) | < 50ms | ✅ Preservado |
| `app.js` | Logs only (HMR-ready) | < 100ms | ✅ Preservado |
| `data.json` | Logs only (HMR-ready) | < 50ms | ✅ Preservado |
| `index.html` | Rebuild + reload | ~2-5s | ⚠️ Parcial |
| `logo.png` | Rebuild + reload | ~2-5s | ⚠️ Parcial |

*(HMR bridge injects a WebSocket client inline for future hot-reload support; currently logs changes only)*

---

## 4. Debugging

### Console Remoto (Chrome DevTools)

Habilite no `MainActivity.kt`:

```kotlin
WebView.setWebContentsDebuggingEnabled(true)  // Adicione no onCreate()
```

Conecte via USB → `chrome://inspect`

### JavaScript Alerts

```javascript
alert('Valor: ' + minhaVariavel);
```

O `WebChromeClient` configurado exibe alerts nativamente.

### Logs no Logcat

```kotlin
Log.d(TAG, "Mensagem de debug")
```

Visualize com:
```bash
adb logcat -s MainActivity
```

### Status do HMR

Status é exibido no log durante execução do `tamk dev`:

```
Build status: monitoring src/main/assets/ (5 builds completed)
```

---

## 5. Customização Nativa

### Modificar o WebView

Edite `src/main/kotlin/com/.../MainActivity.kt`:

```kotlin
webView.settings.apply {
    javaScriptEnabled = true
    domStorageEnabled = true
    allowFileAccess = true
    allowContentAccess = false        // Segurança
    cacheMode = WebSettings.LOAD_DEFAULT
    setSupportZoom(true)
    builtInZoomControls = true
    displayZoomControls = false
}
```

### Adicionar Interface Nativa-JS

```kotlin
class WebAppInterface(private val context: Context) {
    @JavascriptInterface
    fun getDeviceInfo(): String {
        return "Android ${Build.VERSION.RELEASE}"
    }
}

webView.addJavascriptInterface(WebAppInterface(this), "Android")
```

No JavaScript:
```javascript
console.log(Android.getDeviceInfo());
```

### BroadcastReceiver para Refresh

O template já inclui `BroadcastReceiver` para `tamk.ACTION_REFRESH_ASSET` e `tamk.ACTION_REFRESH_ALL`. Use:

```bash
adb shell am broadcast -a tamk.ACTION_REFRESH_ALL
```

---

## 6. Build e Deploy

### Build Completo

```bash
tamk build -p SUA_SENHA
```

Flags:
| Flag | Efeito |
| :--- | :--- |
| `-p SENHA` | Fornece senha (evita prompt) |
| `-V` | Modo verbose com logs de cada etapa |

### Cache Inteligente

O build é pulado se nada mudou:

```
✨ Nada mudou desde o último build. APK atualizado!
```

Force rebuild limpando o cache:
```bash
rm .build_cache
```

### APK Gerado

| Arquivo | Propósito |
| :--- | :--- |
| `{name}-{version}-release.apk` | APK assinado e alinhado (produção) |
| `{name}-{version}-dev.apk` | APK de desenvolvimento (dev mode) |

### Instalação

```bash
# Via T.A.M.K
tamk install

# Via ADB (se configurado)
adb install -r {name}-{version}-release.apk

# Manual
cp {name}-{version}-release.apk /sdcard/Download/
```

---

## 7. CI/CD e Automação

### Build Automatizado

```bash
tamk build -p $KEYSTORE_PASS
```

### Pipeline GitHub Actions (exemplo)

```yaml
- name: Build APK
  run: |
    cd projeto
    tamk build -p ${{ secrets.KEYSTORE_PASS }}
```

### Requisitos para CI

- Go 1.26+, OpenJDK 21, Kotlin
- Android SDK tools (aapt2, apksigner, d8, zipalign)

---

## 🛡️ Boas Práticas

1. **Keystore**: Uma por projeto, guarde em cofre seguro
2. **Versionamento**: Incremente `tamk.config` antes de cada build
3. **Cache**: Limpe `.build_cache` se houver erros estranhos
4. **Backup**: `index.html.tamk_backup` é criado no modo dev
5. **Teste**: Teste em dispositivo real antes de publicar
6. **Segurança**: Desabilite `allowFileAccess` e `allowContentAccess` em produção se não necessário

---

<div align="center">
  <sub>T.A.M.K v1.0.0 — Guia de Desenvolvimento</sub>
</div>
