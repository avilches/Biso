# Especificación

> **Borrador provisional.** Ninguna regla de esta página está aprobada: son propuestas pendientes de
> revisión, y varias dependen de un cambio en `biso` (TASK-73). Ver
> [`pendientes.md`](pendientes.md).

Esta página dice qué hace cada orden, sin decir por qué: las razones están en
[`decisiones.md`](decisiones.md). Hoy solo `import` está especificada. `export` se diseña en TASK-7
y sus reglas ya decididas están en la sección "La exportación" de las decisiones.

Versión de Backlog.md contra la que se ha medido todo lo que sigue: **1.52.0**.

## `backlog.md-migrate import`

### Firma

```
backlog.md-migrate import <backlog-dir> --project <dir> [--out <file|->] [--biso <path>] [--strict]
```

| Parámetro | Oblig. | Por defecto | Qué es |
|---|---|---|---|
| `<backlog-dir>` | sí | | La carpeta de datos de un proyecto de Backlog.md, la que contiene `tasks/`. Se llama `backlog` por defecto, pero Backlog.md admite `.backlog` o cualquier otra ruta. Puede ser un enlace simbólico |
| `--project <dir>` | sí | | Directorio desde el que `biso` resuelve el tablero destino. Se le pasa a `biso` como `--cwd` |
| `--out <file\|->` | no | `-` (salida estándar) | Dónde se escribe el NDJSON |
| `--biso <path>` | no | `biso` del `PATH` | El binario de `biso` que se ejecuta |
| `--strict` | no | falso | Si hay algún hallazgo, no escribe nada y sale con el código 5 |

### Qué lee del origen

- `tasks/` y `completed/`: las tareas, en el mismo formato. Cada fichero `.md` es una tarea.
- `archive/tasks/`: las tareas archivadas. Se importan con `archived: true`.
- `milestones/` y `archive/milestones/`: solo para leer el título de cada milestone (el campo `title`
  de su frontmatter).
- **No lee la configuración de Backlog.md.** Backlog.md la guarda en `<backlog-dir>/config.yml` o, con
  `--config-location root`, en un `backlog.config.yml` junto a la carpeta, y nada de lo que contiene
  hace falta: el prefijo de los ids del origen se deduce de los propios ids (ver "Identificadores").
- `drafts/`, `archive/drafts/`, `docs/` y `decisions/` **no se leen**: los borradores no son tareas
  (tienen un prefijo propio, `DRAFT`, y el estado `Draft`), y los documentos y las decisiones de
  Backlog.md no tienen equivalente en `biso`. Si alguna de esas carpetas tiene ficheros, sale un
  hallazgo con cuántos.

### Qué le pregunta al destino

Ejecuta `biso` con `--cwd <project>` y `--json`, para no depender del formato de texto, y pide solo:

- la configuración (`biso config list --json`): el prefijo de los ids (`task_prefix`) y los estados,
  tipos y prioridades declarados;
- las tareas que ya existen, archivadas y terminales incluidas (`biso export --out -`): de cada una
  se usan su `id`, su `title` y su `createdAt`.

No escribe nada en el tablero destino: la escritura la hace después `biso new --from`. Lo que lee del
destino es una fotografía del momento de la conversión. Si otra persona o agente añade tareas al
destino entre la conversión y la importación, `biso` rechaza el lote entero por id ocupado y no escribe
nada, así que el fallo es seguro y se resuelve repitiendo la conversión.

### La salida

NDJSON: un objeto JSON por línea, con las claves que acepta `biso new --from`, en `camelCase`. Las
tareas salen ordenadas por el número de su id de origen, ascendente, y las claves de cada línea en un
orden fijo, para que dos ejecuciones sobre el mismo tablero den el mismo fichero byte a byte.

### El mapeo de campos

