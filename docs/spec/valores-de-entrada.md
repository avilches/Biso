# Cómo se pasa un valor

## Tres formas de pasar un valor largo

Todo parámetro de tipo texto largo (`--append-desc`, `--append-plan`, `--append-note`,
`--append-summary`, `--comment` y el texto de un criterio) acepta las tres:

| Forma | Significado |
|---|---|
| `--append-desc "texto"` | el texto literal |
| `--append-desc @ruta/fichero.md` | el contenido del fichero, interpretado como UTF-8 |
| `--append-desc -` | todo lo que llegue por la entrada estándar hasta el fin de fichero |

Reglas:

- **Un texto que empieza de verdad por `@` se escribe `@@`.** El primer `@` se descarta y el resto es
  literal. Es la única secuencia de escape del programa.
- **Los campos de persona nunca interpretan el `@`.** `--assignee`, `--author` y `--comment-author`
  toman su valor tal cual, así que `--comment-author @trello:juan` guarda ese texto y no intenta leer
  ningún fichero.
- **`-` solo puede aparecer una vez por invocación.** Dos parámetros que pidan la entrada estándar son
  un error de uso con código 2, porque el segundo leería un flujo agotado y guardaría el vacío sin
  que se note.
- **Un fichero que no existe es código 4**, con el mensaje `error: --append-desc: file not found: docs/x.md`.
  Un fichero que existe pero no se puede leer es código 8.
