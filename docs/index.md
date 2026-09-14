# biso

`biso` es una herramienta de línea de comandos para llevar las tareas de un proyecto, pensada para
que la use un agente automático que trabaja dentro de ese proyecto: la salida es predecible, los
errores se distinguen por su código sin leer el mensaje, y ningún comportamiento depende de dónde se
ejecute el programa.

Esta documentación reúne los documentos del proyecto:

- **[Conceptos](concepts.md)**: los cinco conceptos que conviene tener claros antes de ver un solo
  comando (el tablero, los estados, los criterios, la urgencia y la identidad). Está en inglés, como el
  tutorial.
- **[Tutorial](tutorial/index.md)**: aprender `biso` desde cero siguiendo la vida de una tarea en trece
  situaciones, desde que llegas a un proyecto que no conoces hasta que cierras la tarea. Está en inglés.
- **[Especificación](spec/index.md)**: define todos los comandos con su firma, sus parámetros, su
  comportamiento en los casos límite, la salida literal que imprimen, su esquema JSON y sus códigos
  de salida. Es el documento del que se implementa todo.
- **[Decisiones de diseño](decisiones/index.md)**: la razón de cada decisión de la especificación que
  podría parecer arbitraria, y la evidencia que la sostiene.
- **[Estado del arte](estado-del-arte/index.md)**: el inventario de los gestores de tareas para agentes
  que ya existen, y el catálogo de sus fallos frente a las respuestas de `biso`.
