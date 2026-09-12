<!--
  Generated file. Do not edit by hand: it is overwritten entirely every time
  tutorial/generate.py runs.
  Regenerate it with: uv run --with-requirements docs-requirements.txt --no-project python tutorial/generate.py
-->

# biso tutorial, by scenario

!!! warning "Generated document"
    This page is generated automatically from the fixtures in
    `tutorial/escenarios/` and from `tutorial/conceptos.md`. Do not edit it by
    hand: any change is lost on the next generation. To regenerate it:

    ```
    uv run --with-requirements docs-requirements.txt --no-project python tutorial/generate.py
    ```

## Conceptos

`biso` es la herramienta con la que un agente automático, y la persona que trabaja con él, llevan
las tareas de un proyecto. Antes de ver un solo comando conviene tener claros cinco conceptos, porque
el resto del tutorial da por hecho que ya los conoces.

### El tablero

Un tablero es el conjunto de tareas de un proyecto junto con su configuración: qué estados existen,
qué tipos de tarea hay, qué prioridades se pueden usar y quién es, si alguien lo es, la identidad de
quien trabaja en él. Esa configuración no es un detalle aparte: es lo que hace que un valor tenga
sentido o no. Un tablero no acepta cualquier texto en cualquier campo, sino solo los valores de su
propio vocabulario.

De ahí sale la regla que gobierna todo lo demás: un valor que el tablero no reconoce es siempre un
error, tanto si se está escribiendo como si se está preguntando por él, y con la misma exigencia en
los dos casos. Preguntar por un estado que ese tablero no tiene no devuelve una lista vacía, devuelve
un error. Como consecuencia, cuando una pregunta sí devuelve una lista vacía es porque de verdad no
hay nada que cumpla lo pedido, y eso es información, no un fallo silencioso.

### Los estados

Cada tarea tiene un estado, y el estado es uno de los que ese tablero tiene configurados: no hay una
lista fija de estados válida para todos los proyectos, cada tablero declara los suyos. Lo único que no
se puede configurar libremente es que, de entre esos estados, tres papeles queden siempre cubiertos,
cada uno por un estado distinto:

- El estado **inicial**, donde nace una tarea nueva.
- El estado **activo**, el que se escribe cuando alguien coge una tarea para trabajar en ella.
- El estado **terminal**, el que marca que una tarea está acabada.

Un papel es la función que cumple un estado, no su nombre. Dos tableros pueden llamar de forma
distinta al estado inicial y aun así los dos tienen uno, porque el papel es obligatorio aunque la
palabra con la que se llama no lo sea.

### Criterios de aceptación y definición de hecho

Una tarea puede llevar dos listas de comprobación independientes, con la misma forma: cada elemento
tiene un texto y puede estar marcado o no, y cada elemento nace con una clave numérica propia que no
cambia aunque se quiten otros elementos de la lista. Son dos listas, no una, y se llevan la cuenta por
separado.

La primera son los **criterios de aceptación**: cómo se sabe que el trabajo de esa tarea en concreto
funciona. La segunda es la **definición de hecho**: lo que tiene que ser cierto antes de dar la tarea
por cerrada, más allá de si el resultado funciona, como que otra persona la haya revisado. Ninguna de
las dos es obligatoria, y una tarea puede llegar a su estado terminal con elementos sin marcar en
cualquiera de las dos, porque `biso` avisa de eso pero no lo impide por defecto.

### La urgencia

La urgencia es un número que no se guarda en ningún sitio: se calcula cada vez que se lee la tarea, a
partir de datos que sí están guardados, como la prioridad, si la tarea está activa, si bloquea o la
bloquea otra tarea, si tiene fecha límite cercana, si tiene criterios de aceptación y cuánto tiempo
lleva abierta. Por eso se dice que la urgencia es **derivada**: nadie la escribe directamente, y su
valor de ahora mismo puede no ser el de dentro de un minuto, aunque nadie haya tocado la tarea,
sencillamente porque el tiempo pasó o porque cambió alguna otra tarea de la que depende. Una tarea en
su estado terminal tiene urgencia cero siempre, sin más cálculo.

### La identidad declarada

Un tablero puede tener configurada una identidad, el texto que dice quién eres tú cuando llamas a
`biso`. No es obligatoria: un tablero funciona perfectamente sin que nadie la declare. Lo que cambia
es que algunas operaciones necesitan saber quién eres para tener sentido, como pedir solo las tareas
que son tuyas, quedarte automáticamente asignado una tarea al empezarla, firmar un comentario con tu
nombre por defecto, o dejar una pregunta abierta a la espera de que alguien responda. Si intentas
cualquiera de esas cosas sin identidad declarada, `biso` no adivina ni asume nada: lo dice como un
error explícito. El resto del tablero, lo que no necesita saber quién eres, sigue funcionando igual
con identidad declarada o sin ella.

## Scenarios

