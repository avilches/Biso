# `biso where`

## Firma

```
biso where [--json]
```

Sin parámetros propios.

## Comportamiento

Dice el identificador del tablero, su nombre, la ruta de su directorio, y cuál de las dos vías de la
sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) lo ha elegido. Los tres son datos distintos, y merece la pena verlos juntos porque cada uno
cambia por su cuenta: `biso config set project_name` cambia el nombre y no toca la ruta (["`biso config`"](config.md)), mover el
directorio a mano cambia la ruta y no toca el nombre, y el identificador no cambia jamás. Es el comando al
que remite el error de código 8, y el que hace visible una resolución que de otro modo sería invisible.

| Caso | Qué pasa |
|---|---|
| Hay tablero | Lo imprime con su identificador, su nombre, su ruta y la vía que lo eligió, código 0 |
| El directorio de trabajo es el propio directorio del tablero | Lo imprime igual, con la primera vía de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) en `source` y `path` apuntando al directorio de trabajo, código 0 |
| No hay tablero configurado | Imprime lo que ha buscado y dónde, código 8, `code` `no_board` |
| El puntero nombra un tablero que no está en esta máquina | Imprime que hay un puntero y qué identificador nombra (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), código 8, `code` `pointer_unresolved` |
| Hay más de un candidato | Imprime el elegido y los descartados, con el motivo, código 0 |
| El mismo `id` aparece en dos raíces | Imprime los dos directorios y no elige ninguno, código 11, `code` `ambiguous_board_id` (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)) |
| La base de datos del tablero no se puede leer | El mensaje de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), código 10, `code` `database_unreadable`. `where` no lo esquiva: para decir qué tablero está en uso hay que abrirlo |

## Salida

```
id       3f9a2b1c
board    Kex
path     /Users/avilches/.biso/boards/kex-3f9a2b1c
source   project pointer at /Users/avilches/Hub/Projects/Kex
me       @claude
tasks    248 not archived, 31 archived, highest id ever assigned TASK-290
```

**La fila `path` es siempre la ruta ya resuelta del directorio del tablero, nunca el texto literal que
lleve el puntero.** Cuando el puntero trae una `path` relativa (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), `where` la enseña resuelta
contra el directorio del puntero, que es el que la fila `source` nombra justo debajo, así que las dos
filas juntas dicen a la vez dónde está el tablero y de dónde salió esa respuesta. Enseñar el texto
literal sería enseñar la pregunta en vez de la respuesta: `where` existe para contestar dónde está el
tablero de verdad, y es el comando al que remiten los errores de código 8, donde una ruta que hay que
resolver a mano no sirve de nada. La clave `path` del sobre JSON lleva esa misma ruta resuelta, porque
es el mismo dato en la otra forma.

**Resuelta quiere decir también expandida**, así que la tilde del directorio personal no aparece nunca en
esta fila ni en la clave del JSON, aunque sí aparezca en el valor por defecto de `boards_root`
(sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)) y en los ejemplos de esta especificación, donde se escribe para que se lean. `~` es una
abreviatura que expande el intérprete de órdenes, no una ruta, y una salida que la llevara obligaría a
quien la consume a expandirla por su cuenta.

La fila `source` nombra el directorio del que salió el puntero, y no solo la vía, porque con la
búsqueda de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) ese directorio puede ser cualquier ancestro del de trabajo: enseñarlo es lo
que hace visible de un vistazo el caso de haber heredado el puntero de un proyecto que contiene a este.
Con la otra vía no hay ningún directorio del que salir, porque el tablero es el directorio de trabajo,
y la fila dice `the working directory is this board`.

Y cuando no hay ninguno configurado, por stderr y con código 8:

```
error: no board here, and none configured for this project
searched  this directory: not a board
          pointer:        not found between this directory and /Users/avilches,
                          which is where the search stops (3.2)
hint: `biso init` creates one
```

Y cuando el puntero de este proyecto nombra un tablero que esta máquina no tiene, el mismo mensaje
que da cualquier otro comando en este caso (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), por stderr y con código 8:

```
error: this project's pointer names board 3f9a2b1c, which is not on this machine
hint: `biso init` creates it here, adopting id 3f9a2b1c
```

**La fila de recuentos dice "not archived" y no "active", y la clave JSON se llama `notArchived` por lo
mismo.** La tabla de vocabulario de este documento reserva "active" para el papel del estado, el que
`biso start` usa, y una tarea sin archivar puede estar en cualquiera de los estados, incluido el
terminal. Llamarla activa haría que la misma palabra significara dos cosas en el mismo documento, y en
la fila donde más confunde, porque justo al lado hay un recuento de estados.

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "where",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "id": "3f9a2b1c",
    "board": "Kex",
    "path": "/Users/avilches/.biso/boards/kex-3f9a2b1c",
    "source": "project pointer at /Users/avilches/Hub/Projects/Kex",
    "me": "@claude",
    "counts": { "notArchived": 248, "archived": 31, "highestIdEverAssigned": "TASK-290" }
  }
}
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Tablero encontrado | 0 |
| No hay tablero | 8 |
| Su base de datos no se puede leer | 10 |
| El mismo `id` está en dos raíces | 11 |

## `biso where --help`

```
Usage: biso where [options]

Say which board is in use and which rule picked it. Run it when a command
answers "no board here" and you expected one.

Options:
      --json     machine-readable envelope
  -h, --help     show this help

Exit codes:
  0  a board is in use
  8  no board here
  10 its database could not be read
  11 the same board id is in two places

Examples:
  biso where
  biso -C ~/work/kex where
```

---

