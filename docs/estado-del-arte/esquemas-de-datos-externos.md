# Esquemas de datos externos: seis gestores de tareas e incidencias

Inventario de hechos, sin comparar con `biso` ni proponer correspondencias de campos, para la tarea
TASK-37 del tablero. Investigado el 2026-09-15 sobre documentación oficial, esquemas publicados
(OpenAPI, GraphQL) y, cuando no había documentación exacta, el propio código o los ficheros de
esquema del proyecto. Cada afirmación no trivial lleva su fuente. La sección final, "Qué quedó sin
verificar", enumera lo que no se pudo confirmar del todo.

## 1. Taskwarrior

Fuente principal: la documentación oficial en <https://taskwarrior.org/docs/>. El formato de
almacenamiento interno (TaskChampion) representa cada task como un conjunto de pares clave-valor
de cadena; los atributos de más alto nivel (`project`, `due`, `priority`, `recur`...) se traducen a
esas claves por debajo, pero el usuario y los comandos operan sobre los nombres de atributo, no
sobre esa representación en bruto.

**Atributos, tal como los documenta la página "Task Representation"
(<https://taskwarrior.org/docs/task/>)**: el campo interno usa un prefijo por cada elemento de una
colección, no un array JSON anidado.

| Atributo | Formato interno | Significado |
|---|---|---|
| `status` | un carácter | `P` pendiente (por defecto), `C` completada, `D` borrada, `R` plantilla de recurrencia |
| `description` | cadena | "the one-line summary of the task" |
| `entry` | timestamp UNIX | "the time at which the task was created" |
| `modified` | timestamp UNIX | "the time of the last modification of this task" |
| `start` | timestamp UNIX | "the most recent time at which this task was started (a task with no `start` key is not active)" |
| `end` | timestamp UNIX | "if present, the time at which this task was completed or deleted" |
| `wait` | timestamp UNIX | "indicates the time before which this task should be hidden, as it is not actionable" |
| `tag_<tag>` | una clave por etiqueta | "indicates this task has tag `<tag>` (value is ignored)" |
| `annotation_<timestamp>` | una clave por anotación | "value is an annotation created at the given time" |
| `dep_<uuid>` | una clave por dependencia | "indicates this task depends on another task identified by `<uuid>`; the value is ignored" |

Además de esos, la "Terminology" (<https://taskwarrior.org/docs/terminology/>) documenta `uuid`
("A Universally Unique IDentifier is a 128-bit number used to uniquely identify a task"),
`project` (un atributo de cadena libre, ejemplo `task add project:Home Rake the leaves`), y `due`
("considered the date on which the task must be completed"). La recurrencia
(<https://taskwarrior.org/docs/recurrence/>) añade `recur` (patrón, p. ej. `"monthly"`), `parent`
(el UUID de la tarea plantilla del que procede cada instancia generada: **es una referencia por
UUID distinta del mecanismo de `depends`**) y `mask` (una cadena donde cada carácter resume el
estado de una instancia generada: `-` pendiente, `+` completada, `X` borrada, `W` en espera; "This
is how the template task keeps track of what it does and does not need to generate"); `imask` no
aparece descrito en ninguna página oficial encontrada.

**Priority es en sí misma una UDA desde la versión 2.4.3**, no un atributo del núcleo: "priority is
no longer a core attribute, and is replaced with an equivalent User Defined Attribute", con los
valores por defecto `H`, `M`, `L` y "no priority" (<https://taskwarrior.org/docs/priority/>). Es el
ejemplo más claro de cómo funcionan las UDA en la práctica.

**Cómo se declara una UDA** (<https://taskwarrior.org/docs/udas/>): con `task config`, fijando
`uda.NAME.type` (uno de `string`, `numeric`, `date`, `duration` o `uuid`), opcionalmente
`uda.NAME.label` (cabecera de columna), `uda.NAME.values` (lista cerrada de valores admitidos,
solo para el tipo `string`), `uda.NAME.default` y `urgency.uda.NAME.coefficient` (para que
contribuya al cálculo de urgencia). Cualquier clave no reconocida por Taskwarrior se trata como
UDA: "once configured, a UDA is indistinguishable from core attributes", y se conservan íntegras
en `import` y `export`.

**Exportación JSON** (<https://taskwarrior.org/docs/commands/export/>): `task export` saca un
objeto JSON por tarea; con `json.array` a `1` (su valor por defecto) los envuelve en un array
`[...]` separados por comas, con `json.array=0` escribe un objeto por línea sin envolver. Un
ejemplo de la propia página: `{"description":"Buy some milk","entry":"20141118T050231Z",
"status":"pending","uuid":"a360fc44-315c-4366-b70c-ea7e7520b749"}`. Campos como `id` y `urgency`
son "derived values, not stored values": se calculan al exportar y no viven en el almacén.

**Respuestas:**

- **Comentarios con autor, fecha y cuerpo:** no existe el concepto de comentario con autor.
  Existen **anotaciones** (`annotation_<timestamp>`), que son solo texto con una marca de tiempo,
  sin campo de autor porque Taskwarrior es de un único usuario local. No hay evidencia en la
  documentación de que una anotación se pueda editar tras crearse (solo se documentan los comandos
  para añadir).
- **Subtareas o criterios de aceptación con estructura propia:** no. La única jerarquía es la de
  plantilla de recurrencia y sus instancias (`parent`/`mask`), que no es una lista de subtareas.
- **Campo libre de pares clave-valor arbitrarios:** sí, es el propio mecanismo de UDA: cualquier
  clave no reservada se admite, se tipa y se conserva.
- **Dependencias y jerarquía:** `depends` (`dep_<uuid>`) es el único mecanismo de dependencia entre
  tareas, y es de bloqueo, no de jerarquía. La única relación padre-hijo es la de recurrencia
  (`parent`/`mask`), un mecanismo distinto y no reutilizable para modelar subtareas genéricas.
- **Pregunta abierta dirigida a una persona con estado pendiente/respondida:** no existe; Taskwarrior
  no tiene concepto de destinatario ni de personas más allá de quien ejecuta el comando.
- **Ficheros modificados / documentación asociada / referencias externas:** no hay ningún atributo
  de este tipo en el núcleo; solo se podría simular con una UDA de tipo `string`.
- **Identificador estable en listas:** las anotaciones se identifican por su timestamp de creación
  (que actúa de clave), y las dependencias por el UUID de la tarea referenciada, no por posición.
  El `id` numérico visible en pantalla, en cambio, es **el único identificador inestable**: se
  recalcula en cada informe y no sobrevive a nada; el identificador permanente es siempre el `uuid`.

## 2. Beads (`bd`)

Fuente principal: la documentación del propio repositorio
<https://github.com/gastownhall/beads>, su sitio <https://beads.gascity.com/>, y el código cuando
la documentación no alcanzaba. Beads representa cada incidencia como una fila de una base de datos
(SQLite hasta la versión 0.49, Dolt desde la 0.50) con `issues.jsonl` como exportación de
interoperabilidad, nunca como la fuente de la verdad
(<https://raw.githubusercontent.com/gastownhall/beads/main/docs/core-concepts/sync-concepts.md>).

**Campos de un issue**, según el ejemplo documentado en
`docs/core-concepts/issues.md` (<https://github.com/gastownhall/beads/blob/main/docs/core-concepts/issues.md>)
y los comandos de `bd update` citados en su documentación de referencia:

| Campo | Tipo | Significado |
|---|---|---|
| `id` | cadena | identificador corto, p. ej. `bd-42` (ver más abajo cómo se genera) |
| `title` | cadena | título del issue |
| `description` | cadena | alcance del trabajo, texto libre |
| `type` | cadena cerrada | uno de `bug`, `feature`, `task`, `epic`, `chore` |
| `status` | cadena | p. ej. `open`, `in_progress` |
| `priority` | entero 0-4 | 0 crítico, 4 backlog; "affects the ready queue" |
| `labels` | array de cadenas | etiquetas libres, p. ej. `backend`, `security` |
| `created_at` / `updated_at` | timestamp ISO 8601 | fechas de creación y última modificación |
| `design` | cadena | fijado con `bd update <id> --design`, notas de diseño |
| `acceptance_criteria` | cadena de texto libre | fijado con `bd update <id> --acceptance`; **no es una
  lista estructurada de ítems marcables, es un bloque de texto** |
| `notes` | cadena | fijado con `bd update <id> --notes`, para "rationale or fallback context" |
| `payload` | JSON arbitrario | dato estructurado libre asociado al issue |
| `waiters` | array JSON | usado internamente para saber quién espera a que el issue se desbloquee |
| `metadata` | JSON arbitrario | ver más abajo |

(Los campos `design`, `acceptance_criteria`, `notes`, `payload` y `waiters` están confirmados por
la propia documentación de `bd list --brief`, que dice que ese modificador "omits the free-form
text fields including description, design, acceptance_criteria, notes, payload, and waiters".)

**El campo `metadata` es JSON completamente libre**: "any valid JSON value is stored as-is"
(<https://raw.githubusercontent.com/gastownhall/beads/main/docs/core-concepts/metadata.md>). Se lee
y escribe solo por JSON/`jq` (`bd show <id> --json | jq '.[0] | {id,title,metadata,...}'`); la
propia documentación rechaza explícitamente dar un atajo de línea de comandos para él ("Do not add
a first-class helper such as `bd show <id> --execution`"). Sus usos documentados son tres: pistas
de ejecución para agentes (qué modelo o nivel de razonamiento usar), preservar campos de un sistema
de tickets externo que no tienen equivalente nativo en beads (para sincronización de ida y vuelta),
y como "the preferred extension point" antes de proponer un cambio de esquema nativo. Tiene
prefijos reservados que no se deben usar: `bd:` (uso interno de beads) y `_` (claves privadas).

**Dependencias y jerarquía usan el mismo mecanismo, con un campo de tipo.** Beads modela toda
relación entre issues como una arista con un `type`
(<https://raw.githubusercontent.com/gastownhall/beads/main/docs/core-concepts/dependencies.md>).
Tipos que bloquean el flujo: `blocks` (por defecto, "B cannot start until A closes"),
`parent-child` ("children blocked when parent blocked"), `conditional-blocks` ("B runs only if A
fails"), `waits-for` ("B waits for all of A's children"). Tipos que no bloquean, solo anotan el
grafo: `related`, `tracks`, `discovered-from`, `caused-by`, `validates`, `supersedes`. **La
jerarquía padre-hijo no tiene un campo separado**: es literalmente `--type parent-child` sobre la
misma tabla de aristas que usa `blocks`.

**Comentarios.** Existe `bd comment <id> "texto"`, atajo de `bd comments add <id> "texto"`
(<https://beads.gascity.com/cli-reference/comment>), y la familia `bd comments` para
verlos/añadirlos (<https://beads.gascity.com/cli-reference/comments>), con salida en JSON
(`--json`) y un flag `-a, --author` opcional para fijar el autor. La documentación fetcheada no
muestra ni un subcomando de edición ni uno de borrado de un comentario ya creado, solo `add` y la
vista por issue; no se puede afirmar con certeza que sean inmutables, pero no hay evidencia de lo
contrario.

**Generación de identificadores.** El `id` corto (`bd-a1b2`) se deriva por hash de "issue title,
creation timestamp, random salt", y es "deterministic for same content+timestamp"
(<https://raw.githubusercontent.com/gastownhall/beads/main/docs/core-concepts/hash-ids.md>). En una
colisión al importar, "Beads appends disambiguator, both issues preserved": nunca se pierde un
issue por choque de hash entre ramas.

**Respuestas:**

- **Comentarios con autor, fecha y cuerpo:** sí (autor opcional, marca de tiempo, cuerpo); no se
  encontró documentación de edición o borrado.
- **Subtareas o criterios de aceptación con estructura propia:** no. `acceptance_criteria` es texto
  libre, no una lista de ítems con su propia clave y marcado; la descomposición en piezas más
  pequeñas se hace creando issues hijos por dependencia `parent-child`, no con una sublista dentro
  del propio issue.
- **Campo libre de pares clave-valor arbitrarios:** sí, `metadata`, JSON sin esquema.
- **Dependencias y jerarquía:** mismo mecanismo (aristas con `type`); la jerarquía es solo un valor
  de ese campo (`parent-child`), no una relación distinta.
- **Pregunta abierta a una persona con estado pendiente/respondida:** no se encontró nada así; lo
  más cercano es `waiters`, que es una lista interna de qué espera a qué, no una pregunta dirigida
  a alguien.
- **Ficheros modificados / documentación / referencias externas:** no hay un campo nativo para
  esto; se podría meter en `metadata`, que es justamente el mecanismo pensado para extensión.
- **Identificador estable en listas:** las etiquetas (`labels`) son un array de cadenas sin id
  propio, así que se identifican por su valor. Las dependencias se referencian por el `id` del
  issue relacionado (estable). No se encontró en la documentación fetcheada un ejemplo de campo
  `id` para cada comentario individual, aunque `bd comments --json` sí devuelve estructura por
  comentario.

## 3. Task Master (`claude-task-master`)

Fuente: `docs/task-structure.md` del propio repositorio
(<https://github.com/eyaltoledano/claude-task-master/blob/main/docs/task-structure.md>). Desde la
versión 0.16.2 el almacén tiene un nivel adicional de "tags": `tasks.json` es un objeto cuyas
claves son nombres de tag (por defecto `"master"`) y cuyo valor es `{"tasks": [...]}`; los ficheros
legados de una sola lista se migran en el sitio la primera vez que se usan.

**Campos de una tarea:**

| Campo | Tipo | Obligatorio | Significado |
|---|---|---|---|
| `id` | número | sí | identificador único dentro del tag |
| `title` | cadena | sí | título breve |
| `description` | cadena | sí | resumen de qué implica la tarea |
| `status` | cadena cerrada | sí | uno de `pending`, `in-progress`, `done`, `review`, `deferred`, `cancelled` |
| `dependencies` | array de números | no | ids de tareas previas que deben completarse antes |
| `priority` | cadena cerrada | no | `high`, `medium` o `low`; por defecto `medium` |
| `details` | cadena | no | instrucciones de implementación en profundidad |
| `testStrategy` | cadena | no | cómo verificar que la tarea está bien hecha |
| `subtasks` | array de subtareas | no | ver estructura abajo |

**Estructura de una subtarea:** casi idéntica a la de una tarea, pero su `id` es un número único
solo dentro de la tarea padre (no global), no tiene `priority` propio, y sus `dependencies` pueden
apuntar tanto a otras subtareas hermanas como a ids de tareas de primer nivel. Se referencia desde
fuera con la notación `taskId.subtaskId` (por ejemplo `8.2`), tal como documenta el comando `show`
("Works with both regular tasks and subtasks (using the format taskId.subtaskId)").

**Validaciones documentadas:** ids únicos dentro de cada tag, valores de `status` restringidos al
conjunto cerrado, las `dependencies` deben referenciar ids existentes dentro del mismo tag, ids de
subtarea únicos dentro de su tarea padre, y detección de dependencias circulares. La documentación
no dice qué ocurre con los ids al borrar una tarea (si se renumeran o quedan huecos); no se
encontró una afirmación explícita al respecto.

Hay además un informe aparte, no parte de `tasks.json`, generado por `analyze-complexity`
(`scripts/task-complexity-report.json` por defecto): una puntuación de complejidad 1-10 por tarea,
número de subtareas recomendado, y un prompt de expansión generado por IA para cada una.

**Respuestas:**

- **Comentarios con autor, fecha y cuerpo:** no existe ningún campo de comentario en la estructura
  documentada.
- **Subtareas o criterios de aceptación con estructura propia:** sí hay subtareas con su propia
  estructura (id, título, descripción, estado, dependencias, detalles), pero no hay un campo
  separado de "criterios de aceptación" con su propio marcado; `testStrategy` cumple ese papel como
  texto libre, no como lista de ítems marcables.
- **Campo libre de pares clave-valor arbitrarios:** no se documenta ningún campo de metadata o
  extensión libre en la estructura de tarea.
- **Dependencias y jerarquía:** `dependencies` es el único mecanismo de relación entre tareas, y es
  de bloqueo/orden, no de propiedad. La "jerarquía" tarea-subtarea no es un campo de dependencia
  sino de anidamiento físico (`subtasks` es un array dentro de la tarea), un mecanismo distinto del
  de `dependencies`.
- **Pregunta abierta a una persona con estado pendiente/respondida:** no existe.
- **Ficheros modificados / documentación / referencias externas:** no hay ningún campo así en la
  estructura de tarea o subtarea documentada.
- **Identificador estable en listas:** los ids de tarea y de subtarea son números asignados por
  posición de creación dentro de su ámbito (tag o tarea padre); la documentación no dice si borrar
  una tarea intermedia deja huecos o renumera las siguientes, así que no se puede afirmar que el id
  sobreviva con certeza a una eliminación ajena sin volver a probarlo contra el código.

## 4. GitHub Issues (API REST)

Fuente: la referencia oficial en <https://docs.github.com/en/rest/issues/issues>,
<https://docs.github.com/en/rest/issues/comments>, <https://docs.github.com/en/rest/issues/labels>,
<https://docs.github.com/en/rest/issues/milestones>, <https://docs.github.com/en/rest/issues/sub-issues>
y <https://docs.github.com/en/rest/issues/issue-dependencies>. Los campos puramente de plataforma
(`node_id`, las distintas variantes de `url`, `reactions`, `performed_via_github_app`) se
mencionan solo de pasada, como pide la tarea.

**Campos relevantes al contenido de un issue:**

| Campo | Tipo | Significado |
|---|---|---|
| `title` | cadena | título |
| `body` | cadena o `null` | cuerpo en Markdown; la API también puede devolver `body_html` y `body_text` según la cabecera `Accept` de la petición |
| `state` | `open` \| `closed` | estado binario |
| `state_reason` | `completed` \| `reopened` \| `not_planned` \| `null` | por qué se cerró o reabrió; un `duplicate` se añadió más tarde y generó una discusión pública porque rompía a clientes que asumían el conjunto cerrado original (<https://github.com/orgs/community/discussions/150535>) |
| `user` | Simple User | quien creó el issue |
| `assignees` | array de Simple User | personas asignadas (varias); existe también un campo `assignee` singular por compatibilidad histórica |
| `labels` | array de objetos Label | ver tabla siguiente |
| `milestone` | objeto Milestone o `null` | ver tabla siguiente |
| `locked` / `active_lock_reason` | booleano / cadena o `null` | si la conversación está bloqueada y por qué |
| `comments` | entero | recuento de comentarios (no su contenido) |
| `closed_by` | Simple User o `null` | quien lo cerró |
| `created_at` / `updated_at` / `closed_at` | fecha-hora | ciclo de vida temporal |

**Label** (<https://docs.github.com/en/rest/issues/labels>): `name`, `description` (cadena o
`null`, "must be 100 characters or fewer"), `color` (cadena hexadecimal sin `#`), `default`
(booleano, si es una etiqueta estándar provista por GitHub).

**Milestone** (<https://docs.github.com/en/rest/issues/milestones>): `title`, `description`,
`state` (`open`/`closed`), `creator`, `open_issues`/`closed_issues` (recuentos), `created_at`,
`updated_at`, `due_on`, `closed_at`.

**Comentario** (<https://docs.github.com/en/rest/issues/comments>): `id`, `body` (y también
`body_html`/`body_text` según `Accept`), `user`, `created_at`, `updated_at`. **Son editables**: "You
can use the REST API to update comments on issues and pull requests" mediante `PATCH`. Se listan en
orden ascendente de `id` por defecto, aunque la colección a nivel de repositorio admite ordenar por
`created` o `updated`.

**Jerarquía y dependencias son dos mecanismos distintos, y los dos son de 2024-2025.** La jerarquía
usa la API de sub-issues (<https://docs.github.com/en/rest/issues/sub-issues>):
`GET/POST /repos/{owner}/{repo}/issues/{issue_number}/sub_issues`,
`GET /repos/{owner}/{repo}/issues/{issue_number}/parent`, con un límite de 100 sub-issues por
padre y un endpoint propio para reordenarlos (`PATCH .../sub_issues/priority`). El bloqueo usa la
API de dependencias (<https://docs.github.com/en/rest/issues/issue-dependencies>), añadida en
agosto de 2025: `GET/POST/DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by`
y el simétrico `.../dependencies/blocking`, representando cada dependencia como una referencia a
otro issue por su `issue_id`. El propio issue trae un resumen, `issue_dependencies_summary`, con
los recuentos `blocked_by` y `blocking`.

**Nota aparte sobre metadata:** el objeto Issue en sí no tiene ningún campo de clave-valor libre.
GitHub Projects (v2) sí permite definir campos personalizados (texto, número, fecha, selección
única, iteración...) pero **viven en el objeto "item" del proyecto, no en el Issue**, y se
gestionan por una API GraphQL aparte
(<https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects>).

**Respuestas:**

- **Comentarios con autor, fecha y cuerpo:** sí (`user`, `created_at`/`updated_at`, `body`), y son
  editables mediante `PATCH`, no inmutables.
- **Subtareas o criterios de aceptación con estructura propia:** no como campo del objeto Issue. La
  convención de listas de tareas (`- [ ] texto`) dentro del `body` en Markdown no es un campo de la
  API con estructura propia, es texto; los sub-issues sí son objetos Issue completos enlazados, no
  una lista de ítems ligeros.
- **Campo libre de pares clave-valor arbitrarios:** no en el objeto Issue; existe en Projects v2,
  que es un recurso distinto.
- **Dependencias y jerarquía:** son mecanismos distintos, con familias de endpoints separadas
  (sub-issues para jerarquía, issue-dependencies para bloqueo), no el mismo campo.
- **Pregunta abierta a una persona con estado pendiente/respondida:** no existe en el objeto Issue.
  Lo más cercano en la plataforma es la revisión de un pull request (reviewers solicitados,
  aprobado/cambios solicitados), que es un objeto y un flujo distintos del de un issue.
- **Ficheros modificados / documentación / referencias externas:** el objeto Issue no tiene un
  campo de este tipo (el listado de ficheros modificados es propio del objeto Pull Request, fuera
  de alcance aquí).
- **Identificador estable en listas:** sí. Cada comentario tiene su propio `id` numérico estable
  que sobrevive a que se borre otro comentario de la lista; lo mismo vale para labels, milestones y
  sub-issues, todos referenciados por su `id`/`number`, nunca por posición.

## 5. Linear (API GraphQL)

Fuente: el esquema GraphQL público del SDK oficial,
<https://github.com/linear/linear/blob/master/packages/sdk/src/schema.graphql> (descargado y
consultado línea a línea el 2026-09-15), y la documentación de estados de flujo en
<https://linear.app/docs/configuring-workflows>.

**Campos relevantes del tipo `Issue`** (de entre todos los campos del tipo, bastantes más de los
que interesan aquí, filtrando los puramente internos o de sincronización):

| Campo | Tipo | Significado |
|---|---|---|
| `title` | `String!` | título |
| `description` | `String` | cuerpo en formato Markdown |
| `state` | `WorkflowState!` | estado de flujo (ver abajo) |
| `priority` | `Float!` | prioridad numérica; `priorityLabel` da su nombre legible y `prioritySortOrder` su orden |
| `assignee` | `User` | persona asignada (una sola); `null` si no está asignado |
| `creator` | `User` | quien creó el issue |
| `delegate` | `User` | el agente de IA delegado para trabajar el issue, si lo hay |
| `labels(...)` | `IssueLabelConnection!` | etiquetas asociadas, paginadas |
| `labelIds` | `[String!]!` | los mismos ids de etiqueta, sin paginar |
| `estimate` | `Float` | puntos de estimación |
| `dueDate` | `TimelessDate` | fecha límite, sin hora |
| `parent` | `Issue` | issue padre; `null` si es de nivel superior |
| `children(...)` | `IssueConnection!` | issues hijos (sub-issues) |
| `subIssueSortOrder` | `Float` | orden entre los hijos de un mismo padre |
| `project` | `Project` | proyecto al que pertenece |
| `cycle` | `Cycle` | ciclo (sprint) al que pertenece |
| `comments(...)` | `CommentConnection!` | comentarios, "including inline comments on the issue's description" |
| `relations(...)` | `IssueRelationConnection!` | relaciones salientes con otros issues (ver abajo); `inverseRelations` da las entrantes |
| `identifier` | `String!` | identificador legible, p. ej. `ENG-123` |
| `id` | `ID!` | identificador interno estable (UUID) |
| `createdAt` / `updatedAt` / `completedAt` / `canceledAt` / `startedAt` / `triagedAt` | `DateTime` | ciclo de vida temporal |

**El estado (`state: WorkflowState!`) no es un enum de texto libre**, es una entidad propia con
`name` (el nombre visible, personalizable por equipo), `color`, `position`, y un `type: String!`
que sí es un conjunto fijo documentado en la página de flujos de trabajo
(<https://linear.app/docs/configuring-workflows>): `triage`, `backlog`, `unstarted`, `started`,
`completed`, `canceled`. Ese `type` es lo que da la semántica ("completado", "cancelado"), y `name`
es solo la etiqueta que ve el equipo.

**`Comment`:** `id`, `body` (Markdown; "a derived representation of the canonical bodyData
ProseMirror content"), `user` (`null` si lo creó un bot o una integración), `createdAt`,
`editedAt` (`null` si nunca se editó, lo que confirma que **sí son editables**), `parent`/`parentId`
(comentario padre, para hilos anidados: "Null for top-level comments that are not replies"),
`resolvedAt`/`resolvingUser`/`resolvingComment` (hilo resuelto y quién/qué lo resolvió), `issue`
(a qué issue pertenece, o `null` si pertenece a otro tipo de entidad como un proyecto o una
iniciativa).

**Dependencias y jerarquía son dos mecanismos distintos.** La jerarquía es el par
`parent`/`children` directamente en el tipo `Issue`. Las dependencias de bloqueo son un tipo
aparte, `IssueRelation`, con `issue` (origen), `relatedIssue` (destino) y `type`, cuyo enumerado
`IssueRelationType` tiene los valores `blocks`, `duplicate`, `related`, `similar`. Es decir,
jerarquía y bloqueo no comparten tabla ni campo, a diferencia de Beads.

**No existe ningún campo de metadata o clave-valor arbitraria** en el tipo `Issue` ni en `Comment`:
todos los campos de `Issue` son de dominio fijo (fechas de ciclo de vida, relaciones a
otras entidades tipadas, campos de sincronización con sistemas externos como `syncedWith`, pero
ninguno es un mapa JSON libre para quien usa la API).

**Respuestas:**

- **Comentarios con autor, fecha y cuerpo:** sí (`user`, `createdAt`, `body`), y son editables
  (`editedAt` distingue si se editaron). Además admiten hilos (`parent`/`parentId`) y un estado de
  resuelto/no resuelto (`resolvedAt`).
- **Subtareas o criterios de aceptación con estructura propia:** no hay un campo de checklist
  dentro del issue; la descomposición se hace con sub-issues completos (`children`), que son
  issues de pleno derecho, no ítems ligeros con su propio marcado binario.
- **Campo libre de pares clave-valor arbitrarios:** no existe en el esquema público.
- **Dependencias y jerarquía:** dos mecanismos distintos: `parent`/`children` para jerarquía,
  `IssueRelation` (`type: blocks|duplicate|related|similar`) para todo lo demás, incluido el
  bloqueo.
- **Pregunta abierta a una persona con estado pendiente/respondida:** no existe como concepto de
  primera clase. Lo más cercano es el hilo de comentarios resuelto/no resuelto
  (`resolvedAt`/`resolvingUser`), pero eso es un estado de la conversación entera, no una pregunta
  dirigida a una persona concreta con su propio pendiente/respondida.
- **Ficheros modificados / documentación / referencias externas:** el tipo `Issue` tiene
  `attachments` (no detallado aquí por no ser el foco de la pregunta, pero existe como conexión
  aparte) para adjuntar URLs externas; no hay un campo de "ficheros modificados" porque Linear no
  es un sistema de control de versiones.
- **Identificador estable en listas:** sí. Todo elemento de una conexión (comentarios, relaciones,
  hijos, etiquetas) es un nodo con su propio `id: ID!` estable; el orden se maneja aparte con
  campos explícitos como `subIssueSortOrder` o `sortOrder`, nunca por posición implícita en un
  array.

## 6. Trello (API REST)

Fuente: la especificación OpenAPI oficial publicada en
<https://developer.atlassian.com/cloud/trello/swagger.v3.json> (descargada y consultada el
2026-09-15) y las páginas de referencia en
<https://developer.atlassian.com/cloud/trello/rest/api-group-cards/>,
`.../api-group-checklists/`, `.../api-group-actions/` y `.../api-group-customfields/`.

**Campos de una Card**, según el enumerado `CardFields` del propio esquema OpenAPI (la lista
completa y autorizada de nombres de campo válidos de una tarjeta; algunos no figuran con su tipo en
el objeto `Card` reducido de la misma especificación, pero sí en este enumerado):

`id`, `address`, `badges`, `checkItemStates` (heredado, para compatibilidad), `closed` (si está
archivada), `coordinates`, `creationMethod`, `dueComplete` (booleano, si la fecha límite se marcó
hecha), `dateLastActivity`, `desc` (cuerpo, hasta 16.384 caracteres), `descData`, `due` (fecha
límite), `dueReminder`, `idBoard`, `idChecklists` (array de ids de checklist), `idLabels`,
`idList` (**la lista es el mecanismo de estado de una tarjeta**: no hay un campo `status`
independiente, el estado es en qué lista vive), `idMembers`, `idMembersVoted`, `idShort`,
`idAttachmentCover`, `labels` (objetos completos, no solo ids), `limits`, `locationName`,
`manualCoverAttachment`, `name` (título), `pos` (posición dentro de la lista, de tipo numérico
flotante), `shortLink`, `shortUrl`, `subscribed`, `url`, `cover`, `isTemplate`.

**Label**: `id`, `idBoard`, `name`, `color`.

**Checklist y CheckItem** (esquema OpenAPI, grupo `api-group-checklists`): un `Checklist` tiene
`id`, `name`, `idCard`, `pos`, y una colección de `checkItems`. Cada `CheckItem` tiene `id`,
`idChecklist`, `name`, `nameData` (nulo salvo casos especiales), `pos`, y `state`, un enumerado
cerrado con exactamente dos valores: `complete` e `incomplete`. Es la estructura más parecida a un
criterio de aceptación marcable de las seis: **una clave (`id`), un texto (`name`) y un marcado
binario (`state`)**, aunque su nombre en Trello es "elemento de checklist", no "criterio de
aceptación".

**Custom Fields, el equivalente más cercano a un campo de metadata, pero con esquema, no libre**
(grupo `api-group-customfields`): un `CustomField` se define primero a nivel de tablero, con
`idModel`, `modelType` (`card`, `board` o `member`), `type` (p. ej. `list`, y también texto, número,
fecha, casilla) y, si es de tipo lista, sus opciones (`display.options`, cada una con su propio
`id`, `value.text` y `color`). Solo después de definirlo en el tablero se le puede asignar un valor
a una tarjeta concreta mediante un `CustomFieldItems`, que referencia `idCustomField` + `idModel`
(la tarjeta) y guarda el valor bajo una de las claves tipadas (`value.text`, `value.number`,
`value.date`, o `value.checked` para las casillas). **No es una bolsa JSON arbitraria**: cada
tarjeta solo puede rellenar campos que el tablero ya declaró, con el tipo que el tablero fijó; es
más parecido a añadir columnas a un esquema por adelantado que a un `metadata` libre por tarjeta.

**Comentarios son objetos Action, no un campo de la tarjeta.** Un comentario es una `Action` de
`type: "commentCard"`, con `id`, `idMemberCreator` (y el objeto `memberCreator` completo), `date`,
y el texto en `data.text` (grupo `api-group-actions`). **Son editables y borrables** mediante
endpoints dedicados: `PUT /actions/{id}/text` ("used to edit the content of a comment") y
`DELETE /actions/{id}` ("only comment actions can be deleted"), ambos exigiendo los scopes
`read:board:trello` y `write:board:trello`.

**Adjuntos** (`Attachment`, en el mismo esquema OpenAPI): `id`, `name`, `url`, `date`, `idMember`,
`mimeType`, `bytes`, `pos`. Es el mecanismo de Trello para enlazar una referencia externa
(documento, enlace) a una tarjeta; no hay un campo separado de "documentación asociada".

**Respuestas:**

- **Comentarios con autor, fecha y cuerpo:** sí (`idMemberCreator`/`memberCreator`, `date`,
  `data.text`), y son editables y borrables mediante endpoints propios, no inmutables.
- **Subtareas o criterios de aceptación con estructura propia:** sí, el `CheckItem` dentro de un
  `Checklist`: id, texto (`name`) y marcado binario (`state: complete|incomplete`). Es de los seis
  sistemas el que más se parece literalmente a un criterio de aceptación marcable.
- **Campo libre de pares clave-valor arbitrarios:** no de forma libre; existen los Custom Fields,
  pero exigen declarar el campo (nombre y tipo) a nivel de tablero antes de poder asignarle un
  valor en una tarjeta, así que es un esquema extensible, no una bolsa JSON sin forma.
- **Dependencias y jerarquía:** ninguna de las dos existe de forma nativa en el objeto Card de la
  API pública. No hay campo de "bloqueado por" ni de tarjeta padre/hija; el Power-Up nativo de
  "Card dependencies" que se ve en la interfaz de usuario no aparece como recurso de la API REST
  pública documentada.
- **Pregunta abierta a una persona con estado pendiente/respondida:** no existe.
- **Ficheros modificados / documentación / referencias externas:** sí, como adjuntos (`Attachment`
  con `url`), aunque están pensados para cualquier enlace o fichero, no específicamente para
  ficheros de código o documentación de un repositorio.
- **Identificador estable en listas:** sí. Cada `CheckItem`, cada `Label` y cada `Action`
  (comentario) tiene su propio `id` de tipo `TrelloID` (una cadena estilo ObjectId de MongoDB),
  estable frente al borrado de otros elementos de la misma lista; el campo `pos` solo ordena, no
  identifica.

## Qué quedó sin verificar

1. **La lista completa de atributos del núcleo de Taskwarrior** (`uuid`, `due`, `recur`, `parent`,
   `mask`, `imask`, `priority`) no vive junta en una sola página con las mismas palabras citables:
   se reconstruyó cruzando "Task Representation", "Terminology", "Priority" y "How Recurrence
   Works". El atributo `imask` en concreto no apareció descrito en ninguna página oficial
   consultada, solo mencionado de pasada en el propio JSON de ejemplo de recurrencia.
2. **Si los comentarios de Beads (`bd comment`/`bd comments`) son editables o borrables** no se
   pudo confirmar ni negar con una cita directa: la documentación fetcheada solo documenta `add` y
   la vista por issue, sin mencionar ni un subcomando de edición ni uno de borrado.
3. **El esquema OpenAPI de Trello está repartido en dos objetos parcialmente redundantes**
   (`Card` y el enumerado `CardFields`) que no siempre coinciden: `dueComplete` e `isTemplate`
   aparecen en `CardFields` pero no en las propiedades tipadas de `Card` dentro del mismo fichero.
   Se cruzaron ambos fragmentos del mismo documento para dar la lista completa, pero no se pudo
   confirmar el tipo exacto de cada uno de esos dos campos concretos más allá de su nombre.
4. **Si Task Master renumera o deja huecos en los `id` al borrar una tarea** no está dicho en
   `docs/task-structure.md`; solo se documentan las validaciones de unicidad, no el comportamiento
   de borrado. Haría falta leer el código de `removeTask` para confirmarlo, y no se hizo por no ser
   imprescindible para el inventario de esquema.
5. **Los valores exactos que puede tomar `WorkflowState.type` en Linear** no están en un enum del
   propio esquema GraphQL (el campo está tipado como `String!`, sin restricción a nivel de schema);
   los seis valores citados (`triage`, `backlog`, `unstarted`, `started`, `completed`, `canceled`)
   vienen de la página de producto sobre flujos de trabajo, no de una fuente que los declare como
   conjunto cerrado a nivel de API.
