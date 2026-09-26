# Lo que se deja fuera a propósito

Nombrar lo que no está evita que alguien lo dé por olvidado.

- **No hay `biso delete`.** Está especificado que no existe y qué contesta si se intenta (["`biso archive`"](cmd/archive.md)).
- **No hay entidades de documento ni de decisión.** La documentación se apunta con `--doc`, que es
  una lista de textos, y una decisión de diseño no tiene comando propio: vive en la documentación del
  proyecto, no en el tablero.
- **No hay campo de hito ni de proyecto en la tarea.** Hubo un campo `milestone` y un campo `project`,
  y los dos se retiraron: la agrupación real se resuelve con `parent` (["Las relaciones entre tareas"](modelo-de-datos/relaciones.md):
  una tarea con hijas, de cualquier `type`) y con `type` (una partición plana), sin ningún campo dedicado solo a agrupar. El
  porqué, con las alternativas descartadas, está en
  ["Se retiran `project` y `milestone`"](../decisiones/detalles.md#se-retiran-project-y-milestone).
- **No hay definición de hecho.** Una tarea tiene una sola lista de comprobación, la de criterios de
  aceptación; lo que otra herramienta guardaría en una segunda lista es un criterio más. La
  importación de `biso new --from` sí acepta la clave `definitionOfDone` de un lote ajeno y convierte
  sus elementos en criterios, avisando de ello (["`biso new`"](cmd/new.md)). El porqué, con la medida
  que lo decide, está en
  ["Se retira la definición de hecho"](../decisiones/detalles.md#se-retira-la-definición-de-hecho).
- **No hay un campo de documentación aparte de `references`.** Una tarea tiene un solo campo de
  punteros, y un documento es una referencia más. La importación de `biso new --from` sí acepta la
  clave `documentation` de un lote ajeno y funde sus valores en `references`, avisando de ello
  (["`biso new`"](cmd/new.md)). El porqué, con la medida que lo decide, está en
  ["Se retira `documentation` y `references` queda como único campo de punteros"](../decisiones/detalles.md#se-retira-documentation-y-references-queda-como-único-campo-de-punteros).
- **No hay un campo de ficheros tocados.** Qué código tocó un trabajo lo dice el control de versiones,
  y una ruta que valga la pena señalar es una referencia más. La importación de `biso new --from` sí
  acepta la clave `modifiedFiles` de un lote ajeno y funde sus valores en `references`, avisando de
  ello (["`biso new`"](cmd/new.md)). El porqué, con las medidas que lo deciden, está en
  ["Se retira `modifiedFiles`"](../decisiones/detalles.md#se-retira-modifiedfiles).
- **No hay contextos de sesión**, es decir, filtros por defecto guardados que cambien lo que devuelve
  una consulta sin que se vea en la línea de comandos.
- **No hay recurrencia, ni seguimiento de tiempo, ni subtareas con numeración propia.** Una subtarea
  es una tarea normal con `--parent`, y el mensaje de error de un identificador como `MYP-1.1` lo
  dice. El porqué de no adoptar esa forma con punto, con la alternativa que resuelve sin ella la
  migración desde Backlog.md, está en
  ["No se adoptan identificadores de subtarea con punto"](../decisiones/detalles.md#no-se-adoptan-identificadores-de-subtarea-con-punto).
- **No hay servidor de integración ni protocolo de herramientas.** La interfaz de la versión 1.0 es
  esta línea de comandos y su salida JSON.
- **No hay ningún flag ni variable de entorno que nombre un tablero.** El tablero se elige por
  las dos vías de la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md), y `-C` ya alcanza tanto el directorio de un tablero como el de un
  proyecto que apunte a uno, así que un flag para nombrarlo no añadiría nada (sección ["El presupuesto del mensaje de arranque"](../decisiones/vocabulario-y-mensaje-de-arranque.md#el-presupuesto-del-mensaje-de-arranque)).
- **No hay una interfaz multiproyecto.** Cada invocación resuelve un único tablero (sección ["Cómo se elige el tablero"](resolucion-del-tablero.md)), y no
  hay ningún comando que lea o agregue varios a la vez, aunque la máquina entera tenga más de uno
  (sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina)): quien necesite verlos juntos los recorre uno por uno desde fuera.
- **No hay exportación al formato de Backlog.md.** `biso export` escribe el mismo formato que lee
  `biso new --from`, y traducir a un formato ajeno es trabajo de un conversor aparte, no de este
  comando.
- **No hay sincronización entre máquinas.** Un tablero vive en la máquina donde se creó, y lo que
  cruza a otra es la instantánea que deja `biso snapshot`, para reconstruirlo entero con
  `biso init --from`, no para mantener dos copias vivas al día (sección ["La decisión de persistencia"](../decisiones/persistencia.md#la-decisión-de-persistencia)).
  **Cómo cruza depende de dónde viva el tablero, y las dos vías están especificadas** (["`biso snapshot`"](cmd/snapshot.md)): un
  tablero versionado dentro del proyecto viaja con él y con el remoto que el proyecto ya tenga, sin que
  nadie configure nada, y un tablero con su propio repositorio viaja cuando alguien le da un remoto y
  `biso snapshot --vcs push` lo publica. Lo que no hay es ningún remoto que `biso` configure por su
  cuenta, ni forma de fusionar dos almacenes escritos por separado.
- **No hay un papel de estado para descartar, distinto de terminar.** Una tarea hecha y una abandonada
  hoy comparten el mismo estado terminal. Un papel que obligara a dar un motivo al entrar en él sería
  barato de añadir cuando hiciera falta, pero a diferencia de los demás requisitos de este modelo no
  trae ningún caso real en el que la confusión haya costado algo, así que se queda fuera hasta que
  aparezca uno.

---

