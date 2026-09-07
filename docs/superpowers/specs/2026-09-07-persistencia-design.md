# Diseño: cómo se guardan los datos de biso

Fecha: 2026-09-07

Esta es la decisión que `CLAUDE.md` marcaba como la única que bloqueaba escribir código. La evidencia
sobre cómo lo han resuelto (y roto) las demás herramientas está en
[`../../ESTADO-DEL-ARTE.md`](../../ESTADO-DEL-ARTE.md), y este documento no la repite: la cita.

## 1. Qué se decide

Un tablero es **una base de datos SQLite en un directorio propio, fuera del proyecto que gestiona**,
que el proyecto localiza con un fichero puntero versionado en git. La base de datos no se commitea
nunca. El historial y el respaldo se consiguen con una exportación en texto que sí se commitea, en el
repositorio del propio directorio del tablero.

Las cinco propiedades que hacen que esto funcione:

1. **Un tablero es uno**, con un solo asignador de identificadores, compartido por todas las copias de
   trabajo del proyecto. Nunca se sincroniza fusionando dos almacenes escritos por separado.
2. **El estado de las tareas no hereda la semántica del código.** Cambiar de rama no cambia el
   tablero, porque el tablero no está en el árbol de trabajo.
3. **No hay ningún proceso de fondo.** Ni daemon, ni servidor, ni nada que vigile.
4. **El texto es una salida, nunca la fuente de la verdad ni un canal de sincronización.**
5. **El camino caliente de un comando no ejecuta git.**

## 2. El modelo físico

Una base de datos SQLite por tablero, en modo WAL, con la configuración del tablero dentro de la
propia base de datos junto a las tareas.

Los diez campos derivados de la sección 5 de `SPEC.md` (`urgency`, `acDone`, `acTotal`, `dodDone`,
`dodTotal`, `commentCount`, `blocks`, `ready`, `blocked`, `waiting`) **no tienen columna**. Se calculan
al leer, como el documento ya exige.

Dos datos que la especificación trata como parte del estado del tablero y que por tanto viven en la
base de datos, no derivados de las tareas presentes:

- **El identificador más alto que se ha llegado a asignar** (sección 4.11), que es lo que permite
  distinguir `never_allocated` de `not_found` en la sección 7.3.
- **El contador de claves de criterios de cada tarea**, uno por lista, que solo crece (sección 5.1).

Sobre el contador de criterios conviene ser explícito, porque `max()+1` sobre las claves presentes
parece bastar y no basta. El caso que lo rompe no son los huecos: una tarea con los criterios #1 y #3
tiene máximo 3 y el siguiente sería el #4, correcto. Lo que lo rompe es **quitar el de la clave más
alta**: si la tarea tenía #1, #2 y #3 y se quita el #3, `max()+1` vuelve a dar 3 y la clave se
reasigna, algo que la sección 5.1 prohíbe expresamente.

Y no es un detalle cosmético, porque la misma sección 5.1 dice que los selectores de la sección 8.4
trabajan sobre la clave y **nunca** sobre la posición. Un agente que leyó la tarea cuando el #3 era el
criterio viejo, y que marca el #3 después de esa secuencia, marca en silencio un criterio distinto del
que creía. Es el mismo fallo que reutilizar el identificador de una tarea, en pequeño, y se arregla
igual: **enumerar lo presente dice qué existe ahora, no dice qué se usó ya.** Por eso la sección 4.11
obliga al tablero a recordar el identificador más alto que llegó a asignar, y por eso cada lista de
criterios necesita su contador.

Los comentarios se guardan con su **orden de inserción explícito**, aparte de su fecha, porque
`biso answer` añade al final un comentario con un instante pasado (sección 5.2) y ordenar por fecha
daría un orden equivocado.

### Una restricción del entorno que hay que escribir

El modo WAL exige que todos los procesos que abren la base de datos compartan un poco de memoria, y
por eso **no funciona sobre un sistema de ficheros en red**. La documentación de SQLite sobre uso en
red va más allá y habla de corrupción real, no solo de bloqueos que fallan.

