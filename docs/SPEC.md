# `biso`: especificación del CLI de gestión de tareas

Este documento define `biso`, una herramienta de línea de comandos para llevar las tareas de un
proyecto. Está escrito para que alguien implemente el programa entero a partir de él sin preguntar
nada: cada comando trae su firma, su tabla de parámetros, su comportamiento en los casos límite, la
salida literal que imprime, su esquema JSON, sus códigos de salida y el texto exacto de su ayuda.

El destinatario principal de `biso` es un agente automático que trabaja dentro del proyecto. La
salida es predecible, los errores son distinguibles por su código sin leer el mensaje, y ningún
comportamiento depende de dónde se ejecute el programa.

**Qué define este documento y qué no.** Define la interfaz del programa y el modelo de datos lógico
de una tarea. **No define cómo ni dónde se guardan los datos.** No se menciona ningún formato de
fichero, ninguna base de datos, ninguna ruta ni ningún nombre de fichero de tarea, porque esa
decisión está abierta y la especificación no debe atarla. Donde el almacenamiento importa para el
comportamiento observable, este documento enuncia el requisito (por ejemplo, que dos procesos
simultáneos no puedan asignar el mismo identificador) y deja el mecanismo a quien implemente.

El modelo de tarea es compatible con el de Backlog.md, de modo que se puede importar y exportar entre
las dos herramientas sin perder campos.

**Convención de idioma.** La prosa de este documento va en español. Todo lo que es interfaz del
programa (nombres de comando, banderas, textos de ayuda, mensajes de error, claves JSON y claves de
configuración) va en inglés, porque es lo que la persona o el agente que usa el programa lee y
escribe.

---

## 1. Los principios

Siete reglas. El resto del documento es una consecuencia de ellas.

1. **Un valor que el tablero no conoce es un error, se esté leyendo o escribiendo, y siempre con la
   misma regla de coincidencia.** Un filtro con un valor imposible nunca devuelve una lista vacía.
   Como consecuencia, una lista vacía es un hecho sobre el tablero y quien la recibe puede actuar en
   consecuencia.

2. **Un nombre significa siempre lo mismo, en todos los comandos.** No existen dos banderas con el
   mismo nombre y semántica distinta según dónde se usen, ni dos nombres para el mismo concepto.

3. **El nombre desnudo añade. Sustituir se dice en voz alta.** `--label` añade una etiqueta y
   `--set-label` reemplaza la lista entera.

4. **La salida por defecto de una escritura es lo que quien llama no sabía.** Nunca el eco de lo que
   acaba de escribir.

5. **Un gesto del flujo de trabajo es un comando.** Empezar una tarea y terminarla tienen nombre
   propio y cuestan una llamada cada uno.

6. **Todo lo que se hace una vez se puede hacer cien veces, y se valida antes de escribir nada.**

7. **La forma de la salida no depende de si hay un terminal detrás.** Solo el color mira el terminal.
   Los datos, nunca.

---

## 2. Códigos de salida

Tabla global. Ningún comando usa un código fuera de esta tabla, y ningún código tiene dos
significados. Quien llama puede ramificar sobre el número sin leer el mensaje.

| Código | Nombre | Significado | Ejemplo |
|---:|---|---|---|
| 0 | `OK` | La operación terminó y se aplicó | `biso new "Algo"` |
| 1 | `INTERNAL` | Fallo no previsto del programa | una excepción no capturada |
| 2 | `USAGE` | La línea de comandos está mal formada | bandera desconocida, falta un obligatorio, dos banderas incompatibles, identificador mal formado, bandera de escritura en un comando de lectura |
| 3 | `BAD_VALUE` | El valor que llega es sintácticamente correcto pero el tablero no lo reconoce, o un dato guardado no se puede interpretar | `--status "Pending"` en un tablero cuyos estados son otros |
| 4 | `NOT_FOUND` | La entidad referida no existe | `biso get TASK-999` |
| 5 | `AMBIGUOUS` | La referencia encaja con más de una entidad | `biso get "parser"` con tres coincidencias |
| 6 | `PRECONDITION` | La operación es válida, pero el estado actual del tablero no la permite o no la satisface | `biso finish --strict` con criterios sin marcar, o `biso doctor` con problemas pendientes |
| 7 | `ENVIRONMENT` | Falla el entorno, no la petición | el almacén no responde, no hay permisos, no se puede adquirir el acceso exclusivo, no hay terminal donde hace falta |
| 8 | `NO_BOARD` | No hay tablero accesible desde donde se ha llamado | cualquier comando fuera de un tablero, salvo `init`, `help`, `--help` y `--version`, que no necesitan uno; `biso where` también devuelve 8 cuando no encuentra ninguno |
| 9 | `VALIDATION` | Una validación previa ha fallado y **no se ha escrito nada** | `biso new --from tareas.ndjson` con la línea 47 inválida |

Cuatro reglas que acompañan a la tabla:

- **El código 9 garantiza que no se ha escrito nada.** Si un comando termina con 9, el tablero está
  exactamente como estaba antes. Por eso una validación fallida dentro de un lote se reporta como 9 y
  no como 3 ni como 4, y por eso el 9 llega siempre con el detalle de **todos** los fallos
  encontrados, no solo del primero.
- **Un listado vacío es siempre 0.** Un tablero donde de verdad no hay nada que cumpla un filtro
  válido no es un error.
- **El código 3 cubre dos direcciones.** Un valor de entrada que el tablero no reconoce, y un dato ya
  guardado que el programa no sabe interpretar. Las dos son "el vocabulario no cuadra", y el mensaje
  siempre dice cuál de las dos ha ocurrido.
- **El código 1 es un fallo del programa, no de quien llama.** La reacción correcta es informar, no
  reintentar con otros parámetros.

---

## 3. Banderas globales

Valen para todos los comandos, se pueden escribir antes o después del nombre del comando, y ningún
comando puede redefinir ninguna de ellas ni cambiar su significado.

| Bandera | Corta | Tipo | Por defecto | Qué hace |
|---|---|---|---|---|
| `--cwd <path>` | `-C` | ruta | el directorio actual | Resuelve el tablero desde ahí, sin cambiar el directorio del proceso |
| `--board <name>` | | localizador | el que se resuelva | Usa ese tablero directamente, sin buscar |
| `--json` | | booleano | falso | Toda la salida de datos es JSON, en el sobre de la sección 12 |
| `--quiet` | `-q` | booleano | falso | Reduce la salida a lo mínimo. Ver más abajo |
| `--print` | | booleano | falso | Después de escribir, imprime la ficha completa de cada tarea afectada |
| `--color <when>` | | `auto`, `always`, `never` | `auto` | Control de los códigos de color |
| `--dry-run` | | booleano | falso | Valida todo, no escribe nada. Sale 0 si habría funcionado y 9 si no |
| `--version` | `-V` | booleano | | Imprime `biso 1.0.0` y sale con 0 |
| `--help` | `-h` | booleano | | Imprime la ayuda del comando y sale con 0 |

Reglas de aplicación, que hay que implementar tal cual:

- **`--print` y `--dry-run` solo tienen sentido en los comandos que escriben.** En `prime`, `where`,
  `ls`, `get`, `export`, `config get`, `config list`, `board`, `help` y `biso doctor` sin `--fix` son
  un error de uso con código 2 y el mensaje `error: <flag> does not apply to a read-only command`, con
  `<flag>` igual a `--print` o `--dry-run` según cuál se haya usado. No se ignoran en silencio.
  `biso doctor --fix` es la excepción: con `--fix` es un comando de escritura, y sus dos banderas se
  comportan como en cualquier otro (10.11).
- **`--json` es incompatible con `--quiet`** y con `--print`, porque los tres piden formas distintas
  de la misma salida. Cualquier pareja de las tres da código 2.
- **`--quiet` reduce stdout a los identificadores afectados**, uno por línea, y además suprime las
  líneas informativas de stderr que empiezan por `note:`. **Nunca suprime un `warning:` ni un
  `error:`.** Silenciar un aviso es cosa de quien llama, con `2>/dev/null`. **En un comando de
  lectura no hay identificadores afectados que imprimir**, así que ahí `--quiet` no cambia stdout: solo
  suprime las líneas `note:` de stderr, igual que en un comando de escritura.
- **`--board` acepta el nombre o el localizador de un tablero**, en la forma que el almacenamiento
  imponga. El programa lo trata como una cadena opaca que identifica un tablero, y `biso where` la
  imprime.

### 3.1. Variables de entorno

| Variable | Equivale a | Precedencia |
|---|---|---|
| `BISO_CWD` | `--cwd` | la bandera gana |
| `BISO_BOARD` | `--board` | la bandera gana |
| `BISO_ME` | la identidad de quien llama, para `--mine` y para el autor por defecto de los comentarios | la clave `me` de la configuración gana; si no está, esta variable |
| `BISO_LIMIT` | el límite por defecto de `biso ls` | `--limit` gana, luego esta variable, luego la clave `default_limit`, luego 30 |
| `NO_COLOR` | `--color never`, si está definida con cualquier valor | `--color` gana |

**Qué pasa si no hay identidad**, es decir, ni la clave `me` ni `BISO_ME` están definidas:

| Dónde se usaría `me` | Qué pasa sin ella |
|---|---|
| `biso ls --mine` | Error 6: `error: --mine needs an identity; set it with biso config set me <you> or BISO_ME` |
| `biso start`, autoasignación | No asigna a nadie. Sale `note: no identity configured, task left unassigned` en vez del `note:` de siempre |
| Autor por defecto de un comentario | Error 2 si no se ha pasado `--comment-author`: `error: --comment-author is required, no identity is configured` |
| La línea `you are` de `biso prime` | `you are     (not set)`, con una nota que remite a `biso config set me` |

### 3.2. Cómo se elige el tablero

Un proyecto tiene un tablero, y el programa lo encuentra por este orden. Gana el primero que exista:

1. La bandera `--board`.
2. La variable `BISO_BOARD`.
3. **El puntero del proyecto**, que es una marca que `biso init` deja en el proyecto y que dice qué
   tablero le corresponde. Se busca en el directorio de trabajo y en sus ancestros, con el tope de la
   regla que cierra esta lista.
4. **Un tablero que el propio almacenamiento asocia al directorio de trabajo**, buscado con el mismo
   procedimiento hacia arriba y el mismo tope.

**El tope de la búsqueda hacia arriba** es la raíz del proyecto, entendida como la raíz del
repositorio de control de versiones si lo hay, y si no lo hay, el propio directorio de partida. La
búsqueda **nunca** sube por encima de la raíz del proyecto ni llega al directorio personal, para que
un proyecto no encuentre por accidente el tablero de un proyecto hermano.

Si nada de eso existe, cualquier comando salvo `init`, `where`, `help`, `--help` y `--version` aborta
antes de ejecutar su propia lógica, con código 8 y este mensaje por stderr:

```
error: no board here, and none configured for this project
hint: `biso init` creates one, `biso where` explains what was searched
```

Los cinco exentos no abortan así: `init`, `help`, `--help` y `--version` no necesitan tablero para
hacer su trabajo, y `biso where` lo necesita pero lo comprueba por su cuenta, con su propio mensaje y
su propio código 8 cuando no lo encuentra (sección 10.2).

El puntero del paso 3 es lo que permite que varias copias de trabajo del mismo proyecto compartan un
solo tablero en lugar de tener uno cada una. **`biso init` lo crea siempre que el tablero no quede
dentro del propio proyecto** (sección 10.1), y `biso where` dice cuál se ha usado y por qué
(sección 10.2). No hay ningún caso en el que haya que escribirlo a mano.

---

## 4. Reglas transversales

Estas reglas valen para todos los comandos y no se repiten en cada uno. Un comando solo las menciona
cuando se aparta de ellas, y ninguno lo hace salvo donde se diga.

### 4.1. Interactividad, terminal y color

**Ningún comando abre nunca una interfaz interactiva por su cuenta, y ningún comando pregunta nada.**
No existe la detección de terminal como forma de decidir qué imprime un comando: la salida de
cualquier comando es idéntica byte a byte con terminal y sin él, salvo los códigos de color.

La interfaz interactiva existe, pero es un comando aparte, `biso board`, que solo se ejecuta si se
pide por su nombre y que falla con código 7 si no hay terminal.

Lo único que mira el terminal es el color:

| Situación | Color |
|---|---|
| `--color always` | sí |
| `--color never`, o `NO_COLOR` definida | no |
| `--color auto` y stdout es un terminal | sí |
| `--color auto` y stdout está redirigido | no |

El color se decide por separado para stdout y para stderr, cada uno según su propio destino. Los
códigos de color nunca cambian el texto: quitarlos deja exactamente las líneas documentadas aquí.

### 4.2. stdout, stderr y qué va en cada uno

La regla es fija y no tiene excepciones:

- **stdout lleva datos.** Lo que un programa consumiría: identificadores, listados, fichas, JSON.
- **stderr lleva todo lo demás.** Errores, avisos, notas informativas, sugerencias y el resumen de lo
  que se ha omitido.

En particular, la línea de `biso ls` que dice cuántas tareas se han ocultado va por stderr, porque no
forma parte del listado. Redirigir stdout a un fichero produce un fichero de datos limpio, y
redirigirlo a `/dev/null` no pierde ni un solo aviso.

### 4.3. Notas y avisos

Hay dos clases de mensaje que no son errores, las dos por stderr, y las dos dejan el código de salida
en 0:

- **`note:`** es información de contexto. `--quiet` la suprime.
- **`warning:`** es algo que quien llama necesita saber y que no impide la operación. **Nunca se
  suprime.**

Esta es la lista completa de avisos que el programa emite. No hay ningún otro:

| Aviso | Cuándo |
|---|---|
| `warning: --set-plan replaced 412 bytes of existing content` | cualquier `--set-*` que pise contenido no vacío |
| `warning: TASK-11 moved to Done with 1 of 2 acceptance criteria unchecked` | al llegar a un estado terminal con criterios sin marcar |
| `warning: TASK-11 finished without a final summary` | al llegar a un estado terminal sin resumen |
| `warning: TASK-11 moved to Done with 1 of 3 definition-of-done items unchecked` | al llegar a un estado terminal con la definición de hecho a medias |
| `warning: TASK-11 has unfinished subtasks: TASK-14, TASK-15` | al terminar una tarea con subtareas vivas |
| `warning: TASK-11 is a dependency of TASK-20, which is not finished` | al archivar una tarea de la que dependen otras vivas |
| `warning: --clear-label has no effect on a new task` | cualquier `--clear-*` en `biso new` |
| `warning: TASK-11 has unresolved dependencies: TASK-4 (To Do)` | al empezar una tarea bloqueada |
| `warning: 212 more tasks match; showing 30 of 242` | en `biso ls`, al recortar |
| `warning: --label: "urgent" given twice, kept once` | valor repetido en una bandera de lista |
| `warning: --desc contains a literal \n and no real newline; it will be stored as text` | ver 4.4 |
| `warning: --note: empty value, nothing was added` | valor vacío en una bandera que añade |
| `warning: --due 2026-01-01 is in the past` | fecha límite ya pasada |
| `warning: TASK-11 has no acceptance criteria` | `--check all` sobre una tarea sin criterios |
| `warning: 1 task could not be read and was skipped` | ver 4.11 |
| `warning: <x> is deprecated and will be removed in 2.0` | ver la sección 13 |

### 4.4. Codificación y texto

- La entrada y la salida son **UTF-8**, siempre, sea cual sea la configuración regional del sistema.
  Una secuencia de bytes inválida en un argumento o en un fichero de entrada es un error con código 3
  que señala la posición del byte.
- Los saltos de línea de salida son `\n`. Al leer una entrada, `\r\n` y `\n` se aceptan por igual y
  se normalizan a `\n`.
- El texto se guarda tal cual llega. **Ninguna secuencia de escape se interpreta.** Un `\n` literal
  de dos caracteres se guarda como dos caracteres.
