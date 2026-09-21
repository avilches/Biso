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
admite agrupar las tareas dentro de esas columnas por uno de estos ejes, elegible con un control de
la propia interfaz, sin que ninguno se combine con otro:

- **Por padre** (`parent`): cada tarea con alguna hija actúa de cabecera de un grupo, con su propio
  título y su propio estado, y sus hijas aparecen debajo. Es un nivel exactamente: la jerarquía no se
  recorre más allá del padre inmediato de cada tarea, sea cual sea el `type` del padre, así que no
  hace falta ninguna marca especial para que una tarea grande actúe de epic. Una tarea sin hijas y sin
  padre no entra en ningún grupo. Cerrar o archivar la tarea padre no hace desaparecer el grupo: la
  cabecera se sigue mostrando con su título y su estado mientras alguna hija siga siendo visible con
  los filtros activos, y el grupo entero deja de aparecer solo cuando ninguna hija lo es.
- **Por tipo** (`type`): una partición plana que no mira la jerarquía. Cada tarea cae en un grupo
  según su propio `type`, tantos grupos como valores en uso entre las tareas visibles.
- **Por la clave de una etiqueta con ámbito**
  (["Las etiquetas con ámbito"](../valores-de-entrada.md#las-etiquetas-con-ámbito)): se elige una clave
  de las que el tablero tiene (["Consultar por la clave de una etiqueta con
  ámbito"](../vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito)) y cada valor de esa
  clave es un grupo, más un grupo para las tareas que no llevan ninguna etiqueta suya. Con una clave
  escrita `::` cada tarea cae en un grupo y en uno solo, porque esa es justamente la garantía que da
  el separador; con una clave escrita `:` **una tarea aparece en tantos grupos como valores lleve**,
  que es lo que significa admitir varios, y por eso el eje no promete ser una partición.

Sin agrupar, que es lo que se ve por defecto, las tareas quedan sueltas dentro de cada columna de
estado. La agrupación es solo de esta interfaz: no existe ningún flag equivalente en `biso ls` ni
en `biso prime`, aunque las dos ya permiten filtrar por `--parent`, por `--type` y por la clave de
una etiqueta (secciones [`biso ls`](ls.md) y [`biso set`](set.md)).

## Qué escribe un arrastre

Mover una tarea con el ratón es una escritura del tablero como cualquier otra, y lo que escribe
depende de adónde se la lleve. No hay ningún camino por el que la interfaz escriba algo que la línea
de comandos no pueda escribir también:

| Movimiento | Qué escribe |
|---|---|
| A otra columna | `status`, al de esa columna |
| Al grupo de otro padre, agrupando por padre | `parent` |
| A otro grupo de tipo, agrupando por tipo | `type` |
| A otro grupo de valor, agrupando por la clave de una etiqueta | esa etiqueta: quita la del grupo del que sale y añade la del grupo al que entra, con la misma llamada que haría `--rm-labels k:<origen> --add-labels k:<destino>` (["Escribir una etiqueta con ámbito"](../familias-de-flags.md#escribir-una-etiqueta-con-ámbito)) |
| Al grupo de las que no llevan la clave | solo quita la etiqueta del grupo de origen |
| Desde el grupo de las que no llevan la clave | solo añade la del grupo de destino |
| Dentro del mismo grupo | `ordinal`, la clave de orden manual, que es lo único que reordena (["El orden manual y su clave"](../modelo-de-datos/orden-manual.md)) |

**Un arrastre entre grupos nunca reordena, y un arrastre dentro de un grupo nunca cambia otro campo.**
La regla vale igual para el eje de las etiquetas que para los otros: arrastrar cambia el campo por el
que se agrupa, y solo moverse dentro del grupo toca el orden manual. Con una clave escrita `:`, en la
que una tarea puede estar en varios grupos a la vez, lo que se quita es el valor del grupo del que
sale y no todos los de la clave, que es lo único que conserva el resto de la información de la tarea.

**Un arrastre que hace las dos cosas a la vez**, porque se suelta la tarea en otra columna y dentro
de otro grupo, escribe los dos campos en la misma escritura, con la garantía de todo o nada de
["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables).

### Dónde escribe la clave de orden manual

**Reordenar dentro de la misma columna y el mismo grupo es escribir la clave de orden manual** de la
tarea, la misma que escribirían `--above` y `--below`
(["El orden manual y su clave"](../modelo-de-datos/orden-manual.md)). Las vecinas son siempre las
que la tarea tenga encima y debajo **en el sitio donde se suelta**, es decir dentro de la columna y
del grupo donde cae, y la clave sale del hueco que dejan esas dos. Cualquiera de las dos puede
faltar, y entonces el hueco llega hasta ese extremo, exactamente como cuando falta una vecina en la
línea de comandos.

**La clave que un arrastre escribe es global aunque la vista esté agrupada.** Cada tarea tiene una
sola, así que reordenar dentro de un grupo puede mover la tarea también en la lista sin agrupar, y
eso está aceptado a propósito: la alternativa, una clave por eje de agrupación, está descartada en
["El orden manual es una clave de texto"](../../decisiones/detalles.md#el-orden-manual-es-una-clave-de-texto).

**Un arrastre escribe la clave de la tarea que se suelta y nunca la de sus vecinas**, que es la misma
regla que hace que `--above` sobre una tarea sin clave sea un error en vez de escribirle una. De ahí
salen los tres sitios donde se puede soltar una tarea, y solo esos tres:

- **Entre dos tareas que tienen clave**, que es el hueco de siempre.
- **Delante de todas las de la columna, o detrás de todas.** También es un hueco, y el mismo de
  siempre: la vecina que falta es la que falta, así que soltar arriba del todo da el hueco que va
  desde el principio del orden hasta la primera tarea **de esa columna**, y soltar abajo del todo el
  que va desde la última **de esa columna** hasta el final. No son `--ordinal first` ni
  `--ordinal last`, que miran el tablero entero (["El hueco de cada
  colocación"](../modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación)): en una vista
  agrupada, el extremo de una columna y el extremo del tablero no tienen por qué coincidir. Con el
  tablero sin agrupar y sin filtrar son el mismo sitio, y entonces sí coinciden, que es el caso de
  meter la primera tarea de un tablero donde ninguna tiene clave todavía.
- **Justo en la frontera entre las que tienen clave y las que no**, que es el hueco que deja la
  última tarea con clave de esa columna sin ninguna por debajo: las tareas sin clave van detrás de
  todas las que la tienen (["La regla de orden,
  completa"](ls.md#la-regla-de-orden-completa)), así que ese sitio y el final del orden manual de la
  columna son el mismo sitio.

**Soltar una tarea entre dos que no tienen clave no escribe nada**, y la interfaz lo dice en vez de
dejar la tarea en otro sitio: ahí no hay ningún hueco que nombrar, porque el orden de esa parte de la
lista lo decide la urgencia y no una clave. Quien quiera meterla ahí tiene que darle clave antes a
alguna de las dos vecinas, exactamente lo mismo que le propone la línea de comandos.

Nada de esta página está en la versión 1.0, porque `biso board` queda fuera de ella
(["Qué hay implementado y qué no"](../estado-de-implementacion.md)).

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

