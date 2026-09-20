# `biso snapshot`

## Firma

```
biso snapshot [--vcs <mode>]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--vcs <mode>` | | no | `none`, `commit` o `push` | `commit` | no | no | ninguno |

**`biso snapshot` no cambia ningún dato del tablero**, así que la sección ["Flags globales"](flags-globales.md#flags-globales) lo clasifica junto a
`export` entre los comandos donde `--print` y `--dry-run` son error de uso con código 2: no hay
ninguna tarea afectada que imprimir, ni ninguna escritura de tarea que simular.

**No tiene flag `-o`/`--out`.** A diferencia de `export`, que escribe donde se le diga,
`snapshot` escribe siempre en el propio directorio del tablero (["`biso init`"](init.md)), con nombre fijo:
`snapshot.ndjson` y `board.json`. Es la instantánea del tablero para sí mismo, no un volcado a otra
parte; para volcar a otra parte está `export`.

## Qué escribe, y por qué esos dos ficheros

`snapshot.ndjson` tiene exactamente la forma que fija la garantía de simetría de `biso export` (["`biso export`"](export.md)):
una tarea por línea, con las mismas claves, incluidos los identificadores, las fechas y las claves de
criterio. `board.json` es la configuración del tablero, en la misma forma que imprime
`biso config list --json` (["`biso config`"](config.md)): todas las claves de vocabulario, `task_prefix` y las demás.
`me` y `default_limit` no están entre ellas, porque no son claves de la configuración de un tablero
(sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)), así que restaurar la instantánea de otra persona con `biso init --from`
nunca hereda su identidad ni su límite por defecto. La garantía de simetría de ["`biso export`"](export.md) cubre por
tanto la configuración del tablero entera, campo a campo.

Los dos ficheros son los que después lee `biso init --from` (["`biso init`"](init.md)) para reconstruir el tablero
entero. Van en dos ficheros separados, y no en uno solo, para que los diffs queden legibles: la
configuración cambia pocas veces y las tareas cambian todo el rato, así que mezclarlas habría hecho
que cada revisión de una tarea reescribiera también un bloque de configuración idéntico.

## Cómo se escriben, y qué pasa si el segundo falla

**Cada fichero se escribe en un temporal del mismo directorio y se renombra encima del anterior**, que
es una operación atómica del sistema de ficheros, así que nadie lee nunca medio `snapshot.ndjson`.

**Y los dos temporales se escriben completos antes de renombrar ninguno de los dos.** De ese orden
depende lo único que se puede prometer aquí: si falla la escritura, por disco lleno o por permisos, los
dos ficheros anteriores quedan intactos, no se intenta guardar ninguna revisión, y el código es 8. Entre
el primer renombrado y el segundo queda una ventana de un instante en la que el par no es coherente, y
esta especificación lo dice en vez de prometer una atomicidad de dos ficheros que el sistema de ficheros
no da.

