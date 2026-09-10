# `biso`: especificación del CLI de gestión de tareas

Este documento define `biso`, una herramienta de línea de comandos para llevar las tareas de un
proyecto. Está escrito para que alguien implemente el programa entero a partir de él sin preguntar
nada: cada comando trae su firma, su tabla de parámetros, su comportamiento en los casos límite, la
salida literal que imprime, su esquema JSON, sus códigos de salida y el texto exacto de su ayuda.

El destinatario principal de `biso` es un agente automático que trabaja dentro del proyecto. La
salida es predecible, los errores son distinguibles por su código sin leer el mensaje, y ningún
comportamiento depende de dónde se ejecute el programa.

**Qué define este documento y qué no.** Define la interfaz del programa y el modelo de datos lógico de
una tarea, no el porqué de cómo se guardan los datos: esa decisión, con su razonamiento y su evidencia,
vive en `docs/DECISIONES.md` (sección 12), y este documento no la repite. Aquí aparece el mecanismo
solo donde afecta al comportamiento observable, como el fichero puntero `.biso.json` (sección 3.2) o la
base de datos SQLite que `biso doctor` comprueba (sección 10.11); donde no lo afecta, este documento
enuncia el requisito (por ejemplo, que dos procesos simultáneos no puedan asignar el mismo
identificador) y deja el resto del mecanismo fuera de aquí.

El modelo de tarea es compatible con el de Backlog.md, de modo que se puede importar y exportar entre
las dos herramientas sin perder campos. **La compatibilidad es de modelo de datos y no de formato de
fichero**: los campos se corresponden uno a uno, pero `biso` no lee ni escribe los ficheros Markdown de
esa herramienta, y no hay ninguna intención de que lo haga (sección 14).

**Convención de idioma.** La prosa de este documento va en español. Todo lo que es interfaz del
programa (nombres de comando, banderas, textos de ayuda, mensajes de error, claves JSON y claves de
configuración) va en inglés, porque es lo que la persona o el agente que usa el programa lee y
escribe.

## Vocabulario de este documento

| Español | Qué es | En la interfaz |
|---|---|---|
| **tablero** | Las tareas de un proyecto con su configuración y sus vocabularios | `board` |
| **tarea** | La unidad de trabajo | `task` |
| **estado** | Uno de los valores configurados en `statuses` | `status` |
| **papel** | La función que cumple un estado. Son tres, obligatorios y distintos | las claves `*_status` |
| **inicial** | El papel del estado donde nace una tarea | `initial_status` |
| **activo**, activa | El papel del estado que escribe el agente al coger la tarea | `active_status` |
| **terminal**, terminada | El papel del estado final | `terminal_status` |
| **asignación** | Quién debe hacer la tarea. Es el gesto con el que una persona encarga trabajo | `assignees` |
| **arrendamiento** | Hasta cuándo vale la reserva de un agente sobre una tarea activa. No es un estado | `leaseExpiresAt`, `leaseHolder`, `leaseExpired` |
| **pregunta abierta**, aparcada | Lo que detiene una tarea a la espera de una persona. No es un estado | `question`, `waiting` |
| **archivada** | Fuera del tablero activo sin perder nada. No es un estado | `archived` |
| **bloqueada** | Depende de alguna tarea sin terminar. Solo dependencias, nunca personas | `blocked` |
| **persona** | Quien encarga y quien responde | |
| **agente** | El programa automático que coge tareas y las hace | |
| **criterio** | Un elemento de las dos listas de comprobación | `acceptanceCriteria`, `definitionOfDone` |
| **comentario** | Una entrada inmutable del histórico | `comments` |

Cinco palabras quedan restringidas, en cuatro reglas, y conviene decir a qué en vez de prohibirlas a
secas, porque tres de ellas tienen un uso legítimo:

- **columna** nombra únicamente una columna de la tabla que imprimen `biso ls` y `biso prime`, que
  tiene ocho. Un estado del tablero no se llama nunca columna.
- **tarjeta** y **ticket** nombran únicamente lo que otra herramienta tiene, como una tarjeta de
  Trello. Lo de `biso` es una tarea.
- **panel** no se usa nunca: el conjunto de tareas es el tablero, y lo que abre `biso board` es la
  interfaz web.
- **bloqueada** no se usa nunca referida a una persona. Eso es una pregunta abierta.

---

## 1. Los principios

Siete reglas. El resto del documento es una consecuencia de ellas.

1. **Un valor que el tablero no conoce es un error, se esté leyendo o escribiendo, y siempre con la
   misma regla de coincidencia.** Un filtro con un valor imposible nunca devuelve una lista vacía.
   Como consecuencia, una lista vacía es un hecho sobre el tablero y quien la recibe puede actuar en
   consecuencia.

2. **Un nombre significa siempre lo mismo, en todos los comandos.** No existen dos banderas con el
   mismo nombre y semántica distinta según dónde se usen, ni dos nombres para el mismo concepto.

3. **El nombre desnudo añade. Sustituir se dice en voz alta.** `--label` añade una etiqueta y
   `--set-label` reemplaza la lista entera.

4. **La salida por defecto de una escritura es lo que quien llama no sabía.** Nunca el eco de lo que
   acaba de escribir.

5. **Un gesto del flujo de trabajo es un comando.** Empezar una tarea y terminarla tienen nombre
   propio y cuestan una llamada cada uno.

6. **Todo lo que se hace una vez se puede hacer cien veces, y se valida antes de escribir nada.**

7. **La forma de la salida no depende de si hay un terminal detrás.** Solo el color mira el terminal.
   Los datos, nunca.

---

## 2. Códigos de salida

Tabla global. Ningún comando usa un código fuera de esta tabla, y ningún código tiene dos
significados. Quien llama puede ramificar sobre el número sin leer el mensaje.

| Código | Nombre | Significado | Ejemplo |
|---:|---|---|---|
| 0 | `OK` | La operación terminó y se aplicó | `biso new "Algo"` |
| 1 | `INTERNAL` | Fallo no previsto del programa | una excepción no capturada |
| 2 | `USAGE` | La línea de comandos está mal formada | bandera desconocida, falta un obligatorio, dos banderas incompatibles, identificador mal formado, bandera de escritura en un comando de lectura |
| 3 | `BAD_VALUE` | El valor que llega es sintácticamente correcto pero el tablero no lo reconoce, o un dato guardado no se puede interpretar | `--status "Pending"` en un tablero cuyos estados son otros |
| 4 | `NOT_FOUND` | La entidad referida no existe | `biso get TASK-999` |
| 5 | `AMBIGUOUS` | La referencia encaja con más de una entidad | `biso get "parser"` con tres coincidencias |
| 6 | `PRECONDITION` | La operación es válida, pero el estado actual del tablero no la permite o no la satisface | `biso finish --strict` con criterios sin marcar, o `biso doctor` con problemas pendientes |
| 7 | `ENVIRONMENT` | Falla el entorno, no la petición | el almacén no responde, no hay permisos, no se puede adquirir el acceso exclusivo, no hay terminal donde hace falta |
| 8 | `NO_BOARD` | No hay tablero accesible desde donde se ha llamado | cualquier comando fuera de un tablero, salvo `init`, `help`, `--help` y `--version`, que no necesitan uno; `biso where` también devuelve 8 cuando no encuentra ninguno |
| 9 | `VALIDATION` | Una validación previa ha fallado y **no se ha escrito nada** | `biso new --from tareas.ndjson` con la línea 47 inválida |
| 10 | `DAMAGED` | El tablero está donde tiene que estar, y su almacén no se puede leer | `board.db` que no abre, o que falla su comprobación de integridad (4.12) |
| 11 | `AMBIGUOUS_BOARD` | El mismo identificador de tablero aparece en dos sitios, y elegir uno sería escribir en el que nadie ha nombrado | dos raíces de la sección 3.3 con una carpeta que lleva el mismo marcador (3.2) |

Cinco reglas que acompañan a la tabla:

- **El código 9 garantiza que no se ha escrito nada.** Si un comando termina con 9, el tablero está
  exactamente como estaba antes. Por eso una validación fallida dentro de un lote se reporta como 9 y
  no como 3 ni como 4, y por eso el 9 llega siempre con el detalle de **todos** los fallos
  encontrados, no solo del primero.
- **Un listado vacío es siempre 0.** Un tablero donde de verdad no hay nada que cumpla un filtro
  válido no es un error.
- **El código 3 cubre dos direcciones.** Un valor de entrada que el tablero no reconoce, y un dato ya
  guardado que el programa no sabe interpretar. Las dos son "el vocabulario no cuadra", y el mensaje
  siempre dice cuál de las dos ha ocurrido.
- **El código 1 es un fallo del programa, no de quien llama.** La reacción correcta es informar, no
  reintentar con otros parámetros.
- **Los códigos 8, 10 y 11 son los tres desenlaces malos de resolver el tablero, y son tres porque el
  remedio de cada uno es otro.** Con el 8 no hay tablero y `biso init` lo crea; con el 10 el tablero
  está ahí y hay que reconstruirlo desde una instantánea con `biso init --from`, que es lo único que lo
  arregla; con el 11 hay dos y hay que quitar o renombrar uno de los dos directorios a mano. Por eso el
  daño del almacén no comparte número con la ausencia de tablero, aunque para quien llama las tres
  frases empiecen igual: quien ramifica sobre el número tiene que poder elegir el remedio sin leer el
  mensaje, que es el principio de la sección 1. **Ni el 10 ni el 11 aparecen en la tabla de códigos de
  salida de cada comando**, porque no son desenlaces propios de ninguno sino del tablero entero, igual
  que el 1. La excepción es `biso where`, que existe justamente para explicar la resolución y los lleva
  los dos en su tabla (10.2).

---

## 3. Banderas globales

Valen para todos los comandos, se pueden escribir antes o después del nombre del comando, y ningún
comando puede redefinir ninguna de ellas ni cambiar su significado.

| Bandera | Corta | Tipo | Por defecto | Qué hace |
|---|---|---|---|---|
| `--cwd <path>` | `-C` | ruta | el directorio actual | Resuelve el tablero desde ahí, sin cambiar el directorio del proceso |
| `--json` | | booleano | falso | Toda la salida de datos es JSON, en el sobre de la sección 12 |
| `--quiet` | `-q` | booleano | falso | Reduce la salida a lo mínimo. Ver más abajo |
| `--print` | | booleano | falso | Después de escribir, imprime la ficha completa de cada tarea afectada |
| `--color <when>` | | `auto`, `always`, `never` | `auto` | Control de los códigos de color |
| `--dry-run` | | booleano | falso | Valida todo, no escribe nada. Sale 0 si habría funcionado y 9 si no |
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
  `--fix` es un comando de escritura y la bandera se comporta como en cualquier otro (10.11).
  `snapshot` (10.14) entra en la lista por el mismo motivo que `export`, que escriben ficheros y no
  tocan ninguna tarea.
- **`--dry-run` sí vale en `biso init` y en `biso config set`**, que escriben sin tocar ninguna tarea
  existente y tienen los dos algo que validar antes: `init --from` valida la instantánea entera contra
  el vocabulario que ella misma trae, y sale 0 si habría funcionado y 9 si no, que es exactamente lo
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
  líneas informativas de stderr que empiezan por `note:`. **Nunca suprime un `warning:` ni un
  `error:`.** Silenciar un aviso es cosa de quien llama, con `2>/dev/null`. **En un comando de
  lectura no hay identificadores afectados que imprimir**, así que ahí `--quiet` no cambia stdout: solo
  suprime las líneas `note:` de stderr, igual que en un comando de escritura.

### 3.1. Variables de entorno

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
| Autor de la pregunta, en `biso ask` | Error 2: `error: biso ask needs an identity; set it with biso config set me <you> or BISO_ME` (10.7.5) |
| Autor de la respuesta, en `biso answer` | Error 2: `error: biso answer needs an identity; set it with biso config set me <you> or BISO_ME` (10.7.6) |
| La línea `you are` de `biso prime` | `you are     (not set)`, con una nota que remite a `biso config set me` |

### 3.2. Cómo se elige el tablero

Un proyecto tiene un tablero, y el programa lo encuentra por este orden. Gana el primero que exista:

1. **El directorio de trabajo, cuando es el directorio de un tablero.** Se reconoce porque contiene
   el fichero `board.db`, y entonces el tablero es ese y no se busca nada más.
2. **El puntero del proyecto**, que es una marca que `biso init` deja en el proyecto y que dice qué
   tablero le corresponde. Se busca en el directorio de trabajo y en sus ancestros, con el tope de la
   regla que cierra esta lista.

**Las dos parten del directorio de trabajo**, que es el directorio actual salvo que la bandera global
`-C` o la variable `BISO_CWD` digan otro (sección 3). Por eso `-C` es lo único que hace falta para
trabajar contra otro tablero sin moverse: apuntando al directorio de un tablero se llega por la
primera vía, y apuntando a un proyecto cualquiera se llega por la segunda, al tablero que ese
proyecto tenga.

**La raíz por defecto de la máquina** (sección 3.3) no es una vía más de esta lista: es el
directorio donde `biso init` sin `--at` crea los tableros nuevos. Volver a encontrar un tablero ya
creado depende siempre de una de las dos vías de arriba, nunca de mirar las carpetas de la raíz por
defecto a ver cuál le pega a este proyecto.

**La primera vía no adivina nada, y por eso no contradice el párrafo anterior.** Un directorio que
contiene `board.db` es ese tablero y no puede ser otro, y la configuración completa de un tablero vive
dentro de ese mismo fichero, con su nombre y su `task_prefix` incluidos (sección 10.10), así que ahí no
falta ningún dato que el puntero tuviera que aportar: el puntero solo sirve para encontrar un tablero,
y quien ya está dentro de él no tiene nada que encontrar.

**Y de ahí sale el orden entre las dos**, que al quedar solo dos es toda la precedencia que existe:
estar dentro de un tablero es más específico que apuntar a uno desde un proyecto, así que quien
ejecuta un comando dentro de un tablero se refiere a ese, no al del proyecto que quizá lo contenga.

**Esta vía se conforma con la base de datos y no pide el marcador, y no es un descuido.** Buscar el
tablero de un `id` concreto sí exige las dos cosas, porque ahí la pregunta es cuál de varios directorios
es el que se busca y el marcador es lo único que la contesta. Aquí no hay nada que elegir: el directorio
ya está señalado con el dedo, y un tablero al que le falte el marcador tiene que poder abrirse
precisamente para que `biso doctor --fix` se lo devuelva (10.13). Exigirlo también aquí dejaría sin
arreglo el único estado que ese arreglo existe para arreglar.

**El tope de la búsqueda hacia arriba es el directorio personal de quien llama**, el que dice la
variable `HOME`, cuando el directorio de trabajo está dentro de él: el recorrido comprueba ese
directorio y no sube más, sea cual sea su ruta. Desde `/Users/avilches/Hub/Projects/Biso/src` sube
hasta `/Users/avilches` y para ahí, sin llegar a `/Users` ni a la raíz, y con un directorio personal
en `/root` pararía en `/root`.

**Cuando el directorio de trabajo cae fuera del directorio personal**, o cuando `HOME` no está
definido, el tope es el segundo componente de la ruta, de modo que desde `/Volumes/disco/proyecto` para
en `/Volumes/disco` y desde `/opt/proyecto/src` para en `/opt/proyecto`. Es el mismo criterio de la
regla de arriba aplicado donde no hay directorio personal que lo exprese: no salir del área de trabajo
de quien llama. El número no sale de contar niveles por costumbre, sino de que el primer componente de
una ruta absoluta es siempre un directorio del sistema o el contenedor de los directorios personales de
todo el mundo, y un puntero ahí no puede estar a propósito, solo por accidente.

**Nada en esta búsqueda depende del control de versiones.** No se busca la raíz de ningún repositorio,
no se para al encontrar la marca de uno, y no se ejecuta ningún programa para preguntarlo. La razón es
que la sección 10.14 declara que el control de versiones es opcional y que un tablero funciona igual sin
él: una búsqueda que frenara donde empieza un repositorio pondría la respuesta a cuál es mi tablero en
manos de una herramienta ajena, y además una que la cambiaría sin avisar, porque es `biso snapshot` quien
crea el repositorio del tablero la primera vez que se ejecuta, de modo que el mismo comando en el mismo
sitio respondería una cosa antes y otra después. El precio de no tener ese freno es que un proyecto sin puntero propio hereda el del
proyecto que lo contenga, si lo hay, y se acepta a propósito: la sección 12 de `DECISIONES.md` dice por
qué, y `biso where` enseña siempre de qué directorio salió el puntero que ha resuelto.

Si nada de eso existe, cualquier comando salvo `init`, `where`, `help`, `--help` y `--version` aborta
antes de ejecutar su propia lógica, con código 8 y este mensaje por stderr:

```
error: no board here, and none configured for this project
hint: `biso init` creates one, `biso where` explains what was searched
```

Los cinco exentos no abortan así: `init`, `help`, `--help` y `--version` no necesitan tablero para
hacer su trabajo, y `biso where` lo necesita pero lo comprueba por su cuenta, con su propio mensaje y
su propio código 8 cuando no lo encuentra (sección 10.2).

**El puntero** es el fichero `.biso.json` que `biso init` escribe en el directorio desde el que se le
llama, que es la raíz del proyecto en el uso normal, y que se versiona con el proyecto. Nada obliga a
que esté en la raíz, porque la búsqueda de arriba lo encuentra en cualquier ancestro; ponerlo en la
raíz es lo que hace que lo vean todos los subdirectorios y todas las copias de trabajo. Tiene estas
claves:

| Clave | Tipo | Obligatoria | Notas |
|---|---|---|---|
| `version` | entero | sí | versión del formato del puntero |
| `id` | 8 caracteres hexadecimales en minúscula | sí | la identidad del tablero, y es inmutable. En mayúsculas el puntero es inválido: no se normaliza |
| `path` | ruta del directorio del tablero, absoluta o relativa | no | solo cuando el tablero no vive en una de las raíces de la sección 3.3 |

Al estar versionado, todas las copias de trabajo del proyecto lo ven igual y comparten el mismo
tablero sin ningún paso adicional. Siete reglas gobiernan su lectura:

- Una clave desconocida en el puntero es un error.
- **`path` nombra el directorio del tablero, no el directorio que lo contiene.** Es la ruta que se
  abre, sin concatenarle nada.
- **Una `path` relativa se resuelve respecto al directorio que contiene el fichero puntero**, nunca
  respecto al directorio de trabajo. El motivo es el principio de que ningún comportamiento depende de
  dónde se ejecute el programa (sección 1): con la búsqueda hacia arriba de esta misma sección, el
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
  base de datos discrepan es un error**, y `biso doctor` lo comprueba (10.13).
- **Un directorio con el marcador pero sin la base de datos no es un tablero, y la búsqueda sigue.** No
  es un caso rebuscado: es lo que recibe una copia de trabajo de un proyecto que versionó el directorio
  de su tablero, porque el fichero de exclusión que `init` escribe ahí dentro excluye siempre la base de datos y
  nunca el marcador (10.1). Saltarlo y seguir buscando es lo que encuentra el tablero de verdad un nivel
  más arriba. Si la búsqueda acaba sin nada, el error nombra ese directorio, porque explica el fallo
  mejor que decir solo que no se encontró el tablero.
- **El nombre de la carpeta del tablero es decorativo, y nadie resuelve nunca por él.** Al crearla,
  `biso init` la llama `<slug>-<id>`, por ejemplo `kex-3f9a2b1c`, porque eso hace legible un listado de
  la raíz por defecto. Pero renombrarla no rompe nada, ni la renombra `biso` cuando cambia el nombre
  del tablero (10.10), ni hay comprobación alguna sobre ella.
- **Cómo se busca cuando el `path` no resuelve, o cuando no hay `path`**: se recorren la raíz por
  defecto y después las raíces adicionales (sección 3.3), en ese orden, mirando en cada carpeta si
  contiene el marcador `<id>.id`. Y una `path` relativa que no resuelve en el directorio del puntero se
  prueba, tal cual, contra cada uno de sus ancestros hasta el mismo tope que cierra esta sección. Ese
  último paso es el que hace que un tablero que vive dentro del proyecto se siga encontrando desde una
  copia de trabajo que no lo tiene, como un worktree de git, donde el puntero está versionado y el
  directorio del tablero no.

**Si el mismo `id` aparece en dos sitios, es un error** que nombra los dos directorios y no elige
ninguno, porque elegir sería escribir en un tablero que quien llama no ha nombrado, y dos almacenes con
la misma identidad es exactamente lo que esta persistencia no admite. Sale con **código 11**
(`AMBIGUOUS_BOARD`, sección 2), con la clave `code` `ambiguous_board_id`, y con este mensaje por stderr:

```
error: board 3f9a2b1c is in two places, and biso will not choose between them
        /Users/avilches/.biso/boards/kex-3f9a2b1c
        /Volumes/work/boards/kex-3f9a2b1c
hint: rename or remove one of the two directories
```

Tiene su propio código de salida y no el 8 porque su remedio no se parece a los otros dos: aquí no
falta un tablero que `biso init` pueda crear, sobra uno que solo una persona puede decidir cuál es. El
error lo da cualquier comando que tenga que resolver el tablero, y también `biso init`, que recorre las
mismas raíces para comprobar que el `id` que va a acuñar o a adoptar no exista ya (10.1). **`biso
doctor` no lo comprueba**, y no es un olvido: para llegar a ejecutarse, `doctor` necesita un tablero
resuelto, así que en un tablero duplicado aborta con este mismo error antes de comprobar nada, igual
que hace con la base de datos ilegible de la sección 4.12.

**El `<slug>` del nombre de la carpeta se deriva de `project_name`, nunca es el nombre literal**:
`project_name` es texto libre, y copiarlo tal cual metería espacios y mayúsculas en una ruta que
alguien va a teclear. La derivación reutiliza el mismo paso de `normalizar(x)` (sección 6.1) que ya usa
la derivación de `task_prefix` (sección 4.11): pasar el nombre a minúsculas según Unicode y quitarle
los diacríticos. Después, cada tirada de caracteres que no sean una letra ASCII ni un dígito ASCII se
colapsa en un solo guion, recortando los que queden en los extremos. Así, `Kex` da `kex`, `Mi Proyecto`
da `mi-proyecto`, y `Peña 2026` da `pena-2026`. Ningún carácter de `project_name`, incluido un separador
de ruta, puede romper el nombre de la carpeta, porque cualquier tirada de ellos se colapsa igual en un
guion.

Eso resuelve tres cosas de golpe: que dos proyectos de la misma máquina se puedan llamar igual sin
chocar, porque lo que identifica al tablero no es su nombre sino su `id`; que el puntero siga
resolviendo aunque alguien renombre la carpeta o mueva el tablero a otra raíz, porque el marcador viaja
con el tablero y no con su nombre; y que cambiar el nombre de un tablero no toque el sistema de ficheros
en absoluto (10.10).

**El único caso que sigue pidiendo una corrección a mano** es mover o renombrar a mano el directorio de
un tablero que vive fuera de las raíces de la sección 3.3, porque entonces su `path` deja de resolver y
no hay ninguna raíz que recorrer para encontrarlo. Se arregla con `biso init --at <ruta nueva>`, que
reescribe el puntero adoptando el `id` que ya lleva (10.1).

**El slug y el `task_prefix` arrancan del mismo `project_name` pero fallan por motivos distintos, y
hay que comprobar los dos.** `"2026"` da un slug válido, `2026`, pero no da ningún prefijo, porque no
le queda ninguna letra ASCII (sección 4.11); `"///"` da un slug vacío y también un prefijo vacío. Que
una de las dos derivaciones salga bien no dice nada de la otra, así que ninguna de las dos
comprobaciones sustituye a la otra.

`biso init` genera el `id` de la fuente de números aleatorios del sistema, comprobando que no exista ya
en ninguna de las raíces de la sección 3.3, que es una lectura de directorio por raíz, y escribe el
puntero siempre que no exista ya uno: sin `--at`, con el tablero en la raíz por defecto y **sin clave
`path`**, porque ahí lo encuentra la búsqueda por marcador; con `--at`, con el tablero donde se le diga
y esa ruta en `path` (sección 10.1). Cuando ya existe un puntero sin tablero aquí, la excepción es la de
más abajo: adopta el `id` que ya lleva en vez de generar uno nuevo, y no reescribe el puntero. `biso
where` dice cuál se ha usado y por qué (sección 10.2). No hay ningún caso en el que haya que escribirlo
a mano.

**La forma de la `path` la elige quien llama, con la forma que le da a `--at`** (sección 10.1), y las dos
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

**Dos proyectos distintos pueden apuntar legalmente al mismo tablero.** No hay forma de distinguir "dos
copias de trabajo del mismo proyecto" de "dos proyectos que comparten tablero", porque el mecanismo es
el mismo puntero, y compartir es precisamente para lo que existe.

**Si el puntero se pierde** (se borra a mano, o el proyecto se clona sin haberlo commiteado antes), la
recuperación es explícita, nunca automática, y consiste en una sola cosa: `biso init --at <ruta>` con
la ruta del directorio del tablero que ya existe, que escribe un puntero nuevo con ese `path`,
relativa o absoluta según la regla de arriba (sección 10.1). Es lo único que arregla el proyecto,
porque es lo único que deja el puntero otra vez donde lo ven todos sus subdirectorios y todas sus
copias de trabajo.

**Trabajar con `-C <directorio del tablero>` no es la otra mitad de esa recuperación, sino lo que se
puede hacer mientras tanto**: llega al tablero por la primera vía, pero solo en la llamada en la que
se escribe, y no deja nada apuntado, así que el comando siguiente vuelve a no encontrar nada. Sin
escribir el puntero, el proyecto no vuelve a encontrar su tablero por su cuenta.

**Un tablero cuyo proyecto ya no existe queda huérfano** en la raíz por defecto, y ningún comando de
hoy lo ve.

**Si el puntero existe pero nombra un tablero que esta máquina no tiene** (lo típico al clonar el
proyecto en otro ordenador), el mensaje es otro, porque aquí sí hay un puntero. Dice que hay uno, qué
identificador nombra, y que `biso init` crea el tablero aquí adoptando esa misma identidad:

```
error: this project's pointer names board 3f9a2b1c, which is not on this machine
hint: `biso init` creates it here, adopting id 3f9a2b1c
```

El código de salida sigue siendo 8, porque para quien llama la situación es la misma: no hay tablero
con el que trabajar, y el remedio también es el mismo, `biso init`. Lo que cambia es la clave `code`
del sobre JSON (sección 12), que aquí es `pointer_unresolved` en vez de `no_board`.

---

### 3.3. Configuración de máquina

`biso` necesita, antes de que exista el primer tablero, saber dónde crearlo, y eso no puede depender
de resolver un tablero con la sección 3.2. Por eso vive en un fichero propio de la máquina,
`~/.biso/config.json`, que `biso config` (sección 10.10) no gestiona: aquella es la configuración de
un tablero concreto, y esta es de la máquina entera, independiente de cuántos tableros tenga.

| Clave | Tipo | Por defecto |
|---|---|---|
| `boards_root` | ruta | `~/.biso/boards` |
| `boards_extra_roots` | lista de rutas | vacía |
| `vcs` | `git`, `none` o `custom` | `git` |
| `vcs_custom` | objeto, y solo con `vcs` igual a `custom` | ausente |

`boards_root` es la raíz por defecto de la sección 3.2: el directorio donde `biso init` sin `--at` crea
los tableros nuevos. `boards_extra_roots` son raíces adicionales, para cuando algún tablero vive fuera
de `boards_root`.

**Las dos se recorren igual, y en un orden declarado**: `boards_root` primero y después
`boards_extra_roots` en el orden en que estén escritas. Eso vale en los dos sitios que recorren raíces,
que son la búsqueda de un tablero por su `id` (sección 3.2) y la comprobación de `biso init` de que un
`id` nuevo no exista ya. El orden solo importa para decir qué se lee antes, nunca para elegir entre dos
candidatos: **el mismo `id` en dos raíces es un error** (sección 3.2), no una preferencia. Una raíz de la
lista que no exista o no se pueda leer no es un error, porque una máquina puede tener configurado un
disco que hoy no está montado: se salta y `biso doctor` la reporta como aviso (10.13).

Se leen directamente de ese fichero, sin pasar por la resolución de tablero de la sección 3.2, porque
hace falta conocerlas antes de que exista el primer tablero de la máquina. Y, como en cualquier otro
sitio de `biso`, **una clave desconocida en este fichero es un error**, nunca algo que se ignore en
silencio.

**`vcs` es el sistema de control de versiones con el que `biso snapshot` guarda el historial de la
instantánea** (sección 10.14), y lo único de esta configuración que llega a hacer que `biso` ejecute un
programa ajeno. Vive aquí y no en la configuración de un tablero porque dice qué herramienta hay
instalada en esta máquina, no cómo es un tablero: así no viaja en la instantánea, y restaurar la de otra persona no le impone a nadie el sistema
del ordenador de origen. Vale para todos los tableros de la máquina.

- **`git`** es el valor por defecto y el único sistema que `biso` trae aprendido en la versión 1.0. Su
  receta completa, con las dos preguntas que le hace al repositorio y las tres órdenes que le da, está
  en la sección 10.14, que es la única que lo ejecuta.
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
| `ignore_file` | nombre de fichero | no | el fichero de exclusión que `biso init` escribe dentro del tablero (10.1) |

En `commit` y en `publish`, `{message}` se sustituye por el mensaje que `biso` compone (`biso snapshot:
248 tasks`) y `{files}` por los ficheros de la instantánea, cada uno como un argumento propio. Las dos
se ejecutan con el directorio del tablero como directorio de trabajo, y un código de salida distinto de
cero es error 7 con la clave `code` `vcs_commit_failed`.

---

## 4. Reglas transversales

Estas reglas valen para todos los comandos y no se repiten en cada uno. Un comando solo las menciona
cuando se aparta de ellas, y ninguno lo hace salvo donde se diga.

### 4.1. Interactividad, terminal y color

**Ningún comando abre nunca una interfaz interactiva por su cuenta, y ningún comando pregunta nada.**
No existe la detección de terminal como forma de decidir qué imprime un comando: la salida de
cualquier comando es idéntica byte a byte con terminal y sin él, salvo los códigos de color.

La interfaz interactiva existe, pero es un comando aparte, `biso board`, que solo se ejecuta si se
pide por su nombre y que falla con código 7 si no hay terminal.

Lo único que mira el terminal es el color:

| Situación | Color |
|---|---|
| `--color always` | sí |
| `--color never`, o `NO_COLOR` definida | no |
| `--color auto` y stdout es un terminal | sí |
| `--color auto` y stdout está redirigido | no |

El color se decide por separado para stdout y para stderr, cada uno según su propio destino. Los
códigos de color nunca cambian el texto: quitarlos deja exactamente las líneas documentadas aquí.

### 4.2. stdout, stderr y qué va en cada uno

La regla es fija y no tiene excepciones:

- **stdout lleva datos.** Lo que un programa consumiría: identificadores, listados, fichas, JSON.
- **stderr lleva todo lo demás.** Errores, avisos, notas informativas, sugerencias y el resumen de lo
  que se ha omitido.

En particular, la línea de `biso ls` que dice cuántas tareas se han ocultado va por stderr, porque no
forma parte del listado. Redirigir stdout a un fichero produce un fichero de datos limpio, y
redirigirlo a `/dev/null` no pierde ni un solo aviso.

### 4.3. Notas y avisos

Hay dos clases de mensaje que no son errores, las dos por stderr, y las dos dejan el código de salida
en 0:

- **`note:`** es información de contexto. `--quiet` la suprime.
- **`warning:`** es algo que quien llama necesita saber y que no impide la operación. **Nunca se
  suprime.**

Esta es la lista completa de avisos que el programa emite. No hay ningún otro:

| Aviso | Cuándo |
|---|---|
| `warning: --set-plan replaced 412 bytes of existing content` | cualquier `--set-*` que pise contenido no vacío |
| `warning: TASK-11 moved to Done with 1 of 2 acceptance criteria unchecked` | al llegar a un estado terminal con criterios sin marcar |
| `warning: TASK-11 finished without a final summary` | al llegar a un estado terminal sin resumen |
| `warning: TASK-11 moved to Done with 1 of 3 definition-of-done items unchecked` | al llegar a un estado terminal con la definición de hecho a medias |
| `warning: TASK-11 has unfinished subtasks: TASK-14, TASK-15` | al terminar una tarea con subtareas vivas |
| `warning: TASK-11 is a dependency of TASK-20, which is not finished` | al archivar una tarea de la que dependen otras vivas |
| `warning: --clear-label has no effect on a new task` | cualquier `--clear-*` en `biso new` |
| `warning: TASK-11 has unresolved dependencies: TASK-4 (To Do)` | al empezar una tarea bloqueada |
| `warning: 28 more tasks match; showing 30 of 58` | en `biso ls`, al recortar |
| `warning: --label: "urgent" given twice, kept once` | valor repetido en una bandera de lista |
| `warning: --desc contains a literal \n and no real newline; it will be stored as text` | ver 4.4 |
| `warning: --note: empty value, nothing was added` | valor vacío en una bandera que añade |
| `warning: --due 2026-01-01 is in the past` | fecha límite ya pasada |
| `warning: TASK-11 has no acceptance criteria` | `--check all` sobre una tarea sin criterios |
| `warning: 1 task could not be read and was skipped` | ver 4.12 |
| `warning: <x> is deprecated and will be removed in 2.0` | ver la sección 13 |
| `warning: TASK-11 has an open question, asked by @sara` | al empezar una tarea con una pregunta abierta |
| `warning: TASK-11 moved to Done with an open question, asked by @sara` | al llegar a un estado terminal con una pregunta abierta |
| `warning: TASK-11's lease is held by @sara until 2026-09-08T14:00:00Z` | al escribir sobre una tarea cuyo arrendamiento está vivo y es de otra identidad, con `biso start` o con cualquier otra escritura (sección 5, 9.2 de `DECISIONES.md`, 10.7.1) |

### 4.4. Codificación y texto

- La entrada y la salida son **UTF-8**, siempre, sea cual sea la configuración regional del sistema.
  Una secuencia de bytes inválida en un argumento o en un fichero de entrada es un error con código 3
  que señala la posición del byte.
- Los saltos de línea de salida son `\n`. Al leer una entrada, `\r\n` y `\n` se aceptan por igual y
  se normalizan a `\n`.
- El texto se guarda tal cual llega. **Ninguna secuencia de escape se interpreta.** Un `\n` literal
  de dos caracteres se guarda como dos caracteres.
