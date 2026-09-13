# `biso init`

## Firma

```
biso init [<name>] [--at <dir>] [--statuses <list>]
          [--initial-status <status>] [--active-status <status>]
          [--terminal-status <status>] [--types <list>] [--priorities <list>]
          [--projects <list>] [--extensions <list>] [--prefix <text>]
          [--overwrite-config] [--from <location>]
```

## Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<name>` | | no | texto | el nombre del proyecto | no | no | |
| `--at <dir>` | | no | ruta de un directorio | una carpeta nueva en la raíz por defecto de la máquina (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)) | no | no | |
| `--statuses <list>` | | no | lista | `To Do, In Progress, Done` | sí | sí | |
| `--initial-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--active-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--terminal-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--types <list>` | | no | lista | `task, bug, docs` | sí | sí | |
| `--priorities <list>` | | no | lista | `high, medium, low` | sí | sí | |
| `--projects <list>` | | no | lista | vacía | sí | sí | |
| `--extensions <list>` | | no | lista | vacía | sí | sí | |
| `--prefix <text>` | | no | texto de solo letras | se deriva de `<name>` en mayúsculas (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)) | no | no | |
| `--overwrite-config` | | no | booleano | falso | no | no | |
| `--from <location>` | | no | ruta de un directorio | | no | no | `<name>`, `--statuses`, `--initial-status`, `--active-status`, `--terminal-status`, `--types`, `--priorities`, `--projects`, `--extensions`, `--prefix`, `--overwrite-config` |

**`--at` es la ruta del directorio del tablero que se va a crear, no el directorio donde se crea.** Con
`--at tablero` el tablero queda en `tablero`, no en `tablero/my-project-3f9a2b1c`. Es la misma convención que
la clave `path` del puntero, que también nombra el directorio del tablero y no el que lo contiene
(sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), y no es casualidad: la ruta que recibe `--at` es exactamente la que se escribe en esa
clave, con la misma forma. Puede ser absoluta o relativa al directorio de trabajo, y **su último componente es el nombre de la carpeta, que
es decorativo** (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), así que `--at tablero` es tan válido como `--at my-project-3f9a2b1c`.

## Comportamiento

Crea un tablero vacío con su configuración. **No escribe nunca fuera del tablero**, salvo el puntero
del proyecto que se describe a continuación.

**`<name>` es el `project_name` inicial del tablero.** Cambiarlo más adelante es cosa de `biso config
set project_name`, que no toca el sistema de ficheros (sección ["`biso config`"](config.md)).

`init` escribe, además del tablero, **el puntero del proyecto** (el fichero `.biso.json` de la
sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), y lo escribe siempre que no exista ya uno, porque un tablero no se localiza nunca por su
posición en el disco sino por una de las dos vías de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md). Es la única cosa que `init`
escribe fuera del tablero. La salida siempre confirma que el proyecto apunta al tablero, se haya escrito
el puntero en esta llamada o ya estuviera ahí de antes.

**Sin `--at`, el puntero no lleva clave `path`**: el tablero va a la raíz por defecto de la máquina y
ahí lo encuentra la búsqueda por marcador de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md), así que escribir su ruta sería guardar un
dato que nadie necesita y que dejaría de valer al cambiar `boards_root`.

**Con `--at`, el puntero lleva la ruta en `path`, con la misma forma en que se dio `--at`**: relativa si
`--at` era relativa, absoluta si era absoluta. `--at tablero` escribe `"path": "tablero"`;
`--at /Users/avilches/Hub/Projects/My project/tablero` escribe esa ruta completa. No hay flag para elegir la
forma porque la forma de `--at` ya es la elección, y esa elección es de quien llama y no del programa,
porque depende de un dato que el programa no tiene: **dónde van a vivir las demás copias de trabajo del
proyecto.**

