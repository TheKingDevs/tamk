import os
import re
import sys
import time
import json
import shutil
import signal
import threading
import subprocess
from pathlib import Path
from getpass import getpass
from typing import Optional, Callable, Set, Dict
from http.server import HTTPServer, SimpleHTTPRequestHandler

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from utils.logger import log
from utils.colors import TColor
from utils.banner import banner, banner_box
from config.tamk_config import Config

try:
    import asyncio
    import websockets
    import websockets.server
    WEBSOCKETS_AVAILABLE = True
except ImportError:
    WEBSOCKETS_AVAILABLE = False


class DevServer:
    """
      Servidor WebSocket para comunicação com WebView em modo dev.
    """

    def __init__(self, port: int = 8765):
        self.port = port
        self.clients: Set = set()
        self.server = None
        self.loop: Optional[asyncio.AbstractEventLoop] = None
        self.thread: Optional[threading.Thread] = None
        self._running = False
        self._lock = threading.Lock()        
        self.module_registry: Dict = {}
        self.client_states: Dict = {}

    async def handler(self, websocket):
        """
          Gerencia conexões WebSocket com suporte a HMR.
        """
        client_addr = websocket.remote_address if hasattr(websocket, 'remote_address') else "unknown"
        log(f"📱 Cliente conectado: {client_addr}", "INFO")

        with self._lock:
            self.clients.add(websocket)
            self.client_states[websocket] = {
                'modules': [],
                'hmr_enabled': False,
                'scroll_position': 0,
                'connected_at': time.time()
            }

        try:
            async for message in websocket:
                try:
                    data = json.loads(message)
                    msg_type = data.get('type')
                    
                    if msg_type == 'pong':
                        pass
                    elif msg_type == 'error':
                        log(f"Erro no cliente: {data.get('message')}", "ERROR")
                    elif msg_type == 'hello':
                        # Cliente enviou handshake inicial
                        client_state = self.client_states.get(websocket, {})
                        client_state['hmr_enabled'] = data.get('hmrEnabled', False)
                        client_state['modules'] = data.get('loadedModules', [])
                        client_state['scroll_position'] = data.get('scrollPosition', 0)
                        client_state['version'] = data.get('version', 'unknown')
                        
                        log(f"👋 Hello from client (HMR: {client_state['hmr_enabled']}, Modules: {len(client_state['modules'])})", "INFO")
                        
                        # Envia confirmação
                        await websocket.send(json.dumps({
                            'type': 'hello-ack',
                            'hmrSupported': True,
                            'serverVersion': self.conf.VERSION
                        }))
                    elif msg_type == 'ping':
                        await websocket.send(json.dumps({
                            'type': 'pong',
                            'timestamp': data.get('timestamp'),
                            'serverTime': time.time()
                        }))
                    elif msg_type == 'request-reload':
                        log("🔄 Reload solicitado pelo cliente", "INFO")
                        self.broadcast({'type': 'reload', 'timestamp': time.time()})
                        
                except json.JSONDecodeError:
                    log(f"Mensagem inválida: {message[:100]}", "WARNING")

        except websockets.exceptions.ConnectionClosed:
            pass
        except Exception as e:
            log(f"Erro no handler WebSocket: {e}", "WARNING")
        finally:
            with self._lock:
                self.clients.discard(websocket)
                self.client_states.pop(websocket, None)
            log("📱 Cliente desconectado", "INFO")

    def register_module(self, path: str, content: str) -> None:
        """
          Registra um módulo no registry HMR.
        """
        import hashlib
        self.module_registry[path] = {
            'content': content,
            'hash': hashlib.md5(content.encode()).hexdigest(),
            'timestamp': time.time()
        }

    def get_changed_modules(self, file_paths: list) -> list:
        """
          Compara arquivos modificados com o registry e retorna lista de módulos changed.
        """
        changed = []
        for file_path in file_paths:
            if not file_path.endswith('.js'):
                continue
                
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                import hashlib
                current_hash = hashlib.md5(content.encode()).hexdigest()
                registered = self.module_registry.get(file_path, {})
                
                if registered.get('hash') != current_hash:
                    rel_path = os.path.relpath(file_path, self.assets_dir)
                    changed.append({
                        'path': rel_path,
                        'content': content,
                        'hash': current_hash,
                        'timestamp': time.time()
                    })
                    self.register_module(file_path, content)
                    
            except Exception as e:
                log(f"Erro ao ler módulo {file_path}: {e}", "WARNING")
                
        return changed

    def broadcast_hmr_update(self, file_path: str) -> None:
        """
          Envia update HMR para todos os clientes.
        """
        if not self.clients or not self.loop:
            return
            
        ext = Path(file_path).suffix.lower()
        
        if ext == '.css':
            self.broadcast({
                'type': 'css-update',
                'files': [os.path.basename(file_path)],
                'timestamp': time.time()
            })
            log("🎨 CSS hot-reload enviado", "INFO")
            
        elif ext == '.js':
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                rel_path = os.path.relpath(file_path, self.assets_dir)
                
                self.register_module(file_path, content)
                                
                self.broadcast({
                    'type': 'js-hmr',
                    'modules': [{
                        'path': rel_path,
                        'content': content,
                        'timestamp': time.time()
                    }],
                    'timestamp': time.time()
                })
                log(f"📦 JS HMR update enviado: {rel_path}", "INFO")
                
            except Exception as e:
                log(f"Erro ao enviar HMR update: {e}", "ERROR")
                
        elif ext == '.json':
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                rel_path = os.path.relpath(file_path, self.assets_dir)
                
                self.broadcast({
                    'type': 'json-update',
                    'file': rel_path,
                    'content': content,
                    'timestamp': time.time()
                })
                log(f"📄 JSON update enviado: {rel_path}", "INFO")
                
            except Exception as e:
                log(f"Erro ao enviar JSON update: {e}", "ERROR")
                
        else:
            self.broadcast({
                'type': 'reload',
                'timestamp': time.time(),
                'reason': f'File changed: {os.path.basename(file_path)}'
            })
            log(f"🔄 Full reload enviado (arquivo: {os.path.basename(file_path)})", "INFO")

    def broadcast(self, message: Dict) -> None:
        """
          Envia mensagem para todos os clientes conectados.
        """
        with self._lock:
            clients_copy = set(self.clients)
            loop = self.loop

        if not clients_copy or not loop:
            return

        async def send():
            disconnected = set()

            for client in clients_copy:
                try:
                    if client.open:
                        await client.send(json.dumps(message))
                    else:
                        disconnected.add(client)
                except Exception:
                    disconnected.add(client)

            with self._lock:
                self.clients -= disconnected

        if loop and loop.is_running():
            asyncio.run_coroutine_threadsafe(send(), loop)
    
    def start(self) -> bool:
        """
          Inicia servidor em thread separada com loop asyncio próprio.
        """
        if not WEBSOCKETS_AVAILABLE:
            return False
        
        def run_server():
            self.loop = asyncio.new_event_loop()
            asyncio.set_event_loop(self.loop)

            try:
                start_server = websockets.server.serve(
                    self.handler,
                    "localhost",
                    self.port,
                    ping_interval=10,
                    ping_timeout=5,
                    close_timeout=5
                )

                self.server = self.loop.run_until_complete(start_server)
                self._running = True
                log(f"🌐 WebSocket server na porta {self.port}", "SUCCESS")

                self.loop.run_forever()

            except Exception as e:
                log(f"Erro no servidor WebSocket: {e}", "ERROR")
                self._running = False
            
            finally:
                if self.loop and not self.loop.is_closed():
                    try:
                        pending = asyncio.all_tasks(self.loop)
                        if pending:
                            for task in pending:
                                task.cancel()
                            try:
                                self.loop.run_until_complete(
                                    asyncio.gather(*pending, return_exceptions=True)
                                )
                            except Exception:
                                pass
                        self.loop.close()
                    except Exception:
                        pass
        
        self.thread = threading.Thread(target=run_server, daemon=True)
        self.thread.start()
        
        timeout = 5
        start_time = time.time()
        while not self._running and time.time() - start_time < timeout:
            time.sleep(0.1)
        
        return self._running
    
    def stop(self) -> None:
        """Para o servidor graciosamente."""
        self._running = False

        if self.clients:
            for client in list(self.clients):
                try:
                    if self.loop and self.loop.is_running():
                        coro = client.close()
                        future = asyncio.run_coroutine_threadsafe(coro, self.loop)
                        try:
                            future.result(timeout=1)
                        except Exception:
                            pass
                except Exception:
                    pass
            self.clients.clear()

        if self.server and self.loop and self.loop.is_running():
            try:
                coro = self.server.close()
                future = asyncio.run_coroutine_threadsafe(coro, self.loop)
                try:
                    future.result(timeout=2)
                except Exception:
                    pass
            except Exception:
                pass

        if self.loop and self.loop.is_running():
            try:
                self.loop.call_soon_threadsafe(self.loop.stop)
            except Exception:
                pass

        if self.thread and self.thread.is_alive():
            self.thread.join(timeout=3)

        if self.loop and not self.loop.is_closed():
            try:
                pending = asyncio.all_tasks(self.loop)
                if pending:
                    for task in pending:
                        task.cancel()
                    try:
                        self.loop.run_until_complete(
                            asyncio.gather(*pending, return_exceptions=True)
                        )
                    except Exception:
                        pass
                self.loop.close()
            except Exception:
                pass
        
        self.loop = None
        self.server = None


