# `biso get`

## Firma

```
biso get <ref> [--id] [--match] [--section <name>]... [--explain-urgency]
```

## Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí | referencia | | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |
| `--section <name>` | | no | `meta`, `desc`, `ac`, `dod`, `plan`, `notes`, `summary`, `comments`, `question` | todas | sí | sí | |
| `--explain-urgency` | | no | booleano | falso | no | no | |

`--section` sirve para pedir solo una parte. `biso get MYP-11 --section ac` imprime los criterios con
sus claves y cuesta unas decenas de bytes en vez de la ficha entera, que es lo que hace falta antes de
marcar uno.

## Comportamiento, caso a caso

La resolución de `<ref>` está en la sección ["Cómo se resuelve una referencia a una tarea"](../referencias.md) y no se repite. Lo propio de este comando:

| Caso | Qué pasa |
|---|---|
| La referencia resuelve a una tarea | Se imprime, código 0 |
| La referencia es texto y encaja con varias | Error 5, y las candidatas salen **por stdout** exactamente como las imprimiría `biso ls --search "<texto>"` (["`biso ls`"](ls.md)): mismo orden, mismo límite de 30 y mismo aviso de recorte si hace falta |
| La referencia es texto y encaja con una | Se imprime, con `note: "CRLF" matched MYP-11` por stderr |
| La tarea está archivada | Se imprime, con `note: MYP-11 is archived` por stderr |
| La tarea no se puede leer | Error 3, según la regla de lectura dirigida de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar) |
| `--section` con un nombre inventado | Error 2, con los nueve nombres válidos |
| `--section` de una sección vacía | No imprime esa sección, y si no queda ninguna sección que imprimir, la salida está vacía y el código sigue siendo 0 |

**Sin `--section`, la ficha completa imprime siempre las nueve secciones fijas, vacías incluidas,
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
refs       docs/bugs/BUG-02.md
docs       -
files      -
ext        trello.card=5f2a8c1e3b9d4a7f6e0c2b81

## Description
The diff compares byte by byte and marks as different two lines that only
differ in the line ending.

## Acceptance Criteria
- [x] #1 The diff ignores CRLF
- [ ] #3 There is a test that covers it

## Definition of Done
- [ ] #1 Reviewed by someone else

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

**La línea `lease` sale solo cuando la tarea tiene arrendamiento**, y entonces sale con sus campos:
`lease` es `leaseExpiresAt`, con el mismo formato de instante que `created` y `updated`, y `holder` es
`leaseHolder` (["El vaciado"](../lease.md#el-vaciado) de `lease.md`). Los dos aparecen y desaparecen juntos, porque esa misma regla no admite uno sin
el otro. Pertenece al bloque de metadatos, así que la trae `--section meta` y no ninguna otra sección.
Es la única línea condicional de ese bloque, y por eso va al final de las líneas de dos campos: así
ninguna de las de arriba cambia de sitio según la tarea. Eso no choca con la regla de que la ficha
completa imprime las nueve secciones aunque estén vacías, porque lo condicional es una línea del bloque
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

Es el objeto de `task.list` más los campos del cuerpo:

```json
{
  "schemaVersion": 1,
  "kind": "task.get",
  "generatedAt": "2026-09-06T13:31:09Z",
  "data": {
    "task": {
      "id": "MYP-11",
      "description": "The diff compares byte by byte...",
      "acceptanceCriteria": [ { "key": 1, "text": "The diff ignores CRLF", "checked": true },
                              { "key": 3, "text": "There is a test that covers it", "checked": false } ],
      "definitionOfDone": [ { "key": 1, "text": "Reviewed by someone else", "checked": false } ],
      "plan": "1. Read the parser.\n2. Add the CRLF case.",
      "notes": "The parser already normalized LF, CRLF was missing.",
      "summary": null,
      "comments": [ { "key": 1, "author": "@avilches", "createdAt": "2026-09-06T10:02:11Z", "body": "A user with a Windows clone..." } ],
      "question": null,
      "blocks": ["MYP-40"],
      "urgencyBreakdown": { "priority": 6.0, "active": { "value": 4.0, "reason": null }, "blocking": 8.0,
                            "blocked": 0.0, "due": 0.0, "criteria": 1.0, "age": 0.0 }
    }
  }
}
```

**`urgencyBreakdown` solo sale con `--explain-urgency`**, igual que el desglose de la salida de texto, y
el ejemplo de arriba es el de una llamada que la lleva. Es la única clave de todo el documento que un
flag añade, y la excepción a la regla de las claves siempre presentes está declarada en la sección ["Números, fechas y ausencias"](../contrato-json.md#números-fechas-y-ausencias), junto
con la otra cosa que `biso get` hace con sus flags: recortar `data.task` con `--section`.

`urgencyBreakdown.active` es el único término que no es un número suelto: `value` es el número que
entra en la suma, el producto del coeficiente por el factor, igual que en los demás términos.
`reason` vale `null` cuando el término contribuye, y cuando contribuye `0.0` dice por qué:
`"not_active"` si el estado no es el activo, y `"waiting"` si lo es pero hay una pregunta abierta.

Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas. Con varias
coincidencias, `kind` es `task.candidates`, `data.tasks` es la lista y el código es 5.

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
text of the criteria and of the definition of done, the body of the comments,
the body of the open question and the labels. A match in the title always
wins over a match anywhere else. `biso ls --search` uses this same scope.

Options:
      --id                   force <ref> to be read as an id
      --match                force <ref> to be read as free text
      --section <name>       print only these sections; repeatable or comma
                             separated. One of: meta, desc, ac, dod, plan,
                             notes, summary, comments, question
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

