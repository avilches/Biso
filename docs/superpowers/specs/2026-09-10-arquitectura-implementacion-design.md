# La arquitectura del código de `biso`

## 1. Qué resuelve este documento, y qué no

`docs/spec/` fija el comportamiento observable de `biso`: qué hace cada comando, qué imprime,
qué código de salida devuelve, qué esquema JSON expone. No dice nada de cómo se organiza el código
que lo implementa, porque no le corresponde: eso es una decisión de implementación, no de contrato.

Este documento cubre exactamente ese hueco: los paquetes Go que va a tener el programa, la
responsabilidad de cada uno, sus dependencias entre sí, las librerías externas que hacen falta y
la estrategia de pruebas. No repite ni reinterpreta ningún comportamiento de `docs/spec/`; donde
haya duda sobre qué hace un comando, la respuesta sigue estando allí, no aquí.

**Nota de mantenimiento (2026-09-19).** Este documento se escribió citando `docs/SPEC.md` y
`docs/DECISIONES.md` por número de sección, dos ficheros que ya no existen: la especificación se
partió en `docs/spec/` (TASK-36/TASK-58 y siguientes). Las referencias de abajo ya apuntan a los
ficheros y anclas actuales de `docs/spec/` y `docs/decisiones/`.

**El motivo para escribirlo antes de tocar código** es que va a existir una interfaz gráfica más
adelante, y quiere reutilizar la misma lógica que usa la línea de comandos. Si esa frontera no se
traza ahora, el primer comando que se escriba la fija por accidente, mezclando el analizador de
argumentos con la lógica de negocio, y separarlos después cuesta mucho más que diseñarlos separados
desde el principio.

## 2. El principio: separar "qué hace un comando" de "cómo se invoca"

La idea central es que la lógica de cada comando (`new`, `ls`, `set`, `finish`...) no debe saber si
quien la invocó fue una terminal, un test, o una futura interfaz gráfica. Concretamente: recibe
parámetros ya tipados y validados en su forma, devuelve un resultado tipado, y no imprime nada ni
sabe qué es `--json`.

Todo lo que sabe de texto de ayuda, de `argv`, de códigos de salida de proceso y del contrato JSON
de `docs/spec/contrato-json.md` vive en una sola capa delgada, porque toda la decisión de negocio ya
la tomó la capa de debajo. Esa capa delgada es la única que cambiaría si algún día se añade una
interfaz gráfica; la lógica de los comandos no cambiaría nada.

**Sobre si la interfaz será un binario Go en este mismo módulo o un proceso aparte que hable el
`--json` ya especificado en `docs/spec/contrato-json.md`: queda sin decidir a propósito**, porque no hace falta
decidirlo para trazar esta frontera. Con la lógica de comandos aislada en un paquete `internal/`
con una API en Go limpia, las dos opciones siguen abiertas: si la interfaz acaba siendo un binario
del mismo módulo, importa ese paquete directamente; si acaba siendo un proceso o un módulo aparte,
promoverlo a un paquete público es un cambio de nombre y de ruta de import, no un rediseño. No hay
razón para pagar ese coste antes de necesitarlo.

## 3. El árbol de paquetes

```
cmd/biso/main.go        arranca, llama a cli.Run(argv), hace os.Exit(code)
internal/cli            argv -> Params tipados por comando; Result -> texto o JSON
internal/ops            la lógica de cada comando: Params -> Result, sin I/O de terminal
internal/board          abre un tablero: resuelve su ruta, carga su config, envuelve el Store
internal/store          SQLite: abrir/crear, WAL, transacciones, migraciones
internal/model          Task, Criterion, Comment, Question... tipos puros, sin I/O
internal/match          el algoritmo de coincidencia de docs/spec/vocabularios.md, función pura
internal/vcs            ejecuta git (o el sistema custom) para biso snapshot
```

La regla de dependencias es de arriba abajo y nunca al revés. `internal/model` es la base y no
depende de ningún otro paquete propio. `internal/match` y `internal/vcs` dependen solo de `model`
(el tipo `Error` de la sección 4 de este documento es de allí, y tanto una coincidencia fallida como
una orden de `git` que falla necesitan construir uno), pero no entre sí ni de nada que esté por
encima. `internal/store` también depende solo de `model`, por el mismo motivo con los errores de
disco. `internal/board` depende de `store` y de `model`. `internal/ops` depende de `board`, `model`,
`match` y `vcs`. `internal/cli` depende de `ops` y de `model`. Ningún paquete de abajo importa nunca
uno de arriba; si algún día pareciera necesario, es señal de que algo está en la capa equivocada.