class SimpleHTTPServerThread:
    """Servidor HTTP simples como fallback."""
    
    def __init__(self, root_dir: str, port: int = 8080):
        self.root_dir = os.path.abspath(root_dir)
        self.port = port
        self.server: Optional[HTTPServer] = None
        self.thread: Optional[threading.Thread] = None
        
    def start(self) -> None:
        """Inicia servidor HTTP."""
        
        class Handler(SimpleHTTPRequestHandler):
            def __init__(self, *args, **kwargs):
                self.root = self.root_dir 
                super().__init__(*args, directory=self.root, **kwargs)
            
            def log_message(self, format: str, *args):
                pass
        
        def make_handler():
            def handler(*args, **kwargs):
                return Handler(*args, directory=self.root_dir, **kwargs)
            return handler
        
        def serve():
            try:
                handler_class = make_handler()
                self.server = HTTPServer(('localhost', self.port), handler_class)
                self.server.serve_forever()
            except Exception as e:
                log(f"Erro HTTP server: {e}", "ERROR")
        
        self.thread = threading.Thread(target=serve, daemon=True)
        self.thread.start()
        log(f"🌐 HTTP server na porta {self.port}", "SUCCESS")
    
    def stop(self) -> None:
        if self.server:
            self.server.shutdown()
            self.server.server_close()


