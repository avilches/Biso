# Biso

`biso` es una herramienta de línea de comandos para llevar las tareas de un proyecto, pensada para
que la use un agente automático que trabaja dentro de ese proyecto. La salida es predecible, los
errores se distinguen por su código sin leer el mensaje, y ningún comportamiento depende de dónde se
ejecute el programa.

## Estado: especificación cerrada, sin una línea de código

Lo que hay es [`docs/SPEC.md`](docs/SPEC.md), y es el documento del que se implementa todo. Define
los diecinueve comandos con su firma, su tabla de parámetros, su comportamiento en los casos límite, la
salida literal que imprimen, su esquema JSON, sus códigos de salida y el texto exacto de su ayuda.
Está escrito para que alguien lo implemente entero sin preguntar nada.

**No hay que rediseñar nada por libre.** Si al implementar aparece un caso que la especificación no
cubre, lo correcto es añadirlo a la especificación y luego implementarlo, no resolverlo solo en el
código. Y si una regla parece arbitraria, su razón está en
[`docs/DECISIONES.md`](docs/DECISIONES.md) antes de cambiarla.

Su sección 15 dice por dónde empezar, en el orden en que cada pieza paga lo que cuesta: el modelo de
datos, el algoritmo de coincidencia (una función pura de la que dependen todos los comandos), los
cuatro comandos del trabajo diario, los verbos del ciclo, el mensaje de arranque, el lote y la
exportación, y el resto.

## El modelo de estados está cerrado

Los cuatro requisitos que lo tocaban ya están decididos: dos resueltos sin ningún papel de estado
nuevo, uno aplazado a la decisión de persistencia y uno retirado. El detalle, con la evidencia detrás
de cada decisión, está en la sección 9 de [`docs/DECISIONES.md`](docs/DECISIONES.md).

## Lo que falta por decidir, y bloquea

Es una, y hasta que no esté no se puede escribir código de verdad.

**Cómo se guardan los datos.** La especificación define la interfaz y el modelo de datos lógico, y
deliberadamente no dice si detrás hay ficheros, una base de datos o cualquier otra cosa. Donde el
almacenamiento afecta a lo observable, enuncia el requisito y deja el mecanismo abierto. Al tomar
esta decisión hay que escribir aparte lo que hoy no tiene respuesta: **qué ocurre con una tarea que
existe en una versión del proyecto y no en otra**, porque hoy la única promesa son los tres mensajes
distintos de "no la encuentro" de la sección 7.3. Si la respuesta pasa por git, ahí entra también
todo lo que la sección 14 deja fuera por depender de esto. **De esta decisión depende también el
arrendamiento con caducidad** que el modelo de estados aplazó, para detectar una tarea que un agente
cogió y cuya sesión murió sin liberarla (sección 9.2 de `docs/DECISIONES.md`).

## Cosas que conviene tener presentes al implementar

- **El lenguaje de implementación sigue sin decidir**, pero a diferencia de la persistencia no
  bloquea la especificación. El único requisito que sale del documento es que el programa arranque
  rápido, porque un agente lo invoca muchas veces en una sesión.
- **El mensaje de arranque tiene un tope duro de 5.120 bytes**, repartido en 3.072 para la parte fija
  y 2.048 para el resumen del tablero. No es un objetivo, es una prueba de la suite. El texto actual
  ocupa 4.062 bytes.
- **La simetría entre `biso export` y `biso new --from` es una prueba, no una intención.** Exportar
  un tablero e importarlo en otro vacío tiene que dar dos tableros idénticos campo a campo, con
  identificadores, fechas y claves de criterios incluidas.
- **Los ejemplos de salida del documento se generan, no se escriben a mano.** El algoritmo de
  columnas de `biso ls` y de `biso prime` está especificado como algoritmo justamente para eso.
  Cualquier cambio en él obliga a regenerar los ejemplos y a comprobar que coinciden carácter a
  carácter.
- **Un valor que no existe es siempre un error, se esté escribiendo o leyendo.** Un filtro mal
  escrito nunca puede devolver una lista vacía, porque quien la lee la interpreta como un hecho sobre
  el tablero.

## Reglas de este repositorio

- **El trabajo va en un worktree**, en `.claude/worktrees/<rama>`, nunca editando `main`
  directamente.
- **La documentación y los comentarios van en español.** Los identificadores del código y todo lo que
  es interfaz del programa (comandos, banderas, textos de ayuda, mensajes de error, claves JSON) van
  en inglés.
- **Nunca em-dash**, en ningún texto: ni en documentación, ni en código, ni en mensajes de commit.
- **Los mensajes de commit y las descripciones de PR no llevan coautoría** ni mención de haber sido
  generados por un agente.
