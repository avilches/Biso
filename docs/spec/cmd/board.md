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

