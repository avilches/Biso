<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 8. You get stuck and someone has to decide

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

You're still on `TASK-19`, and you reach a point where you can't go on without someone deciding
something: how many retries at most should be allowed before treating the 5xx as a definitive
failure. You could just make up a number, but that's exactly what rule 11 of `biso`'s startup
message asks you not to do: "Ask instead of guessing." Before parking your question, take a
look at what one looks like from the outside: `TASK-60` already has one open, and this is how it
shows up in the `NEEDS ANSWER` block of the startup message, exactly as `biso prime` prints it:

```
NEEDS ANSWER
  TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
    Should the retry budget be shared with the download endpoint or kept separate?
```

The task doesn't change state just for having an open question: it still counts as in progress
for its record, but it leaves the `IN PROGRESS` block of the startup message and shows up here
instead, so nobody mistakes it for forgotten, or for active at the same time.

!!! abstract "What this scenario teaches"
    - Parking and unparking are their own commands, `biso ask` and `biso answer`, and neither one changes the task's state.
    - A parked task shows up from the outside in the `NEEDS ANSWER` block of the startup message and not in `IN PROGRESS`, even though its state is still the active one.
    - Asking is better than guessing. It's rule 11 of the `biso prime` message, and it's the whole reason `biso ask` exists.
    - The `biso answer` command turns the question into two comments in the same write: one with the original author and moment of the question, and another with the answer, signed now.

You look closely at TASK-60's question, not just the trimmed line from the startup message.

```console
$ biso get TASK-60 --section question
TASK-60  Confirm the retry budget for the upload endpoint

## Open Question
@claude, 2026-09-06 09:30
Should the retry budget be shared with the download endpoint or kept separate?
```

Exit code: `0`

Now yours. Instead of picking a number out of thin air, you leave it written down as a
question.

```console
$ biso ask TASK-19 "What is the maximum number of retries before treating the 5xx as a definitive failure? Is there already a reference value somewhere else in the code, or does a new one need to be set?"
TASK-19  In Progress  ac 0/2  urgency 7.0
```

Exit code: `0`

*(derived output, see [`biso ask`](../spec/cmd/verbos-del-ciclo.md#biso-ask), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: The state doesn't change, it's still In Progress. Urgency does drop, from 11.0 to 7.0: the formula's activity term requires the task to be in the active state *and* to have no open question, so as soon as `waiting` becomes true that term stops adding its 4.0 points. It's the same 4.0 drop that the specification's own example shows for TASK-11 (from 19.0 to 15.0), here with TASK-19's numbers.*

You check that it got written correctly.

```console
$ biso get TASK-19 --section question
TASK-19  Retry the upload on 5xx

## Open Question
@claude, 2026-09-06 14:10
What is the maximum number of retries before treating the 5xx as a definitive failure? Is there already a reference value somewhere else in the code, or does a new one need to be set?
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md), [La pregunta abierta](../spec/modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta); not literal spec text)*

A while later, looking at a similar endpoint, you find the answer yourself: there's already
a value fixed somewhere else in the code. You answer, and the task gets unparked.

```console
$ biso answer TASK-19 "There is already a reference value: the download endpoint uses 3 retries with exponential backoff in retry.go. Use the same value and the same logic here."
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso answer`](../spec/cmd/verbos-del-ciclo.md#biso-answer), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: Urgency is back to its earlier 11.0: waiting is false again, and the activity term goes back to adding its 4.0 points.*

You look at the comments to see how the trail turned out.

```console
$ biso get TASK-19 --section comments
TASK-19  Retry the upload on 5xx

## Comments
#1  @claude, 2026-09-06 14:10
What is the maximum number of retries before treating the 5xx as a definitive failure? Is there already a reference value somewhere else in the code, or does a new one need to be set?

#2  @claude, 2026-09-06 15:30
There is already a reference value: the download endpoint uses 3 retries with exponential backoff in retry.go. Use the same value and the same logic here.
```

Exit code: `0`

*(derived output, see [Los comentarios](../spec/modelo-de-datos/comentarios.md#los-comentarios), [Las fechas](../spec/modelo-de-datos/fechas.md#las-fechas), [`biso get`](../spec/cmd/get.md), [`biso answer`](../spec/cmd/verbos-del-ciclo.md#biso-answer); not literal spec text)*

*Note: The first comment carries the moment the question was asked (14:10), not the moment it was answered (15:30): that's the one documented exception to the rule, the question gets carried over exactly as the program first recorded it. The two comments appear in the order `biso answer` wrote them, which is also their insertion order, and that's the order that gets stored and shown, not the order their timestamps would suggest.*