- Como ese `\n` literal casi siempre es un accidente, un valor de texto que contenga la secuencia de
  dos caracteres `\` `n` y **ningún** salto de línea real produce este aviso, y se guarda igual:
  ```
  warning: --desc contains a literal \n and no real newline; it will be stored as text
  hint: use a real newline, or -d @file.md, or -d - to read from stdin
  ```

### 4.5. Tres formas de pasar un valor largo

Todo parámetro de tipo texto largo (`--desc`, `--plan`, `--note`, `--summary`, `--comment` y el texto
de un criterio o de un elemento de la definición de hecho) acepta las tres:

| Forma | Significado |
|---|---|
| `--desc "texto"` | el texto literal |
| `--desc @ruta/fichero.md` | el contenido del fichero, interpretado como UTF-8 |
| `--desc -` | todo lo que llegue por la entrada estándar hasta el fin de fichero |

Reglas:

- **Un texto que empieza de verdad por `@` se escribe `@@`.** El primer `@` se descarta y el resto es
  literal. Es la única secuencia de escape del programa.
- **Los campos de persona nunca interpretan el `@`.** `--assignee`, `--reporter` y `--comment-author`
  toman su valor tal cual, así que `--comment-author @trello:juan` guarda ese texto y no intenta leer
  ningún fichero.
- **`-` solo puede aparecer una vez por invocación.** Dos parámetros que pidan la entrada estándar son
  un error de uso con código 2, porque el segundo leería un flujo agotado y guardaría el vacío sin
  que se note.
- **Un fichero que no existe es código 4**, con el mensaje `error: --desc: file not found: docs/x.md`.
  Un fichero que existe pero no se puede leer es código 7.
- **Un valor vacío, venga de donde venga, no borra nada.** Ver 4.6.

### 4.6. El valor vacío

Un valor vacío es una cadena sin ningún carácter, o solo con espacios, tanto si llega literalmente
como si llega de un fichero vacío o de una entrada estándar vacía. La regla es única:

| Dónde | Qué pasa |
|---|---|
| En una bandera que añade (`--note`, `--label`, `--ac`, `--desc`) | No se añade nada, se emite `warning: --note: empty value, nothing was added` y el código sigue siendo 0 |
| En una bandera que sustituye (`--set-notes`, `--set-label`) | Deja el campo vacío, igual que `--clear-notes`. Sustituir por nada es vaciar, y eso sí es explícito |
| En un campo escalar (`--type ""`, `--priority ""`) | Error 3. **La cadena vacía nunca es la forma de borrar un escalar**; para eso está `--clear-type` |
| En el título, al crear | Error 2: `error: title cannot be empty` |

**El `code` de un escalar vacío depende de si el campo tiene vocabulario cerrado.** Para `status`,
`type`, `priority` y `project`, una cadena vacía es un valor que no coincide con nada configurado, así
que sigue la regla de 6.1 y el `code` es el de un valor desconocido (`unknown_status` y análogos, con
el mensaje de 6.2). Para los demás escalares (`--reporter ""`, `--ordinal ""`, `--due ""`), que no
tienen vocabulario, el `code` es `empty_scalar_value`.

### 4.7. Valores que empiezan por guion

Tres mecanismos, en orden de preferencia:

1. **`--flag=valor`** funciona siempre y es la forma recomendada: `--desc=-5 grados`.
2. **`--`** termina el análisis de opciones: `biso new -- "-n no es una bandera"`.
3. **Un valor que empieza por guion detrás de una bandera que exige valor se acepta tal cual**, sin
   heurísticas. `biso set TASK-1 --note -x` guarda `-x` como nota.

Como consecuencia de la regla 3, olvidar el valor de una bandera se detecta por lo que sobra después,
no por lo que parece: `biso set TASK-1 --note --priority high` guarda la nota `--priority` y luego
falla con código 2 y `error: unexpected argument: high`.

### 4.8. Repetición y listas separadas por comas

Para toda bandera marcada como repetible:

- Repetirla acumula: `--label a --label b` deja dos etiquetas.
- Si además acepta lista, separar por comas acumula igual: `--label a,b` deja las mismas dos.
- Las dos formas se pueden mezclar.
- **Una coma dentro de un valor se escapa con `\,`.** Es la única forma de meter una coma en una
  etiqueta o en una referencia.
- Los campos de texto largo y los criterios **nunca** se parten por comas.
- Un valor repetido dentro de la misma bandera se guarda una vez y produce
  `warning: --label: "urgent" given twice, kept once`.

Para toda bandera **no** repetible, es decir, los campos escalares, pasarla dos veces con valores
distintos es un error de uso con código 2:

```
error: --status given twice with different values: "In Progress" and "Done"
```

### 4.9. Orden de aplicación dentro de una escritura

Una sola invocación puede tocar muchos campos. El orden en que se aplican es fijo y **no depende del
orden en que aparecen las banderas en la línea de comandos**, para que el resultado sea reproducible:

1. Todos los `--clear-*`.
2. Todos los `--set-*`.
3. Todos los `--rm-*`.
4. Los añadidos, es decir, los nombres desnudos.
5. Los campos escalares.
6. Los marcados de criterios y de definición de hecho.
7. Los comentarios.

Con este orden, `--clear-label --label urgent` deja exactamente una etiqueta, y `--set-ac "A"
--check all` marca los criterios recién puestos. Dentro de un mismo paso manda el orden de la línea
de comandos: `--label b --label a` deja `b` antes que `a`. Las listas nunca se ordenan solas.

### 4.10. Concurrencia, atomicidad y garantías observables

Esta sección no describe un mecanismo: enuncia lo que quien llama tiene derecho a observar. Cómo se
consiga es cosa de quien implemente.

1. **Ninguna escritura se observa a medias.** Un lector concurrente ve el tablero como estaba antes
   de una escritura o como quedó después, nunca en un punto intermedio, y esto vale igual para una
   escritura de una tarea que para un lote de doscientas.
2. **Una escritura que afecta a varias tareas es todo o nada.** Si falla por cualquier motivo, ni una
   sola de las tareas implicadas queda modificada, y el código de salida lo dice: 9 si el fallo se
   detectó al validar, 7 si se detectó al escribir. En los dos casos el mensaje afirma explícitamente
   que no se ha escrito nada.
3. **Dos procesos simultáneos nunca asignan el mismo identificador**, aunque trabajen sobre el mismo
   tablero desde copias de trabajo distintas del proyecto.
4. **Dos escrituras simultáneas sobre la misma tarea no se pierden ni se mezclan.** O se aplican una
   después de otra, o una de las dos falla con código 7.
5. **Si el programa no puede obtener el acceso exclusivo que necesita para escribir**, espera hasta
   cinco segundos y luego falla con código 7 sin escribir nada:
   ```
   error: the board is busy, another process is writing to it
   hint: retry in a moment; nothing was written
   ```
6. **Las lecturas nunca fallan por culpa de una escritura en curso**, y nunca la bloquean.

**Las seis hablan de las escrituras del tablero, y hay un solo comando que escribe ficheros de texto con
nombre fijo, `biso snapshot`.** Sus garantías son otras y están en la sección 10.14: cada fichero se
escribe en un temporal y se renombra encima, los dos temporales se completan antes de renombrar
ninguno, y el comando no toma ningún acceso exclusivo, precisamente para que una copia no pueda hacer
fallar a la escritura de una tarea. Ninguna de las seis de aquí queda tocada por eso.

### 4.11. Identificadores

- Un identificador es `<PREFIX>-<n>`, con `n` entero positivo. `PREFIX` viene de la configuración
  (`task_prefix`).
- **`task_prefix` no tiene un valor fijo por defecto: se deriva del nombre del tablero
  (`project_name`) en mayúsculas.** Dos tableros con `TASK` como valor fijo colisionarían los dos en
  `TASK-1`, y eso haría inservible cualquier vista que junte tareas de varios proyectos.
- **La derivación quita del nombre los caracteres que no son letras y pasa el resto a mayúsculas**,
  así que un tablero llamado `mi-proyecto-2` da el prefijo `MIPROYECTO`. Si al quitarlos no queda
  ninguna letra, como en un tablero llamado `2026`, `biso init` no se inventa un valor: falla y pide
  el prefijo explícitamente con `--prefix` (error 2, `code` `invalid_prefix`, sección 12.3), la misma
  clave que ya cubre un `--prefix` con algo que no sean letras (sección 10.1).
- **"Letra" no incluye los diacríticos**, para que un nombre de tablero con cualquier carácter
  Unicode derive un prefijo predecible. La derivación pasa primero el nombre por el paso de
  `normalizar(x)` (sección 6.1) que quita los acentos, las diéresis y las cedillas, y solo entonces
  se queda con lo que sean letras ASCII. Así un tablero llamado `Peña` deriva `PENA`, y uno llamado
  `Café` deriva `CAFE`.
- **Un identificador no se reutiliza jamás**, ni después de archivar una tarea ni después de
  eliminarla por cualquier vía.
- Los identificadores se asignan de forma creciente, pero **la especificación no promete que la
  secuencia no tenga huecos**. Un hueco es normal y nunca es un error.
- El tablero sabe en todo momento cuál es el identificador más alto que ha llegado a asignar, y ese
  dato se usa en los mensajes de la sección 7.3 y en `biso doctor`.

### 4.12. Qué pasa con un dato que no se puede interpretar

Hay dos motivos distintos por los que un dato resulta imposible de leer, y con una base de datos son
dos casos que hay que separar: uno es que una tarea concreta esté dañada mientras el resto del
tablero sigue legible, y el otro es que no haya tablero legible en absoluto. El segundo no es una
variante del primero.

#### El primer caso: una tarea ilegible

Una tarea puede resultar ilegible: la base de datos devuelve algo corrupto para esa fila, o la tarea
lleva una clave de extensión que la configuración ya no declara. El resto del tablero sigue legible, y
la regla depende del tipo de lectura:

| Tipo de lectura | Qué pasa |
|---|---|
| **Lectura dirigida** a esa tarea, es decir, `get`, o `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer` y `archive` con una referencia que resuelve a ella | Error 3, con el motivo exacto. No se escribe nada |
| **Lectura de conjunto**, es decir, `ls`, `prime`, `export`, `snapshot`, la resolución de una referencia por texto y cualquier filtro | La tarea se salta, se cuenta, y al final se emite `warning: 1 task could not be read and was skipped` con sus identificadores. El resto del resultado es válido y el código es 0, **salvo en `biso export` y en `biso snapshot`, que salen con 6** |
| `biso doctor` | Se reporta como problema y se sigue con las demás. Nunca aborta |

Una lectura de conjunto **nunca** aborta por una tarea mala, y **nunca** la esconde en silencio. Las
dos cosas juntas son lo que impide que un listado incompleto se confunda con un tablero vacío.

**`biso export` y `biso snapshot` son las dos excepciones al código 0 de una lectura de conjunto.**
Los dos escriben igual todo lo que han podido leer, con el mismo aviso por stderr, pero terminan con
**código 6** en vez de 0 cuando han saltado alguna tarea: son los dos comandos cuyo propósito es
servir de copia fiel del tablero, así que una copia incompleta no puede parecer un éxito llano. Un
guion que encadene `biso export -o backup.ndjson && ...` o `biso snapshot && ...` puede comprobar el
código de salida para detectar un volcado incompleto.

#### El segundo caso: la base de datos que no se puede leer

Esto **no es una tarea ilegible: es que no hay tablero legible.** Si la base de datos no abre, o abre
pero falla su comprobación de integridad, ninguna tarea es alcanzable, así que no tiene sentido
presentarlo como "una tarea se ha saltado": no hay nada que saltar, hay un tablero entero fuera de
alcance. La regla es la misma para cualquier operación, sea una lectura dirigida, una de conjunto, una
escritura o `biso doctor`: el comando aborta entero, sin escribir nada, con este mensaje por stderr:

```
error: board 3f9a2b1c's database could not be read
hint: it did not open, or it failed its integrity check, and there is no automatic repair
hint: rebuild it in place with `biso init --from <snapshot dir>`, which keeps its id
```

El código de salida es **10** (`DAMAGED`, sección 2), y no el 8 de la ausencia de tablero, porque el
remedio es otro: aquí el tablero está donde tiene que estar y lo que hay que hacer es reconstruirlo, no
crearlo. La clave `code` del sobre JSON (sección 12.3) es `database_unreadable`.

**El código 10 puede salir de cualquier comando, y por eso no se repite en la tabla de códigos de
salida de cada uno.** Esas tablas dicen los desenlaces propios del comando; este no lo es de ninguno, es
el del tablero entero, igual que el 1 de un fallo del programa, que tampoco aparece en ellas.

**Y el remedio se puede teclear tal cual, porque el segundo `hint` nombra el comando que lo hace.** Un
directorio cuya base de datos no abre no cuenta como tablero accesible para `biso init`, así que
`biso init --from` reconstruye ahí mismo en vez de dar el error 2 de "ya hay uno" (10.1), y adopta el
`id` que nombra el marcador de la instantánea, de modo que el puntero commiteado del proyecto sigue
valiendo. Es también lo que necesita un clon recién traído a otra máquina, que llega con el directorio
del tablero versionado y sin base de datos dentro.

#### Lo que se pierde al pasar de un fichero por tarea a una base de datos compartida

Con un almacén de un fichero por tarea, el aislamiento del daño sale gratis: una tarea corrupta es
exactamente eso, una tarea corrupta, y las demás siguen intactas porque viven en ficheros distintos.
Con una base de datos compartida, ese aislamiento hay que provocarlo a propósito, y cuando falta, una
sola corrupción puede llevarse por delante más de una tarea a la vez, o el tablero entero. Esta
especificación dice las cosas incómodas en voz alta en vez de esconderlas: ese es el coste real de
guardar los datos en una base de datos, frente a la alternativa de un fichero por tarea.

### 4.13. El presupuesto de arranque

**`biso ls` y `biso prime` sobre un tablero de 300 tareas terminan en menos de 25 milisegundos de
reloj.** Es una regla transversal y no solo de esos dos comandos: ninguno de los demás tiene un motivo
para tardar más que ellos. Es una prueba de la suite, no una aspiración, y se mide en la máquina de
referencia.

**La máquina de referencia es la que ejecuta la suite de integración continua del proyecto.** Una
cifra de tiempo sin una máquina no se puede comprobar, porque la misma llamada tarda lo que tarde el
hardware que la ejecuta. Fijar la máquina de referencia como la del propio CI, en vez de describir aquí
un modelo de hardware concreto, es lo que hace que la prueba dé siempre el mismo veredicto para la
misma versión del código, sin que este documento tenga que llevar ni mantener actualizada una ficha
técnica de un ordenador que además dejaría de existir o de venderse.

**Para qué sirve la cifra: detectar una regresión en la máquina que ejecuta la suite, no certificar el
rendimiento de la herramienta sobre un hardware arbitrario.** Quien corra la suite en una máquina
distinta de la de referencia, más lenta o más rápida, no debe leer el resultado como una afirmación
sobre `biso`: debe leerlo como una afirmación sobre esa máquina. Un portátil viejo que supere los
25 ms no dice que la herramienta incumpla su especificación, dice que en ese portátil el número es
otro; lo que sí dice algo es que el número suba en la propia máquina de referencia de una versión a
la siguiente, porque ahí el hardware no ha cambiado y lo único que puede haber cambiado es el código.

**La composición del tablero de 300 tareas es indiferente, y por eso no se fija.** No importa cuántas
estén en cada estado, ni si alguna tiene una pregunta abierta o un arrendamiento vencido: ninguno de
los dos comandos se ramifica según el contenido de una tarea concreta, así que su coste crece de forma
esencialmente lineal con el número de tareas y no con su composición. Fijar un reparto arbitrario no
añadiría ninguna garantía que esta razón no dé ya, y la ambigüedad se cierra con la explicación, no con
una tabla de reparto que nadie necesita reproducir.

Tres reglas protegen ese presupuesto, y ningún comando se aparta de ellas:

1. **Ningún comando hace al arrancar trabajo que nadie ha pedido.** Ni una consulta que no alimente
   una línea de lo que esa invocación va a imprimir, ni una comprobación de más, ni una llamada de
   red: todo lo que no sirve a la salida de la llamada concreta se paga en cada una de las muchas
   veces que un agente ejecuta el programa a lo largo de una sesión, se haya pedido o no. El ejemplo
   documentado está en este documento y no en otra herramienta: la resolución del tablero llevaba una
   comprobación de un `.git` en cada directorio del camino hacia arriba, para frenar la búsqueda del
   puntero, y se retiró al ver que además de costar comprobaciones en cada llamada no protegía de lo
   que pretendía (sección 12 de `DECISIONES.md`).
2. **Ningún comando ejecuta un programa ajeno en su camino caliente.** `biso snapshot` es la única
   excepción, y así lo dice la sección 10.14. Invocar `git`, que es el sistema de control de versiones
   por defecto, cuesta unos 12 milisegundos medidos, casi la mitad de este presupuesto entero gastada en
   una sola llamada. Este presupuesto es además la razón por la que la resolución del tablero de la
   sección 3.2 no mira el control de versiones en absoluto, ni siquiera leyendo ficheros: frenar la
   búsqueda del puntero donde un repositorio empieza obligaría a respetar sus reglas de exclusión para
   que el freno significara algo, y eso no se puede reimplementar de forma fiable ni preguntar sin
   invocar el programa. Que `biso snapshot` sí haga esas dos preguntas (10.14) no contradice nada:
   ese comando ya está fuera del camino caliente por definición.
3. **La palanca mayor no es que cada llamada sea más rápida: es que haga falta hacer menos llamadas.**
   Para eso existe `biso prime` (sección 9), que sustituye el ciclo entero de leer guías sueltas y
   encadenar comandos por un solo mensaje al principio de la sesión; `docs/DECISIONES.md`, sección 3,
   mide lo que cuesta la alternativa de no tenerlo.

---

## 5. El modelo de datos de una tarea

Este es el modelo **lógico**. Describe qué campos tiene una tarea, de qué tipo son y quién los
escribe. No dice nada de cómo se guardan.

| Campo | Tipo lógico | Obligatorio | Quién lo fija | Mutable |
|---|---|---|---|---|
| `id` | identificador `PREFIX-<n>` | sí | el programa | no |
| `title` | texto de una línea | sí | quien llama | sí |
| `status` | uno del vocabulario de estados | sí | quien llama | sí |
| `type` | uno del vocabulario de tipos | no | quien llama | sí |
| `priority` | uno del vocabulario de prioridades | no | quien llama | sí |
| `project` | uno del vocabulario de proyectos | no | quien llama | sí |
| `milestone` | texto de hito | no | quien llama | sí |
| `parent` | referencia a otra tarea | no | quien llama | sí |
| `assignees` | lista de textos de persona | no | quien llama | sí |
| `reporter` | texto de persona | no | el programa al crear, o quien llama; ver 5.6 | sí |
| `labels` | lista de textos | no | quien llama | sí |
| `dependencies` | lista de referencias a tareas | no | quien llama | sí |
| `references` | lista de textos | no | quien llama | sí |
| `documentation` | lista de textos | no | quien llama | sí |
| `modifiedFiles` | lista de textos | no | quien llama | sí |
| `due` | fecha `YYYY-MM-DD` | no | quien llama | sí |
| `ordinal` | entero >= 0 | no | quien llama | sí |
| `createdAt` | instante UTC | sí | el programa | solo al importar |
| `updatedAt` | instante UTC | sí | el programa | solo al importar |
| `archived` | booleano | sí, `false` por defecto | el programa, con `biso archive` | sí, solo con `biso archive` / `--unarchive`, o al importar |
| `leaseExpiresAt` | instante UTC | no | el programa, a `ahora + lease_minutes` (clave de configuración, 10.10); ver la sexta precisión de abajo para cuándo | sí, ver las tres últimas precisiones de abajo, o al importar |
| `leaseHolder` | texto de persona | no | el programa, solo con `biso start` (10.7.1) y con `biso new --start` (10.3); ver la sexta precisión de abajo | sí, solo con esos dos, o al importar; ver las tres últimas precisiones de abajo |
| `urgency` | decimal, derivado | derivado | el programa | no, se recalcula al leer |
| `ext` | mapa de clave declarada a texto | no | quien llama | sí |
| `description` | texto largo | no | quien llama | sí |
| `plan` | texto largo | no | quien llama | sí |
| `notes` | texto largo | no | quien llama | sí |
| `summary` | texto largo | no | quien llama | sí |
| `acceptanceCriteria` | lista de criterios | no | quien llama | sí |
| `definitionOfDone` | lista de criterios | no | quien llama | sí |
| `comments` | lista de comentarios | no | quien llama | solo se añade |
| `question` | registro de tres partes | no | mixto, según la parte; ver 5.7 | sí, solo con `biso ask`, `biso answer`, o al importar |
| `acDone`, `acTotal`, `dodDone`, `dodTotal` | entero, derivado | derivado | el programa | no, se recalculan al leer |
| `commentCount` | entero, derivado | derivado | el programa | no, se recalcula al leer |
| `blocks` | lista de referencias, derivado | derivado | el programa | no, se recalcula al leer |
| `blocked`, `waiting` | booleano, derivado | derivado | el programa | no, se recalculan al leer |
| `leaseExpired` | booleano, derivado | derivado | el programa | no, se recalcula al leer |

Ocho precisiones sobre la mutabilidad:

- **"No mutable" significa que ninguna bandera del programa lo cambia.** `updatedAt` lo reescribe el
  programa en cada operación que cambie algo.
- **Un comentario no se edita ni se borra, solo se añade.** Un comentario es el registro de una
  conversación.
- **`archived` solo lo cambia `biso archive` y `biso archive --unarchive`.** No hay una bandera de
  campo de la sección 8 para él: archivar es un gesto de flujo de trabajo con nombre propio,
  según el principio 5.
- **Los campos marcados "derivado" en esta tabla no se guardan.** Se calculan al leer, y son
  exactamente los campos que `biso export` no escribe (10.9) y que `biso new --from` rechaza como
  clave desconocida (10.3): `urgency`, `acDone`, `acTotal`, `dodDone`, `dodTotal`, `commentCount`,
  `blocks`, `blocked`, `waiting` y `leaseExpired`. Esta es la única lista de campos derivados del
  documento; las demás secciones remiten a ella.
- **`leaseExpired` no cambia el `status` guardado, nunca.** Vale cierto cuando `leaseExpiresAt`
  tiene valor y ese instante es anterior al reloj de quien lee, y **vale falso cuando
  `leaseExpiresAt` está vacío**, que es el caso de toda tarea sin arrendamiento: no hay ningún
  estado en el que este derivado se quede sin valor, porque un derivado que no se pudiera calcular
  es justo lo que el principio 1 de la sección 1 no admite. Dice que el arrendamiento de una tarea
  activa venció, pero el estado guardado sigue siendo el activo hasta que alguien lo cambia con una
  escritura explícita: lo que vence es la reclamación, no el estado (sección 9.2 de
  `DECISIONES.md`). No hay una escritura diferida que la saque del estado activo por su cuenta, porque
  eso haría que un comando tocara tareas que no nombró, y porque `biso prime`, que no escribe nunca,
  mostraría un estado que una escritura ajena y posterior podría cambiar. Liberar el arrendamiento
  vencido es la reclamación explícita que hace `biso start` (10.7.1), no un efecto secundario de
  ningún otro comando.
- **Renovar `leaseExpiresAt` y fijar o transferir `leaseHolder` son cosas distintas, y solo la
  segunda pasa por `biso start` o por su atajo `biso new --start`.** Cualquier escritura sobre una
  tarea activa y asignada renueva `leaseExpiresAt` a `ahora + lease_minutes` (10.10), pero solo
  cuando quien llama ya es `leaseHolder`. **Cualquier escritura son todas**, sin ninguna excepción:
  los seis verbos del ciclo (10.7), `biso set` (10.6) y `biso archive` (10.8), que son los ocho
  comandos que llegan a escribir sobre una tarea que ya existe. Se nombran aquí porque una regla
  general que no nombra a nadie invita a buscarle excepciones donde no las hay. **Una escritura que
  no cambia ningún campo renueva igual**: `biso set` con todas sus banderas dando el valor que la
  tarea ya tiene sale con código 0 y con `note: TASK-11 unchanged` (10.6), y aun así renueva
  `leaseExpiresAt`, porque sigue siendo una escritura del tenedor sobre su tarea y el latido no
  puede depender de si los valores coincidían por casualidad. Esa renovación no toca `updatedAt`,
  porque ningún campo de la tarea ha cambiado, y deja vacía la lista `changed` del esquema JSON de
  10.6; la nota sigue siendo cierta, porque habla de los campos de la tarea y ninguno cambió. Si la
  tarea no tiene arrendamiento todavía, escribir sobre ella no lo crea: fijarlo por primera vez es
  parte de lo que hace `biso start`, igual que reclamarlo vencido o tomarlo de otra identidad
  (10.7.1). Una escritura de una identidad distinta de `leaseHolder` mientras el arrendamiento está
  vivo no toca ninguno de los dos campos: avisa con el mismo
  `warning: TASK-11's lease is held by @sara until 2026-09-08T14:00:00Z` de 10.7.1 y de la tabla de
  la sección 4.3, y el resto de la escritura se hace igual. **Con una sola excepción, y es que esa
  misma escritura rompa la invariante de la precisión siguiente**: si deja la tarea fuera del estado
  activo, sin ninguna persona asignada o archivada, los dos campos se vacían en esa misma escritura,
  sea quien sea quien la haga, y el aviso de que el arrendamiento era de otra identidad se emite
  igual. Una escritura de una identidad distinta mientras el arrendamiento está vencido tampoco lo
  toca, y lo deja vencido: quien comenta, anota o cierra una tarea no ha reclamado nada. **Reclamar
  es de `biso start` (10.7.1) y de su atajo `biso new --start` (10.3), y de nadie más**, con una
  excepción que hay que nombrar porque sin ella la frase sería falsa: `biso start -s <estado>` con
  un estado que no es el activo no fija arrendamiento, ya que fijarlo ahí rompería la invariante de
  la precisión siguiente, y deja los dos campos como los dejaría cualquier otra escritura. Una tarea
  que llega a activa y asignada por cualquier otra vía no tiene arrendamiento hasta que alguien
  llame a `biso start` sobre ella, y esas vías son exactamente dos: las banderas de campo de la
  sección 8, por ejemplo `biso set --status`, ninguna de las cuales lo puede crear, y la
  importación, que es de lo que trata la última precisión.
- **Los dos campos solo tienen valor en una tarea activa y asignada, y se vacían al perder
  cualquiera de las dos condiciones, no solo la primera.** Una escritura que saca la tarea del
  estado activo (`biso finish`, o `biso set --status` a cualquier otro valor) vacía
  `leaseExpiresAt` y `leaseHolder` en esa misma escritura. Y como la condición que los sostiene es
  la conjunción de las dos cosas, perder la segunda los vacía igual: `--clear-assignee` o
  `--rm-assignee` (8.2) sobre una tarea activa que se queda sin ninguna persona asignada vacía los
  dos campos en esa misma escritura, sea quien sea quien la haga. **`biso archive` (10.8) los vacía
  también**, aunque `archived` no sea un estado y archivar no saque la tarea del estado activo:
  archivar es dejar de trabajar en la tarea, y un arrendamiento es la afirmación de que alguien está
  trabajando ahora, así que conservarlo lo guardaría donde nadie lo ve, porque `biso prime` y
  `biso ls` excluyen las archivadas por defecto, y `--unarchive` la devolvería al tablero semanas
  después a nombre de una sesión que ya murió. **Esta precisión gana siempre sobre la anterior, y
  por eso la invariante se enuncia aquí y el aviso allí.** Cuando quien escribe no es
  `leaseHolder`, el aviso de que el arrendamiento es de otra identidad se emite igual, pero los dos
  campos se vacían: `@sara` haciendo `biso finish TASK-11` sobre una tarea arrendada por `@claude` la
  deja terminada y sin arrendamiento. Con la precedencia al revés quedaría una tarea terminada con un
  arrendamiento vivo, que es exactamente lo que la última precisión rechaza al importar, así que
  `biso export` produciría un fichero que su propio `biso init --from` rechaza y la prueba de
  simetría de la sección 13 fallaría (sección 9.2 de `DECISIONES.md`). **Y los dos campos van
  siempre juntos**: ninguna escritura, y tampoco la importación, deja uno con valor y el otro vacío.
- **La importación los escribe con el valor que traiga el fichero, y es la única vía que lo hace.**
  Los dos son campos guardados y no derivados, así que `biso export` los escribe y `biso new --from`
  los lee de vuelta como cualquier otro, que es lo que hace cierta la garantía de simetría de 10.9
  sin una lista de excepciones que mantener. La invariante de la precisión anterior se comprueba al
  importar, y en sus dos mitades. Una línea que traiga `leaseExpiresAt` o `leaseHolder` sobre una
  tarea que no esté a la vez en el estado activo y asignada a alguien es un fallo de validación del
  lote (10.3), igual que una clave desconocida. Y una línea que traiga uno de los dos campos y no el
  otro es el mismo fallo, con el mismo trato: los dos vienen juntos o no viene ninguno, porque un
  `leaseHolder` sin `leaseExpiresAt` sería un arrendamiento que no caduca nunca, y un
  `leaseExpiresAt` sin `leaseHolder` una reserva de nadie. Un arrendamiento importado no privilegia
  a nadie: `leaseExpired` se recalcula contra el reloj de la máquina que lee, así que el que llegue
  caducado sale caducado y `biso start` lo reclama (10.7.1), y el que llegue vivo a nombre de otra
  identidad solo produce el aviso de la sección 4.3 hasta que caduque.

### 5.1. Los criterios y sus claves estables

`acceptanceCriteria` y `definitionOfDone` son listas del mismo tipo, y cada elemento tiene tres cosas:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `key` | entero positivo | el programa al crear el elemento |
| `text` | texto | quien llama |
| `checked` | booleano | quien llama |

**La clave se asigna al crear el elemento, con un contador propio de esa lista dentro de esa tarea, y
no se reasigna nunca.** Quitar un elemento no mueve las claves de los demás: una tarea puede tener
perfectamente los criterios `#1` y `#3` y ninguno más. Cada tarea lleva dos contadores, uno por
lista, que solo crecen.

Consecuencias que hay que respetar en toda la implementación:

- Los selectores de la sección 8.4 trabajan sobre la clave, **nunca** sobre la posición.
- `acTotal` y `dodTotal`, allá donde aparezcan, son **el número de elementos presentes**, nunca la
  clave más alta. Una tarea con los criterios `#1` y `#3` tiene `acTotal` igual a 2.
- Los elementos se muestran y se exportan en el orden en que están en la lista, que es el orden en
  que se crearon salvo que se haya sustituido la lista entera.

### 5.2. Los comentarios

Cada comentario tiene autor, instante y cuerpo:

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | texto libre | quien llama, y por defecto la identidad `me` |
| `createdAt` | instante UTC | el programa, salvo al importar |
| `body` | texto largo | quien llama |

**El autor es texto libre y no se valida contra nada.** Un comentario puede venir de alguien que no
existe en este tablero, y un sistema externo puede usar su propia convención, por ejemplo
`@trello:juan`.

**Los comentarios se guardan y se muestran en orden de inserción, no en orden de `createdAt`.** El
instante de cada uno sigue diciendo la verdad sobre cuándo se escribió, aunque la lista completa no
quede ordenada por él: `biso answer` añade al final un comentario con un instante pasado, el de la
pregunta que responde.

### 5.3. Las fechas

`createdAt`, `updatedAt` y el instante de cada comentario los pone el programa con el reloj del
sistema, en UTC y con precisión de segundo.

**Se pueden fijar solo al importar**, es decir, en `biso new --from`. En cualquier otro sitio son un
hecho observado y no un dato que se negocie.

**Una excepción de forma, no de fondo:** `biso answer` escribe el comentario en que se convierte la
pregunta con el instante en que esa pregunta se hizo, no con el de la respuesta. No negocia nada,
porque ese instante ya lo había observado el programa al crear la pregunta; solo lo traslada.

`question.askedAt` es una cuarta fecha importable, junto a `createdAt`, `updatedAt` y el instante de
cada comentario, y sigue la misma regla que ellas: es opcional, y si `biso new --from` no la trae,
toma el instante de la importación.

**`leaseExpiresAt` es la quinta fecha importable y es la única que no sigue esa regla**, así que se
cuenta aparte a propósito. Si no viene, no se rellena con nada: la tarea llega sin arrendamiento, que
es lo que significa no traerlo. Rellenarla con el instante de la importación crearía un arrendamiento
que nadie ha reclamado, y encima a nombre de nadie, porque `leaseHolder` no es una fecha y no tiene
ningún valor por defecto que ponerle. Los dos vienen juntos o no viene ninguno (precisión octava de
esta misma sección).

### 5.4. La urgencia

`urgency` es un decimal derivado que se recalcula en cada lectura y **nunca se guarda**. Es el segundo
criterio de la tupla de orden por defecto de `biso ls`, después de `ordinal` (10.4), y el que ordena
el resumen de `biso prime`.

```
Si el estado de la tarea es el terminal, urgency = 0.0 y no se calcula nada mas.

En cualquier otro caso:

urgency = 6.0  * prioridad         (high 1.0, medium 0.5, low 0.0, sin prioridad 0.3)
        + 4.0  * activa            (1.0 si el estado es el activo y no hay pregunta abierta, 0.0 si no)
        + 8.0  * bloquea           (1.0 si alguna tarea sin terminar depende de esta)
        - 5.0  * bloqueada         (1.0 si depende de alguna tarea sin terminar)
        + 12.0 * proximidad        (ver la regla siguiente)
        + 1.0  * tiene_criterios   (1.0 si tiene al menos un criterio de aceptacion)
        + 0.5  * min(edad_dias / 30, 4.0)

El resultado se redondea a un decimal.
```

**La regla de `proximidad`, sin ambigüedad:**

```
dias = fecha_limite - hoy, en dias (puede ser negativo si la fecha ya paso)

si la tarea no tiene fecha limite:  proximidad = 0.0
si la tiene:                        proximidad = clamp((30 - dias) / 30, 0.0, 1.0)
```

Una tarea vencida tiene `dias` negativo, así que `(30 - dias) / 30` supera 1 y el resultado se acota
en **1.0**, el mismo máximo que una tarea que vence hoy. Una tarea vencida no suma más que una que
vence hoy; para distinguirlas está el filtro `--overdue` de `biso ls`, no un término sin tope en la
fórmula.

Un ejemplo completo, que es el que imprime `biso get --explain-urgency` en la sección 10.5: una tarea
de prioridad alta, en el estado activo, de la que depende otra tarea sin terminar, sin fecha límite,
con dos criterios y creada hoy, suma `6.0 + 4.0 + 8.0 + 0.0 + 0.0 + 1.0 + 0.0`, es decir **19.0**.

**Los coeficientes configurables son exactamente siete, bajo `urgency.`, uno por término de la
fórmula**: `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`,
`urgency.criteria` y `urgency.age`, con los valores de arriba (6.0, 4.0, 8.0, -5.0, 12.0, 1.0 y 0.5)
como valores por defecto. **Los pesos por prioridad no son configurables**: `high 1.0, medium 0.5,
low 0.0, sin prioridad 0.3` son parte de la estructura fija de la fórmula, que no cambia en la
versión 1.0.

