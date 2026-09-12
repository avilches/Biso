#!/usr/bin/env python3
"""Busca frases que cuentan elementos que el propio documento enumera.

La regla que vigila: en la prosa no se escribe un numero que cuente elementos que el
documento enumera y que pueden crecer, porque ese numero se queda desactualizado en
silencio en cuanto aparece uno mas. Se dice "las reglas de abajo", no "siete reglas".

La excepcion: si el numero es la regla y hay una prueba que lo comprueba, se queda. Los
5.120 bytes del mensaje de arranque, los 25 milisegundos del presupuesto, el recorte a 100
celdas, las ocho columnas fijas de `biso ls`. Esas frases se declaran una por linea en
tools/recuentos-normativos.txt, por su texto y no por su numero de linea, para que la lista
no se invalide al editar el fichero alrededor.

Se ejecuta:

    python tools/comprobar_recuentos.py docs/spec/*.md docs/spec/cmd/*.md docs/*.md

Termina con codigo 0 si no hay ningun recuento sin declarar y 1 si hay alguno.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

CARDINALES = (
    "dos|tres|cuatro|cinco|seis|siete|ocho|nueve|diez|once|doce|trece|catorce|quince|"
    "dieciseis|dieciséis|diecisiete|dieciocho|diecinueve|veinte|veintiun|veintiún|"
    "veintiuna|veintidos|veintidós|treinta|[0-9]+"
)
SUSTANTIVOS = (
    "reglas|filas|comandos|codigos|códigos|banderas|secciones|apartados|documentos|"
    "principios|garantias|garantías|mensajes|campos|verbos|criterios|columnas|entradas|"
    "tipos|familias|errores|avisos|estados|precisiones|comprobaciones|requisitos|decisiones"
)
RECUENTO_RE = re.compile(rf"\b({CARDINALES})\s+({SUSTANTIVOS})\b", re.IGNORECASE)
CERCA_RE = re.compile(r"^\s*```")
NORMATIVAS_PATH = Path(__file__).resolve().parent / "recuentos-normativos.txt"


def cargar_normativas(ruta: Path = NORMATIVAS_PATH) -> set[str]:
    if not ruta.is_file():
        return set()
    lineas = ruta.read_text(encoding="utf-8").splitlines()
    return {l.strip() for l in lineas if l.strip() and not l.startswith("#")}


def frases_que_cuentan(ruta: Path, normativas: set[str]) -> list[tuple[int, str]]:
    encontradas: list[tuple[int, str]] = []
    dentro_de_bloque = False
    for numero, linea in enumerate(ruta.read_text(encoding="utf-8").splitlines(), start=1):
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            continue
        if dentro_de_bloque:
            continue
        for coincidencia in RECUENTO_RE.finditer(linea):
            frase = coincidencia.group(0)
            if frase.lower() in {n.lower() for n in normativas}:
                continue
            encontradas.append((numero, frase))
    return encontradas


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("uso: comprobar_recuentos.py <fichero>...", file=sys.stderr)
        return 2
    normativas = cargar_normativas()
    total = 0
    for nombre in argv[1:]:
        ruta = Path(nombre)
        for numero, frase in frases_que_cuentan(ruta, normativas):
            print(f"{ruta}:{numero}: cuenta elementos del documento: {frase!r}")
            total += 1
    if total:
        print(f"\n{total} frases que cuentan sin declarar como normativas", file=sys.stderr)
        return 1
    print("ninguna frase cuenta elementos del documento sin declararlo")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
