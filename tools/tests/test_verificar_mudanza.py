"""Pruebas del comprobador de identidad de lineas de tools/verificar_mudanza.py."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from verificar_mudanza import comparar

ORIGINAL = """## 1. Primera seccion

Una linea de contenido.
Otra linea.

## 2. Segunda seccion

Una linea de contenido.
"""


def escribir(tmp_path, nombre, texto):
    ruta = tmp_path / nombre
    ruta.write_text(texto, encoding="utf-8")
    return ruta


def test_mudanza_fiel_no_reporta_nada(tmp_path):
    original = escribir(tmp_path, "original.md", ORIGINAL)
    uno = escribir(tmp_path, "uno.md", "# Primera seccion\n\nUna linea de contenido.\nOtra linea.\n")
    dos = escribir(tmp_path, "dos.md", "# Segunda seccion\n\nUna linea de contenido.\n")
    informe = comparar(original, [uno, dos])
    assert informe.ok
    assert informe.lineas_perdidas == []
    assert informe.lineas_sobrantes == []


def test_detecta_una_linea_perdida(tmp_path):
    original = escribir(tmp_path, "original.md", ORIGINAL)
    uno = escribir(tmp_path, "uno.md", "# Primera seccion\n\nUna linea de contenido.\n")
    dos = escribir(tmp_path, "dos.md", "# Segunda seccion\n\nUna linea de contenido.\n")
    informe = comparar(original, [uno, dos])
    assert not informe.ok
    assert informe.lineas_perdidas == ["Otra linea."]


def test_detecta_una_linea_sobrante(tmp_path):
    original = escribir(tmp_path, "original.md", ORIGINAL)
    uno = escribir(tmp_path, "uno.md", "# Primera seccion\n\nUna linea de contenido.\nOtra linea.\n")
    dos = escribir(tmp_path, "dos.md", "# Segunda seccion\n\nUna linea de contenido.\nInventada.\n")
    informe = comparar(original, [uno, dos])
    assert not informe.ok
    assert informe.lineas_sobrantes == ["Inventada."]


def test_una_linea_repetida_tiene_que_aparecer_las_mismas_veces(tmp_path):
    """Una linea que sale dos veces en el original y una en el reparto es una perdida."""
    original = escribir(tmp_path, "original.md", ORIGINAL)
    uno = escribir(tmp_path, "uno.md", "# Primera seccion\n\nUna linea de contenido.\nOtra linea.\n")
    dos = escribir(tmp_path, "dos.md", "# Segunda seccion\n")
    informe = comparar(original, [uno, dos])
    assert not informe.ok
    assert informe.lineas_perdidas == ["Una linea de contenido."]


def test_detecta_un_encabezado_perdido(tmp_path):
    """Las lineas de contenido cuadran, pero falta un encabezado entero."""
    original = escribir(tmp_path, "original.md", ORIGINAL)
    uno = escribir(
        tmp_path,
        "uno.md",
        "# Primera seccion\n\nUna linea de contenido.\nOtra linea.\n\nUna linea de contenido.\n",
    )
    informe = comparar(original, [uno])
    assert not informe.ok
    assert informe.lineas_perdidas == []
    assert informe.encabezados_perdidos == ["Segunda seccion"]


def test_el_numero_de_seccion_no_cuenta_para_comparar_encabezados(tmp_path):
    """Quitar el 1. de "## 1. Primera seccion" es el cambio que la mudanza hace a proposito."""
    original = escribir(tmp_path, "original.md", "## 1. Primera seccion\n\nContenido.\n")
    uno = escribir(tmp_path, "uno.md", "# Primera seccion\n\nContenido.\n")
    informe = comparar(original, [uno])
    assert informe.ok


def test_el_nivel_del_encabezado_no_cuenta(tmp_path):
    """Al pasar a fichero propio, un ### se convierte en # y un #### en ##."""
    original = escribir(tmp_path, "original.md", "### 10.4. Algo\n\n#### Salida\n\nContenido.\n")
    uno = escribir(tmp_path, "uno.md", "# Algo\n\n## Salida\n\nContenido.\n")
    informe = comparar(original, [uno])
    assert informe.ok


def test_una_almohadilla_dentro_de_un_bloque_de_codigo_es_contenido(tmp_path):
    """Un # dentro de ``` no es un encabezado, es una linea de contenido."""
    original = escribir(tmp_path, "original.md", "## Uno\n\n```bash\n# un comentario de shell\n```\n")
    uno = escribir(tmp_path, "uno.md", "# Uno\n\n```bash\n# un comentario de shell\n```\n")
    informe = comparar(original, [uno])
    assert informe.ok


def test_el_titulo_del_documento_sobrevive_en_la_portada(tmp_path):
    """El caso real del reparto: SPEC.md tiene un H1 que docs/spec/index.md conserva.

    Un H1 del original que no aparece en ningun fichero nuevo SI se reporta como perdido, y
    eso es lo correcto: es la unica forma de enterarse de que un titulo desaparecio. Lo que
    esta prueba fija es que cuando el titulo si se conserva, no se reporta nada.
    """
    original = escribir(
        tmp_path,
        "original.md",
        "# El documento entero\n\nPresentacion.\n\n## 1. Primera seccion\n\nContenido.\n",
    )
    portada = escribir(tmp_path, "index.md", "# El documento entero\n\nPresentacion.\n")
    primera = escribir(tmp_path, "primera.md", "# Primera seccion\n\nContenido.\n")
    informe = comparar(original, [portada, primera])
    assert informe.ok


def test_un_titulo_que_no_conserva_nadie_se_reporta(tmp_path):
    """La cara contraria de la prueba anterior, que es la que da valor al comprobador."""
    original = escribir(tmp_path, "original.md", "# El documento entero\n\n## 1. Primera\n\nContenido.\n")
    primera = escribir(tmp_path, "primera.md", "# Primera\n\nContenido.\n")
    informe = comparar(original, [primera])
    assert not informe.ok
    assert informe.encabezados_perdidos == ["El documento entero"]
