<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# biso tutorial, by scenario

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/` and from `tutorial/conceptos.md`. Do not edit it by
    hand: any change is lost on the next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

## Concepts

`biso` is the tool an automated agent, and the person working alongside it, use to run the tasks of
a project. Before looking at a single command, five concepts are worth having straight, because the
rest of the tutorial takes them for granted.

### The board

A board is the set of a project's tasks together with its configuration: what states exist, what
task types there are, what priorities can be used, and who, if anyone, the identity of whoever works
on it is. That configuration is not a side detail: it is what makes a value meaningful or not. A
board does not accept any text in any field, only the values of its own vocabulary.

From that comes the rule that governs everything else: a value the board does not recognize is
always an error, whether you are writing it or asking about it, and with the same strictness in both
cases. Asking about a state that board does not have does not return an empty list, it returns an
error. As a consequence, when a question does return an empty list it is because there really is
nothing that matches what was asked, and that is information, not a silent failure.

### States

Every task has a state, and the state is one of the ones that board has configured: there is no
fixed list of states valid for every project, each board declares its own. The only thing that
cannot be freely configured is that, among those states, three roles always stay covered, each by a
different state:

- The **initial** state, where a new task is born.
- The **active** state, the one written when someone picks up a task to work on it.
- The **terminal** state, the one that marks a task as done.

A role is the function a state serves, not its name. Two boards can call the initial state
something different and both still have one, because the role is mandatory even though the word
used for it is not.

### Acceptance criteria and definition of done

A task can carry two independent checklists, with the same shape: each item has text and can be
checked or not, and each item is born with its own numeric key that does not change even if other
items are removed from the list. They are two lists, not one, and each is counted separately.

The first are the **acceptance criteria**: how you know that particular task's work actually works.
The second is the **definition of done**: what has to be true before calling the task closed, beyond
whether the result works, such as someone else having reviewed it. Neither is mandatory, and a task
can reach its terminal state with unchecked items in either of the two, because `biso` warns about
that but does not block it by default.

### Urgency

Urgency is a number that is not stored anywhere: it is computed every time the task is read, from
data that is stored, such as priority, whether the task is active, whether it blocks or is blocked
by another task, whether it has a due date coming up, whether it has acceptance criteria, and how
long it has been open. That is why urgency is called **derived**: nobody writes it directly, and its
value right now may not be its value a minute from now even if nobody touched the task, simply
because time passed or some other task it depends on changed. A task in its terminal state always
has urgency zero, with no further calculation.

### Declared identity

A board can have an identity configured: the text that says who you are when you call `biso`. It is
not mandatory: a board works perfectly well with nobody declaring one. What changes is that some
operations need to know who you are to make sense, such as asking only for the tasks that are yours,
getting automatically assigned a task when you start it, signing a comment with your name by default,
or leaving an open question waiting for someone to answer. If you try any of those without a declared
identity, `biso` does not guess or assume anything: it says so as an explicit error. The rest of the
board, whatever does not need to know who you are, keeps working the same with or without a declared
identity.

## Scenarios

1. [You land on a project you don't know](#escenario-01)
2. [Something comes to mind and you don't want to lose it](#escenario-02)
3. [What now?](#escenario-03)
4. [You hand out the work](#escenario-04)
5. [You get to work on it](#escenario-05)
6. [Two at once on the same board](#escenario-06)
7. [Halfway through, a criterion is missing](#escenario-07)
8. [You get stuck and someone has to decide](#escenario-08)
9. [The person and the agent talk](#escenario-09)
10. [You close it](#escenario-10)
11. [You get it wrong](#escenario-11)
12. [You park something halfway](#escenario-12)
13. [You touch twenty at once](#escenario-13)

## 1. You land on a project you don't know {: #escenario-01 }

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

COMMANDS  (`biso <cmd> --help` for the detail of any flag)
  biso ls [-s STATUS] [--type T] [-l LABEL] [--mine] [--search TEXT]
  biso get <ref> [--section ac]
  biso new "TITLE" [-d TEXT] [--ac TEXT]... [--type T] [--priority P]
  biso start <ref>... [--plan TEXT]
  biso note <ref> "TEXT"
  biso ask <ref> "QUESTION"
  biso answer <ref> "TEXT"
  biso finish <ref>... [--summary "TEXT"] [--check all] [--check-dod all]
  biso set <ref>... [any field flag]
  biso comment <ref> "TEXT" [--comment-author @who]

FIELD FLAGS  (same names, same meaning, in every command above that writes)
  -t --title  -s --status  --type   --priority  --project      -a --assignee
  -l --label  -d --desc    --ac     --dod       --plan         --note
  --summary   --dep        --ref    --doc       --file         -m --milestone
  -p --parent --due        --ordinal --ext K=V  --reporter     --comment
  --check     --uncheck             --check-dod --uncheck-dod

RULES  (none of these are guessable; they are the whole learning curve)
  1. Every write goes through biso. Nothing else touches the board.
  2. A bare field flag ADDS. Replacing and removing are explicit: --label X
     adds, --set-label X replaces the list, --rm-label X drops one, and
     --clear-label empties it. Same four shapes for every list field.
  3. <ref> is an id (TASK-12), a bare number (12) or free text ("CRLF"). Text
     matching several tasks is an error that lists them, never a guess. `note`,
     `comment`, `ask` and `answer` take one <ref>; `set`, `start` and `finish`
     take several.
  4. Filters reject values this board does not have: `-s Pending` is an error,
     not an empty list. Case, spaces, hyphens and underscores are ignored, so
     `-s todo`, `-s "To Do"` and `-s TO_DO` are one and the same filter. An
     empty list is therefore a fact about the board that you can act on.
  5. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says
     on stderr what it left out. --all lifts the limit, --any-status includes
     Done, --archived reaches the archive.
  6. --check and --uncheck take all, 3, 1-4, 1,3,7 or the criterion text. The
     numbers are the stable #N keys that `biso get` shows, and they never
     shift when one criterion is removed.
  7. `biso new` prints the new id and nothing else. Every other write prints one
     line per task: id, status, criteria, urgency. Add --print for the whole
     record, or --json for a versioned envelope.
  8. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.
  9. Long text: a real newline works, and so do -d @file.md and -d - for stdin.
 10. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,
     6 precondition not met, 7 environment, 8 no board here, 9 nothing written.
 11. `biso ask <ref> "..."` parks a task on a question and `biso answer` unparks
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

Pick one, `biso start <ref> --plan "..."`, work, `biso note <ref> "..."` as you go,
and close with `biso finish <ref> --check all --check-dod all --summary "..."`.
That is the loop. Create a task when the work needs planning or review; do small
edits directly.
```

