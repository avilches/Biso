# Estado del arte: gestores de tareas para agentes, y por qué fallan

Investigado el 2026-09-07 para tomar la decisión de persistencia de `biso`
([diseño](superpowers/specs/2026-09-07-persistencia-design.md)).

La primera mitad es un inventario de las herramientas del espacio, y se cierra con una comparación
aparte: cómo se le dice a cada una que trabaje con un almacén distinto del que encuentra sola, donde
entran también herramientas de fuera del espacio. La segunda es un catálogo de
problemas: cada uno con quién lo sufre, la evidencia con su enlace, y la respuesta de `biso`. La
segunda mitad es la que sirve para explicar la herramienta a alguien que viene de otra.

## Aviso de método

Lo que sigue se apoya en repositorios de GitHub (código, issues, documentación), en Hacker News, y en
documentación oficial. **No hay ninguna cita de Reddit**: el dominio rechaza las descargas
automatizadas y también fallaron las vías alternativas, así que las opiniones de r/ClaudeAI y
r/ChatGPTCoding no están representadas.

Cinco cosas concretas quedaron sin verificar del todo, y se marcan también donde aparecen:

1. **La cita de Steve Yegge** sobre que los conflictos del JSONL de Beads los resuelve la propia IA al
   fusionar. El artículo de Medium devolvió un error 403 y nunca se leyó directamente.
2. **Las versiones exactas del fallo de WAL** de SQLite entre procesos concurrentes (3.7.0 a 3.51.2,
   corregido en 3.51.3), que salen de un resumen de la página oficial y no de su registro de cambios.
3. **El número de estrellas de Spec Kit**, leído de la página y no de la interfaz de programación.
4. **Las tasas de conflicto del 19,8% y del 41,7%** entre pull requests de agentes, que vienen de un
   único artículo y no se pudieron cruzar con una segunda fuente.
5. **La corrupción de SQLite específicamente sobre iCloud Drive o Dropbox por su nombre.** La
   documentación oficial habla de sistemas de ficheros en red en general; aplicarlo a esas carpetas es
   una extrapolación razonable, no un hecho citado.

---

# Parte 1. Las herramientas

## Backlog.md

`https://github.com/MrLesk/Backlog.md`, 6.664 estrellas, 402 forks, 48 issues abiertos, TypeScript,
MIT. Última versión consultada, la 1.51.0. Unas 59.000 descargas mensuales en npm.

Plantea el problema como uno de atención humana, no de generación: los agentes producen en una hora más
código plausible del que puedes leer con cuidado en un día, así que el cuello de botella es tu
atención. Propone tres puntos de control (revisar la especificación de la tarea, revisar el plan que el
agente escribe dentro de ella, revisar el código) bajo la regla de una tarea, una ventana de contexto,
un pull request.

**Cómo guarda los datos.** Un fichero markdown por tarea, versionado en git, bajo una carpeta
configurable (`backlog/` por defecto), repartido en `tasks/`, `drafts/`, `completed/`, `docs/`,
`decisions/`, `milestones/` y `archive/`. El nombre del fichero es `<prefijo>-<id> - <Título>.md`. El
frontmatter es YAML en `snake_case` (`created_date`, `due_date`, `parent_task_id`, `modified_files`),
con la excepción de `onStatusChange`, que se escribe en camelCase. El cuerpo delimita cada sección con
comentarios HTML explícitos (`<!-- SECTION:DESCRIPTION:BEGIN -->`, `<!-- AC:BEGIN -->`,
`<!-- DOD:BEGIN -->`), y cada criterio de aceptación es una casilla numerada. **No hay índice ni
caché**: cada comando recalcula escaneando el directorio, y para asignar un identificador puede llegar
a hacer un fetch de git y recorrer todas las ramas remotas.

**Cómo lo usa un agente.** Línea de comandos con `--plain` y `--json` (este último bajo un esquema
versionado con `schemaVersion`), servidor de protocolo de contexto de modelo, y escritura de
instrucciones en `AGENTS.md` o `CLAUDE.md`.

## Beads (`bd`)

`https://github.com/gastownhall/beads` (antes `steveyegge/beads`), 26.963 estrellas, más de 1.000
issues abiertos. Anunciado en octubre de 2025.

