"""Tests for the reading-notes injection hook."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import notas_lectura


def test_notes_disabled_by_default(monkeypatch):
    monkeypatch.delenv("BISO_DOCS_SERVE", raising=False)
    assert notas_lectura.notes_enabled() is False


def test_notes_enabled_when_env_var_set(monkeypatch):
    monkeypatch.setenv("BISO_DOCS_SERVE", "1")
    assert notas_lectura.notes_enabled() is True


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
