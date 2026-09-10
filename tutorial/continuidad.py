"""Comprueba que los contadores del tablero encadenan entre escenarios consecutivos.

Cada fixture declara en `tablero_entra` y `tablero_sale` con que estado del tablero empieza y con
cual acaba. Este script extrae de esa prosa los contadores con la forma "To Do / In Progress / Done"
y comprueba que lo que un escenario entrega es lo que el siguiente recibe. Un desfase ahi significa
que dos escenarios cuentan historias distintas, y es la rotura que ningun escritor individual puede
ver desde su propio fichero.

    uv run --with-requirements docs-requirements.txt --no-project python tutorial/continuidad.py

Sale 0 si la cadena cuadra y 1 si no.
"""

import pathlib
import re
import sys

import yaml

DIR = pathlib.Path(__file__).resolve().parent / "escenarios"
# Los fixtures citan los contadores de dos formas, las dos legitimas: la abreviada "54/5/190" y la
# larga "To Do 54 | In Progress 5 | Done 190", que es la que imprime `biso prime`.
ABREVIADA = re.compile(r"\b(\d+)\s*/\s*(\d+)\s*/\s*(\d+)\b")
LARGA = re.compile(
    r"To Do\s+(\d+)\s*\|\s*In Progress\s+(\d+)\s*\|\s*Done\s+(\d+)", re.IGNORECASE
)


def contadores(texto):
    """Devuelve el ultimo trio de contadores citado en el texto, que es el que vale."""
    texto = texto or ""
    encontrados = [(m.end(), m.groups()) for m in ABREVIADA.finditer(texto)]
    encontrados += [(m.end(), m.groups()) for m in LARGA.finditer(texto)]
    if not encontrados:
        return None
    encontrados.sort()
    return tuple(int(n) for n in encontrados[-1][1])


def main():
    problemas = []
    anterior = None
    print("%-38s %-14s %-14s" % ("escenario", "entra", "sale"))
    for fichero in sorted(DIR.glob("*.yaml")):
        datos = yaml.safe_load(fichero.read_text(encoding="utf-8"))
        entra = contadores(datos.get("tablero_entra"))
        sale = contadores(datos.get("tablero_sale"))

        def formatea(c):
            return "%d/%d/%d" % c if c else "(sin declarar)"

        print("%-38s %-14s %-14s" % (fichero.stem, formatea(entra), formatea(sale)))

        if entra is None or sale is None:
            problemas.append("%s no declara los contadores en uno de los dos campos" % fichero.name)
        elif anterior and anterior[1] != entra:
            problemas.append(
                "%s entra en %s pero %s salia en %s"
                % (fichero.name, formatea(entra), anterior[0], formatea(anterior[1]))
            )
        if sale:
            anterior = (fichero.name, sale)

    print()
    if problemas:
        for p in problemas:
            print("ERROR: %s" % p)
        return 1
    print("La cadena de contadores cuadra en los %d escenarios." % len(list(DIR.glob("*.yaml"))))
    return 0


if __name__ == "__main__":
    sys.exit(main())
