#!/usr/bin/env python3
"""Generate docs/TUTORIAL.md from the fixtures in tutorial/escenarios/.

Reads every tutorial/escenarios/NN-name.yaml in the order fixed by its numeric prefix, validates
that it has the shape required by
docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md, and writes docs/TUTORIAL.md
preceded by the concepts section (tutorial/conceptos.md).

It depends on nothing beyond the standard library and PyYAML, which MkDocs already pulls in (see
docs-requirements.txt). Run it with:

    uv run --with-requirements docs-requirements.txt --no-project python tutorial/generate.py

When a fixture is incomplete or malformed the script writes nothing and exits non-zero, listing
every problem it found with the file and step each one is in. A generator that swallows an
incomplete fixture and produces an incomplete page destroys the only guarantee this design offers:
that what you read in docs/TUTORIAL.md is exactly what the fixtures say.

Note on language: the code is English, per the rule in CLAUDE.md. The Spanish strings this script
emits *into* the generated page are documentation content, not code, and they all live together in
the PAGE COPY section below.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Any

import yaml

REPO_ROOT = Path(__file__).resolve().parent.parent
SCENARIOS_DIR = REPO_ROOT / "tutorial" / "escenarios"
CONCEPTS_PATH = REPO_ROOT / "tutorial" / "conceptos.md"
OUTPUT_PATH = REPO_ROOT / "docs" / "TUTORIAL.md"

REGENERATE_CMD = (
    "uv run --with-requirements docs-requirements.txt --no-project "
    "python tutorial/generate.py"
)

FILENAME_RE = re.compile(r"^(\d+)-([a-z0-9]+(?:-[a-z0-9]+)*)\.yaml$")
SOURCE_RE = re.compile(r"^(literal|derivada) SPEC \S.*$")

# The fixture files are documentation, so their keys are Spanish. Naming them here keeps that
# Spanish confined to one block instead of scattering it through the code.
KEY_ID = "id"
KEY_TITLE = "titulo"
KEY_SITUATION = "situacion"
KEY_TEACHES = "ensena"
KEY_STEPS = "pasos"
KEY_BOARD_IN = "tablero_entra"
KEY_BOARD_OUT = "tablero_sale"
KEY_NARRATION = "narracion"
KEY_CMD = "cmd"
KEY_OUTPUT = "salida"
KEY_EXIT = "exit"
KEY_SOURCE = "origen"
KEY_REMARK = "comentario"

REQUIRED_SCENARIO_FIELDS = (
    KEY_ID,
    KEY_TITLE,
    KEY_SITUATION,
    KEY_TEACHES,
    KEY_STEPS,
    KEY_BOARD_IN,
    KEY_BOARD_OUT,
)
REQUIRED_STEP_FIELDS = (KEY_CMD, KEY_OUTPUT, KEY_EXIT, KEY_SOURCE)


# --------------------------------------------------------------------------
# PAGE COPY. Spanish on purpose: this is what goes into the generated document.
# --------------------------------------------------------------------------

PAGE_TITLE = "# Tutorial de biso por escenarios"
INDEX_HEADING = "## Escenarios"
TEACHES_HEADING = '!!! abstract "Qué enseña este escenario"'
EXIT_CODE_LINE = "Código de salida: `%s`"
DERIVED_NOTICE = (
    "*(salida derivada de SPEC %s, no es texto literal de la especificación)*"
)
REMARK_LINE = "*Nota: %s*"

HEADER_COMMENT = (
    "<!--\n"
    "  Fichero generado. No lo edites a mano: se sobrescribe entero cada vez que se\n"
    "  ejecuta tutorial/generate.py.\n"
    f"  Se regenera con: {REGENERATE_CMD}\n"
    "-->"
)

GENERATED_ADMONITION = (
    '!!! warning "Documento generado"\n'
    "    Esta página se genera automáticamente a partir de los fixtures de\n"
    "    `tutorial/escenarios/` y de `tutorial/conceptos.md`. No la edites a mano:\n"
    "    cualquier cambio se pierde en la siguiente generación. Para regenerarla:\n"
    "\n"
    "    ```\n"
    f"    {REGENERATE_CMD}\n"
    "    ```"
)


# --------------------------------------------------------------------------
# Helpers
# --------------------------------------------------------------------------


def _rel(path: Path) -> str:
    try:
        return str(path.relative_to(REPO_ROOT))
    except ValueError:
        return str(path)


def _flatten(text: Any) -> str:
    """Collapse multi-line text (titles, `ensena` items) onto a single line."""
    return " ".join(str(text).split())


def _is_missing(value: Any) -> bool:
    return value is None or value == "" or value == []


# --------------------------------------------------------------------------
# Loading and validation
# --------------------------------------------------------------------------


def discover_files() -> tuple[list[tuple[int, Path]], list[str]]:
    errors: list[str] = []
    if not SCENARIOS_DIR.is_dir():
        errors.append(f"{_rel(SCENARIOS_DIR)}: the scenarios directory does not exist")
        return [], errors

    paths = sorted(SCENARIOS_DIR.glob("*.yaml"))
    if not paths:
        errors.append(f"{_rel(SCENARIOS_DIR)}: there is no .yaml file here")
        return [], errors

    numbered: list[tuple[int, Path]] = []
    for path in paths:
        match = FILENAME_RE.match(path.name)
        if not match:
            errors.append(f"{_rel(path)}: the file name does not follow the NN-name.yaml format")
            continue
        numbered.append((int(match.group(1)), path))

    numbered.sort(key=lambda item: item[0])
    return numbered, errors


def validate_step(rel: str, index: int, step: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(step, dict):
        return [f"{rel}, step {index}: is not a mapping of fields"]

    for field in REQUIRED_STEP_FIELDS:
        if field not in step:
            errors.append(f"{rel}, step {index}: missing required field '{field}'")

    # 'cmd' and the source can never be empty. The output can: SPEC 10.5 says that with --section an
    # empty section prints nothing, and that "nothing" is a legitimate result of the step rather than
    # a field somebody forgot to fill in. The one thing the output must not be is null (a bare
    # 'salida:' in the YAML), because that really is a half-written field.
    if KEY_CMD in step and step[KEY_CMD] in (None, ""):
        errors.append(f"{rel}, step {index}: '{KEY_CMD}' cannot be empty")

    if KEY_OUTPUT in step and step[KEY_OUTPUT] is None:
        errors.append(
            f"{rel}, step {index}: '{KEY_OUTPUT}' cannot be null "
            '(use "" for a deliberately empty output)'
        )

    if KEY_EXIT in step and step[KEY_EXIT] is None:
        errors.append(f"{rel}, step {index}: '{KEY_EXIT}' cannot be null")
    elif KEY_EXIT in step:
        exit_value = step[KEY_EXIT]
        if isinstance(exit_value, bool) or not isinstance(exit_value, int):
            errors.append(
                f"{rel}, step {index}: '{KEY_EXIT}' must be an integer, not {exit_value!r}"
            )

    if KEY_SOURCE in step and step[KEY_SOURCE] in (None, ""):
        errors.append(f"{rel}, step {index}: '{KEY_SOURCE}' cannot be empty")
    elif KEY_SOURCE in step:
        source = str(step[KEY_SOURCE])
        if not SOURCE_RE.match(source):
            errors.append(
                f"{rel}, step {index}: '{KEY_SOURCE}' must read 'literal SPEC <section>' "
                f"or 'derivada SPEC <section>', not {source!r}"
            )

    return errors


def validate_scenario(rel: str, path: Path, data: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(data, dict):
        return [f"{rel}: the file does not describe a scenario (a mapping was expected)"]

    for field in REQUIRED_SCENARIO_FIELDS:
        if field not in data or _is_missing(data[field]):
            errors.append(f"{rel}: missing required field '{field}'")

    if KEY_ID in data and not _is_missing(data[KEY_ID]) and data[KEY_ID] != path.stem:
        errors.append(
            f"{rel}: the '{KEY_ID}' field ({data[KEY_ID]!r}) does not match the file name "
            f"({path.stem!r})"
        )

    teaches = data.get(KEY_TEACHES)
    if teaches is not None and (not isinstance(teaches, list) or not teaches):
        errors.append(f"{rel}: '{KEY_TEACHES}' must be a list with at least one item")

    steps = data.get(KEY_STEPS)
    if steps is None:
        return errors  # already reported above as a missing required field
    if not isinstance(steps, list) or not steps:
        errors.append(f"{rel}: '{KEY_STEPS}' must be a list with at least one item")
        return errors

    for index, step in enumerate(steps, start=1):
        errors.extend(validate_step(rel, index, step))

    return errors


def load_scenarios() -> tuple[list[dict], list[str]]:
    numbered, errors = discover_files()
    scenarios: list[dict] = []

    for number, path in numbered:
        rel = _rel(path)
        try:
            text = path.read_text(encoding="utf-8")
        except OSError as exc:
            errors.append(f"{rel}: could not be read ({exc})")
            continue
        try:
            data = yaml.safe_load(text)
        except yaml.YAMLError as exc:
            errors.append(f"{rel}: invalid YAML: {exc}")
            continue

        scenario_errors = validate_scenario(rel, path, data)
        if scenario_errors:
            errors.extend(scenario_errors)
            continue

        data["_number"] = number
        data["_slug"] = f"escenario-{number:02d}"
        scenarios.append(data)

    return scenarios, errors


# --------------------------------------------------------------------------
# Markdown rendering
# --------------------------------------------------------------------------


def demote_headings(markdown_text: str, levels: int = 1) -> str:
    """Push Markdown headings down by `levels`, leaving alone any inside a ``` fence."""
    out_lines = []
    in_fence = False
    heading_re = re.compile(r"^(#+)( .*)$")
    for line in markdown_text.splitlines():
        if line.strip().startswith("```"):
            in_fence = not in_fence
            out_lines.append(line)
            continue
        if not in_fence:
            match = heading_re.match(line)
            if match:
                out_lines.append("#" * (len(match.group(1)) + levels) + match.group(2))
                continue
        out_lines.append(line)
    return "\n".join(out_lines)


