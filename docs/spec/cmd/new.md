# `biso new`

## Firma

```
biso new [<title>] [--start] [--from <file|->] [cualquier flag de campo de las familias de flags, `--title` incluido]
```

## Parámetros propios

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `<title>` | sí, salvo con `--title` o `--from` | texto | | no | no | `--title`, `--from` |
| `--start` | no | booleano | falso | no | no | `--status`, `--from` |
| `--from <file|->` | no | ruta o `-` | | no | no | `<title>` y todas las de campo |

Todos los flags de campo de la sección ["Las familias de flags"](../familias-de-flags.md) valen aquí. En una tarea nueva no hay nada que sustituir
ni que quitar: `--replace-*` se acepta y deja la lista igual que `--add-*`, porque no hay nada previo
que sustituir, y `--rm-*` se acepta pero no tiene ningún elemento sobre el que actuar. `--clear-*` no
hace nada y avisa. Las que se usan de verdad al crear son
`--append-desc`, `--add-ac`, `--type`, `--priority`, `--add-labels`,
`--add-assignees`, `--add-refs`, `--add-deps`, `--parent`,
`--due`, `--ordinal`, `--above`, `--below`, `--author`, `--append-plan`, `--append-note`,
`--append-summary` y `--comment`.

- **El título llega por un sitio, y solo por uno.** Se escribe como argumento (`biso new "Fix the parser"`)
  o con `--title` (`biso new --title "Fix the parser"`), que es el mismo flag de campo que usa `biso set`
  (["Campos escalares"](../familias-de-flags.md#campos-escalares)). Darlo por los dos a la vez es un error
  aunque los dos textos sean iguales, porque quedarse con uno y descartar el otro en silencio sería
  una llamada que hace algo distinto de lo que se le pidió y sale con código 0. `--title` no tiene forma
  corta (["Una forma corta solo existe si nadie más reclama su inicial"](../../decisiones/comandos-y-flags.md#una-forma-corta-solo-existe-si-nadie-más-reclama-su-inicial)).
- **`--start`** crea la tarea directamente en el estado activo, asignada a `me` y con el arrendamiento
  tomado a favor de quien llama (`leaseExpiresAt` y `leaseHolder`, ["El arrendamiento de una tarea"](../lease.md)), exactamente como lo haría
  `biso start` sobre ella. Es el atajo de esas dos llamadas, así que la equivalencia tiene que ser real:
  si `--start` dejara la tarea activa y asignada sin arrendamiento, `biso new "X" --start` y
  `biso new "X"` seguido de `biso start` darían dos tareas distintas. Es, junto con `biso start`, la
  única vía que fija `leaseHolder` fuera de la importación.
- **Una tarea puede nacer colocada en el orden manual.** `--ordinal first`, `--ordinal last`,
  `--above <ref>` y `--below <ref>` valen aquí con el mismo significado que en `biso set`
  (["El orden manual"](../familias-de-flags.md#el-orden-manual)), así que
  `biso new "Fix the parser" --below MYP-11` crea la tarea justo debajo de `MYP-11`. Sin ninguno de
  los cuatro, la tarea nace sin clave, que es lo normal.
- **`--comment` funciona al crear**, igual que en cualquier otro comando de escritura.
- **`--append-plan`, `--append-note` y `--append-summary` no están restringidos por el estado.** Se
  pueden escribir al crear, en cualquier estado.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Título vacío, solo espacios o ausente, ya sea como argumento o con `--title` | Error 2: `error: title cannot be empty`, con el `code` `missing_title` |
| El título dado como argumento y con `--title` | Error 2, con el `code` `incompatible_flags`: `error: --title and the title argument cannot be used together` |
| El título dado solo con `--title` | Se crea la tarea, igual que con el argumento |
| Título muy largo | Se acepta entero, sin recortar |
| Título repetido | Se acepta sin aviso. Dos tareas pueden llamarse igual, para eso está el identificador |
| Valor fuera de un vocabulario cerrado | Error 3, con la lista de válidos |
| `--add-deps` a una tarea inexistente | Error 4. **Las dependencias se validan al escribirlas** |
| `--add-deps` a la propia tarea | Error 2. Un ciclo, en cambio, no puede darse al crear una sola tarea: no tiene identificador todavía y nada puede apuntar a ella (["Las familias de flags"](../familias-de-flags.md#campos-de-lista-que-admiten-coma)) |
| `--parent` inexistente | Error 4. Un ciclo de padres tampoco puede darse aquí, por el mismo motivo |
| `--due` con formato incorrecto | Error 2, señalando `YYYY-MM-DD` |
| `--due` en el pasado | Se acepta, con aviso |
| `--above` o `--below` a una tarea inexistente | Error 4, como cualquier otra referencia |
| `--above` o `--below` sobre una tarea que no tiene clave de orden | Error 6, con el `code` `neighbour_without_ordinal`, y la tarea no se crea (["El orden manual"](../familias-de-flags.md#el-orden-manual)) |
| `--append-desc @fichero` que no existe | Error 4 |
| `--start` sin ninguna identidad configurada (["Variables de entorno"](../invocacion.md#variables-de-entorno)) y sin `--add-assignees` | La tarea se crea en el estado activo y sin asignar, con `note: no identity configured, task left unassigned`, y **sin arrendamiento**: no hay ninguna identidad a la que atribuírselo, y una tarea sin asignar no puede tenerlo (["El arrendamiento de una tarea"](../lease.md)). Es el mismo caso que la fila equivalente de `biso start` (["`biso start`"](verbos-del-ciclo.md#biso-start)) |
| `--start` con `--add-assignees @sara` y una identidad configurada distinta | La tarea queda asignada a `@sara` y el arrendamiento es de quien llama, igual que en `biso start`: quien lo toma es quien escribe, no quien figura en `assignees` |
| Todo bien | Se crea la tarea, código 0 |

## Salida

Por defecto, **una línea por tarea creada, con el identificador y nada más**:

```
MYP-101
```

`biso new` es el único comando de escritura cuya salida por defecto es distinta de la línea de estado
de la sección ["`biso set`"](set.md), y así está dicho en el mensaje de arranque.

Con `--print`, la ficha completa en el formato de `biso get` **sustituye** esa línea del
identificador en vez de venir debajo de ella, que es la regla general del flag
(["Flags globales"](flags-globales.md)): el identificador es la primera cosa que imprime la ficha, así que
imprimir las dos cosas lo repetiría. Con `--quiet`, solo el identificador y ninguna nota.

**`biso new` no anuncia la clave de un `--add-ac` creado al mismo tiempo que la
tarea, a diferencia de la línea de estado de `biso set` (["`biso set`"](set.md#salida)).** Una tarea nace sin ningún
criterio, así que su contador de claves empieza siempre en 1: el primer `--add-ac` de la llamada es
la `#1`, el segundo la `#2`, y así en el mismo orden en que se escribieron los flags
(["Los criterios y sus claves estables"](../modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables)). Quien llama ya lo sabe sin preguntar, así que
imprimirlo sería el eco que el principio 4 prohíbe (["Los principios"](../principios.md)), y no la clase de dato que ese principio
manda enseñar.

### `--dry-run` sobre una sola tarea

Con `--dry-run`, `biso new` valida exactamente lo mismo que validaría la escritura real y no escribe
nada: ni la tarea, ni el identificador, que no llega a gastarse. Si todo está bien, no imprime ningún
identificador, porque no hay ninguno que enseñar, y sale con código 0 y esta línea por stderr, que es
la misma del lote en singular:

```
1 task would be created, nothing was written (--dry-run)
```

Con `--json`, el sobre es el de siempre y su `data.tasks` es una lista vacía, por el mismo motivo:
no hay ninguna tarea creada a la que nombrar. La línea de arriba se sigue imprimiendo por stderr,
porque no es una nota y no la lleva ningún sobre.

**Una vista previa emite los mismos avisos que emitiría la llamada real, con una excepción: los tres
de llegar a un estado terminal** (`terminal_ac_unchecked`, `terminal_no_summary` y
`open_question_on_terminal`, ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)). Los tres nombran la tarea, en su frase y en su
campo `task`, y aquí no hay ninguna que nombrar: el identificador no se ha gastado. Escribirlos con
el hueco vacío daría una frase con dos espacios seguidos y un `task` con la cadena vacía, que es justo
lo que prohíbe ["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias), y es el mismo motivo por el que los avisos
`imported_dod_merged`, `imported_documentation_merged` e `imported_modified_files_merged` del lote nombran la línea del fichero en vez de la tarea. Los demás avisos, que
no nombran ninguna tarea, salen igual con `--dry-run` que sin él: `due_in_past`, `clear_on_new_task`,
`duplicate_flag_value` y el resto. Crear la tarea de verdad en un estado terminal sí emite los tres,
porque entonces el identificador ya existe.

**Y si no está bien, el código nunca es el 7: es el código específico del fallo.** Una sola tarea no
puede producir más de un fallo a la vez, y cualquiera de los que puede producir es atribuible a un
elemento señalable, así que conserva su propio código igual que lo conservaría sin `--dry-run`: 2 para
un título vacío, 3 para un valor fuera de un vocabulario, 4 para un `--parent` que no
existe, 5 para una referencia de texto con varias coincidencias. Es la regla general de
["El código 7 garantiza que no se ha escrito nada, y el código específico siempre gana sobre
él"](../codigos-de-salida.md#el-código-7-garantiza-que-no-se-ha-escrito-nada-y-el-código-específico-siempre-gana-sobre-él),
aplicada aquí: el 7 queda para los fallos que no apuntan a un elemento concreto, y el único sitio de
`biso new` donde eso ocurre es el modo lote, donde `--from` puede traer muchas líneas malas por
motivos distintos. La fila del 7 de la tabla de códigos de salida de más abajo es, por tanto, una
fila del lote y solo del lote.

## El modo lote

```
biso new --from tareas.ndjson
biso new --from -
biso new --from tareas.ndjson --dry-run
```

La entrada es **NDJSON**: un objeto JSON por línea. Las líneas vacías y las que empiezan por `#` se
ignoran. Las claves son las del modelo de datos de la sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md), en `camelCase`.

Ejemplo de una línea, con todos los tipos compuestos:

```json
{"id":"MYP-101","title":"Normalize CRLF in the diff","type":"bug","priority":"high","status":"Done","description":"...","labels":["parser"],"references":["docs/bugs/BUG-02.md"],"dependencies":["MYP-90"],"acceptanceCriteria":[{"key":1,"text":"The diff ignores CRLF","checked":true},{"key":3,"text":"There is a test","checked":false}],"comments":[{"key":1,"author":"@avilches","createdAt":"2026-08-14T10:22:00Z","body":"Reported from Windows"}],"question":{"author":"@avilches","askedAt":"2026-08-16T09:00:00Z","body":"Is it a CRLF, or also a lone CR?"},"createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
```

Las reglas del lote, todas obligatorias:

- **`acceptanceCriteria` acepta dos formas.** Una cadena, que crea un elemento sin marcar con la
  siguiente clave libre, o un objeto con `key`, `text` y `checked`. Las dos formas se pueden mezclar
  dentro de la misma lista. Una `key` repetida dentro de la misma tarea es un fallo de validación.
- **El contador de claves se sitúa por encima de la clave mayor que tenga la tarea al acabar de
  importarla**, de modo que un criterio añadido después nunca choca con uno importado. El contador no
  es una clave del formato: se deduce.
- **`definitionOfDone` se acepta, se convierte en criterios de aceptación y avisa.** Es una de las tres
  claves ajenas al modelo que no son un fallo de validación (las otras dos, `documentation` y
  `modifiedFiles`, siguen a esta regla), y existe por una razón concreta: la definición de
  hecho estuvo en `biso` y sigue estando en Backlog.md, de donde viene la mayoría de los lotes de
  importación (["Se retira la definición de hecho"](../../decisiones/detalles.md#se-retira-la-definición-de-hecho)).
  Acepta las mismas dos formas que `acceptanceCriteria`, la cadena y el objeto. La conversión es un
  procedimiento en tres pasos, en este orden:

    1. Se importa `acceptanceCriteria` con sus propias reglas, y el contador de la tarea queda por
       encima de la clave mayor que haya entrado por ahí.
    2. Cada elemento de `definitionOfDone` se añade al final de la lista, en el orden en que venía,
       conservando su `text` y su `checked` y **tomando la siguiente clave libre del contador**. Su
       `key` original, si la trae, se descarta sin mirarla: las dos listas tenían contadores
       independientes, así que un elemento de cada una puede traer perfectamente la misma, y una
       `key` repetida dentro de `definitionOfDone` tampoco es un fallo de validación por el mismo
       motivo. Es la única diferencia con `acceptanceCriteria`, donde la `key` sí se respeta y
       repetirla sí es un fallo.
    3. Si se convirtió **al menos un** elemento, la tarea emite el aviso `imported_dod_merged`
       (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)). Una lista vacía no avisa de
       nada, porque no se ha convertido nada, y el lote no falla en ninguno de los dos casos.

    `definitionOfDone: null` equivale a que la clave no viniera, y no es un fallo de validación: la
    regla de que `null` en una lista es un fallo vale para las listas del modelo, y esta no lo es.
    Estas dos líneas importan la misma tarea:
  ```json
  {"title":"Normalize CRLF","acceptanceCriteria":[{"key":1,"text":"The diff ignores CRLF","checked":true}],"definitionOfDone":[{"key":1,"text":"Reviewed","checked":false}]}
  {"title":"Normalize CRLF","acceptanceCriteria":[{"key":1,"text":"The diff ignores CRLF","checked":true},{"key":2,"text":"Reviewed","checked":false}]}
  ```
  La primera avisa y la segunda no, y las dos dejan la misma tarea.
- **`documentation` se acepta, se funde en `references` y avisa.** Es la segunda de las tres claves
  ajenas al modelo que no son un fallo de validación, y existe por la misma razón que `definitionOfDone`: el campo
  estuvo en `biso` y sigue estando en Backlog.md, de donde viene la mayoría de los lotes de
  importación (["Se retira `documentation` y `references` queda como único campo de punteros"](../../decisiones/detalles.md#se-retira-documentation-y-references-queda-como-único-campo-de-punteros)).
  Es una lista de textos como `references`, con las mismas reglas que ella, y la fusión es un
  procedimiento en tres pasos, en este orden:

    1. Se importa `references` con sus propias reglas.
    2. Cada elemento de `documentation` se añade al final de esa lista, en el orden en que venía y
       conservando su texto. Un valor que la lista ya tenía, porque estaba en `references` o porque
       `documentation` lo repetía, no se añade otra vez. Es una regla propia de esta fusión, no la de
       los flags (["Repetición y listas separadas por comas"](../valores-de-entrada.md#repetición-y-listas-separadas-por-comas)):
       el lote no deduplica nada más, y un `references` que ya trae un valor repetido lo guarda
       repetido.
    3. Si la línea trajo **al menos un** elemento de `documentation`, la tarea emite el aviso
       `imported_documentation_merged` (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)),
       que cuenta los elementos que llegaron, aunque alguno se haya guardado una sola vez por
       repetido. Una lista vacía no avisa de nada, porque no se ha fundido nada, y el lote no falla en
       ninguno de los dos casos.

    `documentation: null` equivale a que la clave no viniera y no es un fallo de validación, por la
    misma razón que `definitionOfDone: null`: la regla de que `null` en una lista es un fallo vale
    para las listas del modelo, y esta no lo es. Estas dos líneas importan la misma tarea:
  ```json
  {"title":"Normalize CRLF","references":["docs/bugs/BUG-02.md"],"documentation":["docs/parser.md"]}
  {"title":"Normalize CRLF","references":["docs/bugs/BUG-02.md","docs/parser.md"]}
  ```
  La primera avisa y la segunda no, y las dos dejan la misma tarea.
- **`modifiedFiles` se acepta, se funde en `references` y avisa, con la misma regla que
  `documentation`.** Es la tercera clave ajena al modelo que no es un fallo de validación, y existe
  por la misma razón: el campo estuvo en `biso` y sigue estando en Backlog.md
  (["Se retira `modifiedFiles`"](../../decisiones/detalles.md#se-retira-modifiedfiles)). Sus elementos
  se funden en `references` con el mismo procedimiento de tres pasos, y el aviso se llama
  `imported_modified_files_merged` (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)).
  **Si una línea trae las dos claves, el orden es fijo y no depende del orden en que estén escritas en
  el fichero:** primero `references`, luego los elementos de `documentation` y por último los de
  `modifiedFiles`. Un valor que ya estaba se guarda una sola vez, donde apareció por primera vez, y
  la línea emite ambos avisos, el de `documentation` primero. Cada aviso cuenta los elementos que
  trajo su propia clave, aunque alguno se haya guardado una sola vez por repetido. `modifiedFiles: null`
  equivale a que la clave no viniera, por la misma razón que `documentation: null`. Estas dos
  líneas importan la misma tarea:
  ```json
  {"title":"Normalize CRLF","references":["docs/bugs/BUG-02.md"],"modifiedFiles":["internal/diff/diff.go"]}
  {"title":"Normalize CRLF","references":["docs/bugs/BUG-02.md","internal/diff/diff.go"]}
  ```
  La primera avisa y la segunda no, y las dos dejan la misma tarea.
- **Las claves que la conversión crea son lo único que `biso new` sí anuncia.** La salida de `new` no
  dice nunca qué clave recibió un criterio, porque quien llama puede deducirla (["Dónde se anuncia la
  clave de un criterio recién creado"](../../decisiones/detalles.md#dónde-se-anuncia-la-clave-de-un-criterio-recién-creado)),
  y en un lote con `definitionOfDone` eso deja de ser cierto: las claves salen de un contador que
  depende de lo que trajera `acceptanceCriteria` en esa misma línea. Por eso `acAdded` lleva, para esa
  tarea, las claves de los elementos convertidos, en el orden en que se crearon, y el aviso dice
  cuántos fueron.
- **`comments` es una lista de objetos** con `author`, `createdAt`, `body` y, opcionalmente, `key`.
  `createdAt` es opcional y, si falta, se pone el instante de la importación. `key` sigue la misma
  regla que la de `acceptanceCriteria`
  (["Los criterios y sus claves estables"](../modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables)): si falta, se asigna con la siguiente clave libre del
  contador de comentarios de esa tarea, y una `key` repetida dentro de los comentarios de la misma
  tarea es un fallo de validación. Es lo que hace cierta la simetría de `biso export` para la clave de
  un comentario (["`biso export`"](export.md)).
- **`question` se acepta como objeto** con `author`, `askedAt` y `body` (["La pregunta abierta"](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta)) en el lote de `--from`.
  `askedAt` es opcional y, si falta, se pone el instante de la importación, igual que `createdAt` en
  `comments`. Ausente la clave, la tarea se importa sin pregunta abierta.
- **`id`, `createdAt` y `updatedAt` se aceptan aquí y solo aquí.** Un `id` ya ocupado es un fallo de
  validación; un `id` libre se reserva y el tablero no lo volverá a asignar. **Estar ocupado tiene
  dos formas distintas y el mensaje las distingue**, porque en la segunda el tablero no tiene nada
  que ver: lo puede tener ya el tablero de destino, o lo puede haber tomado una línea anterior de
  este mismo fichero, y entonces el mensaje nombra esa línea y no el tablero. Las dos llevan el mismo
  `code` `id_taken` (["Los identificadores de error"](../contrato-json.md#los-identificadores-de-error)):
  ```
  line 7: id "MYP-11" is already taken on this board
  line 9: id "MYP-12" is already taken by line 3 of this file
  ```
- **`leaseExpiresAt` y `leaseHolder` se aceptan aquí con el valor que traiga el fichero**, que es lo
  que hace cierta la garantía de simetría de ["`biso export`"](export.md) para ellos dos. La invariante de ["El vaciado"](../lease.md#el-vaciado) de `lease.md` se
  comprueba en la validación, en sus dos mitades, y cada una es un fallo que nombra la línea y el campo.
  Una línea que traiga cualquiera de los dos sobre una tarea que no esté a la vez en el estado activo y
  asignada a alguien es un fallo de validación. Y una línea que traiga uno de los dos y no el otro
  también lo es, aunque la tarea esté activa y asignada: los campos van juntos, porque
  `leaseExpired` se calcula comparando `leaseExpiresAt` con el reloj de quien lee y con ese campo vacío
  no habría nada que comparar. Los dos fallos se ven así:
  ```
  line 14: leaseHolder on a task that is not both active and assigned
  line 31: leaseHolder given without leaseExpiresAt; the two go together
  ```
- **Un `id` explícito tiene que llevar el `task_prefix` del tablero de destino.** Si no lo lleva, es
  un fallo de validación, igual que un `id` ya ocupado: es la misma protección que hace inmutable a
  `task_prefix` en la sección ["`biso config`"](config.md), cerrando la tercera vía hacia el mismo tablero de identificadores
  mixtos que esa inmutabilidad ya evita en las otras dos (cambiar `--prefix` a mano, o renombrar el
  tablero). No es una restricción nueva sobre la simetría: exportar un tablero y restaurarlo con
  `biso snapshot` y `biso init --from` (["`biso export`"](export.md), ["`biso snapshot`"](snapshot.md)) trae también su `task_prefix`, así que los `id`
  de su `snapshot.ndjson` siempre lo llevan puesto.
- **`ordinal` es el único sitio donde una clave de orden llega escrita, y se valida.** Es una cadena
  con la forma que fija ["El orden manual y su clave"](../modelo-de-datos/orden-manual.md): símbolos
  de `0-9a-z` y sin `0` final. Una que no la cumpla es un fallo de validación con `code` propio,
  `malformed_ordinal`, y no el error de un número mal escrito, porque aquí no hay ningún número:
  ```
  line 14: malformed ordinal: "3000" (an ordinal key is made of 0-9 and a-z, and never ends in 0)
  ```
  **La cadena vacía tampoco cumple la forma**, así que es ese mismo fallo y no una forma de decir
  "sin clave": la que sí lo dice es `null`, o no escribir la clave, como en cualquier otro escalar
  opcional (más abajo en esta misma lista). Es la misma frontera que en la línea de órdenes, donde
  `--ordinal ""` es un error y `--clear-ordinal` es lo que quita la clave
  (["El valor vacío"](../valores-de-entrada.md#el-valor-vacío)).
  Un valor que no sea una cadena, por ejemplo el `3000` sin comillas de un tablero exportado por otra
  herramienta, es un valor del tipo equivocado y cae en `invalid_line`, como cualquier otro
  (["Los identificadores de error"](../contrato-json.md#los-identificadores-de-error)). **Dos líneas
  pueden traer la misma clave**, y no es un fallo: las claves no son únicas, y el listado desempata
  por identificador (["La regla de orden, completa"](ls.md#la-regla-de-orden-completa)). La clave se
  guarda tal cual llega, sin recalcular nada, que es lo que hace exacta la simetría con
  [`biso export`](export.md).
- **`archived` se acepta como booleano.** Por defecto, si la clave no aparece, la tarea se crea sin
  archivar. Ningún otro comando tiene un flag de campo para él: fuera de la importación,
  archivar se hace con `biso archive`.
- **`null` explícito en un escalar opcional (`due`, `ordinal`, `parent`) equivale a que la clave no
  viniera.** En una lista (`labels`, `references`, `dependencies`,
  `acceptanceCriteria`, `comments`), en cambio, `null` es
  un fallo de validación: su forma de estar vacío es `[]`, nunca `null`, la misma regla que
  ["El valor vacío"](../valores-de-entrada.md#el-valor-vacío) aplica a un escalar en la línea de
  órdenes. `null` en `question` equivale también a ausente, sin pregunta abierta.
- **`labels` se valida línea a línea con la regla de las etiquetas con ámbito.** Una etiqueta mal
  formada, una clave escrita con `::` que recibe más de un valor, una clave con los dos separadores
  en la misma línea, o un valor o un separador que la lista `labels` de la configuración no admite,
  son fallos de validación de esa línea
  (["Escribir una etiqueta con ámbito"](../familias-de-flags.md#escribir-una-etiqueta-con-ámbito) y
  ["La lista `labels`"](config.md#la-lista-labels)). **Aquí una clave con `::` repetida no se queda con
  el último valor**, a diferencia de lo que hace esa misma pareja escrita en dos `--add-labels`: en una
  línea de comandos los flags son una secuencia y el último es la intención más reciente, mientras que
  la lista `labels` de una línea describe el estado guardado de una tarea, y quedarse con uno de los
  valores en silencio perdería un dato que el fichero afirmaba. Los mensajes, con la forma de línea
  del informe del lote:
  ```
  line 7: malformed label: "size:"
  line 9: labels mix the two separators of the key "size": "size:a" and "size::b"
  line 12: labels give the key "size" more than one value, and :: allows at most one: "size::a" and "size::b"
  line 14: unknown label value: "size::xl" (valid: size::s, size::m, size::l)
  line 18: wrong separator for the label key "milestone": "milestone:m1"
  ```
  Sus `code`, dentro de `details`, son los mismos que fuera del lote: `malformed_label`,
  `mixed_label_separators`, `exclusive_label_conflict`, `unknown_label_value` y
  `wrong_label_separator` (["Los identificadores de error"](../contrato-json.md#los-identificadores-de-error)).
- **Una clave desconocida es un fallo de validación, no se ignora.** Ni la línea ni el lote se
  escriben, y el mensaje dice la línea y la clave. Las únicas excepciones son `definitionOfDone`,
  `documentation` y `modifiedFiles`, con cualquiera de sus valores admitidos, que se convierten con la regla de más
  arriba en vez de fallar.
- **Los campos derivados de la sección ["El modelo de datos de una tarea"](../modelo-de-datos/index.md) no se aceptan.** En la entrada son claves desconocidas y
  por tanto un fallo de validación.
- **Se valida el fichero entero antes de escribir nada**, y se aplica la garantía de todo o nada de
  la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables).
- Un lote no admite `--start` ni ningún flag de campo: todo va en el fichero.
- **Un lote no emite los avisos de llegar a un estado terminal**
  (`terminal_ac_unchecked`, `terminal_no_summary` y `open_question_on_terminal`). Todos hablan de
  llegar, y una tarea importada no llega a ninguna parte: ya estaba donde el fichero la pone, igual
  que una escritura sobre una tarea que ya estaba en el estado terminal tampoco los repite
  (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)). Los únicos avisos propios del lote son
  `imported_dod_merged`, `imported_documentation_merged` e `imported_modified_files_merged`.
- **`--print` es error 2 en un lote**, con el mensaje `--print does not apply to a batch, which
  affects no task that existed before`. Es la misma razón por la que lo es en
  [`biso init --from`](init.md): un lote crea de cero todas las tareas del fichero, así que no hay
  ninguna tarea anterior cuya ficha enseñar, e imprimir doscientas fichas recién creadas no informa
  de nada que no diga ya `biso ls`.
- **El fichero es UTF-8**, como cualquier otra entrada del programa: un byte que no lo sea es error 3
  con la clave `invalid_encoding`, señalando su posición
  (["Codificación y texto"](../salida-y-terminal.md#codificación-y-texto)).
- **El autor de una tarea importada es el que traiga la línea, y nunca la identidad de quien
  importa.** Una línea sin `author` deja la tarea sin autor, a diferencia de `biso new "X"`, que
  pone el de quien llama (["El autor"](../modelo-de-datos/autor.md)): el lote describe tareas que ya
  existían en otra parte, así que firmarlas con quien las trae sería inventarse un dato. Lo mismo
  vale para el autor de un comentario y para el de la pregunta abierta.
- **`parent` y `dependencies` son identificadores y nunca un texto que buscar.** La gramática de
  ["Cómo se resuelve una referencia a una tarea"](../referencias.md) no se aplica aquí: un fichero lo
  escribe un programa y no una persona, y resolver texto haría que la misma línea significara tareas
  distintas según lo que el tablero tuviera ese día. Cada uno se comprueba contra las tareas del
  tablero **y contra las del propio fichero**, así que una línea puede depender de otra que una línea
  posterior crea; un identificador que no está en ninguno de los dos sitios es un fallo de validación
  con el código 4, el mismo que `--add-deps` a una tarea que no existe. Los ciclos, de dependencia y
  de padres, se comprueban también sobre el resultado.
- **Una línea aporta un solo fallo al informe**, el primero que se encuentra al leerla en el orden de
  esta lista. Si además tiene claves desconocidas, se nombra la primera en orden alfabético y no la
  primera escrita, para que el mismo fichero dé siempre el mismo mensaje. Lo que el informe no hace
  nunca es pararse: las demás líneas se siguen leyendo, y el código 7 llega con todas.

Salida del lote, una línea por tarea, en el orden del fichero:

```
MYP-101
MYP-102
MYP-103
```

Salida de `--dry-run` cuando todo está bien, por stderr y con código 0:

```
242 tasks would be created, nothing was written (--dry-run)
```

Y cuando no, por stderr y con código 7, **con todos los fallos, no solo el primero**:

```
error: 5 of 242 lines are invalid, nothing was written
  line 12: id "OTHER-5" does not match this board's task prefix "MYP"
  line 47: unknown status: "Pendiente" (valid: To Do, In Progress, Done)
  line 88: unknown key: "trelloCard"
  line 130: parent names "MYP-900", which is not on this board and not in this file
  line 201: title cannot be empty
```

**El verbo concuerda con el número de fallos, no con el total.** Un solo fallo en un fichero de una
sola línea es:

```
error: 1 of 1 line is invalid, nothing was written
  line 1: unknown key: "nosuch"
```

**Los fallos salen siempre en orden ascendente de línea**, sea cual sea el momento en que se
descubren. No es una consecuencia gratuita de leer el fichero de arriba abajo: el `parent` y las
`dependencies` de una línea no se pueden juzgar hasta haber leído el fichero entero, porque pueden
nombrar una tarea que crea una línea posterior, así que esos dos fallos se encuentran en una segunda
pasada. La línea 130 del bloque de arriba es justo uno de ellos, y aun así sale entre la 88 y la 201.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea o lote creado | 0 |
| `--dry-run` que habría funcionado | 0 |
| Falta el título, flags incompatibles, fecha mal formada | 2 |
| Valor fuera de un vocabulario, entrada no interpretable | 3 |
| `--add-deps` o `--parent` a una tarea que no existe, o fichero de `@` que no existe | 4 |
| `--add-deps` o `--parent` por texto con varias coincidencias | 5 |
| La vecina de `--above` o de `--below` no tiene clave de orden | 6 |
| Cualquier fallo de validación en el lote de `--from`, o un `--dry-run` de ese lote que no pasa. **Solo del lote**: un `--dry-run` sobre una sola tarea nunca da 7, sino el código específico de su fallo (["`--dry-run` sobre una sola tarea"](#--dry-run-sobre-una-sola-tarea)) | 7 |
| El almacén falla, o no se obtiene el acceso exclusivo | 8 |
| No hay tablero | 20 |

## `biso new --help`

Además de los flags de uso más frecuente, la ayuda glosa `--add-deps`, `--parent` y `--add-refs`, y
dice que `--add-deps` nombra lo que va antes de la tarea nueva. No hay un flag para que la tarea
nueva bloquee a una existente: se crea y después se escribe la arista con `biso set` en la existente,
como dice la ayuda
(["La ayuda enseña la dirección de una dependencia"](../../decisiones/detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia)).

```
Usage: biso new <title> [options]
       biso new --title <text> [options]
       biso new --from <file|-> [options]

Create a task and print its id. Every field flag of `biso set` works here.

Arguments:
  title                      task title (required unless --title or --from is given)

Most used:
      --title <text>          the title, instead of the argument; never both
      --append-desc <text>    description; repeat to append paragraphs
      --add-ac <text>         add an acceptance criterion; repeatable
      --type <value>          configured type
      --priority <value>      configured priority
      --status <value>        configured status (default: the initial one)
      --add-labels <value>    add a label; repeatable or comma-separated
      --add-assignees <@who>  add an assignee; repeatable or comma-separated
      --add-deps <ref>        tasks that must be done first, so each blocks the
                              new task; repeatable, checked to exist
      --parent <ref>          the task this one is part of; at most one
      --add-refs <text>       a path, URL or task id to look at; repeatable
      --due <YYYY-MM-DD>      due date
      --comment <text>        add a discussion comment; repeatable
      --append-plan <text>    implementation plan
      --start                 create it already in the active status, assigned
                              to you, with the lease claimed for you

Every other field flag of `biso set --help` is accepted too.

Dependencies are written on the task that waits: `--add-deps MYP-4` means
MYP-4 goes first and blocks it. There is no flag for the opposite, a new task
that blocks MYP-10; create it, then run `biso set MYP-10 --add-deps <new id>`.

Batch:
      --from <file|->        NDJSON, one task object per line. The only place
                             where id, createdAt, updatedAt, criterion keys,
                             comment timestamps and question timestamps can be
                             given. Validated whole before anything is written.

Any text option also takes @file to read a file, or - to read stdin.

Exit codes:
  0  created            5  a text reference matched several tasks
  2  bad usage          6  --above or --below on a task with no place
  3  unknown value      8  the board could not be written
  4  a referenced task or file does not exist
  7  batch or --dry-run validation failed, nothing was written
                        20 no board here

Examples:
  biso new "Normalize CRLF in the diff" --type bug --priority high
  biso new "Add OAuth" --add-ac "Login succeeds" --add-ac "Token refreshes"
  biso new "Parse the header" --parent MYP-10 --add-deps MYP-4
  biso new "Rewrite the installer" --append-desc @docs/installer.md --start
  biso new --from tasks.ndjson --dry-run
```

---

