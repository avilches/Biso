# Cómo se elige el tablero

## Cómo nace un tablero

Un tablero nace con `biso init` (sección ["`biso init`"](cmd/init.md)), que se ejecuta una vez desde la
raíz del proyecto. Hace dos cosas: crea el tablero, que vive en un directorio propio, y escribe en el
proyecto un puntero, el fichero `.biso.json`, que dice a qué tablero pertenece ese proyecto. El puntero
se versiona con el proyecto; la base de datos no se versiona nunca. Esta página explica ese puntero y
cómo cualquier comando posterior vuelve a encontrar el tablero a partir de él; los flags de `init`, sus
mensajes y sus casos límite están en su propia página.

Dónde vive el tablero lo decide el flag `--at`, y hay tres formas habituales de usarlo.

### El tablero en la raíz por defecto de la máquina

Es lo que pasa con `biso init` a secas, y lo recomendable si no hay un motivo para otra cosa. El
proyecto solo recibe el puntero:

```
/Users/avilches/Hub/Projects/My project/
└── .biso.json        <- {"version": 1, "id": "3f9a2b1c"}: apunta al tablero 3f9a2b1c
```

Y el tablero se crea en la raíz por defecto, `~/.biso/boards` salvo que la
["Configuración de máquina"](invocacion.md#configuración-de-máquina) diga otra cosa:

```
/Users/avilches/.biso/boards/
└── my-project-3f9a2b1c/
    ├── board.db      <- la base de datos
    ├── 3f9a2b1c.id   <- el marcador con la identidad del tablero
    └── .gitignore    <- excluye board.db, para cuando se haga biso snapshot
```

El puntero no lleva ruta, porque cualquier comando encuentra el tablero buscando su marcador en las
raíces de la máquina. No hay nada que añadir al `.gitignore` del proyecto, porque dentro del proyecto
no hay nada del tablero salvo el puntero.

### El tablero fuera del proyecto, en una ruta propia

Con `--at` y una ruta absoluta, por ejemplo `biso init --at /Volumes/Datos/tableros/my-project`, el
tablero va donde se diga y el puntero guarda esa ruta:

```
/Users/avilches/Hub/Projects/My project/
└── .biso.json        <- {"version": 1, "id": "3f9a2b1c", "path": "/Volumes/Datos/tableros/my-project"}
```

```
/Volumes/Datos/tableros/my-project/
├── board.db
├── 3f9a2b1c.id
└── .gitignore
```

Sirve para un tablero que tenga que vivir en un disco concreto. La ruta absoluta se encuentra desde
cualquier copia de trabajo del proyecto, viva donde viva, mientras el tablero no se mueva.

### El tablero dentro del proyecto

Con `--at` y una ruta relativa que cae dentro del proyecto, por ejemplo `biso init --at .biso-board`,
el tablero vive junto al código y el puntero guarda la ruta relativa:

```
/Users/avilches/Hub/Projects/My project/
├── .biso.json        <- {"version": 1, "id": "3f9a2b1c", "path": ".biso-board"}
├── .gitignore        <- lo edita quien llama, si decide ignorar .biso-board/
└── .biso-board/
    ├── board.db
    ├── 3f9a2b1c.id
    └── .gitignore    <- lo escribe init: excluye board.db
```

Aquí quien llama tiene que decidir si el proyecto versiona la carpeta del tablero o la ignora. `init`
se lo recuerda con una nota, pero no toca el `.gitignore` del proyecto:

- **Si la ignora**, añadiendo `.biso-board/` al `.gitignore` del proyecto, el tablero lleva su propio
  historial: `biso snapshot` crea un repositorio dentro de la carpeta la primera vez.
- **Si la versiona**, las instantáneas de `biso snapshot` se commitean en el repositorio del proyecto,
  junto al código, y viajan con él a otras máquinas. La base de datos no entra nunca, porque la excluye
  el `.gitignore` que `init` dejó dentro del tablero.

En los dos casos, una copia de trabajo que viva dentro del proyecto, como un worktree de git en
`.claude/worktrees/`, encuentra el tablero subiendo por sus ancestros. Una copia que viva fuera del
proyecto no lo encuentra con la ruta relativa, y ese caso pide la ruta absoluta del apartado anterior
(["La ruta relativa o absoluta"](#la-ruta-relativa-o-absoluta)).

## Cómo se busca el tablero

### El orden de búsqueda

Un proyecto tiene un tablero, y el programa lo encuentra por este orden. Gana el primero que exista:

1. **El directorio de trabajo, cuando es el directorio de un tablero.** Se reconoce porque contiene
   el fichero `board.db`, y entonces el tablero es ese y no se busca nada más.
2. **El puntero del proyecto**, que es una marca que `biso init` deja en el proyecto y que dice qué
   tablero le corresponde. Se busca en el directorio de trabajo y en sus ancestros, con el tope que
   explica ["El tope de la búsqueda hacia arriba"](#el-tope-de-la-búsqueda-hacia-arriba).

**Las dos parten del directorio de trabajo**, que es el directorio actual salvo que el flag global
`-C` o la variable `BISO_CWD` digan otro (sección ["Flags globales"](cmd/flags-globales.md#flags-globales)). Por eso `-C` es lo único que hace falta para
trabajar contra otro tablero sin moverse: apuntando al directorio de un tablero se llega por la
primera vía, y apuntando a un proyecto cualquiera se llega por la segunda, al tablero que ese
proyecto tenga.

Esto es lo que hace la segunda vía cuando se llama desde un subdirectorio del proyecto: sube por los
ancestros hasta encontrar el puntero.

```
/Users/avilches/Hub/Projects/My project/
├── .biso.json
└── src/
    └── cli/          <- biso se ejecuta aquí
```

Desde `src/cli/` no hay ni `board.db` ni `.biso.json`, así que la búsqueda sube: `src/` tampoco lo
tiene, y `My project/` sí. Ahí encuentra el puntero y resuelve el tablero que nombra, sin que importe
desde qué subdirectorio se haya llamado.

**La raíz por defecto de la máquina** (sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina)) no es una vía más de esta lista: es el
directorio donde `biso init` sin `--at` crea los tableros nuevos. Volver a encontrar un tablero ya
creado depende siempre de una de las dos vías de arriba, nunca de mirar las carpetas de la raíz por
defecto a ver cuál le pega a este proyecto.

**La primera vía no adivina nada, y por eso no contradice el párrafo anterior.** Un directorio que
contiene `board.db` es ese tablero y no puede ser otro, y la configuración completa de un tablero vive
dentro de ese mismo fichero, con su nombre y su `task_prefix` incluidos (sección ["`biso config`"](cmd/config.md)), así que ahí no
falta ningún dato que el puntero tuviera que aportar: el puntero solo sirve para encontrar un tablero,
y quien ya está dentro de él no tiene nada que encontrar.

**Y de ahí sale el orden entre las dos**, que al quedar solo dos es toda la precedencia que existe:
estar dentro de un tablero es más específico que apuntar a uno desde un proyecto, así que quien
ejecuta un comando dentro de un tablero se refiere a ese, no al del proyecto que quizá lo contenga.

**Esta vía se conforma con la base de datos y no pide el marcador, y no es un descuido.** Buscar el
tablero de un `id` concreto sí exige las dos cosas, porque ahí la pregunta es cuál de varios directorios
es el que se busca y el marcador es lo único que la contesta. Aquí no hay nada que elegir: el directorio
ya está señalado con el dedo, y un tablero al que le falte el marcador tiene que poder abrirse
precisamente para que `biso doctor --fix` se lo devuelva (["`biso doctor`"](cmd/doctor.md#qué-comprueba)). Exigirlo también aquí dejaría sin
arreglo el único estado que ese arreglo existe para arreglar.

### El tope de la búsqueda hacia arriba

**El tope de la búsqueda hacia arriba es el directorio personal de quien llama**, el que dice la
variable `HOME`, cuando el directorio de trabajo está dentro de él: el recorrido comprueba ese
directorio y no sube más, sea cual sea su ruta.

```
/Users/avilches/                     <- HOME: aquí para la búsqueda
└── Hub/
    └── Projects/
        └── Biso/
            └── src/                 <- biso se ejecuta aquí
```

Desde `/Users/avilches/Hub/Projects/Biso/src` la búsqueda sube por `Biso/`, `Projects/` y `Hub/` y
para en `/Users/avilches`, sin llegar a `/Users` ni a la raíz. Si ninguno de esos directorios tiene el
puntero, el resultado es el mensaje de la subsección ["Cuando no hay tablero que encontrar"](#cuando-no-hay-tablero-que-encontrar). Con un directorio
personal en `/root` el recorrido pararía en `/root`.

**Cuando el directorio de trabajo cae fuera del directorio personal**, o cuando `HOME` no está
definido, el tope es el segundo componente de la ruta:

```
/Volumes/disco/                      <- tope de la búsqueda
└── proyecto/
    └── ...                          <- biso se ejecuta en algún punto de aquí
```

Desde `/Volumes/disco/proyecto` la búsqueda para en `/Volumes/disco`, y desde `/opt/proyecto/src`
pararía en `/opt/proyecto`. Es el mismo criterio de la regla de arriba aplicado donde no hay directorio
personal que lo exprese: no salir del área de trabajo de quien llama. El número no sale de contar
niveles por costumbre, sino de que el primer componente de una ruta absoluta es siempre un directorio
del sistema o el contenedor de los directorios personales de todo el mundo, y un puntero ahí no puede
estar a propósito, solo por accidente.

### La independencia del control de versiones

**Nada en esta búsqueda depende del control de versiones.** No se busca la raíz de ningún repositorio,
no se para al encontrar la marca de uno, y no se ejecuta ningún programa para preguntarlo. La razón es
que la sección ["`biso snapshot`"](cmd/snapshot.md) declara que el control de versiones es opcional y que un tablero funciona igual sin
él: una búsqueda que frenara donde empieza un repositorio pondría la respuesta a cuál es mi tablero en
manos de una herramienta ajena, y además una que la cambiaría sin avisar, porque es `biso snapshot` quien
crea el repositorio del tablero la primera vez que se ejecuta, de modo que el mismo comando en el mismo
sitio respondería una cosa antes y otra después. El precio de no tener ese freno es que un proyecto sin puntero propio hereda el del
proyecto que lo contenga, si lo hay, y se acepta a propósito: la sección ["La decisión de persistencia"](../decisiones/persistencia.md#la-decisión-de-persistencia) dice por
qué, y `biso where` enseña siempre de qué directorio salió el puntero que ha resuelto.

### Cuando no hay tablero que encontrar

Si nada de eso existe, cualquier comando salvo `init`, `where`, `help`, `--help` y `--version` aborta
antes de ejecutar su propia lógica, con código 20 y este mensaje por stderr:

```
error: no board here, and none configured for this project
hint: `biso init` creates one, `biso where` explains what was searched
```

Los cinco exentos no abortan así: `init`, `help`, `--help` y `--version` no necesitan tablero para
hacer su trabajo, y `biso where` lo necesita pero lo comprueba por su cuenta, con su propio mensaje y
su propio código 20 cuando no lo encuentra (sección ["`biso where`"](cmd/where.md)).

## El puntero del proyecto

### El fichero `.biso.json` y sus claves

**El puntero** es el fichero `.biso.json` que `biso init` escribe en el directorio desde el que se le
llama, que es la raíz del proyecto en el uso normal, y que se versiona con el proyecto. Nada obliga a
que esté en la raíz, porque la búsqueda de arriba lo encuentra en cualquier ancestro; ponerlo en la
raíz es lo que hace que lo vean todos los subdirectorios y todas las copias de trabajo. Tiene estas
claves:

| Clave | Tipo | Obligatoria | Notas |
|---|---|---|---|
| `version` | entero | sí | versión del formato del puntero |
| `id` | 8 caracteres hexadecimales en minúscula | sí | la identidad del tablero, y es inmutable. En mayúsculas el puntero es inválido: no se normaliza |
| `path` | ruta del directorio del tablero, absoluta o relativa | no | solo cuando el tablero no vive en una de las raíces de la sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina) |

Al estar versionado, todas las copias de trabajo del proyecto lo ven igual y comparten el mismo
tablero sin ningún paso adicional.

### Cómo se lee el puntero

Estas reglas gobiernan la lectura del puntero:

- Una clave desconocida en el puntero es un error.
- **`path` nombra el directorio del tablero, no el directorio que lo contiene.** Es la ruta que se
  abre, sin concatenarle nada.
- **Una `path` relativa se resuelve respecto al directorio que contiene el fichero puntero**, nunca
  respecto al directorio de trabajo. El motivo es el principio de que ningún comportamiento depende de
  dónde se ejecute el programa (sección ["Los principios"](principios.md)): con la búsqueda por ancestros explicada más arriba, en
  ["Cómo se busca el tablero"](#cómo-se-busca-el-tablero), el
  puntero se encuentra desde cualquier subdirectorio del proyecto, así que resolver contra el
  directorio de trabajo haría que el mismo puntero nombrara tableros distintos según desde qué
  subdirectorio se llamara. Una `path` que empiece por `~/` se expande al directorio personal de quien
  llama antes de resolverla.
- **El `id` manda y el `path` es solo una pista que puede no resolver**, y esta regla es la que lo hace
  cierto. Un directorio es el tablero `<id>` cuando contiene el fichero **`<id>.id`**, por ejemplo
  `3f9a2b1c.id`, **y también la base de datos**. Ese fichero es la identidad del tablero en el sistema
  de ficheros: el mismo `id` está guardado dentro de la base de datos, y el marcador lo repite en su
  nombre para que encontrar un tablero sea leer nombres de un directorio, sin abrir ninguna base de
  datos. Lleva dentro la versión del formato del almacén. **Si el nombre del marcador y el `id` de la
  base de datos discrepan es un error**, y `biso doctor` lo comprueba (["`biso doctor`"](cmd/doctor.md#qué-comprueba)).
- **Un directorio con el marcador pero sin la base de datos no es un tablero, y la búsqueda sigue.** No
  es un caso rebuscado: es lo que recibe una copia de trabajo de un proyecto que versionó el directorio
  de su tablero, porque el fichero de exclusión que `init` escribe ahí dentro excluye siempre la base de datos y
  nunca el marcador (["`biso init`"](cmd/init.md)). Saltarlo y seguir buscando es lo que encuentra el tablero de verdad un nivel
  más arriba. Si la búsqueda acaba sin nada, el error es el de la subsección
  ["El puntero nombra un tablero que no está en esta máquina"](#el-puntero-nombra-un-tablero-que-no-está-en-esta-máquina), y nombra ese directorio, porque
  explica el fallo mejor que decir solo que no se encontró el tablero.
- **El nombre de la carpeta del tablero es decorativo, y nadie resuelve nunca por él.** Al crearla,
  `biso init` la llama `<slug>-<id>`, por ejemplo `my-project-3f9a2b1c`, porque eso hace legible un listado de
  la raíz por defecto. Pero renombrarla no rompe nada, ni la renombra `biso` cuando cambia el nombre
  del tablero (["`biso config`"](cmd/config.md)), ni hay comprobación alguna sobre ella.
- **Cómo se busca cuando el `path` no resuelve, o cuando no hay `path`**: se recorren la raíz por
  defecto y después las raíces adicionales (sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina)), en ese orden, mirando en cada carpeta si
  contiene el marcador `<id>.id`. Y una `path` relativa que no resuelve en el directorio del puntero se
  prueba, tal cual, contra cada uno de sus ancestros hasta el mismo tope explicado antes, en
  ["Cómo se busca el tablero"](#cómo-se-busca-el-tablero). Ese
  último paso es el que hace que un tablero que vive dentro del proyecto se siga encontrando desde una
  copia de trabajo que no lo tiene, como un worktree de git, donde el puntero está versionado y el
  directorio del tablero no.

### Cómo se deriva el nombre de la carpeta

**El `<slug>` del nombre de la carpeta se deriva de `project_name`, nunca es el nombre literal**:
`project_name` es texto libre, y copiarlo tal cual metería espacios y mayúsculas en una ruta que
alguien va a teclear. La derivación reutiliza el mismo paso de `normalizar(x)` (sección ["El algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia)) que ya usa
la derivación de `task_prefix` (sección ["Identificador de tarea"](modelo-de-datos/identificadores.md#identificador-de-tarea)): pasar el nombre a minúsculas según Unicode y quitarle
los diacríticos. Después, cada tirada de caracteres que no sean una letra ASCII ni un dígito ASCII se
colapsa en un solo guion, recortando los que queden en los extremos. Así, `Kex` da `kex`, `Mi Proyecto`
da `mi-proyecto`, y `Peña 2026` da `pena-2026`. Ningún carácter de `project_name`, incluido un separador
de ruta, puede romper el nombre de la carpeta, porque cualquier tirada de ellos se colapsa igual en un
guion.

Eso resuelve tres cosas de golpe: que dos proyectos de la misma máquina se puedan llamar igual sin
chocar, porque lo que identifica al tablero no es su nombre sino su `id`; que el puntero siga
resolviendo aunque alguien renombre la carpeta o mueva el tablero a otra raíz, porque el marcador viaja
con el tablero y no con su nombre; y que cambiar el nombre de un tablero no toque el sistema de ficheros
en absoluto (["`biso config`"](cmd/config.md)).

**El slug y el `task_prefix` arrancan del mismo `project_name` pero fallan por motivos distintos, y
hay que comprobar los dos.** `"2026"` da un slug válido, `2026`, pero no da ningún prefijo, porque no
le queda ninguna letra ASCII (sección ["Identificador de tarea"](modelo-de-datos/identificadores.md#identificador-de-tarea)); `"///"` da un slug vacío y también un prefijo vacío. Que
una de las dos derivaciones salga bien no dice nada de la otra, así que ninguna de las dos
comprobaciones sustituye a la otra.

### Cómo `biso init` genera el id y escribe el puntero

`biso init` genera el `id` de la fuente de números aleatorios del sistema, comprobando que no exista ya
en ninguna de las raíces de la sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina), que es una lectura de directorio por raíz, y escribe el
puntero siempre que no exista ya uno: sin `--at`, con el tablero en la raíz por defecto y **sin clave
`path`**, porque ahí lo encuentra la búsqueda por marcador; con `--at`, con el tablero donde se le diga
y esa ruta en `path` (sección ["`biso init`"](cmd/init.md)). Cuando ya existe un puntero sin tablero aquí, la excepción es la de
la subsección ["El puntero nombra un tablero que no está en esta máquina"](#el-puntero-nombra-un-tablero-que-no-está-en-esta-máquina): adopta el `id` que ya lleva en vez de generar uno nuevo, y no reescribe el puntero. `biso
where` dice cuál se ha usado y por qué (sección ["`biso where`"](cmd/where.md)). No hay ningún caso en el que haya que escribirlo
a mano.

**Hay una segunda excepción, hermana de la anterior pero al revés.** Cuando este proyecto no resuelve ya,
por su cuenta, a ningún tablero accesible (no hay puntero aquí, o el que hay no resuelve a nada en esta
máquina), y `--at` nombra un directorio que ya es un tablero íntegro, con `board.db` legible, `init`
tampoco acuña un `id` nuevo: adopta el que ya lleva el marcador de ese directorio. Pero aquí sí escribe,
o reescribe, el puntero, porque esta vez no había ninguno correcto que conservar: es la única forma de
que este proyecto vuelva a encontrar ese tablero por su cuenta. No toca ni la configuración ni las
tareas del tablero destino. Esta excepción no se aplica cuando el proyecto ya resuelve, por su cuenta, a
un tablero accesible, aunque sea distinto del destino: ahí `--at` a otro tablero es el Error 2 de "ya
hay un tablero accesible desde aquí" (sección ["`biso init`"](cmd/init.md)), visto desde el lado del destino. Son los remedios de
["El puntero se pierde"](#el-puntero-se-pierde) y de ["La ruta relativa o absoluta"](#la-ruta-relativa-o-absoluta), más abajo.

### La ruta relativa o absoluta

**La forma de la `path` la elige quien llama, con la forma que le da a `--at`** (sección ["`biso init`"](cmd/init.md)), y las dos
sobreviven a cosas distintas. La relativa sobrevive a que el proyecto entero se mueva de sitio con su
tablero dentro. La absoluta sobrevive a que se trabaje desde una copia del proyecto que vive fuera de
él. Ninguna de las dos sobrevive a un clon en otra máquina, y conviene no atribuírselo a la relativa,
porque ahí el directorio del tablero no viaja de ninguna forma.

**El caso que decide entre las dos es dónde viven las copias de trabajo del proyecto**, y es el motivo de
que la búsqueda por ancestros exista. Un worktree de git recibe el puntero, porque está versionado, y no
recibe el directorio del tablero, porque está ignorado. Cuando el worktree vive dentro del proyecto, que
es la convención de esta máquina, la búsqueda por ancestros encuentra el tablero un par de niveles más
arriba y la ruta relativa funciona. Cuando el worktree vive fuera del proyecto, no hay ancestro común que
lo contenga y la relativa no puede resolver: ahí hace falta la absoluta, y por eso la elección no puede
ser del programa.

**El único caso que sigue pidiendo una corrección a mano** es mover o renombrar a mano el directorio de
un tablero que vive fuera de las raíces de la sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina), porque entonces su `path` deja de resolver y
no hay ninguna raíz que recorrer para encontrarlo. Se arregla con `biso init --at <ruta nueva>`,
apuntando al directorio movido. Como la `path` vieja ya no resuelve a nada, este proyecto no resuelve ya
a ningún tablero accesible por su cuenta, así que tampoco aquí `init` da el Error 2 de "ya hay un
tablero": adopta el `id` que el marcador de ese directorio ya lleva y reescribe el puntero con la `path`
nueva (secciones ["`biso init`"](cmd/init.md) y ["Cómo `biso init` genera el id y escribe el puntero"](#cómo-biso-init-genera-el-id-y-escribe-el-puntero), más arriba).

## Casos especiales y errores

### El mismo id en dos sitios

**Si el mismo `id` aparece en dos sitios, es un error** que nombra los dos directorios y no elige
ninguno, porque elegir sería escribir en un tablero que quien llama no ha nombrado, y dos almacenes con
la misma identidad es exactamente lo que esta persistencia no admite.

```
/Users/avilches/.biso/boards/
└── my-project-3f9a2b1c/
    ├── board.db
    └── 3f9a2b1c.id

/Volumes/work/boards/
└── my-project-3f9a2b1c/
    ├── board.db
    └── 3f9a2b1c.id
```

Sale con **código 22** (`AMBIGUOUS_BOARD`, sección ["Códigos de salida"](codigos-de-salida.md)), con la clave `code` `ambiguous_board_id`, y con este mensaje por stderr:

```
error: board 3f9a2b1c is in two places, and biso will not choose between them
        /Users/avilches/.biso/boards/my-project-3f9a2b1c
        /Volumes/work/boards/my-project-3f9a2b1c
hint: rename or remove one of the two directories
```

Tiene su propio código de salida y no el 20 porque su remedio no se parece a los otros dos: aquí no
falta un tablero que `biso init` pueda crear, sobra uno que solo una persona puede decidir cuál es. El
error lo da cualquier comando que tenga que resolver el tablero, y también `biso init`, que recorre las
mismas raíces para comprobar que el `id` que va a acuñar o a adoptar no exista ya (["`biso init`"](cmd/init.md)). **`biso
doctor` no lo comprueba**, y no es un olvido: para llegar a ejecutarse, `doctor` necesita un tablero
resuelto, así que en un tablero duplicado aborta con este mismo error antes de comprobar nada, igual
que hace con la base de datos ilegible de la sección ["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).

### El puntero se pierde

**Si el puntero se pierde** (se borra a mano, o el proyecto se clona sin haberlo commiteado antes), la
recuperación es explícita, nunca automática, y consiste en una sola cosa: `biso init --at <ruta>` con
la ruta del directorio del tablero que ya existe. Como ese directorio ya es un tablero íntegro y este
proyecto, sin puntero, no resuelve ya a ningún otro tablero accesible, `init` no lo rechaza con el
Error 2 de "ya hay un tablero": adopta el `id` que el marcador de ese directorio ya lleva y escribe un
puntero nuevo con ese `id` y con ese `path`, relativa o absoluta según la regla de la subsección
["La ruta relativa o absoluta"](#la-ruta-relativa-o-absoluta) (secciones ["`biso init`"](cmd/init.md) y ["Cómo `biso init` genera el id y escribe el puntero"](#cómo-biso-init-genera-el-id-y-escribe-el-puntero), más arriba). Es lo único que arregla el proyecto,
porque es lo único que deja el puntero otra vez donde lo ven todos sus subdirectorios y todas sus
copias de trabajo.

**Trabajar con `-C <directorio del tablero>` no es la otra mitad de esa recuperación, sino lo que se
puede hacer mientras tanto**: llega al tablero por la primera vía, pero solo en la llamada en la que
se escribe, y no deja nada apuntado, así que el comando siguiente vuelve a no encontrar nada. Sin
escribir el puntero, el proyecto no vuelve a encontrar su tablero por su cuenta.

### El puntero nombra un tablero que no está en esta máquina

**Si el puntero existe pero nombra un tablero que esta máquina no tiene** (lo típico al clonar el
proyecto en otro ordenador), el mensaje es otro, porque aquí sí hay un puntero. Dice que hay uno, qué
identificador nombra, y cómo recuperarlo, y ese "cómo" depende de si la búsqueda, al subir por los
ancestros probando la `path` relativa (["Cómo se lee el puntero"](#cómo-se-lee-el-puntero)), se cruzó por el camino con un directorio a
medias: el que deja un proyecto que versiona la carpeta de su tablero, con el marcador pero sin
`board.db` (["El tablero dentro del proyecto"](#el-tablero-dentro-del-proyecto)). Sin ese directorio no hay ninguna pista de dónde recuperar el
tablero; con él, sí.

**Sin ningún directorio a medias de por medio** (el puntero no lleva `path`, porque el tablero vivía en
la raíz por defecto de la máquina, que no viaja con el proyecto), el mensaje es el genérico, y el
remedio es crear el tablero adoptando la identidad que el puntero ya nombra:

```
error: this project's pointer names board 3f9a2b1c, which is not on this machine
hint: `biso init` creates it here, adopting id 3f9a2b1c
```

**Con un directorio a medias** (la `path` del puntero resuelve, o resolvió contra alguno de sus
ancestros, en un directorio que lleva el marcador `3f9a2b1c.id` pero no `board.db`), el mensaje nombra
ese directorio y distingue si hay algo que restaurar de él, porque un `biso init` a secas lo ignoraría
por completo y crearía un tablero vacío en la raíz por defecto, dejando huérfanas las tareas que
pudiera haber en ese directorio.

Si el directorio también lleva `snapshot.ndjson` y `board.json`, es decir, si `biso snapshot` llegó a
correr y a comitearse al menos una vez, hay tareas que recuperar:

```
error: this project's pointer names board 3f9a2b1c, which is not on this machine
hint: /Users/avilches/Hub/Projects/My project/.biso-board has this board's marker
      and snapshot, versioned without its database
hint: `biso init --from "/Users/avilches/Hub/Projects/My project/.biso-board"
      --at "/Users/avilches/Hub/Projects/My project/.biso-board"` restores it
      there, adopting id 3f9a2b1c
```

Si el directorio solo lleva el marcador, porque la carpeta se versionó nada más crearla y antes de la
primera instantánea, no hay tareas que recuperar y el remedio es crear el tablero vacío ahí mismo, con
el mismo `id`:

```
error: this project's pointer names board 3f9a2b1c, which is not on this machine
hint: /Users/avilches/Hub/Projects/My project/.biso-board has this board's marker,
      versioned before any snapshot was ever taken
hint: `biso init --at "/Users/avilches/Hub/Projects/My project/.biso-board"`
      creates it there, adopting id 3f9a2b1c
```

**En los tres casos el código de salida sigue siendo 20**, porque para quien llama la situación es la
misma: no hay tablero con el que trabajar. Lo que cambia es qué `hint` explica mejor el remedio, y la
clave `code` del sobre JSON (sección ["El contrato JSON"](contrato-json.md)) sigue siendo `pointer_unresolved` en los tres, no `no_board`. Ni `--from`
ni `--at` a secas dan el Error 2 de "ya hay un tablero" sobre ese directorio a medias, porque un
directorio sin una base de datos legible no cuenta como tablero accesible (sección ["`biso init`"](cmd/init.md)).

### Un tablero huérfano

**Un tablero cuyo proyecto ya no existe queda huérfano** en la raíz por defecto, y ningún comando de
hoy lo ve.

### Compartir un tablero entre proyectos

**Dos proyectos distintos pueden apuntar legalmente al mismo tablero.** No hay forma de distinguir "dos
copias de trabajo del mismo proyecto" de "dos proyectos que comparten tablero", porque el mecanismo es
el mismo puntero, y compartir es precisamente para lo que existe.

---
