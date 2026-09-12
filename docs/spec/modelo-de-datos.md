# El modelo de datos de una tarea

Este es el modelo **lógico**. Describe qué campos tiene una tarea, de qué tipo son y quién los
escribe. No dice nada de cómo se guardan.

**Todo campo es de una de seis formas, y la tabla de abajo dice cuál para cada uno, sin excepción:**

| Forma | Qué es | Ningún elemento tiene fecha/autor propios, salvo que se diga | Se direcciona por |
|---|---|---|---|
| escalar | un solo valor: texto, número, booleano, fecha o instante | (no aplica, no es una lista) | no aplica |
| lista de tokens | varios valores del mismo tipo simple, separados por coma al escribir | cierto | por su propio valor, nunca por posición ni por clave |
| bloque de prosa | un solo texto largo, que puede tener saltos de línea | (no aplica) | no aplica |
| mapa de claves | pares clave-valor, con las claves declaradas de antemano | cierto | por su clave |
| lista de objetos | varios elementos, cada uno con más de un campo propio | depende del campo, se dice en su fila | por una clave estable que el programa asigna al crear, nunca por posición |
| registro de tres partes | como un elemento de una lista de objetos, pero solo hay uno, nunca una lista | depende del campo, se dice en su fila | no aplica, solo hay uno |

Con esto, la pregunta "¿tiene fecha y autor propios, y se puede señalar uno a uno?" se contesta para
cualquier campo sin salir de este documento: si su forma es "lista de objetos" o "registro de tres
partes", la fila de la tabla siguiente o su sección propia lo dice explícitamente; para las otras
cuatro formas la respuesta es siempre la misma y ya está en la tabla de arriba.

