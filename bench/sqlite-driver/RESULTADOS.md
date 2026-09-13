# El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`

Este documento cierra la decisión que quedó aplazada a propósito y que ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go) declaraba como lo primero que la implementación tenía que
resolver. Está escrito para que su contenido acabe en esa misma sección. El banco de pruebas del que
salen todas las cifras está en este mismo directorio, y su [`README.md`](README.md) dice cómo se
vuelve a ejecutar.

**El presupuesto se cumple, y con mucho margen.** Los cuatro candidatos entran en los 25
milisegundos de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque) sobre un tablero de 300 tareas, en las dos
plataformas donde se midió. En la máquina de referencia, que la propia sección define como la
que ejecuta la integración continua y por tanto es Linux, el más lento de los cuatro tarda **3,2
milisegundos**, casi ocho veces por debajo del presupuesto. En un portátil macOS, el más lento tarda
**14,5 milisegundos**, y de esos 8,1 son el suelo del sistema operativo para arrancar cualquier
proceso.

**Así que el rendimiento no decide, y lo que decide es la distribución del binario.** Ahí la
diferencia no es de grado: el enlace con la biblioteca en C no puede producir desde esta máquina un
binario para Linux ni para Windows, mientras que las tres traducciones a Go puro producen los tres
binarios con un `go build` y nada más instalado.

---

## 1. Con qué se midió

| | |
|---|---|
| Máquina | Apple M3 Max, 16 núcleos (12 de rendimiento), 64 GiB de RAM |
| Sistema | macOS 26.5.2 (25F84), Darwin 25.5.0 arm64 |
| Go | go1.27.1 darwin/arm64 |
| Compilador de C | Apple clang 21.0.0 (clang-2100.1.1.101) |
| Linux | contenedor `golang:1.27-alpine` arm64 sobre Dory, núcleo 6.12.30 |
| Fecha | 2026-09-10 |

Los cuatro candidatos, en las versiones exactas con las que se midió. Los cuatro llevan dentro la
**misma versión de SQLite, la 3.53.4**, así que ninguna diferencia de las que siguen es del motor:

| Candidato | Versión | Cómo lleva SQLite dentro | `cgo` |
|---|---|---|---|
| `github.com/mattn/go-sqlite3` | v1.14.52 | enlaza la biblioteca en C | sí |
| `modernc.org/sqlite` | v1.58.0, sobre `modernc.org/libc` v1.75.6 | el código de C traducido a Go | no |
| `zombiezen.com/go/sqlite` | v1.4.2 | el mismo motor traducido, con otra interfaz | no |
| `github.com/ncruces/go-sqlite3` | v0.35.4, con `go-sqlite3-wasm/v5` v5.0.35304 | SQLite compilado a WebAssembly, ejecutado con wazero | no |

**El tablero de prueba.** Lo genera `cmd/gen` con una semilla fija, así que sale igual cada vez:
390 tareas en la tabla, de las cuales **300 son las que `biso ls` lista** (las vivas y no
terminales, que es el filtro por defecto de [`biso ls`](../../docs/spec/cmd/ls.md)), con 168 dependencias, 990 criterios
de aceptación y de definición de terminado, 536 comentarios, y prosa en los cuatro campos largos.
Ocupa 1.441.792 bytes en modo WAL. La tabla es a propósito mayor que lo que la consulta devuelve,
porque un tablero real tiene tareas terminadas y archivadas.

**Los cuatro binarios hacen el mismo trabajo, y se comprueba antes de medir.** Abren la base de
datos, ejecutan las cinco consultas que `biso ls` necesita, calculan ["la urgencia"](../../docs/spec/modelo-de-datos/urgencia.md#la-urgencia),
ordenan con la tupla de [`biso ls`](../../docs/spec/cmd/ls.md), recortan al límite de 30 e imprimen las ocho columnas alineadas en
celdas de terminal. Sobre el mismo tablero los cuatro producen una salida **idéntica byte a byte**,
en macOS y en Linux, y el script se niega a dar la medida por buena si dejan de coincidir.

---

## 2. El reloj de pared, que es lo que fija el presupuesto

Milisegundos de reloj de pared del proceso completo, medidos desde fuera, con el `fork` y el `exec`
dentro. La columna que importa es la mediana.

**En Linux, que es la máquina de referencia de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque)** (contenedor, 200 repeticiones):

