# El lenguaje de implementacion y el rendimiento

## El lenguaje de implementación es Go

**Los dos candidatos reales eran Go y Rust**, y la elección es Go. Los dos cumplen con holgura el
presupuesto del apartado anterior en la única carga que hay medida, y los separa menos de un
milisegundo, así que la decisión no se toma por rendimiento: se toma por lo que cuesta escribir el
programa, porque es lo único que de verdad los separa aquí.

**Por rendimiento la diferencia son seis décimas de milisegundo.** Con las medidas de
["El coste de arranque y el coste de contexto"](../estado-del-arte/catalogo-de-problemas.md#12-el-coste-de-arranque-y-el-coste-de-contexto), sobre el suelo de 5,2 milisegundos que cuesta arrancar cualquier proceso,
Rust añade 1,6 milisegundos y Go 2,2. Esa diferencia es el **2,4 por ciento** de un presupuesto de 25
milisegundos que sobra tres veces sobre el total medido de 8,7, que es lo que cuesta leer 300 tareas de
un JSON y no lo que costará leerlas de SQLite (["El origen de la cifra de 25 milisegundos"](persistencia.md#el-origen-de-la-cifra-de-25-milisegundos)). El margen puede encogerse cuando se mida
el almacén de verdad, pero las seis décimas no dependen de eso: son el arranque del propio binario, la
misma cifra fija sea cual sea el trabajo que venga después. Para calibrar cuánto es: `rg`, que es
la herramienta más rápida de las que se midieron instaladas, tarda 7,1 milisegundos, y
`git --version` tarda 12,3. Ganar seis décimas en un programa cuyo competidor de referencia gasta doce
milisegundos en imprimir su propia versión no cambia nada que un usuario pueda notar. El presupuesto,
además, **se midió con un binario de Go**, así que la cifra que la especificación exige no se extrapola
de otro lenguaje: es lo que el lenguaje elegido hizo en esa máquina, con la carga que
["El origen de la cifra de 25 milisegundos"](persistencia.md#el-origen-de-la-cifra-de-25-milisegundos) dice.

**Lo que decide es el ciclo de desarrollo, y en particular el ciclo de un agente.** Este documento y
[`docs/spec/`](../spec/index.md) están escritos para que alguien implemente el programa entero sin preguntar, por pasos y
comprobando cada uno antes de seguir, y ese alguien va a ser en buena parte un
agente automático. En ese modo de trabajo, el coste dominante no es el tiempo de ejecución del programa
sino **el número de vueltas entre escribir y ver el resultado**, y ahí Go gana por dos motivos
distintos. El primero es que compila muy rápido, así que cada vuelta es corta. El segundo es más
importante: el modelo de propiedad y préstamo de memoria de Rust es exactamente la clase de cosa que un
modelo de lenguaje **no acierta a la primera**, y cada fallo obliga a una ronda de corrección que no
trata del problema que se está resolviendo sino de satisfacer al compilador. Esas rondas se suman, hacen
falta revisiones constantes, y no dejan nada mejor en el producto: el mismo programa, escrito en Go, no
las paga.

**El coste de elegir Go se acepta a ojos vistas y aquí queda escrito.** El recolector de basura, que es
la objeción habitual, no tiene ningún efecto medible en una herramienta cuyo proceso vive milisegundos,
lee 300 tareas y termina, porque no llega a haber presión de memoria que recoger. El binario es mayor
que el equivalente en Rust, lo que da igual en algo que se instala una vez. Y se renuncia a las seis
décimas del párrafo anterior, que es lo que se está comprando.

**Quedaba una cosa por comprobar, y no era del lenguaje sino de su encuentro con SQLite.** La
persistencia de ["La decisión de persistencia"](persistencia.md#la-decisión-de-persistencia) es una base de datos SQLite, y en Go hay tres formas de hablar con ella: un
enlace con la biblioteca en C, que obliga a compilar con `cgo`; una traducción de ese código de C a Go
puro; y SQLite compilado a WebAssembly y ejecutado por un motor escrito en Go. La elección afectaba al
arranque y a cómo se distribuye el programa, y nunca a la del lenguaje, porque las tres son de Go. **Ya
está medida, y la cierra el apartado siguiente.**

### El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`

Medido el 2026-09-10 con el banco de pruebas que vive en `bench/sqlite-driver/` de este mismo
repositorio, que se rehace con dos órdenes sin argumentos. Las tablas completas, la composición del
tablero de prueba y las versiones exactas de todo están en `bench/sqlite-driver/RESULTADOS.md`; aquí va
la decisión y lo que la sostiene. Se midieron los cuatro candidatos vivos, y los cuatro llevan dentro la
misma versión de SQLite, la 3.53.4, así que ninguna diferencia de las que siguen es del motor.

**El presupuesto se cumple con los cuatro, y con mucho margen.** Sobre el tablero de 300 tareas de
["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque), en Linux el más lento tarda 3,2 milisegundos, casi ocho veces por debajo de
los 25. En un portátil macOS van de 11 a 14,5 milisegundos, y de esos 8,1 son el suelo del sistema para
arrancar cualquier proceso. Los cuatro binarios producen además una salida idéntica byte a byte en las
dos plataformas, y ninguno necesitó una sola línea de SQL distinta: mismo esquema, mismas cinco
consultas, mismos recorridos. La incompatibilidad que uno teme al elegir un controlador de SQLite no
apareció por ningún lado, y eso es en sí mismo un argumento para quedarse en la interfaz estándar
`database/sql`.

**Así que el rendimiento no decide, y lo que decide es la distribución del binario.** Es el mismo
criterio con el que el apartado anterior eligió Go frente a Rust, aplicado otra vez, y aquí la
diferencia no es de grado.

**El enlace con la biblioteca en C queda descartado, y eso es lo primero, porque descarta una familia
entera.** Es el más rápido de los cuatro en las dos plataformas y aun así se cae, por dos cosas
medidas. La primera es que desde una máquina macOS no produce un binario para Linux ni para Windows,
porque la compilación cruzada falla en `runtime/cgo` al no haber un compilador de C para el sistema
destino. La segunda es peor que un fallo de compilación: si se apaga `cgo`, el binario compila, enlaza,
arranca y muere al tocar la base de datos con el mensaje `go-sqlite3 requires cgo to work. This is a
stub`. Un fallo que la construcción deja pasar y que solo aparece al ejecutar es el peor desenlace para
algo que se reparte como binario, porque no lo ve quien lo construye sino quien lo usa. Los otros tres
candidatos compilan para `linux/amd64`, `linux/arm64` y `windows/amd64` con un `go build` y nada más
instalado, y salen estáticos.

**Entre los tres que quedan gana `modernc`, y por lo que cuesta escribir el programa.** Los separan 0,68
milisegundos, que es el 2,7 por ciento del presupuesto, así que otra vez no decide el reloj. `modernc` es
a la vez la interfaz estándar y el motor sobre el que están construidos los otros dos candidatos de Go
puro. `zombiezen` es ese mismo motor con una interfaz propia que obliga a escribir a mano lo que
`database/sql` da hecho, y su ganancia son 0,28 milisegundos. `ncruces` es el más rápido de los tres,
pero su binario es un 48 por ciento mayor y su búsqueda por texto completo hay que registrarla por
conexión, bajando por debajo de `database/sql`.

**Y esto es lo que se acepta a cambio, con nombre y número.** En macOS, y solo en macOS, `modernc` gasta
entre 2,6 y 3,5 milisegundos antes de que `main` empiece, en el arranque del paquete
`modernc.org/libc/honnef.co/go/netdb`, que lee `/etc/services` y `/etc/protocols` del disco y los
convierte en estructuras de Go para ofrecer una función de red que `biso` no usa nunca y que no se puede
desactivar. Son entre el 10 y el 14 por ciento del presupuesto tirados en cada invocación. En Linux ese
arranque baja a 0,046 milisegundos, porque el paquete no se importa allí, y eso se comprobó ejecutando
los binarios dentro de un contenedor y no leyendo el código. Se acepta porque incluso pagándolo la
mediana en macOS es de 14,5 sobre 25, y porque el remedio no está en `biso` sino aguas arriba.

**La salida de emergencia queda dicha por adelantado, para no tener que discutirla desde cero.** Si ese
peaje llegase a molestar de verdad, el relevo es `ncruces`, y el cambio cuesta una línea de `import` y
la cadena con el nombre del controlador, porque los dos hablan por `database/sql` y el resto del código
que consulta el tablero seguiría siendo el mismo. Esa es también la razón de no elegir `zombiezen`: no
es que sea peor, es que salirse de `database/sql` convierte esa salida de emergencia de una línea en una
reescritura.

**Lo que se comprobó que soporta, porque la especificación lo da por hecho.** Modo WAL, `BEGIN
IMMEDIATE` como acceso exclusivo de escritura, puntos de retorno, `busy_timeout`, claves ajenas,
`user_version`, el `integrity_check` que necesita `biso doctor`, `wal_checkpoint(TRUNCATE)`, `ANALYZE`,
consultas recursivas, el módulo JSON y las funciones de ventana. Y leer desde una segunda conexión
mientras una primera tiene una escritura abierta, que es literalmente lo que promete
["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables). La búsqueda por texto completo con FTS5, que sería la vía barata para
["La búsqueda por texto"](../spec/referencias.md#la-búsqueda-por-texto), viene
puesta. Y un dato para cuando se implemente esa búsqueda: ningún controlador hace `LIKE` insensible a
mayúsculas con acentos, porque eso es lo que hace SQLite sin la biblioteca ICU.

## El presupuesto de arranque se mide con la muestra más rápida

**La prueba del presupuesto de ["El presupuesto de arranque"](../spec/presupuestos.md#el-presupuesto-de-arranque) sigue midiendo reloj, el del proceso
entero desde antes de lanzarlo hasta después de su salida, y su veredicto es la muestra más rápida de
las que toma, no la mediana.** Ejecuta `biso ls` y `biso prime`, cada uno hasta 30 veces, sobre el tablero de
300 tareas y pasa en cuanto una sola de esas ejecuciones termina en menos de 25 milisegundos; solo
falla si ninguna de las 30 lo consigue. Cada ejecución tiene además que terminar con código de salida
0, y quien lee el fallo ve el mínimo, la mediana y el máximo de las muestras tomadas, para distinguir
una regresión de una máquina desbordada. La cifra no cambia: siguen siendo 25 milisegundos de reloj,
y lo que se precisa es qué muestra de reloj es la que se compara con ella.

**Por qué el mínimo y no otro estadístico.** El ruido de una máquina ocupada, ya sea otro proceso
disputando el núcleo, el planificador o un agente de seguridad leyendo cada fichero que se abre, solo
puede añadir tiempo a una ejecución, nunca quitárselo. Ninguna ejecución termina antes de lo que el
programa necesita para hacer su trabajo, así que la más rápida es la mejor cota superior observable de
lo que cuesta el programa por sí mismo, y es la única que la carga puntual no puede empeorar mientras
haya una ejecución que se libre de ella. Una regresión real, en cambio, sube el mínimo igual que
cualquier otra muestra, porque el código nuevo lo paga cada ejecución. Es la propiedad que hace falta:
que la prueba dependa del código y no del estado de la máquina en ese instante. Salir en la primera
ejecución que cabe no cambia el veredicto y hace que la prueba cueste una sola ejecución por comando
cuando la máquina está en reposo.

**Lo medido el 2026-09-21 en la máquina de desarrollo** (Apple M3 Max de 16 núcleos, macOS, 40
ejecuciones por comando y por condición, tablero de 300 tareas con los cuatro bloques de `biso prime`
poblados). La carga artificial fueron procesos que giran sin parar, el doble y el cuádruple que
núcleos hay. Los milisegundos son de reloj y salen como mínimo, mediana y máximo:

| Condición | `biso ls` | `biso prime` |
|---|---|---|
| En reposo | 11,3 / 12,0 / 39,2 | 13,5 / 14,6 / 41,3 |
| 32 procesos ocupando la CPU | 13,8 / 19,2 / 30,3 | 19,1 / 27,5 / 89,5 |
| 64 procesos ocupando la CPU | 13,4 / 20,2 / 48,6 | 18,3 / 23,8 / 63,3 |

Aun en reposo hay ejecuciones sueltas a casi el triple de la mediana, y con carga la mediana de
`biso prime` llega a 27,5 y supera el presupuesto sin que el binario haya cambiado, que es exactamente
el fallo que motivó esta decisión. El mínimo de `biso prime`, en cambio, se queda entre 18 y 20 milisegundos con el
doble de procesos que núcleos y con el cuádruple: el ruido lo desplaza, pero no lo lleva a la cifra. Un
`biso prime` que costara 12 milisegundos más de lo que cuesta hoy pondría el mínimo en reposo por
encima de 25 y la prueba fallaría, con la máquina libre y con la máquina cargada.

**Los 30 se eligen porque hacen falta muy pocos para que un mínimo sea fiable y porque solo se pagan
cuando algo va mal.** En reposo la primera ejecución ya cabe y la prueba termina ahí. Si ninguna cabe
son 30 ejecuciones de unos 20 a 40 milisegundos por comando, poco más de un segundo, un coste que solo
se paga cuando la prueba va a fallar de todos modos.

**Lo que esta decisión no cubre, y se acepta.** Una máquina tan saturada que ninguna de las 30
ejecuciones cabe en 25 milisegundos hace fallar la prueba, y eso es correcto: en esa máquina y en ese
momento el programa no arranca dentro del presupuesto, y el mensaje con el mínimo, la mediana y el
máximo lo deja a la vista para que quien lo lee no tenga que adivinarlo. Y una regresión menor que el
hueco que hay entre el mínimo en reposo y 25 milisegundos no se detecta, pero eso es propio de tener
un presupuesto y no del estadístico: la cifra se vigila entera, no se vigila que no suba.

**Descartado: la mediana de pocas ejecuciones, que era lo que había.** Con cinco ejecuciones la mediana
es la tercera más rápida, así que basta con que la carga afecte a tres para que dé el veredicto. Bajo
carga mide la máquina, como enseña la tabla de arriba, y falla sin que haya cambiado una línea.

**Descartado: el tiempo de CPU (usuario más sistema) en lugar del reloj, y también como condición
añadida.** No ve las esperas: un `sleep`, una espera de disco, una sincronización a disco o una
llamada de red gastan reloj y no gastan CPU, así que una regresión que añadiera una espera de 30
milisegundos pasaría la prueba sin que nada la detectara, y el presupuesto existe justo para lo que
espera quien lanza el comando. Tampoco es inmune a la carga: en las mismas medidas la mediana de CPU de
`biso ls` pasó de 9,1 a 14,1 milisegundos por la contención de caché y de frecuencia. Y un proceso con
varios hilos puede gastar más CPU que reloj, lo que no coincide con lo que dice la especificación.
Como condición añadida exigiría además una segunda cifra que la especificación no tiene.

**Descartado: subir el umbral de la prueba.** Una prueba con un límite distinto del de la
especificación no comprueba la especificación, y un límite que nadie comprueba no es un límite. Es el
mismo razonamiento por el que la prueba se salta entera bajo el detector de carreras en vez de tener
un tope más holgado.

**Descartado: repetir la prueba entera cuando falla.** Es el mínimo de varias muestras hecho a mano y a
escondidas, con la diferencia de que entrena a quien la ve fallar a lanzarla otra vez hasta que pase y
no deja constancia de qué ejecuciones se tomaron. El mínimo lo hace una sola vez, dentro de la propia
prueba y con las muestras impresas.

**Descartado: saltar la prueba cuando la máquina está cargada.** Exigiría leer la carga del sistema,
que no se lee igual en todas las plataformas, y convertiría un fallo en un silencio: una máquina que
siempre está cargada no comprobaría nunca el presupuesto. Solo el detector de carreras justifica un
salto, porque ahí se conoce la causa y se declara.

**Descartado: un percentil bajo, como el décimo, en lugar del mínimo.** Protegería de una muestra
anormalmente baja, pero no existe: ninguna ejecución va más rápido de lo que el programa cuesta, y la
ejecución de calentamiento ya quita el efecto de un fichero leído por primera vez. Exigiría además
tomar todas las muestras siempre, sin poder salir en la primera que cabe.
