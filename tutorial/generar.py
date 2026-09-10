#!/usr/bin/env python3
"""Genera docs/TUTORIAL.md a partir de los fixtures de tutorial/escenarios/.

Lee cada fichero de tutorial/escenarios/NN-nombre.yaml en el orden que fija su prefijo
numérico, valida que tiene la forma que docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md
exige, y escribe docs/TUTORIAL.md precedido de la sección de conceptos (tutorial/conceptos.md).

No tiene más dependencias que la biblioteca estándar y PyYAML, que ya arrastra MkDocs
(ver docs-requirements.txt). Se ejecuta con:

    uv run --with-requirements docs-requirements.txt --no-project python tutorial/generar.py

Si algún fixture está incompleto o mal formado, el script no escribe nada y termina con
un código de salida distinto de cero, listando todos los problemas encontrados con el
fichero y el paso donde está cada uno. Un generador que se traga un fixture incompleto y
produce una página incompleta destruye la única garantía que ofrece este diseño: que lo
que se lee en docs/TUTORIAL.md es exactamente lo que hay en los fixtures.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Any

import yaml

REPO_ROOT = Path(__file__).resolve().parent.parent
ESCENARIOS_DIR = REPO_ROOT / "tutorial" / "escenarios"
CONCEPTOS_PATH = REPO_ROOT / "tutorial" / "conceptos.md"
OUTPUT_PATH = REPO_ROOT / "docs" / "TUTORIAL.md"

REGENERATE_CMD = (
    "uv run --with-requirements docs-requirements.txt --no-project "
    "python tutorial/generar.py"
)

FILENAME_RE = re.compile(r"^(\d+)-([a-z0-9]+(?:-[a-z0-9]+)*)\.yaml$")
ORIGEN_RE = re.compile(r"^(literal|derivada) SPEC \S.*$")

REQUIRED_SCENARIO_FIELDS = (
    "id",
    "titulo",
    "situacion",
    "ensena",
    "pasos",
    "tablero_entra",
    "tablero_sale",
)
REQUIRED_STEP_FIELDS = ("cmd", "salida", "exit", "origen")


def _rel(path: Path) -> str:
    try:
        return str(path.relative_to(REPO_ROOT))
    except ValueError:
        return str(path)


def _flatten(text: Any) -> str:
    """Convierte texto multilínea (títulos, items de 'ensena') en una sola línea."""
    return " ".join(str(text).split())


def _is_missing(value: Any) -> bool:
    return value is None or value == "" or value == []


# --------------------------------------------------------------------------
# Carga y validación
# --------------------------------------------------------------------------


def discover_files() -> tuple[list[tuple[int, Path]], list[str]]:
    errors: list[str] = []
    if not ESCENARIOS_DIR.is_dir():
        errors.append(f"{_rel(ESCENARIOS_DIR)}: no existe el directorio de escenarios")
        return [], errors

    paths = sorted(ESCENARIOS_DIR.glob("*.yaml"))
    if not paths:
        errors.append(f"{_rel(ESCENARIOS_DIR)}: no hay ningún fichero .yaml")
        return [], errors

    numbered: list[tuple[int, Path]] = []
    for path in paths:
        m = FILENAME_RE.match(path.name)
        if not m:
            errors.append(
                f"{_rel(path)}: el nombre del fichero no sigue el formato "
                "NN-nombre.yaml"
            )
            continue
        numbered.append((int(m.group(1)), path))

    numbered.sort(key=lambda item: item[0])
    return numbered, errors


def validate_step(rel: str, index: int, paso: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(paso, dict):
        return [f"{rel}, paso {index}: no es un mapa de campos"]

    for field in REQUIRED_STEP_FIELDS:
        if field not in paso:
            errors.append(f"{rel}, paso {index}: falta el campo obligatorio '{field}'")

    # 'cmd' y 'origen' nunca pueden estar vacíos. 'salida' sí puede: la sección 10.5 de
    # la SPEC dice que, con --section, una sección vacía no imprime nada, y ese "nada" es
    # un resultado legítimo del paso, no un campo que se olvidó de rellenar. Lo único que
    # 'salida' no admite es quedar sin valor (un 'salida:' vacío en el YAML, que parsea
    # como None): eso sí es un campo a medio escribir.
    if "cmd" in paso and paso["cmd"] in (None, ""):
        errors.append(f"{rel}, paso {index}: 'cmd' no puede estar vacío")

    if "salida" in paso and paso["salida"] is None:
        errors.append(
            f"{rel}, paso {index}: 'salida' no puede ser nulo (usa \"\" para una "
            "salida deliberadamente vacía)"
        )

    if "exit" in paso and paso["exit"] is None:
        errors.append(f"{rel}, paso {index}: 'exit' no puede ser nulo")
    elif "exit" in paso:
        exit_value = paso["exit"]
        if isinstance(exit_value, bool) or not isinstance(exit_value, int):
            errors.append(
                f"{rel}, paso {index}: 'exit' tiene que ser un entero, no "
                f"{exit_value!r}"
            )

    if "origen" in paso and paso["origen"] in (None, ""):
        errors.append(f"{rel}, paso {index}: 'origen' no puede estar vacío")
    elif "origen" in paso:
        origen = str(paso["origen"])
        if not ORIGEN_RE.match(origen):
            errors.append(
                f"{rel}, paso {index}: 'origen' tiene que tener la forma "
                f"'literal SPEC <sección>' o 'derivada SPEC <sección>', no {origen!r}"
            )

    return errors


def validate_scenario(rel: str, path: Path, data: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(data, dict):
        return [f"{rel}: el fichero no describe un escenario (se esperaba un mapa)"]

    for field in REQUIRED_SCENARIO_FIELDS:
        if field not in data or _is_missing(data[field]):
            errors.append(f"{rel}: falta el campo obligatorio '{field}'")

    if "id" in data and not _is_missing(data["id"]) and data["id"] != path.stem:
        errors.append(
            f"{rel}: el campo 'id' ({data['id']!r}) no coincide con el nombre del "
            f"fichero ({path.stem!r})"
        )

    ensena = data.get("ensena")
    if ensena is not None and (not isinstance(ensena, list) or not ensena):
        errors.append(f"{rel}: 'ensena' tiene que ser una lista con al menos un elemento")

    pasos = data.get("pasos")
    if pasos is None:
        return errors  # ya reportado como campo obligatorio ausente, arriba
    if not isinstance(pasos, list) or not pasos:
        errors.append(f"{rel}: 'pasos' tiene que ser una lista con al menos un elemento")
        return errors

    for i, paso in enumerate(pasos, start=1):
        errors.extend(validate_step(rel, i, paso))

    return errors


def load_scenarios() -> tuple[list[dict], list[str]]:
    numbered, errors = discover_files()
    scenarios: list[dict] = []

    for numero, path in numbered:
        rel = _rel(path)
        try:
            text = path.read_text(encoding="utf-8")
        except OSError as exc:
            errors.append(f"{rel}: no se pudo leer ({exc})")
            continue
        try:
            data = yaml.safe_load(text)
        except yaml.YAMLError as exc:
            errors.append(f"{rel}: YAML inválido: {exc}")
            continue

        scenario_errors = validate_scenario(rel, path, data)
        if scenario_errors:
            errors.extend(scenario_errors)
            continue

        data["_numero"] = numero
        data["_slug"] = f"escenario-{numero:02d}"
        scenarios.append(data)

    return scenarios, errors


# --------------------------------------------------------------------------
# Generación del Markdown
# --------------------------------------------------------------------------

HEADER_COMMENT = (
    "<!--\n"
    "  Fichero generado. No lo edites a mano: se sobrescribe entero cada vez que se\n"
    f"  ejecuta tutorial/generar.py.\n"
    f"  Se regenera con: {REGENERATE_CMD}\n"
    "-->"
)


def generated_admonition() -> str:
    return (
        '!!! warning "Documento generado"\n'
        "    Esta página se genera automáticamente a partir de los fixtures de\n"
        "    `tutorial/escenarios/` y de `tutorial/conceptos.md`. No la edites a mano:\n"
        "    cualquier cambio se pierde en la siguiente generación. Para regenerarla:\n"
        "\n"
        "    ```\n"
        f"    {REGENERATE_CMD}\n"
        "    ```"
    )


def demote_headings(markdown_text: str, levels: int = 1) -> str:
    """Baja el nivel de los encabezados Markdown en `levels`, sin tocar los que están
    dentro de un bloque de código delimitado por ```."""
    out_lines = []
    in_fence = False
    heading_re = re.compile(r"^(#+)( .*)$")
    for line in markdown_text.splitlines():
        stripped = line.strip()
        if stripped.startswith("```"):
            in_fence = not in_fence
            out_lines.append(line)
            continue
        if not in_fence:
            m = heading_re.match(line)
            if m:
                out_lines.append("#" * (len(m.group(1)) + levels) + m.group(2))
                continue
        out_lines.append(line)
    return "\n".join(out_lines)