| Binario | mínimo | **mediana** | p90 | máximo |
|---|---|---|---|---|
| proceso vacío, sin controlador | 0,32 | **0,37** | 0,72 | 1,47 |
| `mattn` (C, con `cgo`) | 1,89 | **2,10** | 2,26 | 2,55 |
| `ncruces` (WebAssembly) | 2,32 | **2,47** | 2,77 | 6,67 |
| `zombiezen` (traducido) | 2,33 | **2,87** | 3,20 | 4,45 |
| `modernc` (traducido) | 2,70 | **3,15** | 3,37 | 4,02 |

**En macOS, que es la máquina de desarrollo** (300 repeticiones):

| Binario | mínimo | **mediana** | p90 | máximo |
|---|---|---|---|---|
| proceso vacío, sin controlador | 4,84 | **8,07** | 9,87 | 18,04 |
| `mattn` (C, con `cgo`) | 7,67 | **10,99** | 12,32 | 20,92 |
| `ncruces` (WebAssembly) | 8,27 | **11,55** | 12,97 | 20,44 |
| `zombiezen` (traducido) | 10,54 | **14,10** | 15,64 | 18,93 |
| `modernc` (traducido) | 10,78 | **14,51** | 16,04 | 23,49 |

Tres cosas que estas dos tablas dicen y conviene leer despacio.