**El `ordinal` no forma parte de la urgencia.** Es un orden manual que se aplica aparte, según la
regla de orden completa de la sección 10.4.

### 5.5. Los campos externos

`ext` es un mapa de clave a texto para guardar la identidad de una tarea en otro sistema. La regla
es la siguiente:

- El tablero **declara** en su configuración qué claves admite, en la lista `extensions`.
- Escribir una clave declarada funciona: `biso set TASK-1 --ext trello.card=5f2a8c1e3b9d4a7f`.
- Escribir una clave no declarada es error 3:
  ```
  error: unknown extension key: "jira.key"
         declared keys on this board: trello.card, github.issue
  ```
- Una tarea que ya guarda una clave que la configuración no declara **no se lee en silencio ni se
  reescribe perdiéndola**: se aplica la regla de 4.12, y `biso doctor` la reporta.

### 5.6. Quién reporta una tarea

`reporter` se fija una sola vez, al crear la tarea, y después solo cambia si alguien pasa
`--reporter` de forma explícita.

| Al crear la tarea | Valor de `reporter` |
|---|---|
| Se pasa `--reporter <persona>` | esa persona, tal cual |
| No se pasa, y hay identidad configurada | la identidad de quien llama, según la precedencia de 3.1 |
| No se pasa, y no hay identidad configurada | vacío, sin aviso |
| Se pasa `--reporter ""` | vacío |

El caso sin identidad no es un error y no imprime nada: a diferencia de `--mine`, de la
autoasignación de `biso start`, del autor de un comentario, de `biso ask` y de `biso answer`, que sí
la necesitan y están cubiertos por la tabla de 3.1, una tarea sin quien la reporte es válida.

En el lote de `biso new --from`, un objeto que trae `reporter` conserva ese valor, y uno que no lo
trae aplica las mismas reglas de esta tabla.

### 5.7. La pregunta abierta

`question` es un registro de tres partes, con la misma forma que un comentario (5.2):

| Parte | Tipo | Quién la fija |
|---|---|---|
| `author` | texto libre | el programa, con la identidad `me`, salvo al importar |
| `askedAt` | instante UTC | el programa, salvo al importar |
| `body` | texto largo | quien llama |

Vacío es lo normal. Con contenido significa que la tarea espera la respuesta de una persona, esté en
el estado que esté, y entonces el derivado `waiting` es cierto; vacío, `waiting` es falso. Lleva tres
partes y no una sola porque al responderse se convierte literalmente en un comentario, con `biso
answer`, y para eso hacen falta su autor y su instante originales, no los de quien responde.

---

## 6. Los vocabularios del tablero y la regla de validación

Tres campos tienen vocabulario cerrado, definido en la configuración: `status`, `type` y `priority`.
Un cuarto, `project`, lo tiene solo si el tablero declara proyectos. Para todos ellos rige una sola
regla, **idéntica al escribir y al leer**.

### 6.1. El algoritmo de coincidencia

Dado un valor de entrada `v` y la lista de valores configurados, el programa calcula así:

```
normalizar(x):
  1. pasar x a minusculas segun Unicode
  2. descomponer y quitar los diacriticos (acentos, dieresis, cedillas)
  3. eliminar TODOS los caracteres que sean espacio, tabulador, guion (-) o guion bajo (_)
  4. devolver lo que queda

coincidir(v, configurados):
  a. si existe un configurado c con c == v exactamente, devolver c
  b. si no, calcular normalizar(v) y compararlo con normalizar(c) de cada configurado
  c. si exactamente un configurado coincide, devolverlo
  d. si ninguno coincide, error 3
  e. si coinciden dos o mas, error 3 con los dos listados, porque el tablero
     tiene dos valores que se normalizan igual y hay que desambiguarlos
```

Con este algoritmo, y para un tablero cuyo estado es `To Do`:

| Entrada | `normalizar` | Resultado |
|---|---|---|
| `To Do` | `todo` | coincide, por el paso a |
| `todo` | `todo` | coincide |
| `TODO` | `todo` | coincide |
| `To-Do` | `todo` | coincide |
| `TO_DO` | `todo` | coincide |
| `to  do` | `todo` | coincide |
| `Pending` | `pending` | error 3 |
| `Todos` | `todos` | error 3 |

**No hay coincidencia por prefijo ni por parecido.**

### 6.2. El mismo texto vale lo mismo en los dos sentidos

Esta tabla es el contrato, y es la prueba de aceptación que hay que poder ejecutar. Tablero con los
estados `To Do`, `In Progress` y `Done`:

| Entrada | `biso set TASK-1 -s <v>` | `biso ls -s <v>` |
|---|---|---|
| `To Do` | escribe | filtra |
| `todo` | escribe | filtra |
| `TO_DO` | escribe | filtra |
| `In-Progress` | escribe | filtra |
| `Pending` | error 3 | error 3 |
| `""` | error 3 | error 3 |

El mensaje es el mismo en los dos sentidos:

```
error: unknown status: "Pending"
       valid statuses on this board: To Do, In Progress, Done
```

### 6.3. Qué valida cada filtro, y contra qué

| Filtro | Conjunto contra el que valida | Si no encaja |
|---|---|---|
| `--status`, `--type`, `--priority`, `--project` | el vocabulario configurado | error 3 |
| `--label` y `--label-or` | el conjunto de etiquetas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--assignee` | el conjunto de personas del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--milestone` | el conjunto de hitos del tablero, definido abajo | error 3, con las cinco más parecidas |
| `--parent` | la resolución de referencias de la sección 7 | error 2, 4 o 5 |
| `--search` | nada, es texto libre | nunca falla |

**El conjunto de etiquetas del tablero** es la unión de las etiquetas declaradas en la clave `labels`
de la configuración y de las que lleva cualquier tarea del tablero, **incluidas las archivadas y las
que están en el estado terminal**. **El conjunto de personas se define con la clave `assignees` de la
configuración y con los valores de `assignees` de cualquier tarea, archivadas y terminadas incluidas.
Los valores de `reporter` no entran en este conjunto**, porque no hay ningún filtro `--reporter`: una
persona que solo ha reportado tareas y nunca las ha tenido asignadas no pertenece al conjunto contra
el que valida `--assignee`.

**El conjunto de hitos del tablero es solo derivado**: son los valores de `milestone` que lleva
cualquier tarea del tablero, **archivadas y terminadas incluidas**, y nada más. Es el único de los
tres que no tiene mitad declarada, porque no existe ninguna clave `milestones` en la configuración
(10.10) ni ninguna bandera de `biso init` que la escriba, así que **un hito existe exactamente
mientras alguna tarea lo lleve escrito**. Un tablero en el que ninguna tarea tiene hito tiene el
conjunto vacío, y entonces cualquier `--milestone` es error 3; el mensaje lo dice tal cual, sin
sugerencias, porque no hay ninguna que ofrecer.

**Ni las etiquetas, ni las personas, ni los hitos tienen vocabulario cerrado al escribir.** Escribir
una etiqueta nueva la incorpora al conjunto, y a partir de ese momento filtrar por ella funciona. Con
el hito pasa lo mismo: `biso set TASK-1 -m "v1.2"` es lo que hace que `v1.2` exista para
`biso ls -m "v1.2"`, y la última tarea que deja de llevarlo lo saca del conjunto.

Está la bandera `--unchecked` de `biso ls` y `biso export`, que apaga **las tres comprobaciones contra
estos conjuntos, las de etiquetas, personas e hitos, y ninguna otra**: los vocabularios configurados
de `--status`, `--type`, `--priority` y `--project` siguen validando, y `--parent` sigue resolviendo
su referencia. La bandera no cambia ninguna otra cosa.

---

## 7. Cómo se resuelve una referencia a una tarea

Todos los comandos que reciben `<ref>` usan exactamente esta rutina. No hay variantes por comando.

### 7.1. La gramática

| Forma | Ejemplo | Interpretación |
|---|---|---|
| `PREFIX-<n>` | `TASK-11` | identificador, sin distinguir mayúsculas en el prefijo |
| `<n>` | `11` | identificador, con el prefijo del tablero |
| `#<n>` | `#11` | igual que el anterior |
| cualquier otra cosa | `"CRLF"` | consulta de texto |

Dos banderas fuerzan la interpretación, y valen en todos los comandos que aceptan una referencia:

- `--id` obliga a interpretar como identificador. Con un valor que no encaje en la gramática, error 2.
- `--match` obliga a interpretar como texto, y sirve para buscar una tarea que se llame "42".

### 7.2. La búsqueda por texto

**Hay un solo ámbito de búsqueda de texto en todo el programa**, y es el que usan tanto la resolución
de una referencia como el filtro `--search` de `biso ls` y `biso export`. Busca, sin distinguir
mayúsculas ni acentos, en:

el título, la descripción, el plan, las notas, el resumen final, el texto de los criterios de
aceptación, el texto de la definición de hecho, el cuerpo de los comentarios, el cuerpo de la
pregunta abierta y las etiquetas.

No busca en los identificadores, ni en las referencias, ni en la documentación, ni en los campos de
extensión.

**Alcanzar el cuerpo de la pregunta abierta tiene dos consecuencias, y ambas se aceptan a
propósito.** La primera es que la resolución de una referencia por texto también llega ahí, así que
`biso get "CRLF"` puede resolver a una tarea porque ese texto está en su pregunta. La segunda es que
una pregunta puede crear una ambigüedad de código 5 donde antes no la había. Lo contrario sería peor:
que el texto de una pregunta solo se pudiera encontrar al dejar de estar abierta, cuando se convierte
en comentario, y no mientras espera respuesta.

Cuando se usa para resolver una referencia, y solo entonces, se aplican además estas reglas:

| Coincidencias | Qué pasa |
|---:|---|
| exactamente 1 | se usa esa tarea, con `note: "CRLF" matched TASK-11` por stderr |
| 0 | error 4 |
| más de 1 | error 5, con las candidatas por stdout en el formato de `biso ls` |

**Una coincidencia en el título gana sobre una coincidencia en cualquier otro sitio.** Si el texto
aparece en el título de una sola tarea, esa es la respuesta aunque aparezca en el cuerpo de otras
diez, y no hay ambigüedad. La búsqueda para resolver una referencia mira solo las tareas **no
archivadas**; el filtro `--search` mira las que digan los demás filtros.

### 7.3. Los tres mensajes de "no la encuentro"

**Identificador mal formado**, código 2, `code` igual a `malformed_id`:

```
error: malformed task id: "TASK-1.1"
hint: ids look like TASK-11 or 11. A subtask is an ordinary task with --parent TASK-1
```

**Identificador bien formado que el tablero nunca ha llegado a asignar**, código 4, `code` igual a
`never_allocated`:

```
error: TASK-999 has never existed on this board
note: the highest id ever assigned here is TASK-90
```

**Identificador que el tablero asignó alguna vez y que ahora no está**, código 4, `code` igual a
`not_found`:

```
error: TASK-53 is not on this board
note: TASK-53 was assigned at some point, so it was archived and then removed
hint: `biso ls --archived` lists what is archived
```

Los tres códigos son distintos: 2, y luego 4 con dos `code` distintos.

---

## 8. Las familias de banderas

Esta sección define de una vez la forma de todas las banderas de escritura. Los comandos no la
repiten: cada uno dice qué campos acepta, y esta sección dice qué forma tiene cada campo.

### 8.1. La regla

**El nombre desnudo añade. `set-` delante sustituye. `rm-` delante quita uno. `clear-` delante
vacía.** La forma de una bandera se deduce siempre del nombre del campo, sin nombres propios y sin
que haya que consultar nada.

Las cuatro variantes existen para **todo campo que guarde una lista de elementos**. Cada clase de
campo tiene estas variantes:

| Clase de campo | Variantes |
|---|---|
| Lista de elementos | las cuatro |
| Bloque de prosa | añadir, sustituir, vaciar |
| Mapa de claves | fijar una clave, quitar una clave, vaciar |
| Escalar | fijar, vaciar |
| Lista inmutable (comentarios) | solo añadir, `--comment` |

**`question` no entra en esta tabla.** Es un registro de tres partes (5.7), no una lista, ni un
bloque de prosa, ni un mapa, ni un escalar, así que ninguna de estas clases lo describe. **Ninguna
bandera de campo escribe `question`**: lo escriben `biso ask`, `biso answer` y la importación de
`biso new --from`, y nadie más, igual que `archived` solo lo cambia `biso archive` (5).

**El significado no cambia entre comandos.** `--ac` añade un criterio en `biso new`, en `biso set`, en
`biso start` y en `biso finish`, y todos los comandos de escritura aceptan todas estas banderas.

La regla tiene **dos desviaciones de nombre en todo el programa**, y las dos son de forma: ninguna
cambia lo que la bandera hace, pero en las dos el nombre no se deduce entero del campo. La primera es
el número gramatical de las notas: el añadido se llama `--note`, en singular, y su sustitución se llama
`--set-notes`, en plural. La segunda es el sufijo de las marcas: marcar un criterio de aceptación es
`--check` y marcar un elemento de la definición de hecho es `--check-dod`, o sea que el nombre desnudo
está reservado para los criterios y solo la otra lista lleva el sufijo del campo, aunque las demás
banderas de las dos lo lleven siempre (`--ac` y `--dod`, `--rm-ac` y `--rm-dod`). Lo mismo vale para
`--uncheck` frente a `--uncheck-dod`.

### 8.2. Campos de lista

| Campo | Añade | Sustituye | Quita | Vacía | Acepta lista por comas |
|---|---|---|---|---|---|
| etiquetas | `-l, --label` | `--set-label` | `--rm-label` | `--clear-label` | sí |
| personas asignadas | `-a, --assignee` | `--set-assignee` | `--rm-assignee` | `--clear-assignee` | sí |
| referencias | `--ref` | `--set-ref` | `--rm-ref` | `--clear-ref` | sí |
| documentación | `--doc` | `--set-doc` | `--rm-doc` | `--clear-doc` | sí |
| dependencias | `--dep` | `--set-dep` | `--rm-dep` | `--clear-dep` | sí |
| ficheros tocados | `--file` | `--set-file` | `--rm-file` | `--clear-file` | sí |
| criterios de aceptación | `--ac` | `--set-ac` | `--rm-ac` | `--clear-ac` | **no** |
| definición de hecho | `--dod` | `--set-dod` | `--rm-dod` | `--clear-dod` | **no** |

Todas las de "añade" y "sustituye" son repetibles. `--rm-ac` y `--rm-dod` toman un selector de la
sección 8.4.

**`--set-ac` y `--set-dod` crean elementos nuevos, con claves nuevas y sin marcar**, y las claves de
los elementos anteriores no se reutilizan. Es coherente con 5.1: la clave se asigna al crear el
elemento, y sustituir la lista crea elementos.

### 8.3. Campos de prosa

| Campo | Añade al final | Sustituye | Vacía |
|---|---|---|---|
| descripción | `-d, --desc` | `--set-desc` | `--clear-desc` |
| plan | `--plan` | `--set-plan` | `--clear-plan` |
| notas | `--note` | `--set-notes` | `--clear-notes` |
| resumen final | `--summary` | `--set-summary` | `--clear-summary` |

- Añadir a un campo vacío es lo mismo que fijarlo, así que al crear una tarea las dos columnas
  coinciden y no hay nada que decidir.
- Al añadir sobre contenido existente se intercala una línea en blanco, y cada repetición de la
  bandera en la misma invocación produce su propio párrafo.
- Añadir un valor vacío no hace nada y avisa, según 4.6.

### 8.4. Selectores de criterios

`--check`, `--uncheck`, `--rm-ac`, `--check-dod`, `--uncheck-dod` y `--rm-dod` toman un selector.
Todos son repetibles.

| Selector | Ejemplo | Qué elige |
|---|---|---|
| `all` | `--check all` | todos los elementos de esa lista en esa tarea |
| una clave | `--check 3` | el elemento `#3` |
| un rango de claves | `--check 1-4` | las claves de la 1 a la 4 que existan |
| varias claves | `--check 1,3,7` | esas tres |
| texto | `--check "cubre CRLF"` | el elemento cuyo texto contenga ese fragmento |

**La regla de desambiguación, que hay que implementar tal cual.** El valor se trata como lista de
claves **solo si el valor entero** encaja con `^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$`. En cualquier otro
caso es un texto literal, comas incluidas. Así, `--check "1, 2 y el ultimo"` es una búsqueda de texto
que no encontrará nada y dará error 4, en vez de convertirse en algo a medias.

| Caso límite | Resultado |
|---|---|
| clave que no existe | error 4: `no acceptance criterion #7 on TASK-11 (keys: 1, 3)` |
| texto que no encaja con ninguno | error 4, con los textos de los elementos listados |
| texto que encaja con dos | error 5, con los dos listados |
| rango donde faltan claves intermedias | se aplican las que hay, sin aviso |
| rango invertido, `4-1` | error 2 |
| marcar un elemento ya marcado | se queda marcado, sin aviso, la operación es idempotente |
| `--check all` en una tarea sin criterios | sin efecto, con `warning: TASK-11 has no acceptance criteria` |
| `--check all` sobre varias tareas | válido, cada tarea marca los suyos |
| una clave, un rango, una lista o un texto sobre varias tareas | error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |

**La regla de solape se aplica sobre el conjunto ya resuelto, no sobre el texto del selector.** Si
después de resolver `--check` y `--uncheck` un mismo elemento aparece en los dos conjuntos, es error
2, y da igual que se haya escrito `--check 3 --uncheck 3` o `--check all --uncheck 3`:

```
error: --check and --uncheck both select acceptance criterion #3 of TASK-11
```

### 8.5. Campos escalares

| Campo | Fija | Vacía |
|---|---|---|
| título | `-t, --title` | no se puede, es obligatorio |
| estado | `-s, --status` | no se puede, es obligatorio |
| tipo | `--type` | `--clear-type` |
| prioridad | `--priority` | `--clear-priority` |
| proyecto | `--project` | `--clear-project` |
| hito | `-m, --milestone` | `--clear-milestone` |
| tarea padre | `-p, --parent` | `--clear-parent` |
| fecha límite | `--due` | `--clear-due` |
| orden manual | `--ordinal` | `--clear-ordinal` |
| persona que reporta | `--reporter` | `--clear-reporter` |

Un escalar **nunca** se borra pasándole la cadena vacía, según 4.6.

### 8.6. Campos externos

| Operación | Bandera | Repetible |
|---|---|---|
| fijar una clave | `--ext <clave>=<valor>` | sí |
| quitar una clave | `--rm-ext <clave>` | sí |
| vaciar el mapa entero | `--clear-ext` | no |

**No existe `--set-ext`.** Fijar una clave con `--ext` ya sustituye su valor. Vaciar el mapa entero es
`--clear-ext`, y es la única forma de vaciarlo.

---

## 9. `biso prime`, el arranque de una sesión

### 9.1. Qué resuelve este comando

`biso` **no escribe nunca fuera del tablero, salvo el puntero del proyecto de la sección 3.2.** No
modifica ningún otro fichero del proyecto, ni al crear el tablero ni nunca. Todo lo que hace falta
para empezar a trabajar cabe en un solo comando, cuya salida es un solo mensaje.

`biso prime` se ejecuta al empezar la sesión, de la manera que convenga a quien lo use: como una
orden del asistente, desde un fichero de memoria del proyecto que diga en una línea "ejecuta
`biso prime` antes de tocar tareas", o a mano.

**El criterio de diseño del mensaje es exigente y está pensado para poderse comprobar:** quien lo lea
y no haya visto nunca la herramienta tiene que poder completar un ciclo de trabajo entero, desde
crear una tarea hasta cerrarla, sin leer nada más y sin ninguna interacción adicional.

### 9.2. Firma

```
biso prime [--full] [--limit <n>] [--json]
```

### 9.3. Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--full` | | no | booleano | falso | no | no | `--json` |
| `--limit <n>` | | no | entero >= 0 | 5 | no | no | ninguno |

- `--limit` acota juntas las secciones `ASSIGNED TO YOU` y `NEXT UP`: su valor son filas repartidas
  entre las dos, en ese orden de preferencia, con una sola línea de recuento al final de la última que
  se imprima. Con `0`, las dos desaparecen y se queda solo esa línea.
- `--full` añade al final la lista completa de banderas de `biso new` y `biso set`. Es para una
  persona que está aprendiendo la herramienta, no para el arranque de un agente.
- `--json` es la bandera global de la sección 3, y aquí es lo único que la restringe: no se puede
  combinar con `--full`, porque el JSON no lleva texto de ayuda.

### 9.4. Qué hace, caso a caso

| Situación | Qué pasa |
|---|---|
| Hay tablero y tiene tareas | Imprime el mensaje de 9.7 por stdout, código 0 |
| Hay tablero y está vacío | Igual, con los cuatro bloques de tareas sustituidos por las tres líneas de 9.8 |
| No hay tablero | Código 8, y por stderr el mensaje de la sección 3.2 |
| Alguna tarea no se puede leer | El mensaje sale igual, con el aviso de 4.12, código 0 |
| `--limit` negativo | Código 2 |

`biso prime` **no escribe nada, nunca**, y no necesita acceso exclusivo. Es seguro llamarlo en
paralelo desde varias sesiones y mientras otro proceso escribe.

### 9.5. El presupuesto de tamaño

El mensaje tiene un **tope duro de 5.120 bytes**, que se comprueba en la suite de pruebas y se reparte
en dos partes que suman exactamente ese tope:

- **La parte fija no pasa de 3.456 bytes.** Es la línea de título, `COMMANDS`, `FIELD FLAGS`, `RULES` y
  el párrafo final ("Pick one, ..."): nada de esto depende del contenido del tablero.
- **El resumen del tablero no pasa de 1.664 bytes.** Es el bloque `BOARD` (nombre, recuento por
  estado, vocabularios, identidad), `IN PROGRESS`, `NEEDS ANSWER`, `ASSIGNED TO YOU`, `NEXT UP`
  y las líneas de recuento: todo lo que cambia según qué haya en el tablero.

Los bloques no son contiguos entre sí, así que hay líneas en blanco de separación entre ellos: **cada
línea en blanco se cuenta en la parte a la que pertenece el bloque que la precede.** Con esta regla,
la línea en blanco que sigue al título es parte fija, la que sigue a `BOARD` es resumen, las que
siguen a `COMMANDS`, `FIELD FLAGS` y `RULES` son parte fija, y las que siguen a `IN PROGRESS`,
`NEEDS ANSWER`, `ASSIGNED TO YOU` y a `NEXT UP` son resumen.

Si el resumen no cupiera en su parte, el orden de recorte es completo y no deja ningún caso sin
definir:

1. Se reduce primero el número de filas de `NEXT UP`.
2. Si no basta, el de `ASSIGNED TO YOU`.
3. Si no basta, el de `NEEDS ANSWER`.
4. Si no basta, el de `IN PROGRESS`.
5. Si aun así no cupiera, cada uno de los cuatro bloques se reduce a su sola línea de recuento.

`ASSIGNED TO YOU` y `NEXT UP` comparten la línea de recuento que ya define 9.7. `IN PROGRESS` y
`NEEDS ANSWER` llevan cada uno la suya, con el mismo patrón: cuántas tareas del bloque quedan
fuera por el recorte y el comando para verlas completas. Para `IN PROGRESS` es
`N more not shown: 'biso ls --active'`, y para `NEEDS ANSWER` es
`N more not shown: 'biso ls --waiting'`.

Con esa lista el tope deja de ser una aspiración y pasa a ser alcanzable siempre.

El texto literal de la sección 9.7 ocupa **4.818 bytes** con el tablero del ejemplo: **3.327** de
parte fija y **1.491** de resumen. Las dos partes caben dentro de su tope.

**El número que congela el contrato de estabilidad de la sección 13 es el total, 5.120 bytes**, porque
es el único que quien llama observa. El reparto entre las dos partes puede cambiar sin romper ese
contrato.

### 9.6. Qué entra en el mensaje y qué se relega a `--help`

El criterio es uno solo: **entra lo que no se puede adivinar y hace falta para la primera acción; se
queda fuera lo que se puede consultar en el momento exacto en que hace falta.**

Entra:

- Las diez órdenes del ciclo de trabajo con su forma de uso. Quien no sabe que existe `biso finish`
  no va a escribir `biso finish --help`.
- **Los nombres de todas las banderas de campo**, en una rejilla de cinco líneas.
- El vocabulario real de este tablero, con **el recuento por estado** y con la marca de cuál es el
  estado de las tareas nuevas, cuál el activo y cuál el terminal.
- Las once reglas que no son adivinables.
- Los códigos de salida, en dos líneas.
- El estado del tablero: lo que está en curso y lo más urgente de lo que no ha empezado.

Se queda fuera, y va a `biso <cmd> --help`:

- Los valores, los tipos y las incompatibilidades de cada bandera. El mensaje da los nombres, que es
  lo que no se puede adivinar; la ayuda da el detalle, que es lo que se consulta cuando se necesita.
- El formato de lote de `biso new --from` y el esquema JSON completo.
- `biso export`, `biso config`, `biso doctor`, `biso archive`, `biso where` y `biso board`, que no
  aparecen en el ciclo de trabajo normal.
- Todos los casos límite: el rango invertido, las dos entradas estándar, la coma dentro de una
  etiqueta.
- La política de cuándo merece la pena crear una tarea, que es una decisión del proyecto y no de la
  herramienta. El mensaje la resume en una línea y no la desarrolla.

### 9.7. La salida literal

Esto es exactamente lo que `biso prime` imprime por stdout con un tablero de ejemplo. No imprime nada
por stderr.

**Ese tablero fija `task_prefix` a `TASK` explícitamente**, en vez de dejar que se derive de
`project_name` como haría por defecto (sección 4.11), para que los identificadores de todos los ejemplos
de este documento no dependan del nombre que le toque al tablero de turno. De paso queda demostrado que
`task_prefix` se puede fijar a mano.

```
biso 1.0.0 - the task board of this project. This message is all you need to start.

BOARD  Kex
  To Do 54 | In Progress 4 | Done 190
  new tasks start in To Do; `biso start` moves to In Progress; `biso finish` to Done
  types       idea, memory, task, bug, docs
  priorities  high, medium, low
  you are     @claude

COMMANDS  (`biso <cmd> --help` for the detail of any flag)
  biso ls [-s STATUS] [--type T] [-l LABEL] [--mine] [--search TEXT]
  biso get <ref> [--section ac]
  biso new "TITLE" [-d TEXT] [--ac TEXT]... [--type T] [--priority P]
  biso start <ref>... [--plan TEXT]
  biso note <ref> "TEXT"
  biso ask <ref> "QUESTION"
  biso answer <ref> "TEXT"
  biso finish <ref>... [--summary "TEXT"] [--check all] [--check-dod all]
  biso set <ref>... [any field flag]
  biso comment <ref> "TEXT" [--comment-author @who]

FIELD FLAGS  (same names, same meaning, in every command above that writes)
  -t --title  -s --status  --type   --priority  --project      -a --assignee
  -l --label  -d --desc    --ac     --dod       --plan         --note
  --summary   --dep        --ref    --doc       --file         -m --milestone
  -p --parent --due        --ordinal --ext K=V  --reporter     --comment
  --check     --uncheck             --check-dod --uncheck-dod

RULES  (none of these are guessable; they are the whole learning curve)
  1. Every write goes through biso. Nothing else touches the board.
  2. A bare field flag ADDS. Replacing and removing are explicit: --label X
     adds, --set-label X replaces the list, --rm-label X drops one, and
     --clear-label empties it. Same four shapes for every list field.
  3. <ref> is an id (TASK-12), a bare number (12) or free text ("CRLF"). Text
     matching several tasks is an error that lists them, never a guess. `note`,
     `comment`, `ask` and `answer` take one <ref>; `set`, `start` and `finish`
     take several.
  4. Filters reject values this board does not have: `-s Pending` is an error,
     not an empty list. Case, spaces, hyphens and underscores are ignored, so
     `-s todo`, `-s "To Do"` and `-s TO_DO` are one and the same filter. An
     empty list is therefore a fact about the board that you can act on.
  5. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says
     on stderr what it left out. --all lifts the limit, --any-status includes
     Done, --archived reaches the archive.
  6. --check and --uncheck take all, 3, 1-4, 1,3,7 or the criterion text. The
     numbers are the stable #N keys that `biso get` shows, and they never
     shift when one criterion is removed.
  7. `biso new` prints the new id and nothing else. Every other write prints one
     line per task: id, status, criteria, urgency. Add --print for the whole
     record, or --json for a versioned envelope.
  8. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.
  9. Long text: a real newline works, and so do -d @file.md and -d - for stdin.
 10. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,
     6 precondition not met, 7 environment, 8 no board here, 9 nothing written.
 11. `biso ask <ref> "..."` parks a task on a question and `biso answer` unparks
     it, writing both into the comments. Ask instead of guessing. A task
     assigned to you is one a person decided you should do.

IN PROGRESS
  TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
  TASK-52  In Progress  task  low     Document the release checklist                    ac 0/1  -        -
    lease expired 2026-09-05T09:00:00Z, was held by @bob
  TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -

NEEDS ANSWER
  TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
    Should the retry budget be shared with the download endpoint or kept separate?

ASSIGNED TO YOU
  TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
  TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -

NEXT UP  (not assigned to you, by urgency)
  TASK-7   To Do        bug   high    Crash on an empty repository                      ac 0/4  -        2026-09-08
  TASK-19  To Do        task  high    Retry the upload on 5xx                           ac 0/2  -        -
  TASK-44  To Do        bug   low     Wrong column width on narrow ttys                 ac 0/1  -        -
  49 more not shown: `biso ls --not-active --not-waiting`

Pick one, `biso start <ref> --plan "..."`, work, `biso note <ref> "..."` as you go,
and close with `biso finish <ref> --check all --check-dod all --summary "..."`.
That is the loop. Create a task when the work needs planning or review; do small
edits directly.
```

Cómo se calcula el resumen, para que la implementación sea única:

- La línea de recuento del bloque `BOARD` tiene **un número por cada estado configurado**, en el
  orden en que están configurados, y cuenta las tareas no archivadas de ese estado. No hay ninguna
  categoría inventada como "abiertas" que no se corresponda con un estado del tablero.
- Los cuatro bloques `IN PROGRESS`, `NEEDS ANSWER`, `ASSIGNED TO YOU` y `NEXT UP` se reparten
  el tablero por esta precedencia, y cada tarea cae en el primero que la acepte:
  1. `NEEDS ANSWER`, si tiene una pregunta abierta.
  2. `IN PROGRESS`, si está en el estado activo.
  3. `ASSIGNED TO YOU`, si está asignada a la identidad configurada.
  4. `NEXT UP`, el resto.
  **Ninguna tarea aparece en dos bloques.** Una tarea aparcada no sale en `IN PROGRESS` aunque esté en
  el estado activo, porque ese bloque significa que alguien está trabajando y ahí no lo está nadie.
  Los cuatro excluyen las tareas terminadas y las archivadas.
- Se imprimen en este orden: `IN PROGRESS`, `NEEDS ANSWER`, `ASSIGNED TO YOU` y `NEXT UP`.
  **Un bloque sin filas no se imprime**, ni siquiera su encabezado. **Sin identidad configurada,
  `ASSIGNED TO YOU` no se imprime nunca**, aunque el resto del mensaje se imprime igual, con
  `you are (not set)` en el bloque `BOARD`.
- `IN PROGRESS` lista las tareas del estado activo sin pregunta abierta, ordenadas por la regla de
  orden de 10.4, sin límite. Cada tarea cuyo arrendamiento está vencido (el campo derivado
  `leaseExpired` de la sección 5) lleva, igual que `NEEDS ANSWER` con su pregunta, una segunda línea
  indentada con la forma `lease expired <leaseExpiresAt>, was held by <leaseHolder>`. Es el único de
  los cuatro bloques que la lleva, porque es el único cuya etiqueta afirma que alguien está
  trabajando ahora mismo, y un arrendamiento vencido contradice justo esa afirmación. Esta línea, como
  la de la pregunta, no cuenta para el ancho de las columnas. El hecho que la provoca sí viaja en el
  esquema JSON de 9.9, como el campo `leaseExpired`, y sus dos detalles no: quien los quiera los pide
  con `biso get`, que los imprime en su línea `lease` (10.5), igual que pide el cuerpo de la pregunta.
- `NEEDS ANSWER` lista las tareas con pregunta abierta, ordenadas igual, sin límite. Cada tarea
  ocupa **dos líneas**: la fila de siempre, con las ocho columnas del algoritmo de `biso ls`, y debajo
  una línea indentada con la pregunta recortada a **100 celdas**, con la misma regla exacta que el
  algoritmo aplica a los títulos. Los saltos de línea reales del cuerpo se sustituyen por un espacio
  antes de recortar, y el recorte usa el mismo sufijo `...` que 10.4 usa para los títulos. Va en línea
  propia y no en una novena columna, porque el algoritmo tiene ocho exactas y una regla que dice que
  la octava nunca se rellena.
- `ASSIGNED TO YOU` lista las tareas asignadas a la identidad configurada que no estén ya en
  `NEEDS ANSWER` ni en `IN PROGRESS`, ordenadas igual. `NEXT UP` lista el resto, ordenadas
  igual.
- `ASSIGNED TO YOU` y `NEXT UP` **comparten el límite de `--limit`**: su valor son filas repartidas
  entre las dos, en ese orden de preferencia, con una sola línea de recuento al final de la última que
  se imprima. Con `--limit 0` desaparecen los dos y queda solo esa línea. Si cada bloque tuviera su
  propio límite, el resumen crecería al doble sin que `--limit` lo notara.
- La línea de recuento dice cuántas tareas quedan fuera de `ASSIGNED TO YOU` y `NEXT UP` juntas por el
  corte.
- Las filas usan exactamente el algoritmo de columnas de `biso ls` de la sección 10.4, con una
  diferencia declarada aquí: el ancho de las columnas 1 a 7 se calcula sobre las filas de los cuatro
  bloques juntas, **sin contar las líneas de pregunta ni las de arrendamiento vencido**, que no son
  filas de la tabla, para que los cuatro bloques se lean como una sola tabla.
- `NEXT UP` es lo que no cae en ninguno de los tres bloques anteriores, no "lo que no ha empezado".
  Por eso su rótulo es `NEXT UP  (not assigned to you, by urgency)`, su línea de recuento tiene la
  forma `N more not shown: 'biso ls --not-active --not-waiting'`, y su clave en el esquema JSON del
  apartado 9.9 es `hiddenCount`.

### 9.8. Tablero vacío

Cuando no hay ninguna tarea, los cuatro bloques de tareas de 9.7 se sustituyen por esto, sin imprimir
ninguno de sus encabezados, y el resto del mensaje no cambia:

```
THE BOARD IS EMPTY
  Create the first one:
  biso new "Title" -d "What and why" --ac "How we will know it works"
```

