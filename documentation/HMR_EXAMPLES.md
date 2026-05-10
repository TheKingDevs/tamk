# 🎯 Exemplos Práticos de HMR

Coleção de padrões e receitas para usar o sistema HMR do T.A.M.K de forma eficiente.

---

## 📦 Estrutura de Projeto de Exemplo

```
MyWebApp/
├── src/main/assets/
│   ├── index.html
│   ├── css/
│   │   ├── styles.css
│   │   ├── components.css
│   │   └── themes/
│   │       ├── dark.css
│   │       └── light.css
│   ├── js/
│   │   ├── app.js              # Main bootstrap
│   │   ├── modules/
│   │   │   ├── navigation.js
│   │   │   ├── cart.js
│   │   │   ├── products.js
│   │   │   └── user.js
│   │   ├── components/
│   │   │   ├── button.js
│   │   │   ├── modal.js
│   │   │   └── carousel.js
│   │   └── utils/
│   │       ├── storage.js
│   │       └── api.js
│   ├── data/
│   │   ├── products.json
│   │   ├── config.json
│   │   └── translations/
│   │       ├── pt-BR.json
│   │       └── en-US.json
│   └── images/
│       ├── logo.png
│       ├── icons/
│       └── backgrounds/
├── AndroidManifest.xml
├── tamk.config
└── secret/project.keystore
```

---

## 🏗️ Exemplo 1: App com Módulos Independentes

### Contexto

App de e-commerce com módulos separados que podem ser atualizados independentemente sem perder carrinho ou sessão.

### Estrutura JS

```javascript
// js/app.js - Bootstrap principal
import { Navigation } from './modules/navigation.js';
import { Cart } from './modules/cart.js';
import { Products } from './modules/products.js';

window.App = {
  version: '1.0.0',
  modules: {},

  init: function() {
    console.log('[App] Inicializando v' + this.version);
    
    // Inicializa cada módulo
    this.modules.navigation = new Navigation();
    this.modules.cart = new Cart();
    this.modules.products = new Products();
    
    // Registra no TAMK_HMR
    this.registerHMR();
    
    // Renderiza UI inicial
    this.render();
  },

  registerHMR: function() {
    // Handler para cada módulo
    TAMK_HMR.accept('js/modules/navigation.js', (update) => {
      console.log('[HMR] Navigation atualizado');
      this.saveModuleState('navigation');
      eval(update.newContent);
      this.modules.navigation = new Navigation();
      this.restoreModuleState('navigation');
      return true;
    });

    TAMK_HMR.accept('js/modules/cart.js', (update) => {
      console.log('[HMR] Cart atualizado');
      this.saveModuleState('cart');
      eval(update.newContent);
      this.modules.cart = new Cart();
      this.restoreModuleState('cart');
      this.updateCartUI();
      return true;
    });

    TAMK_HMR.accept('js/modules/products.js', (update) => {
      console.log('[HMR] Products atualizado');
      eval(update.newContent);
      this.modules.products = new Products();
      this.renderProducts();
      return true;
    });
  },

  saveModuleState: function(moduleName) {
    if (this.modules[moduleName] && this.modules[moduleName].getState) {
      const state = this.modules[moduleName].getState();
      TAMK_HMR.saveState(moduleName, state);
      console.log('[HMR] Estado salvo:', moduleName, state);
    }
  },

  restoreModuleState: function(moduleName) {
    const saved = TAMK_HMR.getState(moduleName);
    if (saved && this.modules[moduleName] && this.modules[moduleName].setState) {
      this.modules[moduleName].setState(saved);
      console.log('[HMR] Estado restaurado:', moduleName, saved);
    }
  },

  render: function() {
    document.getElementById('app').innerHTML = `
      <nav id="nav"></nav>
      <div id="products"></div>
      <div id="cart"></div>
    `;
    this.modules.navigation.render();
    this.renderProducts();
    this.updateCartUI();
  },

  renderProducts: function() {
    const container = document.getElementById('products');
    if (container && this.modules.products) {
      container.innerHTML = this.modules.products.getHTML();
    }
  },

  updateCartUI: function() {
    const container = document.getElementById('cart');
    if (container && this.modules.cart) {
      container.innerHTML = this.modules.cart.getHTML();
    }
  }
};

// Auto-init
document.addEventListener('DOMContentLoaded', () => window.App.init());
```

