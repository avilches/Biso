# Flags globales

## Banderas globales

Valen para todos los comandos, se pueden escribir antes o después del nombre del comando, y ningún
comando puede redefinir ninguna de ellas ni cambiar su significado.

| Bandera | Corta | Tipo | Por defecto | Qué hace |
|---|---|---|---|---|
| `--cwd <path>` | `-C` | ruta | el directorio actual | Resuelve el tablero desde ahí, sin cambiar el directorio del proceso |
| `--json` | | booleano | falso | Toda la salida de datos es JSON, en el sobre de la sección ["El sobre"](../contrato-json.md#el-sobre) |
| `--quiet` | `-q` | booleano | falso | Reduce la salida a lo mínimo. Ver más abajo |
| `--print` | | booleano | falso | Después de escribir, imprime la ficha completa de cada tarea afectada |
| `--color <when>` | | `auto`, `always`, `never` | `auto` | Control de los códigos de color |
| `--dry-run` | | booleano | falso | Valida todo, no escribe nada. Sale 0 si habría funcionado y 7 si no |
| `--version` | `-V` | booleano | | Imprime `biso 1.0.0` y sale con 0 |
| `--help` | `-h` | booleano | | Imprime la ayuda del comando y sale con 0 |

Reglas de aplicación, que hay que implementar tal cual:

- **Ninguna de las dos se ignora nunca en silencio**, y las dos son error de uso con código 2 allí
  donde no tienen nada que hacer. Lo que cambia es dónde es eso, porque cada una está definida sobre
  una cosa distinta: `--print` sobre las tareas que una escritura afecta, y `--dry-run` sobre la
  validación que precede a una escritura.
- **`--dry-run` es error 2 en los comandos de lectura.** En `prime`, `where`, `ls`, `get`, `export`,
  `snapshot`, `config get`, `config list`, `board`, `help` y `biso doctor` sin `--fix`, con el mensaje
  `error: --dry-run does not apply to a read-only command`. `biso doctor --fix` es la excepción: con
  `--fix` es un comando de escritura y la bandera se comporta como en cualquier otro (["`biso doctor`"](doctor.md)).
  `snapshot` (["`biso snapshot`"](snapshot.md)) entra en la lista por el mismo motivo que `export`, que escriben ficheros y no
  tocan ninguna tarea.
- **`--dry-run` sí vale en `biso init` y en `biso config set`**, que escriben sin tocar ninguna tarea
  existente y tienen los dos algo que validar antes: `init --from` valida la instantánea entera contra
  el vocabulario que ella misma trae, y sale 0 si habría funcionado y 7 si no, que es exactamente lo
  que la definición de la bandera promete; `config set` valida el valor contra el tablero. Validar en
  seco la restauración de un tablero de doscientas tareas sin crear nada es el caso donde más vale, así
  que dejarla fuera de `init` sería perder lo mejor que tiene.
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
