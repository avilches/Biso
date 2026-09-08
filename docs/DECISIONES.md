# Por qué `biso` es como es

`docs/SPEC.md` dice qué hace el programa y nunca por qué. Este documento es el complemento: la razón
de cada decisión que podría parecer arbitraria, y la evidencia que la sostiene.

**Sirve para una cosa concreta:** antes de cambiar una regla de la especificación, hay que mirar aquí
si esa regla existe por algo. Varias de ellas parecen caprichos de estilo y son la respuesta a un
fallo medido en herramientas reales.

Este documento usa las mismas palabras que la especificación y con el mismo significado; la tabla de
"Vocabulario de este documento" al principio de `docs/SPEC.md` es la referencia para las dos.

La evidencia viene de dos sitios. El primero es un estudio del uso real de un gestor de tareas por
agentes automáticos: 856 invocaciones de línea de comandos en 60 sesiones y 10 proyectos a lo largo
de seis días, con la salida de cada llamada medida en bytes. El segundo es la comparación de dos
gestores de tareas maduros, Backlog.md y Taskwarrior, y de los fallos documentados de ambos.

---

## 1. La evidencia detrás de los siete principios

Los siete principios de la sección 1 de la especificación están ahí enunciados sin su procedencia,
porque un principio se aplica igual se sepa o no de dónde viene. Aquí está de dónde viene cada uno.

**Principio 1, que un valor desconocido es un error al leer y al escribir.** Es el fallo más peligroso
que se midió. En la herramienta estudiada, `task list --status "Todo"` contesta `No tasks found.` con
código de salida 0, mientras que `task edit -s "Todo"` escribe correctamente el mismo valor. Un agente
que escribe mal un filtro no recibe un error: recibe una afirmación falsa sobre el tablero, y "no hay
tareas" es exactamente la clase de respuesta sobre la que se construye la siguiente decisión, incluido
informar al usuario de que algo no existe.

**Principio 2, que un nombre significa lo mismo en todos los comandos.** En la herramienta estudiada,
`--ref` y `--acceptance-criteria` añaden en el comando de creación y reemplazan en el de edición. Cada
comando por separado es defendible; juntos son una trampa, porque quien escribe una secuencia los mira
juntos.

**Principio 3, que el nombre desnudo añade.** Se midieron seis casos de agentes usando la variante
destructiva de forma repetida creyendo que añadían: la bandera de plan aplicada hasta cinco veces
sobre la misma tarea, la de referencias tres veces sobre otra, y un caso en el que un agente ejecutó
sobre una misma tarea `--ref`, `--ref`, `--ref`, `--add-ref`, `--remove-ref` y `--clear-refs`, que es
alguien probando a ver cuál de las seis hace lo que quiere. Ninguna de esas llamadas dio error, y el
daño es silencioso: cada una borró lo que había escrito la anterior.

**Principio 4, que la salida por defecto de una escritura es lo que quien llama no sabía.** De las 237
creaciones medidas, 215 llevaban una bandera que devolvía la ficha entera de la tarea recién creada, y
sumaron 179.369 bytes, casi la cuarta parte de toda la salida del estudio, sin dar el único dato que
el agente no tenía, que es el identificador. La mediana de una creación pasa de 154 bytes sin esa
bandera a 1.333 con ella, un factor de 8,7.

**Principio 5, que un gesto del flujo de trabajo es un comando.** De las 159 ediciones medidas, 112
cambian exactamente un campo, hay 82 pares de ediciones consecutivas sobre la misma tarea, y el ciclo
de vida típico cuesta entre seis y doce llamadas. **Y no era una limitación de la herramienta**: se
comprobó que una sola llamada aceptaba el cierre entero y funcionaba. La fragmentación venía de que
sus guías presentaban el trabajo como una lista numerada con un comando por paso.

**Principio 6, el lote con validación previa.** Una migración real creó 242 tareas de una en una, y
cuando el entorno bloqueó los comandos compuestos que hacían falta, los agentes acabaron escribiendo
85 ficheros de tarea a mano, 76 ediciones y 9 creaciones, que es exactamente lo que la herramienta
prohíbe en la instrucción que ella misma inyecta en cada proyecto.

**Principio 7, que la salida no depende del terminal.** La bandera `--plain` de la herramienta
estudiada tiene doble vida: en unos comandos apaga una interfaz interactiva y en otros enciende un
volcado completo. Medido fuera de un terminal, en los comandos de lectura no cambia un solo byte y
aparece 191 veces sin ningún efecto; en el de creación multiplica la salida por 3,3 y en el de edición
por 36.

---

## 2. La regla de coincidencia de vocabulario

Es la pieza que cierra el principio 1, y merece contarse entera porque tiene una trampa.

**Se eligió eliminar los separadores, no colapsarlos.** Normalizar es pasar a minúsculas, quitar
diacríticos y quitar espacios, guiones y guiones bajos; comparar es igualdad. Con esa regla, `todo`,
`To-Do`, `TO_DO` y `to do` son todos `To Do`.

La alternativa era colapsar cualquier tramo de separadores a un solo espacio, y **no funciona**:
colapsando, `To-Do` se convierte en `to do` y coincide, pero `todo` no tiene ningún separador que
colapsar, se queda en `todo`, y no coincide con `to do`. Es un error fácil de cometer al redactar la
regla en prosa, porque el ejemplo que uno escribe a continuación parece cierto y no lo es.

Se eligió eliminar por dos motivos. El primero es que es lo que hacen las herramientas existentes al
escribir, así que ningún texto que hoy vale deja de valer. El segundo es que produce una función más
simple de escribir y de probar.

**Por eso la regla está escrita en la especificación como pseudocódigo con pasos numerados y no como
prosa.** Escrita en prosa volvería a poder decir dos cosas a la vez, y ya lo hizo una vez.

Una consecuencia que conviene no perder: el ejemplo canónico de un valor inválido no puede ser `Todo`,
porque `Todo` es válido. En la especificación el ejemplo es `Pending`, que no existe en ningún
vocabulario.

---

## 3. El presupuesto del mensaje de arranque

`biso prime` sustituye por completo a las guías de instrucciones y a cualquier inyección de texto en
los ficheros de convenciones del proyecto. Ese diseño se toma de una medición concreta.

En la herramienta estudiada, las cuatro guías de instrucciones se leyeron 69 veces en seis días y
suman **201.500 bytes, el 27,5% de toda la salida** que la herramienta devolvió a los agentes, más que
sus dos comandos más usados juntos.

| Lectura | Bytes | Cuándo la exige ese diseño |
|---|---:|---|
| Guía general | 2.365 | al empezar |
| Guía de creación | 4.014 | antes de crear |
| Guía de ejecución | 3.807 | antes de planificar o actualizar |
| Guía de finalización | 2.719 | antes de terminar |
| **Ciclo completo** | **12.905** | una sesión que crea, trabaja y cierra |

A eso hay que sumar la inyección de instrucciones en el fichero de convenciones del repositorio, que
se paga en todas las sesiones aunque no se toque el tablero.

El mensaje de `biso prime` mide **4.746 bytes**, 3.255 de parte fija y 1.491 de resumen del tablero
(sección 9.5 de `docs/SPEC.md`), contra un tope duro de 5.120 repartido en dos mitades de 3.456 y
1.664.

| Magnitud | Herramienta estudiada | `biso` |
|---|---:|---|
| Peor caso por sesión, con ciclo completo | 12.905 bytes | 4.746 bytes |
| Media medida por sesión | 3.358 bytes | 4.746 bytes |
| Lecturas obligatorias por sesión | entre 1 y 4 | 1 |
| Contexto gastado en sesiones que no tocan tareas | la inyección en el fichero de convenciones | 0 |

**La media sube ligeramente, y conviene decirlo en vez de esconderlo.** Lo que cambia es otra cosa: el
coste pasa a ser fijo, conocido y acotado por una prueba, en vez de depender de cuántas guías decida
leer el agente, y el peor caso cae a menos de un tercio. El ahorro grande no está aquí, está en que la
salida de las escrituras deje de ser un eco y en las llamadas que desaparecen al fusionar el ciclo.

**El tope es una prueba de la suite, no un objetivo.** Y el reparto en dos mitades existe para que el
resumen del tablero, que crece con el tablero, no pueda comerse el sitio de las reglas.

