import os
import time
import threading
from pathlib import Path
from typing import Callable, Set, Optional

try:
    from watchdog.observers import Observer
    from watchdog.events import FileSystemEventHandler, FileModifiedEvent, FileCreatedEvent
    WATCHDOG_AVAILABLE = True
except ImportError:
    WATCHDOG_AVAILABLE = False

from utils.logger import log


class WebAppFileHandler(FileSystemEventHandler):
    """
      Handler que detecta mudanças em arquivos web.
    """
    
    def __init__(self, callback: Callable[[str], None], debounce_seconds: float = 0.5):
        self.callback = callback
        self.debounce_seconds = debounce_seconds
        self._timers: dict = {}
        self._lock = threading.Lock()
        
        self.watch_extensions: Set[str] = {
            '.html', '.css', '.js', '.json',
            '.png', '.jpg', '.jpeg', '.svg', '.webp'
        }
        self.ignore_patterns: Set[str] = {
            '__pycache__', '.git', 'node_modules', 
            'assets/cache', 'secret', '.idea'
        }
        
    def _should_process(self, path: str) -> bool:
        """
          Verifica se arquivo deve ser processado.
        """
        for pattern in self.ignore_patterns:
            if pattern in path:
                return False
        
        ext = Path(path).suffix.lower()
        return ext in self.watch_extensions
    
    def _debounced_callback(self, event_path: str) -> None:
        """
          Executa callback com debounce.
        """
        with self._lock:
            if event_path in self._timers:
                self._timers[event_path].cancel()
            
            def delayed():
                with self._lock:
                    self._timers.pop(event_path, None)
                self.callback(event_path)
            
            timer = threading.Timer(self.debounce_seconds, delayed)
            self._timers[event_path] = timer
            timer.start()
    
    def on_modified(self, event):
        if isinstance(event, FileModifiedEvent) and not event.is_directory:
            if self._should_process(event.src_path):
                self._debounced_callback(event.src_path)
    
    def on_created(self, event):
        if isinstance(event, FileCreatedEvent) and not event.is_directory:
            if self._should_process(event.src_path):
                self._debounced_callback(event.src_path)


class FileWatcher:
    """
      Gerenciador do observador de arquivos.
    """
    
    def __init__(self, watch_path: str, on_change_callback: Callable[[str], None]):
        self.watch_path = os.path.abspath(watch_path)
        self.on_change = on_change_callback
        self.observer: Optional[Observer] = None
        self._running = False
        
    def start(self) -> bool:
        """
          Inicia monitoramento.
        """
        if not WATCHDOG_AVAILABLE:
            log("watchdog não instalado", "ERROR")
            return False
            
        if not os.path.exists(self.watch_path):
            log(f"Pasta não existe: {self.watch_path}", "ERROR")
            return False
        
        try:
            event_handler = WebAppFileHandler(self.on_change)
            self.observer = Observer()
            self.observer.schedule(event_handler, self.watch_path, recursive=True)
            self.observer.start()
            self._running = True
            
            log(f"👁️  Monitorando: {self.watch_path}", "INFO")
            return True
            
        except Exception as e:
            log(f"Erro ao iniciar watcher: {e}", "ERROR")
            return False
    
    def stop(self) -> None:
        """
          Para monitoramento e limpa recursos.
        """
        if self.observer and self._running:
            self.observer.stop()
            self.observer.join(timeout=2)
            self._running = False
            log("Monitoramento encerrado", "INFO")
    
    def is_running(self) -> bool:
        return self._running
