# `biso set`

El comando de edición general. Todo lo que hacen los verbos de flujo de la sección ["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](verbos-del-ciclo.md) se puede hacer aquí, con
más palabras.

## Firma

```
biso set <ref>... [cualquier bandera de campo de las familias de banderas]
         [--check-ac <sel>]... [--uncheck-ac <sel>]... [--check-dod <sel>]... [--uncheck-dod <sel>]...
         [--comment <text>]... [--comment-author <@who>] [--id] [--match]
```

## Parámetros propios

**Todas** las banderas de las secciones ["Campos de lista que admiten coma"](../familias-de-banderas.md#campos-de-lista-que-admiten-coma), ["Campos de lista sin coma (criterios)"](../familias-de-banderas.md#campos-de-lista-sin-coma-criterios), ["Campos de prosa"](../familias-de-banderas.md#campos-de-prosa), ["Campos escalares"](../familias-de-banderas.md#campos-escalares) y ["Campos externos"](../familias-de-banderas.md#campos-externos) valen aquí, con exactamente el mismo
significado que en cualquier otro comando. Lo propio de `set`:

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--check-ac <sel>` | | no | selector ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | | sí | ver ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | solape con `--uncheck-ac` |
| `--uncheck-ac <sel>` | | no | selector ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | | sí | ver ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | solape con `--check-ac` |
| `--check-dod <sel>` | | no | selector ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | | sí | ver ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | solape con `--uncheck-dod` |
| `--uncheck-dod <sel>` | | no | selector ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | | sí | ver ["Selectores de criterios"](../familias-de-banderas.md#selectores-de-criterios) | solape con `--check-dod` |
| `--comment <text>` | | no | texto largo | | sí | no | |
| `--comment-author <@who>` | | no | texto libre | `me` | no | no | requiere `--comment` |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

**El autor de un comentario se llama `--comment-author` en todos los comandos que lo aceptan**, sin
excepción, aunque en `biso comment` el prefijo parezca redundante. Un concepto, un nombre.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Ninguna bandera de cambio | Error 2: `error: nothing to change` con un puntero a `biso get` |
| Varias referencias | El mismo cambio se aplica a todas, con la garantía de todo o nada de la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables) |
| Varias referencias y un selector que no sea `all` (clave, rango, lista o texto) | Error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |
| Varias referencias y `--check-ac all` | Válido |
| Una de varias referencias no existe | Error 4, y **no se escribe ninguna**, ni siquiera las buenas |
| Un `--replace-*` sustituye una lista no vacía | Se hace, con el aviso de la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) diciendo cuántos elementos ha reemplazado |
| Paso a un estado terminal con criterios sin marcar | Se hace, con aviso |
| Paso a un estado terminal con una pregunta abierta (["La pregunta abierta"](../modelo-de-datos.md#la-pregunta-abierta)) | Se hace, con aviso, igual que en `biso finish` (["`biso finish`"](verbos-del-ciclo.md#biso-finish)) y como atribuye la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) a cualquier llegada al estado terminal |
| Todas las banderas dejan la tarea igual | Código 0, con `note: MYP-11 unchanged`. Ningún campo de la tarea se escribe, `updatedAt` no cambia y `changed` sale vacía, pero si quien llama es `leaseHolder` **el arrendamiento se renueva igual**: es una escritura del tenedor sobre su tarea, y el latido no depende de si los valores coincidían (sexta precisión de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md)) |
| `--status` a un estado que no es el activo, `--clear-assignees` o `--rm-assignees` que deja la tarea sin nadie, sobre una tarea con arrendamiento | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea el arrendamiento; si era de otra identidad, sale además el aviso de la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) (séptima precisión de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md)) |
| `--comment-author` sin `--comment` | Error 2 |
| `--comment` sin `--comment-author` y sin ninguna identidad configurada (["Variables de entorno"](../invocacion.md#variables-de-entorno)) | Error 2 |
| La tarea no se puede leer | Error 3, y no se escribe nada |

## Salida

Por defecto, **una línea por tarea afectada** con lo que quien llama no sabía: el estado resultante,
el avance de criterios y la urgencia recalculada. Los tres datos son derivados, y ninguno se puede
conocer sin leer la tarea. La MYP-11 de los ejemplos tiene además una definición de hecho de un
elemento, así que su línea trae también el avance de esa segunda lista:

```
MYP-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

**Esta es la línea de estado, y la imprimen también los seis verbos del ciclo de la sección ["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](verbos-del-ciclo.md) y
`biso archive`.** La única excepción es `biso new`, por el motivo que da la sección ["`biso new`"](new.md). Cada comando la enseña
con su propio ejemplo, pero las reglas de su forma se dicen aquí y no se repiten:

1. El trozo `ac <marcados>/<total>` sale siempre que la tarea tenga criterios de aceptación.
2. El trozo `dod <marcados>/<total>` sale siempre que tenga definición de hecho. Una tarea sin ninguna
   de las dos listas imprime solo el identificador, el estado y la urgencia.
3. La palabra `archived` cierra la línea cuando la tarea queda archivada, y solo entonces. Es lo único
   que un comando puede añadirle, y quien lo añade es `biso archive` (["`biso archive`"](archive.md)).

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
        "dodDone": 0, "dodTotal": 1, "urgency": 19.0,
        "changed": ["labels", "status"] }
    ],
    "warnings": [ { "code": "overwrite", "field": "labels", "count": 2, "task": "MYP-11" } ]
  }
}
```

`kind` es `task.write` para `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer` y
`archive`, para que quien consuma la salida no tenga que distinguir qué verbo la produjo. `changed`
dice qué campos han cambiado de verdad, que no es lo mismo que qué banderas se han pasado. En el lote
de `new --from`, las 242 tareas van en `data.tasks` de **un solo sobre**, no en 242 objetos sueltos.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Cambio aplicado, o nada que cambiar | 0 |
| Sin banderas de cambio, banderas incompatibles, selector por clave con varias tareas, solape | 2 |
| Valor fuera de un vocabulario, clave de extensión no declarada, tarea ilegible | 3 |
| Alguna referencia no existe, o un selector de texto no encaja con ningún criterio | 4 |
| Alguna referencia de texto encaja con varias tareas, o un selector con varios criterios | 5 |
| `--dry-run` que no pasa la validación | 9 |
| El almacén falla, o no se obtiene el acceso exclusivo | 7 |
| No hay tablero | 8 |

## `biso set --help`

```
Usage: biso set <ref>... [options]

