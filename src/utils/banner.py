"""
  Utilitário para exibição dinâmica de banners no terminal.
  Adapta-se automaticamente ao tamanho do terminal.
"""

import os
import re
import signal
import subprocess
from enum import Enum
from typing import Optional, List, Callable

from .colors import TColor


class BannerStyle(Enum):
    """Estilos predefinidos de banner."""
    DEFAULT = "default"
    SUCCESS = "success"
    ERROR = "error"
    WARNING = "warning"
    INFO = "info"
    HEADER = "header"


class Banner:
    """
      Gerador de banners dinâmicos para terminal.
    """

    def __init__(self):
        self._cols = self._get_terminal_width()

    def _get_terminal_width(self) -> int:
        """
          Obtém largura atual do terminal.
        """
        try:
            return int(subprocess.check_output(['tput', 'cols'], text=True).strip())
        except (subprocess.CalledProcessError, FileNotFoundError):
            return 80

    def _refresh_width(self) -> None:
        """Atualiza largura do terminal."""
        self._cols = self._get_terminal_width()

    def _strip_ansi(self, text: str) -> str:
        """Remove códigos ANSI de uma string."""
        return re.sub(r'\x1b\[[0-9;]*[mGKF]', '', text)

    def _center_text(self, text: str, fillchar: str = ' ') -> str:
        """
          Centraliza texto no terminal.
          
          Args:
            text: Texto a centralizar
            fillchar: Caractere de preenchimento
        """
        self._refresh_width()
        clean_text = self._strip_ansi(text)
        padding = max(0, (self._cols - len(clean_text)) // 2)
        return fillchar * padding + text

    def _create_separator(self, char: str = '=', color: Optional[str] = None, 
                          width: Optional[int] = None) -> str:
        """
          Cria linha separadora.
          
          Args:
            char: Caractere do separador
            color: Cor ANSI opcional
            width: Largura fixa (opcional)
        """
        self._refresh_width()
        sep_width = width if width else self._cols
        line = char * sep_width
        if color:
            return f"{color}{line}{TColor.RESET}"
        return line

    def _create_box(self, lines: List[str], 
                    border_char: str = '│', 
                    corner_top: str = '╔',
                    corner_bottom: str = '╚',
                    color: Optional[str] = None) -> str:
        """
          Cria caixa com bordas.
          
          Args:
            lines: Linhas de conteúdo
            border_char: Caractere da borda vertical
            corner_top: Caractere do canto superior
            corner_bottom: Caractere do canto inferior
            color: Cor ANSI opcional
        """
        self._refresh_width()
        
        if not lines:
            return ""
        
        max_width = max(len(self._strip_ansi(line)) for line in lines)
        padding = max(0, (self._cols - max_width - 4) // 2)
        
        top_border = f"{corner_top}{'═' * (max_width + 2)}╗"
        bottom_border = f"{corner_bottom}{'═' * (max_width + 2)}╝"
        
        if color:
            top_border = f"{color}{top_border}{TColor.RESET}"
            bottom_border = f"{color}{bottom_border}{TColor.RESET}"
        
        result = [self._center_text(top_border, fillchar=' ')]
        
        for line in lines:
            clean_line = self._strip_ansi(line)
            line_padding = max_width - len(clean_line)
            right_pad = line_padding // 2
            left_pad = line_padding - right_pad
            
            bordered = f"{border_char} {' ' * left_pad}{line}{' ' * right_pad} {border_char}"
            if color:
                bordered = f"{color}{bordered}{TColor.RESET}"
            
            result.append(self._center_text(bordered, fillchar=' '))
        
        result.append(self._center_text(bottom_border, fillchar=' '))
        
        return "\n".join(result)

    def show_art(self, text: str, 
                 font: str = "standard", 
                 effects: str = "metal",
                 fallback: bool = True) -> str:
        """
          Exibe arte ASCII usando toilet/figlet.
          
          Args:
            text: Texto a renderizar
            font: Fonte do toilet
            effects: Efeitos do toilet
            fallback: Se True, usa fallback simples em caso de erro
        """
        try:
            art = subprocess.getoutput(f"toilet -f {font} -F {effects} '{text}'")
            if art and not art.startswith("Cannot"):
                centered_lines = []
                for line in art.splitlines():
                    centered_lines.append(self._center_text(line))
                return "\n".join(centered_lines)
        except Exception:
            pass
        
        if fallback:
            bordered = f"=== {text} ==="
            return self._center_text(bordered)
        
        return ""

    def show(self, 
             title: Optional[str] = None,
             subtitle: Optional[str] = None,
             lines: Optional[List[str]] = None,
             style: BannerStyle = BannerStyle.DEFAULT,
             separator: str = '=',
             show_top: bool = True,
             show_bottom: bool = True) -> None:
        """
          Exibe banner completo.
          
          Args:
            title: Título principal (será centralizado)
            subtitle: Subtítulo (será centralizado)
            lines: Linhas adicionais de conteúdo
            style: Estilo predefinido do banner
            separator: Caractere do separador
            show_top: Mostra separador superior
            show_bottom: Mostra separador inferior
        """
        self._refresh_width()
        
        colors = {
            BannerStyle.DEFAULT: TColor.CYAN,
            BannerStyle.SUCCESS: TColor.GREEN,
            BannerStyle.ERROR: TColor.RED,
            BannerStyle.WARNING: TColor.YELLOW,
            BannerStyle.INFO: TColor.BLUE,
            BannerStyle.HEADER: TColor.MAGENTA
        }
        
        color = colors.get(style, TColor.CYAN)
        
        all_texts = []
        if title:
            all_texts.append(self._strip_ansi(title))
        if subtitle:
            all_texts.append(self._strip_ansi(subtitle))
        if lines:
            for line in lines:
                all_texts.append(self._strip_ansi(line))
        
        max_width = max(len(t) for t in all_texts) if all_texts else 40
        sep_width = max(min(max_width + 4, self._cols), 40)
        
        if show_top:
            print(self._center_text(self._create_separator(separator, color, sep_width)))
        
        if title:
            print(self._center_text(f"{color}{title}{TColor.RESET}"))
        
        if subtitle:
            print(self._center_text(f"{color}{subtitle}{TColor.RESET}"))
        
        if lines:
            for line in lines:
                print(self._center_text(line))
        
        if show_bottom:
            print(self._center_text(self._create_separator(separator, color, sep_width)))

    def show_box(self, 
                 lines: List[str],
                 style: BannerStyle = BannerStyle.DEFAULT,
                 title: Optional[str] = None) -> None:
        """
          Exibe conteúdo em caixa com bordas.
          
          Args:
            lines: Linhas de conteúdo
            style: Estilo predefinido
            title: Título opcional da caixa
        """
        colors = {
            BannerStyle.DEFAULT: TColor.CYAN,
            BannerStyle.SUCCESS: TColor.GREEN,
            BannerStyle.ERROR: TColor.RED,
            BannerStyle.WARNING: TColor.YELLOW,
            BannerStyle.INFO: TColor.BLUE,
            BannerStyle.HEADER: TColor.MAGENTA
        }
        
        color = colors.get(style, TColor.CYAN)
        
        if title:
            print(self._center_text(f"{color}┌─ {title} ─┐{TColor.RESET}"))
        
        print(self._create_box(lines, color=color))
        
        if title:
            print(self._center_text(f"{color}└{'─' * (len(title) + 4)}┘{TColor.RESET}"))

    def show_art_banner(self, 
                        art_text: str,
                        subtitle: Optional[str] = None,
                        separator: str = '=',
                        style: BannerStyle = BannerStyle.DEFAULT) -> None:
        """
          Exibe banner com arte ASCII + subtítulos.
          
          Args:
            art_text: Texto para arte ASCII
            subtitle: Subtítulo abaixo da arte
            separator: Caractere do separador
            style: Estilo do banner
        """
        colors = {
            BannerStyle.DEFAULT: TColor.CYAN,
            BannerStyle.SUCCESS: TColor.GREEN,
            BannerStyle.ERROR: TColor.RED,
            BannerStyle.WARNING: TColor.YELLOW,
            BannerStyle.INFO: TColor.BLUE,
            BannerStyle.HEADER: TColor.MAGENTA
        }
        
        color = colors.get(style, TColor.CYAN)
        
        art = self.show_art(art_text)
        if art:
            print(art)
        
        art_lines = art.splitlines() if art else []
        art_max_width = max(len(self._strip_ansi(line)) for line in art_lines) if art_lines else 40
        
        if subtitle:
            print(self._center_text(f"{color}{subtitle}{TColor.RESET}"))
            sep_width = max(min(max(art_max_width, len(subtitle) + 4), self._cols), 40)
        else:
            sep_width = max(min(art_max_width, self._cols), 40)
        
        print(self._center_text(self._create_separator(separator, color, sep_width)))


class DynamicBanner:
    """
      Banner dinâmico que se redesenha automaticamente ao redimensionar.
    """
    
    def __init__(self, title: str, subtitle: Optional[str] = None, 
                 style: BannerStyle = BannerStyle.DEFAULT,
                 separator_top: str = '=', separator_bottom: str = '=',
                 custom_separator_width: Optional[int] = None):
        self.title = title
        self.subtitle = subtitle
        self.style = style
        self.separator_top = separator_top
        self.separator_bottom = separator_bottom
        self.custom_separator_width = custom_separator_width
        self.banner = Banner()
        self._needs_redraw = True
        self._content_callback: Optional[Callable] = None
        
    def _handle_resize(self, signum, frame):
        """Handler para SIGWINCH."""
        self._needs_redraw = True
        
    def set_content_callback(self, callback: Callable):
        """Define callback para conteúdo dinâmico."""
        self._content_callback = callback
        
    def show_once(self):
        """Exibe banner uma vez (sem redraw automático)."""
        self.banner.show_art_banner(
            art_text=self.title,
            subtitle=self.subtitle,
            separator=self.separator_bottom,
            style=self.style
        )
    
    def show_custom(self, top_sep: Optional[str] = None, 
                    bottom_sep: Optional[str] = None,
                    extra_lines: Optional[List[str]] = None):
        """
          Exibe banner com separadores customizados.
          
          Args:
            top_sep: Separador superior (sobrescreve o padrão)
            bottom_sep: Separador inferior (sobrescreve o padrão)
            extra_lines: Linhas extras de conteúdo
        """
        self.banner._refresh_width()
        
        colors = {
            BannerStyle.DEFAULT: TColor.CYAN,
            BannerStyle.SUCCESS: TColor.GREEN,
            BannerStyle.ERROR: TColor.RED,
            BannerStyle.WARNING: TColor.YELLOW,
            BannerStyle.INFO: TColor.BLUE,
            BannerStyle.HEADER: TColor.MAGENTA
        }
        
        color = colors.get(self.style, TColor.CYAN)
        
        art = self.banner.show_art(self.title)
        if art:
            print(art)
        
        if self.custom_separator_width:
            sep_width = self.custom_separator_width
        elif extra_lines:
            all_widths = [len(self.banner._strip_ansi(line)) for line in extra_lines]
            if self.subtitle:
                all_widths.append(len(self.subtitle))
            max_width = max(all_widths) if all_widths else 40
            sep_width = max(min(max_width + 4, self.banner._cols), 40)
        else:
            art_lines = art.splitlines() if art else []
            art_max = max(len(self.banner._strip_ansi(line)) for line in art_lines) if art_lines else 40
            sep_width = max(min(art_max, self.banner._cols), 40)
        
        if top_sep:
            print(self.banner._center_text(
                self.banner._create_separator(top_sep, color, sep_width)
            ))
        
        if self.subtitle:
            print(self.banner._center_text(f"{color}{self.subtitle}{TColor.RESET}"))
        
        if extra_lines:
            for line in extra_lines:
                print(self.banner._center_text(line))
        
        if bottom_sep:
            print(self.banner._center_text(
                self.banner._create_separator(bottom_sep, color, sep_width)
            ))
    
    def show_continuous(self, render_callback: Optional[Callable] = None,
                        top_sep: str = '=', bottom_sep: str = '='):
        """
          Exibe banner com redraw automático ao redimensionar.
          
          Args:
            render_callback: Função opcional para renderizar conteúdo adicional
            top_sep: Separador superior
            bottom_sep: Separador inferior
        """
        signal.signal(signal.SIGWINCH, self._handle_resize)
        
        try:
            while True:
                if self._needs_redraw:
                    os.system('clear')
                    self.show_custom(
                        top_sep=top_sep,
                        bottom_sep=bottom_sep,
                        extra_lines=render_callback() if render_callback else None
                    )
                    self._needs_redraw = False
                
                import time
                time.sleep(0.1)
        except KeyboardInterrupt:
            pass
        finally:
            signal.signal(signal.SIGWINCH, signal.SIG_DFL)


def banner(title: str, 
           subtitle: Optional[str] = None,
           style: str = "default",
           separator: str = '=') -> None:
    """
      Função simplificada para exibir banner.
      
      Args:
        title: Título do banner
        subtitle: Subtítulo opcional
        style: Estilo (default, success, error, warning, info, header)
        separator: Caractere do separador
    """
    b = Banner()
    b.show(
        title=title,
        subtitle=subtitle,
        style=BannerStyle(style.lower()),
        separator=separator
    )


def banner_art(text: str,
               subtitle: Optional[str] = None,
               style: str = "default",
               font: str = "standard",
               effects: str = "metal") -> None:
    """
      Função simplificada para exibir banner com arte ASCII.
      
      Args:
        text: Texto para arte ASCII
        subtitle: Subtítulo opcional
        style: Estilo do banner
        font: Fonte do toilet
        effects: Efeitos do toilet
    """
    b = Banner()
    b.show_art_banner(
        art_text=text,
        subtitle=subtitle,
        style=BannerStyle(style.lower()),
        separator='='
    )


def banner_box(lines: List[str],
               style: str = "default",
               title: Optional[str] = None) -> None:
    """
      Função simplificada para exibir caixa com bordas.
      
      Args:
        lines: Linhas de conteúdo
        style: Estilo do banner
        title: Título opcional
    """
    b = Banner()
    b.show_box(
        lines=lines,
        style=BannerStyle(style.lower()),
        title=title
    )


def success_banner(title: str, subtitle: Optional[str] = None) -> None:
    """Exibe banner de sucesso."""
    banner(title, subtitle, style="success")


def error_banner(title: str, subtitle: Optional[str] = None) -> None:
    """Exibe banner de erro."""
    banner(title, subtitle, style="error")


def warning_banner(title: str, subtitle: Optional[str] = None) -> None:
    """Exibe banner de aviso."""
    banner(title, subtitle, style="warning")


def info_banner(title: str, subtitle: Optional[str] = None) -> None:
    """Exibe banner informativo."""
    banner(title, subtitle, style="info")


def header_art(text: str, subtitle: Optional[str] = None) -> None:
    """Exibe header com arte ASCII."""
    banner_art(text, subtitle, style="header")


def success_art(text: str, subtitle: Optional[str] = None) -> None:
    """Exibe arte de sucesso."""
    banner_art(text, subtitle, style="success")