### Módulo Cart com Estado

```javascript
// js/modules/cart.js
class Cart {
  constructor() {
    this.items = [];
    this.loadFromStorage();
  }

  add(product) {
    const existing = this.items.find(i => i.id === product.id);
    if (existing) {
      existing.qty++;
    } else {
      this.items.push({ ...product, qty: 1 });
    }
    this.saveToStorage();
  }

  remove(productId) {
    this.items = this.items.filter(i => i.id !== productId);
    this.saveToStorage();
  }

  getTotal() {
    return this.items.reduce((sum, item) => sum + (item.price * item.qty), 0);
  }

  getState() {
    return {
      items: this.items,
      total: this.getTotal()
    };
  }

  setState(state) {
    if (state.items) this.items = state.items;
    if (state.total) console.log('Total restaurado:', state.total);
  }

  saveToStorage() {
    localStorage.setItem('cart', JSON.stringify(this.items));
  }

  loadFromStorage() {
    const saved = localStorage.getItem('cart');
    if (saved) {
      try {
        this.items = JSON.parse(saved);
      } catch (e) {
        this.items = [];
      }
    }
  }

  getHTML() {
    if (this.items.length === 0) {
      return '<p>Carrinho vazio</p>';
    }
    return `
      <h3>Carrinho (${this.items.length} itens)</h3>
      <ul>
        ${this.items.map(item => `
          <li>${item.name} - Qtd: ${item.qty} - R$ ${(item.price * item.qty).toFixed(2)}</li>
        `).join('')}
      </ul>
      <p><strong>Total: R$ ${this.getTotal().toFixed(2)}</strong></p>
    `;
  }
}

window.Cart = Cart;
```

**Resultado:** Ao editar `cart.js` (lógica do carrinho), o estado (itens no carrinho) é preservado!

---

## 🎨 Exemplo 2: Temas Dinâmicos com CSS Variables

### CSS com Variáveis

```css
/* css/styles.css */

:root {
  --primary-color: #667eea;
  --secondary-color: #764ba2;
  --background-color: #f8f9fa;
  --text-color: #212529;
  --border-radius: 8px;
  --transition-speed: 0.3s;
}

body {
  background-color: var(--background-color);
  color: var(--text-color);
  font-family: 'Inter', sans-serif;
  transition: background-color var(--transition-speed) ease,
              color var(--transition-speed) ease;
}

.button {
  background-color: var(--primary-color);
  border-radius: var(--border-radius);
  padding: 12px 24px;
  transition: all var(--transition-speed) ease;
}

.button:hover {
  background-color: var(--secondary-color);
  transform: translateY(-2px);
}

.card {
  background: white;
  border-radius: var(--border-radius);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}
```

### Tema Dark/Light

```css
/* css/themes/dark.css */
:root {
  --primary-color: #7c8ff7;
  --secondary-color: #9b59b6;
  --background-color: #1a1a2e;
  --text-color: #eaeaea;
  --border-radius: 8px;
}

/* Herda tudo de styles.css, sobrescreve só o necessário */
```

```css
/* css/themes/light.css */
:root {
  --primary-color: #667eea;
  --secondary-color: #764ba2;
  --background-color: #ffffff;
  --text-color: #212529;
}
```

### JavaScript para Trocar Tema