- Como ese `\n` literal casi siempre es un accidente, un valor de texto que contenga la secuencia de
  dos caracteres `\` `n` y **ningún** salto de línea real produce este aviso, y se guarda igual:
  ```
  warning: --desc contains a literal \n and no real newline; it will be stored as text
  hint: use a real newline, or -d @file.md, or -d - to read from stdin
  ```

### 4.5. Tres formas de pasar un valor largo

Todo parámetro de tipo texto largo (`--desc`, `--plan`, `--note`, `--summary`, `--comment` y el texto
de un criterio o de un elemento de la definición de hecho) acepta las tres:

| Forma | Significado |
|---|---|
| `--desc "texto"` | el texto literal |
| `--desc @ruta/fichero.md` | el contenido del fichero, interpretado como UTF-8 |
| `--desc -` | todo lo que llegue por la entrada estándar hasta el fin de fichero |

Reglas:

- **Un texto que empieza de verdad por `@` se escribe `@@`.** El primer `@` se descarta y el resto es
  literal. Es la única secuencia de escape del programa.
- **Los campos de persona nunca interpretan el `@`.** `--assignee`, `--reporter` y `--comment-author`
  toman su valor tal cual, así que `--comment-author @trello:juan` guarda ese texto y no intenta leer
  ningún fichero.
- **`-` solo puede aparecer una vez por invocación.** Dos parámetros que pidan la entrada estándar son
  un error de uso con código 2, porque el segundo leería un flujo agotado y guardaría el vacío sin
  que se note.
- **Un fichero que no existe es código 4**, con el mensaje `error: --desc: file not found: docs/x.md`.
  Un fichero que existe pero no se puede leer es código 7.
- **Un valor vacío, venga de donde venga, no borra nada.** Ver 4.6.

### 4.6. El valor vacío

Un valor vacío es una cadena sin ningún carácter, o solo con espacios, tanto si llega literalmente
como si llega de un fichero vacío o de una entrada estándar vacía. La regla es única:

| Dónde | Qué pasa |
|---|---|
| En una bandera que añade (`--note`, `--label`, `--ac`, `--desc`) | No se añade nada, se emite `warning: --note: empty value, nothing was added` y el código sigue siendo 0 |
| En una bandera que sustituye (`--set-notes`, `--set-label`) | Deja el campo vacío, igual que `--clear-notes`. Sustituir por nada es vaciar, y eso sí es explícito |
| En un campo escalar (`--type ""`, `--priority ""`) | Error 3. **La cadena vacía nunca es la forma de borrar un escalar**; para eso está `--clear-type` |
| En el título, al crear | Error 2: `error: title cannot be empty` |

**El `code` de un escalar vacío depende de si el campo tiene vocabulario cerrado.** Para `status`,
`type`, `priority` y `project`, una cadena vacía es un valor que no coincide con nada configurado, así
que sigue la regla de 6.1 y el `code` es el de un valor desconocido (`unknown_status` y análogos, con
el mensaje de 6.2). Para los demás escalares (`--reporter ""`, `--ordinal ""`, `--due ""`), que no
tienen vocabulario, el `code` es `empty_scalar_value`.

### 4.7. Valores que empiezan por guion

Tres mecanismos, en orden de preferencia:

1. **`--flag=valor`** funciona siempre y es la forma recomendada: `--desc=-5 grados`.
2. **`--`** termina el análisis de opciones: `biso new -- "-n no es una bandera"`.
3. **Un valor que empieza por guion detrás de una bandera que exige valor se acepta tal cual**, sin
   heurísticas. `biso set TASK-1 --note -x` guarda `-x` como nota.

Como consecuencia de la regla 3, olvidar el valor de una bandera se detecta por lo que sobra después,
no por lo que parece: `biso set TASK-1 --note --priority high` guarda la nota `--priority` y luego
falla con código 2 y `error: unexpected argument: high`.

### 4.8. Repetición y listas separadas por comas

Para toda bandera marcada como repetible:

- Repetirla acumula: `--label a --label b` deja dos etiquetas.
- Si además acepta lista, separar por comas acumula igual: `--label a,b` deja las mismas dos.
- Las dos formas se pueden mezclar.
- **Una coma dentro de un valor se escapa con `\,`.** Es la única forma de meter una coma en una
  etiqueta o en una referencia.
- Los campos de texto largo y los criterios **nunca** se parten por comas.
- Un valor repetido dentro de la misma bandera se guarda una vez y produce
  `warning: --label: "urgent" given twice, kept once`.

Para toda bandera **no** repetible, es decir, los campos escalares, pasarla dos veces con valores
distintos es un error de uso con código 2:

```
error: --status given twice with different values: "In Progress" and "Done"
```

### 4.9. Orden de aplicación dentro de una escritura

Una sola invocación puede tocar muchos campos. El orden en que se aplican es fijo y **no depende del
orden en que aparecen las banderas en la línea de comandos**, para que el resultado sea reproducible:

1. Todos los `--clear-*`.
2. Todos los `--set-*`.
3. Todos los `--rm-*`.
4. Los añadidos, es decir, los nombres desnudos.
5. Los campos escalares.
6. Los marcados de criterios y de definición de hecho.
7. Los comentarios.

Con este orden, `--clear-label --label urgent` deja exactamente una etiqueta, y `--set-ac "A"
--check all` marca los criterios recién puestos. Dentro de un mismo paso manda el orden de la línea
de comandos: `--label b --label a` deja `b` antes que `a`. Las listas nunca se ordenan solas.

### 4.10. Concurrencia, atomicidad y garantías observables

Esta sección no describe un mecanismo: enuncia lo que quien llama tiene derecho a observar. Cómo se
consiga es cosa de quien implemente.

1. **Ninguna escritura se observa a medias.** Un lector concurrente ve el tablero como estaba antes
   de una escritura o como quedó después, nunca en un punto intermedio, y esto vale igual para una
   escritura de una tarea que para un lote de doscientas.
2. **Una escritura que afecta a varias tareas es todo o nada.** Si falla por cualquier motivo, ni una
   sola de las tareas implicadas queda modificada, y el código de salida lo dice: 9 si el fallo se
   detectó al validar, 7 si se detectó al escribir. En los dos casos el mensaje afirma explícitamente
   que no se ha escrito nada.
3. **Dos procesos simultáneos nunca asignan el mismo identificador**, aunque trabajen sobre el mismo
   tablero desde copias de trabajo distintas del proyecto.
4. **Dos escrituras simultáneas sobre la misma tarea no se pierden ni se mezclan.** O se aplican una
   después de otra, o una de las dos falla con código 7.
5. **Si el programa no puede obtener el acceso exclusivo que necesita para escribir**, espera hasta
   cinco segundos y luego falla con código 7 sin escribir nada:
   ```
   error: the board is busy, another process is writing to it
   hint: retry in a moment; nothing was written
   ```
6. **Las lecturas nunca fallan por culpa de una escritura en curso**, y nunca la bloquean.

### 4.11. Identificadores

- Un identificador es `<PREFIX>-<n>`, con `n` entero positivo. `PREFIX` viene de la configuración
  (`task_prefix`, por defecto `TASK`).
- **Un identificador no se reutiliza jamás**, ni después de archivar una tarea ni después de
  eliminarla por cualquier vía.
- Los identificadores se asignan de forma creciente, pero **la especificación no promete que la
  secuencia no tenga huecos**. Un hueco es normal y nunca es un error.
- El tablero sabe en todo momento cuál es el identificador más alto que ha llegado a asignar, y ese
  dato se usa en los mensajes de la sección 7.3 y en `biso doctor`.

### 4.12. Qué pasa con un dato que no se puede interpretar

Una tarea puede resultar ilegible: el almacén devuelve algo corrupto, o la tarea lleva una clave de
extensión que la configuración ya no declara. La regla es única y depende del tipo de lectura:

| Tipo de lectura | Qué pasa |
|---|---|
| **Lectura dirigida** a esa tarea, es decir, `get`, o `set`, `start`, `note`, `comment`, `finish` y `archive` con una referencia que resuelve a ella | Error 3, con el motivo exacto. No se escribe nada |
| **Lectura de conjunto**, es decir, `ls`, `prime`, `export`, la resolución de una referencia por texto y cualquier filtro | La tarea se salta, se cuenta, y al final se emite `warning: 1 task could not be read and was skipped` con sus identificadores. El resto del resultado es válido y el código es 0, **salvo en `biso export`, que sale con 6** |
| `biso doctor` | Se reporta como problema y se sigue con las demás. Nunca aborta |

Una lectura de conjunto **nunca** aborta por una tarea mala, y **nunca** la esconde en silencio. Las
dos cosas juntas son lo que impide que un listado incompleto se confunda con un tablero vacío.

**`biso export` es la única excepción al código 0 de una lectura de conjunto.** Escribe igual todo lo
que ha podido leer, con el mismo aviso por stderr, pero termina con **código 6** en vez de 0 cuando ha
saltado alguna tarea. Un guion que encadene `biso export -o backup.ndjson && ...` puede comprobar el
código de salida para detectar un volcado incompleto.

---

## 5. El modelo de datos de una tarea

Este es el modelo **lógico**. Describe qué campos tiene una tarea, de qué tipo son y quién los
escribe. No dice nada de cómo se guardan.

| Campo | Tipo lógico | Obligatorio | Quién lo fija | Mutable |
|---|---|---|---|---|
| `id` | identificador `PREFIX-<n>` | sí | el programa | no |
| `title` | texto de una línea | sí | quien llama | sí |
| `status` | uno del vocabulario de estados | sí | quien llama | sí |
| `type` | uno del vocabulario de tipos | no | quien llama | sí |
| `priority` | uno del vocabulario de prioridades | no | quien llama | sí |
| `project` | uno del vocabulario de proyectos | no | quien llama | sí |
| `milestone` | texto libre | no | quien llama | sí |
| `parent` | referencia a otra tarea | no | quien llama | sí |
| `assignees` | lista de textos de persona | no | quien llama | sí |
| `reporter` | texto de persona | no | el programa al crear, o quien llama; ver 5.6 | sí |
| `labels` | lista de textos | no | quien llama | sí |
| `dependencies` | lista de referencias a tareas | no | quien llama | sí |
| `references` | lista de textos | no | quien llama | sí |
| `documentation` | lista de textos | no | quien llama | sí |
| `modifiedFiles` | lista de textos | no | quien llama | sí |
| `due` | fecha `YYYY-MM-DD` | no | quien llama | sí |
| `ordinal` | entero >= 0 | no | quien llama | sí |
| `createdAt` | instante UTC | sí | el programa | solo al importar |
| `updatedAt` | instante UTC | sí | el programa | solo al importar |
| `archived` | booleano | sí, `false` por defecto | el programa, con `biso archive` | sí, solo con `biso archive` / `--unarchive`, o al importar |
| `urgency` | decimal, derivado | derivado | el programa | no, se recalcula al leer |
| `ext` | mapa de clave declarada a texto | no | quien llama | sí |
| `description` | texto largo | no | quien llama | sí |
| `plan` | texto largo | no | quien llama | sí |
| `notes` | texto largo | no | quien llama | sí |
| `summary` | texto largo | no | quien llama | sí |
| `acceptanceCriteria` | lista de criterios | no | quien llama | sí |
| `definitionOfDone` | lista de criterios | no | quien llama | sí |
| `comments` | lista de comentarios | no | quien llama | solo se añade |
| `acDone`, `acTotal`, `dodDone`, `dodTotal` | entero, derivado | derivado | el programa | no, se recalculan al leer |
| `commentCount` | entero, derivado | derivado | el programa | no, se recalcula al leer |
| `blocks` | lista de referencias, derivado | derivado | el programa | no, se recalcula al leer |
| `ready`, `blocked` | booleano, derivado | derivado | el programa | no, se recalculan al leer |

Cuatro precisiones sobre la mutabilidad:

- **"No mutable" significa que ninguna bandera del programa lo cambia.** `updatedAt` lo reescribe el
  programa en cada operación que cambie algo.
- **Un comentario no se edita ni se borra, solo se añade.** Un comentario es el registro de una
  conversación.
- **`archived` solo lo cambia `biso archive` y `biso archive --unarchive`.** No hay una bandera de
  campo de la sección 8 para él: archivar es un gesto de flujo de trabajo con nombre propio,
  según el principio 5.
- **Los campos marcados "derivado" en esta tabla no se guardan.** Se calculan al leer, y son
  exactamente los campos que `biso export` no escribe (10.9) y que `biso new --from` rechaza como
  clave desconocida (10.3): `urgency`, `acDone`, `acTotal`, `dodDone`, `dodTotal`, `commentCount`,
  `blocks`, `ready` y `blocked`. Esta es la única lista de campos derivados del documento; las
  demás secciones remiten a ella.

### 5.1. Los criterios y sus claves estables

`acceptanceCriteria` y `definitionOfDone` son listas del mismo tipo, y cada elemento tiene tres cosas:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | entero positivo | el programa al crear el elemento |
| `text` | texto | quien llama |
| `checked` | booleano | quien llama |

**La clave se asigna al crear el elemento, con un contador propio de esa lista dentro de esa tarea, y
no se reasigna nunca.** Quitar un elemento no mueve las claves de los demás: una tarea puede tener
perfectamente los criterios `#1` y `#3` y ninguno más. Cada tarea lleva dos contadores, uno por
lista, que solo crecen.

Consecuencias que hay que respetar en toda la implementación:

- Los selectores de la sección 8.4 trabajan sobre la clave, **nunca** sobre la posición.
- `acTotal` y `dodTotal`, allá donde aparezcan, son **el número de elementos presentes**, nunca la
  clave más alta. Una tarea con los criterios `#1` y `#3` tiene `acTotal` igual a 2.
- Los elementos se muestran y se exportan en el orden en que están en la lista, que es el orden en
  que se crearon salvo que se haya sustituido la lista entera.

### 5.2. Los comentarios

Cada comentario tiene autor, instante y cuerpo:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | texto libre | quien llama, y por defecto la identidad `me` |
| `createdAt` | instante UTC | el programa, salvo al importar |
| `body` | texto largo | quien llama |

**El autor es texto libre y no se valida contra nada.** Un comentario puede venir de alguien que no
existe en este tablero, y un sistema externo puede usar su propia convención, por ejemplo
`@trello:juan`.

### 5.3. Las fechas

`createdAt`, `updatedAt` y el instante de cada comentario los pone el programa con el reloj del
sistema, en UTC y con precisión de segundo.

**Se pueden fijar solo al importar**, es decir, en `biso new --from`. En cualquier otro sitio son un
hecho observado y no un dato que se negocie.

### 5.4. La urgencia

`urgency` es un decimal derivado que se recalcula en cada lectura y **nunca se guarda**. Es el segundo
criterio de la tupla de orden por defecto de `biso ls`, después de `ordinal` (10.4), y el que ordena
el resumen de `biso prime`.

```
Si el estado de la tarea es el terminal, urgency = 0.0 y no se calcula nada mas.

En cualquier otro caso:

urgency = 6.0  * prioridad         (high 1.0, medium 0.5, low 0.0, sin prioridad 0.3)
        + 4.0  * activa            (1.0 si el estado es el activo, 0.0 si no)
        + 8.0  * bloquea           (1.0 si alguna tarea sin terminar depende de esta)
        - 5.0  * bloqueada         (1.0 si depende de alguna tarea sin terminar)
        + 12.0 * proximidad        (ver la regla siguiente)
        + 1.0  * tiene_criterios   (1.0 si tiene al menos un criterio de aceptacion)
        + 0.5  * min(edad_dias / 30, 4.0)

El resultado se redondea a un decimal.
```

**La regla de `proximidad`, sin ambigüedad:**

```
dias = fecha_limite - hoy, en dias (puede ser negativo si la fecha ya paso)

si la tarea no tiene fecha limite:  proximidad = 0.0
si la tiene:                        proximidad = clamp((30 - dias) / 30, 0.0, 1.0)
```

Una tarea vencida tiene `dias` negativo, así que `(30 - dias) / 30` supera 1 y el resultado se acota
en **1.0**, el mismo máximo que una tarea que vence hoy. Una tarea vencida no suma más que una que
vence hoy; para distinguirlas está el filtro `--overdue` de `biso ls`, no un término sin tope en la
fórmula.

Un ejemplo completo, que es el que imprime `biso get --explain-urgency` en la sección 10.5: una tarea
de prioridad alta, en el estado activo, de la que depende otra tarea sin terminar, sin fecha límite,
con dos criterios y creada hoy, suma `6.0 + 4.0 + 8.0 + 0.0 + 0.0 + 1.0 + 0.0`, es decir **19.0**.

**Los coeficientes configurables son exactamente siete, bajo `urgency.`, uno por término de la
fórmula**: `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`,
`urgency.criteria` y `urgency.age`, con los valores de arriba (6.0, 4.0, 8.0, -5.0, 12.0, 1.0 y 0.5)
como valores por defecto. **Los pesos por prioridad no son configurables**: `high 1.0, medium 0.5,
low 0.0, sin prioridad 0.3` son parte de la estructura fija de la fórmula, que no cambia en la
versión 1.0.

**El `ordinal` no forma parte de la urgencia.** Es un orden manual que se aplica aparte, según la
regla de orden completa de la sección 10.4.

### 5.5. Los campos externos

`ext` es un mapa de clave a texto para guardar la identidad de una tarea en otro sistema. La regla
es la siguiente:

- El tablero **declara** en su configuración qué claves admite, en la lista `extensions`.
- Escribir una clave declarada funciona: `biso set TASK-1 --ext trello.card=5f2a8c1e3b9d4a7f`.
- Escribir una clave no declarada es error 3:
  ```
  error: unknown extension key: "jira.key"
         declared keys on this board: trello.card, github.issue
  ```
- Una tarea que ya guarda una clave que la configuración no declara **no se lee en silencio ni se
  reescribe perdiéndola**: se aplica la regla de 4.12, y `biso doctor` la reporta.

### 5.6. Quién reporta una tarea

`reporter` se fija una sola vez, al crear la tarea, y después solo cambia si alguien pasa
`--reporter` de forma explícita.

| Al crear la tarea | Valor de `reporter` |
|---|---|
| Se pasa `--reporter <persona>` | esa persona, tal cual |
| No se pasa, y hay identidad configurada | la identidad de quien llama, según la precedencia de 3.1 |
| No se pasa, y no hay identidad configurada | vacío, sin aviso |
| Se pasa `--reporter ""` | vacío |

El caso sin identidad no es un error y no imprime nada: a diferencia de `--mine`, de la
autoasignación de `biso start` y del autor de un comentario, que sí la necesitan y están cubiertos
por la tabla de 3.1, una tarea sin quien la reporte es válida.

En el lote de `biso new --from`, un objeto que trae `reporter` conserva ese valor, y uno que no lo
trae aplica las mismas reglas de esta tabla.

---

## 6. Los vocabularios del tablero y la regla de validación

Tres campos tienen vocabulario cerrado, definido en la configuración: `status`, `type` y `priority`.
Un cuarto, `project`, lo tiene solo si el tablero declara proyectos. Para todos ellos rige una sola
regla, **idéntica al escribir y al leer**.

### 6.1. El algoritmo de coincidencia

Dado un valor de entrada `v` y la lista de valores configurados, el programa calcula así:

```
normalizar(x):
  1. pasar x a minusculas segun Unicode
  2. descomponer y quitar los diacriticos (acentos, dieresis, cedillas)
  3. eliminar TODOS los caracteres que sean espacio, tabulador, guion (-) o guion bajo (_)
  4. devolver lo que queda

coincidir(v, configurados):
  a. si existe un configurado c con c == v exactamente, devolver c
  b. si no, calcular normalizar(v) y compararlo con normalizar(c) de cada configurado
  c. si exactamente un configurado coincide, devolverlo
  d. si ninguno coincide, error 3
  e. si coinciden dos o mas, error 3 con los dos listados, porque el tablero
     tiene dos valores que se normalizan igual y hay que desambiguarlos
```

Con este algoritmo, y para un tablero cuyo estado es `To Do`:

