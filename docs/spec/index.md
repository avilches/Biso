# `biso`: especificación del CLI de gestión de tareas

Esta especificación define `biso`, una herramienta de línea de comandos para llevar las tareas de un
proyecto. Sirve a la vez de documentación de referencia y de base para implementar el programa desde
cero sin preguntar nada: cada comando trae su firma, su tabla de parámetros, su comportamiento en los
casos límite, la salida literal que imprime, su esquema JSON, sus códigos de salida y el texto exacto
de su ayuda. Esta página es su portada y su guía de lectura.

El destinatario principal de `biso` es un agente automático que trabaja dentro del proyecto. La
salida es predecible, los errores son distinguibles por su código sin leer el mensaje, y ningún
comportamiento depende de dónde se ejecute el programa.

**Qué define esta especificación y qué no.** Define la interfaz del programa y el modelo de datos
lógico de una tarea, no el porqué de cómo se guardan los datos: esa decisión, con su razonamiento y su
evidencia, vive en ["Decisiones de diseño"](../decisiones/index.md), y ninguno de estos documentos la repite.
Aquí aparece el mecanismo solo donde afecta al comportamiento observable, como el fichero puntero
`.biso.json` de [Cómo se elige el tablero](resolucion-del-tablero.md) o la base de datos SQLite que
[`biso doctor`](cmd/doctor.md) comprueba; donde no lo afecta, la especificación enuncia el requisito
(por ejemplo, que dos procesos simultáneos no puedan asignar el mismo identificador) y deja el resto
del mecanismo fuera de aquí.

**El modelo de tarea se comprobó, campo a campo, frente a Backlog.md y frente a otros seis gestores
de tareas e incidencias, y ninguno alcanza compatibilidad completa.** El más cercano es Backlog.md,
que cubre una parte alta de los campos sin perder información en ningún sentido, pero pierde
precisamente la clave estable de los criterios y de los comentarios, y ninguno de los siete tiene
nada parecido a `leaseExpiresAt`/`leaseHolder` ni a `question`. La comparación completa, con lo que
se pierde en cada sentido para cada sistema, está en ["Compatibilidad del modelo de datos con otros
gestores"](../estado-del-arte/compatibilidad-de-modelos.md). **La compatibilidad que sí existe es de
modelo de datos y no de formato de fichero**: `biso` no lee ni escribe los ficheros Markdown de
ninguna de esas herramientas, y no hay ninguna intención de que lo haga, como consta en [Lo que se
deja fuera a propósito](fuera-de-alcance.md).

## Cómo está organizada

### Los fundamentos

[Vocabulario de esta especificación](vocabulario.md)
fija los nombres de cada elemento del programa y cómo se llaman en estos documentos.
[Los principios](principios.md) son las reglas de las que el resto de la especificación es
consecuencia, y explican por qué los comandos se comportan como se comportan. [Códigos de salida](codigos-de-salida.md)
es la tabla global a la que apela cualquier comportamiento en un caso límite: sin haberla visto, una
referencia a un código concreto en otra página no dice nada.

### Cómo se invoca el programa

[Flags globales](cmd/flags-globales.md), [Entorno y configuración de
máquina](invocacion.md) y [Cómo se elige el tablero](resolucion-del-tablero.md) explican qué pasa
antes de que el programa llegue a interpretar el comando: qué flags valen para todos, y con qué
tablero va a trabajar. [Terminal, flujos de salida y codificación](salida-y-terminal.md) y [Cómo se
pasa un valor](valores-de-entrada.md) terminan ese bloque: qué imprime el programa según haya terminal
de por medio o no, y de qué formas se le puede pasar el valor de cualquier flag. Ninguna de estas
páginas depende de un comando concreto.

### El modelo de datos

[Orden de escritura, concurrencia y datos
dañados](garantias.md) dice qué garantiza el programa antes de que aparezca ningún campo. [El modelo de
datos de una tarea](modelo-de-datos/index.md) es la lista de esos campos, y [El arrendamiento de una
tarea](lease.md) desarrolla aparte las reglas de `leaseExpiresAt` y `leaseHolder`, que dominan la
mutabilidad de una tarea mucho más que cualquier otro campo. [Los presupuestos de arranque y de
tamaño](presupuestos.md) son los límites que ese modelo y el arranque del programa tienen que respetar.
[Los vocabularios del tablero y la regla de validación](vocabularios.md) cierra el bloque explicando
cómo se valida el valor de los campos que tienen un vocabulario cerrado.

### La gramática de la entrada que comparten los comandos

[Cómo se resuelve una referencia a una
tarea](referencias.md) y [Las familias de flags](familias-de-flags.md) definen de una vez la
forma que tiene nombrar una tarea y la forma que tiene cada flag de escritura. Estas reglas valen
para todos los comandos y no se repiten en cada uno: un comando solo las menciona cuando se aparta de
ellas, y ninguno lo hace salvo donde se diga.

### Los comandos

[Los comandos](cmd/index.md) explica cómo se agrupan y cuáles
aparecen en `biso --help` frente a `biso help all`. El primero de la lista es [`biso prime`](cmd/prime.md),
el comando con el que arranca una sesión y que resume el estado del tablero; el resto de los comandos
sigue después, en el orden de esa página.

### Los contratos y lo que queda fuera

[El contrato JSON](contrato-json.md) y [El contrato de
estabilidad](estabilidad.md) dicen qué forma no cambia mientras la versión mayor sea `1`.
[Lo que se deja fuera a propósito](fuera-de-alcance.md) nombra lo que la especificación decide no
tener, para que nadie lo dé por olvidado.