Ofrece memoria persistente y estructurada para agentes, y sustituye los planes desordenados en markdown
por un grafo con dependencias. Su propio texto pide instalar en el `AGENTS.md` del proyecto la
instrucción de no usar listas de tareas en markdown.

**Cómo guarda los datos, en dos arquitecturas separadas por una fecha.** Hasta la versión 0.49.x
(febrero de 2026), SQLite local como copia de trabajo por máquina, ignorada por git, y
`.beads/issues.jsonl` como fuente de la verdad versionada. Identificadores por hash corto de un
identificador aleatorio (`bd-a1b2`) para que dos ramas no colisionaran. Merge driver propio de git,
instalado con `git config merge.beads.driver`, que resolvía campo a campo: las fechas por máximo, las
dependencias por unión de conjuntos, el estado y la prioridad a tres bandas. En la versión 0.50 migró a
**Dolt**, una base de datos SQL con ramas y fusión a nivel de celda nativas, sincronizando por una
referencia propia de git, y el JSONL pasó a ser solo exportación.

**Su documentación avisa de dos cosas que importan.** Que el JSONL no sustituye a la sincronización
propia, porque la importación es solo de inserción y actualización y **no puede saber si un registro
ausente fue borrado o simplemente no se exportó**
(`https://raw.githubusercontent.com/gastownhall/beads/main/docs/core-concepts/sync-concepts.md`). Y, en
la época del daemon, que **el modo daemon no funcionaba correctamente con los worktrees de git** por el
estado de base de datos compartido.

**Cómo terminó.** El cambio a Dolt costó la atomicidad de commitear código y tareas juntos
(`https://github.com/gastownhall/beads/issues/2489`), y el 2 de abril de 2026 DoltHub publicó una marcha
atrás parcial, "Restoring Beads Classic"
(`http://www.dolthub.com/blog/2026-04-02-restoring-beads-classic/`), reconociendo que el paso a Dolt
añadió fricción a los usuarios en solitario, acostumbrados al modelo más simple de SQLite y git que no
exigía un servidor externo. El daemon se eliminó por completo.

## Task Master (`claude-task-master`)

`https://github.com/eyaltoledano/claude-task-master`, 28.056 estrellas, 2.623 forks, 212 issues
abiertos. Detrás hay ahora una empresa. Es la más popular por estrellas y, curiosamente, la que menos
tracción tuvo en Hacker News.

Convierte un documento de requisitos en tareas con dependencias.

**Cómo guarda los datos.** JSON en `.taskmaster/tasks.json`, o un fichero por etiqueta si hay varias.
Su respuesta a los conflictos de fusión **no es un algoritmo, es evitar que dos ramas escriban el mismo
fichero**: su documentación dice que las etiquetas evitan conflictos cuando varios miembros crean tareas
en ramas distintas, con `tm add-tag --from-branch`. No hay mecanismo para cuando dos ramas tocan la
misma etiqueta. Hay un pull request abierto que propone migrar a SQLite con sincronización a JSONL.

**Un dato de coste de contexto.** Su servidor de protocolo expone 36 herramientas por defecto, unos
21.000 tokens, y tuvieron que añadir un modo reducido de 7 herramientas y unos 5.000.

## Spec Kit (GitHub)

`https://github.com/github/spec-kit`, MIT, anunciado el 2 de septiembre de 2025. Estrellas muy altas
pero sin verificar con precisión.

Desarrollo dirigido por especificaciones, con el flujo de especificar, planificar, generar tareas e
implementar. Todo son ficheros versionados en `.specify/` y `specs/`, numerados secuencialmente por
funcionalidad. Sin base de datos, sin servidor de protocolo: interviene con comandos de barra dentro del
agente anfitrión.

## Vibe Kanban

`https://github.com/BloopAI/vibe-kanban`, 28.030 estrellas, 3.002 forks, 539 issues abiertos. **El
equipo lo abandonó el 10 de abril de 2026** (`https://www.vibekanban.com/blog/shutdown`): miles de
ingenieros lo usaban a diario pero casi todos gratis, y no encontraron modelo de negocio.

