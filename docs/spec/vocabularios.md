# Los vocabularios del tablero y la regla de validación

Hay campos con vocabulario cerrado, definido en la configuración: `status`, `type` y `priority`. Para
todos ellos rige una sola regla, **idéntica al escribir y al leer**.

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

Un ejemplo con un tablero creado con los estados por defecto. La configuración de un tablero vive
dentro de su base de datos y no en un fichero que se edite a mano, así que se consulta con
["`biso config`"](cmd/config.md), y el trozo que importa aquí es este:

```
$ biso config list
project_name = My project
statuses = To Do,In Progress,Done
initial_status = To Do
active_status = In Progress
terminal_status = Done
...
```

Uno de sus tres estados es `To Do`. Con el algoritmo de arriba, estas entradas se resuelven así:

| Entrada | `normalizar` | Resultado |
|---|---|---|
| `To Do` | `todo` | coincide exactamente con `To Do`, por el paso a |
| `todo` | `todo` | coincide con `To Do`, por el paso c |
| `TODO` | `todo` | coincide con `To Do` |
| `To-Do` | `todo` | coincide con `To Do`, porque el guion se elimina |
| `TO_DO` | `todo` | coincide con `To Do`, porque el guion bajo se elimina |
| `to  do` | `todo` | coincide con `To Do`, porque los espacios se eliminan todos, sean uno o varios |
| `To Do.` | `todo.` | error 3: el punto no está entre los caracteres que se eliminan, así que `todo.` no es igual a `todo` ni a la forma normalizada de ningún otro estado (`inprogress`, `done`) |
| `To.Do` | `to.do` | error 3, por lo mismo: el punto se queda y `to.do` no coincide con ningún estado |

**No hay coincidencia por prefijo ni por parecido.**

## El mismo texto vale lo mismo en los dos sentidos

Esta tabla es el contrato, y es la prueba de aceptación que hay que poder ejecutar. Tablero con los
estados `To Do`, `In Progress` y `Done`:

| Entrada | `biso set MYP-1 -s <v>` | `biso ls -s <v>` |
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
| `--status`, `--type`, `--priority` | el vocabulario configurado | error 3 |
| `--label` y `--label-or` | el conjunto de etiquetas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--assignee` | el conjunto de personas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--parent` | la resolución de referencias de la sección ["Cómo se resuelve una referencia a una tarea"](referencias.md) | error 2, 4 o 5 |
| `--search` | nada, es texto libre | nunca falla |

**El conjunto de etiquetas del tablero** es la unión de las etiquetas declaradas en la clave `labels`
de la configuración y de las que lleva cualquier tarea del tablero, **incluidas las archivadas y las
que están en el estado terminal**. **El conjunto de personas se define con la clave `assignees` de la
configuración y con los valores de `assignees` de cualquier tarea, archivadas y terminadas incluidas.
Los valores de `author` no entran en este conjunto**, porque no hay ningún filtro `--author`: una
persona que solo consta como autora de la tarea y nunca la ha tenido asignada no pertenece al
conjunto contra el que valida `--assignee`.

**Ni las etiquetas ni las personas tienen vocabulario cerrado al escribir.** Escribir una etiqueta
nueva la incorpora al conjunto, y a partir de ese momento filtrar por ella funciona.

Está el flag `--unchecked` de `biso ls` y `biso export`, que apaga **las comprobaciones contra
estos conjuntos, las de etiquetas y personas, y ninguna otra**: los vocabularios configurados de
`--status`, `--type` y `--priority` siguen validando, y `--parent` sigue resolviendo su referencia.
El flag no cambia ninguna otra cosa.

---

