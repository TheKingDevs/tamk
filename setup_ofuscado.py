#!/usr/bin/env python3
"""
Setup script para criar wheel do TAMK ofuscado.
Distribuição como pacote Python proprietário.
"""

import os
import sys
import shutil
import subprocess
from pathlib import Path
from setuptools import setup, find_packages

# Garantir que estamos no diretório correto
ROOT = Path(__file__).parent


def build_obfuscated():
    """Ofusca o código antes do build usando PyArmor 9+"""
    print("🔒 Ofuscando código fonte...")

    dist_dir = ROOT / "dist" / "obfuscated"
    if dist_dir.exists():
        shutil.rmtree(dist_dir)

    # PyArmor 9+: usa 'gen' com -O (output) e -r (recursive)
    # Rodamos de dentro de src/ para gerar: dist/obfuscated/main.py (sem prefixo src/)
    # Isso evita conflito de módulos no PyInstaller
    abs_output = dist_dir.resolve()
    result = subprocess.run(
        f"cd src && pyarmor gen -O {abs_output} -r .",
        shell=True,
        capture_output=True,
        text=True
    )

    if result.returncode != 0:
        print(f"❌ Erro no PyArmor:\n{result.stderr}")
        sys.exit(1)

    # Verificar estrutura gerada (main.py na raiz do obfuscated)
    main_file = dist_dir / "main.py"
    if not main_file.exists():
        print(f"❌ Arquivo esperado não encontrado: {main_file}")
        sys.exit(1)

    print("✅ Código ofuscado em dist/obfuscated/")
    return dist_dir


def prepare_package():
    """Prepara estrutura do pacote para wheel"""
    print("📦 Preparando pacote...")

    package_dir = ROOT / "tamk_proprietary"
    if package_dir.exists():
        shutil.rmtree(package_dir)
    package_dir.mkdir()

    # Ofuscar código fonte
    obfuscated = build_obfuscated()

    # Copiar código ofuscado (main.py e subpacotes ficam na raiz do obfuscated)
    for item in obfuscated.iterdir():
        if item.name == "pyarmor_runtime_000000":
            continue  # copiado separado abaixo
        dst = package_dir / item.name
        if item.is_dir():
            shutil.copytree(item, dst)
        else:
            shutil.copy2(item, dst)

    # Copiar runtime do PyArmor
    runtime_src = obfuscated / "pyarmor_runtime_000000"
    if runtime_src.exists():
        shutil.copytree(runtime_src, package_dir / "pyarmor_runtime_000000")

    # Criar __init__.py se não existir
    init_file = package_dir / "__init__.py"
    if not init_file.exists():
        init_file.write_text("")

    # Copiar assets para o pacote
    if (ROOT / "assets").exists():
        shutil.copytree(ROOT / "assets", package_dir / "assets")

    print(f"✅ Pacote preparado em {package_dir}")


def create_wheel():
    """Cria wheel do pacote"""
    print("🚀 Criando wheel...")

    # Preparar pacote
    prepare_package()

    # Configuração do setup
    setup(
        name="tamk-proprietary",
        version="2026.3.0",
        author="Shadow Developer",
        description="T.A.M.K - Termux APK Manager Kit (Proprietary)",
        long_description=(ROOT / "README.md").read_text(encoding='utf-8'),
        long_description_content_type="text/markdown",
        packages=find_packages(include=["tamk_proprietary", "tamk_proprietary.*"]),
        package_data={
            "tamk_proprietary": [
                "assets/**/*",
                "pyarmor_runtime_000000/*",
                "**/*.so",
            ],
        },
        include_package_data=True,
        install_requires=[
            "watchdog>=3.0.0",
            "websockets>=12.0",
        ],
        entry_points={
            "console_scripts": [
                "tamk=tamk_proprietary.main:main",
            ],
        },
        python_requires=">=3.9",
        classifiers=[
            "Development Status :: 4 - Beta",
            "Environment :: Console",
            "Intended Audience :: Developers",
            "License :: Other/Proprietary License",
            "Programming Language :: Python :: 3",
            "Topic :: Software Development :: Build Tools",
        ],
        license="Proprietary",
    )


if __name__ == "__main__":
    create_wheel()