### 3.1. `internal/model`

Tipos y validación, sin tocar disco ni red: `Task`, `Criterion` (con su clave estable,
`docs/spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables`), `Comment`,
`Question`, `Urgency`, y el tipo de error transversal de la sección 4 de este documento. Ningún
tipo de aquí sabe qué es SQLite ni qué es JSON; sus métodos son de validación pura (por ejemplo, si
un campo escalar admite cadena vacía o no, `docs/spec/valores-de-entrada.md#el-valor-vacío`).

### 3.2. `internal/match`

El algoritmo de coincidencia de vocabulario (`docs/spec/vocabularios.md#el-algoritmo-de-coincidencia`)
y sus reglas relacionadas (`#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos`, la equivalencia en
los dos sentidos; `#qué-valida-cada-filtro-y-contra-qué`, qué valida cada filtro). Es un paquete propio,
aunque pequeño, porque la propia especificación insiste en que es una función pura de veinte líneas
de la que dependen `ls`, `set`, `new` y todos los vocabularios cerrados: aislarla la hace trivial de
probar con una tabla de casos, y evita que cada comando reimplemente su propia variante de la
regla.

### 3.3. `internal/store`

La única capa que importa `modernc.org/sqlite` y `database/sql`. Abre o crea la base de datos,
activa el modo WAL, gestiona las migraciones del esquema, y expone una única función
`WithTx(func(*sql.Tx) error) error` que es el único punto por el que pasa cualquier escritura: de
ahí salen las seis garantías de `docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables`.
No sabe qué es una `Task`, solo filas
y columnas; traducir entre unas y otras es trabajo de `internal/board`.

`Open` recibe dos cosas, `Open(id, path string)`. El `id` es el del tablero
(`docs/spec/resolucion-del-tablero.md`), que quien abre el almacén ya ha leído del puntero, y el
almacén no lo usa para nada más que nombrar el tablero en el mensaje del error de base de datos
ilegible, que `docs/spec/garantias.md` fija palabra por palabra. Sigue sin saber qué es una `Task`:
un identificador de tablero es una cadena para un mensaje, no el modelo de datos.

Las decisiones de este paquete que se tomaron al implementarlo (TASK-4) y que no se deducen de la
especificación, porque son del controlador y no del comportamiento observable:

- **La ruta del fichero se escapa antes de meterla en el DSN, y se escribe como la espera el
  analizador de URIs de SQLite.** El controlador abre la base de datos con `SQLITE_OPEN_URI` y con
  el prefijo `file:`, así que SQLite lee la ruta como un URI: un `#` en el nombre de un directorio
  corta ahí el nombre del fichero y abre otro distinto sin avisar, y un `?` empieza la lista de
  parámetros. Por eso `uriPath` sustituye `%`, `?` y `#` por su forma porcentual. Y como el
  controlador se eligió precisamente porque compila cruzado a Windows
  (`docs/decisiones/lenguaje-y-rendimiento.md`), esa misma función escribe las rutas de esa
  plataforma: las barras invertidas pasan a barras normales, una letra de unidad recibe delante la
  barra que la capa Windows de SQLite vuelve a descartar (`/C:/...`), y una ruta UNC recibe las
  cinco barras iniciales que SQLite documenta para ellas, para que el nombre del servidor no se lea
  como la autoridad del URI. Está probado con un tablero dentro de un directorio llamado `board #1`
  y con una tabla de casos sobre `uriPath`, que no necesita una máquina Windows.
- **Una sola conexión por proceso** (`SetMaxOpenConns(1)`). En modo WAL solo hay un escritor a la
  vez, así que una segunda conexión del mismo proceso únicamente podría esperar a la primera y
  agotar su tiempo contra sí misma. La contrapartida es que el código que corre dentro de `WithTx`
  usa siempre la transacción que recibe y nunca el manejador del `Store`, que esa transacción tiene
  tomado.
