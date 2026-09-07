# Plan de aplicación del modelo de estados

> **Para quien ejecute esto de forma automática:** SUB-SKILL OBLIGATORIA. Usa
> `superpowers:subagent-driven-development` (recomendada) o `superpowers:executing-plans` para
> ejecutar tarea a tarea. Los pasos usan casillas (`- [ ]`) para llevar la cuenta.

**Objetivo:** aplicar a `docs/SPEC.md` y a `docs/DECISIONES.md` el diseño del modelo de estados, de
modo que los dos documentos queden sin una sola contradicción entre sí ni consigo mismos.

**Enfoque:** no hay código. Lo que se edita son dos documentos de especificación, así que el ciclo de
prueba no es una suite sino **una comprobación de texto que hoy falla y que después de la edición
tiene que pasar**: un `grep` que debe dar cero, un recuento que debe cuadrar, o una medida de bytes
que debe caber en su tope. Cada tarea trae la suya, y ninguna se da por buena sin ejecutarla.

**Herramientas:** `grep`, `wc -c`, `git`. Nada más.

**Diseño del que sale este plan:** [`../specs/2026-09-06-modelo-de-estados-design.md`](../specs/2026-09-06-modelo-de-estados-design.md).
**Hay que leerlo entero antes de empezar.** Este plan dice dónde tocar y cómo comprobarlo; el diseño
dice qué escribir y por qué, y su sección 11 numera los treinta y ocho cambios de `SPEC.md` y los nueve
de fuera. Cada tarea de aquí cita los números de esa lista que cubre.

## Restricciones globales

Salen del `CLAUDE.md` del proyecto y de `DECISIONES.md`, y valen para todas las tareas:

- **La prosa va en español. Todo lo que es interfaz del programa va en inglés**: nombres de comando,
  banderas, textos de ayuda, mensajes de error, claves JSON y claves de configuración.
- **Nunca un em-dash (`—`)**, en ningún texto: ni en los documentos, ni en los mensajes de commit.
- **Los mensajes de commit no llevan coautoría** ni mención de haber sido generados por un agente.
- **Cuando un dato tenga que aparecer en dos sitios, uno remite al otro** en vez de repetirlo.
- **Los ejemplos de salida se generan ejecutando el algoritmo, nunca se escriben a mano.**
- **Al corregir una afirmación, hay que buscarla en todo el documento** antes de darla por corregida.
- **El trabajo va en el worktree** `.claude/worktrees/modelo-de-estados`, nunca en `main`.
- **Ninguna implementación se ramifica por el nombre de un estado.** Si al escribir aparece la
  tentación de citar `In Progress` en una regla, la regla está mal: tiene que hablar del papel.

## Punto de partida medido

Estas son las cuentas de `docs/SPEC.md` antes de tocar nada. Sirven para saber que una comprobación
empieza fallando, que es lo que la hace valer de algo:

| Patrón | Apariciones hoy |
|---|---:|
| `default_status` | 9 |
| `defaultStatus` | 2 |
| `default_assignee` | 5 |
| `Blocked` | 16 |
| `Ideas` | 19 |
| `notStartedHidden` | 1 |
| `penúltimo` | 2 |
| `dos estados` | 2 |

---

## Tarea 1: el vocabulario

Cubre los cambios 1 y 38 de la sección 11.1 del diseño.

**Ficheros:**
- Modificar: `docs/SPEC.md`, entre el párrafo "Convención de idioma" y la línea `## 1. Los principios`

**Lo que produce para las tareas siguientes:** los términos que todas ellas tienen que usar. Se hace
primero justamente por eso.

- [ ] **Paso 1: comprobar que hoy no existe**

```bash
grep -c '^## Vocabulario de este documento' docs/SPEC.md
```

Esperado: `0`.

- [ ] **Paso 2: escribir la sección**

Copiar la tabla de la sección 2 del diseño, entera y sin cambios, bajo el encabezado
`## Vocabulario de este documento`, sin número, seguida de la lista de las cuatro palabras
restringidas. Va detrás del párrafo de convención de idioma y delante del separador `---` que precede
a `## 1. Los principios`, para no renumerar las quince secciones.

- [ ] **Paso 3: comprobar que existe y que no rompe la numeración**