def render_console_block(cmd: str, salida: str) -> str:
    cmd_lines = str(cmd).rstrip("\n").split("\n")
    body = [f"$ {cmd_lines[0]}", *cmd_lines[1:]]
    salida_stripped = str(salida).rstrip("\n")
    if salida_stripped:
        body.extend(salida_stripped.split("\n"))
    return "```console\n" + "\n".join(body) + "\n```"


def render_step(paso: dict, index: int) -> str:
    parts: list[str] = []

    narracion = paso.get("narracion")
    if narracion and not _is_missing(narracion):
        parts.append(str(narracion).strip("\n"))

    parts.append(render_console_block(paso["cmd"], paso["salida"]))
    parts.append(f"Código de salida: `{paso['exit']}`")

    origen = str(paso["origen"])
    if origen.startswith("derivada"):
        seccion = origen.split("SPEC", 1)[1].strip()
        parts.append(
            f"*(salida derivada de SPEC {seccion}, no es texto literal de la "
            "especificación)*"
        )

    comentario = paso.get("comentario")
    if comentario and not _is_missing(comentario):
        parts.append(f"*Nota: {_flatten(comentario)}*")

    return "\n\n".join(parts)


def render_ensena(ensena: list) -> str:
    lines = ['!!! abstract "Qué enseña este escenario"']
    for item in ensena:
        lines.append(f"    - {_flatten(item)}")
    return "\n".join(lines)


