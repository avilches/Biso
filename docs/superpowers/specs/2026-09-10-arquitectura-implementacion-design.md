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

### 3.4. `internal/board`

Junta tres cosas en un único tipo `Board`: dónde está el tablero (la resolución de
`docs/spec/resolucion-del-tablero.md`), su `config.json` propio ya cargado, y el `Store` ya abierto
sobre él. Es
deliberadamente delgado, solo resuelve y abre; no contiene lógica de negocio de ningún comando. Es
el primer argumento que recibe cada función de `internal/ops`.

### 3.5. `internal/ops`

Una función por comando: `New`, `Ls`, `Get`, `Set`, `Start`, `Note`, `Comment`, `Finish`, `Ask`,
`Answer`, `Archive`, `Export`, `Snapshot`, `Config`, `Doctor`, `Where`, `Init`, `Prime`, `BoardInfo`
(el comando `biso board`, con nombre distinto del tipo `board.Board` para no leer `board.Board`
como argumento de una función llamada `Board`), `Help`. La firma de cada una es
`func(b *board.Board, p XParams) (XResult, error)`. Aquí vive toda
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
`docs/spec/cmd/snapshot.md#el-sistema-de-control-de-versiones`) para `biso snapshot`: las dos
preguntas que le hace al repositorio y las tres órdenes que le da, según
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
    Hint     string   // the text that follows "hint: " on stderr, empty when the case has none

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

**`Hint` decidido el 2026-09-19.** Buena parte de los casos de `docs/spec/` llevan una segunda línea
`hint: ...` por stderr, con texto que casi siempre depende de datos concretos del caso (los dos
directorios de un `id` duplicado, cuántas tareas usan un estado, el `id` que hay que adoptar). Vive en
`model.Error`, en el mismo sitio y con la misma disciplina que `Message`, en vez de en una tabla
`code -> hint` aparte en `internal/cli`, porque la mayoría de los hints no son fijos: solo quien
detecta el error, en la capa que sea, tiene los datos concretos para construirlo. `cmd/biso/main.go`
imprime la segunda línea solo cuando `Hint` no está vacío. **No entra en el sobre JSON**:
`docs/spec/contrato-json.md#los-errores-en-json` no tiene una clave `hint` en el objeto de error, así
que este campo es exclusivo de la salida de texto.

(El código Go va siempre en inglés, según el `CLAUDE.md` del proyecto; la prosa de este documento
sigue en español.)

Nace en la capa más profunda que detecta el caso (normalmente `internal/ops`, a veces
`internal/store` si el fallo es de disco) ya con los campos de detalle que le correspondan, por
ejemplo `internal/match` es quien conoce la lista `Valid` cuando un valor no coincide con ningún
vocabulario, así que es quien la rellena antes de devolver el error hacia arriba. Desde que se crea
viaja sin cambios hasta `cmd/biso/main.go`, que es el único sitio que lo traduce a texto en
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
