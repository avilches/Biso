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

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: Not even the header shows up. `biso get`'s rule is literal: with `--section`, an empty section doesn't get printed, and if no section is left to print, the whole output is empty. That's different from asking for the full record, where every empty section shows up anyway, marked `(empty)`. The flag changes the standard on purpose: without it, leaving a section out would blur "empty" with "not requested"; with it, you asked for exactly that section, so "empty" no longer needs saying.*

You write down the first thing that comes to mind: that the change should be reviewed by
someone else before it counts as done.

```console
$ biso set TASK-19 --add-dod "Someone else reviews the change before the task counts as done"
TASK-19  In Progress  ac 0/2  dod 0/1  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: `--add-dod` adds. It's the same verb you already know from `--add-ac`: it behaves the same way for any list field.*

And a second thing that needs to be clear before closing: that the retry behavior gets
documented somewhere, for whoever reads this later.

```console
$ biso set TASK-19 --add-dod "The retry behavior is documented in the README"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

You check how they turned out, with their keys.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #1 Someone else reviews the change before the task counts as done
- [ ] #2 The retry behavior is documented in the README
```

Exit code: `0`

*(derived output, see [`biso get`](../spec/cmd/get.md); not literal spec text)*

On second thought, the first one is redundant: having someone else review the change is
already team policy for any PR, not something specific to this task. You remove it.

```console
$ biso set TASK-19 --rm-dod 1
TASK-19  In Progress  ac 0/2  dod 0/1  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [Selectores de criterios](../spec/familias-de-flags.md#selectores-de-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

You check the list again to make sure what's left hasn't changed number.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #2 The retry behavior is documented in the README
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](../spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables), [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: This is the central point of the scenario. The surviving element is still `#2`, it hasn't become `#1`. The program fixes a key when the element is created and never moves it: that's why checking or removing by number is safe even when the list has lost elements along the way, and why a task can perfectly well have criteria `#2` and `#7` without that being anyone's mistake.*

A coworker reviews the wording and suggests rewriting the whole list at once, adding a
second entry about the tests that are needed.

```console
$ biso set TASK-19 --clear-dods --add-dod "The README explains when and how many times it retries" --add-dod "The test suite covers the 5xx case"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [Sustituir un campo que no tiene bandera de \"sustituir entera\"](../spec/familias-de-flags.md#sustituir-un-campo-que-no-tiene-bandera-de-sustituir-entera), [`biso set`](../spec/cmd/set.md); not literal spec text)*

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

*(derived output, see [Los criterios y sus claves estables](../spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables), [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso get`](../spec/cmd/get.md); not literal spec text)*

*Note: The keys are `#3` and `#4`, not `#1` and `#2`. This list's counter, inside this task, was already at 2 before anything got cleared, and clearing doesn't reset it: "the keys of earlier elements are never reused" holds even when the clear-and-add comes after an `--rm-`, not just right after creating the task.*

Almost right away you realize you're polishing the wording before the code is properly
settled. You'd rather leave it blank until the work is more mature.

```console
$ biso set TASK-19 --clear-dods
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md); not literal spec text)*

*Note: The list ends up empty, and that's why the `dod` chunk disappears from the status line: it only shows up "whenever the task has a definition of done," and an empty list doesn't count as having one. `ac` keeps showing because those two criteria haven't been touched.*

A while later, with the retry already working, you go back to the final list. It's the
same wording as before.

```console
$ biso set TASK-19 --add-dod "The README explains when and how many times it retries" --add-dod "The test suite covers the 5xx case"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista sin coma (criterios)](../spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios), [`biso set`](../spec/cmd/set.md), [La urgencia](../spec/modelo-de-datos/urgencia.md#la-urgencia); not literal spec text)*

*Note: `--add-dod` again. The keys of these two elements are `#5` and `#6`: the list's counter hasn't been reset either by the earlier `--rm-dod` or by the later `--clear-dods`, it "only ever grows," no matter what happens to the list's content.*
