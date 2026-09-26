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
llevan esta restricción; el porqué de cada una está más abajo. El `:` de una etiqueta tiene además un
significado, que fija ["Las etiquetas con ámbito"](#las-etiquetas-con-ámbito).

**El problema que la motiva.** La especificación no restringía ningún carácter en `labels`,
`assignees`, `references`, `documentation`, `dependencies` ni `modifiedFiles`
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

**Por qué el código de salida es 2 y no 3.** El código 3 (["El código 3 cubre dos direcciones"](../spec/codigos-de-salida.md#el-código-3-cubre-dos-direcciones)) es para un valor
sintácticamente correcto que el tablero no reconoce, y `assignees` no tiene vocabulario cerrado al
escribir, ni `labels` mientras su lista de la configuración no restrinja la clave de la etiqueta
(["Qué valida cada filtro, y contra qué"](../spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué)
y ["Las etiquetas con ámbito"](#las-etiquetas-con-ámbito)): cualquier texto que cumpla el alfabeto es
válido sin que el tablero lo declare antes. Un carácter fuera del alfabeto no es un problema de
reconocimiento sino de forma, la misma clase de fallo que un identificador mal formado, que ya es código
2. Tratarlo como código 3 habría mezclado dos preguntas distintas bajo el mismo número: "¿es
sintácticamente válido?" y "¿el tablero lo tiene?".

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
entera, sin ningún campo que los sustituya. Agrupar tareas sueltas, sin ciclo de vida ni tarea padre,
lo dan las etiquetas con ámbito de ["Las etiquetas con ámbito"](#las-etiquetas-con-ámbito).

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
límite que no tenía nada que ver con lo que el campo hacía, que era texto puro sin fecha propia. Lo
cubren las etiquetas con ámbito al estilo GitLab (`milestone::m1`, `group::Decisiones`), que dan
exclusividad real, como mucho un valor por clave en la misma tarea, sin ningún campo nuevo y sin
declarar nada de antemano en la configuración: ["Las etiquetas con ámbito"](#las-etiquetas-con-ámbito).
**La vía descartada que esta entrada dejaba anotada como pendiente ya no lo es: la sustituye esa otra
entrada.**

---

## Las etiquetas con ámbito

**La decisión.** Una etiqueta que contiene `:` es una etiqueta con ámbito. Su clave es el texto anterior
a los primeros dos puntos y su valor es el resto, y el separador, uno o dos puntos, decide cuántos
valores de esa clave admite una tarea: `clave:valor` admite varios y `clave::valor` admite como mucho
uno. La clave y el valor no pueden estar vacíos y el valor no puede empezar ni terminar en `:`; cualquier
otra forma con `:` (`a:`, `:a`, `a:::b`) es `malformed_label`, código 2. La clave se compara plegando
mayúsculas y minúsculas en todas partes, y el separador no cuenta al comparar dos etiquetas por su
valor, así que `milestone:m1` y `milestone::m1` son la misma etiqueta al consultar y al quitar. Se
consulta por clave con `--label clave:` (o `--label clave::`, que es lo mismo), que encuentra cualquier
valor de esa clave y se acepta también en `--label-or`. Y la lista `labels` de la configuración puede
declarar claves y valores, y con eso restringir lo que se escribe, sin que exista ninguna clave de
configuración nueva.

**Escribir.** `--add-labels k::v` deja a `k::v` como única etiqueta de la clave `k` en la tarea: quita
las demás de esa clave, sean `k:x` o `k::y`, y avisa nombrando cada una que quitó. `--add-labels k:v`
sobre una tarea que lleva `k::x` es error 6, porque el estado guardado no lo permite, y el mensaje
propone quitar primero la exclusiva. Dentro de una misma llamada, `k:a` junto a `k::b` es error 2 sin
mirar el orden, y dos exclusivas de la misma clave escritas en el mismo flag dejan la última con un
aviso; repartidas entre `--add-labels` y `--replace-labels` no decide la línea de comandos sino el
orden de los pasos de una escritura, que pone sustituir antes que añadir
(["Escribir una etiqueta con ámbito"](../spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito)).
La comprobación se hace contra el estado que queda después de aplicar los
quitar de esa llamada, así que `--rm-labels k::1 --add-labels k:2` en una sola línea funciona. La misma
regla vale para cada línea de un lote de `biso new --from`, con una salvedad: ahí dos exclusivas de la
misma clave no dejan la última sino que son un fallo de validación de esa línea, porque la lista
`labels` de una línea describe un estado guardado y quedarse con uno de los valores en silencio
perdería un dato que el fichero afirmaba (["El modo lote"](../spec/cmd/new.md#el-modo-lote)).
`--rm-labels k:v` quita `k::v` y al revés.

**Consultar.** `--label k:` y `--label k::` significan cualquier etiqueta de la clave `k`, con cualquier
valor y cualquier separador, y se unen con `y` a los demás `--label` como hoy. Como `a:` y `a::` nunca
son una etiqueta guardable, esa forma queda libre como sintaxis de filtro sin chocar con ninguna
etiqueta real. Una clave que el tablero no tiene es error 3 con las claves más parecidas, y
`--unchecked` apaga esa comprobación igual que las demás de etiquetas. El JSON de una tarea no lleva
ningún campo derivado por clave: quien lea `labels` parte por los primeros dos puntos.

**La lista `labels` de la configuración.** Vacía, que es como nace, deja pasar cualquier cosa. Con
entradas, cada una es de una de tres formas: una etiqueta plana (`pepe`), que solo se ofrece, por
ejemplo a una interfaz que muestre las etiquetas al crear una tarea, y no restringe nada; un par exacto
(`size::m`), que restringe la clave `size` a los valores declarados y fija su separador; y una clave
abierta (`milestone::`), que no restringe el valor y fija el separador. Una etiqueta que ninguna entrada
nombra sigue siendo libre. Escribir un valor que una clave restringida no declara, o la clave con el
otro separador, es error 3, porque es un valor bien formado que el tablero no reconoce, y el mensaje
lista los valores que esa clave admite. Una misma clave no puede aparecer en la lista con los dos
separadores, ni como clave abierta y con valores exactos a la vez. `biso config set labels` falla si
alguna tarea lleva un valor o un separador que la lista nueva prohíbe, y quitar una etiqueta plana no
falla nunca porque no restringía. `biso doctor` señala lo guardado que la lista no cubre. Una clave
declarada cuenta como conocida al consultar aunque ninguna tarea la use todavía. El mensaje de arranque
no lista la clave `labels`: un agente que quiera verla tiene `biso config get labels`, y el error de una
clave restringida ya dice qué valores admite.

**Agrupar.** El prefijo de una etiqueta con ámbito es un tercer eje de agrupación válido para
`biso board`, junto a `parent` y `type`. Arrastrar una tarea de un grupo a otro no escribe su orden
sino el campo por el que se agrupa: otra columna cambia `status`, el grupo de otro padre cambia `parent`,
otro grupo de tipo cambia `type` y otro valor de una clave con ámbito cambia esa etiqueta; moverse
dentro de un grupo, en cambio, reordena (["El orden manual es una clave de texto"](#el-orden-manual-es-una-clave-de-texto)).
Nada de esto está en la versión 1.0, porque `biso board` queda fuera de ella.

**Descartado: una lista `labels` cerrada entera.** Que, con entradas en la lista, cualquier etiqueta que
no case con ninguna sea error. Quien solo quisiera declarar `milestone::` tendría que listar además
todas las demás etiquetas que quisiera admitir, y las etiquetas planas de la lista existen para
ofrecerse, no para prohibir el resto.

**Descartado: una clave de configuración nueva para declarar las claves de ámbito.** La lista `labels`
ya existía, ya era el sitio de las etiquetas declaradas, y con esta decisión solo gana el papel de
restringir por clave. Una clave nueva habría sido la vigésima, con su validación, su flag en
`biso init`, su presencia en `biso snapshot` y su superficie en `biso config`.

**Descartado: dejar las claves sin declarar del todo, o declararlas siempre.** Sin ninguna declaración
posible, una clave mal escrita al guardar crea una clave nueva sin avisar y la cardinalidad queda a
merced del primer uso; con declaración obligatoria, el tablero pierde la libertad que las etiquetas
tienen hoy. La lista opcional da las dos cosas: libre si está vacía, restringida en lo que nombra.

**Descartado: que `k:v` sustituya a `k::x`, o que `k::v` sobre `k:x` falle.** La asimetría es
deliberada. Escribir `::` es declarar la intención de que sea el único valor, así que sustituir es lo
esperado. Escribir `k:v` sobre una clave que ya es exclusiva es casi siempre un dos puntos olvidado, y
aceptarlo rompería la exclusividad en silencio.

**Descartado: que añadir una exclusiva sobre otra sea error.** Pasar la tarea al valor siguiente
(`milestone::m1` a `milestone::m2`) es un solo gesto, y con un error habría que quitar antes la vieja.
El aviso, que nombra lo que se quitó, deja el mismo rastro.

**Descartado: consultar con un comodín, con un flag nuevo o con la clave a secas.** `--label 'k:*'`
falla en zsh sin comillas con "no matches found" antes de llegar a `biso`. Un flag `--label-key` suma
flags a la familia. `--label k` a secas choca con una etiqueta plana que se llame `k`. La
forma `k:` no necesita nada de eso, porque queda libre.

**Descartado: un campo derivado en el JSON con el valor de cada clave.** Es superficie de contrato que
se congela en la 1.0, y quien lo necesita, que es una vista agrupada, ya tiene `labels` entera y la
regla de análisis de esta entrada. Añadirlo después es un cambio compatible.

**Descartado: no plegar la clave, o tratar las formas raras como etiquetas planas.** Con la clave sin
plegar, `Milestone::a` y `milestone::b` en la misma tarea saltarían la exclusividad tecleando una
mayúscula. Con `milestone::` (el valor olvidado) admitida como etiqueta plana, un error de tecleo
entraría en silencio, que es justo lo que el vocabulario cerrado del resto del programa evita.

**Por qué ahora.** El contrato de estabilidad (["El contrato de estabilidad"](../spec/estabilidad.md))
obliga a partir de la versión 1.0, que no se ha publicado. Cambiar el alfabeto (`a:` deja de ser una
etiqueta legal), añadir la regla de análisis y dar significado a una clave de configuración que ya
existía es gratis hoy y costaría un ciclo de aviso después.

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
el campo pediría un segundo para todo lo demás, que es justo el problema de tener dos campos.

**Alternativa descartada: una lista de adjuntos con nombre, al estilo de `attachments`.** Es lo que
tienen Linear y Trello, y resolvería de paso cualquier separación futura, pero es otro objeto con
título y dirección por elemento, con sus propios flags y su propia forma en la ficha y en el JSON, para
un campo que hoy es una lista de textos que nadie interpreta. Si algún día hace falta distinguir la
naturaleza de un puntero, el sitio por donde entrar es este.

**Qué cuesta, y qué se libera.** La retirada quita de la rejilla `FIELD FLAGS` del mensaje de arranque la línea de los flags
de `documentation`. El mensaje tiene un tope duro de 5.504 bytes (["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño))
y con el tablero del ejemplo el mensaje medía entonces 5.089 bytes, y la retirada le quitó 51, todos de
la parte fija, que pasó de 3.600 a 3.549 de los 3.840 que tiene asignados. Son medidas del momento de
esta decisión, no las vigentes, que están en esa misma página del presupuesto. El campo deja además de contar entre los de lista,
que pasan de seis a cinco. También toca la tabla de
correspondencia con otros modelos, donde la cobertura se recalcula sin el campo retirado
(["Compatibilidad de modelos"](../estado-del-arte/compatibilidad-de-modelos.md)). Y hay una pérdida que
se acepta a propósito: al importar desde Backlog.md ya no queda constancia de cuál de las dos listas
era cada puntero, así que el conversor que reconstruya el formato de Backlog.md tendrá que guardar por
su cuenta, por ejemplo en etiquetas con ámbito, de qué campo venía cada valor si quiere la ida y vuelta
exacta. Se hace ahora porque `biso` todavía no se ha publicado: no hay ningún
tablero ajeno con el campo escrito que migrar.

**Una consecuencia que sí existe: los tableros de desarrollo con datos en el campo retirado.** La migración
inicial se editó en el sitio, sin subir `user_version`, así que un tablero creado con un binario anterior
y con datos en `documentation` conserva su esquema viejo y queda parcialmente ilegible con el binario nuevo:
`biso ls` avisa y salta esas tareas, `biso get` falla con el error 3, y `biso doctor` las cuenta como
errores. No hay migración, porque `biso` no está publicado y el único remedio que se ofrece es recrear el tablero.

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
   (["Los principios"](../spec/principios.md)). En el código eran diecisiete líneas en nueve ficheros,
   sin contar las pruebas, y todas de transporte: sin filtro, sin cálculo, sin validación propia y sin aviso.

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
para `documentation`, donde se conservó `references`, la palabra que abarca el conjunto, y no la estrecha.

**Alternativa descartada: hacer `modifiedFiles` un campo validado**, que exija rutas relativas al
repositorio y compruebe que existen. Convertiría un campo que nadie rellena en el único de la tarea que
depende del sistema de ficheros de quien escribe, y un tablero que se lee desde otra máquina o desde
otra rama daría avisos por ficheros que en esa copia no están.

**Qué cuesta, y qué se libera.** La retirada quita de la rejilla `FIELD FLAGS` del mensaje de arranque
la línea entera de los flags del campo, 55 bytes. Con el tablero del ejemplo el mensaje medía entonces 5.038
bytes, y la retirada le quitó 55, todos de la parte fija, que pasó de 3.549 a 3.494 de los 3.840 que tiene asignados.
Son medidas del momento de esta decisión, no las vigentes, que están en
["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño). Los campos de lista
pasan de cinco a cuatro. La tabla de correspondencia con otros modelos se recalcula con un campo menos en el
total (["Compatibilidad de modelos"](../estado-del-arte/compatibilidad-de-modelos.md)). Y hay la
misma pérdida que se aceptó con `documentation`: al importar desde Backlog.md ya no queda constancia de
cuál de las tres listas era cada puntero, así que el conversor que reconstruya el formato de Backlog.md
tendrá que guardarlo por su cuenta, por ejemplo en etiquetas con ámbito, si quiere la ida y vuelta exacta. Se hace
ahora porque el contrato de estabilidad (["El contrato de estabilidad"](../spec/estabilidad.md)) obliga a
partir de la versión 1.0, que no se ha publicado: no hay ningún tablero ajeno con el campo escrito que
migrar.

**Una consecuencia que sí existe: los tableros de desarrollo con datos en el campo retirado.** Igual que con
`documentation`, la migración inicial se editó en el sitio sin subir `user_version`, así que un tablero
creado con un binario anterior y con datos en `modifiedFiles` queda parcialmente ilegible: `biso ls` avisa y
salta esas tareas, `biso get` falla con el error 3 y `biso doctor` las cuenta como errores. No hay
migración, porque `biso` no está publicado, y el único remedio que se ofrece es recrear el tablero.

---

## Se retira `ext`

**La decisión.** La tarea no tiene ningún campo de extensión. `ext` se retira de la especificación
entera, con sus flags `--ext`, `--rm-ext` y `--clear-ext`, su clave de configuración `extensions`, su
paso propio en el orden de aplicación de una escritura, sus cuatro identificadores de error
(`malformed_extension_key`, `unknown_extension_key`, `undeclared_extension_key` y `duplicate_ext_key`),
su comprobación de `biso doctor` y la página de campos externos. No hay ningún campo que lo sustituya:
un valor corto asociado a una clave se guarda como etiqueta con ámbito
(["Las etiquetas con ámbito"](#las-etiquetas-con-ámbito)), que sí se puede consultar.

**Qué cuesta, y qué rendía.** `ext` no rendía nada que se pudiera consultar: no entra en ningún cálculo,
no lo alcanza ninguno de los filtros de `biso ls` ni la búsqueda de texto, y solo se guardaba, se
imprimía en `biso get` y viajaba en la exportación. Costaba una tabla propia y con ella una de las cinco
consultas de una lectura completa del tablero, que es el camino caliente de `biso ls` y `biso prime`;
uno de los nueve pasos del orden de aplicación de una escritura; una de las veinte claves de
configuración, con su validación de vocabulario en uso, su flag en `biso init` y su presencia en la
simetría de `biso snapshot` con `biso init --from`; cuatro identificadores de error; una de las dos
formas `clave=valor` de la línea de comandos, la que se cortaba por el primer `=`, con una regla
contraria a la de `--set-comment-date`, que sigue existiendo y corta por el último; uno de los dos
alfabetos cerrados del programa; unas 158 líneas de Go fuera de pruebas y otras 155 dentro; quince páginas de la especificación que lo nombran; y 33 bytes de la parte
fija del mensaje de arranque. Una medición de 2026-09-21 sobre las 535 tareas de seis tableros de
Backlog.md no encontró ni una vez la necesidad que el campo decía cubrir. Backlog.md no tiene campo de
extensión, así que ese cero no prueba nada por sí solo: lo que decide es que el campo cuesta todo lo
anterior y no rinde nada consultable.

**Qué se pierde.** El valor de una etiqueta tiene el alfabeto cerrado de ["El juego de caracteres de un
token"](#el-juego-de-caracteres-de-un-token), sin espacios, sin `/`, sin `#` y sin una URL, mientras que
el de `ext` era texto libre de una línea. Ya no hay dónde guardar un texto libre por clave. Y `ext` era
la única puerta de salida de un modelo por lo demás cerrado: lo próximo que alguien quiera guardar en
una tarea, si no cabe en una etiqueta con ámbito, se decide y se especifica como campo. Se acepta.

**Descartado: dejarlo tal cual**, como un campo de constancia sin consulta. Seguiría pagando la lista
entera de costes para no rendir nada consultable.

**Descartado: abaratarlo**, con el mapa en una columna de texto de la tabla `task` y con el vocabulario
de claves abierto. La columna quita la tabla y una consulta del camino caliente, pero deja un segundo
mecanismo de clave y valor junto al de las etiquetas con ámbito. Y abrir el vocabulario quita la clave
`extensions`, su validación y parte de sus errores, a costa de que una clave mal escrita deje de
fallar en el momento, que es lo que el vocabulario cerrado existe para evitar y va contra la regla de
que en un campo de vocabulario cerrado un valor que no existe es siempre un error.

**Por qué ahora.** El contrato de estabilidad (["El contrato de estabilidad"](../spec/estabilidad.md))
obliga a partir de la versión 1.0, que no se ha publicado: retirarlo después costaría un ciclo de aviso
de al menos una versión menor, cuatro identificadores de error congelados que seguir emitiendo y una
clave de configuración que seguir leyendo. Hoy es gratis.

---

## No se adoptan identificadores de subtarea con punto

**La decisión.** Un identificador sigue siendo siempre `<PREFIJO>-<n>`, con `n` entero positivo
(["Identificador de tarea"](../spec/modelo-de-datos/identificadores.md)); `biso` no adopta la forma
`<PREFIJO>-<n>.<m>` que usa Backlog.md para nombrar una subtarea, ni al escribir ni al leer. Conservar
ese identificador de origen para el viaje de ida y vuelta es responsabilidad del conversor de
Backlog.md a `biso` (`tools/backlog.md-migrate`, TASK-70 y TASK-7): decidirá y documentará su propio
mecanismo, en su propio registro de decisiones, cuando se implemente esa tarea, sin comprometerlo
aquí.

**Por qué no es un cambio de borde.** La forma `<PREFIJO>-<n>` es un invariante que atraviesa el
programa entero: la gramática de referencias admite además las formas abreviadas `<n>` y `#<n>` sobre
ese mismo entero, `biso ls` ordena por él, el tablero guarda el identificador más alto que ha llegado a
asignar y ese dato aparece en los tres mensajes de "no la encuentro" y en `biso doctor`
(["Identificador de tarea"](../spec/modelo-de-datos/identificadores.md)), y ya hay un mensaje de error
dedicado a cerrarle la puerta a un identificador con punto
(["Los tres mensajes de \"no la encuentro\""](../spec/referencias.md#los-tres-mensajes-de-no-la-encuentro)).
Adoptar el punto habría abierto preguntas de diseño en ese mismo centro, como el orden entre `1.2` y
`1.10` o qué significa el identificador más alto asignado cuando `<n>.<m>` existe sin que exista `<n>`,
sin que ninguna tuviera una respuesta barata, mientras que resolverlo en el conversor no toca ninguna
de esas piezas.

**Qué se pierde, sin medir.** Una mención suelta a `TASK-56.1` escrita en prosa, dentro de una
descripción o de un mensaje de commit, deja de resolver como referencia: la gramática de
["Cómo se resuelve una referencia a una tarea"](../spec/referencias.md#la-gramática) no reconoce esa
forma, así que con `--id` es `malformed_id` y sin él cae en la búsqueda de texto. No hay ninguna medida
de que esa mención se use hoy: es una limitación que se acepta sin comprobar, no una que se sepa
barata, y queda pendiente de lo que se aprenda al escribir el conversor de TASK-70.

**Descartado: adoptar `<PREFIJO>-<n>.<m>` como identificador legal.** Es lo que pedía la tarea
original. Se descarta porque el coste cae en el centro del modelo, arriba, a cambio de una ganancia
que un conversor fuera de `biso` ya obtiene sin él.

---

## El orden manual es una clave de texto

**La decisión.** El orden manual de una tarea, `ordinal`, es una clave de texto opcional y no un número.
Se compone de los símbolos `0-9a-z`, no puede acabar en `0` y se compara por puntos de código, como
`--sort title`, de modo que entre dos claves cualesquiera siempre cabe otra y siempre hay una menor y una
mayor que cualquiera dada. No se puede teclear: se escribe con `--above <ref>` y `--below <ref>`, que
colocan la tarea justo antes o justo después de la vecina nombrada, y con `--ordinal first` y
`--ordinal last`, que la colocan delante o detrás de todas las que tienen clave, y se quita con
`--clear-ordinal`. La clave cruda solo viaja en el JSON (la salida de `biso ls` y `biso get`, la
exportación y el lote de `biso new --from`), y por eso la simetría entre `export` y `new --from` se
conserva. Cada tarea tiene una sola clave, global, y los grupos de una vista son solo presentación.

**El punto medio.** Entre dos claves se elige el centro del alfabeto y no el sucesor, para que las claves
crezcan despacio. El caso peor es meter siempre una tarea en el mismo hueco, y añade un carácter cada
cinco inserciones más o menos. La longitud no tiene tope y no hay renumerado: no hace falta, y si algún
día hiciera falta sería un comando nuevo, que es un añadido compatible.

**Las vecinas y los extremos.** `--above` sobre la primera tarea con clave y `--below` sobre la última no
son error, porque siempre hay una clave menor y una mayor. Una vecina sin clave es error 6, porque las
tareas sin clave van detrás de todas las que la tienen, ordenadas por urgencia, y "justo debajo de una"
no se puede cumplir escribiendo solo la clave de la tarea que se mueve; el mensaje propone
`biso set <vecina> --ordinal last` y luego `--below`. `--above`, `--below` y `--ordinal` son
incompatibles entre sí (error 2), y una tarea no puede ser su propia vecina (error 2). `biso set A B
--below C` deja `C`, `A`, `B`, y con `--above C` deja `A`, `B`, `C`. `--sort ordinal` es orden ascendente
por puntos de código con las tareas sin clave al final, como hoy, y el orden por defecto de `biso ls` no
cambia. La clave de una línea de lote se valida (el alfabeto y el `0` final) con un error propio en vez
del `invalid_number` del entero.

**Reordenar dentro de un grupo.** Como la clave es global, reordenar una tarea dentro de un grupo de
`biso board` escribe una clave entre las vecinas de ese grupo, y la posición global de la tarea puede
moverse como consecuencia. Es el precio de tener una sola clave, y es barato: en la medición de
2026-09-21 solo siete de las 535 tareas de seis tableros de Backlog.md tenían un reordenamiento manual de
verdad. Cruzar de un grupo a otro no reordena, sino que edita el campo por el que se agrupa (["Las
etiquetas con ámbito"](#las-etiquetas-con-ámbito)). Nada de esto está en la versión 1.0, porque
`biso board` queda fuera de ella y la línea de comandos no arrastra nada.

**Descartado: un entero con huecos.** Hace falta que algo asigne los valores espaciados, como hace
Backlog.md con sus múltiplos de 1000, y un paso de renumerado que reparta huecos iguales cuando se
agoten, que hoy no existe. Quien se encierre con enteros consecutivos solo puede ir tarea por tarea.

**Descartado: un decimal.** Un `float64` tiene 52 bits de mantisa, así que entre dos valores caben unas
52 bisecciones y ni una más, y el caso peor, mover siempre una tarea al mismo sitio, es el más común. No
elimina el renumerado, lo pospone y lo vuelve impredecible. Y la simetría entre `biso export` y
`biso new --from` es una prueba de la suite, y un `float64` necesita 17 dígitos significativos para
volver bit a bit, mientras que una cadena vuelve exacta sin especificar nada.

**Descartado: una clave de orden por eje de agrupación**, es decir, un mapa de clave de eje a clave de
orden, con el eje nombrado por un prefijo de etiqueta. Es la única que resuelve el problema de la
posición global en vez de aceptar su consecuencia, y no se descarta por ser una mala idea: Linear tiene
dos claves de orden por tarea (`subIssueSortOrder` y `sortOrder`, en ["Compatibilidad de
modelos"](../estado-del-arte/compatibilidad-de-modelos.md)). Se descarta ahora por lo que cuesta un mapa
por tarea, con claves que se quedan obsoletas en silencio cuando cambia el campo por el que se agrupaba
y un mapa que crece sin límite si los ejes son abiertos, y porque es un añadido estricto sobre esta
decisión: la clave global es exactamente la caída hacia atrás que necesitaría, así que pasar a ella
después de la 1.0 es añadir un campo y no cambiar un tipo.

**Descartado: el orden manual solo en la vista sin agrupar.** Dentro de un grupo mandaría el orden por
defecto y no se podría reordenar a mano.

**Descartado: teclear la clave cruda con `--ordinal <clave>`.** Nadie debería teclear `m8`, y aceptarlo
abre la puerta a claves mal formadas que hay que validar en un flag además de en el lote. `first` y
`last` cubren los extremos y el arranque de un tablero sin ninguna clave, y al ser palabras no chocan con
ningún símbolo del alfabeto.

**Descartado: `0-9A-Za-z` como alfabeto.** Gana menos de un bit por carácter frente a `0-9a-z`, a cambio
de una clave con mayúsculas que depende del sistema donde se copie.

**Descartado: dar clave a la vecina sin clave, o fijar la posición y avisar.** Lo primero escribe en una
tarea que no se nombró como destino de la escritura y la sube por encima de otras sin clave. Lo segundo
hace algo distinto de lo pedido y solo lo avisa.

**Por qué ahora.** Cambiar el tipo de `ordinal` es un cambio del contrato JSON y del nombre de un flag,
y el contrato de estabilidad (["El contrato de estabilidad"](../spec/estabilidad.md)) obliga a partir de
la versión 1.0, que no se ha publicado. Hoy es gratis.

---

## La ayuda enseña la dirección de una dependencia

**La decisión.** Hacia dónde apunta una dependencia se enseña en la ayuda de `biso set` y de
`biso new`, con una frase y un ejemplo con identificadores del proyecto de ejemplo: la arista se
escribe siempre en la tarea que espera, y `biso set MYP-10 --add-deps MYP-4` dice que `MYP-4` va
primero y bloquea a `MYP-10`. La misma ayuda glosa `--parent` y `--add-refs`, que no tienen glosa en
ningún otro comando de escritura (`--parent` solo la tiene como filtro de lectura en `biso ls`), y
`biso set --help` no afirma que ningún flag exija aprender nada más allá de su nombre, porque eso es
cierto para la forma de las listas y falso para el significado de los campos de relación. La dirección
en sí no cambia: es la de
["Las relaciones entre tareas"](../spec/modelo-de-datos/relaciones.md#dependencies-la-precedencia).
El mensaje de arranque lleva la misma enseñanza en una sola regla, la 11 de `RULES`: "A dependency is
written on the task that waits: `biso set MYP-10 --add-deps MYP-4` means MYP-4 blocks MYP-10.", con los
identificadores de la ayuda. Sigue sin glosar los flags: `FIELD FLAGS` es una rejilla de nombres y no
lleva ninguna explicación de lo que significan.

**Por qué hace falta.** Es el único de los errores posibles al escribir una relación que el programa no
puede detectar. Un `--parent` a una tarea que no existe, o una dependencia que cierra un ciclo, fallan
al escribirse; una dependencia escrita al revés es una dependencia válida, y todo el cálculo de
bloqueo y de urgencia queda invertido sin ningún aviso. Con `biso new` hay además un caso propio:
`--add-deps MYP-4` hace que `MYP-4` vaya antes de la tarea nueva, y para que la tarea nueva bloquee a
una existente hay que crearla y luego escribir la arista con `biso set` en la existente, cosa que la
ayuda de `biso new` dice.

**Descartado: dejar el mensaje de arranque sin nada de esto.** El argumento a favor es que la parte
fija tiene un tope propio
(["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)) y que el sitio natural
es la ayuda de cada comando. Se descarta porque el mensaje de arranque es lo único que lee quien empieza
a trabajar en un tablero, y ni sus reglas mencionan las dependencias ni `FIELD FLAGS` dice nada más que
nombres: quien no pide la ayuda de `set` antes de escribir una arista no tiene ninguna pista de la
dirección, y es justo el error que el programa no puede detectar. Costó 120 bytes: con el tablero del
ejemplo el mensaje medía entonces 4.983 bytes y pasó a 5.103, y la parte fija pasó de 3.494 a 3.614, con
un margen de 226 sobre sus 3.840; los tres tableros de las pruebas de presupuesto quedaban entre 5.103 y
5.278 de los 5.504. Son medidas del momento de esta decisión, no las vigentes: las de hoy están en
["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño).
El contrato de estabilidad no lo impide: congela el tope de bytes y no el texto de `biso prime`
(["El contrato de estabilidad"](../spec/estabilidad.md)).

**Descartado: glosar `--add-deps` dentro de `FIELD FLAGS`.** Rompe la rejilla, que es una lista de nombres
alineada para que se lea de un vistazo, y gastaría más bytes que una regla, porque habría que repetir la
explicación en la línea de cada flag de relación que la necesite.

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

---

## Un elemento vacío de un lote se descarta y avisa

**La decisión.** En un lote, `biso new --from` y `biso init --from`, un elemento de una lista que es una
cadena vacía o de solo espacios se descarta: no se guarda, no es un error y la tarea queda como si esa
posición no existiera. Es la misma regla del flag que añade (`--add-refs ""` no añade nada y avisa),
aplicada a cada elemento, y vale para todas las listas que un lote acepta: `assignees`, `labels`,
`dependencies`, `references`, el texto de cada criterio de `acceptanceCriteria` (dado como cadena o como
objeto) y las tres claves ajenas que se convierten al importar, `definitionOfDone`, `documentation` y
`modifiedFiles`. `biso new --from` avisa con `imported_empty_dropped`, una vez por cada línea y cada
lista de la que se descartó algo, con la cuenta; `biso init --from` no avisa, igual que no avisa de lo
que funde. Un `null` que ocupa el sitio de un elemento no es un elemento vacío sino un fallo de
validación de la línea, porque un elemento tiene que ser texto. Los elementos que no están vacíos se
guardan como llegan, sin recortar espacios, como hace el flag. La regla está en
["Un elemento vacío en un lote"](../spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote) y en
["El modo lote"](../spec/cmd/new.md#el-modo-lote).

**Para `labels`, `assignees` y `dependencies` es una relajación deliberada.** Un elemento vacío en
cualquiera de ellas era un fallo de la línea (una etiqueta mal formada, un asignado mal formado, una
dependencia que no está en el tablero ni en el fichero) y pasa a descartarse con aviso. Es un cambio
decidido, no un efecto secundario: un elemento vacío no dice nada, y lo que esas listas siguen
rechazando es un elemento que sí dice algo y lo dice mal.

**Los comentarios quedan fuera de esta decisión.** Esta regla cubre las listas de texto libre y de
tokens; los comentarios son objetos con un cuerpo y se tratan aparte. Lo que hacía la versión sin
corregir con un elemento vacío de `comments`, guardarlo tal cual sin descartarlo ni avisar, ya no es
cierto: tiene su propia regla, más estricta que esta, en
["Un comentario vacío o `null` en un lote es un fallo de validación"](#un-comentario-vacío-o-null-en-un-lote-es-un-fallo-de-validación).

**Por qué se descarta y no se rechaza.** La regla de que un valor que no existe es siempre un error
protege a quien lee una lista de una lectura falsa: un filtro mal escrito que devolviera vacío se
tomaría por un hecho sobre el tablero. Aquí no hay ningún valor que pudiera no existir, porque un
elemento vacío no dice nada: descartarlo no pierde un dato que el fichero afirmaba. Es lo que separa
este caso del de dos etiquetas exclusivas de la misma clave en una línea, que sí es un fallo porque
quedarse con una perdería un valor que el fichero decía
(["Las etiquetas con ámbito"](#las-etiquetas-con-ámbito)). Además, `labels` y `assignees` solo
tienen vocabulario cerrado al leer y no al escribir, y `references` y el texto de un criterio no
tienen ninguno, así que ninguna regla de vocabulario pide rechazar. Y una fuente habitual de un
elemento vacío es inocente, una cadena vacía que se parte por comas y deja un elemento, o un campo en
blanco de un formato de origen, y eso es lo que el flag ya tolera.

**Por qué avisa por línea y por lista.** El flag avisa por cada elemento vacío porque una línea de
comandos tiene pocos. Un lote puede tener cientos de líneas, y quien lo lee necesita saber dónde
está el problema del generador del fichero, no un aviso por cada hueco: la línea y la lista, con la
cuenta, bastan para encontrarlo. Nombra la línea y no la tarea por la misma razón que los avisos de
conversión, que el identificador puede no existir todavía.

**Por qué `init --from` no avisa.** Una restauración no emite avisos ni tiene un campo `warnings` en
su salida, y la instantánea que escribe `biso snapshot` de un tablero escrito con esta regla nunca
trae un elemento vacío, porque ningún camino del programa lo guarda. El único caso es una instantánea editada a mano o escrita por otro
programa, y ahí descartar un elemento que no dice nada es inocuo. Añadir un canal de avisos a la
restauración para ese caso no compensa. Lo de que `biso snapshot` no escribe nunca un elemento vacío
vale para un tablero escrito con esta regla, según el apartado siguiente.

**La simetría se mantiene.** Como el flag no guarda un elemento vacío y el lote lo descarta, ninguna
lista de un tablero escrito con esta regla lo contiene, así que `biso export` no lo escribe nunca y
reimportar la salida deja el mismo tablero campo a campo.

**Un tablero con vacíos guardados antes de esta regla.** El programa no ha tenido una versión
publicada que escribiera elementos vacíos, así que solo un tablero de desarrollo puede tenerlos
guardados; ese tablero los sigue exportando y se limpia con `biso export | biso new --from` sobre un
tablero vacío, o con `biso snapshot` y `biso init --from`, que los descartan al leer. `biso doctor` no
los detecta ni los repara.

**Alternativas descartadas, y por qué.**

- **Rechazar la línea con un error de validación** se descarta porque hace que el mismo contenido sea
  válido en la línea de comandos y inválido en un fichero, y porque haría fallar todo el lote, que se
  valida entero antes de escribir, por un hueco que no cambia el sentido de nada.
- **Guardarlo tal cual**, en `references`, `documentation`, `modifiedFiles` y el texto de los
  criterios, se descarta porque deja una tarea que ningún flag puede producir: la ficha de `biso
  get` muestra una línea de `refs` con huecos entre comas, y el flag `--rm-refs` no puede quitarlo, porque
  un valor vacío ahí es un error.
- **Descartarlo sin avisar** se descarta porque la importación cambia un dato del fichero, aunque sea
  uno que no dice nada, y el flag que añade ya avisa; sin aviso, el generador del fichero nunca se
  entera de que produce huecos.
- **Tratar `null` como un elemento vacío** se descarta porque un elemento de una lista es texto, y
  `null` no lo es, igual que un número no lo es y hoy es un fallo. Es también la extensión natural de
  la regla que ya rechaza `null` como lista entera, y aceptarlo dejaría dos maneras de escribir lo mismo
  donde el resto del formato solo admite una.
- **Un aviso por cada elemento descartado, como el flag**, se descarta por lo dicho arriba: un lote
  con miles de huecos inundaría stderr sin decir nada más que la cuenta y la línea.
- **Que `init --from` también avise** se descarta por lo dicho arriba, y porque exigiría un canal de
  avisos que la restauración no tiene.
- **Que `biso doctor` detecte y repare con `--fix` los vacíos ya guardados** se descarta porque ninguna
  versión publicada los escribió: solo un tablero de desarrollo los tiene y se limpia una vez con
  `export | new --from`, mientras que una comprobación nueva de `doctor` sería permanente para un caso
  que no puede darse en un tablero escrito con esta regla.

---

## Un comentario vacío o `null` en un lote es un fallo de validación

**La decisión.** En `comments`, a diferencia de `assignees`, `labels`, `dependencies`, `references` y
`acceptanceCriteria`, un elemento vacío no se descarta con aviso: la línea entera falla como
`invalid_line`, dentro del `batch_invalid` de siempre. Falla igual un objeto cuyo `body` está vacío o
es solo espacios que un elemento `null` en el lugar de un comentario; ninguno de los dos guarda nada
y ninguno se limita a avisar. La regla vive en
["El modo lote"](../spec/cmd/new.md#el-modo-lote) y en
["Un elemento vacío en un lote"](../spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote), que
remite aquí.

**Por qué un comentario no sigue la regla del elemento vacío.** Esa regla, de la entrada anterior,
vale para una lista de texto libre o de tokens, donde un elemento vacío no dice nada y descartarlo no
pierde ningún dato que el fichero afirmara. Un comentario no es eso: es un objeto con autor y cuerpo,
el registro de que alguien dijo algo, y un cuerpo vacío no es una forma de "nada que añadir" sino un
comentario al que le falta lo único que lo hace un comentario. La comparación que importa no es con
`references` sino con el título: un título vacío tampoco se descarta en silencio al importar un lote,
es el mismo `error: title cannot be empty` que da la línea de comandos
(["El modo lote"](../spec/cmd/new.md#el-modo-lote) trae el ejemplo `line 201: title cannot be
empty`). `comments` sigue esa misma familia y no la de las listas de texto libre.

**La misma regla ya vale en la línea de comandos, y esta decisión la extiende al fichero.** `--comment
""`, tanto en `biso set` y en `biso new` como en el propio `biso comment` (por su flag `--comment` o
por su texto posicional, que es la misma escritura escrita de otra forma), es siempre `error:
--comment cannot be empty`, código 2, nunca el aviso `empty_append` que sí reciben `--append-note` o
`--add-labels`. Antes de esta decisión el texto posicional de `biso comment` no seguía esa regla:
avisaba y no añadía nada, como si fuera un `--append-note`, en vez de fallar como ya hacía su propio
flag `--comment`. Es una incoherencia dentro del mismo comando, no solo entre comandos distintos, y
esta decisión también la cierra: las dos formas de escribir el mismo comentario dan ahora el mismo
error. Con la línea de comandos y el fichero de acuerdo, ningún camino del programa puede dejar
guardado un comentario sin cuerpo.

**La simetría se mantiene, y de forma más fuerte que en la entrada anterior.** Allí la simetría se
sostenía en que ningún flag guarda un elemento vacío y el lote tampoco. Aquí, además, ningún camino
puede siquiera crear un comentario sin cuerpo, así que ningún tablero llega a tenerlo, y la pregunta
de qué debería hacer `biso export` si alguna vez escribiera uno vacío no llega a plantearse: no hay
ninguna instantánea válida que lo contenga. Es la misma garantía que ya vale para `title`.

**Alternativas descartadas, y por qué.**

- **Descartar el elemento con aviso, como las demás listas.** Es la opción que la entrada anterior
  elige para `references` y compañía, y aquí se descarta por la razón de arriba: un cuerpo vacío no es
  un hueco inocente que se pueda quitar sin perder nada, es el único dato que hace de un objeto un
  comentario. Tratarlo como un hueco dejaría entrar por el fichero lo que la línea de comandos ya
  rechaza, exactamente la asimetría que se discutió y se descartó para el título.
- **Aceptarlo y guardar un comentario sin cuerpo**, que es lo que hacía la versión sin corregir: crea
  un `Comment` con `author: null` y `body: ""`. Se descarta porque ningún flag puede producir ese
  mismo objeto, así que sería una tarea que solo un fichero de importación puede dejar, y la ficha de
  `biso get` mostraría un comentario en blanco que `--rm-comment` sí puede borrar pero que nada más en
  el programa sabe explicar ni sabe producir.
- **Dejar el texto posicional de `biso comment` avisando, como hacía antes, y hacer fallar solo el
  flag `--comment` y el lote.** Mantendría la incoherencia dentro del mismo comando: la misma llamada,
  escrita de dos maneras, seguiría dando dos resultados distintos según cuál de las dos se usara. Se
  descarta por la misma razón que unifica el flag y el lote entre sí.
- **Un `code` propio para el cuerpo vacío**, en vez de reutilizar `invalid_line`. Se descarta porque
  `invalid_line` ya cubre un `null` de tipo equivocado en cualquier otra lista y una clave de
  comentario repetida: un cuerpo vacío es la misma familia de fallo, una línea que no se puede
  interpretar como una tarea válida del formato, y no una regla de vocabulario nueva que necesite
  distinguirse de las demás por su código.

---

## La ficha escapa la coma y la barra invertida de una lista

**La decisión.** La ficha de `biso get` escribe cada valor de una lista con la misma regla de escape
que se usa al guardarlo, y ninguna otra: una coma dentro de un valor sale como `\,`, una barra
invertida sale como `\\`, y el resto de los caracteres sale tal cual. Los valores de una línea siguen
separados por coma y espacio, y el separador es siempre una coma sin escapar, así que las
referencias `a,b` y `c.md` salen como `a\,b, c.md` y las referencias `a` y `b` como `a, b`. La regla es
exactamente la inversa de la de entrada de
["Repetición y listas separadas por comas"](../spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas):
para cualquier valor posible, salvo uno con un salto de línea, cuyo tratamiento se decide aparte,
deshacer el escape de lo que imprime la ficha da el valor guardado.
Vale para las cinco listas de la ficha (`assignees`, `labels`, `depends`, `blocks` y `refs`), pero solo
`refs` admite una coma o una barra invertida, así que solo su línea puede cambiar. `--json` no cambia:
cada valor es un elemento de la lista y nunca lleva escape. La regla está en
["`biso get`"](../spec/cmd/get.md#salida).

**Por qué se escapa también la barra invertida.** Escapar solo la coma no basta, porque entonces la
línea deja de decir sin ambigüedad qué valores hay. Las referencias `a\` y `b` se imprimirían como
`a\, b`, y la única referencia `a, b` se imprimiría igual, como `a\, b`: la barra final de la primera se
comería el separador. Duplicar la barra es lo que la regla de entrada ya define para que un valor pueda
terminar en barra invertida justo antes de una coma, así que no se inventa nada. La otra mitad de la
regla de entrada, que una barra delante de cualquier otro carácter se guarda tal cual, no obliga a
escribir esa barra sola en la salida: `C:\dir` se leería igual que `C:\\dir`, pero decidir cuándo
basta la barra simple obliga a mirar el carácter siguiente y el final del valor. Duplicar siempre es
una sola frase, sin excepciones, y se comprueba de un vistazo.

**Por qué la ficha y la entrada son simétricas.** La ficha es la única salida de texto que lista las
referencias de una tarea: `biso ls` y `biso prime` no las traen, y `biso export` y `biso snapshot`
escriben JSON. Quien la lee, sea una persona o un agente, la usa para volver a escribir un valor, por
ejemplo con `--rm-refs`, y si la ficha muestra `a,b` pero el flag necesita `a\,b`, copiar el valor
falla: `--rm-refs 'a,b'` intenta quitar `a` y `b` y avisa de que ninguna está. Con la misma regla en
los dos sentidos, lo que se ve es lo que se teclea, y un agente no tiene que aprender una notación
para leer y otra para escribir. Lo que se copia a un flag es el valor de una referencia, no la línea
entera: el espacio que sigue a cada coma separadora es de la ficha y no del valor, y `--rm-refs 'a\,b, c.md'`
entiende su segundo valor como ` c.md`, con el espacio, y avisa de que no lo tiene.

**Lo que no cambia.** Un aviso que cita un solo valor entre comillas (`warning: --rm-refs: "zz,y" not
present, nothing removed`) no es una lista y sigue con su propia forma de citar, la de una cadena entre comillas, que no es la
regla de la lista: la referencia `a,b` se cita `"a,b"` en el aviso y sale `a\,b` en la ficha, y la
referencia `C:\dir\x` se cita `"C:\\dir\\x"`, con las barras dobladas porque una cadena entre comillas
también las dobla, mientras que la ficha la escribe `C:\\dir\\x` por la regla de la lista. Que las dos
formas coincidan con las barras no las convierte en la misma regla: difieren en la coma. Una lista sin
valores sigue siendo un guion. Una referencia que fuera literalmente `-` sigue imprimiéndose como el
guion de una lista vacía, igual que antes, porque la regla de entrada no define ningún escape que la
distinga y una referencia así no es un caso que valga añadir una notación nueva.

**Alternativas descartadas, y por qué.**

- **Conservar la regla anterior, sin escapar y con `--json` como lectura exacta**, se descarta porque
  la ambigüedad no es cosmética. `a,b, c.md` se lee a simple vista como dos o como tres valores, y es
  la única forma de ver las referencias sin `--json`. Que un lector con un programa disponga de la
  lectura exacta no ayuda al que lee la ficha, y la ficha existe justo para ese lector. La ambigüedad
  tampoco era un coste que compensara nada: la regla nueva no añade una segunda vía de lectura, solo
  hace que la única sea inequívoca.
- **Cambiar el separador de esa línea** (`;`, ` | ` o cualquier otro) se descarta porque una
  referencia es texto libre, y una URL o una ruta puede llevar cualquiera de ellos, así que el problema
  solo se mueve. Y dejaría a `refs` con un separador distinto del de las demás listas de la misma
  ficha.
- **Imprimir cada referencia en su propia línea** se descarta porque el último renglón del bloque de
  metadatos pasaría de ser una línea fija a un número variable de ellas, con lo que el bloque
  dejaría de tener la forma que esta especificación le fija, y quien lee la ficha con un programa
  tendría que saber dónde termina la lista. Además no evita todo escape: una referencia con un
  salto de línea seguiría rompiendo el formato.
- **Entrecomillar cada valor** (`"a,b"`) se descarta porque obliga a definir un segundo escape, el de
  la comilla, junto al de entrada, y a que el lector conozca las dos notaciones.
- **Escapar la coma pero no la barra invertida** se descarta por lo dicho arriba: deja dos
  listas distintas con la misma línea.
- **Escapar la barra invertida solo donde hace falta** (delante de otra barra, delante de una coma o al
  final de un valor) se descarta porque la regla depende de lo que rodea a cada barra, dos
  implementaciones razonables pueden discrepar en un borde, y lo único que ahorra es una barra doble
  en una ruta de Windows, un caso poco frecuente en una lista de referencias que casi siempre son rutas
  con `/`, URLs o identificadores.

---

## Una referencia con un salto de línea se rechaza al escribirla

**La decisión.** Una referencia que contiene un `\r` o un `\n` literal se rechaza al escribirla, con
el mismo error que ya usan `title`, `author` y el texto de un criterio: `malformed_string_value`,
código 2, con `field` igual a `reference`. La comprobación se aplica en `Task.Validate()`, el único
punto por el que pasan las tres vías que pueden dejar un salto de línea en una referencia, `--add-refs`,
`biso new --from` y `biso init --from`, así que un solo cambio basta para las tres. No se sustituye
el salto de línea por otra cosa, ni se escapa al imprimirlo en la ficha: una referencia con un salto de
línea, sencillamente, no llega a guardarse. Ningún otro campo de lista necesita esta misma regla:
`labels` y `assignees` ya la excluyen porque su alfabeto cerrado (letras, dígitos y los símbolos
`- _ . : @`, ["El juego de caracteres de un token"](#el-juego-de-caracteres-de-un-token)) no admite
un salto de línea entre sus caracteres válidos, y `dependencies` siempre se resuelve a un identificador
de tarea válido antes de guardarse (["Las relaciones entre tareas"](../spec/modelo-de-datos/relaciones.md#dependencies-la-precedencia)),
así que tampoco hay ningún camino por el que un salto de línea sobreviva hasta el almacén.

**Por qué.** `references` es del tipo `list<string>`
(["El modelo de datos de una tarea"](../spec/modelo-de-datos/index.md#el-modelo-de-datos-de-una-tarea)),
y cada uno de sus elementos es, por tanto, un `string`: el mismo tipo que ya declara, sin excepción,
que "ningún `string` admite un salto de línea literal"
(["El salto de línea en un campo `string`"](../spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)).
Antes de esta decisión, `references` era el único campo de lista de la especificación sin ninguna
restricción de caracteres, precisamente porque su contenido es una URL o una ruta y no admite el
alfabeto cerrado de `labels` y `assignees`
(["El juego de caracteres de un token"](#el-juego-de-caracteres-de-un-token)); pero un salto de línea
no es un carácter que una URL o una ruta necesiten, así que rechazarlo no deja fuera ningún caso
legítimo, y aplicar aquí la misma regla que ya rige el resto de los `string` de la tarea es la opción
más consistente con lo que la especificación ya afirmaba. Tampoco hace falta un tercer mecanismo de
escape junto a los dos que ya usa la ficha de `biso get` para una lista, la coma y la barra invertida
(["La ficha escapa la coma y la barra invertida de una lista"](#la-ficha-escapa-la-coma-y-la-barra-invertida-de-una-lista)):
rechazar el valor en el punto de entrada quita el problema de raíz, en vez de arrastrarlo hasta la
salida y pedirle a la ficha que lo disfrace.

**Alternativas descartadas, y por qué.**

- **Sustituir el salto de línea por otra cosa** (un espacio, por ejemplo) al guardar la referencia. Se
  descarta porque el programa no cambia en silencio el valor que alguien escribió en ningún otro campo
  de texto libre: `title`, `author` y el texto de un criterio se rechazan, nunca se corrigen solos.
  Sustituir aquí rompería esa garantía, y nadie podría predecir, sin leer el código, en qué se convirtió
  el texto que tecleó.
- **Escaparlo en la ficha** (imprimirlo como la secuencia literal `\n`, por ejemplo). Se descarta porque
  añadiría un tercer carácter de escape a la línea `refs`, junto a la coma y la barra invertida que ya
  usa, y solo arreglaría la lectura por `biso get`: el valor seguiría siendo guardable por
  `--json` y por el lote, dejando en el almacén un `string` con un salto de línea que contradice la
  definición de ese tipo en el resto del modelo de datos.

---

## Una tarea ilegible es la misma para todos los comandos de lectura

**La decisión.** Una tarea es ilegible por una sola definición, la de la lista completa de
["Qué se comprueba"](../spec/garantias.md#qué-se-comprueba): `status`, `type` y `priority` con un valor
que la configuración no declara (comparado letra por letra, y con `status` obligatorio), cada campo de
fecha que no es una fecha o que es obligatorio y está vacío, un nombre de campo de lista desconocido y
una columna cuyo contenido no es del tipo del modelo. La definición se aplica al tablero entero al
leer, antes de mirar los filtros de la llamada, y todos los comandos de lectura la comparten. `biso ls`
y `biso prime` saltan la tarea, la cuentan y avisan (`prime` dentro de su mensaje). `biso get` de esa
tarea sale con código 3 y el motivo exacto. `biso export` y `biso snapshot` escriben las tareas
legibles, no copian la ilegible, y salen con código 6. `biso doctor` la reporta como error y sigue.
Una escritura dirigida a ella es error 3 sin escribir nada, salvo que lo que escribe la deje legible,
que es el remedio de un valor fuera de vocabulario. El aviso nombra todas las tareas ilegibles del
tablero, casen o no con los filtros. Las personas y las etiquetas no entran: son abiertas al escribir y
solo cerradas al leer, contra el conjunto en uso, del que un valor guardado forma parte por definición.
La regla está en ["Qué pasa con un dato que no se puede interpretar"](../spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)
y en la página de cada comando de lectura. No cambia el mensaje de arranque: la línea `unreadable` de
`biso prime` ya existía y no se toca ni una palabra de su parte fija.

**El problema que la motiva.** Antes de esta regla la única comprobación de vocabulario al leer era un
efecto secundario del cálculo de la urgencia, que mira la prioridad y ningún otro campo, y solo
ocurría en los comandos que calculan la urgencia. Sobre un tablero con la base de datos editada a mano,
cada campo daba un resultado distinto:

| Valor guardado | Lo que hacían los comandos |
|---|---|
| Una prioridad que el tablero no declara | `ls` la saltaba con aviso, `get` fallaba y `doctor` la reportaba, pero `export` y `snapshot` la volcaban entera con código 0. El `snapshot.ndjson` resultante lo rechazaba `biso init --from` con el error 7 de la línea inválida |
| Un estado que el tablero no declara | Desaparecía de `ls`, de `export`, de `snapshot` y del recuento de `prime` sin ningún aviso, porque el estado por defecto y el de `--any-status` son los configurados, mientras que `get` la mostraba y `set` la escribía. Es exactamente la lista incompleta que se confunde con un tablero sin esa tarea |
| Un tipo que el tablero no declara | `ls` la listaba con un tipo que `ls --type` rechaza como inexistente, y `export` la volcaba |
| Una fecha que no es una fecha | Ya era coherente: `ls` la saltaba, `get` fallaba, `export` y `snapshot` escribían el resto y salían con 6 |
| Una fecha obligatoria vacía | Pasaba en silencio en todos los comandos |
| Una columna del tipo equivocado | Abortaba todas las lecturas con el código 1, aunque la tarea fuera una sola |
| Una prioridad fuera de vocabulario, con un filtro que no dejaba pasar a la tarea | El aviso no salía, porque el salto ocurría después de filtrar: `ls --status Done` decía `no tasks match` sin más |

**Por qué una sola definición.** La regla del proyecto es que en un campo de vocabulario cerrado un
valor que no existe es un error, se esté escribiendo o leyendo, y que un resultado incompleto nunca
puede parecer un hecho sobre el tablero. Con cada comando decidiendo por su cuenta, la misma tarea
salía en un listado y no en otro, o en ninguno sin decirlo. Una definición única, evaluada una vez al
leer, hace que la pregunta "¿qué comandos ven esta tarea?" tenga la misma respuesta para todos, y que
añadir un comando de lectura no obligue a decidir otra vez qué hace con un dato dañado.

**Por qué el estado, el tipo y la prioridad entran y las etiquetas y las personas no.** El criterio es
si un filtro valida el campo contra la configuración o contra lo que hay guardado. `--status`,
`--type` y `--priority` validan contra la configuración, así que solo en ellos un valor guardado puede
estar fuera de lo que el filtro acepta, y una tarea que se lista con un tipo que `ls --type` rechaza
contradice la regla de que el mismo texto vale lo mismo al escribir y al leer
(["El mismo texto vale lo mismo en los dos sentidos"](../spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos)).
`--label` y `--assignee` validan contra el conjunto de valores en uso, y un valor guardado siempre
pertenece a él, así que ninguna lectura puede contradecirse. Que la lista `labels` de la configuración
prohíba una etiqueta guardada es una incoherencia que `doctor` reporta, pero la etiqueta se lee y se
filtra sin ambigüedad, y volver ilegible la tarea por eso la sacaría de los listados por un dato que
se puede mostrar bien.

**Por qué el estado en concreto no puede quedarse en un aviso de `doctor`.** Los papeles de los
estados (inicial, activo y terminal) se calculan comparando el estado de la tarea con los configurados.
Una tarea con un estado que no es ninguno de ellos no está viva ni terminada: no tiene significado, y
todo lo derivado (cuántas hay en cada estado, qué bloquea, qué es lo siguiente) sería una suposición.

**Por qué la comparación es exacta.** Cada escritura resuelve lo tecleado a la grafía configurada, así
que lo guardado tiene siempre esa grafía. Un valor con otra capitalización o con guiones no lo produjo
el programa, y como el estado de una tarea se compara con el terminal letra por letra para saber si
está terminada, aceptar `done` donde el estado es `Done` la contaría como viva. Usar la normalización de
`coincidir()` sería aceptar un dato que el resto del programa no reconoce.

**Por qué `export` y `snapshot` escriben lo legible y salen con 6.** Son los comandos cuyo propósito es
no perder nada, y son los que más importa que no entreguen una copia que el programa no sabría leer
después. Escribir solo lo legible cumple las dos cosas: la copia se importa siempre entera con
`biso init --from` o con `biso new --from`, y el código 6 más el aviso dejan a la vista que falta algo.
Y no se niegan a escribir porque un valor de fecha malformado no tiene remedio dentro de `biso`:
ningún flag escribe esas fechas y la fila no llega a cargarse para editarla, así que una negativa
detendría las copias durante todo el tiempo que dure un daño que el propio programa no puede arreglar,
que es justo cuando más falta hacen. La tarea ilegible no viaja en la copia ni siquiera cuando su dato
se podría leer, como una prioridad desconocida, porque una línea que el importador rechaza convertiría
la restauración entera en un error.

**Por qué `get` sale con 3 y no con 6.** La tabla de ["Códigos de salida"](../spec/codigos-de-salida.md)
reserva el 3 para "un dato guardado que no se puede interpretar", que es exactamente lo que dice
`get`, y el 6 para una operación válida cuyo resultado el estado del tablero no permite satisfacer por
completo, que es lo que le pasa a `export` y `snapshot`, que sí producen algo. Es la misma frontera que
ya separaba a `get` de `export` para una fecha malformada, extendida a los demás campos.

**Por qué una escritura que deja la tarea legible se aplica.** Un valor fuera de vocabulario es un
daño que un comando del propio programa puede arreglar, y exigir una vía de reparación aparte
(`doctor --fix`) chocaría con la razón por la que `--fix` no decide por quien llama qué valor quiso
decir (["Para qué sirve `biso doctor`, y para qué no"](../spec/cmd/doctor.md#para-qué-sirve-biso-doctor-y-para-qué-no)).
Juzgar la tarea como quedaría tras la escritura, y no como está, deja a `biso set MYP-2 --priority
medium` arreglarla y a `biso set MYP-2 --title X` fallar, sin ninguna excepción nueva que enseñar:
es la misma regla de validación que ya se aplica a toda escritura.

**Por qué el aviso nombra todas las tareas ilegibles.** Una tarea ilegible no se puede comparar con un
filtro, y no se puede afirmar que no lo cumple. `ls --status Done` que imprime solo `note: no tasks
match` sobre un tablero con una tarea ilegible dice algo que nadie sabe, que es la lista incompleta
que se toma por un tablero vacío y que la especificación prohíbe.

**Consecuencias aceptadas.**

- Una tarea que depende de una ilegible no consta como bloqueada por ella: la ilegible no participa en
  el cálculo, igual que ya pasaba con una fecha rota. El aviso la nombra en la misma llamada.
- El `snapshot.ndjson` que se escribe mientras hay una tarea ilegible no la lleva, y la versión
  anterior de esa tarea solo sigue en la historia del repositorio del tablero.
- Una fecha malformada o una columna del tipo equivocado no se arreglan con ningún comando: se corrige
  la base de datos a mano o se recupera la tarea de una instantánea anterior.
- **`biso doctor` no reporta `task_unreadable` para `ordinal` ni para el nombre de un campo de lista
  cuando el valor guardado viola una restricción `CHECK` que el esquema de la base de datos sigue
  declarando, porque su comprobación de integridad general encuentra antes esa misma violación y
  aborta el comando entero con el código 21** de
  ["Qué pasa con un dato que no se puede interpretar"](../spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar),
  sin llegar a construir el informe. Esa situación concreta solo se produce escribiendo el valor
  mientras se ignora el `CHECK` por fuera de cualquier operación de `biso`, por ejemplo con `UPDATE`
  bajo `PRAGMA ignore_check_constraints=1`, dejando la declaración intacta en el esquema
  (`internal/store/migrations.go`). No se cambia el código porque ese caso concreto solo se alcanza
  corrompiendo el fichero por fuera del programa, y sobre esa misma base de datos `biso ls`, `biso
  get` y `biso export` ya dan el diagnóstico exacto con el trato normal de tarea ilegible, porque
  ninguno de los tres corre la comprobación de integridad al abrir el tablero. Cuando en cambio el
  propio esquema deja de declarar el `CHECK`, por ejemplo en un tablero de una versión anterior a él,
  el mismo valor guardado no viola nada de lo que la comprobación de integridad verifica, y `biso
  doctor` sí reporta `task_unreadable` con normalidad: es el caso que ya ejercita
  `cmd/biso/ordinal_unreadable_test.go` para `ordinal`.

**Alternativas descartadas, y por qué.**

- **Que `export` y `snapshot` no escriban nada y salgan con 6** se descarta porque el remedio de una
  fecha malformada no está en `biso`: la negativa dejaría sin copias, y sin historia, un tablero cuyas
  demás tareas están sanas, hasta que alguien edite la base de datos a mano. Escribir lo legible con
  el código 6 protege lo mismo (la copia no se toma por completa) sin ese coste.
- **Que copien la fila tal cual, con o sin aviso** se descarta porque la copia dejaría de poder
  leerse: `biso init --from` rechaza una línea con un valor fuera de vocabulario, y una fecha rota no se
  puede ni escribir. Es el comportamiento anterior para las prioridades, y es lo que la tarea llamaba
  una copia que el propio programa no sabrá leer.
- **Que salgan con 0 y aviso**, como una lectura de conjunto cualquiera, se descarta porque un guion
  que encadene `biso snapshot && ...` no vería el aviso, que va por stderr, y la copia parecería
  completa. Es la razón que ya tenía la excepción del 6 (["Comandos y flags"](comandos-y-flags.md)).
- **Que la comprobación de vocabulario sea cosa de `doctor` y una tarea ilegible sea solo una fila
  corrupta** se descarta porque deja a `ls` listando lo que `ls --type` rechaza y haciendo desaparecer
  sin aviso las tareas con un estado desconocido. Es lo que pasaba, y se trata de una lista incompleta
  que se confunde con un hecho sobre el tablero.
- **Comparar con la normalización de `coincidir()`** se descarta por lo dicho arriba: aceptaría valores
  que ninguna escritura produce y que el cálculo de estados no reconoce.
- **Que `get` imprima la ficha de una tarea con un valor fuera de vocabulario y un aviso** se descarta
  porque `get` es donde se lee el motivo exacto de que una tarea no se lea, y una ficha con un estado
  sin significado calcularía la urgencia y los bloqueos con un valor inventado. Tampoco se puede
  extender a las fechas, donde la ficha no se puede ni construir, y un comando con una regla para cada
  motivo sería justo lo que esta decisión quita.
- **Que el aviso nombre solo las tareas ilegibles que el filtro habría dejado pasar** se descarta
  porque una ilegible no se puede comparar con el filtro, y porque el aviso dependería de qué filtros
  se escribieran.
- **Volver ilegible una tarea con una etiqueta prohibida por la lista `labels`, una dependencia rota o
  un arrendamiento a medias** se descarta porque esos datos se leen y se muestran bien, `doctor` ya
  los reporta, y esconder la tarea de los listados por ellos sería más daño que el que denuncian.
- **Dejar que `doctor --fix` repare las tareas ilegibles** se descarta porque un valor fuera de
  vocabulario tiene más de un arreglo válido, y elegir uno destruye información; y una fecha
  malformada no tiene ninguno que el programa pueda decidir. Es exactamente el criterio con el que
  `doctor` deja fuera de `--fix` lo que necesita una decisión o viene de un daño externo.
- **Contar la tarea ilegible en los recuentos de `biso prime`, con su estado guardado** se descarta
  porque su estado puede ser justo lo ilegible, y la línea de recuento pasaría a describir tareas que
  el mensaje no supo mirar, que es lo que ya decía `prime` de las tareas saltadas.

---

## El cierre transitivo del grafo de dependencias

**La decisión.** El flag nuevo de `biso get` que imprime el cierre transitivo de `dependencies` se
llama `--closure`. Imprime el grafo completo, con las tareas terminadas y archivadas incluidas, en las
dos direcciones (["`biso get`"](../spec/cmd/get.md#--closure)); los dos recuentos que sí van siempre
en la ficha, sin el flag, cuentan solo lo que sigue sin terminar (`blocked by`/`unblocks` en texto,
`blockedByCount`/`unblocksCount` en JSON). Solo `biso get`, sobre una tarea, lo calcula: ni
`biso ls` ni las cuatro listas de `biso prime` lo llevan, salvo la única línea con nombre propio de la
primera fila de `NEXT UP` (["`biso prime`"](../spec/cmd/prime.md#la-salida-literal)). Un ciclo de
dependencias o una dependencia que apunta a una tarea que no existe no hacen fallar ni colgarse el
recorrido: lo primero se recorre sin repetir un identificador ya visitado y sin avisar de nada, lo
segundo simplemente no añade ningún identificador al cierre.

**Por qué `--closure` y no `--transitive` ni `--graph`.** `--transitive` es un adjetivo sin sujeto: no
dice qué es lo que se vuelve transitivo sin leer la ayuda, mientras que "cierre" (`closure`) es el
sustantivo exacto del concepto de teoría de grafos que el flag calcula, el mismo que usa la
descripción de esta misma decisión. `--graph` se descarta porque sugiere una vista del grafo entero
del tablero, y el flag es estrictamente el cierre de una sola tarea, la que se pidió con `<ref>`; un
nombre que prometiera el tablero entero decepcionaría a quien lo probara. Un tercer candidato,
`--explain-blocking`, con el mismo prefijo que `--explain-urgency`, se descarta porque "blocking" y
"blocked" ya nombran los dos términos booleanos de la fórmula de urgencia
(["La urgencia"](../spec/modelo-de-datos/urgencia.md#la-urgencia)): reutilizar ese vocabulario para un
recuento transitivo distinto invitaría a confundir el término de la fórmula, que sigue siendo
booleano, con el cierre completo, que no lo es.

**Por qué la lista que imprime el flag es el grafo completo y el recuento no.** Las dos cifras
existen para responder preguntas distintas: el recuento responde "¿cuánto de esto sigue bloqueando de
verdad, hoy?", y por eso cuenta solo lo que sigue sin terminar, con la misma definición que ya usan
los términos `bloquea` y `bloqueada` de la urgencia. El flag responde "¿cómo es el grafo alrededor de
esta tarea?", una pregunta de auditoría a la que una tarea ya terminada le sigue perteneciendo: sigue
siendo verdad que depende de ella, o que ella depende de otra, aunque ya no bloquee nada. Filtrar
también la lista habría dejado dos cifras que siempre miden lo mismo, el tamaño de la lista y el
recuento, y no habría ninguna razón para tener las dos.

**Por qué el recorrido no señala un ciclo ni avisa de una dependencia rota.** `biso doctor` ya
diagnostica los dos casos, con sus propios `code` (`dependency_cycle` y `dependency_not_found`,
["`biso doctor`"](../spec/cmd/doctor.md#qué-comprueba)), y repetir el diagnóstico en cada
`biso get --closure` que pase por la tarea dañada sería ruido y no información nueva, además de una
segunda fuente de verdad sobre el mismo hecho. `--closure` es una vista de lectura, así que se limita
a dar una respuesta correcta y terminar: ignora el identificador dañino y sigue con el resto del
grafo, que es la respuesta más útil a "qué desbloquea esto" cuando parte del grafo está corrupto, en
vez de negarse a contestar nada.

**Por qué solo `biso get` lo calcula, y ninguna fila de `biso ls` ni de `biso prime`.** El coste de un
cierre transitivo es proporcional al tamaño de la cadena de dependencias de una tarea, no una
constante, y `biso ls` imprime hasta 300 tareas de una vez: repetirlo fila a fila multiplicaría ese
coste por cada una, sin que nadie lo hubiera pedido, justo lo que prohíbe la primera regla de
["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque). `biso get` lee una
sola tarea, así que paga el coste de una sola cadena; `biso prime` paga el de una, la primera fila de
`NEXT UP`, y ninguna más, por la misma razón.

**Alternativas descartadas, y por qué.**

- **Que el cierre en sí, no solo el recuento, se filtre a tareas sin terminar**, como ya hace `blocks`
  con sus dependientes directos, se descarta por la razón de arriba: habría dejado la lista y el
  recuento siempre iguales, sin ganar nada por tener los dos.
- **Que un ciclo se señale con una nota por stderr o con una marca en la lista** se descarta porque
  `biso doctor` ya es el sitio con nombre propio para ese diagnóstico, y una vista de lectura que
  avisara de daño en el almacén cada vez que lo atraviesa duplicaría esa responsabilidad sin
  añadir ninguna información que `doctor` no dé ya, más despacio y en cada llamada en vez de una sola
  vez.
- **Que `biso ls --json` y las cuatro listas de `biso prime --json` lleven `blockedByCount` y
  `unblocksCount` en cada tarea**, para simetría con `blocks` o `urgency`, se descarta por el coste:
  son las dos únicas claves de un objeto de tarea que no valen una constante, y añadirlas a un listado
  de 300 tareas sería el mismo trabajo que 300 llamadas a `biso get`, en una sola invocación que hoy
  cabe en el presupuesto de 25 milisegundos.
- **Que `biso prime` lleve la línea de desbloqueo transitivo en más de una fila de `NEXT UP`**, o en
  las de `IN PROGRESS`, `NEEDS ANSWER` y `ASSIGNED TO YOU`, se descarta por la misma razón de coste:
  cada fila adicional es un cierre transitivo más que nadie pidió, y el mensaje de arranque ya tiene
  su propio tope de tamaño (["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)) que una línea por fila agotaría mucho antes.