| Backlog.md | biso (`new --from`) | Regla |
|---|---|---|
| `id` | `id` | Prefijo del destino y número del origen, salvo colisión. Ver "Identificadores" |
| `title` | `title` | Con las menciones de ids reescritas |
| `status` | `status` | Se casa con el vocabulario del destino. Ver "Vocabularios del destino" |
| `type`, `priority` | `type`, `priority` | Igual que `status` |
| `assignee` (lista) | `assignees` | Cada valor se valida contra el alfabeto de un token |
| `labels` (lista) | `labels` | Igual, más las etiquetas del milestone y del proyecto |
| `milestone` | una etiqueta en `labels` | `milestone:<slug>`. Ver "Milestone" |
| `project` | una etiqueta en `labels` | `project:<slug>`, con el mismo slug que el milestone |
| `dependencies` | `dependencies` | Con los ids reescritos |
| `parent_task_id` | `parent` | Con el id reescrito |
| `ordinal` | `ordinal` | Tal cual |
| `due_date` | `due` | `YYYY-MM-DD`, tal cual |
| `documentation` | `documentation` | Lista tal cual |
| `references` | `references` | Lista tal cual |
| `modified_files` | `modifiedFiles` | Lista tal cual |
| `created_date` | `createdAt` | Ver "Fechas" |
| `updated_date` | `updatedAt` | Ver "Fechas" |
| sección `DESCRIPTION` | `description` | El texto entre las marcas `<!-- SECTION:DESCRIPTION:BEGIN -->` y `END`, sin el salto de línea inicial y final. Los `##` de dentro del texto no se interpretan como secciones |
| sección `PLAN` | `plan` | Ídem |
| sección `NOTES` | `notes` | Ídem |
| sección `FINAL_SUMMARY` | `summary` | Ídem |
| criterios de aceptación | `acceptanceCriteria` | Casillas `- [x] #n texto`. La clave es `n` y `checked` sale de la casilla. Ver "Criterios de aceptación" |
| definición de hecho | `acceptanceCriteria` | Con el sufijo ` #dod`. Ver "Definición de hecho" |
| sección `COMMENTS` | `comments` | Un comentario por bloque, con su autor, su fecha y su cuerpo. Ver "Comentarios" |

**Cualquier clave del frontmatter o sección del cuerpo que no aparezca en esta tabla es un
hallazgo**, con el nombre del fichero y de la clave o la sección. No se ignora en silencio ni
detiene la conversión. Las claves conocidas del frontmatter, medidas con tareas creadas por el CLI de
Backlog.md 1.52.0, son `id`, `title`, `status`, `assignee`, `created_date`, `updated_date`,
`due_date`, `labels`, `milestone`, `dependencies`, `references`, `documentation`, `modified_files`,
`priority`, `type`, `project`, `ordinal` y `parent_task_id`. Las secciones conocidas del cuerpo son
`DESCRIPTION`, `PLAN`, `NOTES`, `FINAL_SUMMARY`, los criterios de aceptación (`AC`), la definición de
hecho (`DOD`) y los comentarios (`COMMENTS`).

### Identificadores

1. **Forma de un id de origen.** Es `<PREFIJO>-<n>` o, para una subtarea, `<PREFIJO>-<n>.<m>`
   (Backlog.md crea `XYZ-001.01` o `ABC-1.2` cuando se le da un padre), con o sin ceros a la izquierda
   (`XYZ-001`). El prefijo del origen se deduce de los propios ids: todos tienen que compartirlo, sin
   distinguir mayúsculas, porque Backlog.md escribe siempre el prefijo en mayúsculas en el id aunque su
   configuración lo guarde de otra forma. Un id con otra forma, o dos prefijos distintos, es un error de
   origen (código 3). El prefijo del destino sale de `task_prefix` del destino.
2. **Un id simple de origen conserva su número con el prefijo del destino** (`TASK-70` pasa a
   `BISO-70`, `XYZ-001` pasa a `BISO-1`) siempre que ese número no exista ya en el destino. Los ceros a
   la izquierda no se escriben.
