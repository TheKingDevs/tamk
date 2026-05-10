import os
import re
import shutil
from pathlib import Path
from typing import Optional

from config.tamk_config import Config

conf = Config()


class ConsoleStructure:
    """
      Estrutura de projeto para aplicativos de console Kotlin.
    """
    
    RESERVED_NAMES = {
        'src', 'libs', 'build', 'out', 'test', 'main', 'kotlin', 'java'
    }

    SEMVER_PATTERN = re.compile(
        r'^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)'
        r'(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?'
        r'(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'
    )

    def setup(self, name: str, version: str, author: str, 
              password: Optional[str] = None, web_url: Optional[str] = None) -> bool:
        """
          Configura estrutura de projeto Console.
        """
        safe_name = self._sanitize_name(name)
        
        is_valid, normalized = self._validate_version(version)
        if not is_valid:
            print(f"❌ Versão inválida: '{version}'")
            print("   Formato esperado: MAJOR.MINOR.PATCH (ex: 1.0.0)")
            return False
        
        version = normalized
        
        base_path = os.path.join(os.getcwd(), safe_name)
        
        if os.path.exists(base_path):
            counter = 1
            original = base_path
            while os.path.exists(base_path):
                base_path = f"{original}_{counter}"
                counter += 1

        folders = ["src", "libs"]
        for f in folders:
            os.makedirs(os.path.join(base_path, f), exist_ok=True)

        template_path = os.path.join(conf.get_template_dir("console"), "Main.kt.tmpl")

        if os.path.exists(template_path):
            with open(template_path, "r", encoding='utf-8') as f:
                content = f.read()
            
            content = content.replace("{{NAME}}", safe_name)
            content = content.replace("{{VERSION}}", version)
            content = content.replace("{{AUTHOR}}", author)
        else:
            content = f"""/**
 * Project: {safe_name}
 * Version: {version}
 * Author: {author}
 * Tipo: Console (Kotlin)
 */
fun main() {{
    println("Olá do Console T.A.M.K!")
    println("Projeto: {safe_name}")
    println("Versão: {version}")
}}
"""

        main_kt_path = os.path.join(base_path, "src", "Main.kt")
        with open(main_kt_path, "w", encoding='utf-8') as f:
            f.write(content)

        config_path = os.path.join(base_path, "tamk.config")
        with open(config_path, "w", encoding='utf-8') as f:
            f.write(f"type=console\nname={safe_name}\nversion={version}\nauthor={author}\n")

        print(f"✅ Projeto Console '{safe_name}' criado em: {base_path}")
        print(f"   Execute: cd {os.path.basename(base_path)} && tamk --run")
        return True

    def _sanitize_name(self, name: str) -> str:
        """  
          Sanitiza nome do projeto.
        """
        if not name:
            raise ValueError("Nome não pode ser vazio")
        clean = re.sub(r'[^a-zA-Z0-9\-_]', '', name.strip())[:50]
        if not clean:
            raise ValueError("Nome inválido")
        if clean[0].isdigit():
            clean = "app_" + clean
        return clean

    def _validate_version(self, version: str) -> tuple[bool, str]:
        """
          Valida formato semver.
        """
        if not version:
            return False, "1.0.0"
        
        version = version.strip()
        if version.startswith('v') or version.startswith('V'):
            version = version[1:]
        
        if not self.SEMVER_PATTERN.match(version):
            return False, version
        
        return True, version
