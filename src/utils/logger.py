import sys
from datetime import datetime
from typing import Optional, TextIO


class Colors:
    """
      Códigos ANSI para colorização.
    """
    HEADER = '\033[95m'
    INFO = '\033[94m'
    SUCCESS = '\033[92m'
    WARNING = '\033[93m'
    ERROR = '\033[91m'
    END = '\033[0m'
    BOLD = '\033[1m'
    GRAY = '\033[90m'


class Logger:
    """
      Logger estruturado com níveis e saída configurável.
    """
          
    LEVELS = {
        'DEBUG': 10,
        'INFO': 20,
        'STEP': 25,
        'SUCCESS': 30,
        'WARNING': 40,
        'ERROR': 50
    }
    
    def __init__(self, level: str = 'INFO', output: Optional[TextIO] = None):
        self.level = self.LEVELS.get(level, 20)
        self.output = output or sys.stdout
        self.use_colors = self._supports_color()
    
    def _supports_color(self) -> bool:
        """
          Detecta se terminal suporta cores.
        """
        try:
            import os
            return hasattr(self.output, 'isatty') and self.output.isatty()
        except Exception:
            return False
    
    def _format_message(self, message: str, level: str) -> str:
        """
          Formata mensagem com timestamp e nível.
        """
        timestamp = datetime.now().strftime('%H:%M:%S')
        return f"[{timestamp}] [{level}] {message}"
    
    def log(self, message: str, level: str = "INFO") -> None:
        """
        Loga mensagem se nível permitido.
        """
        level_upper = level.upper()
        level_num = self.LEVELS.get(level_upper, 20)
        
        if level_num < self.level:
            return
        
        formatted = self._format_message(message, level_upper)
        
        if self.use_colors:
            color_map = {
                'DEBUG': Colors.GRAY,
                'INFO': Colors.INFO,
                'STEP': Colors.BOLD,
                'SUCCESS': Colors.SUCCESS,
                'WARNING': Colors.WARNING,
                'ERROR': Colors.ERROR
            }
            prefix = color_map.get(level_upper, '')
            suffix = Colors.END
            formatted = f"{prefix}{formatted}{suffix}"
        
        self.output.write(formatted + '\n')
        self.output.flush()
    
    def debug(self, msg: str) -> None:
        self.log(msg, 'DEBUG')
    
    def info(self, msg: str) -> None:
        self.log(msg, 'INFO')
    
    def step(self, msg: str) -> None:
        self.log(msg, 'STEP')
    
    def success(self, msg: str) -> None:
        self.log(msg, 'SUCCESS')
    
    def warning(self, msg: str) -> None:
        self.log(msg, 'WARNING')
    
    def error(self, msg: str) -> None:
        self.log(msg, 'ERROR')


_logger = Logger()

def log(message: str, level: str = "INFO") -> None:
    """
      Função convenience para logging.
    """
    _logger.log(message, level)


def set_log_level(level: str) -> None:
    """
      Define nível de log global.
    """
    _logger.level = Logger.LEVELS.get(level.upper(), 20)


def set_output(output: TextIO) -> None:
    """
      Define saída de log.
    """
    _logger.output = output
    _logger.use_colors = _logger._supports_color()
