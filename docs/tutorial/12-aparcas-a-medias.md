<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 12. You park something halfway

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

In scenario 6 the agent picked up `TASK-52`, the task of documenting the release checklist that
`@bob` had left abandoned. It seemed like someone had to pick it up.

Now, with the upload already fixed and closed, it looks at it again and changes its mind:
nobody has shipped a release in weeks, there's no rush, and keeping it in progress under its
name gives a false picture of what's actually happening on the board. It's not that it's done.
It's that it's not the right time.

That has its own gesture for it, and it's not the one most people reach for first.

!!! abstract "What this scenario teaches"
    - Archiving isn't finishing. An archived task doesn't move to the terminal state: it leaves the active board in whatever state it already had.
    - There's no command to delete a task, and that absence is itself part of the specification. An archived task's identifier is never reused.
    - Archiving clears the lease, even if it was still alive and even though archiving doesn't move the task out of an active state. A lease asserts that someone is working on it right now, and archiving says exactly the opposite.

Almost everyone's first instinct is to look for how to delete it.

```console
$ biso delete TASK-52
error: there is no delete command, on purpose
hint: `biso archive <ref>` takes it off the board and keeps the history
      an archived task still exists: `biso ls --archived` lists them, and the
      id is never reused
```

Exit code: `2`

*Note: This command's absence is written into the specification, with its own message and its own code, instead of letting `biso delete` fall into the generic unknown-command error. The difference matters: an unknown command makes you wonder whether you typed it wrong, and this message tells you the decision is deliberate and which gesture you were actually looking for.*

The agent takes the task off the board.

```console
$ biso archive TASK-52
TASK-52  In Progress  ac 0/1  urgency 5.4  archived
```

Exit code: `0`

*(derived output, see [`biso archive`](../spec/cmd/archive.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: Read that line carefully, because it says two things that look incompatible: the task is still `In Progress` and it's `archived`. There's no contradiction. Archiving isn't a state and it doesn't touch the state: it's a separate flag that takes the task off the active board. If it were unarchived tomorrow, it would go back to `In Progress`, which is where it already was. And there's something the line doesn't say: the lease the agent had claimed back in scenario 6 has just been cleared. That always happens on archiving, whoever holds the lease and whether it's still alive or already expired, and the reason is that keeping it around would stash it somewhere nobody looks, because both the startup message and the listing hide archived tasks by default. If this task came back to the board three weeks from now with a lease still held by a session that had already died, that would be garbage wearing the costume of information. Urgency is still being calculated, 5.4, because the task is still in a state that isn't the terminal one. Archiving doesn't zero it out: finishing it does.*

It checks that it hasn't lost the task.

```console
$ biso ls --archived
TASK-52  In Progress  task  low  Document the release checklist  ac 0/1  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](../spec/cmd/ls.md), [`biso archive`](../spec/cmd/archive.md); not literal spec text)*

*Note: There it is, intact, still assigned to the same person. The only thing it's lost is the lease. Notice that `--archived` is what brings it back into view: without that flag it shows up in neither `biso ls` nor `biso prime`, and that's what "leaving the active board" actually means. The task hasn't gone anywhere, it's just been moved out of the way.*

And to close the idea, it checks what's still in progress now.

```console
$ biso ls --status "In Progress"
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](../spec/cmd/ls.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: Three tasks where there used to be four, and `TASK-19` isn't in either list anymore because it got finished back in scenario 10. The board now tells the truth about what's actually being worked on, which is exactly what archiving something you're not going to touch is for.*