| Entrada | `normalizar` | Resultado |
|---|---|---|
| `To Do` | `todo` | coincide, por el paso a |
| `todo` | `todo` | coincide |
| `TODO` | `todo` | coincide |
| `To-Do` | `todo` | coincide |
| `TO_DO` | `todo` | coincide |
| `to  do` | `todo` | coincide |
| `Pending` | `pending` | error 3 |
| `Todos` | `todos` | error 3 |

**No hay coincidencia por prefijo ni por parecido.**

### 6.2. El mismo texto vale lo mismo en los dos sentidos

Esta tabla es el contrato, y es la prueba de aceptación que hay que poder ejecutar. Tablero con los
estados `Ideas`, `To Do`, `In Progress`, `Blocked` y `Done`:

| Entrada | `biso set TASK-1 -s <v>` | `biso ls -s <v>` |
|---|---|---|
| `To Do` | escribe | filtra |
| `todo` | escribe | filtra |
| `TO_DO` | escribe | filtra |
| `In-Progress` | escribe | filtra |
| `Pending` | error 3 | error 3 |
| `""` | error 3 | error 3 |

El mensaje es el mismo en los dos sentidos:

```
error: unknown status: "Pending"
       valid statuses on this board: Ideas, To Do, In Progress, Blocked, Done
```

### 6.3. Qué valida cada filtro, y contra qué

| Filtro | Conjunto contra el que valida | Si no encaja |
|---|---|---|
| `--status`, `--type`, `--priority`, `--project` | el vocabulario configurado | error 3 |
| `--label` y `--label-or` | el conjunto de etiquetas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--assignee` | el conjunto de personas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--parent` | la resolución de referencias de la sección 7 | error 2, 4 o 5 |
| `--milestone` | nada, es texto libre; se compara con la regla de 6.1 | nunca falla, puede no devolver nada |
| `--search` | nada, es texto libre | nunca falla |

**El conjunto de etiquetas del tablero** es la unión de las etiquetas declaradas en la clave `labels`
de la configuración y de las que lleva cualquier tarea del tablero, **incluidas las archivadas y las
que están en el estado terminal**. **El conjunto de personas se define con la clave `assignees` de la
configuración y con los valores de `assignees` de cualquier tarea, archivadas y terminadas incluidas.
Los valores de `reporter` no entran en este conjunto**, porque no hay ningún filtro `--reporter`: una
persona que solo ha reportado tareas y nunca las ha tenido asignadas no pertenece al conjunto contra
el que valida `--assignee`.

**Las etiquetas y las personas no tienen vocabulario cerrado al escribir.** Escribir una etiqueta nueva
la incorpora al conjunto, y a partir de ese momento filtrar por ella funciona.

Está la bandera `--unchecked` de `biso ls` y `biso export`, que apaga la comprobación de etiquetas y
personas y solo esa; no cambia ninguna otra cosa.

---

## 7. Cómo se resuelve una referencia a una tarea

Todos los comandos que reciben `<ref>` usan exactamente esta rutina. No hay variantes por comando.

### 7.1. La gramática

| Forma | Ejemplo | Interpretación |
|---|---|---|
| `PREFIX-<n>` | `TASK-11` | identificador, sin distinguir mayúsculas en el prefijo |
| `<n>` | `11` | identificador, con el prefijo del tablero |
| `#<n>` | `#11` | igual que el anterior |
| cualquier otra cosa | `"CRLF"` | consulta de texto |

Dos banderas fuerzan la interpretación, y valen en todos los comandos que aceptan una referencia:

- `--id` obliga a interpretar como identificador. Con un valor que no encaje en la gramática, error 2.
- `--match` obliga a interpretar como texto, y sirve para buscar una tarea que se llame "42".

### 7.2. La búsqueda por texto

**Hay un solo ámbito de búsqueda de texto en todo el programa**, y es el que usan tanto la resolución
de una referencia como el filtro `--search` de `biso ls` y `biso export`. Busca, sin distinguir
mayúsculas ni acentos, en:

el título, la descripción, el plan, las notas, el resumen final, el texto de los criterios de
aceptación, el texto de la definición de hecho, el cuerpo de los comentarios y las etiquetas.

No busca en los identificadores, ni en las referencias, ni en la documentación, ni en los campos de
extensión.

Cuando se usa para resolver una referencia, y solo entonces, se aplican además estas reglas:

| Coincidencias | Qué pasa |
|---:|---|
| exactamente 1 | se usa esa tarea, con `note: "CRLF" matched TASK-11` por stderr |
| 0 | error 4 |
| más de 1 | error 5, con las candidatas por stdout en el formato de `biso ls` |

**Una coincidencia en el título gana sobre una coincidencia en cualquier otro sitio.** Si el texto
aparece en el título de una sola tarea, esa es la respuesta aunque aparezca en el cuerpo de otras
diez, y no hay ambigüedad. La búsqueda para resolver una referencia mira solo las tareas **no
archivadas**; el filtro `--search` mira las que digan los demás filtros.

### 7.3. Los tres mensajes de "no la encuentro"

**Identificador mal formado**, código 2, `code` igual a `malformed_id`:

```
error: malformed task id: "TASK-1.1"
hint: ids look like TASK-11 or 11. A subtask is an ordinary task with --parent TASK-1
```

**Identificador bien formado que el tablero nunca ha llegado a asignar**, código 4, `code` igual a
`never_allocated`:

```
error: TASK-999 has never existed on this board
note: the highest id ever assigned here is TASK-90
```

**Identificador que el tablero asignó alguna vez y que ahora no está**, código 4, `code` igual a
`not_found`:

```
error: TASK-53 is not on this board
note: TASK-53 was assigned at some point, so it was archived and then removed, or it belongs
      to a version of the project that this board does not have
hint: `biso ls --archived` lists what is archived
```

Los tres códigos son distintos: 2, y luego 4 con dos `code` distintos.

---

## 8. Las familias de banderas

Esta sección define de una vez la forma de todas las banderas de escritura. Los comandos no la
repiten: cada uno dice qué campos acepta, y esta sección dice qué forma tiene cada campo.

### 8.1. La regla

**El nombre desnudo añade. `set-` delante sustituye. `rm-` delante quita uno. `clear-` delante
vacía.** La forma de una bandera se deduce siempre del nombre del campo, sin nombres propios y sin
que haya que consultar nada.

Las cuatro variantes existen para **todo campo que guarde una lista de elementos**. Cada clase de
campo tiene estas variantes:

| Clase de campo | Variantes |
|---|---|
| Lista de elementos | las cuatro |
| Bloque de prosa | añadir, sustituir, vaciar |
| Mapa de claves | fijar una clave, quitar una clave, vaciar |
| Escalar | fijar, vaciar |
| Lista inmutable (comentarios) | solo añadir, `--comment` |

**El significado no cambia entre comandos.** `--ac` añade un criterio en `biso new`, en `biso set`, en
`biso start` y en `biso finish`, y todos los comandos de escritura aceptan todas estas banderas.

La regla tiene **una sola desviación de nombre en todo el programa**: el añadido de notas se llama
`--note`, en singular, y su sustitución se llama `--set-notes`, en plural.

### 8.2. Campos de lista

| Campo | Añade | Sustituye | Quita | Vacía | Acepta lista por comas |
|---|---|---|---|---|---|
| etiquetas | `-l, --label` | `--set-label` | `--rm-label` | `--clear-label` | sí |
| personas asignadas | `-a, --assignee` | `--set-assignee` | `--rm-assignee` | `--clear-assignee` | sí |
| referencias | `--ref` | `--set-ref` | `--rm-ref` | `--clear-ref` | sí |
| documentación | `--doc` | `--set-doc` | `--rm-doc` | `--clear-doc` | sí |
| dependencias | `--dep` | `--set-dep` | `--rm-dep` | `--clear-dep` | sí |
| ficheros tocados | `--file` | `--set-file` | `--rm-file` | `--clear-file` | sí |
| criterios de aceptación | `--ac` | `--set-ac` | `--rm-ac` | `--clear-ac` | **no** |
| definición de hecho | `--dod` | `--set-dod` | `--rm-dod` | `--clear-dod` | **no** |

Todas las de "añade" y "sustituye" son repetibles. `--rm-ac` y `--rm-dod` toman un selector de la
sección 8.4.

**`--set-ac` y `--set-dod` crean elementos nuevos, con claves nuevas y sin marcar**, y las claves de
los elementos anteriores no se reutilizan. Es coherente con 5.1: la clave se asigna al crear el
elemento, y sustituir la lista crea elementos.

### 8.3. Campos de prosa

| Campo | Añade al final | Sustituye | Vacía |
|---|---|---|---|
| descripción | `-d, --desc` | `--set-desc` | `--clear-desc` |
| plan | `--plan` | `--set-plan` | `--clear-plan` |
| notas | `--note` | `--set-notes` | `--clear-notes` |
| resumen final | `--summary` | `--set-summary` | `--clear-summary` |

- Añadir a un campo vacío es lo mismo que fijarlo, así que al crear una tarea las dos columnas
  coinciden y no hay nada que decidir.
- Al añadir sobre contenido existente se intercala una línea en blanco, y cada repetición de la
  bandera en la misma invocación produce su propio párrafo.
- Añadir un valor vacío no hace nada y avisa, según 4.6.

### 8.4. Selectores de criterios

`--check`, `--uncheck`, `--rm-ac`, `--check-dod`, `--uncheck-dod` y `--rm-dod` toman un selector.
Todos son repetibles.

| Selector | Ejemplo | Qué elige |
|---|---|---|
| `all` | `--check all` | todos los elementos de esa lista en esa tarea |
| una clave | `--check 3` | el elemento `#3` |
| un rango de claves | `--check 1-4` | las claves de la 1 a la 4 que existan |
| varias claves | `--check 1,3,7` | esas tres |
| texto | `--check "cubre CRLF"` | el elemento cuyo texto contenga ese fragmento |

**La regla de desambiguación, que hay que implementar tal cual.** El valor se trata como lista de
claves **solo si el valor entero** encaja con `^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$`. En cualquier otro
caso es un texto literal, comas incluidas. Así, `--check "1, 2 y el ultimo"` es una búsqueda de texto
que no encontrará nada y dará error 4, en vez de convertirse en algo a medias.

| Caso límite | Resultado |
|---|---|
| clave que no existe | error 4: `no acceptance criterion #7 on TASK-11 (keys: 1, 3)` |
| texto que no encaja con ninguno | error 4, con los textos de los elementos listados |
| texto que encaja con dos | error 5, con los dos listados |
| rango donde faltan claves intermedias | se aplican las que hay, sin aviso |
| rango invertido, `4-1` | error 2 |
| marcar un elemento ya marcado | se queda marcado, sin aviso, la operación es idempotente |
| `--check all` en una tarea sin criterios | sin efecto, con `warning: TASK-11 has no acceptance criteria` |
| `--check all` sobre varias tareas | válido, cada tarea marca los suyos |
| una clave, un rango, una lista o un texto sobre varias tareas | error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |

**La regla de solape se aplica sobre el conjunto ya resuelto, no sobre el texto del selector.** Si
después de resolver `--check` y `--uncheck` un mismo elemento aparece en los dos conjuntos, es error
2, y da igual que se haya escrito `--check 3 --uncheck 3` o `--check all --uncheck 3`:

```
error: --check and --uncheck both select acceptance criterion #3 of TASK-11
```

### 8.5. Campos escalares

| Campo | Fija | Vacía |
|---|---|---|
| título | `-t, --title` | no se puede, es obligatorio |
| estado | `-s, --status` | no se puede, es obligatorio |
| tipo | `--type` | `--clear-type` |
| prioridad | `--priority` | `--clear-priority` |
| proyecto | `--project` | `--clear-project` |
| hito | `-m, --milestone` | `--clear-milestone` |
| tarea padre | `-p, --parent` | `--clear-parent` |
| fecha límite | `--due` | `--clear-due` |
| orden manual | `--ordinal` | `--clear-ordinal` |
| persona que reporta | `--reporter` | `--clear-reporter` |

Un escalar **nunca** se borra pasándole la cadena vacía, según 4.6.

### 8.6. Campos externos

| Operación | Bandera | Repetible |
|---|---|---|
| fijar una clave | `--ext <clave>=<valor>` | sí |
| quitar una clave | `--rm-ext <clave>` | sí |
| vaciar el mapa entero | `--clear-ext` | no |

**No existe `--set-ext`.** Fijar una clave con `--ext` ya sustituye su valor. Vaciar el mapa entero es
`--clear-ext`, y es la única forma de vaciarlo.

---

## 9. `biso prime`, el arranque de una sesión

### 9.1. Qué resuelve este comando

`biso` **no escribe nunca fuera del tablero, salvo el puntero del proyecto de la sección 3.2.** No
modifica ningún otro fichero del proyecto, ni al crear el tablero ni nunca. Todo lo que hace falta
para empezar a trabajar cabe en un solo comando, cuya salida es un solo mensaje.

`biso prime` se ejecuta al empezar la sesión, de la manera que convenga a quien lo use: como una
orden del asistente, desde un fichero de memoria del proyecto que diga en una línea "ejecuta
`biso prime` antes de tocar tareas", o a mano.

**El criterio de diseño del mensaje es exigente y está pensado para poderse comprobar:** quien lo lea
y no haya visto nunca la herramienta tiene que poder completar un ciclo de trabajo entero, desde
crear una tarea hasta cerrarla, sin leer nada más y sin ninguna interacción adicional.

### 9.2. Firma

```
biso prime [--full] [--limit <n>] [--json]
```

### 9.3. Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--full` | | no | booleano | falso | no | no | `--json` |
| `--limit <n>` | | no | entero >= 0 | 5 | no | no | ninguno |

- `--limit` controla cuántas tareas se listan en la sección `NEXT UP`. Con `0`, la sección desaparece
  y se queda solo el recuento.
- `--full` añade al final la lista completa de banderas de `biso new` y `biso set`. Es para una
  persona que está aprendiendo la herramienta, no para el arranque de un agente.
- `--json` es la bandera global de la sección 3, y aquí es lo único que la restringe: no se puede
  combinar con `--full`, porque el JSON no lleva texto de ayuda.

### 9.4. Qué hace, caso a caso

| Situación | Qué pasa |
|---|---|
| Hay tablero y tiene tareas | Imprime el mensaje de 9.7 por stdout, código 0 |
| Hay tablero y está vacío | Igual, con las dos últimas secciones sustituidas por las tres líneas de 9.8 |
| No hay tablero | Código 8, y por stderr el mensaje de la sección 3.2 |
| Alguna tarea no se puede leer | El mensaje sale igual, con el aviso de 4.12, código 0 |
| `--limit` negativo | Código 2 |

`biso prime` **no escribe nada, nunca**, y no necesita acceso exclusivo. Es seguro llamarlo en
paralelo desde varias sesiones y mientras otro proceso escribe.

### 9.5. El presupuesto de tamaño

El mensaje tiene un **tope duro de 5.120 bytes**, que se comprueba en la suite de pruebas y se reparte
en dos mitades exactas:

- **La parte fija no pasa de 3.072 bytes.** Es la línea de título, `COMMANDS`, `FIELD FLAGS`, `RULES` y
  el párrafo final ("Pick one, ..."): nada de esto depende del contenido del tablero.
- **El resumen del tablero no pasa de 2.048 bytes.** Es el bloque `BOARD` (nombre, recuento por
  estado, vocabularios, identidad), `IN PROGRESS`, `NEXT UP` y la línea de recorte: todo lo que
  cambia según qué haya en el tablero.

Los bloques no son contiguos entre sí, así que hay líneas en blanco de separación entre ellos: **cada
línea en blanco se cuenta en la mitad del bloque que la precede.** Con esta regla, la línea en blanco
que sigue al título es parte fija, la que sigue a `BOARD` es resumen, las que siguen a `COMMANDS`,
`FIELD FLAGS` y `RULES` son parte fija, y las que siguen a `IN PROGRESS` y a `NEXT UP` son resumen.

Si el resumen no cupiera en su mitad, se recorta el número de tareas listadas en `NEXT UP` antes que
cualquier otra cosa, y la línea de recuento lo dice.

El texto literal de la sección 9.7 ocupa **4.062 bytes** con el tablero del ejemplo: **2.963** de
parte fija y **1.099** de resumen. Las dos mitades caben dentro de su tope.

### 9.6. Qué entra en el mensaje y qué se relega a `--help`

El criterio es uno solo: **entra lo que no se puede adivinar y hace falta para la primera acción; se
queda fuera lo que se puede consultar en el momento exacto en que hace falta.**

Entra:

- Las ocho órdenes del ciclo de trabajo con su forma de uso. Quien no sabe que existe `biso finish`
  no va a escribir `biso finish --help`.
- **Los nombres de todas las banderas de campo**, en una rejilla de cinco líneas.
- El vocabulario real de este tablero, con **el recuento por estado** y con la marca de cuál es el
  estado de las tareas nuevas, cuál el activo y cuál el terminal.
- Las diez reglas que no son adivinables.
- Los códigos de salida, en dos líneas.
- El estado del tablero: lo que está en curso y lo más urgente de lo que no ha empezado.

Se queda fuera, y va a `biso <cmd> --help`:

- Los valores, los tipos y las incompatibilidades de cada bandera. El mensaje da los nombres, que es
  lo que no se puede adivinar; la ayuda da el detalle, que es lo que se consulta cuando se necesita.
- El formato de lote de `biso new --from` y el esquema JSON completo.
- `biso export`, `biso config`, `biso doctor`, `biso archive`, `biso where` y `biso board`, que no
  aparecen en el ciclo de trabajo normal.
- Todos los casos límite: el rango invertido, las dos entradas estándar, la coma dentro de una
  etiqueta.
- La política de cuándo merece la pena crear una tarea, que es una decisión del proyecto y no de la
  herramienta. El mensaje la resume en una línea y no la desarrolla.

### 9.7. La salida literal

