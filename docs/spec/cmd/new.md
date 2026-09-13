# `biso new`

## Firma

```
biso new [<title>] [--start] [--from <file|->] [cualquier bandera de campo de las familias de banderas]
```

## Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<title>` | | sí, salvo con `--from` | texto | | no | no | `--from` |
| `--start` | | no | booleano | falso | no | no | `-s`, `--from` |
| `--from <file|->` | | no | ruta o `-` | | no | no | `<title>` y todas las de campo |

Todas las banderas de campo de la sección ["Las familias de banderas"](../familias-de-banderas.md) valen aquí. En una tarea nueva no hay nada que sustituir
ni que quitar: `--replace-*` se acepta y deja la lista igual que `--add-*`, porque no hay nada previo
que sustituir, y `--rm-*` se acepta pero no tiene ningún elemento sobre el que actuar. `--clear-*` no
hace nada y avisa. Las que se usan de verdad al crear son
`-d/--append-desc`, `--add-ac`, `--add-dod`, `--type`, `--priority`, `-l/--add-labels`,
`-a/--add-assignees`, `--add-refs`, `--add-docs`, `--add-deps`, `-p/--parent`,
`--due`, `--ordinal`, `--reporter`, `--ext`, `--append-plan`, `--append-note`,
`--append-summary` y `--comment`.

- **`--start`** crea la tarea directamente en el estado activo, asignada a `me` y con el arrendamiento
  tomado a favor de quien llama (`leaseExpiresAt` y `leaseHolder`, sección ["El modelo de datos de una tarea"](../modelo-de-datos.md)), exactamente como lo haría
  `biso start` sobre ella. Es el atajo de esas dos llamadas, así que la equivalencia tiene que ser real:
  si `--start` dejara la tarea activa y asignada sin arrendamiento, `biso new "X" --start` y
  `biso new "X"` seguido de `biso start` darían dos tareas distintas. Es, junto con `biso start`, la
  única vía que fija `leaseHolder` fuera de la importación.
- **`--comment` funciona al crear**, igual que en cualquier otro comando de escritura.
- **`--append-plan`, `--append-note` y `--append-summary` no están restringidos por el estado.** Se
  pueden escribir al crear, en cualquier estado.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Título vacío o solo espacios | Error 2: `error: title cannot be empty` |