Change any field of one or more tasks, all or nothing. Every flag here means
the same in `biso new`, `biso start`, `biso note`, `biso comment`, `biso ask`,
`biso answer`, `biso finish` and `biso archive`. Every flag name says what it
does; there is no rule to learn beyond the name.

List fields that take comma-separated values have four shapes, and there is no
field that breaks them:
  --add-labels X      add one or more       --replace-labels X   replace the whole list
  --rm-labels X       remove one or more    --clear-labels       empty the list
The same works for --assignees, --refs, --docs, --deps and --files.

Criteria and definition of done have three, because a criterion's text can
contain a comma and so is never split on one. There is no whole-list replace;
do it by clearing and adding in the same call.
      --add-ac <text>        add a criterion; repeatable
      --rm-ac <sel>          remove by selector; sel is all, 3, 1-4, 1,3,7 or
                             the criterion text. The numbers are stable #N
                             keys. With several tasks, sel has to be all
      --clear-acs            empty the list
      --add-dod / --rm-dod / --clear-dods    the same, for definition of done
      --check-ac <sel>       check criteria, by the same kind of selector
      --uncheck-ac <sel>     the opposite
      --check-dod <sel>      the same for definition-of-done items
      --uncheck-dod <sel>    the opposite

Prose fields have two, because a block of text has no single item to remove.
Replace by clearing and appending in the same call.
      --append-desc X (-d)   append a paragraph
      --append-plan X
      --append-note X
      --append-summary X
      --clear-desc / --clear-plan / --clear-notes / --clear-summary

External fields have three: --ext key=value sets that one key, --rm-ext key
drops it, --clear-ext empties the map. There is no --replace-ext: setting a
key already replaces its value.

Scalars just take a value: -t/--title, -s/--status, --type, --priority,
--project, -m/--milestone, -p/--parent, --due, --ordinal, --reporter. Each has
a --clear-<field>. An empty string is never a way to clear anything.

Comments:
      --comment <text>       append a comment; repeatable
      --comment-author <@w>  who wrote it (default: you)

Resolution:
      --id / --match         force <ref> to be an id, or free text

Within one call, every --rm-*/--clear-* is applied before every --add-*/
--append-*, regardless of the order they were written in. A --replace-* over
a non-empty list is allowed and warns on stderr with how many items it
replaced.

Exit codes:
  0  done                    5  something matched more than one thing
  2  bad usage               7  the board could not be written
  3  unknown value           9  --dry-run did not pass
  4  a task or a criterion was not found
                             8  no board here

Examples:
  biso set MYP-11 --priority high --add-labels parser
  biso set MYP-11 --check-ac 1,3 --append-note "Both covered by diff_test.rs"
  biso set MYP-11 MYP-12 --milestone "v1.2"
  biso set "CRLF" --clear-desc --append-desc @docs/bugs/BUG-02.md
```

---

