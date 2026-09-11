# Diseño: el modelo de estados de `biso`

> **Este documento cita la especificación por el número de sus secciones**, como se escribió en su día.
> El 2026-09-10 `docs/SPEC.md` se repartió en los documentos de `docs/spec/`, que se citan por el título
> de sus secciones. Estas referencias se conservan sin tocar porque este documento es el acta de una
> sesión cerrada. Para traducir una de ellas, mira el mapa de la tabla de la tarea 3 de
> [el plan del reparto](../plans/2026-09-10-reparto-de-la-spec.md).

Este documento cierra los cuatro requisitos que la sección 9 de [`DECISIONES.md`](../../DECISIONES.md)
dejó identificados y sin incorporar. No es la especificación: es la decisión y su razón, escrita para
que aplicarla a [`SPEC.md`](../../spec/index.md) y a `DECISIONES.md` sea mecánico. La sección 11 es la lista
de esos cambios, uno a uno.

Lo que decide, en una frase: **ningún papel de estado nuevo, un campo nuevo con dos verbos, y el
encargo resuelto con la asignación que ya existe, retirando lo que la ensuciaba.**

Este documento va por su tercera versión. Las dos primeras pasaron sendas revisiones adversariales que
encontraron veinticinco problemas entre las dos, incluidos un tope que rompía el contrato de
estabilidad, un bloque nuevo que se tragaba el tablero, una bandera que hoy es un error de uso, y
quince sitios de `SPEC.md` sin contabilizar. Lo que sigue lleva esos arreglos, y la sección 12 recoge
lo que queda como riesgo aceptado.

---

## 1. El criterio que ordena todo lo demás

Los cuatro requisitos se enunciaron como si los cuatro pidieran un papel de estado. No es cierto, y
distinguirlo es lo que hace que el resultado sea pequeño. La regla es esta:

> **Algo es un estado cuando es excluyente con los demás y dice en qué punto del camino está la
> tarea. Es un campo cuando puede convivir con cualquier punto del camino.**

El documento ya aplicaba esta regla sin enunciarla: `archived` es un campo y no un estado,
precisamente porque una tarea archivada conserva el estado que tenía. Lo que faltaba era el criterio
escrito, para no volver a meter en el vocabulario de estados algo que no es un punto del camino.

| Requisito | Forma | Por qué |
|---|---|---|
| 9.1, distinguir el encargo de la ejecución | **Nada nuevo**: la asignación | "Esto lo tiene que hacer un agente" convive con cualquier punto del camino |
| 9.2, saber si alguien trabaja de verdad | Un campo, **aplazado** | Es un dato sobre quién y hasta cuándo, no sobre en qué punto está |
| 9.3, señalar lo que espera a una persona | Un campo, `question` | Una tarea puede esperar una respuesta desde cualquier estado |
| 9.4, distinguir terminar de descartar | **Retirado** | Falta la evidencia que lo justifique |

---

## 2. Vocabulario

Esta tabla entra en `SPEC.md` como sección sin numerar, detrás de la convención de idioma y delante de
la sección 1, para no renumerar las quince secciones existentes. `DECISIONES.md` remite a ella.

| Español | Qué es | En la interfaz |
|---|---|---|
| **tablero** | Las tareas de un proyecto con su configuración y sus vocabularios | `board` |
| **tarea** | La unidad de trabajo | `task` |
| **estado** | Uno de los valores configurados en `statuses` | `status` |
| **papel** | La función que cumple un estado. Son tres, obligatorios y distintos | las claves `*_status` |
| **inicial** | El papel del estado donde nace una tarea | `initial_status` |
| **activo**, activa | El papel del estado que escribe el agente al coger la tarea | `active_status` |
| **terminal**, terminada | El papel del estado final | `terminal_status` |
| **asignación** | Quién debe hacer la tarea. Es el gesto con el que una persona encarga trabajo | `assignees` |
| **arrendamiento** | Hasta cuándo vale la reserva de un agente sobre una tarea activa. No es un estado | pendiente de la persistencia |
| **pregunta abierta**, aparcada | Lo que detiene una tarea a la espera de una persona. No es un estado | `question`, `waiting` |
| **archivada** | Fuera del tablero activo sin perder nada. No es un estado | `archived` |
| **bloqueada** | Depende de alguna tarea sin terminar. Solo dependencias, nunca personas | `blocked` |
| **persona** | Quien encarga y quien responde | |
| **agente** | El programa automático que coge tareas y las hace | |
| **criterio** | Un elemento de las dos listas de comprobación | `acceptanceCriteria`, `definitionOfDone` |
| **comentario** | Una entrada inmutable del histórico | `comments` |

Cinco palabras quedan restringidas, en cuatro reglas, y conviene decir a qué en vez de prohibirlas a
secas, porque tres de ellas tienen un uso legítimo:

- **columna** nombra únicamente una columna de la tabla que imprimen `biso ls` y `biso prime`, que
  tiene ocho. Un estado del tablero no se llama nunca columna.
- **tarjeta** y **ticket** nombran únicamente lo que otra herramienta tiene, como una tarjeta de
  Trello. Lo de `biso` es una tarea.
- **panel** no se usa nunca: el conjunto de tareas es el tablero, y lo que abre `biso board` es la
  interfaz web.
- **bloqueada** no se usa nunca referida a una persona. Eso es una pregunta abierta.

---

## 3. Los tres papeles de estado

### 3.1. Los nombres son cosméticos, los papeles son el contrato

Esta frase entra en `SPEC.md` y es la que faltaba:

> Ninguna implementación se ramifica nunca por el nombre de un estado. Que un estado se llame
> `In Progress` o `En curso` no cambia una línea de código, porque todo lo que el programa decide lo
> decide por el papel. Cuántos estados hay y cuál cumple cada papel sí es estructural.

Es la regla cuya ausencia produjo la contradicción que se documenta en 3.3. Y obliga a algo que la
primera versión de este diseño incumplía en su propio ejemplo: **si nada puede depender del nombre de
un estado, tiene que haber una forma de filtrar por papel**, o quien consulte acabará cableando el
nombre. De ahí salen los filtros `--active` y `--not-active` del apartado 5.4.

### 3.2. `default_status` pasa a llamarse `initial_status`

`default` no dice nada, porque todo tiene un valor por defecto. `initial` dice exactamente lo que es,
el estado donde nace una tarea. Se descarta `init_status` porque `init` es un comando del programa y la
clave se leería como "el estado que crea `biso init`", que no es lo que significa.

El renombrado alcanza también a la clave `defaultStatus` de los esquemas JSON de `biso prime` y de
`biso init`, porque dos nombres para el mismo concepto es lo que prohíbe el principio 2. Por qué eso
no choca con el contrato de estabilidad está en el apartado 6.3.

### 3.3. La regla posicional se retira, porque estaba rota

`biso init` dice hoy que el penúltimo estado se guarda como estado activo. El tablero de ejemplo del
documento es `Ideas, To Do, In Progress, Blocked, Done`, cuyo penúltimo es `Blocked`, y hay un ejemplo
literal de `biso init` que lo crea así. Pero tanto la salida de `biso config list` como el esquema JSON
de `biso prime` declaran que en ese mismo tablero el estado activo es `In Progress`. Las dos cosas no
pueden ser ciertas. La regla funcionaba de casualidad porque el tablero por defecto tenía exactamente
tres estados.

No se arregla, se retira:

- **Sin `--statuses`**, `biso init` crea `To Do, In Progress, Done`, con los papeles inicial, activo y
  terminal en ese orden.
- **Con `--statuses`**, hacen falta `--initial-status`, `--active-status` y `--terminal-status`, con
  los mismos nombres que las claves de configuración.

El argumento para exigirlas está ya escrito en el apartado 6.1 de `DECISIONES.md`: una regla que
adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

Los cinco casos límite de `biso init`, todos con error 2 porque son problemas de argumentos, que es
como esa sección clasifica hoy cualquier problema de sus banderas:

| Caso | Mensaje |
|---|---|
| Falta alguna de las tres banderas de papel, habiendo `--statuses` | Las tres nombradas, diciendo cuáles faltan |
| Una bandera de papel sin `--statuses` | Que los papeles solo se fijan junto a la lista de estados |
| Una bandera de papel nombra un estado que no está en `--statuses` | El valor y la lista de estados |
| Dos banderas de papel nombran el mismo estado | Los dos papeles y el estado que comparten |
| `--statuses` con menos de tres estados | Cuántos hacen falta y por qué |

Y los dos que aparecen en `biso config set`, los dos con **error 6** y el identificador
`board_inconsistent` que ya existe, porque son de la misma familia que las cinco comprobaciones que esa
sección tiene hoy: dejar `statuses` con menos de tres estados, y dar a un papel un estado que ya tiene
otro papel.

### 3.4. El mínimo pasa de dos estados a tres

Los tres papeles son obligatorios y **distintos entre sí**, así que un tablero necesita al menos tres
estados. Con eso desaparece el caso de los dos estados, y con él su aviso de que `biso start` no cambia
el estado, que hay que quitar de la tabla de la sección 4.3 y de la salida de `biso init`.

Un tablero puede tener más de tres, y los que sobran no tienen papel. `Ideas`, `Review` o `Blocked`
siguen siendo estados perfectamente válidos; lo que ya no pueden es ser un papel del modelo.

`biso doctor` gana la comprobación de esta invariante, que hoy no la hace nadie: que `statuses` tenga
al menos tres elementos y que los tres papeles apunten a estados distintos y existentes.

### 3.5. Por qué el tablero por defecto no trae `Ideas`

Un estado `Ideas` dice lo mismo que estar sin asignar, que ya se consulta con `biso ls --unassigned`.
Y el otro matiz que aporta, el de "esto quizá no lo hagamos nunca", tiene desde el apartado 9.5 de
`DECISIONES.md` una decisión con evidencia detrás: se cubre con un tipo más del vocabulario, no con un
estado.

Hay un motivo peor para no ponerlo por defecto: si `Ideas` fuera el estado inicial, toda tarea nueva
nacería ahí y el bloque `NEXT UP` de `biso prime` mezclaría "algún día quizá" con "hay que hacerlo",
que es justo lo que ese bloque separa.

Se queda como estado opcional, y el documento lo usa en **un** ejemplo suelto, para ilustrar que un
tablero puede tener estados de más sin que nada se rompa.

---

## 4. El encargo es la asignación

### 4.1. El requisito estaba mal diagnosticado

El requisito 9.1 dice que con los papeles actuales el gesto de una persona que encarga trabajo y el de
un agente que lo coge **son el mismo dato**. Eso es falso, y por eso no hacía falta ningún estado
nuevo: la persona escribe `assignees`, el agente escribe el estado activo. Son dos datos distintos
escritos por dos actores distintos, y la asimetría que el requisito pedía existe ya.

Aplicando el criterio de la sección 1, además, no podía ser un estado: "esto lo tiene que hacer un
agente" convive con cualquier punto del camino, porque una tarea puede estar recién creada, a medias o
aparcada en una pregunta y seguir siendo suya.

### 4.2. Se retira `default_assignee`