- **Esa contrapartida es un error y no una espera sin fin.** Usar el manejador del `Store` mientras
  una transacción está viva esperaría a una conexión que solo esa transacción puede devolver: para
  siempre, sin error y sin tiempo de espera, justo lo contrario de lo que promete la garantía 5
  (fallar en cinco segundos con código 8). El `Store` lleva un indicador atómico que dice si hay
  transacción viva, y `WithTx`, `Query`, `Exec` y la lectura de pragmas lo miran antes de tocar el
  manejador: devuelven `ErrTxInProgress`. Vale igual para un `WithTx` anidado.
- **`WithTx` deshace la transacción también si la función que recibe entra en pánico**, y deja que
  el pánico siga su camino. Sin eso, un fallo así dejaría tomada la única conexión del punto
  anterior y el indicador del punto anterior armado para siempre.
- **La migración comprueba y actúa dentro de una única transacción inmediata.** Leer
  `PRAGMA user_version` por fuera y decidir a partir de ahí no es atómico: dos procesos que abren a
  la vez el mismo fichero rancio verían los dos la versión vieja y aplicarían los dos el mismo
  script, y el segundo fallaría con "table already exists". Por eso la transacción vuelve a leer la
  versión ya con el bloqueo de escritura tomado, aplica lo que falte y la actualiza en el mismo
  commit. La lectura de antes de la transacción no desaparece, pero solo decide si hay algo que
  hacer: abrir un tablero ya migrado, que es el caso de cada invocación, no puede tomar el bloqueo
  de escritura, porque entonces una lectura fallaría mientras otro proceso escribe, contra la
  garantía 6.
- **La comprobación de integridad es un método aparte, `CheckIntegrity`, y no parte de abrir.**
  `PRAGMA integrity_check` recorre el fichero entero, y el presupuesto de arranque de
  `docs/spec/presupuestos.md#el-presupuesto-de-arranque` no lo paga en cada invocación. Quien la
  ejecuta es `biso doctor` (`docs/spec/cmd/doctor.md`); devuelve el mismo error que un fichero que
  no abre, que es lo que pide `docs/spec/garantias.md`.

**El único sitio donde la espera de la garantía 5 se hace a mano: crear el fichero.** Varias
conexiones que crean el mismo fichero desde cero en el mismo instante no se serializan solas: la
primera que llega lo pasa a modo WAL, y SQLite rechaza ese cambio de modo de diario mientras haya
otra conexión activa devolviendo `SQLITE_BUSY` de inmediato, sin consultar el manejador de ocupado
y por tanto sin pasar por el tiempo de espera. Medido: cuatro conexiones creando el mismo fichero a
la vez fallan en varios intentos de veinte, mientras que sobre un fichero ya en modo WAL, aunque
esté vacío de esquema, fallan cero de veinte.

Eso **sí afectaba a la garantía 5**, que promete esperar hasta cinco segundos antes de salir con
código 8 y no exime al comando que crea el tablero, así que el almacén no se rinde ahí: la primera
lectura de `PRAGMA user_version`, que es la sentencia que de verdad conecta, se reintenta con
espera creciente hasta agotar el tiempo configurado, y solo entonces devuelve el error de ocupado.
Cualquier fallo que no sea el bloqueo de escritura no se reintenta, porque significa que no se pudo
leer ni la cabecera de la base de datos, que es el otro caso de `docs/spec/garantias.md`. Lo cubre
`TestConcurrentCreationOfTheSameFileWaitsInsteadOfFailingAtOnce`, veinte intentos de cuatro
conexiones simultáneas sobre un fichero que no existe.

**Los dos errores de la especificación que nacen aquí.** El almacén es la capa más profunda que
distingue estos dos casos, así que construye el `model.Error` completo de cada uno y lo deja subir
sin que nadie lo reinterprete:

| Caso | Código | Clave | Dónde se detecta |
|---|---|---|---|
| El bloqueo de escritura no se consiguió en el tiempo configurado | 8 | `busy` | al abrir, al migrar y en `WithTx` |
| La base de datos no se puede leer, o falla su comprobación de integridad | 21 | `database_unreadable` | al abrir y en `CheckIntegrity` |

