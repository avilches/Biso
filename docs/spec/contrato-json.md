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
incluidos los valores que se resuelven por defecto (el estado terminal excluido sin `--status`, o `--mine`
resuelto a la identidad concreta que se usó). Como cualquier otra clave documentada de este contrato,
**está siempre presente**, con la forma de esta tabla:

| Clave | Flag(s) de `biso ls` | Forma | Valor cuando no se filtra por él |
|---|---|---|---|
| `status` | `--status` | `list<string>` | el valor por defecto ya resuelto: todo menos el estado terminal, o **todos** cuando se escribió `--any-status`, porque ese es el filtro que de verdad se aplicó |
| `notStatus` | `--not-status` | `list<string>` | `[]` |
| `anyStatus` | `--any-status` | `bool` | `false` |
| `archived` | `--archived` | `bool` | `false` |
| `onlyArchived` | `--only-archived` | `bool` | `false` |
| `type` | `--type` | `list<string>` | `[]` |
| `priority` | `--priority` | `list<string>` | `[]` |
| `label` | `--label` | `list<string>` | `[]` |
| `labelOr` | `--label-or` | `list<string>` | `[]` |
| `assignee` | `--assignee`, `--mine` | `list<string>` | `[]` |
| `unassigned` | `--unassigned` | `bool` | `false` |
| `parent` | `--parent` | `string \| null` | `null` |
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
  tecleado tras `--parent` (["Cómo se resuelve una referencia a una tarea"](referencias.md)).
- **Un filtro por la clave de una etiqueta viaja en `label` y en `labelOr` con un solo dos puntos**,
  `milestone:`, aunque se escribiera `milestone::`, porque las dos formas son el mismo filtro
  (["Consultar por la clave de una etiqueta con ámbito"](vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito)).
  Lo que sí conserva es la grafía de la clave tal como se tecleó: se compara plegada y no hay ninguna
  grafía configurada a la que resolverla, al contrario de lo que pasa con un `status`.