| Campo | Forma | Tipo lógico | Obligatorio | Quién lo fija | Mutable |
|---|---|---|---|---|---|
| `id` | escalar | identificador `PREFIX-<n>` | sí | el programa | no |
| `title` | escalar | texto de una línea | sí | quien llama | sí |
| `status` | escalar | uno del vocabulario de estados | sí | quien llama | sí |
| `type` | escalar | uno del vocabulario de tipos | no | quien llama | sí |
| `priority` | escalar | uno del vocabulario de prioridades | no | quien llama | sí |
| `project` | escalar | uno del vocabulario de proyectos | no | quien llama | sí |
| `milestone` | escalar | texto de hito | no | quien llama | sí |
| `parent` | escalar | referencia a otra tarea | no | quien llama | sí |
| `assignees` | lista de tokens | textos de persona | no | quien llama | sí |
| `reporter` | escalar | texto de persona | no | el programa al crear, o quien llama; ver ["Quién reporta una tarea"](#quién-reporta-una-tarea) | sí |
| `labels` | lista de tokens | textos | no | quien llama | sí |
| `dependencies` | lista de tokens | referencias a tareas | no | quien llama | sí |
| `references` | lista de tokens | textos | no | quien llama | sí |
| `documentation` | lista de tokens | textos | no | quien llama | sí |
| `modifiedFiles` | lista de tokens | textos | no | quien llama | sí |
| `due` | escalar | fecha `YYYY-MM-DD` | no | quien llama | sí |
| `ordinal` | escalar | entero >= 0 | no | quien llama | sí |
| `createdAt` | escalar | instante UTC | sí | el programa | solo al importar |
| `updatedAt` | escalar | instante UTC | sí | el programa | solo al importar |
| `archived` | escalar | booleano | sí, `false` por defecto | el programa, con `biso archive` | sí, solo con `biso archive` / `--unarchive`, o al importar |
| `leaseExpiresAt` | escalar | instante UTC | no | el programa, a `ahora + lease_minutes` (clave de [configuración](cmd/config.md)); ver la sexta precisión de abajo para cuándo | sí, ver las tres últimas precisiones de abajo, o al importar |
| `leaseHolder` | escalar | texto de persona | no | el programa, solo con [`biso start`](cmd/verbos-del-ciclo.md#biso-start) y con [`biso new --start`](cmd/new.md); ver la sexta precisión de abajo | sí, solo con esos dos, o al importar; ver las tres últimas precisiones de abajo |
| `urgency` | escalar | decimal, derivado | derivado | el programa | no, se recalcula al leer |
| `ext` | mapa de claves | clave declarada a texto | no | quien llama | sí |
| `description` | bloque de prosa | texto largo | no | quien llama | sí |
| `plan` | bloque de prosa | texto largo | no | quien llama | sí |
| `notes` | bloque de prosa | texto largo | no | quien llama | sí |
| `summary` | bloque de prosa | texto largo | no | quien llama | sí |
| `acceptanceCriteria` | lista de objetos | criterios, sin fecha ni autor propios; ver ["Los criterios y sus claves estables"](#los-criterios-y-sus-claves-estables) | no | quien llama | sí |
| `definitionOfDone` | lista de objetos | criterios, sin fecha ni autor propios; ver ["Los criterios y sus claves estables"](#los-criterios-y-sus-claves-estables) | no | quien llama | sí |
| `comments` | lista de objetos | comentarios, con fecha y autor propios; ver ["Los comentarios"](#los-comentarios) | no | quien llama | se añade, se borra entero, o se corrige solo la fecha; ver la precisión de abajo |
| `question` | registro de tres partes | con fecha y autor propios, igual que un comentario; ver ["La pregunta abierta"](#la-pregunta-abierta) | no | mixto, según la parte; ver ["La pregunta abierta"](#la-pregunta-abierta) | sí, solo con `biso ask`, `biso answer`, o al importar |
| `acDone`, `acTotal`, `dodDone`, `dodTotal` | escalar | entero, derivado | derivado | el programa | no, se recalculan al leer |
| `commentCount` | escalar | entero, derivado | derivado | el programa | no, se recalcula al leer |
| `blocks` | lista de tokens | referencias, derivado | derivado | el programa | no, se recalcula al leer |
| `blocked`, `waiting` | escalar | booleano, derivado | derivado | el programa | no, se recalculan al leer |
| `leaseExpired` | escalar | booleano, derivado | derivado | el programa | no, se recalcula al leer |

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
- **`leaseExpired` no cambia el `status` guardado, nunca.** Vale cierto cuando `leaseExpiresAt`
  tiene valor y ese instante es anterior al reloj de quien lee, y **vale falso cuando
  `leaseExpiresAt` está vacío**, que es el caso de toda tarea sin arrendamiento: no hay ningún
  estado en el que este derivado se quede sin valor, porque un derivado que no se pudiera calcular
  es justo lo que el principio 1 de la sección ["Los principios"](principios.md) no admite. Dice que el arrendamiento de una tarea
  activa venció, pero el estado guardado sigue siendo el activo hasta que alguien lo cambia con una
  escritura explícita: lo que vence es la reclamación, no el estado (sección ["Saber si alguien está trabajando de verdad"](../DECISIONES.md#saber-si-alguien-está-trabajando-de-verdad) de
  `DECISIONES.md`). No hay una escritura diferida que la saque del estado activo por su cuenta, porque
  eso haría que un comando tocara tareas que no nombró, y porque `biso prime`, que no escribe nunca,
  mostraría un estado que una escritura ajena y posterior podría cambiar. Liberar el arrendamiento
  vencido es la reclamación explícita que hace [`biso start`](cmd/verbos-del-ciclo.md#biso-start), no un efecto secundario de
  ningún otro comando.
- **Renovar `leaseExpiresAt` y fijar o transferir `leaseHolder` son cosas distintas, y solo la
  segunda pasa por `biso start` o por su atajo `biso new --start`.** Cualquier escritura sobre una
  tarea activa y asignada renueva `leaseExpiresAt` a `ahora + lease_minutes`, pero solo
  cuando quien llama ya es `leaseHolder`. **Cualquier escritura son todas**, sin ninguna excepción:
  los [seis verbos del ciclo](cmd/verbos-del-ciclo.md), [`biso set`](cmd/set.md) y [`biso archive`](cmd/archive.md), que son los
  comandos que llegan a escribir sobre una tarea que ya existe. Se nombran aquí porque una regla
  general que no nombra a nadie invita a buscarle excepciones donde no las hay. **Una escritura que
  no cambia ningún campo renueva igual**: [`biso set`](cmd/set.md) con todas sus banderas dando el valor que la
  tarea ya tiene sale con código 0 y con `note: MYP-11 unchanged`, y aun así renueva
  `leaseExpiresAt`, porque sigue siendo una escritura del tenedor sobre su tarea y el latido no
  puede depender de si los valores coincidían por casualidad. Esa renovación no toca `updatedAt`,
  porque ningún campo de la tarea ha cambiado, y deja vacía la lista `changed` del [esquema JSON](cmd/set.md); la nota sigue siendo cierta, porque habla de los campos de la tarea y ninguno cambió. Si la
  tarea no tiene arrendamiento todavía, escribir sobre ella no lo crea: fijarlo por primera vez es
  parte de lo que hace `biso start`, igual que reclamarlo vencido o tomarlo de [otra identidad](cmd/verbos-del-ciclo.md#biso-start). Una escritura de una identidad distinta de `leaseHolder` mientras el arrendamiento está
  vivo no toca ninguno de los campos: avisa con el mismo
  `warning: MYP-11's lease is held by @sara until 2026-09-08T14:00:00Z` de [`biso start`](cmd/verbos-del-ciclo.md#biso-start) y de la tabla de
  la sección ["Notas y avisos"](salida-y-terminal.md#notas-y-avisos), y el resto de la escritura se hace igual. **Con una sola excepción, y es que esa
  misma escritura rompa la invariante de la precisión siguiente**: si deja la tarea fuera del estado
  activo, sin ninguna persona asignada o archivada, los campos se vacían en esa misma escritura,
  sea quien sea quien la haga, y el aviso de que el arrendamiento era de otra identidad se emite
  igual. Una escritura de una identidad distinta mientras el arrendamiento está vencido tampoco lo
  toca, y lo deja vencido: quien comenta, anota o cierra una tarea no ha reclamado nada. **Reclamar
  es de [`biso start`](cmd/verbos-del-ciclo.md#biso-start) y de su atajo [`biso new --start`](cmd/new.md), y de nadie más**, con una
  excepción que hay que nombrar porque sin ella la frase sería falsa: `biso start -s <estado>` con
  un estado que no es el activo no fija arrendamiento, ya que fijarlo ahí rompería la invariante de
  la precisión siguiente, y deja los campos como los dejaría cualquier otra escritura. Una tarea
  que llega a activa y asignada por cualquier otra vía no tiene arrendamiento hasta que alguien
  llame a `biso start` sobre ella, y esas vías son exactamente dos: las banderas de campo de la
  sección ["Las familias de banderas"](familias-de-banderas.md), por ejemplo `biso set --status`, ninguna de las cuales lo puede crear, y la
  importación, que es de lo que trata la última precisión.
- **Los campos solo tienen valor en una tarea activa y asignada, y se vacían al perder
  cualquiera de las dos condiciones, no solo la primera.** Una escritura que saca la tarea del
  estado activo (`biso finish`, o `biso set --status` a cualquier otro valor) vacía
  `leaseExpiresAt` y `leaseHolder` en esa misma escritura. Y como la condición que los sostiene es
  la conjunción de las dos cosas, perder la segunda los vacía igual: `--clear-assignees` o
  [`--rm-assignees`](familias-de-banderas.md#campos-de-lista-que-admiten-coma) sobre una tarea activa que se queda sin ninguna persona asignada vacía los
  campos en esa misma escritura, sea quien sea quien la haga. **[`biso archive`](cmd/archive.md) los vacía
  también**, aunque `archived` no sea un estado y archivar no saque la tarea del estado activo:
  archivar es dejar de trabajar en la tarea, y un arrendamiento es la afirmación de que alguien está
  trabajando ahora, así que conservarlo lo guardaría donde nadie lo ve, porque `biso prime` y
  `biso ls` excluyen las archivadas por defecto, y `--unarchive` la devolvería al tablero semanas
  después a nombre de una sesión que ya murió. **Esta precisión gana siempre sobre la anterior, y
  por eso la invariante se enuncia aquí y el aviso allí.** Cuando quien escribe no es
  `leaseHolder`, el aviso de que el arrendamiento es de otra identidad se emite igual, pero los dos
  campos se vacían: `@sara` haciendo `biso finish MYP-11` sobre una tarea arrendada por `@claude` la
  deja terminada y sin arrendamiento. Con la precedencia al revés quedaría una tarea terminada con un
  arrendamiento vivo, que es exactamente lo que la última precisión rechaza al importar, así que
  `biso export` produciría un fichero que su propio `biso init --from` rechaza y la prueba de
  simetría de la sección ["El contrato de estabilidad"](estabilidad.md) fallaría (sección ["Saber si alguien está trabajando de verdad"](../DECISIONES.md#saber-si-alguien-está-trabajando-de-verdad) de `DECISIONES.md`). **Y los campos van
  siempre juntos**: ninguna escritura, y tampoco la importación, deja uno con valor y el otro vacío.
- **La importación los escribe con el valor que traiga el fichero, y es la única vía que lo hace.**
  Los dos son campos guardados y no derivados, así que [`biso export`](cmd/export.md) los escribe y [`biso new --from`](cmd/new.md)
  los lee de vuelta como cualquier otro, que es lo que hace cierta la garantía de simetría de [`biso export`](cmd/export.md)
  sin una lista de excepciones que mantener. La invariante de la precisión anterior se comprueba al
  importar, y en sus dos mitades. Una línea que traiga `leaseExpiresAt` o `leaseHolder` sobre una
  tarea que no esté a la vez en el estado activo y asignada a alguien es un fallo de validación del
  [lote](cmd/new.md), igual que una clave desconocida. Y una línea que traiga uno de los campos y no el
  otro es el mismo fallo, con el mismo trato: los dos vienen juntos o no viene ninguno, porque un
  `leaseHolder` sin `leaseExpiresAt` sería un arrendamiento que no caduca nunca, y un
  `leaseExpiresAt` sin `leaseHolder` una reserva de nadie. Un arrendamiento importado no privilegia
  a nadie: `leaseExpired` se recalcula contra el reloj de la máquina que lee, así que el que llegue
  caducado sale caducado y [`biso start`](cmd/verbos-del-ciclo.md#biso-start) lo reclama, y el que llegue vivo a nombre de otra
  identidad solo produce el aviso de la sección ["Notas y avisos"](salida-y-terminal.md#notas-y-avisos) hasta que caduque.

## Los criterios y sus claves estables

`acceptanceCriteria` y `definitionOfDone` son listas del mismo tipo, y cada elemento tiene tres cosas:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | entero positivo | el programa al crear el elemento |
| `text` | texto | quien llama |
| `checked` | booleano | quien llama |

**La clave se asigna al crear el elemento, con un contador propio de esa lista dentro de esa tarea, y
no se reasigna nunca.** Quitar un elemento no mueve las claves de los demás: una tarea puede tener
perfectamente los criterios `#1` y `#3` y ninguno más. Cada tarea lleva dos contadores, uno por
lista, que solo crecen.

Consecuencias que hay que respetar en toda la implementación:

- Los selectores de la sección ["Selectores de criterios"](familias-de-banderas.md#selectores-de-criterios) trabajan sobre la clave, **nunca** sobre la posición.
- `acTotal` y `dodTotal`, allá donde aparezcan, son **el número de elementos presentes**, nunca la
  clave más alta. Una tarea con los criterios `#1` y `#3` tiene `acTotal` igual a 2.
- Los elementos se muestran y se exportan en el orden en que están en la lista, que es el orden en
  que se crearon salvo que se haya sustituido la lista entera.

## Los comentarios

Cada comentario tiene clave, autor, instante y cuerpo:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | entero positivo | el programa al crear el comentario |
| `author` | texto libre | quien llama, y por defecto la identidad `me` |
| `createdAt` | instante UTC | el programa, salvo al importar o con `--set-comment-date` (["Comentarios"](familias-de-banderas.md#comentarios)) |
| `body` | texto largo | quien llama |

**La clave se asigna al crear el comentario, con un contador propio de esa lista dentro de esa
tarea, y no se reasigna nunca**, exactamente igual que la de un criterio
(["Los criterios y sus claves estables"](#los-criterios-y-sus-claves-estables)). Borrar un comentario no mueve las claves de los
demás.

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
ningún valor por defecto que ponerle. Los dos vienen juntos o no viene ninguno (precisión octava de
esta misma sección).

## La urgencia

`urgency` es un decimal derivado que se recalcula en cada lectura y **nunca se guarda**. Es el segundo
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

`ext` es un mapa de clave a texto para guardar la identidad de una tarea en otro sistema. La regla
es la siguiente:

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

`question` es un registro de tres partes, con la misma forma que [un comentario](#los-comentarios):

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | texto libre | el programa, con la identidad `me`, salvo al importar |
| `askedAt` | instante UTC | el programa, salvo al importar |
| `body` | texto largo | quien llama |

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
  normalizar(x)` (sección ["El algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia)) que quita los acentos, las diéresis y las cedillas, y solo entonces
  se queda con lo que sean letras ASCII. Así un tablero llamado `Peña` deriva `PENA`, y uno llamado
  `Café` deriva `CAFE`.
- **Un identificador no se reutiliza jamás**, ni después de archivar una tarea ni después de
  eliminarla por cualquier vía.
- Los identificadores se asignan de forma creciente, pero **la especificación no promete que la
  secuencia no tenga huecos**. Un hueco es normal y nunca es un error.
- El tablero sabe en todo momento cuál es el identificador más alto que ha llegado a asignar, y ese
  dato se usa en los mensajes de la sección ["Los tres mensajes de \"no la encuentro\""](referencias.md#los-tres-mensajes-de-no-la-encuentro) y en `biso doctor`.

