# `biso export`

## Firma

```
biso export [-o <file|->] [--no-archived] [cualquier filtro de biso ls, salvo --archived y --only-archived]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--out <file|->` | `-o` | no | ruta o `-` | `-`, es decir stdout | no | no | |
| `--no-archived` | | no | booleano | falso | no | no | |
| filtros de `ls` | | no | | | | | `--sort`, `--limit`, `--all`, `--ids`, `--count`, `--archived`, `--only-archived` |

**`biso export` sin filtros exporta el tablero entero**: todos los estados, el terminal incluido, y
todas las tareas, las archivadas incluidas. **No hereda ni el límite por defecto de `biso ls` ni su
exclusión del estado terminal**, y no existe aquí ninguna bandera `--all`. Los filtros de `biso ls` se
aceptan para acotar a propósito, y las banderas de forma de `ls` no, porque un volcado no tiene forma
que elegir. **Tampoco se aceptan `--archived` ni `--only-archived`**, porque las archivadas ya salen
por defecto: la única bandera de `export` sobre el archivo es `--no-archived`.

## La garantía de simetría

La salida es NDJSON, una tarea por línea, con **exactamente** las claves que acepta `biso new --from`,
en la forma de objeto que esa sección define para los criterios, la definición de hecho, los
comentarios y la pregunta abierta, e incluyendo `id`, `createdAt`, `updatedAt`, `archived`, `question`
y las claves estables de cada criterio **y de cada comentario** (["Los criterios y sus claves estables"](../modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables), ["Los comentarios"](../modelo-de-datos/comentarios.md#los-comentarios)).

**Los únicos campos que no salen son los derivados de la sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md).** `question` sale en `export` y
entra de vuelta con `new --from`, con sus tres partes completas.

`export` solo lleva las tareas: reconstruir un tablero entero, con su vocabulario y no solo con sus
datos, es lo que hace [`biso snapshot`](snapshot.md), cuyo `snapshot.ndjson` tiene exactamente esta misma forma
y se lee de vuelta con el `--from` de [`biso init`](init.md), no con el de `biso new`.

La garantía que la suite de pruebas comprueba:

```bash
biso snapshot
# escribe snapshot.ndjson y board.json en ~/.biso/boards/my-project-3f9a2b1c, el propio
# directorio del tablero de origen (biso where lo muestra en su fila "path")
biso -C /tmp init --at /tmp/tablero-nuevo --from ~/.biso/boards/my-project-3f9a2b1c
# los dos tableros son identicos en todos los campos no derivados, incluidos
# los identificadores, las fechas, las claves de los criterios y sus marcas,
# las claves de los comentarios, y en toda su configuracion: estados, tipos,
# extensiones y task_prefix
```

El `init` se ejecuta con `-C /tmp`, fuera del proyecto de origen, exactamente como en la versión
anterior de esta prueba: así su puntero de proyecto no choca con el que el proyecto de origen ya
tiene.

`biso init --from` lee el vocabulario del propio `board.json` de la instantánea, así que el tablero
de destino no necesita declarar nada a mano: nace con el mismo `task_prefix`, los mismos estados y
los mismos tipos que el de origen, y por eso la importación de su `snapshot.ndjson` nunca falla por
vocabulario distinto. Comparar esto con la vía manual de [`biso new --from`](new.md): esa sigue
existiendo para importar un NDJSON suelto en un tablero cuyo vocabulario ya se ha declarado por
separado, pero ya no es la única manera de reconstruir un tablero entero.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Exportado, aunque sean cero tareas | 0 |
| Alguna tarea se ha saltado por ilegible | 6 |
| Banderas de forma de `ls`, `--archived`, `--only-archived`, `--json`, o incompatibles | 2 |
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
  -o, --out <file|->   where to write (default: stdout)
      --no-archived    leave the archived tasks out
  -h, --help           show this help

Every filter of `biso ls` works here except --archived and --only-archived,
which do not apply because archived tasks are already included by default.
Its shaping flags (--sort, --limit, --all, --ids, --count) do not apply either.
--json is rejected with code 2: this output is already one JSON object per
line, while --json means the single envelope every other command prints.

Derived fields are never written: urgency, acDone, acTotal, dodDone, dodTotal,
commentCount, blocks, blocked, waiting, leaseExpired.

Exit codes:
  0  exported       3  a filter value does not exist here
  2  bad usage      6  some task was skipped, unreadable
  8  cannot write there                20 no board here

Examples:
  biso export -o backup.ndjson
  biso export -s Done --no-archived -o done.ndjson
```

---