Consecuencia práctica: **el directorio de un tablero no puede estar en una carpeta sincronizada**
(iCloud Drive, Dropbox, un montaje de red) ni dentro de un vault cifrado montado por red. Esto no se
puede impedir, pero sí detectar: `biso doctor` lo comprueba y lo reporta.

## 3. Dónde vive el tablero

### El directorio del tablero

Un directorio que contiene la base de datos, sus ficheros auxiliares de WAL, y el material versionado
de la sección 8. Su ubicación se decide así, y esto **no cambia** la sección 3.2 de `SPEC.md`, solo
rellena su paso 4:

1. `--board` gana siempre.
2. `BISO_BOARD` después.
3. El **puntero del proyecto**, que es el caso normal.
4. La **raíz por defecto de la máquina**, que es lo que hasta ahora estaba enunciado como "un tablero
   que el propio almacenamiento asocia al directorio de trabajo".

`biso init` sin argumentos crea el tablero dentro de la raíz por defecto y escribe el puntero.
`biso init --at <ruta>` lo pone donde se le diga y escribe el puntero igual.

### El puntero

Un fichero pequeño en la raíz del proyecto, **versionado en git**, con dos cosas y la versión de su
propio formato:

- **`id`**, obligatorio: un identificador **inmutable** de ocho caracteres hexadecimales, generado al
  crear el tablero. Es la identidad, y no se puede cambiar nunca.
- **`path`**, opcional: la ruta completa, solo cuando el tablero no está en la raíz por defecto. Puede
  coincidir con el nombre del tablero o no, y da igual.

**El nombre humano del tablero no va aquí: vive en la configuración del tablero**, junto al prefijo de
los identificadores. El motivo no es que el puntero se pueda editar (eso vale para cualquier cosa que
lleve dentro, y el resultado siempre es un error claro de tablero no encontrado, no un daño silencioso).
El motivo es que **un nombre es una etiqueta humana y las etiquetas colisionan entre personas**: bajo la
decisión de compartir más adelante, dos máquinas pueden tener cada una un tablero llamado `backend` que
son tableros distintos. Una etiqueta no puede ser la identidad.

Esa separación tiene una consecuencia buena: **renombrar el tablero no toca ningún puntero**, así que no
hay que commitear nada en ninguna copia de trabajo para renombrarlo.

**El `id` manda y el `path` es una pista.** El identificador es independiente de la máquina y la ruta no
lo es, así que un fichero versionado que llega a otro ordenador traerá un `path` que allí no existe. Eso
no es un problema nuevo: cae exactamente en el caso de error que esta sección ya necesitaba, el del
puntero que nombra un tablero que en esta máquina no está, y `biso init` adopta el `id` para que las dos
máquinas sigan hablando del mismo tablero.

**Cómo se llama la carpeta, y cómo se busca.** La carpeta del tablero se llama
**`<nombre>-<id>`**, por ejemplo `kex-3f9a2b1c`. El identificador nunca es el nombre de la carpeta por sí
solo: la raíz por defecto se lee con los ojos.

```
~/.biso/boards/kex-3f9a2b1c/
~/.biso/boards/biso-7d10e4a2/
~/.biso/boards/hub-c05b91ff/
```

Llevar el identificador en el nombre de la carpeta resuelve dos cosas de golpe:

- **Dos proyectos distintos de la misma máquina pueden llamarse igual.** Sin el sufijo, el segundo
  `biso init` choca con la carpeta del primero y hay que inventar una regla de desambiguación. Con el
  sufijo no choca nunca.
- **Encontrar el tablero desde el puntero es una sola búsqueda de patrón** sobre la raíz por defecto,
  sin abrir ni leer la configuración de ningún tablero. Y como el identificador va en el nombre, sigue
  resolviendo aunque el nombre haya cambiado: renombrar lleva `kex-3f9a2b1c` a
  `nuevonombre-3f9a2b1c` y el puntero no se toca.