3. **Una tarea que ya está en el destino no se vuelve a importar.** Es la misma tarea de una
   importación anterior si el destino tiene una con el mismo `title` y el mismo `createdAt`, sea cual
   sea su id. No se escribe su línea, su id de origen se equipara al que ya tiene en el destino, y es un
   hallazgo (`already on the destination, skipped`). Así ejecutar `import` dos veces no duplica el
   tablero, subtareas incluidas. No actualiza la tarea que ya estaba, y una tarea que se haya editado
   en el destino después de importarla (otro título) ya no se reconoce.
4. **Reciben un número nuevo los ids simples que chocan con el destino y todos los ids de subtarea**,
   porque un id de `biso` es siempre `<PREFIJO>-<n>` y no admite el punto. Se les asigna, en el orden
   natural de su id de origen, el siguiente número libre a partir de `max(mayor número del destino,
   mayor número del origen) + 1`. Solo esos ids cambian de número; el resto no se desplaza.
5. Con esa tabla de equivalencias se reescriben el `id`, `parent` y `dependencies` de todas las
   tareas, y **toda mención de un id de origen en el texto** (`title`, `description`, `plan`,
   `notes`, `summary`, el texto de los criterios y el cuerpo de los comentarios): `TASK-12` pasa a
   `BISO-12`, o al número reasignado. Todas las menciones se sustituyen en una sola pasada, de modo que
   un número reasignado no vuelve a sustituirse. La coincidencia es el prefijo de origen en
   mayúsculas seguido de `-`, dígitos y, si los hay, de `.` y dígitos (`XYZ-001.01`),
   **distinguiendo mayúsculas**, y solo si no va precedida de una letra, un dígito, `_` o `-`, ni
   seguida de una letra, un dígito, `_` o de `-` y una letra o un dígito. Así `TASK-10-modelo`, que
   podría ser el nombre de una rama, no se toca, y tampoco `SUBTASK-12`. Se reescribe en cualquier
   lugar del texto, un bloque de código incluido. No se toca `documentation` ni `references`, que son
   rutas y URLs opacas.
6. Una mención que tiene la forma de un id de origen pero no corresponde a ninguna tarea del origen, o
   que solo difiere en las mayúsculas (`Xyz-002`, `task-12`), se deja como está y es un hallazgo,
   uno por fichero y campo con el recuento, para que quien ejecuta pueda revisarla.
7. Un `parent` o una dependencia que nombra un id que no está entre las tareas del origen se quita
   de la línea y es un hallazgo, porque `biso` rechaza el lote entero (código 4) por un solo
   identificador que no existe. Los ciclos de padres o de dependencias no se detectan: los rechaza
   `biso new --from --dry-run`, y por eso el ensayo forma parte del uso.
8. Cada id reasignado es un hallazgo: `XYZ-001.01: id BISO-97 assigned (subtask ids have no equivalent)`
   o `TASK-12: id BISO-12 is taken on the destination, reassigned to BISO-97`.

### Milestone y proyecto

El `milestone` de una tarea es un id de milestone (`m-4`). El título se lee del fichero
`milestones/m-4 - *.md` o del de `archive/milestones/`. El slug es el título en minúsculas, sin
diacríticos, con cada tramo de caracteres que no sean letras ni dígitos Unicode convertido en un solo
`-` y sin `-` en los extremos: `Puesta en uso` da `puesta-en-uso` e `Implementación` da
`implementacion`. La tarea recibe la etiqueta `milestone:puesta-en-uso`, después de las etiquetas que
ya tuviera. Una tarea sin milestone no recibe ninguna etiqueta y eso no es un hallazgo. Usa el id como
slug (`milestone:m-9`), con un hallazgo, cuando el fichero del milestone no existe o cuando el slug
sale vacío. Dos milestones distintos que dan el mismo slug se fusionarían en una sola etiqueta, así que
es un hallazgo.

