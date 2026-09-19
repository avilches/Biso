# Flags globales

Valen para todos los comandos, se pueden escribir antes o después del nombre del comando, y ningún
comando puede redefinir ninguno de ellos ni cambiar su significado.

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
| `--dry-run` | | Valida todo, no escribe nada. Sale 0 si habría funcionado y 7 si no |
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
- **`--dry-run` sí vale en `biso init` y en `biso config set`**, que escriben sin tocar ninguna tarea
  existente y tienen los dos algo que validar antes: `init --from` valida la instantánea entera contra
  el vocabulario que ella misma trae, y sale 0 si habría funcionado y 7 si no, que es exactamente lo
  que la definición del flag promete; `config set` valida el valor contra el tablero. Validar en
  seco la restauración de un tablero de doscientas tareas sin crear nada es el caso donde más vale, así
  que dejarla fuera de `init` sería perder lo mejor que tiene.
- **`--dry-run` emite los mismos avisos que emitiría la escritura real**, por stderr y con el mismo
  `code` en `data.warnings` (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)), porque un
  aviso dice algo que quien llama necesita saber **antes** de escribir y callarlo en seco dejaría a la
  vista previa diciendo menos que la llamada que simula. Los avisos que dependen de datos que solo
  existen tras escribir, como el identificador que recibe una tarea nueva, se emiten igualmente
  nombrando lo que sí existe en ese momento, que es la línea del fichero de entrada.
- **`biso doctor --fix --dry-run` es la única excepción al código 7 de `--dry-run` en un comando de
  escritura.** `doctor` nunca tuvo código 7 en su propia tabla de códigos, así que su vista previa
  devuelve el código que daría la llamada real que le sigue (0 o 6), nunca un 7 que esa tabla no tiene;
  el razonamiento completo está en ["El informe en seco de `--fix --dry-run`"](doctor.md#el-informe-en-seco-de---fix---dry-run).
- **`--print` es error 2 en todos esos comandos y también en `biso init` y `biso config set`**, con el
  mensaje `error: --print does not apply to a command that affects no task`. Ninguno de los dos afecta
  a ninguna tarea que existiera antes: `config set` no toca ninguna nunca, e `init --from` las crea de
  cero en un tablero que acaba de nacer, así que imprimir doscientas fichas recién importadas no
  informa de nada que no diga ya `biso ls`.
- **`--json` es incompatible con `--quiet`** y con `--print`, porque los tres piden formas distintas
  de la misma salida. Cualquier pareja de las tres da código 2.
- **`--quiet` reduce stdout a los identificadores afectados**, uno por línea, y además suprime las
  líneas informativas de stderr que empiezan por `note:`. **Nunca suprime un `warning:`, un `error:` ni
  la salida reenviada de un programa ajeno** (["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)). Silenciar un aviso es cosa de quien llama, con
  `2>/dev/null`. **En un comando de
  lectura no hay identificadores afectados que imprimir**, así que ahí `--quiet` no cambia stdout: solo
  suprime las líneas `note:` de stderr, igual que en un comando de escritura.
