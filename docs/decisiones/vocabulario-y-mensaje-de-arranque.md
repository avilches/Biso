# El vocabulario y el mensaje de arranque

## La regla de coincidencia de vocabulario

Es la pieza que cierra el principio 1, y merece contarse entera porque tiene una trampa.

**Se eligió eliminar los separadores, no colapsarlos.** Normalizar es pasar a minúsculas, quitar
diacríticos y quitar espacios, guiones y guiones bajos; comparar es igualdad. Con esa regla, `todo`,
`To-Do`, `TO_DO` y `to do` son todos `To Do`.

La alternativa era colapsar cualquier tramo de separadores a un solo espacio, y **no funciona**:
colapsando, `To-Do` se convierte en `to do` y coincide, pero `todo` no tiene ningún separador que
colapsar, se queda en `todo`, y no coincide con `to do`. Es un error fácil de cometer al redactar la
regla en prosa, porque el ejemplo que uno escribe a continuación parece cierto y no lo es.

Se eligió eliminar por dos motivos. El primero es que es lo que hacen las herramientas existentes al
escribir, así que ningún texto que hoy vale deja de valer. El segundo es que produce una función más
simple de escribir y de probar.

**Por eso la regla está escrita en la especificación como pseudocódigo con pasos numerados y no como
prosa.** Escrita en prosa volvería a poder decir dos cosas a la vez, y ya lo hizo una vez.

Una consecuencia que conviene no perder: el ejemplo canónico de un valor inválido no puede ser `Todo`,
porque `Todo` es válido. En la especificación el ejemplo es `Pending`, que no existe en ningún
vocabulario.

---

## El algoritmo de sugerencias: Levenshtein con umbral

**Decisión vigente.** Las cinco sugerencias de "lo más parecido" que promete la especificación
(claves de `biso config`, nombres de comando de `biso help`, etiquetas y personas de `--label` /
`--label-or` / `--assignee`) usan la distancia de Levenshtein sobre la forma que ya produce
`normalizar()`, con un umbral de corte: solo entran las candidatas cuya distancia no supere la
mitad de la longitud del valor normalizado de entrada, redondeada hacia arriba. De las que pasan
se toman como máximo `N` (tres para claves y comandos, cinco para etiquetas y personas), por
distancia ascendente y con el empate roto por orden alfabético de la forma normalizada. La
sección ["El algoritmo de sugerencias más parecidas"](../spec/vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas) tiene el algoritmo completo.

**Por qué Levenshtein y no prefijo común.** El prefijo común más largo es más simple de
especificar y de calcular, pero falla justo en el caso que una sugerencia de "lo más parecido"
tiene que cubrir: un error de tecleo que no está al final de la palabra. `sttaus`, con una
transposición al principio, tiene prefijo común de longitud cero con `status`, mientras que
Levenshtein lo pone a distancia 1, la máxima cercanía posible sin ser exacto. Encaja además con
el espíritu que `coincidir()` ya fija para el vocabulario cerrado: cubrir la forma de escribir
algo mal, no solo dónde empieza a escribirse.

**Por qué reutilizar `normalizar()` en vez de una métrica sobre el texto tal cual.** Calcular la
distancia sobre el valor tal cual escribiría dos veces la misma insensibilidad a mayúsculas,
guiones y guiones bajos que `coincidir()` ya resuelve, y podría desacompasarse de ella con el
tiempo. Con `normalizar()` compartido, un valor que ya coincide por el paso c de `coincidir()`
nunca necesita sugerencia, y las que sí la necesitan usan exactamente la misma noción de "casi
igual" en los dos sitios.

**Por qué un umbral y no siempre las `N` más cercanas que haya.** Sin umbral, un valor que no se
parece a nada del vocabulario recibiría igualmente `N` sugerencias, porque siempre hay *alguna*
candidata menos lejana que las demás. Eso es peor que no sugerir nada: quien recibe el error
puede tomar una de las `N` sin comprobarla, confiando en que "más parecida" significa "parecida
de verdad". El umbral hace que una entrada sin nada realmente cercano devuelva una lista vacía en
vez de rellenar el hueco con ruido, y ese es el caso de `xyz` contra un vocabulario de etiquetas
en el ejemplo de la sección enlazada arriba.

**Alternativas descartadas:**

- *Prefijo común más largo (longest common prefix), sin umbral.* Más simple y más rápida, pero no
  cubre el error de tecleo real (una transposición o una letra de más al principio colapsan el
  prefijo común a casi nada), que es justo el caso que una sugerencia tiene que resolver.
- *Levenshtein sin umbral.* Comparte la métrica y el desempate de la decisión vigente, pero
  siempre rellena las `N` sugerencias con lo que haya, aunque nada se parezca de verdad a la
  entrada; descartada por el mismo motivo que se explica arriba para el umbral.

---

## El presupuesto del mensaje de arranque

`biso prime` sustituye por completo a las guías de instrucciones y a cualquier inyección de texto en
los ficheros de convenciones del proyecto. Ese diseño se toma de una medición concreta.

En la herramienta estudiada, las cuatro guías de instrucciones se leyeron 69 veces en seis días y
suman **201.500 bytes, el 27,5% de toda la salida** que la herramienta devolvió a los agentes, más que
sus comandos más usados juntos.

| Lectura | Bytes | Cuándo la exige ese diseño |
|---|---:|---|
| Guía general | 2.365 | al empezar |
| Guía de creación | 4.014 | antes de crear |
| Guía de ejecución | 3.807 | antes de planificar o actualizar |
| Guía de finalización | 2.719 | antes de terminar |
| **Ciclo completo** | **12.905** | una sesión que crea, trabaja y cierra |

A eso hay que sumar la inyección de instrucciones en el fichero de convenciones del repositorio, que
se paga en todas las sesiones aunque no se toque el tablero.

