"""
  Sistema de verificação de atualizações com níveis de prioridade.
  Detecta atualizações críticas, comuns e opcionais.
"""

import json
import re
import subprocess
import urllib.request
import urllib.error
from typing import Optional, Dict, Any, Tuple
from enum import Enum
from datetime import datetime, timedelta

from .colors import TColor
from .banner import Banner, BannerStyle, warning_banner, success_banner, info_banner


class UpdateLevel(Enum):
    """Níveis de prioridade para atualizações."""
    CRITICAL = "critical"      # Patch de segurança/bug crítico - auto install
    MAJOR = "major"            # Nova funcionalidade importante - perguntar
    MINOR = "minor"            # Melhorias menores - perguntar
    PATCH = "patch"            # Correções menores - auto install
    OPTIONAL = "optional"      # Atualização cosmética - apenas notificar


class VersionInfo:
    """Informações sobre uma versão."""
    
    def __init__(self, version_str: str, release_notes: str = "", 
                 level: UpdateLevel = UpdateLevel.MINOR,
                 release_date: Optional[datetime] = None):
        self.version_str = version_str
        self.release_notes = release_notes
        self.level = level
        self.release_date = release_date or datetime.now()
        self.major, self.minor, self.patch = self._parse_version(version_str)
    
    def _parse_version(self, version_str: str) -> Tuple[int, int, int]:
        """Extrai componentes da versão (major, minor, patch)."""
        clean_version = re.sub(r'[-+].*$', '', version_str)
        parts = clean_version.split('.')
        
        try:
            major = int(parts[0]) if len(parts) > 0 else 0
            minor = int(parts[1]) if len(parts) > 1 else 0
            patch = int(parts[2]) if len(parts) > 2 else 0
            return major, minor, patch
        except ValueError:
            return 0, 0, 0
    
    def __gt__(self, other: 'VersionInfo') -> bool:
        """Compara se esta versão é maior que outra."""
        if self.major != other.major:
            return self.major > other.major
        if self.minor != other.minor:
            return self.minor > other.minor
        return self.patch > other.patch
    
    def __eq__(self, other: 'VersionInfo') -> bool:
        """Compara se versões são iguais."""
        return (self.major == other.major and 
                self.minor == other.minor and 
                self.patch == other.patch)
    
    def __str__(self) -> str:
        return self.version_str


