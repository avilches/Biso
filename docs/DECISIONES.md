# Por qué `biso` es como es

`docs/SPEC.md` dice qué hace el programa y nunca por qué. Este documento es el complemento: la razón
de cada decisión que podría parecer arbitraria, y la evidencia que la sostiene.

**Sirve para una cosa concreta:** antes de cambiar una regla de la especificación, hay que mirar aquí
si esa regla existe por algo. Varias de ellas parecen caprichos de estilo y son la respuesta a un
fallo medido en herramientas reales.

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

El mensaje de `biso prime` mide **4.062 bytes**, 2.963 de parte fija y 1.099 de resumen del tablero,
contra un tope duro de 5.120 repartido en dos mitades de 3.072 y 2.048.

| Magnitud | Herramienta estudiada | `biso` |
|---|---:|---|
| Peor caso por sesión, con ciclo completo | 12.905 bytes | 4.062 bytes |
| Media medida por sesión | 3.358 bytes | 4.062 bytes |
| Lecturas obligatorias por sesión | entre 1 y 4 | 1 |
| Contexto gastado en sesiones que no tocan tareas | la inyección en el fichero de convenciones | 0 |

**La media sube ligeramente, y conviene decirlo en vez de esconderlo.** Lo que cambia es otra cosa: el
coste pasa a ser fijo, conocido y acotado por una prueba, en vez de depender de cuántas guías decida
leer el agente, y el peor caso cae a menos de un tercio. El ahorro grande no está aquí, está en que la
salida de las escrituras deje de ser un eco y en las llamadas que desaparecen al fusionar el ciclo.

**El tope es una prueba de la suite, no un objetivo.** Y el reparto en dos mitades existe para que el
resumen del tablero, que crece con el tablero, no pueda comerse el sitio de las reglas.

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

**10.3, por qué existe `--start` al crear una tarea.** Evita que crear una tarea para ponerse con ella
en el mismo minuto cueste dos llamadas. Es el principio 5 aplicado a un caso medido.

**10.3, por qué `--comment` funciona al crear.** Por lo mismo: una tarea que nace con un comentario
cuesta una llamada.

**10.4, por qué el filtro de etiquetas es el único que combina sus valores con "y".** Porque el uso
normal de varias etiquetas es acotar, no ampliar.

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
otros tres sitios donde hace falta una identidad: el filtro `--mine` falla, la autoasignación de
`start` avisa, y el autor de un comentario es un error. La razón es que un tablero de una sola persona
no tiene por qué configurar su identidad solo para poder crear tareas.

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
