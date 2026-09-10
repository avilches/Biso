# `biso set`

El comando de edición general. Todo lo que hacen los verbos de flujo de 10.7 se puede hacer aquí, con
más palabras.

## Firma

```
biso set <ref>... [cualquier bandera de campo de la seccion 8]
         [--check <sel>]... [--uncheck <sel>]... [--check-dod <sel>]... [--uncheck-dod <sel>]...
         [--comment <text>]... [--comment-author <@who>] [--id] [--match]
```

## Parámetros propios

**Todas** las banderas de las secciones 8.2, 8.3, 8.5 y 8.6 valen aquí, con exactamente el mismo
significado que en cualquier otro comando. Lo propio de `set`:

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--check <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--uncheck` |
| `--uncheck <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--check` |
| `--check-dod <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--uncheck-dod` |
| `--uncheck-dod <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--check-dod` |
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
| Varias referencias | El mismo cambio se aplica a todas, con la garantía de todo o nada de 4.10 |
| Varias referencias y un selector que no sea `all` (clave, rango, lista o texto) | Error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |
| Varias referencias y `--check all` | Válido |
| Una de varias referencias no existe | Error 4, y **no se escribe ninguna**, ni siquiera las buenas |
| Un `--set-*` pisa contenido no vacío | Se hace, con el aviso de 4.3 diciendo cuántos bytes ha reemplazado |
| Paso a un estado terminal con criterios sin marcar | Se hace, con aviso |
| Paso a un estado terminal con una pregunta abierta (5.7) | Se hace, con aviso, igual que en `biso finish` (10.7.4) y como atribuye 4.3 a cualquier llegada al estado terminal |
| Todas las banderas dejan la tarea igual | Código 0, con `note: TASK-11 unchanged`. Ningún campo de la tarea se escribe, `updatedAt` no cambia y `changed` sale vacía, pero si quien llama es `leaseHolder` **el arrendamiento se renueva igual**: es una escritura del tenedor sobre su tarea, y el latido no depende de si los valores coincidían (sexta precisión de la sección 5) |
| `--status` a un estado que no es el activo, `--clear-assignee` o `--rm-assignee` que deja la tarea sin nadie, sobre una tarea con arrendamiento | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea el arrendamiento; si era de otra identidad, sale además el aviso de 4.3 (séptima precisión de la sección 5) |
| `--comment-author` sin `--comment` | Error 2 |
| `--comment` sin `--comment-author` y sin ninguna identidad configurada (3.1) | Error 2 |
| La tarea no se puede leer | Error 3, y no se escribe nada |

## Salida

Por defecto, **una línea por tarea afectada** con lo que quien llama no sabía: el estado resultante,
el avance de criterios y la urgencia recalculada. Los tres datos son derivados, y ninguno se puede
conocer sin leer la tarea. La TASK-11 de los ejemplos tiene además una definición de hecho de un
elemento, así que su línea trae también el avance de esa segunda lista:

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

**Esta es la línea de estado, y la imprimen también los seis verbos del ciclo de 10.7 y
`biso archive`.** La única excepción es `biso new`, por el motivo que da 10.3. Cada comando la enseña
con su propio ejemplo, pero las tres reglas de su forma se dicen aquí y no se repiten:

1. El trozo `ac <marcados>/<total>` sale siempre que la tarea tenga criterios de aceptación.
2. El trozo `dod <marcados>/<total>` sale siempre que tenga definición de hecho. Una tarea sin ninguna
   de las dos listas imprime solo el identificador, el estado y la urgencia.
3. La palabra `archived` cierra la línea cuando la tarea queda archivada, y solo entonces. Es lo único
   que un comando puede añadirle, y quien lo añade es `biso archive` (10.8).

Los avisos van por stderr:

```
warning: --set-plan replaced 412 bytes of existing content
```

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.write",
  "generatedAt": "2026-09-06T11:40:18Z",
  "data": {
    "tasks": [
      { "id": "TASK-11", "status": "In Progress", "acDone": 1, "acTotal": 2,
        "dodDone": 0, "dodTotal": 1, "urgency": 19.0,
        "changed": ["plan", "status"] }
    ],
    "warnings": [ { "code": "overwrite", "field": "plan", "bytes": 412, "task": "TASK-11" } ]
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
`biso answer`, `biso finish` and `biso archive`.

The four shapes, and there is no field that breaks them:
  --label X        add one          --set-label X   replace the whole list
  --rm-label X     remove one       --clear-label   empty the list
The same works for --assignee, --ref, --doc, --dep, --file, --ac and --dod.

Prose fields have three, because a block of text has no single item to remove:
  --desc X adds, --set-desc X replaces, --clear-desc empties. The same for
  --plan, --note (whose replacement is --set-notes) and --summary.

External fields have three too: --ext key=value sets that one key, --rm-ext key
drops it, --clear-ext empties the map.

Scalars just take a value: -t/--title, -s/--status, --type, --priority,
--project, -m/--milestone, -p/--parent, --due, --ordinal, --reporter. Each has
a --clear-<field>. An empty string is never a way to clear anything.

Criteria and definition of done:
      --check <sel>          check criteria; sel is all, 3, 1-4, 1,3,7 or the
                             criterion text. The numbers are stable #N keys.
                             With several tasks, sel has to be all
      --uncheck <sel>        the opposite
      --check-dod <sel>      the same for definition-of-done items
      --uncheck-dod <sel>    the opposite
      --rm-ac <sel>          remove criteria by the same selector
      --rm-dod <sel>         remove definition-of-done items

Comments:
      --comment <text>       append a comment; repeatable
      --comment-author <@w>  who wrote it (default: you)

Resolution:
      --id / --match         force <ref> to be an id, or free text

A --set-* over existing content is allowed and warns on stderr with how many
bytes it replaced.

Exit codes:
  0  done                    5  something matched more than one thing
  2  bad usage               7  the board could not be written
  3  unknown value           9  --dry-run did not pass
  4  a task or a criterion was not found
                             8  no board here

Examples:
  biso set TASK-11 --priority high --label parser
  biso set TASK-11 --check 1,3 --note "Both covered by diff_test.rs"
  biso set TASK-11 TASK-12 --milestone "v1.2"
  biso set "CRLF" --set-desc @docs/bugs/BUG-02.md
```

---

