# 🔄 HMR - Hot Module Replacement Guide

> **Versão:** 2026.3.0-HMR  
> **Status:** ✅ Implementado

---

## 📖 Visão Geral

O **HMR (Hot Module Replacement)** do T.A.M.K permite atualizar módulos JavaScript, CSS e JSON em tempo real durante o desenvolvimento, **sem recarregar a página inteira** e **preservando o estado da aplicação**.

### ✨ Benefícios

- **Preserva estado**: Formulários preenchidos, scroll position, e dados em memória são mantidos
- **Feedback instantâneo**: Veja mudanças em < 100ms
- **Desenvolvimento fluido**: Não perca o contexto do que está desenvolvendo

---

## 🚀 Como Funciona

### Fluxo HMR

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Developer  │────▶│  T.A.M.K Dev │────▶│  WebView    │
│   Edita     │     │   Server     │     │  (App)      │
│  arquivo    │     │  (WebSocket) │     │             │
└─────────────┘     └──────────────┘     └─────────────┘
                           │                    │
                           │ 1. Detecta mudança │
                           │ 2. Processa HMR    │
                           │ 3. Envia update    │
                           │                    │
                           │              4. Aplica módulo
                           │              5. Preserva estado
                           │              6. Executa handler
```

### Tipos de Arquivo e Comportamento

| Extensão | Comportamento | Reload | State |
|----------|--------------|--------|-------|
| `.css` | Hot Reload | ❌ Não | ✅ Preserva |
| `.js` | HMR Module Injection | ❌ Não | ✅ Preserva |
| `.json` | Data Update | ❌ Não | ✅ Preserva |
| `.html` | Rebuild + Reload | ✅ Sim | ⚠️ Parcial |
| `.png, .jpg` | Rebuild + Reload | ✅ Sim | ⚠️ Parcial |

---

## 🛠️ Usando o HMR

### 1. Inicie o Modo Dev

```bash
cd seu-projeto-webapp
tamk --dev
```

### 2. Abra o App no Dispositivo

O app deve estar instalado e aberto. Ele se conectará automaticamente ao servidor WebSocket.

### 3. Edite Arquivos

Edite qualquer arquivo em `src/main/assets/`:

- **CSS**: Atualização instantânea sem reload
- **JS**: Injeta novo código sem perder estado
- **JSON**: Atualiza dados em tempo real

---

## 📚 HMR API

O T.A.M.K expõe uma API global para controle fino do HMR.

### `TAMK_HMR.accept(handler)`

Registra um handler para updates de módulos específicos.

```javascript
// Aceita updates de qualquer módulo
TAMK_HMR.accept((update) => {
    console.log('Módulo atualizado:', update.modulePath);
    // Retorna false para forçar full reload
    return true;
});

// Aceita updates de módulo específico
TAMK_HMR.accept('js/app.js', (update) => {
    console.log('app.js atualizado!');
    // Re-inicializa apenas este módulo
    if (window.myApp && typeof window.myApp.reload === 'function') {
        window.myApp.reload();
    }
});
```

### `TAMK_HMR.saveState(key, value)`

Salva estado para recuperação após update.

```javascript
// Salva estado antes de um update
TAMK_HMR.saveState('scrollPosition', window.scrollY);
TAMK_HMR.saveState('formData', {
    name: document.getElementById('name').value,
    email: document.getElementById('email').value
});
```

### `TAMK_HMR.getState(key)`

Recupera estado salvo.

```javascript
// Recupera estado após update
var scroll = TAMK_HMR.getState('scrollPosition');
if (scroll !== undefined) {
    window.scrollTo(0, parseInt(scroll));
}
```

### `TAMK_HMR.dispose(modulePath)`

Limpa handlers e módulos registrados.

```javascript
// Cleanup de módulo
TAMK_HMR.dispose('js/old-module.js');
```

---

## 💡 Exemplos Práticos

### Exemplo 1: Preservando Estado de Formulário

```javascript
// Salva estado automaticamente
var formFields = document.querySelectorAll('input, textarea');
formFields.forEach(function(field) {
    field.addEventListener('input', function() {
        var state = {};
        formFields.forEach(function(f) {
            state[f.id] = f.value;
        });
        TAMK_HMR.saveState('formState', state);
    });
});