Los dos llevan el texto literal de `docs/spec/garantias.md`, mensaje y líneas de `hint` incluidas.
Cualquier camino que se tope con el bloqueo devuelve el primero: no vale que abrir o migrar
devuelvan un error del controlador en crudo, porque entonces el mismo estorbo daría códigos de
salida distintos según en qué momento apareciera.

### 3.4. `internal/board`

Junta tres cosas en un único tipo `Board`: dónde está el tablero (la resolución de
`docs/spec/resolucion-del-tablero.md`), su configuración propia ya cargada, y el `Store` ya abierto
sobre él. Es
deliberadamente delgado, solo resuelve y abre; no contiene lógica de negocio de ningún comando. Es
el primer argumento que recibe cada función de `internal/ops`.

**La configuración de un tablero vive dentro de su base de datos, no en un `config.json` propio**,
que es lo que decía este documento antes del paso 4 y ya no es cierto. Lo manda
`docs/spec/resolucion-del-tablero.md#el-orden-de-búsqueda`, que se apoya en que un directorio con
`board.db` es ese tablero y no le falta ningún dato, y `docs/spec/cmd/config.md`, donde renombrar un
tablero es una escritura en su base de datos con las mismas garantías que cualquier otra. La escribe
la segunda migración de `internal/store`, en una tabla de una fila por clave, con la identidad del
tablero en una tabla aparte porque no es configuración: `biso config list` no la lista y nada puede
cambiarla.

Aquí vive además la traducción entre una tarea y sus filas, en el tipo `Tasks`, que existe desde
TASK-10. Sus decisiones de implementación, que no se deducen de la especificación:

- **La cadena vacía es la ausencia de valor en todo campo `string` y `text`, y el instante cero lo
  es en todo campo de fecha.** `docs/spec/valores-de-entrada.md#el-valor-vacío` hace que la cadena
  vacía no sea nunca un valor que quien llama pueda guardar: en un escalar es un error, y vaciar un
  campo es un flag propio. Así que "" y "sin valor" no se pueden distinguir por nada observable, y
  ni el modelo ni el esquema cargan con la diferencia; el contrato JSON los escribe como `null`. **Ya
  no hay excepción.** La tuvo `ordinal` mientras fue un entero, porque el 0 era un valor que se podía
  pedir; desde que es la clave de texto de
  ["El orden manual y su clave"](../../spec/modelo-de-datos/orden-manual.md), cuyo alfabeto no puede
  escribir la cadena vacía, sigue la regla común: ninguna columna de la tabla `task` admite `NULL` y
  ningún campo del tipo `Task` es un puntero por ese motivo.
- **Los seis campos `list<string>` comparten una tabla**, `task_list_item`, con el nombre del campo
  como columna y un `CHECK` que la deja cerrada, en vez de una tabla por campo. Tienen la misma
  forma, el vocabulario cerrado se comprueba igual en SQL que en `internal/model`, y una sola tabla
  significa que una consulta trae todas las listas de todas las tareas. Cada elemento guarda su
  posición, porque `docs/spec/garantias.md` dice que las listas no se ordenan solas: el orden en que
  se escribieron es un dato.
- **Los dos contadores de claves de una tarea se guardan con ella**, `next_criterion_key` y
  `next_comment_key`, en vez de deducirse de la clave más alta presente. Son la razón de que quitar
  el criterio de en medio no libere su clave: deducirla del máximo la reasignaría en cuanto se
  borrara el último.
- **Leer el tablero son cinco consultas y nunca una por tarea**: la de las tareas y una por cada
  tabla hija, ordenadas por tarea y por posición, que se juntan en memoria. Es lo que hace que el
  presupuesto de arranque de `docs/spec/presupuestos.md` sea un puñado de recorridos y no mil
  quinientas idas y venidas.
- **`Create` asigna el identificador dentro de la misma transacción en la que escribe la tarea**, con
  un contador propio del tablero (`board_counter`) que solo crece. Es lo que hace cierta la garantía
  3, y de paso que un identificador no se reutilice nunca. `Create` rellena con el reloj las fechas
  que la tarea no trae y respeta las que sí, que es justo lo que necesita importar un lote.

### 3.5. `internal/ops`

