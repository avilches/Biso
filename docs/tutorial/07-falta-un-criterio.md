<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
-->

# 7. Halfway through, a criterion is missing

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/`. Do not edit it by hand: any change is lost on the
    next generation. To regenerate it:

    ```
    uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
    ```

You've been at `TASK-19` for a while. You're noting down what you check, and you realize the
task never said anything about the change being reviewed, or about the new retry behavior
ending up in the documentation. You stop to write that down before going on.

You check first what the task already carries.

!!! abstract "What this scenario teaches"
    - A task has one checklist, not two. Whatever has to be true before closing, including what another tool would keep in a separate definition of done, is one more acceptance criterion.
    - Each field flag says what it does in its own name: `--add-` adds, `--rm-` removes, `--clear-` empties. For criteria there's no whole-list replace; doing that means clearing and adding in the same call.
    - The `#N` keys of a criterion get assigned when the element is created and are never reassigned. Removing one from the middle doesn't renumber the ones that are left.
    - The key counter only ever grows, even after `--clear-acs`. Elements added after emptying the list don't get back the keys of the ones that were removed.

You ask for just that section, so you don't pull the whole record for something you only
want to read one part of.

```console
$ biso get TASK-19 --section ac
TASK-19  Retry the upload on 5xx

## Acceptance Criteria
- [ ] #1 A request that gets a 5xx is retried automatically
- [ ] #2 The number of retries is configurable
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: These are the two criteria the task was born with, in the initial inventory. Neither says anything about review or documentation.*

You write down the first thing that comes to mind: that the change should be reviewed by
someone else before it counts as done.

```console
$ biso set TASK-19 --add-ac "Someone else reviews the change before the task counts as done"
TASK-19  In Progress  ac 0/3  urgency 11.0  added ac #3
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: Look at what just happened: "someone else reviews it" is not about whether the retry works, and another tool would file it under a separate definition of done. In `biso` there is no second list, so it goes here, as one more criterion. The line ends in `added ac #3` because the call created an element and you couldn't have known its key without reading the task first.*

And a second thing that needs to be clear before closing: that the retry behavior gets
documented somewhere, for whoever reads this later.

```console
$ biso set TASK-19 --add-ac "The retry behavior is documented in the README"
TASK-19  In Progress  ac 0/4  urgency 11.0  added ac #4
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

You check how they turned out, with their keys.

```console
$ biso get TASK-19 --section ac
TASK-19  Retry the upload on 5xx

## Acceptance Criteria
- [ ] #1 A request that gets a 5xx is retried automatically
- [ ] #2 The number of retries is configurable
- [ ] #3 Someone else reviews the change before the task counts as done
- [ ] #4 The retry behavior is documented in the README
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

On second thought, `#3` is redundant: having someone else review the change is already team
policy for any PR, not something specific to this task. You remove it.

```console
$ biso set TASK-19 --rm-ac 3
TASK-19  In Progress  ac 0/3  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

You check the list again to make sure what's left hasn't changed number.

```console
$ biso get TASK-19 --section ac
TASK-19  Retry the upload on 5xx

## Acceptance Criteria
- [ ] #1 A request that gets a 5xx is retried automatically
- [ ] #2 The number of retries is configurable
- [ ] #4 The retry behavior is documented in the README
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](../spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables), [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: This is the central point of the scenario. The surviving element is still `#4`, it hasn't become `#3`. The program fixes a key when the element is created and never moves it: that's why checking or removing by number is safe even when the list has lost elements along the way, and why a task can perfectly well have criteria `#2` and `#7` without that being anyone's mistake.*

A coworker reviews the wording and finds all three too vague to tell whether they're met.
Between you, you rewrite the whole list at once.

```console
$ biso set TASK-19 --clear-acs --add-ac "The README explains when and how many times it retries" --add-ac "The test suite covers the 5xx case"
TASK-19  In Progress  ac 0/2  urgency 11.0  added ac #5, #6
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [Sustituir un campo que no tiene flag de \"sustituir entera\"](../spec/familias-de-flags.md#sustituir-un-campo-que-no-tiene-flag-de-sustituir-entera), [`biso set`](../spec/cmd/set.md); not literal spec text)*

*Note: There's no flag that replaces the whole list of criteria in one shot, so this clears it and adds the two new ones in the same call. The three old elements disappear, and the two that come in are brand new. No warning shows up here, unlike a `--replace-*` on a comma-separated list field that overwrites non-empty content: clearing was explicit, you typed it yourself, so nothing gets overwritten silently and there's nothing to warn about.*

You check the keys of the two new ones.

```console
$ biso get TASK-19 --section ac
TASK-19  Retry the upload on 5xx

## Acceptance Criteria
- [ ] #5 The README explains when and how many times it retries
- [ ] #6 The test suite covers the 5xx case
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](../spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables), [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: The keys are `#5` and `#6`, not `#1` and `#2`. This task's counter was already at 4 before anything got cleared, and clearing doesn't reset it: "the keys of earlier elements are never reused" holds even when the clear-and-add comes after an `--rm-ac`, not just right after creating the task.*

Almost right away you realize you're polishing the wording before the code is properly
settled. You'd rather leave the list blank until the work is more mature.

```console
$ biso set TASK-19 --clear-acs
TASK-19  In Progress  urgency 10.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: Two things change in that line. The `ac` chunk disappears, because it only shows up when the task has criteria and an empty list doesn't count as having any. And the urgency drops from 11.0 to 10.0, because having at least one acceptance criterion is worth 1.0 in the formula: a task nobody can tell is finished is, by that much, less ready to be picked up.*

You ask for the section again, to see what an empty list looks like.

```console
$ biso get TASK-19 --section ac
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: Not even the header shows up. `biso get`'s rule is literal: with `--section`, an empty section doesn't get printed, and if no section is left to print, the whole output is empty. That's different from asking for the full record, where every empty section shows up anyway, marked `(empty)`. The flag changes the standard on purpose: without it, leaving a section out would blur "empty" with "not requested"; with it, you asked for exactly that section, so "empty" no longer needs saying.*

A while later, with the retry already working, you go back to the list. It's the same
wording you and your coworker had agreed on.

```console
$ biso set TASK-19 --add-ac "The README explains when and how many times it retries" --add-ac "The test suite covers the 5xx case"
TASK-19  In Progress  ac 0/2  urgency 11.0  added ac #7, #8
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: The keys of these two elements are `#7` and `#8`: the counter hasn't been reset either by the earlier `--rm-ac` or by the two `--clear-acs`, it "only ever grows," no matter what happens to the list's content. The urgency goes back up to 11.0 for the same reason it went down: the task has criteria again.*
