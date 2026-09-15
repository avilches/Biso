#!/usr/bin/env python3
"""Compute the urgency of every task on the tutorial's example board, per SPEC section 5.4.

This exists so that no fixture has to do that arithmetic by hand. The tutorial's scenarios order
lists by urgency, and that order has to be reproducible instead of taken on trust from a number
somebody typed. Run it with:

    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \
        python docs/docs-tooling/tutorial/urgency.py

It prints every task with its urgency and the breakdown of each term, already sorted by the default
ordering rule of `biso ls` (SPEC 10.4): tasks with an ordinal first, then the rest by descending
urgency, with any tie broken by ascending identifier.

All arithmetic is done with exact fractions, never with floats, so that the final rounding to one
decimal is exact too: a float sum of these terms can land a hair below or above a true .x5 tie
depending on binary representation, and rounding that approximation would silently pick the wrong
side. See SPEC modelo-de-datos/urgencia.md for the rounding rule (round half away from zero) and for
why the priority weight is derived from position and not from a name.
"""

import datetime
import pathlib
import sys
from fractions import Fraction

import yaml

HERE = pathlib.Path(__file__).resolve().parent
BOARD_PATH = HERE / "tablero.yaml"

# Coefficients from SPEC 5.4. A real board can override them through the seven urgency.* keys of
# SPEC 10.10; these are the defaults, which are what the example board uses.
WEIGHT_PRIORITY = Fraction(6)
WEIGHT_ACTIVE = Fraction(4)
WEIGHT_BLOCKING = Fraction(8)
WEIGHT_BLOCKED = Fraction(-5)
WEIGHT_DUE = Fraction(12)
WEIGHT_CRITERIA = Fraction(1)
WEIGHT_AGE = Fraction(1, 2)

NO_PRIORITY_WEIGHT = Fraction(3, 10)


def priority_weight(priority_name, priorities):
    """Weight of a priority level, per SPEC modelo-de-datos/urgencia.md: derived from its position
    in the board's configured `priorities`, from the most urgent (index 0) to the least urgent,
    never from its name. A task with no priority assigned always scores 0.3, regardless of the
    vocabulary or how many levels it has.
    """
    if priority_name is None:
        return NO_PRIORITY_WEIGHT
    n = len(priorities)
    if n == 1:
        return Fraction(1)
    i = priorities.index(priority_name)
    return Fraction(n - 1 - i, n - 1)


def round_half_away_from_zero(value: Fraction, ndigits: int = 1) -> Fraction:
    """Round an exact Fraction to `ndigits` decimals, half away from zero: an exact tie rounds away
    from zero (1.45 -> 1.5, -1.45 -> -1.5), never toward the nearest even digit. Exact because the
    input is a Fraction and not a float, so there is no representation error to break the tie.
    """
    scale = Fraction(10) ** ndigits
    scaled = value * scale
    half = Fraction(1, 2)
    if scaled >= 0:
        rounded = (scaled + half).__floor__()
    else:
        rounded = -((-scaled + half).__floor__())
    return Fraction(rounded, 1) / scale


def as_date(value):
    return datetime.date.fromisoformat(str(value)[:10])


def role(board, name):
    return next(entry["name"] for entry in board["statuses"] if entry.get("role") == name)


def urgency_of(task, board, tasks_by_id, today):
    """Return the rounded urgency (a float, for printing) and the breakdown of the exact terms that
    produced it.
    """
    active_status = role(board, "active")
    terminal_status = role(board, "terminal")

    if task["status"] == terminal_status:
        return 0.0, {"terminal": True}

    terms = {}

    terms["priority"] = WEIGHT_PRIORITY * priority_weight(task.get("priority"), board["priorities"])

    # The active term only counts when the task sits in the active status AND has no open question:
    # a parked task is not being worked on by anybody.
    has_open_question = bool(task.get("question"))
    is_active = task["status"] == active_status and not has_open_question
    terms["active"] = WEIGHT_ACTIVE * (Fraction(1) if is_active else Fraction(0))

    # "blocking" holds when some unfinished task depends on this one. An archived task that never
    # finished counts as finished here, same as a task in the terminal status (SPEC
    # modelo-de-datos/urgencia.md): archiving it unblocks whoever depended on it.
    blocking = any(
        task["id"] in (other.get("depends") or [])
        and other["status"] != terminal_status
        and not other.get("archived")
        for other in tasks_by_id.values()
    )
    terms["blocking"] = WEIGHT_BLOCKING * (Fraction(1) if blocking else Fraction(0))

    # "blocked" holds when this one depends on some unfinished, non-archived task.
    blocked = any(
        tasks_by_id[dep]["status"] != terminal_status and not tasks_by_id[dep].get("archived")
        for dep in (task.get("depends") or [])
        if dep in tasks_by_id
    )
    terms["blocked"] = WEIGHT_BLOCKED * (Fraction(1) if blocked else Fraction(0))

    if task.get("due"):
        days_left = (as_date(task["due"]) - today).days
        # An overdue task has negative days_left, so the ratio goes past 1 and is clamped: being
        # overdue never scores higher than being due today (SPEC 5.4).
        closeness = max(Fraction(0), min(Fraction(30 - days_left, 30), Fraction(1)))
    else:
        closeness = Fraction(0)
    terms["due"] = WEIGHT_DUE * closeness

    has_criteria = Fraction(1) if (task.get("acceptance_criteria") or []) else Fraction(0)
    terms["criteria"] = WEIGHT_CRITERIA * has_criteria

    # age_days is an integer: the difference between the calendar date of "today" and the calendar
    # date of createdAt, never the time of day (SPEC modelo-de-datos/urgencia.md).
    age_days = (today - as_date(task["created_at"])).days if task.get("created_at") else 0
    terms["age"] = WEIGHT_AGE * min(Fraction(age_days, 30), Fraction(4))

    total = sum(terms.values(), Fraction(0))
    return float(round_half_away_from_zero(total, 1)), terms


def sort_key(urgency, task):
    """The default ordering tuple of SPEC 10.4."""
    ordinal = task.get("ordinal")
    return (
        0 if ordinal is not None else 1,
        ordinal if ordinal is not None else 0,
        -urgency,
        int(task["id"].split("-")[1]),
    )


def main():
    data = yaml.safe_load(BOARD_PATH.read_text(encoding="utf-8"))
    board = data["board"]
    today = as_date(board["hoy"])
    tasks_by_id = {task["id"]: task for task in data["tasks"]}

    rows = []
    for task in data["tasks"]:
        urgency, terms = urgency_of(task, board, tasks_by_id, today)
        rows.append((urgency, task, terms))
    rows.sort(key=lambda row: sort_key(row[0], row[1]))

    print("today = %s, lease_minutes = %s" % (board["hoy"], board.get("lease_minutes")))
    print()
    print("%-9s %7s  %-12s %s" % ("id", "urgency", "status", "breakdown"))
    for urgency, task, terms in rows:
        if terms.get("terminal"):
            breakdown = "terminal status, urgency 0.0 with nothing else computed"
        else:
            breakdown = "  ".join(
                "%s %+.3f" % (name, float(value)) for name, value in terms.items() if value
            )
        print("%-9s %7.1f  %-12s %s" % (task["id"], urgency, task["status"], breakdown))
    return 0


if __name__ == "__main__":
    sys.exit(main())
