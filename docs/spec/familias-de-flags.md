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
comparten la forma "escalar". `list<string>` es "lista de tokens"; `text` es "bloque de prosa";
`map<string,string>` es "mapa de claves"; y `list<Criterion>` / `list<Comment>` son "lista de
objetos". **Dos formas se reparten en más de una fila** cuando, dentro de la misma forma, hay más de
un conjunto de operaciones posible: `acceptanceCriteria`, `definitionOfDone` y `comments` son los tres
"lista de objetos", pero los criterios y los comentarios no comparten los mismos flags, así que la
forma aparece dos veces, una por cada conjunto de operaciones.

| Forma | Clase de campo | Variantes |
|---|---|---|
| escalar | Escalar | fijar, vaciar |
| lista de tokens | Lista de tokens | añadir, quitar, vaciar, sustituir entera |
| bloque de prosa | Bloque de prosa | añadir al final, vaciar |
| mapa de claves | Mapa de claves | fijar una clave, quitar una clave, vaciar |
| lista de objetos | Criterios (`acceptanceCriteria`, `definitionOfDone`) | añadir, quitar, vaciar; y aparte, marcar y desmarcar (["Selectores de criterios"](#selectores-de-criterios)) |
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

Los criterios (`ac`, `dod`) y la prosa no tienen un flag de "sustituir entera" propia
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
| etiquetas | `-l, --add-labels` | `--rm-labels` | `--clear-labels` | `--replace-labels` |
| personas asignadas | `-a, --add-assignees` | `--rm-assignees` | `--clear-assignees` | `--replace-assignees` |
| referencias | `--add-refs` | `--rm-refs` | `--clear-refs` | `--replace-refs` |
| documentación | `--add-docs` | `--rm-docs` | `--clear-docs` | `--replace-docs` |
| dependencias | `--add-deps` | `--rm-deps` | `--clear-deps` | `--replace-deps` |
| ficheros tocados | `--add-files` | `--rm-files` | `--clear-files` | `--replace-files` |

Todas son repetibles, y admiten lista separada por comas además de repetición. Las de "añade" y las de
"sustituye" se acumulan igual dentro de la misma llamada: `--replace-labels a --replace-labels b` dejaría
la lista en `{a, b}`, no solo en `{b}`, exactamente como se acumula `--add-labels a --add-labels b`
(["Repetición y listas separadas por comas"](valores-de-entrada.md#repetición-y-listas-separadas-por-comas)). La diferencia entre añadir y sustituir no está en cómo se
escriben los valores, está en qué hacen con el resultado: uno se suma a lo que ya había, el otro lo
sustituye entero.

## Campos de lista sin coma (criterios)

| Campo | Añade | Quita (selector) | Vacía |
|---|---|---|---|
| criterios de aceptación | `--add-ac` | `--rm-ac` | `--clear-acs` |
| definición de hecho | `--add-dod` | `--rm-dod` | `--clear-dods` |

**No existe `--replace-ac` ni `--replace-dod`.** Sustituir la lista entera de criterios se hace
vaciando y añadiendo en la misma llamada (["Sustituir un campo que no tiene flag de \"sustituir entera\""](#sustituir-un-campo-que-no-tiene-flag-de-sustituir-entera)):
`biso set MYP-11 --clear-acs --add-ac "First" --add-ac "Second"`.

`--add-ac` y `--add-dod` son repetibles pero **no** aceptan lista por comas, porque el texto de un
criterio puede contener comas. `--rm-ac` y `--rm-dod` toman un selector de la sección
["Selectores de criterios"](#selectores-de-criterios).

**Los elementos nuevos se crean con claves nuevas, sin marcar, y las claves de los elementos
anteriores no se reutilizan.** Es coherente con
["Los criterios y sus claves estables"](modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables): la
clave se asigna al crear el elemento.

**Qué clave le toca a un elemento creado con `--add-ac` o `--add-dod` es algo que quien llama no
puede saber de antemano**, salvo en `biso new`: en cualquier otro comando de escritura, el contador de
esa lista ya venía de antes, y consultarlo exigiría leer la tarea primero. Por el principio 4
(["Los principios"](principios.md), "la salida por defecto de una escritura es lo que quien llama no sabía"), esa clave es
justo la clase de dato que la salida por defecto tiene que enseñar sin que haga falta pedirlo aparte.
La forma exacta de cómo se enseña, en la línea de estado de ["`biso set`"](cmd/set.md#salida) y no en `biso new`, está en esa misma
sección.

## Campos de prosa

| Campo | Añade al final | Vacía |
|---|---|---|
| descripción | `-d, --append-desc` | `--clear-desc` |
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

## Selectores de criterios

`--check-ac`, `--uncheck-ac`, `--rm-ac`, `--check-dod`, `--uncheck-dod` y `--rm-dod` toman un selector.
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

| Caso límite | Resultado |
|---|---|
| clave que no existe | error 4: `no acceptance criterion #7 on MYP-11 (keys: 1, 3)` |
| texto que no encaja con ninguno | error 4, con los textos de los elementos listados |
| texto que encaja con dos | error 5, con los dos listados |
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
cuerpo del comentario en vez del texto del criterio para la forma de texto:

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

**Esto no es la misma regla que `--ext <clave>=<valor>`, y no hace falta que lo sea.** `--ext` corta
por el primer `=` porque puede: el alfabeto de una clave de `ext` ya excluye el propio `=`
(["El juego de caracteres de un token"](valores-de-entrada.md#el-juego-de-caracteres-de-un-token)), así que el primer `=` de la cadena es siempre el único
`=` que puede separar la clave del valor, y da igual por cuál de los dos extremos se busque. El
selector de `--set-comment-date`, en cambio, puede ser un texto libre sin alfabeto cerrado, así que
necesita su propia regla de corte, y esa regla es "por el último" precisamente porque aquí sí puede
haber más de un `=` en la cadena.

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
| rango invertido | error 2 |
| `--set-comment-date` con un instante mal formado | error 2, señalando el formato ISO 8601 |
| la misma clave en dos `--set-comment-date` con instantes distintos | error 2, misma regla que un escalar repetido con valores distintos (["Repetición y listas separadas por comas"](valores-de-entrada.md#repetición-y-listas-separadas-por-comas)) |
| la misma clave en dos `--set-comment-date` con el mismo instante | se aplica una vez, sin aviso |
| `--rm-comment` y `--set-comment-date` sobre la misma clave en la misma llamada | error 2, detectado en la validación previa de arriba: borrar y corregir la fecha del mismo comentario a la vez es una petición contradictoria |
| `--rm-comment all` en una tarea sin comentarios | sin efecto, con `warning: MYP-11 has no comments`, igual que `--check-ac all` sin criterios (["Selectores de criterios"](#selectores-de-criterios)) |
| `--rm-comment all` o `--set-comment-date all=<instante>` sobre varias tareas | válido, cada tarea actúa sobre los suyos |
| una clave, un rango, una lista o un texto sobre varias tareas | error 2, misma regla que la de ["Selectores de criterios"](#selectores-de-criterios): el selector de una tarea no tiene por qué significar lo mismo en otra |

## Campos escalares

| Campo | Fija | Vacía |
|---|---|---|
| título | `-t, --title` | no se puede, es obligatorio |
| estado | `-s, --status` | no se puede, es obligatorio |
| tipo | `--type` | `--clear-type` |
| prioridad | `--priority` | `--clear-priority` |
| tarea padre | `-p, --parent` | `--clear-parent` |
| fecha límite | `--due` | `--clear-due` |
| orden manual | `--ordinal` | `--clear-ordinal` |
| autor de la tarea | `--author` | `--clear-author` |

**Un escalar guarda un único valor, así que fijarlo con su propio nombre nunca es ambiguo con
"añadir": no hay nada que añadir a un valor que no es una lista.** Por eso estos nombres se quedan sin
prefijo, a diferencia de los campos de lista: no es una excepción a la regla de esta sección, es la
regla aplicada a una forma de dato que solo admite una operación de escritura.

Un escalar **nunca** se borra pasándole la cadena vacía, según ["El valor vacío"](valores-de-entrada.md#el-valor-vacío).

## Campos externos

| Operación | Flag | Repetible |
|---|---|---|
| fijar una clave | `--ext <clave>=<valor>` | sí |
| quitar una clave | `--rm-ext <clave>` | sí |
| vaciar el mapa entero | `--clear-ext` | no |

**No existe `--replace-ext`.** Fijar una clave con `--ext` ya sustituye su valor, así que un segundo
flag para lo mismo solo serviría para equivocarse. Vaciar el mapa entero es `--clear-ext`, y es la
única forma de vaciarlo. Este campo ya era explícito antes del resto del rediseño de esta sección: no
cambia nada aquí.

---
