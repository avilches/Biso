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


def _clasificar(texto: str) -> list[tuple[str, bool]]:
    """Clasifica cada linea de un texto como contenido o encabezado, en su orden.

    Devuelve una lista paralela a las lineas del fichero: para cada una, la linea tal cual y si
    es un encabezado. Hay que clasificar el fichero entero de una vez, empezando por su primera
    linea, porque saber si una linea esta dentro de un bloque de codigo depende de todas las
    cercas anteriores. Clasificar un trozo suelto da un resultado distinto y equivocado: un
    trozo que empieza dentro de un bloque cree estar fuera, y una linea de comentario de shell
    se toma por un encabezado.
    """
    clasificadas: list[tuple[str, bool]] = []
    dentro_de_bloque = False
    for linea in texto.splitlines():
        if CERCA_RE.match(linea):
            dentro_de_bloque = not dentro_de_bloque
            clasificadas.append((linea, False))
            continue
        clasificadas.append((linea, not dentro_de_bloque and bool(ENCABEZADO_RE.match(linea))))
    return clasificadas


def _repartir(texto: str) -> tuple[Counter[str], Counter[str]]:
    """Devuelve el multiconjunto de lineas de contenido y el de textos de encabezado."""
    lineas: Counter[str] = Counter()
    encabezados: Counter[str] = Counter()
    for linea, es_encabezado in _clasificar(texto):
        if es_encabezado:
            coincidencia = ENCABEZADO_RE.match(linea)
            encabezados[coincidencia.group(3).strip()] += 1
        elif linea.strip():
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


def _contenido_de_rangos(
    clasificadas: list[tuple[str, bool]], rangos: list[tuple[int, int]]
) -> Counter[str]:
    """El contenido que le toca a un fichero, sacado de la clasificacion del original entero.

    Recibe el original ya clasificado y no un trozo, justamente para no perder el contexto de
    los bloques de codigo en los bordes de cada rango.
    """
    contenido: Counter[str] = Counter()
    for primera, ultima in rangos:
        for linea, es_encabezado in clasificadas[primera - 1 : ultima]:
            if not es_encabezado and linea.strip():
                contenido[linea] += 1
    return contenido


def comparar_por_fichero(
    original: Path, entradas: list[tuple[Path, list[tuple[int, int]]]]
) -> list[str]:
    """Comprueba que cada fichero nuevo lleva exactamente el contenido de los rangos que le tocan.

    Esto es lo que atrapa un trasvase: dos ficheros que se intercambian su contenido pasan la
    comparacion global, porque ninguna linea se ha perdido, pero no pasan esta.

    Solo compara lineas de contenido. Los encabezados se comparan aparte y en conjunto, porque
    el reparto les cambia el nivel y a algunos el texto, a proposito.
    """
    clasificadas = _clasificar(original.read_text(encoding="utf-8"))
    total = len(clasificadas)
    problemas: list[str] = []
    dueno_de_la_linea: dict[int, Path] = {}
    for fichero, rangos in entradas:
        fuera_de_rango = False
        for primera, ultima in rangos:
            if primera < 1 or ultima > total:
                problemas.append(
                    f"{fichero}: el rango {primera}-{ultima} se sale del original, "
                    f"que acaba en la linea {total}"
                )
                fuera_de_rango = True
                continue
            for numero in range(primera, ultima + 1):
                anterior = dueno_de_la_linea.get(numero)
                if anterior is not None:
                    problemas.append(
                        f"{fichero}: la linea {numero} del original ya la reclamaba {anterior}"
                    )
                else:
                    dueno_de_la_linea[numero] = fichero
        if fuera_de_rango:
            continue
        if not fichero.is_file():
            problemas.append(f"{fichero}: el manifiesto lo nombra pero no existe")
            continue
        esperado = _contenido_de_rangos(clasificadas, rangos)
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


