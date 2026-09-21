# `biso export`

## Firma

```
biso export [--out <file|->] [--no-archived] [cualquier filtro de biso ls, salvo --archived y --only-archived]
```

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `--out <file|->` | no | ruta o `-` | `-`, es decir stdout | no | no | |
| `--no-archived` | no | booleano | falso | no | no | |
| filtros de `ls` | no | | | | | `--sort`, `--limit`, `--all`, `--ids`, `--count`, `--archived`, `--only-archived` |

**`biso export` sin filtros exporta el tablero entero**: todos los estados, el terminal incluido, y
todas las tareas, las archivadas incluidas. **No hereda ni el límite por defecto de `biso ls` ni su
exclusión del estado terminal**, y no existe aquí ningún flag `--all`. Los filtros de `biso ls` se
aceptan para acotar a propósito, y los flags de forma de `ls` no, porque un volcado no tiene forma
que elegir. **Tampoco se aceptan `--archived` ni `--only-archived`**, porque las archivadas ya salen
por defecto: el único flag de `export` sobre el archivo es `--no-archived`.

Cuando `export` acepta `--status` como cualquier otro filtro de `biso ls`, se aplica igual que allí: un
`--status` explícito filtra por ese valor tal cual, terminal incluido (["`biso ls`"](ls.md#comportamiento-caso-a-caso)), así que
`biso export --status Done --no-archived --out done.ndjson` exporta exactamente las tareas `Done` vivas. La
diferencia con `ls` es solo la base cuando no se pasa ningún `--status`: `export` parte de todos los
estados, `ls` parte de todos menos el terminal.

## La garantía de simetría

**Esta garantía asume un tablero destino con vocabulario compatible**: el mismo `task_prefix`, los
mismos estados, tipos y prioridades que el tablero de origen. Si no lo es, la
importación falla con los errores que ya define ["El modo lote"](new.md#el-modo-lote) de `biso new`
(un `id` sin el prefijo correcto, un valor sin vocabulario), no es
un fallo de esta garantía. Reconstruir también el vocabulario, para un tablero destino que no lo
declara de antemano, es lo que hace `snapshot` con `init --from`, más abajo.

La salida es NDJSON, una tarea por línea, con las claves que acepta `biso new --from`,
en la forma de objeto que esa sección define para los criterios, los comentarios y la pregunta
abierta, e incluyendo `id`, `createdAt`, `updatedAt`, `archived`, `question`
y las claves estables de cada criterio **y de cada comentario** (["Los criterios y sus claves estables"](../modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables), ["Los comentarios"](../modelo-de-datos/comentarios.md#los-comentarios)).

**El formato de entrada de `biso new --from` es un superconjunto del de salida, y la diferencia son
tres claves.** `new --from` acepta además `definitionOfDone`, que convierte en criterios de aceptación,
y `documentation` y `modifiedFiles`, que funde en `references` (["`biso new`"](new.md)). `export` no
escribe ninguna de las tres nunca, porque no son campos del modelo. La garantía de la ida y vuelta no se resiente: lo que
`export` escribe, `new --from` lo lee campo a campo, y esas claves solo aparecen en lotes que vengan
de fuera.

**Los únicos campos que no salen son los derivados de la sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md).** `question` sale en `export` y
entra de vuelta con `new --from`, con sus tres partes completas.

**La clave de orden manual sale tal cual está guardada.** `ordinal` es una cadena, no un número
(["El orden manual y su clave"](../modelo-de-datos/orden-manual.md)), y `export` la escribe sin
recalcular nada: `biso new --from` la vuelve a guardar igual, así que las tareas del tablero de
destino quedan en el mismo orden manual que las del de origen. Es justamente lo que un decimal no
podía prometer sin fijar cuántos dígitos se escriben.

**Un escalar opcional sin valor sale como `null`; una lista o un mapa sin elementos sale como `[]` o
`{}`, nunca como `null`.** Es la misma regla de coerción de ["El modo lote"](new.md#el-modo-lote) de
`biso new`, en el sentido contrario, la que hace que reimportar la salida reproduzca la tarea exacta.

`export` solo lleva las tareas: reconstruir un tablero entero, con su vocabulario y no solo con sus
datos, es lo que hace [`biso snapshot`](snapshot.md), cuyo `snapshot.ndjson` tiene exactamente esta misma forma
y se lee de vuelta con el `--from` de [`biso init`](init.md), no con el de `biso new`.

La garantía que la suite de pruebas comprueba:

```bash
biso snapshot
# escribe snapshot.ndjson y board.json en ~/.biso/boards/my-project-3f9a2b1c, el propio
# directorio del tablero de origen (biso where lo muestra en su fila "path")
HOME=/tmp/otra-maquina biso -C /tmp init --at /tmp/tablero-nuevo --from ~/.biso/boards/my-project-3f9a2b1c
# los dos tableros son identicos en todos los campos no derivados, incluidos
# los identificadores, las fechas, las claves de los criterios y sus marcas,
# las claves de los comentarios, y en toda su configuracion: estados, tipos
# y task_prefix. La unica salvedad son los dos contadores de
# claves, que no son una clave del formato y se deducen al importar: ver
# "El contador de claves no es una clave del formato", mas abajo
```

**El `init` se ejecuta con las raíces de otra máquina**, que es lo que de verdad pasa cuando una
instantánea viaja: el `id` que el marcador trae está libre allí. Hacerlo con las raíces de la máquina
de origen, donde ese `id` ya vive, es el error de identidad duplicada de la tabla de casos de
["`biso init`"](init.md), y no depende de que `--at` apunte fuera de las raíces, porque lo que
quedaría duplicado es la identidad y no la carpeta. El `-C /tmp` sigue estando por otra razón, la de
la versión anterior de esta prueba: se llama desde fuera del proyecto de origen para que el puntero
que ese proyecto ya tiene no entre en juego.

La comparación de la prueba no es la de los dos volcados de `biso export`, que los escribe la misma
función en los dos lados y por tanto no puede ver un campo que esa función deje de escribir: es la de
las dos bases de datos, tabla por tabla y columna por columna. El formato en sí lo fija aparte un
fichero de referencia escrito a mano con la línea exportada de una tarea con todos sus campos
puestos, que no cambia cuando cambia el código.

`biso init --from` lee el vocabulario del propio `board.json` de la instantánea, así que el tablero
de destino no necesita declarar nada a mano: nace con el mismo `task_prefix`, los mismos estados y
los mismos tipos que el de origen, y por eso la importación de su `snapshot.ndjson` nunca falla por
vocabulario distinto. Comparar esto con la vía manual de [`biso new --from`](new.md): esa sigue
existiendo para importar un NDJSON suelto en un tablero cuyo vocabulario ya se ha declarado por
separado, pero ya no es la única manera de reconstruir un tablero entero.

## El rechazo de `--json`

`--json` es aquí la parte inválida de la llamada, así que el rechazo es texto plano por stderr y no
el sobre de error, como en cualquier otro comando que no acepta el flag en absoluto
(["Los errores en JSON"](../contrato-json.md#los-errores-en-json)):

```
error: --json does not apply to export, whose output is already NDJSON
```

## El contador de claves no es una clave del formato

`export` no escribe los dos contadores de una tarea, los de las claves de sus criterios y de sus
comentarios, porque no son campos del modelo que nadie pueda escribir: se deducen al importar, por
encima de la clave mayor que traiga la línea (["`biso new`"](new.md#el-modo-lote)). En la práctica
eso los reproduce siempre que el contador siga valiendo uno más que la clave mayor que la tarea
conserva, que es lo normal.

**La excepción es una tarea a la que se le ha quitado el criterio de clave mayor**, o el comentario
de clave mayor. Ahí el contador se queda por encima de una clave que ya no existe, la importación lo
baja a la mayor que quede más uno, y el siguiente criterio de esa tarea recibe una clave que ya se
había usado antes. Quitarlos todos es el caso extremo de eso mismo, con el contador de vuelta a 1, y
no un caso aparte: basta con quitar el último para que la ida y vuelta pierda el contador. Es el
único dato del tablero que no conserva, y se deja así a propósito: guardarlo obligaría a añadir al
formato dos claves que ningún comando escribe y que solo servirían para ese caso.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Exportado, aunque sean cero tareas | 0 |
| Alguna tarea se ha saltado por ilegible | 6 |
| Flags de forma de `ls`, `--archived`, `--only-archived`, `--json`, o incompatibles | 2 |
| Un valor de filtro no existe en el tablero | 3 |
| No se puede escribir el fichero de salida | 8 |
| No hay tablero | 20 |

## `biso export --help`

```
Usage: biso export [options]

Write the board as NDJSON, one task per line, in exactly the shape that
`biso new --from` reads back. Round-tripping every non-derived field is a
tested guarantee: ids, dates, criterion and comment keys, and checkmarks
included.

With no filters it exports everything, the finished and the archived included.
It never inherits the default limit or the default status filter of `biso ls`.
If a task cannot be read, the rest is still written and the exit code is 6,
not 0: this is the one command whose purpose is to lose nothing.

Options:
      --out <file|->   where to write (default: stdout)
      --no-archived    leave the archived tasks out
  -h, --help           show this help

Every filter of `biso ls` works here except --archived and --only-archived,
which do not apply because archived tasks are already included by default.
Its shaping flags (--sort, --limit, --all, --ids, --count) do not apply either.
--json is rejected with code 2: this output is already one JSON object per
line, while --json means the single envelope every other command prints.

Derived fields are never written: urgency, acDone, acTotal, commentCount,
blocks, blocked, waiting, leaseExpired.

Exit codes:
  0  exported       3  a filter value does not exist here
  2  bad usage      6  some task was skipped, unreadable
  8  cannot write there                20 no board here

Examples:
  biso export --out backup.ndjson
  biso export --status Done --no-archived --out done.ndjson
```

---

