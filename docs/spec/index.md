# `biso`: especificación del CLI de gestión de tareas

Esta especificación define `biso`, una herramienta de línea de comandos para llevar las tareas de un
proyecto. Está escrita para que alguien implemente el programa entero a partir de ella sin preguntar
nada: cada comando trae su firma, su tabla de parámetros, su comportamiento en los casos límite, la
salida literal que imprime, su esquema JSON, sus códigos de salida y el texto exacto de su ayuda.
Antes estaba en un solo fichero; ahora está repartida en los documentos de abajo, y esta página es su
portada y su guía de lectura.

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

El modelo de tarea es compatible con el de Backlog.md, de modo que se puede importar y exportar entre
las dos herramientas sin perder campos. **La compatibilidad es de modelo de datos y no de formato de
fichero**: los campos se corresponden uno a uno, pero `biso` no lee ni escribe los ficheros Markdown de
esa herramienta, y no hay ninguna intención de que lo haga, como consta en
[Lo que se deja fuera a propósito](fuera-de-alcance.md).

**Convención de idioma.** La prosa de estos documentos va en español. Todo lo que es interfaz del
programa (nombres de comando, flags, textos de ayuda, mensajes de error, claves JSON y claves de
configuración) va en inglés, porque es lo que la persona o el agente que usa el programa lee y
escribe. **El contenido de los ejemplos va también en inglés**, aunque la prosa que los rodea siga en
español: el título, la descripción, los criterios de aceptación, la definición de hecho, el plan, las
notas, el resumen final, los comentarios y la pregunta abierta de cualquier tarea de ejemplo (MYP-11
y las demás que aparecen en los documentos de `cmd/`) se escriben en inglés, por el mismo motivo que
el resto de la interfaz: es contenido que en un tablero real escribiría la persona o el agente que usa
el programa, no prosa de la especificación.

## Por dónde empezar y en qué orden

### Los fundamentos, antes que cualquier regla concreta

[Vocabulario de esta especificación](vocabulario.md)
fija los nombres en español que usa el resto de los documentos y su equivalente en la interfaz.
[Los principios](principios.md) son las reglas de las que el resto de la especificación es
consecuencia, así que conviene tenerlas presentes antes de leer un comando. [Códigos de salida](codigos-de-salida.md)
es la tabla global a la que apela cualquier comportamiento en un caso límite: sin haberla visto, una
referencia a un código concreto en otra página no dice nada.

### Cómo se invoca el programa, antes de tocar ningún dato

[Flags globales](cmd/flags-globales.md), [Entorno y configuración de
máquina](invocacion.md) y [Cómo se elige el tablero](resolucion-del-tablero.md) explican qué pasa
antes de que el programa llegue a interpretar el comando: qué flags valen para todos, y con qué
tablero va a trabajar. [Terminal, flujos de salida y codificación](salida-y-terminal.md) y [Cómo se
pasa un valor](valores-de-entrada.md) terminan ese bloque: qué imprime el programa según haya terminal
de por medio o no, y de qué formas se le puede pasar el valor de cualquier flag. Ninguna de estas
páginas depende de un comando concreto.

### El modelo de datos, con las reglas que lo escriben

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

### Los comandos, empezando por `prime`

[Los comandos](cmd/index.md) explica cómo se agrupan y cuáles
aparecen en `biso --help` frente a `biso help all`. El primero de la lista es [`biso prime`](cmd/prime.md),
el comando con el que arranca una sesión y que resume el estado del tablero; el resto de los comandos
sigue después, en el orden de esa página.

### Al final, los contratos y lo que falta

[El contrato JSON](contrato-json.md) y [El contrato de
estabilidad](estabilidad.md) dicen qué forma no cambia mientras la versión mayor sea `1`.
[Lo que se deja fuera a propósito](fuera-de-alcance.md) nombra lo que la especificación decide no
tener, para que nadie lo dé por olvidado. Y [Por dónde empezar a implementar](por-donde-empezar.md)
remata la especificación con el orden en que cada pieza del programa paga lo que cuesta.
