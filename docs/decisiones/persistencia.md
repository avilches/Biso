# La persistencia

## La decisión de persistencia

La especificación dejaba deliberadamente abierto cómo se guardan los datos. La decisión es: un tablero
es una base de datos SQLite en un directorio propio fuera del proyecto, localizado por un fichero
puntero versionado con el proyecto (`.biso.json`, ["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md)), con una exportación de texto
que sí se guarda en el control de versiones para el historial (`biso snapshot`, ["`biso snapshot`"](../spec/cmd/snapshot.md)). Sin daemon, y sin fusionar nunca dos almacenes escritos por separado.

La evidencia detrás de cada pieza de esta decisión, con sus enlaces, está en
["Estado del arte"](../estado-del-arte/index.md), el inventario de las herramientas del espacio y el
catálogo de sus fallos. Lo que sigue aquí es el porqué de cada pieza, no la evidencia en bruto.

**Por qué el tablero no se versiona con el código.** Cualquier herramienta que guarde las tareas como
ficheros del árbol de trabajo hereda su peor propiedad: el estado se bifurca con la rama, así que una
incidencia cerrada en una rama vuelve a aparecer abierta al volver a la principal (["El estado de las tareas se bifurca con la rama"](../estado-del-arte/catalogo-de-problemas.md#2-el-estado-de-las-tareas-se-bifurca-con-la-rama)). Sacar el tablero del árbol de trabajo no resuelve ese problema, lo disuelve:
una tarea cerrada está cerrada, no cerrada en esta rama, porque no hay una rama que la contenga. El
precio es que el tablero no viaja al clonar el proyecto en otra máquina, y se paga a propósito a cambio
de que el estado de una tarea sea uno solo.

**Por qué no hay daemon, y por qué el motivo es aritmético y no de gusto.** El coste dominante de una
invocación de `biso` es arrancar un proceso, no el trabajo que hace una vez arrancado: ["El coste de arranque y el coste de contexto"](../estado-del-arte/catalogo-de-problemas.md#12-el-coste-de-arranque-y-el-coste-de-contexto)
mide el suelo del sistema en 5,2 milisegundos y el total de leer, ordenar e
imprimir 300 tareas en 8,7 milisegundos con un binario de Go, suelo incluido y leyendo un JSON en vez de
la base de datos (["El origen de la cifra de 25 milisegundos"](#el-origen-de-la-cifra-de-25-milisegundos)). Un daemon solo puede ahorrar lo que hay por encima del suelo, no el
suelo, porque el cliente que hablaría con él por un socket es también un proceso y paga el mismo suelo de
arranque para lanzarse. Así que un daemon competiría por uno o dos milisegundos de
unos ocho, pagando a cambio una arquitectura entera: un proceso de fondo que hay que arrancar, vigilar y
matar, y que si se cuelga hace fallar también las lecturas que ["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) promete que
nunca fallan por una escritura en curso. Beads tuvo uno, hacía una sola cosa, y se eliminó por completo
al cambiar de motor; quien lo reemplazó por algo más simple cuenta que se pasaba varias veces por semana
peleándose con él (["El proceso de fondo"](../estado-del-arte/catalogo-de-problemas.md#11-el-proceso-de-fondo)).

**Por qué el texto es una salida, y nunca un canal de vuelta.** `biso snapshot` escribe `snapshot.ndjson` y
`board.json` para que el historial de git cuente lo que pasó y para que `biso init --from` pueda
reconstruir el tablero entero en otra máquina, pero nada dentro de `biso` vuelve a leer esos ficheros
como si fueran la verdad. Beads documenta por qué esa asimetría es obligatoria y no una elección
estética: su importación es solo de inserción y actualización, y no puede saber si un registro ausente
en el texto fue borrado a propósito o simplemente no se llegó a exportar (["El almacén binario no se versiona, y el texto no tiene transacciones"](../estado-del-arte/catalogo-de-problemas.md#10-el-almacén-binario-no-se-versiona-y-el-texto-no-tiene-transacciones)). Tratar el texto como una fuente además de como una salida reintroduce esa
ambigüedad en `biso`, así que no se hace nunca: la base de datos es la única verdad, y `export` y
`snapshot` son su proyección de solo lectura hacia fuera.

**Por qué nunca se sincroniza fusionando dos almacenes escritos por separado.** Es el sitio donde se
han estrellado todas las herramientas del espacio, de formas distintas pero con la misma raíz: dos
copias de trabajo que asignan el mismo identificador a tareas distintas (["Dos copias de trabajo asignan el mismo identificador"](../estado-del-arte/catalogo-de-problemas.md#1-dos-copias-de-trabajo-asignan-el-mismo-identificador)), un bloqueo de fichero que no cruza remotos de git (["Un bloqueo que no cruza máquinas, o que se queda huérfano"](../estado-del-arte/catalogo-de-problemas.md#4-un-bloqueo-que-no-cruza-máquinas-o-que-se-queda-huérfano)), y un fichero
de log que se fusiona por unión de líneas y resucita las que se habían borrado, porque una fusión de
texto concatena y solo quita duplicados exactos, sin razonar sobre qué falta ni por qué (["El almacén binario no se versiona, y el texto no tiene transacciones"](../estado-del-arte/catalogo-de-problemas.md#10-el-almacén-binario-no-se-versiona-y-el-texto-no-tiene-transacciones)). La
respuesta seria de ese último problema, sustituir el motor por uno con fusión a nivel de celda, es la que
tomó Beads, y es coherente pero cara. `biso` no la necesita porque no la tiene que resolver: un tablero
vive en una sola máquina y no hay una segunda copia escribible con la que fusionarse, así que la
comprobación de identidad de ["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) basta y no hace falta un algoritmo de fusión.

**Por qué `--fix` no es una comodidad, sino el consentimiento.** Esta decisión añade a `biso doctor`
(["`biso doctor`"](../spec/cmd/doctor.md)) las comprobaciones que no existían antes de que hubiera una base de
datos real detrás del tablero: la integridad de esa base de datos, y el aviso de un sistema de ficheros
donde el modo WAL de SQLite no da las garantías de atomicidad que ["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) exige.
Las dos son daño externo puro, porque nada dentro de `biso` corrompe su propia base de datos ni decide
en qué disco vive el tablero, y por eso ninguna de las dos es reparable ni con `--fix`: la integridad se
repara restaurando de una copia, fuera de `biso` por completo, y el sistema de ficheros no es algo que
la herramienta pueda cambiar. Que aparezcan justo con esta decisión confirma la regla que ya ordenaba el
resto de `doctor`, en vez de ponerla a prueba: `--fix` es el único sitio de todo `biso` donde quien
llama dice "te autorizo a escribir cosas que no te he pedido una por una", y solo entra ahí lo que de
verdad se puede arreglar sin decidir por alguien; lo que no, se reporta y se deja donde está.

**Por qué la búsqueda del puntero no frena en la raíz de un repositorio, ni sabe que git existe.** El
puntero se busca subiendo desde el directorio de trabajo, y lo primero que se escribió fue que el
recorrido paraba en la raíz del proyecto, detectada buscando un directorio `.git` hacia arriba. La idea
era proteger de que un proyecto encontrara el tablero de otro que lo contuviera. Se retira, y conviene
saber por qué, porque parece una protección gratis y no lo es.

El primer motivo es que hace que la resolución del tablero dependa de una herramienta que este mismo
documento declara opcional. ["`biso snapshot`"](../spec/cmd/snapshot.md) dice que git puede no estar instalado y que
un tablero funciona igual sin él, y a la vez el freno haría que la respuesta a cuál es mi tablero
saliera de si existe cierto directorio que crea git. El segundo es peor, porque no es de principios sino
de comportamiento observable: es `biso snapshot` quien convierte el directorio del tablero en un
repositorio, de forma perezosa y la primera vez que se ejecuta, así que con el freno puesto el mismo
comando en el mismo directorio contestaba una cosa antes de la primera instantánea y otra después, sin
que nadie hubiera cambiado ninguna configuración. El tercero es que el freno estaba mal escrito de una
forma que solo se ve al ir a implementarlo: en un worktree de git y en un submódulo, `.git` no es un
directorio sino un fichero, de modo que la raíz de un worktree nunca se habría detectado y el recorrido
se habría pasado de largo hasta el repositorio que lo contiene. Como en esta máquina los worktrees viven
dentro del propio checkout, el efecto habría sido que un `biso` ejecutado en un worktree resolviera el
puntero del checkout principal.

Y el cuarto motivo es que el freno no protegía de lo que decía proteger. El caso que preocupaba es un
proyecto anidado en otro que sí tiene tablero, y ahí el puntero del proyecto de fuera está en el mismo
directorio donde el freno habría parado, así que el freno llega tarde y el puntero se hereda igual. La
única regla que de verdad lo evitaría es respetar el `.gitignore` del proyecto de fuera, porque en el
caso real que motivó la discusión ese fichero excluye la carpeta del proyecto de dentro. Eso está
descartado por ["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque): interpretar un `.gitignore` de verdad no
se puede reimplementar de forma fiable, y preguntárselo a git cuesta 12 milisegundos medidos de los 25
que hay para todo.

**Lo que se acepta a cambio, dicho claro.** Un proyecto sin puntero propio hereda el del proyecto que lo
contenga, si lo hay. Se acepta porque es una situación poco frecuente, porque cuando ocurre lo razonable
es que se note en vez de que se adivine, y porque tiene un remedio que no necesita ninguna regla nueva:
mover el tablero de fuera a un subdirectorio que no esté en la línea de subida, que es exactamente lo
que hubo que hacer con Backlog.md en esta máquina por el mismo motivo. Que se note es cosa de dos
salidas que ya existen: el bloque `BOARD` del mensaje de arranque dice el nombre del tablero en su
primera línea, así que un agente que empieza la sesión con `biso prime` ve enseguida cuál le ha tocado,
y `biso where` dice en su fila `source` el directorio concreto del que salió el puntero. Se descartó
emitir además una nota en cada comando: saldría también en el caso normal de trabajar desde un
subdirectorio del propio proyecto, que es la inmensa mayoría de las llamadas, y una nota que sale
siempre enseña a ignorarla.

**Por qué estar dentro del directorio del tablero es una vía de resolución y no un error.** Antes de
decidir lo de arriba se consideró lo contrario, que un comando ejecutado dentro del directorio de un
tablero fallara siempre, con el argumento de que ese directorio es almacenamiento y no un proyecto. Se
descarta porque ahí no falta ningún dato: la configuración completa de un tablero, con su nombre y su
`task_prefix`, vive dentro de su propia base de datos, y el puntero solo sirve para encontrar un tablero,
cosa que quien ya está dentro de él no necesita. Un directorio que contiene `board.db` es ese tablero y
no puede ser otro, así que no hay nada que adivinar, y es un hecho más específico que cualquier puntero
heredado, de donde sale que gane al puntero en el orden de ["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md). Esta vía se lleva por delante
el otro motivo que tenía el freno de git para existir, porque el caso que hacía falta proteger ahora se
resuelve solo.

**Por qué el nombre `board.db` es interfaz y no un detalle interno.** La vía de arriba necesita
reconocer un directorio de tablero, y sin un nombre declarado no habría una sola forma de hacerlo. El
documento hablaba del fichero de la base de datos sin nombrarlo nunca, lo que bastaba mientras nada
dependiera de reconocerlo desde fuera.

**Por qué la identidad del tablero vive en un fichero y no en el nombre de su carpeta.** El diseño
anterior decía dos cosas que no podían ser verdad a la vez: que el nombre de la carpeta del tablero era
decorativo y que nadie resolvía por él, y que localizar un tablero desde su puntero era buscar el patrón
`*-<id>` en la raíz por defecto. Lo segundo es resolver por el nombre. De esa contradicción salían tres
fallos, y los tres se cerraron de golpe sacando el identificador del nombre y metiéndolo en un fichero
marcador, `<id>.id`, dentro del directorio del tablero.

El primero: `biso config set project_name` movía la carpeta para mantener el slug al día, y prometía no
tocar ningún puntero "porque el tablero se localiza por el patrón". Eso solo valía para los tableros de
la raíz por defecto, que es el único sitio donde se buscaba el patrón. Un tablero en cualquier otra parte
se quedaba con un puntero que nombraba una carpeta ya inexistente y sin ninguna red debajo. El segundo:
la comprobación de `biso doctor` que avisaba cuando el nombre de la carpeta no coincidía con el slug
denunciaba como problema un tablero perfectamente sano creado con `--at`, cuyo nombre de carpeta lo había
elegido quien llamaba, y `--fix` lo "reparaba" renombrando la carpeta, con lo que rompía el puntero que
`init` acababa de escribir. El tercero: `boards_extra_roots` estaba declarada en la configuración de
máquina y ninguna regla la leía, porque la búsqueda nombraba solo `boards_root`.

Con el marcador, el nombre de la carpeta es decorativo de verdad. Renombrarla no rompe nada, cambiar el
nombre del tablero **no toca el sistema de ficheros en absoluto**, la comprobación de `doctor` sobre el
nombre desaparece porque ya no hay nada que comprobar, y la búsqueda por identificador puede recorrer
cualquier raíz mirando marcadores, lo que da sentido a las raíces adicionales. **El premio grande es lo
que se quita**: renombrar un tablero era la única operación del programa que mezclaba una transacción de
SQLite con un movimiento en el sistema de ficheros, cosa que no puede ser atómica, y por eso arrastraba
un orden declarado, un estado intermedio observable y un código de error propio. Nada de eso existe ya.

El coste es que buscar un tablero por su identificador pasa de leer un nombre de carpeta a mirar dentro
de cada carpeta de cada raíz. Sigue sin abrir ninguna base de datos, así que son lecturas de directorio,
pero con veinte tableros son veintiuna en vez de una, y hay que medirlo contra ["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque) cuando el programa exista. Se descartó a propósito el atajo de buscar primero por el nombre
y caer al marcador solo si falla: sería más rápido y volvería a poner el mismo dato en dos sitios, que es
lo que se acaba de quitar.

**Por qué la `path` del puntero puede ser relativa, y por qué la forma la elige quien llama.** Un puntero
que dice `/Users/avilches/Hub/Projects/My project/tablero` solo resuelve en el ordenador donde el proyecto está
en esa ruta exacta, y el puntero se versiona precisamente para que viaje. Con `tablero` guardado como
ruta relativa, mover el proyecto entero con su tablero dentro no rompe nada.

La ruta relativa **se resuelve contra el directorio que contiene el fichero puntero, nunca contra el
directorio de trabajo.** El puntero se busca subiendo desde el directorio de trabajo, así que el mismo
fichero se lee desde cualquier subdirectorio del proyecto: resolver contra el directorio de trabajo haría
que el mismo puntero nombrara un tablero distinto por cada subdirectorio desde el que se llamara, y que
casi ninguno de ellos existiera. Eso rompería el principio de ["Los principios"](../spec/principios.md) de que ningún
comportamiento depende de dónde se ejecute el programa.

**La primera redacción de esto se apoyaba en un escenario que la desmentía, y conviene dejarlo escrito
para no repetirlo.** Decía que una ruta absoluta no resuelve "ni siquiera en un worktree del mismo
proyecto", y es al revés. Un worktree de git recibe el puntero, porque está versionado, y no recibe el
directorio del tablero, porque el propio `biso init` recomienda ignorarlo y los ficheros ignorados no se
comparten entre árboles de trabajo. Así que en un worktree la ruta absoluta sigue apuntando al tablero
real y la relativa apunta dentro del worktree, donde no hay nada. El cambio empeoraba justo el caso con
el que se justificaba. Y el desenlace era peor que el fallo: el mensaje que sale entonces invita a
ejecutar `biso init` para adoptar ese identificador, lo que desde un worktree habría creado un segundo
tablero con el mismo `id`, que es lo único que esta persistencia promete no permitir nunca. El error de
razonamiento fue tratar un worktree como el proyecto en otra ruta: lo es para lo versionado, pero el
tablero no está versionado y sigue existiendo en un solo sitio del disco.

De ahí salen las dos piezas que lo arreglan. Una es que la búsqueda pruebe una `path` relativa que no
resuelve contra cada ancestro del directorio del puntero, que es lo que encuentra el tablero del proyecto
desde un worktree que viva dentro de él. La otra es que **la forma de la ruta la elija quien llama**, con
la forma que le dé a `--at`. Se había descartado dar esa elección con el argumento de que quien llama no
tiene ningún dato que el programa no tenga, y ese argumento era falso: el dato que decide es dónde van a
vivir las demás copias de trabajo del proyecto, y eso solo lo sabe una persona. Con los worktrees dentro
del proyecto, que es la convención de esta máquina, la relativa es la buena; con los worktrees fuera del
proyecto, la única que resuelve es la absoluta, porque el directorio del tablero no está ni en la copia ni
en ninguno de sus ancestros. No hace falta ninguna bandera nueva para ofrecer esa elección, porque la
forma de `--at` ya la expresa.

**Por qué guardar siempre la ruta absoluta se consideró y se descartó.** Es tentador, porque una ruta
absoluta contiene más información que una relativa y por tanto se puede degradar: si no resuelve, todavía
queda su último componente para probarlo contra el directorio del puntero, con lo que hasta mover el
proyecto se sobreviviría, y no habría dos formas ni ninguna elección que explicar. Se descartó por lo que
cuesta en el fichero que se commitea: una ruta absoluta lleva al repositorio el nombre de usuario y la
estructura de directorios de quien ejecutó `init`, y eso queda en el historial de un fichero que otras
personas ven. La elección se conserva porque es de verdad del usuario, y el coste de conservarla es una
nota y una regla de búsqueda, no un mecanismo.

**Y por qué la nota es una nota y no un aviso.** ["Notas y avisos"](../spec/salida-y-terminal.md#notas-y-avisos) reserva `warning:` para lo
que está mal hecho o es arriesgado, y su tabla de avisos es cerrada. Guardar una ruta relativa no está mal
hecho: es la elección correcta para quien tiene sus worktrees dentro del proyecto, que es el caso normal.
La nota está redactada como hecho más consecuencia, y no como el consejo de "considera usar rutas
absolutas", porque un consejo obliga a quien lo lee a averiguar si le aplica, mientras que decir qué se ha
guardado y qué no va a funcionar hace que quien trabaja así se reconozca y el resto pueda seguir.

**La condición de la nota nombra dos cosas y las dos hacen falta.** Una copia de trabajo que vive fuera
del proyecto se queda sin el directorio del tablero solo porque el proyecto lo ignora, que es una de las
dos salidas que la otra nota de `init` describe. Al escribirlo apareció el caso contrario, que no estaba
cubierto: si nadie ignora esa carpeta, la copia de trabajo recibe el marcador y los dos ficheros de texto
pero nunca `board.db`, porque el fichero de exclusión que `init` escribe dentro del tablero lo excluye
siempre. Queda un directorio que parece el tablero y no lo es, así que la resolución exige las dos cosas,
el marcador y la base de datos, y sigue buscando cuando falta la segunda.

**Y esa otra nota dejó de recomendar y pasó a contar las dos salidas**, porque al mirar el caso contrario
se vio que no era el accidente que la primera redacción suponía, sino la configuración que hace que la
instantánea cruce a otra máquina sola: un tablero versionado dentro del proyecto viaja con el remoto que
el proyecto ya tiene. Recomendar ignorar la carpeta habría sido recomendar renunciar a eso sin decirlo.

**Por qué el tope de la búsqueda es el directorio personal y no un número de niveles.** La primera
redacción decía que el recorrido no comprueba ningún directorio con menos de dos componentes de ruta, y
funcionaba, pero por casualidad: acierta solo mientras el directorio personal tenga esa profundidad. Con
un directorio personal en `/root`, que es el del superusuario, el tope habría caído por debajo de él y un
puntero puesto ahí no se habría leído nunca. La regla dice ahora lo que quiere decir, que el recorrido no
sale del directorio personal de quien llama, y guarda el número solo para el caso en que no haya
directorio personal que lo exprese, con el trabajo fuera de la home o sin `HOME` definido. El primer
componente de una ruta absoluta es siempre un directorio del sistema o el contenedor de los directorios
personales de todo el mundo, así que un puntero ahí no puede estar a propósito.

**Por qué se retiran la bandera `--board` y la variable `BISO_BOARD`.** Eran las dos primeras de las
cuatro vías por las que ["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md) encontraba un tablero, y la tabla de banderas
globales las resumía como "Usa ese tablero directamente, sin buscar", con un valor que ["Banderas globales"](../spec/cmd/flags-globales.md#banderas-globales)
describía como "el nombre o el localizador de un tablero, en la forma que el almacenamiento imponga".
Esa vaguedad se sostenía mientras la persistencia estuviera sin decidir. Con la persistencia ya decidida
había que contestar si el valor era un nombre, un identificador de ocho hexadecimales o una ruta, sobre
qué raíces buscaba y qué pasaba cuando encajaba con dos tableros de la máquina. La respuesta no es
elegir una de las tres formas, ni partir la bandera en tres: es que la bandera ya no hace falta. El
razonamiento tiene cuatro patas, y las cuatro hacen falta.

**La primera: lo que `--board` prometía ya lo daba `-C`.** La bandera `--cwd`, con su forma corta `-C`,
dice "resuelve el tablero desde ahí, sin cambiar el directorio del proceso", y la resolución que arranca
desde ese directorio ya reconoce las dos cosas que alguien querría nombrar. Si el directorio contiene
`board.db`, la vía del directorio de trabajo dice que el tablero es ese y no se busca nada más. Si el
directorio es el de un proyecto, la vía del puntero lee su `.biso.json` ahí o en cualquier ancestro. Así
que `-C <directorio del tablero>` y `-C <directorio del proyecto>` cubren juntos todo lo que la bandera
retirada podía nombrar sin ambigüedad. Y para fijarlo durante una sesión entera, que era el único uso
propio de `BISO_BOARD` frente a la bandera, ya está `BISO_CWD`, que ["Variables de entorno"](../spec/invocacion.md#variables-de-entorno) declara equivalente a
`--cwd` con la bandera ganando.

**La segunda: por qué no se sustituye por una bandera que acepte el nombre del tablero.** Es la pata que
más falta va a hacer, porque nombrar el tablero por su nombre es lo primero que se le ocurre a
cualquiera. El obstáculo es que **el nombre de un tablero no está en el sistema de ficheros**. Lo que hay
en el disco es una carpeta llamada `<slug>-<id>`, y ese trozo legible es decorativo a propósito:
["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md) dice que nadie resuelve nunca por él, y ["`biso config`"](../spec/cmd/config.md) dice que cambiar `project_name`, que es el
nombre del tablero, no toca el sistema de ficheros en absoluto. De ahí que el slug pueda ser el de un
nombre anterior y no haya nada que lo corrija, ni falta que hace. El nombre de verdad vive dentro de la
base de datos, junto con el resto de la configuración. Resolver por nombre sería, entonces, abrir la base
de datos de cada carpeta de cada raíz para preguntarle cómo se llama, y eso choca de frente con
["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque), que da 25 milisegundos para todo: con veinte tableros son veinte
aperturas de SQLite antes de empezar a hacer el trabajo que se ha pedido, y eso es exactamente el trabajo
que nadie ha pedido que la primera regla de esa misma sección prohíbe. La salida evidente, guardar el
nombre en el fichero marcador para poder leerlo sin abrir ninguna base de datos, reintroduce justo lo que
el rediseño de la identidad acababa de quitar: que renombrar un tablero vuelva a escribir en el sistema
de ficheros, con lo que vuelve el mismo dato en dos sitios y la posibilidad de que discrepen. Y aunque
todo eso saliera gratis, el nombre no sirve para elegir, porque no es único: ["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md) declara legal
que dos proyectos de la misma máquina se llamen igual, y lo declara como una de las tres cosas que
resuelve de golpe sacar la identidad del nombre, precisamente porque lo que identifica a un tablero es su
`id` y no cómo se llama.

**La tercera: por qué tampoco una bandera que acepte el identificador.** Aquí no hay ningún obstáculo
técnico, hay algo peor: no hay ningún caso de uso que se sostenga. El identificador de ocho hexadecimales
no aparece en el trabajo diario, porque las tareas se nombran con `<PREFIX>-<n>` y ese prefijo se deriva
de `project_name`, no del `id` del tablero (["Identificador de tarea"](../spec/modelo-de-datos/identificadores.md#identificador-de-tarea)), así que nadie lo tiene delante ni lo teclea. El
único momento en que el `id` manda es cuando el puntero lo trae y su `path` no resuelve, y ahí la
búsqueda **ya recorre sola** la raíz por defecto y las raíces adicionales mirando el marcador `<id>.id` de
cada carpeta, sin que nadie tenga que pasar ninguna bandera. El otro caso imaginable, que un mensaje de
error te enseñe un identificador y quieras usarlo, es precisamente aquel en el que ese tablero **no está
en esta máquina**, que es lo que el error `pointer_unresolved` dice con todas las letras: ninguna bandera
alcanza un tablero que no existe en el disco.

**La cuarta: la garantía que se pierde, y por qué no valía una bandera para conservarla.** Hay una
diferencia real entre las banderas, y conviene no disimularla. `--board` prometía usar ese tablero
directamente, sin buscar, mientras que `-C` sí sube por los ancestros, de modo que apuntar con `-C` a un
directorio equivocado puede terminar en silencio en el tablero del proyecto que lo contenga. Parece un
argumento para conservar la bandera, y se cae solo en cuanto se lee ["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md) entera, porque ese
riesgo ya está aceptado en el caso general: "El precio de no tener ese freno es que un proyecto sin
puntero propio hereda el del proyecto que lo contenga, si lo hay, y se acepta a propósito". Añadir una
bandera cuyo único valor fuera evitar en un caso concreto un riesgo que la especificación acepta a
propósito en el caso general sería incoherente, y encima daría la impresión de que el caso general está
protegido cuando no lo está. Lo que sí hace visible el caso, y ya existe, es `biso where`, cuya fila
`source` nombra el directorio del que salió el puntero, y el bloque `BOARD` del mensaje de arranque, que
dice el nombre del tablero en su primera línea.

Cómo eligen su almacén las demás herramientas del espacio, y las de fuera de él, está en
["Cómo se apunta a un almacén distinto del que la herramienta encuentra sola"](../estado-del-arte/herramientas.md#cómo-se-apunta-a-un-almacén-distinto-del-que-la-herramienta-encuentra-sola),
que cierra la parte 1 de "Estado del arte". El resumen es que ninguna acepta el nombre legible de un
almacén para elegirlo, que todas apuntan con una ruta, y que las dos que sí admiten un nombre lo hacen
en una bandera aparte de la de la ruta y contra un registro previo que lo declara.

**Y la consecuencia que esto deja, que conviene ver antes de tocar nada.** Con las dos vías retiradas
quedan dos, y la primera, la que reconoce un directorio como tablero porque contiene `board.db`, deja de
ser una comodidad y pasa a ser **la única forma que queda de tocar un tablero al que el proyecto actual
no apunta sin escribir antes en el disco**, siempre para una sola invocación y con `-C` apuntando a su
directorio. Cuando se decidió, se justificó solo con que ahí no hay nada que adivinar; ahora carga además
con este trabajo. Se nota en el puntero perdido, donde ["Cómo se elige el tablero"](../spec/resolucion-del-tablero.md) ofrecía dos remedios, `--board`
apuntando al tablero o `biso init --at`, y ha quedado con uno solo, que escribe un puntero nuevo en el
proyecto. Quien en el futuro quiera endurecer esa vía, por ejemplo exigiéndole también el marcador
`<id>.id` como hace la búsqueda por identificador, tiene que saber que estaría cerrando la última puerta
que queda, y que la propia sección la deja abierta a propósito para que `biso doctor --fix` pueda
devolver un marcador que falte.

### El control de versiones, la instantánea y los códigos que salieron de ahí

Todo este apartado sale de una sola pregunta que la especificación tenía mal contestada: ["Lo que se deja fuera a propósito"](../spec/fuera-de-alcance.md)
decía que lo que cruza a otra máquina es la instantánea, y ningún comando le daba una vía para
cruzar, porque el repositorio donde vive está fuera del proyecto y no tiene remoto.

**Por qué el sistema de control de versiones es configurable, y no git a secas.** Cablear git habría
dejado sin historial a cualquiera que use otro sistema, y no por una limitación real: lo único que `biso`
necesita de él son cuatro operaciones, ver si un directorio está en un repositorio, crear uno, guardar una
revisión y publicarla. La clave `vcs` de ["Configuración de máquina"](../spec/invocacion.md#configuración-de-máquina) las nombra, con `git` por defecto
porque es el dominante y el único medido, `none` para no ejecutar nada y `custom` para el resto. El
catálogo existe porque un sistema conocido permite decir cosas que un comando opaco no puede: el
identificador de la revisión, que no había nada que guardar, y en qué repositorio ha acabado. Con `custom`
el contrato se reduce a lo único honesto, ejecutar la orden y mirar su código de salida, y la salida en
JSON lo refleja llevando `null` donde no puede saber.

**Por qué la clave vive en la configuración de la máquina y no en la del tablero.** Dice qué herramienta
hay instalada aquí, no cómo es un tablero, y esa diferencia tiene una consecuencia concreta: la
configuración del tablero es la que viaja en la instantánea, así que puesta ahí, restaurar la instantánea
de otra persona le impondría el sistema del ordenador de origen. Es el mismo error que la clave `me`, que
en esta misma ronda se sacó de la instantánea por exactamente esa razón.

**Por qué la revisión va donde vive el tablero, y por qué hay que preguntar por la exclusión.** Un tablero
puede estar en tres situaciones, y las tres tienen los mismos ficheros en el mismo sitio: ser su propio
repositorio, estar dentro del repositorio del proyecto, o no estar en ninguno. Lo único que distingue la
segunda de tener el directorio ignorado es una línea en el fichero de exclusión del proyecto, así que
adivinarlo mirando el disco es imposible y `snapshot` lo pregunta (["`biso snapshot`"](../spec/cmd/snapshot.md)). El caso que
justifica el trabajo es el bueno: cuando el proyecto versiona la carpeta del tablero, la revisión va al
repositorio del código y la instantánea cruza a otra máquina con el proyecto, sin que nadie configure un
remoto. Leída al pie de la letra, la redacción anterior habría creado ahí un repositorio dentro de otro
repositorio, que es el peor de los resultados posibles.

**El riesgo que se acepta con `--vcs push` en ese caso.** Publicar el repositorio del proyecto arrastra
también los commits de código que estuvieran pendientes en esa rama. Se acepta porque es lo que la bandera
promete y porque quien la escribe ya está pidiendo publicar: negarse a hacerlo, o hacerlo a medias, sería
sorprender a quien pidió una cosa clara. La alternativa que se descartó, error de uso en ese caso, rompía
un guion que llame igual desde varios proyectos configurados de formas distintas.

**Por qué la instantánea guarda tres ficheros y no dos.** El marcador `<id>.id` entra en la revisión
porque es lo que hace que la identidad del tablero viaje, y de eso depende que el puntero commiteado del
proyecto siga valiendo después de restaurar. La redacción anterior se contradecía: decía que el marcador
quedaba versionado y a la vez que la revisión añadía solo los dos ficheros de datos. Se nombran los tres
uno a uno, y no el directorio entero, para que un fichero que alguien deje ahí a mano no acabe en el
historial.

**Por qué `snapshot` no toma ningún acceso exclusivo.** Es una lectura, y ["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables)
promete que una lectura nunca hace fallar a una escritura. Si tomara el acceso exclusivo de las
escrituras, una copia podría hacer terminar con error a un `biso set` que llegara a la vez, que es un daño
sobre el trabajo diario. Sin él, el único desenlace malo es que dos instantáneas simultáneas choquen al
guardar la revisión, y ese daño cae sobre una copia que se puede repetir: sale con el código 8 y su
mensaje dice que basta volver a llamar. Un acceso exclusivo propio de este comando evitaría también ese
choque, y se descartó por lo que cuesta especificar bien un fichero de bloqueo persistente y la limpieza
de los que deja atrás un proceso que muere, para un caso que solo ocurre si dos sesiones terminan en el
mismo segundo sobre el mismo tablero.

**Por qué el daño de la base de datos estrena el código 21 en vez de compartir el 20.** El 20 promete un
remedio, `biso init` crea el tablero, y con la base de datos dañada ese remedio no arregla nada: hay que
reconstruir desde una instantánea. Tres situaciones con tres remedios no pueden compartir número si el
principio de ["Los principios"](../spec/principios.md) dice que quien llama ramifica sobre el número sin leer el mensaje.
Y como el remedio ahora tiene un comando que lo hace, el mensaje lo nombra en vez de decir "restaura de
una copia": `biso init --from` reconstruye en el sitio, adoptando el `id` del marcador, y para eso hubo
que declarar que un directorio cuya base de datos no abre no cuenta como tablero accesible.

**Por qué el identificador duplicado estrena el 22 en vez de reusar el 5.** El 5 es la referencia que
encaja con más de una entidad, y en la práctica siempre habla de una tarea: quien lo recibe afina la
referencia. Aquí no hay ninguna referencia que afinar, hay dos directorios en el disco con la misma
identidad, y el remedio es renombrar o quitar uno. Compartir el número habría obligado a leer el mensaje
para saber cuál de los dos remedios aplicar, que es justo lo que los códigos existen para evitar.

**Por qué las columnas se miden en celdas de terminal.** Los títulos son texto libre en UTF-8, y contar
puntos de código desalinea la tabla en cuanto aparece un acento combinante, un ideograma o un emoji,
porque lo que suman en pantalla no es lo que suman como caracteres. La celda es la única unidad que
alinea de verdad, y medirla no contradice la prohibición de mirar el terminal de
["Interactividad, terminal y color"](../spec/salida-y-terminal.md#interactividad-terminal-y-color): la anchura de un carácter es una propiedad de Unicode, igual en cualquier máquina, mientras
que lo que esa sección prohíbe es preguntarle a la ventana cuántas columnas tiene. El recorte del título usa la
misma unidad y no parte nunca un grafema, así que la promesa es un tope de 100 celdas y no una longitud
exacta.

---

### Cómo se ejecuta el sistema de control de versiones

Cerrada la ronda que hizo configurable el sistema de control de versiones, quedaban tres rincones del
contrato de `vcs_custom` sin decidir: qué se hacía con lo que las órdenes escribieran, si había un tiempo
máximo de espera, y si `{files}` daba rutas relativas o absolutas.

**Lo primero que salió al mirarlos fue que dos de los tres no eran de `custom`.** Le pasan igual a `git`:
si `git commit` falla porque un hook lo rechaza o porque nadie ha configurado una identidad, la pregunta
de qué se hace con lo que ese `git` escribió es exactamente la misma. Así que las reglas se escribieron
una sola vez, en un apartado de ["`biso snapshot`"](../spec/cmd/snapshot.md) que vale para los dos sistemas, y `custom`
las hereda. Sale más corto que un contrato paralelo por sistema y deja menos rincones donde volver a
decidir lo mismo.

**La salida se reenvía siempre, y por stderr.** Las tres opciones eran descartarla, reenviarla solo al
fallar y reenviarla siempre. Descartarla deja a quien depura sin saber por qué su hook rechazó el commit,
que es justo el caso en el que alguien va a leer esa salida. Reenviarla solo al fallar pierde el aviso
legítimo de una orden que acaba bien, y si algo escribe un programa ajeno, `biso` no está en posición de
juzgar si sobra. Se eligió reenviarla siempre, y **reenviar las dos corrientes de la orden, no solo la de
error**, porque `git commit` escribe su resumen por la estándar y `git push` su progreso por la de error:
quedarse con una sola pierde la mitad de lo útil. El precio es que el orden relativo entre las dos
corrientes no se puede garantizar, y el documento lo dice en vez de prometer algo que no se cumple.

Por stdout no va nunca nada, porque ["stdout, stderr y qué va en cada uno"](../spec/salida-y-terminal.md#stdout-stderr-y-qué-va-en-cada-uno) lo reserva para los datos. Y con `--json`
tampoco va por stderr, donde ["Los errores en JSON"](../spec/contrato-json.md#los-errores-en-json) pone el sobre de error: ahí las líneas entran en
el propio sobre, en `data.vcsOutput` o en `error.vcsOutput`. Eso añadió una quinta clave de detalle a la
tabla de esa sección, que ya se gobierna aparte del contrato de estabilidad de las salidas de
datos precisamente para poder crecer.

**`--quiet` no las suprime**, por el mismo motivo por el que ["Notas y avisos"](../spec/salida-y-terminal.md#notas-y-avisos) nunca suprime un `warning:`.
`--quiet` calla las líneas `note:` porque las escribió `biso`, que sabe que son trivia; de una línea que
escribió un programa ajeno no puede saberlo. Quien quiera silencio tiene el `2>/dev/null` que esa misma
sección ya nombra.

**La salida de las dos preguntas de la receta sí se descarta, y es la única asimetría.** Apareció al
escribir la regla: `git rev-parse --show-toplevel` falla cuando no hay ningún repositorio y
`git check-ignore` termina distinto de cero cuando la carpeta no está ignorada, y las dos cosas son
respuestas normales, no fallos. Reenviarlas habría puesto un `fatal: not a git repository` alarmante en el
camino normal de cualquier tablero que viva fuera de un repositorio, que es el caso más común. Se reenvía,
por tanto, lo que escriben las órdenes que actúan, y no lo que escriben las que preguntan.

**No hay tiempo máximo de espera, y es una decisión, no un olvido.** La alternativa era un tope, uno solo
o uno por orden, y se descartó por dos razones. La primera es que `biso` no puede interrumpir con
seguridad una orden que está a medio escribir en un repositorio ajeno: matarla a mitad es peor que
esperarla. La segunda es que un `push` legítimo contra un repositorio grande por una red lenta tarda
minutos, así que cualquier tope lo bastante corto para proteger de un cuelgue rompería un uso normal. Y lo
que de verdad cuelga una orden para siempre no es que sea lenta, es que espere una entrada que nadie va a
dar: eso se resuelve ejecutándola con la entrada estándar cerrada y sin terminal, con lo que falla en vez
de esperar. Matar el proceso a mano tampoco cuesta nada, porque los dos ficheros de la instantánea ya
están escritos antes de que se ejecute la primera orden.

**`{files}` da rutas relativas al directorio del tablero.** Las absolutas funcionan aunque la orden se
cambie de directorio por su cuenta, pero meten la ruta de una máquina concreta en cualquier sitio donde la
orden la escriba, como un mensaje de commit o un registro, y hacen que la misma configuración no valga en
dos máquinas. Se eligieron relativas, sin `./` delante, que es además lo que ya hace la receta de `git`, y
quien necesite absolutas puede envolver su orden en un script. Se descartó una clave nueva para elegir
entre las dos, por no añadir una decisión que nadie ha pedido.

**Y al escribirlo apareció un error del documento.** ["Configuración de máquina"](../spec/invocacion.md#configuración-de-máquina) decía que `{files}` se sustituía por
"los ficheros de la instantánea", pero la revisión lleva tres ficheros (`snapshot.ndjson`, `board.json` y
el marcador `<id>.id`) mientras que la clave `files` del JSON de `biso snapshot` enseña solo los dos que
ese comando escribe. Eran dos conjuntos distintos con nombres casi iguales, y no había manera de saber si
`{files}` eran dos o tres. Son los tres, los mismos que entran en la revisión, y ahora esas secciones
lo dicen y se nombran la una a la otra.

## El origen de la cifra de 25 milisegundos

El tope de bytes del mensaje de arranque (["El presupuesto del mensaje de arranque"](vocabulario-y-mensaje-de-arranque.md#el-presupuesto-del-mensaje-de-arranque)) trae su medida. ["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque),
25 milisegundos de reloj para `biso ls` y `biso prime` sobre un tablero de
300 tareas, no la tenía escrita en ningún sitio, y esta sección es esa medida.

**Medido en la misma máquina que documenta ["Estado del arte"](../estado-del-arte/index.md)** (Apple M3 Max, macOS 26.5.2, 300
iteraciones, ["El coste de arranque y el coste de contexto"](../estado-del-arte/catalogo-de-problemas.md#12-el-coste-de-arranque-y-el-coste-de-contexto)). El suelo del sistema operativo para arrancar cualquier
proceso, sin ejecutar ninguna línea propia todavía, es **5,2 milisegundos**. Un binario de Go añade
**2,2 milisegundos** encima de ese suelo. Y un programa en Go que lee 300 tareas **de un fichero
JSON**, las ordena y las imprime tardó **8,7 milisegundos en total**, suelo, arranque de Go y trabajo
real incluidos. La cifra del presupuesto, 25 milisegundos, deja **unas tres veces de margen** sobre ese
total medido.

**Y hay que decir con qué se midió ese total, porque no es el almacén que se acabó eligiendo.** Los
8,7 milisegundos salen de leer un fichero JSON, no la base de datos SQLite que decide
["La decisión de persistencia"](#la-decisión-de-persistencia), que en aquel momento todavía no estaba decidida. La medida vale para lo que se usó: fijar un
presupuesto que un lenguaje compilado cumple de sobra y que un interpretado no cumple. Lo que no es es
una medida del programa terminado: abrir el fichero de la base de datos, preparar sentencias y recorrer
índices no cuesta lo mismo que leer un fichero de texto de una vez.

**El almacén real ya está medido, y esto es lo que le pasó a esa cifra de margen** (["El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`"](lenguaje-y-rendimiento.md#el-controlador-de-sqlite-es-moderncorgsqlite-sin-cgo)). Con
la base de datos SQLite de verdad y el controlador elegido, leer el tablero de 300 tareas cuesta **14,5
milisegundos** en un portátil macOS y **3,2 milisegundos** en Linux. O sea que las tres veces de margen
se quedan en **una vez y siete décimas** en el portátil, y crecen a casi ocho veces en Linux. El
presupuesto se cumple en las dos, y desde ahora la frase "sobra tres veces" hay que leerla como una
cifra de la carga de JSON y no del programa terminado.

**Y una corrección al suelo, que no era una constante.** Los 5,2 milisegundos del párrafo anterior son
un suelo de macOS. El equivalente medido al elegir el controlador, un binario de Go que no hace
absolutamente nada, o sea suelo del sistema más arranque de Go, da **8,1 milisegundos** de mediana en el
portátil macOS, coherente con los 7,4 que suman las dos cifras de arriba. Ese mismo binario tarda
**0,37 milisegundos** en Linux, veintidós veces menos. No toca el presupuesto, que ["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque)
amarra a la máquina que ejecuta la integración continua y no a un modelo de hardware, pero
refuerza el argumento de ["La decisión de persistencia"](#la-decisión-de-persistencia) contra tener un daemon: en el portátil el suelo se come 8,1 de los
14,5 milisegundos que cuesta la lectura real, y un daemon no puede ahorrar el suelo, porque el cliente
que hablaría con él es también un proceso.

**La cifra excluye a propósito los lenguajes interpretados.** En la misma máquina, el solo arranque de
Python 3.14 añade 24,5 milisegundos por delante de cualquier trabajo real, y el de Node 25.6 añade 33.
Un presupuesto que tuviera que cubrir ese arranque dejaría de medir la herramienta y pasaría a medir el
lenguaje, así que la cifra se fija mirando el suelo que un lenguaje compilado permite. Eso acotó la
lista de candidatos a los compilados, y el apartado siguiente cierra la elección dentro de esa lista.
