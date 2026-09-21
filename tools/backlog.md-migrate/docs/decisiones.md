# Decisiones de diseño

> **Provisional.** Ninguna de estas decisiones está cerrada: son propuestas que se van a repasar una
> por una, y varias dependen de un cambio en `biso` (TASK-73). La lista completa de lo que
> falta revisar está en [`pendientes.md`](pendientes.md).

Cada entrada dice primero la decisión vigente, en un párrafo que se puede citar sin más contexto. Las
alternativas que se consideraron van después, marcadas como **descartadas** y con la razón. Si una
decisión nueva deja obsoleta una anterior, la vieja se corrige aquí mismo.

## Es un proyecto independiente de `biso`

**La decisión.** `backlog.md-migrate` vive en `tools/backlog.md-migrate/` del repositorio de `biso`,
con su propio módulo Go, su propia documentación y sus propias decisiones. No importa ningún paquete de
`internal/` de `biso`, no cuelga de su documentación ni de su sitio, y habla con `biso` ejecutando su
binario. Una utilidad de migración es utillaje que se usa una vez o unas pocas, no parte del producto,
y así no toca el binario Go de `biso`, ni su presupuesto de arranque, ni el recuento de comandos de su
especificación.

**Descartado: un subcomando de `biso`** (`biso import-backlog`). Obligaba a añadirlo a la
especificación, al recuento de comandos y a la ayuda, y a meter en el binario un lector de un formato
ajeno. **Descartado: un repositorio aparte.** La prueba necesita un `biso` compilado y viviendo en el
mismo repositorio lo encuentra en `bin/biso`, sin fijar cómo se localiza entre repositorios.

## Dos órdenes independientes, una dirección cada vez

**La decisión.** `import` y `export` son dos órdenes hermanas. Cada una convierte en un sentido y se
ejecuta una vez, o unas pocas mientras los dos tableros conviven. Ninguna mira el estado de la otra.

**Descartado: un sincronizador continuo.** Exige decidir qué gana cuando cambian los dos lados, y ese
es justo el tipo de problema que el proyecto no quiere tener. El nombre `backlog.md-migrate` lo dice.

## El milestone es una etiqueta `milestone:<slug>`

**La decisión.** `biso` no tiene milestone en su modelo de datos, y en Backlog.md es un dato que se
guarda en cada tarea que pertenece a uno. Cada tarea recibe una etiqueta `milestone:<slug>`, con el slug del título del
milestone. El alfabeto de las etiquetas admite los dos puntos, no añade ningún campo al modelo, se
filtra con `--label` y se ve en `biso ls`.

**Descartado: un campo de extensión (`ext`).** Su clave es un identificador en otro sistema y no está
claro que se pueda filtrar como una etiqueta. **Descartado: asumir la pérdida.** Se perdería la
agrupación por hitos de los tableros que la usan.

**El campo `project` de Backlog.md sigue la misma regla**, con la etiqueta `project:<slug>`. Es otro
dato de agrupación que `biso` no modela y que Backlog.md guarda como un texto en cada tarea, y darle un
trato distinto al del milestone no tendría justificación.

## Los identificadores conservan su número y cambian de prefijo

**La decisión.** El tablero destino se crea antes, con el prefijo que se quiera, y el convertidor lo
lee de `biso`. Cada id de origen conserva su número con el prefijo del destino. Solo los ids cuyo número
ya existe en el destino cambian de número, con el siguiente libre por encima del mayor que haya entre
origen y destino, y las menciones de esos ids se reescriben en todo el texto, títulos incluidos. Ninguna
tarea que no choque se desplaza. Así el destino puede tener el prefijo que quiera y no tiene por qué
estar vacío. **Una excepción: una tarea que ya está en el destino no se vuelve a importar.** Si el
destino tiene una tarea con el mismo título y la misma fecha de creación, sea cual sea su id, es la
misma tarea de una importación anterior y se salta con un hallazgo. Sin esto, ejecutar `import` dos
veces duplicaría el tablero: `biso` rechazaría el lote por ids ocupados, pero el convertidor lo evitaría
reasignando todos los ids.

