# El arrendamiento de una tarea

Esta página reúne las reglas de `leaseExpiresAt` y `leaseHolder`, los dos campos de ["El modelo de
datos de una tarea"](modelo-de-datos/index.md) que dicen si alguien está trabajando ahora mismo en una tarea
activa y hasta cuándo vale esa reserva. El nombre "arrendamiento", y por qué no es un estado, están en
["Vocabulario de esta especificación"](vocabulario.md).

## Cuándo cuenta como vencido

`leaseExpired` es el campo derivado que contesta la pregunta. Vale cierto cuando `leaseExpiresAt`
tiene valor y ese instante es anterior al reloj de quien lee, y **vale falso cuando `leaseExpiresAt`
está vacío**, que es el caso de toda tarea sin arrendamiento: no hay ningún estado en el que este
derivado se quede sin valor, porque un derivado que no se pudiera calcular es justo lo que el
principio 1 de ["Los principios"](principios.md) no admite.

**`leaseExpired` no cambia nunca el `status` guardado.** Dice que el arrendamiento de una tarea activa
venció, pero el estado guardado sigue siendo el activo hasta que alguien lo cambia con una escritura
explícita: lo que vence es la reclamación, no el estado (sección ["Saber si alguien está trabajando de
verdad"](../decisiones/modelo-de-estados.md#saber-si-alguien-está-trabajando-de-verdad)). No hay una
escritura diferida que la saque del estado activo por su cuenta, porque eso haría que un comando
tocara tareas que no nombró, y porque `biso prime`, que no escribe nunca, mostraría un estado que una
escritura ajena y posterior podría cambiar. Liberar el arrendamiento vencido es la reclamación
explícita que hace [`biso start`](cmd/verbos-del-ciclo.md#biso-start), no un efecto secundario de
ningún otro comando.

## La renovación

**Renovar `leaseExpiresAt` y fijar o transferir `leaseHolder` son cosas distintas, y solo la segunda
pasa por `biso start` o por su atajo `biso new --start`.** Cualquier escritura sobre una tarea activa
y asignada renueva `leaseExpiresAt` a `ahora + lease_minutes` (clave de [configuración](cmd/config.md)),
pero solo cuando quien llama ya es `leaseHolder`. **Cualquier escritura son todas**, sin ninguna
excepción: los [seis verbos del ciclo](cmd/verbos-del-ciclo.md), [`biso set`](cmd/set.md) y
[`biso archive`](cmd/archive.md), que son los comandos que llegan a escribir sobre una tarea que ya
existe. Se nombran aquí porque una regla general que no nombra a nadie invita a buscarle excepciones
donde no las hay.

**Una escritura que no cambia ningún campo renueva igual**: [`biso set`](cmd/set.md) con todas sus
flags dando el valor que la tarea ya tiene sale con código 0 y con `note: MYP-11 unchanged`, y aun
así renueva `leaseExpiresAt`, porque sigue siendo una escritura del tenedor sobre su tarea y el latido
no puede depender de si los valores coincidían por casualidad. Esa renovación no toca `updatedAt`,
porque ningún campo de la tarea ha cambiado, y deja vacía la lista `changed` del [esquema
JSON](cmd/set.md); la nota sigue siendo cierta, porque habla de los campos de la tarea y ninguno
cambió.

Si la tarea no tiene arrendamiento todavía, escribir sobre ella no lo crea: fijarlo por primera vez es
parte de lo que hace `biso start`, igual que reclamarlo vencido o tomarlo de [otra
identidad](cmd/verbos-del-ciclo.md#biso-start). Una escritura de una identidad distinta de
`leaseHolder` mientras el arrendamiento está vivo no toca ninguno de los campos: avisa con el mismo
`warning: MYP-11's lease is held by @sara until 2026-09-08T14:00:00Z` de [`biso
start`](cmd/verbos-del-ciclo.md#biso-start) y de la tabla de la sección ["Notas y
avisos"](salida-y-terminal.md#notas-y-avisos), y el resto de la escritura se hace igual. **Con una
sola excepción, y es que esa misma escritura rompa la invariante de [El vaciado](#el-vaciado)**: si
deja la tarea fuera del estado activo, sin ninguna persona asignada o archivada, los campos se vacían
en esa misma escritura, sea quien sea quien la haga, y el aviso de que el arrendamiento era de otra
identidad se emite igual. Una escritura de una identidad distinta mientras el arrendamiento está
vencido tampoco lo toca, y lo deja vencido: quien comenta, anota o cierra una tarea no ha reclamado
nada.

**Reclamar es de [`biso start`](cmd/verbos-del-ciclo.md#biso-start) y de su atajo [`biso new
--start`](cmd/new.md), y de nadie más**, con una excepción que hay que nombrar porque sin ella la
frase sería falsa: `biso start -s <estado>` con un estado que no es el activo no fija arrendamiento,
ya que fijarlo ahí rompería la invariante de [El vaciado](#el-vaciado), y deja los campos como los
dejaría cualquier otra escritura. Una tarea que llega a activa y asignada por cualquier otra vía no
tiene arrendamiento hasta que alguien llame a `biso start` sobre ella, y esas vías son exactamente
dos: los flags de campo de la sección ["Las familias de flags"](familias-de-flags.md), por
ejemplo `biso set --status`, ninguna de las cuales lo puede crear, y [la importación](#la-importación).

## El vaciado

**Los campos solo tienen valor en una tarea activa y asignada, y se vacían al perder cualquiera de
las dos condiciones, no solo la primera.** Una escritura que saca la tarea del estado activo
(`biso finish`, o `biso set --status` a cualquier otro valor) vacía `leaseExpiresAt` y `leaseHolder`
en esa misma escritura. Y como la condición que los sostiene es la conjunción de las dos cosas, perder
la segunda los vacía igual: `--clear-assignees` o
[`--rm-assignees`](familias-de-flags.md#campos-de-lista-que-admiten-coma) sobre una tarea activa
que se queda sin ninguna persona asignada vacía los campos en esa misma escritura, sea quien sea quien
la haga.

**[`biso archive`](cmd/archive.md) los vacía también**, aunque `archived` no sea un estado y archivar
no saque la tarea del estado activo: archivar es dejar de trabajar en la tarea, y un arrendamiento es
la afirmación de que alguien está trabajando ahora, así que conservarlo lo guardaría donde nadie lo
ve, porque `biso prime` y `biso ls` excluyen las archivadas por defecto, y `--unarchive` la devolvería
al tablero semanas después a nombre de una sesión que ya murió.

**Y por el mismo motivo, [`biso start`](cmd/verbos-del-ciclo.md#biso-start) se niega a tomar el
arrendamiento de una tarea archivada.** Reintroducir la afirmación "alguien trabaja en esto ahora"
sobre una tarea que se sacó del tablero activo a propósito contradice de frente la razón de este
vaciado, así que `start` es error 6 ahí, con la pista de `biso archive --unarchive` primero
(["`biso start`"](cmd/verbos-del-ciclo.md#biso-start)). Ningún otro verbo del ciclo se comporta así:
`set`, `note`, `comment`, `finish`, `ask` y `answer` escriben sobre una tarea archivada igual que
sobre cualquier otra, porque ninguno de ellos reclama un arrendamiento ni afirma que alguien esté
trabajando ahora mismo.

**Esta regla gana siempre sobre la de [La renovación](#la-renovación), y por eso la invariante se
enuncia aquí y el aviso allí.** Cuando quien escribe no es `leaseHolder`, el aviso de que el
arrendamiento es de otra identidad se emite igual, pero los dos campos se vacían: `@sara` haciendo
`biso finish MYP-11` sobre una tarea arrendada por `@claude` la deja terminada y sin arrendamiento.
Con la precedencia al revés quedaría una tarea terminada con un arrendamiento vivo, que es exactamente
lo que [La importación](#la-importación) rechaza al importar, así que `biso export` produciría un
fichero que su propio `biso init --from` rechaza y la prueba de simetría de ["El contrato de
estabilidad"](estabilidad.md) fallaría (sección ["Saber si alguien está trabajando de
verdad"](../decisiones/modelo-de-estados.md#saber-si-alguien-está-trabajando-de-verdad)). **Y los
campos van siempre juntos**: ninguna escritura, y tampoco la importación, deja uno con valor y el otro
vacío.

## La importación

**La importación los escribe con el valor que traiga el fichero, y es la única vía que lo hace.** Los
dos son campos guardados y no derivados, así que [`biso export`](cmd/export.md) los escribe y [`biso
new --from`](cmd/new.md) los lee de vuelta como cualquier otro, que es lo que hace cierta la garantía
de simetría de [`biso export`](cmd/export.md) sin una lista de excepciones que mantener.

La invariante de [El vaciado](#el-vaciado) se comprueba al importar, y en sus dos mitades. Una línea
que traiga `leaseExpiresAt` o `leaseHolder` sobre una tarea que no esté a la vez en el estado activo y
asignada a alguien es un fallo de validación del [lote](cmd/new.md), igual que una clave desconocida.
Y una línea que traiga uno de los campos y no el otro es el mismo fallo, con el mismo trato: los dos
vienen juntos o no viene ninguno, porque un `leaseHolder` sin `leaseExpiresAt` sería un arrendamiento
que no caduca nunca, y un `leaseExpiresAt` sin `leaseHolder` una reserva de nadie.

Un arrendamiento importado no privilegia a nadie: `leaseExpired` se recalcula contra el reloj de la
máquina que lee, así que el que llegue caducado sale caducado y [`biso
start`](cmd/verbos-del-ciclo.md#biso-start) lo reclama, y el que llegue vivo a nombre de otra
identidad solo produce el aviso de la sección ["Notas y avisos"](salida-y-terminal.md#notas-y-avisos)
hasta que caduque.