### 9.9. El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "prime",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "tool": { "name": "biso", "version": "1.0.0" },
    "board": {
      "name": "Kex",
      "me": "@claude",
      "statuses": ["To Do", "In Progress", "Done"],
      "initialStatus": "To Do",
      "activeStatus": "In Progress",
      "terminalStatus": "Done",
      "types": ["idea", "memory", "task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "extensions": ["trello.card"],
      "countByStatus": { "To Do": 54, "In Progress": 4, "Done": 190 }
    },
    "inProgress": [
      { "id": "TASK-11", "title": "Normalize CRLF in the diff", "status": "In Progress",
        "type": "bug", "priority": "high", "assignees": ["@claude"], "due": null,
        "acDone": 1, "acTotal": 2, "urgency": 19.0, "leaseExpired": false }
    ],
    "needsAnswer": [
      { "id": "TASK-60", "title": "Confirm the retry budget for the upload endpoint",
        "status": "In Progress", "type": "task", "priority": "high", "assignees": ["@claude"],
        "due": null, "acDone": 0, "acTotal": 2, "urgency": 15.2, "leaseExpired": false }
    ],
    "assignedToYou": [
      { "id": "TASK-61", "title": "Rewrite the install section", "status": "To Do",
        "type": "docs", "priority": "medium", "assignees": ["@claude"], "due": null,
        "acDone": 0, "acTotal": 1, "urgency": 12.4, "leaseExpired": false }
    ],
    "nextUp": [
      { "id": "TASK-7", "title": "Crash on an empty repository", "status": "To Do",
        "type": "bug", "priority": "high", "assignees": [], "due": "2026-09-08",
        "acDone": 0, "acTotal": 4, "urgency": 18.2, "leaseExpired": false }
    ],
    "hiddenCount": 49
  }
}
```

Las reglas y los nombres de las banderas no viajan en el JSON: quien pide JSON es un programa, y un
programa no necesita que le expliquen que el nombre desnudo añade.

**Ninguna de las cuatro listas trae el cuerpo de la pregunta**, por el mismo motivo que el esquema de
`task.list` en 10.4 no trae el cuerpo de la tarea: es texto largo. Lo que sí llevan es la posición de
cada tarea (en `needsAnswer` o en cualquier otro de los cuatro bloques), que ya dice si está
aparcada, igual que el campo derivado `waiting` de `task.list`. Quien necesite leer la pregunta usa
`biso get --section question`.

**`leaseExpired` sale en los cuatro bloques**, y en `assignedToYou` y en `nextUp` vale siempre `false`.
Esos dos bloques no pueden contener ninguna tarea en el estado activo, que es la única clase de tarea
en la que un arrendamiento puede existir: la precedencia de 9.7 manda toda tarea activa a `inProgress`
o a `needsAnswer`. Quien implemente puede apoyarse en ese valor constante, pero la clave se emite igual,
porque la regla de 12.4 prohíbe que una clave aparezca o desaparezca según los datos y quien lee esta
salida no sabe de antemano en qué bloque va a caer una tarea. Va en el JSON aunque la segunda línea
indentada del texto salga solo en `inProgress`, porque no es texto largo y esconderlo obligaría a quien
consume JSON a llamar a `biso get` tarea por tarea para saber algo que el mensaje de texto ya enseña.
Sus dos detalles, `leaseExpiresAt` y `leaseHolder`, no salen aquí: para eso está `task.list` (10.4), y
en texto la línea `lease` de la ficha de `biso get` (10.5).

### 9.10. Códigos de salida

| Desenlace | Código |
|---|---:|
| Mensaje impreso | 0 |
| Mensaje impreso con alguna tarea ilegible | 0, con aviso |
| `--limit` negativo, o `--full` junto con `--json` | 2 |
| No hay tablero | 8 |

### 9.11. `biso prime --help`

```
Usage: biso prime [options]

Print everything needed to start working on this board: the commands, the field
flags, the rules that are not guessable, the board vocabulary and what is in
flight. Run it once at the start of a session. It writes nothing.

Options:
  --full          also list every flag of `biso new` and `biso set` in detail
  --limit <n>     rows shown across ASSIGNED TO YOU and NEXT UP together
                  (default 5, 0 hides both)
  --json          machine-readable envelope instead of the message
  -h, --help      show this help

Exit codes:
  0  message printed
  2  bad usage
  8  no board here

Examples:
  biso prime
  biso prime --limit 10
  biso -C ~/work/kex prime
```

---

## 10. Los comandos

Veinte comandos. Los once primeros son el ciclo de trabajo y aparecen en `biso --help`; los nueve
restantes son de administración y aparecen en `biso help all`. De esos once, diez son los que 9.6
cuenta como "las órdenes del ciclo de trabajo" del bloque `COMMANDS` de `biso prime`: `prime` es el
undécimo, y no se lista a sí mismo en su propio mensaje.

| Comando | Qué hace | En `biso --help` |
|---|---|---|
| `prime` | El mensaje de arranque de la sección 9 | sí |
| `ls` | Lista tareas | sí |
| `get` | Muestra una tarea | sí |
| `new` | Crea una o muchas tareas | sí |
| `set` | Cambia campos de una o varias tareas | sí |
| `start` | Toma una tarea y la pone en curso | sí |
| `note` | Añade una nota de implementación | sí |
| `comment` | Añade un comentario con autor | sí |
| `finish` | Cierra una tarea | sí |
| `ask` | Aparca una tarea en una pregunta abierta | sí |
| `answer` | Responde la pregunta abierta y desaparca la tarea | sí |
| `archive` | Saca una tarea del tablero activo | no |
| `export` | Vuelca el tablero en el formato de entrada de `new --from` | no |
| `init` | Crea un tablero | no |
| `where` | Explica qué tablero se está usando y por qué | no |
| `config` | Lee y cambia la configuración | no |
| `doctor` | Comprueba y repara la integridad | no |
| `board` | Abre la interfaz interactiva | no |
| `help` | La ayuda de primer nivel y la de cada comando | no |
| `snapshot` | Escribe la instantánea del tablero en su propio directorio y la guarda en el control de versiones | no |

**Las banderas globales de la sección 3 valen en todos ellos y no se repiten en las tablas de
parámetros de cada comando.** Un comando solo las menciona cuando le impone una restricción
adicional, y esas restricciones son exactamente cuatro en todo el documento: `prime --full` no se
combina con `--json`, `config` solo acepta `--json` en su subcomando `list`, `export` rechaza
`--json` con código 2, y `--print` y `--dry-run` no valen donde no tienen nada que hacer, cada una
en su propia lista, según la regla de la sección 3.

El caso de `export` merece una línea, porque es el único comando cuya salida ya es JSON sin pedirlo:
son objetos JSON, uno por línea, y `--json` pide el sobre único de la sección 12, que es otra forma
distinta. Pasarlo es error 2:

```
error: --json does not apply to export
       its output is already one JSON object per line
```

**Todos los comandos que escriben aceptan todas las banderas de campo de la sección 8**, con el mismo
nombre y el mismo significado. Eso vale para `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`,
`answer` y `archive`. Lo que distingue a unos de otros no es qué campos aceptan, sino qué hacen por
defecto. Las tablas de parámetros de cada comando enumeran solo lo que es propio de ese comando.

### 10.1. `biso init`

#### Firma

```
biso init [<name>] [--at <dir>] [--statuses <list>]
          [--initial-status <status>] [--active-status <status>]
          [--terminal-status <status>] [--types <list>] [--priorities <list>]
          [--projects <list>] [--extensions <list>] [--prefix <text>]
          [--overwrite-config] [--from <location>]
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<name>` | | no | texto | el nombre del proyecto | no | no | |
| `--at <dir>` | | no | ruta de un directorio | una carpeta nueva en la raíz por defecto de la máquina (sección 3.3) | no | no | |
| `--statuses <list>` | | no | lista | `To Do, In Progress, Done` | sí | sí | |
| `--initial-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--active-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--terminal-status <status>` | | sí, si hay `--statuses` | uno de `--statuses` | | no | no | requiere `--statuses` |
| `--types <list>` | | no | lista | `task, bug, docs` | sí | sí | |
| `--priorities <list>` | | no | lista | `high, medium, low` | sí | sí | |
| `--projects <list>` | | no | lista | vacía | sí | sí | |
| `--extensions <list>` | | no | lista | vacía | sí | sí | |
| `--prefix <text>` | | no | texto de solo letras | se deriva de `<name>` en mayúsculas (sección 4.11) | no | no | |
| `--overwrite-config` | | no | booleano | falso | no | no | |
| `--from <location>` | | no | ruta de un directorio | | no | no | `<name>`, `--statuses`, `--initial-status`, `--active-status`, `--terminal-status`, `--types`, `--priorities`, `--projects`, `--extensions`, `--prefix`, `--overwrite-config` |

**`--at` es la ruta del directorio del tablero que se va a crear, no el directorio donde se crea.** Con
`--at tablero` el tablero queda en `tablero`, no en `tablero/kex-3f9a2b1c`. Es la misma convención que
la clave `path` del puntero, que también nombra el directorio del tablero y no el que lo contiene
(sección 3.2), y no es casualidad: la ruta que recibe `--at` es exactamente la que se escribe en esa
clave, con la misma forma. Puede ser absoluta o relativa al directorio de trabajo, y **su último componente es el nombre de la carpeta, que
es decorativo** (sección 3.2), así que `--at tablero` es tan válido como `--at kex-3f9a2b1c`.

#### Comportamiento

Crea un tablero vacío con su configuración. **No escribe nunca fuera del tablero**, salvo el puntero
del proyecto que se describe a continuación.

**`<name>` es el `project_name` inicial del tablero.** Cambiarlo más adelante es cosa de `biso config
set project_name`, que no toca el sistema de ficheros (sección 10.10).

`init` escribe, además del tablero, **el puntero del proyecto** (el fichero `.biso.json` de la
sección 3.2), y lo escribe siempre que no exista ya uno, porque un tablero no se localiza nunca por su
posición en el disco sino por una de las dos vías de la sección 3.2. Es la única cosa que `init`
escribe fuera del tablero. La salida siempre confirma que el proyecto apunta al tablero, se haya escrito
el puntero en esta llamada o ya estuviera ahí de antes.

**Sin `--at`, el puntero no lleva clave `path`**: el tablero va a la raíz por defecto de la máquina y
ahí lo encuentra la búsqueda por marcador de la sección 3.2, así que escribir su ruta sería guardar un
dato que nadie necesita y que dejaría de valer al cambiar `boards_root`.

**Con `--at`, el puntero lleva la ruta en `path`, con la misma forma en que se dio `--at`**: relativa si
`--at` era relativa, absoluta si era absoluta. `--at tablero` escribe `"path": "tablero"`;
`--at /Users/avilches/Hub/Projects/Kex/tablero` escribe esa ruta completa. No hay bandera para elegir la
forma porque la forma de `--at` ya es la elección, y esa elección es de quien llama y no del programa,
porque depende de un dato que el programa no tiene: **dónde van a vivir las demás copias de trabajo del
proyecto.**

La consecuencia práctica hay que decirla, porque es lo único que distingue a las dos formas. Una `path`
relativa se resuelve contra el directorio del puntero y, si ahí no hay tablero, contra sus ancestros
(sección 3.2), así que vale en cualquier copia de trabajo que tenga el tablero dentro o por encima: es
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
ancestros de la sección 3.2 mantiene alcanzable desde una copia de trabajo que no tenga el directorio
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
forma perezosa la primera vez que corre ahí (10.14), y publicarlo es entonces un trabajo aparte.
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
datos: **eso no es un tablero**, y la resolución de la sección 3.2 no lo acepta como tal, sigue buscando
en los ancestros y en las raíces, y así encuentra el tablero de verdad. Si no lo encuentra en ninguna
parte, el error nombra ese directorio a medias, porque es la pista de lo que ha pasado.

**Si ya existe un puntero pero el tablero que nombra no está en esta máquina** (sección 3.2), `init`
no acuña un `id` nuevo: usa el que ya lleva el puntero, para que las dos máquinas sigan hablando del
mismo tablero. Y no reescribe el puntero, porque ya era correcto.

**Sin `--statuses`**, el tablero nace con `To Do, In Progress, Done`, con los papeles inicial, activo y
terminal en ese orden. **Con `--statuses`**, hacen falta las tres banderas de papel,
`--initial-status`, `--active-status` y `--terminal-status`, con los mismos nombres que las claves de
configuración a las que corresponden.

**`--from <location>` restaura una instantánea, en vez de crear un tablero en blanco.** `<location>`
es el directorio de un tablero que ha escrito `biso snapshot` (sección 10.14), es decir, el que
contiene `snapshot.ndjson` y `board.json`. En una sola invocación, `init --from` hace lo que sería
crear el tablero con la configuración de `board.json` e importar `snapshot.ndjson` con las mismas
reglas del lote de `biso new --from` (sección 10.3): valida el fichero de tareas entero contra el
vocabulario de `board.json` antes de escribir nada y, solo si todo es válido, escribe primero la
configuración y después las tareas. Como `board.json` ya trae el nombre del tablero, los estados,
los tipos, las prioridades, los proyectos, las extensiones y el prefijo del tablero de origen,
**`--from` es incompatible con `<name>` y con cualquier bandera de vocabulario**: no hay nada que
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
error: this project already has board 3f9a2b1c, at /Users/avilches/.biso/boards/kex-3f9a2b1c
hint: `biso where` says which rule picked it
hint: --overwrite-config rewrites its configuration and never touches its tasks
```

**`--dry-run` vale en este comando** (sección 3), y es donde más sirve: valida los argumentos y, con
`--from`, la instantánea entera contra el vocabulario que ella misma trae, sin crear ni escribir nada,
y sale 0 si habría funcionado y 9 si no. `--print`, en cambio, es error 2, porque ninguna tarea que
existiera antes queda afectada.

| Caso | Qué pasa |
|---|---|
| Ya hay un tablero accesible desde aquí | Error 2, salvo con `--overwrite-config`, que reescribe la configuración y **nunca toca las tareas** |
| El directorio de destino tiene una base de datos que no se puede leer (4.12) | No cuenta como tablero accesible, así que `--from` reconstruye ahí mismo adoptando el `id` del marcador, código 0. Es el remedio que el `hint` del error 10 nombra, y también lo que necesita un clon traído a otra máquina, que llega con la carpeta versionada y sin base de datos |
| El directorio de trabajo es ya el directorio de un tablero | Es el caso de la fila de arriba, alcanzado por la primera vía de 3.2, y se resuelve igual: Error 2, y con `--overwrite-config` se reescribe la configuración de ese tablero, que es exactamente lo que esa bandera significa. Un tablero no se crea nunca dentro de otro |
| Ya hay un puntero, pero el tablero que nombra no está en esta máquina | No es un error: se crea el tablero adoptando el `id` que el puntero ya lleva, y el puntero no se reescribe porque ya era correcto, código 0 |
| `--at` a un directorio que ya es el directorio de un tablero | Error 2, con el mismo motivo visto desde el otro lado: el destino ya es un tablero |
| `--at` con una ruta relativa | No es un error: el tablero se crea ahí y el puntero lleva esa misma ruta relativa, código 0 |
| `--at` con una ruta absoluta | No es un error: el tablero se crea ahí y el puntero lleva esa misma ruta absoluta, código 0 |
| `--at` con una ruta relativa que sale del proyecto, como `../tableros/kex` | No es un error, y el puntero la guarda tal cual: resuelve mientras la posición relativa entre el puntero y el tablero se mantenga, y el marcador confirma que el directorio al que llega es el tablero que el `id` nombra |
| `--overwrite-config` sobre un tablero con alguna tarea, si el prefijo resultante (el de `--prefix`, o el que se derive de `<name>` cuando no se da) no coincide con el `task_prefix` que el tablero ya tiene | Error 6, la misma inmutabilidad que la sección 10.10 aplica a `task_prefix` |
| Falta alguna de las tres banderas de papel, habiendo `--statuses` | Error 2, con las tres nombradas y cuáles faltan |
| Una bandera de papel sin `--statuses` | Error 2, diciendo que los papeles solo se fijan junto a la lista de estados |
| Una bandera de papel nombra un estado que no está en `--statuses` | Error 2, con el valor y la lista de estados |
| Dos banderas de papel nombran el mismo estado | Error 2, con los dos papeles y el estado que comparten |
| `--statuses` con menos de tres estados | Error 2, diciendo cuántos hacen falta y por qué |
| `--prefix` con algo que no sean letras | Error 2, `code` `invalid_prefix` |
| Sin `--prefix`, el nombre del tablero no deja ninguna letra al derivar el prefijo (sección 4.11) | Error 2, `code` `invalid_prefix`, pidiendo `--prefix` explícito |
| `--at` a un directorio donde no se puede escribir | Error 7 |
| `--from` junto con `<name>`, con cualquier bandera de vocabulario, o con `--overwrite-config` | Error 2 |
| `--from` a un directorio al que le falta `snapshot.ndjson`, `board.json`, o los dos (una instantánea a medias) | Error 4, `code` `file_not_found`, nombrando qué fichero falta |
| `--from` cuyo `board.json` no se puede interpretar como JSON, o lleva una clave desconocida | Error 2, `code` `invalid_snapshot_config` |
| `--from` cuyo `board.json` tiene el mismo problema que haría fallar con Error 2 a la bandera de vocabulario equivalente (por ejemplo, `statuses` con menos de tres elementos, o un `task_prefix` sin letras) | Error 2, con el mismo `code` que usaría esa bandera |
| `--from` cuyo `snapshot.ndjson` está vacío (una instantánea con configuración pero sin tareas) | No es un error: se crea el tablero con esa configuración y cero tareas, código 0 |
| `--from` cuyo `board.json` declara un vocabulario que ninguna tarea de `snapshot.ndjson` usa | No es un error: el tablero se crea con ese vocabulario tal cual lo declara `board.json`, tenga tareas que lo usen entero o no |
| `--from` cuyo `board.json` trae `me` o `default_limit`, porque se escribió a mano o con una versión anterior | No es un error, y tampoco se importan: `biso snapshot` no las escribe (10.14) y `init --from` no las lee, con `note: me and default_limit are not restored, they belong to whoever uses the board`. No son claves desconocidas, así que no caen en el error 2 de la fila de arriba |
| `--from` cuyas tareas usan un valor, una clave de extensión o un `id` que `board.json` no hace válido | Error 9, la misma regla del lote de `biso new --from` (sección 10.3), con el detalle de qué falta línea a línea |

**Los tres estados especiales se guardan como valores explícitos en la configuración, no como
posiciones.** Cambiar `statuses` después no los mueve nunca. Si al cambiar `statuses` uno de los tres
deja de existir, el comando que lo hace falla, según la sección 10.10.

**El fichero de la base de datos se llama `board.db`**, con `board.db-wal` y `board.db-shm` como sus
ficheros auxiliares. El nombre es fijo y forma parte de la interfaz, no un detalle interno, porque es
lo que hace reconocible un directorio de tablero: la primera vía de la sección 3.2 se apoya en él, y
sin un nombre declarado esa vía no sería implementable de una sola manera.

**Junto a él, `init` escribe el marcador de identidad `<id>.id`**, por ejemplo `3f9a2b1c.id`, que es lo
que permite encontrar el tablero por su identificador leyendo nombres de un directorio, sin abrir
ninguna base de datos (sección 3.2). Su nombre es el dato; su contenido es
`{ "storeVersion": 1 }`, la versión del formato del almacén, que no repite el identificador para que no
haya dos sitios donde pueda decir cosas distintas. El mismo identificador está guardado dentro de la
base de datos, y esa redundancia es a propósito: es la que permite comprobar que el marcador de un
directorio corresponde de verdad al tablero que contiene, cosa que `biso doctor` hace (10.13). Los dos
ficheros, la base de datos y el marcador, son lo único que hace falta para que un directorio sea un
tablero.

**El directorio del tablero es también, si es posible, su propio repositorio, pero `init` no lo crea.**
Lo que `init` sí escribe es el **fichero de exclusión del sistema de control de versiones que la máquina tenga
configurado** (sección 3.3), con `board.db` y sus dos ficheros auxiliares dentro, dejándolo listo para
el día en que el directorio llegue a ser un repositorio: lo que se versiona entonces es
`snapshot.ndjson`, `board.json` y el marcador, nunca el binario. Con `git`, ese fichero es
`.gitignore`; con `custom`, el que declare la clave `ignore_file`, y ninguno si no la declara; con
`none`, ninguno, porque no hay nada de lo que excluirse. Escribir un fichero de texto no es ejecutar
ningún programa, así que esto no contradice que `biso snapshot` sea el único que lo hace.

**Si alguien cambia la clave `vcs` después, ese fichero se queda con el nombre del sistema anterior**, y
`biso` no lo renombra ni escribe otro: `init` solo se ejecuta una vez por tablero, y adivinar cuándo hay
que reescribir un fichero de exclusión ajeno sería pasarse. Quien cambie de sistema tiene que escribir a
mano el fichero que el nuevo espere, con las tres líneas de la base de datos.

**El marcador no se excluye, y eso resuelve algo.** Al quedar versionado con la instantánea, el
identificador del tablero viaja en ella, así que restaurar con `biso init --from <instantánea>` puede
recuperar la identidad y no solo los datos: el tablero restaurado adopta el `id` que el marcador de la
instantánea nombra, en vez de acuñar uno nuevo, y por eso el puntero commiteado del proyecto sigue
valiendo después de restaurar. Sin eso, restaurar en una máquina nueva daba un tablero correcto que el
proyecto no podía encontrar. Si el `id` de la instantánea ya existe en esta máquina, es el error de
identidad duplicada de la sección 3.2, no una adopción silenciosa.

**Es `biso snapshot`, no `init`, quien convierte el directorio en un repositorio, y lo hace de forma
perezosa**: la primera vez que `snapshot` corre sobre un directorio que no está en ninguno, crea ahí el
del sistema de control de versiones configurado antes de guardar la revisión (sección 10.14). Y si el
directorio ya está dentro del repositorio del proyecto, que es el caso de un tablero versionado con el
código, no crea ninguno: la revisión va a ese. La base de datos no se versiona nunca, ni siquiera
después de eso, porque cada escritura suya reescribe páginas internas: cada revisión guardaría una copia
completa y ningún sistema podría diferenciarla de una forma legible.

**El repositorio es opcional y su ausencia no rompe nada.** Si el sistema configurado no está instalado,
`snapshot` sigue escribiendo sus dos ficheros igual y sirviendo para restaurar con `--from`; lo único
que se pierde es el historial. `biso` no puede exigir que haya un sistema de control de versiones instalado, así
que esto nunca hace fallar ni a `init` ni a `snapshot` por esta sola razón (la sección 10.14 sí
distingue un fallo de entorno una vez que la revisión se intenta de verdad, con un repositorio ya
existente). Quien quiera dar historial a un tablero que nació sin él puede crear el repositorio a mano
en su directorio en cualquier momento: la siguiente instantánea lo detecta y empieza a guardar
revisiones.

#### Salida

Esto es lo que imprime la tercera invocación de los ejemplos de ayuda,
`biso init Kex --at kex-board --prefix TASK --extensions trello.card` (con `--json` para el esquema
de más abajo). El prefijo sale `TASK` porque lo fija `--prefix`, no porque se derive del nombre
`Kex`, que sin esa bandera daría `KEX` (sección 4.11).

```
Created board "Kex"
  statuses    To Do (initial) | In Progress (active) | Done (terminal)
  types       task, bug, docs
  priorities  high, medium, low
  prefix      TASK
This project now points at that board.
Run `biso prime` to see how to use it.
```

La línea "This project now points at that board." aparece siempre, porque el proyecto siempre queda
apuntando a ese tablero, se escriba el puntero en esta llamada o ya estuviera escrito de antes
(sección 3.2).

**Esta invocación emite además las dos notas**, porque `--at kex-board` es una ruta relativa que cae
dentro del proyecto, que es justo el caso que las dispara. Por stderr sale esto, en este orden:

```
note: the board lives inside this project. Ignore kex-board/ and the board keeps
      its own history; version it and the snapshot travels with your code.
      Either way the database stays out, kex-board/.gitignore excludes it
note: the location is stored as the relative path "kex-board". A working copy
      outside this project will not have that folder while git ignores it, so
      it will not find the board: use an absolute --at if you work that way
```

