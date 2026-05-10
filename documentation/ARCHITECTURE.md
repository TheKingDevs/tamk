> Este documento detalha a arquitetura interna do T.A.M.K, com foco na interação entre os componentes e no fluxo de dados, especialmente após a introdução do suporte a WebApps.

## Visão Geral da Arquitetura

A arquitetura do T.A.M.K é baseada em um padrão de **injeção de dependência** e **factory**, onde a interface de linha de comando (CLI) atua como o ponto de entrada que orquestra diferentes controladores e serviços. O sistema é projetado para ser modular, permitindo que novos tipos de projeto, como o de WebApp, sejam adicionados com impacto mínimo no núcleo existente.

```mermaid
graph TD
    subgraph "Entrada do Usuário"
        A[CLI: main.py]
    end

    subgraph "Orquestração"
        B(ProjectManager)
        C(BuildController)
        D(SetupController)
    end

    subgraph "Criação de Projetos"
        E{ProjectFactory}
    end

    subgraph "Estruturas de Projeto"
        F[UI/APK Structure]
        G[Console Structure]
        H[WebApp Structure]
    end

    subgraph "Motor de Templates"
        I[Template Engine]
    end

    subgraph "Saída"
        J((Projeto Gerado))
        K((APK Final))
    end

    A -- tamk --create --> B;
    A -- tamk --build --> C;
    A -- tamk --setup --> D;

    B --> E;

    E -- "ui_apk" --> F;
    E -- "console" --> G;
    E -- "webapp" --> H;

    F & G & H --> I;
    I --> J;
    C --> K;
```

## Componentes Principais

| Componente | Arquivo(s) | Responsabilidade |
| :--- | :--- | :--- |
| **Interface CLI** | `src/main.py` | Ponto de entrada que interpreta os argumentos da linha de comando (`--create`, `--build`, etc.) e aciona os controladores correspondentes. |
| **Controladores** | `src/controllers/` | Contêm a lógica de negócio principal. `BuildController` gerencia a compilação, `ProjectManager` orquestra a criação de projetos e `SetupController` cuida da configuração do ambiente. |
| **Project Factory** | `src/organization/factory.py` | Componente central para a criação de projetos. Com base no tipo de projeto selecionado pelo usuário, ele instancia a classe de estrutura apropriada (`WebAppStructure`, `UIAppStructure`, etc.). |
| **Estruturas** | `src/organization/structures/` | Definem o "esqueleto" de cada tipo de projeto. A `WebAppStructure`, por exemplo, sabe quais pastas criar (`src/main/assets`) e quais templates processar para um projeto WebApp. |
| **Template Engine** | (Implementado nas classes de estrutura) | Um mecanismo simples de busca e substituição que lê os arquivos `.tmpl` da pasta `assets/templates/`, substitui os placeholders (ex: `{{NAME}}`, `{{PACKAGE}}`) e grava os arquivos finais no diretório do projeto gerado. |

## Fluxo de Criação de um WebApp

1.  O usuário executa `tamk --create`.
2.  O `main.py` aciona o `ProjectManager`.
3.  O `ProjectManager` inicia o assistente interativo e pergunta o tipo de projeto.
4.  O usuário seleciona "WebApp".
5.  O `ProjectManager` invoca `ProjectFactory.create("webapp", ...)` com os dados do usuário.
6.  O `ProjectFactory` instancia `WebAppStructure`.
7.  A `WebAppStructure.setup()` é executada, realizando as seguintes ações:
    *   Cria a estrutura de diretórios, incluindo `src/main/assets/`.
    *   Processa os templates específicos de WebApp (como `MainActivity.kt.tmpl` que contém a lógica do `WebView`) e os templates comuns (ícones, estilos).
    *   Gera uma Keystore privada para o projeto.
8.  O resultado é um projeto Android completo e autônomo, pronto para ter seu conteúdo web adicionado à pasta `assets`.

## O Papel do WebView

O coração do projeto WebApp é o componente `WebView` do Android. A classe `MainActivity.kt` gerada pelo T.A.M.K é responsável por configurar este componente de forma otimizada:

- **JavaScript Ativado**: `webView.settings.javaScriptEnabled = true` permite que toda a lógica do seu site funcione como esperado.
- **Armazenamento DOM**: `webView.settings.domStorageEnabled = true` é crucial para frameworks modernos que utilizam `localStorage`.
- **Navegação Interna**: O `WebViewClient` é configurado para que todos os links clicados dentro do seu site abram no próprio aplicativo, em vez de em um navegador externo.
- **Carregamento Local**: O `WebView` é instruído a carregar o arquivo `file:///android_asset/index.html`, que aponta diretamente para a pasta `assets` do seu projeto APK.