El campo `project` de una tarea es un texto que ya viene escrito (`alpha`), sin fichero aparte. Recibe
la etiqueta `project:<slug>`, con el mismo cálculo del slug y después de la del milestone.

### Fechas

Backlog.md guarda `YYYY-MM-DD HH:mm`, en UTC. Se convierte a `YYYY-MM-DDTHH:mm:00Z`. Una fecha sin hora
(`YYYY-MM-DD`) pasa a `YYYY-MM-DDT00:00:00Z`. Si una tarea no tiene `updated_date`, `updatedAt` es su
`createdAt`. Una fecha que no encaja en ninguna de las dos formas es un hallazgo, y el campo se omite.

### Vocabularios del destino

`status`, `type` y `priority` se casan con los valores que declara el destino con la misma regla de
coincidencia que usa `biso`: un valor idéntico gana; si no, se comparan tras plegar mayúsculas y
minúsculas, quitar los diacríticos y quitar los espacios, guiones y guiones bajos. Si casa uno solo, es
ese. Si no casa ninguno, o casan varios, el campo se omite de la línea y es un hallazgo con el valor
original y la lista de los declarados. Omitir `status` hace que `biso` ponga el estado inicial.

### Alfabeto de un token

Un valor de `labels` o `assignees` solo puede contener en `biso` letras y dígitos Unicode y los
símbolos `- _ . : @`. Se aplica en este orden:

1. **Los espacios se convierten en guiones.** Se quita el espacio en blanco del principio y del final,
   y cada tramo de espacios, tabuladores u otro espacio en blanco pasa a un solo guion normal
   (`-`, U+002D): `with space` da `with-space` y `Sara Smith` da `Sara-Smith`. Cada conversión es un
   hallazgo.
2. Un valor que sigue sin cumplir el alfabeto (`a/b`, `c!`) se quita de la lista y es un hallazgo.
3. Si tras convertir dos valores de la misma tarea quedan iguales (`a b` y `a-b`), se deja uno solo.

Es la regla de `biso` al escribir, aplicada aquí para que el lote no falle entero por una etiqueta. La
conversión de los espacios no se deshace al exportar: `with-space` se queda como `with-space`.

### Criterios de aceptación

Una casilla empieza por `- [ ] #n ` o `- [x] #n `. **Un criterio puede ocupar varias líneas**: las
líneas que siguen a la casilla y no empiezan por otra casilla son la continuación de su texto
(`- [ ] #1 Multi` seguido de `line criterion` es un solo criterio). Como el texto de un criterio de
`biso` es un `string` y no admite un salto de línea, las líneas se unen con un espacio y es un
hallazgo. Una `#` o unos corchetes dentro del texto (`#hash`, `[x]`) no son casillas. Una sección sin
ninguna casilla no da criterios.

### Definición de hecho

La sección tiene el mismo formato que la de criterios: un encabezado `## Definition of Done`, las
marcas `<!-- DOD:BEGIN -->` y `<!-- DOD:END -->` y casillas `- [ ] #n texto`. Cada elemento se añade a
`acceptanceCriteria` con el texto seguido de un espacio y `#dod`, después de los criterios propios. Su
clave es la siguiente a la mayor clave de los criterios de la tarea, en orden. Su casilla se conserva.
La sección puede no existir y en ese caso no se añade nada. **La línea no lleva la clave
`definitionOfDone`**, así que `biso` no emite el aviso `imported_dod_merged`.

Backlog.md puede aplicar a cada tarea nueva una lista de elementos por defecto de la configuración del
proyecto (`definitionOfDone`), que se copia dentro de la sección de cada tarea junto con los propios.
En el fichero no se distinguen, así que todos reciben el sufijo, y un tablero configurado así da a cada
tarea esos criterios extra.

### Comentarios

La sección tiene un encabezado `## Comments`, las marcas `<!-- COMMENTS:BEGIN -->` y
`<!-- COMMENTS:END -->`, y dentro un bloque por comentario:

