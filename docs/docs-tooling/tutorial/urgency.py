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
"""

import datetime
import pathlib
import sys

import yaml

HERE = pathlib.Path(__file__).resolve().parent
BOARD_PATH = HERE / "tablero.yaml"

# Coefficients from SPEC 5.4. A real board can override them through the seven urgency.* keys of
# SPEC 10.10; these are the defaults, which are what the example board uses.
WEIGHT_PRIORITY = 6.0
WEIGHT_ACTIVE = 4.0
WEIGHT_BLOCKING = 8.0
WEIGHT_BLOCKED = -5.0
WEIGHT_DUE = 12.0
WEIGHT_CRITERIA = 1.0
WEIGHT_AGE = 0.5

PRIORITY_VALUE = {"high": 1.0, "medium": 0.5, "low": 0.0, None: 0.3}


def as_date(value):
    return datetime.date.fromisoformat(str(value)[:10])


def role(board, name):
    return next(entry["name"] for entry in board["statuses"] if entry.get("role") == name)


def urgency_of(task, board, tasks_by_id, today):
    """Return the rounded urgency and the breakdown of the terms that produced it."""
    active_status = role(board, "active")
    terminal_status = role(board, "terminal")

    if task["status"] == terminal_status:
        return 0.0, {"terminal": True}

    terms = {}

    terms["priority"] = WEIGHT_PRIORITY * PRIORITY_VALUE[task.get("priority")]

    # The active term only counts when the task sits in the active status AND has no open question:
    # a parked task is not being worked on by anybody.
    has_open_question = bool(task.get("question"))
    is_active = task["status"] == active_status and not has_open_question
    terms["active"] = WEIGHT_ACTIVE * (1.0 if is_active else 0.0)

    # "blocking" holds when some unfinished task depends on this one.
    blocking = any(
        task["id"] in (other.get("depends") or []) and other["status"] != terminal_status
        for other in tasks_by_id.values()
    )
    terms["blocking"] = WEIGHT_BLOCKING * (1.0 if blocking else 0.0)

    # "blocked" holds when this one depends on some unfinished task.
    blocked = any(
        tasks_by_id[dep]["status"] != terminal_status
        for dep in (task.get("depends") or [])
        if dep in tasks_by_id
    )
    terms["blocked"] = WEIGHT_BLOCKED * (1.0 if blocked else 0.0)

    if task.get("due"):
        days_left = (as_date(task["due"]) - today).days
        # An overdue task has negative days_left, so the ratio goes past 1 and is clamped: being
        # overdue never scores higher than being due today (SPEC 5.4).
        closeness = min(max((30 - days_left) / 30, 0.0), 1.0)
    else:
        closeness = 0.0
    terms["due"] = WEIGHT_DUE * closeness

    terms["criteria"] = WEIGHT_CRITERIA * (1.0 if (task.get("acceptance_criteria") or []) else 0.0)

    age_days = (today - as_date(task["created_at"])).days if task.get("created_at") else 0
    terms["age"] = WEIGHT_AGE * min(age_days / 30, 4.0)

    return round(sum(terms.values()), 1), terms


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
                "%s %+.3f" % (name, value) for name, value in terms.items() if value
            )
        print("%-9s %7.1f  %-12s %s" % (task["id"], urgency, task["status"], breakdown))
    return 0


if __name__ == "__main__":
    sys.exit(main())