def render_console_block(cmd: str, output: str) -> str:
    cmd_lines = str(cmd).rstrip("\n").split("\n")
    body = [f"$ {cmd_lines[0]}", *cmd_lines[1:]]
    stripped = str(output).rstrip("\n")
    if stripped:
        body.extend(stripped.split("\n"))
    return "```console\n" + "\n".join(body) + "\n```"


def render_step(step: dict) -> str:
    parts: list[str] = []

    narration = step.get(KEY_NARRATION)
    if narration and not _is_missing(narration):
        parts.append(str(narration).strip("\n"))

    parts.append(render_console_block(step[KEY_CMD], step[KEY_OUTPUT]))
    parts.append(EXIT_CODE_LINE % step[KEY_EXIT])

    source = str(step[KEY_SOURCE])
    if source.startswith("derivada"):
        parts.append(DERIVED_NOTICE % source.split("SPEC", 1)[1].strip())

    remark = step.get(KEY_REMARK)
    if remark and not _is_missing(remark):
        parts.append(REMARK_LINE % _flatten(remark))

    return "\n\n".join(parts)


def render_teaches(teaches: list) -> str:
    lines = [TEACHES_HEADING]
    for item in teaches:
        lines.append(f"    - {_flatten(item)}")
    return "\n".join(lines)