Tablero para orquestar varios agentes en paralelo, revisando diferencias y fusionando desde una sola
interfaz. **SQLite local, fuera del repositorio**, con la advertencia oficial de que borrar ese
directorio pierde todas las tareas y ajustes. Lo único que toca git son los espacios de trabajo, cada
uno con su worktree y su rama. Solo servidor de protocolo, sin línea de comandos de gestión.

## Conductor

`conductor.build`, de Melty Labs. Aplicación nativa de macOS, software cerrado, presentada en julio de
2025. Corre varios agentes en paralelo, cada uno en su worktree bajo
`~/conductor/workspaces/<repo>/<workspace>`, con una base SQLite local para su propio estado y una capa
de nube solo en planes de pago. Las notas de traspaso entre agentes van a una carpeta deliberadamente
excluida de git.

## Amp (Sourcegraph)

**No tiene tablero ni gestor de tareas.** Su unidad de trabajo es la conversación persistente,
sincronizada entre web, aplicaciones y línea de comandos, guardada en PostgreSQL en la nube y sin
despliegue autohospedado (`https://ampcode.com/security`). No hay exportación de una conversación a
markdown ni a JSON. Las quejas públicas encontradas son todas sobre precio.

## GitHub Issues como tablero

La acción oficial de Anthropic se dispara con una mención en un issue o comentario
(`https://code.claude.com/docs/en/github-actions`), y el agente de Copilot funciona asignándole el issue
como a un compañero. La ventaja que la gente destaca es que no hay que construir infraestructura de
coordinación, porque los tickets ya sirven de memoria entre invocaciones y de bus de mensajes vía
comentarios.

Un detalle revelador: la acción oficial **rechaza por defecto a los actores que son bots, para evitar
que un bot dispare a Claude en un bucle.** El propio fabricante reconoce que un tablero compartido es
también un bucle de retroalimentación compartido.

## Ficheros `TODO.md` y `PLAN.md` sueltos

La opción de cero herramientas, con su especificación de facto en
`https://github.com/todomd/todo.md`. Su uso más citado es recuperar contexto al agotar la ventana. Sus
dos quejas también están documentadas, y están en la parte 2.

## Servidores de protocolo genéricos

El registro oficial ya no mantiene catálogo de terceros. Con tracción propia se encontraron **MCP
Shrimp Task Manager** (`https://github.com/cjo4m06/mcp-shrimp-task-manager`, 2.153 estrellas), que
guarda en `tasks.json` y hace algo poco común, inicializar un repositorio de git dentro de su propio
directorio de datos y commitear tras cada cambio; y **taskboard**
(`https://github.com/tcarac/taskboard`, 29 estrellas), explícitamente SQLite y binario único.

## Cómo se apunta a un almacén distinto del que la herramienta encuentra sola

Todo lo anterior dice dónde guarda los datos cada herramienta. Esto es lo otro: cómo se le dice a una
herramienta que trabaje con un almacén que no es el que habría encontrado por su cuenta. Investigado el
2026-09-09 sobre documentación oficial, y sobre el código en el único caso en que la documentación no
llega, `restic`. Aquí entran también herramientas de fuera del espacio de la gestión de tareas, porque
el problema es el mismo y las maduras llevan más tiempo tropezando con él.

**La conclusión, primero: al almacén se apunta con una ruta, no con un nombre.** Ninguna de las
herramientas miradas acepta el nombre legible de un almacén para elegirlo. Los nombres solo aparecen
donde existe un registro previo que los declara, como los contextos de `kubectl` o los sockets con
nombre de `tmux`, y siempre en una bandera distinta de la de la ruta, con una regla escrita que dice
cuál gana cuando se dan las dos, y gana la ruta. El segundo patrón es igual de consistente: la vía a
otro almacén se reparte en varias banderas estrechas, cada una con su variable de entorno, en vez de una
sola cadena que lo admita todo.

**git.** Tres mandos independientes: `-C` cambia el directorio de partida, `--git-dir` nombra el
almacén y `--work-tree` la copia de trabajo, y los dos últimos tienen su variable, `GIT_DIR` y
`GIT_WORK_TREE`. `-C` se aplica antes que los otros dos, así que sus rutas se resuelven contra él. La
propia documentación avisa de que dar `--git-dir` apaga el descubrimiento hacia arriba. No hay ninguna
forma de nombrar un repositorio local por un nombre: solo rutas.

