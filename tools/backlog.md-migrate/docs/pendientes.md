# Decisiones pendientes de revisión

**El repaso de las secciones A y B ya se hizo, decisión por decisión, con quien encarga el proyecto.**
Todo lo que dicen [`decisiones.md`](decisiones.md) y [`especificacion.md`](especificacion.md) nació de
una conversación, y esta página se escribió para juntar cada propuesta y repasarla antes de escribir
código. Ese repaso ya terminó: las secciones A y B quedan ratificadas tal como están escritas, y las dos
preguntas que quedaban abiertas en la sección C también están cerradas. No hay código Go todavía, pero
ya no es por falta de decisiones: lo que sigue pendiente es otro tipo de trabajo, no una decisión, y está
en la sección D (mediciones sin hacer) y en el diseño completo de la exportación (TASK-7).

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

- **Otras versiones de Backlog.md.** Todo se midió con la 1.52.0. El formato de las fechas, los ids y
  los comentarios puede ser distinto en otras.
- **La configuración fuera de la carpeta de datos** (`backlog.config.yml`). Se comprobó que existe y
  dónde, pero como la herramienta no la lee, no se ha medido nada más de ella.
- **Fichero con saltos de línea de Windows, con marca de orden de bytes o con el frontmatter mal
  formado.** No se ha probado cómo los lee Backlog.md ni qué debe hacer la herramienta.
- **Tareas de otras ramas.** Backlog.md puede mostrar tareas que viven en otras ramas de git
  (`check_active_branches`). La herramienta lee solo el directorio que se le da.
- **Un comentario sin `created:`.** No se ha visto que Backlog.md lo produzca.
- **La exportación entera.** No hay nada medido: el formato de la sección de Markdown reversible, si
  Backlog.md la conserva cuando edita una tarea, el formato de los ficheros de milestone que se crean
  y la forma de los ids al escribir. Todo eso es el diseño de TASK-7.
- **La estabilidad de las órdenes de `biso` que se usan.** No he encontrado una garantía escrita para
  `biso config list --json` y `biso export`.
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

Lo único que sigue pendiente es otro tipo de trabajo, no una decisión: las mediciones sin hacer de la
sección D, y el diseño completo de la exportación (A9, A10, TASK-7), que es tarea aparte.