- **Ninguna tarea trae un campo derivado por clave de etiqueta**, en ningún `kind`: `labels` es la
  lista entera, y quien quiera el valor de una clave la parte por los primeros dos puntos con la
  regla de ["Las etiquetas con ámbito"](valores-de-entrada.md#las-etiquetas-con-ámbito). Añadir ese
  campo después sería un cambio compatible; quitarlo, no, y por eso no entra en la versión 1.0
  (["Las etiquetas con ámbito"](../decisiones/detalles.md#las-etiquetas-con-ámbito)).
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
| `valid` | En los que rechazan un valor contra un conjunto conocido: los del 3 sobre vocabulario, y los del 2 sobre un dominio cerrado, como el modo de `--vcs`. En `ambiguous_vocabulary` no es el conjunto entero, sino **solo los valores configurados que empatan**, que es lo que hay que desambiguar (["Cuando el tablero tiene dos valores que se normalizan igual"](vocabularios.md#cuando-el-tablero-tiene-dos-valores-que-se-normalizan-igual)). `wrong_label_separator` no lo lleva, porque lo que rechaza no es un valor contra un conjunto sino un separador contra el que la clave declara, y una clave abierta no tiene ningún valor que ofrecer (["La lista `labels`"](cmd/config.md#la-lista-labels)) |
| `details` | Solo en `batch_invalid`, y es una lista de objetos de esta misma forma, uno por fallo |
| `vcsOutput` | Solo en `vcs_commit_failed` y en `vcs_push_failed`, y es la lista de líneas que escribió **la orden que falló**, no las de las que fueron bien antes, que sí están todas en `data.vcsOutput` cuando la llamada acaba bien (["`biso snapshot`"](cmd/snapshot.md)) |
| `warnings` | Al mismo nivel que `error`, no dentro de él, en cualquier `code` cuya llamada haya producido al menos un aviso antes de fallar, con la misma forma que `data.warnings` (["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)) |

**La fila de `field` y `given` tiene excepciones declaradas, y no son todas del mismo signo.** La
primera es `incompatible_flags`, que **no lleva ninguna de las dos**: nombra un par de flags, y
ninguno de los dos es más culpable que el otro, así que elegir uno para `field` sería inventarse una
atribución que la llamada no tiene. Los demás errores del código 2 que nombran un flag sí las llevan,
`read_only_flag` incluido, que nombra uno solo (["Los flags globales"](cmd/flags-globales.md)).

La siguiente es `invalid_lease`, que **lleva `field` y no lleva `given`**. Sus dos mitades
(["`biso new`"](cmd/new.md#el-modo-lote)) no reprochan ningún valor: la primera reprocha el estado
de la tarea sobre la que llegó la clave, y la segunda que falte la otra clave de la pareja. No hay
nada que citar, así que la clave no viaja, en vez de viajar vacía: una `given` de cadena vacía
significa "el valor que llegó estaba vacío", que es lo que contesta `--rm-labels ""`, y no "no había
ningún valor". Es la misma regla que ["Números, fechas y ausencias"](#números-fechas-y-ausencias)
aplica a una salida de datos, aquí aplicada a la única clave de un error que puede faltar sin que
falte también su pareja.

Y la última es `mixed_label_separators`, que **lleva `field` igual a `labels` y tampoco lleva
`given`**, por la misma razón que `incompatible_flags` no elige flag: su mensaje cita dos etiquetas
que se contradicen entre sí (["Escribir una etiqueta con ámbito"](familias-de-flags.md#escribir-una-etiqueta-con-ámbito))
y ninguna de las dos es la culpable, así que copiar una en `given` diría que el problema es esa y no
la pareja. `field` sí se puede nombrar sin inventar nada, porque el campo en discordia es uno solo.
`exclusive_label_conflict`, en cambio, no lleva ninguna de las dos claves y no es ninguna excepción:
sale con el código 6, y la fila de arriba solo se las promete a los del 3 y a los del 2 que nombran
un flag.

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
| 2 | `incompatible_flags`, `duplicate_scalar_flag`, `unexpected_argument`, `missing_value`, `unknown_flag`, `unknown_command`, `unknown_section`, `unknown_sort_field`, `missing_title`, `missing_ref`, `nothing_to_change`, `malformed_id`, `malformed_label`, `malformed_assignee`, `malformed_string_value`, `mixed_label_separators`, `id_like_positional`, `missing_text`, `inverted_range`, `key_selector_with_many_tasks`, `criterion_selector_overlap`, `comment_selector_overlap`, `two_stdin`, `read_only_flag`, `invalid_date`, `invalid_number`, `invalid_prefix`, `dependency_cycle`, `parent_cycle`, `self_dependency`, `board_exists`, `delete_not_supported`, `missing_identity`, `invalid_status_roles`, `unknown_status_role`, `too_few_statuses`, `invalid_snapshot_config`, `invalid_snapshot_id`, `unknown_key`, `id_taken`, `invalid_lease`, `invalid_vcs_mode`, `invalid_color_mode`, `vcs_push_unavailable`, `vcs_commit_unavailable` |
| 3 | `unknown_status`, `unknown_type`, `unknown_priority`, `unknown_label`, `unknown_label_key`, `unknown_label_value`, `wrong_label_separator`, `unknown_assignee`, `ambiguous_vocabulary`, `empty_scalar_value`, `bad_config_value`, `undecodable_task`, `invalid_line`, `invalid_encoding` |
| 4 | `not_found`, `never_allocated`, `no_such_command`, `unknown_config_key`, `criterion_not_found`, `comment_not_found`, `file_not_found` |
| 5 | `ambiguous_reference`, `criterion_ambiguous`, `comment_ambiguous` |
| 6 | `already_finished`, `precondition_failed`, `board_inconsistent`, `doctor_problems`, `open_question_exists`, `no_open_question`, `mine_requires_identity`, `exclusive_label_conflict` |
| 7 | `batch_invalid` |
| 8 | `busy`, `io_error`, `file_unreadable`, `lease_lost`, `no_terminal`, `port_in_use`, `vcs_commit_failed`, `vcs_push_failed` |
| 20 | `no_board`, `pointer_unresolved` |
| 21 | `database_unreadable` |
| 22 | `ambiguous_board_id` |

**Dos `code` de la fila del 8 están reservados y la versión 1.0 no los emite nunca**: `no_terminal`
y `port_in_use` son de [`biso board`](cmd/board.md), la interfaz interactiva que
["Qué hay implementado y qué no"](estado-de-implementacion.md) declara fuera del alcance de 1.0. No
se quitan de la lista porque
["El contrato de estabilidad"](estabilidad.md) fija que las entradas son permanentes; se nombran
aquí para que quien recorra la lista buscando qué produce cada una no los busque en vano.

**Los cinco `code` del lote y de la instantánea se añadieron al implementarlos**, porque la
especificación describía sus mensajes sin darles identificador: `unknown_key` (una clave que el
formato de intercambio no declara), `id_taken` (un `id` que el tablero ya tiene, o que aparece dos
veces en el mismo fichero), `invalid_lease` (cualquiera de las dos mitades de la invariante de
["El vaciado"](lease.md#el-vaciado) rota en una línea), `invalid_line` (una línea que no se puede
interpretar como una tarea del formato: JSON mal formado, un valor del tipo equivocado, un `null` en
una lista, una fecha ilegible o una clave de criterio o de comentario repetida) y
`invalid_snapshot_id`, que ["`biso init`"](cmd/init.md) ya nombraba en su tabla de casos y que esta
lista no llevaba. Los cuatro primeros viajan siempre dentro de `details`, porque el lote los agrupa
bajo un `batch_invalid`.

**`unknown_section` y `unknown_sort_field` salen con código 2 y no con 3**, aunque los dos empiecen
por `unknown_` como los del vocabulario. La diferencia es contra qué se comprueba el valor: el
vocabulario de `status`, `type` o `priority` lo configura cada tablero, así que un valor que no está
en él es un hecho sobre ese tablero y sale con 3; el juego de secciones de ["`biso get`"](cmd/get.md) y el de campos de
`--sort` de ["`biso ls`"](cmd/ls.md) son del programa, iguales en todas partes, así que escribir uno que no existe es
una línea de comandos mal escrita, exactamente igual que `--color rosa` (`invalid_color_mode`) o
`--vcs fossil` (`invalid_vcs_mode`), que ya salen con 2. Las dos tablas de códigos de esas dos
páginas lo dicen así desde siempre; lo que se corrigió, al implementarlas, fue la fila de esta lista,
que los tenía agrupados con los del 3.

**`no_such_command` sale con código 4 y `unknown_command` con 2, y nombran la misma clase de cosa por
una razón distinta.** `unknown_command` es lo que contesta el analizador cuando la primera palabra de
la línea no es un comando: ahí la llamada no se puede llevar a cabo en absoluto, y eso es el código 2.
`no_such_command` es lo que contesta [`biso help`](cmd/help.md) cuando la llamada está bien formada,
se entiende entera, y nombra como argumento algo que no existe, que es exactamente lo que significa el
código 4 en cualquier otro comando (["Códigos de salida"](codigos-de-salida.md)). La diferencia se ve
en que `biso fnish` no puede hacer nada y `biso help fnish` sí podría, si ese comando existiera.

**`doctor_problems` es el único `code` de esta lista que no viaja nunca en un sobre de error.**
[`biso doctor`](cmd/doctor.md) sale con 6 cuando quedan errores, pero lo que imprime entonces es su
informe, con su propio `kind`, y no un sobre de error: el hallazgo es el resultado del comando y no el
motivo de que no lo haya. La clave sigue en la lista porque nombra ese desenlace, no porque alguna
salida la lleve.

**`missing_text` es el posicional de texto que falta** en los verbos del ciclo que llevan uno
(["Los verbos del ciclo"](cmd/verbos-del-ciclo.md)), y no se confunde con `missing_value`, que es una
flag escrita sin su valor: una es un argumento que no está y la otra una flag a medias.
**`lease_lost` es la reclamación de un arrendamiento vencido que perdió la carrera** contra otra
simultánea (["`biso start`"](cmd/verbos-del-ciclo.md#biso-start)); sale con 8 porque lo que falló no
fue la petición, que era correcta, sino conseguir el acceso exclusivo que hacía falta para servirla,
que es exactamente lo que ese código cubre en la tabla de ["`biso set`"](cmd/set.md#códigos-de-salida).

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
- `urgency` es un decimal con un solo dígito tras el punto, **y también lo es cada uno de los siete
  términos de `urgencyBreakdown`**, incluido `urgencyBreakdown.active.value`
  (["`biso get`"](cmd/get.md#el-esquema-json)). Los términos son los sumandos de ese mismo número, así que se escriben
  igual que él: `6.0` y no `6`, y los siete ceros de una tarea en estado terminal son `0.0` y no `0`.
  Decirlo solo de `urgency` fue lo que dejó que el desglose se serializara como entero.
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

