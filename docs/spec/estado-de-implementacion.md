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
| 4 | [`init`](cmd/init.md) y [`where`](cmd/where.md) | hecho con matices | TASK-12 |
| 5 | [`new`](cmd/new.md), [`ls`](cmd/ls.md), [`get`](cmd/get.md), [`set`](cmd/set.md) | hecho con matices | TASK-13 |
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
| [Cómo se elige el tablero](resolucion-del-tablero.md) | paso 4 | hecho |
| [Terminal, flujos de salida y codificación](salida-y-terminal.md) | pasos 4, 5 y 7 | en curso |
| [Cómo se pasa un valor](valores-de-entrada.md) | pasos 2 y 4 a 9 | en curso |
| [Orden de escritura, concurrencia y datos dañados](garantias.md) | pasos 1, 2, 4, 5 y 8 | en curso |
| [El arrendamiento de una tarea](lease.md) | pasos 2, 5 y 6 | en curso |
| [Los presupuestos de arranque y de tamaño](presupuestos.md) | pasos 1 y 7 | en curso |
| [Cómo se resuelve una referencia a una tarea](referencias.md) | pasos 5 a 9 | en curso |
| [Las familias de flags](familias-de-flags.md) | pasos 5 y 6 | en curso |
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

### Qué dejó hecho el paso 4, y la única cosa que no

El paso 4 dejó el programa ejecutable: `cmd/biso/main.go`, el reparto de `internal/cli` en analizador
y capa de salida, `internal/ops` con la lógica de cada comando, y el tipo `Board` de `internal/board`
con la resolución entera de ["Cómo se elige el tablero"](resolucion-del-tablero.md), que queda hecha
al cerrarse esta tarea porque ningún otro paso la ejerce.

De ["`biso init`"](cmd/init.md) queda **una sola cosa fuera: `--from` no está implementado**, y no es
un descuido sino alcance declarado del paso 8 (TASK-16), porque restaurar una instantánea necesita el
formato de intercambio que trae ese paso. El flag se analiza y sus incompatibilidades se comprueban,
pero la llamada termina con **código 1** y la clave `internal`, diciendo que todavía no existe: es la
única invocación de todo `biso` que hoy contesta ese código, y es un código que su propia página no
lista, así que quien la lea encontrará una diferencia mientras TASK-16 no cierre. Es también el motivo
de que la fila del paso 4 diga "hecho con matices" y no "hecho". Todo lo demás de esa página está: las
dos adopciones de identidad, `--overwrite-config` con sus errores de código 6, `--dry-run` con sus dos
notas, las dos notas de `--at`, el fichero de exclusión y el marcador.