Una función por comando: `New`, `Ls`, `Get`, `Set`, `Start`, `Note`, `Comment`, `Finish`, `Ask`,
`Answer`, `Archive`, `Export`, `Snapshot`, `Config`, `Doctor`, `Where`, `Init`, `Prime`, `BoardInfo`
(el comando `biso board`, con nombre distinto del tipo `board.Board` para no leer `board.Board`
como argumento de una función llamada `Board`), `Help`. La firma de cada una es
`func(b *board.Board, p XParams) (XResult, error)`.

**`Init` y `Where` son la excepción, y son la razón de que exista el entorno que reciben en su
lugar**, un tipo `Env` con el directorio de trabajo, la configuración de máquina, la identidad de
quien llama y las dos fuentes que una prueba necesita poder sustituir, el reloj y el generador de
identificadores. A ninguno de los dos se le puede entregar un tablero ya abierto: uno lo crea y el
otro existe para explicar cómo se encontró, incluso cuando no se encuentra ninguno. Cada una de las
dos abre el suyo por dentro y lo cierra antes de devolver.

**Ese `Env` lo construye esta capa y no la de arriba**, con `ops.NewEnv`, que es quien lee
`~/.biso/config.json` (`docs/spec/invocacion.md#configuración-de-máquina`) y quien resuelve la
identidad de quien llama entre `BISO_ME` y la clave `me`. `internal/cli` le pasa solo lo que sabe
sin abrir ningún fichero: el directorio de trabajo ya resuelto, el directorio personal y el valor
de las variables de entorno. Si la construyera `internal/cli`, esa capa tendría que importar
`internal/board`, que es justo lo que la regla de dependencias de la sección 3 no admite.

Aquí vive toda
la lógica de negocio de `docs/spec/`: qué combinación de campos es válida, qué garantías hay que
respetar dentro de la transacción, qué mensaje de error corresponde a qué caso. Nada de esta capa
imprime texto ni construye JSON.

**La resolución de una referencia a una tarea (`docs/spec/referencias.md`)** la usan `get`, `set`,
`archive`, `--dep`, `--parent` y el `-p` de `ls`, pero a diferencia del algoritmo de coincidencia de
`internal/match` necesita leer el tablero (para la búsqueda por texto de
`docs/spec/referencias.md#la-búsqueda-por-texto`), así que no
puede ser una función pura de un paquete hoja. Vive como una función interna de `internal/ops`,
compartida por las funciones que la necesitan, en vez de un paquete propio: la diferencia con
`match` no es de importancia sino de si depende o no de `internal/board`.

**La simetría entre `biso export` (`docs/spec/cmd/export.md`) y `biso new --from`
(`docs/spec/cmd/new.md`), y la garantía que exige el `CLAUDE.md` del proyecto,** se resuelve con una única definición del formato de
intercambio NDJSON, en un fichero propio de `internal/ops` (`ops/interchange.go`), con una función de
codificación y una de decodificación que usan directamente las etiquetas `json` de los tipos de
`internal/model`. `Export` llama a la de codificación y `New` (en su variante `--from`) llama a la de
decodificación; no hay una segunda implementación en ningún lado que pueda desincronizarse de la
primera. Esto no contradice que `internal/model` no sepa qué es JSON: los tipos llevan las etiquetas
que fijan su forma en la conexión, pero ninguna función de `internal/model` codifica ni decodifica
nada, eso ocurre siempre desde `internal/ops`.

### 3.6. `internal/cli`

Traduce `argv` a los `Params` tipados de cada comando, invoca la función de `internal/ops`
correspondiente, y traduce el `Result` (o el error) a la salida literal de texto o al sobre JSON de
`docs/spec/contrato-json.md`, según `--json` y la detección de terminal de
`docs/spec/salida-y-terminal.md#interactividad-terminal-y-color`. Es la
única capa que conoce el texto exacto de cada `--help`, que vive como constante propia por
comando (ver sección 6 de este documento).

### 3.7. `internal/vcs`

Ejecuta el sistema de control de versiones configurado (`git` en la versión 1.0, según
`docs/spec/cmd/snapshot.md#el-sistema-de-control-de-versiones`) para `biso snapshot`: las cinco
preguntas que le hace al repositorio y las cuatro órdenes que le da, según
`docs/spec/cmd/snapshot.md#cómo-se-ejecutan-las-órdenes`. Es el único paquete que lanza procesos
externos, y solo
lo usa `internal/ops` al implementar `Snapshot`; ningún otro comando lo toca.