```bash
grep -c '^## Vocabulario de este documento' docs/SPEC.md
grep -n '^## [0-9]' docs/SPEC.md | head -3
```

Esperado: `1`, y que la primera sección numerada siga siendo `## 1. Los principios`.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Añade el vocabulario del proyecto a la especificación"
```

---

## Tarea 2: renombrar `default_status` a `initial_status`

Cubre los cambios 21 y 30, y la parte de renombrado del 28.

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 9.9, 10.1, 10.10 y 10.11

**Lo que consume:** nada. **Lo que produce:** el nombre `initial_status` y su forma JSON
`initialStatus`, que las tareas 4, 10 y 12 dan por hecho.

- [ ] **Paso 1: ver dónde está, para no dejarse ninguna**

```bash
grep -n 'default_status\|defaultStatus' docs/SPEC.md
```

Esperado: 11 líneas, repartidas entre 9.9, 10.1, 10.10 y 10.11.

- [ ] **Paso 2: renombrar las once**

`default_status` pasa a `initial_status` en la prosa, en las tablas de configuración y en los textos de
ayuda. `defaultStatus` pasa a `initialStatus` en los dos esquemas JSON. **En el texto de ayuda de
`biso config --help` hay que revisar además la alineación de la columna de descripción**, porque el
nombre crece en dos caracteres.

- [ ] **Paso 3: comprobar que no queda ninguno**

```bash
grep -c 'default_status\|defaultStatus' docs/SPEC.md
grep -c 'initial_status' docs/SPEC.md
```

Esperado: `0` en la primera, y al menos `9` en la segunda.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Renombra default_status a initial_status"
```

---

## Tarea 3: retirar `default_assignee`

Cubre los cambios 13 (la parte de retirada) y 21 (la parte de retirada).

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 10.3 y 10.10

- [ ] **Paso 1: ver las cinco apariciones**

```bash
grep -n 'default_assignee\|defaultAssignee' docs/SPEC.md
```

Esperado: 5 líneas, en la tabla de claves de configuración, en el comportamiento de `biso new`, en el
esquema JSON de `config list` y en `biso config --help`.

- [ ] **Paso 2: quitarlas**

Se retira la clave entera. Ojo con la viñeta de `biso new` que hoy dice que con `--start` y sin `-a`
manda `--start`: esa frase existe solo para resolver el conflicto entre las dos, así que se va con
ella, y lo que queda es que `--start` asigna `me`.

- [ ] **Paso 3: comprobar**

```bash
grep -c 'default_assignee\|defaultAssignee' docs/SPEC.md
```

Esperado: `0`.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Retira default_assignee, que vaciaba de significado a la asignación"
```

---

## Tarea 4: el mínimo de tres estados y los tres papeles distintos

Cubre los cambios 12, 21 y 22, y la parte de la tarea que quita el aviso de los dos estados del
cambio 2.

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 4.3, 10.1, 10.10 y 10.11

**Lo que consume:** el nombre `initial_status` de la tarea 2.

- [ ] **Paso 1: localizar la regla rota**

```bash
grep -n 'penúltimo\|dos estados\|Menos de dos' docs/SPEC.md
```

Esperado: cinco líneas. Las dos de `penúltimo` son la regla posicional, en 10.1 y en su ayuda.

- [ ] **Paso 2: retirar la regla posicional y escribir las banderas**

En 10.1: se van las tres filas que asignan papeles por posición, la fila de "exactamente dos estados" y
la de "menos de dos estados". Entran `--initial-status`, `--active-status` y `--terminal-status` en la
firma, en la tabla de parámetros y en la ayuda, y la tabla de los cinco casos límite del apartado 3.3
del diseño, todos con error 2.

En 10.10: el mínimo de `statuses` pasa a tres, y entran las dos comprobaciones nuevas con error 6 y el
identificador `board_inconsistent`, que ya existe.

En 10.11: `biso doctor` gana la comprobación de que `statuses` tiene al menos tres elementos y de que
los tres papeles apuntan a estados distintos y existentes.

En 4.3 y en la salida de 10.1: se quita el aviso `warning: with two statuses, "start" cannot change the
status`, que ya no puede darse.

- [ ] **Paso 3: comprobar**

```bash
grep -c 'penúltimo' docs/SPEC.md
grep -c 'with two statuses' docs/SPEC.md
grep -c 'initial-status' docs/SPEC.md
```

Esperado: `0`, `0`, y al menos `3`.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Retira la regla posicional de init y exige tres estados con papeles distintos"
```

