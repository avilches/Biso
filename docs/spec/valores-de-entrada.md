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
  que se note. **La cuenta se hace sobre la línea de comandos entera y antes de leer ningún valor**,
  igual que la comprobación de codificación de ["Codificación y texto"](salida-y-terminal.md#codificación-y-texto):
  una llamada que va a ser rechazada por esto no llega a tocar el flujo, así que no lo vacía para
  quien la hizo y no emite ningún aviso sobre un valor que nunca se iba a guardar. Cuentan los dos
  sitios donde un `-` significa la entrada estándar, el valor de un flag de texto largo y el
  posicional que un verbo lee como texto, y da igual cuál de los dos vaya primero.
- **Un guion suelto significa siempre la entrada estándar, y no hay ninguna forma de escribirlo como
  texto literal.** El escape `@@` existe solo para el `@`, y no hay ningún `--` ni ninguna otra
  sintaxis que convierta ese guion en un valor. Para guardar un texto que sea exactamente un guion,
  la forma es la del fichero: escribirlo en uno y pasarlo con `@ruta`. No se inventa una sintaxis
  nueva para un caso que ya tiene solución, porque cada escape que se añade hay que recordarlo en
  todos los flags de texto largo, y este es el único carácter del que no se puede hablar sin él.
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
  referencia o en un fichero tocado. Una etiqueta, una persona asignada o una
  clave de `ext` nunca llevan coma, así que en ninguno de esos campos hay nada que escapar
  (["El juego de caracteres de un token"](#el-juego-de-caracteres-de-un-token)).
- **Una barra invertida doble, `\\`, es una barra invertida literal**, y es la única forma de que un
  valor termine en barra invertida justo antes de una coma que separa:
  `--add-refs 'C:\\dir\\,notes/b.md'` añade dos referencias, `C:\dir\` y `notes/b.md`. Una barra
  invertida delante de cualquier otro carácter no escapa nada y se guarda tal cual, así que una ruta
  de Windows o una expresión regular conservan lo que traen.
- Los campos de texto largo y los criterios **nunca** se parten por comas.
- Un valor repetido dentro del mismo flag se guarda una vez y produce
  `warning: --add-labels: "urgent" given twice, kept once`. Con tres apariciones o más el aviso sigue
  siendo **uno solo** por valor, y las cuenta: `warning: --add-labels: "urgent" given 3 times, kept once`.

Para todo flag **no** repetible, es decir, los campos escalares, pasarlo dos veces con valores
distintos es un error de uso con código 2:

```
error: --status given twice with different values: "In Progress" and "Done"
```

## El juego de caracteres de un token

`labels`, `assignees` y las claves de `ext` (["Campos externos"](modelo-de-datos/campos-externos.md#los-campos-externos)) son los únicos
campos de esta sección cuyo alfabeto está cerrado. Los demás campos de lista de la tabla de
["Campos de lista que admiten coma"](familias-de-flags.md#campos-de-lista-que-admiten-coma), es decir `references`,
`dependencies` y `modifiedFiles`, son texto libre y no tienen ninguna restricción de caracteres: una
referencia puede ser una URL, y un fichero tocado es una ruta, y ninguna de las
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
  `-Cdir` no son un par de flags ni un flag con su valor, son un flag desconocido, porque una
  línea que se puede leer de más de una forma deja de ser predecible.
- **Un flag que no lleva valor rechaza el que se le pegue.** `--json=true` es error 2, con el `code`
  `unexpected_argument`.
- **Los flags globales se pueden escribir antes o después del nombre del comando; los propios de un
  comando, solo después**, porque antes del nombre no hay todavía ninguna tabla donde buscarlos:
  `biso --type bug new` es un flag desconocido y `biso new --type bug` no lo es.
- **`--help` y `--version` terminan el programa en cuanto se leen**, así que lo que venga detrás no
  se analiza: `biso ls --help --loquesea` imprime la ayuda y sale con 0, mientras que
  `biso ls --loquesea --help` falla por el flag desconocido, que se lee antes.
- **Los argumentos sueltos de una llamada van todos delante de los flags de su comando.** El bloque
  empieza justo detrás del nombre del comando y lo cierra el primer flag que venga después; a partir
  de ahí, cualquier argumento suelto es error 2. Es la forma en que están escritas todas las firmas y
  todos los ejemplos de ["Los comandos"](cmd/index.md), donde `<ref>`, `<title>`, `<key>` y `<value>`
  preceden siempre a los flags, y es lo que hace que olvidar un valor se detecte por lo que sobra
  después (["Valores que empiezan por guion"](#valores-que-empiezan-por-guion)): en
  `biso set MYP-1 --append-note --priority high`, el `high` que sobra llega con el bloque ya cerrado.
  Un flag global escrito **antes** del nombre del comando no cierra nada, porque ahí no puede haber
  ningún argumento suelto: el primero que no es un flag es el nombre del comando, así que
  `biso -C ~/work/my-project set MYP-1` es correcto.
- **Detrás de `--`, todo es un argumento suelto**, esté donde esté en la línea y sin que le afecte la
  regla anterior. Es la forma que siempre funciona.
- **`-` significa la entrada estándar venga como venga.** `--append-desc -` y `--append-desc=-` son
  lo mismo, porque las formas de ["Tres formas de pasar un valor largo"](#tres-formas-de-pasar-un-valor-largo)
  hablan del valor y no de cómo se pegó al flag.
- **Un mensaje nombra siempre el flag por su forma larga**, aunque se haya escrito la corta, para que
  el texto de un error no dependa de cómo se tecleó la llamada: `-C ""` falla nombrando
  `--cwd`, con `error: --cwd cannot be empty`.
- **Un valor vacío en un flag que no añade, no sustituye y no es un escalar de tarea de
  ["El valor vacío"](#el-valor-vacío)**, por ejemplo `--rm-labels ""`, `--cwd ""` o `--ext k=`, es
  **error 2** con el `code` `unexpected_argument`: donde la especificación no documenta un valor
  vacío, escribirlo es un error y no algo que se ignore en silencio. Es el 2 y no el 3 porque el 3 es
  "el valor llega bien formado pero el tablero no lo reconoce"
  (["Códigos de salida"](codigos-de-salida.md)), y el vocabulario del tablero no tiene nada que decir
  sobre el directorio de trabajo o sobre una etiqueta que se quería quitar. El escalar de una tarea
  sigue siendo la excepción, con su `empty_scalar_value` de código 3, porque ahí sí es el vocabulario
  quien juzga.
- **La codificación se comprueba sobre la línea de comandos entera antes de leer nada más.** Cada
  argumento, incluidos el nombre del comando y la grafía de un flag, tiene que ser UTF-8 válido
  (["Codificación y texto"](salida-y-terminal.md#codificación-y-texto)), y el argumento que no lo sea
  viaja al mensaje y a `given` con sus bytes escapados, nunca tal cual, para que un byte inválido no
  termine convertido en el carácter de reemplazo dentro de la salida.

### Los mensajes del análisis

| Caso | Código | `code` | Mensaje |
|---|---:|---|---|
| El comando no existe | 2 | `unknown_command` | `error: unknown command: "sett"` |
| El flag no existe | 2 | `unknown_flag` | `error: unknown flag: --nope` |
| Falta el valor de un flag que lo exige | 2 | `missing_value` | `error: --status requires a value` |
| Un flag sin valor con un valor pegado | 2 | `unexpected_argument` | `error: --json takes no value` |
| Un argumento suelto con el bloque ya cerrado | 2 | `unexpected_argument` | `error: unexpected argument: high` |
| Un escalar repetido con valores distintos | 2 | `duplicate_scalar_flag` | `error: --status given twice with different values: "In Progress" and "Done"` |
| La misma clave de `--set-comment-date` con dos instantes (["Comentarios"](familias-de-flags.md#comentarios)) | 2 | `duplicate_scalar_flag` | `error: --set-comment-date: key "3" given twice with different values: "2026-08-14T10:22:00Z" and "2026-08-15T10:22:00Z"` |
| Un valor de pareja al que le falta su `=` | 2 | `unexpected_argument` | `error: --ext: expected <key>=<value>, got "trello"` |
| Más de un flag pidiendo la entrada estándar | 2 | `two_stdin` | `error: - can be given only once per invocation; --append-desc and --append-plan both read stdin` |
| El mismo flag repetible pidiéndola dos veces | 2 | `two_stdin` | `error: - can be given only once per invocation; --append-desc reads stdin twice` |
| Un valor fuera del dominio de `--color` | 2 | `invalid_color_mode` | `error: --color: unknown value: "sometimes"` |
| `BISO_LIMIT` con algo que no es un número de filas (["Variables de entorno"](invocacion.md#variables-de-entorno)) | 2 | `invalid_number` | `error: BISO_LIMIT: not a whole number of rows: "lots"` |
| Una pareja de `--json`, `--quiet` y `--print` | 2 | `incompatible_flags` | `error: --json and --quiet cannot be used together` |
| Un flag que exige otro, sin ese otro | 2 | `incompatible_flags` | `error: --comment-author requires --comment` |
| El título de `biso new` dado como argumento y con `--title` | 2 | `incompatible_flags` | `error: --title and the title argument cannot be used together` |
| `--dry-run` en un comando de lectura | 2 | `read_only_flag` | `error: --dry-run does not apply to a read-only command` |
| `--print` donde no afecta a ninguna tarea | 2 | `read_only_flag` | `error: --print does not apply to a command that affects no task` |
| Un valor vacío donde no se documenta ninguno | 2 | `unexpected_argument` | `error: --rm-labels cannot be empty` |
| El valor vacío de una pareja | 2 | `unexpected_argument` | `error: --ext: the value of key "k" cannot be empty` |
| Un salto de línea en un campo de una línea, la mitad derecha de una pareja incluida (["El salto de línea en un campo `string`"](#el-salto-de-línea-en-un-campo-string)) | 2 | `malformed_string_value` | `error: malformed ext: "first line\nsecond line"` |
| Un escalar de tarea vacío sin vocabulario cerrado | 3 | `empty_scalar_value` | `error: --author cannot be empty`, con `hint: to clear it, use --clear-author` cuando el campo tiene un flag que lo vacía |
| Un argumento que no es UTF-8 | 3 | `invalid_encoding` | `error: invalid UTF-8 in argument 4 at byte 2: "ok\xffbad"`, con `field` igual a `argument` |
| Un fichero o una entrada estándar que no es UTF-8 | 3 | `invalid_encoding` | `error: --append-desc: invalid UTF-8 at byte 12`, con `given` igual a lo que se escribió detrás del flag |
| El fichero de un `@` que no existe | 4 | `file_not_found` | `error: --append-desc: file not found: docs/x.md` |
| El fichero de un `@` que no se puede leer | 8 | `file_unreadable` | `error: --append-desc: file cannot be read: docs/x.md` |
| La entrada estándar que falla al leerse | 8 | `io_error` | `error: --append-desc: stdin cannot be read: <lo que dijo el sistema>` |

**Un mensaje que nombra un par de flags los nombra en el orden de la tabla de especificación**, primero los
globales de ["Los flags globales"](cmd/flags-globales.md) y después los propios del comando, cada
grupo en el orden de su propia tabla de parámetros. `--json --quiet` y `--quiet --json` fallan con el
mismo texto, que es lo que exige la regla de que el texto de un error no dependa de cómo se tecleó la
llamada. La excepción de `incompatible_flags`, que no lleva `field` ni `given`, vive junto a la tabla
que enuncia esa obligación, en ["Los errores en JSON"](contrato-json.md#los-errores-en-json).
