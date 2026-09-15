# Cómo se resuelve una referencia a una tarea

Todos los comandos que reciben `<ref>` usan exactamente esta rutina. No hay variantes por comando.

## La gramática

| Forma | Ejemplo | Interpretación |
|---|---|---|
| `PREFIX-<n>` | `MYP-11` | identificador, sin distinguir mayúsculas en el prefijo |
| `<n>` | `11` | identificador, con el prefijo del tablero |
| `#<n>` | `#11` | igual que el anterior |
| cualquier otra cosa | `"CRLF"` | consulta de texto |

Estos flags fuerzan la interpretación, y valen en todos los comandos que aceptan una referencia:

- `--id` obliga a interpretar como identificador. Con un valor que no encaje en la gramática, error 2.
- `--match` obliga a interpretar como texto, y sirve para buscar una tarea que se llame "42".

## La búsqueda por texto

**Hay un solo ámbito de búsqueda de texto en todo el programa**, y es el que usan tanto la resolución
de una referencia como el filtro `--search` de `biso ls` y `biso export`. Busca, sin distinguir
mayúsculas ni acentos, en:

el título, la descripción, el plan, las notas, el resumen final, el texto de los criterios de
aceptación, el texto de la definición de hecho, el cuerpo de los comentarios, el cuerpo de la
pregunta abierta y las etiquetas.

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
| 0 | error 4 |
| más de 1 | error 5, con las candidatas por stdout en el formato de `biso ls` |

El caso de más de una coincidencia lleva, por stderr y antes de esas filas, esta línea, con `code`
igual a `ambiguous_reference` (["El contrato JSON"](contrato-json.md#los-identificadores-de-error)):

```
error: "CRLF" matches 3 tasks
```

**Una coincidencia en el título gana sobre una coincidencia en cualquier otro sitio.** Si el texto
aparece en el título de una sola tarea, esa es la respuesta aunque aparezca en el cuerpo de otras
diez, y no hay ambigüedad. La búsqueda para resolver una referencia mira solo las tareas **no
archivadas**; el filtro `--search` mira las que digan los demás filtros.

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

