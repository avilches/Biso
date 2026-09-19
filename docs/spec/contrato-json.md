# El contrato JSON

## El sobre

Toda salida con `--json` es **un solo objeto JSON**, con o sin sangrado, indistintamente, y tiene una
de dos formas disjuntas según haya salido bien o mal: la que sigue, con `data`, cuando la llamada
termina bien, o la de ["Los errores en JSON"](#los-errores-en-json), con `error` en su lugar, cuando
no. Un sobre nunca lleva las dos claves a la vez.

```json
{ "schemaVersion": 1, "kind": "<tipo>", "generatedAt": "<ISO 8601 UTC>", "data": { } }
```

| `kind` | Lo produce | `data` contiene |
|---|---|---|
| `prime` | `prime` | Sección ["El esquema JSON"](cmd/prime.md#el-esquema-json) |
| `where` | `where` | `id`, `board`, `path`, `source`, `me`, `counts`, `discarded`. Ejemplo en ["`biso where`"](cmd/where.md) |
| `init` | `init` | `board`, `pointerCreated`. Ejemplo en ["`biso init`"](cmd/init.md) |
| `task.list` | `ls` | `tasks`, `shown`, `matched`, `hidden`, `truncated`, `skipped`, `sort`, `filters`, `warnings`. `filters` tiene su propio esquema en ["Los filtros de `biso ls`"](#los-filtros-de-biso-ls) |
| `task.get` | `get` | `task` |
| `task.candidates` | `get` con varias coincidencias | `tasks` |
| `task.write` | `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer`, `archive` | `tasks`, `warnings` |
| `config` | `config list` | `config`. Ejemplo en ["`biso config`"](cmd/config.md) |
| `doctor` | `doctor` | `problems`, `warnings`, `fixed`. Ejemplo en ["`biso doctor`"](cmd/doctor.md) |
| `snapshot` | `snapshot` | `tasks`, `files`, `vcs`, `committed`, `commit`, `repository`, `pushed`, `vcsOutput`, `stagedOutsideBoard`, `skipped`. Ejemplo en ["`biso snapshot`"](cmd/snapshot.md) |
| `board` | `board` | `url`, `port`, `opened`. Ejemplo en ["`biso board`"](cmd/board.md), y se imprime al arrancar el servidor |
| `help` | `help` | `commands`, con el nombre y el resumen de cada uno. Ejemplo en ["`biso help`"](cmd/help.md#biso-help) |
| `error` | cualquier fallo | Ver ["Los errores en JSON"](#los-errores-en-json) |

Un lote de doscientas cuarenta y dos tareas es **un solo sobre** con doscientas cuarenta y dos
entradas en `data.tasks`, nunca doscientos cuarenta y dos objetos sueltos. La única salida del
programa que es una secuencia de objetos, uno por línea, es `biso export`, que no lleva sobre porque
su formato es NDJSON por definición.

## Los filtros de `biso ls`

`data.filters`, dentro del sobre de `task.list` (["`biso ls`"](cmd/ls.md#el-esquema-json)), no es un
eco de los flags que se escribieron: es el filtro **efectivo, ya resuelto**, que produjo `data.tasks`,
incluidos los valores que se resuelven por defecto (el estado terminal excluido sin `-s`, o `--mine`
resuelto a la identidad concreta que se usó). Como cualquier otra clave documentada de este contrato,
**está siempre presente**, con la forma de esta tabla:

| Clave | Flag(s) de `biso ls` | Forma | Valor cuando no se filtra por él |
|---|---|---|---|
| `status` | `-s/--status` | `list<string>` | el valor por defecto ya resuelto: todo menos el estado terminal |
| `notStatus` | `--not-status` | `list<string>` | `[]` |
| `anyStatus` | `--any-status` | `bool` | `false` |
| `archived` | `--archived` | `bool` | `false` |
| `onlyArchived` | `--only-archived` | `bool` | `false` |
| `type` | `--type` | `list<string>` | `[]` |
| `priority` | `--priority` | `list<string>` | `[]` |
| `label` | `-l/--label` | `list<string>` | `[]` |
| `labelOr` | `--label-or` | `list<string>` | `[]` |
| `assignee` | `-a/--assignee`, `--mine` | `list<string>` | `[]` |
| `unassigned` | `--unassigned` | `bool` | `false` |
| `parent` | `-p/--parent` | `string \| null` | `null` |
| `blocked` | `--blocked` / `--not-blocked` | `bool \| null` | `null` |
| `waiting` | `--waiting` / `--not-waiting` | `bool \| null` | `null` |
| `active` | `--active` / `--not-active` | `bool \| null` | `null` |
| `overdue` | `--overdue` | `bool` | `false` |
| `dueBefore` | `--due-before` | `string \| null` (`YYYY-MM-DD`) | `null` |
| `search` | `--search` | `string \| null` | `null` |
| `unchecked` | `--unchecked` | `bool` | `false` |

Precisiones:

- **Una lista con más de un valor es siempre un "o"**, salvo `label`, la única que se combina con "y"
  (["Reglas de combinación de filtros"](cmd/ls.md#parámetros)). El JSON no lo distingue por forma, las
  dos son `list<string>`; lo distingue la clave.
- **`assignee` ya trae `--mine` resuelto** a la identidad concreta que se usó, igual que `status`
  resuelve su valor por defecto. `unassigned` es una clave aparte porque "nadie asignado" no es una
  persona que se pueda meter en esa lista.
- **`parent` es el identificador ya resuelto** (`MYP-11`), nunca el texto de búsqueda que se haya
  tecleado tras `-p` (["Cómo se resuelve una referencia a una tarea"](referencias.md)).
- **`blocked`, `waiting` y `active` son los tres únicos filtros con un opuesto explícito que compite
  por el mismo bit.** `null` es "no se pidió ninguno de los dos", la misma convención que usa el resto
  de este contrato para "sin valor" (["Números, fechas y ausencias"](#números-fechas-y-ausencias)), y
  no `false`, que ya significa "se pidió la variante negativa". Los demás booleanos de la tabla son
  unarios, sin opuesto, así que `false` ya significa por sí solo "no se filtró por esto".

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

**Cuando el propio `--json` es la parte inválida de la llamada, el error es texto plano por stderr,
no este sobre.** Pasa cuando `--json` no se combina con otro flag de la misma llamada (`export
--json`, `prime --full --json`) o cuando el comando o subcomando no lo acepta en absoluto (`config
get --json`, `config set --json`): el análisis de flags rechaza la combinación antes de que `--json`
llegue a establecer un modo de salida válido, así que no hay sobre que envolver el error. Cada uno de
esos comandos documenta su propio mensaje.

Cuando un solo comando produce varios fallos, como un lote inválido, `error.details` es una lista de
objetos con la misma forma, uno por fallo, y `error.code` es `batch_invalid`.

**Si la llamada que falla habría emitido avisos de haber terminado bien, esos avisos viajan dentro
del mismo sobre, en una clave `warnings` al mismo nivel que `error`**, con la misma forma que
`data.warnings` de una salida sin error (["El esquema JSON"](cmd/set.md#el-esquema-json) y el catálogo de
["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)). Es la única excepción a "stderr solo lleva el
objeto de error": en vez de imprimir primero el texto de cada `warning:` y luego el objeto JSON, que
mezclaría dos formas distintas en el mismo flujo, los dos viajan dentro del único objeto que sale por
stderr. `warnings` solo aparece cuando la llamada fallida produjo al menos un aviso; su ausencia no se
distingue de una lista vacía, así que un consumidor que no la encuentre puede tratarlo como "ningún
aviso".

**Un sobre de error no es una salida de datos, y se gobierna aparte de la promesa de la sección ["Números, fechas y ausencias"](#números-fechas-y-ausencias).**
Esa promesa existe porque quien consume una salida de datos no puede prever qué habrá dentro, así que
tiene derecho a que la forma no dependa del contenido. En un error sí puede preverlo, porque lo primero
que hace es leer `code`, y cada `code` trae siempre las mismas claves. Estas son las tres que están en
todos los errores, las cinco de detalle dentro de `error`, y una sexta que vive fuera de él, con la
regla de cuándo acompañan:

| Clave | En qué errores aparece |
|---|---|
| `exitCode`, `code`, `message` | En todos, siempre |
| `field` y `given` | En los que nombran un flag, una clave de configuración o un valor de entrada concreto: todos los del código 3, y los del 2 que nombran un flag |
| `valid` | En los que rechazan un valor contra un conjunto conocido: los del 3 sobre vocabulario, y los del 2 sobre un dominio cerrado, como el modo de `--vcs` |
| `details` | Solo en `batch_invalid` y en `dry_run_failed`, y es una lista de objetos de esta misma forma, uno por fallo |
| `vcsOutput` | Solo en `vcs_commit_failed` y en `vcs_push_failed`, y es la lista de líneas que escribió la orden que falló (["`biso snapshot`"](cmd/snapshot.md)) |
| `warnings` | Al mismo nivel que `error`, no dentro de él, en cualquier `code` cuya llamada haya producido al menos un aviso antes de fallar, con la misma forma que `data.warnings` (["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)) |

Las cinco de detalle van juntas con su `code` y no con su código de salida, que es lo que hace la regla
comprobable: quien ramifica sobre `unknown_status` sabe que va a tener `field`, `given` y `valid`, y
quien ramifica sobre `busy` sabe que no va a tener ninguna de las cinco. `warnings` es la excepción:
no depende del `code`, sino de si esa llamada en concreto llegó a producir algún aviso antes de fallar,
así que puede acompañar a cualquiera de ellos.

## Los identificadores de error

Un `code` estable es lo que permite ramificar sin analizar prosa. Esta es la lista de la versión 1.0,
agrupada por el código de salida con el que sale cada uno:

| Código de salida | `code` |
|---:|---|
| 1 | `internal` |
| 2 | `incompatible_flags`, `duplicate_scalar_flag`, `unexpected_argument`, `missing_value`, `unknown_flag`, `unknown_command`, `missing_title`, `nothing_to_change`, `malformed_id`, `malformed_label`, `malformed_assignee`, `malformed_extension_key`, `malformed_string_value`, `id_like_positional`, `inverted_range`, `key_selector_with_many_tasks`, `criterion_selector_overlap`, `comment_selector_overlap`, `two_stdin`, `read_only_flag`, `invalid_date`, `invalid_number`, `invalid_prefix`, `dependency_cycle`, `parent_cycle`, `self_dependency`, `board_exists`, `delete_not_supported`, `missing_identity`, `invalid_status_roles`, `unknown_status_role`, `too_few_statuses`, `invalid_snapshot_config`, `invalid_vcs_mode`, `vcs_push_unavailable` |
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
- **La única excepción son las claves que gobierna un flag**, y se sostiene porque quien llama sabe
  qué flags ha escrito: no tiene que mirar la salida para averiguar qué va a encontrarse en ella. Lo
  que la regla de arriba prohíbe es lo otro, que la presencia de una clave dependa de los datos, que es
  justo lo que el consumidor no puede prever. Estas son todas las que hay en el documento:

  | Clave | `kind` | El flag que la gobierna |
  |---|---|---|
  | `data.task.urgencyBreakdown` | `task.get` | Solo aparece con `--explain-urgency` (["`biso get`"](cmd/get.md)) |
  | Las demás claves de `data.task` | `task.get` | Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas, y ninguna otra (["`biso get`"](cmd/get.md)) |
- Una lista vacía es `[]` y un mapa vacío es `{}`, nunca `null`.
- **Todo lo de arriba gobierna las salidas de datos, y el sobre de error se gobierna aparte** (["Los errores en JSON"](#los-errores-en-json)). La
  razón es que quien consume una salida de datos no puede prever qué habrá dentro, y quien recibe un
  error sí: lo primero que lee es `code`, y cada `code` trae siempre las mismas claves.

---

