# El modelo de estados

## El modelo de estados: cuatro requisitos, cerrados

Salen de un diseño anterior de gestor de tareas que no llegó a escribirse, y de la evidencia que
aquel diseño recogió sobre la herramienta que usaba esta máquina antes. Los cuatro tocaban el modelo
de estados y se decidieron a la vez: dos quedan resueltos sin ningún papel de estado nuevo, uno se
resuelve con la enmienda de la decisión de persistencia (["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad)), y uno se retira. El criterio que ordenó
las cuatro decisiones, y que conviene aplicar la próxima vez que alguien proponga un papel de estado,
está en ["El criterio de estado frente a campo"](#el-criterio-de-estado-frente-a-campo).

### Distinguir el encargo de la ejecución: resuelto sin estado nuevo

El requisito decía que, con los tres papeles de la especificación, el gesto de una persona que encarga
trabajo y el de un agente que lo coge **son el mismo dato**, y que hacía falta un cuarto papel de
estado para separarlos.

**El diagnóstico era falso, y la solución sale de corregirlo, no de añadir nada.** La persona ya
escribe un dato: `assignees`. El agente escribe otro: el estado activo, al ejecutar `biso start`. Son
dos datos distintos, escritos por dos actores distintos, y esa asimetría existía ya en la
especificación antes de este trabajo; nadie la había mirado como la respuesta al requisito. Y
aplicando el criterio de ["El criterio de estado frente a campo"](#el-criterio-de-estado-frente-a-campo), tampoco podía ser nunca un estado: "esto lo tiene que hacer un
agente" convive con cualquier punto del camino, porque una tarea puede estar recién creada, a medias,
o aparcada en una pregunta abierta y seguir siendo de quien se la asignaron.

**Esta lectura solo es cierta sin la clave de configuración `default_assignee`.** Esa clave habría
asignado una persona a toda tarea creada sin `-a`. En un tablero que la usara, absolutamente todo
nacería asignado, y la asignación dejaría de significar que alguien decidió encargarte justo esa
tarea: la consulta de arranque de un agente devolvería el backlog entero disfrazado de encargo. Se
retira la clave entera: su fila en la tabla de configuración, su comportamiento en `biso new`, su
aparición en el esquema JSON de `config list` y su línea de `biso config --help`. La autoasignación de
`biso start` se queda, porque no ensucia la señal: cuando `start` asigna a `me`, la tarea entra a la
vez en el estado activo, así que nunca queda en el conjunto de "asignada y sin empezar", que es el que
importa.

**Lo que esta decisión cuesta, para que no se descubra por sorpresa.** El caso de uso que motivó el
requisito venía de un proyecto en el que Backlog.md se sincronizaba con Trello, y Trello hacía de
interfaz en el móvil: ahí se arrastraba la tarjeta de una columna a otra y el cambio llegaba al
tablero. `biso` no tiene arrastre ni columnas que mover, así que ese gesto no existe aquí. Con la
asignación como encargo, el gesto equivalente es asignar la tarea a un agente, que sigue siendo viable
desde el móvil pero es un gesto distinto: nombrar a alguien, no desplazar nada. Cuando se especifique
la sincronización con un sistema externo como Trello, mover una tarjeta de columna en ese sistema no
significará nada para `biso` por sí mismo, y esa correspondencia habrá que definirla entonces a
propósito, no por sorpresa. Y no hay forma de decir "esto es tuyo, pero todavía no": con la asignación
como única señal, asignar autoriza a empezar de inmediato, y quien necesite esa espera tiene que no
asignar hasta que toque, o usar una fecha límite.

### Saber si alguien está trabajando de verdad

Una tarea que un agente coge antes de que su sesión muera se queda en el estado activo
indefinidamente, y nada lo detecta.

La evidencia es concreta. La herramienta que se usaba antes tenía un campo para apuntar quién está
trabajando en una tarea, y el tablero web lo pintaba, pero era una lista a la que se añade un nombre
al empezar y que se vacía al terminar: **sin latido, sin marca de tiempo y sin caducidad**, de modo
que una sesión que muere deja su nombre ahí para siempre. Y había un segundo fallo encima: solo lo
escribían su servidor de integración y su API, nunca su línea de comandos, aunque su propio mensaje
de arranque pedía al agente que mandara su nombre al cambiar de estado.

Hay además un dato medido que descarta la solución fácil: **el evento de fin de sesión de Claude Code
trae un motivo con cinco valores posibles y ninguno corresponde a un proceso muerto a lo bruto**, así
que un hook de cierre no puede ser la única señal de que una tarea quedó huérfana, porque el caso que
importa es justamente el que no lo dispara.

La forma conocida de resolverlo es un arrendamiento con caducidad: latido mientras se trabaja, y algo
que libere la tarea cuya sesión murió. Arrendar sería asignar con esa caducidad sobre una tarea ya
activa, y reclamarla sería la misma escritura que hoy hace `biso start`, así que la decisión de este
apartado sigue encajando con la de ["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo) aunque aquel, al final, no trajera ningún papel nuevo.

**Esta dependía de cómo se persistan los datos**, y se cerró con esa decisión tomada.

**Cerrada, con una enmienda a lo que este apartado decía antes.** La redacción aplazada decía que la
caducidad, al vencer, "saca la tarea del estado activo sin tocar la asignación". Esa frase no se
sostiene y es la que cede: `status` es un campo guardado, con un valor del vocabulario del tablero, y
un campo derivado no puede cambiar un campo guardado. Si nada escribe, la tarea sigue teniendo el
estado activo guardado por muy vencido que esté su arrendamiento.

Lo que se guarda es el instante de caducidad (`leaseExpiresAt`) y quién tiene el arrendamiento
(`leaseHolder`), los dos sin valor salvo en una tarea activa y asignada (["El vaciado"](../spec/lease.md#el-vaciado) de `lease.md`). Lo
que caduca sigue siendo, como decía la redacción original, "estoy en ello" y no "esto es tuyo": el
estado guardado no cambia solo, nunca (["Cuándo cuenta como vencido"](../spec/lease.md#cuándo-cuenta-como-vencido) de `lease.md`). Liberarlo es una escritura explícita, y sigue sin hacer falta un
comando nuevo para eso: es la misma que ya hace [`biso start`](../spec/cmd/verbos-del-ciclo.md#biso-start), con la comprobación de
tenedor que se explica más abajo.

**Por qué se enmienda en vez de reescribirse sin más.** La redacción aplazada no decía cómo se
liberaba una caducidad, y la lectura más directa de "saca la tarea del estado activo" es una
escritura diferida: que la siguiente escritura cualquiera, o un proceso de fondo, arrastrara el
saneamiento de las tareas vencidas. Esa vía se descarta explícitamente al tomar esta decisión, en
["Una tarea se queda cogida porque la sesión murió"](../estado-del-arte/catalogo-de-problemas.md#6-una-tarea-se-queda-cogida-porque-la-sesión-murió), porque haría que un comando tocara tareas que no nombró, y
porque `biso prime`, que no escribe nunca, mostraría un estado que una escritura ajena y posterior
podría cambiar. No es que la decisión siempre hubiera sido la de hoy: es que la única forma de
sostenerla, al escribirla de verdad, obligaba a mover la contradicción con `status` a otro sitio en
vez de resolverla.

**Lo que aporta el estado del arte, mirado al cerrar esta decisión.** El patrón tiene nombre propio
fuera de aquí: un arrendamiento con caducidad, renovado por latido, para evitar la doble reclamación.
La pieza que le faltaba a la redacción aplazada es la comprobación del tenedor, y la idea viene de un
artículo que formaliza el patrón: propone el límite de tiempo **más un token de vallado que rechace las
escrituras del propietario antiguo** (["Una tarea se queda cogida porque la sesión murió"](../estado-del-arte/catalogo-de-problemas.md#6-una-tarea-se-queda-cogida-porque-la-sesión-murió), con su enlace). Sin nada
de eso, el tenedor viejo puede despertar, escribir, y robar de vuelta una tarea que ya había reclamado
otro. Comprobar quién lo tiene sale gratis aquí, porque el arrendamiento ya guarda ese dato y la
comprobación es comparar y sustituir dentro de una transacción que ya existía por otro motivo
(["Concurrencia, atomicidad y garantías observables"](../spec/garantias.md#concurrencia-atomicidad-y-garantías-observables)).

**Pero lo que se implementa no es ese token de vallado, y apartarse de él es deliberado.** Un token de
vallado rechaza la escritura entera de quien ya no es el propietario. `biso` la acepta: la escritura de
una identidad que no es `leaseHolder` se hace igual y solo deja intactos los campos del
arrendamiento, con el aviso de ["Notas y avisos"](../spec/salida-y-terminal.md#notas-y-avisos) (["La renovación"](../spec/lease.md#la-renovación) de `lease.md`). Así que el
agujero que el artículo cierra aquí queda entreabierto: el tenedor viejo que despierta puede comentar,
anotar o cerrar la tarea que otro reclamó, y si vuelve a llamar a `biso start` se la lleva de vuelta,
con aviso y sin impedimento. Se acepta porque rechazar la escritura sería lo único de todo el programa
que impide trabajar por el estado en que está una tarea: las dependencias sin terminar avisan, la
pregunta abierta avisa, y un arrendamiento ajeno avisa igual, por el motivo de ["El porqué de reglas concretas"](comandos-y-flags.md#el-porqué-de-reglas-concretas), que un
bloqueo de flujo no evita el trabajo duplicado y sí empuja a rodear la herramienta. Lo que la
comprobación del tenedor sí cierra, y es lo que se gana, son las dos cosas que el aviso no puede dar:
que una escritura de otra identidad nunca renueve el plazo ajeno ni se atribuya el arrendamiento, y que
de dos reclamaciones simultáneas de un arrendamiento vencido solo gane una, porque `biso start`
comprueba dentro de su propia transacción que seguía vencido. La diferencia con el artículo queda
anotada como riesgo aceptado en ["Riesgos conocidos y aceptados del modelo de estados"](#riesgos-conocidos-y-aceptados-del-modelo-de-estados).

**Por qué el arrendamiento se exporta e importa como cualquier otro campo.** Al añadir los campos
guardados quedó sin decir si viajan en `biso export`, y las dos respuestas eran defendibles: dejarlos
fuera, porque un arrendamiento es la reserva de una sesión concreta en una máquina concreta, o dejarlos
entrar, porque son campos guardados y no derivados y la garantía de simetría de ["`biso export`"](../spec/cmd/export.md)
promete que todos ellos van y vuelven. Se eligió la segunda, y en rigor no era una elección
libre: el ["contrato de estabilidad"](../spec/estabilidad.md) ya enuncia esa simetría como una prueba
de la suite "sobre todos los campos no derivados", sin lista de excepciones, así que dejar fuera el
arrendamiento habría obligado a abrir una y a mantenerla, que es exactamente la clase de enumeración
que se desincroniza (["Una advertencia sobre cómo se mantiene la especificación"](principios-y-mantenimiento.md#una-advertencia-sobre-cómo-se-mantiene-la-especificación)). La otra razón, la que hace que la primera no salga cara, es que el
diseño ya toleraba un arrendamiento ajeno: `leaseExpired` se recalcula contra el reloj de
quien lee, así que el arrendamiento que llega caducado sale caducado y `biso start` lo reclama, y el
que llega vivo a nombre de otra identidad produce el aviso de ["Notas y avisos"](../spec/salida-y-terminal.md#notas-y-avisos) y nada más, porque `biso
start` avisa y coge la tarea igual. El caso que de verdad importa, restaurar un respaldo del propio
tablero, sale además mejor así: la tarea que estaba en marcha sigue constando en marcha y a nombre de
quien la llevaba, en vez de aparecer activa y sin dueño del arrendamiento. Lo único que hay que
custodiar es la invariante de que los campos solo tienen valor en una tarea activa y asignada, y se
custodia en dos sitios: en la validación previa de `biso new --from`, donde se custodia todo lo demás
del lote y que rechaza el fichero entero antes de escribir nada, y en `biso doctor`, que la comprueba
como comprueba las demás invariantes del tablero y la repara con `--fix` vaciando los campos
(["`biso doctor`"](../spec/cmd/doctor.md)). El segundo hace falta porque la importación no es la única forma de que
una base de datos llegue a incumplirla: puede venir escrita a mano, restaurada a medias o de otra
versión, y sin esa comprobación la única invariante que este apartado dice que hay que custodiar sería
la única que `doctor` no mira.

**La invariante gana siempre sobre el aviso de un arrendamiento ajeno, y esa precedencia hay que
escribirla.** Estas reglas se enfrentan en un caso corriente: una escritura de una identidad que no es
`leaseHolder` sobre una tarea con arrendamiento vivo, cuando esa misma escritura además saca la tarea
del estado activo o la deja sin nadie asignado. El argumento no es de gusto: la regla del aviso es una
cortesía hacia el tenedor y la del vaciado es la invariante de la que dependen la validación del lote y
`doctor`, así que con la precedencia al revés quedaría una tarea terminada con un arrendamiento vivo,
rompiendo la prueba de simetría del ["contrato de estabilidad"](../spec/estabilidad.md) (["El vaciado"](../spec/lease.md#el-vaciado) de `lease.md`,
con el ejemplo completo). Por el mismo argumento, `biso archive` vacía también los campos, aunque
`archived` no sea un estado y archivar no saque la tarea del estado activo (["El vaciado"](../spec/lease.md#el-vaciado) de `lease.md`).

**Una escritura que no cambia ningún campo también late, y esto es lo que un implementador desharía
creyendo que optimiza.** `biso set` con todos sus flags dando el valor que la tarea ya tiene renueva
`leaseExpiresAt` igual si quien llama es el tenedor (["La renovación"](../spec/lease.md#la-renovación) de `lease.md`). Salta a la vista el atajo
contrario, no escribir nada cuando no hay nada que escribir, y es un error: el latido de este
arrendamiento no es un comando propio, es cualquier escritura que el agente ya hace, y si la renovación
dependiera de que algún valor hubiera cambiado de verdad, un agente que repite una escritura idempotente
creería estar latiendo sin latir, y perdería la tarea al vencer el plazo por haber hecho justo lo que la
especificación le dice que basta.

**Y `biso new --start` reclama el arrendamiento, porque es el atajo de dos llamadas y la equivalencia
tiene que ser real.** ["El porqué de reglas concretas"](comandos-y-flags.md#el-porqué-de-reglas-concretas) justifica ese flag como el ahorro de `biso new` más `biso start`;
si creara la tarea activa y asignada pero sin arrendamiento, las dos vías darían dos tareas distintas y
la única forma de saberlo sería leer la letra pequeña. La excepción simétrica es `biso start -s` con un
estado que no es el activo: ahí no se fija arrendamiento, porque fijarlo rompería la invariante de
arriba, y hay que nombrarla explícitamente para que la frase "solo `start` reclama" no se lea como una
regla sin excepciones.

**Dos remates que la primera redacción dejó a medias.** El primero: los campos ya se podían pedir
con `biso get --json`, y el mensaje de arranque remitía a `biso get` para verlos, pero su ficha de texto
no tenía dónde enseñarlos, así que la remisión era falsa para quien no pide JSON. Ahora la ficha imprime
una línea `lease` con los dos, y solo cuando la tarea tiene arrendamiento, porque una fila con dos
guiones aparecería en la ficha de casi todas las tareas del tablero (["`biso get`"](../spec/cmd/get.md)). El
segundo: la validación rechazaba los campos sobre una tarea que no estuviera activa y asignada, pero
no rechazaba que llegara uno solo de los dos, y con `leaseExpiresAt` vacío el derivado `leaseExpired`
comparaba un instante que no existe contra el reloj. Se cierra exigiendo que los dos campos vayan
siempre juntos (["La importación"](../spec/lease.md#la-importación) de `lease.md`): un derivado tiene que valer algo en todos los
tableros posibles, no solo en los que se importaron bien.

### Señalar lo que espera a una persona

Se puede configurar un estado tipo `Blocked`, pero es un estado más: `biso prime` no lo distingue, así
que una tarea parada esperando una decisión humana no se ve donde se mira.

La evidencia: en el tablero que se estudió había cuatro tareas paradas por una pregunta sin responder
y nada lo señalaba. Eran, además, preguntas abiertas disfrazadas de trabajo pendiente, porque una
decisión no tenía dónde vivir y la única forma de registrarla era crear una tarea.

Lo que falta es un papel que marque un estado como "espera a una persona", y que el mensaje de
arranque lo destaque en su propio bloque.

**Queda resuelta, y no con el papel que este apartado imaginaba.** Aplicando el criterio de
["El criterio de estado frente a campo"](#el-criterio-de-estado-frente-a-campo), una pregunta abierta puede detener una tarea en cualquier punto del camino, así que no podía ser
un estado: es el campo `question` de ["La pregunta abierta"](../spec/modelo-de-datos/pregunta-abierta.md#la-pregunta-abierta), con su derivado `waiting`, los
verbos `biso ask` y `biso answer`, y el bloque `NEEDS ANSWER` del mensaje de arranque, que es
exactamente el bloque propio que este apartado pedía.

### Distinguir terminar de descartar: retirado

Se retira, y conviene decir por qué el argumento que parecía bueno no lo era, para no repetir el error
si alguien vuelve a proponerlo.

**El argumento que no vale es que archivar y descartar son lo mismo.** No se sostiene contra lo que ya
dice `biso archive`: existe `--unarchive`, que devuelve la tarea al tablero con el estado que tenía, y
la ayuda del comando presenta el archivo como sacar tareas del tablero sin perderlas. Archivar es
reversible, así que archivar es aparcar, y aparcar no es descartar.

**Y la deducción que se apoyaba en él tampoco vale.** La idea era que una tarea archivada que nunca
llegó al estado terminal se pudiera etiquetar como abandonada. No se puede calcular, porque el modelo
guarda el estado actual de una tarea y no un histórico de sus estados anteriores. Con `biso start
--reopen` una tarea terminada vuelve al estado activo, y si se archivara desde ahí, esa deducción la
llamaría abandonada habiendo estado hecha.

**El argumento que sí vale es que a este requisito le falta la evidencia que los otros tres sí traen.**
["Distinguir el encargo de la ejecución: resuelto sin estado nuevo"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo) corrige un diagnóstico sobre una asimetría real de la especificación,
["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad) trae el fallo
medido de una herramienta que se usaba antes (un campo sin latido, sin marca de tiempo y sin
caducidad), y ["Señalar lo que espera a una persona"](#señalar-lo-que-espera-a-una-persona) trae cuatro tareas paradas por una pregunta sin responder en el tablero que se
estudió. Este requisito dice que una tarea hecha y una abandonada "se confunden", sin un solo caso en
el que esa confusión haya costado algo. Un papel de estado que obligue a dar un motivo es barato de
añadir cuando haga falta y caro de quitar si sobra, así que se queda fuera hasta que aparezca un caso
real que lo pida, y entonces se engancha a `biso archive`. Se anota en ["Lo que se deja fuera a propósito"](../spec/fuera-de-alcance.md).

### Lo que se miró de ese diseño anterior y se descarta

**Un modelo con entidades separadas para tarea, idea, aprendizaje y decisión**, cada una con sus
propios estados y campos. La necesidad que lo motivaba es real, porque sin sitio donde ponerlas las
decisiones se disfrazan de tareas. Pero hay un dato posterior que contradice la solución: en el
estudio de uso de una herramienta que sí tiene comandos dedicados para documentos y decisiones,
**no hubo una sola llamada a ninguno de ellos en seis días, sesenta sesiones y diez proyectos**. La
conclusión es que la necesidad se cubre mejor con un tipo más en el vocabulario que ya existe que con
una entidad y un comando propios.

**Una forma concreta de guardar los datos**, con un directorio por entidad para que añadir un
comentario no reescriba nada de lo demás. La [decisión de persistencia](persistencia.md#la-decisión-de-persistencia) no la adopta: una
base de datos SQLite da la misma propiedad, una escritura por tarea sin reescribir el tablero entero, y
además la transacción que un directorio de ficheros habría tenido que construir a mano.

### Un contraste que conviene mirar antes de implementar

Aquel diseño se imponía una regla contraria a la de esta especificación: **los valores por defecto
viven completos en el binario, el comando de creación del tablero no escribe ninguna configuración, y
un repositorio normal no tiene fichero de configuración nunca**. Su argumento era evitar acabar con un
motor genérico que no sabe hacer nada solo y que obliga a configurar antes de escribir la primera
tarea.

Aquí `biso init` sí escribe la configuración. El argumento contrario es bueno y merece mirarse antes
de dar la decisión por hecha.

---

## El criterio de estado frente a campo

Es la regla que ordenó las cuatro decisiones de ["El modelo de estados: cuatro requisitos, cerrados"](#el-modelo-de-estados-cuatro-requisitos-cerrados), y conviene tenerla escrita aparte porque
se va a volver a necesitar la próxima vez que alguien proponga un papel de estado:

> Algo es un estado cuando es excluyente con los demás y dice en qué punto del camino está la tarea.
> Es un campo cuando puede convivir con cualquier punto del camino.

La especificación ya la aplicaba sin decirla: `archived` es un campo y no un estado precisamente
porque una tarea archivada conserva el estado que tenía al archivarse. Lo que faltaba era el criterio
escrito, para no volver a meter en el vocabulario de estados algo que no es un punto del camino.
Aplicado a los cuatro requisitos de ["El modelo de estados: cuatro requisitos, cerrados"](#el-modelo-de-estados-cuatro-requisitos-cerrados), deja solo uno pidiendo de verdad algo excluyente y
ligado al camino, el de ["Distinguir terminar de descartar: retirado"](#distinguir-terminar-de-descartar-retirado), y ese es justo el que se retira por falta de evidencia; los otros tres son un
gesto que ya existía (["Distinguir el encargo de la ejecución"](#distinguir-el-encargo-de-la-ejecución-resuelto-sin-estado-nuevo)), una reserva con caducidad (["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad)) o un campo que puede convivir con cualquier
estado (["Señalar lo que espera a una persona"](#señalar-lo-que-espera-a-una-persona)).

**Por qué esta condición no se hace configurable.** Dejar que cada tablero declarase sus propios
estados como excluyentes o no destruiría la garantía en la que descansa el resto del modelo: que
`biso start`, `biso finish` y los filtros por papel (`--active`, `--not-active`) puedan asumir siempre
que una tarea está en exactamente un estado a la vez. Si esa garantía dependiera de la configuración de
cada tablero, todo comando tendría que consultarla antes de decidir qué significa "activa", y eso es
justo la clase de comportamiento que depende de dónde y con qué se ejecuta el programa que la
introducción de la especificación descarta. La condición es del modelo, no de un tablero concreto.

---

## Riesgos conocidos y aceptados del modelo de estados

Se aceptan a propósito, y conviene anotar por qué en cada uno para no tropezar dos veces con lo mismo.

- **Dos agentes con la misma identidad ven la misma cola.** Dos sesiones con el mismo `BISO_ME` no se
  distinguen entre sí, y las dos podrían coger la misma tarea a la vez. Es exactamente lo que resuelve
  el arrendamiento de ["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad): la segunda sesión ve el aviso de que el arrendamiento está vivo y
  a nombre de esa misma identidad, pero `biso start` avisa y la coge igual, así que dos sesiones que
  comparten identidad siguen pudiendo pisarse. El arrendamiento defiende de una sesión muerta, no de
  dos sesiones vivas con el mismo nombre.
- **El tenedor viejo de un arrendamiento no tiene prohibido escribir.** El artículo que formaliza el
  patrón propone, junto a la caducidad, un token de vallado que rechace las escrituras del propietario
  antiguo (["Una tarea se queda cogida porque la sesión murió"](../estado-del-arte/catalogo-de-problemas.md#6-una-tarea-se-queda-cogida-porque-la-sesión-murió)), y `biso` hace a propósito algo más débil: la
  escritura de una identidad que no es `leaseHolder` se acepta entera y solo deja intactos
  `leaseExpiresAt` y `leaseHolder`, con el aviso de ["Notas y avisos"](../spec/salida-y-terminal.md#notas-y-avisos). Un agente que despierta
  después de que otro reclamara su tarea puede comentarla, anotarla o cerrarla, y con `biso start`
  llevársela de vuelta. Se acepta por coherencia con "avisa, no impide", el mismo argumento de
  ["El porqué de reglas concretas"](comandos-y-flags.md#el-porqué-de-reglas-concretas) que vale para las dependencias sin terminar y para la pregunta abierta: un bloqueo de flujo no evita
  el trabajo duplicado, solo empuja a rodear la herramienta. Lo que el arrendamiento sí garantiza es que
  una tarea no se quede cogida para siempre, que nadie renueve ni se atribuya un arrendamiento ajeno, y
  que de dos reclamaciones simultáneas del mismo arrendamiento vencido solo gane una (["Saber si alguien está trabajando de verdad"](#saber-si-alguien-está-trabajando-de-verdad)).
- **Esto ya no es cierto.** Este riesgo describía que la clave `me` ganaba sobre `BISO_ME`, así que un
  tablero con `me` puesta hacía que todo el mundo compartiera identidad y `--mine` dejara de
  significar nada. La decisión ["La identidad de quien llama"](detalles.md#la-identidad-de-quien-llama) lo cierra: `me` ya no vive en el
  tablero, sino en la configuración de la máquina, y es `BISO_ME` quien gana. Un agente que fija su
  propio `BISO_ME` no hereda nunca la identidad que la máquina tenga configurada.
- **La persona no tiene un canal hacia el agente que se vea en el mensaje de arranque.** El agente
  pregunta con `biso ask` y la persona responde con `biso answer`, pero si la persona quiere decirle
  algo por iniciativa propia lo escribe en un comentario, y el mensaje de arranque no muestra
  comentarios. El agente lo ve al hacer `biso get`.
- **El listado no enseña el texto de la pregunta, solo dice qué tareas la tienen.** Es el precio de no
  meter texto largo en el listado, medido en el principio 4 de ["La evidencia detrás de los siete principios"](principios-y-mantenimiento.md#la-evidencia-detrás-de-los-siete-principios): 215 fichas completas
  sumaron 179.369 bytes, casi la cuarta parte de la salida del estudio. Quien quiera leer la pregunta
  usa `biso get --section question`, o mira el mensaje de arranque, que sí la enseña.
- **Una pregunta abierta sobre una tarea terminada o archivada desaparece de la vista.** `biso finish`
  avisa pero no impide, los bloques del mensaje de arranque excluyen terminadas y archivadas, y
  `biso ls` excluye el estado terminal por defecto, así que `biso ls --waiting` no la encuentra sin
  `--any-status`. Se acepta porque la alternativa, impedir cerrar una tarea con una pregunta abierta,
  empujaría a rodear la herramienta, el mismo argumento que ya vale en ["El porqué de reglas concretas"](comandos-y-flags.md#el-porqué-de-reglas-concretas) para `finish` y
  los criterios sin marcar.
- **El filtro de dependencias no excluye las tareas aparcadas.** `--not-blocked` mira solo
  dependencias, así que un agente que elija trabajo únicamente con ese flag se lleva también las
  que esperan una respuesta. La consulta correcta añade `--not-waiting`, y así lo dicen tanto la
  descripción del flag como el ejemplo de la ayuda de `biso ls`. El riesgo se queda, pero
  encogido: el nombre ya no promete que la tarea esté lista para trabajar, solo que no esté bloqueada,
  que es lo que mide. El porqué del nombre está en ["El porqué de reglas concretas"](comandos-y-flags.md#el-porqué-de-reglas-concretas).