---

## Tarea 5: el campo `question` en el modelo de datos

Cubre los cambios 4, 5, 6, 7, 8 y 9.

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 5, 5.2, 5.3, 5.4, 7.2 y 8

**Lo que produce:** el campo `question` con sus partes `author`, `askedAt` y `body`, y el derivado
`waiting`. Las tareas 6, 7, 8 y 10 los usan con esos nombres exactos.

- [ ] **Paso 1: comprobar que no existen**

```bash
grep -c 'question' docs/SPEC.md
```

Esperado: `0`.

- [ ] **Paso 2: escribir los seis cambios**

En la tabla de campos de la sección 5: `question` como campo no derivado, y `waiting` en la fila de los
derivados, junto a `ready` y `blocked`. En 5.2: que los comentarios se guardan y se muestran en orden
de inserción, y que el instante de cada uno dice la verdad aunque la lista no esté ordenada por él. En
5.3: la excepción del instante que escribe `biso answer`, y `askedAt` como cuarta fecha importable,
opcional, que si falta toma el instante de la importación. En 5.4: el término de actividad pasa a
`(1.0 si el estado es el activo y no hay pregunta abierta, 0.0 si no)`. En 7.2: el ámbito único de
búsqueda alcanza el cuerpo de la pregunta abierta, con la consecuencia dicha de que la resolución de
una referencia por texto también, y de que una pregunta puede crear una ambigüedad de código 5. En 8:
la frase de que `question` no entra en la tabla de clases de campo y por qué.

- [ ] **Paso 3: comprobar que el campo está y que sigue siendo no derivado**

```bash
grep -n 'question' docs/SPEC.md | head -20
grep -n 'urgency`, `acDone' docs/SPEC.md
```

Esperado: `question` presente en la sección 5, y **ausente** de la lista de campos derivados, donde sí
tiene que estar `waiting`.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Añade el campo question y el derivado waiting al modelo de datos"
```

---

## Tarea 6: los verbos `ask` y `answer`

Cubre los cambios 17, 18 y 19.

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 10.7, añadiendo 10.7.5 y 10.7.6 detrás de `finish`

**Lo que consume:** el campo `question` de la tarea 5.

- [ ] **Paso 1: comprobar que no existen**

```bash
grep -c 'biso ask\|biso answer' docs/SPEC.md
```

Esperado: `0`.

- [ ] **Paso 2: escribir los dos apartados**

Con la estructura que usan los otros cuatro verbos: firma, tabla de parámetros propios, qué hace, las
tablas de casos del apartado 5.2 del diseño, salida, códigos de salida y ayuda. **Ninguno de los dos
acepta `--comment-author`**, y la ayuda de `answer` lo dice, porque la sección 10.6 declara esa bandera
como "requiere `--comment`".

Este es el texto literal de las dos ayudas:

```
Usage: biso ask <ref> <text>... [options]

Park ONE task on a question for a person. The task keeps its status, but it
leaves the IN PROGRESS block of `biso prime` and shows up under WAITING ON A
PERSON until somebody runs `biso answer`.

Arguments:
  ref                one task: an id, a bare number or free text
  text               the question; @file and - work too

Options:
      --id / --match force <ref> to be an id, or free text
  -h, --help         show this help

Every field flag of `biso set --help` works here too, but nothing except this
command and `biso answer` ever writes the question itself.

A task holds one open question at a time. Answer it before asking another.

Exit codes:
  0  asked          4  not found        7  could not be written
  2  bad usage      5  ambiguous        8  no board here
  3  empty question, or unreadable      9  --dry-run did not pass
  6  already asking, or already finished

Examples:
  biso ask TASK-11 "Do we normalize binary files too, or only text?"
  biso ask 11 @/tmp/question.md
```

```
Usage: biso answer <ref> <text>... [options]

Answer the open question of ONE task and unpark it. In a single write this
moves the question into the comments with its original author and time, adds
your answer behind it, and clears the question.

Arguments:
  ref                one task: an id, a bare number or free text
  text               the answer; @file and - work too

Options:
      --id / --match force <ref> to be an id, or free text
  -h, --help         show this help

