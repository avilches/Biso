# Quién reporta una tarea

`reporter` se fija una sola vez, al crear la tarea, y después solo cambia si alguien pasa
`--reporter` de forma explícita.

| Al crear la tarea | Valor de `reporter` |
|---|---|
| Se pasa `--reporter <persona>` | esa persona, tal cual |
| No se pasa, y hay identidad configurada | la identidad de quien llama, según la precedencia de ["Variables de entorno"](../invocacion.md#variables-de-entorno) |
| No se pasa, y no hay identidad configurada | vacío, sin aviso |
| Se pasa `--reporter ""` | vacío |

El caso sin identidad no es un error y no imprime nada: a diferencia de `--mine`, de la
autoasignación de `biso start`, del autor de un comentario, de `biso ask` y de `biso answer`, que sí
la necesitan y están cubiertos por la tabla de ["Variables de entorno"](../invocacion.md#variables-de-entorno), una tarea sin quien la reporte es válida.

En el lote de `biso new --from`, un objeto que trae `reporter` conserva ese valor, y uno que no lo
trae aplica las mismas reglas de esta tabla.
