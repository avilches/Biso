# Los vocabularios del tablero y la regla de validación

Tres campos tienen vocabulario cerrado, definido en la configuración: `status`, `type` y `priority`.
Un cuarto, `project`, lo tiene solo si el tablero declara proyectos. Para todos ellos rige una sola
regla, **idéntica al escribir y al leer**.

## El algoritmo de coincidencia

Dado un valor de entrada `v` y la lista de valores configurados, el programa calcula así:

```
normalizar(x):
  1. pasar x a minusculas segun Unicode
  2. descomponer y quitar los diacriticos (acentos, dieresis, cedillas)
  3. eliminar TODOS los caracteres que sean espacio, tabulador, guion (-) o guion bajo (_)
  4. devolver lo que queda

coincidir(v, configurados):
  a. si existe un configurado c con c == v exactamente, devolver c
  b. si no, calcular normalizar(v) y compararlo con normalizar(c) de cada configurado
  c. si exactamente un configurado coincide, devolverlo
  d. si ninguno coincide, error 3
  e. si coinciden dos o mas, error 3 con los dos listados, porque el tablero
     tiene dos valores que se normalizan igual y hay que desambiguarlos
```

Con este algoritmo, y para un tablero cuyo estado es `To Do`:

| Entrada | `normalizar` | Resultado |
|---|---|---|
| `To Do` | `todo` | coincide, por el paso a |
| `todo` | `todo` | coincide |
| `TODO` | `todo` | coincide |
| `To-Do` | `todo` | coincide |
| `TO_DO` | `todo` | coincide |
| `to  do` | `todo` | coincide |
| `Pending` | `pending` | error 3 |
| `Todos` | `todos` | error 3 |

**No hay coincidencia por prefijo ni por parecido.**

## El mismo texto vale lo mismo en los dos sentidos

Esta tabla es el contrato, y es la prueba de aceptación que hay que poder ejecutar. Tablero con los
estados `To Do`, `In Progress` y `Done`:

| Entrada | `biso set TASK-1 -s <v>` | `biso ls -s <v>` |
|---|---|---|
| `To Do` | escribe | filtra |
| `todo` | escribe | filtra |
| `TO_DO` | escribe | filtra |
| `In-Progress` | escribe | filtra |
| `Pending` | error 3 | error 3 |
| `""` | error 3 | error 3 |

El mensaje es el mismo en los dos sentidos:

```
error: unknown status: "Pending"
       valid statuses on this board: To Do, In Progress, Done
```

## Qué valida cada filtro, y contra qué

| Filtro | Conjunto contra el que valida | Si no encaja |
|---|---|---|
| `--status`, `--type`, `--priority`, `--project` | el vocabulario configurado | error 3 |
| `--label` y `--label-or` | el conjunto de etiquetas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--assignee` | el conjunto de personas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--milestone` | el conjunto de hitos del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--parent` | la resolución de referencias de la sección 7 | error 2, 4 o 5 |
| `--search` | nada, es texto libre | nunca falla |

**El conjunto de etiquetas del tablero** es la unión de las etiquetas declaradas en la clave `labels`
de la configuración y de las que lleva cualquier tarea del tablero, **incluidas las archivadas y las
que están en el estado terminal**. **El conjunto de personas se define con la clave `assignees` de la
configuración y con los valores de `assignees` de cualquier tarea, archivadas y terminadas incluidas.
Los valores de `reporter` no entran en este conjunto**, porque no hay ningún filtro `--reporter`: una
persona que solo ha reportado tareas y nunca las ha tenido asignadas no pertenece al conjunto contra
el que valida `--assignee`.

**El conjunto de hitos del tablero es solo derivado**: son los valores de `milestone` que lleva
cualquier tarea del tablero, **archivadas y terminadas incluidas**, y nada más. Es el único de los
tres que no tiene mitad declarada, porque no existe ninguna clave `milestones` en la configuración
(10.10) ni ninguna bandera de `biso init` que la escriba, así que **un hito existe exactamente
mientras alguna tarea lo lleve escrito**. Un tablero en el que ninguna tarea tiene hito tiene el
conjunto vacío, y entonces cualquier `--milestone` es error 3; el mensaje lo dice tal cual, sin
sugerencias, porque no hay ninguna que ofrecer.

**Ni las etiquetas, ni las personas, ni los hitos tienen vocabulario cerrado al escribir.** Escribir
una etiqueta nueva la incorpora al conjunto, y a partir de ese momento filtrar por ella funciona. Con
el hito pasa lo mismo: `biso set TASK-1 -m "v1.2"` es lo que hace que `v1.2` exista para
`biso ls -m "v1.2"`, y la última tarea que deja de llevarlo lo saca del conjunto.

Está la bandera `--unchecked` de `biso ls` y `biso export`, que apaga **las tres comprobaciones contra
estos conjuntos, las de etiquetas, personas e hitos, y ninguna otra**: los vocabularios configurados
de `--status`, `--type`, `--priority` y `--project` siguen validando, y `--parent` sigue resolviendo
su referencia. La bandera no cambia ninguna otra cosa.

---