**gh.** Una sola bandera, `-R, --repo`, con una sola gramática, `[HOST/]OWNER/REPO`. No admite ni una
URL, ni el nombre de un remoto, ni una ruta local. No tiene equivalente de `-C`, y lo piden desde 2020
en `https://github.com/cli/cli/issues/2228`, que sigue sin resolver. Qué gana entre la bandera y la
variable `GH_REPO` no está dicho en su documentación, así que aquí no se afirma.

**Taskwarrior.** Su `data.location` es una ruta, `TASKDATA` la sobrescribe, y la línea de comandos gana
sobre las dos con `rc.data.location=`. Lo que más importa para comparar es otra cosa: su concepto de
contexto **no elige almacén**. Es un filtro permanente sobre el mismo conjunto de datos, así que quien
viene de ahí puede esperar que cambiar de contexto cambie de almacén, cuando lo que cambia es la vista.

**Backlog.md.** No tiene ninguna bandera general para esto. Localiza por raíz de git más la carpeta
`backlog/`, y su directorio se fija al inicializar y a partir de ahí es de solo lectura. Sus tropiezos
están documentados y encajan con problemas que este documento ya cataloga:
`https://github.com/MrLesk/Backlog.md/issues/466` pide precisamente un argumento o una variable de
entorno para apuntar a otro sitio, `https://github.com/MrLesk/Backlog.md/issues/446` cuenta que
ejecutarlo en un subdirectorio creaba una carpeta nueva en vez de encontrar la de arriba, y
`https://github.com/MrLesk/Backlog.md/issues/558` y `https://github.com/MrLesk/Backlog.md/issues/689`
son su servidor escribiendo en el repositorio principal en vez de en el worktree desde el que se lanzó.

**Beads.** Tres vías, todas rutas, ordenadas por grano: `BEADS_DIR` fuerza el directorio del almacén y
apaga el descubrimiento, `BD_DB` apunta al fichero de base de datos, y `--db` lo sobrescribe para una
sola invocación. Sus fallos en esta zona son de la misma familia que los de Backlog.md:
`https://github.com/gastownhall/beads/issues/6222` es un `GIT_DIR` heredado del entorno que envenena la
resolución, y `https://github.com/gastownhall/beads/issues/6353` es su sincronización escribiendo en el
repositorio equivocado.

**El contraejemplo que mide lo que cuesta la cadena polimórfica es `restic`**, que sí tiene una sola
bandera, `-r`, capaz de admitir una ruta local, un servidor y varios servicios de almacenamiento. Lo
paga por escrito y en su propio código: si la cadena no es una ruta existente y contiene dos puntos hay
que decidir si eso es un esquema o parte de un nombre, y ese caso está marcado como ambiguo; hace falta
además un caso especial para las unidades de Windows, donde `C:` no es ningún esquema; y su mensaje de
error tiene que enseñar al usuario un prefijo `local:` que nadie escribiría por su cuenta, solo para
poder desambiguar lo que la gramática única no distingue. Una sola bandera que lo admite todo no ahorra
las reglas, las esconde dentro y las paga en mensajes de error.

**Dos precedentes de desempate.** `kubectl` separa `--kubeconfig`, que es una ruta, de `--context`, que
es un nombre, y documenta dos cadenas de precedencia distintas para cada uno; el nombre solo significa
algo dentro del fichero que la ruta selecciona. `tmux` separa `-L`, que es el nombre de un socket dentro
de su directorio, de `-S`, que es la ruta completa del socket, y documenta que dar `-S` hace que `-L`
se ignore. En los dos casos el nombre existe porque hay un registro previo donde buscarlo, y en los dos
la ruta gana.

---

# Parte 2. El catálogo de problemas

## 1. Dos copias de trabajo asignan el mismo identificador

