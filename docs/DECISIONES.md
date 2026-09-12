# Por qué `biso` es como es

[`docs/spec/`](spec/index.md) dice qué hace el programa y nunca por qué. Este documento es el complemento: la razón
de cada decisión que podría parecer arbitraria, y la evidencia que la sostiene.

**Sirve para una cosa concreta:** antes de cambiar una regla de la especificación, hay que mirar aquí
si esa regla existe por algo. Varias de ellas parecen caprichos de estilo y son la respuesta a un
fallo medido en herramientas reales.

Este documento usa las mismas palabras que la especificación y con el mismo significado; la tabla de
["Vocabulario de esta especificación"](spec/vocabulario.md) es la referencia para las dos.

La evidencia viene de dos sitios. El primero es un estudio del uso real de un gestor de tareas por
agentes automáticos: 856 invocaciones de línea de comandos en 60 sesiones y 10 proyectos a lo largo
de seis días, con la salida de cada llamada medida en bytes. El segundo es la comparación de dos
gestores de tareas maduros, Backlog.md y Taskwarrior, y de los fallos documentados de ambos.

---

## La evidencia detrás de los siete principios

Los principios de ["Los principios"](spec/principios.md) están ahí enunciados sin su procedencia,
porque un principio se aplica igual se sepa o no de dónde viene. Aquí está de dónde viene cada uno.

**Principio 1, que un valor desconocido es un error al leer y al escribir.** Es el fallo más peligroso
que se midió. En la herramienta estudiada, `task list --status "Todo"` contesta `No tasks found.` con
código de salida 0, mientras que `task edit -s "Todo"` escribe correctamente el mismo valor. Un agente
que escribe mal un filtro no recibe un error: recibe una afirmación falsa sobre el tablero, y "no hay
tareas" es exactamente la clase de respuesta sobre la que se construye la siguiente decisión, incluido
informar al usuario de que algo no existe.

**Principio 2, que un nombre significa lo mismo en todos los comandos.** En la herramienta estudiada,
`--ref` y `--acceptance-criteria` añaden en el comando de creación y reemplazan en el de edición. Cada
comando por separado es defendible; juntos son una trampa, porque quien escribe una secuencia los mira
juntos.