Every field flag of `biso set --help` works here too, so you can answer and
refine in one call.

Both comments are signed with your configured identity. This command does not
take --comment-author.

Exit codes:
  0  answered       4  not found        7  could not be written
  2  bad usage      5  ambiguous        8  no board here
  3  unreadable     6  no open question 9  --dry-run did not pass

Examples:
  biso answer TASK-11 "Only text files. Binary ones are skipped entirely."
  biso answer 11 "Yes" --ac "A binary file is never touched"
```

- [ ] **Paso 3: los dos avisos de `start` y `finish`**

Son los cambios 18 y 19, y van en esta tarea porque hablan del campo que esta tarea introduce. En
10.7.1, la tabla de casos de `biso start` gana una fila: sobre una tarea con pregunta abierta se
empieza igual, con aviso, exactamente como hace hoy con las dependencias sin terminar. En 10.7.4, la de
`biso finish` gana otra: se cierra igual, con aviso, como hace hoy con los criterios sin marcar.
**Ninguno de los dos lo impide**, porque impedirlo empuja a rodear la herramienta con `biso set`.

Los textos de los dos avisos van también a la tabla de la sección 4.3 en la tarea 11, que es la que
lleva esa lista.

- [ ] **Paso 4: comprobar que las dos ayudas están, que ninguna nombra `--comment-author` como opción, y que los dos avisos existen**

```bash
grep -c 'Usage: biso ask\|Usage: biso answer' docs/SPEC.md
awk '/Usage: biso ask/,/^```$/' docs/SPEC.md | grep -c 'comment-author'
grep -c 'open question' docs/SPEC.md
```

Esperado: `2`, `0`, y al menos `4`, que son las dos ayudas y las dos filas nuevas de `start` y
`finish`.

- [ ] **Paso 5: commit**

```bash
git add docs/SPEC.md
git commit -m "Especifica los verbos ask y answer"
```

---

## Tarea 7: los cuatro filtros nuevos

Cubre los cambios 14 y 20.

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 10.4 y 10.9

**Lo que consume:** el derivado `waiting` de la tarea 5.

- [ ] **Paso 1: comprobar que no existen**

```bash
grep -c '\-\-not-active\|\-\-not-waiting' docs/SPEC.md
```

Esperado: `0`.

- [ ] **Paso 2: escribirlos**

En la firma de `biso ls`, en su tabla de parámetros con su columna "Incompatible con", en las reglas de
combinación, en el JSON del listado y en la ayuda. `--waiting` es incompatible con `--not-waiting`, y
`--active` con `--not-active`. **Son compatibles con `-s`, con `--not-status` y con `--any-status`**, y
la regla general que hay que escribir en las reglas de combinación es esta:

> Dos filtros que se contradicen por construcción son incompatibles. Una combinación de filtros
> válidos que resulte vacía en este tablero es un hecho legítimo sobre el tablero, no un error.

En el JSON del listado entra `waiting`, junto a `ready` y `blocked`, y **no entra `question`**, porque
su cuerpo es texto largo y el listado no lleva texto largo nunca.

En 10.9 los cuatro filtros se aceptan igual que el resto de filtros de `ls`, y `question` sí entra en
lo que escribe `export`, con sus tres partes.

**Y hay que arreglar el guion de la garantía de simetría de esa misma sección**, que hoy invoca
`biso init Kex --statuses "..."` sin banderas de papel. Desde la tarea 4, esa llamada es un error 2,
así que el documento estaría publicando como prueba un guion que no se puede ejecutar. Hay que
añadirle `--initial-status`, `--active-status` y `--terminal-status` con los estados que corresponda.
No cambies ahí la lista de estados: eso es la tarea 10.

Y en 10.3, que es el otro extremo de la simetría, el formato de lote de `biso new --from` acepta
`question` como objeto con `author`, `askedAt` y `body`, con `askedAt` opcional. **La línea de ejemplo
del NDJSON, que el documento presenta como la que lleva todos los tipos compuestos, tiene que
incluirlo**, o deja de cumplir lo que promete.

- [ ] **Paso 3: comprobar la simetría de export**

```bash
grep -n 'question' docs/SPEC.md | grep -i 'export\|from'
awk '/"blocked"/{print NR": "$0}' docs/SPEC.md
```

Esperado: `question` presente en 10.9 y en el formato de lote de 10.3, y `waiting` junto a `blocked` en
el JSON de `ls`.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Añade los filtros waiting y active, que hablan por papel y no por nombre"
```