Esta decisión no estaba en la primera versión de este diseño, y sin ella todo lo anterior es falso.

`default_assignee` es una clave de configuración que asigna una persona a toda tarea creada sin `-a`.
En un tablero que la use, **absolutamente todo nace asignado**, y con eso la asignación deja de
significar que alguien decidió que esa tarea la hagas tú. La consulta de arranque de un agente
devolvería el backlog entero, y el bloque `ASSIGNED TO YOU` de `biso prime` sería un volcado del
tablero.

Es una comodidad para ahorrar teclas que destruye la única señal sobre la que descansa la decisión del
apartado 4.1, así que se retira: la clave, su fila en la tabla de configuración, su comportamiento en
`biso new`, su aparición en el esquema JSON de `config list` y su línea en `biso config --help`.

**La autoasignación de `biso start` se queda**, porque no ensucia nada. Cuando `start` asigna a `me`,
la tarea pasa a la vez al estado activo, así que no entra en el conjunto de "asignada y sin empezar",
que es el que significa encargo.

### 4.3. Cómo queda el flujo

Un agente que arranca pregunta por lo suyo, y lo hace **por papel y no por nombre de estado**:

```
biso ls --mine --not-active --not-waiting   # lo que puede empezar
biso ls --mine --active --not-waiting       # lo que dejo a medias
```

Y una persona encarga con la misma llamada que crea la tarea:

```
biso new "Normalizar CRLF en el diff" --type bug -a @claude \
    -d "El parser normaliza LF pero no CRLF" \
    --ac "Un diff con CRLF da el mismo resultado que uno con LF" \
    --comment "Mira el parser, no el formateador"
```

### 4.4. Lo que esta decisión cuesta, dicho a propósito

**El caso de uso que motivó todo el trabajo deja de ser un arrastre.** El requisito 9.1 lo describe
como "arrastrar una tarjeta desde el móvil para que un agente se ponga con esa tarea". Con la
asignación como encargo, el gesto pasa a ser asignar un miembro, que en un tablero de móvil son dos
toques en vez de un arrastre y sigue siendo perfectamente viable, pero la frase del requisito ya no
describe la herramienta y hay que reescribirla.

La consecuencia mayor llega después: **cuando se especifique la sincronización con un sistema externo,
mover una tarjeta de columna en ese sistema no significará nada para `biso`**. Eso hay que decidirlo
entonces a propósito, y queda anotado aquí para que no se descubra por sorpresa.

**No hay forma de decir "esto es tuyo, pero todavía no".** Con la asignación como única señal, asignar
autoriza a empezar de inmediato. Quien necesite esa espera tiene dos salidas dentro del modelo: no
asignar hasta que toque, o usar una fecha límite.

**Solo el papel activo tiene filtros, y es deliberado.** No hay `--initial` ni `--terminal`. La
consecuencia es que en un tablero con estados de más, `--not-active` incluye esos estados extra, así
que un `Ideas` asignado al agente le llega mezclado con el trabajo. Se acepta porque la mezcla solo
ocurre si alguien asignó esa tarea a propósito, y `--mine` ya lleva esa intención dentro. Si algún día
molesta, se añaden los dos filtros que faltan sin tocar nada más.

---

## 5. La pregunta abierta

### 5.1. El campo y su booleano

Un campo nuevo de la tarea, `question`, con **las mismas tres partes que un comentario** de la sección
5.2 de `SPEC.md`:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | texto libre | el programa, con la identidad `me`, salvo al importar |
| `askedAt` | instante UTC | el programa, salvo al importar |
| `body` | texto largo | quien llama |

Vacío es lo normal. Con contenido significa que la tarea espera a una persona, esté en el estado que
esté. Lleva tres partes y no una porque al responderse **se convierte literalmente en un comentario**,
y para eso hacen falta su autor y su instante originales.

Y un booleano derivado, `waiting`, que es cierto cuando el campo tiene contenido. **No es una copia del
campo, es la relación que la especificación ya usa entre `dependencies` y `blocked`**: el dato y el
hecho que se deduce de él viajan juntos en la ficha, y los filtros y los listados hablan del hecho. Sin
él, `biso ls --json` no tendría forma de decir que una tarea está aparcada sin llevarse el cuerpo de la
pregunta, que es texto largo y es justo lo que el listado no lleva nunca.

**No es un campo escalar y no entra en la tabla de clases de campo de la sección 8.** Es un registro
con tres partes, y la tabla de clases, según el apartado 6 de `DECISIONES.md`, "no es una lista de
excepciones a la regla: es la regla aplicada a cada forma de dato". En su lugar se aplica el patrón que
la especificación ya usa para `archived`:

> **Ninguna bandera de campo escribe `question`.** Lo escriben `biso ask`, `biso answer` y la
> importación de `biso new --from`, y nadie más.

Eso resuelve tres problemas de una vez. No hay `--question` que sustituya en silencio una pregunta que
el verbo protege con un error. No hay `--clear-question` que la tire sin pasar por el histórico. Y no
hay que inventar una clase de campo para un registro compuesto.

### 5.2. Los dos verbos

Son verbos y no banderas por el principio 5, que dice que un gesto del flujo de trabajo es un comando.
Sin ellos, responder cuesta dos llamadas, una para el comentario y otra para vaciar el campo.

**Ninguno de los dos acepta `--comment-author`.** El autor de la pregunta y el de la respuesta son
siempre la identidad configurada, y sin ella los dos son error 2. La razón es literal: la sección 10.6
declara `--comment-author` como "requiere `--comment`" y convierte en error 2 pasarla sin él. Darle un
segundo destino aquí violaría el principio 2 y obligaría a cambiar esa regla. Quien necesite firmar
distinto tiene `--comment`, que sigue funcionando en los dos verbos con su autor de siempre.

