# `biso get`

## Firma

```
biso get <ref> [--id] [--match] [--section <name>]... [--explain-urgency]
```

## Parámetros

| Parámetro | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|
| `<ref>` | sí | referencia | | no | no | |
| `--id` | no | booleano | falso | no | no | `--match` |
| `--match` | no | booleano | falso | no | no | `--id` |
| `--section <name>` | no | `meta`, `desc`, `ac`, `plan`, `notes`, `summary`, `comments`, `question` | todas | sí | sí | |
| `--explain-urgency` | no | booleano | falso | no | no | |

`--section` sirve para pedir solo una parte. `biso get MYP-11 --section ac` imprime los criterios con
sus claves y cuesta unas decenas de bytes en vez de la ficha entera, que es lo que hace falta antes de
marcar uno.

**Varias secciones salen siempre en el orden fijo de la ficha completa** (`meta`, `desc`, `ac`,
`plan`, `notes`, `summary`, `comments`, `question`), nunca en el orden en que se pidieron: `--section
plan,ac` y `--section ac,plan` imprimen lo mismo. Una sección repetida se guarda una vez, con la misma
regla y el mismo aviso que cualquier flag repetible (["Repetición y listas separadas por
comas"](../valores-de-entrada.md#repetición-y-listas-separadas-por-comas)).

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

**La línea `lease` sale solo cuando la tarea tiene arrendamiento**, y entonces sale con sus campos:
`lease` es `leaseExpiresAt`, con el mismo formato de instante que `created` y `updated`, y `holder` es
`leaseHolder` (["El vaciado"](../lease.md#el-vaciado) de `lease.md`). Los dos aparecen y desaparecen juntos, porque esa misma regla no admite uno sin
el otro. Pertenece al bloque de metadatos, así que la trae `--section meta` y no ninguna otra sección.
Es la única línea condicional de ese bloque, y por eso va al final de las líneas de dos campos: así
ninguna de las de arriba cambia de sitio según la tarea. Eso no choca con la regla de que la ficha
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

## El esquema JSON

`data.task` lleva todas las claves de un objeto de ["`task.list`"](ls.md#el-esquema-json) más las del
cuerpo (`description`, `acceptanceCriteria`, `plan`, `notes`, `summary`,
`comments`, `question`). La lista completa de las primeras vive solo en `ls.md`; el ejemplo de abajo
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
      "description": "The diff compares byte by byte...",
      "acceptanceCriteria": [ { "key": 1, "text": "The diff ignores CRLF", "checked": true },
                              { "key": 3, "text": "There is a test that covers it", "checked": false } ],
      "plan": "1. Read the parser.\n2. Add the CRLF case.",
      "notes": "The parser already normalized LF, CRLF was missing.",
      "summary": null,
      "comments": [ { "key": 1, "author": "@avilches", "createdAt": "2026-09-06T10:02:11Z", "body": "A user with a Windows clone..." } ],
      "question": null,
      "urgencyBreakdown": { "priority": 6.0, "active": { "value": 4.0, "reason": null }, "blocking": 8.0,
                            "blocked": 0.0, "due": 0.0, "criteria": 1.0, "age": 0.0 }
    }
  }
}
```

Sobre una tarea en el estado terminal, `urgencyBreakdown` sale igualmente con `--explain-urgency`,
con sus siete términos a `0.0` y `active.reason` en `"not_active"`: es la forma que le corresponde a
una urgencia que vale cero por definición, sin ningún término calculado, y la clave sigue apareciendo
siempre que se escribe el flag, como promete la regla de abajo.

**`urgencyBreakdown` solo sale con `--explain-urgency`**, igual que el desglose de la salida de texto, y
el ejemplo de arriba es el de una llamada que la lleva. Es la única clave de todo el documento que un
flag añade, y la excepción a la regla de las claves siempre presentes está declarada en la sección ["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias), junto
con la otra cosa que `biso get` hace con sus flags: recortar `data.task` con `--section`.

`urgencyBreakdown.active` es el único término que no es un número suelto: `value` es el número que
entra en la suma, el producto del coeficiente por el factor, igual que en los demás términos.
`reason` vale `null` cuando el término contribuye, y cuando contribuye `0.0` dice por qué:
`"not_active"` si el estado no es el activo, y `"waiting"` si lo es pero hay una pregunta abierta.

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
```

---

