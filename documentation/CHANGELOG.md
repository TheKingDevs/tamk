# 📋 Changelog do T.A.M.K

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo. O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/), e este projeto adere ao [Versionamento Semântico](https://semver.org/lang/pt-BR/).

---

## [ 2.3.2 ] - 2026-03-29

### 🔧 Corrigido

-   **Caminho de Instalação**: Correção no export do caminho de instalação e atualização do ambiente esperado.
    -   Path resolution corrigido para Termux.
    -   Variáveis de ambiente de output atualizadas.

---

## [ 2.3.1 ] - 2026-03-29

### 🎨 Adicionado

-   **Identidade Visual**: Integração do logo do projeto.

### 🔧 Corrigido

-   **Sistema de Banner**: Correção no sistema de renderização e atualização do banner.
    -   Lógica de update do banner corrigida.
    -   Renderização via toilet ajustada.

---

## [ 2.3.0 ] - 2026-03-29

### ✨ Adicionado

-   **HMR - Hot Module Replacement**: Sistema completo de atualização em tempo real para WebApps.
    -   **CSS Hot Reload**: Atualização de CSS sem reload da página, preservando scroll e estado.
    -   **JavaScript HMR**: Injeção de módulos JS sem reload, com suporte a handlers customizados.
    -   **JSON HMR**: Atualização de dados JSON em tempo real.
    -   **State Preservation**: Estado da aplicação (formulários, scroll, dados) é preservado entre updates.
    -   Nova classe `DevServer` com módulo registry e client state tracking em `src/controllers/dev_controller.py`.
    -   Bridge JavaScript atualizado para versão `2026.3.0-HMR` em `assets/templates/webapp/dev_bridge.js.tmpl`.
    -   API HMR exposta globalmente: `TAMK_HMR.accept()`, `TAMK_HMR.saveState()`, `TAMK_HMR.getState()`.
    -   Nova documentação: `HMR_GUIDE.md` com exemplos e troubleshooting.

### 🔧 Modificado

-   **`DevController`**:
    -   Adicionado método `_on_file_changed()` com suporte a HMR por tipo de arquivo.
    -   Método `_quick_assets_build()` agora envia notificação de reload com state preservation.
    -   Instructions atualizadas com informações sobre HMR API.
    -   Comando de status `s` agora mostra módulos registrados e clientes HMR.
-   **`DevServer`**:
    -   Adicionado `handler()` para gerenciar conexões WebSocket com handshake HMR.
    -   Adicionado `register_module()` para registry de módulos JS.
    -   Adicionado `broadcast_hmr_update()` para envio seletivo de updates por tipo de arquivo.
    -   Adicionado `client_states` tracking para monitorar estado de cada cliente.
-   **`dev_bridge.js.tmpl`**:
    -   Atualizado para versão 2026.3.0-HMR.
    -   Adicionado objeto `TAMK_DEV.hmr` com registry de módulos e handlers.
    -   Adicionado método `handleJavaScriptHMR()` para injeção de módulos.
    -   Adicionado `captureFormData()` e `restoreFormData()` para preservação de estado.
    -   Adicionado heartbeat com envio de módulos carregados.

### 📚 Documentação

-   **`HMR_GUIDE.md`**: Guia completo de Hot Module Replacement com exemplos práticos.

---

## [ 2.2.0 ] - 2026-02-27

### ✨ Adicionado

-   **WebApps Hospedados**: Suporte para criação de WebApps via URL.
    -   Carregamento de aplicações web remotas no WebView.
    -   Flexibilidade para apps baseados em serviços externos.

### 🔗 Novos Casos de Uso

-   Desenvolvedores podem agora criar APKs que carregam aplicações web já hospedadas.
-   Ideal para wrappers de PWA e aplicações SaaS.

---

## [ 2.1.2 ] - 2026-01-21

### 📌 Atualizado

-   **Créditos de Contribuidor**: Atualização no README com créditos de contribuidores.

---

## [ 2.1.1 ] - 2026-01-21

### 🔧 Corrigido

-   **Projetos Console**: Correção de erros na construção e estrutura de projetos Console.
    -   Template `Main.kt` implementado para entrada de aplicações console.
    -   Estrutura de diretórios corrigida.

### 📚 Documentação

-   **`CONSOLE_TEMPLATES.md`**: Documentação completa de templates para projetos Console.

---

## [ 2.1.0 ] - 2026-01-20

### ✨ Adicionado

-   **CLI Modernizada**: Nova interface de linha de comando com UX aprimorada.
    -   Sistema de input moderno com suporte a `tput`.
    -   Centralização de layout via `toilet`.
    -   Esquema de cores atualizado em `colors.py`.

### 🔧 Melhorado

-   **URLs do Repositório**: Correção de todas as URLs nos documentos:
    -   `README.md`, `CONTRIBUTING.md`, `QUICKSTART.md` atualizados.
-   **Interface do Usuário**: Experiência de terminal profissional e consistente.

---

## [ 2.0.1 ] - 2026-01-20

### 📝 Corrigido

-   **README**: Ajuste de formatação no título.

---

## [ 2.0.0 ] - 2026-01-20

### ✨ Adicionado

-   **Suporte a WebApps**: Introdução de um novo tipo de projeto que permite encapsular aplicações web (HTML, CSS, JavaScript) em um APK nativo do Android.
    -   Nova classe `WebAppStructure` em `src/organization/structures/webapp.py`.
    -   Novos templates em `assets/templates/webapp/`:
        -   `AndroidManifest.xml.tmpl` com permissões de internet.
        -   `MainActivity.kt.tmpl` com configuração otimizada de `WebView`.
        -   `index.html.tmpl` como ponto de partida para desenvolvimento web.
    -   Criação automática da pasta `src/main/assets/` em projetos WebApp, onde o desenvolvedor coloca seus arquivos web.

### 📚 Documentação

-   **Documentação Completa**: Novos arquivos de documentação para guiar desenvolvedores:
    -   `ARCHITECTURE.md`: Visão geral da arquitetura do sistema.
    -   `API_COMPONENTS.md`: Referência técnica de classes e módulos.
    -   `DEV_GUIDE.md`: Guia prático de desenvolvimento e deployment.
    -   `WEBAPP_TEMPLATES.md`: Código-fonte completo de todos os templates de WebApp.
    -   `CONSOLE_TEMPLATES.md`: Templates de projetos Console.
    -   `FAQ.md`: Perguntas frequentes.
    -   `QUICKSTART.md`: Guia de início rápido.

### 🔧 Modificado

-   **`BuildController`**: Atualizado para detectar e empacotar a pasta `src/main/assets/` durante o build, usando a flag `-A` do `aapt2 link`.
-   **`ProjectFactory`**: Adicionado o mapeamento `"webapp": WebAppStructure()` para suportar a criação de projetos WebApp.
-   **`README.md`**: Atualizado para incluir informações sobre o novo tipo de projeto WebApp e seu fluxo de trabalho.

### 🐛 Corrigido

-   Correção na lógica de detecção de Keystore no `BuildController`, garantindo que a senha correta seja usada para projetos com Keystore privada.

### ⚠️ Breaking Changes

-   **Documentação movida**: A documentação foi movida da raiz do projeto para a pasta `documentation/`.

---

## [ 1.0.0 ] - 2026-01-20

### ✨ Adicionado

-   **Lançamento Inicial**: Primeira versão estável do T.A.M.K.
    -   Suporte para criação de projetos do tipo **UI APK** (aplicativos Android nativos com interface XML).
    -   Suporte para criação de projetos do tipo **Console** (aplicações Kotlin de linha de comando).
    -   Pipeline completo de build: compilação de recursos, código Kotlin, geração de DEX, assinatura e alinhamento de APK.
    -   Sistema de templates modular com placeholders (`{{NAME}}`, `{{PACKAGE}}`, etc.).
    -   Script de instalação `setup-install.sh` para Termux.
    -   Keystore privada por projeto.
    -   Factory pattern para estruturas de projeto.

### 🏗️ Arquitetura

-   **Controllers**: `ProjectManager`, `BuildController`, `SetupController`.
-   **Structures**: `UIAppStructure`, `ConsoleStructure`, `WebAppStructure`.
-   **Templates**: Sistema de templates `.tmpl` com injeção de placeholders.

---

## Formato de Versionamento

O T.A.M.K utiliza o **Versionamento Semântico (SemVer)** no formato `MAJOR.MINOR.PATCH`:

-   **MAJOR**: Mudanças incompatíveis ou marcos significativos
-   **MINOR**: Novas funcionalidades compatíveis
-   **PATCH**: Correções de bugs e melhorias menores

**Exemplo**: `2.3.0` indica a segunda major release, terceira minor release, versão estável.

### Tags de Release

Cada versão é marcada com uma tag anotada no Git:
```bash
git tag -l          # Listar todas as tags
git show v2.3.0     # Ver detalhes de uma release
```