**La primera: el suelo del sistema se come casi todo.** En macOS, arrancar un binario de Go que no
hace absolutamente nada cuesta 8,1 milisegundos, de modo que de los 14,5 milisegundos del candidato
que se recomienda solo 6,4 son atribuibles a leer el tablero. En Linux ese suelo es de 0,37
milisegundos, veintidós veces menos. Esto **confirma con más fuerza el argumento aritmético de
["La decisión de persistencia"](../../docs/decisiones/persistencia.md#la-decisión-de-persistencia) contra tener un daemon**: un daemon solo puede ahorrar lo que hay por encima del suelo, y
en la máquina donde el suelo es 8,1 de 14,5 no hay casi nada que ahorrar.

**La segunda: la cifra de 5,2 milisegundos que ["El origen de la cifra de 25 milisegundos"](../../docs/decisiones/persistencia.md#el-origen-de-la-cifra-de-25-milisegundos) da como "el suelo del sistema
operativo" es una cifra de macOS, no una constante.** Aquí ese suelo sale en 8,1 de mediana y 4,8
de mínimo, del mismo orden que lo que esa misma sección midió, pero en Linux es 0,37. Eso no invalida el
presupuesto de 25 milisegundos, porque un presupuesto holgado sigue siendo holgado, pero sí hay que
dejar de leerlo como una propiedad de los procesos y empezar a leerlo como una propiedad de este
portátil.

**Y la tercera: en macOS el margen de la cola no es el margen de la mediana.** El peor caso de
`modernc` en la tabla de arriba es de 23,5 milisegundos, el 94 por ciento del presupuesto, y con
`--all`, imprimiendo las 300 filas, su peor caso es de 34,7, por encima. Son valores extremos de una
máquina cargada y con un agente de seguridad interceptando cada `exec`, no una regresión, y la propia
sección de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque) lo dice: quien corre la suite en una máquina distinta de la de referencia no debe leer
el resultado como una afirmación sobre `biso`. Pero explica por qué la cifra que se lee en macOS es
la mediana y por qué la máquina de referencia no es esta.

### El tablero de 3.000 tareas, diez veces el de referencia

No entra en el presupuesto, que habla de 300, pero dice cómo escala y por tanto cuándo habría que
volver a mirar. Medianas:

| Candidato | Linux | macOS |
|---|---|---|
| `mattn` | 12,89 | 23,79 |
| `zombiezen` | 18,77 | 28,75 |
| `ncruces` | 19,37 | 29,93 |
| `modernc` | 22,20 | 31,42 |

El coste crece de forma esencialmente lineal, como ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque) anticipa. Con diez veces las
tareas, los cuatro siguen dentro del presupuesto en Linux y ninguno lo cumple ya en macOS. Es decir
que el margen de "tres veces" del que hablaba ["El origen de la cifra de 25 milisegundos"](../../docs/decisiones/persistencia.md#el-origen-de-la-cifra-de-25-milisegundos) se conserva en la máquina de referencia
y no en el portátil, y que un tablero de miles de tareas obligaría a mirar esto otra vez.

---

## 3. Dónde se va el tiempo: no en la consulta

Aquí está el hallazgo que explica toda la tabla anterior, y no estaba donde se esperaba.

**Dentro del proceso, los cuatro son casi iguales.** Mediana de 50 ejecuciones en macOS, en
milisegundos, medida desde dentro de `main`:

| Candidato | abrir la base de datos | las cinco consultas | urgencia, orden y formato | total dentro de `main` |
|---|---|---|---|---|
| `mattn` | 0,507 | 1,170 | 0,121 | 1,837 |
| `zombiezen` | 0,856 | 1,166 | 0,121 | 2,181 |
| `modernc` | 0,843 | 1,407 | 0,117 | 2,396 |
| `ncruces` | 0,945 | 1,436 | 0,122 | 2,518 |

Leer las 300 tareas cuesta entre 1,17 y 1,44 milisegundos, y entre el más rápido y el más lento hay
**0,27 milisegundos**, el 1,1 por ciento del presupuesto. El total dentro de `main` los separa 0,68
milisegundos. Pero en el reloj de pared de macOS los separaban **3,5**. Faltaban casi tres
milisegundos por explicar, y no estaban dentro de `main`.

**Estaban antes de `main`, en el `init` de un paquete.** Con `GODEBUG=inittrace=1`, en macOS:

| Candidato | suma de todos los `init` | el paquete que se lo lleva | memoria que reserva al arrancar |
|---|---|---|---|
| `mattn` | 0,16 ms | ninguno destacable | 8 KB |
| `ncruces` | 0,13 ms | ninguno destacable | 8 KB |
| `zombiezen` | 3,19 ms | `modernc.org/libc/honnef.co/go/netdb`, 2,6 ms | **3,66 MB en 44.061 reservas** |
| `modernc` | 2,79 ms | `modernc.org/libc/honnef.co/go/netdb`, 2,6 ms | **3,66 MB en 44.061 reservas** |

Entre pasadas esas cifras se mueven unas décimas, y en las tres que se hicieron el paquete de la
cuarta columna siempre estuvo entre 2,5 y 3,5 milisegundos.

Ese paquete lee `/etc/protocols` y `/etc/services` del disco y los parsea a estructuras de Go, en
el `init`, para que la traducción de `libc` pueda ofrecer `getservbyname`. En esta máquina
`/etc/services` tiene 13.926 líneas. Es decir que **`biso ls`, un programa que no abre un socket en
su vida, gastaría entre el 10 y el 13 por ciento de su presupuesto entero parseando la lista de
puertos de Internet antes de mirar el tablero**, en cada una de las muchas invocaciones que un
agente hace a lo largo de una sesión. No hay ninguna bandera ni etiqueta de compilación que lo
apague: `libc_darwin.go` importa ese paquete sin condiciones.

**Y en Linux no existe.** Esta es la razón por la que había que medir en Linux y no extrapolar. El
mismo paquete no aparece en la traza, y el `init` más caro de `modernc` pasa a ser de 0,046
milisegundos. La causa es que `libc_linux.go` no importa ese paquete, y no que el fichero falte: la
imagen donde se midió tiene su `/etc/services`, con 361 líneas. Comprobado ejecutando los binarios
dentro del contenedor, no deducido leyendo el código:

| Candidato | `init` más caro en macOS | `init` más caro en Linux |
|---|---|---|
| `modernc` | 2,6 ms (`netdb`) | 0,046 ms (`libc`) |
| `zombiezen` | 2,6 ms (`netdb`) | 0,047 ms (`libc`) |
| `mattn` | 0,034 ms (`os`) | 0,021 ms (`runtime`) |
| `ncruces` | 0,039 ms (`os`) | 0,016 ms (`runtime`) |

**Dicho de otra forma: el peaje de los dos candidatos traducidos es un peaje de macOS, y la máquina
de referencia no lo paga.** En el portátil sí, y eso es lo que hay que apuntar para no volver a
descubrirlo.

Un detalle de `ncruces` que la tabla no cuenta sola: no tiene `init` caro porque su coste está en
otro sitio, en preparar el módulo de WebAssembly al abrir la primera conexión, y por eso es el que
tiene la fase `open` más lenta de los cuatro, 0,945 milisegundos. El coste no desaparece, se mueve.

---

## 4. La distribución del binario, que es lo que de verdad decide

Esta es la sección que cierra la decisión, y su resultado no es un número sino un sí y un no.

**Compilar desde esta máquina para otra plataforma, con lo que ya hay instalado:**

| Candidato | `linux/amd64` | `linux/arm64` | `windows/amd64` |
|---|---|---|---|
| `modernc` | sí, 10,0 MB | sí, 9,7 MB | sí, 10,1 MB |
| `zombiezen` | sí, 9,5 MB | sí, 9,3 MB | sí, 9,6 MB |
| `ncruces` | sí, 16,4 MB | sí, 15,5 MB | sí, 16,6 MB |
| `mattn` con `cgo` | **falla** | no probado | **falla** |
| `mattn` sin `cgo` | compila, pero no funciona | compila, pero no funciona | compila, pero no funciona |

Las dos filas de `mattn` hay que leerlas juntas, porque por separado engañan.

Con `CGO_ENABLED=1`, que es lo que ese controlador necesita para hacer algo, la compilación cruzada
**falla en `runtime/cgo`**: hace falta un compilador de C que produzca código para el sistema
destino, y esta máquina solo tiene el de Apple para macOS. No es un obstáculo insalvable, hay
cadenas de herramientas cruzadas y hay contenedores, pero es exactamente lo que ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go)
llamaba "complica generar binarios para otras plataformas", y ahora está comprobado en vez de
supuesto.

