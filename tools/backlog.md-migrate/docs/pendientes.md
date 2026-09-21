# Decisiones pendientes de revisión

**Nada de este proyecto está dado por hecho.** Todo lo que dicen [`decisiones.md`](decisiones.md) y
[`especificacion.md`](especificacion.md) es una propuesta provisional, nacida de una conversación, y
falta repasar cada decisión una por una antes de escribir código. Esta página las junta todas para
poder hacer ese repaso, y dice qué las bloquea. No hay código Go todavía.

**Qué bloquea la implementación.** TASK-70 (`import`) y TASK-7 (`export`) están en el tablero en estado
`Blocked` y dependen de un cambio en `biso`, TASK-73, descrito en la sección C. Había otro, TASK-72
(aceptar espacios en una etiqueta), que se descartó: ver A11.

## A. Lo que se habló y se dio por bueno de palabra

Sigue pendiente de repaso: dar algo por bueno en una conversación no es cerrarlo.

| # | Decisión | Dónde está escrita |
|---|---|---|
| A1 | Herramienta general e independiente de `biso`, en `tools/backlog.md-migrate/`, un programa Go con módulo propio y documentación propia | "Es un proyecto independiente de `biso`" y `CLAUDE.md` |
| A2 | Nombre `backlog.md-migrate`, con dos órdenes, `import` y `export`, una dirección cada vez y sin sincronización | "Dos órdenes independientes" |
| A3 | El milestone se convierte en la etiqueta `milestone:<slug>` | "El milestone es una etiqueta" |
| A4 | El destino se crea antes con el prefijo que se quiera, los ids conservan su número, y solo los que chocan reciben un número nuevo, con sus menciones reescritas | "Los identificadores conservan su número" |
| A5 | La definición de hecho entra como criterio con el sufijo ` #dod`, y al exportar vuelve a su sección; sin el sufijo se queda como criterio | "La definición de hecho se marca con el sufijo" |
| A6 | Estados, tipos y prioridades son los del destino, casados con la regla de coincidencia de `biso` | "Estados, tipos y prioridades" |
| A7 | `documentation` y `references` van tal cual, sin comprobación de caracteres | "`documentation` y `references`" |
| A8 | Las fechas se leen como UTC y se les añade `:00` y `Z` | "Las fechas se leen como UTC" |
| A9 | La exportación es de mejor esfuerzo y reversible: lo que Backlog.md no tiene se escribe como una sección de Markdown delimitada, y se crean los ficheros de milestone que falten | "La exportación" |
| A10 | Un formato Markdown propio de `biso` es un proyecto aparte, no parte de esta herramienta | "La exportación" |
| A11 | Los espacios de una etiqueta se convierten en guiones normales al importar (`with space` pasa a `with-space`), en vez de cambiar `biso` para que los admita. TASK-72 se descartó por esto. La conversión no se deshace al exportar | "Los espacios de una etiqueta o un asignado se convierten en guiones" |

## B. Lo que propuse yo y no se ha aprobado

| # | Propuesta | Dónde está escrita |
|---|---|---|
| B1 | El campo `project` de Backlog.md se convierte en la etiqueta `project:<slug>`, igual que el milestone | Especificación, "Milestone y proyecto" |
| B2 | Si una tarea ya está en el destino (mismo título y misma fecha de creación, con cualquier id), no se importa otra vez, y se informa | Especificación, "Identificadores", regla 3 |
| B3 | Los ids de subtarea (`ABC-1.2`) reciben siempre un número nuevo. **Se cae si se hace TASK-73** | Especificación, "Identificadores", regla 4 |
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

## C. Lo que espera a un cambio en `biso`

Es un cambio en la especificación y el código de `biso`, y está en el tablero como tarea aparte.
Cuando se haga, hay que volver aquí y quitar las reglas que dejan de hacer falta.

- **TASK-73, aceptar identificadores de subtarea con punto.** Al hacerse, las subtareas conservan su id
  con el prefijo del destino y desaparece la regla B3. La regla de la tarea ya importada (B2) podría
  volver a comparar también el id.
- **Descartada, TASK-72 (aceptar espacios dentro de una etiqueta).** Se resolvió en el convertidor, con
  A11 y B18. **Pregunta que queda:** los demás caracteres que Backlog.md admite en una etiqueta y
  `biso` no (`/`, `!`) siguen quitando el valor con aviso (B7). ¿Se acepta eso, o se transforman
  también de algún modo?

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

## E. Cómo propongo repasarlo

En este orden, uno por uno: las decisiones estructurales (A1 y A2), las de mapeo (A3 a A8, A11 y B1 a B6),
las de lectura y salida (B7 a B18), y por último la exportación (A9 y A10). Antes de escribir código,
TASK-73 tiene que estar resuelta o descartada, y esta página tiene que quedar sin propuestas
pendientes.
