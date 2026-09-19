<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 6. Two at once on the same board

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

This is the scenario that teaches the most rules per line, and none of them are guesswork.

Sara and the agent work on the same board at the same time. Nothing stops them from writing to
the same task, and `biso` isn't going to stop them either: what it does is keep a record of who
said they were working on what, so the other one knows before duplicating the effort.

That's the lease, and up to now you've seen it come into being without looking at it head on.

!!! abstract "What this scenario teaches"
    - A lease is not a state, it's until when a reservation is good. A task can be in the active state with an expired lease, and that is not a contradiction.
    - Any write by whoever holds the lease renews it, including a write that changes no field. That is why there is no heartbeat command: working already is the heartbeat.
    - Writing to a task leased by another identity warns you, but doesn't stop anything. A flow block doesn't prevent duplicate work, it only pushes people to route around the tool.
    - An expired lease doesn't privilege anyone, and freeing it is an explicit claim made by `biso start`. No other command does it on its own, not even on a read.

Sara is reviewing the CRLF problem and finds something that looks useful for `TASK-11`,
which belongs to the agent and which he has under way right now. She jots it down without
stopping to think about whose lease it was.

```console
$ biso note TASK-11 "The CRLF also shows up in the test files, not just in the diff"
warning: TASK-11's lease is held by @claude until 2026-09-06T15:40:18Z
TASK-11  In Progress  ac 1/2  urgency 19.0
```

Exit code: `0`

*(derived output, see [Notas y avisos](../spec/salida-y-terminal.md#notas-y-avisos), [`biso start`](../spec/cmd/verbos-del-ciclo.md#biso-start), [El modelo de datos de una tarea](../spec/modelo-de-datos/index.md); not literal spec text)*

*Note: First thing: the note got written. The exit code is 0 and the task has changed. The warning is not a rejection, it's information. That's deliberate, and the specification argues for it: a flow block doesn't prevent duplicate work. If `biso` had told her "you can't write here", Sara would have written the note somewhere else, or edited the store by hand, and then the board would be lying. It's the same call as with unfinished dependencies and open questions: it warns, it doesn't block. Second thing, and it's easy to miss: this write **hasn't touched the lease**. It hasn't renewed it, because Sara isn't the one holding it, and it hasn't stolen it, because only `biso start` does that. It's still `@claude`'s and it still expires at the same time. The urgency is still 19.0 because a note doesn't change anything urgency measures. And that 19.0 breaks down as: 6.0 for high priority, 4.0 for being active, 8.0 because another unfinished task depends on it, and 1.0 for having acceptance criteria.*

The agent, in his own session, keeps going with his work and notes his own progress on the
same task.

```console
$ biso note TASK-11 "The normalizer now handles the test files case"
TASK-11  In Progress  ac 1/2  urgency 19.0
```

Exit code: `0`

*(derived output, see [`biso note`](../spec/cmd/verbos-del-ciclo.md#biso-note), [El modelo de datos de una tarea](../spec/modelo-de-datos/index.md); not literal spec text)*

*Note: No warning, because the lease is his. And something the output doesn't say: it just got renewed until 240 minutes after this moment. That's the central idea. There's no `biso heartbeat`, no `biso renew`, no flag to ask for one. Any write by the holder on their task renews the reservation, and "any" means all of them: the six cycle verbs, `biso set`, and `biso archive`. Working is what keeps the reservation alive, which is exactly what you want it to mean.*

The agent wants to confirm that the task's priority is still what he thinks it is. He
writes it again with the same value it already had.

```console
$ biso set TASK-11 --priority high
note: TASK-11 unchanged
TASK-11  In Progress  ac 1/2  urgency 19.0
```

Exit code: `0`

*(derived output, see [`biso set`](../spec/cmd/set.md), [El modelo de datos de una tarea](../spec/modelo-de-datos/index.md); not literal spec text)*

*Note: This step looks useless, and it's the one that closes the argument. It hasn't changed any field, the note says so, and yet **the lease has been renewed**. It had to work that way: if the heartbeat depended on values actually changing, a write that happens to match what was already there would let die the reservation of someone who really is working. The renewal looks at who is writing and to what, not whether they happened to change something. The other side of it: since no field on the task changed, the task's last-modified date is **not** touched. The reservation is renewed and the task stays clean.*

Now the other half of the matter. `TASK-52`, the one about documenting the release
checklist, has been in progress for days and was reserved by `@bob`, who hasn't shown up in
a while. The agent looks at what's currently in progress.

```console
$ biso ls -s "In Progress"
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-19  In Progress  task  high    Retry the upload on 5xx                           ac 0/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-52  In Progress  task  low     Document the release checklist                    ac 0/1  -        -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](../spec/cmd/ls.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: `TASK-52` is still `In Progress`, perfectly normally, and has nobody assigned. Its lease expired on September 5th and the stored state hasn't moved an inch. That's not an oversight, it's the rule: what expires is the claim, not the state. Nothing is going to pull it out of `In Progress` on its own, because doing so would mean a command touching tasks you didn't name, and because `biso prime`, which never writes, would end up showing a state that someone else's next write could change. The order of this list, in case you're wondering, is descending urgency and has nothing to do with state: 19.0, 11.0, 7.0, 5.4, and 3.1. `TASK-40` comes last, despite being in progress and medium priority, because it depends on `TASK-11` and being blocked subtracts 5.0.*

The agent decides to take on that abandoned task himself.

```console
$ biso start TASK-52
note: TASK-52 was already In Progress
TASK-52  In Progress  ac 0/1  urgency 5.4
```

Exit code: `0`

*(derived output, see [`biso start`](../spec/cmd/verbos-del-ciclo.md#biso-start), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: With a single call he's done three things: he's seen that the state was already active and says so with that note, he's been assigned the task because nobody had it, and **he's claimed the expired lease** in his own name. That claim happens inside the same transaction that checks it was still expired, and that detail matters: if two agents claim the same expired lease at the same instant, only one wins. There's no window where both believe it's theirs. What that check doesn't do is silence `@bob` if he ever comes back. He can still note, comment on, and close this task, because writing is never forbidden; what he can no longer do is renew or reclaim a reservation that now belongs to someone else, and if he wants it back he has to ask for it again with `biso start`, and he'll get the same warning as the first step. It is a deliberate difference from the classic fencing-token reservation pattern, and it's documented as an accepted risk in the design decisions, under the heading about known risks of the state model.*

Before going back to his own work, the agent checks that his main task is still in good
shape.

```console
$ biso set TASK-19 --priority high
note: TASK-19 unchanged
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso set`](../spec/cmd/set.md), [El modelo de datos de una tarea](../spec/modelo-de-datos/index.md); not literal spec text)*

*Note: Another write that changes nothing and renews `TASK-19`'s reservation until 240 minutes after this moment. With this, the task enters the next scenario with a fresh lease, and you already know that didn't cost any special command.*
