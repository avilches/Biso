# Las familias de banderas

Esta sección define de una vez la forma de todas las banderas de escritura. Los comandos no la
repiten: cada uno dice qué campos acepta, y esta sección dice qué forma tiene cada campo.

## La regla

**El nombre desnudo añade. `set-` delante sustituye. `rm-` delante quita uno. `clear-` delante
vacía.** La forma de una bandera se deduce siempre del nombre del campo, sin nombres propios y sin
que haya que consultar nada.

Las cuatro variantes existen para **todo campo que guarde una lista de elementos**. Cada clase de
campo tiene estas variantes:

| Clase de campo | Variantes |
|---|---|
| Lista de elementos | las cuatro |
| Bloque de prosa | añadir, sustituir, vaciar |
| Mapa de claves | fijar una clave, quitar una clave, vaciar |
| Escalar | fijar, vaciar |
| Lista inmutable (comentarios) | solo añadir, `--comment` |

**`question` no entra en esta tabla.** Es un registro de tres partes (["La pregunta abierta"](modelo-de-datos.md#la-pregunta-abierta)), no una
lista, ni un bloque de prosa, ni un mapa, ni un escalar, así que ninguna de estas clases lo describe.
**Ninguna bandera de campo escribe `question`**: lo escriben `biso ask`, `biso answer` y la importación
de `biso new --from`, y nadie más, igual que `archived` solo lo cambia `biso archive` (sección
["El modelo de datos de una tarea"](modelo-de-datos.md)).

**El significado no cambia entre comandos.** `--ac` añade un criterio en `biso new`, en `biso set`, en
`biso start` y en `biso finish`, y todos los comandos de escritura aceptan todas estas banderas.

La regla tiene **dos desviaciones de nombre en todo el programa**, y las dos son de forma: ninguna
cambia lo que la bandera hace, pero en las dos el nombre no se deduce entero del campo. La primera es
el número gramatical de las notas: el añadido se llama `--note`, en singular, y su sustitución se llama
`--set-notes`, en plural. La segunda es el sufijo de las marcas: marcar un criterio de aceptación es
`--check` y marcar un elemento de la definición de hecho es `--check-dod`, o sea que el nombre desnudo
está reservado para los criterios y solo la otra lista lleva el sufijo del campo, aunque las demás
banderas de las dos lo lleven siempre (`--ac` y `--dod`, `--rm-ac` y `--rm-dod`). Lo mismo vale para
`--uncheck` frente a `--uncheck-dod`.

## Campos de lista

| Campo | Añade | Sustituye | Quita | Vacía | Acepta lista por comas |
|---|---|---|---|---|---|
| etiquetas | `-l, --label` | `--set-label` | `--rm-label` | `--clear-label` | sí |
| personas asignadas | `-a, --assignee` | `--set-assignee` | `--rm-assignee` | `--clear-assignee` | sí |
| referencias | `--ref` | `--set-ref` | `--rm-ref` | `--clear-ref` | sí |
| documentación | `--doc` | `--set-doc` | `--rm-doc` | `--clear-doc` | sí |
| dependencias | `--dep` | `--set-dep` | `--rm-dep` | `--clear-dep` | sí |
| ficheros tocados | `--file` | `--set-file` | `--rm-file` | `--clear-file` | sí |
| criterios de aceptación | `--ac` | `--set-ac` | `--rm-ac` | `--clear-ac` | **no** |
| definición de hecho | `--dod` | `--set-dod` | `--rm-dod` | `--clear-dod` | **no** |

Todas las de "añade" y "sustituye" son repetibles. `--rm-ac` y `--rm-dod` toman un selector de la
sección ["Selectores de criterios"](#selectores-de-criterios).

**`--set-ac` y `--set-dod` crean elementos nuevos, con claves nuevas y sin marcar**, y las claves de
los elementos anteriores no se reutilizan. Es coherente con
["Los criterios y sus claves estables"](modelo-de-datos.md#los-criterios-y-sus-claves-estables): la
clave se asigna al crear el elemento, y sustituir la lista crea elementos.

## Campos de prosa

| Campo | Añade al final | Sustituye | Vacía |
|---|---|---|---|
| descripción | `-d, --desc` | `--set-desc` | `--clear-desc` |
| plan | `--plan` | `--set-plan` | `--clear-plan` |
| notas | `--note` | `--set-notes` | `--clear-notes` |
| resumen final | `--summary` | `--set-summary` | `--clear-summary` |

- Añadir a un campo vacío es lo mismo que fijarlo, así que al crear una tarea las columnas "Añade al
  final" y "Sustituye" coinciden y no hay nada que decidir.
- Al añadir sobre contenido existente se intercala una línea en blanco, y cada repetición de la
  bandera en la misma invocación produce su propio párrafo.
- Añadir un valor vacío no hace nada y avisa, según ["El valor vacío"](valores-de-entrada.md#el-valor-vacío).

## Selectores de criterios

`--check`, `--uncheck`, `--rm-ac`, `--check-dod`, `--uncheck-dod` y `--rm-dod` toman un selector.
Todos son repetibles.

| Selector | Ejemplo | Qué elige |
|---|---|---|
| `all` | `--check all` | todos los elementos de esa lista en esa tarea |
| una clave | `--check 3` | el elemento `#3` |
| un rango de claves | `--check 1-4` | las claves de la 1 a la 4 que existan |
| varias claves | `--check 1,3,7` | esas tres |
| texto | `--check "cubre CRLF"` | el elemento cuyo texto contenga ese fragmento |

**La regla de desambiguación, que hay que implementar tal cual.** El valor se trata como lista de
claves **solo si el valor entero** encaja con `^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$`. En cualquier otro
caso es un texto literal, comas incluidas. Así, `--check "1, 2 y el ultimo"` es una búsqueda de texto
que no encontrará nada y dará error 4, en vez de convertirse en algo a medias.

| Caso límite | Resultado |
|---|---|
| clave que no existe | error 4: `no acceptance criterion #7 on TASK-11 (keys: 1, 3)` |
| texto que no encaja con ninguno | error 4, con los textos de los elementos listados |
| texto que encaja con dos | error 5, con los dos listados |
| rango donde faltan claves intermedias | se aplican las que hay, sin aviso |
| rango invertido, `4-1` | error 2 |
| marcar un elemento ya marcado | se queda marcado, sin aviso, la operación es idempotente |
| `--check all` en una tarea sin criterios | sin efecto, con `warning: TASK-11 has no acceptance criteria` |
| `--check all` sobre varias tareas | válido, cada tarea marca los suyos |
| una clave, un rango, una lista o un texto sobre varias tareas | error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |

**La regla de solape se aplica sobre el conjunto ya resuelto, no sobre el texto del selector.** Si
después de resolver `--check` y `--uncheck` un mismo elemento aparece en los dos conjuntos, es error
2, y da igual que se haya escrito `--check 3 --uncheck 3` o `--check all --uncheck 3`:

```
error: --check and --uncheck both select acceptance criterion #3 of TASK-11
```

## Campos escalares

| Campo | Fija | Vacía |
|---|---|---|
| título | `-t, --title` | no se puede, es obligatorio |
| estado | `-s, --status` | no se puede, es obligatorio |
| tipo | `--type` | `--clear-type` |
| prioridad | `--priority` | `--clear-priority` |
| proyecto | `--project` | `--clear-project` |
| hito | `-m, --milestone` | `--clear-milestone` |
| tarea padre | `-p, --parent` | `--clear-parent` |
| fecha límite | `--due` | `--clear-due` |
| orden manual | `--ordinal` | `--clear-ordinal` |
| persona que reporta | `--reporter` | `--clear-reporter` |

Un escalar **nunca** se borra pasándole la cadena vacía, según ["El valor vacío"](valores-de-entrada.md#el-valor-vacío).

## Campos externos

| Operación | Bandera | Repetible |
|---|---|---|
| fijar una clave | `--ext <clave>=<valor>` | sí |
| quitar una clave | `--rm-ext <clave>` | sí |
| vaciar el mapa entero | `--clear-ext` | no |

**No existe `--set-ext`.** Fijar una clave con `--ext` ya sustituye su valor. Vaciar el mapa entero es
`--clear-ext`, y es la única forma de vaciarlo.

---

