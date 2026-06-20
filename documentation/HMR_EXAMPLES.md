# 🎯 Exemplos Práticos de HMR

Coleção de padrões e receitas para usar o HMR do T.A.M.K de forma eficiente.

---

## Estrutura de Exemplo

```
src/main/assets/
├── index.html
├── css/
│   ├── styles.css
│   ├── components.css
│   └── themes/ (dark.css, light.css)
├── js/
│   ├── app.js
│   ├── modules/
│   │   ├── navigation.js
│   │   ├── cart.js
│   │   ├── products.js
│   │   └── user.js
│   ├── components/
│   │   ├── button.js
│   │   ├── modal.js
│   │   └── carousel.js
│   └── utils/
│       ├── storage.js
│       └── api.js
├── data/
│   ├── products.json
│   ├── config.json
│   └── translations/ (pt-BR.json, en-US.json)
└── images/
```

---

## Exemplo 1: App Modular (E-commerce)

### Bootstrap (`js/app.js`)

```javascript
import { Navigation } from './modules/navigation.js';
import { Cart } from './modules/cart.js';
import { Products } from './modules/products.js';

window.App = {
  version: '1.0.0',
  modules: {},

  init() {
    this.modules.navigation = new Navigation();
    this.modules.cart = new Cart();
    this.modules.products = new Products();
    this.registerHMR();
    this.render();
  },

  registerHMR() {
    // Exemplo conceitual (HMR bridge ainda em desenvolvimento)
    // O bridge atual injeta WebSocket client que escuta reload/css-update
    ws.onmessage = function(e) {
      var msg = JSON.parse(e.data);
      if (msg.type === 'reload') { location.reload(); }
    };
  },

  saveModuleState(name) {
    if (this.modules[name]?.getState) {
      TAMK_HMR.saveState(name, this.modules[name].getState());
    }
  },

  restoreModuleState(name) {
    const saved = TAMK_HMR.getState(name);
    if (saved && this.modules[name]?.setState) {
      this.modules[name].setState(saved);
    }
  },

  render() {
    document.getElementById('app').innerHTML = `
      <nav id="nav"></nav>
      <div id="products"></div>
      <div id="cart"></div>`;
    this.modules.navigation.render();
    this.renderProducts();
    this.updateCartUI();
  }
};
```

### Módulo Cart com Preservação de Estado (`js/modules/cart.js`)

```javascript
class Cart {
  constructor() {
    this.items = JSON.parse(localStorage.getItem('cart')) || [];
  }

  add(product) {
    const existing = this.items.find(i => i.id === product.id);
    existing ? existing.qty++ : this.items.push({...product, qty: 1});
    this.save();
  }

  remove(id) {
    this.items = this.items.filter(i => i.id !== id);
    this.save();
  }

  getTotal() { return this.items.reduce((s, i) => s + i.price * i.qty, 0); }
  getState() { return { items: this.items, total: this.getTotal() }; }
  setState(state) { if (state.items) this.items = state.items; }
  save() { localStorage.setItem('cart', JSON.stringify(this.items)); }
}
```

---

## Exemplo 2: CSS Hot Reload com Tema Dinâmico

```javascript
// js/themes.js
const ThemeManager = {
  current: 'dark',
  
  init() {
    this.current = localStorage.getItem('theme') || 'dark';
    this.apply();
  },
  
  apply() {
    document.documentElement.setAttribute('data-theme', this.current);
    document.getElementById('themeBtn').textContent =
      this.current === 'dark' ? '☀️ Claro' : '🌙 Escuro';
  },
  
  toggle() {
    this.current = this.current === 'dark' ? 'light' : 'dark';
    localStorage.setItem('theme', this.current);
    this.apply();
  }
};

document.addEventListener('DOMContentLoaded', () => ThemeManager.init());
```

CSS updates via TAMK_HMR afetam o tema atual instantaneamente sem perder a seleção.

---

## Exemplo 3: Dados em Tempo Real com JSON HMR

```javascript
// js/data-manager.js
const DataManager = {
  config: {},
  products: [],
  
  async load() {
    this.config = await this.fetchJSON('data/config.json');
    this.products = await this.fetchJSON('data/products.json');
    this.render();
  },
  
  async fetchJSON(path) {
    const resp = await fetch(path);
    return resp.json();
  },
  
  render() {
    if (this.config.theme) ThemeManager.current = this.config.theme;
    // Renderiza produtos...
  }
};

// JSON updates via HMR recarregam dados sem refresh
TAMK_HMR.accept((update) => {
  if (update.type === 'json') {
    DataManager.load();
  }
});
```

---

## Exemplo 4: Estado de Scroll

```javascript
// Salva scroll antes de qualquer rebuild
window.addEventListener('beforeunload', () => {
  TAMK_HMR.saveState('scroll', window.scrollY);
});

// Restaura após HMR ou rebuild
TAMK_HMR.accept(() => {
  const scroll = TAMK_HMR.getState('scroll');
  if (scroll !== undefined) setTimeout(() => window.scrollTo(0, scroll), 100);
});
```

---

## Boas Práticas

1. **Módulos pequenos** — Cada arquivo JS deve ser um módulo independente
2. **Estado exportado** — Módulos devem expor `getState()`/`setState()`
3. **Handlers específicos** — Prefira `TAMK_HMR.accept('path', handler)` ao invés de handler global
4. **Fallback para reload** — Se o HMR falhar, retorne `false` no handler para forçar full reload
5. **Cleanup** — Use `TAMK_HMR.dispose()` quando remover módulos dinamicamente

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Exemplos Práticos</sub>
</div>
