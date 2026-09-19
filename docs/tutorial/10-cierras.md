<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 10. You close it

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

The retry works, it's documented, and someone has given it a quick look over in the comments.
Before closing TASK-19 you go back over the acceptance criteria, and realize that several
things you've already done were never written down as a criterion: the backoff, the test for the
success case, the endpoint documentation and the log. You add them before checking anything off,
so the final record tells the whole story.

!!! abstract "What this scenario teaches"
    - The `--check-ac` and `--uncheck-ac` flags accept five selector forms, all repeatable and all combinable in the same call: a single key, a range, a comma-separated list, `all`, or the criterion's text.
    - Checking or unchecking is always safe to repeat. `--check-ac all` breaks nothing even if everything was already checked, and that's why closing out a cycle can always use it without first checking what's missing.
    - Closing a task always clears its lease, no matter who's holding it. It's the invariant that governs any write that takes a task out of the active state, and `biso finish` applies it without exception.

You add the four criteria you were missing, all at once.

```console
$ biso set TASK-19 --add-ac "The backoff follows the same pattern as the download endpoint" --add-ac "There's a test that covers the success case after retrying" --add-ac "The endpoint documentation mentions the retry limit" --add-ac "The number of attempts shows up in the log"
TASK-19  In Progress  ac 0/6  urgency 11.0  added ac #9, #10, #11, #12
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [Los criterios y sus claves estables](../spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables); not literal spec text)*

*Note: Four `--add-ac` flags in the same call add four new elements, in the order they're written, with new keys, #9, #10, #11 and #12, after the #7 and #8 the task was left with at the end of scenario 7. The keys are that high because that scenario burned #1 through #6 rewriting the list, and a key is never reused.*

You check off the documentation one, which you finished days ago.

```console
$ biso set TASK-19 --check-ac 8
TASK-19  In Progress  ac 1/6  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

The three you closed this morning, one after another, with a range.

```console
$ biso set TASK-19 --check-ac 9-11
TASK-19  In Progress  ac 4/6  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

And the two that aren't consecutive, with a comma-separated list.

```console
$ biso set TASK-19 --check-ac 7,12
TASK-19  In Progress  ac 6/6  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

Almost right away you realize you marked those last two too soon: the retry still gives up
one attempt earlier than the configured limit, and the number of attempts doesn't show up in
the log, only the final result. You uncheck both.

```console
$ biso set TASK-19 --uncheck-ac 7,12
TASK-19  In Progress  ac 4/6  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

*Note: `--uncheck-ac` takes the same selector as `--check-ac`, with the same effect inverted, here the comma-separated list, just like the call right before it.*

You fix the off-by-one in the limit, and check that criterion off by searching for its text
instead of by number.

```console
$ biso set TASK-19 --check-ac "retry happens"
TASK-19  In Progress  ac 5/6  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

*Note: "retry happens" doesn't match the key grammar (^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$), so it's treated as literal text and matches the element whose text contains it. Only #7 ("The retry happens automatically on a 5xx, up to a configurable limit") matches: #11 mentions the retry limit and #10 mentions retrying, but neither contains that exact text.*

You add the attempt count to the log, check that it now shows up, and close the task. You
check off anything left, just in case you missed something, and write the summary.

```console
$ biso finish TASK-19 --check-ac all --append-summary "Retry with the same backoff as the download endpoint, capped at 3 attempts, attempt count visible in the log."
TASK-19  Done  ac 6/6  urgency 0.0
```

Exit code: `0`

*(derived output, see [`biso set`](../spec/cmd/set.md), [`biso finish`](../spec/cmd/verbos-del-ciclo.md#biso-finish), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: `--check-ac all` checks #12, the only one left, and doesn't touch the other five, they were already checked, and checking the same criterion twice "stays checked, with no warning, the operation is idempotent." Urgency drops to 0.0 because the terminal state fixes it there by definition, not because it gets recalculated with the usual formula. What this line doesn't say, and is worth knowing, is that closing a task always clears `leaseExpiresAt` and `leaseHolder`, whoever is holding it. TASK-19 had been carrying @claude's lease since scenario 5; after this `biso finish` it's left without a lease, and it would end up just as empty even if another identity had been holding it. The invariant always wins over the general rule that "a write by someone else doesn't touch the lease": a finished task with a live lease is a state that not even importing the board would accept.*