No están en el bloque de arriba porque ese bloque es stdout, y las notas van por stderr como todas
(sección 4.3). El puntero que esta llamada escribe es
`{ "version": 1, "id": "3f9a2b1c", "path": "kex-board" }`.

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "init",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "board": {
      "name": "Kex",
      "statuses": ["To Do", "In Progress", "Done"],
      "initialStatus": "To Do", "activeStatus": "In Progress", "terminalStatus": "Done",
      "types": ["task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "taskPrefix": "TASK"
    },
    "pointerCreated": true
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tablero creado, o restaurado con `--from` | 0 |
| Ya existía y no hay `--overwrite-config` | 2 |
| Argumentos inválidos, incluido un `board.json` de `--from` inválido, o `--from` junto con `--overwrite-config` | 2 |
| `--overwrite-config` cambiaría `task_prefix` con tareas ya creadas | 6 |
| No se puede escribir | 7 |
| `--from` a un directorio sin `snapshot.ndjson`, sin `board.json`, o sin los dos | 4 |
| El lote de `snapshot.ndjson` de `--from` falla su validación | 9 |

#### `biso init --help`

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
  7  cannot write there
  9  --from's snapshot.ndjson failed batch validation

Examples:
  biso init
  biso init Kex --statuses "Ideas,To Do,In Progress,Done" \
      --initial-status Ideas --active-status "In Progress" \
      --terminal-status Done
  biso init Kex --at kex-board --prefix TASK --extensions trello.card
  biso init --at /tmp/tablero-nuevo --from ~/.biso/boards/kex-3f9a2b1c
```

El segundo ejemplo deja `To Do` sin ningún papel a propósito: un tablero puede llevar estados que no
son ni el inicial, ni el activo, ni el terminal, y eso no rompe nada.

---

### 10.2. `biso where`

#### Firma

```
biso where [--json]
```

Sin parámetros propios.

#### Comportamiento

Dice el identificador del tablero, su nombre, la ruta de su directorio, y cuál de las dos vías de la
sección 3.2 lo ha elegido. Los tres son datos distintos, y merece la pena verlos juntos porque cada uno
cambia por su cuenta: `biso config set project_name` cambia el nombre y no toca la ruta (10.10), mover el
directorio a mano cambia la ruta y no toca el nombre, y el identificador no cambia jamás. Es el comando al
que remite el error de código 8, y el que hace visible una resolución que de otro modo sería invisible.

| Caso | Qué pasa |
|---|---|
| Hay tablero | Lo imprime con su identificador, su nombre, su ruta y la vía que lo eligió, código 0 |
| El directorio de trabajo es el propio directorio del tablero | Lo imprime igual, con la primera vía de 3.2 en `source` y `path` apuntando al directorio de trabajo, código 0 |
| No hay tablero configurado | Imprime lo que ha buscado y dónde, código 8, `code` `no_board` |
| El puntero nombra un tablero que no está en esta máquina | Imprime que hay un puntero y qué identificador nombra (sección 3.2), código 8, `code` `pointer_unresolved` |
| Hay más de un candidato | Imprime el elegido y los descartados, con el motivo, código 0 |
| El mismo `id` aparece en dos raíces | Imprime los dos directorios y no elige ninguno, código 11, `code` `ambiguous_board_id` (sección 3.2) |
| La base de datos del tablero no se puede leer | El mensaje de la sección 4.12, código 10, `code` `database_unreadable`. `where` no lo esquiva: para decir qué tablero está en uso hay que abrirlo |

#### Salida

```
id       3f9a2b1c
board    Kex
path     /Users/avilches/.biso/boards/kex-3f9a2b1c
source   project pointer at /Users/avilches/Hub/Projects/Kex
me       @claude
tasks    248 not archived, 31 archived, highest id ever assigned TASK-290
```

**La fila `path` es siempre la ruta ya resuelta del directorio del tablero, nunca el texto literal que
lleve el puntero.** Cuando el puntero trae una `path` relativa (sección 3.2), `where` la enseña resuelta
contra el directorio del puntero, que es el que la fila `source` nombra justo debajo, así que las dos
filas juntas dicen a la vez dónde está el tablero y de dónde salió esa respuesta. Enseñar el texto
literal sería enseñar la pregunta en vez de la respuesta: `where` existe para contestar dónde está el
tablero de verdad, y es el comando al que remiten los errores de código 8, donde una ruta que hay que
resolver a mano no sirve de nada. La clave `path` del sobre JSON lleva esa misma ruta resuelta, porque
es el mismo dato en la otra forma.

**Resuelta quiere decir también expandida**, así que la tilde del directorio personal no aparece nunca en
esta fila ni en la clave del JSON, aunque sí aparezca en el valor por defecto de `boards_root`
(sección 3.3) y en los ejemplos de esta especificación, donde se escribe para que se lean. `~` es una
abreviatura que expande el intérprete de órdenes, no una ruta, y una salida que la llevara obligaría a
quien la consume a expandirla por su cuenta.

La fila `source` nombra el directorio del que salió el puntero, y no solo la vía, porque con la
búsqueda de la sección 3.2 ese directorio puede ser cualquier ancestro del de trabajo: enseñarlo es lo
que hace visible de un vistazo el caso de haber heredado el puntero de un proyecto que contiene a este.
Con la otra vía no hay ningún directorio del que salir, porque el tablero es el directorio de trabajo,
y la fila dice `the working directory is this board`.

Y cuando no hay ninguno configurado, por stderr y con código 8:

```
error: no board here, and none configured for this project
searched  this directory: not a board
          pointer:        not found between this directory and /Users/avilches,
                          which is where the search stops (3.2)
hint: `biso init` creates one
```

Y cuando el puntero de este proyecto nombra un tablero que esta máquina no tiene, el mismo mensaje
que da cualquier otro comando en este caso (sección 3.2), por stderr y con código 8:

```
error: this project's pointer names board 3f9a2b1c, which is not on this machine
hint: `biso init` creates it here, adopting id 3f9a2b1c
```

**La fila de recuentos dice "not archived" y no "active", y la clave JSON se llama `notArchived` por lo
mismo.** La tabla de vocabulario de este documento reserva "active" para el papel del estado, el que
`biso start` usa, y una tarea sin archivar puede estar en cualquiera de los estados, incluido el
terminal. Llamarla activa haría que la misma palabra significara dos cosas en el mismo documento, y en
la fila donde más confunde, porque justo al lado hay un recuento de estados.

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "where",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "id": "3f9a2b1c",
    "board": "Kex",
    "path": "/Users/avilches/.biso/boards/kex-3f9a2b1c",
    "source": "project pointer at /Users/avilches/Hub/Projects/Kex",
    "me": "@claude",
    "counts": { "notArchived": 248, "archived": 31, "highestIdEverAssigned": "TASK-290" }
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tablero encontrado | 0 |
| No hay tablero | 8 |
| Su base de datos no se puede leer | 10 |
| El mismo `id` está en dos raíces | 11 |

#### `biso where --help`

```
Usage: biso where [options]

Say which board is in use and which rule picked it. Run it when a command
answers "no board here" and you expected one.

Options:
      --json     machine-readable envelope
  -h, --help     show this help

Exit codes:
  0  a board is in use
  8  no board here
  10 its database could not be read
  11 the same board id is in two places

Examples:
  biso where
  biso -C ~/work/kex where
```

---

### 10.3. `biso new`

#### Firma

```
biso new [<title>] [--start] [--from <file|->] [cualquier bandera de campo de la seccion 8]
```

#### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<title>` | | sí, salvo con `--from` | texto | | no | no | `--from` |
| `--start` | | no | booleano | falso | no | no | `-s`, `--from` |
| `--from <file\|->` | | no | ruta o `-` | | no | no | `<title>` y todas las de campo |

Todas las banderas de campo de la sección 8 valen aquí. En una tarea nueva no hay nada que sustituir
ni que quitar, así que `--set-*`, `--rm-*` y `--clear-*` se aceptan y hacen lo mismo que el nombre
desnudo, salvo `--clear-*`, que no hace nada y avisa. Las que se usan de verdad al crear son
`-d/--desc`, `--ac`, `--dod`, `--type`, `--priority`, `-l/--label`, `-a/--assignee`, `--ref`,
`--doc`, `--dep`, `-m/--milestone`, `-p/--parent`, `--due`, `--ordinal`, `--project`, `--reporter`,
`--ext`, `--plan`, `--note`, `--summary` y `--comment`.

- **`--start`** crea la tarea directamente en el estado activo, asignada a `me` y con el arrendamiento
  tomado a favor de quien llama (`leaseExpiresAt` y `leaseHolder`, sección 5), exactamente como lo haría
  `biso start` sobre ella. Es el atajo de esas dos llamadas, así que la equivalencia tiene que ser real:
  si `--start` dejara la tarea activa y asignada sin arrendamiento, `biso new "X" --start` y
  `biso new "X"` seguido de `biso start` darían dos tareas distintas. Es, junto con `biso start`, la
  única vía que fija `leaseHolder` fuera de la importación.
- **`--comment` funciona al crear**, igual que en cualquier otro comando de escritura.
- **`--plan`, `--note` y `--summary` no están restringidos por el estado.** Se pueden escribir al
  crear, en cualquier estado.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Título vacío o solo espacios | Error 2: `error: title cannot be empty` |
| Título muy largo | Se acepta entero, sin recortar |
| Título repetido | Se acepta sin aviso. Dos tareas pueden llamarse igual, para eso está el identificador |
| Valor fuera de un vocabulario cerrado | Error 3, con la lista de válidos |
| `--dep` a una tarea inexistente | Error 4. **Las dependencias se validan al escribirlas** |
| `--dep` a la propia tarea, o que cerraría un ciclo | Error 2 |
| `--parent` inexistente, o que cerraría un ciclo | Error 4 y error 2 respectivamente |
| `--ext` con una clave no declarada | Error 3 |
| `--due` con formato incorrecto | Error 2, señalando `YYYY-MM-DD` |
| `--due` en el pasado | Se acepta, con aviso |
| `-d @fichero` que no existe | Error 4 |
| `--start` sin ninguna identidad configurada (3.1) y sin `-a` | La tarea se crea en el estado activo y sin asignar, con `note: no identity configured, task left unassigned`, y **sin arrendamiento**: no hay ninguna identidad a la que atribuírselo, y una tarea sin asignar no puede tenerlo (sección 5). Es el mismo caso que la fila equivalente de `biso start` (10.7.1) |
| `--start` con `-a @sara` y una identidad configurada distinta | La tarea queda asignada a `@sara` y el arrendamiento es de quien llama, igual que en `biso start`: quien lo toma es quien escribe, no quien figura en `assignees` |
| Todo bien | Se crea la tarea, código 0 |

#### Salida

Por defecto, **una línea por tarea creada, con el identificador y nada más**:

```
TASK-101
```

`biso new` es el único comando de escritura cuya salida por defecto es distinta de la línea de estado
de 10.6, y así está dicho en el mensaje de arranque.

Con `--print`, después de la línea del identificador viene la ficha completa en el formato de
`biso get`. Con `--quiet`, solo el identificador y ninguna nota.

#### El modo lote

```
biso new --from tareas.ndjson
biso new --from -
biso new --from tareas.ndjson --dry-run
```

La entrada es **NDJSON**: un objeto JSON por línea. Las líneas vacías y las que empiezan por `#` se
ignoran. Las claves son las del modelo de datos de la sección 5, en `camelCase`.

Ejemplo de una línea, con todos los tipos compuestos:

```json
{"id":"TASK-101","title":"El diff no normaliza CRLF","type":"bug","priority":"high","status":"Done","description":"...","labels":["parser"],"references":["docs/bugs/BUG-02.md"],"dependencies":["TASK-90"],"ext":{"trello.card":"5f2a8c1e"},"acceptanceCriteria":[{"key":1,"text":"El diff ignora el CRLF","checked":true},{"key":3,"text":"Hay un test","checked":false}],"definitionOfDone":[{"key":1,"text":"Revisado","checked":true}],"comments":[{"author":"@avilches","createdAt":"2026-08-14T10:22:00Z","body":"Reportado desde Windows"}],"question":{"author":"@avilches","askedAt":"2026-08-16T09:00:00Z","body":"Es un CRLF o tambien un CR suelto?"},"createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
```

Las reglas del lote, todas obligatorias:

- **`acceptanceCriteria` y `definitionOfDone` aceptan dos formas.** Una cadena, que crea un elemento
  sin marcar con la siguiente clave libre, o un objeto con `key`, `text` y `checked`. Las dos formas
  se pueden mezclar dentro de la misma lista. Una `key` repetida dentro de la misma tarea es un fallo
  de validación.
- **El contador de claves de cada lista se sitúa por encima de la clave mayor importada**, de modo que
  un criterio añadido después nunca choca con uno importado. El contador no es una clave del formato:
  se deduce.
- **`comments` es una lista de objetos** con `author`, `createdAt` y `body`. `createdAt` es opcional y,
  si falta, se pone el instante de la importación.
- **`question` se acepta como objeto** con `author`, `askedAt` y `body` (5.7) en el lote de `--from`.
  `askedAt` es opcional y, si falta, se pone el instante de la importación, igual que `createdAt` en
  `comments`. Ausente la clave, la tarea se importa sin pregunta abierta.
- **`id`, `createdAt` y `updatedAt` se aceptan aquí y solo aquí.** Un `id` ya ocupado es un fallo de
  validación; un `id` libre se reserva y el tablero no lo volverá a asignar.
- **`leaseExpiresAt` y `leaseHolder` se aceptan aquí con el valor que traiga el fichero**, que es lo
  que hace cierta la garantía de simetría de 10.9 para ellos dos. La invariante de la sección 5 se
  comprueba en la validación, en sus dos mitades, y cada una es un fallo que nombra la línea y el campo.
  Una línea que traiga cualquiera de los dos sobre una tarea que no esté a la vez en el estado activo y
  asignada a alguien es un fallo de validación. Y una línea que traiga uno de los dos y no el otro
  también lo es, aunque la tarea esté activa y asignada: los dos campos van juntos, porque
  `leaseExpired` se calcula comparando `leaseExpiresAt` con el reloj de quien lee y con ese campo vacío
  no habría nada que comparar. Los dos fallos se ven así:
  ```
  line 14: leaseHolder on a task that is not both active and assigned
  line 31: leaseHolder given without leaseExpiresAt; the two go together
  ```
- **Un `id` explícito tiene que llevar el `task_prefix` del tablero de destino.** Si no lo lleva, es
  un fallo de validación, igual que un `id` ya ocupado: es la misma protección que hace inmutable a
  `task_prefix` en la sección 10.10, cerrando la tercera vía hacia el mismo tablero de identificadores
  mixtos que esa inmutabilidad ya evita en las otras dos (cambiar `--prefix` a mano, o renombrar el
  tablero). No es una restricción nueva sobre la simetría: exportar un tablero y restaurarlo con
  `biso snapshot` y `biso init --from` (10.9, 10.14) trae también su `task_prefix`, así que los `id`
  de su `snapshot.ndjson` siempre lo llevan puesto.
- **`archived` se acepta como booleano.** Por defecto, si la clave no aparece, la tarea se crea sin
  archivar. Ningún otro comando tiene una bandera de campo para él: fuera de la importación,
  archivar se hace con `biso archive`.
- **Una clave desconocida es un fallo de validación, no se ignora.** Ni la línea ni el lote se
  escriben, y el mensaje dice la línea y la clave.
- **Los campos derivados de la sección 5 no se aceptan.** En la entrada son claves desconocidas y
  por tanto un fallo de validación.
- **Se valida el fichero entero antes de escribir nada**, y se aplica la garantía de todo o nada de
  la sección 4.10.
- Un lote no admite `--start` ni ninguna bandera de campo: todo va en el fichero.

Salida del lote, una línea por tarea, en el orden del fichero:

```
TASK-101
TASK-102
TASK-103
```

Salida de `--dry-run` cuando todo está bien, por stderr y con código 0:

```
242 tasks would be created, nothing was written (--dry-run)
```

Y cuando no, por stderr y con código 9, **con todos los fallos, no solo el primero**:

```
error: 4 of 242 lines are invalid, nothing was written
  line 12: id "OTHER-5" does not match this board's task prefix "TASK"
  line 47: unknown status: "Pendiente" (valid: To Do, In Progress, Done)
  line 88: unknown key: "trelloCard"
  line 201: title cannot be empty
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea o lote creado | 0 |
| `--dry-run` que habría funcionado | 0 |
| Falta el título, banderas incompatibles, fecha mal formada, ciclo de dependencias o de padres | 2 |
| Valor fuera de un vocabulario, clave de extensión no declarada, entrada no interpretable | 3 |
| `--dep` o `--parent` a una tarea que no existe, o fichero de `@` que no existe | 4 |
| `--dep` o `--parent` por texto con varias coincidencias | 5 |
| Cualquier fallo de validación en un lote, o un `--dry-run` que no pasa | 9 |
| El almacén falla, o no se obtiene el acceso exclusivo | 7 |
| No hay tablero | 8 |

#### `biso new --help`

```
Usage: biso new <title> [options]
       biso new --from <file|-> [options]

Create a task and print its id. Every field flag of `biso set` works here.

Arguments:
  title                      task title (required unless --from is given)

Most used:
  -d, --desc <text>          description; repeat to append paragraphs
      --ac <text>            add an acceptance criterion; repeatable
      --dod <text>           add a definition-of-done item; repeatable
      --type <value>         configured type
      --priority <value>     configured priority
  -s, --status <value>       configured status (default: the initial one)
  -l, --label <value>        add a label; repeatable or comma-separated
  -a, --assignee <@who>      add an assignee; repeatable or comma-separated
      --dep <ref>            add a dependency; validated, repeatable
      --due <YYYY-MM-DD>     due date
      --comment <text>       add a discussion comment; repeatable
      --plan <text>          implementation plan
      --start                create it already in the active status, assigned
                             to you, with the lease claimed for you

Every other field flag of `biso set --help` is accepted too.

Batch:
      --from <file|->        NDJSON, one task object per line. The only place
                             where id, createdAt, updatedAt, criterion keys,
                             comment timestamps and question timestamps can be
                             given. Validated whole before anything is written.

Any text option also takes @file to read a file, or - to read stdin.

Exit codes:
  0  created            4  a referenced task or file does not exist
  2  bad usage          5  a text reference matched several tasks
  3  unknown value      7  the board could not be written
  9  batch or --dry-run validation failed, nothing was written
                        8  no board here

Examples:
  biso new "Normalize CRLF in the diff" --type bug --priority high
  biso new "Add OAuth" --ac "Login succeeds" --ac "Token refreshes"
  biso new "Rewrite the installer" -d @docs/installer.md --start
  biso new --from tasks.ndjson --dry-run
```

---

### 10.4. `biso ls`

#### Firma

```
biso ls [-s <status>]... [--not-status <status>]... [--any-status] [--archived] [--only-archived]
        [--type <v>]... [--priority <v>]... [--project <v>]...
        [-l <label>]... [--label-or <label>]... [-a <@who>]... [--mine] [--unassigned]
        [-m <milestone>] [-p <ref>] [--blocked] [--not-blocked] [--waiting] [--not-waiting]
        [--active] [--not-active] [--overdue] [--due-before <date>]
        [--search <text>] [--unchecked]
        [--sort <field>] [--reverse] [--limit <n>] [--all] [--ids] [--count]
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--status <v>` | `-s` | no | vocabulario | todos menos el terminal | sí | sí | `--any-status` |
| `--not-status <v>` | | no | vocabulario | | sí | sí | `--any-status` |
| `--any-status` | | no | booleano | falso | no | no | `-s`, `--not-status` |
| `--archived` | | no | booleano | falso | no | no | `--only-archived` |
| `--only-archived` | | no | booleano | falso | no | no | `--archived` |
| `--type <v>` | | no | vocabulario | | sí | sí | |
| `--priority <v>` | | no | vocabulario | | sí | sí | |
| `--project <v>` | | no | vocabulario | | sí | sí | |
| `--label <l>` | `-l` | no | etiqueta | | sí | sí | |
| `--label-or <l>` | | no | etiqueta | | sí | sí | |
| `--assignee <@w>` | `-a` | no | persona | | sí | sí | `--mine`, `--unassigned` |
| `--mine` | | no | booleano | falso | no | no | `-a`, `--unassigned` |
| `--unassigned` | | no | booleano | falso | no | no | `-a`, `--mine` |
| `--milestone <m>` | `-m` | no | hito | | no | no | |
| `--parent <ref>` | `-p` | no | referencia | | no | no | |
| `--blocked` | | no | booleano | falso | no | no | `--not-blocked` |
| `--not-blocked` | | no | booleano | falso | no | no | `--blocked` |
| `--waiting` | | no | booleano | falso | no | no | `--not-waiting` |
| `--not-waiting` | | no | booleano | falso | no | no | `--waiting` |
| `--active` | | no | booleano | falso | no | no | `--not-active` |
| `--not-active` | | no | booleano | falso | no | no | `--active` |
| `--overdue` | | no | booleano | falso | no | no | |
| `--due-before <d>` | | no | `YYYY-MM-DD` | | no | no | |
| `--search <text>` | | no | texto libre | | no | no | |
| `--unchecked` | | no | booleano | falso | no | no | |
| `--sort <field>` | | no | `urgency`, `id`, `ordinal`, `due`, `updated`, `created`, `title` | el orden de abajo | no | no | |
| `--reverse` | | no | booleano | falso | no | no | |
| `--limit <n>` | | no | entero >= 0 | 30 | no | no | `--all` |
| `--all` | | no | booleano | falso | no | no | `--limit` |
| `--ids` | | no | booleano | falso | no | no | `--count` |
| `--count` | | no | booleano | falso | no | no | `--ids` |

Reglas de combinación de filtros:

- **Filtros de campos distintos se combinan con `y`.** `-s "To Do" --type bug` son las que cumplen las
  dos cosas.
- **Valores repetidos del mismo campo se combinan con `o`.** `--type bug --type docs` son las de
  cualquiera de los dos tipos. Esto vale para `--status`, `--type`, `--priority`, `--project`,
  `--assignee` y `--label-or`.
- **`-l/--label` es la única que se combina con `y`.** `-l frontend -l bug` son las que llevan las
  dos. Para el `o` está `--label-or`, que valida igual.
- **`--unchecked` apaga la comprobación de existencia de `-l`, `--label-or`, `-a` y `-m`, y ninguna
  otra.** No cambia cómo se combinan ni afecta a ningún otro filtro. Los vocabularios configurados
  siguen validando, y `-p/--parent` sigue resolviendo su referencia.
- **El estado terminal se excluye por defecto**, y `--any-status` es la única forma de incluirlo.
- **Las archivadas se excluyen por defecto.** `--archived` las añade a las vivas y `--only-archived`
  deja solo las archivadas.
- **`--blocked` es incompatible con `--not-blocked`.** Las dos miran las dependencias sin terminar y
  no el estado, así que se combinan con cualquier filtro de estado y con los dos pares de abajo.
  `--not-blocked` por sí sola no dice que la tarea se pueda coger: descarta la que espera a otra
  tarea, no la que espera una respuesta ni la que ya lleva alguien.
- **`--waiting` es incompatible con `--not-waiting`, y `--active` con `--not-active`, cada una con su
  opuesta.** `--active` y `--not-active` filtran por el papel del estado y no por su nombre, que es su
  razón de ser: sin ellas, pedir la cola activa obligaría a escribir `-s "In Progress"`, el nombre
  concreto de un tablero concreto, y la misma consulta dejaría de servir en otro. Las cuatro son
  compatibles con `-s`, con `--not-status` y con `--any-status`, porque filtran sobre el mismo eje sin
  contradecirse: `-s "To Do" --active` es una lista vacía en unos tableros y no en otros. La regla
  general: **dos filtros que se contradicen por construcción son incompatibles. Una combinación de
  filtros válidos que resulte vacía en este tablero es un hecho legítimo sobre el tablero, no un
  error.**

#### La regla de orden, completa

Sin `--sort` se aplica el orden por defecto, que es esta tupla, en este orden y sin excepciones:

1. Las tareas que tienen `ordinal` van antes que las que no lo tienen.
2. Entre las que lo tienen, `ordinal` ascendente.
3. Entre las que no lo tienen, `urgency` descendente.
4. Cualquier empate se rompe por identificador ascendente, siempre.

Un `--sort` explícito sustituye los pasos 1 a 3 por ese campo, ascendente salvo `urgency`, que es
descendente por ser una medida de prioridad, y el paso 4 se sigue aplicando. `--reverse` invierte el
resultado final, incluido el desempate. **El orden nunca depende del estado**, porque el listado no
agrupa por estado.

Un `--sort due` o `--sort ordinal` sobre tareas que no tienen ese campo las pone al final, en bloque,
ordenadas por identificador.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Filtro con un valor fuera del vocabulario | Error 3, con la lista de válidos |
| `-l` con una etiqueta, `-a` con una persona o `-m` con un hito que el tablero no tiene | Error 3, con las cinco más parecidas |
| Lo mismo con `--unchecked` | Se acepta, y probablemente no devuelve nada |
| Filtro válido sin resultados | Ninguna línea por stdout, `note: no tasks match` por stderr, código **0** |
| Hay más resultados que el límite | Se imprimen los primeros y sale el aviso de recorte |
| `--limit 0` | No imprime ninguna fila, solo el aviso de recorte con el total. Es la forma de contar sin `--count` |
| `--count` | Un número por stdout y nada más |
| `--ids` | Identificadores, uno por línea, sin cabeceras ni columnas |
| Alguna tarea ilegible | Se salta, con el aviso de 4.12, y el resto del listado es válido |

#### Salida

Ocho columnas fijas, separadas por dos espacios, en este orden y con estos contenidos:

| Columna | Contenido | Cuando está vacío |
|---|---|---|
| 1 | identificador | nunca lo está |
| 2 | estado | nunca lo está |
| 3 | tipo | `-` |
| 4 | prioridad | `-` |
| 5 | título, recortado a **100 celdas siempre**, con `...` al final si se recorta | nunca lo está |
| 6 | `ac <marcados>/<total>` | `-` si la tarea no tiene criterios de aceptación |
| 7 | primera persona asignada, con `+<n>` si hay más | `-` |
| 8 | fecha límite | `-` |

**El formato se calcula así, en dos pasos, siempre en este orden:**

1. El título de cada tarea se recorta primero a 100 celdas, **contando los tres puntos**, así que la
   cadena que se imprime no pasa nunca de 100: un título más largo deja 97 celdas suyas y `...`
   detrás. Esto pasa antes de calcular ningún ancho de columna.
2. Para cada una de las columnas 1 a 7, el ancho de esa columna es la anchura del valor más largo
   que le corresponde entre las filas que se van a imprimir en esta llamada, y cada valor se rellena
   con espacios a la derecha hasta ese ancho. **La columna 8 nunca se rellena**, porque es la última
   y no hay nada después que alinear.

**La unidad de los dos pasos es la celda de un terminal monoespaciado, no el carácter.** Un título es
texto libre en UTF-8 (4.4), así que puede llevar ideogramas, emoji o acentos combinantes, y esas tres
cosas ocupan en pantalla algo distinto de lo que suman sus puntos de código: una marca combinante mide
cero celdas porque se pinta sobre la letra anterior, un ideograma de Asia oriental o un emoji miden dos,
y todo lo demás mide una. La tabla que lo dice es la de Unicode, la de anchura de Asia oriental más la
categoría de las marcas combinantes, y contar en cualquier otra unidad desalinea la tabla en cuanto un
título deja de ser ASCII.

**Medir en celdas no es mirar el terminal, así que no contradice la sección 4.1.** La anchura de un
carácter es una propiedad de Unicode, la misma en cualquier máquina y con cualquier ventana, y por eso
la salida sigue sin depender de dónde se ejecute el programa. Lo que 4.1 prohíbe es lo otro: preguntar
cuántas columnas tiene la ventana, o si hay color, para decidir qué se imprime.

**Y el recorte nunca parte un grafema por la mitad.** Si cortar exactamente en la celda 97 separaría una
letra de su acento combinante, o partiría un emoji compuesto, se corta en la frontera anterior, así que
el título recortado puede medir 96 o 95 celdas en vez de 97. La promesa es el tope, nunca la longitud
exacta: la cadena impresa no pasa de 100 celdas.

Entre columna y columna van siempre **dos espacios literales**, se haya rellenado o no la columna
anterior. Con estas cuatro tareas, el título más largo mide 28 caracteres y por eso la columna 5 se
rellena a ese ancho, no a uno fijo:

```
TASK-7   To Do        bug   high    Crash on an empty repository  ac 0/4  -        2026-09-08
TASK-11  In Progress  bug   high    Normalize CRLF in the diff    ac 1/2  @claude  -
TASK-19  To Do        task  high    Retry the upload on 5xx       ac 0/2  -        -
TASK-23  To Do        docs  medium  Rewrite the install section   ac 0/1  @sara+1  -
```

Y por stderr, siempre que se haya recortado:

```
warning: 28 more tasks match; showing 30 of 58
hint: narrow with -s, --type or -l, or ask for everything with --all
```

Con `--ids`:

```
TASK-7
TASK-11
```

Con `--count`:

```
58
```

**No hay agrupación por estado.** El estado es una columna más, para que cada línea se pueda tratar
igual que las demás.

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.list",
  "generatedAt": "2026-09-06T13:26:41Z",
  "data": {
    "tasks": [
      {
        "id": "TASK-11",
        "title": "Normalize CRLF in the diff",
        "status": "In Progress",
        "type": "bug",
        "priority": "high",
        "project": null,
        "assignees": ["@claude"],
        "reporter": "@avilches",
        "labels": ["parser"],
        "milestone": null,
        "parent": null,
        "dependencies": [],
        "references": ["docs/bugs/BUG-02.md"],
        "documentation": [],
        "modifiedFiles": [],
        "due": null,
        "ordinal": null,
        "createdAt": "2026-09-06T09:12:04Z",
        "updatedAt": "2026-09-06T11:40:18Z",
        "leaseExpiresAt": "2026-09-06T15:40:18Z",
        "leaseHolder": "@claude",
        "acDone": 1,
        "acTotal": 2,
        "dodDone": 0,
        "dodTotal": 1,
        "commentCount": 1,
        "urgency": 19.0,
        "blocks": ["TASK-40"],
        "blocked": false,
        "waiting": false,
        "leaseExpired": false,
        "archived": false,
        "ext": { "trello.card": "5f2a8c1e3b9d4a7f6e0c2b81" }
      }
    ],
    "shown": 30,
    "matched": 58,
    "hidden": 28,
    "truncated": true,
    "skipped": [],
    "sort": "default",
    "filters": { "status": ["To Do", "In Progress"], "type": [], "label": [] }
  }
}
```

**El listado nunca trae el cuerpo de la tarea**: ni descripción, ni plan, ni notas, ni criterios, ni
comentarios. Para eso está `biso get`. Los diez campos derivados de la sección 5 sí están todos,
`blocks` incluido. `truncated` es explícito para que nadie tenga que comparar `shown` con `matched`,
y `skipped` lleva los identificadores de las tareas ilegibles que se han saltado.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Listado, incluso vacío o con tareas saltadas | 0 |
| Un valor de filtro no existe en el tablero | 3 |
| Banderas incompatibles, `--limit` negativo, `--sort` inventado, fecha mal formada | 2 |
| `--parent` a una tarea que no existe | 4 |
| `--parent` por texto con varias coincidencias | 5 |
| `--mine` sin ninguna identidad configurada (3.1) | 6 |
| El almacén no responde | 7 |
| No hay tablero | 8 |

#### `biso ls --help`

```
Usage: biso ls [options]

List tasks, one per line. Shows 30 by default, hides the Done ones and the
archived ones, and says on stderr what it left out. A filter value the board
does not have is an error, never an empty list, so an empty list is a fact.

Filters (repeat or comma-separate; same field is OR, different fields are AND):
  -s, --status <value>       configured status (default: all but the terminal)
      --not-status <value>   exclude a status
      --any-status           include the terminal status too
      --archived             include archived tasks
      --only-archived        only archived tasks
      --type <value>         configured type
      --priority <value>     configured priority
      --project <value>      configured project
  -l, --label <value>        label; several labels are ANDed
      --label-or <value>     label; several are ORed
  -a, --assignee <@who>      assignee
      --mine                 assigned to you
      --unassigned           assigned to nobody
  -m, --milestone <text>     milestone, matched like any board value
  -p, --parent <ref>         subtasks of this task
      --blocked              something unfinished blocks it
      --not-blocked          nothing unfinished blocks it; it may still be
                             waiting on an answer, so add --not-waiting
      --waiting              has an open question
      --not-waiting          has no open question
      --active               in the board's active status
      --not-active           not in the active status
      --overdue              past its due date
      --due-before <date>    due before YYYY-MM-DD
      --search <text>        free text; see `biso get --help` for the scope
      --unchecked            do not check that the labels, assignees and
                             milestones you filter by exist on the board;
                             nothing else changes

Shape:
      --sort <field>         urgency, id, ordinal, due, updated, created, title
      --reverse              flip the whole order, tie-breaks included
      --limit <n>            how many rows to print (default 30, 0 prints none)
      --all                  print every match
      --ids                  print only ids, one per line
      --count                print only how many match

Columns: id, status, type, priority, title, criteria, assignee, due. Empty
cells print a dash. The title is cut at 100 characters, always, before any
column width is computed. Columns 1 to 7 are padded to the widest value
printed; column 8 never is. Two spaces always separate columns.

Exit codes:
  0  listed, even when empty      5  --parent matched several tasks
  2  bad usage                    6  --mine with no identity configured
  3  a filter value does not exist here
  4  --parent does not exist      7  the board could not respond
                                  8  no board here

Examples:
  biso ls
  biso ls -s "In Progress" --mine
  biso ls --type bug --priority high --limit 10
  biso ls --not-blocked --not-waiting --ids
  biso ls --any-status --archived --all
```

---

### 10.5. `biso get`

#### Firma

```
biso get <ref> [--id] [--match] [--section <name>]... [--explain-urgency]
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí | referencia | | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |
| `--section <name>` | | no | `meta`, `desc`, `ac`, `dod`, `plan`, `notes`, `summary`, `comments`, `question` | todas | sí | sí | |
| `--explain-urgency` | | no | booleano | falso | no | no | |

`--section` sirve para pedir solo una parte. `biso get TASK-11 --section ac` imprime los criterios con
sus claves y cuesta unas decenas de bytes en vez de la ficha entera, que es lo que hace falta antes de
marcar uno.

#### Comportamiento, caso a caso

La resolución de `<ref>` está en la sección 7 y no se repite. Lo propio de este comando:

| Caso | Qué pasa |
|---|---|
| La referencia resuelve a una tarea | Se imprime, código 0 |
| La referencia es texto y encaja con varias | Error 5, y las candidatas salen **por stdout** en el formato de `biso ls` |
| La referencia es texto y encaja con una | Se imprime, con `note: "CRLF" matched TASK-11` por stderr |
| La tarea está archivada | Se imprime, con `note: TASK-11 is archived` por stderr |
| La tarea no se puede leer | Error 3, según la regla de lectura dirigida de 4.12 |
| `--section` con un nombre inventado | Error 2, con los nueve nombres válidos |
| `--section` de una sección vacía | No imprime esa sección, y si no queda ninguna sección que imprimir, la salida está vacía y el código sigue siendo 0 |

**Sin `--section`, la ficha completa imprime siempre las nueve secciones fijas, vacías incluidas,
marcadas con `(empty)`.** Es solo con `--section` que una sección vacía se omite en vez de imprimirse
vacía; sin la bandera, omitir una sección la confundiría con una que no se ha pedido.

#### Salida

```
TASK-11  Normalize CRLF in the diff
status     In Progress          type       bug
priority   high                 urgency    19.0
assignees  @claude              reporter   @avilches
labels     parser               milestone  -
parent     -                    due        -
project    -                    ordinal    -
created    2026-09-06 09:12     updated    2026-09-06 11:40
depends    -                    blocks     TASK-40
lease      2026-09-06 15:40     holder     @claude
refs       docs/bugs/BUG-02.md
docs       -
files      -
ext        trello.card=5f2a8c1e3b9d4a7f6e0c2b81

## Description
El diff compara byte a byte y marca como distintas dos lineas que solo difieren
en el fin de linea.

## Acceptance Criteria
- [x] #1 El diff ignora el CRLF
- [ ] #3 Hay un test que lo cubre

## Definition of Done
- [ ] #1 Revisado por otra persona

## Implementation Plan
1. Leer el parser.
2. Anadir el caso CRLF.

## Implementation Notes
El parser ya normalizaba LF, faltaba CRLF.

## Final Summary
(empty)

## Comments
@avilches, 2026-09-06 10:02
Esto lo reporto un usuario con un repositorio clonado en Windows.

## Open Question
(empty)
```

Los encabezados de esta salida son un formato de presentación, no un formato de almacenamiento.

**La línea `lease` sale solo cuando la tarea tiene arrendamiento**, y entonces sale con sus dos campos:
`lease` es `leaseExpiresAt`, con el mismo formato de instante que `created` y `updated`, y `holder` es
`leaseHolder` (sección 5). Los dos aparecen y desaparecen juntos, porque la sección 5 no admite uno sin
el otro. Pertenece al bloque de metadatos, así que la trae `--section meta` y no ninguna otra sección.
Es la única línea condicional de ese bloque, y por eso va al final de las líneas de dos campos: así
ninguna de las de arriba cambia de sitio según la tarea. Eso no choca con la regla de que la ficha
completa imprime las nueve secciones aunque estén vacías, porque lo condicional es una línea del bloque
y no el bloque. Una tarea sin arrendamiento **no imprime la línea**, en vez de imprimirla con dos
guiones, porque eso pondría dos guiones en la ficha de casi todas las tareas del tablero y la ausencia
de la línea dice lo mismo. Esta es la única forma de ver los dos campos sin `--json`: `biso prime` no
los trae (9.7) y `biso ls` tampoco (10.4). Si el arrendamiento está vencido, el instante ya lo dice y la
ficha no añade ninguna marca; el derivado `leaseExpired` ya calculado está en `--json`.

Con `--section ac`, solo el encabezado con el identificador y el título, y la sección pedida:

```
TASK-11  Normalize CRLF in the diff

## Acceptance Criteria
- [x] #1 El diff ignora el CRLF
- [ ] #3 Hay un test que lo cubre
```

Con `--section question` sobre TASK-60, la tarea con la pregunta abierta del ejemplo de 9.7, la
sección sale rellena con la misma forma que ya usa `## Comments`: el autor y el instante en una línea
y el cuerpo debajo.

```
TASK-60  Confirm the retry budget for the upload endpoint

## Open Question
@claude, 2026-09-06 09:30
Should the retry budget be shared with the download endpoint or kept separate?
```

Con `--explain-urgency`, al final y por stdout:

```
urgency 19.0
  priority high      6.0 * 1.00 =   6.00
  active             4.0 * 1.00 =   4.00
  blocking           8.0 * 1.00 =   8.00
  blocked           -5.0 * 0.00 =   0.00
  due               12.0 * 0.00 =   0.00
  has criteria       1.0 * 1.00 =   1.00
  age 0 days         0.5 * 0.00 =   0.00
                                 -------
                                   19.00
```

El término `active` vale `1.00` solo si el estado es el activo y la tarea no tiene una pregunta
abierta (5.7); en cualquier otro caso vale `0.00`, y la etiqueta dice cuál de los dos motivos se
aplica: `not active` si el estado no es el activo, `active, waiting` si lo es pero la tarea espera una
respuesta.

Sobre una tarea en el estado terminal, el desglose se sustituye por una línea:

```
urgency 0.0
  terminal status, urgency is zero by definition
```

#### El esquema JSON

Es el objeto de `task.list` más los campos del cuerpo:

```json
{
  "schemaVersion": 1,
  "kind": "task.get",
  "generatedAt": "2026-09-06T13:31:09Z",
  "data": {
    "task": {
      "id": "TASK-11",
      "description": "El diff compara byte a byte...",
      "acceptanceCriteria": [ { "key": 1, "text": "El diff ignora el CRLF", "checked": true },
                              { "key": 3, "text": "Hay un test que lo cubre", "checked": false } ],
      "definitionOfDone": [ { "key": 1, "text": "Revisado por otra persona", "checked": false } ],
      "plan": "1. Leer el parser.\n2. Anadir el caso CRLF.",
      "notes": "El parser ya normalizaba LF, faltaba CRLF.",
      "summary": null,
      "comments": [ { "author": "@avilches", "createdAt": "2026-09-06T10:02:11Z", "body": "Esto lo reporto..." } ],
      "question": null,
      "blocks": ["TASK-40"],
      "urgencyBreakdown": { "priority": 6.0, "active": { "value": 4.0, "reason": null }, "blocking": 8.0,
                            "blocked": 0.0, "due": 0.0, "criteria": 1.0, "age": 0.0 }
    }
  }
}
```

**`urgencyBreakdown` solo sale con `--explain-urgency`**, igual que el desglose de la salida de texto, y
el ejemplo de arriba es el de una llamada que la lleva. Es la única clave de todo el documento que una
bandera añade, y la excepción a la regla de las claves siempre presentes está declarada en 12.4, junto
con la otra cosa que `biso get` hace con sus banderas: recortar `data.task` con `--section`.

`urgencyBreakdown.active` es el único término que no es un número suelto: `value` es el número que
entra en la suma, el producto del coeficiente por el factor, igual que en los demás términos.
`reason` vale `null` cuando el término contribuye, y cuando contribuye `0.0` dice por qué:
`"not_active"` si el estado no es el activo, y `"waiting"` si lo es pero hay una pregunta abierta.

Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas. Con varias
coincidencias, `kind` es `task.candidates`, `data.tasks` es la lista y el código es 5.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Tarea impresa | 0 |
| `<ref>` mal formada, `--section` inventada, `--id` con `--match` | 2 |
| La tarea no se puede leer | 3 |
| No existe, o existió y ya no está | 4 |
| Texto con varias coincidencias | 5 |
| El almacén no responde | 7 |
| No hay tablero | 8 |

#### `biso get --help`

```
Usage: biso get <ref> [options]

Show one task. <ref> is an id (TASK-11), a bare number (11) or free text
("CRLF"). Free text that matches several tasks lists them and exits 5; it
never picks one for you.

Free text searches the title, description, plan, notes, final summary, the
text of the criteria and of the definition of done, the body of the comments,
the body of the open question and the labels. A match in the title always
wins over a match anywhere else. `biso ls --search` uses this same scope.

Options:
      --id                   force <ref> to be read as an id
      --match                force <ref> to be read as free text
      --section <name>       print only these sections; repeatable or comma
                             separated. One of: meta, desc, ac, dod, plan,
                             notes, summary, comments, question
      --explain-urgency      show how the urgency number is built
  -h, --help                 show this help

Exit codes:
  0  printed          4  not on this board
  2  bad usage        5  the text matched several tasks
  3  the task could not be read
  7  the board could not respond
  8  no board here

Examples:
  biso get TASK-11
  biso get 11 --section ac
  biso get "CRLF"
  biso get TASK-11 --explain-urgency
```

---

### 10.6. `biso set`

El comando de edición general. Todo lo que hacen los verbos de flujo de 10.7 se puede hacer aquí, con
más palabras.

#### Firma

```
biso set <ref>... [cualquier bandera de campo de la seccion 8]
         [--check <sel>]... [--uncheck <sel>]... [--check-dod <sel>]... [--uncheck-dod <sel>]...
         [--comment <text>]... [--comment-author <@who>] [--id] [--match]
```

#### Parámetros propios

**Todas** las banderas de las secciones 8.2, 8.3, 8.5 y 8.6 valen aquí, con exactamente el mismo
significado que en cualquier otro comando. Lo propio de `set`:

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--check <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--uncheck` |
| `--uncheck <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--check` |
| `--check-dod <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--uncheck-dod` |
| `--uncheck-dod <sel>` | | no | selector 8.4 | | sí | ver 8.4 | solape con `--check-dod` |
| `--comment <text>` | | no | texto largo | | sí | no | |
| `--comment-author <@who>` | | no | texto libre | `me` | no | no | requiere `--comment` |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

**El autor de un comentario se llama `--comment-author` en todos los comandos que lo aceptan**, sin
excepción, aunque en `biso comment` el prefijo parezca redundante. Un concepto, un nombre.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Ninguna bandera de cambio | Error 2: `error: nothing to change` con un puntero a `biso get` |
| Varias referencias | El mismo cambio se aplica a todas, con la garantía de todo o nada de 4.10 |
| Varias referencias y un selector que no sea `all` (clave, rango, lista o texto) | Error 2, porque el selector de una tarea no tiene por qué significar lo mismo en otra |
| Varias referencias y `--check all` | Válido |
| Una de varias referencias no existe | Error 4, y **no se escribe ninguna**, ni siquiera las buenas |
| Un `--set-*` pisa contenido no vacío | Se hace, con el aviso de 4.3 diciendo cuántos bytes ha reemplazado |
| Paso a un estado terminal con criterios sin marcar | Se hace, con aviso |
| Paso a un estado terminal con una pregunta abierta (5.7) | Se hace, con aviso, igual que en `biso finish` (10.7.4) y como atribuye 4.3 a cualquier llegada al estado terminal |
| Todas las banderas dejan la tarea igual | Código 0, con `note: TASK-11 unchanged`. Ningún campo de la tarea se escribe, `updatedAt` no cambia y `changed` sale vacía, pero si quien llama es `leaseHolder` **el arrendamiento se renueva igual**: es una escritura del tenedor sobre su tarea, y el latido no depende de si los valores coincidían (sexta precisión de la sección 5) |
| `--status` a un estado que no es el activo, `--clear-assignee` o `--rm-assignee` que deja la tarea sin nadie, sobre una tarea con arrendamiento | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea el arrendamiento; si era de otra identidad, sale además el aviso de 4.3 (séptima precisión de la sección 5) |
| `--comment-author` sin `--comment` | Error 2 |
| `--comment` sin `--comment-author` y sin ninguna identidad configurada (3.1) | Error 2 |
| La tarea no se puede leer | Error 3, y no se escribe nada |

#### Salida

Por defecto, **una línea por tarea afectada** con lo que quien llama no sabía: el estado resultante,
el avance de criterios y la urgencia recalculada. Los tres datos son derivados, y ninguno se puede
conocer sin leer la tarea. La TASK-11 de los ejemplos tiene además una definición de hecho de un
elemento, así que su línea trae también el avance de esa segunda lista:

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

**Esta es la línea de estado, y la imprimen también los seis verbos del ciclo de 10.7 y
`biso archive`.** La única excepción es `biso new`, por el motivo que da 10.3. Cada comando la enseña
con su propio ejemplo, pero las tres reglas de su forma se dicen aquí y no se repiten:

1. El trozo `ac <marcados>/<total>` sale siempre que la tarea tenga criterios de aceptación.
2. El trozo `dod <marcados>/<total>` sale siempre que tenga definición de hecho. Una tarea sin ninguna
   de las dos listas imprime solo el identificador, el estado y la urgencia.
3. La palabra `archived` cierra la línea cuando la tarea queda archivada, y solo entonces. Es lo único
   que un comando puede añadirle, y quien lo añade es `biso archive` (10.8).

Los avisos van por stderr:

```
warning: --set-plan replaced 412 bytes of existing content
```

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "task.write",
  "generatedAt": "2026-09-06T11:40:18Z",
  "data": {
    "tasks": [
      { "id": "TASK-11", "status": "In Progress", "acDone": 1, "acTotal": 2,
        "dodDone": 0, "dodTotal": 1, "urgency": 19.0,
        "changed": ["plan", "status"] }
    ],
    "warnings": [ { "code": "overwrite", "field": "plan", "bytes": 412, "task": "TASK-11" } ]
  }
}
```

`kind` es `task.write` para `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer` y
`archive`, para que quien consuma la salida no tenga que distinguir qué verbo la produjo. `changed`
dice qué campos han cambiado de verdad, que no es lo mismo que qué banderas se han pasado. En el lote
de `new --from`, las 242 tareas van en `data.tasks` de **un solo sobre**, no en 242 objetos sueltos.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Cambio aplicado, o nada que cambiar | 0 |
| Sin banderas de cambio, banderas incompatibles, selector por clave con varias tareas, solape | 2 |
| Valor fuera de un vocabulario, clave de extensión no declarada, tarea ilegible | 3 |
| Alguna referencia no existe, o un selector de texto no encaja con ningún criterio | 4 |
| Alguna referencia de texto encaja con varias tareas, o un selector con varios criterios | 5 |
| `--dry-run` que no pasa la validación | 9 |
| El almacén falla, o no se obtiene el acceso exclusivo | 7 |
| No hay tablero | 8 |

#### `biso set --help`

```
Usage: biso set <ref>... [options]

Change any field of one or more tasks, all or nothing. Every flag here means
the same in `biso new`, `biso start`, `biso note`, `biso comment`, `biso ask`,
`biso answer`, `biso finish` and `biso archive`.

The four shapes, and there is no field that breaks them:
  --label X        add one          --set-label X   replace the whole list
  --rm-label X     remove one       --clear-label   empty the list
The same works for --assignee, --ref, --doc, --dep, --file, --ac and --dod.

Prose fields have three, because a block of text has no single item to remove:
  --desc X adds, --set-desc X replaces, --clear-desc empties. The same for
  --plan, --note (whose replacement is --set-notes) and --summary.

External fields have three too: --ext key=value sets that one key, --rm-ext key
drops it, --clear-ext empties the map.

Scalars just take a value: -t/--title, -s/--status, --type, --priority,
--project, -m/--milestone, -p/--parent, --due, --ordinal, --reporter. Each has
a --clear-<field>. An empty string is never a way to clear anything.

Criteria and definition of done:
      --check <sel>          check criteria; sel is all, 3, 1-4, 1,3,7 or the
                             criterion text. The numbers are stable #N keys.
                             With several tasks, sel has to be all
      --uncheck <sel>        the opposite
      --check-dod <sel>      the same for definition-of-done items
      --uncheck-dod <sel>    the opposite
      --rm-ac <sel>          remove criteria by the same selector
      --rm-dod <sel>         remove definition-of-done items

Comments:
      --comment <text>       append a comment; repeatable
      --comment-author <@w>  who wrote it (default: you)

Resolution:
      --id / --match         force <ref> to be an id, or free text

A --set-* over existing content is allowed and warns on stderr with how many
bytes it replaced.

Exit codes:
  0  done                    5  something matched more than one thing
  2  bad usage               7  the board could not be written
  3  unknown value           9  --dry-run did not pass
  4  a task or a criterion was not found
                             8  no board here

Examples:
  biso set TASK-11 --priority high --label parser
  biso set TASK-11 --check 1,3 --note "Both covered by diff_test.rs"
  biso set TASK-11 TASK-12 --milestone "v1.2"
  biso set "CRLF" --set-desc @docs/bugs/BUG-02.md
```

---

### 10.7. Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`

Los seis aceptan **todas** las banderas de campo de la sección 8, igual que `set`. No son un
subconjunto: lo que aportan es un nombre y unos valores por defecto, de modo que el gesto frecuente
cabe en una llamada corta y el gesto raro sigue cabiendo en la misma llamada.

Los seis son escrituras sobre la tarea, así que a los cinco que no son `start` (`note`, `comment`,
`finish`, `ask`, `answer`) les aplica la regla general de la sección 5: si quien llama ya es
`leaseHolder`, renuevan `leaseExpiresAt`; si no lo es y el arrendamiento está vivo, no tocan ninguno de
los dos campos y avisan; y si no lo es y está vencido, lo dejan vencido. Ninguno de los cinco fija ni
transfiere `leaseHolder`: reclamar es de `start`, y de `biso new --start` al crear (10.3).
**Y por encima de todo eso está la invariante**: la escritura que saca la tarea del estado activo o la
deja sin ninguna persona asignada vacía los dos campos, sea quien sea quien la haga, así que
`biso finish` los vacía siempre y el aviso de un arrendamiento ajeno no lo impide (séptima precisión de
la sección 5).

Los seis imprimen también la misma línea de estado que `set`, con la forma y las reglas que define
10.6. Los ejemplos de más abajo son esa línea con los datos de la TASK-11, que tiene dos criterios de
aceptación y un elemento de definición de hecho.

El ciclo entero de una tarea es esto:

```
biso start  TASK-11 --plan "1. Leer el parser. 2. Anadir el caso CRLF."
biso note   TASK-11 "El parser ya normalizaba LF, faltaba CRLF"
biso finish TASK-11 --check all --check-dod all --summary "Normaliza CRLF en el diff, verificado con las pruebas."
```

#### 10.7.1. `biso start`

##### Firma

```
biso start <ref>... [--plan <text>] [-a <@who>]... [-s <v>] [--reopen]
           [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--status <v>` | `-s` | no | vocabulario | `active_status` | no | no | |
| `--reopen` | | no | booleano | falso | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

`--plan` y `-a/--assignee` son las banderas de campo de la sección 8, con su significado de siempre:
**las dos añaden**. `--plan` añade al plan existente y `--set-plan` lo reemplaza; `-a` añade una
persona y `--set-assignee` reemplaza la lista.

##### Qué hace

Cuatro cosas en una escritura: pone el estado activo, **asigna la tarea a `me` si no tiene ninguna
persona asignada**, toma el arrendamiento (`leaseExpiresAt`, `leaseHolder`, sección 5) a favor de
quien llama (renovándolo si ya era suyo, reclamándolo si estaba vencido, o tomándolo si era de otra
identidad: es el único comando que hace las tres cosas sobre una tarea que ya existe, sección 5), y
añade el plan si se ha pasado. Con `-s` a un estado que no es el activo no hay arrendamiento que tomar,
y la fila correspondiente de la tabla dice qué pasa entonces.

| Caso | Qué pasa |
|---|---|
| La tarea ya está en el estado activo | Se aplica el resto igual, con `note: TASK-11 was already In Progress` |
| La tarea ya está en el estado terminal | Error 6, salvo con `--reopen`, que la devuelve al estado activo |
| La tarea tiene dependencias sin terminar | Se empieza igual, con el aviso correspondiente. **Avisa, no impide** |
| La tarea tiene una pregunta abierta (5.7) | Se empieza igual, con el aviso correspondiente. **Avisa, no impide**, exactamente como con las dependencias sin terminar |
| El arrendamiento de la tarea está vencido (`leaseExpired`, sección 5) | Se reclama dentro de la misma transacción: `leaseHolder` pasa a ser quien llama y `leaseExpiresAt` se renueva, comprobando en esa misma transacción que seguía vencido, **para que de dos reclamaciones simultáneas del mismo arrendamiento vencido solo gane una**. Lo que esa comprobación no hace es impedirle escribir al tenedor viejo cuando despierte: ninguna escritura corriente suya renueva ni recupera un arrendamiento que ya es de otra identidad (sexta precisión de la sección 5), pero puede seguir anotando, comentando y cerrando la tarea, y con otro `biso start` se la lleva de vuelta con el aviso de la fila siguiente. Es la diferencia deliberada con el token de vallado del patrón, anotada como riesgo aceptado en la sección 11 de `DECISIONES.md` |
| El arrendamiento de la tarea está vivo y es de otra identidad | Se coge igual, con `warning: TASK-11's lease is held by @sara until 2026-09-08T14:00:00Z`. **Avisa, no impide**, por el mismo motivo que las dependencias sin terminar y la pregunta abierta: un bloqueo de flujo no evita el trabajo duplicado, solo empuja a rodear la herramienta modificando datos que no deberían tocarse |
| `-s` con un estado que no es el activo, por ejemplo `biso start TASK-1 -s "To Do"` | Se aplica todo lo demás, pero **no se fija ningún arrendamiento**, y si la tarea lo tenía se vacía como en cualquier otra escritura que la saque del estado activo (séptima precisión de la sección 5). Fijarlo ahí rompería la invariante de que los dos campos solo tienen valor en una tarea activa y asignada, y `-s` acepta cualquier estado del vocabulario, así que este caso existe. Sale `note: TASK-1 was moved to To Do, no lease was claimed` |
| La tarea ya tiene otra persona asignada | No se añade `me`, y sale `note: TASK-11 is assigned to @sara, left as is`. Con `-a` explícito, se añade lo que diga `-a` |
| No hay ninguna identidad configurada (3.1) y no se pasa `-a` | No asigna a nadie, con `note: no identity configured, task left unassigned`, y tampoco se fija el arrendamiento: no hay ninguna identidad a la que atribuírselo |
| La tarea ya tiene plan y se pasa `--plan` | Se añade al final, como toda bandera desnuda |
| Varias referencias | Todo o nada |

##### Salida

```
TASK-11  In Progress  ac 0/2  dod 0/1  urgency 19.0
```

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Empezada | 0 |
| Ya estaba terminada y no hay `--reopen` | 6 |
| Referencia mal formada, banderas incompatibles | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Referencia inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso start --help`

```
Usage: biso start <ref>... [options]

Take one or more tasks: move them to the active status, claim the lease for
you, assign them to you if nobody has them, and record a plan. One call.

Options:
      --plan <text>      add to the implementation plan; repeatable, and takes
                         @file and - like every text option
  -a, --assignee <@who>  add an assignee (--set-assignee replaces the list)
  -s, --status <value>   use another status instead of the active one; no lease
                         is claimed then, a lease only exists on an active task
      --reopen           allow starting a task that is already finished
      --id / --match     force <ref> to be an id, or free text
  -h, --help             show this help

Every field flag of `biso set --help` works here too.

Unresolved dependencies produce a warning, not an error: you decide. Taking
over a live lease held by someone else is the same: it warns, it does not
refuse.

Exit codes:
  0  started        4  not found        7  the board could not be written
  2  bad usage      5  ambiguous        9  --dry-run did not pass
  3  unknown value  6  already finished, use --reopen
                    8  no board here

Examples:
  biso start TASK-11 --plan "1. Read the parser. 2. Add the CRLF case."
  biso start 11
  biso start TASK-11 TASK-12
```

#### 10.7.2. `biso note`

##### Firma

```
biso note <ref> [<text>...] [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí, salvo con `--note` | texto largo | | sí, como posicional | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Cada texto es un párrafo propio en la sección de notas. Aceptan `@fichero` y `-` como cualquier texto
largo.

##### El posicional que parece un identificador

`biso note` toma **una sola** referencia, mientras que `set`, `start`, `finish` y `archive` toman
varias. Para que esa diferencia no produzca basura en silencio, **un posicional de texto que encaje
con la gramática de identificador de la sección 7.1 es un error 2**:

```
error: "TASK-2" looks like a task id, and `biso note` takes only one task
hint: to note the same thing on several tasks: biso set TASK-1 TASK-2 --note "..."
      to write that text literally:            biso note TASK-1 --note "TASK-2"
```

**`--note`, la bandera de campo de la sección 8, nunca pasa por esa comprobación**, porque no es un
posicional: es la vía para escribir una nota que de verdad diga `TASK-2`. La misma regla vale para
`biso comment`, con `--comment`.

##### Qué hace

Añade uno o más párrafos a las notas de implementación. **Nunca reemplaza.** Para reemplazar está
`biso set <ref> --set-notes`, que este comando no acepta por su nombre desnudo pero sí como bandera de
campo de la sección 8, igual que las demás.

| Caso | Qué pasa |
|---|---|
| Sin ningún texto y sin ninguna bandera de campo | Error 2 |
| Texto vacío | No añade nada y avisa, según 4.6 |
| La tarea no tiene notas todavía | Se crean |
| Varios textos | Un párrafo por texto, en el orden dado |

##### Salida

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Añadida | 0 |
| Sin texto, o un posicional que parece un identificador | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente, o fichero de `@` inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso note --help`

```
Usage: biso note <ref> <text>... [options]

Append one or more paragraphs to the implementation notes of ONE task. It never
replaces anything; `biso set <ref> --set-notes` does that.

Arguments:
  ref            one task: an id, a bare number or free text
  text           one paragraph per argument; @file and - work here too

Options:
      --id / --match   force <ref> to be an id, or free text
  -h, --help           show this help

Every field flag of `biso set --help` works here too. Use `--note <text>` for a
paragraph that is not checked against the id grammar, for when the note
itself looks like an id.

To note the same thing on several tasks, use `biso set A B --note "..."`.

Exit codes:
  0  appended       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso note TASK-11 "The parser already normalized LF, CRLF was missing"
  biso note 11 "First finding" "Second finding"
  biso note TASK-11 @/tmp/benchmark-output.txt
```

#### 10.7.3. `biso comment`

Un comentario y una nota son cosas distintas. Una nota de implementación es el registro técnico de
quien hace el trabajo. Un comentario es una conversación, tiene autor y fecha, y es el canal por el
que entra lo que viene de fuera.

##### Firma

```
biso comment <ref> [<text>...] [--comment-author <@who>]
             [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí, salvo con `--comment` | texto largo | | sí | no | |
| `--comment-author <@who>` | | no | texto libre | `me` | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Se aplican las mismas reglas de posicional que en `biso note`, incluida la del texto que parece un
identificador. El autor es texto libre, no se valida contra nada y no interpreta el `@` inicial. **Sin
`--comment-author` y sin ninguna identidad configurada (3.1), es error 2**: `error: --comment-author is
required, no identity is configured`. Lo mismo vale para `--comment` en cualquier otro comando de
escritura.

##### Salida

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Y por stderr, `note: comment #2 by @trello:juan`.

##### Códigos de salida

Los mismos de `biso note`, más el error 2 de `--comment-author` sin identidad configurada.

##### `biso comment --help`

```
Usage: biso comment <ref> <text>... [options]

Append a discussion comment to ONE task, with an author and a timestamp. This
is not `biso note`, which records what you did while implementing.

Arguments:
  ref                      one task: an id, a bare number or free text
  text                     one comment per argument; @file and - work too

Options:
      --comment-author <@who>  free text author (default: you). An external
                               system can use its own convention, such as
                               @trello:juan. The @ is never a file reference
      --id / --match           force <ref> to be an id, or free text
  -h, --help                   show this help

Every field flag of `biso set --help` works here too. Use `--comment <text>`
for a comment that is not checked against the id grammar.

Comments are append-only: they are never edited and never deleted.

Exit codes:
  0  appended       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso comment TASK-11 "A user with a Windows clone reported this"
  biso comment 11 "Moved to Doing from the phone" --comment-author @trello:avilches
```

#### 10.7.4. `biso finish`

##### Firma

```
biso finish <ref>... [--summary <text>] [--check <sel>]... [--check-dod <sel>]...
            [--note <text>]... [--file <path>]... [-s <v>] [--strict] [--no-checks]
            [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--status <v>` | `-s` | no | vocabulario | `terminal_status` | no | no | |
| `--strict` | | no | booleano | el valor de `finish_strict` | no | no | `--no-checks` |
| `--no-checks` | | no | booleano | falso | no | no | `--strict` |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

`--summary`, `--check`, `--check-dod`, `--note` y `--file` son las banderas de campo de siempre.

##### Qué hace

Marca criterios y elementos de la definición de hecho, escribe la última nota y el resumen, apunta los
ficheros tocados y mueve al estado terminal, todo en una escritura.

| Caso | Qué pasa |
|---|---|
| Quedan criterios sin marcar y no se pasó `--check` | **Se cierra igual**, con el aviso y la lista de los que faltan |
| Quedan elementos de la definición de hecho sin marcar | Igual, con su propio aviso |
| Lo mismo, con `--strict` | Error 6, y no se escribe nada |
| Sin `--summary` | Se cierra igual, con `warning: TASK-11 finished without a final summary` |
| Sin `--summary` y con `--strict` | Error 6 |
| La tarea tiene subtareas sin terminar | Aviso con la lista. Con `--strict`, error 6 |
| La tarea tiene una pregunta abierta (5.7) | Se cierra igual, con el aviso correspondiente. **Avisa, no impide, ni con `--strict`**: impedirlo empujaría a rodear la herramienta con `biso set` |
| La tarea ya estaba terminada | Se aplica el resto sin cambiar el estado, con un `note:` |
| La tarea tiene el arrendamiento vivo de otra identidad | Se cierra igual, con el aviso de 4.3 de que era de otra persona, y `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura. La invariante gana sobre el "no tocar los dos campos" de una escritura ajena, porque una tarea terminada con arrendamiento vivo es un tablero que su propia importación rechazaría (séptima precisión de la sección 5) |
| La tarea tiene el arrendamiento y `-s` la lleva a otro estado que tampoco es el activo | Los dos campos se vacían igual: lo que los sostiene es estar en el estado activo, no llegar al terminal |
| `--no-checks` | Se salta todas las comprobaciones y no emite ninguno de esos avisos, incluido el de la pregunta abierta |
| Varias referencias | Todo o nada |

Quien quiera la política dura tiene `--strict`, y puede fijarla por defecto con
`biso config set finish_strict true`.

##### Salida

```
TASK-11  Done  ac 2/2  dod 1/1  urgency 0.0
```

La urgencia de una tarea en el estado terminal es cero por definición, según 5.4.

Por stderr, cuando toca:

```
warning: TASK-11 moved to Done with 1 of 2 acceptance criteria unchecked
  #3 Hay un test que lo cubre
```

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Cerrada | 0 |
| `--strict` y falta algo | 6, y no se escribe nada |
| Banderas incompatibles, selector por clave con varias tareas | 2 |
| Valor fuera de un vocabulario, tarea ilegible | 3 |
| Referencia o criterio inexistente | 4 |
| Referencia o criterio ambiguo | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso finish --help`

```
Usage: biso finish <ref>... [options]

Close one or more tasks: check criteria, add the last note, write the final
summary and move to the terminal status. One call.

Options:
      --summary <text>   add to the final summary; repeatable, takes @file and -
      --check <sel>      check criteria: all, 3, 1-4, 1,3,7 or the text. With
                         several tasks the selector has to be `all`
      --check-dod <sel>  the same for the definition of done
      --note <text>      one last implementation note; repeatable
      --file <path>      record a modified file; repeatable
  -s, --status <value>   use another status instead of the terminal one
      --strict           refuse to finish with unchecked criteria, unchecked
                         definition of done, unfinished subtasks or no summary
                         (default: warn and go on; see finish_strict)
      --no-checks        skip every check and every warning
      --id / --match     force <ref> to be an id, or free text
  -h, --help             show this help

Every field flag of `biso set --help` works here too.

Exit codes:
  0  finished       3  unknown value    6  --strict and something is missing
  2  bad usage      4  not found        7  the board could not be written
                    5  ambiguous        9  --dry-run did not pass
                                        8  no board here

Examples:
  biso finish TASK-11 --check all --check-dod all --summary "Normalizes CRLF"
  biso finish TASK-11 --check "covers CRLF" --note "313 tests green"
  biso finish TASK-11 TASK-12 --check all --summary "Both closed by PR 42"
```

#### 10.7.5. `biso ask`

##### Firma

```
biso ask <ref> <text>... [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí | texto largo | | sí, como posicional | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Cada texto es un párrafo propio del cuerpo de la pregunta, igual que en `biso note` (10.7.2). Acepta
`@fichero` y `-` como cualquier texto largo (4.5).

Se aplica la misma regla del texto que parece un identificador que `biso note` (10.7.2): un
posicional que encaja con la gramática de identificador de la sección 7.1 es error 2. Pero el mensaje
es propio, porque `biso ask` no tiene una bandera de campo que escriba `question`, así que la única
salida es `@fichero` o `-`, nunca `--note`:

```
error: "TASK-2" looks like a task id, and `biso ask` takes only one task
hint: to write that text literally, use @file or - for stdin
```

**`biso ask` no acepta `--comment-author`**: el autor de la pregunta es siempre la identidad
configurada (3.1).

##### Qué hace

Llena el campo `question` (5.7) con el autor y el instante que fija el programa y el texto dado,
junto con cualquier otra bandera de campo que se haya pasado en la misma escritura. **No cambia el
estado de la tarea.**

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Se llena el campo. **No cambia el estado** |
| La tarea ya tiene una pregunta abierta | Error 6, para que la segunda no borre a la primera en silencio |
| La tarea está en el estado terminal | Error 6, igual que `biso start`, con la pista de reabrirla |
| La tarea está archivada | Se hace, con `note: TASK-11 is archived` por stderr, igual que `biso get` |
| El texto está vacío | Error 3: `error: the question cannot be empty`, `code` `empty_scalar_value` (4.6) |
| Sin identidad configurada (3.1) | Error 2: `error: biso ask needs an identity; set it with biso config set me <you> or BISO_ME` |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que `biso note` (10.7.2) |
| Varias referencias | No se admiten: toma exactamente una, como `biso note` y `biso comment` |

Los dos errores 6 llevan pista:

```
error: TASK-11 already has an open question
hint: answer it first with `biso answer TASK-11 <text>`
```

```
error: TASK-11 is already Done
hint: reopen it first with `biso start TASK-11 --reopen`
```

##### Salida

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 15.0
```

La urgencia queda por debajo de los 19.0 del ejemplo de 5.4 porque el término de actividad exige
también que no haya pregunta abierta: la tarea sigue en el estado activo, pero `waiting` ya es
cierto.

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Preguntada | 0 |
| Ya hay una pregunta abierta, o la tarea ya está en el estado terminal | 6 |
| Referencia mal formada, banderas incompatibles, posicional que parece un identificador | 2 |
| Sin identidad configurada | 2 |
| Pregunta vacía, tarea ilegible | 3 |
| Referencia inexistente, o fichero de `@` inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso ask --help`

```
Usage: biso ask <ref> <text>... [options]

Park ONE task on a question for a person. The task keeps its status, but it
leaves the IN PROGRESS block of `biso prime` and shows up under NEEDS ANSWER
until somebody runs `biso answer`.

Arguments:
  ref                one task: an id, a bare number or free text
  text               the question; @file and - work too

Options:
      --id / --match force <ref> to be an id, or free text
  -h, --help         show this help

Every field flag of `biso set --help` works here too, but nothing except this
command and `biso answer` ever writes the question itself.

A task holds one open question at a time. Answer it before asking another.

Exit codes:
  0  asked          4  not found        7  could not be written
  2  bad usage      5  ambiguous        8  no board here
  3  empty question, or unreadable      9  --dry-run did not pass
  6  already asking, or already finished

Examples:
  biso ask TASK-11 "Do we normalize binary files too, or only text?"
  biso ask 11 @/tmp/question.md
```

#### 10.7.6. `biso answer`

##### Firma

```
biso answer <ref> <text>... [--id] [--match] [cualquier bandera de campo de la seccion 8]
```

##### Parámetros propios

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, exactamente una | referencia | | no | no | |
| `<text>` | | sí | texto largo | | sí, como posicional | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Cada texto es un párrafo propio de la respuesta, igual que en `biso ask` (10.7.5) y en `biso note`
(10.7.2). Acepta `@fichero` y `-` como cualquier texto largo (4.5).

Se aplica la misma regla del texto que parece un identificador que `biso note` (10.7.2): un
posicional que encaja con la gramática de identificador de la sección 7.1 es error 2. Pero el mensaje
es propio, porque `biso answer` no tiene una bandera de campo que escriba `question`, así que la única
salida es `@fichero` o `-`, nunca `--note`:

```
error: "TASK-2" looks like a task id, and `biso answer` takes only one task
hint: to write that text literally, use @file or - for stdin
```

**`biso answer` no acepta `--comment-author`**: los dos comentarios que
escribe van siempre firmados por la identidad configurada (3.1). Quien necesite firmar un comentario
con otro autor tiene `biso comment --comment-author`, que sigue funcionando como siempre.

##### Qué hace

Vacía el campo `question` (5.7) en una sola escritura, con tres efectos en este orden exacto:

1. Añade al histórico de comentarios (5.2) uno con el `author`, el `askedAt` y el `body` que guardaba
   el campo: la pregunta se convierte literalmente en un comentario, con su autor y su instante
   originales (5.3).
2. Añade detrás un segundo comentario con el texto de la respuesta, firmado por la identidad
   configurada (3.1) y con el instante de ahora.
3. Vacía el campo.

**El orden no depende de la línea de comandos, según la regla de 4.9.** Dentro del paso de los
comentarios, los dos que escribe este verbo van siempre antes que cualquier `--comment` que se haya
pasado en la misma escritura. El vaciado del campo `question` es un paso propio de `biso answer`,
posterior a todos los de 4.9, y es siempre el último efecto de la escritura.

Como los comentarios se guardan y se muestran en orden de inserción y no de instante (5.2), el
comentario de la pregunta queda antes que el de la respuesta aunque su instante sea anterior, y el
instante de cada uno sigue diciendo la verdad.

| Caso | Qué pasa |
|---|---|
| La tarea no tiene pregunta abierta | Error 6, con la pista de usar `biso comment` |
| Falta el positional del texto | Error 2. Una respuesta sin respuesta no cierra nada |
| El texto está vacío (`biso answer TASK-11 ""`) | Error 3: `error: the answer cannot be empty`, `code` `empty_scalar_value` (4.6) |
| Sin identidad configurada (3.1) | Error 2: `error: biso answer needs an identity; set it with biso config set me <you> or BISO_ME` |
| Un posicional que encaja con la gramática de identificador | Error 2, la misma regla que `biso note` (10.7.2) |
| Se pasan además banderas de campo | Se aplican igual, como en cualquier verbo del ciclo |
| Varias referencias | No se admiten: toma exactamente una, como `biso note` y `biso comment` |
| La tarea está archivada | Se hace, con `note: TASK-11 is archived` por stderr, igual que `biso get` |
| La tarea está en el estado terminal | Se hace igual que en cualquier otro estado |

**No es simétrico con `biso ask`, y es a propósito.** `ask` sobre una tarea terminada es error 6,
porque no tiene sentido abrir una pregunta sobre algo que ya está cerrado. Pero responder una pregunta
que se quedó abierta al cerrar la tarea es la única vía de recuperación que existe: `biso finish`
avisa sin impedirlo (4.3) y esa pregunta desaparece de los bloques de `biso prime` y del `biso ls` por
defecto, riesgo que `docs/DECISIONES.md` (sección 11) acepta a propósito. Impedir `answer` sobre una
tarea terminada cerraría esa única vía.

```
error: TASK-11 has no open question
hint: use `biso comment` to add a comment
```

##### Salida

```
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

La urgencia recupera el término de actividad de 5.4, porque `waiting` vuelve a ser falso.

##### Códigos de salida

| Desenlace | Código |
|---|---:|
| Respondida | 0 |
| La tarea no tiene pregunta abierta | 6 |
| Referencia mal formada, banderas incompatibles, posicional que parece un identificador, falta el texto | 2 |
| Sin identidad configurada | 2 |
| Respuesta vacía, tarea ilegible | 3 |
| Referencia inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

##### `biso answer --help`

```
Usage: biso answer <ref> <text>... [options]

Answer the open question of ONE task and unpark it. In a single write this
moves the question into the comments with its original author and time, adds
your answer behind it, and clears the question.

Arguments:
  ref                one task: an id, a bare number or free text
  text               the answer; @file and - work too

Options:
      --id / --match force <ref> to be an id, or free text
  -h, --help         show this help

Every field flag of `biso set --help` works here too, so you can answer and
refine in one call.

Both comments are signed with your configured identity. This command does not
take --comment-author.

Exit codes:
  0  answered       4  not found        7  could not be written
  2  bad usage      5  ambiguous        8  no board here
  3  empty answer, or unreadable        9  --dry-run did not pass
  6  no open question

Examples:
  biso answer TASK-11 "Only text files. Binary ones are skipped entirely."
  biso answer 11 "Yes" --ac "A binary file is never touched"
```

---

### 10.8. `biso archive`

#### Firma

```
biso archive <ref>... [--unarchive] [--id] [--match]
             [cualquier bandera de campo de la seccion 8]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<ref>` | | sí, una o más | referencia | | sí | no | |
| `--unarchive` | | no | booleano | falso | no | no | |
| `--id` | | no | booleano | falso | no | no | `--match` |
| `--match` | | no | booleano | falso | no | no | `--id` |

Saca la tarea del tablero activo. **La tarea sigue existiendo**, su identificador sigue reservado,
`biso get` la encuentra avisando de que está archivada, y `biso ls --archived` la lista.

| Caso | Qué pasa |
|---|---|
| Ya estaba archivada | Código 0, con un `note:`, sin escribir |
| Otras tareas vivas dependen de ella | Aviso con la lista, se archiva igual |
| La tarea tiene arrendamiento, vivo o vencido | `leaseExpiresAt` y `leaseHolder` se vacían en esa misma escritura, sea de quien sea (séptima precisión de la sección 5). Si estaba vivo y era de otra identidad, sale además el aviso de 4.3 |
| `--unarchive` | La devuelve al tablero con el estado que tenía, y sin arrendamiento: si vuelve al estado activo, quien quiera trabajar en ella lo toma con `biso start` |
| Varias referencias | Todo o nada |

#### `biso delete` no existe, y su ausencia está especificada

Cualquier invocación de `biso delete`, `biso rm` o `biso remove` termina con código 2 y este mensaje
por stderr, en vez de volcar la lista de comandos:

```
error: there is no delete command, on purpose
hint: `biso archive <ref>` takes it off the board and keeps the history
      an archived task still exists: `biso ls --archived` lists them, and the
      id is never reused
```

#### Salida

```
TASK-11  Done  ac 2/2  dod 1/1  urgency 0.0  archived
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Archivada o desarchivada | 0 |
| `biso delete`, o banderas incompatibles | 2 |
| Tarea ilegible | 3 |
| Tarea inexistente | 4 |
| Referencia ambigua | 5 |
| El almacén falla | 7 |
| `--dry-run` que no pasa | 9 |
| No hay tablero | 8 |

#### `biso archive --help`

```
Usage: biso archive <ref>... [options]

Take tasks off the board without losing them. An archived task still exists,
`biso get` still finds it, `biso ls --archived` lists it, and its id is never
reused.

Options:
      --unarchive      put them back on the board
      --id / --match   force <ref> to be an id, or free text
  -h, --help           show this help

Every field flag of `biso set --help` works here too.

There is no delete command. Archiving is the way.

Exit codes:
  0  archived       3  the task could not be read    7  could not be written
  2  bad usage      4  not found                     9  --dry-run did not pass
                    5  ambiguous                      8  no board here

Examples:
  biso archive TASK-11
  biso archive TASK-11 TASK-12 TASK-13
  biso archive TASK-11 --unarchive
```

---

### 10.9. `biso export`

#### Firma

```
biso export [-o <file|->] [--no-archived] [cualquier filtro de biso ls, salvo --archived y --only-archived]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--out <file\|->` | `-o` | no | ruta o `-` | `-`, es decir stdout | no | no | |
| `--no-archived` | | no | booleano | falso | no | no | |
| filtros de `ls` | | no | | | | | `--sort`, `--limit`, `--all`, `--ids`, `--count`, `--archived`, `--only-archived` |

**`biso export` sin filtros exporta el tablero entero**: todos los estados, el terminal incluido, y
todas las tareas, las archivadas incluidas. **No hereda ni el límite por defecto de `biso ls` ni su
exclusión del estado terminal**, y no existe aquí ninguna bandera `--all`. Los filtros de `biso ls` se
aceptan para acotar a propósito, y las banderas de forma de `ls` no, porque un volcado no tiene forma
que elegir. **Tampoco se aceptan `--archived` ni `--only-archived`**, porque las archivadas ya salen
por defecto: la única bandera de `export` sobre el archivo es `--no-archived`.

#### La garantía de simetría

La salida es NDJSON, una tarea por línea, con **exactamente** las claves que acepta `biso new --from`,
en la forma de objeto que esa sección define para los criterios, la definición de hecho, los
comentarios y la pregunta abierta, e incluyendo `id`, `createdAt`, `updatedAt`, `archived`, `question`
y las claves estables de cada criterio.

**Los únicos campos que no salen son los derivados de la sección 5.** `question` sale en `export` y
entra de vuelta con `new --from`, con sus tres partes completas.

`export` solo lleva las tareas: reconstruir un tablero entero, con su vocabulario y no solo con sus
datos, es lo que hace `biso snapshot` (10.14), cuyo `snapshot.ndjson` tiene exactamente esta misma forma
y se lee de vuelta con el `--from` de `biso init` (10.1), no con el de `biso new`.

La garantía que la suite de pruebas comprueba:

```bash
biso snapshot
# escribe snapshot.ndjson y board.json en ~/.biso/boards/kex-3f9a2b1c, el propio
# directorio del tablero de origen (biso where lo muestra en su fila "path")
biso -C /tmp init --at /tmp/tablero-nuevo --from ~/.biso/boards/kex-3f9a2b1c
# los dos tableros son identicos en todos los campos no derivados, incluidos
# los identificadores, las fechas, las claves de los criterios y sus marcas,
# y en toda su configuracion: estados, tipos, extensiones y task_prefix
```

El `init` se ejecuta con `-C /tmp`, fuera del proyecto de origen, exactamente como en la versión
anterior de esta prueba: así su puntero de proyecto no choca con el que el proyecto de origen ya
tiene.

`biso init --from` lee el vocabulario del propio `board.json` de la instantánea, así que el tablero
de destino no necesita declarar nada a mano: nace con el mismo `task_prefix`, los mismos estados y
los mismos tipos que el de origen, y por eso la importación de su `snapshot.ndjson` nunca falla por
vocabulario distinto. Comparar esto con la vía manual de `biso new --from` (10.3): esa sigue
existiendo para importar un NDJSON suelto en un tablero cuyo vocabulario ya se ha declarado por
separado, pero ya no es la única manera de reconstruir un tablero entero.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Exportado, aunque sean cero tareas | 0 |
| Alguna tarea se ha saltado por ilegible | 6 |
| Banderas de forma de `ls`, `--archived`, `--only-archived`, `--json`, o incompatibles | 2 |
| Un valor de filtro no existe en el tablero | 3 |
| No se puede escribir el fichero de salida | 7 |
| No hay tablero | 8 |

#### `biso export --help`

```
Usage: biso export [options]

Write the board as NDJSON, one task per line, in exactly the shape that
`biso new --from` reads back. Round-tripping every non-derived field is a
tested guarantee: ids, dates, criterion keys and their checkmarks included.

With no filters it exports everything, the finished and the archived included.
It never inherits the default limit or the default status filter of `biso ls`.
If a task cannot be read, the rest is still written and the exit code is 6,
not 0: this is the one command whose purpose is to lose nothing.

Options:
  -o, --out <file|->   where to write (default: stdout)
      --no-archived    leave the archived tasks out
  -h, --help           show this help

Every filter of `biso ls` works here except --archived and --only-archived,
which do not apply because archived tasks are already included by default.
Its shaping flags (--sort, --limit, --all, --ids, --count) do not apply either.
--json is rejected with code 2: this output is already one JSON object per
line, while --json means the single envelope every other command prints.

Derived fields are never written: urgency, acDone, acTotal, dodDone, dodTotal,
commentCount, blocks, blocked, waiting, leaseExpired.

Exit codes:
  0  exported       3  a filter value does not exist here
  2  bad usage      6  some task was skipped, unreadable
  7  cannot write there                8  no board here

Examples:
  biso export -o backup.ndjson
  biso export -s Done --no-archived -o done.ndjson
```

---

### 10.10. `biso config`

#### Firma

```
biso config get <key>
biso config set <key> <value>
biso config list
```

#### Parámetros

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<key>` | | sí en `get` y `set` | una clave de la tabla de abajo | | no | no | |
| `<value>` | | sí en `set` | según la clave | | no | sí en las claves de lista | |

`--json` solo se acepta en `config list`. En `get` la salida ya es un solo valor y en `set` no hay
salida por stdout, así que en los dos es un error de uso con código 2.

#### Las claves

| Clave | Tipo | Por defecto |
|---|---|---|
| `project_name` | texto | el nombre del proyecto |
| `statuses` | lista, mínimo tres | `To Do, In Progress, Done` |
| `initial_status` | uno de `statuses` | `To Do`, al crear el tablero sin `--statuses` |
| `active_status` | uno de `statuses` | `In Progress`, al crear el tablero sin `--statuses` |
| `terminal_status` | uno de `statuses` | `Done`, al crear el tablero sin `--statuses` |
| `types` | lista | `task, bug, docs` |
| `priorities` | lista | `high, medium, low` |
| `projects` | lista | vacía |
| `labels` | lista | vacía |
| `assignees` | lista | vacía |
| `extensions` | lista | vacía |
| `task_prefix` | texto de solo letras | se deriva de `project_name` en mayúsculas (sección 4.11) |
| `me` | texto de persona | `BISO_ME` si está definida |
| `default_limit` | entero >= 0 | 30 |
| `finish_strict` | booleano | falso |
| `lease_minutes` | entero > 0 | 240 |
| `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`, `urgency.criteria`, `urgency.age` | decimal | ver 5.4 para el término de cada uno y su valor por defecto |

**`project_name` es el nombre del tablero, y cambiarlo no toca el sistema de ficheros.** Ninguna clave
de esta tabla lo hace. La carpeta del tablero se queda con el nombre que tenga, aunque sea el slug de un
nombre anterior, porque ese nombre es decorativo y nadie resuelve por él (sección 3.2): la identidad del
tablero está en el marcador `<id>.id` y dentro de la base de datos, no en el nombre de la carpeta. Así
que renombrar un tablero es una escritura en su base de datos y nada más, con la misma transacción y las
mismas garantías que cualquier otra.

**Eso quita de en medio la única operación que no podía ser atómica.** Mover un directorio no cabe dentro
de una transacción de SQLite, así que renombrar el tablero habría sido escribir la configuración y
después mover la carpeta, con un estado intermedio observable si la segunda mitad fallaba, un error 7
propio para el fallo de permisos, y la posibilidad de dejar sin resolver el puntero de un tablero que
viviera fuera de las raíces de la sección 3.3. Nada de eso existe: no hay dos mitades.

**Ningún carácter de `project_name` puede romper nada, y ahora por un motivo más simple**: no entra en
ninguna ruta. Sigue habiendo un slug derivado de él (sección 3.2), pero solo se usa para dar nombre a la
carpeta cuando `biso init` la crea, y ahí el valor ya está comprobado. Un `project_name` cuyo slug quede
vacío sigue siendo un error, porque el slug es un dato del tablero y la regla de que un valor no válido
nunca se acepta vale igual.

**Renombrar no toca nunca el `task_prefix`.** Se derivó una vez al crear el tablero y desde entonces
vive por su cuenta en esa clave. Cambiar `project_name` no lo recalcula, aunque el nombre nuevo diera un
`task_prefix` distinto si el tablero se creara hoy. Y si el nombre nuevo no deja ninguna letra con la
que derivar uno (sección 4.11), tampoco es un error aquí, porque `task_prefix` ya está fijado y no se
recalcula al renombrar.

Lo mismo vale para `biso init --overwrite-config` (sección 10.1), y solo cuando se da `<name>`
explícito: si ese `<name>` difiere del `project_name` que el tablero ya tenía, lo cambia igual que lo
haría `config set project_name`, y tampoco mueve nada. **Sin `<name>` explícito, `project_name` se
conserva tal cual estaba**, aunque el valor por defecto de `<name>` sea el nombre del directorio del
proyecto: ese valor por defecto tiene sentido como punto de partida al crear un tablero nuevo, no como
instrucción de renombrar uno que ya existe.

**`me` gana sobre `BISO_ME` cuando las dos están puestas.** Por eso un tablero compartido entre una
persona y un agente tiene que dejar `me` sin configurar: si la lleva puesta, todo el mundo comparte
identidad y `--mine` deja de significar nada (sección 11 de `docs/DECISIONES.md`).

**`lease_minutes` fija cuánto dura el arrendamiento de una tarea activa y asignada (sección 5), y se
puede cambiar libremente en cualquier momento, sin caer nunca en el error 6.** A diferencia de
`task_prefix` o de `statuses` en uso, esta clave no queda incrustada en ninguna tarea existente:
`leaseExpiresAt` se calcula al escribir, así que cambiar `lease_minutes` solo afecta a los
arrendamientos que se renueven desde ese momento, nunca a los ya fijados, y el tablero nunca queda
inconsistente por ello. El valor por defecto, 240 minutos, viene de que **el error dañino es el falso
vencido, no el vencido tardío**: un arrendamiento demasiado corto hace que un agente reclame una tarea
que otro está trabajando de verdad, mientras que uno demasiado largo solo retrasa el aviso. Y las
consecuencias de un valor mal calibrado son más pequeñas de lo que parecen, porque `biso start` avisa
y coge la tarea igual incluso con el arrendamiento vivo (10.7.1): una duración mal puesta produce un
informe equivocado, no datos equivocados.

**Los tres estados especiales son valores explícitos, no posiciones.** Se escriben al crear el tablero
y **cambiar `statuses` no los mueve nunca**. Esta es la diferencia que evita que añadir un estado al
final cambie en silencio a dónde va `biso finish`.

**`task_prefix` es inmutable en cuanto el tablero tiene alguna tarea.** Es la única clave cuyo valor
queda incrustado en datos que ya existen, porque cada identificador ya asignado lleva el prefijo
grabado. Mientras el tablero está vacío no hay ningún identificador con el que pueda entrar en
conflicto, así que cambiarla funciona sin más; en cuanto existe una sola tarea, cambiarla es error 6
(tabla de abajo), con el mismo motivo por el que no se toca `statuses` en uso. La misma regla vale
para `biso init --overwrite-config` (sección 10.1): reescribir la configuración de un tablero con
tareas nunca puede cambiar el `task_prefix` que ya tenía, se pase `--prefix` explícito o no.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Clave inexistente | Error 4, con las tres claves más parecidas |
| Valor del tipo equivocado, por ejemplo `finish_strict maybe` | Error 3, diciendo qué tipo esperaba |
| `initial_status` a un valor que no está en `statuses` | Error 3 |
| `lease_minutes` a cero o negativo | Error 3, el mismo trato que cualquier valor fuera de dominio de esta tabla |
| Quitar de `statuses` un estado que alguna tarea usa | Error 6, con cuántas tareas lo usan y en cuáles |
| Quitar de `statuses` un estado que es `initial_status`, `active_status` o `terminal_status` | Error 6, diciendo cuál de los tres y que hay que cambiarlo antes |
| Dejar `statuses` con menos de tres elementos | Error 6, diciendo cuántos hacen falta |
| Dar a un papel (`initial_status`, `active_status` o `terminal_status`) el mismo estado que otro papel ya tiene | Error 6, con los dos papeles y el estado que comparten |
| Quitar de `extensions` una clave que alguna tarea usa | Error 6, con la lista de tareas |
| Quitar de `types` o `priorities` un valor en uso | Error 6, igual |
| Cambiar `task_prefix` cuando el tablero ya tiene alguna tarea | Error 6, remitiendo a exportar el tablero, reescribir los identificadores e importarlos en un tablero nuevo |
| Cambiar `project_name` a un valor vacío, o a uno cuyo slug (sección 3.2) quede vacío tras derivarlo | Error 3, en los dos casos |
| Cambiar `project_name` al mismo valor que ya tiene | El `set` se completa igual, con su `note:` |
| Cambiar `project_name` a un valor que no dejaría ninguna letra para derivar un prefijo | No es error: el `task_prefix` ya está fijado y no se recalcula al renombrar |
| Cambiar `project_name` en un tablero cuyo puntero lleva `path`, o cuya carpeta ya no se llama como el nombre viejo | No es un caso especial: no se mueve nada y el puntero sigue valiendo, porque nada resuelve por el nombre de la carpeta (sección 3.2) |
| `get` de una clave de lista | Los valores separados por comas, en una línea |
| `set` correcto | Sin salida por stdout, con `note:` por stderr diciendo el valor nuevo |

Ningún cambio de configuración toca ninguna tarea, nunca.

#### Salida

```
$ biso config get statuses
To Do,In Progress,Done

$ biso config list
project_name = Kex
statuses = To Do,In Progress,Done
initial_status = To Do
active_status = In Progress
terminal_status = Done
types = idea,memory,task,bug,docs
priorities = high,medium,low
projects =
labels =
assignees =
extensions = trello.card
task_prefix = TASK
me = @claude
default_limit = 30
finish_strict = false
lease_minutes = 240
urgency.priority = 6.0
urgency.active = 4.0
urgency.blocking = 8.0
urgency.blocked = -5.0
urgency.due = 12.0
urgency.criteria = 1.0
urgency.age = 0.5
```

**`config list` imprime las veintitrés claves, siempre, en el orden de la tabla de claves de arriba**, y
los siete coeficientes de la urgencia con el nombre con el que `config set` los acepta, uno por línea.
Lo que `list` enseña es exactamente el conjunto de claves que `set` admite, y por eso no puede haber
ninguna que solo se vea con `--json`: una clave escondida es una clave que nadie sabe que puede cambiar.

**Una lista vacía se imprime como la clave, el igual y nada detrás**, que es lo mismo que hace
`config get` de una lista vacía, así que las tres listas que nacen vacías (`projects`, `labels` y
`assignees`) aparecen igual en un tablero recién creado.

#### El esquema JSON

Solo `config list` acepta `--json`:

```json
{
  "schemaVersion": 1,
  "kind": "config",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "config": {
      "project_name": "Kex",
      "statuses": ["To Do", "In Progress", "Done"],
      "initial_status": "To Do",
      "active_status": "In Progress",
      "terminal_status": "Done",
      "types": ["idea", "memory", "task", "bug", "docs"],
      "priorities": ["high", "medium", "low"],
      "projects": [],
      "labels": [],
      "assignees": [],
      "extensions": ["trello.card"],
      "task_prefix": "TASK",
      "me": "@claude",
      "default_limit": 30,
      "finish_strict": false,
      "lease_minutes": 240,
      "urgency": { "priority": 6.0, "active": 4.0, "blocking": 8.0, "blocked": -5.0,
                   "due": 12.0, "criteria": 1.0, "age": 0.5 }
    }
  }
}
```

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Hecho | 0 |
| Sintaxis, o `--json` fuera de `list` | 2 |
| Valor de tipo o de dominio incorrecto | 3 |
| Clave inexistente | 4 |
| El cambio dejaría el tablero inconsistente | 6 |
| No se puede escribir la configuración | 7 |
| No hay tablero | 8 |

#### `biso config --help`

```
Usage: biso config get <key>
       biso config set <key> <value>
       biso config list [--json]

Read and change the board configuration. List values are comma-separated.
No configuration change ever touches a task.

Keys:
  project_name       board name; changing it never moves anything on disk,
                     the board folder keeps whatever name it has
  statuses           the board statuses, in order
  initial_status     status of a new task           (one of statuses)
  active_status      what `biso start` sets         (one of statuses)
  terminal_status    what `biso finish` sets        (one of statuses)
  types              configured task types
  priorities         configured priorities
  projects           configured projects
  labels             labels that filters accept on top of the ones in use
  assignees          assignees that filters accept on top of the ones in use
  extensions         declared external field keys, such as trello.card
  task_prefix        id prefix, letters only (default: derived from
                     project_name); immutable once the board has a task
  me                 who you are, for --mine and for comment authorship
  default_limit      how many rows `biso ls` prints (default 30)
  finish_strict      make `biso finish` refuse an incomplete task
  lease_minutes      lease duration in minutes (default 240); free to change
                     at any time, it only affects future renewals
  urgency.priority, urgency.active, urgency.blocking, urgency.blocked,
  urgency.due, urgency.criteria, urgency.age
                     the seven urgency coefficients; see `biso get --explain-urgency`

The three special statuses are stored as explicit values. Changing `statuses`
never moves them; if a change would remove one of them, it fails and says so.

Removing any configured value that a task still uses is refused, never applied
silently.

Options:
      --json         machine-readable output, `list` only
  -h, --help         show this help

Exit codes:
  0  done            4  no such key
  2  bad usage       6  the change would leave the board inconsistent
  3  bad value       7  the configuration could not be written
                     8  no board here

Examples:
  biso config get active_status
  biso config set statuses "To Do,In Progress,Done"
  biso config set finish_strict true
  biso config list --json
```

---

### 10.11. `biso doctor`

#### Firma

```
biso doctor [--fix]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--fix` | | no | booleano | falso | no | no | ninguno |

**`biso doctor` sin `--fix` es de solo lectura**, y `--print` y `--dry-run` de la sección 3 son error 2
igual que en cualquier otro comando de lectura. **Con `--fix` es un comando de escritura**: ahí
`--dry-run` reporta qué se repararía sin reparar nada, y `--print` no añade nada, porque `doctor` no
imprime fichas de tareas.

#### Para qué sirve `biso doctor`, y para qué no

**Ningún comando remite a `biso doctor` para algo que podía arreglar dentro de lo que ya se le
pidió.** Si el arreglo no necesita ninguna decisión y cae dentro del trabajo que el comando iba a
hacer de todas formas, lo arregla y sigue, avisando con un `warning:` si merece la pena saberlo. El
ejemplo que ya se cumple: el contador del identificador más alto lo repara `biso new` por necesidad,
porque para asignar el siguiente tiene que saber el máximo de verdad, y eso no es reparar de paso sino
hacer bien su trabajo.

Los demás comandos no reparan lo que se encuentran, y hay tres razones, las tres apoyadas en promesas
que este documento ya hace:

- **Casi ningún comando ve nada que arreglar.** `biso get` lee una tarea y no puede detectar que otra
  tenga una dependencia rota. Lo que hace útil a `doctor` no es saber reparar, es mirar el tablero
  entero.
- **Los comandos de lectura no pueden escribir.** La sección 4.10 promete que las lecturas nunca
  fallan por una escritura en curso y nunca la bloquean, y la 9.4 que `biso prime` no escribe nunca y
  es seguro en paralelo. Un `ls` o un `prime` que repararan al pasar necesitarían acceso exclusivo,
  podrían esperar cinco segundos y fallar con código 7, y se perdería justo la garantía de los dos
  comandos que un agente llama sin parar.
- **Un comando de escritura tiene permiso para lo que se le pidió, no para más.** Y no es solo
  cuestión de sorpresa: con el todo o nada de la sección 4.10, si `biso set` cambiara un título y
  además reparara otra cosa, y la reparación fallara, habría que decidir si se deshace el título, o
  sea una transacción que abarca dos intenciones sin relación.

Eso le da a `--fix` su sentido exacto: **`--fix` no es una comodidad, es el consentimiento.** Es donde
quien llama dice "te autorizo a escribir cosas que no te he pedido una por una", y es precisamente la
autorización que ningún otro comando tiene.

Sale a `doctor` solo lo que cae en uno de estos dos casos:

- **Lo que necesita una decisión humana**, porque hay más de un arreglo válido y elegir por su cuenta
  destruiría información. Un ciclo de dependencias es el ejemplo: no se puede adivinar qué arista
  sobra.
- **Lo que solo pasa por daño externo**, porque nada dentro de `biso` lo produce: una base de datos
  corrupta, una carpeta que alguien movió a mano, un fichero que perdió sus permisos.

Y de ahí sale el corolario que hace falta para que las dos reglas no se peleen: **un comando que
tropieza con un problema reparable que no le toca arreglar lo dice con un `warning:` y nombra `biso
doctor --fix`.** Eso no es remitir a una llamada que se iba a hacer igual, que es justo lo que se
prohíbe: es contar algo que quien llama no sabía y que no iba a descubrir por su cuenta. La regla que
se prohíbe es remitir a `doctor` para algo que el comando ya tenía permiso para arreglar; avisar de lo
que no puede tocar es lo contrario de esconderlo.

Y queda una tercera razón para que `doctor` exista, que no es de reparación: es el único sitio donde
se puede preguntar **"le pasa algo a este tablero"** sin haber intentado antes una operación. Eso es
un diagnóstico que se corre cuando se sospecha, no una llamada que se iba a hacer igual.

**Ningún comando de `biso` pregunta nada por la entrada estándar, ni `doctor` con `--fix` ni ninguno
otro** (sección 4.1). La misma orden sirve para una persona y para un agente, y lo único que cambia es
quién lee la salida: la persona lee el informe, el agente mira el código de salida.

#### Errores y avisos, y solo los errores sacan el código 6

Lo que `doctor` reporta se divide en dos niveles:

- **Errores**: dejan el tablero inconsistente, o hacen imposible una operación, o vuelven un dato poco
  fiable. Si queda alguno sin reparar, el código de salida es 6.
- **Avisos**: son verdad, merece la pena saberlos, y no rompen nada. **No cambian el código de
  salida.**

**Se llaman avisos, pero no usan el prefijo `warning:`.** Ese prefijo es de stderr: la sección 4.3
dice que su tabla es la lista completa de avisos que el programa emite por ahí, y no hay ningún otro.
El informe de `doctor` va por stdout (ver más abajo por qué), así que para distinguir sus dos niveles
usa un formato propio, no un token que otra sección ya reserva para otra cosa: el encabezado de la
salida, `Errors:` y `Warnings:`, y el recuento de la primera línea. La palabra "aviso" sigue nombrando
el concepto; lo único que cambia es la marca literal.

**Un aviso tiene que ser accionable sin investigar nada**, o no sirve para lo que un agente necesita.
La regla: dice qué hay y qué se esperaba, con los dos valores literales al lado. Para la raíz adicional
que no se puede leer, tal como aparece bajo `Warnings:` en la salida de abajo:

```
extra board root "/Volumes/disco/boards" cannot be read (skipped when looking up boards by id)
```

Ahí el arreglo está a la vista, y se ve además la consecuencia: no es que algo esté roto, es que un
tablero que viviera ahí no se encontraría por su identificador mientras esa raíz no se pueda leer. Quien
lo lea sabe si le importa, montando el disco o quitando la raíz de la configuración de la máquina, sin
abrir nada más.

**El informe de `biso doctor` viaja entero por stdout, con sus errores y sus avisos juntos.** La
sección 4.2 dice que stdout lleva lo que un programa consumiría, y el informe es exactamente eso: es
el resultado que se ha pedido, no un mensaje que acompaña a otro trabajo. Los avisos y las notas de
la sección 4.3 son mensajes que un comando emite al lado de lo que produce; los hallazgos de `doctor`
son lo que produce, así que stdout le corresponde por la regla general, no aparte de ella. La sección
ya lo hacía así antes de esta tarea, cuando un tablero limpio imprime `no problems found` por stdout:
aquí no se cambia nada, se explica lo que ya era. Los avisos de los demás comandos, los que lista la
tabla cerrada de la sección 4.3, siguen yendo por stderr sin cambiar, y el informe de `doctor` no le
añade ninguna fila ni reutiliza su prefijo `warning:`.

#### Qué comprueba

| Comprobación | Nivel | Reparable con `--fix` |
|---|---|---|
| Identificadores duplicados | error | no, hay que decidir a mano |
| Tareas que no se pueden leer | error | no |
| Claves de extensión no declaradas | error | no |
| Estados, tipos, prioridades o proyectos que ya no están configurados | error | no |
| `initial_status`, `active_status` o `terminal_status` que no están en `statuses` | error | no |
| `statuses` con menos de tres elementos, o dos de los tres papeles apuntando al mismo estado | error | no |
| Dependencias que apuntan a tareas inexistentes | error | no |
| Ciclos de dependencias | error | no |
| Ciclos de tarea padre | error | no |
| Claves de criterio repetidas dentro de una tarea | error | no |
| `leaseExpiresAt` o `leaseHolder` con valor en una tarea que no está a la vez en el estado activo y asignada, o uno de los dos con valor y el otro vacío | error | sí, vaciando los dos |
| El identificador más alto que el tablero recuerda haber asignado (4.11) es menor que el identificador más alto de una tarea existente | error | sí |
| Falta el marcador `<id>.id` en el directorio del tablero (sección 3.2) | error | sí, escribiéndolo con el `id` que lleva la base de datos |
| El marcador `<id>.id` nombra un `id` distinto del que lleva la base de datos | error | no, hay que decidir a mano |
| Una raíz de `boards_extra_roots` (sección 3.3) no existe o no se puede leer | aviso | no, es configuración de la máquina o un disco sin montar |
| La comprobación de integridad de la base de datos falla (4.12) | error | no, es daño externo; la reparación es restaurar de una copia |
| El directorio del tablero está en un sistema de ficheros donde el modo WAL de SQLite no es seguro | aviso | no, es una propiedad del sistema de ficheros, no algo que `biso` pueda cambiar |
| Huecos en la numeración | no es un problema | no son un problema, no se reportan |

**No hay ninguna comprobación sobre el nombre de la carpeta del tablero, y no es un olvido.** El nombre
es decorativo y nadie resuelve por él (sección 3.2), así que una carpeta con el nombre de un `project_name`
anterior, o con un nombre que alguien puso a mano, no es un problema del que informar. Denunciarlo sería
denunciar algo que la sección 10.10 permite explícitamente.

**La fila del arrendamiento repara en una sola dirección, y por eso `--fix` la hace solo.** Los dos
campos son lo que sobra, y el estado y la lista de personas asignadas son el dato: vaciarlos deja la
tarea exactamente como estaba, mientras que arreglarla al revés, poniéndola activa o asignándole a
alguien para justificar el arrendamiento, cambiaría el trabajo del tablero para salvar una reserva que
ya no vale. La invariante que comprueba es la de la séptima precisión de la sección 5, la misma que
`biso new --from` aplica al importar (10.3), y cae en el segundo de los dos casos de arriba: ninguna
escritura de `biso` la puede romper, así que si un tablero llega a ese estado es por daño externo, como
una base de datos escrita a mano, restaurada a medias o venida de otra versión. Sin `--fix` sale bajo
`Errors:`:

```
  TASK-52  has a lease but is not both active and assigned
```

Y con `--fix`, en el grupo de lo reparado:

```
  TASK-52 had a lease but was not both active and assigned; cleared leaseExpiresAt and leaseHolder
```

Su `code` en el JSON es `lease_invariant` en los dos sitios.

**Las dos filas del marcador se parecen y se reparan al revés, y por eso son dos.** Que falte tiene una
sola lectura posible: el `id` de verdad es el que lleva la base de datos, y el marcador es su copia en el
sistema de ficheros, así que escribirlo con ese valor no puede equivocarse y `--fix` lo hace solo. Que
discrepe no tiene una sola lectura: reescribir el marcador con el `id` de la base de datos dejaría de
resolver a todos los punteros que nombran el `id` viejo, y reescribir la base de datos cambiaría la
identidad del tablero. Las dos direcciones pierden algo, así que se decide a mano, con el mismo criterio
que los identificadores duplicados de la primera fila.

**La comprobación de integridad de la base de datos y el aviso del sistema de ficheros son las dos
comprobaciones que añade esta misma decisión de persistencia**, y caen cada una en uno de los dos
casos de arriba. La primera es daño externo puro, sección 4.12: nada dentro de `biso` corrompe su
propia base de datos, así que no hay ninguna dirección que reparar por su cuenta, y por eso es un
error, no un aviso, aunque tampoco sea reparable. La segunda no reporta un daño ya hecho, sino un
riesgo: en ese sistema de ficheros el modo WAL de SQLite no ofrece las garantías de atomicidad que la
sección 4.10 exige, pero el tablero de hoy puede estar perfectamente sano. Eso es exactamente lo que
distingue a un aviso de un error, así que es aviso.

**Y la primera tiene una peculiaridad que la separa de las demás filas de error: nunca aparece como
una línea del informe.** Las otras catorce comprobaciones de error sí producen una entrada en la lista
de problemas cuando se disparan, pero esta no, porque cuando se dispara no hay informe de `doctor`
que mostrarla: hay el abort completo de la sección 4.12, con su propio mensaje y su propio código 10,
antes de que `doctor` llegue a comprobar nada más (ver la tabla de comportamiento más abajo). La fila
está en esta tabla para decir que existe como comprobación y cuál es su nivel, no porque vaya a
verse alguna vez junto a las demás.

Con esto, de las dieciocho filas de la tabla, diecisiete son problemas (quince errores y dos avisos) y
una, los huecos en la numeración, no lo es y no se reporta nunca. De esas diecisiete, solo dieciséis
llegan a aparecer alguna vez como una línea del informe: la comprobación de integridad de la base de
datos es la única que, aun siendo un problema real, no se manifiesta ahí, por la razón de arriba.

#### Atomicidad de `--fix` con varias reparaciones

Cuando `--fix` tiene que aplicar más de una reparación de tipo distinto, por ejemplo corregir el
contador del identificador más alto y escribir el marcador `<id>.id` que falta, no hay una sola
operación que las cubra a las dos: **las reparaciones de datos van en una sola transacción de la base
de datos, todo o nada, y la escritura del marcador va después y por separado.** Si esa escritura
falla, no deshace las reparaciones de datos que ya se aplicaron.

Esto no es una preferencia de diseño, es una imposibilidad: **escribir un fichero no puede estar dentro
de una transacción de SQLite.** Son dos sistemas distintos, el motor de la base de datos y el sistema de
ficheros, y no existe manera de hacerlos atómicos juntos. Cualquier redacción de esta sección que
prometiera una atomicidad conjunta estaría prometiendo algo que no se puede implementar. Es la única
reparación de `--fix` que sale de la base de datos, y por eso esta sección existe.

Y esto no rompe la garantía de la sección 4.10, porque esa sección promete sobre las escrituras del
tablero, es decir, sobre sus datos, y el marcador no es un dato del tablero: es una copia de su `id` en
el sistema de ficheros, puesta ahí para poder encontrarlo sin abrirlo (sección 3.2). Por eso el orden
importa y hay que decirlo explícito: primero la transacción de datos, después el marcador. Si el
marcador falla, las reparaciones de datos quedan hechas y son definitivas, el comando termina con el
código de no poder escribir (7, el mismo de cualquier otro fallo de entorno al reparar), y la falta del
marcador vuelve a aparecer como error la próxima vez que se ejecute `doctor`, porque sigue siendo verdad.
Ese desenlace es coherente consigo mismo: no hay ningún dato del tablero observado a medias, y lo único
que queda pendiente es una reparación que ya se sabe cómo repetir.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Tablero limpio | `no problems found` por stdout, código 0 |
| Solo avisos, sin ningún error | Se reportan bajo `Warnings:`, código 0 |
| Solo errores reparables, con `--fix` | Se reparan y se reporta cada uno, código 0 |
| Quedan errores sin reparar | Código 6, aunque se haya reparado algo o se hayan reportado avisos |
| Una tarea ilegible (4.12) | Se reporta como error y se sigue con las demás. **Nunca aborta** |
| La base de datos no se puede leer (4.12) | El comando entero aborta con el mensaje y el código 10 de 4.12, antes de comprobar nada más |
| `--fix` sin poder escribir | Código 7. Si falla la escritura del marcador después de la transacción de datos, esta ya quedó aplicada (ver arriba) |
| `--fix --dry-run` | Reporta qué se repararía, sin reparar nada, código 0 |

#### Salida

```
2 errors found, 1 warning found
Errors:
  TASK-40  dependency TASK-99 does not exist
Warnings:
  extra board root "/Volumes/disco/boards" cannot be read (skipped when looking up boards by id)
1 error fixed
  the highest recorded id was TASK-40 and tasks go up to TASK-52; recorded TASK-52
```

Los errores y los avisos se agrupan bajo su propio encabezado, `Errors:` y `Warnings:`; ninguno de
los dos usa el prefijo `warning:` de stderr, que la sección 4.3 reserva para lo que va por ahí. Un
grupo vacío no se imprime: si no hay avisos no aparece `Warnings:`, y si no hay errores no aparece
`Errors:`. `1 error fixed` cuenta lo reparado aparte, después de los dos grupos.

**El recuento de la primera línea es de lo que la comprobación encontró, esté reparado o no**, así que
no cambia según se haya pedido `--fix` o no: el mismo tablero dice `2 errors found` con `--fix` y sin
él. Lo que cambia con `--fix` es dónde sale cada error. **Un error reparado no se lista dos veces: sale
solo en el grupo de lo reparado, y desaparece de `Errors:`**, que es siempre la lista de lo que queda
por hacer, y por eso puede ser más corta que el número de la primera línea. La resta la explica la
línea de lo reparado, que por eso dice `error` y no `problem`: `2 errors found` arriba, uno bajo
`Errors:` y `1 error fixed` abajo cuadran a la vista sin que quien lee tenga que suponer nada. La
misma cuenta es la que sostiene el código de salida 0 de más abajo, "nothing wrong, or every error
found was fixed": lo encontrado y lo reparado se cuentan sobre el mismo conjunto.

Los avisos no tienen esa distinción porque ninguno es reparable (ver la tabla de comprobaciones), así
que `Warnings:` siempre los lista todos y su recuento siempre coincide con su lista.

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "doctor",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "problems": [
      { "task": "TASK-40", "code": "dependency_not_found", "message": "dependency TASK-99 does not exist" }
    ],
    "warnings": [
      { "task": null, "code": "extra_root_unreadable", "message": "extra board root \"/Volumes/disco/boards\" cannot be read (skipped when looking up boards by id)" }
    ],
    "fixed": [
      { "code": "highest_id_behind", "message": "the highest recorded id was TASK-40 and tasks go up to TASK-52; recorded TASK-52" }
    ]
  }
}
```

No hay ninguna clave de recuento: los tres números de la salida de texto se sacan de la longitud de
las tres listas, y hay que sacarlos igual que los saca el texto. **`problems` son los errores que
quedan, no todos los que se encontraron**, porque un error reparado se mueve a `fixed`, así que
`2 errors found` de la primera línea es `len(problems) + len(fixed)`, y `1 warning found` es
`len(warnings)`. Con esto un consumidor del JSON llega exactamente al mismo número que imprime el
texto, en vez de a uno menor.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Nada mal, o todo lo encontrado se ha reparado | 0 |
| Quedan errores | 6 |
| Sintaxis | 2 |
| No se puede escribir al reparar | 7 |
| No hay tablero | 8 |
| Su base de datos no se puede leer (4.12) | 10 |

#### `biso doctor --help`

```
Usage: biso doctor [options]

Check the board for duplicate ids, unreadable tasks, undeclared extension keys,
values that are no longer configured, a broken status-role invariant, broken
dependencies, dependency cycles, parent cycles, repeated criterion keys, a lease
on a task that is not both active and assigned, a recorded highest id that has
fallen behind, a database that fails its integrity check, a missing or
mismatched <id>.id marker, an extra board root that cannot be read, and a board
directory on a filesystem where SQLite's WAL mode is not safe.

Options:
      --fix      repair what can be repaired without a decision
  -h, --help     show this help

Without --fix this is a read-only command: --print and --dry-run are bad usage
here, same as in any other read-only command. With --fix, --dry-run reports
what would be fixed without fixing it.

Findings come in two levels: errors, which leave the board inconsistent or
unreliable, and warnings, which are true and worth knowing but fix nothing.
Only remaining errors produce exit code 6.

An unreadable task is reported and skipped, never a reason to stop. A database
that cannot be opened, or that fails its integrity check, is not a finding: the
whole command fails instead, with exit code 10.
Gaps in the id sequence are normal and are not reported.

Exit codes:
  0  nothing wrong, or every error found was fixed
  2  bad usage
  6  errors remain
  7  cannot write while fixing
  8  no board here
  10 its database could not be read

Examples:
  biso doctor
  biso doctor --fix
```

---

### 10.12. `biso board`

#### Firma

```
biso board [--port <n>] [--no-open]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--port <n>` | | no | entero entre 1024 y 65535 | 6420 | no | no | ninguno |
| `--no-open` | | no | booleano | falso | no | no | ninguno |

La interfaz interactiva, y **el único comando del programa que abre una interfaz**. Ningún otro puede
abrirla, ni la abre nadie por su cuenta.

| Caso | Qué pasa |
|---|---|
| No hay terminal | Error 7: `error: biso board needs a terminal; every other command works without one` |
| El puerto está ocupado | Error 7, diciendo el puerto |
| Puerto fuera de rango | Error 2 |
| `--no-open` | Arranca y solo imprime la dirección |
| El tablero cambia mientras está abierto | La interfaz recarga. Nunca muestra una versión en caché de una tarea que otro proceso ha cambiado |

#### Salida

```
Board at http://127.0.0.1:6420. Ctrl-C to stop.
```

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "board",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "url": "http://127.0.0.1:6420",
    "port": 6420,
    "opened": true
  }
}
```

**El sobre se imprime al arrancar, no al terminar**, que es la única diferencia de este comando con
todos los demás en JSON: los otros escriben su salida cuando han acabado, y este se queda corriendo
hasta que alguien lo pare, así que el sobre sale en cuanto el servidor escucha. Es lo que necesita un
guion que lo arranque con `--no-open` y tenga que saber a qué dirección apuntar. `opened` dice si se ha
abierto un navegador, y es `false` con `--no-open`.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| El servidor se ha parado limpiamente | 0 |
| Puerto fuera de rango, o banderas incompatibles | 2 |
| No hay terminal, o el puerto está ocupado | 7 |
| No hay tablero | 8 |

#### `biso board --help`

```
Usage: biso board [options]

Open the interactive board in a browser. This is the only command that opens
an interface: every other one prints text and exits, with or without a
terminal.

Options:
      --port <n>   port to listen on, 1024 to 65535 (default 6420)
      --no-open    print the address and do not open a browser
      --json       machine-readable envelope, printed when the server starts
  -h, --help       show this help

Exit codes:
  0  stopped cleanly
  2  bad usage
  7  no terminal, or the port is taken
  8  no board here

Examples:
  biso board
  biso board --port 7000 --no-open
```

---

### 10.13. `biso help`

#### Firma

```
biso help [<command> | all]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `<command>` | | no | nombre de comando, o `all` | ninguno | no | no | |

| Caso | Qué pasa |
|---|---|
| Sin argumento | Imprime la ayuda de primer nivel de la sección 11, igual que `biso --help` |
| Con un nombre de comando | Imprime la ayuda de ese comando, igual que `biso <cmd> --help` |
| Con `all` | Imprime la ayuda de primer nivel más la lista de los nueve comandos de administración, cada uno con su línea |
| Con un nombre que no existe | Error 4, con los tres nombres más parecidos |

`biso help` funciona **sin tablero**.

#### Salida de `biso help all`

Es la de la sección 11, seguida de:

```
Administration:
  init               create a board for this project
  where              say which board is in use and why
  archive <ref>      take a task off the board (there is no delete)
  export             dump the board as NDJSON that `biso new --from` reads back
  config             read and change the board configuration
  doctor             check the board, and repair what can be repaired
  board              open the interactive board
  help [cmd|all]     this
  snapshot           write snapshot.ndjson and board.json, then record them
```

#### El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "help",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "commands": [
      { "name": "prime", "summary": "print the session briefing" },
      { "name": "new", "summary": "create a task" }
    ]
  }
}
```

**En JSON, este comando emite la lista de comandos con su resumen de una línea, y nunca la ayuda en
prosa.** Un texto de ayuda es prosa escrita para leerse, con sus ejemplos y sus párrafos, y meterla en
una clave sería mover el problema de analizarla a otro sitio; la lista de comandos, en cambio, es un dato
y es lo que le sirve a un agente para descubrir la interfaz sin leer nada. Los resúmenes son los mismos
que imprime la ayuda de primer nivel (sección 11).

`biso help --json` y `biso help all --json` traen la lista entera, y `biso help <comando> --json` la trae
con un solo elemento, el de ese comando. La clave `commands` está siempre y siempre es una lista, así que
nadie tiene que mirar el argumento para saber qué forma va a recibir. La ayuda completa de un comando se
sigue pidiendo como siempre, sin `--json`, con `biso help <comando>` o `biso <comando> --help`.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Ayuda impresa | 0 |
| El comando no existe | 4 |

#### `biso help --help`

```
Usage: biso help [command|all]

Print the top-level help, or the help of one command, or the top-level help
plus the nine administrative commands with `all`. Works without a board.

With --json this prints the command list and its one-line summaries, never
the prose help: a help text is written to be read, and the list is the part
that is data.

Arguments:
  command        a command name, or `all`

Options:
      --json     machine-readable envelope with the command list

Exit codes:
  0  help printed
  4  no such command

Examples:
  biso help
  biso help finish
  biso help all
```

---

### 10.14. `biso snapshot`

#### Firma

```
biso snapshot [--vcs <mode>]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--vcs <mode>` | | no | `none`, `commit` o `push` | `commit` | no | no | ninguno |

**`biso snapshot` no cambia ningún dato del tablero**, así que la sección 3 lo clasifica junto a
`export` entre los comandos donde `--print` y `--dry-run` son error de uso con código 2: no hay
ninguna tarea afectada que imprimir, ni ninguna escritura de tarea que simular.

**No tiene bandera `-o`/`--out`.** A diferencia de `export`, que escribe donde se le diga,
`snapshot` escribe siempre en el propio directorio del tablero (10.1), con nombre fijo:
`snapshot.ndjson` y `board.json`. Es la instantánea del tablero para sí mismo, no un volcado a otra
parte; para volcar a otra parte está `export`.

#### Qué escribe, y por qué esos dos ficheros

`snapshot.ndjson` tiene exactamente la forma que fija la garantía de simetría de `biso export` (10.9):
una tarea por línea, con las mismas claves, incluidos los identificadores, las fechas y las claves de
criterio. `board.json` es la configuración del tablero, en la misma forma que imprime
`biso config list --json` (10.10): todas las claves de vocabulario, `task_prefix` y las demás.

**Con dos excepciones, `me` y `default_limit`, que no se escriben nunca.** Las dos son preferencias de
quien usa el tablero y no propiedades suyas, y la primera hace daño de verdad al viajar: restaurar la
instantánea de otra persona con `biso init --from` dejaría su identidad configurada como la del tablero,
y un tablero con `me` puesto anula `--mine`, porque entonces todo el mundo comparte identidad (sección
11 de `DECISIONES.md`). Un tablero restaurado nace por tanto sin identidad y con el límite por defecto,
y quien restaura pone la suya con `biso config set me`. La garantía de simetría de 10.9 sigue cubriendo
todo lo demás y **no cuenta estas dos claves**, que es la única cosa que exportar e importar no
reproduce campo a campo.

Los dos ficheros son los que después lee `biso init --from` (10.1) para reconstruir el tablero
entero. Van en dos ficheros separados, y no en uno solo, para que los diffs queden legibles: la
configuración cambia pocas veces y las tareas cambian todo el rato, así que mezclarlas habría hecho
que cada revisión de una tarea reescribiera también un bloque de configuración idéntico.

#### Cómo se escriben, y qué pasa si el segundo falla

**Cada fichero se escribe en un temporal del mismo directorio y se renombra encima del anterior**, que
es una operación atómica del sistema de ficheros, así que nadie lee nunca medio `snapshot.ndjson`.

**Y los dos temporales se escriben completos antes de renombrar ninguno de los dos.** De ese orden
depende lo único que se puede prometer aquí: si falla la escritura, por disco lleno o por permisos, los
dos ficheros anteriores quedan intactos, no se intenta guardar ninguna revisión, y el código es 7. Entre
el primer renombrado y el segundo queda una ventana de un instante en la que el par no es coherente, y
esta especificación lo dice en vez de prometer una atomicidad de dos ficheros que el sistema de ficheros
no da.

**`snapshot` no toma ningún acceso exclusivo**, ni el de las escrituras de tareas ni uno propio. Es una
lectura del tablero, y las lecturas nunca bloquean a nadie (4.10, punto 6): una instantánea de un
tablero grande no puede hacer fallar a un `biso set` que llegue a la vez. Lo que sí puede pasar es que
dos instantáneas simultáneas choquen al guardar la revisión, porque el sistema de control de versiones
se protege con su propia marca de bloqueo; entonces una de las dos sale con código 7 y su mensaje dice
que basta volver a llamar. Los ficheros de las dos quedan enteros de todos modos, porque cada una
escribió su temporal.

#### El sistema de control de versiones

El sistema lo dice la clave `vcs` de la configuración de máquina (sección 3.3), con `git` por defecto,
`none` para no ejecutar nada y `custom` para uno que `biso` no conoce. La bandera `--vcs` de este
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
`biso ls` y `biso prime` sobre un tablero de 300 tareas (sección 4.13) no admite ese coste en el camino
caliente de ningún comando. `biso snapshot` es la única excepción, precisamente porque quien lo llama ya
está pidiendo explícitamente esa operación; y es también, por lo mismo, el único comando que llega a
crear un repositorio: sin haber corrido nunca `biso snapshot` con `commit` o con `push` sobre él, el
directorio de un tablero nunca pasa a ser uno.

#### Dónde va la revisión, y cómo se decide

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
especificación no puede dejar a quien implemente adivinándola. Son las dos preguntas de arriba y tres
órdenes:

| Pregunta u orden | Qué se ejecuta, con el directorio del tablero como directorio de trabajo |
|---|---|
| ¿Hay un repositorio que lo contenga, y cuál es su raíz? | `git rev-parse --show-toplevel`, cuyo fallo significa que no hay ninguno |
| ¿Ese repositorio ignora la carpeta del tablero? | `git check-ignore` con la ruta del directorio |
| Crear el repositorio propio | `git init` |
| Guardar la revisión | `git add` con los tres ficheros nombrados por su ruta, y después `git commit` con el mensaje de abajo |
| Publicar | `git push` en el repositorio donde haya ido la revisión |

El identificador que la salida devuelve es el del commit que resulta, completo en el JSON y abreviado en
el texto, y el caso de "no hay nada que guardar" es el que reconoce el propio `git commit` cuando no hay
cambios que registrar.

**Preguntar por la exclusión es lo que separa los dos casos de un tablero que vive dentro del
proyecto**, y no se puede deducir mirando el sistema de ficheros: las dos situaciones tienen los mismos
ficheros en los mismos sitios, y lo único que las diferencia es una línea en el fichero de exclusión del
proyecto. Sin esa pregunta, un tablero ignorado acabaría con su revisión pedida al repositorio del
código, que no puede añadir lo que ignora.

**El tablero dentro del proyecto y versionado con él es el único caso en el que la instantánea cruza a
otra máquina sin que nadie configure nada**, porque el repositorio del proyecto ya tiene remoto. En los
demás, el historial del tablero es estrictamente local hasta que alguien le añada uno a mano, y entonces
`--vcs push` lo publica. Publicar no es trabajo de `biso` mientras no se le pida: la sección 14 lo dice.

**Y hay que decir qué arrastra `--vcs push` cuando la revisión ha ido al repositorio del proyecto**:
publica esa rama entera, así que se lleva también los commits de código que estuvieran pendientes.
Es lo que la bandera promete, y quien la escribe ya está pidiendo publicar, así que `biso` no se niega
ni publica a medias; lo que hace es no ponerla por defecto, que es el motivo de que el valor por defecto
sea `commit`.

**Con `custom` no hay ninguna de esas dos preguntas.** `biso` ejecuta la orden `commit` que declare la
configuración, en el directorio del tablero, y mira su código de salida; dónde acabe la revisión es cosa
de esa orden. Por eso una instantánea con `custom` no lleva identificador de revisión en su salida.

#### Qué entra en la revisión

Tres ficheros, nombrados uno a uno: **`snapshot.ndjson`, `board.json` y el marcador `<id>.id`.** No se
añade el directorio entero, para que un fichero que alguien deje ahí a mano no se cuele en el historial.

El marcador entra porque es lo que hace que la identidad del tablero viaje con la instantánea: al leerla
de vuelta, `biso init --from` adopta ese `id` en vez de acuñar uno nuevo, y por eso el puntero
commiteado del proyecto sigue valiendo después de restaurar (10.1). La base de datos no entra nunca, y
de eso se encarga el fichero de exclusión que `init` dejó escrito dentro del tablero, no este comando.

El mensaje de la revisión es `biso snapshot: 248 tasks`, con el recuento real de cada vez.

#### Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Hay repositorio y hay cambios desde la última instantánea | Escribe los dos ficheros, guarda la revisión, código 0 |
| Hay repositorio y no hay ningún cambio desde la última instantánea | Escribe los dos ficheros (con el mismo contenido de antes) y no hay nada que guardar; `note: nothing to commit, snapshot.ndjson and board.json are unchanged since the last snapshot`, código 0 |
| `--vcs none`, o la clave `vcs` en `none` | Escribe los dos ficheros y no ejecuta nada, código 0 |
| El directorio del tablero no está en ningún repositorio, y crearlo funciona | Lo crea, mete los tres ficheros en la primera revisión, código 0 |
| El directorio del tablero está dentro del repositorio del proyecto, que no lo ignora | Guarda la revisión ahí, con los tres ficheros nombrados por su ruta, código 0 |
| El sistema configurado no está instalado, o crear el repositorio falla | Escribe los dos ficheros, `note: no version control here, skipping the commit`, código 0 |
| La revisión falla por una razón de entorno, con un repositorio ya existente (sistema sin configurar, sin permiso, disco lleno, otra instantánea guardando a la vez) | Los dos ficheros ya han quedado escritos antes de intentarlo; Error 7, `code` `vcs_commit_failed`, y el mensaje dice que nada se ha perdido y que basta volver a llamar |
| `--vcs push` y la publicación falla | Los ficheros están escritos y la revisión guardada; Error 7, `code` `vcs_push_failed` |
| `--vcs push` con `vcs` igual a `custom` y sin orden `publish` declarada | Error 2, `code` `vcs_push_unavailable`, antes de escribir nada |
| Alguna tarea no se puede leer (4.12) | Se salta, se cuenta, `warning: 1 task could not be read and was skipped`, y el código es 6 en vez de 0, igual que en `biso export` |
| No se puede escribir alguno de los dos ficheros | Error 7, con los dos ficheros anteriores intactos y sin intentar la revisión |
| No hay tablero | Error 8 |

#### Salida

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
Committed a1b2c3d to the board's own repository
```

Cuando la revisión va al repositorio que contiene al tablero, la segunda línea lo nombra, porque es el
dato que distingue este caso y el que alguien querrá comprobar:

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
Committed a1b2c3d to /Users/avilches/Hub/Projects/Kex, the repository this board lives in
```

Con `--vcs push`, una tercera línea dice que se ha publicado. Sin nada que guardar, la primera línea por
stdout y la nota por stderr:

```
Snapshot written: snapshot.ndjson, board.json (248 tasks)
```

```
note: nothing to commit, snapshot.ndjson and board.json are unchanged since the last snapshot
```

Con `--vcs none`, o si no hay sistema instalado, la salida por stdout es solo la primera línea; en el
segundo caso, además, la nota `note: no version control here, skipping the commit` por stderr.

#### El esquema JSON

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
    "repository": "/Users/avilches/.biso/boards/kex-3f9a2b1c",
    "pushed": false,
    "skipped": []
  }
}
```

`vcs` es el sistema que se ha usado en esta llamada, y vale `none` cuando no se ha ejecutado nada.
`commit` y `repository` son `null` cuando `committed` es `false`, y `commit` también con `custom`, que
no devuelve identificador. `repository` es la raíz del repositorio donde ha ido la revisión, que es lo
que dice en qué caso de los tres se estaba. `pushed` es `false` salvo con `--vcs push` cumplido.
`skipped` lleva los identificadores de las tareas ilegibles que se han saltado, igual que en `biso ls`
(10.4): vacío salvo cuando el código de salida es 6.

#### Códigos de salida

| Desenlace | Código |
|---|---:|
| Escrito, y guardado si procedía | 0 |
| Sintaxis, o `--vcs push` sin orden de publicar configurada | 2 |
| Alguna tarea se ha saltado por ilegible | 6 |
| No se pueden escribir los ficheros, o falla la revisión o la publicación | 7 |
| No hay tablero | 8 |

#### `biso snapshot --help`

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

Exit codes:
  0  written, and recorded if that applied
  2  bad usage, or --vcs push with no publish command configured
  6  some task was skipped, unreadable
  7  cannot write there, or the commit or the push failed
  8  no board here

Examples:
  biso snapshot
  biso snapshot --vcs none
  biso snapshot --vcs push
```