def cargar_excepciones(ruta: Path) -> tuple[set[str], set[str]]:
    """Lee las excepciones de encabezado declaradas: las que nacen y las que se pierden.

    Sin esta declaracion el comprobador no puede distinguir un encabezado que cambia a
    proposito de uno que se perdio por error, y un reparto correcto nunca podria terminar con
    codigo 0.
    """
    nuevos: set[str] = set()
    perdidos: set[str] = set()
    for numero, linea in enumerate(ruta.read_text(encoding="utf-8").splitlines(), start=1):
        limpia = linea.strip()
        if not limpia or limpia.startswith("#"):
            continue
        clase, _, texto = limpia.partition(":")
        clase, texto = clase.strip(), texto.strip()
        if clase == "nuevo" and texto:
            nuevos.add(texto)
        elif clase == "perdido" and texto:
            perdidos.add(texto)
        else:
            raise ValueError(
                f'{ruta}:{numero}: cada linea es "nuevo: <texto>" o "perdido: <texto>"'
            )
    return nuevos, perdidos


def adjudicar_encabezados(informe: Informe, nuevos: set[str], perdidos: set[str]) -> list[str]:
    """Compara los encabezados que cambiaron contra los que se declaro que iban a cambiar.

    Falla en las dos direcciones a proposito. Un encabezado que cambio sin estar declarado es un
    error de la mudanza. Y una excepcion declarada que no llego a ocurrir tambien lo es, porque
    significa que la declaracion ya no describe el reparto que hay.
    """
    problemas: list[str] = []
    observados_nuevos = set(informe.encabezados_sobrantes)
    observados_perdidos = set(informe.encabezados_perdidos)
    for texto in sorted(observados_nuevos - nuevos):
        problemas.append(f"encabezado nuevo sin declarar: {texto!r}")
    for texto in sorted(observados_perdidos - perdidos):
        problemas.append(f"encabezado perdido sin declarar: {texto!r}")
    for texto in sorted(nuevos - observados_nuevos):
        problemas.append(f"se declaro como nuevo un encabezado que no aparece: {texto!r}")
    for texto in sorted(perdidos - observados_perdidos):
        problemas.append(f"se declaro como perdido un encabezado que sigue estando: {texto!r}")
    return problemas


def _uso() -> int:
    print(
        "uso: verificar_mudanza.py [--excepciones <fichero>] <original> <fichero-nuevo>...\n"
        "     verificar_mudanza.py [--excepciones <fichero>] --manifiesto <manifiesto> <original>",
        file=sys.stderr,
    )
    return 2


def main(argv: list[str]) -> int:
    argumentos = argv[1:]
    excepciones: Path | None = None
    if argumentos[:1] == ["--excepciones"]:
        if len(argumentos) < 2:
            return _uso()
        excepciones = Path(argumentos[1])
        argumentos = argumentos[2:]
    try:
        if argumentos[:1] == ["--manifiesto"]:
            if len(argumentos) < 3:
                return _uso()
            entradas = cargar_manifiesto(Path(argumentos[1]))
            original = Path(argumentos[2])
            nuevos = [fichero for fichero, _ in entradas if fichero.is_file()]
            problemas = comparar_por_fichero(original, entradas)
            problemas += cercas_desbalanceadas(nuevos)
        elif len(argumentos) >= 2:
            original = Path(argumentos[0])
            nuevos = [Path(a) for a in argumentos[1:]]
            problemas = []
        else:
            return _uso()
        informe = comparar(original, nuevos)
        if excepciones is not None:
            declarados_nuevos, declarados_perdidos = cargar_excepciones(excepciones)
            problemas += adjudicar_encabezados(informe, declarados_nuevos, declarados_perdidos)
    except (OSError, ValueError) as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    for problema in problemas:
        print(problema)
    for elemento in informe.lineas_perdidas:
        print(f"linea que se perdio: {elemento!r}")
    for elemento in informe.lineas_sobrantes:
        print(f"linea que aparecio de la nada: {elemento!r}")
    if excepciones is None:
        for elemento in informe.encabezados_perdidos:
            print(f"encabezado que se perdio: {elemento!r}")
        for elemento in informe.encabezados_sobrantes:
            print(f"encabezado que aparecio de la nada: {elemento!r}")
    hay_lineas = bool(informe.lineas_perdidas or informe.lineas_sobrantes)
    hay_encabezados = bool(informe.encabezados_perdidos or informe.encabezados_sobrantes)
    if excepciones is not None:
        correcto = not problemas and not hay_lineas
    else:
        correcto = not problemas and not hay_lineas and not hay_encabezados
    if correcto:
        print("la mudanza es fiel: cada fichero lleva su contenido y no falta ni sobra nada")
        return 0
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
