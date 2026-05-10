import os
import shutil
import subprocess
from typing import Optional

from utils.logger import log


class InstallController:
    """
      Controller de instalação de APKs no dispositivo.
    """
    
    @staticmethod
    def install_apk(apk_path: Optional[str] = None) -> bool:
        """
          Instala APK no dispositivo Android.
        
          Args:
            apk_path: Caminho do APK (padrão: app-final.apk no diretório atual)
            
        Returns:
            bool: True se instalação iniciada com sucesso
        """
        if apk_path is None:
            apk_path = os.path.join(os.getcwd(), "app-final.apk")
        
        apk_path = os.path.abspath(apk_path)
        cwd = os.path.abspath(os.getcwd())
        
        if not apk_path.startswith(cwd):
            log("Caminho do APK inválido (fora do diretório do projeto)", "ERROR")
            return False
        
        if not os.path.exists(apk_path):
            log(f"APK não encontrado: {os.path.basename(apk_path)}", "ERROR")
            log("Execute 'tamk --build' primeiro", "INFO")
            return False
        
        if not InstallController._verify_apk_signature(apk_path):
            log("Aviso: Não foi possível verificar assinatura do APK", "WARNING")

        log(f"📦 Solicitando instalação: {os.path.basename(apk_path)}", "STEP")
        
        if shutil.which("termux-share"):
            try:
                result = subprocess.run(
                    ["termux-share", apk_path],
                    capture_output=True,
                    text=True,
                    timeout=30
                )
                
                if result.returncode == 0:
                    log("Instalador do Android iniciado", "SUCCESS")
                    return True
                else:
                    log(f"termux-share falhou: {result.stderr}", "ERROR")
                    return InstallController._install_fallback(apk_path)
                    
            except subprocess.TimeoutExpired:
                log("Timeout ao iniciar instalador", "ERROR")
                return False
            except Exception as e:
                log(f"Erro: {str(e)}", "ERROR")
                return InstallController._install_fallback(apk_path)
        else:
            return InstallController._install_fallback(apk_path)

    @staticmethod
    def _verify_apk_signature(apk_path: str) -> bool:
        """
          Verifica se APK está assinado.
        """
        try:
            result = subprocess.run(
                ["apksigner", "verify", apk_path],
                capture_output=True,
                text=True,
                timeout=10
            )
            return result.returncode == 0
        except Exception:
            return False

    @staticmethod
    def _install_fallback(apk_path: str) -> bool:
        """
          Método alternativo de instalação usando am (Activity Manager).
        """
        try:
            tmp_path = "/sdcard/Download/tamk_temp.apk"
            shutil.copy2(apk_path, tmp_path)
            
            result = subprocess.run(
                ["am", "start", "-a", "android.intent.action.VIEW",
                 "-t", "application/vnd.android.package-archive",
                 "-d", f"file://{tmp_path}"],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            if result.returncode == 0:
                log("Instalador iniciado (método alternativo)", "SUCCESS")
                return True
            else:
                log("Falha no método alternativo", "ERROR")
                log(f"Instale manualmente: {tmp_path}", "INFO")
                return False
                
        except Exception as e:
            log(f"Fallback falhou: {str(e)}", "ERROR")
            return False
