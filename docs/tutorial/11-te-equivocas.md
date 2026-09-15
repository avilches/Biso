<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 11. You get it wrong

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

Ten chapters in, you've been using `biso` without thinking much about it: you ask for a list,
you ask for a task, and it gives it to you. It's easy to forget that behind every query there's
a rule deciding what happens when what you're asking for doesn't exist. Today brings three
stumbles in a row, on purpose, because they teach the single most important rule in the whole
specification: a value the board doesn't know is an error, never an empty list. None of the
three writes anything: this is a reading chapter, not a writing one, so the board comes out
exactly as it went in.

!!! abstract "What this scenario teaches"
    - A value the board doesn't know is an error, whether you're reading or writing, and never an empty list. That's why an empty list IS a fact about the board, and whoever gets one can act on it.
    - Vocabulary matching ignores case, spaces, hyphens and underscores, so `TO_DO`, `to do` and `To-Do` aren't three similar filters: they're the same filter.
    - A text reference that matches several tasks doesn't guess which one you meant: it lists the candidates and lets you choose.
    - A well-formed reference the board never got around to assigning is a different error from one for a task that did exist and is gone now, even though the two messages start the same way.
    - Each stumble has its own exit code, so an agent can branch on the number without reading the message.

You want to see what's still pending and type the status name the way you'd call it in any
other task tracker.

```console
$ biso ls -s Pending
error: unknown status: "Pending"
       valid statuses on this board: To Do, In Progress, Done
```

Exit code: `3`

*Note: There's no «Pending» on this board, so the program doesn't hand you an empty list pretending it understood: it fails with code 3 (BAD_VALUE) and tells you what the valid values are, so you don't have to guess the exact name. It's the same message, word for word, whether the impossible value shows up in a `biso ls` filter or in a `biso set` write, because the same matching rule holds the same whether you're reading or writing.*

You try a value that does exist, but typed however it comes out: in uppercase, with an
underscore.

```console
$ biso ls --count -s TO_DO
54
```

Exit code: `0`

*(derived output, see [El algoritmo de coincidencia](../spec/vocabularios.md#el-algoritmo-de-coincidencia); not literal spec text)*

*Note: `TO_DO` doesn't appear as-is in the board's configuration, which stores `To Do`. But the matching algorithm lowercases both values and strips spaces, hyphens and underscores before comparing, so `TO_DO` normalizes to `todo`, same as `To Do`. The board hasn't changed since the previous chapter ended: there are still 54 tasks in `To Do`. `--count` prints just that number, with no other line.*

You try again, now with a hyphen and mixed case.

```console
$ biso ls --count -s To-Do
54
```

Exit code: `0`

*(derived output, see [El algoritmo de coincidencia](../spec/vocabularios.md#el-algoritmo-de-coincidencia); not literal spec text)*

*Note: Same result, because it's the same filter: `To-Do` also normalizes to `todo`. The same rule strips every space too, not just repeated ones, so `to do`, with two spaces in the middle, would have worked just as well (it's the literal example from the matching algorithm's own table). Four different ways to type it, one single value.*

You've forgotten the number of the retry task, so you type the first thing you remember
from its title.

```console
$ biso get "retry"
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-19  Done         task  high    Retry the upload on 5xx                           ac 6/6  @claude  -
```

Exit code: `5`

*(derived output, see [La búsqueda por texto](../spec/referencias.md#la-búsqueda-por-texto); not literal spec text)*

*Note: All three tasks have «retry» in the title, so none of them wins by having it in the title while the others only have it in the body (the text search rule): all three are candidates equally. The program doesn't choose for you. Exit code 5 (AMBIGUOUS), and the three rows go to stdout, in the same column format as `biso ls`, sorted by urgency because none of them carries an `ordinal`: TASK-60 has an open question, so it doesn't add the active-task term, but it's high priority; TASK-33 is medium priority; TASK-19 is already done, so its urgency is 0.0 (a task in a terminal state always has urgency 0.0) and it comes last. TASK-19's `ac 6/6` carries straight over from how chapter 10 closes it: the two original criteria plus the four it adds there, all checked.*

You type the number by hand, but an extra digit sneaks in.

```console
$ biso get TASK-99
error: TASK-99 has never existed on this board
note: the highest id ever assigned here is TASK-62
```

Exit code: `4`

*(derived output, see [Los tres mensajes de "no la encuentro"](../spec/referencias.md#los-tres-mensajes-de-no-la-encuentro); not literal spec text)*

*Note: `TASK-99` is well-formed as an identifier, so this isn't a usage error (code 2): it's an error because that identifier has never existed, code 4 (NOT_FOUND), with the `never_allocated` key. The text is the literal message for this case, with the highest-ever identifier swapped for this board's own, `TASK-62`, the task created back in chapter 2. There's a third «can't find it» message in the same family, for a task the board did get around to assigning and that's gone now (archived and then deleted): it shares code 4 but not the key, and there's no need to reproduce it here for the difference between the two to be clear. And there's a fourth stumble we haven't even touched, a malformed identifier like `TASK-1.1`, which is code 2 instead of 4, because there the problem isn't that the task doesn't exist but that what you typed doesn't have the shape of a reference. Four ways to get it wrong, four different codes: 2, 3, 4 and 5.*
