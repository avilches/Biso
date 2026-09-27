# Decisiones de diseño

> **Cerradas e implementadas.** Cada decisión de esta página está implementada y verificada con pruebas
> reales; el repaso completo, decisión por decisión, está en [`pendientes.md`](pendientes.md).

Cada entrada dice primero la decisión vigente, en un párrafo que se puede citar sin más contexto. Las
alternativas que se consideraron van después, marcadas como **descartadas** y con la razón. Si una
decisión nueva deja obsoleta una anterior, la vieja se corrige aquí mismo.

## Es un proyecto independiente de `biso`

**La decisión.** `backlog.md-migrate` vive en `tools/backlog.md-migrate/` del repositorio de `biso`,
con su propio módulo Go, su propia documentación y sus propias decisiones. No importa ningún paquete de
`internal/` de `biso`, no cuelga de su documentación ni de su sitio, y habla con `biso` ejecutando su
binario. Una utilidad de migración es utillaje que se usa una vez o unas pocas, no parte del producto,
y así no toca el binario Go de `biso`, ni su presupuesto de arranque, ni el recuento de comandos de su
especificación.

**Descartado: un subcomando de `biso`** (`biso import-backlog`). Obligaba a añadirlo a la
especificación, al recuento de comandos y a la ayuda, y a meter en el binario un lector de un formato
ajeno. **Descartado: un repositorio aparte.** La prueba necesita un `biso` compilado y viviendo en el
mismo repositorio lo encuentra en `bin/biso`, sin fijar cómo se localiza entre repositorios.

## Dos órdenes independientes, una dirección cada vez

**La decisión.** `import` y `export` son dos órdenes hermanas. Cada una convierte en un sentido y se
ejecuta una vez, o unas pocas mientras los dos tableros conviven. Ninguna mira el estado de la otra.

**Descartado: un sincronizador continuo.** Exige decidir qué gana cuando cambian los dos lados, y ese
es justo el tipo de problema que el proyecto no quiere tener. El nombre `backlog.md-migrate` lo dice.

## El milestone es una etiqueta con ámbito `milestone::<slug>`

**La decisión.** `biso` no tiene milestone en su modelo de datos, y en Backlog.md es un dato que se
guarda en cada tarea que pertenece a uno. Cada tarea recibe una etiqueta con ámbito `milestone::<slug>`,
con el slug del título del milestone. El separador es el doble `::` y no el simple `:`, porque una tarea
de Backlog.md pertenece como mucho a un milestone, y `::` es justo la forma con la que una etiqueta con
ámbito de `biso` deja como mucho un valor de esa clave por tarea, aplicado y comprobado por el propio
`biso` al escribir, no por convención del convertidor
(["Las etiquetas con ámbito"](../../../docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito)). No
añade ningún campo al modelo, se filtra con `--label` (con `milestone:` o con `milestone::`, indistinto
al consultar) y se ve en `biso ls`.

**Descartado: la etiqueta plana `milestone:<slug>` con separador simple.** Es lo que se propuso antes de
que existieran las etiquetas con ámbito (TASK-85). Admite varios valores de la misma clave en la misma
tarea, así que nada impide que una tarea termine con `milestone::a` y `milestone::b` a la vez si el
convertidor tiene un error o si alguien la edita a mano después; `biso` no lo habría rechazado. El doble
`::` traslada esa garantía al propio tablero de destino. **Descartado: un campo de extensión (`ext`).**
Además de no estar claro que se pudiera filtrar como una etiqueta, el campo ya no existe en `biso`: se
retiró (TASK-83). **Descartado: asumir la pérdida.** Se perdería la agrupación por hitos de los tableros
que la usan.

**El campo `project` de Backlog.md sigue la misma regla**, con la etiqueta con ámbito `project::<slug>`.
Es otro dato de agrupación que `biso` no modela, que Backlog.md guarda como un texto en cada tarea y del
que también hay como mucho uno por tarea, y darle un trato distinto al del milestone no tendría
justificación.

## Los identificadores conservan su número y cambian de prefijo

**La decisión.** El tablero destino se crea antes, con el prefijo que se quiera, y el convertidor lo
lee de `biso`. Cada id de origen conserva su número con el prefijo del destino. Solo los ids cuyo número
ya existe en el destino cambian de número, con el siguiente libre por encima del mayor que haya entre
origen y destino, y las menciones de esos ids se reescriben en todo el texto, títulos incluidos. Ninguna
tarea que no choque se desplaza. Así el destino puede tener el prefijo que quiera y no tiene por qué
estar vacío. **Una excepción: una tarea que ya está en el destino no se vuelve a importar.** Si el
destino tiene una tarea con el mismo título y la misma fecha de creación, sea cual sea su id, es la
misma tarea de una importación anterior y se salta con un hallazgo. Sin esto, ejecutar `import` dos
veces duplicaría el tablero: `biso` rechazaría el lote por ids ocupados, pero el convertidor lo evitaría
reasignando todos los ids.

