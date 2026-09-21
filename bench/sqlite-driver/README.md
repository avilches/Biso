# Banco de pruebas: con qué controlador habla `biso` con SQLite

> **Este banco es una fotografía de un momento, no una herramienta mantenida.** Se midió el
> 2026-09-10 con el modelo de datos de entonces, y desde entonces el programa retiró tres campos
> que el banco todavía guarda: `ext`, `documentation` y `modifiedFiles`. El contenido de los dos
> últimos, cuando lo hay, vive ahora en `references` (ver ["Se retira `documentation` y `references`
> queda como único campo de
> punteros"](../../docs/decisiones/detalles.md#se-retira-documentation-y-references-queda-como-único-campo-de-punteros)
> y ["Se retira `modifiedFiles`"](../../docs/decisiones/detalles.md#se-retira-modifiedfiles)); `ext`
> no tiene sustituto directo (["Se retira `ext`"](../../docs/decisiones/detalles.md#se-retira-ext)). El esquema
> y el generador de tableros los conservan a propósito, y las cifras de `RESULTADOS.md` y las que cita
> ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go)
> son las de ese modelo. La razón está en ["Por qué se deja como está"](#por-qué-se-deja-como-está).

Este directorio contesta con números la única decisión que quedaba antes de escribir código,
aplazada a propósito hasta tener datos: cómo habla el programa con SQLite.
El resultado, con su tabla y su recomendación, está en
[`RESULTADOS.md`](RESULTADOS.md), y su contenido está escrito para acabar en ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go).

Lo que se mide es el presupuesto de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque): 25 milisegundos de reloj
para `biso ls` sobre un tablero de 300 tareas.

## Por qué se deja como está

**La decisión vigente.** El banco no se actualiza al modelo de datos actual: se conserva tal como
produjo sus cifras, con el esquema y el generador de entonces, y este README lo declara arriba del
todo. Esa es la regla para los tres campos retirados por igual (`ext`, `documentation` y
`modifiedFiles`), sin distinguir entre ellos. Si algún día hay que volver a medir, se actualiza el
esquema a la vez que se rehacen las cifras y su decisión, nunca solo lo primero.

**Cuánto pesan esos campos en lo que se midió.** Se regeneró el tablero de 300 tareas con el
generador sin tocar y salió con las mismas filas y el mismo tamaño que el de
`results/01-tableros.txt` (390 tareas, 1.441.792 bytes),
así que los tamaños son los medidos. Los porcentajes son del orden de lo que pesan, no cifras
exactas: el de las dos filas de abajo sale de repartir el tamaño de `textlist` en proporción al
número de filas (804 de 1.197), porque una tabla no se parte por filas a nivel de página:

| Contenido | Filas | Bytes en el fichero | Parte del tablero |
|---|---|---|---|
| tabla `textlist` entera (`references`, `documentation` y `modifiedFiles`) | 1.197 | 73.728 | 5,1 % |
| de ella, las filas de `documentation` y `modifiedFiles` | 804 | unos 49.000 | unos 3,4 % |
| tabla `ext` | 106 | 4.096 | 0,3 % |

Y lo que se mide, `biso ls`, no lee ninguna de las dos tablas: las cinco consultas de
`internal/board/schema.go` recorren `task`, `assignee`, `criterion`, `dependency` y `board`, porque
la primera regla de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque)
prohíbe leer lo que la invocación no imprime. Las cifras de reloj, de fases, de `init`, de enlazado
y de compilación no dependen de esos campos; lo único que se movería es el tamaño del fichero del
tablero, en unos pocos puntos por ciento.

**Alternativa descartada: actualizar el generador y remedir.** Quitar los tres campos es fácil, pero
no cuesta lo mismo que editar dos líneas: el generador saca todos sus valores de un único generador
pseudoaleatorio con semilla fija, y cada llamada que se quita desplaza todas las siguientes, así que
saldría un tablero distinto tarea a tarea. Habría que repetir entera la medición en macOS y en Linux
(este último dentro de un contenedor), reescribir las tablas de `RESULTADOS.md` y revisar las cifras
que cita ["El lenguaje de implementación es Go"](../../docs/decisiones/lenguaje-y-rendimiento.md#el-lenguaje-de-implementación-es-go),
todo para llegar a la misma recomendación de controlador, porque nada de lo que decide el veredicto
depende de esos campos. Se descarta por desproporcionada; no porque no sea posible.

**Alternativa descartada: actualizar el esquema sin remedir.** Dejaría unas cifras medidas con un
modelo bajo un README que dice otro, que es peor que una fotografía honesta.

## Cómo se vuelve a ejecutar

Dos órdenes, y ninguna necesita argumentos:

```sh
./run.sh      # todo lo que se mide en macOS
./linux.sh    # lo mismo en Linux, dentro de un contenedor
```

`run.sh` deja sus resultados en `results/`, un fichero por bloque de medida, y no toca nada
fuera de este directorio: los binarios van a `bin/`, los tableros generados a `work/`, y las dos
carpetas están en el `.gitignore` porque se rehacen solas. La medida en frío usa una caché de
compilación propia dentro de `work/`, así que no borra la caché del usuario.

`linux.sh` hace falta porque la máquina de referencia de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque) es la que ejecuta la
integración continua, y esa no es macOS. No es una formalidad: uno de los controladores paga en
macOS un peaje de arranque que en Linux no existe, y sin verlo el veredicto sale al revés.
Necesita un Docker que responda, que en esta máquina es Dory, y compila también dentro del
contenedor, que es la única forma de medir allí el controlador de C.

Para una pasada rápida mientras se trastea, las repeticiones se pueden bajar:

```sh
RUNS=30 WARMUP=5 PHASE_RUNS=10 ./run.sh
```

Y para volver a mirar solo una cosa, cada pieza vale por separado:

```sh
./inittrace.sh                             # qué cuesta el init de cada paquete
./bin/gen -out work/board.db -tasks 300    # regenerar el tablero
./bin/ls-modernc work/board.db             # ejecutar un candidato a mano
BENCH_PHASES=1 ./bin/ls-modernc work/board.db    # con el desglose por fases
BENCH_FEATURES=1 ./bin/ls-modernc                # qué soporta ese controlador
```

Si mañana cambia la versión de Go o de un controlador, la medición se rehace con `go get`,
`go mod tidy` y otra vez `./run.sh` y `./linux.sh`. Las versiones exactas con las que se midió
quedan escritas en `results/00-entorno.txt`, así que la comparación no se pierde.

## Qué se compara

Cuatro candidatos, que son todos los que están vivos:

| Candidato | Cómo lleva SQLite dentro | Necesita `cgo` |
|---|---|---|
| `github.com/mattn/go-sqlite3` | enlaza la biblioteca en C | sí |
| `modernc.org/sqlite` | traducción del código de C a Go | no |
| `zombiezen.com/go/sqlite` | el motor traducido de arriba, con otra interfaz | no |
| `github.com/ncruces/go-sqlite3` | SQLite compilado a WebAssembly, ejecutado con wazero | no |

Los descartados sin medir, y por qué, están en la última sección de
[`RESULTADOS.md`](RESULTADOS.md).

## Cómo está montado

La honestidad de la medida depende de una cosa: que los cuatro binarios hagan **el mismo
trabajo**, y que lo único distinto sea hablar con la base de datos. Por eso el código está
partido así:

- `internal/board/` es el trabajo de verdad, y lo comparten los cuatro: el modelo de tarea
  reducido a lo que `biso ls` imprime, la fórmula de ["La urgencia"](../../docs/spec/modelo-de-datos/urgencia.md#la-urgencia), la
  regla de orden y el formato de las ocho columnas de [`biso ls`](../../docs/spec/cmd/ls.md), medido en celdas de terminal.
- `internal/board/schema.go` tiene el esquema del tablero y las **cinco consultas** que
  `biso ls` necesita, y ni una más: la primera regla de ["El presupuesto de arranque"](../../docs/spec/presupuestos.md#el-presupuesto-de-arranque) prohíbe leer lo que la
  invocación no va a imprimir, así que ahí no se leen comentarios, ni etiquetas, ni prosa.
- `internal/board/loadsql.go` recorre esas cinco consultas por `database/sql`, y lo comparten
  los tres controladores que ofrecen esa interfaz. Entre ellos, lo único que cambia es una línea
  de `import` y el nombre del controlador.
- `cmd/ls-zombiezen/` es el único que no puede reutilizarlo, porque ese controlador no implementa
  `database/sql`. Tiene su propio recorrido escrito con `sqlitex`, y eso ya es un dato sobre lo
  que cuesta usarlo.

`run.sh` comprueba antes de medir que los cuatro producen una salida **idéntica byte a byte**
sobre el mismo tablero. Si dejaran de coincidir, la comparación no valdría y el script lo dice.

## Qué mide cada fichero de `results/`

| Fichero | Qué contiene |
|---|---|
| `00-entorno.txt` | la máquina, la versión de Go y la versión exacta de cada controlador |
| `01-tableros.txt` | el tablero generado: cuántas tareas tiene y cuántas lista `biso ls` |
| `02-salida-identica.txt` | la comprobación de que los cuatro dan la misma salida |
| `03-reloj-de-pared.txt` | el reloj de pared del proceso completo, medido desde fuera |
| `04-fases.txt` | cuánto de eso es abrir la base de datos y cuánto consultarla |
| `05-init.txt` | lo que cuesta el `init` de cada paquete antes de llegar a `main` |
| `06-tamano-y-enlazado.txt` | tamaño del binario y bibliotecas del sistema de las que depende |
| `07-compilacion.txt` | compilar en frío y volver a compilar tras tocar un fichero |
| `08-compilacion-cruzada.txt` | si sale un binario para Linux y para Windows desde aquí |
| `09-soporte.txt` | si el controlador soporta lo que el esquema da por hecho |
| `10-linux.txt` | todo lo anterior repetido en Linux, que lo escribe `linux.sh` |

## Las tres herramientas de medida, y por qué son estas

**`cmd/timeit`** mide el reloj de pared del proceso completo desde fuera: lanza el binario,
espera a que termine y se queda con la distribución. Es el sustituto de `hyperfine`, que no está
instalado en esta máquina. Lo importante es que mide desde fuera, con el `fork` y el `exec`
dentro, porque un `Benchmark` de Go mediría un proceso que ya arrancó y el arranque es
justamente la mitad de lo que aquí se discute. Imprime la mediana y la dispersión, y también el
mínimo, que en una máquina con otras cosas encima es el estimador menos contaminado.

**`internal/phase`** parte el interior del proceso en abrir, consultar, calcular y escribir. Se
enciende con `BENCH_PHASES=1` y no cuesta nada cuando está apagado.

**`inittrace.sh`** usa el `GODEBUG=inittrace=1` del runtime de Go, y es la pieza sin la que la
medida no se habría entendido. El medidor de fases arranca ya dentro de `main`, así que no puede
ver nada de lo que pasa antes, y resulta que ahí es donde se va la mayor parte de la diferencia
entre los candidatos.

## Una advertencia sobre las cifras

La máquina donde se midió no estaba en silencio, y lo tenía encima un agente de seguridad que
intercepta cada `exec`. Eso infla el suelo de arrancar un proceso y ensancha la cola de los
tiempos: en macOS arrancar un binario de Go que no hace nada sale en 8,1 milisegundos de mediana,
mientras que dentro de un contenedor de Linux en la misma máquina sale en 0,37. La comparación
**entre los cuatro candidatos** sigue valiendo, porque los cuatro pagan el mismo suelo, y por eso
las tablas de `RESULTADOS.md` dan siempre el suelo medido al lado. Lo que no vale es leer una
cifra suelta de macOS como si fuera lo que costará en la máquina de referencia.