```javascript
// js/app.js - adicione esta função
window.App = {
  // ... outras propriedades ...

  setTheme: function(themeName) {
    const link = document.getElementById('theme-stylesheet');
    if (link) {
      link.href = `css/themes/${themeName}.css`;
      TAMK_HMR.saveState('theme', themeName);
    }
  },

  init: function() {
    // Restaura tema salvo
    const savedTheme = TAMK_HMR.getState('theme') || 'dark';
    this.setTheme(savedTheme);
    
    // Botões de tema
    document.getElementById('theme-dark').onclick = () => this.setTheme('dark');
    document.getElementById('theme-light').onclick = () => this.setTheme('light');
    
    // ... resto do init ...
  }
};
```

**Hot Reload:** Edite as variáveis CSS e o tema muda instantaneamente!

---

## 🔄 Exemplo 3: Auto-Save de Formulário + HMR

### HTML

```html
<!-- index.html -->
<form id="contact-form">
  <input type="text" id="name" placeholder="Seu nome">
  <input type="email" id="email" placeholder="Seu email">
  <textarea id="message" placeholder="Mensagem"></textarea>
  <button type="submit">Enviar</button>
</form>

<script src="js/app.js"></script>
```

### JavaScript

```javascript
// js/form-handler.js
class FormHandler {
  constructor(formId) {
    this.form = document.getElementById(formId);
    this.fields = {};
    this.setupAutoSave();
    this.setupHMR();
  }

  setupAutoSave() {
    this.form.querySelectorAll('input, textarea').forEach(field => {
      const key = 'form_' + field.id;
      
      // Restore ao carregar
      const saved = TAMK_HMR.getState(key);
      if (saved !== undefined) field.value = saved;
      
      // Save ao digitar
      field.addEventListener('input', () => {
        TAMK_HMR.saveState(key, field.value);
      });
    });
  }

  setupHMR() {
    TAMK_HMR.accept((update) => {
      console.log('[Form] HMR detectado, restaurando...');
      
      // Restaura todos os campos
      this.form.querySelectorAll('input, textarea').forEach(field => {
        const key = 'form_' + field.id;
        const saved = TAMK_HMR.getState(key);
        if (saved !== undefined) field.value = saved;
      });
      
      return true;
    });
  }

  getData() {
    return {
      name: document.getElementById('name').value,
      email: document.getElementById('email').value,
      message: document.getElementById('message').value
    };
  }

  clear() {
    this.form.reset();
    this.form.querySelectorAll('input, textarea').forEach(field => {
      const key = 'form_' + field.id;
      TAMK_HMR.saveState(key, '');
    });
  }
}

window.FormHandler = FormHandler;
```

**Uso:**

```javascript
// js/app.js
document.addEventListener('DOMContentLoaded', () => {
  const contactForm = new FormHandler('contact-form');
  
  document.getElementById('submit-btn').addEventListener('click', () => {
    const data = contactForm.getData();
    console.log('Form data:', data);
    // Envia para API...
  });
});
```

**Resultado:** Se o app der reload (por HTML mudou), o formulário volta com os dados que você estava digitando!

---

## 📡 Exemplo 4: Transição de Página com Preservação de Estado

### Sistema de Roteamento Simples

```javascript
// js/router.js
class Router {
  constructor() {
    this.routes = {};
    this.currentRoute = null;
    this.stateStack = [];
    
    // Intercepta updates HMR
    TAMK_HMR.accept('js/router.js', (update) => {
      this.handleRouterUpdate();
      return true;
    });
  }

  addRoute(path, handler) {
    this.routes[path] = handler;
  }

  navigate(path) {
    // Salva estado da rota atual
    if (this.currentRoute && this.routes[this.currentRoute]) {
      const currentHandler = this.routes[this.currentRoute];
      if (currentHandler.saveState) {
        const state = currentHandler.saveState();
        this.stateStack[this.currentRoute] = state;
        TAMK_HMR.saveState('route_' + this.currentRoute, state);
      }
    }
    
    // Carrega nova rota
    this.currentRoute = path;
    if (this.routes[path]) {
      const savedState = TAMK_HMR.getState('route_' + path);
      if (savedState && this.routes[path].restoreState) {
        this.routes[path].restoreState(savedState);
      }
      this.routes[path].render();
    }
  }

  handleRouterUpdate() {
    // Restaura estado após HMR
    const saved = TAMK_HMR.getState('router_state');
    if (saved) {
      this.currentRoute = saved.currentRoute || this.currentRoute;
      this.stateStack = saved.stateStack || this.stateStack;
    }
  }

  saveState() {
    return {
      currentRoute: this.currentRoute,
      stateStack: this.stateStack
    };
  }
}

window.Router = Router;
```

