# Los principios

Las reglas de esta lista. El resto del documento es una consecuencia de ellas.

1. **Un valor que el tablero no conoce es un error, se esté leyendo o escribiendo, y siempre con la
   misma regla de coincidencia.** Un filtro con un valor imposible nunca devuelve una lista vacía.
   Como consecuencia, una lista vacía es un hecho sobre el tablero y quien la recibe puede actuar en
   consecuencia.

2. **Un nombre significa siempre lo mismo, en todos los comandos.** No existen banderas con el
   mismo nombre y semántica distinta según dónde se usen, ni dos nombres para el mismo concepto.

3. **Ninguna bandera de escritura depende de una regla que haya que conocer de antemano.** Cada una
   lleva su propio verbo en el nombre: `--add-labels` añade, `--rm-labels` quita, `--clear-labels`
   vacía y `--replace-labels` sustituye la lista entera. Ninguna forma se deriva de otra ni de la
   ausencia de un prefijo.

4. **La salida por defecto de una escritura es lo que quien llama no sabía.** Nunca el eco de lo que
   acaba de escribir.

5. **Un gesto del flujo de trabajo es un comando.** Empezar una tarea y terminarla tienen nombre
   propio y cuestan una llamada cada uno.

6. **Todo lo que se hace una vez se puede hacer cien veces, y se valida antes de escribir nada.**

7. **La forma de la salida no depende de si hay un terminal detrás.** Solo el color mira el terminal.
   Los datos, nunca.

---

