# Cómo se resuelve una referencia a una tarea

Todos los comandos que reciben `<ref>` usan exactamente esta rutina. No hay variantes por comando.

## La gramática

| Forma | Ejemplo | Interpretación |
|---|---|---|
| `PREFIX-<n>` | `MYP-11` | identificador, sin distinguir mayúsculas en el prefijo |
| `<n>` | `11` | identificador, con el prefijo del tablero |
| `#<n>` | `#11` | igual que el anterior |
| cualquier otra cosa | `"CRLF"` | consulta de texto |

Estos flags fuerzan la interpretación, y valen sobre el `<ref>` posicional de todos los comandos que
lo llevan así (`biso get <ref>`, `biso set <ref>...` y los demás verbos del ciclo):

- `--id` obliga a interpretar como identificador. Con un valor que no encaje en la gramática, error 2.
- `--match` obliga a interpretar como texto, y sirve para buscar una tarea que se llame "42".

**No hay forma de forzar la interpretación cuando la referencia es el valor de un flag**, como
`--parent <ref>` en `biso ls`: ahí decide solo la gramática de arriba, sin `--id` ni `--match` propios.
Es una limitación conocida, sin consecuencia práctica salvo para una tarea cuyo título sea puramente
numérico y se quiera referenciar como padre por ese título en vez de por su identificador.

## La búsqueda por texto

**Hay un solo ámbito de búsqueda de texto en todo el programa**, y es el que usan tanto la resolución
de una referencia como el filtro `--search` de `biso ls` y `biso export`. Busca, sin distinguir
mayúsculas ni acentos, en:

el título, la descripción, el plan, las notas, el resumen final, el texto de los criterios de
aceptación, el cuerpo de los comentarios, el cuerpo de la pregunta abierta y las etiquetas.

No busca en los identificadores, ni en las referencias, ni en la documentación, ni en los campos de
extensión.

**Alcanzar el cuerpo de la pregunta abierta tiene dos consecuencias, y ambas se aceptan a
propósito.** La primera es que la resolución de una referencia por texto también llega ahí, así que
`biso get "CRLF"` puede resolver a una tarea porque ese texto está en su pregunta. La segunda es que
una pregunta puede crear una ambigüedad de código 5 donde antes no la había. Lo contrario sería peor:
que el texto de una pregunta solo se pudiera encontrar al dejar de estar abierta, cuando se convierte
en comentario, y no mientras espera respuesta.

Cuando se usa para resolver una referencia, y solo entonces, se aplican además estas reglas:

| Coincidencias | Qué pasa |
|---:|---|
| exactamente 1 | se usa esa tarea, con `note: "CRLF" matched MYP-11` por stderr |
| 0 | error 4, con `code` igual a `not_found`: `error: no task matches "CRLF"` |
| más de 1 | error 5, con las candidatas por stdout en el formato de `biso ls` |

**Dos coincidencias en el título son la ambigüedad, y los cuerpos no la agrandan.** La regla de que
el título gana no es solo un desempate para el caso de una: si el texto aparece en el título de dos
tareas, las candidatas del error 5 son esas dos y ninguna más, aunque el mismo texto esté además en
el cuerpo de otras diez. Solo cuando no aparece en ningún título se buscan las candidatas en el
resto del ámbito.

El caso de más de una coincidencia lleva, por stderr y antes de esas filas, esta línea, con `code`
igual a `ambiguous_reference` (["El contrato JSON"](contrato-json.md#los-identificadores-de-error)):

```
error: "CRLF" matches 3 tasks
```

**Las candidatas se imprimen en todos los comandos que resuelven una referencia**, no solo en
`biso get`: `biso set`, los verbos del ciclo y el `-p` de `biso ls` las sacan igual, en el mismo
formato y con el mismo orden por defecto, porque esta rutina no tiene variantes por comando. Con
`--json`, en cambio, solo `biso get` contesta el sobre de datos `task.candidates`; los demás
contestan el sobre de error con el mismo código 5, según la tabla de `kind` de ["El sobre"](contrato-json.md#el-sobre).

**Una coincidencia en el título gana sobre una coincidencia en cualquier otro sitio.** Si el texto
aparece en el título de una sola tarea, esa es la respuesta aunque aparezca en el cuerpo de otras
diez, y no hay ambigüedad. La búsqueda para resolver una referencia mira solo las tareas **no
archivadas**; el filtro `--search` mira las que digan los demás filtros.

**El único filtro que se aplica al resolver una referencia es ese, el de archivada, y el estado no
filtra nada.** Una tarea en el estado terminal se resuelve como cualquier otra y aparece entre las
candidatas de un error 5, aunque `biso ls --search "<texto>"` no la traiga: ese listado la deja
fuera por el valor por defecto de su `-s` (["`biso ls`"](cmd/ls.md)), que es un filtro suyo y no una regla de esta
rutina. Resolver una referencia tiene que poder llegar a cualquier tarea del tablero, y una tarea
terminada se lee, se comenta y se reabre igual que las demás; si las candidatas la escondieran,
`biso get "CRLF"` diría que el texto encaja con una sola tarea mientras la lista de candidatas de un
error 5 posterior contradiría esa cuenta. Cuando la tabla de ["`biso get`"](cmd/get.md#comportamiento-caso-a-caso) dice que las
candidatas salen "exactamente" como las imprimiría ese listado, habla de la forma de imprimirlas, que
son tres cosas y solo tres: el orden, el límite de treinta y el aviso de recorte.

## Los tres mensajes de "no la encuentro"

**Identificador mal formado**, código 2, `code` igual a `malformed_id`:

```
error: malformed task id: "MYP-1.1"
hint: ids look like MYP-11 or 11. A subtask is an ordinary task with --parent MYP-1
```

**Identificador bien formado que el tablero nunca ha llegado a asignar**, código 4, `code` igual a
`never_allocated`:

```
error: MYP-999 has never existed on this board
note: the highest id ever assigned here is MYP-90
```

**Identificador que el tablero asignó alguna vez y que ahora no está**, código 4, `code` igual a
`not_found`:

```
error: MYP-53 is not on this board
note: MYP-53 was assigned at some point, but this board's current data does not have it
hint: this only happens when something outside biso touched the data, such as a
      snapshot restored over a newer one or a database edited by hand;
      `biso doctor` diagnoses damage to the board
```

Los códigos de estos tres casos son distintos entre sí: el primero sale con 2, y los otros dos comparten
el código 4 pero llevan un `code` distinto.

---