**Principio 3, que ninguna bandera depende de una regla que haya que conocer de antemano.** Se
midieron seis casos de agentes usando la variante destructiva de forma repetida creyendo que
añadían: la bandera de plan aplicada hasta cinco veces sobre la misma tarea, la de referencias tres
veces sobre otra, y un caso en el que un agente ejecutó sobre una misma tarea `--ref`, `--ref`,
`--ref`, `--add-ref`, `--remove-ref` y `--clear-refs`, que es alguien probando a ver cuál de las seis
hace lo que quiere. Ninguna de esas llamadas dio error, y el daño es silencioso: cada una borró lo
que había escrito la anterior. La primera forma de este principio resolvía el problema con una regla
única, "el nombre desnudo añade", que había que conocer de antemano para no adivinar; se sustituyó
después por dar a cada operación su propio verbo explícito, sin ninguna regla que aprender, por el
motivo que cuenta ["La regla"](spec/familias-de-banderas.md#la-regla) de `familias-de-banderas.md`.

**Principio 4, que la salida por defecto de una escritura es lo que quien llama no sabía.** De las 237
creaciones medidas, 215 llevaban una bandera que devolvía la ficha entera de la tarea recién creada, y
sumaron 179.369 bytes, casi la cuarta parte de toda la salida del estudio, sin dar el único dato que
el agente no tenía, que es el identificador. La mediana de una creación pasa de 154 bytes sin esa
bandera a 1.333 con ella, un factor de 8,7.

**Principio 5, que un gesto del flujo de trabajo es un comando.** De las 159 ediciones medidas, 112
cambian exactamente un campo, hay 82 pares de ediciones consecutivas sobre la misma tarea, y el ciclo
de vida típico cuesta entre seis y doce llamadas. **Y no era una limitación de la herramienta**: se
comprobó que una sola llamada aceptaba el cierre entero y funcionaba. La fragmentación venía de que
sus guías presentaban el trabajo como una lista numerada con un comando por paso.

**Principio 6, el lote con validación previa.** Una migración real creó 242 tareas de una en una, y
cuando el entorno bloqueó los comandos compuestos que hacían falta, los agentes acabaron escribiendo
85 ficheros de tarea a mano, 76 ediciones y 9 creaciones, que es exactamente lo que la herramienta
prohíbe en la instrucción que ella misma inyecta en cada proyecto.

**Principio 7, que la salida no depende del terminal.** La bandera `--plain` de la herramienta
estudiada tiene doble vida: en unos comandos apaga una interfaz interactiva y en otros enciende un
volcado completo. Medido fuera de un terminal, en los comandos de lectura no cambia un solo byte y
aparece 191 veces sin ningún efecto; en el de creación multiplica la salida por 3,3 y en el de edición
por 36.

---

## Una advertencia sobre cómo se mantiene la especificación

La especificación pasó por cuatro revisiones adversariales antes de darse por buena. El patrón de
fallo dominante, y con diferencia, fue siempre el mismo: **dos copias distantes de un mismo dato que
dejan de coincidir**. Una lista de campos que aparece en varias secciones, una cifra publicada en tres
sitios, un código de salida que está en la tabla de un comando pero no en su texto de ayuda.

De ahí salen tres costumbres que conviene mantener al editar:

1. **Cuando un dato tenga que aparecer en dos sitios, que uno remita al otro** en vez de repetirlo.
2. **Los ejemplos de salida se generan ejecutando el algoritmo, no se escriben a mano.** Los del
   listado y los del mensaje de arranque fallaron tres revisiones seguidas mientras se escribieron a
   mano, y dejaron de fallar en cuanto se generaron.
3. **Al corregir una afirmación, búscala en todo el documento** antes de darla por corregida.

Y una cuarta, sobre este documento en particular: la especificación no justifica sus decisiones, y esa
regla es fácil de romper sin darse cuenta. La justificación no solo se esconde en la prosa, también en
la estructura. Una tabla llegó a tener una columna titulada "Por qué" que sobrevivió a cuatro
revisiones, dos de ellas dedicadas expresamente a cazar justificaciones, porque todo el mundo buscaba
frases y esa vivía en una celda.

---

## Lo que se deja fuera, y por qué

- **Los hitos como entidad.** En el estudio, el comando de crear hitos se usó 22 veces, pero los
  comandos de documentos y de decisiones no se usaron ni una sola vez en seis días. `biso` conserva el
  hito como campo de la tarea y no crea una entidad con ciclo de vida propio: no hay comando que cree
  un hito, ni clave de configuración que lo declare, ni fecha ni estado propios. Que no haya entidad
  no quiere decir que la bandera no valide, y conviene no confundir las dos cosas: `--milestone`
  valida contra el conjunto de hitos que las tareas usan de hecho (["Qué valida cada filtro, y contra qué"](spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué)), que es derivado
  y se llena solo. Lo que sí sería contradictorio es hacerla validar contra "los hitos definidos",
  porque nada declara ese conjunto y nadie podría llenarlo.
- **El servidor de integración.** En 1.280 transcripciones no hubo una sola llamada al servidor de
  herramientas que la herramienta estudiada ofrece, pese a estar disponible. Al analizarlo se vio que
  arregla buena parte de los errores de parámetros y **ninguno** de los problemas de granularidad. Las
  mejoras de nomenclatura que sí acierta, distinguir por el nombre lo que añade de lo que sustituye,
  están adoptadas en la línea de comandos de `biso`.
- **Los contextos de sesión**, es decir, filtros por defecto guardados que cambian lo que devuelve una
  consulta sin que se vea en la línea de comandos. Es la clase de estado invisible que hace que quien
  lee un listado saque conclusiones falsas, y contradice el principio 1. Se descarta a propósito y no
  por olvido.
- **El commit automático por cada llamada.** Se midió que convierte el ciclo de una tarea en siete
  commits, seis de ellos con el mensaje idéntico. Presuponía que las tareas eran ficheros versionados
  del propio proyecto, y esa premisa se descarta entera con la [decisión de persistencia](#la-decisión-de-persistencia):
  el tablero vive en un almacén aparte, y lo único que se versiona es la instantánea de texto que
  `biso snapshot` escribe cuando quien llama lo pide, un solo commit por invocación y nunca uno por
  cada escritura de tarea.
- **La visibilidad entre versiones del proyecto.** Se habían medido cinco fallos reales de las copias
  de trabajo paralelas de otras herramientas: dos "tarea no encontrada" sobre tareas que existían en
  otra rama, tres volcados de pila al leer de una rama remota, y un identificador ambiguo. La [decisión
  de persistencia](#la-decisión-de-persistencia) no los resuelve, los disuelve: el tablero no vive en el árbol de
  trabajo, así que una tarea cerrada está cerrada y no hay una rama de la que leerla ni una remota que
  le falte. Lo que queda, y sigue siendo independiente del almacenamiento, son los tres mensajes
  distintos de ["no la encuentro"](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro) y la garantía de que ninguna lectura de
  conjunto aborta por una tarea que no se puede leer.
- **La sincronización con sistemas externos.** No está, pero sí están las cuatro piezas que la hacen
  posible, y esa es la única razón por la que existen: las claves declaradas de `ext` para guardar la
  identidad de la tarea en el otro sistema, el autor libre en los comentarios, las fechas fijables al
  importar y la simetría de `export` con `new --from`.
- **La clave de configuración `default_assignee`.** Habría asignado una persona a toda tarea creada
  sin `-a`. Se descarta porque en un tablero que la usara, absolutamente todo nacería asignado, y la
  asignación dejaría de significar que alguien decidió encargarte justo esa tarea: la consulta de
  arranque de un agente devolvería el backlog entero disfrazado de encargo. Es la comodidad concreta
  que habría destruido la señal en la que se apoya la decisión de ["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo), que la asignación sea
  el gesto con el que una persona encarga trabajo.

---

## Cuatro requisitos aprendidos de otras herramientas

Estos no salen del estudio de uso, sino de comparar Backlog.md con Taskwarrior y mirar qué falla en
cada uno.

**Ningún campo desconocido se acepta ni se descarta en silencio.** Las dos herramientas lo hacen mal:
Backlog.md borra el campo sin avisar en la siguiente reescritura del fichero, y Taskwarrior, ante un
atributo personalizado no declarado, en el peor caso se come la descripción entera de la tarea. La
regla correcta es rechazar con un error explícito, y de ahí salen las claves declaradas de `ext` y la
regla de que una tarea con una clave no declarada falla en una lectura dirigida en vez de perderse.

**Cualquier fecha se puede fijar al importar.** Ninguna de las dos lo permite desde su interfaz
pública. Backlog.md no tiene un solo parámetro de fecha, ni para la de creación ni para el instante de
un comentario, y por eso un importador de histórico no puede preservar las fechas reales usando su
línea de comandos, que es su única vía legítima de escritura. Un detalle que conviene recordar: dentro
de Backlog.md existe una función interna que sí acepta una fecha explícita para un comentario, pero no
está expuesta ni por su CLI, ni por su servidor, ni como librería.

**La exportación es literalmente el formato de importación.** Ninguna de las dos lo cumple: la
exportación de Backlog.md es un informe de solo lectura, y la de Taskwarrior mezcla datos reales con
derivados como el identificador de sesión y la urgencia sin separarlos. En `biso`, `export` escribe
todos los campos no derivados y ninguno derivado, y la ida y vuelta es una prueba de la suite.

Hay una trampa concreta en esto, y cuesta verla: **si el formato de lote admite los criterios de
aceptación como simples cadenas de texto, la simetría es falsa por construcción**, porque un criterio
tiene clave, texto y marca de cumplido, y al reimportar se pierden las claves y las marcas. Por eso el
formato admite objetos.

**Una funcionalidad no se retira sin anunciarla, y un cambio de formato lleva su camino de migración.**
Taskwarrior 3.0 retiró el historial de una de sus vistas sin aviso, y su cambio de motor de
almacenamiento provocó un caso real documentado de pérdida total de la base de tareas en Arch Linux,
porque el paquete se actualizó sin incluir el script de migración. De ahí sale el ciclo de aviso
obligatorio del contrato de estabilidad.

---

## El modelo de estados: cuatro requisitos, cerrados

Salen de un diseño anterior de gestor de tareas que no llegó a escribirse, y de la evidencia que
aquel diseño recogió sobre la herramienta que usaba esta máquina antes. Los cuatro tocaban el modelo
de estados y se decidieron a la vez: dos quedan resueltos sin ningún papel de estado nuevo, uno se
resuelve con la enmienda de la decisión de persistencia (["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad)), y uno se retira. El criterio que ordenó
las cuatro decisiones, y que conviene aplicar la próxima vez que alguien proponga un papel de estado,
está en ["El criterio de estado frente a campo"](#el-criterio-de-estado-frente-a-campo).

### Distinguir el encargo de la ejecución: resuelto sin estado nuevo

El requisito decía que, con los tres papeles de la especificación, el gesto de una persona que encarga
trabajo y el de un agente que lo coge **son el mismo dato**, y que hacía falta un cuarto papel de
estado para separarlos.

**El diagnóstico era falso, y la solución sale de corregirlo, no de añadir nada.** La persona ya
escribe un dato: `assignees`. El agente escribe otro: el estado activo, al ejecutar `biso start`. Son
dos datos distintos, escritos por dos actores distintos, y esa asimetría existía ya en la
especificación antes de este trabajo; nadie la había mirado como la respuesta al requisito. Y
aplicando el criterio de ["El criterio de estado frente a campo"](#el-criterio-de-estado-frente-a-campo), tampoco podía ser nunca un estado: "esto lo tiene que hacer un
agente" convive con cualquier punto del camino, porque una tarea puede estar recién creada, a medias,
o aparcada en una pregunta abierta y seguir siendo de quien se la asignaron.

**Esta lectura solo es cierta sin la clave de configuración `default_assignee`.** Esa clave habría
asignado una persona a toda tarea creada sin `-a`. En un tablero que la usara, absolutamente todo
nacería asignado, y la asignación dejaría de significar que alguien decidió encargarte justo esa
tarea: la consulta de arranque de un agente devolvería el backlog entero disfrazado de encargo. Se
retira la clave entera: su fila en la tabla de configuración, su comportamiento en `biso new`, su
aparición en el esquema JSON de `config list` y su línea de `biso config --help`. La autoasignación de
`biso start` se queda, porque no ensucia la señal: cuando `start` asigna a `me`, la tarea entra a la
vez en el estado activo, así que nunca queda en el conjunto de "asignada y sin empezar", que es el que
importa.

**Lo que esta decisión cuesta, para que no se descubra por sorpresa.** El caso de uso que motivó el
requisito era arrastrar una tarjeta desde el móvil para que un agente se ponga con una tarea. Con la
asignación como encargo, el gesto pasa a ser asignar a un miembro, que sigue siendo viable desde un
móvil pero ya no es un arrastre, así que la frase que describía el caso de uso queda anticuada y no
describe ya la herramienta. Cuando se especifique la sincronización con un sistema externo, mover una
tarjeta de columna en ese sistema no significará nada para `biso`, y esa decisión hay que tomarla
entonces a propósito, no por sorpresa. Y no hay forma de decir "esto es tuyo, pero todavía no": con la
asignación como única señal, asignar autoriza a empezar de inmediato, y quien necesite esa espera tiene
que no asignar hasta que toque, o usar una fecha límite.

### Saber si alguien está trabajando de verdad

Una tarea que un agente coge antes de que su sesión muera se queda en el estado activo
indefinidamente, y nada lo detecta.

La evidencia es concreta. La herramienta que se usaba antes tenía un campo para apuntar quién está
trabajando en una tarea, y el tablero web lo pintaba, pero era una lista a la que se añade un nombre
al empezar y que se vacía al terminar: **sin latido, sin marca de tiempo y sin caducidad**, de modo
que una sesión que muere deja su nombre ahí para siempre. Y había un segundo fallo encima: solo lo
escribían su servidor de integración y su API, nunca su línea de comandos, aunque su propio mensaje
de arranque pedía al agente que mandara su nombre al cambiar de estado.

Hay además un dato medido que descarta la solución fácil: **el evento de fin de sesión de Claude Code
trae un motivo con cinco valores posibles y ninguno corresponde a un proceso muerto a lo bruto**, así
que un hook de cierre no puede ser la única señal de que una tarea quedó huérfana, porque el caso que
importa es justamente el que no lo dispara.

La forma conocida de resolverlo es un arrendamiento con caducidad: latido mientras se trabaja, y algo
que libere la tarea cuya sesión murió. Arrendar sería asignar con esa caducidad sobre una tarea ya
activa, y reclamarla sería la misma escritura que hoy hace `biso start`, así que la decisión de este
apartado sigue encajando con la de ["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo) aunque aquel, al final, no trajera ningún papel nuevo.

**Esta dependía de cómo se persistan los datos**, y se cerró con esa decisión tomada.

**Cerrada, con una enmienda a lo que este apartado decía antes.** La redacción aplazada decía que la
caducidad, al vencer, "saca la tarea del estado activo sin tocar la asignación". Esa frase no se
sostiene y es la que cede: `status` es un campo guardado, con un valor del vocabulario del tablero, y
un campo derivado no puede cambiar un campo guardado. Si nada escribe, la tarea sigue teniendo el
estado activo guardado por muy vencido que esté su arrendamiento.

Lo que se guarda es el instante de caducidad (`leaseExpiresAt`) y quién tiene el arrendamiento
(`leaseHolder`), los dos sin valor salvo en una tarea activa y asignada (["El modelo de datos de una tarea"](spec/modelo-de-datos.md)). Lo
que caduca sigue siendo, como decía la redacción original, "estoy en ello" y no "esto es tuyo": el
campo derivado `leaseExpired` dice que el arrendamiento venció, pero el estado guardado no cambia
solo, nunca. Liberarlo es una escritura explícita, y sigue sin hacer falta un comando nuevo para eso:
es la misma que ya hace `biso start`, que reclama el arrendamiento vencido a favor de quien llama
comprobando quién lo tenía dentro de la misma transacción, para que el tenedor viejo no lo recupere con
su siguiente escritura al despertar (["`biso start`"](spec/cmd/verbos-del-ciclo.md#biso-start)). Con un arrendamiento vivo de otra
identidad, `biso start` avisa y la coge igual: el mismo "avisa, no impide" que ya aplicaba a las
dependencias sin terminar y a la pregunta abierta.

**Por qué se enmienda en vez de reescribirse sin más.** La redacción aplazada no decía cómo se
liberaba una caducidad, y la lectura más directa de "saca la tarea del estado activo" es una
escritura diferida: que la siguiente escritura cualquiera, o un proceso de fondo, arrastrara el
saneamiento de las tareas vencidas. Esa vía se descarta explícitamente al tomar esta decisión, en
["Una tarea se queda cogida porque la sesión murió"](ESTADO-DEL-ARTE.md#6-una-tarea-se-queda-cogida-porque-la-sesión-murió), porque haría que un comando tocara tareas que no nombró, y
porque `biso prime`, que no escribe nunca, mostraría un estado que una escritura ajena y posterior
podría cambiar. No es que la decisión siempre hubiera sido la de hoy: es que la única forma de
sostenerla, al escribirla de verdad, obligaba a mover la contradicción con `status` a otro sitio en
vez de resolverla.

**Lo que aporta el estado del arte, mirado al cerrar esta decisión.** El patrón tiene nombre propio
fuera de aquí: un arrendamiento con caducidad, renovado por latido, para evitar la doble reclamación.
La pieza que le faltaba a la redacción aplazada es la comprobación del tenedor, y la idea viene de un
artículo que formaliza el patrón: propone el límite de tiempo **más un token de vallado que rechace las
escrituras del propietario antiguo** (["Una tarea se queda cogida porque la sesión murió"](ESTADO-DEL-ARTE.md#6-una-tarea-se-queda-cogida-porque-la-sesión-murió), con su enlace). Sin nada
de eso, el tenedor viejo puede despertar, escribir, y robar de vuelta una tarea que ya había reclamado
otro. Comprobar quién lo tiene sale gratis aquí, porque el arrendamiento ya guarda ese dato y la
comprobación es comparar y sustituir dentro de una transacción que ya existía por otro motivo
(["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables)).

**Pero lo que se implementa no es ese token de vallado, y apartarse de él es deliberado.** Un token de
vallado rechaza la escritura entera de quien ya no es el propietario. `biso` la acepta: la escritura de
una identidad que no es `leaseHolder` se hace igual y solo deja intactos los campos del
arrendamiento, con el aviso de ["Notas y avisos"](spec/salida-y-terminal.md#notas-y-avisos) (sexta precisión de ["El modelo de datos de una tarea"](spec/modelo-de-datos.md)). Así que el
agujero que el artículo cierra aquí queda entreabierto: el tenedor viejo que despierta puede comentar,
anotar o cerrar la tarea que otro reclamó, y si vuelve a llamar a `biso start` se la lleva de vuelta,
con aviso y sin impedimento. Se acepta porque rechazar la escritura sería lo único de todo el programa
que impide trabajar por el estado en que está una tarea: las dependencias sin terminar avisan, la
pregunta abierta avisa, y un arrendamiento ajeno avisa igual, por el motivo de ["El porqué de reglas concretas"](#el-porqué-de-reglas-concretas), que un
bloqueo de flujo no evita el trabajo duplicado y sí empuja a rodear la herramienta. Lo que la
comprobación del tenedor sí cierra, y es lo que se gana, son las dos cosas que el aviso no puede dar:
que una escritura de otra identidad nunca renueve el plazo ajeno ni se atribuya el arrendamiento, y que
de dos reclamaciones simultáneas de un arrendamiento vencido solo gane una, porque `biso start`
comprueba dentro de su propia transacción que seguía vencido. La diferencia con el artículo queda
anotada como riesgo aceptado en ["Riesgos conocidos y aceptados del modelo de estados"](#riesgos-conocidos-y-aceptados-del-modelo-de-estados).

**Por qué el arrendamiento se exporta e importa como cualquier otro campo.** Al añadir los campos
guardados quedó sin decir si viajan en `biso export`, y las dos respuestas eran defendibles: dejarlos
fuera, porque un arrendamiento es la reserva de una sesión concreta en una máquina concreta, o dejarlos
entrar, porque son campos guardados y no derivados y la garantía de simetría de ["`biso export`"](spec/cmd/export.md)
promete que todos ellos van y vuelven. Se eligió la segunda, y en rigor no era una elección
libre: el ["contrato de estabilidad"](spec/estabilidad.md) ya enuncia esa simetría como una prueba
de la suite "sobre todos los campos no derivados", sin lista de excepciones, así que dejar fuera el
arrendamiento habría obligado a abrir una y a mantenerla, que es exactamente la clase de enumeración
que se desincroniza (["Una advertencia sobre cómo se mantiene la especificación"](#una-advertencia-sobre-cómo-se-mantiene-la-especificación)). La otra razón, la que hace que la primera no salga cara, es que el
diseño ya toleraba un arrendamiento ajeno: `leaseExpired` se recalcula contra el reloj de
quien lee, así que el arrendamiento que llega caducado sale caducado y `biso start` lo reclama, y el
que llega vivo a nombre de otra identidad produce el aviso de ["Notas y avisos"](spec/salida-y-terminal.md#notas-y-avisos) y nada más, porque `biso
start` avisa y coge la tarea igual. El caso que de verdad importa, restaurar un respaldo del propio
tablero, sale además mejor así: la tarea que estaba en marcha sigue constando en marcha y a nombre de
quien la llevaba, en vez de aparecer activa y sin dueño del arrendamiento. Lo único que hay que
custodiar es la invariante de que los campos solo tienen valor en una tarea activa y asignada, y se
custodia en dos sitios: en la validación previa de `biso new --from`, donde se custodia todo lo demás
del lote y que rechaza el fichero entero antes de escribir nada, y en `biso doctor`, que la comprueba
como comprueba las demás invariantes del tablero y la repara con `--fix` vaciando los campos
(["`biso doctor`"](spec/cmd/doctor.md)). El segundo hace falta porque la importación no es la única forma de que
una base de datos llegue a incumplirla: puede venir escrita a mano, restaurada a medias o de otra
versión, y sin esa comprobación la única invariante que este apartado dice que hay que custodiar sería
la única que `doctor` no mira.

**La invariante gana siempre sobre el aviso de un arrendamiento ajeno, y esa precedencia hay que
escribirla.** Estas reglas se enfrentan en un caso corriente: `@claude` tiene el arrendamiento vivo de
una tarea y `@sara` ejecuta `biso finish` sobre ella. Una regla dice que la escritura de otra identidad
no toca ninguno de los campos, y solo avisa; la otra dice que la escritura que saca la tarea del
estado activo los vacía. Manda la segunda, y el aviso se emite igual. El argumento no es de gusto: la
primera regla es una cortesía hacia el tenedor y la segunda es la invariante de la que dependen la
validación del lote y `doctor`, así que con la precedencia al revés quedaría una tarea terminada con un
arrendamiento vivo, y `biso export` de ese tablero produciría un fichero que su propio
`biso init --from` rechazaría, rompiendo la prueba de simetría del ["contrato de estabilidad"](spec/estabilidad.md). Por el
mismo argumento, `biso archive` vacía también los campos, aunque `archived` no sea un estado y
archivar no saque la tarea del estado activo: un arrendamiento afirma que alguien está trabajando ahora,
y archivar es dejar de trabajar, así que conservarlo lo esconde donde nadie lo ve (`biso prime` y
`biso ls` excluyen las archivadas) hasta que `--unarchive` lo devuelve semanas después a nombre de una
sesión muerta.

**Una escritura que no cambia ningún campo también late, y esto es lo que un implementador desharía
creyendo que optimiza.** `biso set` con todas sus banderas dando el valor que la tarea ya tiene sale con
código 0 y con `note: MYP-11 unchanged`, y aun así renueva `leaseExpiresAt` si quien llama es el
tenedor. Salta a la vista el atajo contrario, no escribir nada cuando no hay nada que escribir, y es un
error: el latido de este arrendamiento no es un comando propio, es cualquier escritura que el agente ya
hace, y si la renovación dependiera de que algún valor hubiera cambiado de verdad, un agente que repite
una escritura idempotente creería estar latiendo sin latir, y perdería la tarea al vencer el plazo por
haber hecho justo lo que la especificación le dice que basta. Lo que no cambia en ese caso es
`updatedAt`, porque ningún campo de la tarea cambió, y la lista `changed` del JSON sale vacía: la
renovación es un hecho del arrendamiento, no una modificación de la tarea, y la nota sigue siendo cierta
porque habla de la tarea.

**Y `biso new --start` reclama el arrendamiento, porque es el atajo de dos llamadas y la equivalencia
tiene que ser real.** ["El porqué de reglas concretas"](#el-porqué-de-reglas-concretas) justifica esa bandera como el ahorro de `biso new` más `biso start`;
si creara la tarea activa y asignada pero sin arrendamiento, las dos vías darían dos tareas distintas y
la única forma de saberlo sería leer la letra pequeña. La excepción simétrica es `biso start -s` con un
estado que no es el activo: ahí no se fija arrendamiento, porque fijarlo rompería la invariante de
arriba, y hay que nombrarla explícitamente para que la frase "solo `start` reclama" no se lea como una
regla sin excepciones.

**Dos remates que la primera redacción dejó a medias.** El primero: los campos ya se podían pedir
con `biso get --json`, y el mensaje de arranque remitía a `biso get` para verlos, pero su ficha de texto
no tenía dónde enseñarlos, así que la remisión era falsa para quien no pide JSON. Ahora la ficha imprime
una línea `lease` con los dos, y solo cuando la tarea tiene arrendamiento, porque una fila con dos
guiones aparecería en la ficha de casi todas las tareas del tablero (["`biso get`"](spec/cmd/get.md)). El
segundo: la validación rechazaba los campos sobre una tarea que no estuviera activa y asignada, pero
no rechazaba que llegara uno solo de los dos, y con `leaseExpiresAt` vacío el derivado `leaseExpired`
comparaba un instante que no existe contra el reloj. Los campos van juntos o no viene ninguno, y
`leaseExpired` es falso cuando no hay `leaseExpiresAt`: un derivado tiene que valer algo en todos los
tableros posibles, no solo en los que se importaron bien.

### Señalar lo que espera a una persona

Se puede configurar un estado tipo `Blocked`, pero es un estado más: `biso prime` no lo distingue, así
que una tarea parada esperando una decisión humana no se ve donde se mira.

La evidencia: en el tablero que se estudió había cuatro tareas paradas por una pregunta sin responder
y nada lo señalaba. Eran, además, preguntas abiertas disfrazadas de trabajo pendiente, porque una
decisión no tenía dónde vivir y la única forma de registrarla era crear una tarea.

Lo que falta es un papel que marque un estado como "espera a una persona", y que el mensaje de
arranque lo destaque en su propio bloque.

**Queda resuelta, y no con el papel que este apartado imaginaba.** Aplicando el criterio de
["El criterio de estado frente a campo"](#el-criterio-de-estado-frente-a-campo), una pregunta abierta puede detener una tarea en cualquier punto del camino, así que no podía ser
un estado: es el campo `question` de ["La pregunta abierta"](spec/modelo-de-datos.md#la-pregunta-abierta), con su derivado `waiting`, los
verbos `biso ask` y `biso answer`, y el bloque `NEEDS ANSWER` del mensaje de arranque, que es
exactamente el bloque propio que este apartado pedía.

### Distinguir terminar de descartar: retirado

Se retira, y conviene decir por qué el argumento que parecía bueno no lo era, para no repetir el error
si alguien vuelve a proponerlo.

**El argumento que no vale es que archivar y descartar son lo mismo.** No se sostiene contra lo que ya
dice `biso archive`: existe `--unarchive`, que devuelve la tarea al tablero con el estado que tenía, y
la ayuda del comando presenta el archivo como sacar tareas del tablero sin perderlas. Archivar es
reversible, así que archivar es aparcar, y aparcar no es descartar.

**Y la deducción que se apoyaba en él tampoco vale.** La idea era que una tarea archivada que nunca
llegó al estado terminal se pudiera etiquetar como abandonada. No se puede calcular, porque el modelo
guarda el estado actual de una tarea y no un histórico de sus estados anteriores. Con `biso start
--reopen` una tarea terminada vuelve al estado activo, y si se archivara desde ahí, esa deducción la
llamaría abandonada habiendo estado hecha.

**El argumento que sí vale es que a este requisito le falta la evidencia que los otros tres sí traen.**
["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo) corrige un diagnóstico sobre una asimetría real de la especificación,
["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad) trae el fallo
medido de una herramienta que se usaba antes (un campo sin latido, sin marca de tiempo y sin
caducidad), y ["Señalar lo que espera a una persona"](#señalar-lo-que-espera-a-una-persona) trae cuatro tareas paradas por una pregunta sin responder en el tablero que se
estudió. Este requisito dice que una tarea hecha y una abandonada "se confunden", sin un solo caso en
el que esa confusión haya costado algo. Un papel de estado que obligue a dar un motivo es barato de
añadir cuando haga falta y caro de quitar si sobra, así que se queda fuera hasta que aparezca un caso
real que lo pida, y entonces se engancha a `biso archive`. Se anota en ["Lo que se deja fuera a propósito"](spec/fuera-de-alcance.md).

### Lo que se miró de ese diseño anterior y se descarta

**Un modelo con entidades separadas para tarea, idea, aprendizaje y decisión**, cada una con sus
propios estados y campos. La necesidad que lo motivaba es real, porque sin sitio donde ponerlas las
decisiones se disfrazan de tareas. Pero hay un dato posterior que contradice la solución: en el
estudio de uso de una herramienta que sí tiene comandos dedicados para documentos y decisiones,
**no hubo una sola llamada a ninguno de ellos en seis días, sesenta sesiones y diez proyectos**. La
conclusión es que la necesidad se cubre mejor con un tipo más en el vocabulario que ya existe que con
una entidad y un comando propios.

**Una forma concreta de guardar los datos**, con un directorio por entidad para que añadir un
comentario no reescriba nada de lo demás. La [decisión de persistencia](#la-decisión-de-persistencia) no la adopta: una
base de datos SQLite da la misma propiedad, una escritura por tarea sin reescribir el tablero entero, y
además la transacción que un directorio de ficheros habría tenido que construir a mano.

### Un contraste que conviene mirar antes de implementar

Aquel diseño se imponía una regla contraria a la de esta especificación: **los valores por defecto
viven completos en el binario, el comando de creación del tablero no escribe ninguna configuración, y
un repositorio normal no tiene fichero de configuración nunca**. Su argumento era evitar acabar con un
motor genérico que no sabe hacer nada solo y que obliga a configurar antes de escribir la primera
tarea.

Aquí `biso init` sí escribe la configuración. El argumento contrario es bueno y merece mirarse antes
de dar la decisión por hecha.

---

## El criterio de estado frente a campo

Es la regla que ordenó las cuatro decisiones de ["El modelo de estados: cuatro requisitos, cerrados"](#el-modelo-de-estados-cuatro-requisitos-cerrados), y conviene tenerla escrita aparte porque
se va a volver a necesitar la próxima vez que alguien proponga un papel de estado:

> Algo es un estado cuando es excluyente con los demás y dice en qué punto del camino está la tarea.
> Es un campo cuando puede convivir con cualquier punto del camino.

La especificación ya la aplicaba sin decirla: `archived` es un campo y no un estado precisamente
porque una tarea archivada conserva el estado que tenía al archivarse. Lo que faltaba era el criterio
escrito, para no volver a meter en el vocabulario de estados algo que no es un punto del camino.
Aplicado a los cuatro requisitos de ["El modelo de estados: cuatro requisitos, cerrados"](#el-modelo-de-estados-cuatro-requisitos-cerrados), deja solo uno pidiendo de verdad algo excluyente y
ligado al camino, el de ["Distinguir terminar de descartar: retirado"](#distinguir-terminar-de-descartar-retirado), y ese es justo el que se retira por falta de evidencia; los otros tres son un
gesto que ya existía (["Distinguir el encargo de la ejecución"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo)), una reserva con caducidad (["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad)) o un campo que puede convivir con cualquier
estado (["Señalar lo que espera a una persona"](#señalar-lo-que-espera-a-una-persona)).

**Por qué esta condición no se hace configurable.** Dejar que cada tablero declarase sus propios
estados como excluyentes o no destruiría la garantía en la que descansa el resto del modelo: que
`biso start`, `biso finish` y los filtros por papel (`--active`, `--not-active`) puedan asumir siempre
que una tarea está en exactamente un estado a la vez. Si esa garantía dependiera de la configuración de
cada tablero, todo comando tendría que consultarla antes de decidir qué significa "activa", y eso es
justo la clase de comportamiento que depende de dónde y con qué se ejecuta el programa que la
introducción de la especificación descarta. La condición es del modelo, no de un tablero concreto.

---

## Riesgos conocidos y aceptados del modelo de estados

Se aceptan a propósito, y conviene anotar por qué en cada uno para no tropezar dos veces con lo mismo.

- **Dos agentes con la misma identidad ven la misma cola.** Dos sesiones con el mismo `BISO_ME` no se
  distinguen entre sí, y las dos podrían coger la misma tarea a la vez. Es exactamente lo que resuelve
  el arrendamiento de ["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad): la segunda sesión ve el aviso de que el arrendamiento está vivo y
  a nombre de esa misma identidad, pero `biso start` avisa y la coge igual, así que dos sesiones que
  comparten identidad siguen pudiendo pisarse. El arrendamiento defiende de una sesión muerta, no de
  dos sesiones vivas con el mismo nombre.
- **El tenedor viejo de un arrendamiento no tiene prohibido escribir.** El artículo que formaliza el
  patrón propone, junto a la caducidad, un token de vallado que rechace las escrituras del propietario
  antiguo (["Una tarea se queda cogida porque la sesión murió"](ESTADO-DEL-ARTE.md#6-una-tarea-se-queda-cogida-porque-la-sesión-murió)), y `biso` hace a propósito algo más débil: la
  escritura de una identidad que no es `leaseHolder` se acepta entera y solo deja intactos
  `leaseExpiresAt` y `leaseHolder`, con el aviso de ["Notas y avisos"](spec/salida-y-terminal.md#notas-y-avisos). Un agente que despierta
  después de que otro reclamara su tarea puede comentarla, anotarla o cerrarla, y con `biso start`
  llevársela de vuelta. Se acepta por coherencia con "avisa, no impide", el mismo argumento de
  ["El porqué de reglas concretas"](#el-porqué-de-reglas-concretas) que vale para las dependencias sin terminar y para la pregunta abierta: un bloqueo de flujo no evita
  el trabajo duplicado, solo empuja a rodear la herramienta. Lo que el arrendamiento sí garantiza es que
  una tarea no se quede cogida para siempre, que nadie renueve ni se atribuya un arrendamiento ajeno, y
  que de dos reclamaciones simultáneas del mismo arrendamiento vencido solo gane una (["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad)).
- **Un tablero con la clave `me` configurada anula la distinción entre persona y agente.** La clave
  `me` gana sobre `BISO_ME`, así que en un tablero que la tenga puesta todo el mundo comparte
  identidad y `--mine` deja de significar nada. Un tablero compartido entre una persona y un agente
  tiene que dejar `me` sin configurar. **Lo único que se ha cerrado de este riesgo es la vía por la que
  llegaba sin que nadie lo decidiera**: la instantánea de `biso snapshot` no escribe `me` (ni
  `default_limit`), así que restaurar el tablero de otra persona con `biso init --from` ya no hereda su
  identidad, y el tablero restaurado nace sin ninguna. Configurarla sigue siendo posible, y sigue
  teniendo esta consecuencia.
- **La persona no tiene un canal hacia el agente que se vea en el mensaje de arranque.** El agente
  pregunta con `biso ask` y la persona responde con `biso answer`, pero si la persona quiere decirle
  algo por iniciativa propia lo escribe en un comentario, y el mensaje de arranque no muestra
  comentarios. El agente lo ve al hacer `biso get`.
- **El listado no enseña el texto de la pregunta, solo dice qué tareas la tienen.** Es el precio de no
  meter texto largo en el listado, medido en el principio 4 de ["La evidencia detrás de los siete principios"](#la-evidencia-detrás-de-los-siete-principios): 215 fichas completas
  sumaron 179.369 bytes, casi la cuarta parte de la salida del estudio. Quien quiera leer la pregunta
  usa `biso get --section question`, o mira el mensaje de arranque, que sí la enseña.
- **Una pregunta abierta sobre una tarea terminada o archivada desaparece de la vista.** `biso finish`
  avisa pero no impide, los bloques del mensaje de arranque excluyen terminadas y archivadas, y
  `biso ls` excluye el estado terminal por defecto, así que `biso ls --waiting` no la encuentra sin
  `--any-status`. Se acepta porque la alternativa, impedir cerrar una tarea con una pregunta abierta,
  empujaría a rodear la herramienta, el mismo argumento que ya vale en ["El porqué de reglas concretas"](#el-porqué-de-reglas-concretas) para `finish` y
  los criterios sin marcar.
- **El filtro de dependencias no excluye las tareas aparcadas.** `--not-blocked` mira solo
  dependencias, así que un agente que elija trabajo únicamente con esa bandera se lleva también las
  que esperan una respuesta. La consulta correcta añade `--not-waiting`, y así lo dicen tanto la
  descripción de la bandera como el ejemplo de la ayuda de `biso ls`. El riesgo se queda, pero
  encogido: el nombre ya no promete que la tarea esté lista para trabajar, solo que no esté bloqueada,
  que es lo que mide. El porqué del nombre está en ["El porqué de reglas concretas"](#el-porqué-de-reglas-concretas).

---

## La decisión de persistencia

La especificación dejaba deliberadamente abierto cómo se guardan los datos. La decisión es: un tablero
es una base de datos SQLite en un directorio propio fuera del proyecto, localizado por un fichero
puntero versionado con el proyecto (`.biso.json`, ["Cómo se elige el tablero"](spec/resolucion-del-tablero.md)), con una exportación de texto
que sí se guarda en el control de versiones para el historial (`biso snapshot`, ["`biso snapshot`"](spec/cmd/snapshot.md)). Sin daemon, y sin fusionar nunca dos almacenes escritos por separado.

La evidencia detrás de cada pieza de esta decisión, con sus enlaces, está en
[`docs/ESTADO-DEL-ARTE.md`](ESTADO-DEL-ARTE.md), el inventario de las herramientas del espacio y el
catálogo de sus fallos. Lo que sigue aquí es el porqué de cada pieza, no la evidencia en bruto.

**Por qué el tablero no se versiona con el código.** Cualquier herramienta que guarde las tareas como
ficheros del árbol de trabajo hereda su peor propiedad: el estado se bifurca con la rama, así que una
incidencia cerrada en una rama vuelve a aparecer abierta al volver a la principal (["El estado de las tareas se bifurca con la rama"](ESTADO-DEL-ARTE.md#2-el-estado-de-las-tareas-se-bifurca-con-la-rama)). Sacar el tablero del árbol de trabajo no resuelve ese problema, lo disuelve:
una tarea cerrada está cerrada, no cerrada en esta rama, porque no hay una rama que la contenga. El
precio es que el tablero no viaja al clonar el proyecto en otra máquina, y se paga a propósito a cambio
de que el estado de una tarea sea uno solo.

**Por qué no hay daemon, y por qué el motivo es aritmético y no de gusto.** El coste dominante de una
invocación de `biso` es arrancar un proceso, no el trabajo que hace una vez arrancado: ["El coste de arranque y el coste de contexto"](ESTADO-DEL-ARTE.md#12-el-coste-de-arranque-y-el-coste-de-contexto)
mide el suelo del sistema en 5,2 milisegundos y el total de leer, ordenar e
imprimir 300 tareas en 8,7 milisegundos con un binario de Go, suelo incluido y leyendo un JSON en vez de
la base de datos (["El origen de la cifra de 25 milisegundos"](#el-origen-de-la-cifra-de-25-milisegundos)). Un daemon solo puede ahorrar lo que hay por encima del suelo, no el
suelo, porque el cliente que hablaría con él por un socket es también un proceso y paga el mismo suelo de
arranque para lanzarse. Así que un daemon competiría por uno o dos milisegundos de
unos ocho, pagando a cambio una arquitectura entera: un proceso de fondo que hay que arrancar, vigilar y
matar, y que si se cuelga hace fallar también las lecturas que ["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) promete que
nunca fallan por una escritura en curso. Beads tuvo uno, hacía una sola cosa, y se eliminó por completo
al cambiar de motor; quien lo reemplazó por algo más simple cuenta que se pasaba varias veces por semana
peleándose con él (["El proceso de fondo"](ESTADO-DEL-ARTE.md#11-el-proceso-de-fondo)).

**Por qué el texto es una salida, y nunca un canal de vuelta.** `biso snapshot` escribe `snapshot.ndjson` y
`board.json` para que el historial de git cuente lo que pasó y para que `biso init --from` pueda
reconstruir el tablero entero en otra máquina, pero nada dentro de `biso` vuelve a leer esos ficheros
como si fueran la verdad. Beads documenta por qué esa asimetría es obligatoria y no una elección
estética: su importación es solo de inserción y actualización, y no puede saber si un registro ausente
en el texto fue borrado a propósito o simplemente no se llegó a exportar (["El almacén binario no se versiona, y el texto no tiene transacciones"](ESTADO-DEL-ARTE.md#10-el-almacén-binario-no-se-versiona-y-el-texto-no-tiene-transacciones)). Tratar el texto como una fuente además de como una salida reintroduce esa
ambigüedad en `biso`, así que no se hace nunca: la base de datos es la única verdad, y `export` y
`snapshot` son su proyección de solo lectura hacia fuera.

**Por qué nunca se sincroniza fusionando dos almacenes escritos por separado.** Es el sitio donde se
han estrellado todas las herramientas del espacio, de formas distintas pero con la misma raíz: dos
copias de trabajo que asignan el mismo identificador a tareas distintas (["Dos copias de trabajo asignan el mismo identificador"](ESTADO-DEL-ARTE.md#1-dos-copias-de-trabajo-asignan-el-mismo-identificador)), un bloqueo de fichero que no cruza remotos de git (["Un bloqueo que no cruza máquinas, o que se queda huérfano"](ESTADO-DEL-ARTE.md#4-un-bloqueo-que-no-cruza-máquinas-o-que-se-queda-huérfano)), y un fichero
de log que se fusiona por unión de líneas y resucita las que se habían borrado, porque una fusión de
texto concatena y solo quita duplicados exactos, sin razonar sobre qué falta ni por qué (["El almacén binario no se versiona, y el texto no tiene transacciones"](ESTADO-DEL-ARTE.md#10-el-almacén-binario-no-se-versiona-y-el-texto-no-tiene-transacciones)). La
respuesta seria de ese último problema, sustituir el motor por uno con fusión a nivel de celda, es la que
tomó Beads, y es coherente pero cara. `biso` no la necesita porque no la tiene que resolver: un tablero
vive en una sola máquina y no hay una segunda copia escribible con la que fusionarse, así que la
comprobación de identidad de ["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) basta y no hace falta un algoritmo de fusión.

**Por qué `--fix` no es una comodidad, sino el consentimiento.** Esta decisión añade a `biso doctor`
(["`biso doctor`"](spec/cmd/doctor.md)) las comprobaciones que no existían antes de que hubiera una base de
datos real detrás del tablero: la integridad de esa base de datos, y el aviso de un sistema de ficheros
donde el modo WAL de SQLite no da las garantías de atomicidad que ["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) exige.
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
documento declara opcional. ["`biso snapshot`"](spec/cmd/snapshot.md) dice que git puede no estar instalado y que
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
descartado por ["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque): interpretar un `.gitignore` de verdad no
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
heredado, de donde sale que gane al puntero en el orden de ["Cómo se elige el tablero"](spec/resolucion-del-tablero.md). Esta vía se lleva por delante
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
pero con veinte tableros son veintiuna en vez de una, y hay que medirlo contra ["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque) cuando el programa exista. Se descartó a propósito el atajo de buscar primero por el nombre
y caer al marcador solo si falla: sería más rápido y volvería a poner el mismo dato en dos sitios, que es
lo que se acaba de quitar.

**Por qué la `path` del puntero puede ser relativa, y por qué la forma la elige quien llama.** Un puntero
que dice `/Users/avilches/Hub/Projects/Kex/tablero` solo resuelve en el ordenador donde el proyecto está
en esa ruta exacta, y el puntero se versiona precisamente para que viaje. Con `tablero` guardado como
ruta relativa, mover el proyecto entero con su tablero dentro no rompe nada.

La ruta relativa **se resuelve contra el directorio que contiene el fichero puntero, nunca contra el
directorio de trabajo.** El puntero se busca subiendo desde el directorio de trabajo, así que el mismo
fichero se lee desde cualquier subdirectorio del proyecto: resolver contra el directorio de trabajo haría
que el mismo puntero nombrara un tablero distinto por cada subdirectorio desde el que se llamara, y que
casi ninguno de ellos existiera. Eso rompería el principio de ["Los principios"](spec/principios.md) de que ningún
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

**Y por qué la nota es una nota y no un aviso.** ["Notas y avisos"](spec/salida-y-terminal.md#notas-y-avisos) reserva `warning:` para lo
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
cuatro vías por las que ["Cómo se elige el tablero"](spec/resolucion-del-tablero.md) encontraba un tablero, y la tabla de banderas
globales las resumía como "Usa ese tablero directamente, sin buscar", con un valor que ["Banderas globales"](spec/invocacion.md#banderas-globales)
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
propio de `BISO_BOARD` frente a la bandera, ya está `BISO_CWD`, que ["Variables de entorno"](spec/invocacion.md#variables-de-entorno) declara equivalente a
`--cwd` con la bandera ganando.

**La segunda: por qué no se sustituye por una bandera que acepte el nombre del tablero.** Es la pata que
más falta va a hacer, porque nombrar el tablero por su nombre es lo primero que se le ocurre a
cualquiera. El obstáculo es que **el nombre de un tablero no está en el sistema de ficheros**. Lo que hay
en el disco es una carpeta llamada `<slug>-<id>`, y ese trozo legible es decorativo a propósito:
["Cómo se elige el tablero"](spec/resolucion-del-tablero.md) dice que nadie resuelve nunca por él, y ["`biso config`"](spec/cmd/config.md) dice que cambiar `project_name`, que es el
nombre del tablero, no toca el sistema de ficheros en absoluto. De ahí que el slug pueda ser el de un
nombre anterior y no haya nada que lo corrija, ni falta que hace. El nombre de verdad vive dentro de la
base de datos, junto con el resto de la configuración. Resolver por nombre sería, entonces, abrir la base
de datos de cada carpeta de cada raíz para preguntarle cómo se llama, y eso choca de frente con
["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque), que da 25 milisegundos para todo: con veinte tableros son veinte
aperturas de SQLite antes de empezar a hacer el trabajo que se ha pedido, y eso es exactamente el trabajo
que nadie ha pedido que la primera regla de esa misma sección prohíbe. La salida evidente, guardar el
nombre en el fichero marcador para poder leerlo sin abrir ninguna base de datos, reintroduce justo lo que
el rediseño de la identidad acababa de quitar: que renombrar un tablero vuelva a escribir en el sistema
de ficheros, con lo que vuelve el mismo dato en dos sitios y la posibilidad de que discrepen. Y aunque
todo eso saliera gratis, el nombre no sirve para elegir, porque no es único: ["Cómo se elige el tablero"](spec/resolucion-del-tablero.md) declara legal
que dos proyectos de la misma máquina se llamen igual, y lo declara como una de las tres cosas que
resuelve de golpe sacar la identidad del nombre, precisamente porque lo que identifica a un tablero es su
`id` y no cómo se llama.

**La tercera: por qué tampoco una bandera que acepte el identificador.** Aquí no hay ningún obstáculo
técnico, hay algo peor: no hay ningún caso de uso que se sostenga. El identificador de ocho hexadecimales
no aparece en el trabajo diario, porque las tareas se nombran con `<PREFIX>-<n>` y ese prefijo se deriva
de `project_name`, no del `id` del tablero (["Identificadores"](spec/modelo-de-datos.md#identificadores)), así que nadie lo tiene delante ni lo teclea. El
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
argumento para conservar la bandera, y se cae solo en cuanto se lee ["Cómo se elige el tablero"](spec/resolucion-del-tablero.md) entera, porque ese
riesgo ya está aceptado en el caso general: "El precio de no tener ese freno es que un proyecto sin
puntero propio hereda el del proyecto que lo contenga, si lo hay, y se acepta a propósito". Añadir una
bandera cuyo único valor fuera evitar en un caso concreto un riesgo que la especificación acepta a
propósito en el caso general sería incoherente, y encima daría la impresión de que el caso general está
protegido cuando no lo está. Lo que sí hace visible el caso, y ya existe, es `biso where`, cuya fila
`source` nombra el directorio del que salió el puntero, y el bloque `BOARD` del mensaje de arranque, que
dice el nombre del tablero en su primera línea.

Cómo eligen su almacén las demás herramientas del espacio, y las de fuera de él, está en la sección de
`docs/ESTADO-DEL-ARTE.md` que cierra su parte 1. El resumen es que ninguna acepta el nombre legible de un
almacén para elegirlo, que todas apuntan con una ruta, y que las dos que sí admiten un nombre lo hacen
en una bandera aparte de la de la ruta y contra un registro previo que lo declara.

**Y la consecuencia que esto deja, que conviene ver antes de tocar nada.** Con las dos vías retiradas
quedan dos, y la primera, la que reconoce un directorio como tablero porque contiene `board.db`, deja de
ser una comodidad y pasa a ser **la única forma que queda de tocar un tablero al que el proyecto actual
no apunta sin escribir antes en el disco**, siempre para una sola invocación y con `-C` apuntando a su
directorio. Cuando se decidió, se justificó solo con que ahí no hay nada que adivinar; ahora carga además
con este trabajo. Se nota en el puntero perdido, donde ["Cómo se elige el tablero"](spec/resolucion-del-tablero.md) ofrecía dos remedios, `--board`
apuntando al tablero o `biso init --at`, y ha quedado con uno solo, que escribe un puntero nuevo en el
proyecto. Quien en el futuro quiera endurecer esa vía, por ejemplo exigiéndole también el marcador
`<id>.id` como hace la búsqueda por identificador, tiene que saber que estaría cerrando la última puerta
que queda, y que la propia sección la deja abierta a propósito para que `biso doctor --fix` pueda
devolver un marcador que falte.

### El control de versiones, la instantánea y los códigos que salieron de ahí

Todo este apartado sale de una sola pregunta que la especificación tenía mal contestada: ["Lo que se deja fuera a propósito"](spec/fuera-de-alcance.md)
decía que lo que cruza a otra máquina es la instantánea, y ningún comando le daba una vía para
cruzar, porque el repositorio donde vive está fuera del proyecto y no tiene remoto.

**Por qué el sistema de control de versiones es configurable, y no git a secas.** Cablear git habría
dejado sin historial a cualquiera que use otro sistema, y no por una limitación real: lo único que `biso`
necesita de él son cuatro operaciones, ver si un directorio está en un repositorio, crear uno, guardar una
revisión y publicarla. La clave `vcs` de ["Configuración de máquina"](spec/invocacion.md#configuración-de-máquina) las nombra, con `git` por defecto
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
adivinarlo mirando el disco es imposible y `snapshot` lo pregunta (["`biso snapshot`"](spec/cmd/snapshot.md)). El caso que
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

**Por qué `snapshot` no toma ningún acceso exclusivo.** Es una lectura, y ["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables)
promete que una lectura nunca hace fallar a una escritura. Si tomara el acceso exclusivo de las
escrituras, una copia podría hacer terminar con error a un `biso set` que llegara a la vez, que es un daño
sobre el trabajo diario. Sin él, el único desenlace malo es que dos instantáneas simultáneas choquen al
guardar la revisión, y ese daño cae sobre una copia que se puede repetir: sale con el código 7 y su
mensaje dice que basta volver a llamar. Un acceso exclusivo propio de este comando evitaría también ese
choque, y se descartó por lo que cuesta especificar bien un fichero de bloqueo persistente y la limpieza
de los que deja atrás un proceso que muere, para un caso que solo ocurre si dos sesiones terminan en el
mismo segundo sobre el mismo tablero.

**Por qué el daño de la base de datos estrena el código 10 en vez de compartir el 8.** El 8 promete un
remedio, `biso init` crea el tablero, y con la base de datos dañada ese remedio no arregla nada: hay que
reconstruir desde una instantánea. Tres situaciones con tres remedios no pueden compartir número si el
principio de ["Los principios"](spec/principios.md) dice que quien llama ramifica sobre el número sin leer el mensaje.
Y como el remedio ahora tiene un comando que lo hace, el mensaje lo nombra en vez de decir "restaura de
una copia": `biso init --from` reconstruye en el sitio, adoptando el `id` del marcador, y para eso hubo
que declarar que un directorio cuya base de datos no abre no cuenta como tablero accesible.

**Por qué el identificador duplicado estrena el 11 en vez de reusar el 5.** El 5 es la referencia que
encaja con más de una entidad, y en la práctica siempre habla de una tarea: quien lo recibe afina la
referencia. Aquí no hay ninguna referencia que afinar, hay dos directorios en el disco con la misma
identidad, y el remedio es renombrar o quitar uno. Compartir el número habría obligado a leer el mensaje
para saber cuál de los dos remedios aplicar, que es justo lo que los códigos existen para evitar.

**Por qué las columnas se miden en celdas de terminal.** Los títulos son texto libre en UTF-8, y contar
puntos de código desalinea la tabla en cuanto aparece un acento combinante, un ideograma o un emoji,
porque lo que suman en pantalla no es lo que suman como caracteres. La celda es la única unidad que
alinea de verdad, y medirla no contradice la prohibición de mirar el terminal de
["Interactividad, terminal y color"](spec/salida-y-terminal.md#interactividad-terminal-y-color): la anchura de un carácter es una propiedad de Unicode, igual en cualquier máquina, mientras
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
una sola vez, en un apartado de ["`biso snapshot`"](spec/cmd/snapshot.md) que vale para los dos sistemas, y `custom`
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

Por stdout no va nunca nada, porque ["stdout, stderr y qué va en cada uno"](spec/salida-y-terminal.md#stdout-stderr-y-qué-va-en-cada-uno) lo reserva para los datos. Y con `--json`
tampoco va por stderr, donde ["Los errores en JSON"](spec/contrato-json.md#los-errores-en-json) pone el sobre de error: ahí las líneas entran en
el propio sobre, en `data.vcsOutput` o en `error.vcsOutput`. Eso añadió una quinta clave de detalle a la
tabla de esa sección, que ya se gobierna aparte del contrato de estabilidad de las salidas de
datos precisamente para poder crecer.

**`--quiet` no las suprime**, por el mismo motivo por el que ["Notas y avisos"](spec/salida-y-terminal.md#notas-y-avisos) nunca suprime un `warning:`.
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

**Y al escribirlo apareció un error del documento.** ["Configuración de máquina"](spec/invocacion.md#configuración-de-máquina) decía que `{files}` se sustituía por
"los ficheros de la instantánea", pero la revisión lleva tres ficheros (`snapshot.ndjson`, `board.json` y
el marcador `<id>.id`) mientras que la clave `files` del JSON de `biso snapshot` enseña solo los dos que
ese comando escribe. Eran dos conjuntos distintos con nombres casi iguales, y no había manera de saber si
`{files}` eran dos o tres. Son los tres, los mismos que entran en la revisión, y ahora esas secciones
lo dicen y se nombran la una a la otra.

## El origen de la cifra de 25 milisegundos

El tope de bytes del mensaje de arranque (["El presupuesto del mensaje de arranque"](#el-presupuesto-del-mensaje-de-arranque)) trae su medida. ["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque),
25 milisegundos de reloj para `biso ls` y `biso prime` sobre un tablero de
300 tareas, no la tenía escrita en ningún sitio, y esta sección es esa medida.

**Medido en la misma máquina que documenta `docs/ESTADO-DEL-ARTE.md`** (Apple M3 Max, macOS 26.5.2, 300
iteraciones, ["El coste de arranque y el coste de contexto"](ESTADO-DEL-ARTE.md#12-el-coste-de-arranque-y-el-coste-de-contexto)). El suelo del sistema operativo para arrancar cualquier
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

**El almacén real ya está medido, y esto es lo que le pasó a esa cifra de margen** (["El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`"](#el-controlador-de-sqlite-es-moderncorgsqlite-sin-cgo)). Con
la base de datos SQLite de verdad y el controlador elegido, leer el tablero de 300 tareas cuesta **14,5
milisegundos** en un portátil macOS y **3,2 milisegundos** en Linux. O sea que las tres veces de margen
se quedan en **una vez y siete décimas** en el portátil, y crecen a casi ocho veces en Linux. El
presupuesto se cumple en las dos, y desde ahora la frase "sobra tres veces" hay que leerla como una
cifra de la carga de JSON y no del programa terminado.

**Y una corrección al suelo, que no era una constante.** Los 5,2 milisegundos del párrafo anterior son
un suelo de macOS. El equivalente medido al elegir el controlador, un binario de Go que no hace
absolutamente nada, o sea suelo del sistema más arranque de Go, da **8,1 milisegundos** de mediana en el
portátil macOS, coherente con los 7,4 que suman las dos cifras de arriba. Ese mismo binario tarda
**0,37 milisegundos** en Linux, veintidós veces menos. No toca el presupuesto, que ["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque)
amarra a la máquina que ejecuta la integración continua y no a un modelo de hardware, pero
refuerza el argumento de ["La decisión de persistencia"](#la-decisión-de-persistencia) contra tener un daemon: en el portátil el suelo se come 8,1 de los
14,5 milisegundos que cuesta la lectura real, y un daemon no puede ahorrar el suelo, porque el cliente
que hablaría con él es también un proceso.

**La cifra excluye a propósito los lenguajes interpretados.** En la misma máquina, el solo arranque de
Python 3.14 añade 24,5 milisegundos por delante de cualquier trabajo real, y el de Node 25.6 añade 33.
Un presupuesto que tuviera que cubrir ese arranque dejaría de medir la herramienta y pasaría a medir el
lenguaje, así que la cifra se fija mirando el suelo que un lenguaje compilado permite. Eso acotó la
lista de candidatos a los compilados, y el apartado siguiente cierra la elección dentro de esa lista.

---

## El lenguaje de implementación es Go

**Los dos candidatos reales eran Go y Rust**, y la elección es Go. Los dos cumplen con holgura el
presupuesto del apartado anterior en la única carga que hay medida, y los separa menos de un
milisegundo, así que la decisión no se toma por rendimiento: se toma por lo que cuesta escribir el
programa, porque es lo único que de verdad los separa aquí.

**Por rendimiento la diferencia son seis décimas de milisegundo.** Con las medidas de
["El coste de arranque y el coste de contexto"](ESTADO-DEL-ARTE.md#12-el-coste-de-arranque-y-el-coste-de-contexto), sobre el suelo de 5,2 milisegundos que cuesta arrancar cualquier proceso,
Rust añade 1,6 milisegundos y Go 2,2. Esa diferencia es el **2,4 por ciento** de un presupuesto de 25
milisegundos que sobra tres veces sobre el total medido de 8,7, que es lo que cuesta leer 300 tareas de
un JSON y no lo que costará leerlas de SQLite (["El origen de la cifra de 25 milisegundos"](#el-origen-de-la-cifra-de-25-milisegundos)). El margen puede encogerse cuando se mida
el almacén de verdad, pero las seis décimas no dependen de eso: son el arranque del propio binario, la
misma cifra fija sea cual sea el trabajo que venga después. Para calibrar cuánto es: `rg`, que es
la herramienta más rápida de las que se midieron instaladas, tarda 7,1 milisegundos, y
`git --version` tarda 12,3. Ganar seis décimas en un programa cuyo competidor de referencia gasta doce
milisegundos en imprimir su propia versión no cambia nada que un usuario pueda notar. El presupuesto,
además, **se midió con un binario de Go**, así que la cifra que la especificación exige no se extrapola
de otro lenguaje: es lo que el lenguaje elegido hizo en esa máquina, con la carga que
["El origen de la cifra de 25 milisegundos"](#el-origen-de-la-cifra-de-25-milisegundos) dice.

**Lo que decide es el ciclo de desarrollo, y en particular el ciclo de un agente.** Este documento y
[`docs/spec/`](spec/index.md) están escritos para que alguien implemente el programa entero sin preguntar, con la
comprobación frecuente que ["Por dónde empezar a implementar"](spec/por-donde-empezar.md) ordena, y ese alguien va a ser en buena parte un
agente automático. En ese modo de trabajo, el coste dominante no es el tiempo de ejecución del programa
sino **el número de vueltas entre escribir y ver el resultado**, y ahí Go gana por dos motivos
distintos. El primero es que compila muy rápido, así que cada vuelta es corta. El segundo es más
importante: el modelo de propiedad y préstamo de memoria de Rust es exactamente la clase de cosa que un
modelo de lenguaje **no acierta a la primera**, y cada fallo obliga a una ronda de corrección que no
trata del problema que se está resolviendo sino de satisfacer al compilador. Esas rondas se suman, hacen
falta revisiones constantes, y no dejan nada mejor en el producto: el mismo programa, escrito en Go, no
las paga.

**El coste de elegir Go se acepta a ojos vistas y aquí queda escrito.** El recolector de basura, que es
la objeción habitual, no tiene ningún efecto medible en una herramienta cuyo proceso vive milisegundos,
lee 300 tareas y termina, porque no llega a haber presión de memoria que recoger. El binario es mayor
que el equivalente en Rust, lo que da igual en algo que se instala una vez. Y se renuncia a las seis
décimas del párrafo anterior, que es lo que se está comprando.

**Quedaba una cosa por comprobar, y no era del lenguaje sino de su encuentro con SQLite.** La
persistencia de ["La decisión de persistencia"](#la-decisión-de-persistencia) es una base de datos SQLite, y en Go hay tres formas de hablar con ella: un
enlace con la biblioteca en C, que obliga a compilar con `cgo`; una traducción de ese código de C a Go
puro; y SQLite compilado a WebAssembly y ejecutado por un motor escrito en Go. La elección afectaba al
arranque y a cómo se distribuye el programa, y nunca a la del lenguaje, porque las tres son de Go. **Ya
está medida, y la cierra el apartado siguiente.**

### El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`

Medido el 2026-09-10 con el banco de pruebas que vive en `bench/sqlite-driver/` de este mismo
repositorio, que se rehace con dos órdenes sin argumentos. Las tablas completas, la composición del
tablero de prueba y las versiones exactas de todo están en `bench/sqlite-driver/RESULTADOS.md`; aquí va
la decisión y lo que la sostiene. Se midieron los cuatro candidatos vivos, y los cuatro llevan dentro la
misma versión de SQLite, la 3.53.4, así que ninguna diferencia de las que siguen es del motor.

**El presupuesto se cumple con los cuatro, y con mucho margen.** Sobre el tablero de 300 tareas de
["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque), en Linux el más lento tarda 3,2 milisegundos, casi ocho veces por debajo de
los 25. En un portátil macOS van de 11 a 14,5 milisegundos, y de esos 8,1 son el suelo del sistema para
arrancar cualquier proceso. Los cuatro binarios producen además una salida idéntica byte a byte en las
dos plataformas, y ninguno necesitó una sola línea de SQL distinta: mismo esquema, mismas cinco
consultas, mismos recorridos. La incompatibilidad que uno teme al elegir un controlador de SQLite no
apareció por ningún lado, y eso es en sí mismo un argumento para quedarse en la interfaz estándar
`database/sql`.

**Así que el rendimiento no decide, y lo que decide es la distribución del binario.** Es el mismo
criterio con el que el apartado anterior eligió Go frente a Rust, aplicado otra vez, y aquí la
diferencia no es de grado.

**El enlace con la biblioteca en C queda descartado, y eso es lo primero, porque descarta una familia
entera.** Es el más rápido de los cuatro en las dos plataformas y aun así se cae, por dos cosas
medidas. La primera es que desde una máquina macOS no produce un binario para Linux ni para Windows,
porque la compilación cruzada falla en `runtime/cgo` al no haber un compilador de C para el sistema
destino. La segunda es peor que un fallo de compilación: si se apaga `cgo`, el binario compila, enlaza,
arranca y muere al tocar la base de datos con el mensaje `go-sqlite3 requires cgo to work. This is a
stub`. Un fallo que la construcción deja pasar y que solo aparece al ejecutar es el peor desenlace para
algo que se reparte como binario, porque no lo ve quien lo construye sino quien lo usa. Los otros tres
candidatos compilan para `linux/amd64`, `linux/arm64` y `windows/amd64` con un `go build` y nada más
instalado, y salen estáticos.

**Entre los tres que quedan gana `modernc`, y por lo que cuesta escribir el programa.** Los separan 0,68
milisegundos, que es el 2,7 por ciento del presupuesto, así que otra vez no decide el reloj. `modernc` es
a la vez la interfaz estándar y el motor sobre el que están construidos los otros dos candidatos de Go
puro. `zombiezen` es ese mismo motor con una interfaz propia que obliga a escribir a mano lo que
`database/sql` da hecho, y su ganancia son 0,28 milisegundos. `ncruces` es el más rápido de los tres,
pero su binario es un 48 por ciento mayor y su búsqueda por texto completo hay que registrarla por
conexión, bajando por debajo de `database/sql`.

**Y esto es lo que se acepta a cambio, con nombre y número.** En macOS, y solo en macOS, `modernc` gasta
entre 2,6 y 3,5 milisegundos antes de que `main` empiece, en el arranque del paquete
`modernc.org/libc/honnef.co/go/netdb`, que lee `/etc/services` y `/etc/protocols` del disco y los
convierte en estructuras de Go para ofrecer una función de red que `biso` no usa nunca y que no se puede
desactivar. Son entre el 10 y el 14 por ciento del presupuesto tirados en cada invocación. En Linux ese
arranque baja a 0,046 milisegundos, porque el paquete no se importa allí, y eso se comprobó ejecutando
los binarios dentro de un contenedor y no leyendo el código. Se acepta porque incluso pagándolo la
mediana en macOS es de 14,5 sobre 25, y porque el remedio no está en `biso` sino aguas arriba.

**La salida de emergencia queda dicha por adelantado, para no tener que discutirla desde cero.** Si ese
peaje llegase a molestar de verdad, el relevo es `ncruces`, y el cambio cuesta una línea de `import` y
la cadena con el nombre del controlador, porque los dos hablan por `database/sql` y el resto del código
que consulta el tablero seguiría siendo el mismo. Esa es también la razón de no elegir `zombiezen`: no
es que sea peor, es que salirse de `database/sql` convierte esa salida de emergencia de una línea en una
reescritura.

**Lo que se comprobó que soporta, porque la especificación lo da por hecho.** Modo WAL, `BEGIN
IMMEDIATE` como acceso exclusivo de escritura, puntos de retorno, `busy_timeout`, claves ajenas,
`user_version`, el `integrity_check` que necesita `biso doctor`, `wal_checkpoint(TRUNCATE)`, `ANALYZE`,
consultas recursivas, el módulo JSON y las funciones de ventana. Y leer desde una segunda conexión
mientras una primera tiene una escritura abierta, que es literalmente lo que promete
["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables). La búsqueda por texto completo con FTS5, que sería la vía barata para
["La búsqueda por texto"](spec/referencias.md#la-búsqueda-por-texto), viene
puesta. Y un dato para cuando se implemente esa búsqueda: ningún controlador hace `LIKE` insensible a
mayúsculas con acentos, porque eso es lo que hace SQLite sin la biblioteca ICU.

---

## La regla de coincidencia de vocabulario

Es la pieza que cierra el principio 1, y merece contarse entera porque tiene una trampa.

**Se eligió eliminar los separadores, no colapsarlos.** Normalizar es pasar a minúsculas, quitar
diacríticos y quitar espacios, guiones y guiones bajos; comparar es igualdad. Con esa regla, `todo`,
`To-Do`, `TO_DO` y `to do` son todos `To Do`.

La alternativa era colapsar cualquier tramo de separadores a un solo espacio, y **no funciona**:
colapsando, `To-Do` se convierte en `to do` y coincide, pero `todo` no tiene ningún separador que
colapsar, se queda en `todo`, y no coincide con `to do`. Es un error fácil de cometer al redactar la
regla en prosa, porque el ejemplo que uno escribe a continuación parece cierto y no lo es.

Se eligió eliminar por dos motivos. El primero es que es lo que hacen las herramientas existentes al
escribir, así que ningún texto que hoy vale deja de valer. El segundo es que produce una función más
simple de escribir y de probar.

**Por eso la regla está escrita en la especificación como pseudocódigo con pasos numerados y no como
prosa.** Escrita en prosa volvería a poder decir dos cosas a la vez, y ya lo hizo una vez.

Una consecuencia que conviene no perder: el ejemplo canónico de un valor inválido no puede ser `Todo`,
porque `Todo` es válido. En la especificación el ejemplo es `Pending`, que no existe en ningún
vocabulario.

---

## El presupuesto del mensaje de arranque

`biso prime` sustituye por completo a las guías de instrucciones y a cualquier inyección de texto en
los ficheros de convenciones del proyecto. Ese diseño se toma de una medición concreta.

En la herramienta estudiada, las cuatro guías de instrucciones se leyeron 69 veces en seis días y
suman **201.500 bytes, el 27,5% de toda la salida** que la herramienta devolvió a los agentes, más que
sus comandos más usados juntos.

| Lectura | Bytes | Cuándo la exige ese diseño |
|---|---:|---|
| Guía general | 2.365 | al empezar |
| Guía de creación | 4.014 | antes de crear |
| Guía de ejecución | 3.807 | antes de planificar o actualizar |
| Guía de finalización | 2.719 | antes de terminar |
| **Ciclo completo** | **12.905** | una sesión que crea, trabaja y cierra |

A eso hay que sumar la inyección de instrucciones en el fichero de convenciones del repositorio, que
se paga en todas las sesiones aunque no se toque el tablero.

El mensaje de `biso prime` mide **4.818 bytes**, 3.327 de parte fija y 1.491 de resumen del tablero
(["El presupuesto de tamaño"](spec/presupuestos.md#el-presupuesto-de-tamaño)), contra un tope duro de 5.120 repartido en dos partes de 3.456 y
1.664.

| Magnitud | Herramienta estudiada | `biso` |
|---|---:|---|
| Peor caso por sesión, con ciclo completo | 12.905 bytes | 4.818 bytes |
| Media medida por sesión | 3.358 bytes | 4.818 bytes |
| Lecturas obligatorias por sesión | entre 1 y 4 | 1 |
| Contexto gastado en sesiones que no tocan tareas | la inyección en el fichero de convenciones | 0 |

**La media sube ligeramente, y conviene decirlo en vez de esconderlo.** Lo que cambia es otra cosa: el
coste pasa a ser fijo, conocido y acotado por una prueba, en vez de depender de cuántas guías decida
leer el agente, y el peor caso cae a menos de un tercio. El ahorro grande no está aquí, está en que la
salida de las escrituras deje de ser un eco y en las llamadas que desaparecen al fusionar el ciclo.

**El tope es una prueba de la suite, no un objetivo.** Y el reparto en dos partes existe para que el
resumen del tablero, que crece con el tablero, no pueda comerse el sitio de las reglas.

**El reparto entre las dos partes cambió con el modelo de estados, y el total no.** La parte fija
sube de 3.072 a 3.456 bytes porque el bloque `COMMANDS` gana las dos órdenes nuevas del ciclo, `ask` y
`answer`, y el bloque `RULES` gana una regla más, la undécima, sobre esos verbos y sobre qué
significa una tarea asignada; el resumen del tablero baja de 2.048 a 1.664.
Eso solo es seguro de hacer porque, a la vez, el orden de recorte del resumen deja de estar incompleto:
antes nombraba un solo bloque y decía "antes que cualquier otra cosa" sin nombrar ninguna otra, así que
un tablero con muchas tareas en curso podía rebasar el tope sin que hubiera una conducta definida para
ese caso. Con los cinco pasos completos que trae ahora ["El presupuesto de tamaño"](spec/presupuestos.md#el-presupuesto-de-tamaño), el resumen ya no
crece sin límite, y darle menos sitio cuesta filas mostradas, no correcciones. La parte fija, en
cambio, no se puede recortar sola: o cabe entera o hay que quitar contenido a mano, así que es la parte
que necesita más margen.

**El tope total, 5.120 bytes, no se mueve, y el motivo no es de contrato.** El contrato de estabilidad
solo obliga desde la versión 1.0, que todavía no está publicada, así que subir el tope no rompería
ninguna promesa hecha a nadie. La razón es de fondo: un tope que se sube cada vez que aprieta deja de
ser un tope, y su valor entero está en que obligue a elegir qué entra en el mensaje y qué se relega a
`--help`. Por eso, si al escribir el texto real de ["La salida literal"](spec/cmd/prime.md#la-salida-literal) los números no cupieran, lo que se
recorta es contenido, no el tope.

**Esto no es "el tope nunca sube", es "el tope sube solo cuando reducir ya no es posible sin perder
algo".** La medida de hoy, 4.818 de 5.120 bytes, tiene 302 de margen: nunca hizo falta apretar para
caber, así que esta regla no se ha puesto a prueba todavía. Pero el margen que de verdad manda no es
ese, sino el de la parte fija, que con las banderas nuevas `--check-dod` y `--uncheck-dod` en la
rejilla de `FIELD FLAGS` ha bajado a **129 bytes** de los 3.456: cualquier texto nuevo en el bloque
fijo tiene que caber ahí, no en los 302 del total. Si en el futuro un comando nuevo obliga a
recortar el bloque fijo (`COMMANDS`, `FIELD FLAGS`, `RULES`) y esa reducción sale limpia, sin perder
información que un agente necesite para arrancar bien, es que había margen y el tope hizo su trabajo.
Pero si reducir más solo se puede ya a costa de quitar algo así, mantener el tope fijo deja de ser
disciplina y pasa a ser dañar el mensaje a propósito; en ese punto, subirlo es lo correcto.

---

## El porqué de reglas concretas

Cada entrada dice la sección de la especificación a la que corresponde.

**["El algoritmo de coincidencia"](spec/vocabularios.md#el-algoritmo-de-coincidencia), por qué no hay coincidencia por prefijo ni por parecido al resolver una referencia.** Una regla
que adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

**["Qué valida cada filtro, y contra qué"](spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué), por qué las tareas archivadas cuentan en los conjuntos de etiquetas, personas e hitos contra
los que validan los filtros.** Es lo que impide que un filtro que hoy funciona deje de funcionar
mañana por archivar la última tarea que lo usaba.

**["Qué valida cada filtro, y contra qué"](spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué), por qué las etiquetas, las personas y los hitos no tienen vocabulario cerrado al escribir, pero
sus filtros sí validan.** No tienen vocabulario cerrado porque su utilidad es que se puedan inventar
sobre la marcha. Y validar al leer no es una asimetría con la escritura: es la aplicación del
principio 1, que dice que un filtro que no puede encajar con nada es un error y no una respuesta
vacía.

**["Qué valida cada filtro, y contra qué"](spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué), por qué el hito validaba y ahora valida.** El hito era la excepción, con un `--milestone` que
nunca fallaba y devolvía una lista vacía ante cualquier errata. La razón que se daba era que no hay
entidad de hito, pero el argumento del párrafo de arriba no distingue en nada al hito de la etiqueta:
las dos son texto que quien llama se inventa, ninguna de las dos se declara antes de usarla, y en las
dos una errata al filtrar produce exactamente el fallo que el principio 1 existe para evitar. Lo que
hacía falta no era una entidad, sino un conjunto contra el que comparar, y ese conjunto ya estaba
ahí sin que nadie lo escribiera: los hitos que las tareas llevan de hecho. Con eso, **el hito era la
última asimetría del principio 1 dentro de la especificación, y deja de serlo.**

La excepción que queda, la bandera `--unchecked`, no es de la misma clase y por eso se conserva. La
diferencia está en quién decide: una asimetría es el programa el que decide callar, sin que quien
llama lo sepa ni pueda evitarlo, mientras que `--unchecked` la pide quien llama, en la misma línea de
comandos, y quien la escribe está declarando que acepta una lista vacía sin garantía. Un
comportamiento que se pide no engaña a nadie. Por eso `--unchecked` pasa a apagar también la
comprobación del hito: dejar el hito fuera de la escapatoria declarada crearía una asimetría nueva
justo al quitar la vieja.

**["Los tres mensajes de \"no la encuentro\""](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), por qué son distintos.** Porque las tres situaciones
piden acciones distintas de quien llama: corregir la sintaxis, dejar de buscar, o mirar en el archivo.

**["La regla"](spec/familias-de-banderas.md#la-regla), por qué cada clase de campo tiene las variantes de bandera que tiene, y por qué ninguna se
llama con el nombre desnudo del campo.** Antes había cuatro variantes por cada campo de lista
(`--campo` añadía, `--set-campo` sustituía, `--rm-campo` quitaba, `--clear-campo` vaciaba), y las
demás clases de campo tenían el subconjunto de esas cuatro que tenía sentido para su forma de dato:
un bloque de prosa no tiene elementos que quitar de uno en uno, así que no tenía `rm-`; un mapa de
claves se manipula por clave y no por posición; un escalar solo se fija o se vacía; los comentarios,
en aquel momento, solo admitían añadir. **Esa última pieza ya no es cierta**: ["Borrar o corregir la fecha de un comentario"](#borrar-o-corregir-la-fecha-de-un-comentario), más abajo en
este documento, añade dos operaciones más, y la forma vigente de la clase está en
["Comentarios"](spec/familias-de-banderas.md#comentarios) y no aquí. Lo que sigue siendo cierto, y es lo que este párrafo quería decir, es
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
  acumulado (["Repetición y listas separadas por comas"](spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas)), que es una segunda pieza de comportamiento no obvio
  encima del nombre. Sustituir esas dos listas se hace vaciando y añadiendo en la misma llamada
  (["Orden de aplicación dentro de una escritura"](spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura)), que ya hacía falta declarar para todo lo demás y cubre el mismo caso sin una
  bandera más que aprender.
- **`--check` y `--uncheck` pasan a `--check-ac` y `--uncheck-ac`, simétricos con `--check-dod` y
  `--uncheck-dod`.** La forma anterior reservaba el nombre desnudo para los criterios de aceptación y
  obligaba a la definición de hecho a llevar el sufijo, una asimetría que la propia especificación
  declaraba como excepción sin más justificación que la de ser la primera lista de las dos. Bajo la
  regla nueva no hay sitio para reservar un nombre desnudo a nada, así que la asimetría desaparece
  sola en vez de quedar documentada como caso especial.

**Esto deja pendiente el mensaje de arranque.** El bloque `FIELD FLAGS` de
["La salida literal"](spec/cmd/prime.md#la-salida-literal) enseñaba antes una sola bandera por campo (la que añadía) y una regla en `RULES`
explicaba cómo derivar las otras tres; ese ahorro de espacio dependía exactamente del mecanismo que
se acaba de retirar. Medido: los nombres de las banderas de campo explícitas, sin ninguna tabla que
las derive, ocupan por sí solos más de 700 bytes, frente a los 436 del bloque `FIELD FLAGS` actual
completo, y la parte fija del mensaje de arranque solo tiene 129 bytes de margen
(["El presupuesto del mensaje de arranque"](#el-presupuesto-del-mensaje-de-arranque)). Cómo se enseñan las banderas nuevas dentro de ese presupuesto
queda sin decidir a propósito, y no se ha tocado `cmd/prime.md` ni `presupuestos.md` hasta que se
decida.

**["Selectores de criterios"](spec/familias-de-banderas.md#selectores-de-criterios), por qué quitar un criterio de aceptación toma un selector y no un texto.** Porque quitarlo por
su texto exacto es más frágil que quitarlo por su clave.

**["Campos externos"](spec/familias-de-banderas.md#campos-externos), por qué no existe una bandera que sustituya el mapa de campos externos entero.** Fijar una clave
ya es sustituir su valor, así que una segunda bandera para lo mismo solo serviría para equivocarse. Y
una que sustituyese el mapa entero con la sintaxis `clave=valor` sería una forma silenciosa de borrar
la identidad externa de una tarea al escribir otra.

**["La salida literal"](spec/cmd/prime.md#la-salida-literal), por qué el bloque de tareas en curso del mensaje de arranque no tiene límite.** Porque en un
tablero sano son pocas.

**["`biso init`"](spec/cmd/init.md), por qué el puntero del proyecto es la única cosa que `init` escribe fuera del tablero.** Sin
ella, un tablero creado en otra ubicación no lo encontraría ningún comando posterior.

**["`biso init`"](spec/cmd/init.md), por qué se retira la regla posicional que guardaba el estado activo como el penúltimo de
`--statuses`.** La regla estaba rota, y la contradicción que la delata vive en el propio documento: el
tablero de ejemplo era `Ideas, To Do, In Progress, Blocked, Done`, cuyo penúltimo es `Blocked`, y había
un ejemplo literal de `biso init` que lo creaba así, mientras que tanto la salida de `biso config list`
como el esquema JSON de `biso prime` declaraban que el estado activo de ese mismo tablero era
`In Progress`. Las dos cosas no podían ser ciertas a la vez, y la regla solo parecía funcionar porque
el tablero por defecto tenía justo tres estados. Se sustituye por tres banderas explícitas,
`--initial-status`, `--active-status` y `--terminal-status`, con el mismo argumento de ["El algoritmo de coincidencia"](spec/vocabularios.md#el-algoritmo-de-coincidencia): una regla
que adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

**["`biso init`"](spec/cmd/init.md), por qué el tablero por defecto no trae un estado `Ideas`.** Un estado `Ideas` no dice nada que
no diga ya estar sin asignar, que se consulta con `biso ls --unassigned`. El matiz que sí aporta,
"esto quizá no lo hagamos nunca", tiene ya una decisión con evidencia detrás en ["Lo que se miró de ese diseño anterior y se descarta"](#lo-que-se-miró-de-ese-diseño-anterior-y-se-descarta): se
cubre con un tipo más del vocabulario que ya existe y no con un estado. Y hay un motivo peor para no
ponerlo por defecto: si `Ideas` fuera el estado inicial, toda tarea nueva nacería ahí, y el bloque
`NEXT UP` del mensaje de arranque mezclaría "algún día quizá" con "hay que hacerlo", que es justo la
distinción que ese bloque existe para hacer.

**["`biso new`"](spec/cmd/new.md), por qué existe `--start` al crear una tarea.** Evita que crear una tarea para ponerse con ella
en el mismo minuto cueste dos llamadas. Es el principio 5 aplicado a un caso medido.

**["`biso new`"](spec/cmd/new.md), por qué `--comment` funciona al crear.** Por lo mismo: una tarea que nace con un comentario
cuesta una llamada.

**["`biso ls`"](spec/cmd/ls.md), por qué el filtro de etiquetas es el único que combina sus valores con "y".** Porque el uso
normal de varias etiquetas es acotar, no ampliar.

**["`biso ls`"](spec/cmd/ls.md), por qué el filtro de dependencias es `--blocked` y `--not-blocked`, y no `--ready`.** El hecho
que se calcula es uno solo, que alguna dependencia esté sin terminar, así que se nombra una vez y su
negación se forma con el mismo prefijo que los otros dos pares booleanos de `biso ls`,
`--waiting`/`--not-waiting` y `--active`/`--not-active`. El nombre `ready` sobraba por dos motivos
distintos. El primero es que hacía viajar el mismo hecho dos veces en el JSON, como `ready` y como
`blocked`, y campos que dicen lo mismo acaban divergiendo. El segundo es que prometía más de lo
que cumplía: miraba solo dependencias, así que `biso ls --ready` devolvía también las tareas aparcadas
en una pregunta, que es justo lo que un agente no puede coger. De los dos nombres sobrevive `blocked`
porque ya tiene entrada propia en la tabla de ["Vocabulario de esta especificación"](spec/vocabulario.md), porque da nombre al
término `urgency.blocked` de la fórmula de urgencia, y porque nombra el hecho que de verdad se calcula.
Y el nombre nuevo tampoco promete estar lista para trabajar, porque ninguna bandera sola puede: eso
son varios filtros, y cuántos depende de qué se busque. Descartar lo bloqueado y lo aparcado son dos,
`--not-blocked --not-waiting`; quien quiera además tarea sin empezar añade `--not-active`, que es el
filtro con el que el propio mensaje de arranque describe su bloque `NEXT UP` (["La salida literal"](spec/cmd/prime.md#la-salida-literal)); y quien la quiera
sin dueño, `--unassigned`. El cambio quita una clave del JSON y renombra una bandera,
que son las dos cosas que el ["contrato de estabilidad"](spec/estabilidad.md) promete no tocar nunca, y por eso se hace
ahora: ese contrato obliga a partir de la versión 1.0 y todavía no hay ninguna versión publicada.
Después de 1.0 esta limpieza ya no se podría hacer.

**["`biso set`"](spec/cmd/set.md), por qué `set` no repite en su tabla las banderas de campo.** Porque repetirlas invitaría a que
divergieran, que es como se rompen los documentos largos.

**["`biso set`"](spec/cmd/set.md), por qué la línea de estado encoge cuando la tarea no tiene criterios, en vez de imprimir un
guion como hace el listado.** Las dos salidas parecen contradecirse y no lo hacen, porque no son la
misma clase de cosa. El listado de ["`biso ls`"](spec/cmd/ls.md) es una tabla: sus columnas se rellenan al ancho del valor más
largo de la llamada, así que una celda vacía tiene que ocupar su sitio o las filas de abajo se
descolocan, y para eso está el guion. La línea de estado sale una por tarea afectada, sin ancho
compartido y sin nada que alinear debajo, de modo que un hueco no descoloca nada y un guion solo
añadiría un símbolo más que interpretar. Quien quiera los contadores siempre, estén las listas vacías
o no, pide `--json`, que trae los cuatro como números.

**["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](spec/cmd/verbos-del-ciclo.md), por qué `finish` avisa de los criterios sin marcar y no lo impide.** Un criterio puede haber
quedado obsoleto, y un comando que no deja cerrar empuja a rodearlo con `set`, que es como se aprende
a esquivar una herramienta. Para quien quiera la política dura está `--strict`.

**["`biso export`"](spec/cmd/export.md), por qué `export` no hereda los valores por defecto de `ls`, y por qué sale con código 6 y no
con 0 cuando salta una tarea ilegible.** Porque exportar de más nunca hace daño y exportar de menos en
silencio arruina una copia de seguridad. Es el único comando cuyo propósito es no perder nada, y por
eso es la única excepción a la regla general de las lecturas de conjunto.

**["`biso help`"](spec/cmd/help.md#biso-help), por qué `help` funciona sin tablero.** Porque es lo primero que alguien ejecuta cuando algo
no va.

**["Repetición y listas separadas por comas"](spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas), por qué los campos de texto largo no se parten por comas.** Porque una coma dentro de una frase
es normal, y partir por ella convertiría una descripción en varias.

**["Las fechas"](spec/modelo-de-datos.md#las-fechas), por qué las fechas se pueden fijar al importar y no en el uso normal.** Sin esa excepción no se
puede importar el histórico de otro sistema conservando cuándo pasó cada cosa, que es el tercer
requisito de ["Cuatro requisitos aprendidos de otras herramientas"](#cuatro-requisitos-aprendidos-de-otras-herramientas), en este mismo documento.

---

## Decisiones de detalle que cuesta reconstruir

**Una tarea sin quien la reporte es válida.** El campo `reporter` toma la identidad configurada al
crear la tarea, y si no hay ninguna se queda vacío **sin avisar**. Es deliberadamente distinto de los
otros cinco sitios donde hace falta una identidad (tabla de ["Variables de entorno"](spec/invocacion.md#variables-de-entorno)): el
filtro `--mine` falla, la autoasignación de `start` avisa, y el autor de un comentario, `biso ask` y
`biso answer` son un error. La razón es que un tablero de una sola persona no tiene por qué
configurar su identidad solo para poder crear tareas.

**`biso export --json` es un error y no una bandera sin efecto.** `export` es el único comando cuya
salida ya es JSON sin pedirlo, en forma de un objeto por línea, mientras que `--json` significa el
sobre único que imprimen todos los demás. Son dos formas distintas, y aceptar la bandera en silencio
dejaría en duda cuál de las dos sale.

---

## El juego de caracteres de un token

La especificación no restringía ningún carácter en `labels`, `assignees`, `references`,
`documentation`, `dependencies`, `modifiedFiles` ni en una clave de `ext`: la única regla escrita era
que una coma dentro de un valor se escapa con `\,`, lo que de hecho dejaba pasar espacios, saltos de
línea y cualquier símbolo. `labels` y `assignees` se imprimen en las columnas de ancho fijo de
`biso ls` y `biso prime`, así que un espacio o un salto de línea dentro de uno de esos tokens rompía
la tabla sin que hubiera forma de distinguir, al leerla, un token con espacio de dos tokens
separados. Hacía falta cerrar el alfabeto en algún punto, y la pregunta era dónde.

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

**La decisión.** Cerrar el alfabeto de `labels` y `assignees` a letras y dígitos Unicode más los
símbolos `- _ . : @`, sin espacio, siguiendo a Taskwarrior y Jira y no a GitHub: biso es un programa
que un agente maneja tecleando líneas de comandos, no un formulario web, así que cada etiqueta con
espacio sería una comilla que ese agente tendría que acordarse de poner siempre, para ganar exactamente
lo mismo que ya ofrecen `-` y `_`. Y olvidar la comilla no siempre falla alto: según qué banderas haya
alrededor, la palabra suelta puede convertirse en un argumento inesperado (el caso bueno, error 2) o
colarse donde no tocaba. Cerrar el alfabeto quita el problema de raíz en vez de pedir disciplina.

Cerrar el alfabeto de estos dos campos tiene una consecuencia que también hay que anotar: **el escape
de coma deja de aplicarles**. La única razón para escapar una coma es poder meterla como contenido
literal de un valor, y una coma no está en el alfabeto cerrado de `labels` ni de `assignees`, así que
ahí nunca hay una coma legítima que escapar. El escape sigue haciendo falta para `references`,
`documentation` y `modifiedFiles`, que siguen siendo texto libre.

**Por qué `references`, `documentation`, `dependencies` y `modifiedFiles` quedan fuera.** Los tres
primeros de la lista de campos de lista con coma que no son `labels` ni `assignees` guardan contenido
cuyo alfabeto no lo decide biso: una referencia o una documentación pueden ser una URL, y un fichero
tocado es una ruta del sistema de ficheros. Cerrarles el alfabeto dejaría fuera casos legítimos
(`/`, `?`, `#` de una URL; `/` de una ruta) a cambio de nada, porque ninguno de los tres se imprime en
una columna de ancho fijo con otros de su misma clase de la forma en que lo hacen las etiquetas.
`dependencies` no necesita una regla nueva porque ya tiene la suya: cada elemento es un `<ref>` y lo
gobierna entera la gramática de ["Cómo se resuelve una referencia a una tarea"](spec/referencias.md), que ya distingue un identificador
mal formado de una consulta de texto libre.

**Por qué la clave de `ext` es un tercer alfabeto y no el mismo que `labels`.** Una clave de `ext`
tiene la forma de un espacio de nombres con punto (`trello.card`, `github.issue`), así que necesita el
punto y admite `-`/`_` para el segmento, pero no tiene ningún uso documentado para `@` ni para `:`. Y
no puede admitir `=`, porque `--ext <clave>=<valor>` ya usa ese carácter para separar la clave del
valor: permitirlo dentro de la clave haría ambiguo dónde termina una y empieza el otro en
`--ext a=b=c`.

**Por qué el código de salida es 2 y no 3.** El código 3 (["Códigos de salida"](spec/codigos-de-salida.md)) es para un valor
sintácticamente correcto que el tablero no reconoce, y `labels` y `assignees` no tienen vocabulario
cerrado al escribir (["Qué valida cada filtro, y contra qué"](spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué)): cualquier texto que cumpla el alfabeto es válido sin que el
tablero lo declare antes. Un carácter fuera del alfabeto no es un problema de reconocimiento sino de
forma, la misma clase de fallo que un identificador mal formado, que ya es código 2. Tratarlo como
código 3 habría mezclado dos preguntas distintas bajo el mismo número: "¿es sintácticamente válido?" y
"¿el tablero lo tiene?".

---

## Dónde se anuncia la clave de un criterio recién creado

**La decisión.** Cuando `--add-ac` o `--add-dod` crea un elemento en un comando distinto de
`biso new`, la línea de estado por defecto (["`biso set`"](spec/cmd/set.md#salida) y los verbos del ciclo) añade `added ac #<clave>`
o `added dod #<clave>` al final, y solo cuando la llamada crea de verdad algo. `biso new` no lo
anuncia nunca, ni siquiera cuando crea criterios a la vez que la tarea. En `--json`, `acAdded` y
`dodAdded` se añaden a `data.tasks` del esquema `task.write`, presentes siempre (vacíos si no se creó
nada) en todos los comandos que comparten ese `kind`, incluido `new`.

**Por qué.** El principio 4 (["Los principios"](spec/principios.md)) dice que la salida por defecto de una escritura es lo que
quien llama no sabía, nunca el eco de lo que acaba de escribir. Sobre una tarea que ya existía, el
contador de claves de sus criterios viene de antes, y quien llama no puede saberlo sin leer la tarea
primero: es justo el dato que ese principio manda enseñar, y por eso va en la misma línea de estado
que ya enseña el resto de derivados (`ac X/Y`, `dod X/Y`, `urgency`), no en un sitio aparte. Sobre una
tarea recién creada con `biso new`, en cambio, el contador de cada lista siempre empieza en 1, así que
la clave de cada `--add-ac` es el mismo orden en que se escribieron las banderas
(["Los criterios y sus claves estables"](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables)): quien llama ya lo sabe, y anunciarlo sería el eco que el principio 4
prohíbe.

**Alternativas descartadas, y por qué.** Una nota de stderr, con la misma forma que `note:` o
`warning:` de ["Notas y avisos"](spec/salida-y-terminal.md#notas-y-avisos), encajaría con la convención existente, pero separaría este
dato del resto de derivados que viven en la línea de estado por la misma razón exacta (no se pueden
saber sin leer la tarea), y lo dejaría invisible para cualquier consumo que descarte stderr, sin que
haga falta: el contrato de estabilidad (["El contrato de estabilidad"](spec/estabilidad.md)) ya dice que el texto exacto de la
línea de estado puede cambiar entre versiones menores, así que alargarla no rompe ninguna promesa. Una
línea aparte en stdout, por la misma razón, multiplicaría las líneas de salida por tarea sin necesidad:
el dato cabe en la línea que ya existe.

---

## Borrar o corregir la fecha de un comentario

**La decisión.** `biso` añade banderas dedicadas para comentarios ya escritos: `--rm-comment <sel>`
borra uno o varios enteros, y `--set-comment-date <sel>=<instante>` corrige solo su fecha. El
selector es el mismo de ["Selectores de criterios"](spec/familias-de-banderas.md#selectores-de-criterios), con la clave del comentario en vez de la del
criterio. **No existe, y no va a existir, ninguna forma de editar el cuerpo o el autor de un
comentario ya escrito.** Cada comentario recibe además una clave estable, igual que un criterio de
aceptación, porque sin ella no hay forma de señalar cuál se quiere borrar o corregir sin que se mueva
al borrar otro.

**El caso medido que lo motiva.** En una migración real con otra herramienta que guarda los
comentarios como texto en un fichero, se detectaron agentes editando ese fichero a mano para que las
fechas de varios comentarios coincidieran entre sí, precisamente porque la herramienta no ofrecía
ninguna vía legítima para corregir la fecha de un comentario ya escrito fuera de una importación
inicial. `biso` ya resuelve la mitad del problema: cualquier fecha se puede fijar al crear una tarea
por lote (["Cuatro requisitos aprendidos de otras herramientas"](#cuatro-requisitos-aprendidos-de-otras-herramientas), en este documento). Pero esa vía sirve para
poblar un tablero vacío, no para corregir una tarea que ya existe: un `id` ya ocupado falla al
importar (["`biso new`"](spec/cmd/new.md)), así que hoy no hay ninguna forma de arreglar una fecha equivocada en una tarea
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
