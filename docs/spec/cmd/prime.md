# `biso prime`, el arranque de una sesión

## Qué resuelve este comando

`biso` **no escribe nunca fuera del tablero, salvo el puntero del proyecto de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md).** No
modifica ningún otro fichero del proyecto, ni al crear el tablero ni nunca. Todo lo que hace falta
para empezar a trabajar cabe en un solo comando, cuya salida es un solo mensaje.

`biso prime` se ejecuta al empezar la sesión, de la manera que convenga a quien lo use: como una
orden del asistente, desde un fichero de memoria del proyecto que diga en una línea "ejecuta
`biso prime` antes de tocar tareas", o a mano.

**El criterio de diseño del mensaje es exigente y está pensado para poderse comprobar:** quien lo lea
y no haya visto nunca la herramienta tiene que poder completar un ciclo de trabajo entero, desde
crear una tarea hasta cerrarla, sin leer nada más y sin ninguna interacción adicional.

## Firma

```
biso prime [--full] [--limit <n>] [--json]
```

## Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--full` | | no | booleano | falso | no | no | `--json` |
| `--limit <n>` | | no | entero >= 0 | 5 | no | no | ninguno |

- `--limit` acota juntas las secciones `ASSIGNED TO YOU` y `NEXT UP`: su valor son filas repartidas
  entre las dos, en ese orden de preferencia, con una sola línea de recuento al final de la última que
  se imprima. Con `0`, las dos desaparecen y se queda solo esa línea.
- `--full` añade al final la lista completa de banderas de `biso new` y `biso set`. Es para una
  persona que está aprendiendo la herramienta, no para el arranque de un agente.
