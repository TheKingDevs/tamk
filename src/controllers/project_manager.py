import os
import re
import sys
import subprocess
from getpass import getpass
from time import sleep as delay
from typing import Dict, Optional

from utils.logger import log
from utils.colors import TColor, ask_factory
from organization.factory import ProjectFactory
from utils.banner import banner_art, success_art, Banner, BannerStyle, DynamicBanner


class ProjectManager:
    """
      Gerenciador interativo de criação de projetos.
    """

    def __init__(self):
        self.options: Dict[str, str] = {
            "1": "console",
            "2": "ui_apk",
            "3": "webapp"
        }
        self._banner_instance = DynamicBanner(
            title="T.A.M.K",
            subtitle="Termux Apk Manager Kit • Factory (2026)",
            style=BannerStyle.DEFAULT,
            separator_top='═',
            separator_bottom='═'
        )

    def get_cols(self) -> int:
        """
          Obtém largura do terminal.
        """
        try:
            return int(subprocess.check_output(['tput', 'cols'], text=True).strip())
        except (subprocess.CalledProcessError, FileNotFoundError):
            return 80

    def center_block(self, text: str) -> str:
        """
          Centraliza texto no terminal.
        """
        cols = self.get_cols()
        output = []
        for line in text.splitlines():
            clean_line = re.sub(r'\x1b\[[0-9;]*[mGKF]', '', line)
            padding = max(0, (cols - len(clean_line)) // 2)
            output.append(" " * padding + line)
        return "\n".join(output)

    def show_banner(self):
        """
          Exibe banner inicial.
        """
        self._banner_instance.show_custom(
            top_sep='═',
            bottom_sep='═'
        )
        print("")

    def _clear_lines(self, count: int):
        """
          Limpa linhas do terminal.
        """
        for _ in range(count):
            sys.stdout.write(f"{TColor.UP}{TColor.CLEAR_LINE}")
        sys.stdout.flush()

    def _validate_project_name(self, name: str) -> bool:
        """
          Valida nome do projeto.
        """
        if not name or len(name.strip()) < 2:
            log("Nome deve ter pelo menos 2 caracteres", "ERROR")
            return False
        if name.startswith(('-', '_')):
            log("Nome não pode começar com hífen ou underscore", "ERROR")
            return False
        if not re.match(r'^[a-zA-Z0-9\-_]+$', name):
            log("Nome contém caracteres inválidos (use apenas letras, números, hífen e underscore)", "ERROR")
            return False
        return True

    def _validate_version(self, version: str) -> tuple[bool, str]:
        """
          Valida versão no wizard.
        """
        import re
        pattern = re.compile(
            r'^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)'
            r'(?:-([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'
        )
        
        version = version.strip()
        if version.startswith('v') or version.startswith('V'):
            version = version[1:]
        
        if not pattern.match(version):
            return False, version
        
        return True, version

    def start_wizard(self) -> None:
        """
          Inicia assistente interativo de criação.
        """
        os.system('clear')
        self.show_banner()

        while True:
            name = ask_factory("Nome do projeto", "MyApp")
            self._clear_lines(1)
            if self._validate_project_name(name):
                break
            self._clear_lines(1)
        
        author = ask_factory("Nome do Desenvolvedor", os.getenv("USER", "Developer"))
        self._clear_lines(1)

        while True:
            version = ask_factory("Versão do Release", "1.0.0")
            is_valid, normalized = self._validate_version(version)
            self._clear_lines(1)
            
            if is_valid:
                version = normalized
                break
            else:
                log(f"Versão inválida: '{version}'", "ERROR")
                log("Use formato: MAJOR.MINOR.PATCH (ex: 1.0.0, 2.3.1-beta)", "INFO")
                delay(4.5)
                self._clear_lines(2)

        cols = self.get_cols()
        indent = " " * int(cols * 0.02)
        print(f"{TColor.BOLD}{TColor.YELLOW}SELECIONE A ENGINE:{TColor.RESET}\n")
        print(f"{TColor.CYAN}[1] Standard Console{TColor.RESET}")
        print(f"{indent}└─ Scripts e automações CLI\n")
        print(f"{TColor.GREEN}[2] Native Android (UI/APK){TColor.RESET}")
        print(f"{indent}└─ Interface nativa XML/Kotlin\n")
        print(f"{TColor.MAGENTA}[3] Universal WebApp (HTML/JS){TColor.RESET}")
        print(f"{indent}└─ WebView híbrido\n")
        
        choice = ask_factory(f"Engine (1-3, padrão: 2)", "2")
        p_type = self.options.get(choice.strip(), "ui_apk")
        self._clear_lines(12)

        web_url = "file:///android_asset/index.html"
        if p_type == "webapp":
            print(f"{TColor.YELLOW}TIPO DE CONTEÚDO WEB:{TColor.RESET}")
            print(f"{TColor.CYAN}[1] Interno{TColor.RESET} (Pasta src/main/assets)")
            print(f"{TColor.CYAN}[2] Externo{TColor.RESET} (URL Remota)")
            web_choice = ask_factory("Opção", "1")
            
            if web_choice.strip() == "2":
                while True:
                    url = ask_factory("URL da aplicação", "https://example.com")
                    if url.startswith(('http://', 'https://')):
                        web_url = url
                        break
                    log("URL deve começar com http:// ou https://", "ERROR")
                    self._clear_lines(2)
            self._clear_lines(4)

        password: Optional[str] = None
        if p_type in ["ui_apk", "webapp"]:
            print(f"{TColor.YELLOW}⚠ SEGURANÇA: Chave para assinatura do APK{TColor.RESET}")
            print(f"{TColor.GRAY}(Mínimo 6 caracteres){TColor.RESET}")
            
            while True:
                try:
                    password = getpass(f"{TColor.CYAN}❯{TColor.RESET} Senha da Keystore: ")
                    if len(password) >= 6:
                        break
                    log("Senha deve ter pelo menos 6 caracteres", "ERROR")
                except (EOFError, KeyboardInterrupt):
                    print(f"\n{TColor.RED}Cancelado.{TColor.RESET}")
                    return
            
            self._clear_lines(3)

        print(f"🚀 {TColor.GREEN}Criando projeto {p_type}...{TColor.RESET}")
        
        print(f"Nome: {name}")
        print(f"Autor: {author}")
        print(f"Versão: {version}")
        if p_type == "webapp":
            print(f"URL: {web_url}")
        print("")

        success = ProjectFactory.create(
            p_type=p_type,
            name=name,
            version=version,
            author=author,
            password=password,
            web_url=web_url
        )

        if success:
            self._clear_lines(1)
            success_banner = DynamicBanner(
                title="PRONTO",
                subtitle=f"PROJETO '{name.upper()}' CRIADO!",
                style=BannerStyle.SUCCESS,
                separator_top='═',
                separator_bottom='═'
            )
            success_banner.show_custom(top_sep='═', bottom_sep='═')
            print(f"\n{TColor.YELLOW}Próximo passo:{TColor.RESET} cd {name} && tamk --build\n")
        else:
            log("Falha na criação do projeto. Verifique os logs acima.", "ERROR")


if __name__ == "__main__":
    ProjectManager().start_wizard()
