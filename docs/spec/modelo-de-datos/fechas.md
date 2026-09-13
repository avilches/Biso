# Las fechas

`createdAt`, `updatedAt` y el instante de cada comentario los pone el programa con el reloj del
sistema, en UTC y con precisión de segundo.

**Se pueden fijar solo al importar**, es decir, en `biso new --from`. En cualquier otro sitio son un
hecho observado y no un dato que se negocie, **con una única excepción**: el instante de un
comentario ya escrito se puede corregir con `--set-comment-date`
(["Comentarios"](../familias-de-flags.md#comentarios), ["Borrar o corregir la fecha de un comentario"](../../decisiones/detalles.md#borrar-o-corregir-la-fecha-de-un-comentario)). Es una corrección de un dato ya
observado, no una negociación nueva, y por eso no abre la puerta a hacer lo mismo con `createdAt`,
`updatedAt` ni con `question.askedAt`.

**Una excepción de forma, no de fondo:** `biso answer` escribe el comentario en que se convierte la
pregunta con el instante en que esa pregunta se hizo, no con el de la respuesta. No negocia nada,
porque ese instante ya lo había observado el programa al crear la pregunta; solo lo traslada.

`question.askedAt` es una cuarta fecha importable, junto a `createdAt`, `updatedAt` y el instante de
cada comentario, y sigue la misma regla que ellas: es opcional, y si `biso new --from` no la trae,
toma el instante de la importación.

**`leaseExpiresAt` es la quinta fecha importable y es la única que no sigue esa regla**, así que se
cuenta aparte a propósito. Si no viene, no se rellena con nada: la tarea llega sin arrendamiento, que
es lo que significa no traerlo. Rellenarla con el instante de la importación crearía un arrendamiento
que nadie ha reclamado, y encima a nombre de nadie, porque `leaseHolder` no es una fecha y no tiene
ningún valor por defecto que ponerle. Los dos vienen juntos o no viene ninguno (la regla de
[La importación](../lease.md#la-importación) en `lease.md`).
