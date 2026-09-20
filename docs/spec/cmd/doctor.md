# `biso doctor`

## Firma

```
biso doctor [--fix]
```

| Parámetro | Corto | Oblig. | Tipo | Por defecto | Repetible | Lista | Incompatible con |
|---|---|---|---|---|---|---|---|
| `--fix` | | no | booleano | falso | no | no | ninguno |

**`biso doctor` sin `--fix` es de solo lectura**, y `--print` y `--dry-run` de la sección ["Flags globales"](flags-globales.md#flags-globales) son error 2
igual que en cualquier otro comando de lectura. **Con `--fix` es un comando de escritura**: ahí
`--dry-run` es una vista previa pura de lo que haría la llamada real que le sigue, con el mismo informe
y el mismo código de salida de la tabla de más abajo (nunca un 7 propio, que `doctor` no tiene, sección
["Flags globales"](flags-globales.md#flags-globales)), sin reparar nada; y `--print` sigue siendo
error 2, con `--fix` y sin él, porque `doctor` no imprime fichas de tareas y lo que `--fix` repara
(un arrendamiento que ninguna tarea justifica, un contador y un fichero de marcador) no es ninguna
tarea de la que imprimir una. Es la regla general de que `--print` y `--dry-run` no se ignoran nunca
en silencio, aplicada al único comando donde `--fix` podría hacer dudar.

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
  podrían esperar cinco segundos y fallar con código 8, y se perdería justo la garantía de los dos
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
| Estados, tipos o prioridades que ya no están configurados | error | no |
| `initial_status`, `active_status` o `terminal_status` que no están en `statuses` | error | no |
| `statuses` con menos de tres elementos, o dos de los tres papeles apuntando al mismo estado | error | no |
| Dependencias que apuntan a tareas inexistentes | error | no |
| Ciclos de dependencias | error | no |
| Ciclos de tarea padre | error | no |
| Claves de criterio repetidas dentro de una tarea | error | no |
| `leaseExpiresAt` o `leaseHolder` con valor en una tarea que no está a la vez en el estado activo y asignada, o uno de los dos con valor y el otro vacío | error | sí, vaciando los dos |
| El identificador más alto que el tablero recuerda haber asignado (["Identificador de tarea"](../modelo-de-datos/identificadores.md#identificador-de-tarea)) es menor que el identificador más alto de una tarea existente | error | sí |
| Falta el marcador `<id>.id` en el directorio del tablero (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)) | error | sí, escribiéndolo con el `id` que lleva la base de datos |
| El marcador `<id>.id` nombra un `id` distinto del que lleva la base de datos | error | no, hay que decidir a mano |
| Una raíz de `boards_extra_roots` (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)) no existe o no se puede leer | aviso | no, es configuración de la máquina o un disco sin montar |
| La comprobación de integridad de la base de datos falla (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | error | no, es daño externo; la reparación es restaurar de una copia |
| El directorio del tablero está en un sistema de ficheros donde el modo WAL de SQLite no es seguro | aviso | no, es una propiedad del sistema de ficheros, no algo que `biso` pueda cambiar |
| El fichero de exclusión no corresponde al `vcs` configurado en la máquina | aviso | no, hay que escribirlo a mano |
| Huecos en la numeración | no es un problema | no son un problema, no se reportan |

**No hay ninguna comprobación sobre el nombre de la carpeta del tablero, y no es un olvido.** El nombre
es decorativo y nadie resuelve por él (sección ["Cómo se elige el tablero"](../resolucion-del-tablero.md)), así que una carpeta con el nombre de un `project_name`
anterior, o con un nombre que alguien puso a mano, no es un problema del que informar. Denunciarlo sería
denunciar algo que la sección ["`biso config`"](config.md) permite explícitamente.

**La fila del arrendamiento repara en una sola dirección, y por eso `--fix` la hace solo.** Los dos
campos son lo que sobra, y el estado y la lista de personas asignadas son el dato: vaciarlos deja la
tarea exactamente como estaba, mientras que arreglarla al revés, poniéndola activa o asignándole a
alguien para justificar el arrendamiento, cambiaría el trabajo del tablero para salvar una reserva que
ya no vale. La invariante que comprueba es la de ["El vaciado"](../lease.md#el-vaciado) de `lease.md`, la misma que
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

### El sondeo del sistema de ficheros

**El aviso se dispara por un sondeo funcional en el propio directorio del tablero, no por el nombre o
el tipo del sistema de ficheros.** `doctor` comprueba en caliente los dos primitivos que el modo WAL de
SQLite necesita, un bloqueo por rango de bytes fiable y un `mmap` compartido que de verdad sincroniza
con una lectura corriente, en vez de adivinarlos por una tabla de sistemas de ficheros conocidos, que
se quedaría corta ante un FUSE nuevo o un disco de red que se anuncia con un tipo normal. El controlador
es `modernc.org/sqlite`, sin `cgo` (raíz de este repositorio), así que el sondeo usa
`golang.org/x/sys/unix` en macOS y Linux y `golang.org/x/sys/windows` en Windows, las dos bibliotecas
puras en Go que ya evitan el mismo problema de compilación cruzada que evita el controlador.

El protocolo, en el directorio del tablero (donde vive `board.db`, no en un directorio temporal del
sistema, porque lo que importa es el sistema de ficheros de ese directorio en concreto):

1. **Crea un fichero de prueba**, `.biso-wal-probe-<pid>-<aleatorio>`, con permisos `0600`, le escribe
   64 KiB de ceros y hace `fsync`. Si crear o escribir el fichero falla (por ejemplo `EACCES` o
   `EROFS`), el sondeo entero se marca como fallido sin seguir a los pasos siguientes: si `biso` no
   puede ni escribir ahí, no hay manera de comprobar nada más, y esa incapacidad ya es en sí misma un
   sistema de ficheros donde WAL no puede funcionar.
2. **Bloqueo por rango de bytes**, en dos mitades, porque son dos preguntas distintas.

   La primera es si este sistema de ficheros tiene bloqueos por rango de bytes: toma un bloqueo
   exclusivo sobre el byte `[0, 1)` con `unix.FcntlFlock` (`F_SETLK`) en Unix, o con `LockFileEx` y
   un rango de bytes en Windows, y lo libera. Es el mismo primitivo que toma SQLite, así que
   cualquier error aquí (`ENOLCK`, `ENOSYS`, `EINVAL`, o el equivalente de Windows) es ya un
   sistema de ficheros donde el controlador no puede hacer su propio bloqueo, y el sondeo falla sin
   seguir.

   La segunda es si ese bloqueo se hace cumplir de verdad entre dos descriptores independientes, y
   **no se puede preguntar con `F_SETLK`**: un bloqueo de registro de POSIX pertenece al proceso y
   no al descriptor, así que el mismo proceso que lo pide dos veces lo recibe las dos, por
   definición y no por un fallo del sistema de ficheros. Un sondeo escrito así diría que todos los
   sistemas de ficheros del mundo son inseguros. La pregunta se hace con el bloqueo por rango de
   bytes ligado al descriptor abierto, `F_OFD_SETLK` en Linux y en macOS, que es el mismo bloqueo
   con otra pertenencia: abre una segunda vez el mismo fichero con un descriptor independiente,
   toma el bloqueo `[0, 1)` desde el primero, intenta el mismo desde el segundo y tiene que fallar
   con `EAGAIN`/`EACCES`. Si en cambio se concede, el sistema de ficheros no está haciendo cumplir
   el bloqueo, que es el síntoma clásico de un NFS que concede bloqueos en local sin coordinarlos
   con el otro extremo. En Windows la pregunta se hace con `LockFileEx`, cuyos bloqueos ya son del
   manejador y no del proceso, así que ahí no hace falta ninguna variante. Libera los dos bloqueos
   y cierra el segundo descriptor.

   **En una plataforma sin bloqueos ligados al descriptor abierto** (los BSD, que no tienen
   `F_OFD_SETLK`) esta segunda mitad se salta y el sondeo se queda con la primera. No inventarse un
   fallo es lo correcto: el aviso dice que el sondeo corrió y algo no funcionó, y una comprobación
   que no se puede hacer no es una que haya salido mal.
3. **`mmap` compartido con una escritura visible fuera del mapeo.** Mapea el fichero con
   `unix.Mmap` (`MAP_SHARED`, `PROT_READ|PROT_WRITE`) en Unix, o `CreateFileMapping` /
   `MapViewOfFile` (`PAGE_READWRITE`, `FILE_MAP_WRITE`) en Windows. Escribe un patrón de 8 bytes al
   principio del mapeo, lo sincroniza (`msync(MS_SYNC)` en Unix, `FlushViewOfFile` seguido de
   `FlushFileBuffers` en Windows) y vuelve a leer esos mismos 8 bytes con una lectura corriente del
   fichero, no a través del mapeo. Si el `mmap` en sí falla (`ENODEV`, `ENOTSUP`, `EOPNOTSUPP`), o si la
   lectura corriente no ve el patrón que se acaba de escribir, el sondeo se marca como fallido: es la
   forma en que un `MAP_SHARED` que en realidad no comparte memoria con el fichero se delata.
4. **Tope de tiempo.** Los tres pasos anteriores corren dentro de un único plazo de 2 segundos. Ninguna
   de las llamadas del sistema que usan tiene un timeout propio, así que el sondeo entero se lanza en su
   propia goroutine y se espera con `context.WithTimeout`; si el plazo se cumple antes de que la
   goroutine termine (un montaje de red colgado en `fcntl` es el caso real que esto cubre), el sondeo
   se da por fallido y `doctor` sigue sin esperar a que la goroutine acabe alguna vez, porque Go no
   puede cancelar una llamada al sistema bloqueada. La goroutine abandonada borra el fichero de prueba
   por su cuenta si es que llega a terminar; `doctor` no depende de que lo haga.
5. **Limpieza.** Si el sondeo termina dentro del plazo, borra el fichero de prueba antes de devolver el
   resultado, se haya marcado como seguro o como fallido.

**Cualquier fallo o comportamiento degradado de los pasos 1 a 4 dispara el aviso**: no crearse el
fichero, no hacer cumplir el bloqueo, que el `mmap` falle o que su escritura no se vea desde una lectura
corriente, o que el plazo se cumpla. Solo cuando los tres pasos funcionales pasan dentro del plazo se
considera el sistema de ficheros seguro para WAL y el aviso no sale.

**El sondeo puede dar un falso negativo, y esta sección no lo esconde.** Un solo proceso en un solo
host no puede reproducir la condición de carrera entre dos hosts distintos escribiendo el mismo
`board.db` por NFS, que es la que de verdad rompe WAL en red; el sondeo detecta la falta de soporte
local de los dos primitivos, no la ausencia de coordinación entre extremos. Es la misma limitación que
ya reconoce la decisión de esta tarea en `docs/decisiones/`, y el motivo por el que sigue siendo aviso
y no error: un sondeo que pasa no es una garantía, es la ausencia de la señal más barata de comprobar.

**Y la primera tiene una peculiaridad que la separa de las demás filas de error: nunca aparece como
una línea del informe.** Las demás comprobaciones de error sí producen una entrada en la lista
de problemas cuando se disparan, pero esta no, porque cuando se dispara no hay informe de `doctor`
que mostrarla: hay el abort completo de la sección ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), con su propio mensaje y su propio código 21,
antes de que `doctor` llegue a comprobar nada más (ver la tabla de comportamiento más abajo). La fila
está en esta tabla para decir que existe como comprobación y cuál es su nivel, no porque vaya a
verse alguna vez junto a las demás.

**El fichero de exclusión desactualizado no se repara con `--fix`, por el mismo motivo que ya fija la
sección ["`biso init`"](init.md).** `vcs` es configuración de la máquina, no del
tablero (sección ["Configuración de máquina"](../invocacion.md#configuración-de-máquina)), así que el
fichero que un tablero tiene que llevar depende de qué esté configurado ahora, no de un historial que
`biso` no guarda. El único nombre que la especificación fija de antemano es `.gitignore`, el de `git`;
el de `custom` lo declara `ignore_file`, un valor arbitrario que la máquina puede haber cambiado o
dejado de declarar, y `none` no espera ninguno. Por eso `doctor` solo puede reconocer con certeza un
caso: que exista `.gitignore` en el directorio del tablero y que el `vcs` de la máquina ya no sea `git`
ni un `custom` cuyo `ignore_file` sea también `.gitignore`. El caso contrario, un `ignore_file` de un
`custom` anterior que quedó huérfano, no es detectable, porque `biso` no recuerda cuál era ese nombre
antes de que `vcs` cambiara.

Es aviso, no error, por la misma razón que la raíz adicional que no se puede leer: no dice que el
tablero esté roto ni que ningún dato de `biso` sea poco fiable, dice que hay un fichero de la máquina
que se quedó con un nombre antiguo. Y no es reparable con `--fix`, porque la sección ["`biso init`"](init.md)
ya deja dicho que `biso` nunca renombra ni escribe otro fichero de exclusión por su cuenta cuando `vcs`
cambia: `doctor --fix` sería incoherente con esa misma decisión si lo hiciera. Se decide a mano, igual
que la fila del marcador `<id>.id` discrepante. Sale bajo `Warnings:`:

```
  .gitignore does not match the configured vcs "none" (left over from git, biso does not rewrite it automatically)
```

Su `code` en el JSON es `ignore_file_mismatch`, con `task` a `null`, porque es un hallazgo del tablero
y no de ninguna tarea, igual que la raíz adicional que no se puede leer.

Con esto, todas las filas de la tabla son un problema real salvo los huecos en la numeración, que no
lo son y no se reportan nunca. Y de las que sí lo son, solo la comprobación de integridad de la base de
datos no llega a aparecer nunca como una línea del informe, por la razón de arriba.

### El `code` y el mensaje de cada comprobación

**Cada fila de la tabla de arriba, salvo los huecos en la numeración, tiene un `code`** en `snake_case`
e inglés, que describe el hallazgo y no el mecanismo que lo detecta, con el mismo estilo que los cinco
que ya existían: `dependency_not_found`, `lease_invariant`, `highest_id_behind`, `extra_root_unreadable`
e `ignore_file_mismatch`. El mensaje sigue la misma regla que cualquier aviso accionable de esta sección:
dice qué hay y qué se esperaba, con los dos valores literales al lado, en inglés. `task` lleva el
identificador de la tarea afectada, o `null` cuando el hallazgo es del tablero y no de ninguna tarea en
concreto, el mismo criterio que ya usan `extra_root_unreadable` e `ignore_file_mismatch`.

| Comprobación | `code` | `task` | Mensaje |
|---|---|---|---|
| Identificadores duplicados | `duplicate_id` | el id compartido | `id "MYP-40" is used by 2 tasks, ids must be unique` |
| Tareas que no se pueden leer | `task_unreadable` | el id, si se puede recuperar, o `null` | `task "MYP-40" could not be parsed: unexpected end of JSON input` |
| Claves de extensión no declaradas | `undeclared_extension_key` | la tarea | `ext key "trello.card" is not declared, declared keys are "jira.issue"` |
| Estados, tipos o prioridades ya no configurados | `value_not_configured` | la tarea | `status "Blocked" is not one of the configured statuses "To Do, In Progress, Done"` (o el mismo mensaje con `type` o `priority`, según cuál sea) |
| `initial_status`, `active_status` o `terminal_status` fuera de `statuses` | `status_role_unknown` | `null` | `active_status "Doing" is not one of the configured statuses "To Do, In Progress, Done"` |
| `statuses` con menos de tres elementos, o dos papeles apuntando al mismo estado | `status_role_invalid` | `null` | `statuses has 2 elements, at least 3 are required`, o `active_status and terminal_status are both "Done", the three roles must be distinct` |
| Dependencias que apuntan a tareas inexistentes | `dependency_not_found` | la tarea que declara la dependencia | `dependency MYP-99 does not exist` |
| Ciclos de dependencias | `dependency_cycle` | una tarea del ciclo | `MYP-11 is part of a dependency cycle: MYP-11 -> MYP-12 -> MYP-11` |
| Ciclos de tarea padre | `parent_cycle` | una tarea del ciclo | `MYP-11 is part of a parent cycle: MYP-11 -> MYP-12 -> MYP-11` |
| Claves de criterio repetidas dentro de una tarea | `duplicate_criterion_key` | la tarea | `acceptance criterion key #3 is used by 2 criteria, keys must be unique within a task` |
| Arrendamiento sin tarea activa y asignada | `lease_invariant` | la tarea | `has a lease but is not both active and assigned` |
| Identificador más alto por detrás | `highest_id_behind` | `null` | `the highest recorded id was MYP-40 and tasks go up to MYP-52` |
| Falta el marcador `<id>.id` | `marker_missing` | `null` | `board directory has no <id>.id marker, the database says id is "3f9a2b1c"` |
| El marcador `<id>.id` nombra un `id` distinto | `marker_id_mismatch` | `null` | `marker file names id "a1b2c3d4", the database says id is "3f9a2b1c"` |
| Raíz de `boards_extra_roots` no legible | `extra_root_unreadable` | `null` | `extra board root "/Volumes/disco/boards" cannot be read (skipped when looking up boards by id)` |
| **Integridad de la base de datos falla** | **sin `code`** | | **Nunca entra en `problems`, `warnings` ni `fixed`: aborta el comando entero con su propio mensaje y el código 21, antes de que `doctor` construya ningún JSON** (sección ["El sondeo del sistema de ficheros"](#el-sondeo-del-sistema-de-ficheros) de más arriba explica el porqué con más detalle) |
| Sistema de ficheros inseguro para WAL | `unsafe_wal_filesystem` | `null` | `board directory "/Users/avilches/.biso/boards/my-project-3f9a2b1c" is on a filesystem where SQLite's WAL mode is not safe (the byte-range lock or the shared mmap probe failed)` |
| El fichero de exclusión no corresponde al `vcs` configurado | `ignore_file_mismatch` | `null` | `.gitignore does not match the configured vcs "none" (left over from git, biso does not rewrite it automatically)` |

### Las dos formas de un mensaje reparable

**Las comprobaciones que `--fix` repara dicen dos cosas distintas, y las dos son normativas**: una
bajo `Errors:`, mientras el problema sigue ahí, y otra en el grupo de lo reparado, cuando `--fix` ya lo
ha arreglado. La tabla de arriba fija la primera, que es la que se ve en un tablero sin tocar; esta fija
las dos, una al lado de la otra, para que no haya que deducir ninguna. El `code` es el mismo en los dos
sitios, y las dos formas tienen su prueba de fichero dorado.

| `code` | Bajo `Errors:` | En el grupo de lo reparado |
|---|---|---|
| `lease_invariant` | `has a lease but is not both active and assigned` | `MYP-52 had a lease but was not both active and assigned; cleared leaseExpiresAt and leaseHolder` |
| `highest_id_behind` | `the highest recorded id was MYP-40 and tasks go up to MYP-52` | `the highest recorded id was MYP-40 and tasks go up to MYP-52; recorded MYP-52` |
| `marker_missing` | `board directory has no <id>.id marker, the database says id is "3f9a2b1c"` | `wrote the <id>.id marker, the database says id is "3f9a2b1c"` |

La única de las tres que nombra la tarea en su propio texto es la del arrendamiento, y solo en la forma
reparada: bajo `Errors:` el identificador va en la columna de la izquierda, y el grupo de lo reparado no
tiene esa columna. Las otras dos son hallazgos del tablero, con `task` a `null`, así que no nombran
ninguna. Las cuatro líneas del contador y del marcador, tal como salen en el informe:

```
  the highest recorded id was MYP-40 and tasks go up to MYP-52
```

```
  the highest recorded id was MYP-40 and tasks go up to MYP-52; recorded MYP-52
```

```
  board directory has no <id>.id marker, the database says id is "3f9a2b1c"
```

```
  wrote the <id>.id marker, the database says id is "3f9a2b1c"
```

### Cuando el tablero no recuerda haber asignado ningún identificador

El contador del identificador más alto vale cero en un tablero que nunca ha asignado ninguno, y cero no
es un identificador: no existe ninguna tarea `MYP-0` y el mensaje no se la inventa. **Cuando el contador
está a cero, el mensaje dice que no hay ninguno registrado** en vez de nombrar uno que nadie podría
buscar:

```
  no id is recorded as handed out and tasks go up to MYP-52
```

La forma reparada es la misma regla de la subsección anterior, la frase de arriba seguida de lo que se
registró: `no id is recorded as handed out and tasks go up to MYP-52; recorded MYP-52`. El `code` sigue
siendo `highest_id_behind`, porque el hallazgo es el mismo y quien ramifica sobre el `code` no tiene que
distinguir dos casos que se arreglan igual. Un tablero sin ninguna tarea tampoco dispara nada, porque el
contador a cero solo está por detrás cuando existe alguna tarea con un número por delante de él.

### Cuando la lista que el mensaje cita está vacía

Dos de los mensajes citan una lista configurada: la de claves de extensión declaradas y la del
vocabulario cerrado que un valor incumple. **`types`, `priorities` y `extensions` se pueden dejar
vacías** (["`biso config`"](config.md)), y entonces citar la lista imprimiría unas comillas con nada
dentro, que no dice nada y se lee como un fallo del programa. **Con la lista vacía el mensaje dice que
no hay ninguna**, en vez de citarla:

```
  ext key "trello.card" is not declared, and the board declares none
```

```
  type "bug" is not one of the configured types, and the board configures none
```

Los `code` no cambian, `undeclared_extension_key` y `value_not_configured`: es el mismo hallazgo
contado con las palabras que le tocan. `statuses` no necesita esta forma, porque nunca puede quedarse
vacía: la comprobación de la tabla de arriba exige al menos tres.

### El orden en que sale el informe

**Los hallazgos salen en el orden de la tabla ["Qué comprueba"](#qué-comprueba)**, de arriba abajo, y
los avisos siguen esa misma tabla en su propia lista. Dentro de una misma comprobación el orden es el
del tablero, es decir, identificador ascendente para todo lo que se pregunta de una tarea, y el orden
en que están declaradas para lo que se pregunta de la configuración de la máquina.

Es un orden fijo y no una consecuencia de cómo se implemente: varias de las comprobaciones que miran
una tarea se hacen en una sola pasada por el tablero, que es lo barato, así que sin esta regla el
informe saldría agrupado por tarea en un sitio y por comprobación en otro. Con ella, dos ejecuciones
sobre el mismo tablero imprimen exactamente lo mismo, y quien lee el informe puede encontrar una
comprobación por su posición en la tabla.

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
código de no poder escribir (8, el mismo de cualquier otro fallo de entorno al reparar), y la falta del
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
| La base de datos no se puede leer (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | El comando entero aborta con el mensaje y el código 21 de ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), antes de comprobar nada más |
| `--fix` sin poder escribir | Código 8. Si falla la escritura del marcador después de la transacción de datos, esta ya quedó aplicada (ver arriba) |
| `--fix --dry-run` | Vista previa pura de `--fix`: el mismo informe, con `Errors:` y `Warnings:` idénticos a los de una llamada real y el bloque de reparaciones renombrado a `N error(s) would be fixed, nothing was written (--dry-run)`, sin escribir nada. El código sigue la tabla de códigos de `doctor` de más abajo aplicada a ese mismo informe: 6 si queda algún error sin reparar tras la vista previa del `--fix`, 0 si no queda ninguno |

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

## El informe en seco de `--fix --dry-run`

`--fix --dry-run` no es una excepción de formato: reusa el mismo informe de `--fix` entero, con
`Errors:` y `Warnings:` idénticos, porque son hallazgos y ningún hallazgo cambia por no escribir nada.
Lo único que cambia es el último bloque, que en una llamada real dice `N error(s) fixed` y en la vista
previa dice `N error(s) would be fixed, nothing was written (--dry-run)`, con el mismo detalle línea a
línea de lo que se habría reparado.

Con el mismo tablero del ejemplo de más arriba, donde el único error reparable es el contador de
identificador más alto y la dependencia rota de `MYP-40` no lo es:

```
$ biso doctor --fix --dry-run
2 errors found, 1 warning found
Errors:
  MYP-40  dependency MYP-99 does not exist
Warnings:
  extra board root "/Volumes/disco/boards" cannot be read (skipped when looking up boards by id)
1 error would be fixed, nothing was written (--dry-run)
  the highest recorded id was MYP-40 and tasks go up to MYP-52; recorded MYP-52
```

**El código de salida de esta llamada es 6**, no 0, porque `MYP-40` sigue bajo `Errors:` después de la
vista previa: es un error que hoy no es reparable con `--fix` (tabla ["Qué comprueba"](#qué-comprueba)), así que
seguiría estando ahí después de la llamada real que le siga, y el `--dry-run` tiene que decirlo con el
mismo código que diría esa llamada, no con un 0 que sugeriría que todo va a quedar bien. Solo cuando
`Errors:` queda vacío después de la vista previa del `--fix`, la llamada sale con 0.

**El `--json` de esta llamada usa el mismo `kind: "doctor"` y el mismo esquema** que una llamada real,
con `fixed` llevando lo que se repararía en vez de lo que se reparó; no hay ninguna clave que distinga
un informe en seco de uno real, porque las tres listas (`problems`, `warnings`, `fixed`) ya dicen todo
lo que hace falta y el propio `--dry-run` de la llamada ya lo dice quien la hizo.

Esto no es una excepción nueva al contrato genérico de `--dry-run` de la sección ["Flags globales"](flags-globales.md#flags-globales):
ese contrato ya dice 0 si habría funcionado y 7 si no para el resto de comandos, pero `doctor` nunca
tuvo código 7 (tabla de códigos de más abajo), así que su `--dry-run` nunca pudo devolverlo. `--fix
--dry-run` es sencillamente el mismo caso llevado a sus últimas consecuencias: una vista previa pura del
código que daría la llamada real que le sigue, tomado de la propia tabla de códigos de `doctor`
(0, 6, 2, 8, 20 o 21), nunca un 7 que esta tabla no tiene.

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
| No se puede escribir al reparar | 8 |
| No hay tablero | 20 |
| Su base de datos no se puede leer (["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) | 21 |

**Esta misma tabla vale para `--fix --dry-run`**, sección ["El informe en seco de `--fix
--dry-run`"](#el-informe-en-seco-de---fix---dry-run): es una vista previa del código que daría la
llamada real que le sigue, nunca un 7 propio, porque `doctor` no tiene código 7.

## `biso doctor --help`

```
Usage: biso doctor [options]

Check the board for duplicate ids, unreadable tasks, undeclared extension keys,
values that are no longer configured, a broken status-role invariant, broken
dependencies, dependency cycles, parent cycles, repeated criterion keys, a lease
on a task that is not both active and assigned, a recorded highest id that has
fallen behind, a database that fails its integrity check, a missing or
mismatched <id>.id marker, an extra board root that cannot be read, an exclusion
file that no longer matches the configured vcs, and a board directory on a
filesystem where SQLite's WAL mode is not safe.

Options:
      --fix      repair what can be repaired without a decision
  -h, --help     show this help

Without --fix this is a read-only command: --print and --dry-run are bad usage
here, same as in any other read-only command. With --fix, --dry-run reports
what would be fixed without fixing it, same report as a real run, same exit
code too: 6 if an error would remain unfixed, 0 otherwise. Never its own 7,
doctor has none.

Findings come in two levels: errors, which leave the board inconsistent or
unreliable, and warnings, which are true and worth knowing but fix nothing.
Only remaining errors produce exit code 6.

An unreadable task is reported and skipped, never a reason to stop. A database
that cannot be opened, or that fails its integrity check, is not a finding: the
whole command fails instead, with exit code 21.
Gaps in the id sequence are normal and are not reported.

Exit codes:
  0  nothing wrong, or every error found was fixed
  2  bad usage
  6  errors remain
  8  cannot write while fixing
  20 no board here
  21 its database could not be read

Examples:
  biso doctor
  biso doctor --fix
```

---

