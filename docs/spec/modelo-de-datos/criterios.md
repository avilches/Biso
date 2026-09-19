# Los criterios y sus claves estables

`acceptanceCriteria` y `definitionOfDone` son listas del mismo tipo, `Criterion`, y cada elemento
tiene tres partes, sin fecha ni autor propios:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | `int` positivo | el programa al crear el elemento |
| `text` | `string` (["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)) | quien llama |
| `checked` | `bool` | quien llama |

**La clave se asigna al crear el elemento, con un contador propio de esa lista dentro de esa tarea, y
no se reasigna nunca.** Quitar un elemento no mueve las claves de los demás: una tarea puede tener
perfectamente los criterios `#1` y `#3` y ninguno más. Cada tarea lleva dos contadores, uno por
lista, que solo crecen. Un `Criterion` se direcciona siempre por su `key`, nunca por su posición en
la lista.

Consecuencias que hay que respetar en toda la implementación:

- Los selectores de la sección ["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios) trabajan sobre la clave, **nunca** sobre la posición.
- `acTotal` y `dodTotal`, allá donde aparezcan, son **el número de elementos presentes**, nunca la
  clave más alta. Una tarea con los criterios `#1` y `#3` tiene `acTotal` igual a 2.
- Los elementos se muestran y se exportan en el orden en que están en la lista, que es el orden en
  que se crearon salvo que se haya sustituido la lista entera.