**`biso ask <ref> <text>...`** llena el campo.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Se llena el campo. **No cambia el estado** |
| La tarea ya tiene una pregunta abierta | Error 6, para que la segunda no borre a la primera en silencio |
| La tarea está en el estado terminal | Error 6, igual que `biso start`, con la pista de reabrirla |
| La tarea está archivada | Se hace, con `note: TASK-11 is archived`, igual que `biso get` |
| El texto está vacío | Error 3, con el `code` `empty_scalar_value` que ya existe. Aparcar una tarea en una pregunta que no dice nada es peor que negarse |
| Sin identidad configurada | Error 2, con el `code` nuevo del apartado siguiente |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que `biso note` |
| Varias referencias | No se admiten: toma exactamente una, como `note` y `comment` |

**`biso answer <ref> <text>...`** vacía el campo, en una sola escritura y con tres efectos:

1. Añade al histórico un comentario con el `author`, el `askedAt` y el `body` que guardaba el campo.
2. Añade detrás un segundo comentario con la respuesta, firmado por la identidad configurada y con el
   instante de ahora.
3. Vacía el campo.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Error 6, con la pista de usar `biso comment` |
| Sin texto | Error 2. Una respuesta sin respuesta no cierra nada |
| Sin identidad configurada | Error 2, el mismo `code` que `ask` |
| Se pasan además banderas de campo | Se aplican igual, como en cualquier verbo del ciclo |

**El orden dentro de la escritura**, que la sección 4.9 exige fijar y no dejar al orden de la línea de
comandos: los dos comentarios del verbo se añaden **antes** que cualquier comentario que venga de
`--comment`, y el vaciado del campo es lo último.

Eso deja los comentarios en orden de inserción y no de instante, porque el de la pregunta lleva una
fecha pasada y se añade al final. **Los comentarios se guardan y se muestran en orden de inserción**,
que es lo que `SPEC.md` hace hoy sin decirlo, y el instante de cada uno dice la verdad. Conviene
escribirlo donde se define la lista.

Los dos aceptan las demás banderas de campo de la sección 8, como el resto de verbos del ciclo. El caso
de uso real de eso es responder y concretar a la vez:

```
biso answer TASK-11 "Solo los de texto. Los binarios se saltan enteros." \
    --ac "Un fichero binario no se toca"
```

**Códigos de salida de los dos**: los mismos de `biso comment` (0, 2, 3, 4, 5, 7, 8, 9), más el 6 de
sus tablas. Eso son **siete identificadores `code` nuevos** en la sección 12.3: tres del código 6 para
los tres casos de arriba, uno del código 2 para la falta de identidad, que hoy no tiene ninguno pese a
existir el error, y otros tres del código 2 para los casos de `biso init` que no encajan en los que ya
hay.

### 5.3. Tres excepciones declaradas

**A la sección 5.3 de `SPEC.md`**, que dice que el instante de un comentario lo pone el programa y solo
se puede fijar al importar. `biso answer` escribe un comentario con un instante pasado. No es una
excepción de fondo, porque ese instante lo observó el propio programa cuando se hizo la pregunta y
quien llama no lo negocia, pero es una excepción de forma y va escrita donde está la regla.

**A la lista de fechas importables**, que hoy son tres: `createdAt`, `updatedAt` y el instante de cada
comentario. `question.askedAt` es la cuarta, y sigue la misma regla que el instante de un comentario:
**es opcional al importar, y si falta se pone el instante de la importación**.

**Al ámbito único de búsqueda de texto de la sección 7.2**, que no lo usa solo `--search`: lo usa
también la resolución de una referencia por texto. Ampliarlo al cuerpo de la pregunta abierta significa
que `biso get "CRLF"` puede resolver a una tarea porque ese texto está en su pregunta, y que una
pregunta puede crear una ambigüedad de código 5 donde antes no la había. **Se acepta a propósito**: lo
contrario sería que el texto de una pregunta se pudiera encontrar justo al dejar de estar abierta,
cuando pasa a ser comentario, y no antes.

### 5.4. Alrededor del campo

- **`--waiting` y `--not-waiting`** en `biso ls` y en `biso export`, **incompatibles entre sí**. Son
  dos y no una porque el agente necesita excluir las aparcadas para elegir trabajo. Y son incompatibles
  porque juntas dan cero resultados en cualquier tablero, que es la afirmación falsa que el principio 1
  existe para impedir. La regla general que conviene escribir: **dos filtros que se contradicen por
  construcción son incompatibles; una combinación de filtros válidos que resulte vacía en este tablero
  es un hecho legítimo.**
- **`--active` y `--not-active`** en los dos mismos comandos, incompatibles entre sí por la misma
  razón. Filtran por el papel y no por el nombre del estado, y sin ellos no hay consulta portable entre
  tableros. Son compatibles con `-s`, con `--not-status` y con `--any-status`, porque filtran sobre el
  mismo eje pero no se contradicen: `-s "To Do" --active` es una lista vacía en unos tableros y no en
  otros, así que es un hecho y no una contradicción.
- **`question` viaja en `biso export` y lo acepta `biso new --from`**, con sus tres partes. La línea de
  ejemplo del NDJSON, que se presenta como la que lleva todos los tipos compuestos, tiene que
  incluirlo.
- **`biso ls --json` lleva `waiting` y no `question`**, exactamente como lleva `blocked` y no
  `dependencies`. El cuerpo de una pregunta es texto largo, y el motivo por el que el listado no lleva
  texto largo está medido en el principio 4: 215 fichas completas sumaron 179.369 bytes, casi la cuarta
  parte de la salida del estudio.
