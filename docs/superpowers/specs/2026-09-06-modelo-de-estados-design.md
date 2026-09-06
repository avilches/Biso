# Diseño: el modelo de estados de `biso`

Este documento cierra los cuatro requisitos que la sección 9 de [`DECISIONES.md`](../../DECISIONES.md)
dejó identificados y sin incorporar. No es la especificación: es la decisión y su razón, escrita para
que aplicarla a [`SPEC.md`](../../SPEC.md) y a `DECISIONES.md` sea mecánico. La sección 11 de este
documento es la lista de esos cambios, uno a uno.

Lo que decide, en una frase: **ningún papel de estado nuevo, un campo nuevo con dos verbos, y el
encargo resuelto con la asignación que ya existe, retirando lo que la ensuciaba.**

Este documento va por su segunda versión. La primera pasó una revisión adversarial que encontró trece
sitios de `SPEC.md` sin contabilizar, una subida de tope que rompía el contrato de estabilidad, y un
bloque nuevo de `biso prime` que se tragaba el tablero entero. Lo que sigue ya lleva esos arreglos, y
la sección 12 recoge lo que quedó como riesgo aceptado.

---

## 1. El criterio que ordena todo lo demás

Los cuatro requisitos se enunciaron como si los cuatro pidieran un papel de estado. No es cierto, y
distinguirlo es lo que hace que el resultado sea pequeño. La regla es esta:

> **Algo es un estado cuando es excluyente con los demás y dice en qué punto del camino está la
> tarea. Es un campo cuando puede convivir con cualquier punto del camino.**

El documento ya aplicaba esta regla sin enunciarla: `archived` es un campo y no un estado,
precisamente porque una tarea archivada conserva el estado que tenía. Lo que faltaba era el criterio
escrito, para no volver a meter en el vocabulario de estados algo que no es un punto del camino.

Aplicado a los cuatro requisitos:

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
| **pregunta abierta**, aparcada | Lo que detiene una tarea a la espera de una persona. No es un estado | `question` |
| **archivada** | Fuera del tablero activo sin perder nada. No es un estado | `archived` |
| **bloqueada** | Depende de alguna tarea sin terminar. Solo dependencias, nunca personas | `blocked` |
| **persona** | Quien encarga y quien responde | |
| **agente** | El programa automático que coge tareas y las hace | |
| **criterio** | Un elemento de las dos listas de comprobación | `acceptanceCriteria`, `definitionOfDone` |
| **comentario** | Una entrada inmutable del histórico | `comments` |

Palabras que el documento no vuelve a usar, porque cada una tiene ya su término: **columna** (es un
estado), **panel** (es el tablero; lo que abre `biso board` es la interfaz web), **tarjeta** y
**ticket** (es una tarea), y **bloqueada** referida a una persona (eso es una pregunta abierta).

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

No se arregla, se retira. En su lugar:

- **Sin `--statuses`**, `biso init` crea `To Do, In Progress, Done`, con los papeles inicial, activo y
  terminal en ese orden.
- **Con `--statuses`**, hacen falta `--initial-status`, `--active-status` y `--terminal-status`, con
  los mismos nombres que las claves de configuración.

El argumento para exigirlas está ya escrito en el apartado 6.1 de `DECISIONES.md`: una regla que
adivina acierta a veces, y acertar a veces es peor que fallar siempre, porque enseña a confiar.

Los cuatro casos límite de `biso init`, todos con error 2 porque son problemas de argumentos, que es
como esa sección clasifica hoy cualquier problema de sus banderas:

| Caso | Mensaje |
|---|---|
| Falta alguna de las tres banderas de papel | Las tres nombradas, diciendo cuáles faltan |
| Una bandera de papel nombra un estado que no está en `--statuses` | El valor y la lista de estados |
| Dos banderas de papel nombran el mismo estado | Los dos papeles y el estado que comparten |
| `--statuses` con menos de tres estados | Cuántos hacen falta y por qué |