Por eso el identificador puede ser corto, ya que se lee con los ojos: **ocho caracteres hexadecimales en
minúscula**, generados por `biso init` de la fuente de números aleatorios del sistema, comprobando que no
exista ya en la raíz por defecto, que es una lectura de directorio y sale gratis. Es la misma forma que
usa Beads para sus tareas, así que está probada como legible.

**El formato es JSON, y no YAML.** El motivo no es estético: la experiencia con Backlog.md en esta
máquina, anotada en la memoria global, es que su interfaz usa camelCase y su fichero snake_case, y que
**una clave mal escrita en el YAML se ignora en silencio**. El principio de `biso` es que un valor que no
existe es siempre un error, así que una clave desconocida en el puntero es un error con su código, no un
encogimiento de hombros. Y JSON es lo que la herramienta ya habla en todas partes.

Que esté versionado es lo que hace que funcionen tres cosas de golpe:

- **Los worktrees salen gratis.** El fichero está presente e idéntico en cada worktree del proyecto,
  así que todos resuelven al mismo tablero sin lógica especial y sin subir por el árbol de
  directorios.
- **Sobrevive a volver a clonar** en la misma máquina, porque el puntero vuelve con el clon y
  encuentra el tablero que ya existía.
- **La identidad compartida para el futuro ya existe y ya viaja.** Cuando algún día se comparta entre
  máquinas, no hay que inventarse una clave: la lleva el repositorio desde el primer día.

### Un caso nuevo de error que la especificación no tiene

El puntero puede nombrar un tablero que **en esta máquina no está**, porque alguien acaba de clonar el
proyecto en otro ordenador. Hoy la sección 3.2 solo tiene el código 8 con
`error: no board here, and none configured for this project`, y ese mensaje sería falso: sí hay uno
configurado para este proyecto, lo que no hay es el tablero.

Hace falta un mensaje propio que diga eso, y `biso init` en esa situación **adopta la identidad del
puntero** en vez de acuñar una nueva, para que las dos máquinas sigan hablando del mismo tablero.

### Dos casos más que conviene enunciar antes de que sorprendan

**Dos proyectos distintos pueden apuntar al mismo tablero, y es legal.** No hay forma de distinguir "dos
copias de trabajo del mismo proyecto" de "dos proyectos que comparten tablero", porque el mecanismo es el
mismo fichero puntero, y compartir es precisamente para lo que la sección 3.2 dice que existe. Conviene
escribirlo para que nadie lo trate luego como un fallo que hay que impedir.

**Un tablero cuyo proyecto ya no existe se queda huérfano y nadie lo ve.** El caso contrario, el puntero
sin tablero, tiene su error. Este no: si borras el proyecto, su tablero se queda en la raíz por defecto
para siempre, y `biso doctor` no puede saberlo porque trabaja sobre un tablero, no sobre la raíz. No se
resuelve aquí, pero se enuncia: la enumeración de la sección 10 es también lo que algún día permitirá
listarlos y ver los que ya no tienen dueño.

## 4. Los identificadores

Se quedan como están en la sección 4.11: `<PREFIX>-<n>`, `n` entero positivo, crecientes, **con huecos
permitidos**, y nunca reutilizados. El prefijo ya viene de la configuración, en `task_prefix`, con
`TASK` por defecto, así que un tablero que quiera identificadores como `KEX-1` no necesita nada nuevo.

**Pero el valor por defecto hay que cambiarlo.** Que dos tableros distintos empiecen los dos en
`TASK-1` hace que la vista multiproyecto de la sección 10 nazca inservible, salvo que el usuario se
acuerde de poner un prefijo distinto en cada tablero, que es precisamente la clase de cosa que no hay
que pedirle. La propuesta es que **`task_prefix` se derive por defecto del nombre del tablero** en
mayúsculas: un tablero llamado `kex` da `KEX-1` sin que nadie configure nada, y sigue siendo
configurable para quien quiera otra cosa.

Los dos valores no son la misma cosa aunque uno salga del otro. El nombre del tablero es una etiqueta
que se puede cambiar, y hay un comando para hacerlo. El prefijo está incrustado en cada identificador
para siempre. Así que se deriva al crear el tablero y desde entonces viven separados.