**Quién lo sufre.** Backlog.md, por diseño y reconocido por su mantenedor. En
`https://github.com/MrLesk/Backlog.md/issues/711` lo explica así: la asignación secuencial no puede
hacerse segura en su topología objetivo, porque cada clon calcula el siguiente número contra lo que
puede ver, así que dos escritores creando tareas entre dos pushes acaban asignando el mismo
identificador de forma determinista. Y añade que lo sufrieron en producción dos veces, con el mismo
identificador en dos tareas distintas, apareciendo como conflictos al commitear y, peor, como
referencias ambiguas desde las dependencias. **El issue está cerrado como "no planeado".** El pull
request que sí llegó (`https://github.com/MrLesk/Backlog.md/pull/749`) dice explícitamente que se
detiene en el diagnóstico y la recuperación con intervención humana, y que la prevención no está
incluida.

Peor que el conflicto es el silencio. En `https://github.com/MrLesk/Backlog.md/issues/632` el mantenedor
narra que dos sesiones de agente en historias divergentes calcularon el mismo número, y como el nombre
de fichero incluye el título, los dos ficheros tenían nombres distintos: **git fusionó las dos ramas sin
ningún conflicto** y quedaron dos tareas con el mismo identificador sin que nada avisara. Su conclusión:
un bloqueo local previene la creación concurrente en una máquina, pero no ayuda cuando dos historias
independientes divergen y se fusionan después.

Spec Kit tiene la misma forma del problema con la numeración de funcionalidades: dos ramas generan el
mismo índice y al fusionar quedan duplicados
(`https://github.com/github/spec-kit/discussions/2116`), con la sugerencia de usar marcas de tiempo o
una convención de equipo.

Beads lo resolvió con identificadores aleatorios cortos, y pagó con no poder decir "la tarea 5" en voz
alta.

