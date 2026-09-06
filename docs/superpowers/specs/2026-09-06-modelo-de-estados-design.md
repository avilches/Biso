# Diseño: el modelo de estados de `biso`

Este documento cierra los cuatro requisitos que la sección 9 de [`DECISIONES.md`](../../DECISIONES.md)
dejó identificados y sin incorporar. No es la especificación: es la decisión y su razón, escrita para
que aplicarla a [`SPEC.md`](../../SPEC.md) y a `DECISIONES.md` sea mecánico. La sección 10 de este
documento es la lista de esos cambios, uno a uno.

Lo que decide, en una frase: **ningún papel de estado nuevo, un campo nuevo con dos verbos, y el
encargo resuelto con la asignación que ya existe.**

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
| 9.4, distinguir terminar de descartar | **Retirado** | Archivar y descartar son lo mismo |

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

Es la regla cuya ausencia produjo la contradicción que se documenta en 3.3.

### 3.2. `default_status` pasa a llamarse `initial_status`

`default` no dice nada, porque todo tiene un valor por defecto. `initial` dice exactamente lo que es,
el estado donde nace una tarea. Se descarta `init_status` porque `init` es un comando del programa y la
clave se leería como "el estado que crea `biso init`", que no es lo que significa.

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

Los cuatro casos límite de `biso init`, todos con error 2 porque son problemas de argumentos:

| Caso | Mensaje |
|---|---|
| Falta alguna de las tres banderas de papel | Las tres nombradas, diciendo cuáles faltan |
| Una bandera de papel nombra un estado que no está en `--statuses` | El valor y la lista de estados |
| Dos banderas de papel nombran el mismo estado | Los dos papeles y el estado que comparten |
| `--statuses` con menos de tres estados | Cuántos hacen falta y por qué |

Y los dos que aparecen en `biso config set`, los dos con error 3 porque son problemas de valor:
dejar `statuses` con menos de tres estados, y dar a un papel un estado que ya tiene otro papel.

### 3.4. El mínimo pasa de dos estados a tres

Los tres papeles son obligatorios y **distintos entre sí**, así que un tablero necesita al menos tres
estados. Con eso desaparece el caso de los dos estados y su aviso de que `biso start` no cambia el
estado.

Un tablero puede tener más de tres, y los que sobran no tienen papel. `Ideas`, `Review` o `Blocked`
siguen siendo estados perfectamente válidos; lo que ya no pueden es ser un papel del modelo.

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

### 4.2. Cómo queda el flujo

Un agente que arranca pregunta por lo suyo con lo que ya existe:

```
biso ls --mine --not-status "In Progress" --not-waiting   # lo que puede empezar
biso ls --mine -s "In Progress"                           # lo que dejo a medias
```

Y una persona encarga con la misma llamada que crea la tarea:

```
biso new "Normalizar CRLF en el diff" --type bug -a @claude \
    -d "El parser normaliza LF pero no CRLF" \
    --ac "Un diff con CRLF da el mismo resultado que uno con LF" \
    --comment "Mira el parser, no el formateador"
```

### 4.3. Lo que esta decisión cuesta, dicho a propósito

**El caso de uso que motivó todo el trabajo deja de ser un arrastre.** El requisito 9.1 lo describe
como "arrastrar una tarjeta desde el móvil para que un agente se ponga con esa tarea". Con la
asignación como encargo, el gesto pasa a ser asignar un miembro, que en un tablero de móvil son dos
toques en vez de un arrastre y sigue siendo perfectamente viable, pero la frase del requisito ya no
describe la herramienta y hay que reescribirla.

La consecuencia mayor llega después: **cuando se especifique la sincronización con un sistema externo,
mover una tarjeta de columna en ese sistema no significará nada para `biso`**. Eso hay que decidirlo
entonces a propósito, y queda anotado aquí para que no se descubra por sorpresa.

**Una tarea que un agente cogió por su cuenta acaba pareciéndose a una que le encargó una persona**,
porque `biso start` asigna a `me` cuando la tarea no tiene a nadie. Operativamente da igual, porque en
los dos casos la conclusión es la misma, pero conviene que esté escrito.