def render_scenario(data: dict) -> str:
    numero = data["_numero"]
    slug = data["_slug"]
    titulo = _flatten(data["titulo"])
    heading = f"## {numero}. {titulo} {{: #{slug} }}"
    situacion = str(data["situacion"]).strip("\n")
    ensena_block = render_ensena(data["ensena"])
    pasos_md = [render_step(paso, i) for i, paso in enumerate(data["pasos"], start=1)]
    return "\n\n".join([heading, situacion, ensena_block, *pasos_md])


def render_index(scenarios: list[dict]) -> str:
    lines = ["## Escenarios", ""]
    for data in scenarios:
        titulo = _flatten(data["titulo"])
        lines.append(f"{data['_numero']}. [{titulo}](#{data['_slug']})")
    return "\n".join(lines)


def build_document(scenarios: list[dict], conceptos_md: str) -> str:
    parts = [
        HEADER_COMMENT,
        "# Tutorial de biso por escenarios",
        generated_admonition(),
        demote_headings(conceptos_md.strip("\n")),
        render_index(scenarios),
    ]
    parts.extend(render_scenario(data) for data in scenarios)
    return "\n\n".join(parts) + "\n"


# --------------------------------------------------------------------------
# Entrada
# --------------------------------------------------------------------------


def main() -> int:
    scenarios, errors = load_scenarios()

    if not CONCEPTOS_PATH.is_file():
        errors.append(f"{_rel(CONCEPTOS_PATH)}: no existe")

    if errors:
        for error in errors:
            print(f"error: {error}", file=sys.stderr)
        print(
            f"error: {len(errors)} problema(s) encontrados; no se ha escrito "
            f"{_rel(OUTPUT_PATH)}",
            file=sys.stderr,
        )
        return 1

    conceptos_md = CONCEPTOS_PATH.read_text(encoding="utf-8")
    document = build_document(scenarios, conceptos_md)
    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT_PATH.write_text(document, encoding="utf-8")
    print(f"escrito {_rel(OUTPUT_PATH)} a partir de {len(scenarios)} escenario(s)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
