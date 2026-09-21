# Las familias de flags

Esta sección define de una vez la forma de todos los flags de escritura. Los comandos no la
repiten: cada uno dice qué campos acepta, y esta sección dice qué forma tiene cada campo.

## La regla

**Todo flag de escritura dice en su propio nombre qué hace, sin excepción.** No hay ningún nombre
"desnudo" cuyo significado dependa de una regla aparte que haya que conocer de antemano: quien lee
`--add-labels`, `--rm-labels`, `--clear-labels` o `--replace-labels` no necesita saber nada más que esas
cuatro palabras para saber qué hace cada una, ni falta que consultar esta sección para adivinarlo.

Cada clase de campo tiene exactamente las operaciones que tienen sentido para esa forma de dato, ni
una más ni una menos:

**La "Forma" de esta tabla es una clasificación de comportamiento de escritura, no el tipo del
campo.** Se deriva del tipo concreto de ["El modelo de datos de una tarea"](modelo-de-datos/index.md): a
efectos de qué flag tiene sentido ofrecer, da igual si un valor único es `string`, `enum(...)`,
`date`, `int`, `float` o `bool`, porque a todos les basta con fijar y vaciar, así que esos tipos
comparten la forma "escalar". `list<string>` es "lista de tokens"; `text` es "bloque de prosa"; y
`list<Criterion>` / `list<Comment>` son "lista de objetos".
**Una forma se reparte en más de una fila** cuando, dentro de la misma forma, hay más de un
conjunto de operaciones posible: `acceptanceCriteria` y `comments` son los dos "lista de objetos",
pero los criterios y los comentarios no comparten los mismos flags, así que la forma aparece dos
veces, una por cada conjunto de operaciones.

