# Plan: procesar las notas de lectura del 2026-09-13

Worktree: `.claude/worktrees/notas-lectura`, rama `worktree-notas-lectura`. Línea base: `make docs-doctor`
pasa entero antes de empezar.

Dos fases. La fase 1 solo renombra o mueve: no cambia lo que la spec dice que hace el programa. La
fase 2 mejora el texto. Cada fase termina con `make docs-doctor` en verde y un commit.

## Reglas para todos los pasos

- Documentación en español, código (Go, Python) entero en inglés. Nunca em-dash.
- `docs/tutorial/*.md` y `docs/concepts.md` son generados: se editan sus fuentes
  (`docs/docs-tooling/tutorial/escenarios/*.yaml`, `docs/docs-tooling/tutorial/conceptos.md`) y se
  regeneran con `make docs-build`.
- Las tareas de `backlog/` no se editan a mano: solo con el CLI `backlog`.
- Un renombrado de fichero o de título cambia rutas y anclas: hay que actualizar todos los enlaces
  que apunten a ellos, `docs/docs-tooling/mkdocs/mkdocs.yml` (nav) y
  `docs/docs-tooling/tools/mapa-de-secciones.txt`. `make docs-doctor` lo comprueba.
- Si un paso encuentra una contradicción de la spec que no es mecánica, no la resuelve: la anota en
  su informe final.
- Los pasos de la fase 1 se ejecutan en serie, en este orden, porque tocan los mismos ficheros.

## Fase 1: renombrar y mover

### 1.1 Estructura de páginas

a. **Flags globales a su propia página** (nota 19). La sección "Banderas globales" de
   `docs/spec/invocacion.md`, con su tabla y sus reglas de aplicación, pasa a
   `docs/spec/cmd/flags-globales.md`, con título `# Flags globales`, primera entrada bajo "Los
   comandos" en el nav. `invocacion.md` queda con "Variables de entorno" y "Configuración de máquina"
   subidas a `##`, y título `# Entorno y configuración de máquina`. Actualizar todos los enlaces a
   `invocacion.md#banderas-globales` y los recuentos de `docs/spec/cmd/index.md` si cuentan páginas.
b. **Modelo de datos** (notas 23, 24, 26, 36). En `docs/spec/modelo-de-datos/index.md`: borrar el
   párrafo "Las tablas siguientes agrupan los campos según quién decide su valor..."; el título "Los
   campos que fija el programa" pasa a "Campos automáticos" y "Los campos que fija quien llama" a
   "Campos fijados por el usuario". `identificadores.md` cambia su título a "Identificador de tarea".
   El orden del nav bajo el modelo de datos queda: Identificador de tarea, Las fechas, La urgencia, El
   autor de una tarea (ver 1.2), Los criterios y sus claves estables, La pregunta abierta, Los
   comentarios, Los campos externos. Si `index.md` enumera las subpáginas, mismo orden.
c. **Ficheros con "banderas" en el nombre**: `docs/spec/familias-de-banderas.md` pasa a
   `docs/spec/familias-de-flags.md` y `docs/decisiones/comandos-y-banderas.md` a
   `docs/decisiones/comandos-y-flags.md` (con `git mv`). Los títulos se cambian en 1.4.

### 1.2 `reporter` pasa a `author` (nota 35)

- Campo `reporter` a `author`, flag `--reporter` a `--author`, en spec, decisiones, tutorial
  (fixtures y lagunas), y en `bench/sqlite-driver/` (`cmd/gen/main.go`, `internal/board/schema.go`,
  y lo que compile con ellos: `go build ./...` y `go vet ./...` en `bench/sqlite-driver` deben pasar).
- `docs/spec/modelo-de-datos/quien-reporta.md` pasa a `autor.md`, título "El autor de una tarea".
- En prosa, "quién reporta", "el que reporta", "reportar una tarea" pasan a "el autor de la tarea" o
  "crear la tarea" según el caso. Como comentarios y preguntas ya tienen su propio `author`, la prosa
  dice siempre "el autor de la tarea" frente a "el autor de un comentario" o "de la pregunta",
  nunca "el autor" a secas si puede confundirse.
