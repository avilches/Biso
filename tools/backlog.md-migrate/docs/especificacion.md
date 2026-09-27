# Especificación

> **Borrador provisional.** Ninguna regla de esta página está aprobada: son propuestas pendientes de
> revisión. TASK-73 ya se resolvió (`biso` no adopta identificadores de subtarea con punto), así que
> lo que queda pendiente es una decisión propia de este proyecto y el resto de lo que ya recogía
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
- **Un fichero de `tasks/`, `completed/` o `archive/tasks/` cuyo frontmatter no se puede interpretar
  como YAML válido (una comilla sin cerrar, una clave duplicada) se salta**, con un hallazgo que trae
  el fichero y el error de parseo, y el resto del lote sigue. Backlog.md 1.53.0 descarta esa misma
  tarea en silencio (no sale en `backlog task list` y ni siquiera `backlog task view` la encuentra),
  pero el convertidor lee estas carpetas directamente con su propio analizador de YAML, así que sí se
  la encuentra. No es el mismo tipo de hallazgo que una clave o un vocabulario no reconocido dentro de
  una tarea que sí se pudo leer (ver "El mapeo de campos"): este fichero ni siquiera llegó a
  convertirse en una tarea.
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
  se usan su `id`, su `title`, su `createdAt` y su `ordinal`. Ver "Orden manual" para qué hace con este
  último.

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
| `milestone` | una etiqueta en `labels` | `milestone::<slug>`. Ver "Milestone y proyecto" |
| `project` | una etiqueta en `labels` | `project::<slug>`, con el mismo slug que el milestone |
| `dependencies` | `dependencies` | Con los ids reescritos |
| `parent_task_id` | `parent` | Con el id reescrito |
| `ordinal` | `ordinal` | Se recalcula, no se copia. Ver "Orden manual" |
| `due_date` | `due` | `YYYY-MM-DD`, tal cual |
| `documentation` | `documentation` | Clave de compatibilidad de `biso new --from`, que la funde en `references`. Ver "Documentación y ficheros tocados" |
| `references` | `references` | Lista tal cual |
| `modified_files` | `modifiedFiles` | Clave de compatibilidad de `biso new --from`, que la funde en `references`. Ver "Documentación y ficheros tocados" |
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
   origen (código 3). El prefijo del destino sale de `task_prefix` del destino. Dos ids de origen cuentan
   como el mismo id, a los efectos de ese mismo error de código 3 ("dos ficheros traen el mismo id", ver
   "Códigos de salida"), cuando su prefijo coincide sin distinguir mayúsculas, cuando los dos son simples
   o los dos son de subtarea, cuando su número principal es igual y, si los dos son de subtarea, también
   lo es su número de subtarea, todo interpretado como número entero y no como cadena: es la misma forma
   canónica que la regla 2 usa para decidir si un número choca con el destino. Un id simple y un id de
   subtarea con el mismo número principal nunca son el mismo id a estos efectos, precisamente porque uno
   tiene número de subtarea y el otro no: `TASK-1` y `TASK-1.2` son dos ids distintos, la tarea y una de
   sus subtareas.
2. **Un id simple de origen conserva su número con el prefijo del destino** (`TASK-70` pasa a
   `BISO-70`, `XYZ-001` pasa a `BISO-1`) siempre que ese número no exista ya en el destino. Los ceros a
   la izquierda no se escriben.
