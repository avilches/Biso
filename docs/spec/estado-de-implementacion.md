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
- **especificado, no implementado**: la especificación ya lo define entero, con sus mensajes
  literales y sus códigos de salida, y no hay ni una línea de código que lo ejerza. Es el estado de
  una decisión que se escribió en la especificación antes de implementarla, y el paso siguiente es
  la tarea que la lleva al programa.
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
| 4 | [`init`](cmd/init.md) y [`where`](cmd/where.md) | hecho | TASK-12 |
| 5 | [`new`](cmd/new.md), [`ls`](cmd/ls.md), [`get`](cmd/get.md), [`set`](cmd/set.md) | hecho | TASK-13 |
| 6 | [Los verbos del ciclo](cmd/verbos-del-ciclo.md) | hecho | TASK-14 |
| 7 | [`prime`](cmd/prime.md) y la medida real del presupuesto de arranque | hecho | TASK-15 |
| 8 | El lote de `new --from`, [`export`](cmd/export.md), [`snapshot`](cmd/snapshot.md) e `init --from` | hecho | TASK-16 |
| 9 | El resto: [`archive`](cmd/archive.md), [`config`](cmd/config.md), [`doctor`](cmd/doctor.md), [`help`](cmd/help.md) | hecho con matices | TASK-17 |
| | El orden manual ([abajo](#el-orden-manual-especificado-y-todavía-no-implementado)) | especificado, no implementado | sin tarea |
| | [`board`](cmd/board.md) | fuera de alcance de 1.0 | sin tarea |

## Los documentos transversales

Estos no son un paso propio: los ejerce cada tarea de la tabla de arriba que toca un campo o un
comando concreto, así que se dan por hechos solo cuando todas las tareas que los usan están
cerradas.

| Documento | Se ejerce en los pasos | Estado |
|---|---|---|
| [Vocabulario de esta especificación](vocabulario.md) | ninguno; fija nombres, no comportamiento | guía, no verificable |
| [Los principios](principios.md) | todos | guía, no verificable |
| [Códigos de salida](codigos-de-salida.md) | todos los que devuelven un código de error | hecho |
| [Flags globales](cmd/flags-globales.md) | pasos 4 a 9 | hecho con matices |
| [Entorno y configuración de máquina](invocacion.md) | pasos 4 a 9 | hecho |
| [Cómo se elige el tablero](resolucion-del-tablero.md) | pasos 4 y 9 | hecho |
| [Terminal, flujos de salida y codificación](salida-y-terminal.md) | pasos 4, 5 y 7 | hecho con matices |
| [Cómo se pasa un valor](valores-de-entrada.md) | pasos 2 y 4 a 9 | hecho |
| [Orden de escritura, concurrencia y datos dañados](garantias.md) | pasos 1, 2, 4, 5, 8 y 9 | hecho |
| [El arrendamiento de una tarea](lease.md) | pasos 2, 5 y 6 | hecho |
| [Los presupuestos de arranque y de tamaño](presupuestos.md) | pasos 1 y 7 | hecho |
| [Cómo se resuelve una referencia a una tarea](referencias.md) | pasos 5 a 9 | hecho |
| [Las familias de flags](familias-de-flags.md) | pasos 5 y 6 | hecho |
| [El contrato JSON](contrato-json.md) | pasos 4 a 9 | hecho con matices |
| [El contrato de estabilidad](estabilidad.md) | pasos 7 y 8 | hecho |
| [Lo que se deja fuera a propósito](fuera-de-alcance.md) | ninguno | fuera de alcance |

### Los tres matices de la tabla de arriba

Ninguna fila queda ya "en curso": ese estado significa, según la leyenda de esta página, que hay una
tarea abierta que lo cubre, y las nueve subtareas de la implementación están cerradas. Las tres que
quedan con matices son estas, y el matiz es siempre el mismo tipo de cosa, algo que la
especificación describe y que la versión 1.0 acepta sin llegar a ejercer:

- **[Flags globales](cmd/flags-globales.md)** y **[Terminal, flujos de salida y
  codificación](salida-y-terminal.md)**: `--color` y `NO_COLOR` se analizan, se validan y deciden
  correctamente si habría color, pero **ninguna salida de la versión 1.0 lleva color**, porque la
  especificación fija cuándo lo habría y no dice en ningún sitio qué se pinta. Está escrito en esa
  misma página y hay una prueba que comprueba que no cambian un byte.
- **[El contrato JSON](contrato-json.md)**: dos de sus `code`, `no_terminal` y `port_in_use`, están
  reservados para [`biso board`](cmd/board.md) y la versión 1.0 no los emite nunca. No se quitan
  porque ["El contrato de estabilidad"](estabilidad.md) fija que las entradas son permanentes.

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
de llegar ahí que hay hoy, y las tres están probadas.

La otra mitad, la que se ve desde fuera, la cerró el paso 8: el aviso
`warning: 1 task could not be read and was skipped: MYP-2` con sus identificadores (y, con más de una, la forma plural con `were`), y el código 6 de
`biso export` y de `biso snapshot` cuando han saltado alguna, cada uno con su prueba. Del texto
exacto del error de la lectura dirigida no queda nada pendiente de implementar, porque la
especificación no lo fija: el código y la clave sí están.

**La garantía 5 la completó también el paso 2 en el único camino que se la saltaba**: crear el
fichero del tablero desde varias conexiones a la vez daba un código 8 instantáneo, porque SQLite
rechaza el paso a modo WAL sin consultar el tiempo de espera. Ahora ese camino espera el tiempo
configurado antes de rendirse, y lo prueban veinte intentos de cuatro conexiones simultáneas.

**Del modelo de datos**, el paso 2 hace cumplir al escribir el título obligatorio (código 2,
`missing_title`), la distinción entre un campo `string` de una línea y un campo `text` (código 2,
`malformed_string_value`, sobre `title`, `author` y el texto de un criterio),
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

De ["`biso init`"](cmd/init.md) quedó fuera **`--from`**, y no fue un descuido sino alcance declarado
del paso 8 (TASK-16), porque restaurar una instantánea necesita el formato de intercambio que trae ese
paso. Mientras tanto el flag se analizaba y sus incompatibilidades se comprobaban, pero la llamada
terminaba con código 1. **Desde el paso 8 ese hueco ya no existe**, y por eso la fila del paso 4
dice "hecho". Todo lo demás de esa página estaba desde el
principio: las
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

De los transversales, el paso 4 no ejerce **el color** de
["Terminal, flujos de salida y codificación"](salida-y-terminal.md), y ningún paso posterior lo
ejerce tampoco: la decisión está implementada y probada contra su tabla, y nada la llama, porque
ninguna salida de la versión 1.0 lleva color. No es una deuda de este paso ni de `ls` y `prime`, que
ya están: la especificación fija cuándo habría color y no dice en ningún sitio qué se pinta, así
que esa página lo declara ahora como el comportamiento de la versión y no como algo pendiente
(["Interactividad, terminal y color"](salida-y-terminal.md#interactividad-terminal-y-color)).
`--color` y `NO_COLOR` se aceptan y se validan, y una prueba comprueba que no cambian un byte. De ["El contrato JSON"](contrato-json.md) están el sobre, los dos
esquemas de estos comandos y el sobre de error con sus claves de detalle, `warnings` incluida; falta
el resto de los `kind`. Y de ["Flags globales"](cmd/flags-globales.md) están todos salvo `--print` en
un comando que sí afecta a alguna tarea, que no existe todavía.

### Qué dejó hecha la primera mitad del paso 5

La primera mitad del paso 5 son los comandos que escriben, [`biso new`](cmd/new.md) y
[`biso set`](cmd/set.md), y con ellos las tres piezas transversales que hacían falta para
escribirlos y que los demás comandos reusarán tal cual:

- **[Las familias de flags](familias-de-flags.md) enteras**, en una sola tabla de `internal/cli` que
  todo comando de escritura toma igual, y un solo motor en `internal/ops` que las aplica en los ocho
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

- **El lote de `biso new --from`** era alcance declarado del paso 8 (TASK-16), porque necesita el
  formato de intercambio que trae ese paso, y hasta entonces el flag ni siquiera se analizaba. Desde
  ese paso está implementado.

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
  una referencia, que hoy son `get`, `set` y el `--parent` de `ls`.
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

**Lo único que faltaba del paso 5 era el lote de `biso new --from`, que no formaba parte de ese
paso.** Es alcance declarado del paso 8 (TASK-16), porque
necesita el formato de intercambio que trae ese paso, y hasta entonces el flag ni siquiera se
analizaba: escribirlo respondía `unknown flag: --from` con código 2, que es la respuesta de un flag
que el programa no conoce y no la de un flag documentado que todavía no hace nada. **Desde el paso 8
existe**, con todas las reglas que documenta ["`biso new`"](cmd/new.md#el-modo-lote), así que esa
fila dice "hecho". Todo lo demás de las cuatro páginas está implementado y probado.

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
  defecto de su `--status`. El código ya lo hacía; el texto se podía leer de las dos maneras y ahora lo dice
  explícitamente (["Cómo se resuelve una referencia a una tarea"](referencias.md#la-búsqueda-por-texto)).
- Las otras dos eran huecos de prueba y no de comportamiento: el paso 2 del orden de aplicación (los
  `--replace-*`) no lo cubría ninguna prueba, y una frase de ["`biso new`"](cmd/new.md#salida) seguía diciendo que la
  ficha de `--print` venía debajo de la línea del identificador en vez de sustituirla.

### Qué dejó hecho el paso 6

Los seis verbos del ciclo están enteros, y la forma en que lo están es lo que más conviene saber:
**ninguno tiene ruta de escritura propia**. `biso set` se partió en un bucle compartido que recibe
un `verb` con tres momentos (antes de aplicar las flags de campo, después de los ocho pasos de
["Orden de aplicación dentro de una escritura"](garantias.md#orden-de-aplicación-dentro-de-una-escritura),
y después de arreglar el arrendamiento), más un cuarto que `biso answer` usa para meter sus dos
comentarios por delante de cualquier `--comment` de la misma llamada. Los seis toman la misma tabla
de flags de campo que `biso new` y `biso set`, así que la promesa de que una flag significa lo mismo
en todas partes no depende de que nadie la copie bien.

Tres piezas del motor son nuevas y sirven a más de un verbo:

- **Los avisos pueden llevar líneas de detalle.** Las usa el de criterios sin marcar, que lista
  debajo los que faltan, en cualquier comando que llegue al estado terminal y no solo en
  `biso finish` (["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)).
- **Una escritura puede llevar reclamaciones condicionales de arrendamiento**, que `internal/board`
  comprueba dentro de la misma transacción que escribe las filas. Es lo que hace cierto que de dos
  reclamaciones simultáneas de un arrendamiento vencido solo gane una; la que pierde no escribe nada
  y sale con código 8 y la clave `lease_lost`, que es nueva en
  ["Los identificadores de error"](contrato-json.md#los-identificadores-de-error).
- **Los argumentos posicionales de texto pasan por las tres formas de
  ["Cómo se pasa un valor"](valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo)**, así que
  `biso note MYP-11 @findings.md` vale lo mismo que `biso set MYP-11 --append-note @findings.md`. Lo
  que la regla del posicional que parece un identificador juzga es lo que se escribió, no lo que el
  fichero traía, que es lo que mantiene abierta esa vía de escape.

La especificación se corrigió en el mismo cambio, con los textos literales que le faltaban: el error
6 de `biso start` sobre una tarea ya terminada, el de `biso finish --strict` con todo lo que falta
nombrado de una vez, el del posicional que parece un identificador en `biso comment`, y el de una
llamada sin texto, que estrena la clave `missing_text`. Las tres cosas que la página no decidía y hubo
que decidir son esas mismas: qué dice cada uno de esos errores.

La revisión del paso corrigió cuatro cosas más, tres de ellas también en la especificación:

- **Los avisos de cierre de `biso finish` son los de llegar a un estado terminal**, así que con `--status`
  hacia otro estado no sale ninguno, ni siquiera convertido en el error 6 de `--strict`. La página
  del verbo no tenía fila para ese caso y ahora la tiene (["`biso finish`"](cmd/verbos-del-ciclo.md#biso-finish)).
- **La llamada sin ningún texto de `biso ask` y de `biso answer` tiene su bloque literal**, como ya lo
  tenía la de `biso note`, y el de `answer` dice `needs an answer` y no `needs a answer`.
- **El error `missing_text` no lleva `field` ni `given`.** Esas dos claves son de los errores que
  nombran un flag, una clave de configuración o un valor concreto, y un posicional que nadie escribió
  no es ninguno de los tres (["Los errores en JSON"](contrato-json.md#los-errores-en-json)).
- **`biso comment` con solo `--comment-author` es esa misma llamada sin texto**, porque ese flag firma
  un comentario pero no escribe ninguno.

Y una quinta que venía del paso 5 y se arregló aquí, por ser el mismo camino de código: **`--dry-run`
ejecuta también la última capa de validación**, la del modelo. Una vista previa ya no puede salir con
0 donde la escritura de verdad sale con 3, que es lo que prometen ["Flags globales"](cmd/flags-globales.md)
y la sección de salida de ["`biso set`"](cmd/set.md#salida).

### Qué dejó hecho el paso 7

`biso prime` está entero, con sus dos salidas, sus cuatro bloques, su recorte en cascada y su
esquema JSON. Lo que más conviene saber es cómo quedaron las dos cifras de
["Los presupuestos de arranque y de tamaño"](presupuestos.md), que son las que este paso existía para cerrar:

- **El tope de tamaño se comprueba sobre el mensaje que imprime el proceso**, no sobre lo que
  devuelve una función, y se comprueba en siete sitios: el tablero del ejemplo, un tablero con una
  tarea ilegible, un tablero vacío, cuatro tableros de 300 tareas que fuerzan el recorte escalón a
  escalón, un tablero sin ninguna tarea pero con cuarenta estados, cuarenta tipos y cuarenta
  prioridades de nombre largo, y un tablero cuyo nombre son diez mil caracteres. Los dos últimos son
  los que hacen cierta la palabra "siempre": los cuatro primeros escalones solo recortan tareas, así
  que un tablero sin ninguna se escapaba del tope por el bloque `BOARD` (imprimía 8.465 bytes) hasta
  que se añadieron los escalones 6 y 7. Con el tablero del ejemplo el mensaje mide **5.070 bytes** de
  los 5.504, **3.581** de parte fija y **1.489** de resumen, que es exactamente lo que dice
  ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño); con el de los vocabularios
  largos, **5.217**, y con el del nombre kilométrico, **5.245**.
- **La medida del presupuesto de arranque ya es la de verdad**, en `cmd/biso/budget_test.go`:
  ejecuta el binario compilado, `biso ls` y `biso prime`, sobre un tablero real de 300 tareas con
  los cuatro bloques poblados, y mide de la llamada al código de salida, arranque del proceso
  incluido. En la máquina de desarrollo da unos **13 ms** para cada uno de los dos, entre 12 y 14
  según la ejecución, frente al tope de 25. Se queda con la mediana de cinco ejecuciones tras una de
  calentamiento, y **bajo el detector de carreras se salta declarándolo**, igual que la medida del
  paso 1: el proceso que se mide no lleva el detector, porque `go build` lo compila sin él, pero sí
  lo lleva todo lo que corre alrededor, y esa no es la máquina ociosa de la que habla la cifra. La
  medida sintética del paso 1 sigue donde estaba, en `internal/board`, y ahora dice lo que es: el
  suelo de esta, útil porque una regresión ahí señala el almacén y no un punto cualquiera del
  camino.

**El mensaje ofrecía un comando que no existía, y ya existe.** Su bloque `COMMANDS` empieza por
`` `biso help <cmd>...` for the detail of any ``, y [`biso help`](cmd/help.md) era del paso 9, que
lo cerró: hasta entonces esa línea contestaba `unknown command: "help"` con código 2, y hoy
contesta la ayuda de los comandos que se le nombren, varios en la misma llamada.

Cuatro cosas que la especificación no decidía y hubo que decidir, todas escritas en su página antes
de escribir el código:

- **La parte fija del mensaje es literal, ejemplos incluidos** (["Lo que no depende del tablero"](cmd/prime.md#lo-que-no-depende-del-tablero)).
  El `MYP-12` de la regla 2 y el `Done` de la regla 4 ilustran la forma de un identificador y la de
  un estado terminal y no describen el tablero que se tiene delante, que ya está descrito dos
  bloques más arriba. Es lo que hace exacta la aritmética del presupuesto: la parte fija mide
  siempre lo mismo y todo el margen que queda es para el resumen.
- **Qué se ve del recorte en cascada** (["El recorte en cascada"](cmd/prime.md#el-recorte-en-cascada)): un bloque que se queda sin filas
  pierde también su encabezado y se queda en su línea de recuento, el recorte es del texto y no toca
  la salida de `--json`, y el tope se mide sobre el mensaje sin `--full`. El bloque `BOARD` se
  recorta el último y solo cuando los cuatro bloques de tareas ya no pueden dar más: sus tres listas
  pierden elementos, primero `priorities`, luego `types` y por último la línea de recuento por
  estado, cada una diciendo `+N more`, y si ni así cabe, el bloque se corta en seco con tres puntos.
  Son los escalones 6 y 7 de ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño).
- **El texto de `--full`** (["`--full`"](cmd/prime.md#--full)), que la especificación mencionaba sin escribir. Agrupa los
  flags por el campo que escriben y en el orden en que una escritura los aplica, y se genera desde
  la tabla única de `internal/cli/fields.go`, así que un flag nuevo aparece ahí sin que nadie tenga
  que acordarse. La rejilla `FIELD FLAGS` del mensaje, en cambio, sí es texto literal: sus doce
  líneas son una disposición que la tabla no lleva dentro. Lo que la ata a la tabla es una prueba de
  `internal/cli` que compara las dos en los dos sentidos y se pone roja en cuanto un flag entra en
  una y no en la otra.
- **Tres casos límite pequeños**: un vocabulario vacío escribe `(none)`, la línea `unreadable` va la
  última del bloque `BOARD`, y "tablero vacío" quiere decir sin ninguna tarea, archivadas incluidas,
  así que un tablero cuyas tareas están todas terminadas no imprime ni los cuatro bloques ni el
  texto que los sustituye. Un `--limit` negativo contesta el mismo error que el de `biso ls`, que es
  el mismo flag. El `(none)` no es letra muerta: el estado se alcanza con
  `biso config set types ""` sobre un tablero donde ningún valor de esa lista está en uso, un caso
  que ["`biso config`"](cmd/config.md#comportamiento-caso-a-caso) recoge ahora explícitamente, y la
  prueba lo construye escribiendo la configuración porque ese comando es del paso 9.

**Por qué [el contrato de estabilidad](estabilidad.md) queda "hecho".** Ese documento congela
ocho cosas, y las ocho existen hoy en código y tienen prueba: los códigos de salida y los
identificadores `code`, el nombre y el significado de cada comando y de cada flag, las claves de
`data` en cada `kind` de JSON, el algoritmo de coincidencia idéntico al leer y al escribir, la
estabilidad de las claves de los criterios, el tope de tamaño del mensaje de `biso prime`, y las
dos simetrías, la de `biso export` con `biso new --from` y la de `biso snapshot` con
`biso init --from`. Las dos últimas las cierra el paso 8, y su prueba compara un volcado de las
dos bases de datos, no dos salidas del mismo codificador, que es lo que hace que pueda fallar.

### Qué dejó hecho el paso 8

El paso 8 es el lote de comandos que mueve un tablero entero de un sitio a otro, y van juntos
porque la garantía de simetría los necesita a la vez: [`export`](cmd/export.md), el modo lote de
[`biso new --from`](cmd/new.md#el-modo-lote), [`snapshot`](cmd/snapshot.md) e
[`init --from`](cmd/init.md).

**El formato de intercambio está definido una sola vez**, en `internal/ops/interchange.go`: una
estructura con sus etiquetas de serialización, una función que la escribe y otra que la lee, y la
lista de claves admitidas deducida de esas mismas etiquetas en vez de repetida a mano. `board.json`
tiene su propia mitad, con la misma regla, en `internal/ops/boardfile.go`. Es lo que hace que la
prueba de simetría signifique algo: no hay una segunda implementación del formato que pueda
desincronizarse de la primera.

**La prueba de simetría son dos, porque las dos vías ejercitan código distinto.** La de
`export` con `new --from` importa el volcado en un tablero que declara el mismo vocabulario a mano,
y la de `snapshot` con `init --from` restaura un tablero entero sin declarar nada, en una máquina con
otras raíces, que es donde el identificador de la instantánea está libre.

**Lo que las dos comparan es un volcado de las dos bases de datos, tabla por tabla y columna por
columna, y no las dos exportaciones.** Comparar las exportaciones era comparar la salida de una
función con la salida de esa misma función: un campo que el codificador dejara de escribir
desaparecía igual en los dos lados y la prueba seguía en verde, cosa que se comprobó poniendo a nulo
el instante de creación y viendo pasar la suite entera. El volcado lo escribe SQL de la propia
prueba y no comparte nada con el programa, así que ese campo perdido sale como dos tableros
distintos. El formato en sí lo fija aparte un fichero de referencia escrito a mano
(`cmd/biso/testdata/export-line.txt`) con la línea exportada de una tarea con todos sus campos
puestos: una clave que desaparezca del formato deja de coincidir con él. Las dos mutaciones con las
que se comprobó que la prueba nueva sí falla son esa fecha de creación a nulo y la clave estable de
un criterio renumerada al vuelo.

Comparar las bases enteras descubrió además una diferencia real entre dos tableros que se comportan
igual: una tarea creada por `biso new` guardaba sus dos contadores de claves a 0 y una importada a 1.
Ahora la traducción a filas los normaliza a 1, que es el valor que la ida y vuelta reproduce.

La revisión adversarial del paso corrigió once cosas más, y estas cambiaron la especificación: la
regla real del contador de claves, que se pierde en cuanto se quita el criterio o el comentario de
clave mayor y no solo al quitarlos todos
(["`biso export`"](cmd/export.md#el-contador-de-claves-no-es-una-clave-del-formato)); que restaurar
una instantánea cuyo identificador ya vive en una raíz de esta máquina es el error de identidad
duplicada aunque `--at` apunte fuera de las raíces, con su mensaje literal, lo que corrigió el
ejemplo de la garantía de simetría que hacía justo eso (["`biso init`"](cmd/init.md),
["`biso export`"](cmd/export.md#la-garantía-de-simetría)); que los fallos de un lote salen siempre en
orden de línea, con un fallo de grafo en el bloque de ejemplo para que se vea; y el mensaje de un
`id` ocupado, que ahora es distinto según lo tenga el tablero o una línea anterior del mismo fichero
(["`biso new`"](cmd/new.md#el-modo-lote)). En el código quedaron además la clave `unknown_key` con su
código 2 y no envuelta en un `invalid_line`, la urgencia y la lista de lo que cambió en el sobre de
cada tarea del lote, y la lista de claves cuyo vacío es `[]` o `{}` deducida por reflexión como ya lo
estaban las claves admitidas.

Lagunas que se cerraron en la especificación al implementarlos, todas escritas en su página antes de
escribir el código: el texto literal con el que `export` rechaza `--json`
(["`biso export`"](cmd/export.md#el-rechazo-de---json)); que los dos contadores de claves se deducen
y qué caso no reproducen (["`biso export`"](cmd/export.md#el-contador-de-claves-no-es-una-clave-del-formato));
que el autor de una tarea importada nunca es quien importa, que `parent` y `dependencies` son
identificadores comprobados contra el tablero y contra el propio fichero, y que cada línea aporta un
solo fallo al informe (["`biso new`"](cmd/new.md#el-modo-lote)); la tercera línea de la salida de
`--vcs push` y la abreviatura del identificador de la revisión
(["`biso snapshot`"](cmd/snapshot.md#salida)); de qué marcador sale la identidad de un tablero
restaurado (["`biso init`"](cmd/init.md)); y los cinco `code` que faltaban en
["Los identificadores de error"](contrato-json.md#los-identificadores-de-error), `invalid_snapshot_id`
incluido, que `init` ya nombraba y esa lista no llevaba.

### Qué dejó hecho el paso 9

Los comandos del último paso están enteros: `biso archive`, `biso config`, `biso doctor` y
`biso help`. Con ellos, el mensaje de `biso prime` deja de ofrecer nada que no exista, y la única
página de comando que sigue siendo solo texto es [`biso board`](cmd/board.md), fuera de alcance de
la versión 1.0 por decisión explícita de `TASK-55`.

Lo que conviene saber de cómo quedaron:

- **`biso archive` no tiene ruta de escritura propia**, como los seis verbos del paso 6: es el mismo
  bucle compartido con un campo que ningún flag escribe, sus dos notas idempotentes y el aviso por
  cada tarea viva que dependía de la archivada. Vaciar el arrendamiento ya lo hacía el motor.
- **`biso config` valida en dos mitades**, que es como su tabla de casos los separa: primero si
  el valor es un valor de su clave, que es el código 3 y no lee ninguna tarea, y después si el
  tablero seguiría entero con él, que es el código 6 y las lee todas. La segunda mitad es la misma
  pregunta que [`biso doctor`](cmd/doctor.md) hace al revés: lo que este comando no deja que pase es
  lo que aquel reporta cuando pasó de todas formas.
- **`biso help` resuelve contra un catálogo**, que es la lista de todos los comandos con el resumen
  de una línea que imprime la ayuda de primer nivel. Un nombre está en el catálogo aunque esta
  compilación no lleve todavía la lógica detrás, así que `biso help board`, `biso help export` y
  `biso help snapshot` contestan su texto: la ayuda de un comando es parte de la interfaz que fija la
  especificación, y `biso help all` ya imprime esos nombres.
- **`biso doctor` no ejecuta la comprobación de integridad al abrir**, sino dentro del comando, por
  la razón por la que existe el presupuesto de arranque: recorre el fichero entero.

**El sondeo del sistema de ficheros estaba especificado de una forma que no podía funcionar, y la
página se corrigió.** Su paso 2 pedía comprobar que un bloqueo por rango de bytes se hace cumplir
tomándolo desde un descriptor y pidiéndolo desde un segundo, los dos del mismo proceso; pero un
bloqueo de registro de POSIX pertenece al proceso y no al descriptor, así que el segundo se concede
siempre, por definición, y el sondeo llamaba inseguro a cualquier disco local. Ahora esa mitad se
pregunta con `F_OFD_SETLK`, el mismo bloqueo ligado al descriptor abierto, en Linux y en macOS, y se
salta en un Unix que no lo tenga. Vive en un paquete propio, `internal/walprobe`, con pruebas que
afirman lo que un sondeo roto no dice: que un directorio local corriente es seguro y que el fichero
de prueba no se queda atrás.

**De ahí sale la única dependencia externa nueva**, `golang.org/x/sys`, que ahora es directa y no
heredada del controlador de SQLite. La autoriza la propia página de
[`biso doctor`](cmd/doctor.md#el-sondeo-del-sistema-de-ficheros), que la nombra por el mismo motivo
por el que el controlador no usa `cgo`: es Go puro y compila de forma cruzada.

**Lo que no está**, además de `biso board`: **el sondeo no tiene mitad de Windows**. La página
describe `LockFileEx` y `CreateFileMapping`, y ahí hoy no hay sondeo ninguno, así que `biso doctor`
en Windows nunca levanta ese aviso. Eso es lo correcto mientras no exista: el aviso dice que el
sondeo corrió y algo falló, y llamar inseguro a lo que no se ha medido sería peor que callar. Es el
motivo, con `board`, de que la fila del paso 9 diga "hecho con matices".

**Y no está el color**, que no es de este paso sino de todos: la tabla de decisión de
["Interactividad, terminal y color"](salida-y-terminal.md#interactividad-terminal-y-color) está
implementada y probada, y no la llama nadie, porque ninguna salida lleva códigos de escape. A
diferencia de las dos de arriba, esto ya no es un hueco pendiente: la especificación dice desde esta
revisión que la versión 1.0 no emite color, y una prueba comprueba que `--color always` y
`NO_COLOR` dan la misma salida byte a byte. Lo que faltaría para pintar algo no es código, es la
decisión de qué se pinta, que ninguna página toma.

Lo que la especificación no decidía, o decidía mal, y se corrigió en el mismo cambio:

- **`--print` es error 2 también en `biso doctor --fix`.** La página decía que ahí "no añade nada",
  que se podía leer como que se acepta y no hace nada; la regla general de
  ["Flags globales"](cmd/flags-globales.md#flags-globales) dice que ni `--print` ni `--dry-run` se
  ignoran nunca en silencio, así que se acepta esa y se dice explícito en los dos sitios.
- **`no_such_command` es una clave nueva** de
  ["Los identificadores de error"](contrato-json.md#los-identificadores-de-error), con código 4: es
  lo que contesta `biso help` ante un nombre que no existe, y no el `unknown_command` de código 2 del
  analizador, porque ahí la llamada está bien formada y lo que no existe es su argumento. En la misma
  lista se dejó dicho que `doctor_problems` no viaja nunca en un sobre de error.
- **El ejemplo JSON de [`biso help`](cmd/help.md#el-esquema-json) llevaba unos resúmenes** que no
  eran los que imprime la ayuda de primer nivel, cuando el párrafo de al lado decía que son los
  mismos. Ahora lo son, y una prueba dorada los compara.
- **La nota de vaciar una lista de configuración** dejaba un espacio suelto detrás del igual, donde
  `config list` y `config get` no lo dejan; y su vista previa habría dicho "would be set to " sin
  nada detrás, así que dice "would be emptied" (["`biso config`"](cmd/config.md#comportamiento-caso-a-caso)).
- **`biso archive --unarchive` con una referencia de texto no podía funcionar.** La búsqueda de
  ["Cómo se resuelve una referencia a una tarea"](referencias.md#la-búsqueda-por-texto) mira solo
  las tareas no archivadas, y lo que esa llamada nombra está archivado por definición, así que
  contestaba siempre que el texto no encaja con ninguna tarea. Es ahora la única excepción de esa
  regla, y mira el tablero entero para que nombrar por texto una tarea que ya está en el tablero dé
  la misma nota idempotente que nombrarla por su identificador
  (["La referencia de `--unarchive`"](cmd/archive.md#la-referencia-de---unarchive)).

Y dos arreglos de código que la revisión del paso no habría visto desde fuera:

- **`biso doctor` comparaba el marcador `<id>.id` con una copia de sí mismo.** Cuando el tablero se
  abre desde dentro de su propio directorio, el `id` de la localización sale del marcador, así que la
  fila del marcador discrepante no se disparaba nunca. Ahora lee el `id` de la base de datos, que es
  contra lo que esa fila compara.
- **La prueba de `biso prime` que comprueba el `(none)` de un vocabulario vacío** escribía la lista
  vacía directamente en la tabla de configuración, porque `biso config set types ""` no existía. Ya
  existe, y ahora la prueba pasa por ahí, que era la única puerta a ese estado. Lo mismo con el
  ayudante que archivaba una tarea a golpe de `UPDATE` en las pruebas de los verbos del ciclo.

### Qué cerró la revisión del paso 9

La revisión del paso encontró dos fallos que no eran de los comandos de ese paso sino de capas de abajo, y
que se habían escapado precisamente porque eran de todos y de ninguno. Los dos están arreglados en la
capa donde viven, `internal/store`, y probados desde varios comandos a la vez:

- **Un fallo de escritura por permisos salía con código 1 y el mensaje crudo del controlador.** El
  almacén solo clasificaba el tablero ocupado, la base de datos corrupta y el fichero que no es una
  base de datos; el fichero de solo lectura, el disco lleno y el fallo de entrada/salida caían al error
  genérico. Ahora son el código 8 con la clave `io_error`, que es lo que
  ["Códigos de salida"](codigos-de-salida.md) promete para un entorno que falla, y la
  especificación estrena la sección que fija su mensaje
  (["Qué pasa cuando el almacén no se puede escribir"](garantias.md#qué-pasa-cuando-el-almacén-no-se-puede-escribir)).
  La otra mitad del mismo camino, un directorio de tablero sin permiso de escritura, salía con 21 y un
  hueco donde iría el `id`, porque el modo WAL necesita escribir para abrirse y ahí todavía no hay `id`
  que leer: sale también con 8, y el mensaje nombra el fichero, que siempre se conoce.
- **Un daño que solo se ve al leer una página salía con código 1.** Un fichero que no es una base de
  datos falla al abrirse y siempre dio 21, pero uno cuya cabecera está intacta y cuyas páginas no
  falla a mitad de una consulta, y ese error subía sin clasificar. La garantía de
  ["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)
  es de todos los comandos, así que la clasificación está ahora en el almacén y alcanza también a lo
  que falla recorriendo filas ya obtenidas.

De `biso doctor` cerró lo que le faltaba de red y de texto:

- **Las comprobaciones que `--fix` repara tienen sus dos formas escritas**, la de encontrado y la de
  reparado, en la tabla normativa y con prueba de fichero dorado cada una
  (["Las dos formas de un mensaje reparable"](cmd/doctor.md#las-dos-formas-de-un-mensaje-reparable)). Dos de las tres decían
  algo que la especificación no recogía.
- **Dos casos de borde que la especificación no cubría**: un contador a cero no es el identificador
  `MYP-0` y el mensaje ya no se lo inventa, y una lista de vocabulario vacía ya no se cita con unas
  comillas sin nada dentro.
- **El informe sale en el orden de la tabla de comprobaciones**, que es lo que su comentario ya
  prometía y lo que ahora fija la especificación
  (["El orden en que sale el informe"](cmd/doctor.md#el-orden-en-que-sale-el-informe)). Las comprobaciones no corren en ese orden,
  porque las que miran una tarea se hacen en una sola pasada.
- **Las comprobaciones que no tenían ninguna prueba ya la tienen**: la clave de extensión no
  declarada, los dos papeles de estado, el ciclo de dependencias, el de tarea padre y el aviso del
  sistema de ficheros inseguro, que era el único que nadie había visto dispararse nunca. Y la sección
  de atomicidad de una reparación con varios arreglos, la única que existe solo por este comando,
  tiene ahora una prueba que la ejerce entera: la transacción de datos queda aplicada, la escritura
  del marcador falla, el comando sale con 8 y el marcador vuelve a aparecer como error.

Las que siguen sin prueba son las que el esquema de hoy contesta por construcción,
los identificadores duplicados y las claves de criterio repetidas, y el código lo dice donde están.

Y **["Cómo se elige el tablero"](resolucion-del-tablero.md) acota a qué vía se refiere** cuando dice
que un tablero sin marcador tiene que poder abrirse para que `biso doctor --fix` se lo devuelva: habla
de la primera vía, la del directorio de trabajo, y no de las dos. El puntero sigue exigiendo el
marcador aunque su `path` resuelva, porque ahí manda el `id`, y esa única puerta basta para que el
arreglo tenga algo que arreglar.

### Qué cerró la auditoría de la especificación contra el binario

Con los nueve pasos ya mezclados, una auditoría recorrió `docs/spec/` entera comparando cada ancla
con lo que hace el programa. Encontró nueve divergencias, y estas son las que cambiaron algo:

- **El límite de filas configurable no existía.** `biso ls` usaba siempre treinta, y la clave
  `default_limit` se leía del fichero de máquina, se validaba y se tiraba. Ahora la precedencia
  entera está implementada, con prueba de sus cuatro escalones, y `BISO_LIMIT` se valida como se
  valida esa clave (["Variables de entorno"](invocacion.md#variables-de-entorno)).
- **El color no se emite, y ahora la especificación lo dice.** Ver el párrafo del paso 9 de más
  arriba.
- **`dry_run_failed` no lo producía ningún camino**, y solo vivía en un comentario. Un `--dry-run`
  que no pasa contesta el código de su fallo, que es la decisión deliberada del paso 5, y el único
  7 que existe es el del lote. El identificador se ha quitado de
  ["El contrato JSON"](contrato-json.md), y con él las filas del 7 de las tablas de códigos y de
  las ayudas de `set`, `archive` y los seis verbos del ciclo, que prometían un desenlace inalcanzable.
- **Una instantánea que fallaba al publicar se tragaba lo que sí había pasado.** Ahora imprime la
  línea de los ficheros escritos y las notas que le tocaban, con el error detrás
  (["`biso snapshot`"](cmd/snapshot.md#salida)).
- **La regla de una sola entrada estándar se comprobaba después de leerla.** El primer `-` vaciaba el
  flujo y emitía su aviso antes de que la llamada fuera rechazada; la cuenta se hace ahora sobre la
  línea entera y antes de leer ningún valor
  (["Tres formas de pasar un valor largo"](valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo)).
- **Tres arreglos pequeños de texto y de forma**: la pista de un identificador que el tablero ya no
  tiene sale partida en sus tres líneas, como la imprime su página; el ejemplo del identificador mal
  formado de ["Cómo se resuelve una referencia a una tarea"](referencias.md) lleva ahora el `--id`
  que hace falta para llegar a él, porque sin él la gramática de esa misma página manda la cadena a
  la búsqueda por texto; y los dos fallos del arrendamiento de un lote ya no llevan una clave `given`
  vacía, que es una segunda excepción declarada de
  ["Los errores en JSON"](contrato-json.md#los-errores-en-json).

### Qué dejó hecha la retirada de las formas cortas

La tarea `TASK-69` aplicó la regla de
["Una forma corta solo existe si nadie más reclama su inicial"](../decisiones/comandos-y-flags.md#una-forma-corta-solo-existe-si-nadie-más-reclama-su-inicial)
al programa y a la especificación:

- **Las formas cortas de la tabla de la decisión ya no existen.** Usarlas es un error 2 con el `code`
  `unknown_flag`. La medida de la tarea contaba seis, y al escribir la prueba que aplica la regla a
  todos los comandos apareció una más, la de `--out` en `biso export`, que chocaba con `--overdue`.
  Las de los flags globales se quedan.
- **Las tablas de parámetros de `docs/spec/cmd/` ya no tienen la columna `Corto`**, salvo la de los
  flags globales, que es donde la forma corta sigue existiendo.
- **`biso new` acepta el título por `--title` y rechaza darlo también como argumento**, con el `code`
  `incompatible_flags` (["`biso new`"](cmd/new.md#comportamiento-caso-a-caso)), y `biso new --help`
  lista `--title`.
- **Las pruebas** están en `cmd/biso/short_flags_test.go`, que ejecuta el binario con cada forma
  retirada, con el choque de título y con la ayuda de `new`, y en `internal/cli/short_flags_test.go`,
  que recorre la tabla de todos los comandos y falla si una forma corta la reclama otro flag.
- **El mensaje de arranque medía entonces 5.089 bytes** con el tablero del ejemplo, cincuenta más que
  antes, porque sus flags se escriben enteros
  (["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño)). Esa cifra ya no es la
  vigente: la retirada de `documentation` la bajó a 5.038
  (["Qué dejó hecha la fusión de `documentation` en `references`"](#qué-dejó-hecha-la-fusión-de-documentation-en-references))
  y la de `modifiedFiles` a 4.983
  (["Qué dejó hecha la retirada de `modifiedFiles`"](#qué-dejó-hecha-la-retirada-de-modifiedfiles)).
  La regla 11 del mensaje de arranque la subió después a 5.103
  (["Qué dejó hecha la ayuda de la dirección de una dependencia"](#qué-dejó-hecha-la-ayuda-de-la-dirección-de-una-dependencia)),
  y la retirada de `ext` la bajó de nuevo. La cifra vigente es la de
  ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño).

### Qué dejó hecha la fusión de `documentation` en `references`

La tarea `TASK-75` aplicó
["Se retira `documentation` y `references` queda como único campo de punteros"](../decisiones/detalles.md#se-retira-documentation-y-references-queda-como-único-campo-de-punteros)
al programa y a la especificación. Las anclas que cubre son el modelo de datos, las familias de
flags, los valores de entrada, el contrato JSON de `get` y `ls`, la ficha de `biso get`, las ayudas
de `biso set` y `biso prime`, el mensaje de arranque, el modo lote de `biso new` y el formato de
`biso export`:

- **El campo `documentation` ya no existe en ningún sitio.** Ni en el modelo, ni en la lista de
  campos permitidos de las listas del almacén, ni en el JSON de `get`, `ls` y `export`, ni en la
  ficha, ni en los flags `--add-docs`, `--rm-docs`, `--clear-docs` y `--replace-docs`, que ahora
  son un error 2 con el `code` `unknown_flag` como cualquier flag que no existe. Como `biso` no se ha
  publicado, se ha editado el primer script de migración en vez de añadir uno.
- **`biso new --from` acepta `documentation` y lo funde en `references`**, con el aviso
  `imported_documentation_merged`, con el mismo patrón que `definitionOfDone`: la clave no cuenta como
  desconocida, `null` equivale a ausente, la lista vacía no avisa, y el aviso nombra la línea y
  cuenta los elementos (["`biso new`"](cmd/new.md#el-modo-lote)). Decisión propia de la
  implementación, anotada en la especificación: un valor que ya estaba en `references` se guarda una
  sola vez. Las pruebas están en `internal/ops/batch_test.go`.
- **El mensaje de arranque medía entonces 5.038 bytes** con el tablero del ejemplo, 51 menos que antes: la
  parte fija pasaba de 3.600 a **3.549** y el resumen seguía en **1.489**. La rejilla `FIELD FLAGS` pasó de
  trece líneas a doce. Esa cifra tampoco es ya la vigente: la retirada de `modifiedFiles` la bajó a 4.983.
- **La cobertura de `compatibilidad-de-modelos.md` se recalculó** sin el campo retirado, y el porcentaje de cada sistema cambió en consecuencia.

### Qué dejó hecha la retirada de `modifiedFiles`

La tarea `TASK-80` aplicó
["Se retira `modifiedFiles`"](../decisiones/detalles.md#se-retira-modifiedfiles)
al programa y a la especificación. Las anclas que cubre son el modelo de datos, las familias de
flags, los valores de entrada, el contrato JSON de `get`, `ls` y `export`, la ficha de `biso get`, las
ayudas de `biso set`, `biso finish` y `biso prime`, el mensaje de arranque, el modo lote de `biso new`,
el formato de `biso export`, la tabla de avisos y la fuera de alcance:

- **El campo `modifiedFiles` ya no existe en ningún sitio.** Ni en el modelo, ni en la lista de
  campos permitidos de las listas del almacén, ni en el JSON de `get`, `ls` y `export`, ni en la
  ficha, ni en los flags `--add-files`, `--rm-files`, `--clear-files` y `--replace-files`, que ahora
  son un error 2 con el `code` `unknown_flag` como cualquier flag que no existe. `biso finish` deja
  de anunciar `--add-files` en su firma, su ayuda y su descripción: quien cierre una tarea y quiera
  dejar una ruta usa `--add-refs`, que ya funcionaba ahí como todo flag de campo. Como `biso` no se ha
  publicado, se ha editado el primer script de migración en vez de añadir uno.
- **`biso new --from` acepta `modifiedFiles` y lo funde en `references`**, con el aviso
  `imported_modified_files_merged`, con el mismo patrón que `documentation`: la clave no cuenta como
  desconocida, `null` equivale a ausente, la lista vacía no avisa, y el aviso nombra la línea y cuenta
  los elementos (["`biso new`"](cmd/new.md#el-modo-lote)). Decisión propia de la implementación, anotada
  en la especificación: si una línea trae `documentation` y `modifiedFiles` a la vez, el orden en
  `references` es fijo, primero lo que ya había, luego `documentation` y por último `modifiedFiles`,
  con ambos avisos en ese mismo orden. Las pruebas están en `internal/ops/batch_test.go`.
- **La simetría entre `biso export` y `biso new --from` sigue pasando sin el campo.** Las dos
  pruebas que llevaban un `modifiedFiles` en su tablero (`symmetry_test.go` y `export_test.go`) lo
  sustituyen por un segundo valor de `references`, para que la lista siga teniendo más de un elemento.
- **El mensaje de arranque medía entonces 4.983 bytes** con el tablero del ejemplo, 55 menos que antes: la
  parte fija pasaba de 3.549 a **3.494** y el resumen seguía en **1.489**. La rejilla `FIELD FLAGS` pasa de
  doce líneas a once, y es la línea entera de los flags del campo lo que se libera. El tope total de 5.504
  no cambia y el margen de la parte fija sube de 291 a 346 bytes. Esa cifra ya no es la vigente: la regla 11
  la subió a 5.103, y la retirada de `ext` la bajó después.
- **La cobertura de `compatibilidad-de-modelos.md` se recalculó** sin el campo retirado: un campo menos en el total sobre el que se calcula cada porcentaje.
- **La familia de los campos de lista queda en cuatro**: `labels`, `assignees`, `references` y
  `dependencies`.

### Qué dejó hecha la ayuda de la dirección de una dependencia

La tarea `TASK-77` aplicó
["La ayuda enseña la dirección de una dependencia"](../decisiones/detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia)
al programa y a la especificación. Las anclas que cubre son la ayuda de `biso set`, la de `biso new`,
la página de relaciones entre tareas y el presupuesto de tamaño del mensaje de arranque:

- **La ayuda de `biso set` tiene un bloque `Relations`** que glosa `--parent`, `--add-deps` y
  `--add-refs`, dice que una dependencia se escribe en la tarea que espera y da el ejemplo
  `biso set MYP-10 --add-deps MYP-4` para "`MYP-4` bloquea a `MYP-10`", con la advertencia de que la
  arista contraria es igual de válida y el programa no puede avisar de ella. La frase que decía que
  ningún flag exigía aprender nada más allá de su nombre se reformula.
- **La ayuda de `biso new` glosa `--add-deps`, `--parent` y `--add-refs`** en sus flags más usados,
  dice que `--add-deps` nombra lo que va antes de la tarea nueva y explica cómo se consigue lo
  contrario (crearla y luego `biso set MYP-10 --add-deps <new id>`). Lleva un ejemplo con `--parent` y
  `--add-deps`.
- **Las dos ayudas, la especificación y los ficheros de `cmd/biso/testdata/` llevan el mismo texto**,
  y la comparación carácter a carácter que ya existía lo comprueba. La prueba nueva
  `TestTheHelpOfSetAndNewSaysWhichWayADependencyPoints` comprueba además que las frases de la
  dirección están, para que un retoque posterior de la ayuda no las pierda sin que falle nada.
- **El mensaje de arranque no llevó nada de esto en la primera versión, y ahora lleva una regla.** La
  implementación de la tarea no tocó `FIELD FLAGS` ni `RULES`, porque la tarea decía que el sitio natural es
  la ayuda de `set` y de `new`, y el mensaje seguía midiendo 4.983 bytes. Esa decisión ya no es la vigente:
  quien solo lee `biso prime` no tenía ninguna pista de la dirección de una dependencia, y la regla 11 de
  `RULES` la dice en dos líneas (["La ayuda enseña la dirección de una dependencia"](../decisiones/detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia)).
  Con el tablero del ejemplo el mensaje medía entonces **5.103 bytes**, 120 más: **3.614** de parte fija,
  con un margen de 226 bytes sobre los 3.840, y **1.489** de resumen. Con el de los vocabularios largos
  medía 5.250 y con el del nombre kilométrico 5.278, los dos por debajo del tope de 5.504. Esas cifras ya
  no son las vigentes: la retirada de `ext` las bajó después, y las de hoy están en
  ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño). La prueba de que
  la parte fija no pasa de 3.840 bytes y el total de 5.504 ya existía
  (`TestTheFixedPartOfTheMessageIsAlwaysTheSame` y `assertWithinBudget`), y
  `TestTheBudgetConstantsAreTheNumbersOfTheSpecification` compara las dos constantes con los números que
  imprime `presupuestos.md`, para que nadie suba el tope para hacer sitio.
- **La prueba de la ayuda comprueba también las glosas de `--add-deps`**, la de `set` (`so each blocks it`) y
  la de `new` (`so each blocks the new task`), de modo que invertirlas falla aunque se edite a la vez el
  código, los ficheros de `cmd/biso/testdata/` y la especificación.
- **No hay `--add-blocks`**, y la razón queda escrita en la decisión.

### Qué dejó hecha la retirada de `ext`

La tarea `TASK-87` aplicó ["Se retira `ext`"](../decisiones/detalles.md#se-retira-ext) al programa y a la
especificación, en ese orden. Las anclas que cubre son el modelo de datos, las familias de flags, el orden
de escritura, la configuración, `init`, `doctor`, el intercambio de `new` y `export`, `get`, `ls`, `prime`
y el contrato JSON:

- **El campo `ext` ya no existe en ningún sitio.** Ni en el modelo, ni en la tabla `task_ext`, ni en su
  lectura, ni en el JSON de `get`, `ls` y `export`, ni en la ficha, ni en los flags `--ext`, `--rm-ext` y
  `--clear-ext`, que ahora son un error 2 con el `code` `unknown_flag` como cualquier flag que no existe.
  Un lote de `biso new --from` que traiga la clave `ext` falla como con cualquier clave desconocida. Como
  `biso` no se ha publicado, se ha editado el primer script de migración en vez de añadir uno, igual que
  en las retiradas anteriores.
- **Los tableros creados antes del cambio dejan de abrirse.** Su configuración lleva una fila
  `extensions` que el programa ya no conoce, y una clave desconocida es un dato que no se puede
  interpretar, código 3. Solo afecta a tableros de ensayo de esta máquina, sin ningún dato real.
- **El orden de aplicación de una escritura pasa de nueve pasos a ocho.** Desaparece el quinto, el de los
  campos de mapa, y los escalares, los marcados de criterios, la corrección de fechas y los comentarios
  suben un puesto. Lo comprueba `TestTheStepsOfBothLayersLineUp`.
- **La configuración pasa de veinte claves a diecinueve.** `biso init` pierde `--extensions`, y `biso
  config list` y `biso prime --json` dejan de llevar `extensions`.
- **`biso doctor` pierde su comprobación de claves de extensión no declaradas**, con su `code`
  `undeclared_extension_key`. La prueba del orden del informe usa ahora un tipo no configurado, que se
  encuentra en la misma pasada sobre las tareas.
- **La forma clave y valor deja de existir en la línea de comandos**, salvo en `--set-comment-date`, que
  corta por el último `=`. El analizador pierde la categoría de paso de `ext`, su alfabeto y la regla de
  que la última clave repetida gana, y con ella el aviso `duplicate_ext_key`.
- **El mensaje de arranque medía entonces 5.070 bytes** con el tablero del ejemplo, 33 menos que antes: la parte
  fija pasaba de 3.614 a **3.581** y el resumen seguía en **1.489**. La rejilla `FIELD FLAGS` perdió la línea de
  los flags del campo. El tope total de 5.504 no cambió y el margen de la parte fija subió de 226 a 259
  bytes. Las cifras vigentes son siempre las de
  ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño).
- **La simetría entre `biso export` y `biso new --from` sigue pasando.** Las pruebas que llevaban un `ext`
  en su tablero lo dejan de llevar, y se eliminan las que solo ejercían `ext`.

### El orden manual: especificado y todavía no implementado

La especificación ya recoge entera la decisión
["El orden manual es una clave de texto"](../decisiones/detalles.md#el-orden-manual-es-una-clave-de-texto),
y **el programa no sabe nada de ella**: hoy `biso` guarda `ordinal` como un entero no negativo, lo
acepta escrito con `--ordinal 3000`, rechaza el negativo con `invalid_number` y no conoce ni
`--above` ni `--below`. Esta entrada está aquí para que nadie lea las páginas de arriba como una
descripción de lo que el binario hace ahora mismo.

Las anclas que la especificación cubre ya, y que la tarea de implementación tendrá que ejercer:

- **La clave y su algoritmo**, en ["El orden manual y su clave"](modelo-de-datos/orden-manual.md):
  qué es una clave, cómo se comparan dos, la función del punto medio con su tabla de casos, el hueco
  que le toca a cada flag, el bloque de varias tareas en una llamada y por qué no hace falta
  renumerar.
- **Los flags**, en ["El orden manual"](familias-de-flags.md#el-orden-manual), con sus casos en la
  tabla de ["`biso set`"](cmd/set.md#comportamiento-caso-a-caso) y en la de
  ["`biso new`"](cmd/new.md#comportamiento-caso-a-caso), y el paso de escritura que les toca en
  ["Orden de aplicación dentro de una escritura"](garantias.md#orden-de-aplicación-dentro-de-una-escritura).
- **La lectura**, en ["La regla de orden, completa"](cmd/ls.md#la-regla-de-orden-completa), en el
  sobre de ["`biso ls`"](cmd/ls.md#el-esquema-json) y en el de
  ["`biso get`"](cmd/get.md#el-esquema-json), más la celda `manual` de la ficha de texto
  (["`biso get`"](cmd/get.md#salida)).
- **El intercambio**, con la clave escrita y validada en
  ["El modo lote"](cmd/new.md#el-modo-lote) y devuelta tal cual por
  ["`biso export`"](cmd/export.md#la-garantía-de-simetría).
- **Los `code` nuevos**, en
  ["Los identificadores de error"](contrato-json.md#los-identificadores-de-error):
  `invalid_ordinal_value`, `self_ordinal_neighbour` y `malformed_ordinal` con el código 2, y
  `neighbour_without_ordinal` con el 6. Ningún aviso nuevo.
- **El arrastre de [`biso board`](cmd/board.md#qué-escribe-un-arrastre)**, que sigue fuera de la
  versión 1.0 y por tanto no lo ejerce nadie todavía.

**El mensaje de arranque crece, y sus cifras hay que volver a medirlas al implementarlo.** El orden
manual pasa a tener su propia línea en la rejilla `FIELD FLAGS`, que gana una línea y pierde
`--ordinal --clear-ordinal` de la de los escalares. Contando el texto literal de
["La salida literal"](cmd/prime.md#la-salida-literal) con el tablero del ejemplo, el mensaje pasa de
5.070 a 5.112 bytes y la parte fija de 3.581 a 3.623 de los 3.840 que tiene asignados, que es lo que
dice ahora ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño); el resumen no
cambia. La medida que manda es la del proceso, así que la tarea que implemente esto tiene que
confirmarla con la prueba del presupuesto y corregir esa página si sale otra cosa. La lista de
`--full` gana los flags nuevos en su línea de `ordinal`, y no cuenta para el tope.

**Lo que hay que hacer con los tableros de desarrollo.** El campo cambia de tipo, así que un tablero
creado con un binario anterior guarda enteros donde el binario nuevo espera claves. Como `biso` no se
ha publicado, vale la misma salida que en las retiradas anteriores: editar el primer script de
migración en vez de añadir uno, y recrear los tableros de ensayo de esta máquina.

## Antes de empezar un paso

Al planificar la tarea de un paso (el plan que se registra antes de tocar código, según
`backlog instructions task-execution`), el plan enumera las anclas exactas de la especificación que
va a cubrir, tomadas de las dos tablas de arriba, y se confirma con quien encargó la tarea antes de
empezar a escribir código. Es el punto en el que se corrige esta página si el alcance real resulta
distinto del que aquí consta.
