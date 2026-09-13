<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 5. You get to work on it

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

The agent already knows `TASK-19` is his. Now he's going to actually start it, and that isn't
done by changing the state by hand: there's a command named exactly after the move you're
making.

One single call does four things at once, and you didn't ask for one of them.

!!! abstract "What this scenario teaches"
    - Starting a task is a command with its own name, `biso start`, and it costs one call. It is not `biso set --status`, even though the state ends up the same.
    - Starting creates a lease: a reservation with an expiry date that says someone is working on that task right now. No other command creates one.
    - A lease is not a state. The task is `In Progress` and also holds a reservation; those are two separate things stored apart from each other.

The agent takes the task, and along the way writes down how he plans to attack it, so Sara
can read it without having to ask.

```console
$ biso start TASK-19 --append-plan "Reproduce the 5xx with a test server, and wrap the upload in a retry with exponential backoff"
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso start`](../spec/cmd/verbos-del-ciclo.md#biso-start), [La urgencia](../spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: The urgency has gone up from 7.0 to 11.0, and that's exactly the 4.0 of the active-task term. Nobody wrote that number: it gets recalculated on every read. What the output doesn't say is the most important part of this step. `biso start` has done four things in a single write: it set the active state, it checked that the task already had someone assigned (`@claude`, since scenario 4, so it didn't touch that), **it took the lease** in the caller's name, and it added the plan. If the task had had nobody assigned, `start` would have assigned it to whoever called it, and that's the difference with what Sara did in the previous scenario: she assigned without starting, and `start` starts by assigning if it has to.*

The lease doesn't show up in the status line, so the agent looks at the task's record to
understand what he just reserved.

```console
$ biso get TASK-19 --section plan
TASK-19  Retry the upload on 5xx

## Implementation Plan
Reproduce the 5xx with a test server, and wrap the upload in a retry with exponential
backoff
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: The plan is right where it should be. The lease needs explaining, because this step doesn't show it directly: it's two fields stored on the task, until when the reservation is good and who holds it, and the program just set them to "now plus 240 minutes" and `@claude`. Those 240 minutes are the `lease_minutes` key in the board's configuration, and you can change it whenever you want without breaking anything: the expiry is computed at write time, so changing it only affects leases that get renewed from that point on. The default is generous on purpose. The mistake that costs you is a false expiry, not a late one: too short a lease lets another agent claim a task someone is genuinely working on, while too long a lease only delays a warning. And as you'll see in the next scenario, that warning never stops anyone from working.*
