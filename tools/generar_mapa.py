#!/usr/bin/env python3
"""Genera el mapa de numero de seccion a fichero y ancla del reparto.

Lee los encabezados numerados de docs/SPEC.md y el reparto declarado en el manifiesto, y para
cada numero de seccion dice a que fichero de docs/spec/ y a que ancla hay que apuntar. Es lo
que convierte la reescritura de las 646 referencias en un trabajo mecanico en vez de en 646
decisiones sueltas.

El ancla se calcula igual que el slugify de pymdownx que mkdocs.yml configura: minusculas,
fuera todo lo que no sea letra, numero, espacio o guion, y cada racha de espacios convertida
en un guion. Los acentos se conservan.

Cuando el encabezado es el que se convierte en el H1 de su fichero, el ancla queda vacia y la
referencia es al fichero entero.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

# Los siete ficheros cuyo H1 nace de la nada, de la tabla de excepciones de esta tarea. En
# ellos ningun encabezado viejo se gasta como H1: todos se conservan como subsecciones, asi
# que todos tienen ancla. En los demas, el H1 sale de la primera linea de su primer rango, que
# es un encabezado, y esa seccion se cita por el fichero entero y sin ancla.
H1_NUEVO = {
    "vocabulario.md",
    "invocacion.md",
    "salida-y-terminal.md",
    "valores-de-entrada.md",
    "garantias.md",
    "presupuestos.md",
    "cmd/help.md",
}

ENCABEZADO_RE = re.compile(r"^(#{2,6})\s+(\d+(?:\.\d+)*)\.?\s+(.*)$")
CERCA_RE = re.compile(r"^\s*```")
NO_ANCLA_RE = re.compile(r"[^\w\s-]", re.UNICODE)


def ancla_de(titulo: str) -> str:
    limpio = NO_ANCLA_RE.sub("", titulo.strip().lower())
    return re.sub(r"\s+", "-", limpio)


def cargar_rangos(manifiesto: Path) -> list[tuple[str, list[tuple[int, int]]]]:
    """Lee los rangos del mismo manifiesto que usa tools/verificar_mudanza.py.

    Se leen de ahi y no se repiten en este fichero para que no puedan desincronizarse: el
    manifiesto es la unica declaracion de que rangos le tocan a cada fichero.
    """
    entradas: list[tuple[str, list[tuple[int, int]]]] = []
    for linea in manifiesto.read_text(encoding="utf-8").splitlines():
        limpia = linea.strip()
        if not limpia or limpia.startswith("#"):
            continue
        ruta, _, rangos = limpia.partition(":")
        lista = []
        for trozo in rangos.split(","):
            primera, ultima = trozo.strip().split("-")
            lista.append((int(primera), int(ultima)))
        entradas.append((ruta.strip().removeprefix("docs/spec/"), lista))
    return entradas


def fichero_de_linea(rangos_por_fichero, numero: int) -> str | None:
    for fichero, rangos in rangos_por_fichero:
        for primera, ultima in rangos:
            if primera <= numero <= ultima:
                return fichero
    return None


def main() -> int:
    spec = Path("docs/SPEC.md")
    manifiesto = Path(sys.argv[1] if len(sys.argv) > 1 else "tools/manifiesto-del-reparto.txt")
    rangos_por_fichero = cargar_rangos(manifiesto)
    primera_linea_de = {f: r[0][0] for f, r in rangos_por_fichero}
    filas: list[tuple[str, str, str, str]] = []
    dentro_de_bloque = False
    for numero, linea in enumerate(spec.read_text(encoding="utf-8").splitlines(), start=1):
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            continue
        if dentro_de_bloque:
            continue
        coincidencia = ENCABEZADO_RE.match(linea)
        if not coincidencia:
            continue
        seccion, titulo = coincidencia.group(2), coincidencia.group(3).strip()
        fichero = fichero_de_linea(rangos_por_fichero, numero)
        if fichero is None:
            print(f"AVISO: la seccion {seccion} de la linea {numero} no cae en ningun rango", file=sys.stderr)
            continue
        es_h1 = fichero not in H1_NUEVO and primera_linea_de[fichero] == numero
        ancla = "" if es_h1 else ancla_de(titulo)
        filas.append((seccion, fichero, ancla, titulo))
    ancho = max(len(f[0]) for f in filas)
    for seccion, fichero, ancla, titulo in filas:
        destino = fichero if not ancla else f"{fichero}#{ancla}"
        print(f"{seccion:<{ancho}}  {destino}  \"{titulo}\"")
    print(f"\n{len(filas)} secciones numeradas mapeadas", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
