<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

!!! warning "Generated document"
    This page is generated automatically from `tutorial/conceptos.md`. Do not
    edit it by hand: any change is lost on the next generation. To regenerate
    it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

# Concepts

`biso` is the tool an automated agent, and the person working alongside it, use to run the tasks of
a project. Before looking at a single command, five concepts are worth having straight, because the
rest of the tutorial takes them for granted.

## The board

A board is the set of a project's tasks together with its configuration: what states exist, what
task types there are, what priorities can be used, and who, if anyone, the identity of whoever works
on it is. That configuration is not a side detail: it is what makes a value meaningful or not. A
board does not accept any text in any field, only the values of its own vocabulary.

From that comes the rule that governs everything else: a value the board does not recognize is
always an error, whether you are writing it or asking about it, and with the same strictness in both
cases. Asking about a state that board does not have does not return an empty list, it returns an
error. As a consequence, when a question does return an empty list it is because there really is
nothing that matches what was asked, and that is information, not a silent failure.

## States

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

## Acceptance criteria and definition of done

A task can carry two independent checklists, with the same shape: each item has text and can be
checked or not, and each item is born with its own numeric key that does not change even if other
items are removed from the list. They are two lists, not one, and each is counted separately.

The first are the **acceptance criteria**: how you know that particular task's work actually works.
The second is the **definition of done**: what has to be true before calling the task closed, beyond
whether the result works, such as someone else having reviewed it. Neither is mandatory, and a task
can reach its terminal state with unchecked items in either of the two, because `biso` warns about
that but does not block it by default.

## Urgency

Urgency is a number that is not stored anywhere: it is computed every time the task is read, from
data that is stored, such as priority, whether the task is active, whether it blocks or is blocked
by another task, whether it has a due date coming up, whether it has acceptance criteria, and how
long it has been open. That is why urgency is called **derived**: nobody writes it directly, and its
value right now may not be its value a minute from now even if nobody touched the task, simply
because time passed or some other task it depends on changed. A task in its terminal state always
has urgency zero, with no further calculation.

## Declared identity

A board can have an identity configured: the text that says who you are when you call `biso`. It is
not mandatory: a board works perfectly well with nobody declaring one. What changes is that some
operations need to know who you are to make sense, such as asking only for the tasks that are yours,
getting automatically assigned a task when you start it, signing a comment with your name by default,
or leaving an open question waiting for someone to answer. If you try any of those without a declared
identity, `biso` does not guess or assume anything: it says so as an explicit error. The rest of the
board, whatever does not need to know who you are, keeps working the same with or without a declared
identity.
