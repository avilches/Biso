"""Calcula la urgencia de cada tarea del tablero de ejemplo segun la seccion 5.4 de docs/SPEC.md.

Existe para que ningun fixture tenga que hacer esa aritmetica a mano. Los escenarios del tutorial
ordenan listas por urgencia, y ese orden hay que poder reproducirlo sin creerse un numero escrito
por alguien. Se ejecuta con:

    uv run --with-requirements docs-requirements.txt --no-project python tutorial/urgencia.py

Imprime cada tarea con su urgencia y el desglose de los terminos que la componen, ya ordenadas por
la regla de orden por defecto de `biso ls` (seccion 10.4): primero las que tienen `ordinal`, luego
las demas por urgencia descendente, y cualquier empate por identificador ascendente.
"""

import datetime
import pathlib
import sys

import yaml

RAIZ = pathlib.Path(__file__).resolve().parent
TABLERO = RAIZ / "tablero.yaml"

# Coeficientes de la seccion 5.4. Son configurables en un tablero real (las siete claves
# urgency.* de 10.10); aqui se usan los valores por defecto, que son los del tablero de ejemplo.
PESO_PRIORIDAD = 6.0
PESO_ACTIVA = 4.0
PESO_BLOQUEA = 8.0
PESO_BLOQUEADA = -5.0
PESO_PROXIMIDAD = 12.0
PESO_CRITERIOS = 1.0
PESO_EDAD = 0.5

VALOR_PRIORIDAD = {"high": 1.0, "medium": 0.5, "low": 0.0, None: 0.3}


def dia(texto):
    return datetime.date.fromisoformat(str(texto)[:10])


def calcular(tarea, tablero, tareas_por_id, hoy):
    activo = next(e["name"] for e in tablero["statuses"] if e.get("role") == "active")
    terminal = next(e["name"] for e in tablero["statuses"] if e.get("role") == "terminal")

    if tarea["status"] == terminal:
        return 0.0, {"terminal": True}

    terminos = {}

    terminos["prioridad"] = PESO_PRIORIDAD * VALOR_PRIORIDAD[tarea.get("priority")]

    tiene_pregunta = bool(tarea.get("question"))
    activa = 1.0 if (tarea["status"] == activo and not tiene_pregunta) else 0.0
    terminos["activa"] = PESO_ACTIVA * activa

    # "bloquea" es cierto si alguna tarea sin terminar depende de esta.
    bloquea = any(
        tarea["id"] in (otra.get("depends") or []) and otra["status"] != terminal
        for otra in tareas_por_id.values()
    )
    terminos["bloquea"] = PESO_BLOQUEA * (1.0 if bloquea else 0.0)

    # "bloqueada" es cierto si esta depende de alguna tarea sin terminar.
    bloqueada = any(
        tareas_por_id[dep]["status"] != terminal
        for dep in (tarea.get("depends") or [])
        if dep in tareas_por_id
    )
    terminos["bloqueada"] = PESO_BLOQUEADA * (1.0 if bloqueada else 0.0)

    if tarea.get("due"):
        dias = (dia(tarea["due"]) - hoy).days
        proximidad = min(max((30 - dias) / 30, 0.0), 1.0)
    else:
        proximidad = 0.0
    terminos["proximidad"] = PESO_PROXIMIDAD * proximidad

    criterios = 1.0 if (tarea.get("acceptance_criteria") or []) else 0.0
    terminos["criterios"] = PESO_CRITERIOS * criterios

    if tarea.get("created_at"):
        edad_dias = (hoy - dia(tarea["created_at"])).days
    else:
        edad_dias = 0
    terminos["edad"] = PESO_EDAD * min(edad_dias / 30, 4.0)

    return round(sum(terminos.values()), 1), terminos


def clave_de_orden(fila):
    urgencia, tarea = fila
    tiene_ordinal = tarea.get("ordinal") is not None
    numero = int(tarea["id"].split("-")[1])
    return (
        0 if tiene_ordinal else 1,
        tarea.get("ordinal") if tiene_ordinal else 0,
        -urgencia,
        numero,
    )


def main():
    datos = yaml.safe_load(TABLERO.read_text(encoding="utf-8"))
    tablero = datos["board"]
    hoy = dia(tablero["hoy"])
    tareas_por_id = {t["id"]: t for t in datos["tasks"]}

    filas = []
    for tarea in datos["tasks"]:
        urgencia, terminos = calcular(tarea, tablero, tareas_por_id, hoy)
        filas.append((urgencia, tarea, terminos))

    filas.sort(key=lambda f: clave_de_orden((f[0], f[1])))

    print("hoy = %s, lease_minutes = %s" % (tablero["hoy"], tablero.get("lease_minutes")))
    print()
    print("%-9s %7s  %-12s %s" % ("id", "urgency", "status", "desglose"))
    for urgencia, tarea, terminos in filas:
        if terminos.get("terminal"):
            desglose = "estado terminal, urgencia 0.0 sin calcular nada mas"
        else:
            desglose = "  ".join(
                "%s %+.3f" % (nombre, valor)
                for nombre, valor in terminos.items()
                if valor
            )
        print("%-9s %7.1f  %-12s %s" % (tarea["id"], urgencia, tarea["status"], desglose))
    return 0


if __name__ == "__main__":
    sys.exit(main())
