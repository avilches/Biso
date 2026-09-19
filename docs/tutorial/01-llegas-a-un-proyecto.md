<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 1. You land on a project you don't know

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

You've just been dropped into a project you've never touched before. People have been
working on it for a while, there are tasks half done, and you don't even know where to
start. Before touching anything, you want to know what's there: what's being worked on
right now, what's waiting for an answer, what's yours, and what rules govern this board
specifically. You don't want to read an entire manual just for that.

!!! abstract "What this scenario teaches"
    - You read the startup message once, at the start of the session, and that's enough: you never need to look anything else up to complete a full work cycle.
    - `biso where` exists for the day you're not sure which board you're actually touching.

The first thing you do, before looking at anything else, is this:

```console
$ biso prime
biso 1.0.0 - the task board of this project. This message is all you need to start.

BOARD  Kex
  To Do 54 | In Progress 4 | Done 190
  new tasks start in To Do; `biso start` moves to In Progress; `biso finish` to Done
  types       idea, memory, task, bug, docs
  priorities  high, medium, low
  you are     @claude

COMMANDS  (`biso help <cmd>...` for the detail of any, several at once)
  biso ls [-s STATUS] [--type T] [-l LABEL] [--mine] [--search TEXT]
  biso get <ref> [--section ac]
  biso new "TITLE" [-d TEXT] [--add-ac TEXT]... [--type T] [--priority P]
  biso start <ref>... [--append-plan TEXT]
  biso note <ref> "TEXT"
  biso ask <ref> "QUESTION"
  biso answer <ref> "TEXT"
  biso finish <ref>... [--append-summary "TEXT"] [--check-ac all]
  biso set <ref>... [any field flag]
  biso comment <ref> "TEXT" [--comment-author @who]

FIELD FLAGS  (same names, same meaning, in every command above that writes)
  -t --title  -s --status  --type --clear-type  --priority --clear-priority
  -p --parent --clear-parent  --due --clear-due  --ordinal --clear-ordinal  --author --clear-author
  -l --add-labels --rm-labels --clear-labels --replace-labels
  -a --add-assignees --rm-assignees --clear-assignees --replace-assignees
  --add-refs --rm-refs --clear-refs --replace-refs
  --add-docs --rm-docs --clear-docs --replace-docs
  --add-deps --rm-deps --clear-deps --replace-deps
  --add-files --rm-files --clear-files --replace-files
  --add-ac --rm-ac --clear-acs   --check-ac --uncheck-ac
  -d --append-desc --clear-desc  --append-plan --clear-plan
  --append-note --clear-notes  --append-summary --clear-summary
  --comment --rm-comment --set-comment-date
  --ext K=V --rm-ext --clear-ext

RULES  (none of these are guessable; they are the whole learning curve)
  1. Every write goes through biso. Nothing else touches the board.
  2. <ref> is an id (TASK-12), a bare number (12) or free text ("CRLF"). Text
     matching several tasks is an error that lists them, never a guess. `note`,
     `comment`, `ask` and `answer` take one <ref>; `set`, `start` and `finish`
     take several.
  3. Filters reject values this board does not have: `-s Pending` is an error,
     not an empty list. Case, spaces, hyphens and underscores are ignored, so
     `-s todo`, `-s "To Do"` and `-s TO_DO` are one and the same filter. An
     empty list is therefore a fact about the board that you can act on.
  4. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says
     on stderr what it left out. --all lifts the limit, --any-status includes
     Done, --archived reaches the archive.
  5. --check-ac and --uncheck-ac take all, 3, 1-4, 1,3,7 or the criterion text. The
     numbers are the stable #N keys that `biso get` shows, and they never shift
     when one criterion is removed.
  6. `biso new` prints the new id and nothing else. Every other write prints one
     line per task: id, status, criteria, urgency. Add --print for the whole
     record, or --json for a versioned envelope.
  7. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.
  8. Long text: a real newline works, and so do -d @file.md and -d - for stdin.
  9. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,
     6 precondition not met, 7 nothing written, 8 environment, 20 no board here.
 10. `biso ask <ref> "..."` parks a task on a question and `biso answer` unparks
     it, writing both into the comments. Ask instead of guessing. A task
     assigned to you is one a person decided you should do.

IN PROGRESS
  TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
  TASK-52  In Progress  task  low     Document the release checklist                    ac 0/1  -        -
    lease expired 2026-09-05T09:00:00Z, was held by @bob
  TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -

NEEDS ANSWER
  TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
    Should the retry budget be shared with the download endpoint or kept separate?

ASSIGNED TO YOU
  TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
  TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -

NEXT UP  (not assigned to you, by urgency)
  TASK-7   To Do        bug   high    Crash on an empty repository                      ac 0/4  -        2026-09-08
  TASK-19  To Do        task  high    Retry the upload on 5xx                           ac 0/2  -        -
  TASK-44  To Do        bug   low     Wrong column width on narrow ttys                 ac 0/1  -        -
  49 more not shown: `biso ls --not-active --not-waiting`

Pick one, `biso start <ref> --append-plan "..."`, work, `biso note <ref> "..."` as you go,
and close with `biso finish <ref> --check-ac all --append-summary "..."`.
That is the loop. Create a task when the work needs planning or review; do small
edits directly.
```

Exit code: `0`

*(derived output, see [La salida literal](../spec/cmd/prime.md#la-salida-literal); not literal spec text)*

*Note: This is the whole message, top to bottom, and there is no second screen after it. Every fixed part of it, the command list, the field flags and the rules, is copied from the specification; what changes here is the board summary, which is this tutorial's board, and the example id in rule 2, written with this board's prefix (see tutorial/lagunas/01-03.md). The BOARD block states the actual vocabulary of this board (the types, the priorities, the three states and their role) and who you are here. The four sections below split up the entire board without any task showing up in two places: what's in progress, what's waiting on an answer from you, what's yours, and what's next by urgency. No one writes that urgency by hand, it's recalculated every time you ask for it (you'll come back to this in scenario 3, "What now?"). With this alone you could already create a task, start it, work on it and close it without opening any other document: that's the standard this message is written to.*

Even so, there's one doubt `prime` doesn't answer: this board, the one you just read,
is it exactly the one you think it is? If you work with several copies of the project at
once (a different worktree, a folder cloned twice), it's worth checking before you write
anything.

```console
$ biso where
id       3f9a2b1c
board    Kex
path     /Users/avilches/.biso/boards/kex-3f9a2b1c
source   project pointer at /Users/avilches/Hub/Projects/Kex
me       @claude
tasks    248 not archived, 31 archived, highest id ever assigned TASK-290
```

Exit code: `0`

*(derived output, see [`biso where`](../spec/cmd/where.md); not literal spec text)*

*Note: Four pieces of data, each changing on its own: the id never changes, the name changes with `biso config set project_name`, the path changes if someone moves the folder by hand, and `source` says where that answer came from, which here is the project pointer (how the board gets resolved in the first place). If there were no board at all, both `prime` and `where` would fail with exit code 20, and `where`'s own error message points to `biso init` as the fix: it creates a new board and leaves it pointed at from this project. All you need right now is to know it exists; its flags matter only when you actually need to create one.*
