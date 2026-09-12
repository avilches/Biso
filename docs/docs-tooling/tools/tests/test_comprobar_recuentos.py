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