Exit code: `0`

*Note: This is the whole message, top to bottom, and there is no second screen after it. The BOARD block states the actual vocabulary of this board (the types, the priorities, the three states and their role) and who you are here. The four sections below split up the entire board without any task showing up in two places: what's in progress, what's waiting on an answer from you, what's yours, and what's next by urgency. No one writes that urgency by hand, it's recalculated every time you ask for it (you'll come back to this in scenario 3, "What now?"). With this alone you could already create a task, start it, work on it and close it without opening any other document: that's the standard this message is written to.*

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

*Note: Four pieces of data, each changing on its own: the id never changes, the name changes with `biso config set project_name`, the path changes if someone moves the folder by hand, and `source` says where that answer came from, which here is the project pointer (how the board gets resolved in the first place). If there were no board at all, both `prime` and `where` would fail with exit code 8, and `where`'s own error message points to `biso init` as the fix: it creates a new board and leaves it pointed at from this project. All you need right now is to know it exists; its flags matter only when you actually need to create one.*

## 2. Something comes to mind and you don't want to lose it {: #escenario-02 }

You're in the middle of something else when you notice something that needs fixing in
the file uploader. It's not what you're working on right now, and if you keep pulling on
that thread you lose track of the real task. All you need is to write it down somewhere
it won't get forgotten, and keep going.

!!! abstract "What this scenario teaches"
    - Each field flag says what it does in its own name. Repeating `--add-ac` twice leaves two criteria, not one overwriting the other.
    - `biso new` prints the new task's id and nothing else: no title echoed back as confirmation, no status line.

Just typing the title would already keep it safe, but since it's already in your
head, it's worth also writing down why it matters and what criterion will tell you
it's resolved. That's two more flags on the same line, and `new` accepts all of them:

```console
$ biso new "Log the diff size before uploading" --add-ac "The log prints the diff size before each upload" --add-ac "The size also shows up when the upload fails"
TASK-62
```

Exit code: `0`

*(derived output, see [`biso new`](spec/cmd/new.md); not literal spec text)*

*Note: No "created TASK-62", no repeating the title: the only line is the id, because that's exactly what you couldn't have known beforehand, and you already knew everything else (rule 7 of the startup message, and one of the core principles: the default output of a write is whatever the caller didn't already know, never an echo of what it just wrote). The two `--add-ac` flags haven't replaced anything, they've added two separate criteria to a list that started out empty, because that is exactly what `--add-ac` says it does. There is no flag that replaces the whole list of criteria: doing that means clearing it and adding again in the same call, though you didn't need that here, a freshly created task has nothing to clear yet.*

## 3. What now? {: #escenario-03 }

You already know which board this is and you've already kept your idea safe. But the
startup message showed you the whole map of the project, not your own corner of it, and
you have twenty minutes before the next meeting. You want to see what's on your plate
right now, then pick one and read it in full before touching anything.

!!! abstract "What this scenario teaches"
    - `biso ls` shows at most 30 tasks and leaves out the ones already Done; what gets left out is reported on stderr, never mixed in with the data rows on stdout.
    - No one writes down the urgency that decides the order, it's recalculated every time you ask for the list, which is why it never shows up as a column.

You look at what's assigned to you, with no other filter:

```console
$ biso ls --mine
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md); not literal spec text)*

*Note: Five rows, not one more: since that's fewer than the 30 `ls` shows by default, no truncation warning comes out on stderr, just the rows on stdout. The order isn't by id or by when they were created: it's descending urgency, which is the default criterion whenever no task in the list carries a manual `ordinal` (per the rules for `biso ls`). TASK-11 comes first because it adds up high priority, being active, and blocking another unfinished task; TASK-40 comes last despite also being in progress, because it is itself blocked by TASK-11, and that subtracts instead of adding. No one writes that number down or stores it anywhere: it's recalculated the instant you read it (per how urgency is defined), which is why there's no "urgency" column in this table, only the order it produces.*

On a real board that list doesn't always fit in 30 rows, and the notice about what got
left out never mixes in with the tasks: it goes to stderr, so redirecting the output to
a file leaves a clean data file. To see that without waiting for the board to grow, you
set a limit lower than what's assigned to you:

```console
$ biso ls --mine --limit 3
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
warning: 2 more tasks match; showing 3 of 5
hint: narrow with -s, --type or -l, or ask for everything with --all
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md); not literal spec text)*