---

## Tarea 8: `biso get`

Cubre el cambio 15.

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 10.5

- [ ] **Paso 1: contar las apariciones del ocho, que son cuatro**

```bash
grep -n 'ocho secciones\|meta`, `desc\|the eight\|eight sections' docs/SPEC.md
```

Esperado: las cuatro apariciones dentro de 10.5, en la tabla de parámetros, en el mensaje de error de
`--section` inventada, en la frase "las ocho secciones fijas" y en la ayuda.

- [ ] **Paso 2: los cuatro cambios de esta sección**

`question` entra como novena sección de `--section` en los cuatro sitios donde hoy pone ocho. La ficha
la imprime como las demás, con `(empty)` cuando no hay pregunta. `--explain-urgency` cambia en sus dos
formas: la línea de texto del término activo dice cuál de los dos motivos lo anula, y
`urgencyBreakdown.active` del JSON deja de ser un número suelto para llevar el motivo con él. Y el
ámbito de búsqueda de texto que la ayuda de este comando repite tiene que decir lo mismo que 7.2.

- [ ] **Paso 3: comprobar que no queda ningún ocho**

```bash
grep -c 'ocho secciones\|eight sections' docs/SPEC.md
grep -c 'nueve secciones\|nine sections' docs/SPEC.md
```

Esperado: `0` en la primera y al menos `1` en la segunda.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Añade la sección question a biso get y desambigua explain-urgency"
```

---

## Tarea 9: `biso prime`, el texto y las reglas

Cubre los cambios 24, 25, 26, 27, 28 y 29. **Es la tarea más grande y la que más se puede equivocar**,
porque toca seis apartados que se citan entre sí.

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 9.3, 9.5, 9.6, 9.7, 9.9 y 9.11

**Lo que consume:** el campo `question` de la tarea 5 y los verbos de la tarea 6.

- [ ] **Paso 1: comprobar el estado de partida**

```bash
grep -c 'notStartedHidden' docs/SPEC.md
grep -n 'ocho órdenes\|diez reglas\|3.072\|2.048\|5.120' docs/SPEC.md
```

Esperado: `1` en la primera, y en la segunda las cifras del presupuesto y las dos cuentas de 9.6.

- [ ] **Paso 2: los seis apartados**

En 9.3: `--limit` acota `ASSIGNED TO YOU` y `NEXT UP` **juntos**, cinco filas repartidas entre los dos
en ese orden de preferencia, y con `0` desaparecen los dos y queda solo la línea de recuento.

En 9.5: el reparto pasa a 3.456 y 1.664 con total 5.120; la enumeración de qué bloques van en cada
mitad incluye los dos nuevos; la regla de a qué mitad se imputa cada línea en blanco los nombra; el
orden de recorte se completa con los cinco pasos del apartado 6.1 del diseño; y entra la frase de cuál
de los dos números congela el contrato de estabilidad, que es el total.

En 9.6: de ocho órdenes del ciclo a diez, y de diez reglas a once.

En 9.7: los dos bloques con la precedencia de cuatro pasos, los tres renombrados de la tabla del
apartado 6.1 del diseño, la regla de que un bloque sin filas no se imprime, la de que sin identidad
`ASSIGNED TO YOU` no se imprime nunca, y la de que los anchos de las columnas 1 a 7 se calculan sobre
las filas de los cuatro bloques juntas, sin contar las líneas de pregunta. La pregunta va en línea
propia indentada, recortada a **100 caracteres**, la misma cifra que los títulos.

Esta es la regla 11 literal que entra en el bloque `RULES`:

```
 11. `biso ask <ref> "..."` parks a task on a question and `biso answer` unparks
     it, writing both into the comments. Ask instead of guessing. A task
     assigned to you is one a person decided you should do.
```

Y estas las dos líneas que entran en `COMMANDS`:

```
  biso ask <ref> "QUESTION"
  biso answer <ref> "TEXT"