Con `CGO_ENABLED=0` pasa algo peor, y es el argumento más fuerte de todo este documento. **El
binario compila, enlaza y arranca**, y falla al primer contacto con la base de datos con este
mensaje:

```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

Un fallo que la compilación deja pasar y que solo aparece al ejecutar es el peor de los desenlaces
posibles para una herramienta que se distribuye como binario: no lo ve quien la construye, lo ve
quien la usa. Y es un pie fácil de pisar, porque `CGO_ENABLED=0` es lo que uno pone precisamente
para que un binario sea estático.

**Y el binario de C, incluso construido en su plataforma, no es estático.** Compilado dentro del
contenedor de Alpine, `ldd` dice que depende de `libc.musl-aarch64.so.1`, así que ese binario no
corre en un sistema con glibc. Los otros tres, al pasarles `ldd`, contestan `Not a valid dynamic
program`, que es la forma que tiene `ldd` de decir que no hay nada que resolver. Un binario de Go
puro es un fichero que se copia y funciona; el de `cgo` se lleva consigo la pregunta de contra qué
`libc` se enlazó.

En macOS esta diferencia no se ve, y por eso hubo que ir a Linux: `otool -L` da la misma respuesta
para los cuatro, `libSystem.B.dylib` y `libresolv.9.dylib`, que es lo mismo que necesita un
programa de Go que no importa nada.

### Tamaño del binario y tiempo de compilación

Binario para macOS arm64, y compilación en esta máquina. La columna "en caliente" es tocar un
fichero del propio programa y volver a compilar, que es la vuelta que paga un agente en cada
iteración:

| Candidato | tamaño | compilar en frío | compilar en caliente |
|---|---|---|---|
| `mattn` | 7,19 MB | 14,25 s | 0,19 s |
| `zombiezen` | 9,55 MB | 5,66 s | 0,19 s |
| `modernc` | 10,07 MB | 5,53 s | 0,19 s |
| `ncruces` | 14,93 MB | 6,58 s | 0,15 s |
| referencia: sin controlador | 2,08 MB | | |

`mattn` da el binario más pequeño y tarda **dos veces y media más en compilar en frío**, porque hay
que compilar la biblioteca de SQLite en C, que son 269.653 líneas en un solo fichero. En caliente los cuatro
son iguales, porque ninguno recompila la dependencia. ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go) escogió Go frente a Rust
poniendo el ciclo de desarrollo por delante del rendimiento, y ese mismo criterio aquí penaliza a
`cgo` en la primera compilación de cada máquina y de cada `run` de integración continua, no en la
del día a día.

Sobre el tamaño, esa misma sección ya dijo lo que había que decir: "El binario es mayor que el
equivalente en Rust, lo que da igual en algo que se instala una vez". Ninguno de estos cuatro
números cambia esa frase.

---

## 5. Lo que cada controlador soporta

Comprobado ejecutando las consultas de verdad contra cada uno, no leyendo su documentación. Está en
`internal/features/features.go`, que además dice de dónde sale la exigencia de cada línea.

**Los cuatro soportan todo lo que el esquema y la especificación dan por hecho**, y lo hacen con la
misma versión de SQLite, la 3.53.4: modo WAL, `BEGIN IMMEDIATE` como acceso exclusivo de escritura
de ["Concurrencia, atomicidad y garantías observables"](../../docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables), puntos de retorno, vuelta atrás de una transacción, `busy_timeout`, claves
ajenas, `user_version` para versionar el esquema, `PRAGMA integrity_check` que `biso doctor`
necesita desde ["La decisión de persistencia"](../../docs/decisiones/persistencia.md#la-decisión-de-persistencia), `wal_checkpoint(TRUNCATE)`, `ANALYZE`, consultas recursivas para el
árbol de `parent` y el grafo de dependencias, el módulo JSON y las funciones de ventana. Y los
cuatro dejan leer desde una segunda conexión mientras una primera tiene una escritura abierta, que
es literalmente lo que esa misma sección de garantías promete.

Dos diferencias que sí hay, y ninguna decide:

**FTS5, que sería la vía barata para ["la búsqueda por texto"](../../docs/spec/referencias.md#la-búsqueda-por-texto), no está puesta en
todos por defecto.** `modernc` y `zombiezen` la traen compilada. `mattn` no, y se activa con
`go build -tags sqlite_fts5`, que cuesta 540 KB más de binario y **nada medible en reloj**: 10,19
milisegundos de mediana con y sin ella. `ncruces` tampoco la trae activa, y la ofrece en su paquete
`ext/fts5` con un registro por conexión, que es lo más incómodo de los tres caminos porque hay que
llegar a la conexión en crudo por debajo de `database/sql`.

**Ninguno de los cuatro hace `LIKE` insensible a mayúsculas con letras acentuadas**: los cuatro
contestan falso a `'Ábaco' LIKE 'ábaco'`, porque es lo que hace SQLite sin ICU. No es una diferencia
entre candidatos, es un dato para quien implemente ["la búsqueda por texto"](../../docs/spec/referencias.md#la-búsqueda-por-texto): no puede confiar
en `LIKE` para texto no ASCII con ninguno de ellos.

---

## 6. Lo que costó escribir cada prueba

Esta sección no tiene cifras de reloj: es lo que se aprendió al programar de verdad las cuatro
pruebas, que es un dato del mismo tipo que el que ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go) usó para elegir Go frente a Rust.
Como allí, lo que se mide es el coste de escribir el programa, no el de ejecutarlo.

**Los tres que hablan por `database/sql` no son tres implementaciones, son una.** El cargador
`internal/board/loadsql.go`, que recorre las cinco consultas y llena el modelo, lo comparten
`mattn`, `modernc` y `ncruces` sin una sola condición dentro. Entre los tres binarios lo único que
cambia es **una línea de `import` y la cadena con el nombre del controlador**, y en un caso ni eso,
porque `mattn` y `ncruces` se registran los dos como `sqlite3`. Compiló a la primera, y los tres
dieron la misma salida byte a byte en la primera ejecución, sin ajustar nada. Eso es la mejor
noticia de todo el ejercicio, porque significa que la decisión que este documento toma es
reversible por casi nada.

**`zombiezen` fue el único que costó trabajo, y el trabajo fue duplicar.** No implementa
`database/sql`, así que hubo que escribirle un recorrido propio con `sqlitex.Execute` y una función
de vuelta por fila, unas 110 líneas que no comparte con nadie, y después una segunda copia de las
comprobaciones de soporte, otras 170 líneas, porque tampoco ahí se podía reutilizar nada. Su
interfaz tiene virtudes reales: es explícita, y preguntar `ColumnType(i) != TypeNull` es más honesto
que envolver un entero en `sql.NullInt64` para averiguar lo mismo. Pero el precio es que cada pieza
auxiliar te la escribes tú, y que la salida de emergencia hacia otro controlador deja de ser una
línea y pasa a ser una reescritura. Para un programa cuya especificación ordena comprobar a menudo,
eso pesa.

**Lo que sí costó tiempo no fue ningún controlador, fueron tres detalles del entorno.** El primero:
en modo WAL los datos se quedan en el fichero `-wal`, así que el tablero generado medía 4 KB hasta
que se añadió `PRAGMA wal_checkpoint(TRUNCATE)` al final del generador, y sin eso la medida habría
comparado leer un fichero vacío. El segundo: `mattn` y `ncruces` registran los dos el nombre
`sqlite3` en `database/sql`, de modo que dos controladores no caben en el mismo binario y el segundo
en registrarse hace `panic`; por eso las comprobaciones de soporte son un modo de cada binario de
medida y no un programa aparte. El tercero: `ncruces` avisa por la salida de error si importas su
paquete `embed`, que es lo que su documentación antigua pedía, y ese aviso ensuciaba la medida.

**Y lo que más sorprendió: ninguno de los cuatro necesitó una sola línea de SQL distinta.** Mismo
esquema, mismas cinco consultas, mismos `Scan`. La incompatibilidad que uno teme al elegir un
controlador de SQLite no apareció por ningún lado, y eso es en sí mismo el argumento para quedarse
en `database/sql`.

**El coste de `cgo` no estuvo en escribir el código, que fue idéntico, sino en todo lo demás.**
Compilar en frío tarda dos veces y media más, la compilación cruzada no existe, y su modo
degradado produce un binario que engaña. Ninguna de esas tres cosas se ve mirando el `main.go`,
que es exactamente por lo que había que medirlas.

## 7. La decisión

**El controlador es `modernc.org/sqlite`, sobre `database/sql`, sin `cgo`.** Y el orden de los
motivos importa, porque el primero descarta una familia entera y los demás solo eligen dentro de la
que queda.

**Primero: `cgo` queda descartado por la distribución, no por la velocidad.** Es el más rápido de
los cuatro en las dos plataformas, y aun así se descarta, porque desde esta máquina no produce un
binario para Linux ni para Windows, y porque su modo sin `cgo` compila un binario que falla al
ejecutarse en vez de fallar al construirse. Para una herramienta que se distribuye como binario y
que se ejecuta cientos de veces por sesión, un fallo que solo se ve en la máquina del usuario pesa
más que el milisegundo que gana en Linux.

**Segundo: entre los tres que quedan, el rendimiento tampoco decide.** En la máquina de referencia
están entre 2,47 y 3,15 milisegundos, los tres casi ocho veces por debajo del presupuesto, y los
separa 0,68 milisegundos, que es el 2,7 por ciento de 25. Es la misma situación que ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go)
describió al elegir el lenguaje, y merece la misma respuesta: cuando el rendimiento ya no distingue, decide lo
que cuesta escribir y mantener el programa.

**Tercero: por ese criterio gana `modernc`.** Es el único de los tres que es a la vez la interfaz
estándar y el motor que los otros usan. `zombiezen` envuelve a `modernc`, así que elegirlo es
elegir el mismo motor con una interfaz propia que obliga a escribir a mano todo lo que
`database/sql` da hecho, y su ganancia son 0,28 milisegundos en Linux. `ncruces` es el más rápido de
los tres y no tiene el peaje de arranque, pero cuesta un binario un 48 por ciento mayor, y su FTS5
hay que registrarla por conexión bajando por debajo de `database/sql`.

**Y esto es lo que se acepta a cambio, dicho claro, porque tiene nombre y número.** En macOS, y solo
en macOS, `modernc` gasta entre 2,6 y 3,5 milisegundos antes de `main` parseando `/etc/services` en
el `init` de `modernc.org/libc/honnef.co/go/netdb`, con 3,66 MB en 44.061 reservas de memoria, para
una funcionalidad de red que `biso` no usa nunca y que no se puede desactivar. Es entre el 10 y el 14
por ciento del presupuesto, tirado. Se acepta porque la máquina de referencia es Linux y allí ese
`init` no existe, porque incluso pagándolo la mediana en macOS es de 14,5 milisegundos sobre un
presupuesto de 25, y porque el remedio no está en `biso` sino aguas arriba.

**Lo que haría cambiar la decisión, dicho por adelantado para que no haya que volver a discutirlo
desde cero.** Si el peaje de arranque en macOS llegase a molestar de verdad, el relevo es `ncruces`,
y el cambio cuesta **una línea de `import` y una cadena con el nombre del controlador**, porque los
dos hablan por `database/sql` y todo el código de `biso` que consulta el tablero seguiría siendo el
mismo. Eso es también la razón para no elegir `zombiezen`: no es que sea peor, es que salirse de
`database/sql` convierte esa salida de emergencia de una línea en una reescritura.

---

## 8. Los candidatos que se descartaron sin medir

Se comprobó que estuvieran vivos antes de descartarlos, porque descartar por antigüedad sin mirar la
fecha es lo mismo que no mirar.

| Candidato | Última versión | Por qué no se mide |
|---|---|---|
| `github.com/glebarez/go-sqlite` | v1.23.0, 2026-08-06 | Está vivo, pero es un fork de `modernc.org/sqlite` empaquetado para GORM. Es el mismo motor y el mismo `libc`, así que mediría lo mismo que `modernc` y traería un intermediario más. |
| `github.com/tailscale/sqlite` | 2026-03-06 | Vivo, pero usa `cgo` igual que `mattn` y añade instrumentación pensada para el producto de Tailscale. Hereda entera la razón por la que `cgo` queda descartado. |
| `github.com/eatonphil/gosqlite` | v0.10.0, 2024-08-11 | Usa `cgo`, y además es una interfaz de bajo nivel sin `database/sql`, así que junta las dos cosas que aquí se descartan. |
| `crawshaw.io/sqlite` | v0.3.2, 2020-06-07 | Sin publicar nada en más de seis años. `zombiezen.com/go/sqlite`, que sí se mide, es su continuación. |

No hay ningún quinto camino: o se enlaza la biblioteca en C, o se traduce su código a Go, o se
compila a WebAssembly y se ejecuta con un motor de WebAssembly escrito en Go. Los cuatro candidatos medidos
cubren los tres.
