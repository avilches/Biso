<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 13. You touch twenty at once

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

Last stop. So far everything you've done has been on one task at a time, and that's how you
work ninety percent of the time.

But the day comes when you need to touch twenty at once. The team decides documentation tasks
drop in priority this quarter, or a list of incidents arrives from another system and all of
them need to go in. And that's where a scary question shows up: if something fails halfway
through the batch, what ends up written?

`biso`'s answer is nothing. It's worth seeing why.

!!! abstract "What this scenario teaches"
    - Anything you can do once you can do a hundred times, with the same flag names. There's no separate syntax for batches.
    - Validation happens upfront and in full: the whole batch gets checked before anything is written. A batch never ends up half-done, so you never have to figure out where it broke.
    - `--dry-run` validates without writing and answers with the exit code: 0 if it would have worked, 9 if not. It's how you look before you leap.

The agent starts with the easy part: two tasks that need their priority lowered, in a
single call.

```console
$ biso set TASK-44 TASK-61 --priority low
TASK-44  To Do  ac 0/1  urgency 1.6
TASK-61  To Do  ac 0/1  urgency 1.4
```

Exit code: `0`

*(derived output, see [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: One line per task, in the same format as when you touch a single one. There's no batch mode that changes the shape of the output, and no summary that replaces the detail: what you learned with one applies to two just the same. `TASK-44` was already low priority, so its urgency doesn't move. `TASK-61` drops from 4.5 to 1.4, and there's a hidden lesson in there about reading these numbers: the real drop is exactly 3.0, what medium priority was contributing, but the two figures you see are rounded to one decimal from 4.45 and 1.45, and those two roundings don't go the same direction. Urgency is for ordering a list, not for doing arithmetic on it.*

Now a longer list, with a typo hidden inside: one of the references doesn't exist.

```console
$ biso set TASK-7 TASK-33 TASK-99 --priority medium
error: TASK-99 has never existed on this board
note: the highest id ever assigned here is TASK-62
```

Exit code: `4`

*(derived output, see [Los tres mensajes de "no la encuentro"](../spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), [`biso set`](../spec/cmd/set.md); not literal spec text)*

*Note: And here's what matters: **`TASK-7` and `TASK-33` haven't been touched**. It's not that they got written and then undone: it's that nothing ever got written, because validating all three references happens before the first write. Think about what that saves you. If the batch applied task by task and stopped on the first failure, after this error you'd have to figure out how many had already been written so you wouldn't repeat them, and there's no way to do that reliably from a program. The contract here is simple: all of them or none of them.*

It fixes the typo and this time, before writing, asks for it to be validated without
touching anything.

```console
$ biso set TASK-7 TASK-33 --priority medium --dry-run
2 tasks would be updated, nothing was written (--dry-run)
```

Exit code: `0`

*(derived output, see [Banderas globales](../spec/invocacion.md#banderas-globales), [`biso set`](../spec/cmd/set.md); not literal spec text)*

*Note: Exit code 0, so it would have worked. That code is the real answer, more than the text: a program calling `biso` doesn't need to read anything to know whether the batch is valid. The exact text of this line is a derivation. The specification gives the literal phrase for `biso new --from`'s batch ("242 tasks would be created, nothing was written (--dry-run)") and says `--dry-run` works on every command that writes, but it doesn't spell out the phrase for `biso set`. Here the same form was used with the verb swapped, which is the most likely choice, and it's noted in tutorial/lagunas/11-13.md.*

With validation green, it runs it for real.

```console
$ biso set TASK-7 TASK-33 --priority medium
TASK-7   To Do  ac 0/4  urgency 15.3
TASK-33  To Do  ac 1/3  urgency 4.3
```

Exit code: `0`

*(derived output, see [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: `TASK-7` drops from 18.3 to 15.3, the 3.0 it loses going from high to medium priority, and it's still the most urgent task on the board by a wide margin: it still has the deadline term, contributing 11.2 because it's due in two days. `TASK-33` doesn't move, since it was already medium priority, and it still shows up in the output anyway: it tells you where every task you named ends up, not just the ones that changed. Notice the alignment of the first column too. Its width is set by the longest identifier in the output, so `TASK-7` carries an extra space to line up with `TASK-33`.*

To finish, the other kind of batch that exists: bringing in new tasks from a file. The
agent got three incidents through another channel and has them in a file, one per line.
Before importing them, it validates them.

```console
$ biso new --from incidencias.ndjson --dry-run
3 tasks would be created, nothing was written (--dry-run)
```

Exit code: `0`

*Note: This sentence really is literal from the specification, just with a different number. And this is where `--dry-run` earns its keep most, because an import file can bring in a type, a status or a key this board doesn't know, and a value the board doesn't know is always an error. With an import batch that means a single misspelled field on line 200 stops the 199 before it from being created at all. That sounds harsh and it's exactly what you want: the alternative would be a half-imported board and no reliable way to tell where it stopped. If validation failed, the code would be 9, not 4 or 3: nine means exactly "nothing was written because validation didn't pass," and it's distinct from the error for one specific bad value precisely because what it's reporting on is the whole batch.*
