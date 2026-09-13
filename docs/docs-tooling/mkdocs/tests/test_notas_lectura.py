"""Tests for the reading-notes injection hook."""

import io
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import notas_lectura


def test_notes_disabled_by_default(monkeypatch):
    monkeypatch.delenv("BISO_DOCS_SERVE", raising=False)
    assert notas_lectura.notes_enabled() is False


def test_notes_enabled_when_env_var_set(monkeypatch):
    monkeypatch.setenv("BISO_DOCS_SERVE", "1")
    assert notas_lectura.notes_enabled() is True


def test_notes_disabled_when_env_var_set_but_empty(monkeypatch):
    monkeypatch.setenv("BISO_DOCS_SERVE", "")
    assert notas_lectura.notes_enabled() is False


def test_build_snippet_wraps_css_and_js():
    snippet = notas_lectura.build_snippet("console.log(1)", "body{color:red}")
    assert "<style>body{color:red}</style>" in snippet
    assert "<script>console.log(1)</script>" in snippet


def test_on_post_page_leaves_output_untouched_when_disabled(monkeypatch):
    monkeypatch.delenv("BISO_DOCS_SERVE", raising=False)
    output = "<html><body>hola</body></html>"
    assert notas_lectura.on_post_page(output, page=None, config=None) == output


def test_on_post_page_injects_snippet_before_closing_body(monkeypatch, tmp_path):
    js_file = tmp_path / "notas_lectura.js"
    css_file = tmp_path / "notas_lectura.css"
    js_file.write_text("console.log('hola')", encoding="utf-8")
    css_file.write_text("body{color:red}", encoding="utf-8")
    monkeypatch.setattr(notas_lectura, "JS_PATH", js_file)
    monkeypatch.setattr(notas_lectura, "CSS_PATH", css_file)
    monkeypatch.setenv("BISO_DOCS_SERVE", "1")

    output = "<html><body>hola</body></html>"
    result = notas_lectura.on_post_page(output, page=None, config=None)

    assert "console.log('hola')" in result
    assert "body{color:red}" in result
    assert result.endswith("</body></html>")
    assert result.index("<script>") < result.index("</body>")


def test_append_note_writes_a_trimmed_line(tmp_path):
    notes_file = tmp_path / "notas-pendientes.md"
    notes_file.write_text("<!-- header -->\n", encoding="utf-8")

    notas_lectura.append_note("  - [ ] /x/ | una nota  \n", notes_path=notes_file)

    assert notes_file.read_text(encoding="utf-8") == "<!-- header -->\n- [ ] /x/ | una nota\n"


def test_append_note_rejects_an_empty_line(tmp_path):
    notes_file = tmp_path / "notas-pendientes.md"
    notes_file.write_text("", encoding="utf-8")

    with pytest.raises(ValueError):
        notas_lectura.append_note("   ", notes_path=notes_file)


def _call_wsgi_app(app, method, path_info, body=b""):
    environ = {
        "REQUEST_METHOD": method,
        "PATH_INFO": path_info,
        "CONTENT_LENGTH": str(len(body)),
        "wsgi.input": io.BytesIO(body),
    }
    captured = {}

    def start_response(status, headers):
        captured["status"] = status
        captured["headers"] = headers

    result = list(app(environ, start_response))
    return captured["status"], captured["headers"], result


def test_save_app_appends_the_posted_note_and_returns_204(tmp_path, monkeypatch):
    notes_file = tmp_path / "notas-pendientes.md"
    notes_file.write_text("", encoding="utf-8")
    monkeypatch.setattr(notas_lectura, "NOTES_PATH", notes_file)

    status, _headers, _body = _call_wsgi_app(
        notas_lectura.save_app, "POST", notas_lectura.SAVE_PATH_INFO, b"- [ ] /x/ | hola"
    )

    assert status == "204 No Content"
    assert notes_file.read_text(encoding="utf-8") == "- [ ] /x/ | hola\n"


def test_save_app_rejects_an_empty_body(tmp_path, monkeypatch):
    notes_file = tmp_path / "notas-pendientes.md"
    notes_file.write_text("", encoding="utf-8")
    monkeypatch.setattr(notas_lectura, "NOTES_PATH", notes_file)

    status, _headers, _body = _call_wsgi_app(
        notas_lectura.save_app, "POST", notas_lectura.SAVE_PATH_INFO, b"   "
    )

    assert status == "400 Bad Request"
    assert notes_file.read_text(encoding="utf-8") == ""


def test_wrap_app_routes_the_save_path_to_save_app(monkeypatch, tmp_path):
    notes_file = tmp_path / "notas-pendientes.md"
    notes_file.write_text("", encoding="utf-8")
    monkeypatch.setattr(notas_lectura, "NOTES_PATH", notes_file)

    def original_app(environ, start_response):
        start_response("200 OK", [])
        return [b"original"]

    wrapped = notas_lectura.wrap_app_with_save_endpoint(original_app)

    status, _headers, body = _call_wsgi_app(
        wrapped, "POST", notas_lectura.SAVE_PATH_INFO, b"- [ ] /x/ | hola"
    )
    assert status == "204 No Content"

    status, _headers, body = _call_wsgi_app(wrapped, "GET", "/spec/cmd/ls/")
    assert status == "200 OK"
    assert body == [b"original"]


def test_wrap_app_delegates_a_get_on_the_save_path_to_the_original_app(monkeypatch, tmp_path):
    notes_file = tmp_path / "notas-pendientes.md"
    notes_file.write_text("", encoding="utf-8")
    monkeypatch.setattr(notas_lectura, "NOTES_PATH", notes_file)

    def original_app(environ, start_response):
        start_response("404 Not Found", [])
        return [b"not the save endpoint"]

    wrapped = notas_lectura.wrap_app_with_save_endpoint(original_app)

    status, _headers, body = _call_wsgi_app(wrapped, "GET", notas_lectura.SAVE_PATH_INFO)

    assert status == "404 Not Found"
    assert body == [b"not the save endpoint"]
    assert notes_file.read_text(encoding="utf-8") == ""


def test_on_serve_wraps_the_app_only_when_notes_enabled(monkeypatch):
    class FakeServer:
        def __init__(self):
            self._app = "original-app"

        def get_app(self):
            return self._app

        def set_app(self, app):
            self._app = app

    monkeypatch.delenv("BISO_DOCS_SERVE", raising=False)
    server = FakeServer()
    notas_lectura.on_serve(server, config=None, builder=None)
    assert server.get_app() == "original-app"

    monkeypatch.setenv("BISO_DOCS_SERVE", "1")
    server = FakeServer()
    notas_lectura.on_serve(server, config=None, builder=None)
    assert server.get_app() != "original-app"
    assert callable(server.get_app())