- Comprobar que no existe ya un `--author` con otro significado. No existe hoy: si aparece, parar e
  informar.

### 1.3 Renumerar los códigos de salida (nota 14, opción B)

| Nombre | Antes | Después |
|---|---:|---:|
| `OK` .. `PRECONDITION` | 0-6 | 0-6, sin cambio |
| `VALIDATION` | 9 | 7 |
| `ENVIRONMENT` | 7 | 8 |
| `NO_BOARD` | 8 | 20 |
| `DAMAGED` | 10 | 21 |
| `AMBIGUOUS_BOARD` | 11 | 22 |

- La tabla de `docs/spec/codigos-de-salida.md` en el orden nuevo, y una frase que diga que los
  códigos del 20 en adelante son los desenlaces de resolver el tablero.
- Cada mención de un código 7, 8, 9, 10 u 11 en todo el repositorio se revisa **leyendo su
  contexto**: tablas de códigos de cada comando, "sale con", "error 9", "código 8", `"exitCode": 9`,
  claves de salida de los fixtures del tutorial, `bench/`. Un 7 o un 9 que no es código de salida
  (número de línea, recuento, tiempo) no se toca. Hacer el cambio en un solo recorrido con el mapa de
  arriba, nunca en cadena (si se cambia 9 a 7 y luego 7 a 8, el 9 original acaba en 8).
- Las tablas de códigos de cada comando se reordenan por número.
- Al terminar, listar cada fichero tocado con el número de sustituciones, para la revisión.

### 1.4 `bandera` pasa a `flag` en todo el repositorio (nota 12)

- Todo el repositorio: `docs/` entero (incluidos `decisiones/`, `estado-del-arte/`, `superpowers/`),
  `CLAUDE.md` y los demás `CLAUDE.md`, `docs/docs-tooling/` (fixtures, lagunas, `mkdocs.yml`,
  `tools/*.py`, `tools/*.txt`), `bench/` y las tareas de `backlog/`.
