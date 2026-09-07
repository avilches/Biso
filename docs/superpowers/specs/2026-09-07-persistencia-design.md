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
  Una tarea puede tener los criterios #1 y #3 y ningún #2, así que el contador no se puede deducir del
  máximo de las claves presentes.

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

Un fichero pequeño en la raíz del proyecto, **versionado en git**, con la identidad del tablero y la
versión del formato del puntero. Nada más: ni rutas absolutas, ni configuración, ni nada que caduque.

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

## 4. Los identificadores

Se quedan como están en la sección 4.11: `PREFIX-<n>`, `n` entero positivo, crecientes, **con huecos
permitidos**, y nunca reutilizados.

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

## 14. Lo que queda por decidir dentro de esta decisión

Cosas que el plan tiene que resolver y que aquí se dejan enunciadas a propósito:

- **La cifra del presupuesto de arranque**, que hay que medir en la máquina de referencia y no estimar.
- **La forma exacta de la instantánea restaurable** (un fichero con la configuración al lado del
  NDJSON, o un solo fichero con las dos cosas) y qué comando la restaura.
- **El nombre y el formato del fichero puntero.**
- **Cómo se llama el paso de exportar y commitear**, y si es un comando o un efecto opcional de otro.
- **Si el lenguaje se decide ya**, ahora que el presupuesto de arranque acota la lista.
