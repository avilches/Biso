# Los presupuestos de arranque y de tamaño

## El presupuesto de arranque

**`biso ls` y `biso prime` sobre un tablero de 300 tareas terminan en menos de 25 milisegundos de
reloj.** Es una regla transversal a todo el camino caliente, el de los comandos que un agente ejecuta
muchas veces a lo largo de una sesión: ninguno de ellos tiene un motivo para tardar más que estos dos.
`biso snapshot` (regla 2 de más abajo) y el sondeo de `biso doctor`
(["El sondeo del sistema de ficheros"](cmd/doctor.md#el-sondeo-del-sistema-de-ficheros), hasta 2
segundos) quedan fuera de esta cifra porque no están en ese camino: son comandos que se llaman a
propósito y de forma esporádica, no en cada consulta. Es una prueba de la suite, no una aspiración, y
se mide en la máquina de referencia.

**La máquina de referencia es la que ejecuta la suite de integración continua del proyecto.** Una
cifra de tiempo sin una máquina no se puede comprobar, porque la misma llamada tarda lo que tarde el
hardware que la ejecuta. Fijar la máquina de referencia como la del propio CI, en vez de describir aquí
un modelo de hardware concreto, es lo que hace que la prueba dé siempre el mismo veredicto para la
misma versión del código, sin que este documento tenga que llevar ni mantener actualizada una ficha
técnica de un ordenador que además dejaría de existir o de venderse.

**Para qué sirve la cifra: detectar una regresión en la máquina que ejecuta la suite, no certificar el
rendimiento de la herramienta sobre un hardware arbitrario.** Quien corra la suite en una máquina
distinta de la de referencia, más lenta o más rápida, no debe leer el resultado como una afirmación
sobre `biso`: debe leerlo como una afirmación sobre esa máquina. Un portátil viejo que supere los
25 ms no dice que la herramienta incumpla su especificación, dice que en ese portátil el número es
otro; lo que sí dice algo es que el número suba en la propia máquina de referencia de una versión a
la siguiente, porque ahí el hardware no ha cambiado y lo único que puede haber cambiado es el código.

**La composición del tablero de 300 tareas es indiferente, y por eso no se fija.** No importa cuántas
estén en cada estado, ni si alguna tiene una pregunta abierta o un arrendamiento vencido: ninguno de
los comandos se ramifica según el contenido de una tarea concreta, así que su coste crece de forma
esencialmente lineal con el número de tareas y no con su composición. Fijar un reparto arbitrario no
añadiría ninguna garantía que esta razón no dé ya, y la ambigüedad se cierra con la explicación, no con
una tabla de reparto que nadie necesita reproducir.

Estas reglas protegen ese presupuesto, y ningún comando se aparta de ellas:

1. **Ningún comando hace al arrancar trabajo que nadie ha pedido.** Ni una consulta que no alimente
   una línea de lo que esa invocación va a imprimir, ni una comprobación de más, ni una llamada de
   red: todo lo que no sirve a la salida de la llamada concreta se paga en cada una de las muchas
   veces que un agente ejecuta el programa a lo largo de una sesión, se haya pedido o no. El ejemplo
   documentado está en este documento y no en otra herramienta: la resolución del tablero llevaba una
   comprobación de un `.git` en cada directorio del camino hacia arriba, para frenar la búsqueda del
   puntero, y se retiró al ver que además de costar comprobaciones en cada llamada no protegía de lo
   que pretendía (sección ["La decisión de persistencia"](../decisiones/persistencia.md#la-decisión-de-persistencia)).
2. **Ningún comando ejecuta un programa ajeno en su camino caliente.** `biso snapshot` es la única
   excepción, y así lo dice la sección ["`biso snapshot`"](cmd/snapshot.md). Invocar `git`, que es el sistema de control de versiones
   por defecto, cuesta unos 12 milisegundos medidos, casi la mitad de este presupuesto entero gastada en
   una sola llamada. Este presupuesto es además la razón por la que la resolución del tablero de la
   sección ["Cómo se elige el tablero"](resolucion-del-tablero.md) no mira el control de versiones en absoluto, ni siquiera leyendo ficheros: frenar la
   búsqueda del puntero donde un repositorio empieza obligaría a respetar sus reglas de exclusión para
   que el freno significara algo, y eso no se puede reimplementar de forma fiable ni preguntar sin
   invocar el programa. Que `biso snapshot` sí haga esas preguntas (["`biso snapshot`"](cmd/snapshot.md)) no contradice nada:
   ese comando ya está fuera del camino caliente por definición.
3. **La palanca mayor no es que cada llamada sea más rápida: es que haga falta hacer menos llamadas.**
   Para eso existe `biso prime` (sección ["`biso prime`, el arranque de una sesión"](cmd/prime.md)), que sustituye el ciclo entero de leer guías sueltas y
   encadenar comandos por un solo mensaje al principio de la sesión; la sección ["El presupuesto del mensaje de arranque"](../decisiones/vocabulario-y-mensaje-de-arranque.md#el-presupuesto-del-mensaje-de-arranque)
   mide lo que cuesta la alternativa de no tenerlo.

---

## El presupuesto de tamaño

El mensaje tiene un **tope duro de 5.504 bytes**, que se comprueba en la suite de pruebas y se reparte
en dos partes que suman exactamente ese tope:

- **La parte fija no pasa de 3.840 bytes.** Es la línea de título, `COMMANDS`, `FIELD FLAGS`, `RULES` y
  el párrafo final ("Pick one, ..."): nada de esto depende del contenido del tablero.
- **El resumen del tablero no pasa de 1.664 bytes.** Es el bloque `BOARD` (nombre, recuento por
  estado, vocabularios, identidad, el aviso de una tarea ilegible), `IN PROGRESS`, `NEEDS ANSWER`,
  `ASSIGNED TO YOU`, `NEXT UP` y las líneas de recuento: todo lo que cambia según qué haya en el
  tablero.

Los bloques no son contiguos entre sí, así que hay líneas en blanco de separación entre ellos: **cada
línea en blanco se cuenta en la parte a la que pertenece el bloque que la precede.** Con esta regla,
la línea en blanco que sigue al título es parte fija, la que sigue a `BOARD` es resumen, las que
siguen a `COMMANDS`, `FIELD FLAGS` y `RULES` son parte fija, y las que siguen a `IN PROGRESS`,
`NEEDS ANSWER`, `ASSIGNED TO YOU` y a `NEXT UP` son resumen.

Si el resumen no cupiera en su parte, el orden de recorte es completo y no deja ningún caso sin
definir:

1. Se reduce primero el número de filas de `NEXT UP`.
2. Si no basta, el de `ASSIGNED TO YOU`.
3. Si no basta, el de `NEEDS ANSWER`.
4. Si no basta, el de `IN PROGRESS`.
5. Si aun así no cupiera, cada uno de los cuatro bloques se reduce a su sola línea de recuento.

`ASSIGNED TO YOU` y `NEXT UP` comparten la línea de recuento que ya define la sección ["La salida literal"](cmd/prime.md#la-salida-literal). `IN PROGRESS` y
`NEEDS ANSWER` llevan cada uno la suya, con el mismo patrón: cuántas tareas del bloque quedan
fuera por el recorte y el comando para verlas completas. Para `IN PROGRESS` es
`N more not shown: 'biso ls --active'`, y para `NEEDS ANSWER` es
`N more not shown: 'biso ls --waiting'`.

Con esa lista el tope deja de ser una aspiración y pasa a ser alcanzable siempre. Lo que se ve en el
texto al aplicarla, y las dos cosas que el recorte no toca (el bloque `BOARD` y la salida de
`--json`), están en ["El recorte en cascada"](cmd/prime.md#el-recorte-en-cascada). **El tope se mide
sobre el mensaje sin `--full`**, que es ayuda para quien aprende la herramienta y no parte del
arranque.

El texto literal de la sección ["La salida literal"](cmd/prime.md#la-salida-literal) ocupa **5.039 bytes** con el tablero del ejemplo: **3.550** de
parte fija y **1.489** de resumen. Las dos partes caben dentro de su tope.

**El número que congela el contrato de estabilidad de la sección ["El contrato de estabilidad"](estabilidad.md) es el total, 5.504 bytes**, porque
es el único que quien llama observa. El reparto entre las dos partes puede cambiar sin romper ese
contrato.