**`snapshot` no toma ningún acceso exclusivo**, ni el de las escrituras de tareas ni uno propio. Es una
lectura del tablero, y las lecturas nunca bloquean a nadie (["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables), punto 6): una instantánea de un
tablero grande no puede hacer fallar a un `biso set` que llegue a la vez. Lo que sí puede pasar es que
dos instantáneas simultáneas choquen al guardar la revisión, porque el sistema de control de versiones
se protege con su propia marca de bloqueo; entonces una de las dos sale con código 8 y su mensaje dice
que basta volver a llamar. Los ficheros de las dos quedan enteros de todos modos, porque cada una
escribió su temporal.

## El sistema de control de versiones

El sistema lo dice la clave `vcs` de la configuración de máquina (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)), con `git` por defecto,
`none` para no ejecutar nada y `custom` para uno que `biso` no conoce. El flag `--vcs` de este
comando elige qué se hace en esta llamada, y no cambia esa configuración:

| `--vcs` | Qué hace |
|---|---|
| `none` | Escribe los dos ficheros y no ejecuta nada, ni siquiera para preguntar |
| `commit` (por defecto) | Escribe los dos ficheros y guarda una revisión con el sistema configurado |
| `push` | Lo anterior y además publica, con la orden de publicar de ese sistema |

Con `vcs` igual a `none` en la configuración, `commit` y `push` no tienen a quién pedírselo: los dos
escriben los ficheros y emiten `note: vcs is set to none, skipping the commit`, con código 0.

**Ningún otro comando de `biso` ejecuta nunca un programa ajeno.** Invocar el sistema de control de
versiones cuesta unos 12 milisegundos medidos, y el presupuesto de arranque de 25 milisegundos para
`biso ls` y `biso prime` sobre un tablero de 300 tareas (sección ["El presupuesto de arranque"](../presupuestos.md#el-presupuesto-de-arranque)) no admite ese coste en el camino
caliente de ningún comando. `biso snapshot` es la única excepción, precisamente porque quien lo llama ya
está pidiendo explícitamente esa operación; y es también, por lo mismo, el único comando que llega a
crear un repositorio: sin haber corrido nunca `biso snapshot` con `commit` o con `push` sobre él, el
directorio de un tablero nunca pasa a ser uno.

## Dónde va la revisión, y cómo se decide

El directorio del tablero puede estar en tres situaciones distintas, y **`snapshot` las distingue
preguntándole al propio sistema**, no adivinando por el disco. La receta de cada sistema del catálogo
trae esas dos preguntas: si hay un repositorio que contenga a este directorio y cuál es su raíz, y si
ese repositorio ignora la carpeta del tablero. Con eso:

| Situación | Dónde va la revisión |
|---|---|
| La raíz del repositorio es el propio directorio del tablero | Ahí. Es el caso de un tablero que ya tiene historial propio |
| La raíz está por encima y el repositorio **no** ignora la carpeta del tablero | En ese repositorio, que es el del proyecto. La instantánea queda versionada junto al código y viaja con él, con su remoto incluido |
| La raíz está por encima y el repositorio **sí** ignora la carpeta, o no hay ningún repositorio | En un repositorio propio del tablero, que `snapshot` crea de forma perezosa la primera vez |

**La receta de `git` es esta, entera**, porque es el único sistema del catálogo en la versión 1.0 y la
especificación no puede dejar a quien implemente adivinándola. Son cinco preguntas, de las que solo
importa la respuesta, y cuatro órdenes, que son las que actúan:

| Pregunta u orden | Qué se ejecuta, con el directorio del tablero como directorio de trabajo |
|---|---|
| ¿Hay un repositorio que lo contenga, y cuál es su raíz? | `git rev-parse --show-toplevel`, cuyo fallo significa que no hay ninguno |
| ¿Ese repositorio ignora la carpeta del tablero? | `git check-ignore` con la ruta del directorio detrás del separador `--`, como el resto de la receta, para que el nombre de la carpeta no pueda leerse como ninguna otra cosa |
| ¿El índice tiene algo preparado fuera de esos tres ficheros? | `git diff --cached --no-relative --name-only -z`, cuya salida se descarta y solo se cuenta |
| ¿Los tres ficheros difieren de lo que ya está guardado? | `git diff --cached --quiet` con los tres nombrados por su ruta, ya preparados, y su código de salida cero significa que no hay nada que guardar |
| ¿Cuál es el identificador de la revisión que ha resultado? | `git rev-parse HEAD`, ya guardada |
| Crear el repositorio propio | `git init` |
| Preparar los tres ficheros | `git add` con los tres nombrados por su ruta |
| Guardar la revisión | `git commit` con los tres ficheros nombrados por su ruta como argumentos de la propia orden, y con el mensaje de abajo |
| Publicar | `git push` en el repositorio donde haya ido la revisión |

**Nombrar los tres ficheros como argumentos de `git commit` es lo que acota el commit, y el `git add`
previo es lo que hace que exista el primero.** Con rutas explícitas en la línea de órdenes, `git commit`
activa por defecto su modo `--only`: toma el contenido actual del árbol de trabajo de esas rutas y las
comitea, sin tocar lo que hubiera ya preparado en el índice para cualquier otro fichero, y eso vale
igual con el `git add` delante, porque preparar los tres ficheros del tablero no prepara ninguno más.
El `git add` no se puede quitar: `git commit` con rutas solo acepta rutas que el repositorio ya conoce,
y responde `pathspec did not match any file(s) known to git` a un fichero sin seguir, que es
exactamente lo que son los tres la primera vez que `snapshot` corre sobre un repositorio recién
creado. Sin él, esa primera instantánea no guardaría nada y diría además que no había nada que
guardar.

**El caso de no haber nada que guardar se reconoce preguntando, no leyendo lo que `git commit`
escribe.** `git commit` termina distinto de cero tanto cuando no hay ningún cambio que registrar como
cuando la revisión falla de verdad, y distinguir los dos por su mensaje es justo lo que la receta de
un sistema no debe hacer. Por eso se pregunta antes, con `git diff --cached --quiet` sobre los tres
ficheros ya preparados: si termina en cero no había nada que guardar y el fallo de `git commit` es ese
y no otro, con código 0 y su nota; si termina distinto de cero sí lo había, y entonces un `git commit`
que falla es el error 8 de la tabla de abajo.

**Y antes de comitear se pregunta si el índice tiene algo más preparado**, con
`git diff --cached --no-relative --name-only -z`, cuyas dos banderas finales no son adorno y sin ellas
el recuento sale mal en una máquina corriente. `--no-relative` da las rutas relativas a la raíz del
repositorio sea cual sea el directorio de trabajo: sin ella, una máquina con `diff.relative` activado
listaría solo lo que cuelga del directorio del tablero, que es el directorio de trabajo de la
pregunta, y lo preparado en el resto del proyecto sería invisible y se contaría como cero. `-z`
separa las rutas con un byte cero y las escribe tal cual: sin ella, `core.quotePath`, que viene
activado de serie, devuelve entrecomillada y con escapes octales cualquier ruta con un carácter no
ASCII, así que los tres ficheros de un tablero que viva en una carpeta con un acento no se
reconocerían como suyos y se contarían como cambios ajenos.
Cuenta como ajeno cualquier ruta que aparezca en esa lista y no sea una
de las tres del tablero, midiendo sobre el índice entero del repositorio donde va a ir la revisión, no
solo sobre lo que hay dentro del propio directorio del tablero: cuando el tablero tiene su propio
repositorio, o cuando `snapshot` acaba de crearlo, la raíz del repositorio y el directorio del tablero son
el mismo sitio y las dos medidas coinciden; cuando el tablero vive dentro del repositorio del proyecto, la
raíz está por encima, y un cambio preparado en cualquier otro fichero del proyecto cuenta igual de ajeno
que uno preparado dentro del propio directorio del tablero. Si hay alguno, el `--only` de arriba ya
garantiza que `git commit` no lo toca, y `snapshot` además lo dice: `note: 2 staged change(s) outside the
board were left untouched`, por stderr y en `data.stagedOutsideBoard` del JSON.

El identificador que la salida devuelve es el del commit que resulta, completo en el JSON y abreviado en
el texto, donde son sus siete primeros caracteres. Con `custom`, que no devuelve ninguno, la línea se
queda sin él y dice `Committed to ...` sin más. `git commit` no escribe el identificador entero, así que se pregunta después con
`git rev-parse HEAD`, que es la quinta pregunta de la receta.

**Preguntar por la exclusión es lo que separa los dos casos de un tablero que vive dentro del
proyecto**, y no se puede deducir mirando el sistema de ficheros: las dos situaciones tienen los mismos
ficheros en los mismos sitios, y lo único que las diferencia es una línea en el fichero de exclusión del
proyecto. Sin esa pregunta, un tablero ignorado acabaría con su revisión pedida al repositorio del
código, que no puede añadir lo que ignora.

**El tablero dentro del proyecto y versionado con él es el único caso en el que la instantánea cruza a
otra máquina sin que nadie configure nada**, porque el repositorio del proyecto ya tiene remoto. En los
demás, el historial del tablero es estrictamente local hasta que alguien le añada uno a mano, y entonces
`--vcs push` lo publica. Publicar no es trabajo de `biso` mientras no se le pida: la sección ["Lo que se deja fuera a propósito"](../fuera-de-alcance.md) lo dice.

**Y hay que decir qué arrastra `--vcs push` cuando la revisión ha ido al repositorio del proyecto**:
publica esa rama entera, así que se lleva también los commits de código que estuvieran pendientes.
Es lo que el flag promete, y quien lo escribe ya está pidiendo publicar, así que `biso` no se niega
ni publica a medias; lo que hace es no ponerlo por defecto, que es el motivo de que el valor por defecto
sea `commit`.

**Con `custom` no hay ninguna de esas tres preguntas.** `biso` ejecuta la orden `commit` que declare la
configuración, en el directorio del tablero, y mira su código de salida; dónde acabe la revisión es cosa
de esa orden. Por eso una instantánea con `custom` no lleva identificador de revisión en su salida.

## Qué entra en la revisión

Tres ficheros, nombrados uno a uno: **`snapshot.ndjson`, `board.json` y el marcador `<id>.id`.** No se
añade el directorio entero, para que un fichero que alguien deje ahí a mano no se cuele en el historial.

El marcador entra porque es lo que hace que la identidad del tablero viaje con la instantánea: al leerla
de vuelta, `biso init --from` adopta ese `id` en vez de acuñar uno nuevo, y por eso el puntero
commiteado del proyecto sigue valiendo después de restaurar (["`biso init`"](init.md)). La base de datos no entra nunca, y
de eso se encarga el fichero de exclusión que `init` dejó escrito dentro del tablero, no este comando.

El mensaje de la revisión es `biso snapshot: 248 tasks`, con el recuento real de cada vez.

## Cómo se ejecutan las órdenes

Esto vale igual para `git` y para `custom`, porque no es de un sistema concreto sino de ejecutar un
programa que no es `biso`.

**Todas se ejecutan con el directorio del tablero como directorio de trabajo, y con la entrada estándar
cerrada y sin terminal.** Lo segundo evita el único cuelgue que de verdad ocurre: una orden esperando una
contraseña o una confirmación que nadie va a escribir. Con la entrada cerrada falla en vez de esperar, y
ese fallo se cuenta como cualquier otro.

**No hay tiempo máximo de espera, y es deliberado.** `biso` no puede interrumpir con seguridad una orden
que está a medio escribir en un repositorio ajeno, y un `push` legítimo contra un repositorio grande por
una red lenta tarda lo que tarda. Matar el proceso a mano no pierde nada, porque los dos ficheros de la
instantánea ya están escritos antes de que se ejecute la primera orden, que es la misma promesa que hace
la tabla de casos límite de aquí abajo.

**Lo que las órdenes escriben se reenvía por stderr**, línea a línea, cada una prefijada con el valor de
la clave `vcs` y dos puntos, o sea `git:` o `custom:`. Se reenvían las dos corrientes de la orden, la
estándar y la de error, porque las dos llevan cosas que quien llama querrá leer: `git commit` escribe su
resumen por la estándar y `git push` su progreso por la de error, así que quedarse con una sola pierde la
mitad. **El orden relativo entre las dos corrientes no está garantizado**; el de las líneas dentro de cada
una sí. Las líneas en blanco se descartan, y una última línea sin salto final cuenta como línea.

Nada de eso va nunca por stdout, que la sección ["stdout, stderr y qué va en cada uno"](../salida-y-terminal.md#stdout-stderr-y-qué-va-en-cada-uno) reserva para los datos, y `--quiet` no lo suprime,
por el motivo que da la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos).

**Y tiene una consecuencia visible que no es un fallo**: cuando no hay nada que guardar, la línea que
`git commit` escribe por su cuenta se reenvía igual, así que sale al lado de la nota que `biso` emite
para decir lo mismo. Se deja así a propósito, porque recortar la salida de un programa ajeno para que no
repita lo que `biso` ya dice obligaría a reconocer sus mensajes uno a uno, y eso es justo lo que la
receta de un sistema no debe hacer: la receta le pregunta cosas y mira códigos de salida, nunca lee lo
que escribe.

**La salida de las preguntas de la receta sí se descarta**, y es la única asimetría. Sus fallos o su
contenido son respuestas legítimas: `git rev-parse --show-toplevel` falla cuando no hay ningún
repositorio, `git check-ignore` termina distinto de cero cuando la carpeta no está ignorada,
`git diff --cached --no-relative --name-only -z` simplemente lista lo que haya, vacía cuando no hay nada ajeno preparado,
y las otras dos preguntas se hacen para que `biso` pueda decir en qué caso está, no para que nadie las
lea. Reenviar la primera pareja llenaría de un `fatal: not a git repository` alarmante el camino normal
de cualquier tablero que viva fuera de un repositorio, y la tercera no es un mensaje para quien llama
tal cual: `biso` la cuenta y la convierte en la nota de arriba. Se reenvía, por tanto, solo lo que
escriben las órdenes que actúan: `init`, `add`, `commit` y `push` en `git`, y `commit` y `publish` en
`custom`.

**Con `--json` no van por stderr.** Ahí stderr lleva el sobre de error (["Los errores en JSON"](../contrato-json.md#los-errores-en-json)) y no puede llevar además
texto suelto, así que esas líneas van dentro del sobre: en `data.vcsOutput` cuando la operación acaba
bien, y en `error.vcsOutput` cuando falla. Es una lista de cadenas, una por línea, y **ahí van sin el
prefijo**, que existe solo para separarlas a la vista en un terminal.

**Las dos claves se llaman igual y no llevan lo mismo.** `data.vcsOutput` lleva las líneas de todas
las órdenes que se ejecutaron, en el orden en que se ejecutaron, porque ahí no hay ninguna que
señalar. `error.vcsOutput` lleva **solo las de la orden que falló** (["Los errores en JSON"](../contrato-json.md#los-errores-en-json)): quien recibe un
`vcs_push_failed` está leyendo por qué falló la publicación, y las líneas del `commit` que sí
funcionó antes solo le harían buscar la causa donde no está.

**Una orden que no se puede ni lanzar no escribe nada, y aun así deja una línea.** El texto que
explica por qué no se pudo lanzar, que es el que nombra el programa que falta, viaja como una línea
más de esa salida, en las dos claves. Sin él, una orden `publish` de `custom` que apunta a un
programa inexistente daría un error 8 sin una sola pista de qué instalar.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Hay repositorio y hay cambios desde la última instantánea | Escribe los dos ficheros, guarda la revisión, código 0 |
| Hay repositorio y no hay ningún cambio desde la última instantánea | Escribe los dos ficheros (con el mismo contenido de antes) y no hay nada que guardar; `note: nothing to commit, snapshot.ndjson and board.json are unchanged since the last snapshot`, código 0 |
| Hay repositorio y el índice tenía algo más preparado, fuera de los tres ficheros del tablero | Guarda la revisión igual, con los tres ficheros nombrados por su ruta; lo demás sigue preparado, sin tocar; `note: N staged change(s) outside the board were left untouched`, código 0 |
| `--vcs none`, o la clave `vcs` en `none` | Escribe los dos ficheros y no ejecuta nada, código 0 |
| El directorio del tablero no está en ningún repositorio, y crearlo funciona | Lo crea, mete los tres ficheros en la primera revisión, código 0 |
| El directorio del tablero está dentro del repositorio del proyecto, que no lo ignora | Guarda la revisión ahí, con los tres ficheros nombrados por su ruta, código 0 |
| El sistema configurado no está instalado, o crear el repositorio falla | Escribe los dos ficheros, `note: no version control here, skipping the commit`, código 0. Con `custom`, "no está instalado" es que la orden `commit` no se pueda ni lanzar; una que sí se lanza y termina distinto de cero es el error de la fila de abajo |
| `--vcs push` y no había nada que guardar | Publica igual, y la nota de que no había nada que guardar sale como siempre. Lo que `--vcs push` promete no depende de que esta llamada haya guardado una revisión: puede haber revisiones anteriores sin publicar |
| La revisión falla por una razón de entorno, con un repositorio ya existente (sistema sin configurar, sin permiso, disco lleno, otra instantánea guardando a la vez) | Los dos ficheros ya han quedado escritos antes de intentarlo; Error 8, `code` `vcs_commit_failed`, y el mensaje dice que nada se ha perdido y que basta volver a llamar |
| `--vcs push` y la publicación falla | Los ficheros están escritos y la revisión guardada; Error 8, `code` `vcs_push_failed`. Con `custom`, una orden `publish` que no se puede ni lanzar cuenta también como publicación fallida y no como sistema sin instalar, porque la orden `commit` sí se lanzó: lo que falta no es el sistema, es esa orden |
| `--vcs push` con `vcs` igual a `custom` y sin orden `publish` declarada | Error 2, `code` `vcs_push_unavailable`, antes de escribir nada |
| `vcs` igual a `custom` y sin orden `commit` declarada, con `--vcs commit` o `--vcs push` | Error 2, `code` `vcs_commit_unavailable`, antes de escribir nada. `commit` es obligatoria (["Configuración de máquina"](../invocacion.md#configuración-de-máquina)), así que su ausencia es una configuración que no puede funcionar y no una revisión que se haya intentado: nunca es el error 8 de una revisión que falla. Con `--vcs none` no hay nada que declarar y no es error |
| Alguna tarea no se puede leer (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | Se salta, se cuenta, `warning: 1 task could not be read and was skipped`, y el código es 6 en vez de 0, igual que en `biso export` |
| No se puede escribir alguno de los dos ficheros | Error 8, con los dos ficheros anteriores intactos y sin intentar la revisión |
| No hay tablero | Error 20 |

## Salida

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
Committed a1b2c3d to the board's own repository
```

Cuando la revisión va al repositorio que contiene al tablero, la segunda línea lo nombra, porque es el
dato que distingue este caso y el que alguien querrá comprobar:

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
Committed a1b2c3d to /Users/avilches/Hub/Projects/My project, the repository this board lives in
```

Con `--vcs push`, una tercera línea dice que se ha publicado, y es esta:

```
Pushed that branch to its remote
```

Sin nada que guardar, la primera línea por stdout y la nota por stderr:

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
```

```
note: nothing to commit, snapshot.ndjson and board.json are unchanged since the last snapshot
```

Si el índice tenía algo más preparado fuera de esos tres ficheros, la nota sale igual junto a la primera
línea, se haya guardado la revisión o no:

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
Committed a1b2c3d to the board's own repository
```

```
note: 2 staged change(s) outside the board were left untouched
```

Con `--vcs none`, o si no hay sistema instalado, la salida por stdout es solo la primera línea; en el
segundo caso, además, la nota `note: no version control here, skipping the commit` por stderr.

Y por stderr sale también lo que hayan escrito las órdenes, prefijado, antes de cualquier nota y de la
línea de error si la hay:

```
git: [main a1b2c3d] biso snapshot: 248 tasks
git:  3 files changed, 12 insertions(+), 4 deletions(-)
```

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "snapshot",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "tasks": 248,
    "files": ["snapshot.ndjson", "board.json"],
    "vcs": "git",
    "committed": true,
    "commit": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
    "repository": "/Users/avilches/.biso/boards/my-project-3f9a2b1c",
    "pushed": false,
    "vcsOutput": ["[main a1b2c3d] biso snapshot: 248 tasks", " 3 files changed, 12 insertions(+), 4 deletions(-)"],
    "stagedOutsideBoard": 0,
    "skipped": []
  }
}
```

`vcs` es el sistema que se ha usado en esta llamada. Vale `none` exactamente cuando no se ha llegado a
intentar ninguna orden, que son los dos casos declarados: `--vcs none`, y la clave `vcs` de la máquina
en `none`. Con la clave en `git` o en `custom` vale eso, aunque el sistema resulte no estar instalado:
la alternativa sería que la misma nota `no version control here` saliera unas veces con `none` y otras
con `git`, según si lo que falló fue encontrar el programa o crear el repositorio, que es una
diferencia sobre la que nadie puede actuar.
`commit` y `repository` son `null` cuando `committed` es `false`, y `commit` también con `custom`, que
no devuelve identificador. `repository` es la raíz del repositorio donde ha ido la revisión, que es lo
que dice en qué caso de los tres se estaba. `pushed` es `false` salvo con `--vcs push` cumplido.
`stagedOutsideBoard` es el recuento de rutas que el índice tenía preparadas fuera de los tres ficheros
del tablero antes de comitear (["Dónde va la revisión, y cómo se decide"](#dónde-va-la-revisión-y-cómo-se-decide)): `0` salvo con `vcs` igual a `git` y algo
ajeno preparado, y siempre `0` con `none` o `custom`, que no hacen esa pregunta.
`skipped` lleva los identificadores de las tareas ilegibles que se han saltado, igual que en `biso ls`
(["`biso ls`"](ls.md)): vacío salvo cuando el código de salida es 6.

`files` son los dos ficheros que este comando escribe, **no los tres que entran en la revisión**: el
marcador `<id>.id` ya estaba ahí y lo escribió `biso init`. El `{files}` de la configuración de `custom`
(["Configuración de máquina"](../invocacion.md#configuración-de-máquina)) sí son los tres, y son dos conjuntos distintos con nombres parecidos.

`vcsOutput` son las líneas que escribieron todas las órdenes que se ejecutaron, en el orden en que se
ejecutaron y sin el prefijo que llevan en el modo de texto, y está vacía cuando no se ha ejecutado
ninguna. El `vcsOutput` de un sobre de error es otra cosa y lleva solo las líneas de la orden que
falló (["Cómo se ejecutan las órdenes"](#cómo-se-ejecutan-las-órdenes)).

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Escrito, y guardado si procedía | 0 |
| Sintaxis, o una orden de `custom` que hace falta y no está configurada | 2 |
| Alguna tarea se ha saltado por ilegible | 6 |
| No se pueden escribir los ficheros, o falla la revisión o la publicación | 8 |
| No hay tablero | 20 |

## `biso snapshot --help`

```
Usage: biso snapshot [options]

Write snapshot.ndjson and board.json into the board's own directory, then
record them with the version control system this machine is configured for
(the vcs key, git by default). Those two files and the <id>.id marker are
what `biso init --from` reads back to rebuild a board whole: its tasks, in
the same shape `biso export` writes, its configuration, in the same shape
`biso config list --json` prints, and its identity.

Where the revision lands follows the board's directory. If it is a
repository of its own, there. If it sits inside another repository that does
not ignore it, in that one, beside the code, which is what makes the
snapshot travel to other machines on its own. Otherwise this command creates
the board its own repository, lazily, the first time it runs there.

biso snapshot is the only command that ever runs another program, and the
only one that creates a repository. Invoking version control costs about
12ms, more than the 25ms startup budget for biso ls and biso prime allows on
the hot path.

Options:
      --vcs <mode>   none, commit or push (default commit)
  -h, --help         show this help

The two files are always written, whatever version control does. If it is
not installed, or creating the repository fails, the commit is skipped with
a note: it is optional, and its absence never fails this command. A commit
that fails once it is really attempted is an error, nothing is lost, and
running this command again after fixing the reason is all it takes.

Whatever those commands print is forwarded on stderr, prefixed with the
system name, and never suppressed. There is no timeout: interrupting a
half-written revision is not safe, and nothing is lost by killing this
command, since both files are on disk before the first one runs.

Exit codes:
  0  written, and recorded if that applied
  2  bad usage, or --vcs push with no publish command configured
  6  some task was skipped, unreadable
  8  cannot write there, or the commit or the push failed
  20 no board here

Examples:
  biso snapshot
  biso snapshot --vcs none
  biso snapshot --vcs push
```

---