---

## 5. La pregunta abierta

### 5.1. El campo

Un campo nuevo de la tarea, `question`, con **las mismas tres partes que un comentario** de la sección
5.2 de `SPEC.md`:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | texto libre | quien llama, y por defecto la identidad `me` |
| `askedAt` | instante UTC | el programa, salvo al importar |
| `body` | texto largo | quien llama |

Vacío es lo normal. Con contenido significa que la tarea espera a una persona, esté en el estado que
esté. Lleva tres partes y no una porque al responderse **se convierte literalmente en un comentario**,
y para eso hacen falta su autor y su instante originales.

No se añade ningún booleano derivado: que el campo tenga contenido ya lo dice todo, y añadir uno
repetiría la redundancia que ya existe entre `ready` y `blocked`.

### 5.2. Los dos verbos

Son verbos y no solo banderas por el principio 5, que dice que un gesto del flujo de trabajo es un
comando. Sin ellos, responder cuesta dos llamadas, una para el comentario y otra para vaciar el campo.

**`biso ask <ref> [<text>...] [--comment-author <@who>]`** llena el campo.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Se llena el campo. **No cambia el estado** |
| La tarea ya tiene una pregunta abierta | Error 6, para que la segunda no borre a la primera en silencio |
| Sin `--comment-author` y sin identidad configurada | Error 2, el mismo de `biso comment` |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que `biso note` |
| Varias referencias | No se admiten: toma exactamente una, como `note` y `comment` |

**`biso answer <ref> [<text>...] [--comment-author <@who>]`** vacía el campo, en una sola escritura y
con tres efectos:

1. Añade al histórico un comentario con el `author`, el `askedAt` y el `body` que guardaba el campo.
2. Añade detrás un segundo comentario con la respuesta, firmado por quien contesta y con el instante
   de ahora.
3. Vacía el campo.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Error 6, con la pista de usar `biso comment` |
| Sin `--comment-author` y sin identidad configurada | Error 2, el mismo de `biso comment` |
| Se pasan además banderas de campo | Se aplican igual, como en cualquier verbo del ciclo |

Los dos aceptan todas las banderas de campo de la sección 8, como el resto de verbos del ciclo. El caso
de uso real de eso es responder y concretar a la vez:

```
biso answer TASK-11 "Solo los de texto. Los binarios se saltan enteros." \
    --ac "Un fichero binario no se toca"
```

**Códigos de salida de los dos**: los mismos de `biso comment` (0, 2, 3, 4, 5, 7, 8, 9), más el 6 de
las dos filas de arriba.

### 5.3. Una excepción declarada a la sección 5.3 de `SPEC.md`

Esa sección dice que el instante de un comentario lo pone el programa y **solo se puede fijar al
importar**. `biso answer` escribe un comentario con un instante pasado. No es una excepción de fondo,
porque ese instante lo observó el propio programa cuando se hizo la pregunta y quien llama no lo
negocia, pero es una excepción de forma y tiene que estar escrita donde está la regla.

### 5.4. Alrededor del campo

- **`--question <texto>` y `--clear-question`**, las dos variantes que la sección 8.5 da a cualquier
  escalar, aceptadas por todos los comandos que escriben.
- **`--waiting` y `--not-waiting`** en `biso ls` y en `biso export`. Son dos y no una porque el agente
  necesita **excluir** las aparcadas para elegir trabajo, y ese es su filtro más usado.
- **`question` viaja en `biso export` y lo acepta `biso new --from`**, con sus tres partes, porque no
  es un campo derivado y la ida y vuelta es una prueba de la suite.
- **`biso get` gana una novena sección**, `question`, que se suma a las ocho de `--section`.
- **`biso start` sobre una tarea aparcada avisa y no lo impide**, igual que hace hoy con las
  dependencias sin terminar. Impedirlo empujaría a rodear la herramienta con `set`.