**El reparto entre las dos mitades cambió con el modelo de estados, y el total no.** La parte fija
sube de 3.072 a 3.456 bytes porque el bloque `COMMANDS` gana las dos órdenes nuevas del ciclo, `ask` y
`answer`, y el bloque `RULES` gana una regla más, la undécima, sobre esos dos verbos y sobre qué
significa una tarea asignada; el resumen del tablero baja de 2.048 a 1.664.
Eso solo es seguro de hacer porque, a la vez, el orden de recorte del resumen deja de estar incompleto:
antes nombraba un solo bloque y decía "antes que cualquier otra cosa" sin nombrar ninguna otra, así que
un tablero con muchas tareas en curso podía rebasar el tope sin que hubiera una conducta definida para
ese caso. Con los cinco pasos completos que trae ahora la sección 9.5 de `SPEC.md`, el resumen ya no
crece sin límite, y darle menos sitio cuesta filas mostradas, no correcciones. La parte fija, en
cambio, no se puede recortar sola: o cabe entera o hay que quitar contenido a mano, así que es la mitad
que necesita más margen.

**El tope total, 5.120 bytes, no se mueve, y el motivo no es de contrato.** El contrato de estabilidad
solo obliga desde la versión 1.0, que todavía no está publicada, así que subir el tope no rompería
ninguna promesa hecha a nadie. La razón es de fondo: un tope que se sube cada vez que aprieta deja de
ser un tope, y su valor entero está en que obligue a elegir qué entra en el mensaje y qué se relega a
`--help`. Por eso, si al escribir el texto real de la sección 9.7 los números no cupieran, lo que se
recorta es contenido, no el tope.

**Esto no es "el tope nunca sube", es "el tope sube solo cuando reducir ya no es posible sin perder
algo".** La medida de hoy, 4.746 de 5.120 bytes, tiene 374 de margen: nunca hizo falta apretar para
caber, así que esta regla no se ha puesto a prueba todavía. Si en el futuro un comando nuevo obliga a
recortar el bloque fijo (`COMMANDS`, `FIELD FLAGS`, `RULES`) y esa reducción sale limpia, sin perder
información que un agente necesite para arrancar bien, es que había margen y el tope hizo su trabajo.
Pero si reducir más solo se puede ya a costa de quitar algo así, mantener el tope fijo deja de ser
disciplina y pasa a ser dañar el mensaje a propósito; en ese punto, subirlo es lo correcto.

---

## 4. Lo que se deja fuera, y por qué

- **Los hitos como entidad.** En el estudio, el comando de crear hitos se usó 22 veces, pero los
  comandos de documentos y de decisiones no se usaron ni una sola vez en seis días. `biso` conserva el
  hito como campo de texto libre de la tarea y no crea una entidad con ciclo de vida propio. Cuidado
  con una tentación concreta: declarar que el hito es texto libre y a la vez que la bandera valida
  contra "los hitos definidos" es contradictorio, porque nada puede llenar ese conjunto.
- **El servidor de integración.** En 1.280 transcripciones no hubo una sola llamada al servidor de
  herramientas que la herramienta estudiada ofrece, pese a estar disponible. Al analizarlo se vio que
  arregla buena parte de los errores de parámetros y **ninguno** de los problemas de granularidad. Las
  mejoras de nomenclatura que sí acierta, distinguir por el nombre lo que añade de lo que sustituye,
  están adoptadas en la línea de comandos de `biso`.
- **Los contextos de sesión**, es decir, filtros por defecto guardados que cambian lo que devuelve una
  consulta sin que se vea en la línea de comandos. Es la clase de estado invisible que hace que quien
  lee un listado saque conclusiones falsas, y contradice el principio 1. Se descarta a propósito y no
  por olvido.
- **El commit automático por cada llamada.** Se midió que convierte el ciclo de una tarea en siete
  commits, seis de ellos con el mensaje idéntico. Presuponía que las tareas eran ficheros versionados
  del propio proyecto, y esa premisa se descarta entera con la decisión de persistencia (sección 12):
  el tablero vive en un almacén aparte, y lo único que se versiona es la instantánea de texto que
  `biso snapshot` escribe cuando quien llama lo pide, un solo commit por invocación y nunca uno por
  cada escritura de tarea.
- **La visibilidad entre versiones del proyecto.** Se habían medido cinco fallos reales de las copias
  de trabajo paralelas de otras herramientas: dos "tarea no encontrada" sobre tareas que existían en
  otra rama, tres volcados de pila al leer de una rama remota, y un identificador ambiguo. La decisión
  de persistencia (sección 12) no los resuelve, los disuelve: el tablero no vive en el árbol de
  trabajo, así que una tarea cerrada está cerrada y no hay una rama de la que leerla ni una remota que
  le falte. Lo que queda, y sigue siendo independiente del almacenamiento, son los tres mensajes
  distintos de "no la encuentro" de la sección 7.3 de `SPEC.md` y la garantía de que ninguna lectura de
  conjunto aborta por una tarea que no se puede leer.
- **La sincronización con sistemas externos.** No está, pero sí están las cuatro piezas que la hacen
  posible, y esa es la única razón por la que existen: las claves declaradas de `ext` para guardar la
  identidad de la tarea en el otro sistema, el autor libre en los comentarios, las fechas fijables al
  importar y la simetría de `export` con `new --from`.
- **La clave de configuración `default_assignee`.** Habría asignado una persona a toda tarea creada
  sin `-a`. Se descarta porque en un tablero que la usara, absolutamente todo nacería asignado, y la
  asignación dejaría de significar que alguien decidió encargarte justo esa tarea: la consulta de
  arranque de un agente devolvería el backlog entero disfrazado de encargo. Es la comodidad concreta
  que habría destruido la señal en la que se apoya la decisión del apartado 9.1, que la asignación sea
  el gesto con el que una persona encarga trabajo.

---

## 5. Cuatro requisitos aprendidos de otras herramientas

Estos no salen del estudio de uso, sino de comparar Backlog.md con Taskwarrior y mirar qué falla en
cada uno.

**Ningún campo desconocido se acepta ni se descarta en silencio.** Las dos herramientas lo hacen mal:
Backlog.md borra el campo sin avisar en la siguiente reescritura del fichero, y Taskwarrior, ante un
atributo personalizado no declarado, en el peor caso se come la descripción entera de la tarea. La
regla correcta es rechazar con un error explícito, y de ahí salen las claves declaradas de `ext` y la
regla de que una tarea con una clave no declarada falla en una lectura dirigida en vez de perderse.

**Cualquier fecha se puede fijar al importar.** Ninguna de las dos lo permite desde su interfaz
pública. Backlog.md no tiene un solo parámetro de fecha, ni para la de creación ni para el instante de
un comentario, y por eso un importador de histórico no puede preservar las fechas reales usando su
línea de comandos, que es su única vía legítima de escritura. Un detalle que conviene recordar: dentro
de Backlog.md existe una función interna que sí acepta una fecha explícita para un comentario, pero no
está expuesta ni por su CLI, ni por su servidor, ni como librería.

**La exportación es literalmente el formato de importación.** Ninguna de las dos lo cumple: la
exportación de Backlog.md es un informe de solo lectura, y la de Taskwarrior mezcla datos reales con
derivados como el identificador de sesión y la urgencia sin separarlos. En `biso`, `export` escribe
todos los campos no derivados y ninguno derivado, y la ida y vuelta es una prueba de la suite.

Hay una trampa concreta en esto, y cuesta verla: **si el formato de lote admite los criterios de
aceptación como simples cadenas de texto, la simetría es falsa por construcción**, porque un criterio
tiene clave, texto y marca de cumplido, y al reimportar se pierden las claves y las marcas. Por eso el
formato admite objetos.

**Una funcionalidad no se retira sin anunciarla, y un cambio de formato lleva su camino de migración.**
Taskwarrior 3.0 retiró el historial de una de sus vistas sin aviso, y su cambio de motor de
almacenamiento provocó un caso real documentado de pérdida total de la base de tareas en Arch Linux,
porque el paquete se actualizó sin incluir el script de migración. De ahí sale el ciclo de aviso
obligatorio del contrato de estabilidad.

---

## 6. El porqué de reglas concretas

Cada entrada dice la sección de la especificación a la que corresponde.

**6.1, por qué no hay coincidencia por prefijo ni por parecido al resolver una referencia.** Una regla
que adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

**6.3, por qué las tareas archivadas cuentan en el conjunto de etiquetas y personas contra el que
validan los filtros.** Es lo que impide que un filtro que hoy funciona deje de funcionar mañana por
archivar la última tarea que lo usaba.

