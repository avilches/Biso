# `biso board`

## Firma

```
biso board [--port <n>] [--no-open]
```

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `--port <n>` | no | entero entre 1024 y 65535 | 6420 | no | no | ninguno |
| `--no-open` | no | booleano | falso | no | no | ninguno |

La interfaz interactiva, y **el único comando del programa que abre una interfaz**. Ningún otro puede
abrirla, ni la abre nadie por su cuenta.

| Caso | Qué pasa |
|---|---|
| No hay terminal, sin `--no-open` | Error 8: `error: biso board needs a terminal; every other command works without one` |
| No hay terminal, con `--no-open` | No es un error: `--no-open` es el modo sin terminal, para un guion que arranca el servidor y solo necesita la dirección |
| El puerto está ocupado | Error 8, diciendo el puerto |
| Puerto fuera de rango | Error 2 |
| `--no-open` | Arranca y solo imprime la dirección |
| El tablero cambia mientras está abierto | La interfaz recarga. Nunca muestra una versión en caché de una tarea que otro proceso ha cambiado |

## Agrupación visual

La interfaz enseña un Kanban con las columnas fijas por `status`, en el orden configurado. Además,
admite agrupar las tareas dentro de esas columnas de dos formas independientes, elegibles con un
control de la propia interfaz y que no se combinan entre sí:

- **Por padre** (`parent`): cada tarea con alguna hija actúa de cabecera de un grupo, con su propio
  título y su propio estado, y sus hijas aparecen debajo. Es un nivel exactamente: la jerarquía no se
  recorre más allá del padre inmediato de cada tarea, sea cual sea el `type` del padre, así que no
  hace falta ninguna marca especial para que una tarea grande actúe de epic. Una tarea sin hijas y sin
  padre no entra en ningún grupo. Cerrar o archivar la tarea padre no hace desaparecer el grupo: la
  cabecera se sigue mostrando con su título y su estado mientras alguna hija siga siendo visible con
  los filtros activos, y el grupo entero deja de aparecer solo cuando ninguna hija lo es.
- **Por tipo** (`type`): una partición plana que no mira la jerarquía. Cada tarea cae en un grupo
  según su propio `type`, tantos grupos como valores en uso entre las tareas visibles.

Sin agrupar, que es lo que se ve por defecto, las tareas quedan sueltas dentro de cada columna de
estado. La agrupación es solo de esta interfaz: no existe ningún flag equivalente en `biso ls` ni
en `biso prime`, aunque las dos ya permiten filtrar por `--parent` y por `--type`
(secciones [`biso ls`](ls.md) y [`biso set`](set.md)).

## Qué escribe un arrastre

Mover una tarea con el ratón es una escritura del tablero como cualquier otra, y lo que escribe
depende de adónde se la lleve. No hay ningún camino por el que la interfaz escriba algo que la línea
de comandos no pueda escribir también:

- **Dentro de la misma columna y el mismo grupo, un arrastre reordena**, y eso es escribir la clave
  de orden manual de la tarea, la misma que escribirían `--above` y `--below`
  (["El orden manual y su clave"](../modelo-de-datos/orden-manual.md)). Las vecinas son las que la
  tarea tenga encima y debajo **en el sitio donde se suelta**, y la clave sale del hueco que dejan.
- **Cruzar de una columna a otra escribe el `status`**, que es lo que significa la columna.
- **Cruzar de un grupo a otro no reordena: edita el campo por el que se agrupa.** Con la agrupación
  por padre, el `parent`; con la agrupación por tipo, el `type`. La clave de orden manual no se toca,
  porque el grupo no es un orden sino un reparto de las mismas tareas.
- **Un arrastre que hace las dos cosas a la vez**, porque se suelta la tarea en otra columna y dentro
  de otro grupo, escribe los dos campos en la misma escritura, con la garantía de todo o nada de
  ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables).

**La clave que un arrastre escribe es global aunque la vista esté agrupada.** Cada tarea tiene una
sola, así que reordenar dentro de un grupo puede mover la tarea también en la lista sin agrupar, y
eso está aceptado a propósito: la alternativa, una clave por eje de agrupación, está descartada en
["El orden manual es una clave de texto"](../../decisiones/detalles.md#el-orden-manual-es-una-clave-de-texto).

**Un arrastre escribe la clave de la tarea que se suelta y nunca la de sus vecinas**, que es la misma
regla que hace que `--above` sobre una tarea sin clave sea un error en vez de escribirle una. De ahí
salen los tres sitios donde se puede soltar una tarea, y solo esos tres:

- **Delante de todas las de la columna**, que es `--ordinal first`, y **detrás de todas**, que es
  `--ordinal last`. Es la forma de meter la primera tarea de un tablero donde ninguna tiene clave
  todavía.
- **Entre dos tareas que tienen clave**, que es el hueco de siempre.
- **Justo en la frontera entre las que tienen clave y las que no**, que es `--ordinal last`: las
  tareas sin clave van detrás de todas las que la tienen (["La regla de orden,
  completa"](ls.md#la-regla-de-orden-completa)), así que ese sitio y el final del orden manual son el
  mismo sitio.

**Soltar una tarea entre dos que no tienen clave no escribe nada**, y la interfaz lo dice en vez de
dejar la tarea en otro sitio: ahí no hay ningún hueco que nombrar, porque el orden de esa parte de la
lista lo decide la urgencia y no una clave. Quien quiera meterla ahí tiene que darle clave antes a
alguna de las dos vecinas, exactamente lo mismo que le propone la línea de comandos.

## Salida

```
Board at http://127.0.0.1:6420. Ctrl-C to stop.
```

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "board",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "url": "http://127.0.0.1:6420",
    "port": 6420,
    "opened": true
  }
}
```

**El sobre se imprime al arrancar, no al terminar**, que es la única diferencia de este comando con
todos los demás en JSON: los otros escriben su salida cuando han acabado, y este se queda corriendo
hasta que alguien lo pare, así que el sobre sale en cuanto el servidor escucha. Es lo que necesita un
guion que lo arranque con `--no-open` y tenga que saber a qué dirección apuntar, sin necesitar él mismo
un terminal: `--no-open` es la única forma de arrancar `biso board` sin uno. `opened` dice si se ha
abierto un navegador, y es `false` con `--no-open`.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| El servidor se ha parado limpiamente | 0 |
| Puerto fuera de rango, o flags incompatibles | 2 |
| No hay terminal sin `--no-open`, o el puerto está ocupado | 8 |
| No hay tablero | 20 |

## `biso board --help`

```
Usage: biso board [options]

Open the interactive board in a browser. This is the only command that opens
an interface: every other one prints text and exits, with or without a
terminal.

Options:
      --port <n>   port to listen on, 1024 to 65535 (default 6420)
      --no-open    print the address, do not open a browser, and do not
                   require a terminal
      --json       machine-readable envelope, printed when the server starts
  -h, --help       show this help

Exit codes:
  0  stopped cleanly
  2  bad usage
  8  no terminal without --no-open, or the port is taken
  20 no board here

Examples:
  biso board
  biso board --port 7000 --no-open
```

---