**Las subtareas siempre reciben un número nuevo.** Backlog.md da a una subtarea un id con punto
(`XYZ-001.01`, `ABC-1.2`), y un id de `biso` es siempre `<PREFIJO>-<n>`. No hay forma de conservarlo, así
que se le asigna un número libre y se reescriben sus menciones. El parentesco no se pierde: va en
`parent`. Por eso la regla de la tarea ya importada no compara los ids, sino el título y la fecha.

**La coincidencia de menciones distingue mayúsculas y no toca lo que va pegado a otras palabras.** En el
tablero medido hay cinco menciones en minúsculas que son nombres de ramas y de worktrees
(`task-10-modelo`), y reescribirlas sin distinguir mayúsculas las habría corrompido.

**Descartado: obligar a que el destino tenga el mismo prefijo que el origen** (`--prefix TASK`).
Dejaba un tablero con un prefijo que no elige nadie y que no se puede cambiar después. **Descartado:
renumerar todo desde el siguiente libre.** Cambia todos los números aunque no choque nada y rompe el
parecido entre los dos tableros. **Descartado: dejar que `biso` asigne los ids.** Un lote no puede
expresar un padre o una dependencia sobre un id que todavía no existe.

## Las fechas se leen como UTC

**La decisión.** Backlog.md 1.52.0 guarda las fechas como `YYYY-MM-DD HH:mm` en UTC, y `biso` usa
instantes UTC con precisión de segundo. La conversión añade `:00` y la `Z`. Se comprobó contra un
reloj independiente: el commit automático que hizo Backlog.md al editar una tarea lleva la hora
`17:30:22-04:00`, y el `updated_date` de esa misma tarea dice `21:30`, es decir, UTC y no la hora local
de esa máquina. Sin `updated_date`, `updatedAt` toma la
`createdAt`, porque `biso` pondría de otro modo el instante de la importación y la tarea parecería
recién modificada.

**Descartado: un flag de zona horaria.** Hasta que aparezca una versión de Backlog.md que no guarde en
UTC, sería configuración para un caso que no existe.

## Estados, tipos y prioridades son los del destino

**La decisión.** La utilidad no impone ningún vocabulario. El destino se crea antes con el suyo, y el
convertidor casa cada valor de origen con uno de los declarados, con la misma regla de coincidencia que
usa `biso` (plegar mayúsculas, quitar diacríticos, espacios, guiones y guiones bajos). Lo que no case se
omite de la línea y se informa, en lugar de corregirlo por su cuenta.

**Descartado: declarar en el destino los cinco estados de Backlog.md por defecto.** Sirve para el
tablero de este proyecto, pero una utilidad general no puede suponer los estados de un tablero ajeno.

## Los espacios de una etiqueta o un asignado se convierten en guiones

**La decisión.** Backlog.md admite espacios dentro de una etiqueta (`with space`) y de un asignado
(`Sara Smith`), y `biso` no los admite en ninguno de los dos. Al importar, el convertidor transforma
cada tramo de espacios en un guion normal (`with-space`, `Sara-Smith`), quita los del principio y del
final, y avisa de cada conversión. Los demás caracteres que `biso` no admite (`a/b`, `c!`) siguen
quitando el valor, con aviso. La conversión no se deshace al exportar: `with-space` se queda así.

**La pega.** Exportar un tablero importado no devuelve el texto original de esas etiquetas, y dos
etiquetas que solo se diferencian por un espacio o un guion (`a b` y `a-b`) se funden en una.

**Descartado: hacer que `biso` admita espacios en una etiqueta** (era TASK-72, ya descartada). Cambiaba el
modelo de `biso` y reabría su decisión sobre el alfabeto de un token solo para servir a una utilidad de
migración. **Descartado: quitar la etiqueta con aviso.** Perdía un dato que se puede conservar con una
transformación evidente.

**Se aplica también a los asignados.** El encargo hablaba de las etiquetas, pero los asignados tienen el
mismo alfabeto y el mismo problema, y darles un trato distinto no tendría justificación. Pendiente de
confirmar.

## `documentation` y `references` van tal cual

**La decisión.** En `biso` los campos `documentation` y `references` son texto libre, sin restricción de
caracteres, así que las listas de Backlog.md pasan sin comprobación ni cambio. El alfabeto cerrado de
`biso` (letras, dígitos y `- _ . : @`) rige solo para `labels`, `assignees` y las claves de `ext`, y
solo esos campos se validan. Un valor inválido se quita de la lista y se informa, para que un lote no
falle entero por una etiqueta.

