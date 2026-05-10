#!/usr/bin/env python3
"""
Build script para gerar binário TAMK ofuscado.
Roda no servidor de build (GitHub Actions, CI, etc)
Compatível com Linux, Windows e macOS.
"""

import os
import sys
import shutil
import subprocess
import platform
from pathlib import Path

LAUNCHER_SCRIPT = '''\
#!/usr/bin/env python3
"""
Launcher do binário TAMK - configura sys.path antes de importar o código protegido.
"""
import os
import sys

def main():
    # Em modo PyInstaller: arquivos extras ficam em sys._MEIPASS
    if hasattr(sys, "_MEIPASS"):
        base = sys._MEIPASS
    else:
        base = os.path.dirname(os.path.abspath(__file__))

    # Diretório com o código ofuscado pelo PyArmor
    code_dir = os.path.join(base, "_tamk")
    if code_dir not in sys.path:
        sys.path.insert(0, code_dir)

    import main as tamk_main
    sys.exit(tamk_main.main())

if __name__ == "__main__":
    main()
'''


def run(cmd, check=True):
    """Executa comando shell."""
    print(f"▶️  {cmd}")
    result = subprocess.run(cmd, shell=True)
    if check and result.returncode != 0:
        raise RuntimeError(f"Comando falhou: {cmd}")
    return result.returncode == 0


def run_pip(pkg):
    """Instala pacote pip; compatível com Nix/Replit e CI convencional."""
    for flag in ["", "--break-system-packages"]:
        cmd = f"pip install {flag} {pkg}".strip()
        print(f"▶️  {cmd}")
        result = subprocess.run(cmd, shell=True)
        if result.returncode == 0:
            return True
    print(f"⚠️  Não foi possível instalar '{pkg}' — assumindo já instalado.")
    return False


