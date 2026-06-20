# Sistema de Atualizações do T.A.M.K

Sistema de atualizações automáticas com níveis de prioridade.

---

## Visão Geral

- Detecção automática via GitHub API
- Níveis de prioridade (CRITICAL → auto-install, MAJOR/MINOR → prompt)
- Cache de 6 horas para evitar requisições excessivas
- Notificações visuais com banners coloridos
- `UpdateUseCase.Check()` consulta a GitHub API, retorna info, exibe ao usuário

---

## Níveis de Atualização

| Nível | Descrição | Comportamento |
| :--- | :--- | :--- |
| 🔴 **CRITICAL** | Segurança, bugs críticos | Instalação **automática** |
| 🟡 **PATCH** | Correções menores | Instalação **automática** |
| 🟢 **MAJOR** | Novas funcionalidades | **Pergunta** ao usuário |
| 🔵 **MINOR** | Melhorias incrementais | **Pergunta** ao usuário |
| ⚪ **OPTIONAL** | Cosmético/docs | **Notifica** apenas |

### Detecção de Nível

1. **Versão semântica**: Compara major.minor.patch
2. **Palavras-chave nas notas de release**:
   - `critical`, `security`, `urgent` → CRITICAL
   - `patch`, `bugfix`, `fix` → PATCH
   - Mudança major → MAJOR
   - Mudança minor → MINOR

---

## Como Usar

```bash
# Verificação manual
tamk update
```

---

## Funcionamento Interno

### Cache
- Local: `os.TempDir() + "/tamk-update/update_cache.json"`
- TTL: 6 horas
- Ignorado em `update` ou se corrompido

### Update Methods (auto-detected)

1. **git**: `git pull --rebase --autostash` in TAMK_HOME
2. **go install**: `go install github.com/TheKingDevs/tamk/cmd/tamk@latest`

### Fluxo

```
1. Verificar GitHub API → cache válido?
2. Nova versão? → Detectar nível
3. CRITICAL/PATCH → Auto-instalar
4. MAJOR/MINOR → Perguntar usuário
5. Atualizar → Sucesso? → OK
```

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Update System</sub>
</div>
