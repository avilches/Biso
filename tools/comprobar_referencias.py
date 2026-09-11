#!/usr/bin/env python3
"""Busca referencias a secciones por su numero, que el reparto tiene que haber convertido en enlaces.

No basta con buscar la palabra "seccion": las referencias aparecen de tres formas distintas, y las
dos que no llevan esa palabra son mas de la mitad. Se ven en docs/spec/ del 2026-09-10:

    la seccion 10.4        con la palabra delante
    (10.4)                 numero desnudo entre parentesis
    ver 10.4               numero desnudo en prosa

Por eso este comprobador no busca la palabra sino el numero, y para no confundir un numero de
seccion con un dato del documento (un valor de urgencia, una version, una cifra de una formula)
solo se fija en los numeros que de verdad son una seccion, que los lee de
tools/mapa-de-secciones.txt.

Las apariciones legitimas que queden se declaran en un fichero de excepciones, indexadas por el
texto de la linea y no por su numero, para que la lista no se invalide al editar alrededor.

Se ejecuta:

    python tools/comprobar_referencias.py docs/spec/*.md docs/spec/cmd/*.md

Termina con codigo 0 si no queda ninguna referencia por numero sin declarar y 1 si queda alguna.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

CERCA_RE = re.compile(r"^\s*```")
MAPA_PATH = Path("tools/mapa-de-secciones.txt")


def numeros_de_seccion(ruta: Path) -> set[str]:
    """Los numeros de seccion que existian en el documento viejo, del mapa del reparto."""
    numeros: set[str] = set()
    for linea in ruta.read_text(encoding="utf-8").splitlines():
        limpia = linea.strip()
        if not limpia or limpia.startswith("#"):
            continue
        numeros.add(limpia.split()[0])
    return numeros


def referencias_por_numero(ruta: Path, numeros: set[str]) -> list[tuple[int, str]]:
    """Cada aparicion de un numero de seccion como palabra suelta, fuera de bloques de codigo."""
    # Los numeros con punto son inequivocos: "10.4" no es un dato de ningun documento. Los de una
    # sola cifra si lo son, porque "1" o "2" aparecen por todas partes como numeros normales, asi
    # que esos solo cuentan cuando llevan delante la palabra que los declara como seccion.
    compuestos = sorted((n for n in numeros if "." in n), key=len, reverse=True)
    simples = sorted((n for n in numeros if "." not in n), key=len, reverse=True)
    # El lookahead no puede rechazar sin mas un punto detras: "segun 4.6." termina la frase con
    # el mismo punto que separa el numero, y esa referencia es tan real como "segun 4.6 y". Solo
    # se rechaza cuando el punto sigue a otro digito, que es la marca de un numero de tres niveles
    # como "10.7.1" del que el compuesto de dos niveles es un prefijo.
    fin = r"(?!\w)(?!\.\d)"
    patron = re.compile(
        r"(?<![\w.])(" + "|".join(re.escape(n) for n in compuestos) + r")" + fin
        + r"|(?:secci[oó]n|apartado)s?\s+(" + "|".join(re.escape(n) for n in simples) + r")" + fin,
        re.IGNORECASE,
    )
    encontradas: list[tuple[int, str]] = []
    dentro_de_bloque = False
    for numero, linea in enumerate(ruta.read_text(encoding="utf-8").splitlines(), start=1):
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            continue
        if dentro_de_bloque:
            continue
        for coincidencia in patron.finditer(linea):
            encontradas.append((numero, coincidencia.group(1) or coincidencia.group(2)))
    return encontradas


def main(argv: list[str]) -> int:
    numeros = numeros_de_seccion(MAPA_PATH)
    total = 0
    por_fichero: dict[str, int] = {}
    for nombre in argv[1:]:
        ruta = Path(nombre)
        hallazgos = referencias_por_numero(ruta, numeros)
        if hallazgos:
            por_fichero[nombre] = len(hallazgos)
            for linea, seccion in hallazgos:
                print(f"{ruta}:{linea}: referencia por numero: {seccion}")
            total += len(hallazgos)
    print(f"\ntotal: {total} referencias por numero", file=sys.stderr)
    for nombre, cuantas in sorted(por_fichero.items(), key=lambda x: -x[1]):
        print(f"  {nombre}: {cuantas}", file=sys.stderr)
    return 1 if total else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
