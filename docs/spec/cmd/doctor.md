# `biso doctor`

## Firma

```
biso doctor [--fix]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--fix` | | no | booleano | falso | no | no | ninguno |

**`biso doctor` sin `--fix` es de solo lectura**, y `--print` y `--dry-run` de la sección ["Banderas globales"](../invocacion.md#banderas-globales) son error 2
igual que en cualquier otro comando de lectura. **Con `--fix` es un comando de escritura**: ahí
`--dry-run` reporta qué se repararía sin reparar nada, y `--print` no añade nada, porque `doctor` no
imprime fichas de tareas.

## Para qué sirve `biso doctor`, y para qué no

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
- **Los comandos de lectura no pueden escribir.** La sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables) promete que las lecturas nunca
  fallan por una escritura en curso y nunca la bloquean, y la sección ["Qué hace, caso a caso"](prime.md#qué-hace-caso-a-caso) que `biso prime` no escribe nunca y
  es seguro en paralelo. Un `ls` o un `prime` que repararan al pasar necesitarían acceso exclusivo,
  podrían esperar cinco segundos y fallar con código 7, y se perdería justo la garantía de los dos
  comandos que un agente llama sin parar.
- **Un comando de escritura tiene permiso para lo que se le pidió, no para más.** Y no es solo
  cuestión de sorpresa: con el todo o nada de la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables), si `biso set` cambiara un título y
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

Y de ahí sale el corolario que hace falta para que las reglas no se peleen: **un comando que
tropieza con un problema reparable que no le toca arreglar lo dice con un `warning:` y nombra `biso
doctor --fix`.** Eso no es remitir a una llamada que se iba a hacer igual, que es justo lo que se
prohíbe: es contar algo que quien llama no sabía y que no iba a descubrir por su cuenta. La regla que
se prohíbe es remitir a `doctor` para algo que el comando ya tenía permiso para arreglar; avisar de lo
que no puede tocar es lo contrario de esconderlo.

Y queda una tercera razón para que `doctor` exista, que no es de reparación: es el único sitio donde
se puede preguntar **"le pasa algo a este tablero"** sin haber intentado antes una operación. Eso es
un diagnóstico que se corre cuando se sospecha, no una llamada que se iba a hacer igual.

**Ningún comando de `biso` pregunta nada por la entrada estándar, ni `doctor` con `--fix` ni ninguno
otro** (sección ["Interactividad, terminal y color"](../salida-y-terminal.md#interactividad-terminal-y-color)). La misma orden sirve para una persona y para un agente, y lo único que cambia es
quién lee la salida: la persona lee el informe, el agente mira el código de salida.

## Errores y avisos, y solo los errores sacan el código 6

Lo que `doctor` reporta se divide en dos niveles:

- **Errores**: dejan el tablero inconsistente, o hacen imposible una operación, o vuelven un dato poco
  fiable. Si queda alguno sin reparar, el código de salida es 6.
- **Avisos**: son verdad, merece la pena saberlos, y no rompen nada. **No cambian el código de
  salida.**

**Se llaman avisos, pero no usan el prefijo `warning:`.** Ese prefijo es de stderr: la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos)
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
sección ["stdout, stderr y qué va en cada uno"](../salida-y-terminal.md#stdout-stderr-y-qué-va-en-cada-uno) dice que stdout lleva lo que un programa consumiría, y el informe es exactamente eso: es
el resultado que se ha pedido, no un mensaje que acompaña a otro trabajo. Los avisos y las notas de
la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) son mensajes que un comando emite al lado de lo que produce; los hallazgos de `doctor`
son lo que produce, así que stdout le corresponde por la regla general, no aparte de ella. La sección
ya lo hacía así antes de esta tarea, cuando un tablero limpio imprime `no problems found` por stdout:
aquí no se cambia nada, se explica lo que ya era. Los avisos de los demás comandos, los que lista la
tabla cerrada de la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos), siguen yendo por stderr sin cambiar, y el informe de `doctor` no le
añade ninguna fila ni reutiliza su prefijo `warning:`.