```

En 9.9: las dos listas nuevas del esquema JSON, `defaultStatus` ya renombrada por la tarea 2, y
`notStartedHidden` a `hiddenCount`.

En 9.11: la ayuda describe el `--limit` nuevo.

**No se toca todavía el ejemplo literal de salida de 9.7.** Eso es la tarea 10, porque se genera y no
se escribe.

- [ ] **Paso 3: comprobar**

```bash
grep -c 'notStartedHidden' docs/SPEC.md
grep -c 'hiddenCount' docs/SPEC.md
grep -c '5.504' docs/SPEC.md
grep -n '3.456\|1.664\|5.120' docs/SPEC.md
```

Esperado: `0`, al menos `1`, `0` (el tope no sube), y las tres cifras del reparto presentes.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Añade a prime los bloques de preguntas y de asignadas, con su presupuesto y su recorte"
```

---

## Tarea 10: regenerar los ejemplos y medir

Cubre los cambios 36 y 37. **Es la única tarea que puede invalidar a las anteriores**, porque es la que
comprueba si el presupuesto cuadra de verdad.

**Ficheros:**
- Modificar: `docs/SPEC.md`, los ejemplos de salida de las secciones 9.7, 9.8, 10.4 y 10.5, y todas las
  apariciones del tablero de ejemplo

- [ ] **Paso 1: cambiar el tablero de ejemplo**

`Ideas, To Do, In Progress, Blocked, Done` pasa a `To Do, In Progress, Done` en las 35 apariciones de
`Ideas` y `Blocked`, incluidos los recuentos por estado, los mensajes de error que listan los estados
válidos y los ejemplos de `biso init` y de `biso config set`. **Una excepción deliberada**: un único
ejemplo conserva un estado de más, para ilustrar que se puede.

- [ ] **Paso 2: regenerar los ejemplos aplicando el algoritmo a mano**

Los ejemplos de `biso ls` y de `biso prime` se calculan con el algoritmo de columnas de 10.4 y con la
regla de anchos de 9.7, no se escriben a ojo. Hay que rehacer los anchos porque los títulos y los
estados han cambiado de longitud.

- [ ] **Paso 3: medir el mensaje y comprobar que cabe**

```bash
awk '/^biso 1\.0\.0 - the task board/,/^That is the loop\.$/' docs/SPEC.md > /tmp/prime.txt
wc -c /tmp/prime.txt
```

Esperado: **como mucho 5.120 bytes en total**, con la parte fija por debajo de 3.456 y el resumen por
debajo de 1.664. Hay que medir las dos mitades por separado, imputando cada línea en blanco a la mitad
del bloque que la precede, que es la regla de 9.5.

- [ ] **Paso 4: si no cabe, recortar contenido, nunca subir el tope**

El orden para recortar es el del apartado 6.2 del diseño: se recorta contenido del mensaje, y el tope
no se toca. Si hay que quitar algo, lo primero que sobra es detalle de las reglas, no una orden del
ciclo.

- [ ] **Paso 5: actualizar las cifras publicadas**

`SPEC.md` publica el trío medido, hoy 4.062, 2.963 y 1.099, y `DECISIONES.md` lo repite en su sección
3. Los tres números salen de la medida del paso 3, no de la estimación del diseño.

- [ ] **Paso 6: comprobar que el tablero viejo no queda en ningún sitio**

```bash
grep -c 'Blocked' docs/SPEC.md
grep -c 'Ideas' docs/SPEC.md
```

Esperado: `0` en la primera, y `1` o `2` en la segunda, solo las del ejemplo deliberado.

- [ ] **Paso 7: commit**

```bash
git add docs/SPEC.md
git commit -m "Regenera los ejemplos con el tablero de tres estados y mide el mensaje de arranque"
```

---

## Tarea 11: las listas y las cuentas que enumeran comandos

Cubre los cambios 2, 3, 10, 11, 16, 23, 31, 32, 33, 34 y 35. Van juntas porque **son el patrón de
fallo dominante del proyecto**: el mismo dato copiado en sitios distantes.

**Ficheros:**
- Modificar: `docs/SPEC.md`, secciones 4.3, 4.12, 10 (cabecera y tabla), 10.6, 11, 12.1, 12.3, 13,
  14 y 15

- [ ] **Paso 1: encontrar todas las copias**

```bash
grep -n 'Diecisiete comandos\|Son treinta líneas\|cuatro verbos de ciclo' docs/SPEC.md
grep -n '`new`, `set`, `start`, `note`, `comment`, `finish` y `archive`' docs/SPEC.md
```

