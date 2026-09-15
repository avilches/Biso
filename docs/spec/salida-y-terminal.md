# Terminal, flujos de salida y codificación

## Interactividad, terminal y color

**Ningún comando abre nunca una interfaz interactiva por su cuenta, y ningún comando pregunta nada.**
No existe la detección de terminal como forma de decidir qué imprime un comando: la salida de
cualquier comando es idéntica byte a byte con terminal y sin él, salvo los códigos de color.

La interfaz interactiva existe, pero es un comando aparte, `biso board`, que solo se ejecuta si se
pide por su nombre y que falla con código 8 si no hay terminal.

Lo único que mira el terminal es el color:

| Situación | Color |
|---|---|
| `--color always` | sí |
| `--color never`, o `NO_COLOR` definida | no |
| `--color auto` y stdout es un terminal | sí |
| `--color auto` y stdout está redirigido | no |

El color se decide por separado para stdout y para stderr, cada uno según su propio destino. Los
códigos de color nunca cambian el texto: quitarlos deja exactamente las líneas documentadas aquí.

## stdout, stderr y qué va en cada uno

La regla es fija y no tiene excepciones:

- **stdout lleva datos.** Lo que un programa consumiría: identificadores, listados, fichas, JSON.
- **stderr lleva todo lo demás.** Errores, avisos, notas informativas, sugerencias y el resumen de lo
  que se ha omitido.

En particular, la línea de `biso ls` que dice cuántas tareas se han ocultado va por stderr, porque no
forma parte del listado. Redirigir stdout a un fichero produce un fichero de datos limpio, y
redirigirlo a `/dev/null` no pierde ni un solo aviso.

## Notas y avisos

Hay tres clases de línea que no son errores, las tres por stderr:

- **`note:`** es información de contexto. `--quiet` la suprime.
- **`warning:`** es algo que quien llama necesita saber y que no impide la operación. **Nunca se
  suprime.**
- **La salida de un programa ajeno**, prefijada con el nombre del sistema de control de versiones y dos
  puntos, o sea `git:` o `custom:`. Solo la emite `biso snapshot`, que es el único comando que ejecuta
  otro programa, y las reglas de cuándo aparece están en la sección ["`biso snapshot`"](cmd/snapshot.md). **Nunca se suprime**, tampoco
  con `--quiet`: de una línea que `biso` no ha escrito no puede juzgar si sobra.

Las dos primeras dejan el código de salida en 0. La tercera acompaña igual a una operación que va bien
que a una que falla, y ahí el código lo decide el resultado de la orden, nunca la línea.

Esta es la lista completa de avisos que el programa emite. No hay ningún otro:

| Aviso | Cuándo |
|---|---|
| `warning: --replace-labels replaced 2 existing labels` | cualquier `--replace-*` que sustituya una lista no vacía |
| `warning: MYP-11 moved to Done with 1 of 2 acceptance criteria unchecked` | al llegar a un estado terminal con criterios sin marcar |
| `warning: MYP-11 finished without a final summary` | al llegar a un estado terminal sin resumen |
| `warning: MYP-11 moved to Done with 1 of 3 definition-of-done items unchecked` | al llegar a un estado terminal con la definición de hecho a medias |
| `warning: MYP-11 has unfinished subtasks: MYP-14, MYP-15` | al terminar una tarea con subtareas vivas. Una subtarea archivada sin terminar se marca `MYP-15 (archived)` dentro de la misma lista, en vez de listarse igual que una viva (["`biso finish`"](cmd/verbos-del-ciclo.md#biso-finish)) |
| `warning: MYP-11 is a dependency of MYP-20, which is not finished` | al archivar una tarea de la que dependen otras vivas |
| `warning: --clear-labels has no effect on a new task` | cualquier `--clear-*` en `biso new` |
| `warning: MYP-11 has unresolved dependencies: MYP-4 (To Do)` | al empezar una tarea bloqueada |
| `warning: 28 more tasks match; showing 30 of 58` | en `biso ls`, al recortar |
| `warning: --add-labels: "urgent" given twice, kept once` | valor repetido en un flag de lista |
| `warning: --append-desc contains a literal \n and no real newline; it will be stored as text` | ver ["Codificación y texto"](#codificación-y-texto) |
| `warning: --append-note: empty value, nothing was added` | valor vacío en un flag que añade |
| `warning: --due 2026-01-01 is in the past` | fecha límite ya pasada |
| `warning: MYP-11 has no acceptance criteria` | `--check-ac all` sobre una tarea sin criterios |
| `warning: MYP-11 has no comments` | `--rm-comment all` sobre una tarea sin comentarios (["Comentarios"](familias-de-flags.md#comentarios)) |
| `warning: 1 task could not be read and was skipped` | ver ["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar) |
| `warning: <x> is deprecated and will be removed in 2.0` | ver la sección ["El contrato de estabilidad"](estabilidad.md) |
| `warning: MYP-11 has an open question, asked by @sara` | al empezar una tarea con una pregunta abierta |
| `warning: MYP-11 moved to Done with an open question, asked by @sara` | al llegar a un estado terminal con una pregunta abierta |
| `warning: MYP-11's lease is held by @sara until 2026-09-08T14:00:00Z` | al escribir sobre una tarea cuyo arrendamiento está vivo y es de otra identidad, con `biso start` o con cualquier otra escritura (["La renovación"](lease.md#la-renovación) de `lease.md`, ["Saber si alguien está trabajando de verdad"](../decisiones/modelo-de-estados.md#saber-si-alguien-está-trabajando-de-verdad), y ["`biso start`"](cmd/verbos-del-ciclo.md#biso-start)) |

## Codificación y texto

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
  warning: --append-desc contains a literal \n and no real newline; it will be stored as text
  hint: use a real newline, or -d @file.md, or -d - to read from stdin
  ```

