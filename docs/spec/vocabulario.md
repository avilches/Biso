# Vocabulario de esta especificación

| Español | Qué es | En la interfaz |
|---|---|---|
| **tablero** | Las tareas de un proyecto con su configuración y sus vocabularios | `board` |
| **tarea** | La unidad de trabajo | `task` |
| **estado** | Uno de los valores configurados en `statuses` | `status` |
| **papel** | La función que cumple un estado. Son tres, obligatorios y distintos | las claves `*_status` |
| **inicial** | El papel del estado donde nace una tarea | `initial_status` |
| **activo**, activa | El papel del estado que escribe el agente al coger la tarea | `active_status` |
| **terminal**, terminada | El papel del estado final | `terminal_status` |
| **asignación** | Quién debe hacer la tarea. Es el gesto con el que una persona encarga trabajo | `assignees` |
| **arrendamiento** | Hasta cuándo vale la reserva de un agente sobre una tarea activa. No es un estado | `leaseExpiresAt`, `leaseHolder`, `leaseExpired` |
| **pregunta abierta**, aparcada | Lo que detiene una tarea a la espera de una persona. No es un estado | `question`, `waiting` |
| **archivada** | Fuera del tablero activo sin perder nada. No es un estado | `archived` |
| **bloqueada** | Depende de alguna tarea sin terminar. Solo dependencias, nunca personas | `blocked` |
| **persona** | Quien encarga y quien responde | |
| **agente** | El programa automático que coge tareas y las hace | |
| **criterio** | Un elemento de las dos listas de comprobación | `acceptanceCriteria`, `definitionOfDone` |
| **comentario** | Una entrada inmutable del histórico | `comments` |

Cinco palabras quedan restringidas, en cuatro reglas, y conviene decir a qué en vez de prohibirlas a
secas, porque tres de ellas tienen un uso legítimo:

- **columna** nombra únicamente una columna de la tabla que imprimen `biso ls` y `biso prime`, que
  tiene ocho. Un estado del tablero no se llama nunca columna.
- **tarjeta** y **ticket** nombran únicamente lo que otra herramienta tiene, como una tarjeta de
  Trello. Lo de `biso` es una tarea.
- **panel** no se usa nunca: el conjunto de tareas es el tablero, y lo que abre `biso board` es la
  interfaz web.
- **bloqueada** no se usa nunca referida a una persona. Eso es una pregunta abierta.

---