- **Un valor vacío, venga de donde venga, no borra nada.** Ver ["El valor vacío"](#el-valor-vacío).

## El valor vacío

Un valor vacío es una cadena sin ningún carácter, o solo con espacios, tanto si llega literalmente
como si llega de un fichero vacío o de una entrada estándar vacía. La regla es única:

| Dónde | Qué pasa |
|---|---|
| En un flag que añade (`--append-note`, `--add-labels`, `--add-ac`, `--append-desc`) | No se añade nada, se emite `warning: --append-note: empty value, nothing was added` y el código sigue siendo 0 |
| En un flag que sustituye (`--replace-labels`) | Deja el campo vacío, igual que `--clear-labels`. Sustituir por nada es vaciar, y eso sí es explícito |
| En un campo escalar (`--type ""`, `--priority ""`) | Error 3. **La cadena vacía nunca es la forma de borrar un escalar**; para eso está `--clear-type` |
| En el título, al crear | Error 2: `error: title cannot be empty` |

**El `code` de un escalar vacío depende de si el campo tiene vocabulario cerrado.** Para `status`,
`type` y `priority`, una cadena vacía es un valor que no coincide con nada configurado, así
que sigue la regla del ["algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia) y el
`code` es el de un valor desconocido (`unknown_status` y análogos, con el mensaje de
["El mismo texto vale lo mismo en los dos sentidos"](vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos)).
Para los demás escalares (`--author ""`, `--ordinal ""`, `--due ""`), que no
tienen vocabulario, el `code` es `empty_scalar_value`.

## Valores que empiezan por guion

Tres mecanismos, en orden de preferencia:

1. **`--flag=valor`** funciona siempre y es la forma recomendada: `--append-desc=-5 grados`.
2. **`--`** termina el análisis de opciones: `biso new -- "-n no es un flag"`.
3. **Un valor que empieza por guion detrás de un flag que exige valor se acepta tal cual**, sin
   heurísticas. `biso set MYP-1 --append-note -x` guarda `-x` como nota.

Como consecuencia de la regla 3, olvidar el valor de un flag se detecta por lo que sobra después,
no por lo que parece: `biso set MYP-1 --append-note --priority high` guarda la nota `--priority` y
luego falla con código 2 y `error: unexpected argument: high`.

## Repetición y listas separadas por comas

Para todo flag marcado como repetible:

- Repetirlo acumula: `--add-labels a --add-labels b` deja dos etiquetas.
- Si además acepta lista, separar por comas acumula igual: `--add-labels a,b` deja las mismas dos.
- Las dos formas se pueden mezclar.
- **Una coma dentro de un valor se escapa con `\,`.** Es la única forma de meter una coma en una
  referencia, en una documentación o en un fichero tocado. Una etiqueta, una persona asignada o una
  clave de `ext` nunca llevan coma, así que en ninguno de esos campos hay nada que escapar
  (["El juego de caracteres de un token"](#el-juego-de-caracteres-de-un-token)).
- Los campos de texto largo y los criterios **nunca** se parten por comas.
- Un valor repetido dentro del mismo flag se guarda una vez y produce
  `warning: --add-labels: "urgent" given twice, kept once`.

Para todo flag **no** repetible, es decir, los campos escalares, pasarlo dos veces con valores
distintos es un error de uso con código 2:

```
error: --status given twice with different values: "In Progress" and "Done"
```

## El juego de caracteres de un token

`labels`, `assignees` y las claves de `ext` (["Campos externos"](modelo-de-datos/campos-externos.md#los-campos-externos)) son los únicos
campos de esta sección cuyo alfabeto está cerrado. Los demás campos de lista de la tabla de
["Campos de lista que admiten coma"](familias-de-flags.md#campos-de-lista-que-admiten-coma), es decir `references`, `documentation`,
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
comparación entera, está en ["El juego de caracteres de un token"](../decisiones/detalles.md#el-juego-de-caracteres-de-un-token).

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

## El salto de línea en un campo `string`

`title`, `author` de tarea/comentario/pregunta, `Criterion.text` de `acceptanceCriteria`,
y los valores (no las claves) de `ext` son del tipo `string` de ["El modelo de datos de una
tarea"](modelo-de-datos/index.md#el-modelo-de-datos-de-una-tarea), es decir, texto de **una línea**.
Ninguno de ellos admite un `\r` o un `\n` literal: si lo llevara, dejaría de ser una línea, y `title`
en particular rompería la promesa de ["`biso ls`"](cmd/ls.md#salida) de que cada tarea ocupa
exactamente una línea de la tabla.

**Un `\r` o un `\n` en cualquiera de estos campos es error 2 (`USAGE`)**, con el mismo tratamiento que
un carácter fuera del alfabeto de la sección anterior:

```
error: malformed title: "first line\nsecond line"
hint: a string field cannot contain a newline or a carriage return
```

El `code` es `malformed_string_value`, con `field` igual a `title`, `author`, `criterion_text` o
`ext`, según cuál sea el campo. Está en la tabla de ["Los identificadores de
error"](contrato-json.md#los-identificadores-de-error).

**Esto rige al escribir**, con la misma excepción que la sección anterior: un valor ya guardado con un
salto de línea, porque se escribió antes de que existiera esta regla o llegó por una vía que no pasa
por esta validación, se trata como un dato que no se puede interpretar (["Qué pasa con un dato que no
se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)).

`Comment.body` y `Question.body` son del tipo `text`, no `string`, y admiten salto de línea sin
ninguna restricción: la pregunta y su respuesta pueden ser tan largas y estructuradas como haga
falta.


## Cómo se lee la línea de comandos

Lo de arriba dice qué significa un valor. Esto dice cómo se separa un flag de su valor y qué es un
argumento suelto, que es lo que hace falta para llegar a aplicarlo.

- **Un flag largo se escribe `--nombre`, `--nombre=valor` o `--nombre valor`**, y uno corto `-x`,
  `-x=valor` o `-x valor`. **Los flags cortos no se agrupan ni se pegan a su valor**: `-qV` y
  `-lparser` no son un par de flags ni un flag con su valor, son un flag desconocido, porque una
  línea que se puede leer de más de una forma deja de ser predecible.
- **Un flag que no lleva valor rechaza el que se le pegue.** `--json=true` es error 2, con el `code`
  `unexpected_argument`.
- **Los flags globales se pueden escribir antes o después del nombre del comando; los propios de un
  comando, solo después**, porque antes del nombre no hay todavía ninguna tabla donde buscarlos:
  `biso --type bug new` es un flag desconocido y `biso new --type bug` no lo es.
- **`--help` y `--version` terminan el programa en cuanto se leen**, así que lo que venga detrás no
  se analiza: `biso ls --help --loquesea` imprime la ayuda y sale con 0, mientras que
  `biso ls --loquesea --help` falla por el flag desconocido, que se lee antes.
- **Los argumentos sueltos de una llamada forman un único bloque contiguo.** Pueden ir todos delante
  de los flags o todos detrás, pero no a los lados de ellos: en cuanto aparece un flag después de
  uno de ellos, el bloque se cierra, y cualquier argumento suelto posterior es error 2. Es lo que
  hace que olvidar un valor se detecte por lo que sobra después
  (["Valores que empiezan por guion"](#valores-que-empiezan-por-guion)): en
  `biso set MYP-1 --append-note --priority high`, el `high` que sobra llega con el bloque ya cerrado.
- **Detrás de `--`, todo es un argumento suelto**, esté donde esté en la línea y sin que le afecte la
  regla anterior. Es la forma que siempre funciona.
- **`-` significa la entrada estándar venga como venga.** `--append-desc -` y `--append-desc=-` son
  lo mismo, porque las formas de ["Tres formas de pasar un valor largo"](#tres-formas-de-pasar-un-valor-largo)
  hablan del valor y no de cómo se pegó al flag.
- **Un mensaje nombra siempre el flag por su forma larga**, aunque se haya escrito la corta, para que
  el texto de un error no dependa de cómo se tecleó la llamada: `-l 'urgent!'` falla nombrando
  `--add-labels`.
- **Un valor vacío en un flag que no añade, no sustituye y no es un escalar de
  ["El valor vacío"](#el-valor-vacío)**, por ejemplo `--rm-labels ""` o `--cwd ""`, es error 3 con el
  `code` `empty_scalar_value`, igual que un escalar: donde la especificación no documenta un valor
  vacío, escribirlo es un error y no algo que se ignore en silencio.

### Los mensajes del análisis

| Caso | Código | `code` | Mensaje |
|---|---:|---|---|
| El comando no existe | 2 | `unknown_command` | `error: unknown command: "sett"` |
| El flag no existe | 2 | `unknown_flag` | `error: unknown flag: --nope` |
| Falta el valor de un flag que lo exige | 2 | `missing_value` | `error: --status requires a value` |
| Un flag sin valor con un valor pegado | 2 | `unexpected_argument` | `error: --json takes no value` |
| Un argumento suelto con el bloque ya cerrado | 2 | `unexpected_argument` | `error: unexpected argument: high` |
| Un escalar repetido con valores distintos | 2 | `duplicate_scalar_flag` | `error: --status given twice with different values: "In Progress" and "Done"` |
| Un valor de pareja al que le falta su `=` | 2 | `unexpected_argument` | `error: --ext: expected <key>=<value>, got "trello"` |
| Más de un flag pidiendo la entrada estándar | 2 | `two_stdin` | `error: - can be given only once per invocation; --append-desc and --append-plan both read stdin` |
| Un valor fuera del dominio de `--color` | 2 | `invalid_color_mode` | `error: --color: unknown value: "sometimes"` |
| Una pareja de `--json`, `--quiet` y `--print` | 2 | `incompatible_flags` | `error: --json and --quiet cannot be used together` |
| Un flag que exige otro, sin ese otro | 2 | `incompatible_flags` | `error: --comment-author requires --comment` |
| `--dry-run` en un comando de lectura | 2 | `read_only_flag` | `error: --dry-run does not apply to a read-only command` |
| `--print` donde no afecta a ninguna tarea | 2 | `read_only_flag` | `error: --print does not apply to a command that affects no task` |
| Un escalar vacío sin vocabulario cerrado | 3 | `empty_scalar_value` | `error: --author cannot be empty`, con `hint: to clear it, use --clear-author` cuando el campo tiene un flag que lo vacía |
| Una secuencia de bytes que no es UTF-8 | 3 | `invalid_encoding` | `error: --append-desc: invalid UTF-8 at byte 12` |
| El fichero de un `@` que no existe | 4 | `file_not_found` | `error: --append-desc: file not found: docs/x.md` |
| El fichero de un `@` que no se puede leer | 8 | `file_unreadable` | `error: --append-desc: file cannot be read: docs/x.md` |
| La entrada estándar que falla al leerse | 8 | `io_error` | `error: --append-desc: stdin cannot be read: <lo que dijo el sistema>` |

**`incompatible_flags` no lleva `field` ni `given`**, a diferencia de los demás errores de esta tabla
que nombran un flag. Nombra un par de ellos, y ninguno es más culpable que el otro, así que elegir
uno para `field` sería inventarse una atribución que la llamada no tiene.
