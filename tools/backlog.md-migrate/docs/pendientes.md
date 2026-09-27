# Decisiones pendientes de revisión

**El repaso de las secciones A y B ya se hizo, decisión por decisión, con quien encarga el proyecto.**
Todo lo que dicen [`decisiones.md`](decisiones.md) y [`especificacion.md`](especificacion.md) nació de
una conversación, y esta página se escribió para juntar cada propuesta y repasarla antes de escribir
código. Ese repaso ya terminó: las secciones A y B quedan ratificadas tal como están escritas, y las dos
preguntas que quedaban abiertas en la sección C también están cerradas. La implementación de `import`
ya está terminada; lo que sigue pendiente ya no es ninguna decisión: es otro tipo de trabajo, y está en
la sección D (mediciones sin hacer) y en el diseño completo de la exportación (TASK-7).

**Qué bloqueaba la implementación, y ya no.** TASK-70 (`import`) y TASK-7 (`export`) esperaban a que
`biso` resolviera TASK-73 (aceptar identificadores de subtarea con punto). TASK-73 ya se resolvió, y en
el sentido contrario al que pedía la tarea original: `biso` no adopta la forma con punto, y dejó en
manos de este proyecto decidir cómo conservar el id de origen de una subtarea. Esa decisión ya está
tomada, en ["El identificador de origen de una subtarea se guarda en una etiqueta con
ámbito"](decisiones.md#el-identificador-de-origen-de-una-subtarea-se-guarda-en-una-etiqueta-con-ámbito)
de `decisiones.md`, y la sección C de esta página recoge lo que queda de aquello, ya cerrado. Había otro
cambio pedido a `biso`, TASK-72 (aceptar espacios en una etiqueta), que se descartó: ver A11. El repaso
de las secciones A y B, que era lo único que seguía bloqueando el paso a código, también terminó y no
dejó ninguna decisión pendiente. Lo que queda en esta página, la sección D y la exportación (TASK-7), es
trabajo de medición y de diseño, no propuestas a la espera de repaso.

## A. Lo que se habló y se dio por bueno de palabra

El repaso terminó: todas las entradas quedan ratificadas tal como están escritas, sin ninguna
bifurcación pendiente. La única precisión añadida es en A11, donde queda confirmado que la conversión de
espacios en guiones se aplica igual a los asignados que a las etiquetas (ver decisiones.md).

| # | Decisión | Dónde está escrita |
|---|---|---|
| A1 | Herramienta general e independiente de `biso`, en `tools/backlog.md-migrate/`, un programa Go con módulo propio y documentación propia | "Es un proyecto independiente de `biso`" y `CLAUDE.md` |
| A2 | Nombre `backlog.md-migrate`, con dos órdenes, `import` y `export`, una dirección cada vez y sin sincronización | "Dos órdenes independientes" |
| A3 | El milestone se convierte en la etiqueta con ámbito `milestone::<slug>` (doble `::`, porque una tarea tiene como mucho un milestone; TASK-85 hizo posible esta forma) | "El milestone es una etiqueta con ámbito" |
| A4 | El destino se crea antes con el prefijo que se quiera, los ids conservan su número, y solo los que chocan reciben un número nuevo, con sus menciones reescritas | "Los identificadores conservan su número" |
| A12 | El id de origen de una subtarea (`TASK-56.1`) se guarda en la etiqueta con ámbito `backlog.id::TASK-56.1`, para que `export` lo pueda reconstruir | "El identificador de origen de una subtarea se guarda en una etiqueta con ámbito" |
| A13 | El `ordinal` de Backlog.md no se copia: se recalcula con el algoritmo del punto medio de `biso`, ordenando las tareas de origen por su `ordinal` y asignando claves nuevas que conservan el orden relativo | "El orden manual se recalcula, no se copia" |
| A14 | `documentation` y `modified_files` de Backlog.md se escriben tal cual en las claves de compatibilidad `documentation` y `modifiedFiles` del lote, y es `biso new --from` quien las funde en `references` | "`documentation`, `references` y `modified_files` van tal cual, y `documentation`/`modified_files` se dejan fundir en `references`" |
| A5 | La definición de hecho entra como criterio con el sufijo ` #dod`, y al exportar vuelve a su sección; sin el sufijo se queda como criterio | "La definición de hecho se marca con el sufijo" |
| A6 | Estados, tipos y prioridades son los del destino, casados con la regla de coincidencia de `biso` | "Estados, tipos y prioridades" |
| A7 | `references` va tal cual, sin comprobación de caracteres; `biso` ya no tiene campo `documentation`, ver A14 | "`documentation`, `references` y `modified_files` van tal cual, y `documentation`/`modified_files` se dejan fundir en `references`" |
| A8 | Las fechas se leen como UTC y se les añade `:00` y `Z` | "Las fechas se leen como UTC" |
| A9 | La exportación es de mejor esfuerzo y reversible: lo que Backlog.md no tiene se escribe como una sección de Markdown delimitada, y se crean los ficheros de milestone que falten | "La exportación" |
| A10 | Un formato Markdown propio de `biso` es un proyecto aparte, no parte de esta herramienta | "La exportación" |
| A11 | Los espacios de una etiqueta se convierten en guiones normales al importar (`with space` pasa a `with-space`), en vez de cambiar `biso` para que los admita. TASK-72 se descartó por esto. La conversión no se deshace al exportar | "Los espacios de una etiqueta o un asignado se convierten en guiones" |

## B. Lo que propuse yo y no se ha aprobado

El repaso terminó: todas las propuestas quedan ratificadas tal como están escritas, porque ninguna tenía
una alternativa real documentada.

| # | Propuesta | Dónde está escrita |
|---|---|---|
| B1 | El campo `project` de Backlog.md se convierte en la etiqueta con ámbito `project::<slug>`, igual que el milestone | Especificación, "Milestone y proyecto" |
| B2 | Si una tarea ya está en el destino (mismo título y misma fecha de creación, con cualquier id), no se importa otra vez, y se informa | Especificación, "Identificadores", regla 3 |
| B3 | Los ids de subtarea (`ABC-1.2`) reciben siempre un número nuevo. **TASK-73 se resolvió confirmando esta regla: no hay ninguna forma con punto que adoptar, y el id de origen se conserva aparte, con A12** | Especificación, "Identificadores", regla 4 |
| B4 | Un criterio de aceptación de varias líneas se une con espacios y se informa, porque el texto de un criterio de `biso` no admite saltos de línea | Especificación, "Criterios de aceptación" |
| B5 | Un tablero configurado con elementos de definición de hecho por defecto llena cada tarea de criterios `#dod`; no hay opción para omitirlos | Especificación, "Definición de hecho" |
| B6 | Un estado, tipo o prioridad que el destino no declara se omite de la línea y se informa; sin estado, `biso` pone el inicial | Especificación, "Vocabularios del destino" |
| B7 | Una etiqueta o un asignado que sigue sin encajar en el alfabeto de un token tras convertir los espacios (`a/b`, `c!`) se quita y se informa | Especificación, "Alfabeto de un token" |
| B18 | La conversión de espacios en guiones (A11) se aplica también a los asignados (`Sara Smith` pasa a `Sara-Smith`), se quita el espacio de los bordes, cada tramo de espacios da un solo guion, y cada conversión se informa | Especificación, "Alfabeto de un token" |
| B8 | Un `parent` o una dependencia que nombra una tarea inexistente se quita y se informa; los ciclos los detecta `biso new --from --dry-run` | Especificación, "Identificadores", regla 7 |
| B9 | La reescritura de menciones distingue mayúsculas, respeta los límites de palabra, incluye el título y los bloques de código, y avisa de las casi menciones (`Xyz-002`, `task-12`) | Especificación, "Identificadores", reglas 5 y 6 |
| B10 | Un comentario sin `author:` se importa sin autor y sin hallazgo | Especificación, "Comentarios" |
| B11 | El slug del milestone quita diacríticos y separa por guiones; si sale vacío o dos coinciden, se usa el id y se informa | Especificación, "Milestone y proyecto" |
| B12 | Sin flag de zona horaria, y sin `updated_date` se usa la de creación | Especificación, "Fechas" |
| B13 | Firma: `import <backlog-dir> --project <dir> [--out] [--biso] [--strict]`; códigos de salida 0, 2, 3, 4 y 5; la salida es NDJSON ordenado y determinista; los hallazgos van por la salida de errores, una línea cada uno | Especificación, "Firma", "Los hallazgos" y "Códigos de salida" |
| B14 | Lee `tasks/`, `completed/` y `archive/tasks/` (archivadas con `archived: true`) y los milestones; no lee borradores, documentos ni decisiones de Backlog.md, solo avisa de cuántos hay | Especificación, "Qué lee del origen" |
| B15 | La herramienta no lee la configuración de Backlog.md: deduce el prefijo de los ids | Especificación, "Qué lee del origen" y "Identificadores" |
| B16 | Le pide al destino su configuración y sus tareas ejecutando `biso ... --json` y `biso export`, sin importar código de `biso` | Especificación, "Qué le pregunta al destino" |
| B17 | `--strict` se niega a escribir nada si hay algún hallazgo, y el código 5 no es un fallo | Especificación, "Códigos de salida" |

## C. Lo que pedía un cambio en `biso`, y cómo se resolvió

Los dos puntos de esta sección pedían un cambio en la especificación y el código de `biso`. Los dos
están resueltos, y ninguno de los dos con un cambio en `biso`. Las dos preguntas que quedaban abiertas
también se han cerrado.

- **TASK-73, aceptar identificadores de subtarea con punto: resuelta, en el sentido contrario al que
  pedía la tarea.** `biso` decidió no adoptar la forma `<PREFIJO>-<n>.<m>`
  (["No se adoptan identificadores de subtarea con punto"](../../../docs/decisiones/detalles.md#no-se-adoptan-identificadores-de-subtarea-con-punto)
  en la raíz del repositorio), así que la regla B3 **se confirma y no desaparece**: una subtarea siempre
  recibe un número nuevo, con el prefijo del destino, igual que cualquier id que choque. Lo que sí es
  nuevo es que este proyecto ya tiene su propio mecanismo para no perder el id de origen: la etiqueta con
  ámbito `backlog.id::<id>` de A12, documentada en decisiones.md
  (["El identificador de origen de una subtarea se guarda en una etiqueta con
  ámbito"](decisiones.md#el-identificador-de-origen-de-una-subtarea-se-guarda-en-una-etiqueta-con-ámbito)).
  **Pregunta cerrada.** Se planteó si la regla de la tarea ya importada (B2) debería usar también
  `backlog.id::` para reconocer una subtarea repetida, además de título y fecha de creación. La
  respuesta es no: B2 se deja sin cambios, solo título y fecha de creación, porque ya cubre el caso real
  y añadir una segunda vía de comparación sería complejidad extra para un caso que no se ha visto
  necesario.
- **Descartada, TASK-72 (aceptar espacios dentro de una etiqueta).** Se resolvió en el convertidor, con
  A11 y B18. **Pregunta cerrada.** Se planteó si los demás caracteres que Backlog.md admite en una
  etiqueta y `biso` no (`/`, `!`) deberían transformarse de algún modo en vez de quitarse con aviso (B7).
  La respuesta es no: se mantiene B7 tal cual. Ampliar el alfabeto de `labels`/`assignees` de `biso` para
  admitir `/` o `!` repetiría exactamente el mismo caso que ya se descartó para los espacios en TASK-72,
  cambiar una decisión central del modelo de `biso`, que rige en todo el programa y no solo en la
  importación, para servir a una utilidad de migración. Si en algún momento se quiere reconsiderar el
  alfabeto de `biso` en general, sería una tarea aparte en el tablero de `biso`, no una decisión de este
  convertidor.

## D. Lo que sigue sin medir

- **Otras versiones de Backlog.md: comprobado que 1.53.0 no rompe nada de lo medido con 1.52.0.**
  Se repitió con la 1.53.0 (`backlog --version` lo confirma) la comprobación de cuatro de los hechos
  que esta página y `decisiones.md` daban por medidos con la 1.52.0, creando un tablero de prueba con
  `backlog init "Medicion153" --defaults`: el formato de fecha `YYYY-MM-DD HH:mm` en UTC (una tarea
  creada a las `2026-09-26 17:02 -0400` según `date` quedó con `created_date: '2026-09-26 21:02'`, la
  hora UTC), la forma con punto del id de una subtarea (`backlog task create "..." --parent FEAT-1`
  creó `FEAT-1.1`, con `parent_task_id: FEAT-1` en su frontmatter), la ubicación de
  `backlog.config.yml` fuera de la carpeta de datos con `--config-location root` (queda junto a
  `AGENTS.md`, no dentro de `backlog/`) y la capitalización del prefijo, que no coincide entre la
  configuración y los ids: con `task_prefix: "task"` en minúsculas la tarea creada fue `TASK-1`, y con
  `task_prefix: "Feat"` fue `FEAT-1`, siempre en mayúsculas pase lo que pase en la configuración. Los
  cuatro se comportan igual que con la 1.52.0. No hace falta instalar ni probar ninguna otra versión
  para esta revisión; queda abierto, como ya decía esta página, que una versión futura cambie algo que
  no se haya vuelto a comprobar.
- **La configuración fuera de la carpeta de datos (`backlog.config.yml`): confirmado que no hace
  falta leerla.** Se probó, sobre el tablero de prueba de la 1.53.0, si `statuses`, `types`,
  `priorities` o el prefijo declarados en la configuración esconden algún dato que el convertidor
  necesitaría y que no esté también en las propias tareas. No lo esconden: cada tarea escribe el valor
  elegido tal cual en su propio frontmatter, exista o no ese valor en la lista de la configuración. Con
  `backlog task create "Tarea con tipo" --type spike --priority High` el fichero queda con
  `priority: high` y `type: spike` en minúsculas, aunque `backlog config list` muestre
  `priorities: [High, Medium, Low]` con mayúscula; y `type` ni siquiera aparecía como clave en
  `config.yml` (solo en la salida de `config list`, que trae sus valores por defecto sin que estén
  escritos en el fichero). Lo mismo pasa con `project`: `backlog config get projects` obliga a declarar
  antes una lista en el fichero de configuración para poder usar `--project`, pero una vez creada la
  tarea con `--project ProyectoX` el valor `ProyectoX` queda escrito tal cual en `project:` de la
  tarea, legible sin volver a mirar la configuración. El `milestone` de una tarea guarda el id del
  milestone (`milestone: m-0`), no su título; el título vive en el frontmatter del propio fichero de
  `backlog/milestones/m-0 - hito-uno.md` (`title: "Hito uno"`), que la herramienta ya lee según B14, no
  en `backlog.config.yml`. Como A6 dice que el vocabulario que manda es el del destino, el vocabulario
  que el origen declaró pero no llegó a usar en ninguna tarea nunca hace falta conocerlo. B15 (la
  herramienta no lee `backlog.config.yml`) queda confirmada, y de hecho leerla habría sido
  contraproducente para el prefijo, por la capitalización distinta del punto anterior.
- **Fichero con CRLF, con marca de orden de bytes o con el frontmatter mal formado: medido.** Sobre una
  tarea creada con el CLI (`backlog task create "Tarea normal" --status "In Progress" -l alpha,beta`)
  se probaron tres ediciones a mano. Saltos de línea CRLF: `backlog task list` y `backlog task view` la
  leen sin problema, y al editarla con `backlog task edit TASK-1 -d "..."` Backlog.md reescribe el
  fichero entero con saltos LF, normalizándolo sin avisar. Marca de orden de bytes UTF-8 al principio
  del fichero: se lee igual de bien, sin ningún aviso ni diferencia en la salida. Un frontmatter YAML
  roto a propósito de dos formas (una comilla sin cerrar en `title: "Tarea rota`, y una clave `status`
  duplicada) hace que Backlog.md 1.53.0 **descarte la tarea en silencio**: no sale en
  `backlog task list`, ni en `stdout` ni en `stderr` hay ningún aviso de que el fichero existe y no se
  pudo leer, y `backlog task view TASK-3` responde `Task TASK-3 not found`, el mismo mensaje que si el
  fichero no existiera. Esto no era un problema mientras el convertidor leyera a través del CLI de
  Backlog.md, pero lee `tasks/`, `completed/` y `archive/tasks/` **directamente**, con su propio
  analizador de YAML, así que sí se va a encontrar un fichero así en un tablero real. **Decidido:** ese
  fichero se salta, se informa como hallazgo con el fichero y el error de parseo, y el resto del lote
  sigue adelante; con `--strict`, cuenta igual que cualquier otro hallazgo y no se escribe nada. La
  decisión está en ["Nada se pierde en
  silencio"](decisiones.md#nada-se-pierde-en-silencio) de `decisiones.md`, y la regla en
  "Qué lee del origen", "Los hallazgos" y "Códigos de salida" de `especificacion.md`.
- **Tareas de otras ramas: medido, y confirma la suposición de la especificación con un matiz.** Con un
  tablero de prueba, una tarea creada y comiteada en `main` y otra tarea creada y comiteada en una rama
  `feature-y` sin mezclar, al volver a `main`: `backlog task list` solo muestra la tarea de `main`
  (`TASK-1`), nunca la de `feature-y`. En cambio `backlog board` sí cruza ramas: imprime
  `Indexing 1 other local branches...` y el tablero resultante incluye las dos tareas, `TASK-1` y
  `TASK-2`. Repetido con `backlog config set checkActiveBranches false`, `board` deja de cruzar ramas y
  vuelve a mostrar solo `TASK-1`. Es decir, `check_active_branches` sí existe y sí mezcla ramas, pero
  solo afecta a comandos como `board` (probablemente también a la interfaz web), no a `task list`. La
  especificación de este convertidor asume leer solo el directorio de trabajo actual, que es
  exactamente lo que hace `task list` y no lo que hace `board`; la suposición es correcta, y conviene
  seguir describiéndola como "igual que `task list`", no como "igual que Backlog.md en general".
- **Un comentario sin `created:`: confirmado que el CLI nunca lo omite, y que si falta por edición a
  mano Backlog.md lo tolera en vez de rechazarlo.** `backlog task edit TASK-1 --comment "..."` siempre
  escribe `created: <fecha>` como primera línea del comentario, con o sin `--comment-author`. No hay
  ninguna combinación de flags que lo omita: solo llega sin `created:` editando el fichero a mano, como
  ya decía esta página. Se probó también qué pasa si se hace: quitando a mano la línea `created:` de un
  comentario ya escrito, `backlog task view` lo sigue mostrando (como comentario `#1`, sin fecha) en vez
  de descartar la tarea entera, a diferencia del frontmatter roto de más arriba. Es un dato a favor de
  B10 (un comentario sin autor se importa sin aviso): un comentario sin `created:` puede llegar por la
  misma vía, edición manual, y Backlog.md no lo rechaza, así que el convertidor debería poder leerlo sin
  fallar aunque solo el CLI nunca lo produzca.
- **La exportación entera.** No hay nada medido: el formato de la sección de Markdown reversible, si
  Backlog.md la conserva cuando edita una tarea, el formato de los ficheros de milestone que se crean
  y la forma de los ids al escribir. Todo eso es el diseño de TASK-7.
- **La estabilidad de las órdenes de `biso` que se usan: es una lectura de la especificación de `biso`,
  no una medición, y sí hay garantía escrita para las dos.** `docs/spec/estabilidad.md` de la raíz del
  repositorio no nombra `biso config list --json` ni `biso export` una por una, pero los cubre con dos
  reglas generales. Para `config list --json`: "Las claves de `data` en cada `kind` de JSON... se
  pueden añadir claves; las que hay no se quitan ni cambian de tipo", y `docs/spec/cmd/config.md` fija
  que su `kind: "config"` lleva `data.config` con, entre otras, `statuses`, `types`, `priorities` y
  `task_prefix`, justo las claves que este convertidor necesita. Para `export`: "La simetría entre
  `biso export` y `biso new --from` sobre todos los campos no derivados, que es una prueba de la suite y
  no una intención", que es una garantía explícita sobre el formato de salida de `export`
  (`docs/spec/cmd/export.md` la describe como NDJSON con las claves que acepta `new --from`). Lo que sí
  puede cambiar entre versiones menores, según el mismo documento, es el texto de los mensajes de error
  y avisos y la disposición en columnas de la salida en texto de `biso ls`/`biso get` (para eso está
  `--json`), pero eso no afecta a este convertidor porque ya usa `--json` y `export` en vez de la salida
  de texto. Corrijo así lo que decía esta página: no es que falte la garantía, es que está repartida en
  dos reglas generales del contrato en vez de una promesa por comando.
- **Cómo es de verdad el `ordinal` de Backlog.md: medido, pero no con el CLI.** A diferencia del resto
  de esta página, esto se midió contra los ficheros reales de `backlog/tasks/*.md` del tablero de este
  mismo proyecto (111 tareas), no creando un tablero de prueba con el CLI de Backlog.md como pide "El
  formato de Backlog.md se mide con su CLI, no con un tablero" de decisiones.md. Es una muestra de un
  solo tablero, con un patrón claro y consistente, pero no tiene la misma garantía metodológica que las
  mediciones hechas con el CLI. El resultado: el `ordinal` es un único contador global, no uno por
  columna de estado (`TASK-69`, estado Done, `ordinal: 200`; `TASK-70`, To Do, `ordinal: 300`; `TASK-68`,
  To Do, `ordinal: 400`; `TASK-71`, Done, `ordinal: 600`, entremezclados sin importar el estado). No es
  único: `TASK-3` y `TASK-8` comparten `ordinal: 3000`, el único par que lo hace en las 111 tareas del
  tablero, todas con el campo presente. Backlog.md inserta con valores intermedios al reordenar a mano,
  sin renumerar el resto (`ordinal` en incrementos de 500, como `26500`, `26750` y `41500`, intercalados
  entre los de incrementos de 1000). La regla de desempate para el caso de empate está en "El orden
  manual se recalcula, no se copia" de decisiones.md. Sigue siendo deseable, aunque no bloqueante,
  repetir esta comprobación creando un tablero de prueba con el CLI de Backlog.md y forzando a propósito
  un empate de `ordinal` entre tareas de distinto estado, para tener la misma garantía metodológica que
  el resto de lo medido.

## E. El repaso terminó

El repaso se hizo en el orden previsto: las decisiones estructurales (A1 y A2), las de mapeo (A3 a A8,
A11 a A14 y B1 a B6) y las de lectura y salida (B7 a B18). Las tres capas quedan ratificadas, con la
única precisión de A11 sobre los asignados (ver decisiones.md) y con las dos preguntas de la sección C
ya cerradas. Esta página no tiene ya ninguna propuesta pendiente de decisión.

Lo único que sigue pendiente es otro tipo de trabajo, no una decisión: lo que queda sin medir de la
sección D (repetir la comprobación del `ordinal` con el CLI, y el diseño completo de la exportación,
A9, A10, TASK-7, que es tarea aparte). La decisión nueva que salió de medir la sección D, qué hace
`import` con un fichero de `tasks/` cuyo frontmatter YAML no se puede analizar, ya está cerrada (ver
sección D).
