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


RANGO_RE = re.compile(r"^(\d+)-(\d+)$")


def cargar_manifiesto(ruta: Path) -> list[tuple[Path, list[tuple[int, int]]]]:
    """Lee el manifiesto del reparto: una linea por fichero nuevo con los rangos que le tocan.

    El formato es "ruta: primera-ultima[, primera-ultima]...", con las lineas en blanco y las
    que empiezan por almohadilla ignoradas. Los numeros son de lineas del fichero original,
    con los dos extremos incluidos.
    """
    entradas: list[tuple[Path, list[tuple[int, int]]]] = []
    for numero, linea in enumerate(ruta.read_text(encoding="utf-8").splitlines(), start=1):
        limpia = linea.strip()
        if not limpia or limpia.startswith("#"):
            continue
        if ":" not in limpia:
            raise ValueError(f"{ruta}:{numero}: falta el dos puntos que separa el fichero de sus rangos")
        fichero, _, rangos = limpia.partition(":")
        lista: list[tuple[int, int]] = []
        for trozo in rangos.split(","):
            coincidencia = RANGO_RE.match(trozo.strip())
            if not coincidencia:
                raise ValueError(f"{ruta}:{numero}: {trozo.strip()!r} no tiene la forma <primera>-<ultima>")
            primera, ultima = int(coincidencia.group(1)), int(coincidencia.group(2))
            if primera > ultima:
                raise ValueError(f"{ruta}:{numero}: el rango {primera}-{ultima} empieza despues de acabar")
            lista.append((primera, ultima))
        entradas.append((Path(fichero.strip()), lista))
    return entradas


def _contenido_de_rangos(lineas: list[str], rangos: list[tuple[int, int]]) -> Counter[str]:
    trozo: list[str] = []
    for primera, ultima in rangos:
        trozo.extend(lineas[primera - 1 : ultima])
    contenido, _ = _repartir("\n".join(trozo))
    return contenido


def comparar_por_fichero(original: Path, entradas: list[tuple[Path, list[tuple[int, int]]]]) -> list[str]:
    """Comprueba que cada fichero nuevo lleva exactamente el contenido de los rangos que le tocan.

    Esto es lo que atrapa un trasvase: dos ficheros que se intercambian su contenido pasan la
    comparacion global, porque ninguna linea se ha perdido, pero no pasan esta.

    Solo compara lineas de contenido. Los encabezados se comparan aparte y en conjunto, porque
    el reparto les cambia el nivel y a algunos el texto, a proposito.
    """
    lineas_original = original.read_text(encoding="utf-8").splitlines()
    problemas: list[str] = []
    for fichero, rangos in entradas:
        if not fichero.is_file():
            problemas.append(f"{fichero}: el manifiesto lo nombra pero no existe")
            continue
        esperado = _contenido_de_rangos(lineas_original, rangos)
        real, _ = _repartir(fichero.read_text(encoding="utf-8"))
        for linea in sorted((esperado - real).elements()):
            problemas.append(f"{fichero}: le falta una linea de sus rangos: {linea!r}")
        for linea in sorted((real - esperado).elements()):
            problemas.append(f"{fichero}: tiene una linea que no sale de sus rangos: {linea!r}")
    return problemas


def cercas_desbalanceadas(ficheros: list[Path]) -> list[str]:
    """Un fichero con un numero impar de cercas tiene un bloque de codigo sin cerrar.

    Pasa cuando un corte cae en mitad de un bloque, y la comparacion de lineas no lo ve porque
    ninguna linea se pierde: solo se rompe el markdown.
    """
    problemas: list[str] = []
    for fichero in ficheros:
        cercas = sum(1 for linea in fichero.read_text(encoding="utf-8").splitlines() if CERCA_RE.match(linea))
        if cercas % 2:
            problemas.append(f"{fichero}: numero impar de cercas de bloque de codigo, hay un bloque sin cerrar")
    return problemas


def main(argv: list[str]) -> int:
    try:
        if len(argv) >= 4 and argv[1] == "--manifiesto":
            manifiesto, original = Path(argv[2]), Path(argv[3])
            entradas = cargar_manifiesto(manifiesto)
            nuevos = [fichero for fichero, _ in entradas]
            problemas = comparar_por_fichero(original, entradas)
            problemas += cercas_desbalanceadas([f for f in nuevos if f.is_file()])
        elif len(argv) >= 3 and argv[1] != "--manifiesto":
            original = Path(argv[1])
            nuevos = [Path(a) for a in argv[2:]]
            problemas = []
        else:
            print(
                "uso: verificar_mudanza.py <original> <fichero-nuevo>...\n"
                "     verificar_mudanza.py --manifiesto <manifiesto> <original>",
                file=sys.stderr,
            )
            return 2
        informe = comparar(original, nuevos)
    except (OSError, ValueError) as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    for problema in problemas:
        print(problema)
    for etiqueta, elementos in (
        ("linea que se perdio", informe.lineas_perdidas),
        ("linea que aparecio de la nada", informe.lineas_sobrantes),
        ("encabezado que se perdio", informe.encabezados_perdidos),
        ("encabezado que aparecio de la nada", informe.encabezados_sobrantes),
    ):
        for elemento in elementos:
            print(f"{etiqueta}: {elemento!r}")
    if informe.ok and not problemas:
        print("la mudanza es fiel: cada fichero lleva su contenido y no falta ni sobra nada")
        return 0
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
