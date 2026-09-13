# Qué hay implementado y qué no

Esta página separa la ambición de la especificación (todo lo que hay más arriba en la barra
lateral) de lo que existe hoy en código. No es normativa: no define ningún comportamiento, solo
dice en qué punto está cada parte frente al texto que la define. Se actualiza al cerrar la tarea de
Backlog.md que implementa un paso, nunca antes; si al implementarlo el comportamiento real
terminó siendo distinto del texto original, la spec se corrige en ese mismo cambio y esta tabla
enlaza a la nota que explica la diferencia, en vez de dejar las dos versiones conviviendo sin
avisar.

**Los estados posibles:**

- **pendiente**: ninguna tarea de Backlog.md lo ha tocado todavía.
- **en curso**: hay una tarea abierta que lo cubre.
- **hecho**: la tarea que lo implementó está cerrada, con sus criterios de aceptación verificados.
- **hecho con matices**: implementado, pero con una diferencia frente al texto original de la
  especificación; la fila enlaza a la nota que la explica.
- **guía, no verificable**: el documento fija vocabulario o principios, no un comportamiento que un
  test pueda comprobar.
- **fuera de alcance**: la propia especificación excluye este documento a propósito, así que nunca
  tendrá una fila "pendiente".

## Los pasos de [Por dónde empezar a implementar](por-donde-empezar.md)

| Paso | Qué cubre | Estado | Tarea |
|---|---|---|---|
| 1 | El almacén SQLite, sus migraciones y `WithTx` | pendiente | TASK-4 |
| 2 | El modelo de datos lógico ([`modelo-de-datos/`](modelo-de-datos/index.md)) y la garantía de identificadores únicos | pendiente | TASK-10 |
| 3 | El algoritmo de coincidencia ([`vocabularios.md`](vocabularios.md#el-algoritmo-de-coincidencia)) | pendiente | TASK-11 |
| 4 | [`init`](cmd/init.md) y [`where`](cmd/where.md) | pendiente | TASK-12 |
| 5 | [`new`](cmd/new.md), [`ls`](cmd/ls.md), [`get`](cmd/get.md), [`set`](cmd/set.md) | pendiente | TASK-13 |
| 6 | [Los verbos del ciclo](cmd/verbos-del-ciclo.md) | pendiente | TASK-14 |
| 7 | [`prime`](cmd/prime.md) y la medida real del presupuesto de arranque | pendiente | TASK-15 |
| 8 | El lote de `new --from`, [`export`](cmd/export.md), [`snapshot`](cmd/snapshot.md) e `init --from` | pendiente | TASK-16 |
| 9 | El resto: [`archive`](cmd/archive.md), [`config`](cmd/config.md), [`doctor`](cmd/doctor.md), [`board`](cmd/board.md), [`help`](cmd/help.md) | pendiente | TASK-17 |

## Los documentos transversales

Estos no son un paso propio: los ejerce cada tarea de la tabla de arriba que toca un campo o un
comando concreto, así que se dan por hechos solo cuando todas las tareas que los usan están
cerradas.

| Documento | Se ejerce en los pasos | Estado |
|---|---|---|
| [Vocabulario de esta especificación](vocabulario.md) | ninguno; fija nombres, no comportamiento | guía, no verificable |
| [Los principios](principios.md) | todos | guía, no verificable |
| [Códigos de salida](codigos-de-salida.md) | todos los que devuelven un código de error | pendiente |
| [Banderas globales, entorno y configuración de máquina](invocacion.md) | pasos 4 a 9 | pendiente |
| [Cómo se elige el tablero](resolucion-del-tablero.md) | paso 4 | pendiente |
| [Terminal, flujos de salida y codificación](salida-y-terminal.md) | pasos 5 y 7 | pendiente |
| [Cómo se pasa un valor](valores-de-entrada.md) | pasos 4 a 9 | pendiente |
| [Orden de escritura, concurrencia y datos dañados](garantias.md) | pasos 1, 2 y 4 | pendiente |
| [El arrendamiento de una tarea](lease.md) | pasos 2 y 6 | pendiente |
| [Los presupuestos de arranque y de tamaño](presupuestos.md) | pasos 1 y 7 | pendiente |
| [Cómo se resuelve una referencia a una tarea](referencias.md) | pasos 5 a 9 | pendiente |
| [Las familias de banderas](familias-de-banderas.md) | pasos 5 y 6 | pendiente |
| [El contrato JSON](contrato-json.md) | pasos 4 a 9 | pendiente |
| [El contrato de estabilidad](estabilidad.md) | paso 7 | pendiente |
| [Lo que se deja fuera a propósito](fuera-de-alcance.md) | ninguno | fuera de alcance |

## Antes de empezar un paso

Al planificar la tarea de un paso (el plan que se registra antes de tocar código, según
`backlog instructions task-execution`), el plan enumera las anclas exactas de la especificación que
va a cubrir, tomadas de las dos tablas de arriba, y se confirma con quien encargó la tarea antes de
empezar a escribir código. Es el punto en el que se corrige esta página si el alcance real resulta
distinto del que aquí consta.