- **`biso finish` sobre una tarea aparcada avisa y no lo impide**, por el mismo motivo que no lo impide
  con los criterios sin marcar.

---

## 6. `biso prime`

### 6.1. Dos bloques nuevos

Entre `IN PROGRESS` y `NEXT UP` entran dos bloques, los dos sin límite de filas por el mismo argumento
que ya se aplica a `IN PROGRESS`: en un tablero sano son pocas.

- **`WAITING ON A PERSON`**: las tareas con pregunta abierta. Cada fila lleva, además de las columnas
  de siempre, la pregunta recortada a una línea por el mismo algoritmo de columnas que recorta los
  títulos, porque la pregunta es el dato que hace falta para actuar.
- **`ASSIGNED TO YOU`**: las asignadas a la identidad configurada que no estén en el estado activo.

**Ninguna tarea aparece en dos bloques.** Los cuatro se reparten el tablero por esta precedencia, y
cada tarea cae en el primero que la acepte:

1. `WAITING ON A PERSON`, si tiene una pregunta abierta.
2. `IN PROGRESS`, si está en el estado activo.
3. `ASSIGNED TO YOU`, si está asignada a la identidad configurada.
4. `NEXT UP`, el resto.

De esa precedencia salen dos consecuencias que conviene ver escritas. **Una tarea aparcada no sale en
`IN PROGRESS` aunque esté en el estado activo**, porque ese bloque significa que alguien está
trabajando y ahí no lo está nadie. Y **`NEXT UP` deja de incluir lo que ya salió en los dos bloques
nuevos**, así que su línea de recuento final cuenta solo lo que él mismo recorta.

Los cuatro bloques excluyen las tareas terminadas y las archivadas, igual que hoy.

### 6.2. El presupuesto sube, con el número medido

El añadido a la parte fija son 55 bytes de las dos órdenes nuevas, 12 de la bandera de campo y 214 de
la regla nueva, en total **281**. La parte fija pasa de 2.963 a **3.244** contra un tope de 3.072, y el
mensaje entero a unos **4.731** contra un tope total de 5.120.

Es decir, **el tope total no se rompe, solo el reparto interno**. El reparto pasa a ser:

| Mitad | Antes | Ahora |
|---|---:|---:|
| Parte fija | 3.072 | **3.456** |
| Resumen del tablero | 2.048 | 2.048 |
| **Total** | **5.120** | **5.504** |

Los 212 bytes de holgura de la parte fija son deliberados: quedarse otra vez a ochenta bytes del techo
garantiza volver a esta discusión al siguiente cambio. El argumento del apartado 3 de `DECISIONES.md`
no se resiente, porque 4.731 bytes medidos siguen siendo poco más de un tercio de los 12.905 que
costaba un ciclo completo en la herramienta estudiada.

---

## 7. La urgencia

El término de actividad de la fórmula de la sección 5.4 pasa a estar condicionado:

```
+ 4.0 * activa   (1.0 si el estado es el activo Y no hay pregunta abierta, 0.0 si no)
```

Sin coeficiente nuevo y sin término nuevo: **siguen siendo exactamente siete**, que es como el
documento los describe. El motivo es que una tarea aparcada no la está trabajando nadie, y sin esta
condición la fórmula afirmaría lo contrario y la pondría arriba del listado del que un agente elige.

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

**Archivar y descartar son lo mismo.** Si cuesta explicar la diferencia entre los dos, es que no la
hay, y una herramienta con dos gestos indistinguibles enseña a dudar de cuál usar.

La distinción que el requisito pedía sigue siendo legible sin guardar nada: **una tarea archivada que
nunca llegó al estado terminal es exactamente una tarea abandonada**, y eso se deduce de dos campos que
ya existen.

Lo único que se pierde es la obligación de dar un motivo al abandonar. Si algún día hace falta, se
engancha a `biso archive` y no a un estado nuevo.

---

## 10. La lista de cambios

### 10.1. En `SPEC.md`