**6.3, por qué las etiquetas y las personas no tienen vocabulario cerrado al escribir, pero sus
filtros sí validan.** No tienen vocabulario cerrado porque su utilidad es que se puedan inventar sobre
la marcha. Y validar al leer no es una asimetría con la escritura: es la aplicación del principio 1,
que dice que un filtro que no puede encajar con nada es un error y no una respuesta vacía.

**7.3, por qué los tres mensajes de "no la encuentro" son distintos.** Porque las tres situaciones
piden acciones distintas de quien llama: corregir la sintaxis, dejar de buscar, o mirar en el archivo.

**8.1, por qué cada clase de campo tiene las variantes de bandera que tiene.** Las cuatro variantes
(`--campo`, `--set-campo`, `--rm-campo`, `--clear-campo`) existen para todo campo que guarde una
lista, porque sobre una lista se pueden hacer las cuatro operaciones. Un bloque de prosa no tiene
elementos que quitar de uno en uno, así que no tiene `rm-`. Un mapa de claves se manipula por clave y
no por posición, así que su "quitar" toma una clave. Un escalar solo se fija o se vacía. Y los
comentarios son una lista inmutable, de la que no se quita nada, así que solo admiten añadir. **La
tabla de clases de campo no es una lista de excepciones a la regla: es la regla aplicada a cada forma
de dato.**

**8.4, por qué quitar un criterio de aceptación toma un selector y no un texto.** Porque quitarlo por
su texto exacto es más frágil que quitarlo por su clave.

**8.6, por qué no existe una bandera que sustituya el mapa de campos externos entero.** Fijar una clave
ya es sustituir su valor, así que una segunda bandera para lo mismo solo serviría para equivocarse. Y
una que sustituyese el mapa entero con la sintaxis `clave=valor` sería una forma silenciosa de borrar
la identidad externa de una tarea al escribir otra.

**9.7, por qué el bloque de tareas en curso del mensaje de arranque no tiene límite.** Porque en un
tablero sano son pocas.

**10.1, por qué el puntero del proyecto es la única cosa que `init` escribe fuera del tablero.** Sin
ella, un tablero creado en otra ubicación no lo encontraría ningún comando posterior.

**10.1, por qué se retira la regla posicional que guardaba el estado activo como el penúltimo de
`--statuses`.** La regla estaba rota, y la contradicción que la delata vive en el propio documento: el
tablero de ejemplo era `Ideas, To Do, In Progress, Blocked, Done`, cuyo penúltimo es `Blocked`, y había
un ejemplo literal de `biso init` que lo creaba así, mientras que tanto la salida de `biso config list`
como el esquema JSON de `biso prime` declaraban que el estado activo de ese mismo tablero era
`In Progress`. Las dos cosas no podían ser ciertas a la vez, y la regla solo parecía funcionar porque
el tablero por defecto tenía justo tres estados. Se sustituye por tres banderas explícitas,
`--initial-status`, `--active-status` y `--terminal-status`, con el mismo argumento de 6.1: una regla
que adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

**10.1, por qué el tablero por defecto no trae un estado `Ideas`.** Un estado `Ideas` no dice nada que
no diga ya estar sin asignar, que se consulta con `biso ls --unassigned`. El matiz que sí aporta,
"esto quizá no lo hagamos nunca", tiene ya una decisión con evidencia detrás en el apartado 9.5: se
cubre con un tipo más del vocabulario que ya existe y no con un estado. Y hay un motivo peor para no
ponerlo por defecto: si `Ideas` fuera el estado inicial, toda tarea nueva nacería ahí, y el bloque
`NEXT UP` del mensaje de arranque mezclaría "algún día quizá" con "hay que hacerlo", que es justo la
distinción que ese bloque existe para hacer.

**10.3, por qué existe `--start` al crear una tarea.** Evita que crear una tarea para ponerse con ella
en el mismo minuto cueste dos llamadas. Es el principio 5 aplicado a un caso medido.

**10.3, por qué `--comment` funciona al crear.** Por lo mismo: una tarea que nace con un comentario
cuesta una llamada.

**10.4, por qué el filtro de etiquetas es el único que combina sus valores con "y".** Porque el uso
normal de varias etiquetas es acotar, no ampliar.

**10.4, por qué el filtro de dependencias es `--blocked` y `--not-blocked`, y no `--ready`.** El hecho
que se calcula es uno solo, que alguna dependencia esté sin terminar, así que se nombra una vez y su
negación se forma con el mismo prefijo que los otros dos pares booleanos de `biso ls`,
`--waiting`/`--not-waiting` y `--active`/`--not-active`. El nombre `ready` sobraba por dos motivos
distintos. El primero es que hacía viajar el mismo hecho dos veces en el JSON, como `ready` y como
`blocked`, y dos campos que dicen lo mismo acaban divergiendo. El segundo es que prometía más de lo
que cumplía: miraba solo dependencias, así que `biso ls --ready` devolvía también las tareas aparcadas
en una pregunta, que es justo lo que un agente no puede coger. De los dos nombres sobrevive `blocked`
porque ya tiene entrada propia en la tabla de vocabulario de `docs/SPEC.md`, porque da nombre al
término `urgency.blocked` de la fórmula de urgencia, y porque nombra el hecho que de verdad se calcula.
Y el nombre nuevo tampoco promete estar lista para trabajar, porque ninguna bandera sola puede: eso
son varios filtros, y cuántos depende de qué se busque. Descartar lo bloqueado y lo aparcado son dos,
`--not-blocked --not-waiting`; quien quiera además tarea sin empezar añade `--not-active`, que es el
filtro con el que el propio mensaje de arranque describe su bloque `NEXT UP` (9.7); y quien la quiera
sin dueño, `--unassigned`. El cambio quita una clave del JSON y renombra una bandera,
que son las dos cosas que el contrato de estabilidad de 13 promete no tocar nunca, y por eso se hace
ahora: ese contrato obliga a partir de la versión 1.0 y todavía no hay ninguna versión publicada.
Después de 1.0 esta limpieza ya no se podría hacer.

**10.6, por qué `set` no repite en su tabla las banderas de campo.** Porque repetirlas invitaría a que
divergieran, que es como se rompen los documentos largos.

**10.6, por qué la línea de estado encoge cuando la tarea no tiene criterios, en vez de imprimir un
guion como hace el listado.** Las dos salidas parecen contradecirse y no lo hacen, porque no son la
misma clase de cosa. El listado de 10.4 es una tabla: sus columnas se rellenan al ancho del valor más
largo de la llamada, así que una celda vacía tiene que ocupar su sitio o las filas de abajo se
descolocan, y para eso está el guion. La línea de estado sale una por tarea afectada, sin ancho
compartido y sin nada que alinear debajo, de modo que un hueco no descoloca nada y un guion solo
añadiría un símbolo más que interpretar. Quien quiera los contadores siempre, estén las listas vacías
o no, pide `--json`, que trae los cuatro como números.

**10.7, por qué `finish` avisa de los criterios sin marcar y no lo impide.** Un criterio puede haber
quedado obsoleto, y un comando que no deja cerrar empuja a rodearlo con `set`, que es como se aprende
a esquivar una herramienta. Para quien quiera la política dura está `--strict`.

**10.9, por qué `export` no hereda los valores por defecto de `ls`, y por qué sale con código 6 y no
con 0 cuando salta una tarea ilegible.** Porque exportar de más nunca hace daño y exportar de menos en
silencio arruina una copia de seguridad. Es el único comando cuyo propósito es no perder nada, y por
eso es la única excepción a la regla general de las lecturas de conjunto.

**10.13, por qué `help` funciona sin tablero.** Porque es lo primero que alguien ejecuta cuando algo
no va.

**4.8, por qué los campos de texto largo no se parten por comas.** Porque una coma dentro de una frase
es normal, y partir por ella convertiría una descripción en varias.

**5.3, por qué las fechas se pueden fijar al importar y no en el uso normal.** Sin esa excepción no se
puede importar el histórico de otro sistema conservando cuándo pasó cada cosa, que es el tercer
requisito de la sección 5 de este documento.

---

## 7. Dos decisiones de detalle que cuesta reconstruir

**Una tarea sin quien la reporte es válida.** El campo `reporter` toma la identidad configurada al
crear la tarea, y si no hay ninguna se queda vacío **sin avisar**. Es deliberadamente distinto de los
otros cinco sitios donde hace falta una identidad (tabla de la sección 3.1 de `docs/SPEC.md`): el
filtro `--mine` falla, la autoasignación de `start` avisa, y el autor de un comentario, `biso ask` y
`biso answer` son un error. La razón es que un tablero de una sola persona no tiene por qué
configurar su identidad solo para poder crear tareas.