*Note: The first three rows go to stdout, and the warning and the hint go to stderr: in your terminal you see them one after another, but if you save the output to a file (`biso ls --mine --limit 3 > tareas.txt`) that file has three lines, not five. Here you set the limit yourself with `--limit 3` to see this with little data; without that flag the default limit is 30, and with a board this size (54 To Do plus 4 In Progress, not counting the 190 Done that `ls` hides by default) it's easy to reach those 30 without even trying. `--all` removes the limit entirely, and `--any-status` is the only thing that brings back the finished tasks.*

Of the five, TASK-33 is the one that looks most like what you just left behind in the
previous scenario: written down, with a criterion or two already thought out but not
started. Before deciding whether it's really yours, you read it in full:

```console
$ biso get TASK-33
TASK-33  Add a retry counter to the upload log
status     To Do                type       task
priority   medium               urgency    4.3
assignees  @claude              reporter   @avilches
labels     -                    milestone  -
parent     -                    due        -
project    -                    ordinal    -
created    2026-08-20 09:40     updated    2026-09-03 16:15
depends    -                    blocks     -
refs       -
docs       -
files      -
ext        -

## Description
Every failed upload retries silently. The log does not say how many attempts it took,
so debugging one means reconstructing it by hand from scattered timestamps.

## Acceptance Criteria
- [x] #1 Every retry is counted in the log
- [ ] #2 The counter shows up in the final summary
- [ ] #3 A test covers three retries in a row

## Definition of Done
(empty)

## Implementation Plan
(empty)

## Implementation Notes
(empty)

## Final Summary
(empty)

## Comments
(empty)

## Open Question
(empty)
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: There's no `lease` line: that only shows up when the task has a lease, and a task in To Do can't have one (per the task data model). The nine sections always appear, even when empty and marked `(empty)`, because leaving out one that wasn't requested would be confused with one that was requested and came back empty. Notice the `#1`, `#2` and `#3` keys in front of each criterion: they're stable, not a position in the list, so if `#2` ever gets removed the other two stay `#1` and `#3`, they never get renumbered.*

## 4. You hand out the work {: #escenario-04 }

So far you've been looking at the board and jotting things down. This is different: there's a
task someone has to do, and someone has to say who. On this project a person, `@sara`, and an
agent, `@claude`, work on the same board, and `TASK-19` (retry the upload when the server
returns a 5xx) has been sitting there unowned since the start.

Sara decides the agent should do it. What follows is how that gets said, and how he finds out.

!!! abstract "What this scenario teaches"
    - Assigning is one person's decision about what another should do. It is not an automatic handout: `biso` never assigns anything on its own.
    - A board shared between a person and an agent leaves the `me` key unset on purpose, and each one declares their identity in their own session with `BISO_ME`. If the board had it set, everyone would share the same identity and `--mine` would stop meaning anything.
    - Assigning a task does not create a lease. Receiving work and starting it are two different things, and the second one has its own command.

Sara opens a new terminal and, before anything else, wants to see what she has pending. She
hasn't declared who she is in this session yet.

```console
$ biso ls --mine
error: --mine needs an identity; set it with biso config set me <you> or BISO_ME
```

Exit code: `6`

*Note: Notice that it doesn't return an empty list. It could have, and that would have been the easy thing to implement: with no identity, no task is "mine", so zero results. But then whoever calls it couldn't tell "I have nothing assigned" apart from "you haven't said who you are", and that is exactly what the first principle of the specification forbids. The error also has its own code, 6, so a program calling `biso` can react without reading the message. The message offers two ways out, and they aren't equivalent. `biso config set me` writes it into the board's configuration, where everyone would see it; `BISO_ME` declares it only for this session. On a shared board you have to use the second one.*

Sara declares her identity for this session, which is an environment variable and not a
`biso` command, and hands the task to the agent.

```console
$ biso set TASK-19 -a @claude
TASK-19  To Do  ac 0/2  urgency 7.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: The output doesn't repeat what Sara just wrote: it doesn't say "assignee: @claude". It says the state the task is left in, its acceptance criteria progress, and its urgency, which is what she didn't know. That's the fourth principle of the specification, and it holds for every write. The urgency is 7.0, which comes from adding 6.0 for being high priority and 1.0 for having acceptance criteria. It doesn't add the active-task term, worth 4.0, because the task is still `To Do`: assigning a task doesn't put it in motion. The `dod` chunk doesn't show up because `TASK-19` has no definition of done, and that chunk only appears when there is one.*

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

*(derived output, see [`biso ls`](spec/cmd/ls.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Six rows where there used to be five. `TASK-19` comes in second, and that spot isn't random: the list sorts by descending urgency, and `TASK-19` scores 7.0, same as `TASK-60`. A tie is always broken by ascending identifier, and 19 comes before 60. There's a reason `TASK-60` scores the same as a task that hasn't even started, despite being high priority and already in progress: `TASK-60` has an open question, and the active-task term only adds up when the task is in the active state **and** isn't waiting on an answer. A parked task isn't being worked on by anyone, so it doesn't compete for your attention. Scenario 8 shows the inside of that. And watch what this list doesn't say: `TASK-19` already belongs to the agent, but nobody has started working on it. There is no lease. That's what the next scenario's command is for.*

The startup message tells the same story another way. If the agent ran it again now,
`TASK-19` would have moved.

```console
$ biso ls --mine -s "To Do"
TASK-19  To Do  task  high    Retry the upload on 5xx                ac 0/2  @claude  -
TASK-61  To Do  docs  medium  Rewrite the install section            ac 0/1  @claude  -
TASK-33  To Do  task  medium  Add a retry counter to the upload log  ac 1/3  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md); not literal spec text)*

