# Los principios, y como se mantiene la especificacion

## La evidencia detrás de los siete principios

Los principios de ["Los principios"](../spec/principios.md) están ahí enunciados sin su procedencia,
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

**Principio 3, que ningún flag depende de una regla que haya que conocer de antemano.** Se
midieron seis casos de agentes usando la variante destructiva de forma repetida creyendo que
añadían: el flag de plan aplicado hasta cinco veces sobre la misma tarea, el de referencias tres
veces sobre otra, y un caso en el que un agente ejecutó sobre una misma tarea `--ref`, `--ref`,
`--ref`, `--add-ref`, `--remove-ref` y `--clear-refs`, que es alguien probando a ver cuál de las seis
hace lo que quiere. Ninguna de esas llamadas dio error, y el daño es silencioso: cada una borró lo
que había escrito la anterior. La primera forma de este principio resolvía el problema con una regla
única, "el nombre desnudo añade", que había que conocer de antemano para no adivinar; se sustituyó
después por dar a cada operación su propio verbo explícito, sin ninguna regla que aprender, por el
motivo que cuenta ["La regla"](../spec/familias-de-flags.md#la-regla) de `familias-de-flags.md`.

**Principio 4, que la salida por defecto de una escritura es lo que quien llama no sabía.** De las 237
creaciones medidas, 215 llevaban un flag que devolvía la ficha entera de la tarea recién creada, y
sumaron 179.369 bytes, casi la cuarta parte de toda la salida del estudio, sin dar el único dato que
el agente no tenía, que es el identificador. La mediana de una creación pasa de 154 bytes sin ese
flag a 1.333 con ella, un factor de 8,7.

**Principio 5, que un gesto del flujo de trabajo es un comando.** De las 159 ediciones medidas, 112
cambian exactamente un campo, hay 82 pares de ediciones consecutivas sobre la misma tarea, y el ciclo
de vida típico cuesta entre seis y doce llamadas. **Y no era una limitación de la herramienta**: se
comprobó que una sola llamada aceptaba el cierre entero y funcionaba. La fragmentación venía de que
sus guías presentaban el trabajo como una lista numerada con un comando por paso.

**Principio 6, el lote con validación previa.** Una migración real creó 242 tareas de una en una, y
cuando el entorno bloqueó los comandos compuestos que hacían falta, los agentes acabaron escribiendo
85 ficheros de tarea a mano, 76 ediciones y 9 creaciones, que es exactamente lo que la herramienta
prohíbe en la instrucción que ella misma inyecta en cada proyecto.

**Principio 7, que la salida no depende del terminal.** El flag `--plain` de la herramienta
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
  comandos de documentos y de decisiones no se usaron ni una sola vez en seis días. Por eso `biso`
  conservó el hito como campo de la tarea en vez de crear una entidad con ciclo de vida propio: sin
  comando que lo creara, sin clave de configuración que lo declarase, sin fecha ni estado propios.
  **Esta pieza ya no es cierta: `milestone` se retiró entero, no solo como entidad.**
  ["Se retiran `project` y `milestone`"](detalles.md#se-retiran-project-y-milestone), más abajo en este
  documento, explica por qué el campo dejó de hacer falta y qué ocupa su lugar.
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
  del propio proyecto, y esa premisa se descarta entera con la [decisión de persistencia](persistencia.md#la-decisión-de-persistencia):
  el tablero vive en un almacén aparte, y lo único que se versiona es la instantánea de texto que
  `biso snapshot` escribe cuando quien llama lo pide, un solo commit por invocación y nunca uno por
  cada escritura de tarea.
- **La visibilidad entre versiones del proyecto.** Se habían medido cinco fallos reales de las copias
  de trabajo paralelas de otras herramientas: dos "tarea no encontrada" sobre tareas que existían en
  otra rama, tres volcados de pila al leer de una rama remota, y un identificador ambiguo. La [decisión
  de persistencia](persistencia.md#la-decisión-de-persistencia) no los resuelve, los disuelve: el tablero no vive en el árbol de
  trabajo, así que una tarea cerrada está cerrada y no hay una rama de la que leerla ni una remota que
  le falte. Lo que queda, y sigue siendo independiente del almacenamiento, son los tres mensajes
  distintos de ["no la encuentro"](../spec/referencias.md#los-tres-mensajes-de-no-la-encuentro) y la garantía de que ninguna lectura de
  conjunto aborta por una tarea que no se puede leer.
- **La sincronización con sistemas externos.** No está, pero sí están las cuatro piezas que la hacen
  posible, y esa es la única razón por la que existen: las claves declaradas de `ext` para guardar la
  identidad de la tarea en el otro sistema, el autor libre en los comentarios, las fechas fijables al
  importar y la simetría de `export` con `new --from`.
- **La clave de configuración `default_assignee`.** Habría asignado una persona a toda tarea creada
  sin `-a`. Se descarta porque en un tablero que la usara, absolutamente todo nacería asignado, y la
  asignación dejaría de significar que alguien decidió encargarte justo esa tarea: la consulta de
  arranque de un agente devolvería el backlog entero disfrazado de encargo. Es la comodidad concreta
  que habría destruido la señal en la que se apoya la decisión de ["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](modelo-de-estados.md#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo), que la asignación sea
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