| # | Dónde | Qué |
|---|---|---|
| 1 | Antes de la sección 1 | Entra la tabla de vocabulario de la sección 2 de este documento |
| 2 | Sección 5 | Nuevo campo `question` con sus tres partes, no derivado |
| 3 | Sección 5.3 | La excepción del instante de un comentario escrito por `biso answer` |
| 4 | Sección 5.4 | El término de actividad pasa a exigir que no haya pregunta abierta |
| 5 | Secciones 6 y 10.10 | `default_status` pasa a `initial_status`; mínimo de tres estados; los tres papeles distintos |
| 6 | Sección 8.5 | `--question` y `--clear-question` |
| 7 | Sección 9.5 | El reparto del presupuesto pasa a 3.456 y 2.048, con total 5.504 |
| 8 | Sección 9.6 | De ocho órdenes del ciclo a diez, y de diez reglas a once |
| 9 | Secciones 9.7 y 9.9 | Los bloques `WAITING ON A PERSON` y `ASSIGNED TO YOU`, en el texto y en el JSON |
| 10 | Sección 10.1 | Se retira la regla posicional; `--statuses` exige las tres banderas de papel |
| 11 | Sección 10.4 | Filtros `--waiting` y `--not-waiting` |
| 12 | Sección 10.5 | Novena sección `question` en la ficha y en `--section` |
| 13 | Sección 10.7 | Los dos verbos nuevos, `ask` y `answer` |
| 14 | Sección 10.7.1 | `biso start` avisa sobre una tarea aparcada |
| 15 | Sección 10.7.4 | `biso finish` avisa sobre una tarea aparcada |
| 16 | Sección 10.9 | `question` entra en `export`, y `--waiting` y `--not-waiting` entre sus filtros |
| 17 | Sección 10.3 | `biso new --from` acepta `question` con sus tres partes |
| 18 | Sección 11 | La ayuda de primer nivel lista `ask` y `answer` |
| 19 | Sección 12.3 | Los identificadores de error de las dos condiciones nuevas del código 6 |
| 20 | Todo el documento | El tablero de ejemplo pasa a `To Do, In Progress, Done`; desaparece el estado `Blocked` |
| 21 | Todo el documento | Se regeneran los ejemplos de `ls` y de `prime`, que no se escriben a mano |
| 22 | Todo el documento | Se aplica el vocabulario y se retiran las palabras prohibidas |

Los cambios 8 y 12 tocan tres cuentas publicadas en prosa que cambian a la vez: las ocho órdenes del
ciclo y las diez reglas de `biso prime`, y las ocho secciones fijas de `biso get`. Van en la misma
pasada, porque su patrón de fallo es el que la sección 8 de `DECISIONES.md` señala como dominante.

### 10.2. En `DECISIONES.md`

| # | Dónde | Qué |
|---|---|---|
| 1 | Sección 9 | Se sustituye entera: los cuatro requisitos pasan a resueltos, aplazado o retirado |
| 2 | Sección 9.1 | El diagnóstico correcto: el encargo es la asignación, y qué cuesta eso |
| 3 | Sección 9.4 | Por qué se retira: archivar y descartar son lo mismo |
| 4 | Sección 3 | El presupuesto nuevo, con los bytes medidos |
| 5 | Sección 6 | Por qué se retira la regla posicional, con la contradicción que la delató |
| 6 | Sección 6 | Por qué el tablero por defecto no trae `Ideas` |
| 7 | Nueva | El criterio de estado frente a campo, y por qué la condición no es configurable |
| 8 | Nueva | Los riesgos conocidos de la sección 11 de este documento |

---

## 11. Riesgos conocidos y aceptados

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

**`biso set --clear-question` tira la pregunta sin responderla.** Cumple el principio 3, porque el
nombre dice lo que hace, pero la pierde del histórico en vez de moverla a los comentarios.

**`ready` y `blocked` siguen siendo complementarios**, dos campos derivados que viajan en el JSON y en
la ficha para contar un solo hecho. Es anterior a este trabajo y queda anotado como limpieza aparte.
