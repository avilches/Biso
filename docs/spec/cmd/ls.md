# `biso ls`

## Firma

```
biso ls [-s <status>]... [--not-status <status>]... [--any-status] [--archived] [--only-archived]
        [--type <v>]... [--priority <v>]...
        [-l <label>]... [--label-or <label>]... [-a <@who>]... [--mine] [--unassigned]
        [-p <ref>] [--blocked] [--not-blocked] [--waiting] [--not-waiting]
        [--active] [--not-active] [--overdue] [--due-before <date>]
        [--search <text>] [--unchecked]
        [--sort <field>] [--reverse] [--limit <n>] [--all] [--ids] [--count]
```

## Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--status <v>` | `-s` | no | vocabulario | todos menos el terminal | sí | sí | `--any-status` |
| `--not-status <v>` | | no | vocabulario | | sí | sí | `--any-status` |
| `--any-status` | | no | booleano | falso | no | no | `-s`, `--not-status` |
| `--archived` | | no | booleano | falso | no | no | `--only-archived` |
| `--only-archived` | | no | booleano | falso | no | no | `--archived` |
| `--type <v>` | | no | vocabulario | | sí | sí | |
| `--priority <v>` | | no | vocabulario | | sí | sí | |
| `--label <l>` | `-l` | no | etiqueta | | sí | sí | |
| `--label-or <l>` | | no | etiqueta | | sí | sí | |
| `--assignee <@w>` | `-a` | no | persona | | sí | sí | `--mine`, `--unassigned` |
| `--mine` | | no | booleano | falso | no | no | `-a`, `--unassigned` |
| `--unassigned` | | no | booleano | falso | no | no | `-a`, `--mine` |
| `--parent <ref>` | `-p` | no | referencia | | no | no | |
| `--blocked` | | no | booleano | falso | no | no | `--not-blocked` |
| `--not-blocked` | | no | booleano | falso | no | no | `--blocked` |
| `--waiting` | | no | booleano | falso | no | no | `--not-waiting` |
| `--not-waiting` | | no | booleano | falso | no | no | `--waiting` |
| `--active` | | no | booleano | falso | no | no | `--not-active` |
| `--not-active` | | no | booleano | falso | no | no | `--active` |
| `--overdue` | | no | booleano | falso | no | no | |
| `--due-before <d>` | | no | `YYYY-MM-DD` | | no | no | |
| `--search <text>` | | no | texto libre | | no | no | |
| `--unchecked` | | no | booleano | falso | no | no | |
| `--sort <field>` | | no | `urgency`, `id`, `ordinal`, `due`, `updated`, `created`, `title` | el orden de abajo | no | no | |
| `--reverse` | | no | booleano | falso | no | no | |
| `--limit <n>` | | no | entero >= 0 | 30 | no | no | `--all` |
| `--all` | | no | booleano | falso | no | no | `--limit` |
| `--ids` | | no | booleano | falso | no | no | `--count` |
| `--count` | | no | booleano | falso | no | no | `--ids` |

Reglas de combinación de filtros:

- **Filtros de campos distintos se combinan con `y`.** `-s "To Do" --type bug` son las que cumplen las
  dos cosas.
- **Valores repetidos del mismo campo se combinan con `o`.** `--type bug --type docs` son las de
  cualquiera de los tipos. Esto vale para `--status`, `--type`, `--priority`, `--assignee` y
  `--label-or`.
- **`-l/--label` es la única que se combina con `y`.** `-l frontend -l bug` son las que llevan las
  dos. Para el `o` está `--label-or`, que valida igual.
- **`--unchecked` apaga la comprobación de existencia de `-l`, `--label-or` y `-a`, y ninguna otra.**
  No cambia cómo se combinan ni afecta a ningún otro filtro. Los vocabularios configurados siguen
  validando, y `-p/--parent` sigue resolviendo su referencia.
