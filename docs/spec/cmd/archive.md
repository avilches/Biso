# `biso archive`

## Firma

```
biso archive <ref>... [--unarchive] [--id] [--match]
             [cualquier bandera de campo de la seccion 8]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--unarchive` | | no | booleano | falso | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Saca la tarea del tablero activo. **La tarea sigue existiendo**, su identificador sigue reservado,
`biso get` la encuentra avisando de que está archivada, y `biso ls --archived` la lista.

| Caso | Qué pasa |
|---|---|
| Ya estaba archivada | Código 0, con un `note:`, sin escribir |
| Otras tareas vivas dependen de ella | Aviso con la lista, se archiva igual |
| La tarea tiene arrendamiento, vivo o vencido | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea (séptima precisión de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md)). Si estaba vivo y era de otra identidad, sale además el aviso de ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) |
| `--unarchive` | La devuelve al tablero con el estado que tenía, y sin arrendamiento: si vuelve al estado activo, quien quiera trabajar en ella lo toma con `biso start` |
| Varias referencias | Todo o nada |

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
TASK-11  Done  ac 2/2  dod 1/1  urgency 0.0  archived
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Archivada o desarchivada | 0 |
| `biso delete`, o banderas incompatibles | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

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
  0  archived       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso archive TASK-11
  biso archive TASK-11 TASK-12 TASK-13
  biso archive TASK-11 --unarchive
```

---

