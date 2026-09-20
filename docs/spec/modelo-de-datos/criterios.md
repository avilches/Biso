# Los criterios y sus claves estables

`acceptanceCriteria` es una lista de elementos del tipo `Criterion`, y cada elemento tiene tres
partes, sin fecha ni autor propios:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | `int` positivo | el programa al crear el elemento |
| `text` | `string` (["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string)) | quien llama |
| `checked` | `bool` | quien llama |

**Es la única lista de comprobación de una tarea.** No hay una segunda lista para lo que haga falta
antes de cerrar, del tipo "alguien más lo ha revisado": eso es un criterio de aceptación más. El
porqué, con la medida que lo sostiene, está en
["Se retira la definición de hecho"](../../decisiones/detalles.md#se-retira-la-definición-de-hecho).

**La clave se asigna al crear el elemento, con un contador propio de esa tarea, y no se reasigna
nunca.** Quitar un elemento no mueve las claves de los demás: una tarea puede tener perfectamente los
criterios `#1` y `#3` y ninguno más. Ese contador solo crece. Un `Criterion` se direcciona siempre
por su `key`, nunca por su posición en la lista.

Consecuencias que hay que respetar en toda la implementación:

- Los selectores de la sección ["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios) trabajan sobre la clave, **nunca** sobre la posición.
- `acTotal`, allá donde aparezca, es **el número de elementos presentes**, nunca la clave más alta.
  Una tarea con los criterios `#1` y `#3` tiene `acTotal` igual a 2.
- Los elementos se muestran y se exportan en el orden en que están en la lista, que es el orden en
  que se crearon salvo que se haya sustituido la lista entera.
- Una lista con un elemento cuya clave no sea un entero positivo, o con dos elementos que compartan
  clave, **no se guarda**. Las claves las pone el programa, así que eso no es un error de quien llama
  sino un fallo de quien construyó la tarea, y se para antes de tocar el almacén: nunca sale como un
  error de la base de datos nombrando una tabla. **Y a propósito no tiene un `code` propio de
  ["Los identificadores de error"](../contrato-json.md#los-identificadores-de-error)**, porque
  ninguna invocación puede provocarlo: si ocurre, el programa está roto, así que sale por el
  código 1 de ["El código 1 es un fallo del programa, no de quien
  llama"](../codigos-de-salida.md#el-código-1-es-un-fallo-del-programa-no-de-quien-llama). Darle un
  `code` lo presentaría como un caso previsto de la especificación, que es justo lo que no es.