class UpdateChecker:
    """
      Verificador de atualizações com suporte a níveis de prioridade.
    """
    
    GITHUB_API_URL = "https://api.github.com/repos/Shadw-Developer/tamk/releases"
    GITHUB_REPO_URL = "https://github.com/Shadw-Developer/tamk"
    CACHE_DURATION_HOURS = 6
    
    def __init__(self, current_version: str):
        self.current_version = VersionInfo(current_version)
        self.cache_file = self._get_cache_path()
        self._latest_release: Optional[Dict[str, Any]] = None
        self._last_check: Optional[datetime] = None
    
    def _get_cache_path(self) -> str:
        """Obtém caminho para arquivo de cache."""
        import os
        tamk_home = os.environ.get('TAMK_HOME', '')
        if tamk_home:
            cache_dir = os.path.join(tamk_home, '.cache')
        else:
            cache_dir = os.path.expanduser('~/.tamk_cache')
        
        os.makedirs(cache_dir, exist_ok=True)
        return os.path.join(cache_dir, 'update_cache.json')
    
    def _load_from_cache(self) -> Optional[Dict[str, Any]]:
        """Carrega dados do cache se ainda válidos."""
        try:
            with open(self.cache_file, 'r') as f:
                data = json.load(f)
            
            check_time = datetime.fromisoformat(data['check_time'])
            if datetime.now() - check_time < timedelta(hours=self.CACHE_DURATION_HOURS):
                self._last_check = check_time
                return data.get('release')
        except (FileNotFoundError, json.JSONDecodeError, KeyError):
            pass
        return None
    
    def _save_to_cache(self, release_data: Dict[str, Any]) -> None:
        """Salva dados no cache."""
        try:
            cache_data = {
                'check_time': datetime.now().isoformat(),
                'release': release_data
            }
            with open(self.cache_file, 'w') as f:
                json.dump(cache_data, f, indent=2)
        except Exception:
            pass
    
    def _detect_update_level(self, latest_version: VersionInfo, 
                              release_body: str) -> UpdateLevel:
        """
          Detecta o nível da atualização baseado na versão e notas de release.
        
          Regras:
            - CRITICAL: Contém "critical", "security", "urgent" nas notas
            - PATCH: Apenas correções de bugs menores
            - MAJOR: Nova funcionalidade significativa
            - MINOR: Melhorias incrementais
        """
        body_lower = release_body.lower()
        
        critical_keywords = ['critical', 'security', 'urgent', 'importante', 'obrigatória']
        if any(keyword in body_lower for keyword in critical_keywords):
            return UpdateLevel.CRITICAL
        
        patch_keywords = ['patch', 'bugfix', 'correção', 'fix', 'hotfix']
        if any(keyword in body_lower for keyword in patch_keywords):
            if latest_version.major == self.current_version.major and \
               latest_version.minor == self.current_version.minor:
                return UpdateLevel.PATCH
        
        if latest_version.major > self.current_version.major:
            return UpdateLevel.MAJOR
        
        if latest_version.minor > self.current_version.minor:
            return UpdateLevel.MINOR
        
        return UpdateLevel.MINOR
    
    def check_for_updates(self, force: bool = False) -> Optional[VersionInfo]:
        """
          Verifica se há atualizações disponíveis.
        
          Args:
            force: Se True, ignora o cache e força nova verificação
            
          Returns:
            VersionInfo da nova versão ou None se estiver atualizado
        """

        if not force:
            cached = self._load_from_cache()
            if cached:
                self._latest_release = cached
                latest = VersionInfo(
                    cached['tag_name'].lstrip('v') if cached.get('tag_name') else cached.get('name', '0.0.0'),
                    cached.get('body', ''),
                    self._detect_update_level(
                        VersionInfo(cached['tag_name'].lstrip('v') if cached.get('tag_name') else '0.0.0', 
                                   cached.get('body', '')),
                        cached.get('body', '')
                    )
                )
                if latest > self.current_version:
                    return latest
                return None
        
        try:
            req = urllib.request.Request(
                self.GITHUB_API_URL,
                headers={'User-Agent': 'TAMK-Update-Checker'}
            )
            
            with urllib.request.urlopen(req, timeout=10) as response:
                releases = json.loads(response.read().decode('utf-8'))
            
            if not releases:
                return None
            
            for release in releases:
                if release.get('draft') or release.get('prerelease'):
                    continue
                
                self._latest_release = release
                version_str = release.get('tag_name', '').lstrip('v') or release.get('name', '0.0.0')
                release_body = release.get('body', '')
                
                latest_version = VersionInfo(version_str, release_body)
                latest_version.level = self._detect_update_level(latest_version, release_body)
                
                self._save_to_cache(release)
                
                if latest_version > self.current_version:
                    return latest_version
            
            return None
            
        except (urllib.error.URLError, urllib.error.HTTPError, json.JSONDecodeError, Exception) as e:
            return None
    
    def get_update_info(self) -> Optional[Dict[str, Any]]:
        """Obtém informações detalhadas sobre a atualização."""
        if not self._latest_release:
            return None
        
        return {
            'version': self._latest_release.get('tag_name', '').lstrip('v'),
            'name': self._latest_release.get('name', ''),
            'body': self._latest_release.get('body', ''),
            'url': self._latest_release.get('html_url', ''),
            'published_at': self._latest_release.get('published_at', ''),
            'level': self._detect_update_level(
                VersionInfo(self._latest_release.get('tag_name', '').lstrip('v', '0.0.0')),
                self._latest_release.get('body', '')
            ).value
        }
    
    def get_download_url(self) -> Optional[str]:
        """Obtém URL de download do release."""
        if not self._latest_release:
            return None
        
        assets = self._latest_release.get('assets', [])
        if assets:
            return assets[0].get('browser_download_url')
        
        return self._latest_release.get('html_url', self.GITHUB_REPO_URL)