- **`biso get` gana una novena sección**, `question`, que se suma a las ocho de `--section`.
- **`biso start` sobre una tarea aparcada avisa y no lo impide**, igual que hace hoy con las
  dependencias sin terminar. Impedirlo empujaría a rodear la herramienta con `set`.
- **`biso finish` sobre una tarea aparcada avisa y no lo impide**, por el mismo motivo que no lo impide
  con los criterios sin marcar.

Los dos avisos entran en la tabla de la sección 4.3, que se declara a sí misma como la lista completa
de avisos y que ya hoy no lo es, porque le falta el de los dos estados de `biso init`. Ese, además,
desaparece con el mínimo de tres estados.

---

## 6. `biso prime`

### 6.1. Cuatro bloques, todos acotados

Entre `IN PROGRESS` y `NEXT UP` entran dos bloques:

- **`WAITING ON A PERSON`**: las tareas con pregunta abierta. Cada tarea ocupa **dos líneas**: la fila
  de siempre, con las ocho columnas del algoritmo de `biso ls`, y debajo una línea indentada con la
  pregunta **recortada a 100 caracteres**, la misma cifra exacta que el algoritmo aplica a los títulos.
  Va en línea propia y no en una novena columna porque el algoritmo tiene ocho exactas y una regla que
  dice que la octava nunca se rellena.
- **`ASSIGNED TO YOU`**: las asignadas a la identidad configurada que no estén en el estado activo y no
  tengan pregunta abierta.

**Ninguna tarea aparece en dos bloques.** Los cuatro se reparten el tablero por esta precedencia, y
cada tarea cae en el primero que la acepte:

1. `WAITING ON A PERSON`, si tiene una pregunta abierta.
2. `IN PROGRESS`, si está en el estado activo.
3. `ASSIGNED TO YOU`, si está asignada a la identidad configurada.
4. `NEXT UP`, el resto.

De ahí salen cuatro consecuencias. **Una tarea aparcada no sale en `IN PROGRESS`** aunque esté en el
estado activo, porque ese bloque significa que alguien está trabajando y ahí no lo está nadie. **Un
bloque sin filas no se imprime**, ni siquiera su encabezado, que es lo que ya hace `--limit 0` hoy con
`NEXT UP`. **Sin identidad configurada, `ASSIGNED TO YOU` no se imprime nunca**, igual que `biso prime`
ya tolera hoy no tener identidad imprimiendo `you are (not set)`. Y **`NEXT UP` deja de significar "sin
empezar"**, así que hay que renombrar tres cosas, con estos nombres concretos para que aplicarlo no
exija decidir nada:

| Hoy | Pasa a ser |
|---|---|
| `NEXT UP  (not started, not done, by urgency)` | `NEXT UP  (not assigned to you, by urgency)` |
| `54 more not started: 'biso ls --all'` | `54 more not shown: 'biso ls --not-active --not-waiting'` |
| La clave JSON `notStartedHidden` | `hiddenCount` |

**`--limit` acota los dos bloques juntos, no cada uno por su lado.** Su valor sigue siendo 5, y son
cinco filas repartidas entre `ASSIGNED TO YOU` y `NEXT UP`, en ese orden de preferencia, con **una sola
línea de recuento** al final del último de los dos que se imprima. Con `--limit 0` desaparecen los dos
y queda solo esa línea. Si cada bloque tuviera su propio límite, el resumen crecería al doble sin que
`--limit` lo notara.

**El orden de recorte se completa**, y esto arregla un agujero anterior a este diseño. La regla de hoy
dice que se recorta `NEXT UP` "antes que cualquier otra cosa" y no nombra ninguna otra cosa, así que un
tablero con muchas tareas en curso rebasa el tope sin conducta definida. El orden pasa a ser completo:

1. `NEXT UP`.
2. `ASSIGNED TO YOU`.
3. `WAITING ON A PERSON`.
4. `IN PROGRESS`.
5. Si aun así no cupiera, cada bloque se reduce a su línea de recuento.

Con esa lista **el tope deja de ser una aspiración y pasa a ser alcanzable siempre**, que es lo que una
prueba de la suite necesita.

**Los anchos de las columnas 1 a 7 se calculan sobre las filas que se van a imprimir en los cuatro
bloques juntas**, que es la extensión natural de la regla de hoy, escrita para dos, y lo que hace que
los cuatro se lean como una sola tabla. Las líneas de pregunta no entran en ese cálculo, porque no son
filas de la tabla.

Los cuatro bloques excluyen las tareas terminadas y las archivadas, igual que hoy.

### 6.2. El presupuesto: el tope total no se mueve

El añadido a la parte fija son unos 55 bytes de las dos órdenes nuevas y 214 de la regla nueva, en
total **269**. No hay ninguna bandera de campo nueva, así que la rejilla `FIELD FLAGS` no cambia. La
parte fija pasa de 2.963 a unos **3.232**.

El resumen del tablero pierde lo que ocupaban `Ideas` y `Blocked` en la línea de recuento y en las
filas, gana el bloque de preguntas con sus dos líneas por tarea, y gana un encabezado más sin ganar
filas, porque `--limit` acota los dos bloques juntos. Sale alrededor de **1.480** frente a los 1.099 de
hoy.

| Mitad | Antes | Ahora |
|---|---:|---:|
| Parte fija | 3.072 | **3.456** |
| Resumen del tablero | 2.048 | **1.664** |
| **Total** | **5.120** | **5.120** |