Esto es exactamente lo que `biso prime` imprime por stdout con un tablero de ejemplo. No imprime nada
por stderr.

```
biso 1.0.0 - the task board of this project. This message is all you need to start.

BOARD  Kex
  Ideas 12 | To Do 42 | In Progress 3 | Blocked 5 | Done 187
  new tasks start in Ideas; `biso start` moves to In Progress; `biso finish` to Done
  types       idea, memory, task, bug, docs
  priorities  high, medium, low
  you are     @claude

COMMANDS  (`biso <cmd> --help` for the detail of any flag)
  biso ls [-s STATUS] [--type T] [-l LABEL] [--mine] [--search TEXT]
  biso get <ref> [--section ac]
  biso new "TITLE" [-d TEXT] [--ac TEXT]... [--type T] [--priority P]
  biso start <ref>... [--plan TEXT]
  biso note <ref> "TEXT"
  biso finish <ref>... [--summary "TEXT"] [--check all]
  biso set <ref>... [any field flag]
  biso comment <ref> "TEXT" [--comment-author @who]

FIELD FLAGS  (same names, same meaning, in every command above that writes)
  -t --title  -s --status  --type   --priority  --project      -a --assignee
  -l --label  -d --desc    --ac     --dod       --plan         --note
  --summary   --dep        --ref    --doc       --file         -m --milestone
  -p --parent --due        --ordinal --ext K=V  --reporter     --comment
  --check     --uncheck

RULES  (none of these are guessable; they are the whole learning curve)
  1. Every write goes through biso. Nothing else touches the board.
  2. A bare field flag ADDS. Replacing and removing are explicit: --label X
     adds, --set-label X replaces the list, --rm-label X drops one, and
     --clear-label empties it. Same four shapes for every list field.
  3. <ref> is an id (TASK-12), a bare number (12) or free text ("CRLF"). Text
     matching several tasks is an error that lists them, never a guess. `note`
     and `comment` take one <ref>; `set`, `start` and `finish` take several.
  4. Filters reject values this board does not have: `-s Pending` is an error,
     not an empty list. Case, spaces, hyphens and underscores are ignored, so
     `-s todo`, `-s "To Do"` and `-s TO_DO` are one and the same filter. An
     empty list is therefore a fact about the board that you can act on.
  5. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says
     on stderr what it left out. --all lifts the limit, --any-status includes
     Done, --archived reaches the archive.
  6. --check and --uncheck take all, 3, 1-4, 1,3,7 or the criterion text. The
     numbers are the stable #N keys that `biso get` shows, and they never
     shift when one criterion is removed.
  7. `biso new` prints the new id and nothing else. Every other write prints one
     line per task: id, status, criteria, urgency. Add --print for the whole
     record, or --json for a versioned envelope.
  8. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.
  9. Long text: a real newline works, and so do -d @file.md and -d - for stdin.
 10. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,
     6 precondition not met, 7 environment, 8 no board here, 9 nothing written.

IN PROGRESS
  TASK-11  In Progress  bug   high    Normalize CRLF in the diff         ac 1/2  @claude  -
  TASK-52  In Progress  task  low     Document the release checklist     ac 0/1  -        -
  TASK-40  In Progress  task  medium  Split the config loader            ac 0/2  @claude  -

NEXT UP  (not started, not done, by urgency)
  TASK-7   To Do        bug   high    Crash on an empty repository       ac 0/4  -        2026-09-08
  TASK-19  To Do        task  high    Retry the upload on 5xx            ac 0/2  -        -
  TASK-23  Blocked      docs  medium  Rewrite the install section        ac 0/1  -        -
  TASK-31  To Do        task  medium  Cache the parsed manifest          ac 0/3  -        -
  TASK-44  Ideas        bug   low     Wrong column width on narrow ttys  ac 0/1  -        -
  54 more not started: `biso ls --all`

Pick one, `biso start <ref> --plan "..."`, work, `biso note <ref> "..."` as you go,
and close with `biso finish <ref> --check all --summary "..."`. That is the loop.
Create a task when the work needs planning or review; do small edits directly.
```

Cómo se calcula el resumen, para que la implementación sea única:

- La línea de recuento tiene **un número por cada estado configurado**, en el orden en que están
  configurados, y cuenta las tareas no archivadas de ese estado. No hay ninguna categoría inventada
  como "abiertas" que no se corresponda con una columna del tablero.
- `IN PROGRESS` lista las tareas del estado activo, ordenadas por la regla de orden de 10.4, sin
  límite.
- `NEXT UP` lista las tareas que no están ni en el estado activo ni en el terminal, ordenadas igual,
  y corta en `--limit`.
- La línea de recuento final dice cuántas tareas quedan fuera de `NEXT UP` por el corte.
- Las filas usan exactamente el algoritmo de columnas de `biso ls` de la sección 10.4, con una
  diferencia declarada aquí: el ancho de las columnas 1 a 7 se calcula sobre las filas de
  `IN PROGRESS` y `NEXT UP` juntas, para que las dos secciones se lean como una sola tabla.

### 9.8. Tablero vacío

Cuando no hay ninguna tarea, `IN PROGRESS` y `NEXT UP` se sustituyen por esto, y el resto del mensaje
no cambia:

```
THE BOARD IS EMPTY
  Create the first one:
  biso new "Title" -d "What and why" --ac "How we will know it works"
```

### 9.9. El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "prime",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "tool": { "name": "biso", "version": "1.0.0" },
    "board": {
      "name": "Kex",
      "me": "@claude",
      "statuses": ["Ideas", "To Do", "In Progress", "Blocked", "Done"],
      "defaultStatus": "Ideas",
      "activeStatus": "In Progress",
      "terminalStatus": "Done",
      "types": ["idea", "memory", "task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "extensions": ["trello.card"],
      "countByStatus": { "Ideas": 12, "To Do": 42, "In Progress": 3, "Blocked": 5, "Done": 187 }
    },
    "inProgress": [
      { "id": "TASK-11", "title": "Normalize CRLF in the diff", "status": "In Progress",
        "type": "bug", "priority": "high", "assignees": ["@claude"], "due": null,
        "acDone": 1, "acTotal": 2, "urgency": 19.0 }
    ],
    "nextUp": [
      { "id": "TASK-7", "title": "Crash on an empty repository", "status": "To Do",
        "type": "bug", "priority": "high", "assignees": [], "due": "2026-09-08",
        "acDone": 0, "acTotal": 4, "urgency": 18.2 }
    ],
    "notStartedHidden": 54
  }
}
```

Las reglas y los nombres de las banderas no viajan en el JSON: quien pide JSON es un programa, y un
programa no necesita que le expliquen que el nombre desnudo añade.

### 9.10. Códigos de salida

| Desenlace | Código |
|---|---:|
| Mensaje impreso | 0 |
| Mensaje impreso con alguna tarea ilegible | 0, con aviso |
| `--limit` negativo, o `--full` junto con `--json` | 2 |
| No hay tablero | 8 |

### 9.11. `biso prime --help`

```
Usage: biso prime [options]

Print everything needed to start working on this board: the commands, the field
flags, the rules that are not guessable, the board vocabulary and what is in
flight. Run it once at the start of a session. It writes nothing.

Options:
  --full          also list every flag of `biso new` and `biso set` in detail
  --limit <n>     how many not-started tasks to show (default 5, 0 hides them)
  --json          machine-readable envelope instead of the message
  -h, --help      show this help

Exit codes:
  0  message printed
  2  bad usage
  8  no board here

Examples:
  biso prime
  biso prime --limit 10
  biso -C ~/work/kex prime
```

---

## 10. Los comandos

Diecisiete comandos. Los nueve primeros son el ciclo de trabajo y aparecen en `biso --help`; los ocho
restantes son de administración y aparecen en `biso help all`.

| Comando | Qué hace | En `biso --help` |
|---|---|---|
| `prime` | El mensaje de arranque de la sección 9 | sí |
| `ls` | Lista tareas | sí |
| `get` | Muestra una tarea | sí |
| `new` | Crea una o muchas tareas | sí |
| `set` | Cambia campos de una o varias tareas | sí |
| `start` | Toma una tarea y la pone en curso | sí |
| `note` | Añade una nota de implementación | sí |
| `comment` | Añade un comentario con autor | sí |
| `finish` | Cierra una tarea | sí |
| `archive` | Saca una tarea del tablero activo | no |
| `export` | Vuelca el tablero en el formato de entrada de `new --from` | no |
| `init` | Crea un tablero | no |
| `where` | Explica qué tablero se está usando y por qué | no |
| `config` | Lee y cambia la configuración | no |
| `doctor` | Comprueba y repara la integridad | no |
| `board` | Abre la interfaz interactiva | no |
| `help` | La ayuda de primer nivel y la de cada comando | no |

**Las banderas globales de la sección 3 valen en todos ellos y no se repiten en las tablas de
parámetros de cada comando.** Un comando solo las menciona cuando le impone una restricción
adicional, y esas restricciones son exactamente cuatro en todo el documento: `prime --full` no se
combina con `--json`, `config` solo acepta `--json` en su subcomando `list`, `export` rechaza
`--json` con código 2, y `--print` con `--dry-run` no valen en los comandos de lectura según la
regla de la sección 3.

El caso de `export` merece una línea, porque es el único comando cuya salida ya es JSON sin pedirlo:
son objetos JSON, uno por línea, y `--json` pide el sobre único de la sección 12, que es otra forma
distinta. Pasarlo es error 2:

```
error: --json does not apply to export
       its output is already one JSON object per line
```

**Todos los comandos que escriben aceptan todas las banderas de campo de la sección 8**, con el mismo
nombre y el mismo significado. Eso vale para `new`, `set`, `start`, `note`, `comment`, `finish` y
`archive`. Lo que distingue a unos de otros no es qué campos aceptan, sino qué hacen por defecto. Las
tablas de parámetros de cada comando enumeran solo lo que es propio de ese comando.

### 10.1. `biso init`

#### Firma

```
biso init [<name>] [--at <location>] [--statuses <list>] [--types <list>]
          [--priorities <list>] [--projects <list>] [--extensions <list>]
          [--prefix <text>] [--overwrite-config]
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<name>` | | no | texto | el nombre del proyecto | no | no | |
| `--at <location>` | | no | localizador | el que el almacenamiento asocia por defecto a este proyecto | no | no | |
| `--statuses <list>` | | no | lista | `To Do, In Progress, Done` | sí | sí | |
| `--types <list>` | | no | lista | `task, bug, docs` | sí | sí | |
| `--priorities <list>` | | no | lista | `high, medium, low` | sí | sí | |
| `--projects <list>` | | no | lista | vacía | sí | sí | |
| `--extensions <list>` | | no | lista | vacía | sí | sí | |
| `--prefix <text>` | | no | texto de solo letras | `TASK` | no | no | |
| `--overwrite-config` | | no | booleano | falso | no | no | |

`--at` acepta el nombre o el localizador de un tablero, en la forma que el almacenamiento imponga,
igual que la bandera global `--board`.

#### Comportamiento

Crea un tablero vacío con su configuración. **No escribe nunca fuera del tablero**, salvo el puntero
del proyecto que se describe a continuación.

`init` escribe, además del tablero, **el puntero del proyecto** de la sección 3.2, y lo hace siempre
que el tablero no quede dentro del propio proyecto, es decir, siempre que se use `--at` apuntando
fuera. Es la única cosa que `init` escribe fuera del tablero. La salida dice siempre si el puntero se
ha creado.

| Caso | Qué pasa |
|---|---|
| Ya hay un tablero accesible desde aquí | Error 2, salvo con `--overwrite-config`, que reescribe la configuración y **nunca toca las tareas** |
| El primer estado de `--statuses` | Se guarda como `default_status` |
| El penúltimo estado | Se guarda como `active_status` |
| El último estado | Se guarda como `terminal_status` |
| Menos de dos estados | Error 2 |
| Exactamente dos estados | Válido. El primero es a la vez `default_status` y `active_status`, así que `biso start` no cambia el estado: solo asigna y añade el plan. Se avisa en la salida |
| `--prefix` con algo que no sean letras | Error 2 |
| `--at` a un localizador donde no se puede escribir | Error 7 |

**Los tres estados especiales se guardan como valores explícitos en la configuración, no como
posiciones.** Cambiar `statuses` después no los mueve nunca. Si al cambiar `statuses` uno de los tres
deja de existir, el comando que lo hace falla, según la sección 10.10.

#### Salida

```
Created board "Kex"
  statuses    To Do (default) | In Progress (active) | Done (terminal)
  types       task, bug, docs
  priorities  high, medium, low
  prefix      TASK
This project now points at that board.
Run `biso prime` to see how to use it.
```

La última línea sobre el puntero solo aparece cuando el puntero se ha creado. Con exactamente dos
estados, se añade además `warning: with two statuses, "start" cannot change the status; it still
assigns and records the plan`.

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "init",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "board": {
      "name": "Kex",
      "statuses": ["To Do", "In Progress", "Done"],
      "defaultStatus": "To Do", "activeStatus": "In Progress", "terminalStatus": "Done",
      "types": ["task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "taskPrefix": "TASK"
    },
    "pointerCreated": false
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tablero creado | 0 |
| Ya existía y no hay `--overwrite-config` | 2 |
| Argumentos inválidos | 2 |
| No se puede escribir | 7 |

#### `biso init --help`

```
Usage: biso init [name] [options]

Create a task board for this project. It writes the board and, when the board
lives outside the project, a pointer inside the project so every copy of the
project finds the same board. It never writes outside the board otherwise.

Arguments:
  name                   board name (default: the project directory name)

Options:
  --at <location>        where the board lives, in whatever form the storage
                         takes (default: what the storage associates with this
                         project)
  --statuses <list>      comma-separated. The first is the default status of a
                         new task, the second to last is the active one that
                         `biso start` sets, and the last is the terminal one
                         that `biso finish` sets. All three are then stored as
                         explicit values and never move again.
                         (default: "To Do,In Progress,Done")
  --types <list>         comma-separated (default: "task,bug,docs")
  --priorities <list>    comma-separated (default: "high,medium,low")
  --projects <list>      comma-separated (default: none)
  --extensions <list>    comma-separated declared external field keys, such as
                         trello.card (default: none)
  --prefix <text>        task id prefix, letters only (default: TASK)
  --overwrite-config     replace the configuration of an existing board,
                         keeping every task
  -h, --help             show this help

Exit codes:
  0  board created
  2  bad usage, or a board is already reachable from here
  7  cannot write there

Examples:
  biso init
  biso init Kex --statuses "Ideas,To Do,In Progress,Blocked,Done"
  biso init Kex --at kex-board --prefix KEX --extensions trello.card
```

---

### 10.2. `biso where`

#### Firma

```
biso where [--json]
```

Sin parámetros propios.

#### Comportamiento

Dice qué tablero se está usando y por qué regla de la sección 3.2 se ha elegido. Es el comando al que
remite el error de código 8, y el que hace visible una resolución que de otro modo sería invisible.

| Caso | Qué pasa |
|---|---|
| Hay tablero | Lo imprime con la regla que lo eligió, código 0 |
| No hay tablero | Imprime lo que ha buscado y dónde, código 8 |
| Hay más de un candidato | Imprime el elegido y los descartados, con el motivo, código 0 |

#### Salida

```
board    Kex
source   project pointer at the root of this project
me       @claude
tasks    249 active, 31 archived, highest id ever assigned TASK-290
```

Y cuando no hay ninguno, por stderr y con código 8:

```
error: no board here, and none configured for this project
searched  --board:      not given
          BISO_BOARD:   not set
          pointer:      not found between this directory and the project root
          local board:  not found between this directory and the project root
hint: `biso init` creates one
```

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "where",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "board": "Kex",
    "source": "project pointer at the root of this project",
    "me": "@claude",
    "counts": { "active": 249, "archived": 31, "highestIdEverAssigned": "TASK-290" }
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tablero encontrado | 0 |
| No hay tablero | 8 |

#### `biso where --help`

```
Usage: biso where [options]

Say which board is in use and which rule picked it. Run it when a command
answers "no board here" and you expected one.

Options:
      --json     machine-readable envelope
  -h, --help     show this help

Exit codes:
  0  a board is in use
  8  no board here

Examples:
  biso where
  biso -C ~/work/kex where
```

---

### 10.3. `biso new`

#### Firma

```
biso new [<title>] [--start] [--from <file|->] [cualquier bandera de campo de la seccion 8]
```

#### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<title>` | | sí, salvo con `--from` | texto | | no | no | `--from` |
| `--start` | | no | booleano | falso | no | no | `-s`, `--from` |
| `--from <file>` | | no | ruta o `-` | | no | no | `<title>` y todas las de campo |

Todas las banderas de campo de la sección 8 valen aquí. En una tarea nueva no hay nada que sustituir
ni que quitar, así que `--set-*`, `--rm-*` y `--clear-*` se aceptan y hacen lo mismo que el nombre
desnudo, salvo `--clear-*`, que no hace nada y avisa. Las que se usan de verdad al crear son
`-d/--desc`, `--ac`, `--dod`, `--type`, `--priority`, `-l/--label`, `-a/--assignee`, `--ref`,
`--doc`, `--dep`, `-m/--milestone`, `-p/--parent`, `--due`, `--ordinal`, `--project`, `--reporter`,
`--ext`, `--plan`, `--note`, `--summary` y `--comment`.

- **`--start`** crea la tarea directamente en el estado activo y asignada a `me`.
- **`--comment` funciona al crear**, igual que en cualquier otro comando de escritura.
- **`--plan`, `--note` y `--summary` no están restringidos por el estado.** Se pueden escribir al
  crear, en cualquier estado.
- **`default_assignee` de la configuración asigna esa persona cuando no se ha pasado `-a`.** Con
  `--start` y sin `-a`, `--start` asigna `me` en su lugar; `default_assignee` solo se aplica cuando
  ni `-a` ni `--start` han asignado a nadie.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Título vacío o solo espacios | Error 2: `error: title cannot be empty` |
