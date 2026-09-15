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

**No hay coincidencia por prefijo ni por parecido.** `coincidir()` nunca sugiere nada: o
encuentra una coincidencia exacta o normalizada, o es error 3. Lo que sí sugiere candidatos
parecidos es un algoritmo aparte, para el mensaje de error de las demás reglas de la
especificación que lo prometen.

## El algoritmo de sugerencias más parecidas

Cinco sitios de la especificación prometen, en su mensaje de error, una lista de los valores
más parecidos a lo que se tecleó: `--label` y `--label-or` (hasta cinco etiquetas, más abajo en
esta misma sección), `--assignee` (hasta cinco personas, igual), `biso config` con una clave
inexistente (hasta tres claves, sección ["`biso config`"](cmd/config.md#comportamiento-caso-a-caso)) y `biso help` con uno o
varios nombres de comando, si alguno no existe (hasta tres nombres, sección ["La ayuda"](cmd/help.md)). Estos casos
comparten un solo algoritmo, con el mismo espíritu que `coincidir()`: una sola regla, y solo
cambia el tope `N` según el caso.

**La métrica es la distancia de Levenshtein, sobre la forma normalizada.** Dadas dos cadenas `a`
y `b`, la distancia de Levenshtein es el número mínimo de inserciones, eliminaciones o
sustituciones de un carácter que hace falta para convertir `a` en `b`. Se calcula con la matriz
de programación dinámica habitual: con `m = longitud(a)` y `n = longitud(b)`, una tabla `d` de
`(m+1) × (n+1)` donde `d[i][0] = i`, `d[0][j] = j` para todo `i, j`, y para `i, j > 0`:

```
d[i][j] = d[i-1][j-1]                                si a[i] == b[j]
d[i][j] = 1 + min(d[i-1][j], d[i][j-1], d[i-1][j-1])  si a[i] != b[j]
```

El resultado es `d[m][n]`.

```
sugerir(v, candidatos, N):
  1. calcular normalizar(v), la misma normalizacion que usa coincidir()
  2. para cada candidato c, calcular normalizar(c) y la distancia de Levenshtein
     entre normalizar(v) y normalizar(c)
  3. descartar los candidatos cuya distancia supere el umbral: la mitad de la
     longitud de normalizar(v), redondeada hacia arriba
  4. ordenar lo que queda por distancia ascendente; en caso de empate, por
     orden alfabetico de la forma normalizada
  5. devolver como maximo los N primeros que queden
```

**Si ningún candidato pasa el umbral, el resultado es una lista vacía, no un error distinto.**
El mensaje de error se queda con lo que ya dice sin sugerencia, igual que cuando el vocabulario
contra el que se sugiere está vacío (por ejemplo, un tablero sin ninguna etiqueta): forzar `N`
sugerencias cuando nada se parece de verdad haría más probable que se tomara una sin
comprobarla que ayudar a corregir el error.

Un ejemplo con las etiquetas `api, backend, bug, docs, frontend, infra, parser, security, ui,
urgent`:

| Entrada | `normalizar` | Umbral (mitad de la longitud, hacia arriba) | Qué pasa el umbral | Sugerencia (máximo 5) |
|---|---|---|---|---|
| `fronted` | `fronted` (7) | 4 | `frontend`, distancia 1 | `frontend` |
| `xyz` | `xyz` (3) | 2 | ninguna: la más cercana, `api`, está a distancia 3 | ninguna, lista vacía |

**El orden de salida es siempre por cercanía, nunca alfabético puro**, salvo para romper un
empate entre dos candidatos a la misma distancia.

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
| `--label` y `--label-or` | el conjunto de etiquetas del tablero, definido abajo | error 3, con hasta cinco de las más parecidas (las que pasen el umbral de ["El algoritmo de sugerencias más parecidas"](#el-algoritmo-de-sugerencias-más-parecidas)) |
| `--assignee` | el conjunto de personas del tablero, definido abajo | error 3, con hasta cinco de las más parecidas (mismo algoritmo) |
| `--parent` | la resolución de referencias de la sección ["Cómo se resuelve una referencia a una tarea"](referencias.md) | error 2, 4 o 5 |
| `--search` | nada, es texto libre | nunca falla |

**El conjunto de etiquetas del tablero** es la unión de las etiquetas declaradas en la clave `labels`
de la configuración y de las que lleva cualquier tarea del tablero, **incluidas las archivadas y las
que están en el estado terminal**. **El conjunto de personas se define con la clave `assignees` de la
configuración y con los valores de `assignees` de cualquier tarea, archivadas y terminadas incluidas.
Los valores de `author` no entran en este conjunto**, porque no hay ningún filtro `--author`: una
persona que solo consta como autora de la tarea y nunca la ha tenido asignada no pertenece al
conjunto contra el que valida `--assignee`.

Un tablero con las etiquetas `frontend`, `backend`, `bug`, `docs` y `parser`, y las personas
`@claude`, `@sara` y `@avilches`:

```
$ biso ls -l fronted
error: unknown label: "fronted"
hint: did you mean: frontend?

$ biso ls -a @clude
error: unknown assignee: "@clude"
hint: did you mean: @claude?
```

**Ni las etiquetas ni las personas tienen vocabulario cerrado al escribir.** Escribir una etiqueta
nueva la incorpora al conjunto, y a partir de ese momento filtrar por ella funciona.

Está el flag `--unchecked` de `biso ls` y `biso export`, que apaga **las comprobaciones contra
estos conjuntos, las de etiquetas y personas, y ninguna otra**: los vocabularios configurados de
`--status`, `--type` y `--priority` siguen validando, y `--parent` sigue resolviendo su referencia.
El flag no cambia ninguna otra cosa.

---