- "flag" es masculino: "la bandera global" a "el flag global", "banderas incompatibles" a "flags
  incompatibles", "una bandera de escritura" a "un flag de escritura", "esa bandera" a "ese flag".
  Revisar artículos, demostrativos, adjetivos y participios de la frase ("la bandera está
  definida" a "el flag está definido"). Nada de "flagear".
- Títulos: "Las familias de banderas" a "Las familias de flags", "Comandos y banderas" a "Comandos y
  flags", etc. Con ellos cambian anclas (`#...banderas...`), que hay que seguir hasta sus enlaces.
- Los comprobadores de `docs/docs-tooling/tools/` pueden buscar la palabra literal (por ejemplo
  `comprobar_recuentos.py` y `recuentos-normativos.txt`): actualizar el patrón y comprobar que
  siguen encontrando lo mismo. Los ficheros Python quedan en inglés.
- **Tareas de backlog** (12 ficheros en `backlog/tasks/`, la mayoría por la ruta
  `familias-de-banderas.md`): por cada una, `backlog task view <id> --plain`, y cambiar solo los
  campos que contengan la palabra con `backlog task edit`: `-t` (título), `-d` (descripción),
  `--plan`, `--notes`, `--final-summary`. Si un criterio de aceptación la contiene, sustituir la
  lista con `--acceptance-criteria` repetido y volver a marcar con `--check-ac` los que estaban
  marcados. Si está en un comentario, que el CLI no edita, anotarlo en el informe. Nunca editar el
  `.md` de la tarea a mano. `auto_commit` hará un commit por edición: es lo esperado.
- Comprobación final: `grep -rniI "bandera" . --exclude-dir=.git --exclude-dir=site` no devuelve
  nada salvo `notas-pendientes.md` y este plan, que nombra la palabra vieja a propósito.

### 1.5 "Tarjeta" y "ticket" (nota 11)

- Quitar la viñeta de `docs/spec/vocabulario.md` sobre **tarjeta** y **ticket**.
- En `docs/spec/cmd/board.md`, "tarjetas" pasa a "tareas".
- `docs/decisiones/modelo-de-estados.md` se reescribe en la fase 2 (paso 2.1).

### 1.6 Que no se corten palabras (nota 17)

- En `docs/stylesheets/extra.css`, que el contenido parta líneas solo entre palabras: comprobar qué
  regla de Material provoca el corte (por ejemplo `word-break` o `overflow-wrap: anywhere` en tablas o
  en `code`) y anularla con `overflow-wrap: normal; word-break: normal; hyphens: manual;` en el ámbito
  que toque, sin romper el scroll horizontal de tablas y bloques de código.

### Cierre de la fase 1

`make docs-doctor` en verde, `go build ./... && go vet ./...` en `bench/sqlite-driver`, revisión con
agente, commit.

## Fase 2: mejorar el texto

Los pasos son independientes entre sí (tocan ficheros distintos) y pueden ir en paralelo.

### 2.1 Índice, vocabulario y estados

- `docs/spec/index.md` (nota 7): quitar la frase "El orden de la barra lateral no es alfabético..."
  y cualquier otra mención a "el orden de la barra lateral" (línea ~87).
- `docs/spec/vocabulario.md` (nota 8): en la fila **persona**, la columna "En la interfaz" nombra
  los campos que llevan un texto de persona: `assignees`, `author` y el `author` de comentarios y
  preguntas, y la descripción aclara que no es un campo sino el tipo de valor que llevan esos campos.
  En la fila **criterio**, la descripción dice para qué sirve cada lista: una comprueba si el trabajo
  cumple lo pedido (`acceptanceCriteria`) y la otra si la tarea está lista para cerrarse
  (`definitionOfDone`). Redacción propia, no la de la nota.
- `docs/spec/vocabularios.md` (nota 10): en la tabla del algoritmo de coincidencia, quitar las filas
  de valores que no existen en el tablero (`Pending`, `Todos`) y usar solo ejemplos de escrituras
  distintas del mismo valor que la normalización resuelve. Si hace falta mostrar un fallo, que sea uno
  de escritura (por ejemplo `To Do.` o `To-Do-`, si según el algoritmo no se normalizan igual);
  verificar cada fila contra el algoritmo que hay justo encima. La segunda tabla ("El mismo texto vale
  lo mismo en los dos sentidos") conserva su fila de error, porque ahí es el contrato del mensaje.
- `docs/decisiones/modelo-de-estados.md`: reescribir el pasaje de "arrastrar una tarjeta desde el
  móvil" para que diga el contexto real: un proyecto en el que Backlog.md se sincronizaba con Trello,
  y Trello hacía de interfaz en el móvil, donde se arrastraba la tarjeta y el cambio llegaba al
  tablero. Dejar explícito que `biso` no tiene arrastre ni columnas que mover. Seguir la regla de
  `CLAUDE.md` sobre cómo se escribe una decisión.

### 2.2 Códigos de salida, flags globales y entorno

- `docs/spec/codigos-de-salida.md` (nota 13): el párrafo de los códigos del tablero se reescribe como
  una frase de entrada y una lista, un elemento por código (20, 21, 22) con qué significa y cuál es su
  remedio, y después, en frases cortas, por qué son tres y por qué no están en la tabla de cada
  comando (con la excepción de `biso where`).
- `docs/spec/cmd/flags-globales.md` (nota 15): dos tablas. Una de flags con valor (`--cwd`,
  `--color`), con su columna "Por defecto". Otra de flags booleanos, sin columna de valor por defecto,
  precedida de una frase que diga qué significa booleano aquí: poner el flag lo activa, no ponerlo lo
  deja desactivado, y no admite valor.
- Misma página (nota 16): la regla "Ninguna de las dos se ignora nunca en silencio" nombra
  explícitamente `--print` y `--dry-run`, y se entiende sin leer las siguientes.
- `docs/spec/invocacion.md` (nota 18): comprobar la tabla "Qué pasa si no hay identidad" contra cada
  comando que cita (`--mine` en `ls`, `biso start`, `biso comment`, `biso ask`, `biso answer`, el
  autor de la tarea en `biso new`), mirando mensaje, código de salida y comportamiento. Si coinciden,
  no se toca. Si no, informar de cada discrepancia con las dos citas, sin resolverla.

### 2.3 Resolución del tablero (notas 20 y 21)

- `docs/spec/resolucion-del-tablero.md`: reorganizar en secciones y subsecciones con sentido, sin
  quitar ninguna regla normativa. Añadir al principio una sección breve sobre cómo nace un tablero
  (`biso init`, el fichero puntero en el proyecto, el directorio del tablero fuera de él), que remita a
  `cmd/init.md` para el detalle en vez de duplicarlo. Añadir ejemplos con árboles de directorios (en
  bloques de código) para los casos principales: el proyecto con su puntero, el tablero en su raíz,
  un subdirectorio que resuelve hacia arriba, la ambigüedad del código 22.
- Todas las anclas que cambien se siguen hasta sus enlaces, y `mapa-de-secciones.txt` se actualiza.

### 2.4 Modelo de datos

- `index.md`, tabla "Campos automáticos" (nota 25): la columna "Obligatorio" pasa a "Nullable".
- Tabla "Campos fijados por el usuario" (nota 27): se quitan las columnas "Obligatorio" y "Mutable" y
  se añade una columna "Notas", vacía salvo en `title` y `status` (obligatorios) y en `comments` y
  `question` (cómo se modifican, con el texto que hoy está en "Mutable").
- "Precisiones generales sobre la mutabilidad" (nota 28): reformular para que no remita a una columna
  que ya no existe.
- La frase "¿Tiene fecha y autor propios, y se puede señalar uno a uno? Solo se aplica a los tipos con
  nombre..." (nota 22): reescribir en lenguaje llano. "Tipos con nombre" son `Criterion`, `Comment` y
  `Question`; "señalar uno a uno" es que tenga una clave para referirse a un elemento concreto de la
  lista. Decirlo así.
- `comentarios.md` (nota 29): dejar claro que el orden es el de inserción, que cambiar la fecha de un
  comentario con `--set-comment-date` no lo mueve de sitio en la lista, y que `biso answer` añade al
  final un comentario con fecha pasada.
- `fechas.md` (notas 30, 32, 33): tres secciones. Una con `task.createdAt`, `task.updatedAt` y
  `comment.createdAt`; otra con `question.askedAt`; otra con `leaseExpiresAt`. Los campos llevan
  siempre su prefijo de objeto en esta página. Comprobar el nombre real del instante de un comentario
  en el modelo antes de escribir `comment.createdAt`.
- `campos-externos.md` (nota 34): `ext` guarda cualquier valor que el usuario quiera asociar a la
  tarea con una clave; la identidad en otro sistema (Trello) es un ejemplo, no su definición.

### Cierre de la fase 2

`make docs-doctor` en verde, revisión con agente, commit, informe con las discrepancias de 2.2 si las
hay.

## Precisiones tras la validación del plan

Dos agentes revisaron el plan contra el repositorio. Lo que sigue manda sobre lo anterior donde lo
precise.

### Fase 1

- **Enlaces a rutas renombradas también en los fixtures.** Los YAML de
  `docs/docs-tooling/tutorial/escenarios/` llevan rutas y anclas escritas a mano
  (`spec/familias-de-banderas.md#...` en 07, 09 y 10; `spec/invocacion.md#banderas-globales` en 13).
  Todo renombrado de 1.1 y 1.2 los incluye, no solo `mkdocs.yml` y `mapa-de-secciones.txt`.
- **`identificadores.md#identificadores`** tiene 11 enlaces entrantes (spec, decisiones, estado del
  arte). Al cambiar el título a "Identificador de tarea", el ancla pasa a `#identificador-de-tarea` y
  hay que seguirlos todos.
- **Subtítulos con la palabra vieja citados desde fuera**, por ejemplo
  `## Sustituir un campo que no tiene bandera de "sustituir entera"` en `familias-de-banderas.md`,
  enlazado desde `escenarios/07-falta-un-criterio.yaml`. Antes de cambiar un título, buscar su ancla.
- **`doctor.py` no pasa los comprobadores por `docs/spec/modelo-de-datos/*.md`** (sus `glob` de las
  líneas ~74-90). Se añade ese directorio a los dos `glob`, y lo que salga por ello se arregla si es
  mecánico o se informa si no.
- **`mapa-de-secciones.txt`**: solo se actualizan ruta, ancla y título de las filas cuya página o
  título cambia en esta fase, incluidas las de modelo de datos que hoy apuntan a
  `modelo-de-datos.md#...` si el comprobador lo exige. Los números de sección no se renumeran.
- **1.2**: además de las frases citadas, buscar `report` en todas sus formas ("quien la reporte",
  "reportante"...) y decidir por contexto. Ya existe `--comment-author`; no choca, pero en las tablas de
  los verbos del ciclo conviven los dos y la descripción de cada uno debe dejar claro de quién es. La
  línea del nav de 1.1b se completa aquí con la ruta `autor.md`.
- **1.3, falsos positivos**: la tabla `| 7 | ... | | 8 | ... |` de `docs/spec/cmd/ls.md` (~121) son
  números de columna, no códigos. No se toca.
- **1.3, formas sin la palabra "código"**: "el 8" a secas en `docs/decisiones/persistencia.md` (~364),
  "devuelve 8" dentro de una celda en `codigos-de-salida.md`, la tabla número a claves JSON de
  `docs/spec/contrato-json.md` (~87-91). Todas entran.
- **1.3, alcance**: `docs/superpowers/` queda fuera de la renumeración. Son planes y diseños ya
  ejecutados que describen el estado de entonces. El cambio de palabra de 1.4 sí los cubre.
- **1.4**: `comprobar_recuentos.py` (~33) tiene `banderas` en `SUSTANTIVOS`: pasa a `flags`.
  `recuentos-normativos.txt` (~13) declara la excepción "tres banderas", que aparece en
  `cmd/init.md` (dos veces) y `decisiones/comandos-y-banderas.md`: las tres y la excepción cambian a la
  vez.

### Fase 2

- **2.1, ejemplos de fallo por escritura verificados** contra el algoritmo: `To Do.` normaliza a
  `todo.` y `To.Do` a `to.do`, y ninguno coincide con `todo`. `To-Do-` normaliza a `todo` y **sí
  coincide**: no vale como ejemplo de fallo.
- **2.1**: la historia de la tarjeta solo está en `docs/decisiones/modelo-de-estados.md` (~36-44).
- **2.2, booleanos**: `--json`, `--quiet`, `--print` y `--dry-run` son interruptores (sin ponerlos,
  desactivados). `--version` y `--help` son acciones que imprimen y salen. La frase de la tabla de
  booleanos lo distingue, o se ponen en una tercera agrupación dentro de la misma tabla con su nota.
- **2.2, nota 16**: la frase ya nombra `--print` y `--dry-run`, pero al final; quien lee "Ninguna de
  las dos" no sabe de cuáles habla. Reescribirla para que los nombre al principio.
- **2.2, identidad**: la comprobación ya está hecha. Coinciden `ls --mine`, `start`, `comment`, `ask` y
  `answer`. **Hay una discrepancia**: `invocacion.md` (~69) dice que `biso prime` imprime
  `you are (not set)` "con una nota que remite a `biso config set me`", y `cmd/prime.md` (~81-82 y
  ~198-200) dice que `prime` no escribe nada por stderr y no menciona ninguna nota. No se resuelve en
  esta fase: se informa al usuario. `biso new` no está en esa tabla a propósito (sin identidad, el autor
  queda vacío sin aviso) y no se añade.
- **2.3**: el fichero no tiene hoy ningún `##`, y ningún enlace entrante apunta a un ancla interna. Se
  conserva el H1 `# Cómo se elige el tablero`. `cmd/init.md` ya explica cómo nace el tablero: la
  sección nueva es un resumen corto que remite ahí.
- **2.4**: el instante de un comentario se llama `createdAt` y `--set-comment-date` existe con ese
  nombre. `leaseExpiresAt` se define en `docs/spec/lease.md`: la sección de `fechas.md` remite ahí.
- **Paralelo**: si un paso necesita tocar `mapa-de-secciones.txt`, no lo edita: deja la línea nueva en
  su informe y la aplica el coordinador.
