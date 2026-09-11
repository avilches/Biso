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
