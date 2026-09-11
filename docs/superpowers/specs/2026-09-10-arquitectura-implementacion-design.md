# La arquitectura del código de `biso`

## 1. Qué resuelve este documento, y qué no

`docs/SPEC.md` fija el comportamiento observable de `biso`: qué hace cada comando, qué imprime,
qué código de salida devuelve, qué esquema JSON expone. No dice nada de cómo se organiza el código
que lo implementa, porque no le corresponde: eso es una decisión de implementación, no de contrato.

Este documento cubre exactamente ese hueco: los paquetes Go que va a tener el programa, la
responsabilidad de cada uno, sus dependencias entre sí, las librerías externas que hacen falta y
la estrategia de pruebas. No repite ni reinterpreta ningún comportamiento de `SPEC.md`; donde haya
duda sobre qué hace un comando, la respuesta sigue estando allí, no aquí.

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
de la sección 12 de `SPEC.md` vive en una sola capa delgada, porque toda la decisión de negocio ya
la tomó la capa de debajo. Esa capa delgada es la única que cambiaría si algún día se añade una
interfaz gráfica; la lógica de los comandos no cambiaría nada.

**Sobre si la interfaz será un binario Go en este mismo módulo o un proceso aparte que hable el
`--json` ya especificado en la sección 12: queda sin decidir a propósito**, porque no hace falta
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
internal/match          el algoritmo de coincidencia de la sección 6.1, función pura
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

Tipos y validación, sin tocar disco ni red: `Task`, `Criterion` (con su clave estable, sección 5.1
de `SPEC.md`), `Comment`, `Question`, `Urgency`, y el tipo de error transversal de la sección 4 de
este documento. Ningún tipo de aquí sabe qué es SQLite ni qué es JSON; sus métodos son de
validación pura (por ejemplo, si un campo escalar admite cadena vacía o no, sección 4.6 de
`SPEC.md`).

### 3.2. `internal/match`

El algoritmo de coincidencia de vocabulario (sección 6.1 de `SPEC.md`) y sus reglas relacionadas
(6.2, la equivalencia en los dos sentidos; 6.3, qué valida cada filtro). Es un paquete propio,
aunque pequeño, porque la propia especificación insiste en que es una función pura de veinte líneas
de la que dependen `ls`, `set`, `new` y todos los vocabularios cerrados: aislarla la hace trivial de
probar con una tabla de casos, y evita que cada comando reimplemente su propia variante de la
regla.

### 3.3. `internal/store`

La única capa que importa `modernc.org/sqlite` y `database/sql`. Abre o crea la base de datos,
activa el modo WAL, gestiona las migraciones del esquema, y expone una única función
`WithTx(func(*sql.Tx) error) error` que es el único punto por el que pasa cualquier escritura: de
ahí salen las seis garantías de la sección 4.10 de `SPEC.md`. No sabe qué es una `Task`, solo filas
y columnas; traducir entre unas y otras es trabajo de `internal/board`.

### 3.4. `internal/board`

Junta tres cosas en un único tipo `Board`: dónde está el tablero (la resolución de la sección 3.2
de `SPEC.md`), su `config.json` propio ya cargado, y el `Store` ya abierto sobre él. Es
deliberadamente delgado, solo resuelve y abre; no contiene lógica de negocio de ningún comando. Es
el primer argumento que recibe cada función de `internal/ops`.

### 3.5. `internal/ops`

Una función por comando: `New`, `Ls`, `Get`, `Set`, `Start`, `Note`, `Comment`, `Finish`, `Ask`,
`Answer`, `Archive`, `Export`, `Snapshot`, `Config`, `Doctor`, `Where`, `Init`, `Prime`, `BoardInfo`
(el comando `biso board`, con nombre distinto del tipo `board.Board` para no leer `board.Board`
como argumento de una función llamada `Board`), `Help`. La firma de cada una es
`func(b *board.Board, p XParams) (XResult, error)`. Aquí vive toda
la lógica de negocio de `SPEC.md`: qué combinación de campos es válida, qué garantías hay que
respetar dentro de la transacción, qué mensaje de error corresponde a qué caso. Nada de esta capa
imprime texto ni construye JSON.

**La resolución de una referencia a una tarea (sección 7 de `SPEC.md`)** la usan `get`, `set`,
`archive`, `--dep`, `--parent` y el `-p` de `ls`, pero a diferencia del algoritmo de coincidencia de
`internal/match` necesita leer el tablero (para la búsqueda por texto de la sección 7.2), así que no
puede ser una función pura de un paquete hoja. Vive como una función interna de `internal/ops`,
compartida por las funciones que la necesitan, en vez de un paquete propio: la diferencia con
`match` no es de importancia sino de si depende o no de `internal/board`.

**La simetría entre `biso export` y `biso new --from` (secciones 10.3 y 10.9 de `SPEC.md`, y la
garantía que exige el `CLAUDE.md` del proyecto)** se resuelve con una única definición del formato de
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
la sección 12 de `SPEC.md`, según `--json` y la detección de terminal de la sección 4.1. Es la
única capa que conoce el texto exacto de cada `--help`, que vive como constante propia por
comando (ver sección 6 de este documento).

### 3.7. `internal/vcs`

Ejecuta el sistema de control de versiones configurado (`git` en la versión 1.0, según la sección
3.3 de `SPEC.md`) para `biso snapshot`: las dos preguntas que le hace al repositorio y las tres
órdenes que le da, según la sección 10.14. Es el único paquete que lanza procesos externos, y solo
lo usa `internal/ops` al implementar `Snapshot`; ningún otro comando lo toca.

## 4. El tipo de error transversal

Un único tipo, en `internal/model`, con los campos que exige la tabla de la sección 12.2 de
`SPEC.md`, no solo los tres que están siempre presentes:

```go
type Error struct {
    ExitCode int      // el código de salida de la sección 2
    Code     string   // el identificador de la sección 12.3
    Message  string   // el texto que sigue a "error: " en stderr

    // Las cinco claves de detalle de la sección 12.2, cada una presente solo en el
    // subconjunto de codes que le corresponde según esa misma tabla; en los demás
    // casos queda en su cero y se omite al serializar (omitempty).
    Field     string   // en los del código 3, y los del 2 que nombran una bandera
    Given     string   // acompaña siempre a Field
    Valid     []string // en los que rechazan un valor contra un conjunto conocido
    Details   []*Error // solo en batch_invalid y dry_run_failed, misma forma, uno por fallo
    VCSOutput []string // solo en vcs_commit_failed y vcs_push_failed
}
```

Nace en la capa más profunda que detecta el caso (normalmente `internal/ops`, a veces
`internal/store` si el fallo es de disco) ya con los campos de detalle que le correspondan, por
ejemplo `internal/match` es quien conoce la lista `Valid` cuando un valor no coincide con ningún
vocabulario, así que es quien la rellena antes de devolver el error hacia arriba. Desde que se crea
viaja sin cambios hasta `cmd/biso/main.go`, que es el único sitio que lo traduce a texto en
`stderr`, a la envoltura JSON de error de la sección 12.2, y al código de salida del proceso.
Ningún paquete intermedio reinterpreta ni envuelve este error; solo lo propaga.

## 5. Librerías externas

Además de `modernc.org/sqlite` (ya elegido, apartado 14.1 de `DECISIONES.md`), la única librería
externa que hace falta es `golang.org/x/term`, para detectar terminal interactiva y color (sección
4.1 de `SPEC.md`); es la extensión oficial de Go para eso y no trae nada más consigo.

No hace falta ninguna librería de parseo de línea de comandos: se descarta explícitamente cobra,
kong y equivalentes (sección 6 de este documento explica por qué). Tampoco hace falta ninguna
librería de fechas (las fechas son instantes UTC con precisión de segundo, `time.Parse(time.RFC3339)`
de la biblioteca estándar basta) ni de JSON (`encoding/json` de la biblioteca estándar). `git` se
invoca con `os/exec` desde `internal/vcs`, sin ninguna librería intermedia.

## 6. El analizador de línea de comandos, y por qué es propio

El contrato de `SPEC.md` incluye reglas que una librería de propósito general no ofrece:

- Tres formas de dar el mismo valor largo (literal, `@fichero`, `-` para entrada estándar), con
  `@@` como único escape y con la regla de que `-` solo puede aparecer una vez por invocación
  entera, no una vez por bandera (sección 4.5).
- Un valor vacío que no borra nada salvo en las banderas que sustituyen (sección 4.6).
- Un valor que empieza por guion se acepta tal cual detrás de una bandera que exige valor, sin
  heurísticas; olvidar un valor se detecta por lo que sobra después, no por lo que parece una
  bandera (sección 4.7).
- El texto literal de cada `--help` está fijado carácter a carácter en `SPEC.md`.

Ninguna librería de parseo genera esto por defecto, y forzarla a hacerlo significa sobreescribir su
comportamiento en casi todos los comandos, que es perder la ventaja de usarla. Por eso
`internal/cli` tiene su propio analizador: una tabla de especificación por comando (nombre de
bandera, si acepta valor, si es de las "tres formas" de la 4.5, si es lista o escalar, sección 8 de
`SPEC.md`) recorrida por una única función de parseo genérica, de forma que las reglas
transversales se escriben una sola vez y no trece. Además de 4.5 a 4.7, esa tabla es también la que
resuelve 4.8 (una bandera de lista marcada como repetible acumula valores en cada aparición, en vez
de quedarse con la última, que es lo que hace una librería genérica por defecto) y 4.9 (la función
de parseo no aplica los cambios en el orden en que llegaron en `argv`, sino que los clasifica en las
siete categorías fijas de esa sección, de "todos los `--clear-*`" hasta "los comentarios", y dentro
de cada categoría sí respeta el orden de `argv`). El texto de cada `--help` no se deriva de esa tabla: es una constante
propia por comando, transcrita de `SPEC.md`, para no introducir una capa de generación que tendría
que mantenerse sincronizada a mano de todos modos.

## 7. Estrategia de pruebas

Dado el peso que `SPEC.md` pone en el texto literal, cada comando lleva pruebas de fichero dorado:
la salida exacta de cada bloque de la sección 10 vive como fixture junto al test, y la prueba
compara carácter a carácter; lo mismo para cada esquema JSON de la sección correspondiente. Una
regeneración de un fixture es un cambio deliberado y visible en el diff, nunca automática.

`internal/match` se prueba con una tabla de casos, sin ningún I/O.

El presupuesto de arranque de la sección 4.13 de `SPEC.md` se mide con el binario real compilado,
como ya advierte el paso 1 de la sección 15: no es una prueba unitaria de una función, es una
prueba de integración que ejecuta `biso` de verdad y mide el reloj.

## 8. Lo que este documento deja fuera a propósito

- **El nombre del módulo Go** (la primera línea de `go.mod`) no está decidido aquí porque depende
  de dónde acabe viviendo el repositorio, y no cambia nada de la arquitectura: se fija al crear
  `go.mod`, en el primer paso de la sección 15 de `SPEC.md`.
- **Si la interfaz gráfica futura vive en este módulo o en otro** queda abierto, como se explica en
  la sección 2. No bloquea nada de lo descrito aquí.
- **El orden de implementación no cambia.** Este documento describe cómo se organiza el código
  dentro de los nueve pasos de la sección 15 de `SPEC.md`; no propone un orden distinto.