def format_update_message(latest_version: VersionInfo, level: UpdateLevel) -> str:
    """Formata mensagem de atualização baseada no nível."""
    messages = {
        UpdateLevel.CRITICAL: f"⚠️  ATUALIZAÇÃO CRÍTICA: v{latest_version} disponível",
        UpdateLevel.PATCH: f"🔧 Patch disponível: v{latest_version}",
        UpdateLevel.MAJOR: f"🚀 Nova versão maior: v{latest_version}",
        UpdateLevel.MINOR: f"✨ Nova versão: v{latest_version}",
        UpdateLevel.OPTIONAL: f"📦 Atualização disponível: v{latest_version}"
    }
    return messages.get(level, f"📦 Atualização disponível: v{latest_version}")


def show_update_notification(latest_version: VersionInfo, level: UpdateLevel,
                             release_notes: str = "") -> None:
    """
      Exibe notificação visual de atualização.
    
      Args:
        latest_version: Informações da nova versão
        level: Nível da atualização
        release_notes: Notas de release
    """
    banner = Banner()
    
    if level == UpdateLevel.CRITICAL:
        lines = [
            f"{TColor.RED}{TColor.BOLD}Uma atualização CRÍTICA está disponível!{TColor.RESET}",
            f"",
            f"Versão atual: {TColor.YELLOW}v{latest_version}{TColor.RESET}",
            f"Nova versão: {TColor.GREEN}v{latest_version}{TColor.RESET}",
            f"",
            f"{TColor.BOLD}Esta atualização será instalada automaticamente.{TColor.RESET}"
        ]
        banner.show_box(lines, style=BannerStyle.ERROR, title="⚠️ ATUALIZAÇÃO CRÍTICA")
        
    elif level == UpdateLevel.PATCH:
        lines = [
            f"{TColor.CYAN}Um patch de correção está disponível!{TColor.RESET}",
            f"",
            f"Versão: {TColor.GREEN}v{latest_version}{TColor.RESET}",
            f"",
            f"{TColor.BOLD}Este patch será instalado automaticamente.{TColor.RESET}"
        ]
        banner.show_box(lines, style=BannerStyle.INFO, title="🔧 PATCH DISPONÍVEL")
        
    elif level in [UpdateLevel.MAJOR, UpdateLevel.MINOR]:
        lines = [
            f"{TColor.GREEN}Nova versão disponível!{TColor.RESET}",
            f"",
            f"Versão: {TColor.CYAN}v{latest_version}{TColor.RESET}",
            f"",
            f"Execute {TColor.BOLD}'tamk --update'{TColor.RESET} para atualizar.",
        ]
        banner.show_box(lines, style=BannerStyle.SUCCESS, title="✨ ATUALIZAÇÃO")
        
    else:
        info_banner("Atualização Disponível", f"v{latest_version} disponível")


def check_and_notify(current_version: str, force: bool = False) -> Optional[VersionInfo]:
    """
      Função convenience para verificar e notificar sobre atualizações.
    
      Args:
        current_version: Versão atual do T.A.M.K
        force: Força verificação ignorando cache
        
      Returns:
        VersionInfo se houver atualização, None caso contrário
    """
    checker = UpdateChecker(current_version)
    latest = checker.check_for_updates(force=force)
    
    if latest:
        show_update_notification(latest, latest.level, checker.get_update_info().get('body', '') if checker.get_update_info() else '')
    
    return latest