---

## 11. La ayuda de primer nivel

`biso --help` y `biso help` imprimen esto, y solo esto:

```
biso 1.0.0 - the task board of this project.

Usage: biso [global options] <command> [options]

Start here:
  prime              everything you need to work on this board, in one message

Daily work:
  ls                 list tasks, most urgent first
  get <ref>          show one task
  new "TITLE"        create a task and print its id
  set <ref>...       change any field
  start <ref>...     take a task
  note <ref> TEXT    append an implementation note
  comment <ref> TEXT append a discussion comment
  finish <ref>...    close a task
  ask <ref> TEXT     park on a question
  answer <ref> TEXT  answer it and unpark

Global options:
  -C, --cwd <path>   resolve the board from there, instead of cd-ing
      --json         machine-readable output
  -q, --quiet        print only ids
      --print        print the whole record after writing
      --color <when> auto (default), always, never
      --dry-run      validate, write nothing (writing commands only)
  -V, --version      print the version
  -h, --help         this, or the help of a command

More: `biso <command> --help`, and `biso help all` for the administrative
commands (init, where, archive, export, config, doctor, board, help, snapshot).
```

Son treinta y una líneas, y no incluyen los nueve comandos de administración.

---

## 12. El contrato JSON

### 12.1. El sobre

Toda salida con `--json` es **un solo objeto JSON**, con o sin sangrado, indistintamente, y siempre
con esta forma:

