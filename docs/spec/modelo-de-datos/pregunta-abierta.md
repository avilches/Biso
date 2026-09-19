# La pregunta abierta

`question` es del tipo `Question`, un único valor y nunca una lista. Comparte con
[`Comment`](comentarios.md#los-comentarios) el autor, el instante y el cuerpo, pero no lleva `key`:
un valor único no es un elemento de una lista que haga falta direccionar.

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | `string` libre (["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)) | el programa, con la identidad `me`, salvo al importar |
| `askedAt` | `date` (instante UTC) | el programa, salvo al importar |
| `body` | `text` | quien llama |

Vacío es lo normal. Con contenido significa que la tarea espera la respuesta de una persona, esté en
el estado que esté, y entonces el derivado `waiting` es cierto; vacío, `waiting` es falso. Lleva tres
partes y no una sola porque al responderse se convierte literalmente en un comentario, con `biso
answer`, y para eso hacen falta su autor y su instante originales, no los de quien responde.