Y los dos que aparecen en `biso config set`, los dos con **error 6** y el identificador
`board_inconsistent` que ya existe, porque son de la misma familia que las cinco comprobaciones que esa
sección tiene hoy: dejar `statuses` con menos de tres estados, y dar a un papel un estado que ya tiene
otro papel.

### 3.4. El mínimo pasa de dos estados a tres

Los tres papeles son obligatorios y **distintos entre sí**, así que un tablero necesita al menos tres
estados. Con eso desaparece el caso de los dos estados y su aviso de que `biso start` no cambia el
estado.

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

Un agente que arranca pregunta por lo suyo, y lo hace **por papel y no por nombre de estado**, con los
filtros del apartado 5.4:

```
biso ls --mine --not-active --not-waiting   # lo que puede empezar
biso ls --mine --active                     # lo que dejo a medias
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
asignar hasta que toque, o usar una fecha límite. Un estado extra sin papel no sirve, porque un agente
no puede consultarlo de forma portable entre tableros sin cablear su nombre, que es justo lo que
prohíbe el apartado 3.1.

---

## 5. La pregunta abierta

### 5.1. El campo

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

No se añade ningún booleano derivado: que el campo tenga contenido ya lo dice todo, y añadir uno
repetiría la redundancia que ya existe entre `ready` y `blocked`.

**No es un campo escalar y no entra en la tabla de clases de campo de la sección 8.** Es un registro
con tres partes, y la tabla de clases, según el apartado 6 de `DECISIONES.md`, "no es una lista de
excepciones a la regla: es la regla aplicada a cada forma de dato". Declararlo escalar sería la primera
excepción de esa tabla. En su lugar se aplica el patrón que la especificación ya usa para `archived`:

> **Ninguna bandera de campo escribe `question`.** Lo escriben `biso ask`, `biso answer` y la
> importación de `biso new --from`, y nadie más.

Eso resuelve tres problemas de una vez. No hay `--question` que sustituya en silencio una pregunta que
el verbo protege con un error. No hay `--clear-question` que la tire sin pasar por el histórico. Y no
hay que inventar una clase de campo para un registro compuesto.

### 5.2. Los dos verbos

Son verbos y no banderas por el principio 5, que dice que un gesto del flujo de trabajo es un comando.
Sin ellos, responder cuesta dos llamadas, una para el comentario y otra para vaciar el campo.

**`biso ask <ref> [<text>...]`** llena el campo. El autor es siempre la identidad configurada y no se
puede pasar por bandera, para que `--comment-author` tenga un solo destino en todo el programa, que es
el autor de un comentario.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Se llena el campo. **No cambia el estado** |
| La tarea ya tiene una pregunta abierta | Error 6, para que la segunda no borre a la primera en silencio |
| Sin identidad configurada | Error 2, con el mismo mensaje que el autor de un comentario |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que `biso note` |
| Varias referencias | No se admiten: toma exactamente una, como `note` y `comment` |

**`biso answer <ref> [<text>...] [--comment-author <@who>]`** vacía el campo, en una sola escritura y
con tres efectos:

1. Añade al histórico un comentario con el `author`, el `askedAt` y el `body` que guardaba el campo.
2. Añade detrás un segundo comentario con la respuesta, firmado por quien contesta y con el instante
   de ahora. `--comment-author` afecta a este y solo a este.
3. Vacía el campo.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Error 6, con la pista de usar `biso comment` |
| Sin `--comment-author` y sin identidad configurada | Error 2, el mismo de `biso comment` |
| Se pasan además banderas de campo | Se aplican igual, como en cualquier verbo del ciclo |

**El orden dentro de la escritura**, que la sección 4.9 exige fijar y no dejar al orden de la línea de
comandos: los dos comentarios del verbo se añaden **antes** que cualquier comentario que venga de
`--comment`, y el vaciado del campo es lo último. Así la conversación queda en el orden en que ocurrió.

Los dos aceptan las demás banderas de campo de la sección 8, como el resto de verbos del ciclo. El caso
de uso real de eso es responder y concretar a la vez:

```
biso answer TASK-11 "Solo los de texto. Los binarios se saltan enteros." \
    --ac "Un fichero binario no se toca"