**`biso export --json` es un error y no una bandera sin efecto.** `export` es el único comando cuya
salida ya es JSON sin pedirlo, en forma de un objeto por línea, mientras que `--json` significa el
sobre único que imprimen todos los demás. Son dos formas distintas, y aceptar la bandera en silencio
dejaría en duda cuál de las dos sale.

---

## 8. Una advertencia sobre cómo se mantiene la especificación

La especificación pasó por cuatro revisiones adversariales antes de darse por buena. El patrón de
fallo dominante, y con diferencia, fue siempre el mismo: **dos copias distantes de un mismo dato que
dejan de coincidir**. Una lista de campos que aparece en dos secciones, una cifra publicada en tres
sitios, un código de salida que está en la tabla de un comando pero no en su texto de ayuda.

De ahí salen tres costumbres que conviene mantener al editar:

1. **Cuando un dato tenga que aparecer en dos sitios, que uno remita al otro** en vez de repetirlo.
2. **Los ejemplos de salida se generan ejecutando el algoritmo, no se escriben a mano.** Los del
   listado y los del mensaje de arranque fallaron tres revisiones seguidas mientras se escribieron a
   mano, y dejaron de fallar en cuanto se generaron.
3. **Al corregir una afirmación, búscala en todo el documento** antes de darla por corregida.

Y una cuarta, sobre este documento en particular: la especificación no justifica sus decisiones, y esa
regla es fácil de romper sin darse cuenta. La justificación no solo se esconde en la prosa, también en
la estructura. Una tabla llegó a tener una columna titulada "Por qué" que sobrevivió a cuatro
revisiones, dos de ellas dedicadas expresamente a cazar justificaciones, porque todo el mundo buscaba
frases y esa vivía en una celda.

---

## 9. El modelo de estados: cuatro requisitos, cerrados

Salen de un diseño anterior de gestor de tareas que no llegó a escribirse, y de la evidencia que
aquel diseño recogió sobre la herramienta que usaba esta máquina antes. Los cuatro tocaban el modelo
de estados y se decidieron a la vez: dos quedan resueltos sin ningún papel de estado nuevo, uno se
resuelve con la enmienda de la decisión de persistencia (9.2), y uno se retira. El criterio que ordenó
las cuatro decisiones, y que conviene aplicar la próxima vez que alguien proponga un papel de estado,
está en el apartado 10.

### 9.1. Distinguir el encargo de la ejecución: resuelto sin estado nuevo

El requisito decía que, con los tres papeles de la especificación, el gesto de una persona que encarga
trabajo y el de un agente que lo coge **son el mismo dato**, y que hacía falta un cuarto papel de
estado para separarlos.

**El diagnóstico era falso, y la solución sale de corregirlo, no de añadir nada.** La persona ya
escribe un dato: `assignees`. El agente escribe otro: el estado activo, al ejecutar `biso start`. Son
dos datos distintos, escritos por dos actores distintos, y esa asimetría existía ya en la
especificación antes de este trabajo; nadie la había mirado como la respuesta al requisito. Y
aplicando el criterio del apartado 10, tampoco podía ser nunca un estado: "esto lo tiene que hacer un
agente" convive con cualquier punto del camino, porque una tarea puede estar recién creada, a medias,
o aparcada en una pregunta abierta y seguir siendo de quien se la asignaron.

**Esta lectura solo es cierta sin la clave de configuración `default_assignee`.** Esa clave habría
asignado una persona a toda tarea creada sin `-a`. En un tablero que la usara, absolutamente todo
nacería asignado, y la asignación dejaría de significar que alguien decidió encargarte justo esa
tarea: la consulta de arranque de un agente devolvería el backlog entero disfrazado de encargo. Se
retira la clave entera: su fila en la tabla de configuración, su comportamiento en `biso new`, su
aparición en el esquema JSON de `config list` y su línea de `biso config --help`. La autoasignación de
`biso start` se queda, porque no ensucia la señal: cuando `start` asigna a `me`, la tarea entra a la
vez en el estado activo, así que nunca queda en el conjunto de "asignada y sin empezar", que es el que
importa.

**Lo que esta decisión cuesta, para que no se descubra por sorpresa.** El caso de uso que motivó el
requisito era arrastrar una tarjeta desde el móvil para que un agente se ponga con una tarea. Con la
asignación como encargo, el gesto pasa a ser asignar a un miembro, que sigue siendo viable desde un
móvil pero ya no es un arrastre, así que la frase que describía el caso de uso queda anticuada y no
describe ya la herramienta. Cuando se especifique la sincronización con un sistema externo, mover una
tarjeta de columna en ese sistema no significará nada para `biso`, y esa decisión hay que tomarla
entonces a propósito, no por sorpresa. Y no hay forma de decir "esto es tuyo, pero todavía no": con la
asignación como única señal, asignar autoriza a empezar de inmediato, y quien necesite esa espera tiene
que no asignar hasta que toque, o usar una fecha límite.

### 9.2. Saber si alguien está trabajando de verdad

Una tarea que un agente coge antes de que su sesión muera se queda en el estado activo
indefinidamente, y nada lo detecta.

La evidencia es concreta. La herramienta que se usaba antes tenía un campo para apuntar quién está
trabajando en una tarea, y el tablero web lo pintaba, pero era una lista a la que se añade un nombre
al empezar y que se vacía al terminar: **sin latido, sin marca de tiempo y sin caducidad**, de modo
que una sesión que muere deja su nombre ahí para siempre. Y había un segundo fallo encima: solo lo
escribían su servidor de integración y su API, nunca su línea de comandos, aunque su propio mensaje
de arranque pedía al agente que mandara su nombre al cambiar de estado.

Hay además un dato medido que descarta la solución fácil: **el evento de fin de sesión de Claude Code
trae un motivo con cinco valores posibles y ninguno corresponde a un proceso muerto a lo bruto**, así
que un hook de cierre no puede ser la única señal de que una tarea quedó huérfana, porque el caso que
importa es justamente el que no lo dispara.

La forma conocida de resolverlo es un arrendamiento con caducidad: latido mientras se trabaja, y algo
que libere la tarea cuya sesión murió. Arrendar sería asignar con esa caducidad sobre una tarea ya
activa, y reclamarla sería la misma escritura que hoy hace `biso start`, así que la decisión de este
apartado sigue encajando con la del 9.1 aunque aquel, al final, no trajera ningún papel nuevo.

**Esta dependía de cómo se persistan los datos**, y se cerró con esa decisión tomada.

**Cerrada, con una enmienda a lo que este apartado decía antes.** La redacción aplazada decía que la
caducidad, al vencer, "saca la tarea del estado activo sin tocar la asignación". Esa frase no se
sostiene y es la que cede: `status` es un campo guardado, con un valor del vocabulario del tablero, y
un campo derivado no puede cambiar un campo guardado. Si nada escribe, la tarea sigue teniendo el
estado activo guardado por muy vencido que esté su arrendamiento.

Lo que se guarda es el instante de caducidad (`leaseExpiresAt`) y quién tiene el arrendamiento
(`leaseHolder`), los dos sin valor salvo en una tarea activa y asignada (sección 5 de `SPEC.md`). Lo
que caduca sigue siendo, como decía la redacción original, "estoy en ello" y no "esto es tuyo": el
campo derivado `leaseExpired` dice que el arrendamiento venció, pero el estado guardado no cambia
solo, nunca. Liberarlo es una escritura explícita, y sigue sin hacer falta un comando nuevo para eso:
es la misma que ya hace `biso start`, que reclama el arrendamiento vencido a favor de quien llama
comprobando quién lo tenía dentro de la misma transacción, para que el tenedor viejo no la recupere al
despertar (sección 10.7.1 de `SPEC.md`). Con un arrendamiento vivo de otra identidad, `biso start`
avisa y la coge igual: el mismo "avisa, no impide" que ya aplicaba a las dependencias sin terminar y a
la pregunta abierta.

