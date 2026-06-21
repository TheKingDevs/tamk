# ❓ Perguntas Frequentes (FAQ)

---

## Geral

### O que é o T.A.M.K?

Framework de automação para desenvolver, compilar e assinar APKs Android diretamente no Termux. Suporta UI nativa (XML/Kotlin), Console (Kotlin CLI) e WebApps (WebView HTML/CSS/JS).

### Quais os requisitos?

Android com Termux, Go 1.26+, OpenJDK 21, Kotlin, aapt2, apksigner, zip, wget. O `setup-install.sh` automatiza tudo.

### Funciona no SmartIDE?

Sim. O instalador detecta SmartIDE e ajusta paths automaticamente.

### Preciso de root?

Não. O T.A.M.K funciona sem root. A instalação de APK usa um servidor HTTP local com QR code para download direto.

---

## WebApps

### O que é um WebApp no T.A.M.K?

App Android que usa WebView para renderizar HTML/CSS/JS dentro de um APK nativo.

### Qual a diferença para um site responsivo?

WebApp é instalado, tem ícone na tela inicial, pode funcionar offline (assets internos) e pode acessar recursos nativos via JavaScriptInterface.

### Posso usar React, Vue ou Angular?

Sim. Compile para estático (`npm run build`) e copie a pasta `dist/` para `src/main/assets/`.

### Acessa câmera, GPS?

Por padrão não. Adicione permissões no `AndroidManifest.xml` e implemente `@JavascriptInterface` no `MainActivity.kt`.

### Funciona offline?

Sim, se o conteúdo estiver em `src/main/assets/`. Para APIs externas, implemente Service Workers ou cache local.

### PWA?

Sim — encapsule um PWA no WebView. Mas ele não será "progressivo" no sentido de instalação via navegador.

---

## Build e Deploy

### O que é uma Keystore?

Arquivo com chave criptográfica para assinar o APK. Sem assinatura, o Android não instala. Perdeu a keystore? Não pode atualizar o app na Play Store.

### Publicar na Google Play?

Sim. Use keystore de release, crie conta no Google Play Console, siga as diretrizes.

### Como atualizar um app publicado?

Incremente `versionCode`/`versionName` no `AndroidManifest.xml`, faça o build, assine com a **mesma keystore**.

### CI/CD?

Tecnicamente possível. Requer ambiente com Java, Kotlin, aapt2, apksigner. Não oficialmente suportado.

---

## Troubleshooting

### Build falha com "Senha incorreta"?

Use a senha definida na criação: `tamk build -p SUA_SENHA`. Perdeu? Não há recuperação.

### WebApp exibe tela em branco?

1. `webView.loadUrl()` aponta para `file:///android_asset/index.html`?
2. `javaScriptEnabled = true`?
3. Habilite `chrome://inspect` para ver erros no console.

### `tamk` não reconhecido?

Reinicie o Termux ou `source ~/.bashrc`. Verifique se `$PREFIX/bin/tamk` existe e tem permissão de execução.

### Build pula ("Nada mudou")?

Remova `.build_cache` para forçar rebuild.

### Modo dev não inicia?

```bash
tamk build -p senha  # Precisa de APK base primeiro
```

### Erro "aapt2 not found"?

```bash
pkg install aapt2
```

### Erro "SDK not configured"?

```bash
tamk setup
```

### HMR não conecta?

```bash
tamk dev --no-ws  # Usa HTTP fallback
```

---

## Contribuição

### Como contribuir?

Consulte [CONTRIBUTING.md](CONTRIBUTING.md). Issues, PRs, novos templates e documentação são bem-vindos.

---

<div align="center">
  <sub>T.A.M.K v1.0.0 — FAQ</sub>
</div>
