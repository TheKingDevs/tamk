import os
import shutil
import hashlib
import zipfile
import subprocess
from pathlib import Path
from typing import Optional

from utils.logger import log
from config.tamk_config import Config


class SetupController:
    """
      Controller de configuração de ambiente T.A.M.K.
    """
    
    SDK_URL = "https://dl.google.com/android/repository/platform-30_r03.zip"
    SDK_EXPECTED_HASH = "a1b2c3d4..."
    
    def __init__(self):
        self.conf = Config()
        self._ensure_directories()

    def _ensure_directories(self) -> None:
        """
          Garante que diretórios necessários existem.
        """
        os.makedirs(os.path.dirname(self.conf.SDK_PATH), exist_ok=True)
        os.makedirs(os.path.dirname(self.conf.KEYSTORE), exist_ok=True)

    def _download_with_progress(self, url: str, dest: str) -> bool:
        """
          Download com wget mostrando progresso.
        """
        log(f"Baixando SDK do Google...", "INFO")
        
        try:
            result = subprocess.run(
                ["wget", "-q", "--show-progress", url, "-O", dest],
                capture_output=False,
                text=True,
                timeout=300
            )
            return result.returncode == 0
        except subprocess.TimeoutExpired:
            log("Timeout no download do SDK", "ERROR")
            return False
        except FileNotFoundError:
            log("wget não encontrado. Instale com: pkg install wget", "ERROR")
            return False

    def _verify_hash(self, filepath: str, expected_hash: str) -> bool:
        """
          Verifica SHA-256 de arquivo.
        """
        if not os.path.exists(filepath):
            return False
        
        sha256 = hashlib.sha256()
        try:
            with open(filepath, 'rb') as f:
                while chunk := f.read(8192):
                    sha256.update(chunk)
            return sha256.hexdigest() == expected_hash
        except IOError:
            return False

    def setup_environment(self) -> None:
        """
          Configura ambiente local: SDK Android e keystore de debug.
        """
        log("Configurando ambiente local...", "STEP")

        if not os.path.exists(self.conf.SDK_PATH):
            zip_tmp = os.path.join(self.conf.DEV_DIR, "sdk_temp.zip")
            
            if not self._download_with_progress(self.SDK_URL, zip_tmp):
                return
            
            log("Extraindo android.jar...", "INFO")
            try:
                with zipfile.ZipFile(zip_tmp, 'r') as zip_ref:
                    jar_in_zip = "android-11/android.jar"
                    if jar_in_zip not in zip_ref.namelist():
                        jar_candidates = [n for n in zip_ref.namelist() 
                                        if n.endswith('android.jar')]
                        if not jar_candidates:
                            raise ValueError("android.jar não encontrado no SDK")
                        jar_in_zip = jar_candidates[0]
                    
                    with zip_ref.open(jar_in_zip) as source:
                        with open(self.conf.SDK_PATH, 'wb') as target:
                            shutil.copyfileobj(source, target)
                
                os.remove(zip_tmp)
                log("SDK configurada com sucesso", "SUCCESS")
                
            except (zipfile.BadZipFile, ValueError) as e:
                log(f"Erro ao extrair SDK: {e}", "ERROR")
                for f in [zip_tmp, self.conf.SDK_PATH]:
                    try:
                        if os.path.exists(f):
                            os.remove(f)
                    except OSError:
                        pass
                return
            except Exception as e:
                log(f"Erro inesperado: {e}", "ERROR")
                return

        if not os.path.exists(self.conf.KEYSTORE):
            log("Gerando keystore de desenvolvimento...", "INFO")
            
            try:
                result = subprocess.run([
                    "keytool", "-genkey", "-v",
                    "-keystore", self.conf.KEYSTORE,
                    "-alias", "androiddebugkey",
                    "-keyalg", "RSA", "-keysize", "2048", "-validity", "10000",
                    "-storepass", "android",
                    "-keypass", "android",
                    "-dname", "CN=Android Debug, O=Android, C=US"
                ], capture_output=True, text=True, timeout=60)
                
                if result.returncode != 0:
                    log(f"Falha ao gerar keystore: {result.stderr}", "ERROR")
                    return
                
                try:
                    os.chmod(self.conf.KEYSTORE, 0o600)
                except Exception:
                    pass
                    
                log("Keystore de debug criada", "SUCCESS")
                
            except FileNotFoundError:
                log("keytool não encontrado. Instale OpenJDK.", "ERROR")
            except subprocess.TimeoutExpired:
                log("Timeout ao gerar keystore", "ERROR")

    def verify_environment(self) -> bool:
        """
          Verifica se ambiente está configurado corretamente.
        
          Returns:
            bool: True se tudo OK, False caso contrário
        """
        checks = {
            "SDK": os.path.exists(self.conf.SDK_PATH),
            "Keystore": os.path.exists(self.conf.KEYSTORE),
            "aapt2": shutil.which("aapt2") is not None,
            "kotlinc": shutil.which("kotlinc") is not None,
            "d8": shutil.which("d8") is not None,
            "apksigner": shutil.which("apksigner") is not None,
        }
        
        all_ok = all(checks.values())
        
        if not all_ok:
            log("Verificação de ambiente falhou:", "WARNING")
            for item, ok in checks.items():
                status = "✓" if ok else "✗"
                log(f"  [{status}] {item}", "INFO" if ok else "ERROR")
        
        return all_ok
