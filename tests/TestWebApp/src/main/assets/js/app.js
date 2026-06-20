/**
 * T.A.M.K WebApp - JavaScript Principal
 * Projeto: TestWebApp
 * Versão: 0.1.0
 */

(function() {
  'use strict';

  // Inicialização do App
  console.log('[TestWebApp] App inicializado');
  console.log('[TestWebApp] Versão: 0.1.0');

  // Detecta se está em modo desenvolvimento
  const isDevMode = window.location.hostname === 'localhost' || 
                    window.location.hostname === '127.0.0.1';

  if (isDevMode) {
    console.log('[TestWebApp] Modo desenvolvimento detectado');
  }

  // Funções utilitárias globais
  window.App = {
    name: 'TestWebApp',
    version: '0.1.0',
    author: 'Dev',

    /**
     * Exibe informações do app no console
     */
    info: function() {
      console.group('📱 TestWebApp');
      console.log('Versão: 0.1.0');
      console.log('Autor: Dev');
      console.log('Package: com.dev.testwebapp');
      console.log('Desenvolvido com T.A.M.K v2026.3.0-HMR');
      console.groupEnd();
    },

    /**
     * Navega para uma URL interna
     * @param {string} path - Caminho relativo
     */
    navigate: function(path) {
      console.log('[Navigation]', path);
      // Implementar navegação customizada se necessário
    },

    /**
     * Carrega dados de forma assíncrona
     * @param {string} url - URL para fetch
     * @returns {Promise<any>}
     */
    fetch: async function(url) {
      try {
        const response = await window.fetch(url);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return await response.json();
      } catch (error) {
        console.error('[Fetch Error]', error);
        throw error;
      }
    }
  };

  // Auto-executa informações no console
  window.App.info();

})();
