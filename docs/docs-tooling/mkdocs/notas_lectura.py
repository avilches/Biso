"""MkDocs hook that injects the reading-notes button only for `mkdocs serve` runs.

`mkdocs build` and `mkdocs serve` read the same `mkdocs.yml`, and anything under
`docs/docs-tooling/` (where this file lives) is excluded from the site by `exclude_docs`, so it
is never copied into `site/` and `extra_javascript`/`extra_css` cannot reach it. Instead, this
hook reads its sibling `notas_lectura.js` and `notas_lectura.css` from disk and inlines their
content into each rendered page, right before `</body>`, only when the `BISO_DOCS_SERVE`
environment variable is set. The `docs-serve` target of the Makefile sets it before invoking
`mkdocs serve`; `docs-build` and `docs-doctor` never do, so the built site never contains a
trace of the button.
"""

from __future__ import annotations

import os
from pathlib import Path

NOTES_ENV_VAR = "BISO_DOCS_SERVE"
HERE = Path(__file__).resolve().parent
JS_PATH = HERE / "notas_lectura.js"
CSS_PATH = HERE / "notas_lectura.css"


def notes_enabled(env: dict[str, str] | None = None) -> bool:
    source = env if env is not None else os.environ
    return bool(source.get(NOTES_ENV_VAR))


def build_snippet(js_text: str, css_text: str) -> str:
    return f"<style>{css_text}</style>\n<script>{js_text}</script>\n"


def on_post_page(output, page, config):
    if not notes_enabled():
        return output
    snippet = build_snippet(
        JS_PATH.read_text(encoding="utf-8"), CSS_PATH.read_text(encoding="utf-8")
    )
    return output.replace("</body>", snippet + "</body>", 1)
