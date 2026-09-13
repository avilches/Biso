# El contrato JSON

## El sobre

Toda salida con `--json` es **un solo objeto JSON**, con o sin sangrado, indistintamente, y siempre
con esta forma:

```json
{ "schemaVersion": 1, "kind": "<tipo>", "generatedAt": "<ISO 8601 UTC>", "data": { } }
```

| `kind` | Lo produce | `data` contiene |
|---|---|---|
| `prime` | `prime` | Sección ["El esquema JSON"](cmd/prime.md#el-esquema-json) |
| `where` | `where` | `id`, `board`, `path`, `source`, `me`, `counts`. Ejemplo en ["`biso where`"](cmd/where.md) |
| `init` | `init` | `board`, `pointerCreated`. Ejemplo en ["`biso init`"](cmd/init.md) |
| `task.list` | `ls` | `tasks`, `shown`, `matched`, `hidden`, `truncated`, `skipped`, `sort`, `filters` |
| `task.get` | `get` | `task` |
| `task.candidates` | `get` con varias coincidencias | `tasks` |
| `task.write` | `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer`, `archive` | `tasks`, `warnings` |
| `config` | `config list` | `config`. Ejemplo en ["`biso config`"](cmd/config.md) |
| `doctor` | `doctor` | `problems`, `warnings`, `fixed`. Ejemplo en ["`biso doctor`"](cmd/doctor.md) |
| `snapshot` | `snapshot` | `tasks`, `files`, `vcs`, `committed`, `commit`, `repository`, `pushed`, `vcsOutput`, `skipped`. Ejemplo en ["`biso snapshot`"](cmd/snapshot.md) |
| `board` | `board` | `url`, `port`, `opened`. Ejemplo en ["`biso board`"](cmd/board.md), y se imprime al arrancar el servidor |
| `help` | `help` | `commands`, con el nombre y el resumen de cada uno. Ejemplo en ["`biso help`"](cmd/help.md#biso-help) |
| `error` | cualquier fallo | Ver ["Los errores en JSON"](#los-errores-en-json) |

Un lote de doscientas cuarenta y dos tareas es **un solo sobre** con doscientas cuarenta y dos
entradas en `data.tasks`, nunca doscientos cuarenta y dos objetos sueltos. La única salida del
programa que es una secuencia de objetos, uno por línea, es `biso export`, que no lleva sobre porque
su formato es NDJSON por definición.

## Los errores en JSON

Con `--json`, un error sale **por stderr**, como un objeto con esta forma, y el código de salida del
proceso es el de la tabla de la sección ["Códigos de salida"](codigos-de-salida.md):

```json
{
  "schemaVersion": 1,
  "kind": "error",
  "generatedAt": "2026-09-06T09:12:04Z",
  "error": {
    "exitCode": 3,
    "code": "unknown_status",
    "message": "unknown status: \"Pending\"",
    "field": "status",
    "given": "Pending",
    "valid": ["To Do", "In Progress", "Done"]
  }
}
```

Cuando un solo comando produce varios fallos, como un lote inválido, `error.details` es una lista de
objetos con la misma forma, uno por fallo, y `error.code` es `batch_invalid`.

**Un sobre de error no es una salida de datos, y se gobierna aparte de la promesa de la sección ["Números, fechas y ausencias"](#números-fechas-y-ausencias).**
Esa promesa existe porque quien consume una salida de datos no puede prever qué habrá dentro, así que
tiene derecho a que la forma no dependa del contenido. En un error sí puede preverlo, porque lo primero
que hace es leer `code`, y cada `code` trae siempre las mismas claves. Estas son las tres que están en
todos los errores y las cinco de detalle, con la regla de cuándo acompañan:

| Clave | En qué errores aparece |
|---|---|
| `exitCode`, `code`, `message` | En todos, siempre |
| `field` y `given` | En los que nombran una bandera, una clave de configuración o un valor de entrada concreto: todos los del código 3, y los del 2 que nombran una bandera |
| `valid` | En los que rechazan un valor contra un conjunto conocido: los del 3 sobre vocabulario, y los del 2 sobre un dominio cerrado, como el modo de `--vcs` |
| `details` | Solo en `batch_invalid` y en `dry_run_failed`, y es una lista de objetos de esta misma forma, uno por fallo |
| `vcsOutput` | Solo en `vcs_commit_failed` y en `vcs_push_failed`, y es la lista de líneas que escribió la orden que falló (["`biso snapshot`"](cmd/snapshot.md)) |

Las cinco de detalle van juntas con su `code` y no con su código de salida, que es lo que hace la regla
comprobable: quien ramifica sobre `unknown_status` sabe que va a tener `field`, `given` y `valid`, y
quien ramifica sobre `busy` sabe que no va a tener ninguna de las cinco.

## Los identificadores de error

Un `code` estable es lo que permite ramificar sin analizar prosa. Esta es la lista de la versión 1.0,
agrupada por el código de salida con el que sale cada uno:

| Código de salida | `code` |
|---:|---|
| 1 | `internal` |
| 2 | `incompatible_flags`, `duplicate_scalar_flag`, `unexpected_argument`, `missing_value`, `unknown_flag`, `unknown_command`, `missing_title`, `nothing_to_change`, `malformed_id`, `malformed_label`, `malformed_assignee`, `malformed_extension_key`, `id_like_positional`, `inverted_range`, `key_selector_with_many_tasks`, `criterion_selector_overlap`, `comment_selector_overlap`, `two_stdin`, `read_only_flag`, `invalid_date`, `invalid_number`, `invalid_prefix`, `dependency_cycle`, `parent_cycle`, `self_dependency`, `board_exists`, `delete_not_supported`, `missing_identity`, `invalid_status_roles`, `unknown_status_role`, `too_few_statuses`, `invalid_snapshot_config`, `invalid_vcs_mode`, `vcs_push_unavailable` |
| 3 | `unknown_status`, `unknown_type`, `unknown_priority`, `unknown_label`, `unknown_assignee`, `unknown_extension_key`, `unknown_section`, `unknown_sort_field`, `ambiguous_vocabulary`, `empty_scalar_value`, `bad_config_value`, `undecodable_task`, `invalid_encoding` |
| 4 | `not_found`, `never_allocated`, `unknown_config_key`, `criterion_not_found`, `comment_not_found`, `file_not_found` |
| 5 | `ambiguous_reference`, `criterion_ambiguous`, `comment_ambiguous` |
| 6 | `already_finished`, `precondition_failed`, `board_inconsistent`, `doctor_problems`, `open_question_exists`, `no_open_question`, `mine_requires_identity` |
| 7 | `batch_invalid`, `dry_run_failed` |
| 8 | `busy`, `io_error`, `file_unreadable`, `no_terminal`, `port_in_use`, `vcs_commit_failed`, `vcs_push_failed` |
| 20 | `no_board`, `pointer_unresolved` |
| 21 | `database_unreadable` |
| 22 | `ambiguous_board_id` |

**La lista es ampliable y las entradas son permanentes.** Una versión posterior puede añadir un `code`
nuevo, pero ninguno de los de arriba cambiará de significado, cambiará de código de salida ni
desaparecerá. Quien ramifique sobre un `code` desconocido debe tratarlo por su código de salida, que
sí está cerrado.

`missing_identity` (código 2, en `biso ask`, `biso answer` y el autor de un comentario) y
`mine_requires_identity` (código 6, en `--mine`) son la falta de identidad de la tabla de ["Variables de entorno"](invocacion.md#variables-de-entorno), pero
con códigos de salida distintos. No son un mismo concepto duplicado: como esta tabla está
agrupada por código de salida y ninguno de los dos se mueve nunca, la misma falta de identidad no
puede compartir un `code` cuando sale con códigos distintos. La asimetría entre esos códigos es
anterior a esta rama.

## Números, fechas y ausencias

- Las fechas son ISO 8601 en UTC terminadas en `Z`, con precisión de segundo. Nunca hora local, nunca
  sin zona. `due` es la excepción, porque es un día y no un instante, y viaja como `YYYY-MM-DD`.
- `urgency` es un decimal con un solo dígito tras el punto.
- Un campo sin valor es `null`, nunca la cadena vacía ni la ausencia de la clave. **Ninguna clave va ni
  viene según los datos**: la que está documentada para un `kind` aparece siempre que se emite ese
  `kind`, valga lo que valga, para que nadie tenga que distinguir entre "no está" y "no tiene valor".
- **La única excepción son las claves que gobierna una bandera**, y se sostiene porque quien llama sabe
  qué banderas ha escrito: no tiene que mirar la salida para averiguar qué va a encontrarse en ella. Lo
  que la regla de arriba prohíbe es lo otro, que la presencia de una clave dependa de los datos, que es
  justo lo que el consumidor no puede prever. Estas son todas las que hay en el documento:

  | Clave | `kind` | La bandera que la gobierna |
  |---|---|---|
  | `data.task.urgencyBreakdown` | `task.get` | Solo aparece con `--explain-urgency` (["`biso get`"](cmd/get.md)) |
  | Las demás claves de `data.task` | `task.get` | Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas, y ninguna otra (["`biso get`"](cmd/get.md)) |
- Una lista vacía es `[]` y un mapa vacío es `{}`, nunca `null`.
- **Todo lo de arriba gobierna las salidas de datos, y el sobre de error se gobierna aparte** (["Los errores en JSON"](#los-errores-en-json)). La
  razón es que quien consume una salida de datos no puede prever qué habrá dentro, y quien recibe un
  error sí: lo primero que lee es `code`, y cada `code` trae siempre las mismas claves.

---