## Qué comprueba

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
| El identificador más alto que el tablero recuerda haber asignado (["Identificadores"](../modelo-de-datos.md#identificadores)) es menor que el identificador más alto de una tarea existente | error | sí |
| Falta el marcador `<id>.id` en el directorio del tablero (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)) | error | sí, escribiéndolo con el `id` que lleva la base de datos |
| El marcador `<id>.id` nombra un `id` distinto del que lleva la base de datos | error | no, hay que decidir a mano |
| Una raíz de `boards_extra_roots` (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)) no existe o no se puede leer | aviso | no, es configuración de la máquina o un disco sin montar |
| La comprobación de integridad de la base de datos falla (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | error | no, es daño externo; la reparación es restaurar de una copia |
| El directorio del tablero está en un sistema de ficheros donde el modo WAL de SQLite no es seguro | aviso | no, es una propiedad del sistema de ficheros, no algo que `biso` pueda cambiar |
| Huecos en la numeración | no es un problema | no son un problema, no se reportan |

**No hay ninguna comprobación sobre el nombre de la carpeta del tablero, y no es un olvido.** El nombre
es decorativo y nadie resuelve por él (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), así que una carpeta con el nombre de un `project_name`
anterior, o con un nombre que alguien puso a mano, no es un problema del que informar. Denunciarlo sería
denunciar algo que la sección ["`biso config`"](config.md) permite explícitamente.

**La fila del arrendamiento repara en una sola dirección, y por eso `--fix` la hace solo.** Los dos
campos son lo que sobra, y el estado y la lista de personas asignadas son el dato: vaciarlos deja la
tarea exactamente como estaba, mientras que arreglarla al revés, poniéndola activa o asignándole a
alguien para justificar el arrendamiento, cambiaría el trabajo del tablero para salvar una reserva que
ya no vale. La invariante que comprueba es la de la séptima precisión de la sección ["El modelo de datos de una tarea"](../modelo-de-datos.md), la misma que
`biso new --from` aplica al importar (["`biso new`"](new.md)), y cae en el segundo de los dos casos de arriba: ninguna
escritura de `biso` la puede romper, así que si un tablero llega a ese estado es por daño externo, como
una base de datos escrita a mano, restaurada a medias o venida de otra versión. Sin `--fix` sale bajo
`Errors:`:

```
  MYP-52  has a lease but is not both active and assigned
```

Y con `--fix`, en el grupo de lo reparado:

```
  MYP-52 had a lease but was not both active and assigned; cleared leaseExpiresAt and leaseHolder
```

Su `code` en el JSON es `lease_invariant` en los dos sitios.

**Las filas del marcador se parecen y se reparan al revés.** Que falte tiene una
sola lectura posible: el `id` de verdad es el que lleva la base de datos, y el marcador es su copia en el
sistema de ficheros, así que escribirlo con ese valor no puede equivocarse y `--fix` lo hace solo. Que
discrepe no tiene una sola lectura: reescribir el marcador con el `id` de la base de datos dejaría de
resolver a todos los punteros que nombran el `id` viejo, y reescribir la base de datos cambiaría la
identidad del tablero. Las dos direcciones pierden algo, así que se decide a mano, con el mismo criterio
que los identificadores duplicados de la primera fila.

**La comprobación de integridad de la base de datos y el aviso del sistema de ficheros son las dos
comprobaciones que añade esta misma decisión de persistencia**, y caen cada una en uno de los dos
casos de arriba. La primera es daño externo puro, sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar): nada dentro de `biso` corrompe su
propia base de datos, así que no hay ninguna dirección que reparar por su cuenta, y por eso es un
error, no un aviso, aunque tampoco sea reparable. La segunda no reporta un daño ya hecho, sino un
riesgo: en ese sistema de ficheros el modo WAL de SQLite no ofrece las garantías de atomicidad que la
sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables) exige, pero el tablero de hoy puede estar perfectamente sano. Eso es exactamente lo que
distingue a un aviso de un error, así que es aviso.

**Y la primera tiene una peculiaridad que la separa de las demás filas de error: nunca aparece como
una línea del informe.** Las demás comprobaciones de error sí producen una entrada en la lista
de problemas cuando se disparan, pero esta no, porque cuando se dispara no hay informe de `doctor`
que mostrarla: hay el abort completo de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), con su propio mensaje y su propio código 10,
antes de que `doctor` llegue a comprobar nada más (ver la tabla de comportamiento más abajo). La fila
está en esta tabla para decir que existe como comprobación y cuál es su nivel, no porque vaya a
verse alguna vez junto a las demás.

