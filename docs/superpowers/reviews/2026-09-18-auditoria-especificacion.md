# Auditoría de la especificación, 2026-09-18

## Alcance y método

Esta es una auditoría interna de consistencia de `docs/spec/`. No modifica la especificación ni
decide cuál de dos reglas contradictorias debe prevalecer. Su propósito es dejar localizables los
hallazgos antes de convertir los que procedan en cambios normativos.

La revisión se repartió entre tres lectores independientes: comandos, modelo y salidas, y ciclo de
vida del tablero. Además, los comprobadores de recuentos y referencias por número terminaron sin
hallazgos, y `mkdocs build --strict` terminó correctamente. Esas comprobaciones validan estructura y
enlaces, no las contradicciones semánticas de abajo.

## Hallazgos que bloquean una implementación inequívoca

### `--author ""` tiene dos significados incompatibles

`modelo-de-datos/autor.md:11` dice que `--author ""` crea un autor vacío. En cambio,
`valores-de-entrada.md:46-47` dice que el mismo valor es `empty_scalar_value`, con código 3.

Hay que elegir una regla. Si el autor vacío es una excepción válida, debe declararse junto a la regla
general de escalares vacíos. Si no lo es, hay que retirar el ejemplo y conservar `--clear-author`
para borrar un valor existente.

### La renovación del lease contradice `updatedAt`

`modelo-de-datos/index.md:30` exige actualizar `updatedAt` en cada escritura que cambie algo. Pero
`lease.md:37-43` y `cmd/set.md:49` renuevan `leaseExpiresAt` y afirman que ni la tarea ni
`updatedAt` cambian. `leaseExpiresAt` es un campo almacenado de la tarea en
`modelo-de-datos/index.md:32`.

La especificación debe declarar de forma expresa que el heartbeat es una excepción y decidir si se
incluye en `changed`, o bien actualizar `updatedAt` al renovar el lease.

### `biso doctor` de solo lectura escribe en el tablero

`cmd/flags-globales.md:33-36` trata `biso doctor` sin `--fix` como comando de solo lectura. Sin
embargo, el sondeo WAL de `cmd/doctor.md:190-223` crea, escribe 64 KiB, sincroniza y borra un fichero
en el directorio del tablero. Incluso puede dejar una goroutine que lo limpie después de un timeout.

Hay que reclasificar el comando y sus flags, o rediseñar el sondeo para que no escriba en el tablero.
También falta definir el error y su código cuando ese sondeo no puede escribir.

## Contratos globales contradictorios

### Precedencia de `NO_COLOR`

`salida-y-terminal.md:14-19` prescribe que la presencia de `NO_COLOR` impide siempre el color.
`invocacion.md:18-20` prescribe que `--color always` gana y sí lo activa.

La tabla de precedencia debe ser única. Una forma directa es: flag explícito, después `NO_COLOR`, y
después el comportamiento automático.

### Errores de `export --json`

El contrato de `contrato-json.md:33-36` exige un objeto JSON único por `stderr` ante cualquier fallo
con `--json`. `cmd/index.md:39-46` y `cmd/export.md:72,100-101` prescriben en cambio texto literal
para el rechazo de `export --json`.

El rechazo debe usar el sobre de error JSON cuando el flag esté presente, dejando el texto literal
para la llamada sin `--json`, o debe documentarse una excepción deliberada al contrato.

### `doctor --fix --dry-run` no sigue el resultado de `--dry-run`

`cmd/flags-globales.md:23` promete código 0 si una operación en seco habría funcionado y 7 si no.
La misma página dice que `doctor --fix` se comporta como cualquier escritura. No obstante,
`cmd/doctor.md:395-409,452-454` fija códigos 0 o 6, nunca 7.

Hay que declarar una excepción explícita que conserve 0 o 6, o adaptar `doctor` al código 7 de los
dry-runs generales.

### El sobre JSON universal y el de error no tienen la misma forma

`contrato-json.md:5-10` afirma que toda salida con `--json` lleva `data`. El ejemplo de error de
`contrato-json.md:38-51` lleva en su lugar `error`. Más adelante la página los gobierna por separado,
pero no corrige la promesa universal inicial.

Conviene definir desde el encabezado dos sobres disjuntos: éxito con `data` y error con `error`.

### Los avisos JSON de `ls` no tienen dónde ir

`salida-y-terminal.md:51-54` dice que todo `warning:` de éxito con `--json` viaja estructurado en
`data.warnings`. `ls` puede emitir `list_truncated` según `salida-y-terminal.md:77`, pero
`contrato-json.md:17` y `cmd/ls.md:205-248` no incluyen `warnings` en `task.list`.

Se debe añadir `warnings` al esquema de los kinds que puedan avisar, o restringir la regla general a
los kinds que sí lo exponen.

## Restauración y simetrías incompletas

### `init --from` necesita un fichero que no forma parte de la entrada definida

`cmd/init.md:288-294` dice que la restauración toma la identidad de `<id>.id`. Sin embargo,
`cmd/init.md:170-172,247,378` define la instantánea solo con `snapshot.ndjson` y `board.json`.