**El tope total no se mueve.** Lo que cambia es el reparto, y la dirección merece justificarse, porque
a primera vista contradice el motivo por el que existe el reparto. La sección 3 de `DECISIONES.md` dice
que las dos mitades están para que el resumen, que crece con el tablero, no se coma el sitio de las
reglas. Este diseño le quita sitio precisamente a esa mitad, y puede hacerlo por una razón que antes no
se cumplía: **con el orden de recorte completo del apartado anterior, el resumen ya no crece sin
límite**. Darle menos sitio cuesta filas mostradas, no correcciones. La parte fija, en cambio, no se
puede recortar sola: o cabe o hay que quitar contenido a mano.

Las cifras del añadido son estimaciones sobre texto que todavía no está escrito, y en particular la
regla nueva a la que se imputan 214 bytes no está redactada en ninguna parte. **Los números
definitivos salen de regenerar el ejemplo**, que es lo que el apartado 8 de `DECISIONES.md` exige. Si
al regenerarlo no cupieran, se recorta contenido, no se sube el tope.

La sección 9.5 gana además la frase que faltaba: **cuál de los dos números congela el contrato de
estabilidad**, que es el total, porque es el único que quien llama observa.

### 6.3. Las cinco cosas que este diseño hace y el contrato prohíbe

La sección 13 de `SPEC.md` lista lo que no cambia nunca. Este diseño la toca cinco veces, y conviene
tenerlas todas delante en vez de dos:

1. Renombra la clave JSON `defaultStatus`, y las claves de `data` "no se quitan".
2. Renombra la clave JSON `notStartedHidden`, por lo mismo.
3. Cambia la semántica de `--limit` en `biso prime`, y "una bandera nunca cambia de semántica".
4. Retira `default_assignee`, y retirar una funcionalidad exige un ciclo de aviso.
5. Mueve el reparto interno del presupuesto, aunque no el tope.

**El contrato obliga a partir de la 1.0, que no está publicada**, y la sección 13 no lo dice hoy: es lo
primero que hay que escribir ahí. Con eso, las cinco son legales.

La razón para no subir el tope de todas formas no es legal, es de fondo: **un tope que se sube cada vez
que aprieta no es un tope**, y su valor entero está en que obligue a elegir qué entra. Los cuatro
cambios restantes no debilitan ninguna promesa: corrigen nombres y una semántica antes de que nadie
dependa de ellos.

---

## 7. La urgencia

El término de actividad de la fórmula de la sección 5.4 pasa a estar condicionado:

```
+ 4.0 * activa   (1.0 si el estado es el activo Y no hay pregunta abierta, 0.0 si no)
```

Sin coeficiente nuevo y sin término nuevo: **siguen siendo exactamente siete**. El motivo es que una
tarea aparcada no la está trabajando nadie, y sin esta condición la fórmula afirmaría lo contrario y la
pondría arriba del listado del que un agente elige.

Eso obliga a tocar `biso get --explain-urgency` **en sus dos formas**. En el texto, la línea del
término activo dice cuál de los dos motivos lo anula. Y en el JSON, `urgencyBreakdown.active` es hoy un
número suelto, así que un programa vería el mismo `0.0` ambiguo: necesita acompañarlo del motivo.

---

## 8. El arrendamiento, aplazado

**La forma.** Arrendar es asignar con caducidad. Como la asignación ya existe, lo único que falta es un
instante de caducidad sobre una tarea activa y asignada. Un arrendamiento vencido **saca la tarea del
estado activo y le deja la asignación puesta**: lo que caduca es "estoy en ello", no "esto es tuyo".

**Lo que queda para la decisión de persistencia**: el nombre y el tipo del campo, cómo se renueva
mientras se trabaja, y quién detecta la caducidad sin que cueste caro.

**Por qué no se escribe el campo ahora.** Sería un campo que nada mantiene, que `biso export` tendría
que llevar y que la prueba de simetría tendría que cubrir, para un mecanismo que no existe.

Entra en la sección 14 de `SPEC.md`, cuyo propósito declarado es que nombrar lo que no está evite que
alguien lo dé por olvidado.

---

## 9. Lo retirado: el requisito 9.4

Se retira, y el argumento correcto no es el que parecía.

**El argumento que no vale.** Decir que archivar y descartar son lo mismo no se sostiene contra lo que
`SPEC.md` dice hoy de `biso archive`: existe `--unarchive`, que devuelve la tarea al tablero con el
estado que tenía, y la ayuda presenta el archivo como sacar tareas del tablero sin perderlas. Archivar
es reversible, así que es aparcar, y aparcar no es descartar.

**Y la deducción que proponía tampoco vale.** "Una tarea archivada que nunca llegó al estado terminal
es una tarea abandonada" no se puede calcular, porque el modelo guarda el estado actual y no un
histórico de estados. Con `biso start --reopen` una tarea terminada vuelve al activo, y archivada desde
ahí la deducción la llamaría abandonada habiendo estado hecha.

**El argumento que sí vale.** Los otros tres requisitos traen cada uno su evidencia medida: llamadas
contadas, bytes medidos, fallos reproducidos. Este no trae ninguna. Dice que hecha y abandonada "se
confunden", sin un solo caso en el que esa confusión haya costado algo. Un papel de estado que obliga a
dar un motivo es barato de añadir cuando haga falta y caro de quitar si sobra, así que **se queda fuera
hasta que haya un caso real que lo pida**, y entonces se engancha a `biso archive`. Entra también en la
sección 14.

---

## 10. El recorrido completo

Tablero `Kex` con estados `To Do, In Progress, Done`. La persona trabaja con `BISO_ME=@avilches` y el
agente con `BISO_ME=@claude`, sin la clave `me` en la configuración del tablero por el motivo que
explica la sección 12.

