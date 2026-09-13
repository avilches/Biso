"""Hook de MkDocs que saca docs/superpowers/ del indice de busqueda.

El plugin de busqueda de mkdocs-material (9.7.7) no tiene una opcion de configuracion
para excluir un directorio entero: solo lee `page.meta["search"]["exclude"]`, que se
declara pagina a pagina en su frontmatter. Poner ese frontmatter a mano en cada fichero
de superpowers/ se olvidaria en el siguiente plan o diseno que se anada ahi, asi que este
hook lo fija en cuanto MkDocs lee cada pagina, para las que ya existen y las que vengan.

superpowers/ se sigue publicando (lo exige el enlace de docs/estado-del-arte/index.md a un
diseno concreto) y se sigue sacando del menu con `not_in_nav` en mkdocs.yml: esto solo
evita que una busqueda devuelva un plan de sesion en vez de un documento principal.
"""

from __future__ import annotations


def on_page_markdown(markdown, page, config, files):
    if page.file.src_uri.startswith("superpowers/"):
        page.meta.setdefault("search", {})["exclude"] = True
    return markdown