```
author: @ann
created: 2026-09-20 22:08
---
First comment
---
```

El cuerpo puede tener varias líneas y no puede contener una línea que sea solo `---`, porque Backlog.md
la reserva. Cada bloque da un elemento de `comments` con `author`, `createdAt` (la fecha, convertida
como en "Fechas") y `body`; la clave la asigna `biso` con su siguiente clave libre.

**La línea `author:` es opcional**: un comentario escrito sin `--comment-author` no la lleva y el bloque
empieza directamente por `created:`. Es un comentario sin autor, y `biso` admite un comentario sin autor
en un lote, así que no es un hallazgo. Un bloque sin `created` sí lo es, y la fecha se omite.

### Casos que no son obvios

Medidos con el CLI de Backlog.md 1.52.0:

- **El frontmatter es YAML de verdad.** Un título con comillas, dos puntos, `#` o corchetes se guarda
  entre comillas simples (`title: 'Title: with "quotes" # hash'`), así que se lee con un intérprete de
  YAML completo y no línea a línea.
- **Una tarea puede no tener ninguna sección.** Una tarea creada solo con título tiene el frontmatter y
  nada más. Todo campo ausente se omite de la línea.
- **Una sección puede contener `##` y `---`** en su texto sin dejar de ser una sola sección: se delimita
  por sus marcas, no por lo que contenga.
- **`priority` y `type` se guardan en minúsculas** aunque la configuración los escriba con mayúscula
  inicial (`priority: high` con la prioridad configurada `High`). El casado con el vocabulario del destino
  lo absorbe.
- **Los asignados se guardan tal como se dieron**, con o sin `@` y con espacios (`- Sara Smith`), y se
  tratan como cualquier otro valor del alfabeto de un token, con la conversión de los espacios.
- **Las etiquetas admiten cualquier carácter** (`a/b`, `c!`, `ñandú`, `with space`). Los espacios se
  convierten en guiones y solo pasan los demás que encajen en el alfabeto de `biso`. Ver "Alfabeto de
  un token".
- **`updated_date` no existe hasta que la tarea se edita.**
- **Un borrador promovido es una tarea normal**: `backlog draft promote` la mueve a `tasks/` con un id
  nuevo del prefijo de las tareas, el estado inicial y la fecha de creación que ya tenía. Como los
  borradores no se leen, solo llegan si están promovidos.
- **`backlog cleanup` mueve las tareas terminadas a `completed/`** sin cambiar su formato.

### Los hallazgos

Cada hallazgo es una línea por la salida de errores:

```
warning: <fichero>: <campo>: <mensaje>
```

`<fichero>` es el nombre del fichero de origen (o `-` si el hallazgo es de todo el tablero) y
`<campo>` la clave del frontmatter, la sección o el campo de `biso` afectado. El mensaje va en inglés.

### Códigos de salida

| Código | Cuándo |
|---|---|
| 0 | Convertido y sin ningún hallazgo |
| 2 | Uso incorrecto: falta un argumento o hay uno desconocido |
| 3 | El origen no se puede leer: no existe la carpeta, no tiene `tasks/`, un YAML no se puede interpretar, un id tiene una forma que no es la de "Identificadores", los ids no comparten prefijo o dos ficheros traen el mismo id |
| 4 | El destino no responde: no se encuentra `biso`, o falla una de sus órdenes |
| 5 | Convertido con hallazgos. El NDJSON está escrito, salvo con `--strict`, donde no se escribe nada |

**El código 5 no es un fallo**: el NDJSON está escrito. Encadenar `import` con `biso new --from` en un
`&&` se detiene en el 5, así que quien automatice el paso debe aceptarlo explícitamente.

**El convertidor no escribe nunca en el tablero destino.** Un fallo a mitad de la conversión deja el
destino intacto y, con `--out` a un fichero, no deja un fichero a medias: se escribe en un temporal y
se renombra al terminar.