La consecuencia práctica hay que decirla, porque es lo único que distingue a las dos formas. Una `path`
relativa se resuelve contra el directorio del puntero y, si ahí no hay tablero, contra sus ancestros
(sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), así que vale en cualquier copia de trabajo que tenga el tablero dentro o por encima: es
lo que quiere quien pone el tablero dentro del proyecto y trabaja en copias que también viven dentro del
proyecto, como los worktrees de git en `.claude/worktrees/`. Una `path` absoluta vale desde cualquier
sitio del disco mientras el proyecto no se mueva, y es lo que hace falta cuando las copias de trabajo
viven fuera del proyecto: ahí el directorio del tablero no está ni en la copia ni en ninguno de sus
ancestros, y una ruta relativa no lo alcanzaría. La elección, entonces, es entre sobrevivir a que el
proyecto se mueva y sobrevivir a que se trabaje desde fuera de él, y solo quien llama sabe cuál de las
dos le pasa.

**Por eso `init` lo dice cuando guarda una ruta relativa**, y solo entonces, con esta nota por stderr:

```
note: the location is stored as the relative path "tablero". A working copy
      outside this project will not have that folder while git ignores it, so
      it will not find the board: use an absolute --at if you work that way
```

La condición que la nota nombra es la exacta, y las dos mitades hacen falta: una copia de trabajo que
viva fuera del proyecto solo se queda sin el directorio del tablero **porque git lo ignora**, que es lo
que la otra nota de este comando recomienda hacer. Sin lo segundo, la carpeta viajaría con la copia y la
ruta relativa resolvería. La nota no dice "considera usar rutas absolutas" a secas porque quien lo leyera
no sabría si le aplica: diciendo qué se ha guardado y qué no va a funcionar, quien trabaja así se
reconoce y el resto puede seguir.

**Un `--at` que caiga dentro del proyecto está permitido, y entonces `init` lo dice.** Es una
configuración legítima, para quien quiera que su tablero viva junto a su proyecto y viaje en la misma
copia de seguridad, y sigue funcionando igual porque la resolución del tablero no depende de dónde esté
la carpeta. Es el caso al que sirve la `path` relativa del párrafo anterior, y el que la búsqueda por
ancestros de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) mantiene alcanzable desde una copia de trabajo que no tenga el directorio
del tablero. Lo que queda por decidir es si el proyecto versiona esa carpeta o la ignora, y **las dos
cosas funcionan y llevan a sitios distintos**, así que `init` no recomienda ninguna: dice cuáles son,
por stderr y con el nombre real de la carpeta, que es el que se le acaba de dar en `--at`.

```
note: the board lives inside this project. Ignore tablero/ and the board keeps
      its own history; version it and the snapshot travels with your code.
      Either way the database stays out, tablero/.gitignore excludes it
```

Es una nota y no un aviso porque no hay nada mal hecho, y `init` **no toca el fichero de exclusión del
proyecto** en ninguno de los dos casos: el puntero sigue siendo la única cosa que este comando escribe
fuera del tablero, y tocar la configuración de un proyecto ajeno sería pasarse de ahí.

**La última línea de esa nota nombra el fichero de exclusión que `init` acaba de escribir dentro del
tablero**, que es el del sistema configurado en la máquina, `.gitignore` con el valor por defecto. Con
`vcs` en `none` no se escribe ninguno, así que esa línea no sale y la nota se queda en las dos primeras:
sin fichero de exclusión, quien versione la carpeta se llevaría también la base de datos, y decir lo
contrario sería mentir.

**Las dos salidas de esa elección cambian dónde acaba el historial**, y por eso merece la pena decirlas
juntas. Ignorar la carpeta deja al tablero con su propio repositorio, el que `biso snapshot` crea de
forma perezosa la primera vez que corre ahí (["`biso snapshot`"](snapshot.md)), y publicarlo es entonces un trabajo aparte.
Versionarla mete los dos ficheros de la instantánea en el repositorio del código, así que `snapshot`
guarda su revisión ahí mismo, junto a los cambios del proyecto, y la instantánea cruza a otra máquina
con él sin que nadie configure nada. La base de datos no entra en ninguno de los dos casos.

**Con `--at` relativo dentro del proyecto salen las dos notas, en este orden**, y se quedan separadas
porque dicen cosas de naturaleza distinta: la primera describe una elección que se toma ahora, y la
segunda avisa de una consecuencia que solo le ocurrirá a quien trabaje desde fuera del proyecto.
Juntarlas en un solo mensaje haría que quien no está en ese caso tuviera que leer la condición para
descartarla.

