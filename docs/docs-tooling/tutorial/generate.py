#!/usr/bin/env python3
"""Generate the tutorial pages from the fixtures in tutorial/escenarios/ and tutorial/conceptos.md.

Reads every tutorial/escenarios/NN-name.yaml in the order fixed by its numeric prefix, validates
that it has the shape required by
docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md and
docs/superpowers/specs/2026-09-11-origen-combinado-y-tutorial-en-ingles-design.md, and writes:

- docs/tutorial/index.md: the tutorial's landing page, with the numbered list of scenarios.
- docs/tutorial/NN-name.md: one page per scenario, named after its fixture.
- docs/concepts.md: the concepts page, copied from tutorial/conceptos.md.

It depends on nothing beyond the standard library and PyYAML, which MkDocs already pulls in (see
docs-requirements.txt). Run it with:

    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \
        python docs/docs-tooling/tutorial/generate.py

When a fixture is incomplete or malformed the script writes nothing and exits non-zero, listing
every problem it found with the file and step each one is in. A generator that swallows an
incomplete fixture and produces an incomplete page destroys the only guarantee this design offers:
that what you read in the generated pages is exactly what the fixtures say.
"""

from __future__ import annotations

import re
import shutil
import sys
from pathlib import Path
from typing import Any

import yaml

TUTORIAL_DIR = Path(__file__).resolve().parent
REPO_ROOT = TUTORIAL_DIR.parent.parent.parent
SCENARIOS_DIR = TUTORIAL_DIR / "escenarios"
CONCEPTS_PATH = TUTORIAL_DIR / "conceptos.md"
TUTORIAL_OUTPUT_DIR = REPO_ROOT / "docs" / "tutorial"
CONCEPTS_OUTPUT_PATH = REPO_ROOT / "docs" / "concepts.md"
OBSOLETE_OUTPUT_PATH = REPO_ROOT / "docs" / "TUTORIAL.md"

REGENERATE_CMD = (
    "uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project "
    "python docs/docs-tooling/tutorial/generate.py"
)

FILENAME_RE = re.compile(r"^(\d+)-([a-z0-9]+(?:-[a-z0-9]+)*)\.yaml$")
SOURCE_KINDS = ("literal", "derived")
SOURCE_ITEM_RE = re.compile(r'^(\S+\.md(?:#\S+)?) "(.+)"$')

KEY_ID = "id"
KEY_TITLE = "title"
KEY_SITUATION = "situation"
KEY_TEACHES = "teaches"
KEY_STEPS = "steps"
KEY_BOARD_IN = "board_in"
KEY_BOARD_OUT = "board_out"
KEY_NARRATION = "narration"
KEY_CMD = "command"
KEY_OUTPUT = "output"
KEY_EXIT = "exit_code"
KEY_SOURCE_KIND = "source_kind"
KEY_SOURCE = "source"
KEY_REMARK = "remark"

REQUIRED_SCENARIO_FIELDS = (
    KEY_ID,
    KEY_TITLE,
    KEY_SITUATION,
    KEY_TEACHES,
    KEY_STEPS,
    KEY_BOARD_IN,
    KEY_BOARD_OUT,
)
REQUIRED_STEP_FIELDS = (KEY_CMD, KEY_OUTPUT, KEY_EXIT, KEY_SOURCE_KIND, KEY_SOURCE)


# --------------------------------------------------------------------------
# Page copy
# --------------------------------------------------------------------------

INDEX_TITLE = "# biso tutorial, by scenario"
INDEX_INTRO = (
    "Thirteen scenarios follow the life of `TASK-19`, from the moment it is handed out to the "
    "moment it closes. Read them in order: each one continues on the same example board where "
    "the last left off.\n"
    "\n"
    "New to `biso`'s vocabulary? Start with [Concepts](../concepts.md)."
)
INDEX_HEADING = "## Scenarios"
TEACHES_HEADING = '!!! abstract "What this scenario teaches"'
EXIT_CODE_LINE = "Exit code: `%s`"
DERIVED_NOTICE = "*(derived output, see %s; not literal spec text)*"
REMARK_LINE = "*Note: %s*"