| Título muy largo | Se acepta entero, sin recortar |
| Título repetido | Se acepta sin aviso. Dos tareas pueden llamarse igual, para eso está el identificador |
| Valor fuera de un vocabulario cerrado | Error 3, con la lista de válidos |
| `--dep` a una tarea inexistente | Error 4. **Las dependencias se validan al escribirlas** |
| `--dep` a la propia tarea, o que cerraría un ciclo | Error 2 |
| `--parent` inexistente, o que cerraría un ciclo | Error 4 y error 2 respectivamente |
| `--ext` con una clave no declarada | Error 3 |
| `--due` con formato incorrecto | Error 2, señalando `YYYY-MM-DD` |
| `--due` en el pasado | Se acepta, con aviso |
| `-d @fichero` que no existe | Error 4 |
| Todo bien | Se crea la tarea, código 0 |

#### Salida

Por defecto, **una línea por tarea creada, con el identificador y nada más**:

```
TASK-101
```

`biso new` es el único comando de escritura cuya salida por defecto es distinta de la línea de estado
de 10.6, y así está dicho en el mensaje de arranque.

Con `--print`, después de la línea del identificador viene la ficha completa en el formato de
`biso get`. Con `--quiet`, solo el identificador y ninguna nota.

#### El modo lote

```
biso new --from tareas.ndjson
biso new --from -
biso new --from tareas.ndjson --dry-run
```

La entrada es **NDJSON**: un objeto JSON por línea. Las líneas vacías y las que empiezan por `#` se
ignoran. Las claves son las del modelo de datos de la sección 5, en `camelCase`.

Ejemplo de una línea, con todos los tipos compuestos:

```json
{"id":"TASK-101","title":"El diff no normaliza CRLF","type":"bug","priority":"high","status":"Done","description":"...","labels":["parser"],"references":["docs/bugs/BUG-02.md"],"dependencies":["TASK-90"],"ext":{"trello.card":"5f2a8c1e"},"acceptanceCriteria":[{"key":1,"text":"El diff ignora el CRLF","checked":true},{"key":3,"text":"Hay un test","checked":false}],"definitionOfDone":[{"key":1,"text":"Revisado","checked":true}],"comments":[{"author":"@avilches","createdAt":"2026-08-14T10:22:00Z","body":"Reportado desde Windows"}],"createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
```

Las reglas del lote, todas obligatorias:

- **`acceptanceCriteria` y `definitionOfDone` aceptan dos formas.** Una cadena, que crea un elemento
  sin marcar con la siguiente clave libre, o un objeto con `key`, `text` y `checked`. Las dos formas
  se pueden mezclar dentro de la misma lista. Una `key` repetida dentro de la misma tarea es un fallo
  de validación.
- **El contador de claves de cada lista se sitúa por encima de la clave mayor importada**, de modo que
  un criterio añadido después nunca choca con uno importado. El contador no es una clave del formato:
  se deduce.
- **`comments` es una lista de objetos** con `author`, `createdAt` y `body`. `createdAt` es opcional y,
  si falta, se pone el instante de la importación.
- **`id`, `createdAt` y `updatedAt` se aceptan aquí y solo aquí.** Un `id` ya ocupado es un fallo de
  validación; un `id` libre se reserva y el tablero no lo volverá a asignar.
- **`archived` se acepta como booleano.** Por defecto, si la clave no aparece, la tarea se crea sin
  archivar. Ningún otro comando tiene una bandera de campo para él: fuera de la importación,
  archivar se hace con `biso archive`.
- **Una clave desconocida es un fallo de validación, no se ignora.** Ni la línea ni el lote se
  escriben, y el mensaje dice la línea y la clave.
- **Los campos derivados de la sección 5 no se aceptan.** En la entrada son claves desconocidas y
  por tanto un fallo de validación.
- **Se valida el fichero entero antes de escribir nada**, y se aplica la garantía de todo o nada de
  la sección 4.10.
- Un lote no admite `--start` ni ninguna bandera de campo: todo va en el fichero.

Salida del lote, una línea por tarea, en el orden del fichero:

```
TASK-101
TASK-102
TASK-103
```

Salida de `--dry-run` cuando todo está bien, por stderr y con código 0:

```
242 tasks would be created, nothing was written (--dry-run)
```

Y cuando no, por stderr y con código 9, **con todos los fallos, no solo el primero**:

```
error: 3 of 242 lines are invalid, nothing was written
  line 47: unknown status: "Pendiente" (valid: Ideas, To Do, In Progress, Blocked, Done)
  line 88: unknown key: "trelloCard"
  line 201: title cannot be empty
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea o lote creado | 0 |
| `--dry-run` que habría funcionado | 0 |
| Falta el título, banderas incompatibles, fecha mal formada, ciclo de dependencias o de padres | 2 |
| Valor fuera de un vocabulario, clave de extensión no declarada, entrada no interpretable | 3 |
| `--dep` o `--parent` a una tarea que no existe, o fichero de `@` que no existe | 4 |
| `--dep` o `--parent` por texto con varias coincidencias | 5 |
| Cualquier fallo de validación en un lote, o un `--dry-run` que no pasa | 9 |
| El almacén falla, o no se obtiene el acceso exclusivo | 7 |
| No hay tablero | 8 |

#### `biso new --help`

```
Usage: biso new <title> [options]
       biso new --from <file|-> [options]

Create a task and print its id. Every field flag of `biso set` works here.

Arguments:
  title                      task title (required unless --from is given)

Most used:
  -d, --desc <text>          description; repeat to append paragraphs
      --ac <text>            add an acceptance criterion; repeatable
      --dod <text>           add a definition-of-done item; repeatable
      --type <value>         configured type
      --priority <value>     configured priority
  -s, --status <value>       configured status (default: the board default)
  -l, --label <value>        add a label; repeatable or comma-separated
  -a, --assignee <@who>      add an assignee; repeatable or comma-separated
      --dep <ref>            add a dependency; validated, repeatable
      --due <YYYY-MM-DD>     due date
      --comment <text>       add a discussion comment; repeatable
      --plan <text>          implementation plan
      --start                create it already in the active status, assigned
                             to you

Every other field flag of `biso set --help` is accepted too.

Batch:
      --from <file|->        NDJSON, one task object per line. The only place
                             where id, createdAt, updatedAt, criterion keys and
                             comment timestamps can be given. Validated whole
                             before anything is written.

Any text option also takes @file to read a file, or - to read stdin.

Exit codes:
  0  created            4  a referenced task or file does not exist
  2  bad usage          5  a text reference matched several tasks
  3  unknown value      7  the board could not be written
  9  batch or --dry-run validation failed, nothing was written
                        8  no board here

Examples:
  biso new "Normalize CRLF in the diff" --type bug --priority high
  biso new "Add OAuth" --ac "Login succeeds" --ac "Token refreshes"
  biso new "Rewrite the installer" -d @docs/installer.md --start
  biso new --from tasks.ndjson --dry-run
```

---

### 10.4. `biso ls`

#### Firma

```
biso ls [-s <status>]... [--not-status <status>]... [--any-status] [--archived] [--only-archived]
        [--type <v>]... [--priority <v>]... [--project <v>]...
        [-l <label>]... [--label-or <label>]... [-a <@who>]... [--mine] [--unassigned]
        [-m <milestone>] [-p <ref>] [--ready] [--blocked] [--overdue] [--due-before <date>]
        [--search <text>] [--unchecked]
        [--sort <field>] [--reverse] [--limit <n>] [--all] [--ids] [--count]
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--status <v>` | `-s` | no | vocabulario | todos menos el terminal | sí | sí | `--any-status` |
| `--not-status <v>` | | no | vocabulario | | sí | sí | `--any-status` |
| `--any-status` | | no | booleano | falso | no | no | `-s`, `--not-status` |
| `--archived` | | no | booleano | falso | no | no | `--only-archived` |
| `--only-archived` | | no | booleano | falso | no | no | `--archived` |
| `--type <v>` | | no | vocabulario | | sí | sí | |
| `--priority <v>` | | no | vocabulario | | sí | sí | |
| `--project <v>` | | no | vocabulario | | sí | sí | |
| `--label <l>` | `-l` | no | etiqueta | | sí | sí | |
| `--label-or <l>` | | no | etiqueta | | sí | sí | |
| `--assignee <@w>` | `-a` | no | persona | | sí | sí | `--mine`, `--unassigned` |
| `--mine` | | no | booleano | falso | no | no | `-a`, `--unassigned` |
| `--unassigned` | | no | booleano | falso | no | no | `-a`, `--mine` |
| `--milestone <m>` | `-m` | no | texto libre | | no | no | |
| `--parent <ref>` | `-p` | no | referencia | | no | no | |
| `--ready` | | no | booleano | falso | no | no | `--blocked` |
| `--blocked` | | no | booleano | falso | no | no | `--ready` |
| `--overdue` | | no | booleano | falso | no | no | |
| `--due-before <d>` | | no | `YYYY-MM-DD` | | no | no | |
| `--search <text>` | | no | texto libre | | no | no | |
| `--unchecked` | | no | booleano | falso | no | no | |
| `--sort <field>` | | no | `urgency`, `id`, `ordinal`, `due`, `updated`, `created`, `title` | el orden de abajo | no | no | |
| `--reverse` | | no | booleano | falso | no | no | |
| `--limit <n>` | | no | entero >= 0 | 30 | no | no | `--all` |
| `--all` | | no | booleano | falso | no | no | `--limit` |
| `--ids` | | no | booleano | falso | no | no | `--count` |
| `--count` | | no | booleano | falso | no | no | `--ids` |

Reglas de combinación de filtros:

- **Filtros de campos distintos se combinan con `y`.** `-s "To Do" --type bug` son las que cumplen las
  dos cosas.
- **Valores repetidos del mismo campo se combinan con `o`.** `--type bug --type docs` son las de
  cualquiera de los dos tipos. Esto vale para `--status`, `--type`, `--priority`, `--project`,
  `--assignee` y `--label-or`.
- **`-l/--label` es la única que se combina con `y`.** `-l frontend -l bug` son las que llevan las
  dos. Para el `o` está `--label-or`, que valida igual.
- **`--unchecked` apaga la comprobación de existencia de `-l`, `--label-or` y `-a`, y solo esa.** No
  cambia cómo se combinan ni afecta a ningún otro filtro. Los vocabularios cerrados siguen validando.
- **El estado terminal se excluye por defecto**, y `--any-status` es la única forma de incluirlo.
- **Las archivadas se excluyen por defecto.** `--archived` las añade a las vivas y `--only-archived`
  deja solo las archivadas.

#### La regla de orden, completa

`--sort` sin valor aplica el orden por defecto, que es esta tupla, en este orden y sin excepciones:

1. Las tareas que tienen `ordinal` van antes que las que no lo tienen.
2. Entre las que lo tienen, `ordinal` ascendente.
3. Entre las que no lo tienen, `urgency` descendente.
4. Cualquier empate se rompe por identificador ascendente, siempre.

Un `--sort` explícito sustituye los pasos 1 a 3 por ese campo, ascendente salvo `urgency`, que es
descendente por ser una medida de prioridad, y el paso 4 se sigue aplicando. `--reverse` invierte el
resultado final, incluido el desempate. **El orden nunca depende del estado**, porque el listado no
agrupa por estado.

Un `--sort due` o `--sort ordinal` sobre tareas que no tienen ese campo las pone al final, en bloque,
ordenadas por identificador.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Filtro con un valor fuera del vocabulario | Error 3, con la lista de válidos |
| `-l` con una etiqueta que no existe en el tablero | Error 3, con las cinco más parecidas |
| Lo mismo con `--unchecked` | Se acepta, y probablemente no devuelve nada |
| Filtro válido sin resultados | Ninguna línea por stdout, `note: no tasks match` por stderr, código **0** |
| Hay más resultados que el límite | Se imprimen los primeros y sale el aviso de recorte |
| `--limit 0` | No imprime ninguna fila, solo el aviso de recorte con el total. Es la forma de contar sin `--count` |
| `--count` | Un número por stdout y nada más |
| `--ids` | Identificadores, uno por línea, sin cabeceras ni columnas |
| Alguna tarea ilegible | Se salta, con el aviso de 4.12, y el resto del listado es válido |

#### Salida

Ocho columnas fijas, separadas por dos espacios, en este orden y con estos contenidos:

| Columna | Contenido | Cuando está vacío |
|---|---|---|
| 1 | identificador | nunca lo está |
| 2 | estado | nunca lo está |
| 3 | tipo | `-` |
| 4 | prioridad | `-` |
| 5 | título, recortado a **100 caracteres siempre**, con `...` al final si se recorta | nunca lo está |
| 6 | `ac <marcados>/<total>` | `-` si la tarea no tiene criterios |
| 7 | primera persona asignada, con `+<n>` si hay más | `-` |
| 8 | fecha límite | `-` |

**El formato se calcula así, en dos pasos, siempre en este orden:**

1. El título de cada tarea se recorta primero a 100 caracteres, con `...` al final si se ha
   recortado. Esto pasa antes de calcular ningún ancho de columna.
2. Para cada una de las columnas 1 a 7, el ancho de esa columna es la longitud del valor más largo
   que le corresponde entre las filas que se van a imprimir en esta llamada, y cada valor se rellena
   con espacios a la derecha hasta ese ancho. **La columna 8 nunca se rellena**, porque es la última
   y no hay nada después que alinear.

Entre columna y columna van siempre **dos espacios literales**, se haya rellenado o no la columna
anterior. Con estas cuatro tareas, el título más largo mide 28 caracteres y por eso la columna 5 se
rellena a ese ancho, no a uno fijo:

```
TASK-7   To Do        bug   high    Crash on an empty repository  ac 0/4  -        2026-09-08
TASK-11  In Progress  bug   high    Normalize CRLF in the diff    ac 1/2  @claude  -
TASK-19  To Do        task  high    Retry the upload on 5xx       ac 0/2  -        -
TASK-23  Blocked      docs  medium  Rewrite the install section   ac 0/1  @sara+1  -
```

Y por stderr, siempre que se haya recortado:

```
warning: 212 more tasks match; showing 30 of 242
hint: narrow with -s, --type or -l, or ask for everything with --all
```

Con `--ids`:

```
TASK-7
TASK-11
```

Con `--count`:

```
242
```

**No hay agrupación por estado.** El estado es una columna más, para que cada línea se pueda tratar
igual que las demás.

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.list",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "tasks": [
      {
        "id": "TASK-11",
        "title": "Normalize CRLF in the diff",
        "status": "In Progress",
        "type": "bug",
        "priority": "high",
        "project": null,
        "assignees": ["@claude"],
        "reporter": "@avilches",
        "labels": ["parser"],
        "milestone": null,
        "parent": null,
        "dependencies": [],
        "references": ["docs/bugs/BUG-02.md"],
        "documentation": [],
        "modifiedFiles": [],
        "due": null,
        "ordinal": null,
        "createdAt": "2026-09-06T09:12:04Z",
        "updatedAt": "2026-09-06T11:40:18Z",
        "acDone": 1,
        "acTotal": 2,
        "dodDone": 0,
        "dodTotal": 0,
        "commentCount": 1,
        "urgency": 19.0,
        "blocks": ["TASK-40"],
        "ready": true,
        "blocked": false,
        "archived": false,
        "ext": { "trello.card": "5f2a8c1e3b9d4a7f6e0c2b81" }
      }
    ],
    "shown": 30,
    "matched": 242,
    "hidden": 212,
    "truncated": true,
    "skipped": [],
    "sort": "default",
    "filters": { "status": ["Ideas", "To Do", "In Progress", "Blocked"], "type": [], "label": [] }
  }
}
```

**El listado nunca trae el cuerpo de la tarea**: ni descripción, ni plan, ni notas, ni criterios, ni
comentarios. Para eso está `biso get`. Los nueve campos derivados de la sección 5 sí están todos,
`blocks` incluido. `truncated` es explícito para que nadie tenga que comparar `shown` con `matched`,
y `skipped` lleva los identificadores de las tareas ilegibles que se han saltado.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Listado, incluso vacío o con tareas saltadas | 0 |
| Un valor de filtro no existe en el tablero | 3 |
| Banderas incompatibles, `--limit` negativo, `--sort` inventado, fecha mal formada | 2 |
| `--parent` a una tarea que no existe | 4 |
| `--parent` por texto con varias coincidencias | 5 |
| `--mine` sin ninguna identidad configurada (3.1) | 6 |
| El almacén no responde | 7 |
| No hay tablero | 8 |

#### `biso ls --help`

```
Usage: biso ls [options]

List tasks, one per line. Shows 30 by default, hides the Done ones and the
archived ones, and says on stderr what it left out. A filter value the board
does not have is an error, never an empty list, so an empty list is a fact.

Filters (repeat or comma-separate; same field is OR, different fields are AND):
  -s, --status <value>       configured status (default: all but the terminal)
      --not-status <value>   exclude a status
      --any-status           include the terminal status too
      --archived             include archived tasks
      --only-archived        only archived tasks
      --type <value>         configured type
      --priority <value>     configured priority
      --project <value>      configured project
  -l, --label <value>        label; several labels are ANDed
      --label-or <value>     label; several are ORed
  -a, --assignee <@who>      assignee
      --mine                 assigned to you
      --unassigned           assigned to nobody
  -m, --milestone <text>     milestone, matched like any board value
  -p, --parent <ref>         subtasks of this task
      --ready                nothing unfinished blocks it
      --blocked              something unfinished blocks it
      --overdue              past its due date
      --due-before <date>    due before YYYY-MM-DD
      --search <text>        free text; see `biso get --help` for the scope
      --unchecked            do not check that the labels and assignees you
                             filter by exist on the board; nothing else changes

Shape:
      --sort <field>         urgency, id, ordinal, due, updated, created, title
      --reverse              flip the whole order, tie-breaks included
      --limit <n>            how many rows to print (default 30, 0 prints none)
      --all                  print every match
      --ids                  print only ids, one per line
      --count                print only how many match

Columns: id, status, type, priority, title, criteria, assignee, due. Empty
cells print a dash. The title is cut at 100 characters, always, before any
column width is computed. Columns 1 to 7 are padded to the widest value
printed; column 8 never is. Two spaces always separate columns.

Exit codes:
  0  listed, even when empty      5  --parent matched several tasks
  2  bad usage                    6  --mine with no identity configured
  3  a filter value does not exist here
  4  --parent does not exist      7  the board could not respond
                                  8  no board here

Examples:
  biso ls
  biso ls -s "In Progress" --mine
  biso ls --type bug --priority high --limit 10
  biso ls --ready --ids
  biso ls --any-status --archived --all
```

