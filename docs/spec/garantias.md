# Orden de escritura, concurrencia y datos dañados

## Orden de aplicación dentro de una escritura

Una sola invocación puede tocar muchos campos. El orden en que se aplican es fijo y **no depende del
orden en que aparecen los flags en la línea de comandos**, para que el resultado sea reproducible y
determinista:

1. Todos los `--clear-*`.
2. Todos los `--replace-*`.
3. Todos los `--rm-*`, incluido `--rm-comment`.
4. Los añadidos: `--add-*` y `--append-*`.
5. Los campos escalares.
6. Los marcados de criterios (`--check-ac`, `--uncheck-ac`).
7. `--set-comment-date`.
8. Los comentarios, es decir `--comment`, el que añade.

Con este orden, `--clear-labels --add-labels urgent` deja exactamente una etiqueta, y `--clear-acs
--add-ac "A" --check-ac all` marca el criterio recién puesto. Dentro de un mismo paso manda el orden
de la línea de comandos: `--add-labels b --add-labels a` deja `b` antes que `a`. Las listas nunca se
ordenan solas.

**El orden manual no añade ningún paso: cae en el 5, con los demás escalares.** `--ordinal first`,
`--ordinal last`, `--above` y `--below` escriben el campo `ordinal` igual que `--due` escribe el
suyo, y `--clear-ordinal` es un `--clear-*` corriente del paso 1; por eso los cuatro primeros son
incompatibles con él y entre sí, y no hace falta ninguna regla de orden que los combine
(["El orden manual"](familias-de-flags.md#el-orden-manual)). Lo que sí hay que fijar es **contra qué
tablero se busca la vecina: contra el de antes de la escritura**, igual que los selectores de
comentarios de más abajo. Una llamada que mueve varias tareas no se estorba a sí misma, porque el
hueco se calcula una sola vez y descontando las claves de las tareas que ella misma mueve, y dentro
de ese hueco las claves se escriben en el orden en que se escribieron las referencias
(["Varias tareas en la misma llamada"](modelo-de-datos/orden-manual.md#varias-tareas-en-la-misma-llamada)).

**`--rm-comment` y `--set-comment-date` solo pueden señalar un comentario que ya existiera al empezar
la llamada.** Como `--comment` va en el último paso, un comentario que la propia llamada añade nunca
es un objetivo válido de las otras dos: su clave se resuelve contra la lista de antes de que
`--rm-comment` actúe en el paso 3, así que referenciarla en la misma llamada es la misma clase de
fallo que una clave que no existe todavía (["Comentarios"](familias-de-flags.md#comentarios)).

## Concurrencia, atomicidad y garantías observables

Esta sección no describe un mecanismo: enuncia lo que quien llama tiene derecho a observar. Cómo se
consiga es cosa de quien implemente.

1. **Ninguna escritura se observa a medias.** Un lector concurrente ve el tablero como estaba antes
   de una escritura o como quedó después, nunca en un punto intermedio, y esto vale igual para una
   escritura de una tarea que para un lote de doscientas.
2. **Una escritura que afecta a varias tareas es todo o nada.** Si falla por cualquier motivo, ni una
   sola de las tareas implicadas queda modificada, y el código de salida lo dice: el código específico
   del fallo (3, 4, 5...) si se detectó al validar y es atribuible a un elemento concreto, 7 si se
   detectó al validar y no lo es, 8 si se detectó al escribir. En todos los casos el mensaje afirma
   explícitamente que no se ha escrito nada. Cuál gana entre el código específico y el 7 está en
   ["El código 7 garantiza que no se ha escrito nada, y el código específico siempre gana sobre
   él"](codigos-de-salida.md#el-código-7-garantiza-que-no-se-ha-escrito-nada-y-el-código-específico-siempre-gana-sobre-él).
3. **Dos procesos simultáneos nunca asignan el mismo identificador**, aunque trabajen sobre el mismo
   tablero desde copias de trabajo distintas del proyecto.
4. **Dos escrituras simultáneas sobre la misma tarea no se pierden ni se mezclan.** O se aplican una
   después de otra, o una de las dos falla con código 8.
5. **Si el programa no puede obtener el acceso exclusivo que necesita para escribir**, espera hasta
   cinco segundos y luego falla con código 8 sin escribir nada:
   ```
   error: the board is busy, another process is writing to it
   hint: retry in a moment; nothing was written
   ```
6. **Las lecturas nunca fallan por culpa de una escritura en curso**, y nunca la bloquean.

**Las seis hablan de las escrituras del tablero, y hay un solo comando que escribe ficheros de texto con
nombre fijo, `biso snapshot`.** Sus garantías son otras y están en [`biso snapshot`](cmd/snapshot.md): cada fichero se
escribe en un temporal y se renombra encima, los dos temporales se completan antes de renombrar
ninguno, y el comando no toma ningún acceso exclusivo, precisamente para que una copia no pueda hacer
fallar a la escritura de una tarea. Ninguna de las seis de aquí queda tocada por eso.

## Qué pasa con un dato que no se puede interpretar

Hay dos motivos distintos por los que un dato resulta imposible de leer, y con una base de datos son
dos casos que hay que separar: uno es que una tarea concreta esté dañada mientras el resto del
tablero sigue legible, y el otro es que no haya tablero legible en absoluto. El segundo no es una
variante del primero.

### El primer caso: una tarea ilegible

Una tarea es **ilegible** cuando el programa no puede darle a alguno de sus campos el significado que
el tablero le da: la base de datos devuelve algo corrupto para esa fila, un campo de vocabulario
cerrado guarda un valor que la configuración no declara, o una fecha guardada no es una fecha. El
resto del tablero sigue legible.

**La definición es una sola y vale igual para todos los comandos de lectura.** La tabla de la sección
siguiente decide qué tarea es ilegible, y lo decide para el tablero entero antes de mirar ningún
filtro de la llamada, no cada comando a su manera. Es la regla de
["Los vocabularios del tablero y la regla de validación"](vocabularios.md), que rige "idéntica al
escribir y al leer", aplicada a lo que ya está guardado: un valor que el tablero no declara es un error,
y una tarea que lo guarda no puede salir en un listado como si estuviera bien, cuando
`biso ls --priority urgent` sobre ese mismo tablero es un error. La razón de fondo y las alternativas
que se descartaron están en
["Una tarea ilegible es la misma para todos los comandos de lectura"](../decisiones/detalles.md#una-tarea-ilegible-es-la-misma-para-todos-los-comandos-de-lectura).

#### Qué se comprueba

Esta es la lista completa. Un campo que no está en ella no vuelve ilegible una tarea.

| Campo | La tarea es legible si | `field` del error cuando no lo es |
|---|---|---|
| `status` | su valor es exactamente uno de los de `statuses`; vacío no lo es | `status` |
| `type` | está vacío, o su valor es exactamente uno de los de `types` | `type` |
| `priority` | está vacía, o su valor es exactamente uno de los de `priorities` | `priority` |
| `due` | está vacía, o es un día real escrito `YYYY-MM-DD` | `due` |
| `createdAt` y `updatedAt` | son un instante escrito `YYYY-MM-DDTHH:MM:SSZ`; ninguna puede estar vacía | `createdAt` o `updatedAt` |
| `leaseExpiresAt` | está vacía, o es un instante | `leaseExpiresAt` |
| `question.askedAt` | si la tarea tiene pregunta, es un instante y no está vacía | `question.askedAt` |
| `createdAt` de un comentario | es un instante y no está vacía | `comment #<clave> createdAt` |
| El nombre de un campo de lista | es uno de los que el modelo conoce | el nombre guardado |
| Cualquier otra columna de la tarea, de sus criterios o de sus comentarios | su contenido es del tipo que el modelo le da: un booleano es 0 o 1, un entero es un número | el nombre del campo |

**Los campos de vocabulario cerrado son los que un filtro valida contra la configuración: `status`,
`type` y `priority`.** Son los únicos donde un valor guardado puede quedar fuera de lo que un filtro
acepta, y por eso son los únicos donde la tarea deja de ser legible.

**La comparación es exacta y no usa el algoritmo de coincidencia.** Lo que se guarda es siempre la
grafía configurada, porque cada escritura resuelve lo tecleado a ella
(["El algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia)). Un `done` guardado en
un tablero cuyo estado terminal es `Done` no es la misma cosa escrita de otro modo, es un valor que
ninguna escritura del programa produce, y como el programa compara el estado de una tarea con el
terminal letra por letra para saber si está terminada, darlo por bueno la contaría como viva.

**Una fecha vacía solo es válida donde la fecha es opcional.** `due` y `leaseExpiresAt` pueden no
tener valor. `createdAt`, `updatedAt`, la fecha de la pregunta cuando la hay y la de cada comentario
son no nulas en ["El modelo de datos de una tarea"](modelo-de-datos/index.md), así que una cadena
vacía ahí es un dato que falta y no un dato ausente. El formato de cada fecha es el de
["Números, fechas y ausencias"](contrato-json.md#números-fechas-y-ausencias), y `due` es un día y no un
instante: `2026-09-21T10:00:00Z` no es un `due` legible.

**Lo que no vuelve ilegible una tarea, y por qué.**

- Las personas (`author`, `assignees`, `leaseHolder` y el autor de cada comentario y de la pregunta) y
  las etiquetas son cerradas solo al leer y abiertas al escribir
  (["Qué valida cada filtro, y contra qué"](vocabularios.md#qué-valida-cada-filtro-y-contra-qué)): un
  filtro las valida contra las que el tablero tiene en uso, y un valor guardado pertenece por
  definición a ese conjunto. No hay ninguna lectura en la que un filtro rechace lo que un listado
  muestra, que es lo que hace ilegible a un `type` desconocido.
- Una etiqueta que la lista `labels` de la configuración prohíbe, una dependencia o un padre que no
  existen y un arrendamiento a medias son incoherencias entre datos que se leen bien. `biso doctor` las
  reporta como error (["`biso doctor`"](cmd/doctor.md#qué-comprueba)), pero la tarea se lee, se muestra
  y se filtra con su sentido intacto. Un `ordinal` fuera de rango no se comprueba al leer.

#### Qué hace cada comando

| Comando | Con una tarea ilegible |
|---|---|
| `biso ls` | La salta, la cuenta y emite el aviso de abajo por stderr. No aparece en las filas y no entra en `--count`. El resto del listado es válido y el código es 0 |
| `biso prime` | La deja fuera del recuento por estado y de todos sus bloques, y añade la línea `unreadable` al bloque `BOARD` en vez de emitir el aviso por stderr (["`biso prime`"](cmd/prime.md#la-salida-literal)). Código 0 |
| `biso get` de esa tarea | Error 3 con `code` `undecodable_task` y el motivo exacto, sin imprimir nada de la ficha, con `--section` o sin él |
| `biso get` de otra tarea | Imprime esa ficha y emite el aviso, porque resolver la referencia lee el resto del tablero |
| `biso export` y `biso snapshot` | Escriben todas las tareas legibles y ninguna de las ilegibles, emiten el aviso y salen con código 6 |
| `biso doctor` | La reporta como error con el `code` de su fila (`value_not_configured` para un valor fuera de vocabulario, `task_unreadable` para lo demás), sigue con las demás y nunca aborta. Código 6 |
| Una escritura dirigida a ella: `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer` y `archive` con una referencia que resuelve a ella | Error 3, con el motivo exacto. No se escribe nada. `set`, `start` y `finish` tienen la excepción del valor fuera de vocabulario de ["Cómo se arregla"](#cómo-se-arregla-una-tarea-ilegible); `note`, `comment`, `ask`, `answer` y `archive` no la tienen nunca, escriban lo que escriban |
| Resolver una referencia por texto, y cualquier filtro | La tarea no participa, y el aviso la nombra |

**Los comandos de lectura de conjunto no abortan nunca por una tarea mala, y no la esconden nunca en
silencio.** Las dos cosas juntas son lo que impide que un listado incompleto se confunda con un
tablero vacío. El aviso, por stderr, es
`warning: 1 task could not be read and was skipped: MYP-2` con sus identificadores, o, con más de una,
`warning: 2 tasks could not be read and were skipped: MYP-2, MYP-7`, y `biso prime` lo integra en su
propio mensaje por stdout en vez de emitirlo por stderr.

**El aviso nombra todas las tareas ilegibles del tablero, casen o no con los filtros de la llamada.**
Una tarea ilegible no se puede comparar con un filtro, así que no se puede afirmar que no lo cumple:
`biso ls --status Done` sobre un tablero con una tarea ilegible imprime `note: no tasks match` y el aviso
juntos, y nunca solo la nota, que quien lee tomaría por un hecho sobre el tablero. Cuentan también las
archivadas y las que están en el estado terminal.

**El aviso dice qué tarea y no por qué.** El motivo exacto lo da `biso get` de esa tarea, con el valor
que hay guardado, y `biso doctor` la reporta junto al resto de problemas del tablero. El error de
`biso get` es, en texto, una línea por stderr, y en JSON un sobre de error con `code` `undecodable_task`
que lleva `field` y `given` siempre, y `valid` cuando el campo es de vocabulario
(["Los errores en JSON"](contrato-json.md#los-errores-en-json)). El texto de los tres motivos que se
reconocen por su forma, con `MYP-2` como ejemplo:

| Motivo | Mensaje |
|---|---|
| Un valor de `status`, `type` o `priority` que el tablero no declara | `MYP-2 cannot be read: its priority is "urgent", which this board does not configure` |
| Una fecha que no es un día, en `due` | `MYP-2 cannot be read: due is not a calendar day (YYYY-MM-DD): "2026-9-1"` |
| Una fecha que no es un instante, en cualquier otro campo de fecha, o una fecha obligatoria vacía | `MYP-2 cannot be read: createdAt is not an instant (YYYY-MM-DDTHH:MM:SSZ): "nope"` |

En la primera fila, `priority` es el nombre del campo que falla, así que es `status` o `type` según cuál
sea. Los otros dos motivos, un nombre de campo de lista desconocido y una columna del tipo equivocado,
nombran el campo y lo guardado, sin un texto fijo. En ningún caso el mensaje lleva el texto de un error del lenguaje
o de la biblioteca de fechas, que no es parte del contrato.

#### `biso export` y `biso snapshot`

**Son las dos excepciones al código 0 de una lectura de conjunto.** Los dos escriben todo lo que han
podido leer, con el mismo aviso por stderr, pero terminan con **código 6** en vez de 0 cuando han
saltado alguna tarea: son los comandos cuyo propósito es servir de copia fiel del tablero, así que una
copia incompleta no puede parecer un éxito llano. Un guion que encadene
`biso export --out backup.ndjson && ...` o `biso snapshot && ...` puede comprobar el código de salida
para detectar un volcado incompleto.

**Lo que escriben es siempre algo que el propio programa sabe volver a leer.** Una tarea ilegible no se
copia: ni entera, ni con el campo dañado, ni con un valor por defecto en su lugar. Un
`snapshot.ndjson` con una tarea cuya prioridad el tablero no declara lo rechazaría `biso init --from`
con el error 7 de la línea inválida, y con una fecha rota no se puede ni escribir la línea. Por eso la
copia lleva las tareas legibles y solo esas, y se importa siempre entera.

**Tampoco se niegan a escribir, aunque haya una ilegible.** Es la otra mitad de la misma decisión. Las
fechas malformadas no tienen remedio dentro de `biso` (la sección siguiente), así que una negativa
pararía las copias justo mientras hay un daño que el propio programa no puede arreglar. Escribir lo
legible y salir con 6 conserva todo lo que se puede conservar y deja el fallo a la vista. A cambio, el
`snapshot.ndjson` nuevo no lleva la tarea ilegible, y la versión anterior de esa tarea solo sigue en la
historia del repositorio del tablero.

#### Cómo se arregla una tarea ilegible

El remedio depende de cuál sea el dato dañado, y `biso doctor` y `biso get` dicen cuál es.

| Qué está mal | Cómo se arregla |
|---|---|
| Un valor fuera de vocabulario en `status`, `type` o `priority` | Escribir uno que el tablero declare, con `biso set MYP-2 --priority medium` (o `--status`, o `--type`; `--clear-type` y `--clear-priority` dejan el campo sin valor, y `status` no puede quedar vacío), o declarar el valor con `biso config set` si era el correcto |
| Una fecha que no es una fecha, o una columna del tipo equivocado | No tiene remedio con ningún comando de `biso`: ninguno escribe `createdAt`, `updatedAt` ni las fechas de la pregunta y de un comentario, y una fila que no se decodifica no se puede ni cargar para editarla. Es daño externo, porque nada dentro del programa escribe una fecha mal formada: se corrige el valor en la base de datos (`board.db`, un fichero SQLite), o se recupera la tarea de una instantánea anterior |

**Una escritura dirigida a una tarea ilegible solo por su vocabulario se aplica si lo que escribe la
deja legible, y solo `biso set`, `biso start` y `biso finish` tienen esta excepción**, porque son los
únicos comandos que pueden escribir `status`, `type` o `priority`. En cualquier otro caso, incluidos
`biso note`, `biso comment`, `biso ask`, `biso answer` y `biso archive` sobre una tarea ilegible, es
error 3 sin escribir nada, aunque el campo que esos cinco escriben no tenga nada que ver con el
vocabulario: no reparan nunca, porque no son la vía que la especificación declara para hacerlo. Se
juzga la tarea como quedaría después de la escritura, no como está: `biso set MYP-2 --priority medium`
arregla la prioridad, y `biso set MYP-2 --title "Otro"` falla porque la tarea seguiría con su
prioridad fuera de vocabulario. Lo mismo vale para `biso start` y `biso finish`, que fijan el estado
por su cuenta; y la propia precondición de `start` y de `finish` (estar archivada, estar ya en el estado
terminal, no estar lista para cerrar) nunca contesta antes que la ilegibilidad de un campo que la
llamada no toca. Con `--dry-run` la vista previa contesta lo mismo que la llamada real. Es lo que
permite el remedio de arriba sin una vía especial de reparación, y no hay nada equivalente para una
fila que no se decodifica, porque esa no llega a cargarse.

### El segundo caso: la base de datos que no se puede leer

Esto **no es una tarea ilegible: es que no hay tablero legible.** Si la base de datos no abre, o abre
pero falla su comprobación de integridad, ninguna tarea es alcanzable, así que no tiene sentido
presentarlo como "una tarea se ha saltado": no hay nada que saltar, hay un tablero entero fuera de
alcance. La regla es la misma para cualquier operación, sea una lectura dirigida, una de conjunto, una
escritura o `biso doctor`: el comando aborta entero, sin escribir nada, con este mensaje por stderr:

```
error: board 3f9a2b1c's database could not be read
hint: it did not open, or it failed its integrity check, and there is no automatic repair
hint: rebuild it in place with `biso init --from <snapshot dir>`, which keeps its id
```

El código de salida es **21** (`DAMAGED`, ver ["Códigos de salida"](codigos-de-salida.md)), y no el 20 de
la ausencia de tablero, porque el remedio es otro: aquí el tablero está donde tiene que estar y lo que
hay que hacer es reconstruirlo, no crearlo. La clave `code` del sobre JSON (sección
["Los identificadores de error"](contrato-json.md#los-identificadores-de-error)) es `database_unreadable`.

**El código 21 puede salir de cualquier comando**, así que no aparece en la tabla de códigos de salida
de ninguno, salvo las dos excepciones de la sección ["Códigos de salida"](codigos-de-salida.md#los-códigos-20-21-y-22-son-los-tres-desenlaces-malos-de-resolver-el-tablero-y-cada-uno-tiene-un-remedio-distinto).

**Y el remedio se puede teclear tal cual, porque el segundo `hint` nombra el comando que lo hace.** Un
directorio cuya base de datos no abre no cuenta como tablero accesible para `biso init`, así que
`biso init --from` reconstruye ahí mismo en vez de dar el error 2 de "ya hay uno" de [`biso init`](cmd/init.md), y adopta el
`id` que nombra el marcador de la instantánea, de modo que el puntero commiteado del proyecto sigue
valiendo. Es también lo que necesita un clon recién traído a otra máquina, que llega con el directorio
del tablero versionado y sin base de datos dentro.

## Qué pasa cuando el almacén no se puede escribir

Hay un tercer motivo por el que una llamada no llega a su fin, y no es ninguno de los dos de arriba:
el tablero está donde tiene que estar y su base de datos está perfectamente sana, pero el entorno se
niega a dejar escribir en ella. El fichero o su directorio no tienen permiso de escritura, están en un
montaje de solo lectura, el disco está lleno, o el sistema operativo falla la operación. **Eso no es un
dato que no se pueda interpretar, así que no es el código 21**: es el entorno fallando y no la
petición, que es exactamente lo que nombra el código 8 (`ENVIRONMENT`, ver ["Códigos de salida"](codigos-de-salida.md)).
La clave `code` del sobre JSON es `io_error`, la misma que ya usa cualquier otra escritura de fichero
que el sistema rechaza.

```
error: /Users/avilches/.biso/boards/my-project-3f9a2b1c/board.db cannot be written: attempt to write a readonly database (8)
hint: check its permissions and the free space of its directory; nothing was written
```

**El mensaje nombra el fichero y no el tablero**, y esa elección es deliberada. El modo WAL necesita
escribir para abrirse, así que este fallo puede ocurrir al abrir, antes de que nadie haya leído el `id`
que la base de datos guarda dentro; un mensaje que nombrara el tablero saldría con un hueco donde
debería ir el identificador, mientras que la ruta del fichero es siempre conocida y es además lo único
que hay que mirar para arreglarlo.

**Vale para cualquier comando y para cualquier momento**, no solo para las escrituras declaradas: un
`biso ls` sobre un directorio de tablero sin permiso de escritura falla igual, y con el mismo código,
porque abrir en modo WAL ya es escribir. Y vale también para `biso doctor --fix`, cuya tabla de códigos
de salida ya prometía el 8 para "no se puede escribir al reparar" (["`biso doctor`"](cmd/doctor.md#códigos-de-salida)).