### Uso em Páginas

```javascript
// js/pages/products-page.js
class ProductsPage {
  constructor() {
    this.filter = 'all';
    this.sortBy = 'name';
    this.page = 1;
  }

  render() {
    // Renderiza lista de produtos conforme filtros
    document.getElementById('page-content').innerHTML = `
      <h1>Produtos</h1>
      <div class="filters">
        <select id="filter-select">
          <option value="all" ${this.filter === 'all' ? 'selected' : ''}>Todos</option>
          <option value="available" ${this.filter === 'available' ? 'selected' : ''}>Disponíveis</option>
        </select>
        <select id="sort-select">
          <option value="name" ${this.sortBy === 'name' ? 'selected' : ''}>Nome</option>
          <option value="price" ${this.sortBy === 'price' ? 'selected' : ''}>Preço</option>
        </select>
      </div>
      <div class="products-grid">...produtos...</div>
    `;
    
    this.bindEvents();
  }

  bindEvents() {
    document.getElementById('filter-select').onchange = (e) => {
      this.filter = e.target.value;
      this.render();
    };
    document.getElementById('sort-select').onchange = (e) => {
      this.sortBy = e.target.value;
      this.render();
    };
  }

  saveState() {
    return {
      filter: this.filter,
      sortBy: this.sortBy,
      page: this.page
    };
  }

  restoreState(state) {
    if (state.filter) this.filter = state.filter;
    if (state.sortBy) this.sortBy = state.sortBy;
    if (state.page) this.page = state.page;
  }
}

window.ProductsPage = ProductsPage;
```

---

## 📊 Exemplo 5: Dashboard com Gráficos (Chart.js)

### HTML

```html
<div id="dashboard">
  <canvas id="sales-chart"></canvas>
  <canvas id="users-chart"></canvas>
</div>
```

### JavaScript com HMR

```javascript
// js/dashboard.js
class Dashboard {
  constructor() {
    this.charts = {};
    this.data = {};
    this.loadData();
  }

  async loadData() {
    const response = await fetch('data/dashboard.json');
    this.data = await response.json();
    this.renderCharts();
  }

  renderCharts() {
    this.renderSalesChart();
    this.renderUsersChart();
  }

  renderSalesChart() {
    const ctx = document.getElementById('sales-chart').getContext('2d');
    
    if (this.charts.sales) {
      this.charts.sales.destroy(); // Limpa gráfico anterior
    }
    
    this.charts.sales = new Chart(ctx, {
      type: 'line',
      data: {
        labels: this.data.sales.labels,
        datasets: [{
          label: 'Vendas',
          data: this.data.sales.values,
          borderColor: '#667eea',
          fill: true
        }]
      }
    });
  }

  renderUsersChart() {
    const ctx = document.getElementById('users-chart').getContext('2d');
    
    if (this.charts.users) {
      this.charts.users.destroy();
    }
    
    this.charts.users = new Chart(ctx, {
      type: 'bar',
      data: {
        labels: this.data.users.labels,
        datasets: [{
          label: 'Usuários',
          data: this.data.users.values,
          backgroundColor: '#764ba2'
        }]
      }
    });
  }

  // HMR handler
  static acceptHMR() {
    TAMK_HMR.accept('js/dashboard.js', (update) => {
      console.log('[Dashboard] Recarregando...');
      
      // Salva zoom/configurações atuais
      const zoomLevel = document.body.style.zoom || 1;
      
      // Eval novo código
      eval(update.newContent);
      
      // Recarrega dados
      const newDashboard = new Dashboard();
      
      // Restaura zoom
      document.body.style.zoom = zoomLevel;
      
      console.log('[Dashboard] Recarregado com sucesso!');
      return true;
    });
    
    // Handler para dados JSON
    TAMK_HMR.accept('data/dashboard.json', (update) => {
      console.log('[Dashboard] Dados atualizados');
      if (window.dashboard) {
        window.dashboard.data = JSON.parse(update.newContent);
        window.dashboard.renderCharts();
      }
      return true;
    });
  }
}

Dashboard.acceptHMR();
window.dashboard = new Dashboard();
```

