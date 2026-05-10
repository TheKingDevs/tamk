# 🤝 Guia de Contribuição para o T.A.M.K

Agradecemos profundamente seu interesse em contribuir com o **Termux APK Manager Kit (T.A.M.K)**! Este projeto nasceu da necessidade de democratizar o desenvolvimento Android nativo, e cada contribuição nos aproxima desse objetivo. Sua participação é valiosa para tornar o desenvolvimento acessível a todos, diretamente do dispositivo móvel.

---

## 🏗️ Arquitetura: Entendendo o Motor

Antes de mergulhar no código, é crucial compreender a arquitetura do T.A.M.K para garantir que suas contribuições estejam alinhadas com o design do sistema.

### Fluxo de Dados Principal

O T.A.M.K opera através de uma orquestração clara entre a interface de linha de comando e os controladores de lógica:

1.  **CLI (`main.py`)**: Gerencia os argumentos de entrada e orquestra a execução.
2.  **Controller Factory**: Instancia o controlador adequado com base na ação solicitada.
3.  **Controllers**: Contêm a lógica de negócio específica (`BuildController`, `SetupController`, `InstallController`).
4.  **Template Engine**: Realiza a substituição de placeholders (ex: `{{VAR}}`) nos arquivos `.tmpl`.
5.  **Saída**: Gera os arquivos finais (Kotlin/XML/HTML) integrados a uma SDK isolada.

### Componentes Chave

| Componente | Função |
| :--- | :--- |
| **Motor (CLI)** | Ponto de entrada que valida comandos e parâmetros. |
| **Controllers** | Implementam as operações pesadas como compilação e download de SDK. |
| **Template Engine** | Sistema de busca e substituição para geração dinâmica de código. |
| **Project Factory** | Cria a estrutura de pastas isoladas e garante a portabilidade do projeto. |

---

## 🎯 Áreas de Contribuição

### 1. Novos Templates (Prioridade Alta)

Estamos em busca de novos layouts profissionais para expandir as possibilidades do T.A.M.K. Você pode contribuir criando novos arquivos em `assets/templates/ui_apk/` ou `assets/templates/webapp/`.

**Tipos de Templates Desejados:**
- `login_activity.xml.tmpl` + `LoginActivity.kt.tmpl`
- `settings_activity.xml.tmpl` com `PreferenceFragment`
- `recycler_template.xml.tmpl` com adapter base
- `compose_activity.tmpl` (Jetpack Compose)
- Templates de WebApp otimizados para frameworks como React ou Vue.

### 2. Melhorias nos Controllers

- **BuildController (`src/controllers/build.py`)**: Otimizações em compilação incremental, cache de dependências e suporte a multithreading.
- **SetupController (`src/controllers/setup.py`)**: Download paralelo de componentes da SDK, validação de checksum SHA-256 e sistemas de fallback para mirrors.

### 3. Utilitários e Ferramentas

- **Logger Avançado**: Implementação de barras de progresso ASCII e spinners de atividade em `src/utils/logger.py`.
- **Scripts de Instalação**: Suporte para diferentes shells como Zsh e Fish, além de scripts de desinstalação robustos.

### 4. Testes e Qualidade

- Testes unitários para a **Template Engine**.
- Testes de integração para a **Build Pipeline**.
- Benchmarks de performance e validação de compatibilidade entre diferentes versões do Android/Termux.

---

## 📜 Regras de Estilo e Convenções

### Python (PEP 8 Plus)

Sempre utilize tipagem estática e docstrings detalhadas para manter a manutenibilidade do código.

```python
def compile_apk(project_path: str, keystore_pass: str) -> dict:
    """Compila o projeto APK com as credenciais fornecidas.

    Args:
        project_path: Caminho absoluto para o projeto
        keystore_pass: Senha da keystore (criptografada)

    Returns:
        dict: Resultado da compilação com status e caminho do APK
    """
    # Implementação
    pass
```

### Kotlin e XML Templates

- **Kotlin**: Use 4 espaços para indentação, nomes de variáveis descritivos e separe a lógica de inicialização da lógica de negócio.
- **XML**: Organize os atributos de forma lógica (IDs primeiro, depois layout, depois estilo) e use 4 espaços para indentação.
- **Placeholders**: Devem ser sempre em **MAIÚSCULAS** com underscores (ex: `{{PACKAGE_NAME}}`, `{{ACTIVITY_NAME}}`).

---

## 🚀 Processo de Pull Request

### 1. Pré-requisitos

Antes de enviar seu PR, certifique-se de que seu código passa nos testes e no linting:

```bash
# Executar testes
python3 -m pytest tests/ --verbose

# Verificar linting
python3 -m pylint src/ --rcfile=.pylintrc
```

### 2. Fluxo de Trabalho

1.  **Fork e Clone**: Crie seu fork e clone localmente.
2.  **Branch Descritiva**: Use prefixos como `feature/`, `fix/` ou `docs/`.
3.  **Commits Atômicos**: Faça commits pequenos e com mensagens claras.
4.  **Rebase**: Mantenha sua branch atualizada com a `upstream/main`.
5.  **Push e PR**: Envie suas alterações e abra o Pull Request detalhando as mudanças.

---

## 🐛 Reportando Bugs

Ao abrir uma Issue para reportar um erro, utilize o seguinte template para facilitar o diagnóstico:

- **Ambiente**: Dispositivo, versão do Android, versão do Termux e versão do T.A.M.K.
- **Comando Executado**: O comando exato que causou o erro (use `--verbose`).
- **Comportamento**: O que era esperado vs. o que aconteceu.
- **Logs**: Cole a saída completa do erro.

---

## 🧪 Ambiente de Desenvolvimento

Para configurar seu ambiente de contribuição rapidamente:

```bash
git clone https://github.com/Shadw-Developer/tamk.git
cd tamk
pip install -r requirements-dev.txt
pre-commit install
python -m pytest tests/ -v
```

**Ferramentas Recomendadas:**
- **Editor**: VS Code com extensões Python e Kotlin.
- **Linting/Formatação**: `pylint`, `black` (Python) e `ktlint` (Kotlin).
- **Testes**: `pytest` com `pytest-cov`.

---

🏆 **Reconhecimento**: Todas as contribuições serão creditadas no arquivo `CREDITS.md`. Contribuidores regulares serão convidados para a lista de mantenedores do projeto.

---

<div align="center">
  <sub>Feito com ❤️ por @mrx_dev</sub>
</div>