**Las subtareas siempre reciben un número nuevo.** Backlog.md da a una subtarea un id con punto
(`XYZ-001.01`, `ABC-1.2`), y un id de `biso` es siempre `<PREFIJO>-<n>`. No hay forma de conservarlo, así
que se le asigna un número libre y se reescriben sus menciones. El parentesco no se pierde: va en
`parent`. Por eso la regla de la tarea ya importada no compara los ids, sino el título y la fecha. Esto
ya no es una regla provisional a la espera de TASK-73: TASK-73 se resolvió confirmándola, `biso` no va a
adoptar la forma con punto, y lo único que queda de esa tarea es la decisión de este proyecto para
conservar el id de origen, en la siguiente sección.

**La coincidencia de menciones distingue mayúsculas y no toca lo que va pegado a otras palabras.** En el
tablero medido hay cinco menciones en minúsculas que son nombres de ramas y de worktrees
(`task-10-modelo`), y reescribirlas sin distinguir mayúsculas las habría corrompido.

**Descartado: obligar a que el destino tenga el mismo prefijo que el origen** (`--prefix TASK`).
Dejaba un tablero con un prefijo que no elige nadie y que no se puede cambiar después. **Descartado:
renumerar todo desde el siguiente libre.** Cambia todos los números aunque no choque nada y rompe el
parecido entre los dos tableros. **Descartado: dejar que `biso` asigne los ids.** Un lote no puede
expresar un padre o una dependencia sobre un id que todavía no existe.

**Backlog.md reutiliza el número de una tarea archivada, y el import lo tolera en vez de rechazarlo.**
Con el CLI real (1.53.0): crear `TASK-1` y `TASK-2`, archivar `TASK-2`, y crear una tercera tarea le da
a esta tercera el id `TASK-2` otra vez, el mismo número que la archivada y no uno nuevo; el fichero de
`archive/tasks/` sigue con `id: TASK-2` en su frontmatter, y el nuevo fichero de `tasks/` también. El
ciclo se puede repetir sin que en ningún momento haya dos copias activas a la vez: archivar también esa
reutilización deja dos ficheros en `archive/tasks/` con `id: TASK-2` y ninguna copia activa ni terminada,
una secuencia perfectamente normal de un tablero real que lleva tiempo en marcha, no un tablero editado a
mano. Tratar "dos ficheros traen el mismo id" como fatal sin excepción rompería este uso normal: la
primera vez que Backlog.md reutiliza un número, `import` rechazaría el tablero entero. Por eso el único
caso fatal pasa a ser que MÁS DE UNA copia no archivada comparta el id: es lo único que Backlog.md nunca
puede producir por sí mismo, y sigue siendo el error fatal de código 3. Con cero o una copia no
archivada, nunca es fatal, sea cual sea la forma del id compartido. Cuando hay exactamente una copia no
archivada, esa es la que conserva el número si el id es simple, mientras que cada una de las archivadas se
reasigna con el motivo de reutilización; si el id compartido tiene forma de subtarea, la copia no
archivada se reasigna igual que las demás, porque ninguna subtarea conserva su número nunca. Cuando
ninguna copia está activa ni terminada, ninguna conserva el número: todas se reasignan, sin que haga falta
ningún desempate entre las archivadas para decidir cuál es "la buena", el mismo tratamiento que ya reciben
todas las subtareas.