El mensaje de `biso prime` mide **4.818 bytes**, 3.327 de parte fija y 1.491 de resumen del tablero
(["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)), contra un tope duro de 5.120 repartido en dos partes de 3.456 y
1.664.

| Magnitud | Herramienta estudiada | `biso` |
|---|---:|---|
| Peor caso por sesión, con ciclo completo | 12.905 bytes | 4.818 bytes |
| Media medida por sesión | 3.358 bytes | 4.818 bytes |
| Lecturas obligatorias por sesión | entre 1 y 4 | 1 |
| Contexto gastado en sesiones que no tocan tareas | la inyección en el fichero de convenciones | 0 |

**La media sube ligeramente, y conviene decirlo en vez de esconderlo.** Lo que cambia es otra cosa: el
coste pasa a ser fijo, conocido y acotado por una prueba, en vez de depender de cuántas guías decida
leer el agente, y el peor caso cae a menos de un tercio. El ahorro grande no está aquí, está en que la
salida de las escrituras deje de ser un eco y en las llamadas que desaparecen al fusionar el ciclo.

**El tope es una prueba de la suite, no un objetivo.** Y el reparto en dos partes existe para que el
resumen del tablero, que crece con el tablero, no pueda comerse el sitio de las reglas.

**El reparto entre las dos partes cambió con el modelo de estados, y el total no.** La parte fija
sube de 3.072 a 3.456 bytes porque el bloque `COMMANDS` gana las dos órdenes nuevas del ciclo, `ask` y
`answer`, y el bloque `RULES` gana una regla más, la undécima, sobre esos verbos y sobre qué
significa una tarea asignada; el resumen del tablero baja de 2.048 a 1.664.
Eso solo es seguro de hacer porque, a la vez, el orden de recorte del resumen deja de estar incompleto:
antes nombraba un solo bloque y decía "antes que cualquier otra cosa" sin nombrar ninguna otra, así que
un tablero con muchas tareas en curso podía rebasar el tope sin que hubiera una conducta definida para
ese caso. Con los cinco pasos completos que trae ahora ["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño), el resumen ya no
crece sin límite, y darle menos sitio cuesta filas mostradas, no correcciones. La parte fija, en
cambio, no se puede recortar sola: o cabe entera o hay que quitar contenido a mano, así que es la parte
que necesita más margen.

**El tope total, 5.120 bytes, no se mueve, y el motivo no es de contrato.** El contrato de estabilidad
solo obliga desde la versión 1.0, que todavía no está publicada, así que subir el tope no rompería
ninguna promesa hecha a nadie. La razón es de fondo: un tope que se sube cada vez que aprieta deja de
ser un tope, y su valor entero está en que obligue a elegir qué entra en el mensaje y qué se relega a
`--help`. Por eso, si al escribir el texto real de ["La salida literal"](../spec/cmd/prime.md#la-salida-literal) los números no cupieran, lo que se
recorta es contenido, no el tope.

**Esto no es "el tope nunca sube", es "el tope sube solo cuando reducir ya no es posible sin perder
algo".** La medida de aquel momento, 4.818 de 5.120 bytes, tenía 302 de margen: nunca había hecho falta
apretar para caber, así que esta regla no se había puesto a prueba todavía. El margen que de verdad
mandaba no era ese, sino el de la parte fija, que con los flags nuevos `--check-dod` y
`--uncheck-dod` en la rejilla de `FIELD FLAGS` había bajado a 129 bytes de los 3.456: cualquier texto
nuevo en el bloque fijo tenía que caber ahí, no en los 302 del total. Esos flags, y los demás
de la definición de hecho, se retiraron después (["Se retira la definición de hecho"](detalles.md#se-retira-la-definición-de-hecho)),
lo que devolvió a la parte fija unos 125 bytes; la cifra de 129 es la de aquel momento, no la de hoy.

**Ese punto llegó, y las cifras de este apartado son historia.** El rediseño de los flags de campo
de TASK-24 dejó pendiente cómo enseñarlos dentro de este presupuesto, y el experimento con un agente
real de TASK-25 (["El grid completo de flags de campo en el mensaje de arranque, medido con un
agente real"](#el-grid-completo-de-flags-de-campo-en-el-mensaje-de-arranque-medido-con-un-agente-real),
más abajo en este documento) subió el tope de 5.120 a **5.504 bytes**, y la parte fija de 3.456 a
**3.840**, con datos y no por conveniencia: es justo el caso que este apartado predijo, uno donde
reducir el bloque fijo ya solo se podía a costa de perder información que un agente necesita para no
adivinar mal. Las cifras de las tablas de arriba, y el texto literal de la sección
["La salida literal"](../spec/cmd/prime.md#la-salida-literal) que citan, quedan como estaban en el
momento del modelo de estados; la medida vigente hoy está en la entrada enlazada.

---

## El grid completo de flags de campo en el mensaje de arranque, medido con un agente real

**Decisión vigente.** El bloque `FIELD FLAGS` de `biso prime` enseña las cuatro formas explícitas de
cada campo de lista (añadir, quitar, vaciar, sustituir), sin ninguna regla que las derive, y el tope
de bytes del mensaje sube de 5.120 a **5.504** (parte fija de 3.456 a **3.840**) para que quepan. Es
el diseño que TASK-24 había dejado descartado por grande sin medirlo; el experimento de TASK-25 lo
midió contra las otras dos alternativas y ganó.

**El experimento.** Sobre tres redacciones candidatas del bloque `FIELD FLAGS`/`RULES` (mostrar solo
la forma de añadir con una regla de sufijos que deriva las otras tres; no mostrar ningún nombre de
flag y remitir a `--help`; mostrar las cuatro formas explícitas de cada campo) se dieron veinte
encargos en lenguaje llano a Claude Sonnet por variante y tipo de encargo, con `claude-ibm -p --bare
--system-prompt`, sin más contexto que el mensaje de `prime` simulado de cada variante y sin acceso a
ninguna herramienta. Los encargos: crear una tarea asignándola a alguien (`--add-assignees`,
un campo que solo aparece en `FIELD FLAGS`), añadir una etiqueta a una existente (`--add-labels`) y
quitarle una etiqueta (`--rm-labels`). Se contó en cuántas de las veinte el comando propuesto usaba el
nombre de flag correcto a la primera.

| Variante | Asignar | Añadir etiqueta | Quitar etiqueta |
|---|---:|---:|---:|
| Remitir a `--help`, sin ningún nombre en el mensaje | 0/20 | 0/20 | 0/20 |
| Regla de sufijos corta, con una plantilla `--rm-X` abstracta | 20/20 | 20/20 | 9/20 |
| Regla de sufijos, con los nombres ya compuestos en vez de una plantilla | (heredado, no repetido) | 20/20 | 20/20 |
| Grid completo, las cuatro formas explícitas de cada campo | 20/20 | 20/20 | 20/20 |

**Remitir a `--help` falla siempre, y de una forma que explica por qué.** Sin ningún nombre de
flag en el mensaje, el modelo no se queda callado ni consulta nada, porque el encargo es de una
sola vuelta y no tiene ninguna herramienta que llamar: inventa un nombre plausible y equivocado,
siempre del mismo tipo (`--label`, `--add-label`, `--rm-label` en singular, cuando el real es
plural). Esto confirma con datos lo que ya decía la sección
["Qué entra en el mensaje y qué se relega a `--help`"](../spec/cmd/prime.md#qué-entra-en-el-mensaje-y-qué-se-relega-a---help) de `prime.md`: un nombre que no está en el
mensaje no se consulta, se adivina.

**La primera regla de sufijos falla solo en un caso, y el fallo es de redacción, no del enfoque.**
Las once respuestas equivocadas de veinte escriben `biso set MYP-33 --rm-urgent`, tratando el valor
("urgent") como si fuera el nombre del campo, porque la regla explicaba el patrón `--rm-X` de forma
abstracta en vez de dar ya compuesto el nombre real (`--rm-labels`). Reescrita para dar los nombres ya
compuestos en cada caso citado, en vez de una plantilla con una X que sustituir, la regla de sufijos
sube a 20/20 en los dos encargos que se repitieron, empatando con el grid completo.

**Con las dos empatadas en acierto, decide el tamaño, y ahí gana el grid completo.** Medida la regla
de sufijos corregida dentro del texto real de `prime.md`, su rejilla `FIELD FLAGS` más corta no
compensa lo que cuesta la propia regla una vez que tiene que ser precisa para no confundir: **4.250**
bytes de parte fija, 794 por encima del tope de entonces. El grid completo, sin ninguna regla que
derivar, mide **3.695**: 239 bytes por encima, menos de un tercio de la diferencia. La razón es que
una regla que tiene que ser precisa para no confundir acaba repitiendo casi toda la información que
ya está en el grid, así que ahorra la rejilla pero no gana lo suficiente en la regla para compensarlo.

**Por qué subir el tope y no seguir recortando.** El grid completo, en su primera redacción, cabía
con solo 64 bytes de exceso sobre el tope total de entonces y 239 sobre la parte fija: mucho más
cerca de caber que la regla de sufijos. Recortar más sin volver a medir arriesgaba reintroducir el
mismo tipo de confusión que este mismo experimento acaba de encontrar y corregir una vez, así que se
prefirió subir el tope con la evidencia en la mano, siguiendo la propia regla de la sección
["El presupuesto del mensaje de arranque"](#el-presupuesto-del-mensaje-de-arranque): "sube solo cuando reducir ya no es posible sin perder
algo". Los números nuevos, 5.504 bytes de tope total y 3.840 de parte fija, dejan 145 bytes de margen
en la parte fija y 320 en el total, proporciones parecidas a las que dejaba el reparto anterior.

**Alternativas descartadas:**

- *Remitir a `--help`.* Descartada por el 0/60 medido arriba, y porque contradice el criterio ya
  escrito de que los nombres de flag, al ser lo que no se puede adivinar, tienen que entrar en el
  mensaje.
- *Regla de sufijos, incluso corregida.* Descartada pese a igualar en acierto al grid completo,
  porque mide más bytes en total una vez escrita con la precisión que exige no confundir.

**Lo que este experimento no mide.** `--help` sigue haciendo falta para todo lo que
["Qué entra en el mensaje y qué se relega a `--help`"](../spec/cmd/prime.md#qué-entra-en-el-mensaje-y-qué-se-relega-a---help) ya declaraba fuera del mensaje de arranque:
valores, tipos e incompatibilidades de cada flag, el formato de lote de `new --from`, los comandos
de administración y los casos límite. El grid completo elimina la adivinanza del *nombre* de un
flag de escritura, no la necesidad de `--help` para todo lo demás.

---

## Protocolo y resultado: comprobar con un agente fresco la dirección de una dependencia

**El protocolo ya se ejecutó, una revisión adversarial le encontró dos sesgos posibles, los dos se
comprobaron empíricamente, y el resultado sigue siendo que la regla 11 no se toca.** Lo que sigue
describe el protocolo de `TASK-89` tal como se diseñó, escrito para que otro agente lo ejecutara tal
cual sin decidir nada por su cuenta, con la tabla medida y el análisis que exige su criterio de
aceptación (["Los resultados medidos"](#cómo-se-tabula-y-qué-se-escribe-como-análisis)), y termina con
las dos contrapruebas que pidió la revisión
(["La revisión adversarial y sus dos hallazgos"](#la-revisión-adversarial-y-sus-dos-hallazgos)). Entre
las doce ejecuciones originales y las ocho de las contrapruebas, veinte ejecuciones acertaron la
dirección de la arista sin ninguna excepción, así que el texto de la regla 11 sigue siendo el de arriba
y no se tocó ni la especificación, ni el código, ni los ficheros de referencia.

**El problema que comprueba.** La regla 11 de `RULES`
(["La salida literal"](../spec/cmd/prime.md#la-salida-literal)) dice hacia dónde apunta una
dependencia:

```
 11. A dependency is written on the task that waits:
     `biso set MYP-10 --add-deps MYP-4` means MYP-4 blocks MYP-10.
```

Esa frase se escribió razonando sobre el texto y comprobando el comportamiento del programa
(["La ayuda enseña la dirección de una dependencia"](detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia)),
pero nunca se puso delante de un agente que no supiera nada más de `biso` que lo que ese mensaje le
cuenta. Este protocolo hace exactamente eso.

### El tablero de ejemplo

El tablero vive fuera del repositorio, en un directorio temporal fijo para que las órdenes de abajo se
puedan copiar y pegar sin sustituir nada. Quien ejecute el protocolo compila el binario y crea el
tablero con estas órdenes, en este orden, desde la raíz del repositorio:

```
mkdir -p /tmp/biso-task89/bin
go build -o /tmp/biso-task89/bin/biso ./cmd/biso

mkdir -p /tmp/biso-task89/template
/tmp/biso-task89/bin/biso -C /tmp/biso-task89/template init "Example Project" --prefix EXP --at ./board-data
/tmp/biso-task89/bin/biso -C /tmp/biso-task89/template new "Alpha"
/tmp/biso-task89/bin/biso -C /tmp/biso-task89/template new "Beta"
```

`new "Alpha"` es la primera tarea del tablero y recibe el identificador `EXP-1`; `new "Beta"` recibe
`EXP-2`. Ningún título describe una acción real ni sugiere un orden entre las dos, a propósito: si
"Alpha" y "Beta" fueran, por ejemplo, "Preparar el entorno" y "Ejecutar las pruebas", un agente podría
acertar la dirección por el sentido común del propio trabajo y no por haber leído la regla 11, y eso
falsearía la medida. No se fija ninguna identidad (`BISO_ME` no se declara), no se añade ningún tipo,
prioridad, etiqueta ni descripción: cualquier campo de más es una variable que el protocolo no necesita
controlar.

`/tmp/biso-task89/template` queda como la copia maestra, sin ninguna dependencia entre sus dos tareas.
No se toca nunca directamente: cada una de las doce ejecuciones de más abajo trabaja sobre su propia
copia.

**El texto exacto que imprime `biso prime` sobre este tablero**, capturado con
`biso -C /tmp/biso-task89/template prime`, y que hay que pegar sin cambiar ni un carácter donde este
protocolo lo pide:

```
biso 1.0.0 - the task board of this project. This message is all you need to start.

BOARD  Example Project
  To Do 2 | In Progress 0 | Done 0
  new tasks start in To Do; `biso start` moves to In Progress; `biso finish` to Done
  types       task, bug, docs
  priorities  high, medium, low
  you are     (not set: run biso as BISO_ME=@you biso ...)

COMMANDS  (`biso help <cmd>...` for the detail of any, several at once)
  biso ls [--status STATUS] [--type T] [--label LABEL] [--mine] [--search TEXT]
  biso get <ref> [--section ac]
  biso new "TITLE" [--append-desc TEXT] [--add-ac TEXT]... [--type T] [--priority P]
  biso start <ref>... [--append-plan TEXT]
  biso note <ref> "TEXT"
  biso ask <ref> "QUESTION"
  biso answer <ref> "TEXT"
  biso finish <ref>... [--append-summary "TEXT"] [--check-ac all]
  biso set <ref>... [any field flag]
  biso comment <ref> "TEXT" [--comment-author @who]

FIELD FLAGS  (same names, same meaning, in every command above that writes)
  --title  --status  --type --clear-type  --priority --clear-priority
  --parent --clear-parent  --due --clear-due  --author --clear-author
  --ordinal first|last  --above <ref> --below <ref>  --clear-ordinal
  --add-labels --rm-labels --clear-labels --replace-labels
  --add-assignees --rm-assignees --clear-assignees --replace-assignees
  --add-refs --rm-refs --clear-refs --replace-refs
  --add-deps --rm-deps --clear-deps --replace-deps
  --add-ac --rm-ac --clear-acs   --check-ac --uncheck-ac
  --append-desc --clear-desc  --append-plan --clear-plan
  --append-note --clear-notes  --append-summary --clear-summary
  --comment --rm-comment --set-comment-date

RULES  (none of these are guessable; they are the whole learning curve)
  1. Every write goes through biso. Nothing else touches the board.
  2. <ref> is an id (MYP-12), a bare number (12) or free text ("CRLF"). Text
     matching several tasks is an error that lists them, never a guess. `note`,
     `comment`, `ask` and `answer` take one <ref>; `set`, `start` and `finish`
     take several.
  3. Filters reject values this board does not have: `--status Pending` is an error,
     not an empty list. Case, spaces, hyphens and underscores are ignored, so
     `--status todo`, `--status "To Do"` and `--status TO_DO` are one and the same filter. An
     empty list is therefore a fact about the board that you can act on.
  4. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says
     on stderr what it left out. --all lifts the limit, --any-status includes
     Done, --archived reaches the archive.
  5. --check-ac and --uncheck-ac take all, 3, 1-4, 1,3,7 or the criterion text. The
     numbers are the stable #N keys that `biso get` shows, and they never shift
     when one criterion is removed.
  6. `biso new` prints the new id and nothing else. Every other write prints one
     line per task: id, status, criteria, urgency. Add --print for the whole
     record, or --json for a versioned envelope.
  7. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.
  8. Long text: a real newline works, and so do --append-desc @file.md and --append-desc - for stdin.
  9. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,
     6 precondition not met, 7 nothing written, 8 environment, 20 no board here.
 10. `biso ask <ref> "..."` parks a task on a question and `biso answer` unparks
     it, writing both into the comments. Ask instead of guessing. A task
     assigned to you is one a person decided you should do.
 11. A dependency is written on the task that waits:
     `biso set MYP-10 --add-deps MYP-4` means MYP-4 blocks MYP-10.

NEXT UP  (not assigned to you, by urgency)
  EXP-1  To Do  -  -  Alpha  -  -  -
  EXP-2  To Do  -  -  Beta   -  -  -

Pick one, `biso start <ref> --append-plan "..."`, work, `biso note <ref> "..."` as you go,
and close with `biso finish <ref> --check-ac all --append-summary "..."`.
That is the loop. Create a task when the work needs planning or review; do small
edits directly.
```

Si al ejecutar el protocolo esta captura sale distinta (otra versión en la primera línea, otra fecha,
otro texto de `RULES`), es la nueva captura la que se usa en todo lo que sigue, no la de aquí arriba:
esta es la fotografía del momento en que se escribió el protocolo, y la única que cuenta de verdad es
la que produce el binario que se está probando. La cadena de versión de la primera línea puede cambiar
sin que eso invalide nada, igual que dice
["La salida literal"](../spec/cmd/prime.md#la-salida-literal) de `prime.md`.

### Las cuatro formulaciones del encargo

Cuatro maneras de pedir la misma cosa en el mundo real: que `Alpha` vaya primero y que `Beta` espere
por ella. Ninguna nombra un comando de `biso` ni dice la palabra "dependency" o "dependencia" salvo la
que la necesita por construcción (la segunda, que es justamente la forma "B depende de A"). Van en
inglés, como el resto del texto que en un tablero real escribiría quien usa el programa (la interfaz
de `biso`, sus mensajes y sus ejemplos son ingleses de punta a punta), y para no mezclar dos idiomas
dentro de la misma conversación del agente fresco, que lee la regla 11 también en inglés.

| Formulación | Texto exacto del encargo |
|---|---|
| F1, "A bloquea a B" | `Alpha blocks Beta.` |
| F2, "B depende de A" | `Beta depends on Alpha.` |
| F3, "A antes que B" | `Alpha has to happen before Beta.` |
| F4, "B no puede empezar hasta que A termine" | `Beta cannot start until Alpha is finished.` |

Cada celda de la columna derecha es el mensaje completo que recibe el agente fresco: ni una palabra
más, ni una coma menos.

### Qué recibe cada agente fresco, y qué no

**El sitio de trabajo.** Bajo `/tmp/biso-task89/runs/` hay doce carpetas, una por ejecución, con esta
forma fija (`<slug>` es uno de `f1-blocks`, `f2-depends`, `f3-before`, `f4-cannot-start`, y `<n>` es
`1`, `2` o `3`):

```
/tmp/biso-task89/runs/<slug>/<n>/board/
```

Cada una de esas doce carpetas `board/` es una copia completa e independiente de
`/tmp/biso-task89/template/`, copiada así:

```
mkdir -p /tmp/biso-task89/runs/<slug>/<n>
cp -R /tmp/biso-task89/template/. /tmp/biso-task89/runs/<slug>/<n>/board/
```

Toda copia cuelga de la misma ruta relativa (`board/` dentro de su propia carpeta de ejecución), así
que la orden que copia, la que lanza el agente y la que verifica el resultado son idénticas en las doce
salvo por `<slug>` y `<n>`. `--at` se guardó como ruta relativa (`./board-data`) al crear la maestra, así
que cada copia resuelve su propio `board-data/` sin arrastrar nada de las demás
(["`biso init`"](../spec/cmd/init.md)).

**El agente fresco es una invocación de `claude` en modo no interactivo**, no un subagente de esta
misma sesión de Claude Code: un subagente lanzado con la herramienta `Agent` heredaría, por estar
dentro de este repositorio, los ficheros `CLAUDE.md` del proyecto (el de la raíz, el de `.claude/` y
este mismo documento en el índice de memoria), que hablan de la especificación, del tablero de
Backlog.md y de cómo se trabaja en `biso`; eso es exactamente el contexto que el protocolo tiene que
excluir. `/tmp/biso-task89/` está fuera de cualquier repositorio con `CLAUDE.md`, así que lanzar `claude`
desde ahí, con `--bare`, no descubre ninguno.

Antes de lanzar las doce ejecuciones se escriben dos ficheros de texto, palabra por palabra:

`/tmp/biso-task89/prompts/system.txt`, igual para las doce ejecuciones (sustituir
`<PRIME_TEXT>` por el bloque completo capturado en la sección anterior, sin tocarlo):

```
You are a software engineer with a shell. Bash is your only tool. The project you are working in
uses a command-line program called biso to track its tasks, and biso is already on your PATH.
Before you read anything else, running `biso prime` printed this automatically:

<PRIME_TEXT>

Nothing else about biso or about this project has been explained to you. A colleague is about to
tell you one fact about the board. Decide which single biso command makes the board match what they
say, run it, and then reply with exactly one line: the command you ran and nothing else.
```

`/tmp/biso-task89/prompts/<slug>.txt`, uno por formulación, con el texto exacto de la tabla de arriba
y nada más (por ejemplo `f1-blocks.txt` contiene la única línea `Alpha blocks Beta.`).

**La orden que lanza una ejecución**, igual para las doce salvo `<slug>` y `<n>`:

```
cd /tmp/biso-task89/runs/<slug>/<n>/board
PATH="/tmp/biso-task89/bin:$PATH" \
claude -p \
  --bare \
  --model sonnet \
  --tools "Bash" \
  --strict-mcp-config \
  --disable-slash-commands \
  --permission-mode bypassPermissions \
  --max-budget-usd 1 \
  --output-format json \
  --system-prompt "$(cat /tmp/biso-task89/prompts/system.txt)" \
  "$(cat /tmp/biso-task89/prompts/<slug>.txt)" \
  > /tmp/biso-task89/runs/<slug>/<n>/result.json \
  2> /tmp/biso-task89/runs/<slug>/<n>/stderr.log
```

`--bare` apaga el descubrimiento de `CLAUDE.md`, los hooks, los plugins y la memoria automática; junto
con `--strict-mcp-config` y `--disable-slash-commands` deja al agente sin más herramienta que `Bash` y
sin ningún comando propio que invocar. `--permission-mode bypassPermissions` es necesario porque
`--print` no tiene a nadie que conteste un permiso; es aceptable porque la carpeta sobre la que actúa
es una copia desechable fuera del repositorio. `--max-budget-usd 1` es un límite de seguridad, no una
medida del experimento: ninguna de las cuatro formulaciones debería necesitar gastar cerca de eso para
una sola escritura.

**Qué recibe, en resumen: el texto de `biso prime` de la sección anterior (dentro de `--system-prompt`,
sin nada más alrededor salvo el envoltorio fijo de arriba), la formulación exacta que le toca (como
único mensaje de usuario), una copia propia del tablero en el directorio desde el que se lanza
`claude`, y la instrucción de ejecutar el comando de `biso` que corresponda y decir cuál fue.**

**Qué no recibe, explícitamente: ningún fichero de `docs/spec/`, ningún fichero `.go` del repositorio,
este documento ni ninguna otra página de `docs/decisiones/`, ninguna mención de que se está midiendo
la dirección de una dependencia, ni acceso a ninguna herramienta que no sea `Bash`.** No hace falta
ninguna orden que se lo impida aparte de las banderas de arriba: `--tools "Bash"` ya deja fuera
`Read`, `Grep` y cualquier otra que pudiera usar para explorar el repositorio, y de todas formas el
directorio de trabajo (`board/`, una copia de `template/`) no contiene ni la especificación ni el
código fuente, solo el puntero y la base de datos del tablero.

**Cómo se ejecutó en la práctica.** El entorno donde se corrió `TASK-89` no podía invocar el binario
`claude` como proceso externo, así que las doce ejecuciones se lanzaron con la herramienta `Agent` de
esa misma sesión de Claude Code (`subagent_type: "general-purpose"`, sin `isolation` y sin pasarle
ningún contexto de la conversación que coordinaba el protocolo), una llamada independiente por
ejecución. Es la misma salvedad que el párrafo de arriba advertía que había que evitar, aceptada aquí
como la única vía disponible: cada agente fresco recibió, como todo su encargo, el texto exacto que
este apartado describe (el bloque de `biso prime` capturado más arriba, la formulación que le tocaba y
una frase neutra con la ruta de su copia del tablero y la del binario), sin ninguna mención de "TASK-89",
de "protocolo", de "regla 11" ni de que se estuviera midiendo nada. La diferencia frente al diseño
original es que el agente no llevaba `--tools "Bash"` forzado por la línea de comandos, así que en
teoría podía usar `Read` o `Grep`, y su directorio de trabajo por defecto no era la copia del tablero
sino el que hereda de la sesión que lo lanza. **Esa teoría resultó ser el problema real**, medido y
corregido en ["La revisión adversarial y sus dos hallazgos"](#la-revisión-adversarial-y-sus-dos-hallazgos),
más abajo: la carpeta desde la que trabaja un agente lanzado así no es la copia de `template/` que se
le nombra en el texto, y sí lleva de serie los `CLAUDE.md` del proyecto.

### Cómo se verifica el resultado

Tras cada ejecución, con `<slug>` y `<n>` de esa ejecución:

```
BOARD=/tmp/biso-task89/runs/<slug>/<n>/board
D1=$(/tmp/biso-task89/bin/biso -C "$BOARD" get EXP-1 --json | jq -r '.data.task.dependencies | join(",")')
D2=$(/tmp/biso-task89/bin/biso -C "$BOARD" get EXP-2 --json | jq -r '.data.task.dependencies | join(",")')
```

`dependencies` es el campo que se escribe (["Las relaciones entre tareas"](../spec/modelo-de-datos/relaciones.md#dependencies-la-precedencia));
`blocks` es su inverso derivado y no hace falta leerlo aparte para clasificar, aunque coincide siempre
con `dependencies` por construcción. La clasificación de esa ejecución, por este orden:

1. **Acierto.** `$D2` es `EXP-1` (y, por construcción, `$D1` es vacío). `Beta` quedó esperando a
   `Alpha`, que es lo que piden las cuatro formulaciones dicho de cuatro maneras.
2. **Inversión.** `$D1` es `EXP-2` (y, por construcción, `$D2` es vacío). La arista se escribió al
   revés: `Alpha` quedó esperando a `Beta`.
3. **Sin resultado.** Cualquier otro caso: `$D1` y `$D2` vacíos los dos (no se escribió ninguna
   dependencia), o cualquier valor que no sea exactamente uno de los dos patrones de arriba (por
   ejemplo, una referencia añadida con `--add-refs` en vez de `--add-deps`, o un identificador que no
   es ni `EXP-1` ni `EXP-2`). También cae aquí una ejecución que no terminó: el proceso `claude` salió
   con código distinto de cero, agotó `--max-budget-usd`, o `result.json` no contiene una respuesta
   final. `stderr.log` y `result.json` de esa ejecución se guardan igual y se leen a mano para escribir
   la nota del caso en el análisis.

**"Sin resultado" no es ni acierto ni inversión, y se cuenta en su propia columna.** Descartarlo del
recuento, o repartirlo entre las otras dos categorías porque "en el fondo no invirtió nada", escondería
un fallo distinto: el agente no llegó a actuar sobre la regla 11 en absoluto, por el motivo que sea, y
eso es un dato sobre el protocolo o sobre el propio agente, no sobre si la regla enseña bien la
dirección.

### Cómo se tabula y qué se escribe como análisis

La tabla, sobre tres ejecuciones por formulación, se rellena con las cifras medidas y no antes:

| Formulación | Aciertos | Inversiones | Sin resultado |
|---|---|---|---|
| F1, `Alpha blocks Beta.` | 3 | 0 | 0 |
| F2, `Beta depends on Alpha.` | 3 | 0 | 0 |
| F3, `Alpha has to happen before Beta.` | 3 | 0 | 0 |
| F4, `Beta cannot start until Alpha is finished.` | 3 | 0 | 0 |
| **Total** | 12 | 0 | 0 |

Cada fila suma la cantidad fija de ejecuciones de esa formulación; la fila de total suma las doce.

**El análisis.** Las doce ejecuciones acertaron, sin una sola excepción: en las cuatro formulaciones,
los tres agentes frescos de cada una escribieron la dependencia con `biso set EXP-2 --add-deps EXP-1`
(verificado con `biso get EXP-1 --json` y `biso get EXP-2 --json` sobre cada una de las doce copias del
tablero, que dan `EXP-1` con `dependencies` vacío y `EXP-2` con `dependencies: ["EXP-1"]` en las doce).
No hay ninguna formulación que falle una sola vez, ni un patrón que aparezca en una ejecución y no en
las otras dos de la misma fila: las tres repeticiones de cada formulación coinciden entre sí.

El dato más informativo no es que acertaran, sino que las cuatro lo hicieron con el mismo comando pese
a estar redactadas de maneras muy distintas. F1 ("Alpha blocks Beta") nombra el verbo activo
(`blocks`) que también usa la propia regla 11 al describir el ejemplo, así que un acierto ahí es
consistente con leer la regla y aplicarla, pero también con adivinar por la forma de la frase sin haber
entendido la regla. F2 ("Beta depends on Alpha") es la única que usa la palabra "depends" y la nombra
en la dirección contraria a como la nombra F1: si el acierto de F1 fuera un efecto de superficie del
verbo "blocks" y no de haber entendido la regla, F2 tendría que fallar o al menos vacilar, porque pide
traducir "depende de" a "escribe la dependencia en la tarea que depende", que es exactamente el paso
que la regla 11 explica y que ninguna otra parte del mensaje de arranque dice. F2 acertó las tres veces
igual que F1. F3 y F4 no usan ni "blocks" ni "depends": expresan la misma relación con un orden temporal
("has to happen before") y con una condición de bloqueo ("cannot start until... is finished"), y
tampoco fallaron. Que las cuatro formulaciones, con vocabulario disjunto entre sí, conviertan al mismo
comando sin ninguna inversión es la evidencia de que la regla 11 enseña la dirección con claridad
suficiente para un agente que parte de cero, y no solo para quien ya sabe interpretar el ejemplo. Esta
lectura da por hecho que el agente partía de cero y que el orden de creación no jugaba ningún papel;
las dos cosas se pusieron a prueba después, en
["La revisión adversarial y sus dos hallazgos"](#la-revisión-adversarial-y-sus-dos-hallazgos).

Con el criterio fijado antes de medir (una sola inversión atribuible a la regla 11, en cualquiera de las
doce ejecuciones, bastaría para reescribir su texto), el resultado no lo activa: cero de doce
ejecuciones invirtieron la arista y cero se quedaron sin resultado. El texto de la regla 11 se deja tal
como está, y no se corrigen ni la especificación, ni el código, ni los ficheros de referencia de
`cmd/biso/testdata/`.

### Qué pasa si algún agente invierte la arista

**El criterio: una sola inversión atribuible a la regla 11, en cualquiera de las doce ejecuciones,
basta para reescribir su texto.** No hace falta que sea mayoría dentro de una formulación. La razón es
la misma que justificó escribir la regla la primera vez
(["La ayuda enseña la dirección de una dependencia"](detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia)):
una dependencia invertida es un dato válido que el programa nunca detecta, y deja el cálculo de
bloqueo y de urgencia corrompido en silencio para siempre, no solo en esa sesión. Un umbral de mayoría
toleraría a propósito una tasa de inversión medida y real (por ejemplo una de tres, un tercio de las
ejecuciones de esa formulación) para el único tipo de error que el programa no puede avisar, lo que
contradice el motivo por el que la regla existe. Se descarta por eso: el coste de corregir el texto es
bajo (cabe de sobra en el margen medido más abajo) frente al coste de dejar una frase que se sabe que
confunde a un agente de cada tantos.

**Antes de contar una inversión hacia este criterio, se lee su transcripción.** `result.json` de esa
ejecución trae la respuesta completa del agente fresco. Si el motivo fue no entender hacia dónde
apunta la regla 11 (leyó "Alpha blocks Beta" y escribió la dependencia en `EXP-1` en vez de en `EXP-2`,
por ejemplo), cuenta hacia el criterio de arriba. Si el motivo es otro y no tiene que ver con la
dirección (confundió qué identificador era `Alpha` y cuál `Beta`, por ejemplo, algo que ninguna
redacción de la regla 11 podría arreglar), se anota igual en el análisis, pero no obliga a tocar el
texto de la regla: ese fallo no lo causa la regla, lo causaría igual cualquier redacción.

**Si ninguna de las doce invierte la arista, la entrada lo dice así, con la tabla de arriba ya rellena
de ceros en esa columna, y el texto de la regla 11 no se toca.** Es lo que pasó: las doce ejecuciones
midieron `EXP-2` con `EXP-1` como dependencia y `EXP-1` sin ninguna, la columna de inversiones de la
tabla de arriba queda en cero en cada formulación, y el párrafo siguiente sobre el margen de bytes
queda como referencia para la próxima vez que haga falta tocar `RULES`, no como algo que esta medición
haya usado.

**El margen de bytes disponible hoy, medido con el binario y no con la cifra ya escrita en
["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño):** la parte fija del
mensaje mide **3.623 bytes** sobre el propio tablero de este protocolo (`EXP-1`/`EXP-2`, sin ninguna
tarea más), exactamente la misma cifra que mide sobre el tablero de la especificación
(["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)), porque la parte fija
no depende del contenido del tablero. El tope de la parte fija es **3.840 bytes**
(["El presupuesto de tamaño"](../spec/presupuestos.md#el-presupuesto-de-tamaño)), así que quedan
**217 bytes** libres hoy para ampliar la regla 11 (o cualquier otra parte del bloque fijo) sin tocar el
tope. Si la corrección no cupiera en esos 217 bytes, el orden a seguir es el de
["El presupuesto del mensaje de arranque"](#el-presupuesto-del-mensaje-de-arranque): apretar el resto
del texto fijo antes de subir el tope, y subirlo solo con la misma clase de evidencia medida que subió
el tope de 5.120 a 5.504 en
["El grid completo de flags de campo en el mensaje de arranque, medido con un agente real"](#el-grid-completo-de-flags-de-campo-en-el-mensaje-de-arranque-medido-con-un-agente-real).

**Si el texto se corrige, el protocolo se repite entero sobre el texto nuevo**, con un tablero de
ejemplo nuevo (mismas órdenes, misma ruta) y las mismas doce ejecuciones, porque un texto corregido sin
volver a medirlo es exactamente la situación que este protocolo existe para no dar por buena. Cuando la
corrección exista, la especificación
(["La salida literal"](../spec/cmd/prime.md#la-salida-literal)), el código y los ficheros de referencia
de `cmd/biso/testdata/` cambian a la vez, como pide la regla de este repositorio sobre ejemplos y
fixtures.

### La revisión adversarial y sus dos hallazgos

Una revisión posterior, hecha por un agente que no había escrito ni ejecutado nada de lo de arriba,
encontró dos maneras en que la medición de doce aciertos y cero inversiones podía estar sesgada. Las
dos se comprobaron empíricamente, no se descartaron ni se dieron por buenas de palabra.

**Primer hallazgo: el orden de creación coincidía con la respuesta correcta en las doce ejecuciones.**
`EXP-1` ("Alpha", creada primero) quedó siempre sin dependencias y `EXP-2` ("Beta", creada segunda)
quedó siempre con la dependencia. Un agente que ignorase la regla 11 por completo y aplicase la
heurística "lo creado antes no depende de nada, lo creado después es lo que espera" habría acertado
igual las doce veces, sin haber entendido la dirección que enseña la regla.

La contraprueba fue un segundo tablero, creado con el orden invertido: `biso new "Beta"` primero
(`EXP-1`) y `biso new "Alpha"` segundo (`EXP-2`), dejando intacto todo lo demás (mismos títulos, mismo
significado de las cuatro formulaciones). En este tablero la respuesta correcta según la regla 11 es
la contraria a la que predice la heurística del orden: `Beta` sigue siendo la que espera, así que la
dependencia va en `EXP-1` (aunque se creó primero) apuntando a `EXP-2` (aunque se creó segundo). La
heurística del orden, de estar operando, habría escrito `EXP-2 --add-deps EXP-1`, que en este tablero
es la dirección equivocada.

Se probó una ejecución por formulación sobre este tablero invertido (cuatro en total, no doce, porque
esto es una contraprueba de un sesgo concreto y no una repetición de la medida). Las cuatro acertaron
la dirección real y ninguna aplicó la heurística del orden:

| Formulación | Comando ejecutado | Correcto según la regla 11 |
|---|---|---|
| F1, `Alpha blocks Beta.` | `biso set Beta --add-deps Alpha` (`EXP-1 --add-deps EXP-2` por resolución de título) | sí |
| F2, `Beta depends on Alpha.` | `biso set EXP-1 --add-deps EXP-2` | sí |
| F3, `Alpha has to happen before Beta.` | `biso set EXP-1 --add-deps EXP-2` | sí |
| F4, `Beta cannot start until Alpha is finished.` | `biso set EXP-1 --add-deps EXP-2` | sí |

Verificado con `biso get EXP-1 --json` y `biso get EXP-2 --json` sobre cada copia: las cuatro dieron
`EXP-1` con `dependencies: ["EXP-2"]` y `EXP-2` sin ninguna, que es la dirección que pide la regla 11 y
la contraria a la que predice "lo creado antes nunca depende de nada". El sesgo del orden de creación
queda descartado con esta evidencia: los dieciséis aciertos conjuntos (los doce originales más estos
cuatro) no se explican por una heurística de identificador, porque cuando el identificador y la regla
piden cosas distintas, los agentes siguieron a la regla.

**Segundo hallazgo: los agentes frescos podían estar recibiendo el `CLAUDE.md` del proyecto.** El
párrafo de arriba, escrito antes de correr el protocolo, ya avisaba de que el diseño original quería
excluirlo y que lanzar los doce con la herramienta `Agent` en vez de con `claude -p --bare` external era
la salvedad que rompía esa exclusión. La comprobación pendiente era si de verdad pasaba.

Se comprobó con tres agentes de diagnóstico, cada uno con una única instrucción neutra (ejecutar `pwd`,
buscar `CLAUDE.md` en cada directorio ascendiente hasta `/`, y decir si ya traían contexto de proyecto
antes de leer el mensaje), lanzados con las tres formas disponibles de la herramienta `Agent`:

| Modo | `pwd` | `CLAUDE.md` encontrados subiendo hasta `/` | ¿Traía contexto de Biso ya puesto? |
|---|---|---|---|
| Sin `isolation` | `/Users/avilches/Hub/Projects/Biso` | `Biso/CLAUDE.md`, `Hub/CLAUDE.md` | sí, los dos completos más la memoria |
| `isolation: "worktree"` | `.../Biso/.claude/worktrees/agent-<id>` | el `CLAUDE.md` del propio worktree, `.claude/CLAUDE.md`, `Biso/CLAUDE.md`, `Hub/CLAUDE.md` | sí, igual que sin `isolation` |
| `isolation: "remote"` | el mismo patrón `.../worktrees/agent-<id>` que `worktree` | los mismos cuatro | sí, igual que los otros dos |

Ninguno de los tres modos evita la fuga: los tres heredan como directorio de trabajo por defecto el de
la sesión que los lanza, no la ruta que se les nombra dentro del mensaje, y esa ruta tiene `CLAUDE.md`
en varios niveles. `isolation: "worktree"` no ayuda porque el worktree que crea es del propio
repositorio de Biso, así que arrastra su `CLAUDE.md` igual; `isolation: "remote"` mostró el mismo
patrón de directorio que `worktree`, así que tampoco lo evitó en este entorno. La herramienta `Agent`
de esta sesión no tiene ningún parámetro para arrancar un agente en un directorio limpio sin
ascendientes con `CLAUDE.md`, así que no hay, dentro de lo que ofrece esta herramienta, una forma de
replicar la exclusión que lograba `claude -p --bare --tools "Bash"` del diseño original. Esto también
significa que un agente lanzado así no estaba limitado a `Bash`: en teoría podía usar `Read` o `Grep`
para abrir `docs/spec/cmd/set.md` o `docs/decisiones/detalles.md`, que explican la dirección de una
dependencia en prosa mucho más extensa que la regla 11, y acertar por haber leído la especificación en
vez de por haber entendido `biso prime`. Ninguna de las dieciséis ejecuciones que se pueden auditar
(ver el párrafo siguiente) lo hizo, pero la posibilidad estuvo abierta en las doce originales sin que
nada la cerrara.

La mitigación disponible, a falta de una `isolation` que dé un directorio limpio, fue instruir al
agente dentro del propio mensaje: una frase explícita pidiéndole que ignore cualquier contexto de
proyecto que ya tuviera puesto y que no lea, busque ni abra ningún fichero salvo para ejecutar el
comando de `biso`, más una segunda línea de respuesta obligatoria diciendo si leyó algo más. Con esa
instrucción se repitió una ejecución por formulación (cuatro en total) sobre una copia nueva del
tablero original (`EXP-1` "Alpha" sin dependencias, `EXP-2` "Beta" con la dependencia, el mismo reparto
que las doce primeras ejecuciones):

| Formulación | Comando ejecutado | ¿Leyó algún fichero además del binario? |
|---|---|---|
| F1, `Alpha blocks Beta.` | `biso set EXP-2 --add-deps EXP-1` | no |
| F2, `Beta depends on Alpha.` | `biso set EXP-2 --add-deps EXP-1` | no (consultó `biso --help` y `biso set --help`, ningún fichero) |
| F3, `Alpha has to happen before Beta.` | `biso set EXP-2 --add-deps EXP-1` | no |
| F4, `Beta cannot start until Alpha is finished.` | `biso set EXP-2 --add-deps EXP-1` | no |

Las cuatro acertaron, verificado igual que las demás con `biso get --json`, y las cuatro declararon no
haber leído nada fuera del binario de `biso` (dos de ellas sí consultaron su propia ayuda con
`--help`, que no es un fichero del repositorio). Esa declaración es autoinforme del propio agente, no
una garantía estructural como la que daba `--tools "Bash"` en el diseño original, y por eso es una
evidencia más débil que una exclusión real; combinada con que ninguna de las ocho ejecuciones nuevas
(las cuatro de este hallazgo más las cuatro del sesgo de orden) usó más de un puñado de llamadas a
herramientas, ninguna de ellas coherente con haber abierto y leído una página de la especificación,
apoya que la fuga de `CLAUDE.md`, aunque real y confirmada, no cambió el resultado medido: ningún
agente de los veinte (doce originales, cuatro de la contraprueba del orden, cuatro de esta
confirmación) escribió la dependencia al revés.

**Qué demuestra el experimento corregido y qué no.** Demuestra que, sobre las veinte ejecuciones que se
han corrido en total sobre este texto de la regla 11, ninguna la invirtió, ni cuando el identificador
más bajo era el que debía depender (contraprueba del orden) ni cuando se le pidió al agente que
ignorase cualquier fuga de contexto de proyecto (contraprueba del `CLAUDE.md`). No demuestra que la
fuga de `CLAUDE.md` sea inofensiva en general: solo que en las ocho ejecuciones donde se comprobó
expresamente, ninguna explotó el acceso a herramientas de más para ir a leer la especificación, a
juzgar por su propio informe y por su número de llamadas a herramientas. Una repetición futura de este
protocolo con una herramienta capaz de arrancar un agente en un directorio sin ningún `CLAUDE.md`
ascendiente (el equivalente real a `claude -p --bare --tools "Bash"`) cerraría esta pega por completo;
hasta entonces, la mitigación de instruir al agente a ignorar el contexto y a no leer ficheros de más
es la mejor disponible con las herramientas de esta sesión, y quedó aplicada y verificada en las ocho
ejecuciones de arriba.

Con el criterio de una sola inversión ya fijado, y sin que ninguna de las veinte ejecuciones lo
active, la conclusión no cambia: el texto de la regla 11 sigue siendo el de arriba, y no se toca ni la
especificación, ni el código, ni los ficheros de referencia.

**Tercer hallazgo (menor, no obliga a nada): la formulación en voz pasiva.** Se sugirió añadir una
quinta formulación del tipo "Beta is blocked by Alpha." como contraprueba adicional de vocabulario. Se
descarta por escrito: las cuatro formulaciones que ya tiene la tabla alternan cuál de las dos tareas se
nombra primero (F1 y F3 empiezan por `Alpha`, F2 y F4 empiezan por `Beta`) y usan vocabulario disjunto
entre sí (`blocks`, `depends on`, `has to happen before`, `cannot start until... is finished`), así
que ya cubren tanto el orden de mención como la ausencia de una palabra común entre todas. Una quinta
formulación en voz pasiva pondría a prueba una construcción gramatical distinta, no una dirección
semántica distinta, y el criterio de aceptación de `TASK-89` pide formulaciones que digan lo mismo de
maneras distintas, no que cubran todas las construcciones gramaticales posibles del inglés.