- **Sin `-s` explícito, el estado terminal se excluye salvo `--any-status`; con `-s` explícito se
  filtra por ese valor tal cual, terminal incluido.** La exclusión del terminal es el comportamiento
  del valor por defecto de `-s`, no una regla aparte que se superponga a él: `biso ls -s Done`
  devuelve las tareas `Done`, exactamente como pide cualquier otro valor de `-s`. `--not-status <x>`
  sin `-s` sigue restando sobre la base por defecto (todos menos el terminal), así que por sí solo no
  reintroduce el terminal: `biso ls --not-status "To Do"` excluye `"To Do"` y sigue sin traer `Done`.
  `--any-status` sigue siendo la única forma de traer el terminal sin nombrarlo con `-s`.
- **Las archivadas se excluyen por defecto.** `--archived` las añade a las vivas y `--only-archived`
  deja solo las archivadas.
- **`--blocked` es incompatible con `--not-blocked`.** Los dos miran las dependencias sin terminar y
  no el estado, así que se combinan con cualquier filtro de estado y con los dos pares de abajo.
  `--not-blocked` por sí solo no dice que la tarea se pueda coger: descarta la que espera a otra
  tarea, no la que espera una respuesta ni la que ya lleva alguien. **Una tarea archivada sin
  terminar cuenta como terminada aquí**, igual que en la urgencia (["La
  urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)): no hace `--blocked` a quien depende de
  ella.