Con esto, todas las filas de la tabla son un problema real salvo los huecos en la numeración, que no
lo son y no se reportan nunca. Y de las que sí lo son, solo la comprobación de integridad de la base de
datos no llega a aparecer nunca como una línea del informe, por la razón de arriba.

## Atomicidad de `--fix` con varias reparaciones

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

Y esto no rompe la garantía de la sección ["Concurrencia, atomicidad y garantías observables"](../garantias.md#concurrencia-atomicidad-y-garantías-observables), porque esa sección promete sobre las escrituras del
tablero, es decir, sobre sus datos, y el marcador no es un dato del tablero: es una copia de su `id` en
el sistema de ficheros, puesta ahí para poder encontrarlo sin abrirlo (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)). Por eso el orden
importa y hay que decirlo explícito: primero la transacción de datos, después el marcador. Si el
marcador falla, las reparaciones de datos quedan hechas y son definitivas, el comando termina con el
código de no poder escribir (7, el mismo de cualquier otro fallo de entorno al reparar), y la falta del
marcador vuelve a aparecer como error la próxima vez que se ejecute `doctor`, porque sigue siendo verdad.
Ese desenlace es coherente consigo mismo: no hay ningún dato del tablero observado a medias, y lo único
que queda pendiente es una reparación que ya se sabe cómo repetir.

## Comportamiento, caso a caso

| Caso | Qué pasa |
|---|---|
| Tablero limpio | `no problems found` por stdout, código 0 |
| Solo avisos, sin ningún error | Se reportan bajo `Warnings:`, código 0 |
| Solo errores reparables, con `--fix` | Se reparan y se reporta cada uno, código 0 |
| Quedan errores sin reparar | Código 6, aunque se haya reparado algo o se hayan reportado avisos |
| Una tarea ilegible (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | Se reporta como error y se sigue con las demás. **Nunca aborta** |
| La base de datos no se puede leer (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | El comando entero aborta con el mensaje y el código 10 de ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), antes de comprobar nada más |
| `--fix` sin poder escribir | Código 7. Si falla la escritura del marcador después de la transacción de datos, esta ya quedó aplicada (ver arriba) |
| `--fix --dry-run` | Reporta qué se repararía, sin reparar nada, código 0 |

## Salida

```
2 errors found, 1 warning found
Errors:
  MYP-40  dependency MYP-99 does not exist
Warnings:
  extra board root "/Volumes/disco/boards" cannot be read (skipped when looking up boards by id)
1 error fixed
  the highest recorded id was MYP-40 and tasks go up to MYP-52; recorded MYP-52
```

Los errores y los avisos se agrupan bajo su propio encabezado, `Errors:` y `Warnings:`; ninguno de
los dos usa el prefijo `warning:` de stderr, que la sección ["Notas y avisos"](../salida-y-terminal.md#notas-y-avisos) reserva para lo que va por ahí. Un
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

## El esquema JSON

```json
{
  "schemaVersion": 1,
  "kind": "doctor",
  "generatedAt": "2026-09-06T09:12:04Z",
  "data": {
    "problems": [
      { "task": "MYP-40", "code": "dependency_not_found", "message": "dependency MYP-99 does not exist" }
    ],
    "warnings": [
      { "task": null, "code": "extra_root_unreadable", "message": "extra board root \"/Volumes/disco/boards\" cannot be read (skipped when looking up boards by id)" }
    ],
    "fixed": [
      { "code": "highest_id_behind", "message": "the highest recorded id was MYP-40 and tasks go up to MYP-52; recorded MYP-52" }
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

## Códigos de salida

| Desenlace | Código |
|---|---:|
| Nada mal, o todo lo encontrado se ha reparado | 0 |
| Quedan errores | 6 |
| Sintaxis | 2 |
| No se puede escribir al reparar | 7 |
| No hay tablero | 8 |
| Su base de datos no se puede leer (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | 10 |

## `biso doctor --help`

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

