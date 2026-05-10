> Este guia prático aborda o fluxo de trabalho para desenvolver e implantar WebApps usando o T.A.M.K, desde a configuração inicial até a preparação para distribuição.

## 1. Ambiente de Desenvolvimento

O T.A.M.K foi projetado para ter o mínimo de dependências externas, mas uma configuração inicial correta é essencial.

### Instalação e Verificação

O script `setup-install.sh` é a maneira recomendada de instalar o T.A.M.K. Ele automatiza as seguintes etapas:

1.  **Verificação de Dependências**: Garante que pacotes como `python`, `openjdk-21`, `aapt2`, e `apksigner` estejam presentes no ambiente Termux.
2.  **Cópia dos Arquivos**: Move o código-fonte do T.A.M.K para um diretório padronizado (`$PREFIX/opt/tamk`).
3.  **Criação do Link Simbólico**: Cria um executável global em `$PREFIX/bin/tamk`, permitindo que você chame o kit de qualquer lugar no seu sistema.

Após a instalação, sempre execute `tamk --version` para confirmar que o ambiente está configurado corretamente.

### Configuração do Projeto

Uma vez que um projeto é criado com `tamk --create`, ele se torna um ambiente autônomo. A etapa mais importante é a configuração da SDK local, que é feita automaticamente na primeira criação de projeto ou manualmente com `tamk --setup`.

> **Nota de Segurança**: Cada projeto pode ter sua própria Keystore (`secret/project.keystore`), que é usada para assinar o APK. Se uma Keystore de projeto não for encontrada, o T.A.M.K utilizará a Keystore de debug global. **Nunca compartilhe suas Keystores.**

## 2. Ciclo de Vida de um WebApp

O desenvolvimento de um WebApp com o T.A.M.K segue um ciclo simples e direto.

### Passo 1: Criação

Comece com o assistente interativo:

```bash
tamk --create
```

-   Escolha a opção **WebApp**.
-   Forneça as informações solicitadas (nome, pacote, autor, senha da Keystore).

### Passo 2: Desenvolvimento Web

Esta é a fase principal. Navegue até o diretório do seu projeto recém-criado. Você encontrará a pasta `src/main/assets/`. **Este é o seu workspace.**

-   **Substitua ou modifique** o `index.html` de exemplo com o seu próprio site.
-   Você pode criar subdiretórios para organizar seus arquivos (`css/`, `js/`, `images/`). Todas as pastas e arquivos dentro de `assets/` serão incluídos no APK.
-   Desenvolva seu site como faria normalmente. O T.A.M.K simplesmente "empacota" este conteúdo.

### Passo 3: Build e Teste

Quando estiver pronto para testar seu WebApp em um dispositivo real, execute o processo de build:

```bash
# Dentro do diretório do seu projeto
tamk --build -p SUA_SENHA
```

Este comando compila os recursos, empacota seus arquivos web, compila o código Kotlin do `WebView` e assina o APK. O resultado é o arquivo `app-final.apk`.

Para instalar, use:

```bash
tamk --install
```

## 3. Customização e Debugging

### Customizando o Comportamento Nativo

Se você precisar alterar o comportamento do `WebView` (por exemplo, habilitar novas permissões, adicionar uma interface de JavaScript para comunicação entre a web e o nativo, etc.), você pode editar diretamente o arquivo `src/main/kotlin/com/example/seuwebapp/MainActivity.kt` no seu projeto gerado. Após a modificação, basta rodar o comando de build novamente.

### Debugging

Debugging de WebApps pode ser desafiador. Aqui estão algumas dicas:

-   **Alerts**: Use `alert('Minha variável: ' + minhaVariavel);` no seu JavaScript para inspecionar valores. O `WebChromeClient` configurado pelo T.A.M.K garante que os alertas funcionem.
-   **Console Remoto**: Para um debugging mais avançado, você pode habilitar o debugging remoto do `WebView`. Adicione o seguinte código ao método `onCreate` da sua `MainActivity.kt`:

    ```kotlin
    WebView.setWebContentsDebuggingEnabled(true);
    ```

    Depois, conecte seu dispositivo ao computador via USB, abra o Google Chrome no desktop e navegue para `chrome://inspect`. Seu `WebView` deverá aparecer como um alvo inspecionável, dando acesso ao console, DOM e network.

## 4. Preparando para Deployment

Antes de distribuir seu aplicativo, considere os seguintes pontos:

-   **Ícone do App**: Substitua os ícones padrão localizados em `res/mipmap/` por ícones personalizados.
-   **Versão do App**: Edite o arquivo `tamk.config` na raiz do seu projeto para incrementar o número da versão (`version=1.0.1`). Esta informação será usada no `AndroidManifest.xml` durante o próximo build.
-   **Build de Release**: O processo de build padrão já gera um APK assinado e alinhado (`zipalign`), que é considerado pronto para release. Não há uma distinção entre build de "debug" e "release" no fluxo atual do T.A.M.K, pois a senha da Keystore é sempre exigida.
-   **Ofuscação (ProGuard/R8)**: Atualmente, o T.A.M.K não integra ferramentas de ofuscação de código Kotlin. Para WebApps, a maior parte da sua lógica estará em JavaScript, que pode ser minificada e ofuscada usando ferramentas web padrão (como Webpack, Terser) antes de ser colocada na pasta `assets`.
