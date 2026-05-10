> Este documento fornece uma referência técnica sobre as principais classes, módulos e templates que compõem o T.A.M.K, com ênfase nas adições para o suporte a WebApps.

## Módulos Python (`src/`)

O núcleo do T.A.M.K é escrito em Python e está localizado no diretório `src/`. A seguir, detalhamos os componentes mais importantes.

### `organization/factory.py`

**Classe: `ProjectFactory`**

Esta classe atua como um ponto central de distribuição para a criação de diferentes tipos de projetos. Seu método estático `create` é a chave para a extensibilidade do sistema.

```python
class ProjectFactory:
    @staticmethod
    def create(p_type, name, version, author, password):
        structures = {
            "ui_apk": UIAppStructure(),
            "console": ConsoleStructure(),
            "webapp": WebAppStructure() # <-- Ponto de extensão
        }
        builder = structures.get(p_type)
        if builder:
            return builder.setup(name, version, author, password)
        return False
```

-   **Uso**: O `ProjectManager` chama este método com o tipo de projeto (`"webapp"`) e os metadados coletados do usuário. A fábrica então delega a tarefa de construção para a classe de estrutura correspondente.

### `organization/structures/webapp.py`

**Classe: `WebAppStructure`**

Esta é a nova classe responsável por definir a estrutura e os arquivos de um projeto WebApp. Ela herda a responsabilidade de orquestrar a criação de pastas e o processamento de templates.

**Principais Métodos:**

-   `setup()`: O método principal que coordena todo o processo de criação.
-   `_create_folders()`: Cria a árvore de diretórios necessária, incluindo a crucial pasta `src/main/assets`, que servirá como a raiz para os arquivos web (HTML, CSS, JS).
-   `_generate_file()`: Lê um arquivo de template (`.tmpl`), substitui os placeholders e o salva no diretório de destino do projeto.
-   `_generate_keystore()`: Gera uma chave de assinatura privada para o projeto, garantindo que cada WebApp seja um artefato de software seguro e independente.

### `controllers/build_controller.py`

**Classe: `BuildController`**

Este controlador foi atualizado para lidar com as particularidades dos WebApps. A principal mudança está no comando `aapt2 link`, que agora inclui a flag `-A` para empacotar o diretório `assets` dentro do APK.

```python
# ... dentro de build_apk()

assets_path = "src/main/assets"
assets_flag = f"-A {assets_path}" if os.path.exists(assets_path) else ""

link_cmd = (f"aapt2 link -I {self.conf.SDK_PATH} --manifest AndroidManifest.xml "
            f"--java {self.cache_dir}/gen -o {self.cache_dir}/app.apk "
            f"{assets_flag} {self.cache_dir}/res.zip --auto-add-overlay")
```

-   **Lógica**: O controlador verifica a existência da pasta `src/main/assets`. Se ela existir, a flag `-A` é adicionada ao comando `aapt2 link`, instruindo a ferramenta a incluir todo o conteúdo dessa pasta na raiz do diretório `assets/` dentro do APK final.

## Templates (`assets/templates/`)

Os templates são a base para a geração de código. Eles contêm placeholders que são substituídos pelos metadados do projeto durante a criação.

### `webapp/MainActivity.kt.tmpl`

Este é o template Kotlin que define a atividade principal do WebApp. Ele contém a lógica para inicializar e configurar o `WebView`.

```kotlin
package {{PACKAGE}}

import android.app.Activity
import android.os.Bundle
import android.webkit.WebChromeClient
import android.webkit.WebView
import android.webkit.WebViewClient

class MainActivity : Activity() {

    private lateinit var webView: WebView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        webView = WebView(this)
        
        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            allowFileAccess = true
        }

        webView.webViewClient = WebViewClient()
        webView.webChromeClient = WebChromeClient()

        setContentView(webView)

        webView.loadUrl("file:///android_asset/index.html")
    }
    
    // ... Lógica para o botão "voltar"
}
```

-   **`file:///android_asset/index.html`**: Esta URL é o ponto de conexão crucial. Ela instrui o `WebView` a carregar o arquivo `index.html` localizado dentro da pasta `assets/` do APK, que é exatamente onde o `BuildController` coloca os arquivos do seu site.

### `webapp/index.html.tmpl`

Um template HTML simples é fornecido como ponto de partida. Ele demonstra que o ambiente está funcionando e que o JavaScript pode ser executado.

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{NAME}}</title>
    <!-- ... estilos ... -->
</head>
<body>
    <h1>Olá, WebApp!</h1>
    <p>Rodando nativamente no Android via T.A.M.K</p>
    <button onclick="alert('JavaScript funcionando!')">Testar JS</button>
</body>
</html>
```

### Placeholders Comuns

| Placeholder | Descrição | Exemplo |
| :--- | :--- | :--- |
| `{{NAME}}` | O nome do aplicativo, definido pelo usuário. | `Meu WebApp Incrível` |
| `{{PACKAGE}}` | O nome do pacote Java/Kotlin. | `com.exemplo.meuwebapp` |
| `{{VERSION}}` | A versão do aplicativo. | `1.0.0` |
| `{{AUTHOR}}` | O nome do autor do projeto. | `João da Silva` |
