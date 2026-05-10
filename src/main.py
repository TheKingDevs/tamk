#!/usr/bin/env python3
import os
import sys
import argparse
from typing import Optional

# Garante que imports funcionam
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from config.tamk_config import Config
from utils.logger import log, set_log_level
from utils.auto_updater import run_update_check, interactive_update
from controllers.run_controller import RunController
from controllers.dev_controller import start_dev_mode
from controllers.project_manager import ProjectManager
from controllers.build_controller import BuildController
from controllers.setup_controller import SetupController
from controllers.install_controller import InstallController


def create_parser() -> argparse.ArgumentParser:
    """
      Cria parser de argumentos.
    """
    parser = argparse.ArgumentParser(
        prog='tamk',
        description="📱 T.A.M.K - Termux APK Manager Kit",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Exemplos:
  tamk --create              # Cria novo projeto
  tamk --build -p minhasenha # Builda projeto atual
  tamk --dev                 # Modo desenvolvimento com live reload
  tamk --run                 # Executa projeto console
        """
    )

    group = parser.add_mutually_exclusive_group()
    group.add_argument("-b", "--build", action="store_true",
                      help="Builda o projeto atual")
    group.add_argument("--create", action="store_true",
                      help="Inicia wizard de criação de projeto")
    group.add_argument("--setup", action="store_true",
                      help="Configura SDK e keystore global")
    group.add_argument("-r", "--run", nargs='?', const=True, metavar="ARQUIVO",
                      help="Executa projeto console ou arquivo Kotlin/Java")
    group.add_argument("-l", "--install", action="store_true",
                      help="Instala o APK gerado no dispositivo")
    group.add_argument("-i", "--ide", action="store_true",
                      help="Abre no SmartIDE")
    group.add_argument("-v", "--version", action="store_true",
                      help="Mostra versão")
    group.add_argument("-d", "--dev", action="store_true",
                      help="Inicia modo desenvolvimento com live reload")
    group.add_argument("-u", "--update", action="store_true",
                      help="Verifica e instala atualizações")
    group.add_argument("--no-update-check", action="store_true",
                      help="Desabilita verificação automática de atualizações")

    parser.add_argument("-p", "--password", metavar="SENHA",
                       help="Senha da keystore (será solicitada se não fornecida)")
    parser.add_argument("-V", "--verbose", action="store_true",
                       help="Modo detalhado (debug)")
    parser.add_argument("--no-ws", action="store_true",
                       help="Desabilita WebSocket no modo dev (usa HTTP)")
    parser.add_argument("--ws-port", type=int, default=8765, metavar="PORTA",
                       help="Porta do WebSocket (padrão: 8765)")

    return parser


def handle_build(args) -> int:
    """
      Handler para o comando ( --build ).
    """
    try:
        password = args.password
        
        controller = BuildController(verbose=args.verbose, password=password)
        controller.build_apk(password=password)
        return 0
    except Exception as e:
        log(f"Erro no build: {e}", "ERROR")
        return 1


def handle_create(args) -> int:
    """
      Handler para comando --create.
    """
    try:
        ProjectManager().start_wizard()
        return 0
    except KeyboardInterrupt:
        print("\nCancelado.")
        return 130
    except Exception as e:
        log(f"Erro ao criar projeto: {e}", "ERROR")
        return 1


def handle_setup(args) -> int:
    """
      Handler para o comando ( --setup. )
    """
    try:
        SetupController().setup_environment()
        return 0
    except Exception as e:
        log(f"Erro no setup: {e}", "ERROR")
        return 1


def handle_run(args) -> int:
    """
      Handler para o comando ( --run. )
    """
    try:
        file_to_run = None
        if args.run is not True and args.run is not None:
            file_to_run = str(args.run)
        
        controller = RunController(file_path=file_to_run, verbose=args.verbose)
        controller.execute_snippet()
        return 0
    except Exception as e:
        log(f"Erro na execução: {e}", "ERROR")
        return 1


def handle_install(args) -> int:
    """
      Handler para o comando ( --install. )
    """
    try:
        success = InstallController.install_apk()
        return 0 if success else 1
    except Exception as e:
        log(f"Erro na instalação: {e}", "ERROR")
        return 1


def handle_ide(args) -> int:
    """
      Handler para o comando ( --ide. )
    """
    try:
        conf = Config()
        log(f"Abrindo SmartIDE em: {conf.SMARTIDE_PATH}", "INFO")
        result = os.system("am start -n org.smartide.code/.MainActivity 2>/dev/null")
        return 0 if result == 0 else 1
    except Exception as e:
        log(f"Erro ao abrir IDE: {e}", "ERROR")
        return 1


def handle_dev(args) -> int:
    """
      Handler para o comando ( --dev. )
    """
    try:
        success = start_dev_mode(
            password=args.password,
            verbose=args.verbose,
            no_ws=args.no_ws,
            ws_port=args.ws_port
        )
        return 0 if success else 1
    except KeyboardInterrupt:
        print("\nEncerrado.")
        return 130
    except Exception as e:
        log(f"Erro no modo dev: {e}", "ERROR")
        return 1


def handle_version(args) -> int:
    """
      Handler para o comando ( --version. )
    """
    conf = Config()
    print(f"✨ T.A.M.K Version: {conf.VERSION}")
    print(f"   Environment: {conf.ENV}")
    print(f"   Home: {conf.TAMK_HOME}")
    return 0


def handle_update(args) -> int:
    """
      Handler para o comando ( --update. )
    """
    try:
        conf = Config()
        success = interactive_update(conf.VERSION)
        return 0 if success else 1
    except Exception as e:
        log(f"Erro na atualização: {e}", "ERROR")
        return 1


def main() -> int:
    """
      Função principal.
    """
    parser = create_parser()
    args = parser.parse_args()

    if args.verbose:
        set_log_level('DEBUG')

    try:
        conf = Config()
        
        # Verifica atualizações (exceto se --update ou --no-update-check)
        if not args.update and not args.no_update_check and not args.version:
            run_update_check(conf.VERSION, auto_critical=True)
        
        if args.version:
            return handle_version(args)
        elif args.setup:
            return handle_setup(args)
        elif args.create:
            return handle_create(args)
        elif args.run is not None:
            return handle_run(args)
        elif args.build:
            return handle_build(args)
        elif args.ide:
            return handle_ide(args)
        elif args.install:
            return handle_install(args)
        elif args.dev:
            return handle_dev(args)
        elif args.update:
            return handle_update(args)
        else:
            parser.print_help()
            return 0

    except KeyboardInterrupt:
        print("\n\nOperação cancelada pelo usuário.")
        return 130
    except Exception as e:
        log(f"Erro inesperado: {e}", "ERROR")
        if args.verbose:
            import traceback
            traceback.print_exc()
        return 1


if __name__ == "__main__":
    sys.exit(main())
