import os
from pathlib import Path
from typing import Optional


class Config:
    """
      Configuração centralizada do T.A.M.K.
    """

    VERSION: str = "2026.3.0-HMR"
    
    def __init__(self):
        self._detect_environment()
                
        self.DEV_DIR = self._secure_path(os.getcwd(), "development")        
        self.SDK_PATH = self._secure_path(self.DEV_DIR, "sdk", "android.jar")
        self.KEYSTORE = self._secure_path(self.DEV_DIR, "secret", "debug.keystore")
        
        self._validate_paths()

    def _detect_environment(self) -> None:
        """
          Detecta ambiente de execução.
        """
        prefix = os.environ.get('PREFIX', '')
        
        if '/org.smartide.code' in prefix:
            self.ENV = "smartide"
            self.SMARTIDE_PATH = "/data/data/org.smartide.code/files/home/"
        elif '/com.termux' in prefix:
            self.ENV = "termux"
            self.SMARTIDE_PATH = os.path.expanduser("~")
        else:
            self.ENV = "unknown"
            self.SMARTIDE_PATH = os.path.expanduser("~")
        
        self.TAMK_HOME = os.environ.get('TAMK_HOME', os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

    def _secure_path(self, *parts: str) -> str:
        """
          Constrói path de forma segura, prevenindo path traversal.
        """
        path = os.path.join(*parts)
        return os.path.abspath(os.path.normpath(path))

    def _validate_paths(self) -> None:
        """
          Valida que paths não estão fora de diretórios permitidos.
        """
        home = os.path.expanduser("~")
        
        for attr_name, path in [
            ('DEV_DIR', self.DEV_DIR),
            ('SDK_PATH', self.SDK_PATH),
            ('KEYSTORE', self.KEYSTORE)
        ]:
            if not (path.startswith(home) or path.startswith(os.getcwd())):
                raise RuntimeError(
                    f"Path inválido para {attr_name}: {path} "
                    f"(fora de {home} ou {os.getcwd()})"
                )

    def get_template_dir(self, template_type: str) -> str:
        """
          Retorna diretório de templates de forma segura.

          Args:
            template_type: Tipo de template ('webapp', 'ui_apk', 'console')
        """
        allowed = {'webapp', 'ui_apk', 'console'}
        if template_type not in allowed:
            raise ValueError(f"Tipo de template inválido: {template_type}")

        candidates = []
        
        if os.environ.get('TAMK_HOME'):
            candidates.append(Path(os.environ['TAMK_HOME']) / "assets" / "templates" / template_type)
                    
        candidates.append(Path(__file__).parent.parent.parent / "assets" / "templates" / template_type)        
        candidates.append(Path(os.getcwd()) / "assets" / "templates" / template_type)      
        candidates.append(Path("/data/data/com.termux/files/usr/opt/tamk/assets/templates") / template_type)
        candidates.append(Path("/usr/opt/tamk/assets/templates") / template_type)

        for candidate in candidates:
            if candidate.exists():
                return str(candidate)

        return str(candidates[0]) if candidates else f"/data/data/com.termux/files/usr/opt/tamk/assets/templates/{template_type}"

    def ensure_directories(self) -> None:
        """
          Garante que diretórios necessários existem.
        """
        os.makedirs(self.DEV_DIR, mode=0o755, exist_ok=True)
        os.makedirs(os.path.dirname(self.SDK_PATH), mode=0o755, exist_ok=True)
        os.makedirs(os.path.dirname(self.KEYSTORE), mode=0o700, exist_ok=True)
