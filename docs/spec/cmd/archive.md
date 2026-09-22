# `biso archive`

## Firma

```
biso archive <ref>... [--unarchive] [--id] [--match]
             [cualquier flag de campo de las familias de flags]
```

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `<ref>` | sí, una o más | referencia | | sí | no | |
| `--unarchive` | no | booleano | falso | no | no | |
| `--id` | no | booleano | falso | no | no | `--match` |
| `--match` | no | booleano | falso | no | no | `--id` |

Saca la tarea del tablero activo. **La tarea sigue existiendo**, su identificador sigue reservado,
`biso get` la encuentra avisando de que está archivada, y `biso ls --archived` la lista.

| Caso | Qué pasa |
|---|---|
| Ya estaba archivada | Código 0, con un `note:`, sin escribir |
| Otras tareas vivas dependen de ella | Aviso con la lista, se archiva igual |
| La tarea tiene arrendamiento, vivo o vencido | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea (["El vaciado"](../lease.md#el-vaciado) de `lease.md`). Si estaba vivo y era de otra identidad, sale además el aviso de ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) |
| `--unarchive` sobre una tarea archivada | La devuelve al tablero con el estado que tenía, y sin arrendamiento: si vuelve al estado activo, quien quiera trabajar en ella lo toma con `biso start` |
| `--unarchive` sobre una tarea que no está archivada | Código 0, idempotente, con `note: MYP-11 was not archived`, calcado del trato que ya recibe "Ya estaba archivada" en la fila de arriba |
| Varias referencias | Todo o nada |

**Archivar una tarea sin terminar desbloquea, en la misma lectura, a quien dependía de ella.** Para
`blocked`, `blocks` y los términos `bloquea`/`bloqueada` de la urgencia, una tarea archivada sin
terminar cuenta como terminada, así que archivarla quita de inmediato el bloqueo de sus
dependientes, sin que nadie escriba nada más (["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)). **Esto no aplica al aviso de
"subtareas sin terminar" de [`biso finish`](verbos-del-ciclo.md#biso-finish) sobre un padre**: una
subtarea archivada sin terminar sigue contando ahí como sin terminar, y el aviso la marca
`(archived)` en la lista para distinguirla de una subtarea viva de verdad. La diferencia es
deliberada: las dependencias hablan de orden de trabajo, y una vez archivada la que bloqueaba ya no
impide avanzar a nadie; el aviso de `finish` habla del alcance del padre, y una subtarea que
desapareció del tablero activo sin haberse marcado nunca hecha es justo lo que ese aviso existe para
sacar a la luz.

## La referencia de `--unarchive`

**Con `--unarchive`, una referencia de texto mira el tablero entero, archivadas incluidas.** Es la
única excepción a la regla de ["La búsqueda por texto"](../referencias.md#la-búsqueda-por-texto), que
para cualquier otra referencia de cualquier otro comando mira solo las tareas que no están
archivadas, y la excepción se explica sola: lo que `--unarchive` nombra está archivado por
definición, así que buscarlo entre las que no lo están no podría encontrarlo nunca. Sin esto,
`biso archive "the parser" --unarchive` contestaría siempre que no encaja con ninguna tarea, que
además es falso.

**Mira el tablero entero y no solo la mitad archivada**, para que nombrar por texto una tarea que ya
está en el tablero siga dando el código 0 y el `note: MYP-11 was not archived` de la tabla de arriba,
igual que si se hubiera nombrado por su identificador: la idempotencia de esa fila no puede depender
de cómo se escribió la referencia. Y si el texto encaja a la vez con una tarea archivada y con una
viva, eso es una ambigüedad de verdad y sale el código 5 con sus candidatas, como cualquier otra.

Una referencia por identificador no cambia aquí ni en ningún sitio: siempre llega a cualquier tarea
del tablero, archivada o no.

## `biso delete` no existe, y su ausencia está especificada

Cualquier invocación de `biso delete`, `biso rm` o `biso remove` termina con código 2 y este mensaje
por stderr, en vez de volcar la lista de comandos:

```
error: there is no delete command, on purpose
hint: `biso archive <ref>` takes it off the board and keeps the history
      an archived task still exists: `biso ls --archived` lists them, and the
      id is never reused
```

## Salida

```
MYP-11  Done  ac 2/2  urgency 0.0  archived
```

Con `--dry-run`, la misma línea marcada como hipotética, con la regla y el encabezado que fija
["`biso set`"](set.md#salida):

```
$ biso archive MYP-11 --dry-run
1 task would be affected (--dry-run)
MYP-11  Done  ac 2/2  urgency 0.0  archived
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Archivada o desarchivada | 0 |
| `biso delete`, o flags incompatibles | 2 |
| Tarea ilegible (el motivo, `code` `undecodable_task`, va en el mensaje; `archive` no cambia ningún campo de vocabulario, así que no tiene la excepción de `biso set`) | 3 |
| Tarea inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 8 |
| No hay tablero | 20 |

## `biso archive --help`

```
Usage: biso archive <ref>... [options]

Take tasks off the board without losing them. An archived task still exists,
`biso get` still finds it, `biso ls --archived` lists it, and its id is never
reused.

Options:
      --unarchive      put them back on the board
      --id / --match   force <ref> to be an id, or free text
  -h, --help           show this help

Every field flag of `biso set --help` works here too.

There is no delete command. Archiving is the way.

Exit codes:
  0  archived       3  the task could not be read      8  could not be written
  2  bad usage      4  not found                      20  no board here
                    5  ambiguous

Examples:
  biso archive MYP-11
  biso archive MYP-11 MYP-12 MYP-13
  biso archive MYP-11 --unarchive
```

---

