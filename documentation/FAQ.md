# ❓ Perguntas Frequentes (FAQ)

> Este documento responde às dúvidas mais comuns sobre o T.A.M.K, com foco especial no suporte a WebApps.

## Geral

### O que é o T.A.M.K?

O **T.A.M.K (Termux APK Manager Kit)** é um framework de automação que permite desenvolver, compilar e assinar aplicativos Android (APKs) diretamente no ambiente Termux, sem a necessidade de um computador desktop ou Android Studio. Ele suporta aplicativos nativos com interface XML, aplicações de console em Kotlin e, desde a versão 2026.2.0, **WebApps** que encapsulam sites HTML/CSS/JS em um APK instalável.

### Quais são os requisitos para usar o T.A.M.K?

Você precisa de um dispositivo Android com o Termux instalado e os seguintes pacotes: `python`, `openjdk-21`, `kotlin`, `aapt2`, `apksigner`, `zip` e `wget`. O script de instalação `setup-install.sh` automatiza a instalação dessas dependências.

### O T.A.M.K funciona no SmartIDE?

Sim. O T.A.M.K foi projetado para funcionar tanto no Termux quanto no SmartIDE. O script de instalação detecta automaticamente o ambiente e ajusta os caminhos de instalação.

---

## WebApps

### O que é um WebApp no contexto do T.A.M.K?

Um **WebApp** é um aplicativo Android que utiliza um componente `WebView` para renderizar uma aplicação web (HTML, CSS, JavaScript) dentro de um APK nativo. Isso permite que desenvolvedores web distribuam seus sites como aplicativos instaláveis na Google Play Store ou através de instalação direta (sideloading).

### Qual é a diferença entre um WebApp e um site responsivo?

Um site responsivo é acessado através de um navegador e depende de uma conexão com a internet. Um WebApp, por outro lado, é um aplicativo instalado no dispositivo. Ele pode funcionar offline (se o conteúdo estiver empacotado no APK), ter um ícone na tela inicial, e potencialmente acessar recursos nativos do Android através de interfaces JavaScript customizadas.

### Posso usar frameworks como React, Vue ou Angular em um WebApp?

Sim. Desde que você compile seu projeto para arquivos estáticos (HTML, CSS, JS), você pode colocá-los na pasta `src/main/assets/` do seu projeto T.A.M.K. Frameworks modernos geralmente possuem um comando de build (como `npm run build`) que gera uma pasta `dist/` ou `build/` com os arquivos prontos para produção. Basta copiar o conteúdo dessa pasta para `assets/`.

### Meu WebApp pode acessar a câmera, GPS ou outros recursos do dispositivo?

Por padrão, o template básico do T.A.M.K não configura essas permissões. No entanto, você pode adicionar as permissões necessárias ao `AndroidManifest.xml` do seu projeto e implementar uma interface JavaScript no `MainActivity.kt` para expor essas funcionalidades ao seu código web. Isso requer conhecimento intermediário de desenvolvimento Android.

### Como faço para que meu WebApp funcione offline?

Se todo o conteúdo do seu site estiver na pasta `src/main/assets/`, o WebApp funcionará offline por padrão, pois o `WebView` carrega os arquivos diretamente do APK. Se seu site faz requisições a APIs externas, você precisará implementar uma estratégia de cache (usando Service Workers, por exemplo) ou armazenar os dados localmente usando `localStorage`.

### Posso usar o T.A.M.K para criar um Progressive Web App (PWA)?

Sim e não. Um PWA é, por definição, uma aplicação web que funciona no navegador. No entanto, você pode encapsular um PWA em um WebApp do T.A.M.K. A principal diferença é que, uma vez encapsulado, ele não será mais "progressivo" no sentido tradicional (não será instalado via navegador), mas sim um aplicativo nativo que contém seu PWA.

---

## Build e Deployment

### O que é uma Keystore e por que ela é importante?

Uma **Keystore** é um arquivo que contém uma chave criptográfica usada para assinar digitalmente seu APK. A assinatura garante a autenticidade e integridade do aplicativo. Sem uma assinatura válida, o Android não permitirá a instalação do APK. Cada projeto criado pelo T.A.M.K pode ter sua própria Keystore, ou você pode usar a Keystore de debug global.

### Posso distribuir meu APK na Google Play Store?

Sim, mas há alguns requisitos adicionais. A Google Play Store exige que os APKs sejam assinados com uma chave de release (não de debug), que você deve gerar e manter em segurança. Além disso, você precisará criar uma conta de desenvolvedor na Google Play Console e seguir as diretrizes de publicação da plataforma.

### Como faço para atualizar um aplicativo já publicado?

Para atualizar um aplicativo, você deve incrementar o `versionCode` e o `versionName` no `AndroidManifest.xml` do seu projeto, fazer as alterações desejadas no código ou no conteúdo web, e então gerar um novo APK. Ao publicar a atualização, você **deve** assinar o novo APK com a **mesma Keystore** usada na versão anterior. Se você perder a Keystore original, não será possível atualizar o aplicativo na Play Store, e você terá que publicá-lo como um novo aplicativo.

### O T.A.M.K suporta builds automatizados (CI/CD)?

Atualmente, o T.A.M.K é projetado para uso interativo no Termux. No entanto, como ele é baseado em scripts Python e comandos shell, é tecnicamente possível integrá-lo a um pipeline de CI/CD, desde que o ambiente de execução tenha as dependências necessárias (Java, Kotlin, aapt2, etc.). Isso requer configuração manual e não é oficialmente suportado.

---

## Troubleshooting

### Meu build falha com o erro "Senha incorreta". O que fazer?

Certifique-se de que você está fornecendo a senha correta da Keystore. Se você criou o projeto com uma senha específica, use essa senha no comando de build: `tamk --build -p SUA_SENHA`. Se você não se lembra da senha, infelizmente não há como recuperá-la. Você precisará gerar uma nova Keystore (o que significa que não poderá atualizar versões anteriores do aplicativo assinadas com a Keystore antiga).

### Meu WebApp exibe uma tela em branco. O que pode estar errado?

Verifique os seguintes pontos:

1.  **Caminho do arquivo**: Certifique-se de que o `webView.loadUrl()` no `MainActivity.kt` está apontando para `file:///android_asset/index.html` e que o arquivo `index.html` existe em `src/main/assets/`.
2.  **JavaScript habilitado**: Confirme que `javaScriptEnabled` está definido como `true` nas configurações do `WebView`.
3.  **Erros no console**: Habilite o debugging remoto (veja a seção de debugging no `DEV_GUIDE.md`) e inspecione o console do Chrome DevTools para identificar erros de JavaScript ou problemas de carregamento de recursos.

### O comando `tamk` não é reconhecido após a instalação. Como resolver?

Isso geralmente significa que o diretório `$PREFIX/bin` não está no seu `PATH`. Tente reiniciar o Termux ou execute `source ~/.bashrc` (ou `source ~/.zshrc` se você usa Zsh). Se o problema persistir, verifique se o script de instalação foi executado com sucesso e se o arquivo `$PREFIX/bin/tamk` existe e tem permissões de execução.

### Posso contribuir com novos templates ou funcionalidades para o T.A.M.K?

Sim! O T.A.M.K é um projeto de código aberto. Consulte o arquivo `CONTRIBUTING.md` para diretrizes sobre como contribuir. Novas ideias, templates e correções de bugs são sempre bem-vindas.