### El prefijo no se puede cambiar, y hoy la especificación deja que se cambie

Este es un agujero que hay que cerrar. La sección 10.10 lista `task_prefix` como una clave de
configuración normal y termina diciendo que **ningún cambio de configuración toca ninguna tarea, nunca**.
Las dos cosas juntas hacen que `biso config set task_prefix KEX` se acepte hoy y no toque nada, dejando
un tablero cuyas tareas se llaman `TASK-1` a `TASK-90` y cuya configuración afirma que el prefijo es
`KEX`. `task_prefix` es la única clave cuyo valor está incrustado en datos que ya existen, y por eso es
la única a la que esa frase no le sirve.

**La solución es prohibirlo**: error 6 en cuanto el tablero tenga alguna tarea, con la misma forma que ya
tiene quitar de `statuses` un estado que alguna tarea usa. No un comando amable que lo cambie, y por
cuatro daños concretos:

- Rompe las referencias que otros sistemas ya apuntaron: mensajes de commit, nombres de rama,
  descripciones de pull request, y la memoria del propio agente.
- Rompe el mensaje de la sección 7.3 que dice cuál es el identificador más alto asignado, y la
  comprobación de identificador mal formado, que tendría que aceptar dos prefijos a la vez.
- Rompe la simetría de exportar e importar, que preserva los identificadores exactamente: una
  instantánea vieja devuelve el prefijo antiguo.
- Y deja el tablero en un estado que ningún mensaje de la especificación sabe describir.

**La salida existe, y es un regalo de la decisión de la sección 8.** Quien de verdad quiera cambiar el
prefijo exporta, reescribe los identificadores en el NDJSON con cualquier herramienta, y lo importa en un
tablero nuevo. Es cirugía, explícita y de un solo sentido, y es posible precisamente porque la instantánea
es texto. Mejor eso que un comando que finja que la operación es segura.

No se pasan a identificadores aleatorios, que es la solución de Beads, porque el precio es perder
poder decir "la tarea 5" en voz alta y no hace falta: aquí hay un solo asignador por tablero, así que
la colisión que sufre Backlog.md no puede ocurrir. Su fallo no viene de que los números sean
secuenciales, viene de que **cada copia de trabajo asigna por su cuenta**.

Y el camino a varias máquinas ya está abierto sin cambiar nada, porque los huecos son legales por
escrito: **se reservan rangos por máquina**. Los huecos que eso deja no son una anomalía que haya que
explicar, son un caso que la especificación ya admite.

## 5. Concurrencia y atomicidad

Lo que SQLite en modo WAL garantiza y lo que la sección 4.10 de `SPEC.md` exige son el mismo contrato
escrito dos veces:

| Sección 4.10 exige | WAL da |
|---|---|
| Ninguna escritura se observa a medias, ni de una tarea ni de un lote de doscientas | Una transacción |
| Las lecturas nunca fallan por una escritura en curso y nunca la bloquean | Lectores y escritor no se estorban |
| Dos escrituras simultáneas sobre la misma tarea no se pierden ni se mezclan | Un solo escritor a la vez |
| Si no se consigue el acceso exclusivo, se espera hasta cinco segundos y se falla con código 7 sin escribir nada | `busy_timeout` de cinco segundos, y el `SQLITE_BUSY` que quede se traduce a código 7 |
| Dos procesos nunca asignan el mismo identificador, aunque trabajen desde copias de trabajo distintas | Asignar el identificador ocurre dentro de la transacción de creación |
| `biso prime` no escribe nunca y no necesita acceso exclusivo (sección 9.4) | Es una lectura |

Las operaciones multitarea (`new --from` con 242 líneas, `set` con varias referencias, `start`,
`finish` y `archive` con varias) son **una transacción**, lo que da el todo o nada sin ningún trabajo
extra. La validación completa antes de escribir sigue siendo necesaria, porque los códigos 9 y 7
distinguen el fallo detectado al validar del detectado al escribir.