**Y hay que decir qué pasa en la mitad que versiona la carpeta**, porque tiene una consecuencia que no
se ve venir. El fichero de exclusión que `init` escribe dentro del tablero excluye siempre la base de
datos, así que versionar el directorio del tablero versiona su marcador y sus dos ficheros de texto,
pero nunca `board.db`. Una copia de trabajo recibiría entonces un directorio con el marcador correcto y sin base de
datos: **eso no es un tablero**, y la resolución de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) no lo acepta como tal, sigue buscando
en los ancestros y en las raíces, y así encuentra el tablero de verdad. Si no lo encuentra en ninguna
parte, el error nombra ese directorio a medias, porque es la pista de lo que ha pasado.

**Si ya existe un puntero pero el tablero que nombra no está en esta máquina** (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), `init`
no acuña un `id` nuevo: usa el que ya lleva el puntero, para que las dos máquinas sigan hablando del
mismo tablero. Y no reescribe el puntero, porque ya era correcto.

**Sin `--statuses`**, el tablero nace con `To Do, In Progress, Done`, con los papeles inicial, activo y
terminal en ese orden. **Con `--statuses`**, hacen falta los tres flags de papel,
`--initial-status`, `--active-status` y `--terminal-status`, con los mismos nombres que las claves de
configuración a las que corresponden.

**`--from <location>` restaura una instantánea, en vez de crear un tablero en blanco.** `<location>`
es el directorio de un tablero que ha escrito ["`biso snapshot`"](snapshot.md), es decir, el que
contiene `snapshot.ndjson` y `board.json`. En una sola invocación, `init --from` hace lo que sería
crear el tablero con la configuración de `board.json` e importar `snapshot.ndjson` con las mismas
reglas del lote de `biso new --from` (sección ["`biso new`"](new.md)): valida el fichero de tareas entero contra el
vocabulario de `board.json` antes de escribir nada y, solo si todo es válido, escribe primero la
configuración y después las tareas. Como `board.json` ya trae el nombre del tablero, los estados,
los tipos, las prioridades, los proyectos, las extensiones y el prefijo del tablero de origen,
**`--from` es incompatible con `<name>` y con cualquier flag de vocabulario**: no hay nada que
decidir, todo viene del fichero. `--at` sigue valiendo igual que en un `init` normal, porque gobierna
dónde queda el tablero nuevo, no su vocabulario. **`--overwrite-config` en cambio es incompatible con
`--from`.** No es una restricción arbitraria: `--overwrite-config` reescribe la configuración de un
tablero que ya existe sin tocar sus tareas, y `--from` restaura un tablero entero, configuración y
tareas, en uno nuevo. Combinar las dos sería importar las tareas de la instantánea en un tablero que
ya tiene las suyas mientras se le cambia el vocabulario, y eso no es restaurar: es dejar el destino
como una segunda copia viva del tablero de origen, escritas las dos por separado, que es justo lo que
esta decisión de persistencia rechaza. Importar tareas en un tablero que ya las tiene sí está
permitido, y lo hace `biso new --from`; lo que no existe es la copia paralela. Si el destino de
`--from` ya tiene un tablero, ese caso ya está cubierto por la primera fila de la tabla siguiente: es
el mismo Error 2 de "ya hay uno accesible desde aquí", y no hace falta `--overwrite-config` para
distinguirlo porque `--from` siempre crea un tablero nuevo, nunca reescribe uno existente.

**El error de que ya hay un tablero es el más probable de este comando, y este es su mensaje**, con
código 2 y la clave `code` `board_exists`. Nombra el tablero y su ruta, porque lo que quien llama
necesita saber es cuál se ha encontrado, y remite a los dos caminos que hay desde ahí:

```
error: this project already has board 3f9a2b1c, at /Users/avilches/.biso/boards/my-project-3f9a2b1c
hint: `biso where` says which rule picked it
hint: --overwrite-config rewrites its configuration and never touches its tasks
```