```
# 1. La persona encarga. El encargo es la asignacion.
biso new "Normalizar CRLF en el diff" --type bug -a @claude \
    -d "El parser normaliza LF pero no CRLF" \
    --ac "Un diff con CRLF da el mismo resultado que uno con LF" \
    --comment "Mira el parser, no el formateador"

# 2. El agente arranca y ve su cola.
biso prime
biso ls --mine --not-active --not-waiting

# 3. El agente la coge. Aqui el dato deja de ser ambiguo:
#    la persona escribio assignees, el agente escribe el estado.
biso start TASK-11 --plan "1. Leer el parser. 2. Anadir el caso CRLF."

# 4. El agente se atasca. No cambia el estado, pero sale de IN PROGRESS.
biso ask TASK-11 "Se normalizan tambien los binarios, o solo los de texto?"

# 5. La persona la ve donde mira.
biso prime
biso ls --waiting
biso get TASK-11 --section question

# 6. La persona contesta y concreta a la vez. Una sola escritura.
biso answer TASK-11 "Solo los de texto. Los binarios se saltan enteros." \
    --ac "Un fichero binario no se toca"

# 7. El agente sigue y cierra.
biso finish TASK-11 --check all --summary "Normaliza CRLF en texto, salta binarios."
```

Los filtros del día a día. **Las consultas del agente llevan `--not-waiting` porque `biso prime` aplica
la precedencia y `biso ls` no**, así que sin esa bandera los dos leerían el mismo tablero de dos
maneras distintas:

| Quién | Qué quiere saber | Comando |
|---|---|---|
| Agente | Todo lo necesario al arrancar | `biso prime` |
| Agente | Lo que puede empezar ahora | `biso ls --mine --not-active --not-waiting` |
| Agente | Lo que dejó a medias | `biso ls --mine --active --not-waiting` |
| Agente | Si algo suyo está esperando respuesta | `biso ls --mine --waiting` |
| Persona | Qué espera una respuesta suya | `biso ls --waiting` |
| Persona | Qué le ha encargado al agente | `biso ls -a @claude --not-active` |
| Persona | Qué está haciendo el agente ahora mismo | `biso ls -a @claude --active` |
| Persona | Qué no ha encargado a nadie | `biso ls --unassigned` |

---

## 11. La lista de cambios

### 11.1. En `SPEC.md`

**El modelo de datos y las reglas transversales**

| # | Dónde | Qué |
|---|---|---|
| 1 | Antes de la sección 1 | Entra la tabla de vocabulario de la sección 2 de este documento |
| 2 | Sección 4.3 | Los dos avisos nuevos; se quita el de los dos estados, que ya sobraba y ahora además no existe |
| 3 | Sección 4.12 | `ask` y `answer` entran en la enumeración de qué es una lectura dirigida |
| 4 | Sección 5 | Nuevo campo `question` con sus tres partes, y el derivado `waiting` |
| 5 | Sección 5.2 | Los comentarios se guardan y se muestran en orden de inserción |
| 6 | Sección 5.3 | La excepción del instante escrito por `answer`, y `askedAt` como cuarta fecha importable |
| 7 | Sección 5.4 | El término de actividad exige que no haya pregunta abierta |
| 8 | Sección 7.2 | El ámbito único de búsqueda alcanza el cuerpo de la pregunta, para `--search` y para la resolución de referencias |
| 9 | Sección 8 | `question` no entra en la tabla de clases de campo, y se dice por qué |

**Los comandos**

| # | Dónde | Qué |
|---|---|---|
| 10 | Sección 10, cabecera y tabla | De diecisiete comandos a diecinueve, y de nueve del ciclo a once |
| 11 | Sección 10, texto de cierre | `ask` y `answer` entran en la lista de comandos que aceptan banderas de campo |
| 12 | Sección 10.1 | Se retira la regla posicional; los cinco casos límite; se quita el aviso de los dos estados |
| 13 | Sección 10.3 | Se retira `default_assignee`; `new --from` acepta `question`; el ejemplo de NDJSON lo incluye |
| 14 | Sección 10.4 | Los cuatro filtros nuevos con sus incompatibilidades; `waiting` en el JSON del listado |
| 15 | Sección 10.5 | Novena sección `question`, con su cifra escrita cuatro veces en esa sección; `--explain-urgency` en texto y en JSON; el ámbito de búsqueda de su ayuda |
| 16 | Sección 10.6 | La lista de productores de `kind` `task.write` gana los dos verbos |
| 17 | Sección 10.7 | Los dos verbos nuevos, con su orden de aplicación dentro de la escritura |
| 18 | Sección 10.7.1 | `biso start` avisa sobre una tarea aparcada |
| 19 | Sección 10.7.4 | `biso finish` avisa sobre una tarea aparcada |
| 20 | Sección 10.9 | `question` en `export`, los cuatro filtros, y el guion de simetría con las banderas de papel |
| 21 | Sección 10.10 | `initial_status`; mínimo de tres estados; papeles distintos; se retira `default_assignee` |
| 22 | Sección 10.11 | `biso doctor` comprueba la invariante de los tres papeles |
| 23 | Sección 11 | La ayuda de primer nivel lista `ask` y `answer`, y su cifra de líneas |

**`biso prime` y los contratos**

