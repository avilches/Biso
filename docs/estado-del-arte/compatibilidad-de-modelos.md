# 4. Compatibilidad del modelo de datos con otros gestores

Esta página cruza el modelo de datos de `biso`
([El modelo de datos de una tarea](../spec/modelo-de-datos/index.md) y sus páginas hijas) contra el
inventario de campos de [Esquemas de datos externos](esquemas-de-datos-externos.md), que reúne los
hechos citados de cada sistema externo, campo a campo, y decide qué compatibilidad se puede prometer
en la especificación.

Backlog.md se investigó aparte, directamente contra el CLI 1.52.0 instalado en esta máquina (su
`--help`, su `Input schema`, y tareas de prueba creadas y borradas en un worktree para observar el
frontmatter real), porque es la herramienta que este mismo repositorio usa para su propio tablero y
la que motivó la pregunta.

**Esta página ya no cruza el campo `definitionOfDone`,** que se retiró del modelo de datos el
2026-09-19 (["Se retira la definición de hecho"](../decisiones/detalles.md#se-retira-la-definición-de-hecho)).
Lo que queda vivo no es una correspondencia sino una regla de entrada: un lote de `biso new --from`
que traiga esa clave, como el que saldría de un tablero de Backlog.md, convierte cada elemento en un
criterio de aceptación conservando su texto y su marcado, y avisa de ello (["`biso new`"](../spec/cmd/new.md)).
Las tablas de abajo, por tanto, tienen una fila menos que las versiones anteriores de este documento.

## Método

Cada campo guardado de una tarea de `biso` (los automáticos y los fijados por el
usuario; los derivados no entran, porque nunca se guardan y ["El modelo de
datos"](../spec/modelo-de-datos/index.md#los-campos-derivados) ya dice que ni `biso export` los
escribe ni `biso new --from` los acepta) recibe una de tres correspondencias frente a cada sistema:

- **Directa**: el sistema tiene un campo de dominio, reconocido por sus propios comandos o su API,
  con la misma semántica; solo cambia el nombre o el formato.
- **Con transformación**: hay un campo o mecanismo capaz de llevarse el valor sin perder la
  estructura (una lista sigue siendo una lista, una fecha sigue siendo una fecha), pero exige mapear
  un vocabulario distinto, cambiar de tipo, o aceptar una pérdida de precisión o de granularidad que
  se explica en la fila.
- **Sin equivalente**: no hay ningún campo del sistema, reconocido por sus propias herramientas, que
  pueda recibir el valor sin aplanarlo a texto opaco o sin perder la estructura por completo.

**Una decisión importante, y deliberadamente conservadora: un campo de extensión libre sin
esquema (la `metadata` JSON de Beads, o una UDA de Taskwarrior inventada por `biso` con un nombre que
ningún otro proyecto usa) no cuenta como transformación para los campos que ese sistema no modela de
forma nativa, aunque el dato quepa ahí técnicamente.** Meter ahí el resto del modelo de `biso`
preservaría los bytes, pero no la compatibilidad: ningún comando nativo de esa herramienta sabría
filtrar, mostrar o editar ese dato, así que la tarea se volvería ilegible para quien no conozca la
convención de `biso`. La única excepción es el propio campo `ext` de `biso`, que por construcción ya
es un mapa de extensión y por tanto sí corresponde, de forma directa o con transformación de tipo,
con el mecanismo de extensión del sistema destino cuando existe uno.

**Los dos sentidos casi siempre dan el mismo número, y donde no, se explica en prosa y no en la
cifra.** La compatibilidad estructural (si los dos modelos tienen un sitio equivalente para un campo)
es la misma mirando desde cualquiera de los dos lados; lo que cambia de sentido es la pérdida
concreta, no la categoría: perder granularidad de `status` al exportar a un sistema con menos
estados no es lo mismo que no perder nada al importar esos mismos estados de vuelta, pero en ambos
casos el campo sigue siendo "con transformación". Cada tabla dice explícitamente, en su prosa de
después, en qué sentido concreto ocurre cada pérdida.

**El porcentaje es (directa + transformación) / 28**, sobre el total de campos guardados de `biso`,
igual en los dos sentidos salvo que la prosa de después de la tabla diga lo contrario para un campo
suelto.

## Resumen

| Sistema | Cobertura al exportar | Cobertura al importar |
|---|---|---|
| Backlog.md | 23/28 ≈ 82 % | 23/28 ≈ 82 % |
| Linear | 17/28 ≈ 61 % | 17/28 ≈ 61 % |
| Trello | 16/28 ≈ 57 % | 16/28 ≈ 57 % |
| Beads (`bd`) | 16/28 ≈ 57 % | 16/28 ≈ 57 % |
| GitHub Issues | 13/28 ≈ 46 % | 13/28 ≈ 46 % |
| Taskwarrior | 12/28 ≈ 43 % | 12/28 ≈ 43 % |
| Task Master | 9/28 ≈ 32 % | 9/28 ≈ 32 % |

**La misma comparación, en una sola tabla.** "Sí" es la correspondencia directa, "Transf." es con
transformación, y "No" agrupa tanto la ausencia de campo como los pocos casos marcados como "sin
equivalente confirmado" en la tabla de su sistema (donde la investigación no llegó a comprobar si
existe o no); el matiz de cada celda, y por qué, está en la tabla detallada de la sección de ese
sistema.

| Campo de `biso` | Backlog.md | Linear | Trello | Beads | GitHub Issues | Taskwarrior | Task Master |
|---|---|---|---|---|---|---|---|
| `id` | Sí | Sí | Transf. | Transf. | Transf. | Transf. | Transf. |
| `createdAt` | Transf. | Sí | No | Sí | Sí | Transf. | No |
| `updatedAt` | Transf. | Sí | Transf. | Sí | Sí | Transf. | No |
| `archived` | Transf. | No | Sí | No | No | Transf. | No |
| `leaseExpiresAt` | No | No | No | No | No | No | No |
| `leaseHolder` | No | No | No | No | No | No | No |
| `title` | Sí | Sí | Sí | Sí | Sí | Sí | Sí |
| `status` | Sí | Sí | Transf. | Transf. | Transf. | Transf. | Transf. |
| `type` | Sí | No | No | Transf. | Transf. | No | No |
| `priority` | Sí | Transf. | Transf. | Transf. | No | Transf. | Transf. |
| `parent` | Sí | Sí | No | Transf. | Sí | No | Transf. |
| `assignees` | Sí | Transf. | Sí | No | Sí | No | No |
| `author` | No | Sí | No | No | Sí | No | No |
| `labels` | Sí | Sí | Sí | Sí | Sí | Sí | No |
| `dependencies` | Sí | Sí | No | Transf. | Sí | Sí | Sí |
| `references` | Sí | Transf. | Transf. | No | No | No | No |
| `documentation` | Sí | Transf. | Transf. | No | No | No | No |
| `modifiedFiles` | Sí | No | No | No | No | No | No |
| `due` | Sí | Sí | Sí | No | No | Sí | No |
| `ordinal` | Sí | Transf. | Transf. | No | No | No | No |
| `ext` | No | No | Transf. | Sí | No | Sí | No |
| `description` | Sí | Sí | Sí | Sí | Sí | No | Transf. |
| `plan` | Sí | No | No | Transf. | No | No | Transf. |
| `notes` | Sí | No | No | Sí | No | No | No |
| `summary` | Sí | No | No | No | No | No | No |
| `acceptanceCriteria` | Transf. | No | Sí | Transf. | No | No | Transf. |
| `comments` | Transf. | Transf. | Transf. | Transf. | Transf. | Transf. | No |
| `question` | No | No | No | No | No | No | No |

**Ninguno llega a la compatibilidad campo a campo que la portada de la especificación afirmaba sin
haberla comprobado.** Ni siquiera Backlog.md, el más cercano por diseño: pierde la clave estable de
los criterios y de los comentarios, que es precisamente la pieza que
["Los criterios y sus claves estables"](../spec/modelo-de-datos/criterios.md) y ["Los
comentarios"](../spec/modelo-de-datos/comentarios.md) tratan como una garantía central del modelo de
`biso`. Ningún sistema de los siete tiene nada parecido a `question` (una pregunta dirigida a una
persona concreta con su propio pendiente/respondida), ni a `leaseExpiresAt`/`leaseHolder` (un
arrendamiento con expiración), que son, junto con las claves estables, los tres rasgos que distinguen
el modelo de `biso` de los siete investigados.

## Backlog.md

Comprobado contra el CLI 1.52.0 instalado (`backlog task create --help`, `backlog task edit --help`,
y el frontmatter YAML real de tareas de prueba). Es, de los siete, el modelo más parecido: mismo
concepto de tarea versionada como fichero, mismos nombres para casi todo.

| Campo de `biso` | Backlog.md | Correspondencia |
|---|---|---|
| `id` | `id` (`TASK-<n>`, o `TASK-<n>.<m>` para una subtarea) | Directa |
| `createdAt` | `created_date` | Con transformación: precisión de minuto, no de segundo, y no se confirmó que sea UTC |
| `updatedAt` | `updated_date` | Con transformación: mismo caso |
| `archived` | mover el fichero a `archive/` con `backlog task archive` | Con transformación: es una ubicación de fichero, no un campo booleano |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `title` | Directa |
| `status` | `status`, enum configurable por proyecto | Directa |
| `type` | `type`, enum configurable por proyecto | Directa |
| `priority` | `priority`, enum configurable por proyecto | Directa |
| `parent` | `parent_task_id` | Directa (el `id` de la propia subtarea pasa a forma jerárquica `TASK-56.1`, pero el campo en sí no pierde el valor) |
| `assignees` | `assignee` (lista de `@nombre`) | Directa |
| `author` | — (solo hay `assignee`, no un campo de quien creó la tarea) | Sin equivalente |
| `labels` | `labels` | Directa |
| `dependencies` | `dependencies` | Directa |
| `references` | `references` | Directa |
| `documentation` | `documentation` | Directa |
| `modifiedFiles` | `modified_files` | Directa |
| `due` | `due_date` (`YYYY-MM-DD`) | Directa |
| `ordinal` | `ordinal` | Directa |
| `ext` | — | Sin equivalente |
| `description` | `description` | Directa |
| `plan` | `plan` | Directa |
| `notes` | `notes` | Directa |
| `summary` | `final_summary` | Directa |
| `acceptanceCriteria` | Acceptance Criteria, casilla numerada `#1`, `#2`... | Con transformación: la clave no es estable, ver nota |
| `comments` | Comments (`author` opcional, `created`, cuerpo) | Con transformación: sin clave propia, ver nota |
| `question` | — | Sin equivalente |

**Nota verificada sobre la clave de los criterios.** Se creó una tarea de prueba con los criterios
`#1 primero`, `#2 segundo` y `#3 tercero`, y se borró el `#2` con `--remove-ac 2`: el que era `#3`
pasó a numerarse `#2`. Backlog.md referencia por posición, no por una clave estable como la de
`biso` (["Los criterios y sus claves estables"](../spec/modelo-de-datos/criterios.md)): una tarea de
`biso` con los criterios `#1` y `#3` (sin `#2`, porque se borró) no tiene ese hueco en Backlog.md, y
un selector que apuntara al `#3` de `biso` puede dejar de significar lo mismo tras un borrado ajeno.

**Nota verificada sobre los comentarios.** El frontmatter de un comentario de Backlog.md es
`author` (opcional), `created` y el cuerpo; no lleva una clave propia, y no se encontró un flag de
edición o de borrado de un comentario individual (a diferencia de `--rm-comment` y
`--set-comment-date` de `biso`).

**Qué se pierde al exportar (`biso` → Backlog.md).** La clave estable de los criterios y de los
comentarios: si `biso` tiene los criterios `#1` y `#3`, Backlog.md los guarda como `#1` y `#2`,
seguidos, y ya no queda constancia del hueco. `author`, `ext` y `question` desaparecen del todo. Los
segundos de `createdAt`/`updatedAt` se redondean al minuto.

**Qué se pierde al importar (Backlog.md → `biso`).** Nada de estructura: como Backlog.md nunca tuvo
huecos en sus claves ni un campo de autor de tarea, `biso` simplemente asigna claves nuevas
consecutivas y deja `author` vacío, que son exactamente sus valores por defecto. La única pérdida
real en este sentido es la de precisión de fecha, que ya venía perdida del lado de Backlog.md.

**Cobertura: 23/28 ≈ 82 % en los dos sentidos.**

## Linear

Fuente: `docs/estado-del-arte/esquemas-de-datos-externos.md#5-linear-api-graphql`, extraído
directamente del `schema.graphql` público del SDK oficial.

| Campo de `biso` | Linear | Correspondencia |
|---|---|---|
| `id` | `identifier` (`ENG-123`) | Directa |
| `createdAt` | `createdAt` | Directa |
| `updatedAt` | `updatedAt` | Directa |
| `archived` | — (no verificado en esta investigación; el esquema consultado no incluyó un campo de archivado) | Sin equivalente confirmado |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `title` | Directa |
| `status` | `state` (`WorkflowState`: `name` personalizable + `type` de un conjunto fijo) | Directa: mismo mecanismo de enum configurable con categoría subyacente |
| `type` | — | Sin equivalente (simulable con `labels`, que no es lo mismo) |
| `priority` | `priority` (`Float`, con `priorityLabel`) | Con transformación: escala numérica fija, no enum de texto configurable |
| `parent` | `parent` / `children` | Directa |
| `assignees` | `assignee` (un único valor) | Con transformación: Linear solo admite una persona asignada |
| `author` | `creator` | Directa |
| `labels` | `labels` | Directa |
| `dependencies` | `IssueRelation` con `type: blocks` | Directa |
| `references` | `attachments` (no detallado en la investigación) | Con transformación |
| `documentation` | `attachments`, mismo mecanismo | Con transformación |
| `modifiedFiles` | — | Sin equivalente |
| `due` | `dueDate` | Directa |
| `ordinal` | `subIssueSortOrder` (entre hermanos) o `sortOrder` (mencionado, no detallado) | Con transformación: orden relativo, no un entero libre |
| `ext` | — | Sin equivalente |
| `description` | `description` | Directa |
| `plan` | — | Sin equivalente |
| `notes` | — | Sin equivalente |
| `summary` | — | Sin equivalente |
| `acceptanceCriteria` | — (la descomposición se hace con sub-issues completos, no con una lista ligera) | Sin equivalente |
| `comments` | `Comment` (`user`, `createdAt`, `body`, `editedAt`, hilos, resuelto/no resuelto) | Con transformación: son editables, con hilos y resolución que `biso` no modela |
| `question` | — | Sin equivalente |

**Qué se pierde al exportar (`biso` → Linear).** Si una tarea tiene más de una persona en
`assignees`, solo la primera sobrevive. El resto de campos sin equivalente desaparece del todo:
`type`, `leaseExpiresAt`/`leaseHolder`, `ext`, `plan`, `notes`, `summary`,
`acceptanceCriteria`, `question`. Un comentario exportado deja de ser inmutable:
Linear permite editarlo después, y `biso` ya no lo sabría.

**Qué se pierde al importar (Linear → `biso`).** Nada adicional a lo anterior: un solo `assignee`
cabe siempre en la lista de `biso`, así que no hay pérdida en ese sentido, y todo lo que Linear no
modela tampoco llega para perderse. Sí se pierde la información propia de Linear que no tiene sitio
en `biso`: los hilos de comentarios, el estado resuelto/no resuelto, el `cycle` y el `project` a los
que pertenece el issue.

**Cobertura: 17/28 ≈ 61 % en los dos sentidos.**

## Trello

Fuente: `docs/estado-del-arte/esquemas-de-datos-externos.md#6-trello-api-rest`, del esquema OpenAPI
oficial.

| Campo de `biso` | Trello | Correspondencia |
|---|---|---|
| `id` | `id` (`TrelloID`) | Con transformación: formato distinto, mismo rol |
| `createdAt` | — (no se encontró un campo de fecha de creación de la tarjeta) | Sin equivalente confirmado |
| `updatedAt` | `dateLastActivity` | Con transformación: incluye cualquier actividad, no solo cambios de campos |
| `archived` | `closed` (booleano) | Directa |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `name` | Directa |
| `status` | `idList` (la lista es el estado) | Con transformación: es pertenencia a una lista, no un valor de enum |
| `type` | — | Sin equivalente |
| `priority` | Custom Field, si el tablero lo declaró | Con transformación |
| `parent` | — (no existe jerarquía nativa en el objeto Card) | Sin equivalente |
| `assignees` | `idMembers` | Directa |
| `author` | — (no se documentó un creador de la tarjeta) | Sin equivalente |
| `labels` | `labels` / `idLabels` | Directa |
| `dependencies` | — (el Power-Up de dependencias no está en la API REST pública) | Sin equivalente |
| `references` | `Attachment` (`url`, `name`...) | Con transformación |
| `documentation` | `Attachment`, mismo mecanismo | Con transformación |
| `modifiedFiles` | — | Sin equivalente |
| `due` | `due` | Directa |
| `ordinal` | `pos` (flotante, orden dentro de la lista) | Con transformación |
| `ext` | Custom Fields (declarados de antemano, tipados) | Con transformación |
| `description` | `desc` (hasta 16.384 caracteres) | Directa |
| `plan` | — | Sin equivalente |
| `notes` | — | Sin equivalente |
| `summary` | — | Sin equivalente |
| `acceptanceCriteria` | `Checklist` + `CheckItem` (`id`, `name`, `state: complete\|incomplete`) | Directa |
| `comments` | `Action` de `type: commentCard` (`idMemberCreator`, `date`, `data.text`) | Con transformación: editables y borrables por endpoints propios |
| `question` | — | Sin equivalente |

**Qué se pierde al exportar (`biso` → Trello).** `type`, `parent`, `author`, `dependencies`,
`modifiedFiles`, `plan`, `notes`, `summary` y `question` no tienen dónde ir. Un comentario exportado
deja de ser inmutable, igual que en Linear y GitHub. `priority` y `ext` solo sobreviven si el tablero
de destino ya declaró los Custom Fields correspondientes; si no, hay que crearlos antes de exportar.

**Qué se pierde al importar (Trello → `biso`).** Nada adicional: `dependencies` y `parent` no
existen en el origen y por tanto no hay nada que perder en ese sentido. Se pierde, en cambio, todo lo
propio de Trello sin sitio en `biso`: `pos` como orden fino entre tarjetas de una misma lista, los
Custom Fields no declarados de antemano en un tablero de `biso` (que no tiene ese concepto), y el
board/list como jerarquía de agrupación.

**Cobertura: 16/28 ≈ 57 % en los dos sentidos.**

## Beads (`bd`)

Fuente: `docs/estado-del-arte/esquemas-de-datos-externos.md#2-beads-bd`.

| Campo de `biso` | Beads | Correspondencia |
|---|---|---|
| `id` | `id` (`bd-a1b2`, hash determinista) | Con transformación: formato distinto, mismo rol |
| `createdAt` | `created_at` | Directa |
| `updatedAt` | `updated_at` | Directa |
| `archived` | — (no se documentó un estado distinto de `status`) | Sin equivalente |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `title` | Directa |
| `status` | `status` (cadena, p. ej. `open`, `in_progress`) | Con transformación |
| `type` | `type` (`bug`, `feature`, `task`, `epic`, `chore`; cerrado) | Con transformación: vocabulario fijo, no configurable |
| `priority` | `priority` (entero 0-4) | Con transformación |
| `parent` | arista `type: parent-child` (misma tabla que `dependencies`) | Con transformación |
| `assignees` | — | Sin equivalente |
| `author` | — | Sin equivalente |
| `labels` | `labels` | Directa |
| `dependencies` | arista `type: blocks` (por defecto) | Con transformación: mismo mecanismo que `parent`, distinguido solo por el valor de `type` |
| `references` | — | Sin equivalente |
| `documentation` | — | Sin equivalente |
| `modifiedFiles` | — | Sin equivalente |
| `due` | — | Sin equivalente |
| `ordinal` | — | Sin equivalente |
| `ext` | `metadata` (JSON arbitrario) | Directa |
| `description` | `description` | Directa |
| `plan` | `design` | Con transformación: nombre más específico ("notas de diseño") que el "plan" genérico de `biso` |
| `notes` | `notes` | Directa |
| `summary` | — | Sin equivalente |
| `acceptanceCriteria` | `acceptance_criteria` (texto libre, no una lista marcable) | Con transformación: se pierde la clave y el marcado por ítem |
| `comments` | `bd comment`/`bd comments` (autor opcional, fecha, cuerpo) | Con transformación: no se confirmó ni edición ni borrado, ni una clave estable por comentario |
| `question` | — (`waiters` es una lista interna de qué espera a qué, no una pregunta dirigida a alguien) | Sin equivalente |

**Qué se pierde al exportar (`biso` → Beads).** `assignees`, `author`, `references`,
`documentation`, `modifiedFiles`, `due`, `ordinal`, `summary` y `question` desaparecen.
`acceptanceCriteria` cae en el único campo de texto libre de Beads, perdiendo el marcado individual
de cada criterio.

**Qué se pierde al importar (Beads → `biso`).** Nada adicional: Beads no define ninguno de esos
campos, así que no hay nada que traer y perder. Sí se pierde lo propio de Beads sin sitio en `biso`:
el `payload` JSON asociado al issue, la lista `waiters`, y los tipos de relación que no son de
bloqueo ni de jerarquía (`related`, `tracks`, `discovered-from`, `caused-by`, `validates`,
`supersedes`).

**Cobertura: 16/28 ≈ 57 % en los dos sentidos.**

## GitHub Issues

Fuente: `docs/estado-del-arte/esquemas-de-datos-externos.md#4-github-issues-api-rest`.

| Campo de `biso` | GitHub Issues | Correspondencia |
|---|---|---|
| `id` | `number` | Con transformación: numérico, sin el prefijo de proyecto de `biso` |
| `createdAt` | `created_at` | Directa |
| `updatedAt` | `updated_at` | Directa |
| `archived` | — | Sin equivalente |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `title` | Directa |
| `status` | `state` (`open`/`closed`, binario) + `state_reason` | Con transformación: pérdida de granularidad, ver nota |
| `type` | no confirmado en esta investigación (no se investigaron los "Issue Types" de organización, solo `labels`) | Con transformación, vía `labels` |
| `priority` | — (vive en Projects v2, un recurso distinto del Issue) | Sin equivalente en el objeto Issue |
| `parent` | API de sub-issues (`parent`/`sub_issues`) | Directa |
| `assignees` | `assignees` | Directa |
| `author` | `user` | Directa |
| `labels` | `labels` | Directa |
| `dependencies` | API de issue-dependencies (`blocked_by`/`blocking`) | Directa |
| `references` | — (solo menciones de texto en el `body`, no un campo estructurado) | Sin equivalente |
| `documentation` | — | Sin equivalente |
| `modifiedFiles` | — (es propio del objeto Pull Request, no del Issue) | Sin equivalente |
| `due` | — (solo existe `due_on` en el Milestone compartido, no por Issue) | Sin equivalente |
| `ordinal` | — (existe orden dentro de Projects v2, no en el Issue) | Sin equivalente |
| `ext` | — (los campos personalizados de Projects v2 viven en el objeto "item" del proyecto, no en el Issue) | Sin equivalente |
| `description` | `body` | Directa |
| `plan` | — | Sin equivalente |
| `notes` | — | Sin equivalente |
| `summary` | — | Sin equivalente |
| `acceptanceCriteria` | — (una lista de tareas en Markdown dentro del `body` es texto, no un campo con estructura) | Sin equivalente estructurado |
| `comments` | `id`, `body`, `user`, `created_at`/`updated_at` | Con transformación: son editables mediante `PATCH` |
| `question` | — | Sin equivalente |

**Nota sobre `status`.** GitHub solo admite los valores `open` y `closed`; el vocabulario configurable
de `biso` (por ejemplo `to do`/`in progress`/`blocked`/`done`) se colapsa a esos dos al exportar, y
`state_reason` (`completed`/`not_planned`) es lo más cerca que hay de matizar un cierre.

**Qué se pierde al exportar (`biso` → GitHub Issues).** La granularidad de `status` (varios estados
activos se ven todos como `open`). Todo lo que no tiene equivalente: `priority`, `references`,
`documentation`, `modifiedFiles`, `due` por tarea, `ordinal`, `ext`, `plan`, `notes`, `summary`,
`acceptanceCriteria`, `question`. Un comentario exportado deja de ser inmutable.

**Qué se pierde al importar (GitHub Issues → `biso`).** Nada adicional a lo de arriba: `open`/`closed`
se mapea sin pérdida a dos valores del vocabulario de `biso`. Se pierde lo propio de GitHub sin sitio
en `biso`: `milestone`, `locked`/`active_lock_reason`, y el resumen `issue_dependencies_summary`.

**Cobertura: 13/28 ≈ 46 % en los dos sentidos.**

## Taskwarrior

Fuente: `docs/estado-del-arte/esquemas-de-datos-externos.md#1-taskwarrior`.

| Campo de `biso` | Taskwarrior | Correspondencia |
|---|---|---|
| `id` | `uuid` | Con transformación |
| `createdAt` | `entry` | Con transformación: timestamp UNIX, no fecha ISO |
| `updatedAt` | `modified` | Con transformación |
| `archived` | `status: D` (borrada) | Con transformación, con pérdida de matiz: borrado y archivado no son lo mismo |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `description` | Directa (Taskwarrior llama "description" al título de una línea) |
| `status` | `status` (`P`/`C`/`D`/`R`, cerrado) | Con transformación |
| `type` | — | Sin equivalente |
| `priority` | `priority` (UDA desde 2.4.3; `H`/`M`/`L`/vacío) | Con transformación |
| `parent` | — (el `parent` real de Taskwarrior es solo el de las plantillas de recurrencia, semántica distinta) | Sin equivalente |
| `assignees` | — (Taskwarrior es de un único usuario local) | Sin equivalente |
| `author` | — | Sin equivalente |
| `labels` | `tag_<tag>` | Directa |
| `dependencies` | `dep_<uuid>` | Directa |
| `references` | — | Sin equivalente |
| `documentation` | — | Sin equivalente |
| `modifiedFiles` | — | Sin equivalente |
| `due` | `due` | Directa |
| `ordinal` | — | Sin equivalente |
| `ext` | UDA (user defined attributes) | Directa |
| `description` (texto largo) | — (el único campo de texto de Taskwarrior ya es el título) | Sin equivalente |
| `plan` | — | Sin equivalente |
| `notes` | — | Sin equivalente |
| `summary` | — | Sin equivalente |
| `acceptanceCriteria` | — | Sin equivalente |
| `comments` | `annotation_<timestamp>` (texto + fecha, sin autor) | Con transformación, con pérdida del autor |
| `question` | — | Sin equivalente |

**Por qué no se cuentan más campos "con transformación" vía UDA.** Una UDA es de un único valor
escalar (`string`, `numeric`, `date`, `duration` o `uuid`); no admite una lista, así que
`assignees`/`references`/`documentation`/`modifiedFiles` no caben sin aplanarlos a texto separado por
comas, perdiendo la propiedad de lista real que Taskwarrior reconoce. Y una lista de objetos con
clave propia (`acceptanceCriteria`, `comments` con estructura completa, `question`) no cabe en absoluto en un valor escalar sin volverse texto opaco que ningún comando de
Taskwarrior sabría interpretar como checklist.

**Qué se pierde al exportar (`biso` → Taskwarrior).** Todo lo que no tiene equivalente de la tabla:
en particular, la distinción entre archivar y borrar, y toda estructura de listas (`assignees`,
`acceptanceCriteria`, etc.). Las anotaciones que reciben los comentarios pierden el autor.

**Qué se pierde al importar (Taskwarrior → `biso`).** Nada adicional: Taskwarrior no define esos
campos. Se pierde lo propio de Taskwarrior sin sitio en `biso`: `wait` (ocultar una tarea hasta una
fecha), `recur`/`mask` (la recurrencia entera, que `biso` no modela), y cualquier UDA que el usuario
haya declarado con su propio significado.

**Cobertura: 12/28 ≈ 43 % en los dos sentidos.**

## Task Master (`claude-task-master`)

Fuente: `docs/estado-del-arte/esquemas-de-datos-externos.md#3-task-master-claude-task-master`.

| Campo de `biso` | Task Master | Correspondencia |
|---|---|---|
| `id` | `id` (número, único dentro del tag) | Con transformación |
| `createdAt` | — | Sin equivalente |
| `updatedAt` | — | Sin equivalente |
| `archived` | — | Sin equivalente |
| `leaseExpiresAt` | — | Sin equivalente |
| `leaseHolder` | — | Sin equivalente |
| `title` | `title` | Directa |
| `status` | `status` (`pending`/`in-progress`/`done`/`review`/`deferred`/`cancelled`, cerrado) | Con transformación |
| `type` | — | Sin equivalente |
| `priority` | `priority` (`high`/`medium`/`low`, cerrado) | Con transformación |
| `parent` | anidamiento físico (`subtasks` dentro de la tarea), no un campo de referencia | Con transformación |
| `assignees` | — | Sin equivalente |
| `author` | — | Sin equivalente |
| `labels` | — | Sin equivalente |
| `dependencies` | `dependencies` (array de ids numéricos) | Directa |
| `references` | — | Sin equivalente |
| `documentation` | — | Sin equivalente |
| `modifiedFiles` | — | Sin equivalente |
| `due` | — | Sin equivalente |
| `ordinal` | — (el orden es la posición en el array, sin campo dedicado) | Sin equivalente |
| `ext` | — | Sin equivalente |
| `description` | `description` (resumen corto, no un bloque de prosa libre) | Con transformación |
| `plan` | `details` | Con transformación: nombre distinto, mismo papel |
| `notes` | — | Sin equivalente |
| `summary` | — | Sin equivalente |
| `acceptanceCriteria` | `testStrategy` (texto libre sobre cómo verificar, no una lista marcable) | Con transformación, con pérdida de estructura |
| `comments` | — | Sin equivalente |
| `question` | — | Sin equivalente |

**Qué se pierde al exportar (`biso` → Task Master).** Las fechas (`createdAt`, `updatedAt`, `due`),
todo lo relativo a personas (`assignees`, `author`), `labels`, `references`, `documentation`,
`modifiedFiles`, `ext`, `notes`, `summary`, `comments` y `question`.
`acceptanceCriteria` se funde en `testStrategy` como texto, perdiendo el marcado por ítem.

**Qué se pierde al importar (Task Master → `biso`).** Nada adicional: ninguno de esos campos existe
en origen. Se pierde lo propio de Task Master sin sitio en `biso`: el informe aparte de complejidad
(`task-complexity-report.json`), y el nivel de "tags" que agrupa varios `tasks.json` en uno.

**Cobertura: 9/28 ≈ 32 % en los dos sentidos.**

## Qué quedó sin verificar

Todo lo que ya declara sin verificar
[`esquemas-de-datos-externos.md`](esquemas-de-datos-externos.md#qué-quedó-sin-verificar) sigue sin
verificar aquí, porque esta página no repite esa investigación. Además, específicamente para la
comparación de esta página:

1. **Si Linear tiene un campo de archivado** (`archivedAt` o similar): el fragmento del esquema
   GraphQL consultado no lo cubrió, y no se afirma su existencia ni su ausencia con certeza.
2. **Si GitHub tiene un campo `type` de dominio nativo** (los "Issue Types" a nivel de organización,
   añadidos después del período que cubrió la investigación de origen): se trató como no verificado y
   se usó `labels` como la vía práctica conocida, no como constancia de que no exista una vía mejor.
3. **Si un `Checklist` de Trello admite más de una instancia por tarjeta sin límite**: no se comprobó
   un límite explícito en la documentación consultada, solo que el recurso lo permite en general. La
   pregunta sostenía la fila de `definitionOfDone` como un segundo checklist y hoy no decide nada,
   porque ese campo ya no existe en `biso`; se conserva porque es un hecho sobre Trello que otra
   correspondencia futura podría necesitar.
