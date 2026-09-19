# El autor de una tarea

`author` se fija una sola vez, al crear la tarea, y después solo cambia si alguien pasa
`--author` de forma explícita (nuevo valor) o `--clear-author`
(["Campos escalares"](../familias-de-flags.md#campos-escalares), lo vacía).

| Al crear la tarea | Valor de `author` |
|---|---|
| Se pasa `--author <persona>` | esa persona, tal cual |
| No se pasa, y hay identidad configurada | la identidad de quien llama, según la precedencia de ["Variables de entorno"](../invocacion.md#variables-de-entorno) |
| No se pasa, y no hay identidad configurada | vacío, sin aviso |

`--author ""` no es una forma de crear la tarea con autor vacío: sigue la regla general de
["El valor vacío"](../valores-de-entrada.md#el-valor-vacío), que la trata como cualquier otro escalar
sin vocabulario. `author` es del tipo `string` y por tanto no admite un salto de línea literal
(["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)).

El caso sin identidad no es un error y no imprime nada: a diferencia de `--mine`, de la
autoasignación de `biso start`, del autor de un comentario, de `biso ask` y de `biso answer`, que sí
la necesitan y están cubiertos por la tabla de ["Variables de entorno"](../invocacion.md#variables-de-entorno), una tarea sin autor es válida.

En el lote de `biso new --from`, un objeto que trae `author` conserva ese valor, y uno que no lo
trae aplica las mismas reglas de esta tabla.