## 6. El arrendamiento con caducidad, sin nada que vigile

La sección 9.2 de `DECISIONES.md` dejaba aplazado **quién detecta la caducidad sin que cueste caro**.
La respuesta es que no hace falta que nadie la detecte, porque **la caducidad no es un suceso, es un
campo derivado**.

La sección 5 ya tiene diez campos que no se guardan y se recalculan al leer, tres de ellos
(`ready`, `blocked`, `waiting`) del mismo estilo. "Arrendamiento vencido" es exactamente igual: una
tarea activa cuyo instante de caducidad ya pasó. Nadie tiene que darse cuenta en el momento; se sabe la
próxima vez que alguien mira, que es cuando importa.

Lo que sí se guarda es el instante de caducidad y **quién tiene el arrendamiento**. El latido que lo
renueva no es un proceso periódico: es cualquier escritura que el agente ya hace sobre esa tarea.

### Lo que vence es la reclamación, no el estado

Aquí hay una contradicción con la sección 9.2 de `DECISIONES.md` que hay que resolver a favor de este
diseño. Esa sección dice que la caducidad, al vencer, **saca la tarea del estado activo** sin tocar la
asignación. Pero `status` es un campo guardado, con un valor del vocabulario del tablero, y un campo
derivado no puede cambiar un campo guardado. Si nada escribe, la tarea sigue teniendo el estado activo
guardado por muy vencido que esté su arrendamiento.

Las dos salidas alternativas son peores:

- **Escritura diferida**, en la que la siguiente escritura cualquiera arrastre el saneamiento de las
  tareas vencidas, hace que un comando toque tareas que no nombró. Eso rompe la previsibilidad, y además
  `biso prime`, que no escribe nunca, mostraría un estado que una escritura ajena y posterior cambiaría.
- **Un proceso que vigile** está descartado en la sección 12.

Así que se separan las dos mitades: **el estado guardado no cambia solo, nunca.** Un campo derivado dice
que el arrendamiento está vencido, y liberarlo es una operación explícita que ejecuta el agente o una
persona. Es lo más honesto que se puede afirmar con lo que se sabe: nadie ha tocado esto en mucho tiempo,
y la tarea sigue donde la dejaron.

**Esto obliga a enmendar la sección 9.2 de `DECISIONES.md`**, cuya frase sobre sacar la tarea del estado
activo se escribió antes de esta decisión y es la que tiene que ceder.

### El token de vallado, que sale gratis

El estado del arte avisa de que un arrendamiento sin más deja un agujero: el agente A pierde el
arrendamiento por caducidad, B coge la tarea, y luego A despierta y escribe, robándola sin que nada lo
note. La defensa que la literatura llama token de vallado aquí no es un concepto nuevo: como el
arrendamiento guarda quién lo tiene, **una escritura que presume tenerlo lo comprueba dentro de la
misma transacción**. Eso es comparar y sustituir, y SQLite lo da por el hecho de estar en una
transacción.

## 7. Qué significa ahora "una tarea ilegible"

La sección 4.12 está escrita pensando en un fichero por tarea: si una no se puede leer, las lecturas
dirigidas dan error 3, las de conjunto la saltan y avisan, `export` sale con código 6, y `biso doctor`
la reporta. Con una base de datos eso hay que reescribirlo, porque **son dos casos distintos y hoy se
dicen como uno**:

- **Una tarea cuyo contenido no se puede interpretar.** Existe y se lee su fila, pero algún campo
  compuesto no tiene la forma esperada. Aquí el comportamiento de la sección 4.12 se mantiene
  entero y tal cual.
- **La base de datos entera no abre, o falla su comprobación de integridad.** Esto no es una tarea
  ilegible, es que no hay tablero legible, y necesita su propio código y su propio mensaje. No puede
  presentarse como "una tarea se ha saltado".

`biso doctor` gana una comprobación de integridad de la base de datos, que es barata y que hoy no
tiene equivalente.