**Por qué se enmienda en vez de reescribirse sin más.** La redacción aplazada no decía cómo se
liberaba una caducidad, y la lectura más directa de "saca la tarea del estado activo" es una
escritura diferida: que la siguiente escritura cualquiera, o un proceso de fondo, arrastrara el
saneamiento de las tareas vencidas. Esa vía se descarta explícitamente al tomar esta decisión, en la
sección 6 del diseño de persistencia, porque haría que un comando tocara tareas que no nombró, y
porque `biso prime`, que no escribe nunca, mostraría un estado que una escritura ajena y posterior
podría cambiar. No es que la decisión siempre hubiera sido la de hoy: es que la única forma de
sostenerla, al escribirla de verdad, obligaba a mover la contradicción con `status` a otro sitio en
vez de resolverla.

**Lo que aporta el estado del arte, mirado al cerrar esta decisión.** El patrón tiene nombre propio
fuera de aquí: un arrendamiento con caducidad, renovado por latido, para evitar la doble reclamación.
Y la pieza que le faltaba a la redacción aplazada, la que un artículo que formaliza el patrón señala
como la que cierra el agujero, es la comprobación del tenedor: sin ella, el tenedor viejo puede
despertar, escribir, y robar de vuelta una tarea que ya había reclamado otro. Aquí sale gratis, porque
el arrendamiento ya guarda quién lo tiene y la comprobación es comparar y sustituir dentro de una
transacción que ya existía por otro motivo (sección 4.10).

**Por qué el arrendamiento se exporta e importa como cualquier otro campo.** Al añadir los dos campos
guardados quedó sin decir si viajan en `biso export`, y las dos respuestas eran defendibles: dejarlos
fuera, porque un arrendamiento es la reserva de una sesión concreta en una máquina concreta, o dejarlos
entrar, porque son campos guardados y no derivados y la garantía de simetría de la sección 10.9 de
`SPEC.md` promete que todos ellos van y vuelven. Se eligió la segunda, y en rigor no era una elección
libre: el contrato de estabilidad de la sección 13 de `SPEC.md` ya enuncia esa simetría como una prueba
de la suite "sobre todos los campos no derivados", sin lista de excepciones, así que dejar fuera el
arrendamiento habría obligado a abrir una y a mantenerla, que es exactamente la clase de enumeración
que se desincroniza (apartado 8). La otra razón, la que hace que la primera no salga cara, es que el
diseño ya toleraba un arrendamiento ajeno: `leaseExpired` se recalcula contra el reloj de
quien lee, así que el arrendamiento que llega caducado sale caducado y `biso start` lo reclama, y el
que llega vivo a nombre de otra identidad produce el aviso de la sección 4.3 y nada más, porque `biso
start` avisa y coge la tarea igual. El caso que de verdad importa, restaurar un respaldo del propio
tablero, sale además mejor así: la tarea que estaba en marcha sigue constando en marcha y a nombre de
quien la llevaba, en vez de aparecer activa y sin dueño del arrendamiento. Lo único que hay que
custodiar es la invariante de que los dos campos solo tienen valor en una tarea activa y asignada, y se
custodia donde se custodia todo lo demás del lote: en la validación previa de `biso new --from`, que
rechaza el fichero entero antes de escribir nada.

### 9.3. Señalar lo que espera a una persona

Se puede configurar un estado tipo `Blocked`, pero es un estado más: `biso prime` no lo distingue, así
que una tarea parada esperando una decisión humana no se ve donde se mira.

La evidencia: en el tablero que se estudió había cuatro tareas paradas por una pregunta sin responder
y nada lo señalaba. Eran, además, preguntas abiertas disfrazadas de trabajo pendiente, porque una
decisión no tenía dónde vivir y la única forma de registrarla era crear una tarea.

Lo que falta es un papel que marque un estado como "espera a una persona", y que el mensaje de
arranque lo destaque en su propio bloque.

**Queda resuelta, y no con el papel que este apartado imaginaba.** Aplicando el criterio del apartado
10, una pregunta abierta puede detener una tarea en cualquier punto del camino, así que no podía ser
un estado: es el campo `question` de la sección 5.7 de `SPEC.md`, con su derivado `waiting`, los
verbos `biso ask` y `biso answer`, y el bloque `NEEDS ANSWER` del mensaje de arranque, que es
exactamente el bloque propio que este apartado pedía.

### 9.4. Distinguir terminar de descartar: retirado

Se retira, y conviene decir por qué el argumento que parecía bueno no lo era, para no repetir el error
si alguien vuelve a proponerlo.

**El argumento que no vale es que archivar y descartar son lo mismo.** No se sostiene contra lo que ya
dice `biso archive`: existe `--unarchive`, que devuelve la tarea al tablero con el estado que tenía, y
la ayuda del comando presenta el archivo como sacar tareas del tablero sin perderlas. Archivar es
reversible, así que archivar es aparcar, y aparcar no es descartar.

**Y la deducción que se apoyaba en él tampoco vale.** La idea era que una tarea archivada que nunca
llegó al estado terminal se pudiera etiquetar como abandonada. No se puede calcular, porque el modelo
guarda el estado actual de una tarea y no un histórico de sus estados anteriores. Con `biso start
--reopen` una tarea terminada vuelve al estado activo, y si se archivara desde ahí, esa deducción la
llamaría abandonada habiendo estado hecha.

**El argumento que sí vale es que a este requisito le falta la evidencia que los otros tres sí traen.**
El 9.1 corrige un diagnóstico sobre una asimetría real de la especificación, el 9.2 trae el fallo
medido de una herramienta que se usaba antes (un campo sin latido, sin marca de tiempo y sin
caducidad), y el 9.3 trae cuatro tareas paradas por una pregunta sin responder en el tablero que se
estudió. Este requisito dice que una tarea hecha y una abandonada "se confunden", sin un solo caso en
el que esa confusión haya costado algo. Un papel de estado que obligue a dar un motivo es barato de
añadir cuando haga falta y caro de quitar si sobra, así que se queda fuera hasta que aparezca un caso
real que lo pida, y entonces se engancha a `biso archive`. Se anota en la sección 14 de `SPEC.md`.

### 9.5. Lo que se miró de ese diseño anterior y se descarta

**Un modelo con entidades separadas para tarea, idea, aprendizaje y decisión**, cada una con sus
propios estados y campos. La necesidad que lo motivaba es real, porque sin sitio donde ponerlas las
decisiones se disfrazan de tareas. Pero hay un dato posterior que contradice la solución: en el
estudio de uso de una herramienta que sí tiene comandos dedicados para documentos y decisiones,
**no hubo una sola llamada a ninguno de ellos en seis días, sesenta sesiones y diez proyectos**. La
conclusión es que la necesidad se cubre mejor con un tipo más en el vocabulario que ya existe que con
una entidad y un comando propios.

**Una forma concreta de guardar los datos**, con un directorio por entidad para que añadir un
comentario no reescriba nada de lo demás. La decisión de persistencia (sección 12) no la adopta: una
base de datos SQLite da la misma propiedad, una escritura por tarea sin reescribir el tablero entero, y
además la transacción que un directorio de ficheros habría tenido que construir a mano.

### 9.6. Un contraste que conviene mirar antes de implementar

Aquel diseño se imponía una regla contraria a la de esta especificación: **los valores por defecto
viven completos en el binario, el comando de creación del tablero no escribe ninguna configuración, y
un repositorio normal no tiene fichero de configuración nunca**. Su argumento era evitar acabar con un
motor genérico que no sabe hacer nada solo y que obliga a configurar antes de escribir la primera
tarea.

Aquí `biso init` sí escribe la configuración. El argumento contrario es bueno y merece mirarse antes
de dar la decisión por hecha.

---

## 10. El criterio de estado frente a campo

Es la regla que ordenó las cuatro decisiones del apartado 9, y conviene tenerla escrita aparte porque
se va a volver a necesitar la próxima vez que alguien proponga un papel de estado:

> Algo es un estado cuando es excluyente con los demás y dice en qué punto del camino está la tarea.
> Es un campo cuando puede convivir con cualquier punto del camino.

La especificación ya la aplicaba sin decirla: `archived` es un campo y no un estado precisamente
porque una tarea archivada conserva el estado que tenía al archivarse. Lo que faltaba era el criterio
escrito, para no volver a meter en el vocabulario de estados algo que no es un punto del camino.
Aplicado a los cuatro requisitos del apartado 9, deja solo uno pidiendo de verdad algo excluyente y
ligado al camino, el 9.4, y ese es justo el que se retira por falta de evidencia; los otros tres son un
gesto que ya existía (9.1), una reserva con caducidad (9.2) o un campo que puede convivir con cualquier
estado (9.3).