Debe exigirse y validarse `<id>.id`, incluido su error y la comprobación en `--dry-run`, o debe
incluirse la identidad validada en `board.json`.

### `export` y `new --from` no definen la entrada `null`

El contrato JSON representa los opcionales sin valor como `null` en `contrato-json.md:125-137`.
`cmd/export.md:30-36` promete emitir exactamente el formato que acepta `new --from`, pero
`cmd/new.md:85-143` no define cómo importar `null` para escalares opcionales, colecciones, mapas o
`question`.

Hace falta una tabla de coerción de NDJSON: por ejemplo, `null` como ausencia para opcionales, `[]` y
`{}` obligatorios para colecciones, y rechazo de `null` en campos obligatorios.

### La simetría de `export` requiere una configuración de destino no declarada

`cmd/export.md:30-36,82-85` llama a `export` entrada exacta de `new --from`, pero un tablero destino
necesita vocabularios, prefijo y extensiones compatibles. La simetría tablero completo que sí tiene
procedimiento definido es `snapshot` con `init --from`, en `cmd/export.md:44-52`.

Hay que separar ambos contratos: `export` a `new --from` bajo precondiciones de configuración
compatibles, y `snapshot` a `init --from` para una reconstrucción completa.

## Referencias y forma observable de la salida

### `ls --parent` no puede forzar el tipo de referencia

`referencias.md:14-17` da `--id` y `--match` a todos los comandos que aceptan una referencia.
`cmd/ls.md:9,31,295` acepta `--parent <ref>`, pero no documenta cómo aplicar esos flags al valor.

Hay que limitar la garantía a referencias posicionales o añadir una forma de forzar la interpretación
en cada flag cuyo valor sea `<ref>`.

### `get --section` no define duplicados ni orden

`cmd/get.md:16,226-228` admite repetir `--section` y listas separadas por coma, pero no define qué
ocurre al pedir la misma sección varias veces ni el orden de salida de una selección reordenada.

La especificación debe fijar deduplicación y orden, tanto para texto como JSON. El orden canónico de
secciones o el orden de primera aparición son alternativas comprobables.

### El esquema de `task.get` no es completo

`contrato-json.md:125-127` exige que toda clave documentada para un kind aparezca siempre.
`cmd/get.md:157` define `task.get` como el objeto de `task.list` más el cuerpo, pero su ejemplo en
`cmd/get.md:165-179` omite claves básicas y no publica una tabla completa de `data.task`.

Debe publicarse el esquema íntegro y hacer el ejemplo conforme, dejando `--section` como la única
reducción declarada.

### Listas derivadas sin orden estable

`modelo-de-datos/index.md:96` define `blocks` como lista de identificadores, y
`modelo-de-datos/urgencia.md:95-98` explica cuáles entran, pero no en qué orden.

La salida JSON estable requiere fijar el orden, por ejemplo por identificador ascendente, y qué hacer
si la tarea que determina una dependencia no se puede decodificar.

## Configuración y límites de comandos

### `init --at` y el puntero discrepan sobre `path`

`cmd/init.md:67-70,229-232` obliga a escribir `path` para cualquier uso de `--at`.
`resolucion-del-tablero.md:213-218` dice que `path` solo aparece si el tablero queda fuera de las
raíces configuradas.

Hay que elegir una regla única. Una opción es escribir `path` cuando se usó `--at` y omitirla solo en
la ubicación predeterminada implícita.

### El presupuesto general contradice excepciones propias

`presupuestos.md:5-8` presenta 25 ms para `ls` y `prime` como regla transversal y niega que otro
comando tenga motivo para durar más. `cmd/doctor.md:215-221` permite un sondeo de hasta 2 s, y
`presupuestos.md:42-50` ya exceptúa `snapshot` por procesos externos.

La redacción debe limitar el presupuesto al camino caliente de `ls` y `prime`, o enumerar de forma
expresa las excepciones.

### `board --no-open` es incompatible con su uso anunciado en scripts

`cmd/board.md:17-23` exige un terminal aun con `--no-open`. `cmd/board.md:67-71` justifica el JSON
para un script que arranca el servidor con ese flag, aunque un script normalmente no tiene TTY.

Hay que permitir un modo headless con `--no-open`, o retirar y aclarar esa promesa.

## Precisión adicional recomendada

- ~~`modelo-de-datos/autor.md:3-4` e `index.md:77-79` presentan el autor como inmutable tras
  crear~~ resuelto: `autor.md` nombra ahora las dos vías de cambio, `--author` y `--clear-author`.