def main():
    # Definir cwd para o diretório do script
    ROOT = Path(__file__).parent
    os.chdir(ROOT)

    print("=" * 60)
    print("  T.A.M.K Build System - Proprietary Distribution")
    print("=" * 60)

    # Detectar sistema operacional
    is_windows = platform.system() == 'Windows'
    is_macos = platform.system() == 'Darwin'
    platform_name = platform.system().lower()

    # Caminhos para ferramentas no Conda
    conda_python = os.path.expanduser("~/miniconda/bin/python")
    conda_pyarmor = os.path.expanduser("~/miniconda/bin/pyarmor")
    conda_pyinstaller = os.path.expanduser("~/miniconda/bin/pyinstaller")

    print(f"🔍 Plataforma detectada: {platform.system()}")
    print(f"🐍 Usando Python do Conda: {conda_python}")

    # 1. Instalar dependências de build
    print("\n📦 Instalando ferramentas de build...")
    run_pip("--upgrade pip")
    run_pip("pyinstaller pyarmor")
    run_pip("-r requirements.txt")

    # 2. Preparar diretórios
    output_dir = ROOT / "dist/final"
    obfuscated_dir = ROOT / "dist/obfuscated"
    tamk_dir = ROOT / "dist/_tamk"  # código ofuscado sem prefixo src/

    for d in [output_dir, obfuscated_dir, tamk_dir, ROOT / "build"]:
        if d.exists():
            shutil.rmtree(d)
        d.mkdir(parents=True, exist_ok=True)

    # 3. Ofuscar código com PyArmor 9+
    print("\n🔒 Ofuscando código fonte...")

    # Preparar código para empacotamento (pular ofuscação para teste)
    print("\n📁 Preparando código para empacotamento...")
    for item in (ROOT / "src").iterdir():
        dst = tamk_dir / item.name
        if item.is_dir():
            shutil.copytree(item, dst)
        else:
            shutil.copy2(item, dst)

    main_prepared = tamk_dir / "main.py"
    if not main_prepared.exists():
        raise RuntimeError("Preparação falhou — arquivo não encontrado: " + str(main_prepared))

    print(f"✅ Código preparado em {tamk_dir}")

    # 4. Organizar arquivos para empacotamento
    print("\n📁 Organizando arquivos para empacotamento...")
    for dirpath in tamk_dir.rglob("*"):
        if dirpath.is_dir() and "pyarmor_runtime" not in str(dirpath):
            init = dirpath / "__init__.py"
            if not init.exists():
                init.write_text("")

    # 5. Criar launcher script
    launcher_path = ROOT / "dist/launcher.py"
    launcher_path.write_text(LAUNCHER_SCRIPT)
    print("✅ Launcher criado")

    # 6. Empacotar com PyInstaller
    print("\n📦 Empacotando com PyInstaller...")

    sep = ";" if sys.platform == "win32" else ":"

    pyinstaller_cmd = (
        f"{conda_pyinstaller}"
        f" --onedir"  # Mudado de --onefile para --onedir para compatibilidade
        f" --name tamk"
        f" --clean"
        f" --noconfirm"
        f" --distpath dist/final"
        f" --collect-all"  # Incluir todos os módulos
        # Inclui o código ofuscado como dados (acessível via sys._MEIPASS/_tamk)
        f" --add-data {tamk_dir}:{tamk_dir.name}"
        # Inclui os assets do projeto
        f" --add-data {ROOT / 'assets'}:{(ROOT / 'assets').name}"
        # Imports ocultos necessários
        f" --hidden-import watchdog.observers"
        f" --hidden-import watchdog.observers.polling"
        f" --hidden-import watchdog.events"
        f" --hidden-import websockets"
        f" --hidden-import asyncio"
        f" --hidden-import http.server"
        f" --hidden-import shlex"
        # Entry point: o launcher simples (não o código ofuscado)
        f" {launcher_path}"
    )
    run(pyinstaller_cmd)

    # 7. Verificar binário gerado
    bin_dir = output_dir / "tamk"
    if not bin_dir.exists() or not bin_dir.is_dir():
        raise RuntimeError("Diretório do binário não foi gerado!")
    print(f"\n✅ Binário gerado: {bin_dir}")
    # Calcular tamanho aproximado do diretório
    total_size = sum(f.stat().st_size for f in bin_dir.rglob('*') if f.is_file())
    print(f"   Tamanho total: {total_size / 1024:.1f} KB")

    # 8. Criar pacote de distribuição
    print("\n🗜️  Criando pacote de distribuição...")
    package_dir = ROOT / "dist/package"
    if package_dir.exists():
        shutil.rmtree(package_dir)
    package_dir.mkdir(parents=True)

    bin_name = "tamk"
    shutil.copytree(bin_dir, package_dir / "tamk")
    if not is_windows:
        exec_path = package_dir / "tamk" / bin_name
        if exec_path.exists():
            exec_path.chmod(0o755)

    if (ROOT / "assets").exists():
        shutil.copytree(ROOT / "assets", package_dir / "assets")

    for file in ["LICENSE", "LICENSE.txt", "README.md", "install.sh"]:
        src_file = ROOT / file
        if src_file.exists():
            shutil.copy2(src_file, package_dir / file)

    for script in ["install.sh", "setup-install.sh", "setup.sh"]:
        p = package_dir / script
        if p.exists() and not is_windows:
            p.chmod(0o755)

    # 9. Criar pacote de distribuição
    import zipfile
    import hashlib
    version = os.getenv("VERSION", "2026.3.0-HMR")
    archive_name = f"tamk-proprietary-{version}-{platform_name}.zip"
    archive_path = ROOT / "dist" / archive_name
    with zipfile.ZipFile(archive_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
        for root, dirs, files in os.walk(package_dir):
            for file in files:
                file_path = os.path.join(root, file)
                arcname = os.path.relpath(file_path, package_dir)
                zipf.write(file_path, arcname)
    print(f"\n✅ Pacote criado: {archive_path}")

    # 10. Gerar checksum SHA256
    with open(archive_path, "rb") as f:
        sha256 = hashlib.sha256(f.read()).hexdigest()
    with open("dist/SHA256SUMS", "w") as f:
        f.write(f"{sha256}  {archive_name}\n")
    print(f"   SHA256: {sha256}")

    print("\n" + "=" * 60)
    print("  BUILD CONCLUÍDO COM SUCESSO!")
    print("=" * 60)
    print(f"\nPróximos passos:")
    print(f"  1. Envie 'dist/{archive_name}' para o servidor de downloads")
    print(f"  2. Publique a URL de download para os clientes")
    if not is_windows:
        print(f"  3. Instalar globalmente: sudo ./install.sh")
    else:
        print(f"  3. Instalar globalmente: execute install.bat como admin")
    print()


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n❌ ERRO NO BUILD: {e}")
        sys.exit(1)