**Por qué esta condición no se hace configurable.** Dejar que cada tablero declarase sus propios
estados como excluyentes o no destruiría la garantía en la que descansa el resto del modelo: que
`biso start`, `biso finish` y los filtros por papel (`--active`, `--not-active`) puedan asumir siempre
que una tarea está en exactamente un estado a la vez. Si esa garantía dependiera de la configuración de
cada tablero, todo comando tendría que consultarla antes de decidir qué significa "activa", y eso es
justo la clase de comportamiento que depende de dónde y con qué se ejecuta el programa que la
introducción de la especificación descarta. La condición es del modelo, no de un tablero concreto.

---

## 11. Riesgos conocidos y aceptados del modelo de estados

Se aceptan a propósito, y conviene anotar por qué en cada uno para no tropezar dos veces con lo mismo.

- **Dos agentes con la misma identidad ven la misma cola.** Dos sesiones con el mismo `BISO_ME` no se
  distinguen entre sí, y las dos podrían coger la misma tarea a la vez. Es exactamente lo que resuelve
  el arrendamiento del apartado 9.2: la segunda sesión ve el aviso de que el arrendamiento está vivo y
  a nombre de esa misma identidad, pero `biso start` avisa y la coge igual, así que dos sesiones que
  comparten identidad siguen pudiendo pisarse. El arrendamiento defiende de una sesión muerta, no de
  dos sesiones vivas con el mismo nombre.
- **Un tablero con la clave `me` configurada anula la distinción entre persona y agente.** La clave
  `me` gana sobre `BISO_ME`, así que en un tablero que la tenga puesta todo el mundo comparte
  identidad y `--mine` deja de significar nada. Un tablero compartido entre una persona y un agente
  tiene que dejar `me` sin configurar.
- **La persona no tiene un canal hacia el agente que se vea en el mensaje de arranque.** El agente
  pregunta con `biso ask` y la persona responde con `biso answer`, pero si la persona quiere decirle
  algo por iniciativa propia lo escribe en un comentario, y el mensaje de arranque no muestra
  comentarios. El agente lo ve al hacer `biso get`.
- **El listado no enseña el texto de la pregunta, solo dice qué tareas la tienen.** Es el precio de no
  meter texto largo en el listado, medido en el principio 4 de la sección 1: 215 fichas completas
  sumaron 179.369 bytes, casi la cuarta parte de la salida del estudio. Quien quiera leer la pregunta
  usa `biso get --section question`, o mira el mensaje de arranque, que sí la enseña.
- **Una pregunta abierta sobre una tarea terminada o archivada desaparece de la vista.** `biso finish`
  avisa pero no impide, los bloques del mensaje de arranque excluyen terminadas y archivadas, y
  `biso ls` excluye el estado terminal por defecto, así que `biso ls --waiting` no la encuentra sin
  `--any-status`. Se acepta porque la alternativa, impedir cerrar una tarea con una pregunta abierta,
  empujaría a rodear la herramienta, el mismo argumento que ya vale en el apartado 6 para `finish` y
  los criterios sin marcar.
- **El filtro de dependencias no excluye las tareas aparcadas.** `--not-blocked` mira solo
  dependencias, así que un agente que elija trabajo únicamente con esa bandera se lleva también las
  que esperan una respuesta. La consulta correcta añade `--not-waiting`, y así lo dicen tanto la
  descripción de la bandera como el ejemplo de la ayuda de `biso ls`. El riesgo se queda, pero
  encogido: el nombre ya no promete que la tarea esté lista para trabajar, solo que no esté bloqueada,
  que es lo que mide. El porqué del nombre está en el apartado 6.

---

## 12. La decisión de persistencia

La especificación dejaba deliberadamente abierto cómo se guardan los datos. La decisión es: un tablero
es una base de datos SQLite en un directorio propio fuera del proyecto, localizado por un fichero
puntero versionado en git (`.biso.json`, sección 3.2 de `SPEC.md`), con una exportación de texto que sí
se commitea para el historial (`biso snapshot`, sección 10.14 de `SPEC.md`). Sin daemon, y sin fusionar
nunca dos almacenes escritos por separado.

La evidencia detrás de cada pieza de esta decisión, con sus enlaces, está en
[`docs/ESTADO-DEL-ARTE.md`](ESTADO-DEL-ARTE.md), el inventario de las herramientas del espacio y el
catálogo de sus fallos. Lo que sigue aquí es el porqué de cada pieza, no la evidencia en bruto.

**Por qué el tablero no se versiona con el código.** Cualquier herramienta que guarde las tareas como
ficheros del árbol de trabajo hereda su peor propiedad: el estado se bifurca con la rama, así que una
incidencia cerrada en una rama vuelve a aparecer abierta al volver a la principal (sección 2 de
`docs/ESTADO-DEL-ARTE.md`). Sacar el tablero del árbol de trabajo no resuelve ese problema, lo disuelve:
una tarea cerrada está cerrada, no cerrada en esta rama, porque no hay una rama que la contenga. El
precio es que el tablero no viaja al clonar el proyecto en otra máquina, y se paga a propósito a cambio
de que el estado de una tarea sea uno solo.

**Por qué no hay daemon, y por qué el motivo es aritmético y no de gusto.** El coste dominante de una
invocación de `biso` es arrancar un proceso, no el trabajo que hace una vez arrancado: la sección 12 de
`docs/ESTADO-DEL-ARTE.md` mide el suelo del sistema en 5,2 milisegundos y el trabajo real de leer,
ordenar e imprimir 300 tareas en 8,7 milisegundos con un binario de Go. Un daemon solo puede ahorrar la
segunda cifra, no la primera, porque el cliente que hablaría con él por un socket es también un proceso y
paga el mismo suelo de arranque para lanzarse. Así que un daemon competiría por uno o dos milisegundos de
unos ocho, pagando a cambio una arquitectura entera: un proceso de fondo que hay que arrancar, vigilar y
matar, y que si se cuelga hace fallar también las lecturas que la sección 4.10 de `SPEC.md` promete que
nunca fallan por una escritura en curso. Beads tuvo uno, hacía una sola cosa, y se eliminó por completo
al cambiar de motor; quien lo reemplazó por algo más simple cuenta que se pasaba varias veces por semana
peleándose con él (sección 11 de `docs/ESTADO-DEL-ARTE.md`).

**Por qué el texto es una salida, y nunca un canal de vuelta.** `biso snapshot` escribe `tasks.ndjson` y
`config.json` para que el historial de git cuente lo que pasó y para que `biso init --from` pueda
reconstruir el tablero entero en otra máquina, pero nada dentro de `biso` vuelve a leer esos ficheros
como si fueran la verdad. Beads documenta por qué esa asimetría es obligatoria y no una elección
estética: su importación es solo de inserción y actualización, y no puede saber si un registro ausente
en el texto fue borrado a propósito o simplemente no se llegó a exportar (sección 10 de
`docs/ESTADO-DEL-ARTE.md`). Tratar el texto como una fuente además de como una salida reintroduce esa
ambigüedad en `biso`, así que no se hace nunca: la base de datos es la única verdad, y `export` y
`snapshot` son su proyección de solo lectura hacia fuera.

**Por qué nunca se sincroniza fusionando dos almacenes escritos por separado.** Es el sitio donde se
han estrellado todas las herramientas del espacio, de formas distintas pero con la misma raíz: dos
copias de trabajo que asignan el mismo identificador a tareas distintas (sección 1 de
`docs/ESTADO-DEL-ARTE.md`), un bloqueo de fichero que no cruza remotos de git (sección 4), y un fichero
de log que se fusiona por unión de líneas y resucita las que se habían borrado, porque una fusión de
texto concatena y solo quita duplicados exactos, sin razonar sobre qué falta ni por qué (sección 10). La
respuesta seria de ese último problema, sustituir el motor por uno con fusión a nivel de celda, es la que
tomó Beads, y es coherente pero cara. `biso` no la necesita porque no la tiene que resolver: un tablero
vive en una sola máquina y no hay una segunda copia escribible con la que fusionarse, así que la
comprobación de identidad de la sección 4.10 de `SPEC.md` basta y no hace falta un algoritmo de fusión.