```

**Códigos de salida de los dos**: los mismos de `biso comment` (0, 2, 3, 4, 5, 7, 8, 9), más el 6 de
las dos filas de arriba, con dos identificadores `code` nuevos en la tabla de la sección 12.3. El error
2 por falta de identidad reutiliza el `code` que ya use `biso comment` para ese mismo caso.

### 5.3. Dos excepciones declaradas

**A la sección 5.3 de `SPEC.md`**, que dice que el instante de un comentario lo pone el programa y solo
se puede fijar al importar. `biso answer` escribe un comentario con un instante pasado. No es una
excepción de fondo, porque ese instante lo observó el propio programa cuando se hizo la pregunta y
quien llama no lo negocia, pero es una excepción de forma y va escrita donde está la regla.

**A la lista de fechas importables**, que hoy son tres (`id` aparte): `createdAt`, `updatedAt` y el
instante de cada comentario. `question.askedAt` es la cuarta, y sigue la misma regla que el instante de
un comentario: **es opcional al importar, y si falta se pone el instante de la importación**.

### 5.4. Alrededor del campo

- **`--waiting` y `--not-waiting`** en `biso ls` y en `biso export`. Son dos y no una porque el agente
  necesita **excluir** las aparcadas para elegir trabajo, y ese es su filtro más usado.
- **`--active` y `--not-active`** en los dos mismos comandos, que filtran por el papel y no por el
  nombre del estado. Sin ellos no hay forma de escribir una consulta portable entre tableros, que es lo
  que exige el apartado 3.1. Son incompatibles entre sí, como `--ready` y `--blocked`.
- **`question` viaja en `biso export` y lo acepta `biso new --from`**, con sus tres partes, porque no
  es un campo derivado y la ida y vuelta es una prueba de la suite. La línea de ejemplo del NDJSON, que
  se presenta como la que lleva todos los tipos compuestos, tiene que incluirlo.
- **`biso get` gana una novena sección**, `question`, que se suma a las ocho de `--section`.
- **`biso ls --json` lleva `question`**, entero. Es la única forma de que `biso ls --waiting --json`
  sirva para algo, y no choca con la regla de que el listado no trae el cuerpo de la tarea, porque una
  pregunta abierta no es cuerpo: es el dato que hace falta para actuar sobre la fila.
- **`--search` busca dentro del cuerpo de la pregunta abierta**, porque si no lo hiciera el texto de
  una pregunta sería encontrable justo al dejar de estar abierta, cuando pasa a ser un comentario, y no
  antes.
- **`biso start` sobre una tarea aparcada avisa y no lo impide**, igual que hace hoy con las
  dependencias sin terminar. Impedirlo empujaría a rodear la herramienta con `set`.
- **`biso finish` sobre una tarea aparcada avisa y no lo impide**, por el mismo motivo que no lo impide
  con los criterios sin marcar.

Los dos avisos entran en la tabla de la sección 4.3, que se declara a sí misma como la lista completa
de avisos que el programa emite.

---

## 6. `biso prime`

### 6.1. Dos bloques nuevos, los dos acotados

Entre `IN PROGRESS` y `NEXT UP` entran dos bloques:

- **`WAITING ON A PERSON`**: las tareas con pregunta abierta. Cada tarea ocupa **dos líneas**: la fila
  de siempre, con las ocho columnas del algoritmo de `biso ls`, y debajo una línea indentada con la
  pregunta recortada a una línea. Va en línea propia y no en una novena columna porque el algoritmo de
  columnas tiene ocho exactas y una regla que dice que la octava nunca se rellena; una novena obligaría
  a cambiarlo, y está publicado en dos sitios.
- **`ASSIGNED TO YOU`**: las asignadas a la identidad configurada que no estén en el estado activo y no
  tengan pregunta abierta.

**Ninguna tarea aparece en dos bloques.** Los cuatro se reparten el tablero por esta precedencia, y
cada tarea cae en el primero que la acepte:

1. `WAITING ON A PERSON`, si tiene una pregunta abierta.
2. `IN PROGRESS`, si está en el estado activo.
3. `ASSIGNED TO YOU`, si está asignada a la identidad configurada.
4. `NEXT UP`, el resto.

De ahí salen tres consecuencias que conviene ver escritas. **Una tarea aparcada no sale en
`IN PROGRESS`** aunque esté en el estado activo, porque ese bloque significa que alguien está
trabajando y ahí no lo está nadie. **`NEXT UP` deja de significar "sin empezar"**, así que su rótulo,
su línea de recuento y la clave `notStartedHidden` del JSON cambian de nombre. Y **sin identidad
configurada el bloque `ASSIGNED TO YOU` no se imprime**, igual que `biso prime` ya tolera hoy no tener
identidad imprimiendo `you are (not set)`.

**`--limit` acota los dos bloques juntos, no cada uno por su lado.** Su valor sigue siendo 5, y son
cinco filas repartidas entre `ASSIGNED TO YOU` y `NEXT UP`, en ese orden de preferencia, con una sola
línea de recuento al final que dice cuántas quedaron fuera de los dos. Esa es la diferencia que impide
que el mensaje crezca con el tablero: si cada bloque tuviera su propio límite, el resumen crecería al
doble sin que `--limit` lo notara.

`IN PROGRESS` y `WAITING ON A PERSON` siguen sin límite, por el argumento que ya se aplica al primero:
en un tablero sano son pocas. Si aun así el resumen no cupiera en su mitad, **el orden de recorte es
`NEXT UP`, luego `ASSIGNED TO YOU`, luego `WAITING ON A PERSON`**, y la línea de recuento lo dice. Esa
regla existe hoy nombrando un solo bloque y hay que ampliarla.

Los cuatro bloques excluyen las tareas terminadas y las archivadas, igual que hoy.

### 6.2. El presupuesto: el tope total no se mueve

El añadido a la parte fija son unos 55 bytes de las dos órdenes nuevas, 12 de la línea de la rejilla y
214 de la regla nueva, en total **281**. La parte fija pasa de 2.963 a unos **3.244**.

El resumen del tablero pierde 12 bytes por el estado `Blocked` que desaparece de la línea de recuento,
gana el bloque de preguntas con sus dos líneas por tarea, y gana un encabezado más sin ganar filas,
porque `--limit` acota los dos bloques juntos. Sale alrededor de **1.497** frente a los 1.099 de hoy.

| Mitad | Antes | Ahora |
|---|---:|---:|
| Parte fija | 3.072 | **3.456** |
| Resumen del tablero | 2.048 | **1.664** |
| **Total** | **5.120** | **5.120** |

**El tope total no se mueve.** Lo que cambia es el reparto interno, porque la parte fija crece con la
herramienta y el resumen no tiene por qué. Con esas cifras quedan 212 bytes de holgura en la parte fija
y unos 167 en el resumen.

Las tres cifras del añadido son estimaciones sobre texto que todavía no está escrito, y las del resumen
dependen de cuántas preguntas abiertas haya. **Los números definitivos salen de regenerar el ejemplo**,
que es lo que el apartado 8 de `DECISIONES.md` exige para cualquier salida del documento. Si al
regenerarlo no cupieran, se recorta contenido, no se sube el tope.

La sección 9.5 gana además la frase que faltaba: **cuál de los dos números congela el contrato de
estabilidad**, que es el total, porque es el único que quien llama observa.

### 6.3. Por qué el contrato de estabilidad no se toca, y por qué el renombrado sí puede

La sección 13 de `SPEC.md` lista entre lo que no cambia nunca "El tope de tamaño del mensaje de
`biso prime`" y "Las claves de `data` en cada `kind` de JSON", que no se quitan. Este diseño mantiene el
tope y renombra una clave, así que hay que decir por qué eso no es incoherente.

**El contrato obliga a partir de la 1.0, que no está publicada**, y la sección 13 no lo dice hoy: es lo
primero que hay que escribir ahí. Con eso, ni el renombrado ni una subida de tope serían ilegales
todavía.

La razón para no subir el tope de todas formas no es legal, es de fondo: **un tope que se sube cada vez
que aprieta no es un tope**, y su valor entero está en que obligue a elegir qué entra. El renombrado,
en cambio, no debilita ninguna promesa: corrige un nombre antes de que nadie dependa de él.

---

## 7. La urgencia

El término de actividad de la fórmula de la sección 5.4 pasa a estar condicionado:

```
+ 4.0 * activa   (1.0 si el estado es el activo Y no hay pregunta abierta, 0.0 si no)
```

Sin coeficiente nuevo y sin término nuevo: **siguen siendo exactamente siete**, que es como el
documento los describe. El motivo es que una tarea aparcada no la está trabajando nadie, y sin esta
condición la fórmula afirmaría lo contrario y la pondría arriba del listado del que un agente elige.

Eso obliga a tocar también `biso get --explain-urgency`, que imprime una línea por término. Con la
condición nueva, una tarea en el estado activo y con una pregunta abierta muestra `0.00` en esa línea,
y quien lo lea no puede saber si es porque el estado no es el activo o porque hay una pregunta. La
línea tiene que decir cuál de los dos motivos se aplica.

---

## 8. El arrendamiento, aplazado

Se decide su forma y se aplaza su mecanismo, y la separación es limpia porque la parte que tocaba el
modelo de estados ya no existe.

**La forma.** Arrendar es asignar con caducidad. Como la asignación ya existe, lo único que falta es un
instante de caducidad sobre una tarea activa y asignada. Un arrendamiento vencido **saca la tarea del
estado activo y le deja la asignación puesta**: lo que caduca es "estoy en ello", no "esto es tuyo".

**Lo que queda para la decisión de persistencia**: el nombre y el tipo del campo, cómo se renueva
mientras se trabaja, y quién detecta la caducidad sin que cueste caro.

**Por qué no se escribe el campo ahora.** Sería un campo que nada mantiene, que `biso export` tendría
que llevar y que la prueba de simetría tendría que cubrir, para un mecanismo que no existe.

---

## 9. Lo retirado: el requisito 9.4

Se retira, y el argumento correcto no es el que parecía.

**El argumento que no vale.** Decir que archivar y descartar son lo mismo no se sostiene contra lo que
`SPEC.md` dice hoy de `biso archive`: existe `--unarchive`, que devuelve la tarea al tablero con el
estado que tenía, y la ayuda del comando presenta el archivo como sacar tareas del tablero sin
perderlas. Archivar es reversible, así que es aparcar, y aparcar no es descartar.

**Y la deducción que proponía tampoco vale.** "Una tarea archivada que nunca llegó al estado terminal
es una tarea abandonada" no se puede calcular, porque el modelo de datos guarda el estado actual y no
un histórico de estados. Con `biso start --reopen` una tarea terminada vuelve al activo, y archivada
desde ahí la deducción la llamaría abandonada habiendo estado hecha.

**El argumento que sí vale.** Los otros tres requisitos de la sección 9 traen cada uno su evidencia
medida: llamadas contadas, bytes medidos, fallos reproducidos. Este no trae ninguna. Dice que hecha y
abandonada "se confunden", sin un solo caso en el que esa confusión haya costado algo. Un papel de
estado que obliga a dar un motivo es barato de añadir cuando haga falta y caro de quitar si sobra, así
que **se queda fuera hasta que haya un caso real que lo pida**, y entonces se engancha a `biso archive`
en vez de a un estado nuevo.

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

Los filtros que cada uno usa en el día a día:

| Quién | Qué quiere saber | Comando |
|---|---|---|
| Agente | Todo lo necesario al arrancar | `biso prime` |
| Agente | Lo que puede empezar ahora | `biso ls --mine --not-active --not-waiting` |
| Agente | Lo que dejó a medias | `biso ls --mine --active` |
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
| 2 | Sección 4.3 | Los dos avisos nuevos, sobre una lista que se declara completa |
| 3 | Sección 4.12 | `ask` y `answer` entran en la enumeración de qué es una lectura dirigida |
| 4 | Sección 5 | Nuevo campo `question` con sus tres partes, no derivado |
| 5 | Sección 5.3 | La excepción del instante escrito por `answer`, y `askedAt` como cuarta fecha importable |
| 6 | Sección 5.4 | El término de actividad exige que no haya pregunta abierta |
| 7 | Sección 7.2 | `--search` alcanza el cuerpo de la pregunta abierta |
| 8 | Sección 8 | `question` no entra en la tabla de clases de campo, y se dice por qué |

**Los comandos**

| # | Dónde | Qué |
|---|---|---|
| 9 | Sección 10, cabecera y tabla | De diecisiete comandos a diecinueve, y de nueve del ciclo a once |
| 10 | Sección 10, texto de cierre | `ask` y `answer` entran en la lista de comandos que aceptan banderas de campo |
| 11 | Sección 10.1 | Se retira la regla posicional; `--statuses` exige las tres banderas de papel |
| 12 | Sección 10.3 | Se retira `default_assignee`; `new --from` acepta `question`; la línea de ejemplo del NDJSON la incluye |
| 13 | Sección 10.4 | Filtros `--waiting`, `--not-waiting`, `--active` y `--not-active`; `question` en el JSON del listado |
| 14 | Sección 10.5 | Novena sección `question`; `--explain-urgency` distingue los dos motivos del término activo |
| 15 | Sección 10.6 | La lista de productores de `kind` `task.write` gana los dos verbos |
| 16 | Sección 10.7 | Los dos verbos nuevos, con su orden de aplicación dentro de la escritura |
| 17 | Sección 10.7.1 | `biso start` avisa sobre una tarea aparcada |
| 18 | Sección 10.7.4 | `biso finish` avisa sobre una tarea aparcada |
| 19 | Sección 10.9 | `question` en `export`, los cuatro filtros nuevos, y el guion de simetría con las banderas de papel |
| 20 | Sección 10.10 | `initial_status`; mínimo de tres estados; papeles distintos; se retira `default_assignee` |
| 21 | Sección 10.11 | `biso doctor` comprueba la invariante de los tres papeles |
| 22 | Sección 11 | La ayuda de primer nivel lista `ask` y `answer`, y su cifra de líneas |

**`biso prime` y los contratos**

| # | Dónde | Qué |
|---|---|---|
| 23 | Sección 9.5 | El reparto pasa a 3.456 y 1.664 con total 5.120; qué bloques van en cada mitad; el orden de recorte; qué número congela el contrato |
| 24 | Sección 9.6 | De ocho órdenes del ciclo a diez, y de diez reglas a once |
| 25 | Sección 9.7 | Los dos bloques nuevos, la precedencia, el reparto de `--limit`, y el rótulo de `NEXT UP` |
| 26 | Secciones 9.9, 10.1 y 10.10 | `defaultStatus` pasa a `initialStatus`; `notStartedHidden` cambia de nombre |
| 27 | Sección 12.3 | Seis identificadores `code` nuevos: dos del código 6 para los verbos y cuatro del código 2 para `biso init`. Los dos casos de `config set` reutilizan `board_inconsistent`, que ya existe |
| 28 | Sección 13 | Desde cuándo obliga el contrato de estabilidad |
| 29 | Sección 15 | De cuatro verbos de ciclo a seis |

**Barridos sobre todo el documento**

| # | Qué |
|---|---|
| 30 | El tablero de ejemplo pasa a `To Do, In Progress, Done`; desaparece el estado `Blocked` |
| 31 | Se regeneran los ejemplos de `ls` y de `prime`, que no se escriben a mano |
| 32 | Se aplica el vocabulario y se retiran las palabras prohibidas |

Los cambios 22 y 24 tocan cuatro cuentas publicadas en prosa, y el 9 y el 29 otras tres. Van en la
misma pasada, porque su patrón de fallo es el que la sección 8 de `DECISIONES.md` señala como
dominante. Hay una octava fuera de `SPEC.md`: el `CLAUDE.md` del proyecto dice "los quince comandos" y
el documento dice diecisiete, así que ya estaban descuadrados.

### 11.2. En `DECISIONES.md`

| # | Dónde | Qué |
|---|---|---|
| 1 | Sección 9 | Se sustituye entera: los cuatro requisitos pasan a resueltos, aplazado o retirado |
| 2 | Sección 9.1 | El diagnóstico correcto, la retirada de `default_assignee`, y qué cuesta eso |
| 3 | Sección 9.4 | Por qué se retira: le falta la evidencia que los otros tres sí traen |
| 4 | Sección 3 | El presupuesto nuevo, con el reparto y el motivo de no subir el total |
| 5 | Sección 4 | `default_assignee` entra en la lista de lo que se deja fuera, con su razón |
| 6 | Sección 6 | Por qué se retira la regla posicional, con la contradicción que la delató |
| 7 | Sección 6 | Por qué el tablero por defecto no trae `Ideas` |
| 8 | Nueva | El criterio de estado frente a campo, y por qué la condición no es configurable |
| 9 | Nueva | Los riesgos conocidos de la sección 12 de este documento |

---

## 12. Riesgos conocidos y aceptados

**Dos agentes con la misma identidad ven la misma cola.** Dos sesiones con `BISO_ME=@claude` no se
distinguen entre sí y cogerían la misma tarea. Es exactamente lo que resuelve el arrendamiento, que
está aplazado, así que hasta entonces el modelo no protege contra eso.

**Un tablero con `me` en su configuración anula la distinción entre persona y agente.** La clave `me`
gana sobre `BISO_ME` según la tabla de la sección 3.1, así que en un tablero que la tenga puesta todo
el mundo es la misma identidad y `--mine` deja de significar nada. Un tablero compartido entre una
persona y un agente **tiene que dejar `me` sin configurar** y depender de la variable de entorno. Hay
que escribirlo donde se explica la clave, porque hoy no está y es una trampa silenciosa.

**La persona no tiene canal hacia el agente que se vea en `biso prime`.** El agente pregunta y la
persona responde, pero si la persona quiere decirle algo por iniciativa propia lo escribe en un
comentario, y `prime` no muestra comentarios. El agente lo ve al hacer `biso get`, así que el documento
tiene que decir que ese es el camino.

**Una pregunta abierta sobre una tarea terminada o archivada desaparece de la vista.** `finish` avisa
pero no impide, los bloques de `prime` excluyen terminadas y archivadas, y `biso ls` excluye el estado
terminal por defecto. Así que `biso ls --waiting` no la encuentra sin `--any-status`. Se acepta porque
la alternativa, impedir cerrar una tarea con una pregunta abierta, empuja a rodear la herramienta, que
es el fallo que el apartado 10.7 de `SPEC.md` evita a propósito en el caso equivalente de los criterios
sin marcar.

**`--ready` no excluye las aparcadas.** `ready` mira solo dependencias, así que una tarea parada en una
pregunta sigue siendo `ready` y un agente que elija trabajo con esa bandera, que es lo que su nombre
invita a hacer, se las lleva. La consulta correcta es la del apartado 10, con `--not-waiting`. Está
anotado porque el nombre engaña.

**`ready` y `blocked` siguen siendo complementarios**, dos campos derivados que viajan en el JSON y en
la ficha para contar un solo hecho. Es anterior a este trabajo y queda como limpieza aparte.
