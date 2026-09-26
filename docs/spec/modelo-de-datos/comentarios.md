# Los comentarios

Cada elemento de `comments` es del tipo `Comment`, con clave, autor, instante y cuerpo propios:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | `int` positivo | el programa al crear el comentario |
| `author` | `string` libre (["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)) | quien llama, y por defecto la identidad `me` |
| `createdAt` | `date` (instante UTC) | el programa, salvo al importar o con `--set-comment-date` (["Comentarios"](../familias-de-flags.md#comentarios)) |
| `body` | `text` | quien llama |

**La clave se asigna al crear el comentario, con un contador propio de esa lista dentro de esa
tarea, y no se reasigna nunca**, exactamente igual que la de un `Criterion`
(["Los criterios y sus claves estables"](criterios.md#los-criterios-y-sus-claves-estables)). Borrar un comentario no mueve las claves de los
demás, y un `Comment` se direcciona siempre por su `key`, nunca por su posición.

**El autor es texto libre y no se valida contra ningún vocabulario.** Un comentario puede venir de
alguien que no existe en este tablero, y un sistema externo puede usar su propia convención, por
ejemplo `@trello:juan`. Sí sigue la regla general de un campo `string`: no admite un salto de línea
literal (["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)).

**El cuerpo nunca está vacío.** Ningún camino del programa puede dejar guardado un `Comment` sin
`body`: ni el flag `--comment`, ni el texto posicional de `biso comment` que es la misma escritura
escrita de otra forma, ni un lote de `biso new --from` o de `biso init --from`, que rechaza la línea
en vez de descartar el comentario
(["El valor vacío"](../valores-de-entrada.md#el-valor-vacío), ["El modo lote"](../cmd/new.md#el-modo-lote)). Es la misma garantía que ya vale para
`title`, y la razón está en
["Un comentario vacío o `null` en un lote es un fallo de validación"](../../decisiones/detalles.md#un-comentario-vacío-o-null-en-un-lote-es-un-fallo-de-validación).

**Los comentarios se guardan y se muestran en el orden en que se crean, no en el de `createdAt`.**
Cada comentario nuevo se añade al final de la lista, y ese es el orden en que se listan siempre.
Corregir la fecha de un comentario con `--set-comment-date` no lo mueve de sitio, aunque la fecha
nueva sea futura: sigue apareciendo donde estaba. El instante de cada uno sigue diciendo la verdad
sobre cuándo se escribió, aunque la lista completa no quede ordenada por él: `biso answer` añade al
final un comentario con un instante pasado, el de la pregunta que responde.

**El cuerpo y el autor de un comentario no se editan nunca, por ninguna vía.** Un comentario es el
registro de una conversación, y lo que se dijo no se reescribe. Lo que sí se puede corregir, con los
flags dedicados de ["Comentarios"](../familias-de-flags.md#comentarios) y nunca con una escritura general sobre la
tarea, es borrar el comentario entero (`--rm-comment`) o corregir únicamente su fecha
(`--set-comment-date`). La razón, con el caso que la motiva, está en
["Borrar o corregir la fecha de un comentario"](../../decisiones/detalles.md#borrar-o-corregir-la-fecha-de-un-comentario).
