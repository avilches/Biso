# El modelo de datos de una tarea

Este es el modelo **lógico**. Describe qué campos tiene una tarea, de qué tipo son y quién los
escribe. No dice nada de cómo se guardan.

**Todo campo tiene un tipo concreto, y las tablas de abajo lo dicen para cada uno, sin excepción.**
Los tipos son `string` (texto de una línea; ningún `string` admite un salto de línea literal, ver
["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)),
`text` (bloque de prosa, con saltos de línea),
`enum(...)` (un vocabulario cerrado, entre paréntesis; dice "configurable" cuando el tablero puede
ampliarlo), `date` (una fecha o un instante; el formato exacto se dice en la fila del campo), `int`,
`float`, `bool`, `list<string>` (varios valores simples separados por coma al escribir,
direccionables por su propio valor, nunca por posición), y `map<string,string>` (pares clave-valor
con las claves declaradas de antemano, direccionables por su clave). Los campos con estructura
propia usan tipos con nombre, cada uno definido en su propia página: `list<Criterion>`,
`list<Comment>` y `Question` (esta última con un único valor, nunca una lista).

**¿Tiene fecha y autor propios, y una clave estable para señalar un elemento suyo dentro de una
lista?** Esa pregunta solo tiene sentido para los tipos con nombre, es decir, los que tienen
estructura propia en vez de ser un valor simple como `string` o `list<string>`: `Criterion`,
`Comment` y `Question`. Cada uno la contesta en su propia página: ["Los criterios y sus claves
estables"](criterios.md) para `Criterion`, ["Los comentarios"](comentarios.md) para `Comment`,
["La pregunta abierta"](pregunta-abierta.md) para `Question`. Ningún otro tipo tiene fecha ni autor
propios.

## Campos automáticos

| Campo | Tipo | Nullable | Mutable |
|---|---|---|---|
| `id` | `string` (`PREFIX-<n>`) | no | no |
| `createdAt` | `date` (instante UTC, precisión de segundo) | no | solo al importar |
| `updatedAt` | `date` (instante UTC, precisión de segundo) | no | sí, en cada escritura que cambie algo, o al importar; no cuenta un cambio en `leaseExpiresAt` o `leaseHolder` por sí solo |
| `archived` | `bool` | no, `false` por defecto | sí, solo con `biso archive` / `--unarchive`, o al importar |
| `leaseExpiresAt` | `date` (instante UTC) | sí | sí, ver [`lease.md`](../lease.md), o al importar |
| `leaseHolder` | `string` (texto de persona) | sí | sí, solo con esos dos, o al importar; ver [`lease.md`](../lease.md) |

Precisiones:

- **`archived` solo lo cambia `biso archive` y `biso archive --unarchive`.** No hay un flag de
  campo de la sección ["Las familias de flags"](../familias-de-flags.md) para él: archivar es un gesto de flujo de trabajo con nombre propio,
  según el principio 5.
- **`leaseExpiresAt` lo fija el programa**, nunca quien llama; ver [La renovación](../lease.md#la-renovación) en `lease.md` para la fórmula y para cuándo.
- **`leaseHolder` solo lo fija el programa, con [`biso start`](../cmd/verbos-del-ciclo.md#biso-start) y con [`biso new --start`](../cmd/new.md)**; ver [La renovación](../lease.md#la-renovación) en `lease.md`.
- **Las reglas de `leaseExpiresAt` y `leaseHolder`** (cuándo cuenta como vencido el arrendamiento,
  cuándo se renueva, cuándo se vacía y qué hace la importación con ellos) tienen su propio documento,
  [`lease.md`](../lease.md), porque dominan la mutabilidad de una tarea mucho más que cualquier otro
  campo.

## Campos fijados por el usuario

| Campo | Tipo | Notas |
|---|---|---|
| `title` | `string` | obligatorio |
| `status` | `enum(...)`, configurable (ver `statuses` en [`biso config`](../cmd/config.md)) | obligatorio |
| `type` | `enum(...)`, configurable (ver `types` en [`biso config`](../cmd/config.md)) | |
| `priority` | `enum(...)`, configurable (ver `priorities` en [`biso config`](../cmd/config.md)) | |
| `parent` | `string` (referencia a otra tarea); ver ["Las relaciones entre tareas"](relaciones.md) | |
| `assignees` | `list<string>` (textos de persona) | |
| `author` | `string` (texto de persona); ver ["El autor de una tarea"](autor.md) | |
| `labels` | `list<string>` | |
| `dependencies` | `list<string>` (referencias a tareas); ver ["Las relaciones entre tareas"](relaciones.md) | |
| `references` | `list<string>`; ver ["Las relaciones entre tareas"](relaciones.md) | el único campo de punteros: un documento es una referencia más |
| `due` | `date` (`YYYY-MM-DD`) | |
| `ordinal` | `int` (>= 0) | |
| `description` | `text` | |
| `plan` | `text` | |
| `notes` | `text` | |
| `summary` | `text` | |
| `acceptanceCriteria` | `list<Criterion>`; ver ["Los criterios y sus claves estables"](criterios.md) | |
| `comments` | `list<Comment>`; ver ["Los comentarios"](comentarios.md) | se añade, se borra entero, o se corrige solo la fecha; nunca se edita el cuerpo ni el autor |
| `question` | `Question`; ver ["La pregunta abierta"](pregunta-abierta.md) | solo se cambia con `biso ask`, `biso answer`, o al importar |

Precisiones para los campos de esta tabla que no son enteramente de quien llama:

- **`author` se fija una sola vez, al crear la tarea, y con reglas propias** que dependen de si se
  pasa `--author` y de si hay identidad configurada; están completas en ["El autor de una
  tarea"](autor.md).
- **`comments` no se edita nunca por una escritura general sobre la tarea.** El cuerpo y el autor de
  un comentario no se editan jamás, por ninguna vía; lo único que admiten los flags dedicados de
  ["Comentarios"](../familias-de-flags.md#comentarios) es borrar el comentario entero
  (`--rm-comment`) o corregir únicamente su fecha (`--set-comment-date`). La razón, con el caso que
  la motiva, está en
  ["Borrar o corregir la fecha de un comentario"](../../decisiones/detalles.md#borrar-o-corregir-la-fecha-de-un-comentario).
- **`question` se llena con `biso ask` (autor e instante los fija el programa, el cuerpo lo da quien
  llama) y se vacía con `biso answer`.** Los tres detalles están en ["La pregunta abierta"](pregunta-abierta.md).
- **Un `ordinal` negativo es error 2 (`USAGE`)**, con el `code` `invalid_number` y el mensaje
  `error: ordinal cannot be negative: -1`. El tipo de la tabla es `int (>= 0)`, y el cero es un valor
  legítimo y no una ausencia: la forma de dejar el campo sin valor es `--clear-ordinal`.

## Los campos derivados

| Campo | Tipo |
|---|---|
| `urgency` | `float`; ver ["La urgencia"](urgencia.md) |
| `acDone`, `acTotal` | `int` |
| `commentCount` | `int` |
| `blocks` | `list<string>` |
| `blocked`, `waiting` | `bool` |
| `leaseExpired` | `bool` |

**Ninguno de estos campos se guarda.** Se calculan al leer, y son exactamente los campos que
[`biso export`](../cmd/export.md) no escribe y que [`biso new --from`](../cmd/new.md) rechaza como
clave desconocida: `urgency`, `acDone`, `acTotal`, `commentCount`, `blocks`,
`blocked`, `waiting` y `leaseExpired`. Esta es la única lista de campos derivados del documento; las
demás páginas remiten a ella.

**`blocks` se ordena por identificador ascendente**, el mismo criterio de desempate que usa
["`biso ls`"](../cmd/ls.md#comportamiento-caso-a-caso) para cualquier listado de tareas, para no tener
un segundo criterio de orden en el programa. Una tarea dependiente que no se puede decodificar al leer
se excluye de `blocks`, igual que una archivada sin terminar (["La urgencia"](urgencia.md)): no se
puede afirmar que bloquea nada de un dato que no se puede interpretar.

## Precisiones generales sobre la mutabilidad

- **"No mutable" significa que ningún flag del programa lo cambia**, como pasa con `id` en la tabla
  de Campos automáticos. `updatedAt` lo reescribe el programa en cada operación que cambie algo,
  aunque ningún flag lo controle.
- **Se pueden fijar fechas solo al importar**, es decir, en `biso new --from`; los detalles completos
  están en ["Las fechas"](fechas.md).
