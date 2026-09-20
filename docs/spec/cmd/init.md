# `biso init`

## Firma

```
biso init [<name>] [--at <dir>] [--statuses <list>]
          [--initial-status <status>] [--active-status <status>]
          [--terminal-status <status>] [--types <list>] [--priorities <list>]
          [--extensions <list>] [--prefix <text>]
          [--overwrite-config] [--from <location>]
```

## Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<name>` | | no | texto | el nombre del proyecto | no | no | |
| `--at <dir>` | | no | ruta de un directorio | una carpeta nueva en la raíz por defecto de la máquina (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)) | no | no | |
| `--statuses <list>` | | no | lista | `To Do, In Progress, Done` ¹ | sí | sí | |
| `--initial-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--active-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--terminal-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--types <list>` | | no | lista | `task, bug, docs` ¹ | sí | sí | |
| `--priorities <list>` | | no | lista | `high, medium, low` ¹ | sí | sí | |
| `--extensions <list>` | | no | lista | vacía ¹ | sí | sí | |
| `--prefix <text>` | | no | texto de solo letras | se deriva de `<name>` en mayúsculas (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)) ¹ | no | no | |
| `--overwrite-config` | | no | booleano | falso | no | no | |
| `--from <location>` | | no | ruta de un directorio | | no | no | `<name>`<br>`--statuses`<br>`--initial-status`<br>`--active-status`<br>`--terminal-status`<br>`--types`<br>`--priorities`<br>`--extensions`<br>`--prefix`<br>`--overwrite-config` |

¹ Solo aplica al crear un tablero nuevo. Con `--overwrite-config` sobre un tablero que ya existe,
no pasar este flag no vuelve a este valor por defecto: conserva el valor que el tablero ya tenía
para esa clave (tabla de casos más abajo).

**Repetible** dice si el flag se puede dar más de una vez en la misma llamada, cada vez se acumula.
**Lista** dice si, además, admite varios valores separados por coma dentro de un solo `--flag a,b`. En
esta tabla las dos siempre coinciden porque `--statuses`, `--types`, `--priorities` y `--extensions`
son todas listas de tokens que admiten las dos formas a la vez; no es la regla general de la
especificación, donde hay flags repetibles que no admiten coma, como `--comment` (sección
["`biso set`"](set.md)). El detalle exacto de cómo se acumulan las dos formas está en
["Repetición y listas separadas por comas"](../valores-de-entrada.md#repetición-y-listas-separadas-por-comas).

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

**Si el proyecto versiona la carpeta del tablero, la base de datos no viaja con ella**, y eso tiene una
consecuencia que no se ve venir. El fichero de exclusión que `init` escribe dentro del tablero excluye siempre la base de
datos, así que versionar el directorio del tablero versiona su marcador y sus dos ficheros de texto,
pero nunca `board.db`. Una copia de trabajo recibiría entonces un directorio con el marcador correcto y sin base de
datos: **eso no es un tablero**, y la resolución de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md) no lo acepta como tal, sigue buscando
en los ancestros y en las raíces, y así encuentra el tablero de verdad. Si no lo encuentra en ninguna
parte, el error nombra ese directorio a medias, porque es la pista de lo que ha pasado. Ese error y sus
dos remedios, según si el directorio también lleva la instantánea o solo el marcador, están en la
sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md#el-puntero-nombra-un-tablero-que-no-está-en-esta-máquina).

**Un directorio de destino cuya base de datos está pero no se puede leer se trata igual que uno al que
le falta**, que es lo que dice la tabla de abajo, en sus filas sobre un destino sin base de datos
legible, y la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar): no
cuenta como tablero accesible, así que `init` no da el Error 2 de "ya hay uno" sino que crea el tablero
ahí mismo, adoptando el `id` del marcador, con código 0. **Para poder hacerlo, sustituye el fichero**:
borra `board.db` y sus dos auxiliares y escribe una base de datos nueva encima. Es lo único que este
comando borra en toda la especificación, y no hay otra forma de cumplir lo que esa tabla promete,
porque reconstruir en el sitio (el remedio que el `hint` del error 21 manda teclear) es exactamente
poner una base de datos donde estaba la que no se puede leer. Un fichero que el programa no sabe leer
tampoco es uno que pueda conservar, y quien tenga una instantánea recupera además las tareas con
`--from` en vez de quedarse con un tablero vacío.

**Si ya existe un puntero pero el tablero que nombra no está en esta máquina** (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), `init`
no acuña un `id` nuevo: usa el que ya lleva el puntero, para que las dos máquinas sigan hablando del
mismo tablero. Y no reescribe el puntero, porque ya era correcto.