```json
{ "schemaVersion": 1, "kind": "<tipo>", "generatedAt": "<ISO 8601 UTC>", "data": { } }
```

| `kind` | Lo produce | `data` contiene |
|---|---|---|
| `prime` | `prime` | Sección 9.9 |
| `where` | `where` | `id`, `board`, `path`, `source`, `me`, `counts`. Ejemplo en 10.2 |
| `init` | `init` | `board`, `pointerCreated`. Ejemplo en 10.1 |
| `task.list` | `ls` | `tasks`, `shown`, `matched`, `hidden`, `truncated`, `skipped`, `sort`, `filters` |
| `task.get` | `get` | `task` |
| `task.candidates` | `get` con varias coincidencias | `tasks` |
| `task.write` | `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`, `answer`, `archive` | `tasks`, `warnings` |
| `config` | `config list` | `config`. Ejemplo en 10.10 |
| `doctor` | `doctor` | `problems`, `warnings`, `fixed`. Ejemplo en 10.11 |
| `snapshot` | `snapshot` | `tasks`, `files`, `vcs`, `committed`, `commit`, `repository`, `pushed`, `skipped`. Ejemplo en 10.14 |
| `board` | `board` | `url`, `port`, `opened`. Ejemplo en 10.12, y se imprime al arrancar el servidor |
| `help` | `help` | `commands`, con el nombre y el resumen de cada uno. Ejemplo en 10.13 |
| `error` | cualquier fallo | Ver 12.2 |

