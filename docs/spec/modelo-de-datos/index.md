# El modelo de datos de una tarea

Este es el modelo **lógico**. Describe qué campos tiene una tarea, de qué tipo son y quién los
escribe. No dice nada de cómo se guardan.

**Todo campo tiene un tipo concreto, y las tablas de abajo lo dicen para cada uno, sin excepción.**
Los tipos son `string` (texto de una línea), `text` (bloque de prosa, con saltos de línea),
`enum(...)` (un vocabulario cerrado, entre paréntesis; dice "configurable" cuando el tablero puede
ampliarlo), `date` (una fecha o un instante; el formato exacto se dice en la fila del campo), `int`,
`float`, `bool`, `list<string>` (varios valores simples separados por coma al escribir,
direccionables por su propio valor, nunca por posición), y `map<string,string>` (pares clave-valor
con las claves declaradas de antemano, direccionables por su clave). Los campos con estructura
propia usan tres tipos con nombre, cada uno definido en su propia página: `list<Criterion>`,
`list<Comment>` y `Question` (esta última con un único valor, nunca una lista).

**¿Tiene fecha y autor propios, y se puede señalar uno a uno?** Solo se aplica a los tipos con
nombre, y cada uno lo contesta en su propia página: ["Los criterios y sus claves
estables"](criterios.md) para `Criterion`, ["Los comentarios"](comentarios.md) para `Comment`,
["La pregunta abierta"](pregunta-abierta.md) para `Question`. Ningún otro tipo tiene fecha ni autor
propios.

Las tablas siguientes agrupan los campos según quién decide su valor: los que fija el programa
solo, los que fija quien llama, y los que se derivan y nunca se guardan. Un campo con reglas mixtas
aparece en la tabla del que lo fija normalmente, con una nota que explica la mezcla.

## Los campos que fija el programa

| Campo | Tipo | Obligatorio | Mutable |
|---|---|---|---|
| `id` | `string` (`PREFIX-<n>`) | sí | no |
| `createdAt` | `date` (instante UTC, precisión de segundo) | sí | solo al importar |
| `updatedAt` | `date` (instante UTC, precisión de segundo) | sí | solo al importar |
| `archived` | `bool` | sí, `false` por defecto | sí, solo con `biso archive` / `--unarchive`, o al importar |
| `leaseExpiresAt` | `date` (instante UTC) | no | sí, ver [`lease.md`](../lease.md), o al importar |
| `leaseHolder` | `string` (texto de persona) | no | sí, solo con esos dos, o al importar; ver [`lease.md`](../lease.md) |

Precisiones:

- **`archived` solo lo cambia `biso archive` y `biso archive --unarchive`.** No hay una bandera de
  campo de la sección ["Las familias de banderas"](../familias-de-banderas.md) para él: archivar es un gesto de flujo de trabajo con nombre propio,
  según el principio 5.
- **`leaseExpiresAt` lo fija el programa a `ahora + lease_minutes`** (clave de [configuración](../cmd/config.md)); ver [La renovación](../lease.md#la-renovación) en `lease.md` para cuándo.
- **`leaseHolder` solo lo fija el programa, con [`biso start`](../cmd/verbos-del-ciclo.md#biso-start) y con [`biso new --start`](../cmd/new.md)**; ver [La renovación](../lease.md#la-renovación) en `lease.md`.
- **Las reglas de `leaseExpiresAt` y `leaseHolder`** (cuándo cuenta como vencido el arrendamiento,
  cuándo se renueva, cuándo se vacía y qué hace la importación con ellos) tienen su propio documento,
  [`lease.md`](../lease.md), porque dominan la mutabilidad de una tarea mucho más que cualquier otro
  campo.

## Los campos que fija quien llama

| Campo | Tipo | Obligatorio | Mutable |
|---|---|---|---|
| `title` | `string` | sí | sí |
| `status` | `enum(...)`, configurable (ver `statuses` en [`biso config`](../cmd/config.md)) | sí | sí |
| `type` | `enum(...)`, configurable (ver `types` en [`biso config`](../cmd/config.md)) | no | sí |
| `priority` | `enum(...)`, configurable (ver `priorities` en [`biso config`](../cmd/config.md)) | no | sí |
| `parent` | `string` (referencia a otra tarea) | no | sí |
| `assignees` | `list<string>` (textos de persona) | no | sí |
| `reporter` | `string` (texto de persona); ver ["Quién reporta una tarea"](quien-reporta.md) | no | sí |
| `labels` | `list<string>` | no | sí |
| `dependencies` | `list<string>` (referencias a tareas) | no | sí |
| `references` | `list<string>` | no | sí |
| `documentation` | `list<string>` | no | sí |
| `modifiedFiles` | `list<string>` | no | sí |
| `due` | `date` (`YYYY-MM-DD`) | no | sí |
| `ordinal` | `int` (>= 0) | no | sí |
| `ext` | `map<string,string>` | no | sí |
| `description` | `text` | no | sí |
| `plan` | `text` | no | sí |
| `notes` | `text` | no | sí |
| `summary` | `text` | no | sí |
| `acceptanceCriteria` | `list<Criterion>`; ver ["Los criterios y sus claves estables"](criterios.md) | no | sí |
| `definitionOfDone` | `list<Criterion>`; ver ["Los criterios y sus claves estables"](criterios.md) | no | sí |
| `comments` | `list<Comment>`; ver ["Los comentarios"](comentarios.md) | no | se añade, se borra entero, o se corrige solo la fecha; nunca se edita el cuerpo ni el autor |
| `question` | `Question`; ver ["La pregunta abierta"](pregunta-abierta.md) | no | sí, solo con `biso ask`, `biso answer`, o al importar |

Precisiones para los campos de esta tabla que no son enteramente de quien llama:

- **`reporter` se fija una sola vez, al crear la tarea, y con reglas propias** que dependen de si se
  pasa `--reporter` y de si hay identidad configurada; están completas en ["Quién reporta una
  tarea"](quien-reporta.md).
- **`comments` no se edita nunca por una escritura general sobre la tarea.** El cuerpo y el autor de
  un comentario no se editan jamás, por ninguna vía; lo único que admiten las banderas dedicadas de
  ["Comentarios"](../familias-de-banderas.md#comentarios) es borrar el comentario entero
  (`--rm-comment`) o corregir únicamente su fecha (`--set-comment-date`). La razón, con el caso que
  la motiva, está en
  ["Borrar o corregir la fecha de un comentario"](../../DECISIONES.md#borrar-o-corregir-la-fecha-de-un-comentario) de `DECISIONES.md`.
- **`question` se llena con `biso ask` (autor e instante los fija el programa, el cuerpo lo da quien
  llama) y se vacía con `biso answer`.** Los tres detalles están en ["La pregunta abierta"](pregunta-abierta.md).

## Los campos derivados

| Campo | Tipo |
|---|---|
| `urgency` | `float`; ver ["La urgencia"](urgencia.md) |
| `acDone`, `acTotal`, `dodDone`, `dodTotal` | `int` |
| `commentCount` | `int` |
| `blocks` | `list<string>` |
| `blocked`, `waiting` | `bool` |
| `leaseExpired` | `bool` |

**Ninguno de estos campos se guarda.** Se calculan al leer, y son exactamente los campos que
[`biso export`](../cmd/export.md) no escribe y que [`biso new --from`](../cmd/new.md) rechaza como
clave desconocida: `urgency`, `acDone`, `acTotal`, `dodDone`, `dodTotal`, `commentCount`, `blocks`,
`blocked`, `waiting` y `leaseExpired`. Esta es la única lista de campos derivados del documento; las
demás páginas remiten a ella.

## Precisiones generales sobre la mutabilidad

- **"No mutable" significa que ninguna bandera del programa lo cambia.** `updatedAt` lo reescribe el
  programa en cada operación que cambie algo.
- **Se pueden fijar fechas solo al importar**, es decir, en `biso new --from`; los detalles completos
  están en ["Las fechas"](fechas.md).
