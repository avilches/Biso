# Decisiones de detalle

## Decisiones de detalle que cuesta reconstruir

**Una tarea sin autor es válida.** El campo `author` toma la identidad configurada al
crear la tarea, y si no hay ninguna se queda vacío **sin avisar**. Es deliberadamente distinto de los
otros cinco sitios donde hace falta una identidad (tabla de ["Variables de entorno"](../spec/invocacion.md#variables-de-entorno)): el
filtro `--mine` falla, la autoasignación de `start` avisa, y el autor de un comentario, `biso ask` y
`biso answer` son un error. La razón es que un tablero de una sola persona no tiene por qué
configurar su identidad solo para poder crear tareas.

**`biso export --json` es un error y no un flag sin efecto.** `export` es el único comando cuya
salida ya es JSON sin pedirlo, en forma de un objeto por línea, mientras que `--json` significa el
sobre único que imprimen todos los demás. Son dos formas distintas, y aceptar el flag en silencio
dejaría en duda cuál de las dos sale.

---

## El juego de caracteres de un token

**La decisión.** Cerrar el alfabeto de `labels` y `assignees` a letras y dígitos Unicode más los
símbolos `- _ . : @`, sin espacio, siguiendo a Taskwarrior y Jira y no a GitHub (comparación completa
más abajo): biso es un programa que un agente maneja tecleando líneas de comandos, no un formulario
web, así que cada etiqueta con espacio sería una comilla que ese agente tendría que acordarse de poner
siempre, para ganar exactamente lo mismo que ya ofrecen `-` y `_`. Y olvidar la comilla no siempre
falla alto: según qué flags haya alrededor, la palabra suelta puede convertirse en un argumento
inesperado (el caso bueno, error 2) o colarse donde no tocaba. Cerrar el alfabeto quita el problema de
raíz en vez de pedir disciplina. `references` y `dependencies` no
llevan esta restricción, y una clave de `ext` lleva un tercer alfabeto distinto; el porqué de cada una
está más abajo.

**El problema que la motiva.** La especificación no restringía ningún carácter en `labels`,
`assignees`, `references`, `documentation`, `dependencies`, `modifiedFiles` ni en una clave de `ext`
(`documentation` y `modifiedFiles` eran entonces campos distintos de `references` y ya no existen, ver
["Se retira `documentation` y `references` queda como único campo de punteros"](#se-retira-documentation-y-references-queda-como-único-campo-de-punteros)
y ["Se retira `modifiedFiles`"](#se-retira-modifiedfiles)):
la única regla escrita era que una coma dentro de un valor se escapa con `\,`, lo que de hecho dejaba
pasar espacios, saltos de línea y cualquier símbolo. `labels` y `assignees` se imprimen en las columnas
de ancho fijo de `biso ls` y `biso prime`, así que un espacio o un salto de línea dentro de uno de esos
tokens rompía la tabla sin que hubiera forma de distinguir, al leerla, un token con espacio de dos
tokens separados. Hacía falta cerrar el alfabeto en algún punto, y la pregunta era dónde.

**Qué hacen las herramientas comparables.** Dos de ellas escriben una etiqueta como palabra suelta de
una línea de comandos, que es exactamente la situación de biso:

- **Taskwarrior** exige que una etiqueta sea una sola palabra que no empiece por dígito, puntuación ni
  operador matemático. No hay espacio posible, por regla.
- **Jira** rechaza directamente cualquier etiqueta con espacio, con el mensaje literal
  `Labels can't have spaces`, y recomienda `-` o `_` para una etiqueta de varias palabras.
- **GitHub** sí permite etiquetas de varias palabras (`good first issue`), pero nunca se enfrenta al
  problema de biso: una etiqueta de GitHub se elige en un desplegable de un formulario web o llega ya
  como cadena entrecomillada dentro de un JSON de su API, nunca como una palabra suelta que una shell
  tenga que trocear antes de que el programa la vea.

Cerrar el alfabeto de estos dos campos tiene una consecuencia que también hay que anotar: **el escape
de coma deja de aplicarles**. La única razón para escapar una coma es poder meterla como contenido
literal de un valor, y una coma no está en el alfabeto cerrado de `labels` ni de `assignees`, así que
ahí nunca hay una coma legítima que escapar. El escape sigue haciendo falta para `references`,
que sigue siendo texto libre.

**Por qué `references` y `dependencies` quedan fuera.** De los
campos de lista con coma que no son `labels` ni `assignees`, uno (`references`)
guarda contenido cuyo alfabeto no lo decide biso: una referencia puede ser una URL o la ruta del
sistema de ficheros de un documento o de un fichero de código. Cerrarle el alfabeto dejaría fuera
casos legítimos (`/`, `?`, `#` de una URL; `/` de una ruta) a cambio de nada, porque no se imprime en
una columna de ancho fijo con otros de su misma clase de la forma en que lo hacen las etiquetas. El
otro, `dependencies`, no guarda ni una URL ni una ruta y no necesita esta razón: no necesita
ninguna regla nueva porque ya tiene la suya, distinta de la de `references`. Cada elemento es un `<ref>` y lo gobierna entera la gramática de ["Cómo se resuelve una referencia a una tarea"](../spec/referencias.md), que ya distingue un identificador
mal formado de una consulta de texto libre.

**Por qué la clave de `ext` es un tercer alfabeto y no el mismo que `labels`.** Una clave de `ext`
tiene la forma de un espacio de nombres con punto (`trello.card`, `github.issue`), así que necesita el
punto y admite `-`/`_` para el segmento, pero no tiene ningún uso documentado para `@` ni para `:`. Y
no puede admitir `=`, porque `--ext <clave>=<valor>` ya usa ese carácter para separar la clave del
valor: permitirlo dentro de la clave haría ambiguo dónde termina una y empieza el otro en
`--ext a=b=c`.

**Por qué el código de salida es 2 y no 3.** El código 3 (["El código 3 cubre dos direcciones"](../spec/codigos-de-salida.md#el-código-3-cubre-dos-direcciones)) es para un valor
sintácticamente correcto que el tablero no reconoce, y `labels` y `assignees` no tienen vocabulario
cerrado al escribir (["Qué valida cada filtro, y contra qué"](../spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué)): cualquier texto que cumpla el alfabeto es válido sin que el
tablero lo declare antes. Un carácter fuera del alfabeto no es un problema de reconocimiento sino de
forma, la misma clase de fallo que un identificador mal formado, que ya es código 2. Tratarlo como
código 3 habría mezclado dos preguntas distintas bajo el mismo número: "¿es sintácticamente válido?" y
"¿el tablero lo tiene?".

---

## Dónde se anuncia la clave de un criterio recién creado

**La decisión.** Cuando `--add-ac` crea un elemento en un comando distinto de
`biso new`, la línea de estado por defecto (["`biso set`"](../spec/cmd/set.md#salida) y los verbos del ciclo) añade `added ac #<clave>`
al final, y solo cuando la llamada crea de verdad algo. `biso new` no lo
anuncia nunca, ni siquiera cuando crea criterios a la vez que la tarea. En `--json`, `acAdded`
se añade a `data.tasks` del esquema `task.write`, presente siempre (vacío si no se creó
nada) en todos los comandos que comparten ese `kind`, incluido `new`.

**Por qué.** El principio 4 (["Los principios"](../spec/principios.md)) dice que la salida por defecto de una escritura es lo que
quien llama no sabía, nunca el eco de lo que acaba de escribir. Sobre una tarea que ya existía, el
contador de claves de sus criterios viene de antes, y quien llama no puede saberlo sin leer la tarea
primero: es justo el dato que ese principio manda enseñar, y por eso va en la misma línea de estado
que ya enseña el resto de derivados (`ac X/Y`, `urgency`), no en un sitio aparte. Sobre una
tarea recién creada con `biso new`, en cambio, el contador siempre empieza en 1, así que
la clave de cada `--add-ac` es el mismo orden en que se escribieron los flags
(["Los criterios y sus claves estables"](../spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables)): quien llama ya lo sabe, y anunciarlo sería el eco que el principio 4
prohíbe.

**Alternativas descartadas, y por qué.** Una nota de stderr, con la misma forma que `note:` o
`warning:` de ["Notas y avisos"](../spec/salida-y-terminal.md#notas-y-avisos), encajaría con la convención existente, pero separaría este
dato del resto de derivados que viven en la línea de estado por la misma razón exacta (no se pueden
saber sin leer la tarea), y lo dejaría invisible para cualquier consumo que descarte stderr, sin que
haga falta: el contrato de estabilidad (["El contrato de estabilidad"](../spec/estabilidad.md)) ya dice que el texto exacto de la
línea de estado puede cambiar entre versiones menores, así que alargarla no rompe ninguna promesa. Una
línea aparte en stdout, por la misma razón, multiplicaría las líneas de salida por tarea sin necesidad:
el dato cabe en la línea que ya existe.

---

## Borrar o corregir la fecha de un comentario

**La decisión.** `biso` añade flags dedicados para comentarios ya escritos: `--rm-comment <sel>`
borra uno o varios enteros, y `--set-comment-date <sel>=<instante>` corrige solo su fecha. El
selector es el mismo de ["Selectores de criterios"](../spec/familias-de-flags.md#selectores-de-criterios), con la clave del comentario en vez de la del
criterio. **No existe, y no va a existir, ninguna forma de editar el cuerpo o el autor de un
comentario ya escrito.** Cada comentario recibe además una clave estable, igual que un criterio de
aceptación, porque sin ella no hay forma de señalar cuál se quiere borrar o corregir sin que se mueva
al borrar otro.

**El caso medido que lo motiva.** En una migración real con otra herramienta que guarda los
comentarios como texto en un fichero, se detectaron agentes editando ese fichero a mano para que las
fechas de varios comentarios coincidieran entre sí, precisamente porque la herramienta no ofrecía
ninguna vía legítima para corregir la fecha de un comentario ya escrito fuera de una importación
inicial. `biso` ya resuelve la mitad del problema: cualquier fecha se puede fijar al crear una tarea
por lote (["Cuatro requisitos aprendidos de otras herramientas"](requisitos-de-otras-herramientas.md)). Pero esa vía sirve para
poblar un tablero vacío, no para corregir una tarea que ya existe: un `id` ya ocupado falla al
importar (["`biso new`"](../spec/cmd/new.md)), así que hoy no hay ninguna forma de arreglar una fecha equivocada en una tarea
existente sin destruirla y recrearla entera. Ese hueco es el que empuja a la misma clase de atajo que
ya se vio en la otra herramienta, y en `biso` el atajo equivalente sería manipular directamente el
fichero de la base de datos SQLite por fuera de la CLI, que es peor que editar un markdown a mano
porque además puede dejar el almacén en un estado que `biso doctor` no sabe explicar.

**Por qué se acepta corregir la fecha o borrar, y no editar el cuerpo o el autor.** La garantía que
importa preservar es que un comentario es el registro de lo que se dijo, no de cuándo se archivó el
registro. Editar el cuerpo o el autor reescribiría la conversación misma, que es exactamente lo que
esta lista existe para impedir. Corregir la fecha, o borrar el comentario entero cuando de verdad
sobra, deja intacto el contenido de la conversación y solo toca un metadato o la presencia del
registro completo, así que no compromete esa garantía. Es la misma distinción que ya usa el modelo con
`leaseExpiresAt`/`leaseHolder`: se puede corregir un dato operativo sin que eso abra la puerta a
reescribir el historial de lo que pasó.

**Por qué una clave estable, y no un selector por posición.** Las claves de los criterios ya resuelven
el mismo problema (una posición se mueve al borrar un elemento de en medio, una clave no) y ya tienen
su propio selector completo. Darle a los comentarios una segunda forma de direccionarse, distinta y
más pobre, solo para ahorrarse un campo, habría creado dos maneras de resolver "cuál elemento de una
lista" en el mismo documento en vez de una. Reutilizar el selector entero, en cambio, significa que
quien ya sabe usar `--rm-ac` no aprende nada nuevo para usar `--rm-comment`.

---

## Se retiran `project` y `milestone`

**La decisión.** La tarea no tiene ningún campo dedicado a agrupar trabajo. La agrupación real, la de
una tarea grande con subtareas propias, se resuelve con `parent` (sección ["El modelo de datos de una tarea"](../spec/modelo-de-datos/index.md)):
cualquier tarea con hijas, sea cual sea su `type`, actúa como grupo, sin que haga falta marcarla de
ninguna forma especial. Quien quiera además distinguir esas tareas grandes de las demás por su
naturaleza puede declarar un valor `epic` en `types` (sección [`biso config`](../spec/cmd/config.md)) y
usarlo como cualquier otro tipo, sin que eso exija ningún campo ni comportamiento nuevo. La partición
plana, la que separa tareas por su naturaleza sin mirar la jerarquía, ya la da ese mismo `type`. Ni
`project`/`projects` ni `milestone` sobreviven a esta decisión: se retiran de la especificación
entera, sin sustituto.

**Por qué no un campo `project` en la tarea.** Era un escalar con vocabulario cerrado, declarado en la
lista `projects` de la configuración, igual que `type` o `priority`. Compartía nombre con
`project_name`, la clave que da nombre al tablero entero, así que dos conceptos completamente
distintos ("a qué proyecto lógico pertenece esta tarea dentro del tablero" y "cómo se llama el
tablero") competían por la misma palabra en la misma herramienta. Y no había ningún caso de uso
medido, en ninguna de las herramientas comparables de ["Estado del arte"](../estado-del-arte/index.md), que pidiera agrupar las
tareas de un mismo tablero por un proyecto declarado aparte.

**Por qué no un campo `epic` nuevo.** Se consideró, y se descarta. `parent` ya resuelve exactamente lo
mismo con una tarea real detrás, con su propio `status` y sus propios criterios, en vez de un texto
suelto sin ciclo de vida. Añadir `epic` como escalar habría creado dos nombres para el mismo concepto,
"de qué agrupación mayor es esto parte", en contra del principio 2 de la especificación. La
confirmación de que `parent` es la vía correcta viene de herramientas pensadas específicamente para
agentes de código: Beads, de Steve Yegge, no tiene ningún campo `epic` separado, trata `epic` como un
valor más de su `type` y agrupa enteramente por el grafo de padres; TaskMaster AI resuelve lo mismo
con subtareas anidadas de identificador estable, también sin campo dedicado.

**Por qué se retira también `milestone`, después de haberlo defendido como campo.** Corrige
["Los hitos como entidad"](lo-que-se-deja-fuera.md), en la página de lo que se deja fuera. `milestone`
cubría un caso real y distinto del de `parent`: un cajón para agrupar tareas sueltas sin crear una
tarea nueva, sin exigirles ciclo de vida y sin configuración previa. Ese caso sigue siendo real, pero
`milestone` no era la única manera de resolverlo y arrastraba una palabra con connotación de fecha
límite que no tenía nada que ver con lo que el campo hacía, que era texto puro sin fecha propia. La
vía que queda escrita para el futuro, si el caso vuelve a aparecer con datos que lo justifiquen, son
las labels con ámbito al estilo GitLab (`clave::valor`, por ejemplo `group::Decisiones`), que ya caben
en el alfabeto cerrado de `labels` (sección ["El juego de caracteres de un token"](#el-juego-de-caracteres-de-un-token): el `:` ya está
permitido) y que añadirían exclusividad real, como mucho un valor por `clave` en la misma tarea, sin
declarar nada de antemano en la configuración, a diferencia de `ext`. No se implementa ahora: es la
vía descartada que queda anotada para no reabrir la pregunta sin motivo, con el mismo criterio de
evidencia con el que se había defendido `milestone` la primera vez.

**Por qué no `ext`.** `ext` ya es un mapa de clave declarada a texto, pero su papel declarado es
guardar la identidad de la tarea en otro sistema, no agrupar, y no tiene ningún filtro en `biso ls` ni
vocabulario derivado. Usarlo para agrupar habría exigido, o bien construirle un filtro y un vocabulario
derivado que hoy no tiene, que es exactamente lo que ya hacía `milestone` y por tanto no ahorra nada,
o bien agrupar en el board por una dimensión que no se puede consultar por ningún otro sitio de la
herramienta, una asimetría nueva entre lo que se ve y lo que se puede pedir por la línea de comandos.

---

## Se retira la definición de hecho

**La decisión.** Una tarea tiene una sola lista de comprobación, `acceptanceCriteria`. El campo
`definitionOfDone` no existe, ni sus flags (`--add-dod`, `--rm-dod`, `--clear-dods`, `--check-dod`,
`--uncheck-dod`), ni sus derivados (`dodDone`, `dodTotal`, `dodAdded`), ni el trozo `dod X/Y` de la
línea de estado, ni el aviso de cerrar con la definición de hecho a medias. Lo que otra herramienta
guardaría en una segunda lista, del tipo "alguien más lo ha revisado", es un criterio de aceptación
más. La única huella que queda es de entrada: un lote de `biso new --from` que traiga
`definitionOfDone` no falla, sino que convierte cada elemento en un criterio de aceptación y avisa
(sección [`biso new`](../spec/cmd/new.md)).

**La medida que lo decide.** El 2026-09-19 se contaron las tareas de los cinco tableros de Backlog.md
de esta máquina, la herramienta de la que `biso` hereda el campo: 420 tareas entre los cinco, con
criterios de aceptación en las de los tableros que los usan y **cero** con una sección de definición
de hecho. En 2,1 GB de transcripciones de sesiones de agente hay una sola ejecución real de `--dod`,
del 2026-09-15, y fue para inspeccionar qué forma tenía el Markdown resultante, no para trabajar. El
campo no se usaba poco: no se había usado nunca.

**Por qué no se usaba, que no es lo mismo que por qué sobra.** En Backlog.md la definición de hecho
está pensada como plantilla de proyecto, una lista reutilizable que se aplica sola a cada tarea nueva.
Consta en la propia herramienta, versión 1.52.0 instalada en esta máquina: `backlog task create`
ofrece `--no-dod-defaults` para desactivar esos valores por defecto del proyecto, y la guía que el
CLI da a los agentes, `backlog instructions task-creation`, dice literalmente que esos valores se
aplican solos y que solo hay que añadir elementos propios de una tarea cuando esa tarea necesite
higiene de cierre extra. Esa plantilla nunca se configuró en
ninguno de los cinco tableros, así que la sección no llegó a aparecer jamás. `biso` había copiado la
mitad equivocada de ese diseño: se trajo la lista de la tarea y dejó fuera el nivel de proyecto, que
era de donde venía el sentido. Una definición de hecho que hay que teclear tarea por tarea no es una
definición de hecho, es una segunda lista de criterios de aceptación con otro nombre, y el principio 2
de la especificación (["Los principios"](../spec/principios.md)) prohíbe exactamente eso, dos nombres
para el mismo concepto.

**El estado del arte confirma que el campo es una rareza.** De los seis gestores investigados en
["Esquemas de datos externos"](../estado-del-arte/esquemas-de-datos-externos.md), ninguno tiene dos
listas de comprobación fijas. Taskwarrior no tiene ninguna; Beads tiene `acceptance_criteria` como un
único bloque de texto libre, no una lista de elementos marcables; Task Master resuelve con subtareas
anidadas y un campo `testStrategy`; GitHub Issues y Linear no tienen campo de checklist y descomponen
con sub-issues; y Trello tiene `Checklist` con `CheckItem`, que es la forma exacta de un criterio,
pero en un array de cuantas listas nombradas quiera quien las cree, que es la generalización a N
listas y no un modelo de dos niveles. Fuera de esos seis, Jira, la herramienta donde nació el
vocabulario de Scrum del que salen los dos términos, no tiene ninguno de los dos como campo nativo, y
la documentación de Atlassian dice por qué: los criterios de aceptación pertenecen a un elemento
concreto, mientras que la definición de hecho aplica a todos los elementos del sprint o del proyecto.
Es un acuerdo de equipo, no un campo de la tarea.

**Alternativa descartada: darle a `biso` una definición de hecho por tablero.** Habría sido la forma
fiel al concepto original, una lista declarada en la configuración que se copia en cada tarea nueva.
Se descarta porque añade una clave de configuración, un momento de aplicación que hay que especificar
(qué pasa con las tareas ya creadas cuando la plantilla cambia) y una lista que crece sola en cada
tarea, todo ello para un caso de uso que en 420 tareas reales no apareció ni una vez. Si algún día
aparece, el sitio por donde entrar está escrito aquí.

**Alternativa descartada: N listas nombradas, al estilo de Trello.** Es la generalización correcta y
resuelve de paso cualquier separación futura, pero multiplica la superficie por la vía contraria a la
que sigue la herramienta: cada lista necesitaría nombre, un selector que la nombre en todos los flags
de marcado, y un sitio propio en la línea de estado, que hoy cabe entera en una línea justamente
porque solo hay una lista que contar.

**Qué cuesta, y por qué ahora.** La retirada toca diecinueve ficheros de la especificación, nueve de
los trece escenarios del tutorial (el séptimo reescrito entero y los demás retocados), la tabla de
correspondencia con otros modelos y unos 125 bytes de la parte fija del mensaje de arranque, que
tiene tope duro (["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)).
Se hace ahora porque no hay ni una línea de código Go escrita, así que el cambio es enteramente de
documentación; cada día que el campo siguiera en la especificación sería un día más de superficie que
después habría que implementar, probar y mantener para algo que nadie rellena.

---

## Se retira `documentation` y `references` queda como único campo de punteros

**La decisión.** Una tarea tiene un solo campo de punteros, `references`, y un documento es una
referencia más. El campo `documentation` no existe, ni sus flags (`--add-docs`, `--rm-docs`,
`--clear-docs`, `--replace-docs`), ni su línea `docs` en la ficha de `biso get`, ni su clave en el
JSON, ni su línea en la rejilla `FIELD FLAGS` del mensaje de arranque. La única huella que queda es de
entrada: un lote de `biso new --from` que traiga `documentation` no falla, sino que añade cada
elemento al final de `references` y avisa (sección [`biso new`](../spec/cmd/new.md)). Se conserva
`references` y no `documentation` porque su palabra abarca el conjunto: un informe de bug o una URL no
son documentación, y la ruta de una página de la especificación sí es una referencia.

**La medida que lo decide.** El 2026-09-20 se contaron las 88 tareas del tablero de este mismo
proyecto: 34 usan `documentation`, 5 usan `references` y **ninguna usa las dos**. De las 5 que usan
`references`, tres apuntan a rutas de la especificación, que es exactamente lo que ponen las otras 34
en `documentation`; una lleva prosa y otra el identificador de otra tarea. Los dos campos se llenan
con la misma clase de contenido y nadie los ha usado a la vez, así que no hay ninguna tarea donde la
distinción diga algo.

**Por qué no había distinción que conservar.** La especificación nunca definió en qué se
diferenciaban, y ninguna de las dos listas se valida, se resuelve a una tarea, se filtra ni entra en
ningún cálculo: las dos son texto libre que se guarda, se imprime y viaja en la exportación
(["Los punteros: `references`"](../spec/modelo-de-datos/relaciones.md#los-punteros-references)).
Backlog.md, de donde `biso` copió los dos campos, tampoco los distingue: su ayuda dice `add reference
URL or file path` para uno y `add documentation URL or file path` para el otro. Y de las siete
herramientas de ["Estado del arte"](../estado-del-arte/index.md), Backlog.md es la única que tiene
estos dos campos: Linear y Trello tienen `attachments`, que es un solo objeto y no dos, y las otras
cuatro no tienen nada. Dos nombres para el mismo concepto es justo lo que prohíbe el principio 2 de
la especificación (["Los principios"](../spec/principios.md)).

**Alternativa descartada: conservar los dos y definir la diferencia.** Por ejemplo, `documentation`
para lo que gobierna el trabajo y `references` para lo que solo lo acompaña. Se descarta porque la
definición tendría que inventarse ahora, sin ningún uso que la respalde (34 tareas frente a 5, y ni
una con las dos), y porque una frontera así obligaría a quien escribe a decidir en cada puntero de
qué lado cae, para un dato que ni se valida ni se consulta.

**Alternativa descartada: conservar `documentation` y retirar `references`.** Es la que más tareas de
este tablero usan hoy, pero su palabra es más estrecha: un informe de bug, una dirección web o el
identificador de otra tarea con la que hay que ser coherente no son documentación, y con ese nombre
el campo pediría un segundo para todo lo demás, que es donde se empezó.

**Alternativa descartada: una lista de adjuntos con nombre, al estilo de `attachments`.** Es lo que
tienen Linear y Trello, y resolvería de paso cualquier separación futura, pero es otro objeto con
título y dirección por elemento, con sus propios flags y su propia forma en la ficha y en el JSON, para
un campo que hoy es una lista de textos que nadie interpreta. Si algún día hace falta distinguir la
naturaleza de un puntero, el sitio por donde entrar es este.

**Qué cuesta, y qué se libera.** La retirada quita de la rejilla `FIELD FLAGS` del mensaje de arranque la línea de los flags
de `documentation`. El mensaje tiene un tope duro de 5.504 bytes (["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño))
y con el tablero del ejemplo pasa de 5.089 a 5.038 bytes, 51 menos, todos de la parte fija, que baja
de 3.600 a 3.549 de los 3.840 que tiene asignados. El campo deja además de contar entre los de lista,
que pasan de seis a cinco. También toca la tabla de
correspondencia con otros modelos, donde la cobertura se recalcula sin el campo retirado
(["Compatibilidad de modelos"](../estado-del-arte/compatibilidad-de-modelos.md)). Y hay una pérdida que
se acepta a propósito: al importar desde Backlog.md ya no queda constancia de cuál de las dos listas
era cada puntero, así que el conversor que reconstruya el formato de Backlog.md tendrá que guardar en
`ext` de qué campo venía cada valor si quiere la ida y vuelta exacta, igual que se propuso para los
identificadores con punto. Se hace ahora porque `biso` todavía no se ha publicado: no hay ningún
tablero ajeno con el campo escrito que migrar.

---

## Se retira `modifiedFiles`

**La decisión.** Una tarea no tiene un campo de ficheros tocados. El campo `modifiedFiles` no existe,
ni sus flags (`--add-files`, `--rm-files`, `--clear-files`, `--replace-files`, y con ellos el
`--add-files` que anunciaba `biso finish`), ni su línea `files` en la ficha de `biso get`, ni su
clave en el JSON, ni su línea en la rejilla `FIELD FLAGS` del mensaje de arranque. Una ruta que valga
la pena señalar es una referencia más, en `references`, y la pregunta de qué código tocó un trabajo la
contesta el control de versiones. La única huella que queda es de entrada: un lote de
`biso new --from` que traiga `modifiedFiles` no falla, sino que añade cada elemento al final de
`references`, detrás de los de `documentation` si la misma línea también los trae, y avisa con
`imported_modified_files_merged` (sección [`biso new`](../spec/cmd/new.md)).

**Las tres medidas que lo deciden.**

1. **Uso real: 0 de 535.** El 2026-09-21 se contaron las tareas de los seis tableros de Backlog.md de
   esta máquina: 94 en Biso, 271 en Kex, 56 en HubApp, 43 en LedgerDashboard, 38 en Health y 33 en
   dotfiles. Ninguna tiene `modified_files`. No es un cero de imposibilidad: el CLI de Backlog.md
   1.52.0 ofrece `--modified-file`, así que quien quiso rellenarlo pudo. Es la misma clase de medida,
   sobre una base mayor, con la que se retiró la definición de hecho
   (["Se retira la definición de hecho"](#se-retira-la-definición-de-hecho)).
2. **Una sola de siete herramientas.** De las siete de
   ["Estado del arte"](../estado-del-arte/index.md), solo Backlog.md tiene el campo. La nota de
   GitHub Issues es la que más dice: allí los ficheros cambiados son una propiedad del objeto Pull
   Request y no del Issue, es decir, de la revisión del código y no de la tarea que la motivó
   (["Compatibilidad de modelos"](../estado-del-arte/compatibilidad-de-modelos.md)).
3. **Solape con `references`, una vez fundido `documentation`.** Con `documentation` ya dentro de
   `references` (["Se retira `documentation` y `references` queda como único campo de punteros"](#se-retira-documentation-y-references-queda-como-único-campo-de-punteros)),
   `references` y `modifiedFiles` eran las dos únicas listas de texto libre sin validación de ninguna
   clase: el mismo tipo, la misma ausencia de restricción de caracteres, la misma regla de escape de la
   coma, la misma exclusión de la búsqueda de texto, el mismo destino en la exportación y en la ficha.
   `dependencies` queda aparte porque sí se valida y porque entra en dos términos de la urgencia. Lo
   único que distinguía a las dos era la etiqueta que imprime la ficha, `refs` frente a `files`, y dos
   nombres para el mismo concepto es justo lo que prohíbe el principio 2 de la especificación
   (["Los principios"](../spec/principios.md)). En el código eran trece apariciones fuera de las
   pruebas y todas de transporte: sin filtro, sin cálculo, sin validación propia y sin aviso.

**Alternativa descartada: conservarlo porque es el único campo que contesta qué código tocó este
trabajo.** Es el caso en contra más serio, porque esa pregunta es propia del público al que apunta
`biso`: agentes que trabajan sobre un repositorio. Se descarta porque ninguna parte del programa hace
cumplir esa distinción. Nada valida que una entrada de `modifiedFiles` sea una ruta, ni impide que sea
una URL o una frase, así que el campo era una convención, y una convención no necesita un campo del
esquema, un valor en un `CHECK` de SQL, una familia de flags y una línea del mensaje de arranque. Quien quiera
la convención la sigue con `--add-refs` y una ruta, y ni siquiera hace falta que la respete el resto del
equipo, porque el campo tampoco lo garantizaba. Y la respuesta fiable a esa pregunta no está en la
tarea sino en `git diff --name-only` sobre las ramas del trabajo, que no se queda desfasada cuando
alguien olvida apuntar un fichero.

**Alternativa descartada: conservarlo y fundir `references` en él.** Habría dejado un campo llamado
`modifiedFiles` que guarda URLs e identificadores de otras tareas. Es el problema que ya se resolvió
para `documentation`, con la palabra estrecha ganando sobre la ancha.

**Alternativa descartada: hacer `modifiedFiles` un campo validado**, que exija rutas relativas al
repositorio y compruebe que existen. Convertiría un campo que nadie rellena en el único de la tarea que
depende del sistema de ficheros de quien escribe, y un tablero que se lee desde otra máquina o desde
otra rama daría avisos por ficheros que en esa copia no están.

**Qué cuesta, y qué se libera.** La retirada quita de la rejilla `FIELD FLAGS` del mensaje de arranque
la línea entera de los flags del campo, 55 bytes. Con el tablero del ejemplo el mensaje pasa de 5.038 a
4.983 bytes, todos de la parte fija, que baja de 3.549 a 3.494 de los 3.840 que tiene asignados
(["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)). Los campos de lista
pasan de cinco a cuatro. La tabla de correspondencia con otros modelos se recalcula con un campo menos en el
total (["Compatibilidad de modelos"](../estado-del-arte/compatibilidad-de-modelos.md)). Y hay la
misma pérdida que se aceptó con `documentation`: al importar desde Backlog.md ya no queda constancia de
cuál de las tres listas era cada puntero, así que el conversor que reconstruya el formato de Backlog.md
tendrá que guardar en `ext` de qué campo venía cada valor si quiere la ida y vuelta exacta. Se hace
ahora porque el contrato de estabilidad (["El contrato de estabilidad"](../spec/estabilidad.md)) obliga a
partir de la versión 1.0, que no se ha publicado: no hay ningún tablero ajeno con el campo escrito que
migrar.

---

## La ayuda enseña la dirección de una dependencia

**La decisión.** Hacia dónde apunta una dependencia se enseña en la ayuda de `biso set` y de
`biso new`, con una frase y un ejemplo con identificadores del proyecto de ejemplo: la arista se
escribe siempre en la tarea que espera, y `biso set MYP-10 --add-deps MYP-4` dice que `MYP-4` va
primero y bloquea a `MYP-10`. La misma ayuda glosa `--parent` y `--add-refs`, que hasta entonces no
tenían glosa en ningún comando de escritura (`--parent` solo la tenía como filtro de lectura en
`biso ls`), y la frase de `biso set --help`
que decía que ningún flag exigía aprender nada más allá de su nombre se corrige, porque era cierta
para la forma de las listas y falsa para el significado de los campos de relación. La dirección
en sí no cambia: es la de
["Las relaciones entre tareas"](../spec/modelo-de-datos/relaciones.md#dependencies-la-precedencia).
El mensaje de arranque no lleva glosas: su parte fija tiene un tope propio
(["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)) y `FIELD FLAGS`
sigue siendo una rejilla de nombres.

**Por qué hace falta.** Es el único de los errores posibles al escribir una relación que el programa no
puede detectar. Un `--parent` a una tarea que no existe, o una dependencia que cierra un ciclo, fallan
al escribirse; una dependencia escrita al revés es una dependencia válida, y todo el cálculo de
bloqueo y de urgencia queda invertido sin ningún aviso. Con `biso new` hay además un caso propio:
`--add-deps MYP-4` hace que `MYP-4` vaya antes de la tarea nueva, y para que la tarea nueva bloquee a
una existente hay que crearla y luego escribir la arista con `biso set` en la existente, cosa que la
ayuda de `biso new` dice.

**Descartado: un flag inverso, `--add-blocks`.** Habría sido la forma de decir "esta tarea bloquea a
MYP-10" sin invertir la frase en la cabeza. Se descarta porque escribiría el campo `dependencies` de
una tarea que no es la nombrada en `<ref>`, y eso rompe la regla de que una escritura toca las tareas
que se nombran y ninguna más. `blocks` sigue siendo solo un campo derivado que se calcula al leer
(["Los campos derivados"](../spec/modelo-de-datos/index.md#los-campos-derivados)), y con un flag
inverso habría además dos maneras de escribir la misma arista, con la pregunta añadida de cuál gana
cuando se dan a la vez.

---

## La identidad de quien llama

**La decisión.** `me` y `default_limit` no son propiedades de un tablero, sino preferencias de quien lo
usa, y viven en la configuración de máquina, `~/.biso/config.json` (sección ["Configuración de máquina"](../spec/invocacion.md#configuración-de-máquina)),
junto a `boards_root` y `vcs`: `biso config` no las conoce, y se editan a mano en ese fichero como
cualquier otra clave suya. Entre la variable de entorno `BISO_ME` y la clave `me` de esa configuración
gana la variable, la misma precedencia que ya sigue el resto de la especificación (flag, luego
variable, luego configuración) y de la que esta era la única excepción. Con eso, un tablero compartido
entre una persona y un agente no necesita ninguna regla aparte para distinguirlos: la persona puede
dejar su `me` puesto en la configuración de su máquina, y un agente que fija su propio `BISO_ME` en el
entorno nunca lo hereda, tenga la máquina la clave puesta o no.

**Por qué no hay un comando que escriba `me` en la configuración de máquina.** Ninguna de las claves de
esa configuración tiene hoy un comando de escritura, `boards_root` y `vcs` incluidas, y `me` no rompe
esa regla. La falta de un atajo de una línea es justo lo que empuja hacia `BISO_ME`: la única receta
que cabe en un mensaje de error y se ejecuta sin abrir un editor es fijar la variable, así que todo
mensaje que hoy pedía una identidad nombra primero `BISO_ME` y, como segunda vía, la clave `me` de
`~/.biso/config.json` por su ruta completa, para que baste con leer el error para saber qué fichero y
qué clave tocar sin ir a buscarlo en esta página.

**Alternativas descartadas, y por qué.** Un flag `--me` se descarta porque `BISO_ME=@quien biso ...`
ya cubre exactamente el mismo caso, fijar la identidad para una sola invocación, sin añadir un flag
global más al contrato de estabilidad. Dejar `me` y `default_limit` en la configuración del tablero, y
limitarse a invertir la precedencia, se descarta porque entonces seguirían haciendo falta las tres
excepciones que hoy tienen: `biso snapshot` sin escribirlas, `biso init --from` ignorándolas con una
nota, y la promesa de simetría de ["El contrato de estabilidad"](../spec/estabilidad.md) con una excepción declarada; sacarlas
del tablero quita las tres a la vez, porque deja de haber nada que excluir. Un fichero hermano del
puntero del proyecto, buscado hacia arriba sin el tope de la sección ["El tope de la búsqueda hacia arriba"](../spec/resolucion-del-tablero.md#el-tope-de-la-búsqueda-hacia-arriba),
se descarta por dos motivos: fuera del control de versiones no resuelve nada que la configuración de
máquina no resuelva ya, y una identidad heredada en silencio de un directorio varios niveles por
encima es precisamente el error que esta decisión corrige, solo que en un sitio nuevo. Y un flag
`--global` o `--machine` para que `biso config` escriba `~/.biso/config.json` se descarta por ahora,
porque ninguna de las otras claves de esa configuración lo tiene y `me` no es un motivo suficiente
para dárselo solo a ella.
