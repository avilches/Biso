# Cómo se pasa un valor

## Tres formas de pasar un valor largo

Todo parámetro de tipo texto largo (`--append-desc`, `--append-plan`, `--append-note`,
`--append-summary`, `--comment` y el texto de un criterio o de un elemento de la definición de hecho)
acepta las tres:

| Forma | Significado |
|---|---|
| `--append-desc "texto"` | el texto literal |
| `--append-desc @ruta/fichero.md` | el contenido del fichero, interpretado como UTF-8 |
| `--append-desc -` | todo lo que llegue por la entrada estándar hasta el fin de fichero |

Reglas:

- **Un texto que empieza de verdad por `@` se escribe `@@`.** El primer `@` se descarta y el resto es
  literal. Es la única secuencia de escape del programa.
- **Los campos de persona nunca interpretan el `@`.** `--assignee`, `--reporter` y `--comment-author`
  toman su valor tal cual, así que `--comment-author @trello:juan` guarda ese texto y no intenta leer
  ningún fichero.
- **`-` solo puede aparecer una vez por invocación.** Dos parámetros que pidan la entrada estándar son
  un error de uso con código 2, porque el segundo leería un flujo agotado y guardaría el vacío sin
  que se note.
- **Un fichero que no existe es código 4**, con el mensaje `error: --append-desc: file not found: docs/x.md`.
  Un fichero que existe pero no se puede leer es código 7.