- `--json` es la bandera global de la sección ["Banderas globales"](../invocacion.md#banderas-globales), y aquí es lo único que la restringe: no se puede
  combinar con `--full`, porque el JSON no lleva texto de ayuda.

## Qué hace, caso a caso

| Situación | Qué pasa |
|---|---|
| Hay tablero y tiene tareas | Imprime el mensaje de la sección ["La salida literal"](#la-salida-literal) por stdout, código 0 |
| Hay tablero y está vacío | Igual, con los cuatro bloques de tareas sustituidos por las tres líneas de la sección ["Tablero vacío"](#tablero-vacío) |
| No hay tablero | Código 8, y por stderr el mensaje de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) |
| Alguna tarea no se puede leer | El mensaje sale igual, con el aviso de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), código 0 |
| `--limit` negativo | Código 2 |

`biso prime` **no escribe nada, nunca**, y no necesita acceso exclusivo. Es seguro llamarlo en
paralelo desde varias sesiones y mientras otro proceso escribe.

## Qué entra en el mensaje y qué se relega a `--help`

El criterio es uno solo: **entra lo que no se puede adivinar y hace falta para la primera acción; se
queda fuera lo que se puede consultar en el momento exacto en que hace falta.**

Entra:

- Las diez órdenes del ciclo de trabajo con su forma de uso. Quien no sabe que existe `biso finish`
  no va a escribir `biso finish --help`.
- **Los nombres de todas las banderas de campo**, en una rejilla de cinco líneas.
- El vocabulario real de este tablero, con **el recuento por estado** y con la marca de cuál es el
  estado de las tareas nuevas, cuál el activo y cuál el terminal.
- Las reglas que no son adivinables.
- Los códigos de salida, en dos líneas.
- El estado del tablero: lo que está en curso y lo más urgente de lo que no ha empezado.

Se queda fuera, y va a `biso <cmd> --help`:

- Los valores, los tipos y las incompatibilidades de cada bandera. El mensaje da los nombres, que es
  lo que no se puede adivinar; la ayuda da el detalle, que es lo que se consulta cuando se necesita.
- El formato de lote de `biso new --from` y el esquema JSON completo.
- `biso export`, `biso config`, `biso doctor`, `biso archive`, `biso where` y `biso board`, que no
  aparecen en el ciclo de trabajo normal.
- Todos los casos límite: el rango invertido, pedir la entrada estándar más de una vez, la coma
  dentro de una etiqueta.
- La política de cuándo merece la pena crear una tarea, que es una decisión del proyecto y no de la
  herramienta. El mensaje la resume en una línea y no la desarrolla.

## La salida literal

Esto es exactamente lo que `biso prime` imprime por stdout con un tablero de ejemplo. No imprime nada
por stderr.

**Ese tablero fija `task_prefix` a `MYP` explícitamente**, en vez de dejar que se derive de
`project_name` como haría por defecto (la sección ["Identificadores"](../modelo-de-datos.md#identificadores)), para que los identificadores de todos los ejemplos
de este documento no dependan del nombre que le toque al tablero de turno. De paso queda demostrado que
`task_prefix` se puede fijar a mano.

```
biso 1.0.0 - the task board of this project. This message is all you need to start.

BOARD  My project
  To Do 54 | In Progress 4 | Done 190
  new tasks start in To Do; `biso start` moves to In Progress; `biso finish` to Done
  types       idea, memory, task, bug, docs
  priorities  high, medium, low
  you are     @claude

COMMANDS  (`biso <cmd> --help` for the detail of any flag)
  biso ls [-s STATUS] [--type T] [-l LABEL] [--mine] [--search TEXT]
  biso get <ref> [--section ac]
  biso new "TITLE" [-d TEXT] [--add-ac TEXT]... [--type T] [--priority P]
  biso start <ref>... [--plan TEXT]
  biso note <ref> "TEXT"
  biso ask <ref> "QUESTION"
  biso answer <ref> "TEXT"
  biso finish <ref>... [--summary "TEXT"] [--check-ac all] [--check-dod all]
  biso set <ref>... [any field flag]
  biso comment <ref> "TEXT" [--comment-author @who]

FIELD FLAGS  (same names, same meaning, in every command above that writes)
  -t --title  -s --status  --type --clear-type  --priority --clear-priority
  --project --clear-project  -m --milestone --clear-milestone
  -p --parent --clear-parent  --due --clear-due  --ordinal --clear-ordinal  --reporter --clear-reporter
  -l --add-labels --rm-labels --clear-labels --replace-labels
  -a --add-assignees --rm-assignees --clear-assignees --replace-assignees
  --add-refs --rm-refs --clear-refs --replace-refs
  --add-docs --rm-docs --clear-docs --replace-docs
  --add-deps --rm-deps --clear-deps --replace-deps
  --add-files --rm-files --clear-files --replace-files
  --add-ac --rm-ac --clear-acs   --add-dod --rm-dod --clear-dods
  --check-ac --uncheck-ac --check-dod --uncheck-dod
  -d --append-desc --clear-desc  --append-plan --clear-plan
  --append-note --clear-notes  --append-summary --clear-summary
  --comment --rm-comment --set-comment-date
  --ext K=V --rm-ext --clear-ext

RULES  (none of these are guessable; they are the whole learning curve)
  1. Every write goes through biso. Nothing else touches the board.
  2. <ref> is an id (MYP-12), a bare number (12) or free text ("CRLF"). Text
     matching several tasks is an error that lists them, never a guess. `note`,
     `comment`, `ask` and `answer` take one <ref>; `set`, `start` and `finish`
     take several.
  3. Filters reject values this board does not have: `-s Pending` is an error,
     not an empty list. Case, spaces, hyphens and underscores are ignored, so
     `-s todo`, `-s "To Do"` and `-s TO_DO` are one and the same filter. An
     empty list is therefore a fact about the board that you can act on.
  4. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says
     on stderr what it left out. --all lifts the limit, --any-status includes
     Done, --archived reaches the archive.
  5. --check-ac, --uncheck-ac, --check-dod and --uncheck-dod take all, 3, 1-4, 1,3,7 or the
     criterion text. The numbers are the stable #N keys that `biso get` shows, and they never
     shift when one criterion is removed.
  6. `biso new` prints the new id and nothing else. Every other write prints one
     line per task: id, status, criteria, urgency. Add --print for the whole
     record, or --json for a versioned envelope.
  7. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.
  8. Long text: a real newline works, and so do -d @file.md and -d - for stdin.
  9. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,
     6 precondition not met, 7 environment, 8 no board here, 9 nothing written.
 10. `biso ask <ref> "..."` parks a task on a question and `biso answer` unparks
     it, writing both into the comments. Ask instead of guessing. A task
     assigned to you is one a person decided you should do.

IN PROGRESS
  MYP-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
  MYP-52  In Progress  task  low     Document the release checklist                    ac 0/1  -        -
    lease expired 2026-09-05T09:00:00Z, was held by @bob
  MYP-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -

NEEDS ANSWER
  MYP-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
    Should the retry budget be shared with the download endpoint or kept separate?

ASSIGNED TO YOU
  MYP-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
  MYP-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -

NEXT UP  (not assigned to you, by urgency)
  MYP-7   To Do        bug   high    Crash on an empty repository                      ac 0/4  -        2026-09-08
  MYP-19  To Do        task  high    Retry the upload on 5xx                           ac 0/2  -        -
  MYP-44  To Do        bug   low     Wrong column width on narrow ttys                 ac 0/1  -        -
  49 more not shown: `biso ls --not-active --not-waiting`

Pick one, `biso start <ref> --plan "..."`, work, `biso note <ref> "..."` as you go,
and close with `biso finish <ref> --check-ac all --check-dod all --summary "..."`.
That is the loop. Create a task when the work needs planning or review; do small
edits directly.
```

La cadena de versión de la primera línea es la que tenga instalada quien ejecute el ejemplo, y el
contrato de estabilidad permite que cambie entre versiones, así que puede no ser exactamente esta.

Cómo se calcula el resumen, para que la implementación sea única:

- La línea de recuento del bloque `BOARD` tiene **un número por cada estado configurado**, en el
  orden en que están configurados, y cuenta las tareas no archivadas de ese estado. No hay ninguna
  categoría inventada como "abiertas" que no se corresponda con un estado del tablero.
- Los cuatro bloques `IN PROGRESS`, `NEEDS ANSWER`, `ASSIGNED TO YOU` y `NEXT UP` se reparten
  el tablero por esta precedencia, y cada tarea cae en el primero que la acepte:
    1. `NEEDS ANSWER`, si tiene una pregunta abierta.
    2. `IN PROGRESS`, si está en el estado activo.
    3. `ASSIGNED TO YOU`, si está asignada a la identidad configurada.
    4. `NEXT UP`, el resto.
  **Ninguna tarea aparece en dos bloques.** Una tarea aparcada no sale en `IN PROGRESS` aunque esté en
  el estado activo, porque ese bloque significa que alguien está trabajando y ahí no lo está nadie.
  Los cuatro excluyen las tareas terminadas y las archivadas.
- Se imprimen en este orden: `IN PROGRESS`, `NEEDS ANSWER`, `ASSIGNED TO YOU` y `NEXT UP`.
  **Un bloque sin filas no se imprime**, ni siquiera su encabezado. **Sin identidad configurada,
  `ASSIGNED TO YOU` no se imprime nunca**, aunque el resto del mensaje se imprime igual, con
  `you are (not set)` en el bloque `BOARD`.
- `IN PROGRESS` lista las tareas del estado activo sin pregunta abierta, ordenadas por la regla de
  orden de la sección ["`biso ls`"](ls.md), sin límite. Cada tarea cuyo arrendamiento está vencido (el campo derivado
  `leaseExpired` de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md)) lleva, igual que `NEEDS ANSWER` con su pregunta, una segunda línea
  indentada con la forma `lease expired <leaseExpiresAt>, was held by <leaseHolder>`. Es el único de
  los cuatro bloques que la lleva, porque es el único cuya etiqueta afirma que alguien está
  trabajando ahora mismo, y un arrendamiento vencido contradice justo esa afirmación. Esta línea, como
  la de la pregunta, no cuenta para el ancho de las columnas. El hecho que la provoca sí viaja en el
  esquema JSON de la sección ["El esquema JSON"](#el-esquema-json), como el campo `leaseExpired`, y sus dos detalles no: quien los quiera los pide
  con `biso get`, que los imprime en su línea `lease` (["`biso get`"](get.md)), igual que pide el cuerpo de la pregunta.
- `NEEDS ANSWER` lista las tareas con pregunta abierta, ordenadas igual, sin límite. Cada tarea
  ocupa **dos líneas**: la fila de siempre, con las ocho columnas del algoritmo de `biso ls`, y debajo
  una línea indentada con la pregunta recortada a **100 celdas**, con la misma regla exacta que el
  algoritmo aplica a los títulos. Los saltos de línea reales del cuerpo se sustituyen por un espacio
  antes de recortar, y el recorte usa el mismo sufijo `...` que la sección ["`biso ls`"](ls.md) usa para los títulos. Va en línea
  propia y no en una novena columna, porque el algoritmo tiene ocho exactas y una regla que dice que
  la octava nunca se rellena.
- `ASSIGNED TO YOU` lista las tareas asignadas a la identidad configurada que no estén ya en
  `NEEDS ANSWER` ni en `IN PROGRESS`, ordenadas igual. `NEXT UP` lista el resto, ordenadas
  igual.
- `ASSIGNED TO YOU` y `NEXT UP` **comparten el límite de `--limit`**: su valor son filas repartidas
  entre las dos, en ese orden de preferencia, con una sola línea de recuento al final de la última que
  se imprima. Con `--limit 0` desaparecen los dos y queda solo esa línea. Si cada bloque tuviera su
  propio límite, el resumen crecería al doble sin que `--limit` lo notara.
- La línea de recuento dice cuántas tareas quedan fuera de `ASSIGNED TO YOU` y `NEXT UP` juntas por el
  corte.
- Las filas usan exactamente el algoritmo de columnas de `biso ls` de la sección ["`biso ls`"](ls.md), con una
  diferencia declarada aquí: el ancho de las columnas 1 a 7 se calcula sobre las filas de los cuatro
  bloques juntas, **sin contar las líneas de pregunta ni las de arrendamiento vencido**, que no son
  filas de la tabla, para que los cuatro bloques se lean como una sola tabla.
- `NEXT UP` es lo que no cae en ninguno de los tres bloques anteriores, no "lo que no ha empezado".
  Por eso su rótulo es `NEXT UP  (not assigned to you, by urgency)`, su línea de recuento tiene la
  forma `N more not shown: 'biso ls --not-active --not-waiting'`, y su clave en el esquema JSON del
  apartado ["El esquema JSON"](#el-esquema-json) es `hiddenCount`.

## Tablero vacío

Cuando no hay ninguna tarea, los cuatro bloques de tareas de la sección ["La salida literal"](#la-salida-literal) se sustituyen por esto, sin imprimir
ninguno de sus encabezados, y el resto del mensaje no cambia:

```
THE BOARD IS EMPTY
  Create the first one:
  biso new "Title" -d "What and why" --ac "How we will know it works"
```

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "prime",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "tool": { "name": "biso", "version": "1.0.0" },
    "board": {
      "name": "My project",
      "me": "@claude",
      "statuses": ["To Do", "In Progress", "Done"],
      "initialStatus": "To Do",
      "activeStatus": "In Progress",
      "terminalStatus": "Done",
      "types": ["idea", "memory", "task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "extensions": ["trello.card"],
      "countByStatus": { "To Do": 54, "In Progress": 4, "Done": 190 }
    },
    "inProgress": [
      { "id": "MYP-11", "title": "Normalize CRLF in the diff", "status": "In Progress",
        "type": "bug", "priority": "high", "assignees": ["@claude"], "due": null,
        "acDone": 1, "acTotal": 2, "urgency": 19.0, "leaseExpired": false }
    ],
    "needsAnswer": [
      { "id": "MYP-60", "title": "Confirm the retry budget for the upload endpoint",
        "status": "In Progress", "type": "task", "priority": "high", "assignees": ["@claude"],
        "due": null, "acDone": 0, "acTotal": 2, "urgency": 15.2, "leaseExpired": false }
    ],
    "assignedToYou": [
      { "id": "MYP-61", "title": "Rewrite the install section", "status": "To Do",
        "type": "docs", "priority": "medium", "assignees": ["@claude"], "due": null,
        "acDone": 0, "acTotal": 1, "urgency": 12.4, "leaseExpired": false }
    ],
    "nextUp": [
      { "id": "MYP-7", "title": "Crash on an empty repository", "status": "To Do",
        "type": "bug", "priority": "high", "assignees": [], "due": "2026-09-08",
        "acDone": 0, "acTotal": 4, "urgency": 18.2, "leaseExpired": false }
    ],
    "hiddenCount": 49
  }
}
```

Los valores de urgencia del ejemplo salen de los coeficientes por defecto, que el contrato de
estabilidad permite cambiar entre versiones menores, así que las cifras exactas pueden no ser estas.

Las reglas y los nombres de las banderas no viajan en el JSON: quien pide JSON es un programa, y un
programa no necesita que le expliquen en prosa cómo se nombran las banderas de escritura.

**Ninguna de las cuatro listas trae el cuerpo de la pregunta**, por el mismo motivo que el esquema de
`task.list` en la sección ["`biso ls`"](ls.md) no trae el cuerpo de la tarea: es texto largo. Lo que sí llevan es la posición de
cada tarea (en `needsAnswer` o en cualquier otro de los cuatro bloques), que ya dice si está
aparcada, igual que el campo derivado `waiting` de `task.list`. Quien necesite leer la pregunta usa
`biso get --section question`.

**`leaseExpired` sale en los cuatro bloques**, y en `assignedToYou` y en `nextUp` vale siempre `false`.
Esos dos bloques no pueden contener ninguna tarea en el estado activo, que es la única clase de tarea
en la que un arrendamiento puede existir: la precedencia de la sección ["La salida literal"](#la-salida-literal) manda toda tarea activa a `inProgress`
o a `needsAnswer`. Quien implemente puede apoyarse en ese valor constante, pero la clave se emite igual,
porque la regla de la sección ["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias) prohíbe que una clave aparezca o desaparezca según los datos y quien lee esta
salida no sabe de antemano en qué bloque va a caer una tarea. Va en el JSON aunque la segunda línea
indentada del texto salga solo en `inProgress`, porque no es texto largo y esconderlo obligaría a quien
consume JSON a llamar a `biso get` tarea por tarea para saber algo que el mensaje de texto ya enseña.
Sus dos detalles, `leaseExpiresAt` y `leaseHolder`, no salen aquí: para eso está `task.list` (["`biso ls`"](ls.md)), y
en texto la línea `lease` de la ficha de `biso get` (["`biso get`"](get.md)).

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Mensaje impreso | 0 |
| Mensaje impreso con alguna tarea ilegible | 0, con aviso |
| `--limit` negativo, o `--full` junto con `--json` | 2 |
| No hay tablero | 8 |

## `biso prime --help`

```
Usage: biso prime [options]

Print everything needed to start working on this board: the commands, the field
flags, the rules that are not guessable, the board vocabulary and what is in
flight. Run it once at the start of a session. It writes nothing.

Options:
  --full          also list every flag of `biso new` and `biso set` in detail
  --limit <n>     rows shown across ASSIGNED TO YOU and NEXT UP together
                  (default 5, 0 hides both)
  --json          machine-readable envelope instead of the message
  -h, --help      show this help

Exit codes:
  0  message printed
  2  bad usage
  8  no board here

Examples:
  biso prime
  biso prime --limit 10
  biso -C ~/work/my-project prime
```

---

