# `biso ls`

## Firma

```
biso ls [--status <status>]... [--not-status <status>]... [--any-status] [--archived] [--only-archived]
        [--type <v>]... [--not-type <v>]... [--priority <v>]... [--not-priority <v>]...
        [--label <label>]... [--label-or <label>]... [--not-label <label>]...
        [--assignee <@who>]... [--not-assignee <@who>]... [--mine] [--unassigned] [--author <@who>]...
        [--parent <ref>] [--root]
        [--blocked] [--not-blocked] [--waiting] [--not-waiting]
        [--active] [--not-active] [--overdue] [--due-before <date>]
        [--created-after <date>] [--created-before <date>]
        [--updated-after <date>] [--updated-before <date>]
        [--ref <text>]... [--not-ref <text>]... [--search <text>] [--unchecked]
        [--sort <field>] [--reverse] [--limit <n>] [--all] [--ids] [--count]
```

## Parámetros

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `--status <v>` | no | vocabulario | todos menos el terminal | sí | sí | `--any-status` |
| `--not-status <v>` | no | vocabulario | | sí | sí | `--any-status` |
| `--any-status` | no | booleano | falso | no | no | `--status`, `--not-status` |
| `--archived` | no | booleano | falso | no | no | `--only-archived` |
| `--only-archived` | no | booleano | falso | no | no | `--archived` |
| `--type <v>` | no | vocabulario | | sí | sí | |
| `--not-type <v>` | no | vocabulario | | sí | sí | |
| `--priority <v>` | no | vocabulario | | sí | sí | |
| `--not-priority <v>` | no | vocabulario | | sí | sí | |
| `--label <l>` | no | etiqueta | | sí | sí | |
| `--label-or <l>` | no | etiqueta | | sí | sí | |
| `--not-label <l>` | no | etiqueta | | sí | sí | |
| `--assignee <@w>` | no | persona | | sí | sí | `--mine`, `--unassigned` |
| `--not-assignee <@w>` | no | persona | | sí | sí | |
| `--mine` | no | booleano | falso | no | no | `--assignee`, `--unassigned` |
| `--unassigned` | no | booleano | falso | no | no | `--assignee`, `--mine` |
| `--author <@w>` | no | persona, sin vocabulario cerrado | | sí | sí | |
| `--parent <ref>` | no | referencia | | no | no | `--root` |
| `--root` | no | booleano | falso | no | no | `--parent` |
| `--blocked` | no | booleano | falso | no | no | `--not-blocked` |
| `--not-blocked` | no | booleano | falso | no | no | `--blocked` |
| `--waiting` | no | booleano | falso | no | no | `--not-waiting` |
| `--not-waiting` | no | booleano | falso | no | no | `--waiting` |
| `--active` | no | booleano | falso | no | no | `--not-active` |
| `--not-active` | no | booleano | falso | no | no | `--active` |
| `--overdue` | no | booleano | falso | no | no | |
| `--due-before <d>` | no | `YYYY-MM-DD` | | no | no | |
| `--created-after <d>` | no | `YYYY-MM-DD` | | no | no | |
| `--created-before <d>` | no | `YYYY-MM-DD` | | no | no | |
| `--updated-after <d>` | no | `YYYY-MM-DD` | | no | no | |
| `--updated-before <d>` | no | `YYYY-MM-DD` | | no | no | |
| `--ref <text>` | no | texto libre | | sí | sí | |
| `--not-ref <text>` | no | texto libre | | sí | sí | |
| `--search <text>` | no | texto libre | | no | no | |
| `--unchecked` | no | booleano | falso | no | no | |
| `--sort <field>` | no | `urgency`, `id`, `ordinal`, `due`, `updated`, `created`, `title`, `priority` | el orden de abajo | no | no | |
| `--reverse` | no | booleano | falso | no | no | |
| `--limit <n>` | no | entero >= 0 | 30 | no | no | `--all` |
| `--all` | no | booleano | falso | no | no | `--limit` |
| `--ids` | no | booleano | falso | no | no | `--count` |
| `--count` | no | booleano | falso | no | no | `--ids` |

Reglas de combinación de filtros:

- **Filtros de campos distintos se combinan con `y`.** `--status "To Do" --type bug` son las que cumplen las
  dos cosas.
- **Valores repetidos del mismo campo se combinan con `o`.** `--type bug --type docs` son las de
  cualquiera de los tipos. Esto vale para `--status`, `--type`, `--priority`, `--assignee`,
  `--author`, `--ref` y `--label-or`.