class DevController:
    """
      Controlador do modo desenvolvimento com live reload.
    """
    
    WATCH_EXTENSIONS: Set[str] = {
        '.html', '.css', '.js', '.json', 
        '.png', '.jpg', '.jpeg', '.svg', '.webp',
        '.xml', '.kt'
    }
    
    IGNORE_PATTERNS: Set[str] = {
        '__pycache__', '.git', 'node_modules',
        'assets/cache', 'secret', '.idea', '.vscode',
        'build', 'dist', '.gradle', 'tamk_dev_'
    }

    def __init__(self, password: Optional[str] = None, verbose: bool = False, no_ws: bool = False, ws_port: int = 8765):
        self.conf = Config()
        self.password = password
        self.verbose = verbose
        self.no_ws = no_ws
        self.ws_port = ws_port
        
        self.root_dir = os.path.abspath(os.getcwd())
        self.assets_dir = os.path.join(self.root_dir, "src", "main", "assets")
        self.cache_dir = os.path.join(self.root_dir, "assets", "cache")
        
        self.ws_server: Optional[DevServer] = None
        self.http_server: Optional[SimpleHTTPServerThread] = None
        self.watcher: Optional['FileWatcher'] = None
        
        self.last_build_time = 0.0
        self.build_cooldown = 1.0
        self.build_count = 0
        self.start_time = time.time()
        self._running = False
        
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)

    def _signal_handler(self, signum, frame):
        log(f"\nSinal {signum} recebido. Encerrando...", "INFO")
        try:
            self.stop()
        except Exception as e:
            log(f"Erro ao encerrar: {e}", "WARNING")

    def _is_webapp_project(self) -> bool:
        config_path = os.path.join(self.root_dir, "tamk.config")
        if not os.path.exists(config_path):
            return False
        
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                content = f.read().lower()
                return "type=webapp" in content
        except Exception:
            return False

    def _check_prerequisites(self) -> bool:
        if not self._is_webapp_project():
            log("Este comando só funciona em projetos WebApp!", "ERROR")
            return False
        
        if not os.path.exists(self.assets_dir):
            log(f"Pasta assets não encontrada: {self.assets_dir}", "ERROR")
            return False
        
        if not os.path.exists(os.path.join(self.root_dir, "app-final.apk")):
            if not os.path.exists(os.path.join(self.root_dir, "app-dev.apk")):
                log("APK base não encontrado. Execute 'tamk --build' primeiro.", "ERROR")
                return False
        
        return True

    def _should_process_file(self, file_path: str) -> bool:
        for pattern in self.IGNORE_PATTERNS:
            if pattern in file_path:
                return False
        
        ext = Path(file_path).suffix.lower()
        return ext in self.WATCH_EXTENSIONS

    def _get_cached_password(self) -> str:
        """
          Obtém senha do cache ou solicita ao usuário.
        """
        if self.password:
            return self.password
        
        self.password = getpass("Senha da Keystore: ")
        return self.password

    def _quick_assets_build(self) -> bool:
        """
          Build rápido apenas dos assets.
        """
        current_time = time.time()
        if current_time - self.last_build_time < self.build_cooldown:
            return False

        self.last_build_time = current_time
        self.build_count += 1

        log(f"🔄 Rebuild #{self.build_count}...", "STEP")

        try:
            from controllers.build_controller import BuildController

            pwd = self._get_cached_password()

            builder = BuildController(verbose=self.verbose, password=pwd)
            success = builder.build_assets_only(pwd)

            if success and self.ws_server:
                self.ws_server.broadcast({
                    "type": "reload",
                    "timestamp": time.time(),
                    "build": self.build_count,
                    "reason": "Assets rebuild"
                })

            elapsed = time.time() - self.start_time
            log(f"✅ Build #{self.build_count} em {elapsed:.1f}s", "SUCCESS")
            return success

        except Exception as e:
            log(f"Erro no build: {str(e)}", "ERROR")
            return False

    def _on_file_changed(self, file_path: str) -> None:
        if not self._should_process_file(file_path):
            return

        rel_path = os.path.relpath(file_path, self.root_dir)
        log(f"📝 Modificado: {rel_path}", "INFO")

        ext = Path(file_path).suffix.lower()

        if self.ws_server and self.ws_server.clients:
            if ext in ('.css', '.js', '.json'):
                self.ws_server.broadcast_hmr_update(file_path)
                return
            elif ext in ('.html', '.png', '.jpg', '.jpeg', '.svg', '.webp'):
                log(f"📄 Arquivo {ext} detectado, rebuild necessário", "INFO")
        
        self._quick_assets_build()

    def _inject_dev_bridge(self) -> None:
        """
          Cria arquivo dev-bridge.js na pasta js/ e adiciona referência no index.html.
        """
        index_path = os.path.join(self.assets_dir, "index.html")
        if not os.path.exists(index_path):
            log("index.html não encontrado", "WARNING")
            return

        js_dir = os.path.join(self.assets_dir, "js")
        os.makedirs(js_dir, exist_ok=True)
        bridge_js_path = os.path.join(js_dir, "tamk-dev-bridge.js")

        try:
            tmpl_dir = self.conf.get_template_dir("webapp")
            bridge_tmpl_path = os.path.join(tmpl_dir, "dev_bridge.js.tmpl")

            if not os.path.exists(bridge_tmpl_path):
                log(f"Template dev_bridge.js.tmpl não encontrado em {bridge_tmpl_path}", "WARNING")
                return

            with open(bridge_tmpl_path, 'r', encoding='utf-8') as f:
                bridge_content = f.read()

            bridge_content = bridge_content.replace('{{DEV_PORT}}', str(self.ws_port))
            bridge_content = bridge_content.replace('{{VERSION}}', self.conf.VERSION)

            with open(bridge_js_path, 'w', encoding='utf-8') as f:
                f.write(bridge_content)

            log(f"🔌 Dev bridge criado: js/tamk-dev-bridge.js (porta {self.ws_port})", "INFO")

            with open(index_path, 'r', encoding='utf-8') as f:
                content = f.read()

            if 'tamk-dev-bridge.js' in content:
                return

            backup_path = index_path + '.tamk_backup'
            if not os.path.exists(backup_path):
                shutil.copy2(index_path, backup_path)
                log("📦 Backup criado: index.html.tamk_backup", "INFO")

            script_tag = f'  <!-- TAMK Dev Bridge -->\n<script src="js/tamk-dev-bridge.js"></script>\n</body>'
            
            if '</body>' in content:
                content = content.replace('</body>', script_tag)
            else:
                content += f'\n<!-- TAMK Dev Bridge -->\n<script src="js/tamk-dev-bridge.js"></script>\n'

            with open(index_path, 'w', encoding='utf-8') as f:
                f.write(content)

            log("✅ Bridge injetado no index.html", "INFO")

        except Exception as e:
            log(f"Erro ao injetar bridge: {e}", "WARNING")

    def _restore_backup(self) -> None:
        """
          Restaura index.html apenas se usuário não tiver editado.
          Compara com backup para detectar mudanças.
        """
        index_path = os.path.join(self.assets_dir, "index.html")
        backup_path = index_path + '.tamk_backup'

        if not os.path.exists(backup_path):
            return

        try:
            with open(index_path, 'r', encoding='utf-8') as f:
                current = f.read()
            with open(backup_path, 'r', encoding='utf-8') as f:
                original = f.read()

            if 'tamk-dev-bridge.js' in current:
                current = current.replace('  <!-- TAMK Dev Bridge -->\n<script src="js/tamk-dev-bridge.js"></script>\n</body>', '</body>')
                current = current.replace('<!-- TAMK Dev Bridge -->\n<script src="js/tamk-dev-bridge.js"></script>\n', '')
                
                if current.strip() == original.strip():
                    log("📄 index.html limpo (apenas bridge removido)", "INFO")
                else:
                    log("✨ Edições do usuário preservadas", "INFO")
                
                with open(index_path, 'w', encoding='utf-8') as f:
                    f.write(current)

            os.remove(backup_path)
            log("🧹 Backup removido", "INFO")

        except Exception as e:
            log(f"Erro ao restaurar: {e}", "WARNING")

    def _show_instructions(self) -> None:
        """Exibe instruções do modo dev."""
        banner_box(
            lines=[
                f"{TColor.BOLD}✨ RECURSOS HMR:{TColor.RESET}",
                f"  CSS → Atualiza sem reload (preserva estado)",
                f"  JS → Injeta módulos sem reload",
                f"  JSON → Atualiza dados sem reload",
                f"  HTML/Imagens → Rebuild + reload automático",
                "",
                f"{TColor.BOLD}📦 SERVIDORES:{TColor.RESET}",
                f"  WebSocket: ws://localhost:8765",
                f"  HTTP: http://localhost:8080 (fallback)",
                "",
                f"{TColor.BOLD}⌨️ COMANDOS:{TColor.RESET}",
                f"  Ctrl+C → Encerrar | b → Rebuild | i → Install",
                f"  h → Ajuda | s → Status",
            ],
            style="info",
            title="🚀 MODO DESENVOLVIMENTO COM HMR"
        )

    def _handle_input(self) -> None:
        """Thread para comandos do teclado."""
        while self._running:
            try:
                cmd = input().strip().lower()
                
                if cmd in ('q', 'quit', 'exit'):
                    break
                elif cmd == 'b':
                    log("🔨 Forçando rebuild...", "STEP")
                    self._quick_assets_build()
                elif cmd == 'i':
                    log("📦 Instalando APK...", "STEP")
                    subprocess.run(["tamk", "--install"], capture_output=True)
                elif cmd == 'h':
                    self._show_instructions()
                elif cmd == 's':
                    uptime = time.time() - self.start_time
                    clients = len(self.ws_server.clients) if self.ws_server else 0
                    modules = len(self.ws_server.module_registry) if self.ws_server else 0
                    hmr_clients = sum(1 for c in self.ws_server.client_states.values() if c.get('hmr_enabled')) if self.ws_server else 0
                    log(f"Uptime: {uptime:.0f}s | Builds: {self.build_count} | Clientes: {clients} (HMR: {hmr_clients}) | Módulos: {modules}", "INFO")

            except EOFError:
                break
            except KeyboardInterrupt:
                break

    def start(self) -> bool:
        """
          Inicia modo desenvolvimento.
        """
        if not self._check_prerequisites():
            return False
        
        os.makedirs(self.cache_dir, exist_ok=True)
        
        self._inject_dev_bridge()
        
        if WEBSOCKETS_AVAILABLE and not self.no_ws:
            self.ws_server = DevServer(port=8765)
            if self.ws_server.start():
                log("✅ WebSocket server iniciado", "SUCCESS")
            else:
                log("⚠️  WebSocket falhou, usando HTTP", "WARNING")
                self.http_server = SimpleHTTPServerThread(self.assets_dir, port=8080)
                self.http_server.start()
        else:
            self.http_server = SimpleHTTPServerThread(self.assets_dir, port=8080)
            self.http_server.start()
        
        try:
            from utils.watcher import FileWatcher
            self.watcher = FileWatcher(self.assets_dir, self._on_file_changed)
            if not self.watcher.start():
                log("Falha no watcher", "ERROR")
                return False
        except ImportError:
            log("watchdog não instalado", "ERROR")
            return False
        
        self._running = True
        self._show_instructions()
        
        log("🔨 Build inicial...", "STEP")
        self._quick_assets_build()
        
        try:
            self._handle_input()
        except KeyboardInterrupt:
            pass
        finally:
            self.stop()
        
        return True

    def stop(self) -> None:
        """
          Encerra modo dev.
        """
        if not self._running:
            return

        self._running = False
        log("Encerrando modo desenvolvimento...", "STEP")

        if self.watcher:
            self.watcher.stop()

        if self.ws_server:
            self.ws_server.stop()

        if self.http_server:
            self.http_server.stop()

        bridge_js_path = os.path.join(self.assets_dir, "js", "tamk-dev-bridge.js")
        if os.path.exists(bridge_js_path):
            try:
                os.remove(bridge_js_path)
                log("🧹 tamk-dev-bridge.js removido", "INFO")
            except Exception as e:
                log(f"Erro ao remover bridge: {e}", "WARNING")

        self._restore_backup()

        log("👋 Modo dev encerrado", "SUCCESS")


def start_dev_mode(password: Optional[str] = None, verbose: bool = False, no_ws: bool = False, ws_port: int = 8765) -> bool:
    """
      Entry point.
    """
    controller = DevController(password=password, verbose=verbose, no_ws=no_ws, ws_port=ws_port)
    return controller.start()


if __name__ == "__main__":
    start_dev_mode()