1. [Llegas a un proyecto que no conoces](#escenario-01)
2. [Se te ocurre algo y no quieres que se pierda](#escenario-02)
3. [¿Y ahora qué hago?](#escenario-03)
4. [Repartes el trabajo](#escenario-04)
5. [Te pones con ella](#escenario-05)
6. [Dos a la vez sobre el mismo tablero](#escenario-06)
7. [A mitad, falta un criterio](#escenario-07)
8. [Te atascas y alguien tiene que decidir](#escenario-08)
9. [La persona y el agente hablan](#escenario-09)
10. [Cierras](#escenario-10)
11. [Te equivocas](#escenario-11)
12. [Aparcas algo a medias](#escenario-12)
13. [Tocas veinte de golpe](#escenario-13)

## 1. Llegas a un proyecto que no conoces {: #escenario-01 }

Te acaban de meter en un proyecto que no habías tocado nunca. Hay gente trabajando en él
desde hace tiempo, hay tareas a medias, y tú no sabes ni por dónde se entra. Antes de tocar
nada, quieres saber qué hay: qué se está haciendo ahora mismo, qué espera respuesta, qué te
toca a ti y qué reglas rigen este tablero en concreto. No quieres leerte un manual entero
para eso.

!!! abstract "What this scenario teaches"
    - El arranque se lee entero una vez, al principio de la sesión, y con eso basta, sin volver a consultar nada más para completar un ciclo de trabajo completo.
    - `biso where` existe para el día en que dudes de qué tablero estás tocando de verdad.

Lo primero que haces, antes de mirar nada más, es esto:

```console
$ biso prime
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

Exit code: `0`

*Note: Esto es todo el mensaje, de arriba a abajo, y no hay una segunda pantalla después. El bloque BOARD dice el vocabulario real de este tablero (los tipos, las prioridades, los tres estados con su papel) y quién eres tú aquí. Las cuatro secciones de abajo reparten el tablero entero sin que ninguna tarea aparezca en dos sitios: lo que está en marcha, lo que espera una respuesta tuya, lo que es tuyo y lo siguiente por urgencia. Nadie escribe esa urgencia a mano, se recalcula cada vez que la pides (lo retomas en el escenario 3, "¿Y ahora qué hago?"). Con esto ya podrías crear una tarea, empezarla, trabajarla y cerrarla sin abrir ningún otro documento: ese es el criterio con el que está escrito este mensaje.*

Aún así te queda una duda que `prime` no contesta: este tablero, el que acabas de leer,
¿es exactamente el que crees que es? Si trabajas con varias copias del proyecto a la vez
(un worktree distinto, una carpeta clonada dos veces) conviene comprobarlo antes de
escribir nada.

```console
$ biso where
id       3f9a2b1c
board    Kex
path     /Users/avilches/.biso/boards/kex-3f9a2b1c
source   project pointer at /Users/avilches/Hub/Projects/Kex
me       @claude
tasks    248 not archived, 31 archived, highest id ever assigned TASK-290
```

Exit code: `0`

*Note: Cuatro datos que cambian cada uno por su cuenta: el identificador no cambia jamás, el nombre lo cambia `biso config set project_name`, la ruta cambia si alguien mueve la carpeta a mano, y `source` dice de dónde ha salido esa respuesta, que aquí es el puntero del proyecto (sección 3.2). Si no hubiera ningún tablero, tanto `prime` como `where` fallarían con el código 8, y el propio mensaje de error de `where` remite a `biso init` como el remedio: crea un tablero nuevo y lo deja apuntado desde este proyecto. No hace falta más que saber que existe para este momento; sus banderas se ven cuando de verdad haga falta crear uno.*

## 2. Se te ocurre algo y no quieres que se pierda {: #escenario-02 }

Estás a mitad de otra cosa cuando caes en algo que hay que arreglar en el subidor de
archivos. No es lo que estás haciendo ahora, y si sigues tirando del hilo pierdes el hilo
de lo de verdad. Lo único que necesitas es dejarlo apuntado en algún sitio donde no se te
vaya a olvidar, y seguir.

!!! abstract "What this scenario teaches"
    - El nombre desnudo de una bandera añade. `--ac` repetida dos veces deja dos criterios, no uno que pisa al otro.
    - `biso new` imprime el identificador de la tarea nueva y nada más: ni un título repetido a modo de confirmación, ni una línea de estado.

Bastaría con teclear el título y ya quedaría a salvo, pero ya que lo tienes en la
cabeza, merece la pena dejar también el porqué y con qué criterio sabrás que está
resuelto. Son dos banderas más en la misma línea, y `new` las acepta todas:

```console
$ biso new "Log the diff size before uploading" --ac "The log prints the diff size before each upload" --ac "The size also shows up when the upload fails"
TASK-62
```

Exit code: `0`

*(derived output, see [`biso new`](spec/cmd/new.md); not literal spec text)*

*Note: Nada de "created TASK-62" ni de repetir el título: la única línea es el identificador, porque eso es justo lo que no podías saber de antemano y lo demás ya lo sabías (regla 7 del mensaje de arranque, principio 4 de la sección 1). Las dos banderas `--ac` no han sustituido nada, han añadido dos criterios distintos a una lista que empezaba vacía, porque el nombre desnudo de un campo de lista siempre añade (sección 8.1). Si en vez de sumar un segundo criterio hubieras querido reemplazar el primero, la bandera habría sido `--set-ac`, no `--ac` otra vez, aunque aquí no hacía falta: una tarea recién creada no tiene nada que sustituir todavía.*

## 3. ¿Y ahora qué hago? {: #escenario-03 }

Ya sabes qué tablero es este y ya has dejado tu idea a salvo. Pero el arranque te ha
enseñado el mapa entero del proyecto, no lo tuyo en concreto, y tienes veinte minutos
antes de la siguiente reunión. Quieres ver qué tienes tú encima ahora mismo, y luego
coger una y leerla entera antes de tocar nada.

!!! abstract "What this scenario teaches"
    - `biso ls` enseña como mucho 30 tareas y deja fuera las que ya están Done; lo que se queda fuera lo dice por stderr, nunca mezclado con las filas de datos de stdout.
    - La urgencia que decide el orden no la escribe nadie, se recalcula cada vez que pides la lista, y por eso nunca aparece como columna.

Miras qué tienes asignado, sin filtrar nada más:

```console
$ biso ls --mine
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md); not literal spec text)*

*Note: Cinco filas, ni una más: como son menos de las 30 que `ls` enseña por defecto, no sale ningún aviso de recorte por stderr, solo las filas por stdout. El orden no es el del identificador ni el de cuándo se crearon: es urgencia descendente, que es el criterio por defecto cuando ninguna tarea de la lista lleva un `ordinal` manual (sección 10.4). TASK-11 va primera porque suma prioridad alta, estar activa y bloquear a otra tarea sin terminar; TASK-40 va la última pese a estar también en marcha, porque a ella la bloquea TASK-11 y eso resta en vez de sumar. Ese número no lo escribe nadie ni se guarda en ningún sitio: se recalcula en el instante de leer (sección 5.4), y por eso no hay una columna "urgency" en esta tabla, solo el orden que produce.*

En un tablero real esa lista no siempre cabe en 30 filas, y el aviso de lo que se ha
quedado fuera no se mezcla nunca con las tareas: va por stderr, para que redirigir la
salida a un fichero deje un fichero de datos limpio. Para verlo sin esperar a que el
tablero crezca, le pones un límite más bajo que el que tienes asignado:

```console
$ biso ls --mine --limit 3
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
warning: 2 more tasks match; showing 3 of 5
hint: narrow with -s, --type or -l, or ask for everything with --all
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md); not literal spec text)*

*Note: Las tres primeras filas van por stdout, y la advertencia y la sugerencia van por stderr: en tu terminal se ven seguidas, pero si guardas la salida en un fichero (`biso ls --mine --limit 3 > tareas.txt`) ese fichero tiene tres líneas, no cinco. Aquí el límite lo has puesto tú con `--limit 3` para verlo con pocos datos; sin esa bandera el límite por defecto es 30, y con un tablero de este tamaño (54 To Do más 4 In Progress, sin contar las 190 Done que `ls` esconde por defecto) es fácil llegar a esas 30 sin buscarlo. `--all` quita el límite entero, y `--any-status` es lo único que trae de vuelta las tareas terminadas.*

De las cinco, TASK-33 es la que más se parece a lo que acabas de dejar tú en el
escenario anterior: apuntada, con algún criterio ya pensado pero sin empezar. Antes de
decidir si es tuya de verdad, la lees entera:

```console
$ biso get TASK-33
TASK-33  Add a retry counter to the upload log
status     To Do                type       task
priority   medium               urgency    4.3
assignees  @claude              reporter   @avilches
labels     -                    milestone  -
parent     -                    due        -
project    -                    ordinal    -
created    2026-08-20 09:40     updated    2026-09-03 16:15
depends    -                    blocks     -
refs       -
docs       -
files      -
ext        -

## Description
Every failed upload retries silently. The log does not say how many attempts it took,
so debugging one means reconstructing it by hand from scattered timestamps.

## Acceptance Criteria
- [x] #1 Every retry is counted in the log
- [ ] #2 The counter shows up in the final summary
- [ ] #3 A test covers three retries in a row

## Definition of Done
(empty)

## Implementation Plan
(empty)

## Implementation Notes
(empty)

## Final Summary
(empty)

## Comments
(empty)

## Open Question
(empty)
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: No hay línea `lease`: esa solo sale cuando la tarea tiene un arrendamiento, y una tarea en To Do no puede tenerlo (sección 5). Las nueve secciones salen siempre, aunque estén vacías y marcadas con `(empty)`, porque omitir una que no se ha pedido se confundiría con una que sí se ha pedido y ha salido vacía. Fíjate en las claves `#1`, `#2` y `#3` delante de cada criterio: son estables, no una posición en la lista, así que si algún día se quita el `#2` los otros dos siguen siendo `#1` y `#3`, nunca se renumeran.*

## 4. Repartes el trabajo {: #escenario-04 }

Hasta ahora has estado mirando el tablero y apuntando cosas. Esto es distinto: hay una tarea
que alguien tiene que hacer, y hay que decir quién. En este proyecto trabajan una persona,
`@sara`, y un agente, `@claude`, sobre el mismo tablero, y `TASK-19` (reintentar la subida
cuando el servidor responde 5xx) lleva ahí desde el principio sin dueño.

Sara decide que la haga el agente. Lo que sigue es cómo se dice eso, y cómo se entera él.

!!! abstract "What this scenario teaches"
    - Asignar es la decisión de una persona sobre lo que otra debe hacer. No es un reparto automático: `biso` no asigna nada por su cuenta.
    - Un tablero compartido entre una persona y un agente deja la clave `me` sin configurar a propósito, y cada uno declara su identidad en su propia sesión con `BISO_ME`. Si el tablero la llevase puesta, todos compartirían identidad y `--mine` dejaría de significar nada.
    - Asignar una tarea no crea ningún arrendamiento. Recibir trabajo y empezarlo son dos cosas distintas, y la segunda tiene su propio comando.

Sara abre una terminal nueva y, antes de nada, quiere ver qué tiene ella pendiente. Todavía
no ha declarado quién es en esta sesión.

```console
$ biso ls --mine
error: --mine needs an identity; set it with biso config set me <you> or BISO_ME
```

Exit code: `6`

*Note: Fíjate en que no devuelve una lista vacía. Podría haberlo hecho, y sería lo cómodo de programar: sin identidad, ninguna tarea es "mía", así que cero resultados. Pero entonces quien lo llama no podría distinguir "no tengo nada asignado" de "no has dicho quién eres", y eso es justo lo que el primer principio de la especificación prohíbe. El error tiene además su propio código, el 6, así que un programa que llame a `biso` puede reaccionar sin leer el mensaje. El mensaje ofrece las dos salidas, y no son equivalentes. `biso config set me` lo escribe en la configuración del tablero, donde lo verían todos; `BISO_ME` lo declara solo para esta sesión. En un tablero compartido hay que usar la segunda.*

Sara declara su identidad para esta sesión, que es una variable de entorno y no un comando
de `biso`, y le pasa la tarea al agente.

```console
$ biso set TASK-19 -a @claude
TASK-19  To Do  ac 0/2  urgency 7.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: La salida no repite lo que Sara acaba de escribir: no dice "assignee: @claude". Dice el estado en que queda la tarea, su avance de criterios y su urgencia, que es lo que ella no sabía. Es el cuarto principio de la especificación, y se cumple en todas las escrituras. La urgencia es 7.0, y sale de sumar 6.0 por ser de prioridad alta y 1.0 por tener criterios de aceptación. No suma el término de tarea activa, que vale 4.0, porque sigue en `To Do`: asignar una tarea no la pone en marcha. No aparece el trozo `dod` porque `TASK-19` no tiene definición de hecho, y ese trozo solo sale cuando la hay.*

En su propia sesión, el agente pregunta qué le toca. Aquí `--mine` sí funciona, porque su
identidad está declarada.

```console
$ biso ls --mine
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-19  To Do        task  high    Retry the upload on 5xx                           ac 0/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-61  To Do        docs  medium  Rewrite the install section                       ac 0/1  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Seis filas donde antes había cinco. `TASK-19` entra en segunda posición, y ese sitio no es casual: la lista va por urgencia descendente y `TASK-19` vale 7.0, igual que `TASK-60`. Un empate se rompe siempre por identificador ascendente, y 19 va antes que 60. Que `TASK-60` valga lo mismo que una tarea que ni ha empezado, teniendo ella prioridad alta y estando en curso, tiene una explicación: `TASK-60` tiene una pregunta abierta, y el término de tarea activa solo suma cuando la tarea está en el estado activo **y** no está esperando respuesta. Una tarea aparcada no está siendo trabajada por nadie, así que no compite por tu atención. En el escenario 8 se ve por dentro. Y ojo con lo que esta lista no dice: `TASK-19` ya es del agente, pero nadie ha empezado a trabajar en ella. No hay ningún arrendamiento. Para eso hace falta el comando del escenario siguiente.*

El mensaje de arranque cuenta lo mismo de otra forma. Si el agente volviera a lanzarlo
ahora, `TASK-19` habría cambiado de sitio.

```console
$ biso ls --mine -s "To Do"
TASK-19  To Do  task  high    Retry the upload on 5xx                ac 0/2  @claude  -
TASK-61  To Do  docs  medium  Rewrite the install section            ac 0/1  @claude  -
TASK-33  To Do  task  medium  Add a retry counter to the upload log  ac 1/3  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md); not literal spec text)*

*Note: En `biso prime`, estas tres son las que salen en el bloque `ASSIGNED TO YOU`, que significa exactamente esto: tareas que una persona decidió que hicieras tú y que todavía no están en marcha. En el escenario 1 viste ese mensaje entero, y entonces `TASK-19` no estaba en ese bloque porque no era de nadie. Fíjate también en que las columnas se han estrechado. El ancho de cada una lo decide el contenido de la lista que se está imprimiendo, no una plantilla fija: al filtrar por un solo estado, la columna de estado ya no necesita hueco para `In Progress`, y al desaparecer las tareas de título largo, la columna del título se encoge con ellas.*

## 5. Te pones con ella {: #escenario-05 }

El agente ya sabe que `TASK-19` es suya. Ahora va a empezarla de verdad, y eso no consiste en
cambiarle el estado a mano: hay un comando que se llama justo como el gesto que estás haciendo.

Con una sola llamada pasan cuatro cosas a la vez, y una de ellas no la has pedido.

!!! abstract "What this scenario teaches"
    - Empezar una tarea es un comando con nombre propio, `biso start`, y cuesta una llamada. No es `biso set --status`, aunque el estado acabe igual.
    - Al empezar nace un arrendamiento: una reserva con fecha de caducidad que dice que alguien está trabajando en esa tarea ahora mismo. Ningún otro comando lo crea.
    - Un arrendamiento no es un estado. La tarea está en `In Progress` y además tiene una reserva; son dos cosas separadas que se guardan aparte.

El agente coge la tarea, y de paso deja escrito por dónde piensa atacarla, para que Sara
pueda leerlo sin preguntar.

```console
$ biso start TASK-19 --plan "Reproducir el 5xx con un servidor de prueba, y envolver la subida en un reintento con espera exponencial"
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso start`](spec/cmd/verbos-del-ciclo.md#biso-start), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: La urgencia ha subido de 7.0 a 11.0, y son exactamente los 4.0 del término de tarea activa. Nadie ha escrito ese número: se recalcula al leer, cada vez. Lo que la salida no dice es lo más importante de este paso. `biso start` ha hecho cuatro cosas en una sola escritura: ha puesto el estado activo, ha comprobado que la tarea ya tenía a alguien asignado (a `@claude`, desde el escenario 4, así que no ha tocado nada), **ha tomado el arrendamiento** a nombre de quien llama, y ha añadido el plan. Si la tarea no hubiera tenido a nadie asignado, `start` le habría puesto a quien llama, y esa es la diferencia con lo que hizo Sara en el escenario anterior: ella asignó sin empezar, y `start` empieza asignando si hace falta.*

El arrendamiento no se ve en la línea de estado, así que el agente mira la ficha para
entender qué acaba de reservar.

```console
$ biso get TASK-19 --section plan
TASK-19  Retry the upload on 5xx

## Implementation Plan
Reproducir el 5xx con un servidor de prueba, y envolver la subida en un reintento con espera
exponencial
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: El plan está donde debía. Lo del arrendamiento hay que explicarlo, porque no se enseña aquí: son dos campos guardados en la tarea, hasta cuándo vale la reserva y de quién es, y el programa acaba de ponerlos en "ahora más 240 minutos" y en `@claude`. Esos 240 minutos son la clave `lease_minutes` de la configuración del tablero, y se pueden cambiar cuando se quiera sin romper nada: el vencimiento se calcula al escribir, así que cambiarla solo afecta a los arrendamientos que se renueven desde ese momento. El valor por defecto es generoso a propósito. El error que hace daño es el falso vencido, no el vencido tarde: un arrendamiento demasiado corto hace que otro agente reclame una tarea que alguien está trabajando de verdad, mientras que uno demasiado largo solo retrasa un aviso. Y como vas a ver en el escenario siguiente, ese aviso nunca impide trabajar.*

## 6. Dos a la vez sobre el mismo tablero {: #escenario-06 }

Este es el escenario que más reglas enseña por línea, y ninguna de ellas se adivina.

Sara y el agente trabajan sobre el mismo tablero al mismo tiempo. Nada les impide escribir sobre
la misma tarea, y `biso` no va a impedírselo: lo que hace es dejar constancia de quién dijo que
estaba trabajando en qué, para que el otro lo sepa antes de duplicar el esfuerzo.

Eso es el arrendamiento, y hasta ahora lo has visto nacer sin mirarlo de frente.

!!! abstract "What this scenario teaches"
    - Un arrendamiento no es un estado, es hasta cuándo vale una reserva. La tarea puede estar en el estado activo con el arrendamiento vencido, y eso no es una contradicción.
    - Cualquier escritura de quien tiene el arrendamiento lo renueva, incluida una escritura que no cambia ningún campo. Por eso no existe ningún comando de latido: trabajar ya es el latido.
    - Escribir sobre una tarea arrendada por otra identidad avisa, pero no impide nada. Un bloqueo de flujo no evita el trabajo duplicado, solo empuja a rodear la herramienta.
    - Un arrendamiento vencido no privilegia a nadie, y liberarlo es una reclamación explícita de `biso start`. Ningún otro comando lo hace por su cuenta, ni siquiera al leer.

Sara está revisando el problema del CRLF y descubre algo que le parece útil para `TASK-11`,
que es del agente y que él tiene en marcha ahora mismo. Lo apunta sin pensar en quién la
tenía.

```console
$ biso note TASK-11 "El CRLF tambien aparece en los ficheros de prueba, no solo en el diff"
warning: TASK-11's lease is held by @claude until 2026-09-06T15:40:18Z
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Exit code: `0`

*(derived output, see [Notas y avisos](spec/salida-y-terminal.md#notas-y-avisos), [`biso start`](spec/cmd/verbos-del-ciclo.md#biso-start), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: Lo primero: la nota se ha escrito. El código de salida es 0 y la tarea ha cambiado. El aviso no es un rechazo, es información. Eso es deliberado y la especificación lo argumenta: un bloqueo de flujo no evita el trabajo duplicado. Si `biso` le hubiera dicho "no puedes escribir aquí", Sara habría escrito la nota en otro sitio, o habría editado el almacén por su cuenta, y entonces el tablero mentiría. Es la misma decisión que con las dependencias sin terminar y con las preguntas abiertas: avisa, no impide. Lo segundo, y es fácil pasarlo por alto: esta escritura **no ha tocado el arrendamiento**. No lo ha renovado, porque Sara no es quien lo tiene, y no lo ha robado, porque solo `biso start` hace eso. Sigue siendo de `@claude` y sigue venciendo a la misma hora. La urgencia sigue en 19.0 porque una nota no cambia nada de lo que la urgencia mide. Y esos 19.0 se descomponen así: 6.0 por prioridad alta, 4.0 por estar activa, 8.0 porque hay otra tarea sin terminar que depende de ella, y 1.0 por tener criterios de aceptación.*

El agente, en su sesión, sigue con lo suyo y anota su propio avance en la misma tarea.

```console
$ biso note TASK-11 "El normalizador ya pasa el caso de los ficheros de prueba"
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Exit code: `0`

*(derived output, see [`biso note`](spec/cmd/verbos-del-ciclo.md#biso-note), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: Ningún aviso, porque el arrendamiento es suyo. Y algo que la salida no dice: acaba de renovarse hasta 240 minutos después de este momento. Ahí está la idea central. No hay un `biso heartbeat`, ni un `biso renew`, ni una bandera para pedirlo. Cualquier escritura del tenedor sobre su tarea renueva la reserva, y "cualquiera" son todas: los seis verbos del ciclo, `biso set` y `biso archive`. Trabajar es lo que mantiene la reserva viva, que es exactamente lo que quieres que signifique.*

El agente quiere confirmar que la prioridad de esa tarea sigue siendo la que cree. La
escribe otra vez con el mismo valor que ya tenía.

```console
$ biso set TASK-11 --priority high
note: TASK-11 unchanged
TASK-11  In Progress  ac 1/2  dod 0/1  urgency 19.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: Este paso parece inútil y es el que cierra el razonamiento. No ha cambiado ningún campo, lo dice la nota, y aun así **el arrendamiento se ha renovado**. Tenía que ser así: si el latido dependiera de que los valores hubieran cambiado de verdad, una escritura que por casualidad coincide con lo que ya había dejaría morir la reserva de alguien que sí está trabajando. La renovación mira quién escribe y sobre qué, no si acertó a cambiar algo. La otra cara: como ningún campo de la tarea ha cambiado, la fecha de última modificación de la tarea **no** se toca. La reserva se renueva y la tarea no se ensucia.*

Ahora la otra mitad del asunto. `TASK-52`, la del documento de la lista de publicación,
lleva días en marcha y la tenía reservada `@bob`, que ya no aparece por aquí. El agente mira
qué hay en curso.

```console
$ biso ls -s "In Progress"
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-19  In Progress  task  high    Retry the upload on 5xx                           ac 0/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-52  In Progress  task  low     Document the release checklist                    ac 0/1  -        -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `TASK-52` sigue en `In Progress`, con toda normalidad, y no tiene a nadie asignado. Su arrendamiento venció el 5 de septiembre y el estado guardado no se ha movido ni un milímetro. Eso no es un descuido, es la regla: lo que vence es la reclamación, no el estado. Nadie va a sacarla de `In Progress` por su cuenta, porque hacerlo significaría que un comando toca tareas que no le nombraste, y porque `biso prime`, que no escribe nunca, acabaría mostrando un estado que la siguiente escritura de otro podría cambiar. El orden de esta lista, por si te lo preguntas, es urgencia descendente y no tiene nada que ver con el estado: 19.0, 11.0, 7.0, 5.4 y 3.1. `TASK-40` va última, pese a estar en marcha y ser de prioridad media, porque depende de `TASK-11` y estar bloqueada resta 5.0.*

El agente decide encargarse él de esa tarea abandonada.

```console
$ biso start TASK-52
note: TASK-52 was already In Progress
TASK-52  In Progress  ac 0/1  urgency 5.4
```

Exit code: `0`

*(derived output, see [`biso start`](spec/cmd/verbos-del-ciclo.md#biso-start), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Con una sola llamada ha hecho tres cosas: ha visto que el estado ya era el activo y lo dice con esa nota, se ha asignado la tarea porque no la tenía nadie, y **ha reclamado el arrendamiento vencido** a su nombre. Esa reclamación ocurre dentro de la misma transacción que comprueba que seguía vencido, y ese detalle importa: si dos agentes reclaman el mismo arrendamiento vencido en el mismo instante, solo gana uno. No hay una ventana en la que los dos crean que es suyo. Lo que esa comprobación no hace es callar a `@bob` si algún día vuelve. Él puede seguir anotando, comentando y cerrando esta tarea, porque escribir nunca está prohibido; lo que ya no puede es renovar ni recuperar una reserva que ahora es de otro, y si quiere recuperarla tiene que pedirla otra vez con `biso start` y saldrá el aviso del primer paso. Es una diferencia deliberada con el patrón clásico de reserva con testigo, y está anotada como riesgo aceptado en la sección 11 de las decisiones de diseño.*

Antes de volver a lo suyo, el agente comprueba que su tarea principal sigue en pie.

```console
$ biso set TASK-19 --priority high
note: TASK-19 unchanged
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [El modelo de datos de una tarea](spec/modelo-de-datos.md); not literal spec text)*

*Note: Otra escritura que no cambia nada y que renueva la reserva de `TASK-19` hasta 240 minutos después de este momento. Con esto la tarea entra en el escenario siguiente con el arrendamiento fresco, y ya sabes que eso no ha costado ningún comando especial.*

## 7. A mitad, falta un criterio {: #escenario-07 }

Llevas un rato con `TASK-19` abierta. Vas anotando lo que compruebas, pero todavía no le has
puesto ninguna definición de hecho: ni una palabra sobre qué tiene que pasar, aparte de los
criterios de aceptación, para que esta tarea se pueda dar por cerrada de verdad. Te paras a
escribirla antes de seguir.

Miras primero si ya hay algo.

!!! abstract "What this scenario teaches"
    - El nombre desnudo de un campo de lista añade; `--set-` sustituye la lista entera; `--rm-` quita uno; `--clear-` la vacía. Las cuatro formas valen igual para cualquier campo de lista, no solo para la definición de hecho.
    - Las claves `#N` de un criterio, sea de aceptación o de la definición de hecho, se asignan al crear el elemento y no se reasignan nunca. Quitar uno de en medio no corre la numeración de los que quedan.
    - El contador de claves de una lista solo crece, también después de `--clear-`. Los elementos que se añaden después de vaciar la lista no recuperan las claves de los que se quitaron.

Pides solo esa sección, para no traerte la ficha entera por algo que sabes que va a estar
vacío.

```console
$ biso get TASK-19 --section dod
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: No sale ni la cabecera. La regla de 10.5 es literal: con `--section`, una sección vacía no se imprime, y si no queda ninguna sección que imprimir la salida entera está vacía. Eso es distinto de pedir la ficha completa, donde toda sección vacía sale igual, marcada con `(empty)`. La bandera cambia el criterio a propósito: sin ella, omitir confundiría "vacío" con "no pedido"; con ella pediste justo eso, así que "vacío" ya no necesita decirse.*

Escribes lo primero que se te ocurre: que el cambio lo revise alguien más antes de darlo
por bueno.

```console
$ biso set TASK-19 --dod "El cambio lo revisa otra persona antes de dar la tarea por terminada"
TASK-19  In Progress  ac 0/2  dod 0/1  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: El nombre desnudo `--dod` añade. Es el mismo verbo que ya conoces de `--ac`: se comporta igual en cualquier campo de lista.*

Y una segunda cosa que tiene que quedar clara antes de cerrar: que el comportamiento del
reintento quede explicado en algún sitio, para quien lo lea después.

```console
$ biso set TASK-19 --dod "El comportamiento del retry queda documentado en el README"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

Compruebas cómo han quedado, con sus claves.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #1 El cambio lo revisa otra persona antes de dar la tarea por terminada
- [ ] #2 El comportamiento del retry queda documentado en el README
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

Pensándolo mejor, el primero sobra: que otra persona revise el cambio ya es política del
equipo para cualquier PR, no algo específico de esta tarea. Lo quitas.

```console
$ biso set TASK-19 --rm-dod 1
TASK-19  In Progress  ac 0/2  dod 0/1  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

Vuelves a mirar la lista para comprobar que lo que queda no ha cambiado de número.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #2 El comportamiento del retry queda documentado en el README
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables), [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: Este es el punto central del escenario. El elemento que sobrevive sigue siendo `#2`, no ha pasado a `#1`. La clave la fija el programa al crear el elemento y no se mueve nunca (5.1): por eso marcar o quitar por número es seguro incluso cuando la lista ha perdido elementos por el medio, y por eso una tarea puede tener perfectamente los criterios `#2` y `#7` sin que eso sea un error de nadie.*

Un compañero revisa la frase y sugiere reescribir toda la lista de una vez, con una
segunda entrada sobre las pruebas que hacen falta.

```console
$ biso set TASK-19 --set-dod "El README explica cuando y cuantas veces se reintenta" --set-dod "El paquete de pruebas cubre el caso 5xx"
warning: --set-dod replaced 58 bytes of existing content
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Notas y avisos](spec/salida-y-terminal.md#notas-y-avisos), [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: `--set-dod` sustituye la lista entera. El elemento `#2` desaparece con el resto de la lista vieja, y los dos que entran son elementos nuevos. El aviso de bytes reemplazados lo describe 4.3 para "cualquier --set-* que pise contenido no vacío", sin restringirlo a los campos de prosa, pero el único ejemplo que trae la especificación es de `--set-plan` (una cadena suelta), y no dice cómo se cuentan los bytes cuando lo que se sustituye es una lista con varios elementos, cada uno con su propio texto. Aquí se ha tomado como los bytes UTF-8 del texto del único elemento que había antes de sustituir (58, los de "El comportamiento del retry queda documentado en el README"), por ser la lectura más cercana al caso de una sola cadena que sí cubre la especificación. Queda anotado como laguna en tutorial/lagunas/07-10.md.*

Compruebas las claves de los dos nuevos.

```console
$ biso get TASK-19 --section dod
TASK-19  Retry the upload on 5xx

## Definition of Done
- [ ] #3 El README explica cuando y cuantas veces se reintenta
- [ ] #4 El paquete de pruebas cubre el caso 5xx
```

Exit code: `0`

*(derived output, see [Los criterios y sus claves estables](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables), [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: Las claves son `#3` y `#4`, no `#1` y `#2`. El contador de esta lista, dentro de esta tarea, ya iba por 2 antes de sustituir nada, y sustituir no lo reinicia: "las claves de los elementos anteriores no se reutilizan" vale también cuando la sustitución viene después de un `--rm-`, no solo justo después de crear la tarea.*

Casi al momento te das cuenta de que estás afinando la redacción antes de tener el código
encajado del todo. Prefieres dejarla en blanco hasta que el trabajo esté más maduro.

```console
$ biso set TASK-19 --clear-dod
TASK-19  In Progress  ac 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: La lista queda vacía, y por eso el trozo `dod` desaparece de la línea de estado: solo sale "siempre que la tarea tenga definición de hecho" (10.6), y una lista vacía no cuenta como tenerla. `ac` sigue saliendo porque esos dos criterios no se han tocado.*

Un rato después, con el reintento ya funcionando, retomas la lista definitiva. Es la misma
redacción de antes.

```console
$ biso set TASK-19 --dod "El README explica cuando y cuantas veces se reintenta" --dod "El paquete de pruebas cubre el caso 5xx"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Nombre desnudo, añade otra vez. Las claves de estos dos elementos son `#5` y `#6`: el contador de la lista no se ha reiniciado ni con el `--rm-dod` de antes ni con el `--clear-dod` de después, "solo crece" en palabras de 5.1, pase lo que pase con el contenido de la lista.*

## 8. Te atascas y alguien tiene que decidir {: #escenario-08 }

Sigues con `TASK-19`, y llegas a un punto en el que no puedes seguir sin que alguien decida
algo: cuántos reintentos como máximo hace falta permitir antes de dar el 5xx por un fallo
definitivo. Podrías inventarte un número, pero eso es justo lo que la regla 11 del mensaje de
arranque de `biso` pide no hacer: "Ask instead of guessing." Antes de aparcar tu pregunta,
mira cómo se ve una desde fuera: `TASK-60` ya tiene una abierta, y así aparece en el bloque
`NEEDS ANSWER` del arranque, tal y como lo imprime `biso prime` (sección 9.7 de la
especificación):

```
NEEDS ANSWER
  TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
    Should the retry budget be shared with the download endpoint or kept separate?
```

La tarea no cambia de estado por tener una pregunta abierta: sigue contando como en curso a
efectos de su ficha, pero deja el bloque `IN PROGRESS` del arranque y aparece aquí en su
lugar, para que nadie la dé por olvidada ni por activa a la vez.

!!! abstract "What this scenario teaches"
    - Aparcar y desaparcar son comandos propios, `biso ask` y `biso answer`, y ninguno de los dos cambia el estado de la tarea.
    - Una tarea aparcada se ve desde fuera, en el bloque `NEEDS ANSWER` del arranque y no en `IN PROGRESS`, aunque su estado siga siendo el activo.
    - Preguntar es preferible a adivinar. Es la regla 11 del mensaje de `biso prime`, y es la razón de ser de `biso ask`.
    - El comando `biso answer` convierte la pregunta en dos comentarios en la misma escritura, uno con el autor y el instante originales de la pregunta, y otro con la respuesta, firmado ahora.

Miras la pregunta de TASK-60 de cerca, no solo la línea recortada del arranque.

```console
$ biso get TASK-60 --section question
TASK-60  Confirm the retry budget for the upload endpoint

## Open Question
@claude, 2026-09-06 09:30
Should the retry budget be shared with the download endpoint or kept separate?
```

Exit code: `0`

Ahora la tuya. En vez de fijar un número a ojo, la dejas escrita como pregunta.

```console
$ biso ask TASK-19 "¿Cuál es el número máximo de reintentos antes de dar el 5xx por fallo definitivo? ¿Hay ya un valor de referencia en otra parte del código, o hay que fijar uno nuevo?"
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 7.0
```

Exit code: `0`

*(derived output, see [`biso ask`](spec/cmd/verbos-del-ciclo.md#biso-ask), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: El estado no cambia, sigue en In Progress. La urgencia sí baja, de 11.0 a 7.0: el término de actividad de la fórmula (5.4) exige que la tarea esté en el estado activo *y* que no tenga pregunta abierta, así que en cuanto `waiting` pasa a cierto ese término deja de sumar sus 4.0 puntos. Es la misma caída de 4.0 que muestra el ejemplo de 10.7.5 sobre TASK-11 (de 19.0 a 15.0), con los números de TASK-19.*

Compruebas que ha quedado bien escrita.

```console
$ biso get TASK-19 --section question
TASK-19  Retry the upload on 5xx

## Open Question
@claude, 2026-09-06 14:10
¿Cuál es el número máximo de reintentos antes de dar el 5xx por fallo definitivo? ¿Hay ya un valor de referencia en otra parte del código, o hay que fijar uno nuevo?
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md), [La pregunta abierta](spec/modelo-de-datos.md#la-pregunta-abierta); not literal spec text)*

Un rato después, mirando un endpoint parecido, encuentras la respuesta tú mismo: ya hay un
valor fijado en otra parte del código. Respondes y la tarea queda desaparcada.

```console
$ biso answer TASK-19 "Ya hay un valor de referencia: el endpoint de descargas usa 3 reintentos con backoff exponencial en retry.go. Usa el mismo valor y la misma lógica aquí."
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso answer`](spec/cmd/verbos-del-ciclo.md#biso-answer), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: La urgencia recupera los 11.0 de antes: waiting vuelve a ser falso y el término de actividad vuelve a sumar sus 4.0 puntos.*

Miras los comentarios para ver cómo ha quedado el rastro.

```console
$ biso get TASK-19 --section comments
TASK-19  Retry the upload on 5xx

## Comments
@claude, 2026-09-06 14:10
¿Cuál es el número máximo de reintentos antes de dar el 5xx por fallo definitivo? ¿Hay ya un valor de referencia en otra parte del código, o hay que fijar uno nuevo?

@claude, 2026-09-06 15:30
Ya hay un valor de referencia: el endpoint de descargas usa 3 reintentos con backoff exponencial en retry.go. Usa el mismo valor y la misma lógica aquí.
```

Exit code: `0`

*(derived output, see [Los comentarios](spec/modelo-de-datos.md#los-comentarios), [Las fechas](spec/modelo-de-datos.md#las-fechas), [`biso get`](spec/cmd/get.md), [`biso answer`](spec/cmd/verbos-del-ciclo.md#biso-answer); not literal spec text)*

*Note: El primer comentario lleva el instante en que se hizo la pregunta (14:10), no el de la respuesta (15:30): es la excepción de forma que anota 5.3, la pregunta se traslada tal cual la vivió el programa la primera vez. Los dos van en el orden en que `biso answer` los escribió, que es también su orden de inserción, y 5.2 dice que ese es el orden que se guarda y se muestra, no el que marca cada instante.*

## 9. La persona y el agente hablan {: #escenario-09 }

Con la duda del reintento resuelta, sigues implementando TASK-19. Vas dejando constancia de lo
que compruebas, para que quien lea la tarea después no tenga que releer el diff entero. Pero
también te llega algo de fuera que no es tuyo: un usuario ha reportado un problema por otro
canal, y alguien te pide que lo dejes anotado en la tarea. Son dos cosas distintas y es fácil
confundirlas: una nota de implementación es tu cuaderno, sin firma ni fecha propia por
entrada; un comentario es una conversación, y siempre lleva quién lo escribió y cuándo.

!!! abstract "What this scenario teaches"
    - El comando `biso note` añade un párrafo a las notas de implementación. Es un bloque de prosa sin autor ni instante por elemento, el cuaderno técnico de quien hace el trabajo.
    - El comando `biso comment` añade un comentario con autor e instante. Es la vía por la que entra lo que viene de fuera, y el canal de conversación con una persona.
    - El autor por defecto de un comentario es la identidad configurada (`me`). `--comment-author` deja firmarlo con otra, texto libre y sin validar, para retransmitir algo que vino de otro sistema.

Terminas de encajar el reintento con el mismo patrón que ya usaba el endpoint de
descargas, y dejas constancia de cómo quedó.

```console
$ biso note TASK-19 "El backoff exponencial de retry.go se puede reutilizar tal cual; solo hay que fijar el límite a 3."
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso note`](spec/cmd/verbos-del-ciclo.md#biso-note), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

Miras cómo ha quedado esa nota.

```console
$ biso get TASK-19 --section notes
TASK-19  Retry the upload on 5xx

## Implementation Notes
El backoff exponencial de retry.go se puede reutilizar tal cual; solo hay que fijar el límite a 3.
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md); not literal spec text)*

*Note: No hay ningún autor ni ningún instante junto al texto. Las notas no tienen esa estructura: son un solo bloque de prosa al que cada `--note` añade un párrafo (8.3), no una lista de entradas con autor como los comentarios (5.2). Si necesitaras saber quién escribió una nota o cuándo, no lo sabrías: no es ese tipo de dato.*

Te llega un aviso: un usuario con conexión móvil ve fallos incluso con el reintento puesto.
No es algo que hayas visto tú, así que lo entras como comentario, firmado por quien lo
reportó.

```console
$ biso comment TASK-19 "Un usuario con conexión móvil ve fallos incluso después de los 3 reintentos, ¿deberíamos subir el límite en ese caso?" --comment-author @trello:juan
note: comment #3 by @trello:juan
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso comment`](spec/cmd/verbos-del-ciclo.md#biso-comment), [Los comentarios](spec/modelo-de-datos.md#los-comentarios); not literal spec text)*

*Note: "note: comment #3 by @trello:juan" sale por stderr, con la forma exacta que da el ejemplo de 10.7.3. El número no es una clave estable como la de un criterio (5.1): es la posición del comentario en la lista, y como los comentarios nunca se editan ni se borran (append-only, 10.7.3), esa posición no vuelve a cambiar. Este es el tercero de TASK-19 porque `biso answer` ya había añadido dos en el escenario anterior.*

Respondes tú, como tú mismo, sin banderas de autor: la identidad configurada firma por
defecto.

```console
$ biso comment TASK-19 "De momento no, 3 reintentos es la política del resto del sistema; si se repite lo revisamos."
note: comment #4 by @claude
TASK-19  In Progress  ac 0/2  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [`biso comment`](spec/cmd/verbos-del-ciclo.md#biso-comment), [Los comentarios](spec/modelo-de-datos.md#los-comentarios); not literal spec text)*

*Note: Sin --comment-author, el autor es `me`, la identidad configurada de este tablero: @claude. Es el mismo `--comment-author` que existe en todos los comandos de escritura, no solo en `biso comment`: el nombre no cambia porque en `biso comment` parezca redundante (10.6).*

Pones las dos secciones una junto a otra para ver el contraste de un vistazo.

```console
$ biso get TASK-19 --section notes,comments
TASK-19  Retry the upload on 5xx

## Implementation Notes
El backoff exponencial de retry.go se puede reutilizar tal cual; solo hay que fijar el límite a 3.

## Comments
@claude, 2026-09-06 14:10
¿Cuál es el número máximo de reintentos antes de dar el 5xx por fallo definitivo? ¿Hay ya un valor de referencia en otra parte del código, o hay que fijar uno nuevo?

@claude, 2026-09-06 15:30
Ya hay un valor de referencia: el endpoint de descargas usa 3 reintentos con backoff exponencial en retry.go. Usa el mismo valor y la misma lógica aquí.

@trello:juan, 2026-09-06 16:05
Un usuario con conexión móvil ve fallos incluso después de los 3 reintentos, ¿deberíamos subir el límite en ese caso?

@claude, 2026-09-06 16:20
De momento no, 3 reintentos es la política del resto del sistema; si se repite lo revisamos.
```

Exit code: `0`

*(derived output, see [`biso get`](spec/cmd/get.md), [Los comentarios](spec/modelo-de-datos.md#los-comentarios); not literal spec text)*

*Note: Las notas son un único bloque; los comentarios son cuatro entradas firmadas, en el orden en que se escribieron. `## Implementation Notes` sale antes que `## Comments` aunque en la bandera se haya pedido al revés (`notes,comments`): el orden de impresión de las secciones es el orden fijo de la ficha completa (10.5), no el orden en que se listan en `--section`. La especificación no lo dice de forma explícita para el caso de varias secciones a la vez; queda anotado en tutorial/lagunas/07-10.md.*

## 10. Cierras {: #escenario-10 }

El reintento funciona, está documentado y hay quien lo ha revisado por encima en los
comentarios. Antes de cerrar TASK-19 repasas los criterios de aceptación, y te das cuenta de
que dos cosas que ya has hecho nunca quedaron anotadas como criterio: el backoff y las
pruebas. Las añades antes de marcar nada, para que la ficha final cuente la historia completa.

!!! abstract "What this scenario teaches"
    - Las banderas `--check` y `--uncheck` (y sus pares `--check-dod` y `--uncheck-dod`) aceptan cinco formas de selector, todas repetibles y todas combinables en la misma llamada, una clave suelta, un rango, una lista separada por comas, `all`, o el texto del criterio.
    - Marcar o desmarcar es siempre seguro de repetir. `--check all` no rompe nada aunque ya estuviera todo marcado, y por eso el cierre de un ciclo siempre puede llevarlo sin mirar antes qué falta.
    - Cerrar una tarea vacía siempre su arrendamiento, lo tenga quien lo tenga. Es la invariante que gobierna cualquier escritura que saca una tarea del estado activo, y `biso finish` la aplica sin excepción.

Añades los dos criterios que te faltaban, de una vez.

```console
$ biso set TASK-19 --ac "El backoff sigue el mismo patron que el endpoint de descargas" --ac "Hay un test que cubre el caso de 5xx repetido" --ac "Hay un test que cubre el caso de exito tras reintentar" --ac "La documentacion del endpoint menciona el limite de reintentos"
TASK-19  In Progress  ac 0/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Campos de lista](spec/familias-de-banderas.md#campos-de-lista), [`biso set`](spec/cmd/set.md), [Los criterios y sus claves estables](spec/modelo-de-datos.md#los-criterios-y-sus-claves-estables); not literal spec text)*

*Note: Cuatro `--ac` en la misma llamada añaden cuatro elementos nuevos, en el orden en que se escriben (4.9), con claves nuevas, #3, #4, #5 y #6, detrás de los #1 y #2 que ya tenía la tarea desde el inventario inicial.*

Marcas el primero, el que ya tenías comprobado desde el principio.

```console
$ biso set TASK-19 --check 1
TASK-19  In Progress  ac 1/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

Los tres que probaste juntos con el mismo test, de una vez, con un rango.

```console
$ biso set TASK-19 --check 3-5
TASK-19  In Progress  ac 4/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

Y dos que no son consecutivos, con una lista separada por comas.

```console
$ biso set TASK-19 --check 2,6
TASK-19  In Progress  ac 6/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

Casi al momento te das cuenta de que el #2 (el del contador en el log) lo diste por bueno
demasiado pronto, el número de intentos todavía no se ve en el log, solo el resultado
final. Lo desmarcas.

```console
$ biso set TASK-19 --uncheck 2
TASK-19  In Progress  ac 5/6  dod 0/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: `--uncheck` toma el mismo selector que `--check`, con el mismo efecto invertido, aquí "uno suelto", igual que el primer `--check` de este escenario.*

En la definición de hecho, marcas el que ya está cerrado buscando por su texto, no por
número.

```console
$ biso set TASK-19 --check-dod "README"
TASK-19  In Progress  ac 5/6  dod 1/2  urgency 11.0
```

Exit code: `0`

*(derived output, see [Selectores de criterios](spec/familias-de-banderas.md#selectores-de-criterios), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: "README" no encaja con la gramática de claves de 8.4 (^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$), así que se trata como texto literal y busca el elemento cuyo texto lo contenga. Solo el #5 ("El README explica cuándo y cuántas veces se reintenta") encaja, así que no hay ambigüedad.*

Añades el número de intentos al log, comprueba que ahora sí aparece, y cierras la tarea.
Marcas todo lo que quede de una vez, por si te has dejado algo suelto, y escribes el
resumen.

```console
$ biso finish TASK-19 --check all --check-dod all --summary "Retry con el mismo backoff que el endpoint de descargas, tope de 3 intentos, contador visible en el log."
TASK-19  Done  ac 6/6  dod 2/2  urgency 0.0
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [`biso finish`](spec/cmd/verbos-del-ciclo.md#biso-finish), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `--check all` vuelve a marcar el #2, que habías desmarcado, y no toca los otros cinco, ya estaban marcados, y marcar dos veces el mismo criterio "se queda marcado, sin aviso, la operación es idempotente" (8.4). `--check-dod all` marca el #6, el único que quedaba. La urgencia pasa a 0.0 porque el estado terminal la fija así por definición (5.4), no porque se recalcule con la fórmula de siempre. Lo que esta línea no dice, y conviene saber, es que cerrar una tarea vacía siempre `leaseExpiresAt` y `leaseHolder`, sea quien sea quien la tenga (sección 5, séptima precisión, citada de nuevo en 10.6 y en 10.7.4). TASK-19 llevaba el arrendamiento de @claude desde el escenario 5; con este `biso finish` queda sin arrendamiento, y quedaría igual de vacío aunque lo hubiera tenido otra identidad. La invariante gana siempre sobre la regla general de "una escritura ajena no toca el arrendamiento": una tarea terminada con arrendamiento vivo es un estado que ni la propia importación del tablero aceptaría.*

## 11. Te equivocas {: #escenario-11 }

Llevas diez capítulos usando `biso` sin pensar demasiado en ello: pides una lista, pides una
tarea, y te la da. Es fácil olvidar que detrás de cada consulta hay una regla que decide qué
pasa cuando lo que pides no existe. Hoy tocan tres tropiezos seguidos, a propósito, porque son
los que enseñan la regla más importante de toda la especificación: un valor que el tablero no
conoce es un error, nunca una lista vacía. Ninguno de los tres escribe nada: este es un
capítulo de consulta, no de escritura, así que el tablero sale exactamente como entra.

!!! abstract "What this scenario teaches"
    - Un valor que el tablero no conoce es un error, se esté leyendo o escribiendo, y nunca una lista vacía. Por eso una lista vacía sí es un hecho sobre el tablero, y quien la recibe puede actuar en consecuencia.
    - La coincidencia de vocabulario ignora mayúsculas, espacios, guiones y guiones bajos, así que `TO_DO`, `to do` y `To-Do` no son tres filtros parecidos: son el mismo filtro.
    - Una referencia de texto que encaja con varias tareas no adivina cuál querías decir: lista las candidatas y te deja elegir.
    - Una referencia bien formada que el tablero nunca ha llegado a asignar es un error distinto del de una tarea que sí existió y ya no está, aunque los dos mensajes empiecen igual.
    - Cada tropiezo tiene su propio código de salida, así que un agente puede ramificar sobre el número sin leer el mensaje.

Quieres ver qué queda pendiente y escribes el nombre del estado tal y como lo llamarías
en cualquier otro gestor de tareas.

```console
$ biso ls -s Pending
error: unknown status: "Pending"
       valid statuses on this board: To Do, In Progress, Done
```

Exit code: `3`

*Note: No hay ningún «Pending» en este tablero, así que el programa no te devuelve una lista vacía fingiendo que lo ha entendido: falla con el código 3 (BAD_VALUE) y te dice cuáles son los valores válidos, para que no tengas que adivinar el nombre exacto. Es el mismo mensaje, palabra por palabra, tanto si el valor imposible llega en un filtro de `biso ls` como si llega en una escritura de `biso set` (sección 6.2).*

Pruebas con un valor que sí existe, pero escrito como te salga: en mayúsculas, con guion
bajo.

```console
$ biso ls --count -s TO_DO
54
```

Exit code: `0`

*(derived output, see [El algoritmo de coincidencia](spec/vocabularios.md#el-algoritmo-de-coincidencia); not literal spec text)*

*Note: `TO_DO` no aparece tal cual en la configuración del tablero, que guarda `To Do`. Pero el algoritmo de coincidencia de la sección 6.1 pasa los dos valores a minúsculas y les quita los espacios, los guiones y los guiones bajos antes de comparar, así que `TO_DO` normaliza a `todo`, igual que `To Do`. El tablero no ha cambiado desde que terminó el capítulo anterior: sigue habiendo 54 tareas en `To Do`. `--count` imprime solo ese número, sin ninguna otra línea (sección 10.4).*

Vuelves a probar, ahora con guion y mezclando mayúsculas y minúsculas.

```console
$ biso ls --count -s To-Do
54
```

Exit code: `0`

*(derived output, see [El algoritmo de coincidencia](spec/vocabularios.md#el-algoritmo-de-coincidencia); not literal spec text)*

*Note: Mismo resultado, porque es el mismo filtro: `To-Do` normaliza también a `todo`. La misma regla quita además los espacios repetidos, así que `to do`, con dos espacios en medio, también habría servido (es el ejemplo literal de la propia tabla de la sección 6.1). Cuatro formas distintas de teclearlo, un solo valor.*

Se te ha olvidado el número de la tarea de los reintentos y escribes lo primero que
recuerdas del título.

```console
$ biso get "retry"
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-33  To Do        task  medium  Add a retry counter to the upload log             ac 1/3  @claude  -
TASK-19  Done         task  high    Retry the upload on 5xx                           ac 2/2  @claude  -
```

Exit code: `5`

*(derived output, see [La búsqueda por texto](spec/referencias.md#la-búsqueda-por-texto); not literal spec text)*

*Note: Las tres tareas llevan «retry» en el título, así que ninguna gana por estar en el título mientras las otras solo lo llevan en el cuerpo (la regla de la sección 7.2): las tres son candidatas por igual. El programa no elige por ti. Código 5 (AMBIGUOUS), y las tres filas por stdout, en el mismo formato de columnas de `biso ls` (sección 10.4), ordenadas por urgencia porque ninguna lleva `ordinal`: TASK-60 tiene una pregunta abierta, así que no suma el término de tarea activa, pero es de prioridad alta; TASK-33 es de prioridad media; TASK-19 ya está terminada, así que su urgencia es 0,0 (sección 5.4, «si el estado de la tarea es el terminal, urgency = 0.0») y va la última. El valor de `ac` de TASK-19 depende de cómo se cerrara en el capítulo 10, que no está a la vista de este fichero: se asume que llegó a `ac 2/2`, siguiendo el cierre que recomienda el propio mensaje de `biso prime` («biso finish <ref> --check all --check-dod all --summary "..."»). Queda anotado en `tutorial/lagunas/11-13.md`.*

Escribes el número a mano, pero se te cuela un dígito de más.

```console
$ biso get TASK-99
error: TASK-99 has never existed on this board
note: the highest id ever assigned here is TASK-62
```

Exit code: `4`

*(derived output, see [Los tres mensajes de "no la encuentro"](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro); not literal spec text)*

*Note: `TASK-99` está bien formado como identificador, así que esto no es un error de uso (código 2): es un error de que ese identificador nunca ha existido, código 4 (NOT_FOUND), con la clave `never_allocated`. El texto es el literal de la sección 7.3, con el número del identificador más alto sustituido por el de este tablero, `TASK-62`, que es la tarea que se creó en el capítulo 2. Hay un tercer mensaje de «no la encuentro» en la misma sección, para una tarea que el tablero sí llegó a asignar y que ahora ya no está (archivada y después eliminada): comparte el código 4 pero no la clave, y no hace falta reproducirlo aquí para que quede clara la diferencia entre los dos. Y queda un cuarto tropiezo que ni siquiera hemos tocado, el identificador mal formado como `TASK-1.1`, que es código 2 en vez de 4, porque ahí el problema no es que la tarea no exista sino que lo que has escrito no tiene la forma de una referencia. Cuatro formas de equivocarte, cuatro códigos distintos: 2, 3, 4 y 5.*

## 12. Aparcas algo a medias {: #escenario-12 }

En el escenario 6 el agente se quedó con `TASK-52`, la tarea de documentar la lista de
publicación que `@bob` había dejado abandonada. Le pareció que alguien tenía que recogerla.

Ahora, con la subida ya arreglada y cerrada, la mira otra vez y cambia de opinión: nadie ha
publicado nada en semanas, no hay prisa, y tenerla en marcha a su nombre da una imagen falsa de
lo que está pasando en el tablero. No es que esté terminada. Es que no toca.

Eso tiene un gesto propio, y no es el que la mayoría busca primero.

!!! abstract "What this scenario teaches"
    - Archivar no es terminar. Una tarea archivada no pasa al estado terminal: sale del tablero activo con el estado que tuviera.
    - No existe ningún comando para borrar una tarea, y esa ausencia está especificada. El identificador de una tarea archivada no se reutiliza nunca.
    - Archivar vacía el arrendamiento, aunque estuviera vivo y aunque archivar no saque la tarea del estado activo. Un arrendamiento afirma que alguien está trabajando ahora, y archivar dice justo lo contrario.

El primer impulso de casi todo el mundo es buscar cómo se borra.

```console
$ biso delete TASK-52
error: there is no delete command, on purpose
hint: `biso archive <ref>` takes it off the board and keeps the history
      an archived task still exists: `biso ls --archived` lists them, and the
      id is never reused
```

Exit code: `2`

*Note: La ausencia de este comando está escrita en la especificación, con su mensaje y su código, en vez de dejar que `biso delete` caiga en el error genérico de comando desconocido. La diferencia importa: un comando desconocido te hace dudar de si lo escribiste mal, y este mensaje te dice que la decisión es deliberada y cuál es el gesto que buscabas.*

El agente saca la tarea del tablero.

```console
$ biso archive TASK-52
TASK-52  In Progress  ac 0/1  urgency 5.4  archived
```

Exit code: `0`

*(derived output, see [`biso archive`](spec/cmd/archive.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Lee con cuidado esa línea, porque dice dos cosas que parecen incompatibles: la tarea sigue en `In Progress` y está `archived`. No hay contradicción. Archivar no es un estado y no toca el estado: es una marca aparte que la saca del tablero activo. Si mañana se desarchivara, volvería a `In Progress`, que es donde estaba. Y hay algo que la línea no dice: el arrendamiento que el agente había reclamado en el escenario 6 acaba de vaciarse. Eso pasa siempre al archivar, sea el arrendamiento de quien sea y esté vivo o vencido, y el motivo es que conservarlo lo guardaría donde nadie lo ve, porque el arranque y el listado esconden las archivadas por defecto. Si esta tarea volviera al tablero dentro de tres semanas con un arrendamiento a nombre de una sesión que ya murió, ese dato sería basura con aspecto de información. La urgencia sigue calculándose, 5.4, porque la tarea sigue en un estado que no es el terminal. Archivar no la pone a cero: eso lo hace terminarla.*

Comprueba que no la ha perdido.

```console
$ biso ls --archived
TASK-52  In Progress  task  low  Document the release checklist  ac 0/1  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md), [`biso archive`](spec/cmd/archive.md); not literal spec text)*

*Note: Ahí sigue, entera, con su asignación intacta. Lo único que ha perdido es el arrendamiento. Fíjate en que `--archived` es lo que la trae de vuelta a la vista: sin esa bandera no aparece ni en `biso ls` ni en `biso prime`, y eso es lo que significa "salir del tablero activo". La tarea no se ha ido a ninguna parte, se ha quitado de en medio.*

Y para cerrar la idea, mira qué queda en marcha ahora.

```console
$ biso ls -s "In Progress"
TASK-11  In Progress  bug   high    Normalize CRLF in the diff                        ac 1/2  @claude  -
TASK-60  In Progress  task  high    Confirm the retry budget for the upload endpoint  ac 0/2  @claude  -
TASK-40  In Progress  task  medium  Split the config loader                           ac 0/2  @claude  -
```

Exit code: `0`

*(derived output, see [`biso ls`](spec/cmd/ls.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Tres tareas donde antes había cuatro, y `TASK-19` ya no está en ninguna de las dos listas porque se terminó en el escenario 10. El tablero dice ahora la verdad sobre lo que se está haciendo, que es para lo que sirve archivar algo que no vas a tocar.*

## 13. Tocas veinte de golpe {: #escenario-13 }

Última parada. Hasta aquí todo lo que has hecho ha sido sobre una tarea a la vez, y así es como
se trabaja el noventa por ciento del tiempo.

Pero llega el día en que hay que retocar veinte. El equipo decide que las tareas de documentación
bajan de prioridad este trimestre, o llega una lista de incidencias de otro sistema y hay que
meterlas todas. Y ahí aparece una pregunta que da miedo: si algo falla a mitad del lote, ¿qué
queda escrito?

La respuesta de `biso` es que nada. Vale la pena ver por qué.

!!! abstract "What this scenario teaches"
    - Todo lo que se puede hacer una vez se puede hacer cien, y con los mismos nombres de bandera. No hay una sintaxis aparte para el lote.
    - La validación es previa y total: se comprueba el lote entero antes de escribir nada. Un lote no se queda a medias, así que no hay que averiguar por dónde se rompió.
    - `--dry-run` valida sin escribir y contesta con el código de salida: 0 si habría funcionado, 9 si no. Es la forma de mirar antes de saltar.

El agente empieza por lo fácil: dos tareas que hay que bajar de prioridad, en una sola
llamada.

```console
$ biso set TASK-44 TASK-61 --priority low
TASK-44  To Do  ac 0/1  urgency 1.6
TASK-61  To Do  ac 0/1  urgency 1.4
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: Una línea por tarea, con el mismo formato que cuando tocas una sola. No hay un modo lote que cambie la forma de la salida ni un resumen que sustituya al detalle: lo que aprendiste con una vale para las dos. `TASK-44` ya era de prioridad baja, así que su urgencia no se mueve. `TASK-61` baja de 4.5 a 1.4, y ahí hay una lección escondida sobre leer estos números: la caída real es de exactamente 3.0, los que aportaba la prioridad media, pero las dos cifras que ves están redondeadas a un decimal desde 4.45 y 1.45, y esos dos redondeos no van en la misma dirección. La urgencia sirve para ordenar una lista, no para hacer aritmética con ella.*

Ahora una lista más larga, y con una errata dentro: una de las referencias no existe.

```console
$ biso set TASK-7 TASK-33 TASK-99 --priority medium
error: TASK-99 has never existed on this board
note: the highest id ever assigned here is TASK-62
```

Exit code: `4`

*(derived output, see [Los tres mensajes de "no la encuentro"](spec/referencias.md#los-tres-mensajes-de-no-la-encuentro), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: Y esto es lo importante: **`TASK-7` y `TASK-33` no se han tocado**. No es que se hayan escrito y luego se haya deshecho: es que no se llegó a escribir nada, porque la validación de las tres referencias ocurre antes de la primera escritura. Piensa en lo que te ahorra. Si el lote se aplicase tarea a tarea y se detuviese al fallar, después de este error tendrías que averiguar cuántas se habían escrito ya para no repetirlas, y eso es imposible de hacer bien desde un programa. Aquí el contrato es simple: o todas o ninguna.*

Corrige la errata y esta vez, antes de escribir, pide que se lo valide sin tocar nada.

```console
$ biso set TASK-7 TASK-33 --priority medium --dry-run
2 tasks would be updated, nothing was written (--dry-run)
```

Exit code: `0`

*(derived output, see [Banderas globales](spec/invocacion.md#banderas-globales), [`biso set`](spec/cmd/set.md); not literal spec text)*

*Note: Código 0, así que habría funcionado. Ese código es la respuesta de verdad, más que el texto: un programa que llame a `biso` no necesita leer nada para saber si el lote es válido. El texto exacto de esta línea es una derivación. La especificación da la frase literal para el lote de `biso new --from` ("242 tasks would be created, nothing was written (--dry-run)") y dice que `--dry-run` vale en todos los comandos que escriben, pero no escribe la frase para `biso set`. Aquí se ha usado la misma forma cambiando el verbo, que es lo más probable, y queda anotado en tutorial/lagunas/11-13.md.*

Con la validación en verde, ejecuta de verdad.

```console
$ biso set TASK-7 TASK-33 --priority medium
TASK-7   To Do  ac 0/4  urgency 15.3
TASK-33  To Do  ac 1/3  urgency 4.3
```

Exit code: `0`

*(derived output, see [`biso set`](spec/cmd/set.md), [La urgencia](spec/modelo-de-datos.md#la-urgencia); not literal spec text)*

*Note: `TASK-7` cae de 18.3 a 15.3, que son los 3.0 que pierde al pasar de prioridad alta a media, y sigue siendo la más urgente del tablero con diferencia: le queda el término de fecha límite, que aporta 11.2 porque vence en dos días. `TASK-33` no se mueve, porque ya era de prioridad media, y aun así aparece en la salida: te dice en qué queda cada tarea que nombraste, no solo las que cambiaron. Fíjate también en la alineación de la primera columna. El ancho lo fija el identificador más largo de la salida, así que `TASK-7` lleva un espacio de más para cuadrar con `TASK-33`.*

Para acabar, el otro lote que existe: meter tareas nuevas desde un fichero. El agente ha
recibido tres incidencias por otro canal y las tiene en un fichero, una por línea. Antes de
importarlas, las valida.

```console
$ biso new --from incidencias.ndjson --dry-run
3 tasks would be created, nothing was written (--dry-run)
```

Exit code: `0`

*Note: Esta frase sí es literal de la especificación, con otro número. Y aquí `--dry-run` es donde más sirve, porque un fichero de importación puede traer un tipo, un estado o una clave que este tablero no conoce, y un valor que el tablero no conoce siempre es un error. Con el lote de importación eso significa que un solo campo mal escrito en la línea 200 impide crear las 199 anteriores. Suena severo y es lo que quieres: al revés tendrías un tablero a medio importar y ninguna forma fiable de saber por dónde continuar. Si la validación fallase, el código sería 9, no 4 ni 3: nueve significa exactamente "no se escribió nada porque la validación no pasó", y se distingue del error de un valor concreto justamente porque lo que informa es sobre el lote entero.*