- **`--overdue` es `dias < 0`**, la misma cuenta que usa el término `proximidad` de la urgencia
  (["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)): una tarea que vence hoy tiene
  `dias = 0` y no es `--overdue`.
- **`--waiting` es incompatible con `--not-waiting`, y `--active` con `--not-active`, cada uno con su
  opuesto.** `--active` y `--not-active` filtran por el papel del estado y no por su nombre, que es su
  razón de ser: sin ellos, pedir la cola activa obligaría a escribir `-s "In Progress"`, el nombre
  concreto de un tablero concreto, y la misma consulta dejaría de servir en otro. Los cuatro son
  compatibles con `-s`, con `--not-status` y con `--any-status`, porque filtran sobre el mismo eje sin
  contradecirse: `-s "To Do" --active` es una lista vacía en unos tableros y no en otros. La regla
  general: **dos filtros que se contradicen por construcción son incompatibles. Una combinación de
  filtros válidos que resulte vacía en este tablero es un hecho legítimo sobre el tablero, no un
  error.**

## La regla de orden, completa

Sin `--sort` se aplica el orden por defecto, que es esta tupla, en este orden y sin excepciones:

1. Las tareas que tienen `ordinal` van antes que las que no lo tienen.
2. Entre las que lo tienen, `ordinal` ascendente.
3. Entre las que no lo tienen, `urgency` descendente.
4. Cualquier empate se rompe por identificador ascendente, siempre.

Un `--sort` explícito sustituye los pasos 1 a 3 por ese campo, ascendente salvo `urgency`, que es
descendente por ser una medida de prioridad, y el paso 4 se sigue aplicando. `--reverse` invierte el
resultado final, incluido el desempate. **El orden nunca depende del estado**, porque el listado no
agrupa por estado.

Un `--sort due` o `--sort ordinal` sobre tareas que no tienen ese campo las pone al final, en bloque,
ordenadas por identificador.

`--sort title` compara los títulos por sus puntos de código Unicode, de menor a mayor, y no por las
reglas de intercalación de ningún idioma: una `Z` va antes que una `a`, y `ñ` va después de `z`. Es
la única comparación de texto del orden, y se fija así porque una intercalación local haría que el
mismo tablero se ordenara distinto según la configuración regional de la máquina, que es justo lo
que ["Los principios"](../principios.md) no admite. Para un título mal ordenado a ojo está
`--sort ordinal`, que es lo que existe para decidir un orden a mano.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Filtro con un valor fuera del vocabulario | Error 3, con la lista de válidos |
| `-l` con una etiqueta o `-a` con una persona que el tablero no tiene | Error 3, con hasta cinco de las más parecidas, igual que en la sección ["Qué valida cada filtro, y contra qué"](../vocabularios.md#qué-valida-cada-filtro-y-contra-qué) |
| Lo mismo con `--unchecked` | Se acepta, y probablemente no devuelve nada |
| Filtro válido sin resultados | Ninguna línea por stdout, `note: no tasks match` por stderr, código **0** |
| Hay más resultados que el límite | Se imprimen los primeros y sale el aviso de recorte |
| `--limit 0` | No imprime ninguna fila, solo el aviso de recorte con el total. Es la forma de contar sin `--count` |
| `--count` | Un número por stdout y nada más: ni filas, ni aviso de recorte, ni la nota de "sin resultados" |
| `--ids` | Identificadores, uno por línea, sin cabeceras ni columnas. Es un listado, así que el límite y su aviso de recorte se aplican igual que con las columnas |
| Alguna tarea ilegible | Se salta, con el aviso de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), y el resto del listado es válido |

## Salida

Ocho columnas fijas, separadas por dos espacios, en este orden y con estos contenidos:

| Columna | Contenido | Cuando está vacío |
|---|---|---|
| 1 | identificador | nunca lo está |
| 2 | estado | nunca lo está |
| 3 | tipo | `-` |
| 4 | prioridad | `-` |
| 5 | título, recortado a **100 celdas siempre**, con `...` al final si se recorta | nunca lo está |
| 6 | `ac <marcados>/<total>` | `-` si la tarea no tiene criterios de aceptación |
| 7 | primera persona asignada, con `+<n>` si hay más | `-` |
| 8 | fecha límite | `-` |

**El formato se calcula así, en dos pasos, siempre en este orden:**

1. El título de cada tarea se recorta primero a 100 celdas, **contando los tres puntos**, así que la
   cadena que se imprime no pasa nunca de 100: un título más largo deja 97 celdas suyas y `...`
   detrás. Esto pasa antes de calcular ningún ancho de columna.
2. Para cada una de las columnas 1 a 7, el ancho de esa columna es la anchura del valor más largo
   que le corresponde entre las filas que se van a imprimir en esta llamada, y cada valor se rellena
   con espacios a la derecha hasta ese ancho. **La columna 8 nunca se rellena**, porque es la última
   y no hay nada después que alinear.

**La unidad de los dos pasos es la celda de un terminal monoespaciado, no el carácter.** Un título es
texto libre en UTF-8 (["Codificación y texto"](../salida-y-terminal.md#codificación-y-texto)), así que puede llevar ideogramas, emoji o acentos combinantes, y esas tres
cosas ocupan en pantalla algo distinto de lo que suman sus puntos de código: una marca combinante mide
cero celdas porque se pinta sobre la letra anterior, un ideograma de Asia oriental o un emoji miden dos,
y todo lo demás mide una. La tabla que lo dice es la de Unicode, la de anchura de Asia oriental más la
categoría de las marcas combinantes, y contar en cualquier otra unidad desalinea la tabla en cuanto un
título deja de ser ASCII. **Los caracteres de formato (categoría Cf) miden también cero**, por el
mismo motivo que una marca combinante: no se dibujan. El caso que importa es el juntador de ancho
cero (U+200D) que une las piezas de un emoji compuesto, que no ocupa ninguna celda y que contarlo
como una desalinearía la fila.

**Medir en celdas no es mirar el terminal, así que no contradice la sección ["Interactividad, terminal y color"](../salida-y-terminal.md#interactividad-terminal-y-color).** La anchura de un
carácter es una propiedad de Unicode, la misma en cualquier máquina y con cualquier ventana, y por eso
la salida sigue sin depender de dónde se ejecute el programa. Lo que la sección ["Interactividad, terminal y color"](../salida-y-terminal.md#interactividad-terminal-y-color) prohíbe es lo otro: preguntar
cuántas columnas tiene la ventana, o si hay color, para decidir qué se imprime.

**Y el recorte nunca parte un grafema por la mitad.** Si cortar exactamente en la celda 97 separaría una
letra de su acento combinante, o partiría un emoji compuesto, se corta en la frontera anterior, así que
el título recortado puede medir 96 o 95 celdas en vez de 97. La promesa es el tope, nunca la longitud
exacta: la cadena impresa no pasa de 100 celdas.

Entre columna y columna van siempre **dos espacios literales**, se haya rellenado o no la columna
anterior. Con estas cuatro tareas, el título más largo mide 28 caracteres y por eso la columna 5 se
rellena a ese ancho, no a uno fijo:

```
MYP-7   To Do        bug   high    Crash on an empty repository  ac 0/4  -        2026-09-08
MYP-11  In Progress  bug   high    Normalize CRLF in the diff    ac 1/2  @claude  -
MYP-19  To Do        task  high    Retry the upload on 5xx       ac 0/2  -        -
MYP-23  To Do        docs  medium  Rewrite the install section   ac 0/1  @sara+1  -
```

Y por stderr, siempre que se haya recortado:

```
warning: 28 more tasks match; showing 30 of 58
hint: narrow with -s, --type or -l, or ask for everything with --all
```

Con `--ids`:

```
MYP-7
MYP-11
```

Con `--count`:

```
58
```

**`--count` no recorta nada, porque no imprime ninguna fila.** El número que da es el de las tareas
que encajan, no el de las que se habrían impreso, así que `--limit` no tiene nada que cortar y no hay
aviso de recorte que emitir: la llamada contesta `matched` y ya está. En el sobre JSON eso se ve
igual, con `data.tasks` vacío, `shown` y `hidden` a cero y `truncated` en `false`. `--ids`, en
cambio, sí es un listado: imprime las mismas filas que las columnas, con el mismo límite y el mismo
aviso, solo que con una columna en vez de ocho.

**No hay agrupación por estado.** El estado es una columna más, para que cada línea se pueda tratar
igual que las demás.

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.list",
  "generatedAt": "2026-09-06T13:26:41Z",
  "data": {
    "tasks": [
      {
        "id": "MYP-11",
        "title": "Normalize CRLF in the diff",
        "status": "In Progress",
        "type": "bug",
        "priority": "high",
        "assignees": ["@claude"],
        "author": "@avilches",
        "labels": ["parser"],
        "parent": null,
        "dependencies": [],
        "references": ["docs/bugs/BUG-02.md"],
        "documentation": [],
        "modifiedFiles": [],
        "due": null,
        "ordinal": null,
        "createdAt": "2026-09-06T09:12:04Z",
        "updatedAt": "2026-09-06T11:40:18Z",
        "leaseExpiresAt": "2026-09-06T15:40:18Z",
        "leaseHolder": "@claude",
        "acDone": 1,
        "acTotal": 2,
        "commentCount": 1,
        "urgency": 19.0,
        "blocks": ["MYP-40"],
        "blocked": false,
        "waiting": false,
        "leaseExpired": false,
        "archived": false,
        "ext": { "trello.card": "5f2a8c1e3b9d4a7f6e0c2b81" }
      }
    ],
    "shown": 30,
    "matched": 58,
    "hidden": 28,
    "truncated": true,
    "skipped": [],
    "sort": "default",
    "filters": {
      "status": ["To Do", "In Progress"],
      "notStatus": [],
      "anyStatus": false,
      "archived": false,
      "onlyArchived": false,
      "type": [],
      "priority": [],
      "label": [],
      "labelOr": [],
      "assignee": [],
      "unassigned": false,
      "parent": null,
      "blocked": null,
      "waiting": null,
      "active": null,
      "overdue": false,
      "dueBefore": null,
      "search": null,
      "unchecked": false
    },
    "warnings": [ { "code": "list_truncated", "shown": 30, "matched": 58 } ]
  }
}
```

**`truncated` sigue siendo la forma de saber que se recortó sin tener que mirar `warnings`**, pero
cuando `truncated` es `true`, `data.warnings` lleva el mismo `list_truncated` de la lista general de
avisos (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)), con la misma regla que cualquier
otro comando: el texto humano sale por stderr y el objeto estructurado va en el sobre.

**La forma completa de `data.filters`, con una clave por cada filtro de esta página y el porqué de
cada una, está en ["Los filtros de `biso ls`"](../contrato-json.md#los-filtros-de-biso-ls).**

La cifra de urgencia del ejemplo puede no ser esta; el motivo está en la sección
["Urgencia"](../modelo-de-datos/urgencia.md).

**El listado nunca trae el cuerpo de la tarea**: ni descripción, ni plan, ni notas, ni criterios, ni
comentarios. Para eso está `biso get`. Los campos derivados de la sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md) sí están todos,
`blocks` incluido. `truncated` es explícito para que nadie tenga que comparar `shown` con `matched`,
y `skipped` lleva los identificadores de las tareas ilegibles que se han saltado.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Listado, incluso vacío o con tareas saltadas | 0 |
| Un valor de filtro no existe en el tablero | 3 |
| Flags incompatibles, `--limit` negativo, `--sort` inventado (`code` igual a `unknown_sort_field`), fecha mal formada | 2 |
| `--parent` a una tarea que no existe | 4 |
| `--parent` por texto con varias coincidencias, que además imprime las candidatas por stdout como cualquier otra referencia ambigua (["La búsqueda por texto"](../referencias.md#la-búsqueda-por-texto)) | 5 |
| `--mine` sin ninguna identidad configurada (["Variables de entorno"](../invocacion.md#variables-de-entorno)) | 6 |
| El almacén no responde | 8 |
| No hay tablero | 20 |

## `biso ls --help`

```
Usage: biso ls [options]

List tasks, one per line. Shows 30 by default, hides the Done ones and the
archived ones, and says on stderr what it left out. A filter value the board
does not have is an error, never an empty list, so an empty list is a fact.

Filters (repeat or comma-separate; same field is OR, different fields are AND):
  -s, --status <value>       configured status (default: all but the terminal)
      --not-status <value>   exclude a status
      --any-status           include the terminal status too
      --archived             include archived tasks
      --only-archived        only archived tasks
      --type <value>         configured type
      --priority <value>     configured priority
  -l, --label <value>        label; several labels are ANDed
      --label-or <value>     label; several are ORed
  -a, --assignee <@who>      assignee
      --mine                 assigned to you
      --unassigned           assigned to nobody
  -p, --parent <ref>         subtasks of this task
      --blocked              something unfinished blocks it
      --not-blocked          nothing unfinished blocks it; it may still be
                             waiting on an answer, so add --not-waiting
      --waiting              has an open question
      --not-waiting          has no open question
      --active               in the board's active status
      --not-active           not in the active status
      --overdue              past its due date
      --due-before <date>    due before YYYY-MM-DD
      --search <text>        free text; see `biso get --help` for the scope
      --unchecked            do not check that the labels and assignees you
                             filter by exist on the board; nothing else
                             changes

Shape:
      --sort <field>         urgency, id, ordinal, due, updated, created, title
      --reverse              flip the whole order, tie-breaks included
      --limit <n>            how many rows to print (default 30, 0 prints none)
      --all                  print every match
      --ids                  print only ids, one per line
      --count                print only how many match

Columns: id, status, type, priority, title, criteria, assignee, due. Empty
cells print a dash. The title is cut at 100 characters, always, before any
column width is computed. Columns 1 to 7 are padded to the widest value
printed; column 8 never is. Two spaces always separate columns.

Exit codes:
  0  listed, even when empty      5  --parent matched several tasks
  2  bad usage                    6  --mine with no identity configured
  3  a filter value does not exist here
  4  --parent does not exist      8  the board could not respond
                                  20 no board here

Examples:
  biso ls
  biso ls -s "In Progress" --mine
  biso ls --type bug --priority high --limit 10
  biso ls --not-blocked --not-waiting --ids
  biso ls --any-status --archived --all
```

---

