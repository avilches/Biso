# `biso get`

## Firma

```
biso get <ref> [--id] [--match] [--section <name>]... [--explain-urgency] [--closure]
```

## Parámetros

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `<ref>` | sí | referencia | | no | no | |
| `--id` | no | booleano | falso | no | no | `--match` |
| `--match` | no | booleano | falso | no | no | `--id` |
| `--section <name>` | no | `meta`, `desc`, `ac`, `plan`, `notes`, `summary`, `comments`, `question` | todas | sí | sí | |
| `--explain-urgency` | no | booleano | falso | no | no | |
| `--closure` | no | booleano | falso | no | no | |

`--section` sirve para pedir solo una parte. `biso get MYP-11 --section ac` imprime los criterios con
sus claves y cuesta unas decenas de bytes en vez de la ficha entera, que es lo que hace falta antes de
marcar uno.

**Varias secciones salen siempre en el orden fijo de la ficha completa** (`meta`, `desc`, `ac`,
`plan`, `notes`, `summary`, `comments`, `question`), nunca en el orden en que se pidieron: `--section
plan,ac` y `--section ac,plan` imprimen lo mismo. Una sección repetida se guarda una vez, con la misma
regla y el mismo aviso que cualquier flag repetible (["Repetición y listas separadas por
comas"](../valores-de-entrada.md#repetición-y-listas-separadas-por-comas)).

**`--closure` es compatible con `--section`, con `--id` y con `--match`, sin ninguna
excepción**, exactamente igual que `--explain-urgency`: los dos son flags que añaden un bloque propio
al final de la ficha y no recortan ni dependen de qué secciones se pidieron
(["`--closure`"](#--closure)). Sobre una referencia que no resuelve a una sola tarea (varias
coincidencias, error 5; la tarea no se puede leer, error 3; no existe, error 4) no hay ninguna ficha
que completar, así que `--closure` no tiene ningún efecto observable y no añade ningún código de
salida propio.

## Comportamiento, caso a caso

La resolución de `<ref>` está en la sección ["Cómo se resuelve una referencia a una tarea"](../referencias.md) y no se repite. Lo propio de este comando:

| Caso | Qué pasa |
|---|---|
| La referencia resuelve a una tarea | Se imprime, código 0 |
| La referencia es texto y encaja con varias | Error 5, y las candidatas salen **por stdout** exactamente como las imprimiría `biso ls --search "<texto>"` (["`biso ls`"](ls.md)), y ese "exactamente" son tres cosas y solo tres: el mismo orden, el mismo límite de 30 y el mismo aviso de recorte si hace falta. **No son los filtros de ese listado**: las candidatas salen del tablero entero y una tarea en el estado terminal aparece entre ellas, aunque `biso ls` la deje fuera por el valor por defecto de su `--status` (["Cómo se resuelve una referencia a una tarea"](../referencias.md#la-búsqueda-por-texto)) |
| La referencia es texto y encaja con una | Se imprime, con `note: "CRLF" matched MYP-11` por stderr |
| La tarea está archivada | Se imprime, con `note: MYP-11 is archived` por stderr |
| La tarea no se puede leer: un valor de `status`, `type` o `priority` que el tablero no declara, una fecha que no es una fecha, o cualquier otro motivo de ["Qué se comprueba"](../garantias.md#qué-se-comprueba) | Error 3 con `code` `undecodable_task`, sin imprimir nada de la ficha, con `--section` o sin él. El mensaje dice qué campo y qué valor hay guardado, y el texto de cada motivo está en ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), junto con cómo se arregla |
| Otra tarea del tablero no se puede leer | La ficha se imprime igual, y por stderr sale el aviso que nombra las tareas ilegibles del tablero, con código 0 |
| `--section` con un nombre inventado | Error 2, `code` igual a `unknown_section`, con los ocho nombres válidos |
| `--section` de una sección vacía | No imprime esa sección, y si no queda ninguna sección que imprimir, la salida está vacía y el código sigue siendo 0 |

**Sin `--section`, la ficha completa imprime siempre las ocho secciones fijas, vacías incluidas,
marcadas con `(empty)`.** Es solo con `--section` que una sección vacía se omite en vez de imprimirse
vacía; sin el flag, omitir una sección la confundiría con una que no se ha pedido.

## Salida

```
MYP-11  Normalize CRLF in the diff
status     In Progress          type       bug
priority   high                 urgency    19.0
assignees  @claude              author     @avilches
labels     parser               parent     -
due        -                    ordinal    -
created    2026-09-06 09:12     updated    2026-09-06 11:40
depends    -                    blocks     MYP-40
blocked by 0 total
unblocks   2 total
lease      2026-09-06 15:40     holder     @claude
refs       docs/bugs/BUG-02.md, notes/a\,b.md

## Description
The diff compares byte by byte and marks as different two lines that only
differ in the line ending.

## Acceptance Criteria
- [x] #1 The diff ignores CRLF
- [ ] #3 There is a test that covers it

## Implementation Plan
1. Read the parser.
2. Add the CRLF case.

## Implementation Notes
The parser already normalized LF, CRLF was missing.

## Final Summary
(empty)

## Comments
#1  @avilches, 2026-09-06 10:02
A user with a Windows clone reported this.

## Open Question
(empty)
```

La cifra de urgencia del ejemplo puede no ser esta; el motivo está en la sección
["Urgencia"](../modelo-de-datos/urgencia.md).

Los encabezados de esta salida son un formato de presentación, no un formato de almacenamiento.

**Cómo se escribe cada valor de ese bloque.** Una lista (`assignees`, `labels`, `depends`, `blocks`,
`refs`) va en una línea, con sus valores separados por coma y espacio, y **cada valor se escribe con la
misma regla de escape que se usa al guardarlo**
(["Repetición y listas separadas por comas"](../valores-de-entrada.md#repetición-y-listas-separadas-por-comas)):
una coma dentro de un valor sale como `\,`, una barra invertida sale como `\\`, y ningún otro carácter
se toca. Es exactamente la regla de entrada al revés, así que lo que se lee en la ficha es, valor a valor,
lo que habría que teclear en `--add-refs` para guardarlo. Solo `refs` puede llevar esos dos caracteres:
en `assignees`, `labels`, `depends` y `blocks` ningún valor los admite, y su línea sale sin cambios.

El separador es siempre una coma **sin escapar** seguida de un espacio, y una coma escapada nunca separa.
Por eso las referencias `a,b` y `c.md` salen como `a\,b, c.md`, y las referencias `a` y `b` salen como
`a, b`; una ruta `C:\dir\notes.md` sale como `C:\\dir\\notes.md`. Los espacios de un valor se conservan
tal cual, los de los extremos incluidos, porque quien separa es la coma y no el espacio. El ejemplo de
arriba lo muestra: su línea `refs` lleva dos referencias, `docs/bugs/BUG-02.md` y `notes/a,b.md`.

**Cómo se lee "sin escapar".** La línea se recorre de izquierda a derecha y una barra invertida se lleva
consigo el carácter que le sigue. Una referencia que termina en barra invertida sale con la barra doble, y
esa barra doble ya está completa, así que la coma que viene detrás sí separa: las referencias `a\` y `b`
salen como `a\\, b`, que se parte en dos, mientras que la única referencia `a, b` sale como `a\, b`, que
no se parte. Para leer la línea de vuelta se parte por cada coma sin escapar, se descarta el espacio que
la sigue y se deshace el escape de cada trozo.

**Lo que se copia a un flag es el valor de una referencia, no la línea entera.** El valor de una
referencia, tal como lo muestra la ficha, se puede pasar a `--add-refs` o a `--rm-refs` y nombra esa
referencia. La línea entera no: el espacio que sigue a cada coma separadora forma parte de la ficha y no
del valor, y un flag que recibe `a\,b, c.md` entiende la segunda referencia como ` c.md`, con su espacio.

**Una excepción a la lectura exacta.** Una referencia cuyo valor es exactamente `-` se imprime igual que
una lista vacía, porque la regla de entrada no define ningún escape que la distinga. No hace falta una
excepción parecida para el salto de línea: una referencia con un `\r` o un `\n` se rechaza al
escribirla (["Una referencia con un salto de línea se rechaza al
escribirla"](../../decisiones/detalles.md#una-referencia-con-un-salto-de-línea-se-rechaza-al-escribirla)),
así que la ficha nunca tiene una que partir.

La ficha sigue siendo un formato de presentación, y quien lea los valores con un programa los lee en
`--json`, donde cada uno es un elemento de la lista y nunca lleva escape. La razón de esta regla, y las
salidas que se descartaron, están en
["La ficha escapa la coma y la barra invertida de una lista"](../../decisiones/detalles.md#la-ficha-escapa-la-coma-y-la-barra-invertida-de-una-lista).

Un campo sin valor es un guion, como en las columnas de
["`biso ls`"](ls.md). Los instantes (`created`, `updated`, `lease`) se escriben con el día y la hora
hasta el minuto, `YYYY-MM-DD HH:MM`, que es la precisión que se lee: el segundo está en `--json`,
que es donde lo lee un programa.

**`ordinal` es el único valor del bloque que no se imprime tal cual está guardado.** Lo que guarda es
una clave de texto que no se puede teclear y que no dice nada que quien la lee pueda usar
(["El orden manual y su clave"](../modelo-de-datos/orden-manual.md)), así que la ficha imprime
`manual` cuando la tarea tiene una y el guion de siempre cuando no. Lo que de verdad hace falta saber
de ese campo es si la tarea tiene un sitio decidido a mano, que es lo que permite nombrarla como
vecina de un `--above` o de un `--below`; la clave en crudo está en `--json`, que es donde la lee un
programa.

**`blocked by` y `unblocks` son las dos líneas que dan el recuento del cierre transitivo del grafo de
dependencias**, es decir, no solo lo que esta tarea depende directamente y lo que depende
directamente de ella (`depends` y `blocks`, arriba), sino toda la cadena: `blocked by` cuenta cuántas
tareas tienen que terminar, siguiendo `dependencies` las veces que haga falta, antes de que esta sea
tocable, y `unblocks` cuántas dependen de esta por esa misma cadena, en cualquier profundidad. Las dos
cuentan solo tareas **sin terminar**, con la misma definición exacta que usan los términos `bloquea` y
`bloqueada` de la fórmula de urgencia: ni en el estado terminal ni archivada
(["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia)).

Van en una línea propia cada una, sin pareja, con la forma `blocked by <n> total` y
`unblocks <n> total`, entre `depends`/`blocks` y la línea condicional `lease`/`holder`. **Las dos
salen siempre, aunque el recuento sea cero**, a diferencia de `lease`: un recuento de cero es un hecho
sobre la tarea (no depende transitivamente de nada sin terminar, o no desbloquea nada) y no la ausencia
de un dato, así que esconder la línea confundiría las dos cosas: es el mismo principio que en
`--json` prohíbe que una clave presente se sustituya por su ausencia según el valor
(["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias)), aplicado aquí a la
salida de texto. Pertenecen al bloque de metadatos, así que las trae `--section meta` y ninguna otra
sección, igual que el resto de las líneas de este bloque.

**Es distinto de `blocks`, que sigue siendo solo directo y sin cambios.** `blocks` es la lista de
identificadores que dependen directamente de esta tarea y siguen sin terminar
(["Los campos derivados"](../modelo-de-datos/index.md#los-campos-derivados)); `unblocks` es su cierre
transitivo, cuántos hay en cualquier profundidad de esa misma cadena. `MYP-11` del ejemplo de arriba
tiene `blocks MYP-40` (un solo dependiente directo) y a la vez `unblocks 2 total`, porque `MYP-40`
desbloquea a su vez a otra tarea sin terminar (["`--closure`"](#--closure) trae la cadena entera).
`depends` es la contrapartida directa de `blocked by`, con la misma relación.

**Estos dos recuentos son de `biso get` sobre una sola tarea, no de `biso ls`.** No están en el
esquema de `task.list` (["`biso ls`"](ls.md#el-esquema-json)) ni en las cuatro listas de
`biso prime` (["`biso prime`"](prime.md#el-esquema-json)), salvo la única excepción con nombre propio
de la sección ["`biso prime`"](prime.md#la-salida-literal). La razón es
["El presupuesto de arranque"](../presupuestos.md#el-presupuesto-de-arranque): calcular el cierre
transitivo de una tarea cuesta proporcional al tamaño de su propia cadena de dependencias, y
`biso ls` imprime hasta 300 a la vez, así que repetirlo tarea a tarea en un listado multiplicaría ese
coste por cada fila sin que nadie lo hubiera pedido, justo lo que la regla 1 de ese presupuesto
prohíbe. `biso get` lee una tarea, así que paga el coste de una sola cadena.

Para el árbol completo de identificadores de cada cierre, con las tareas terminadas y archivadas
incluidas, está el flag `--closure` (["`--closure`"](#--closure)).

**La línea `lease` sale solo cuando la tarea tiene arrendamiento**, y entonces sale con sus campos:
`lease` es `leaseExpiresAt`, con el mismo formato de instante que `created` y `updated`, y `holder` es
`leaseHolder` (["El vaciado"](../lease.md#el-vaciado) de `lease.md`). Los dos aparecen y desaparecen juntos, porque esa misma regla no admite uno sin
el otro. Pertenece al bloque de metadatos, así que la trae `--section meta` y no ninguna otra sección.
Es la única línea condicional de ese bloque, la única que aparece o desaparece según la tarea, y por
eso va al final de todas las líneas del bloque: así ninguna de las de arriba, `blocked by` y `unblocks`
incluidas, cambia de sitio según la tarea. Eso no choca con la regla de que la ficha
completa imprime las ocho secciones aunque estén vacías, porque lo condicional es una línea del bloque
y no el bloque. Una tarea sin arrendamiento **no imprime la línea**, en vez de imprimirla con dos
guiones, porque eso pondría dos guiones en la ficha de casi todas las tareas del tablero y la ausencia
de la línea dice lo mismo. Esta es la única forma de ver los campos sin `--json`: `biso prime` no
los trae (["La salida literal"](prime.md#la-salida-literal)) y `biso ls` tampoco (["`biso ls`"](ls.md)). Si el arrendamiento está vencido, el instante ya lo dice y la
ficha no añade ninguna marca; el derivado `leaseExpired` ya calculado está en `--json`.

Con `--section ac`, solo el encabezado con el identificador y el título, y la sección pedida:

```
MYP-11  Normalize CRLF in the diff

## Acceptance Criteria
- [x] #1 The diff ignores CRLF
- [ ] #3 There is a test that covers it
```

Con `--section question` sobre MYP-60, la tarea con la pregunta abierta del ejemplo de la sección ["La salida literal"](prime.md#la-salida-literal), la
sección sale rellena con una forma parecida a la de `## Comments`: el autor y el instante en una línea
y el cuerpo debajo, pero **sin la clave**, porque `question` es del tipo `Question`, un valor único y
no una lista direccionable (["La pregunta abierta"](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta)): no hay un selector que pueda señalar "la pregunta número tal".

```
MYP-60  Confirm the retry budget for the upload endpoint

## Open Question
@claude, 2026-09-06 09:30
Should the retry budget be shared with the download endpoint or kept separate?
```

Con `--explain-urgency`, al final y por stdout:

```
urgency 19.0
  priority high      6.0 * 1.00 =   6.00
  active             4.0 * 1.00 =   4.00
  blocking           8.0 * 1.00 =   8.00
  blocked           -5.0 * 0.00 =   0.00
  due               12.0 * 0.00 =   0.00
  has criteria       1.0 * 1.00 =   1.00
  age 0 days         0.5 * 0.00 =   0.00
                                 -------
                                   19.00
```

La primera línea nombra la prioridad de la tarea junto al término, y una tarea que no tiene ninguna
imprime `priority (none)`, que es el caso con peso propio de ["La urgencia"](../modelo-de-datos/urgencia.md#la-urgencia) y no el de una
prioridad que el tablero ya no declara, que hace la tarea ilegible.

El término `active` vale `1.00` solo si el estado es el activo y la tarea no tiene una pregunta
abierta (["La pregunta abierta"](../modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta)); en cualquier otro caso vale `0.00`, y la etiqueta dice cuál de los dos motivos se
aplica: `not active` si el estado no es el activo, `active, waiting` si lo es pero la tarea espera una
respuesta.

Sobre una tarea en el estado terminal, el desglose se sustituye por una línea:

```
urgency 0.0
  terminal status, urgency is zero by definition
```

### `--closure`

Con `--closure`, al final y por stdout, después del bloque de `--explain-urgency` si los dos se piden
a la vez:

```
closure
  blocked by 0 total  -
  unblocks   2 total  MYP-40, MYP-71, MYP-72
```

Es el cierre transitivo completo del grafo de `dependencies`, en las dos direcciones, sobre `MYP-11`
del ejemplo de arriba. En este tablero ficticio, `MYP-40` (la tarea que `MYP-11` bloquea
directamente) tiene a su vez dos dependientes encadenados: `MYP-71`, "Migrate callers to the new
config loader" (`To Do`), que depende de `MYP-40`, y `MYP-72`, "Remove the deprecated config loader
path" (`Done`), que depende de `MYP-71`. El cierre hacia abajo de `MYP-11` es por tanto
`MYP-40`, `MYP-71` y `MYP-72`, los tres identificadores completos, con `MYP-72` incluido aunque ya
está terminada.

**La lista es completa; el recuento no.** `unblocks 2 total` cuenta solo `MYP-40` y `MYP-71`, las dos
tareas del cierre que siguen sin terminar, con la misma definición de más arriba (["Salida"](#salida));
`MYP-72` entra en la lista igualmente, porque la lista describe el grafo entero tal cual está, no lo
que queda por hacer. Es la diferencia deliberada entre lo que imprime el flag (el grafo, completo) y
lo que imprime la ficha sin él (cuánto de ese grafo sigue bloqueando de verdad). No hay una tercera
cifra "de la lista": quien la necesite cuenta los identificadores que imprime.

Cada línea lleva la etiqueta rellenada a las once celdas del bloque de metadatos
(["Salida"](#salida)), el recuento con la forma `N total`, dos espacios literales y la lista de
identificadores separados por coma y espacio, en el mismo orden ascendente que usa `blocks`
(["Los campos derivados"](../modelo-de-datos/index.md#los-campos-derivados)). Un identificador de
tarea no puede llevar coma ni barra invertida
(["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)), así que
la lista no necesita ningún escape. Con un cierre vacío en esa dirección, la lista es el guion de
siempre, como en `blocked by 0 total  -` de arriba, y nunca una lista vacía a secas.

**El recorrido nunca falla ni se cuelga, sea cual sea el estado del almacén.** Es una vista de
lectura, y cada cosa que podría romperla tiene una salida definida:

| Caso | Qué imprime `--closure` |
|---|---|
| Un ciclo de dependencias en el grafo (["`biso doctor`"](doctor.md#qué-comprueba), `dependency_cycle`) | El recorrido no repite un identificador ya visitado y sigue por el resto del grafo; no señala el ciclo, porque diagnosticarlo es trabajo de `biso doctor` y no de esta vista de lectura |
| Una dependencia que apunta a una tarea que no existe (["`biso doctor`"](doctor.md#qué-comprueba), `dependency_not_found`) | Ese identificador no entra en el cierre ni en el recuento: no hay ninguna tarea que añadir, y el recorrido no sigue más allá de él |
| Una tarea del cierre que no se puede leer (["Qué se comprueba"](../garantias.md#qué-se-comprueba)) | Se excluye del cierre entero, de la lista y del recuento, con la misma razón que excluye a un dependiente ilegible de `blocks` (["Los campos derivados"](../modelo-de-datos/index.md#los-campos-derivados)): no se puede afirmar nada, ni que bloquea ni que sigue el grafo, de un dato que no se puede interpretar. El recorrido no sigue las dependencias de esa tarea, porque no se pueden leer |
| Un ciclo que vuelve a la propia tarea de la que se pidió el cierre | La tarea de partida nunca aparece en su propio cierre, en ninguna de las dos direcciones, aunque el ciclo la alcance |

Ninguno de los tres primeros casos avisa por stderr ni cambia el código de salida: `biso doctor` ya es
el sitio donde se diagnostican, y repetir el aviso en cada `biso get --closure` sobre una tarea
alcanzada por el daño sería ruido, no información nueva.

## El esquema JSON

`data.task` lleva todas las claves de un objeto de ["`task.list`"](ls.md#el-esquema-json) más las del
cuerpo (`description`, `acceptanceCriteria`, `plan`, `notes`, `summary`,
`comments`, `question`), más dos claves que no están en `task.list`: `blockedByCount` y
`unblocksCount`. La lista completa de las primeras vive solo en `ls.md`; el ejemplo de abajo
las repite todas para que sirva de esquema completo, verificable clave a clave:

```json
{
  "schemaVersion": 1,
  "kind": "task.get",
  "generatedAt": "2026-09-06T13:31:09Z",
  "data": {
    "task": {
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
      "archived": false,
      "blockedByCount": 0,
      "unblocksCount": 2,
      "description": "The diff compares byte by byte...",
      "acceptanceCriteria": [ { "key": 1, "text": "The diff ignores CRLF", "checked": true },
                              { "key": 3, "text": "There is a test that covers it", "checked": false } ],
      "plan": "1. Read the parser.\n2. Add the CRLF case.",
      "notes": "The parser already normalized LF, CRLF was missing.",
      "summary": null,
      "comments": [ { "key": 1, "author": "@avilches", "createdAt": "2026-09-06T10:02:11Z", "body": "A user with a Windows clone..." } ],
      "question": null,
      "urgencyBreakdown": { "priority": 6.0, "active": { "value": 4.0, "reason": null }, "blocking": 8.0,
                            "blocked": 0.0, "due": 0.0, "criteria": 1.0, "age": 0.0 },
      "closure": { "blockedBy": [], "unblocks": ["MYP-40", "MYP-71", "MYP-72"] }
    }
  }
}
```

Sobre una tarea en el estado terminal, `urgencyBreakdown` sale igualmente con `--explain-urgency`,
con sus siete términos a `0.0` y `active.reason` en `"not_active"`: es la forma que le corresponde a
una urgencia que vale cero por definición, sin ningún término calculado, y la clave sigue apareciendo
siempre que se escribe el flag, como promete la regla de abajo.

**`urgencyBreakdown` solo sale con `--explain-urgency`**, igual que el desglose de la salida de texto, y
el ejemplo de arriba es el de una llamada que la lleva. Junto con `closure`, de más abajo, son las dos
únicas claves de todo el documento que añade un flag en vez de una sección, y la excepción a la regla
de las claves siempre presentes está declarada en la sección
["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias), junto con la otra
cosa que `biso get` hace con sus flags: recortar `data.task` con `--section`.

`urgencyBreakdown.active` es el único término que no es un número suelto: `value` es el número que
entra en la suma, el producto del coeficiente por el factor, igual que en los demás términos.
`reason` vale `null` cuando el término contribuye, y cuando contribuye `0.0` dice por qué:
`"not_active"` si el estado no es el activo, y `"waiting"` si lo es pero hay una pregunta abierta.

**`blockedByCount` y `unblocksCount` son el equivalente en JSON de las líneas `blocked by` y
`unblocks` de la ficha de texto** (["Salida"](#salida)): dos enteros que cuentan las tareas sin
terminar del cierre transitivo de `dependencies`, hacia arriba y hacia abajo. **Están siempre
presentes**, con o sin `--closure`, porque a diferencia de `urgencyBreakdown` y de `closure` no los
gobierna ningún flag: son dos claves ordinarias del bloque de metadatos, como `blocks` o `urgency`, y
por eso `--section` sí los recorta (los trae `meta` y ninguna otra sección), mientras que ni
`--section` ni la ausencia de `--closure` los hacen desaparecer del bloque `meta`. **No están en
`task.list`, a diferencia del resto de claves de este bloque**: calcular el cierre transitivo de una
tarea cuesta proporcional al tamaño de su propia cadena, y repetirlo en cada fila de un listado de
hasta 300 violaría la primera regla de ["El presupuesto de arranque"](../presupuestos.md#el-presupuesto-de-arranque), así que
`biso ls --json` y las cuatro listas de `biso prime --json` no las llevan
(["`biso ls`"](ls.md#el-esquema-json), ["`biso prime`"](prime.md#el-esquema-json)).

**`closure` solo sale con `--closure`**, con la misma regla que `urgencyBreakdown` y el mismo motivo:
es el equivalente en JSON del bloque de texto de ["`--closure`"](#--closure), y `--section` no lo
gobierna, así que aparece con el flag sea cual sea lo que se pidió con `--section`. `blockedBy` y
`unblocks` son el cierre transitivo completo en cada dirección: todos los identificadores alcanzados,
sin filtrar por si la tarea que nombran está terminada o archivada, ordenados igual que `blocks`
(["Los campos derivados"](../modelo-de-datos/index.md#los-campos-derivados)); una lista vacía es `[]`,
nunca `null`, con la regla general de ["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias). Es la diferencia
con `blockedByCount`/`unblocksCount`: la lista cuenta el grafo entero, el recuento solo lo que sigue
sin terminar (["`--closure`"](#--closure) trae el porqué completo, con un ejemplo donde los dos
difieren).

**`ordinal` es una cadena o `null`**, y este sobre es el único sitio de `biso get` donde se puede
leer la clave en crudo, porque la ficha de texto imprime `manual` en su lugar. La tarea del ejemplo
no tiene ninguna; una que la tuviera traería, por ejemplo, `"ordinal": "m"`. Pertenece al bloque de
metadatos, así que la trae `--section meta`, como el resto de las claves de ese bloque.

Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas. **Una sección
pedida que esté vacía sigue trayendo su clave, con el valor `null` o la lista vacía que le
corresponda**: quitarla haría que la presencia de una clave dependiera de los datos, que es justo lo
que prohíbe ["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias). Omitir la sección vacía es cosa de la salida de texto, donde
lo que sobra es un encabezado sin nada debajo.

Con varias coincidencias, `kind` es `task.candidates`, `data.tasks` es la lista y el código es 5.
Ese sobre lleva `data` y no `error`, aunque el código no sea cero, y es el único de todo el programa
que hace eso: las candidatas son un dato, no la descripción de un fallo. **`biso get` es también el
único comando que lo emite**: cualquier otro que resuelva una referencia ambigua con `--json`, como
`biso set` o el `--parent` de `biso ls`, contesta el sobre de error de ["Los errores en JSON"](../contrato-json.md#los-errores-en-json) con el mismo código 5,
porque la tabla de `kind` de ["El sobre"](../contrato-json.md#el-sobre) le da `task.candidates` a `get` y a ningún otro.

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea impresa | 0 |
| `<ref>` mal formada, `--section` inventada, `--id` con `--match` | 2 |
| La tarea no se puede leer | 3 |
| No existe, o existió y ya no está | 4 |
| Texto con varias coincidencias | 5 |
| El almacén no responde | 8 |
| No hay tablero | 20 |

`--closure` no añade ningún código propio, con la misma tabla de arriba con `--closure` que sin él:
como `--explain-urgency`, es un flag que decora una ficha ya resuelta y nunca decide por sí solo si la
llamada acaba bien o mal.

## `biso get --help`

```
Usage: biso get <ref> [options]

Show one task. <ref> is an id (MYP-11), a bare number (11) or free text
("CRLF"). Free text that matches several tasks lists them and exits 5; it
never picks one for you.

Free text searches the title, description, plan, notes, final summary, the
text of the criteria, the body of the comments, the body of the open question
and the labels. A match in the title always
wins over a match anywhere else. `biso ls --search` uses this same scope.

Options:
      --id                   force <ref> to be read as an id
      --match                force <ref> to be read as free text
      --section <name>       print only these sections; repeatable or comma
                             separated. One of: meta, desc, ac, plan, notes,
                             summary, comments, question
      --explain-urgency      show how the urgency number is built
      --closure              show the full transitive closure of dependencies,
                             both directions
  -h, --help                 show this help

Exit codes:
  0  printed          4  not on this board
  2  bad usage        5  the text matched several tasks
  3  the task could not be read
  8  the board could not respond
  20 no board here

Examples:
  biso get MYP-11
  biso get 11 --section ac
  biso get "CRLF"
  biso get MYP-11 --explain-urgency
  biso get MYP-11 --closure
```

---

