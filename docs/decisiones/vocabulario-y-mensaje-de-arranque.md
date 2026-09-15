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
nuevo en el bloque fijo tenía que caber ahí, no en los 302 del total.

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
