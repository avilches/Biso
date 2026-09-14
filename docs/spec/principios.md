# Los principios

Las reglas de esta lista. El resto del documento es una consecuencia de ellas.

1. **En un campo de vocabulario cerrado (`status`, `type`, `priority`), un valor que el tablero no
   conoce es un error, se esté leyendo o escribiendo, y siempre con la misma regla de coincidencia.**
   Un filtro con un valor imposible nunca devuelve una lista vacía. Así, que un listado, con filtros o
   sin ellos, no devuelva ninguna tarea significa que de verdad no hay tareas que cumplan lo pedido, y
   quien lo recibe puede actuar en consecuencia; nunca que un filtro mal escrito haya ocultado todos
   los resultados. Las etiquetas y las
   personas son distintas a propósito: su vocabulario solo es cerrado al leer, no al escribir (sección
   ["Los vocabularios del tablero y la regla de validación"](vocabularios.md)).

2. **Un nombre significa siempre lo mismo, en todos los comandos.** No existen flags con el
   mismo nombre y semántica distinta según dónde se usen, ni dos nombres para el mismo concepto.

3. **Ningún flag de escritura depende de una regla que haya que conocer de antemano.** Cada una
   lleva su propio verbo en el nombre: `--add-labels` añade, `--rm-labels` quita, `--clear-labels`
   vacía y `--replace-labels` sustituye la lista entera. Ninguna forma se deriva de otra ni de la
   ausencia de un prefijo.

4. **La salida por defecto de una escritura es lo que quien llama no sabía.** Nunca el eco de lo que
   acaba de escribir.

5. **Una operación del flujo de trabajo es un comando.** Empezar una tarea y terminarla tienen nombre
   propio y cuestan una llamada cada uno.

6. **Lo que se puede hacer sobre una tarea se puede hacer sobre muchas en una sola llamada, y la
   llamada entera se valida antes de escribir nada.** Crear doscientas tareas es un `biso new --from`
   con un fichero, no doscientas llamadas, y cambiar diez es un `biso set` con diez referencias. Si
   una sola de las tareas o de las líneas no es válida, no se escribe ninguna. No es idempotencia:
   repetir la misma llamada repite su efecto, y añadir un comentario dos veces deja dos comentarios.

7. **La forma de la salida no depende de si hay un terminal detrás.** Solo el color mira el terminal.
   Los datos, nunca.

---