**Cualquier referencia a un id compartido resuelve a una única tarea, y deja un hallazgo.** Una mención
en texto libre, un `parent` o un elemento de `dependencies` que nombra un id que dos o más ficheros
comparten no puede distinguir a cuál de las tareas se refería quien lo escribió: Backlog.md no guarda esa
intención en ningún sitio, solo el número. Si exactamente una de ellas no está archivada, la resolución
elige esa: es la misma que conserva el número cuando el id compartido es simple, la que de verdad existe
hoy en el tablero, mientras que la archivada, o las archivadas si hay varias, ya cerraron su ciclo y no
son a las que tendría sentido seguir apuntando. Si NINGUNA de las tareas que comparten el id está activa
ni terminada, no hay ninguna "existente hoy" a la que apuntar: si al menos una de esas copias archivadas
tiene una `created_date` válida, la resolución elige la de fecha más reciente entre ellas y, si dos o más
empatan también en la fecha exacta, gana la última por ruta relativa ascendente entre las que empatan, el
mismo mecanismo que usa la regla 4 de "Identificadores" para su propio desempate; si ninguna copia tiene
una fecha válida, elige la última por ruta relativa ascendente. En cualquiera de los dos casos es la
copia más cercana de todas ellas a ser "la actual". Es una simplificación deliberada en los dos casos,
no una detección: el convertidor no intenta
adivinar cuál de las tareas tenía en mente quien escribió la referencia, solo resuelve de la única forma
que tiene sentido, tanto si el id compartido es simple como si tiene forma de subtarea. Esta resolución
deja un hallazgo, no bloqueante, con el fichero, el id compartido y el id final al que se resolvió:
["Nada se pierde en silencio"](#nada-se-pierde-en-silencio), más abajo, incluye este caso, porque una
referencia ambigua resuelta sin avisar sería justo el tipo de pérdida silenciosa que ese principio
prohíbe, aunque aquí lo que se pierde no es un dato sino la certeza de a qué tarea apuntaba de verdad
quien escribió la referencia.

**Descartado: dejar la referencia sin resolver**, tratándola como la mención no reconocida de la regla 6
de "Identificadores". Perdería la referencia por completo en el caso más común, precisamente aquel en el
que la interpretación "la tarea que existe hoy en el tablero" sí es la correcta: la inmensa mayoría de
las referencias a un id reutilizado, cuando la tarea que lo tenía ya está archivada, quieren decir la
tarea activa, no un limbo sin resolver.

**Descartado: resolver por la fecha de la referencia**, con un argumento honesto y no uno débil: la cota
sí resuelve sin ambigüedad el caso más común, una tarea activa y una sola archivada compartiendo el id.
Si la fecha de creación o de actualización de la tarea que menciona el id es anterior a la fecha de
creación de la tarea activa, esa mención solo puede referirse a la archivada, y con una sola archivada eso
ya deja una única candidata sin ambigüedad. Aun así se descarta, porque añadir esta heurística complica
la implementación con sus propios casos límite: qué pasa si hay más de una archivada y la cota descarta
solo a algunas de ellas, qué fecha del que menciona se usa si le falta tanto `updated_date` como
`created_date`, o qué pasa en un empate exacto de fechas. El beneficio que ganaría, precisión en un caso
que hoy queda marcado como ambiguo, ya lo cubre de otra forma el hallazgo de resolución: al quedar visible
y no bloqueante, quien le importe el caso concreto puede revisarlo y corregirlo a mano. La simplicidad de
"siempre la no archivada, o la más reciente si ninguna lo es" es preferible a una regla más precisa pero
más frágil.

**La comparación de "ya está en el destino" usa el título ya reescrito, no el crudo.** El título de una
tarea de origen puede mencionar el id de otra tarea (`"Follow-up of TASK-1"`), y la regla 5 de
"Identificadores" lo reescribe en el título que de verdad se escribe en el destino (`"Follow-up of
BISO-1"`). Si la comparación de la excepción anterior usara el título crudo de origen, la segunda
ejecución de `import` compararía `"Follow-up of TASK-1"` contra el `"Follow-up of BISO-1"` que ya está
guardado, no encontraría la coincidencia y duplicaría la tarea, justo lo que esta misma decisión promete
evitar. Por eso esa comparación usa una reescritura de menciones más simple que la definitiva de la
regla 5, y que no depende de si un id choca o no en esta ejecución concreta (evita la circularidad de
necesitar saber qué tareas se saltan antes de poder decidir qué tareas se saltan): sustituye por el
prefijo del destino y el mismo número mencionado solo la mención (misma forma y límites de palabra que
la regla 5) cuyo id corresponde a AL MENOS una tarea SIMPLE del lote de origen actual (puede corresponder
a más de una si el número se reutilizó tras archivar, regla 1 de "Identificadores"; basta con que al
menos una la tenga), sin tener en cuenta si ese número choca con algo en el destino ni si esa tarea se va
a reasignar; cualquier otra mención con forma de id, sea porque no corresponde a ninguna tarea del lote o porque corresponde a una
subtarea, se deja tal cual, con el mismo criterio con el que la regla 6 de "Identificadores" deja tal
cual una mención que no reconoce. Esta reescritura "ingenua" (naive en el código) es estable entre
ejecuciones: la primera y la segunda vez que se lee el mismo origen dan el mismo título para comparar,
sin importar el estado del destino en ese momento.

**Una limitación que se acepta.** Que una mención no corresponda a ninguna tarea del lote de origen no
es un problema: tanto la reescritura ingenua de una segunda ejecución como la reescritura definitiva de
la primera dejan esa mención sin tocar, así que el título candidato y el título ya guardado coinciden en
esa parte y la comparación funciona sin necesidad de aceptar nada aquí. La limitación real cubre tres
casos. El primero: si la tarea SIMPLE mencionada en el título fue ella misma reasignada a un número
distinto en la primera importación, porque su número chocaba en aquel momento, la reescritura ingenua de
la segunda ejecución no reproduce ese número reasignado, y la comparación vuelve a fallar. El segundo,
más amplio: cualquier mención de una SUBTAREA, choque o no, porque una subtarea siempre recibe un número
nuevo (regla 4 de "Identificadores"), y la reescritura ingenua no puede predecir ese número sin conocer
ya el resultado de la reasignación, la misma circularidad que motiva toda esta reescritura; así que un
título que menciona una subtarea se duplicará en cada reimportación, no solo cuando esa subtarea choca.
El tercero: si el id SIMPLE mencionado en el título lo comparten varias tareas y NINGUNA está activa ni
terminada (todas archivadas, regla 1 de "Identificadores"), la copia a la que resuelve esa mención en la
reescritura definitiva de la primera importación (la de creación más reciente, o la última por ruta si
ninguna tiene fecha, según la decisión de más arriba) recibe siempre un número nuevo, porque ninguna
copia archivada conserva el suyo cuando todas están archivadas; la reescritura ingenua de la segunda
ejecución, que solo comprueba si el número está presente entre los ids simples del lote sin conocer
reasignaciones, sigue usando el número original sin reasignar, y la comparación vuelve a fallar igual que
en los otros dos casos. En los tres casos la tarea se duplicaría. La identidad por título es frágil por
construcción, y esta decisión ya acepta esa fragilidad para una tarea cuyo título se editó a mano en el
destino después de importarla ("ya no se reconoce"); la reasignación por colisión, la mención de una
subtarea y la de un id compartido sin ninguna copia activa son manifestaciones de esa misma fragilidad,
no de una fragilidad de una naturaleza distinta. No compensa resolverla con más mecanismo, como guardar
aparte un identificador estable: sería una complejidad nueva para casos que siguen siendo específicos
(un título que menciona a una subtarea, a un id compartido sin ninguna copia activa, o a una tarea simple
que además choca en la primera importación) frente al caso hermano, ya aceptado, del título editado a
mano.

**El mismo número, no la misma cadena, también decide cuándo dos ids de origen chocan entre sí, pero
solo entre ids de la misma forma.** La regla 1 de "Identificadores" usa esta misma forma canónica
(prefijo sin distinguir mayúsculas, número entero, número de subtarea entero si lo hay) con la que esta
decisión ya compara números para saber si uno choca con el destino: dos ids simples, o dos ids de
subtarea, con el mismo prefijo y el mismo número principal (y, si son de subtarea, el mismo número de
subtarea) son el mismo id de origen aunque las cadenas sean distintas, como `TASK-1` y `TASK-001`. Un id
simple y un id de subtarea nunca son el mismo id aunque compartan el número principal: `TASK-1` y
`TASK-1.2` son dos ids distintos, la tarea y una de sus subtareas, justo el caso que la regla 4 y la
sección anterior dan por hecho que conviven sin ser un error.

**Corrección:** esta entrada decía que dos ficheros que traen el mismo id por esta forma canónica
siempre paran el import con el código 3, para no convertirse en silencio en la misma tarea de destino y
perder una tarea entera sin avisar. Ya no es del todo cierto: la excepción de "Backlog.md reutiliza el
número de una tarea archivada, y el import lo tolera en vez de rechazarlo", más arriba en esta misma
sección, hace que el código 3 solo se dispare cuando MÁS DE UNA de las copias que comparten el id no está
archivada; con cero o una copia no archivada, el lote nunca se aborta por este motivo, sea cual sea el
número de copias archivadas que compartan el id.

**Esa misma forma canónica también decide a qué tarea corresponde una mención en el texto.** La
reescritura definitiva de menciones (regla 5 de "Identificadores"), el hallazgo de una mención que no
corresponde a ninguna tarea (regla 6) y la reescritura ingenua de la comparación de "ya está en el
destino" (regla 3) buscan los tres en la tabla de equivalencias por el mismo número entero, no por la
cadena exacta: es una única forma canónica la que decide, en cualquier punto de la especificación donde
hace falta, a qué tarea de origen corresponde un id. Sin esto, una tarea de origen con ceros en su id
(`TASK-001`) y un título o una descripción que la mencionan sin ceros (`TASK-1`) quedarían sin conectar:
la mención no se reescribiría en la
primera importación, la comparación de "ya está en el destino" (que sí usa forma canónica) no
reconocería lo que quedó escrito en la segunda, y la tarea se duplicaría. El patrón que decide si algo
cuenta como una mención en absoluto no cambia: sigue distinguiendo mayúsculas y los mismos límites de
palabra; la forma canónica solo entra después, para decidir a qué tarea del lote corresponde una mención
ya reconocida como válida.

**`parent` y `dependencies` se resuelven por la forma canónica entera, prefijo insensible a mayúsculas
incluido, porque no son texto libre.** El patrón que reconoce una mención en la regla 5 distingue
mayúsculas por una razón concreta: protege texto libre donde un id puede aparecer por casualidad, como
un nombre de rama (`task-10-modelo`) que no es una mención real. Un valor de `parent_task_id` o de
`dependencies` no corre ese riesgo, porque el campo entero es un id y nada más; no hay nada que proteger
de una coincidencia accidental. Por eso ahí se aplica la forma canónica completa de la regla 1 sin la
salvedad de mayúsculas: un `parent_task_id: TASK-1` que apunta a la tarea de origen `TASK-001` se
reconoce igual que si apuntara con el prefijo en minúsculas. Cuando dos elementos de `dependencies` de
la misma tarea resuelven al mismo id final por esta forma canónica, se deja uno solo, el primero en el
orden original, el mismo criterio de deduplicación que ya usa "Alfabeto de un token" para dos valores de
`labels` o `assignees` que quedan iguales tras la conversión.

## El identificador de origen de una subtarea se guarda en una etiqueta con ámbito

**La decisión.** `biso` decidió no adoptar la forma `<PREFIJO>-<n>.<m>` de Backlog.md, y dejó
explícitamente en manos de este proyecto decidir cómo conservar ese identificador de origen para el
viaje de ida y vuelta
(["No se adoptan identificadores de subtarea con punto"](../../../docs/decisiones/detalles.md#no-se-adoptan-identificadores-de-subtarea-con-punto)).
La respuesta es una etiqueta con ámbito exclusiva, `backlog.id::<id-de-origen>`, con el id completo tal
como lo escribe Backlog.md, punto incluido (`backlog.id::TASK-56.1`), añadida a la tarea junto con las
demás etiquetas de la importación. El separador es `::` porque una tarea tiene como mucho un id de
origen. El alfabeto de una etiqueta con ámbito admite letras, dígitos y los símbolos `- _ . : @`
(["El juego de caracteres de un token"](../../../docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token)),
y un id de Backlog.md solo usa letras, dígitos, el guion del prefijo y el punto de la subtarea, así que
el valor cabe entero sin ninguna conversión. La clave `backlog.id` no choca con `milestone` ni con
`project`, las otras dos claves con ámbito que ya usa este convertidor.

Con esta etiqueta, `export` (TASK-7) puede reconstruir el id exacto de Backlog.md de una tarea que fue
una subtarea al importarla, en vez de inventar uno nuevo: lee `backlog.id::` si existe y, si no, genera
un id propio de Backlog.md a partir del `parent` y de una numeración nueva. La etiqueta se añade
**solo** a las tareas cuyo id de origen llevaba punto; una tarea con un id simple (`TASK-70`) no la
necesita, porque su id de destino ya se reconstruye con la regla de "Los identificadores conservan su
número y cambian de prefijo": basta con volver a poner el prefijo de origen delante del mismo número,
salvo que ese id fuera de los reasignados por colisión o por id reutilizado tras archivar (regla 1 de
"Identificadores"), casos en los que de todos modos no hay un número de origen que reconstruir con
sentido, porque el conflicto, o la reutilización, ya dicen que ese número no era único en el tablero de
origen en el momento de exportar.

**Descartado: guardar el id de origen para toda tarea, no solo las subtareas.** Sería una etiqueta más
por tarea sin ninguna ganancia: un id simple ya se reconstruye con la regla de prefijo y número, y
guardarlo de todos modos duplicaría un dato que no hace falta leer nunca.

**Descartado: un campo de extensión (`ext`).** Es la solución obvia para un identificador corto asociado
a una clave, y es justo el ejemplo que pone la propia especificación de `biso` para lo que sustituye a
`ext`. Pero `ext` no existe en `biso`: se retiró (TASK-83), así que ni siquiera es una alternativa
disponible hoy.

**Descartado: guardarlo en `references`**, como un puntero de texto libre (`backlog:TASK-56.1`, por
ejemplo). `references` no tiene vocabulario ni estructura: nada impide que se acumulen varias entradas
parecidas, no se puede filtrar por él con una sintaxis dedicada como `--label backlog.id:`, y ya va a
llevar los punteros reales de la tarea (ver "Documentación y ficheros tocados" en la especificación), así
que mezclar ahí un dato de contabilidad interna del convertidor haría más difícil distinguir un puntero
de verdad de una anotación de importación.

**Descartado: guardar solo el sufijo de la subtarea (`.1`) en vez del id completo.** El sufijo por sí
solo no sirve para nada sin saber a qué padre pertenece y con qué prefijo se escribió en origen; el id
completo es autocontenido y se puede leer sin cruzarlo con `parent`.

## Una etiqueta con ámbito derivada del dato real gana a la etiqueta de origen que choca con ella

**La decisión.** Las tres etiquetas con ámbito que añade el convertidor, `milestone::<slug>`,
`project::<slug>` y `backlog.id::<id-de-origen>`, se derivan siempre de un dato real de la tarea de
origen: su `milestone`, su `project`, o su id con punto. Si esa misma tarea ya trae en su propia lista de
`labels` un valor con esa misma clave (por ejemplo, un equipo que ya usaba `milestone::sprint-3` como
convención propia antes de migrar), escribir las dos etiquetas de la misma clave con ámbito haría que
`biso` rechazara el lote entero, porque una clave con `::` deja como mucho un valor por tarea
(["Escribir una etiqueta con
ámbito"](../../../docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito)):
`error: MYP-11 already has "size::s", and :: allows at most one value of the key "size"`. Gana la
etiqueta que deriva el convertidor del dato real, no la que ya traía el origen: es el dato que
efectivamente vive en la tarea de Backlog.md (su milestone real, su proyecto real, su id de origen real),
mientras que la etiqueta de `labels` con esa misma clave es, en el peor de los casos, una convención local
del tablero de origen que no tiene por qué significar lo mismo en el destino. La etiqueta de origen que
choca se quita de `labels` antes de escribir el lote, y se informa con el mismo tratamiento que un
carácter fuera del alfabeto de un token: una línea de hallazgo con el fichero, la tarea y el valor
descartado. La tarea no se salta: perder una etiqueta plana es una pérdida menor y reversible con un
aviso, mientras que saltar la tarea entera perdería mucho más (el resto de sus campos, sus criterios, sus
comentarios) por una sola etiqueta que ya iba a quedar duplicada.

**Descartado: dejar que gane la etiqueta que ya traía el origen.** Invertiría el propósito de la etiqueta
sintética: `milestone::`, `project::` y `backlog.id::` existen para que `biso` y `export` puedan confiar
en que esa clave contiene el dato real de Backlog.md, no una convención de texto libre que alguien haya
usado antes con la misma forma por casualidad. **Descartado: saltar la tarea entera cuando hay
colisión.** Convertiría un choque de una sola etiqueta en la pérdida de toda la tarea, un coste
desproporcionado para un caso que se resuelve quitando un valor y avisando.

## Las fechas se leen como UTC

**La decisión.** Backlog.md 1.52.0 guarda las fechas como `YYYY-MM-DD HH:mm` en UTC, y `biso` usa
instantes UTC con precisión de segundo. La conversión añade `:00` y la `Z`. Se comprobó contra un
reloj independiente: el commit automático que hizo Backlog.md al editar una tarea lleva la hora
`17:30:22-04:00`, y el `updated_date` de esa misma tarea dice `21:30`, es decir, UTC y no la hora local
de esa máquina. Sin `updated_date`, `updatedAt` toma la
`createdAt`, porque `biso` pondría de otro modo el instante de la importación y la tarea parecería
recién modificada.

**Descartado: un flag de zona horaria.** Hasta que aparezca una versión de Backlog.md que no guarde en
UTC, sería configuración para un caso que no existe.

## El orden manual se recalcula, no se copia

**La decisión.** El `ordinal` de Backlog.md es un número; el `ordinal` de `biso` es una clave de texto en
base 36 que no puede terminar en `0` y que no guarda ninguna relación aritmética con un número
(["El orden manual y su clave"](../../../docs/spec/modelo-de-datos/orden-manual.md)). No se copia tal
cual, ni siquiera convertido a su grafía en base 36: un `ordinal` como `1000` terminaría en `0`, que es
justo la forma que `biso` rechaza al escribir. El convertidor lee también el `ordinal` de las tareas que
ya existen en el destino y toma la mayor de esas claves como ancla; después ordena las tareas de origen
por su `ordinal` ascendente y les asigna claves nuevas que conservan ese orden relativo, con el mismo
algoritmo del punto medio que usan `--ordinal last`, `--above` y `--below` de `biso`, como si el lote
entero se colocara con `--ordinal last` detrás de esa ancla, una tarea tras otra. Anclarlas después de la
última clave del destino, en vez de calcularlas como si el destino estuviera vacío, es lo que hace que
ninguna clave nueva choque nunca con una que ya exista: cada clave calculada es estrictamente mayor que
el ancla y que la clave que se acaba de asignar antes que ella, así que cae siempre en un hueco donde no
hay ninguna tarea del destino, sea cual sea el estado del destino. Una tarea sin `ordinal` en origen no
recibe ninguna clave, igual que en `biso` una tarea sin orden manual no ocupa ningún sitio de él.

**El `ordinal` es un único contador global, no uno por columna de estado.** Medido contra los ficheros
reales de `backlog/tasks/*.md` del tablero de este mismo proyecto (111 tareas), no con el CLI de
Backlog.md sobre un tablero de prueba controlado como el resto de lo medido en esta página (ver "El
formato de Backlog.md se mide con su CLI, no con un tablero"): es una muestra de un solo tablero, no una
garantía verificada con el CLI. Prueba: `TASK-69` tiene `status: Done` y `ordinal: 200`; `TASK-70` tiene
`status: To Do` y `ordinal: 300`; `TASK-68` tiene `status: To Do` y `ordinal: 400`; `TASK-71` tiene
`status: Done` y `ordinal: 600`. Tareas de estados distintos se entremezclan en la misma secuencia
ascendente, lo que descarta un contador separado por columna. El mismo barrido encontró que el `ordinal`
no es necesariamente único: `TASK-3` y `TASK-8` comparten `ordinal: 3000`, el único par que lo hace entre
las 111 tareas del tablero. Cuando dos tareas de origen comparten `ordinal`, el convertidor las ordena
entre sí por su id de origen ascendente, con el mismo orden natural que usa "Identificadores" para
asignar número nuevo a los ids que chocan, antes de asignarles las claves nuevas de orden manual; así dos
ejecuciones sobre el mismo origen dan siempre el mismo resultado.

**Descartado: convertir el número a su grafía en base 36 y usarla tal cual.** Falla por partida doble:
puede terminar en `0` (`1000` en base 36 es `rs`, que no termina en cero por casualidad, pero `36000` sí
lo haría) y, sobre todo, no preserva el orden relativo entre dos claves de longitudes distintas de la
forma en que lo exige `biso`: la comparación es por punto de código, no numérica, así que una conversión
de base ingenua no basta y hay que pasar por el algoritmo del punto medio de todos modos.

**Descartado: no importar el orden manual.** Se perdería un dato real: la propia medición de `biso`
sobre seis tableros de Backlog.md encontró reordenamientos manuales de verdad
(["El orden manual es una clave de texto"](../../../docs/decisiones/detalles.md#el-orden-manual-es-una-clave-de-texto)),
así que no es un campo que nadie use.

## Estados, tipos y prioridades son los del destino

**La decisión.** La utilidad no impone ningún vocabulario. El destino se crea antes con el suyo, y el
convertidor casa cada valor de origen con uno de los declarados, con la misma regla de coincidencia que
usa `biso` (plegar mayúsculas, quitar diacríticos, espacios, guiones y guiones bajos). Lo que no case se
omite de la línea y se informa, en lugar de corregirlo por su cuenta.

**Descartado: declarar en el destino los cinco estados de Backlog.md por defecto.** Sirve para el
tablero de este proyecto, pero una utilidad general no puede suponer los estados de un tablero ajeno.

## Los espacios de una etiqueta o un asignado se convierten en guiones

**La decisión.** Backlog.md admite espacios dentro de una etiqueta (`with space`) y de un asignado
(`Sara Smith`), y `biso` no los admite en ninguno de los dos. Al importar, el convertidor transforma
cada tramo de espacios en un guion normal (`with-space`, `Sara-Smith`), quita los del principio y del
final, y avisa de cada conversión. Los demás caracteres que `biso` no admite (`a/b`, `c!`) siguen
quitando el valor, con aviso. La conversión no se deshace al exportar: `with-space` se queda así.

**La pega.** Exportar un tablero importado no devuelve el texto original de esas etiquetas, y dos
etiquetas que solo se diferencian por un espacio o un guion (`a b` y `a-b`) se funden en una.

**Descartado: hacer que `biso` admita espacios en una etiqueta** (era TASK-72, ya descartada). Cambiaba el
modelo de `biso` y reabría su decisión sobre el alfabeto de un token solo para servir a una utilidad de
migración. **Descartado: quitar la etiqueta con aviso.** Perdía un dato que se puede conservar con una
transformación evidente.

**Se aplica también a los asignados.** El encargo hablaba de las etiquetas, pero los asignados tienen el
mismo alfabeto cerrado que las etiquetas en `biso`, así que se les aplica exactamente la misma
conversión de espacios a guion, sin ningún trato distinto.

## `documentation`, `references` y `modified_files` van tal cual, y `documentation`/`modified_files` se dejan fundir en `references`

**La decisión.** `biso` no tiene campo `documentation` ni campo `modifiedFiles`: los dos se retiraron
(TASK-83 retiró `ext`, y por separado se retiraron `documentation` y `modifiedFiles`) y `references`
quedó como el único campo de punteros de una tarea. Es texto libre, sin restricción de caracteres, así
que las listas de Backlog.md pasan sin comprobación ni cambio. El alfabeto cerrado de `biso` (letras,
dígitos y `- _ . : @`) rige solo para `labels` y `assignees` (ya no para las claves de `ext`, que no
existen), y solo esos dos campos se validan. Un valor inválido se quita de la lista y se informa, para
que un lote no falle entero por una etiqueta.

`documentation` y `modified_files` de Backlog.md se escriben, tal cual, en las claves `documentation` y
`modifiedFiles` del lote: son las claves de compatibilidad que `biso new --from` sigue aceptando de un
lote ajeno, y que funde en `references` por su cuenta, avisando de ello. El convertidor no las funde él
mismo. El detalle y la pega que esto acepta (la pérdida de la distinción entre las tres listas de origen
al volver a exportar) están en la especificación, sección ["Documentación y ficheros
tocados"](especificacion.md#documentación-y-ficheros-tocados).

## La definición de hecho se marca con el sufijo `#dod`

**La decisión.** `biso` no tiene definición de hecho: al importar, la convierte en criterios de
aceptación y ya no se distinguen. Para que `export` pueda devolverlos a su sección, el convertidor los
añade como criterios con el texto seguido de un espacio y `#dod`. En la exportación, un criterio cuyo
texto acaba en ` #dod` va a la definición de hecho, sin el sufijo, y uno que no lo lleva se queda como
criterio de aceptación. Es una convención de texto en la frontera con otra herramienta, no un campo
nuevo, y no contradice la decisión de `biso` de tener una sola lista de comprobación.

**La pega.** Un criterio que acabe por casualidad en ` #dod` se exportará como definición de hecho.
Se acepta porque es un sufijo muy poco probable en un criterio real.

**Descartado: dejar que `biso` convierta `definitionOfDone` por su cuenta.** Es lo que ya hace, pero no
deja huella para volver, y el objetivo es que exportar e importar de nuevo conserve todo.
**Descartado: perder la distinción.** Que los tableros con los que se midió no usen la definición de
hecho no dice nada de otros tableros, y la utilidad es general.

## El formato de Backlog.md se mide con su CLI, no con un tablero

**La decisión.** Lo que la especificación dice de cómo guarda Backlog.md sus campos, sus secciones y sus
ids se ha comprobado creando un proyecto de prueba con el propio CLI de Backlog.md 1.52.0 y usando todas
las opciones de `backlog task create`, `edit`, `archive` y `complete`, y leyendo los ficheros que
resultan. Los ficheros de prueba de la suite se generan igual, con el CLI, y no se escriben a mano. Es
lo que ha descubierto cosas que ningún tablero real habría enseñado: las subtareas con id con punto, el
campo `project`, los comentarios, la definición de hecho por defecto que se copia a cada tarea, la
configuración que puede vivir fuera de la carpeta de datos y el prefijo cuya capitalización de la
configuración no coincide con la de los ids.

**Descartado: deducir el formato del tablero de este proyecto.** Es una muestra de una sola
configuración y de las opciones que alguien usó, y la herramienta es general.

## Nada se pierde en silencio

**La decisión.** Todo lo que el convertidor no puede mapear, cambia o quita, sale como un hallazgo por
la salida de errores, con el fichero y el campo. Incluye una clave del frontmatter o una sección del
cuerpo que no reconoce. El código de salida 5 avisa de que hubo hallazgos y `--strict` permite negarse
a escribir nada si los hay. Quien ejecuta decide. También resolver una referencia ambigua a un id
compartido (["Los identificadores conservan su número y cambian de
prefijo"](#los-identificadores-conservan-su-número-y-cambian-de-prefijo)), a la tarea no archivada o, si
ninguna lo está, a la seleccionada entre las que lo comparten, sigue esta misma regla: es un hallazgo, no
un fallo silencioso, con el fichero, el id compartido y el id final al que se resolvió.

**Un fichero de tarea cuyo frontmatter no se puede interpretar como YAML válido sigue la misma regla:
es un hallazgo, no un fichero que aborta el lote.** Medido contra el CLI real de Backlog.md 1.53.0: un
frontmatter con una comilla sin cerrar (`title: "Tarea rota`) o con la clave `status` duplicada hace
que el propio Backlog.md descarte esa tarea en silencio, sin ningún aviso en `stdout` ni en `stderr`, y
ni siquiera `backlog task view TASK-3` la encuentra (responde `Task TASK-3 not found`, el mismo mensaje
que si el fichero no existiera). El convertidor, en cambio, lee `tasks/`, `completed/` y
`archive/tasks/` directamente con su propio analizador de YAML, así que sí se va a encontrar tableros
reales con un fichero así, que el CLI de Backlog.md nunca marcó como roto. Esa tarea se salta, se
informa con el fichero y el error de parseo, y el resto del lote sigue adelante; con `--strict`, este
hallazgo cuenta igual que cualquier otro y no se escribe nada.

**Descartado: abortar en el primer hallazgo.** Un solo estado desconocido impediría ver el resto de
problemas del tablero de una sola vez.

**Descartado: abortar el import entero por un fichero con el frontmatter roto.** Es el mismo argumento
que abortar en el primer hallazgo, aplicado a un caso concreto: un solo fichero mal escrito a mano no
puede impedir ver el resto de un tablero real, y menos cuando ni siquiera el propio Backlog.md se dio
cuenta de que estaba roto.

## La exportación: mejor esfuerzo, pero reversible

**La decisión (sin implementar; su diseño y su prueba son TASK-7).** `export` escribe un directorio
`backlog/` que Backlog.md abre. Los campos de `biso` que Backlog.md no tiene (comentarios, pregunta
abierta y arrendamiento) se escriben como una sección de Markdown delimitada y legible por máquina,
dentro de la descripción o de las notas, y `import` la reconstruye como campos: exportar e importar de
nuevo conserva todo. `ext` y `modifiedFiles` no están en esta lista porque no son campos de `biso`: se
retiraron los dos, y no hay nada que reconstruir para ellos. Las reglas de arriba se aplican al revés:
el prefijo de `biso` pasa al del Backlog.md de destino con las mismas reglas de colisión, la etiqueta
`milestone::<slug>` vuelve a ser un milestone y el exportador crea el fichero de milestone que falte,
` #dod` vuelve a la definición de hecho, las fechas se escriben al minuto en UTC y la clave de orden
manual se traduce de vuelta a un `ordinal` numérico que conserva el mismo orden relativo (el mismo
problema que "Orden manual" en la especificación, en el sentido contrario). `references` vuelve entero
al campo `references` de Backlog.md: no se reparte entre `references`, `documentation` y
`modified_files`, por la pérdida ya aceptada en la sección enlazada arriba.

**Riesgo abierto.** Backlog.md pierde las claves de frontmatter que no conoce cuando edita una tarea.
Hay que medir si conserva también una sección desconocida del cuerpo.

**Descartado: reutilizar el formato Markdown de Backlog.md como formato de intercambio propio de
`biso`.** Heredaría sus bugs conocidos, no tiene sitio para lo que `biso` sí modela, y la garantía de
simetría de `biso export` con `biso new --from` no se podría cumplir. Un Markdown propio de `biso` sería
una función nueva de `biso`, no de esta utilidad.
