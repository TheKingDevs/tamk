# 📄 Templates para Console

Templates em `templates/console/`.

---

## Main.kt.tmpl

```kotlin
/**
 * Project: {{NAME}}
 * Version: {{VERSION}}
 * Author: {{AUTHOR}}
 */
fun main() {
    println("Olá do Console T.A.M.K!")
    println("Projeto: {{NAME}}")
    println("Versão: {{VERSION}}")
    println("Autor: {{AUTHOR}}")
}
```

**Placeholders:** `{{NAME}}`, `{{VERSION}}`, `{{AUTHOR}}`

---

## Execução

```bash
# No diretório do projeto Console
tamk run

# Ou arquivo específico
tamk run src/MeuScript.kt
```

---

## Console Template

A classe `ConsoleStructure` gerencia a criação:

- Cria pastas `src/` e `libs/`
- Processa `Main.kt.tmpl` ou usa fallback inline
- Salva `tamk.config` com `type=console`
- Não gera keystore (console não precisa de assinatura)

---

<div align="center">
  <sub>T.A.M.K v1.0.0 — Templates Console</sub>
</div>