**Qué hace `biso`.** Un tablero es uno, con un solo asignador. Los identificadores siguen siendo
secuenciales y legibles, porque el fallo no viene de que sean secuenciales, viene de que cada copia de
trabajo asigne por su cuenta. Compartir entre máquinas más adelante se hace reservando rangos, y los
huecos que eso deja ya son legales por escrito en ["Identificadores"](spec/modelo-de-datos.md#identificadores).

## 2. El estado de las tareas se bifurca con la rama

**Quién lo sufre.** Cualquiera que guarde las tareas como ficheros del árbol de trabajo. La
formulación más clara está en un repaso de estas herramientas
(`https://nesbitt.io/2026/08/20/issues-in-the-repo.html`): crear una rama de trabajo bifurca el estado
de la incidencia junto con el código, así que **una incidencia cerrada en la rama vuelve a aparecer
abierta en el momento en que te cambias a la principal.**

Backlog.md tiene además un issue abierto sobre worktrees que no ven los cambios sin commitear hechos en
otro (`https://github.com/MrLesk/Backlog.md/issues/689`).

Las herramientas que lo esquivan sacando el estado del repositorio (Vibe Kanban, Conductor) cambian el
problema por el contrario: el estado deja de seguir a la rama, que a veces es lo que quieres y a veces
no.

**Qué hace `biso`.** Sacar el tablero del árbol de trabajo, a propósito y como decisión enunciada. Una
tarea cerrada está cerrada, no cerrada en esta rama. El precio, que el tablero no viaje al clonar en
otra máquina, se paga a cambio de que el estado sea uno.

## 3. Dos escrituras a la vez, y una se pierde sin decir nada

**Quién lo sufre.** Backlog.md mide en `https://github.com/MrLesk/Backlog.md/issues/843` que **doce de
doce ediciones concurrentes se pierden**, y que los dos procesos salen con código cero y los dos
imprimen que actualizaron la tarea.

Task Master tenía el mismo agujero: sus funciones de leer y escribir el JSON no tenían bloqueo alguno,
así que dos ventanas de agente escribiendo a la vez hacían que la última pisara a la anterior
(`https://github.com/eyaltoledano/claude-task-master/issues/1567`). Se arregló con una librería de
ficheros de bloqueo y escrituras atómicas.

Claude Code sufre lo mismo en su propio fichero de configuración, que varias sesiones truncan a mitad
de escritura (`https://github.com/anthropics/claude-code/issues/28973`).

**Qué hace `biso`.** Lo tenía escrito antes de elegir el mecanismo:
["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) exige que ninguna
escritura se observe a medias, que dos escrituras simultáneas sobre la misma tarea no se pierdan ni se
mezclen, y que si no se consigue el acceso exclusivo se espere hasta cinco segundos y se falle con
código 7 **sin escribir nada**. Una transacción de SQLite en modo WAL da exactamente eso.

## 4. Un bloqueo que no cruza máquinas, o que se queda huérfano

**Quién lo sufre.** Backlog.md endureció su bloqueo dos veces y sigue reconociendo el límite: el issue
`https://github.com/MrLesk/Backlog.md/issues/565` admite que el intento anterior mantenía el bloqueo
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

**Quién lo sufre.** Backlog.md, en `https://github.com/MrLesk/Backlog.md/issues/997`: crear una tarea
escanea las carpetas de activas y completadas pero no la de archivadas, así que el identificador vuelve
a estar libre y quedan dos ficheros distintos con el mismo identificador en carpetas distintas.

**Qué hace `biso`.** ["Identificadores"](spec/modelo-de-datos.md#identificadores) lo prohíbe expresamente: un identificador no se reutiliza jamás, y
el tablero sabe en todo momento cuál es el más alto que ha llegado a asignar, dato que se guarda aparte
de las tareas presentes. Eso es también lo que le permite dar
[tres mensajes distintos de "no la encuentro"](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), distinguiendo un identificador mal formado, uno que nunca existió, y uno
que existió y ya no está.

## 6. Una tarea se queda cogida porque la sesión murió

**Quién lo sufre.** Todo el mundo, y el vocabulario ya está inventado. Un proyecto describe su
arquitectura como un sistema de arrendamiento renovado por latido para evitar la doble reclamación
(`https://news.ycombinator.com/item?id=47341352`). Un artículo lo formaliza señalando que un bloqueo
simple se queda para siempre si su propietario se detiene, y propone un arrendamiento con límite de
tiempo **más un token de vallado que rechace las escrituras del propietario antiguo**
(`https://zenn.dev/agentmemories/articles/agent-task-lease?locale=en`). Hay incluso un issue con ese
vocabulario en el título (`https://github.com/xorbitsai/xagent/issues/1230`).

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
["Saber si alguien está trabajando de verdad"](DECISIONES.md#saber-si-alguien-está-trabajando-de-verdad) de `docs/DECISIONES.md` cuenta por qué se aparta a propósito,
y ["Riesgos conocidos y aceptados del modelo de estados"](DECISIONES.md#riesgos-conocidos-y-aceptados-del-modelo-de-estados) lo anota como riesgo aceptado.

## 7. El agente escribe mal el formato y se traga datos

**Quién lo sufre.** Backlog.md tiene tres issues abiertos que son el mismo problema por tres sitios. En
`https://github.com/MrLesk/Backlog.md/issues/990`, si un agente copia la descripción ya renderizada,
con sus marcadores de sección incluidos, y la reenvía como entrada, los marcadores se anidan y las
ediciones sucesivas se van tragando los criterios de aceptación y las notas, que desaparecen de la
salida. En `https://github.com/MrLesk/Backlog.md/issues/1000`, confundir banderas crea un criterio
fantasma en silencio. En `https://github.com/MrLesk/Backlog.md/issues/1008`, un subencabezado dentro de
una sección la trunca, porque el patrón que la extrae no distingue dos almohadillas de tres.

En Hacker News está la queja de fondo sobre editar texto con un agente: tiene problemas para editar
ficheros grandes y no puede añadir texto al final de uno grande por el tamaño de su ventana de
contexto.

**Qué hace `biso`.** Nadie escribe el almacén a mano, ni el agente ni la persona: se escribe por
comandos, y cada comando valida. Es el sitio donde la base de datos gana de forma más clara, porque no
hay un formato de texto que un agente pueda malinterpretar al reenviarlo.

Con una honestidad al lado: **con un fichero por tarea, aislar el daño de una tarea corrupta es gratis**,
y ["Qué pasa con un dato que no se puede interpretar"](spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar) está escrita pensando en eso. Con una base de datos hay que separar dos
casos que hoy se dicen como uno: una tarea cuyo contenido no se interpreta, y una base de datos que no
abre.

## 8. Un dato que la herramienta no conoce se pierde al editar

**Quién lo sufre.** Backlog.md, en `https://github.com/MrLesk/Backlog.md/issues/918`, etiquetado como
bug: los ficheros de tarea son markdown por diseño, lo que invita a que otras herramientas añadan sus
propias claves, y hoy cualquier clave así se pierde en silencio en la siguiente edición. Se reproduce
añadiendo una clave a mano y haciendo una edición trivial.

**Qué hace `biso`.** Los campos externos son un mecanismo declarado, no un hueco: el tablero declara
qué claves admite, y una tarea con una clave que la configuración ya no declara **no se lee en silencio
ni se reescribe perdiéndola**, sino que sigue la regla de
["Qué pasa con un dato que no se puede interpretar"](spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar) y `biso doctor` la reporta.

## 9. Estado compartido que cada invocación pisa

**Quién lo sufre.** Spec Kit, en `https://github.com/github/spec-kit/issues/4128`: un fichero
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
copia completa en vez de un cambio (`https://ongardie.net/blog/sqlite-in-git`). Git LFS no ayuda,
porque resuelve el peso del repositorio y no la fusionabilidad. Litestream tampoco, porque es
continuidad operativa con un único escritor primario, no colaboración. La respuesta seria del espacio
es Dolt, que sustituye el motor por uno con fusión a nivel de celda, y es lo que eligió Beads.

**La mitad de texto.** Un fichero por tarea no tiene transacciones: el renombrado atómico lo da el
sistema operativo para un fichero, no para un directorio, así que el todo o nada sobre doscientas
tareas repartidas en doscientos ficheros hay que construirlo. Y un solo fichero de log fusionable tiene
su propia trampa: decirle a git que fusione por unión de líneas **resucita las líneas borradas**, porque
concatena la unión de lo que hay en los dos lados y solo quita duplicados exactos, sin razonar sobre qué
falta ni por qué
(`https://stackoverflow.com/questions/12947436/git-union-merge-brings-back-some-deleted-lines`). GitLab
y scikit-learn lo usan para sus registros de cambios y documentan sus efectos, incluidas entradas
duplicadas en dos secciones (`https://github.com/scikit-learn/scikit-learn/issues/21516`).

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
sus banderas sugerían que hacía mucho más, lo que generaba expectativas equivocadas. **Se eliminó por
completo** al cambiar de motor. Y el comentario más citado de quien lo reemplazó por algo más simple
dice que empezó a pelearse con él varias veces por semana porque su daemon de fondo se puso a
sincronizar las cosas equivocadas (`https://news.ycombinator.com/item?id=46487580`). En el mismo hilo,
otro lo llama un lío incomprensible para algo que en el fondo es una idea simple.

Y su documentación avisaba de que **el modo daemon no funcionaba correctamente con los worktrees de
git** por el estado de base de datos compartido.

Cuando su sucesor exigió un servidor externo, los autores del motor publicaron una marcha atrás
reconociendo la fricción que eso añadía a quien trabaja solo.

**Qué hace `biso`.** No hay daemon, y el motivo es aritmético antes que estético. El coste dominante de
una invocación es arrancar un proceso, y el cliente que hablaría por el socket también es un proceso, así
que el daemon compite por uno o dos milisegundos de unos ocho pagando con una arquitectura entera. Un
daemon colgado, además, haría fallar también las lecturas, que
["Concurrencia, atomicidad y garantías observables"](spec/garantias.md#concurrencia-atomicidad-y-garantías-observables) promete que nunca fallan
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
(`https://replit.com/blog/golang-performance`). El pecado es hacer al arrancar trabajo que nadie pidió.

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
(`https://github.com/MrLesk/Backlog.md/issues/821`). MCP Shrimp Task Manager commitea tras cada cambio
en su propio directorio de datos.

**Qué hace `biso`.** El paso de exportar y commitear es explícito y **nunca está en el camino caliente
de un comando**, que además de evitar el ruido evita pagar los 12 ms de `git` en cada invocación. La
granularidad del historial y por tanto del rollback es la frecuencia con la que se toma la instantánea,
y eso se dice en voz alta en vez de disimularse.

## 14. Un valor que no existe devuelve una lista vacía en vez de un error

**Quién lo sufre.** Backlog.md, en `https://github.com/MrLesk/Backlog.md/issues/919`: pasar un hito que
no coincide con nada **crea un hito virtual en silencio**. Y en
`https://github.com/MrLesk/Backlog.md/issues/824`, la bandera `--json` estaba documentada en el README
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
(`https://www.tamirdresher.com/blog/2026/03/21/rate-limiting-multi-agent`), sin ningún mecanismo de
prioridad entre el agente que sondeaba en segundo plano y el que necesitaba la cuota para trabajo
crítico.

Y hay una queja de otra naturaleza: no se distingue lo que escribió un humano de lo que escribió un
agente. Una discusión formal pide poder bloquear issues generados por Copilot y dice que no hay ninguna
indicación visible de que fuera generado, ni en la interfaz ni en la interfaz de programación
(`https://github.com/orgs/community/discussions/159749`).

**Qué hace `biso`.** Es local y no tiene cuota. Y distingue quién escribió qué **mientras cada quien
tenga su propia identidad**, con autor en los comentarios y en la pregunta abierta. Esa condición hay que
decirla, porque un tablero con la clave `me` configurada la destruye: `me` gana sobre `BISO_ME`, así que
ahí todo el mundo comparte identidad y la distinción deja de existir, lo que
["Riesgos conocidos y aceptados del modelo de estados"](DECISIONES.md#riesgos-conocidos-y-aceptados-del-modelo-de-estados) anota como riesgo aceptado. Un tablero compartido entre una persona y un agente tiene que
dejar `me` sin configurar, y de la vía por la que esa clave llegaba sin que nadie la eligiera, restaurar
la instantánea de otra persona, ya se encarga `biso snapshot`, que no la escribe.

## 16. El agente deja de mirar el tablero cuando el proyecto crece

**Quién lo sufre.** Task Master, según un comentario en el hilo de Backlog.md: estaba bastante bien,
pero a veces el agente lo ignoraba a medida que el proyecto crecía
(`https://news.ycombinator.com/item?id=44486442`). Y con ficheros markdown sueltos la queja es la
inversa y complementaria: hay riesgo de que los agentes tropiecen con tickets obsoletos y consuman
tokens (`https://news.ycombinator.com/item?id=46509745`).

**Qué hace `biso`.** Es el trabajo de `biso prime`: un mensaje de arranque de tamaño acotado por
contrato, que cabe entero en el contexto y dice lo que hace falta para empezar bien, en vez de dejar que
el agente decida cuánto tablero leer.

---

# Parte 3. Lo que la gente quiere conservar, y lo que rechaza

De todo el ruido salen tres cosas que nadie quiere perder: **el grafo de dependencias**, **poder dejar
fuera del listado lo que no se puede coger ahora**, y **poder fichar una tarea de forma atómica**. Las
tres están ya en la especificación de `biso` como `dependencies`, como los filtros que se combinan en
`biso ls --not-blocked --not-waiting`, y como `biso start`. La segunda no es una sola bandera a
propósito: ninguna puede decir por sí misma que una tarea esté lista, porque cuántos filtros hace falta
descartar depende de qué se busque, y ["El porqué de reglas concretas"](DECISIONES.md#el-porqué-de-reglas-concretas) cuenta por qué se retiró el
nombre `--ready`, que lo prometía sin poder cumplirlo.

Y dos que rechaza de forma consistente: **el proceso en segundo plano** y **el almacén opaco**. El
primero está descartado. El segundo es la tensión real de esta decisión, y la respuesta es que el
almacén no sea opaco por fuera aunque sea una base de datos por dentro: hay una exportación fiel y
probada, versionada en el propio directorio del tablero, y `biso where` explica siempre qué tablero se
usó y por qué.

El ciclo del espacio se ha completado una vez en menos de un año. Los ficheros markdown sueltos
generaron Beads para darles dependencias y un grafo. La complejidad de Beads generó un reemplazo
deliberadamente simple, de vuelta al texto plano pero conservando el grafo. Y el motor de Beads tuvo que
publicar una marcha atrás para recuperar a quien trabaja solo. Merece la pena tenerlo presente antes de
añadir cualquier cosa.