HEADER_COMMENT = (
    "<!--\n"
    "  Generated file. Do not edit by hand: it is overwritten entirely every time\n"
    "  tutorial/generate.py runs.\n"
    f"  Regenerate it with: {REGENERATE_CMD}\n"
    "-->"
)

GENERATED_ADMONITION = (
    '!!! warning "Generated document"\n'
    "    This page is generated automatically from the fixtures in\n"
    "    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the\n"
    "    next generation. To regenerate it:\n"
    "\n"
    "    ```\n"
    f"    {REGENERATE_CMD}\n"
    "    ```"
)

CONCEPTS_ADMONITION = (
    '!!! warning "Generated document"\n'
    "    This page is generated automatically from `tutorial/conceptos.md`. Do not\n"
    "    edit it by hand: any change is lost on the next generation. To regenerate\n"
    "    it:\n"
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
    """Collapse multi-line text (titles, `teaches` items) onto a single line."""
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


def validate_source(rel: str, index: int, step: dict) -> list[str]:
    errors: list[str] = []

    kind = step.get(KEY_SOURCE_KIND)
    if _is_missing(kind):
        errors.append(f"{rel}, step {index}: '{KEY_SOURCE_KIND}' cannot be empty")
    elif kind not in SOURCE_KINDS:
        errors.append(
            f"{rel}, step {index}: '{KEY_SOURCE_KIND}' must be 'literal' or 'derived', not {kind!r}"
        )

    source = step.get(KEY_SOURCE)
    if _is_missing(source):
        errors.append(f"{rel}, step {index}: '{KEY_SOURCE}' cannot be empty")
        return errors
    if not isinstance(source, list):
        errors.append(f"{rel}, step {index}: '{KEY_SOURCE}' must be a list")
        return errors

    if kind == "literal" and len(source) != 1:
        errors.append(
            f"{rel}, step {index}: a 'literal' '{KEY_SOURCE_KIND}' must have exactly one "
            f"'{KEY_SOURCE}' item, not {len(source)}"
        )

    for item in source:
        if not SOURCE_ITEM_RE.match(str(item)):
            errors.append(
                f"{rel}, step {index}: '{KEY_SOURCE}' item must read 'path/to/file.md "
                f'"Section title"\', not {item!r}'
            )

    return errors


def validate_step(rel: str, index: int, step: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(step, dict):
        return [f"{rel}, step {index}: is not a mapping of fields"]

    for field in REQUIRED_STEP_FIELDS:
        if field not in step:
            errors.append(f"{rel}, step {index}: missing required field '{field}'")

    # The command and the source can never be empty. The output can: the spec says that with
    # --section an empty section prints nothing, and that "nothing" is a legitimate result of the
    # step rather than a field somebody forgot to fill in. The one thing the output must not be is
    # null (a bare 'output:' in the YAML), because that really is a half-written field.
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

    if KEY_SOURCE_KIND in step or KEY_SOURCE in step:
        errors.extend(validate_source(rel, index, step))

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
        data["_filename"] = f"{path.stem}.md"
        scenarios.append(data)

    return scenarios, errors


# --------------------------------------------------------------------------
# Markdown rendering
# --------------------------------------------------------------------------


def render_console_block(cmd: str, output: str) -> str:
    cmd_lines = str(cmd).rstrip("\n").split("\n")
    body = [f"$ {cmd_lines[0]}", *cmd_lines[1:]]
    stripped = str(output).rstrip("\n")
    if stripped:
        body.extend(stripped.split("\n"))
    return "```console\n" + "\n".join(body) + "\n```"


def render_source_links(source: list[str]) -> str:
    """Turn ['a.md "Title"', ...] into linked, code-formatted references.

    Each path in a fixture's `source` is written relative to `docs/`, per the contract in
    tutorial/CLAUDE.md. Scenario pages live one level below that, in docs/tutorial/, so the link
    needs the matching `../` to still resolve.
    """
    links = []
    for item in source:
        match = SOURCE_ITEM_RE.match(item)
        path, title = match.group(1), match.group(2)
        links.append(f"[{title}](../{path})")
    return ", ".join(links)


def render_step(step: dict) -> str:
    parts: list[str] = []

    narration = step.get(KEY_NARRATION)
    if narration and not _is_missing(narration):
        parts.append(str(narration).strip("\n"))

    parts.append(render_console_block(step[KEY_CMD], step[KEY_OUTPUT]))
    parts.append(EXIT_CODE_LINE % step[KEY_EXIT])

    if step.get(KEY_SOURCE_KIND) == "derived":
        parts.append(DERIVED_NOTICE % render_source_links(step[KEY_SOURCE]))

    remark = step.get(KEY_REMARK)
    if remark and not _is_missing(remark):
        parts.append(REMARK_LINE % _flatten(remark))

    return "\n\n".join(parts)


def render_teaches(teaches: list) -> str:
    lines = [TEACHES_HEADING]
    for item in teaches:
        lines.append(f"    - {_flatten(item)}")
    return "\n".join(lines)


def render_scenario_page(data: dict) -> str:
    heading = f"# {data['_number']}. {_flatten(data[KEY_TITLE])}"
    situation = str(data[KEY_SITUATION]).strip("\n")
    steps = [render_step(step) for step in data[KEY_STEPS]]
    parts = [
        HEADER_COMMENT,
        heading,
        GENERATED_ADMONITION,
        situation,
        render_teaches(data[KEY_TEACHES]),
        *steps,
    ]
    return "\n\n".join(parts) + "\n"


def render_index_page(scenarios: list[dict]) -> str:
    lines = [INDEX_HEADING, ""]
    for data in scenarios:
        lines.append(f"{data['_number']}. [{_flatten(data[KEY_TITLE])}]({data['_filename']})")
    parts = [
        HEADER_COMMENT,
        INDEX_TITLE,
        GENERATED_ADMONITION,
        INDEX_INTRO,
        "\n".join(lines),
    ]
    return "\n\n".join(parts) + "\n"


def render_concepts_page(concepts_md: str) -> str:
    parts = [HEADER_COMMENT, CONCEPTS_ADMONITION, concepts_md.strip("\n")]
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
            f"error: {len(errors)} problem(s) found; the tutorial pages were not written",
            file=sys.stderr,
        )
        return 1

    index_page = render_index_page(scenarios)
    concepts_page = render_concepts_page(CONCEPTS_PATH.read_text(encoding="utf-8"))
    scenario_pages = {data["_filename"]: render_scenario_page(data) for data in scenarios}

    if TUTORIAL_OUTPUT_DIR.exists():
        shutil.rmtree(TUTORIAL_OUTPUT_DIR)
    TUTORIAL_OUTPUT_DIR.mkdir(parents=True)
    (TUTORIAL_OUTPUT_DIR / "index.md").write_text(index_page, encoding="utf-8")
    for filename, content in scenario_pages.items():
        (TUTORIAL_OUTPUT_DIR / filename).write_text(content, encoding="utf-8")

    CONCEPTS_OUTPUT_PATH.write_text(concepts_page, encoding="utf-8")

    if OBSOLETE_OUTPUT_PATH.exists():
        OBSOLETE_OUTPUT_PATH.unlink()

    print(
        f"wrote {_rel(TUTORIAL_OUTPUT_DIR)}/ ({len(scenarios)} scenario page(s) + index.md) "
        f"and {_rel(CONCEPTS_OUTPUT_PATH)}"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