## La definición de hecho se marca con el sufijo `#dod`

**La decisión.** `biso` no tiene definición de hecho: al importar, la convierte en criterios de
aceptación y ya no se distinguen. Para que `export` pueda devolverlos a su sección, el convertidor los
añade como criterios con el texto seguido de un espacio y `#dod`. En la exportación, un criterio cuyo
texto acaba en ` #dod` va a la definición de hecho, sin el sufijo, y uno que no lo lleva se queda como
criterio de aceptación. Es una convención de texto en la frontera con otra herramienta, no un campo
nuevo, y no contradice la decisión de `biso` de tener una sola lista de comprobación.

**La pega.** Un criterio que acabe por casualidad en ` #dod` se exportará como definición de hecho.
Se acepta porque es un sufijo muy poco probable en un criterio real.

**Descartado: dejar que `biso` convierta `definitionOfDone` por su cuenta.** Es lo que ya hace, pero no
deja huella para volver, y el objetivo es que exportar e importar de nuevo conserve todo.
**Descartado: perder la distinción.** Que los tableros con los que se midió no usen la definición de
hecho no dice nada de otros tableros, y la utilidad es general.

## El formato de Backlog.md se mide con su CLI, no con un tablero

**La decisión.** Lo que la especificación dice de cómo guarda Backlog.md sus campos, sus secciones y sus
ids se ha comprobado creando un proyecto de prueba con el propio CLI de Backlog.md 1.52.0 y usando todas
las opciones de `backlog task create`, `edit`, `archive` y `complete`, y leyendo los ficheros que
resultan. Los ficheros de prueba de la suite se generan igual, con el CLI, y no se escriben a mano. Es
lo que ha descubierto cosas que ningún tablero real habría enseñado: las subtareas con id con punto, el
campo `project`, los comentarios, la definición de hecho por defecto que se copia a cada tarea, la
configuración que puede vivir fuera de la carpeta de datos y el prefijo cuya capitalización de la
configuración no coincide con la de los ids.

**Descartado: deducir el formato del tablero de este proyecto.** Es una muestra de una sola
configuración y de las opciones que alguien usó, y la herramienta es general.

## Nada se pierde en silencio

**La decisión.** Todo lo que el convertidor no puede mapear, cambia o quita, sale como un hallazgo por
la salida de errores, con el fichero y el campo. Incluye una clave del frontmatter o una sección del
cuerpo que no reconoce. El código de salida 5 avisa de que hubo hallazgos y `--strict` permite negarse
a escribir nada si los hay. Quien ejecuta decide.

**Descartado: abortar en el primer hallazgo.** Un solo estado desconocido impediría ver el resto de
problemas del tablero de una sola vez.

## La exportación: mejor esfuerzo, pero reversible

**La decisión (sin implementar; su diseño y su prueba son TASK-7).** `export` escribe un directorio
`backlog/` que Backlog.md abre. Los campos de `biso` que Backlog.md no tiene (comentarios, pregunta
abierta, arrendamiento, `ext`, `modifiedFiles`) se escriben como una sección de Markdown delimitada y
legible por máquina, dentro de la descripción o de las notas, y `import` la reconstruye como campos:
exportar e importar de nuevo conserva todo. Las reglas de arriba se aplican al revés: el prefijo de
`biso` pasa al del Backlog.md de destino con las mismas reglas de colisión, la etiqueta
`milestone:<slug>` vuelve a ser un milestone y el exportador crea el fichero de milestone que falte,
` #dod` vuelve a la definición de hecho y las fechas se escriben al minuto en UTC.

**Riesgo abierto.** Backlog.md pierde las claves de frontmatter que no conoce cuando edita una tarea.
Hay que medir si conserva también una sección desconocida del cuerpo.

**Descartado: reutilizar el formato Markdown de Backlog.md como formato de intercambio propio de
`biso`.** Heredaría sus bugs conocidos, no tiene sitio para lo que `biso` sí modela, y la garantía de
simetría de `biso export` con `biso new --from` no se podría cumplir. Un Markdown propio de `biso` sería
una función nueva de `biso`, no de esta utilidad.
