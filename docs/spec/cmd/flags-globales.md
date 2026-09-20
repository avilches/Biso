# Flags globales

Valen para todos los comandos, se pueden escribir antes o después del nombre del comando, y ningún
comando puede redefinir ninguno de ellos ni cambiar su significado.

**Estos son los únicos flags de `biso` con forma corta**: `-C`, `-q`, `-V` y `-h`. Ningún flag propio de
un comando la tiene, porque una forma corta solo existe si ningún otro flag del mismo comando empieza por
su letra (["Una forma corta solo existe si nadie más reclama su inicial"](../../decisiones/comandos-y-flags.md#una-forma-corta-solo-existe-si-nadie-más-reclama-su-inicial)).
Escribir la inicial de cualquier otro flag es un error 2 con el `code` `unknown_flag`, como con cualquier flag que
el comando no tiene: `error: unknown flag: -x`.

Flags con valor:

| Flag | Corta | Valores | Por defecto | Qué hace |
|---|---|---|---|---|
| `--cwd <path>` | `-C` | una ruta | el directorio actual | Resuelve el tablero desde ahí, sin cambiar el directorio del proceso |
| `--color <when>` | | `auto`, `always`, `never` | `auto` | Control de los códigos de color |

Los flags que siguen no llevan valor: ponerlos lo activa, y no ponerlos lo deja como está.
`--json`, `--quiet`, `--print` y `--dry-run` activan un modo mientras están puestos, y ese modo está
desactivado si no se ponen. `--version` y `--help` no son un modo: son una acción que imprime algo y
termina el programa con código 0 en cuanto se lee, sin más.

| Flag | Corta | Qué hace |
|---|---|---|
| `--json` | | Toda la salida de datos es JSON, en el sobre de la sección ["El sobre"](../contrato-json.md#el-sobre) |
| `--quiet` | `-q` | Reduce la salida a lo mínimo. Ver más abajo |
| `--print` | | Después de escribir, imprime la ficha completa de cada tarea afectada |
| `--dry-run` | | Valida todo, no escribe nada. Sale 0 si habría funcionado, y si no, el código del fallo |
| `--version` | `-V` | Imprime `biso 1.0.0` y sale con 0 |
| `--help` | `-h` | Imprime la ayuda del comando y sale con 0 |

Reglas de aplicación, que hay que implementar tal cual:

- **`--print` y `--dry-run` no se ignoran nunca en silencio: allí donde no tienen nada que hacer, los
  dos son error de uso con código 2.** Lo que cambia entre ellos es dónde es eso, porque cada uno está
  definido sobre una cosa distinta: `--print` sobre las tareas que una escritura afecta, y `--dry-run`
  sobre la validación que precede a una escritura.
- **`--dry-run` es error 2 en los comandos de lectura.** "De lectura" quiere decir que ninguna tarea
  del tablero cambia, no que el comando no toque el sistema de ficheros: `export`, `snapshot` y el
  sondeo de `biso doctor` sin `--fix` (["El sondeo del sistema de ficheros"](doctor.md#el-sondeo-del-sistema-de-ficheros)) escriben
  ficheros propios sin que eso los saque de esta lista. En `prime`, `where`, `ls`, `get`, `export`,
  `snapshot`, `config get`, `config list`, `board`, `help` y `biso doctor` sin `--fix`, con el mensaje
  `error: --dry-run does not apply to a read-only command`. `biso doctor --fix` es la excepción: con
  `--fix` es un comando de escritura y el flag se comporta como en cualquier otro (["`biso doctor`"](doctor.md)).
  `snapshot` (["`biso snapshot`"](snapshot.md)) entra en la lista por el mismo motivo que `export`, que escriben ficheros y no
  tocan ninguna tarea.
- **El código de un `--dry-run` que no pasa es el del fallo, y nunca uno propio de la vista previa.**
  Una vista previa contesta lo mismo que contestaría la llamada que simula: 2 para un título vacío,
  3 para un valor fuera de un vocabulario, 4 para una referencia que no existe, y 7 solo donde el 7
  ya vivía sin `--dry-run`, que es la validación de conjunto de un lote
  (["El código 7 garantiza que no se ha escrito nada, y el código específico siempre gana sobre
  él"](../codigos-de-salida.md#el-código-7-garantiza-que-no-se-ha-escrito-nada-y-el-código-específico-siempre-gana-sobre-él)).
  Que la escritura no se haya llegado a intentar no cambia a quién es atribuible el fallo, y un
  código propio de "la vista previa no pasó" obligaría a leer la prosa del mensaje para saber qué
  falló, que es justo lo que un código específico existe para evitar.
- **`--dry-run` sí vale en `biso init` y en `biso config set`**, que escriben sin tocar ninguna tarea
  existente y tienen los dos algo que validar antes: `init --from` valida la instantánea entera contra
  el vocabulario que ella misma trae, y sale 0 si habría funcionado y 7 si el lote no es válido, que
  es el mismo 7 del lote de `biso new --from`; `config set` valida el valor contra el tablero y
  contesta el código de ese valor. Validar en
  seco la restauración de un tablero de doscientas tareas sin crear nada es el caso donde más vale, así
  que dejarla fuera de `init` sería perder lo mejor que tiene.
- **`--dry-run` emite los mismos avisos que emitiría la escritura real**, por stderr y con el mismo
  `code` en `data.warnings` (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)), porque un
  aviso dice algo que quien llama necesita saber **antes** de escribir y callarlo en seco dejaría a la
  vista previa diciendo menos que la llamada que simula. Los avisos que dependen de datos que solo
  existen tras escribir, como el identificador que recibe una tarea nueva, se emiten igualmente
  nombrando lo que sí existe en ese momento, que es la línea del fichero de entrada.
- **`biso doctor --fix --dry-run` no contesta el código de un hallazgo suelto, sino el de su informe
  entero.** Su vista previa devuelve el código que daría la llamada real que le sigue, 0 o 6, que es
  lo que dice su propia tabla de códigos; el razonamiento completo está en
  ["El informe en seco de `--fix --dry-run`"](doctor.md#el-informe-en-seco-de---fix---dry-run).
- **La ficha que imprime `--print` es la de ["`biso get`"](get.md), entera y sin `--section`, y sustituye la
  salida por defecto del comando en vez de añadirse debajo.** Los tres datos de la línea de estado de
  ["`biso set`"](set.md) (el estado, el avance de los criterios y la urgencia) están ya dentro de esa ficha, así que
  imprimir las dos cosas sería repetirlos. Con varias tareas afectadas va una ficha por tarea,
  separadas por una línea en blanco. Con `--dry-run` se imprimen igual, bajo la misma cabecera de
  vista previa, porque la ficha describe cómo quedaría la tarea. La excepción es `biso new --dry-run`,
  que no imprime ninguna: una tarea que no se ha creado no tiene identificador, y la ficha empieza
  por él (["`biso new`"](new.md#--dry-run-sobre-una-sola-tarea)).
- **`--print` es error 2 en todos esos comandos y también en `biso init`, `biso config set` y
  `biso doctor --fix`**, con el mensaje `error: --print does not apply to a command that affects no
  task`. Ninguno de los tres afecta a ninguna tarea que existiera antes: `config set` no toca ninguna
  nunca, `init --from` las crea de cero en un tablero que acaba de nacer, así que imprimir doscientas
  fichas recién importadas no informa de nada que no diga ya `biso ls`, y lo que repara
  `doctor --fix` no es ninguna tarea (["`biso doctor`"](doctor.md)). `doctor --fix` es el único de
  los tres que sí es un comando de escritura, y aun así `--print` no tiene ahí nada que imprimir: el
  flag está definido sobre las tareas que una escritura afecta, no sobre el hecho de escribir.
- **`--json` es incompatible con `--quiet`** y con `--print`, porque los tres piden formas distintas
  de la misma salida. Cualquier pareja de las tres da código 2.
- **`--quiet` reduce stdout a los identificadores afectados**, uno por línea, y además suprime las
  líneas informativas de stderr que empiezan por `note:`. **Nunca suprime un `warning:`, un `error:` ni
  la salida reenviada de un programa ajeno** (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)). Silenciar un aviso es cosa de quien llama, con
  `2>/dev/null`. **En un comando de
  lectura no hay identificadores afectados que imprimir**, así que ahí `--quiet` no cambia stdout: solo
  suprime las líneas `note:` de stderr, igual que en un comando de escritura.
