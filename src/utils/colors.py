import re
import sys
from typing import Optional


class TColor:
    """
      Cores e estilos para terminal.
    """
    
    # Cores Primárias
    CYAN = '\033[96m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    MAGENTA = '\033[95m'
    BLUE = '\033[94m'
    RED = '\033[91m'
    
    # Tons de Interface
    GRAY = '\033[90m'
    RESET = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
    
    # Controle de Terminal
    CLEAR_LINE = '\033[K'
    UP = '\033[F'
    HIDE_CURSOR = '\033[?25l'
    SHOW_CURSOR = '\033[?25h'


def strip_ansi(text: str) -> str:
    """
      Remove códigos ANSI de string.
    """
    return re.sub(r'\x1b\[[0-9;]*[mGKF]', '', text)


def ask_factory(question: str, default: str, password: bool = False) -> str:
    """
      Prompt interativo com estilo moderno.
    
      Args:
        question: Pergunta a ser feita
        default: Valor padrão
        password: Se True, oculta input
    """
    arrow = f"{TColor.CYAN}❯{TColor.RESET}"
    dot = f"{TColor.GREEN}•{TColor.RESET}"
    
    prompt = (
        f"{dot} {TColor.BOLD}{question}{TColor.RESET} "
        f"{TColor.GRAY}({default}){TColor.RESET} "
        f"{TColor.CYAN}~{arrow}{TColor.RESET} "
    )
    
    try:
        if password:
            from getpass import getpass
            clean_prompt = strip_ansi(prompt)
            user_input = getpass(clean_prompt)
        else:
            user_input = input(prompt).strip()
        
        return user_input if user_input else default
        
    except KeyboardInterrupt:
        print(f"\n{TColor.RED}Sessão encerrada.{TColor.RESET}")
        sys.exit(0)
    except EOFError:
        return default


def confirm(question: str, default: bool = False) -> bool:
    """
      Pergunta de confirmação sim/não.
    """
    suffix = "[Y/n]" if default else "[y/N]"
    prompt = f"{TColor.YELLOW}?{TColor.RESET} {question} {TColor.GRAY}{suffix}{TColor.RESET} "
    
    try:
        response = input(prompt).strip().lower()
        if not response:
            return default
        return response in ('y', 'yes', 's', 'sim')
    except KeyboardInterrupt:
        print()
        sys.exit(0)
    except EOFError:
        return default


def print_success(message: str) -> None:
    """
      Imprime mensagem de sucesso.
    """
    print(f"{TColor.GREEN}✔{TColor.RESET} {message}")


def print_error(message: str) -> None:
    """
      Imprime mensagem de erro.
    """
    print(f"{TColor.RED}✖{TColor.RESET} {message}")


def print_warning(message: str) -> None:
    """
      Imprime mensagem de aviso.
    """
    print(f"{TColor.YELLOW}⚠{TColor.RESET} {message}")


def print_info(message: str) -> None:
    """
      Imprime mensagem informativa.
    """
    print(f"{TColor.BLUE}ℹ{TColor.RESET} {message}")