**`--dry-run` vale en este comando** (sección ["Flags globales"](flags-globales.md#flags-globales)), y es donde más sirve: valida los argumentos y, con
`--from`, la instantánea entera contra el vocabulario que ella misma trae, sin crear ni escribir nada,
y sale 0 si habría funcionado y 7 si no. `--print`, en cambio, es error 2, porque ninguna tarea que
existiera antes queda afectada.

| Caso | Qué pasa |
|---|---|
| Ya hay un tablero accesible desde aquí | Error 2, salvo con `--overwrite-config`, que reescribe la configuración y **nunca toca las tareas** |
| El directorio de destino tiene una base de datos que no se puede leer (sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | No cuenta como tablero accesible, así que `--from` reconstruye ahí mismo adoptando el `id` del marcador, código 0. Es el remedio que el `hint` del error 21 nombra, y también lo que necesita un clon traído a otra máquina, que llega con la carpeta versionada y sin base de datos |
| El directorio de trabajo es ya el directorio de un tablero | Es el caso de la fila de arriba, alcanzado por la primera vía de ["Cómo se elige el tablero"](../resolucion-del-tablero.md), y se resuelve igual: Error 2, y con `--overwrite-config` se reescribe la configuración de ese tablero, que es exactamente lo que ese flag significa. Un tablero no se crea nunca dentro de otro |
| Ya hay un puntero, pero el tablero que nombra no está en esta máquina | No es un error: se crea el tablero adoptando el `id` que el puntero ya lleva, y el puntero no se reescribe porque ya era correcto, código 0 |
| `--at` a un directorio que ya es el directorio de un tablero | Error 2, con el mismo motivo visto desde el otro lado: el destino ya es un tablero |
| `--at` con una ruta relativa | No es un error: el tablero se crea ahí y el puntero lleva esa misma ruta relativa, código 0 |
| `--at` con una ruta absoluta | No es un error: el tablero se crea ahí y el puntero lleva esa misma ruta absoluta, código 0 |
| `--at` con una ruta relativa que sale del proyecto, como `../tableros/my-project` | No es un error, y el puntero la guarda tal cual: resuelve mientras la posición relativa entre el puntero y el tablero se mantenga, y el marcador confirma que el directorio al que llega es el tablero que el `id` nombra |
| `--overwrite-config` sobre un tablero con alguna tarea, si el prefijo resultante (el de `--prefix`, o el que se derive de `<name>` cuando no se da) no coincide con el `task_prefix` que el tablero ya tiene | Error 6, la misma inmutabilidad que la sección ["`biso config`"](config.md) aplica a `task_prefix` |
| Falta alguno de los tres flags de papel, habiendo `--statuses` | Error 2, con las tres nombradas y cuáles faltan |
| Un flag de papel sin `--statuses` | Error 2, diciendo que los papeles solo se fijan junto a la lista de estados |
| Un flag de papel nombra un estado que no está en `--statuses` | Error 2, con el valor y la lista de estados |
| Varios flags de papel nombran el mismo estado | Error 2, con los papeles y el estado que comparten |
| `--statuses` con menos de tres estados | Error 2, diciendo cuántos hacen falta y por qué |
| `--prefix` con algo que no sean letras | Error 2, `code` `invalid_prefix` |
| Sin `--prefix`, el nombre del tablero no deja ninguna letra al derivar el prefijo (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)) | Error 2, `code` `invalid_prefix`, pidiendo `--prefix` explícito |
| `--at` a un directorio donde no se puede escribir | Error 8 |
| `--from` junto con `<name>`, con cualquier flag de vocabulario, o con `--overwrite-config` | Error 2 |
| `--from` a un directorio al que le falta `snapshot.ndjson`, `board.json`, o los dos (una instantánea a medias) | Error 4, `code` `file_not_found`, nombrando qué fichero falta |
| `--from` cuyo `board.json` no se puede interpretar como JSON, o lleva una clave desconocida | Error 2, `code` `invalid_snapshot_config` |
| `--from` cuyo `board.json` tiene el mismo problema que haría fallar con Error 2 al flag de vocabulario equivalente (por ejemplo, `statuses` con menos de tres elementos, o un `task_prefix` sin letras) | Error 2, con el mismo `code` que usaría ese flag |
| `--from` cuyo `snapshot.ndjson` está vacío (una instantánea con configuración pero sin tareas) | No es un error: se crea el tablero con esa configuración y cero tareas, código 0 |
| `--from` cuyo `board.json` declara un vocabulario que ninguna tarea de `snapshot.ndjson` usa | No es un error: el tablero se crea con ese vocabulario tal cual lo declara `board.json`, tenga tareas que lo usen entero o no |
| `--from` cuyo `board.json` trae `me` o `default_limit`, porque se escribió a mano o con una versión anterior | No es un error, y tampoco se importan: `biso snapshot` no las escribe y `init --from` no las lee, con `note: me and default_limit are not restored, they belong to whoever uses the board`. No son claves desconocidas, así que no caen en el error 2 de la fila de arriba |
| `--from` cuyas tareas usan un valor, una clave de extensión o un `id` que `board.json` no hace válido | Error 7, la misma regla del lote de `biso new --from` (sección ["`biso new`"](new.md)), con el detalle de qué falta línea a línea |

**Los tres estados especiales se guardan como valores explícitos en la configuración, no como
posiciones.** Cambiar `statuses` después no los mueve nunca. Si al cambiar `statuses` uno de los tres
deja de existir, el comando que lo hace falla, según la sección ["`biso config`"](config.md).

**El fichero de la base de datos se llama `board.db`**, con `board.db-wal` y `board.db-shm` como sus
ficheros auxiliares. El nombre es fijo y forma parte de la interfaz, no un detalle interno, porque es
lo que hace reconocible un directorio de tablero: la primera vía de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) se apoya en él, y
sin un nombre declarado esa vía no sería implementable de una sola manera.

**Junto a él, `init` escribe el marcador de identidad `<id>.id`**, por ejemplo `3f9a2b1c.id`, que es lo
que permite encontrar el tablero por su identificador leyendo nombres de un directorio, sin abrir
ninguna base de datos (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)). Su nombre es el dato; su contenido es
`{ "storeVersion": 1 }`, la versión del formato del almacén, que no repite el identificador para que no
haya dos sitios donde pueda decir cosas distintas. El mismo identificador está guardado dentro de la
base de datos, y esa redundancia es a propósito: es la que permite comprobar que el marcador de un
directorio corresponde de verdad al tablero que contiene, cosa que `biso doctor` hace (sección ["`biso doctor`"](doctor.md#qué-comprueba)). Los dos
ficheros, la base de datos y el marcador, son lo único que hace falta para que un directorio sea un
tablero.

**El directorio del tablero es también, si es posible, su propio repositorio, pero `init` no lo crea.**
Lo que `init` sí escribe es el **fichero de exclusión del sistema de control de versiones que la máquina tenga
configurado** (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)), con `board.db` y sus dos ficheros auxiliares dentro, dejándolo listo para
el día en que el directorio llegue a ser un repositorio: lo que se versiona entonces es
`snapshot.ndjson`, `board.json` y el marcador, nunca el binario. Con `git`, ese fichero es
`.gitignore`; con `custom`, el que declare la clave `ignore_file`, y ninguno si no la declara; con
`none`, ninguno, porque no hay nada de lo que excluirse. Escribir un fichero de texto no es ejecutar
ningún programa, así que esto no contradice que `biso snapshot` sea el único que lo hace.

**Si alguien cambia la clave `vcs` después, ese fichero se queda con el nombre del sistema anterior**, y
`biso` no lo renombra ni escribe otro: `init` solo se ejecuta una vez por tablero, y adivinar cuándo hay
que reescribir un fichero de exclusión ajeno sería pasarse. Quien cambie de sistema tiene que escribir a
mano el fichero que el nuevo espere, con las tres líneas de la base de datos. `biso doctor` avisa de
este desajuste cuando lo puede reconocer (sección ["`biso doctor`"](doctor.md#qué-comprueba)).

**El marcador no se excluye, y eso resuelve algo.** Al quedar versionado con la instantánea, el
identificador del tablero viaja en ella, así que restaurar con `biso init --from <instantánea>` puede
recuperar la identidad y no solo los datos: el tablero restaurado adopta el `id` que el marcador de la
instantánea nombra, en vez de acuñar uno nuevo, y por eso el puntero commiteado del proyecto sigue
valiendo después de restaurar. Sin eso, restaurar en una máquina nueva daba un tablero correcto que el
proyecto no podía encontrar. Si el `id` de la instantánea ya existe en esta máquina, es el error de
identidad duplicada de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md), no una adopción silenciosa.

**Es `biso snapshot`, no `init`, quien convierte el directorio en un repositorio, y lo hace de forma
perezosa**: la primera vez que `snapshot` corre sobre un directorio que no está en ninguno, crea ahí el
del sistema de control de versiones configurado antes de guardar la revisión (sección ["`biso snapshot`"](snapshot.md)). Y si el
directorio ya está dentro del repositorio del proyecto, que es el caso de un tablero versionado con el
código, no crea ninguno: la revisión va a ese. La base de datos no se versiona nunca, ni siquiera
después de eso, porque cada escritura suya reescribe páginas internas: cada revisión guardaría una copia
completa y ningún sistema podría diferenciarla de una forma legible.

**El repositorio es opcional y su ausencia no rompe nada.** Si el sistema configurado no está instalado,
`snapshot` sigue escribiendo sus dos ficheros igual y sirviendo para restaurar con `--from`; lo único
que se pierde es el historial. `biso` no puede exigir que haya un sistema de control de versiones instalado, así
que esto nunca hace fallar ni a `init` ni a `snapshot` por esta sola razón (la sección ["`biso snapshot`"](snapshot.md) sí
distingue un fallo de entorno una vez que la revisión se intenta de verdad, con un repositorio ya
existente). Quien quiera dar historial a un tablero que nació sin él puede crear el repositorio a mano
en su directorio en cualquier momento: la siguiente instantánea lo detecta y empieza a guardar
revisiones.

## Salida

Esto es lo que imprime la tercera invocación de los ejemplos de ayuda,
`biso init "My project" --prefix MYP --at my-project-board --extensions trello.card` (con `--json` para el
esquema de más abajo). El prefijo sale `MYP` porque lo fija `--prefix`, no porque se derive del nombre
`My project`, que sin ese flag daría `MYPROJECT` (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)).

```
Created board "My project"
  statuses    To Do (initial) | In Progress (active) | Done (terminal)
  types       task, bug, docs
  priorities  high, medium, low
  prefix      MYP
This project now points at that board.
Run `biso prime` to see how to use it.
```

La línea "This project now points at that board." aparece siempre, porque el proyecto siempre queda
apuntando a ese tablero, se escriba el puntero en esta llamada o ya estuviera escrito de antes
(sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)).