**Hay una excepción hermana, al revés: cuando este proyecto no resuelve ya, por su cuenta, a ningún
tablero accesible, y `--at` nombra un directorio que ya es un tablero íntegro, con `board.db` legible.**
Es lo que pasa cuando no hay ningún puntero aquí, o cuando el que hay no resuelve a nada en esta
máquina (los dos remedios de la sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md#el-puntero-se-pierde), "El puntero se pierde" y
["La ruta relativa o absoluta"](../resolucion-del-tablero.md#la-ruta-relativa-o-absoluta)). Ahí tampoco `init` acuña un `id` nuevo: adopta el que ya lleva el marcador del destino. Pero a
diferencia del caso anterior, esta vez sí escribe, o reescribe, el puntero de este proyecto con ese `id`
y con la `path` que apunta a ese destino, porque no había ningún puntero correcto que conservar. No
toca ni la configuración ni las tareas del tablero destino: la operación es puramente local a este
proyecto, indicarle dónde está su tablero. **Esta excepción no se aplica cuando el proyecto que llama ya
resuelve, por su cuenta, a un tablero accesible**, aunque sea distinto del destino: ahí sigue siendo el
Error 2 de la fila "Ya hay un tablero accesible desde aquí", solo que visto desde el lado del destino en
vez de desde el lado de este proyecto.

**Las dos excepciones son la misma regla, y así es como se decide siempre si el puntero se escribe:
`init` lo escribe cuando lo que hay apuntado no encontraría el tablero que esta llamada acaba de
crear o de adoptar, y lo deja en paz cuando sí lo encontraría.** En el primer caso el puntero ya
nombraba ese `id` y el tablero nace donde ese puntero lo iba a buscar, así que no hay nada que
corregir; en el segundo no había ningún puntero correcto que conservar. Y queda un tercer caso que
las dos frases anteriores no cubrían por separado: un puntero con el `id` correcto cuya clave `path`
ya no nombra dónde ha ido a parar el tablero, que es lo que pasa al mover un tablero con
`biso init --at <ruta nueva>` (sección ["La ruta relativa o absoluta"](../resolucion-del-tablero.md#la-ruta-relativa-o-absoluta)). Ahí el puntero se reescribe, por la
misma razón: dejarlo como está sería dejar al proyecto sin poder encontrar su tablero. En la práctica
la comparación es entre el `id` y la `path` que el puntero tiene y los que tendría que tener, que son
el `id` en uso y el texto de `--at` tal cual se escribió, o ninguna `path` si no hubo `--at`.

**La regla vale igual con `--overwrite-config`**, que reescribe la configuración de un tablero que ya
existe en vez de crear uno: si al llamar no hay ningún puntero entre el directorio de trabajo y el
tope de la búsqueda, `init` lo escribe. Solo puede pasar llegando al tablero por la primera vía, es
decir estando dentro de su directorio, porque llegar por la segunda es precisamente tener un puntero.
La razón es la línea "This project now points at that board.", que sale siempre: dejar al proyecto sin
nada apuntado y decir a la vez que apunta sería mentir, y de las dos formas de arreglarlo (escribir el
puntero o callarse la línea) solo la primera deja al proyecto encontrando su tablero desde cualquier
otro sitio. La `path` que se escribe es la que haría falta para encontrarlo: ninguna si el tablero está
directamente bajo una de las raíces de la máquina, la de `--at` tal cual se escribió si esta llamada lo
pasó, y la ruta del directorio del tablero en cualquier otro caso.

La clave `pointerCreated` del sobre JSON dice exactamente eso: si esta llamada escribió el fichero.

**Sin `--statuses`**, el tablero nace con `To Do, In Progress, Done`, con los papeles inicial, activo y
terminal en ese orden. **Con `--statuses`**, hacen falta los tres flags de papel,
`--initial-status`, `--active-status` y `--terminal-status`, con los mismos nombres que las claves de
configuración a las que corresponden.

**`--from <location>` restaura una instantánea, en vez de crear un tablero en blanco.** `<location>`
es el directorio de un tablero que ha escrito ["`biso snapshot`"](snapshot.md), con los tres ficheros
de ["Qué entra en la revisión"](snapshot.md#qué-entra-en-la-revisión): `snapshot.ndjson`, `board.json`
y el marcador `<id>.id`. En una sola invocación, `init --from` hace lo que sería
crear el tablero con la configuración de `board.json` e importar `snapshot.ndjson` con las mismas
reglas del lote de `biso new --from` (sección ["`biso new`"](new.md)): valida el fichero de tareas entero contra el
vocabulario de `board.json` antes de escribir nada y, solo si todo es válido, escribe primero la
configuración y después las tareas. Como `board.json` ya trae el nombre del tablero, los estados,
los tipos, las prioridades, las extensiones y el prefijo del tablero de origen,
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

**La identidad que `--from` adopta es la del marcador de la instantánea, salvo que el destino ya
tenga uno propio, y entonces manda el del destino.** Los dos casos no compiten en la práctica, porque
el único destino que trae marcador es el tablero que se está reconstruyendo en su sitio, y ahí los
dos marcadores llevan el mismo `id`; la regla se escribe de todas formas para que la reconstrucción
en el sitio no dependa de que la instantánea que se le pase sea la suya.

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

**Sin `--from`, la frase de `--dry-run` es una sola nota por stderr**, porque no hay ninguna tarea que
enseñar en una línea de estado y lo único que hay que confirmar es que el tablero se habría creado:

```
$ biso init "My project" --dry-run
note: board would be created at /Users/avilches/.biso/boards/my-project-3f9a2b1c (--dry-run)
```

**Con `--overwrite-config`, esa nota diría algo falso**, porque ahí no se crea ningún tablero: el que
hay ya existe y lo que la llamada haría es reescribir su configuración. La frase es entonces esta otra,
y nombra el tablero por su `id` y por su ruta, como hace el error de que ya existe:

```
note: the configuration of board 3f9a2b1c at /Users/avilches/.biso/boards/my-project-3f9a2b1c
      would be rewritten, and no task would change (--dry-run)
```

El salto de línea está donde está, entre la ruta y el verbo, y no depende de lo larga que sea la ruta,
igual que en las dos notas de `--at`.

Con `--at`, la ruta que nombra la nota es la que `--at` da, con la misma forma (relativa o absoluta)
que llevaría el puntero si la llamada fuera real. **Con `--from`, en cambio, la frase es la del lote de
importación** (sección ["`biso new`"](new.md#el-modo-lote)), `<N> tasks would be created, nothing was
written (--dry-run)`, porque ahí sí hay tareas que contar; `init --from` no repite la nota del tablero
además de esa frase, para no decir dos veces que no se ha escrito nada.

| Caso | Qué pasa |
|---|---|
| Ya hay un tablero accesible desde aquí | Error 2, salvo con `--overwrite-config`, que reescribe la configuración y **nunca toca las tareas** |
| `--overwrite-config` sin uno de los flags de vocabulario (`--statuses` y sus tres papeles, `--types`, `--priorities`, `--extensions`) | No es un error: esa clave conserva el valor que el tablero ya tenía, igual que `project_name` sin `<name>` explícito (sección ["`biso config`"](config.md)). Como consecuencia, el prefijo sin `<name>` ni `--prefix` también se conserva, porque se deriva de `project_name`, que a su vez se conserva; no hace falta ningún caso especial para él |
| El directorio de destino no tiene una base de datos legible: le falta, o no se puede leer (sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | No cuenta como tablero accesible, así que `--from` reconstruye ahí mismo, con `--at` apuntando a ese mismo directorio, adoptando el `id` del marcador, código 0. Es el remedio que el `hint` del error 21 nombra, y también el que necesita un clon traído a otra máquina que llega con la carpeta del tablero versionada y sin base de datos: el mismo remedio lo repite el `hint` del error `pointer_unresolved` (código 20) cuando es la resolución normal, no `init`, quien encuentra ese directorio a medias (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md#el-puntero-nombra-un-tablero-que-no-está-en-esta-máquina)) |
| `--at`, sin `--from`, al mismo directorio de la fila anterior: con el marcador pero sin una base de datos legible | No cuenta como tablero accesible, igual que en la fila anterior: `init` lo crea ahí mismo, adoptando el `id` del marcador, código 0. Sin `--from` no hay instantánea que restaurar, así que el tablero nace vacío; es el remedio de un clon cuyo directorio del tablero se versionó antes de la primera instantánea |
| El directorio de trabajo es ya el directorio de un tablero | Es el caso de la fila "Ya hay un tablero accesible desde aquí", alcanzado por la primera vía de ["Cómo se elige el tablero"](../resolucion-del-tablero.md), y se resuelve igual: Error 2, y con `--overwrite-config` se reescribe la configuración de ese tablero, que es exactamente lo que ese flag significa. Un tablero no se crea nunca dentro de otro. Si la base de datos que ese directorio tiene no se puede leer, no es un tablero accesible y manda la fila de arriba: `init` lo crea ahí mismo, adoptando el `id` de su marcador, código 0 |
| Ya hay un puntero, pero el tablero que nombra no está en esta máquina | No es un error: se crea el tablero adoptando el `id` que el puntero ya lleva, y el puntero no se reescribe porque ya era correcto, código 0 |
| `--at` a un directorio que ya es el directorio de un tablero, con `board.db` legible | Depende de si este proyecto ya resuelve, por su cuenta, a un tablero accesible (fila "Ya hay un tablero accesible desde aquí"). Si lo resuelve, Error 2, con el mismo motivo visto desde el otro lado: el destino también es un tablero. Si no lo resuelve, porque no hay puntero aquí o el que hay no resuelve a nada en esta máquina, no es un error: adopta el `id` del marcador del destino y escribe, o reescribe, el puntero de este proyecto con ese `id` y esa `path`, sin tocar la configuración ni las tareas del destino, código 0 |
| `--at` con una ruta relativa | No es un error: el tablero se crea ahí y el puntero lleva esa misma ruta relativa, código 0 |
| `--at` con una ruta absoluta | No es un error: el tablero se crea ahí y el puntero lleva esa misma ruta absoluta, código 0 |
| `--at` con una ruta relativa que sale del proyecto, como `../tableros/my-project` | No es un error, y el puntero la guarda tal cual: resuelve mientras la posición relativa entre el puntero y el tablero se mantenga, y el marcador confirma que el directorio al que llega es el tablero que el `id` nombra |
| `--overwrite-config` sobre un tablero con alguna tarea, con `--prefix` explícito que no coincide con el `task_prefix` que el tablero ya tiene | Error 6, la misma inmutabilidad que la sección ["`biso config`"](config.md) aplica a `task_prefix`. Sin `--prefix`, el prefijo se conserva (fila de arriba) y este error no puede darse |
| `--overwrite-config` con `--statuses` explícito (y sus tres papeles) que dejaría el tablero inconsistente, sobre un tablero con tareas | Error 6, las mismas reglas de la tabla de casos de ["`biso config`"](config.md#comportamiento-caso-a-caso), aplicadas solo porque `--statuses` se pasó explícito: una clave conservada no puede quitar nada que ya estuviera en uso |
| `--overwrite-config` con `--types` explícito que quita un tipo en uso, sobre un tablero con tareas | Error 6, la misma regla de la tabla de casos de ["`biso config`"](config.md#comportamiento-caso-a-caso) |
| `--overwrite-config` con `--priorities` explícito que quita una prioridad en uso, sobre un tablero con tareas | Error 6, la misma regla de la tabla de casos de ["`biso config`"](config.md#comportamiento-caso-a-caso) |
| `--overwrite-config` con `--extensions` explícito que quita una clave en uso, sobre un tablero con tareas | Error 6, la misma regla de la tabla de casos de ["`biso config`"](config.md#comportamiento-caso-a-caso) |
| Falta alguno de los tres flags de papel, habiendo `--statuses` | Error 2, con los tres nombrados y cuáles faltan |
| Un flag de papel sin `--statuses` | Error 2, diciendo que los papeles solo se fijan junto a la lista de estados |
| Un flag de papel nombra un estado que no está en `--statuses` | Error 2, con el valor y la lista de estados |
| Varios flags de papel nombran el mismo estado | Error 2, con los papeles y el estado que comparten |
| `--statuses` con menos de tres estados | Error 2, diciendo cuántos hacen falta y por qué |
| `--prefix` con algo que no sean letras | Error 2, `code` `invalid_prefix` |
| Sin `--prefix`, el nombre del tablero no deja ninguna letra al derivar el prefijo (sección ["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)) | Error 2, `code` `invalid_prefix`, pidiendo `--prefix` explícito |
| El nombre del tablero no deja ninguna letra ni ningún dígito con los que derivar el slug de su carpeta (sección ["Cómo se deriva el nombre de la carpeta"](../resolucion-del-tablero.md#cómo-se-deriva-el-nombre-de-la-carpeta)), como `"///"` | Error 3, `code` `bad_config_value`, la misma regla que aplica ["`biso config`"](config.md) a `project_name`: el slug es un dato del tablero, así que un valor que no sirve no se acepta ni aquí ni ahí. Se comprueba aunque haya `--at`, porque el slug es del tablero y no de la carpeta que esta llamada le da |
| El mismo `id` aparece en dos raíces de la máquina | Error 22, `code` `ambiguous_board_id` (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md#el-mismo-id-en-dos-sitios)). `init` recorre las mismas raíces para comprobar que el `id` que va a acuñar o a adoptar no exista ya, así que da el mismo error que cualquier otro comando |
| `--at` a un directorio donde no se puede escribir | Error 8 |
| `--from` junto con `<name>`, con cualquier flag de vocabulario, o con `--overwrite-config` | Error 2 |
| `--from` a un directorio al que le falta `snapshot.ndjson`, `board.json`, el marcador `<id>.id`, o varios de los tres (una instantánea a medias) | Error 4, `code` `file_not_found`, nombrando qué fichero falta |
| `--from` cuyo `board.json` no se puede interpretar como JSON, o lleva una clave desconocida | Error 2, `code` `invalid_snapshot_config` |
| `--from` cuyo `<id>.id` no tiene el contenido `{ "storeVersion": 1 }` que esta misma página describe más arriba | Error 2, `code` `invalid_snapshot_id` |
| `--from` cuyo `<id>.id` nombra un identificador que ya existe en esta máquina | El error de identidad duplicada de ["Cómo se elige el tablero"](../resolucion-del-tablero.md), no una adopción silenciosa |
| `--from` cuyo `board.json` tiene el mismo problema que haría fallar con Error 2 al flag de vocabulario equivalente (por ejemplo, `statuses` con menos de tres elementos, o un `task_prefix` sin letras) | Error 2, con el mismo `code` que usaría ese flag |
| `--from` cuyo `snapshot.ndjson` está vacío (una instantánea con configuración pero sin tareas) | No es un error: se crea el tablero con esa configuración y cero tareas, código 0 |
| `--from` cuyo `board.json` declara un vocabulario que ninguna tarea de `snapshot.ndjson` usa | No es un error: el tablero se crea con ese vocabulario tal cual lo declara `board.json`, tenga tareas que lo usen entero o no |
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

**El contenido de ese fichero son tres líneas**, la base de datos y sus dos ficheros auxiliares, cada
uno en su línea y sin comodines, porque el nombre de los tres es fijo:

```
board.db
board.db-wal
board.db-shm
```

**Si el directorio ya tenía uno, `init` no lo toca.** Es lo que le pasa a un clon que trae versionada
la carpeta del tablero con su fichero de exclusión dentro, y reescribir el fichero de otro sería
pasarse igual que tocar el del proyecto.

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

**La primera línea dice cuál de los tres desenlaces ha ocurrido**, porque "Created" sería mentira
sobre un tablero que ya estaba ahí. Las filas de abajo y la línea del puntero son las mismas en los
tres:

| Desenlace | Primera línea | Última línea |
|---|---|---|
| Se ha creado un tablero, acuñando o adoptando su identidad | `Created board "My project"` | la de arriba, la que nombra `biso prime` |
| `--at` nombraba un tablero ya íntegro y este proyecto solo ha aprendido dónde está | `Adopted board "My project"` | la misma |
| `--overwrite-config` ha reescrito la configuración de un tablero que ya existía | `Rewrote the configuration of board "My project"` | no sale |

La línea de `biso prime` no sale en el tercero porque ese tablero ya estaba en uso: quien reescribe
su configuración no acaba de empezar con él.

**Con `--quiet`, stdout se queda vacío.** Ese flag reduce stdout a los identificadores de las tareas
afectadas (sección ["Flags globales"](flags-globales.md#flags-globales)), y este comando no afecta a
ninguna, así que no hay nada que imprimir. Las notas de más abajo tampoco salen, porque son `note:`.

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
| Tablero creado, restaurado con `--from`, o el puntero (re)escrito adoptando un tablero ya existente | 0 |
| Ya existía y no hay `--overwrite-config` | 2 |
| Argumentos inválidos, incluido un `board.json` de `--from` inválido, o `--from` junto con `--overwrite-config` | 2 |
| El nombre del tablero no deja ningún slug con el que nombrar su carpeta | 3 |
| `--from` a un directorio sin `snapshot.ndjson`, sin `board.json`, o sin los dos | 4 |
| `--overwrite-config` cambiaría `task_prefix` con tareas ya creadas | 6 |
| El lote de `snapshot.ndjson` de `--from` falla su validación | 7 |
| No se puede escribir | 8 |
| El mismo `id` está en dos raíces | 22 |

El 22 sale en esta tabla, y no solo en la de ["Códigos de salida"](../codigos-de-salida.md), porque
este comando recorre las raíces de la máquina por su cuenta, para comprobar que el `id` que va a
acuñar o a adoptar no exista ya, así que puede darlo incluso desde un proyecto que no tiene tablero
todavía.

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

