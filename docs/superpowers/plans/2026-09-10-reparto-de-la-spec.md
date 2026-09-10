# Plan de implementación del reparto de `SPEC.md`

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Partir `docs/SPEC.md` en 34 documentos que se citan por el título de sus secciones y no por su
número, sin perder una sola línea de contenido, y dejar tres comprobadores que impidan que el problema
vuelva.

**Architecture:** El trabajo se hace en dos mitades con pruebas distintas. La mudanza mueve texto sin
cambiar ni una letra del contenido, y se prueba con una comparación exacta de líneas contra el fichero
viejo. La reescritura de referencias cambia líneas por diseño, y se prueba con `mkdocs build --strict`,
que valida cada enlace y cada ancla y aborta con código 1 si alguno no resuelve. Nunca se hacen las dos
cosas en el mismo commit.

**Tech Stack:** MkDocs 1.6.1 con el tema Material, pymdown-extensions 11.0.2, y scripts de Python 3 de
la biblioteca estándar, ejecutados con `uv` sin crear entorno virtual en el repositorio. Las pruebas de
los scripts, con pytest.

**Spec:** [`docs/superpowers/specs/2026-09-10-reparto-de-la-spec-design.md`](../specs/2026-09-10-reparto-de-la-spec-design.md)

## Global Constraints

- **Nunca em-dash** (`—`) ni en documentación, ni en código, ni en comentarios, ni en mensajes de commit.
- **La documentación y los comentarios van en español.** Los identificadores del código en inglés.
- **Los mensajes de commit van sin acentos**, siguiendo la convención de los commits del repositorio, y
  **no llevan coautoría** ni mención de haber sido generados por un agente.
- **El trabajo va en el worktree `.claude/worktrees/reparto-de-la-spec`**, rama
  `worktree-reparto-de-la-spec`. Nunca se edita el checkout principal.
- **Ningún nombre de fichero de `docs/spec/` empieza por un número.**
- **Ningún fichero de `docs/spec/` pasa de 700 líneas.**
- Las órdenes de `uv` van siempre con `--no-project`, para no crear entorno en el repositorio.
- El comando de construcción de la documentación es
  `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`.

## Estado de partida, medido el 2026-09-10

- `docs/SPEC.md`: 5.762 líneas. La numeración de líneas que usa la tarea 3 es la de ese fichero en el
  commit `3bb97ed`, del que sale esta rama. **Si `SPEC.md` cambia antes de ejecutar la tarea 3, hay que
  volver a calcular el mapa de líneas.**
- `mkdocs build --strict` con `validation.anchors` termina con código 0 y cero avisos. Se parte de verde.
- 646 referencias por número, ninguna en forma de enlace.

---

## Task 1: El comprobador de identidad de líneas

Es la red de seguridad de toda la mudanza, así que se escribe antes de mover nada. Compara un fichero
original con un conjunto de ficheros nuevos y responde si el contenido sobrevivió entero.

Compara dos cosas por separado, y las dos importan:

1. **Las líneas de contenido**, que son todas las que no empiezan por `#`. Se comparan como multiconjunto,
   de modo que una línea repetida tres veces en el original tiene que aparecer tres veces en el reparto.
2. **Los textos de los encabezados**, normalizados quitando las almohadillas, el número de sección y los
   espacios de los extremos. También como multiconjunto. Sin esto, perder un encabezado entero pasaría
   inadvertido, porque la primera comparación los ignora.

**Files:**
- Create: `tools/verificar_mudanza.py`
- Create: `tools/requirements.txt`
- Test: `tools/tests/test_verificar_mudanza.py`

**Interfaces:**
- Produces: `comparar(original: Path, nuevos: list[Path]) -> Informe`, donde `Informe` es un dataclass con
  los campos `lineas_perdidas: list[str]`, `lineas_sobrantes: list[str]`,
  `encabezados_perdidos: list[str]`, `encabezados_sobrantes: list[str]`, y una propiedad `ok: bool` que es
  cierta cuando las cuatro listas están vacías.
- Produces: interfaz de línea de comandos
  `python tools/verificar_mudanza.py <original> <fichero-nuevo>...`, que imprime el informe y termina con
  código 0 si todo cuadra y 1 si no.

- [ ] **Step 1: Averiguar la versión de pytest y fijarla**

No inventes el número de versión. Pregúntaselo a `uv` y escribe el que responda:

```bash
uv run --with pytest --no-project python -c "import pytest; print(pytest.__version__)"
```

Crea `tools/requirements.txt` con esa versión exacta y este comentario encima:

```
# Dependencias de las pruebas de los scripts de tools/. Se ejecutan con
# uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q
pytest==<la version que respondio el comando de arriba>
```

- [ ] **Step 2: Escribir las pruebas que fallan**

Crea `tools/tests/test_verificar_mudanza.py`:

```python
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
```

- [ ] **Step 3: Ejecutar las pruebas y ver que fallan**

Run: `uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q`
Expected: FAIL, con `ModuleNotFoundError: No module named 'verificar_mudanza'`.

- [ ] **Step 4: Escribir el script**

Crea `tools/verificar_mudanza.py`:

```python
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
```

- [ ] **Step 5: Ejecutar las pruebas y ver que pasan**

Run: `uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q`
Expected: PASS, las diez pruebas.

- [ ] **Step 6: Commit**

```bash
git add tools/verificar_mudanza.py tools/requirements.txt tools/tests/test_verificar_mudanza.py
git commit -m "Anade el comprobador de identidad de lineas del reparto"
```

---

## Task 2: La validación de enlaces y anclas de MkDocs

Sin esto, la reescritura de referencias no tiene red. Se pone antes de reescribir nada, y se comprueba que
el estado actual la pasa.

**Files:**
- Modify: `mkdocs.yml`

