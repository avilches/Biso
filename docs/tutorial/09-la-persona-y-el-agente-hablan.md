<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 9. The person and the agent talk

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

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

*(derived output, see [`biso note`](../spec/cmd/verbos-del-ciclo.md#biso-note), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

You check how that note turned out.

```console
$ biso get TASK-19 --section notes
TASK-19  Retry the upload on 5xx

## Implementation Notes
The exponential backoff in retry.go can be reused as-is; the limit just needs to be set to 3.
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

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

*(derived output, see [`biso comment`](../spec/cmd/verbos-del-ciclo.md#biso-comment), [Los comentarios](../spec/modelo-de-datos/comentarios.md#los-comentarios), [Comentarios](../spec/familias-de-banderas.md#comentarios); not literal spec text)*

*Note: "note: comment #3 by @trello:juan" goes to stderr, in the exact shape the specification's own example gives. The number is a stable key, exactly like a criterion's: it's assigned once when the comment is created and never reassigned, even if an earlier comment is later removed with `--rm-comment`. A comment's body and author are never edited, by any flag; only its date can be corrected, with `--set-comment-date`, and the whole comment can be removed entirely. This is TASK-19's third comment because `biso answer` had already added two in the previous scenario.*

You answer yourself, as yourself, with no author flags: the configured identity signs by
default.

```console
$ biso comment TASK-19 "Not for now, 3 retries is the policy for the rest of the system; if it happens again we'll revisit it."
note: comment #4 by @claude
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso comment`](../spec/cmd/verbos-del-ciclo.md#biso-comment), [Los comentarios](../spec/modelo-de-datos/comentarios.md#los-comentarios); not literal spec text)*

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

*(derived output, see [`biso get`](../spec/cmd/get.md), [Los comentarios](../spec/modelo-de-datos/comentarios.md#los-comentarios); not literal spec text)*

*Note: Notes are a single block; comments are four signed entries, in the order they were written. `## Implementation Notes` comes before `## Comments` even though the flag asked for them in the opposite order (`notes,comments`): the print order of sections is the fixed order of the full record, not the order they're listed in `--section`. The specification doesn't say this explicitly for the case of several sections at once; noted as an open gap in tutorial/lagunas/07-10.md.*