## 4. El tipo de error transversal

Un único tipo, en `internal/model`, con los campos que exige la tabla de
`docs/spec/contrato-json.md#los-errores-en-json`, no solo los tres que están siempre presentes:

```go
type Error struct {
    ExitCode int      // the exit code, see docs/spec/codigos-de-salida.md
    Code     string   // the identifier, see docs/spec/contrato-json.md#los-identificadores-de-error
    Message  string   // the text that follows "error: " on stderr
    Detail   []string // verbatim display lines printed right after the message
    Notes    []string // the "note: " lines that follow those, in order
    Hints    []string // the "hint: " lines that follow those, in order

    // The five detail fields from docs/spec/contrato-json.md#los-errores-en-json, each present
    // only on the subset of codes it corresponds to per that same table; on the rest they stay
    // at their zero value and are omitted when serialized (omitempty).
    Field     string   // on exit code 3, and exit code 2 errors that name a flag
    Given     string   // always accompanies Field
    Valid     []string // on errors that reject a value against a known set
    Details   []*Error // only on batch_invalid and dry_run_failed, same shape, one per failure
    VCSOutput []string // only on vcs_commit_failed and vcs_push_failed
}
```

**Los hints son una lista, `Hints []string`.** Buena parte de los casos de `docs/spec/` llevan una
línea `hint: ...` por stderr detrás del mensaje, con texto que casi siempre depende de datos
concretos del caso (los dos directorios de un `id` duplicado, cuántas tareas usan un estado, el `id`
que hay que adoptar), y dos casos llevan dos de esas líneas: la base de datos ilegible de
`docs/spec/garantias.md` y el tablero ya existente de `docs/spec/cmd/init.md`. `Hints` guarda una
entrada por línea, en el orden en que la especificación las imprime, y queda vacía cuando el caso no
tiene ninguna. Vive en `model.Error`, en el mismo sitio y con la misma disciplina que `Message`, en
vez de en una tabla `code -> hint` aparte en `internal/cli`, porque la mayoría de los hints no son
fijos: solo quien detecta el error, en la capa que sea, tiene los datos concretos para construirlo.
`cmd/biso/main.go` imprime una línea `hint: ` por entrada. **No entra en el sobre JSON**:
`docs/spec/contrato-json.md#los-errores-en-json` no tiene una clave `hint` en el objeto de error, así
que este campo es exclusivo de la salida de texto.

**Y las notas son otra lista igual, `Notes []string`,** por el mismo motivo y con la misma
disciplina. Hay una tercera clase de línea por stderr detrás del mensaje, la que empieza por
`note: `, y dos errores de la especificación la llevan: los dos de
`docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro` que dicen que una tarea no está,
donde la nota da el dato que distingue un caso del otro (el identificador más alto que el tablero
ha llegado a asignar, o que ese en concreto se asignó alguna vez). El campo lo añadió TASK-10, que
es la primera tarea que construye esos dos errores; antes el tipo no podía llevar esas líneas.
Tampoco entra en el sobre JSON, por la misma razón que `Hints`.

(Descartado: un único `Hint string` con las dos líneas separadas por un salto de línea. Obligaría a
quien imprime a partir la cadena para poner el prefijo `hint: ` en cada línea, es decir, a
reinterpretar el contenido del error, que es justo lo que este tipo evita.)

**Una entrada de `Hints` o de `Notes` sí puede llevar saltos de línea dentro, y eso no contradice lo
anterior.** Lo descartado era meter dos hints distintos en una cadena; esto es un solo hint que la
especificación imprime repartido en varias líneas de pantalla, y **dónde parte es texto fijo y no un
ancho**: las dos notas de `docs/spec/cmd/init.md` cortan entre las mismas palabras tanto con
`--at tablero` como con `--at my-project-board`, así que no hay ningún ajuste de línea que reproduzca
la especificación, y calcularlo daría un texto distinto del que esa página fija. Quien imprime pone el
prefijo en la primera línea y alinea las demás debajo, que son seis espacios para `note: ` y para
`hint: `. Sigue sin reinterpretar nada: solo sangra.

