# `biso config`

## Firma

```
biso config get <key>
biso config set <key> <value>
biso config list
```

## Parámetros

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `<key>` | sí en `get` y `set` | una clave de la tabla de abajo | | no | no | |
| `<value>` | sí en `set` | según la clave | | no | sí en las claves de lista | |

`--json` solo se acepta en `config list`. En `get` la salida ya es un solo valor y en `set` no hay
salida por stdout, así que en los dos es un error de uso con código 2. El rechazo es texto plano por
stderr, no el sobre de error de ["El contrato JSON"](../contrato-json.md#los-errores-en-json), porque
es `--json` mismo la parte inválida de la llamada:

```
error: --json only applies to config list
```

## Las claves

| Clave | Tipo | Por defecto |
|---|---|---|
| `project_name` | texto | el nombre del proyecto |
| `statuses` | lista, mínimo tres | `To Do, In Progress, Done` |
| `initial_status` | uno de `statuses` | `To Do`, al crear el tablero sin `--statuses` |
| `active_status` | uno de `statuses` | `In Progress`, al crear el tablero sin `--statuses` |
| `terminal_status` | uno de `statuses` | `Done`, al crear el tablero sin `--statuses` |
| `types` | lista | `task, bug, docs` |
| `priorities` | lista | `high, medium, low` |
| `labels` | lista | vacía |
| `assignees` | lista | vacía |
| `task_prefix` | texto de solo letras | se deriva de `project_name` en mayúsculas (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)) |
| `finish_strict` | booleano | falso |
| `lease_minutes` | entero > 0 | 240 |
| `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`, `urgency.criteria`, `urgency.age` | decimal | ver ["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia) para el término de cada uno y su valor por defecto |

**`project_name` es el nombre del tablero, y cambiarlo no toca el sistema de ficheros.** Ninguna clave
de esta tabla lo hace. La carpeta del tablero se queda con el nombre que tenga, aunque sea el slug de un
nombre anterior, porque ese nombre es decorativo y nadie resuelve por él (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)): la identidad del
tablero está en el marcador `<id>.id` y dentro de la base de datos, no en el nombre de la carpeta. Así
que renombrar un tablero es una escritura en su base de datos y nada más, con la misma transacción y las
mismas garantías que cualquier otra.

**Eso quita de en medio la única operación que no podía ser atómica.** Mover un directorio no cabe dentro
de una transacción de SQLite, así que renombrar el tablero habría sido escribir la configuración y
después mover la carpeta, con un estado intermedio observable si la segunda mitad fallaba, un error 8
propio para el fallo de permisos, y la posibilidad de dejar sin resolver el puntero de un tablero que
viviera fuera de las raíces de la sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina). Nada de eso existe: no hay dos mitades.

**Ningún carácter de `project_name` puede romper nada, y ahora por un motivo más simple**: no entra en
ninguna ruta. Sigue habiendo un slug derivado de él (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), pero solo se usa para dar nombre a la
carpeta cuando `biso init` la crea, y ahí el valor ya está comprobado. Un `project_name` cuyo slug quede
vacío sigue siendo un error, porque el slug es un dato del tablero y la regla de que un valor no válido
nunca se acepta vale igual.

**Renombrar no toca nunca el `task_prefix`.** Se derivó una vez al crear el tablero y desde entonces
vive por su cuenta en esa clave. Cambiar `project_name` no lo recalcula, aunque el nombre nuevo diera un
`task_prefix` distinto si el tablero se creara hoy. Y si el nombre nuevo no deja ninguna letra con la
que derivar uno (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)), tampoco es un error aquí, porque `task_prefix` ya está fijado y no se
recalcula al renombrar.

Lo mismo vale para `biso init --overwrite-config` (sección ["`biso init`"](init.md)), y solo cuando se da `<name>`
explícito: si ese `<name>` difiere del `project_name` que el tablero ya tenía, lo cambia igual que lo
haría `config set project_name`, y tampoco mueve nada. **Sin `<name>` explícito, `project_name` se
conserva tal cual estaba**, aunque el valor por defecto de `<name>` sea el nombre del directorio del
proyecto: ese valor por defecto tiene sentido como punto de partida al crear un tablero nuevo, no como
instrucción de renombrar uno que ya existe.