- **Un valor vacío, venga de donde venga, no borra nada.** Ver ["El valor vacío"](#el-valor-vacío).

## El valor vacío

Un valor vacío es una cadena sin ningún carácter, o solo con espacios, tanto si llega literalmente
como si llega de un fichero vacío o de una entrada estándar vacía. La regla es única:

| Dónde | Qué pasa |
|---|---|
| En una bandera que añade (`--append-note`, `--add-labels`, `--add-ac`, `--append-desc`) | No se añade nada, se emite `warning: --append-note: empty value, nothing was added` y el código sigue siendo 0 |
| En una bandera que sustituye (`--replace-labels`) | Deja el campo vacío, igual que `--clear-labels`. Sustituir por nada es vaciar, y eso sí es explícito |
| En un campo escalar (`--type ""`, `--priority ""`) | Error 3. **La cadena vacía nunca es la forma de borrar un escalar**; para eso está `--clear-type` |
| En el título, al crear | Error 2: `error: title cannot be empty` |

**El `code` de un escalar vacío depende de si el campo tiene vocabulario cerrado.** Para `status`,
`type` y `priority`, una cadena vacía es un valor que no coincide con nada configurado, así
que sigue la regla del ["algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia) y el
`code` es el de un valor desconocido (`unknown_status` y análogos, con el mensaje de
["El mismo texto vale lo mismo en los dos sentidos"](vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos)).
Para los demás escalares (`--reporter ""`, `--ordinal ""`, `--due ""`), que no
tienen vocabulario, el `code` es `empty_scalar_value`.

## Valores que empiezan por guion

Tres mecanismos, en orden de preferencia:

1. **`--flag=valor`** funciona siempre y es la forma recomendada: `--append-desc=-5 grados`.
2. **`--`** termina el análisis de opciones: `biso new -- "-n no es una bandera"`.
3. **Un valor que empieza por guion detrás de una bandera que exige valor se acepta tal cual**, sin
   heurísticas. `biso set MYP-1 --append-note -x` guarda `-x` como nota.

Como consecuencia de la regla 3, olvidar el valor de una bandera se detecta por lo que sobra después,
no por lo que parece: `biso set MYP-1 --append-note --priority high` guarda la nota `--priority` y
luego falla con código 2 y `error: unexpected argument: high`.

## Repetición y listas separadas por comas

Para toda bandera marcada como repetible:

- Repetirla acumula: `--add-labels a --add-labels b` deja dos etiquetas.
- Si además acepta lista, separar por comas acumula igual: `--add-labels a,b` deja las mismas dos.
- Las dos formas se pueden mezclar.
- **Una coma dentro de un valor se escapa con `\,`.** Es la única forma de meter una coma en una
  referencia, en una documentación o en un fichero tocado. Una etiqueta, una persona asignada o una
  clave de `ext` nunca llevan coma, así que en ninguno de esos campos hay nada que escapar
  (["El juego de caracteres de un token"](#el-juego-de-caracteres-de-un-token)).
- Los campos de texto largo y los criterios **nunca** se parten por comas.
- Un valor repetido dentro de la misma bandera se guarda una vez y produce
  `warning: --add-labels: "urgent" given twice, kept once`.

Para toda bandera **no** repetible, es decir, los campos escalares, pasarla dos veces con valores
distintos es un error de uso con código 2:

```
error: --status given twice with different values: "In Progress" and "Done"
```

## El juego de caracteres de un token

`labels`, `assignees` y las claves de `ext` (["Campos externos"](modelo-de-datos.md#los-campos-externos)) son los únicos
campos de esta sección cuyo alfabeto está cerrado. Los demás campos de lista de la tabla de
["Campos de lista que admiten coma"](familias-de-banderas.md#campos-de-lista-que-admiten-coma), es decir `references`, `documentation`,
`dependencies` y `modifiedFiles`, son texto libre y no tienen ninguna restricción de caracteres: una
referencia o una documentación pueden ser una URL, y un fichero tocado es una ruta, y ninguna de las
dos cosas admite cerrarle el alfabeto sin dejar fuera casos legítimos. `dependencies` tampoco la
necesita: cada elemento es un `<ref>` y ya lo gobierna entera la gramática de ["Cómo se resuelve una referencia a una tarea"](referencias.md).

| Campo | Alfabeto |
|---|---|
| `labels`, `assignees` | letras y dígitos Unicode, y los símbolos `- _ . : @` |
| clave de `ext` | letras y dígitos Unicode, y los símbolos `- _ .` |

**Ninguno de los dos alfabetos admite el espacio.** Dos herramientas comparables que escriben una
etiqueta como palabra suelta de una línea de comandos, en vez de elegirla en un formulario web,
la prohíben: Taskwarrior exige que una etiqueta sea una sola palabra, y Jira rechaza directamente
cualquier etiqueta con espacio. GitHub sí permite etiquetas de varias palabras, pero nunca se
enfrenta a este problema porque una etiqueta de GitHub nunca se teclea suelta en una shell: se elige
en un desplegable o llega ya como cadena entrecomillada dentro de un JSON. La razón completa, con la
comparación entera, está en ["El juego de caracteres de un token"](../DECISIONES.md#el-juego-de-caracteres-de-un-token) de `DECISIONES.md`.

**La clave de `ext` no admite `@` ni `:`, porque no tienen ningún uso documentado ahí, ni tampoco
`=`, porque `--ext <clave>=<valor>` ya usa ese carácter para separar la clave del valor**: si se
permitiera dentro de la clave, `--ext a=b=c` sería ambiguo sobre dónde termina la clave.

**Un carácter fuera del alfabeto que le toca es error 2 (`USAGE`)**, en la misma familia que un
identificador mal formado (["Los tres mensajes de \"no la encuentro\""](referencias.md#los-tres-mensajes-de-no-la-encuentro)): es un problema de forma, no de que el
tablero no reconozca el valor, así que no es el código 3 de ["Los vocabularios del tablero y la regla de validación"](vocabularios.md), y de hecho ni `labels` ni
`assignees` tienen vocabulario cerrado al escribir (["Qué valida cada filtro, y contra qué"](vocabularios.md#qué-valida-cada-filtro-y-contra-qué)).

```
error: malformed label: "urgent!"
hint: a label may contain letters, digits, and - _ . : @

error: malformed assignee: "sara smith"
hint: an assignee may contain letters, digits, and - _ . : @

error: malformed extension key: "trello=card"
hint: an extension key may contain letters, digits, and - _ .
```

**Esto rige al escribir.** Un valor ya guardado que no cumple este alfabeto, porque se escribió antes
de que existiera esta regla o porque llegó por una vía que no pasa por esta validación, no es un error
nuevo distinto: es un dato que el programa no puede interpretar, y se trata con la regla general de
["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), la misma que ya cubre una clave de `ext` que la
configuración ha dejado de declarar.

Los `code` correspondientes, `malformed_label`, `malformed_assignee` y `malformed_extension_key`,
están en la tabla de ["Los identificadores de error"](contrato-json.md#los-identificadores-de-error).