**`Detail []string` es una tercera clase de línea, y es literal.** Son las líneas de pantalla que dos
casos de la especificación imprimen entre el mensaje y los hints, con su propia sangría y sus propias
columnas internas: los dos directorios de un `id` duplicado
(`docs/spec/resolucion-del-tablero.md#el-mismo-id-en-dos-sitios`) y el bloque `searched` de
`biso where` (`docs/spec/cmd/where.md`). No caben en `Notes`, que les pondría el prefijo `note: `
delante, ni en `Message`, que es una línea y además viaja al sobre JSON. Quien imprime las escribe tal
cual, sin tocarlas. El campo lo añadió TASK-12, la primera tarea que construye esos dos errores, y
tampoco entra en el sobre JSON.

(El código Go va siempre en inglés, según el `CLAUDE.md` del proyecto; la prosa de este documento
sigue en español.)

Nace en la capa más profunda que detecta el caso (normalmente `internal/ops`, a veces
`internal/store` si el fallo es de disco) ya con los campos de detalle que le correspondan, por
ejemplo `internal/match` es quien conoce la lista `Valid` cuando un valor no coincide con ningún
vocabulario, así que es quien la rellena antes de devolver el error hacia arriba. Desde que se crea
viaja sin cambios hasta la capa de salida de `internal/cli`, a la que `cmd/biso/main.go` entrega la
llamada entera y de la que recibe el código de salida del proceso: es el único sitio que lo traduce a
texto en
`stderr`, a la envoltura JSON de error de `docs/spec/contrato-json.md#los-errores-en-json`, y al
código de salida del proceso.
Ningún paquete intermedio reinterpreta ni envuelve este error; solo lo propaga.

## 5. Librerías externas

Además de `modernc.org/sqlite` (ya elegido,
`docs/decisiones/lenguaje-y-rendimiento.md#el-controlador-de-sqlite-es-moderncorgsqlite-sin-cgo`),
la única librería externa que hace falta es `golang.org/x/term`, para detectar terminal interactiva
y color (`docs/spec/salida-y-terminal.md#interactividad-terminal-y-color`); es la extensión oficial
de Go para eso y no trae nada más consigo.

No hace falta ninguna librería de parseo de línea de comandos: se descarta explícitamente cobra,
kong y equivalentes (sección 6 de este documento explica por qué). Tampoco hace falta ninguna
librería de fechas (las fechas son instantes UTC con precisión de segundo, `time.Parse(time.RFC3339)`
de la biblioteca estándar basta) ni de JSON (`encoding/json` de la biblioteca estándar). `git` se
invoca con `os/exec` desde `internal/vcs`, sin ninguna librería intermedia.

## 6. El analizador de línea de comandos, y por qué es propio

El contrato de `docs/spec/valores-de-entrada.md` y `docs/spec/garantias.md` incluye reglas que una
librería de propósito general no ofrece:

- Tres formas de dar el mismo valor largo (literal, `@fichero`, `-` para entrada estándar), con
  `@@` como único escape y con la regla de que `-` solo puede aparecer una vez por invocación
  entera, no una vez por flag (`docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo`).
- Un valor vacío que no borra nada salvo en los flags que sustituyen
  (`docs/spec/valores-de-entrada.md#el-valor-vacío`).
- Un valor que empieza por guion se acepta tal cual detrás de un flag que exige valor, sin
  heurísticas; olvidar un valor se detecta por lo que sobra después, no por lo que parece un
  flag (`docs/spec/valores-de-entrada.md#valores-que-empiezan-por-guion`).
- El texto literal de cada `--help` está fijado carácter a carácter en la propia página de cada
  comando, bajo `docs/spec/cmd/`.

Ninguna librería de parseo genera esto por defecto, y forzarla a hacerlo significa sobreescribir su
comportamiento en casi todos los comandos, que es perder la ventaja de usarla. Por eso
`internal/cli` tiene su propio analizador: una tabla de especificación por comando (nombre de
flag, si acepta valor, si es de las "tres formas" de arriba, si es lista o escalar, según la tabla
de parámetros de cada página de `docs/spec/cmd/`) recorrida por una única función de parseo
genérica, de forma que las reglas transversales se escriben una sola vez y no trece. Además de las
tres formas de arriba, esa tabla es también la que resuelve la acumulación de un flag de lista
marcado como repetible (`docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas`:
acumula valores en cada aparición, en vez de quedarse con la última, que es lo que hace una
librería genérica por defecto) y el orden de aplicación
(`docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura`: la función de parseo no
aplica los cambios en el orden en que llegaron en `argv`, sino que los clasifica en las nueve
categorías fijas de esa sección, de "todos los `--clear-*`" hasta "los comentarios", y dentro de
cada categoría sí respeta el orden de `argv`). El texto de cada `--help` no se deriva de esa tabla:
es una constante propia por comando, transcrita de su página de `docs/spec/cmd/`, para no
introducir una capa de generación que tendría que mantenerse sincronizada a mano de todos modos.