Hay que ser honesto sobre lo que se pierde: con un fichero por tarea, el aislamiento del daño es
gratis y natural. Con una base de datos hay que provocarlo a propósito, y una corrupción de página
puede llevarse más de una tarea.

## 8. Historial, respaldo y rollback

El directorio del tablero es **su propio repositorio de git**. Lo que se versiona ahí no es la base de
datos, es la exportación:

- **Se versiona** la exportación en NDJSON de las tareas y **la configuración del tablero**.
- **Se ignora** el fichero de la base de datos y sus auxiliares de WAL.

**El repositorio de git del directorio del tablero es opcional, y su ausencia no rompe nada.** La
exportación se escribe igual, y sigue sirviendo para restaurar y para leer con los ojos; lo único que se
pierde es el historial. Esto importa por dos motivos: `biso` no puede exigir que git esté instalado para
funcionar, y hay quien no quiere un repositorio más en su máquina. Que `biso init` lo cree o no es una
decisión de comodidad, no de arquitectura.

Versionar la base de datos sería lo natural y es lo que no se hace, por tres motivos que se suman:
cada escritura reescribe páginas internas, así que cada commit guardaría una copia completa del
fichero en vez de un cambio; git no puede diffear un binario, así que el historial no diría qué pasó; y
ejecutar `git` cuesta unos 12 milisegundos medidos, más que el resto de la invocación junta.

**El paso de exportar y commitear es explícito y nunca está en el camino caliente de un comando.** Lo
llama un hook o el propio agente en los momentos naturales. Commitear en cada escritura, además de
romper el presupuesto de arranque, produce el ruido en git del que ya se queja la gente que usa
`auto_commit` en otras herramientas.

### El rollback, y la pieza que falta hoy

Un rollback es sacar la instantánea de un commit viejo, crear un tablero vacío e importarla. Funciona
porque la simetría entre `biso export` y `biso new --from` ya es una prueba de la suite: los dos
tableros quedan idénticos campo a campo, con identificadores, fechas y claves de criterios incluidas.

Dos cosas hay que decir con precisión:

- **La granularidad del rollback es la frecuencia de la instantánea.** Se vuelve a los momentos en que
  se exportó y se commiteó, no a cualquier instante. Es el precio de no ejecutar git en el camino
  caliente, y es el precio correcto.
- **Hoy la exportación no basta para reconstruir el tablero.** `biso export` lleva las tareas y no
  lleva la configuración: el propio ejemplo de la simetría en la sección 10.9 vuelve a declarar los
  estados, los tipos y las extensiones en el `init` del tablero de destino, y el documento avisa de que
  si el vocabulario no coincide el lote entero falla con código 9. **La instantánea tiene que incluir
  la configuración**, y restaurarla tiene que ser una operación y no una receta a mano.

Un rollback es de tablero completo a un punto del tiempo. No es deshacer la última operación, y no
pretende serlo.

## 9. El presupuesto de arranque

Hoy "que arranque rápido" vive solo en `CLAUDE.md` y no tiene número, lo que significa que nadie puede
incumplirlo. Pasa a ser **un requisito con una cifra y una prueba de la suite**, como ya lo son los
5.120 bytes del mensaje de arranque.

Con tres reglas que salen de lo que le pasó a otros:

- **No hacer al arrancar trabajo que nadie pidió.** Replit tenía un binario de Go que arrancaba en 11
  milisegundos y cuya herramienta real tardaba 277, por inicializar de golpe un mapa de 25 MB.
- **No ejecutar git en el camino caliente.** Backlog.md hace un fetch y recorre las ramas remotas para
  asignar un identificador, y tuvo que añadir dos opciones de configuración solo para acotar ese coste.
- **La palanca mayor no es hacer cada llamada más rápida, es hacer menos llamadas**, que es
  exactamente para qué existe `biso prime`.

Esto tiene una consecuencia sobre la otra decisión abierta del proyecto, que no se toma aquí pero que
queda acotada: con un presupuesto de un dígito o dos de milisegundos, los lenguajes interpretados
quedan fuera por aritmética. En esta máquina el suelo del sistema para arrancar cualquier proceso es
5,2 milisegundos, un binario de Go añade 2,2 encima, y Python y Node añaden 24,5 y 33.

