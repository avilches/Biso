# Lo que se deja fuera, y por qué

["Lo que se deja fuera a propósito"](../spec/fuera-de-alcance.md) nombra, sin justificarlo, lo que la
especificación decide no tener. Esta página es el porqué de seis de esas piezas, con la evidencia que
las descarta.

- **Los hitos como entidad.** En el estudio, el comando de crear hitos se usó 22 veces, pero los
  comandos de documentos y de decisiones no se usaron ni una sola vez en seis días. Por eso `biso`
  conservó el hito como campo de la tarea en vez de crear una entidad con ciclo de vida propio: sin
  comando que lo creara, sin clave de configuración que lo declarase, sin fecha ni estado propios.
  **Esta pieza ya no es cierta: `milestone` se retiró entero, no solo como entidad.**
  ["Se retiran `project` y `milestone`"](detalles.md#se-retiran-project-y-milestone) explica por qué
  el campo dejó de hacer falta y qué ocupa su lugar.
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
- **La clave de configuración `default_assignee`.** Habría asignado una persona a toda tarea creada
  sin `--add-assignees`. Se descarta porque en un tablero que la usara, absolutamente todo nacería asignado, y la
  asignación dejaría de significar que alguien decidió encargarte justo esa tarea: la consulta de
  arranque de un agente devolvería el backlog entero disfrazado de encargo. Es la comodidad concreta
  que habría destruido la señal en la que se apoya la decisión de ["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](modelo-de-estados.md#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo), que la asignación sea
  el gesto con el que una persona encarga trabajo.
