# 2. El catálogo de problemas

## 1. Dos copias de trabajo asignan el mismo identificador

**Quién lo sufre.** Backlog.md, por diseño y reconocido por su mantenedor. En
<https://github.com/MrLesk/Backlog.md/issues/711> lo explica así: la asignación secuencial no puede
hacerse segura en su topología objetivo, porque cada clon calcula el siguiente número contra lo que
puede ver, así que dos escritores creando tareas entre dos pushes acaban asignando el mismo
identificador de forma determinista. Y añade que lo sufrieron en producción dos veces, con el mismo
identificador en dos tareas distintas, apareciendo como conflictos al commitear y, peor, como
referencias ambiguas desde las dependencias. **El issue está cerrado como "no planeado".** El pull
request que sí llegó (<https://github.com/MrLesk/Backlog.md/pull/749>) dice explícitamente que se
detiene en el diagnóstico y la recuperación con intervención humana, y que la prevención no está
incluida. Entre worktrees del mismo repositorio sí se previene desde el pull request
<https://github.com/MrLesk/Backlog.md/pull/710>, que al asignar un número mira también los ficheros de
tareas de los demás worktrees, commiteados o no, y comparte el bloqueo de creación entre todos ellos.
Eso funciona porque todos esos worktrees están en el mismo disco; entre dos clones no hay nada que mirar.

Peor que el conflicto es el silencio. En <https://github.com/MrLesk/Backlog.md/issues/632> el mantenedor
narra que dos sesiones de agente en historias divergentes calcularon el mismo número, y como el nombre
de fichero incluye el título, los dos ficheros tenían nombres distintos: **git fusionó las dos ramas sin
ningún conflicto** y quedaron dos tareas con el mismo identificador sin que nada avisara. Su conclusión:
un bloqueo local previene la creación concurrente en una máquina, pero no ayuda cuando dos historias
independientes divergen y se fusionan después.

Spec Kit tiene la misma forma del problema con la numeración de funcionalidades: dos ramas generan el
mismo índice y al fusionar quedan duplicados
(<https://github.com/github/spec-kit/discussions/2116>), con la sugerencia de usar marcas de tiempo o
una convención de equipo.

Beads lo resolvió con identificadores aleatorios cortos, y pagó con no poder decir "la tarea 5" en voz
alta.

**Qué hace `biso`.** Un tablero es uno, con un solo asignador. Los identificadores siguen siendo
secuenciales y legibles, porque el fallo no viene de que sean secuenciales, viene de que cada copia de
trabajo asigne por su cuenta. Compartir entre máquinas más adelante se hace reservando rangos, y los
huecos que eso deja ya son legales por escrito en ["Identificador de tarea"](../spec/modelo-de-datos/identificadores.md#identificador-de-tarea).

## 2. El estado de las tareas se bifurca con la rama

**Quién lo sufre.** Cualquiera que guarde las tareas como ficheros del árbol de trabajo. La
formulación más clara está en un repaso de estas herramientas
(<https://nesbitt.io/2026/08/20/issues-in-the-repo.html>): crear una rama de trabajo bifurca el estado
de la incidencia junto con el código, así que **una incidencia cerrada en la rama vuelve a aparecer
abierta en el momento en que te cambias a la principal.**

Backlog.md lo sufre, y su respuesta ha sido ir leyendo con git el estado de las demás ramas. Con la
opción `check_active_branches`, buscar o listar tareas mira también las ramas locales y los cambios
commiteados en los demás worktrees del mismo repositorio, y con `remote_operations` también las ramas
remotas. El issue <https://github.com/MrLesk/Backlog.md/issues/689> se quejaba de que un worktree no
veía lo que se hacía en otro, y está cerrado como resuelto desde el 2026-07-01, pero solo en parte. El
pull request <https://github.com/MrLesk/Backlog.md/pull/706> arregló que un proceso de larga duración
(el tablero web) sirviera resultados viejos, y dice expresamente que no cambia la visibilidad de los
cambios **sin commitear** de otro worktree. El pull request <https://github.com/MrLesk/Backlog.md/pull/710>
sí mira esos ficheros sin commitear, pero solo para no repetir identificadores al crear, no para listar.
El mantenedor pidió abrir otro issue si ese caso seguía importando. Así que Backlog.md lo sigue
sufriendo: una tarea cerrada en un worktree y sin commitear sigue abierta vista desde los demás.

Y leer las demás ramas trae sus propios fallos. Una misma tarea puede aparecer en dos ramas con
contenido distinto, y entonces la herramienta tiene que decidir cuál es la buena. Los issues
<https://github.com/MrLesk/Backlog.md/issues/783> (una tarea idéntica en dos ramas se marcaba como
ambigua) y <https://github.com/MrLesk/Backlog.md/issues/818> (editar una tarea sin commitear la volvía
ambigua con `check_active_branches` activado) son dos casos de eso. Además, cada consulta paga ejecutar
git sobre varias ramas, y con ramas remotas también la red.

Las herramientas que lo esquivan sacando el estado del repositorio (Vibe Kanban, Conductor) cambian el
problema por el contrario: el estado deja de seguir a la rama, que a veces es lo que quieres y a veces
no, y además se pierde el historial que daba el control de versiones.

**Qué hace `biso`.** Lo mismo que Vibe Kanban y Conductor en la mitad del estado, y lo contrario en la
mitad del historial. El estado vive fuera del árbol de trabajo y es uno solo: una tarea cerrada está
cerrada, no cerrada en esta rama, y ningún comando tiene que preguntarle a git por otras ramas para
saberlo. El historial no se pierde, porque `biso snapshot` escribe el tablero entero como texto y lo
commitea, en el repositorio del proyecto junto al código o en uno aparte (["`biso snapshot`"](../spec/cmd/snapshot.md)).
Ese texto es una salida y nunca una entrada: nada lo vuelve a leer como si fuera el estado, salvo
`biso init --from` para reconstruir un tablero desde cero. El razonamiento completo está en
["La decisión de persistencia"](../decisiones/persistencia.md#la-decisión-de-persistencia).

## 3. Dos escrituras a la vez, y una se pierde sin decir nada

**Quién lo sufre.** Backlog.md mide en <https://github.com/MrLesk/Backlog.md/issues/843> que **doce de
doce ediciones concurrentes se pierden**, y que los dos procesos salen con código cero y los dos
imprimen que actualizaron la tarea.

Task Master tenía el mismo agujero: sus funciones de leer y escribir el JSON no tenían bloqueo alguno,
así que dos ventanas de agente escribiendo a la vez hacían que la última pisara a la anterior
(<https://github.com/eyaltoledano/claude-task-master/issues/1567>). Se arregló con una librería de
ficheros de bloqueo y escrituras atómicas.

Claude Code sufre lo mismo en su propio fichero de configuración, que varias sesiones truncan a mitad
de escritura (<https://github.com/anthropics/claude-code/issues/28973>).

**Qué hace `biso`.** Lo tenía escrito antes de elegir el mecanismo:
["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) exige que ninguna
escritura se observe a medias, que dos escrituras simultáneas sobre la misma tarea no se pierdan ni se
mezclen, y que si no se consigue el acceso exclusivo se espere hasta cinco segundos y se falle con
código 8 **sin escribir nada**. Una transacción de SQLite en modo WAL da exactamente eso.

## 4. Un bloqueo que no cruza máquinas, o que se queda huérfano

**Quién lo sufre.** Backlog.md endureció su bloqueo dos veces y sigue reconociendo el límite: el issue
<https://github.com/MrLesk/Backlog.md/issues/565> admite que el intento anterior mantenía el bloqueo
más tiempo que la ventana crítica, y el issue 711 dice que un bloqueo de fichero solo serializa
escritores en el mismo sistema de ficheros y **no puede cruzar remotos de git**. El issue 937, todavía
abierto, propone coordinación por referencias de git con compare y sustituye, reconociendo que
comprobar antes de editar no cierra la carrera distribuida.

El bloqueo huérfano tiene su ejemplo canónico en el `index.lock` de git, y la defensa habitual es
caducar el fichero de bloqueo por su fecha de modificación en vez de confiar en que el proceso lo
limpie.

**Qué hace `biso`.** No usa ficheros de bloqueo propios: usa el bloqueo del motor, que un proceso
muerto libera solo. Y no intenta cruzar máquinas, porque un tablero es uno.

## 5. Archivar una tarea libera su identificador y el siguiente lo reutiliza

**Quién lo sufre.** Backlog.md, en <https://github.com/MrLesk/Backlog.md/issues/997>: crear una tarea
escanea las carpetas de activas y completadas pero no la de archivadas, así que el identificador vuelve
a estar libre y quedan dos ficheros distintos con el mismo identificador en carpetas distintas.

**Qué hace `biso`.** ["Identificador de tarea"](../spec/modelo-de-datos/identificadores.md#identificador-de-tarea) lo prohíbe expresamente: un identificador no se reutiliza jamás, y
el tablero sabe en todo momento cuál es el más alto que ha llegado a asignar, dato que se guarda aparte
de las tareas presentes. Eso es también lo que le permite dar
[tres mensajes distintos de "no la encuentro"](../spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), distinguiendo un identificador mal formado, uno que nunca existió, y uno
que existió y ya no está.

## 6. Una tarea se queda cogida porque la sesión murió

**Quién lo sufre.** Todo el mundo, y el vocabulario ya está inventado. Un proyecto describe su
arquitectura como un sistema de arrendamiento renovado por latido para evitar la doble reclamación
(<https://news.ycombinator.com/item?id=47341352>). Un artículo lo formaliza señalando que un bloqueo
simple se queda para siempre si su propietario se detiene, y propone un arrendamiento con límite de
tiempo **más un token de vallado que rechace las escrituras del propietario antiguo**
(<https://zenn.dev/agentmemories/articles/agent-task-lease?locale=en>). Hay incluso un issue con ese
vocabulario en el título (<https://github.com/xorbitsai/xagent/issues/1230>).

**Qué hace `biso`.** El arrendamiento vencido es **un campo derivado**, no un suceso: una tarea activa
cuyo instante de caducidad ya pasó. No hace falta nada que vigile, porque se sabe la próxima vez que
alguien mira, que es cuando importa. El latido que lo renueva es cualquier escritura que el agente ya
hace. De las dos piezas que propone el artículo se lleva entera la caducidad, y del token de vallado
solo la mitad: el arrendamiento guarda quién lo tiene, y `biso start` compara ese dato dentro de la
misma transacción, de modo que nadie renueva ni se atribuye un arrendamiento ajeno y de dos
reclamaciones simultáneas de uno vencido solo gana una. Lo que no hace es rechazar la escritura del
tenedor antiguo, que es la otra mitad del token: la acepta entera y solo deja intactos los dos campos
del arrendamiento, avisando por stderr, así que quien despierta tarde puede comentar, anotar o cerrar la
tarea que otro ya reclamó.
["Saber si alguien está trabajando de verdad"](../decisiones/modelo-de-estados.md#saber-si-alguien-está-trabajando-de-verdad) de "Decisiones de diseño" cuenta por qué se aparta a propósito,
y ["Riesgos conocidos y aceptados del modelo de estados"](../decisiones/modelo-de-estados.md#riesgos-conocidos-y-aceptados-del-modelo-de-estados) lo anota como riesgo aceptado.

## 7. El agente escribe mal el formato y se traga datos

**Quién lo sufre.** Backlog.md tiene tres issues abiertos que son el mismo problema por tres sitios. En
<https://github.com/MrLesk/Backlog.md/issues/990>, si un agente copia la descripción ya renderizada,
con sus marcadores de sección incluidos, y la reenvía como entrada, los marcadores se anidan y las
ediciones sucesivas se van tragando los criterios de aceptación y las notas, que desaparecen de la
salida. En <https://github.com/MrLesk/Backlog.md/issues/1000>, confundir flags crea un criterio
fantasma en silencio. En <https://github.com/MrLesk/Backlog.md/issues/1008>, un subencabezado dentro de
una sección la trunca, porque el patrón que la extrae no distingue dos almohadillas de tres.

En Hacker News está la queja de fondo sobre editar texto con un agente: tiene problemas para editar
ficheros grandes y no puede añadir texto al final de uno grande por el tamaño de su ventana de
contexto.

**Qué hace `biso`.** Nadie escribe el almacén a mano, ni el agente ni la persona: se escribe por
comandos, y cada comando valida. Es el sitio donde la base de datos gana de forma más clara, porque no
hay un formato de texto que un agente pueda malinterpretar al reenviarlo.

Con una honestidad al lado: **con un fichero por tarea, aislar el daño de una tarea corrupta es gratis**,
y ["Qué pasa con un dato que no se puede interpretar"](../spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar) está escrita pensando en eso. Con una base de datos hay que separar dos
casos que hoy se dicen como uno: una tarea cuyo contenido no se interpreta, y una base de datos que no
abre.

## 8. Un dato que la herramienta no conoce se pierde al editar

**Quién lo sufre.** Backlog.md, en <https://github.com/MrLesk/Backlog.md/issues/918>, etiquetado como
bug: los ficheros de tarea son markdown por diseño, lo que invita a que otras herramientas añadan sus
propias claves, y hoy cualquier clave así se pierde en silencio en la siguiente edición. Se reproduce
añadiendo una clave a mano y haciendo una edición trivial.

**Qué hace `biso`.** El problema no se plantea igual, porque nadie edita a mano el almacén: cada
escritura pasa por un comando que valida, y un dato que `biso` no conoce **nunca se guarda a medias ni
se pierde en la siguiente edición, sino que se rechaza en el momento** con un error. Un flag que no
existe es error 2, una clave que no existe en una línea de un lote de `biso new --from` es error 2 con
el código `unknown_key`, y una clave de configuración que no existe es error 4.

Lo que `biso` no ofrece es un campo de extensión para que otra herramienta guarde sus propios datos en
una tarea: no lo tiene. La respuesta prevista para un dato corto asociado a una clave es una etiqueta
con ámbito, como `milestone::m1`, que se escribe y se consulta con las etiquetas de siempre. Es una
decisión escrita y todavía no implementada, en
["Las etiquetas con ámbito"](../decisiones/detalles.md#las-etiquetas-con-ámbito). Y lo que no cabe en
una etiqueta, un texto libre por clave, no tiene sitio hoy: si hace falta, se decide y se especifica
como un campo, según ["Se retira `ext`"](../decisiones/detalles.md#se-retira-ext).

## 9. Estado compartido que cada invocación pisa

**Quién lo sufre.** Spec Kit, en <https://github.com/github/spec-kit/issues/4128>: un fichero
compartido guarda qué funcionalidad está activa, y cada invocación de agente lo sobrescribe sin avisar,
así que con dos agentes concurrentes uno fija el contexto y el otro lo pisa, sin ningún error visible.
Sigue abierto sin respuesta.

**Qué hace `biso`.** Es el motivo de rechazar un registro global de tableros que se actualice solo, por
mucho que parezca la opción completa. La enumeración se hace por convención (una raíz por defecto donde
caen los tableros) más una lista explícita en la configuración, sin ningún índice que pueda quedarse
obsoleto ni convertirse en una segunda fuente de la verdad.

## 10. El almacén binario no se versiona, y el texto no tiene transacciones

**La mitad binaria.** Versionar un SQLite en git no funciona: cada escritura reescribe páginas
internas, así que a veces el fichero cambia sin que cambie su contenido lógico, y cada commit guarda una
copia completa en vez de un cambio (<https://ongardie.net/blog/sqlite-in-git>). Git LFS no ayuda,
porque resuelve el peso del repositorio y no la fusionabilidad. Litestream tampoco, porque es
continuidad operativa con un único escritor primario, no colaboración. La respuesta seria del espacio
es Dolt, que sustituye el motor por uno con fusión a nivel de celda, y es lo que eligió Beads.

**La mitad de texto.** Un fichero por tarea no tiene transacciones: el renombrado atómico lo da el
sistema operativo para un fichero, no para un directorio, así que el todo o nada sobre doscientas
tareas repartidas en doscientos ficheros hay que construirlo. Y un solo fichero de log fusionable tiene
su propia trampa: decirle a git que fusione por unión de líneas **resucita las líneas borradas**, porque
concatena la unión de lo que hay en los dos lados y solo quita duplicados exactos, sin razonar sobre qué
falta ni por qué
(<https://stackoverflow.com/questions/12947436/git-union-merge-brings-back-some-deleted-lines>). GitLab
y scikit-learn lo usan para sus registros de cambios y documentan sus efectos, incluidas entradas
duplicadas en dos secciones (<https://github.com/scikit-learn/scikit-learn/issues/21516>).

**Qué hace `biso`.** Se queda con la mitad que da transacciones y no versiona el binario: la base de
datos es la verdad y no entra en git, y lo que se versiona es una exportación en texto, una tarea por
línea, más la configuración del tablero. El texto es una salida, nunca un canal de vuelta. Esa
distinción es exactamente lo que Beads avisa por escrito: su importación es solo de inserción y
actualización, y no puede saber si un registro ausente fue borrado o simplemente no se exportó.

Hay además un patrón verificado que hace bien la mitad contraria y del que se toma prestada una idea:
`org-roam` usa ficheros de texto como verdad y SQLite como caché, y detecta que la caché quedó obsoleta
**comparando un hash del contenido**, no la fecha de modificación, que se engaña con una copia que
preserva marcas de tiempo.

## 11. El proceso de fondo

**Quién lo sufre.** Beads. Tuvo daemon, hacía una sola cosa (traerse cambios en la máquina receptora) y
sus flags sugerían que hacía mucho más, lo que generaba expectativas equivocadas. **Se eliminó por
completo** al cambiar de motor. Y el comentario más citado de quien lo reemplazó por algo más simple
dice que empezó a pelearse con él varias veces por semana porque su daemon de fondo se puso a
sincronizar las cosas equivocadas (<https://news.ycombinator.com/item?id=46487580>). En el mismo hilo,
otro lo llama un lío incomprensible para algo que en el fondo es una idea simple.

Y su documentación avisaba de que **el modo daemon no funcionaba correctamente con los worktrees de
git** por el estado de base de datos compartido.

Cuando su sucesor exigió un servidor externo, los autores del motor publicaron una marcha atrás
reconociendo la fricción que eso añadía a quien trabaja solo.

**Qué hace `biso`.** No hay daemon, y el motivo es aritmético antes que estético. El coste dominante de
una invocación es arrancar un proceso, y el cliente que hablaría por el socket también es un proceso, así
que el daemon compite por uno o dos milisegundos de unos ocho pagando con una arquitectura entera. Un
daemon colgado, además, haría fallar también las lecturas, que
["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) promete que nunca fallan
por una escritura en curso.

## 12. El coste de arranque y el coste de contexto

**Medido en esta máquina** (Apple M3 Max, macOS 26.5.2, 300 iteraciones). El suelo del sistema para
arrancar cualquier proceso es 5,2 ms, y hay que restarlo para comparar lenguajes: C añade 0,5 ms, Rust
1,6 ms, Go 2,2 ms, Python 3.14 24,5 ms y Node 25.6 33 ms. Leer 300 tareas de un JSON, ordenarlas e
imprimirlas costó 8,7 ms en Go y 47,1 ms en Python. Doscientas invocaciones seguidas: 1,58 s en Go, 7,97
s en Node, 9,22 s en Python. Herramientas reales instaladas, para calibrar: `rg` 7,1 ms, `git --version`
12,3 ms, `kubectl` 29,4 ms, `gh` 46,8 ms, `npm` 85,3 ms, `aws` 247,2 ms.

**El aviso importante no es sobre el lenguaje.** Replit documenta un binario de Go que arrancaba en 11
ms y cuya herramienta real tardaba 277, porque inicializaba de golpe un mapa generado de 25 MB
(<https://replit.com/blog/golang-performance>). El pecado es hacer al arrancar trabajo que nadie pidió.

**El coste de contexto es el otro.** El servidor de protocolo de Task Master expone 36 herramientas,
unos 21.000 tokens, y tuvo que añadir un modo de 7 herramientas y 5.000. Y en el hilo de Backlog.md se
pregunta cómo repartir tareas a un agente sin agotar su contexto navegando muchos ficheros.

**Qué hace `biso`.** El presupuesto de arranque pasa a ser un requisito con cifra y con prueba de la
suite, como ya lo son los 5.504 bytes del mensaje de arranque. Y la palanca principal no es hacer cada
llamada más rápida sino hacer menos llamadas, que es para lo que existe `biso prime`: una invocación que
sustituye a muchas, con su tamaño acotado por contrato.

## 13. El commit automático hace ruido

**Quién lo sufre.** Backlog.md tiene un feature request abierto pidiendo un modo que enmiende su propio
commit precisamente para reducir el ruido en git
(<https://github.com/MrLesk/Backlog.md/issues/821>). MCP Shrimp Task Manager commitea tras cada cambio
en su propio directorio de datos.

**Qué hace `biso`.** El paso de exportar y commitear es explícito y **nunca está en el camino caliente
de un comando**, que además de evitar el ruido evita pagar los 12 ms de `git` en cada invocación. La
granularidad del historial y por tanto del rollback es la frecuencia con la que se toma la instantánea,
y eso se dice en voz alta en vez de disimularse.

## 14. Un valor que no existe devuelve una lista vacía en vez de un error

**Quién lo sufre.** Backlog.md, en <https://github.com/MrLesk/Backlog.md/issues/919>: pasar un hito que
no coincide con nada **crea un hito virtual en silencio**. Y en
<https://github.com/MrLesk/Backlog.md/issues/824>, el flag `--json` estaba documentada en el README
y rechazada por todos los comandos de lectura, con el reportante encontrando la cadena compilada en el
binario pero no conectada.

**Qué hace `biso`.** Es uno de sus principios y está en su `CLAUDE.md`: un valor que no existe es
siempre un error, se esté escribiendo o leyendo, porque un filtro mal escrito nunca puede devolver una
lista vacía, ya que quien la lee la interpreta como un hecho sobre el tablero.

## 15. Usar GitHub Issues cuesta cuota y latencia

**Quién lo sufre.** Quien lo elige por no montar nada. GitHub mantiene cuotas separadas de unas 5.000
unidades por hora. Un incidente narrado con números: nueve agentes lanzados a la vez abrieron 10 pull
requests en 22 minutos, y en el minuto 8 empezaron los errores de límite excedido, agotando las 5.000
peticiones en 90 segundos, con más de 60 fallos en cascada y hasta una hora de recuperación
(<https://www.tamirdresher.com/blog/2026/03/21/rate-limiting-multi-agent>), sin ningún mecanismo de
prioridad entre el agente que sondeaba en segundo plano y el que necesitaba la cuota para trabajo
crítico.

Y hay una queja de otra naturaleza: no se distingue lo que escribió un humano de lo que escribió un
agente. Una discusión formal pide poder bloquear issues generados por Copilot y dice que no hay ninguna
indicación visible de que fuera generado, ni en la interfaz ni en la interfaz de programación
(<https://github.com/orgs/community/discussions/159749>).

**Qué hace `biso`.** Es local y no tiene cuota. Y distingue quién escribió qué **mientras cada quien
declare su propia identidad**, con autor en los comentarios y en la pregunta abierta. Esa condición no
depende de la configuración de ningún tablero: la identidad vive en la configuración de la máquina de
cada quien, y la variable de entorno `BISO_ME` gana siempre sobre ella, así que un agente que fija su
propia `BISO_ME` nunca hereda la identidad configurada en la máquina, comparta tablero con quien
comparta (["La identidad de quien llama"](../decisiones/detalles.md#la-identidad-de-quien-llama)).

## 16. El agente deja de mirar el tablero cuando el proyecto crece

**Quién lo sufre.** Task Master, según un comentario en el hilo de Backlog.md: estaba bastante bien,
pero a veces el agente lo ignoraba a medida que el proyecto crecía
(<https://news.ycombinator.com/item?id=44486442>). Y con ficheros markdown sueltos la queja es la
inversa y complementaria: hay riesgo de que los agentes tropiecen con tickets obsoletos y consuman
tokens (<https://news.ycombinator.com/item?id=46509745>).

**Qué hace `biso`.** Es el trabajo de `biso prime`: un mensaje de arranque de tamaño acotado por
contrato, que cabe entero en el contexto y dice lo que hace falta para empezar bien, en vez de dejar que
el agente decida cuánto tablero leer.