- `modelo-de-datos/index.md:6-14` llama `string` a texto de una línea, pero no fija una validación de
  CR y LF. **Pendiente de decidir antes de tocar la spec**: confirmado que `Comment.body` y
  `Question.body` ya son `text` (permiten salto de línea) y no entran en esto; los campos que sí son
  `string` y quedarían sujetos a la regla son `title` e `author` de la tarea (`index.md:51,57`,
  `autor.md`), `Comment.author` (`comentarios.md:8`), `Question.author` (`pregunta-abierta.md:8`),
  `Criterion.text` de `acceptanceCriteria`/`definitionOfDone` (`criterios.md:9`), y los valores de
  `ext`, que es `map<string,string>` (`index.md:65`). La regla de alfabeto ya existente para
  labels/assignees/claves de `ext` vive en `valores-de-entrada.md#el-juego-de-caracteres-de-un-token`
  y sería el sitio natural para una regla de CR/LF análoga, si se decide añadirla.
- ~~`modelo-de-datos/pregunta-abierta.md:3-10` dice que `Question` tiene la misma forma que
  `Comment`~~ resuelto: acotada a autor, instante y cuerpo, sin `key`.
- ~~`contrato-json.md:15` omite `discarded` del objeto `where`~~ resuelto: añadido a la tabla de kinds.
- **Pendiente, trabajo de diseño más grande**: `task.list.filters` solo tiene un ejemplo con tres
  claves (`status`, `type`, `label`), sin definición textual de qué lleva en general. `ls` tiene más
  de una decena de filtros sin sitio en ese esquema: `--priority`, `--assignee`/`--mine`/`--unassigned`,
  `--parent`, `--blocked`/`--not-blocked`, `--waiting`/`--not-waiting`, `--active`/`--not-active`,
  `--overdue`, `--due-before`, `--search`, `--unchecked`, `--any-status`,
  `--archived`/`--only-archived`, `--not-status`, `--label-or`. Hace falta decidir, para cada uno, su
  clave en `data.filters` y su forma (booleano, valor único, lista), antes de escribirlo en
  `contrato-json.md` y en el ejemplo de `cmd/ls.md`.
- ~~`cmd/init.md:176-178` aún nombra `projects` en `board.json`~~ resuelto: quitado de la lista, ya
  que el campo está fuera de alcance.

## Resolución, 2026-09-19

Los diecinueve hallazgos y los cinco puntos de precisión adicional se revisaron uno por uno contra el
texto citado, con confirmación explícita antes de cada cambio. Diecisiete hallazgos y tres puntos de
precisión quedaron resueltos en el propio texto de la especificación; dos puntos de precisión quedan
pendientes de una decisión de diseño más amplia (marcados arriba), y su detalle vive también en el
catálogo de fuentes únicas de
["Los principios, y cómo se mantiene la especificación"](../../decisiones/principios-y-mantenimiento.md#el-catálogo-de-datos-que-viven-en-más-de-un-sitio).

| Hallazgo | Resuelto en |
|---|---|
| `--author ""` | [`modelo-de-datos/autor.md`](../../spec/modelo-de-datos/autor.md) |
| Renovación del lease y `updatedAt` | [`modelo-de-datos/index.md`](../../spec/modelo-de-datos/index.md#campos-automáticos) |
| `doctor` de solo lectura y su sondeo | [`cmd/flags-globales.md`](../../spec/cmd/flags-globales.md) |
| Precedencia de `NO_COLOR` | [`salida-y-terminal.md`](../../spec/salida-y-terminal.md) |
| Errores de `export --json` (y `prime --full`, `config`) | [`contrato-json.md`](../../spec/contrato-json.md#los-errores-en-json), [`cmd/export.md`](../../spec/cmd/export.md), [`cmd/prime.md`](../../spec/cmd/prime.md), [`cmd/config.md`](../../spec/cmd/config.md) |
| `doctor --fix --dry-run` y el código 7 | [`cmd/flags-globales.md`](../../spec/cmd/flags-globales.md) |
| Sobre JSON universal y de error | [`contrato-json.md`](../../spec/contrato-json.md#el-sobre) |
| `warnings` de `ls` en JSON | [`cmd/ls.md`](../../spec/cmd/ls.md#el-esquema-json), [`contrato-json.md`](../../spec/contrato-json.md) |
| `init --from` y `<id>.id` | [`cmd/init.md`](../../spec/cmd/init.md) |
| `null` en el lote de `new --from` | [`cmd/new.md`](../../spec/cmd/new.md#el-modo-lote), [`cmd/export.md`](../../spec/cmd/export.md) |
| Simetría de `export`/`new --from` | [`cmd/export.md`](../../spec/cmd/export.md#la-garantía-de-simetría) |
| `ls --parent` y el forzado de referencia | [`referencias.md`](../../spec/referencias.md) |
| `get --section`, duplicados y orden | [`cmd/get.md`](../../spec/cmd/get.md) |
| Esquema de `task.get` | [`cmd/get.md`](../../spec/cmd/get.md#el-esquema-json) |
| Orden de `blocks` | [`modelo-de-datos/index.md`](../../spec/modelo-de-datos/index.md#los-campos-derivados) |
| `init --at` y `path` en el puntero | [`resolucion-del-tablero.md`](../../spec/resolucion-del-tablero.md) |
| Presupuesto de 25 ms y sus excepciones | [`presupuestos.md`](../../spec/presupuestos.md) |
| `board --no-open` sin terminal | [`cmd/board.md`](../../spec/cmd/board.md) |