## 7. Estrategia de pruebas

Dado el peso que `docs/spec/` pone en el texto literal, cada comando lleva pruebas de fichero
dorado: la salida exacta de cada bloque de `docs/spec/cmd/<comando>.md` vive como fixture junto al
test, y la prueba compara carácter a carácter; lo mismo para cada esquema JSON de la página
correspondiente. Una regeneración de un fixture es un cambio deliberado y visible en el diff, nunca
automática.

`internal/match` se prueba con una tabla de casos, sin ningún I/O.

El presupuesto de arranque de `docs/spec/presupuestos.md#el-presupuesto-de-arranque` se mide con el
binario real compilado, como ya advierte `TASK-4` (el paso 1 de `TASK-55`): no es una prueba
unitaria de una función, es una prueba de integración que ejecuta `biso` de verdad y mide el reloj.

### 7.1. El round-trip de `Params`, decidido el 2026-09-19

No hay un contrato JSON genérico para los `Params` de todos los comandos: casi ningún comando aparte
de `ls` hace eco de sus parámetros en su propia salida JSON (`new`/`set` devuelven la tarea
resultante, no los flags que la produjeron), así que inventar una forma JSON para esos `Params`
sería una capa sin consumidor real.

**Para `biso ls` sí.** `Params` de `ls` lleva las mismas etiquetas `json` que
`docs/spec/contrato-json.md#los-filtros-de-biso-ls` fija para `data.filters`. `internal/cli` no
traduce nada a mano para producir `data.filters`: hace `json.Marshal` directamente sobre `Params`,
así que no hay una segunda implementación que se pueda desincronizar de la primera, el mismo
principio que ya sigue `ops/interchange.go` con la simetría export/import (sección 3.5). La prueba
de round-trip es serializar y volver a deserializar `Params`, comparando structs, para que un
filtro nuevo nunca se pierda en silencio al codificarlo o decodificarlo.

**Para todos los comandos, una segunda propiedad, sin JSON de por medio.** El analizador de línea de
comandos de la sección 6 se prueba con una tabla de invocaciones representativas por comando:
`parse(argv) -> Params`, luego una función `render(Params) -> argv` que solo existe en el test (no
es una interfaz pública, nadie la usa en producción) reconstruye un `argv` canónico a partir de esa
misma tabla de especificación por flag, y `parse(render(Params))` tiene que dar un `Params`
idéntico al primero. Esto comprueba que el analizador es determinista sin necesitar una segunda
implementación de referencia, y atrapa la clase de bug donde dos invocaciones equivalentes se
resuelven a valores distintos.

**La simetría export/import no cambia** respecto a la sección 3.5: sigue siendo la única definición
NDJSON de `ops/interchange.go`, usada por `Export` y por `New --from`, con la prueba de que exportar
e importar en un tablero vacío da dos tableros idénticos campo a campo (criterio de aceptación de
`TASK-16`, y la garantía que exige el `CLAUDE.md` del proyecto).

## 8. Lo que este documento deja fuera a propósito

- **El nombre del módulo Go** (la primera línea de `go.mod`) no está decidido aquí porque depende
  de dónde acabe viviendo el repositorio, y no cambia nada de la arquitectura: se fija al crear
  `go.mod`, en `TASK-4`, el primer paso de `TASK-55`.
- **Si la interfaz gráfica futura vive en este módulo o en otro** queda abierto, como se explica en
  la sección 2. No bloquea nada de lo descrito aquí.
- **El orden de implementación no cambia.** Este documento describe cómo se organiza el código
  dentro de los nueve pasos de `TASK-55`; no propone un orden distinto.
