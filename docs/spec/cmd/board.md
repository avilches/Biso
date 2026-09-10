# `biso board`

## Firma

```
biso board [--port <n>] [--no-open]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--port <n>` | | no | entero entre 1024 y 65535 | 6420 | no | no | ninguno |
| `--no-open` | | no | booleano | falso | no | no | ninguno |

La interfaz interactiva, y **el único comando del programa que abre una interfaz**. Ningún otro puede
abrirla, ni la abre nadie por su cuenta.

| Caso | Qué pasa |
|---|---|
| No hay terminal | Error 7: `error: biso board needs a terminal; every other command works without one` |
| El puerto está ocupado | Error 7, diciendo el puerto |
| Puerto fuera de rango | Error 2 |
| `--no-open` | Arranca y solo imprime la dirección |
| El tablero cambia mientras está abierto | La interfaz recarga. Nunca muestra una versión en caché de una tarea que otro proceso ha cambiado |

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
guion que lo arranque con `--no-open` y tenga que saber a qué dirección apuntar. `opened` dice si se ha
abierto un navegador, y es `false` con `--no-open`.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| El servidor se ha parado limpiamente | 0 |
| Puerto fuera de rango, o banderas incompatibles | 2 |
| No hay terminal, o el puerto está ocupado | 7 |
| No hay tablero | 8 |

## `biso board --help`

```
Usage: biso board [options]

Open the interactive board in a browser. This is the only command that opens
an interface: every other one prints text and exits, with or without a
terminal.

Options:
      --port <n>   port to listen on, 1024 to 65535 (default 6420)
      --no-open    print the address and do not open a browser
      --json       machine-readable envelope, printed when the server starts
  -h, --help       show this help

Exit codes:
  0  stopped cleanly
  2  bad usage
  7  no terminal, or the port is taken
  8  no board here

Examples:
  biso board
  biso board --port 7000 --no-open
```

---