| # | Dónde | Qué |
|---|---|---|
| 24 | Sección 9.3 | `--limit` acota los dos bloques juntos, y qué hace `--limit 0` |
| 25 | Sección 9.5 | El reparto y el total; qué bloques van en cada mitad; el orden de recorte completo; qué número congela el contrato |
| 26 | Sección 9.6 | De ocho órdenes del ciclo a diez, y de diez reglas a once |
| 27 | Sección 9.7 | Los dos bloques, la precedencia, el reparto de `--limit`, los tres renombrados y los anchos de columna |
| 28 | Sección 9.9 | Las dos listas nuevas del esquema JSON, `defaultStatus` a `initialStatus` y `notStartedHidden` a `hiddenCount` |
| 29 | Sección 9.11 | La ayuda de `prime` describe el `--limit` nuevo |
| 30 | Secciones 10.1 y 10.10 | `defaultStatus` e `initial_status` en sus esquemas JSON y en sus ayudas |
| 31 | Sección 12.1 | La fila `task.write` de la tabla gana los dos verbos |
| 32 | Sección 12.3 | Siete identificadores `code` nuevos, según el apartado 5.2 |
| 33 | Sección 13 | Desde cuándo obliga el contrato de estabilidad |
| 34 | Sección 14 | El arrendamiento aplazado y el requisito 9.4 retirado |
| 35 | Sección 15 | De cuatro verbos de ciclo a seis |

**Barridos sobre todo el documento**

| # | Qué |
|---|---|
| 36 | El tablero de ejemplo pasa a `To Do, In Progress, Done`; desaparecen `Ideas` y `Blocked` |
| 37 | Se regeneran los ejemplos de `ls` y de `prime`, que no se escriben a mano |
| 38 | Se aplica el vocabulario y se retiran las palabras prohibidas |

**Ocho cuentas publicadas en prosa cambian a la vez**, y su patrón de fallo es el que la sección 8 de
`DECISIONES.md` señala como dominante: los diecisiete comandos y los nueve del ciclo del cambio 10, las
ocho secciones fijas de `biso get` del cambio 15, que dentro de la sección 10.5 está escrita cuatro
veces, las treinta líneas del cambio 23, las ocho órdenes y las diez reglas del cambio 26, y los cuatro
verbos de ciclo del cambio 35. Hay una novena fuera de `SPEC.md`: el `CLAUDE.md` del proyecto dice "los
quince comandos" y el documento dice diecisiete, así que ya estaban descuadrados.

### 11.2. Fuera de `SPEC.md`

| # | Fichero | Qué |
|---|---|---|
| 1 | `DECISIONES.md` sección 9 | Se sustituye entera: los cuatro requisitos pasan a resueltos, aplazado o retirado |
| 2 | `DECISIONES.md` sección 9.1 | El diagnóstico correcto, la retirada de `default_assignee`, y qué cuesta eso |
| 3 | `DECISIONES.md` sección 9.4 | Por qué se retira: le falta la evidencia que los otros tres sí traen |
| 4 | `DECISIONES.md` sección 3 | El presupuesto, el reparto, el orden de recorte y el motivo de no subir el total |
| 5 | `DECISIONES.md` sección 4 | `default_assignee` entra en la lista de lo que se deja fuera, con su razón |
| 6 | `DECISIONES.md` sección 6 | Por qué se retira la regla posicional, y por qué el tablero por defecto no trae `Ideas` |
| 7 | `DECISIONES.md`, nueva | El criterio de estado frente a campo, y por qué la condición no es configurable |
| 8 | `DECISIONES.md`, nueva | Los riesgos conocidos de la sección 12 de este documento |
| 9 | `CLAUDE.md` | Su sección de cuatro requisitos sin incorporar queda obsoleta, y su cifra de comandos está mal |

---

## 12. Riesgos conocidos y aceptados

**Dos agentes con la misma identidad ven la misma cola.** Dos sesiones con `BISO_ME=@claude` no se
distinguen entre sí y cogerían la misma tarea. Es lo que resuelve el arrendamiento, que está aplazado.

**Un tablero con `me` en su configuración anula la distinción entre persona y agente.** La clave `me`
gana sobre `BISO_ME`, así que en un tablero que la tenga puesta todo el mundo es la misma identidad y
`--mine` deja de significar nada. Un tablero compartido entre una persona y un agente **tiene que dejar
`me` sin configurar**. Hay que escribirlo donde se explica la clave, porque hoy no está.

**La persona no tiene canal hacia el agente que se vea en `biso prime`.** El agente pregunta y la
persona responde, pero si la persona quiere decirle algo por iniciativa propia lo escribe en un
comentario, y `prime` no muestra comentarios. El agente lo ve al hacer `biso get`.

**`biso ls` no enseña la pregunta, solo dice qué tareas la tienen.** Es el precio de no meter texto
largo en el listado. Quien quiera leerlas hace `biso get --section question`, o mira `biso prime`, que
sí las enseña. El requisito 9.3 queda cubierto donde se mira al arrancar, no en el listado.

**Una pregunta abierta sobre una tarea terminada o archivada desaparece de la vista.** `finish` avisa
pero no impide, los bloques de `prime` excluyen terminadas y archivadas, y `biso ls` excluye el estado
terminal por defecto, así que `biso ls --waiting` no la encuentra sin `--any-status`. Se acepta porque
la alternativa, impedir cerrar una tarea con una pregunta abierta, empuja a rodear la herramienta.

**`--ready` no excluye las aparcadas.** `ready` mira solo dependencias, así que un agente que elija
trabajo con esa bandera, que es lo que su nombre invita a hacer, se las lleva. La consulta correcta
lleva `--not-waiting`. Está anotado porque el nombre engaña.

**`ready` y `blocked` siguen siendo complementarios**, dos campos derivados para un solo hecho. Es
anterior a este trabajo y queda como limpieza aparte.
