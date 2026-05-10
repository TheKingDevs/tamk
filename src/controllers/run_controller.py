import os
import re
import tempfile
import subprocess
from pathlib import Path
from typing import Optional

from utils.logger import log
from utils.banner import Banner
from utils.colors import TColor
from config.tamk_config import Config


class RunController:
    """
      Controller de execução de snippets Kotlin/Java.
    """

    SUPPORTED_EXTENSIONS = {'.kt', '.java'}
    MAX_FILE_SIZE = 1024 * 1024  # 1MB

    def __init__(self, file_path: Optional[str] = None, verbose: bool = False):
        self.file_path = file_path
        self.verbose = verbose
        self.conf = Config()
        self.cache_dir = os.path.join("assets", "cache", "run")
        self._banner = Banner()

    def _get_terminal_columns(self) -> int:
        """
          Obtém largura do terminal.
        """
        return self._banner._get_terminal_width()

    def _validate_file(self, path: str) -> bool:
        """
          Valida arquivo para execução.
        """
        if not os.path.exists(path):
            log(f"Arquivo não encontrado: {path}", "ERROR")
            return False
        
        ext = Path(path).suffix.lower()
        if ext not in self.SUPPORTED_EXTENSIONS:
            log(f"Extensão não suportada: {ext}. Use: {self.SUPPORTED_EXTENSIONS}", "ERROR")
            return False
        
        try:
            size = os.path.getsize(path)
            if size > self.MAX_FILE_SIZE:
                log(f"Arquivo muito grande: {size/1024:.0f}KB (máx: 1MB)", "ERROR")
                return False
        except OSError:
            pass
        
        abs_path = os.path.abspath(path)
        home = os.path.expanduser("~")
        if not (abs_path.startswith(home) or abs_path.startswith(os.getcwd())):
            log("Arquivo fora de diretórios permitidos", "ERROR")
            return False
        
        return True

    def _detect_project_file(self) -> Optional[str]:
        """
          Detecta arquivo principal do projeto atual.
        """
        config_path = "tamk.config"
        if not os.path.exists(config_path):
            return None
        
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            if "type=console" not in content:
                log("Comando --run sem argumentos só funciona em projetos Console", "ERROR")
                return None
            
            if os.path.exists("src/Main.kt"):
                return "src/Main.kt"
            elif os.path.exists("src/main.kt"):
                return "src/main.kt"
            else:
                for f in Path("src").glob("*.kt"):
                    return str(f)
                log("Nenhum arquivo Kotlin encontrado em src/", "ERROR")
                return None
                
        except Exception as e:
            log(f"Erro ao ler config: {e}", "ERROR")
            return None

    def execute_snippet(self) -> None:
        """
          Compila e executa arquivo Kotlin ou Java.
        """
        target_file = self.file_path

        if target_file is None or target_file is True:
            target_file = self._detect_project_file()
            if not target_file:
                log("Nenhum arquivo especificado e nenhum projeto detectado", "ERROR")
                return

        if not self._validate_file(target_file):
            return

        os.makedirs(self.cache_dir, exist_ok=True)
        
        filename = os.path.basename(target_file)
        name_only = Path(filename).stem
        ext = Path(filename).suffix.lower()

        log(f"Compilando: {filename}...", "INFO")

        if ext == '.kt':
            jar_path = os.path.join(self.cache_dir, f"{name_only}.jar")
            compile_cmd = [
                "kotlinc",
                target_file,
                "-include-runtime",
                "-d", jar_path
            ]
            run_cmd = ["java", "-jar", jar_path]
            
        elif ext == '.java':
            class_dir = self.cache_dir
            compile_cmd = [
                "javac",
                "-d", class_dir,
                target_file
            ]
            run_cmd = ["java", "-cp", class_dir, name_only]
        else:
            return

        if self.verbose:
            safe_cmd = ' '.join(shlex.quote(str(c)) for c in compile_cmd)
            log(f"Compilando: {safe_cmd}", "DEBUG")

        try:
            result = subprocess.run(
                compile_cmd,
                capture_output=not self.verbose,
                text=True,
                timeout=60
            )
            
            if result.returncode != 0:
                log("Falha na compilação", "ERROR")
                if result.stderr:
                    safe_stderr = re.sub(r'/(home|data)/[^/\s]+', '/[REDACTED]', result.stderr)
                    print(safe_stderr)
                return

        except subprocess.TimeoutExpired:
            log("Timeout na compilação (60s)", "ERROR")
            return
        except FileNotFoundError as e:
            log(f"Compilador não encontrado: {e.filename}", "ERROR")
            return

        print(self._banner._create_separator('=', TColor.CYAN))
        print(self._banner._center_text(f"{TColor.BOLD}SAÍDA DO PROGRAMA{TColor.RESET}"))
        print(self._banner._create_separator('=', TColor.CYAN))
        print()

        try:
            subprocess.run(run_cmd, timeout=300)
        except subprocess.TimeoutExpired:
            log("\n\nTimeout de execução (5 minutos)", "ERROR")
        except KeyboardInterrupt:
            print("\n\nExecução interrompida pelo usuário")
        except Exception as e:
            log(f"\nErro na execução: {e}", "ERROR")