Del resto del paso conviene saber implementadas estas reglas, comprobadas al revisarlo, porque son
las que un comando posterior heredará sin volver a escribirlas: un directorio cuya base de
datos no se puede leer no cuenta como tablero para `init`, que lo reconstruye ahí mismo adoptando el
`id` de su marcador (["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar));
un aviso nunca se pierde porque la llamada falle, y viaja como texto o dentro del sobre de error según
haya `--json` (["Los errores en JSON"](contrato-json.md#los-errores-en-json)); y una `note:` se sigue
imprimiendo como texto con `--json`, porque ningún sobre la lleva
(["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)).

De los transversales, el paso 4 no ejerce todavía **el color** de
["Terminal, flujos de salida y codificación"](salida-y-terminal.md): la decisión está implementada y
probada contra su tabla, pero ninguna salida de `init` ni de `where` lleva color, así que nada la usa
hasta que lleguen `ls` y `prime`. De ["El contrato JSON"](contrato-json.md) están el sobre, los dos
esquemas de estos comandos y el sobre de error con sus claves de detalle, `warnings` incluida; falta
el resto de los `kind`. Y de ["Flags globales"](cmd/flags-globales.md) están todos salvo `--print` en
un comando que sí afecta a alguna tarea, que no existe todavía.

### Qué dejó hecha la primera mitad del paso 5

La primera mitad del paso 5 son los comandos que escriben, [`biso new`](cmd/new.md) y
[`biso set`](cmd/set.md), y con ellos las tres piezas transversales que hacían falta para
escribirlos y que los demás comandos reusarán tal cual:

- **[Las familias de flags](familias-de-flags.md) enteras**, en una sola tabla de `internal/cli` que
  todo comando de escritura toma igual, y un solo motor en `internal/ops` que las aplica en los nueve
  pasos de ["Orden de aplicación dentro de una escritura"](garantias.md#orden-de-aplicación-dentro-de-una-escritura).
  Los seis verbos del ciclo y `biso archive` son ese mismo motor con un nombre y unos valores por
  defecto encima.
- **[Cómo se resuelve una referencia a una tarea](referencias.md)**, como función interna compartida
  de `internal/ops`: la gramática, `--id` y `--match`, el ámbito de búsqueda de texto y los tres
  desenlaces. **Lo único que falta de esa página es imprimir las candidatas** del error 5 en el
  formato de `biso ls`: el error ya las lleva dentro, y la mitad que las imprime llega con `ls`.
- **[El arrendamiento](lease.md)** en todo lo que toca a una escritura: la renovación del tenedor, el
  aviso del arrendamiento ajeno, el vaciado al perder el estado activo, la última persona asignada o
  al archivar, y la reclamación de `biso new --start`, que es el atajo de `biso start` y deja la misma
  tarea. Lo que queda es `biso start` mismo, del paso 6.

De estos comandos falta lo siguiente, y nada de ello es un descuido:

- **El lote de `biso new --from`** es alcance declarado del paso 8 (TASK-16), porque necesita el
  formato de intercambio que trae ese paso. El flag **no se analiza todavía**, así que escribirlo hoy
  responde `unknown flag: --from`.

Al implementarlos se cerraron varias lagunas de la especificación, todas escritas en su página antes
de escribir el código: el desenlace de `--dry-run` sobre una sola tarea y por qué ahí nunca sale el
código 7 (["`biso new`"](cmd/new.md#--dry-run-sobre-una-sola-tarea)); el mensaje de una referencia de
texto que no encaja con ninguna tarea y qué pasa cuando encaja con dos títulos
(["Cómo se resuelve una referencia a una tarea"](referencias.md#la-búsqueda-por-texto)); los dos
avisos de añadir un valor que ya estaba y de quitar uno que no estaba, que la lista de
["Notas y avisos"](salida-y-terminal.md#notas-y-avisos) se declaraba completa sin llevarlos; los
mensajes literales de los selectores de texto, del rango invertido, del solape de comentarios, de
`--due` y de `--ordinal` (["Las familias de flags"](familias-de-flags.md)); y cómo se calcula
`changed` (["`biso set`"](cmd/set.md#el-esquema-json)).

### Qué dejó hecha la segunda mitad del paso 5

La segunda mitad son los comandos que leen, [`biso ls`](cmd/ls.md) y [`biso get`](cmd/get.md), y con
ellos las piezas que faltaban de estos transversales:

- **El algoritmo de columnas** de ["`biso ls`"](cmd/ls.md#salida) entero, medido en celdas de terminal y no en
  caracteres, con el recorte del título que nunca parte un grafema. La tabla de anchura de Asia
  oriental vive escrita en `internal/cli/width.go`, por la misma razón que la de acentos de
  `internal/match`: las únicas dependencias externas permitidas son el controlador de SQLite y
  `golang.org/x/term`.
- **Las candidatas de una referencia ambigua**, que es lo que le faltaba a ["Cómo se resuelve una referencia a una tarea"](referencias.md): salen por
  stdout en el formato del listado, con su orden y su límite, y lo hacen en todo comando que resuelva
  una referencia, que hoy son `get`, `set` y el `-p` de `ls`.
- **`--print`**, que ya imprime la ficha completa de cada tarea afectada, porque esa ficha es la de
  `biso get`.

De ["El contrato JSON"](contrato-json.md) quedan hechos los `kind` `task.list`, `task.get` y `task.candidates`, con
`data.filters` serializado desde los propios parámetros del comando y no desde una traducción escrita
a mano. Y la tabla de ["El mismo texto vale lo mismo en los dos sentidos"](vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos) queda probada de punta a punta en
los dos sentidos, que era el hueco que el paso 3 no podía cerrar desde dentro de `internal/match`.

Lagunas que se cerraron en la especificación al implementarlos, todas escritas en su página antes de
escribir el código: qué hace `--count` con el límite y qué imprime `--ids`, cómo ordena `--sort
title` y cuánto mide un carácter de formato (["`biso ls`"](cmd/ls.md)); cómo se escribe cada valor del bloque de
metadatos, qué dice el desglose de una tarea sin prioridad y qué trae el sobre de una sección vacía
que se pidió (["`biso get`"](cmd/get.md)); que la ficha de `--print` sustituye a la salida por defecto
(["Flags globales"](cmd/flags-globales.md)); y que las candidatas de una referencia ambigua salen en todos los comandos
(["Cómo se resuelve una referencia a una tarea"](referencias.md)). Además se corrigió una contradicción de
["Los identificadores de error"](contrato-json.md#los-identificadores-de-error), que agrupaba `unknown_section` y `unknown_sort_field` con los códigos 3
mientras las páginas de `ls` y de `get` los daban por errores de uso: se han movido a la fila del
código 2, que es la que dicen esas páginas y la que corresponde a un dominio cerrado del programa.

### Qué le falta al paso 5, y qué corrigió su revisión

**La fila del paso 5 dice "hecho con matices" y no "hecho" por una sola cosa: el lote de
`biso new --from` no existe.** Es alcance declarado del paso 8 (TASK-16), porque necesita el formato
de intercambio que trae ese paso, y hasta entonces el flag **ni siquiera se analiza**: escribirlo hoy
responde `unknown flag: --from` con código 2, que es la respuesta de un flag que el programa no
conoce y no la de un flag documentado que todavía no hace nada. Quien lea ["`biso new`"](cmd/new.md#el-modo-lote), que sí lo
documenta entero, encontrará esa diferencia mientras TASK-16 no cierre. Todo lo demás de las cuatro
páginas está implementado y probado.

La revisión adversarial del paso corrigió seis cosas, y conviene saber en qué quedaron porque cuatro
de ellas cambiaron la especificación:

- **Un filtro de etiqueta o de persona no distingue mayúsculas**, como manda ["Campos de lista que admiten coma"](familias-de-flags.md#campos-de-lista-que-admiten-coma).
  Comparaba letra por letra, así que un tablero cuya única etiqueta fuera `Parser` contestaba error 3
  a `--label parser` en vez del listado. El plegado se queda en las mayúsculas: no toca los acentos,
  porque una etiqueta es un token corto y no prosa.
- **Los avisos de llegar a un estado terminal nombran la tarea**, así que `biso new` los construye
  cuando el identificador ya está asignado. Antes salían con el hueco vacío. De ahí sale la regla
  nueva de que una vista previa de `biso new` no los emite, porque no hay identificador que gastar
  (["`biso new`"](cmd/new.md#--dry-run-sobre-una-sola-tarea)).
- **Los siete términos de `urgencyBreakdown` son decimales**, igual que la `urgency` de la que son
  sumandos. La regla estaba escrita solo para la urgencia, y por eso el desglose se escapó: ahora
  ["Números, fechas y ausencias"](contrato-json.md#números-fechas-y-ausencias) la dice de los dos.
- **Las candidatas de una referencia ambigua salen del tablero entero**, sin filtrar por estado, así
  que una tarea terminada aparece entre ellas aunque `biso ls --search` la deje fuera por el valor por
  defecto de su `-s`. El código ya lo hacía; el texto se podía leer de las dos maneras y ahora lo dice
  explícitamente (["Cómo se resuelve una referencia a una tarea"](referencias.md#la-búsqueda-por-texto)).
- Las otras dos eran huecos de prueba y no de comportamiento: el paso 2 del orden de aplicación (los
  `--replace-*`) no lo cubría ninguna prueba, y una frase de ["`biso new`"](cmd/new.md#salida) seguía diciendo que la
  ficha de `--print` venía debajo de la línea del identificador en vez de sustituirla.

## Antes de empezar un paso

Al planificar la tarea de un paso (el plan que se registra antes de tocar código, según
`backlog instructions task-execution`), el plan enumera las anclas exactas de la especificación que
va a cubrir, tomadas de las dos tablas de arriba, y se confirma con quien encargó la tarea antes de
empezar a escribir código. Es el punto en el que se corrige esta página si el alcance real resulta
distinto del que aquí consta.
