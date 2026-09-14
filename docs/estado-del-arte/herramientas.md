# Parte 1. Las herramientas

## Backlog.md

<https://github.com/MrLesk/Backlog.md>, 6.664 estrellas, 402 forks, 48 issues abiertos, TypeScript,
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

<https://github.com/gastownhall/beads> (antes `steveyegge/beads`), 26.963 estrellas, más de 1.000
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
(<https://raw.githubusercontent.com/gastownhall/beads/main/docs/core-concepts/sync-concepts.md>). Y, en
la época del daemon, que **el modo daemon no funcionaba correctamente con los worktrees de git** por el
estado de base de datos compartido.

**Cómo terminó.** El cambio a Dolt costó la atomicidad de commitear código y tareas juntos
(<https://github.com/gastownhall/beads/issues/2489>), y el 2 de abril de 2026 DoltHub publicó una marcha
atrás parcial, "Restoring Beads Classic"
(<http://www.dolthub.com/blog/2026-04-02-restoring-beads-classic/>), reconociendo que el paso a Dolt
añadió fricción a los usuarios en solitario, acostumbrados al modelo más simple de SQLite y git que no
exigía un servidor externo. El daemon se eliminó por completo.

## Task Master (`claude-task-master`)

<https://github.com/eyaltoledano/claude-task-master>, 28.056 estrellas, 2.623 forks, 212 issues
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

<https://github.com/github/spec-kit>, MIT, anunciado el 2 de septiembre de 2025. Estrellas muy altas
pero sin verificar con precisión.

Desarrollo dirigido por especificaciones, con el flujo de especificar, planificar, generar tareas e
implementar. Todo son ficheros versionados en `.specify/` y `specs/`, numerados secuencialmente por
funcionalidad. Sin base de datos, sin servidor de protocolo: interviene con comandos de barra dentro del
agente anfitrión.

## Vibe Kanban

<https://github.com/BloopAI/vibe-kanban>, 28.030 estrellas, 3.002 forks, 539 issues abiertos. **El
equipo lo abandonó el 10 de abril de 2026** (<https://www.vibekanban.com/blog/shutdown>): miles de
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
despliegue autohospedado (<https://ampcode.com/security>). No hay exportación de una conversación a
markdown ni a JSON. Las quejas públicas encontradas son todas sobre precio.

## GitHub Issues como tablero

La acción oficial de Anthropic se dispara con una mención en un issue o comentario
(<https://code.claude.com/docs/en/github-actions>), y el agente de Copilot funciona asignándole el issue
como a un compañero. La ventaja que la gente destaca es que no hay que construir infraestructura de
coordinación, porque los tickets ya sirven de memoria entre invocaciones y de bus de mensajes vía
comentarios.

Un detalle revelador: la acción oficial **rechaza por defecto a los actores que son bots, para evitar
que un bot dispare a Claude en un bucle.** El propio fabricante reconoce que un tablero compartido es
también un bucle de retroalimentación compartido.

## Ficheros `TODO.md` y `PLAN.md` sueltos

La opción de cero herramientas, con su especificación de facto en
<https://github.com/todomd/todo.md>. Su uso más citado es recuperar contexto al agotar la ventana. Sus
dos quejas también están documentadas, y están en la parte 2.

## Servidores de protocolo genéricos

El registro oficial ya no mantiene catálogo de terceros. Con tracción propia se encontraron **MCP
Shrimp Task Manager** (<https://github.com/cjo4m06/mcp-shrimp-task-manager>, 2.153 estrellas), que
guarda en `tasks.json` y hace algo poco común, inicializar un repositorio de git dentro de su propio
directorio de datos y commitear tras cada cambio; y **taskboard**
(<https://github.com/tcarac/taskboard>, 29 estrellas), explícitamente SQLite y binario único.

## Cómo se apunta a un almacén distinto del que la herramienta encuentra sola

Todo lo anterior dice dónde guarda los datos cada herramienta. Esto es lo otro: cómo se le dice a una
herramienta que trabaje con un almacén que no es el que habría encontrado por su cuenta. Investigado el
2026-09-09 sobre documentación oficial, y sobre el código en el único caso en que la documentación no
llega, `restic`. Aquí entran también herramientas de fuera del espacio de la gestión de tareas, porque
el problema es el mismo y las maduras llevan más tiempo tropezando con él.

**La conclusión, primero: al almacén se apunta con una ruta, no con un nombre.** Ninguna de las
herramientas miradas acepta el nombre legible de un almacén para elegirlo. Los nombres solo aparecen
donde existe un registro previo que los declara, como los contextos de `kubectl` o los sockets con
nombre de `tmux`, y siempre en un flag distinto del de la ruta, con una regla escrita que dice
cuál gana cuando se dan las dos, y gana la ruta. El segundo patrón es igual de consistente: la vía a
otro almacén se reparte en varios flags estrechos, cada uno con su variable de entorno, en vez de una
sola cadena que lo admita todo.

**git.** Tres mandos independientes: `-C` cambia el directorio de partida, `--git-dir` nombra el
almacén y `--work-tree` la copia de trabajo, y los dos últimos tienen su variable, `GIT_DIR` y
`GIT_WORK_TREE`. `-C` se aplica antes que los otros dos, así que sus rutas se resuelven contra él. La
propia documentación avisa de que dar `--git-dir` apaga el descubrimiento hacia arriba. No hay ninguna
forma de nombrar un repositorio local por un nombre: solo rutas.

**gh.** Un solo flag, `-R, --repo`, con una sola gramática, `[HOST/]OWNER/REPO`. No admite ni una
URL, ni el nombre de un remoto, ni una ruta local. No tiene equivalente de `-C`, y lo piden desde 2020
en <https://github.com/cli/cli/issues/2228>, que sigue sin resolver. Qué gana entre el flag y la
variable `GH_REPO` no está dicho en su documentación, así que aquí no se afirma.

**Taskwarrior.** Su `data.location` es una ruta, `TASKDATA` la sobrescribe, y la línea de comandos gana
sobre las dos con `rc.data.location=`. Lo que más importa para comparar es otra cosa: su concepto de
contexto **no elige almacén**. Es un filtro permanente sobre el mismo conjunto de datos, así que quien
viene de ahí puede esperar que cambiar de contexto cambie de almacén, cuando lo que cambia es la vista.

**Backlog.md.** No tiene ningún flag general para esto. Localiza por raíz de git más la carpeta
`backlog/`, y su directorio se fija al inicializar y a partir de ahí es de solo lectura. Sus tropiezos
están documentados y encajan con problemas que este documento ya cataloga:
<https://github.com/MrLesk/Backlog.md/issues/466> pide precisamente un argumento o una variable de
entorno para apuntar a otro sitio, <https://github.com/MrLesk/Backlog.md/issues/446> cuenta que
ejecutarlo en un subdirectorio creaba una carpeta nueva en vez de encontrar la de arriba, y
<https://github.com/MrLesk/Backlog.md/issues/558> y <https://github.com/MrLesk/Backlog.md/issues/689>
son su servidor escribiendo en el repositorio principal en vez de en el worktree desde el que se lanzó.

**Beads.** Tres vías, todas rutas, ordenadas por grano: `BEADS_DIR` fuerza el directorio del almacén y
apaga el descubrimiento, `BD_DB` apunta al fichero de base de datos, y `--db` lo sobrescribe para una
sola invocación. Sus fallos en esta zona son de la misma familia que los de Backlog.md:
<https://github.com/gastownhall/beads/issues/6222> es un `GIT_DIR` heredado del entorno que envenena la
resolución, y <https://github.com/gastownhall/beads/issues/6353> es su sincronización escribiendo en el
repositorio equivocado.

**El contraejemplo que mide lo que cuesta la cadena polimórfica es `restic`**, que sí tiene una sola
flag, `-r`, capaz de admitir una ruta local, un servidor y varios servicios de almacenamiento. Lo
paga por escrito y en su propio código: si la cadena no es una ruta existente y contiene dos puntos hay
que decidir si eso es un esquema o parte de un nombre, y ese caso está marcado como ambiguo; hace falta
además un caso especial para las unidades de Windows, donde `C:` no es ningún esquema; y su mensaje de
error tiene que enseñar al usuario un prefijo `local:` que nadie escribiría por su cuenta, solo para
poder desambiguar lo que la gramática única no distingue. Un solo flag que lo admite todo no ahorra
las reglas, las esconde dentro y las paga en mensajes de error.

**Dos precedentes de desempate.** `kubectl` separa `--kubeconfig`, que es una ruta, de `--context`, que
es un nombre, y documenta dos cadenas de precedencia distintas para cada uno; el nombre solo significa
algo dentro del fichero que la ruta selecciona. `tmux` separa `-L`, que es el nombre de un socket dentro
de su directorio, de `-S`, que es la ruta completa del socket, y documenta que dar `-S` hace que `-L`
se ignore. En los dos casos el nombre existe porque hay un registro previo donde buscarlo, y en los dos
la ruta gana.