Esperado: las tres cuentas, y **dos** apariciones de la lista de siete comandos, una en 10.6 y otra en
la tabla de 12.1.

- [ ] **Paso 2: las once ediciones**

Diecisiete comandos pasan a diecinueve y nueve del ciclo a once, en la cabecera de la sección 10 y en
su tabla, que gana dos filas. Las dos listas de productores de `task.write` ganan los dos verbos. La
tabla de lectura dirigida de 4.12 los gana también. La tabla de avisos de 4.3 gana los dos avisos
nuevos. La ayuda de primer nivel de la sección 11 los lista y su cuenta de líneas cambia. La tabla de
identificadores de 12.3 gana los siete `code` del apartado 5.2 del diseño. La sección 13 dice desde
cuándo obliga el contrato. La sección 14 gana el arrendamiento aplazado y el requisito del descarte
retirado. Y la sección 15 pasa de cuatro verbos de ciclo a seis.

- [ ] **Paso 3: comprobar que las cuentas cuadran contando de verdad**

```bash
grep -c '^| `' docs/SPEC.md
awk '/^## 11\./,/^## 12\./' docs/SPEC.md | grep -c '^'
grep -c 'Diecisiete comandos' docs/SPEC.md
```

Esperado: `0` en la última. Las otras dos son para contar a mano las filas de la tabla de comandos y
las líneas del bloque de ayuda, y comprobar que coinciden con lo que dice la prosa.

- [ ] **Paso 4: commit**

```bash
git add docs/SPEC.md
git commit -m "Actualiza las listas y las cuentas de comandos, y el contrato de estabilidad"
```

---

## Tarea 12: `DECISIONES.md`

Cubre los nueve cambios de la sección 11.2 del diseño, menos el de `CLAUDE.md`.

**Ficheros:**
- Modificar: `docs/DECISIONES.md`, secciones 3, 4, 6 y 9, más dos secciones nuevas

**Lo que consume:** las cifras medidas en la tarea 10.

- [ ] **Paso 1: comprobar el estado de partida**

```bash
grep -n '4.062\|2.963\|1.099' docs/DECISIONES.md
grep -c '^### 9\.' docs/DECISIONES.md
```

Esperado: las tres cifras del presupuesto viejo, y seis apartados en la sección 9.

- [ ] **Paso 2: reescribir**

La sección 9 entera pasa de "cuatro requisitos identificados y no incorporados" a los cuatro resueltos,
aplazado o retirado, con el diagnóstico correcto del 9.1, la retirada de `default_assignee` y lo que
esa decisión cuesta, y el argumento bueno para retirar el 9.4. La sección 3 recibe las cifras medidas y
el motivo de no subir el total. La sección 4 recibe `default_assignee` en la lista de lo que se deja
fuera. La sección 6 recibe por qué se retira la regla posicional, con la contradicción del penúltimo, y
por qué el tablero por defecto no trae `Ideas`. Y entran dos secciones nuevas: el criterio de estado
frente a campo, y los riesgos conocidos de la sección 12 del diseño.

- [ ] **Paso 3: comprobar que las cifras coinciden con las de `SPEC.md`**

```bash
grep -o '[0-9]\.[0-9][0-9][0-9] bytes' docs/SPEC.md | sort -u
grep -o '[0-9]\.[0-9][0-9][0-9] bytes' docs/DECISIONES.md | sort -u
```

Esperado: que las dos listas cuadren. Es el patrón de fallo dominante y es lo único que hay que
comprobar aquí de verdad.

- [ ] **Paso 4: commit**

```bash
git add docs/DECISIONES.md
git commit -m "Cierra en DECISIONES los cuatro requisitos del modelo de estados"
```

---

## Tarea 13: `CLAUDE.md`

Cubre el noveno cambio de la sección 11.2 del diseño.

**Ficheros:**
- Modificar: `CLAUDE.md`

- [ ] **Paso 1: ver lo que ha quedado obsoleto**

```bash
grep -n 'quince comandos\|Cuatro requisitos' CLAUDE.md
```

Esperado: las dos cosas. La cifra de comandos ya estaba mal antes de este trabajo, porque decía quince
y `SPEC.md` decía diecisiete.

- [ ] **Paso 2: actualizar**

La cifra pasa a diecinueve. La sección de cuatro requisitos sin incorporar se sustituye por una frase
que diga que el modelo de estados está cerrado y remita al documento de diseño. La sección de lo que
falta por decidir se queda con una sola cosa bloqueante, la persistencia, porque el lenguaje sigue
abierto pero no bloquea la especificación.

- [ ] **Paso 3: comprobar**

```bash
grep -c 'quince comandos' CLAUDE.md
bash ~/Hub/dotfiles/scripts/link-agent-instructions.sh --check . || true
```

Esperado: `0` en la primera. La segunda comprueba que los tres agentes siguen leyendo lo mismo, que es
una regla de la máquina y no de este proyecto.

- [ ] **Paso 4: commit**

```bash
git add CLAUDE.md
git commit -m "Actualiza CLAUDE.md tras cerrar el modelo de estados"
```

---

## Tarea 14: la pasada final de coherencia

Cubre el cambio 38, el barrido del vocabulario, más la búsqueda de afirmaciones a medias. El cambio 38
estaba asignado a la tarea 1 y se movió aquí durante la ejecución, porque es un barrido de documento
entero y las tareas 2 a 13 reescriben buena parte de ese documento: hacerlo antes obligaría a
rehacerlo. Lo demás existe porque la sección 8 de `DECISIONES.md` dice que al corregir una afirmación
hay que buscarla en todo el documento, y esa búsqueda solo se puede hacer al final.

- [ ] **Paso 1: los greps que tienen que dar cero**

```bash
for p in 'default_status' 'defaultStatus' 'default_assignee' 'Blocked' \
         'notStartedHidden' 'penúltimo' 'with two statuses' '—'; do
  printf '%-20s %s\n' "$p" "$(grep -c "$p" docs/SPEC.md)"