| Forma | Clase de campo | Variantes |
|---|---|---|
| escalar | Escalar | fijar, vaciar |
| lista de tokens | Lista de tokens | añadir, quitar, vaciar, sustituir entera |
| bloque de prosa | Bloque de prosa | añadir al final, vaciar |
| lista de objetos | Criterios (`acceptanceCriteria`) | añadir, quitar, vaciar; y aparte, marcar y desmarcar (["Selectores de criterios"](#selectores-de-criterios)) |
| lista de objetos | Comentarios (`comments`) | añadir (`--comment`), quitar uno o varios enteros, corregir solo su fecha; nunca editar cuerpo ni autor |

**`question` no entra en esta tabla.** Es de tipo `Question`, un valor único y no una lista, y esa
forma no tiene ninguna fila aquí porque no existe ningún flag de campo que la escriba: ninguna
de las demás filas la describe, y no hace falta una fila vacía solo para nombrarla.
**Ningún flag de campo escribe `question`**: lo escriben `biso ask`, `biso answer` y la importación
de `biso new --from`, y nadie más, igual que `archived` solo lo cambia `biso archive` (sección
["El modelo de datos de una tarea"](modelo-de-datos/index.md)).

**El significado no cambia entre comandos.** `--add-ac` añade un criterio en `biso new`, en `biso set`, en
`biso start` y en `biso finish`, y todos los comandos de escritura aceptan todos estos flags.

**Por qué no hay nombre desnudo.** Antes lo había: el nombre desnudo del campo añadía, y `set-`,
`rm-` y `clear-` delante cambiaban esa operación. Funcionaba, y de hecho resolvía un fallo real medido
en otra herramienta (Principio 3 de ["La evidencia detrás de los siete principios"](../decisiones/principios-y-mantenimiento.md#la-evidencia-detrás-de-los-siete-principios)), pero exigía conocer esa regla de antemano: la
única forma de saber qué hacía `--label` era haberla leído en algún sitio, porque el nombre por sí solo
no lo dice. Cada flag de esta sección lleva ahora su propio verbo (`add`, `rm`, `clear`,
`replace`, `append`, `check`, `uncheck`), así que no hay ninguna regla que aprender antes de usarla:
la razón completa, con la medición de por qué la forma anterior tampoco cabía ya en el mensaje de
arranque, está en ["El grid completo de flags de campo en el mensaje de arranque, medido con un agente real"](../decisiones/vocabulario-y-mensaje-de-arranque.md#el-grid-completo-de-flags-de-campo-en-el-mensaje-de-arranque-medido-con-un-agente-real).

## Sustituir un campo que no tiene flag de "sustituir entera"

Los criterios (`ac`) y la prosa no tienen un flag de "sustituir entera" propia
(["Campos de lista sin coma"](#campos-de-lista-sin-coma-criterios) y ["Campos de prosa"](#campos-de-prosa)). Sustituirlos se hace vaciando y añadiendo en la misma
llamada:

```
biso set MYP-11 --clear-labels --add-labels parser,urgent
biso set MYP-11 --clear-desc --append-desc "Reescrito entero"
biso set MYP-11 --clear-acs --add-ac "First" --add-ac "Second"
```

Cada una deja el campo con exactamente lo que se acaba de añadir, nunca mezclado con lo que hubiera
antes, y el orden en que se escriben los flags en la línea de comandos no importa: **el orden fijo
de aplicación dentro de una escritura, con los borrados siempre antes que los añadidos, está en
["Orden de aplicación dentro de una escritura"](garantias.md#orden-de-aplicación-dentro-de-una-escritura)** y es el mismo para todos los campos, no una regla aparte de esta
sección.

Ese orden también dice qué pasa si el mismo valor aparece a la vez en la parte que añade y en la
parte que quita de un campo: no es un error, gana el añadido, porque se resuelve después. Esto es
distinto de `--check-ac` y `--uncheck-ac` sobre el mismo criterio, que sí es un error
(["Selectores de criterios"](#selectores-de-criterios)): marcar y desmarcar el mismo criterio son dos peticiones contradictorias
sobre el mismo booleano y no hay un orden que las resuelva con sentido, mientras que quitar y volver
a añadir un elemento de una lista es una operación con un resultado bien definido.

## Campos de lista que admiten coma

| Campo | Añade | Quita (uno o varios) | Vacía | Sustituye entera |
|---|---|---|---|---|
| etiquetas | `--add-labels` | `--rm-labels` | `--clear-labels` | `--replace-labels` |
| personas asignadas | `--add-assignees` | `--rm-assignees` | `--clear-assignees` | `--replace-assignees` |
| referencias | `--add-refs` | `--rm-refs` | `--clear-refs` | `--replace-refs` |
| dependencias | `--add-deps` | `--rm-deps` | `--clear-deps` | `--replace-deps` |

Todas son repetibles, y admiten lista separada por comas además de repetición. Las de "añade" y las de
"sustituye" se acumulan igual dentro de la misma llamada: `--replace-labels a --replace-labels b` dejaría
la lista en `{a, b}`, no solo en `{b}`, exactamente como se acumula `--add-labels a --add-labels b`
(["Repetición y listas separadas por comas"](valores-de-entrada.md#repetición-y-listas-separadas-por-comas)). La diferencia entre añadir y sustituir no está en cómo se
escriben los valores, está en qué hacen con el resultado: uno se suma a lo que ya había, el otro lo
sustituye entero.

Por ejemplo, sobre una tarea que ya tiene la etiqueta `cli`, estas tres llamadas dejan exactamente las
mismas etiquetas, `cli`, `parser` y `urgent`:

```
biso set MYP-11 --add-labels parser,urgent
biso set MYP-11 --add-labels parser --add-labels urgent
biso set MYP-11 --add-labels parser,urgent
```

Y esta, en cambio, deja solo `parser` y `urgent`, porque sustituye la lista y `cli` desaparece:

```
biso set MYP-11 --replace-labels parser,urgent
```

Una etiqueta o una persona nunca llevan coma, pero una referencia sí puede, y entonces la coma se escapa con `\,`
(["Repetición y listas separadas por comas"](valores-de-entrada.md#repetición-y-listas-separadas-por-comas)):
`--add-refs 'notes/a\,b.md'` añade una sola referencia, `notes/a,b.md`.

**Las dependencias se validan al escribirlas, y solo al escribirlas.** Lo que significa una dependencia, y hacia dónde apunta, está en ["Las relaciones entre tareas"](modelo-de-datos/relaciones.md); aquí solo está lo que pasa al escribirla. `--add-deps` y
`--replace-deps` resuelven cada valor con la rutina de ["Cómo se resuelve una referencia a una
tarea"](referencias.md) y guardan el identificador al que resuelve, así que `--add-deps "CRLF"`
deja guardado `MYP-11` y no el texto que se tecleó, y una referencia que no existe o que encaja con
varias tareas falla con el código que le toca (4 o 5). `--rm-deps` y `--clear-deps` **no** resuelven
nada: quitan el valor literal, porque exigir que una dependencia exista para poder quitarla haría
imposible limpiar una que apunta a una tarea que ya no está, y porque quitar un valor que la lista no
tiene es el aviso tolerante de más abajo y no un error.

Una dependencia que fuera la propia tarea, o que cerrara un ciclo, es error 2, y el mensaje enseña el
camino que se cerraría para que no haya que reconstruirlo a mano. Lo mismo vale para `--parent`:

```
error: MYP-11 cannot depend on itself
error: --add-deps would close a dependency cycle: MYP-11 -> MYP-4 -> MYP-11
error: MYP-11 cannot be its own parent
error: --parent would close a parent cycle: MYP-11 -> MYP-4 -> MYP-11
```

Sus `code` son `self_dependency`, `dependency_cycle` y `parent_cycle`
(["Los identificadores de error"](contrato-json.md#los-identificadores-de-error)). **En una sola
tarea de `biso new` ninguno de los cuatro puede llegar a darse**, porque una tarea que todavía no
existe no tiene identificador y nada puede apuntar a ella; las comprobaciones se hacen igual, con la
misma función, y simplemente no encuentran nada. **El lote de `biso new --from` sí los produce**, y
ahí no salen con el código 2 sino con el 7 del informe del lote, porque cada uno se atribuye a una
línea del fichero y el fichero puede traer varios: una línea puede nombrar como `parent` o como
dependencia a una tarea que otra línea del mismo fichero crea, así que el ciclo existe en cuanto el
fichero se lee entero (["El modo lote"](cmd/new.md#el-modo-lote)).

**Añadir un valor que la lista ya tiene, o quitar uno que no tiene, avisa pero nunca falla.** Ninguna
de las dos operaciones exige leer la tarea primero para no fallar, que es justo lo que este diseño
evita en cualquier otro sitio donde hay una alternativa tolerante. Las dos terminan con código 0:

```
warning: --add-labels: "urgent" already present, kept once
warning: --rm-labels: "urgent" not present, nothing removed
```

El mismo patrón vale para cualquier fila de la tabla de arriba, `--rm-deps` incluido:
`warning: --rm-deps: "MYP-4" not present, nothing removed`. No es la misma regla que el aviso de
`--add-labels: "urgent" given twice, kept once` de más abajo: esa otra es un valor repetido dentro de
la misma llamada, y esta es un valor que ya estaba en la tarea antes de la llamada; las dos pueden
darse a la vez y cada una avisa por su cuenta.

**Una etiqueta o una persona asignada distinguen mayúsculas al guardar, y las ignoran al filtrar.**
`--add-labels Parser` y `--add-labels parser` en la misma tarea quedan como dos etiquetas distintas de
verdad: nunca se funden en silencio, porque el vocabulario de etiquetas y personas solo es cerrado al
leer y no al escribir, y fundirlas sería perder un dato que nadie pidió perder. Un
filtro de lectura, en cambio, no distingue: `biso ls --label parser` encuentra las tareas etiquetadas
`Parser` y las etiquetadas `parser` por igual. La comparación de lectura pliega mayúsculas y minúsculas
(case-fold Unicode), pero no toca acentos, porque una etiqueta o una persona son tokens cortos y no
prosa, y no comparten la regla de acentos de los selectores de texto de la sección
["Selectores de criterios"](#selectores-de-criterios).

### Escribir una etiqueta con ámbito

Una etiqueta que lleva `:` se analiza con la regla de
["Las etiquetas con ámbito"](valores-de-entrada.md#las-etiquetas-con-ámbito), que fija cuál es su clave,
cuál su valor y qué forma es mal formada. Lo que añade esta sección es lo único que el separador
decide al escribir: **una clave escrita con `::` deja como mucho una etiqueta suya en la tarea, y una
clave escrita con `:` admite todas las que se le pongan.**

**La exclusividad se comprueba sobre la lista que queda al final de la escritura**, no sobre la que
la tarea tenía al empezar. Con el orden fijo de
["Orden de aplicación dentro de una escritura"](garantias.md#orden-de-aplicación-dentro-de-una-escritura),
eso significa que los pasos que vacían, sustituyen y quitan ya se han aplicado cuando se juzga lo que
se añade, así que `biso set MYP-11 --rm-labels milestone::m1 --add-labels milestone:m2` funciona en
una sola llamada aunque la tarea empezara con una etiqueta exclusiva de esa clave. La comprobación
es de la fase de validación, así que cuando falla no se escribe nada.

| Caso | Qué pasa | Código |
|---|---|---:|
| `--add-labels k::v` sobre una tarea que conserva otras etiquetas de la clave `k` | `k::v` queda como única etiqueta de `k`: las demás se quitan, sean `k:x` o `k::y`, y el aviso las nombra una a una | 0 |
| `--add-labels k:v` sobre una tarea que conserva un `k::x`, incluso si `x` es el mismo `v` | Error 6, y no se escribe nada. El separador no es un detalle del valor: escribir `k:v` pide que la clave admita varios, y la tarea dice que admite uno | 6 |
| `k:a` y `k::b` de la misma clave en la misma llamada, en cualquier orden y repartidas como sea entre `--add-labels` y `--replace-labels` | Error 2. Los valores de `--rm-labels` no cuentan aquí, porque no escriben ninguna etiqueta: son justamente lo que deja sitio a la que se añade | 2 |
| `k::a` y `k::b` en la misma llamada | Gana la última escrita en la línea de comandos, con aviso | 0 |
| `--rm-labels k:v` | Quita la etiqueta de esa clave y ese valor, sea `k:v` o `k::v`: al quitar, el separador no cuenta | 0 |
| `--rm-labels k:` o `--rm-labels k::` | Error 2, etiqueta mal formada: la forma sin valor es sintaxis de filtro y no de escritura, y para vaciar la lista entera está `--clear-labels` | 2 |

Los avisos de esos casos, con su `code` en ["Notas y avisos"](salida-y-terminal.md#notas-y-avisos):

```
warning: --add-labels: "size::m" replaced size::s, size:l on MYP-11
warning: --add-labels: key "size" given twice with ::, kept "size::b"
```

El primero nombra lo que quitó en el orden en que la tarea lo tenía guardado, porque las listas nunca
se ordenan solas. El segundo cuenta las apariciones cuando son más de dos, igual que
`duplicate_flag_value`: `key "size" given 3 times with ::, kept "size::c"`.

Y los mensajes que rechazan la escritura:

```
error: MYP-11 already has "size::s", and :: allows at most one value of the key "size"
hint: drop it first, as in --rm-labels size::s --add-labels size:m

error: "size:a" and "size::b" mix the two separators of the key "size"
hint: a key takes either several values with :, or at most one with ::
```

Sus `code` son `exclusive_label_conflict` (código 6) y `mixed_label_separators` (código 2), los dos en
["Los identificadores de error"](contrato-json.md#los-identificadores-de-error). El segundo nombra los
dos valores en el orden en que se escribieron en la línea de comandos y **no culpa a ninguno de los
dos**, porque ninguno lo es más que el otro; qué claves lleva su objeto de error está en
["Los errores en JSON"](contrato-json.md#los-errores-en-json).

**Nada de esto es propio de `--add-labels`.** `--replace-labels` deja la lista que se le da y la misma
regla la juzga entera, así que `--replace-labels k:a,k::b` es el mismo error 2 y
`--replace-labels k::a,k::b` deja `k::b` con el mismo aviso. Lo que sí cambia con `--replace-labels`
y con `--clear-labels` es que la tarea no conserva nada de antes, así que una etiqueta exclusiva que
estuviera guardada no puede entrar en conflicto con nada: se fue en su propio paso.

**Y la misma regla vale para cada línea de un lote de `biso new --from`**, donde no hay flags sino la
lista `labels` de la línea, con una diferencia declarada en ["El modo lote"](cmd/new.md#el-modo-lote):
ahí dos valores `::` de la misma clave son un fallo de validación en vez de quedarse con el último.

**Si la lista `labels` de la configuración restringe la clave**, escribir un valor que no declara, o
la clave con el otro separador, es error 3, y esa comprobación va antes que todo lo de arriba
(["La lista `labels`"](cmd/config.md#la-lista-labels)).

## Campos de lista sin coma (criterios)

| Campo | Añade | Quita (selector) | Vacía |
|---|---|---|---|
| criterios de aceptación | `--add-ac` | `--rm-ac` | `--clear-acs` |

**No existe `--replace-ac`.** Sustituir la lista entera de criterios se hace
vaciando y añadiendo en la misma llamada (["Sustituir un campo que no tiene flag de \"sustituir entera\""](#sustituir-un-campo-que-no-tiene-flag-de-sustituir-entera)):
`biso set MYP-11 --clear-acs --add-ac "First" --add-ac "Second"`.

`--add-ac` es repetible pero **no** acepta lista por comas, porque el texto de un criterio puede
contener comas. `--rm-ac` toma un selector de la sección
["Selectores de criterios"](#selectores-de-criterios).

Para añadir varios criterios en una llamada, se repite el flag, uno por criterio:

```
biso set MYP-11 --add-ac "The parser accepts CRLF" --add-ac "Dates keep their time zone" \
                --add-ac "Reviewed by someone else"
```

Cada `--add-ac` de esa llamada añade un criterio de aceptación. Una coma dentro del texto no parte
nada: `--add-ac "Handles CRLF, LF and CR"` es un solo criterio.

**Los elementos nuevos se crean con claves nuevas, sin marcar, y las claves de los elementos
anteriores no se reutilizan.** Es coherente con
["Los criterios y sus claves estables"](modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables): la
clave se asigna al crear el elemento.

**Qué clave le toca a un elemento creado con `--add-ac` es algo que quien llama no
puede saber de antemano**, salvo en `biso new`: en cualquier otro comando de escritura, el contador de
esa lista ya venía de antes, y consultarlo exigiría leer la tarea primero. Por el principio 4
(["Los principios"](principios.md), "la salida por defecto de una escritura es lo que quien llama no sabía"), esa clave es
justo la clase de dato que la salida por defecto tiene que enseñar sin que haga falta pedirlo aparte.
La forma exacta de cómo se enseña, en la línea de estado de ["`biso set`"](cmd/set.md#salida) y no en `biso new`, está en esa misma
sección.

## Campos de prosa

| Campo | Añade al final | Vacía |
|---|---|---|
| descripción | `--append-desc` | `--clear-desc` |
| plan | `--append-plan` | `--clear-plan` |
| notas | `--append-note` | `--clear-notes` |
| resumen final | `--append-summary` | `--clear-summary` |

**No existe un flag que sustituya un bloque de prosa entero.** Un bloque de prosa es un único
texto (["El modelo de datos de una tarea"](modelo-de-datos/index.md)), sin elementos direccionables que
quitar uno a uno, así que no hay `--rm-*` para ninguno de los campos de la tabla de arriba. Sustituirlo
entero se hace vaciando y añadiendo en la misma llamada: `biso set MYP-11 --clear-plan --append-plan "Nuevo plan"`.

`--append-note` está en singular porque cada llamada añade un párrafo; `--clear-notes` está en plural
porque vacía el campo `notes` entero. No es una excepción a ninguna regla: cada palabra nombra
exactamente lo que el flag hace.

- Añadir a un campo vacío es lo mismo que fijarlo, así que al crear una tarea da igual usar `--append-*`
  o dejarlo vacío y añadir después: el resultado es el mismo.
- Al añadir sobre contenido existente se intercala una línea en blanco, y cada repetición del
  flag en la misma invocación produce su propio párrafo.
- Añadir un valor vacío no hace nada y avisa, según ["El valor vacío"](valores-de-entrada.md#el-valor-vacío).

Por ejemplo, sobre una tarea sin notas:

```
biso set MYP-11 --append-note "Rewrote the date parser"
biso set MYP-11 --append-note "CRLF still fails on Windows" --append-note "Asked @sara about it"
```

Después de las dos llamadas, el campo `notes` tiene tres párrafos separados por una línea en blanco:

```
Rewrote the date parser

CRLF still fails on Windows

Asked @sara about it
```

Un texto con varias líneas se pasa entre comillas con los saltos de línea dentro, o, más cómodo para
textos largos, desde un fichero con `@` o desde la entrada estándar con `-`
(["Tres formas de pasar un valor largo"](valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo)):

```
biso set MYP-11 --append-plan "1. Reproduce the CRLF failure
2. Normalize line endings before parsing
3. Add a test per platform"

biso set MYP-11 --append-plan @plan.md

cat plan.md | biso set MYP-11 --append-plan -
```

Y para sustituir el plan entero, se vacía y se añade en la misma llamada:

```
biso set MYP-11 --clear-plan --append-plan @plan.md
```

## Selectores de criterios

`--check-ac`, `--uncheck-ac` y `--rm-ac` toman un selector.
Todos son repetibles.

**Las claves de un selector son las mismas que enseña `biso get`, y no hay otro sitio donde
consultarlas.** Un criterio se lista como `- [x] #1 The diff ignores CRLF` (["`biso get`"](cmd/get.md)); ese `#1` es
literalmente la clave que acepta el selector, tanto en la salida de texto como en el JSON (`"key": 1`).

| Selector | Ejemplo | Qué elige |
|---|---|---|
| `all` | `--check-ac all` | todos los elementos de esa lista en esa tarea |
| una clave | `--check-ac 3` | el elemento `#3` |
| un rango de claves | `--check-ac 1-4` | las claves de la 1 a la 4 que existan |
| varias claves | `--check-ac 1,3,7` | esas tres |
| texto | `--check-ac "covers CRLF"` | el elemento cuyo texto contenga ese fragmento |

**La regla de desambiguación, que hay que implementar tal cual.** El valor se trata como lista de
claves **solo si el valor entero** encaja con `^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$`. En cualquier otro
caso es un texto literal, comas incluidas. Así, `--check-ac "1, 2 and the last one"` es una búsqueda de
texto que no encontrará nada y dará error 4, en vez de convertirse en algo a medias.

**Un selector de texto ignora mayúsculas y acentos.** La comparación se hace en dos pasos, sobre el
fragmento buscado y sobre el texto de cada elemento por igual: primero se pliegan mayúsculas y
minúsculas (case-fold Unicode), y luego se aplica normalización NFKD y se descartan las marcas
combinantes (categoría Unicode `Mn`), que es lo que quita los acentos sin tocar el resto del
carácter. Así, `--check-ac "codigo"` encuentra un criterio guardado como `El código ignora CRLF`, y
`--check-ac "CRLF"` encuentra uno guardado como `Handles crlf`. La regla de desambiguación de arriba se
aplica primero, sobre el valor tal cual llega, y solo si el resultado es un texto se le aplica este
plegado; una lista de claves nunca lo necesita.

| Caso límite | Resultado |
|---|---|
| clave que no existe | error 4: `no acceptance criterion #7 on MYP-11 (keys: 1, 3)` |
| texto que no encaja con ninguno | error 4, con los textos de los elementos listados |
| texto que encaja con dos | error 5, con los dos listados |
| texto que solo encaja ignorando mayúsculas o acentos, por ejemplo `"codigo"` contra `"El código..."` | encaja igual que si coincidiera carácter a carácter |
| rango donde faltan claves intermedias | se aplican las que hay, sin aviso |
| rango invertido, `4-1` | error 2 |
| marcar un elemento ya marcado | se queda marcado, sin aviso, la operación es idempotente |
| `--check-ac all` en una tarea sin criterios | sin efecto, con `warning: MYP-11 has no acceptance criteria` |
| `--check-ac all` sobre varias tareas | válido, cada tarea marca los suyos |
| una clave, un rango, una lista o un texto sobre varias tareas | error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |

**La regla de solape se aplica sobre el conjunto ya resuelto, no sobre el texto del selector.** Si
después de resolver `--check-ac` y `--uncheck-ac` un mismo elemento aparece en los dos conjuntos, es error
2, y da igual que se haya escrito `--check-ac 3 --uncheck-ac 3` o `--check-ac all --uncheck-ac 3`:

```
error: --check-ac and --uncheck-ac both select acceptance criterion #3 of MYP-11
```

Esta es la única familia de flags de esta sección donde el solape es un error en vez de resolverse
por orden (["Sustituir un campo que no tiene flag de \"sustituir entera\""](#sustituir-un-campo-que-no-tiene-flag-de-sustituir-entera) explica por qué).

**Un selector de texto que falla lista los elementos debajo de su mensaje**, indentados dos espacios y
con la clave de cada uno delante, que es la clave que el selector siempre acepta. El de cero
coincidencias sale con código 4 y el de dos o más con código 5:

```
error: no acceptance criterion of MYP-11 matches "covers CRLF"
  #1 The diff ignores CRLF
  #3 There is a test

error: "test" matches 2 acceptance criteria of MYP-11
  #3 There is a test
  #4 There is a second test
```

Un rango invertido nombra el trozo que lo está, y no el selector entero:

```
error: --check-ac: inverted range: "4-1"
hint: a range goes from the lower key to the higher one, as in 1-4
```

Y un selector que no es `all` sobre varias tareas dice cuál se escribió, porque el remedio es
cambiarlo:

```
error: --check-ac: with several tasks the selector has to be all, and this one is "3"
hint: the keys of the acceptance criteria of one task do not name the same thing in another
```

## Comentarios

| Operación | Flag | Repetible |
|---|---|---|
| añadir | `--comment <text>` | sí |
| borrar uno o varios enteros | `--rm-comment <sel>` | sí |
| corregir solo la fecha de uno o varios | `--set-comment-date <sel>=<instante>` | sí |

**No existe un flag que edite el cuerpo o el autor de un comentario ya escrito, y no va a
existir.** Un comentario es el registro de una conversación, y lo único que se concede aquí es
corregir un metadato (la fecha) o retirar el comentario entero, nunca reescribir lo que se dijo. La
razón, con el caso medido que la motiva, está en ["Borrar o corregir la fecha de un comentario"](../decisiones/detalles.md#borrar-o-corregir-la-fecha-de-un-comentario).

**`--rm-comment` y `--set-comment-date` toman el mismo selector que `--rm-ac` y `--check-ac`**
(["Selectores de criterios"](#selectores-de-criterios)), con la clave de un comentario en vez de la de un criterio y el
cuerpo del comentario en vez del texto del criterio para la forma de texto, y **heredan la misma
regla de plegado de mayúsculas y acentos** que esa sección fija para el fragmento de texto: un
`--rm-comment "codigo"` encuentra un comentario cuyo cuerpo dice "El código..." igual que lo hace
`--check-ac` con un criterio.

```
biso set MYP-11 --rm-comment 3
biso set MYP-11 --rm-comment all
biso set MYP-11 --set-comment-date 3=2026-08-14T10:22:00Z
```

**No existe `--clear-comments`.** `--rm-comment all` ya vacía la lista, y esta familia no necesita
una segunda forma de decir lo mismo: a diferencia de los criterios, que tienen `--clear-acs` además de
`--rm-ac` (["Campos de lista sin coma (criterios)"](#campos-de-lista-sin-coma-criterios)), aquí no hace falta la redundancia porque nada en esta
familia sustituye la lista entera de un tirón, y vaciarla del todo es un gesto tan deliberado como
borrar uno por uno.

**`--set-comment-date` toma un instante UTC completo (`YYYY-MM-DDTHH:MM:SSZ`), no una fecha suelta.**
`createdAt` es un instante, no un día, así que la corrección tiene que poder fijar la hora y no solo la
fecha; `--due` es distinto porque `due` sí es un día (["El modelo de datos de una tarea"](modelo-de-datos/index.md)).

**Dónde corta el `=` de `--set-comment-date`, cuando el selector es un texto que a su vez puede traer
el signo `=`.** El valor se divide por el **último** `=` de la cadena, nunca por el primero: todo lo
que queda a la derecha tiene que cumplir el formato de instante de arriba, y si no lo cumple es error
2 de instante mal formado, aunque el fragmento de texto de la izquierda contenga otro `=` suelto. Es
la misma familia de problema que ya resuelve la regla de desambiguación de ["Selectores de criterios"](#selectores-de-criterios) (clave
o texto, según si el valor entero encaja con la gramática de claves), solo que aquí el corte no
depende de una gramática cerrada sino de que el instante tiene una forma fija y reconocible. Un cuerpo
de comentario que termine literalmente en algo con forma de instante detrás de un `=` es el único caso
que esta regla no puede resolver por texto; para ese caso, la clave sigue siendo el selector que
siempre funciona.

**El solape entre `--rm-comment` y `--set-comment-date` se detecta antes de aplicar ninguno de los
dos, no durante el orden de escritura.** Caen en pasos distintos de ["Orden de aplicación dentro de una escritura"](garantias.md#orden-de-aplicación-dentro-de-una-escritura)
(`--rm-comment` en el 3, `--set-comment-date` en el 7), así que si se dejara que cada uno resolviera
su selector en su propio paso, `--rm-comment` ya habría borrado el comentario para cuando
`--set-comment-date` intentara corregirle la fecha, y el resultado sería un error 4 de "no existe" en
vez de un conflicto. Para que la regla no dependa de ese orden, cada selector se resuelve contra la
lista de comentarios de **antes** de la escritura, en la fase de validación que exige el principio 6
(["Los principios"](principios.md)), exactamente como ya hace la regla de solape de `--check-ac`/`--uncheck-ac`
(["Selectores de criterios"](#selectores-de-criterios)) sobre el conjunto ya resuelto: si una misma clave aparece en los dos
selectores resueltos, es error 2 y no se aplica ni el borrado ni la corrección.

| Caso límite | Resultado |
|---|---|
| clave que no existe | error 4: `no comment #7 on MYP-11 (keys: 1)` |
| texto que no encaja con ningún comentario | error 4, con los cuerpos de los comentarios listados |
| texto que encaja con dos o más | error 5, con los dos listados |
| texto que solo encaja ignorando mayúsculas o acentos | encaja igual que si coincidiera carácter a carácter, misma regla que ["Selectores de criterios"](#selectores-de-criterios) |
| rango invertido | error 2 |
| `--set-comment-date` con un instante mal formado | error 2, señalando el formato ISO 8601: `error: --set-comment-date: invalid instant: "2026-08-14"`, con `hint: an instant is written YYYY-MM-DDTHH:MM:SSZ, in UTC` y el `code` `invalid_date` |
| la misma clave en dos `--set-comment-date` con instantes distintos | error 2, misma regla que un escalar repetido con valores distintos (["Repetición y listas separadas por comas"](valores-de-entrada.md#repetición-y-listas-separadas-por-comas)) |
| la misma clave en dos `--set-comment-date` con el mismo instante | se aplica una vez, sin aviso |
| `--rm-comment` y `--set-comment-date` sobre la misma clave en la misma llamada | error 2, detectado en la validación previa de arriba: borrar y corregir la fecha del mismo comentario a la vez es una petición contradictoria. El mensaje es `error: --rm-comment and --set-comment-date both select comment #3 of MYP-11`, con el `code` `comment_selector_overlap` |
| `--rm-comment all` en una tarea sin comentarios | sin efecto, con `warning: MYP-11 has no comments`, igual que `--check-ac all` sin criterios (["Selectores de criterios"](#selectores-de-criterios)) |
| `--rm-comment all` o `--set-comment-date all=<instante>` sobre varias tareas | válido, cada tarea actúa sobre los suyos |
| una clave, un rango, una lista o un texto sobre varias tareas | error 2, misma regla que la de ["Selectores de criterios"](#selectores-de-criterios): el selector de una tarea no tiene por qué significar lo mismo en otra |

## Campos escalares

| Campo | Fija | Vacía |
|---|---|---|
| título | `--title` | no se puede, es obligatorio |
| estado | `--status` | no se puede, es obligatorio |
| tipo | `--type` | `--clear-type` |
| prioridad | `--priority` | `--clear-priority` |
| tarea padre | `--parent` | `--clear-parent` |
| fecha límite | `--due` | `--clear-due` |
| orden manual | `--ordinal` | `--clear-ordinal` |
| autor de la tarea | `--author` | `--clear-author` |

**Un escalar guarda un único valor, así que fijarlo con su propio nombre nunca es ambiguo con
"añadir": no hay nada que añadir a un valor que no es una lista.** Por eso estos nombres se quedan sin
prefijo, a diferencia de los campos de lista: no es una excepción a la regla de esta sección, es la
regla aplicada a una forma de dato que solo admite una operación de escritura.

Un escalar **nunca** se borra pasándole la cadena vacía, según ["El valor vacío"](valores-de-entrada.md#el-valor-vacío).

**Dos de estos escalares no son texto libre y rechazan lo que no cumple su forma**, los dos con
código 2:

```
error: --due: invalid date: "20/09/2026"
hint: a due date is written YYYY-MM-DD

error: --ordinal: not a whole number: "first"
```

Su `code` es `invalid_date` y `invalid_number` respectivamente, los mismos que ya llevan la fecha mal
formada y el `ordinal` negativo de ["El modelo de datos de una tarea"](modelo-de-datos/index.md).
`--due` con una fecha ya pasada, en cambio, no es un error: se acepta con el aviso `due_in_past`
(["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)).

## Casos límite de añadir, quitar y fijar

| Caso límite | Resultado | Código |
|---|---|---|
| `--add-labels`, o cualquier otro `--add-*`/`--append-*` de lista de tokens, con un valor que la tarea ya tiene | Se queda igual, sin duplicar, con `warning: --add-labels: "urgent" already present, kept once` | 0 |
| `--rm-labels`, `--rm-deps` o cualquier otro `--rm-*` de lista de tokens, sobre un valor que la tarea no tiene | Sin efecto, con `warning: --rm-labels: "urgent" not present, nothing removed` | 0 |
| Mayúsculas en una etiqueta o una persona asignada, por ejemplo `--add-labels Parser --add-labels parser` | Quedan como dos valores distintos al guardar; un filtro de lectura como `ls --label parser` encuentra los dos | 0 |
| Mayúsculas y acentos en el selector de texto de un criterio o de un comentario | Se pliegan las mayúsculas y se descartan los acentos antes de comparar (normalización NFKD, sin marcas combinantes); no cambia si el resultado es 0, 4 o 5, solo qué encuentra | sin cambio |
| Una etiqueta con ámbito en cualquiera de los flags de etiquetas | Lo decide ["Escribir una etiqueta con ámbito"](#escribir-una-etiqueta-con-ámbito), según el separador y según lo que la tarea conserve tras los `--rm-labels` de la misma llamada | 0, 2, 3 o 6 |

---