**Interfaces:**
- Produces: el comando `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
  pasa a fallar con código 1 ante cualquier enlace o ancla que no resuelva dentro de `docs/`.

- [ ] **Step 1: Añadir la validación y el slugify que conserva los acentos**

En `mkdocs.yml`, justo antes del bloque `markdown_extensions:`, añade:

```yaml
# Un enlace a un fichero que no existe o a un ancla que no existe pasa a ser un aviso, y con
# --strict, un fallo de build. Es lo que sostiene la convencion de citar por el titulo de la
# seccion en vez de por su numero: una referencia rota deja de ser un error silencioso.
validation:
  anchors: warn
  unrecognized_links: warn
```

Y dentro de `markdown_extensions:`, cambia la entrada `toc` para que quede así:

```yaml
  - toc:
      permalink: true
      # Sin esto, el titulo "Codigos de salida" genera el ancla "codigos-de-salida", sin los
      # acentos, y un enlace que los lleve falla. Con el slugify de pymdownx las anclas
      # conservan los acentos, quedan legibles y coinciden con las que genera GitHub, asi que
      # un enlace funciona igual en el sitio y leyendo el fichero en crudo.
      slugify: !!python/object/apply:pymdownx.slugs.slugify {kwds: {case: lower}}
```

- [ ] **Step 2: Comprobar que el estado actual pasa**

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0, cero avisos. Medido el 2026-09-10: así es. Si aparece algún aviso, arréglalo antes de
seguir, porque el resto del plan da por hecho que se parte de verde.

- [ ] **Step 3: Comprobar que de verdad detecta una ancla rota**

Una red que no se prueba no es una red. Añade al final de `docs/index.md`, temporalmente:

```markdown
[prueba de la red](SPEC.md#esta-ancla-no-existe)
```

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: FAIL con código 1 y el aviso `does not contain an anchor '#esta-ancla-no-existe'`.

Quita esa línea y vuelve a construir. Expected: código 0.

- [ ] **Step 4: Commit**

```bash
git add mkdocs.yml
git commit -m "Valida enlaces y anclas en el build de la documentacion"
```

---

## Task 3: La mudanza de `SPEC.md` a `docs/spec/`

**Esta tarea no cambia ni una letra de contenido.** Al terminar, las referencias siguen diciendo
"sección 10.4" y el sitio sigue construyendo. Lo único que cambia son los encabezados, que pierden su
número y suben de nivel.

**Files:**
- Delete: `docs/SPEC.md`
- Create: los 34 ficheros de la tabla de abajo
- Modify: `mkdocs.yml` (el `nav`)
- Modify: `docs/index.md` (el enlace a la especificación)

**Interfaces:**
- Consumes: `tools/verificar_mudanza.py` de la tarea 1.
- Produces: los 34 ficheros de `docs/spec/`, que todas las tareas siguientes citan por su ruta.

### El mapa, línea a línea

Los rangos son del `docs/SPEC.md` del commit `3bb97ed`, ambos extremos incluidos. Cubren las 5.762
líneas sin solaparse y sin dejar ninguna fuera. Cuando un fichero recibe dos rangos, van en el orden en
que aparecen en la tabla.

| Fichero | Rangos | Líneas | Título (H1) |
|---|---|---|---|
| `docs/spec/index.md` | 1-29, 507-510 | 33 | `biso`: especificación del CLI de gestión de tareas |
| `docs/spec/vocabulario.md` | 30-63 | 34 | Vocabulario de esta especificación |
| `docs/spec/principios.md` | 64-91 | 28 | Los principios |
| `docs/spec/codigos-de-salida.md` | 92-137 | 46 | Códigos de salida |
| `docs/spec/invocacion.md` | 138-205, 434-505 | 140 | Banderas globales, entorno y configuración de máquina |
| `docs/spec/resolucion-del-tablero.md` | 206-433 | 228 | Cómo se elige el tablero |
| `docs/spec/salida-y-terminal.md` | 511-598 | 88 | Terminal, flujos de salida y codificación |
| `docs/spec/valores-de-entrada.md` | 599-674 | 76 | Cómo se pasa un valor |
| `docs/spec/garantias.md` | 675-721, 746-812 | 114 | Orden de escritura, concurrencia y datos dañados |
| `docs/spec/modelo-de-datos.md` | 722-745, 868-1170 | 327 | El modelo de datos de una tarea |
| `docs/spec/presupuestos.md` | 813-867, 1545-1585 | 96 | Los presupuestos de arranque y de tamaño |
| `docs/spec/vocabularios.md` | 1171-1271 | 101 | Los vocabularios del tablero y la regla de validación |
| `docs/spec/referencias.md` | 1272-1352 | 81 | Cómo se resuelve una referencia a una tarea |
| `docs/spec/familias-de-banderas.md` | 1353-1494 | 142 | Las familias de banderas |
| `docs/spec/cmd/index.md` | 1876-1926 | 51 | Los comandos |
| `docs/spec/cmd/prime.md` | 1495-1544, 1586-1875 | 340 | `biso prime`, el arranque de una sesión |
| `docs/spec/cmd/init.md` | 1927-2349 | 423 | `biso init` |
| `docs/spec/cmd/where.md` | 2350-2485 | 136 | `biso where` |
| `docs/spec/cmd/new.md` | 2486-2705 | 220 | `biso new` |
| `docs/spec/cmd/ls.md` | 2706-3035 | 330 | `biso ls` |
| `docs/spec/cmd/get.md` | 3036-3279 | 244 | `biso get` |
| `docs/spec/cmd/set.md` | 3280-3454 | 175 | `biso set` |
| `docs/spec/cmd/verbos-del-ciclo.md` | 3455-4127 | 673 | Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer` |
| `docs/spec/cmd/archive.md` | 4128-4216 | 89 | `biso archive` |
| `docs/spec/cmd/export.md` | 4217-4325 | 109 | `biso export` |
| `docs/spec/cmd/config.md` | 4326-4597 | 272 | `biso config` |
| `docs/spec/cmd/doctor.md` | 4598-4936 | 339 | `biso doctor` |
| `docs/spec/cmd/board.md` | 4937-5024 | 88 | `biso board` |
| `docs/spec/cmd/help.md` | 5025-5126, 5471-5512 | 144 | La ayuda: `biso help` y `biso --help` |
| `docs/spec/cmd/snapshot.md` | 5127-5470 | 344 | `biso snapshot` |
| `docs/spec/contrato-json.md` | 5513-5641 | 129 | El contrato JSON |
| `docs/spec/estabilidad.md` | 5642-5690 | 49 | El contrato de estabilidad |
| `docs/spec/fuera-de-alcance.md` | 5691-5734 | 44 | Lo que se deja fuera a propósito |
| `docs/spec/por-donde-empezar.md` | 5735-5762 | 28 | Por dónde empezar a implementar |

Las cinco mudanzas que no son un corte limpio, y que el diseño justifica:

- **`biso prime` se une a los comandos** y su presupuesto de tamaño (9.5, líneas 1545-1585) se va a
  `presupuestos.md` con el de arranque (4.13, líneas 813-867). Por eso `cmd/prime.md` recibe dos rangos
  con un hueco en medio.
- **La resolución del tablero** (3.2) sale de dentro de las banderas globales, que se quedan con el
  resto de la sección 3 en `invocacion.md`, también con un hueco en medio.
- **Los identificadores** (4.11, líneas 722-745) se van con el modelo de datos, porque hablan de los
  identificadores de las tareas.
- **La ayuda se reúne**: el comando `biso help` (10.13) y la ayuda de primer nivel (sección 11) van al
  mismo fichero.
- **La introducción de las reglas transversales** (líneas 507-510) va a `docs/spec/index.md`, porque
  describe un grupo que deja de existir como sección y su contenido pasa a formar parte de la explicación
  del orden de lectura. Sus líneas de contenido sobreviven, que es lo que el comprobador exige. La línea
  506, que es el encabezado `## 4. Reglas transversales`, es la única línea del documento viejo que no va
  a ningún sitio, y está declarada en la tabla de excepciones.

**Comprobación aritmética de la tabla:** los rangos cubren las líneas 1 a 5.762 sin solaparse, con la
línea 506 como única excepción declarada. Antes de empezar a cortar, verifícalo con un script de tres
líneas que ordene los rangos y confirme que cada uno empieza donde acabó el anterior. Si no cuadra, el
mapa está mal y hay que arreglarlo antes de tocar nada, no después.

### Las reglas de la mudanza

- **El encabezado de cada fichero pierde su número y pasa a ser un H1** con el título de la tabla.
- **Las subsecciones suben un nivel**: un `###` que era hijo de la sección que ahora es fichero pasa a
  `##`, un `####` pasa a `###`, y así. El comprobador no mira el nivel, solo el texto.
- **Cuando un fichero reúne dos secciones que antes estaban separadas, cada una conserva su propio
  encabezado un nivel por debajo del H1.** Esta regla no es estética, es lo que mantiene corta la lista de
  excepciones: si `cmd/help.md` se queda con un `## \`biso help\`` y un `## La ayuda de primer nivel`, esos
  dos títulos siguen existiendo y el comprobador no los echa en falta.
- **Nada más cambia.** Ni una palabra del cuerpo, ni una referencia, ni un ejemplo, ni una tabla.

### Las once excepciones de encabezado, y no hay más

El comprobador del paso 3 va a reportar exactamente estas y ninguna otra. Cualquier cosa fuera de esta
lista es un error de la mudanza.

**Siete H1 que aparecen de la nada**, porque su fichero reúne piezas que antes no tenían un título común,
o porque el título viejo hablaba de un documento en singular:

| Fichero | H1 nuevo |
|---|---|
| `docs/spec/vocabulario.md` | Vocabulario de esta especificación |
| `docs/spec/invocacion.md` | Banderas globales, entorno y configuración de máquina |
| `docs/spec/salida-y-terminal.md` | Terminal, flujos de salida y codificación |
| `docs/spec/valores-de-entrada.md` | Cómo se pasa un valor |
| `docs/spec/garantias.md` | Orden de escritura, concurrencia y datos dañados |
| `docs/spec/presupuestos.md` | Los presupuestos de arranque y de tamaño |
| `docs/spec/cmd/help.md` | La ayuda: `biso help` y `biso --help` |

**Dos títulos viejos que se pierden**, porque nombraban agrupaciones que dejan de existir:

| Título que desaparece | Por qué |
|---|---|
| Vocabulario de este documento | Lo sustituye el H1 de `vocabulario.md`, porque ya no hay un documento en singular |
| Reglas transversales | Era el nombre del cajón de la sección 4, que se reparte en cuatro ficheros por temas. Su prosa introductoria (líneas 507-510) sobrevive en `docs/spec/index.md`, y solo se pierde la línea del encabezado |

**Los demás H1 tienen que coincidir carácter a carácter con el encabezado del que salen**, quitándole el
número. Si el comprobador reporta un octavo H1 nuevo o un tercer título perdido, hay un corte mal hecho.

- [ ] **Step 1: Crear los ficheros con los rangos de la tabla**

Trabaja fichero a fichero, y extrae los rangos con `sed` en vez de copiando a mano, que es donde se
pierden líneas. Para un fichero de un solo rango:

```bash
mkdir -p docs/spec/cmd
{ echo '# `biso ls`'; echo; sed -n '2707,3035p' docs/SPEC.md; } > docs/spec/cmd/ls.md
```

Fíjate en que el rango empieza en 2707 y no en 2706: la primera línea del rango es el encabezado viejo
`### 10.4. \`biso ls\``, que se sustituye por el H1 nuevo y no se copia. Para un fichero de dos rangos, los
dos `sed` seguidos en el orden de la tabla, saltándose en cada uno su línea de encabezado solo si ese
encabezado es el que se convierte en el H1.

Después, ajusta el nivel de los encabezados internos, que suben uno: un `####` pasa a `###`. Hazlo con una
sustitución acotada al principio de línea y fuera de los bloques de código.

No borres `docs/SPEC.md` todavía: el comprobador del paso 3 lo necesita.

- [ ] **Step 2: Comprobar que ningún fichero pasa de 700 líneas**

```bash
wc -l docs/spec/*.md docs/spec/cmd/*.md | sort -rn | head -5
```

Expected: el mayor es `docs/spec/cmd/verbos-del-ciclo.md`, con unas 673 líneas más el H1.

- [ ] **Step 3: Comprobar que no se ha perdido nada**

```bash
python3 tools/verificar_mudanza.py docs/SPEC.md docs/spec/*.md docs/spec/cmd/*.md
```

Expected: `la mudanza es fiel: ninguna linea ni encabezado se perdio o aparecio`, código 0.

Los encabezados que el comprobador reporte tienen que ser exactamente los siete H1 nuevos y los dos
títulos perdidos de la tabla de excepciones, y ninguno más. **Si aparece cualquier otro, es un error de la
mudanza y hay que arreglarlo, nunca añadirlo a las excepciones.** Las líneas de contenido, en cambio,
tienen que cuadrar sin ninguna excepción: si el comprobador reporta una sola línea perdida o sobrante, el
paso no está terminado.

- [ ] **Step 4: Borrar `docs/SPEC.md` y rehacer la navegación**

Borra `docs/SPEC.md`. En `mkdocs.yml`, sustituye la entrada `- Especificacion: SPEC.md` por el árbol
completo, en el orden de lectura que fija la tabla, con la sección de comandos anidada. En
`docs/index.md`, cambia el enlace `[Especificación](SPEC.md)` por `[Especificación](spec/index.md)`.

- [ ] **Step 5: Construir el sitio**

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0. Los enlaces internos siguen siendo texto plano con números, así que la validación no
tiene nada que objetar todavía.

- [ ] **Step 6: Commit**

```bash
git add -A docs/spec docs/SPEC.md docs/index.md mkdocs.yml
git commit -m "Reparte SPEC.md en docs/spec/ sin cambiar una letra de contenido"
```

---

## Task 4: La portada de la especificación y el orden de lectura

La mudanza dejó `docs/spec/index.md` con la introducción vieja, que habla de "este documento" en singular
y no dice en qué orden se leen las piezas. Esta tarea es la única del plan que escribe prosa nueva.

**Files:**
- Modify: `docs/spec/index.md`

**Interfaces:**
- Consumes: los 34 ficheros de la tarea 3.

- [ ] **Step 1: Reescribir la introducción para que hable de un conjunto de documentos**

Conserva lo que dice hoy sobre qué define la especificación y qué no, y sobre la convención de idioma.
Cambia las frases que dan por hecho que todo está en un fichero. Añade el orden de lectura: los
fundamentos primero (vocabulario, principios, códigos de salida), luego cómo se invoca y cómo se elige el
tablero, luego el modelo de datos y los vocabularios, luego la gramática de la entrada, luego los
comandos empezando por `prime`, y al final los contratos y lo que queda fuera.

**No escribas un número que cuente los ficheros.** Ni "los 34 documentos" ni "las cinco partes". Es la
regla de la tarea 8 y este es el sitio donde más tienta romperla.

- [ ] **Step 2: Construir el sitio**

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0.

- [ ] **Step 3: Commit**

```bash
git add docs/spec/index.md
git commit -m "Escribe la portada de la especificacion con su orden de lectura"
```

---

## Task 5: Las referencias internas de `docs/spec/`

Son 215, y ahora que el documento está partido muchas cruzan de fichero. Esta tarea sí cambia líneas, y su
prueba es el build.

**Files:**
- Modify: todos los ficheros de `docs/spec/` y `docs/spec/cmd/`

**Interfaces:**
- Consumes: la validación de anclas de la tarea 2.

- [ ] **Step 1: Inventariar las referencias que quedan**

```bash
grep -rEn "(secci[oó]n|apartado)s? +[0-9]+(\.[0-9]+)*" docs/spec/ | wc -l
```

Anota el número de partida. Al terminar la tarea tiene que ser cero.

- [ ] **Step 2: Reescribirlas una por una**

La forma es la del diseño:

```markdown
la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md#cómo-se-elige-el-tablero)
```

Reglas:

- Dentro del mismo fichero, solo el ancla: `["El algoritmo de coincidencia"](#el-algoritmo-de-coincidencia)`.
- Desde `docs/spec/cmd/` hacia `docs/spec/`, un nivel arriba: `../resolucion-del-tablero.md#...`.
- El ancla se saca del título tal cual, en minúsculas, con los espacios convertidos en guiones,
  conservando los acentos y quitando las comillas de código, los dos puntos y las comas. El título
  `` Los verbos del ciclo: `start`, `note`, `comment` `` da `los-verbos-del-ciclo-start-note-comment`.
- Cuando la referencia apuntaba a una sección de nivel 2 que ahora es un fichero entero, el enlace es al
  fichero sin ancla: `[la especificación de `biso doctor`](cmd/doctor.md)`.
- **No inventes el ancla de memoria.** Si dudas, mírala en el encabezado del fichero de destino.

- [ ] **Step 3: Comprobar que no queda ninguna por número**

```bash
grep -rEn "(secci[oó]n|apartado)s? +[0-9]+(\.[0-9]+)*" docs/spec/
```

Expected: sin salida.

- [ ] **Step 4: Construir el sitio, que es donde se ve si algún ancla está mal**

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0. Cada aviso que salga nombra el fichero y el ancla que no existe. Arréglalos y vuelve a
construir hasta que esté en verde.

- [ ] **Step 5: Commit**

```bash
git add docs/spec
git commit -m "Convierte en enlaces por titulo las referencias internas de la especificacion"
```

---

## Task 6: El comprobador de enlaces de fuera de `docs/`, y las referencias de los demás documentos

`mkdocs build --strict` solo valida lo que está dentro de `docs/`. `CLAUDE.md`, `INTEGRATION.md` y los dos
`.md` de `bench/sqlite-driver/` quedan fuera de esa red y necesitan una propia.

**Files:**
- Create: `tools/comprobar_enlaces.py`
- Test: `tools/tests/test_comprobar_enlaces.py`
- Modify: `CLAUDE.md`, `INTEGRATION.md`, `bench/sqlite-driver/README.md`, `bench/sqlite-driver/RESULTADOS.md`
- Modify: `docs/DECISIONES.md`, `docs/PENDIENTES.md`, `docs/ESTADO-DEL-ARTE.md`, `docs/index.md`

**Interfaces:**
- Produces: `enlaces_rotos(ficheros: list[Path], raiz: Path) -> list[str]`, que devuelve una descripción por
  cada enlace relativo que apunte a un fichero que no existe o a un ancla que ese fichero no tiene.
- Produces: la interfaz `python tools/comprobar_enlaces.py <fichero>...`, código 0 si todos resuelven.

- [ ] **Step 1: Escribir las pruebas que fallan**

Crea `tools/tests/test_comprobar_enlaces.py`:

```python
"""Pruebas del comprobador de enlaces de los ficheros de fuera de docs/."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from comprobar_enlaces import enlaces_rotos


def preparar(tmp_path):
    (tmp_path / "docs" / "spec").mkdir(parents=True)
    (tmp_path / "docs" / "spec" / "principios.md").write_text(
        "# Los principios\n\n## El primero\n\nTexto.\n", encoding="utf-8"
    )
    return tmp_path


def test_un_enlace_bueno_no_se_reporta(tmp_path):
    raiz = preparar(tmp_path)
    fichero = raiz / "CLAUDE.md"
    fichero.write_text("[los principios](docs/spec/principios.md#el-primero)\n", encoding="utf-8")
    assert enlaces_rotos([fichero], raiz) == []


def test_un_enlace_a_fichero_que_no_existe_se_reporta(tmp_path):
    raiz = preparar(tmp_path)
    fichero = raiz / "CLAUDE.md"
    fichero.write_text("[fantasma](docs/spec/no-existe.md)\n", encoding="utf-8")
    assert len(enlaces_rotos([fichero], raiz)) == 1


def test_un_enlace_a_ancla_que_no_existe_se_reporta(tmp_path):
    raiz = preparar(tmp_path)
    fichero = raiz / "CLAUDE.md"
    fichero.write_text("[mala](docs/spec/principios.md#el-segundo)\n", encoding="utf-8")
    assert len(enlaces_rotos([fichero], raiz)) == 1


def test_el_ancla_conserva_los_acentos(tmp_path):
    raiz = preparar(tmp_path)
    (raiz / "docs" / "spec" / "acentos.md").write_text(
        "# Uno\n\n## Codigos de salida y por que\n", encoding="utf-8"
    )
    (raiz / "docs" / "spec" / "tildes.md").write_text(
        "# Uno\n\n## Códigos de salida y por qué\n", encoding="utf-8"
    )
    fichero = raiz / "CLAUDE.md"
    fichero.write_text(
        "[a](docs/spec/tildes.md#códigos-de-salida-y-por-qué)\n"
        "[b](docs/spec/acentos.md#codigos-de-salida-y-por-que)\n",
        encoding="utf-8",
    )
    assert enlaces_rotos([fichero], raiz) == []


def test_una_url_externa_no_se_comprueba(tmp_path):
    raiz = preparar(tmp_path)
    fichero = raiz / "CLAUDE.md"
    fichero.write_text("[fuera](https://example.com/x#y)\n", encoding="utf-8")
    assert enlaces_rotos([fichero], raiz) == []


def test_un_enlace_dentro_de_un_bloque_de_codigo_no_se_comprueba(tmp_path):
    raiz = preparar(tmp_path)
    fichero = raiz / "CLAUDE.md"
    fichero.write_text("```\n[ejemplo](docs/spec/no-existe.md)\n```\n", encoding="utf-8")
    assert enlaces_rotos([fichero], raiz) == []
```

- [ ] **Step 2: Ejecutar las pruebas y ver que fallan**

Run: `uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q`
Expected: FAIL con `ModuleNotFoundError: No module named 'comprobar_enlaces'`.

- [ ] **Step 3: Escribir el script**

Crea `tools/comprobar_enlaces.py`. El ancla de un encabezado se calcula igual que lo hace el slugify de
`pymdownx` configurado en `mkdocs.yml`: pasar a minúsculas, quitar todo lo que no sea letra, número,
espacio o guion (las comillas de código, los dos puntos y las comas se van), y convertir cada racha de
espacios en un solo guion. Los acentos se conservan, así que no se normaliza el unicode.

```python
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
```

- [ ] **Step 4: Ejecutar las pruebas y ver que pasan**

Run: `uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q`
Expected: PASS, las dieciseis pruebas de los dos ficheros.

- [ ] **Step 5: Reescribir las referencias de los ocho documentos**

Con la misma forma de la tarea 5, y con la ruta que corresponda al sitio de cada fichero: los de `docs/`
apuntan a `spec/...`, `CLAUDE.md` a `docs/spec/...`, y los de `bench/sqlite-driver/` a
`../../docs/spec/...`. Son 97 en `docs/DECISIONES.md`, 35 en `INTEGRATION.md`, 24 en
`bench/sqlite-driver/RESULTADOS.md`, 11 en `docs/ESTADO-DEL-ARTE.md`, 10 en `CLAUDE.md`, 9 en
`docs/PENDIENTES.md` y 6 en `bench/sqlite-driver/README.md`. En `docs/index.md`, actualiza además la lista
de documentos del proyecto.

- [ ] **Step 6: Comprobar las dos redes**

```bash
python3 tools/comprobar_enlaces.py CLAUDE.md INTEGRATION.md bench/sqlite-driver/README.md bench/sqlite-driver/RESULTADOS.md
```
Expected: `todos los enlaces relativos resuelven`, código 0.

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0.

- [ ] **Step 7: Comprobar que no queda ninguna referencia por número en ningún documento vivo**

```bash
grep -rEn "(secci[oó]n|apartado)s? +[0-9]+(\.[0-9]+)*" --include="*.md" . | grep -v "^./docs/superpowers/"
```

Expected: sin salida.

- [ ] **Step 8: Commit**

```bash
git add tools CLAUDE.md INTEGRATION.md bench/sqlite-driver docs
git commit -m "Cita la especificacion por el titulo de sus secciones en todo el repositorio"
```

---

## Task 7: La nota en los documentos históricos

Los documentos de `docs/superpowers/` anteriores a este reparto no se tocan por dentro. Solo se les avisa
de que citan con la numeración vieja.

**Files:**
- Modify: `docs/superpowers/plans/2026-09-06-modelo-de-estados.md`
- Modify: `docs/superpowers/plans/2026-09-07-persistencia.md`
- Modify: `docs/superpowers/specs/2026-09-06-modelo-de-estados-design.md`
- Modify: `docs/superpowers/specs/2026-09-07-persistencia-design.md`

**El diseño del tutorial no lleva nota, y esto es una corrección al diseño del reparto.** Su decisión 8
lo mete en la lista de históricos, pero la misma decisión dice después que el tutorial necesita una pasada
propia. No puede ser las dos cosas: una nota que diga "estas referencias se conservan sin tocar" y una
tarea que las toque se contradicen. Manda lo específico: el diseño del tutorial es un encargo con trabajo
en curso, no un acta cerrada, así que no lleva nota y sus referencias las reescribe la tarea 10.

- [ ] **Step 1: Añadir la nota al principio de cada uno**

Justo después del título, con este texto, igual en los cuatro:

```markdown
> **Este documento cita la especificación por el número de sus secciones**, como se escribió en su día.
> El 2026-09-10 `docs/SPEC.md` se repartió en los documentos de `docs/spec/`, que se citan por el título
> de sus secciones. Estas referencias se conservan sin tocar porque este documento es el acta de una
> sesión cerrada. Para traducir una de ellas, mira el mapa de la tabla de la tarea 3 de
> [el plan del reparto](../plans/2026-09-10-reparto-de-la-spec.md).
```

Ajusta la ruta relativa del enlace según si el fichero está en `plans/` o en `specs/`. El del propio plan
no lleva nota, y el diseño del reparto tampoco, porque ya lo dice en su cabecera.

- [ ] **Step 2: Construir el sitio**

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0. Si el enlace al plan da un aviso, es que la ruta relativa está mal.

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers
git commit -m "Avisa en los documentos historicos de que citan con la numeracion vieja"
```

---

## Task 8: El comprobador de recuentos, y las frases que cuentan

**Files:**
- Create: `tools/comprobar_recuentos.py`
- Create: `tools/recuentos-normativos.txt`
- Test: `tools/tests/test_comprobar_recuentos.py`
- Modify: los ficheros de `docs/` que tengan frases que cuentan

**Interfaces:**
- Produces: `frases_que_cuentan(ruta: Path, normativas: set[str]) -> list[tuple[int, str]]`, que devuelve el
  número de línea y la frase de cada recuento no declarado como normativo.
- Produces: `python tools/comprobar_recuentos.py <fichero>...`, código 0 si no hay ninguno sin declarar.

- [ ] **Step 1: Escribir las pruebas que fallan**

Crea `tools/tests/test_comprobar_recuentos.py`:

```python
"""Pruebas del comprobador de frases que cuentan elementos del propio documento."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from comprobar_recuentos import frases_que_cuentan


def escribir(tmp_path, texto):
    ruta = tmp_path / "doc.md"
    ruta.write_text(texto, encoding="utf-8")
    return ruta


def test_detecta_un_numero_en_palabra(tmp_path):
    ruta = escribir(tmp_path, "Siete reglas gobiernan su lectura.\n")
    assert frases_que_cuentan(ruta, set()) == [(1, "Siete reglas")]


def test_detecta_un_numero_en_cifra(tmp_path):
    ruta = escribir(tmp_path, "Hay 18 filas en la tabla.\n")
    assert frases_que_cuentan(ruta, set()) == [(1, "18 filas")]


def test_una_frase_declarada_normativa_no_se_reporta(tmp_path):
    ruta = escribir(tmp_path, "Ocho columnas fijas, separadas por dos espacios.\n")
    assert frases_que_cuentan(ruta, {"Ocho columnas"}) == []


def test_un_sustantivo_que_no_es_del_documento_no_se_reporta(tmp_path):
    """"dos procesos" no cuenta partes del documento, cuenta cosas del mundo."""
    ruta = escribir(tmp_path, "Dos procesos simultaneos no pueden asignar el mismo id.\n")
    assert frases_que_cuentan(ruta, set()) == []


def test_no_mira_dentro_de_los_bloques_de_codigo(tmp_path):
    ruta = escribir(tmp_path, "```\nSiete reglas\n```\n")
    assert frases_que_cuentan(ruta, set()) == []


def test_reporta_cada_aparicion_con_su_linea(tmp_path):
    ruta = escribir(tmp_path, "Siete reglas.\n\nY once comandos.\n")
    assert frases_que_cuentan(ruta, set()) == [(1, "Siete reglas"), (3, "once comandos")]
```

- [ ] **Step 2: Ejecutar las pruebas y ver que fallan**

Run: `uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q`
Expected: FAIL con `ModuleNotFoundError: No module named 'comprobar_recuentos'`.

- [ ] **Step 3: Escribir el script**

Crea `tools/comprobar_recuentos.py`. Los sustantivos que busca son los que nombran partes del propio
documento: `reglas`, `filas`, `comandos`, `codigos`, `códigos`, `banderas`, `secciones`, `apartados`,
`documentos`, `principios`, `garantias`, `garantías`, `mensajes`, `campos`, `verbos`, `criterios`,
`columnas`, `entradas`, `tipos`, `familias`, `errores`, `avisos`, `estados`, `precisiones`,
`comprobaciones`, `requisitos`, `decisiones`. Los cardinales, en cifra o en palabra de `dos` a `treinta`.

```python
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
```

- [ ] **Step 4: Ejecutar las pruebas y ver que pasan**

Run: `uv run --with-requirements tools/requirements.txt --no-project pytest tools/tests -q`
Expected: PASS, las veintidos pruebas de los tres ficheros.

- [ ] **Step 5: Crear la lista de normativas vacía y ver qué encuentra**

Crea `tools/recuentos-normativos.txt` con solo la cabecera:

```
# Frases que cuentan y que se quedan, porque el numero es la regla y hay una prueba que lo
# comprueba, no una descripcion del contenido del documento. Una por linea, con su texto
# exacto. Ver tools/comprobar_recuentos.py.
```

Run: `python3 tools/comprobar_recuentos.py docs/spec/*.md docs/spec/cmd/*.md docs/*.md`
Expected: una lista de unas cuarenta frases, con su fichero y su línea.

- [ ] **Step 6: Clasificar cada frase, una por una**

Para cada una, decide si el número es parte del contrato del programa o una descripción del documento.

Al contrato, y por tanto a `tools/recuentos-normativos.txt`: las ocho columnas de `biso ls`, el mínimo de
tres estados, y las que hablen de un número que una prueba de la suite comprueba.

A reescribir sin número: todas las demás. Los casos concretos que ya se conocen son el párrafo de
`biso doctor` con sus seis números encadenados, en `docs/spec/cmd/doctor.md`; el "Veinte comandos. Los
once primeros... los nueve restantes... De esos once, diez son..." de `docs/spec/cmd/index.md`; las "cuatro
restricciones en todo el documento" del mismo fichero; "Las once reglas que no son adivinables" de
`docs/spec/cmd/prime.md`; y la lista de documentos del proyecto de `docs/index.md`. Sustitúyelos por una
frase que no cuente: "las reglas de abajo", "cada fila de la tabla", "las que la tabla marca con un sí".

**El párrafo de `biso doctor` es el primer cliente y el que justifica la tarea.** Al terminar, tiene que
poder añadirse una fila a su tabla sin tocar ninguna frase de prosa.

- [ ] **Step 7: Avisar de los números de ejemplo que el contrato declara cambiantes**

Esto es la otra mitad del problema de los números, y es la decisión 6 del diseño.

**En los bloques de salida de ejemplo no se ponen placeholders.** Los números se quedan tal cual, porque
son datos del tablero de ejemplo, que es un fixture fijo, y porque el ancho de cada columna se calcula
sobre el contenido, así que un `<N>` de otra anchura desalinearía la tabla entera. Además esas salidas se
comparan carácter a carácter.

**Lo que sí se hace** es añadir una nota debajo del ejemplo allí donde el contrato de estabilidad ya
declara que el valor cambia entre versiones menores. Son dos casos y solo dos:

- **Los valores de urgencia**, porque los coeficientes por defecto de la fórmula pueden cambiar. Afecta a
  los ejemplos de `docs/spec/modelo-de-datos.md`, `docs/spec/cmd/ls.md`, `docs/spec/cmd/get.md`,
  `docs/spec/cmd/prime.md` y `docs/spec/cmd/verbos-del-ciclo.md`.
- **La cadena de versión** del programa, en la primera línea del mensaje de arranque.

La nota, una sola frase debajo del bloque, con esta forma:

```markdown
El valor de urgencia del ejemplo sale de los coeficientes por defecto, que el contrato de estabilidad
permite cambiar entre versiones menores, así que la cifra exacta puede no ser esta.
```

No la repitas en cada ejemplo del mismo fichero: una vez por fichero, en el primero que muestre el valor,
y basta.

- [ ] **Step 8: Comprobar que está limpio**

Run: `python3 tools/comprobar_recuentos.py docs/spec/*.md docs/spec/cmd/*.md docs/*.md`
Expected: `ninguna frase cuenta elementos del documento sin declararlo`, código 0.

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0.

- [ ] **Step 9: Commit**

```bash
git add tools docs
git commit -m "Prohibe las frases que cuentan elementos del documento y arregla las que habia"
```

---

## Task 9: `DECISIONES.md` se reordena por temas

No se parte en ficheros. Cambia el orden, que hoy es el de llegada y no el de afinidad, con la misma prueba
de identidad de líneas de la tarea 3.

**Files:**
- Modify: `docs/DECISIONES.md`

**Interfaces:**
- Consumes: `tools/verificar_mudanza.py` de la tarea 1.

- [ ] **Step 1: Guardar una copia del original para poder comparar**

```bash
cp docs/DECISIONES.md /tmp/DECISIONES-antes.md
```

- [ ] **Step 2: Reagrupar las secciones por tema**

Los agrupamientos que el diseño identifica: las secciones del modelo de estados, que hoy son la 9, la 10 y
la 11, van juntas y podrían ser una sola con subsecciones. Las de rendimiento, que hoy son la 13 y la 14 con
su 14.1, van juntas. El resto se ordena de lo general a lo concreto.

Mueve bloques enteros. No reescribas prosa, salvo una frase de transición cuando el orden nuevo la pida, y
en ese caso anótala como excepción para el paso 4.

**Condensa solo donde dos párrafos digan literalmente lo mismo.** El valor de este documento es la
evidencia que guarda, y resumir evidencia es perderla.

- [ ] **Step 3: Renumerar los encabezados según el orden nuevo**

Las secciones de `DECISIONES.md` se quedan numeradas, porque siguen siendo un solo documento y su número es
su orden de lectura. Lo que cambia es que ya nadie las cita por ese número: las referencias entrantes que la
tarea 6 reescribió apuntan a su título.

- [ ] **Step 4: Comprobar que no se ha perdido nada**

```bash
python3 tools/verificar_mudanza.py /tmp/DECISIONES-antes.md docs/DECISIONES.md
```

Expected: código 0. Los encabezados que reporte como perdidos o sobrantes tienen que ser solo los que
cambiaron de número, y las líneas, ninguna. Si aparece una línea perdida, es contenido que se ha ido y hay
que devolverlo.

- [ ] **Step 5: Comprobar que las referencias entrantes siguen resolviendo**

Los enlaces que apuntan a `DECISIONES.md` desde otros ficheros usan el título de la sección, así que
renumerar no los rompe. Pero si alguna sección cambió de título al reagruparse, sí.

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0.

```bash
python3 tools/comprobar_enlaces.py CLAUDE.md INTEGRATION.md bench/sqlite-driver/README.md bench/sqlite-driver/RESULTADOS.md
```
Expected: código 0.

- [ ] **Step 6: Comprobar los recuentos del fichero reordenado**

Run: `python3 tools/comprobar_recuentos.py docs/DECISIONES.md`
Expected: código 0.

- [ ] **Step 7: Commit**

```bash
git add docs/DECISIONES.md
git commit -m "Reordena DECISIONES.md por temas en vez de por orden de llegada"
```

---

## Task 10: El campo `origen` de los fixtures del tutorial

**Precondición:** el trabajo del tutorial vive en el worktree `worktree-tutorial` y a fecha de este plan
está sin commitear. Esta tarea no se puede hacer hasta que ese trabajo esté en una rama. **Si al llegar
aquí sigue sin commitear, para y dilo, no lo commitees tú.**

**Files:**
- Modify: `tutorial/generar.py`
- Modify: `tutorial/escenarios/01-llegas-a-un-proyecto.yaml`
- Modify: `docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md`

- [ ] **Step 1: Cambiar la forma admitida del campo `origen`**

Hoy el generador valida `origen` contra la expresión `^(literal|derivada) SPEC \S.*$`, y los fixtures
escriben `literal SPEC 4.3`. La forma nueva nombra el fichero y el título:

```yaml
origen: literal spec/salida-y-terminal.md "Notas y avisos"
```

Cambia la expresión regular del generador para exigir esa forma, y su mensaje de error para que diga cuál
es. Mantén las dos palabras `literal` y `derivada`, que es la distinción que sostiene la validación del
tutorial.

- [ ] **Step 2: Actualizar el escenario que ya está escrito**

Traduce cada `origen` de `01-llegas-a-un-proyecto.yaml` a la forma nueva, con el mapa de la tarea 3.

- [ ] **Step 3: Actualizar el diseño del tutorial**

En su sección del formato de los fixtures, cambia el ejemplo y las dos formas admitidas.

- [ ] **Step 4: Generar el tutorial y ver que pasa**

Run: `uv run --with-requirements docs-requirements.txt --no-project python tutorial/generar.py`
Expected: código 0, y `docs/TUTORIAL.md` escrito.

Run: `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
Expected: código 0.

- [ ] **Step 5: Commit**

```bash
git add tutorial docs/TUTORIAL.md docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md
git commit -m "Cita la especificacion por titulo en el campo origen de los fixtures"
```

---

## Comprobación final

Antes de dar el trabajo por terminado, las ocho a la vez:

- [ ] `python3 tools/verificar_mudanza.py` no encuentra ninguna línea perdida ni sobrante.
- [ ] `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict` termina con 0.
- [ ] `python3 tools/comprobar_enlaces.py` termina con 0 sobre los cuatro ficheros de fuera de `docs/`.
- [ ] `grep` de "sección N.N" y "apartado N.N" no encuentra nada fuera de `docs/superpowers/`.
- [ ] `python3 tools/comprobar_recuentos.py` termina con 0 sobre todos los documentos vivos.
- [ ] El párrafo de `biso doctor` ya no encadena números, y se le puede añadir una fila a su tabla sin
      tocar prosa.
- [ ] Ningún fichero de `docs/spec/` pasa de 700 líneas.
- [ ] Ningún nombre de fichero de `docs/spec/` empieza por un número.

Y una que no es automática: **abre `docs/spec/index.md` y léela como si no conocieras el proyecto.** Si no
sabes por dónde empezar después de leerla, la portada no está terminada, aunque los ocho comprobadores
estén en verde.

## Lo que este plan no hace

- No implementa nada de `biso`. El paso 1 del orden de implementación sigue siendo el almacén.
- No parte `DECISIONES.md` en ficheros, solo lo reordena.
- No toca por dentro los documentos históricos de `docs/superpowers/`, solo les añade una nota.
- No cierra ninguna de las entradas abiertas de `docs/PENDIENTES.md`. Desbloquea la de `biso doctor`, que
  es otra cosa: después de la tarea 8 se puede hacer, pero hacerla no es parte de este plan.
