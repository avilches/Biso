"""Pruebas del comprobador de identidad de lineas de tools/verificar_mudanza.py."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from verificar_mudanza import (
    adjudicar_encabezados,
    cargar_excepciones,
    cargar_manifiesto,
    cercas_desbalanceadas,
    comparar,
    comparar_por_fichero,
    main,
)

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


CRUZABLE = """### 1. Primero

#### Salida

Arranca la tarea 12.

### 2. Segundo

#### Salida

Para la tarea 12.
"""


def test_el_contenido_cruzado_entre_dos_ficheros_se_detecta(tmp_path):
    """Dos ficheros que se intercambian su contenido pasan la comparacion global, no esta.

    Es el caso que de verdad importa: el encabezado "Salida" sale 17 veces en la
    especificacion real, asi que un corte mal hecho puede dejar la salida de un comando en
    el fichero de otro sin que ninguna linea se pierda.
    """
    original = escribir(tmp_path, "original.md", CRUZABLE)
    primero = escribir(tmp_path, "primero.md", "# Primero\n\n## Salida\n\nPara la tarea 12.\n")
    segundo = escribir(tmp_path, "segundo.md", "# Segundo\n\n## Salida\n\nArranca la tarea 12.\n")
    entradas = [(primero, [(1, 5)]), (segundo, [(7, 11)])]
    assert comparar(original, [primero, segundo]).ok
    problemas = comparar_por_fichero(original, entradas)
    assert len(problemas) == 4


def test_un_fichero_fiel_a_sus_rangos_no_da_problemas(tmp_path):
    original = escribir(tmp_path, "original.md", CRUZABLE)
    primero = escribir(tmp_path, "primero.md", "# Primero\n\n## Salida\n\nArranca la tarea 12.\n")
    segundo = escribir(tmp_path, "segundo.md", "# Segundo\n\n## Salida\n\nPara la tarea 12.\n")
    assert comparar_por_fichero(original, [(primero, [(1, 5)]), (segundo, [(7, 11)])]) == []


def test_un_bloque_de_codigo_sin_cerrar_se_reporta(tmp_path):
    partido = escribir(tmp_path, "partido.md", "# Uno\n\n```bash\nuna orden\n")
    entero = escribir(tmp_path, "entero.md", "# Dos\n\n```bash\nuna orden\n```\n")
    problemas = cercas_desbalanceadas([partido, entero])
    assert len(problemas) == 1
    assert "partido.md" in problemas[0]


def test_un_fichero_que_no_se_puede_leer_da_un_error_limpio(tmp_path):
    """Sin esto, main termina con una traza de Python en vez de con un mensaje."""
    existe = escribir(tmp_path, "existe.md", "# Uno\n\nContenido.\n")
    assert main(["prog", str(tmp_path / "no-existe.md"), str(existe)]) == 2


BLOQUE = """## 1. Primero

```bash
# un comentario de shell
una orden
```

## 2. Segundo