**Por qué `--fix` no es una comodidad, sino el consentimiento.** Esta decisión añade a `biso doctor`
(sección 10.11 de `SPEC.md`) las dos comprobaciones que no existían antes de que hubiera una base de
datos real detrás del tablero: la integridad de esa base de datos, y el aviso de un sistema de ficheros
donde el modo WAL de SQLite no da las garantías de atomicidad que la sección 4.10 de `SPEC.md` exige.
Las dos son daño externo puro, porque nada dentro de `biso` corrompe su propia base de datos ni decide
en qué disco vive el tablero, y por eso ninguna de las dos es reparable ni con `--fix`: la integridad se
repara restaurando de una copia, fuera de `biso` por completo, y el sistema de ficheros no es algo que
la herramienta pueda cambiar. Que aparezcan justo con esta decisión confirma la regla que ya ordenaba el
resto de `doctor`, en vez de ponerla a prueba: `--fix` es el único sitio de todo `biso` donde quien
llama dice "te autorizo a escribir cosas que no te he pedido una por una", y solo entra ahí lo que de
verdad se puede arreglar sin decidir por alguien; lo que no, se reporta y se deja donde está.

**Por qué la búsqueda del puntero no frena en la raíz de un repositorio, ni sabe que git existe.** El
puntero se busca subiendo desde el directorio de trabajo, y lo primero que se escribió fue que el
recorrido paraba en la raíz del proyecto, detectada buscando un directorio `.git` hacia arriba. La idea
era proteger de que un proyecto encontrara el tablero de otro que lo contuviera. Se retira, y conviene
saber por qué, porque parece una protección gratis y no lo es.

El primer motivo es que hace que la resolución del tablero dependa de una herramienta que este mismo
documento declara opcional. La sección 10.14 de `SPEC.md` dice que git puede no estar instalado y que
un tablero funciona igual sin él, y a la vez el freno haría que la respuesta a cuál es mi tablero
saliera de si existe cierto directorio que crea git. El segundo es peor, porque no es de principios sino
de comportamiento observable: es `biso snapshot` quien convierte el directorio del tablero en un
repositorio, de forma perezosa y la primera vez que se ejecuta, así que con el freno puesto el mismo
comando en el mismo directorio contestaba una cosa antes de la primera instantánea y otra después, sin
que nadie hubiera cambiado ninguna configuración. El tercero es que el freno estaba mal escrito de una
forma que solo se ve al ir a implementarlo: en un worktree de git y en un submódulo, `.git` no es un
directorio sino un fichero, de modo que la raíz de un worktree nunca se habría detectado y el recorrido
se habría pasado de largo hasta el repositorio que lo contiene. Como en esta máquina los worktrees viven
dentro del propio checkout, el efecto habría sido que un `biso` ejecutado en un worktree resolviera el
puntero del checkout principal.

Y el cuarto motivo es que el freno no protegía de lo que decía proteger. El caso que preocupaba es un
proyecto anidado en otro que sí tiene tablero, y ahí el puntero del proyecto de fuera está en el mismo
directorio donde el freno habría parado, así que el freno llega tarde y el puntero se hereda igual. La
única regla que de verdad lo evitaría es respetar el `.gitignore` del proyecto de fuera, porque en el
caso real que motivó la discusión ese fichero excluye la carpeta del proyecto de dentro. Eso está
descartado por el presupuesto de la sección 4.13 de `SPEC.md`: interpretar un `.gitignore` de verdad no
se puede reimplementar de forma fiable, y preguntárselo a git cuesta 12 milisegundos medidos de los 25
que hay para todo.

**Lo que se acepta a cambio, dicho claro.** Un proyecto sin puntero propio hereda el del proyecto que lo
contenga, si lo hay. Se acepta porque es una situación poco frecuente, porque cuando ocurre lo razonable
es que se note en vez de que se adivine, y porque tiene un remedio que no necesita ninguna regla nueva:
mover el tablero de fuera a un subdirectorio que no esté en la línea de subida, que es exactamente lo
que hubo que hacer con Backlog.md en esta máquina por el mismo motivo. Que se note es cosa de dos
salidas que ya existen: el bloque `BOARD` del mensaje de arranque dice el nombre del tablero en su
primera línea, así que un agente que empieza la sesión con `biso prime` ve enseguida cuál le ha tocado,
y `biso where` dice en su fila `source` el directorio concreto del que salió el puntero. Se descartó
emitir además una nota en cada comando: saldría también en el caso normal de trabajar desde un
subdirectorio del propio proyecto, que es la inmensa mayoría de las llamadas, y una nota que sale
siempre enseña a ignorarla.

**Por qué estar dentro del directorio del tablero es una vía de resolución y no un error.** Antes de
decidir lo de arriba se consideró lo contrario, que un comando ejecutado dentro del directorio de un
tablero fallara siempre, con el argumento de que ese directorio es almacenamiento y no un proyecto. Se
descarta porque ahí no falta ningún dato: la configuración completa de un tablero, con su nombre y su
`task_prefix`, vive dentro de su propia base de datos, y el puntero solo sirve para encontrar un tablero,
cosa que quien ya está dentro de él no necesita. Un directorio que contiene `board.db` es ese tablero y
no puede ser otro, así que no hay nada que adivinar, y es un hecho más específico que cualquier puntero
heredado, de donde sale que gane al puntero en el orden de la sección 3.2. Esta vía se lleva por delante
el otro motivo que tenía el freno de git para existir, porque el caso que hacía falta proteger ahora se
resuelve solo.

**Por qué el nombre `board.db` es interfaz y no un detalle interno.** La vía de arriba necesita
reconocer un directorio de tablero, y sin un nombre declarado no habría una sola forma de hacerlo. El
documento hablaba del fichero de la base de datos sin nombrarlo nunca, lo que bastaba mientras nada
dependiera de reconocerlo desde fuera.

**Por qué la identidad del tablero vive en un fichero y no en el nombre de su carpeta.** El diseño
anterior decía dos cosas que no podían ser verdad a la vez: que el nombre de la carpeta del tablero era
decorativo y que nadie resolvía por él, y que localizar un tablero desde su puntero era buscar el patrón
`*-<id>` en la raíz por defecto. Lo segundo es resolver por el nombre. De esa contradicción salían tres
fallos, y los tres se cerraron de golpe sacando el identificador del nombre y metiéndolo en un fichero
marcador, `<id>.id`, dentro del directorio del tablero.

El primero: `biso config set project_name` movía la carpeta para mantener el slug al día, y prometía no
tocar ningún puntero "porque el tablero se localiza por el patrón". Eso solo valía para los tableros de
la raíz por defecto, que es el único sitio donde se buscaba el patrón. Un tablero en cualquier otra parte
se quedaba con un puntero que nombraba una carpeta ya inexistente y sin ninguna red debajo. El segundo:
la comprobación de `biso doctor` que avisaba cuando el nombre de la carpeta no coincidía con el slug
denunciaba como problema un tablero perfectamente sano creado con `--at`, cuyo nombre de carpeta lo había
elegido quien llamaba, y `--fix` lo "reparaba" renombrando la carpeta, con lo que rompía el puntero que
`init` acababa de escribir. El tercero: `boards_extra_roots` estaba declarada en la configuración de
máquina y ninguna regla la leía, porque la búsqueda nombraba solo `boards_root`.

Con el marcador, el nombre de la carpeta es decorativo de verdad. Renombrarla no rompe nada, cambiar el
nombre del tablero **no toca el sistema de ficheros en absoluto**, la comprobación de `doctor` sobre el
nombre desaparece porque ya no hay nada que comprobar, y la búsqueda por identificador puede recorrer
cualquier raíz mirando marcadores, lo que da sentido a las raíces adicionales. **El premio grande es lo
que se quita**: renombrar un tablero era la única operación del programa que mezclaba una transacción de
SQLite con un movimiento en el sistema de ficheros, cosa que no puede ser atómica, y por eso arrastraba
un orden declarado, un estado intermedio observable y un código de error propio. Nada de eso existe ya.

El coste es que buscar un tablero por su identificador pasa de leer un nombre de carpeta a mirar dentro
de cada carpeta de cada raíz. Sigue sin abrir ninguna base de datos, así que son lecturas de directorio,
pero con veinte tableros son veintiuna en vez de una, y hay que medirlo contra el presupuesto de la
sección 4.13 cuando el programa exista. Se descartó a propósito el atajo de buscar primero por el nombre
y caer al marcador solo si falla: sería más rápido y volvería a poner el mismo dato en dos sitios, que es
lo que se acaba de quitar.

**Por qué la `path` del puntero puede ser relativa, y por qué la forma la elige quien llama.** Un puntero
que dice `/Users/avilches/Hub/Projects/Kex/tablero` solo resuelve en el ordenador donde el proyecto está
en esa ruta exacta, y el puntero se versiona precisamente para que viaje. Con `tablero` guardado como
ruta relativa, mover el proyecto entero con su tablero dentro no rompe nada.