## 10. Que los tableros se puedan enumerar

Esto no sirve a ningún comando de hoy: sirve para no cerrarle la puerta a una interfaz multiproyecto
mañana. Si cada proyecto puede poner su tablero donde quiera, nada puede listarlos todos.

**Convención más lista explícita.** Hay una raíz por defecto donde caen los tableros nuevos, y
cualquier herramienta que quiera enumerarlos lee esa raíz. Un tablero puesto fuera se añade con una
lista de raíces en la configuración de la máquina.

Lo que **no** se hace es un registro que se actualice solo con cada tablero creado. Parece la opción
completa y es la peor: es estado global mutable, se queda obsoleto en cuanto alguien mueve o borra un
directorio, y se convierte en una segunda fuente de la verdad. Es la clase de fallo que tiene Spec Kit,
cuyo fichero compartido de contexto lo pisa cada invocación de agente en silencio.

## 11. Configuración a nivel de máquina

`biso` necesita poder tener valores por defecto de la máquina, y como mínimo la raíz donde caen los
tableros nuevos y la lista de raíces adicionales.

El motivo es una lección medida en esta máquina y no una preferencia: Backlog.md **no tiene
configuración global**, comprobado con un `$HOME` falso, con un fichero en el directorio padre y con su
variable de entorno, y ninguno se lee. La consecuencia fue que cada repositorio lleva su copia, las
copias se separan solas, y hubo que escribir un script y una plantilla en dotfiles para repartir a mano
lo que debería ser un valor por defecto.

## 12. Lo que queda fuera, y por qué

- **La interfaz multiproyecto.** `biso` hoy no tiene ninguna interfaz, ni web ni de terminal, y añadir
  una es un subsistema nuevo que necesita su propia especificación. De ella aquí solo se recoge el
  requisito que afecta al almacenamiento, que es la sección 10.
- **Exportar al formato de Backlog.md.** No para uso propio, sino para que alguien que ya lo usa pueda
  probar `biso` sin salto al vacío. Importar de él sale barato y merece la pena. Exportar hacia él
  hereda sus bugs abiertos, entre ellos que al editar una tarea pierde las claves de frontmatter que no
  conoce y que sus marcadores de sección se anidan cuando un agente los reenvía, así que es trabajo
  perpetuo. Decisión de más adelante, con esa advertencia por delante.
- **Sincronizar entre máquinas.** No se diseña, pero no se cierra: la identidad compartida ya viaja en
  el puntero y los huecos de numeración ya son legales, que son las dos cosas que habría que retocar
  después si no se previeran ahora.
- **Un daemon.** Descartado con aritmética y no por gusto: el coste dominante de una invocación es
  arrancar un proceso, y el cliente que hablaría por el socket también lo es, así que el daemon compite
  por uno o dos milisegundos de unos ocho. Lo que cuesta está en el estado del arte, incluido que la
  documentación de Beads avisaba de que su modo daemon no funcionaba bien con los worktrees de git.

## 13. Qué hay que cambiar en SPEC.md

Esto es el encargo para el plan, no cambios ya hechos:

1. **Sección 3.2**, paso 4: la asociación entre tablero y directorio de trabajo pasa a ser la raíz por
   defecto de la máquina, y hay que describir el puntero (que va versionado, y qué lleva).
2. **Sección 3.2**, un caso de error nuevo: el puntero nombra un tablero que no está en esta máquina,
   con su mensaje y su código, distinto del código 8 actual. Y `biso init` adopta la identidad del
   puntero en ese caso.
3. **Sección 4.12**, reescrita: separar la tarea cuyo contenido no se interpreta de la base de datos
   que no abre, con código y mensaje propios para el segundo caso.
4. **Sección 10.9 y 10.3**: la instantánea incluye la configuración del tablero, y restaurarla es una
   operación, no una receta. Decidir su forma exacta.
