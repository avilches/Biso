# Comandos y banderas

## El porqué de reglas concretas

Cada entrada dice la sección de la especificación a la que corresponde.

**["El algoritmo de coincidencia"](../spec/vocabularios.md#el-algoritmo-de-coincidencia), por qué no hay coincidencia por prefijo ni por parecido al resolver una referencia.** Una regla
que adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

**["Qué valida cada filtro, y contra qué"](../spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué), por qué las tareas archivadas cuentan en los conjuntos de etiquetas y personas contra
los que validan los filtros.** Es lo que impide que un filtro que hoy funciona deje de funcionar
mañana por archivar la última tarea que lo usaba.

**["Qué valida cada filtro, y contra qué"](../spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué), por qué las etiquetas y las personas no tienen vocabulario cerrado al escribir, pero sus
filtros sí validan.** No tienen vocabulario cerrado porque su utilidad es que se puedan inventar
sobre la marcha. Y validar al leer no es una asimetría con la escritura: es la aplicación del
principio 1, que dice que un filtro que no puede encajar con nada es un error y no una respuesta
vacía.

**Esta sección tenía una tercera entrada, sobre por qué el hito pasó a validar igual que una
etiqueta.** Ya no aplica: `milestone` se retiró entero, no solo su validación.
["Se retiran `project` y `milestone`"](detalles.md#se-retiran-project-y-milestone), más abajo en este documento,
explica la decisión vigente. Lo que queda de aquella entrada, y sigue siendo cierto, es que la
bandera `--unchecked` no es una asimetría del principio 1 porque quien la escribe pide explícitamente
una lista sin garantía, algo que el programa nunca decide callar por su cuenta.

**["Los tres mensajes de \"no la encuentro\""](../spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), por qué son distintos.** Porque las tres situaciones
piden acciones distintas de quien llama: corregir la sintaxis, dejar de buscar, o mirar en el archivo.

**["La regla"](../spec/familias-de-flags.md#la-regla), por qué cada clase de campo tiene las variantes de bandera que tiene, y por qué ninguna se
llama con el nombre desnudo del campo.** Antes había cuatro variantes por cada campo de lista
(`--campo` añadía, `--set-campo` sustituía, `--rm-campo` quitaba, `--clear-campo` vaciaba), y las
demás clases de campo tenían el subconjunto de esas cuatro que tenía sentido para su forma de dato:
un bloque de prosa no tiene elementos que quitar de uno en uno, así que no tenía `rm-`; un mapa de
claves se manipula por clave y no por posición; un escalar solo se fija o se vacía; los comentarios,
en aquel momento, solo admitían añadir. **Esa última pieza ya no es cierta**: ["Borrar o corregir la fecha de un comentario"](detalles.md#borrar-o-corregir-la-fecha-de-un-comentario), más abajo en
este documento, añade dos operaciones más, y la forma vigente de la clase está en
["Comentarios"](../spec/familias-de-flags.md#comentarios) y no aquí. Lo que sigue siendo cierto, y es lo que este párrafo quería decir, es
que la tabla de clases de campo aplica exactamente las operaciones que tienen sentido para cada forma
de dato, no una lista de excepciones: que el conjunto de un escalar sea distinto del de un mapa, o que
el de los comentarios haya crecido de uno a tres elementos, no es una excepción a la regla, es la
regla funcionando.

**Lo que cambió es el nombre desnudo.** Funcionaba, y resolvía el fallo medido del principio 3, pero
exigía conocer la regla de antemano para no adivinar: nada en `--label` dice que añade, hay que
haberlo leído en algún sitio. Se sustituyó por dar a cada operación su propio verbo (`--add-labels`,
`--rm-labels`, `--clear-labels`, `--replace-labels`, y así con cada campo), de modo que cualquier
bandera se entiende por su nombre sin haber leído esta sección. Dos consecuencias de ese cambio:

- **Los campos de lista sin coma (`ac`, `dod`) pierden la variante de "sustituir entera".**
  `--replace-ac`/`--replace-dod` habría sido repetible igual que `--add-ac`, y repetir una bandera de
  sustituir no la sustituye dos veces: acumula sus valores y sustituye una sola vez con el conjunto
  acumulado (["Repetición y listas separadas por comas"](../spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas)), que es una segunda pieza de comportamiento no obvio
  encima del nombre. Sustituir esas dos listas se hace vaciando y añadiendo en la misma llamada
  (["Orden de aplicación dentro de una escritura"](../spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura)), que ya hacía falta declarar para todo lo demás y cubre el mismo caso sin una
  bandera más que aprender.
- **`--check` y `--uncheck` pasan a `--check-ac` y `--uncheck-ac`, simétricos con `--check-dod` y
  `--uncheck-dod`.** La forma anterior reservaba el nombre desnudo para los criterios de aceptación y
  obligaba a la definición de hecho a llevar el sufijo, una asimetría que la propia especificación
  declaraba como excepción sin más justificación que la de ser la primera lista de las dos. Bajo la
  regla nueva no hay sitio para reservar un nombre desnudo a nada, así que la asimetría desaparece
  sola en vez de quedar documentada como caso especial.

**Esto había dejado pendiente el mensaje de arranque, y ya no lo está.** El bloque `FIELD FLAGS` de
["La salida literal"](../spec/cmd/prime.md#la-salida-literal) enseñaba antes una sola bandera por campo (la que añadía) y una regla en `RULES`
explicaba cómo derivar las otras tres; ese ahorro de espacio dependía exactamente del mecanismo que
se acaba de retirar. Cómo enseñar las banderas nuevas dentro del presupuesto de bytes se dejó sin
decidir a propósito, sin tocar `cmd/prime.md` ni `presupuestos.md`, hasta medirlo con un agente real
en vez de razonarlo: la decisión, con el experimento que la sostiene, está en
["El grid completo de banderas de campo en el mensaje de arranque, medido con un agente real"](vocabulario-y-mensaje-de-arranque.md#el-grid-completo-de-banderas-de-campo-en-el-mensaje-de-arranque-medido-con-un-agente-real),
más abajo en este documento.

**["Selectores de criterios"](../spec/familias-de-flags.md#selectores-de-criterios), por qué quitar un criterio de aceptación toma un selector y no un texto.** Porque quitarlo por
su texto exacto es más frágil que quitarlo por su clave.

**["Campos externos"](../spec/familias-de-flags.md#campos-externos), por qué no existe una bandera que sustituya el mapa de campos externos entero.** Fijar una clave
ya es sustituir su valor, así que una segunda bandera para lo mismo solo serviría para equivocarse. Y
una que sustituyese el mapa entero con la sintaxis `clave=valor` sería una forma silenciosa de borrar
la identidad externa de una tarea al escribir otra.

**["La salida literal"](../spec/cmd/prime.md#la-salida-literal), por qué el bloque de tareas en curso del mensaje de arranque no tiene límite.** Porque en un
tablero sano son pocas.

**["`biso init`"](../spec/cmd/init.md), por qué el puntero del proyecto es la única cosa que `init` escribe fuera del tablero.** Sin
ella, un tablero creado en otra ubicación no lo encontraría ningún comando posterior.

**["`biso init`"](../spec/cmd/init.md), por qué se retira la regla posicional que guardaba el estado activo como el penúltimo de
`--statuses`.** La regla estaba rota, y la contradicción que la delata vive en el propio documento: el
tablero de ejemplo era `Ideas, To Do, In Progress, Blocked, Done`, cuyo penúltimo es `Blocked`, y había
un ejemplo literal de `biso init` que lo creaba así, mientras que tanto la salida de `biso config list`
como el esquema JSON de `biso prime` declaraban que el estado activo de ese mismo tablero era
`In Progress`. Las dos cosas no podían ser ciertas a la vez, y la regla solo parecía funcionar porque
el tablero por defecto tenía justo tres estados. Se sustituye por tres banderas explícitas,
`--initial-status`, `--active-status` y `--terminal-status`, con el mismo argumento de ["El algoritmo de coincidencia"](../spec/vocabularios.md#el-algoritmo-de-coincidencia): una regla
que adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

**["`biso init`"](../spec/cmd/init.md), por qué el tablero por defecto no trae un estado `Ideas`.** Un estado `Ideas` no dice nada que
no diga ya estar sin asignar, que se consulta con `biso ls --unassigned`. El matiz que sí aporta,
"esto quizá no lo hagamos nunca", tiene ya una decisión con evidencia detrás en ["Lo que se miró de ese diseño anterior y se descarta"](modelo-de-estados.md#lo-que-se-miró-de-ese-diseño-anterior-y-se-descarta): se
cubre con un tipo más del vocabulario que ya existe y no con un estado. Y hay un motivo peor para no
ponerlo por defecto: si `Ideas` fuera el estado inicial, toda tarea nueva nacería ahí, y el bloque
`NEXT UP` del mensaje de arranque mezclaría "algún día quizá" con "hay que hacerlo", que es justo la
distinción que ese bloque existe para hacer.

**["`biso new`"](../spec/cmd/new.md), por qué existe `--start` al crear una tarea.** Evita que crear una tarea para ponerse con ella
en el mismo minuto cueste dos llamadas. Es el principio 5 aplicado a un caso medido.

**["`biso new`"](../spec/cmd/new.md), por qué `--comment` funciona al crear.** Por lo mismo: una tarea que nace con un comentario
cuesta una llamada.

**["`biso ls`"](../spec/cmd/ls.md), por qué el filtro de etiquetas es el único que combina sus valores con "y".** Porque el uso
normal de varias etiquetas es acotar, no ampliar.

**["`biso ls`"](../spec/cmd/ls.md), por qué el filtro de dependencias es `--blocked` y `--not-blocked`, y no `--ready`.** El hecho
que se calcula es uno solo, que alguna dependencia esté sin terminar, así que se nombra una vez y su
negación se forma con el mismo prefijo que los otros dos pares booleanos de `biso ls`,
`--waiting`/`--not-waiting` y `--active`/`--not-active`. El nombre `ready` sobraba por dos motivos
distintos. El primero es que hacía viajar el mismo hecho dos veces en el JSON, como `ready` y como
`blocked`, y campos que dicen lo mismo acaban divergiendo. El segundo es que prometía más de lo
que cumplía: miraba solo dependencias, así que `biso ls --ready` devolvía también las tareas aparcadas
en una pregunta, que es justo lo que un agente no puede coger. De los dos nombres sobrevive `blocked`
porque ya tiene entrada propia en la tabla de ["Vocabulario de esta especificación"](../spec/vocabulario.md), porque da nombre al
término `urgency.blocked` de la fórmula de urgencia, y porque nombra el hecho que de verdad se calcula.
Y el nombre nuevo tampoco promete estar lista para trabajar, porque ninguna bandera sola puede: eso
son varios filtros, y cuántos depende de qué se busque. Descartar lo bloqueado y lo aparcado son dos,
`--not-blocked --not-waiting`; quien quiera además tarea sin empezar añade `--not-active`, que es el
filtro con el que el propio mensaje de arranque describe su bloque `NEXT UP` (["La salida literal"](../spec/cmd/prime.md#la-salida-literal)); y quien la quiera
sin dueño, `--unassigned`. El cambio quita una clave del JSON y renombra una bandera,
que son las dos cosas que el ["contrato de estabilidad"](../spec/estabilidad.md) promete no tocar nunca, y por eso se hace
ahora: ese contrato obliga a partir de la versión 1.0 y todavía no hay ninguna versión publicada.
Después de 1.0 esta limpieza ya no se podría hacer.

**["`biso set`"](../spec/cmd/set.md), por qué `set` no repite en su tabla las banderas de campo.** Porque repetirlas invitaría a que
divergieran, que es como se rompen los documentos largos.

**["`biso set`"](../spec/cmd/set.md), por qué la línea de estado encoge cuando la tarea no tiene criterios, en vez de imprimir un
guion como hace el listado.** Las dos salidas parecen contradecirse y no lo hacen, porque no son la
misma clase de cosa. El listado de ["`biso ls`"](../spec/cmd/ls.md) es una tabla: sus columnas se rellenan al ancho del valor más
largo de la llamada, así que una celda vacía tiene que ocupar su sitio o las filas de abajo se
descolocan, y para eso está el guion. La línea de estado sale una por tarea afectada, sin ancho
compartido y sin nada que alinear debajo, de modo que un hueco no descoloca nada y un guion solo
añadiría un símbolo más que interpretar. Quien quiera los contadores siempre, estén las listas vacías
o no, pide `--json`, que trae los cuatro como números.

**["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](../spec/cmd/verbos-del-ciclo.md), por qué `finish` avisa de los criterios sin marcar y no lo impide.** Un criterio puede haber
quedado obsoleto, y un comando que no deja cerrar empuja a rodearlo con `set`, que es como se aprende
a esquivar una herramienta. Para quien quiera la política dura está `--strict`.

**["`biso export`"](../spec/cmd/export.md), por qué `export` no hereda los valores por defecto de `ls`, y por qué sale con código 6 y no
con 0 cuando salta una tarea ilegible.** Porque exportar de más nunca hace daño y exportar de menos en
silencio arruina una copia de seguridad. Es el único comando cuyo propósito es no perder nada, y por
eso es la única excepción a la regla general de las lecturas de conjunto.

**["`biso help`"](../spec/cmd/help.md#biso-help), por qué `help` funciona sin tablero.** Porque es lo primero que alguien ejecuta cuando algo
no va.

**["Repetición y listas separadas por comas"](../spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas), por qué los campos de texto largo no se parten por comas.** Porque una coma dentro de una frase
es normal, y partir por ella convertiría una descripción en varias.

**["Las fechas"](../spec/modelo-de-datos/fechas.md#las-fechas), por qué las fechas se pueden fijar al importar y no en el uso normal.** Sin esa excepción no se
puede importar el histórico de otro sistema conservando cuándo pasó cada cosa, que es el tercer
requisito de ["Cuatro requisitos aprendidos de otras herramientas"](principios-y-mantenimiento.md#cuatro-requisitos-aprendidos-de-otras-herramientas), en este mismo documento.

---

## `<command>` repetible en `biso help`

**La decisión.** `biso help` acepta varios nombres de comando en la misma llamada
(["`biso help`"](../spec/cmd/help.md#biso-help)): `biso help finish note` imprime la ayuda completa de
`finish` y luego la de `note`, en el orden pedido, separadas por una línea en blanco, como si se
hubiera llamado `biso <cmd> --help` una vez por nombre. Con `--json`, `commands` trae un elemento por
nombre pedido, en el mismo orden, y sigue siendo solo el resumen de una línea de cada uno, nunca la
prosa. La validación es todo o nada y ocurre antes de imprimir nada: los nombres se resuelven en el
orden en que aparecen, y el primero que no existe hace fallar la llamada entera con el error de
siempre (los tres nombres más parecidos a ese), sin imprimir la ayuda de ningún comando de la
llamada, ni siquiera la de los que sí existen. `all` es un valor especial y no un nombre de comando:
combinarlo con cualquier nombre de comando en la misma llamada es un error de uso, código 2, tanto en
texto como en `--json`.

**Por qué.** El coste que resuelve esta tarea es el mismo de siempre: un agente que necesita el
detalle de las banderas de varios comandos para un solo encargo tenía que llamar a `--help` una vez
por comando, multiplicando llamadas al CLI y salida consumida, aunque los necesitara todos a la vez.
Hacer `<command>` repetible, con la misma notación `<command>...` que ya usa `<ref>...` en `finish`,
`archive` o `set`, resuelve ese coste con la sintaxis que la especificación ya usa para "uno o
varios", sin inventar una segunda forma. La validación todo o nada no es una regla nueva: es el mismo
principio de la sección ["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables),
que dice que una escritura sobre varias tareas no deja la mitad hecha y calla el resto, aplicado aquí
a una lectura. Entregar la ayuda de los nombres válidos y solo fallar en el inválido dejaría a quien
llama sin saber si le falta algo de lo que pidió sin releer la lista completa contra la salida.

**Alternativas descartadas, y por qué.** Aceptar `all` junto con nombres de comando, y tratarlo como
"la lista más la ayuda completa de esos nombres", se descarta porque las dos formas de salida (una
lista de una línea por comando, y la prosa completa de uno) no tienen un orden ni una combinación con
un significado único, y ya existen dos llamadas separadas para pedir cada cosa. Ignorar `all` en
silencio cuando aparece junto a otros nombres, tratándolo como si no se hubiera escrito, se descarta
por la misma razón de siempre contra el silencio: quien lo escribe por error nunca se entera de que
esa parte de su llamada no hizo nada. Devolver la ayuda de los nombres válidos y solo señalar el
inválido con un aviso, en vez de fallar la llamada entera, se descarta porque mezclaría una salida de
datos con un error a medias, y porque nada distingue en ese caso un nombre mal escrito de una llamada
que de verdad quería sólo los nombres que sí existían.
