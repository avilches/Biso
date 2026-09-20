# Especificación

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
| `<backlog-dir>` | sí | | La carpeta `backlog/` de un proyecto de Backlog.md, la que contiene `config.yml` y `tasks/`. Puede ser un enlace simbólico |
| `--project <dir>` | sí | | Directorio desde el que `biso` resuelve el tablero destino. Se le pasa a `biso` como `--cwd` |
| `--out <file\|->` | no | `-` (salida estándar) | Dónde se escribe el NDJSON |
| `--biso <path>` | no | `biso` del `PATH` | El binario de `biso` que se ejecuta |
| `--strict` | no | falso | Si hay algún hallazgo, no escribe nada y sale con el código 5 |

### Qué lee del origen

- `tasks/` y `completed/`: las tareas, en el mismo formato. Cada fichero `.md` es una tarea.
- `archive/tasks/`: las tareas archivadas. Se importan con `archived: true`.
- `milestones/` y `archive/milestones/`: solo para leer el título de cada milestone (el campo `title`
  de su frontmatter).
- `config.yml`: solo la clave `task_prefix`, que es el prefijo de los ids del origen. Sin ella, el
  prefijo es `task`. La comparación con el prefijo de un id ignora las mayúsculas.
- `drafts/` y `archive/drafts/` **no se leen**. Si tienen ficheros, sale un hallazgo con cuántos.

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
| `labels` (lista) | `labels` | Igual, más la etiqueta del milestone |
| `milestone` | una etiqueta en `labels` | `milestone:<slug>`. Ver "Milestone" |
| `dependencies` | `dependencies` | Con los ids reescritos |
| `parent_task_id` | `parent` | Con el id reescrito |
| `ordinal` | `ordinal` | Tal cual |
| `due_date` | `due` | `YYYY-MM-DD`. Sin medir en el tablero real |
| `documentation` | `documentation` | Lista tal cual |
| `references` | `references` | Lista tal cual |
| `created_date` | `createdAt` | Ver "Fechas" |
| `updated_date` | `updatedAt` | Ver "Fechas" |
| sección `DESCRIPTION` | `description` | El texto entre las marcas `<!-- SECTION:DESCRIPTION:BEGIN -->` y `END`, sin el salto de línea inicial y final. Los `##` de dentro del texto no se interpretan como secciones |
| sección `PLAN` | `plan` | Ídem |
| sección `NOTES` | `notes` | Ídem |
| sección `FINAL_SUMMARY` | `summary` | Ídem |
| criterios de aceptación | `acceptanceCriteria` | Casillas `- [x] #n texto`. La clave es `n` y `checked` sale de la casilla |
| definición de hecho | `acceptanceCriteria` | Con el sufijo ` #dod`. Ver "Definición de hecho" |

**Cualquier clave del frontmatter o sección del cuerpo que no aparezca en esta tabla es un
hallazgo**, con el nombre del fichero y de la clave o la sección. No se ignora en silencio ni
detiene la conversión. Dos casos que hay que confirmar con un fichero real antes de darlos por
cerrados: la clave `due_date`, y el modo en que Backlog.md guarda sus comentarios de discusión. Se
miden creando tareas con el CLI de Backlog.md. Hasta entonces, un comentario es un hallazgo.

### Identificadores

1. El prefijo de origen sale de `config.yml` y el de destino de `task_prefix` del destino. Un id de
   origen que no tenga la forma `<PREFIJO>-<dígitos>` (por ejemplo `TASK-5.1`, con un punto) es un
   error de origen, código 3. Los ceros a la izquierda (`TASK-007`) se ignoran al leer el número y no
   se escriben al salir.
2. Se juntan los ids del destino y los del origen. **Un id de origen conserva su número con el
   prefijo del destino** (`TASK-70` pasa a `BISO-70`) siempre que ese número no exista ya en el destino.
3. **Una tarea que ya está en el destino no se vuelve a importar.** Si el id existe en el destino y la
   tarea de allí tiene el mismo `title` y el mismo `createdAt` que la de origen, es la misma tarea de
   una importación anterior: no se escribe su línea, su id de origen se equipara al que ya tiene, y es
   un hallazgo (`already on the destination, skipped`). Así ejecutar `import` dos veces no duplica el
   tablero. No actualiza la tarea que ya estaba.