done
```

Esperado: `0` en los ocho.

- [ ] **Paso 2: los greps que tienen que dar lo mismo en los dos documentos**

```bash
grep -o 'initial_status\|active_status\|terminal_status' docs/SPEC.md | sort | uniq -c
grep -o 'initial_status\|active_status\|terminal_status' docs/DECISIONES.md | sort | uniq -c
```

Esperado: que ningún papel aparezca en un documento con un nombre y en el otro con otro.

- [ ] **Paso 3: el barrido del vocabulario (cambio 38)**

Aplicar la sección `## Vocabulario de este documento` a todo el documento, retirando los usos que ella
misma prohíbe. Cada resultado de estos hay que mirarlo a mano, porque tres de las cinco palabras
tienen un uso legítimo y solo el contexto lo dice:

```bash
grep -n 'columna\|panel\|tarjeta\|ticket' docs/SPEC.md
```

Un `columna` que nombre una de las ocho columnas de la tabla que imprimen `biso ls` y `biso prime` es
correcto y se queda. Un `columna` que nombre un estado del tablero es un fallo y pasa a decir estado:
hay uno conocido alrededor de la línea 1107, en la frase "que no se corresponda con una columna del
tablero". Un `tarjeta` que hable de Trello es correcto. `panel` no es correcto nunca.

- [ ] **Paso 4: comprobar que no ha quedado ninguna regla que cite el nombre de un estado**

```bash
grep -n 'In Progress' docs/SPEC.md | grep -v 'example\|ejemplo\|^\s*[0-9]*:  ' | head -20
```

Cada resultado hay que mirarlo a mano. Un `In Progress` dentro de un ejemplo de salida es correcto. Un
`In Progress` dentro de una regla es un fallo, porque las reglas hablan de papeles.

- [ ] **Paso 4: releer las treinta y ocho entradas del diseño y marcar cada una**

Ir por la sección 11 del documento de diseño de arriba abajo y comprobar una por una que está aplicada.
Es la comprobación que ninguna otra sustituye.

- [ ] **Paso 5: commit**

```bash
git add -A
git commit -m "Pasada final de coherencia sobre los dos documentos"
```

---

## Lo que este plan deja fuera a propósito

- **La decisión de persistencia**, que sigue abierta y con ella el mecanismo del arrendamiento.
- **La decisión de lenguaje**, que sigue abierta y no condiciona a la especificación.
- **La limpieza de `ready` y `blocked`**, que son complementarios y cuentan un solo hecho. Es anterior
  a este trabajo y merece su propia decisión.