3. **Una tarea que ya está en el destino no se vuelve a importar.** Es la misma tarea de una
   importación anterior si el destino tiene una con el mismo `title` y el mismo `createdAt`, sea cual
   sea su id. El `title` que se compara no es el crudo del fichero de origen: es ese mismo título con
   sus menciones de id de origen reescritas de forma ingenua, sustituyendo por el prefijo del destino y
   el mismo número mencionado solo la mención (misma forma y límites de palabra que la regla 5) cuyo id
   corresponde exactamente a una tarea SIMPLE del lote de origen actual, sin tener en cuenta si ese
   número choca con algo en el destino ni si esa tarea se va a reasignar; cualquier otra mención con
   forma de id, sea porque no corresponde a ninguna tarea del lote o porque corresponde a una subtarea,
   se deja tal cual, con el mismo criterio con el que la regla 6 deja tal cual una mención que no
   reconoce. El porqué de esta reescritura, y el caso que aun así se pierde, están en decisiones.md,
   ["Los identificadores conservan su número y cambian de
   prefijo"](decisiones.md#los-identificadores-conservan-su-número-y-cambian-de-prefijo). Ese título
   reescrito de forma ingenua se usa solo para esta comparación y se descarta después: el título que de
   verdad se escribe en el destino, cuando la tarea no es un duplicado, sale siempre de la reescritura
   definitiva de la regla 5. No se escribe su línea, su id de origen se equipara al que ya tiene en el
   destino, y es un hallazgo (`already on the destination, skipped`). Así ejecutar `import` dos veces no
   duplica el tablero, subtareas incluidas. No actualiza la tarea que ya estaba, y una tarea que se haya
   editado en el destino después de importarla (otro título) ya no se reconoce.
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
   podría ser el nombre de una rama, no se toca, y tampoco `SUBTASK-12`. Una vez que este patrón
   reconoce una mención válida, a qué tarea del lote corresponde se decide por la misma forma canónica
   que la regla 1 usa para "mismo id" (mismo prefijo, mismo número principal como entero y, si la
   mención tiene forma de subtarea, mismo número de subtarea como entero), no por la cadena exacta: una
   mención `TASK-1` encuentra en la tabla de equivalencias tanto a una tarea de origen `TASK-1` como a
   una `TASK-001`. Se reescribe en cualquier lugar del texto, un bloque de código incluido. No se toca
   `documentation`, `references` ni `modified_files`, que son rutas y URLs opacas.
6. Una mención que tiene la forma de un id de origen pero no corresponde a ninguna tarea del origen, o
   que solo difiere en las mayúsculas (`Xyz-002`, `task-12`), se deja como está y es un hallazgo,
   uno por fichero y campo con el recuento, para que quien ejecuta pueda revisarla. Si corresponde a
   alguna tarea del origen se decide con esa misma forma canónica de la regla 1, no por cadena exacta:
   una mención `TASK-1` sí corresponde a una tarea de origen `TASK-001`, y no cae en este hallazgo solo
   porque los ceros no coincidan letra por letra.
7. Un `parent` o una dependencia que nombra un id que no está entre las tareas del origen se quita
   de la línea y es un hallazgo, porque `biso` rechaza el lote entero (código 4) por un solo
   identificador que no existe. "Estar entre las tareas del origen" se decide aquí por la forma
   canónica completa de la regla 1, prefijo sin distinguir mayúsculas incluido, a diferencia del patrón
   de mención de la regla 5, que sí distingue mayúsculas para no tocar por accidente un nombre de rama
   en texto libre: un valor estructurado como `parent_task_id` o un elemento de `dependencies` no tiene
   ese riesgo, porque todo su valor es un id, no texto libre donde un id puede aparecer por casualidad.
   Así, un `parent_task_id: TASK-1` que en realidad apunta a la tarea de origen `TASK-001` se reconoce
   igual. Un valor que no tiene siquiera la forma de un id de la regla 1 se quita igual, con su hallazgo.
   Si dos elementos de `dependencies` de la misma tarea resuelven al mismo id final por esta forma
   canónica (`TASK-1` y `TASK-001` en la misma lista, por ejemplo), se deja uno solo, el primero en el
   orden original, con el mismo criterio de deduplicación que ya aplica "Alfabeto de un token" cuando dos
   valores de `labels` o `assignees` de la misma tarea quedan iguales tras la conversión. Los ciclos de
   padres o de dependencias no se detectan: los rechaza `biso new --from --dry-run`, y por eso el ensayo
   forma parte del uso.
8. Cada id reasignado es un hallazgo: `XYZ-001.01: id BISO-97 assigned (subtask ids have no equivalent)`
   o `TASK-12: id BISO-12 is taken on the destination, reassigned to BISO-97`.
9. **El id de origen de una subtarea se guarda además en la etiqueta con ámbito
   `backlog.id::<id-de-origen>`** (`backlog.id::TASK-56.1`), con el id completo tal como lo escribe
   Backlog.md, punto incluido, para que `export` lo pueda reconstruir; el porqué está en decisiones.md,
   ["El identificador de origen de una subtarea se guarda en una etiqueta con
   ámbito"](decisiones.md#el-identificador-de-origen-de-una-subtarea-se-guarda-en-una-etiqueta-con-ámbito).
   Si la tarea de origen ya trae en sus propias `labels` un valor con la clave `backlog.id`, prevalece la
   etiqueta que el convertidor deriva del id de origen real, y la etiqueta de origen que choca se quita
   de la lista antes de escribir el lote, con el mismo hallazgo que "Milestone y proyecto" describe para
   `milestone::` y `project::`.

### Milestone y proyecto

El `milestone` de una tarea es un id de milestone (`m-4`). El título se lee del fichero
`milestones/m-4 - *.md` o del de `archive/milestones/`. El slug es el título en minúsculas, sin
diacríticos, con cada tramo de caracteres que no sean letras ni dígitos Unicode convertido en un solo
`-` y sin `-` en los extremos: `Puesta en uso` da `puesta-en-uso` e `Implementación` da
`implementacion`. La tarea recibe la etiqueta con ámbito `milestone::puesta-en-uso`, después de las
etiquetas que ya tuviera. El separador es `::` y no `:` porque una tarea de Backlog.md pertenece como
mucho a un milestone, y `::` es la forma con la que una etiqueta con ámbito deja como mucho un valor
de su clave por tarea (["Las etiquetas con ámbito"](../../../docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito)):
`biso` mismo hace cumplir esa exclusividad al escribir, en vez de que el convertidor confíe en no
escribirla nunca dos veces. Una tarea sin milestone no recibe ninguna etiqueta y eso no es un hallazgo.
Usa el id como slug (`milestone::m-9`), con un hallazgo, cuando el fichero del milestone no existe o
cuando el slug sale vacío. Dos milestones distintos que dan el mismo slug se fusionarían en una sola
etiqueta, así que es un hallazgo.

El campo `project` de una tarea es un texto que ya viene escrito (`alpha`), sin fichero aparte. Recibe
la etiqueta con ámbito `project::<slug>`, con el mismo cálculo del slug, la misma razón para el `::` (una
tarea tiene como mucho un `project`) y después de la del milestone.

**Si la tarea de origen ya trae en sus propias `labels` un valor con la clave `milestone` o la clave
`project`**, prevalece la etiqueta que el convertidor deriva del dato real de Backlog.md, no la que ya
traía el origen. Escribir las dos dejaría dos valores de la misma clave con ámbito en la misma tarea, y
`biso` rechaza eso al escribir (["Escribir una etiqueta con
ámbito"](../../../docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito)). La etiqueta de origen
que choca se quita de `labels` antes de escribir el lote, y se informa con el mismo tratamiento que un
carácter fuera del alfabeto de un token: una línea de hallazgo con el fichero, la tarea y el valor
descartado. La misma regla se aplica igual a la etiqueta `backlog.id::`, ver "Identificadores". El
razonamiento está en decisiones.md, ["Una etiqueta con ámbito derivada del dato real gana a la etiqueta de
origen que choca con
ella"](decisiones.md#una-etiqueta-con-ámbito-derivada-del-dato-real-gana-a-la-etiqueta-de-origen-que-choca-con-ella).

**Filtrar no cambia.** `--label milestone:` sigue encontrando cualquier etiqueta de esa clave sea cual
sea su separador, y `--label milestone:puesta-en-uso` sigue encontrando `milestone::puesta-en-uso`,
porque el separador no cuenta al consultar ni al quitar, solo al escribir
(["Las etiquetas con ámbito"](../../../docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito)).

### Orden manual

Backlog.md guarda `ordinal` como un número; `biso` lo guarda como una clave de texto en base 36 que no
puede terminar en `0` y que no tiene relación aritmética con el número de origen
(["El orden manual y su clave"](../../../docs/spec/modelo-de-datos/orden-manual.md)). **No es un mapeo
directo.** Copiar el número tal cual fallaría además por la propia forma: un `ordinal` de Backlog.md
como `1000` termina en `0`, que es precisamente la forma que `biso` rechaza.

El convertidor lee también el `ordinal` de las tareas que ya existen en el destino (ver "Qué le pregunta
al destino") y toma la mayor de esas claves como ancla, o ninguna si el destino no tiene todavía ninguna
clave de orden manual. Después ordena las tareas del origen por su `ordinal` ascendente (las que no lo
tienen quedan sin clave, igual que en `biso`) y les asigna claves nuevas que conservan ese mismo orden
relativo, con el mismo algoritmo del punto medio que usan `--ordinal last`, `--above` y `--below`
(["El algoritmo del punto medio"](../../../docs/spec/modelo-de-datos/orden-manual.md#el-algoritmo-del-punto-medio)):
la primera tarea importada toma la clave intermedia entre el ancla del destino y la ausencia de
siguiente, y cada tarea siguiente toma la clave intermedia entre la que se acaba de asignar y la
ausencia de siguiente, como si el lote entero se colocara con `--ordinal last`, una tarea detrás de otra,
después de todo lo que ya hubiera en el destino. Un empate en el `ordinal` de origen se desempata
ordenando esas tareas entre sí por su id de origen ascendente, con el mismo orden natural que usa
"Identificadores", regla 4, para asignar número nuevo a los ids que chocan, para que dos ejecuciones den
el mismo resultado. Las claves calculadas no chocan nunca con
una clave ya existente en el destino, porque cada una es estrictamente mayor que el ancla y que la clave
recién asignada antes que ella, sea cual sea el estado del destino.

### Fechas

Backlog.md guarda `YYYY-MM-DD HH:mm`, en UTC. Se convierte a `YYYY-MM-DDTHH:mm:00Z`. Una fecha sin hora
(`YYYY-MM-DD`) pasa a `YYYY-MM-DDT00:00:00Z`. Si una tarea no tiene `updated_date`, `updatedAt` es su
`createdAt`. Una fecha que no encaja en ninguna de las dos formas es un hallazgo, y el campo se omite.

### Documentación y ficheros tocados

`biso` no tiene un campo `documentation` ni un campo `modifiedFiles`: los dos se retiraron y `references`
quedó como el único campo de punteros de una tarea
(["Las relaciones entre tareas"](../../../docs/spec/modelo-de-datos/relaciones.md#los-punteros-references)).
La única huella que queda de los dos es de entrada: `biso new --from` sigue aceptando las claves
`documentation` y `modifiedFiles` de un lote ajeno, y funde cada una en `references`, avisando con
`imported_documentation_merged` e `imported_modified_files_merged` (["`biso new`"](../../../docs/spec/cmd/new.md)).

El convertidor aprovecha exactamente esa compatibilidad: escribe `documentation` y `modified_files` de
Backlog.md en las claves `documentation` y `modifiedFiles` del lote, tal cual, y deja que `biso new --from`
haga la fusión y emita sus propios avisos. No hace falta que el convertidor las funda él mismo en
`references`.

**La pega, aceptada como tal.** `biso` no conserva de qué lista venía cada puntero: al volver a exportar
una tarea importada, sus tres listas de origen (`references`, `documentation`, `modified_files`)
llegarían todas juntas a un solo campo, y `export` (TASK-7) no tiene forma de repartirlas de vuelta. Es
la misma pérdida que ya acepta la propia decisión de `biso` de retirar los dos campos
(["Se retira `documentation` y `references` queda como único campo de
punteros"](../../../docs/decisiones/detalles.md#se-retira-documentation-y-references-queda-como-único-campo-de-punteros)),
que además apunta la salida de quien la quiera evitar: guardar en una etiqueta con ámbito de qué campo
venía cada valor. Este proyecto no lo hace por el mismo motivo que llevó a `biso` a fusionar los campos:
ninguna tarea real usa las dos listas a la vez, así que la distinción no sostiene el coste de una
etiqueta por cada elemento de `references`.

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
3. Si tras convertir dos valores de la misma tarea quedan iguales (`a b` y `a-b`), se deja uno solo, el
   primero en el orden original.

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
Un frontmatter que no se pudo interpretar como YAML usa `frontmatter` como `<campo>`, porque el fallo
es de todo el fichero y no de una clave concreta.

### Códigos de salida

| Código | Cuándo |
|---|---|
| 0 | Convertido y sin ningún hallazgo |
| 2 | Uso incorrecto: falta un argumento o hay uno desconocido |
| 3 | El origen no se puede leer: no existe la carpeta, no tiene `tasks/`, un id tiene una forma que no es la de "Identificadores", los ids no comparten prefijo o dos ficheros traen el mismo id |
| 4 | El destino no responde: no se encuentra `biso`, o falla una de sus órdenes |
| 5 | Convertido con hallazgos. El NDJSON está escrito, salvo con `--strict`, donde no se escribe nada |

**Corrección.** Esta página decía antes que un frontmatter que no se puede interpretar como YAML era
uno de los motivos del código 3, que aborta el lote entero. Ya no es así: ver "Qué lee del origen" más
arriba. Un fichero así se salta como un hallazgo (código 5), no aborta el import por el resto de
ficheros que sí se pudieron leer.

**El código 5 no es un fallo**: el NDJSON está escrito. Encadenar `import` con `biso new --from` en un
`&&` se detiene en el 5, así que quien automatice el paso debe aceptarlo explícitamente.

**El convertidor no escribe nunca en el tablero destino.** Un fallo a mitad de la conversión deja el
destino intacto y, con `--out` a un fichero, no deja un fichero a medias: se escribe en un temporal y
se renombra al terminar.
