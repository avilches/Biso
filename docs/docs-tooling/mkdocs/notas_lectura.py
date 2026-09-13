"""MkDocs hook that injects the reading-notes button only for `mkdocs serve` runs, and gives
it a place to save a note without triggering a full page reload.

`mkdocs build` and `mkdocs serve` read the same `mkdocs.yml`, and anything under
`docs/docs-tooling/` (where this file lives) is excluded from the site by `exclude_docs`, so it
is never copied into `site/` and `extra_javascript`/`extra_css` cannot reach it. Instead,
`on_post_page` reads its sibling `notas_lectura.js` and `notas_lectura.css` from disk and
inlines their content into each rendered page, right before `</body>`, only when the
`BISO_DOCS_SERVE` environment variable is set. The `docs-serve` target of the Makefile sets it
before invoking `mkdocs serve`; `docs-build` and `docs-doctor` never do, so the built site
never contains a trace of the button.

`on_serve` wraps the WSGI application of MkDocs' own dev server (the `LiveReloadServer`
instance it hands to this hook) to answer `POST /__notas_lectura__/save` by appending a line
to `notas-pendientes.md`. That file lives at the repository root, deliberately outside
`docs_dir`: `mkdocs serve` watches the whole of `docs_dir` and reloads the browser on any
change under it, and a note file that lived there would cause every saved note to reload the
page the person is reading. `on_serve` is only ever called by `mkdocs serve`, never by
`mkdocs build`, so this endpoint cannot exist in a built site either way.
"""

from __future__ import annotations

import os
from pathlib import Path

NOTES_ENV_VAR = "BISO_DOCS_SERVE"
SAVE_PATH_INFO = "/__notas_lectura__/save"
HERE = Path(__file__).resolve().parent
JS_PATH = HERE / "notas_lectura.js"
CSS_PATH = HERE / "notas_lectura.css"
NOTES_PATH = HERE.parent.parent.parent / "notas-pendientes.md"


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


def append_note(line: str, notes_path: Path | None = None) -> None:
    text = line.strip()
    if not text:
        raise ValueError("empty note")
    target = notes_path if notes_path is not None else NOTES_PATH
    with target.open("a", encoding="utf-8") as handle:
        handle.write(text + "\n")


def save_app(environ, start_response):
    try:
        length = int(environ.get("CONTENT_LENGTH") or 0)
        body = environ["wsgi.input"].read(length).decode("utf-8")
        append_note(body)
    except ValueError:
        start_response("400 Bad Request", [("Content-Type", "text/plain")])
        return [b"empty note"]
    except OSError:
        start_response("500 Internal Server Error", [("Content-Type", "text/plain")])
        return [b"failed to save note"]
    start_response("204 No Content", [])
    return [b""]


def wrap_app_with_save_endpoint(app):
    def wrapped(environ, start_response):
        if (
            environ.get("REQUEST_METHOD") == "POST"
            and environ.get("PATH_INFO") == SAVE_PATH_INFO
        ):
            return save_app(environ, start_response)
        return app(environ, start_response)

    return wrapped


def on_serve(server, config, builder):
    if not notes_enabled():
        return server
    server.set_app(wrap_app_with_save_endpoint(server.get_app()))
    return server