**Resultado:** Mude `dashboard.json` → gráficos atualizam. Mude lógica em `dashboard.js` → gráficos recriam mantendo zoom/etc.

---

## 🎮 Exemplo 6: Jogo/App com Loop de Animação

### Canvas Game Loop

```javascript
// js/game.js
class Game {
  constructor() {
    this.canvas = document.getElementById('game-canvas');
    this.ctx = this.canvas.getContext('2d');
    this.running = false;
    this.score = 0;
    this.player = { x: 100, y: 100, size: 20 };
    this.enemies = [];
    
    this.loadState();
    this.setupHMR();
    this.setupControls();
  }

  loadState() {
    const saved = TAMK_HMR.getState('game_state');
    if (saved) {
      this.score = saved.score || 0;
      this.player = saved.player || this.player;
    }
  }

  saveState() {
    TAMK_HMR.saveState('game_state', {
      score: this.score,
      player: this.player
    });
  }

  setupHMR() {
    TAMK_HMR.accept('js/game.js', (update) => {
      console.log('[Game] Recarregando código...');
      
      // Salva estado crítico
      this.saveState();
      
      // Eval novo código
      eval(update.newContent);
      
      console.log('[Game] Código recarregado! Score mantido:', this.score);
      return true;
    });
    
    // Para assets (sprites), também aceita updates
    TAMK_HMR.accept('js/game-sprites.js', (update) => {
      console.log('[Game] Sprites atualizados');
      eval(update.newContent);
      this.reloadSprites();
      return true;
    });
  }

  start() {
    if (!this.running) {
      this.running = true;
      this.gameLoop();
    }
  }

  stop() {
    this.running = false;
  }

  gameLoop() {
    if (!this.running) return;
    
    this.update();
    this.render();
    
    requestAnimationFrame(() => this.gameLoop());
  }

  update() {
    // Atualiza posições, colisões, etc.
    // Score pode aumentar aqui
  }

  render() {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    
    // Desenha jogador
    this.ctx.fillStyle = '#667eea';
    this.ctx.fillRect(this.player.x, this.player.y, this.player.size, this.player.size);
    
    // Desenha inimigos
    this.ctx.fillStyle = '#e74c3c';
    this.enemies.forEach(enemy => {
      this.ctx.fillRect(enemy.x, enemy.y, enemy.size, enemy.size);
    });
    
    // Desenha score
    this.ctx.fillStyle = 'white';
    this.ctx.font = '16px monospace';
    this.ctx.fillText(`Score: ${this.score}`, 10, 20);
  }

  setupControls() {
    document.addEventListener('keydown', (e) => {
      const speed = 10;
      switch(e.key) {
        case 'ArrowUp': this.player.y -= speed; break;
        case 'ArrowDown': this.player.y += speed; break;
        case 'ArrowLeft': this.player.x -= speed; break;
        case 'ArrowRight': this.player.x += speed; break;
        case ' ': this.togglePause(); break;
      }
    });
  }
}

// Auto-start
window.addEventListener('load', () => {
  window.game = new Game();
  window.game.start();
});
```

**Hot Reload:** Mude lógica do jogo (física, inimigos, controles) → score e posição do jogador são mantidos!

