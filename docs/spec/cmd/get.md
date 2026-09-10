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

`--section` sirve para pedir solo una parte. `biso get TASK-11 --section ac` imprime los criterios con
sus claves y cuesta unas decenas de bytes en vez de la ficha entera, que es lo que hace falta antes de
marcar uno.

## Comportamiento, caso a caso

La resolución de `<ref>` está en la sección 7 y no se repite. Lo propio de este comando:

| Caso | Qué pasa |
|---|---|
| La referencia resuelve a una tarea | Se imprime, código 0 |
| La referencia es texto y encaja con varias | Error 5, y las candidatas salen **por stdout** en el formato de `biso ls` |
| La referencia es texto y encaja con una | Se imprime, con `note: "CRLF" matched TASK-11` por stderr |
| La tarea está archivada | Se imprime, con `note: TASK-11 is archived` por stderr |
| La tarea no se puede leer | Error 3, según la regla de lectura dirigida de 4.12 |
| `--section` con un nombre inventado | Error 2, con los nueve nombres válidos |
| `--section` de una sección vacía | No imprime esa sección, y si no queda ninguna sección que imprimir, la salida está vacía y el código sigue siendo 0 |

**Sin `--section`, la ficha completa imprime siempre las nueve secciones fijas, vacías incluidas,
marcadas con `(empty)`.** Es solo con `--section` que una sección vacía se omite en vez de imprimirse
vacía; sin la bandera, omitir una sección la confundiría con una que no se ha pedido.

## Salida

```
TASK-11  Normalize CRLF in the diff
status     In Progress          type       bug
priority   high                 urgency    19.0
assignees  @claude              reporter   @avilches
labels     parser               milestone  -
parent     -                    due        -
project    -                    ordinal    -
created    2026-09-06 09:12     updated    2026-09-06 11:40
depends    -                    blocks     TASK-40
lease      2026-09-06 15:40     holder     @claude
refs       docs/bugs/BUG-02.md
docs       -
files      -
ext        trello.card=5f2a8c1e3b9d4a7f6e0c2b81

## Description
El diff compara byte a byte y marca como distintas dos lineas que solo difieren
en el fin de linea.

## Acceptance Criteria
- [x] #1 El diff ignora el CRLF
- [ ] #3 Hay un test que lo cubre

## Definition of Done
- [ ] #1 Revisado por otra persona

## Implementation Plan
1. Leer el parser.
2. Anadir el caso CRLF.

## Implementation Notes
El parser ya normalizaba LF, faltaba CRLF.

## Final Summary
(empty)

## Comments
@avilches, 2026-09-06 10:02
Esto lo reporto un usuario con un repositorio clonado en Windows.

## Open Question
(empty)
```

Los encabezados de esta salida son un formato de presentación, no un formato de almacenamiento.

**La línea `lease` sale solo cuando la tarea tiene arrendamiento**, y entonces sale con sus dos campos:
`lease` es `leaseExpiresAt`, con el mismo formato de instante que `created` y `updated`, y `holder` es
`leaseHolder` (sección 5). Los dos aparecen y desaparecen juntos, porque la sección 5 no admite uno sin
el otro. Pertenece al bloque de metadatos, así que la trae `--section meta` y no ninguna otra sección.
Es la única línea condicional de ese bloque, y por eso va al final de las líneas de dos campos: así
ninguna de las de arriba cambia de sitio según la tarea. Eso no choca con la regla de que la ficha
completa imprime las nueve secciones aunque estén vacías, porque lo condicional es una línea del bloque
y no el bloque. Una tarea sin arrendamiento **no imprime la línea**, en vez de imprimirla con dos
guiones, porque eso pondría dos guiones en la ficha de casi todas las tareas del tablero y la ausencia
de la línea dice lo mismo. Esta es la única forma de ver los dos campos sin `--json`: `biso prime` no
los trae (9.7) y `biso ls` tampoco (10.4). Si el arrendamiento está vencido, el instante ya lo dice y la
ficha no añade ninguna marca; el derivado `leaseExpired` ya calculado está en `--json`.

Con `--section ac`, solo el encabezado con el identificador y el título, y la sección pedida:

```
TASK-11  Normalize CRLF in the diff

## Acceptance Criteria
- [x] #1 El diff ignora el CRLF
- [ ] #3 Hay un test que lo cubre
```

Con `--section question` sobre TASK-60, la tarea con la pregunta abierta del ejemplo de 9.7, la
sección sale rellena con la misma forma que ya usa `## Comments`: el autor y el instante en una línea
y el cuerpo debajo.

```
TASK-60  Confirm the retry budget for the upload endpoint

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
abierta (5.7); en cualquier otro caso vale `0.00`, y la etiqueta dice cuál de los dos motivos se
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
      "id": "TASK-11",
      "description": "El diff compara byte a byte...",
      "acceptanceCriteria": [ { "key": 1, "text": "El diff ignora el CRLF", "checked": true },
                              { "key": 3, "text": "Hay un test que lo cubre", "checked": false } ],
      "definitionOfDone": [ { "key": 1, "text": "Revisado por otra persona", "checked": false } ],
      "plan": "1. Leer el parser.\n2. Anadir el caso CRLF.",
      "notes": "El parser ya normalizaba LF, faltaba CRLF.",
      "summary": null,
      "comments": [ { "author": "@avilches", "createdAt": "2026-09-06T10:02:11Z", "body": "Esto lo reporto..." } ],
      "question": null,
      "blocks": ["TASK-40"],
      "urgencyBreakdown": { "priority": 6.0, "active": { "value": 4.0, "reason": null }, "blocking": 8.0,
                            "blocked": 0.0, "due": 0.0, "criteria": 1.0, "age": 0.0 }
    }
  }
}
```

**`urgencyBreakdown` solo sale con `--explain-urgency`**, igual que el desglose de la salida de texto, y
el ejemplo de arriba es el de una llamada que la lleva. Es la única clave de todo el documento que una
bandera añade, y la excepción a la regla de las claves siempre presentes está declarada en 12.4, junto
con la otra cosa que `biso get` hace con sus banderas: recortar `data.task` con `--section`.

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
| El almacén no responde | 7 |
| No hay tablero | 8 |

## `biso get --help`

```
Usage: biso get <ref> [options]

Show one task. <ref> is an id (TASK-11), a bare number (11) or free text
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
  7  the board could not respond
  8  no board here

Examples:
  biso get TASK-11
  biso get 11 --section ac
  biso get "CRLF"
  biso get TASK-11 --explain-urgency
```

---

