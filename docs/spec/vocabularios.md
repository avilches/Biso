# Los vocabularios del tablero y la regla de validación

Hay campos con vocabulario cerrado, definido en la configuración: `status`, `type` y `priority`. Para
todos ellos rige una sola regla, **idéntica al escribir y al leer**.

**Idéntica al leer significa también lo ya guardado.** Un valor de `status`, `type` o `priority` que
una tarea guarda y que la configuración no declara es un error igual que uno tecleado en un filtro: la
tarea es ilegible y ningún comando de lectura la trata como si estuviera bien
(["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)).
La comparación con lo guardado es exacta, sin los pasos del algoritmo de abajo, porque lo que se guarda
es siempre la grafía configurada.

## El algoritmo de coincidencia

Dado un valor de entrada `v` y la lista de valores configurados, el programa calcula así:

```
normalizar(x):
  1. plegar mayusculas y minusculas segun Unicode (case folding), no pasar a minusculas
  2. descomponer y quitar los diacriticos (acentos, dieresis, cedillas)
  3. eliminar TODOS los caracteres que sean espacio, tabulador, guion (-) o guion bajo (_)
  4. devolver lo que queda

coincidir(v, configurados):
  a. si existe un configurado c con c == v exactamente, devolver c
  b. si no, calcular normalizar(v) y compararlo con normalizar(c) de cada
     configurado, contando una sola vez los configurados repetidos letra por letra
  c. si exactamente un configurado coincide, devolverlo
  d. si ninguno coincide, error 3
  e. si coinciden dos o mas configurados distintos, error 3 con los dos listados,
     porque el tablero tiene dos valores que se normalizan igual y hay que
     desambiguarlos
```

**Qué es el paso 1, y por qué no es pasar a minúsculas.** El paso 1 es el **plegado de mayúsculas
y minúsculas de Unicode** (el *case folding*), la misma regla que ["Selectores de
criterios"](familias-de-flags.md#selectores-de-criterios) fija para el fragmento de un selector de
texto. Pasar a minúsculas no vale, y la diferencia no es teórica: el griego escribe la sigma final
de una palabra con un carácter distinto (`ς`) del de la sigma de en medio (`σ`), y pasar a
minúsculas la forma en mayúsculas de un estado que acabe en sigma da una cadena que no es igual a
la de ese mismo estado tecleado en minúsculas. Un tablero con el estado `Δοκιμές` rechazaría
`ΔΟΚΙΜΕΣ`, que es su propio estado escrito en mayúsculas. El plegado lleva las dos escrituras al
mismo representante y las dos coinciden. Lo mismo hace con el signo de micro y la `μ`, con el
signo de kelvin y la `k`, y con la ese larga `ſ` y la `s`.

**Qué cuenta como separador en el paso 3.** Exactamente cuatro caracteres, y ningún otro: el
espacio (U+0020), el tabulador (U+0009), el guion (U+002D) y el guion bajo (U+005F). Ningún otro
espacio en blanco es un separador: un espacio duro (U+00A0) o un salto de línea sobreviven a la
normalización, así que un valor que lleve uno dentro no coincide con nada, que es lo que le
corresponde a un valor que casi siempre llegó ahí por accidente.

**Qué cuenta como diacrítico en el paso 2.** La descomposición es la canónica de Unicode (la forma
NFD), y lo que se elimina de ella son las marcas combinantes (la categoría Mn). Quitar el
diacrítico nunca sustituye una letra por otra que no sea su base: `ñ` da `n` y `ü` da `u`, pero
`ß`, `ø` y `æ` se quedan como están, porque ninguna de las tres es una letra con un acento encima.
Una marca combinante suelta se elimina siempre, esté donde esté; un carácter precompuesto se
descompone si pertenece a los bloques latinos, griego o cirílico (U+00C0 a U+024F, U+0370 a U+04FF
y U+1E00 a U+1FFF), y fuera de ellos conserva su diacrítico. La regla se aplica a todo carácter de
esos bloques y no solo a las letras: catorce de ellos son signos y no letras (U+037E, U+0385,
U+0387, U+1FC1, U+1FCD a U+1FCF, U+1FDD a U+1FDF, U+1FED, U+1FEE, U+1FEF y U+1FFD), y también se
descomponen, aunque ninguno aparezca nunca dentro de un estado de un tablero real.

Un ejemplo con un tablero creado con los estados por defecto. La configuración de un tablero vive
dentro de su base de datos y no en un fichero que se edite a mano, así que se consulta con
["`biso config`"](cmd/config.md), y el trozo que importa aquí es este:

```
$ biso config list
project_name = My project
statuses = To Do,In Progress,Done
initial_status = To Do
active_status = In Progress
terminal_status = Done
...
```

Uno de sus tres estados es `To Do`. Con el algoritmo de arriba, estas entradas se resuelven así:

| Entrada | `normalizar` | Resultado |
|---|---|---|
| `To Do` | `todo` | coincide exactamente con `To Do`, por el paso a |
| `todo` | `todo` | coincide con `To Do`, por el paso c |
| `TODO` | `todo` | coincide con `To Do` |
| `To-Do` | `todo` | coincide con `To Do`, porque el guion se elimina |
| `TO_DO` | `todo` | coincide con `To Do`, porque el guion bajo se elimina |
| `to  do` | `todo` | coincide con `To Do`, porque los espacios se eliminan todos, sean uno o varios |
| `To Do.` | `todo.` | error 3: el punto no está entre los caracteres que se eliminan, así que `todo.` no es igual a `todo` ni a la forma normalizada de ningún otro estado (`inprogress`, `done`) |
| `To.Do` | `to.do` | error 3, por lo mismo: el punto se queda y `to.do` no coincide con ningún estado |

**No hay coincidencia por prefijo ni por parecido.** `coincidir()` nunca sugiere nada: o
encuentra una coincidencia exacta o normalizada, o es error 3. Lo que sí sugiere candidatos
parecidos es un algoritmo aparte, para el mensaje de error de las demás reglas de la
especificación que lo prometen.

**Un valor de solo espacios no llega hasta aquí.** Una cadena que solo lleva espacios es el valor
vacío, y quien la convierte en la cadena vacía es la capa que la lee (un flag, un fichero o la
entrada estándar), según ["El valor vacío"](valores-de-entrada.md#el-valor-vacío). `coincidir()`
recibe el valor ya resuelto y no vuelve a aplicar esa regla: cita en su mensaje exactamente lo que
le dan.

### Cuando el tablero tiene dos valores que se normalizan igual

El paso e es el único de los cinco que no culpa a quien escribe, sino al tablero: si sus estados
son `To Do` y `To-Do`, `todo` se parece igual a los dos y elegir uno sería inventarse cuál quiso
decir. El mensaje lo dice así, y lista solo los que empatan:

```
error: ambiguous status: "todo" matches 2 configured values: To Do, To-Do
hint: type one of them exactly, or rename one so the two no longer normalize the same
```

Su `code` es [`ambiguous_vocabulary`](contrato-json.md#los-identificadores-de-error), el mismo sea
cual sea el campo, porque quien ramifica sobre él no tiene que reaccionar distinto según cuál sea:
el remedio es el mismo. Su `field` es el campo, su `given` es lo que se
tecleó, y su `valid` lleva **solo los valores que empatan**, no el vocabulario entero, porque los
que no empatan no son la salida del problema. Si empatan más de dos, el mensaje dice cuántos son y
los lista todos.

**El paso a manda por encima de este**, que es lo que deja una salida sin tocar la configuración:
en ese mismo tablero, `To Do` y `To-Do` escritos exactamente resuelven cada uno al suyo, sin error.

**Dos valores configurados idénticos letra por letra son el mismo valor, no una ambigüedad.** Si la
configuración lista `To Do` dos veces, `todo` resuelve a `To Do` por el paso c, como si estuviera
una sola vez: el paso b descarta los duplicados exactos antes de contar, así que la ambigüedad del
paso e solo existe cuando los que se normalizan igual son valores **distintos**. Inventar un empate
entre un valor y su propia copia le pediría a quien escribe que desambiguara algo que no tiene dos
respuestas.

## El algoritmo de sugerencias más parecidas

Cinco sitios de la especificación prometen, en su mensaje de error, una lista de los valores
más parecidos a lo que se tecleó: `--label`, `--label-or` y `--not-label` (hasta cinco etiquetas, más
abajo en esta misma sección; y hasta cinco claves, cuando lo que se consulta es una clave con la forma
`clave:` de ["Consultar por la clave de una etiqueta con ámbito"](#consultar-por-la-clave-de-una-etiqueta-con-ámbito)), `--assignee` y `--not-assignee` (hasta cinco personas, igual), `biso config` con una clave
inexistente (hasta tres claves, sección ["`biso config`"](cmd/config.md#comportamiento-caso-a-caso)) y `biso help` con uno o
varios nombres de comando, si alguno no existe (hasta tres nombres, sección ["La ayuda"](cmd/help.md)). Estos casos
comparten un solo algoritmo, con el mismo espíritu que `coincidir()`: una sola regla, y solo
cambia el tope `N` según el caso. `--not-label` y `--not-assignee` reusan exactamente el mensaje y el
algoritmo de `--label`/`--label-or` y de `--assignee`, así que no cuentan como un sexto y un séptimo
sitio: son el mismo sitio, alcanzado por un flag distinto.

**La métrica es la distancia de Levenshtein, sobre la forma normalizada.** Dadas dos cadenas `a`
y `b`, la distancia de Levenshtein es el número mínimo de inserciones, eliminaciones o
sustituciones de un carácter que hace falta para convertir `a` en `b`. Se calcula con la matriz
de programación dinámica habitual: con `m = longitud(a)` y `n = longitud(b)`, una tabla `d` de
`(m+1) × (n+1)` donde `d[i][0] = i`, `d[0][j] = j` para todo `i, j`, y para `i, j > 0`:

```
d[i][j] = d[i-1][j-1]                                si a[i] == b[j]
d[i][j] = 1 + min(d[i-1][j], d[i][j-1], d[i-1][j-1])  si a[i] != b[j]
```

El resultado es `d[m][n]`.

```
sugerir(v, candidatos, N):
  1. calcular normalizar(v), la misma normalizacion que usa coincidir()
  2. para cada candidato c, calcular normalizar(c) y la distancia de Levenshtein
     entre normalizar(v) y normalizar(c)
  3. descartar los candidatos cuya distancia supere el umbral: la mitad de la
     longitud de normalizar(v), redondeada hacia arriba
  4. ordenar lo que queda por distancia ascendente; en caso de empate, por
     orden alfabetico de la forma normalizada
  5. devolver como maximo los N primeros que queden
```

**Si ningún candidato pasa el umbral, el resultado es una lista vacía, no un error distinto.**
El mensaje de error se queda con lo que ya dice sin sugerencia, igual que cuando el vocabulario
contra el que se sugiere está vacío (por ejemplo, un tablero sin ninguna etiqueta): forzar `N`
sugerencias cuando nada se parece de verdad haría más probable que se tomara una sin
comprobarla que ayudar a corregir el error.

Un ejemplo con las etiquetas `api, backend, bug, docs, frontend, infra, parser, security, ui,
urgent`:

| Entrada | `normalizar` | Umbral (mitad de la longitud, hacia arriba) | Qué pasa el umbral | Sugerencia (máximo 5) |
|---|---|---|---|---|
| `fronted` | `fronted` (7) | 4 | `frontend`, distancia 1 | `frontend` |
| `xyz` | `xyz` (3) | 2 | ninguna: la más cercana, `api`, está a distancia 3 | ninguna, lista vacía |

**El orden de salida es siempre por cercanía, nunca alfabético puro**, salvo para romper un
empate entre dos candidatos a la misma distancia. Si dos candidatos empatan en distancia **y además se
normalizan igual** (un tablero con las etiquetas `bar-code` y `Bar Code`), el desempate alfabético
de la forma normalizada tampoco los separa, así que decide su forma original. Es un tercer criterio
que no aporta ningún significado, solo existe para que la lista no dependa del orden en que quien
llama montó la suya.

## El mismo texto vale lo mismo en los dos sentidos

Esta tabla es el contrato, y es la prueba de aceptación que hay que poder ejecutar. Tablero con los
estados `To Do`, `In Progress` y `Done`:

| Entrada | `biso set MYP-1 --status <v>` | `biso ls --status <v>` |
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
       valid statuses on this board: To Do, In Progress, Done
```

En `type` y en `priority` es el mismo mensaje cambiando la palabra, en singular en la primera línea
y en plural en la segunda: `unknown type: "epic"` con `valid types on this board: ...`, y
`unknown priority: "urgent"` con `valid priorities on this board: ...`. La primera línea es el
`message` del objeto de error (["Los errores en JSON"](contrato-json.md#los-errores-en-json)); la
segunda no viaja en él, porque la lista que la compone ya está en `valid` y quien lee el JSON no
necesita la frase que la envuelve.

## Qué valida cada filtro, y contra qué

| Filtro | Conjunto contra el que valida | Si no encaja |
|---|---|---|
| `--status`, `--not-status`, `--type`, `--not-type`, `--priority` y `--not-priority` | el vocabulario configurado | error 3 |
| `--label`, `--label-or` y `--not-label` | el conjunto de etiquetas del tablero, definido abajo | error 3, con hasta cinco de las más parecidas (las que pasen el umbral de ["El algoritmo de sugerencias más parecidas"](#el-algoritmo-de-sugerencias-más-parecidas)) |
| `--label`, `--label-or` y `--not-label` con la forma `clave:` | el conjunto de claves del tablero, definido en ["Consultar por la clave de una etiqueta con ámbito"](#consultar-por-la-clave-de-una-etiqueta-con-ámbito) | error 3, con hasta cinco de las claves más parecidas |
| `--assignee` y `--not-assignee` | el conjunto de personas del tablero, definido abajo | error 3, con hasta cinco de las más parecidas (mismo algoritmo) |
| `--parent` | la resolución de referencias de la sección ["Cómo se resuelve una referencia a una tarea"](referencias.md) | error 2, 4 o 5 |
| `--author` | nada, no tiene vocabulario ni siquiera al leer (["El autor de una tarea"](modelo-de-datos/autor.md)) | nunca falla |
| `--ref` y `--not-ref` | nada, es texto libre | nunca falla |
| `--search` | nada, es texto libre | nunca falla |

**El conjunto de etiquetas del tablero** es la unión de las etiquetas declaradas en la clave `labels`
de la configuración y de las que lleva cualquier tarea del tablero, **incluidas las archivadas y las
que están en el estado terminal**. **El conjunto de personas se define con la clave `assignees` de la
configuración y con los valores de `assignees` de cualquier tarea, archivadas y terminadas incluidas.
Los valores de `author` no entran en este conjunto.** `--author` (["`biso ls`"](cmd/ls.md#parámetros))
compara con el mismo `normalizar()` de ["El algoritmo de coincidencia"](#el-algoritmo-de-coincidencia)
que usa `--assignee`, pero no lo resuelve contra ningún conjunto cerrado, porque `author` no tiene
vocabulario, ni siquiera al leer (["El autor de una tarea"](modelo-de-datos/autor.md)): compara el
normalizado del valor tecleado directamente contra el normalizado del `author` de cada tarea, sin el
paso de vocabulario que sí tiene `--assignee`. Una persona que solo consta como autora de la
tarea y nunca la ha tenido asignada sigue sin pertenecer al conjunto contra el que valida
`--assignee`, y un `--author` que ninguna tarea usa es una lista vacía, no un error.

Un tablero con las etiquetas `frontend`, `backend`, `bug`, `docs` y `parser`, y las personas
`@claude`, `@sara` y `@avilches`:

```
$ biso ls --label fronted
error: unknown label: "fronted"
hint: did you mean: frontend?

$ biso ls --assignee @clude
error: unknown assignee: "@clude"
hint: did you mean: @claude?
```

**Ni las etiquetas ni las personas tienen vocabulario cerrado al escribir.** Escribir una etiqueta
nueva la incorpora al conjunto, y a partir de ese momento filtrar por ella funciona.

Está el flag `--unchecked` de `biso ls` y `biso export`, que apaga **las comprobaciones contra
estos conjuntos, las de etiquetas y personas, y ninguna otra**: los vocabularios configurados de
`--status`, `--type` y `--priority` siguen validando, y `--parent` sigue resolviendo su referencia.
El flag no cambia ninguna otra cosa.

### Consultar por la clave de una etiqueta con ámbito

**`--label clave:` significa cualquier etiqueta de esa clave, con cualquier valor y con cualquiera de
los separadores.** `--label clave::` es exactamente el mismo filtro, porque el separador no distingue
nada al leer (["Las etiquetas con ámbito"](valores-de-entrada.md#las-etiquetas-con-ámbito)), y las dos
formas se aceptan igual en `--label-or` y en `--not-label`. Un filtro por la clave se combina con los
demás como cualquier otro valor de su flag: con `y` en `--label`, con `o` en `--label-or` y restando
en `--not-label`, con la misma regla que el resto de negaciones
(["`biso ls`"](cmd/ls.md#parámetros)).

Un filtro que nombra la etiqueta entera sigue validando contra el conjunto de etiquetas de arriba, y
tampoco ahí cuenta el separador: `--label milestone:m1` encuentra también las tareas etiquetadas
`milestone::m1`.

**El conjunto de claves del tablero** son las claves de las etiquetas declaradas en la clave `labels`
de la configuración y las claves de las etiquetas que lleva cualquier tarea, **archivadas y
terminadas incluidas**, con la misma regla que el conjunto de etiquetas. Una clave declarada cuenta
como conocida aunque ninguna tarea la use todavía, que es justo lo que hace útil declararla. Las
claves se comparan plegadas, así que `--label Milestone:` y `--label milestone:` son el mismo filtro.

Una clave que el tablero no tiene es error 3, con hasta cinco de las más parecidas:

```
$ biso ls --label milestne:
error: unknown label key: "milestne"
hint: did you mean: milestone?
```

Su `code` es `unknown_label_key`, con `field` igual a `label` y `valid` con las claves del tablero
(["Los identificadores de error"](contrato-json.md#los-identificadores-de-error)). Su `given` es el
filtro ya normalizado a un solo dos puntos (`milestne:`), exactamente el que viaja en
`data.filters.label` (o en `data.filters.notLabel` cuando el flag era `--not-label`)
(["Los filtros de `biso ls`"](contrato-json.md#los-filtros-de-biso-ls)), así que
un `--label Milestne::` se reprocha como `Milestne:`: la grafía tecleada de la clave se conserva y el
separador no, porque las dos formas son el mismo filtro y citar una de las dos sería elegir por el
lector.

**`--unchecked` apaga también esta comprobación**, igual que las de etiquetas y personas y ninguna
más: con él, `--label milestne:` o `--not-label milestne:` se aceptan y probablemente no devuelven
nada.

**El JSON de una tarea no lleva ningún campo derivado por clave**, ni en `task.get`, ni en
`task.list`, ni en la exportación: quien necesite el valor de una clave parte `labels` por los
primeros dos puntos, con la regla de análisis de
["Las etiquetas con ámbito"](valores-de-entrada.md#las-etiquetas-con-ámbito). El porqué está en
["Las etiquetas con ámbito"](../decisiones/detalles.md#las-etiquetas-con-ámbito).

---