def render_scenario(data: dict) -> str:
    heading = f"## {data['_number']}. {_flatten(data[KEY_TITLE])} {{: #{data['_slug']} }}"
    situation = str(data[KEY_SITUATION]).strip("\n")
    steps = [render_step(step) for step in data[KEY_STEPS]]
    return "\n\n".join([heading, situation, render_teaches(data[KEY_TEACHES]), *steps])


def render_index(scenarios: list[dict]) -> str:
    lines = [INDEX_HEADING, ""]
    for data in scenarios:
        lines.append(f"{data['_number']}. [{_flatten(data[KEY_TITLE])}](#{data['_slug']})")
    return "\n".join(lines)


def build_document(scenarios: list[dict], concepts_md: str) -> str:
    parts = [
        HEADER_COMMENT,
        PAGE_TITLE,
        GENERATED_ADMONITION,
        demote_headings(concepts_md.strip("\n")),
        render_index(scenarios),
    ]
    parts.extend(render_scenario(data) for data in scenarios)
    return "\n\n".join(parts) + "\n"


# --------------------------------------------------------------------------
# Entry point
# --------------------------------------------------------------------------


def main() -> int:
    scenarios, errors = load_scenarios()

    if not CONCEPTS_PATH.is_file():
        errors.append(f"{_rel(CONCEPTS_PATH)}: does not exist")

    if errors:
        for error in errors:
            print(f"error: {error}", file=sys.stderr)
        print(
            f"error: {len(errors)} problem(s) found; {_rel(OUTPUT_PATH)} was not written",
            file=sys.stderr,
        )
        return 1

    document = build_document(scenarios, CONCEPTS_PATH.read_text(encoding="utf-8"))
    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT_PATH.write_text(document, encoding="utf-8")
    print(f"wrote {_rel(OUTPUT_PATH)} from {len(scenarios)} scenario(s)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
