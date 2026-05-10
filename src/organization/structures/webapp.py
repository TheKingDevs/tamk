import os
import re
import shutil
import subprocess
from pathlib import Path
from typing import Optional
from urllib.parse import urlparse

from utils.logger import log
from config.tamk_config import Config

conf = Config()


class WebAppStructure:
    """
      Estrutura de projeto para WebApps.
    """
    
    RESERVED_NAMES = {
        'android', 'com', 'org', 'java', 'kotlin', 'assets', 'res', 
        'src', 'build', 'test', 'main', 'secret', 'tamk'
    }
    
    SEMVER_PATTERN = re.compile(
        r'^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)'
        r'(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?'
        r'(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'
    )
    
    def __init__(self):
        self.base: str = ""
        self.package: str = ""
        self.tmpl_dir: str = ""

    def _sanitize_name(self, name: str) -> str:
        """
          Sanitiza nome do projeto para uso seguro em paths e package names.
        """
        if not name or not isinstance(name, str):
            raise ValueError("Nome do projeto não pode ser vazio")
        
        clean = re.sub(r'[^a-zA-Z0-9\-_]', '', name.strip())
        clean = clean[:50]
        
        if not clean:
            raise ValueError("Nome do projeto inválido após sanitização")
        
        if clean[0].isdigit():
            clean = "app_" + clean

        return clean

    def _sanitize_author(self, author: str) -> str:
        """
          Sanitiza nome do autor para package name.
        """
        if not author or not isinstance(author, str):
            return "developer"
        
        clean = re.sub(r'[^a-zA-Z0-9]', '', author.strip().lower())
        return clean[:20] or "developer"

    def _validate_version(self, version: str) -> tuple[bool, str]:
        """
          Valida formato de versão semver.
        
          Returns:
            tuple: (is_valid, normalized_version)
        """
        if not version or not isinstance(version, str):
            return False, "1.0.0"
        
        version = version.strip()
        
        if version.startswith('v') or version.startswith('V'):
            version = version[1:]
        
        if not self.SEMVER_PATTERN.match(version):
            return False, version
        
        return True, version

    def _validate_package_name(self, package: str) -> bool:
        """
          Valida formato de package name Android.
        """
        pattern = r'^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$'
        return bool(re.match(pattern, package)) and len(package) <= 100

    def _validate_url(self, url: str) -> bool:
        """
          Valida URL para modo externo.
        """
        try:
            parsed = urlparse(url)
            return parsed.scheme in ('http', 'https') and parsed.netloc
        except Exception:
            return False

    def setup(self, name: str, version: str, author: str, 
              password: Optional[str], web_url: str) -> bool:
        """
          Configura estrutura completa de WebApp com validações.
        """
        try:
            safe_name = self._sanitize_name(name)
        except ValueError as e:
            log(f"Erro no nome do projeto: {e}", "ERROR")
            return False
        
        safe_author = self._sanitize_author(author)
        
        is_valid_version, normalized_version = self._validate_version(version)
        if not is_valid_version:
            log(f"Versão inválida: '{version}'", "ERROR")
            log("Formato esperado: MAJOR.MINOR.PATCH (ex: 1.0.0, 2.3.1-beta)", "INFO")
            return False
        
        version = normalized_version
        
        if password and len(password) < 6:
            log("Senha deve ter pelo menos 6 caracteres", "ERROR")
            return False
        
        is_internal = "android_asset" in web_url
        if not is_internal and not self._validate_url(web_url):
            log(f"URL inválida: {web_url}", "ERROR")
            return False

        self.base = os.path.join(os.getcwd(), safe_name)
        
        if os.path.exists(self.base):
            counter = 1
            original_base = self.base
            while os.path.exists(self.base):
                self.base = f"{original_base}_{counter}"
                counter += 1
            log(f"Diretório existente detectado. Usando: {os.path.basename(self.base)}", "WARNING")
        
        self.package = f"com.{safe_author}.{safe_name}"
        if not self._validate_package_name(self.package):
            self.package = f"com.tamk.{safe_name}"

        self.tmpl_dir = conf.get_template_dir("webapp")

        log(f"Criando WebApp: {safe_name} ({self.package})", "INFO")

        try:
            self._create_folders(is_internal)

            if password:
                self._generate_keystore(password, safe_author)
            else:
                log("Aviso: Nenhuma senha fornecida. Usando keystore de debug.", "WARNING")

            mappings = self._get_template_mappings(is_internal)

            replacements = {
                "{{NAME}}": safe_name,
                "{{PACKAGE}}": self.package,
                "{{VERSION}}": version,
                "{{AUTHOR}}": author,
                "{{WEB_URL}}": web_url,
                "{{DEV_MODE}}": "false",
                "{{MIN_SDK}}": "21",
                "{{TARGET_SDK}}": "30",
                "{{TAMK_VERSION}}": conf.VERSION
            }

            mode_text = "Interno (Assets)" if is_internal else f"Externo (URL: {web_url})"
            log(f"Processando templates • Modo: {mode_text}", "STEP")
            
            for dest, tmpl_name in mappings.items():
                self._generate_file(tmpl_name, dest, replacements)

            self._save_config(safe_name, version, author, web_url)
            self._run_local_setup()

            log(f"✅ WebApp '{safe_name}' criado em: {self.base}", "SUCCESS")
            log(f"Package: {self.package}", "INFO")
            log(f"Versão: {version}", "INFO")
            log(f"Próximo passo: cd {os.path.basename(self.base)} && tamk --build", "INFO")
            return True
            
        except Exception as e:
            log(f"Erro ao criar projeto: {str(e)}", "ERROR")
            if os.path.exists(self.base):
                try:
                    shutil.rmtree(self.base)
                    log("Diretório parcial removido", "INFO")
                except Exception:
                    pass
            return False

    def _get_template_mappings(self, is_internal: bool) -> dict:
        """
          Retorna mapeamento de templates baseado no modo.
        """
        package_path = self.package.replace('.', os.sep)

        mappings = {
            os.path.join("src", "main", "kotlin", package_path, "MainActivity.kt"): "MainActivity.kt.tmpl",
            os.path.join("res", "values", "styles.xml"): "styles.xml.tmpl",
            os.path.join("res", "values", "strings.xml"): "strings.xml.tmpl",
            os.path.join("res", "drawable", "ic_launcher.xml"): "icon.xml.tmpl",
            os.path.join("res", "drawable", "ic_launcher_round.xml"): "icon.xml.tmpl",
            os.path.join("res", "xml", "network_security_config.xml"): "network_security_config.xml.tmpl",
            "AndroidManifest.xml": "AndroidManifest.xml.tmpl",
            ".gitignore": "gitignore_root.tmpl"
        }

        if is_internal:
            mappings[os.path.join("src", "main", "assets", "index.html")] = "index.html.tmpl"
            mappings[os.path.join("src", "main", "assets", "css", "styles.css")] = "css/styles.css.tmpl"
            mappings[os.path.join("src", "main", "assets", "js", "app.js")] = "js/app.js.tmpl"
            mappings[os.path.join("src", "main", "assets", ".gitignore")] = "gitignore_assets.tmpl"

        return mappings

    def _create_folders(self, is_internal: bool) -> None:
        """
          Cria estrutura de diretórios de forma segura usando apenas path.join.
        """
        paths = [
            os.path.join("src", "main", "kotlin", *self.package.split('.')),
            os.path.join("res", "values"),
            os.path.join("res", "drawable"),
            os.path.join("res", "xml"),
            "secret"
        ]
        
        if is_internal:
            paths.extend([
                os.path.join("src", "main", "assets"),
                os.path.join("src", "main", "assets", "css"),
                os.path.join("src", "main", "assets", "js"),
                os.path.join("src", "main", "assets", "images")
            ])
        
        for p in paths:
            full_path = os.path.join(self.base, p)
            resolved = os.path.abspath(full_path)
            
            base_resolved = os.path.abspath(self.base)
            if not resolved.startswith(base_resolved):
                raise ValueError(f"Tentativa de path traversal detectada: {p}")
            
            os.makedirs(resolved, exist_ok=True)
            log(f"  📁 {p}", "DEBUG")

    def _generate_file(self, tmpl_name: str, dest_path: str, reps: dict) -> None:
        """
          Gera arquivo a partir de template com validações.
        """
        src = os.path.join(self.tmpl_dir, tmpl_name)
        
        if not os.path.exists(src):
            common_dir = os.path.join(os.path.dirname(self.tmpl_dir), "ui_apk")
            src = os.path.join(common_dir, tmpl_name)
            
        if not os.path.exists(src):
            if tmpl_name.endswith('.gitignore'):
                content = "# Assets do WebApp\n*.tmp\n.cache/\n"
            else:
                log(f"Template não encontrado: {tmpl_name}", "ERROR")
                return
        else:
            with open(src, "r", encoding='utf-8') as f:
                content = f.read()

        for key, value in reps.items():
            content = content.replace(key, str(value))

        final_path = os.path.join(self.base, dest_path)
        
        parent = os.path.dirname(final_path)
        if parent:
            os.makedirs(parent, exist_ok=True)
        
        if os.path.exists(final_path):
            backup = final_path + ".backup"
            shutil.copy2(final_path, backup)
        
        with open(final_path, "w", encoding='utf-8') as f:
            f.write(content)

    def _generate_keystore(self, password: str, author: str) -> None:
        """
          Gera keystore de projeto de forma segura.
        """
        ks_path = os.path.join(self.base, "secret", "project.keystore")
        
        log("Gerando Keystore privada (RSA 2048)...", "INFO")
        
        safe_cn = re.sub(r'[^a-zA-Z0-9\s]', '', author)[:64]
        
        try:
            result = subprocess.run([
                "keytool",
                "-genkey",
                "-v",
                "-keystore", ks_path,
                "-alias", "project_key",
                "-keyalg", "RSA",
                "-keysize", "2048",
                "-validity", "10000",
                "-storepass", password,
                "-keypass", password,
                "-dname", f"CN={safe_cn}, O=TAMK-Web, C=BR"
            ], capture_output=True, text=True, timeout=60)
            
            if result.returncode != 0:
                raise RuntimeError(f"keytool falhou: {result.stderr}")
            
            try:
                os.chmod(ks_path, 0o600)
            except Exception:
                pass
                
        except subprocess.TimeoutExpired:
            raise RuntimeError("Timeout ao gerar keystore")
        except FileNotFoundError:
            raise RuntimeError("keytool não encontrado. Instale OpenJDK.")

    def _save_config(self, name: str, version: str, author: str, web_url: str) -> None:
        """
          Salva configuração do projeto.
        """
        config_path = os.path.join(self.base, "tamk.config")
        
        config_content = f"""# T.A.M.K Project Configuration
# Gerado automaticamente - NÃO EDITE MANUALMENTE

type=webapp
name={name}
version={version}
author={author}
package={self.package}
web_url={web_url}
created_with={conf.VERSION}
"""
        
        with open(config_path, "w", encoding='utf-8') as f:
            f.write(config_content)

    def _run_local_setup(self) -> None:
        """
          Executa setup do ambiente local.
        """
        old_cwd = os.getcwd()
        try:
            os.chdir(self.base)
            try:
                from controllers.setup_controller import SetupController
                SetupController().setup_environment()
            except Exception as e:
                log(f"Aviso: Setup de ambiente falhou (não crítico): {e}", "WARNING")
        finally:
            os.chdir(old_cwd)
