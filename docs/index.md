# biso

`biso` es una herramienta de línea de comandos para llevar las tareas de un proyecto, pensada para
que la use un agente automático que trabaja dentro de ese proyecto: la salida es predecible, los
errores se distinguen por su código sin leer el mensaje, y ningún comportamiento depende de dónde se
ejecute el programa.

Esta documentación reúne los cuatro documentos del proyecto:

- **[Especificación](spec/index.md)**: define todos los comandos con su firma, sus parámetros, su
  comportamiento en los casos límite, la salida literal que imprimen, su esquema JSON y sus códigos
  de salida. Es el documento del que se implementa todo.
- **[Decisiones de diseño](DECISIONES.md)**: la razón de cada decisión de la especificación que
  podría parecer arbitraria, y la evidencia que la sostiene.
- **[Estado del arte](ESTADO-DEL-ARTE.md)**: el inventario de los gestores de tareas para agentes
  que ya existen, y el catálogo de sus fallos frente a las respuestas de `biso`.
- **[Pendientes](PENDIENTES.md)**: los huecos que la especificación todavía no decide, las
  incoherencias que no cambian el comportamiento, y las decisiones aplazadas a propósito.
