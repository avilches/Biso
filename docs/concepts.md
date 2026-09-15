# Concepts

`biso` is the tool an automated agent, and the person working alongside it, use to run the tasks of
a project. Before looking at a single command, five concepts are worth having straight, because the
rest of the tutorial takes them for granted.

## The board

A board is the set of a project's tasks together with its configuration: what states exist, what
task types there are, and what priorities can be used (the full list of keys is in
["Las claves" of `biso config`](spec/cmd/config.md#las-claves)). That configuration is not a side detail: it is what makes a value meaningful or not. A
board does not accept any text in any field, only the values of its own vocabulary.

From that comes the rule that governs everything else: a value the board does not recognize is
always an error, whether you are writing it or asking about it, and with the same strictness in both
cases. Asking about a state that board does not have does not return an empty list, it returns an
error. As a consequence, when a question does return an empty list it is because there really is
nothing that matches what was asked, and that is information, not a silent failure. The exact rule is
in ["Los vocabularios del tablero y la regla de validación"](spec/vocabularios.md).

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
used for it is not. For example:

| Board | `statuses` | initial | active | terminal |
|---|---|---|---|---|
| A new board, with the defaults | `To Do, In Progress, Done` | `To Do` | `In Progress` | `Done` |
| A team that triages first | `Inbox, Ready, Doing, Review, Shipped` | `Inbox` | `Doing` | `Shipped` |
| A board in Spanish | `Pendiente, En curso, Hecha` | `Pendiente` | `En curso` | `Hecha` |

In the second board, `Ready` and `Review` have no role: a task gets there only when someone
sets it explicitly with `biso set`, because no command writes them on its own. `biso start` always writes the active state, and `biso finish` always writes
the terminal one, whatever each board calls them. The keys that hold each role are `initial_status`,
`active_status` and `terminal_status` (["Las claves" of `biso config`](spec/cmd/config.md#las-claves)),
and why these three roles and no others is explained in
["El modelo de estados"](decisiones/modelo-de-estados.md).

State and assignment answer two different questions, and an agent looking for work reads both. The
state says where a task stands in its own life, from born to done; `assignees` says who was handed
it, and that can happen at any point along the way, a task fresh in its initial state included. A
task created in `Inbox` and assigned to an agent is already that agent's, whether or not anyone
moves it to `Ready` first: `--mine` finds it regardless of state, and only running `biso start`
switches it to the active state. There is no extra "available" role for this, because assignment
already carries the "this is yours" signal on its own; adding a state for it would only duplicate
what `assignees` already says. The reasoning behind treating assignment and state as two independent
signals is in
["Distinguir el encargo de la ejecución"](decisiones/modelo-de-estados.md#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo).

## Acceptance criteria and definition of done

A task can carry two independent checklists, with the same shape: each item has text and can be
checked or not, and each item is born with its own numeric key that does not change even if other
items are removed from the list. They are two lists, not one, and each is counted separately.

The first are the **acceptance criteria**: how you know that particular task's work actually works.
The second is the **definition of done**: what has to be true before calling the task closed, beyond
whether the result works, such as someone else having reviewed it. Neither is mandatory, and a task
can reach its terminal state with unchecked items in either of the two, because `biso` warns about
that but does not block it by default. Both lists, and the stable keys of their items, are defined
in ["Los criterios y sus claves estables"](spec/modelo-de-datos/criterios.md).

## Urgency

Urgency is a number that is not stored anywhere: it is computed every time the task is read, from
data that is stored, such as priority, whether the task is active, whether it blocks or is blocked
by another task, whether it has a due date coming up, whether it has acceptance criteria, and how
long it has been open. That is why urgency is called **derived**: nobody writes it directly, and its
value right now may not be its value a minute from now even if nobody touched the task, simply
because time passed or some other task it depends on changed. A task in its terminal state always
has urgency zero, with no further calculation. The formula, with the weight of each term, is in
["La urgencia"](spec/modelo-de-datos/urgencia.md#la-urgencia).

## Declared identity

You can declare an identity: the text that says who you are when you call `biso`, such as `@sara`.
It does not belong to any board: it comes from the `BISO_ME` environment variable or, if that is not
set, from the `me` key of the machine configuration file `~/.biso/config.json`, so the same identity
applies to every board on that machine. It is not mandatory: a board works perfectly well with nobody
declaring one. What changes is that some
operations need to know who you are to make sense, such as asking only for the tasks that are yours,
getting automatically assigned a task when you start it, signing a comment with your name by default,
or leaving an open question waiting for someone to answer. If you try any of those without a declared
identity, `biso` does not guess or assume anything: it says so as an explicit error. The rest of the
board, whatever does not need to know who you are, keeps working the same with or without a declared
identity. Which commands need it, and the exact error each one gives without it, is in
["Variables de entorno"](spec/invocacion.md#variables-de-entorno).
