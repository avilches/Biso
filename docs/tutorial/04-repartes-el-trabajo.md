<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 4. You hand out the work

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

So far you've been looking at the board and jotting things down. This is different: there's a
task someone has to do, and someone has to say who. On this project a person, `@sara`, and an
agent, `@claude`, work on the same board, and `TASK-19` (retry the upload when the server
returns a 5xx) has been sitting there unowned since the start.

Sara decides the agent should do it. What follows is how that gets said, and how he finds out.

!!! abstract "What this scenario teaches"
    - Assigning is one person's decision about what another should do. It is not an automatic handout: `biso` never assigns anything on its own.
    - Identity isn't a property of the board at all: it's a setting of the machine each person or agent runs on, and the `BISO_ME` environment variable always wins over it. A person and an agent sharing a board just each declare their own `BISO_ME` in their own session.
    - Assigning a task does not create a lease. Receiving work and starting it are two different things, and the second one has its own command.

Sara opens a new terminal and, before anything else, wants to see what she has pending. She
hasn't declared who she is in this session yet.

```console
$ biso ls --mine
error: --mine needs an identity; set BISO_ME, or add "me" to ~/.biso/config.json
```

Exit code: `6`

*Note: Notice that it doesn't return an empty list. It could have, and that would have been the easy thing to implement: with no identity, no task is "mine", so zero results. But then whoever calls it couldn't tell "I have nothing assigned" apart from "you haven't said who you are", and that is exactly what the first principle of the specification forbids. The error also has its own code, 6, so a program calling `biso` can react without reading the message. The message offers two ways out, and they aren't equivalent. `~/.biso/config.json` is a setting of this machine, so writing `me` there sets your identity for every board and every session that runs here, agent sessions included unless they declare their own. `BISO_ME` declares it only for this one shell. Sara uses the second, so the agent's own identity keeps winning in its own sessions regardless of what her machine has saved.*

Sara declares her identity for this session, which is an environment variable and not a
`biso` command, and hands the task to the agent.

```console
$ biso set TASK-19 --add-assignees @claude
TASK-19  To Do  ac 0/2  urgency 7.0
```

Exit code: `0`

*(derived output, see [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: The output doesn't repeat what Sara just wrote: it doesn't say "assignee: @claude". It says the state the task is left in, its acceptance criteria progress, and its urgency, which is what she didn't know. That's the fourth principle of the specification, and it holds for every write. The urgency is 7.0, which comes from adding 6.0 for being high priority and 1.0 for having acceptance criteria. It doesn't add the active-task term, worth 4.0, because the task is still `To Do`: assigning a task doesn't put it in motion.*

In his own session, the agent asks what's his to do. Here `--mine` does work, because his
identity is declared.

```console
$ biso ls --mine
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-19  To Do        task  high    Retry the upload on 5xx                           ac 0/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](../spec/cmd/ls.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: Six rows where there used to be five. `TASK-19` comes in second, and that spot isn't random: the list sorts by descending urgency, and `TASK-19` scores 7.0, same as `TASK-60`. A tie is always broken by ascending identifier, and 19 comes before 60. There's a reason `TASK-60` scores the same as a task that hasn't even started, despite being high priority and already in progress: `TASK-60` has an open question, and the active-task term only adds up when the task is in the active state **and** isn't waiting on an answer. A parked task isn't being worked on by anyone, so it doesn't compete for your attention. Scenario 8 shows the inside of that. And watch what this list doesn't say: `TASK-19` already belongs to the agent, but nobody has started working on it. There is no lease. That's what the next scenario's command is for.*

The startup message tells the same story another way. If the agent ran it again now,
`TASK-19` would have moved.

```console
$ biso ls --mine --status "To Do"
TASK-19  To Do  task  high    Retry the upload on 5xx                ac 0/2  @claude  -
TASK-61  To Do  docs  medium  Rewrite the install section            ac 0/1  @claude  -
TASK-33  To Do  task  medium  Add a retry counter to the upload log  ac 1/3  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](../spec/cmd/ls.md); not literal spec text)*

*Note: In `biso prime`, these three are the ones that show up in the `ASSIGNED TO YOU` block, which means exactly this: tasks a person decided you should do and that aren't under way yet. In scenario 1 you saw that whole message, and back then `TASK-19` wasn't in that block because it belonged to no one. Also notice that the columns have narrowed. The width of each one is decided by the content of the list being printed, not a fixed template: once you filter down to a single state, the status column no longer needs room for `In Progress`, and once the long-titled tasks disappear, the title column shrinks along with them.*