---

## 🌐 Exemplo 7: Internacionalização (i18n) Dinâmica

### Arquivos de Tradução

```json
// data/translations/pt-BR.json
{
  "app": {
    "title": "Meu App",
    "welcome": "Bem-vindo",
    "button": {
      "submit": "Enviar",
      "cancel": "Cancelar"
    }
  }
}
```

```json
// data/translations/en-US.json
{
  "app": {
    "title": "My App",
    "welcome": "Welcome",
    "button": {
      "submit": "Submit",
      "cancel": "Cancel"
    }
  }
}
```

### Sistema de Tradução

```javascript
// js/i18n.js
class I18n {
  constructor() {
    this.currentLang = 'pt-BR';
    this.translations = {};
    this.loadTranslations();
    
    TAMK_HMR.accept('data/translations/pt-BR.json', (update) => {
      console.log('[i18n] Traduções pt-BR atualizadas');
      this.translations['pt-BR'] = JSON.parse(update.newContent);
      this.refreshUI();
      return true;
    });
    
    TAMK_HMR.accept('data/translations/en-US.json', (update) => {
      console.log('[i18n] Traduções en-US atualizadas');
      this.translations['en-US'] = JSON.parse(update.newContent);
      this.refreshUI();
      return true;
    });
  }

  async loadTranslations() {
    this.translations['pt-BR'] = await fetch('data/translations/pt-BR.json').then(r => r.json());
    this.translations['en-US'] = await fetch('data/translations/en-US.json').then(r => r.json());
  }

  t(key, lang = this.currentLang) {
    const keys = key.split('.');
    let value = this.translations[lang];
    
    for (const k of keys) {
      if (value && typeof value === 'object') {
        value = value[k];
      } else {
        return key; // fallback
      }
    }
    
    return value || key;
  }

  setLanguage(lang) {
    this.currentLang = lang;
    this.refreshUI();
  }

  refreshUI() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
      const key = el.getAttribute('data-i18n');
      el.textContent = this.t(key);
    });
  }
}

window.I18n = I18n;

// Uso no HTML:
// <h1 data-i18n="app.title"></h1>
// <button data-i18n="app.button.submit"></button>
```

**Hot Reload:** Edite JSON de tradução → interface atualiza instantaneamente!

---

## 💾 Exemplo 8: Persistência Avançada (IndexedDB + HMR)

### Wrapper IndexedDB

```javascript
// js/storage.js
class Storage {
  constructor(dbName) {
    this.dbName = dbName;
    this.db = null;
    this.init();
  }

  async init() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(this.dbName, 1);
      
      request.onupgradeneeded = (e) => {
        const db = e.target.result;
        if (!db.objectStoreNames.contains('data')) {
          db.createObjectStore('data', { keyPath: 'key' });
        }
      };
      
      request.onsuccess = (e) => {
        this.db = e.target.result;
        resolve(this.db);
      };
      
      request.onerror = (e) => reject(e);
    });
  }

  async save(key, value) {
    return new Promise((resolve, reject) => {
      const tx = this.db.transaction('data', 'readwrite');
      const store = tx.objectStore('data');
      store.put({ key, value });
      tx.oncomplete = () => resolve();
      tx.onerror = (e) => reject(e);
    });
  }

  async load(key) {
    return new Promise((resolve, reject) => {
      const tx = this.db.transaction('data', 'readonly');
      const store = tx.objectStore('data');
      const request = store.get(key);
      request.onsuccess = () => resolve(request.result?.value);
      request.onerror = (e) => reject(e);
    });
  }

  // Salva tudo antes do HMR
  saveAllForHMR(dataMap) {
    Object.entries(dataMap).forEach(([key, value]) => {
      TAMK_HMR.saveState(key, value);
    });
  }

  // Restaura tudo após HMR
  restoreAllFromHMR(keys) {
    const restored = {};
    keys.forEach(key => {
      restored[key] = TAMK_HMR.getState(key);
    });
    return restored;
  }
}

window.Storage = Storage;
```