---

### 10.5. `biso get`

#### Firma

```
biso get <ref> [--id] [--match] [--section <name>]... [--explain-urgency]
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí | referencia | | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |
| `--section <name>` | | no | `meta`, `desc`, `ac`, `dod`, `plan`, `notes`, `summary`, `comments` | todas | sí | sí | |
| `--explain-urgency` | | no | booleano | falso | no | no | |

`--section` sirve para pedir solo una parte. `biso get TASK-11 --section ac` imprime los criterios con
sus claves y cuesta unas decenas de bytes en vez de la ficha entera, que es lo que hace falta antes de
marcar uno.

#### Comportamiento, caso a caso

La resolución de `<ref>` está en la sección 7 y no se repite. Lo propio de este comando:

| Caso | Qué pasa |
|---|---|
| La referencia resuelve a una tarea | Se imprime, código 0 |
| La referencia es texto y encaja con varias | Error 5, y las candidatas salen **por stdout** en el formato de `biso ls` |
| La referencia es texto y encaja con una | Se imprime, con `note: "CRLF" matched TASK-11` por stderr |
| La tarea está archivada | Se imprime, con `note: TASK-11 is archived` por stderr |
| La tarea no se puede leer | Error 3, según la regla de lectura dirigida de 4.12 |
| `--section` con un nombre inventado | Error 2, con los ocho nombres válidos |
| `--section` de una sección vacía | No imprime esa sección, y si no queda ninguna sección que imprimir, la salida está vacía y el código sigue siendo 0 |

**Sin `--section`, la ficha completa imprime siempre las ocho secciones fijas, vacías incluidas,
marcadas con `(empty)`.** Es solo con `--section` que una sección vacía se omite en vez de imprimirse
vacía; sin la bandera, omitir una sección la confundiría con una que no se ha pedido.

#### Salida

```
TASK-11  Normalize CRLF in the diff
status     In Progress          type       bug
priority   high                 urgency    19.0
assignees  @claude              reporter   @avilches
labels     parser               milestone  -
parent     -                    due        -
created    2026-09-06 09:12     updated    2026-09-06 11:40
depends    -                    blocks     TASK-40
refs       docs/bugs/BUG-02.md
docs       -
files      -
ext        trello.card=5f2a8c1e3b9d4a7f6e0c2b81

## Description
El diff compara byte a byte y marca como distintas dos lineas que solo difieren
en el fin de linea.

## Acceptance Criteria
- [x] #1 El diff ignora el CRLF
- [ ] #3 Hay un test que lo cubre

## Definition of Done
- [ ] #1 Revisado por otra persona

## Implementation Plan
1. Leer el parser.
2. Anadir el caso CRLF.

## Implementation Notes
El parser ya normalizaba LF, faltaba CRLF.

## Final Summary
(empty)

## Comments
@avilches, 2026-09-06 10:02
Esto lo reporto un usuario con un repositorio clonado en Windows.
```

Los encabezados de esta salida son un formato de presentación, no un formato de almacenamiento.

Con `--section ac`, solo el encabezado con el identificador y el título, y la sección pedida:

```
TASK-11  Normalize CRLF in the diff

## Acceptance Criteria
- [x] #1 El diff ignora el CRLF
- [ ] #3 Hay un test que lo cubre
```

Con `--explain-urgency`, al final y por stdout:

```
urgency 19.0
  priority high      6.0 * 1.00 =   6.00
  active             4.0 * 1.00 =   4.00
  blocking           8.0 * 1.00 =   8.00
  blocked           -5.0 * 0.00 =   0.00
  due               12.0 * 0.00 =   0.00
  has criteria       1.0 * 1.00 =   1.00
  age 0 days         0.5 * 0.00 =   0.00
                                 -------
                                   19.00
```

Sobre una tarea en el estado terminal, el desglose se sustituye por una línea:

```
urgency 0.0
  terminal status, urgency is zero by definition
```

#### El esquema JSON

Es el objeto de `task.list` más los campos del cuerpo:

```json
{
  "schemaVersion": 1,
  "kind": "task.get",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "task": {
      "id": "TASK-11",
      "description": "El diff compara byte a byte...",
      "acceptanceCriteria": [ { "key": 1, "text": "El diff ignora el CRLF", "checked": true },
                              { "key": 3, "text": "Hay un test que lo cubre", "checked": false } ],
      "definitionOfDone": [ { "key": 1, "text": "Revisado por otra persona", "checked": false } ],
      "plan": "1. Leer el parser.\n2. Anadir el caso CRLF.",
      "notes": "El parser ya normalizaba LF, faltaba CRLF.",
      "summary": null,
      "comments": [ { "author": "@avilches", "createdAt": "2026-09-06T10:02:11Z", "body": "Esto lo reporto..." } ],
      "blocks": ["TASK-40"],
      "urgencyBreakdown": { "priority": 6.0, "active": 4.0, "blocking": 8.0, "blocked": 0.0,
                            "due": 0.0, "criteria": 1.0, "age": 0.0 }
    }
  }
}
```

Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas. Con varias
coincidencias, `kind` es `task.candidates`, `data.tasks` es la lista y el código es 5.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea impresa | 0 |
| `<ref>` mal formada, `--section` inventada, `--id` con `--match` | 2 |
| La tarea no se puede leer | 3 |
| No existe, o existió y ya no está | 4 |
| Texto con varias coincidencias | 5 |
| El almacén no responde | 7 |
| No hay tablero | 8 |

#### `biso get --help`

```
Usage: biso get <ref> [options]

Show one task. <ref> is an id (TASK-11), a bare number (11) or free text
("CRLF"). Free text that matches several tasks lists them and exits 5; it
never picks one for you.

Free text searches the title, description, plan, notes, final summary, the
text of the criteria and of the definition of done, the body of the comments
and the labels. A match in the title always wins over a match anywhere else.
`biso ls --search` uses this same scope.

Options:
      --id                   force <ref> to be read as an id
      --match                force <ref> to be read as free text
      --section <name>       print only these sections; repeatable or comma
                             separated. One of: meta, desc, ac, dod, plan,
                             notes, summary, comments
      --explain-urgency      show how the urgency number is built
  -h, --help                 show this help

Exit codes:
  0  printed          4  not on this board
  2  bad usage        5  the text matched several tasks
  3  the task could not be read
  7  the board could not respond
  8  no board here

Examples:
  biso get TASK-11
  biso get 11 --section ac
  biso get "CRLF"
  biso get TASK-11 --explain-urgency
```

---

### 10.6. `biso set`

El comando de edición general. Todo lo que hacen los verbos de flujo de 10.7 se puede hacer aquí, con
más palabras.

#### Firma

```
biso set <ref>... [cualquier bandera de campo de la seccion 8]
         [--check <sel>]... [--uncheck <sel>]... [--check-dod <sel>]... [--uncheck-dod <sel>]...
         [--comment <text>]... [--comment-author <@who>] [--id] [--match]
```

#### Parámetros propios

**Todas** las banderas de las secciones 8.2, 8.3, 8.5 y 8.6 valen aquí, con exactamente el mismo
significado que en cualquier otro comando. Lo propio de `set`:

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--check <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--uncheck` |
| `--uncheck <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--check` |
| `--check-dod <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--uncheck-dod` |
| `--uncheck-dod <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--check-dod` |
| `--comment <text>` | | no | texto largo | | sí | no | |
| `--comment-author <@who>` | | no | texto libre | `me` | no | no | requiere `--comment` |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

**El autor de un comentario se llama `--comment-author` en todos los comandos que lo aceptan**, sin
excepción, aunque en `biso comment` el prefijo parezca redundante. Un concepto, un nombre.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Ninguna bandera de cambio | Error 2: `error: nothing to change` con un puntero a `biso get` |
| Varias referencias | El mismo cambio se aplica a todas, con la garantía de todo o nada de 4.10 |
| Varias referencias y un selector que no sea `all` (clave, rango, lista o texto) | Error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |
| Varias referencias y `--check all` | Válido |
| Una de varias referencias no existe | Error 4, y **no se escribe ninguna**, ni siquiera las buenas |
| Un `--set-*` pisa contenido no vacío | Se hace, con el aviso de 4.3 diciendo cuántos bytes ha reemplazado |
| Paso a un estado terminal con criterios sin marcar | Se hace, con aviso |
| Todas las banderas dejan la tarea igual | Código 0, sin escribir, con `note: TASK-11 unchanged` |
| `--comment-author` sin `--comment` | Error 2 |
| `--comment` sin `--comment-author` y sin ninguna identidad configurada (3.1) | Error 2 |
| La tarea no se puede leer | Error 3, y no se escribe nada |

#### Salida

Por defecto, **una línea por tarea afectada** con lo que quien llama no sabía: el estado resultante,
el avance de criterios y la urgencia recalculada. Los tres datos son derivados, y ninguno se puede
conocer sin leer la tarea.

```
TASK-11  In Progress  ac 1/2  urgency 19.0
```

Cuando la tarea tiene definición de hecho, la línea la incluye igual:

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Los avisos van por stderr:

```
warning: --set-plan replaced 412 bytes of existing content
```

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.write",
  "generatedAt": "2026-09-06T11:40:18Z",
  "data": {
    "tasks": [
      { "id": "TASK-11", "status": "In Progress", "acDone": 1, "acTotal": 2,
        "dodDone": 0, "dodTotal": 1, "urgency": 19.0,
        "changed": ["plan", "status"] }
    ],
    "warnings": [ { "code": "overwrite", "field": "plan", "bytes": 412, "task": "TASK-11" } ]
  }
}
```

`kind` es `task.write` para `new`, `set`, `start`, `note`, `comment`, `finish` y `archive`, para que
quien consuma la salida no tenga que distinguir qué verbo la produjo. `changed` dice qué campos han
cambiado de verdad, que no es lo mismo que qué banderas se han pasado. En el lote de `new --from`, las
242 tareas van en `data.tasks` de **un solo sobre**, no en 242 objetos sueltos.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Cambio aplicado, o nada que cambiar | 0 |
| Sin banderas de cambio, banderas incompatibles, selector por clave con varias tareas, solape | 2 |
| Valor fuera de un vocabulario, clave de extensión no declarada, tarea ilegible | 3 |
| Alguna referencia no existe, o un selector de texto no encaja con ningún criterio | 4 |
| Alguna referencia de texto encaja con varias tareas, o un selector con varios criterios | 5 |
| `--dry-run` que no pasa la validación | 9 |
| El almacén falla, o no se obtiene el acceso exclusivo | 7 |
| No hay tablero | 8 |

#### `biso set --help`

```
Usage: biso set <ref>... [options]

Change any field of one or more tasks, all or nothing. Every flag here means
the same in `biso new`, `biso start`, `biso note`, `biso comment`,
`biso finish` and `biso archive`.

The four shapes, and there is no field that breaks them:
  --label X        add one          --set-label X   replace the whole list
  --rm-label X     remove one       --clear-label   empty the list
The same works for --assignee, --ref, --doc, --dep, --file, --ac and --dod.

Prose fields have three, because a block of text has no single item to remove:
  --desc X adds, --set-desc X replaces, --clear-desc empties. The same for
  --plan, --note (whose replacement is --set-notes) and --summary.

External fields have three too: --ext key=value sets that one key, --rm-ext key
drops it, --clear-ext empties the map.

Scalars just take a value: -t/--title, -s/--status, --type, --priority,
--project, -m/--milestone, -p/--parent, --due, --ordinal, --reporter. Each has
a --clear-<field>. An empty string is never a way to clear anything.

Criteria and definition of done:
      --check <sel>          check criteria; sel is all, 3, 1-4, 1,3,7 or the
                             criterion text. The numbers are stable #N keys.
                             With several tasks, sel has to be all
      --uncheck <sel>        the opposite
      --check-dod <sel>      the same for definition-of-done items
      --uncheck-dod <sel>    the opposite
      --rm-ac <sel>          remove criteria by the same selector
      --rm-dod <sel>         remove definition-of-done items

Comments:
      --comment <text>       append a comment; repeatable
      --comment-author <@w>  who wrote it (default: you)

Resolution:
      --id / --match         force <ref> to be an id, or free text

A --set-* over existing content is allowed and warns on stderr with how many
bytes it replaced.

Exit codes:
  0  done                    5  something matched more than one thing
  2  bad usage               7  the board could not be written
  3  unknown value           9  --dry-run did not pass
  4  a task or a criterion was not found
                             8  no board here

Examples:
  biso set TASK-11 --priority high --label parser
  biso set TASK-11 --check 1,3 --note "Both covered by diff_test.rs"
  biso set TASK-11 TASK-12 --milestone "v1.2"
  biso set "CRLF" --set-desc @docs/bugs/BUG-02.md
