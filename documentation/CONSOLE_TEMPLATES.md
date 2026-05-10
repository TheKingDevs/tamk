# 📄 Templates para Console

> Este documento contém o código-fonte dos templates utilizados na criação de projetos do tipo Console. Estes arquivos estão localizados em `assets/templates/console/`.

## Main.kt.tmpl

Este é o ponto de entrada para aplicações de linha de comando em Kotlin.

```kotlin
/**
 * Project: {{NAME}}
 * Version: {{VERSION}}
 * Author: {{AUTHOR}}
 */
fun main() {
    println("Olá do Console T.A.M.K!")
}
```

**Placeholders Utilizados:**

- `{{NAME}}`: Nome do projeto definido durante a criação.
- `{{VERSION}}`: Versão inicial do projeto.
- `{{AUTHOR}}`: Nome do autor/desenvolvedor.

---

## Execução Simplificada

Diferente dos projetos de APK, os projetos de Console não requerem uma senha de Keystore e podem ser executados diretamente sem um processo de build complexo.

### Como Executar:

1.  **Dentro do Projeto**: Basta digitar `tamk --run` (ou `tamk -r`). O T.A.M.K detectará automaticamente o arquivo `src/Main.kt` e o executará.
2.  **Arquivo Isolado**: Você também pode executar qualquer arquivo Kotlin isolado usando `tamk --run [arquivo.kt]`.

---

## Código Python: `console.py`

A classe `ConsoleStructure` gerencia a criação da estrutura de pastas e o processamento do template acima.

```python
import os

class ConsoleStructure:
    def setup(self, name, version, author, password=None):
        """
        Configura a estrutura de um projeto do tipo Console.
        """
        base_path = os.path.join(os.getcwd(), name)

        # Estrutura simplificada para Console
        folders = ["src", "libs", "build"]
        for f in folders:
            os.makedirs(os.path.join(base_path, f), exist_ok=True)

        # Caminho do template
        template_path = "assets/templates/console/Main.kt.tmpl"
        
        if os.path.exists(template_path):
            with open(template_path, "r") as f:
                content = f.read()
            
            # Substituição de placeholders
            content = content.replace("{{NAME}}", name)
            content = content.replace("{{VERSION}}", version)
            content = content.replace("{{AUTHOR}}", author)
            
            with open(os.path.join(base_path, "src/Main.kt"), "w") as f:
                f.write(content)
        else:
            # Fallback
            # ... (código de fallback)
```