// Restaura após HMR
TAMK_HMR.accept(function(update) {
    var savedState = TAMK_HMR.getState('formState');
    if (savedState) {
        for (var id in savedState) {
            var el = document.getElementById(id);
            if (el) el.value = savedState[id];
        }
    }
});
```

### Exemplo 2: Hot Reload de Componente

```javascript
// Componente com HMR
var CounterComponent = {
    count: 0,
    
    init: function() {
        this.count = 0;
        this.render();
    },
    
    render: function() {
        document.getElementById('count').textContent = this.count;
    },
    
    increment: function() {
        this.count++;
        this.render();
    }
};

// Registra handler HMR
TAMK_HMR.accept('js/counter.js', function(update) {
    // Salva estado
    var savedCount = CounterComponent.count;
    
    // Eval do novo código
    eval(update.newContent);
    
    // Restaura estado
    CounterComponent.count = savedCount;
    CounterComponent.render();
    
    console.log('Componente atualizado com estado preservado!');
});
```

### Exemplo 3: Application State Management

```javascript
// Gerenciador de estado global
window.AppState = {
    user: null,
    settings: {},
    cache: {},
    
    // Salva estado antes de HMR
    saveForHMR: function() {
        TAMK_HMR.saveState('user', this.user);
        TAMK_HMR.saveState('settings', this.settings);
        TAMK_HMR.saveState('cache', this.cache);
    },
    
    // Restaura estado após HMR
    restoreFromHMR: function() {
        this.user = TAMK_HMR.getState('user') || null;
        this.settings = TAMK_HMR.getState('settings') || {};
        this.cache = TAMK_HMR.getState('cache') || {};
    }
};

// Auto-save antes de unload
window.addEventListener('beforeunload', function() {
    window.AppState.saveForHMR();
});

// Auto-restore no init
document.addEventListener('DOMContentLoaded', function() {
    window.AppState.restoreFromHMR();
});
```

---

## 🔧 Configuração

### No index.html

O bridge é injetado automaticamente, mas você pode configurar:

```html
<script>
// Configurações opcionais antes do bridge carregar
window.TAMK_CONFIG = {
    hmrEnabled: true,
    debugMode: true,
    maxReconnectAttempts: 15
};
</script>
```

### Debug Mode

Para ver logs detalhados do HMR:

```javascript
// No console do app
TAMK_DEV.hmr.debug = true;
```

---

## 🐛 Troubleshooting

### HMR não está funcionando

1. **Verifique se o app está conectado:**
   ```javascript
   console.log(TAMK_DEV.ws.readyState);
   // 1 = aberto, outros = fechado/conectando
   ```

2. **Verifique se HMR está habilitado:**
   ```javascript
   console.log(TAMK_DEV.hmr.enabled);
   ```

3. **Lista módulos registrados:**
   ```javascript
   console.log(TAMK_DEV.listModules());
   ```

### Estado não está sendo preservado

1. **Salve estado explicitamente:**
   ```javascript
   TAMK_HMR.saveState('myKey', myValue);
   ```

2. **Use handlers accept:**
   ```javascript
   TAMK_HMR.accept(function(update) {
       // Seu código de restore aqui
   });
   ```

### Módulo JS não atualiza

- Verifique se o arquivo está em `src/main/assets/`
- Verifique se a extensão é `.js`
- Tente forçar reload com `Ctrl+C` e `tamk --dev` novamente

---

## 📊 Status Command

Durante o modo dev, use `s` para ver status:

```
Uptime: 120s | Builds: 5 | Clientes: 1 (HMR: 1) | Módulos: 12
```

- **Clientes**: Apps conectados
- **HMR**: Clientes com HMR habilitado
- **Módulos**: Módulos JS registrados no registry

---

## 🎯 Melhores Práticas

1. **Sempre use handlers accept** para módulos críticos
2. **Salve estado importante** antes de operações que podem trigger reload
3. **Teste o fallback** - assegure que seu app funciona mesmo após full reload
4. **Não dependa de estado volátil** - o HMR é para desenvolvimento, não produção

---

## 🔮 Futuro

Próximas melhorias planejadas:

- [ ] Suporte a TypeScript HMR
- [ ] Module dependency graph
- [ ] Partial page updates
- [ ] React/Vue integration

---

**Veja também:**
- [DEV_GUIDE.md](DEV_GUIDE.md) - Guia completo de desenvolvimento
- [FAQ.md](FAQ.md) - Perguntas frequentes
