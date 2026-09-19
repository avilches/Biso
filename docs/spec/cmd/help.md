# La ayuda: `biso help` y `biso --help`

## `biso help`

### Firma

```
biso help [<command>... | all]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<command>` | | no | nombre de comando, o `all` | ninguno | sí | no | |

| Caso | Qué pasa |
|---|---|
| Sin argumento | Imprime la ["ayuda de primer nivel"](#la-ayuda-de-primer-nivel), igual que `biso --help` |
| Con uno o varios nombres de comando, todos existentes | Imprime la ayuda completa de cada uno, en el orden pedido, en la misma llamada |
| Con `all`, y ningún nombre de comando además | Imprime la ayuda de primer nivel más la lista de los comandos de administración, cada uno con su línea |
| Con `all` junto a uno o más nombres de comando | Error 2: `all` no se combina con nombres de comando |
| Con uno o varios nombres, y alguno no existe | Error 4, con hasta tres de los nombres más parecidos al primero que no existe en el orden pedido (los que pasen el umbral de ["El algoritmo de sugerencias más parecidas"](../vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas)); no imprime la ayuda de ningún comando de la llamada, ni siquiera de los que sí existen |

`biso help` funciona **sin tablero**.

### Varios comandos en una sola llamada

`biso help finish note` imprime la ayuda completa de `finish` y a continuación la de `note`, en el
orden en que se pidieron, cada una igual que si se hubiera llamado por separado con
`biso <cmd> --help`. Entre el bloque de un comando y el del siguiente se intercala una línea en
blanco.

**La validación es todo o nada, y ocurre antes de imprimir nada.** Los nombres pedidos se resuelven
en el orden en que aparecen, y en cuanto uno no existe, la llamada entera falla con el error de
siempre (hasta tres nombres más parecidos a ese, no a los demás) y no se imprime la ayuda de ningún
comando de la llamada, ni la de los que existen antes de él en la lista ni la de los que vienen
después. Es el mismo criterio de todo o nada de la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables),
aplicado aquí a una lectura en vez de a una escritura: una llamada que pide varias cosas a la vez no
entrega la mitad y calla el resto.

```
$ biso help fnish
error: no such command: "fnish"
hint: did you mean: finish, init?
```

Solo salen dos nombres aunque el tope sea tres: ningún otro comando pasa el umbral de
["El algoritmo de sugerencias más parecidas"](../vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas), y la lista se queda corta en vez de rellenarse con algo que no se parece.

**`all` es un valor especial, no un nombre de comando, y no se combina con ninguno en la misma
llamada.** Pide una lista de comandos, no la ayuda completa de uno, y las dos formas de salida no
tienen una combinación con un significado único: no hay un orden razonable entre "la lista" y "la
ayuda completa de X", y quien quiere las dos cosas ya puede pedirlas en dos llamadas. Combinarlos es
un error de uso:

```
error: help: "all" cannot be combined with a command name
hint: run `biso help all` on its own, or list only command names
```

### Salida de `biso help all`

Es la de la ["ayuda de primer nivel"](#la-ayuda-de-primer-nivel), seguida de:

```
Administration:
  init               create a board for this project
  where              say which board is in use and why
  archive <ref>      take a task off the board (there is no delete)
  export             dump the board as NDJSON that `biso new --from` reads back
  config             read and change the board configuration
  doctor             check the board, and repair what can be repaired
  board              open the interactive board
  help [cmd|all]     this
  snapshot           write snapshot.ndjson and board.json, then record them
```

### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "help",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "commands": [
      { "name": "prime", "summary": "print the session briefing" },
      { "name": "new", "summary": "create a task" }
    ]
  }
}
```

**En JSON, este comando emite la lista de comandos con su resumen de una línea, y nunca la ayuda en
prosa.** Un texto de ayuda es prosa escrita para leerse, con sus ejemplos y sus párrafos, y meterla en
una clave sería mover el problema de analizarla a otro sitio; la lista de comandos, en cambio, es un dato
y es lo que le sirve a un agente para descubrir la interfaz sin leer nada. Los resúmenes son los mismos
que imprime la ["ayuda de primer nivel"](#la-ayuda-de-primer-nivel).

`biso help --json` y `biso help all --json` traen la lista entera, y `biso help <comando>... --json`
trae un elemento por cada nombre pedido, en el mismo orden, tal y como hoy trae uno solo cuando se
pide un único nombre. La clave `commands` está siempre y siempre es una lista, así que nadie tiene que
mirar el argumento para saber qué forma va a recibir. La ayuda completa de un comando se sigue
pidiendo como siempre, sin `--json`, con `biso help <comando>` o `biso <comando> --help`; `--json` con
varios nombres no cambia esto, sigue trayendo solo el resumen de cada uno, nunca la prosa. `all`
combinado con un nombre de comando falla igual con `--json` que sin él, antes de que se emita ningún
sobre.

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Ayuda impresa | 0 |
| `all` combinado con uno o más nombres de comando | 2 |
| Alguno de los comandos pedidos no existe | 4 |

### `biso help --help`

```
Usage: biso help [command...|all]

Print the top-level help, the help of one or more commands, or the top-level
help plus the nine administrative commands with `all`. Works without a board.

Given more than one command name, prints the full help of each one, in the
order given, in the same call, as if `biso <cmd> --help` had been called once
per name, with a blank line between one command's help and the next. A name
that does not exist fails the whole call before anything is printed, even the
help of the names that do exist earlier in the list. `all` is not a command
name and does not combine with one.

With --json this prints the command list and its one-line summaries, never
the prose help: a help text is written to be read, and the list is the part
that is data. Given several names, `commands` carries one element per name,
in the same order.

Arguments:
  command        one or more command names, or `all` on its own

Options:
      --json     machine-readable envelope with the command list

Exit codes:
  0  help printed
  2  usage error, such as `all` combined with a command name
  4  no such command

Examples:
  biso help
  biso help finish
  biso help finish note
  biso help all
```

---

## La ayuda de primer nivel

`biso --help` y `biso help` imprimen esto, y solo esto:

```
biso 1.0.0 - the task board of this project.

Usage: biso [global options] <command> [options]

Start here:
  prime              everything you need to work on this board, in one message

Daily work:
  ls                 list tasks, most urgent first
  get <ref>          show one task
  new "TITLE"        create a task and print its id
  set <ref>...       change any field
  start <ref>...     take a task
  note <ref> TEXT    append an implementation note
  comment <ref> TEXT append a discussion comment
  finish <ref>...    close a task
  ask <ref> TEXT     park on a question
  answer <ref> TEXT  answer it and unpark

Global options:
  -C, --cwd <path>   resolve the board from there, instead of cd-ing
      --json         machine-readable output
  -q, --quiet        print only ids
      --print        print the whole record after writing
      --color <when> auto (default), always, never
      --dry-run      validate, write nothing (writing commands only)
  -V, --version      print the version
  -h, --help         this, or the help of a command

More: `biso <command> --help`, and `biso help all` for the administrative
commands (init, where, archive, export, config, doctor, board, help, snapshot).
```

Son treinta y una líneas, y no incluyen los comandos de administración.

**`biso` a secas, sin ningún comando y sin ningún flag, imprime eso mismo y sale con 0.** No es un
error de uso: una llamada que no nombra ningún comando no ha pedido nada mal, no ha pedido nada, y lo
que le hace falta a quien la escribe es justo la lista de lo que puede pedir. Sale por stdout, como
cualquier otra ayuda.

---