4. Los demás ids que existen en el destino son las colisiones. A cada una, en orden de número de
   origen, se le asigna el siguiente número libre a partir de `max(mayor número del destino, mayor
   número del origen) + 1`. Solo esos ids cambian de número; el resto no se desplaza.
5. Con esa tabla de equivalencias se reescriben el `id`, `parent` y `dependencies` de todas las
   tareas, y **toda mención de un id de origen en el texto** (`title`, `description`, `plan`,
   `notes`, `summary` y el texto de los criterios): `TASK-12` pasa a `BISO-12`, o al número
   reasignado. Todas las menciones se sustituyen en una sola pasada, de modo que un número reasignado
   no vuelve a sustituirse. La coincidencia es el prefijo de origen exacto en mayúsculas seguido de
   `-` y dígitos, **distinguiendo mayúsculas**, y solo si no va precedida de una letra, un dígito, `_`
   o `-`, ni seguida de una letra, un dígito, `_` o de `-` y una letra o un dígito. Así `task-10-modelo`,
   que es el nombre de una rama, no se toca, y tampoco `SUBTASK-12`. Se reescribe en cualquier lugar
   del texto, un bloque de código incluido. No se toca `documentation` ni `references`, que son rutas y
   URLs opacas.
6. Una mención con la forma de un id de origen que no corresponde a ninguna tarea del origen se deja
   como está y es un hallazgo.
7. Un `parent` o una dependencia que nombra un id que no está entre las tareas del origen se quita
   de la línea y es un hallazgo, porque `biso` rechaza el lote entero (código 4) por un solo
   identificador que no existe. Los ciclos de padres o de dependencias no se detectan: los rechaza
   `biso new --from --dry-run`, y por eso el ensayo forma parte del uso.
8. Cada id reasignado es un hallazgo: `TASK-12: id BISO-12 is taken on the destination, reassigned to BISO-97`.

### Milestone

El `milestone` de una tarea es un id de milestone (`m-4`). El título se lee del fichero
`milestones/m-4 - *.md` o del de `archive/milestones/`. El slug es el título en minúsculas, sin
diacríticos, con cada tramo de caracteres que no sean letras ni dígitos Unicode convertido en un solo
`-` y sin `-` en los extremos: `Puesta en uso` da `puesta-en-uso` e `Implementación` da
`implementacion`. La tarea recibe la etiqueta `milestone:puesta-en-uso`, después de las etiquetas que
ya tuviera. Una tarea sin milestone no recibe ninguna etiqueta y eso no es un hallazgo. Usa el id como
slug (`milestone:m-9`), con un hallazgo, cuando el fichero del milestone no existe o cuando el slug
sale vacío. Dos milestones distintos que dan el mismo slug se fusionarían en una sola etiqueta, así que
es un hallazgo.

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

Un valor de `labels` o `assignees` solo puede contener letras y dígitos Unicode y los símbolos
`- _ . : @`. Un valor que no lo cumple se quita de la lista y es un hallazgo. Es la regla de `biso`
al escribir, aplicada aquí para que el lote no falle entero por una etiqueta.

### Definición de hecho

Cada elemento de la sección de definición de hecho de una tarea se añade a `acceptanceCriteria` con el
texto seguido de un espacio y `#dod`, después de los criterios propios. Su clave es la siguiente a la
mayor clave de los criterios de la tarea, en orden. Su casilla se conserva. La sección puede no existir
y en ese caso no se añade nada. **La línea no lleva la clave `definitionOfDone`**, así que `biso` no
emite el aviso `imported_dod_merged`.

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
| 3 | El origen no se puede leer: no existe la carpeta, no tiene `tasks/`, un YAML no se puede interpretar, un id no tiene la forma `<PREFIJO>-<dígitos>` o dos ficheros traen el mismo id |
| 4 | El destino no responde: no se encuentra `biso`, o falla una de sus órdenes |
| 5 | Convertido con hallazgos. El NDJSON está escrito, salvo con `--strict`, donde no se escribe nada |

**El código 5 no es un fallo**: el NDJSON está escrito. Encadenar `import` con `biso new --from` en un
`&&` se detiene en el 5, así que quien automatice el paso debe aceptarlo explícitamente.

**El convertidor no escribe nunca en el tablero destino.** Un fallo a mitad de la conversión deja el
destino intacto y, con `--out` a un fichero, no deja un fichero a medias: se escribe en un temporal y
se renombra al terminar.
