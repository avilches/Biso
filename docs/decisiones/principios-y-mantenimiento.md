# Los principios, y como se mantiene la especificacion

## La evidencia detrás de los siete principios

Los principios de ["Los principios"](../spec/principios.md) están ahí enunciados sin su procedencia,
porque un principio se aplica igual se sepa o no de dónde viene. Aquí está de dónde viene cada uno.

**Principio 1, que un valor desconocido es un error al leer y al escribir.** Es el fallo más peligroso
que se midió. En la herramienta estudiada, `task list --status "Todo"` contesta `No tasks found.` con
código de salida 0, mientras que `task edit -s "Todo"` escribe correctamente el mismo valor. Un agente
que escribe mal un filtro no recibe un error: recibe una afirmación falsa sobre el tablero, y "no hay
tareas" es exactamente la clase de respuesta sobre la que se construye la siguiente decisión, incluido
informar al usuario de que algo no existe.

**Principio 2, que un nombre significa lo mismo en todos los comandos.** En la herramienta estudiada,
`--ref` y `--acceptance-criteria` añaden en el comando de creación y reemplazan en el de edición. Cada
comando por separado es defendible; juntos son una trampa, porque quien escribe una secuencia los mira
juntos.

**Principio 3, que ningún flag depende de una regla que haya que conocer de antemano.** Se
midieron seis casos de agentes usando la variante destructiva de forma repetida creyendo que
añadían: el flag de plan aplicado hasta cinco veces sobre la misma tarea, el de referencias tres
veces sobre otra, y un caso en el que un agente ejecutó sobre una misma tarea `--ref`, `--ref`,
`--ref`, `--add-ref`, `--remove-ref` y `--clear-refs`, que es alguien probando a ver cuál de las seis
hace lo que quiere. Ninguna de esas llamadas dio error, y el daño es silencioso: cada una borró lo
que había escrito la anterior. La primera forma de este principio resolvía el problema con una regla
única, "el nombre desnudo añade", que había que conocer de antemano para no adivinar; se sustituyó
después por dar a cada operación su propio verbo explícito, sin ninguna regla que aprender, por el
motivo que cuenta ["La regla"](../spec/familias-de-flags.md#la-regla) de `familias-de-flags.md`.

**Principio 4, que la salida por defecto de una escritura es lo que quien llama no sabía.** De las 237
creaciones medidas, 215 llevaban un flag que devolvía la ficha entera de la tarea recién creada, y
sumaron 179.369 bytes, casi la cuarta parte de toda la salida del estudio, sin dar el único dato que
el agente no tenía, que es el identificador. La mediana de una creación pasa de 154 bytes sin ese
flag a 1.333 con ella, un factor de 8,7.

**Principio 5, que una operación del flujo de trabajo es un comando.** De las 159 ediciones medidas, 112
cambian exactamente un campo, hay 82 pares de ediciones consecutivas sobre la misma tarea, y el ciclo
de vida típico cuesta entre seis y doce llamadas. **Y no era una limitación de la herramienta**: se
comprobó que una sola llamada aceptaba el cierre entero y funcionaba. La fragmentación venía de que
sus guías presentaban el trabajo como una lista numerada con un comando por paso.

**Principio 6, el lote con validación previa.** Una migración real creó 242 tareas de una en una, y
cuando el entorno bloqueó los comandos compuestos que hacían falta, los agentes acabaron escribiendo
85 ficheros de tarea a mano, 76 ediciones y 9 creaciones, que es exactamente lo que la herramienta
prohíbe en la instrucción que ella misma inyecta en cada proyecto.

**Principio 7, que la salida no depende del terminal.** El flag `--plain` de la herramienta
estudiada tiene doble vida: en unos comandos apaga una interfaz interactiva y en otros enciende un
volcado completo. Medido fuera de un terminal, en los comandos de lectura no cambia un solo byte y
aparece 191 veces sin ningún efecto; en el de creación multiplica la salida por 3,3 y en el de edición
por 36.

---

## Una advertencia sobre cómo se mantiene la especificación

La especificación pasó por cuatro revisiones adversariales antes de darse por buena. El patrón de
fallo dominante, y con diferencia, fue siempre el mismo: **dos copias distantes de un mismo dato que
dejan de coincidir**. Una lista de campos que aparece en varias secciones, una cifra publicada en tres
sitios, un código de salida que está en la tabla de un comando pero no en su texto de ayuda.

De ahí salen estas costumbres que conviene mantener al editar:

1. **Cuando un dato tenga que aparecer en dos sitios, que uno remita al otro** en vez de repetirlo.
2. **Los ejemplos de salida se generan ejecutando el algoritmo, no se escriben a mano.** Los del
   listado y los del mensaje de arranque fallaron tres revisiones seguidas mientras se escribieron a
   mano, y dejaron de fallar en cuanto se generaron.
3. **Al corregir una afirmación, búscala en todo el documento** antes de darla por corregida.
4. **Un paso de implementación no se da por empezado sin decir qué anclas de la especificación va a
   cubrir, ni por terminado sin actualizar su fila en [Qué hay implementado y qué
   no](../spec/estado-de-implementacion.md).** Si el comportamiento real terminó siendo distinto del
   texto original, la especificación se corrige en ese mismo cambio: esa página nunca debe quedar
   más optimista que el código.

### El catálogo de datos que viven en más de un sitio

La costumbre 1 de arriba solo funciona si se recuerda cuáles son esos sitios. Esta tabla los junta,
para consultarla antes de editar cualquiera de los dos lados de una fila: la fuente es donde vive la
definición o la regla completa, y el otro sitio solo la usa, la ejemplifica o remite a ella.

| Dato | Fuente | También aparece en | De cuándo |
|---|---|---|---|
| Escalar vacío (`--author ""` y análogos) | [`valores-de-entrada.md`](../spec/valores-de-entrada.md#el-valor-vacío) | [`familias-de-flags.md`](../spec/familias-de-flags.md#campos-escalares), [`modelo-de-datos/autor.md`](../spec/modelo-de-datos/autor.md) | auditoría 2026-09-18 |
| `updatedAt` no cambia con el heartbeat del lease | [`modelo-de-datos/index.md`](../spec/modelo-de-datos/index.md#campos-automáticos) (la excepción) | [`lease.md`](../spec/lease.md), [`cmd/set.md`](../spec/cmd/set.md) (el razonamiento) | auditoría 2026-09-18 |
| Qué cuenta como "comando de lectura" | [`cmd/flags-globales.md`](../spec/cmd/flags-globales.md) | [`cmd/doctor.md`](../spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros) (su sondeo escribe sin dejar de ser de lectura) | auditoría 2026-09-18 |
| Precedencia de `NO_COLOR` frente a `--color` | [`invocacion.md`](../spec/invocacion.md#variables-de-entorno) | [`salida-y-terminal.md`](../spec/salida-y-terminal.md) (la tabla del efecto final) | auditoría 2026-09-18 |
| Cuándo el rechazo de `--json` es texto plano y no el sobre de error | [`contrato-json.md`](../spec/contrato-json.md#los-errores-en-json) | [`cmd/export.md`](../spec/cmd/export.md), [`cmd/prime.md`](../spec/cmd/prime.md), [`cmd/config.md`](../spec/cmd/config.md), [`cmd/index.md`](../spec/cmd/index.md) (cada mensaje literal) | auditoría 2026-09-18 |
| Excepción de `doctor --fix --dry-run` al código 7 | [`cmd/flags-globales.md`](../spec/cmd/flags-globales.md) | [`cmd/doctor.md`](../spec/cmd/doctor.md#el-informe-en-seco-de---fix---dry-run) (el razonamiento completo y su propia tabla de códigos) | auditoría 2026-09-18 |
| Lista de avisos y sus campos JSON | [`salida-y-terminal.md`](../spec/salida-y-terminal.md#notas-y-avisos) | [`cmd/ls.md`](../spec/cmd/ls.md#el-esquema-json) y [`contrato-json.md`](../spec/contrato-json.md) (dónde sale cada uno en `data`) | auditoría 2026-09-18 |
| Los tres ficheros de una instantánea (`snapshot.ndjson`, `board.json`, `<id>.id`) | [`cmd/snapshot.md`](../spec/cmd/snapshot.md#qué-entra-en-la-revisión) | [`cmd/init.md`](../spec/cmd/init.md) (sus tablas de error de `--from`) | auditoría 2026-09-18 |
| Coerción de `null` en el NDJSON de lote | [`cmd/new.md`](../spec/cmd/new.md#el-modo-lote) | [`cmd/export.md`](../spec/cmd/export.md) (debe producir en el sentido contrario) | auditoría 2026-09-18 |
| Validación de vocabulario al importar (`id`, prefijo, claves de `ext`, valores) | [`cmd/new.md`](../spec/cmd/new.md#el-modo-lote) | [`cmd/export.md`](../spec/cmd/export.md) (declara la precondición, no repite los errores) | auditoría 2026-09-18 |
| Identificador ascendente como criterio de orden | [`cmd/ls.md`](../spec/cmd/ls.md#comportamiento-caso-a-caso) (desempate de listados) | [`modelo-de-datos/index.md`](../spec/modelo-de-datos/index.md#los-campos-derivados) (`blocks` reusa el mismo criterio) | auditoría 2026-09-18 |
| Cuándo el puntero lleva `path` | [`cmd/init.md`](../spec/cmd/init.md) (siempre que se usó `--at`) | [`resolucion-del-tablero.md`](../spec/resolucion-del-tablero.md#el-fichero-bisojson-y-sus-claves) | auditoría 2026-09-18 |
| Presupuesto de 25 ms y sus excepciones | [`presupuestos.md`](../spec/presupuestos.md) | [`cmd/snapshot.md`](../spec/cmd/snapshot.md), [`cmd/doctor.md`](../spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros) (sus propios topes) | auditoría 2026-09-18 |
| Forma de `Comment` (autor, instante, cuerpo) | [`modelo-de-datos/comentarios.md`](../spec/modelo-de-datos/comentarios.md) | [`modelo-de-datos/pregunta-abierta.md`](../spec/modelo-de-datos/pregunta-abierta.md) (`Question` la comparte, sin `key`) | auditoría 2026-09-18 |
| Claves del objeto `where` | [`cmd/where.md`](../spec/cmd/where.md) (el ejemplo) | [`contrato-json.md`](../spec/contrato-json.md) (la tabla de kinds) | auditoría 2026-09-18 |
| Todas las claves de `data.task` (`task.list` + cuerpo) | [`cmd/ls.md`](../spec/cmd/ls.md#el-esquema-json) (`task.list`) | [`cmd/get.md`](../spec/cmd/get.md#el-esquema-json) (repite todas a propósito, para que el ejemplo sirva de esquema completo) | auditoría 2026-09-18 |
| Códigos de salida, uno por regla | [`codigos-de-salida.md`](../spec/codigos-de-salida.md) | [`garantias.md`](../spec/garantias.md), [`detalles.md`](detalles.md) | TASK-56, 2026-09-15 |
| La urgencia y su fórmula | [`modelo-de-datos/urgencia.md`](../spec/modelo-de-datos/urgencia.md) | [`cmd/prime.md`](../spec/cmd/prime.md), [`cmd/verbos-del-ciclo.md`](../spec/cmd/verbos-del-ciclo.md) | TASK-56, 2026-09-15 |
| El leasing (renovación, vaciado) | [`lease.md`](../spec/lease.md) | [`modelo-de-estados.md`](modelo-de-estados.md) | TASK-56, 2026-09-15 |
| El presupuesto del mensaje de arranque | [`vocabulario-y-mensaje-de-arranque.md`](vocabulario-y-mensaje-de-arranque.md) | [`estabilidad.md`](../spec/estabilidad.md) | TASK-56, 2026-09-15 |
| El modelo de datos de una tarea | [`modelo-de-datos/index.md`](../spec/modelo-de-datos/index.md) | [`familias-de-flags.md`](../spec/familias-de-flags.md) | 2026-09-12 |

**Dos puntos quedaron pendientes de la auditoría de 2026-09-18, sin fuente decidida todavía**: si un
`string` (texto de una línea: `title`, `author`, `Criterion.text`, valores de `ext`) admite `\r` o
`\n`, y el esquema completo de `data.filters` en `task.list` para los filtros de `biso ls` que hoy no
tienen clave documentada. El detalle de cada uno está en el acta interna de esa auditoría (no se
publica en el sitio, así que no lleva enlace desde aquí).

Y una cuarta, sobre este documento en particular: la especificación no justifica sus decisiones, y esa
regla es fácil de romper sin darse cuenta. La justificación no solo se esconde en la prosa, también en
la estructura. Una tabla llegó a tener una columna titulada "Por qué" que sobrevivió a cuatro
revisiones, dos de ellas dedicadas expresamente a cazar justificaciones, porque todo el mundo buscaba
frases y esa vivía en una celda.
