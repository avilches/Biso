# Orden de escritura, concurrencia y datos dañados

## Orden de aplicación dentro de una escritura

Una sola invocación puede tocar muchos campos. El orden en que se aplican es fijo y **no depende del
orden en que aparecen las banderas en la línea de comandos**, para que el resultado sea reproducible:

1. Todos los `--clear-*`.
2. Todos los `--set-*`.
3. Todos los `--rm-*`.
4. Los añadidos, es decir, los nombres desnudos.
5. Los campos escalares.
6. Los marcados de criterios y de definición de hecho.
7. Los comentarios.

Con este orden, `--clear-label --label urgent` deja exactamente una etiqueta, y `--set-ac "A"
--check all` marca los criterios recién puestos. Dentro de un mismo paso manda el orden de la línea
de comandos: `--label b --label a` deja `b` antes que `a`. Las listas nunca se ordenan solas.

## Concurrencia, atomicidad y garantías observables

Esta sección no describe un mecanismo: enuncia lo que quien llama tiene derecho a observar. Cómo se
consiga es cosa de quien implemente.

1. **Ninguna escritura se observa a medias.** Un lector concurrente ve el tablero como estaba antes
   de una escritura o como quedó después, nunca en un punto intermedio, y esto vale igual para una
   escritura de una tarea que para un lote de doscientas.
2. **Una escritura que afecta a varias tareas es todo o nada.** Si falla por cualquier motivo, ni una
   sola de las tareas implicadas queda modificada, y el código de salida lo dice: 9 si el fallo se
   detectó al validar, 7 si se detectó al escribir. En los dos casos el mensaje afirma explícitamente
   que no se ha escrito nada.
3. **Dos procesos simultáneos nunca asignan el mismo identificador**, aunque trabajen sobre el mismo
   tablero desde copias de trabajo distintas del proyecto.
4. **Dos escrituras simultáneas sobre la misma tarea no se pierden ni se mezclan.** O se aplican una
   después de otra, o una de las dos falla con código 7.
5. **Si el programa no puede obtener el acceso exclusivo que necesita para escribir**, espera hasta
   cinco segundos y luego falla con código 7 sin escribir nada:
   ```
   error: the board is busy, another process is writing to it
   hint: retry in a moment; nothing was written
   ```
6. **Las lecturas nunca fallan por culpa de una escritura en curso**, y nunca la bloquean.

**Las seis hablan de las escrituras del tablero, y hay un solo comando que escribe ficheros de texto con
nombre fijo, `biso snapshot`.** Sus garantías son otras y están en la sección 10.14: cada fichero se
escribe en un temporal y se renombra encima, los dos temporales se completan antes de renombrar
ninguno, y el comando no toma ningún acceso exclusivo, precisamente para que una copia no pueda hacer
fallar a la escritura de una tarea. Ninguna de las seis de aquí queda tocada por eso.

## Qué pasa con un dato que no se puede interpretar

Hay dos motivos distintos por los que un dato resulta imposible de leer, y con una base de datos son
dos casos que hay que separar: uno es que una tarea concreta esté dañada mientras el resto del
tablero sigue legible, y el otro es que no haya tablero legible en absoluto. El segundo no es una
variante del primero.

### El primer caso: una tarea ilegible

Una tarea puede resultar ilegible: la base de datos devuelve algo corrupto para esa fila, o la tarea
lleva una clave de extensión que la configuración ya no declara. El resto del tablero sigue legible, y
la regla depende del tipo de lectura:

| Tipo de lectura | Qué pasa |
|---|---|
| **Lectura dirigida** a esa tarea, es decir, `get`, o `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer` y `archive` con una referencia que resuelve a ella | Error 3, con el motivo exacto. No se escribe nada |
| **Lectura de conjunto**, es decir, `ls`, `prime`, `export`, `snapshot`, la resolución de una referencia por texto y cualquier filtro | La tarea se salta, se cuenta, y al final se emite `warning: 1 task could not be read and was skipped` con sus identificadores. El resto del resultado es válido y el código es 0, **salvo en `biso export` y en `biso snapshot`, que salen con 6** |
| `biso doctor` | Se reporta como problema y se sigue con las demás. Nunca aborta |

Una lectura de conjunto **nunca** aborta por una tarea mala, y **nunca** la esconde en silencio. Las
dos cosas juntas son lo que impide que un listado incompleto se confunda con un tablero vacío.

**`biso export` y `biso snapshot` son las dos excepciones al código 0 de una lectura de conjunto.**
Los dos escriben igual todo lo que han podido leer, con el mismo aviso por stderr, pero terminan con
**código 6** en vez de 0 cuando han saltado alguna tarea: son los dos comandos cuyo propósito es
servir de copia fiel del tablero, así que una copia incompleta no puede parecer un éxito llano. Un
guion que encadene `biso export -o backup.ndjson && ...` o `biso snapshot && ...` puede comprobar el
código de salida para detectar un volcado incompleto.

### El segundo caso: la base de datos que no se puede leer

Esto **no es una tarea ilegible: es que no hay tablero legible.** Si la base de datos no abre, o abre
pero falla su comprobación de integridad, ninguna tarea es alcanzable, así que no tiene sentido
presentarlo como "una tarea se ha saltado": no hay nada que saltar, hay un tablero entero fuera de
alcance. La regla es la misma para cualquier operación, sea una lectura dirigida, una de conjunto, una
escritura o `biso doctor`: el comando aborta entero, sin escribir nada, con este mensaje por stderr:

```
error: board 3f9a2b1c's database could not be read
hint: it did not open, or it failed its integrity check, and there is no automatic repair
hint: rebuild it in place with `biso init --from <snapshot dir>`, which keeps its id
```

El código de salida es **10** (`DAMAGED`, sección 2), y no el 8 de la ausencia de tablero, porque el
remedio es otro: aquí el tablero está donde tiene que estar y lo que hay que hacer es reconstruirlo, no
crearlo. La clave `code` del sobre JSON (sección 12.3) es `database_unreadable`.

**El código 10 puede salir de cualquier comando, y por eso no se repite en la tabla de códigos de
salida de cada uno.** Esas tablas dicen los desenlaces propios del comando; este no lo es de ninguno, es
el del tablero entero, igual que el 1 de un fallo del programa, que tampoco aparece en ellas.

**Y el remedio se puede teclear tal cual, porque el segundo `hint` nombra el comando que lo hace.** Un
directorio cuya base de datos no abre no cuenta como tablero accesible para `biso init`, así que
`biso init --from` reconstruye ahí mismo en vez de dar el error 2 de "ya hay uno" (10.1), y adopta el
`id` que nombra el marcador de la instantánea, de modo que el puntero commiteado del proyecto sigue
valiendo. Es también lo que necesita un clon recién traído a otra máquina, que llega con el directorio
del tablero versionado y sin base de datos dentro.

### Lo que se pierde al pasar de un fichero por tarea a una base de datos compartida

Con un almacén de un fichero por tarea, el aislamiento del daño sale gratis: una tarea corrupta es
exactamente eso, una tarea corrupta, y las demás siguen intactas porque viven en ficheros distintos.
Con una base de datos compartida, ese aislamiento hay que provocarlo a propósito, y cuando falta, una
sola corrupción puede llevarse por delante más de una tarea a la vez, o el tablero entero. Esta
especificación dice las cosas incómodas en voz alta en vez de esconderlas: ese es el coste real de
guardar los datos en una base de datos, frente a la alternativa de un fichero por tarea.

