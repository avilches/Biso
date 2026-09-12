# Lo que se deja fuera a propósito

Nombrar lo que no está evita que alguien lo dé por olvidado.

- **No hay `biso delete`.** Está especificado que no existe y qué contesta si se intenta (["`biso archive`"](cmd/archive.md)).
- **No hay entidades de hito, documento ni decisión.** El hito es un campo de la tarea y no una
  entidad con ciclo de vida propio: no se crea, no se cierra, no tiene fecha ni descripción, y no hay
  ninguna clave de configuración que lo declare. Eso no impide que `-m/--milestone` valide al filtrar,
  porque el conjunto contra el que valida es derivado de lo que las tareas usan (["Qué valida cada filtro, y contra qué"](vocabularios.md#qué-valida-cada-filtro-y-contra-qué)) y no una lista
  que haya que mantener aparte. La documentación se apunta con `--doc`, que es una lista de textos.
- **No hay contextos de sesión**, es decir, filtros por defecto guardados que cambien lo que devuelve
  una consulta sin que se vea en la línea de comandos.
- **No hay recurrencia, ni seguimiento de tiempo, ni subtareas con numeración propia.** Una subtarea
  es una tarea normal con `--parent`, y el mensaje de error de un identificador como `MYP-1.1` lo
  dice.
- **No hay servidor de integración ni protocolo de herramientas.** La interfaz de la versión 1.0 es
  esta línea de comandos y su salida JSON.
- **No hay ninguna bandera ni variable de entorno que nombre un tablero.** El tablero se elige por
  las dos vías de la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md), y `-C` ya alcanza tanto el directorio de un tablero como el de un
  proyecto que apunte a uno, así que una bandera para nombrarlo no añadiría nada (sección ["El presupuesto del mensaje de arranque"](../DECISIONES.md#el-presupuesto-del-mensaje-de-arranque)
  de `docs/DECISIONES.md`).
- **No hay una interfaz multiproyecto.** Cada invocación resuelve un único tablero (sección ["Cómo se elige el tablero"](resolucion-del-tablero.md)), y no
  hay ningún comando que lea o agregue varios a la vez, aunque la máquina entera tenga más de uno
  (sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina)): quien necesite verlos juntos los recorre uno por uno desde fuera.
- **No hay exportación al formato de Backlog.md.** `biso export` escribe el mismo formato que lee
  `biso new --from`, y traducir a un formato ajeno es trabajo de un conversor aparte, no de este
  comando.
- **No hay sincronización con ningún sistema externo.**
- **No hay sincronización entre máquinas.** Un tablero vive en la máquina donde se creó, y lo que
  cruza a otra es la instantánea que deja `biso snapshot`, para reconstruirlo entero con
  `biso init --from`, no para mantener dos copias vivas al día (sección ["La decisión de persistencia"](../DECISIONES.md#la-decisión-de-persistencia) de `docs/DECISIONES.md`).
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

