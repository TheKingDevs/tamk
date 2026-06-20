# 🔄 HMR — Hot Module Replacement Guide

> **Versão:** 2026.3.0-HMR — Guia rápido de referência do sistema HMR.

---

## Visão Geral

HMR permite atualizar módulos JS/CSS/JSON em tempo real **sem recarregar a página** e **preservando o estado**.

### Fluxo

```
Developer → TAMK Dev (WebSocket) → WebView
  1. Edita arquivo
  2. Detecta mudança
  3. Envia update
  4. Aplica módulo
  5. Preserva estado
```

### Tipos

| Extensão | Comportamento | Reload | Estado |
| :--- | :--- | :--- | :--- |
| `.css` | Hot reload | ❌ | ✅ |
| `.js` | Module injection | ❌ | ✅ |
| `.json` | Data update | ❌ | ✅ |
| `.html` | Rebuild + reload | ✅ | ⚠️ Parcial |
| `.png/.jpg` | Rebuild + reload | ✅ | ⚠️ Parcial |

---

## Como Usar

```bash
cd seu-projeto-webapp
tamk dev
```

Edite arquivos em `src/main/assets/` — as mudanças aparecem automaticamente no dispositivo.

---

## API JavaScript

### `TAMK_HMR.accept(handler)`

```javascript
// Global
TAMK_HMR.accept((update) => {
    console.log('Update:', update.modulePath);
    return true; // false = full reload
});

// Específico
TAMK_HMR.accept('js/app.js', (update) => {
    if (window.myApp) window.myApp.reload();
});
```

### `TAMK_HMR.saveState(key, value)`

```javascript
TAMK_HMR.saveState('scrollY', window.scrollY);
TAMK_HMR.saveState('form', { name: 'João', email: 'joao@email.com' });
```

### `TAMK_HMR.getState(key)`

```javascript
const scrollY = TAMK_HMR.getState('scrollY');
if (scrollY) window.scrollTo(0, scrollY);
```

### `TAMK_HMR.dispose(path)`

```javascript
TAMK_HMR.dispose('js/old.js'); // Cleanup
```

---

## Exemplos

### 1. Preservar Formulário

```javascript
document.querySelectorAll('input, textarea').forEach(f => {
    f.addEventListener('input', () => {
        const state = {};
        document.querySelectorAll('input').forEach(i => state[i.id] = i.value);
        TAMK_HMR.saveState('formState', state);
    });
});

TAMK_HMR.accept(() => {
    const saved = TAMK_HMR.getState('formState');
    if (saved) Object.entries(saved).forEach(([id, val]) => {
        const el = document.getElementById(id);
        if (el) el.value = val;
    });
});
```

### 2. Componente com Estado

```javascript
const Counter = {
    count: 0,
    init() { this.render(); },
    render() { document.getElementById('count').textContent = this.count; },
    increment() { this.count++; this.render(); }
};

TAMK_HMR.accept('js/counter.js', (update) => {
    const saved = Counter.count;
    eval(update.newContent);
    Counter.count = saved;
    Counter.init();
});
```

---

## Estrutura Recomendada

```
src/main/assets/
├── index.html
├── css/
│   ├── styles.css
│   ├── components.css
│   └── themes/
├── js/
│   ├── app.js
│   ├── modules/
│   │   ├── navigation.js
│   │   ├── cart.js
│   │   └── products.js
│   └── components/
│       ├── button.js
│       ├── modal.js
│       └── carousel.js
├── data/
│   ├── config.json
│   └── translations/
└── images/
```

---

## Troubleshooting

| Problema | Solução |
| :--- | :--- |
| WebSocket não conecta | `tamk dev --no-ws` |
| HMR não funciona | Verifique se bridge foi injetado |
| Build lento | Ignore `node_modules/` |
| ADB não conecta | Instale manualmente |

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — HMR Guide</sub>
</div>
