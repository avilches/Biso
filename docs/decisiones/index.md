# Por qué `biso` es como es

[`docs/spec/`](../spec/index.md) dice qué hace el programa y nunca por qué. Este documento es el complemento: la razón
de cada decisión que podría parecer arbitraria, y la evidencia que la sostiene.

**Sirve para una cosa concreta:** antes de cambiar una regla de la especificación, hay que mirar aquí
si esa regla existe por algo. Varias de ellas parecen caprichos de estilo y son la respuesta a un
fallo medido en herramientas reales.

Este documento usa las mismas palabras que la especificación y con el mismo significado; la tabla de
["Vocabulario de esta especificación"](../spec/vocabulario.md) es la referencia para las dos.

La evidencia viene de dos sitios. El primero es un estudio del uso real de un gestor de tareas por
agentes automáticos: 856 invocaciones de línea de comandos en 60 sesiones y 10 proyectos a lo largo
de seis días, con la salida de cada llamada medida en bytes. El segundo es la comparación de dos
gestores de tareas maduros, Backlog.md y Taskwarrior, y de los fallos documentados de ambos.

## Las páginas

- [Los principios, y cómo se mantiene la especificación](principios-y-mantenimiento.md)
- [El modelo de estados](modelo-de-estados.md)
- [La persistencia](persistencia.md)
- [El lenguaje de implementación y el rendimiento](lenguaje-y-rendimiento.md)
- [El vocabulario y el mensaje de arranque](vocabulario-y-mensaje-de-arranque.md)
- [Comandos y banderas](comandos-y-flags.md)
- [Decisiones de detalle](detalles.md)
