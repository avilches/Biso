# La ayuda: `biso help` y `biso --help`

## `biso help`

### Firma

```
biso help [<command> | all]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<command>` | | no | nombre de comando, o `all` | ninguno | no | no | |

| Caso | Qué pasa |
|---|---|
| Sin argumento | Imprime la ["ayuda de primer nivel"](#la-ayuda-de-primer-nivel), igual que `biso --help` |
| Con un nombre de comando | Imprime la ayuda de ese comando, igual que `biso <cmd> --help` |
| Con `all` | Imprime la ayuda de primer nivel más la lista de los comandos de administración, cada uno con su línea |
| Con un nombre que no existe | Error 4, con los tres nombres más parecidos |

`biso help` funciona **sin tablero**.

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

`biso help --json` y `biso help all --json` traen la lista entera, y `biso help <comando> --json` la trae
con un solo elemento, el de ese comando. La clave `commands` está siempre y siempre es una lista, así que
nadie tiene que mirar el argumento para saber qué forma va a recibir. La ayuda completa de un comando se
sigue pidiendo como siempre, sin `--json`, con `biso help <comando>` o `biso <comando> --help`.

### Códigos de salida

| Desenlace | Código |
|---|---:|
| Ayuda impresa | 0 |
| El comando no existe | 4 |

### `biso help --help`

```
Usage: biso help [command|all]

Print the top-level help, or the help of one command, or the top-level help
plus the nine administrative commands with `all`. Works without a board.

With --json this prints the command list and its one-line summaries, never
the prose help: a help text is written to be read, and the list is the part
that is data.

Arguments:
  command        a command name, or `all`

Options:
      --json     machine-readable envelope with the command list

Exit codes:
  0  help printed
  4  no such command

Examples:
  biso help
  biso help finish
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

---

