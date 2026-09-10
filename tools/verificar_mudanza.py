#!/usr/bin/env python3
"""Comprueba que un reparto de un documento en varios ficheros no perdio nada.

Compara el fichero original con el conjunto de ficheros que salieron de el, y responde si
el contenido sobrevivio entero. Compara dos cosas por separado:

1. Las lineas de contenido, que son todas las que no son un encabezado, como multiconjunto:
   una linea que sale tres veces en el original tiene que salir tres veces en el reparto.
2. Los textos de los encabezados, sin sus almohadillas ni su numero de seccion, tambien como
   multiconjunto. Sin esto, perder un encabezado entero pasaria inadvertido, porque la
   primera comparacion los ignora.

Lo que se admite que cambie, y por eso se normaliza antes de comparar: el numero de seccion
del encabezado, que la mudanza quita a proposito, y su nivel, porque una subseccion que pasa
a ser un fichero propio sube de ### a #.

Se ejecuta:

    python tools/verificar_mudanza.py docs/SPEC.md docs/spec/*.md docs/spec/cmd/*.md

Termina con codigo 0 si todo cuadra y 1 si no, listando cada diferencia.
"""

from __future__ import annotations

import re
import sys
from collections import Counter
from dataclasses import dataclass, field
from pathlib import Path

# "## 10.4. `biso ls`" -> nivel "##", numero "10.4.", texto "`biso ls`"
ENCABEZADO_RE = re.compile(r"^(#{1,6})\s+(?:(\d+(?:\.\d+)*)\.?\s+)?(.*)$")
CERCA_RE = re.compile(r"^\s*```")


@dataclass
class Informe:
    lineas_perdidas: list[str] = field(default_factory=list)
    lineas_sobrantes: list[str] = field(default_factory=list)
    encabezados_perdidos: list[str] = field(default_factory=list)
    encabezados_sobrantes: list[str] = field(default_factory=list)

    @property
    def ok(self) -> bool:
        return not (
            self.lineas_perdidas
            or self.lineas_sobrantes
            or self.encabezados_perdidos
            or self.encabezados_sobrantes
        )


def _repartir(texto: str) -> tuple[Counter[str], Counter[str]]:
    """Devuelve el multiconjunto de lineas de contenido y el de textos de encabezado."""
    lineas: Counter[str] = Counter()
    encabezados: Counter[str] = Counter()
    dentro_de_bloque = False
    for linea in texto.splitlines():
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            lineas[linea] += 1
            continue
        if not dentro_de_bloque:
            coincidencia = ENCABEZADO_RE.match(linea)
            if coincidencia:
                encabezados[coincidencia.group(3).strip()] += 1
                continue
        if linea.strip():
            lineas[linea] += 1
    return lineas, encabezados


def _diferencia(sobran: Counter[str], faltan: Counter[str]) -> list[str]:
    return sorted((sobran - faltan).elements())


def comparar(original: Path, nuevos: list[Path]) -> Informe:
    lineas_viejas, encabezados_viejos = _repartir(original.read_text(encoding="utf-8"))
    lineas_nuevas: Counter[str] = Counter()
    encabezados_nuevos: Counter[str] = Counter()
    for ruta in nuevos:
        lineas, encabezados = _repartir(ruta.read_text(encoding="utf-8"))
        lineas_nuevas.update(lineas)
        encabezados_nuevos.update(encabezados)
    return Informe(
        lineas_perdidas=_diferencia(lineas_viejas, lineas_nuevas),
        lineas_sobrantes=_diferencia(lineas_nuevas, lineas_viejas),
        encabezados_perdidos=_diferencia(encabezados_viejos, encabezados_nuevos),
        encabezados_sobrantes=_diferencia(encabezados_nuevos, encabezados_viejos),
    )


def main(argv: list[str]) -> int:
    if len(argv) < 3:
        print("uso: verificar_mudanza.py <original> <fichero-nuevo>...", file=sys.stderr)
        return 2
    informe = comparar(Path(argv[1]), [Path(a) for a in argv[2:]])
    for etiqueta, elementos in (
        ("linea que se perdio", informe.lineas_perdidas),
        ("linea que aparecio de la nada", informe.lineas_sobrantes),
        ("encabezado que se perdio", informe.encabezados_perdidos),
        ("encabezado que aparecio de la nada", informe.encabezados_sobrantes),
    ):
        for elemento in elementos:
            print(f"{etiqueta}: {elemento!r}")
    if informe.ok:
        print("la mudanza es fiel: ninguna linea ni encabezado se perdio o aparecio")
        return 0
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
