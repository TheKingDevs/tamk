from typing import Optional
from organization.structures.ui_apk import UIAppStructure
from organization.structures.webapp import WebAppStructure
from organization.structures.console import ConsoleStructure


class ProjectFactory:
    """
      Fábrica de projetos T.A.M.K.
    
      Cria estruturas de projeto de forma segura e validada.
    """
    
    _structures = {
        "ui_apk": UIAppStructure,
        "console": ConsoleStructure,
        "webapp": WebAppStructure
    }

    @classmethod
    def create(cls, p_type: str, name: str, version: str, author: str, 
               password: Optional[str] = None, web_url: Optional[str] = None) -> bool:
        """
          Cria projeto do tipo especificado.
        
          Args:
            p_type: Tipo de projeto ('ui_apk', 'console', 'webapp')
            name: Nome do projeto
            version: Versão
            author: Autor
            password: Senha da keystore (para ui_apk e webapp)
            web_url: URL para webapp (file:// ou http(s)://)
            
        Returns:
            bool: True se sucesso, False caso contrário
        """
        p_type = p_type.lower().strip()
        if p_type not in cls._structures:
            from utils.logger import log
            log(f"Tipo de projeto desconhecido: {p_type}", "ERROR")
            log(f"Tipos suportados: {', '.join(cls._structures.keys())}", "INFO")
            return False
        
        structure_class = cls._structures[p_type]
        structure = structure_class()
        
        try:
            return structure.setup(
                name=name,
                version=version,
                author=author,
                password=password,
                web_url=web_url
            )
        except Exception as e:
            from utils.logger import log
            log(f"Erro na criação do projeto: {str(e)}", "ERROR")
            return False

    @classmethod
    def list_types(cls) -> list:
        """
          Retorna lista de tipos de projeto suportados.
        """
        return list(cls._structures.keys())