**`lease_minutes` fija cuánto dura el arrendamiento de una tarea activa y asignada (["El arrendamiento de una tarea"](../lease.md)), y se
puede cambiar libremente en cualquier momento, sin caer nunca en el error 6.** A diferencia de
`task_prefix` o de `statuses` en uso, esta clave no queda incrustada en ninguna tarea existente:
`leaseExpiresAt` se calcula al escribir, así que cambiar `lease_minutes` solo afecta a los
arrendamientos que se renueven desde ese momento, nunca a los ya fijados, y el tablero nunca queda
inconsistente por ello. El valor por defecto, 240 minutos, viene de que **el error dañino es el falso
vencido, no el vencido tardío**: un arrendamiento demasiado corto hace que un agente reclame una tarea
que otro está trabajando de verdad, mientras que uno demasiado largo solo retrasa el aviso. Y las
consecuencias de un valor mal calibrado son más pequeñas de lo que parecen, porque `biso start` avisa
y coge la tarea igual incluso con el arrendamiento vivo (["`biso start`"](verbos-del-ciclo.md#biso-start)): una duración mal puesta produce un
informe equivocado, no datos equivocados.

## La lista `labels`

**Vacía, que es como nace, no restringe nada.** Cualquier etiqueta bien formada se puede escribir, y
los filtros validan contra el conjunto de etiquetas en uso
(["Qué valida cada filtro, y contra qué"](../vocabularios.md#qué-valida-cada-filtro-y-contra-qué)). Con
entradas, cada una toma una de estas formas, y **la lista solo restringe las claves que nombra**:

| Entrada | Qué es | Qué restringe |
|---|---|---|
| `pepe` | una etiqueta plana | nada. Se ofrece, por ejemplo a una interfaz que enseñe las etiquetas disponibles al crear una tarea, y no prohíbe ninguna otra |
| `size::m` | un par exacto | la clave `size` admite solo los valores declarados así, y solo con el separador declarado |
| `milestone::` | una clave abierta | la clave `milestone` admite cualquier valor, pero solo con el separador declarado |

Una etiqueta cuya clave ninguna entrada nombra sigue siendo libre, y una etiqueta plana que nadie
declara también: declarar `milestone::` no obliga a declarar nada más.

**Dentro de esta lista, y solo aquí, la forma sin valor no es una etiqueta mal formada**
(["Las etiquetas con ámbito"](../valores-de-entrada.md#las-etiquetas-con-ámbito)): es la declaración de
una clave abierta, igual que en un filtro es la consulta por una clave. `milestone::` declara la
clave con el separador que admite como mucho un valor por tarea, y `trello:` la declara con el que
admite varios.

**La lista tiene que ser coherente consigo misma**, y se comprueba al escribirla. Una misma clave no
puede aparecer con los dos separadores, ni como clave abierta y con valores exactos a la vez, porque
en los dos casos la lista diría dos cosas distintas sobre la misma clave:

```
error: labels: the key "size" is declared with both : and ::
hint: a key takes one separator, : for several values or :: for at most one

error: labels: the key "milestone" is declared both as an open key and with exact values
hint: declare milestone:: on its own, or its values, but not both
```

Las dos salen con código 3 y el `code` `bad_config_value`, como cualquier otro valor de configuración
fuera de dominio.

**Escribir en una tarea un valor que una clave restringida no declara, o esa clave con el otro
separador, es error 3**, porque es una etiqueta bien formada que el tablero no reconoce, que es
exactamente lo que ese código significa (["Códigos de salida"](../codigos-de-salida.md#el-código-3-cubre-dos-direcciones)):

```
error: unknown label value: "size::xl"
       valid labels for the key "size" on this board: size::s, size::m, size::l

error: wrong separator for the label key "milestone": "milestone:m1"
hint: this board declares milestone::, at most one value per task
```

Sus `code` son `unknown_label_value` y `wrong_label_separator`
(["Los identificadores de error"](../contrato-json.md#los-identificadores-de-error)). El primero lleva
en `valid` las etiquetas que esa clave admite, escritas enteras y con su separador, que es lo mismo
que lista su segunda línea; el segundo no lleva `valid`, porque la clave abierta no tiene ningún
conjunto de valores que ofrecer.

**Esas comprobaciones son de escritura y no de lectura.** Al leer, el separador no distingue nada
(["Las etiquetas con ámbito"](../valores-de-entrada.md#las-etiquetas-con-ámbito)), así que
`biso ls --label milestone:m1` encuentra igual las tareas etiquetadas `milestone::m1` aunque la lista
declare la clave con `::`. Un valor que no existe en el tablero ya lo rechaza el filtro con su propio
`unknown_label`, sin que esta lista tenga que decir nada.

**`biso config set labels` falla si alguna tarea lleva una etiqueta que la lista nueva prohibiría**,
mirando todas las tareas del tablero, archivadas y terminadas incluidas, que es el mismo conjunto
contra el que validan los filtros. Es la misma regla que impide quitar de `statuses` un estado en
uso: la configuración no puede dejar guardado un valor que ella misma declara imposible. El informe da una línea por etiqueta ofendida, con
las tareas que la llevan en orden ascendente de identificador, y sale con código 6 y el `code`
`board_inconsistent`:

```
error: labels: 3 tasks use a label the new list would not allow
  "milestone:m1" on MYP-11: the new list declares the key "milestone" with ::
  "size::xl" on MYP-19, MYP-23: the key "size" allows size::s, size::m, size::l
```

Con una sola tarea la primera línea es `error: labels: 1 task uses a label the new list would not
allow`. **Quitar de la lista una etiqueta plana no falla nunca**, porque no restringía nada, y
tampoco falla quitar una clave entera: lo que deja de estar declarado vuelve a ser libre.

**El mensaje de arranque no lista esta clave**, y eso está dicho en
["`biso prime`"](prime.md#la-salida-literal) con su razón.

**Los tres estados especiales son valores explícitos, no posiciones.** Se escriben al crear el tablero
y **cambiar `statuses` no los mueve nunca**. Esta es la diferencia que evita que añadir un estado al
final cambie en silencio a dónde va `biso finish`.

**`task_prefix` es inmutable en cuanto el tablero tiene alguna tarea.** Es la única clave cuyo valor
queda incrustado en datos que ya existen, porque cada identificador ya asignado lleva el prefijo
grabado. Mientras el tablero está vacío no hay ningún identificador con el que pueda entrar en
conflicto, así que cambiarla funciona sin más; en cuanto existe una sola tarea, cambiarla es error 6
(tabla de abajo), con el mismo motivo por el que no se toca `statuses` en uso. La misma regla vale
para `biso init --overwrite-config` (sección ["`biso init`"](init.md)): reescribir la configuración de un tablero con
tareas nunca puede cambiar el `task_prefix` que ya tenía, se pase `--prefix` explícito o no.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Clave inexistente | Error 4, con hasta tres de las claves más parecidas (las que pasen el umbral de ["El algoritmo de sugerencias más parecidas"](../vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas)) |
| Valor del tipo equivocado, por ejemplo `finish_strict maybe` | Error 3, diciendo qué tipo esperaba |
| `initial_status` a un valor que no está en `statuses` | Error 3 |
| `lease_minutes` a cero o negativo | Error 3, el mismo trato que cualquier valor fuera de dominio de esta tabla |
| Quitar de `statuses` un estado que alguna tarea usa | Error 6, con cuántas tareas lo usan y en cuáles |
| Quitar de `statuses` un estado que es `initial_status`, `active_status` o `terminal_status` | Error 6, diciendo cuál de los tres y que hay que cambiarlo antes |
| Dejar `statuses` con menos de tres elementos | Error 6, diciendo cuántos hacen falta |
| Dar a un papel (`initial_status`, `active_status` o `terminal_status`) el mismo estado que otro papel ya tiene | Error 6, con los dos papeles y el estado que comparten |
| Quitar de `types` o `priorities` un valor en uso | Error 6, igual |
| Una entrada de `labels` que no es ni una etiqueta bien formada, ni un par exacto, ni una clave abierta | Error 2 con el `code` `malformed_label`, el mismo que en cualquier otro sitio donde se escribe una etiqueta (["Las etiquetas con ámbito"](../valores-de-entrada.md#las-etiquetas-con-ámbito)) |
| La misma clave en `labels` con los dos separadores, o abierta y con valores exactos a la vez | Error 3 (["La lista `labels`"](#la-lista-labels)) |
| `set labels` a una lista que prohibiría una etiqueta que alguna tarea lleva | Error 6, con una línea por etiqueta y las tareas que la llevan (["La lista `labels`"](#la-lista-labels)) |
| Quitar de `labels` una etiqueta plana, o una clave entera | No es un error: lo que deja de estar declarado vuelve a ser libre |
| Vaciar `types` o `priorities` del todo (`biso config set types ""`), sin ningún valor en uso | No es un error: ninguna de las dos tiene mínimo en la tabla de claves, al contrario que `statuses`, y `labels` y `assignees` ya nacen vacías. El tablero queda sin ese vocabulario, así que deja de admitir `--type` (o `--priority`) en cualquier llamada, y ["`biso prime`"](prime.md#la-salida-literal) lo dice en su línea `types` con un `(none)` |
| Cambiar `task_prefix` cuando el tablero ya tiene alguna tarea | Error 6, remitiendo a exportar el tablero, reescribir los identificadores e importarlos en un tablero nuevo |
| Cambiar `project_name` a un valor vacío, o a uno cuyo slug (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)) quede vacío tras derivarlo | Error 3, en los dos casos |
| Cambiar `project_name` al mismo valor que ya tiene | El `set` se completa igual, con su `note:` |
| Cambiar `project_name` a un valor que no dejaría ninguna letra para derivar un prefijo | No es error: el `task_prefix` ya está fijado y no se recalcula al renombrar |
| Cambiar `project_name` en un tablero cuyo puntero lleva `path`, o cuya carpeta ya no se llama como el nombre viejo | No es un caso especial: no se mueve nada y el puntero sigue valiendo, porque nada resuelve por el nombre de la carpeta (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)) |
| `get` de una clave de lista | Los valores separados por comas, en una línea |
| `set` correcto | Sin salida por stdout, con `note:` por stderr diciendo el valor nuevo |

La nota de un `set` correcto lleva la clave y el valor nuevo, con el mismo formato `clave = valor`
que usan `get` y `list`:

```
$ biso config set lease_minutes 45
note: lease_minutes = 45
```

En una clave de lista sale la lista entera ya separada por comas, como la vería un `get` posterior,
nunca solo el elemento que se acaba de añadir o quitar:

```
$ biso config set statuses "To Do,In Progress,Done,Blocked"
note: statuses = To Do,In Progress,Done,Blocked
```

**Vaciar una lista deja la nota sin nada detrás del igual**, exactamente igual que la línea que esa
clave tiene en `config list` y que la respuesta de un `get` posterior, sin espacio suelto al final:

```
$ biso config set types ""
note: types =
```

Ningún cambio de configuración toca ninguna tarea, nunca.

**Una clave inexistente sugiere hasta tres de las más parecidas**, con el algoritmo de
["El algoritmo de sugerencias más parecidas"](../vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas):

```
$ biso config get urgency.pryority
error: no such key: "urgency.pryority"
hint: did you mean: urgency.priority, urgency.criteria, urgency.active?
```

**Con `--dry-run`, la misma nota pero en condicional**, porque `config set` no tiene una línea de
estado que marcar como hipotética (sección ["`biso set`"](set.md#salida)): no hay ninguna tarea
afectada, solo un valor que se habría escrito.

```
$ biso config set lease_minutes 45 --dry-run
note: lease_minutes would be set to 45 (--dry-run)
```

Y con una clave de lista, la lista entera que quedaría, con la misma forma que la nota real:

```
$ biso config set statuses "To Do,In Progress,Done,Blocked" --dry-run
note: statuses would be set to To Do,In Progress,Done,Blocked (--dry-run)
```

**La vista previa de vaciar una lista tiene su propia redacción**, porque "would be set to" sin nada
detrás deja la frase en el aire. Lo que hace un vocabulario que se vacía a propósito es quedarse
vacío, y eso es lo que dice:

```
$ biso config set types "" --dry-run
note: types would be emptied (--dry-run)
```

## Salida

```
$ biso config get statuses
To Do,In Progress,Done

$ biso config list
project_name = My project
statuses = To Do,In Progress,Done
initial_status = To Do
active_status = In Progress
terminal_status = Done
types = idea,memory,task,bug,docs
priorities = high,medium,low
labels =
assignees =
task_prefix = TASK
finish_strict = false
lease_minutes = 240
urgency.priority = 6.0
urgency.active = 4.0
urgency.blocking = 8.0
urgency.blocked = -5.0
urgency.due = 12.0
urgency.criteria = 1.0
urgency.age = 0.5
```

**`config list` imprime las diecinueve claves, siempre, en el orden de la tabla de claves de arriba**, y
los siete coeficientes de la urgencia con el nombre con el que `config set` los acepta, uno por línea.
Lo que `list` enseña es exactamente el conjunto de claves que `set` admite, y por eso no puede haber
ninguna que solo se vea con `--json`: una clave escondida es una clave que nadie sabe que puede cambiar.

**Una lista vacía se imprime como la clave, el igual y nada detrás**, que es lo mismo que hace
`config get` de una lista vacía, así que las dos listas que nacen vacías (`labels` y `assignees`)
aparecen igual en un tablero recién creado.

## El esquema JSON

Solo `config list` acepta `--json`:

```json
{
  "schemaVersion": 1,
  "kind": "config",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "config": {
      "project_name": "My project",
      "statuses": ["To Do", "In Progress", "Done"],
      "initial_status": "To Do",
      "active_status": "In Progress",
      "terminal_status": "Done",
      "types": ["idea", "memory", "task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "labels": [],
      "assignees": [],
      "task_prefix": "TASK",
      "finish_strict": false,
      "lease_minutes": 240,
      "urgency": { "priority": 6.0, "active": 4.0, "blocking": 8.0, "blocked": -5.0,
                   "due": 12.0, "criteria": 1.0, "age": 0.5 }
    }
  }
}
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Hecho | 0 |
| Sintaxis, o `--json` fuera de `list` | 2 |
| Valor de tipo o de dominio incorrecto | 3 |
| Clave inexistente | 4 |
| El cambio dejaría el tablero inconsistente | 6 |
| No se puede escribir la configuración | 8 |
| No hay tablero | 20 |

## `biso config --help`

```
Usage: biso config get <key>
       biso config set <key> <value>
       biso config list [--json]

Read and change the board configuration. List values are comma-separated.
No configuration change ever touches a task.

Keys:
  project_name       board name; changing it never moves anything on disk,
                     the board folder keeps whatever name it has
  statuses           the board statuses, in order
  initial_status     status of a new task           (one of statuses)
  active_status      what `biso start` sets         (one of statuses)
  terminal_status    what `biso finish` sets        (one of statuses)
  types              configured task types
  priorities         configured priorities
  labels             labels that filters accept on top of the ones in use;
                     a key::value or key:: entry also restricts what that key
                     accepts, a plain one only offers itself
  assignees          assignees that filters accept on top of the ones in use
  task_prefix        id prefix, letters only (default: derived from
                     project_name); immutable once the board has a task
  finish_strict      make `biso finish` refuse an incomplete task
  lease_minutes      lease duration in minutes (default 240); free to change
                     at any time, it only affects future renewals
  urgency.priority, urgency.active, urgency.blocking, urgency.blocked,
  urgency.due, urgency.criteria, urgency.age
                     the seven urgency coefficients; see `biso get --explain-urgency`

The three special statuses are stored as explicit values. Changing `statuses`
never moves them; if a change would remove one of them, it fails and says so.

Removing any configured value that a task still uses is refused, never applied
silently.

Options:
      --json         machine-readable output, `list` only
  -h, --help         show this help

Exit codes:
  0  done            4  no such key
  2  bad usage       6  the change would leave the board inconsistent
  3  bad value       8  the configuration could not be written
                     20 no board here

Examples:
  biso config get active_status
  biso config set statuses "To Do,In Progress,Done"
  biso config set finish_strict true
  biso config list --json
```

---