```

---

### 10.7. Los verbos del ciclo: `start`, `note`, `comment`, `finish`

Los cuatro aceptan **todas** las banderas de campo de la sección 8, igual que `set`. No son un
subconjunto: lo que aportan es un nombre y unos valores por defecto, de modo que el gesto frecuente
cabe en una llamada corta y el gesto raro sigue cabiendo en la misma llamada.

El ciclo entero de una tarea es esto:

```
biso start  TASK-11 --plan "1. Leer el parser. 2. Anadir el caso CRLF."
biso note   TASK-11 "El parser ya normalizaba LF, faltaba CRLF"
biso finish TASK-11 --check all --summary "Normaliza CRLF en el diff, verificado con las pruebas."
```

#### 10.7.1. `biso start`

##### Firma

```
biso start <ref>... [--plan <text>] [-a <@who>]... [-s <v>] [--reopen]
           [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--status <v>` | `-s` | no | vocabulario | `active_status` | no | no | |
| `--reopen` | | no | booleano | falso | no | no | |

`--plan` y `-a/--assignee` son las banderas de campo de la sección 8, con su significado de siempre:
**las dos añaden**. `--plan` añade al plan existente y `--set-plan` lo reemplaza; `-a` añade una
persona y `--set-assignee` reemplaza la lista.

##### Qué hace

Tres cosas en una escritura: pone el estado activo, **asigna la tarea a `me` si no tiene ninguna
persona asignada**, y añade el plan si se ha pasado.

| Caso | Qué pasa |
|---|---|
| La tarea ya está en el estado activo | Se aplica el resto igual, con `note: TASK-11 was already In Progress` |
| La tarea ya está en el estado terminal | Error 6, salvo con `--reopen`, que la devuelve al estado activo |
| La tarea tiene dependencias sin terminar | Se empieza igual, con el aviso correspondiente. **Avisa, no impide** |
| La tarea ya tiene otra persona asignada | No se añade `me`, y sale `note: TASK-11 is assigned to @sara, left as is`. Con `-a` explícito, se añade lo que diga `-a` |
| No hay ninguna identidad configurada (3.1) y no se pasa `-a` | No asigna a nadie, con `note: no identity configured, task left unassigned` |
| La tarea ya tiene plan y se pasa `--plan` | Se añade al final, como toda bandera desnuda |
| Varias referencias | Todo o nada |

##### Salida

```
TASK-11  In Progress  ac 0/2  urgency 19.0
```

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Empezada | 0 |
| Ya estaba terminada y no hay `--reopen` | 6 |
| Referencia mal formada, banderas incompatibles | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Referencia inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso start --help`

```
Usage: biso start <ref>... [options]

Take one or more tasks: move them to the active status, assign them to you if
nobody has them, and record a plan. One call.

Options:
      --plan <text>      add to the implementation plan; repeatable, and takes
                         @file and - like every text option
  -a, --assignee <@who>  add an assignee (--set-assignee replaces the list)
  -s, --status <value>   use another status instead of the active one
      --reopen           allow starting a task that is already finished
      --id / --match     force <ref> to be an id, or free text
  -h, --help             show this help

Every field flag of `biso set --help` works here too.

Unresolved dependencies produce a warning, not an error: you decide.

Exit codes:
  0  started        4  not found        7  the board could not be written
  2  bad usage      5  ambiguous        9  --dry-run did not pass
  3  unknown value  6  already finished, use --reopen
                    8  no board here

Examples:
  biso start TASK-11 --plan "1. Read the parser. 2. Add the CRLF case."
  biso start 11
  biso start TASK-11 TASK-12
```

#### 10.7.2. `biso note`

##### Firma

```
biso note <ref> [<text>...] [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí, salvo con `--note` | texto largo | | sí, como posicional | no | |

Cada texto es un párrafo propio en la sección de notas. Aceptan `@fichero` y `-` como cualquier texto
largo.

##### El posicional que parece un identificador

`biso note` toma **una sola** referencia, mientras que `set`, `start`, `finish` y `archive` toman
varias. Para que esa diferencia no produzca basura en silencio, **un posicional de texto que encaje
con la gramática de identificador de la sección 7.1 es un error 2**:

```
error: "TASK-2" looks like a task id, and `biso note` takes only one task
hint: to note the same thing on several tasks: biso set TASK-1 TASK-2 --note "..."
      to write that text literally:            biso note TASK-1 --note "TASK-2"
```

**`--note`, la bandera de campo de la sección 8, nunca pasa por esa comprobación**, porque no es un
posicional: es la vía para escribir una nota que de verdad diga `TASK-2`. La misma regla vale para
`biso comment`, con `--comment`.

##### Qué hace

Añade uno o más párrafos a las notas de implementación. **Nunca reemplaza.** Para reemplazar está
`biso set <ref> --set-notes`, que este comando no acepta por su nombre desnudo pero sí como bandera de
campo de la sección 8, igual que las demás.

| Caso | Qué pasa |
|---|---|
| Sin ningún texto y sin ninguna bandera de campo | Error 2 |
| Texto vacío | No añade nada y avisa, según 4.6 |
| La tarea no tiene notas todavía | Se crean |
| Varios textos | Un párrafo por texto, en el orden dado |

##### Salida

```
TASK-11  In Progress  ac 1/2  urgency 19.0
```

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Añadida | 0 |
| Sin texto, o un posicional que parece un identificador | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente, o fichero de `@` inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso note --help`

```
Usage: biso note <ref> <text>... [options]

Append one or more paragraphs to the implementation notes of ONE task. It never
replaces anything; `biso set <ref> --set-notes` does that.

Arguments:
  ref            one task: an id, a bare number or free text
  text           one paragraph per argument; @file and - work here too

Options:
      --id / --match   force <ref> to be an id, or free text
  -h, --help           show this help

Every field flag of `biso set --help` works here too. Use `--note <text>` for a
paragraph that is not checked against the id grammar, for when the note
itself looks like an id.

To note the same thing on several tasks, use `biso set A B --note "..."`.

Exit codes:
  0  appended       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso note TASK-11 "The parser already normalized LF, CRLF was missing"
  biso note 11 "First finding" "Second finding"
  biso note TASK-11 @/tmp/benchmark-output.txt
```

#### 10.7.3. `biso comment`

Un comentario y una nota son cosas distintas. Una nota de implementación es el registro técnico de
quien hace el trabajo. Un comentario es una conversación, tiene autor y fecha, y es el canal por el
que entra lo que viene de fuera.

##### Firma

```
biso comment <ref> [<text>...] [--comment-author <@who>]
             [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí, salvo con `--comment` | texto largo | | sí | no | |
| `--comment-author <@who>` | | no | texto libre | `me` | no | no | |

Se aplican las mismas reglas de posicional que en `biso note`, incluida la del texto que parece un
identificador. El autor es texto libre, no se valida contra nada y no interpreta el `@` inicial. **Sin
`--comment-author` y sin ninguna identidad configurada (3.1), es error 2**: `error: --comment-author is
required, no identity is configured`. Lo mismo vale para `--comment` en cualquier otro comando de
escritura.

##### Salida

```
TASK-11  In Progress  ac 1/2  urgency 19.0
```

Y por stderr, `note: comment #2 by @trello:juan`.

##### Códigos de salida

Los mismos de `biso note`, más el error 2 de `--comment-author` sin identidad configurada.

##### `biso comment --help`

```
Usage: biso comment <ref> <text>... [options]

Append a discussion comment to ONE task, with an author and a timestamp. This
is not `biso note`, which records what you did while implementing.

Arguments:
  ref                      one task: an id, a bare number or free text
  text                     one comment per argument; @file and - work too

Options:
      --comment-author <@who>  free text author (default: you). An external
                               system can use its own convention, such as
                               @trello:juan. The @ is never a file reference
      --id / --match           force <ref> to be an id, or free text
  -h, --help                   show this help

Every field flag of `biso set --help` works here too. Use `--comment <text>`
for a comment that is not checked against the id grammar.

Comments are append-only: they are never edited and never deleted.

Exit codes:
  0  appended       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso comment TASK-11 "A user with a Windows clone reported this"
  biso comment 11 "Moved to Doing from the phone" --comment-author @trello:avilches
```

#### 10.7.4. `biso finish`

##### Firma

```
biso finish <ref>... [--summary <text>] [--check <sel>]... [--check-dod <sel>]...
            [--note <text>]... [--file <path>]... [-s <v>] [--strict] [--no-checks]
            [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--status <v>` | `-s` | no | vocabulario | `terminal_status` | no | no | |
| `--strict` | | no | booleano | el valor de `finish_strict` | no | no | `--no-checks` |
| `--no-checks` | | no | booleano | falso | no | no | `--strict` |

`--summary`, `--check`, `--check-dod`, `--note` y `--file` son las banderas de campo de siempre.

##### Qué hace

Marca criterios y elementos de la definición de hecho, escribe la última nota y el resumen, apunta los
ficheros tocados y mueve al estado terminal, todo en una escritura.

| Caso | Qué pasa |
|---|---|
| Quedan criterios sin marcar y no se pasó `--check` | **Se cierra igual**, con el aviso y la lista de los que faltan |
| Quedan elementos de la definición de hecho sin marcar | Igual, con su propio aviso |
| Lo mismo, con `--strict` | Error 6, y no se escribe nada |
| Sin `--summary` | Se cierra igual, con `warning: TASK-11 finished without a final summary` |
| Sin `--summary` y con `--strict` | Error 6 |
| La tarea tiene subtareas sin terminar | Aviso con la lista. Con `--strict`, error 6 |
| La tarea ya estaba terminada | Se aplica el resto sin cambiar el estado, con un `note:` |
| `--no-checks` | Se salta todas las comprobaciones y no emite ninguno de esos avisos |
| Varias referencias | Todo o nada |

Quien quiera la política dura tiene `--strict`, y puede fijarla por defecto con
`biso config set finish_strict true`.

##### Salida

```
TASK-11  Done  ac 2/2  dod 1/1  urgency 0.0
```

La urgencia de una tarea en el estado terminal es cero por definición, según 5.4.

Por stderr, cuando toca:

```
warning: TASK-11 moved to Done with 1 of 2 acceptance criteria unchecked
  #3 Hay un test que lo cubre
```

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Cerrada | 0 |
| `--strict` y falta algo | 6, y no se escribe nada |
| Banderas incompatibles, selector por clave con varias tareas | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Referencia o criterio inexistente | 4 |
| Referencia o criterio ambiguo | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso finish --help`

```
Usage: biso finish <ref>... [options]

Close one or more tasks: check criteria, add the last note, write the final
summary and move to the terminal status. One call.

Options:
      --summary <text>   add to the final summary; repeatable, takes @file and -
      --check <sel>      check criteria: all, 3, 1-4, 1,3,7 or the text. With
                         several tasks the selector has to be `all`
      --check-dod <sel>  the same for the definition of done
      --note <text>      one last implementation note; repeatable
      --file <path>      record a modified file; repeatable
  -s, --status <value>   use another status instead of the terminal one
      --strict           refuse to finish with unchecked criteria, unchecked
                         definition of done, unfinished subtasks or no summary
                         (default: warn and go on; see finish_strict)
      --no-checks        skip every check and every warning
      --id / --match     force <ref> to be an id, or free text
  -h, --help             show this help

Every field flag of `biso set --help` works here too.

Exit codes:
  0  finished       3  unknown value    6  --strict and something is missing
  2  bad usage      4  not found        7  the board could not be written
                    5  ambiguous        9  --dry-run did not pass
                                        8  no board here

Examples:
  biso finish TASK-11 --check all --summary "Normalizes CRLF, tests green"
  biso finish TASK-11 --check "covers CRLF" --note "313 tests green"
  biso finish TASK-11 TASK-12 --check all --summary "Both closed by PR 42"
```

---

### 10.8. `biso archive`

#### Firma

```
biso archive <ref>... [--unarchive] [--id] [--match]
             [cualquier bandera de campo de la seccion 8]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--unarchive` | | no | booleano | falso | no | no | |

Saca la tarea del tablero activo. **La tarea sigue existiendo**, su identificador sigue reservado,
`biso get` la encuentra avisando de que está archivada, y `biso ls --archived` la lista.

| Caso | Qué pasa |
|---|---|
| Ya estaba archivada | Código 0, con un `note:`, sin escribir |
| Otras tareas vivas dependen de ella | Aviso con la lista, se archiva igual |
| `--unarchive` | La devuelve al tablero con el estado que tenía |
| Varias referencias | Todo o nada |

#### `biso delete` no existe, y su ausencia está especificada

Cualquier invocación de `biso delete`, `biso rm` o `biso remove` termina con código 2 y este mensaje
por stderr, en vez de volcar la lista de comandos:

```
error: there is no delete command, on purpose
hint: `biso archive <ref>` takes it off the board and keeps the history
      an archived task still exists: `biso ls --archived` lists them, and the
      id is never reused
```

#### Salida

```
TASK-11  Done  ac 2/2  urgency 0.0  archived
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Archivada o desarchivada | 0 |
| `biso delete`, o banderas incompatibles | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

#### `biso archive --help`

```
Usage: biso archive <ref>... [options]

Take tasks off the board without losing them. An archived task still exists,
`biso get` still finds it, `biso ls --archived` lists it, and its id is never
reused.

Options:
      --unarchive      put them back on the board
      --id / --match   force <ref> to be an id, or free text
  -h, --help           show this help

Every field flag of `biso set --help` works here too.

There is no delete command. Archiving is the way.

Exit codes:
  0  archived       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso archive TASK-11
  biso archive TASK-11 TASK-12 TASK-13
  biso archive TASK-11 --unarchive
```

---

### 10.9. `biso export`

#### Firma

```
biso export [-o <file|->] [--no-archived] [cualquier filtro de biso ls, salvo --archived y --only-archived]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--out <file>` | `-o` | no | ruta o `-` | `-`, es decir stdout | no | no | |
| `--no-archived` | | no | booleano | falso | no | no | |
| filtros de `ls` | | no | | | | | `--sort`, `--limit`, `--all`, `--ids`, `--count`, `--archived`, `--only-archived` |

**`biso export` sin filtros exporta el tablero entero**: todos los estados, el terminal incluido, y
todas las tareas, las archivadas incluidas. **No hereda ni el límite por defecto de `biso ls` ni su
exclusión del estado terminal**, y no existe aquí ninguna bandera `--all`. Los filtros de `biso ls` se
aceptan para acotar a propósito, y las banderas de forma de `ls` no, porque un volcado no tiene forma
que elegir. **Tampoco se aceptan `--archived` ni `--only-archived`**, porque las archivadas ya salen
por defecto: la única bandera de `export` sobre el archivo es `--no-archived`.

#### La garantía de simetría

La salida es NDJSON, una tarea por línea, con **exactamente** las claves que acepta `biso new --from`,
en la forma de objeto que esa sección define para los criterios, la definición de hecho y los
comentarios, e incluyendo `id`, `createdAt`, `updatedAt`, `archived` y las claves estables de cada
criterio.

**Los únicos campos que no salen son los derivados de la sección 5.**

La garantía que la suite de pruebas comprueba:

```bash
biso export -o copia.ndjson
biso -C /tmp init nuevo --at /tmp/tablero-nuevo \
     --statuses "Ideas,To Do,In Progress,Blocked,Done" \
     --types "idea,memory,task,bug,docs" --extensions trello.card
biso --board /tmp/tablero-nuevo new --from copia.ndjson
# los dos tableros son identicos en todos los campos no derivados, incluidos
# los identificadores, las fechas, las claves de los criterios y sus marcas
```

El segundo `init` declara los mismos estados, tipos y extensiones que el tablero de origen. El
tablero de destino tiene que declarar el mismo vocabulario que el de origen para que la importación
pase la validación: un tablero con otro vocabulario hace fallar el lote entero con código 9, con el
detalle de qué valores faltan. El `init` se ejecuta con `-C /tmp`, fuera del proyecto de origen.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Exportado, aunque sean cero tareas | 0 |
| Alguna tarea se ha saltado por ilegible | 6 |
| Banderas de forma de `ls`, `--archived`, `--only-archived`, `--json`, o incompatibles | 2 |
| Un valor de filtro no existe en el tablero | 3 |
| No se puede escribir el fichero de salida | 7 |
| No hay tablero | 8 |

#### `biso export --help`

```
Usage: biso export [options]

Write the board as NDJSON, one task per line, in exactly the shape that
`biso new --from` reads back. Round-tripping every non-derived field is a
tested guarantee: ids, dates, criterion keys and their checkmarks included.

With no filters it exports everything, the finished and the archived included.
It never inherits the default limit or the default status filter of `biso ls`.
If a task cannot be read, the rest is still written and the exit code is 6,
not 0: this is the one command whose purpose is to lose nothing.

Options:
  -o, --out <file|->   where to write (default: stdout)
      --no-archived    leave the archived tasks out
  -h, --help           show this help

Every filter of `biso ls` works here except --archived and --only-archived,
which do not apply because archived tasks are already included by default.
Its shaping flags (--sort, --limit, --all, --ids, --count) do not apply either.
--json is rejected with code 2: this output is already one JSON object per
line, while --json means the single envelope every other command prints.

Derived fields are never written: urgency, acDone, acTotal, dodDone, dodTotal,
commentCount, blocks, ready, blocked.

Exit codes:
  0  exported       3  a filter value does not exist here
  2  bad usage      6  some task was skipped, unreadable
  7  cannot write there                8  no board here

Examples:
  biso export -o backup.ndjson
  biso export -s Done --no-archived -o done.ndjson
```

---

### 10.10. `biso config`

#### Firma

```
biso config get <key>
biso config set <key> <value>
biso config list
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<key>` | | sí en `get` y `set` | una clave de la tabla de abajo | | no | no | |
| `<value>` | | sí en `set` | según la clave | | no | sí en las claves de lista | |

`--json` solo se acepta en `config list`. En `get` la salida ya es un solo valor y en `set` no hay
salida por stdout, así que en los dos es un error de uso con código 2.

#### Las claves

| Clave | Tipo | Por defecto |
|---|---|---|
| `project_name` | texto | el nombre del proyecto |
| `statuses` | lista, mínimo dos | `To Do, In Progress, Done` |
| `default_status` | uno de `statuses` | el primero, al crear el tablero |
| `active_status` | uno de `statuses` | el penúltimo, al crear el tablero |
| `terminal_status` | uno de `statuses` | el último, al crear el tablero |
| `types` | lista | `task, bug, docs` |
| `priorities` | lista | `high, medium, low` |
| `projects` | lista | vacía |
| `labels` | lista | vacía |
| `assignees` | lista | vacía |
| `extensions` | lista | vacía |
| `task_prefix` | texto de solo letras | `TASK` |
| `me` | texto de persona | `BISO_ME` si está definida |
| `default_assignee` | texto de persona | vacío |
| `default_limit` | entero >= 0 | 30 |
| `finish_strict` | booleano | falso |
| `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`, `urgency.criteria`, `urgency.age` | decimal | ver 5.4 para el término de cada uno y su valor por defecto |

**Los tres estados especiales son valores explícitos, no posiciones.** Se escriben al crear el tablero
y **cambiar `statuses` no los mueve nunca**. Esta es la diferencia que evita que añadir una columna al
final cambie en silencio a dónde va `biso finish`.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Clave inexistente | Error 4, con las tres claves más parecidas |
| Valor del tipo equivocado, por ejemplo `finish_strict maybe` | Error 3, diciendo qué tipo esperaba |
| `default_status` a un valor que no está en `statuses` | Error 3 |
| Quitar de `statuses` un estado que alguna tarea usa | Error 6, con cuántas tareas lo usan y en cuáles |
| Quitar de `statuses` un estado que es `default_status`, `active_status` o `terminal_status` | Error 6, diciendo cuál de los tres y que hay que cambiarlo antes |
| Quitar de `extensions` una clave que alguna tarea usa | Error 6, con la lista de tareas |
| Quitar de `types` o `priorities` un valor en uso | Error 6, igual |
| `get` de una clave de lista | Los valores separados por comas, en una línea |
| `set` correcto | Sin salida por stdout, con `note:` por stderr diciendo el valor nuevo |

Ningún cambio de configuración toca ninguna tarea, nunca.

#### Salida

```
$ biso config get statuses
Ideas,To Do,In Progress,Blocked,Done

$ biso config list
project_name = Kex
statuses = Ideas,To Do,In Progress,Blocked,Done
default_status = Ideas
active_status = In Progress
terminal_status = Done
types = idea,memory,task,bug,docs
priorities = high,medium,low
extensions = trello.card
task_prefix = TASK
me = @claude
default_limit = 30
finish_strict = false
```

#### El esquema JSON

Solo `config list` acepta `--json`:

```json
{
  "schemaVersion": 1,
  "kind": "config",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "config": {
      "project_name": "Kex",
      "statuses": ["Ideas", "To Do", "In Progress", "Blocked", "Done"],
      "default_status": "Ideas",
      "active_status": "In Progress",
      "terminal_status": "Done",
      "types": ["idea", "memory", "task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "projects": [],
      "labels": [],
      "assignees": [],
      "extensions": ["trello.card"],
      "task_prefix": "TASK",
      "me": "@claude",
      "default_assignee": null,
      "default_limit": 30,
      "finish_strict": false,
      "urgency": { "priority": 6.0, "active": 4.0, "blocking": 8.0, "blocked": -5.0,
                   "due": 12.0, "criteria": 1.0, "age": 0.5 }
    }
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Hecho | 0 |
| Sintaxis, o `--json` fuera de `list` | 2 |
| Valor de tipo o de dominio incorrecto | 3 |
| Clave inexistente | 4 |
| El cambio dejaría el tablero inconsistente | 6 |
| No se puede escribir la configuración | 7 |
| No hay tablero | 8 |

#### `biso config --help`

```
Usage: biso config get <key>
       biso config set <key> <value>
       biso config list [--json]

Read and change the board configuration. List values are comma-separated.
No configuration change ever touches a task.

Keys:
  project_name       board name
  statuses           the columns, in order
  default_status     status of a new task           (one of statuses)
  active_status      what `biso start` sets         (one of statuses)
  terminal_status    what `biso finish` sets        (one of statuses)
  types              configured task types
  priorities         configured priorities
  projects           configured projects
  labels             labels that filters accept on top of the ones in use
  assignees          assignees that filters accept on top of the ones in use
  extensions         declared external field keys, such as trello.card
  task_prefix        id prefix, letters only (default TASK)
  me                 who you are, for --mine and for comment authorship
  default_assignee   assignee of a new task
  default_limit      how many rows `biso ls` prints (default 30)
  finish_strict      make `biso finish` refuse an incomplete task
  urgency.priority, urgency.active, urgency.blocking, urgency.blocked,
  urgency.due, urgency.criteria, urgency.age
                     the seven urgency coefficients; see `biso get --explain-urgency`

The three special statuses are stored as explicit values. Changing `statuses`
never moves them; if a change would remove one of them, it fails and says so.

Removing any configured value that a task still uses is refused, never applied
silently.

Options:
      --json         machine-readable output, `list` only
  -h, --help         show this help

Exit codes:
  0  done            4  no such key
  2  bad usage       6  the change would leave the board inconsistent
  3  bad value       7  the configuration could not be written
                     8  no board here

Examples:
  biso config get active_status
  biso config set statuses "Ideas,To Do,In Progress,Blocked,Done"
  biso config set finish_strict true
  biso config list --json
```

---

### 10.11. `biso doctor`

#### Firma

```
biso doctor [--fix]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--fix` | | no | booleano | falso | no | no | ninguno |

**`biso doctor` sin `--fix` es de solo lectura**, y `--print` y `--dry-run` de la sección 3 son error 2
igual que en cualquier otro comando de lectura. **Con `--fix` es un comando de escritura**: ahí
`--dry-run` reporta qué se repararía sin reparar nada, y `--print` no añade nada, porque `doctor` no
imprime fichas de tareas.

#### Qué comprueba

| Comprobación | Reparable con `--fix` |
|---|---|
| Identificadores duplicados | no, hay que decidir a mano |
| Tareas que no se pueden leer | no |
| Claves de extensión no declaradas | no |
| Estados, tipos, prioridades o proyectos que ya no están configurados | no |
| `default_status`, `active_status` o `terminal_status` que no están en `statuses` | no |
| Dependencias que apuntan a tareas inexistentes | no |
| Ciclos de dependencias | no |
| Ciclos de tarea padre | no |
| Claves de criterio repetidas dentro de una tarea | no |
| El identificador más alto que el tablero recuerda haber asignado (4.11) es menor que el identificador más alto de una tarea existente | sí |
| Huecos en la numeración | no son un problema, no se reportan |

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Tablero limpio | `no problems found` por stdout, código 0 |
| Solo problemas reparables, con `--fix` | Se reparan y se reporta cada uno, código 0 |
| Quedan problemas sin reparar | Código 6, aunque se haya reparado algo |
| Una tarea ilegible | Se reporta y se sigue con las demás. **Nunca aborta** |
| `--fix` sin poder escribir | Código 7 |
| `--fix --dry-run` | Reporta qué se repararía, sin reparar nada, código 0 |

#### Salida

```
2 problems found
  TASK-40  dependency TASK-99 does not exist
  TASK-52  unknown extension key: jira.key (declared: trello.card, github.issue)
1 problem fixed
  the highest recorded id was TASK-40 and tasks go up to TASK-52; recorded TASK-52
```

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "doctor",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "problems": [
      { "task": "TASK-40", "code": "dependency_not_found", "message": "dependency TASK-99 does not exist" },
      { "task": "TASK-52", "code": "unknown_extension_key", "message": "unknown extension key: jira.key (declared: trello.card, github.issue)" }
    ],
    "fixed": [
      { "code": "highest_id_behind", "message": "the highest recorded id was TASK-40 and tasks go up to TASK-52; recorded TASK-52" }
    ]
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Nada mal, o todo lo encontrado se ha reparado | 0 |
| Quedan problemas | 6 |
| Sintaxis | 2 |
| No se puede escribir al reparar | 7 |
| No hay tablero | 8 |

#### `biso doctor --help`

```
Usage: biso doctor [options]

Check the board for duplicate ids, unreadable tasks, undeclared extension keys,
values that are no longer configured, broken or circular dependencies, repeated
criterion keys, and a recorded highest id that has fallen behind.

Options:
      --fix      repair what can be repaired without a decision
  -h, --help     show this help

Without --fix this is a read-only command: --print and --dry-run are bad usage
here, same as in any other read-only command. With --fix, --dry-run reports
what would be fixed without fixing it.

An unreadable task is reported and skipped, never a reason to stop.
Gaps in the id sequence are normal and are not reported.

Exit codes:
  0  nothing wrong, or everything found was fixed
  2  bad usage
  6  problems remain
  7  cannot write while fixing
  8  no board here

Examples:
  biso doctor
  biso doctor --fix
```

---

### 10.12. `biso board`

#### Firma

```
biso board [--port <n>] [--no-open]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--port <n>` | | no | entero entre 1024 y 65535 | 6420 | no | no | ninguno |
| `--no-open` | | no | booleano | falso | no | no | ninguno |

La interfaz interactiva, y **el único comando del programa que abre una interfaz**. Ningún otro puede
abrirla, ni la abre nadie por su cuenta.

| Caso | Qué pasa |
|---|---|
| No hay terminal | Error 7: `error: biso board needs a terminal; every other command works without one` |
| El puerto está ocupado | Error 7, diciendo el puerto |
| Puerto fuera de rango | Error 2 |
| `--no-open` | Arranca y solo imprime la dirección |
| El tablero cambia mientras está abierto | La interfaz recarga. Nunca muestra una versión en caché de una tarea que otro proceso ha cambiado |

#### Salida

```
Board at http://127.0.0.1:6420. Ctrl-C to stop.
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| El servidor se ha parado limpiamente | 0 |
| Puerto fuera de rango, o banderas incompatibles | 2 |
| No hay terminal, o el puerto está ocupado | 7 |
| No hay tablero | 8 |

#### `biso board --help`

```
Usage: biso board [options]

Open the interactive board in a browser. This is the only command that opens
an interface: every other one prints text and exits, with or without a
terminal.

Options:
      --port <n>   port to listen on, 1024 to 65535 (default 6420)
      --no-open    print the address and do not open a browser
  -h, --help       show this help

Exit codes:
  0  stopped cleanly
  2  bad usage
  7  no terminal, or the port is taken
  8  no board here

Examples:
  biso board
  biso board --port 7000 --no-open
```

---

### 10.13. `biso help`

#### Firma

```
biso help [<command> | all]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<command>` | | no | nombre de comando, o `all` | ninguno | no | no | |

| Caso | Qué pasa |
|---|---|
| Sin argumento | Imprime la ayuda de primer nivel de la sección 11, igual que `biso --help` |
| Con un nombre de comando | Imprime la ayuda de ese comando, igual que `biso <cmd> --help` |
| Con `all` | Imprime la ayuda de primer nivel más la lista de los ocho comandos de administración, cada uno con su línea |
| Con un nombre que no existe | Error 4, con los tres nombres más parecidos |

`biso help` funciona **sin tablero**.

#### Salida de `biso help all`

Es la de la sección 11, seguida de:

```
Administration:
  init               create a board for this project
  where              say which board is in use and why
  archive <ref>      take a task off the board (there is no delete)
  export             dump the board as NDJSON that `biso new --from` reads back
  config             read and change the board configuration
  doctor             check the board, and repair what can be repaired
  board              open the interactive board
  help [cmd|all]     this
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Ayuda impresa | 0 |
| El comando no existe | 4 |

#### `biso help --help`

```
Usage: biso help [command|all]

Print the top-level help, or the help of one command, or the top-level help
plus the eight administrative commands with `all`. Works without a board.

Arguments:
  command        a command name, or `all`

Exit codes:
  0  help printed
  4  no such command

Examples:
  biso help
  biso help finish
  biso help all
```

---

## 11. La ayuda de primer nivel

`biso --help` y `biso help` imprimen esto, y solo esto:

```
biso 1.0.0 - the task board of this project.

Usage: biso [global options] <command> [options]

Start here:
  prime              everything you need to work on this board, in one message

Daily work:
  ls                 list tasks, most urgent first
  get <ref>          show one task
  new "TITLE"        create a task and print its id
  set <ref>...       change any field
  start <ref>...     take a task
  note <ref> TEXT    append an implementation note
  comment <ref> TEXT append a discussion comment
  finish <ref>...    close a task

Global options:
  -C, --cwd <path>   resolve the board from there, instead of cd-ing
      --board <name> use this board directly
      --json         machine-readable output
  -q, --quiet        print only ids
      --print        print the whole record after writing
      --color <when> auto (default), always, never
      --dry-run      validate, write nothing (writing commands only)
  -V, --version      print the version
  -h, --help         this, or the help of a command

More: `biso <command> --help`, and `biso help all` for the administrative
commands (init, where, archive, export, config, doctor, board, help).
```

Son treinta líneas, y no incluyen los ocho comandos de administración.

---

## 12. El contrato JSON

### 12.1. El sobre

Toda salida con `--json` es **un solo objeto JSON**, con o sin sangrado, indistintamente, y siempre
con esta forma:

```json
{ "schemaVersion": 1, "kind": "<tipo>", "generatedAt": "<ISO 8601 UTC>", "data": { } }
```

| `kind` | Lo produce | `data` contiene |
|---|---|---|
| `prime` | `prime` | Sección 9.9 |
| `where` | `where` | `board`, `source`, `me`, `counts`. Ejemplo en 10.2 |
| `init` | `init` | `board`, `pointerCreated`. Ejemplo en 10.1 |
| `task.list` | `ls` | `tasks`, `shown`, `matched`, `hidden`, `truncated`, `skipped`, `sort`, `filters` |
| `task.get` | `get` | `task` |
| `task.candidates` | `get` con varias coincidencias | `tasks` |
| `task.write` | `new`, `set`, `start`, `note`, `comment`, `finish`, `archive` | `tasks`, `warnings` |
| `config` | `config list` | `config`. Ejemplo en 10.10 |
| `doctor` | `doctor` | `problems`, `fixed`. Ejemplo en 10.11 |
| `error` | cualquier fallo | Ver 12.2 |

Un lote de doscientas cuarenta y dos tareas es **un solo sobre** con doscientas cuarenta y dos
entradas en `data.tasks`, nunca doscientos cuarenta y dos objetos sueltos. La única salida del
programa que es una secuencia de objetos, uno por línea, es `biso export`, que no lleva sobre porque
su formato es NDJSON por definición.

### 12.2. Los errores en JSON

Con `--json`, un error sale **por stderr**, como un objeto con esta forma, y el código de salida del
proceso es el de la tabla de la sección 2:

```json
{
  "schemaVersion": 1,
  "kind": "error",
  "generatedAt": "2026-09-06T09:12:04Z",
  "error": {
    "exitCode": 3,
    "code": "unknown_status",
    "message": "unknown status: \"Pending\"",
    "field": "status",
    "given": "Pending",
    "valid": ["Ideas", "To Do", "In Progress", "Blocked", "Done"]
  }
}
```

Cuando un solo comando produce varios fallos, como un lote inválido, `error.details` es una lista de
objetos con la misma forma, uno por fallo, y `error.code` es `batch_invalid`.

### 12.3. Los identificadores de error

Un `code` estable es lo que permite ramificar sin analizar prosa. Esta es la lista de la versión 1.0,
agrupada por el código de salida con el que sale cada uno:

| Código de salida | `code` |
|---:|---|
| 2 | `incompatible_flags`, `duplicate_scalar_flag`, `unexpected_argument`, `missing_value`, `unknown_flag`, `unknown_command`, `missing_title`, `nothing_to_change`, `malformed_id`, `id_like_positional`, `inverted_range`, `key_selector_with_many_tasks`, `criterion_selector_overlap`, `two_stdin`, `read_only_flag`, `invalid_date`, `invalid_number`, `invalid_prefix`, `dependency_cycle`, `parent_cycle`, `self_dependency`, `board_exists`, `delete_not_supported` |
| 3 | `unknown_status`, `unknown_type`, `unknown_priority`, `unknown_project`, `unknown_label`, `unknown_assignee`, `unknown_extension_key`, `unknown_section`, `unknown_sort_field`, `ambiguous_vocabulary`, `empty_scalar_value`, `bad_config_value`, `undecodable_task`, `invalid_encoding` |
| 4 | `not_found`, `never_allocated`, `unknown_config_key`, `criterion_not_found`, `file_not_found` |
| 5 | `ambiguous_reference`, `criterion_ambiguous` |
| 6 | `already_finished`, `precondition_failed`, `board_inconsistent`, `doctor_problems` |
| 7 | `busy`, `io_error`, `file_unreadable`, `no_terminal`, `port_in_use` |
| 8 | `no_board` |
| 9 | `batch_invalid`, `dry_run_failed` |
| 1 | `internal` |

**La lista es ampliable y las entradas son permanentes.** Una versión posterior puede añadir un `code`
nuevo, pero ninguno de los de arriba cambiará de significado, cambiará de código de salida ni
desaparecerá. Quien ramifique sobre un `code` desconocido debe tratarlo por su código de salida, que
sí está cerrado.

### 12.4. Números, fechas y ausencias

- Las fechas son ISO 8601 en UTC terminadas en `Z`, con precisión de segundo. Nunca hora local, nunca
  sin zona. `due` es la excepción, porque es un día y no un instante, y viaja como `YYYY-MM-DD`.
- `urgency` es un decimal con un solo dígito tras el punto.
- Un campo sin valor es `null`, nunca la cadena vacía ni la ausencia de la clave. **Todas las claves
  documentadas aparecen siempre**, para que nadie tenga que distinguir entre "no está" y "no tiene
  valor".
- Una lista vacía es `[]` y un mapa vacío es `{}`, nunca `null`.

---

## 13. El contrato de estabilidad

Lo que se promete mientras la versión mayor sea `1`.

**No cambia nunca:**

- Los códigos de salida de la sección 2 y su significado.
- Los identificadores `code` de la sección 12.3, con la regla de ampliación que allí se dice.
- El nombre y el significado de cada comando y de cada bandera. **Una bandera nunca cambia de
  semántica**, y en particular ninguna que hoy añade pasará a reemplazar. Si hiciera falta el
  comportamiento contrario, se añade una bandera nueva con otro nombre.
- Las claves de `data` en cada `kind` de JSON. Se pueden añadir claves; las que hay no se quitan ni
  cambian de tipo.
- El algoritmo de coincidencia de la sección 6.1, idéntico al leer y al escribir.
- La simetría entre `biso export` y `biso new --from` sobre todos los campos no derivados, que es una
  prueba de la suite y no una intención.
- La estabilidad de las claves de los criterios: una clave asignada no se reasigna nunca.
- El tope de tamaño del mensaje de `biso prime`.

**Puede cambiar entre versiones menores, y por eso no hay que analizarlo:**

- El texto exacto de los mensajes de error y de los avisos. Lo estable es el `code`, no la prosa.
- La disposición de las columnas de `biso ls` y de la ficha de `biso get`, y para eso está `--json`.
- El texto de `biso prime`, dentro de su tope, que es donde se espera que la herramienta más aprenda
  con el tiempo.
- Los coeficientes por defecto de la urgencia. La estructura de la fórmula, no.
- Los valores por defecto de la configuración, salvo los que este documento fija dentro de un comando.

**Cómo se anuncia una retirada.** Nada se quita sin un ciclo completo de aviso: primero la
funcionalidad emite `warning: <x> is deprecated and will be removed in 2.0` durante al menos una
versión menor, y solo entonces desaparece. Una funcionalidad nunca desaparece en silencio entre dos
versiones.

**Migración.** Si cambiara la forma en que los datos se guardan, la herramienta migra sola al
detectarlo, y en cualquier caso el volcado de una versión se puede importar en la siguiente, porque
el formato de `export` es el de `new --from` y los dos están en este contrato.

---

## 14. Lo que se deja fuera a propósito

Nombrar lo que no está evita que alguien lo dé por olvidado.

- **No hay `biso delete`.** Está especificado que no existe y qué contesta si se intenta (10.8).
- **No hay entidades de hito, documento ni decisión.** El hito es un campo de texto libre de la tarea,
  no una entidad con ciclo de vida propio, y por eso `-m/--milestone` no valida contra ninguna lista.
  La documentación se apunta con `--doc`, que es una lista de textos.
- **No hay contextos de sesión**, es decir, filtros por defecto guardados que cambien lo que devuelve
  una consulta sin que se vea en la línea de comandos.
- **No hay recurrencia, ni seguimiento de tiempo, ni subtareas con numeración propia.** Una subtarea
  es una tarea normal con `--parent`, y el mensaje de error de un identificador como `TASK-1.1` lo
  dice.
- **No hay servidor de integración ni protocolo de herramientas.** La interfaz de la versión 1.0 es
  esta línea de comandos y su salida JSON.
- **No hay nada sobre control de versiones.** Ni commits automáticos, ni ramas, ni la posibilidad de
  ver una tarea que existe en otra versión del proyecto. Todo eso depende de cómo se guarden los
  datos, y este documento no lo decide. Cuando esa decisión se tome, hará falta especificar aparte qué
  ocurre con las tareas que existen en una versión del proyecto y no en otra: hoy, la única promesa es
  la de los tres mensajes distintos de la sección 7.3.
- **No hay sincronización con ningún sistema externo.**

---

## 15. Por dónde empezar a implementar

En el orden en que cada pieza paga lo que cuesta:

1. **El modelo de datos lógico** de la sección 5, con las claves estables de los criterios y el
   rechazo explícito de lo desconocido.
2. **El algoritmo de coincidencia** de la sección 6.1, que es una función pura de veinte líneas y de
   la que dependen todos los comandos.
3. **`new`, `ls`, `get` y `set`**, que son el trabajo diario.
4. **Los cuatro verbos de ciclo** de 10.7, que son azúcar sobre `set` y se escriben encima.
5. **`prime`**, que es lo que hace que todo lo anterior se use bien sin leer nada más.
6. **El lote de `new --from` y `export`**, con la prueba de simetría, que es lo que convierte una
   migración en una sola operación.
7. **El resto**: `init`, `where`, `archive`, `config`, `doctor`, `board` y `help`.

Las garantías de la sección 4.10 no son un paso de esta lista: hay que respetarlas desde el primer
comando que escriba.