Un lote de doscientas cuarenta y dos tareas es **un solo sobre** con doscientas cuarenta y dos
entradas en `data.tasks`, nunca doscientos cuarenta y dos objetos sueltos. La única salida del
programa que es una secuencia de objetos, uno por línea, es `biso export`, que no lleva sobre porque
su formato es NDJSON por definición.

### 12.2. Los errores en JSON

Con `--json`, un error sale **por stderr**, como un objeto con esta forma, y el código de salida del
proceso es el de la tabla de la sección 2:

```json
{
  "schemaVersion": 1,
  "kind": "error",
  "generatedAt": "2026-09-06T09:12:04Z",
  "error": {
    "exitCode": 3,
    "code": "unknown_status",
    "message": "unknown status: \"Pending\"",
    "field": "status",
    "given": "Pending",
    "valid": ["To Do", "In Progress", "Done"]
  }
}
```

Cuando un solo comando produce varios fallos, como un lote inválido, `error.details` es una lista de
objetos con la misma forma, uno por fallo, y `error.code` es `batch_invalid`.

**Un sobre de error no es una salida de datos, y se gobierna aparte de la promesa de la sección 12.4.**
Esa promesa existe porque quien consume una salida de datos no puede prever qué habrá dentro, así que
tiene derecho a que la forma no dependa del contenido. En un error sí puede preverlo, porque lo primero
que hace es leer `code`, y cada `code` trae siempre las mismas claves. Estas son las tres que están en
todos los errores y las cuatro de detalle, con la regla de cuándo acompañan:

| Clave | En qué errores aparece |
|---|---|
| `exitCode`, `code`, `message` | En todos, siempre |
| `field` y `given` | En los que nombran una bandera, una clave de configuración o un valor de entrada concreto: todos los del código 3, y los del 2 que nombran una bandera |
| `valid` | En los que rechazan un valor contra un conjunto conocido: los del 3 sobre vocabulario, y los del 2 sobre un dominio cerrado, como el modo de `--vcs` |
| `details` | Solo en `batch_invalid` y en `dry_run_failed`, y es una lista de objetos de esta misma forma, uno por fallo |

Las cuatro de detalle van juntas con su `code` y no con su código de salida, que es lo que hace la regla
comprobable: quien ramifica sobre `unknown_status` sabe que va a tener `field`, `given` y `valid`, y
quien ramifica sobre `busy` sabe que no va a tener ninguna de las cuatro.

### 12.3. Los identificadores de error

Un `code` estable es lo que permite ramificar sin analizar prosa. Esta es la lista de la versión 1.0,
agrupada por el código de salida con el que sale cada uno:

| Código de salida | `code` |
|---:|---|
| 2 | `incompatible_flags`, `duplicate_scalar_flag`, `unexpected_argument`, `missing_value`, `unknown_flag`, `unknown_command`, `missing_title`, `nothing_to_change`, `malformed_id`, `id_like_positional`, `inverted_range`, `key_selector_with_many_tasks`, `criterion_selector_overlap`, `two_stdin`, `read_only_flag`, `invalid_date`, `invalid_number`, `invalid_prefix`, `dependency_cycle`, `parent_cycle`, `self_dependency`, `board_exists`, `delete_not_supported`, `missing_identity`, `invalid_status_roles`, `unknown_status_role`, `too_few_statuses`, `invalid_snapshot_config`, `invalid_vcs_mode`, `vcs_push_unavailable` |
| 3 | `unknown_status`, `unknown_type`, `unknown_priority`, `unknown_project`, `unknown_label`, `unknown_assignee`, `unknown_milestone`, `unknown_extension_key`, `unknown_section`, `unknown_sort_field`, `ambiguous_vocabulary`, `empty_scalar_value`, `bad_config_value`, `undecodable_task`, `invalid_encoding` |
| 4 | `not_found`, `never_allocated`, `unknown_config_key`, `criterion_not_found`, `file_not_found` |
| 5 | `ambiguous_reference`, `criterion_ambiguous` |
| 6 | `already_finished`, `precondition_failed`, `board_inconsistent`, `doctor_problems`, `open_question_exists`, `no_open_question`, `mine_requires_identity` |
| 7 | `busy`, `io_error`, `file_unreadable`, `no_terminal`, `port_in_use`, `vcs_commit_failed`, `vcs_push_failed` |
| 8 | `no_board`, `pointer_unresolved` |
| 9 | `batch_invalid`, `dry_run_failed` |
| 10 | `database_unreadable` |
| 11 | `ambiguous_board_id` |
| 1 | `internal` |

**La lista es ampliable y las entradas son permanentes.** Una versión posterior puede añadir un `code`
nuevo, pero ninguno de los de arriba cambiará de significado, cambiará de código de salida ni
desaparecerá. Quien ramifique sobre un `code` desconocido debe tratarlo por su código de salida, que
sí está cerrado.

`missing_identity` (código 2, en `biso ask`, `biso answer` y el autor de un comentario) y
`mine_requires_identity` (código 6, en `--mine`) son la falta de identidad de la tabla de 3.1, pero
con dos códigos de salida distintos. No son un mismo concepto duplicado: como esta tabla está
agrupada por código de salida y ninguno de los dos se mueve nunca, la misma falta de identidad no
puede compartir un `code` cuando sale con códigos distintos. La asimetría entre los dos códigos es
anterior a esta rama.

### 12.4. Números, fechas y ausencias

- Las fechas son ISO 8601 en UTC terminadas en `Z`, con precisión de segundo. Nunca hora local, nunca
  sin zona. `due` es la excepción, porque es un día y no un instante, y viaja como `YYYY-MM-DD`.
- `urgency` es un decimal con un solo dígito tras el punto.
- Un campo sin valor es `null`, nunca la cadena vacía ni la ausencia de la clave. **Ninguna clave va ni
  viene según los datos**: la que está documentada para un `kind` aparece siempre que se emite ese
  `kind`, valga lo que valga, para que nadie tenga que distinguir entre "no está" y "no tiene valor".
- **La única excepción son las claves que gobierna una bandera**, y se sostiene porque quien llama sabe
  qué banderas ha escrito: no tiene que mirar la salida para averiguar qué va a encontrarse en ella. Lo
  que la regla de arriba prohíbe es lo otro, que la presencia de una clave dependa de los datos, que es
  justo lo que el consumidor no puede prever. Estas son todas las que hay en el documento:

  | Clave | `kind` | La bandera que la gobierna |
  |---|---|---|
  | `data.task.urgencyBreakdown` | `task.get` | Solo aparece con `--explain-urgency` (10.5) |
  | Las demás claves de `data.task` | `task.get` | Con `--section`, `data.task` trae solo `id` y las claves de las secciones pedidas, y ninguna otra (10.5) |
- Una lista vacía es `[]` y un mapa vacío es `{}`, nunca `null`.
- **Todo lo de arriba gobierna las salidas de datos, y el sobre de error se gobierna aparte** (12.2). La
  razón es que quien consume una salida de datos no puede prever qué habrá dentro, y quien recibe un
  error sí: lo primero que lee es `code`, y cada `code` trae siempre las mismas claves.

---

## 13. El contrato de estabilidad

Lo que se promete mientras la versión mayor sea `1`. **Obliga a partir de la versión 1.0**, que
todavía no está publicada: hasta que salga, nada de lo de abajo está roto por cambiar.

**No cambia nunca:**

- Los códigos de salida de la sección 2 y su significado.
- Los identificadores `code` de la sección 12.3, con la regla de ampliación que allí se dice.
- El nombre y el significado de cada comando y de cada bandera. **Una bandera nunca cambia de
  semántica**, y en particular ninguna que hoy añade pasará a reemplazar. Si hiciera falta el
  comportamiento contrario, se añade una bandera nueva con otro nombre.
- Las claves de `data` en cada `kind` de JSON. Se pueden añadir claves; las que hay no se quitan ni
  cambian de tipo.
- El algoritmo de coincidencia de la sección 6.1, idéntico al leer y al escribir.
- La simetría entre `biso export` y `biso new --from` sobre todos los campos no derivados, que es una
  prueba de la suite y no una intención. La de `biso snapshot` con `biso init --from` cubre además la
  configuración del tablero, **con la excepción declarada de `me` y `default_limit`**, que no viajan en
  la instantánea (10.14) y por tanto tampoco están en esta promesa.
- La estabilidad de las claves de los criterios: una clave asignada no se reasigna nunca.
- El tope de tamaño del mensaje de `biso prime`.

**Puede cambiar entre versiones menores, y por eso no hay que analizarlo:**

- El texto exacto de los mensajes de error y de los avisos. Lo estable es el `code`, no la prosa.
- La disposición de las columnas de `biso ls` y de la ficha de `biso get`, y para eso está `--json`.
- El texto de `biso prime`, dentro de su tope, que es donde se espera que la herramienta más aprenda
  con el tiempo.
- Los coeficientes por defecto de la urgencia. La estructura de la fórmula, no.
- Los valores por defecto de la configuración, salvo los que este documento fija dentro de un comando.
- **El presupuesto de arranque de 25 milisegundos de la sección 4.13.** No es de la misma naturaleza
  que el tope de bytes de arriba: los 5.120 bytes son una propiedad del texto, así que cualquiera los
  mide y siempre dan lo mismo, mientras que los 25 milisegundos son una propiedad de la máquina de
  referencia. Congelar en este contrato un número que depende del hardware haría que la herramienta
  incumpliera su propia promesa al ejecutarse en un ordenador más lento, sin que nadie hubiera cambiado
  una línea de código. Sigue siendo una prueba de la suite y sigue teniendo que fallar ante una
  regresión real: lo que no es, es una promesa de versión a versión.

**Cómo se anuncia una retirada.** Nada se quita sin un ciclo completo de aviso: primero la
funcionalidad emite `warning: <x> is deprecated and will be removed in 2.0` durante al menos una
versión menor, y solo entonces desaparece. Una funcionalidad nunca desaparece en silencio entre dos
versiones.

**Migración.** Si cambiara la forma en que los datos se guardan, la herramienta migra sola al
detectarlo, y en cualquier caso el volcado de una versión se puede importar en la siguiente, porque
el formato de `export` es el de `new --from` y los dos están en este contrato.

---

## 14. Lo que se deja fuera a propósito

Nombrar lo que no está evita que alguien lo dé por olvidado.

- **No hay `biso delete`.** Está especificado que no existe y qué contesta si se intenta (10.8).
- **No hay entidades de hito, documento ni decisión.** El hito es un campo de la tarea y no una
  entidad con ciclo de vida propio: no se crea, no se cierra, no tiene fecha ni descripción, y no hay
  ninguna clave de configuración que lo declare. Eso no impide que `-m/--milestone` valide al filtrar,
  porque el conjunto contra el que valida es derivado de lo que las tareas usan (6.3) y no una lista
  que haya que mantener aparte. La documentación se apunta con `--doc`, que es una lista de textos.
- **No hay contextos de sesión**, es decir, filtros por defecto guardados que cambien lo que devuelve
  una consulta sin que se vea en la línea de comandos.
- **No hay recurrencia, ni seguimiento de tiempo, ni subtareas con numeración propia.** Una subtarea
  es una tarea normal con `--parent`, y el mensaje de error de un identificador como `TASK-1.1` lo
  dice.
- **No hay servidor de integración ni protocolo de herramientas.** La interfaz de la versión 1.0 es
  esta línea de comandos y su salida JSON.
- **No hay ninguna bandera ni variable de entorno que nombre un tablero.** El tablero se elige por
  las dos vías de la sección 3.2, y `-C` ya alcanza tanto el directorio de un tablero como el de un
  proyecto que apunte a uno, así que una bandera para nombrarlo no añadiría nada (sección 3 de
  `docs/DECISIONES.md`).
- **No hay una interfaz multiproyecto.** Cada invocación resuelve un único tablero (sección 3.2), y no
  hay ningún comando que lea o agregue varios a la vez, aunque la máquina entera tenga más de uno
  (sección 3.3): quien necesite verlos juntos los recorre uno por uno desde fuera.
- **No hay exportación al formato de Backlog.md.** `biso export` escribe el mismo formato que lee
  `biso new --from`, y traducir a un formato ajeno es trabajo de un conversor aparte, no de este
  comando.
- **No hay sincronización con ningún sistema externo.**
- **No hay sincronización entre máquinas.** Un tablero vive en la máquina donde se creó, y lo que
  cruza a otra es la instantánea que deja `biso snapshot`, para reconstruirlo entero con
  `biso init --from`, no para mantener dos copias vivas al día (sección 12 de `docs/DECISIONES.md`).
  **Cómo cruza depende de dónde viva el tablero, y las dos vías están especificadas** (10.14): un
  tablero versionado dentro del proyecto viaja con él y con el remoto que el proyecto ya tenga, sin que
  nadie configure nada, y un tablero con su propio repositorio viaja cuando alguien le da un remoto y
  `biso snapshot --vcs push` lo publica. Lo que no hay es ningún remoto que `biso` configure por su
  cuenta, ni forma de fusionar dos almacenes escritos por separado.
- **No hay un papel de estado para descartar, distinto de terminar.** Una tarea hecha y una abandonada
  hoy comparten el mismo estado terminal. Un papel que obligara a dar un motivo al entrar en él sería
  barato de añadir cuando hiciera falta, pero a diferencia de los demás requisitos de este modelo no
  trae ningún caso real en el que la confusión haya costado algo, así que se queda fuera hasta que
  aparezca uno.

---

## 15. Por dónde empezar a implementar

En el orden en que cada pieza paga lo que cuesta:

1. **El almacén**: abrir o crear la base de datos, el modo WAL y la transacción dentro de la que
   ocurre cualquier escritura, que es de lo que dependen las seis garantías de la sección 4.10. **Aquí
   se mide el presupuesto de arranque de la sección 4.13** con el mecanismo de acceso que se haya
   elegido, porque esa prueba no se puede escribir antes de que el almacén exista, y la elección del
   mecanismo se decide midiendo contra ella.
2. **El modelo de datos lógico** de la sección 5, con las claves estables de los criterios y el
   rechazo explícito de lo desconocido.
3. **El algoritmo de coincidencia** de la sección 6.1, que es una función pura de veinte líneas y de
   la que dependen todos los comandos.
4. **`init` y `where`**, que son crear un tablero y saber cuál es. Van aquí y no al final porque no hay
   forma de ejecutar ni de probar el paso siguiente sin un tablero, y crear un tablero es `init`.
5. **`new`, `ls`, `get` y `set`**, que son el trabajo diario.
6. **Los seis verbos de ciclo** de 10.7, `start`, `note`, `comment`, `finish`, `ask` y `answer`, que
   son azúcar sobre `set` y se escriben encima.
7. **`prime`**, que es lo que hace que todo lo anterior se use bien sin leer nada más.
8. **El lote de `new --from`, `export`, `snapshot` e `init --from`**, los cuatro juntos porque la prueba
   de simetría los necesita a la vez: exportar un tablero e importarlo tiene que dar dos tableros
   idénticos campo a campo, y lo mismo una instantánea restaurada.
9. **El resto**: `archive`, `config`, `doctor`, `board` y `help`.

Las garantías de la sección 4.10 no son un paso de esta lista: hay que respetarlas desde el primer
comando que escriba. El paso 1 es el que da las herramientas para respetarlas, no una excepción a esa
regla.
