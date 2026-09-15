# `biso archive`

## Firma

```
biso archive <ref>... [--unarchive] [--id] [--match]
             [cualquier flag de campo de las familias de flags]
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
| La tarea tiene arrendamiento, vivo o vencido | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea (["El vaciado"](../lease.md#el-vaciado) de `lease.md`). Si estaba vivo y era de otra identidad, sale además el aviso de ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) |
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
MYP-11  Done  ac 2/2  dod 1/1  urgency 0.0  archived
```

Con `--dry-run`, la misma línea marcada como hipotética, con la regla y el encabezado que fija
["`biso set`"](set.md#salida):

```
$ biso archive MYP-11 --dry-run
1 task would be affected (--dry-run)
MYP-11  Done  ac 2/2  dod 1/1  urgency 0.0  archived
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Archivada o desarchivada | 0 |
| `biso delete`, o flags incompatibles | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente | 4 |
| Referencia ambigua | 5 |
| `--dry-run` que no pasa | 7 |
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
  0  archived       3  the task could not be read     7  --dry-run did not pass
  2  bad usage      4  not found                      8  could not be written
                    5  ambiguous                     20  no board here

Examples:
  biso archive MYP-11
  biso archive MYP-11 MYP-12 MYP-13
  biso archive MYP-11 --unarchive
```

---