5. **Sección 10.11**, `biso doctor`: comprobación de integridad de la base de datos, y aviso si el
   directorio del tablero está en un sistema de ficheros donde WAL no es seguro.
6. **Sección 14**: desaparece lo que dependía de esta decisión, y entra lo que queda fuera de verdad.
   La visibilidad entre versiones del proyecto queda resuelta por construcción, porque el tablero no
   está en el árbol de trabajo. Los tres mensajes de la sección 7.3 se mantienen, y el tercero pierde
   la mitad de su explicación actual sobre pertenecer a otra versión del proyecto.
7. **Sección 5**: el arrendamiento, con su instante de caducidad y su tenedor, y el campo derivado que
   dice si está vencido.
8. **Sección nueva o sección 4**: el presupuesto de arranque, con su cifra y su prueba.
9. **Sección 10.10**, configuración: los valores por defecto de la máquina y la lista de raíces.
10. **Sección 13**, contrato de estabilidad: el presupuesto de arranque se suma a los números
    congelados, o se dice explícitamente que no lo está.
11. **Sección 4.11 y sección 10.1**: el valor por defecto de `task_prefix` deja de ser `TASK` fijo y se
    deriva del nombre del tablero, para que dos tableros no colisionen los dos en `TASK-1`. Queda por
    decidir qué pasa con un nombre que no da un prefijo válido.
12. **Sección 10.10**, y es un agujero que ya existe: `task_prefix` pasa a ser inmutable en cuanto el
    tablero tenga alguna tarea, con error 6. Hoy se puede cambiar y la propia sección promete que ningún
    cambio de configuración toca ninguna tarea, así que el tablero queda con las tareas nombradas con el
    prefijo viejo y la configuración diciendo otro. El mensaje del error remite a la ruta de exportar,
    reescribir e importar.
13. **El nombre del tablero y su comando de renombrar**: el nombre vive en la configuración del tablero,
    no en el puntero, y hay un comando que lo cambia. Renombrar mueve la carpeta y no toca ningún
    puntero. Hay que decidir el nombre del comando y qué hace si la carpeta de destino ya existe.
14. **Sección 10.2**, `biso where`: tiene que decir también el identificador del tablero y su nombre, no
    solo la ruta, porque ahora son tres cosas distintas.
15. **La operación de liberar un arrendamiento vencido**, que es explícita y no automática (sección 6).
    Hay que decidir si es un comando propio o una reparación de `biso doctor --fix`.
16. **Sección 3.2**: dejar dicho que dos proyectos distintos pueden apuntar legalmente al mismo tablero,
    porque es el mismo mecanismo que hace que varias copias de trabajo lo compartan.

Y un cambio que no es de `SPEC.md`: **la sección 9.2 de `DECISIONES.md` hay que enmendarla.** Su frase
sobre que la caducidad saca la tarea del estado activo no se sostiene con esta decisión, porque un campo
derivado no puede cambiar un campo guardado. Lo que vence es la reclamación, no el estado, y esa
enmienda tiene que quedar registrada con su motivo y no reescrita sin más.

## 14. Lo que queda por decidir dentro de esta decisión

Cosas que el plan tiene que resolver y que aquí se dejan enunciadas a propósito:

- **La cifra del presupuesto de arranque**, que hay que medir en la máquina de referencia y no estimar.
- **La forma exacta de la instantánea restaurable** (un fichero con la configuración al lado del
  NDJSON, o un solo fichero con las dos cosas) y qué comando la restaura.
- **Cómo se llama el fichero puntero.** Su formato ya está decidido en la sección 3 (JSON, con `id`
  obligatorio e inmutable y `path` opcional), pero no su nombre en el disco.
- **Qué se hace si un nombre de tablero, ya con su sufijo, sigue siendo un nombre de carpeta inválido**
  en el sistema de ficheros, por ejemplo por su longitud o por sus caracteres.
- **Cómo se llama el paso de exportar y commitear**, y si es un comando o un efecto opcional de otro.
- **Si el lenguaje se decide ya**, ahora que el presupuesto de arranque acota la lista.