- **`--label` es la única que se combina con `y`.** `--label frontend --label bug` son las que llevan las
  dos. Para el `o` está `--label-or`, que valida igual.
- **`--label clave:` es cualquier etiqueta de esa clave**, con cualquier valor y con cualquiera de los
  separadores, y `--label clave::` es el mismo filtro
  (["Consultar por la clave de una etiqueta con ámbito"](../vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito)).
  Se combina como cualquier otro valor de su flag, así que `--label milestone: --label bug` son las
  tareas que tienen algún `milestone` **y** la etiqueta `bug`, y `--label-or milestone:`
  se suma con `o` a los demás valores de `--label-or`. Un filtro que nombra la etiqueta entera
  tampoco distingue el separador: `--label milestone:m1` encuentra las tareas con `milestone::m1`.
  **El identificador de una tarea en otro sistema, como una tarjeta de Trello, se consulta con esta
  misma forma** (`--label trello:CARD-123`, o `--label trello:` para cualquier tarea enlazada a
  Trello): no hay un filtro dedicado a esto porque el campo que lo guardaba, `ext`, se retiró en
  TASK-83 y quedó sustituido por las etiquetas con ámbito
  (["Se retira `ext`"](../../decisiones/detalles.md#se-retira-ext)), que ya resuelven llegar a una
  tarea desde ese identificador.
- **Las negaciones (`--not-status`, `--not-type`, `--not-priority`, `--not-label`, `--not-assignee` y
  `--not-ref`) no se combinan entre sí con el `o` de arriba: cada valor repetido resta por su cuenta.**
  `--not-type bug --not-type docs` descarta las dos a la vez, no solo una: sumar exclusiones nunca
  puede devolver una tarea que ya se había descartado. Una negación y el filtro positivo del mismo
  campo se combinan como cualquier par de campos distintos, con `y`, así que nombrar el mismo valor en
  los dos (`--type bug --not-type bug`) da una lista vacía, no un error: es la misma regla general de
  esta sección, más abajo, de que dos filtros que se contradicen por construcción son incompatibles
  mientras que una combinación de filtros válidos que resulte vacía es un hecho legítimo sobre el
  tablero, aplicada al caso en que los dos valores coinciden. Ninguna de las cuatro negaciones nuevas
  declara una incompatibilidad de flag: a diferencia de `--any-status`, que apaga el filtrado por
  estado entero y por eso sí choca con `--not-status`, no hay ningún `--any-type`, `--any-priority`,
  `--any-label` ni `--any-assignee` con los que colisionar. `--not-type` y `--not-priority` validan
  contra el vocabulario configurado, igual que `--type` y `--priority`; `--not-label` y
  `--not-assignee` validan contra el conjunto de etiquetas y el de personas del tablero, igual que
  `--label` y `--assignee`, con el mismo algoritmo de sugerencias
  (["Qué valida cada filtro, y contra qué"](../vocabularios.md#qué-valida-cada-filtro-y-contra-qué)).
  `--not-label` acepta también la forma `clave:`, con la misma regla que `--label`.
- **`--ref <texto>` filtra por subcadena sobre `references`, sin distinguir mayúsculas ni acentos**, con
  el mismo algoritmo de plegado que usa un selector de texto
  (["Selectores de criterios"](../familias-de-flags.md#selectores-de-criterios)): case-fold Unicode y
  descarte de las marcas combinantes de la forma NFKD. `--not-ref <texto>` es su negación, con el mismo
  patrón de nombre que `--not-status` y la misma regla de resta de la negación de arriba. Como desde
  TASK-75 y TASK-80 `references` es el único campo de punteros, `--ref` cubre por igual una ruta de
  fichero, una página de la especificación, una URL o el identificador de otra tarea: todo vive en la
  misma lista (["Los punteros: `references`"](../modelo-de-datos/relaciones.md#los-punteros-references)).
  No hay `--file` ni su negación, porque `modifiedFiles` ya no existe (se retiró en TASK-80) y `--ref`
  cubre también las rutas que antes guardaba. Ni `--ref` ni `--not-ref` tienen ningún vocabulario
  contra el que validar, así que un texto que ninguna tarea referencia da una lista vacía, nunca un
  error 3, la misma regla que `--search`.
- **El ámbito de `--search` y el de la resolución de referencias no incluyen `references`, y eso no
  cambia con `--ref`.** Es una decisión a propósito, con su razón completa en
  ["La búsqueda por texto"](../referencias.md#la-búsqueda-por-texto): meterlo en ese ámbito compartido
  sembraría ambigüedades de código 5 en `biso get` cada vez que dos tareas apuntaran al mismo sitio.
  `--ref` y `--not-ref` son el filtro dedicado que resuelve la consulta sin tocar ese ámbito.
- **`--author <@quien>` filtra por el autor de la tarea, con la misma comparación que `--assignee`: el
  `normalizar()` de ["El algoritmo de coincidencia"](../vocabularios.md#el-algoritmo-de-coincidencia)**
  (plegado de mayúsculas y minúsculas Unicode, descomposición NFD con descarte de diacríticos, y
  eliminación de espacio, tabulador, guion y guion bajo), aplicado por igual al valor de `--author` y
  al `author` de cada tarea. A diferencia de `--assignee`, `--author` no resuelve ese valor
  normalizado contra ningún conjunto configurado ni cerrado, porque no hay ninguno: `author` no tiene
  vocabulario ni siquiera al leer (["El autor de una tarea"](../modelo-de-datos/autor.md)). Por eso
  `--author` compara el normalizado del valor tecleado directamente contra el normalizado del
  `author` de cada tarea, tarea a tarea, en vez de resolverlo primero a una grafía canónica como hace
  `coincidir()` con `--assignee`; el resultado es el mismo algoritmo de plegado, sin el paso de
  vocabulario que no aplica aquí. Un `--author` que ninguna tarea usa es una lista vacía, nunca un
  error 3, y `--unchecked` no le afecta porque no hay ninguna comprobación que apagar. **No hay un
  equivalente de `--mine` para `--author`.** `--mine` existe porque la
  asignación es el eje que cambia con el trabajo diario y hay que preguntarlo en cada `biso ls`; el
  autor de una tarea se fija una sola vez, al crearla, y no vuelve a preguntarse en el mismo bucle de
  trabajo, así que escribir la identidad una vez con `--author @quien` (o leerla de `BISO_ME`
  (["Variables de entorno"](../invocacion.md#variables-de-entorno)) fuera del programa) no tiene la
  fricción diaria que `--mine` existe para quitar.
- **`--created-after`, `--created-before`, `--updated-after` y `--updated-before` toman una fecha
  `YYYY-MM-DD`, con la misma forma y el mismo error 2 que `--due-before` ante una fecha mal formada.**
  Comparan por la fecha de calendario en UTC de `createdAt` o `updatedAt`, ignorando la hora, la misma
  regla que ya fija "hoy" para la antigüedad de la urgencia
  (["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)). `--created-before` y `--updated-before`
  filtran estrictamente antes de la fecha dada, igual que `--due-before` filtra `due` anterior a la
  suya; `--created-after` y `--updated-after` filtran esa fecha **incluida** y en adelante, así que
  `--updated-after 2026-09-01 --updated-before 2026-09-08` es la primera semana de septiembre sin
  solapar con la siguiente. **A diferencia de `--overdue` y `--due-before`, estos cuatro filtros no
  tienen un caso de tablero sin ninguna fecha que comparar**: `createdAt` y `updatedAt` los pone el
  programa en cada tarea y nunca faltan
  (["El modelo de datos de una tarea"](../modelo-de-datos/index.md)), así que un `--created-before` o
  un `--updated-after` siempre tienen con qué comparar.
- **`--root` filtra las tareas sin padre**, el complemento de `--parent`, que da las hijas directas de
  una. Los dos son incompatibles entre sí por construcción: una tarea no puede ser a la vez "sin padre"
  y "hija de esta otra", así que `--root --parent <ref>` es siempre error 2, para cualquier valor de
  `<ref>`, y no una lista vacía que dependa del tablero.
- **`--unchecked` apaga la comprobación de existencia de `--label`, `--label-or`, `--not-label`,
  `--assignee` y `--not-assignee`, y ninguna otra.**
  No cambia cómo se combinan ni afecta a ningún otro filtro. Los vocabularios configurados siguen
  validando, y `--parent` sigue resolviendo su referencia.
- **Sin `--status` explícito, el estado terminal se excluye salvo `--any-status`; con `--status` explícito se
  filtra por ese valor tal cual, terminal incluido.** La exclusión del terminal es el comportamiento
  del valor por defecto de `--status`, no una regla aparte que se superponga a él: `biso ls --status Done`
  devuelve las tareas `Done`, exactamente como pide cualquier otro valor de `--status`. `--not-status <x>`
  sin `--status` sigue restando sobre la base por defecto (todos menos el terminal), así que por sí solo no
  reintroduce el terminal: `biso ls --not-status "To Do"` excluye `"To Do"` y sigue sin traer `Done`.
  `--any-status` sigue siendo la única forma de traer el terminal sin nombrarlo con `--status`.
- **Las archivadas se excluyen por defecto.** `--archived` las añade a las vivas y `--only-archived`
  deja solo las archivadas.
- **`--blocked` es incompatible con `--not-blocked`.** Los dos miran las dependencias sin terminar y
  no el estado, así que se combinan con cualquier filtro de estado y con los dos pares de abajo.
  `--not-blocked` por sí solo no dice que la tarea se pueda coger: descarta la que espera a otra
  tarea, no la que espera una respuesta ni la que ya lleva alguien. **Una tarea archivada sin
  terminar cuenta como terminada aquí**, igual que en la urgencia (["La
  urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)): no hace `--blocked` a quien depende de
  ella.
- **`--overdue` es `dias < 0`**, la misma cuenta que usa el término `proximidad` de la urgencia
  (["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)): una tarea que vence hoy tiene
  `dias = 0` y no es `--overdue`. **En un tablero donde ninguna tarea tiene `due`, `--overdue` no
  marca nunca ninguna**, porque no hay ninguna fecha con la que hacer esa cuenta.
- **`--due-before <date>` filtra las tareas cuya `due` es anterior a la fecha dada.** **En un
  tablero donde ninguna tarea tiene `due`, `--due-before` no filtra nunca nada**, por el mismo
  motivo que `--overdue`: no hay ninguna fecha que comparar.
- **`--waiting` es incompatible con `--not-waiting`, y `--active` con `--not-active`, cada uno con su
  opuesto.** `--active` y `--not-active` filtran por el papel del estado y no por su nombre, que es su
  razón de ser: sin ellos, pedir la cola activa obligaría a escribir `--status "In Progress"`, el nombre
  concreto de un tablero concreto, y la misma consulta dejaría de servir en otro. Los cuatro son
  compatibles con `--status`, con `--not-status` y con `--any-status`, porque filtran sobre el mismo eje sin
  contradecirse: `--status "To Do" --active` es una lista vacía en unos tableros y no en otros. La regla
  general: **dos filtros que se contradicen por construcción son incompatibles. Una combinación de
  filtros válidos que resulte vacía en este tablero es un hecho legítimo sobre el tablero, no un
  error.**

## La regla de orden, completa

Sin `--sort` se aplica el orden por defecto, que es esta tupla, en este orden y sin excepciones:

1. Las tareas que tienen `ordinal` van antes que las que no lo tienen.
2. Entre las que lo tienen, `ordinal` ascendente, comparado por puntos de código como el resto de las
   comparaciones de texto de esta regla (["El orden manual y su
   clave"](../modelo-de-datos/orden-manual.md)).
3. Entre las que no lo tienen, `urgency` descendente.
4. Cualquier empate se rompe por identificador ascendente, siempre. Dos tareas pueden llevar la misma
   clave de orden, y es ahí donde se desempatan.

Un `--sort` explícito sustituye los pasos 1 a 3 por ese campo, ascendente salvo `urgency`, que es
descendente por ser una medida de prioridad, y el paso 4 se sigue aplicando. `--reverse` invierte el
resultado final, incluido el desempate. **El orden nunca depende del estado**, porque el listado no
agrupa por estado.

**`--sort priority` no es una segunda excepción junto a `urgency`.** Ordena por la posición del nivel
de la tarea en `priorities`, la misma `i` de la regla de `prioridad` de la urgencia, 0-indexada desde
el nivel más urgente (["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)). Ascendente por `i`
pone primero la posición 0, que ya es el nivel más urgente, así que "ascendente" aquí ya significa "de
más a menos urgente" sin necesitar la excepción que sí hace falta con `urgency`, que crece hacia la
urgencia y no hacia la posición. Una tarea sin ninguna prioridad asignada va al final, en bloque,
ordenada por identificador, la misma regla que `--sort due` y `--sort ordinal` de arriba. **En un
tablero donde ninguna tarea tiene `priority`, `--sort priority` deja el tablero entero en ese bloque
final, ordenado por identificador**, la misma regla de determinismo que ya vale para `--sort due`.

Un `--sort due` o `--sort ordinal` sobre tareas que no tienen ese campo las pone al final, en bloque,
ordenadas por identificador. **`--sort ordinal` es, por tanto, el orden manual ascendente con las
tareas sin clave detrás**, que es el orden por defecto sin su paso 3: lo que cambia es que las tareas
sin clave se ordenan entre ellas por identificador y no por urgencia. `--reverse` lo invierte entero,
así que pone las tareas sin clave delante. **En un tablero donde ninguna tarea tiene `due`,
`--sort due` deja el tablero entero en ese bloque final, ordenado por identificador**: es la misma
regla de arriba para `--sort ordinal` sobre tareas sin ese campo, aplicada aquí al caso en que
ninguna tarea del tablero lo tiene.

`--sort title` compara los títulos por sus puntos de código Unicode, de menor a mayor, y no por las
reglas de intercalación de ningún idioma: una `Z` va antes que una `a`, y `ñ` va después de `z`. Es
la única comparación de texto del orden, y se fija así porque una intercalación local haría que el
mismo tablero se ordenara distinto según la configuración regional de la máquina, que es justo lo
que ["Los principios"](../principios.md) no admite. Para un título mal ordenado a ojo está
`--sort ordinal`, que es lo que existe para decidir un orden a mano.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Filtro con un valor fuera del vocabulario | Error 3, con la lista de válidos |
| `--label` o `--not-label` con una etiqueta, o `--assignee` o `--not-assignee` con una persona, que el tablero no tiene | Error 3, con hasta cinco de las más parecidas, igual que en la sección ["Qué valida cada filtro, y contra qué"](../vocabularios.md#qué-valida-cada-filtro-y-contra-qué) |
| `--label clave:` o `--not-label clave:` con una clave que el tablero no tiene | Error 3 con el `code` `unknown_label_key`, y hasta cinco de las claves más parecidas: `error: unknown label key: "milestne"` con `hint: did you mean: milestone?` |
| Lo mismo con `--unchecked`, en cualquiera de los tres casos anteriores | Se acepta, y probablemente no devuelve nada |
| `--label` con una etiqueta mal formada, por ejemplo `--label ":m"` | Error 2 con el `code` `malformed_label`, antes de mirar el tablero. Ni `--unchecked` lo apaga: no es una comprobación contra el tablero, es la forma del token (["Las etiquetas con ámbito"](../valores-de-entrada.md#las-etiquetas-con-ámbito)) |
| `--root` junto con `--parent` | Error 2: las dos se contradicen por construcción, para cualquier valor de `--parent` |
| `--author`, `--ref` o `--not-ref` con un valor que ninguna tarea tiene | Lista vacía, nunca error 3: ninguno de los tres tiene vocabulario contra el que validar |
| Filtro válido sin resultados | Ninguna línea por stdout, `note: no tasks match` por stderr, código **0** |
| Hay más resultados que el límite | Se imprimen los primeros y sale el aviso de recorte |
| `--limit 0` | No imprime ninguna fila, solo el aviso de recorte con el total. Es la forma de contar sin `--count` |
| `--count` | Un número por stdout y nada más: ni filas, ni aviso de recorte, ni la nota de "sin resultados" |
| `--ids` | Identificadores, uno por línea, sin cabeceras ni columnas. Es un listado, así que el límite y su aviso de recorte se aplican igual que con las columnas |
| Alguna tarea ilegible: un valor de `status`, `type` o `priority` que el tablero no declara, una fecha que no es una fecha, o cualquier otro motivo de ["Qué se comprueba"](../garantias.md#qué-se-comprueba) | Se salta, no entra en `--count`, y sale el aviso de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), y el resto del listado es válido. El aviso nombra **todas** las ilegibles del tablero, casen o no con los filtros: `--status Done` sin coincidencias imprime `note: no tasks match` y el aviso, nunca solo la nota |

## Salida

Ocho columnas fijas, separadas por dos espacios, en este orden y con estos contenidos:

| Columna | Contenido | Cuando está vacío |
|---|---|---|
| 1 | identificador | nunca lo está |
| 2 | estado | nunca lo está |
| 3 | tipo | `-` |
| 4 | prioridad | `-` |
| 5 | título, recortado a **100 celdas siempre**, con `...` al final si se recorta | nunca lo está |
| 6 | `ac <marcados>/<total>` | `-` si la tarea no tiene criterios de aceptación |
| 7 | primera persona asignada, con `+<n>` si hay más | `-` |
| 8 | fecha límite | `-` |

**La columna 8 es la primera candidata a recortar si algún día hace falta reducir el número de
columnas.** En un tablero donde ninguna tarea usa `due` queda vacía, con `-`, en todas las filas de
todas las llamadas: no hay ninguna fecha límite que mostrar. Cuánto se usa `due` hoy en esta máquina,
medido con su fecha, y por qué el campo se conserva de todas formas, está en la nota sobre
`urgency.due` de ["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia).

**El formato se calcula así, en dos pasos, siempre en este orden:**

1. El título de cada tarea se recorta primero a 100 celdas, **contando los tres puntos**, así que la
   cadena que se imprime no pasa nunca de 100: un título más largo deja 97 celdas suyas y `...`
   detrás. Esto pasa antes de calcular ningún ancho de columna.
2. Para cada una de las columnas 1 a 7, el ancho de esa columna es la anchura del valor más largo
   que le corresponde entre las filas que se van a imprimir en esta llamada, y cada valor se rellena
   con espacios a la derecha hasta ese ancho. **La columna 8 nunca se rellena**, porque es la última
   y no hay nada después que alinear.

**La unidad de los dos pasos es la celda de un terminal monoespaciado, no el carácter.** Un título es
texto libre en UTF-8 (["Codificación y texto"](../salida-y-terminal.md#codificación-y-texto)), así que puede llevar ideogramas, emoji o acentos combinantes, y esas tres
cosas ocupan en pantalla algo distinto de lo que suman sus puntos de código: una marca combinante mide
cero celdas porque se pinta sobre la letra anterior, un ideograma de Asia oriental o un emoji miden dos,
y todo lo demás mide una. La tabla que lo dice es la de Unicode, la de anchura de Asia oriental más la
categoría de las marcas combinantes, y contar en cualquier otra unidad desalinea la tabla en cuanto un
título deja de ser ASCII. **Los caracteres de formato (categoría Cf) miden también cero**, por el
mismo motivo que una marca combinante: no se dibujan. El caso que importa es el juntador de ancho
cero (U+200D) que une las piezas de un emoji compuesto, que no ocupa ninguna celda y que contarlo
como una desalinearía la fila.

**Medir en celdas no es mirar el terminal, así que no contradice la sección ["Interactividad, terminal y color"](../salida-y-terminal.md#interactividad-terminal-y-color).** La anchura de un
carácter es una propiedad de Unicode, la misma en cualquier máquina y con cualquier ventana, y por eso
la salida sigue sin depender de dónde se ejecute el programa. Lo que la sección ["Interactividad, terminal y color"](../salida-y-terminal.md#interactividad-terminal-y-color) prohíbe es lo otro: preguntar
cuántas columnas tiene la ventana, o si hay color, para decidir qué se imprime.

**Y el recorte nunca parte un grafema por la mitad.** Si cortar exactamente en la celda 97 separaría una
letra de su acento combinante, o partiría un emoji compuesto, se corta en la frontera anterior, así que
el título recortado puede medir 96 o 95 celdas en vez de 97. La promesa es el tope, nunca la longitud
exacta: la cadena impresa no pasa de 100 celdas.

Entre columna y columna van siempre **dos espacios literales**, se haya rellenado o no la columna
anterior. Con estas cuatro tareas, el título más largo mide 28 caracteres y por eso la columna 5 se
rellena a ese ancho, no a uno fijo:

```
MYP-7   To Do        bug   high    Crash on an empty repository  ac 0/4  -        2026-09-08
MYP-11  In Progress  bug   high    Normalize CRLF in the diff    ac 1/2  @claude  -
MYP-19  To Do        task  high    Retry the upload on 5xx       ac 0/2  -        -
MYP-23  To Do        docs  medium  Rewrite the install section   ac 0/1  @sara+1  -
```

Y por stderr, siempre que se haya recortado:

```
warning: 28 more tasks match; showing 30 of 58
hint: narrow with --status, --type or --label, or ask for everything with --all
```

Con `--ids`:

```
MYP-7
MYP-11
```

Con `--count`:

```
58
```

**`--count` no recorta nada, porque no imprime ninguna fila.** El número que da es el de las tareas
que encajan, no el de las que se habrían impreso, así que `--limit` no tiene nada que cortar y no hay
aviso de recorte que emitir: la llamada contesta `matched` y ya está. En el sobre JSON eso se ve
igual, con `data.tasks` vacío, `shown` y `hidden` a cero y `truncated` en `false`. `--ids`, en
cambio, sí es un listado: imprime las mismas filas que las columnas, con el mismo límite y el mismo
aviso, solo que con una columna en vez de ocho.

**No hay agrupación por estado.** El estado es una columna más, para que cada línea se pueda tratar
igual que las demás.

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.list",
  "generatedAt": "2026-09-06T13:26:41Z",
  "data": {
    "tasks": [
      {
        "id": "MYP-11",
        "title": "Normalize CRLF in the diff",
        "status": "In Progress",
        "type": "bug",
        "priority": "high",
        "assignees": ["@claude"],
        "author": "@avilches",
        "labels": ["parser"],
        "parent": null,
        "dependencies": [],
        "references": ["docs/bugs/BUG-02.md", "notes/a,b.md"],
        "due": null,
        "ordinal": null,
        "createdAt": "2026-09-06T09:12:04Z",
        "updatedAt": "2026-09-06T11:40:18Z",
        "leaseExpiresAt": "2026-09-06T15:40:18Z",
        "leaseHolder": "@claude",
        "acDone": 1,
        "acTotal": 2,
        "commentCount": 1,
        "urgency": 19.0,
        "blocks": ["MYP-40"],
        "blocked": false,
        "waiting": false,
        "leaseExpired": false,
        "archived": false
      }
    ],
    "shown": 30,
    "matched": 58,
    "hidden": 28,
    "truncated": true,
    "skipped": [],
    "sort": "default",
    "filters": {
      "status": ["To Do", "In Progress"],
      "notStatus": [],
      "anyStatus": false,
      "archived": false,
      "onlyArchived": false,
      "type": [],
      "notType": [],
      "priority": [],
      "notPriority": [],
      "label": [],
      "labelOr": [],
      "notLabel": [],
      "assignee": [],
      "notAssignee": [],
      "unassigned": false,
      "author": [],
      "parent": null,
      "root": false,
      "blocked": null,
      "waiting": null,
      "active": null,
      "overdue": false,
      "dueBefore": null,
      "createdAfter": null,
      "createdBefore": null,
      "updatedAfter": null,
      "updatedBefore": null,
      "ref": [],
      "notRef": [],
      "search": null,
      "unchecked": false
    },
    "warnings": [ { "code": "list_truncated", "shown": 30, "matched": 58 } ]
  }
}
```

**`truncated` sigue siendo la forma de saber que se recortó sin tener que mirar `warnings`**, pero
cuando `truncated` es `true`, `data.warnings` lleva el mismo `list_truncated` de la lista general de
avisos (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)), con la misma regla que cualquier
otro comando: el texto humano sale por stderr y el objeto estructurado va en el sobre.

**La forma completa de `data.filters`, con una clave por cada filtro de esta página y el porqué de
cada una, está en ["Los filtros de `biso ls`"](../contrato-json.md#los-filtros-de-biso-ls).**

La cifra de urgencia del ejemplo puede no ser esta; el motivo está en la sección
["Urgencia"](../modelo-de-datos/urgencia.md).

**`ordinal` viaja como cadena o como `null`, nunca como número**: es la clave de orden manual
(["El orden manual y su clave"](../modelo-de-datos/orden-manual.md)), y este sobre es uno de los
pocos sitios donde se puede leer, porque las columnas de texto no la enseñan. La tarea del ejemplo no
tiene ninguna; una que la tuviera traería, por ejemplo, `"ordinal": "m"`.

**El listado nunca trae el cuerpo de la tarea**: ni descripción, ni plan, ni notas, ni criterios, ni
comentarios. Para eso está `biso get`. Los campos derivados de la sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md) sí están todos,
`blocks` incluido, con la única excepción declarada ahí mismo: `blockedByCount` y `unblocksCount`, que
solo están en `task.get` porque calcular un cierre transitivo por cada fila de un listado de hasta 300
violaría el presupuesto de arranque
(["El modelo de datos de una tarea"](../modelo-de-datos/index.md#los-campos-derivados)). `truncated`
es explícito para que nadie tenga que comparar `shown` con `matched`, y `skipped` lleva los
identificadores de las tareas ilegibles que se han saltado.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Listado, incluso vacío o con tareas saltadas | 0 |
| Un valor de filtro no existe en el tablero | 3 |
| Flags incompatibles, `--limit` negativo, `--sort` inventado (`code` igual a `unknown_sort_field`), fecha mal formada | 2 |
| `--parent` a una tarea que no existe | 4 |
| `--parent` por texto con varias coincidencias, que además imprime las candidatas por stdout como cualquier otra referencia ambigua (["La búsqueda por texto"](../referencias.md#la-búsqueda-por-texto)) | 5 |
| `--mine` sin ninguna identidad configurada (["Variables de entorno"](../invocacion.md#variables-de-entorno)) | 6 |
| El almacén no responde | 8 |
| No hay tablero | 20 |

**`--ref`, `--not-ref` y `--author` nunca producen el código 3.** No tienen ningún vocabulario
cerrado contra el que validar, así que un texto que ninguna tarea referencia, o una persona que solo
consta como autora en ningún sitio, da una lista vacía por el código 0, la misma regla que ya sigue
`--search` (["Qué valida cada filtro, y contra qué"](../vocabularios.md#qué-valida-cada-filtro-y-contra-qué)).
`--not-type`, `--not-priority`, `--not-label` y `--not-assignee` sí pueden dar el código 3, con la
misma regla que sus filtros positivos.

## `biso ls --help`

```
Usage: biso ls [options]

List tasks, one per line. Shows 30 by default, hides the Done ones and the
archived ones, and says on stderr what it left out. A filter value the board
does not have is an error, never an empty list, so an empty list is a fact.

Filters (repeat or comma-separate; same field is OR, different fields are AND):
      --status <value>          configured status (default: all but the terminal)
      --not-status <value>      exclude a status
      --any-status              include the terminal status too
      --archived                include archived tasks
      --only-archived           only archived tasks
      --type <value>            configured type
      --not-type <value>        exclude a type
      --priority <value>        configured priority
      --not-priority <value>    exclude a priority
      --label <value>           label; several labels are ANDed. The form key:
                                matches any value of that scoped-label key
      --label-or <value>        label; several are ORed; takes key: too
      --not-label <value>       exclude a label; several are ORed (excluded if it
                                carries any of them); takes key: too
      --assignee <@who>         assignee
      --not-assignee <@who>     exclude an assignee
      --mine                    assigned to you
      --unassigned              assigned to nobody
      --author <@who>           task author
      --parent <ref>            subtasks of this task
      --root                    no parent
      --blocked                 something unfinished blocks it
      --not-blocked             nothing unfinished blocks it; it may still be
                                waiting on an answer, so add --not-waiting
      --waiting                 has an open question
      --not-waiting             has no open question
      --active                  in the board's active status
      --not-active              not in the active status
      --overdue                 past its due date
      --due-before <date>       due before YYYY-MM-DD
      --created-after <date>    created on or after YYYY-MM-DD
      --created-before <date>   created before YYYY-MM-DD
      --updated-after <date>    updated on or after YYYY-MM-DD
      --updated-before <date>   updated before YYYY-MM-DD
      --ref <text>              substring match on references
      --not-ref <text>          exclude a references substring match
      --search <text>           free text; see `biso get --help` for the scope
      --unchecked               do not check that the labels and assignees you
                                filter by exist on the board; nothing else
                                changes

Shape:
      --sort <field>            urgency, id, ordinal, due, updated, created, title,
                                priority
      --reverse                 flip the whole order, tie-breaks included
      --limit <n>               how many rows to print (default 30, 0 prints none)
      --all                     print every match
      --ids                     print only ids, one per line
      --count                   print only how many match

Columns: id, status, type, priority, title, criteria, assignee, due. Empty
cells print a dash. The title is cut at 100 characters, always, before any
column width is computed. Columns 1 to 7 are padded to the widest value
printed; column 8 never is. Two spaces always separate columns.

Exit codes:
  0  listed, even when empty      5  --parent matched several tasks
  2  bad usage                    6  --mine with no identity configured
  3  a filter value does not exist here
  4  --parent does not exist      8  the board could not respond
                                  20 no board here

Examples:
  biso ls
  biso ls --status "In Progress" --mine
  biso ls --type bug --priority high --limit 10
  biso ls --not-blocked --not-waiting --ids
  biso ls --any-status --archived --all
  biso ls --ref internal/ops/write.go --any-status
  biso ls --root --not-label blocked
```

---

