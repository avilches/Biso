#!/usr/bin/env python3
"""Comprueba los enlaces relativos de los ficheros Markdown que estan fuera de docs/.

`mkdocs build --strict` valida los enlaces y las anclas de todo lo que hay dentro de docs/,
pero no ve CLAUDE.md, INTEGRATION.md ni los .md de bench/. Este script los cubre: por cada
enlace relativo a un fichero del repositorio, comprueba que el fichero existe y que, si el
enlace lleva ancla, ese fichero tiene un encabezado que la genera.

El ancla se calcula igual que el slugify de pymdownx que mkdocs.yml configura, para que los
dos coincidan: minusculas, fuera todo lo que no sea letra, numero, espacio o guion, y cada
racha de espacios convertida en un guion. Los acentos se conservan.

Se ejecuta:

    python tools/comprobar_enlaces.py CLAUDE.md INTEGRATION.md bench/sqlite-driver/*.md

Termina con codigo 0 si todos los enlaces resuelven y 1 si alguno no.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ENLACE_RE = re.compile(r"\[[^\]]*\]\(([^)\s]+)\)")
ENCABEZADO_RE = re.compile(r"^#{1,6}\s+(.*)$")
CERCA_RE = re.compile(r"^\s*```")
NO_ANCLA_RE = re.compile(r"[^\w\s-]", re.UNICODE)


def ancla_de(titulo: str) -> str:
    limpio = NO_ANCLA_RE.sub("", titulo.strip().lower())
    return re.sub(r"\s+", "-", limpio)


def anclas_de(ruta: Path) -> set[str]:
    anclas: set[str] = set()
    dentro_de_bloque = False
    for linea in ruta.read_text(encoding="utf-8").splitlines():
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            continue
        if dentro_de_bloque:
            continue
        coincidencia = ENCABEZADO_RE.match(linea)
        if coincidencia:
            anclas.add(ancla_de(coincidencia.group(1)))
    return anclas


def _enlaces_de(ruta: Path) -> list[str]:
    enlaces: list[str] = []
    dentro_de_bloque = False
    for linea in ruta.read_text(encoding="utf-8").splitlines():
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            continue
        if not dentro_de_bloque:
            enlaces.extend(ENLACE_RE.findall(linea))
    return enlaces


def enlaces_rotos(ficheros: list[Path], raiz: Path) -> list[str]:
    problemas: list[str] = []
    for fichero in ficheros:
        for destino in _enlaces_de(fichero):
            if destino.startswith(("http://", "https://", "mailto:", "#")):
                continue
            ruta, _, ancla = destino.partition("#")
            if not ruta:
                continue
            candidato = (fichero.parent / ruta).resolve()
            if not candidato.is_file():
                candidato = (raiz / ruta).resolve()
            if not candidato.is_file():
                problemas.append(f"{fichero}: el enlace {destino!r} apunta a un fichero que no existe")
                continue
            if ancla and candidato.suffix == ".md" and ancla not in anclas_de(candidato):
                problemas.append(f"{fichero}: el enlace {destino!r} apunta a un ancla que no existe")
    return problemas


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("uso: comprobar_enlaces.py <fichero>...", file=sys.stderr)
        return 2
    raiz = Path.cwd()
    problemas = enlaces_rotos([Path(a) for a in argv[1:]], raiz)
    for problema in problemas:
        print(problema)
    if problemas:
        return 1
    print("todos los enlaces relativos resuelven")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