*Note: In `biso prime`, these three are the ones that show up in the `ASSIGNED TO YOU` block, which means exactly this: tasks a person decided you should do and that aren't under way yet. In scenario 1 you saw that whole message, and back then `TASK-19` wasn't in that block because it belonged to no one. Also notice that the columns have narrowed. The width of each one is decided by the content of the list being printed, not a fixed template: once you filter down to a single state, the status column no longer needs room for `In Progress`, and once the long-titled tasks disappear, the title column shrinks along with them.*

## 5. You get to work on it {: #escenario-05 }

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

*(derived output, see [`biso start`](spec/cmd/verbos-del-ciclo.md#biso-start), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

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

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: The plan is right where it should be. The lease needs explaining, because this step doesn't show it directly: it's two fields stored on the task, until when the reservation is good and who holds it, and the program just set them to "now plus 240 minutes" and `@claude`. Those 240 minutes are the `lease_minutes` key in the board's configuration, and you can change it whenever you want without breaking anything: the expiry is computed at write time, so changing it only affects leases that get renewed from that point on. The default is generous on purpose. The mistake that costs you is a false expiry, not a late one: too short a lease lets another agent claim a task someone is genuinely working on, while too long a lease only delays a warning. And as you'll see in the next scenario, that warning never stops anyone from working.*

## 6. Two at once on the same board {: #escenario-06 }

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
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Exit code: `0`

*(derived output, see [Notas y avisos](spec/salida-y-terminal.md#notas-y-avisos), [`biso start`](spec/cmd/verbos-del-ciclo.md#biso-start), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: First thing: the note got written. The exit code is 0 and the task has changed. The warning is not a rejection, it's information. That's deliberate, and the specification argues for it: a flow block doesn't prevent duplicate work. If `biso` had told her "you can't write here", Sara would have written the note somewhere else, or edited the store by hand, and then the board would be lying. It's the same call as with unfinished dependencies and open questions: it warns, it doesn't block. Second thing, and it's easy to miss: this write **hasn't touched the lease**. It hasn't renewed it, because Sara isn't the one holding it, and it hasn't stolen it, because only `biso start` does that. It's still `@claude`'s and it still expires at the same time. The urgency is still 19.0 because a note doesn't change anything urgency measures. And that 19.0 breaks down as: 6.0 for high priority, 4.0 for being active, 8.0 because another unfinished task depends on it, and 1.0 for having acceptance criteria.*

The agent, in his own session, keeps going with his work and notes his own progress on the
same task.

```console
$ biso note TASK-11 "The normalizer now handles the test files case"
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Exit code: `0`

*(derived output, see [`biso note`](spec/cmd/verbos-del-ciclo.md#biso-note), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: No warning, because the lease is his. And something the output doesn't say: it just got renewed until 240 minutes after this moment. That's the central idea. There's no `biso heartbeat`, no `biso renew`, no flag to ask for one. Any write by the holder on their task renews the reservation, and "any" means all of them: the six cycle verbs, `biso set`, and `biso archive`. Working is what keeps the reservation alive, which is exactly what you want it to mean.*

The agent wants to confirm that the task's priority is still what he thinks it is. He
writes it again with the same value it already had.

```console
$ biso set TASK-11 --priority high
note: TASK-11 unchanged
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

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

*(derived output, see [`biso ls`](spec/cmd/ls.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `TASK-52` is still `In Progress`, perfectly normally, and has nobody assigned. Its lease expired on September 5th and the stored state hasn't moved an inch. That's not an oversight, it's the rule: what expires is the claim, not the state. Nothing is going to pull it out of `In Progress` on its own, because doing so would mean a command touching tasks you didn't name, and because `biso prime`, which never writes, would end up showing a state that someone else's next write could change. The order of this list, in case you're wondering, is descending urgency and has nothing to do with state: 19.0, 11.0, 7.0, 5.4, and 3.1. `TASK-40` comes last, despite being in progress and medium priority, because it depends on `TASK-11` and being blocked subtracts 5.0.*

The agent decides to take on that abandoned task himself.

```console
$ biso start TASK-52
note: TASK-52 was already In Progress
TASK-52  In Progress  ac 0/1  urgency 5.4
```

Exit code: `0`

*(derived output, see [`biso start`](spec/cmd/verbos-del-ciclo.md#biso-start), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: With a single call he's done three things: he's seen that the state was already active and says so with that note, he's been assigned the task because nobody had it, and **he's claimed the expired lease** in his own name. That claim happens inside the same transaction that checks it was still expired, and that detail matters: if two agents claim the same expired lease at the same instant, only one wins. There's no window where both believe it's theirs. What that check doesn't do is silence `@bob` if he ever comes back. He can still note, comment on, and close this task, because writing is never forbidden; what he can no longer do is renew or reclaim a reservation that now belongs to someone else, and if he wants it back he has to ask for it again with `biso start`, and he'll get the same warning as the first step. It is a deliberate difference from the classic fencing-token reservation pattern, and it's documented as an accepted risk in the design decisions, under the heading about known risks of the state model.*

Before going back to his own work, the agent checks that his main task is still in good
shape.

```console
$ biso set TASK-19 --priority high
note: TASK-19 unchanged
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: Another write that changes nothing and renews `TASK-19`'s reservation until 240 minutes after this moment. With this, the task enters the next scenario with a fresh lease, and you already know that didn't cost any special command.*

## 7. Halfway through, a criterion is missing {: #escenario-07 }

You've been at `TASK-19` for a while. You're noting down what you check, but you still haven't
written a definition of done: not a word about what has to happen, beyond the acceptance
criteria, for this task to really count as finished. You stop to write it down before going on.

You check first whether there's already something there.

!!! abstract "What this scenario teaches"
    - Each field flag says what it does in its own name: `--add-` adds, `--rm-` removes, `--clear-` empties. For criteria and definition-of-done items there's no whole-list replace; doing that means clearing and adding in the same call.
    - The `#N` keys of a criterion, whether an acceptance criterion or a definition-of-done item, get assigned when the element is created and are never reassigned. Removing one from the middle doesn't renumber the ones that are left.
    - A list's key counter only ever grows, even after `--clear-`. Elements added after emptying the list don't get back the keys of the ones that were removed.

You ask for just that section, so you don't pull the whole record for something you already
know is going to be empty.

```console
$ biso get TASK-19 --section dod
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: Not even the header shows up. `biso get`'s rule is literal: with `--section`, an empty section doesn't get printed, and if no section is left to print, the whole output is empty. That's different from asking for the full record, where every empty section shows up anyway, marked `(empty)`. The flag changes the standard on purpose: without it, leaving a section out would blur "empty" with "not requested"; with it, you asked for exactly that section, so "empty" no longer needs saying.*

You write down the first thing that comes to mind: that the change should be reviewed by
someone else before it counts as done.

```console
$ biso set TASK-19 --add-dod "Someone else reviews the change before the task counts as done"
TASK-19  In Progress  ac 0/2  dod 0/1  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `--add-dod` adds. It's the same verb you already know from `--add-ac`: it behaves the same way for any list field.*

And a second thing that needs to be clear before closing: that the retry behavior gets
documented somewhere, for whoever reads this later.

```console
$ biso set TASK-19 --add-dod "The retry behavior is documented in the README"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

You check how they turned out, with their keys.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #1 Someone else reviews the change before the task counts as done
- [ ] #2 The retry behavior is documented in the README
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

On second thought, the first one is redundant: having someone else review the change is
already team policy for any PR, not something specific to this task. You remove it.

```console
$ biso set TASK-19 --rm-dod 1
TASK-19  In Progress  ac 0/2  dod 0/1  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

You check the list again to make sure what's left hasn't changed number.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #2 The retry behavior is documented in the README
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables), [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: This is the central point of the scenario. The surviving element is still `#2`, it hasn't become `#1`. The program fixes a key when the element is created and never moves it: that's why checking or removing by number is safe even when the list has lost elements along the way, and why a task can perfectly well have criteria `#2` and `#7` without that being anyone's mistake.*

A coworker reviews the wording and suggests rewriting the whole list at once, adding a
second entry about the tests that are needed.

```console
$ biso set TASK-19 --clear-dods --add-dod "The README explains when and how many times it retries" --add-dod "The test suite covers the 5xx case"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [Sustituir un campo que no tiene bandera de \"sustituir entera\"](spec/familias-de-banderas.md#sustituir-un-campo-que-no-tiene-bandera-de-sustituir-entera), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: There's no flag that replaces the whole list of definition-of-done items in one shot, so this clears it and adds the two new items in the same call. Element `#2` disappears along with the rest of the old list, and the two that come in are brand new elements. No warning shows up here, unlike a `--replace-*` on a comma-separated list field that overwrites non-empty content: clearing was explicit, you typed it yourself, so nothing gets overwritten silently and there's nothing to warn about.*

You check the keys of the two new ones.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #3 The README explains when and how many times it retries
- [ ] #4 The test suite covers the 5xx case
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables), [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: The keys are `#3` and `#4`, not `#1` and `#2`. This list's counter, inside this task, was already at 2 before anything got cleared, and clearing doesn't reset it: "the keys of earlier elements are never reused" holds even when the clear-and-add comes after an `--rm-`, not just right after creating the task.*

Almost right away you realize you're polishing the wording before the code is properly
settled. You'd rather leave it blank until the work is more mature.

```console
$ biso set TASK-19 --clear-dods
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: The list ends up empty, and that's why the `dod` chunk disappears from the status line: it only shows up "whenever the task has a definition of done," and an empty list doesn't count as having one. `ac` keeps showing because those two criteria haven't been touched.*

A while later, with the retry already working, you go back to the final list. It's the
same wording as before.

```console
$ biso set TASK-19 --add-dod "The README explains when and how many times it retries" --add-dod "The test suite covers the 5xx case"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `--add-dod` again. The keys of these two elements are `#5` and `#6`: the list's counter hasn't been reset either by the earlier `--rm-dod` or by the later `--clear-dods`, it "only ever grows," no matter what happens to the list's content.*

## 8. You get stuck and someone has to decide {: #escenario-08 }

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
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 7.0
```

Exit code: `0`

*(derived output, see [`biso ask`](spec/cmd/verbos-del-ciclo.md#biso-ask), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

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

*(derived output, see [`biso get`](spec/cmd/get.md), [La pregunta abierta](spec/modelo-de-datos.md#la-pregunta-abierta); not literal spec text)*

A while later, looking at a similar endpoint, you find the answer yourself: there's already
a value fixed somewhere else in the code. You answer, and the task gets unparked.

```console
$ biso answer TASK-19 "There is already a reference value: the download endpoint uses 3 retries with exponential backoff in retry.go. Use the same value and the same logic here."
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso answer`](spec/cmd/verbos-del-ciclo.md#biso-answer), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

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

*(derived output, see [Los comentarios](spec/modelo-de-datos.md#los-comentarios), [Las fechas](spec/modelo-de-datos.md#las-fechas), [`biso get`](spec/cmd/get.md), [`biso answer`](spec/cmd/verbos-del-ciclo.md#biso-answer); not literal spec text)*

*Note: The first comment carries the moment the question was asked (14:10), not the moment it was answered (15:30): that's the one documented exception to the rule, the question gets carried over exactly as the program first recorded it. The two comments appear in the order `biso answer` wrote them, which is also their insertion order, and that's the order that gets stored and shown, not the order their timestamps would suggest.*

## 9. The person and the agent talk {: #escenario-09 }

With the retry question settled, you keep implementing TASK-19. You keep a record of what you
check, so whoever reads the task later doesn't have to re-read the whole diff. But something
also comes in from outside that isn't yours: a user has reported a problem through another
channel, and someone asks you to leave it noted on the task. These are two different things and
it's easy to mix them up: an implementation note is your own notebook, with no author or
timestamp per entry; a comment is a conversation, and it always carries who wrote it and when.

!!! abstract "What this scenario teaches"
    - The `biso note` command adds a paragraph to the implementation notes. It's a block of prose with no author or timestamp per element, the technical notebook of whoever is doing the work.
    - The `biso comment` command adds a comment with an author and a timestamp. It's the channel for anything that comes from outside, and the way to talk with a person.
    - A comment's default author is the configured identity (`me`). `--comment-author` lets you sign it with another one, free text and unvalidated, to relay something that came from another system.

You finish fitting the retry with the same pattern the download endpoint already used, and
leave a record of how it turned out.

```console
$ biso note TASK-19 "The exponential backoff in retry.go can be reused as-is; the limit just needs to be set to 3."
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso note`](spec/cmd/verbos-del-ciclo.md#biso-note), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

You check how that note turned out.

```console
$ biso get TASK-19 --section notes
TASK-19  Retry the upload on 5xx

## Implementation Notes
The exponential backoff in retry.go can be reused as-is; the limit just needs to be set to 3.
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: There's no author or timestamp next to the text. Notes don't have that structure: they're a single block of prose that each `--note` adds a paragraph to, not a list of entries with an author like comments. If you needed to know who wrote a note or when, you wouldn't be able to: that's not the kind of data it is.*

You get a heads-up: a user on a mobile connection sees failures even with the retry in
place. It's not something you saw yourself, so you enter it as a comment, signed with
whoever reported it.

```console
$ biso comment TASK-19 "A user on a mobile connection sees failures even after the 3 retries, should we raise the limit in that case?" --comment-author @trello:juan
note: comment #3 by @trello:juan
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso comment`](spec/cmd/verbos-del-ciclo.md#biso-comment), [Los comentarios](spec/modelo-de-datos.md#los-comentarios), [Comentarios](spec/familias-de-banderas.md#comentarios); not literal spec text)*

*Note: "note: comment #3 by @trello:juan" goes to stderr, in the exact shape the specification's own example gives. The number is a stable key, exactly like a criterion's: it's assigned once when the comment is created and never reassigned, even if an earlier comment is later removed with `--rm-comment`. A comment's body and author are never edited, by any flag; only its date can be corrected, with `--set-comment-date`, and the whole comment can be removed entirely. This is TASK-19's third comment because `biso answer` had already added two in the previous scenario.*

You answer yourself, as yourself, with no author flags: the configured identity signs by
default.

```console
$ biso comment TASK-19 "Not for now, 3 retries is the policy for the rest of the system; if it happens again we'll revisit it."
note: comment #4 by @claude
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso comment`](spec/cmd/verbos-del-ciclo.md#biso-comment), [Los comentarios](spec/modelo-de-datos.md#los-comentarios); not literal spec text)*

*Note: Without --comment-author, the author is `me`, this board's configured identity: @claude. It's the same `--comment-author` that exists on every write command, not just `biso comment`: the name doesn't change just because it looks redundant on `biso comment`.*

You put the two sections side by side to see the contrast at a glance.

```console
$ biso get TASK-19 --section notes,comments
TASK-19  Retry the upload on 5xx

## Implementation Notes
The exponential backoff in retry.go can be reused as-is; the limit just needs to be set to 3.

## Comments
#1  @claude, 2026-09-06 14:10
What is the maximum number of retries before treating the 5xx as a definitive failure? Is there already a reference value somewhere else in the code, or does a new one need to be set?

#2  @claude, 2026-09-06 15:30
There is already a reference value: the download endpoint uses 3 retries with exponential backoff in retry.go. Use the same value and the same logic here.

#3  @trello:juan, 2026-09-06 16:05
A user on a mobile connection sees failures even after the 3 retries, should we raise the limit in that case?

#4  @claude, 2026-09-06 16:20
Not for now, 3 retries is the policy for the rest of the system; if it happens again we'll revisit it.
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md), [Los comentarios](spec/modelo-de-datos.md#los-comentarios); not literal spec text)*

*Note: Notes are a single block; comments are four signed entries, in the order they were written. `## Implementation Notes` comes before `## Comments` even though the flag asked for them in the opposite order (`notes,comments`): the print order of sections is the fixed order of the full record, not the order they're listed in `--section`. The specification doesn't say this explicitly for the case of several sections at once; noted as an open gap in tutorial/lagunas/07-10.md.*

## 10. You close it {: #escenario-10 }

The retry works, it's documented, and someone has given it a quick look over in the comments.
Before closing TASK-19 you go back over the acceptance criteria, and realize that two things
you've already done were never written down as a criterion: the backoff and the tests. You add
them before checking anything off, so the final record tells the whole story.

!!! abstract "What this scenario teaches"
    - The `--check-ac` and `--uncheck-ac` flags (and their `--check-dod` and `--uncheck-dod` counterparts) accept five selector forms, all repeatable and all combinable in the same call: a single key, a range, a comma-separated list, `all`, or the criterion's text.
    - Checking or unchecking is always safe to repeat. `--check-ac all` breaks nothing even if everything was already checked, and that's why closing out a cycle can always use it without first checking what's missing.
    - Closing a task always clears its lease, no matter who's holding it. It's the invariant that governs any write that takes a task out of the active state, and `biso finish` applies it without exception.

You add the two criteria you were missing, all at once.

```console
$ biso set TASK-19 --add-ac "The backoff follows the same pattern as the download endpoint" --add-ac "There's a test that covers the repeated 5xx case" --add-ac "There's a test that covers the success case after retrying" --add-ac "The endpoint documentation mentions the retry limit"
TASK-19  In Progress  ac 0/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](spec/familias-de-banderas.md#campos-de-lista-sin-coma-criterios), [`biso set`](spec/cmd/set.md), [Los criterios y sus claves estables](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables); not literal spec text)*

*Note: Four `--add-ac` flags in the same call add four new elements, in the order they're written, with new keys, #3, #4, #5 and #6, after the #1 and #2 the task already had from the initial inventory.*

You check off the first one, the one you'd already verified from the start.

```console
$ biso set TASK-19 --check-ac 1
TASK-19  In Progress  ac 1/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

The three you tested together with the same test, all at once, with a range.

```console
$ biso set TASK-19 --check-ac 3-5
TASK-19  In Progress  ac 4/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

And two that aren't consecutive, with a comma-separated list.

```console
$ biso set TASK-19 --check-ac 2,6
TASK-19  In Progress  ac 6/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

Almost right away you realize you marked #2 (the one about the log counter) too soon: the
number of attempts still doesn't show up in the log, only the final result. You uncheck it.

```console
$ biso set TASK-19 --uncheck-ac 2
TASK-19  In Progress  ac 5/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: `--uncheck-ac` takes the same selector as `--check-ac`, with the same effect inverted, here "a single one," just like the first `--check-ac` in this scenario.*

In the definition of done, you check off the one that's already finished by searching for
its text, not by number.

```console
$ biso set TASK-19 --check-dod "README"
TASK-19  In Progress  ac 5/6  dod 1/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: "README" doesn't match the key grammar (^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$), so it's treated as literal text and matches the element whose text contains it. Only #5 ("The README explains when and how many times it retries") matches, so there's no ambiguity.*

You add the attempt count to the log, check that it now shows up, and close the task. You
check off anything left, just in case you missed something, and write the summary.

```console
$ biso finish TASK-19 --check-ac all --check-dod all --append-summary "Retry with the same backoff as the download endpoint, capped at 3 attempts, attempt count visible in the log."
TASK-19  Done  ac 6/6  dod 2/2  urgency 0.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [`biso finish`](spec/cmd/verbos-del-ciclo.md#biso-finish), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `--check-ac all` checks #2 again, the one you'd unchecked, and doesn't touch the other five, they were already checked, and checking the same criterion twice "stays checked, with no warning, the operation is idempotent." `--check-dod all` checks #6, the only one left. Urgency drops to 0.0 because the terminal state fixes it there by definition, not because it gets recalculated with the usual formula. What this line doesn't say, and is worth knowing, is that closing a task always clears `leaseExpiresAt` and `leaseHolder`, whoever is holding it. TASK-19 had been carrying @claude's lease since scenario 5; after this `biso finish` it's left without a lease, and it would end up just as empty even if another identity had been holding it. The invariant always wins over the general rule that "a write by someone else doesn't touch the lease": a finished task with a live lease is a state that not even importing the board would accept.*

## 11. You get it wrong {: #escenario-11 }

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

*(derived output, see [El algoritmo de coincidencia](spec/vocabularios.md#el-algoritmo-de-coincidencia); not literal spec text)*

*Note: `TO_DO` doesn't appear as-is in the board's configuration, which stores `To Do`. But the matching algorithm lowercases both values and strips spaces, hyphens and underscores before comparing, so `TO_DO` normalizes to `todo`, same as `To Do`. The board hasn't changed since the previous chapter ended: there are still 54 tasks in `To Do`. `--count` prints just that number, with no other line.*

You try again, now with a hyphen and mixed case.

```console
$ biso ls --count -s To-Do
54
```

Exit code: `0`

*(derived output, see [El algoritmo de coincidencia](spec/vocabularios.md#el-algoritmo-de-coincidencia); not literal spec text)*

*Note: Same result, because it's the same filter: `To-Do` also normalizes to `todo`. The same rule also collapses repeated spaces, so `to do`, with two spaces in the middle, would have worked too (it's the literal example from the matching algorithm's own table). Four different ways to type it, one single value.*

You've forgotten the number of the retry task, so you type the first thing you remember
from its title.

```console
$ biso get "retry"
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-19  Done         task  high    Retry the upload on 5xx                           ac 6/6  @claude  -
```

Exit code: `5`

*(derived output, see [La búsqueda por texto](spec/referencias.md#la-búsqueda-por-texto); not literal spec text)*

*Note: All three tasks have «retry» in the title, so none of them wins by having it in the title while the others only have it in the body (the text search rule): all three are candidates equally. The program doesn't choose for you. Exit code 5 (AMBIGUOUS), and the three rows go to stdout, in the same column format as `biso ls`, sorted by urgency because none of them carries an `ordinal`: TASK-60 has an open question, so it doesn't add the active-task term, but it's high priority; TASK-33 is medium priority; TASK-19 is already done, so its urgency is 0.0 (a task in a terminal state always has urgency 0.0) and it comes last. TASK-19's `ac 6/6` carries straight over from how chapter 10 closes it: the two original criteria plus the four it adds there, all checked.*

You type the number by hand, but an extra digit sneaks in.

```console
$ biso get TASK-99
error: TASK-99 has never existed on this board
note: the highest id ever assigned here is TASK-62
```

Exit code: `4`

*(derived output, see [Los tres mensajes de "no la encuentro"](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro); not literal spec text)*

*Note: `TASK-99` is well-formed as an identifier, so this isn't a usage error (code 2): it's an error because that identifier has never existed, code 4 (NOT_FOUND), with the `never_allocated` key. The text is the literal message for this case, with the highest-ever identifier swapped for this board's own, `TASK-62`, the task created back in chapter 2. There's a third «can't find it» message in the same family, for a task the board did get around to assigning and that's gone now (archived and then deleted): it shares code 4 but not the key, and there's no need to reproduce it here for the difference between the two to be clear. And there's a fourth stumble we haven't even touched, a malformed identifier like `TASK-1.1`, which is code 2 instead of 4, because there the problem isn't that the task doesn't exist but that what you typed doesn't have the shape of a reference. Four ways to get it wrong, four different codes: 2, 3, 4 and 5.*

## 12. You park something halfway {: #escenario-12 }

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

*(derived output, see [`biso archive`](spec/cmd/archive.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Read that line carefully, because it says two things that look incompatible: the task is still `In Progress` and it's `archived`. There's no contradiction. Archiving isn't a state and it doesn't touch the state: it's a separate flag that takes the task off the active board. If it were unarchived tomorrow, it would go back to `In Progress`, which is where it already was. And there's something the line doesn't say: the lease the agent had claimed back in scenario 6 has just been cleared. That always happens on archiving, whoever holds the lease and whether it's still alive or already expired, and the reason is that keeping it around would stash it somewhere nobody looks, because both the startup message and the listing hide archived tasks by default. If this task came back to the board three weeks from now with a lease still held by a session that had already died, that would be garbage wearing the costume of information. Urgency is still being calculated, 5.4, because the task is still in a state that isn't the terminal one. Archiving doesn't zero it out: finishing it does.*

It checks that it hasn't lost the task.

```console
$ biso ls --archived
TASK-52  In Progress  task  low  Document the release checklist  ac 0/1  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md), [`biso archive`](spec/cmd/archive.md); not literal spec text)*

*Note: There it is, intact, still assigned to the same person. The only thing it's lost is the lease. Notice that `--archived` is what brings it back into view: without that flag it shows up in neither `biso ls` nor `biso prime`, and that's what "leaving the active board" actually means. The task hasn't gone anywhere, it's just been moved out of the way.*

And to close the idea, it checks what's still in progress now.

```console
$ biso ls -s "In Progress"
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Three tasks where there used to be four, and `TASK-19` isn't in either list anymore because it got finished back in scenario 10. The board now tells the truth about what's actually being worked on, which is exactly what archiving something you're not going to touch is for.*

## 13. You touch twenty at once {: #escenario-13 }

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

*(derived output, see [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: One line per task, in the same format as when you touch a single one. There's no batch mode that changes the shape of the output, and no summary that replaces the detail: what you learned with one applies to two just the same. `TASK-44` was already low priority, so its urgency doesn't move. `TASK-61` drops from 4.5 to 1.4, and there's a hidden lesson in there about reading these numbers: the real drop is exactly 3.0, what medium priority was contributing, but the two figures you see are rounded to one decimal from 4.45 and 1.45, and those two roundings don't go the same direction. Urgency is for ordering a list, not for doing arithmetic on it.*

Now a longer list, with a typo hidden inside: one of the references doesn't exist.

```console
$ biso set TASK-7 TASK-33 TASK-99 --priority medium
error: TASK-99 has never existed on this board
note: the highest id ever assigned here is TASK-62
```

Exit code: `4`

*(derived output, see [Los tres mensajes de "no la encuentro"](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: And here's what matters: **`TASK-7` and `TASK-33` haven't been touched**. It's not that they got written and then undone: it's that nothing ever got written, because validating all three references happens before the first write. Think about what that saves you. If the batch applied task by task and stopped on the first failure, after this error you'd have to figure out how many had already been written so you wouldn't repeat them, and there's no way to do that reliably from a program. The contract here is simple: all of them or none of them.*

It fixes the typo and this time, before writing, asks for it to be validated without
touching anything.

```console
$ biso set TASK-7 TASK-33 --priority medium --dry-run
2 tasks would be updated, nothing was written (--dry-run)
```

Exit code: `0`

*(derived output, see [Banderas globales](spec/invocacion.md#banderas-globales), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: Exit code 0, so it would have worked. That code is the real answer, more than the text: a program calling `biso` doesn't need to read anything to know whether the batch is valid. The exact text of this line is a derivation. The specification gives the literal phrase for `biso new --from`'s batch ("242 tasks would be created, nothing was written (--dry-run)") and says `--dry-run` works on every command that writes, but it doesn't spell out the phrase for `biso set`. Here the same form was used with the verb swapped, which is the most likely choice, and it's noted in tutorial/lagunas/11-13.md.*

With validation green, it runs it for real.

```console
$ biso set TASK-7 TASK-33 --priority medium
TASK-7   To Do  ac 0/4  urgency 15.3
TASK-33  To Do  ac 1/3  urgency 4.3
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

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
