# Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`

Los seis aceptan **todos** los flags de campo de la sección ["Las familias de flags"](../familias-de-flags.md), igual que `set`. No son un
subconjunto: lo que aportan es un nombre y unos valores por defecto, de modo que el gesto frecuente
cabe en una llamada corta y el gesto raro sigue cabiendo en la misma llamada.

Los seis son escrituras sobre la tarea, así que a los cinco que no son `start` (`note`, `comment`,
`finish`, `ask`, `answer`) les aplica la regla general de [`lease.md`](../lease.md#la-renovación): si
quien llama ya es `leaseHolder`, renuevan `leaseExpiresAt`; si no lo es y el arrendamiento está vivo,
no tocan ninguno de los campos y avisan; y si no lo es y está vencido, lo dejan vencido. Ninguno de
los cinco fija ni transfiere `leaseHolder`: reclamar es de `start`, y de [`biso new --start`](new.md)
al crear. **Y por encima de todo eso está la invariante**: la escritura que saca la tarea del estado
activo o la deja sin ninguna persona asignada vacía los campos, sea quien sea quien la haga, así que
`biso finish` los vacía siempre y el aviso de un arrendamiento ajeno no lo impide (["El
vaciado"](../lease.md#el-vaciado) de `lease.md`).

Los seis imprimen también la misma línea de estado que `set`, con la forma y las reglas que define
[`biso set`](set.md). Los ejemplos de más abajo son esa línea con los datos de la MYP-11, que tiene dos criterios de
aceptación y un elemento de definición de hecho.

El ciclo entero de una tarea es esto:

```
biso start  MYP-11 --append-plan "1. Read the parser. 2. Add the CRLF case."
biso note   MYP-11 "The parser already normalized LF, CRLF was missing"
biso finish MYP-11 --check-ac all --check-dod all --append-summary "Normalize CRLF in the diff, verified with the tests."
```

## `biso start`

### Firma

```
biso start <ref>... [--append-plan <text>] [-a <@who>]... [-s <v>] [--reopen]
           [--id] [--match] [cualquier flag de campo de las familias de flags]
```

### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--status <v>` | `-s` | no | vocabulario | `active_status` | no | no | |
| `--reopen` | | no | booleano | falso | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

`--append-plan` y `-a/--add-assignees` son los flags de campo de la sección ["Las familias de flags"](../familias-de-flags.md), con su
significado de siempre: **las dos añaden**. `--append-plan` añade al plan existente; sustituirlo entero
se hace vaciando y añadiendo en la misma llamada (`--clear-plan --append-plan ...`), porque un bloque
de prosa no tiene flag de sustituir entera. `-a` añade una persona y `--replace-assignees` reemplaza
la lista.

### Qué hace

Cuatro cosas en una escritura: pone el estado activo, **asigna la tarea a `me` si no tiene ninguna
persona asignada**, toma el arrendamiento (`leaseExpiresAt`, `leaseHolder`, sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md)) a favor de
quien llama (renovándolo si ya era suyo, reclamándolo si estaba vencido, o tomándolo si era de otra
identidad: es el único comando que hace las tres cosas sobre una tarea que ya existe, sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md)), y
añade el plan si se ha pasado. Con `-s` a un estado que no es el activo no hay arrendamiento que tomar,
y la fila correspondiente de la tabla dice qué pasa entonces.

| Caso | Qué pasa |
|---|---|
| La tarea ya está en el estado activo | Se aplica el resto igual, con `note: MYP-11 was already In Progress` |
| La tarea ya está en el estado terminal | Error 6, salvo con `--reopen`, que la devuelve al estado activo |
| La tarea tiene dependencias sin terminar | Se empieza igual, con el aviso correspondiente. **Avisa, no impide** |
| La tarea tiene [una pregunta abierta](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta) | Se empieza igual, con el aviso correspondiente. **Avisa, no impide**, exactamente como con las dependencias sin terminar |
| El arrendamiento de la tarea está vencido (`leaseExpired`, ["Cuándo cuenta como vencido"](../lease.md#cuándo-cuenta-como-vencido) de `lease.md`) | Se reclama dentro de la misma transacción: `leaseHolder` pasa a ser quien llama y `leaseExpiresAt` se renueva, comprobando en esa misma transacción que seguía vencido, **para que de dos reclamaciones simultáneas del mismo arrendamiento vencido solo gane una**. Lo que esa comprobación no hace es impedirle escribir al tenedor viejo cuando despierte: ninguna escritura corriente suya renueva ni recupera un arrendamiento que ya es de otra identidad (["La renovación"](../lease.md#la-renovación) de `lease.md`), pero puede seguir anotando, comentando y cerrando la tarea, y con otro `biso start` se la lleva de vuelta con el aviso de la fila siguiente. Es la diferencia deliberada con el token de vallado del patrón, anotada como riesgo aceptado en la sección ["Riesgos conocidos y aceptados del modelo de estados"](../../decisiones/modelo-de-estados.md#riesgos-conocidos-y-aceptados-del-modelo-de-estados) |
| El arrendamiento de la tarea está vivo y es de otra identidad | Se coge igual, con `warning: MYP-11's lease is held by @sara until 2026-09-08T14:00:00Z`. **Avisa, no impide**, por el mismo motivo que las dependencias sin terminar y la pregunta abierta: un bloqueo de flujo no evita el trabajo duplicado, solo empuja a rodear la herramienta modificando datos que no deberían tocarse |
| `-s` con un estado que no es el activo, por ejemplo `biso start MYP-1 -s "To Do"` | Se aplica todo lo demás, pero **no se fija ningún arrendamiento**, y si la tarea lo tenía se vacía como en cualquier otra escritura que la saque del estado activo (["El vaciado"](../lease.md#el-vaciado) de `lease.md`). Fijarlo ahí rompería la invariante de que los campos solo tienen valor en una tarea activa y asignada, y `-s` acepta cualquier estado del vocabulario, así que este caso existe. Sale `note: MYP-1 was moved to To Do, no lease was claimed` |
| La tarea ya tiene otra persona asignada | No se añade `me`, y sale `note: MYP-11 is assigned to @sara, left as is`. Con `-a` explícito, se añade lo que diga `-a` |
| No hay [ninguna identidad configurada](../invocacion.md#variables-de-entorno) y no se pasa `-a` | No asigna a nadie, con `note: no identity configured, task left unassigned`, y tampoco se fija el arrendamiento: no hay ninguna identidad a la que atribuírselo |
| La tarea ya tiene plan y se pasa `--append-plan` | Se añade al final, como todo flag de añadir |
| Varias referencias | Todo o nada |

### Salida

```
MYP-11  In Progress  ac 0/2  dod 0/1  urgency 19.0
```

El valor de urgencia del ejemplo sale de los coeficientes por defecto, que el contrato de estabilidad
permite cambiar entre versiones menores, así que la cifra exacta puede no ser esta.

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Empezada | 0 |
| Ya estaba terminada y no hay `--reopen` | 6 |
| Referencia mal formada, flags incompatibles | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Referencia inexistente | 4 |
| Referencia ambigua | 5 |
| `--dry-run` que no pasa | 7 |
| El almacén falla | 8 |
| No hay tablero | 20 |

### `biso start --help`

```
Usage: biso start <ref>... [options]

Take one or more tasks: move them to the active status, claim the lease for
you, assign them to you if nobody has them, and record a plan. One call.

Options:
      --append-plan <text>       add to the implementation plan; repeatable,
                                 and takes @file and - like every text option
  -a, --add-assignees <@who>     add an assignee (--replace-assignees replaces
                                 the list)
  -s, --status <value>           use another status instead of the active one;
                                 no lease is claimed then, a lease only exists
                                 on an active task
      --reopen           allow starting a task that is already finished
      --id / --match     force <ref> to be an id, or free text
  -h, --help             show this help

Every field flag of `biso set --help` works here too.

Unresolved dependencies produce a warning, not an error: you decide. Taking
over a live lease held by someone else is the same: it warns, it does not
refuse.

Exit codes:
  0  started        4  not found        7  --dry-run did not pass
  2  bad usage      5  ambiguous        8  the board could not be written
  3  unknown value  6  already finished, use --reopen
                    20 no board here

Examples:
  biso start MYP-11 --append-plan "1. Read the parser. 2. Add the CRLF case."
  biso start 11
  biso start MYP-11 MYP-12
```

## `biso note`

### Firma

```
biso note <ref> [<text>...] [--id] [--match] [cualquier flag de campo de las familias de flags]
```

### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí, salvo con `--note` | texto largo | | sí, como posicional | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Cada texto es un párrafo propio en la sección de notas. Aceptan `@fichero` y `-` como cualquier texto
largo.

### El posicional que parece un identificador

`biso note` toma **una sola** referencia, mientras que `set`, `start`, `finish` y `archive` toman
varias. Para que esa diferencia no produzca basura en silencio, **un posicional de texto que encaje
con la gramática de identificador de la sección ["La gramática"](../referencias.md#la-gramática) es un error 2**:

```
error: "MYP-2" looks like a task id, and `biso note` takes only one task
hint: to note the same thing on several tasks: biso set MYP-1 MYP-2 --append-note "..."
      to write that text literally:            biso note MYP-1 --append-note "MYP-2"
```

**`--append-note`, el flag de campo de la sección ["Las familias de flags"](../familias-de-flags.md), nunca pasa por esa comprobación**, porque no
es un posicional: es la vía para escribir una nota que de verdad diga `MYP-2`. La misma regla vale
para `biso comment`, con `--comment`.

### Qué hace

Añade uno o más párrafos a las notas de implementación. **Nunca reemplaza.** Para reemplazar está
vaciar y añadir en la misma llamada, `biso set <ref> --clear-notes --append-note "..."`, que este
comando no acepta por su verbo propio pero sí como flags de campo de la sección ["Las familias de flags"](../familias-de-flags.md), igual
que las demás.

| Caso | Qué pasa |
|---|---|
| Sin ningún texto y sin ningún flag de campo | Error 2 |
| Texto vacío | No añade nada y avisa, según ["El valor vacío"](../valores-de-entrada.md#el-valor-vacío) |
| La tarea no tiene notas todavía | Se crean |
| Varios textos | Un párrafo por texto, en el orden dado |

### Salida

```
MYP-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Añadida | 0 |
| Sin texto, o un posicional que parece un identificador | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente, o fichero de `@` inexistente | 4 |
| Referencia ambigua | 5 |
| `--dry-run` que no pasa | 7 |
| El almacén falla | 8 |
| No hay tablero | 20 |

### `biso note --help`

```
Usage: biso note <ref> <text>... [options]

Append one or more paragraphs to the implementation notes of ONE task. It never
replaces anything; clearing and appending in the same `biso set <ref>` call does.

Arguments:
  ref            one task: an id, a bare number or free text
  text           one paragraph per argument; @file and - work here too

Options:
      --id / --match   force <ref> to be an id, or free text
  -h, --help           show this help

Every field flag of `biso set --help` works here too. Use `--append-note <text>`
for a paragraph that is not checked against the id grammar, for when the note
itself looks like an id.

To note the same thing on several tasks, use `biso set A B --append-note "..."`.

Exit codes:
  0  appended       3  the task could not be read    7  --dry-run did not pass
  2  bad usage      4  not found                      8  could not be written
                    5  ambiguous                     20  no board here

Examples:
  biso note MYP-11 "The parser already normalized LF, CRLF was missing"
  biso note 11 "First finding" "Second finding"
  biso note MYP-11 @/tmp/benchmark-output.txt
```

## `biso comment`

Un comentario y una nota son cosas distintas. Una nota de implementación es el registro técnico de
quien hace el trabajo. Un comentario es una conversación, tiene autor y fecha, y es el canal por el
que entra lo que viene de fuera.

### Firma

```
biso comment <ref> [<text>...] [--comment-author <@who>]
             [--id] [--match] [cualquier flag de campo de las familias de flags]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí, salvo con `--comment` | texto largo | | sí | no | |
| `--comment-author <@who>` | | no | texto libre | `me` | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Se aplican las mismas reglas de posicional que en `biso note`, incluida la del texto que parece un
identificador. El autor es texto libre, no se valida contra nada y no interpreta el `@` inicial. **Sin
`--comment-author` y sin [ninguna identidad configurada](../invocacion.md#variables-de-entorno), es error 2**: `error: --comment-author is
required, no identity is configured`. Lo mismo vale para `--comment` en cualquier otro comando de
escritura.

### Salida

```
MYP-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Y por stderr, `note: comment #2 by @trello:juan`. Ese `#2` es la `key` que acaba de recibir el
comentario (["Los comentarios"](../modelo-de-datos/comentarios.md#los-comentarios)), la misma que después acepta `--rm-comment` y `--set-comment-date`
(["Comentarios"](../familias-de-flags.md#comentarios)).

### Códigos de salida

Los mismos de `biso note`, más el error 2 de `--comment-author` sin identidad configurada.

### `biso comment --help`

```
Usage: biso comment <ref> <text>... [options]

Append a discussion comment to ONE task, with an author and a timestamp. This
is not `biso note`, which records what you did while implementing.

Arguments:
  ref                      one task: an id, a bare number or free text
  text                     one comment per argument; @file and - work too

Options:
      --comment-author <@who>  free text author (default: you). An external
                               system can use its own convention, such as
                               @trello:juan. The @ is never a file reference
      --id / --match           force <ref> to be an id, or free text
  -h, --help                   show this help

Every field flag of `biso set --help` works here too. Use `--comment <text>`
for a comment that is not checked against the id grammar.

A comment's body and author are never edited, by any flag. The whole comment
can be removed with --rm-comment, and only its date corrected with
--set-comment-date, both in `biso set --help`.

Exit codes:
  0  appended       3  the task could not be read    7  --dry-run did not pass
  2  bad usage      4  not found                      8  could not be written
                    5  ambiguous                     20  no board here

Examples:
  biso comment MYP-11 "A user with a Windows clone reported this"
  biso comment 11 "Moved to Doing from the phone" --comment-author @trello:avilches
```

## `biso finish`

### Firma

```
biso finish <ref>... [--append-summary <text>] [--check-ac <sel>]... [--check-dod <sel>]...
            [--append-note <text>]... [--add-files <path>]... [-s <v>] [--strict] [--no-checks]
            [--id] [--match] [cualquier flag de campo de las familias de flags]
```

### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--status <v>` | `-s` | no | vocabulario | `terminal_status` | no | no | |
| `--strict` | | no | booleano | el valor de `finish_strict` | no | no | `--no-checks` |
| `--no-checks` | | no | booleano | falso | no | no | `--strict` |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

`--append-summary`, `--check-ac`, `--check-dod`, `--append-note` y `--add-files` son los flags de
campo de siempre.

### Qué hace

Marca criterios y elementos de la definición de hecho, escribe la última nota y el resumen, apunta los
ficheros tocados y mueve al estado terminal, todo en una escritura.

| Caso | Qué pasa |
|---|---|
| Quedan criterios sin marcar y no se pasó `--check-ac` | **Se cierra igual**, con el aviso y la lista de los que faltan |
| Quedan elementos de la definición de hecho sin marcar | Igual, con su propio aviso |
| Lo mismo, con `--strict` | Error 6, y no se escribe nada |
| Sin `--summary` | Se cierra igual, con `warning: MYP-11 finished without a final summary` |
| Sin `--summary` y con `--strict` | Error 6 |
| La tarea tiene subtareas sin terminar | Aviso con la lista. Con `--strict`, error 6 |
| La tarea tiene [una pregunta abierta](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta) | Se cierra igual, con el aviso correspondiente. **Avisa, no impide, ni con `--strict`**: impedirlo empujaría a rodear la herramienta con `biso set` |
| La tarea ya estaba terminada | Se aplica el resto sin cambiar el estado, con un `note:` |
| La tarea tiene el arrendamiento vivo de otra identidad | Se cierra igual, con el aviso de ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) de que era de otra persona, y `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura. La invariante gana sobre el "no tocar los campos" de una escritura ajena, porque una tarea terminada con arrendamiento vivo es un tablero que su propia importación rechazaría (["El vaciado"](../lease.md#el-vaciado) de `lease.md`) |
| La tarea tiene el arrendamiento y `-s` la lleva a otro estado que tampoco es el activo | Los campos se vacían igual: lo que los sostiene es estar en el estado activo, no llegar al terminal |
| `--no-checks` | Se salta todas las comprobaciones y no emite ninguno de esos avisos, incluido el de la pregunta abierta |
| Varias referencias | Todo o nada |

Quien quiera la política dura tiene `--strict`, y puede fijarla por defecto con
`biso config set finish_strict true`.

### Salida

```
MYP-11  Done  ac 2/2  dod 1/1  urgency 0.0
```

La urgencia de una tarea en el estado terminal es cero por definición, según la sección ["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia).

Por stderr, cuando toca:

```
warning: MYP-11 moved to Done with 1 of 2 acceptance criteria unchecked
  #3 There is a test that covers it
```

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Cerrada | 0 |
| `--strict` y falta algo | 6, y no se escribe nada |
| Flags incompatibles, selector por clave con varias tareas | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Referencia o criterio inexistente | 4 |
| Referencia o criterio ambiguo | 5 |
| `--dry-run` que no pasa | 7 |
| El almacén falla | 8 |
| No hay tablero | 20 |

### `biso finish --help`

```
Usage: biso finish <ref>... [options]

Close one or more tasks: check criteria, add the last note, write the final
summary and move to the terminal status. One call.

Options:
      --append-summary <text>  add to the final summary; repeatable, takes
                               @file and -
      --check-ac <sel>         check criteria: all, 3, 1-4, 1,3,7 or the text.
                               With several tasks the selector has to be `all`
      --check-dod <sel>        the same for the definition of done
      --append-note <text>     one last implementation note; repeatable
      --add-files <path>       record a modified file; repeatable
  -s, --status <value>         use another status instead of the terminal one
      --strict           refuse to finish with unchecked criteria, unchecked
                         definition of done, unfinished subtasks or no summary
                         (default: warn and go on; see finish_strict)
      --no-checks        skip every check and every warning
      --id / --match     force <ref> to be an id, or free text
  -h, --help             show this help

Every field flag of `biso set --help` works here too.

Exit codes:
  0  finished       3  unknown value    6  --strict and something is missing
  2  bad usage      4  not found        7  --dry-run did not pass
                    5  ambiguous        8  the board could not be written
                                        20 no board here

Examples:
  biso finish MYP-11 --check-ac all --check-dod all --append-summary "Normalizes CRLF"
  biso finish MYP-11 --check-ac "covers CRLF" --append-note "313 tests green"
  biso finish MYP-11 MYP-12 --check-ac all --append-summary "Both closed by PR 42"
```

## `biso ask`

### Firma

```
biso ask <ref> <text>... [--id] [--match] [cualquier flag de campo de las familias de flags]
```

### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí | texto largo | | sí, como posicional | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Cada texto es un párrafo propio del cuerpo de la pregunta, igual que en [`biso note`](#biso-note). Acepta
`@fichero` y `-` como [cualquier texto largo](../valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).

Se aplica la misma regla del texto que parece un identificador que [`biso note`](#biso-note): un
posicional que encaja con la gramática de identificador de la sección ["La gramática"](../referencias.md#la-gramática) es error 2. Pero el mensaje
es propio, porque `biso ask` no tiene un flag de campo que escriba `question`, así que la única
salida es `@fichero` o `-`, nunca `--note`:

```
error: "MYP-2" looks like a task id, and `biso ask` takes only one task
hint: to write that text literally, use @file or - for stdin
```

**`biso ask` no acepta `--comment-author`**: el autor de la pregunta es siempre [la identidad
configurada](../invocacion.md#variables-de-entorno).

### Qué hace

Llena el campo [`question`](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta) con el autor y el instante que fija el programa y el texto dado,
junto con cualquier otro flag de campo que se haya pasado en la misma escritura. **No cambia el
estado de la tarea.**

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Se llena el campo. **No cambia el estado** |
| La tarea ya tiene una pregunta abierta | Error 6, para que la segunda no borre a la primera en silencio |
| La tarea está en el estado terminal | Error 6, igual que `biso start`, con la pista de reabrirla |
| La tarea está archivada | Se hace, con `note: MYP-11 is archived` por stderr, igual que `biso get` |
| El texto está vacío | Error 3: `error: the question cannot be empty`, `code` [`empty_scalar_value`](../valores-de-entrada.md#el-valor-vacío) |
| Sin [identidad configurada](../invocacion.md#variables-de-entorno) | Error 2: `error: biso ask needs an identity; set it with biso config set me <you> or BISO_ME` |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que [`biso note`](#biso-note) |
| Varias referencias | No se admiten: toma exactamente una, como `biso note` y `biso comment` |

Los errores 6 llevan pista:

```
error: MYP-11 already has an open question
hint: answer it first with `biso answer MYP-11 <text>`
```

```
error: MYP-11 is already Done
hint: reopen it first with `biso start MYP-11 --reopen`
```

### Salida

```
MYP-11  In Progress  ac 1/2  dod 0/1  urgency 15.0
```

La urgencia queda por debajo de los 19.0 del ejemplo de ["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia) porque el término de actividad exige
también que no haya pregunta abierta: la tarea sigue en el estado activo, pero `waiting` ya es
cierto.

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Preguntada | 0 |
| Ya hay una pregunta abierta, o la tarea ya está en el estado terminal | 6 |
| Referencia mal formada, flags incompatibles, posicional que parece un identificador | 2 |
| Sin identidad configurada | 2 |
| Pregunta vacía, tarea ilegible | 3 |
| Referencia inexistente, o fichero de `@` inexistente | 4 |
| Referencia ambigua | 5 |
| `--dry-run` que no pasa | 7 |
| El almacén falla | 8 |
| No hay tablero | 20 |

### `biso ask --help`

```
Usage: biso ask <ref> <text>... [options]

Park ONE task on a question for a person. The task keeps its status, but it
leaves the IN PROGRESS block of `biso prime` and shows up under NEEDS ANSWER
until somebody runs `biso answer`.

Arguments:
  ref                one task: an id, a bare number or free text
  text               the question; @file and - work too

Options:
      --id / --match force <ref> to be an id, or free text
  -h, --help         show this help

Every field flag of `biso set --help` works here too, but nothing except this
command and `biso answer` ever writes the question itself.

A task holds one open question at a time. Answer it before asking another.

Exit codes:
  0  asked          4  not found        7  --dry-run did not pass
  2  bad usage      5  ambiguous        8  could not be written
  3  empty question, or unreadable      20 no board here
  6  already asking, or already finished

Examples:
  biso ask MYP-11 "Do we normalize binary files too, or only text?"
  biso ask 11 @/tmp/question.md
```

## `biso answer`

### Firma

```
biso answer <ref> <text>... [--id] [--match] [cualquier flag de campo de las familias de flags]
```

### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí | texto largo | | sí, como posicional | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Cada texto es un párrafo propio de la respuesta, igual que en [`biso ask`](#biso-ask) y en [`biso note`](#biso-note). Acepta `@fichero` y `-` como [cualquier texto largo](../valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).

Se aplica la misma regla del texto que parece un identificador que [`biso note`](#biso-note): un
posicional que encaja con la gramática de identificador de la sección ["La gramática"](../referencias.md#la-gramática) es error 2. Pero el mensaje
es propio, porque `biso answer` no tiene un flag de campo que escriba `question`, así que la única
salida es `@fichero` o `-`, nunca `--note`:

```
error: "MYP-2" looks like a task id, and `biso answer` takes only one task
hint: to write that text literally, use @file or - for stdin
```

**`biso answer` no acepta `--comment-author`**: los dos comentarios que
escribe van siempre firmados por [la identidad configurada](../invocacion.md#variables-de-entorno). Quien necesite firmar un comentario
con otro autor tiene `biso comment --comment-author`, que sigue funcionando como siempre.

### Qué hace

Vacía el campo [`question`](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta) en una sola escritura, con tres efectos en este orden exacto:

1. Añade al [histórico de comentarios](../modelo-de-datos/comentarios.md#los-comentarios) uno con el `author`, el `askedAt` y el `body` que guardaba
   el campo: la pregunta se convierte literalmente en un comentario, con su autor y su [instante](../modelo-de-datos/fechas.md#las-fechas)
   originales.
2. Añade detrás un segundo comentario con el texto de la respuesta, firmado por [la identidad
   configurada](../invocacion.md#variables-de-entorno) y con el instante de ahora.
3. Vacía el campo.

**El orden no depende de la línea de comandos, según la regla de ["Orden de aplicación dentro de una escritura"](../garantias.md#orden-de-aplicación-dentro-de-una-escritura).** Dentro del paso de los
comentarios, los dos que escribe este verbo van siempre antes que cualquier `--comment` que se haya
pasado en la misma escritura. El vaciado del campo `question` es un paso propio de `biso answer`,
posterior a todos los de ["Orden de aplicación dentro de una escritura"](../garantias.md#orden-de-aplicación-dentro-de-una-escritura), y es siempre el último efecto de la escritura.

Como los [comentarios](../modelo-de-datos/comentarios.md#los-comentarios) se guardan y se muestran en orden de inserción y no de instante, el
comentario de la pregunta queda antes que el de la respuesta aunque su instante sea anterior, y el
instante de cada uno sigue diciendo la verdad.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Error 6, con la pista de usar `biso comment` |
| Falta el positional del texto | Error 2. Una respuesta sin respuesta no cierra nada |
| El texto está vacío (`biso answer MYP-11 ""`) | Error 3: `error: the answer cannot be empty`, `code` [`empty_scalar_value`](../valores-de-entrada.md#el-valor-vacío) |
| Sin [identidad configurada](../invocacion.md#variables-de-entorno) | Error 2: `error: biso answer needs an identity; set it with biso config set me <you> or BISO_ME` |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que [`biso note`](#biso-note) |
| Se pasan además flags de campo | Se aplican igual, como en cualquier verbo del ciclo |
| Varias referencias | No se admiten: toma exactamente una, como `biso note` y `biso comment` |
| La tarea está archivada | Se hace, con `note: MYP-11 is archived` por stderr, igual que `biso get` |
| La tarea está en el estado terminal | Se hace igual que en cualquier otro estado |

**No es simétrico con `biso ask`, y es a propósito.** `ask` sobre una tarea terminada es error 6,
porque no tiene sentido abrir una pregunta sobre algo que ya está cerrado. Pero responder una pregunta
que se quedó abierta al cerrar la tarea es la única vía de recuperación que existe: `biso finish`
[avisa](../salida-y-terminal.md#notas-y-avisos) sin impedirlo y esa pregunta desaparece de los bloques de `biso prime` y del `biso ls` por
defecto, riesgo que la sección ["Riesgos conocidos y aceptados del modelo de estados"](../../decisiones/modelo-de-estados.md#riesgos-conocidos-y-aceptados-del-modelo-de-estados) acepta a propósito. Impedir `answer` sobre una
tarea terminada cerraría esa única vía.

```
error: MYP-11 has no open question
hint: use `biso comment` to add a comment
```

### Salida

```
MYP-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

La urgencia recupera el término de actividad de ["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia), porque `waiting` vuelve a ser falso.

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Respondida | 0 |
| La tarea no tiene pregunta abierta | 6 |
| Referencia mal formada, flags incompatibles, posicional que parece un identificador, falta el texto | 2 |
| Sin identidad configurada | 2 |
| Respuesta vacía, tarea ilegible | 3 |
| Referencia inexistente | 4 |
| Referencia ambigua | 5 |
| `--dry-run` que no pasa | 7 |
| El almacén falla | 8 |
| No hay tablero | 20 |

### `biso answer --help`

```
Usage: biso answer <ref> <text>... [options]

Answer the open question of ONE task and unpark it. In a single write this
moves the question into the comments with its original author and time, adds
your answer behind it, and clears the question.

Arguments:
  ref                one task: an id, a bare number or free text
  text               the answer; @file and - work too

Options:
      --id / --match force <ref> to be an id, or free text
  -h, --help         show this help

Every field flag of `biso set --help` works here too, so you can answer and
refine in one call.

Both comments are signed with your configured identity. This command does not
take --comment-author.

Exit codes:
  0  answered       4  not found        7  --dry-run did not pass
  2  bad usage      5  ambiguous        8  could not be written
  3  empty answer, or unreadable        20 no board here
  6  no open question

Examples:
  biso answer MYP-11 "Only text files. Binary ones are skipped entirely."
  biso answer 11 "Yes" --add-ac "A binary file is never touched"
```

---

