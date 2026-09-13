# Entorno y configuración de máquina

## Variables de entorno

| Variable | Equivale a | Precedencia |
|---|---|---|
| `BISO_CWD` | `--cwd` | la bandera gana |
| `BISO_ME` | la identidad de quien llama, para `--mine` y para el autor por defecto de los comentarios | la clave `me` de la configuración gana; si no está, esta variable |
| `BISO_LIMIT` | el límite por defecto de `biso ls` | `--limit` gana, luego esta variable, luego la clave `default_limit`, luego 30 |
| `NO_COLOR` | `--color never`, si está definida con cualquier valor | `--color` gana |

**Qué pasa si no hay identidad**, es decir, ni la clave `me` ni `BISO_ME` están definidas:

| Dónde se usaría `me` | Qué pasa sin ella |
|---|---|
| `biso ls --mine` | Error 6: `error: --mine needs an identity; set it with biso config set me <you> or BISO_ME` |
| `biso start`, autoasignación | No asigna a nadie. Sale `note: no identity configured, task left unassigned` en vez del `note:` de siempre |
| Autor por defecto de un comentario | Error 2 si no se ha pasado `--comment-author`: `error: --comment-author is required, no identity is configured` |
| Autor de la pregunta, en `biso ask` | Error 2: `error: biso ask needs an identity; set it with biso config set me <you> or BISO_ME` (["`biso ask`"](cmd/verbos-del-ciclo.md#biso-ask)) |
| Autor de la respuesta, en `biso answer` | Error 2: `error: biso answer needs an identity; set it with biso config set me <you> or BISO_ME` (["`biso answer`"](cmd/verbos-del-ciclo.md#biso-answer)) |
| La línea `you are` de `biso prime` | `you are     (not set)`, con una nota que remite a `biso config set me` |

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

`boards_root` es la raíz por defecto de la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md): el directorio donde `biso init` sin `--at` crea
los tableros nuevos. `boards_extra_roots` son raíces adicionales, para cuando algún tablero vive fuera
de `boards_root`.

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
  receta completa, con las dos preguntas que le hace al repositorio y las tres órdenes que le da, está
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
| `commit` | lista de argumentos | sí | la orden que guarda una revisión |
| `publish` | lista de argumentos | no | la orden que la publica, y sin ella `--vcs push` es error 2 |
| `ignore_file` | nombre de fichero | no | el fichero de exclusión que `biso init` escribe dentro del tablero (["`biso init`"](cmd/init.md)) |

En `commit` y en `publish`, `{message}` se sustituye por el mensaje que `biso` compone (`biso snapshot:
248 tasks`), y `{files}` por los tres ficheros que entran en la revisión, cada uno como un argumento
propio: `snapshot.ndjson`, `board.json` y el marcador `<id>.id`, siempre en ese orden y siempre como ruta
relativa al directorio del tablero, sin `./` delante.

**Son los tres que nombra la sección ["`biso snapshot`"](cmd/snapshot.md), y no los dos de la clave `files` del JSON de `biso
snapshot`**, que son solo los que ese comando escribe. Los dos nombres se parecen y los dos conjuntos son
distintos, así que conviene leerlos juntos antes de escribir una orden.

Un código de salida distinto de cero es error 7 con la clave `code` `vcs_commit_failed`. Cómo se ejecutan
las dos órdenes, qué se hace con lo que escriban y por qué no hay tiempo máximo de espera lo dice la
sección ["`biso snapshot`"](cmd/snapshot.md), en un apartado que vale igual para `git` y para `custom`.

---

