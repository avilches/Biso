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

## Los pasos de implementación

El orden de estos nueve pasos, y el porqué de ese orden, está en la tarea `TASK-55` del tablero de
Backlog.md, de la que cada tarea de la tabla es una subtarea.

| Paso | Qué cubre | Estado | Tarea |
|---|---|---|---|
| 1 | El almacén SQLite, sus migraciones y `WithTx` | hecho | TASK-4 |
| 2 | El modelo de datos lógico ([`modelo-de-datos/`](modelo-de-datos/index.md)) y la garantía de identificadores únicos | hecho | TASK-10 |
| 3 | El algoritmo de coincidencia ([`vocabularios.md`](vocabularios.md#el-algoritmo-de-coincidencia)) y el de sugerencias ([`vocabularios.md`](vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas)), en `internal/match` | hecho | TASK-11 |
| 4 | [`init`](cmd/init.md) y [`where`](cmd/where.md) | en curso | TASK-12 |
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
| [Códigos de salida](codigos-de-salida.md) | todos los que devuelven un código de error | en curso |
| [Flags globales](cmd/flags-globales.md) | pasos 4 a 9 | en curso |
| [Entorno y configuración de máquina](invocacion.md) | pasos 4 a 9 | en curso |
| [Cómo se elige el tablero](resolucion-del-tablero.md) | paso 4 | en curso |
| [Terminal, flujos de salida y codificación](salida-y-terminal.md) | pasos 4, 5 y 7 | en curso |
| [Cómo se pasa un valor](valores-de-entrada.md) | pasos 2 y 4 a 9 | en curso |
| [Orden de escritura, concurrencia y datos dañados](garantias.md) | pasos 1, 2 y 4 | en curso |
| [El arrendamiento de una tarea](lease.md) | pasos 2 y 6 | pendiente |
| [Los presupuestos de arranque y de tamaño](presupuestos.md) | pasos 1 y 7 | en curso |
| [Cómo se resuelve una referencia a una tarea](referencias.md) | pasos 5 a 9 | pendiente |
| [Las familias de flags](familias-de-flags.md) | pasos 5 y 6 | pendiente |
| [El contrato JSON](contrato-json.md) | pasos 4 a 9 | en curso |
| [El contrato de estabilidad](estabilidad.md) | paso 7 | pendiente |
| [Lo que se deja fuera a propósito](fuera-de-alcance.md) | ninguno | fuera de alcance |

### Qué queda probado ya de los transversales que empezaron los pasos 1 y 2

**De [`garantias.md`](garantias.md)**, el paso 1 dejó probadas en `internal/store` las garantías 1
(ninguna escritura se observa a medias: un lector concurrente no ve la fila sin confirmar), 2 (todo
o nada, incluido cuando la función entra en pánico), 4 y 5 (el bloqueo de escritura se espera el
tiempo configurado y luego falla con código 8 y el texto literal, desde cualquier camino: abrir,
migrar o escribir) y 6 (una lectura no se bloquea ni falla por una escritura en curso). También
quedó implementado el segundo caso de ["Qué pasa con un dato que no se puede
interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar): una base de datos que
no se puede leer falla con código 21, la clave `database_unreadable` y sus dos `hint`, y la
comprobación de integridad que nombra ese texto existe como un método propio del almacén, que es lo
que ejecutará [`biso doctor`](cmd/doctor.md) en el paso 9.

**La garantía 3 (dos procesos simultáneos nunca asignan el mismo identificador) la cierra el paso 2**,
que es cuando existen identificadores que asignar. Se asignan con un contador propio del tablero que
solo crece, dentro de la misma transacción de escritura del paso 1, y la prueba que lo comprueba
abre cuatro almacenes independientes sobre el mismo fichero, cada uno con su conexión, y les hace
crear cien tareas a la vez: ninguna repite identificador y no falta ninguno. De ahí sale también que
un identificador no se reutilice nunca, ni tras archivar la tarea ni tras desaparecer su fila.

El primer caso de esa misma sección, la tarea suelta ilegible, lo cierra el paso 2 en su sustancia:
una lectura de conjunto (`Tasks.All`) nunca aborta por una tarea mala, la deja fuera del listado y la
devuelve aparte con su motivo, y una lectura dirigida (`Tasks.Load`) de esa misma tarea falla con
código 3 y la clave `undecodable_task`. Una fecha que el programa no escribió, un nombre de campo de
lista que el modelo no conoce y una prioridad que la configuración ya no declara son las tres formas
de llegar ahí que hay hoy, y las tres están probadas. El paso 2 rechaza además al escribir una clave
de extensión que el tablero no declara (código 3, `unknown_extension_key`) y una clave fuera del
alfabeto (código 2, `malformed_extension_key`).

Lo que queda de ese caso para los pasos que implementan `ls`, `get`, `export` y `snapshot` es lo que
se ve desde fuera: el texto literal del aviso `warning: 1 task could not be read and was skipped` con
sus identificadores, el código 6 de `biso export` y `biso snapshot` cuando han saltado alguna, y el
texto exacto del error de la lectura dirigida, que la especificación todavía no fija.

**La garantía 5 la completó también el paso 2 en el único camino que se la saltaba**: crear el
fichero del tablero desde varias conexiones a la vez daba un código 8 instantáneo, porque SQLite
rechaza el paso a modo WAL sin consultar el tiempo de espera. Ahora ese camino espera el tiempo
configurado antes de rendirse, y lo prueban veinte intentos de cuatro conexiones simultáneas.

**Del modelo de datos**, el paso 2 hace cumplir al escribir el título obligatorio (código 2,
`missing_title`), la distinción entre un campo `string` de una línea y un campo `text` (código 2,
`malformed_string_value`, sobre `title`, `author`, el texto de un criterio y los valores de `ext`),
el alfabeto de las etiquetas y las personas (`malformed_label` y `malformed_assignee`), el
`ordinal` no negativo (`invalid_number`) y las claves de los criterios. Todo eso se comprueba antes
de abrir la transacción, así que una tarea rechazada no gasta identificador.

**De [`presupuestos.md`](presupuestos.md)**, la medida vive desde el paso 2 en `internal/board` y ya
es sobre el esquema real: abrir el tablero y leer sus 300 tareas enteras, con sus listas, sus
criterios y sus comentarios, tarda unos 2,7 ms frente al tope de 25 que fija
["El presupuesto de arranque"](presupuestos.md#el-presupuesto-de-arranque). Bajo el detector de
carreras la misma lectura tarda unos 80 ms, y esa ejecución **no compara contra ningún otro número**:
la prueba se salta declarándolo, porque el detector multiplica por más de un orden de magnitud cada
lectura de un controlador de SQLite escrito en Go puro y medir contra él mediría el detector. La cifra
de la especificación la afirma la ejecución sin detector, que es la única que puede. Sigue sin ser la
medida del presupuesto: esa es sobre `biso ls` y `biso prime` en el binario compilado y es del paso 7
(TASK-15).
["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño) es entero del paso 7.

### Qué falta del paso 4, que está en curso

El paso 4 dejó el programa ejecutable: `cmd/biso/main.go`, el reparto de `internal/cli` en analizador
y capa de salida, `internal/ops` con la lógica de cada comando, y el tipo `Board` de `internal/board`
con la resolución entera de ["Cómo se elige el tablero"](resolucion-del-tablero.md). De
["`biso init`"](cmd/init.md) queda una sola cosa fuera: **`--from` no está implementado**, porque
restaurar una instantánea necesita el formato de intercambio que trae el paso 8 (TASK-16). El flag se
analiza, sus incompatibilidades se comprueban, y la llamada termina con código 1 diciendo que todavía
no existe. Todo lo demás de esa página está: las dos adopciones de identidad, `--overwrite-config`
con sus errores de código 6, `--dry-run`, las dos notas de `--at`, el fichero de exclusión y el
marcador.

De los transversales, el paso 4 no ejerce todavía **el color** de
["Terminal, flujos de salida y codificación"](salida-y-terminal.md): la decisión está implementada y
probada contra su tabla, pero ninguna salida de `init` ni de `where` lleva color, así que nada la usa
hasta que lleguen `ls` y `prime`. De ["El contrato JSON"](contrato-json.md) están el sobre, los dos
esquemas de estos comandos y el sobre de error con sus claves de detalle; falta el resto de los
`kind`. Y de ["Flags globales"](cmd/flags-globales.md) están todos salvo `--print` en un comando que
sí afecta a alguna tarea, que no existe todavía.

## Antes de empezar un paso

Al planificar la tarea de un paso (el plan que se registra antes de tocar código, según
`backlog instructions task-execution`), el plan enumera las anclas exactas de la especificación que
va a cubrir, tomadas de las dos tablas de arriba, y se confirma con quien encargó la tarea antes de
empezar a escribir código. Es el punto en el que se corrige esta página si el alcance real resulta
distinto del que aquí consta.