La ruta relativa **se resuelve contra el directorio que contiene el fichero puntero, nunca contra el
directorio de trabajo.** El puntero se busca subiendo desde el directorio de trabajo, así que el mismo
fichero se lee desde cualquier subdirectorio del proyecto: resolver contra el directorio de trabajo haría
que el mismo puntero nombrara un tablero distinto por cada subdirectorio desde el que se llamara, y que
casi ninguno de ellos existiera. Eso rompería el principio de la sección 1 de `SPEC.md` de que ningún
comportamiento depende de dónde se ejecute el programa.

**La primera redacción de esto se apoyaba en un escenario que la desmentía, y conviene dejarlo escrito
para no repetirlo.** Decía que una ruta absoluta no resuelve "ni siquiera en un worktree del mismo
proyecto", y es al revés. Un worktree de git recibe el puntero, porque está versionado, y no recibe el
directorio del tablero, porque el propio `biso init` recomienda ignorarlo y los ficheros ignorados no se
comparten entre árboles de trabajo. Así que en un worktree la ruta absoluta sigue apuntando al tablero
real y la relativa apunta dentro del worktree, donde no hay nada. El cambio empeoraba justo el caso con
el que se justificaba. Y el desenlace era peor que el fallo: el mensaje que sale entonces invita a
ejecutar `biso init` para adoptar ese identificador, lo que desde un worktree habría creado un segundo
tablero con el mismo `id`, que es lo único que esta persistencia promete no permitir nunca. El error de
razonamiento fue tratar un worktree como el proyecto en otra ruta: lo es para lo versionado, pero el
tablero no está versionado y sigue existiendo en un solo sitio del disco.

De ahí salen las dos piezas que lo arreglan. Una es que la búsqueda pruebe una `path` relativa que no
resuelve contra cada ancestro del directorio del puntero, que es lo que encuentra el tablero del proyecto
desde un worktree que viva dentro de él. La otra es que **la forma de la ruta la elija quien llama**, con
la forma que le dé a `--at`. Se había descartado dar esa elección con el argumento de que quien llama no
tiene ningún dato que el programa no tenga, y ese argumento era falso: el dato que decide es dónde van a
vivir las demás copias de trabajo del proyecto, y eso solo lo sabe una persona. Con los worktrees dentro
del proyecto, que es la convención de esta máquina, la relativa es la buena; con los worktrees fuera del
proyecto, la única que resuelve es la absoluta, porque el directorio del tablero no está ni en la copia ni
en ninguno de sus ancestros. No hace falta ninguna bandera nueva para ofrecer esa elección, porque la
forma de `--at` ya la expresa.

**Por qué el tope de la búsqueda es el directorio personal y no un número de niveles.** La primera
redacción decía que el recorrido no comprueba ningún directorio con menos de dos componentes de ruta, y
funcionaba, pero por casualidad: acierta solo mientras el directorio personal tenga esa profundidad. Con
un directorio personal en `/root`, que es el del superusuario, el tope habría caído por debajo de él y un
puntero puesto ahí no se habría leído nunca. La regla dice ahora lo que quiere decir, que el recorrido no
sale del directorio personal de quien llama, y guarda el número solo para el caso en que no haya
directorio personal que lo exprese, con el trabajo fuera de la home o sin `HOME` definido. El primer
componente de una ruta absoluta es siempre un directorio del sistema o el contenedor de los directorios
personales de todo el mundo, así que un puntero ahí no puede estar a propósito.

---

## 13. El origen de la cifra de 25 milisegundos

El tope de bytes del mensaje de arranque (apartado 3) trae su medida. El presupuesto de arranque de la
sección 4.13 de `SPEC.md`, 25 milisegundos de reloj para `biso ls` y `biso prime` sobre un tablero de
300 tareas, no la tenía escrita en ningún sitio, y esta sección es esa medida.

**Medido en la misma máquina que documenta `docs/ESTADO-DEL-ARTE.md`** (Apple M3 Max, macOS 26.5.2, 300
iteraciones, sección 12 de ese documento). El suelo del sistema operativo para arrancar cualquier
proceso, sin ejecutar ninguna línea propia todavía, es **5,2 milisegundos**. Un binario de Go añade
**2,2 milisegundos** encima de ese suelo. Y un programa en Go que lee 300 tareas, las ordena y las
imprime tardó **8,7 milisegundos en total**, suelo, arranque de Go y trabajo real incluidos. La cifra
del presupuesto, 25 milisegundos, deja **unas tres veces de margen** sobre ese total medido.

**La cifra excluye a propósito los lenguajes interpretados.** En la misma máquina, el solo arranque de
Python 3.14 añade 24,5 milisegundos por delante de cualquier trabajo real, y el de Node 25.6 añade 33.
Un presupuesto que tuviera que cubrir ese arranque dejaría de medir la herramienta y pasaría a medir el
lenguaje, así que la cifra se fija mirando el suelo que un lenguaje compilado permite. Eso acotó la
lista de candidatos a los compilados, y el apartado siguiente cierra la elección dentro de esa lista.

---

## 14. El lenguaje de implementación es Go

**Los dos candidatos reales eran Go y Rust**, y la elección es Go. Los dos cumplen el presupuesto del
apartado anterior con holgura, así que la decisión no se toma por rendimiento: se toma por lo que cuesta
escribir el programa, porque es lo único que de verdad los separa aquí.

**Por rendimiento la diferencia son seis décimas de milisegundo.** Con las medidas de la sección 12 de
`docs/ESTADO-DEL-ARTE.md`, sobre el suelo de 5,2 milisegundos que cuesta arrancar cualquier proceso,
Rust añade 1,6 milisegundos y Go 2,2. Esa diferencia es el **2,4 por ciento** de un presupuesto de 25
milisegundos que ya sobra tres veces sobre el total medido de 8,7. Para calibrar cuánto es: `rg`, que es
la herramienta más rápida de las que se midieron instaladas, tarda 7,1 milisegundos, y
`git --version` tarda 12,3. Ganar seis décimas en un programa cuyo competidor de referencia gasta doce
milisegundos en imprimir su propia versión no cambia nada que un usuario pueda notar. El presupuesto,
además, **se midió con un binario de Go**, así que la cifra que la especificación exige no es una
extrapolación: es lo que el lenguaje elegido hizo en esa máquina.

**Lo que decide es el ciclo de desarrollo, y en particular el ciclo de un agente.** Este documento y
`SPEC.md` están escritos para que alguien implemente el programa entero sin preguntar, con la
comprobación frecuente que la sección 15 de `SPEC.md` ordena, y ese alguien va a ser en buena parte un
agente automático. En ese modo de trabajo, el coste dominante no es el tiempo de ejecución del programa
sino **el número de vueltas entre escribir y ver el resultado**, y ahí Go gana por dos motivos
distintos. El primero es que compila muy rápido, así que cada vuelta es corta. El segundo es más
importante: el modelo de propiedad y préstamo de memoria de Rust es exactamente la clase de cosa que un
modelo de lenguaje **no acierta a la primera**, y cada fallo obliga a una ronda de corrección que no
trata del problema que se está resolviendo sino de satisfacer al compilador. Esas rondas se suman, hacen
falta revisiones constantes, y no dejan nada mejor en el producto: el mismo programa, escrito en Go, no
las paga.

**El coste de elegir Go se acepta a ojos vistas y aquí queda escrito.** El recolector de basura, que es
la objeción habitual, no tiene ningún efecto medible en una herramienta cuyo proceso vive milisegundos,
lee 300 tareas y termina, porque no llega a haber presión de memoria que recoger. El binario es mayor
que el equivalente en Rust, lo que da igual en algo que se instala una vez. Y se renuncia a las seis
décimas del párrafo anterior, que es lo que se está comprando.

**Queda una cosa por comprobar al empezar a implementar, y no es del lenguaje sino de su encuentro con
SQLite.** La persistencia del apartado 12 es una base de datos SQLite, y en Go hay dos formas de hablar
con ella: un enlace con la biblioteca en C, que obliga a compilar con `cgo` y complica generar binarios
para otras plataformas, o una traducción de SQLite a Go puro, que compila en cualquier sitio sin
herramientas de C. La elección entre las dos afecta al arranque y a cómo se distribuye el programa, así
que **hay que medirla contra el presupuesto de 25 milisegundos antes de comprometerse**, y no está
medida todavía. Es lo primero que la implementación tiene que resolver, y su resultado pertenece a este
mismo apartado cuando se sepa.