**Esta invocación emite además las dos notas**, porque `--at my-project-board` es una ruta relativa que cae
dentro del proyecto, que es justo el caso que las dispara. Por stderr sale esto, en este orden:

```
note: the board lives inside this project. Ignore my-project-board/ and the board keeps
      its own history; version it and the snapshot travels with your code.
      Either way the database stays out, my-project-board/.gitignore excludes it
note: the location is stored as the relative path "my-project-board". A working copy
      outside this project will not have that folder while git ignores it, so
      it will not find the board: use an absolute --at if you work that way
```

No están en el bloque de arriba porque ese bloque es stdout, y las notas van por stderr como todas
(sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)). El puntero que esta llamada escribe es
`{ "version": 1, "id": "3f9a2b1c", "path": "my-project-board" }`.

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "init",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "board": {
      "name": "My project",
      "statuses": ["To Do", "In Progress", "Done"],
      "initialStatus": "To Do", "activeStatus": "In Progress", "terminalStatus": "Done",
      "types": ["task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "taskPrefix": "MYP"
    },
    "pointerCreated": true
  }
}
```

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Tablero creado, o restaurado con `--from` | 0 |
| Ya existía y no hay `--overwrite-config` | 2 |
| Argumentos inválidos, incluido un `board.json` de `--from` inválido, o `--from` junto con `--overwrite-config` | 2 |
| `--from` a un directorio sin `snapshot.ndjson`, sin `board.json`, o sin los dos | 4 |
| `--overwrite-config` cambiaría `task_prefix` con tareas ya creadas | 6 |
| El lote de `snapshot.ndjson` de `--from` falla su validación | 7 |
| No se puede escribir | 8 |

## `biso init --help`

```
Usage: biso init [name] [options]

