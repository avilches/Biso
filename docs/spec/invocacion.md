# Entorno y configuración de máquina

## Variables de entorno

| Variable | Qué hace |
|---|---|
| `BISO_CWD` | Fija el directorio desde el que se resuelve el tablero, igual que el flag global `--cwd` |
| `BISO_ME` | Declara la identidad de quien llama: la usan `biso ls --mine`, la autoasignación de `biso start` y el autor por defecto de los comentarios, de las preguntas y de las respuestas |
| `BISO_LIMIT` | Fija cuántas filas imprime `biso ls` cuando no se le pasa `--limit` |
| `NO_COLOR` | Desactiva los colores, igual que `--color never`. Basta con que esté definida, con cualquier valor |

Cuando la variable y el flag dicen cosas distintas, **el flag de la línea de comandos gana siempre**,
porque es lo más específico de la llamada:

- `--cwd` prevalece sobre `BISO_CWD`.
- `--limit` prevalece sobre `BISO_LIMIT`. Si no hay ni `--limit` ni `BISO_LIMIT`, se usa la clave
  `default_limit` de la ["Configuración de máquina"](#configuración-de-máquina), y si tampoco está, 30.
- `--color` prevalece sobre `NO_COLOR`. `--color` no se escribe solo, siempre lleva uno de sus tres
  valores (`auto`, `always` o `never`, ["Flags globales"](cmd/flags-globales.md)), así que
  `NO_COLOR=1 biso ls --color always` imprime con color.
- `BISO_ME` no tiene flag. Si no está definida, se usa la clave `me` de la
  ["Configuración de máquina"](#configuración-de-máquina), en `~/.biso/config.json`.

**Qué pasa si no hay identidad**, es decir, si no está definida la variable de entorno `BISO_ME` ni la
clave `me` del fichero de configuración de máquina `~/.biso/config.json`:

| Comandos que necesitan identidad | Qué pasa sin ella |
|---|---|
| `biso ls --mine` | Error 6:<br>`error: --mine needs an identity; set BISO_ME, or add "me" to ~/.biso/config.json` |
| `biso start`, que se asigna la tarea a quien llama | No falla, pero no asigna la tarea a nadie. En vez del `note:` de siempre sale:<br>`note: no identity configured, task left unassigned` |
| `biso comment`, o el flag `--comment` de cualquier comando, cuando no se pasa `--comment-author` | Error 2:<br>`error: --comment-author is required, no identity is configured` |
| `biso ask` (autor de la pregunta) | Error 2 (["`biso ask`"](cmd/verbos-del-ciclo.md#biso-ask)):<br>`error: biso ask needs an identity; set BISO_ME, or add "me" to ~/.biso/config.json` |
| `biso answer` (autor de la respuesta) | Error 2 (["`biso answer`"](cmd/verbos-del-ciclo.md#biso-answer)):<br>`error: biso answer needs an identity; set BISO_ME, or add "me" to ~/.biso/config.json` |
| `biso prime` (la línea `you are`) | No falla. La línea sale por stdout, dentro del bloque `BOARD`, sin ninguna `note:` por stderr (["`biso prime`"](cmd/prime.md#la-salida-literal)):<br>`you are     (not set: run biso as BISO_ME=@you biso ...)` |
| `biso where` (la fila `me`) | No falla. La fila sale por stdout, sin ninguna `note:` por stderr, y `data.me` vale `null` en el JSON (["`biso where`"](cmd/where.md)):<br>`me       (not set: run biso as BISO_ME=@you biso ...)` |

## Configuración de máquina

`biso` necesita, antes de que exista el primer tablero, saber dónde crearlo, y eso no puede depender
de resolver un tablero con la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md). Por eso vive en un fichero propio de la máquina,
`~/.biso/config.json`, que `biso config` (la sección ["`biso config`"](cmd/config.md)) no gestiona: aquella es la configuración de
un tablero concreto, y esta es de la máquina entera, independiente de cuántos tableros tenga.

| Clave | Tipo | Por defecto |
|---|---|---|
| `boards_root` | ruta | `~/.biso/boards` |
| `boards_extra_roots` | lista de rutas | vacía |
| `vcs` | `git`, `none` o `custom` | `git` |
| `vcs_custom` | objeto, y solo con `vcs` igual a `custom` | ausente |
| `me` | texto de persona | ausente |
| `default_limit` | entero >= 0 | 30 |

`boards_root` es la raíz por defecto de la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md): el directorio donde `biso init` sin `--at` crea
los tableros nuevos. `boards_extra_roots` son raíces adicionales, para cuando algún tablero vive fuera
de `boards_root`.

**`me` y `default_limit` son preferencias de quien usa `biso` en esta máquina, no propiedades de un
tablero**, y por eso viven aquí y no en la configuración de un tablero (sección ["`biso config`"](cmd/config.md#las-claves)):
la primera es la identidad de la sección ["Variables de entorno"](#variables-de-entorno), y la segunda el límite por defecto de
["`biso ls`"](cmd/ls.md). Las dos se leen igual que `boards_root` o `vcs`, directamente de este fichero y sin pasar por
la resolución de tablero, así que valen para todos los tableros de esta máquina y no viajan en la
instantánea de ["`biso snapshot`"](cmd/snapshot.md). El razonamiento completo está en ["La identidad de quien llama"](../decisiones/detalles.md#la-identidad-de-quien-llama).

**Las dos se recorren igual, y en un orden declarado**: `boards_root` primero y después
`boards_extra_roots` en el orden en que estén escritas. Eso vale en los dos sitios que recorren raíces,
que son la búsqueda de un tablero por su `id` (la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md)) y la comprobación de `biso init` de que un
`id` nuevo no exista ya. El orden solo importa para decir qué se lee antes, nunca para elegir entre dos
candidatos: **el mismo `id` en dos raíces es un error** (la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md)), no una preferencia. Una raíz de la
lista que no exista o no se pueda leer no es un error, porque una máquina puede tener configurado un
disco que hoy no está montado: se salta y `biso doctor` la reporta como aviso (["`biso doctor`"](cmd/doctor.md#qué-comprueba)).

Se leen directamente de ese fichero, sin pasar por la resolución de tablero de la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md), porque
hace falta conocerlas antes de que exista el primer tablero de la máquina. Y, como en cualquier otro
sitio de `biso`, **una clave desconocida en este fichero es un error**, nunca algo que se ignore en
silencio.

**`vcs` es el sistema de control de versiones con el que `biso snapshot` guarda el historial de la
instantánea** (la sección ["`biso snapshot`"](cmd/snapshot.md)), y lo único de esta configuración que llega a hacer que `biso` ejecute un
programa ajeno. Vive aquí y no en la configuración de un tablero porque dice qué herramienta hay
instalada en esta máquina, no cómo es un tablero: así no viaja en la instantánea, y restaurar la de otra persona no le impone a nadie el sistema
del ordenador de origen. Vale para todos los tableros de la máquina.

- **`git`** es el valor por defecto y el único sistema que `biso` trae aprendido en la versión 1.0. Su
  receta completa, con las cinco preguntas que le hace al repositorio y las cuatro órdenes que le da, está
  en la sección ["`biso snapshot`"](cmd/snapshot.md), que es la única que lo ejecuta.
- **`none`** deja a `biso snapshot` escribiendo sus dos ficheros y nada más: no busca repositorio, no
  crea ninguno y no ejecuta ningún programa. Es lo que hay que poner en una máquina sin git, o cuando
  el historial de la instantánea no interesa.
- **`custom`** es el escape para un sistema que `biso` no conoce, y entonces `vcs_custom` declara las
  órdenes. Con él, **`biso` no averigua nada del repositorio**: ejecuta la orden en el directorio del
  tablero y mira su código de salida, y de ahí salen las dos únicas cosas que puede decir después, si
  la orden fue bien o si falló. Por eso una instantánea con `custom` nunca lleva identificador de
  revisión en su salida, ni distingue el caso de no haber nada que guardar.

`vcs_custom` tiene estas claves, y una desconocida es un error como en cualquier otro sitio:

| Clave | Tipo | Obligatoria | Qué es |
|---|---|---|---|
| `commit` | lista de argumentos | sí | la orden que guarda una revisión, y sin ella toda instantánea que pida guardar es error 2, `code` `vcs_commit_unavailable` (["`biso snapshot`"](cmd/snapshot.md)) |
| `publish` | lista de argumentos | no | la orden que la publica, y sin ella `--vcs push` es error 2 |
| `ignore_file` | nombre de fichero | no | el fichero de exclusión que `biso init` escribe dentro del tablero (["`biso init`"](cmd/init.md)) |

En `commit` y en `publish`, `{message}` se sustituye por el mensaje que `biso` compone (`biso snapshot:
248 tasks`), y `{files}` por los tres ficheros que entran en la revisión, cada uno como un argumento
propio: `snapshot.ndjson`, `board.json` y el marcador `<id>.id`, siempre en ese orden y siempre como ruta
relativa al directorio del tablero, sin `./` delante.

**Los dos se sustituyen de forma distinta, y es por lo que cada uno produce.** `{message}` es un texto y
se sustituye allí donde aparezca dentro de un argumento, así que valen igual `["-m", "{message}"]` y
`["--message={message}"]`. `{files}` son tres argumentos, así que solo se sustituye cuando es el argumento
entero; dentro de uno más largo se queda tal cual, porque no hay ninguna forma de meter tres argumentos
dentro de uno y cualquier elección de separador sería inventada.

**Son los tres que nombra la sección ["`biso snapshot`"](cmd/snapshot.md), y no los dos de la clave `files` del JSON de `biso
snapshot`**, que son solo los que ese comando escribe. Los dos nombres se parecen y los dos conjuntos son
distintos, así que conviene leerlos juntos antes de escribir una orden.

Un código de salida distinto de cero es error 8 con la clave `code` `vcs_commit_failed`. Cómo se ejecutan
las dos órdenes, qué se hace con lo que escriban y por qué no hay tiempo máximo de espera lo dice la
sección ["`biso snapshot`"](cmd/snapshot.md), en un apartado que vale igual para `git` y para `custom`.

---

