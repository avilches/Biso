# Vocabulario de esta especificación

Cada fila es un elemento de `biso`. La primera columna es el nombre que usa el propio programa, en
comandos, flags, claves JSON y claves de configuración, y es el nombre que manda. La última es la
palabra con la que estos documentos, que están en español, se refieren a ese elemento en la prosa. Si
la especificación se tradujera algún día al inglés, esa última columna sobraría y la tabla seguiría
valiendo tal cual, porque lo que define cada fila es la columna del medio. Para una introducción
narrativa a los mismos conceptos, con menos detalle, está la página [Concepts](../concepts.md).

| En el programa | Qué es | En esta documentación |
|---|---|---|
| `board` | Las tareas de un proyecto con su configuración y sus vocabularios | **tablero** |
| `task` | La unidad de trabajo | **tarea** |
| `status` | Uno de los valores configurados en `statuses` | **estado** |
| las claves `*_status` | La función que cumple un estado. Son tres, obligatorios y distintos | **papel** |
| `initial_status` | El papel del estado donde nace una tarea | **inicial** |
| `active_status` | El papel del estado que escribe el agente al coger la tarea | **activo**, activa |
| `terminal_status` | El papel del estado final | **terminal**, terminada |
| `assignees` | Quién debe hacer la tarea. Es la forma en que una persona encarga trabajo | **asignación** |
| `leaseExpiresAt`, `leaseHolder`, `leaseExpired` | Hasta cuándo vale la reserva de un agente sobre una tarea activa. No es un estado | **arrendamiento** |
| `question`, `waiting` | Lo que detiene una tarea a la espera de una persona. No es un estado | **pregunta abierta**, aparcada |
| `archived` | Fuera del tablero activo sin perder nada. No es un estado | **archivada** |
| `blocked` | Depende de alguna tarea sin terminar. Solo dependencias, nunca personas: lo que espera a una persona es una pregunta abierta | **bloqueada** |
| `assignees`, `author`, y el `author` de un comentario o de la pregunta | Quien encarga y quien responde. No es un campo, sino el tipo de valor que llevan estos campos | **persona** |
| | El programa automático que coge tareas y las hace | **agente** |
| `acceptanceCriteria`, `definitionOfDone` | Un elemento de las dos listas de comprobación: `acceptanceCriteria` comprueba que el trabajo hace lo que se pidió, `definitionOfDone` que la tarea cumple lo necesario para poder cerrarse | **criterio** |
| `comments` | Una entrada del histórico cuyo cuerpo y autor no se editan nunca, aunque su fecha se pueda corregir o el comentario entero se pueda borrar | **comentario** |

**Un estado no se llama nunca columna.** En muchas herramientas de tareas cada estado se dibuja como
una columna de un tablero kanban, y es tentador llamarlo así. En esta documentación, **columna**
significa únicamente una de las ocho columnas de la tabla de texto que imprimen `biso ls` y
`biso prime` (["Salida" de `biso ls`](cmd/ls.md#salida)), como la del identificador o la del estado.
Así, "la columna de estado" solo puede significar la segunda columna de esa tabla, y nunca "el estado
To Do".

---