Contenido llano.
"""


def test_un_rango_que_empieza_dentro_de_un_bloque_no_pierde_la_linea(tmp_path):
    """El comentario de shell es contenido, no un encabezado, aunque empiece por almohadilla.

    Clasificar el trozo por separado lo tomaria por un encabezado y lo perderia del contenido
    esperado, porque el trozo no sabe que viene detras de una cerca de apertura.
    """
    original = escribir(tmp_path, "original.md", BLOQUE)
    trozo = escribir(tmp_path, "trozo.md", "# Primero\n\n```bash\n# un comentario de shell\nuna orden\n```\n")
    assert comparar_por_fichero(original, [(trozo, [(1, 6)])]) == []


def test_rangos_no_contiguos_no_confunden_el_estado_del_bloque(tmp_path):
    original = escribir(tmp_path, "original.md", BLOQUE)
    juntos = escribir(
        tmp_path,
        "juntos.md",
        "# Juntado\n\n```bash\n# un comentario de shell\nuna orden\n```\n\nContenido llano.\n",
    )
    assert comparar_por_fichero(original, [(juntos, [(3, 6), (10, 10)])]) == []


def test_un_rango_que_se_sale_del_original_se_reporta(tmp_path):
    original = escribir(tmp_path, "original.md", "## 1. Primero\n\nContenido.\n")
    fichero = escribir(tmp_path, "uno.md", "# Primero\n\nContenido.\n")
    problemas = comparar_por_fichero(original, [(fichero, [(1, 99)])])
    assert len(problemas) == 1
    assert "se sale del original" in problemas[0]


def test_dos_ficheros_que_reclaman_la_misma_linea_se_reportan(tmp_path):
    original = escribir(tmp_path, "original.md", "## 1. Primero\n\nContenido.\n")
    uno = escribir(tmp_path, "uno.md", "# Primero\n\nContenido.\n")
    dos = escribir(tmp_path, "dos.md", "# Primero\n\nContenido.\n")
    problemas = comparar_por_fichero(original, [(uno, [(1, 3)]), (dos, [(1, 3)])])
    assert any("ya la reclamaba" in p for p in problemas)


def test_el_mensaje_del_fichero_que_falta_llega_a_imprimirse(tmp_path, capsys):
    """Antes moria: main pasaba el fichero ausente a comparar, que petaba antes de imprimir."""
    original = escribir(tmp_path, "original.md", "## 1. Primero\n\nContenido.\n")
    existe = escribir(tmp_path, "existe.md", "# Primero\n\nContenido.\n")
    manifiesto = escribir(
        tmp_path,
        "manifiesto.txt",
        f"{existe}: 1-3\n{tmp_path / 'no-existe.md'}: 1-3\n",
    )
    codigo = main(["prog", "--manifiesto", str(manifiesto), str(original)])
    salida = capsys.readouterr().out
    assert codigo == 1
    assert "el manifiesto lo nombra pero no existe" in salida


def test_las_excepciones_declaradas_dan_codigo_cero(tmp_path, capsys):
    """Un encabezado que cambia a proposito, declarado, no impide terminar en verde."""
    original = escribir(tmp_path, "original.md", "## 1. Titulo viejo\n\nContenido.\n")
    nuevo = escribir(tmp_path, "nuevo.md", "# Titulo nuevo\n\nContenido.\n")
    declaracion = escribir(
        tmp_path, "excepciones.txt", "nuevo: Titulo nuevo\nperdido: Titulo viejo\n"
    )
    codigo = main(["prog", "--excepciones", str(declaracion), str(original), str(nuevo)])
    assert codigo == 0
    assert "la mudanza es fiel" in capsys.readouterr().out


def test_un_encabezado_que_cambia_sin_declarar_falla(tmp_path, capsys):
    original = escribir(tmp_path, "original.md", "## 1. Titulo viejo\n\nContenido.\n")
    nuevo = escribir(tmp_path, "nuevo.md", "# Titulo nuevo\n\nContenido.\n")
    declaracion = escribir(tmp_path, "excepciones.txt", "# ninguna declarada\n")
    codigo = main(["prog", "--excepciones", str(declaracion), str(original), str(nuevo)])
    salida = capsys.readouterr().out
    assert codigo == 1
    assert "encabezado nuevo sin declarar: 'Titulo nuevo'" in salida
    assert "encabezado perdido sin declarar: 'Titulo viejo'" in salida


def test_una_excepcion_declarada_que_no_ocurre_falla(tmp_path):
    """Si la declaracion promete un cambio que no esta, se ha quedado desfasada."""
    original = escribir(tmp_path, "original.md", "## 1. Igual\n\nContenido.\n")
    nuevo = escribir(tmp_path, "nuevo.md", "# Igual\n\nContenido.\n")
    informe = comparar(original, [nuevo])
    problemas = adjudicar_encabezados(informe, {"Inventado"}, set())
    assert problemas == ["se declaro como nuevo un encabezado que no aparece: 'Inventado'"]
