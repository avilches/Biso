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

El mensaje de `biso prime` mide **4.689 bytes**, 3.255 de parte fija y 1.434 de resumen del tablero
(sección 9.5 de `docs/SPEC.md`), contra un tope duro de 5.120 repartido en dos mitades de 3.456 y
1.664.

| Magnitud | Herramienta estudiada | `biso` |
|---|---:|---|
| Peor caso por sesión, con ciclo completo | 12.905 bytes | 4.689 bytes |
| Media medida por sesión | 3.358 bytes | 4.689 bytes |
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
algo".** La medida de hoy, 4.689 de 5.120 bytes, tiene 431 de margen: nunca hizo falta apretar para
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
- **Todo lo relativo al control de versiones.** No porque el problema no exista: se midió que un
  commit automático por llamada convierte el ciclo de una tarea en siete commits, seis de ellos con el
  mensaje idéntico. Está fuera porque presupone que las tareas son ficheros versionados, y esa
  decisión está abierta. **Cuando se decida cómo persiste el sistema, este problema hay que volver a
  resolverlo**, y la forma la da el diseño actual sin tocar nada: si cerrar una tarea es una llamada,
  es una unidad de cambio, y su mensaje puede decir qué pasó.
- **La visibilidad entre versiones del proyecto.** Se midieron cinco fallos reales relacionados con
  copias de trabajo paralelas: dos "tarea no encontrada" sobre tareas que existían en otra rama, tres
  volcados de pila al intentar leer de una rama remota, y un error de identificador ambiguo. Resolver
  eso exige nombrar el sistema de control de versiones, así que **queda explícitamente sin resolver
  hasta que se decida la persistencia**. Lo que sí está, y es independiente del almacenamiento, son
  los tres mensajes distintos de "no la encuentro" y la garantía de que ninguna lectura de conjunto
  aborta por una tarea que no se puede leer.
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
porque ya tiene entrada propia en el vocabulario de la sección 2, porque da nombre al término
`urgency.blocked` de la fórmula de urgencia, y porque nombra el hecho que de verdad se calcula. Quien
quiera trabajo cogible pide las dos cosas, y así lo enseña el ejemplo de la ayuda:
`biso ls --not-blocked --not-waiting --ids`. El cambio quita una clave del JSON y renombra una bandera,
que son las dos cosas que el contrato de estabilidad de 13 promete no tocar nunca, y por eso se hace
ahora: ese contrato obliga a partir de la versión 1.0 y todavía no hay ninguna versión publicada.
Después de 1.0 esta limpieza ya no se podría hacer.

**10.6, por qué `set` no repite en su tabla las banderas de campo.** Porque repetirlas invitaría a que
divergieran, que es como se rompen los documentos largos.

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
de estados y se decidieron a la vez: dos quedan resueltos sin ningún papel de estado nuevo, uno queda
aplazado a la decisión de persistencia, y uno se retira. El criterio que ordenó las cuatro decisiones,
y que conviene aplicar la próxima vez que alguien proponga un papel de estado, está en el apartado 10.

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

**Esta depende de cómo se persistan los datos** y conviene tomarla con esa.

**Queda aplazada, y la forma ya está decidida.** Un instante de caducidad sobre una tarea activa y
asignada, que al vencer saca la tarea del estado activo sin tocar la asignación, porque lo que caduca
es "estoy en ello" y no "esto es tuyo". Lo que falta, el nombre y el tipo del campo, cómo se renueva
mientras se trabaja y quién detecta la caducidad sin que cueste caro, depende de cómo se persistan los
datos, y se anota en la sección 14 de `SPEC.md` para que no se dé por olvidado.

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
comentario no reescriba nada de lo demás. No se descarta: es una opción razonada para la decisión de
persistencia, que sigue abierta.

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
  el arrendamiento del apartado 9.2, que está aplazado.
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
  descripción de la bandera como el ejemplo de la ayuda de `biso ls`. El riesgo se queda pero encogido:
  el nombre ya no promete estar lista para trabajar, solo no estar bloqueada, que es lo que mide. El
  porqué del nombre está en el apartado 6.
