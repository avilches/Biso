#!/usr/bin/env python3
"""Ejecuta de punta a punta el pipeline de comprobacion y generacion de la documentacion.

Encadena, en este orden exacto, los pasos que hoy se ejecutan sueltos: comprobar los enlaces
de los ficheros que quedan fuera de docs/, regenerar las paginas del tutorial, comprobar los
recuentos y las referencias por numero de docs/spec/, docs/*.md, docs/tutorial/*.md, docs/decisiones/*.md
y docs/estado-del-arte/*.md, y construir el sitio con
MkDocs en modo estricto. El orden no es arbitrario: la regeneracion del tutorial va antes de
comprobar_recuentos.py porque ese comprobador revisa las paginas generadas, y si se ejecutara
sobre la version vieja estaria comprobando lo que no se va a commitear.

No se detiene en el primer fallo: ejecuta los cinco pasos siempre, acumula cuales fallaron e
imprime un resumen al final, igual que hace `biso doctor` con las comprobaciones del programa.

Se ejecuta desde la raiz del repositorio, con MkDocs instalado en el entorno:

    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \\
        python docs/docs-tooling/tools/doctor.py

Termina con codigo 0 si los cinco pasos salen bien y 1 si alguno falla.
"""

from __future__ import annotations

import glob
import subprocess
import sys
from pathlib import Path

TOOLS_DIR = Path(__file__).resolve().parent
DOCS_TOOLING_DIR = TOOLS_DIR.parent
TUTORIAL_DIR = DOCS_TOOLING_DIR / "tutorial"
MKDOCS_DIR = DOCS_TOOLING_DIR / "mkdocs"


def expand(*patterns: str) -> list[str]:
    """Expande patrones estilo shell: subprocess no invoca una shell que lo haga por nosotros."""
    ficheros: list[str] = []
    for patron in patterns:
        ficheros.extend(sorted(glob.glob(patron)))
    return ficheros


def run_step(name: str, command: list[str]) -> bool:
    print(f"==> {name}")
    result = subprocess.run(command)
    ok = result.returncode == 0
    print(f"<== {name}: {'OK' if ok else 'FALLO'} (codigo {result.returncode})\n")
    return ok


def main(argv: list[str]) -> int:
    del argv
    steps: list[tuple[str, list[str]]] = [
        (
            "comprobar_enlaces",
            [
                sys.executable,
                str(TOOLS_DIR / "comprobar_enlaces.py"),
                "CLAUDE.md",
                "bench/sqlite-driver/README.md",
                "bench/sqlite-driver/RESULTADOS.md",
            ],
        ),
        (
            "regenerar las paginas del tutorial",
            [sys.executable, str(TUTORIAL_DIR / "generate.py")],
        ),
        (
            "comprobar_recuentos",
            [
                sys.executable,
                str(TOOLS_DIR / "comprobar_recuentos.py"),
                *expand(
                    "docs/spec/*.md",
                    "docs/spec/cmd/*.md",
                    "docs/spec/modelo-de-datos/*.md",
                    "docs/*.md",
                    "docs/tutorial/*.md",
                    "docs/decisiones/*.md",
                    "docs/estado-del-arte/*.md",
                ),
            ],
        ),
        (
            "comprobar_referencias",
            [
                sys.executable,
                str(TOOLS_DIR / "comprobar_referencias.py"),
                *expand("docs/spec/*.md", "docs/spec/cmd/*.md", "docs/spec/modelo-de-datos/*.md"),
            ],
        ),
        (
            "mkdocs build --strict",
            ["mkdocs", "build", "--strict", "-f", str(MKDOCS_DIR / "mkdocs.yml")],
        ),
    ]

    failed = [name for name, command in steps if not run_step(name, command)]

    print("=== resumen ===")
    for name, _ in steps:
        print(f"{'FALLO' if name in failed else 'OK'}: {name}")

    if failed:
        print(f"\n{len(failed)} de {len(steps)} pasos fallaron", file=sys.stderr)
        return 1
    print("\ntodos los pasos salieron bien")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
