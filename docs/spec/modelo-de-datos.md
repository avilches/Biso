# El modelo de datos de una tarea

Este es el modelo **lógico**. Describe qué campos tiene una tarea, de qué tipo son y quién los
escribe. No dice nada de cómo se guardan.

**Todo campo tiene un tipo concreto, y la tabla de abajo lo dice para cada uno, sin excepción.** Los
tipos son `string` (texto de una línea), `text` (bloque de prosa, con saltos de línea), `enum(...)`
(un vocabulario cerrado, entre paréntesis; dice "configurable" cuando el tablero puede ampliarlo),
`date` (una fecha o un instante; el formato exacto se dice en la fila del campo), `int`, `float`,
`bool`, `list<string>` (varios valores simples separados por coma al escribir, direccionables por su
propio valor, nunca por posición), y `map<string,string>` (pares clave-valor con las claves
declaradas de antemano, direccionables por su clave). Los campos con estructura propia usan tres
tipos con nombre, cada uno definido en su propia sección: `list<Criterion>`, `list<Comment>` y
`Question` (esta última con un único valor, nunca una lista).

**¿Tiene fecha y autor propios, y se puede señalar uno a uno?** Solo se aplica a los tipos con
nombre, y cada uno lo contesta en su propia sección: ["Los criterios y sus claves
estables"](#los-criterios-y-sus-claves-estables) para `Criterion`, ["Los comentarios"](#los-comentarios)
para `Comment`, ["La pregunta abierta"](#la-pregunta-abierta) para `Question`. Ningún otro tipo tiene
fecha ni autor propios.

| Campo | Tipo | Obligatorio | Quién lo fija | Mutable |
|---|---|---|---|---|
| `id` | `string` (`PREFIX-<n>`) | sí | el programa | no |
| `title` | `string` | sí | quien llama | sí |
| `status` | `enum(...)`, configurable (ver `statuses` en [`biso config`](cmd/config.md)) | sí | quien llama | sí |
| `type` | `enum(...)`, configurable (ver `types` en [`biso config`](cmd/config.md)) | no | quien llama | sí |
| `priority` | `enum(...)`, configurable (ver `priorities` en [`biso config`](cmd/config.md)) | no | quien llama | sí |
| `parent` | `string` (referencia a otra tarea) | no | quien llama | sí |
| `assignees` | `list<string>` (textos de persona) | no | quien llama | sí |
| `reporter` | `string` (texto de persona) | no | el programa al crear, o quien llama; ver ["Quién reporta una tarea"](#quién-reporta-una-tarea) | sí |
| `labels` | `list<string>` | no | quien llama | sí |
| `dependencies` | `list<string>` (referencias a tareas) | no | quien llama | sí |
| `references` | `list<string>` | no | quien llama | sí |
| `documentation` | `list<string>` | no | quien llama | sí |
| `modifiedFiles` | `list<string>` | no | quien llama | sí |
| `due` | `date` (`YYYY-MM-DD`) | no | quien llama | sí |
| `ordinal` | `int` (>= 0) | no | quien llama | sí |
| `createdAt` | `date` (instante UTC, precisión de segundo) | sí | el programa | solo al importar |
| `updatedAt` | `date` (instante UTC, precisión de segundo) | sí | el programa | solo al importar |
| `archived` | `bool` | sí, `false` por defecto | el programa, con `biso archive` | sí, solo con `biso archive` / `--unarchive`, o al importar |
| `leaseExpiresAt` | `date` (instante UTC) | no | el programa, a `ahora + lease_minutes` (clave de [configuración](cmd/config.md)); ver [La renovación](lease.md#la-renovación) en `lease.md` para cuándo | sí, ver [`lease.md`](lease.md), o al importar |
| `leaseHolder` | `string` (texto de persona) | no | el programa, solo con [`biso start`](cmd/verbos-del-ciclo.md#biso-start) y con [`biso new --start`](cmd/new.md); ver [La renovación](lease.md#la-renovación) en `lease.md` | sí, solo con esos dos, o al importar; ver [`lease.md`](lease.md) |
| `urgency` | `float`, derivado | derivado | el programa | no, se recalcula al leer |
| `ext` | `map<string,string>` | no | quien llama | sí |
| `description` | `text` | no | quien llama | sí |
| `plan` | `text` | no | quien llama | sí |
| `notes` | `text` | no | quien llama | sí |
| `summary` | `text` | no | quien llama | sí |
| `acceptanceCriteria` | `list<Criterion>`; ver ["Los criterios y sus claves estables"](#los-criterios-y-sus-claves-estables) | no | quien llama | sí |
| `definitionOfDone` | `list<Criterion>`; ver ["Los criterios y sus claves estables"](#los-criterios-y-sus-claves-estables) | no | quien llama | sí |
| `comments` | `list<Comment>`; ver ["Los comentarios"](#los-comentarios) | no | quien llama | se añade, se borra entero, o se corrige solo la fecha; nunca se edita el cuerpo ni el autor |
| `question` | `Question`; ver ["La pregunta abierta"](#la-pregunta-abierta) | no | mixto, según la parte; ver ["La pregunta abierta"](#la-pregunta-abierta) | sí, solo con `biso ask`, `biso answer`, o al importar |
| `acDone`, `acTotal`, `dodDone`, `dodTotal` | `int`, derivado | derivado | el programa | no, se recalculan al leer |
| `commentCount` | `int`, derivado | derivado | el programa | no, se recalcula al leer |
| `blocks` | `list<string>`, derivado | derivado | el programa | no, se recalcula al leer |
| `blocked`, `waiting` | `bool`, derivado | derivado | el programa | no, se recalculan al leer |
| `leaseExpired` | `bool`, derivado | derivado | el programa | no, se recalcula al leer |

Precisiones sobre la mutabilidad:

- **"No mutable" significa que ninguna bandera del programa lo cambia.** `updatedAt` lo reescribe el
  programa en cada operación que cambie algo.
- **El cuerpo y el autor de un comentario no se editan nunca, por ninguna vía.** Un comentario es el
  registro de una conversación, y lo que se dijo no se reescribe. Lo que sí se puede corregir, con las
  banderas dedicadas de ["Comentarios"](familias-de-banderas.md#comentarios) y nunca con una escritura general sobre la
  tarea, es borrar el comentario entero (`--rm-comment`) o corregir únicamente su fecha
  (`--set-comment-date`). La razón, con el caso que la motiva, está en
  ["Borrar o corregir la fecha de un comentario"](../DECISIONES.md#borrar-o-corregir-la-fecha-de-un-comentario) de `DECISIONES.md`.
- **`archived` solo lo cambia `biso archive` y `biso archive --unarchive`.** No hay una bandera de
  campo de la sección ["Las familias de banderas"](familias-de-banderas.md) para él: archivar es un gesto de flujo de trabajo con nombre propio,
  según el principio 5.
- **Los campos marcados "derivado" en esta tabla no se guardan.** Se calculan al leer, y son
  exactamente los campos que [`biso export`](cmd/export.md) no escribe y que [`biso new --from`](cmd/new.md) rechaza como
  clave desconocida: `urgency`, `acDone`, `acTotal`, `dodDone`, `dodTotal`, `commentCount`,
  `blocks`, `blocked`, `waiting` y `leaseExpired`. Esta es la única lista de campos derivados del
  documento; las demás secciones remiten a ella.

**Las reglas de `leaseExpiresAt` y `leaseHolder` (cuándo cuenta como vencido el arrendamiento, cuándo
se renueva, cuándo se vacía y qué hace la importación con ellos) tienen su propio documento,
[`lease.md`](lease.md), porque dominan la mutabilidad de una tarea mucho más que cualquier otro campo
y no son el tema principal de este.**

## Los criterios y sus claves estables

`acceptanceCriteria` y `definitionOfDone` son listas del mismo tipo, `Criterion`, y cada elemento
tiene tres partes, sin fecha ni autor propios:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | `int` positivo | el programa al crear el elemento |
| `text` | `string` | quien llama |
| `checked` | `bool` | quien llama |

**La clave se asigna al crear el elemento, con un contador propio de esa lista dentro de esa tarea, y
no se reasigna nunca.** Quitar un elemento no mueve las claves de los demás: una tarea puede tener
perfectamente los criterios `#1` y `#3` y ninguno más. Cada tarea lleva dos contadores, uno por
lista, que solo crecen. Un `Criterion` se direcciona siempre por su `key`, nunca por su posición en
la lista.

Consecuencias que hay que respetar en toda la implementación:

- Los selectores de la sección ["Selectores de criterios"](familias-de-banderas.md#selectores-de-criterios) trabajan sobre la clave, **nunca** sobre la posición.
- `acTotal` y `dodTotal`, allá donde aparezcan, son **el número de elementos presentes**, nunca la
  clave más alta. Una tarea con los criterios `#1` y `#3` tiene `acTotal` igual a 2.
- Los elementos se muestran y se exportan en el orden en que están en la lista, que es el orden en
  que se crearon salvo que se haya sustituido la lista entera.

## Los comentarios

Cada elemento de `comments` es del tipo `Comment`, con clave, autor, instante y cuerpo propios:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | `int` positivo | el programa al crear el comentario |
| `author` | `string` libre | quien llama, y por defecto la identidad `me` |
| `createdAt` | `date` (instante UTC) | el programa, salvo al importar o con `--set-comment-date` (["Comentarios"](familias-de-banderas.md#comentarios)) |
| `body` | `text` | quien llama |

**La clave se asigna al crear el comentario, con un contador propio de esa lista dentro de esa
tarea, y no se reasigna nunca**, exactamente igual que la de un `Criterion`
(["Los criterios y sus claves estables"](#los-criterios-y-sus-claves-estables)). Borrar un comentario no mueve las claves de los
demás, y un `Comment` se direcciona siempre por su `key`, nunca por su posición.

**El autor es texto libre y no se valida contra nada.** Un comentario puede venir de alguien que no
existe en este tablero, y un sistema externo puede usar su propia convención, por ejemplo
`@trello:juan`.

**Los comentarios se guardan y se muestran en orden de inserción, no en orden de `createdAt`.** El
instante de cada uno sigue diciendo la verdad sobre cuándo se escribió, aunque la lista completa no
quede ordenada por él: `biso answer` añade al final un comentario con un instante pasado, el de la
pregunta que responde.

## Las fechas

`createdAt`, `updatedAt` y el instante de cada comentario los pone el programa con el reloj del
sistema, en UTC y con precisión de segundo.

**Se pueden fijar solo al importar**, es decir, en `biso new --from`. En cualquier otro sitio son un
hecho observado y no un dato que se negocie, **con una única excepción**: el instante de un
comentario ya escrito se puede corregir con `--set-comment-date`
(["Comentarios"](familias-de-banderas.md#comentarios), ["Borrar o corregir la fecha de un comentario"](../DECISIONES.md#borrar-o-corregir-la-fecha-de-un-comentario) de `DECISIONES.md`). Es una corrección de un dato ya
observado, no una negociación nueva, y por eso no abre la puerta a hacer lo mismo con `createdAt`,
`updatedAt` ni con `question.askedAt`.

**Una excepción de forma, no de fondo:** `biso answer` escribe el comentario en que se convierte la
pregunta con el instante en que esa pregunta se hizo, no con el de la respuesta. No negocia nada,
porque ese instante ya lo había observado el programa al crear la pregunta; solo lo traslada.

`question.askedAt` es una cuarta fecha importable, junto a `createdAt`, `updatedAt` y el instante de
cada comentario, y sigue la misma regla que ellas: es opcional, y si `biso new --from` no la trae,
toma el instante de la importación.

**`leaseExpiresAt` es la quinta fecha importable y es la única que no sigue esa regla**, así que se
cuenta aparte a propósito. Si no viene, no se rellena con nada: la tarea llega sin arrendamiento, que
es lo que significa no traerlo. Rellenarla con el instante de la importación crearía un arrendamiento
que nadie ha reclamado, y encima a nombre de nadie, porque `leaseHolder` no es una fecha y no tiene
ningún valor por defecto que ponerle. Los dos vienen juntos o no viene ninguno (la regla de
[La importación](lease.md#la-importación) en `lease.md`).

## La urgencia

`urgency` es un `float` derivado que se recalcula en cada lectura y **nunca se guarda**. Es el segundo
criterio de la tupla de orden por defecto de `biso ls`, después de `ordinal` ([`biso ls`](cmd/ls.md)), y el que ordena
el resumen de `biso prime`.

```
Si el estado de la tarea es el terminal, urgency = 0.0 y no se calcula nada mas.

En cualquier otro caso:

urgency = 6.0  * prioridad         (high 1.0, medium 0.5, low 0.0, sin prioridad 0.3)
        + 4.0  * activa            (1.0 si el estado es el activo y no hay pregunta abierta, 0.0 si no)
        + 8.0  * bloquea           (1.0 si alguna tarea sin terminar depende de esta)
        - 5.0  * bloqueada         (1.0 si depende de alguna tarea sin terminar)
        + 12.0 * proximidad        (ver la regla siguiente)
        + 1.0  * tiene_criterios   (1.0 si tiene al menos un criterio de aceptacion)
        + 0.5  * min(edad_dias / 30, 4.0)

El resultado se redondea a un decimal.
```

**La regla de `proximidad`, sin ambigüedad:**

```
dias = fecha_limite - hoy, en dias (puede ser negativo si la fecha ya paso)

si la tarea no tiene fecha limite:  proximidad = 0.0
si la tiene:                        proximidad = clamp((30 - dias) / 30, 0.0, 1.0)
```

Una tarea vencida tiene `dias` negativo, así que `(30 - dias) / 30` supera 1 y el resultado se acota
en **1.0**, el mismo máximo que una tarea que vence hoy. Una tarea vencida no suma más que una que
vence hoy; para distinguirlas está el filtro `--overdue` de `biso ls`, no un término sin tope en la
fórmula.

Un ejemplo completo, que es el que imprime `biso get --explain-urgency` en la sección [`biso get`](cmd/get.md): una tarea
de prioridad alta, en el estado activo, de la que depende otra tarea sin terminar, sin fecha límite,
con dos criterios y creada hoy, suma `6.0 + 4.0 + 8.0 + 0.0 + 0.0 + 1.0 + 0.0`, es decir **19.0**.

El valor de urgencia del ejemplo sale de los coeficientes por defecto, que el contrato de estabilidad
permite cambiar entre versiones menores, así que la cifra exacta puede no ser esta.

**Los coeficientes configurables son exactamente siete, bajo `urgency.`, uno por término de la
fórmula**: `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`,
`urgency.criteria` y `urgency.age`, con los valores de arriba (6.0, 4.0, 8.0, -5.0, 12.0, 1.0 y 0.5)
como valores por defecto. **Los pesos por prioridad no son configurables**: `high 1.0, medium 0.5,
low 0.0, sin prioridad 0.3` son parte de la estructura fija de la fórmula, que no cambia en la
versión 1.0.

**El `ordinal` no forma parte de la urgencia.** Es un orden manual que se aplica aparte, según la
regla de orden completa de la sección [`biso ls`](cmd/ls.md).

## Los campos externos

`ext` es un `map<string,string>` de clave a texto para guardar la identidad de una tarea en otro
sistema. La regla es la siguiente:

- El tablero **declara** en su configuración qué claves admite, en la lista `extensions`.
- Escribir una clave declarada funciona: `biso set MYP-1 --ext trello.card=5f2a8c1e3b9d4a7f`.
- Escribir una clave no declarada es error 3:
  ```
  error: unknown extension key: "jira.key"
         declared keys on this board: trello.card, github.issue
  ```
- Una tarea que ya guarda una clave que la configuración no declara **no se lee en silencio ni se
  reescribe perdiéndola**: se aplica la regla de ["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), y `biso doctor` la reporta.

## Quién reporta una tarea

`reporter` se fija una sola vez, al crear la tarea, y después solo cambia si alguien pasa
`--reporter` de forma explícita.

| Al crear la tarea | Valor de `reporter` |
|---|---|
| Se pasa `--reporter <persona>` | esa persona, tal cual |
| No se pasa, y hay identidad configurada | la identidad de quien llama, según la precedencia de ["Variables de entorno"](invocacion.md#variables-de-entorno) |
| No se pasa, y no hay identidad configurada | vacío, sin aviso |
| Se pasa `--reporter ""` | vacío |

El caso sin identidad no es un error y no imprime nada: a diferencia de `--mine`, de la
autoasignación de `biso start`, del autor de un comentario, de `biso ask` y de `biso answer`, que sí
la necesitan y están cubiertos por la tabla de ["Variables de entorno"](invocacion.md#variables-de-entorno), una tarea sin quien la reporte es válida.

En el lote de `biso new --from`, un objeto que trae `reporter` conserva ese valor, y uno que no lo
trae aplica las mismas reglas de esta tabla.

## La pregunta abierta

`question` es del tipo `Question`, un único valor y nunca una lista, con la misma forma que un
[`Comment`](#los-comentarios):

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | `string` libre | el programa, con la identidad `me`, salvo al importar |
| `askedAt` | `date` (instante UTC) | el programa, salvo al importar |
| `body` | `text` | quien llama |

Vacío es lo normal. Con contenido significa que la tarea espera la respuesta de una persona, esté en
el estado que esté, y entonces el derivado `waiting` es cierto; vacío, `waiting` es falso. Lleva tres
partes y no una sola porque al responderse se convierte literalmente en un comentario, con `biso
answer`, y para eso hacen falta su autor y su instante originales, no los de quien responde.

---

## Identificadores

- Un identificador es `<PREFIX>-<n>`, con `n` entero positivo. `PREFIX` viene de la configuración
  (`task_prefix`).
- **`task_prefix` no tiene un valor fijo por defecto: se deriva del nombre del tablero
  (`project_name`) en mayúsculas.** Dos tableros con `MYP` como valor fijo colisionarían los dos en
  `MYP-1`, y eso haría inservible cualquier vista que junte tareas de varios proyectos.
- **Los ejemplos de esta especificación pertenecen todos al mismo tablero ficticio, llamado `My
  project`, con `task_prefix` fijado a `MYP`.** Es el prefijo que aparece en los identificadores de
  ejemplo del resto del documento (`MYP-11`, `MYP-60`...), fijado a mano en vez de derivarse del
  nombre (que daría `MYPROJECT`), para que los ejemplos se lean como parte de un mismo tablero
  coherente.
- **La derivación quita del nombre los caracteres que no son letras y pasa el resto a mayúsculas**,
  así que un tablero llamado `mi-proyecto-2` da el prefijo `MIPROYECTO`. Si al quitarlos no queda
  ninguna letra, como en un tablero llamado `2026`, `biso init` no se inventa un valor: falla y pide
  el prefijo explícitamente con `--prefix` (error 2, `code` [`invalid_prefix`](contrato-json.md#los-identificadores-de-error)), la misma
  clave que ya cubre un `--prefix` con algo que no sean letras (la sección [`biso init`](cmd/init.md)).
- **"Letra" no incluye los diacríticos**, para que un nombre de tablero con cualquier carácter
  Unicode derive un prefijo predecible. La derivación pasa primero el nombre por el paso de
  `normalizar(x)` (sección ["El algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia)) que quita los acentos, las diéresis y las cedillas, y solo entonces
  se queda con lo que sean letras ASCII. Así un tablero llamado `Peña` deriva `PENA`, y uno llamado
  `Café` deriva `CAFE`.
- **Un identificador no se reutiliza jamás**, ni después de archivar una tarea ni después de
  eliminarla por cualquier vía.
- Los identificadores se asignan de forma creciente, pero **la especificación no promete que la
  secuencia no tenga huecos**. Un hueco es normal y nunca es un error.
- El tablero sabe en todo momento cuál es el identificador más alto que ha llegado a asignar, y ese
  dato se usa en los mensajes de la sección ["Los tres mensajes de \"no la encuentro\""](referencias.md#los-tres-mensajes-de-no-la-encuentro) y en `biso doctor`.
