import os
import re
import glob
import json
import shlex
import shutil
import hashlib
import zipfile
import tempfile
import subprocess
from pathlib import Path
from getpass import getpass
from typing import Optional, Dict, List

from utils.logger import log
from config.tamk_config import Config


class BuildController:
    def __init__(self, verbose: bool = False, password: Optional[str] = None):
        self.conf = Config()
        self.verbose = verbose
        self.root_dir = os.getcwd()
        self.cache_dir = os.path.join(self.root_dir, "assets", "cache")
        self.hash_file = os.path.join(self.root_dir, ".build_cache")
        
        self._validate_project_structure()

    def _validate_project_structure(self) -> None:
        """
          Valida se o diretório atual é um projeto T.A.M.K válido.
        """
        required_files = ["tamk.config", "AndroidManifest.xml"]
        missing = [f for f in required_files if not os.path.exists(f)]
        if missing:
            raise RuntimeError(f"Estrutura de projeto inválida. Arquivos faltando: {missing}")

    def _sanitize_path(self, path: str) -> str:
        """
          Sanitiza caminhos para prevenir path traversal.
        """
        clean = re.sub(r'[;&|`$]', '', path)
        clean = os.path.normpath(clean)
        resolved = os.path.abspath(os.path.join(self.root_dir, clean))
        if not resolved.startswith(os.path.abspath(self.root_dir)):
            raise ValueError(f"Path traversal detectado: {path}")
        return resolved

    def _execute(self, command: List[str], shell: bool = False, cwd: Optional[str] = None) -> bool:
        """
          Executa comando de forma segura.
        
          Args:
            command: Lista de argumentos (recomendado) ou string (quando inevitável)
            shell: False por padrão - só True quando absolutamente necessário
            cwd: Diretório de trabalho opcional
        """
        if self.verbose:
            safe_cmd = ' '.join(shlex.quote(str(c)) for c in command)
            log(f"Executando: {safe_cmd}", "DEBUG")
        
        try:
            result = subprocess.run(
                command,
                shell=shell,
                capture_output=not self.verbose,
                text=True,
                cwd=cwd,
                timeout=300
            )
            
            if result.returncode != 0:
                log(f"Falha no comando: {command[0] if command else 'unknown'}", "ERROR")
                if not self.verbose and result.stderr:
                    safe_stderr = re.sub(r'--ks-pass\s+\S+', '--ks-pass [REDACTED]', result.stderr)
                    print(safe_stderr)
                return False
            return True
            
        except subprocess.TimeoutExpired:
            log("Comando excedeu tempo limite (5 minutos)", "ERROR")
            return False
        except FileNotFoundError as e:
            log(f"Comando não encontrado: {e.filename}", "ERROR")
            return False
        except Exception as e:
            log(f"Erro inesperado na execução: {str(e)}", "ERROR")
            return False

    def _calculate_project_hash(self) -> Dict[str, str]:
        """
          Calcula hash de todos os arquivos relevantes.
        """
        hashes = {}
        folders_to_watch = ["src", "res", "AndroidManifest.xml"]
        
        for item in folders_to_watch:
            path = os.path.join(self.root_dir, item)
            if os.path.isfile(path):
                hashes[item] = self._file_hash(path)
            elif os.path.isdir(path):
                for root, _, files in os.walk(path):
                    for f in files:
                        full_path = os.path.join(root, f)
                        if any(skip in full_path for skip in ['.git', 'cache', 'secret', '__pycache__']):
                            continue
                        relative_path = os.path.relpath(full_path, self.root_dir)
                        hashes[relative_path] = self._file_hash(full_path)
        return hashes

    def _file_hash(self, path: str) -> str:
        """
          Gera hash SHA-256 de arquivo.
        """
        hasher = hashlib.sha256()
        try:
            with open(path, 'rb') as afile:
                while chunk := afile.read(8192):
                    hasher.update(chunk)
            return hasher.hexdigest()
        except (IOError, OSError) as e:
            log(f"Erro ao ler arquivo para hash: {path}", "WARNING")
            return ""

    def _must_recompile(self, current_hashes: Dict[str, str]) -> bool:
        """
          Verifica se recompilação é necessária.
        """
        if not os.path.exists(self.hash_file) or not os.path.exists("app-final.apk"):
            return True
        
        try:
            with open(self.hash_file, "r", encoding='utf-8') as f:
                old_hashes = json.load(f)
            return old_hashes != current_hashes
        except (json.JSONDecodeError, IOError):
            return True

    def _validate_keystore_password(self, ks_path: str, password: str) -> bool:
        """
          Valida senha da keystore de forma segura.
        """
        try:
            result = subprocess.run(
                [
                    "keytool",
                    "-list",
                    "-keystore", ks_path,
                    "-storepass", password,
                    "-alias", "project_key"
                ],
                capture_output=True,
                text=True,
                timeout=30
            )
            return result.returncode == 0
        except Exception as e:
            log(f"Erro na validação da keystore: {str(e)}", "ERROR")
            return False

    def build_apk(self, password: Optional[str] = None) -> None:
        """
          Pipeline completo de build.
        
          Args:
            password: Senha da keystore (se None, solicita interativamente)
        """
        if not os.path.exists(self.conf.SDK_PATH):
            log("Ambiente local incompleto! Rode 'tamk --setup'", "ERROR")
            return

        if password is None:
            password = getpass("Senha da Keystore: ")
        
        if len(password) < 6:
            log("Senha deve ter pelo menos 6 caracteres", "ERROR")
            return

        log("Verificando integridade dos arquivos...", "INFO")
        current_hashes = self._calculate_project_hash()
        
        if not self._must_recompile(current_hashes):
            log("✨ Nada mudou desde o último build. APK atualizado!", "SUCCESS")
            return

        project_ks = os.path.join(self.root_dir, "secret", "project.keystore")
        ks_path = project_ks if os.path.exists(project_ks) else self.conf.KEYSTORE
        ks_pass = password if os.path.exists(project_ks) else "android"

        log("Validando credenciais da keystore...", "STEP")
        if not self._validate_keystore_password(ks_path, ks_pass):
            log("SENHA INCORRETA ou keystore inválida! Build abortado.", "ERROR")
            password = "0" * len(password)
            return
        
        del password

        log("🚀 Iniciando Build Completo...", "INFO")
        
        try:
            shutil.rmtree(self.cache_dir, ignore_errors=True)
            os.makedirs(os.path.join(self.cache_dir, "gen"), exist_ok=True)
            os.makedirs(os.path.join(self.cache_dir, "obj"), exist_ok=True)
        except OSError as e:
            log(f"Erro ao preparar diretórios de cache: {e}", "ERROR")
            return

        # AAPT2 - Compilação de recursos
        log("Compilando recursos...", "STEP")
        res_zip = os.path.join(self.cache_dir, "res.zip")
        if not self._execute([
            "aapt2", "compile", "--dir", "res", "-o", res_zip
        ]):
            return

        # AAPT2 - Link (com assets se existirem)
        assets_path = os.path.join("src", "main", "assets")
        base_apk = os.path.join(self.cache_dir, "app.apk")
        
        link_cmd = [
            "aapt2", "link",
            "-I", self.conf.SDK_PATH,
            "--manifest", "AndroidManifest.xml",
            "--java", os.path.join(self.cache_dir, "gen"),
            "-o", base_apk,
            "--auto-add-overlay"
        ]
        
        if os.path.exists(assets_path):
            link_cmd.extend(["-A", assets_path])
        
        link_cmd.append(res_zip)
        
        if not self._execute(link_cmd):
            return

        # KOTLINC - Compilação Kotlin
        log("Compilando Kotlin...", "STEP")
        kotlin_src = os.path.join("src", "main", "kotlin")
        if os.path.exists(kotlin_src):
            compile_cmd = [
                "kotlinc",
                kotlin_src,
                os.path.join(self.cache_dir, "gen"),
                "-cp", self.conf.SDK_PATH,
                "-d", os.path.join(self.cache_dir, "obj")
            ]
            if not self._execute(compile_cmd):
                return

        # D8 - Geração de DEX
        log("Gerando DEX...", "STEP")
        classes = glob.glob(
            os.path.join(self.cache_dir, "obj", "**", "*.class"),
            recursive=True
        )
        
        if classes:
            d8_cmd = [
                "d8",
                "--lib", self.conf.SDK_PATH,
                "--release",
                "--output", self.cache_dir
            ] + classes
            
            if not self._execute(d8_cmd):
                return

        # Empacotamento e assinatura
        classes_dex = os.path.join(self.cache_dir, "classes.dex")
        if os.path.exists(classes_dex):
            if not self._execute([
                "zip", "-j", base_apk, classes_dex
            ]):
                return

        # Zipalign
        unsigned_apk = "app-unsigned.apk"
        if not self._execute([
            "zipalign", "-f", "4", base_apk, unsigned_apk
        ]):
            return

        # Assinatura
        final_apk = "app-final.apk"
        sign_result = subprocess.run([
            "apksigner", "sign",
            "--ks", ks_path,
            "--ks-pass", "pass:" + ks_pass,
            "--out", final_apk,
            unsigned_apk
        ], capture_output=True, text=True, timeout=60)

        if sign_result.returncode != 0:
            safe_err = re.sub(r'pass:\S+', 'pass:[REDACTED]', sign_result.stderr)
            log(f"Falha na assinatura: {safe_err}", "ERROR")
            return

        for temp_file in [unsigned_apk, base_apk]:
            try:
                if os.path.exists(temp_file):
                    os.remove(temp_file)
            except OSError:
                pass

        try:
            with open(self.hash_file, "w", encoding='utf-8') as f:
                json.dump(current_hashes, f, indent=2)
        except IOError as e:
            log(f"Aviso: Não foi possível salvar cache de build: {e}", "WARNING")

        log("✅ SUCESSO: app-final.apk gerado corretamente!", "SUCCESS")
        
        try:
            size = os.path.getsize(final_apk) / 1024  # KB
            log(f"📦 Tamanho do APK: {size:.1f} KB", "INFO")
        except OSError:
            pass

    def build_assets_only(self, password: Optional[str] = None) -> bool:
        """
          Build incremental apenas de assets
        """
        log("Build incremental de assets...", "STEP")
        
        assets_path = os.path.join("src", "main", "assets")
        if not os.path.exists(assets_path):
            log("Pasta assets não encontrada", "ERROR")
            return False
        
        base_apk = "app-final.apk"
        if not os.path.exists(base_apk):
            base_apk = "app-dev.apk"
            if not os.path.exists(base_apk):
                log("APK base não encontrado", "ERROR")
                return False
        
        project_ks = os.path.join(self.root_dir, "secret", "project.keystore")
        ks_path = project_ks if os.path.exists(project_ks) else self.conf.KEYSTORE
        ks_pass = password if (password and os.path.exists(project_ks)) else "android"
        
        if not self._validate_keystore_password(ks_path, ks_pass):
            log("Senha incorreta", "ERROR")
            return False

        temp_dir = tempfile.mkdtemp(prefix="tamk_dev_")
        
        try:
            extract_dir = os.path.join(temp_dir, "extracted")
            os.makedirs(extract_dir)
            
            with zipfile.ZipFile(base_apk, 'r') as zip_ref:
                zip_ref.extractall(extract_dir)
            
            old_assets = os.path.join(extract_dir, "assets")
            if os.path.exists(old_assets):
                shutil.rmtree(old_assets)
            
            shutil.copytree(assets_path, old_assets)
            
            temp_apk = os.path.join(temp_dir, "unsigned.apk")
            with zipfile.ZipFile(temp_apk, 'w', zipfile.ZIP_DEFLATED) as zipf:
                for root, dirs, files in os.walk(extract_dir):
                    for file in files:
                        file_path = os.path.join(root, file)
                        arcname = os.path.relpath(file_path, extract_dir)
                        zipf.write(file_path, arcname)
            
            dev_apk = "app-dev.apk"
            sign_result = subprocess.run([
                "apksigner", "sign",
                "--ks", ks_path,
                "--ks-pass", "pass:" + ks_pass,
                "--out", dev_apk,
                temp_apk
            ], capture_output=True, text=True, timeout=60)

            if sign_result.returncode == 0:
                log(f"✅ Assets atualizados: {dev_apk}", "SUCCESS")
                
                self._try_install(dev_apk)
                
                return True
            else:
                safe_err = re.sub(r'pass:\S+', 'pass:[REDACTED]', sign_result.stderr)
                log(f"Falha ao assinar: {safe_err}", "ERROR")
                return False
                
        except Exception as e:
            log(f"Erro no build de assets: {str(e)}", "ERROR")
            return False
            
        finally:
            shutil.rmtree(temp_dir, ignore_errors=True)
    
    def _try_install(self, apk_path: str) -> None:
        """
          Tenta instalar APK automaticamente via adb se disponível.
        """
        try:
            result = subprocess.run(
                ["adb", "devices"],
                capture_output=True,
                text=True,
                timeout=5
            )

            if "device" in result.stdout and "emulator" not in result.stdout:
                install = subprocess.run(
                    ["adb", "install", "-r", apk_path],
                    capture_output=True,
                    text=True,
                    timeout=30
                )

                if install.returncode == 0:
                    log("📱 APK instalado no dispositivo", "SUCCESS")
                else:
                    log("📱 APK pronto (instalação manual necessária)", "INFO")
            else:
                log("📱 APK pronto: app-dev.apk", "INFO")

        except (FileNotFoundError, subprocess.TimeoutExpired):
            log("📱 APK pronto para instalação manual", "INFO")

    def push_asset_to_device(self, asset_path: str, device_path: str = None) -> bool:
        """
          Envia um arquivo de asset específico para o dispositivo via ADB.
          Usado para HTML hot-reload sem rebuild completo.

          Args:
            asset_path: Caminho local do arquivo (dentro de src/main/assets/)
            device_path: Caminho no dispositivo (padrão: /sdcard/tamk_assets/)

          Returns:
            bool: True se envio foi bem-sucedido
        """
        if device_path is None:
            device_path = "/sdcard/tamk_assets"

        # Caminho absoluto do arquivo
        abs_asset = os.path.abspath(asset_path)
        if not os.path.exists(abs_asset):
            log(f"Arquivo não encontrado: {asset_path}", "ERROR")
            return False

        # Verifica se está dentro de assets
        assets_dir = os.path.join(self.root_dir, "src", "main", "assets")
        if not abs_asset.startswith(os.path.abspath(assets_dir)):
            log(f"Arquivo deve estar dentro de src/main/assets/: {asset_path}", "ERROR")
            return False

        # Caminho relativo dentro de assets
        rel_path = os.path.relpath(abs_asset, assets_dir)
        device_file = os.path.join(device_path, rel_path).replace("\\", "/")

        try:
            # 1. Criar diretório no dispositivo
            device_dir = os.path.dirname(device_file)
            subprocess.run(
                ["adb", "shell", "mkdir", "-p", device_dir],
                capture_output=True,
                timeout=10
            )

            # 2. Push do arquivo
            result = subprocess.run(
                ["adb", "push", abs_asset, device_file],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                log(f"📤 Asset enviado: {rel_path} → {device_file}", "SUCCESS")
                return True
            else:
                log(f"❌ Falha ao enviar asset: {result.stderr}", "ERROR")
                return False

        except subprocess.TimeoutExpired:
            log("Timeout ao enviar asset via ADB", "ERROR")
            return False
        except FileNotFoundError:
            log("ADB não encontrado. Instale android-tools", "ERROR")
            return False
        except Exception as e:
            log(f"Erro ao enviar asset: {str(e)}", "ERROR")
            return False

    def notify_device_refresh(self, asset_path: str = None) -> bool:
        """
          Notifica o dispositivo para recarregar um asset específico.
          Envia broadcast intent para o app escutar.

          Args:
            asset_path: Caminho relativo do asset (ex: 'index.html')

          Returns:
            bool: True se notificação foi enviada
        """
        try:
            # Verifica se dispositivo está conectado
            result = subprocess.run(
                ["adb", "devices"],
                capture_output=True,
                text=True,
                timeout=5
            )

            if "device" not in result.stdout or "emulator" in result.stdout:
                log("Dispositivo não conectado via ADB", "WARNING")
                return False

            # Envia broadcast para o app escutar
            # O app precisa ter um BroadcastReceiver registrado
            package_name = self._get_package_name()

            if asset_path:
                # Notifica arquivo específico
                cmd = [
                    "adb", "shell", "am", "broadcast",
                    "-a", "tamk.ACTION_REFRESH_ASSET",
                    "-n", f"{package_name}/.TAMKRefreshReceiver",
                    "--es", "path", asset_path
                ]
            else:
                # Notifica refresh geral
                cmd = [
                    "adb", "shell", "am", "broadcast",
                    "-a", "tamk.ACTION_REFRESH_ALL",
                    "-n", f"{package_name}/.TAMKRefreshReceiver"
                ]

            result = subprocess.run(cmd, capture_output=True, text=True, timeout=10)

            if result.returncode == 0:
                log(f"🔔 Refresh notificado: {asset_path if asset_path else 'todos assets'}", "INFO")
                return True
            else:
                # Broadcast pode falhar se receiver não existe -isso é normal em produção
                # Em dev, o app deve ter o receiver
                if "Broadcast completed" in result.stdout or "result=0" in result.stdout:
                    return True
                return False

        except Exception as e:
            log(f"Erro ao notificar dispositivo: {str(e)}", "ERROR")
            return False

    def _get_package_name(self) -> str:
        """
          Extrai package name do AndroidManifest.xml
        """
        manifest_path = os.path.join(self.root_dir, "AndroidManifest.xml")
        if not os.path.exists(manifest_path):
            return ""

        try:
            with open(manifest_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # Procura por package="..."
            import re
            match = re.search(r'package\s*=\s*["\']([^"\']+)["\']', content)
            if match:
                return match.group(1)
        except Exception:
            pass

        return ""
