# Los comentarios

Cada elemento de `comments` es del tipo `Comment`, con clave, autor, instante y cuerpo propios:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | `int` positivo | el programa al crear el comentario |
| `author` | `string` libre | quien llama, y por defecto la identidad `me` |
| `createdAt` | `date` (instante UTC) | el programa, salvo al importar o con `--set-comment-date` (["Comentarios"](../familias-de-banderas.md#comentarios)) |
| `body` | `text` | quien llama |

**La clave se asigna al crear el comentario, con un contador propio de esa lista dentro de esa
tarea, y no se reasigna nunca**, exactamente igual que la de un `Criterion`
(["Los criterios y sus claves estables"](criterios.md#los-criterios-y-sus-claves-estables)). Borrar un comentario no mueve las claves de los
demás, y un `Comment` se direcciona siempre por su `key`, nunca por su posición.

**El autor es texto libre y no se valida contra nada.** Un comentario puede venir de alguien que no
existe en este tablero, y un sistema externo puede usar su propia convención, por ejemplo
`@trello:juan`.

**Los comentarios se guardan y se muestran en orden de inserción, no en orden de `createdAt`.** El
instante de cada uno sigue diciendo la verdad sobre cuándo se escribió, aunque la lista completa no
quede ordenada por él: `biso answer` añade al final un comentario con un instante pasado, el de la
pregunta que responde.

**El cuerpo y el autor de un comentario no se editan nunca, por ninguna vía.** Un comentario es el
registro de una conversación, y lo que se dijo no se reescribe. Lo que sí se puede corregir, con las
banderas dedicadas de ["Comentarios"](../familias-de-banderas.md#comentarios) y nunca con una escritura general sobre la
tarea, es borrar el comentario entero (`--rm-comment`) o corregir únicamente su fecha
(`--set-comment-date`). La razón, con el caso que la motiva, está en
["Borrar o corregir la fecha de un comentario"](../../DECISIONES.md#borrar-o-corregir-la-fecha-de-un-comentario) de `DECISIONES.md`.
