# Las fechas

Cada campo lleva en esta página su prefijo de objeto (`task.`, `comment.`, `question.`), para no
confundir el instante de la tarea con el de uno de sus comentarios o el de su pregunta.

## `task.createdAt`, `task.updatedAt` y `comment.createdAt`

`task.createdAt`, `task.updatedAt` y `comment.createdAt` los pone el programa con el reloj del
sistema, en UTC y con precisión de segundo.

**Se pueden fijar solo al importar**, es decir, en `biso new --from`. En cualquier otro sitio son un
hecho observado y no un dato que se negocie, **con una única excepción**: el `comment.createdAt` de
un comentario ya escrito se puede corregir con `--set-comment-date`
(["Comentarios"](../familias-de-flags.md#comentarios), ["Borrar o corregir la fecha de un
comentario"](../../decisiones/detalles.md#borrar-o-corregir-la-fecha-de-un-comentario)). Es una
corrección de un dato ya observado, no una negociación nueva, y por eso no abre la puerta a hacer lo
mismo con `task.createdAt`, `task.updatedAt` ni con `question.askedAt`.

**Una excepción de forma, no de fondo:** `biso answer` escribe el comentario en que se convierte la
pregunta con el `comment.createdAt` igual al `question.askedAt` de esa pregunta, no con el instante
de la respuesta. No negocia nada, porque ese instante ya lo había observado el programa al crear la
pregunta; solo lo traslada.

## `question.askedAt`

`question.askedAt` es una cuarta fecha importable, junto a `task.createdAt`, `task.updatedAt` y
`comment.createdAt`, y sigue la misma regla que ellas: es opcional, y si `biso new --from` no la
trae, toma el instante de la importación.

## `task.leaseExpiresAt`

`task.leaseExpiresAt` es la quinta fecha importable y es la única que no sigue esa regla: si `biso
new --from` no la trae, la tarea llega sin arrendamiento en vez de tomar el instante de la
importación, y solo se acepta junto con `task.leaseHolder`, nunca uno sin el otro. El detalle completo,
con el porqué de cada regla, está en [`lease.md`](../lease.md#la-importación).

## Una fecha guardada que no es una fecha

`task.createdAt`, `task.updatedAt`, `comment.createdAt` y `question.askedAt` son instantes
`YYYY-MM-DDTHH:MM:SSZ` y `task.due` es un día `YYYY-MM-DD`
(["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias)). Una fecha guardada
con otra forma, o vacía donde el campo no puede estar vacío, hace ilegible la tarea entera, con el
mismo tratamiento en todos los comandos de lectura (["El primer caso: una tarea ilegible"](../garantias.md#el-primer-caso-una-tarea-ilegible)).
Solo `task.due` y `task.leaseExpiresAt` pueden no tener valor.