| Título muy largo | Se acepta entero, sin recortar |
| Título repetido | Se acepta sin aviso. Dos tareas pueden llamarse igual, para eso está el identificador |
| Valor fuera de un vocabulario cerrado | Error 3, con la lista de válidos |
| `--add-deps` a una tarea inexistente | Error 4. **Las dependencias se validan al escribirlas** |
| `--add-deps` a la propia tarea, o que cerraría un ciclo | Error 2 |
| `--parent` inexistente, o que cerraría un ciclo | Error 4 y error 2 respectivamente |
| `--ext` con una clave no declarada | Error 3 |
| `--due` con formato incorrecto | Error 2, señalando `YYYY-MM-DD` |
| `--due` en el pasado | Se acepta, con aviso |
| `-d @fichero` que no existe | Error 4 |
| `--start` sin ninguna identidad configurada (["Variables de entorno"](../invocacion.md#variables-de-entorno)) y sin `-a` | La tarea se crea en el estado activo y sin asignar, con `note: no identity configured, task left unassigned`, y **sin arrendamiento**: no hay ninguna identidad a la que atribuírselo, y una tarea sin asignar no puede tenerlo (sección ["El modelo de datos de una tarea"](../modelo-de-datos.md)). Es el mismo caso que la fila equivalente de `biso start` (["`biso start`"](verbos-del-ciclo.md#biso-start)) |
| `--start` con `-a @sara` y una identidad configurada distinta | La tarea queda asignada a `@sara` y el arrendamiento es de quien llama, igual que en `biso start`: quien lo toma es quien escribe, no quien figura en `assignees` |
| Todo bien | Se crea la tarea, código 0 |

## Salida

Por defecto, **una línea por tarea creada, con el identificador y nada más**:

```
MYP-101
```

`biso new` es el único comando de escritura cuya salida por defecto es distinta de la línea de estado
de la sección ["`biso set`"](set.md), y así está dicho en el mensaje de arranque.

Con `--print`, después de la línea del identificador viene la ficha completa en el formato de
`biso get`. Con `--quiet`, solo el identificador y ninguna nota.

**`biso new` no anuncia la clave de un `--add-ac` o `--add-dod` creado al mismo tiempo que la
tarea, a diferencia de la línea de estado de `biso set` (["`biso set`"](set.md#salida)).** Una tarea nace sin ningún
criterio, así que su contador de claves empieza siempre en 1: el primer `--add-ac` de la llamada es
la `#1`, el segundo la `#2`, y así en el mismo orden en que se escribieron las banderas
(["Los criterios y sus claves estables"](../modelo-de-datos.md#los-criterios-y-sus-claves-estables)). Quien llama ya lo sabe sin preguntar, así que
imprimirlo sería el eco que el principio 4 prohíbe (["Los principios"](../principios.md)), y no la clase de dato que ese principio
manda enseñar.

## El modo lote

```
biso new --from tareas.ndjson
biso new --from -
biso new --from tareas.ndjson --dry-run
```

La entrada es **NDJSON**: un objeto JSON por línea. Las líneas vacías y las que empiezan por `#` se
ignoran. Las claves son las del modelo de datos de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md), en `camelCase`.

Ejemplo de una línea, con todos los tipos compuestos:

```json
{"id":"MYP-101","title":"Normalize CRLF in the diff","type":"bug","priority":"high","status":"Done","description":"...","labels":["parser"],"references":["docs/bugs/BUG-02.md"],"dependencies":["MYP-90"],"ext":{"trello.card":"5f2a8c1e"},"acceptanceCriteria":[{"key":1,"text":"The diff ignores CRLF","checked":true},{"key":3,"text":"There is a test","checked":false}],"definitionOfDone":[{"key":1,"text":"Reviewed","checked":true}],"comments":[{"key":1,"author":"@avilches","createdAt":"2026-08-14T10:22:00Z","body":"Reported from Windows"}],"question":{"author":"@avilches","askedAt":"2026-08-16T09:00:00Z","body":"Is it a CRLF, or also a lone CR?"},"createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
```

Las reglas del lote, todas obligatorias:

- **`acceptanceCriteria` y `definitionOfDone` aceptan dos formas.** Una cadena, que crea un elemento
  sin marcar con la siguiente clave libre, o un objeto con `key`, `text` y `checked`. Las dos formas
  se pueden mezclar dentro de la misma lista. Una `key` repetida dentro de la misma tarea es un fallo
  de validación.
- **El contador de claves de cada lista se sitúa por encima de la clave mayor importada**, de modo que
  un criterio añadido después nunca choca con uno importado. El contador no es una clave del formato:
  se deduce.
- **`comments` es una lista de objetos** con `author`, `createdAt`, `body` y, opcionalmente, `key`.
  `createdAt` es opcional y, si falta, se pone el instante de la importación. `key` sigue la misma
  regla que la de `acceptanceCriteria` y `definitionOfDone`
  (["Los criterios y sus claves estables"](../modelo-de-datos.md#los-criterios-y-sus-claves-estables)): si falta, se asigna con la siguiente clave libre del
  contador de comentarios de esa tarea, y una `key` repetida dentro de los comentarios de la misma
  tarea es un fallo de validación. Es lo que hace cierta la simetría de `biso export` para la clave de
  un comentario (["`biso export`"](export.md)).
- **`question` se acepta como objeto** con `author`, `askedAt` y `body` (["La pregunta abierta"](../modelo-de-datos.md#la-pregunta-abierta)) en el lote de `--from`.
  `askedAt` es opcional y, si falta, se pone el instante de la importación, igual que `createdAt` en
  `comments`. Ausente la clave, la tarea se importa sin pregunta abierta.
- **`id`, `createdAt` y `updatedAt` se aceptan aquí y solo aquí.** Un `id` ya ocupado es un fallo de
  validación; un `id` libre se reserva y el tablero no lo volverá a asignar.
- **`leaseExpiresAt` y `leaseHolder` se aceptan aquí con el valor que traiga el fichero**, que es lo
  que hace cierta la garantía de simetría de ["`biso export`"](export.md) para ellos dos. La invariante de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md) se
  comprueba en la validación, en sus dos mitades, y cada una es un fallo que nombra la línea y el campo.
  Una línea que traiga cualquiera de los dos sobre una tarea que no esté a la vez en el estado activo y
  asignada a alguien es un fallo de validación. Y una línea que traiga uno de los dos y no el otro
  también lo es, aunque la tarea esté activa y asignada: los campos van juntos, porque
  `leaseExpired` se calcula comparando `leaseExpiresAt` con el reloj de quien lee y con ese campo vacío
  no habría nada que comparar. Los dos fallos se ven así:
  ```
  line 14: leaseHolder on a task that is not both active and assigned
  line 31: leaseHolder given without leaseExpiresAt; the two go together
  ```
- **Un `id` explícito tiene que llevar el `task_prefix` del tablero de destino.** Si no lo lleva, es
  un fallo de validación, igual que un `id` ya ocupado: es la misma protección que hace inmutable a
  `task_prefix` en la sección ["`biso config`"](config.md), cerrando la tercera vía hacia el mismo tablero de identificadores
  mixtos que esa inmutabilidad ya evita en las otras dos (cambiar `--prefix` a mano, o renombrar el
  tablero). No es una restricción nueva sobre la simetría: exportar un tablero y restaurarlo con
  `biso snapshot` y `biso init --from` (["`biso export`"](export.md), ["`biso snapshot`"](snapshot.md)) trae también su `task_prefix`, así que los `id`
  de su `snapshot.ndjson` siempre lo llevan puesto.
- **`archived` se acepta como booleano.** Por defecto, si la clave no aparece, la tarea se crea sin
  archivar. Ningún otro comando tiene una bandera de campo para él: fuera de la importación,
  archivar se hace con `biso archive`.
- **Una clave desconocida es un fallo de validación, no se ignora.** Ni la línea ni el lote se
  escriben, y el mensaje dice la línea y la clave.
- **Los campos derivados de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md) no se aceptan.** En la entrada son claves desconocidas y
  por tanto un fallo de validación.
- **Se valida el fichero entero antes de escribir nada**, y se aplica la garantía de todo o nada de
  la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables).
- Un lote no admite `--start` ni ninguna bandera de campo: todo va en el fichero.

Salida del lote, una línea por tarea, en el orden del fichero:

```
MYP-101
MYP-102
MYP-103
```

Salida de `--dry-run` cuando todo está bien, por stderr y con código 0:

```
242 tasks would be created, nothing was written (--dry-run)
```

Y cuando no, por stderr y con código 9, **con todos los fallos, no solo el primero**:

```
error: 4 of 242 lines are invalid, nothing was written
  line 12: id "OTHER-5" does not match this board's task prefix "MYP"
  line 47: unknown status: "Pendiente" (valid: To Do, In Progress, Done)
  line 88: unknown key: "trelloCard"
  line 201: title cannot be empty
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea o lote creado | 0 |
| `--dry-run` que habría funcionado | 0 |
| Falta el título, banderas incompatibles, fecha mal formada, ciclo de dependencias o de padres | 2 |
| Valor fuera de un vocabulario, clave de extensión no declarada, entrada no interpretable | 3 |
| `--add-deps` o `--parent` a una tarea que no existe, o fichero de `@` que no existe | 4 |
| `--add-deps` o `--parent` por texto con varias coincidencias | 5 |
| Cualquier fallo de validación en un lote, o un `--dry-run` que no pasa | 9 |
| El almacén falla, o no se obtiene el acceso exclusivo | 7 |
| No hay tablero | 8 |

## `biso new --help`

```
Usage: biso new <title> [options]
       biso new --from <file|-> [options]

Create a task and print its id. Every field flag of `biso set` works here.

Arguments:
  title                      task title (required unless --from is given)

Most used:
  -d, --append-desc <text>    description; repeat to append paragraphs
      --add-ac <text>         add an acceptance criterion; repeatable
      --add-dod <text>        add a definition-of-done item; repeatable
      --type <value>          configured type
      --priority <value>      configured priority
  -s, --status <value>        configured status (default: the initial one)
  -l, --add-labels <value>    add a label; repeatable or comma-separated
  -a, --add-assignees <@who>  add an assignee; repeatable or comma-separated
      --add-deps <ref>        add a dependency; validated, repeatable
      --due <YYYY-MM-DD>      due date
      --comment <text>        add a discussion comment; repeatable
      --append-plan <text>    implementation plan
      --start                 create it already in the active status, assigned
                              to you, with the lease claimed for you

Every other field flag of `biso set --help` is accepted too.

Batch:
      --from <file|->        NDJSON, one task object per line. The only place
                             where id, createdAt, updatedAt, criterion keys,
                             comment timestamps and question timestamps can be
                             given. Validated whole before anything is written.

Any text option also takes @file to read a file, or - to read stdin.

Exit codes:
  0  created            4  a referenced task or file does not exist
  2  bad usage          5  a text reference matched several tasks
  3  unknown value      7  the board could not be written
  9  batch or --dry-run validation failed, nothing was written
                        8  no board here

Examples:
  biso new "Normalize CRLF in the diff" --type bug --priority high
  biso new "Add OAuth" --add-ac "Login succeeds" --add-ac "Token refreshes"
  biso new "Rewrite the installer" -d @docs/installer.md --start
  biso new --from tasks.ndjson --dry-run
```

---

