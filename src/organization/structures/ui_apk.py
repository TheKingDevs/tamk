import os
import re
import shutil
import subprocess
from pathlib import Path
from typing import Optional

from utils.logger import log
from config.tamk_config import Config

conf = Config()


class UIAppStructure:
    """
      Estrutura de projeto para aplicativos Android nativos com UI.
    """
    
    RESERVED_NAMES = {
        'android', 'com', 'org', 'java', 'kotlin', 'assets', 
        'res', 'src', 'build', 'test', 'main', 'secret'
    }

    SEMVER_PATTERN = re.compile(
        r'^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)'
        r'(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?'
        r'(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'
    )

    def setup(self, name: str, version: str, author: str, 
              password: Optional[str], web_url: Optional[str] = None) -> bool:
        """
          Configura estrutura de projeto UI APK.
        """
        safe_name = self._sanitize_name(name)
        safe_author = self._sanitize_author(author)
        
        is_valid, normalized = self._validate_version(version)
        if not is_valid:
            log(f"Versão inválida: '{version}'", "ERROR")
            log("Formato esperado: MAJOR.MINOR.PATCH (ex: 1.0.0)", "INFO")
            return False
        
        version = normalized
        
        if password and len(password) < 6:
            log("Senha deve ter pelo menos 6 caracteres", "ERROR")
            return False

        self.base = os.path.join(os.getcwd(), safe_name)
        
        if os.path.exists(self.base):
            counter = 1
            original = self.base
            while os.path.exists(self.base):
                self.base = f"{original}_{counter}"
                counter += 1
        
        self.package = f"com.{safe_author}.{safe_name}"

        self.tmpl_dir = conf.get_template_dir("ui_apk")

        try:
            self._create_folders()
            self._generate_keystore(password, safe_author)
            
            package_path = self.package.replace('.', os.sep)
            
            mappings = {
                "AndroidManifest.xml": "AndroidManifest.xml.tmpl",
                os.path.join("res", "layout", "activity_main.xml"): "activity_main.xml.tmpl",
                os.path.join("res", "values", "strings.xml"): "strings.xml.tmpl",
                os.path.join("res", "values", "styles.xml"): "styles.xml.tmpl",
                os.path.join("res", "drawable", "ic_launcher.xml"): "icon.xml.tmpl",
                os.path.join("res", "drawable", "ic_launcher_round.xml"): "icon.xml.tmpl",
                os.path.join("src", "main", "kotlin", package_path, "MainActivity.kt"): "MainActivity.kt.tmpl"
            }

            replacements = {
                "{{NAME}}": safe_name,
                "{{PACKAGE}}": self.package,
                "{{VERSION}}": version,
                "{{AUTHOR}}": author
            }

            log("Processando templates de UI...", "STEP")
            for dest, tmpl in mappings.items():
                self._generate_file(tmpl, dest, replacements)

            self._save_config(safe_name, version)
            self._run_local_setup()

            log(f"✅ Projeto UI '{safe_name}' criado!", "SUCCESS")
            return True
            
        except Exception as e:
            log(f"Erro: {str(e)}", "ERROR")
            if os.path.exists(self.base):
                shutil.rmtree(self.base, ignore_errors=True)
            return False

    def _sanitize_name(self, name: str) -> str:
        """
          Sanitiza nome do projeto.
        """
        if not name:
            raise ValueError("Nome não pode ser vazio")
        clean = re.sub(r'[^a-zA-Z0-9\-_]', '', name.strip())[:50]
        if not clean:
            raise ValueError("Nome inválido")
        if clean[0].isdigit():
            clean = "app_" + clean
        return clean

    def _sanitize_author(self, author: str) -> str:
        """
          Sanitiza nome do autor.
        """
        if not author:
            return "developer"
        return re.sub(r'[^a-zA-Z0-9]', '', author.lower())[:20] or "developer"

    def _validate_version(self, version: str) -> tuple[bool, str]:
        """
          Valida formato semver.
        """
        if not version:
            return False, "1.0.0"
        
        version = version.strip()
        if version.startswith('v') or version.startswith('V'):
            version = version[1:]
        
        if not self.SEMVER_PATTERN.match(version):
            return False, version
        
        return True, version

    def _create_folders(self) -> None:
        paths = [
            os.path.join("src", "main", "kotlin", *self.package.split('.')),
            os.path.join("res", "layout"),
            os.path.join("res", "values"),
            os.path.join("res", "drawable"),
            "secret"
        ]
        for p in paths:
            os.makedirs(os.path.join(self.base, p), exist_ok=True)

    def _generate_file(self, tmpl_name: str, dest_path: str, reps: dict) -> None:
        """
          Gera arquivo a partir de template.
        """
        src = os.path.join(self.tmpl_dir, tmpl_name)
        if not os.path.exists(src):
            log(f"Template não encontrado: {tmpl_name}", "WARNING")
            return

        with open(src, "r", encoding='utf-8') as f:
            content = f.read()

        for key, value in reps.items():
            content = content.replace(key, value)

        final_path = os.path.join(self.base, dest_path)
        os.makedirs(os.path.dirname(final_path), exist_ok=True)
        
        with open(final_path, "w", encoding='utf-8') as f:
            f.write(content)

    def _generate_keystore(self, password: Optional[str], author: str) -> None:
        """
          Gera keystore do projeto.
        """
        if not password:
            return
            
        ks_path = os.path.join(self.base, "secret", "project.keystore")
        safe_cn = re.sub(r'[^a-zA-Z0-9\s]', '', author)[:64]
        
        log("Gerando Keystore...", "INFO")
        
        result = subprocess.run([
            "keytool", "-genkey", "-v",
            "-keystore", ks_path,
            "-alias", "project_key",
            "-keyalg", "RSA", "-keysize", "2048", "-validity", "10000",
            "-storepass", password, "-keypass", password,
            "-dname", f"CN={safe_cn}, O=TAMK, C=BR"
        ], capture_output=True, text=True, timeout=60)
        
        if result.returncode != 0:
            raise RuntimeError(f"Falha ao gerar keystore: {result.stderr}")
        
        try:
            os.chmod(ks_path, 0o600)
        except Exception:
            pass

    def _save_config(self, name: str, version: str) -> None:
        """
          Salva configuração do projeto.
        """
        config_path = os.path.join(self.base, "tamk.config")
        with open(config_path, "w", encoding='utf-8') as f:
            f.write(f"type=ui_apk\nname={name}\nversion={version}\npackage={self.package}\n")

    def _run_local_setup(self) -> None:
        """
          Executa setup local.
        """
        old_cwd = os.getcwd()
        try:
            os.chdir(self.base)
            from controllers.setup_controller import SetupController
            SetupController().setup_environment()
        except Exception as e:
            log(f"Setup local: {e}", "WARNING")
        finally:
            os.chdir(old_cwd)
