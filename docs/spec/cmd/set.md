# `biso set`

El comando de edición general. Todo lo que hacen los verbos de flujo de la sección ["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](verbos-del-ciclo.md) se puede hacer aquí, con
más palabras.

## Firma

```
biso set <ref>... [cualquier flag de campo de las familias de flags]
         [--check-ac <sel>]... [--uncheck-ac <sel>]...
         [--comment <text>]... [--comment-author <@who>]
         [--rm-comment <sel>]... [--set-comment-date <sel>=<instante>]... [--id] [--match]
```

## Parámetros propios

**Todos** los flags de las secciones ["Campos de lista que admiten coma"](../familias-de-flags.md#campos-de-lista-que-admiten-coma), ["Campos de lista sin coma (criterios)"](../familias-de-flags.md#campos-de-lista-sin-coma-criterios), ["Campos de prosa"](../familias-de-flags.md#campos-de-prosa) y ["Campos escalares"](../familias-de-flags.md#campos-escalares) valen aquí, con exactamente el mismo
significado que en cualquier otro comando. Lo propio de `set`:

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `<ref>` | sí, una o más | referencia | | sí | no | |
| `--check-ac <sel>` | no | selector ["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios) | | sí | ver ["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios) | solape con `--uncheck-ac` |
| `--uncheck-ac <sel>` | no | selector ["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios) | | sí | ver ["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios) | solape con `--check-ac` |
| `--comment <text>` | no | texto largo | | sí | no | |
| `--comment-author <@who>` | no | texto libre | `me` | no | no | requiere `--comment` |
| `--rm-comment <sel>` | no | selector ["Comentarios"](../familias-de-flags.md#comentarios) | | sí | ver ["Comentarios"](../familias-de-flags.md#comentarios) | solape con `--set-comment-date` sobre la misma clave |
| `--set-comment-date <sel>=<instante>` | no | selector ["Comentarios"](../familias-de-flags.md#comentarios) + instante UTC | | sí | ver ["Comentarios"](../familias-de-flags.md#comentarios) | solape con `--rm-comment` sobre la misma clave |
| `--id` | no | booleano | falso | no | no | `--match` |
| `--match` | no | booleano | falso | no | no | `--id` |

**El autor de un comentario se llama `--comment-author` en todos los comandos que lo aceptan**, sin
excepción, aunque en `biso comment` el prefijo parezca redundante. Un concepto, un nombre.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Ninguna referencia | Error 2, con el `code` `missing_ref`: `error: biso set needs at least one task reference`, y `hint: biso set MYP-11 --priority high` |
| Ningún flag de cambio | Error 2: `error: nothing to change` con un puntero a `biso get`, que nombra la primera referencia de la llamada: `hint: \`biso get MYP-11\` shows what the task has now` |
| La misma tarea nombrada dos veces | Se escribe una sola vez. La segunda mención no dice nada que no dijera la primera, y aplicar los cambios dos veces añadiría dos veces el mismo comentario o el mismo criterio |
| Varias referencias | El mismo cambio se aplica a todas, con la garantía de todo o nada de la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables) |
| Varias referencias y un selector que no sea `all` (clave, rango, lista o texto) | Error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |
| Varias referencias y `--check-ac all` | Válido |
| Una de varias referencias no existe | Error 4, y **no se escribe ninguna**, ni siquiera las buenas |
| Un `--replace-*` sustituye una lista no vacía | Se hace, con el aviso de la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) diciendo cuántos elementos ha reemplazado |
| `--add-labels k::v` sobre una tarea que conserva otras etiquetas de la clave `k` | Se hace: `k::v` queda como única etiqueta de esa clave y el aviso nombra las que quitó (["Escribir una etiqueta con ámbito"](../familias-de-flags.md#escribir-una-etiqueta-con-ámbito)) |
| `--add-labels k:v` sobre una tarea que conserva un `k::x` | Error 6 con el `code` `exclusive_label_conflict`, y no se escribe nada, ni en esa tarea ni en las demás de la llamada |
| `k:a` y `k::b` de la misma clave en la misma llamada | Error 2 con el `code` `mixed_label_separators`, sin mirar el orden en que se escribieron |
| Paso a un estado terminal con criterios sin marcar | Se hace, con aviso |
| Paso a un estado terminal con una pregunta abierta (["La pregunta abierta"](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta)) | Se hace, con aviso, igual que en `biso finish` (["`biso finish`"](verbos-del-ciclo.md#biso-finish)) y como atribuye la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) a cualquier llegada al estado terminal |
| Todos los flags dejan la tarea igual | Código 0, con `note: MYP-11 unchanged`. Ningún campo de la tarea se escribe, `updatedAt` no cambia y `changed` sale vacía, pero si quien llama es `leaseHolder` **el arrendamiento se renueva igual**: es una escritura del tenedor sobre su tarea, y el latido no depende de si los valores coincidían (["La renovación"](../lease.md#la-renovación) de `lease.md`) |
| `--status` a un estado que no es el activo, `--clear-assignees` o `--rm-assignees` que deja la tarea sin nadie, sobre una tarea con arrendamiento | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea el arrendamiento; si era de otra identidad, sale además el aviso de la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) (["El vaciado"](../lease.md#el-vaciado) de `lease.md`) |
| `--above` o `--below` con varias referencias | Válido, y es la forma de mover un bloque: las tareas caen en el hueco en el orden en que se escribieron (["El orden manual"](../familias-de-flags.md#el-orden-manual)) |
| `--above` o `--below` nombrando a una de las tareas que se están moviendo | Error 2, con el `code` `self_ordinal_neighbour`: `error: --below: MYP-11 cannot be its own neighbour` |
| `--above` o `--below` sobre una tarea sin clave de orden | Error 6, con el `code` `neighbour_without_ordinal`, y **no se escribe ninguna** de las tareas de la llamada |
| Dos de `--ordinal`, `--above`, `--below` y `--clear-ordinal` en la misma llamada | Error 2, con el `code` `incompatible_flags`: los cuatro escriben el mismo campo |
| `--comment-author` sin `--comment` | Error 2 |
| `--comment` sin `--comment-author` y sin ninguna identidad configurada (["Variables de entorno"](../invocacion.md#variables-de-entorno)) | Error 2 |
| La tarea no se puede leer, por un valor de `status`, `type` o `priority` que el tablero no declara, una fecha malformada o cualquier otro motivo de ["Qué se comprueba"](../garantias.md#qué-se-comprueba) | Error 3 con `code` `undecodable_task`, y no se escribe nada. **La excepción es el valor fuera de vocabulario cuando la propia llamada lo corrige**: `biso set MYP-11 --priority medium` sobre una tarea cuya prioridad el tablero no declara se aplica, porque se juzga la tarea como quedaría, y `biso set MYP-11 --title X` no, porque seguiría ilegible. Una fecha malformada no tiene excepción, porque esa tarea no llega a cargarse. Lo mismo vale para los verbos del ciclo que fijan el estado, `biso start` y `biso finish`, y con `--dry-run` la vista previa contesta lo mismo que la llamada real ("Cómo se arregla una tarea ilegible" en ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#cómo-se-arregla-una-tarea-ilegible)) |
| La tarea está archivada | Se hace, con `note: MYP-11 is archived` por stderr, igual que `biso get`. `biso set` no reclama ningún arrendamiento, así que archivar no le opone ninguna restricción; la única excepción del programa es [`biso start`](verbos-del-ciclo.md#biso-start), que sí lo reclama (["El vaciado"](../lease.md#el-vaciado) de `lease.md`) |

## Salida

Por defecto, **una línea por tarea afectada** con lo que quien llama no sabía: el estado resultante,
el avance de criterios y la urgencia recalculada. Los tres datos son derivados, y ninguno se puede
conocer sin leer la tarea:

```
MYP-11  In Progress  ac 1/2  urgency 19.0
```

**Esta es la línea de estado, y la imprimen también los seis verbos del ciclo de la sección ["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](verbos-del-ciclo.md) y
`biso archive`.** La única excepción es `biso new`, por el motivo que da la sección ["`biso new`"](new.md). Cada comando la enseña
con su propio ejemplo, pero las reglas de su forma se dicen aquí y no se repiten:

1. El trozo `ac <marcados>/<total>` sale siempre que la tarea tenga criterios de aceptación. Una
   tarea sin ninguno imprime solo el identificador, el estado y la urgencia.
2. **Si la llamada crea uno o más criterios de aceptación, la línea añade `added ac #<clave>`,** con
   las claves nuevas separadas por comas si son varias. **Si la llamada no crea ningún criterio, esta
   pieza no aparece**, y la línea es exactamente la de antes. La razón de que vaya aquí y no en
   `biso new` está en
   ["Campos de lista sin coma (criterios)"](../familias-de-flags.md#campos-de-lista-sin-coma-criterios):
   ```
   biso set MYP-11 --add-ac "There is a test" --add-ac "Docs updated"
   MYP-11  In Progress  ac 1/4  urgency 19.0  added ac #4, #5
   ```
3. La palabra `archived` cierra la línea cuando la tarea queda archivada, y solo entonces. Es lo único
   más que una escritura puede añadirle, y quien lo añade es `biso archive` (["`biso archive`"](archive.md)).
4. **Con `--dry-run`, la salida es la misma línea de estado que daría la escritura real, calculada sin
   escribir nada**, con un encabezado delante que la marca como hipotética: `<N> task(s) would be
   affected (--dry-run)`, en singular si `N` es 1 y en plural en cualquier otro caso. Debajo va una
   línea por tarea, con exactamente la misma forma de los puntos 1 a 3 (incluido `added ac #<clave>` si
   la llamada crearía un criterio, y `archived` si la tarea quedaría archivada). Los dos bloques van por
   stdout, en el mismo canal que la línea real, porque siguen siendo el dato que la llamada produce, solo
   que hipotético; ninguno de los dos implica que se haya escrito nada. La regla vale para los seis
   verbos del ciclo (sección ["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`,
   `answer`"](verbos-del-ciclo.md)) y para `biso archive` (sección ["`biso archive`"](archive.md)),
   porque todos reusan esta misma línea de estado; cada uno la enseña con su propio ejemplo.
   ```
   biso set MYP-11 --add-labels parser --dry-run
   1 task would be affected (--dry-run)
   MYP-11  In Progress  ac 1/2  urgency 19.0
   ```

Los avisos van por stderr:

```
warning: --replace-labels replaced 2 existing labels
```

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.write",
  "generatedAt": "2026-09-06T11:40:18Z",
  "data": {
    "tasks": [
      { "id": "MYP-11", "status": "In Progress", "acDone": 1, "acTotal": 2,
        "urgency": 19.0,
        "changed": ["labels", "status"], "acAdded": [] }
    ],
    "warnings": [ { "code": "overwrite", "field": "labels", "count": 2, "task": "MYP-11" } ]
  }
}
```

`kind` es `task.write` para `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer` y
`archive`, para que quien consuma la salida no tenga que distinguir qué verbo la produjo. `changed`
dice qué campos han cambiado de verdad, que no es lo mismo que qué flags se han pasado: se calcula
comparando la tarea de antes con la de después, campo a campo, con los nombres de
["El modelo de datos de una tarea"](../modelo-de-datos/index.md), y en el orden de la tabla de esa
página. **`leaseExpiresAt` y `leaseHolder` no están nunca entre ellos**, porque una renovación no es
un cambio de la tarea, que es lo mismo que dice la fila de `updatedAt` de esa tabla y lo que hace
cierta la fila de "todos los flags dejan la tarea igual" de más arriba. Una lista vacía y una lista
que nunca estuvo son el mismo campo sin valor, así que pasar de una a otra tampoco cuenta como
cambio. En el lote
de `new --from`, las 242 tareas van en `data.tasks` de **un solo sobre**, no en 242 objetos sueltos.

**`acAdded` es la lista de claves que la llamada acaba de crear**, en el mismo orden en que se
crearon, vacía (`[]`) cuando no se creó ninguna. Está presente en las mismas condiciones que el
resto de claves de `data.tasks` (["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias)): siempre, en todos los comandos que
comparten este `kind`, incluido `new`, aunque su equivalente en texto plano solo aparezca en la línea
de estado y nunca en la salida por defecto de `biso new` (["`biso new`"](new.md)).

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Cambio aplicado, o nada que cambiar | 0 |
| Sin flags de cambio, flags incompatibles, selector por clave con varias tareas, solape | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Alguna referencia no existe, o un selector de texto no encaja con ningún criterio o comentario | 4 |
| Alguna referencia de texto encaja con varias tareas, o un selector con varios criterios o comentarios | 5 |
| Una etiqueta con ámbito escrita con `:` sobre una clave que la tarea conserva con `::` (["Escribir una etiqueta con ámbito"](../familias-de-flags.md#escribir-una-etiqueta-con-ámbito)) | 6 |
| La vecina de `--above` o de `--below` no tiene clave de orden | 6 |
| El almacén falla, o no se obtiene el acceso exclusivo | 8 |
| No hay tablero | 20 |

## `biso set --help`

La ayuda glosa los campos de relación (`--parent`, `--add-deps`, `--add-refs`) y dice hacia dónde
apunta una dependencia, con `biso set MYP-10 --add-deps MYP-4` como ejemplo de que `MYP-4` bloquea a
`MYP-10`. Junto con la regla 11 del mensaje de arranque (["La salida literal"](prime.md#la-salida-literal)), es lo que enseña la dirección, y no puede deducirse del
nombre del flag: una dependencia escrita al revés es válida y el programa no la detecta. La
decisión, y por qué no hay un flag inverso, está en
["La ayuda enseña la dirección de una dependencia"](../../decisiones/detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia).

```
Usage: biso set <ref>... [options]

Change any field of one or more tasks, all or nothing. Every flag here means
the same in `biso new`, `biso start`, `biso note`, `biso comment`, `biso ask`,
`biso answer`, `biso finish` and `biso archive`. A flag name says what it does,
and comma lists always have the same four shapes, but what a relation field
means is not in its name: see Relations below.

List fields that take comma-separated values have four shapes, and there is no
field that breaks them:
  --add-labels X      add one or more       --replace-labels X   replace the whole list
  --rm-labels X       remove one or more    --clear-labels       empty the list
The same works for --assignees, --refs and --deps.

A label with a colon is scoped: key:value allows several values of that key
on a task, key::value at most one, and writing key::value drops the other
values of that key and says on stderr which ones it dropped. Filter with
`biso ls --label key:` to get any value of a key.

Criteria have three, because a criterion's text can contain a comma and so is
never split on one. There is no whole-list replace; do it by clearing and
adding in the same call.
      --add-ac <text>        add a criterion; repeatable
      --rm-ac <sel>          remove by selector; sel is all, 3, 1-4, 1,3,7 or
                             the criterion text. The numbers are stable #N
                             keys. With several tasks, sel has to be all
      --clear-acs            empty the list
      --check-ac <sel>       check criteria, by the same kind of selector
      --uncheck-ac <sel>     the opposite

Prose fields have two, because a block of text has no single item to remove.
Replace by clearing and appending in the same call.
      --append-desc X        append a paragraph
      --append-plan X
      --append-note X
      --append-summary X
      --clear-desc / --clear-plan / --clear-notes / --clear-summary

Scalars just take a value: --title, --status, --type, --priority,
--parent, --due, --author. Each has a --clear-<field>. An empty string
is never a way to clear anything.

Manual order is the one scalar you never type. You write where the task
goes, and biso writes the key:
      --ordinal first|last   before or after every task that has a place
      --above <ref>          just above that task
      --below <ref>          just below it
      --clear-ordinal        out of the manual order
Several tasks in one call land in the order you wrote them, so
`biso set MYP-7 MYP-19 --below MYP-40` leaves MYP-40, MYP-7, MYP-19. A
task with no place cannot be a neighbour: give it one first with
`biso set <ref> --ordinal last`.

Relations point from the task you name in <ref> to other tasks:
      --parent <ref>         the task this one is part of; at most one
      --add-deps <ref>       tasks that must be done before this one, so each
                             blocks it; checked to exist, no cycles
      --add-refs <text>      a path, a URL or a task id to look at; free text,
                             never checked
A dependency is written on the task that waits, never on the one that blocks.
To say that MYP-4 blocks MYP-10:
  biso set MYP-10 --add-deps MYP-4
The other way round, `biso set MYP-4 --add-deps MYP-10`, is just as valid and
says the opposite, so biso cannot warn you when it is backwards.

Comments:
      --comment <text>            append a comment; repeatable
      --comment-author <@w>       who wrote it (default: you)
      --rm-comment <sel>          remove one or more, by the same kind of
                                   selector as --rm-ac; body and author are
                                   never edited, by any flag
      --set-comment-date <sel>=<instant>
                                   correct only the date of one or more,
                                   instant is YYYY-MM-DDTHH:MM:SSZ

Resolution:
      --id / --match         force <ref> to be an id, or free text

Within one call, every --rm-*/--clear-* is applied before every --add-*/
--append-*, regardless of the order they were written in. A --replace-* over
a non-empty list is allowed and warns on stderr with how many items it
replaced.

Exit codes:
  0  done                    5  something matched more than one thing
  2  bad usage               6  a scoped label already has its one value,
  3  unknown value              or --above/--below on a task with no place
  4  a task, criterion or comment was not found
                             8  the board could not be written
                             20 no board here

Examples:
  biso set MYP-11 --priority high --add-labels parser
  biso set MYP-11 --check-ac 1,3 --append-note "Both covered by diff_test.rs"
  biso set MYP-11 MYP-12 --due 2026-09-20
  biso set "CRLF" --clear-desc --append-desc @docs/bugs/BUG-02.md
```

---

