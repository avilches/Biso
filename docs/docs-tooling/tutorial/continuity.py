#!/usr/bin/env python3
"""Check that the board counters chain up between consecutive tutorial scenarios.

Every fixture declares, in `board_in` and `board_out`, the board state it starts from and
the one it hands over. This script pulls the "To Do / In Progress / Done" counters out of that prose
and checks that what one scenario hands over is what the next one receives. A mismatch there means
two scenarios are telling different stories, and it is the kind of break no single writer can see
from inside their own file.

    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \
        python docs/docs-tooling/tutorial/continuity.py

Exits 0 when the chain holds and 1 when it does not.
"""

import pathlib
import re
import sys

import yaml

SCENARIOS_DIR = pathlib.Path(__file__).resolve().parent / "escenarios"

# Fixtures quote the counters in two shapes, both legitimate: the short "54/5/190" and the long
# "To Do 54 | In Progress 5 | Done 190", which is the one `biso prime` prints.
SHORT_FORM = re.compile(r"\b(\d+)\s*/\s*(\d+)\s*/\s*(\d+)\b")
LONG_FORM = re.compile(r"To Do\s+(\d+)\s*\|\s*In Progress\s+(\d+)\s*\|\s*Done\s+(\d+)", re.I)


def counters(text):
    """Return the last counter triple quoted in the text, which is the one that counts."""
    text = text or ""
    found = [(m.end(), m.groups()) for m in SHORT_FORM.finditer(text)]
    found += [(m.end(), m.groups()) for m in LONG_FORM.finditer(text)]
    if not found:
        return None
    found.sort()
    return tuple(int(n) for n in found[-1][1])


def show(triple):
    return "%d/%d/%d" % triple if triple else "(not declared)"


def main():
    paths = sorted(SCENARIOS_DIR.glob("*.yaml"))
    problems = []
    previous = None

    print("%-38s %-14s %-14s" % ("scenario", "receives", "hands over"))
    for path in paths:
        data = yaml.safe_load(path.read_text(encoding="utf-8"))
        received = counters(data.get("board_in"))
        handed_over = counters(data.get("board_out"))
        print("%-38s %-14s %-14s" % (path.stem, show(received), show(handed_over)))

        if received is None or handed_over is None:
            problems.append("%s does not declare the counters in one of the two fields" % path.name)
        elif previous and previous[1] != received:
            problems.append(
                "%s receives %s but %s handed over %s"
                % (path.name, show(received), previous[0], show(previous[1]))
            )
        if handed_over:
            previous = (path.name, handed_over)

    print()
    if problems:
        for problem in problems:
            print("error: %s" % problem, file=sys.stderr)
        return 1
    print("The counter chain holds across all %d scenarios." % len(paths))
    return 0


if __name__ == "__main__":
    sys.exit(main())
