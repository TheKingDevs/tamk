"""
  Sistema de atualização automática do T.A.M.K.
  Instala automaticamente patches críticos e pergunta sobre atualizações comuns.
"""

import os
import sys
import json
import shutil
import tempfile
import subprocess
import urllib.error
import urllib.request
from datetime import datetime
from typing import Optional, Tuple

from .colors import TColor
from .update_checker import UpdateChecker, UpdateLevel, VersionInfo, check_and_notify
from .banner import Banner, BannerStyle, success_banner, error_banner, warning_banner, info_banner


class AutoUpdater:
    """
      Gerenciador de atualizações automáticas do T.A.M.K.
    """
    
    def __init__(self, current_version: str):
        self.current_version = current_version
        self.checker = UpdateChecker(current_version)
        self.banner = Banner()
        self.tamk_home = os.environ.get('TAMK_HOME', 
                                         os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
    
    def _print_colored(self, message: str, color: str = TColor.CYAN) -> None:
        """Imprime mensagem colorida."""
        print(f"{color}{message}{TColor.RESET}")
    
    def _run_command(self, command: str, capture: bool = False) -> Tuple[bool, str]:
        """
          Executa comando shell.
        
          Returns:
            Tuple[success, output]
        """
        try:
            if capture:
                result = subprocess.run(
                    command,
                    shell=True,
                    capture_output=True,
                    text=True,
                    timeout=120
                )
                return result.returncode == 0, result.stdout + result.stderr
            else:
                result = subprocess.run(command, shell=True, timeout=120)
                return result.returncode == 0, ""
        except subprocess.TimeoutExpired:
            return False, "Timeout na execução do comando"
        except Exception as e:
            return False, str(e)
    
    def _download_file(self, url: str, dest_path: str) -> bool:
        """Baixa arquivo de URL."""
        try:
            req = urllib.request.Request(
                url,
                headers={'User-Agent': 'TAMK-Auto-Updater'}
            )
            
            with urllib.request.urlopen(req, timeout=60) as response:
                with open(dest_path, 'wb') as out_file:
                    shutil.copyfileobj(response, out_file)
            
            return True
        except Exception as e:
            return False
    
    def _backup_current_installation(self) -> Optional[str]:
        """Cria backup da instalação atual."""
        try:
            backup_dir = tempfile.mkdtemp(prefix='tamk_backup_')
            backup_info = {
                'tamk_home': self.tamk_home,
                'timestamp': datetime.now().isoformat(),
                'version': self.current_version
            }
            
            info_file = os.path.join(backup_dir, 'backup_info.json')
            with open(info_file, 'w') as f:
                json.dump(backup_info, f, indent=2)
            
            src_dir = os.path.join(self.tamk_home, 'src')
            if os.path.exists(src_dir):
                shutil.copytree(src_dir, os.path.join(backup_dir, 'src'))
            
            self._print_colored(f"✓ Backup criado em: {backup_dir}", TColor.GREEN)
            return backup_dir
            
        except Exception as e:
            self._print_colored(f"✗ Falha ao criar backup: {e}", TColor.RED)
            return None
    
    def _restore_from_backup(self, backup_dir: str) -> bool:
        """Restaura instalação a partir do backup."""
        try:
            if not backup_dir or not os.path.exists(backup_dir):
                return False
            
            src_backup = os.path.join(backup_dir, 'src')
            if os.path.exists(src_backup):
                shutil.rmtree(os.path.join(self.tamk_home, 'src'), ignore_errors=True)
                shutil.copytree(src_backup, os.path.join(self.tamk_home, 'src'))
            
            self._print_colored("✓ Backup restaurado com sucesso", TColor.GREEN)
            return True
            
        except Exception as e:
            self._print_colored(f"✗ Falha ao restaurar backup: {e}", TColor.RED)
            return False
    
    def _update_via_git(self) -> bool:
        """Atualiza via git pull."""
        self._print_colored("\n📥 Baixando atualização via git...", TColor.CYAN)
        
        success, output = self._run_command(
            f"cd {self.tamk_home} && git pull --rebase --autostash",
            capture=True
        )
        
        if success:
            self._print_colored("✓ Código atualizado com sucesso", TColor.GREEN)
            return True
        else:
            self._print_colored(f"✗ Falha no git pull: {output}", TColor.RED)
            return False
    
    def _update_via_script(self) -> bool:
        """Atualiza via script de instalação."""
        self._print_colored("\n📥 Executando script de instalação...", TColor.CYAN)
        
        setup_script = os.path.join(self.tamk_home, 'setup-install.sh')
        if not os.path.exists(setup_script):
            self._print_colored("✗ Script de instalação não encontrado", TColor.RED)
            return False
        
        success, output = self._run_command(
            f"bash {setup_script}",
            capture=True
        )
        
        if success:
            self._print_colored("✓ Instalação atualizada com sucesso", TColor.GREEN)
            return True
        else:
            self._print_colored(f"✗ Falha na atualização: {output}", TColor.RED)
            return False
    
    def _update_via_pip(self) -> bool:
        """Atualiza via pip (se instalado como pacote)."""
        self._print_colored("\n📥 Atualizando via pip...", TColor.CYAN)
        
        success, output = self._run_command(
            "pip install --upgrade tamk 2>/dev/null || pip3 install --upgrade tamk 2>/dev/null",
            capture=True
        )
        
        if success:
            self._print_colored("✓ Pacote atualizado com sucesso", TColor.GREEN)
            return True
        else:
            self._print_colored("ℹ Atualização via pip não disponível", TColor.YELLOW)
            return False
    
    def _check_installation_method(self) -> str:
        """Detecta método de instalação atual."""
        git_dir = os.path.join(self.tamk_home, '.git')
        if os.path.exists(git_dir):
            return 'git'
        
        success, output = self._run_command("pip show tamk 2>/dev/null || pip3 show tamk 2>/dev/null", capture=True)
        if success and self.tamk_home in output:
            return 'pip'
        
        return 'script'
    
    def perform_update(self, automatic: bool = False) -> bool:
        """
          Realiza a atualização.
        
          Args:
            automatic: Se True, tenta atualização automática sem interação
            
          Returns:
            True se atualização foi bem sucedida
        """
        latest = self.checker.check_for_updates(force=True)
        
        if not latest:
            self._print_colored("\n✓ Você já está na versão mais recente!", TColor.GREEN)
            return True
        
        update_info = self.checker.get_update_info()
        level = latest.level
        
        self.banner.show(
            title=f"📦 ATUALIZAÇÃO DISPONÍVEL",
            subtitle=f"v{self.current_version} → v{latest.version_str}",
            style=BannerStyle.INFO if level in [UpdateLevel.MAJOR, UpdateLevel.MINOR] else BannerStyle.SUCCESS,
            lines=[
                f"Nível: {TColor.BOLD}{level.value.upper()}{TColor.RESET}",
                f"Publicada: {update_info.get('published_at', 'N/A')[:10] if update_info else 'N/A'}",
                "",
                "Notas de release:",
                f"{update_info.get('body', 'Sem notas')[:200] if update_info else 'Sem notas'}..." if update_info else ""
            ]
        )
        
        should_update = False
        
        if level in [UpdateLevel.CRITICAL, UpdateLevel.PATCH]:
            self._print_colored("\n⚡ Atualização crítica/patch - instalando automaticamente...", TColor.YELLOW)
            should_update = True
        elif automatic:
            should_update = True
        else:
            self._print_colored("\n" + "=" * 50, TColor.CYAN)
            self._print_colored("Deseja atualizar agora?", TColor.BOLD)
            self._print_colored("=" * 50 + "\n", TColor.CYAN)
            
            try:
                response = input(f"{TColor.GREEN}[Y/n]{TColor.RESET}: ").strip().lower()
                should_update = response in ['', 'y', 'yes', 'sim']
            except (EOFError, KeyboardInterrupt):
                self._print_colored("\nAtualização cancelada.", TColor.YELLOW)
                return False
        
        if not should_update:
            self._print_colored("\nℹ Atualização adiada. Execute 'tamk --update' quando desejar.", TColor.YELLOW)
            return True
        
        backup_dir = self._backup_current_installation()
        
        method = self._check_installation_method()
        self._print_colored(f"\n🔧 Método de instalação detectado: {TColor.BOLD}{method}{TColor.RESET}", TColor.CYAN)
        
        success = False
        
        if method == 'git':
            success = self._update_via_git()
        elif method == 'pip':
            success = self._update_via_pip()
        else:
            success = self._update_via_script()
        
        if success:
            success_banner(
                "ATUALIZAÇÃO CONCLUÍDA",
                f"T.A.M.K atualizado para v{latest.version_str}"
            )
            self._print_colored("\n💡 Reinicie o terminal para aplicar as mudanças.", TColor.YELLOW)
            return True
        else:
            if backup_dir:
                self._print_colored("\n🔄 Tentando restaurar backup...", TColor.YELLOW)
                self._restore_from_backup(backup_dir)
            
            error_banner(
                "ATUALIZAÇÃO FALHOU",
                "Ocorreu um erro durante a atualização"
            )
            return False
    
    def check_and_install_critical(self) -> bool:
        """
          Verifica e instala automaticamente apenas atualizações críticas.
        
          Returns:
            True se atualização crítica foi instalada
        """
        latest = self.checker.check_for_updates(force=False)
        
        if not latest:
            return False
        
        if latest.level in [UpdateLevel.CRITICAL, UpdateLevel.PATCH]:
            self._print_colored(
                f"\n{TColor.RED}{TColor.BOLD}⚠️ ATUALIZAÇÃO {latest.level.value.upper()} DETECTADA!{TColor.RESET}",
                TColor.RED
            )
            self._print_colored(
                f"Instalando automaticamente: v{self.current_version} → v{latest.version_str}",
                TColor.YELLOW
            )
            return self.perform_update(automatic=True)
        
        return False


def run_update_check(current_version: str, auto_critical: bool = True) -> None:
    """
      Função principal para verificação de atualizações.
    
      Args:
        current_version: Versão atual do T.A.M.K
        auto_critical: Se True, instala automaticamente atualizações críticas
    """
    updater = AutoUpdater(current_version)
    
    if auto_critical:
        if updater.check_and_install_critical():
            return
    
    latest = updater.checker.check_for_updates(force=False)
    if latest and latest.level not in [UpdateLevel.CRITICAL, UpdateLevel.PATCH]:
        from .update_checker import show_update_notification
        show_update_notification(latest, latest.level)


def interactive_update(current_version: str) -> bool:
    """
      Inicia atualização interativa (comando --update).
    
      Args:
        current_version: Versão atual do T.A.M.K
        
      Returns:
        True se atualização foi bem sucedida
    """
    updater = AutoUpdater(current_version)
    return updater.perform_update(automatic=False)