Create a task board for this project. It writes the board and a pointer
inside the project so every copy of the project finds the same board. It
never writes outside the board otherwise.

Arguments:
  name                   board name (default: the project directory name)

Options:
  --at <dir>                  the board's own directory, not where to put it
                              (default: a new folder in the machine's default
                              boards root). A relative path is stored relative
                              to the pointer; an absolute one is stored as is
  --statuses <list>           comma-separated, at least three
                              (default: "To Do,In Progress,Done")
  --initial-status <status>   status of a new task (default: "To Do")
  --active-status <status>    what `biso start` sets (default: "In Progress")
  --terminal-status <status>  what `biso finish` sets (default: "Done")
  --types <list>              comma-separated (default: "task,bug,docs")
  --priorities <list>         comma-separated (default: "high,medium,low")
  --projects <list>           comma-separated (default: none)
  --extensions <list>         comma-separated declared external field keys,
                              such as trello.card (default: none)
  --prefix <text>             task id prefix, letters only (default: derived
                              from the board name, uppercased)
  --overwrite-config          replace the configuration of an existing board,
                              keeping every task
  --from <location>           restore a snapshot: the directory where `biso
                              snapshot` wrote snapshot.ndjson and board.json.
                              Incompatible with name and with every vocabulary
                              option (which all come from board.json
                              instead), and with --overwrite-config: restoring
                              always creates a new board
  -h, --help                  show this help