---

## 🏆 Exemplo 9: App Completo Integrado

### index.html

```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{NAME}}</title>
  
  <!-- Config HMR ANTES de qualquer coisa -->
  <script>
    window.TAMK_CONFIG = {
      hmrEnabled: true,
      debugMode: true,
      maxReconnectAttempts: 20
    };
  </script>
  
  <!-- CSS com hot reload -->
  <link rel="stylesheet" href="css/styles.css">
  <link rel="stylesheet" href="css/themes/dark.css" id="theme-stylesheet">
</head>
<body>
  <div id="app">
    <header></header>
    <main></main>
    <footer></footer>
  </div>

  <!-- Scripts com HMR module injection -->
  <script type="module" src="js/app.js"></script>
</body>
</html>
```

### js/app.js principal

```javascript
import { Navigation } from './modules/navigation.js';
import { I18n } from './i18n.js';
import { FormHandler } from './form-handler.js';

// Inicialização única
window.addEventListener('DOMContentLoaded', async () => {
  console.log('🚀 App iniciando...');
  
  // Configura debug
  if (window.TAMK_DEV) {
    window.TAMK_DEV.hmr.debug = true;
    console.log('HMR debug ativado');
  }
  
  // Carrega módulos
  const i18n = new I18n();
  const navigation = new Navigation();
  const contactForm = new FormHandler('contact-form');
  
  // Salva tudo antes de unload
  window.addEventListener('beforeunload', () => {
    console.log('Salvando estado antes de sair...');
    TAMK_HMR.saveState('i18n_lang', i18n.currentLang);
  });
  
  // Restaura ao recarregar (HMR ou full)
  const savedLang = TAMK_HMR.getState('i18n_lang');
  if (savedLang) i18n.setLanguage(savedLang);
  
  console.log('✅ App pronto para HMR!');
});
```

---

## ✅ Checklist de Implementação HMR

Use este checklist para garantir que seu WebApp está pronto para HMR:

- [ ] **Bootstrap**: Código principal em `app.js` ou `main.js`
- [ ] **Handlers**: Cada módulo JS tem `TAMK_HMR.accept()` own handler
- [ ] **Estado**: `saveState()` chamado antes de mudanças destrutivas
- [ ] **Restauração**: `getState()` usado no handler para recuperar
- [ ] **Dispose**: `TAMK_HMR.dispose()` para módulos removidos
- [ ] **CSS Variables**: Use variáveis CSS para theming (hot swap mais fácil)
- [ ] **Module Pattern**: Separa lógica em módulos independentes
- [ ] **JSON Data**: Módulos que usam JSON têm handler para data updates
- [ ] **Fallback**: App funciona sem HMR (full reload)
- [ ] **Debug**: `TAMK_DEV.hmr.debug = true` em desenvolvimento

---

## 🔍 Debugging Tips

```javascript
// 1. Verifique conexão WebSocket
console.log('WS:', TAMK_DEV?.ws?.readyState); // 1 = OK
console.log('HMR:', TAMK_DEV?.hmr?.enabled);  // true = OK

// 2. Liste módulos registrados
console.log('Módulos:', TAMK_DEV?.listModules?.());

// 3. Force um update manual (teste)
TAMK_HMR.accept(() => {
  console.log('Manual update trigered');
  return true;
});

// 4. Limpe tudo (útil para debugging)
Object.keys(localStorage).forEach(k => localStorage.removeItem(k));
TAMK_HMR.dispose('*');

// 5. Veja o que está salvo
console.log('Keys salvas:', Object.keys(localStorage).filter(k => k.startsWith('TAMK_')));

// 6. Log detalhado
TAMK_DEV.hmr.debug = true;
```

---

<div align="center">
  <sub>Exemplos práticos HMR © 2026 T.A.M.K</sub>
</div>