`--initial-status`, `--active-status` and `--terminal-status` each name one of
`--statuses`, all three distinct. Giving `--statuses` requires the three
together; giving any of them without `--statuses` is bad usage. They are then
stored as explicit values and never move again.

With --at the board can live inside the project itself, which is fine. Two
things follow, and `init` says both when it applies. Decide whether the
project ignores that folder or versions it: ignore it and the board keeps a
history of its own, version it and its snapshot travels with your code. The
database stays out either way. And mind the form of the path: a relative --at
is stored relative to the pointer and resolves from any working copy that has
the folder inside it or above it, which is the case for worktrees kept under
the project; a working copy that lives outside the project has no such
folder, precisely because it is ignored, so pass an absolute --at if you work
that way.

The board directory can also become a repository of its own, but `init` does
not create it: `init` only writes the ignore file of the version control
system this machine is configured for (the vcs key, .gitignore with git),
holding the database file and its WAL auxiliaries, so that once a repository
exists only snapshot.ndjson, board.json and the <id>.id marker are ever
versioned. `biso snapshot` is the one that creates that repository, lazily,
the first time it runs against a directory that is in none (see `biso
snapshot --help`). Missing version control never fails `init` or `snapshot`;
the board works the same, only its history is lost.

Exit codes:
  0  board created, or restored with --from
  2  bad usage, or a board is already reachable from here
  4  --from points at a directory missing snapshot.ndjson, board.json, or both
  6  --overwrite-config would change task_prefix on a board with tasks
  7  --from's snapshot.ndjson failed batch validation
  8  cannot write there

Examples:
  biso init
  biso init "My project" --statuses "Ideas,To Do,In Progress,Done" \
      --initial-status Ideas --active-status "In Progress" \
      --terminal-status Done
  biso init "My project" --prefix MYP --at my-project-board --extensions trello.card
  biso init --at /tmp/tablero-nuevo --from ~/.biso/boards/my-project-3f9a2b1c
```

El segundo ejemplo deja `To Do` sin ningún papel a propósito: un tablero puede llevar estados que no
son ni el inicial, ni el activo, ni el terminal, y eso no rompe nada.

---

